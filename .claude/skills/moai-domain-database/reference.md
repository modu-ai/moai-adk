# Database Architecture Reference Guide

## Advanced Database Patterns

### 1. High Availability Setup

#### PostgreSQL Patroni Cluster Configuration

```yaml
# patroni.yml
restapi:
  listen: 0.0.0.0:8008
  connect_address: 10.0.0.1:8008

etcd:
  hosts: 10.0.0.10:2379,10.0.0.11:2379,10.0.0.12:2379

bootstrap:
  dcs:
    ttl: 30
    loop_wait: 10
    retry_timeout: 10
    maximum_lag_on_failover: 1048576
    postgresql:
      use_pg_rewind: true
      parameters:
        max_connections: 200
        max_wal_senders: 10
        wal_level: replica
        hot_standby: "on"
        max_prepared_transactions: 200
        max_worker_processes: 8
        max_replication_slots: 10
        wal_keep_size: 1GB

  pg_hba:
    - host replication replicator 10.0.0.0/8 md5
    - host all all 0.0.0.0/0 md5

postgresql:
  listen: 0.0.0.0:5432
  connect_address: 10.0.0.1:5432
  data_dir: /var/lib/postgresql/data
  bin_dir: /usr/lib/postgresql/17/bin
  pgpass: /tmp/pgpass
  authentication:
    replication:
      username: replicator
      password: replicator_password
    superuser:
      username: postgres
      password: postgres_password
    rewind:
      username: rewind_user
      password: rewind_password

tags:
  nofailover: false
  noloadbalance: false
  clonefrom: false
  nosync: false
```

#### MongoDB Replica Set Configuration

```javascript
// MongoDB replica set initialization
rs.initiate({
  _id: "rs0",
  members: [
    {
      _id: 0,
      host: "mongo1:27017",
      priority: 2,
      votes: 1
    },
    {
      _id: 1,
      host: "mongo2:27017",
      priority: 1,
      votes: 1
    },
    {
      _id: 2,
      host: "mongo3:27017",
      priority: 1,
      votes: 1,
      arbiterOnly: false
    }
  ]
});

// Enable sharding for large datasets
sh.enableSharding("myapp");
sh.shardCollection("myapp.users", {"_id": "hashed"});
sh.shardCollection("myapp.events", {"timestamp": 1, "user_id": 1});
```

### 2. Performance Tuning

#### PostgreSQL Configuration Optimization

```ini
# postgresql.conf performance tuning
# Memory Configuration
shared_buffers = 4GB                    # 25% of RAM
effective_cache_size = 12GB             # 75% of RAM
work_mem = 256MB                        # Per operation memory
maintenance_work_mem = 1GB              # Maintenance operations
autovacuum_work_mem = 512MB             # Autovacuum memory

# Connection Settings
max_connections = 200
max_prepared_transactions = 200
superuser_reserved_connections = 3

# WAL Configuration
wal_buffers = 64MB                      # 16MB or 3% of shared_buffers
wal_writer_delay = 200ms
commit_delay = 0
commit_siblings = 5
checkpoint_segments = 32                # WAL segments before checkpoint
checkpoint_timeout = 15min
checkpoint_completion_target = 0.9

# Query Planning
random_page_cost = 1.1                  # SSD optimization
effective_io_concurrency = 200         # SSD concurrent IO
seq_page_cost = 1.0
cpu_tuple_cost = 0.01
cpu_index_tuple_cost = 0.005
cpu_operator_cost = 0.0025

# Parallel Query Settings
max_parallel_workers_per_gather = 4
max_parallel_workers = 8
parallel_tuple_cost = 0.1
parallel_setup_cost = 1000.0

# Background Writer
bgwriter_delay = 200ms
bgwriter_lru_maxpages = 100
bgwriter_lru_multiplier = 2.0

# Statistics
default_statistics_target = 1000
track_activities = on
track_counts = on
track_io_timing = on
track_functions = all
track_activity_query_size = 2048
```

#### Redis Performance Optimization

```conf
# redis.conf performance tuning
# Memory Management
maxmemory 8gb
maxmemory-policy allkeys-lru
maxmemory-samples 5

# Persistence Settings
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes

# AOF Configuration
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
no-appendfsync-on-rewrite no
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb

# Network Settings
tcp-keepalive 300
tcp-backlog 511
timeout 0

# Client Settings
maxclients 10000

# Memory Optimization
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
list-max-ziplist-size -2
list-compress-depth 0
set-max-intset-entries 512
zset-max-ziplist-entries 128
zset-max-ziplist-value 64

# Lazy Freeing
lazyfree-lazy-eviction no
lazyfree-lazy-expire no
lazyfree-lazy-server-del no
slave-lazy-flush no
```

### 3. Advanced Indexing Strategies

#### Partial and Functional Indexes

```sql
-- PostgreSQL partial indexes for performance
CREATE INDEX idx_active_users_email ON users(email) WHERE status = 'active';
CREATE INDEX idx_recent_orders ON orders(created_at) WHERE created_at > NOW() - INTERVAL '30 days';
CREATE INDEX idx_high_value_transactions ON transactions(amount) WHERE amount > 10000;

-- Functional indexes for computed columns
CREATE INDEX idx_users_lower_email ON users(LOWER(email));
CREATE INDEX idx_orders_total_amount ON orders((quantity * unit_price));
CREATE INDEX idx_users_full_name ON users((first_name || ' ' || last_name));

-- Expression indexes for complex queries
CREATE INDEX idx_events_date_hour ON events(date_trunc('hour', created_at));
CREATE INDEX idx_users_age_bucket ON users(EXTRACT(YEAR FROM AGE(NOW(), birth_date)) / 10);
```

#### MongoDB Compound Indexing

```javascript
// MongoDB compound indexes for common query patterns
db.users.createIndex({"status": 1, "created_at": -1});
db.orders.createIndex({"customer_id": 1, "status": 1, "created_at": -1});
db.products.createIndex({"category": 1, "price": 1, "rating": -1});

// Text indexes for search functionality
db.products.createIndex({
  "name": "text",
  "description": "text",
  "tags": "text"
}, {
  weights: {
    "name": 10,
    "description": 5,
    "tags": 8
  },
  name: "product_search_index"
});

// Geospatial indexes for location-based queries
db.locations.createIndex({"coordinates": "2dsphere"});
db.nearby.createIndex({"location": "2dsphere", "category": 1});
```

### 4. Backup and Recovery Strategies

#### PostgreSQL Backup Automation

```bash
#!/bin/bash
# backup_postgresql.sh - Comprehensive backup script

DB_NAME="myapp"
DB_USER="postgres"
BACKUP_DIR="/backups/postgresql"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# Create backup directory
mkdir -p "$BACKUP_DIR/$DATE"

# Full database backup
pg_dump -h localhost -U "$DB_USER" -d "$DB_NAME" \
  --format=custom \
  --compress=9 \
  --verbose \
  --file="$BACKUP_DIR/$DATE/full_backup.dump"

# Schema-only backup
pg_dump -h localhost -U "$DB_USER" -d "$DB_NAME" \
  --schema-only \
  --file="$BACKUP_DIR/$DATE/schema_backup.sql"

# Data-only backup
pg_dump -h localhost -U "$DB_USER" -d "$DB_NAME" \
  --data-only \
  --file="$BACKUP_DIR/$DATE/data_backup.sql"

# Configuration backup
pg_dumpall -h localhost -U "$DB_USER" \
  --roles-only \
  --file="$BACKUP_DIR/$DATE/roles_backup.sql"

# WAL segment backup for point-in-time recovery
pg_receivewal -h localhost -U "$DB_USER" \
  --directory="$BACKUP_DIR/$DATE/wal" \
  --compress=9 \
  --status-interval=1

# Verify backup integrity
pg_restore --list "$BACKUP_DIR/$DATE/full_backup.dump" > "$BACKUP_DIR/$DATE/backup_contents.txt"

# Clean old backups
find "$BACKUP_DIR" -type d -mtime +$RETENTION_DAYS -exec rm -rf {} \;

echo "Backup completed: $BACKUP_DIR/$DATE"
```

#### MongoDB Backup Strategy

```bash
#!/bin/bash
# backup_mongodb.sh - MongoDB backup automation

DB_NAME="myapp"
BACKUP_DIR="/backups/mongodb"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# Create backup directory
mkdir -p "$BACKUP_DIR/$DATE"

# Full database backup
mongodump --db "$DB_NAME" \
  --out "$BACKUP_DIR/$DATE" \
  --gzip \
  --journal

# Oplog backup for point-in-time recovery
mongodump --db local --collection oplog.rs \
  --query '{"ts":{"$gt":{"$timestamp":{"t":1640995200,"i":1}}}}' \
  --out "$BACKUP_DIR/$DATE/oplog" \
  --gzip

# User and role backup
mongodump --db admin --collection system.users \
  --out "$BACKUP_DIR/$DATE/users" \
  --gzip

# Index definitions backup
mongo "$DB_NAME" --eval "
  var collections = db.getCollectionNames();
  for (var i = 0; i < collections.length; i++) {
    var indexes = db.getCollection(collections[i]).getIndexes();
    print('Collection: ' + collections[i]);
    printjson(indexes);
  }
" > "$BACKUP_DIR/$DATE/index_definitions.json"

# Verify backup
ls -la "$BACKUP_DIR/$DATE"

# Clean old backups
find "$BACKUP_DIR" -type d -mtime +$RETENTION_DAYS -exec rm -rf {} \;

echo "MongoDB backup completed: $BACKUP_DIR/$DATE"
```

### 5. Monitoring and Alerting

#### Prometheus Database Metrics

```yaml
# prometheus.yml database scraping configuration
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'postgresql'
    static_configs:
      - targets: ['postgres-exporter:9187']
    scrape_interval: 10s
    metrics_path: /metrics

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
    scrape_interval: 10s

  - job_name: 'mongodb'
    static_configs:
      - targets: ['mongodb-exporter:9216']
    scrape_interval: 10s

rule_files:
  - "database_alerts.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

```yaml
# database_alerts.yml - Database alerting rules
groups:
  - name: postgresql_alerts
    rules:
      - alert: PostgreSQLDown
        expr: up{job="postgresql"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "PostgreSQL is down"
          description: "PostgreSQL instance {{ $labels.instance }} has been down for more than 1 minute."

      - alert: PostgreSQLHighConnections
        expr: pg_stat_activity_count / pg_settings_max_connections > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High PostgreSQL connection usage"
          description: "PostgreSQL connection usage is above 80% on instance {{ $labels.instance }}."

      - alert: PostgreSQLSlowQueries
        expr: rate(pg_stat_statements_mean_time_seconds[5m]) > 1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Slow PostgreSQL queries detected"
          description: "Average query time is {{ $value }}s on instance {{ $labels.instance }}."

  - name: redis_alerts
    rules:
      - alert: RedisDown
        expr: up{job="redis"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Redis is down"
          description: "Redis instance {{ $labels.instance }} has been down for more than 1 minute."

      - alert: RedisHighMemoryUsage
        expr: redis_memory_used_bytes / redis_memory_max_bytes > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High Redis memory usage"
          description: "Redis memory usage is above 90% on instance {{ $labels.instance }}."
```

## Database Security Best Practices

### Encryption and Authentication

```sql
-- PostgreSQL SSL/TLS Configuration
-- Require SSL for all connections
ALTER SYSTEM SET ssl = 'on';
ALTER SYSTEM SET ssl_cert_file = '/var/lib/postgresql/ssl/server.crt';
ALTER SYSTEM SET ssl_key_file = '/var/lib/postgresql/ssl/server.key';
ALTER SYSTEM SET ssl_ca_file = '/var/lib/postgresql/ssl/ca.crt';

-- Create users with password policies
CREATE ROLE app_user WITH LOGIN PASSWORD 'secure_password' VALID UNTIL '2025-12-31';
CREATE ROLE read_only_user WITH LOGIN PASSWORD 'readonly_password' 
  VALID UNTIL '2025-12-31';

-- Implement password complexity
CREATE EXTENSION IF NOT EXISTS passwordcheck;
ALTER SYSTEM SET password_encryption = 'scram-sha-256';

-- Connection security
CREATE USER app_user WITH PASSWORD 'very_secure_password';
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT CONNECT ON DATABASE myapp TO app_user;
GRANT USAGE ON SCHEMA public TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user;
```

```javascript
// MongoDB security configuration
// Enable authentication and authorization
use admin
db.createUser({
  user: "admin",
  pwd: "secure_admin_password",
  roles: [
    { role: "userAdminAnyDatabase", db: "admin" },
    { role: "clusterAdmin", db: "admin" },
    { role: "readAnyDatabase", db: "admin" }
  ]
});

// Enable SSL/TLS
net:
  ssl:
    mode: requireSSL
    PEMKeyFile: /etc/ssl/mongodb.pem
    CAFile: /etc/ssl/ca.pem
    allowInvalidCertificates: false
    allowInvalidHostnames: false

// Enable encryption at rest
security:
  enableEncryption: true
  encryptionKeyFile: /etc/mongodb-keyfile
  kmip:
    clientCertificateFile: /etc/ssl/kmip-client.pem
    serverCAFile: /etc/ssl/kmip-ca.pem
```

## Database Migration Patterns

### Zero-Downtime Migration Strategy

```python
# Zero-downtime database migration
class ZeroDowntimeMigration:
    def __init__(self, source_db, target_db):
        self.source = source_db
        self.target = target_db
    
    async def migrate_table(self, table_name, batch_size=1000):
        """Migrate table with zero downtime"""
        
        # Step 1: Create target table with new schema
        await self._create_target_table(table_name)
        
        # Step 2: Set up change data capture
        cdc_cursor = await self._setup_cdc(table_name)
        
        # Step 3: Migrate historical data in batches
        offset = 0
        while True:
            batch = await self._get_batch(table_name, offset, batch_size)
            if not batch:
                break
                
            await self._insert_batch(self.target, table_name, batch)
            offset += batch_size
            
            # Process CDC changes periodically
            if offset % (batch_size * 10) == 0:
                await self._process_cdc_changes(cdc_cursor)
        
        # Step 4: Final sync and cutover
        await self._final_sync(cdc_cursor)
        await self._switch_traffic(table_name)
        
        # Step 5: Cleanup
        await self._cleanup(table_name)
    
    async def _setup_cdc(self, table_name):
        """Setup change data capture for table"""
        # Implementation depends on database type
        # PostgreSQL: Use logical replication
        # MySQL: Use binlog
        # MongoDB: Use change streams
        pass
    
    async def _switch_traffic(self, table_name):
        """Switch application traffic to new table"""
        # Use feature flags or configuration changes
        # Monitor for errors and rollback if needed
        pass
```

This reference guide provides comprehensive database architecture patterns, performance optimization strategies, backup procedures, and security implementations for enterprise environments.
