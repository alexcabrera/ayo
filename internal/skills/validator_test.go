package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateValidSkill(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "my-skill")
	mustWriteSkill(t, skillDir, "my-skill", "A valid skill description")

	errors := Validate(skillDir)
	if len(errors) > 0 {
		t.Errorf("expected no errors, got: %v", errors)
	}
}

func TestValidateMissingDirectory(t *testing.T) {
	errors := Validate("/nonexistent/path")
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if !containsHelper(errors[0].Message, "does not exist") {
		t.Errorf("expected 'does not exist' error, got: %s", errors[0].Message)
	}
}

func TestValidateMissingSkillMD(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "empty-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}

	errors := Validate(skillDir)
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if !containsHelper(errors[0].Message, "SKILL.md") {
		t.Errorf("expected SKILL.md error, got: %s", errors[0].Message)
	}
}

func TestValidateMissingName(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "bad-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
description: A skill without a name
---
Body`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	errors := Validate(skillDir)
	if len(errors) == 0 {
		t.Fatal("expected validation errors")
	}
	found := false
	for _, e := range errors {
		if e.Field == "name" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected name field error")
	}
}

func TestValidateUppercaseName(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "BadName")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: BadName
description: A skill with uppercase name
---
Body`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	errors := Validate(skillDir)
	if len(errors) == 0 {
		t.Fatal("expected validation errors for uppercase name")
	}
}

func TestValidateNameMismatch(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "dir-name")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: different-name
description: Name doesn't match directory
---
Body`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	errors := Validate(skillDir)
	found := false
	for _, e := range errors {
		if e.Field == "name" && containsHelper(e.Message, "match directory") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected directory match error")
	}
}

func TestValidateConsecutiveHyphens(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "bad--name")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: bad--name
description: Name with consecutive hyphens
---
Body`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	errors := Validate(skillDir)
	found := false
	for _, e := range errors {
		if e.Field == "name" && containsHelper(e.Message, "consecutive") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected consecutive hyphens error")
	}
}

func TestValidateLeadingHyphen(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "-bad")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: -bad
description: Name with leading hyphen
---
Body`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	errors := Validate(skillDir)
	found := false
	for _, e := range errors {
		if e.Field == "name" && containsHelper(e.Message, "start or end") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected leading/trailing hyphen error")
	}
}

func TestValidateDescriptionTooLong(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "long-desc")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	
	longDesc := make([]byte, 1100)
	for i := range longDesc {
		longDesc[i] = 'a'
	}
	
	content := "---\nname: long-desc\ndescription: " + string(longDesc) + "\n---\nBody"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	errors := Validate(skillDir)
	found := false
	for _, e := range errors {
		if e.Field == "description" && containsHelper(e.Message, "1024") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected description length error")
	}
}

func TestValidateCompatibilityTooLong(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "long-compat")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	
	longCompat := make([]byte, 600)
	for i := range longCompat {
		longCompat[i] = 'a'
	}
	
	content := "---\nname: long-compat\ndescription: Valid desc\ncompatibility: " + string(longCompat) + "\n---\nBody"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	errors := Validate(skillDir)
	found := false
	for _, e := range errors {
		if e.Field == "compatibility" && containsHelper(e.Message, "500") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected compatibility length error")
	}
}

func TestValidateUnexpectedField(t *testing.T) {
	root := t.TempDir()
	skillDir := filepath.Join(root, "extra-field")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: extra-field
description: Valid desc
unexpected_field: should cause error
---
Body`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	errors := Validate(skillDir)
	found := false
	for _, e := range errors {
		if e.Field == "unexpected_field" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected unexpected field error")
	}
}

func TestValidationErrorString(t *testing.T) {
	err := ValidationError{Field: "name", Message: "is required"}
	if err.Error() != "name: is required" {
		t.Errorf("unexpected error string: %s", err.Error())
	}

	err2 := ValidationError{Message: "general error"}
	if err2.Error() != "general error" {
		t.Errorf("unexpected error string: %s", err2.Error())
	}
}
