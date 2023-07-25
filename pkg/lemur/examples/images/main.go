package main

import (
	"context"
	"fmt"
	"os"

	lemur "chatgpt-go/pkg/lemur"
)

func main() {
	client := lemur.NewClient(os.Getenv("lemur_API_KEY"))

	respUrl, err := client.CreateImage(
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
	fmt.Println(respUrl.Data[0].URL)
}
