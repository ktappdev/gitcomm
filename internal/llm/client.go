package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/ktappdev/gitcomm/config"
)

type Provider string

const (
	ProviderGroq   Provider = "groq"
	ProviderOpenAI Provider = "openai"

	// Default settings for commit messages
	DefaultMaxTokens   = 50  // About 2 lines of text
	DefaultTemperature = 0.7 // Slightly creative but not too random
)

type Client struct {
	provider    Provider
	apiKey      string
	apiURL      string
	model       string
	maxTokens   int
	temperature float64
	client      *http.Client
}

type ClientConfig struct {
	Provider    Provider
	Model       string
	MaxTokens   int
	Temperature float64
}

func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.Provider == "" {
		cfg.Provider = ProviderGroq
	}

	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = DefaultMaxTokens
	}

	if cfg.Temperature == 0 {
		cfg.Temperature = DefaultTemperature
	}

	var apiKey string
	var apiURL string
	var defaultModel string

	switch cfg.Provider {
	case ProviderGroq:
		apiKey = os.Getenv(config.GroqAPIKeyEnv)
		apiURL = config.GroqAPIURL
		defaultModel = "llama-3.1-70b-versatile"
	case ProviderOpenAI:
		apiKey = os.Getenv(config.OpenAIAPIKeyEnv)
		apiURL = config.OpenAIAPIURL
		defaultModel = "gpt-4o-mini"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key not set for provider %s", cfg.Provider)
	}

	if cfg.Model == "" {
		cfg.Model = defaultModel
	}

	return &Client{
		provider:    cfg.Provider,
		apiKey:      apiKey,
		apiURL:      apiURL,
		model:       cfg.Model,
		maxTokens:   cfg.MaxTokens,
		temperature: cfg.Temperature,
		client:      &http.Client{},
	}, nil
}

func (c *Client) SendPrompt(prompt string) (string, error) {
	requestBody, _ := json.Marshal(map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":  c.maxTokens,
		"temperature": c.temperature,
		"stream":      false, // We don't need streaming for commit messages
	})

	req, _ := http.NewRequest("POST", c.apiURL, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	slog.Debug("Got response from LLM", "response", result)

	return extractContent(result)
}

func extractContent(result map[string]interface{}) (string, error) {
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("unexpected response format")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("content is not a string")
	}

	return content, nil
}
