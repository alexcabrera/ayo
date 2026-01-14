package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// formKeyMap implements help.KeyMap for the wizard.
type formKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Select   key.Binding
	Skip     key.Binding
	Back     key.Binding
	Quit     key.Binding
	navigate key.Binding
	scroll   key.Binding
}

func (k formKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.navigate, k.Select, k.Skip, k.Back, k.Quit}
}

func (k formKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.scroll},
		{k.Select, k.Skip, k.Back, k.Quit},
	}
}

var defaultFormKeyMap = formKeyMap{
	Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Skip:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "skip")),
	Back:     key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("S-tab", "back")),
	Quit:     key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl-c", "cancel")),
	navigate: key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "navigate")),
	scroll:   key.NewBinding(key.WithKeys("pgup", "pgdown"), key.WithHelp("pgup/dn", "scroll")),
}

// fullScreenForm wraps a huh.Form to provide full-screen layout with pinned help.
type fullScreenForm struct {
	form            *huh.Form
	help            help.Model
	keyMap          formKeyMap
	helpStyle       lipgloss.Style
	width           int
	height          int
	ready           bool
	showCancelDialog bool
	cancelSelection  int // 0 = No, go back (default), 1 = Yes, cancel
}

// newFullScreenForm creates a new full-screen form wrapper.
func newFullScreenForm(form *huh.Form) *fullScreenForm {
	h := help.New()
	h.ShowAll = false
	return &fullScreenForm{
		form:   form,
		help:   h,
		keyMap: defaultFormKeyMap,
		helpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 1),
		cancelSelection: 0, // Default to "No, go back"
	}
}

func (m *fullScreenForm) Init() tea.Cmd {
	return m.form.Init()
}

func (m *fullScreenForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle cancel dialog if shown
	if m.showCancelDialog {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "left", "h":
				m.cancelSelection = 0
				return m, nil
			case "right", "l":
				m.cancelSelection = 1
				return m, nil
			case "enter":
				if m.cancelSelection == 1 {
					// User confirmed cancel
					m.form.State = huh.StateAborted
					return m, tea.Quit
				}
				// Go back to form
				m.showCancelDialog = false
				return m, nil
			case "esc", "n":
				// Cancel the dialog, go back to form
				m.showCancelDialog = false
				return m, nil
			case "y":
				// Confirm cancel
				m.form.State = huh.StateAborted
				return m, tea.Quit
			case "ctrl+c":
				// Double ctrl-c confirms cancel
				m.form.State = huh.StateAborted
				return m, tea.Quit
			}
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.showCancelDialog = true
			m.cancelSelection = 0 // Default to "No, go back"
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Account for 1 char padding on all sides (2 chars total width, 2 lines total height)
		innerWidth := msg.Width - 2
		innerHeight := msg.Height - 2

		m.help.Width = innerWidth

		// Reserve space for help bar (2 lines)
		helpHeight := 2
		formHeight := innerHeight - helpHeight

		// Update form with adjusted size
		m.form.WithWidth(innerWidth)
		m.form.WithHeight(formHeight)

		if !m.ready {
			m.ready = true
		}

		// Pass the adjusted size to the form
		adjustedMsg := tea.WindowSizeMsg{
			Width:  innerWidth,
			Height: formHeight,
		}
		_, cmd := m.form.Update(adjustedMsg)
		return m, cmd
	}

	// Pass all other messages to the form
	_, cmd := m.form.Update(msg)

	// Check if form is done
	if m.form.State == huh.StateCompleted || m.form.State == huh.StateAborted {
		return m, tea.Quit
	}

	return m, cmd
}

func (m *fullScreenForm) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Inner dimensions (accounting for 1 char padding)
	innerHeight := m.height - 2
	innerWidth := m.width - 2

	// Help bar at bottom
	helpView := m.helpStyle.Render(m.help.View(m.keyMap))
	helpHeight := lipgloss.Height(helpView)

	// Form content - huh's form already handles its own viewport scrolling
	formView := m.form.View()
	formHeight := lipgloss.Height(formView)

	// Calculate available space for form (reserve space for help)
	availableForForm := innerHeight - helpHeight

	// Add padding between form and help if form is shorter
	var padding string
	if formHeight < availableForForm {
		paddingLines := availableForForm - formHeight
		if paddingLines > 0 {
			padding = strings.Repeat("\n", paddingLines)
		}
	}

	// Combine form, padding, and help
	content := lipgloss.JoinVertical(lipgloss.Left,
		formView,
		padding,
		helpView,
	)

	// If showing cancel dialog, overlay it
	if m.showCancelDialog {
		content = m.renderCancelDialog(innerWidth, innerHeight)
	}

	// Apply 1 character padding around the entire content
	paddedStyle := lipgloss.NewStyle().Padding(1, 1)
	return paddedStyle.Render(content)
}

// renderCancelDialog renders the cancel confirmation dialog.
func (m *fullScreenForm) renderCancelDialog(width, height int) string {
	// Button text constants for consistent sizing
	const noButtonText = "No, go back"
	const yesButtonText = "Yes, cancel"

	// Calculate button width - use the longer button text plus padding
	buttonWidth := len(yesButtonText) + 4 // +4 for padding (2 on each side)
	if len(noButtonText)+4 > buttonWidth {
		buttonWidth = len(noButtonText) + 4
	}

	// Dialog styles
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#a78bfa")).
		Padding(1, 2).
		Width(50)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#a78bfa"))

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e5e7eb"))

	// Base button style with fixed width for consistent sizing
	buttonBaseStyle := lipgloss.NewStyle().
		Width(buttonWidth).
		Align(lipgloss.Center)

	selectedButtonStyle := buttonBaseStyle.
		Bold(true).
		Background(lipgloss.Color("#a78bfa")).
		Foreground(lipgloss.Color("#1a1a2e"))

	unselectedButtonStyle := buttonBaseStyle.
		Foreground(lipgloss.Color("#6b7280"))

	dangerButtonStyle := buttonBaseStyle.
		Bold(true).
		Background(lipgloss.Color("#ef4444")).
		Foreground(lipgloss.Color("#ffffff"))

	unselectedDangerStyle := buttonBaseStyle.
		Foreground(lipgloss.Color("#ef4444"))

	// Build dialog content
	title := titleStyle.Render("Cancel Agent Creation?")
	message := messageStyle.Render("Are you sure you want to cancel?\nAll progress will be lost.")

	// Buttons: [No, go back] on left (default selected), [Yes, cancel] on right (danger)
	var noButton, yesButton string
	if m.cancelSelection == 0 {
		noButton = selectedButtonStyle.Render(noButtonText)
		yesButton = unselectedDangerStyle.Render(yesButtonText)
	} else {
		noButton = unselectedButtonStyle.Render(noButtonText)
		yesButton = dangerButtonStyle.Render(yesButtonText)
	}

	// Fixed spacing between buttons
	buttonSpacer := lipgloss.NewStyle().Width(4).Render("")
	buttons := lipgloss.JoinHorizontal(lipgloss.Center, noButton, buttonSpacer, yesButton)

	// Center the buttons row
	buttonsRow := lipgloss.NewStyle().Width(46).Align(lipgloss.Center).Render(buttons)

	dialogContent := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		message,
		"",
		buttonsRow,
	)

	dialog := dialogStyle.Render(dialogContent)

	// Center the dialog
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}

// Run runs the full-screen form.
func (m *fullScreenForm) Run() error {
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

// State returns the form's state.
func (m *fullScreenForm) State() huh.FormState {
	return m.form.State
}
