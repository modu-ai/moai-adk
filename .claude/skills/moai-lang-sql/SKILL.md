---

name: moai-lang-sql
description: SQL best practices with testing frameworks, query optimization, and migration management. Use when writing or reviewing SQL code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# SQL Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | SQL code discussions, framework guidance, or file extensions such as .sql. |
| Tier | 3 |

## What it does

Provides SQL-specific expertise for database development, including SQL testing strategies, query optimization techniques, and migration management best practices.

## When to use

- Engages when the conversation references SQL work, frameworks, or files like .sql.
- “Writing SQL tests”, “Query optimization”, “Migration management”
- Automatically invoked when working with database projects
- SQL SPEC implementation (`/alfred:2-run`)

## How it works

**Testing Strategies**:
- **pgTAP**: PostgreSQL testing framework
- **DbUnit**: Database testing for JVM
- **SQLTest**: SQL unit testing
- Integration tests with test databases

**Query Optimization**:
- **EXPLAIN/EXPLAIN ANALYZE**: Execution plan analysis
- **Index optimization**: B-tree, Hash, GiST indices
- **Query rewriting**: JOIN optimization
- **Avoiding N+1 queries**: Eager loading

**Migration Management**:
- **Flyway**: Version-based migrations
- **Liquibase**: Changelog-based migrations
- **Alembic**: Python database migrations
- **Rails migrations**: Ruby on Rails approach

**SQL Best Practices**:
- **Normalization**: 3NF compliance
- **Constraints**: Foreign keys, NOT NULL, CHECK
- **Transactions**: ACID properties
- **Prepared statements**: SQL injection prevention

**Database Patterns**:
- Use CTEs (Common Table Expressions) for readability
- Window functions over self-joins
- Avoid SELECT * in production code
- Use parameterized queries

## Examples
```bash
sqlfluff lint sql/ && sqlfluff fix sql/
```

## Inputs
- Language-specific source directories (e.g. `src/`, `app/`).
- Language-specific build/test configuration files (e.g. `package.json`, `pyproject.toml`, `go.mod`).
- Relevant test suites and sample data.

## Outputs
- Test/lint execution plan tailored to the selected language.
- List of key language idioms and review checkpoints.

## Failure Modes
- When the language runtime or package manager is not installed.
- When the main language cannot be determined in a multilingual project.

## Dependencies
- Access to the project file is required using the Read/Grep tool.
- When used with `Skill("moai-foundation-langs")`, it is easy to share cross-language conventions.

## References
- PostgreSQL Global Development Group. "PostgreSQL Documentation." https://www.postgresql.org/docs/ (accessed 2025-03-29).
- SQLFluff. "SQLFluff Documentation." https://docs.sqlfluff.com/en/stable/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (migration validation)
- alfred-code-reviewer (SQL review)
- database-expert (database design)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
