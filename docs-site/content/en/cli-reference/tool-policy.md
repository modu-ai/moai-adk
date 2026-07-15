---
title: moai tool-policy Tool Policy
weight: 88
draft: false
---

`moai tool-policy` manages the tool/permission policy SSOT. `.moai/config/sections/tool-policy.yaml` is the single source of truth; from it, the permissions block of `settings.json` is generated (codegen) and policy entries are queried.

## Subcommands

| Command | Description |
|--------|------|
| `moai tool-policy build` | Regenerate the settings.json permissions block from tool-policy.yaml |
| `moai tool-policy list` | List tool-policy entries (thin query) |

## moai tool-policy build

```bash
moai tool-policy build
moai tool-policy build --local-only
```

Regenerates the permissions block of the local `.claude/settings.json` and the template `settings.json.tmpl`.

| Flag | Description |
|--------|------|
| `--repo-root <path>` | Repository root (default: cwd) |
| `--policy <path>` | tool-policy.yaml path (default: `<repo-root>/.moai/config/sections/tool-policy.yaml`) |
| `--local-only` | Regenerate only the local `.claude/settings.json` (skip the template .tmpl) |
| `--template-only` | Regenerate only the template settings.json.tmpl (skip the local) |
| `--default-mode <mode>` | Override `permissions.defaultMode` (default: preserve the existing value) |
| `--json` | Output the result as JSON |

## moai tool-policy list

```bash
moai tool-policy list
moai tool-policy list --risk-tier irreversible --decision deny
```

| Flag | Description |
|--------|------|
| `--risk-tier <read\|write\|irreversible>` | Filter by risk tier |
| `--decision <allow\|deny\|ask>` | Filter by decision |
| `--tool <name>` | Filter by tool name (exact match) |
| `--format <text\|json>` | Output format |
| `--repo-root <path>` | Repository root (default: cwd) |
| `--policy <path>` | tool-policy.yaml path |

## Related documents

- [settings.json Guide](/en/advanced/settings-json) — permissions block details
- [config Section Reference](/en/advanced/config-sections)
- [CLI Overview](/en/getting-started/cli)
