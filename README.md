# ayo

A command-line tool for running AI agents with tool execution, skills, and agent chaining.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Running Agents](#running-agents)
- [Commands](#commands)
  - [setup](#setup)
  - [init-shell](#init-shell)
  - [agents](#agents)
  - [skills](#skills)
  - [chain](#chain)
  - [completion](#completion)
- [Configuration](#configuration)
- [Directory Structure](#directory-structure)
- [Agent Chaining](#agent-chaining)

---

## Installation

```bash
go install ./cmd/ayo
```

Built-in agents and skills are automatically installed on first run. No manual setup required.

To reinstall built-ins (e.g., after modifying them) or set up shell integration:

```bash
ayo setup              # Reinstall built-ins and configure shell integration
ayo setup --force      # Overwrite modifications without prompting
```

For project-local installation (useful for development or project-specific agents):

```bash
ayo setup --dev
```

This installs to `./.local/share/ayo` and `./.config/ayo` in your current directory instead of the global locations.

---

## Quick Start

```bash
# Start interactive chat with the default @ayo agent
ayo

# Run a single prompt
ayo "tell me a joke"

# Chat with a specific agent
ayo @ayo

# Run a prompt with file attachments
ayo -a file.txt "summarize this"
```

---

## Running Agents

### Usage

```
ayo [command] [@agent] [prompt] [--flags]
```

### Interactive Mode

Start an interactive chat session that continues until you exit:

```bash
ayo              # Chat with default @ayo agent
ayo @ayo         # Chat with a specific agent
```

- First `Ctrl+C` interrupts the current request
- Second `Ctrl+C` (at prompt) exits the session

### Non-Interactive Mode

Execute a single prompt and exit:

```bash
ayo "your prompt here"
ayo @ayo "explain this code"
```

### File Attachments

Attach files to provide context:

```bash
ayo -a file.txt "summarize this"
ayo -a src/main.go -a src/utils.go "review this code"
```

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--attachment` | `-a` | File attachments (can be repeated) |
| `--config` | | Path to config file |
| `--debug` | | Show debug output including raw tool payloads |
| `--help` | `-h` | Help for ayo |
| `--version` | `-v` | Version for ayo |

---

## Commands

### setup

Reinstall built-in agents and skills, create user directories, and configure shell integration.

Built-ins are automatically installed on first run, so this command is only needed to:
- Reinstall after modifying built-in agents/skills
- Set up shell integration for completions and helper functions
- Install to a project-local directory with `--dev`

```bash
ayo setup              # Reinstall built-ins and configure shell integration
ayo setup --dev        # Project-local setup to ./.config/ayo and ./.local/share/ayo
ayo setup --force      # Overwrite modifications without prompting
```

| Flag | Short | Description |
|------|-------|-------------|
| `--dev` | | Install to local project directories instead of global |
| `--force` | `-f` | Overwrite modifications without prompting |

---

### init-shell

Output shell initialization script for completions and helper functions.

```bash
# Add to your .bashrc or .zshrc
eval "$(ayo init-shell)"
```

This enables:
- Shell completions for commands and agent handles
- Helper functions for `ayo agents dir` and `ayo skills dir` (interactive directory selection with [gum](https://github.com/charmbracelet/gum))

---

### agents

Manage agents: list, show, create, update, and navigate.

```bash
ayo agents list                    # List all available agents
ayo agents list --source=user      # Filter by source (user, built-in)
ayo agents show @ayo               # Show agent details
ayo agents create myagent          # Create a new agent
ayo agents dir                     # Go to agents directory (requires shell integration)
ayo agents update                  # Update built-in agents
ayo agents update --force          # Overwrite without checking for modifications
```

#### agents list

```bash
ayo agents list [--source=<source>]
```

| Flag | Description |
|------|-------------|
| `--source` | Filter by source: `user`, `built-in` |

#### agents show

Display detailed information about an agent including its system prompt, skills, and configuration.

```bash
ayo agents show @ayo
```

#### agents create

Create a new agent interactively or with flags.

```bash
ayo agents create @myagent
ayo agents create @myagent --description "My custom agent"
ayo agents create @helper --model gpt-4.1 --system "You are a helpful assistant"
```

| Flag | Description |
|------|-------------|
| `--description` | Agent description |
| `--model` | Model to use |
| `--system` | System message text |
| `--system-file` | Path to system message file |
| `--ignore-shared` | Ignore shared system message |

#### agents dir

Navigate to an agents directory. Requires shell integration (`eval "$(ayo init-shell)"`).

With [gum](https://github.com/charmbracelet/gum) installed, provides interactive selection between user and built-in directories.

#### agents update

Update built-in agents to the latest version embedded in the binary.

```bash
ayo agents update           # Check for modifications first
ayo agents update --force   # Overwrite without prompting
```

---

### skills

Manage skills: list, show, create, validate, update, and navigate.

Skills extend agent capabilities with domain-specific instructions following the [agentskills spec](https://agentskills.org).

```bash
ayo skills list                      # List all available skills
ayo skills list --source=built-in    # Filter by source
ayo skills show debugging            # Show skill details
ayo skills create my-skill           # Create a new skill (agent-specific)
ayo skills create my-skill --shared  # Create in shared skills directory
ayo skills validate ./path           # Validate a skill directory
ayo skills dir                       # Go to skills directory
ayo skills update                    # Update built-in skills
ayo skills update --force            # Overwrite without prompting
```

#### skills list

```bash
ayo skills list [--source=<source>]
```

| Flag | Description |
|------|-------------|
| `--source` | Filter by source: `agent`, `user`, `installed`, `built-in` |

#### skills show

Display detailed information about a skill including its description, metadata, and full content.

```bash
ayo skills show debugging
```

#### skills create

Create a new skill from template.

```bash
ayo skills create <name> [--flags]
```

| Flag | Description |
|------|-------------|
| `--shared` | Create in shared skills directory (available to all agents) |

Without `--shared`, creates an agent-specific skill.

#### skills validate

Validate a skill directory against the agentskills spec.

```bash
ayo skills validate ./my-skill
ayo skills validate ~/.config/ayo/skills/debugging
```

#### skills dir

Navigate to a skills directory. Requires shell integration.

#### skills update

Update built-in skills to the latest version embedded in the binary.

```bash
ayo skills update           # Check for modifications first
ayo skills update --force   # Overwrite without prompting
```

---

### chain

Commands for discovering compatible agents and validating chain connections.

Agents with structured input/output schemas (JSON Schema) can be composed via Unix pipes.

```bash
ayo chain ls                              # List all chainable agents
ayo chain ls --json                       # Output as JSON
ayo chain inspect @ayo.debug.structured-io # Show agent's input and output schemas
ayo chain from @ayo.example.chain.code-reviewer  # List agents that can receive this agent's output
ayo chain to @ayo.example.chain.issue-reporter   # List agents whose output this agent can receive
ayo chain validate @ayo.debug.structured-io '{"environment": "staging", "service": "api"}'  # Validate JSON against input schema
ayo chain example @ayo.debug.structured-io  # Generate example input JSON
```

#### chain ls

List all agents that have input or output schemas (chainable agents).

```bash
ayo chain ls [--json]
```

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON |

#### chain inspect

Show an agent's input and output schemas.

```bash
ayo chain inspect @ayo.debug.structured-io [--json]
```

#### chain from

Find agents that can receive output from the specified agent.

```bash
ayo chain from @ayo.example.chain.code-reviewer
```

#### chain to

Find agents whose output the specified agent can receive.

```bash
ayo chain to @ayo.example.chain.issue-reporter
```

#### chain validate

Validate JSON against an agent's input schema.

```bash
ayo chain validate @ayo.debug.structured-io '{"environment": "staging", "service": "api"}'
echo '{"environment": "staging", "service": "api"}' | ayo chain validate @ayo.debug.structured-io
```

#### chain example

Generate example input JSON based on an agent's input schema.

```bash
ayo chain example @ayo.debug.structured-io
```

---

### completion

Generate shell autocompletion scripts.

#### Bash

```bash
# Current session
source <(ayo completion bash)

# Linux (persistent)
ayo completion bash > /etc/bash_completion.d/ayo

# macOS with Homebrew (persistent)
ayo completion bash > $(brew --prefix)/etc/bash_completion.d/ayo
```

#### Zsh

```bash
# Current session
source <(ayo completion zsh)

# Linux (persistent)
ayo completion zsh > "${fpath[1]}/_ayo"

# macOS with Homebrew (persistent)
ayo completion zsh > $(brew --prefix)/share/zsh/site-functions/_ayo
```

#### Fish

```bash
# Current session
ayo completion fish | source

# Persistent
ayo completion fish > ~/.config/fish/completions/ayo.fish
```

#### PowerShell

```powershell
# Current session
ayo completion powershell | Out-String | Invoke-Expression
```

---

## Configuration

### Config File

The main config file is located at `~/.config/ayo/config.yaml`:

```yaml
# Override default directories
agents_dir: ~/.config/ayo/agents
skills_dir: ~/.config/ayo/skills

# System prompts (applied to all agents)
shared_system_message: ~/.config/ayo/prompts/system.md
system_prefix: ~/.config/ayo/prompts/prefix.md
system_suffix: ~/.config/ayo/prompts/suffix.md

# Model configuration
default_model: gpt-4.1

# Provider configuration
provider:
  name: openai
  id: openai
  api_endpoint: https://api.openai.com/v1
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | API key for OpenAI provider |
| `ANTHROPIC_API_KEY` | API key for Anthropic provider |
| `CATWALK_URL` | URL for Catwalk model proxy (default: `http://localhost:8080`) |

---

## Directory Structure

### Load Priority

Ayo searches for agents and skills in the following order (first found wins):

| Priority | Path | Description |
|----------|------|-------------|
| 1 | `./.config/ayo/` | Local project config |
| 2 | `./.local/share/ayo/` | Local project data |
| 3 | `~/.config/ayo/` | User config |
| 4 | `~/.local/share/ayo/` | Built-in data |

This allows project-specific agents/skills to override user and built-in ones.

### Platform Directories

| Platform | User Config | Built-in Data |
|----------|-------------|---------------|
| macOS/Linux | `~/.config/ayo/` | `~/.local/share/ayo/` |
| Windows | `%LOCALAPPDATA%\ayo\` | `%LOCALAPPDATA%\ayo\` |

### Directory Layout

```
~/.config/ayo/                    # User configuration
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

~/.local/share/ayo/               # Built-in data (managed by ayo setup)
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

### Local Project Setup

With `ayo setup --dev`, ayo creates project-local directories:

```
./                               # Your project root
├── .config/
│   └── ayo/                     # Local project config
│       ├── agents/
│       └── skills/
└── .local/
    └── share/
        └── ayo/                 # Local project data
            ├── agents/
            └── skills/
```

These directories are automatically added to `.gitignore` and `.crushignore`.

---

## Agent Chaining

Agents with structured input/output schemas can be composed via Unix pipes.

### Schema Files

Agents can define optional JSON schemas:

```
@my-agent/
├── config.json
├── system.md
├── input.jsonschema    # Optional: validates input JSON
└── output.jsonschema   # Optional: structures output JSON
```

### Piping Agents

```bash
# Chain two agents (code reviewer -> issue reporter)
ayo @ayo.example.chain.code-reviewer '{"repo":".", "files":["main.go"]}' | ayo @ayo.example.chain.issue-reporter
```

### Pipeline Behavior

- **Stdin is piped**: Agent reads JSON from stdin
- **Stdout is piped**: UI goes to stderr, raw JSON goes to stdout
- **Full UI visible**: Spinners, reasoning, and tool calls always appear on stderr

### Schema Compatibility

When piping agents, ayo validates that:

1. **Exact match**: Output schema identical to input schema
2. **Structural match**: Output has all required fields of input (superset OK)
3. **Freeform**: Target agent has no input schema (accepts anything)

### Discovery Commands

```bash
# List all chainable agents
ayo chain ls

# Show agent's schemas
ayo chain inspect @ayo.example.chain.code-reviewer

# Find compatible agents
ayo chain from @ayo.example.chain.code-reviewer   # What can receive this output?
ayo chain to @ayo.example.chain.issue-reporter    # What can feed this input?

# Validate and test
ayo chain validate @ayo.debug.structured-io '{"environment": "staging", "service": "api"}'
ayo chain example @ayo.debug.structured-io
```

---

## License

See [LICENSE](LICENSE) for details.
