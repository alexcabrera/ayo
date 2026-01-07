package config

import "testing"

func TestConfiguredModelsRequiresAPIKey(t *testing.T) {
	cfg := Default()
	cfg.Provider.Models = nil
	// clear potential env
	t.Setenv("OPENAI_API_KEY", "")

	models := ConfiguredModels(cfg)
	if len(models) != 0 {
		t.Fatalf("expected no models without API key")
	}
}

func TestConfiguredModelsUsesEmbeddedWhenAPIKeyPresent(t *testing.T) {
	cfg := Default()
	cfg.Provider.Models = nil
	t.Setenv("OPENAI_API_KEY", "key")

	models := ConfiguredModels(cfg)
	if len(models) == 0 {
		t.Fatalf("expected embedded models when API key present")
	}
}
