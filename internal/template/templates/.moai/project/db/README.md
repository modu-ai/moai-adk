# .moai/project/db/

_TBD — Initialize this folder via `/moai db init` before running any database documentation tasks._

---

## Purpose

This directory is the **authoritative source** for database schema documentation, migration history,
access-control policies, query patterns, and seed data strategy for this project.

Files are maintained in two ways:
1. **Auto-sync**: The `moai-domain-db-docs` PostToolUse hook detects migration file changes and
   regenerates schema documentation automatically (10-second debounce).
2. **Manual edit**: Any file in this directory can be edited directly. Re-running `/moai db init`
   will **preserve** your edits and only warn — it will not overwrite user-modified files
   (see Auto-sync Policy below and SPEC-DB-CMD-001 for enforcement details).

---

## Auto-sync Policy

| Trigger | Action |
|---------|--------|
| Migration file saved (matches `migration_patterns` in `db.yaml`) | `moai-domain-db-docs` regenerates `schema.md` and `erd.mmd` |
| Files in `.moai/project/db/**` saved | **Excluded** — no recursive trigger |
| Files in `.moai/cache/**` saved | Excluded |
| `**/*.lock` files saved | Excluded |

Debounce: 10 seconds (configurable via `db.auto_sync.debounce_seconds` in `db.yaml`).
User approval required before applying auto-generated changes (`require_user_approval: true`).

---

## Update Workflow

```
1. Edit migration file (e.g., prisma/schema.prisma)
2. PostToolUse hook fires → moai-domain-db-docs analyzes changes
3. Proposed updates presented via AskUserQuestion
4. On approval: schema.md and erd.mmd are updated
5. Manual review: rls-policies.md, queries.md, seed-data.md (not auto-updated)
```

For conflicts (e.g., you edited `schema.md` manually and a migration also changed the schema),
MoAI calls `AskUserQuestion` to resolve the conflict before writing.

---

## File Responsibilities

| File | Purpose | Auto-updated? |
|------|---------|---------------|
| `schema.md` | Tables/collections, relationships, indexes, constraints | Yes (via hook) |
| `erd.mmd` | Mermaid ER diagram — visual representation of schema | Yes (via hook) |
| `migrations.md` | Applied and pending migration history, rollback notes | Partial (applied list) |
| `rls-policies.md` | Row-level security policies, access control matrix | No — edit manually |
| `queries.md` | Common queries, aggregations, reports | No — edit manually |
| `seed-data.md` | Seed strategy, fixture locations, dev vs prod data | No — edit manually |

---

## Excluded Patterns

The following paths are excluded from auto-sync triggering (see `db.yaml`):

```
.moai/project/db/**    # This directory — prevents recursive hook loops
.moai/cache/**         # Cache files
**/*.lock              # Lock files (package-lock.json, yarn.lock, etc.)
```

---

## Configuration

Database documentation behavior is controlled by `.moai/config/sections/db.yaml`.

Key settings:
- `db.enabled` — Must be `true` for auto-sync to activate (set by `/moai db init`)
- `db.engine` — Primary database engine (e.g., PostgreSQL, MySQL, MongoDB)
- `db.orm` — ORM/ODM in use (e.g., Prisma, SQLAlchemy, Mongoose)
- `db.migration_tool` — Migration tool (e.g., Prisma Migrate, Alembic, Rails)

---

_Last reviewed: _TBD__
_Populated by: `/moai db init` interview_
