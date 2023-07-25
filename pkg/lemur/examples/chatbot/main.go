package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	lemur "chatgpt-go/pkg/lemur"
)

func main() {
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
