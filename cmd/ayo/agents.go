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
	var sourceFilter string

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

				if len(handles) == 0 {
					fmt.Println("No agents found.")
					return nil
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
				handleStyle := lipgloss.NewStyle().Foreground(cyan).Bold(true)
				sourceStyle := lipgloss.NewStyle().Foreground(muted)
				descStyle := lipgloss.NewStyle().Foreground(text)
				countStyle := lipgloss.NewStyle().Foreground(muted)
				dividerStyle := lipgloss.NewStyle().Foreground(subtle)

				// Header
				fmt.Println()
				fmt.Println(headerStyle.Render("  Agents"))
				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))

				count := 0
				for _, h := range handles {
					// Determine source
					isBuiltin := builtin.HasAgent(h)
					source := "user"
					if isBuiltin {
						// Check if user has overridden
						userDir := filepath.Join(cfg.AgentsDir, h)
						if _, err := os.Stat(userDir); err == nil {
							source = "user"
						} else {
							source = "built-in"
						}
					}

					// Filter by source if specified
					if sourceFilter != "" && source != sourceFilter {
						continue
					}
					count++

					// Get description
					ag, err := agent.Load(cfg, h)
					desc := ""
					if err == nil {
						desc = ag.Config.Description
					}

					// Format: ◆ @handle                              [source]
					icon := iconStyle.Render("◆")
					handle := handleStyle.Render(h)
					sourceLabel := sourceStyle.Render(source)

					// Calculate padding for alignment
					handleWidth := lipgloss.Width(h)
					padding := 50 - handleWidth
					if padding < 2 {
						padding = 2
					}

					fmt.Printf("  %s %s%s%s\n", icon, handle, strings.Repeat(" ", padding), sourceLabel)

					// Description (truncated, indented)
					if desc != "" {
						if len(desc) > 52 {
							desc = desc[:49] + "..."
						}
						fmt.Printf("    %s\n", descStyle.Render(desc))
					}
				}

				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))
				fmt.Println(countStyle.Render(fmt.Sprintf("  %d agents", count)))
				fmt.Println()

				return nil
			})
		},
	}

	cmd.Flags().StringVar(&sourceFilter, "source", "", "filter by source (user, built-in)")

	return cmd
}

func createAgentCmd(cfgPath *string) *cobra.Command {
	var (
		model        string
		description  string
		system       string
		systemFile   string
		ignoreShared bool
	)

	cmd := &cobra.Command{
		Use:   "create [@handle]",
		Short: "Create a new agent",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var handle string
			if len(args) > 0 {
				handle = agent.NormalizeHandle(args[0])
			}

			return withConfig(cfgPath, func(cfg config.Config) error {
				providerModels := config.ConfiguredModels(cfg)
				modelIDs := make([]string, 0, len(providerModels))
				modelSet := make(map[string]struct{}, len(providerModels))
				for _, m := range providerModels {
					modelIDs = append(modelIDs, m.ID)
					modelSet[m.ID] = struct{}{}
				}

				// Interactive form if required fields missing
				if handle == "" || model == "" || (system == "" && systemFile == "") {
					ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
					defer cancel()

					form := ui.NewInitForm()
					res, err := form.Run(ctx, modelIDs)
					if err != nil {
						return err
					}
					handle = res.Handle
					model = res.Model
					description = res.Description
					ignoreShared = res.IgnoreShared
					system = res.System
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
					data, err := os.ReadFile(systemFile)
					if err != nil {
						return err
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

				agCfg := agent.Config{
					Model:                     model,
					Description:               description,
					IgnoreSharedSystemMessage: ignoreShared,
				}

				ag, err := agent.Save(cfg, handle, agCfg, system)
				if err != nil {
					return err
				}

				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render("✓ Created agent: " + ag.Handle))
				fmt.Printf("  Location: %s\n", ag.Dir)
				fmt.Println("  Edit system.md to customize your agent.")

				return nil
			})
		},
	}

	cmd.Flags().StringVar(&model, "model", "", "model to use")
	cmd.Flags().StringVar(&description, "description", "", "agent description")
	cmd.Flags().StringVar(&system, "system", "", "system message text")
	cmd.Flags().StringVar(&systemFile, "system-file", "", "path to system message file")
	cmd.Flags().BoolVar(&ignoreShared, "ignore-shared", false, "ignore shared system message")

	return cmd
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
