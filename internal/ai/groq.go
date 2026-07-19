package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Sephy314/patchnote/internal/config"
)

const (
	groqBaseURL        = "https://api.groq.com/openai/v1"
	groqDefaultTimeout = 60 * time.Second
)

// GroqClient implements the Client interface for the Groq API.
type GroqClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// NewGroqClient creates a new Groq AI client.
func NewGroqClient(cfg *config.Config) *GroqClient {
	return &GroqClient{
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		baseURL: groqBaseURL,
		httpClient: &http.Client{
			Timeout: groqDefaultTimeout,
		},
	}
}

type chatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

type errorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// Complete sends a chat completion request to Groq.
func (c *GroqClient) Complete(ctx context.Context, req Request) (*Response, error) {
	apiKey := c.apiKey
	if apiKey == "" {
		return nil, fmt.Errorf("API key not configured. Run 'patchnote register' to set it up")
	}

	body := chatCompletionRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}
	if body.Model == "" {
		body.Model = c.model
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("AI request failed: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("AI request failed (%s): %s", errResp.Error.Type, errResp.Error.Message)
		}
		return nil, fmt.Errorf("AI request failed with status %d: %s", httpResp.StatusCode, string(respBody))
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(respBody, &completion); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("AI returned no choices")
	}

	return &Response{
		Content: completion.Choices[0].Message.Content,
		Model:   completion.Model,
		Usage: Usage{
			PromptTokens:     completion.Usage.PromptTokens,
			CompletionTokens: completion.Usage.CompletionTokens,
			TotalTokens:      completion.Usage.TotalTokens,
		},
	}, nil
}

// ValidateKey checks if the given API key is valid by making a lightweight request.
func (c *GroqClient) ValidateKey(ctx context.Context, apiKey string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach Groq API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from Groq API", resp.StatusCode)
	}

	return nil
}
