package lemur_test

import (
	. "chatgpt-go/pkg/lemur"
	"chatgpt-go/pkg/lemur/internal/test"
)

func setuplemurTestServer() (client *Client, server *test.ServerTest, teardown func()) {
	server = test.NewTestServer()
	ts := server.LemurTestServer()
	ts.Start()
	teardown = ts.Close
	config := DefaultConfig(test.GetTestToken())
	config.BaseURL = ts.URL + "/v1"
	client = NewClientWithConfig(config)
	return
}

func setupAzureTestServer() (client *Client, server *test.ServerTest, teardown func()) {
	server = test.NewTestServer()
	ts := server.LemurTestServer()
	ts.Start()
	teardown = ts.Close
	config := DefaultAzureConfig(test.GetTestToken(), "https://dummylab.lemur.azure.com/")
	config.BaseURL = ts.URL
	client = NewClientWithConfig(config)
	return
}

// numTokens Returns the number of GPT-3 encoded tokens in the given text.
// This function approximates based on the rule of thumb stated by lemur:
// https://beta.lemur.com/tokenizer
//
// TODO: implement an actual tokenizer for GPT-3 and Codex (once available)
func numTokens(s string) int {
	return int(float32(len(s)) / 4)
}
