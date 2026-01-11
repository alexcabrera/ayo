package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"ayo/internal/agent"
	"ayo/internal/builtin"
	"ayo/internal/config"
	"ayo/internal/paths"
	"ayo/internal/skills"
	"ayo/internal/ui"
)

func newAgentsCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agents",
		Short:   "Manage agents",
		Aliases: []string{"agent"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to list
			return listAgentsCmd(cfgPath).RunE(cmd, args)
		},
	}

	cmd.AddCommand(listAgentsCmd(cfgPath))
	cmd.AddCommand(createAgentCmd(cfgPath))
	cmd.AddCommand(showAgentCmd(cfgPath))
	cmd.AddCommand(dirAgentCmd())
	cmd.AddCommand(updateAgentsCmd(cfgPath))

	return cmd
}

func listAgentsCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				// Ensure builtins are installed
				if err := builtin.Install(); err != nil {
					return fmt.Errorf("install builtins: %w", err)
				}

				handles, err := agent.ListHandles(cfg)
				if err != nil {
					return err
				}

				// Color palette
				purple := lipgloss.Color("#a78bfa")
				cyan := lipgloss.Color("#67e8f9")
				muted := lipgloss.Color("#6b7280")
				text := lipgloss.Color("#e5e7eb")
				subtle := lipgloss.Color("#374151")

				// Styles
				headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
				sectionStyle := lipgloss.NewStyle().Foreground(muted).Bold(true)
				iconStyle := lipgloss.NewStyle().Foreground(cyan)
				handleStyle := lipgloss.NewStyle().Foreground(cyan).Bold(true)
				descStyle := lipgloss.NewStyle().Foreground(text)
				countStyle := lipgloss.NewStyle().Foreground(muted)
				dividerStyle := lipgloss.NewStyle().Foreground(subtle)
				emptyStyle := lipgloss.NewStyle().Foreground(muted).Italic(true)

				// Categorize agents
				type agentInfo struct {
					handle string
					desc   string
				}
				var userAgents, builtinAgents []agentInfo

				for _, h := range handles {
					// Get description
					ag, err := agent.Load(cfg, h)
					desc := ""
					if err == nil {
						desc = ag.Config.Description
					}

					// Determine source
					isBuiltin := builtin.HasAgent(h)
					if isBuiltin {
						// Check if user has overridden
						userDir := filepath.Join(cfg.AgentsDir, h)
						if _, err := os.Stat(userDir); err == nil {
							userAgents = append(userAgents, agentInfo{h, desc})
						} else {
							builtinAgents = append(builtinAgents, agentInfo{h, desc})
						}
					} else {
						userAgents = append(userAgents, agentInfo{h, desc})
					}
				}

				// Render function for an agent
				renderAgent := func(a agentInfo) {
					icon := iconStyle.Render("◆")
					handle := handleStyle.Render(a.handle)
					fmt.Printf("  %s %s\n", icon, handle)

					// Description (truncated, indented)
					if a.desc != "" {
						desc := a.desc
						if len(desc) > 52 {
							desc = desc[:49] + "..."
						}
						fmt.Printf("    %s\n", descStyle.Render(desc))
					}
				}

				// Header
				fmt.Println()
				fmt.Println(headerStyle.Render("  Agents"))
				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))

				// User-defined agents section
				fmt.Println()
				fmt.Printf("  %s\n", sectionStyle.Render("User-defined"))
				if len(userAgents) == 0 {
					fmt.Printf("    %s\n", emptyStyle.Render("No user-defined agents"))
					fmt.Printf("    %s\n", emptyStyle.Render("Create one with: ayo agents create @name"))
				} else {
					for _, a := range userAgents {
						renderAgent(a)
					}
				}

				// Built-in agents section
				fmt.Println()
				fmt.Printf("  %s\n", sectionStyle.Render("Built-in"))
				if len(builtinAgents) == 0 {
					fmt.Printf("    %s\n", emptyStyle.Render("No built-in agents installed"))
					fmt.Printf("    %s\n", emptyStyle.Render("Run: ayo setup"))
				} else {
					for _, a := range builtinAgents {
						renderAgent(a)
					}
				}

				fmt.Println()
				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))
				fmt.Println(countStyle.Render(fmt.Sprintf("  %d agents", len(handles))))
				fmt.Println()

				return nil
			})
		},
	}

	return cmd
}

func createAgentCmd(cfgPath *string) *cobra.Command {
	var (
		// Core
		model       string
		description string
		system      string
		systemFile  string

		// Tools
		tools []string

		// Skills
		skills_             []string
		excludeSkills       []string
		ignoreBuiltinSkills bool
		ignoreSharedSkills  bool

		// Chaining
		inputSchema  string
		outputSchema string

		// Dev mode
		devMode bool
	)

	cmd := &cobra.Command{
		Use:   "create [@handle]",
		Short: "Create a new agent",
		Long: `Create a new agent with the specified configuration.

If required flags are not provided, an interactive wizard will guide you through
the creation process. For non-interactive use, provide all required flags.

Examples:
  # Interactive mode
  ayo agents create @myagent

  # Full CLI creation
  ayo agents create @code-helper \
    --model gpt-4.1 \
    --description "Helps write clean code" \
    --system "You are an expert programmer..." \
    --tools bash,agent_call \
    --skills debugging

  # Using external files
  ayo agents create @reviewer \
    --model claude-3.5-sonnet \
    --system-file ~/prompts/reviewer.md \
    --input-schema ~/schemas/input.json \
    --output-schema ~/schemas/output.json

  # Create in local project directory for testing
  ayo agents create @myagent --dev`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var handle string
			if len(args) > 0 {
				handle = agent.NormalizeHandle(args[0])
			}

			// If dev mode, use local config directory
			if devMode {
				paths.SetLocalDevMode()
			}

			return withConfig(cfgPath, func(cfg config.Config) error {
				providerModels := config.ConfiguredModels(cfg)
				modelIDs := make([]string, 0, len(providerModels))
				modelSet := make(map[string]struct{}, len(providerModels))
				for _, m := range providerModels {
					modelIDs = append(modelIDs, m.ID)
					modelSet[m.ID] = struct{}{}
				}

				// Determine if we need interactive mode
				needsInteractive := handle == "" || model == "" || (system == "" && systemFile == "")

				if needsInteractive {
					ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
					defer cancel()

					// Discover available skills for the form with proper source tagging
					discoveredSkills := skills.DiscoverAll(skills.DiscoveryOptions{
						UserSharedDir: cfg.SkillsDir,
						BuiltinDir:    builtin.SkillsInstallDir(),
					})

					// Available tools
					availableTools := []string{"bash", "agent_call"}

					form := ui.NewAgentCreateForm(ui.AgentCreateFormOptions{
						Models:          modelIDs,
						AvailableSkills: discoveredSkills.Skills,
						AvailableTools:  availableTools,
						PrefilledHandle: handle,
					})

					res, err := form.Run(ctx)
					if err != nil {
						return err
					}

					// Transfer results from form
					handle = res.Handle
					model = res.Model
					description = res.Description
					tools = res.AllowedTools
					skills_ = res.Skills
					ignoreBuiltinSkills = res.IgnoreBuiltinSkills
					ignoreSharedSkills = res.IgnoreSharedSkills
					system = res.SystemMessage
					inputSchema = res.InputSchemaFile
					outputSchema = res.OutputSchemaFile
				}

				if handle == "" {
					return fmt.Errorf("handle is required")
				}

				// Check reserved namespace
				if agent.IsReservedNamespace(handle) {
					return fmt.Errorf("cannot use reserved handle %s", handle)
				}

				// Check if already exists
				agentDir := filepath.Join(cfg.AgentsDir, handle)
				if _, err := os.Stat(agentDir); err == nil {
					return fmt.Errorf("agent already exists: %s", handle)
				}

				// Load system from file if specified
				if system == "" && systemFile != "" {
					expanded := expandPath(systemFile)
					data, err := os.ReadFile(expanded)
					if err != nil {
						return fmt.Errorf("read system file: %w", err)
					}
					system = string(data)
				}

				// Default system message
				if system == "" {
					system = "You are a helpful assistant."
				}

				// Default model
				if model == "" {
					model = cfg.DefaultModel
				}

				// Validate model if we have a configured set
				if len(modelSet) > 0 {
					if _, ok := modelSet[model]; !ok {
						return fmt.Errorf("model %s is not configured", model)
					}
				}

				// Default tools only for non-interactive mode (CLI flags)
				// Interactive mode explicitly sets tools from wizard selection
				if !needsInteractive && len(tools) == 0 {
					tools = []string{"bash"}
				}

				agCfg := agent.Config{
					Model:               model,
					Description:         description,
					AllowedTools:        tools,
					Skills:              skills_,
					ExcludeSkills:       excludeSkills,
					IgnoreBuiltinSkills: ignoreBuiltinSkills,
					IgnoreSharedSkills:  ignoreSharedSkills,
				}

				ag, err := agent.SaveWithSchemas(cfg, handle, agCfg, system, inputSchema, outputSchema)
				if err != nil {
					return err
				}

				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render("✓ Created agent: " + ag.Handle))
				fmt.Printf("  Location: %s\n", ag.Dir)

				// Show what was configured (from config, not resolved)
				if len(ag.Config.AllowedTools) > 0 {
					fmt.Printf("  Tools: %s\n", strings.Join(ag.Config.AllowedTools, ", "))
				}
				if len(ag.Config.Skills) > 0 {
					fmt.Printf("  Skills: %s\n", strings.Join(ag.Config.Skills, ", "))
				}
				if ag.HasInputSchema() || ag.HasOutputSchema() {
					fmt.Println("  Chaining: enabled")
				}

				return nil
			})
		},
	}

	// Core flags
	cmd.Flags().StringVarP(&model, "model", "m", "", "model to use")
	cmd.Flags().StringVarP(&description, "description", "d", "", "agent description")
	cmd.Flags().StringVarP(&system, "system", "s", "", "system message text")
	cmd.Flags().StringVarP(&systemFile, "system-file", "f", "", "path to system message markdown file")

	// Tool flags
	cmd.Flags().StringSliceVarP(&tools, "tools", "t", nil, "allowed tools (comma-separated or repeated)")

	// Skill flags
	cmd.Flags().StringSliceVar(&skills_, "skills", nil, "skills to include (comma-separated or repeated)")
	cmd.Flags().StringSliceVar(&excludeSkills, "exclude-skills", nil, "skills to exclude")
	cmd.Flags().BoolVar(&ignoreBuiltinSkills, "ignore-builtin-skills", false, "ignore built-in skills")
	cmd.Flags().BoolVar(&ignoreSharedSkills, "ignore-shared-skills", false, "ignore shared skills")

	// Schema flags
	cmd.Flags().StringVar(&inputSchema, "input-schema", "", "path to input JSON schema file (for chaining)")
	cmd.Flags().StringVar(&outputSchema, "output-schema", "", "path to output JSON schema file (for chaining)")

	// Dev mode flag
	cmd.Flags().BoolVar(&devMode, "dev", false, "create agent in local ./.config/ayo/ directory for testing")

	return cmd
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

func showAgentCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <handle>",
		Short: "Show agent details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			handle := agent.NormalizeHandle(args[0])

			return withConfig(cfgPath, func(cfg config.Config) error {
				// Ensure builtins are installed
				if err := builtin.Install(); err != nil {
					return fmt.Errorf("install builtins: %w", err)
				}

				ag, err := agent.Load(cfg, handle)
				if err != nil {
					return fmt.Errorf("agent not found: %s", handle)
				}

				// Color palette
				purple := lipgloss.Color("#a78bfa")
				cyan := lipgloss.Color("#67e8f9")
				muted := lipgloss.Color("#6b7280")
				text := lipgloss.Color("#e5e7eb")
				subtle := lipgloss.Color("#374151")

				// Styles
				headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
				iconStyle := lipgloss.NewStyle().Foreground(cyan)
				labelStyle := lipgloss.NewStyle().Foreground(muted)
				valueStyle := lipgloss.NewStyle().Foreground(text)
				dividerStyle := lipgloss.NewStyle().Foreground(subtle)

				fmt.Println()
				fmt.Println("  " + iconStyle.Render("◆") + " " + headerStyle.Render(ag.Handle))
				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))

				source := "user"
				if ag.BuiltIn {
					source = "built-in"
				}
				fmt.Printf("  %s %s\n", labelStyle.Render("Source:"), valueStyle.Render(source))
				fmt.Printf("  %s  %s\n", labelStyle.Render("Model:"), valueStyle.Render(ag.Model))

				if ag.Config.Description != "" {
					fmt.Printf("  %s   %s\n", labelStyle.Render("Desc:"), valueStyle.Render(ag.Config.Description))
				}

				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))
				fmt.Printf("  %s   %s\n", labelStyle.Render("Path:"), valueStyle.Render(ag.Dir))

				if len(ag.Skills) > 0 {
					skillNames := make([]string, len(ag.Skills))
					for i, s := range ag.Skills {
						skillNames[i] = s.Name
					}
					sort.Strings(skillNames)
					fmt.Printf("  %s %s\n", labelStyle.Render("Skills:"), valueStyle.Render(strings.Join(skillNames, ", ")))
				}

				if len(ag.Config.AllowedTools) > 0 {
					fmt.Printf("  %s  %s\n", labelStyle.Render("Tools:"), valueStyle.Render(strings.Join(ag.Config.AllowedTools, ", ")))
				}

				fmt.Println()

				return nil
			})
		},
	}

	return cmd
}

func dirAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "dir",
		Short:  "Show agents directories",
		Long:   "Shows paths to user and built-in agent directories.",
		Hidden: false,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Agent directories:")
			fmt.Printf("  User:     %s\n", paths.AgentsDir())
			fmt.Printf("  Built-in: %s\n", builtin.InstallDir())
			return nil
		},
	}

	return cmd
}

func updateAgentsCmd(cfgPath *string) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update built-in agents to latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				sui := newSetupUI(cmd.OutOrStdout())

				if !force {
					// Check for modified agents
					modified, err := builtin.CheckModifiedAgents()
					if err != nil {
						return fmt.Errorf("check modified agents: %w", err)
					}

					if len(modified) > 0 {
						sui.Warning("The following agents have local modifications:")
						for _, m := range modified {
							sui.Info(fmt.Sprintf("  %s: %v", m.Handle, m.ModifiedFiles))
						}
						sui.Blank()
						sui.Info("Use --force to overwrite, or copy modifications to user directory first:")
						sui.Info(fmt.Sprintf("  %s", cfg.AgentsDir))
						return fmt.Errorf("agents have local modifications")
					}
				}

				sui.Step("Updating built-in agents...")
				installDir, err := builtin.ForceInstall()
				if err != nil {
					return err
				}
				sui.SuccessPath("Updated agents at", installDir)
				return nil
			})
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "overwrite without checking for modifications")

	return cmd
}
