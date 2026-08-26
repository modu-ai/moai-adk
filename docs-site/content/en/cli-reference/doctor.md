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

## Home Disk Usage check {{< new-badge v3.1.1 >}}

A full `moai doctor` run carries a **Home Disk Usage** entry. It reports how full the `~/.moai` home directory is and is **advisory**: exceeding the threshold never blocks another command.

| Reported item | Content |
|---------------|---------|
| Total size | The total `~/.moai` footprint plus its three largest entries |
| Per-profile breakdown | The size of each `claude-profiles/<profile>` with its category split |
| Release count | How many binaries remain in `releases/`, and the current version |
| Cleanable bytes | The estimate of what `moai clean --home` could actually delete |
| `~/.claude` | Size only — never a cleanup target on any path |

When the cleanable estimate exceeds the threshold (a compiled default of 500 MB) the status turns WARN and recommends `moai clean --home` (dry-run by default). Below it, the status stays OK. When no `~/.moai` exists at all, the check reports "nothing to report" and passes.

The estimate calls **the same scanner** `moai clean --home` uses, so the number doctor quotes and the list clean actually deletes cannot drift apart. Full detail: [Home Directory Hygiene](/en/advanced/home-hygiene).

## Exit codes

Scripts and CI wrappers calling `moai doctor` read the exit code, not the summary line.

| Exit code | Meaning |
|-----------|---------|
| `0` | No failing check. Warnings are advisory and do not change the exit code |
| `1` | One or more checks failed — the summary's `Fail N` carried through |

The Constitution Registry check does more than confirm the registry parses: it runs the **same drift validation** as `moai constitution validate`. Doctor therefore cannot report ok on a checkout where validate fails. Bypassing with `MOAI_CONSTITUTION_SKIP_VALIDATE=1` returns doctor to its structural verdict.

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

Related: [Project Status](/en/cli-reference/status) · [CLI Overview](/en/getting-started/cli)
