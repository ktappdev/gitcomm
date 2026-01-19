package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const (
	OpenRouterAPIKeyEnv = "OPEN_ROUTER_API_KEY"
	OpenRouterAPIURL    = "https://openrouter.ai/api/v1/chat/completions"
)

type Config struct {
	OpenRouterAPIKey string   `json:"open_router_api_key"`
	Models           []string `json:"models,omitempty"`
	MaxTokens        int      `json:"max_tokens,omitempty"`
	Temperature      float64  `json:"temperature,omitempty"`
	APIURL           string   `json:"api_url,omitempty"`
	TimeoutSeconds   int      `json:"timeout_seconds,omitempty"`
}

func LoadConfig() (*Config, error) {
	// Try loading from config file first
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".gitcomm", "config.json")
	config := &Config{}

	// Check if config file exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err == nil {
			json.Unmarshal(data, config)
		}
	}

	// Environment variables override config file
	if envKey := os.Getenv(OpenRouterAPIKeyEnv); envKey != "" {
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
