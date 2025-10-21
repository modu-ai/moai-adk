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
- 도메인 관련 설계 문서 및 사용자 요구사항.
- 프로젝트 기술 스택 및 운영 제약.

## Outputs
- 도메인 특화 아키텍처 또는 구현 가이드라인.
- 연관 서브 에이전트/스킬 권장 목록.

## Failure Modes
- 도메인 근거 문서가 없거나 모호할 때.
- 프로젝트 전략이 미확정이라 구체화할 수 없을 때.

## Dependencies
- `.moai/project/` 문서와 최신 기술 브리핑이 필요합니다.

## References
- Fowler, Martin. "Evolutionary Database Design." https://martinfowler.com/articles/evodb.html (accessed 2025-03-29).
- AWS. "Database Tuning Best Practices." https://aws.amazon.com/blogs/database/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 도메인 스킬에 대한 입력/출력 및 실패 대응을 명문화했습니다.

## Works well with

- alfred-trust-validation (migration testing)
- sql-expert (SQL implementation)
- backend-expert (ORM integration)

## Best Practices
- 도메인 결정 사항마다 근거 문서(버전/링크)를 기록합니다.
- 성능·보안·운영 요구사항을 초기 단계에서 동시에 검토하세요.
