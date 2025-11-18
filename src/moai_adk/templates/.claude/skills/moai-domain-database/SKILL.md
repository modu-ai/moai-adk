---
name: moai-domain-database
version: 4.0.0
updated: '2025-11-18'
status: stable
description: "Enterprise database architecture specialist with PostgreSQL 17, MySQL\
  \ 8.4 LTS, \nMongoDB 8.0, Redis 7.4 expertise. Master connection pooling, query\
  \ optimization, \ncaching strategies, and database DevOps automation. Build scalable,\
  \ resilient \ndatabase systems with comprehensive monitoring and disaster recovery.\n"
allowed-tools:
- Read
- Write
- Edit
- Bash
- Glob
- WebFetch
- WebSearch
tags:
- database
- postgresql
- mysql
- mongodb
- redis
- sql
- nosql
- optimization
stability: stable
---


# Enterprise Database Architecture

## Level 1: Quick Reference

### Core Capabilities
- **Relational Databases**: PostgreSQL 17, MySQL 8.4 LTS, MariaDB 11.4
- **NoSQL Solutions**: MongoDB 8.0, Redis 7.4, Cassandra 5.0
- **ORM & Query Builders**: SQLAlchemy 2.0, Django ORM 5.1, Prisma 5
- **Connection Management**: PgBouncer 1.23, ProxySQL 2.6
- **Performance**: Query optimization, indexing strategies, caching

### Quick Setup Examples

```bash
# PostgreSQL 17 with optimized configuration
docker run -d --name postgres17 \
  -e POSTGRES_DB=myapp -e POSTGRES_USER=admin -e POSTGRES_PASSWORD=secure_pass \
  -p 5432:5432 postgres:17-alpine \
  -c shared_preload_libraries=pg_stat_statements \
  -c max_connections=200 -c shared_buffers=256MB

# Redis 7.4 with persistence
docker run -d --name redis74 \
  -p 6379:6379 redis:7.4-alpine \
  redis-server --appendonly yes --maxmemory 512mb
```

```python
# SQLAlchemy 2.0 async connection pooling
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker

engine = create_async_engine(
    "postgresql+asyncpg://user:pass@localhost/db",
    pool_size=20,
    max_overflow=30,
    pool_pre_ping=True,
    pool_recycle=3600
)

AsyncSessionLocal = sessionmaker(
    engine, class_=AsyncSession, expire_on_commit=False
)
```

## Level 2: Practical Implementation

### Database Architecture Patterns

#### 1. Connection Pool Management

```python
# PgBouncer configuration for transaction pooling
[databases]
myapp = host=localhost port=5432 dbname=myapp

[pgbouncer]
listen_port = 6432
listen_addr = 127.0.0.1
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt
logfile = /var/log/pgbouncer/pgbouncer.log
pidfile = /var/run/pgbouncer/pgbouncer.pid
admin_users = postgres
stats_users = stats, postgres

pool_mode = transaction
max_client_conn = 1000
default_pool_size = 20
min_pool_size = 5
reserve_pool_size = 5
reserve_pool_timeout = 5
max_db_connections = 50
max_user_connections = 50
server_reset_query = DISCARD ALL
server_check_delay = 30
server_check_query = select 1
server_lifetime = 3600
server_idle_timeout = 600
```

#### 2. Query Optimization Patterns

```python
# PostgreSQL 17 advanced query optimization
class QueryOptimizer:
    def __init__(self, db_session):
        self.db = db_session
    
    async def analyze_query_performance(self, query: str):
        """Analyze query execution plan and performance"""
        explain_query = f"EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) {query}"
        result = await self.db.execute(explain_query)
        plan = result.scalar()[0]
        
        return {
            'execution_time': plan['Execution Time'],
            'planning_time': plan['Planning Time'],
            'total_cost': plan['Total Cost'],
            'actual_rows': plan['Actual Rows'],
            'buffers': plan['Buffers'],
            'recommendations': self._generate_optimization_tips(plan)
        }
    
    def _generate_optimization_tips(self, plan):
        """Generate optimization recommendations based on execution plan"""
        tips = []
        
        # Check for sequential scans on large tables
        for node in self._extract_plan_nodes(plan):
            if node.get('Node Type') == 'Seq Scan' and node.get('Actual Rows', 0) > 10000:
                tips.append(f"Consider index on {node['Relation Name']}")
            
            if node.get('Node Type') == 'Hash Join' and node.get('Hash Buckets', 0) > 100000:
                tips.append("Large hash join detected - consider work_mem optimization")
        
        return tips
    
    async def create_missing_indexes(self, table_name: str, columns: list):
        """Automatically create missing indexes based on query patterns"""
        index_name = f"idx_{table_name}_{'_'.join(columns)}"
        
        index_sql = f"""
        CREATE INDEX CONCURRENTLY IF NOT EXISTS {index_name} 
        ON {table_name} ({', '.join(columns)})
        """
        
        await self.db.execute(text(index_sql))
        await self.db.commit()
```

#### 3. Caching Strategy Implementation

```python
# Redis 7.4 advanced caching patterns
class DatabaseCache:
    def __init__(self, redis_client, db_session):
        self.redis = redis_client
        self.db = db_session
    
    async def cached_query(self, key: str, query: str, ttl: int = 3600):
        """Execute query with Redis caching"""
        # Try cache first
        cached_result = await self.redis.get(key)
        if cached_result:
            return json.loads(cached_result)
        
        # Execute query and cache result
        result = await self.db.execute(text(query))
        rows = result.fetchall()
        
        # Convert to JSON-serializable format
        data = [dict(row._mapping) for row in rows]
        
        # Cache with TTL
        await self.redis.setex(key, ttl, json.dumps(data))
        
        return data
    
    async def invalidate_cache_pattern(self, pattern: str):
        """Invalidate cache keys matching pattern"""
        keys = await self.redis.keys(pattern)
        if keys:
            await self.redis.delete(*keys)
    
    async def cache_warm_up(self, queries: list):
        """Pre-warm cache with common queries"""
        for key, query, ttl in queries:
            await self.cached_query(key, query, ttl)
```

### Database Monitoring & Observability

```python
# Comprehensive database monitoring setup
import psycopg2
from prometheus_client import Counter, Histogram, Gauge
import time

# Prometheus metrics
DB_QUERY_DURATION = Histogram('db_query_duration_seconds', 'Database query duration')
DB_CONNECTION_POOL = Gauge('db_connection_pool_active', 'Active database connections')
DB_ERROR_COUNT = Counter('db_error_count', 'Database error count', ['error_type'])

class DatabaseMonitor:
    def __init__(self, connection_string):
        self.connection_string = connection_string
    
    @DB_QUERY_DURATION.time()
    async def execute_with_monitoring(self, query: str, params: dict = None):
        """Execute query with comprehensive monitoring"""
        start_time = time.time()
        
        try:
            async with self.pool.acquire() as conn:
                result = await conn.execute(query, params or {})
                
                # Log slow queries
                duration = time.time() - start_time
                if duration > 1.0:  # Slow query threshold
                    await self._log_slow_query(query, duration, params)
                
                return result
                
        except Exception as e:
            DB_ERROR_COUNT.labels(error_type=type(e).__name__).inc()
            await self._log_database_error(query, e, params)
            raise
    
    async def get_database_metrics(self):
        """Collect comprehensive database metrics"""
        metrics = {}
        
        # Connection pool metrics
        pool_stats = await self._get_pool_stats()
        metrics.update(pool_stats)
        
        # Database performance metrics
        perf_stats = await self._get_performance_stats()
        metrics.update(perf_stats)
        
        # Lock monitoring
        lock_stats = await self._get_lock_stats()
        metrics.update(lock_stats)
        
        return metrics
```

## Level 3: Advanced Integration

### Database DevOps & Automation

#### 1. Database Migration as Code

```python
# Database migration management with Alembic
from alembic import command
from alembic.config import Config
import asyncio

class DatabaseMigrator:
    def __init__(self, database_url: str):
        self.database_url = database_url
        self.alembic_cfg = Config("alembic.ini")
        self.alembic_cfg.set_main_option("sqlalchemy.url", database_url)
    
    async def create_migration(self, message: str):
        """Create new database migration"""
        command.revision(self.alembic_cfg, autogenerate=True, message=message)
    
    async def run_migrations(self):
        """Run pending migrations"""
        command.upgrade(self.alembic_cfg, "head")
    
    async def rollback_migration(self, revision: str):
        """Rollback to specific revision"""
        command.downgrade(self.alembic_cfg, revision)
    
    async def get_migration_history(self):
        """Get migration history"""
        from alembic.runtime.migration import MigrationContext
        from sqlalchemy import create_engine
        
        engine = create_engine(self.database_url)
        with engine.connect() as connection:
            context = MigrationContext.configure(connection)
            return context.get_current_revision()
```

#### 2. Multi-Database Replication Setup

```python
# PostgreSQL logical replication setup
class DatabaseReplication:
    def __init__(self, primary_config, replica_configs):
        self.primary = primary_config
        self.replicas = replica_configs
    
    async def setup_logical_replication(self):
        """Setup logical replication from primary to replicas"""
        
        # Setup replication user on primary
        await self._create_replication_user()
        
        # Create publication on primary
        await self._create_publication()
        
        # Setup subscriptions on replicas
        for replica in self.replicas:
            await self._create_subscription(replica)
    
    async def _create_replication_user(self):
        """Create replication user on primary database"""
        sql = """
        CREATE USER IF NOT EXISTS replicator WITH REPLICATION ENCRYPTED PASSWORD 'replicator_password';
        GRANT CONNECT ON DATABASE {database} TO replicator;
        GRANT USAGE ON SCHEMA public TO replicator;
        GRANT SELECT ON ALL TABLES IN SCHEMA public TO replicator;
        ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO replicator;
        """.format(database=self.primary['database'])
        
        await self._execute_on_primary(sql)
    
    async def _create_publication(self):
        """Create publication for all tables"""
        sql = """
        CREATE PUBLICATION IF NOT EXISTS all_tables_publication 
        FOR ALL TABLES WITH (publish = 'insert, update, delete');
        """
        
        await self._execute_on_primary(sql)
    
    async def _create_subscription(self, replica_config):
        """Create subscription on replica"""
        sql = f"""
        CREATE SUBSCRIPTION IF NOT EXISTS {replica_config['name']}_subscription
        CONNECTION 'host={self.primary['host']} port={self.primary['port']} 
                    dbname={self.primary['database']} user=replicator 
                    password=replicator_password'
        PUBLICATION all_tables_publication
        WITH (slot_name = {replica_config['name']}_slot, 
              create_slot = true, 
              enabled = true);
        """
        
        await self._execute_on_replica(replica_config, sql)
```

### Database Security & Compliance

#### 1. Row-Level Security Implementation

```sql
-- PostgreSQL Row-Level Security (RLS)
-- Enable RLS on sensitive tables
ALTER TABLE sensitive_data ENABLE ROW LEVEL SECURITY;

-- Create policy for user access based on department
CREATE POLICY user_department_access ON sensitive_data
    FOR ALL
    TO application_user
    USING (department = current_setting('app.current_department'));

-- Create policy for admin access
CREATE POLICY admin_full_access ON sensitive_data
    FOR ALL
    TO admin_user
    USING (true);

-- Audit logging for data access
CREATE OR REPLACE FUNCTION audit_sensitive_access()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO audit_log (table_name, operation, user_id, timestamp, row_data)
    VALUES (
        TG_TABLE_NAME,
        TG_OP,
        current_setting('app.current_user_id'),
        NOW(),
        row_to_json(NEW)
    );
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create audit trigger
CREATE TRIGGER sensitive_data_audit
    AFTER INSERT OR UPDATE OR DELETE ON sensitive_data
    FOR EACH ROW EXECUTE FUNCTION audit_sensitive_access();
```

## Related Skills

- **moai-domain-backend**: Microservices database patterns
- **moai-domain-testing**: Database testing strategies  
- **moai-domain-ml-ops**: ML model database integration
- **moai-domain-devops**: Database infrastructure as code

## Quick Start Checklist

- [ ] Select appropriate database engine (PostgreSQL 17/MySQL 8.4/MongoDB 8.0)
- [ ] Configure connection pooling (PgBouncer/ProxySQL)
- [ ] Implement query monitoring and optimization
- [ ] Setup caching strategy with Redis 7.4
- [ ] Configure database backups and replication
- [ ] Implement security measures and RLS
- [ ] Setup monitoring and alerting
- [ ] Create disaster recovery procedures

## Performance Optimization Tips

1. **Connection Pooling**: Always use connection pools with appropriate sizing
2. **Query Optimization**: Use EXPLAIN ANALYZE for slow query analysis
3. **Index Strategy**: Create indexes based on actual query patterns
4. **Caching Layers**: Implement Redis caching for frequently accessed data
5. **Monitoring**: Set up comprehensive monitoring for all database metrics
6. **Security**: Use row-level security and audit logging for sensitive data
7. **Backups**: Implement automated backup with point-in-time recovery
8. **Replication**: Use logical replication for high availability

---

**Enterprise Database Architecture** - Build scalable, secure, and high-performance database systems with modern best practices and comprehensive automation.
