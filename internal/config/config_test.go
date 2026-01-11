package config

import (
	"os"
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

	// System prompts are now resolved at load time via paths.FindPromptFile
	// Default config has empty strings for SystemPrefix and SystemSuffix
	if cfg.SystemPrefix != "" {
		t.Fatalf("expected empty SystemPrefix, got %s", cfg.SystemPrefix)
	}
	if cfg.SystemSuffix != "" {
		t.Fatalf("expected empty SystemSuffix, got %s", cfg.SystemSuffix)
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
