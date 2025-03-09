package config

import (
    "os"
    "path/filepath"
    "encoding/json"
)

const (
    GeminiAPIKeyEnv = "GEMINI_API_KEY"
    GroqAPIKeyEnv   = "GROQ_API_KEY"
    OpenAIAPIKeyEnv = "OPENAI_API_KEY"
    OpenAIAPIURL    = "https://api.openai.com/v1/chat/completions"
    GroqAPIURL      = "https://api.groq.com/openai/v1/chat/completions"
)

type Config struct {
    GeminiAPIKey string `json:"gemini_api_key"`
    GroqAPIKey   string `json:"groq_api_key"`
    OpenAIAPIKey string `json:"openai_api_key"`
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
    if envKey := os.Getenv(GeminiAPIKeyEnv); envKey != "" {
        config.GeminiAPIKey = envKey
    }
    if envKey := os.Getenv(GroqAPIKeyEnv); envKey != "" {
        config.GroqAPIKey = envKey
    }
    if envKey := os.Getenv(OpenAIAPIKeyEnv); envKey != "" {
        config.OpenAIAPIKey = envKey
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
