# SQL Code Examples

Production-ready examples for modern SQL development with PostgreSQL 17.2, MySQL 9.1.0, pgTAP 1.3.3, sqlfluff 3.2.5, and migration management.

---

## Example 1: pgTAP 1.3.3 Unit Testing with PostgreSQL 17.2

### Test File: `tests/001-user-schema.sql`

```sql
-- @TEST:USER-001 | SPEC: SPEC-USER-001.md
-- pgTAP test for user schema validation
BEGIN;

SELECT plan(8);

-- Test 1: Table exists
SELECT has_table('public', 'users', 'users table should exist');

-- Test 2: Column existence
SELECT has_column('public', 'users', 'id', 'users should have id column');
SELECT has_column('public', 'users', 'email', 'users should have email column');
SELECT has_column('public', 'users', 'created_at', 'users should have created_at');

-- Test 3: Primary key constraint
SELECT has_pk('public', 'users', 'users should have primary key');

-- Test 4: Column types
SELECT col_type_is('public', 'users', 'id', 'bigint', 'id should be bigint');
SELECT col_type_is('public', 'users', 'email', 'character varying(255)',
                    'email should be varchar(255)');

-- Test 5: NOT NULL constraints
SELECT col_not_null('public', 'users', 'email', 'email should be NOT NULL');

SELECT * FROM finish();
ROLLBACK;
```

**Key Features**:
- ✅ `BEGIN;` and `ROLLBACK;` for transaction isolation
- ✅ `SELECT plan(N);` declares number of tests upfront
- ✅ Descriptive test names for clear failure messages
- ✅ Schema validation (tables, columns, constraints)
- ✅ Type checking with `col_type_is()`

**Run Commands**:
```bash
# Install pgTAP extension (PostgreSQL 17.2)
psql -d testdb -c "CREATE EXTENSION IF NOT EXISTS pgtap;"

# Run single test file
pg_prove -d testdb tests/001-user-schema.sql

# Run all tests with verbose output
pg_prove -d testdb -v tests/*.sql

# Generate TAP output for CI/CD
pg_prove -d testdb tests/*.sql --formatter TAP::Formatter::HTML > report.html
```

---

## Example 2: Advanced pgTAP Function Testing

### Test File: `tests/002-user-functions.sql`

```sql
-- @TEST:USER-002 | SPEC: SPEC-USER-002.md
-- Test user validation functions
BEGIN;

SELECT plan(12);

-- Setup test data
INSERT INTO users (id, email, username, created_at)
VALUES (1, 'alice@example.com', 'alice', NOW()),
       (2, 'bob@example.com', 'bob', NOW());

-- Test 1: Function exists
SELECT has_function('public', 'validate_email',
                     'validate_email function should exist');

-- Test 2: Function returns correct type
SELECT function_returns('public', 'validate_email', 'boolean',
                        'validate_email should return boolean');

-- Test 3: Valid email passes
SELECT ok(validate_email('test@example.com'),
          'Valid email should pass validation');

-- Test 4: Invalid emails fail
SELECT ok(NOT validate_email('invalid-email'),
          'Invalid email should fail validation');
SELECT ok(NOT validate_email(''),
          'Empty email should fail validation');

-- Test 5: Stored procedure behavior
SELECT lives_ok('CALL update_user_status(1, ''active'')',
                'update_user_status should not throw error');

-- Test 6: Trigger behavior
UPDATE users SET email = 'newemail@example.com' WHERE id = 1;
SELECT ok(
    (SELECT updated_at > created_at FROM users WHERE id = 1),
    'Trigger should update updated_at timestamp'
);

-- Test 7: Data integrity
SELECT is(
    (SELECT COUNT(*) FROM users WHERE email LIKE '%@%'),
    2::bigint,
    'All users should have valid email format'
);

-- Test 8: Aggregate function testing
SELECT ok(
    (SELECT AVG(LENGTH(email)) FROM users) > 10,
    'Average email length should be reasonable'
);

-- Test 9: Exception handling
SELECT throws_ok(
    'INSERT INTO users (id, email) VALUES (1, ''duplicate@example.com'')',
    '23505',
    'duplicate key value violates unique constraint',
    'Duplicate email should raise unique constraint error'
);

-- Cleanup
DELETE FROM users WHERE id IN (1, 2);

SELECT * FROM finish();
ROLLBACK;
```

**Key Features**:
- ✅ `has_function()` and `function_returns()` for function validation
- ✅ `lives_ok()` for testing procedures don't throw errors
- ✅ `throws_ok()` for exception handling verification
- ✅ Data-driven tests with `is()` comparisons
- ✅ Trigger behavior validation

**Run Commands**:
```bash
# Run with coverage reporting (requires pg_stat_statements)
pg_prove -d testdb tests/002-user-functions.sql --verbose

# Run with timing information
pg_prove -d testdb tests/002-user-functions.sql --timer
```

---

## Example 3: sqlfluff 3.2.5 Linting and Formatting

### Project Configuration: `.sqlfluff`

```ini
[sqlfluff]
# Dialect selection (supports postgres, mysql, sqlite, tsql, etc.)
dialect = postgres
templater = raw
sql_file_exts = .sql,.sql.j2,.dml,.ddl

# Maximum line length
max_line_length = 100

# Exclude directories
exclude_rules = None
ignore = .git,venv,.venv,target,dbt_packages

[sqlfluff:indentation]
indent_unit = space
tab_space_size = 2
indented_joins = True
indented_using_on = True

[sqlfluff:rules]
# Capitalization rules
capitalisation_policy = lower
extended_capitalisation_policy = lower

# Layout rules
comma_style = trailing
layout_type = postgres

[sqlfluff:rules:L003]
# Indentation errors
hanging_indents = True

[sqlfluff:rules:L010]
# Capitalization of keywords
capitalisation_policy = lower

[sqlfluff:rules:L014]
# Unquoted identifiers
extended_capitalisation_policy = lower

[sqlfluff:rules:L030]
# Function name capitalization
capitalisation_policy = lower
```

### Example SQL: `queries/user_analytics.sql`

```sql
-- @CODE:ANALYTICS-003 | SPEC: SPEC-ANALYTICS-003.md | TEST: tests/003-analytics.sql
-- User analytics query with proper formatting
select
  u.id,
  u.username,
  u.email,
  count(o.id) as order_count,
  sum(o.total_amount) as total_spent,
  avg(o.total_amount) as avg_order_value,
  max(o.created_at) as last_order_date
from
  users as u
  inner join orders as o
    on u.id = o.user_id
where
  o.status = 'completed'
  and o.created_at >= current_date - interval '30 days'
group by
  u.id,
  u.username,
  u.email
having
  count(o.id) >= 5
order by
  total_spent desc
limit 100;
```

**Key Features**:
- ✅ Lowercase keywords and identifiers
- ✅ Trailing commas (easier to add/remove lines)
- ✅ Indented joins with explicit `on` clauses
- ✅ Consistent 2-space indentation
- ✅ Clear column aliasing

**Run Commands**:
```bash
# Check SQL files without fixing
sqlfluff lint queries/

# Auto-fix linting issues
sqlfluff fix queries/

# Lint specific file with detailed output
sqlfluff lint queries/user_analytics.sql --verbose

# Format SQL (fix + standardize layout)
sqlfluff format queries/user_analytics.sql

# Check differences before applying fixes
sqlfluff fix queries/ --diff

# CI/CD integration (exit code 1 if violations found)
sqlfluff lint queries/ --dialect postgres --format github-annotation
```

---

## Example 4: Database Migration Management

### Migration Structure

```
migrations/
├── V001__create_users_table.sql
├── V002__add_email_index.sql
├── V003__create_orders_table.sql
└── V004__add_user_status.sql
```

### Migration: `V001__create_users_table.sql`

```sql
-- @CODE:MIGRATION-001 | SPEC: SPEC-MIGRATION-001.md
-- Migration: Create users table with proper constraints
-- Version: 1.0.0
-- Date: 2025-10-22

begin;

create table if not exists users (
  id bigserial primary key,
  email varchar(255) not null unique,
  username varchar(100) not null unique,
  password_hash varchar(255) not null,
  status varchar(20) not null default 'active',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),

  -- Constraints
  constraint users_email_format_check
    check (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$'),
  constraint users_status_check
    check (status in ('active', 'inactive', 'suspended', 'deleted'))
);

-- Indexes for performance
create index idx_users_email on users (email);
create index idx_users_username on users (username);
create index idx_users_status on users (status);
create index idx_users_created_at on users (created_at desc);

-- Trigger for updated_at timestamp
create or replace function update_updated_at_column()
returns trigger as $$
begin
  new.updated_at = now();
  return new;
end;
$$ language plpgsql;

create trigger set_users_updated_at
  before update on users
  for each row
  execute function update_updated_at_column();

-- Comments for documentation
comment on table users is 'Core user accounts with authentication';
comment on column users.email is 'Unique email address for user login';
comment on column users.status is 'User account status (active/inactive/suspended/deleted)';

commit;
```

**Key Features**:
- ✅ Transaction wrapping (`begin;` / `commit;`)
- ✅ `if not exists` for idempotency
- ✅ Check constraints for data validation
- ✅ Strategic indexes for common queries
- ✅ Trigger for automatic timestamp updates
- ✅ Table and column comments for documentation

### Migration: `V002__add_email_index.sql`

```sql
-- @CODE:MIGRATION-002 | SPEC: SPEC-MIGRATION-002.md
-- Migration: Add composite index for email queries
-- Version: 1.0.0
-- Date: 2025-10-22

begin;

-- Add composite index for email + status queries
create index concurrently if not exists idx_users_email_status
  on users (email, status)
  where status = 'active';

-- Add GIN index for full-text search on username
create index concurrently if not exists idx_users_username_gin
  on users using gin (to_tsvector('english', username));

commit;
```

**Key Features**:
- ✅ `create index concurrently` (no table locking)
- ✅ Partial index with `where` clause (smaller, faster)
- ✅ GIN index for full-text search
- ✅ `if not exists` for safe re-runs

**Run Commands**:
```bash
# Using Flyway
flyway migrate -configFiles=flyway.conf

# Using Liquibase
liquibase update --changelog-file=changelog.xml

# Using Alembic (Python)
alembic upgrade head

# Manual migration with psql
psql -d proddb -f migrations/V001__create_users_table.sql
```

---

## Example 5: Query Optimization and Performance

### Slow Query: `queries/slow_analytics.sql`

```sql
-- BEFORE: Slow query (no indexes, inefficient joins)
select
  u.username,
  count(*) as order_count
from
  users u,
  orders o
where
  u.id = o.user_id
  and o.status = 'completed'
group by
  u.username
order by
  order_count desc;
```

### Optimized Query: `queries/fast_analytics.sql`

```sql
-- AFTER: Optimized query (proper joins, indexes, CTEs)
-- @CODE:ANALYTICS-005 | SPEC: SPEC-ANALYTICS-005.md | TEST: tests/005-analytics.sql

-- Use CTE for better readability and optimization
with completed_orders as (
  select
    user_id,
    count(*) as order_count
  from
    orders
  where
    status = 'completed'
  group by
    user_id
)
select
  u.username,
  co.order_count
from
  users as u
  inner join completed_orders as co
    on u.id = co.user_id
order by
  co.order_count desc;

-- Required indexes:
-- create index idx_orders_status_user on orders (status, user_id);
-- create index idx_users_id_username on users (id, username);
```

**Optimization Techniques**:
- ✅ Use explicit `inner join` instead of implicit comma joins
- ✅ Add indexes on filter and join columns
- ✅ Use CTEs for complex subqueries
- ✅ Filter early (push down `where` clauses)
- ✅ Use covering indexes to avoid table lookups

### Performance Analysis Commands

```bash
# PostgreSQL: EXPLAIN ANALYZE
psql -d proddb -c "
EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)
SELECT * FROM users WHERE email = 'test@example.com';
" > explain.json

# PostgreSQL: Check slow queries (requires pg_stat_statements)
psql -d proddb -c "
SELECT
  query,
  calls,
  total_exec_time,
  mean_exec_time,
  max_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 20;
"

# MySQL: EXPLAIN ANALYZE
mysql -u root -p proddb -e "
EXPLAIN ANALYZE
SELECT * FROM users WHERE email = 'test@example.com'\G
"

# MySQL: Slow query log analysis
mysqldumpslow -s t -t 10 /var/log/mysql/slow-query.log
```

---

## Example 6: TDD Workflow for SQL Features

### RED Phase: Write Failing Test

**Test File: `tests/003-order-validation.sql`**

```sql
-- @TEST:ORDER-001 | SPEC: SPEC-ORDER-001.md
-- Test order validation logic (RED phase)
BEGIN;

SELECT plan(5);

-- Test: Function should reject orders with negative amounts
SELECT ok(
  NOT validate_order(user_id := 1, amount := -100, status := 'pending'),
  'validate_order should reject negative amounts'
);

-- Test: Function should reject orders with invalid status
SELECT ok(
  NOT validate_order(user_id := 1, amount := 100, status := 'invalid_status'),
  'validate_order should reject invalid status'
);

-- Test: Function should accept valid orders
SELECT ok(
  validate_order(user_id := 1, amount := 100, status := 'pending'),
  'validate_order should accept valid orders'
);

-- Test: Function should enforce user existence
SELECT throws_ok(
  'SELECT validate_order(user_id := 99999, amount := 100, status := ''pending'')',
  'P0001',
  'User does not exist',
  'validate_order should enforce user existence'
);

-- Test: Function should enforce amount range
SELECT ok(
  NOT validate_order(user_id := 1, amount := 1000000, status := 'pending'),
  'validate_order should enforce maximum amount'
);

SELECT * FROM finish();
ROLLBACK;
```

**Run Test (should FAIL)**:
```bash
pg_prove -d testdb tests/003-order-validation.sql
# Expected: 5 failures (function doesn't exist yet)
```

### GREEN Phase: Implement Function

**Implementation: `migrations/V005__add_order_validation.sql`**

```sql
-- @CODE:ORDER-001 | SPEC: SPEC-ORDER-001.md | TEST: tests/003-order-validation.sql
-- GREEN: Implement order validation function
BEGIN;

create or replace function validate_order(
  user_id bigint,
  amount numeric(10, 2),
  status varchar(20)
) returns boolean as $$
declare
  user_exists boolean;
begin
  -- Check if user exists
  select exists(select 1 from users where id = user_id)
  into user_exists;

  if not user_exists then
    raise exception 'User does not exist'
      using errcode = 'P0001';
  end if;

  -- Validate amount
  if amount <= 0 then
    return false;
  end if;

  if amount > 100000 then
    return false;
  end if;

  -- Validate status
  if status not in ('pending', 'completed', 'cancelled') then
    return false;
  end if;

  -- All checks passed
  return true;
end;
$$ language plpgsql;

COMMIT;
```

**Run Test (should PASS)**:
```bash
# Apply migration
psql -d testdb -f migrations/V005__add_order_validation.sql

# Run test again
pg_prove -d testdb tests/003-order-validation.sql
# Expected: 5 passes
```

### REFACTOR Phase: Improve Code

**Refactored: `migrations/V006__refactor_order_validation.sql`**

```sql
-- @CODE:ORDER-001 | SPEC: SPEC-ORDER-001.md | TEST: tests/003-order-validation.sql
-- REFACTOR: Improve order validation with better error messages
BEGIN;

create or replace function validate_order(
  p_user_id bigint,
  p_amount numeric(10, 2),
  p_status varchar(20)
) returns boolean as $$
declare
  v_user_exists boolean;
  v_max_amount constant numeric(10, 2) := 100000.00;
  v_valid_statuses constant varchar[] := array['pending', 'completed', 'cancelled'];
begin
  -- Check user existence (early return)
  select exists(select 1 from users where id = p_user_id)
  into v_user_exists;

  if not v_user_exists then
    raise exception 'User % does not exist', p_user_id
      using errcode = 'P0001',
            hint = 'Please provide a valid user_id';
  end if;

  -- Validate amount (with descriptive messages)
  if p_amount <= 0 then
    raise notice 'Order amount must be positive: %', p_amount;
    return false;
  end if;

  if p_amount > v_max_amount then
    raise notice 'Order amount % exceeds maximum %', p_amount, v_max_amount;
    return false;
  end if;

  -- Validate status using array comparison
  if not (p_status = any(v_valid_statuses)) then
    raise notice 'Invalid status: %. Valid values: %', p_status, v_valid_statuses;
    return false;
  end if;

  return true;
end;
$$ language plpgsql stable;

-- Add comment for documentation
comment on function validate_order is 'Validates order data before insert/update';

COMMIT;
```

**Key Improvements**:
- ✅ Parameter prefix convention (`p_` for params, `v_` for variables)
- ✅ Constants for magic numbers
- ✅ Array-based status validation (easier to maintain)
- ✅ Better error messages with hints
- ✅ `stable` function attribute (optimization hint)
- ✅ Function documentation comment

**Run Test (should still PASS)**:
```bash
psql -d testdb -f migrations/V006__refactor_order_validation.sql
pg_prove -d testdb tests/003-order-validation.sql
# Expected: 5 passes (no behavior change)
```

---

## Example 7: MySQL-specific Testing and Optimization

### MySQL Test File: `tests/mysql_users.sql`

```sql
-- @TEST:USER-MYSQL-001 | SPEC: SPEC-USER-MYSQL-001.md
-- MySQL-specific testing (using mysql-test framework)
-- Run with: mysql-test-run.pl tests/mysql_users.test

-- Setup
DROP TABLE IF EXISTS users;
CREATE TABLE users (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  username VARCHAR(100) NOT NULL UNIQUE,
  status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_email (email),
  INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Test 1: Insert valid user
INSERT INTO users (email, username) VALUES ('alice@example.com', 'alice');
SELECT COUNT(*) = 1 AS test_insert FROM users WHERE email = 'alice@example.com';

-- Test 2: Duplicate email should fail
--error ER_DUP_ENTRY
INSERT INTO users (email, username) VALUES ('alice@example.com', 'alice2');

-- Test 3: ENUM validation
--error ER_TRUNCATED_WRONG_VALUE_FOR_FIELD
INSERT INTO users (email, username, status) VALUES ('bob@example.com', 'bob', 'invalid');

-- Test 4: Auto-update timestamp
INSERT INTO users (email, username) VALUES ('charlie@example.com', 'charlie');
SELECT SLEEP(1);
UPDATE users SET username = 'charlie_updated' WHERE email = 'charlie@example.com';
SELECT updated_at > created_at AS test_timestamp FROM users WHERE email = 'charlie@example.com';

-- Cleanup
DROP TABLE users;
```

**Run Commands**:
```bash
# MySQL Test Framework
mysql-test-run.pl tests/mysql_users.test

# Alternative: Manual testing with mysql client
mysql -u root -p testdb < tests/mysql_users.sql
```

---

## Example 8: CI/CD Integration

### GitHub Actions: `.github/workflows/sql-ci.yml`

```yaml
name: SQL CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.13'

      - name: Install sqlfluff
        run: pip install sqlfluff==3.2.5

      - name: Lint SQL files
        run: sqlfluff lint queries/ migrations/ --dialect postgres --format github-annotation

  test-postgres:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:17.2
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v4

      - name: Install pgTAP
        run: |
          sudo apt-get update
          sudo apt-get install -y postgresql-17-pgtap

      - name: Run migrations
        env:
          PGPASSWORD: postgres
        run: |
          for migration in migrations/*.sql; do
            psql -h localhost -U postgres -d testdb -f "$migration"
          done

      - name: Run pgTAP tests
        env:
          PGPASSWORD: postgres
        run: |
          pg_prove -h localhost -U postgres -d testdb tests/*.sql

  test-mysql:
    runs-on: ubuntu-latest
    services:
      mysql:
        image: mysql:9.1.0
        env:
          MYSQL_ROOT_PASSWORD: root
          MYSQL_DATABASE: testdb
        options: >-
          --health-cmd "mysqladmin ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 3306:3306

    steps:
      - uses: actions/checkout@v4

      - name: Run MySQL migrations
        run: |
          for migration in migrations/mysql/*.sql; do
            mysql -h 127.0.0.1 -u root -proot testdb < "$migration"
          done

      - name: Run MySQL tests
        run: |
          mysql -h 127.0.0.1 -u root -proot testdb < tests/mysql_suite.sql
```

---

## Summary

These examples demonstrate:
- ✅ **pgTAP 1.3.3**: Unit testing with fixtures, functions, triggers, and exception handling
- ✅ **sqlfluff 3.2.5**: Linting, formatting, and style enforcement
- ✅ **PostgreSQL 17.2**: Modern SQL features, migrations, and performance optimization
- ✅ **MySQL 9.1.0**: MySQL-specific testing and optimization
- ✅ **TDD Workflow**: RED → GREEN → REFACTOR cycle with tests and implementations
- ✅ **CI/CD Integration**: Automated testing and linting in GitHub Actions

**Next Steps**:
1. Review `reference.md` for detailed CLI commands and configuration options
2. Integrate pgTAP and sqlfluff into your project's CI/CD pipeline
3. Follow the TDD workflow for all new SQL features
4. Use `EXPLAIN ANALYZE` to optimize slow queries

---

_For detailed reference documentation, see [reference.md](reference.md)_
_Last updated: 2025-10-22_
