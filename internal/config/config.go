package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	OpenRouterAPIKeyEnvPrimary = "OPENROUTER_API_KEY"
	OpenRouterAPIKeyEnvLegacy  = "OPEN_ROUTER_API_KEY"
	OpenRouterAPIURL           = "https://openrouter.ai/api/v1/chat/completions"
	DefaultMaxTokens           = 400
	DefaultTemperature         = 0.7
	DefaultTimeoutSeconds      = 30
	MaxModelNameLength         = 255
)

var (
	DefaultModels = []string{
		"openai/gpt-oss-120b:free",     // Primary: Free and capable
		"openai/gpt-oss-120b",          // Fallback 1: Strong performance
		"google/gemini-2.5-flash-lite", // Fallback 2: Fast and capable
	}
	
	// ModelNameRegex validates provider/model format
	// Provider name: alphanumeric, underscore, hyphen
	// Model name: alphanumeric, period, underscore, hyphen, colon
	// Separated by a single slash
	ModelNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+/[a-zA-Z0-9._:-]+$`)
)

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
	
	// Validate model names for security
	validatedModels := make([]string, 0, len(config.Models))
	for _, model := range config.Models {
		if err := ValidateModelName(model); err != nil {
			// Log validation error but continue with default models
			// This prevents DoS attacks via malformed config files
			continue
		}
		validatedModels = append(validatedModels, model)
	}
	
	// If all models were invalid, use defaults
	if len(validatedModels) == 0 {
		validatedModels = DefaultModels
	}
	config.Models = validatedModels
	
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

	// Validate all model names before saving
	for _, model := range config.Models {
		if err := ValidateModelName(model); err != nil {
			return fmt.Errorf("invalid model name %q: %w", model, err)
		}
	}

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")
	
	// Use atomic write: write to temp file first, then rename
	tempPath := configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary config file: %w", err)
	}
	
	// Rename temp file to final config file (atomic operation)
	if err := os.Rename(tempPath, configPath); err != nil {
		// Clean up temp file if rename fails
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temporary config file: %w", err)
	}

	return nil
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

// ValidateModelName checks if a model name is valid according to security requirements
func ValidateModelName(model string) error {
	if model == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	// Check length limit to prevent DoS via large strings
	if len(model) > MaxModelNameLength {
		return fmt.Errorf("model name exceeds maximum length of %d characters", MaxModelNameLength)
	}

	// Check for slash separator (basic validation for better error messages)
	if !strings.Contains(model, "/") {
		return fmt.Errorf("model name must be in provider/model format (e.g., 'openai/gpt-4o-mini')")
	}

	// Apply stricter regex validation
	if !ModelNameRegex.MatchString(model) {
		return fmt.Errorf("model name must match format: provider/model-name (alphanumeric, underscore, hyphen, period, colon)")
	}

	return nil
}
