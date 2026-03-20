package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigReturnsParseError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".gitcomm"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".gitcomm", "config.json")
	if err := os.WriteFile(path, []byte(`{"models": [`), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "invalid config file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadRuntimeConfigUsesEnvOnInvalidConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(OpenRouterAPIKeyEnvPrimary, "env-key")
	if err := os.MkdirAll(filepath.Join(home, ".gitcomm"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".gitcomm", "config.json")
	if err := os.WriteFile(path, []byte(`{"models": [`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadRuntimeConfig()
	if err == nil {
		t.Fatal("expected warning error from invalid config")
	}
	if cfg.OpenRouterAPIKey != "env-key" {
		t.Fatalf("expected env api key, got %q", cfg.OpenRouterAPIKey)
	}
	if len(cfg.Models) != len(DefaultModels) {
		t.Fatalf("expected default models, got %v", cfg.Models)
	}
}

func TestLoadConfigFallsBackWhenModelsInvalid(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".gitcomm"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".gitcomm", "config.json")
	content := `{"open_router_api_key":"test","models":["bad model"," ","also-bad"]}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if len(cfg.Models) != len(DefaultModels) {
		t.Fatalf("expected default models, got %v", cfg.Models)
	}
}
