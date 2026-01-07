package builtin

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

//go:embed skills/*
var skillsFS embed.FS

// SkillInfo contains summary information about a builtin skill.
type SkillInfo struct {
	Name        string
	Description string
	Path        string // Relative path in embedded FS
}

// ListBuiltinSkills returns all shared built-in skill names.
func ListBuiltinSkills() []string {
	entries, err := skillsFS.ReadDir("skills")
	if err != nil {
		return nil
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names
}

// HasBuiltinSkill checks if a shared built-in skill exists.
func HasBuiltinSkill(name string) bool {
	skillPath := path.Join("skills", name, "SKILL.md")
	_, err := skillsFS.ReadFile(skillPath)
	return err == nil
}

// LoadBuiltinSkill loads a shared built-in skill definition.
func LoadBuiltinSkill(name string) (SkillDefinition, error) {
	skillPath := path.Join("skills", name)
	return loadSkillFromFS(skillsFS, skillPath, name)
}

// ListAgentBuiltinSkills returns skill names for a specific built-in agent.
func ListAgentBuiltinSkills(handle string) []string {
	if !strings.HasPrefix(handle, "@") {
		handle = "@" + handle
	}
	skillsPath := path.Join("agents", handle, "skills")
	entries, err := agentsFS.ReadDir(skillsPath)
	if err != nil {
		return nil
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names
}

// LoadAgentBuiltinSkill loads an agent-specific built-in skill.
func LoadAgentBuiltinSkill(handle, skillName string) (SkillDefinition, error) {
	if !strings.HasPrefix(handle, "@") {
		handle = "@" + handle
	}
	skillPath := path.Join("agents", handle, "skills", skillName)
	return loadSkillFromFS(agentsFS, skillPath, skillName)
}

// loadSkillFromFS loads a skill definition from an embedded filesystem.
func loadSkillFromFS(fsys embed.FS, skillPath, skillName string) (SkillDefinition, error) {
	var skill SkillDefinition

	skillMDPath := path.Join(skillPath, "SKILL.md")
	data, err := fsys.ReadFile(skillMDPath)
	if err != nil {
		return skill, fmt.Errorf("skill %s not found: %w", skillName, err)
	}

	content := string(data)

	// Parse YAML frontmatter
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			frontmatter := parts[1]
			for _, line := range strings.Split(frontmatter, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "name:") {
					skill.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				} else if strings.HasPrefix(line, "description:") {
					skill.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				}
			}
			skill.Content = strings.TrimSpace(parts[2])
		}
	} else {
		skill.Content = content
	}

	if skill.Name == "" {
		skill.Name = skillName
	}

	return skill, nil
}

// SkillsFS returns the embedded filesystem for shared built-in skills.
func SkillsFS() fs.FS {
	sub, _ := fs.Sub(skillsFS, "skills")
	return sub
}

// GetAllBuiltinSkillInfos returns info about all built-in skills (shared + agent-specific).
func GetAllBuiltinSkillInfos() []SkillInfo {
	var infos []SkillInfo

	// Shared built-in skills
	for _, name := range ListBuiltinSkills() {
		skill, err := LoadBuiltinSkill(name)
		if err != nil {
			continue
		}
		infos = append(infos, SkillInfo{
			Name:        skill.Name,
			Description: skill.Description,
			Path:        path.Join("skills", name),
		})
	}

	return infos
}
