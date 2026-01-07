package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
)

// SetupUI provides styled output for setup commands
type SetupUI struct {
	out    io.Writer
	styles Styles
}

// NewSetupUI creates a new setup UI writer
func NewSetupUI(out io.Writer) *SetupUI {
	return &SetupUI{
		out:    out,
		styles: DefaultStyles(),
	}
}

// Header prints a styled section header
func (s *SetupUI) Header(text string) {
	style := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		MarginTop(1)
	fmt.Fprintln(s.out, style.Render(text))
}

// SubHeader prints a styled sub-section header
func (s *SetupUI) SubHeader(text string) {
	style := lipgloss.NewStyle().
		Foreground(colorTextDim).
		MarginLeft(2)
	fmt.Fprintln(s.out, style.Render(text))
}

// Step prints a step description
func (s *SetupUI) Step(text string) {
	style := lipgloss.NewStyle().
		Foreground(colorText)
	fmt.Fprintln(s.out, style.Render(text))
}

// Success prints a success message with checkmark
func (s *SetupUI) Success(text string) {
	icon := lipgloss.NewStyle().Foreground(colorSuccess).Render(IconSuccess)
	msg := lipgloss.NewStyle().Foreground(colorText).Render(text)
	fmt.Fprintf(s.out, "  %s %s\n", icon, msg)
}

// SuccessPath prints a success message with a path
func (s *SetupUI) SuccessPath(text, path string) {
	icon := lipgloss.NewStyle().Foreground(colorSuccess).Render(IconSuccess)
	msg := lipgloss.NewStyle().Foreground(colorText).Render(text)
	pathStyle := lipgloss.NewStyle().Foreground(colorSecondary).Render(path)
	fmt.Fprintf(s.out, "  %s %s %s\n", icon, msg, pathStyle)
}

// Warning prints a warning message
func (s *SetupUI) Warning(text string) {
	icon := lipgloss.NewStyle().Foreground(colorTertiary).Render(IconWarning)
	msg := lipgloss.NewStyle().Foreground(colorTertiary).Render(text)
	fmt.Fprintf(s.out, "  %s %s\n", icon, msg)
}

// WarningDetail prints a warning detail item
func (s *SetupUI) WarningDetail(text string) {
	style := lipgloss.NewStyle().Foreground(colorMuted).MarginLeft(6)
	fmt.Fprintln(s.out, style.Render("- "+text))
}

// Error prints an error message
func (s *SetupUI) Error(text string) {
	icon := lipgloss.NewStyle().Foreground(colorError).Render(IconError)
	msg := lipgloss.NewStyle().Foreground(colorError).Render(text)
	fmt.Fprintf(s.out, "  %s %s\n", icon, msg)
}

// Info prints an info message
func (s *SetupUI) Info(text string) {
	style := lipgloss.NewStyle().Foreground(colorMuted).MarginLeft(2)
	fmt.Fprintln(s.out, style.Render(text))
}

// Code prints a code block (e.g., for shell commands)
func (s *SetupUI) Code(text string) {
	style := lipgloss.NewStyle().
		Foreground(colorSuccess).
		Background(colorBgDark).
		Padding(0, 1).
		MarginLeft(4)
	fmt.Fprintln(s.out, style.Render(text))
}

// Complete prints the final completion message
func (s *SetupUI) Complete(text string) {
	style := lipgloss.NewStyle().
		Foreground(colorSuccess).
		Bold(true).
		MarginTop(1)
	fmt.Fprintln(s.out, style.Render(IconSuccess+" "+text))
}

// Cancelled prints a cancellation message
func (s *SetupUI) Cancelled(text string) {
	style := lipgloss.NewStyle().
		Foreground(colorMuted).
		MarginTop(1)
	fmt.Fprintln(s.out, style.Render(text))
}

// Blank prints a blank line
func (s *SetupUI) Blank() {
	fmt.Fprintln(s.out)
}

// Divider prints a subtle divider line
func (s *SetupUI) Divider() {
	style := lipgloss.NewStyle().Foreground(colorSubtle)
	fmt.Fprintln(s.out, style.Render("────────────────────────────────────────"))
}
