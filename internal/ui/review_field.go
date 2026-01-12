package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// ReviewSection represents a section in the review field.
type ReviewSection struct {
	Label      string
	Value      string
	Expandable bool
	Expanded   bool
	Content    string // Full content for expandable sections
}

// ReviewDataProvider is a function that returns the current sections to display.
type ReviewDataProvider func() []ReviewSection

// ReviewField is a custom huh field for reviewing all collected form data.
// It supports accordion-style expandable sections where only one can be open at a time.
type ReviewField struct {
	dataProvider   ReviewDataProvider
	sections       []ReviewSection
	selectedIdx    int
	viewport       viewport.Model
	terminalHeight int

	// Caching for performance
	cachedContent string // Cached rendered content
	contentDirty  bool   // Whether content needs re-rendering
	styles        *reviewStyles

	focused bool
	width   int
	height  int
	theme   *huh.Theme
	keymap  reviewKeyMap
}

// reviewStyles holds cached lipgloss styles
type reviewStyles struct {
	label           lipgloss.Style
	value           lipgloss.Style
	expandHint      lipgloss.Style
	selected        lipgloss.Style
	expandedContent lipgloss.Style
}

type reviewKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Toggle   key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Next     key.Binding
	Prev     key.Binding
}

// NewReviewField creates a new review field.
func NewReviewField() *ReviewField {
	vp := viewport.New(80, 20)
	vp.MouseWheelEnabled = true

	return &ReviewField{
		sections:     []ReviewSection{},
		viewport:     vp,
		height:       20,
		contentDirty: true,
		keymap: reviewKeyMap{
			Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
			Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
			Toggle:   key.NewBinding(key.WithKeys("enter", " "), key.WithHelp("enter/space", "expand")),
			PageUp:   key.NewBinding(key.WithKeys("pgup", "ctrl+u"), key.WithHelp("pgup", "page up")),
			PageDown: key.NewBinding(key.WithKeys("pgdown", "ctrl+d"), key.WithHelp("pgdn", "page down")),
			Next:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "continue")),
			Prev:     key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "back")),
		},
	}
}

// DataProvider sets the function that provides review sections dynamically.
func (f *ReviewField) DataProvider(provider ReviewDataProvider) *ReviewField {
	f.dataProvider = provider
	return f
}

// Height sets the height of the review field.
func (f *ReviewField) Height(height int) *ReviewField {
	f.height = height
	viewportHeight := height - 2
	if viewportHeight < 5 {
		viewportHeight = 5
	}
	f.viewport.Height = viewportHeight
	return f
}

// loadSections refreshes sections from the data provider, preserving expand state.
func (f *ReviewField) loadSections() {
	if f.dataProvider == nil {
		return
	}

	// Remember which label was expanded
	var expandedLabel string
	for _, s := range f.sections {
		if s.Expanded {
			expandedLabel = s.Label
			break
		}
	}

	// Get fresh sections
	f.sections = f.dataProvider()

	// Restore expanded state
	if expandedLabel != "" {
		for i := range f.sections {
			if f.sections[i].Label == expandedLabel && f.sections[i].Expandable {
				f.sections[i].Expanded = true
				break
			}
		}
	}

	// Ensure selectedIdx is valid and points to an expandable section
	f.selectedIdx = -1
	for i, s := range f.sections {
		if s.Expandable {
			f.selectedIdx = i
			break
		}
	}

	f.contentDirty = true
}

// ensureStyles initializes cached styles if needed.
func (f *ReviewField) ensureStyles() {
	if f.styles != nil {
		return
	}
	f.styles = &reviewStyles{
		label: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}).
			Bold(true).
			Width(14),
		value: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FAFAFA"}),
		expandHint: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#94A3B8"}),
		selected: lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#E0E7FF", Dark: "#3730A3"}),
		expandedContent: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#374151", Dark: "#D1D5DB"}).
			MarginLeft(2).
			PaddingLeft(2).
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#A78BFA", Dark: "#7C3AED"}),
	}
}

// renderContent builds the viewport content.
func (f *ReviewField) renderContent() string {
	f.ensureStyles()

	var sb strings.Builder

	for i, section := range f.sections {
		// Build the line
		label := f.styles.label.Render(section.Label + ":")
		value := f.styles.value.Render(section.Value)

		line := label + " " + value

		if section.Expandable {
			if section.Expanded {
				line += " " + f.styles.expandHint.Render("▼")
			} else {
				line += " " + f.styles.expandHint.Render("▶")
			}
		}

		// Highlight selected row
		if f.focused && i == f.selectedIdx && section.Expandable {
			line = f.styles.selected.Render(line)
		}

		sb.WriteString(line)
		sb.WriteString("\n")

		// Show expanded content
		if section.Expandable && section.Expanded && section.Content != "" {
			content := section.Content
			// Render markdown if it looks like markdown - use fast shared renderer
			if strings.Contains(content, "#") || strings.Contains(content, "```") || strings.Contains(content, "**") || strings.Contains(content, "- ") {
				if r := getFastRenderer(); r != nil {
					rendered, err := r.Render(content)
					if err == nil {
						content = strings.TrimSpace(rendered)
					}
				}
			}
			sb.WriteString(f.styles.expandedContent.Render(content))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// Init initializes the field.
func (f *ReviewField) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (f *ReviewField) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.terminalHeight = msg.Height
		newWidth := msg.Width - 4
		// Only mark dirty if width changed significantly
		if f.viewport.Width > 0 {
			diff := f.viewport.Width - newWidth
			if diff < 0 {
				diff = -diff
			}
			if diff > 5 {
				f.contentDirty = true
			}
		}
		f.viewport.Width = newWidth
		viewportHeight := msg.Height - 10
		if viewportHeight < 5 {
			viewportHeight = 5
		}
		if viewportHeight > f.height {
			viewportHeight = f.height
		}
		f.viewport.Height = viewportHeight
		// Only re-render if dirty
		if f.contentDirty && f.focused {
			f.viewport.SetContent(f.renderContent())
			f.contentDirty = false
		}
	}

	if !f.focused {
		return f, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keymap.Next):
			return f, huh.NextField
		case key.Matches(msg, f.keymap.Prev):
			return f, huh.PrevField
		case key.Matches(msg, f.keymap.Up):
			oldIdx := f.selectedIdx
			f.moveToPrevExpandable()
			if oldIdx != f.selectedIdx {
				f.viewport.SetContent(f.renderContent())
			}
			return f, nil
		case key.Matches(msg, f.keymap.Down):
			oldIdx := f.selectedIdx
			f.moveToNextExpandable()
			if oldIdx != f.selectedIdx {
				f.viewport.SetContent(f.renderContent())
			}
			return f, nil
		case key.Matches(msg, f.keymap.Toggle):
			f.toggleSelected()
			f.viewport.SetContent(f.renderContent())
			return f, nil
		case key.Matches(msg, f.keymap.PageUp):
			f.viewport.ViewUp()
			return f, nil
		case key.Matches(msg, f.keymap.PageDown):
			f.viewport.ViewDown()
			return f, nil
		}
	}

	var cmd tea.Cmd
	f.viewport, cmd = f.viewport.Update(msg)
	return f, cmd
}

// moveToPrevExpandable moves selection to previous expandable section.
func (f *ReviewField) moveToPrevExpandable() {
	for i := f.selectedIdx - 1; i >= 0; i-- {
		if f.sections[i].Expandable {
			f.selectedIdx = i
			return
		}
	}
}

// moveToNextExpandable moves selection to next expandable section.
func (f *ReviewField) moveToNextExpandable() {
	for i := f.selectedIdx + 1; i < len(f.sections); i++ {
		if f.sections[i].Expandable {
			f.selectedIdx = i
			return
		}
	}
}

// toggleSelected toggles the expanded state of the selected section.
// Implements accordion behavior: only one section can be expanded at a time.
func (f *ReviewField) toggleSelected() {
	if f.selectedIdx >= 0 && f.selectedIdx < len(f.sections) {
		if f.sections[f.selectedIdx].Expandable {
			wasExpanded := f.sections[f.selectedIdx].Expanded
			// Close all sections first (accordion behavior)
			for i := range f.sections {
				f.sections[i].Expanded = false
			}
			// Toggle the selected one (if it was closed, open it; if it was open, leave it closed)
			if !wasExpanded {
				f.sections[f.selectedIdx].Expanded = true
			}
		}
	}
}

// View renders the field.
func (f *ReviewField) View() string {
	return f.viewport.View()
}

// Focus focuses the field.
func (f *ReviewField) Focus() tea.Cmd {
	f.focused = true
	// Reload sections from data provider
	f.loadSections()
	f.viewport.SetContent(f.renderContent())
	f.viewport.GotoTop()
	return nil
}

// Blur blurs the field.
func (f *ReviewField) Blur() tea.Cmd {
	f.focused = false
	return nil
}

// Error returns nil (no validation).
func (f *ReviewField) Error() error {
	return nil
}

// Run runs the field standalone.
func (f *ReviewField) Run() error {
	return huh.Run(f)
}

// RunAccessible runs in accessible mode.
func (f *ReviewField) RunAccessible(w io.Writer, r io.Reader) error {
	f.loadSections()
	for _, s := range f.sections {
		_, _ = fmt.Fprintf(w, "%s: %s\n", s.Label, s.Value)
		if s.Expandable && s.Content != "" {
			_, _ = fmt.Fprintf(w, "%s\n", s.Content)
		}
	}
	return nil
}

// Skip returns false - this field should not be skipped.
func (f *ReviewField) Skip() bool {
	return false
}

// Zoom returns false - let the group manage height distribution.
func (f *ReviewField) Zoom() bool {
	return false
}

// KeyBinds returns the keybindings.
func (f *ReviewField) KeyBinds() []key.Binding {
	return []key.Binding{f.keymap.Up, f.keymap.Down, f.keymap.Toggle, f.keymap.Next}
}

// WithTheme sets the theme.
func (f *ReviewField) WithTheme(theme *huh.Theme) huh.Field {
	f.theme = theme
	return f
}

// WithAccessible is deprecated but required.
func (f *ReviewField) WithAccessible(accessible bool) huh.Field {
	return f
}

// WithKeyMap sets the keymap.
func (f *ReviewField) WithKeyMap(k *huh.KeyMap) huh.Field {
	return f
}

// WithWidth sets the width.
func (f *ReviewField) WithWidth(width int) huh.Field {
	f.width = width
	f.viewport.Width = width - 4
	return f
}

// WithHeight sets the height allocated by the form.
func (f *ReviewField) WithHeight(height int) huh.Field {
	f.height = height
	viewportHeight := height - 2
	if viewportHeight < 5 {
		viewportHeight = 5
	}
	f.viewport.Height = viewportHeight
	return f
}

// WithPosition sets position info.
func (f *ReviewField) WithPosition(p huh.FieldPosition) huh.Field {
	f.keymap.Prev.SetEnabled(!p.IsFirst())
	f.keymap.Next.SetEnabled(!p.IsLast())
	return f
}

// GetKey returns empty string (no key).
func (f *ReviewField) GetKey() string {
	return ""
}

// GetValue returns nil (no value).
func (f *ReviewField) GetValue() any {
	return nil
}

// Helper to read file content for expandable sections
func readFileContent(path string) string {
	if path == "" {
		return ""
	}
	expanded := path
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			expanded = strings.Replace(path, "~", home, 1)
		}
	}
	data, err := os.ReadFile(expanded)
	if err != nil {
		return fmt.Sprintf("(error reading file: %v)", err)
	}
	return string(data)
}
