---
title: moai init Initialization
weight: 5
draft: false
---

`moai init` initializes a MoAI project in the current directory or a new folder. It lays down the `.claude/` and `.moai/` structure and settings required for Claude Code integration, and — when needed — configures the project mode, language, and quality gates through an interactive wizard.

## Usage

```bash
moai init [project-name]
```

| Pattern | Behavior |
|------|------|
| `moai init <name>` | Create a `./<name>/` folder and initialize inside it |
| `moai init .` | Initialize in the current directory |
| `moai init` | Initialize in the current directory (same as `moai init .`) |

Accepts at most 1 argument.

## Key Flags

### Deployment scope

| Flag | Description |
|--------|------|
| `--all` | Deploy the full catalog (core + optional packs + harness artifacts). The default is core-only slim mode |
| `--force` | Re-initialize an existing project (backs up the current `.moai/`) |
| `--no-hooks` | Skip git hook installation |

### Project defaults

| Flag | Description |
|--------|------|
| `--root <dir>` | Project root (default: current directory) |
| `--name <name>` | Project name (default: directory name) |
| `--language <lang>` | Primary programming language |
| `--framework <name>` | Framework (default: auto-detect or `none`) |
| `--mode <ddd\|tdd>` | Development methodology (default: tdd) |
| `--non-interactive` | Skip the interactive wizard — use flags and defaults only |

### Wizard phases

| Flag | Description |
|--------|------|
| `--standard` | Present Phase 1 questions (project mode, harness profile, LSP, quality gate, design) |
| `--advanced` | Present Phase 1 + Phase 2 questions (includes `--standard`) |
| `--project-mode <personal\|team>` | Project mode (default: personal) |
| `--harness-profile <name>` | Harness evaluation profile: default, strict, lenient, frontend |
| `--enable-lsp` | Enable LSP integration (default: false) |
| `--enforce-quality` | Enforce quality gates (default: true) |
| `--enable-design` | Enable the design workflow (default: true) |

### Git / model policy

| Flag | Description |
|--------|------|
| `--git-mode <manual\|personal\|team>` | Git workflow mode (default: manual) |
| `--git-provider <github\|gitlab>` | Git provider |
| `--github-username <name>` | GitHub username (required for personal/team mode) |
| `--model-policy <max\|medium\|low>` | Performance tier — stored in `performance_tier` of `llm.yaml` |
| `--plan-type <api\|subscription>` | Billing plan type — stored in `plan_type` of `llm.yaml` |

## Examples

```bash
# Initialize in a new folder
moai init my-app

# Initialize in the current directory
moai init .

# Specify a methodology
moai init --mode tdd

# Deploy the full catalog (bypass slim mode)
moai init --all

# Non-interactive (e.g. CI)
moai init . --non-interactive --language go
```

## Related commands

| Command | Description |
|--------|------|
| `moai update` | Sync templates for an initialized project |
| `moai status` | Check initialization status |
| `moai doctor` | Validate the environment after initialization |

## See also

- [Project Status](/en/cli-reference/status)
- [Update](/en/cli-reference/update)
- [CLI Overview](/en/getting-started/cli)
