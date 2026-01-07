package ui

import (
	"bytes"
	"testing"
)

func TestCleanText(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "strips ansi escape sequences",
			in:   "\x1b[38;5;252mhello\x1b[0m",
			out:  "hello",
		},
		{
			name: "unescapes newline",
			in:   "line1\\nline2",
			out:  "line1\nline2",
		},
		{
			name: "unescapes tab",
			in:   "a\\tb",
			out:  "a\tb",
		},
		{
			name: "passes through plain text",
			in:   "plain text",
			out:  "plain text",
		},
	}

	for _, tc := range cases {
		got := cleanText(tc.in)
		if got != tc.out {
			t.Fatalf("%s: expected %q got %q", tc.name, tc.out, got)
		}
	}
}

func TestUIDepth(t *testing.T) {
	// Test depth 0 (top-level)
	ui0 := NewWithDepth(false, 0)
	if ui0.Depth() != 0 {
		t.Errorf("Depth() = %d, want 0", ui0.Depth())
	}
	if ui0.IsNested() {
		t.Error("depth 0 should not be nested")
	}
	if ui0.indent() != "" {
		t.Errorf("depth 0 indent = %q, want empty", ui0.indent())
	}

	// Test depth 1 (first level sub-agent)
	ui1 := NewWithDepth(false, 1)
	if ui1.Depth() != 1 {
		t.Errorf("Depth() = %d, want 1", ui1.Depth())
	}
	if !ui1.IsNested() {
		t.Error("depth 1 should be nested")
	}
	if ui1.indent() == "" {
		t.Error("depth 1 should have indent")
	}

	// Test depth 2 (nested sub-agent)
	ui2 := NewWithDepth(false, 2)
	if len(ui2.indent()) <= len(ui1.indent()) {
		t.Error("depth 2 indent should be longer than depth 1")
	}
}

func TestUIIndentLines(t *testing.T) {
	ui := NewWithDepth(false, 1)
	indent := ui.indent()

	// Single line
	result := ui.indentLines("hello")
	expected := indent + "hello"
	if result != expected {
		t.Errorf("indentLines single = %q, want %q", result, expected)
	}

	// Multiple lines
	result = ui.indentLines("line1\nline2\nline3")
	expected = indent + "line1\n" + indent + "line2\n" + indent + "line3"
	if result != expected {
		t.Errorf("indentLines multi = %q, want %q", result, expected)
	}
}

func TestUIWritesToBuffer(t *testing.T) {
	var buf bytes.Buffer
	ui := NewWithWriter(false, &buf)

	ui.println("test output")

	if buf.String() != "test output\n" {
		t.Errorf("output = %q, want %q", buf.String(), "test output\n")
	}
}
