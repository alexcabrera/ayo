package ui

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"

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

// Run executes the multi-step form wizard using huh.Form with groups.
func (f *AgentCreateForm) Run(ctx context.Context) (AgentCreateResult, error) {
	if len(f.opts.Models) == 0 {
		return AgentCreateResult{}, fmt.Errorf("no models configured; update config provider models")
	}

	var res AgentCreateResult

	// Handle without @ prefix
	var handleName string
	if strings.HasPrefix(f.opts.PrefilledHandle, "@") {
		handleName = strings.TrimPrefix(f.opts.PrefilledHandle, "@")
	} else {
		handleName = f.opts.PrefilledHandle
	}

	// Default tools selection
	res.AllowedTools = []string{"bash"}

	// Tracking variables for conditional groups
	var systemSource string = "inline"
	var enableChaining bool

	// Build model options
	modelOpts := make([]huh.Option[string], 0, len(f.opts.Models))
	for _, m := range f.opts.Models {
		modelOpts = append(modelOpts, huh.NewOption(m, m))
	}

	// Build tool options with multi-line labels showing full description
	toolDescriptions := map[string]string{
		"bash":       "Execute shell commands to interact with the system",
		"agent_call": "Delegate tasks to other specialized agents",
	}
	toolOpts := make([]huh.Option[string], 0, len(f.opts.AvailableTools))
	for _, t := range f.opts.AvailableTools {
		desc := toolDescriptions[t]
		if desc == "" {
			desc = t
		}
		// Multi-line label: name on first line, description indented on second
		label := fmt.Sprintf("%s\n    %s", t, desc)
		toolOpts = append(toolOpts, huh.NewOption(label, t))
	}

	// Build skill options with multi-line labels showing full description
	skillOpts := make([]huh.Option[string], 0, len(f.opts.AvailableSkills))
	for _, s := range f.opts.AvailableSkills {
		// Multi-line label: name on first line, description indented on second
		// Wrap long descriptions at ~60 chars
		desc := wrapText(s.Description, 60, "    ")
		label := fmt.Sprintf("%s\n%s", s.Name, desc)
		skillOpts = append(skillOpts, huh.NewOption(label, s.Name))
	}

	hasSkills := len(f.opts.AvailableSkills) > 0

	// Calculate step numbers dynamically
	totalSteps := 5
	if !hasSkills {
		totalSteps = 4
	}

	skillsStepNum := 3
	systemStepNum := 4
	chainingStepNum := 5
	if !hasSkills {
		systemStepNum = 3
		chainingStepNum = 4
	}

	// Create the form with all groups
	groups := []*huh.Group{
		// Step 1: Identity
		huh.NewGroup(
			huh.NewNote().
				Title(fmt.Sprintf("Step 1 of %d ─ Identity", totalSteps)).
				Description("Give your agent a unique handle (like @myagent) that you'll use to\ninvoke it. The description helps you remember what this agent does."),
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
				Title(fmt.Sprintf("Step 2 of %d ─ Tools", totalSteps)).
				Description("Tools let your agent interact with the outside world. The bash tool\nallows running shell commands. Use x/space to toggle selection."),
			huh.NewMultiSelect[string]().
				Title("Allowed Tools").
				Options(toolOpts...).
				Value(&res.AllowedTools),
		),
	}

	// Step 3: Skills (conditional)
	if hasSkills {
		groups = append(groups,
			huh.NewGroup(
				huh.NewNote().
					Title(fmt.Sprintf("Step %d of %d ─ Skills", skillsStepNum, totalSteps)).
					Description("Skills are reusable instruction sets that teach your agent specialized\ntasks. Select any that match what you want this agent to do."),
				huh.NewMultiSelect[string]().
					Title("Available Skills").
					Options(skillOpts...).
					Value(&res.Skills).
					Filterable(true),
			),
		)
	}

	// Step: System Prompt Source
	groups = append(groups,
		huh.NewGroup(
			huh.NewNote().
				Title(fmt.Sprintf("Step %d of %d ─ System Prompt", systemStepNum, totalSteps)).
				Description("The system prompt defines your agent's personality, knowledge, and\nbehavior. You can write it inline or load from an existing file."),
			huh.NewSelect[string]().
				Title("Source").
				Options(
					huh.NewOption("Enter inline", "inline"),
					huh.NewOption("Browse for file", "file"),
				).
				Value(&systemSource),
		),
	)

	// Get editor name for help text
	editorName := os.Getenv("EDITOR")
	if editorName == "" {
		editorName = "vim"
	}

	// Step: System Prompt Content (inline) - same step number
	groups = append(groups,
		huh.NewGroup(
			huh.NewNote().
				Title(fmt.Sprintf("Step %d of %d ─ System Prompt", systemStepNum, totalSteps)).
				Description("Enter the system message directly"),
			huh.NewText().
				Title("System message").
				Placeholder("You are a helpful assistant...").
				Description(fmt.Sprintf("ctrl+e to open in %s • alt+enter for newline", editorName)).
				Value(&res.SystemMessage).
				CharLimit(0).
				Lines(10).
				Editor(editorName),
		).WithHideFunc(func() bool {
			return systemSource != "inline"
		}),
	)

	// Step: System Prompt Review (inline) - confirm before proceeding
	var confirmSystemInline bool
	groups = append(groups,
		huh.NewGroup(
			huh.NewNote().
				Title(fmt.Sprintf("Step %d of %d ─ System Prompt", systemStepNum, totalSteps)).
				DescriptionFunc(func() string {
					if res.SystemMessage == "" {
						return "No system message entered (will use default)"
					}
					rendered := renderMarkdownPreview(res.SystemMessage, 60)
					return fmt.Sprintf("Preview:\n%s", rendered)
				}, &res.SystemMessage),
			huh.NewConfirm().
				Title("Continue with this system prompt?").
				Affirmative("Yes").
				Negative("No, go back").
				Value(&confirmSystemInline),
		).WithHideFunc(func() bool {
			return systemSource != "inline"
		}),
	)

	// Step: System Prompt Content (file) - directly show file picker
	groups = append(groups,
		huh.NewGroup(
			huh.NewFilePicker().
				Title(fmt.Sprintf("Step %d of %d ─ System Prompt", systemStepNum, totalSteps)).
				Description("← or backspace to go up a directory").
				CurrentDirectory(currentDir()).
				AllowedTypes([]string{".md", ".txt"}).
				Value(&res.SystemFile).
				Picking(true).
				Height(15),
		).WithHideFunc(func() bool {
			return systemSource != "file"
		}),
	)

	// Step: System Prompt File Preview - show file contents before confirming
	var confirmSystemFile bool
	groups = append(groups,
		huh.NewGroup(
			huh.NewNote().
				Title(fmt.Sprintf("Step %d of %d ─ System Prompt", systemStepNum, totalSteps)).
				DescriptionFunc(func() string {
					if res.SystemFile == "" {
						return "No file selected\n\nshift+tab to go back and select a file"
					}
					content, err := readFileContent(res.SystemFile)
					if err != nil {
						return fmt.Sprintf("Error reading file: %v\n\nshift+tab to go back", err)
					}
					rendered := renderMarkdownPreview(content, 60)
					return fmt.Sprintf("File: %s\n%s", shortenPath(res.SystemFile), rendered)
				}, &res.SystemFile),
			huh.NewConfirm().
				Title("Use this file?").
				Affirmative("Yes").
				Negative("No, go back").
				Value(&confirmSystemFile),
		).WithHideFunc(func() bool {
			return systemSource != "file"
		}),
	)

	// Step: Chaining
	groups = append(groups,
		huh.NewGroup(
			huh.NewNote().
				Title(fmt.Sprintf("Step %d of %d ─ Agent Chaining", chainingStepNum, totalSteps)).
				Description("Chaining lets you connect agents together using Unix pipes, where one\nagent's output becomes another's input. This is useful for building\nmulti-step workflows like: analyze → summarize → format.\n\nTo enable chaining, your agent needs JSON schemas that define the\nstructure of its input and output data."),
			huh.NewConfirm().
				Title("Enable chaining for this agent?").
				Description("You can add this later by creating schema files in the agent directory").
				Value(&enableChaining),
		),
	)

	// Step: Input Schema (conditional on chaining)
	groups = append(groups,
		huh.NewGroup(
			huh.NewFilePicker().
				Title(fmt.Sprintf("Step %d of %d ─ Input Schema", chainingStepNum, totalSteps)).
				Description("← to go up • enter to select or skip").
				CurrentDirectory(currentDir()).
				AllowedTypes([]string{".json", ".jsonschema"}).
				Value(&res.InputSchemaFile).
				Picking(true).
				Height(12),
		).WithHideFunc(func() bool {
			return !enableChaining
		}),
	)

	// Step: Input Schema Preview - show file contents before confirming
	var confirmInputSchema bool
	groups = append(groups,
		huh.NewGroup(
			huh.NewNote().
				Title(fmt.Sprintf("Step %d of %d ─ Input Schema", chainingStepNum, totalSteps)).
				DescriptionFunc(func() string {
					if res.InputSchemaFile == "" {
						return "No file selected (skipping input schema)"
					}
					content, err := readFilePreview(res.InputSchemaFile, 15)
					if err != nil {
						return fmt.Sprintf("Error reading file: %v", err)
					}
					return fmt.Sprintf("File: %s\n\n%s", shortenPath(res.InputSchemaFile), content)
				}, &res.InputSchemaFile),
			huh.NewConfirm().
				Title("Use this file?").
				Affirmative("Yes").
				Negative("No, go back").
				Value(&confirmInputSchema),
		).WithHideFunc(func() bool {
			return !enableChaining || res.InputSchemaFile == ""
		}),
	)

	// Step: Output Schema (conditional on chaining)
	groups = append(groups,
		huh.NewGroup(
			huh.NewFilePicker().
				Title(fmt.Sprintf("Step %d of %d ─ Output Schema", chainingStepNum, totalSteps)).
				Description("← to go up • enter to select or skip").
				CurrentDirectory(currentDir()).
				AllowedTypes([]string{".json", ".jsonschema"}).
				Value(&res.OutputSchemaFile).
				Picking(true).
				Height(12),
		).WithHideFunc(func() bool {
			return !enableChaining
		}),
	)

	// Step: Output Schema Preview - show file contents before confirming
	var confirmOutputSchema bool
	groups = append(groups,
		huh.NewGroup(
			huh.NewNote().
				Title(fmt.Sprintf("Step %d of %d ─ Output Schema", chainingStepNum, totalSteps)).
				DescriptionFunc(func() string {
					if res.OutputSchemaFile == "" {
						return "No file selected (skipping output schema)"
					}
					content, err := readFilePreview(res.OutputSchemaFile, 15)
					if err != nil {
						return fmt.Sprintf("Error reading file: %v", err)
					}
					return fmt.Sprintf("File: %s\n\n%s", shortenPath(res.OutputSchemaFile), content)
				}, &res.OutputSchemaFile),
			huh.NewConfirm().
				Title("Use this file?").
				Affirmative("Yes").
				Negative("No, go back").
				Value(&confirmOutputSchema),
		).WithHideFunc(func() bool {
			return !enableChaining || res.OutputSchemaFile == ""
		}),
	)

	// Step: Review and Confirm
	var reviewChoice string
	groups = append(groups,
		huh.NewGroup(
			huh.NewNote().
				Title("Review Configuration").
				DescriptionFunc(func() string {
					return buildReviewSummary(handleName, res, systemSource, enableChaining)
				}, &res),
			huh.NewSelect[string]().
				Title("What would you like to do?").
				Options(
					huh.NewOption("Create this agent", "create"),
					huh.NewOption("Cancel", "cancel"),
				).
				Value(&reviewChoice),
		),
	)

	// Step: Cancel Confirmation
	var confirmCancel bool
	groups = append(groups,
		huh.NewGroup(
			huh.NewConfirm().
				Title("Are you sure you want to cancel?").
				Description("All entered information will be lost.").
				Affirmative("Yes, cancel").
				Negative("No, go back").
				Value(&confirmCancel),
		).WithHideFunc(func() bool {
			return reviewChoice != "cancel"
		}),
	)

	// Create and run the form in alternate screen mode for clean TUI
	form := huh.NewForm(groups...).
		WithTheme(huh.ThemeCharm()).
		WithShowHelp(true).
		WithProgramOptions(tea.WithAltScreen())

	err := form.Run()
	if err != nil {
		if err.Error() == "user aborted" {
			return AgentCreateResult{}, fmt.Errorf("agent creation cancelled")
		}
		return AgentCreateResult{}, err
	}

	// Check if user cancelled
	if reviewChoice == "cancel" && confirmCancel {
		return AgentCreateResult{}, fmt.Errorf("agent creation cancelled")
	}
	if reviewChoice != "create" {
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
func buildReviewSummary(handleName string, res AgentCreateResult, systemSource string, enableChaining bool) string {
	var b strings.Builder

	handle := "@" + strings.TrimPrefix(handleName, "@")

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Handle:        %s\n", handle))
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
		path := res.SystemFile
		if home := userHomeDir(); home != "" {
			path = strings.Replace(path, home, "~", 1)
		}
		b.WriteString(fmt.Sprintf("  System:        %s\n", path))
	} else if res.SystemMessage != "" {
		preview := res.SystemMessage
		if len(preview) > 35 {
			preview = preview[:32] + "..."
		}
		preview = strings.ReplaceAll(preview, "\n", " ")
		b.WriteString(fmt.Sprintf("  System:        %s\n", preview))
	}

	if enableChaining {
		b.WriteString("  Chaining:      enabled\n")
		if res.InputSchemaFile != "" {
			path := res.InputSchemaFile
			if home := userHomeDir(); home != "" {
				path = strings.Replace(path, home, "~", 1)
			}
			b.WriteString(fmt.Sprintf("    Input:       %s\n", path))
		}
		if res.OutputSchemaFile != "" {
			path := res.OutputSchemaFile
			if home := userHomeDir(); home != "" {
				path = strings.Replace(path, home, "~", 1)
			}
			b.WriteString(fmt.Sprintf("    Output:      %s\n", path))
		}
	}

	b.WriteString("\n")

	return b.String()
}

// userHomeDir returns the user's home directory or empty string if unavailable.
func userHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// currentDir returns the current working directory or home directory as fallback.
func currentDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return userHomeDir()
}

// wrapText wraps text at the specified width, adding a prefix to each line.
func wrapText(text string, width int, prefix string) string {
	if len(text) <= width {
		return prefix + text
	}

	var lines []string
	words := strings.Fields(text)
	var currentLine string

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, prefix+currentLine)
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, prefix+currentLine)
	}

	return strings.Join(lines, "\n")
}

// readFilePreview reads a file and returns the first maxLines lines for preview.
func readFilePreview(path string, maxLines int) (string, error) {
	expanded := expandPath(path)
	data, err := os.ReadFile(expanded)
	if err != nil {
		return "", err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, fmt.Sprintf("... (%d more lines)", len(strings.Split(content, "\n"))-maxLines))
	}

	return strings.Join(lines, "\n"), nil
}

// shortenPath shortens a path by replacing home directory with ~.
func shortenPath(path string) string {
	home := userHomeDir()
	if home != "" && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

// readFileContent reads the full content of a file.
func readFileContent(path string) (string, error) {
	expanded := expandPath(path)
	data, err := os.ReadFile(expanded)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// renderMarkdownPreview renders markdown content using glamour with word wrap.
func renderMarkdownPreview(content string, width int) string {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		// Fallback to plain text if glamour fails
		return content
	}

	rendered, err := r.Render(content)
	if err != nil {
		return content
	}

	return strings.TrimSpace(rendered)
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
