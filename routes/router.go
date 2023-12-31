package routes

import (
	"chatgpt-go/global"
	"chatgpt-go/model"
	"chatgpt-go/service"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"chatgpt-go/pkg/lemur"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func VerifyEndpoint(c *gin.Context) {
	var req model.VerifyRequest
	if err := c.BindJSON(&req); err != nil || !isValidToken(req.Token) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "Error",
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Verify successfully",
		"data":    nil,
	})
}

func isValidToken(token string) bool {
	return token != "" && global.Config.System.AuthSecretKey == token
}

func SessionEndpoint(c *gin.Context) {
	authorizationHeader := "Bearer " + global.OpenAIKey
	c.Request.Header.Set("Authorization", authorizationHeader)

	authSecretKey := global.Config.System.AuthSecretKey
	isAuthenticated := authSecretKey != ""

	response := createResponse(isAuthenticated)
	c.JSON(http.StatusOK, response)
}

func createResponse(isAuthenticated bool) gin.H {
	return gin.H{
		"status":  "Success",
		"message": "",
		"data": gin.H{
			"auth":  isAuthenticated,
			"model": "ChatGPTAPI",
		},
	}
}

func GetConfig(c *gin.Context) {
	response, err := service.ChatConfig()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func ChatProcess(chatStorage *ChatStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置响应头的 Content-Type 为 application/octet-stream
		c.Header("Content-Type", "application/octet-stream")

		// 获取响应写入器对象，并判断是否支持刷新缓冲区
		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			c.AbortWithError(http.StatusInternalServerError, errors.New("Streaming not supported"))
			return
		}

		// 解析请求参数
		var req model.ChatRequest
		err := c.BindJSON(&req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		if global.Config.System.OpenAIKey == "" {
			panic(errors.New("Missing OPENAI_API_KEY environment variable"))
		}

		config := lemur.DefaultConfig(global.Config.System.OpenAIKey)
		socksHost := global.Config.System.SocksHost
		socksPort := global.Config.System.SocksPort
		httpsProxy := global.Config.System.HttpsProxy

		if socksHost != "" && socksPort != "" {
			proxyUrl, err := url.Parse("socks5://" + socksHost + ":" + socksPort)
			if err != nil {
				panic(err)
			}
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			}
			config.HTTPClient = &http.Client{
				Transport: transport,
			}
		} else if httpsProxy != "" {
			proxyUrl, err := url.Parse("https://" + httpsProxy)
			if err != nil {
				panic(err)
			}
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			}
			config.HTTPClient = &http.Client{
				Transport: transport,
			}
		}

		client := lemur.NewClientWithConfig(config)

		if req.Options.ParentMessageId == "" {
			req.Options.ParentMessageId = uuid.NewString()
		}
		newMessageId := uuid.NewString()
		chatStorage.AddMessage(newMessageId, req.Options.ParentMessageId, lemur.ChatCompletionMessage{
			Role:    lemur.ChatMessageRoleUser,
			Content: req.Prompt,
		})
		messages, err := chatStorage.GetMessages(newMessageId)
		reqData := lemur.ChatCompletionRequest{
			Model:    lemur.GPT3Dot5Turbo,
			Messages: messages,
			Stream:   true,
		}

		// fmt.Printf("Request data: %v\n", reqData)
		stream, err := client.CreateChatCompletionStream(c, reqData)
		if err != nil {
			fmt.Printf("CompletionStream error: %v\n", err)
			return
		}
		defer stream.Close()

		text := ""
		messageId := ""
		for {
			response, err := stream.Recv()

			if errors.Is(err, io.EOF) {
				if messageId != "" {
					chatStorage.AddMessage(messageId, newMessageId, lemur.ChatCompletionMessage{
						Role:    lemur.ChatMessageRoleAssistant,
						Content: text,
					})
				}
				fmt.Println("Stream finished")
				return
			}

			if err != nil {
				fmt.Printf("Stream error: %v\n", err)
				return
			}

			// fmt.Printf("		Stream response: %v\n", response)

			messageId = response.ID
			text = text + response.Choices[0].Delta.Content
			resp := model.ChatResponse{
				Role:            lemur.ChatMessageRoleAssistant,
				Id:              response.ID,
				ParentMessageId: newMessageId,
				Text:            text,
				Delta:           response.Choices[0].Delta.Content,
				Detail:          response,
			}
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				fmt.Printf("JSON marshaling error: %v\n", err)
				return
			}

			_, err = c.Writer.Write(jsonResp)
			if err != nil {
				fmt.Printf("Writing response error: %v\n", err)
				return
			}

			// 刷新缓冲区，发送数据
			flusher.Flush()

			// 在 response 结构体后面添加换行符，以便进行流式传输
			_, err = c.Writer.Write([]byte("\n"))
			if err != nil {
				fmt.Printf("Writing newline error: %v\n", err)
				return
			}
		}
	}
}

func ChatProcessV2(chatStorage *ChatStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置响应头的 Content-Type 为 application/octet-stream
		c.Header("Content-Type", "application/octet-stream")

		// 获取响应写入器对象，并判断是否支持刷新缓冲区
		flusher, ok := c.Writer.(http.Flusher) //断言
		if !ok {
			c.AbortWithError(http.StatusInternalServerError, errors.New("Streaming not supported"))
			return
		}

		// 解析请求参数
		var req model.ChatRequest
		err := c.BindJSON(&req)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		// spew.Dump(req)
		if global.Config.System.OpenAIKey == "" {
			panic(errors.New("Missing OPENAI_API_KEY environment variable"))
		}

		config := lemur.DefaultConfig(global.Config.System.OpenAIKey)
		socksHost := global.Config.System.SocksHost
		socksPort := global.Config.System.SocksPort
		httpsProxy := global.Config.System.HttpsProxy
		httpProxy := global.Config.System.HttpProxy

		if socksHost != "" && socksPort != "" {
			proxyUrl, err := url.Parse("socks5://" + socksHost + ":" + socksPort)
			if err != nil {
				panic(err)
			}
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			}
			config.HTTPClient = &http.Client{
				Transport: transport,
			}
		} else if httpsProxy != "" {
			proxyUrl, err := url.Parse("https://" + httpsProxy)
			if err != nil {
				panic(err)
			}
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			}
			config.HTTPClient = &http.Client{
				Transport: transport,
			}
		} else if httpProxy != "" {
			proxyUrl, err := url.Parse("http://" + httpProxy)
			if err != nil {
				panic(err)
			}
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyUrl),
			}
			config.HTTPClient = &http.Client{
				Transport: transport,
			}
		}
		config.BaseURL = "http://lemurchat.anfans.cn/api"
		client := lemur.NewClientWithConfig(config)

		/*
		   1、从客户端解析请求，存入数据库
		*/
		if req.Options.ParentMessageId == "" { // chatcmpl-7c1gUEGvLGP87IsXy7GQAO3oC7EZT
			req.Options.ParentMessageId = "chatcmpl-start"
			// fmt.Println(`req.Options.ParentMessageId == "",`, req.Options.ParentMessageId)
		}
		newMessageIdUser := uuid.NewString()
		err = chatStorage.AddMessage(newMessageIdUser, req.Options.ParentMessageId, lemur.ChatCompletionMessage{
			Role:    lemur.ChatMessageRoleUser,
			Content: req.Prompt,
		})
		if err != nil {
			fmt.Println("Error when chatStorage.AddMessage", err)
		}

		/*
		   1、从数据库中取出，构造request
		*/
		msgCtx, err := chatStorage.GetContextMessages(newMessageIdUser)
		if err != nil {
			fmt.Println("Error when chatStorage.GetContextMessages", err)
		}
		messages := []lemur.ChatCompletionMessageLemur{
			{Content: "", ID: "LEMUR_AI_SYSTEM_SETTING", IsSensitive: false, NeedCheck: false, Role: "system"},
			// {Content: "请求", IsSensitive: false, NeedCheck: true, Role: "user"},
			// {Content: "你可以请求我做什么？请具体指明你需要我做的事情。", IsSensitive: false, NeedCheck: true, Role: "assistant"},
		}
		msgCtxLen := len(msgCtx)
		for idx := range msgCtx {
			messages = append(messages, lemur.ChatCompletionMessageLemur{Content: msgCtx[msgCtxLen-1-idx].Content, IsSensitive: false, NeedCheck: true, Role: msgCtx[msgCtxLen-1-idx].Role})
		}
		// messages = append(messages, lemur.ChatCompletionMessageLemur{Content: req.Prompt, IsSensitive: false, NeedCheck: true, Role: "user"})
		msg, err := json.Marshal(messages)
		if err != nil {
			fmt.Printf("json.Marshal(messages) error: %v\n", err)
			return
		}
		reqBody := lemur.ChatCompletionRequestLemur{
			Messages: string(msg),
		}

		// fmt.Printf("Request data: %v\n", reqBody)
		stream, err := client.CreateChatCompletionStreamLemur(c, reqBody)
		// stream, err := client.CreateChatCompletionStream(c, reqBody)
		if err != nil {
			fmt.Printf("CompletionStream error: %v\n", err)
			return
		}
		defer stream.Close()

		var text string                         // 全局变量
		var currentMessageId = uuid.NewString() // 全局变量
		for {
			response, err := stream.Recv()

			if errors.Is(err, io.EOF) {
				if currentMessageId != "" {
					err := chatStorage.AddMessage(currentMessageId, newMessageIdUser, lemur.ChatCompletionMessage{
						Role:    lemur.ChatMessageRoleAssistant,
						Content: text,
					})
					fmt.Println("Error when chatStorage.AddMessage=", err)
				}
				// fmt.Println("Stream finished")
				// fmt.Println("text=", text)
				return
			}

			if err != nil {
				fmt.Printf("Error when stream.Recv() : %v\n", err)
				return
			}

			// fmt.Printf("		Stream response: %v\n", response)
			json.Marshal(response)

			text = text + response.Choices[0].Delta.Content
			resp := model.ChatResponseLemur{
				Role:            lemur.ChatMessageRoleAssistant,
				Id:              currentMessageId, //response.ID,
				ParentMessageId: newMessageIdUser,
				Text:            text,
				Delta:           response.Choices[0].Delta.Content,
				Detail: lemur.ChatCompletionStreamResponseLemur{
					ID:      response.ID,
					Object:  response.Object,
					Created: int64(response.Created),
					Model:   response.Model,
					// Choices: response.Choices,
				},
			}
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				fmt.Printf("Error when JSON marshaling: %v\n", err)
				return
			}

			_, err = c.Writer.Write(jsonResp)
			if err != nil {
				fmt.Printf("Error when Writing response: %v\n", err)
				return
			}

			// 刷新缓冲区，发送数据
			flusher.Flush()

			// 在 response 结构体后面添加换行符，以便进行流式传输
			_, err = c.Writer.Write([]byte("\n"))
			if err != nil {
				fmt.Printf("Error when Writing newline: %v\n", err)
				return
			}
		}

	}
}
