package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"ayo/internal/agent"
	"ayo/internal/builtin"
	"ayo/internal/config"
	"ayo/internal/paths"
	"ayo/internal/pipe"
	"ayo/internal/run"
)

func newRootCmd() *cobra.Command {
	var cfgPath string
	var attachments []string
	var debug bool

	cmd := &cobra.Command{
		Use:           "ayo [@agent] [prompt]",
		Short:         "Run AI agents",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Auto-install built-in agents and skills if needed (version-based)
			return builtin.Install()
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// Only complete first arg (agent handle)
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			cfg, err := loadConfig(cfgPath)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			handles, err := agent.ListHandles(cfg)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			// Filter by prefix if user has typed something
			var matches []string
			for _, h := range handles {
				if strings.HasPrefix(h, toComplete) {
					matches = append(matches, h)
				}
			}
			return matches, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(&cfgPath, func(cfg config.Config) error {
				if len(args) == 0 {
					// No args: show help
					return cmd.Help()
				}

				// Determine agent handle and remaining args
				var handle string
				var promptArgs []string

				if strings.HasPrefix(args[0], "@") {
					// First arg is an agent handle
					handle = agent.NormalizeHandle(args[0])
					promptArgs = args[1:]
				} else {
					// First arg is not an agent handle: use default agent with all args as prompt
					handle = agent.DefaultAgent
					promptArgs = args
				}

				ag, err := agent.Load(cfg, handle)
				if err != nil {
					return err
				}

				runner, err := run.NewRunnerFromConfig(cfg, debug)
				if err != nil {
					return err
				}

				// Non-interactive mode: prompt provided as positional args or stdin
				if len(promptArgs) > 0 || pipe.IsStdinPiped() {
					var prompt string

					if pipe.IsStdinPiped() {
						// Read from stdin
						stdinData, err := pipe.ReadStdin()
						if err != nil {
							return fmt.Errorf("read stdin: %w", err)
						}
						stdinData = strings.TrimSpace(stdinData)

						if ag.HasInputSchema() {
							// Agent has input schema: stdin must be valid JSON matching schema
							if err := ag.ValidateInput(stdinData); err != nil {
								return formatInputValidationError(err, stdinData, ag)
							}
							prompt = stdinData
						} else {
							// Agent has no input schema: build preamble with context
							prompt = buildFreeformPreamble(stdinData)
						}

						// If there are also positional args, append them
						if len(promptArgs) > 0 {
							prompt = prompt + "\n\n" + strings.Join(promptArgs, " ")
						}
					} else {
						// No stdin, use positional args
						prompt = strings.Join(promptArgs, " ")

						// Validate input against schema if agent has one
						if err := ag.ValidateInput(prompt); err != nil {
							return err
						}
					}

					ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
					defer cancel()

					resp, err := runner.Text(ctx, ag, prompt, attachments)
					if err != nil {
						return err
					}

					// Output to stdout (for piping)
					fmt.Println(resp)
					return nil
				}

				// Interactive mode
				return runInteractiveChat(cmd.Context(), runner, ag, debug)
			})
		},
	}

	cmd.PersistentFlags().StringVar(&cfgPath, "config", defaultConfigPath(), "path to config file")
	cmd.Flags().StringSliceVarP(&attachments, "attachment", "a", nil, "file attachments")
	cmd.Flags().BoolVar(&debug, "debug", false, "show debug output including raw tool payloads")

	// Subcommands
	cmd.AddCommand(newSetupCmd(&cfgPath))
	cmd.AddCommand(newInitShellCmd())
	cmd.AddCommand(newAgentsCmd(&cfgPath))
	cmd.AddCommand(newSkillsCmd(&cfgPath))
	cmd.AddCommand(newChainCmd(&cfgPath))

	return cmd
}

func defaultConfigPath() string {
	return paths.ConfigFile()
}

func loadConfig(cfgPath string) (config.Config, error) {
	return config.Load(cfgPath)
}

func withConfig(cfgPath *string, fn func(config.Config) error) error {
	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		return err
	}
	return fn(cfg)
}

// formatInputValidationError creates a detailed error message for input validation failures.
func formatInputValidationError(err error, receivedInput string, ag agent.Agent) error {
	ctx := pipe.GetChainContext()
	var sourceInfo string
	if ctx != nil && ctx.Source != "" {
		sourceInfo = fmt.Sprintf("\nReceived from: %s", ctx.Source)
		if ctx.SourceDescription != "" {
			sourceInfo += fmt.Sprintf(" (%s)", ctx.SourceDescription)
		}
	}

	// Truncate input if too long
	displayInput := receivedInput
	if len(displayInput) > 500 {
		displayInput = displayInput[:500] + "\n... (truncated)"
	}

	return fmt.Errorf("input validation failed for %s%s\n\nReceived:\n%s\n\nError: %w",
		ag.Handle, sourceInfo, displayInput, err)
}

// buildFreeformPreamble creates a preamble for agents without input schemas
// when receiving piped input from another agent.
func buildFreeformPreamble(jsonInput string) string {
	ctx := pipe.GetChainContext()

	var preamble strings.Builder
	preamble.WriteString("You received structured output from a previous agent in a chain.\n\n")

	if ctx != nil {
		if ctx.Source != "" {
			preamble.WriteString(fmt.Sprintf("Source agent: %s\n", ctx.Source))
		}
		if ctx.SourceDescription != "" {
			preamble.WriteString(fmt.Sprintf("Description: %s\n", ctx.SourceDescription))
		}
		preamble.WriteString(fmt.Sprintf("Chain depth: %d\n", ctx.Depth))
		preamble.WriteString("\n")
	}

	preamble.WriteString("The output is provided below as JSON:\n\n")
	preamble.WriteString("```json\n")
	preamble.WriteString(jsonInput)
	preamble.WriteString("\n```")

	return preamble.String()
}
