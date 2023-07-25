package model

import "chatgpt-go/pkg/lemur"

// 从客户端传上来的请求
type ChatRequest struct {
	Prompt  string             `json:"prompt"`
	Options ChatRequestOptions `json:"options,omitempty"`
}
type ChatRequestOptions struct {
	ParentMessageId string `json:"parentMessageId"`
}

type VerifyRequest struct {
	Token string `json:"token"`
}

/*
返回给客户端的请求
*/
type ChatResponse struct {
	Role            string                             `json:"role"`
	Id              string                             `json:"id"`
	ParentMessageId string                             `json:"parentMessageId"`
	Delta           string                             `json:"delta"`
	Text            string                             `json:"text"`
	Detail          lemur.ChatCompletionStreamResponse `json:"detail"`
}
type ChatResponseLemur struct {
	Role            string                                  `json:"role"`
	Id              string                                  `json:"id"`
	ParentMessageId string                                  `json:"parentMessageId"`
	Delta           string                                  `json:"delta"`
	Text            string                                  `json:"text"`
	Detail          lemur.ChatCompletionStreamResponseLemur `json:"detail"`
}

// api/config接口 返回的结果
type ChatConfig struct {
	Message string         `json:"message"`
	Data    ChatConfigData `json:"data"`
	Status  string         `json:"status"`
}
type ChatConfigData struct {
	APIModel     string `json:"apiModel"`
	ReverseProxy string `json:"reverseProxy"`
	TimeoutMs    int    `json:"timeoutMs"`
	SocksProxy   string `json:"socksProxy"`
	HttpsProxy   string `json:"httpsProxy"`
	Balance      string `json:"balance"`
}
