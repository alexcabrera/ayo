// Package builtin provides embedded built-in agents that ship with ayo.
package builtin

import (
	"embed"
	"encoding/json"
	"io/fs"
	"path"
	"sort"
	"strings"
)

//go:embed agents/*
var agentsFS embed.FS

// AgentDefinition represents a built-in agent's configuration and content
type AgentDefinition struct {
	Handle      string
	Config      AgentConfig
	System      string
	Skills      []SkillDefinition
	Description string
}

// AgentConfig mirrors the user agent config structure
type AgentConfig struct {
	Model                     string   `json:"model,omitempty"`
	IgnoreSharedSystemMessage bool     `json:"ignore_shared_system_message,omitempty"`
	Description               string   `json:"description,omitempty"`
	DelegateHint              string   `json:"delegate_hint,omitempty"`
	AllowedTools              []string `json:"allowed_tools,omitempty"`
}

// SkillDefinition represents a built-in skill
type SkillDefinition struct {
	Name        string
	Description string
	Content     string
}

// ListAgents returns all built-in agent handles
func ListAgents() []string {
	entries, err := agentsFS.ReadDir("agents")
	if err != nil {
		return nil
	}

	var handles []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Directory names include @ prefix (e.g., @ayo)
			handles = append(handles, entry.Name())
		}
	}
	sort.Strings(handles)
	return handles
}

// HasAgent checks if a built-in agent exists with the given handle
func HasAgent(handle string) bool {
	// Normalize to include @ prefix to match directory name
	if !strings.HasPrefix(handle, "@") {
		handle = "@" + handle
	}
	_, err := agentsFS.ReadDir(path.Join("agents", handle))
	return err == nil
}

// LoadAgent loads a built-in agent definition
func LoadAgent(handle string) (AgentDefinition, error) {
	// Normalize to include @ prefix to match directory name
	if !strings.HasPrefix(handle, "@") {
		handle = "@" + handle
	}
	basePath := path.Join("agents", handle)

	def := AgentDefinition{
		Handle: handle,
	}

	// Load config.json
	configData, err := agentsFS.ReadFile(path.Join(basePath, "config.json"))
	if err == nil {
		if err := json.Unmarshal(configData, &def.Config); err != nil {
			return def, err
		}
		def.Description = def.Config.Description
	}

	// Load system.md
	systemData, err := agentsFS.ReadFile(path.Join(basePath, "system.md"))
	if err != nil {
		return def, err
	}
	def.System = strings.TrimSpace(string(systemData))

	// Load skills if present
	skillsPath := path.Join(basePath, "skills")
	skillEntries, err := agentsFS.ReadDir(skillsPath)
	if err == nil {
		for _, entry := range skillEntries {
			if entry.IsDir() {
				skill, err := loadSkill(path.Join(skillsPath, entry.Name()))
				if err == nil {
					def.Skills = append(def.Skills, skill)
				}
			}
		}
	}

	return def, nil
}

func loadSkill(skillPath string) (SkillDefinition, error) {
	var skill SkillDefinition

	// Read SKILL.md for metadata
	skillMD, err := agentsFS.ReadFile(path.Join(skillPath, "SKILL.md"))
	if err != nil {
		return skill, err
	}

	content := string(skillMD)

	// Parse YAML frontmatter if present
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			// Parse frontmatter
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

	// Default name from directory
	if skill.Name == "" {
		skill.Name = path.Base(skillPath)
	}

	return skill, nil
}

// FS returns the embedded filesystem for built-in agents
func FS() fs.FS {
	sub, _ := fs.Sub(agentsFS, "agents")
	return sub
}

// AgentInfo contains summary information about a builtin agent for delegation hints
type AgentInfo struct {
	Handle       string
	Description  string
	DelegateHint string
}

// ListAgentInfo returns info about all builtin agents for use in system prompts
func ListAgentInfo() []AgentInfo {
	handles := ListAgents()
	infos := make([]AgentInfo, 0, len(handles))
	for _, handle := range handles {
		def, err := LoadAgent(handle)
		if err != nil {
			continue
		}
		infos = append(infos, AgentInfo{
			Handle:       handle,
			Description:  def.Config.Description,
			DelegateHint: def.Config.DelegateHint,
		})
	}
	return infos
}
