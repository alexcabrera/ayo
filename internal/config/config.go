package config

import (
	"fmt"
	"os"
	"strings"

	"ayo/internal/paths"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration for ayo.
type Config struct {
	AgentsDir      string           `yaml:"agents_dir"`
	SystemPrefix   string           `yaml:"system_prefix"`
	SystemSuffix   string           `yaml:"system_suffix"`
	SkillsDir      string           `yaml:"skills_dir"`
	DefaultModel   string           `yaml:"default_model"`
	CatwalkBaseURL string           `yaml:"catwalk_base_url"`
	Provider       catwalk.Provider `yaml:"provider"`
}

func apiKeyEnvForProvider(p catwalk.Provider) string {
	if p.ID == "" {
		return ""
	}
	return strings.ToUpper(string(p.ID)) + "_API_KEY"
}

func defaultCatwalkURL() string {
	if env := strings.TrimSpace(os.Getenv("CATWALK_URL")); env != "" {
		return env
	}
	return "http://localhost:8080"
}

// Default returns a Config populated with default values.
func Default() Config {
	return Config{
		AgentsDir:      paths.AgentsDir(),
		SystemPrefix:   "", // Uses paths.FindPromptFile("system-prefix.md")
		SystemSuffix:   "", // Uses paths.FindPromptFile("system-suffix.md")
		SkillsDir:      paths.SkillsDir(),
		DefaultModel:   "gpt-4.1",
		CatwalkBaseURL: defaultCatwalkURL(),
		Provider: catwalk.Provider{
			Name:        "openai",
			ID:          catwalk.InferenceProviderOpenAI,
			APIEndpoint: "https://api.openai.com/v1",
		},
	}
}

// Load reads configuration from the given path, falling back to defaults when missing.
func Load(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	if strings.TrimSpace(cfg.CatwalkBaseURL) == "" {
		cfg.CatwalkBaseURL = defaultCatwalkURL()
	}

	return cfg, nil
}
