---
title: moai inventory Command
weight: 25
draft: false
---

This guide explains the `moai inventory` command for viewing active sessions, worktrees, and harnesses in your project.

{{< callout type="info" >}}
**One-line summary**: `moai inventory` provides a unified view of all active resources (sessions, worktrees, harnesses) in your project at a glance.
{{< /callout >}}

## Overview

`moai inventory` is a read-only command that provides an **integrated inventory** of your current project state.

### Query Targets

| Resource | Description | Location |
|----------|-------------|----------|
| **Active Sessions** | Currently running Claude Code sessions | `.moai/state/active-sessions.json` |
| **Worktrees** | L2/L3 isolated branches for projects | `~/.moai/worktrees/<project>/` |
| **Harnesses** | Generated dynamic agent teams | `.moai/harness/manifest.json` |
| **SPEC Progress** | Progress tracking for active SPECs | `.moai/specs/SPEC-*/progress.md` |

## Command Format

```bash
moai inventory [options]
```

### Basic Usage

```bash
moai inventory
```

Display inventory in basic text format.

### JSON Format Output

```bash
moai inventory --json
```

Output in structured JSON for automated analysis.

### Filtering

Query specific resource types:

```bash
moai inventory --type sessions
moai inventory --type worktrees
moai inventory --type harnesses
moai inventory --type specs
```

### Detailed Information

Include additional information for each resource:

```bash
moai inventory --verbose
moai inventory --verbose --json
```

## Text Format Output

### Basic Output Example

```
MOAI Inventory for moai-adk-go
Project Root: /path/to/your-project
Updated: 2026-07-01T10:15:00Z

========== ACTIVE SESSIONS ==========
Session ID                              Branch        SPEC ID            Status
edc25996-04cb-4139-b2f6-c2968e7337db    main          SPEC-DOCS-001      in-progress
a1b2c3d4-e5f6-7890-1234-567890abcdef    feat/auth     SPEC-AUTH-002      run-phase

========== WORKTREES ==========
Name                    Branch              Created        Status
SPEC-DOCS-001          docs/rebuild        2026-07-01     active
SPEC-AUTH-002          feat/auth            2026-07-01     active

========== HARNESSES ==========
Name                    Version    Teammates    Worktree Isolation    Status
backend-team            1.0.0      3            L1_optional           active
frontend-team           1.0.0      2            none                  active

========== ACTIVE SPECS ==========
SPEC ID                 Status          Phase      Owner           Progress
SPEC-DOCS-001          in-progress     run        manager-develop  M3/6
SPEC-AUTH-002          in-progress     run        manager-develop  M2/5
```

### Detailed Information (`--verbose`)

```
========== ACTIVE SESSIONS (VERBOSE) ==========

Session: edc25996-04cb-4139-b2f6-c2968e7337db
  Created:     2026-06-29T14:30:00Z
  Last Update: 2026-07-01T10:15:00Z
  Branch:      main
  SPEC ID:     SPEC-DOCS-001
  Status:      in-progress (running M3)
  Context:     ~145K / 200K tokens (73%)
  Model:       claude-haiku-4-5
  Resume:      available (.moai/specs/SPEC-DOCS-001/progress.md)

========== WORKTREES (VERBOSE) ==========

Worktree: SPEC-DOCS-001
  Path:         ~/.moai/worktrees/moai-adk-go/SPEC-DOCS-001
  Base Branch:  main (origin/main)
  Created:      2026-07-01T08:00:00Z
  Session:      edc25996-04cb-4139-b2f6-c2968e7337db
  Files Modified: 7
  Files Created:  4
  Commits:       2
```

## JSON Format Output

### Schema

```json
{
  "inventory": {
    "project_root": "/path/to/your-project",
    "timestamp": "2026-07-01T10:15:00Z",
    "sessions": [...],
    "worktrees": [...],
    "harnesses": [...],
    "specs": [...]
  }
}
```

### Session Object

```json
{
  "session_id": "edc25996-04cb-4139-b2f6-c2968e7337db",
  "created_at": "2026-06-29T14:30:00Z",
  "branch": "main",
  "spec_id": "SPEC-DOCS-001",
  "status": "in-progress",
  "context_usage": {
    "current": 145000,
    "total": 200000,
    "percentage": 72.5
  },
  "model": "claude-haiku-4-5",
  "resume_available": true
}
```

### Worktree Object

```json
{
  "name": "SPEC-DOCS-001",
  "path": "~/.moai/worktrees/moai-adk-go/SPEC-DOCS-001",
  "base_branch": "main",
  "created_at": "2026-07-01T08:00:00Z",
  "session_id": "edc25996-04cb-4139-b2f6-c2968e7337db",
  "status": "active",
  "files_modified": 7,
  "files_created": 4,
  "commits": 2
}
```

### Harness Object

```json
{
  "name": "backend-team",
  "version": "1.0.0",
  "created_at": "2026-07-01T10:00:00Z",
  "teammates": 3,
  "worktree_isolation": "L1_optional",
  "status": "active",
  "manifest_path": ".moai/harness/manifest.json"
}
```

### SPEC Object

```json
{
  "spec_id": "SPEC-DOCS-001",
  "title": "Documentation v3 Rebuild",
  "status": "in-progress",
  "phase": "run",
  "current_milestone": 3,
  "total_milestones": 6,
  "owner": "manager-develop",
  "progress_file": ".moai/specs/SPEC-DOCS-001/progress.md",
  "created_at": "2026-06-20T09:00:00Z"
}
```

## Practical Usage Examples

### 1. Detect Multi-Session Race Conditions

```bash
moai inventory --type sessions

# Look for multiple sessions handling the same SPEC → race condition risk
```

### 2. Review Worktree Cleanup Status

```bash
moai inventory --type worktrees --verbose

# Check for stale worktrees, then clean them
moai worktree remove <name>
```

### 3. Query Harness Team List

```bash
moai inventory --type harnesses --json | jq '.inventory.harnesses[] | {name, teammates, status}'

# Example output:
# {
#   "name": "backend-team",
#   "teammates": 3,
#   "status": "active"
# }
```

### 4. Track Active SPEC Progress

```bash
moai inventory --type specs | grep in-progress

# See all currently in-progress SPECs
```

### 5. Use in Automation Scripts

```bash
#!/bin/bash
# Auto-cleanup stale worktrees

moai inventory --type worktrees --json | jq -r '.inventory.worktrees[] | select(.status == "stale") | .name' | while read name; do
  echo "Removing stale worktree: $name"
  moai worktree remove "$name"
done
```

## Output Interpretation

### Status Field

| Status | Meaning |
|--------|---------|
| `active` | Currently in use |
| `idle` | Paused (session in explicit pause state) |
| `stale` | Unused (no access for 7+ days) |
| `error` | Error state (requires investigation) |

### Phase Field

| Phase | Description |
|-------|-------------|
| `plan` | Plan phase execution in progress |
| `run` | Run phase execution in progress |
| `sync` | Sync phase execution in progress |
| `completed` | Completed state |

## Related Documentation

- [SPEC-Based Development](/workflow-commands/moai-plan) — SPEC lifecycle
- [Worktree Management](/getting-started/worktree) — Worktree isolation and lifecycle
- [Harness v4 Builder](/advanced/builder-agents) — Dynamic team management
- [CLI Reference](/getting-started/cli) — Other CLI commands

{{< callout type="info" >}}
**Tip**: `moai inventory` can power automated cleanup scripts and monitoring dashboards. Automated analysis of JSON format keeps your project state always visible.
{{< /callout >}}
