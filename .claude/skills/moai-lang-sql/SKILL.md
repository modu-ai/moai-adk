---

name: moai-lang-sql
description: SQL best practices with pgTAP, sqlfluff 3.2, query optimization, and migration management.

---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: sql, lang, moai  


## Quick Reference

### SQL Standards (November 2025)

**Database Versions:**
- PostgreSQL 17.2 (primary recommendation)
- MySQL 9.1.0 (alternative)
- SQLite 3.45.0 (embedded/testing)

**Code Quality:**
- sqlfluff 3.2.5 for linting and formatting
- pgTAP 1.3.3 for database testing
- sqitch 1.4.1 for migration management

**Query Optimization:**
- EXPLAIN ANALYZE for performance analysis
- Proper indexing strategies
- Query plan optimization
- Connection pooling

**Best Practices:**
- Parameterized queries (prevent SQL injection)
- Transaction management with proper isolation levels
- Schema versioning with migration tools
- Test coverage for critical queries


## Implementation Guide

### Tool Version Matrix (2025-10-22)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **PostgreSQL** | 17.2 | Primary database | ✅ Current |
| **MySQL** | 9.1.0 | Alternative database | ✅ Current |
| **sqlfluff** | 3.2.5 | SQL linting | ✅ Current |
| **pgTAP** | 1.3.3 | Database testing | ✅ Current |
| **sqitch** | 1.4.1 | Migration management | ✅ Current |

### Schema Design Pattern

```sql
-- PostgreSQL schema with best practices
CREATE SCHEMA IF NOT EXISTS app;

-- User table with proper constraints
CREATE TABLE app.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT username_length CHECK (LENGTH(username) >= 3)
);

-- Indexes for common queries
CREATE INDEX idx_users_email ON app.users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_username ON app.users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON app.users(created_at DESC);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION app.update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON app.users
    FOR EACH ROW
    EXECUTE FUNCTION app.update_updated_at();
```

### Query Optimization

```sql
-- Bad: N+1 query pattern
SELECT * FROM orders WHERE user_id = $1;  -- Called 100 times
SELECT * FROM order_items WHERE order_id = $1;  -- Called 100 times

-- Good: JOIN with proper indexing
SELECT 
    o.id AS order_id,
    o.total,
    o.created_at,
    oi.id AS item_id,
    oi.product_id,
    oi.quantity,
    oi.price
FROM orders o
LEFT JOIN order_items oi ON o.id = oi.order_id
WHERE o.user_id = $1
ORDER BY o.created_at DESC;

-- Index to support the query
CREATE INDEX idx_orders_user_created ON orders(user_id, created_at DESC);
CREATE INDEX idx_order_items_order ON order_items(order_id);

-- Analyze query performance
EXPLAIN ANALYZE
SELECT 
    o.id AS order_id,
    o.total,
    o.created_at,
    oi.id AS item_id,
    oi.product_id,
    oi.quantity,
    oi.price
FROM orders o
LEFT JOIN order_items oi ON o.id = oi.order_id
WHERE o.user_id = 'user-uuid-here'
ORDER BY o.created_at DESC;
```

### Transaction Management

```sql
-- Proper transaction handling with isolation level
BEGIN TRANSACTION ISOLATION LEVEL READ COMMITTED;

-- Update user balance
UPDATE app.wallets
SET balance = balance - 100.00,
    updated_at = NOW()
WHERE user_id = $1
  AND balance >= 100.00;

-- Check if update succeeded
IF NOT FOUND THEN
    ROLLBACK;
    RAISE EXCEPTION 'Insufficient balance';
END IF;

-- Create transaction record
INSERT INTO app.transactions (
    user_id,
    amount,
    type,
    description
) VALUES (
    $1,
    -100.00,
    'withdrawal',
    'Account withdrawal'
);

COMMIT;
```

### Database Testing with pgTAP

```sql
-- Test file: tests/user_schema_test.sql
BEGIN;

SELECT plan(5);

-- Test 1: Table exists
SELECT has_table('app', 'users', 'Users table should exist');

-- Test 2: Required columns exist
SELECT has_column('app', 'users', 'id', 'Users table should have id column');
SELECT has_column('app', 'users', 'email', 'Users table should have email column');

-- Test 3: Constraints work
SELECT throws_ok(
    $$INSERT INTO app.users (email, username, password_hash) 
      VALUES ('invalid-email', 'test', 'hash')$$,
    'email_format',
    'Should reject invalid email format'
);

-- Test 4: Unique constraint
SELECT throws_ok(
    $$INSERT INTO app.users (email, username, password_hash) 
      VALUES ('test@example.com', 'testuser', 'hash');
      INSERT INTO app.users (email, username, password_hash) 
      VALUES ('test@example.com', 'testuser2', 'hash2')$$,
    '23505',
    'Should reject duplicate email'
);

SELECT * FROM finish();
ROLLBACK;
```

### Migration Management with Sqitch

```sql
-- deploy/add_users_table.sql
-- Deploy app:add_users_table to pg

BEGIN;

CREATE TABLE app.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;

-- revert/add_users_table.sql
-- Revert app:add_users_table from pg

BEGIN;

DROP TABLE IF EXISTS app.users;

COMMIT;

-- verify/add_users_table.sql
-- Verify app:add_users_table on pg

BEGIN;

SELECT id, email, username, password_hash, created_at
FROM app.users
WHERE FALSE;

ROLLBACK;
```


## Advanced Patterns

### Partitioning for Large Tables

```sql
-- Partition large logs table by month
CREATE TABLE app.logs (
    id BIGSERIAL,
    user_id UUID NOT NULL,
    action VARCHAR(100) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL
) PARTITION BY RANGE (created_at);

-- Create partitions
CREATE TABLE app.logs_2025_01 PARTITION OF app.logs
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE app.logs_2025_02 PARTITION OF app.logs
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

-- Index on each partition
CREATE INDEX idx_logs_2025_01_user ON app.logs_2025_01(user_id);
CREATE INDEX idx_logs_2025_02_user ON app.logs_2025_02(user_id);

-- Query automatically uses appropriate partition
SELECT * FROM app.logs
WHERE created_at BETWEEN '2025-01-15' AND '2025-01-20'
  AND user_id = 'user-uuid';
```

### Full-Text Search

```sql
-- Add full-text search to articles
ALTER TABLE app.articles
ADD COLUMN search_vector tsvector
GENERATED ALWAYS AS (
    to_tsvector('english', 
        coalesce(title, '') || ' ' || 
        coalesce(content, '') || ' ' || 
        coalesce(tags, '')
    )
) STORED;

-- Index for fast full-text search
CREATE INDEX idx_articles_search ON app.articles USING GIN(search_vector);

-- Search query
SELECT 
    id,
    title,
    ts_rank(search_vector, query) AS rank
FROM app.articles,
     to_tsquery('english', 'postgresql & performance') AS query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 10;
```

### JSON/JSONB Operations

```sql
-- Store flexible metadata in JSONB
CREATE TABLE app.products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index on JSONB field
CREATE INDEX idx_products_metadata_category 
    ON app.products((metadata->>'category'));

-- Query JSONB data
SELECT 
    id,
    name,
    metadata->>'category' AS category,
    metadata->'pricing'->>'amount' AS price
FROM app.products
WHERE metadata->>'category' = 'electronics'
  AND (metadata->'pricing'->>'amount')::numeric < 100;

-- Update JSONB field
UPDATE app.products
SET metadata = metadata || '{"featured": true}'::jsonb
WHERE id = 'product-uuid';
```

### Common Table Expressions (CTEs)

```sql
-- Recursive CTE for hierarchical data
WITH RECURSIVE category_tree AS (
    -- Base case: root categories
    SELECT 
        id,
        name,
        parent_id,
        1 AS level,
        ARRAY[id] AS path
    FROM app.categories
    WHERE parent_id IS NULL
    
    UNION ALL
    
    -- Recursive case: child categories
    SELECT 
        c.id,
        c.name,
        c.parent_id,
        ct.level + 1,
        ct.path || c.id
    FROM app.categories c
    INNER JOIN category_tree ct ON c.parent_id = ct.id
)
SELECT 
    id,
    REPEAT('  ', level - 1) || name AS indented_name,
    level,
    path
FROM category_tree
ORDER BY path;
```

### Connection Pooling

```javascript
// Node.js with pg pool
import { Pool } from 'pg';

const pool = new Pool({
  host: process.env.DB_HOST,
  port: parseInt(process.env.DB_PORT || '5432'),
  database: process.env.DB_NAME,
  user: process.env.DB_USER,
  password: process.env.DB_PASSWORD,
  max: 20,                    // Maximum connections
  idleTimeoutMillis: 30000,   // Close idle connections after 30s
  connectionTimeoutMillis: 2000, // Timeout if can't get connection
});

// Use parameterized queries (prevent SQL injection)
async function getUserByEmail(email: string) {
  const result = await pool.query(
    'SELECT id, username, email FROM app.users WHERE email = $1',
    [email]
  );
  return result.rows[0];
}

// Transaction helper
async function transferMoney(fromUserId: string, toUserId: string, amount: number) {
  const client = await pool.connect();
  
  try {
    await client.query('BEGIN');
    
    await client.query(
      'UPDATE app.wallets SET balance = balance - $1 WHERE user_id = $2',
      [amount, fromUserId]
    );
    
    await client.query(
      'UPDATE app.wallets SET balance = balance + $1 WHERE user_id = $2',
      [amount, toUserId]
    );
    
    await client.query('COMMIT');
  } catch (error) {
    await client.query('ROLLBACK');
    throw error;
  } finally {
    client.release();
  }
}
```


## Best Practices

### DO
- Use parameterized queries to prevent SQL injection
- Apply proper indexing for common query patterns
- Test database logic with pgTAP
- Use migrations for schema changes (sqitch, Flyway, Liquibase)
- Analyze query performance with EXPLAIN ANALYZE
- Use transactions for multi-step operations
- Apply appropriate isolation levels
- Use connection pooling in applications
- Document complex queries with comments
- Monitor slow query logs

### DON'T
- Concatenate user input into SQL strings
- Create indexes without analyzing query patterns
- Skip database testing
- Modify production schema manually
- Ignore query performance warnings
- Use SELECT * in production code
- Forget to handle transaction rollbacks
- Leave unused indexes (they slow down writes)
- Use VARCHAR without length limits
- Ignore database-specific optimizations


## Works Well With

- `moai-foundation-trust` (TRUST 5 testing principles)
- `moai-domain-backend` (Backend integration)
- `moai-essentials-perf` (Query performance optimization)
- `moai-security-api` (SQL injection prevention)


## Changelog

- **v2.0.0** (2025-11-21): 3-level structure with comprehensive SQL patterns
- **v1.0.0** (2025-03-29): Initial release


**End of Skill** | Updated 2025-11-21



## Context7 Integration

### Related Libraries & Tools
- [PostgreSQL](/postgres/postgres): Advanced relational database
- [MySQL](/mysql/mysql-server): Popular SQL database

### Official Documentation
- [Documentation](https://www.postgresql.org/docs/)
- [API Reference](https://www.postgresql.org/docs/current/sql.html)

### Version-Specific Guides
Latest stable version: 16
- [Release Notes](https://www.postgresql.org/docs/release/)
- [Migration Guide](https://www.postgresql.org/docs/current/upgrading.html)
