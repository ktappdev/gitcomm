package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ktappdev/gitcomm/internal/diag"
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
		"openrouter/free",
		"openrouter/free",
		"google/gemini-2.5-flash-lite",
	}

	ModelNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+/[a-zA-Z0-9._:-]+$`)
)

type Config struct {
	OpenRouterAPIKey string   `json:"open_router_api_key"`
	Models           []string `json:"models,omitempty"`
	MaxTokens        int      `json:"max_tokens,omitempty"`
	Temperature      float64  `json:"temperature,omitempty"`
	APIURL           string   `json:"api_url,omitempty"`
	TimeoutSeconds   int      `json:"timeout_seconds,omitempty"`
}

func Dir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".gitcomm"), nil
}

func Path() (string, error) {
	configDir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
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
	cfg := DefaultConfig()
	configPath, err := Path()
	if err != nil {
		return nil, err
	}

	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			diag.Error("config", "failed to parse config file", "path", configPath, "error", err, "size_bytes", len(data))
			return nil, fmt.Errorf("invalid config file %s: %w", configPath, err)
		}
		diag.Debug("config", "loaded config file", "path", configPath, "size_bytes", len(data))
	} else if !os.IsNotExist(err) {
		diag.Error("config", "failed to read config file", "path", configPath, "error", err)
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	applyEnvOverrides(cfg)
	normalizeRuntimeConfig(cfg)
	return cfg, nil
}

// LoadRuntimeConfig returns a usable runtime config even when LoadConfig fails.
//
// It preserves the original load error so callers can log or surface the warning,
// while still proceeding with defaults plus environment overrides when safe.
func LoadRuntimeConfig() (*Config, error) {
	cfg, err := LoadConfig()
	if err == nil {
		return cfg, nil
	}

	fallback := DefaultConfig()
	applyEnvOverrides(fallback)
	normalizeRuntimeConfig(fallback)
	diag.Warn("config", "using runtime fallback config after load failure", "error", err, "has_env_api_key", fallback.OpenRouterAPIKey != "")
	return fallback, err
}

func SaveConfig(config *Config) error {
	configDir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}

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
	tempPath := configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write temporary config file: %w", err)
	}
	if err := os.Rename(tempPath, configPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temporary config file: %w", err)
	}

	diag.Info("config", "saved config file", "path", configPath, "models_count", len(config.Models))
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

func normalizeRuntimeConfig(cfg *Config) {
	cfg.Models = normalizeModels(cfg.Models)
	validatedModels := make([]string, 0, len(cfg.Models))
	for _, model := range cfg.Models {
		if err := ValidateModelName(model); err != nil {
			diag.Warn("config", "ignoring invalid configured model", "model", model, "error", err)
			continue
		}
		validatedModels = append(validatedModels, model)
	}
	if len(validatedModels) == 0 {
		validatedModels = append([]string(nil), DefaultModels...)
		if len(cfg.Models) > 0 {
			diag.Warn("config", "all configured models invalid; using defaults", "defaults_count", len(validatedModels))
		}
	}
	cfg.Models = validatedModels

	if cfg.MaxTokens < 0 {
		diag.Warn("config", "negative max_tokens reset to zero", "value", cfg.MaxTokens)
		cfg.MaxTokens = 0
	}
	if cfg.Temperature < 0 {
		diag.Warn("config", "negative temperature reset to zero", "value", cfg.Temperature)
		cfg.Temperature = 0
	}
	if cfg.TimeoutSeconds < 0 {
		diag.Warn("config", "negative timeout reset to zero", "value", cfg.TimeoutSeconds)
		cfg.TimeoutSeconds = 0
	}
}

func applyEnvOverrides(cfg *Config) {
	if envKey := os.Getenv(OpenRouterAPIKeyEnvPrimary); envKey != "" {
		cfg.OpenRouterAPIKey = envKey
	} else if envKey := os.Getenv(OpenRouterAPIKeyEnvLegacy); envKey != "" {
		cfg.OpenRouterAPIKey = envKey
	}
}

func ValidateModelName(model string) error {
	if model == "" {
		return fmt.Errorf("model name cannot be empty")
	}
	if len(model) > MaxModelNameLength {
		return fmt.Errorf("model name exceeds maximum length of %d characters", MaxModelNameLength)
	}
	if !strings.Contains(model, "/") {
		return fmt.Errorf("model name must be in provider/model format (e.g., 'openai/gpt-4o-mini')")
	}
	if !ModelNameRegex.MatchString(model) {
		return fmt.Errorf("model name must match format: provider/model-name (alphanumeric, underscore, hyphen, period, colon)")
	}
	return nil
}
