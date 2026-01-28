package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const (
	OpenRouterAPIKeyEnvPrimary = "OPENROUTER_API_KEY"
	OpenRouterAPIKeyEnvLegacy  = "OPEN_ROUTER_API_KEY"
	OpenRouterAPIURL           = "https://openrouter.ai/api/v1/chat/completions"
	DefaultMaxTokens           = 400
	DefaultTemperature         = 0.7
	DefaultTimeoutSeconds      = 30
)

var DefaultModels = []string{
	"openai/gpt-oss-120b:free",     // Primary: Free and capable
	"openai/gpt-oss-120b",          // Fallback 1: Strong performance
	"google/gemini-2.5-flash-lite", // Fallback 2: Fast and capable
}

func Path() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".gitcomm", "config.json"), nil
}

type Config struct {
	OpenRouterAPIKey string   `json:"open_router_api_key"`
	Models           []string `json:"models,omitempty"`
	MaxTokens        int      `json:"max_tokens,omitempty"`
	Temperature      float64  `json:"temperature,omitempty"`
	APIURL           string   `json:"api_url,omitempty"`
	TimeoutSeconds   int      `json:"timeout_seconds,omitempty"`
}

func DefaultConfig() *Config {
	models := make([]string, len(DefaultModels))
	copy(models, DefaultModels)

	return &Config{
		Models:         models,
		MaxTokens:      DefaultMaxTokens,
		Temperature:    DefaultTemperature,
		APIURL:         OpenRouterAPIURL,
		TimeoutSeconds: DefaultTimeoutSeconds,
	}
}

func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	configPath, err := Path()
	if err != nil {
		return nil, err
	}

	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, config)
	}

	// Environment variables override config file (primary then legacy)
	if envKey := os.Getenv(OpenRouterAPIKeyEnvPrimary); envKey != "" {
		config.OpenRouterAPIKey = envKey
	} else if envKey := os.Getenv(OpenRouterAPIKeyEnvLegacy); envKey != "" {
		config.OpenRouterAPIKey = envKey
	}

	config.Models = normalizeModels(config.Models)
	if config.MaxTokens < 0 {
		config.MaxTokens = 0
	}
	if config.Temperature < 0 {
		config.Temperature = 0
	}
	if config.TimeoutSeconds < 0 {
		config.TimeoutSeconds = 0
	}

	return config, nil
}

func SaveConfig(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".gitcomm")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(configDir, "config.json"), data, 0600)
}

func normalizeModels(models []string) []string {
	if len(models) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(models))
	for _, model := range models {
		trimmed := strings.TrimSpace(model)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}

	return normalized
}
