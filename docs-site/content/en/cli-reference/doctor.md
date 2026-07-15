---
title: moai doctor Diagnostics
weight: 60
draft: false
---

`moai doctor` runs comprehensive system diagnostics. It checks Claude Code configuration, dependencies, project structure, language-specific development tools, and the environment, and can suggest fixes for detected issues.

## Overview

```bash
moai doctor [OPTIONS]
```

## Flags

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Show detailed diagnostic information (tool versions, language detection) |
| `--fix` | Suggest fixes for detected issues |
| `--export` | Export diagnostics to a JSON file |
| `--check <tool>` | Run a specific check only (e.g. git, go, config) |

## Subcommands

`moai doctor` provides subcommands that dive deeper into a specific area.

| Command | Description |
|---------|-------------|
| `moai doctor config` | Configuration diagnostics — inspect merged settings with provenance |
| `moai doctor hook` | Show the 27-event hook coverage table |
| `moai doctor permission` | Diagnose permission resolution |
| `moai doctor sandbox` | Sandbox backend availability diagnostics |

`moai doctor config` in turn offers `dump` (dump merged settings) and `diff <tier-a> <tier-b>` (compare two settings tiers).

## Examples

```bash
# Full diagnostics
moai doctor

# Detailed diagnostics
moai doctor --verbose

# Export diagnostics
moai doctor --export diagnostics.json

# Diagnose a specific area
moai doctor hook          # hook coverage table
moai doctor permission    # permission resolution
moai doctor sandbox       # sandbox backend
```

---

Related: [Project Status](/cli-reference/status) · [CLI Overview](/getting-started/cli)
