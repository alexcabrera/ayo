package ui

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"ayo/internal/skills"
)

// AgentCreateResult contains all the data collected from the agent creation wizard.
type AgentCreateResult struct {
	// Identity
	Handle      string
	Description string
	Model       string

	// Tools
	AllowedTools []string

	// Skills
	Skills              []string
	ExcludeSkills       []string
	IgnoreBuiltinSkills bool
	IgnoreSharedSkills  bool

	// System
	SystemMessage string
	SystemFile    string

	// Chaining
	InputSchemaFile  string
	OutputSchemaFile string
}

// AgentCreateFormOptions provides configuration for the agent creation form.
type AgentCreateFormOptions struct {
	// Models is the list of available models to select from.
	Models []string
	// AvailableSkills is the list of discovered skills.
	AvailableSkills []skills.Metadata
	// AvailableTools is the list of available tools.
	AvailableTools []string
	// PrefilledHandle is an optional pre-filled handle from CLI args.
	PrefilledHandle string
}

// AgentCreateForm is a multi-step wizard for creating agents.
type AgentCreateForm struct {
	opts AgentCreateFormOptions
}

// NewAgentCreateForm creates a new agent creation form.
func NewAgentCreateForm(opts AgentCreateFormOptions) *AgentCreateForm {
	return &AgentCreateForm{opts: opts}
}

// Styles for the wizard
var (
	// Unused but kept for potential future use
	_ = lipgloss.NewStyle()
)

// wizardModel wraps huh.Form with progress indicator
type wizardModel struct {
	form        *huh.Form
	width       int
	height      int
	quitting    bool
	err         error
	groupIndex  int
	totalGroups int
}

func (m wizardModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Update the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	// Check if form is complete
	if m.form.State == huh.StateCompleted {
		return m, tea.Quit
	}

	return m, cmd
}

func (m wizardModel) View() string {
	if m.quitting {
		return ""
	}

	return m.form.View()
}

// Run executes the multi-step form wizard.
func (f *AgentCreateForm) Run(ctx context.Context) (AgentCreateResult, error) {
	if len(f.opts.Models) == 0 {
		return AgentCreateResult{}, fmt.Errorf("no models configured; update config provider models")
	}

	var res AgentCreateResult
	res.Handle = f.opts.PrefilledHandle

	// Build model options
	modelOpts := make([]huh.Option[string], 0, len(f.opts.Models))
	for _, m := range f.opts.Models {
		modelOpts = append(modelOpts, huh.NewOption(m, m))
	}

	// Build tool options with descriptions
	toolDescriptions := map[string]string{
		"bash":       "Execute shell commands",
		"agent_call": "Delegate to other agents",
	}
	toolOpts := make([]huh.Option[string], 0, len(f.opts.AvailableTools))
	for _, t := range f.opts.AvailableTools {
		desc := toolDescriptions[t]
		if desc == "" {
			desc = t
		}
		label := fmt.Sprintf("%s - %s", t, desc)
		toolOpts = append(toolOpts, huh.NewOption(label, t))
	}

	// Tracking variables for conditional groups
	var systemSource string = "inline"
	var enableChaining bool
	var confirmed bool

	// Default tools selection
	res.AllowedTools = []string{"bash"}

	// Handle without @ prefix (we'll add it automatically)
	var handleName string
	if strings.HasPrefix(res.Handle, "@") {
		handleName = strings.TrimPrefix(res.Handle, "@")
	} else {
		handleName = res.Handle
	}

	// Phase 1: Identity, Tools, and basic settings
	form1 := huh.NewForm(
		// Step 1: Identity
		huh.NewGroup(
			huh.NewNote().
				Title("Step 1 of 5 ─ Identity").
				Description("Basic agent information\n"),
			huh.NewInput().
				Title("Handle").
				Prompt("@ ").
				Placeholder("myagent").
				Value(&handleName).
				Validate(func(v string) error {
					if v == "" {
						return fmt.Errorf("required")
					}
					name := strings.TrimPrefix(v, "@")
					if name == "" {
						return fmt.Errorf("required")
					}
					if strings.HasPrefix(name, "ayo.") || name == "ayo" {
						return fmt.Errorf("cannot use reserved 'ayo' namespace")
					}
					if strings.ContainsAny(name, " \t\n") {
						return fmt.Errorf("handle cannot contain spaces")
					}
					return nil
				}),
			huh.NewInput().
				Title("Description").
				Placeholder("A helpful assistant that...").
				Value(&res.Description),
			huh.NewSelect[string]().
				Title("Model").
				Options(modelOpts...).
				Value(&res.Model),
		),

		// Step 2: Tools
		huh.NewGroup(
			huh.NewNote().
				Title("Step 2 of 5 ─ Tools").
				Description("Select capabilities for your agent\n"),
			huh.NewMultiSelect[string]().
				Title("Allowed Tools").
				Description("These tools will be available to the agent").
				Options(toolOpts...).
				Value(&res.AllowedTools),
		),
	).WithTheme(huh.ThemeCharm()).WithShowHelp(true)

	// Run phase 1
	p1 := tea.NewProgram(wizardModel{form: form1}, tea.WithAltScreen())
	finalModel1, err := p1.Run()
	if err != nil {
		return AgentCreateResult{}, err
	}
	wm1 := finalModel1.(wizardModel)
	if wm1.quitting && wm1.form.State != huh.StateCompleted {
		return AgentCreateResult{}, fmt.Errorf("agent creation cancelled")
	}

	// Phase 2: Skills picker
	hasSkills := len(f.opts.AvailableSkills) > 0
	if hasSkills {
		picker := NewSkillsPicker(f.opts.AvailableSkills)
		pickerResult, err := picker.Run()
		if err != nil {
			return AgentCreateResult{}, err
		}
		res.Skills = pickerResult.Skills
	}

	// Phase 3: System prompt source selection
	form2 := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Step 4 of 5 ─ System Prompt").
				Description("Define the agent's personality and behavior\n"),
			huh.NewSelect[string]().
				Title("Source").
				Description("How would you like to provide the system prompt?").
				Options(
					huh.NewOption("Enter inline", "inline"),
					huh.NewOption("Browse for file", "file"),
				).
				Value(&systemSource),
		),
	).WithTheme(huh.ThemeCharm()).WithShowHelp(true)

	p2 := tea.NewProgram(wizardModel{form: form2}, tea.WithAltScreen())
	finalModel2, err := p2.Run()
	if err != nil {
		return AgentCreateResult{}, err
	}
	wm2 := finalModel2.(wizardModel)
	if wm2.quitting && wm2.form.State != huh.StateCompleted {
		return AgentCreateResult{}, fmt.Errorf("agent creation cancelled")
	}

	// Phase 4: System prompt content based on source selection
	if systemSource == "file" {
		// Use file picker
		home, _ := os.UserHomeDir()
		picker := NewFilePicker(FilePickerOptions{
			Title:        "Step 4 of 5 ─ Select System Prompt File",
			StartDir:     home,
			AllowedTypes: []string{".md", ".txt"},
		})
		result, err := picker.Run()
		if err != nil {
			return AgentCreateResult{}, err
		}
		if !result.Selected {
			return AgentCreateResult{}, fmt.Errorf("agent creation cancelled")
		}
		res.SystemFile = result.Path

		// Confirm file selection
		formFile := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("Step 4 of 5 ─ System Prompt").
					Description(fmt.Sprintf("Selected: %s\n", res.SystemFile)),
			),
		).WithTheme(huh.ThemeCharm()).WithShowHelp(true)

		pFile := tea.NewProgram(wizardModel{form: formFile}, tea.WithAltScreen())
		finalFile, err := pFile.Run()
		if err != nil {
			return AgentCreateResult{}, err
		}
		wmFile := finalFile.(wizardModel)
		if wmFile.quitting && wmFile.form.State != huh.StateCompleted {
			return AgentCreateResult{}, fmt.Errorf("agent creation cancelled")
		}
	} else {
		// Inline editor
		formInline := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("Step 4 of 5 ─ System Prompt").
					Description("Enter the system message directly\n"),
				huh.NewText().
					Title("System message").
					Placeholder("You are a helpful assistant...").
					Description("Ctrl+E to open external editor").
					Value(&res.SystemMessage).
					CharLimit(0).
					Lines(8),
			),
		).WithTheme(huh.ThemeCharm()).WithShowHelp(true)

		pInline := tea.NewProgram(wizardModel{form: formInline}, tea.WithAltScreen())
		finalInline, err := pInline.Run()
		if err != nil {
			return AgentCreateResult{}, err
		}
		wmInline := finalInline.(wizardModel)
		if wmInline.quitting && wmInline.form.State != huh.StateCompleted {
			return AgentCreateResult{}, fmt.Errorf("agent creation cancelled")
		}
	}

	// Phase 5: Chaining
	form3 := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Step 5 of 5 ─ Agent Chaining").
				Description("Enable structured I/O for pipeline composition\n"),
			huh.NewConfirm().
				Title("Enable chaining?").
				Description("Add JSON schemas for input/output validation").
				Value(&enableChaining),
		),
	).WithTheme(huh.ThemeCharm()).WithShowHelp(true)

	p3 := tea.NewProgram(wizardModel{form: form3}, tea.WithAltScreen())
	finalModel3, err := p3.Run()
	if err != nil {
		return AgentCreateResult{}, err
	}
	wm3 := finalModel3.(wizardModel)
	if wm3.quitting && wm3.form.State != huh.StateCompleted {
		return AgentCreateResult{}, fmt.Errorf("agent creation cancelled")
	}

	// Phase 6: Schema files (if chaining enabled)
	if enableChaining {
		// Input schema file picker
		home, _ := os.UserHomeDir()
		inputPicker := NewFilePicker(FilePickerOptions{
			Title:        "Step 5 of 5 ─ Select Input Schema (optional)",
			StartDir:     home,
			AllowedTypes: []string{".json", ".jsonschema"},
		})
		inputResult, err := inputPicker.Run()
		if err != nil {
			return AgentCreateResult{}, err
		}
		if inputResult.Selected {
			res.InputSchemaFile = inputResult.Path
		}

		// Output schema file picker
		outputPicker := NewFilePicker(FilePickerOptions{
			Title:        "Step 5 of 5 ─ Select Output Schema (optional)",
			StartDir:     home,
			AllowedTypes: []string{".json", ".jsonschema"},
		})
		outputResult, err := outputPicker.Run()
		if err != nil {
			return AgentCreateResult{}, err
		}
		if outputResult.Selected {
			res.OutputSchemaFile = outputResult.Path
		}
	}

	// Phase 7: Review and confirm
	formReview := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Review").
				DescriptionFunc(func() string {
					reviewRes := res
					name := strings.TrimPrefix(handleName, "@")
					reviewRes.Handle = "@" + name
					return f.buildReviewSummary(reviewRes, systemSource)
				}, &handleName),
			huh.NewConfirm().
				Title("Create this agent?").
				Affirmative("Create").
				Negative("Cancel").
				Value(&confirmed),
		),
	).WithTheme(huh.ThemeCharm()).WithShowHelp(true)

	pReview := tea.NewProgram(wizardModel{form: formReview}, tea.WithAltScreen())
	finalReview, err := pReview.Run()
	if err != nil {
		return AgentCreateResult{}, err
	}
	wmReview := finalReview.(wizardModel)
	if wmReview.quitting && wmReview.form.State != huh.StateCompleted {
		return AgentCreateResult{}, fmt.Errorf("agent creation cancelled")
	}

	if !confirmed {
		return AgentCreateResult{}, fmt.Errorf("agent creation cancelled")
	}

	// Normalize handle
	handleName = strings.TrimPrefix(handleName, "@")
	res.Handle = "@" + handleName

	// If file source was selected, read the file content
	if systemSource == "file" && res.SystemFile != "" {
		expanded := expandPath(res.SystemFile)
		data, err := os.ReadFile(expanded)
		if err != nil {
			return AgentCreateResult{}, fmt.Errorf("read system file: %w", err)
		}
		res.SystemMessage = string(data)
	}

	return res, nil
}

// buildReviewSummary creates a formatted summary of the agent configuration.
func (f *AgentCreateForm) buildReviewSummary(res AgentCreateResult, systemSource string) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Handle:        %s\n", res.Handle))
	if res.Description != "" {
		desc := res.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		b.WriteString(fmt.Sprintf("  Description:   %s\n", desc))
	}
	b.WriteString(fmt.Sprintf("  Model:         %s\n", res.Model))

	if len(res.AllowedTools) > 0 {
		b.WriteString(fmt.Sprintf("  Tools:         %s\n", strings.Join(res.AllowedTools, ", ")))
	} else {
		b.WriteString("  Tools:         (none)\n")
	}

	if len(res.Skills) > 0 {
		skillList := strings.Join(res.Skills, ", ")
		if len(skillList) > 40 {
			skillList = skillList[:37] + "..."
		}
		b.WriteString(fmt.Sprintf("  Skills:        %s\n", skillList))
	} else {
		b.WriteString("  Skills:        (none)\n")
	}

	if systemSource == "file" && res.SystemFile != "" {
		b.WriteString(fmt.Sprintf("  System:        from %s\n", res.SystemFile))
	} else if res.SystemMessage != "" {
		preview := res.SystemMessage
		if len(preview) > 35 {
			preview = preview[:32] + "..."
		}
		preview = strings.ReplaceAll(preview, "\n", " ")
		b.WriteString(fmt.Sprintf("  System:        %s\n", preview))
	}

	if res.InputSchemaFile != "" || res.OutputSchemaFile != "" {
		b.WriteString("  Chaining:      enabled\n")
		if res.InputSchemaFile != "" {
			b.WriteString(fmt.Sprintf("    Input:       %s\n", res.InputSchemaFile))
		}
		if res.OutputSchemaFile != "" {
			b.WriteString(fmt.Sprintf("    Output:      %s\n", res.OutputSchemaFile))
		}
	}

	b.WriteString("\n")

	return b.String()
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return strings.Replace(path, "~", home, 1)
	}
	return path
}
