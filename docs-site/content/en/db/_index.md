---
title: Database Schema Management
description: Automatically track and manage schemas, migrations, and seed data
weight: 15
draft: false
---

MoAI-ADK's database workflow provides centralized management of your project's database metadata. With the `/moai db` command, you can scan migration files, automatically generate schema documentation, and detect drift between documentation and actual migrations.

## Key Features

- **Interactive Initialization** — Run `/moai db init` to select database engine, ORM, and migration tool, then automatically generate metadata templates
- **Automatic Synchronization** — PostToolUse hook automatically detects migration file changes and refreshes documentation
- **Drift Detection** — Use `/moai db verify` to detect inconsistencies between schema documentation and migration files
- **16-Language Support** — Go, Python, TypeScript, Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala, R, Flutter, Swift

## Four Subcommands

```bash
/moai db init      # Initialize DB metadata via interactive interview
/moai db refresh   # Rescan migration files and regenerate schema documentation
/moai db verify    # Check for drift (read-only)
/moai db list      # Display all tables as Markdown table
```

## When to Use

- Set up database metadata for new projects
- Automatically update documentation after adding/editing migration files
- Share current schema status with team members
- Validate consistency between schema documentation and actual migration state

## Next Steps

- **[Getting Started](./getting-started.md)** — Run `/moai db init` and create your first migration
- **[Schema Sync](./schema-sync.md)** — PostToolUse hook and automatic refresh mechanism
- **[Migration Patterns](./migration-patterns.md)** — Default migration paths for 16 languages
- **[Project DB Directory](./project-db-directory.md)** — Introduction to the 7-file template set

## Related Documentation

See the [/moai db command reference](../../reference/moai-db.md) for more details.
