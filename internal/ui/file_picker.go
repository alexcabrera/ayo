package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FilePickerResult contains the result of the file picker.
type FilePickerResult struct {
	Path     string
	Selected bool
}

// FilePickerOptions configures the file picker.
type FilePickerOptions struct {
	// Title shown at the top of the picker
	Title string
	// StartDir is the initial directory to show
	StartDir string
	// AllowedTypes filters files by extension (e.g., ".md", ".json")
	AllowedTypes []string
	// DirAllowed allows selecting directories
	DirAllowed bool
}

// filePickerModel wraps the bubbles filepicker with styling
type filePickerModel struct {
	picker   filepicker.Model
	help     help.Model
	keyMap   filePickerKeyMap
	title    string
	width    int
	height   int
	selected string
	quitting bool
}

type filePickerKeyMap struct {
	Select key.Binding
	Cancel key.Binding
}

func (k filePickerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Cancel}
}

func (k filePickerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Select, k.Cancel}}
}

// NewFilePicker creates a new file picker component.
func NewFilePicker(opts FilePickerOptions) *filePickerModel {
	fp := filepicker.New()

	// Set starting directory
	if opts.StartDir != "" {
		fp.CurrentDirectory = opts.StartDir
	} else if home, err := os.UserHomeDir(); err == nil {
		fp.CurrentDirectory = home
	} else if cwd, err := os.Getwd(); err == nil {
		fp.CurrentDirectory = cwd
	}

	// Configure options
	fp.AllowedTypes = opts.AllowedTypes
	fp.DirAllowed = opts.DirAllowed
	fp.FileAllowed = true
	fp.ShowPermissions = false
	fp.ShowSize = false
	fp.ShowHidden = false
	fp.AutoHeight = false
	fp.Height = 12

	// Style the picker
	purple := lipgloss.Color("#a78bfa")
	cyan := lipgloss.Color("#67e8f9")
	muted := lipgloss.Color("#6b7280")
	text := lipgloss.Color("#e5e7eb")

	fp.Styles.Cursor = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	fp.Styles.Directory = lipgloss.NewStyle().Foreground(purple).Bold(true)
	fp.Styles.File = lipgloss.NewStyle().Foreground(text)
	fp.Styles.Selected = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	fp.Styles.Symlink = lipgloss.NewStyle().Foreground(muted).Italic(true)
	fp.Cursor = ">"

	h := help.New()

	keyMap := filePickerKeyMap{
		Select: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Cancel: key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "cancel")),
	}

	title := opts.Title
	if title == "" {
		title = "Select File"
	}

	return &filePickerModel{
		picker: fp,
		help:   h,
		keyMap: keyMap,
		title:  title,
	}
}

func (m *filePickerModel) Init() tea.Cmd {
	return m.picker.Init()
}

func (m *filePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Cancel):
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.picker, cmd = m.picker.Update(msg)

	// Check if file was selected
	if didSelect, path := m.picker.DidSelectFile(msg); didSelect {
		m.selected = path
		return m, tea.Quit
	}

	// Check if directory was selected (if allowed)
	if didSelect, path := m.picker.DidSelectDisabledFile(msg); didSelect && m.picker.DirAllowed {
		// Check if it's actually a directory
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			m.selected = path
			return m, tea.Quit
		}
	}

	return m, cmd
}

func (m *filePickerModel) View() string {
	// Styles
	purple := lipgloss.Color("#a78bfa")
	muted := lipgloss.Color("#6b7280")
	subtle := lipgloss.Color("#374151")

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
	pathStyle := lipgloss.NewStyle().Foreground(muted)
	dividerStyle := lipgloss.NewStyle().Foreground(subtle)
	helpStyle := lipgloss.NewStyle().Foreground(muted)

	var b strings.Builder

	// Header
	b.WriteString("\n")
	b.WriteString(headerStyle.Render("  " + m.title))
	b.WriteString("\n")
	b.WriteString(dividerStyle.Render("  " + strings.Repeat("─", 58)))
	b.WriteString("\n\n")

	// Current path
	currentPath := m.picker.CurrentDirectory
	if home, err := os.UserHomeDir(); err == nil {
		currentPath = strings.Replace(currentPath, home, "~", 1)
	}
	b.WriteString("  ")
	b.WriteString(pathStyle.Render(currentPath))
	b.WriteString("\n\n")

	// File picker
	pickerView := m.picker.View()
	// Indent each line
	for _, line := range strings.Split(pickerView, "\n") {
		b.WriteString("  ")
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dividerStyle.Render("  " + strings.Repeat("─", 58)))
	b.WriteString("\n")

	// Help
	b.WriteString("  ")
	b.WriteString(helpStyle.Render("↑/↓ navigate • ←/h back • →/l/enter open/select • esc cancel"))
	b.WriteString("\n")

	return b.String()
}

// Run runs the file picker and returns the selected path.
func (m *filePickerModel) Run() (FilePickerResult, error) {
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return FilePickerResult{}, err
	}

	fm := finalModel.(*filePickerModel)
	return FilePickerResult{
		Path:     fm.selected,
		Selected: fm.selected != "" && !fm.quitting,
	}, nil
}

// RunFilePicker is a convenience function to run the file picker.
func RunFilePicker(opts FilePickerOptions) (string, bool, error) {
	picker := NewFilePicker(opts)
	result, err := picker.Run()
	if err != nil {
		return "", false, err
	}
	return result.Path, result.Selected, nil
}

// ExpandPath expands ~ to the user's home directory.
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
