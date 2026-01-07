package config

import (
	"os"
	"strings"

	"github.com/charmbracelet/catwalk/pkg/embedded"
)

type ModelChoice struct {
	ID   string
	Name string
}

func ConfiguredModels(cfg Config) []ModelChoice {
	provider := cfg.Provider
	models := provider.Models

	if len(models) == 0 {
		for _, p := range embedded.GetAll() {
			if p.ID == provider.ID {
				models = p.Models
				break
			}
		}
	}

	envKey := strings.ToUpper(string(provider.ID)) + "_API_KEY"
	if envKey != "" && os.Getenv(envKey) == "" {
		return nil
	}

	choices := make([]ModelChoice, 0, len(models))
	for _, m := range models {
		choices = append(choices, ModelChoice{ID: m.ID, Name: m.Name})
	}
	return choices
}
