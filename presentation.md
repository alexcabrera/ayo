# `ayo` is a cli for agents you orchestrate

- Define agents with Markdown
- Add skills to work with external services
- Chain agents together with structured I/O

---

## 2025 in Review

Successful deployments of _AI Agents™_ in the enterprise has largely consisted:

1. Engineers employing agents to review and generate code
2. SMEs rolling their own solutions by wiring third-party services into agents

The latter is why `n8n` has seen a meteoric rise.

### Wildly Effective

I led two `n8n` implementations in 2025:

- The first, generated $3M in new revenue and pre-sold near $4M for 2026;
- the second is projected to save $35M over the next two years in operational costs.

---

## Theory of the case 

These integrations are because SMEs know their problem domains far better than the product engineering organization.

> "The goal is to become HBO faster than HBO can become us"
> -- Ted Sarandos, Netflix CCO, 2013

The future of enterprise tooling is going be turning individual SMEs on teams into the managers of agents that act as a dedicated interal product development teams.

> Netflix to Buy Warner Bros. and HBO Max
> -- Dec. 5, 2025

---

## Plugging in nodes is still programming.

Most of the issues and fragility in scaled `n8n` deployments comes from users needing to feel their way into some basic programing fundamentals. Even then, most experienced engineers will recoil when they see these workflows. 

The agents they write, however, are **very** impressive.

---

## How `ayo` works

The `ayo` executable installs markdown/jsonschema defined agents into `~/.local/share/ayo/@<agentname>`. By default there is an `@ayo` agent.

```bash
ayo @ayo "tell me a joke"
```

```
→ @ayo
  Why did the scarecrow win an award?
  Because he was outstanding in his field!
```

---

### Agent Structure

The agent structure looks like this:

```
  @my-agent/
  ├── config.json          # Required: Agent configuration
  ├── system.md            # Required: System prompt (instructions)
  ├── input.jsonschema     # Optional: Structured input validation
  ├── output.jsonschema    # Optional: Structured output format
  └── skills/              # Optional: Agent-specific skills
      └── my-skill/
          └── SKILL.md
```

---

#### Agent Configuration

```
  {
    "description": "Analyzes log files and extracts error patterns",
    "allowed_tools": ["bash"],
    "skills": ["debugging"],
    "exclude_skills": [],
    "ignore_builtin_skills": false,
    "ignore_shared_skills": false
  }
```

---

#### Agent System Message


