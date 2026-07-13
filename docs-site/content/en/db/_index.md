---
title: Database Schema Management
description: A workflow that automatically tracks and manages schemas, migrations, and seed data
weight: 70
draft: false
---

MoAI-ADK's database workflow manages your project's schema metadata centrally.
The `/moai db` command scans migration files, auto-generates schema
documentation, and detects drift between the documentation and the actual
migrations.

This schema documentation is not just for humans. From the **agentic harness**
perspective, `.moai/project/db/` is the agent's persistent context (file-based
memory). Instead of re-reading migration files every session to reconstruct
the schema, the agent references one curated schema document — dramatically
reducing the tokens needed to obtain the same information. It is a case of
Tokenomics applied all the way down to the documentation structure.

## Key features

- **Interactive initialization** — choose your database engine, ORM, and migration tool with `/moai db init` and auto-generate the metadata templates
- **Automatic sync** — a PostToolUse hook detects migration file changes and refreshes automatically
- **Drift detection** — check for mismatches between the schema documentation and migration files with `/moai db verify`
- **16 supported languages** — Go, Python, TypeScript, JavaScript, Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift

## 4 subcommands

```bash
/moai db init      # Initialize DB metadata through an interactive interview
/moai db refresh   # Rescan migration files and regenerate schema docs
/moai db verify    # Check for drift (read-only)
/moai db list      # Show all tables as a Markdown table
```

## When to use it

- Set up database metadata when starting a new project
- Auto-update documentation after adding/editing migration files
- Share the current schema state with teammates
- Verify consistency between the schema documentation and the actual migration state

## Next steps

- **[Getting Started](./getting-started.md)** — running `/moai db init` and your first migration
- **[Schema Sync](./schema-sync.md)** — the PostToolUse hook and the automatic refresh mechanism
- **[Migration Patterns](./migration-patterns.md)** — default migration paths for 16 languages
- **[Project DB Directory](./project-db-directory.md)** — introducing the 7-file template set

## Related documentation

For more details, see the [/moai db command guide](/en/db/getting-started).
