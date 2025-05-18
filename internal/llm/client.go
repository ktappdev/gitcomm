package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/generative-ai-go/genai"
	"github.com/ktappdev/gitcomm/internal/config"
	"google.golang.org/api/option"
)

const (
	ProviderGroq   Provider = "groq"
	ProviderOpenAI Provider = "openai"
	ProviderGemini Provider = "gemini"

	// Default settings for commit messages
	DefaultMaxTokens   int32   = 50  // About 2 lines of text
	DefaultTemperature float32 = 0.7 // Slightly creative but not too random
)

type Provider string

type ClientConfig struct {
	Provider    Provider
	Model       string
	MaxTokens   int32
	Temperature float32
}

type Client struct {
	provider     Provider
	apiKey       string
	apiURL       string
	model        string
	maxTokens    int32
	temperature  float32
	client       *http.Client
	geminiClient *genai.Client
}

func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.Provider == "" {
		cfg.Provider = ProviderGemini // Change default to Gemini
	}

	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = DefaultMaxTokens
	}

	if cfg.Temperature == 0 {
		cfg.Temperature = DefaultTemperature
	}

	appConfig, err := config.LoadConfig()
	if err != nil {
		// Log the error but proceed, allowing env vars to still function if file is missing/corrupt
		slog.Warn("Failed to load .gitcomm/config.json", "error", err)
		// Initialize an empty config so that subsequent checks don't nil pointer
		appConfig = &config.Config{}
	}

	var apiKey string
	var apiURL string
	var defaultModel string
	var geminiClient *genai.Client

	switch cfg.Provider {
	case ProviderGemini:
		apiKey = appConfig.GeminiAPIKey // Get from loaded config (JSON or ENV)
		if apiKey == "" {
			return nil, fmt.Errorf("Gemini API key not set in config file or %s environment variable", config.GeminiAPIKeyEnv)
		}
		ctx := context.Background()
		// var err error // err already declared by appConfig loading
		geminiClient, err = genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini client: %v", err)
		}
		defaultModel = "gemini-2.0-flash-lite" // Ensure this is a valid and available model
	case ProviderGroq:
		apiKey = appConfig.GroqAPIKey // Get from loaded config (JSON or ENV)
		apiURL = config.GroqAPIURL
		defaultModel = "llama-3.1-70b-versatile" // Ensure this is valid for Groq
	case ProviderOpenAI:
		apiKey = appConfig.OpenAIAPIKey // Get from loaded config (JSON or ENV)
		apiURL = config.OpenAIAPIURL
		defaultModel = "gpt-4o-mini" // Ensure this is valid for OpenAI
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}

	if cfg.Provider != ProviderGemini && apiKey == "" {
		var envVarName string
		switch cfg.Provider {
		case ProviderGroq:
			envVarName = config.GroqAPIKeyEnv
		case ProviderOpenAI:
			envVarName = config.OpenAIAPIKeyEnv
		}
		return nil, fmt.Errorf("API key for %s not set in config file or %s environment variable", cfg.Provider, envVarName)
	}

	if cfg.Model == "" {
		cfg.Model = defaultModel
	}

	return &Client{
		provider:     cfg.Provider,
		apiKey:       apiKey,
		apiURL:       apiURL,
		model:        cfg.Model,
		maxTokens:    cfg.MaxTokens,
		temperature:  cfg.Temperature,
		client:       &http.Client{},
		geminiClient: geminiClient,
	}, nil
}

func (c *Client) Close() error {
	if c.geminiClient != nil {
		c.geminiClient.Close()
	}
	return nil
}

func (c *Client) SendPrompt(prompt string) (string, error) {
	if c.provider == ProviderGemini {
		return c.sendGeminiPrompt(prompt)
	}

	// For OpenAI-style APIs (Groq and OpenAI)
	response, err := c.sendOpenAIStylePrompt(prompt)
	if err != nil {
		// Try falling back to Gemini if available and the current provider fails
		if c.geminiClient != nil {
			slog.Info("Falling back to Gemini due to error with primary provider", "error", err)
			return c.sendGeminiPrompt(prompt)
		}
		return "", err
	}
	return response, nil
}

func (c *Client) sendGeminiPrompt(prompt string) (string, error) {
	ctx := context.Background()
	model := c.geminiClient.GenerativeModel(c.model)

	model.SetTemperature(float32(c.temperature))
	model.SetMaxOutputTokens(int32(c.maxTokens))

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("error generating content: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	// Extract text from the response part
	if textPart, ok := candidate.Content.Parts[0].(genai.Text); ok {
		return string(textPart), nil
	}

	// If we can't get a text part, return an error
	return "", fmt.Errorf("unexpected response type: %T", candidate.Content.Parts[0])
}

func (c *Client) sendOpenAIStylePrompt(prompt string) (string, error) {
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
		"stream":      false,
	})

	req, _ := http.NewRequest("POST", c.apiURL, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %v", err)
	}
	slog.Debug("Got response from LLM", "response", result)

	return extractContent(result)
}

func extractContent(result map[string]interface{}) (string, error) {
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("unexpected response format: no choices array")
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

	return content, nil
}
