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
	"github.com/ktappdev/gitcomm/internal/diag"
)

const (
	DefaultMaxTokens      int32   = 400
	DefaultTemperature    float32 = 0.7
	DefaultTimeoutSeconds         = 30
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

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    any    `json:"code"`
	} `json:"error,omitempty"`
}

func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = DefaultMaxTokens
	}
	if cfg.Temperature == 0 {
		cfg.Temperature = DefaultTemperature
	}

	appConfig, cfgErr := config.LoadRuntimeConfig()
	if cfgErr != nil {
		diag.Warn("llm", "continuing with runtime fallback config", "error", cfgErr)
	}

	apiKey := appConfig.OpenRouterAPIKey
	if apiKey == "" {
		if cfgErr != nil {
			return nil, fmt.Errorf("configuration is invalid and no OpenRouter API key is available via %s/%s: %w", config.OpenRouterAPIKeyEnvPrimary, config.OpenRouterAPIKeyEnvLegacy, cfgErr)
		}
		return nil, fmt.Errorf("OpenRouter API key not set in config file or %s/%s environment variables", config.OpenRouterAPIKeyEnvPrimary, config.OpenRouterAPIKeyEnvLegacy)
	}

	models := config.DefaultModels
	if len(appConfig.Models) > 0 {
		models = appConfig.Models
	}
	apiURL := config.OpenRouterAPIURL
	if appConfig.APIURL != "" {
		apiURL = appConfig.APIURL
	}
	maxTokens := cfg.MaxTokens
	if appConfig.MaxTokens > 0 {
		maxTokens = int32(appConfig.MaxTokens)
	}
	temperature := cfg.Temperature
	if appConfig.Temperature > 0 {
		temperature = float32(appConfig.Temperature)
	}
	timeoutSeconds := DefaultTimeoutSeconds
	if appConfig.TimeoutSeconds > 0 {
		timeoutSeconds = appConfig.TimeoutSeconds
	}

	diag.Info("llm", "initialized client", "models", strings.Join(models, ","), "timeout_seconds", timeoutSeconds, "max_tokens", maxTokens, "temperature", temperature, "api_url", apiURL, "config_warning", cfgErr != nil)
	return &Client{
		apiKey:      apiKey,
		apiURL:      apiURL,
		maxTokens:   maxTokens,
		temperature: temperature,
		client:      &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second},
		models:      models,
	}, nil
}

func (c *Client) Close() error { return nil }

func (c *Client) SendPrompt(prompt string) (string, error) {
	var lastErr error
	promptBytes := len([]byte(prompt))
	diag.Info("llm", "sending prompt", "models_count", len(c.models), "prompt_chars", len(prompt), "prompt_bytes", promptBytes, "max_tokens", c.maxTokens)

	for i, model := range c.models {
		if i == 0 {
			fmt.Printf("⚡ Using %s\n", getModelDisplayName(model))
		} else {
			fmt.Printf("🔄 Falling back to %s\n", getModelDisplayName(model))
		}
		response, err := c.tryModel(model, prompt, i+1, len(c.models))
		if err == nil {
			diag.Info("llm", "model succeeded", "model", model, "attempt", i+1)
			return response, nil
		}
		lastErr = err
		diag.Warn("llm", "model attempt failed", "model", model, "attempt", i+1, "error", err)
		if i < len(c.models)-1 {
			fmt.Printf("⚠️  %s failed, trying next model...\n", getModelDisplayName(model))
		}
	}

	return "", fmt.Errorf("all models failed; see diagnostics log for details: %w", lastErr)
}

func (c *Client) tryModel(model, prompt string, attempt, total int) (string, error) {
	requestBody, err := json.Marshal(map[string]any{
		"model":       model,
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
		"max_tokens":  c.maxTokens,
		"temperature": c.temperature,
		"reasoning":   map[string]any{"max_tokens": 0},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}
	startedAt := time.Now()
	diag.Info("llm", "starting model attempt", "model", model, "attempt", attempt, "total_attempts", total, "request_bytes", len(requestBody), "prompt_chars", len(prompt))

	req, err := http.NewRequest("POST", c.apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Title", "GitComm")

	resp, err := c.client.Do(req)
	if err != nil {
		diag.Error("llm", "http request failed", "model", model, "attempt", attempt, "elapsed_ms", time.Since(startedAt).Milliseconds(), "error", err)
		return "", fmt.Errorf("request to %s failed: %w", model, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	diag.Info("llm", "received model response", "model", model, "attempt", attempt, "status", resp.StatusCode, "elapsed_ms", time.Since(startedAt).Milliseconds(), "response_bytes", len(body))

	if resp.StatusCode != http.StatusOK {
		return "", formatAPIError(model, resp.StatusCode, body)
	}

	var result chatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		diag.Error("llm", "failed to parse response", "model", model, "attempt", attempt, "error", err, "body_snippet", diag.Snippet(string(body), 300))
		return "", fmt.Errorf("failed to unmarshal response from %s: %w", model, err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("%s returned no choices", model)
	}
	content := strings.TrimSpace(result.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("%s returned empty response content", model)
	}
	return content, nil
}

func formatAPIError(model string, statusCode int, body []byte) error {
	var result chatResponse
	providerMsg := ""
	if err := json.Unmarshal(body, &result); err == nil && result.Error != nil {
		providerMsg = diag.Snippet(strings.TrimSpace(result.Error.Message), 200)
	}
	bodySnippet := diag.Snippet(string(body), 300)
	diag.Error("llm", "provider returned error", "model", model, "status", statusCode, "provider_message", providerMsg, "body_snippet", bodySnippet)

	base := fmt.Sprintf("%s failed with status %d", model, statusCode)
	if providerMsg != "" {
		base += ": " + providerMsg
	}

	switch statusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("%s. This can happen when the diff/prompt is too large or malformed", base)
	case http.StatusPaymentRequired:
		return fmt.Errorf("%s. The model may require credits or be unavailable on your OpenRouter plan", base)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%s. The model is rate limited right now", base)
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("OpenRouter authentication failed (%d)", statusCode)
	default:
		return fmt.Errorf(base)
	}
}

func getModelDisplayName(model string) string {
	switch model {
	case "meta-llama/llama-3.3-8b-instruct:free":
		return "Llama 3.3 8B Instruct (Free)"
	case "meta-llama/llama-4-scout":
		return "Llama 4 Scout"
	case "google/gemini-2.5-flash-lite":
		return "Gemini 2.5 Flash Lite"
	default:
		return model
	}
}
