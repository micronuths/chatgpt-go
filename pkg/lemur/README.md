# Go lemur
[![Go Reference](https://pkg.go.dev/badge/lemur.svg)](https://pkg.go.dev/lemur)
[![Go Report Card](https://goreportcard.com/badge/lemur)](https://goreportcard.com/report/lemur)
[![codecov](https://codecov.io/gh/sashabaranov/go-lemur/branch/master/graph/badge.svg?token=bCbIfHLIsW)](https://codecov.io/gh/sashabaranov/go-lemur)

This library provides unofficial Go clients for [lemur API](https://platform.lemur.com/). We support: 

* ChatGPT
* GPT-3, GPT-4
* DALL·E 2
* Whisper

### Installation:
```
go get lemur
```
Currently, go-lemur requires Go version 1.18 or greater.

### ChatGPT example usage:

```go
package main

import (
	"context"
	"fmt"
	lemur "lemur"
)

func main() {
	client := lemur.NewClient("your token")
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

```

### Getting an lemur API Key:

1. Visit the lemur website at [https://platform.lemur.com/account/api-keys](https://platform.lemur.com/account/api-keys).
2. If you don't have an account, click on "Sign Up" to create one. If you do, click "Log In".
3. Once logged in, navigate to your API key management page.
4. Click on "Create new secret key".
5. Enter a name for your new key, then click "Create secret key".
6. Your new API key will be displayed. Use this key to interact with the lemur API.

**Note:** Your API key is sensitive information. Do not share it with anyone.

### Other examples:

<details>
<summary>ChatGPT streaming completion</summary>

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	lemur "lemur"
)

func main() {
	c := lemur.NewClient("your token")
	ctx := context.Background()

	req := lemur.ChatCompletionRequest{
		Model:     lemur.GPT3Dot5Turbo,
		MaxTokens: 20,
		Messages: []lemur.ChatCompletionMessage{
			{
				Role:    lemur.ChatMessageRoleUser,
				Content: "Lorem ipsum",
			},
		},
		Stream: true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Stream response: ")
	for {
		response, err := stream.Recv()
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
```
</details>

<details>
<summary>GPT-3 completion</summary>

```go
package main

import (
	"context"
	"fmt"
	lemur "lemur"
)

func main() {
	c := lemur.NewClient("your token")
	ctx := context.Background()

	req := lemur.CompletionRequest{
		Model:     lemur.GPT3Ada,
		MaxTokens: 5,
		Prompt:    "Lorem ipsum",
	}
	resp, err := c.CreateCompletion(ctx, req)
	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return
	}
	fmt.Println(resp.Choices[0].Text)
}
```
</details>

<details>
<summary>GPT-3 streaming completion</summary>

```go
package main

import (
	"errors"
	"context"
	"fmt"
	"io"
	lemur "lemur"
)

func main() {
	c := lemur.NewClient("your token")
	ctx := context.Background()

	req := lemur.CompletionRequest{
		Model:     lemur.GPT3Ada,
		MaxTokens: 5,
		Prompt:    "Lorem ipsum",
		Stream:    true,
	}
	stream, err := c.CreateCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("CompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("Stream finished")
			return
		}

		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return
		}


		fmt.Printf("Stream response: %v\n", response)
	}
}
```
</details>

<details>
<summary>Audio Speech-To-Text</summary>

```go
package main

import (
	"context"
	"fmt"

	lemur "lemur"
)

func main() {
	c := lemur.NewClient("your token")
	ctx := context.Background()

	req := lemur.AudioRequest{
		Model:    lemur.Whisper1,
		FilePath: "recording.mp3",
	}
	resp, err := c.CreateTranscription(ctx, req)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return
	}
	fmt.Println(resp.Text)
}
```
</details>

<details>
<summary>Audio Captions</summary>

```go
package main

import (
	"context"
	"fmt"
	"os"

	lemur "lemur"
)

func main() {
	c := lemur.NewClient(os.Getenv("lemur_KEY"))

	req := lemur.AudioRequest{
		Model:    lemur.Whisper1,
		FilePath: os.Args[1],
		Format:   lemur.AudioResponseFormatSRT,
	}
	resp, err := c.CreateTranscription(context.Background(), req)
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
	if _, err := f.WriteString(resp.Text); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}
}
```
</details>

<details>
<summary>DALL-E 2 image generation</summary>

```go
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	lemur "lemur"
	"image/png"
	"os"
)

func main() {
	c := lemur.NewClient("your token")
	ctx := context.Background()

	// Sample image by link
	reqUrl := lemur.ImageRequest{
		Prompt:         "Parrot on a skateboard performs a trick, cartoon style, natural light, high detail",
		Size:           lemur.CreateImageSize256x256,
		ResponseFormat: lemur.CreateImageResponseFormatURL,
		N:              1,
	}

	respUrl, err := c.CreateImage(ctx, reqUrl)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return
	}
	fmt.Println(respUrl.Data[0].URL)

	// Example image as base64
	reqBase64 := lemur.ImageRequest{
		Prompt:         "Portrait of a humanoid parrot in a classic costume, high detail, realistic light, unreal engine",
		Size:           lemur.CreateImageSize256x256,
		ResponseFormat: lemur.CreateImageResponseFormatB64JSON,
		N:              1,
	}

	respBase64, err := c.CreateImage(ctx, reqBase64)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return
	}

	imgBytes, err := base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
	if err != nil {
		fmt.Printf("Base64 decode error: %v\n", err)
		return
	}

	r := bytes.NewReader(imgBytes)
	imgData, err := png.Decode(r)
	if err != nil {
		fmt.Printf("PNG decode error: %v\n", err)
		return
	}

	file, err := os.Create("example.png")
	if err != nil {
		fmt.Printf("File creation error: %v\n", err)
		return
	}
	defer file.Close()

	if err := png.Encode(file, imgData); err != nil {
		fmt.Printf("PNG encode error: %v\n", err)
		return
	}

	fmt.Println("The image was saved as example.png")
}

```
</details>

<details>
<summary>Configuring proxy</summary>

```go
config := lemur.DefaultConfig("token")
proxyUrl, err := url.Parse("http://localhost:{port}")
if err != nil {
	panic(err)
}
transport := &http.Transport{
	Proxy: http.ProxyURL(proxyUrl),
}
config.HTTPClient = &http.Client{
	Transport: transport,
}

c := lemur.NewClientWithConfig(config)
```

See also: https://pkg.go.dev/lemur#ClientConfig
</details>

<details>
<summary>ChatGPT support context</summary>

```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"lemur"
)

func main() {
	client := lemur.NewClient("your token")
	messages := make([]lemur.ChatCompletionMessage, 0)
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Conversation")
	fmt.Println("---------------------")

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)
		messages = append(messages, lemur.ChatCompletionMessage{
			Role:    lemur.ChatMessageRoleUser,
			Content: text,
		})

		resp, err := client.CreateChatCompletion(
			context.Background(),
			lemur.ChatCompletionRequest{
				Model:    lemur.GPT3Dot5Turbo,
				Messages: messages,
			},
		)

		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			continue
		}

		content := resp.Choices[0].Message.Content
		messages = append(messages, lemur.ChatCompletionMessage{
			Role:    lemur.ChatMessageRoleAssistant,
			Content: content,
		})
		fmt.Println(content)
	}
}
```
</details>

<details>
<summary>Azure lemur ChatGPT</summary>

```go
package main

import (
	"context"
	"fmt"

	lemur "lemur"
)

func main() {
	config := lemur.DefaultAzureConfig("your Azure lemur Key", "https://your Azure lemur Endpoint")
	// If you use a deployment name different from the model name, you can customize the AzureModelMapperFunc function
	// config.AzureModelMapperFunc = func(model string) string {
	// 	azureModelMapping = map[string]string{
	// 		"gpt-3.5-turbo": "your gpt-3.5-turbo deployment name",
	// 	}
	// 	return azureModelMapping[model]
	// }

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

```
</details>

<details>
<summary>Azure lemur Embeddings</summary>

```go
package main

import (
	"context"
	"fmt"

	lemur "lemur"
)

func main() {

	config := lemur.DefaultAzureConfig("your Azure lemur Key", "https://your Azure lemur Endpoint")
	config.APIVersion = "2023-05-15" // optional update to latest API version

	//If you use a deployment name different from the model name, you can customize the AzureModelMapperFunc function
	//config.AzureModelMapperFunc = func(model string) string {
	//    azureModelMapping = map[string]string{
	//        "gpt-3.5-turbo":"your gpt-3.5-turbo deployment name",
	//    }
	//    return azureModelMapping[model]
	//}

	input := "Text to vectorize"

	client := lemur.NewClientWithConfig(config)
	resp, err := client.CreateEmbeddings(
		context.Background(),
		lemur.EmbeddingRequest{
			Input: []string{input},
			Model: lemur.AdaEmbeddingV2,
		})

	if err != nil {
		fmt.Printf("CreateEmbeddings error: %v\n", err)
		return
	}

	vectors := resp.Data[0].Embedding // []float32 with 1536 dimensions

	fmt.Println(vectors[:10], "...", vectors[len(vectors)-10:])
}
```
</details>

<details>
<summary>JSON Schema for function calling</summary>

It is now possible for chat completion to choose to call a function for more information ([see developer docs here](https://platform.lemur.com/docs/guides/gpt/function-calling)).

In order to describe the type of functions that can be called, a JSON schema must be provided. Many JSON schema libraries exist and are more advanced than what we can offer in this library, however we have included a simple `jsonschema` package for those who want to use this feature without formatting their own JSON schema payload.

The developer documents give this JSON schema definition as an example:

```json
{
  "name":"get_current_weather",
  "description":"Get the current weather in a given location",
  "parameters":{
    "type":"object",
    "properties":{
        "location":{
          "type":"string",
          "description":"The city and state, e.g. San Francisco, CA"
        },
        "unit":{
          "type":"string",
          "enum":[
              "celsius",
              "fahrenheit"
          ]
        }
    },
    "required":[
        "location"
    ]
  }
}
```

Using the `jsonschema` package, this schema could be created using structs as such:

```go
FunctionDefinition{
  Name: "get_current_weather",
  Parameters: jsonschema.Definition{
    Type: jsonschema.Object,
    Properties: map[string]jsonschema.Definition{
      "location": {
        Type: jsonschema.String,
        Description: "The city and state, e.g. San Francisco, CA",
      },
      "unit": {
        Type: jsonschema.String,
        Enum: []string{"celcius", "fahrenheit"},
      },
    },
    Required: []string{"location"},
  },
}
```

The `Parameters` field of a `FunctionDefinition` can accept either of the above styles, or even a nested struct from another library (as long as it can be marshalled into JSON).
</details>

<details>
<summary>Error handling</summary>

Open-AI maintains clear documentation on how to [handle API errors](https://platform.lemur.com/docs/guides/error-codes/api-errors)

example:
```
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

```
</details>

See the `examples/` folder for more.

### Integration tests:

Integration tests are requested against the production version of the lemur API. These tests will verify that the library is properly coded against the actual behavior of the API, and will  fail upon any incompatible change in the API.

**Notes:**
These tests send real network traffic to the lemur API and may reach rate limits. Temporary network problems may also cause the test to fail.

**Run tests using:**
```
lemur_TOKEN=XXX go test -v -tags=integration ./api_integration_test.go
```

If the `lemur_TOKEN` environment variable is not available, integration tests will be skipped.

## Thank you

We want to take a moment to express our deepest gratitude to the [contributors](https://lemur/graphs/contributors) and sponsors of this project:
- [Carson Kahn](https://carsonkahn.com) of [Spindle AI](https://spindleai.com)

To all of you: thank you. You've helped us achieve more than we ever imagined possible. Can't wait to see where we go next, together!
