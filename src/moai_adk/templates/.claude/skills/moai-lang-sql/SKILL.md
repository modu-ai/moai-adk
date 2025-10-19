---
name: moai-lang-sql
description: SQL best practices with testing frameworks, query optimization, and migration
  management
allowed-tools:
- Read
- Bash
---

# SQL Expert

## What it does

Provides SQL-specific expertise for database development, including SQL testing strategies, query optimization techniques, and migration management best practices.

## When to use

- "SQL 테스트 작성", "쿼리 최적화", "마이그레이션 관리"
- Automatically invoked when working with database projects
- SQL SPEC implementation (`/alfred:2-build`)

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

### Example 1: TDD with pgTAP
User: "/alfred:2-build SCHEMA-001"
Claude: (creates RED test with pgTAP, GREEN schema implementation, REFACTOR with indices)

### Example 2: Query optimization
User: "쿼리 성능 분석"
Claude: (runs EXPLAIN ANALYZE and suggests optimization)

## Works well with

- alfred-trust-validation (migration validation)
- alfred-code-reviewer (SQL review)
- database-expert (database design)
