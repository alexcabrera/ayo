// Package paths provides directory paths for ayo.
//
// Directory Priority Order (first found wins for lookups):
//  1. ./.config/ayo (local project config)
//  2. ./.local/share/ayo (local project data)
//  3. ~/.config/ayo (user config)
//  4. ~/.local/share/ayo (user data / built-ins)
//
// For writes, ayo uses:
//   - User agents/skills: ~/.config/ayo (or ./.config/ayo with --dev)
//   - Built-in installation: ~/.local/share/ayo (or ./.local/share/ayo with --dev)
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

	// localDevMode is set by SetLocalDevMode to force local directory paths
	localDevMode     bool
	localDevModeOnce sync.Once
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

// SetLocalDevMode enables local dev mode, which uses ./.local/share/ayo and ./.config/ayo
// instead of the default directories. This is typically set via the --dev flag on setup.
// Must be called before any directory functions are used.
func SetLocalDevMode() {
	localDevModeOnce.Do(func() {
		localDevMode = true
	})
}

// IsLocalDevMode returns true if local dev mode is enabled via SetLocalDevMode.
func IsLocalDevMode() bool {
	return localDevMode
}

// DataDir returns the data directory for ayo.
//
// Local dev mode: ./.local/share/ayo (current directory)
// Dev mode: {repo}/.ayo (project-local built-ins)
// Production Unix: ~/.local/share/ayo (XDG compliant)
// Production Windows: %LOCALAPPDATA%\ayo
//
// This directory stores built-in agents, built-in skills, and version markers.
// In dev mode, each checkout has its own isolated built-ins.
func DataDir() string {
	if localDevMode {
		wd, _ := os.Getwd()
		return filepath.Join(wd, ".local", "share", "ayo")
	}

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
// Local dev mode: ./.config/ayo (current directory)
// Unix (macOS, Linux): ~/.config/ayo
// Windows: %LOCALAPPDATA%\ayo (same as production DataDir)
//
// This directory stores user configuration and user-created content:
// config.yaml, user agents, user skills, and system prompts.
// This is always the global user directory, even in dev mode (unless local dev mode).
func ConfigDir() string {
	if localDevMode {
		wd, _ := os.Getwd()
		return filepath.Join(wd, ".config", "ayo")
	}

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

// LocalConfigDir returns the local project config directory (./.config/ayo).
// Returns empty string if not in a directory context or on Windows.
func LocalConfigDir() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(wd, ".config", "ayo")
}

// LocalDataDir returns the local project data directory (./.local/share/ayo).
// Returns empty string if not in a directory context or on Windows.
func LocalDataDir() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(wd, ".local", "share", "ayo")
}

// UserConfigDir returns the global user config directory (~/.config/ayo).
// On Windows, returns %LOCALAPPDATA%\ayo.
func UserConfigDir() string {
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, _ := os.UserHomeDir()
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "ayo")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ayo")
}

// UserDataDir returns the global user data directory (~/.local/share/ayo).
// On Windows, returns %LOCALAPPDATA%\ayo.
func UserDataDir() string {
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, _ := os.UserHomeDir()
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "ayo")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "ayo")
}

// HasLocalConfig returns true if a local config directory exists (./.config/ayo).
func HasLocalConfig() bool {
	dir := LocalConfigDir()
	if dir == "" {
		return false
	}
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// HasLocalData returns true if a local data directory exists (./.local/share/ayo).
func HasLocalData() bool {
	dir := LocalDataDir()
	if dir == "" {
		return false
	}
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// AgentsDirs returns all agent directories in lookup priority order.
// Order: local config, local data, user config, user data (built-in).
// Only includes directories that exist.
func AgentsDirs() []string {
	var dirs []string
	check := func(base string) {
		if base == "" {
			return
		}
		dir := filepath.Join(base, "agents")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			dirs = append(dirs, dir)
		}
	}

	check(LocalConfigDir())
	check(LocalDataDir())
	check(UserConfigDir())
	check(UserDataDir())

	return dirs
}

// SkillsDirs returns all skills directories in lookup priority order.
// Order: local config, local data, user config, user data (built-in).
// Only includes directories that exist.
func SkillsDirs() []string {
	var dirs []string
	check := func(base string) {
		if base == "" {
			return
		}
		dir := filepath.Join(base, "skills")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			dirs = append(dirs, dir)
		}
	}

	check(LocalConfigDir())
	check(LocalDataDir())
	check(UserConfigDir())
	check(UserDataDir())

	return dirs
}
