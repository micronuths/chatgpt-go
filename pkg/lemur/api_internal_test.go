package lemur

import (
	"context"
	"testing"
)

func TestlemurFullURL(t *testing.T) {
	cases := []struct {
		Name   string
		Suffix string
		Expect string
	}{
		{
			"ChatCompletionsURL",
			"/chat/completions",
			"https://api.lemur.com/v1/chat/completions",
		},
		{
			"CompletionsURL",
			"/completions",
			"https://api.lemur.com/v1/completions",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			az := DefaultConfig("dummy")
			cli := NewClientWithConfig(az)
			actual := cli.fullURL(c.Suffix)
			if actual != c.Expect {
				t.Errorf("Expected %s, got %s", c.Expect, actual)
			}
			t.Logf("Full URL: %s", actual)
		})
	}
}

func TestRequestAuthHeader(t *testing.T) {
	cases := []struct {
		Name      string
		APIType   APIType
		HeaderKey string
		Token     string
		OrgID     string
		Expect    string
	}{
		{
			"lemurDefault",
			"",
			"Authorization",
			"dummy-token-lemur",
			"",
			"Bearer dummy-token-lemur",
		},
		{
			"lemurOrg",
			APITypelemur,
			"Authorization",
			"dummy-token-lemur",
			"dummy-org-lemur",
			"Bearer dummy-token-lemur",
		},
		{
			"lemur",
			APITypelemur,
			"Authorization",
			"dummy-token-lemur",
			"",
			"Bearer dummy-token-lemur",
		},
		{
			"AzureAD",
			APITypeAzureAD,
			"Authorization",
			"dummy-token-azure",
			"",
			"Bearer dummy-token-azure",
		},
		{
			"Azure",
			APITypeAzure,
			AzureAPIKeyHeader,
			"dummy-api-key-here",
			"",
			"dummy-api-key-here",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			az := DefaultConfig(c.Token)
			az.APIType = c.APIType
			az.OrgID = c.OrgID

			cli := NewClientWithConfig(az)
			req, err := cli.newRequest(context.Background(), "POST", "/chat/completions")
			if err != nil {
				t.Errorf("Failed to create request: %v", err)
			}
			actual := req.Header.Get(c.HeaderKey)
			if actual != c.Expect {
				t.Errorf("Expected %s, got %s", c.Expect, actual)
			}
			t.Logf("%s: %s", c.HeaderKey, actual)
		})
	}
}

func TestAzureFullURL(t *testing.T) {
	cases := []struct {
		Name             string
		BaseURL          string
		AzureModelMapper map[string]string
		Model            string
		Expect           string
	}{
		{
			"AzureBaseURLWithSlashAutoStrip",
			"https://httpbin.org/",
			nil,
			"chatgpt-demo",
			"https://httpbin.org/" +
				"lemur/deployments/chatgpt-demo" +
				"/chat/completions?api-version=2023-05-15",
		},
		{
			"AzureBaseURLWithoutSlashOK",
			"https://httpbin.org",
			nil,
			"chatgpt-demo",
			"https://httpbin.org/" +
				"lemur/deployments/chatgpt-demo" +
				"/chat/completions?api-version=2023-05-15",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			az := DefaultAzureConfig("dummy", c.BaseURL)
			cli := NewClientWithConfig(az)
			// /lemur/deployments/{engine}/chat/completions?api-version={api_version}
			actual := cli.fullURL("/chat/completions", c.Model)
			if actual != c.Expect {
				t.Errorf("Expected %s, got %s", c.Expect, actual)
			}
			t.Logf("Full URL: %s", actual)
		})
	}
}
