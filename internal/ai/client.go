package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// RequestTimeout is the max timeout allowed for every LLM call
const RequestTimeout = 10 * time.Second

type Client struct {
	BaseURL    string //llm's api server address
	Model      string
	APIKey     string //llm's provider check the api adress
	httpClient *http.Client
}

func (c *Client) MockMode() bool {
	return c.BaseURL == ""
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

// errors messages
var ErrBackendUnavailable = errors.New("ai: llm backend unavailable")
var ErrBadResponse = errors.New("ai: llm returned an unparseable response")

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}
type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Message *chatMessage `json:"message,omitempty"`
}

// complete sends a prompt to LLM backend and return the response and error
func (c *Client) complete(ctx context.Context, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, RequestTimeout)
	defer cancel()
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: RequestTimeout}
	}
	reqBody := chatRequest{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
		Stream: false,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)

	}
	url := c.BaseURL + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("%w: building request: %v", ErrBackendUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Covers network failures AND context-deadline timeouts.
		return "", fmt.Errorf("%w: %v", ErrBackendUnavailable, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w: reading body: %v", ErrBadResponse, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%w: backend returned status %d: %s", ErrBackendUnavailable, resp.StatusCode, string(body))
	}

	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("%w: %v", ErrBadResponse, err)
	}

	if len(parsed.Choices) > 0 {
		return parsed.Choices[0].Message.Content, nil
	}
	if parsed.Message != nil {
		return parsed.Message.Content, nil
	}
	return "", fmt.Errorf("%w: no content in response", ErrBadResponse)

}
