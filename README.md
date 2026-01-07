# ayo

A command-line tool for running AI agents.

## Installation

```bash
go install ./cmd/ayo
```

After installation, run the setup command to install built-in agents, skills, and configure shell integration:

```bash
ayo setup
```

## Usage

```
ayo [command] [@agent] [prompt] [--flags]
```

### Running Agents

```bash
# Start interactive chat with the default @ayo agent
ayo

# Run a single prompt (non-interactive)
ayo "tell me a joke"

# Start interactive chat with a specific agent
ayo @myagent

# Run a single prompt with a specific agent
ayo @myagent "explain this code"

# Attach files to your prompt
ayo -a file.txt "summarize this"
```

### Global Flags

| Flag | Description |
|------|-------------|
| `-a, --attachment` | File attachments |
| `--config` | Path to config file |
| `--debug` | Show debug output including raw tool payloads |
| `-h, --help` | Help for ayo |
| `-v, --version` | Version for ayo |

## Commands

### setup

Runs complete ayo setup: installs built-in agents and skills, creates user directories, and configures shell integration.

```bash
ayo setup
ayo setup --force  # Overwrite modifications without prompting
```

### init

Create a new agent.

```bash
ayo init @myagent
ayo init --handle myagent --description "My custom agent"
```

| Flag | Description |
|------|-------------|
| `--handle` | Agent handle (without @) |
| `--description` | Agent description |
| `--model` | Model id |
| `--system` | System message text (overrides file) |
| `--system-file` | Path to system message file |
| `--ignore-shared` | Ignore shared system message |

### init-shell

Output shell initialization script. Add to your `.bashrc` or `.zshrc`:

```bash
eval "$(ayo init-shell)"
```

This enables shell completions and helper functions for commands like `ayo agents dir` and `ayo skills dir`.

### agents

Manage agents.

```bash
ayo agents list              # List all available agents
ayo agents list --source=user     # Filter by source (user, built-in)
ayo agents show @myagent     # Show agent details
ayo agents create myagent    # Create a new agent
ayo agents dir               # Go to agents directory (requires shell integration)
ayo agents update            # Update built-in agents
ayo agents update --force    # Overwrite without checking for modifications
```

#### agents create

```bash
ayo agents create <handle> [--flags]
```

| Flag | Description |
|------|-------------|
| `--description` | Agent description |
| `--model` | Model to use |
| `--system` | System message |

### skills

Manage skills. Skills extend agent capabilities with domain-specific instructions.

```bash
ayo skills list              # List all available skills
ayo skills list --source=built-in  # Filter by source (agent, user, installed, built-in)
ayo skills show debugging    # Show skill details
ayo skills create my-skill   # Create a new skill
ayo skills create my-skill --shared  # Create in shared skills directory
ayo skills validate ./path   # Validate a skill directory
ayo skills dir               # Go to skills directory (requires shell integration)
ayo skills update            # Update built-in skills
ayo skills update --force    # Overwrite without checking for modifications
```

### chain

Commands for discovering compatible agents and validating chain connections. Agents with structured input/output schemas can be composed via Unix pipes.

```bash
ayo chain ls                 # List all chainable agents
ayo chain ls --json          # Output as JSON
ayo chain inspect @agent     # Show agent's input and output schemas
ayo chain from @agent        # List agents that can receive output from this agent
ayo chain to @agent          # List agents whose output this agent can receive
ayo chain validate @agent '{"key": "value"}'  # Validate JSON against input schema
ayo chain example @agent     # Generate example input JSON for an agent
```

#### Piping Agents

```bash
ayo @agent-1 '{"input": "data"}' | ayo @agent-2
```

When piping, the UI (spinners, reasoning, tool calls) goes to stderr while raw JSON output goes to stdout.

### completion

Generate shell autocompletion scripts.

#### bash

```bash
# Current session
source <(ayo completion bash)

# Linux (persistent)
ayo completion bash > /etc/bash_completion.d/ayo

# macOS (persistent)
ayo completion bash > $(brew --prefix)/etc/bash_completion.d/ayo
```

#### zsh

```bash
# Current session
source <(ayo completion zsh)

# Linux (persistent)
ayo completion zsh > "${fpath[1]}/_ayo"

# macOS (persistent)
ayo completion zsh > $(brew --prefix)/share/zsh/site-functions/_ayo
```

#### fish

```bash
# Current session
ayo completion fish | source

# Persistent
ayo completion fish > ~/.config/fish/completions/ayo.fish
```

#### powershell

```powershell
# Current session
ayo completion powershell | Out-String | Invoke-Expression
```

## Configuration

Ayo uses the following directories:

| Platform | User Config | Built-in Data |
|----------|-------------|---------------|
| macOS/Linux | `~/.config/ayo/` | `~/.local/share/ayo/` |
| Windows | `%LOCALAPPDATA%\ayo\` | `%LOCALAPPDATA%\ayo\` |

The config file is located at `~/.config/ayo/config.yaml`.
