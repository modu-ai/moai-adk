---
title: Profile Management
weight: 40
draft: false
---

MoAI-ADK's profile system lets you manage multiple Claude Code configurations in isolation. Separate work vs. personal setups, or high-quality vs. cost-saving sessions, into one profile each, and you no longer need to change model, language, and display settings every time.

## What is a Profile?

A profile is an **isolated Claude Code configuration directory** (`CLAUDE_CONFIG_DIR`). Each profile maintains independent settings, model selection, and language environment.

```
~/.moai/claude-profiles/
├── default/           # Default profile
│   ├── settings.json
│   └── settings.local.json
├── work/              # Work profile
│   ├── settings.json
│   └── settings.local.json
└── personal/          # Personal profile
    └── ...
```

## Command Reference

### moai profile list

Displays all available profiles.

```bash
moai profile list
```

### moai profile setup [name]

Runs the interactive setup wizard.

```bash
moai profile setup          # Set up the default profile
moai profile setup work     # Set up the "work" profile
```

**Wizard configuration items:**
- **Identity**: user name, role
- **Languages**: conversation language, code comment language
- **Model Settings**: default model, 1M context model selection
- **Display**: output style, status line settings

### moai profile current

Displays the name of the currently active profile.

```bash
moai profile current
```

### moai profile delete [name]

Deletes a profile.

```bash
moai profile delete old-profile
```

## Running Claude Code with a Profile

Specify a profile with the `-p` (or `--profile`) flag.

```bash
moai cc -p work          # Run Claude with the work profile
moai glm -p cost-save    # Run GLM with the cost-save profile
moai cg -p team          # Run CG mode with the team profile
```

{{< callout type="info" >}}
When no profile is specified, the default profile is used. On first run, the setup wizard starts automatically.
{{< /callout >}}

## Selecting a 1M Context Model

When configuring a profile, you can select a model that supports a 1M context window. The `[1m]` suffix is not a separate model — it is Claude Code's native context-window modifier.

**Selectable model aliases:**
- `opus` / `opus[1m]`
- `sonnet` / `sonnet[1m]`
- `fable` / `fable[1m]`

Select it in the "Model Settings" step of the setup wizard, or edit the profile settings file directly. 1M context models are well suited for analyzing large codebases or working with long documents.

## Behavior on Profile Switch

| Switch | Behavior |
|------|------|
| `moai cc` → `moai glm` | GLM environment variables injected automatically |
| `moai glm` → `moai cc` | GLM environment variables removed automatically |
| `moai cc` → `moai cg` | GLM env injected into the tmux session only; the Leader stays on Claude |

## Related Documents

- [CLI Reference](/en/getting-started/cli) - Full CLI commands
- [Quick Start](/en/getting-started/quickstart) - Getting started
- [Initial Setup](/en/getting-started/init-wizard) - Project initialization
