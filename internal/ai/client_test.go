package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patchnote/patchnote/internal/config"
)

func TestNewClientGroq(t *testing.T) {
	cfg := &config.Config{
		Provider: "groq",
		Model:    "test-model",
		APIKey:   "test-key",
	}

	client, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewClientUnsupported(t *testing.T) {
	cfg := &config.Config{
		Provider: "openai",
	}

	_, err := New(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported provider")
}

func TestGroqClientComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		var req chatCompletionRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "test-model", req.Model)
		assert.Len(t, req.Messages, 1)
		assert.Equal(t, "user", req.Messages[0].Role)

		resp := chatCompletionResponse{
			Model: "test-model",
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "feat: add test"}},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &GroqClient{
		apiKey:  "test-key",
		model:   "test-model",
		baseURL: server.URL,
		httpClient: &http.Client{
			Transport: http.DefaultTransport,
		},
	}

	resp, err := client.Complete(context.Background(), Request{
		Messages: []Message{
			{Role: "user", Content: "test prompt"},
		},
		Temperature: 0.2,
	})

	require.NoError(t, err)
	assert.Equal(t, "feat: add test", resp.Content)
	assert.Equal(t, "test-model", resp.Model)
	assert.Equal(t, 10, resp.Usage.PromptTokens)
	assert.Equal(t, 5, resp.Usage.CompletionTokens)
	assert.Equal(t, 15, resp.Usage.TotalTokens)
}

func TestGroqClientNoAPIKey(t *testing.T) {
	client := &GroqClient{
		apiKey: "",
		model:  "test",
	}

	_, err := client.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "test"}},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key not configured")
}

func TestGroqClientAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		resp := errorResponse{}
		resp.Error.Message = "invalid api key"
		resp.Error.Type = "invalid_request_error"
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &GroqClient{
		apiKey:  "bad-key",
		model:   "test-model",
		baseURL: server.URL,
		httpClient: &http.Client{
			Transport: http.DefaultTransport,
		},
	}

	_, err := client.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "test"}},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid api key")
}

func TestGroqClientNoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatCompletionResponse{
			Model:   "test",
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &GroqClient{
		apiKey:  "test-key",
		model:   "test",
		baseURL: server.URL,
		httpClient: &http.Client{
			Transport: http.DefaultTransport,
		},
	}

	_, err := client.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "test"}},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no choices")
}

func TestGroqClientValidateKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/models", r.URL.Path)
		assert.Equal(t, "Bearer valid-key", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &GroqClient{
		httpClient: &http.Client{},
		baseURL:    server.URL,
	}

	err := client.ValidateKey(context.Background(), "valid-key")
	assert.NoError(t, err)
}

func TestGroqClientValidateKeyInvalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := &GroqClient{
		httpClient: &http.Client{},
		baseURL:    server.URL,
	}

	err := client.ValidateKey(context.Background(), "bad-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid API key")
}

func TestGroqClientDefaultModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatCompletionRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		resp := chatCompletionResponse{
			Model: req.Model,
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "ok"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &GroqClient{
		apiKey:  "test-key",
		model:   "default-model",
		baseURL: server.URL,
		httpClient: &http.Client{
			Transport: http.DefaultTransport,
		},
	}

	// When no model specified in request, should use client default
	resp, err := client.Complete(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "test"}},
	})

	require.NoError(t, err)
	assert.Equal(t, "default-model", resp.Model)
}
