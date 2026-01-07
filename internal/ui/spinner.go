package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// SpinnerFrames defines the animation frames for the spinner
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner displays an animated spinner with a message
type Spinner struct {
	message   string
	indent    string // Indentation prefix for nested spinners
	startTime time.Time
	done      chan struct{}
	closeOnce sync.Once
	mu        sync.Mutex
	frame     int
	style     lipgloss.Style
	msgStyle  lipgloss.Style
	isTTY     bool
	quiet     bool
}

// NewSpinner creates a new spinner with the given message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message:   message,
		indent:    "",
		startTime: time.Now(),
		done:      make(chan struct{}),
		style:     lipgloss.NewStyle().Foreground(colorPrimary),
		msgStyle:  lipgloss.NewStyle().Foreground(colorTextDim),
		isTTY:     term.IsTerminal(int(os.Stderr.Fd())),
	}
}

// NewSpinnerWithDepth creates a spinner at a specific nesting depth.
// Depth 0 is top-level (no indent), depth 1+ adds visual indentation.
func NewSpinnerWithDepth(message string, depth int) *Spinner {
	indent := ""
	style := lipgloss.NewStyle().Foreground(colorPrimary)
	msgStyle := lipgloss.NewStyle().Foreground(colorTextDim)

	if depth > 0 {
		// Use vertical line indicator and muted colors for nested spinners
		indent = strings.Repeat("  ", depth) + "│ "
		style = lipgloss.NewStyle().Foreground(colorSecondary)
		msgStyle = lipgloss.NewStyle().Foreground(colorMuted)
	}

	return &Spinner{
		message:   message,
		indent:    indent,
		startTime: time.Now(),
		done:      make(chan struct{}),
		style:     style,
		msgStyle:  msgStyle,
		isTTY:     term.IsTerminal(int(os.Stderr.Fd())),
	}
}

// NewQuietSpinner creates a spinner that produces no output
func NewQuietSpinner() *Spinner {
	return &Spinner{
		startTime: time.Now(),
		done:      make(chan struct{}),
		quiet:     true,
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	if s.quiet {
		return
	}
	if !s.isTTY {
		fmt.Fprintf(os.Stderr, "%s%s %s\n", s.indent, SpinnerFrames[0], s.message)
		return
	}

	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.done:
				return
			case <-ticker.C:
				s.mu.Lock()
				s.frame = (s.frame + 1) % len(SpinnerFrames)
				s.render()
				s.mu.Unlock()
			}
		}
	}()

	// Initial render
	s.mu.Lock()
	s.render()
	s.mu.Unlock()
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	s.closeOnce.Do(func() { close(s.done) })

	if s.quiet {
		return
	}

	if s.isTTY {
		fmt.Fprint(os.Stderr, "\r\033[K") // Clear line
	}
}

// StopWithMessage stops the spinner and shows a final message with elapsed time
func (s *Spinner) StopWithMessage(message string) {
	s.closeOnce.Do(func() { close(s.done) })

	if s.quiet {
		return
	}

	elapsed := time.Since(s.startTime)
	elapsedStr := formatDuration(elapsed)

	if s.isTTY {
		fmt.Fprint(os.Stderr, "\r\033[K")
	}

	// Show "Thought for Xs" style message
	msgStyle := lipgloss.NewStyle().Foreground(colorTextDim)
	fmt.Fprintf(os.Stderr, "%s%s %s\n", s.indent, msgStyle.Render(message), msgStyle.Render("("+elapsedStr+")"))
}

// StopWithError stops the spinner and shows an error indicator
func (s *Spinner) StopWithError(message string) {
	s.closeOnce.Do(func() { close(s.done) })

	if s.quiet {
		return
	}

	if s.isTTY {
		fmt.Fprint(os.Stderr, "\r\033[K")
	}

	errorStyle := lipgloss.NewStyle().Foreground(colorError)
	fmt.Fprintf(os.Stderr, "%s%s %s\n", s.indent, errorStyle.Render(IconError), message)
}

// Elapsed returns the time elapsed since the spinner started
func (s *Spinner) Elapsed() time.Duration {
	return time.Since(s.startTime)
}

// render draws the current spinner state (must be called with lock held)
func (s *Spinner) render() {
	if s.quiet {
		return
	}
	frame := s.style.Render(SpinnerFrames[s.frame])
	msg := s.msgStyle.Render(s.message)
	fmt.Fprintf(os.Stderr, "\r\033[K%s%s %s", s.indent, frame, msg)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}
