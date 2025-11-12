# moai-domain-database - CLI Reference

_Last updated: 2025-11-12_

## Stable Version Reference (2025-11)

### Database Systems

**PostgreSQL 17** (Stable):
- Release: 2024-09 (LTS)
- Features: JSON improvements, incremental backup, logical replication enhancements
- Install: `brew install postgresql@17` (macOS), `apt install postgresql-17` (Ubuntu)

**MySQL 8.4 LTS** (Long-Term Support):
- Release: 2024-04 (LTS until 2032)
- Features: InnoDB optimizations, JSON improvements, performance schema enhancements
- Install: `brew install mysql@8.4` (macOS), `apt install mysql-server-8.4` (Ubuntu)

**MongoDB 8.0** (Stable):
- Release: 2024 (Stable)
- Features: Queryable Encryption GA, Time Series improvements, cluster-to-cluster sync
- Install: `brew install mongodb-community@8.0` (macOS)

**Redis 7.4** (Stable):
- Release: 2024 (Stable)
- Features: JSON support improvements, Stream optimizations, Cluster mode enhancements
- Install: `brew install redis@7.4` (macOS), `apt install redis-server` (Ubuntu)

### ORM & Query Builders

**SQLAlchemy 2.0** (Python, Stable):
- Release: 2023-01 (Stable)
- Install: `pip install sqlalchemy[asyncio]>=2.0.0`
- Async Driver: `pip install asyncpg` (PostgreSQL), `pip install aiomysql` (MySQL)

**Prisma 5** (TypeScript, Stable):
- Release: 2023 (Stable)
- Install: `npm install prisma@^5.0.0 @prisma/client@^5.0.0`
- Usage: `npx prisma init`, `npx prisma generate`

**Django ORM 5.1** (Python, LTS):
- Release: 2024-08 (LTS until 2026)
- Install: `pip install django>=5.1.0`

**TypeORM 0.3** (TypeScript, Stable):
- Release: 2022 (Stable)
- Install: `npm install typeorm@^0.3.0 reflect-metadata`

---

## Common Commands

### PostgreSQL 17

```bash
# Start PostgreSQL service
brew services start postgresql@17

# Connect to database
psql -U postgres -d mydb

# Create database
createdb mydb

# Backup database
pg_dump mydb > backup.sql

# Restore database
psql mydb < backup.sql

# Check version
psql --version
```

### MySQL 8.4

```bash
# Start MySQL service
brew services start mysql@8.4

# Connect to database
mysql -u root -p

# Create database
mysql -e "CREATE DATABASE mydb;"

# Backup database
mysqldump -u root -p mydb > backup.sql

# Restore database
mysql -u root -p mydb < backup.sql

# Check version
mysql --version
```

### Redis 7.4

```bash
# Start Redis service
brew services start redis

# Connect to Redis CLI
redis-cli

# Monitor commands
redis-cli MONITOR

# Get server info
redis-cli INFO

# Check version
redis-cli --version
```

### MongoDB 8.0

```bash
# Start MongoDB service
brew services start mongodb-community@8.0

# Connect to MongoDB shell
mongosh

# Create database
mongosh --eval "use mydb"

# Backup database
mongodump --db mydb --out /backup

# Restore database
mongorestore /backup/mydb

# Check version
mongod --version
```

---

## Performance Tuning

### PostgreSQL Configuration (`postgresql.conf`)

```ini
# Memory Settings
shared_buffers = 256MB           # 25% of RAM
effective_cache_size = 1GB       # 50-75% of RAM
work_mem = 4MB                   # Per operation
maintenance_work_mem = 64MB      # For VACUUM, CREATE INDEX

# Connection Settings
max_connections = 100
superuser_reserved_connections = 3

# Query Planner
random_page_cost = 1.1           # SSD: 1.1, HDD: 4.0
effective_io_concurrency = 200   # SSD: 200, HDD: 2

# Write-Ahead Log
wal_buffers = 16MB
checkpoint_completion_target = 0.9
```

### MySQL Configuration (`my.cnf`)

```ini
# InnoDB Settings
innodb_buffer_pool_size = 1G     # 50-80% of RAM
innodb_log_file_size = 256M
innodb_flush_log_at_trx_commit = 2
innodb_flush_method = O_DIRECT

# Connection Settings
max_connections = 150
thread_cache_size = 8

# Query Cache (disabled in 8.0+)
# query_cache_type = OFF
```

### Redis Configuration (`redis.conf`)

```ini
# Memory Settings
maxmemory 256mb
maxmemory-policy allkeys-lru

# Persistence
save 900 1                       # Save if 1 key changed in 900s
save 300 10                      # Save if 10 keys changed in 300s
save 60 10000                    # Save if 10000 keys changed in 60s
appendonly yes
appendfsync everysec
```

---

## Migration Tools

### Alembic (SQLAlchemy)

```bash
# Install Alembic
pip install alembic

# Initialize Alembic
alembic init alembic

# Create migration
alembic revision --autogenerate -m "Add users table"

# Apply migrations
alembic upgrade head

# Rollback migration
alembic downgrade -1

# Check current version
alembic current
```

### Prisma Migrate

```bash
# Create migration
npx prisma migrate dev --name add_users_table

# Apply migrations (production)
npx prisma migrate deploy

# Reset database (development)
npx prisma migrate reset

# Check migration status
npx prisma migrate status
```

---

## Monitoring & Debugging

### PostgreSQL

```sql
-- Active queries
SELECT pid, usename, query, state
FROM pg_stat_activity
WHERE state != 'idle';

-- Slow queries
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- Table sizes
SELECT
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
LIMIT 10;

-- Index usage
SELECT
  schemaname,
  tablename,
  indexname,
  idx_scan,
  idx_tup_read,
  idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan ASC
LIMIT 10;
```

### MySQL

```sql
-- Active queries
SHOW FULL PROCESSLIST;

-- Slow queries (enable slow query log first)
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 1;

-- Table sizes
SELECT
  table_schema,
  table_name,
  ROUND(data_length / 1024 / 1024, 2) AS data_mb,
  ROUND(index_length / 1024 / 1024, 2) AS index_mb
FROM information_schema.tables
ORDER BY data_length + index_length DESC
LIMIT 10;

-- Index usage
SELECT
  database_name,
  table_name,
  index_name,
  INDEX_TYPE
FROM mysql.innodb_index_stats
ORDER BY n_diff_pfx01 DESC;
```

### Redis

```bash
# Monitor real-time commands
redis-cli MONITOR

# Get server info
redis-cli INFO

# Memory usage
redis-cli INFO memory

# Slow log
redis-cli SLOWLOG GET 10

# Client list
redis-cli CLIENT LIST

# Database size
redis-cli DBSIZE
```

---

## Tool Versions (2025-11-12 Stable)

- **PostgreSQL**: 17.5 (LTS)
- **MySQL**: 8.4.3 LTS (LTS until 2032)
- **MongoDB**: 8.0.4 (Stable)
- **Redis**: 7.4.1 (Stable)
- **SQLAlchemy**: 2.0.35 (Stable)
- **Prisma**: 5.22.0 (Stable)
- **Alembic**: 1.13.3 (Stable)
- **asyncpg**: 0.29.0 (PostgreSQL async driver)
- **aiomysql**: 0.2.0 (MySQL async driver)

---

_For detailed usage, see SKILL.md_

## Skill 표준 준수 검증

이 Skill의 Enterprise v4.0 준수 여부를 확인하세요:

### 빠른 검증 (YAML 메타데이터만)

```bash
python3 -c "
import yaml
import sys
try:
    with open('SKILL.md') as f:
        content = f.read()
        yaml_str = content.split('---')[1]
        metadata = yaml.safe_load(yaml_str)
        required = ['name', 'version', 'status', 'description']
        missing = [f for f in required if f not in metadata]
        if not missing:
            print('✅ PASS: YAML metadata complete')
            sys.exit(0)
        else:
            print(f'❌ FAIL: Missing fields: {missing}')
            sys.exit(1)
except Exception as e:
    print(f'❌ ERROR: {str(e)[:100]}')
    sys.exit(1)
"
```

### 전체 검증 (moai-skill-validator 사용)

완전한 Skill 검증을 위해 validator를 호출하세요:

```
Skill("moai-skill-validator")
```

이 명령어는 다음을 검증합니다:
- YAML 메타데이터 구조
- 필수 파일 존재 (SKILL.md, reference.md, examples.md)
- Progressive Disclosure 구조
- 보안 검증 (API 키, eval/exec 패턴 감지)
- TAG 시스템 준수
