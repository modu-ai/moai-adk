---
id: SPEC-STATUS-AUTO-001
version: "1.0.0"
status: draft
created: "2026-04-27"
updated: "2026-04-27"
author: GOOS
priority: P1
labels: [spec-status, automation, hooks, cli, sync-workflow]
---

# SPEC-STATUS-AUTO-001: SPEC Status Auto-Update System

## Problem Statement

SPEC documents in `.moai/specs/` track lifecycle status (`draft` → `planned` → `implemented` → `completed`), but status updates are purely manual. When a SPEC's implementation is merged to main, the SPEC frontmatter is often left at `draft` or `planned` indefinitely. 15 SPECs were found in this exact state during audit on 2026-04-27.

Root causes:
1. No code exists to modify SPEC frontmatter status
2. `/moai sync` Step 2.4 documents status updates but has no implementation
3. SPEC metadata formats are fragmented (6 variants: YAML, Markdown list, table, Korean fields)
4. No trigger mechanism to detect when a SPEC should transition

## Requirements (EARS Format)

### REQ-1: SPEC Status Updater Library (Priority: P0)

**WHEN** `spec.UpdateStatus(specID, newStatus)` is called, **THE SYSTEM SHALL** locate `.moai/specs/<specID>/spec.md`, detect its metadata format (YAML frontmatter, Markdown list, or table), update the status field to `newStatus`, and preserve all other content unchanged.

**Acceptance Criteria:**
- AC-1.1: Supports all 6 format variants:
  - YAML `status:` (Format A, B)
  - Markdown `- **Status**: X` (Format D)
  - Markdown table `| 상태 | X |` or `| Status | X |` (Format E)
  - Adds YAML frontmatter block if none exists (Format F)
- AC-1.2: Validated status values: `draft`, `planned`, `in-progress`, `implemented`, `completed`, `superseded`
- AC-1.3: Returns error if spec.md file not found or status value invalid
- AC-1.4: Does not modify any content outside the status field
- AC-1.5: Unit test coverage >= 90% for all format parsers

### REQ-2: CLI Command (Priority: P0)

**WHEN** `moai spec status <SPEC-ID> <status>` is invoked, **THE SYSTEM SHALL** call the updater library and report success or failure.

**Acceptance Criteria:**
- AC-2.1: Command registered under `moai spec` parent command
- AC-2.2: Validates SPEC-ID exists in `.moai/specs/` before attempting update
- AC-2.3: Prints human-readable confirmation: `SPEC-XXX status updated: draft → completed`
- AC-2.4: Supports `--dry-run` flag to preview change without writing
- AC-2.5: Supports `--list` flag to show all SPECs and their current status

### REQ-3: Commit-Based Auto-Detection (L1) (Priority: P1)

**WHEN** a git commit message contains a pattern matching `SPEC-[A-Z0-9]+-[0-9]+`, **THE SYSTEM SHALL** extract the SPEC-ID(s) and automatically update their status to `implemented`.

**Acceptance Criteria:**
- AC-3.1: Implemented as `internal/hook/spec_status.go` hook handler
- AC-3.2: Triggered by `PostToolUse` event when tool is `Bash` and command matches `git commit`
- AC-3.3: Parses commit message from `git log -1 --format=%s`
- AC-3.4: Extracts all unique SPEC-XXX patterns from the message
- AC-3.5: For each SPEC-ID found, calls `spec.UpdateStatus(specID, "implemented")`
- AC-3.6: Logs updated SPECs to stderr (non-blocking — hook exit code always 0)
- AC-3.7: Gracefully skips if `.moai/specs/` directory does not exist

### REQ-4: Sync Workflow Integration (L2) (Priority: P1)

**WHEN** `/moai sync` completes documentation synchronization, **THE SYSTEM SHALL** update the synced SPEC status to `completed`.

**Acceptance Criteria:**
- AC-4.1: sync.md Phase 2.4 invokes `moai spec status <SPEC-ID> completed`
- AC-4.2: Status update occurs after documentation sync succeeds
- AC-4.3: Failure to update status does not block the sync workflow (warning only)
- AC-4.4: Logs the status transition in sync output

### REQ-5: Batch Status Command (Priority: P2)

**WHEN** `moai spec status --sync-git` is invoked, **THE SYSTEM SHALL** cross-reference all SPECs in `.moai/specs/` against git log on main and update statuses where implementation commits exist.

**Acceptance Criteria:**
- AC-5.1: Scans `git log main --oneline --no-merges` for SPEC-XXX patterns
- AC-5.2: For each SPEC found in commits but not marked `completed`/`implemented`, updates status
- AC-5.3: Reports a summary: `Updated N SPECs, skipped M (already completed), K not found`
- AC-5.4: Requires `--confirm` flag (or `--yes` for non-interactive) before writing changes

## Technical Approach

### Package Structure

```
internal/spec/
  status.go         # UpdateStatus(), ParseStatus(), Format detection
  status_test.go    # Unit tests for all 6 format parsers

internal/hook/
  spec_status.go    # L1 hook handler (PostToolUse)

internal/cli/
  spec.go           # `moai spec` parent command
  spec_status.go    # `moai spec status` subcommand
```

### Format Detection Algorithm

```
1. Read first 30 lines of spec.md
2. If contains "---" delimiter → YAML frontmatter
   a. Parse YAML, look for "status:" key
   b. If no "status:" key → add it
3. Else if contains "| 상태 |" or "| Status |" → Table format
   a. Replace status value in table row
4. Else if contains "- **Status**:" → Markdown list format
   a. Replace status value in list item
5. Else → Prepend YAML frontmatter block with status field
```

### Hook Registration

In `settings.json` hooks:
```json
{
  "PostToolUse": [{
    "matcher": "Bash",
    "hooks": [{
      "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-spec-status.sh\"",
      "timeout": 5
    }]
  }]
}
```

Hook wrapper script reads stdin JSON, checks if bash command was `git commit`, then calls `moai spec status --auto-commit`.

## Scope

### In Scope
- SPEC status updater library with multi-format support
- CLI command `moai spec status`
- L1 PostToolUse hook for commit-based detection
- L2 sync workflow integration
- Batch sync-git command

### Out of Scope
- L3 PR merge detection via GitHub Actions (future SPEC)
- SPEC lifecycle state machine enforcement (validation only)
- Web UI for SPEC status management
- Automatic archiving of completed SPECs

## Dependencies

- Existing `internal/cli/` command structure (cobra)
- Existing `internal/hook/` hook handler pattern
- `.moai/specs/` directory structure (convention)

## Risks

| Risk | Mitigation |
|------|------------|
| YAML frontmatter parsing fragile | Use simple line-based regex, not full YAML parser |
| Hook timeout on large commit messages | 5s timeout, process only first line |
| Race condition on concurrent edits | SPEC files are single-writer (one session at a time) |
| False positive SPEC pattern in commit body | Only match conventional commit title (first line) |
