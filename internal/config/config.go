package config

import (
    "os"
    "path/filepath"
    "encoding/json"
)

const (
    OpenRouterAPIKeyEnv = "OPEN_ROUTER_API_KEY"
    OpenRouterAPIURL    = "https://openrouter.ai/api/v1/chat/completions"
)

type Config struct {
    OpenRouterAPIKey string `json:"open_router_api_key"`
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
