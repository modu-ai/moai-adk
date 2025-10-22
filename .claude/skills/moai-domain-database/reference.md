# Database Design and Development Reference

> Official documentation and standards for database design, schema optimization, indexing, and migration management

---

## Official Documentation Links

### Relational Databases

| Database | Version | Documentation | Status |
|----------|---------|--------------|--------|
| **PostgreSQL** | 17.2 | https://www.postgresql.org/docs/ | ✅ Current (2025) |
| **MySQL** | 9.0 | https://dev.mysql.com/doc/ | ✅ Current (2025) |
| **SQLite** | 3.47 | https://www.sqlite.org/docs.html | ✅ Current (2025) |
| **SQL Server** | 2022 | https://learn.microsoft.com/en-us/sql/ | ✅ Current (2025) |

### NoSQL Databases

| Database | Version | Documentation | Status |
|----------|---------|--------------|--------|
| **MongoDB** | 8.0.0 | https://www.mongodb.com/docs/ | ✅ Current (2025) |
| **Redis** | 7.4.0 | https://redis.io/docs/ | ✅ Current (2025) |
| **Cassandra** | 5.0 | https://cassandra.apache.org/doc/ | ✅ Current (2025) |

### Standards & Guidelines

- **SQL Standard (ISO/IEC 9075)**: https://www.iso.org/standard/76583.html
- **ACID Properties**: https://en.wikipedia.org/wiki/ACID
- **CAP Theorem**: https://en.wikipedia.org/wiki/CAP_theorem

---

## Database Normalization

### Normal Forms Overview

**First Normal Form (1NF)**:
- Eliminate repeating groups
- Each column contains atomic values
- Each record is unique (primary key exists)

**Second Normal Form (2NF)**:
- Meets 1NF requirements
- All non-key attributes fully depend on primary key
- No partial dependencies

**Third Normal Form (3NF)**:
- Meets 2NF requirements
- No transitive dependencies
- Non-key attributes depend only on primary key

**Boyce-Codd Normal Form (BCNF)**:
- Stricter version of 3NF
- Every determinant is a candidate key

### Normalization Best Practices

Apply normalization up to 3NF for most applications. Queries run 23–36% slower with unnormalized structures according to DB-Engines statistics from 2024.

**When to denormalize**:
- High-read, low-write scenarios
- Data warehouse/analytics workloads
- Performance-critical queries requiring minimal joins

---

## Indexing Strategies

### Index Types

**B-Tree Indexes** (Default):
```sql
-- PostgreSQL / MySQL
CREATE INDEX idx_user_email ON users(email);

-- Composite index
CREATE INDEX idx_user_name ON users(last_name, first_name);
```

**Unique Indexes**:
```sql
CREATE UNIQUE INDEX idx_unique_email ON users(email);
```

**Partial Indexes** (PostgreSQL):
```sql
CREATE INDEX idx_active_users ON users(status) WHERE status = 'active';
```

**Full-Text Search Indexes**:
```sql
-- PostgreSQL
CREATE INDEX idx_content_search ON articles USING GIN(to_tsvector('english', content));

-- MySQL
CREATE FULLTEXT INDEX idx_content ON articles(content);
```

**Hash Indexes** (PostgreSQL):
```sql
CREATE INDEX idx_user_id_hash ON sessions USING HASH(user_id);
```

### Indexing Best Practices

1. **Index frequently filtered columns**: `WHERE`, `JOIN`, `ORDER BY`, `GROUP BY`
2. **Avoid over-indexing**: Each index slows down `INSERT`, `UPDATE`, `DELETE`
3. **Use composite indexes wisely**: Order matters (most selective first)
4. **Monitor index usage**: Remove unused indexes

**Performance Impact**:
- Lack of indexes or over-indexing are top sources of performance problems
- Index columns that are frequently filtered or joined in queries
- Balance read performance vs write overhead

---

## Schema Design Patterns

### Primary Key Strategies

**Auto-incrementing Integer**:
```sql
-- PostgreSQL
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL
);

-- MySQL
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL
);
```

**UUID** (Globally unique):
```sql
-- PostgreSQL
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL
);
```

**Composite Primary Key**:
```sql
CREATE TABLE order_items (
    order_id INT,
    product_id INT,
    quantity INT,
    PRIMARY KEY (order_id, product_id)
);
```

### Foreign Key Constraints

```sql
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

**Referential Actions**:
- `CASCADE`: Delete/update child records automatically
- `SET NULL`: Set foreign key to NULL
- `RESTRICT`: Prevent delete/update if children exist
- `NO ACTION`: Similar to RESTRICT (check at end of transaction)

### Check Constraints

```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) CHECK (price > 0),
    stock INT CHECK (stock >= 0)
);
```

---

## Database Migrations

### Migration Best Practices

1. **Version control**: Store migrations in Git
2. **One-way migrations**: Forward-only (no rollback logic)
3. **Idempotent operations**: Safe to run multiple times
4. **Test on staging**: Never test migrations directly on production
5. **Backup first**: Always backup before migrating production

### Migration Tools

| Language | Tool | Documentation |
|----------|------|--------------|
| Python | Alembic | https://alembic.sqlalchemy.org/ |
| Python | Django Migrations | https://docs.djangoproject.com/en/5.0/topics/migrations/ |
| Node.js | Knex.js | https://knexjs.org/guide/migrations.html |
| Ruby | ActiveRecord | https://guides.rubyonrails.org/active_record_migrations.html |
| Go | golang-migrate | https://github.com/golang-migrate/migrate |
| PHP | Laravel Migrations | https://laravel.com/docs/11.x/migrations |

### Safe Migration Patterns

**Adding a column** (safe):
```sql
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
```

**Adding a nullable column with default** (safe):
```sql
ALTER TABLE users ADD COLUMN verified BOOLEAN DEFAULT FALSE;
```

**Making a column NOT NULL** (requires data backfill):
```sql
-- Step 1: Add nullable column
ALTER TABLE users ADD COLUMN age INT;

-- Step 2: Backfill data
UPDATE users SET age = 0 WHERE age IS NULL;

-- Step 3: Make NOT NULL
ALTER TABLE users ALTER COLUMN age SET NOT NULL;
```

**Renaming a column** (requires code changes):
```sql
-- PostgreSQL
ALTER TABLE users RENAME COLUMN old_name TO new_name;
```

---

## Query Optimization

### EXPLAIN Analysis

```sql
-- PostgreSQL
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';

-- MySQL
EXPLAIN SELECT * FROM users WHERE email = 'test@example.com';
```

### Common Optimization Techniques

**Use indexes for filtering**:
```sql
-- Before (full table scan)
SELECT * FROM orders WHERE created_at > '2025-01-01';

-- After (create index)
CREATE INDEX idx_created_at ON orders(created_at);
```

**Avoid SELECT \***:
```sql
-- Inefficient
SELECT * FROM users WHERE id = 1;

-- Efficient (only needed columns)
SELECT id, email, name FROM users WHERE id = 1;
```

**Use LIMIT for pagination**:
```sql
-- Offset-based (slow for large offsets)
SELECT * FROM users ORDER BY id LIMIT 20 OFFSET 1000;

-- Keyset-based (faster)
SELECT * FROM users WHERE id > 1000 ORDER BY id LIMIT 20;
```

**Batch operations**:
```sql
-- Inefficient (N queries)
INSERT INTO users (name) VALUES ('Alice');
INSERT INTO users (name) VALUES ('Bob');

-- Efficient (1 query)
INSERT INTO users (name) VALUES ('Alice'), ('Bob');
```

---

## Transaction Management

### ACID Properties

- **Atomicity**: All or nothing (transactions succeed completely or fail completely)
- **Consistency**: Database remains in valid state
- **Isolation**: Concurrent transactions don't interfere
- **Durability**: Committed data survives crashes

### Isolation Levels

| Level | Dirty Read | Non-repeatable Read | Phantom Read |
|-------|-----------|---------------------|--------------|
| **Read Uncommitted** | ✅ Possible | ✅ Possible | ✅ Possible |
| **Read Committed** | ❌ Prevented | ✅ Possible | ✅ Possible |
| **Repeatable Read** | ❌ Prevented | ❌ Prevented | ✅ Possible |
| **Serializable** | ❌ Prevented | ❌ Prevented | ❌ Prevented |

### Transaction Examples

```sql
-- PostgreSQL / MySQL
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;

-- Rollback on error
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
-- If error occurs:
ROLLBACK;
```

---

## Connection Pooling

### Why Connection Pooling?

- Database connections are expensive to create/destroy
- Connection pools reuse existing connections
- Improves performance and resource utilization

### Popular Connection Pool Libraries

| Language | Library | Documentation |
|----------|---------|--------------|
| Python | SQLAlchemy Pool | https://docs.sqlalchemy.org/en/20/core/pooling.html |
| Python | asyncpg Pool | https://magicstack.github.io/asyncpg/current/api/index.html |
| Node.js | pg Pool | https://node-postgres.com/features/pooling |
| Go | pgxpool | https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool |
| Java | HikariCP | https://github.com/brettwooldridge/HikariCP |

### Connection Pool Settings

```python
# SQLAlchemy example
engine = create_engine(
    'postgresql://user:pass@localhost/db',
    pool_size=10,          # Number of persistent connections
    max_overflow=20,       # Max additional connections
    pool_timeout=30,       # Seconds to wait for connection
    pool_recycle=3600      # Recycle connections after 1 hour
)
```

---

## Security Best Practices

### SQL Injection Prevention

**Always use parameterized queries**:
```python
# ❌ BAD (vulnerable to SQL injection)
query = f"SELECT * FROM users WHERE email = '{user_input}'"

# ✅ GOOD (parameterized)
query = "SELECT * FROM users WHERE email = %s"
cursor.execute(query, (user_input,))
```

### Encryption

**Data at rest**:
- Transparent Data Encryption (TDE) for SQL Server
- Encrypted tablespaces for PostgreSQL
- InnoDB encryption for MySQL

**Data in transit**:
```
# PostgreSQL connection string
postgresql://user:pass@localhost/db?sslmode=require

# MySQL connection string
mysql://user:pass@localhost/db?ssl-mode=REQUIRED
```

### Access Control

**Principle of least privilege**:
```sql
-- Create read-only user
CREATE USER readonly_user WITH PASSWORD 'secure_password';
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly_user;

-- Create application user with limited permissions
CREATE USER app_user WITH PASSWORD 'secure_password';
GRANT SELECT, INSERT, UPDATE, DELETE ON users, orders TO app_user;
```

---

## Backup and Recovery

### Backup Strategies

**Logical Backups** (SQL dumps):
```bash
# PostgreSQL
pg_dump -U postgres -d mydb -F c -f backup.dump

# MySQL
mysqldump -u root -p mydb > backup.sql
```

**Physical Backups** (file system level):
```bash
# PostgreSQL (requires archive mode)
pg_basebackup -U postgres -D /backup/location -Ft -z -P
```

**Continuous Archiving** (point-in-time recovery):
- PostgreSQL WAL archiving
- MySQL binary logs

### Recovery Testing

**Test backups regularly**:
1. Restore to staging environment
2. Verify data integrity
3. Measure recovery time objective (RTO)
4. Document recovery procedures

---

## Performance Monitoring

### Key Metrics

- **Query execution time**: Slow query logs
- **Connection pool usage**: Active vs idle connections
- **Index usage**: Identify unused indexes
- **Cache hit ratio**: Buffer pool efficiency
- **Lock contention**: Blocking queries

### Monitoring Tools

| Tool | Database | Documentation |
|------|----------|--------------|
| **pg_stat_statements** | PostgreSQL | https://www.postgresql.org/docs/current/pgstatstatements.html |
| **Performance Schema** | MySQL | https://dev.mysql.com/doc/refman/8.0/en/performance-schema.html |
| **pgAdmin** | PostgreSQL | https://www.pgadmin.org/ |
| **MySQL Workbench** | MySQL | https://www.mysql.com/products/workbench/ |

---

## Additional Resources

### Learning Platforms

- **PostgreSQL Tutorial**: https://www.postgresqltutorial.com/
- **MySQL Tutorial**: https://www.mysqltutorial.org/
- **SQL Performance Explained**: https://use-the-index-luke.com/
- **Database Design Course**: https://www.coursera.org/learn/database-design

### Books

- "Designing Data-Intensive Applications" by Martin Kleppmann
- "SQL Performance Explained" by Markus Winand
- "High Performance MySQL" by Baron Schwartz

---

**Last Updated**: 2025-10-22
**Database Versions**: PostgreSQL 17.2, MySQL 9.0, MongoDB 8.0.0, Redis 7.4.0
**Standards**: SQL:2023, ACID compliance, CAP theorem
