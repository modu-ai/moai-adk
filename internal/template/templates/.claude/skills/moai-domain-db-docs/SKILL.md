---
name: moai-domain-db-docs
description: >
  Parses DB migration files (Prisma, Alembic, Rails, raw SQL) and keeps
  .moai/project/db/schema.md, erd.mmd, migrations.md in sync. Powers the
  PostToolUse hook and /moai db refresh/verify subcommands.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, TaskCreate, TaskUpdate
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-04-20"
  tags: "db, schema, migration, documentation, sync, drift"

triggers:
  keywords: ["db docs", "schema sync", "migration parse", "erd update"]
  agents: ["expert-backend"]
  phases: ["run", "sync"]
---

# moai-domain-db-docs: DB Documentation Sync Skill

Purpose: Consume `.moai/cache/db-sync/proposal.json` and update the three DB documentation
files (`schema.md`, `erd.mmd`, `migrations.md`) in `.moai/project/db/`.

SPEC: SPEC-DB-SYNC-001

---

## Invocation Modes

This skill is invoked in two ways:

1. **PostToolUse approval flow** (via orchestrator): After the user selects "Apply" in the
   3-option AskUserQuestion dialog, the orchestrator invokes this skill with proposal.json.

2. **Direct subcommand**: Via `/moai db refresh` (full rebuild) or `/moai db verify` (drift check).

---

## Recursion Guard (REQ-019)

Before writing any file, check that the target path is NOT in the Excluded Patterns:

- `.moai/project/db/**`
- `.moai/cache/**`
- `.moai/logs/**`

If the target matches an excluded pattern, abort silently. This is a skill-level guard
complementing the hook-level guard (REQ-004).

---

## Phase A: Apply Proposal (REQ-013, REQ-016, REQ-017, REQ-018)

Triggered when user selects "Apply" in the approval flow.

### A1: Read proposal.json

```
Read .moai/cache/db-sync/proposal.json
```

Extract:
- `file_path`: original migration file
- `parsed_content`: normalized schema representation
- `timestamp`: ISO-8601 creation time

### A2: Update schema.md (REQ-016)

Rules for in-place update:
1. Read existing `.moai/project/db/schema.md`.
2. Identify sections delimited by `## ` headers.
3. For each table/collection referenced in `parsed_content`:
   - If a matching row exists, update column_count and last_migration_file.
   - If no matching row, append a new row preserving `_TBD_` in user-managed columns.
4. Do NOT overwrite user-edited descriptions or `_TBD_` markers.
5. Write updated file using Edit tool.

### A3: Regenerate erd.mmd (REQ-017)

Rules:
1. Read existing `.moai/project/db/erd.mmd` to extract the comment header block (lines starting with `%%`).
2. Rebuild the `erDiagram` body from `parsed_content` table definitions.
3. Preserve the original comment header verbatim.
4. Validate Mermaid syntax: must start with `erDiagram` after the comment header.
5. Write using Edit tool (or Write if file is new).

Template output format:

```
%% MoAI DB ERD — auto-generated. Edit schema.md to add relationships.
%% Last updated: <ISO-8601 timestamp>
erDiagram
    <TABLE_NAME> {
        <type> <column_name>
    }
```

### A4: Append to migrations.md (REQ-018)

1. Read existing `.moai/project/db/migrations.md`.
2. Locate the `## Applied Migrations` table.
3. Append a new row with:
   - `filename`: basename of `file_path`
   - `applied_at`: ISO-8601 timestamp from proposal
   - `checksum`: SHA-256 of the migration file content (compute via `sha256sum` or `shasum -a 256`)
   - `up_summary`: first non-comment line of `parsed_content` (truncated to 80 chars)
4. Write using Edit tool.

---

## Phase B: User Approval Flow (REQ-012, REQ-014, REQ-015)

This phase documents the orchestrator-level behavior that invokes this skill.
The orchestrator (not this skill) calls AskUserQuestion.

### B1: 3-Option AskUserQuestion

When `proposal.json` is present and `decision == "ask-user"`, the orchestrator presents:

Question: "Migration file changed. Update schema documentation?"

Options:
1. Apply (권장) — Apply proposed schema update: invoke moai-domain-db-docs to update
   schema.md, erd.mmd, and migrations.md from the parsed migration file.
2. Review — Review diff first: display the diff between current schema.md and proposed
   changes, then re-ask Apply/Skip.
3. Skip — Skip this time: delete proposal.json and take no action.

### B2: Review diff (REQ-014)

If user selects "Review":
1. Read current `.moai/project/db/schema.md`.
2. Generate a unified diff between current content and proposed content from `parsed_content`.
3. Display diff (truncate to 100 lines; add note "see proposal.json for full diff" if truncated).
4. Re-ask with 2 options: Apply / Skip.

### B3: Skip (REQ-015)

If user selects "Skip":
1. Delete `.moai/cache/db-sync/proposal.json`.
2. Output: "Schema update skipped."

---

## Phase C: /moai db verify (REQ-020, REQ-021, REQ-022)

Triggered by `/moai db verify`. Read-only — MUST NOT modify any files.

### C1: Compute expected schema

1. Scan migration files using patterns from `db.yaml` `migration_patterns`.
2. Extract table names from each migration file (stub: grep for `CREATE TABLE`).
3. Build expected set E of table names.

### C2: Read current schema.md

1. Read `.moai/project/db/schema.md`.
2. Extract registered table names from the `## Tables` section.
3. Build documented set D of table names.

### C3: Diff and output

Compute symmetric difference: `drift = (E - D) ∪ (D - E)`

If no drift:
- Print: `Schema documentation is in sync`
- Exit 0 (REQ-022)

If drift detected:
- Print unified diff to stdout showing added/removed tables (REQ-021)
- Exit 1 (REQ-021)

Format:
```
--- schema.md (documented)
+++ migrations (actual)
@@ -1,N +1,M @@
+ <table added in migrations, missing from schema.md>
- <table in schema.md, missing from migrations>
```

---

## Phase D: /moai db refresh (REQ-023, REQ-024)

Triggered by `/moai db refresh`. Full rebuild from all migration files.

### D1: User confirmation (REQ-024)

Call AskUserQuestion:

Question: "Confirm full rebuild of schema.md, erd.mmd, and migrations.md?"

Options:
1. Apply (권장) — Rebuild all 3 docs from scratch by rescanning all migration files.
2. Cancel — Cancel the rebuild. No files will be modified.

If Cancel: exit 0 with message "Refresh cancelled."

### D2: Full scan

1. Use Glob to find all migration files matching patterns from `db.yaml`.
2. Parse each file using the stub parser (read content).
3. Aggregate all table definitions.

### D3: Rebuild 3 docs

Rebuild in order:
1. `schema.md` — regenerate table registry from aggregated tables.
2. `erd.mmd` — regenerate erDiagram from aggregated tables.
3. `migrations.md` — regenerate Applied Migrations table with all found files.

Use Write tool (not Edit) since this is a full rebuild.

Output summary:
```
Rebuilt schema.md: <N> tables
Rebuilt erd.mmd: <N> entities
Rebuilt migrations.md: <N> migration entries
```

---

## Excluded Patterns Reference

Do NOT write to or process these paths (recursion guard, REQ-019):
- `.moai/project/db/**`
- `.moai/cache/**`
- `.moai/logs/**`

---

## Error Handling

- If any file read/write fails: log to `.moai/logs/db-sync-errors.log` and continue with remaining files.
- If SHA-256 computation fails: use `unknown` as checksum value.
- If `erDiagram` validation fails: write a minimal valid `erDiagram {}` and log the error.
- Never block the user's workflow (always exit 0 on non-fatal errors).
