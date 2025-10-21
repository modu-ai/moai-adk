---
name: moai-domain-database
description: Database design, schema optimization, indexing strategies, and migration management
allowed-tools:
  - Read
  - Bash
tier: 4
auto-load: "false"
---

# Database Expert

## What it does

Provides expertise in database design, schema normalization, indexing strategies, query optimization, and safe migration management for SQL and NoSQL databases.

## When to use

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

### Example 1: Schema design with normalization
User: "/alfred:2-run DB-SCHEMA-001"
Claude: (creates RED schema test, GREEN implementation with constraints, REFACTOR with indices)

### Example 2: Query optimization
User: "Improve query performance"
Claude: (analyzes EXPLAIN output and suggests index creation)

## Works well with

- alfred-trust-validation (migration testing)
- sql-expert (SQL implementation)
- backend-expert (ORM integration)
