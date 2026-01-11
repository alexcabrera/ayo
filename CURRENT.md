# Current Feature: Agent Creation Wizard & System Prompt Refactoring

## Overview

This document summarizes the recent changes to the ayo CLI, focusing on the agent creation wizard improvements and system prompt architecture refactoring.

---

## Agent Creation Wizard (`ayo agents create`)

### Multi-Step Wizard Flow

The agent creation wizard now uses a clean, step-by-step flow with each step on its own dedicated screen (using `tea.WithAltScreen()`):

1. **Step 1 - Identity**: Handle, description, model selection
2. **Step 2 - Tools**: Multi-select for allowed tools (bash, agent_call)
3. **Step 3 - Skills**: Tabbed picker for built-in vs user-defined skills
4. **Step 4 - System Prompt**: Choose source (inline or file), then enter content
5. **Step 5 - Agent Chaining**: Enable structured I/O with JSON schemas
6. **Review**: Summary of all selections with confirm/cancel

### Skills Picker Component

New standalone component (`internal/ui/skills_picker.go`) with:
- Tabbed interface: "Built-in" vs "User-defined" 
- Keyboard navigation: `tab`/`←`/`→` switch tabs, `↑`/`↓` navigate, `space` toggle, `enter` confirm
- Proper source detection for skills (built-in vs user)

### File Picker Component

New file picker component (`internal/ui/file_picker.go`) styled after Crush:
- Used for system prompt file selection and JSON schema selection
- Filters by file extension (`.md`, `.txt` for prompts; `.json`, `.jsonschema` for schemas)
- Vim-style navigation (`h`/`j`/`k`/`l`) and arrow keys
- Current path display with `~` expansion

---

## System Prompt Architecture

### Previous Structure (Removed)

The old system had:
- `SharedSystemMessage` - A shared system.md included for all agents
- `SystemPrefix` / `SystemSuffix` - Prefix and suffix files
- `IgnoreSharedSystemMessage` flag per agent

### New Structure

Simplified to just prefix and suffix:

| Component | Description |
|-----------|-------------|
| **Environment Context** | Auto-generated: OS, datetime, working directory |
| **system-prefix.md** | Prepended to all agents |
| **Agent's system.md** | Agent-specific instructions |
| **system-suffix.md** | Appended to all agents |

### File Locations & Priority

Prompts are discovered with priority (first found wins):

1. `./.config/ayo/prompts/{name}` - Local project override
2. `~/.config/ayo/prompts/{name}` - User override
3. `./.local/share/ayo/prompts/{name}` - Local project built-in
4. `~/.local/share/ayo/prompts/{name}` - Global built-in

### Built-in Prompts

Embedded in binary at `internal/builtin/prompts/`:
- `system-prefix.md` - Default prefix for all agents
- `system-suffix.md` - Default suffix for all agents

Installed via `ayo setup` to `~/.local/share/ayo/prompts/` (or `.local/share/ayo/prompts/` in dev mode).

### Path Resolution Functions

New functions in `internal/paths/paths.go`:
- `BuiltinPromptsDir()` - Returns data dir prompts path
- `UserPromptsDir()` - Returns config dir prompts path  
- `FindPromptFile(name)` - Finds prompt file using priority order

---

## Agents & Skills List UI

Both `ayo agents list` and `ayo skills list` now use grouped sections:

```
  Agents / Skills
  ──────────────────────────────────────────────────────────

  User-defined
    (items or helpful empty state message)

  Built-in  
    (items or helpful empty state message)

  ──────────────────────────────────────────────────────────
  N agents / skills
```

---

## Input Validation Error Messages

Improved error messages for agents with input schemas:

```
  ERROR: Agent requires structured JSON input

  This agent requires structured JSON input.
  
  Your input is not valid JSON.
  
  Expected format:
    {
      "field_name": "...", (required)  // Field description
      ...
    }
```

Shows:
- Clear error header
- Explanation of what went wrong
- Expected JSON schema with field names, types, required markers, and descriptions

---

## Install Script Changes

`./install.sh` now:
- Accepts `--dev` flag to force local installation
- Runs `ayo setup` after building to install built-in agents, skills, and prompts
- Properly handles branch detection for install location

---

## Config Changes

### Removed Fields
- `SharedSystemMessage` - No longer used
- `IgnoreSharedSystemMessage` (agent config) - No longer used

### Updated Fields
- `SystemPrefix` - Now empty by default; uses `paths.FindPromptFile("system-prefix.md")`
- `SystemSuffix` - Now empty by default; uses `paths.FindPromptFile("system-suffix.md")`

---

## Files Changed

### New Files
- `internal/builtin/prompts/system-prefix.md`
- `internal/builtin/prompts/system-suffix.md`
- `internal/ui/skills_picker.go`
- `internal/ui/file_picker.go`
- `CURRENT.md` (this file)

### Modified Files
- `install.sh` - Added `--dev` flag and setup integration
- `cmd/ayo/agents.go` - Updated list format, removed shared system flag
- `cmd/ayo/skills.go` - Updated list format with grouped sections
- `cmd/ayo/main.go` - Custom error handler for input validation
- `cmd/ayo/root.go` - Input validation error printing
- `internal/agent/agent.go` - New prompt loading logic, improved error messages
- `internal/builtin/builtin.go` - Added prompts embed
- `internal/builtin/install.go` - Added prompts extraction
- `internal/config/config.go` - Removed SharedSystemMessage, updated defaults
- `internal/paths/paths.go` - Added prompt path functions
- `internal/ui/agent_create_form.go` - Complete rewrite of wizard flow

### Updated Tests
- `internal/agent/agent_test.go` - Updated for new prompt structure
- `internal/config/config_test.go` - Updated for removed fields

---

## Version

Built-in version bumped to trigger reinstallation of prompts on next `ayo setup`.
