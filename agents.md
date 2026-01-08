# ayo CLI

**CRITICAL: After every code change, you MUST add/maintain tests and keep them passing. NEVER reply to the user while tests are failing. Never skip tests.**
**CRITICAL: Maintain TODO.md checklist at repo root. Add new tasks before work, check them off when done, and keep it up to date.**
**CRITICAL: Do not modify anything under ./tmp (vendored, read-only). Explore only.**
**CRITICAL: Do not modify anything under ./.ayo (local dev config, read-only). This is the project-local built-in data directory used during development.**

## Documentation Guidelines

**CRITICAL: All command examples in documentation (README.md, AGENTS.md, etc.) must work if copy/pasted.**
- Use real agent handles and skill names that exist (e.g., `@ayo`, `@ayo.example.chain.code-reviewer`, `debugging`)
- For commands that create new entities (like `ayo init @myagent`), placeholders are acceptable since they will create the entity
- Directory structure diagrams showing hypothetical user content are acceptable (e.g., `@myagent/` to show where user agents go)
- Never use placeholder names like `@agent`, `@myagent`, `@source-agent` in commands that query or operate on existing entities
- Always test example commands before committing documentation changes

## Preferred Libraries (./tmp)

The `./tmp` directory contains vendored source code from Charm and related libraries. These are **read-only reference implementations** for illustrative purposes only.

**IMPORTANT:**
- These sources are snapshots and may be outdated - always verify against live documentation when implementing
- Use these libraries as the **first-line solution** for any applicable problem
- Do NOT modify files in `./tmp` - explore only
- When in doubt, check the official docs at https://charm.sh/

### Library Reference

| Library | Import | Use For |
|---------|--------|---------|
| **Bubble Tea** | `github.com/charmbracelet/bubbletea` | Interactive TUI apps (Elm Architecture) |
| **Bubbles** | `github.com/charmbracelet/bubbles` | Pre-built TUI components (spinners, inputs, tables) |
| **Lip Gloss** | `github.com/charmbracelet/lipgloss` | Terminal styling (colors, borders, layout) |
| **Glamour** | `github.com/charmbracelet/glamour` | Markdown rendering in terminal |
| **Huh** | `github.com/charmbracelet/huh` | Interactive forms and prompts |
| **Log** | `github.com/charmbracelet/log` | Styled, leveled logging |
| **Harmonica** | `github.com/charmbracelet/harmonica` | Spring physics animations |
| **Fang** | `github.com/charmbracelet/fang` | Cobra CLI enhancements (help, manpages) |
| **Catwalk** | `github.com/charmbracelet/catwalk` | AI provider/model configuration |
| **Fantasy** | `charm.land/fantasy` | Provider-agnostic LLM abstraction (streaming, tools, agents) |

### When to Use Each Library

#### Bubble Tea (`./tmp/bubbletea`)
**Use for:** Full interactive TUI applications with state management
- Complex multi-screen interfaces
- Real-time updates and event handling
- Keyboard/mouse input handling
- Any app needing Model-View-Update pattern

#### Bubbles (`./tmp/bubbles`)
**Use for:** Drop-in UI components for Bubble Tea apps
- `spinner` - Loading indicators
- `textinput` - Single-line text input
- `textarea` - Multi-line text editing
- `table` - Data tables
- `list` - Scrollable lists with filtering
- `viewport` - Scrollable content panes
- `paginator` - Paged content navigation
- `filepicker` - File/directory selection
- `timer` - Countdown/stopwatch

#### Lip Gloss (`./tmp/lipgloss`)
**Use for:** Styling terminal output
- Colors (foreground, background, adaptive)
- Text formatting (bold, italic, underline)
- Borders and padding
- Layout (centering, joining text blocks)
- `lipgloss/table` - Styled tables
- `lipgloss/list` - Styled lists
- `lipgloss/tree` - Tree rendering

#### Glamour (`./tmp/glamour`)
**Use for:** Rendering Markdown in terminal
- Displaying README/documentation
- Help text with formatting
- Any Markdown content (API responses, notes)
- Multiple built-in themes (dark, light, dracula, tokyo-night)

#### Huh (`./tmp/huh`)
**Use for:** Interactive forms and user input collection
- Multi-field forms with validation
- Selection menus (single/multi-select)
- Confirmation prompts
- Text input with placeholders
- File pickers
- Works standalone or embedded in Bubble Tea

#### Log (`./tmp/log`)
**Use for:** Application logging
- Leveled logging (debug, info, warn, error, fatal)
- Colored, styled output
- Structured logging with key-value pairs
- JSON/logfmt formatters
- slog compatibility

#### Harmonica (`./tmp/harmonica`)
**Use for:** Smooth animations
- Spring physics for natural motion
- Animated progress bars
- Smooth scrolling
- UI element transitions

#### Fang (`./tmp/fang`)
**Use for:** Enhancing Cobra CLI apps
- Styled help output
- Automatic `--version` flag
- Manpage generation
- Shell completion setup
- Consistent error handling

### Additional Tools

#### Gum (`./tmp/gum`)
**Use for:** Shell script interactivity (not Go code)
- Quick prompts from bash scripts
- `gum input`, `gum choose`, `gum confirm`, `gum spin`
- Prototyping before building in Go

#### Glow (`./tmp/glow`)
**Use for:** Reference for markdown browsing TUI
- Terminal markdown viewer implementation
- GitHub/GitLab markdown fetching patterns

#### Sequin (`./tmp/sequin`)
**Use for:** Debugging terminal output
- Decoding ANSI escape sequences
- Inspecting TUI rendering
- Validating golden test files

#### Crush (`./tmp/crush`)
**Use for:** Reference AI CLI implementation
- LLM integration patterns
- MCP server support
- LSP context integration
- Session management

## Completion Checklist

**Before reporting any task as complete, you MUST:**

1. **Run the full test suite**: `go test ./...`
   - If any test fails, fix it immediately without asking the user
   - Keep iterating until all tests pass
   - Never report completion while tests are failing

2. **Rebuild the binary**: `go install ./cmd/ayo`
   - This ensures the `ayo` command reflects all changes

A task is NOT complete until both steps pass successfully.

A Go-based command line tool for managing local AI agents.

## Features

- Define, manage, and run AI agents
- Built-in agents shipped with the binary
- Interactive chat sessions within the terminal
- Non-interactive single-prompt mode
- Bash tool as default for task execution
- System prompts assembled from prefix, shared, agent, tools, skills, and suffix
- Configurable models via Catwalk

## Configuration

Ayo uses two directories:

**Unix (macOS, Linux):**
- User config: `~/.config/ayo/` (config.yaml, user agents, user skills, prompts)
- Built-in data: `~/.local/share/ayo/` (agents and skills auto-installed on first run)

**Dev mode:** When running from a source checkout (`go run ./cmd/ayo`), built-in data is stored in `{repo}/.ayo/` instead. User config remains at `~/.config/ayo/`.

**Windows:**
- Both: `%LOCALAPPDATA%\ayo\`

```yaml
# ~/.config/ayo/config.yaml
agents_dir: ~/.config/ayo/agents
skills_dir: ~/.config/ayo/skills
shared_system_message: ~/.config/ayo/prompts/system.md
system_prefix: ~/.config/ayo/prompts/prefix.md
system_suffix: ~/.config/ayo/prompts/suffix.md
default_model: gpt-4.1
provider: {}
```

## Directory Structure

**Production (installed binary):**
```
~/.config/ayo/                    # User configuration (editable)
├── config.yaml                   # Main config file
├── agents/                       # User-defined agents
│   └── @myagent/
│       ├── config.json
│       ├── system.md
│       └── skills/               # Agent-specific skills
├── skills/                       # User-defined shared skills
│   └── my-skill/
│       └── SKILL.md
└── prompts/                      # Custom system prompts
    ├── prefix.md
    ├── system.md
    └── suffix.md

~/.local/share/ayo/               # Built-in data (auto-installed on first run)
├── agents/                       # Built-in agents
│   └── @ayo/
│       ├── config.json
│       ├── system.md
│       └── skills/
├── skills/                       # Built-in shared skills
│   └── debugging/
│       └── SKILL.md
└── .builtin-version              # Version marker
```

**Dev mode (running from source checkout):**
```
~/Code/ayo-skills/                # Your checkout
├── .ayo/                         # Built-in data (project-local)
│   ├── agents/
│   ├── skills/
│   └── .builtin-version
└── ...

~/.config/ayo/                    # User config (shared across all instances)
├── agents/
├── skills/
└── ...
```

This allows multiple dev branches to have isolated built-ins while sharing user-defined agents and skills.

## Loading Priority

**Agents:** User agents (`~/.config/ayo/agents`) take priority over built-in agents (`~/.local/share/ayo/agents`).

**Skills:** Discovery priority (first found wins):
1. Agent-specific skills (in agent's `skills/` directory)
2. User shared skills (`~/.config/ayo/skills`)
3. Built-in skills (`~/.local/share/ayo/skills`)

## Usage

```bash
# Setup (optional - built-ins auto-install on first run)
ayo setup                   # Reinstall built-ins, create user dirs, shell integration
ayo setup --force           # Overwrite modifications without prompting

# Chat
ayo                         # Start interactive chat with default @ayo agent
ayo "tell me a joke"        # Run single prompt with default @ayo agent
ayo @ayo                   # Start interactive chat session with agent
ayo @ayo "tell me a joke"  # Run single prompt (non-interactive)

# Agents management
ayo agents list             # List available agents
ayo agents show @ayo      # Show agent details
ayo agents create <handle>  # Create new agent
ayo agents dir              # Go to agents directory (interactive picker)
ayo agents update           # Update built-in agents
ayo agents update --force   # Update without checking for modifications

# Skills management
ayo skills list             # List available skills
ayo skills show <name>      # Show skill details
ayo skills create <name>    # Create new skill
ayo skills validate <path>  # Validate skill directory
ayo skills dir              # Go to skills directory (interactive picker)
ayo skills update           # Update built-in skills

# Shell integration
ayo init-shell              # Output shell integration script
```

### Shell Integration

For `ayo agents dir` and `ayo skills dir` to work interactively, add shell integration:

```bash
# Add to ~/.zshrc or ~/.bashrc:
eval "$(ayo init-shell)"
```

This enables:
- Interactive directory picker using `gum` (if installed)
- Tab completion for commands and agent handles
- Shell wrapper for cd-based commands

### Default Agent

When no agent is specified, ayo uses the `@ayo` agent (the default built-in agent).

### Interactive Mode

`ayo @agentname` starts an interactive chat session. The conversation continues until you exit with Ctrl+C.

- First Ctrl+C interrupts the current request
- Second Ctrl+C (at prompt) exits the session

### Non-Interactive Mode

`ayo @agentname "Your prompt here"` executes the prompt and exits.

## UI Behavior

Both interactive and non-interactive modes share the same UI components:

### Spinner Feedback

Spinners display progress during async operations:

- **LLM calls**: "Thinking..." while waiting, then "✓ Response received (elapsed)"
- **Tool calls**: Shows LLM-provided description (e.g., "Installing dependencies..."), then "✓ Installing dependencies (1.2s)" or "✕ ... failed (elapsed)"

### Styled Output

- Markdown rendering via glamour with syntax highlighting
- Tool outputs displayed in styled boxes
- Error messages with red styling and icons
- Reasoning/thinking displayed in bordered boxes

### Chat Header (Interactive Only)

Purple styled "Chat with @agentname" header with exit hint.

## Tool System

### Bash Tool

The `bash` tool is the default and primary tool. Agents use it to accomplish any task unless a more specific skill is available.

When calling bash, the LLM must provide:
- `command`: The shell command to execute
- `description`: Human-readable description shown in the spinner (e.g., "Running test suite")

Optional parameters:
- `timeout_seconds`: Command timeout (default 30s)
- `working_dir`: Working directory scoped to project root

### Skills

Skills extend agent capabilities by providing domain-specific instructions. Skills follow the [agentskills spec](https://agentskills.org).

Skills are discovered from multiple sources (in priority order):
1. **Agent-specific**: `{agent_dir}/skills/{skill-name}/`
2. **User shared**: `~/.config/ayo/skills/{skill-name}/`
3. **Built-in**: `~/.local/share/ayo/skills/{skill-name}/`

Each skill is a directory containing a `SKILL.md` with YAML frontmatter:

```markdown
---
name: my-skill
description: What this skill does and when to use it.
license: MIT
metadata:
  author: your-name
  version: "1.0"
---

# Skill Instructions

Detailed instructions for the agent...
```

**Required fields:**
- `name`: 1-64 chars, lowercase, hyphens ok, must match directory name
- `description`: 1-1024 chars, describes what the skill does and when to use it

**Optional fields:**
- `license`: License identifier
- `compatibility`: Environment requirements (max 500 chars)
- `metadata`: Key-value pairs (author, version, etc.)
- `allowed-tools`: Pre-approved tools (experimental)

**Optional directories:**
- `scripts/`: Executable code
- `references/`: Additional documentation
- `assets/`: Templates, data files

#### Agent Config for Skills

Agents can configure which skills are available:

```json
{
  "skills": ["skill-a", "skill-b"],
  "exclude_skills": ["unwanted-skill"],
  "ignore_builtin_skills": false,
  "ignore_shared_skills": false
}
```

#### Skills CLI Commands

```bash
ayo skills list                  # List all available skills
ayo skills list --source=built-in # Filter by source
ayo skills show <name>           # Show skill details
ayo skills validate <path>       # Validate a skill directory
ayo skills create <name>         # Create new skill from template
ayo skills create <name> --shared # Create in shared skills directory
```

#### Built-in Skills

Built-in skills are embedded in the binary and installed via `ayo setup`:

**Source (in repo):**
- Shared: `internal/builtin/skills/{skill-name}/`
- Agent-specific: `internal/builtin/agents/@ayo/skills/{skill-name}/`

**Installed to:**
- `~/.local/share/ayo/skills/`
- `~/.local/share/ayo/agents/@ayo/skills/`

**Current built-in skills:**
- `debugging` - Systematic debugging techniques
- `project-summary` - Project analysis and documentation (for @ayo)

## System Prompt Assembly

Messages are built in order:
1. Combined system (prefix + shared + agent + suffix)
2. Tools prompt (bash instructions)
3. Skills prompt (available skills XML)
4. User message

## Architecture Notes

- **Fantasy provider abstraction**: Uses `charm.land/fantasy` for provider-agnostic LLM calls. Supports OpenAI, Anthropic, Google, OpenRouter, and OpenAI-compatible providers.
- **Agent-based streaming**: Fantasy's `Agent` abstraction handles tool execution and multi-step interactions via callbacks (`OnTextDelta`, `OnToolCall`, `OnToolResult`, etc.)
- UI renders ordered tool outputs with spinner feedback

## Built-in Agents

Built-in agents are embedded in the binary and installed via `ayo setup`.

### Installation

**Source (in repo):** `internal/builtin/agents/{name}/`

**Installed to:** `~/.local/share/ayo/agents/`

**User agents:** `~/.config/ayo/agents/`

User agents take precedence over built-in agents with the same name.

### Structure

Each built-in agent directory contains:
```
internal/builtin/agents/{name}/
├── config.json      # Agent configuration
├── system.md        # System prompt (sandwiched between prefix/suffix)
└── skills/          # Optional agent-specific skills
    └── {skill}/
        └── SKILL.md
```

### Adding a Built-in Agent

1. Create directory: `internal/builtin/agents/{name}/`
2. Add `config.json`:
   ```json
   {
     "description": "Agent description",
     "allowed_tools": ["bash"]
   }
   ```
3. Add `system.md` with the agent's system prompt
4. Optionally add skills in `skills/{skill}/SKILL.md`
5. Bump `Version` constant in `internal/builtin/install.go`
6. The agent is automatically embedded via `//go:embed` and installed on next `ayo setup`

### Current Built-in Agents

- `@ayo` - The default agent, a versatile command-line assistant

The `ayo` namespace is reserved - users cannot create agents with the `@ayo` handle or `@ayo.` prefix.

## Versioning

Ayo uses semantic versioning (semver). The CLI version is defined in `internal/version/version.go`.

### Bumping the Version

When releasing a new version:

1. Update the `Version` constant in `internal/version/version.go`
2. Follow semver conventions:
   - **MAJOR** (1.0.0): Breaking changes
   - **MINOR** (0.2.0): New features, backward compatible
   - **PATCH** (0.1.1): Bug fixes, backward compatible

```go
// internal/version/version.go
const Version = "0.2.0"  // Example: bumping minor version
```

### Checking the Version

```bash
ayo --version
# Output: ayo version 0.1.0
```

## Agent Chaining

Agents can be composed via Unix pipes when they have structured input/output schemas. The output of one agent becomes the input to the next.

### Structured I/O Schemas

Agents can define optional JSON schemas:

- `input.jsonschema` - Validates input; agent only accepts JSON matching this schema
- `output.jsonschema` - Structures output; final response is cast to this format

Example agent structure:
```
@my-agent/
├── config.json
├── system.md
├── input.jsonschema    # Optional: structured input
└── output.jsonschema   # Optional: structured output
```

### Piping Agents

```bash
# Chain two agents (code reviewer -> issue reporter)
ayo @ayo.example.chain.code-reviewer '{"repo":".", "files":["main.go"]}' | ayo @ayo.example.chain.issue-reporter
```

**Pipeline behavior:**
- Stdin is piped → agent reads JSON from stdin
- Stdout is piped → UI goes to stderr, raw JSON goes to stdout
- Full UI (spinners, reasoning, tool calls) always visible on stderr

### Schema Compatibility

When piping agents:

1. **Exact match**: Output schema identical to input schema
2. **Structural match**: Output has all required fields of input (superset OK)
3. **Freeform**: Target agent has no input schema (accepts anything)

If schemas are incompatible, validation fails with a clear error.

### Chain Discovery Commands

```bash
# List all chainable agents (have input or output schema)
ayo chain ls

# Show agent's schemas
ayo chain inspect @ayo.debug.structured-io

# Find agents that can receive this agent's output
ayo chain from @ayo.example.chain.code-reviewer

# Find agents whose output this agent can receive
ayo chain to @ayo.example.chain.issue-reporter

# Validate JSON against agent's input schema
ayo chain validate @ayo.debug.structured-io '{"environment": "staging", "service": "api"}'
echo '{"environment": "staging", "service": "api"}' | ayo chain validate @ayo.debug.structured-io

# Generate example input for an agent
ayo chain example @ayo.debug.structured-io
```

### Chain Context

When agents are chained, context is passed via environment variable:
- `AYO_CHAIN_CONTEXT` contains JSON with `depth`, `source`, and `source_description`
- Freeform agents receive a preamble describing the chain context

### Example Chain Agents

Built-in example agents demonstrating chaining:

```bash
# Code reviewer outputs structured findings
ayo @ayo.example.chain.code-reviewer '{"repo":".", "files":["main.go"]}'

# Issue reporter consumes code reviewer output
ayo @ayo.example.chain.code-reviewer '{"repo":".", "files":["main.go"]}' \
  | ayo @ayo.example.chain.issue-reporter
```
