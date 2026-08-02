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

This value reflects the global record, so when different projects remember different profiles, read it together with the limitations in [Automatic Profile Selection](#automatic-profile-selection).

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
A profile given with `-p` always takes precedence. To see which profile is used when you do not specify one, read [Automatic Profile Selection](#automatic-profile-selection) below. The first time a profile is used, the setup wizard starts automatically.
{{< /callout >}}

## Automatic Profile Selection

Running `moai cc` without `-p` picks a profile by consulting the record in `~/.moai/claude-profiles/launch.yaml`. That record is updated every time you launch a named profile with `-p`.

{{< callout type="note" >}}
The per-project memory described below ships in the next release. The currently released version keeps only a single global record (`last_profile`), so specifying a profile with `-p` in project B overwrites the value project A had remembered.
{{< /callout >}}

Alongside the global record, `launch.yaml` maintains a `projects:` map keyed by the project's absolute path. A launch without `-p` resolves the profile in this order:

1. The profile the current project remembers (a `projects:` entry)
2. The global record (`last_profile`)
3. The default profile

If a recorded profile's directory has already been deleted, it is skipped and resolution moves on to the next step. A name given with `-p` takes precedence over this entire order, and `-p default` can name the default profile explicitly.

To turn off both lookups, set the environment variable:

```bash
MOAI_NO_PROFILE_FALLBACK=1 moai cc    # ignore the record and run the default profile
```

The per-project record is written when you launch with `-p`, and it is updated as well when you switch profiles in the [Web Console](/en/cli-reference/web). The default profile (`default`) is never recorded.

**Limitations to know**

- Moving or renaming a project directory leaves the existing entry matching no path. The entry is skipped silently, so it does not break launching.
- The `projects:` map grows as projects accumulate, and there is no command to prune it yet.
- `moai profile current` reports the global record as-is. So in a project whose remembered profile differs from the global record, the name `moai profile current` gives you can differ from the profile that `moai cc` without `-p` actually launches.

## First Launch of a New Profile

A newly created profile directory does not yet contain `.claude.json`, the file that holds Claude Code's account state. Account state is kept per configuration directory on every platform, so even when your existing session is perfectly healthy, the first launch with a new profile brings up the login / onboarding screen.

{{< callout type="note" >}}
The notice below ships in the next release. The currently released version moves to the login screen with no warning at all.
{{< /callout >}}

Before starting Claude Code, the launcher prints the following to standard error:

```
Notice: profile "work" has no Claude Code configuration yet.
  Claude Code will show the login / onboarding screen on this launch.
  Account state is not inherited between profiles; sign in once and it
  persists for this profile.
```

Nothing is copied or moved into the new profile as credentials. Where account state lives differs by platform, so a copy tailored to one platform would be wrong on another. Once you sign in, that state stays with the profile, so this screen does not appear on later launches.

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

- [moai web Console](/en/cli-reference/web) - Switch and edit profiles from the browser
- [CLI Reference](/en/getting-started/cli) - Full CLI commands
- [Quick Start](/en/getting-started/quickstart) - Getting started
- [Initial Setup](/en/getting-started/init-wizard) - Project initialization
