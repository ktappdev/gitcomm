package llm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ktappdev/gitcomm/internal/config"
)

func TestNewClientUsesRuntimeFallbackConfigWithEnv(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(config.OpenRouterAPIKeyEnvPrimary, "env-key")
	configDir := filepath.Join(home, ".gitcomm")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"models": [`), 0o600); err != nil {
		t.Fatal(err)
	}

	client, err := NewClient(ClientConfig{})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client.apiKey != "env-key" {
		t.Fatalf("expected env api key, got %q", client.apiKey)
	}
	if len(client.models) != len(config.DefaultModels) {
		t.Fatalf("expected default models, got %v", client.models)
	}
}

func TestFormatAPIErrorPreservesProviderReason(t *testing.T) {
	err := formatAPIError("meta-llama/llama-3.3-8b-instruct:free", 400, []byte(`{"error":{"message":"prompt is too long for this model"}}`))
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "prompt is too long for this model") {
		t.Fatalf("provider message missing: %v", err)
	}
	if !strings.Contains(msg, "too large or malformed") {
		t.Fatalf("size guidance missing: %v", err)
	}
}

func TestFormatAPIErrorPaymentRequiredPreservesProviderReason(t *testing.T) {
	err := formatAPIError("meta-llama/llama-4-scout", 402, []byte(`{"error":{"message":"insufficient credits"}}`))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "insufficient credits") {
		t.Fatalf("unexpected error: %v", err)
	}
}
