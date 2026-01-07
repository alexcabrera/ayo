package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSkillValidatesNameAndDescription(t *testing.T) {
	content := `---
name: pdf-processing
description: Extract pdfs
---
Body`
	meta, body, err := parseSkill("/tmp/skills/pdf-processing/SKILL.md", "pdf-processing", content)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if meta.Name != "pdf-processing" || meta.Description != "Extract pdfs" {
		t.Fatalf("metadata mismatch: %+v", meta)
	}
	if body != "Body" {
		t.Fatalf("body mismatch: %q", body)
	}
}

func TestParseSkillRejectsBadName(t *testing.T) {
	content := `---
name: Pdf
description: desc
---
Body`
	_, _, err := parseSkill("/tmp/skills/Pdf/SKILL.md", "Pdf", content)
	if err == nil {
		t.Fatalf("expected error for bad name")
	}
}

func TestParseSkillRequiresFrontmatter(t *testing.T) {
	_, _, err := parseSkill("/tmp/skills/skill/SKILL.md", "skill", "no frontmatter")
	if err == nil {
		t.Fatalf("expected frontmatter error")
	}
}

func TestDiscoverPrefersFirstAndWarnsDuplicates(t *testing.T) {
	root := t.TempDir()
	agentRoot := filepath.Join(root, "agent")
	sharedRoot := filepath.Join(root, "shared")
	mustWriteSkill(t, filepath.Join(agentRoot, "skill-a"), "skill-a", "desc a")
	mustWriteSkill(t, filepath.Join(sharedRoot, "skill-a"), "skill-a", "desc b")

	res := Discover(agentRoot, sharedRoot)
	if len(res.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(res.Skills))
	}
	if res.Skills[0].Description != "desc a" {
		t.Fatalf("expected agent skill to win")
	}
	if len(res.Warnings) == 0 {
		t.Fatalf("expected duplicate warning")
	}
}

func TestParseSkillOptionalFields(t *testing.T) {
	content := `---
name: my-skill
description: A test skill
license: MIT
compatibility: Requires bash
allowed-tools: Bash(git:*) Read
metadata:
  author: test-org
  version: "1.0"
---
Body content here`
	meta, body, err := parseSkill("/tmp/skills/my-skill/SKILL.md", "my-skill", content)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if meta.Name != "my-skill" {
		t.Errorf("name mismatch: %s", meta.Name)
	}
	if meta.Description != "A test skill" {
		t.Errorf("description mismatch: %s", meta.Description)
	}
	if meta.License != "MIT" {
		t.Errorf("license mismatch: %s", meta.License)
	}
	if meta.Compatibility != "Requires bash" {
		t.Errorf("compatibility mismatch: %s", meta.Compatibility)
	}
	if meta.AllowedTools != "Bash(git:*) Read" {
		t.Errorf("allowed-tools mismatch: %s", meta.AllowedTools)
	}
	if meta.RawMetadata == nil {
		t.Fatal("expected metadata map")
	}
	if meta.RawMetadata["author"] != "test-org" {
		t.Errorf("author mismatch: %s", meta.RawMetadata["author"])
	}
	if meta.RawMetadata["version"] != "1.0" {
		t.Errorf("version mismatch: %s", meta.RawMetadata["version"])
	}
	if meta.Version() != "1.0" {
		t.Errorf("Version() mismatch: %s", meta.Version())
	}
	if meta.Author() != "test-org" {
		t.Errorf("Author() mismatch: %s", meta.Author())
	}
	if body != "Body content here" {
		t.Errorf("body mismatch: %q", body)
	}
}

func TestSkillSourceString(t *testing.T) {
	tests := []struct {
		source SkillSource
		want   string
	}{
		{SourceAgentSpecific, "agent"},
		{SourceUserShared, "user"},
		{SourceBuiltIn, "built-in"},
		{SkillSource(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.source.String(); got != tt.want {
			t.Errorf("SkillSource(%d).String() = %q, want %q", tt.source, got, tt.want)
		}
	}
}

func TestDiscoverWithSources(t *testing.T) {
	root := t.TempDir()
	agentDir := filepath.Join(root, "agent")
	userDir := filepath.Join(root, "user")
	builtinDir := filepath.Join(root, "builtin")

	// Create skills in each source
	mustWriteSkill(t, filepath.Join(agentDir, "skill-a"), "skill-a", "agent version")
	mustWriteSkill(t, filepath.Join(userDir, "skill-a"), "skill-a", "user version")
	mustWriteSkill(t, filepath.Join(userDir, "skill-b"), "skill-b", "user only")
	mustWriteSkill(t, filepath.Join(builtinDir, "skill-c"), "skill-c", "builtin only")

	res := DiscoverWithSources([]SkillSourceDir{
		{Path: agentDir, Source: SourceAgentSpecific, Label: "agent"},
		{Path: userDir, Source: SourceUserShared, Label: "user"},
		{Path: builtinDir, Source: SourceBuiltIn, Label: "builtin"},
	})

	if len(res.Skills) != 3 {
		t.Fatalf("expected 3 skills, got %d", len(res.Skills))
	}

	// Check that agent version of skill-a won
	for _, s := range res.Skills {
		if s.Name == "skill-a" && s.Description != "agent version" {
			t.Errorf("skill-a should be agent version, got %s", s.Description)
		}
		if s.Name == "skill-a" && s.Source != SourceAgentSpecific {
			t.Errorf("skill-a should have agent source, got %s", s.Source)
		}
	}

	// Should have warning about duplicate
	foundWarning := false
	for _, w := range res.Warnings {
		if contains(w, "duplicate") && contains(w, "skill-a") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Error("expected duplicate warning for skill-a")
	}
}

func TestDiscoverDetectsOptionalDirs(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "my-skill")
	mustWriteSkill(t, skillDir, "my-skill", "test skill")

	// Create optional directories
	if err := os.MkdirAll(filepath.Join(skillDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(skillDir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Don't create assets

	res := DiscoverWithSources([]SkillSourceDir{
		{Path: root, Source: SourceUserShared, Label: "user"},
	})

	if len(res.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(res.Skills))
	}

	skill := res.Skills[0]
	if !skill.HasScripts {
		t.Error("expected HasScripts to be true")
	}
	if !skill.HasRefs {
		t.Error("expected HasRefs to be true")
	}
	if skill.HasAssets {
		t.Error("expected HasAssets to be false")
	}
}

func TestLoadPreservesSource(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "my-skill")
	mustWriteSkillFull(t, skillDir, "my-skill", "test skill", "MIT", "test-author", "2.0")

	res := DiscoverWithSources([]SkillSourceDir{
		{Path: root, Source: SourceBuiltIn, Label: "builtin"},
	})

	if len(res.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(res.Skills))
	}

	skill, err := Load(res.Skills[0])
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if skill.Metadata.Source != SourceBuiltIn {
		t.Errorf("expected source BuiltIn, got %s", skill.Metadata.Source)
	}
	if skill.Metadata.License != "MIT" {
		t.Errorf("expected license MIT, got %s", skill.Metadata.License)
	}
	if skill.Metadata.Version() != "2.0" {
		t.Errorf("expected version 2.0, got %s", skill.Metadata.Version())
	}
	if skill.Metadata.Author() != "test-author" {
		t.Errorf("expected author test-author, got %s", skill.Metadata.Author())
	}
}

func TestMetadataVersionAuthorEmpty(t *testing.T) {
	meta := Metadata{Name: "test", Description: "test"}
	if meta.Version() != "" {
		t.Error("expected empty version")
	}
	if meta.Author() != "" {
		t.Error("expected empty author")
	}
}

func mustWriteSkill(t *testing.T, dir, name, desc string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := "---\nname: " + name + "\ndescription: " + desc + "\n---\nbody"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func mustWriteSkillFull(t *testing.T, dir, name, desc, license, author, version string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := `---
name: ` + name + `
description: ` + desc + `
license: ` + license + `
metadata:
  author: ` + author + `
  version: "` + version + `"
---
body content`
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
