package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"ayo/internal/agent"
	"ayo/internal/config"
	"ayo/internal/ui"
)

func newInitCmd(cfgPath *string) *cobra.Command {
	var (
		handle       string
		model        string
		description  string
		ignoreShared bool
		systemPath   string
		systemText   string
	)

	const defaultSystemMessage = "You are a helpful assistant"

	cmd := &cobra.Command{
		Use:   "init [@handle]",
		Short: "Create a new agent",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				if len(args) > 0 {
					handle = agent.NormalizeHandle(args[0])
				}

				providerModels := config.ConfiguredModels(cfg)
				modelIDs := make([]string, 0, len(providerModels))
				modelSet := make(map[string]struct{}, len(providerModels))
				for _, m := range providerModels {
					modelIDs = append(modelIDs, m.ID)
					modelSet[m.ID] = struct{}{}
				}

				if handle == "" || model == "" || (systemText == "" && systemPath == "") {
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
					systemText = res.System
				}

				if model == "" {
					model = cfg.DefaultModel
				}

				if handle == "" {
					return fmt.Errorf("handle is required")
				}

				if systemText == "" && systemPath != "" {
					data, err := os.ReadFile(systemPath)
					if err != nil {
						return err
					}
					systemText = string(data)
				}

				if systemText == "" {
					systemText = defaultSystemMessage
				}

				if len(modelSet) > 0 {
					if _, ok := modelSet[model]; !ok {
						return fmt.Errorf("model %s is not configured", model)
					}
				}

				agCfg := agent.Config{
					Model:                     model,
					IgnoreSharedSystemMessage: ignoreShared,
					SystemFile:                systemPath,
					Description:               description,
				}

				ag, err := agent.Save(cfg, handle, agCfg, systemText)
				if err != nil {
					return err
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Created agent %s at %s\n", ag.Handle, ag.Dir)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&handle, "handle", "", "agent handle (without @)")
	cmd.Flags().StringVar(&model, "model", "", "model id")
	cmd.Flags().StringVar(&description, "description", "", "agent description")
	cmd.Flags().BoolVar(&ignoreShared, "ignore-shared", false, "ignore shared system message")
	cmd.Flags().StringVar(&systemPath, "system-file", "", "path to system message file")
	cmd.Flags().StringVar(&systemText, "system", "", "system message text (overrides file)")

	return cmd
}
