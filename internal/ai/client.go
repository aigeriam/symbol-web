package ai

import (
	"net/http"
	"os"
)

type Client struct {
	BaseURL    string //llm's api server address
	Model      string
	APIKey     string //llm's provider check the api adress
	httpClient *http.Client
}

func NewClientFromEnv() *Client {
	return &Client{
		BaseURL: os.Getenv("LLM_BASE_URL"),
		Model:   os.Getenv("LLM_MODEL"),
		APIKey:  os.Getenv("LLM_API_KEY"),
		httpClient: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}
