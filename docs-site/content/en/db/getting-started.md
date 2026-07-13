---
title: Getting Started
description: Initialize your project's database metadata with /moai db init
weight: 10
draft: false
---

One run of `/moai db init` puts your project's database metadata in place.
From then on, the schema documentation follows along every time you add a
migration file, so both agents and teammates always see the latest schema in
the same place.

## Prerequisites

Before starting the database workflow, you need:

1. The `.moai/project/product.md` and `.moai/project/tech.md` files created by the `/moai project` command
2. A supported database engine (PostgreSQL, MySQL, SQLite, MongoDB, etc.)
3. An ORM or query builder (GORM, sqlc, Prisma, SQLAlchemy, ActiveRecord, etc.)
4. A migration tool (golang-migrate, Flyway, Liquibase, Alembic, etc.)

## Step-by-step initialization guide

### Step 1: check the project metadata

First, verify the required files created by `/moai project` exist:

```bash
ls -la .moai/project/
# The following files must exist:
# - product.md
# - tech.md
# - structure.md
```

If these files are missing, run `/moai project` first.

### Step 2: initialize the database metadata

Now run the `/moai db init` command:

```bash
/moai db init
```

### Step 3: answer the interview questions

MoAI interactively asks about the following 4 items:

1. **Database engine** — the database you use (PostgreSQL, MySQL, SQLite, MongoDB, etc.)
2. **ORM/query builder** — your data access layer tooling
3. **Multi-tenant strategy** — single schema, schema-per-tenant, DB-per-tenant, or none
4. **Migration tool** — your schema change management tool

Choose the appropriate option for each question.

### Step 4: review the generated files

After initialization, the following files are created in the
`.moai/project/db/` directory:

```
.moai/project/db/
├── README.md              # DB section overview
├── schema.md              # Auto-generated table registry
├── erd.mmd                # Entity-relationship diagram
├── migrations.md          # Migration file index
├── rls-policies.md        # Row-level security rules (Supabase/Postgres)
├── queries.md             # Common query library
└── seed-data.md           # Seed data patterns
```

What each file does:

- `schema.md` — automatically documents all tables, columns, data types, and constraints
- `erd.mmd` — visualizes table relationships in Mermaid syntax
- `migrations.md` — a timeline of applied migration files
- `queries.md` — a collection of common query examples that AI agents reference

### Step 5: write and sync your first migration

Add a new migration file to the project. For Go/golang-migrate, for example:

```bash
# Create a migration file in the db/migrations/ directory
touch db/migrations/001_create_users_table.sql
```

After writing the migration file, refresh the schema documentation with:

```bash
/moai db refresh
```

This command:
- Scans all migration files
- Adds the new table information to schema.md
- Updates the erd.mmd diagram
- Refreshes the migrations.md timeline

### Step 6: verify drift (optional)

To check for drift:

```bash
/moai db verify
```

Results:

- `Schema documentation is in sync` — migrations and documentation match
- Drift report output — differences shown in detail (exit code: 1)

## Troubleshooting

### "Missing prerequisite files" error

If `.moai/project/product.md` and `.moai/project/tech.md` are missing:

```bash
/moai project
```

Run the command above first to generate the project metadata.

### Migration files are not recognized

Verify the project's language and migration tool were detected correctly:

```bash
cat .moai/config/sections/language.yaml
```

Check the `language` field; if needed, you can manually specify
`migration_patterns` in `.moai/config/sections/db.yaml`.

### Automatic sync is not working

Verify the PostToolUse hook is registered correctly:

```bash
grep -A5 "PostToolUse" .claude/settings.json
```

If the hook is missing, re-run `/moai db init` or register it manually in
`.claude/settings.json`.
