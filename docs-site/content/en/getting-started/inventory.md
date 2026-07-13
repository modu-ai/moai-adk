---
title: moai inventory Command
weight: 25
draft: false
---

A guide to the `moai inventory` command, which queries your project's active sessions, worktrees, and harnesses.

{{< callout type="info" >}}
**One-line summary**: `moai inventory` gives you an at-a-glance view of every active resource in the current project (sessions, worktrees, harnesses).
{{< /callout >}}

## Overview

`moai inventory` is a read-only command that provides a **unified inventory** of the current project state. When you run several parallel sessions and worktrees, you need one place to answer "what is running right now?" — this command is that answer.

### What It Queries

| Resource | Description | Location |
|------|------|------|
| **Active Sessions** | Currently running Claude Code sessions | `.moai/state/active-sessions.json` |
| **Worktrees** | L2/L3 isolated branches for the project | `~/.moai/worktrees/<project>/` |
| **Harnesses** | Generated dynamic agent teams | `.moai/harness/manifest.json` |
| **SPEC Progress** | Progress state of active SPECs | `.moai/specs/SPEC-*/progress.md` |

## Command Format

```bash
moai inventory [options]
```

### Basic Usage

```bash
moai inventory
```

Prints the inventory in the default text format.

### JSON Output

```bash
moai inventory --json
```

Outputs structured JSON, useful for automated analysis.

### Filtering

Query only a specific resource type:

```bash
moai inventory --type sessions
moai inventory --type worktrees
moai inventory --type harnesses
moai inventory --type specs
```

### Verbose Details

Include extra information for each resource:

```bash
moai inventory --verbose
moai inventory --verbose --json
```

## Text Output

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

### Verbose Details (`--verbose`)

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

## JSON Output

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

### 1. Detecting Multi-Session Contention

```bash
moai inventory --type sessions

# In the output, more than 1 session working the same SPEC → contention risk
```

### 2. Checking Worktrees for Cleanup

```bash
moai inventory --type worktrees --verbose

# Identify old worktrees, then clean up
moai worktree remove <name>
```

### 3. Listing Harness Teams

```bash
moai inventory --type harnesses --json | jq '.inventory.harnesses[] | {name, teammates, status}'

# Expected output:
# {
#   "name": "backend-team",
#   "teammates": 3,
#   "status": "active"
# }
```

### 4. Tracking Active SPEC Progress

```bash
moai inventory --type specs | grep in-progress

# See every SPEC currently in progress
```

### 5. Using It in Automation Scripts

```bash
#!/bin/bash
# Automatic worktree cleanup script

moai inventory --type worktrees --json | jq -r '.inventory.worktrees[] | select(.status == "stale") | .name' | while read name; do
  echo "Removing stale worktree: $name"
  moai worktree remove "$name"
done
```

## Interpreting the Output

### The Status Field

| Status | Meaning |
|--------|------|
| `active` | Currently in use |
| `idle` | Suspended (the session is explicitly paused) |
| `stale` | Unused (no access for 7+ days) |
| `error` | Error state (needs attention) |

### The Phase Field

| Phase | Description |
|-------|------|
| `plan` | Plan phase in progress |
| `run` | Run phase in progress |
| `sync` | Sync phase in progress |
| `completed` | Completed |

## Related Documents

- [SPEC-Based Development](/en/workflow-commands/moai-plan) - The SPEC lifecycle
- [Worktree Management](/en/getting-started/worktree) - Worktree isolation and lifecycle
- [Harness v4 Builder](/en/advanced/builder-agents) - Dynamic team management
- [CLI Reference](/en/getting-started/cli) - Other CLI commands

{{< callout type="info" >}}
**Tip**: `moai inventory` works well with automated cleanup scripts and monitoring dashboards. Parse the JSON output automatically and you always know the state of your project.
{{< /callout >}}
