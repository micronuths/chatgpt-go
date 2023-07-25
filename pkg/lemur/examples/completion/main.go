package main

import (
	"context"
	"fmt"
	"os"

	lemur "chatgpt-go/pkg/lemur"
)

func main() {
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
