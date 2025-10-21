---

name: moai-domain-database
description: Database design, schema optimization, indexing strategies, and migration management. Use when working on database integration tasks scenarios.
allowed-tools:
  - Read
  - Bash
---

# Database Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand for data layer design |
| Trigger cues | Schema modeling, migration planning, query optimization, indexing strategy. |
| Tier | 4 |

## What it does

Provides expertise in database design, schema normalization, indexing strategies, query optimization, and safe migration management for SQL and NoSQL databases.

## When to use

- Engages when the conversation focuses on database design or tuning.
- “Database design”, “Schema optimization”, “Index strategy”, “Migration”
- Automatically invoked when working with database projects
- Database SPEC implementation (`/alfred:2-run`)

## How it works

**Schema Design**:
- **Normalization**: 1NF, 2NF, 3NF, BCNF
- **Denormalization**: Performance trade-offs
- **Constraints**: Primary keys, foreign keys, unique, check
- **Data types**: Choosing appropriate types

**Indexing Strategies**:
- **B-tree indices**: General-purpose indexing
- **Hash indices**: Exact match queries
- **Full-text indices**: Text search
- **Composite indices**: Multi-column indexing
- **Index maintenance**: REINDEX, VACUUM

**Query Optimization**:
- **EXPLAIN/EXPLAIN ANALYZE**: Query plan analysis
- **JOIN optimization**: INNER, LEFT, RIGHT, FULL
- **Subquery vs JOIN**: Performance comparison
- **N+1 query problem**: Eager loading
- **Query caching**: Redis, Memcached

**Migration Management**:
- **Version control**: Flyway, Liquibase, Alembic
- **Rollback strategies**: Backward compatibility
- **Zero-downtime migrations**: Expand-contract pattern
- **Data migrations**: Safe data transformations

**Database Types**:
- **SQL**: PostgreSQL, MySQL, SQLite
- **NoSQL**: MongoDB, Redis, Cassandra
- **NewSQL**: CockroachDB, Vitess

## Examples
```bash
$ alembic upgrade head
$ psql -f audits/verify_constraints.sql
```

## Inputs
- Domain-specific design documents and user requirements.
- Project technology stack and operational constraints.

## Outputs
- Domain-specific architecture or implementation guidelines.
- Recommended list of associated sub-agents/skills.

## Failure Modes
- When the domain document does not exist or is ambiguous.
- When the project strategy is unconfirmed and cannot be specified.

## Dependencies
- `.moai/project/` document and latest technical briefing are required.

## References
- Fowler, Martin. "Evolutionary Database Design." https://martinfowler.com/articles/evodb.html (accessed 2025-03-29).
- AWS. "Database Tuning Best Practices." https://aws.amazon.com/blogs/database/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Codified input/output and failure responses for domain skills.

## Works well with

- alfred-trust-validation (migration testing)
- sql-expert (SQL implementation)
- backend-expert (ORM integration)

## Best Practices
- Record supporting documentation (version/link) for each domain decision.
- Review performance, security, and operational requirements simultaneously at an early stage.
