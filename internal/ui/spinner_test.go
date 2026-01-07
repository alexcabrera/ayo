package ui

import (
	"testing"
)

func TestSpinnerFrames(t *testing.T) {
	if len(SpinnerFrames) == 0 {
		t.Error("SpinnerFrames should not be empty")
	}
}

func TestNewSpinner(t *testing.T) {
	s := NewSpinner("test message")
	if s == nil {
		t.Fatal("NewSpinner returned nil")
	}
	if s.message != "test message" {
		t.Errorf("message = %q, want %q", s.message, "test message")
	}
}

func TestNewQuietSpinner(t *testing.T) {
	s := NewQuietSpinner()
	if s == nil {
		t.Fatal("NewQuietSpinner returned nil")
	}
	if !s.quiet {
		t.Error("quiet spinner should have quiet=true")
	}
}

func TestNewSpinnerWithDepth(t *testing.T) {
	// Depth 0 should have no indent
	s0 := NewSpinnerWithDepth("test", 0)
	if s0.indent != "" {
		t.Errorf("depth 0 indent = %q, want empty", s0.indent)
	}

	// Depth 1 should have indent
	s1 := NewSpinnerWithDepth("test", 1)
	if s1.indent == "" {
		t.Error("depth 1 should have indent")
	}

	// Depth 2 should have more indent
	s2 := NewSpinnerWithDepth("test", 2)
	if len(s2.indent) <= len(s1.indent) {
		t.Error("depth 2 indent should be longer than depth 1")
	}
}
