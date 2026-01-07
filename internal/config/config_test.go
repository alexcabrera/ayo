package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ayo/internal/paths"
)

func TestDefaultPaths(t *testing.T) {
	cfg := Default()

	// Verify paths use the platform-specific data directory
	if cfg.AgentsDir != paths.AgentsDir() {
		t.Fatalf("agents dir mismatch: got %s, want %s", cfg.AgentsDir, paths.AgentsDir())
	}
	if cfg.SkillsDir != paths.SkillsDir() {
		t.Fatalf("skills dir mismatch: got %s, want %s", cfg.SkillsDir, paths.SkillsDir())
	}

	// System prompts should be in the prompts directory
	promptsDir := paths.SystemPromptsDir()
	if cfg.SharedSystemMessage != filepath.Join(promptsDir, "system.md") {
		t.Fatalf("shared system path mismatch: got %s", cfg.SharedSystemMessage)
	}
	if cfg.SystemPrefix != filepath.Join(promptsDir, "prefix.md") {
		t.Fatalf("prefix path mismatch: got %s", cfg.SystemPrefix)
	}
	if cfg.SystemSuffix != filepath.Join(promptsDir, "suffix.md") {
		t.Fatalf("suffix path mismatch: got %s", cfg.SystemSuffix)
	}

	// All paths should contain "ayo"
	if !strings.Contains(cfg.AgentsDir, "ayo") {
		t.Fatalf("agents dir should contain 'ayo': %s", cfg.AgentsDir)
	}
}

func TestDefaultCatwalkURLFromEnv(t *testing.T) {
	t.Setenv("CATWALK_URL", "https://catwalk.example")
	cfg := Default()
	if cfg.CatwalkBaseURL != "https://catwalk.example" {
		t.Fatalf("expected catwalk base URL from env, got %q", cfg.CatwalkBaseURL)
	}
}

func TestDefaultCatwalkURLFallback(t *testing.T) {
	t.Setenv("CATWALK_URL", "")
	cfg := Default()
	if cfg.CatwalkBaseURL == "" {
		t.Fatalf("expected default catwalk URL to be set")
	}
}

func mustUserHome(t *testing.T) string {
	t.Helper()
	h, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("home: %v", err)
	}
	return h
}
