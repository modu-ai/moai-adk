---
title: Profile Management
weight: 80
draft: false
---
# Profile Management


MoAI-ADK's profile system lets you manage multiple Claude Code configurations in isolation.

## What Is a Profile?

A profile is an **isolated Claude Code configuration directory** (`CLAUDE_CONFIG_DIR`). Each profile can maintain its own independent settings, model selection, and language environment.

```
~/.moai/claude-profiles/
├── default/           # default profile
│   ├── settings.json
│   └── settings.local.json
├── work/              # work profile
│   ├── settings.json
│   └── settings.local.json
└── personal/          # personal profile
    └── ...
```

## Command Reference

### moai profile list

Displays every available profile.

```bash
moai profile list
```

### moai profile setup [name]

Runs the interactive setup wizard.

```bash
moai profile setup          # set up the default profile
moai profile setup work     # set up the "work" profile
```

**Wizard configuration items:**
- **Identity**: User name, role
- **Languages**: Conversation language, code comment language
- **Model Settings**: Default model, 1M-context model selection
- **Display**: Output style, status line settings

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
moai cc -p work          # run Claude with the work profile
moai glm -p cost-save    # run GLM with the cost-save profile
moai cg -p team          # run CG mode with the team profile
```

{{< callout type="info" >}}
When no profile is specified, the default profile is used. The setup wizard starts automatically on first run.
{{< /callout >}}

## Choosing a 1M-Context Model

When configuring a profile, you can choose a model that supports the 1M-token context window.

**Supported models:**
- `claude-opus-4-8[1m]` - Opus 4.8 (1M context)
- `claude-sonnet-4-6[1m]` - Sonnet 4.6 (1M context)

Select it in the "Model Settings" step of the setup wizard, or edit the profile configuration file directly.

## Behavior on Profile Switch

| Switch | Behavior |
|------|------|
| `moai cc` → `moai glm` | GLM environment variables are injected automatically |
| `moai glm` → `moai cc` | GLM environment variables are removed automatically |
| `moai cc` → `moai cg` | GLM env is injected only into the tmux session; the Leader stays on Claude |

## Related Documentation

- [CLI Reference](/getting-started/cli) - Full CLI command reference
- [Quick Start](/getting-started/quickstart) - Getting started for the first time
- [Initial Setup](/getting-started/init-wizard) - Project initialization
