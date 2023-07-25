package lemur_test

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"chatgpt-go/pkg/lemur"
)

func Example() {
	client := lemur.NewClient(os.Getenv("lemur_API_KEY"))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		lemur.ChatCompletionRequest{
			Model: lemur.GPT3Dot5Turbo,
			Messages: []lemur.ChatCompletionMessage{
				{
					Role:    lemur.ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}

func ExampleClient_CreateChatCompletionStream() {
	client := lemur.NewClient(os.Getenv("lemur_API_KEY"))

	stream, err := client.CreateChatCompletionStream(
		context.Background(),
		lemur.ChatCompletionRequest{
			Model:     lemur.GPT3Dot5Turbo,
			MaxTokens: 20,
			Messages: []lemur.ChatCompletionMessage{
				{
					Role:    lemur.ChatMessageRoleUser,
					Content: "Lorem ipsum",
				},
			},
			Stream: true,
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Stream response: ")
	for {
		var response lemur.ChatCompletionStreamResponse
		response, err = stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished")
			return
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}

		fmt.Printf(response.Choices[0].Delta.Content)
	}
}

func ExampleClient_CreateCompletion() {
	client := lemur.NewClient(os.Getenv("lemur_API_KEY"))
	resp, err := client.CreateCompletion(
		context.Background(),
		lemur.CompletionRequest{
			Model:     lemur.GPT3Ada,
			MaxTokens: 5,
			Prompt:    "Lorem ipsum",
		},
	)
	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return
	}
	fmt.Println(resp.Choices[0].Text)
}

func ExampleClient_CreateCompletionStream() {
	client := lemur.NewClient(os.Getenv("lemur_API_KEY"))
	stream, err := client.CreateCompletionStream(
		context.Background(),
		lemur.CompletionRequest{
			Model:     lemur.GPT3Ada,
			MaxTokens: 5,
			Prompt:    "Lorem ipsum",
			Stream:    true,
		},
	)
	if err != nil {
		fmt.Printf("CompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	for {
		var response lemur.CompletionResponse
		response, err = stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("Stream finished")
			return
		}

		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return
		}

		fmt.Printf("Stream response: %#v\n", response)
	}
}

func ExampleClient_CreateTranscription() {
	client := lemur.NewClient(os.Getenv("lemur_API_KEY"))
	resp, err := client.CreateTranscription(
		context.Background(),
		lemur.AudioRequest{
			Model:    lemur.Whisper1,
			FilePath: "recording.mp3",
		},
	)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return
	}
	fmt.Println(resp.Text)
}

func ExampleClient_CreateTranscription_captions() {
	client := lemur.NewClient(os.Getenv("lemur_API_KEY"))

	resp, err := client.CreateTranscription(
		context.Background(),
		lemur.AudioRequest{
			Model:    lemur.Whisper1,
			FilePath: os.Args[1],
			Format:   lemur.AudioResponseFormatSRT,
		},
	)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return
	}
	f, err := os.Create(os.Args[1] + ".srt")
	if err != nil {
		fmt.Printf("Could not open file: %v\n", err)
		return
	}
	defer f.Close()
	if _, err = f.WriteString(resp.Text); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}
}

func ExampleClient_CreateTranslation() {
	client := lemur.NewClient(os.Getenv("lemur_API_KEY"))
	resp, err := client.CreateTranslation(
		context.Background(),
		lemur.AudioRequest{
			Model:    lemur.Whisper1,
			FilePath: "recording.mp3",
		},
	)
	if err != nil {
		fmt.Printf("Translation error: %v\n", err)
		return
	}
	fmt.Println(resp.Text)
}

func ExampleClient_CreateImage() {
	client := lemur.NewClient(os.Getenv("lemur_API_KEY"))

	respURL, err := client.CreateImage(
		context.Background(),
		lemur.ImageRequest{
			Prompt:         "Parrot on a skateboard performs a trick, cartoon style, natural light, high detail",
			Size:           lemur.CreateImageSize256x256,
			ResponseFormat: lemur.CreateImageResponseFormatURL,
			N:              1,
		},
	)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return
	}
	fmt.Println(respURL.Data[0].URL)
}

func ExampleClient_CreateImage_base64() {
	client := lemur.NewClient(os.Getenv("lemur_API_KEY"))

	resp, err := client.CreateImage(
		context.Background(),
		lemur.ImageRequest{
			Prompt:         "Portrait of a humanoid parrot in a classic costume, high detail, realistic light, unreal engine",
			Size:           lemur.CreateImageSize512x512,
			ResponseFormat: lemur.CreateImageResponseFormatB64JSON,
			N:              1,
		},
	)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return
	}

	b, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		fmt.Printf("Base64 decode error: %v\n", err)
		return
	}

	f, err := os.Create("example.png")
	if err != nil {
		fmt.Printf("File creation error: %v\n", err)
		return
	}
	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		fmt.Printf("File write error: %v\n", err)
		return
	}

	fmt.Println("The image was saved as example.png")
}

func ExampleClientConfig_clientWithProxy() {
	config := lemur.DefaultConfig(os.Getenv("lemur_API_KEY"))
	port := os.Getenv("lemur_PROXY_PORT")
	proxyURL, err := url.Parse(fmt.Sprintf("http://localhost:%s", port))
	if err != nil {
		panic(err)
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	config.HTTPClient = &http.Client{
		Transport: transport,
	}

	client := lemur.NewClientWithConfig(config)

	client.CreateChatCompletion( //nolint:errcheck // outside of the scope of this example.
		context.Background(),
		lemur.ChatCompletionRequest{
			// etc...
		},
	)
}

func Example_chatbot() {
	client := lemur.NewClient(os.Getenv("lemur_API_KEY"))

	req := lemur.ChatCompletionRequest{
		Model: lemur.GPT3Dot5Turbo,
		Messages: []lemur.ChatCompletionMessage{
			{
				Role:    lemur.ChatMessageRoleSystem,
				Content: "you are a helpful chatbot",
			},
		},
	}
	fmt.Println("Conversation")
	fmt.Println("---------------------")
	fmt.Print("> ")
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		req.Messages = append(req.Messages, lemur.ChatCompletionMessage{
			Role:    lemur.ChatMessageRoleUser,
			Content: s.Text(),
		})
		resp, err := client.CreateChatCompletion(context.Background(), req)
		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			continue
		}
		fmt.Printf("%s\n\n", resp.Choices[0].Message.Content)
		req.Messages = append(req.Messages, resp.Choices[0].Message)
		fmt.Print("> ")
	}
}

func ExampleDefaultAzureConfig() {
	azureKey := os.Getenv("AZURE_lemur_API_KEY")       // Your azure API key
	azureEndpoint := os.Getenv("AZURE_lemur_ENDPOINT") // Your azure lemur endpoint
	config := lemur.DefaultAzureConfig(azureKey, azureEndpoint)
	client := lemur.NewClientWithConfig(config)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		lemur.ChatCompletionRequest{
			Model: lemur.GPT3Dot5Turbo,
			Messages: []lemur.ChatCompletionMessage{
				{
					Role:    lemur.ChatMessageRoleUser,
					Content: "Hello Azure lemur!",
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}

// Open-AI maintains clear documentation on how to handle API errors.
//
// see: https://platform.lemur.com/docs/guides/error-codes/api-errors
func ExampleAPIError() {
	var err error // Assume this is the error you are checking.
	e := &lemur.APIError{}
	if errors.As(err, &e) {
		switch e.HTTPStatusCode {
		case 401:
		// invalid auth or key (do not retry)
		case 429:
		// rate limiting or engine overload (wait and retry)
		case 500:
		// lemur server error (retry)
		default:
			// unhandled
		}
	}
}
