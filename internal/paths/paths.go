// Package paths provides directory paths for ayo.
package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	devRoot     string
	devRootOnce sync.Once
)

// IsDevMode returns true if ayo is running from a source checkout.
// In dev mode, built-in data is stored in {repo}/.ayo/ instead of ~/.local/share/ayo/.
func IsDevMode() bool {
	return getDevRoot() != ""
}

// DevRoot returns the repository root if running in dev mode, or empty string otherwise.
func DevRoot() string {
	return getDevRoot()
}

// getDevRoot finds the repository root by checking:
// 1. Walking up from executable location (for built binaries in repo)
// 2. Walking up from current working directory (for go run)
// looking for a go.mod file with "module ayo".
func getDevRoot() string {
	devRootOnce.Do(func() {
		// Try from executable first (handles ./ayo built binary)
		if root := findDevRootFrom(executableDir()); root != "" {
			devRoot = root
			return
		}

		// Try from current working directory (handles go run)
		if wd, err := os.Getwd(); err == nil {
			if root := findDevRootFrom(wd); root != "" {
				devRoot = root
				return
			}
		}
	})
	return devRoot
}

// executableDir returns the directory containing the executable, or empty if unknown.
func executableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return ""
	}
	return filepath.Dir(exe)
}

// findDevRootFrom walks up from the given directory looking for a go.mod with "module ayo".
func findDevRootFrom(startDir string) string {
	if startDir == "" {
		return ""
	}

	dir := startDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(goModPath); err == nil {
			// Check if this is the ayo module
			content := string(data)
			if strings.HasPrefix(content, "module ayo") ||
				strings.Contains(content, "\nmodule ayo") {
				return dir
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached filesystem root
		}
		dir = parent
	}
	return ""
}

// DataDir returns the data directory for ayo.
//
// Dev mode: {repo}/.ayo (project-local built-ins)
// Production Unix: ~/.local/share/ayo (XDG compliant)
// Production Windows: %LOCALAPPDATA%\ayo
//
// This directory stores built-in agents, built-in skills, and version markers.
// In dev mode, each checkout has its own isolated built-ins.
func DataDir() string {
	if root := getDevRoot(); root != "" {
		return filepath.Join(root, ".ayo")
	}

	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, _ := os.UserHomeDir()
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "ayo")
	}
	// Unix (macOS, Linux, etc.) - XDG compliant
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "ayo")
}

// ConfigDir returns the config directory for ayo.
//
// Unix (macOS, Linux): ~/.config/ayo
// Windows: %LOCALAPPDATA%\ayo (same as production DataDir)
//
// This directory stores user configuration and user-created content:
// config.yaml, user agents, user skills, and system prompts.
// This is always the global user directory, even in dev mode.
func ConfigDir() string {
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, _ := os.UserHomeDir()
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "ayo")
	}
	// Unix (macOS, Linux, etc.)
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ayo")
}

// AgentsDir returns the directory for user-created agents.
// Location: ~/.config/ayo/agents (Unix) or %LOCALAPPDATA%\ayo\agents (Windows)
// This is always the global user directory, even in dev mode.
func AgentsDir() string {
	return filepath.Join(ConfigDir(), "agents")
}

// BuiltinAgentsDir returns the directory for installed built-in agents.
// Dev mode: {repo}/.ayo/agents
// Production: ~/.local/share/ayo/agents (Unix) or %LOCALAPPDATA%\ayo\agents (Windows)
func BuiltinAgentsDir() string {
	return filepath.Join(DataDir(), "agents")
}

// SkillsDir returns the directory for user shared skills.
// Location: ~/.config/ayo/skills (Unix) or %LOCALAPPDATA%\ayo\skills (Windows)
// This is always the global user directory, even in dev mode.
func SkillsDir() string {
	return filepath.Join(ConfigDir(), "skills")
}

// BuiltinSkillsDir returns the directory for installed built-in skills.
// Dev mode: {repo}/.ayo/skills
// Production: ~/.local/share/ayo/skills (Unix) or %LOCALAPPDATA%\ayo\skills (Windows)
func BuiltinSkillsDir() string {
	return filepath.Join(DataDir(), "skills")
}

// ConfigFile returns the path to the main config file.
// Location: ~/.config/ayo/config.yaml (Unix) or %LOCALAPPDATA%\ayo\config.yaml (Windows)
// This is always the global user config, even in dev mode.
func ConfigFile() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// SystemPromptsDir returns the directory for system prompt files.
// Location: ~/.config/ayo/prompts (Unix) or %LOCALAPPDATA%\ayo\prompts (Windows)
// This is always the global user directory, even in dev mode.
func SystemPromptsDir() string {
	return filepath.Join(ConfigDir(), "prompts")
}

// VersionFile returns the path to the builtin version marker.
// Dev mode: {repo}/.ayo/.builtin-version
// Production: ~/.local/share/ayo/.builtin-version (Unix) or %LOCALAPPDATA%\ayo\.builtin-version (Windows)
func VersionFile() string {
	return filepath.Join(DataDir(), ".builtin-version")
}
