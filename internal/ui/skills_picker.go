package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ayo/internal/skills"
)

// SkillsPickerResult contains the selected skills.
type SkillsPickerResult struct {
	Skills   []string
	Accepted bool
}

// SkillsPicker is a single-screen component for selecting skills
// with tabs for built-in vs user-defined skills.
type SkillsPicker struct {
	builtinSkills []skills.Metadata
	userSkills    []skills.Metadata

	activeTab int // 0 = built-in, 1 = user
	cursor    int
	selected  map[string]bool

	width  int
	height int
}

// NewSkillsPicker creates a new skills picker.
func NewSkillsPicker(allSkills []skills.Metadata) *SkillsPicker {
	var builtin, user []skills.Metadata
	for _, s := range allSkills {
		switch s.Source {
		case skills.SourceBuiltIn:
			builtin = append(builtin, s)
		default:
			user = append(user, s)
		}
	}

	return &SkillsPicker{
		builtinSkills: builtin,
		userSkills:    user,
		activeTab:     0,
		cursor:        0,
		selected:      make(map[string]bool),
	}
}

// SelectedSkills returns the list of selected skill names.
func (p *SkillsPicker) SelectedSkills() []string {
	var result []string
	for name := range p.selected {
		result = append(result, name)
	}
	return result
}

// SetSelected sets the initially selected skills.
func (p *SkillsPicker) SetSelected(names []string) {
	for _, name := range names {
		p.selected[name] = true
	}
}

func (p *SkillsPicker) currentList() []skills.Metadata {
	if p.activeTab == 0 {
		return p.builtinSkills
	}
	return p.userSkills
}

func (p *SkillsPicker) Init() tea.Cmd {
	return nil
}

func (p *SkillsPicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		return p, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			// Switch tabs
			p.activeTab = (p.activeTab + 1) % 2
			p.cursor = 0
			return p, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			if p.activeTab > 0 {
				p.activeTab--
				p.cursor = 0
			}
			return p, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			if p.activeTab < 1 {
				p.activeTab++
				p.cursor = 0
			}
			return p, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			list := p.currentList()
			if p.cursor > 0 {
				p.cursor--
			} else if len(list) > 0 {
				p.cursor = len(list) - 1
			}
			return p, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			list := p.currentList()
			if p.cursor < len(list)-1 {
				p.cursor++
			} else {
				p.cursor = 0
			}
			return p, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys(" ", "x"))):
			// Toggle selection
			list := p.currentList()
			if len(list) > 0 && p.cursor < len(list) {
				name := list[p.cursor].Name
				if p.selected[name] {
					delete(p.selected, name)
				} else {
					p.selected[name] = true
				}
			}
			return p, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return p, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c", "esc"))):
			p.selected = make(map[string]bool) // Clear on cancel
			return p, tea.Quit
		}
	}

	return p, nil
}

func (p *SkillsPicker) View() string {
	// Styles
	purple := lipgloss.Color("#a78bfa")
	cyan := lipgloss.Color("#67e8f9")
	muted := lipgloss.Color("#6b7280")
	text := lipgloss.Color("#e5e7eb")
	subtle := lipgloss.Color("#374151")

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
	activeTabStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(cyan).
		Background(lipgloss.Color("#1e1e2e")).
		Padding(0, 2)
	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(muted).
		Padding(0, 2)
	dividerStyle := lipgloss.NewStyle().Foreground(subtle)
	selectedStyle := lipgloss.NewStyle().Foreground(cyan).Bold(true)
	unselectedStyle := lipgloss.NewStyle().Foreground(text)
	cursorStyle := lipgloss.NewStyle().Foreground(cyan)
	descStyle := lipgloss.NewStyle().Foreground(muted)
	emptyStyle := lipgloss.NewStyle().Foreground(muted).Italic(true)
	helpStyle := lipgloss.NewStyle().Foreground(muted)

	var b strings.Builder

	// Header
	b.WriteString("\n")
	b.WriteString(headerStyle.Render("  Select Skills"))
	b.WriteString("\n")
	b.WriteString(dividerStyle.Render("  " + strings.Repeat("─", 58)))
	b.WriteString("\n\n")

	// Tabs
	builtinLabel := fmt.Sprintf("Built-in (%d)", len(p.builtinSkills))
	userLabel := fmt.Sprintf("User-defined (%d)", len(p.userSkills))

	var builtinTab, userTab string
	if p.activeTab == 0 {
		builtinTab = activeTabStyle.Render("● " + builtinLabel)
		userTab = inactiveTabStyle.Render("○ " + userLabel)
	} else {
		builtinTab = inactiveTabStyle.Render("○ " + builtinLabel)
		userTab = activeTabStyle.Render("● " + userLabel)
	}

	b.WriteString("  ")
	b.WriteString(builtinTab)
	b.WriteString("  ")
	b.WriteString(userTab)
	b.WriteString("\n\n")

	// Skills list
	list := p.currentList()
	if len(list) == 0 {
		if p.activeTab == 0 {
			b.WriteString("  ")
			b.WriteString(emptyStyle.Render("No built-in skills installed"))
			b.WriteString("\n  ")
			b.WriteString(emptyStyle.Render("Run: ayo setup"))
			b.WriteString("\n")
		} else {
			b.WriteString("  ")
			b.WriteString(emptyStyle.Render("No user-defined skills"))
			b.WriteString("\n  ")
			b.WriteString(emptyStyle.Render("Create one with: ayo skills create <name> --shared"))
			b.WriteString("\n")
		}
	} else {
		for i, s := range list {
			// Cursor
			cursor := "  "
			if i == p.cursor {
				cursor = cursorStyle.Render("> ")
			}

			// Checkbox
			checkbox := "[ ]"
			nameStyle := unselectedStyle
			if p.selected[s.Name] {
				checkbox = "[x]"
				nameStyle = selectedStyle
			}

			b.WriteString(cursor)
			b.WriteString(checkbox)
			b.WriteString(" ")
			b.WriteString(nameStyle.Render(s.Name))
			b.WriteString("\n")

			// Description (indented)
			desc := s.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			b.WriteString("      ")
			b.WriteString(descStyle.Render(desc))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(dividerStyle.Render("  " + strings.Repeat("─", 58)))
	b.WriteString("\n")

	// Help
	b.WriteString("  ")
	b.WriteString(helpStyle.Render("tab/←/→ switch tabs • ↑/↓ navigate • space toggle • enter confirm"))
	b.WriteString("\n")

	return b.String()
}

// Run runs the skills picker as a standalone program.
func (p *SkillsPicker) Run() (SkillsPickerResult, error) {
	prog := tea.NewProgram(p, tea.WithAltScreen())
	finalModel, err := prog.Run()
	if err != nil {
		return SkillsPickerResult{}, err
	}

	picker := finalModel.(*SkillsPicker)
	return SkillsPickerResult{
		Skills:   picker.SelectedSkills(),
		Accepted: len(picker.selected) > 0 || true, // Always accepted unless cancelled
	}, nil
}
