---
name: moai-workflow-db
description: >
  Database metadata workflow for MoAI projects. Manages schema, migrations, and seeds
  through four subcommands: init (interactive setup), refresh (rescan migrations),
  verify (drift detection), and list (read-only table output). Thin Wrapper entry
  point delegates to this skill via Skill("moai") with arguments: db $ARGUMENTS.
user-invocable: false
allowed-tools: AskUserQuestion, Read, Write, Edit, Bash, TaskCreate, TaskUpdate, Glob, Grep
metadata:
  version: "0.1.0"
  category: "workflow"
  status: "active"
  updated: "2026-04-21"
  tags: "db, database, schema, migrations, init, refresh, verify, list"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["db", "database", "schema", "migration", "migrations", "seeds"]
  agents: ["manager-spec", "expert-backend"]
  phases: ["run"]
---

# Workflow: db — Database Metadata Management

Purpose: Manage project database metadata (schema, migrations, seeds) through four subcommands.
This Thin Wrapper router dispatches `$ARGUMENTS[0]` to the appropriate Phase below.

SPEC: SPEC-DB-CMD-001

---

## Phase 0: Router

Read `$ARGUMENTS[0]` (the first token passed to this skill).

Router Table:

| First token | Action | Target Phase(s) |
|---|---|---|
| `init` | Full interactive init | Phase 1 (Preflight) → Phase 2 (Interview) → Phase 3 (Template Render) → Phase 4 (Hook Registration) |
| `refresh` | Rescan migrations, rebuild docs | Phase 5 (Scan) → Phase 6 (Regenerate) |
| `verify` | Drift check (read-only) | Phase 7 (Drift Detection) |
| `list` | Read-only table output | Phase 8 (List Tables) |
| (empty) | AskUserQuestion select | Return to Phase 0 after selection |
| \<unknown\> | Structured error, exit | N/A — no Phase executed |

### Empty first token

If `$ARGUMENTS[0]` is empty or not provided, present subcommand selection via AskUserQuestion:

Question: "Which /moai db subcommand would you like to run?"

Options:
1. init (권장) — Initialize database metadata: run an interactive interview to set up engine, ORM, multi-tenant, and migration tool, then scaffold `.moai/project/db/` artifacts.
2. refresh — Rescan migration files and regenerate schema documentation from current migration state.
3. verify — Read-only drift check: compare schema.md table set against migration files and exit non-zero if drift detected.
4. list — Read-only table listing: display all tables from schema.md in a Markdown aligned table.

After selection, route to the appropriate Phase.

### Unknown first token

If `$ARGUMENTS[0]` is not one of `init`, `refresh`, `verify`, `list`, output a structured error and exit without executing any Phase:

```
Error: Unknown subcommand "$ARGUMENTS[0]"

Valid subcommands for /moai db:
  init     Initialize database metadata via interactive interview
  refresh  Rescan migrations and regenerate schema documentation
  verify   Check for drift between schema.md and migration files
  list     Display all tables from schema.md (read-only)

Usage: /moai db <subcommand> [args]
```

Exit with non-zero status (exit 1).

---

## Phase 1: Preflight

Triggered by: `init`

Check that required prerequisite files exist before any DB metadata is created.

Required files:
- `.moai/project/product.md`
- `.moai/project/tech.md`

Check steps:
1. Use Glob to verify `.moai/project/product.md` exists.
2. Use Glob to verify `.moai/project/tech.md` exists.
3. If either file is missing, output the following message and stop without creating any files:

```
Error: Missing prerequisite files.

Run /moai project first to generate product.md and tech.md before initializing db metadata.
```

If both files exist, proceed to Phase 2.

---

## Phase 2: Interview

Triggered by: `init` (after Phase 1 passes)

Call AskUserQuestion once with exactly 4 questions (Claude Code limit: max 4 questions per call) covering all DB initialization decisions:
1. Database engine (e.g., PostgreSQL, MySQL, SQLite, MongoDB, etc.)
2. ORM or query builder (e.g., GORM, sqlc, Prisma, SQLAlchemy, ActiveRecord, etc.)
3. Multi-tenant strategy (single schema, schema-per-tenant, database-per-tenant, or none)
4. Migration tool and conventions (e.g., golang-migrate, Flyway, Liquibase, Alembic, etc.)

After collecting answers, proceed to Phase 3.

---

## Phase 3: Template Render

Triggered by: `init` (after Phase 2 completes)

Use TaskCreate to track this phase. Update task status with TaskUpdate at each step.

Render DB metadata templates based on interview answers from Phase 2.

Steps:
1. TaskCreate: "Render DB metadata templates"
2. Determine the target directory: `.moai/project/db/`
3. Based on the engine and ORM answers, render appropriate template files:
   - `schema.md` — table/collection registry with columns and metadata
   - `migrations.md` — migration file index
   - `seeds.md` — seed data catalog
4. Write rendered files using Write tool.
5. TaskUpdate: mark complete.

NOTE: Actual template rendering logic is defined in SPEC-DB-TEMPLATES-001. This phase provides the skeleton entry point.

Proceed to Phase 4.

---

## Phase 4: Hook Registration

Triggered by: `init` (after Phase 3 completes)

Register PostToolUse hooks for ongoing schema sync.

NOTE: Hook registration implementation is defined in SPEC-DB-SYNC-001. This phase provides the skeleton entry point.

Steps:
1. Check if `PostToolUse` hook for db sync is already registered in `.claude/settings.json`.
2. If not registered, note that SPEC-DB-SYNC-001 handles this integration.
3. Output a summary of what was initialized.

---

## Phase 5: Scan

Triggered by: `refresh`

Use TaskCreate to track this phase. Update task status with TaskUpdate at each step.

Scan migration files for the current project using language-aware glob patterns.

Steps:
1. TaskCreate: "Scan migration files"
2. Read `.moai/config/sections/language.yaml` to determine the project language via `project_markers` or auto-detection.
3. Look up the canonical migration path for the detected language in the Appendix: Language Migration Path Mapping table.
4. Use Glob to scan the canonical migration path for migration files.
5. If the user has overridden the path in project config, use the override instead.
6. Collect the list of migration files with timestamps/versions.
7. TaskUpdate: mark complete with file count.

Proceed to Phase 6.

---

## Phase 6: Regenerate

Triggered by: `refresh` (after Phase 5 completes)

Regenerate schema documentation from the scanned migration files.

Implementation: Invoke Skill("moai-domain-db-docs") with `refresh` mode.
The skill handles Phase D (full rebuild) including user confirmation (REQ-024),
rescanning all migrations, and rebuilding schema.md / erd.mmd / migrations.md.

Steps (delegated to moai-domain-db-docs):
1. Skill("moai-domain-db-docs") Phase D1: AskUserQuestion confirm full rebuild.
2. Skill("moai-domain-db-docs") Phase D2: Full scan of all migration files.
3. Skill("moai-domain-db-docs") Phase D3: Rebuild 3 docs and output summary.

NOTE: Detailed regeneration logic is in moai-domain-db-docs (SPEC-DB-SYNC-001 Phase D).

---

## Phase 7: Drift Detection

Triggered by: `verify`

Read-only drift check. This phase MUST NOT create or modify any files.

Implementation: Invoke Skill("moai-domain-db-docs") with `verify` mode.
The skill handles Phase C (drift detection, REQ-020–022) and exits with the
appropriate code (0 = in sync, 1 = drift detected).

Steps (delegated to moai-domain-db-docs):
1. Skill("moai-domain-db-docs") Phase C1: Compute expected schema from migration files.
2. Skill("moai-domain-db-docs") Phase C2: Read current schema.md table set.
3. Skill("moai-domain-db-docs") Phase C3: Diff and output:
   - No drift: print "Schema documentation is in sync", exit 0 (REQ-022).
   - Drift: print unified diff, exit 1 (REQ-021).

Fallback (if moai-domain-db-docs is unavailable): apply inline logic below.

Check if `.moai/project/db/schema.md` exists using Glob.
- If missing: output `"Schema not found. Run /moai db init to initialize db schema."` and exit with code 2.

Extract table names from schema.md and migration files, compute symmetric difference.
If drift: output drift report and exit 1. If no drift: output "No drift detected." and exit 0.

---

## Phase 8: List Tables

Triggered by: `list`

Read-only table listing. This phase MUST NOT create or modify any files.

Steps:
1. Check if `.moai/project/db/schema.md` exists using Glob.
   - If missing: output `"Schema not found. Run /moai db init to initialize db schema."` and exit with code 2.
2. Read `.moai/project/db/schema.md`.
3. Extract table metadata: `table_name`, `column_count`, `primary_key`, `last_migration_file`.
4. Output a Markdown aligned table to stdout:

```
| table_name | column_count | primary_key | last_migration_file |
|---|---|---|---|
| users | 8 | id | 20240101_create_users.sql |
| orders | 12 | id | 20240115_create_orders.sql |
```

After outputting the table, exit 0. Do NOT write or modify any files.

---

## Appendix: Language Migration Path Mapping

Reference table for canonical migration tool paths per language (REQ-009).

This table is the single source of truth for default migration file locations.
Users may override these paths via project configuration (SPEC-DB-TEMPLATES-001).

| Language | Canonical Migration Tool | Default Path Pattern |
|---|---|---|
| `go` | golang-migrate | `db/migrations/*.sql` or `migrations/*.sql` |
| `python` | Alembic | `alembic/versions/*.py` |
| `typescript` | Prisma Migrate | `prisma/migrations/**/*.sql` |
| `javascript` | Knex.js | `migrations/*.js` or `knexfile migrations/` |
| `rust` | SQLx | `migrations/*.sql` |
| `java` | Flyway | `src/main/resources/db/migration/V*.sql` |
| `kotlin` | Flyway | `src/main/resources/db/migration/V*.sql` |
| `csharp` | EF Core Migrations | `Migrations/*.cs` |
| `ruby` | ActiveRecord (Rails) | `db/migrate/*.rb` |
| `php` | Laravel Migrations | `database/migrations/*.php` |
| `elixir` | Ecto | `priv/repo/migrations/*.exs` |
| `cpp` | No canonical standard | `db/migrations/*.sql` (convention) |
| `scala` | Slick / Flyway | `src/main/resources/db/migration/V*.sql` |
| `r` | No canonical standard | `migrations/*.sql` (convention) |
| `flutter` | Drift (drift_db_inspector) | `assets/migrations/*.sql` |
| `swift` | GRDB | `Resources/Migrations/*.sql` |

Detection strategy: Read `.moai/config/sections/language.yaml` for `project_markers` or auto-detect by scanning for language-specific marker files (e.g., `go.mod` for Go, `package.json` for JavaScript/TypeScript, `Cargo.toml` for Rust). Use the detected language to look up the path pattern above. If the project language is ambiguous, present options via AskUserQuestion.
