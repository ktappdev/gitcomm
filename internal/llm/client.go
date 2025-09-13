package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ktappdev/gitcomm/internal/config"
)

const (
	// Default settings for commit messages
	DefaultMaxTokens   int32   = 400 // Allow for detailed commit messages
	DefaultTemperature float32 = 0.7 // Slightly creative but not too random
)

type ClientConfig struct {
	MaxTokens   int32
	Temperature float32
}

type Client struct {
	apiKey      string
	apiURL      string
	maxTokens   int32
	temperature float32
	client      *http.Client
	models      []string
}

func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = DefaultMaxTokens
	}

	if cfg.Temperature == 0 {
		cfg.Temperature = DefaultTemperature
	}

	appConfig, err := config.LoadConfig()
	if err != nil {
		// Log the error but proceed, allowing env vars to still function if file is missing/corrupt
		// Initialize an empty config so that subsequent checks don't nil pointer
		appConfig = &config.Config{}
	}

	apiKey := appConfig.OpenRouterAPIKey
	if apiKey == "" {
		return nil, fmt.Errorf("OpenRouter API key not set in config file or %s environment variable", config.OpenRouterAPIKeyEnv)
	}

	// Models to try in order (primary first, then fallbacks)
	models := []string{
		"meta-llama/llama-3.3-8b-instruct:free", // Primary: Free and capable
		"meta-llama/llama-4-scout",             // Fallback 1: Strong performance
		"google/gemini-2.5-flash-lite",         // Fallback 2: Fast and capable
	}

	return &Client{
		apiKey:      apiKey,
		apiURL:      config.OpenRouterAPIURL,
		maxTokens:   cfg.MaxTokens,
		temperature: cfg.Temperature,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		models: models,
	}, nil
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) SendPrompt(prompt string) (string, error) {
	var lastErr error
	
	for i, model := range c.models {
		// Show which model we're trying
		if i == 0 {
			fmt.Printf("⚡ Using %s\n", getModelDisplayName(model))
		} else {
			fmt.Printf("🔄 Falling back to %s\n", getModelDisplayName(model))
		}
		
		response, err := c.tryModel(model, prompt)
		if err == nil {
			return response, nil
		}
		
		lastErr = err
		
		// Only show error if this isn't the last model
		if i < len(c.models)-1 {
			if strings.Contains(lastErr.Error(), "empty response") {
				fmt.Printf("⚠️  %s returned empty response, trying next model...\n", getModelDisplayName(model))
			} else {
				fmt.Printf("⚠️  %s failed, trying next model...\n", getModelDisplayName(model))
			}
		}
	}
	
	return "", fmt.Errorf("all models failed, last error: %w", lastErr)
}

func (c *Client) tryModel(model, prompt string) (string, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":  c.maxTokens,
		"temperature": c.temperature,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Title", "GitComm")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status: %d, response: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices returned from API")
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format: invalid choice format")
	}

	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format: no message in choice")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format: content is not a string")
	}

	// Check if content is empty or just whitespace
	content = strings.TrimSpace(content)
	if content == "" {
		return "", fmt.Errorf("model returned empty response")
	}

	return content, nil
}

func getModelDisplayName(model string) string {
	switch model {
	case "meta-llama/llama-3.3-8b-instruct:free":
		return "Llama 3.3 8B Instruct (Free)"
	case "google/gemini-2.5-flash-lite":
		return "Gemini 2.5 Flash Lite"
	case "meta-llama/llama-4-scout":
		return "Llama 4 Scout"
	case "openai/gpt-oss-20b":
		return "GPT OSS 20B"
	default:
		return model
	}
}
