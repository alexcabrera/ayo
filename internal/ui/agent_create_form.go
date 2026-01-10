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
	SystemMessage      string
	SystemFile         string
	IgnoreSharedSystem bool

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

	// Separate skills by source for tabs
	var builtinSkillOpts, userSkillOpts []huh.Option[string]
	for _, s := range f.opts.AvailableSkills {
		desc := s.Description
		if len(desc) > 45 {
			desc = desc[:42] + "..."
		}
		label := fmt.Sprintf("%s - %s", s.Name, desc)
		opt := huh.NewOption(label, s.Name)

		switch s.Source {
		case skills.SourceBuiltIn:
			builtinSkillOpts = append(builtinSkillOpts, opt)
		default:
			userSkillOpts = append(userSkillOpts, opt)
		}
	}

	// Skills tab selection
	var skillsTab string = "built-in"
	builtinCount := len(builtinSkillOpts)
	userCount := len(userSkillOpts)

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

	// Build the form
	form := huh.NewForm(
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

		// Step 3: Skills - Tab selector
		huh.NewGroup(
			huh.NewNote().
				Title("Step 3 of 5 ─ Skills").
				DescriptionFunc(func() string {
					return renderSkillsTabs(skillsTab, builtinCount, userCount)
				}, &skillsTab),
			huh.NewSelect[string]().
				Title("").
				Options(
					huh.NewOption(fmt.Sprintf("◉ Built-in (%d)", builtinCount), "built-in"),
					huh.NewOption(fmt.Sprintf("◉ User (%d)", userCount), "user"),
				).
				Value(&skillsTab),
		).WithHideFunc(func() bool {
			return builtinCount == 0 && userCount == 0
		}),

		// Step 3a: Built-in skills (conditional)
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Built-in Skills").
				Description("Skills bundled with ayo").
				Options(builtinSkillOpts...).
				Filterable(true).
				Height(8).
				Value(&res.Skills),
		).WithHideFunc(func() bool {
			return skillsTab != "built-in" || builtinCount == 0
		}),

		// Step 3b: User skills (conditional)
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("User Skills").
				Description("Your custom skills").
				Options(userSkillOpts...).
				Filterable(true).
				Height(8).
				Value(&res.Skills),
		).WithHideFunc(func() bool {
			return skillsTab != "user" || userCount == 0
		}),

		// No skills available
		huh.NewGroup(
			huh.NewNote().
				Title("Step 3 of 5 ─ Skills").
				Description("No skills available.\nCreate skills with: ayo skills create <name>"),
		).WithHideFunc(func() bool {
			return builtinCount > 0 || userCount > 0
		}),

		// Step 4a: System source selection
		huh.NewGroup(
			huh.NewNote().
				Title("Step 4 of 5 ─ System Prompt").
				Description("Define the agent's personality and behavior\n"),
			huh.NewSelect[string]().
				Title("Source").
				Description("How would you like to provide the system prompt?").
				Options(
					huh.NewOption("Enter inline", "inline"),
					huh.NewOption("Load from file", "file"),
				).
				Value(&systemSource),
			huh.NewConfirm().
				Title("Ignore shared system message?").
				Description("Skip including ~/.config/ayo/prompts/system.md").
				Value(&res.IgnoreSharedSystem),
		),

		// Step 4b: File path input (conditional)
		huh.NewGroup(
			huh.NewInput().
				Title("System message file").
				Placeholder("~/prompts/system.md").
				Value(&res.SystemFile).
				Validate(func(v string) error {
					if v == "" {
						return fmt.Errorf("file path required")
					}
					expanded := expandPath(v)
					if _, err := os.Stat(expanded); os.IsNotExist(err) {
						return fmt.Errorf("file not found: %s", v)
					}
					return nil
				}),
		).WithHideFunc(func() bool { return systemSource != "file" }),

		// Step 4c: Inline editor (conditional)
		huh.NewGroup(
			huh.NewText().
				Title("System message").
				Placeholder("You are a helpful assistant...").
				Description("Ctrl+E to open external editor").
				Value(&res.SystemMessage).
				CharLimit(0).
				Lines(8),
		).WithHideFunc(func() bool { return systemSource != "inline" }),

		// Step 5: Chaining toggle
		huh.NewGroup(
			huh.NewNote().
				Title("Step 5 of 5 ─ Agent Chaining").
				Description("Enable structured I/O for pipeline composition\n"),
			huh.NewConfirm().
				Title("Enable chaining?").
				Description("Add JSON schemas for input/output validation").
				Value(&enableChaining),
		),

		// Step 5b: Schema files (conditional)
		huh.NewGroup(
			huh.NewInput().
				Title("Input schema file").
				Placeholder("path/to/input.jsonschema (optional)").
				Description("Validates incoming JSON data").
				Value(&res.InputSchemaFile).
				Validate(validateOptionalSchemaFile),
			huh.NewInput().
				Title("Output schema file").
				Placeholder("path/to/output.jsonschema (optional)").
				Description("Structures outgoing JSON data").
				Value(&res.OutputSchemaFile).
				Validate(validateOptionalSchemaFile),
		).WithHideFunc(func() bool { return !enableChaining }),

		// Step 6: Review and confirm
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

	// Run the form with our wrapper
	p := tea.NewProgram(wizardModel{
		form: form,
	})

	finalModel, err := p.Run()
	if err != nil {
		return AgentCreateResult{}, err
	}

	wm := finalModel.(wizardModel)
	if wm.quitting && wm.form.State != huh.StateCompleted {
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

// renderSkillsTabs renders the tab bar for skill selection
func renderSkillsTabs(activeTab string, builtinCount, userCount int) string {
	// Tab styles using box drawing characters
	// Active tab: filled background effect with ◉
	// Inactive tab: dimmed with ○

	var builtinTab, userTab string

	if activeTab == "built-in" {
		builtinTab = fmt.Sprintf("▓▓ ◉ Built-in (%d) ▓▓", builtinCount)
		userTab = fmt.Sprintf("   ○ User (%d)     ", userCount)
	} else {
		builtinTab = fmt.Sprintf("   ○ Built-in (%d) ", builtinCount)
		userTab = fmt.Sprintf("▓▓ ◉ User (%d) ▓▓   ", userCount)
	}

	return fmt.Sprintf("Select skills to enable\n\n%s    %s\n", builtinTab, userTab)
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

	if res.IgnoreSharedSystem {
		b.WriteString("                 (ignoring shared system)\n")
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

// validateOptionalSchemaFile validates an optional schema file path.
func validateOptionalSchemaFile(v string) error {
	if v == "" {
		return nil
	}
	expanded := expandPath(v)
	if _, err := os.Stat(expanded); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", v)
	}
	return nil
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

// sourceDisplayName returns a user-friendly name for a skill source.
func sourceDisplayName(source skills.SkillSource) string {
	switch source {
	case skills.SourceBuiltIn:
		return "built-in"
	case skills.SourceUserShared:
		return "user"
	case skills.SourceAgentSpecific:
		return "agent"
	default:
		return "user"
	}
}
