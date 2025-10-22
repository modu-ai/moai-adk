# SQL CLI Reference

Quick reference for PostgreSQL 17.2, MySQL 9.1.0, pgTAP 1.3.3, sqlfluff 3.2.5, and migration tools.

---

## Tool Versions (2025-10-22)

| Tool | Version | Release Date | Purpose |
|------|---------|--------------|---------|
| **PostgreSQL** | 17.2 | 2025-10-15 | Primary RDBMS |
| **MySQL** | 9.1.0 | 2024-10-21 | Alternative RDBMS |
| **pgTAP** | 1.3.3 | 2024-11-20 | PostgreSQL unit testing |
| **sqlfluff** | 3.2.5 | 2025-09-15 | SQL linting and formatting |
| **Flyway** | 10.29.0 | 2025-10-01 | Schema migration tool |
| **Liquibase** | 4.30.0 | 2025-09-20 | Database change management |

---

## PostgreSQL 17.2

### Installation

```bash
# macOS (Homebrew)
brew install postgresql@17

# Ubuntu/Debian
sudo apt-get update
sudo apt-get install postgresql-17 postgresql-client-17

# Verify installation
psql --version
# Expected: psql (PostgreSQL) 17.2
```

### Common Commands

```bash
# Start PostgreSQL service
brew services start postgresql@17   # macOS
sudo systemctl start postgresql     # Linux

# Create database
createdb mydb

# Connect to database
psql -d mydb

# Execute SQL file
psql -d mydb -f script.sql

# Dump database
pg_dump mydb > backup.sql

# Restore database
psql -d mydb < backup.sql

# List databases
psql -l

# Execute query from command line
psql -d mydb -c "SELECT * FROM users LIMIT 10;"
```

### Configuration Files

```bash
# Main config (find with: psql -c "SHOW config_file;")
/usr/local/var/postgresql@17/postgresql.conf

# Authentication
/usr/local/var/postgresql@17/pg_hba.conf
```

### Performance Tuning

```sql
-- Enable pg_stat_statements for query analysis
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- View top 10 slowest queries
SELECT
  query,
  calls,
  total_exec_time,
  mean_exec_time,
  rows
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Analyze query performance
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT * FROM users WHERE email = 'test@example.com';

-- Check index usage
SELECT
  schemaname,
  tablename,
  indexname,
  idx_scan,
  idx_tup_read
FROM pg_stat_user_indexes
ORDER BY idx_scan ASC;

-- View table sizes
SELECT
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

---

## MySQL 9.1.0

### Installation

```bash
# macOS (Homebrew)
brew install mysql@9.1

# Ubuntu/Debian
wget https://dev.mysql.com/get/mysql-apt-config_0.8.33-1_all.deb
sudo dpkg -i mysql-apt-config_0.8.33-1_all.deb
sudo apt-get update
sudo apt-get install mysql-server

# Verify installation
mysql --version
# Expected: mysql  Ver 9.1.0
```

### Common Commands

```bash
# Start MySQL service
brew services start mysql@9.1       # macOS
sudo systemctl start mysql          # Linux

# Connect to MySQL
mysql -u root -p

# Create database
mysql -u root -p -e "CREATE DATABASE mydb;"

# Execute SQL file
mysql -u root -p mydb < script.sql

# Dump database
mysqldump -u root -p mydb > backup.sql

# Restore database
mysql -u root -p mydb < backup.sql

# List databases
mysql -u root -p -e "SHOW DATABASES;"

# Execute query from command line
mysql -u root -p mydb -e "SELECT * FROM users LIMIT 10;"
```

### Performance Analysis

```sql
-- Enable slow query log
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 1;

-- View query execution plan
EXPLAIN ANALYZE
SELECT * FROM users WHERE email = 'test@example.com';

-- Check index cardinality
SHOW INDEX FROM users;

-- View table sizes
SELECT
  table_name,
  ROUND(((data_length + index_length) / 1024 / 1024), 2) AS size_mb
FROM information_schema.tables
WHERE table_schema = 'mydb'
ORDER BY (data_length + index_length) DESC;

-- Analyze slow queries
SELECT
  query_time,
  sql_text
FROM mysql.slow_log
ORDER BY query_time DESC
LIMIT 10;
```

---

## pgTAP 1.3.3

### Installation

```bash
# PostgreSQL extension
psql -d mydb -c "CREATE EXTENSION IF NOT EXISTS pgtap;"

# pg_prove runner (Perl TAP harness)
cpan TAP::Parser::SourceHandler::pgTAP

# Or via package manager
sudo apt-get install libtap-parser-sourcehandler-pgtap-perl
```

### Common Test Functions

```sql
-- Plan declaration
SELECT plan(N);           -- Declare number of tests
SELECT no_plan();         -- No fixed plan
SELECT * FROM finish();   -- End test suite

-- Schema tests
SELECT has_table('public', 'users');
SELECT has_column('public', 'users', 'email');
SELECT has_pk('public', 'users');
SELECT has_index('public', 'users', 'idx_email');

-- Type tests
SELECT col_type_is('public', 'users', 'id', 'bigint');
SELECT col_not_null('public', 'users', 'email');
SELECT col_has_default('public', 'users', 'status');

-- Function tests
SELECT has_function('public', 'validate_email');
SELECT function_returns('public', 'validate_email', 'boolean');

-- Data tests
SELECT is(actual, expected, 'description');
SELECT ok(condition, 'description');
SELECT lives_ok('SQL statement', 'description');
SELECT throws_ok('SQL statement', 'error_code', 'error_message', 'description');

-- Row comparison
SELECT is(
  (SELECT COUNT(*) FROM users),
  5::bigint,
  'Should have 5 users'
);

-- Set comparison
SELECT results_eq(
  'SELECT id FROM users ORDER BY id',
  ARRAY[1, 2, 3, 4, 5],
  'User IDs should match'
);
```

### Running Tests

```bash
# Single test file
pg_prove -d mydb tests/001-users.sql

# All tests in directory
pg_prove -d mydb tests/*.sql

# Verbose output
pg_prove -d mydb -v tests/*.sql

# With timing
pg_prove -d mydb --timer tests/*.sql

# Generate HTML report
pg_prove -d mydb tests/*.sql --formatter TAP::Formatter::HTML > report.html

# CI/CD mode (exit code 1 on failure)
pg_prove -d mydb tests/*.sql || exit 1
```

### Test File Template

```sql
-- tests/example.sql
BEGIN;

SELECT plan(N);

-- Your tests here
SELECT ok(true, 'This test passes');

SELECT * FROM finish();
ROLLBACK;
```

---

## sqlfluff 3.2.5

### Installation

```bash
# Python pip
pip install sqlfluff==3.2.5

# Python pipx (isolated environment)
pipx install sqlfluff==3.2.5

# Verify installation
sqlfluff --version
# Expected: sqlfluff, version 3.2.5
```

### CLI Commands

```bash
# Lint SQL files
sqlfluff lint queries/

# Lint with specific dialect
sqlfluff lint queries/ --dialect postgres

# Auto-fix issues
sqlfluff fix queries/

# Format SQL (fix + layout standardization)
sqlfluff format queries/

# Show diff before fixing
sqlfluff fix queries/ --diff

# Exclude specific rules
sqlfluff lint queries/ --exclude-rules L003,L010

# Include only specific rules
sqlfluff lint queries/ --rules L001,L002

# Lint specific file
sqlfluff lint queries/users.sql --verbose

# Check specific dialect features
sqlfluff dialects

# Generate configuration file
sqlfluff generate-config > .sqlfluff
```

### Configuration File: `.sqlfluff`

```ini
[sqlfluff]
dialect = postgres
templater = raw
sql_file_exts = .sql,.sql.j2,.dml,.ddl
max_line_length = 100
ignore = .git,venv,.venv,target,dbt_packages

[sqlfluff:indentation]
indent_unit = space
tab_space_size = 2
indented_joins = True
indented_using_on = True

[sqlfluff:rules]
capitalisation_policy = lower
extended_capitalisation_policy = lower
comma_style = trailing

[sqlfluff:rules:L003]
hanging_indents = True

[sqlfluff:rules:L010]
capitalisation_policy = lower
```

### Common Rules

| Rule | Description | Example |
|------|-------------|---------|
| **L001** | Unnecessary trailing whitespace | `SELECT * FROM users ` |
| **L003** | Indentation errors | Inconsistent spacing |
| **L010** | Keyword capitalization | `SELECT` vs `select` |
| **L014** | Unquoted identifiers | `user_id` vs `"userId"` |
| **L019** | Comma placement | Leading vs trailing |
| **L030** | Function name capitalization | `COUNT()` vs `count()` |
| **L042** | Join without ON clause | `FROM a, b` vs `FROM a JOIN b ON ...` |

### CI/CD Integration

```yaml
# GitHub Actions
- name: Lint SQL
  run: sqlfluff lint queries/ --dialect postgres --format github-annotation
```

---

## Migration Tools

### Flyway 10.29.0

```bash
# Installation
brew install flyway       # macOS
wget -qO- https://repo1.maven.org/maven2/org/flywaydb/flyway-commandline/10.29.0/flyway-commandline-10.29.0-linux-x64.tar.gz | tar xvz

# Configuration: flyway.conf
flyway.url=jdbc:postgresql://localhost:5432/mydb
flyway.user=postgres
flyway.password=secret
flyway.locations=filesystem:migrations

# Commands
flyway migrate            # Apply pending migrations
flyway info               # Show migration status
flyway validate           # Validate applied migrations
flyway clean              # Drop all objects (WARNING: destructive)
flyway repair             # Repair metadata table

# Migration naming: V<version>__<description>.sql
# Example: V001__create_users_table.sql
```

### Liquibase 4.30.0

```bash
# Installation
brew install liquibase    # macOS
wget https://github.com/liquibase/liquibase/releases/download/v4.30.0/liquibase-4.30.0.tar.gz

# Configuration: liquibase.properties
changeLogFile=changelog.xml
url=jdbc:postgresql://localhost:5432/mydb
username=postgres
password=secret

# Commands
liquibase update          # Apply pending changesets
liquibase status          # Show pending changesets
liquibase rollback <tag>  # Rollback to specific tag
liquibase tag <name>      # Tag current database state
liquibase diff            # Compare two databases

# Changelog structure
<databaseChangeLog>
  <changeSet id="1" author="alice">
    <createTable tableName="users">
      <column name="id" type="bigint" autoIncrement="true">
        <constraints primaryKey="true"/>
      </column>
    </createTable>
  </changeSet>
</databaseChangeLog>
```

### Alembic (Python)

```bash
# Installation
pip install alembic

# Initialize
alembic init migrations

# Configuration: alembic.ini
sqlalchemy.url = postgresql://postgres:secret@localhost/mydb

# Commands
alembic revision -m "create users table"
alembic upgrade head      # Apply all pending migrations
alembic downgrade -1      # Rollback one migration
alembic current           # Show current revision
alembic history           # Show migration history
```

---

## Performance Optimization

### Index Strategy

```sql
-- B-tree index (default, for equality and range queries)
CREATE INDEX idx_users_email ON users (email);

-- Composite index (multi-column queries)
CREATE INDEX idx_users_email_status ON users (email, status);

-- Partial index (filtered subset)
CREATE INDEX idx_active_users ON users (email) WHERE status = 'active';

-- Unique index (enforce uniqueness)
CREATE UNIQUE INDEX idx_users_email_unique ON users (email);

-- GIN index (full-text search, JSONB)
CREATE INDEX idx_users_username_gin ON users USING gin (to_tsvector('english', username));

-- Concurrent index creation (no table locking)
CREATE INDEX CONCURRENTLY idx_users_created_at ON users (created_at);

-- Drop unused indexes
DROP INDEX IF EXISTS idx_old_index;
```

### Query Optimization

```sql
-- Use EXPLAIN ANALYZE to profile queries
EXPLAIN (ANALYZE, BUFFERS, VERBOSE)
SELECT * FROM users WHERE email = 'test@example.com';

-- Avoid SELECT *
-- Bad
SELECT * FROM users;
-- Good
SELECT id, email, username FROM users;

-- Use CTEs for complex queries
WITH active_users AS (
  SELECT id, email FROM users WHERE status = 'active'
)
SELECT * FROM active_users WHERE email LIKE '%@example.com';

-- Avoid N+1 queries (use JOIN instead)
-- Bad
SELECT * FROM users;
-- Then for each user: SELECT * FROM orders WHERE user_id = ?
-- Good
SELECT u.*, o.* FROM users u LEFT JOIN orders o ON u.id = o.user_id;

-- Use LIMIT for pagination
SELECT * FROM users ORDER BY created_at DESC LIMIT 20 OFFSET 0;
```

---

## TRUST 5 Principles for SQL

### T - Test First

```sql
-- Write tests before implementation (pgTAP)
-- tests/test_user_validation.sql
BEGIN;
SELECT plan(3);
SELECT has_function('validate_email');
SELECT function_returns('validate_email', 'boolean');
SELECT ok(validate_email('test@example.com'), 'Valid email passes');
SELECT * FROM finish();
ROLLBACK;
```

### R - Readable

```sql
-- Use sqlfluff for consistent formatting
sqlfluff format queries/

-- Clear naming conventions
CREATE TABLE users (
  id bigserial PRIMARY KEY,
  email varchar(255) NOT NULL,
  created_at timestamptz DEFAULT now()
);

-- Add comments for documentation
COMMENT ON TABLE users IS 'Core user accounts';
COMMENT ON COLUMN users.email IS 'User login email';
```

### U - Unified

```sql
-- Use schemas for organization
CREATE SCHEMA auth;
CREATE SCHEMA analytics;

-- Consistent naming
-- Tables: plural nouns (users, orders)
-- Columns: snake_case (user_id, created_at)
-- Indexes: idx_<table>_<columns> (idx_users_email)
-- Constraints: <table>_<column>_<type> (users_email_check)
```

### S - Secured

```sql
-- Use parameterized queries (avoid SQL injection)
-- Bad: "SELECT * FROM users WHERE id = " + user_input
-- Good: PREPARE stmt AS SELECT * FROM users WHERE id = $1;

-- Grant minimal permissions
GRANT SELECT ON users TO readonly_user;
GRANT SELECT, INSERT, UPDATE ON orders TO app_user;

-- Use RLS (Row Level Security)
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY user_isolation ON users
  USING (id = current_user_id());

-- Encrypt sensitive data
CREATE EXTENSION IF NOT EXISTS pgcrypto;
UPDATE users SET password_hash = crypt('password', gen_salt('bf'));
```

### T - Trackable

```sql
-- Use @TAG markers in migrations and tests
-- @CODE:USER-001 | SPEC: SPEC-USER-001.md | TEST: tests/test_users.sql

-- Add audit columns
CREATE TABLE users (
  id bigserial PRIMARY KEY,
  email varchar(255) NOT NULL,
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now(),
  created_by varchar(100),
  updated_by varchar(100)
);

-- Track schema changes with migrations
-- migrations/V001__create_users_table.sql
```

---

## Best Practices

### Schema Design

```sql
-- Use appropriate data types
-- Bad: VARCHAR(999999)
-- Good: VARCHAR(255), TEXT for large content

-- Normalize tables (avoid redundancy)
-- Separate concerns (users vs user_profiles)

-- Use foreign keys for referential integrity
CREATE TABLE orders (
  id bigserial PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- Add constraints for data validation
ALTER TABLE users ADD CONSTRAINT email_format_check
  CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$');
```

### Transaction Management

```sql
-- Wrap related operations in transactions
BEGIN;
INSERT INTO users (email, username) VALUES ('alice@example.com', 'alice');
INSERT INTO user_profiles (user_id, bio) VALUES (CURRVAL('users_id_seq'), 'Bio text');
COMMIT;

-- Use savepoints for partial rollback
BEGIN;
SAVEPOINT before_update;
UPDATE users SET status = 'inactive' WHERE id = 1;
ROLLBACK TO before_update;
COMMIT;

-- Set isolation level when needed
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
```

### Backup and Recovery

```bash
# Regular backups
pg_dump mydb > backup_$(date +%Y%m%d).sql

# Restore from backup
psql mydb < backup_20251022.sql

# Continuous archiving (PITR)
# Configure in postgresql.conf:
wal_level = replica
archive_mode = on
archive_command = 'cp %p /archive/%f'
```

---

## Troubleshooting

### Common Issues

```sql
-- Deadlocks
-- Solution: Ensure consistent lock ordering, use shorter transactions

-- Slow queries
-- Solution: Add indexes, use EXPLAIN ANALYZE, optimize JOINs

-- Connection limit reached
-- Solution: Close idle connections, increase max_connections

-- Disk space full
-- Solution: Run VACUUM, delete old data, increase disk size
```

### Debugging Commands

```sql
-- Show active queries
SELECT pid, query, state, query_start
FROM pg_stat_activity
WHERE state = 'active';

-- Cancel running query
SELECT pg_cancel_backend(pid);

-- Terminate connection
SELECT pg_terminate_backend(pid);

-- Show locks
SELECT * FROM pg_locks;

-- Show database size
SELECT pg_database_size('mydb');

-- Vacuum tables
VACUUM ANALYZE users;
```

---

## Additional Resources

- **PostgreSQL Docs**: https://www.postgresql.org/docs/17/
- **MySQL Docs**: https://dev.mysql.com/doc/refman/9.1/en/
- **pgTAP**: https://pgtap.org/documentation.html
- **sqlfluff**: https://docs.sqlfluff.com/
- **Flyway**: https://documentation.red-gate.com/flyway
- **Liquibase**: https://docs.liquibase.com/

---

_Last updated: 2025-10-22_
