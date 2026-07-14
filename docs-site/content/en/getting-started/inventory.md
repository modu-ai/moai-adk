---
title: moai inventory Command
weight: 25
draft: false
---

A guide to the `moai inventory` command, which shows the current project's active sessions, worktrees, and harnesses at a glance.

{{< callout type="info" >}}
**One-line summary**: `moai inventory` shows the current project's active resources (sessions, worktrees, harnesses) read-only. With `--json` you get structured output for use in scripts.
{{< /callout >}}

## Overview

`moai inventory` is a read-only command that provides a unified view to check "what is running right now?" at once when you operate multiple parallel sessions and worktrees.

### What it shows

| Resource | Description | Data source |
|------|------|------------|
| **Sessions** | Active Claude Code sessions | `.moai/state/active-sessions.json` |
| **Worktrees** | Git worktrees for the project | Git worktree list |
| **Harnesses** | Registered harnesses | `.moai/harness/` manifests |

## Command form

```bash
moai inventory [OPTIONS]
```

### Flags

| Flag | Description |
|------|------|
| `--json` | Structured JSON output (machine-readable) |
| `--project-root <path>` | Project root path (default: current directory) |

This command supports only the two flags above. There are no filtering or verbose-mode flags — do any needed processing on the `--json` output with `jq` or similar.

## Basic use

```bash
moai inventory
```

Prints a text-format summary of sessions, worktrees, and harnesses.

## JSON output

```bash
moai inventory --json
```

Outputs structured JSON for use in automated analysis or CI scripts.

## JSON schema

The top-level structure of the `--json` output consists of three sections.

```json
{
  "sessions": { ... },
  "worktrees": { ... },
  "harnesses": { ... }
}
```

Each section has `count`, `entries`, and an optional `error` field.

### Session entry

```json
{
  "session_id": "edc25996",
  "spec_id": "SPEC-DOCS-001",
  "phase": "run"
}
```

| Field | Description |
|------|------|
| `session_id` | Session ID (short form, first 8 characters) |
| `spec_id` | Linked SPEC ID |
| `phase` | Current phase (`plan`, `run`, `sync`, `mx`) |

### Worktree entry

```json
{
  "branch": "feat/auth",
  "path": "/home/user/.moai/worktrees/project/SPEC-AUTH-001",
  "head": "a1b2c3d4"
}
```

| Field | Description |
|------|------|
| `branch` | Worktree branch name |
| `path` | Worktree filesystem path |
| `head` | HEAD commit hash (short form, first 8 characters) |

### Harness entry

```json
{
  "name": "backend-team",
  "domain": "backend",
  "manifest_missing": false
}
```

| Field | Description |
|------|------|
| `name` | Harness name |
| `domain` | Harness domain |
| `manifest_missing` | Whether the manifest file is missing (`true` means the configuration is incomplete) |

### Full output example

```json
{
  "sessions": {
    "count": 2,
    "entries": [
      { "session_id": "edc25996", "spec_id": "SPEC-DOCS-001", "phase": "run" },
      { "session_id": "a1b2c3d4", "spec_id": "SPEC-AUTH-002", "phase": "plan" }
    ]
  },
  "worktrees": {
    "count": 1,
    "entries": [
      { "branch": "feat/auth", "path": "/home/user/.moai/worktrees/project/SPEC-AUTH-001", "head": "a1b2c3d4" }
    ]
  },
  "harnesses": {
    "count": 1,
    "entries": [
      { "name": "backend-team", "domain": "backend", "manifest_missing": false }
    ]
  }
}
```

## Practical usage examples

### 1. Detect multi-session contention

If two or more sessions are working on the same SPEC, there is a contention risk.

```bash
moai inventory --json | jq '[.sessions.entries[] | .spec_id] | group_by(.) | map(select(length > 1))'
```

### 2. List active worktree branches

```bash
moai inventory --json | jq -r '.worktrees.entries[].branch'
```

### 3. Find harnesses with a missing manifest

A harness with `manifest_missing: true` is in an incomplete configuration state.

```bash
moai inventory --json | jq '.harnesses.entries[] | select(.manifest_missing)'
```

### 4. Distribution of currently in-progress phases

```bash
moai inventory --json | jq '[.sessions.entries[].phase] | group_by(.) | map({phase: .[0], count: length})'
```

## Related documents

- [CLI Reference](./cli) — full CLI commands
- [Project Status](./status) — the `moai status` command
- [SPEC-based Development](/en/workflow-commands/moai-plan) — the SPEC lifecycle

{{< callout type="info" >}}
**Tip**: `moai inventory --json` can be used in monitoring dashboards and CI scripts. Since it is a read-only command, it is safe to automate.
{{< /callout >}}
