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

- "데이터베이스 설계", "스키마 최적화", "인덱스 전략", "마이그레이션"
- Automatically invoked when working with database projects
- Database SPEC implementation (`/alfred:2-build`)

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

### Example 1: Schema Design with Normalization

**❌ Before (Denormalized)**:
```sql
-- @CODE:SCHEMA-001: 데이터 중복
CREATE TABLE users (
    id INT PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100),
    city VARCHAR(100),
    city_zip_code VARCHAR(10),
    city_state VARCHAR(20),
    city_country VARCHAR(50)
);

-- 문제:
-- - city 정보 반복 (정규화 위반)
-- - 데이터 중복 → 불일치 위험
-- - 저장 공간 낭비
```

**✅ After (Normalized)**:
```sql
-- @CODE:SCHEMA-001: 정규화 (3NF)
CREATE TABLE countries (
    id INT PRIMARY KEY,
    name VARCHAR(50) UNIQUE
);

CREATE TABLE cities (
    id INT PRIMARY KEY,
    name VARCHAR(100),
    zip_code VARCHAR(10),
    state VARCHAR(20),
    country_id INT REFERENCES countries(id)
);

CREATE TABLE users (
    id INT PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100) UNIQUE,
    city_id INT REFERENCES cities(id)
);

-- 개선:
-- ✅ 데이터 중복 제거
-- ✅ 일관성 보장
-- ✅ 저장 공간 절약
```

### Example 2: Indexing for Query Performance

**RED (Test)**:
```sql
-- @TEST:INDEX-001 | SPEC: SPEC-INDEX-001.md
-- 쿼리 성능 테스트
EXPLAIN ANALYZE
SELECT * FROM orders WHERE customer_id = 123;

-- Before: Seq Scan 2000ms
-- After:  Index Scan 5ms (400배 빠름!)
```

**GREEN (Create Index)**:
```sql
-- @CODE:INDEX-001 | TEST: tests/test_index.sql
CREATE INDEX idx_orders_customer_id ON orders(customer_id);

-- 효과: 쿼리 계획 최적화
```

**REFACTOR (Composite Index)**:
```sql
-- @CODE:INDEX-001:REFACTOR | 복합 인덱스
-- 자주 함께 조회되는 컬럼 (customer_id, created_at)
CREATE INDEX idx_orders_customer_created
ON orders(customer_id, created_at DESC);

-- 개선: 더 효율적인 쿼리 실행
```

### Example 3: Migration Management (Zero-Downtime)

**Before (Downtime)**:
```sql
-- 문제: 컬럼 삭제 시 즉시 오류 발생
ALTER TABLE orders DROP COLUMN legacy_field;
-- 애플리케이션 코드가 아직 사용 중이면 다운타임 발생
```

**After (Expand-Contract Pattern)**:
```bash
# 1단계: Expand (새 구조 추가)
ALTER TABLE orders ADD COLUMN new_total DECIMAL(10,2);

# 2단계: Migrate (기존 데이터 → 새 컬럼)
UPDATE orders SET new_total = (SELECT SUM(price) FROM order_items WHERE order_id = orders.id);

# 3단계: Contract (이전 컬럼 제거)
# 애플리케이션이 new_total 사용하도록 업데이트 후
ALTER TABLE orders DROP COLUMN old_total;

# 장점: 0 다운타임 ✅
```

### Example 4: N+1 Query Problem

**❌ Before (Slow)**:
```python
# @CODE:N-PLUS-ONE-001: N+1 문제
orders = db.query("SELECT * FROM orders LIMIT 10")

for order in orders:
    customer = db.query("SELECT * FROM customers WHERE id = ?", order.customer_id)
    print(customer.name)

# 쿼리: 1 + 10 = 11번 (느림)
```

**✅ After (Eager Loading)**:
```python
# @CODE:N-PLUS-ONE-001: JOIN 최적화
orders = db.query("""
    SELECT o.*, c.name
    FROM orders o
    JOIN customers c ON o.customer_id = c.id
    LIMIT 10
""")

for order in orders:
    print(order.name)

# 쿼리: 1번 (빠름!)
```

## Keywords

"데이터베이스 설계", "정규화", "인덱싱", "마이그레이션", "쿼리 최적화", "N+1 문제", "스키마", "제약조건", "성능", "zero-downtime deployment"

## Reference

- Database design: `.moai/memory/development-guide.md#데이터베이스-설계`
- Query optimization: CLAUDE.md#쿼리-최적화
- Migration patterns: `.moai/memory/development-guide.md#데이터베이스-마이그레이션`

## Works well with

- moai-foundation-trust (마이그레이션 테스트)
- moai-domain-backend (ORM 통합)
- moai-essentials-perf (성능 최적화)
