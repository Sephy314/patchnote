package ai

import (
	"context"
	"fmt"

	"github.com/Sephy314/patchnote/internal/config"
)

// Message represents a single message in a chat conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Request holds the parameters for an AI completion request.
type Request struct {
	Messages    []Message
	Temperature float64
	Model       string
	MaxTokens   int
}

// Response holds the AI completion response.
type Response struct {
	Content string
	Model   string
	Usage   Usage
}

// Usage tracks token consumption.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Client defines the interface for AI providers.
type Client interface {
	Complete(ctx context.Context, req Request) (*Response, error)
	ValidateKey(ctx context.Context, apiKey string) error
}

// New creates an AI client for the configured provider.
func New(cfg *config.Config) (Client, error) {
	switch cfg.Provider {
	case "groq":
		return NewGroqClient(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}
