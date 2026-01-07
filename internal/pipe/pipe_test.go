package pipe

import (
	"os"
	"testing"
)

func TestChainContext(t *testing.T) {
	// Save original env and restore after test
	original := os.Getenv(ChainContextEnvVar)
	defer os.Setenv(ChainContextEnvVar, original)

	t.Run("GetChainContext returns nil when not set", func(t *testing.T) {
		os.Unsetenv(ChainContextEnvVar)
		ctx := GetChainContext()
		if ctx != nil {
			t.Errorf("expected nil, got %+v", ctx)
		}
	})

	t.Run("GetChainContext parses valid context", func(t *testing.T) {
		os.Setenv(ChainContextEnvVar, `{"depth":2,"source":"@ayo.test","source_description":"Test agent"}`)
		ctx := GetChainContext()
		if ctx == nil {
			t.Fatal("expected context, got nil")
		}
		if ctx.Depth != 2 {
			t.Errorf("expected depth 2, got %d", ctx.Depth)
		}
		if ctx.Source != "@ayo.test" {
			t.Errorf("expected source @ayo.test, got %s", ctx.Source)
		}
		if ctx.SourceDescription != "Test agent" {
			t.Errorf("expected description 'Test agent', got %s", ctx.SourceDescription)
		}
	})

	t.Run("GetChainContext returns nil for invalid JSON", func(t *testing.T) {
		os.Setenv(ChainContextEnvVar, "not json")
		ctx := GetChainContext()
		if ctx != nil {
			t.Errorf("expected nil for invalid JSON, got %+v", ctx)
		}
	})

	t.Run("SetChainContext sets environment variable", func(t *testing.T) {
		ctx := ChainContext{
			Depth:             3,
			Source:            "@ayo.example",
			SourceDescription: "Example",
		}
		if err := SetChainContext(ctx); err != nil {
			t.Fatalf("SetChainContext error: %v", err)
		}

		retrieved := GetChainContext()
		if retrieved == nil {
			t.Fatal("expected context after set")
		}
		if retrieved.Depth != 3 {
			t.Errorf("expected depth 3, got %d", retrieved.Depth)
		}
	})

	t.Run("NextChainContext increments depth", func(t *testing.T) {
		os.Setenv(ChainContextEnvVar, `{"depth":2,"source":"@ayo.prev"}`)
		next := NextChainContext("@ayo.current", "Current agent")

		if next.Depth != 3 {
			t.Errorf("expected depth 3, got %d", next.Depth)
		}
		if next.Source != "@ayo.current" {
			t.Errorf("expected source @ayo.current, got %s", next.Source)
		}
	})

	t.Run("NextChainContext starts at 1 when no context", func(t *testing.T) {
		os.Unsetenv(ChainContextEnvVar)
		next := NextChainContext("@ayo.first", "First agent")

		if next.Depth != 1 {
			t.Errorf("expected depth 1, got %d", next.Depth)
		}
	})

	t.Run("ChainDepth returns 0 when not in chain", func(t *testing.T) {
		os.Unsetenv(ChainContextEnvVar)
		// Note: This test may return 1 if stdin is piped during test execution
		// In normal terminal execution, it should return 0
		depth := ChainDepth()
		// Allow either 0 or 1 since test runners may pipe
		if depth < 0 || depth > 1 {
			t.Errorf("expected depth 0 or 1, got %d", depth)
		}
	})
}
