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

## Hook Delivery check {{< new-badge v3.1.4 >}}

`moai update` can pass over hook entries the template newly ships without ever landing them in an existing project's `.claude/settings.json`. The **Hook Delivery** check finds those gaps: within the hook event keys the project already carries, it compares the shipped template against the project file and reports entries the template carries but the project lacks.

| Reported item | Content |
|---------------|---------|
| Missing entries | Named as `hooks.PreToolUse missing handle-pre-tool.sh (matcher AskUserQuestion)` — the event key and matcher included |
| How to fix | Re-add each missing entry under the named event key, copying the block from the template settings of your moai version |
| Post-update verification | The command that shows whether `moai update` deleted managed files (`git status --porcelain \| grep '^ D'`) — on a hit, restore with `git restore -- <path>` before re-adding the entries |

The check is read-only — it never writes `.claude/settings.json`. Event keys the template introduces for the first time and entries you authored yourself are never counted as missing, and opted-out entries are not expected either (the template renders with your project's own `hook.opt_in.enabled` setting). When everything matches, the check reports `ok`.

## Codex Wiring check {{< new-badge v3.1.4 >}}

A full `moai doctor` run carries a **Codex Wiring** entry. It examines whether the project's Codex wiring — the generated hooks, the MCP registration, and the skill mirror — is intact. The check is advisory, fail-open, and read-only: it reports what it finds and never repairs anything itself.

| Check item | Content |
|------------|---------|
| `.codex/hooks.json` presence · key whitelist | Whether the wiring file exists and its keys match the whitelist. One stray key makes codex silently ignore the whole file, so this check observes that silence in its place |
| Sidecar hash | Compares the current file against the hook-content hash recorded at deploy time, catching the divergence left behind after hooks were hand-edited |
| `moai` binary PATH | Generated hook commands are `moai hook ...`, so nothing fires if `moai` cannot be found on PATH |
| `[mcp_servers.moai]` in `.codex/config.toml` | Whether the MCP registration table exists and matches the standard registration shape. The table is user-owned, so doctor reports it and never repairs it |
| `.agents/skills` skill mirror | Codex CLI does not scan `.claude/skills`, so without the mirror — or with a broken link — codex sees none of this project's MoAI skills |
| User-layer `[[skills.config]]` registrations | Inspects the skill registrations in `~/.codex/config.toml` (`$CODEX_HOME/config.toml` when set) for path and `enabled` key shape. This item is checked only when codex is in play — that is, when the project is wired or codex is installed |

When the check finds a problem, the fix directive baked into the code is shown alongside.

| Finding | Directive |
|---------|-----------|
| Project with no wiring at all | `moai init --llm codex` |
| Sidecar divergence after hook changes | Re-trust the changed hooks with `codex /hooks` |
| Missing skill mirror · broken link | `moai update --templates-only --force --yes` |
| Registration whose skill file has vanished | Remove the entry or restore the skill file |

Exactly one of the findings this check can produce is classified fatal: if a user-layer `[[skills.config]]` entry lacks the `enabled` key or carries a value that is not a bare TOML boolean, codex exits 1 on every invocation. Hitting this single fatal finding turns the doctor result into Fail and carries the exit code 1. This behavior was observed on codex-cli 0.153.4 — it is confirmed for that release only and is not generalized to other releases. The correct shape is to state `enabled = true` or `enabled = false` explicitly on every entry. Every finding other than this one is advisory and does not change doctor's exit code.

On a machine without codex, a claude-only project carrying no wiring makes this check skip silently — an informational skip that produces no warning row. For reference, bulk-collecting ghost registrations that point at skill files that no longer exist sits outside this check's directives; it is a separate verb, `moai clean --codex-skills`.

Codex wiring overall — the hook trust model and the skill mirror included — is covered in detail in [Codex Dual Harness](/en/advanced/codex-dual-harness).

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
