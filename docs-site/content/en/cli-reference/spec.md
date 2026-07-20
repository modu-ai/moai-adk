---
title: moai spec Document Management
weight: 35
draft: false
---

`moai spec` manages the SPEC documents in the `.moai/specs/` directory. It provides subcommands for status updates, drift detection, acceptance-criteria viewing, EARS/GEARS linting, atomic closure, era auditing, and archiving.

## Subcommands

| Command | Description |
|--------|------|
| `moai spec status` | Update or list SPEC status |
| `moai spec drift` | Detect drift between frontmatter status and the git log |
| `moai spec view <SPEC-ID>` | View acceptance criteria as a tree |
| `moai spec lint [spec.md...]` | Lint for EARS compliance and structural validity |
| `moai spec close <SPEC-ID>` | Atomic 4-phase closure (status: completed + progress.md backfill) |
| `moai spec audit` | SPEC era classification and modern-era status drift audit |
| `moai spec archive` | Archive closed SPECs out of `.moai/specs/` |

## moai spec status

```bash
moai spec status <SPEC-ID> <new-status>   # Update status
moai spec status --list                   # List all SPECs
moai spec status --sync-git               # Sync status from the git log
```

| Flag | Description |
|--------|------|
| `--dry-run` | Preview changes without writing |
| `--list` | List all SPECs and their status |
| `--sync-git` | Sync SPEC status from main's git log |
| `--yes` | Non-interactive auto-confirmation for `--sync-git` (required for CI/pipes) |

## moai spec drift

```bash
moai spec drift
```

| Flag | Description |
|--------|------|
| `--json` | JSON-format output |
| `--exit-code-on-drift` | Exit code 1 when drift is detected |
| `--count` | Print only the drift count |
| `--no-cache` | Bypass the HEAD-SHA result cache and recompute |

## moai spec lint

```bash
moai spec lint [spec.md...]
```

| Flag | Description |
|--------|------|
| `--json` | JSON-format output |
| `--sarif` | SARIF 2.1.0 format output |
| `--strict` | Treat warnings as errors |
| `--format <fmt>` | Output format (table) |

## moai spec close

```bash
moai spec close SPEC-ID
```

Atomically transitions a SPEC to `status: completed` in a single commit.

| Flag | Description |
|--------|------|
| `--backfill-only` | Perform only the progress.md backfill |
| `--dry-run` | Preview without committing |
| `--force` | Force closure without confirmation |
| `--json` | JSON-format output |

## moai spec audit

```bash
moai spec audit
```

Scans `.moai/specs/SPEC-*/`, classifies each SPEC with the era heuristic, and detects modern-era status drift.

| Flag | Description |
|--------|------|
| `--json` | JSON-format output |
| `--filter-era <era>` | Filter by era |
| `--filter-spec <id>` | Filter by SPEC ID |
| `--include-grandfathered` | Include grandfather-era SPECs |
| `--strict` | Strict mode |

## moai spec archive

```bash
moai spec archive --dry-run   # Check targets (no move)
moai spec archive --yes       # Apply the plan
```

Archives terminal SPECs older than the grace window (default 90 days).

| Flag | Description |
|--------|------|
| `--dry-run` | Report the target set without moving |
| `--yes` | Confirm the move (required to apply) |
| `--grace-days <n>` | Grace-window days (0 = default 90) |
| `--json` | Output the plan as JSON |

## Related documents

- [SPEC-based Development](/en/core-concepts/spec-based-dev)
- [CLI Overview](/en/getting-started/cli)
