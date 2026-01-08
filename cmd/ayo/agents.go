package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"ayo/internal/agent"
	"ayo/internal/builtin"
	"ayo/internal/config"
	"ayo/internal/paths"
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

				// Styles
				headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
				agentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
				sourceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
				descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

				fmt.Println(headerStyle.Render("Available Agents"))
				fmt.Println(strings.Repeat("─", 60))

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

					// Get description
					ag, err := agent.Load(cfg, h)
					desc := ""
					if err == nil {
						desc = ag.Config.Description
					}

					// Agent name and source
					name := agentStyle.Render("🤖 " + h)
					sourceLabel := sourceStyle.Render(fmt.Sprintf("[%s]", source))
					fmt.Printf("%-45s %s\n", name, sourceLabel)

					// Description (truncated)
					if desc != "" {
						if len(desc) > 55 {
							desc = desc[:52] + "..."
						}
						fmt.Printf("   %s\n", descStyle.Render(desc))
					}
				}

				fmt.Println(strings.Repeat("─", 60))
				fmt.Printf("%d agents available\n", len(handles))

				return nil
			})
		},
	}

	cmd.Flags().StringVar(&sourceFilter, "source", "", "filter by source (user, built-in)")

	return cmd
}

func createAgentCmd(cfgPath *string) *cobra.Command {
	var (
		model       string
		description string
		system      string
	)

	cmd := &cobra.Command{
		Use:   "create <handle>",
		Short: "Create a new agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			handle := agent.NormalizeHandle(args[0])

			return withConfig(cfgPath, func(cfg config.Config) error {
				// Check reserved namespace
				if agent.IsReservedNamespace(handle) {
					return fmt.Errorf("cannot use reserved handle %s", handle)
				}

				// Check if already exists
				agentDir := filepath.Join(cfg.AgentsDir, handle)
				if _, err := os.Stat(agentDir); err == nil {
					return fmt.Errorf("agent already exists: %s", handle)
				}

				// Default system message
				if system == "" {
					system = "You are a helpful assistant."
				}

				// Default model
				if model == "" {
					model = cfg.DefaultModel
				}

				agCfg := agent.Config{
					Model:       model,
					Description: description,
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
	cmd.Flags().StringVar(&system, "system", "", "system message")

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

				// Styles
				headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
				labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
				valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

				fmt.Println()
				fmt.Println(headerStyle.Render("🤖 " + ag.Handle))
				fmt.Println(strings.Repeat("─", 60))

				source := "user"
				if ag.BuiltIn {
					source = "built-in"
				}
				fmt.Printf("%s %s\n", labelStyle.Render("Source:"), valueStyle.Render(source))
				fmt.Printf("%s %s\n", labelStyle.Render("Model:"), valueStyle.Render(ag.Model))

				if ag.Config.Description != "" {
					fmt.Printf("%s %s\n", labelStyle.Render("Description:"), valueStyle.Render(ag.Config.Description))
				}

				fmt.Println(strings.Repeat("─", 60))
				fmt.Printf("%s %s\n", labelStyle.Render("Location:"), valueStyle.Render(ag.Dir))

				if len(ag.Skills) > 0 {
					skillNames := make([]string, len(ag.Skills))
					for i, s := range ag.Skills {
						skillNames[i] = s.Name
					}
					sort.Strings(skillNames)
					fmt.Printf("%s %s\n", labelStyle.Render("Skills:"), valueStyle.Render(strings.Join(skillNames, ", ")))
				}

				if len(ag.Config.AllowedTools) > 0 {
					fmt.Printf("%s %s\n", labelStyle.Render("Tools:"), valueStyle.Render(strings.Join(ag.Config.AllowedTools, ", ")))
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
		Short:  "Go to agents directory (requires shell integration)",
		Long:   "Opens an interactive picker to choose between user and built-in agent directories.\nRequires shell integration: run `ayo setup` first.",
		Hidden: false,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// This is the fallback when shell integration is not set up
			fmt.Println("This command requires shell integration.")
			fmt.Println()
			fmt.Println("Run `ayo setup` to configure shell integration.")
			fmt.Println()
			fmt.Println("Agent directories:")
			fmt.Printf("  User:     %s\n", paths.AgentsDir())
			fmt.Printf("  Built-in: %s\n", builtin.InstallDir())
			return nil
		},
	}

	// Hidden subcommand for shell integration
	cmd.AddCommand(&cobra.Command{
		Use:    "pick",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Output choices for gum to pick
			fmt.Println("user")
			fmt.Println("built-in")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:    "path",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "built-in":
				fmt.Print(builtin.InstallDir())
			case "user":
				fmt.Print(paths.AgentsDir())
			default:
				return fmt.Errorf("unknown choice: %s", args[0])
			}
			return nil
		},
	})

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
