# Advanced Database Architecture Patterns

> **Version**: 4.0.0 Enterprise
> **Last Updated**: 2025-11-22
> **Focus**: Enterprise-grade database design, sharding, replication, transactions, distributed architectures

---

## Table of Contents

1. SQL Advanced Patterns (Schema Design, Transactions)
2. NoSQL Distributed Architectures
3. Transaction Patterns and Consistency
4. Data Modeling for Scale
5. Disaster Recovery Architecture

---

## SQL Advanced Patterns

### ACID Transactions and Isolation Levels

**PostgreSQL 17 Transaction Isolation**:

```python
from sqlalchemy import event, text
from sqlalchemy.ext.asyncio import AsyncSession

class TransactionManager:
    """Manage complex ACID transactions across multiple operations"""

    async def transfer_funds(
        self,
        session: AsyncSession,
        from_account: int,
        to_account: int,
        amount: decimal
    ):
        """Atomic fund transfer with serializable isolation"""
        try:
            # Set SERIALIZABLE isolation level for strongest guarantees
            await session.execute(
                text("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
            )

            # Lock source account (SELECT FOR UPDATE)
            from_stmt = select(Account).where(Account.id == from_account)
            from_row = (await session.execute(from_stmt)).scalar()

            if from_row.balance < amount:
                raise InsufficientFundsError()

            # Deduct from source
            from_row.balance -= amount

            # Add to destination
            to_stmt = select(Account).where(Account.id == to_account)
            to_row = (await session.execute(to_stmt)).scalar()
            to_row.balance += amount

            # Create audit trail
            audit = TransactionLog(
                from_id=from_account,
                to_id=to_account,
                amount=amount,
                timestamp=datetime.utcnow()
            )
            session.add(audit)

            await session.commit()
            return True

        except Exception as e:
            await session.rollback()
            raise TransactionError(f"Transfer failed: {e}")
```

### Schema Design for Performance

**Partitioning Strategy**:

```sql
-- Time-based range partitioning for time-series data
CREATE TABLE events (
    id BIGSERIAL,
    event_type VARCHAR(50),
    user_id BIGINT,
    created_at TIMESTAMP,
    data JSONB
) PARTITION BY RANGE (DATE_TRUNC('month', created_at));

-- Create monthly partitions
CREATE TABLE events_2025_01 PARTITION OF events
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE events_2025_02 PARTITION OF events
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

-- Hash partitioning for distribution
CREATE TABLE user_sessions (
    session_id UUID,
    user_id BIGINT,
    data JSONB
) PARTITION BY HASH (user_id);

CREATE TABLE user_sessions_part_0 PARTITION OF user_sessions
    FOR VALUES WITH (MODULUS 4, REMAINDER 0);
```

### JSON/JSONB Optimization

```python
class JSONQueryBuilder:
    """Build optimized JSON queries with indexing strategies"""

    @staticmethod
    async def create_json_indexes(session: AsyncSession):
        """Create GIN indexes for JSONB searches"""
        # GIN index for general JSONB queries
        await session.execute(text(
            "CREATE INDEX idx_events_data_gin ON events USING GIN(data)"
        ))

        # Expression index for frequently accessed paths
        await session.execute(text(
            "CREATE INDEX idx_events_user_type ON events "
            "USING BTREE((data->>'user_type'), (data->>'action'))"
        ))

    @staticmethod
    async def query_nested_json(
        session: AsyncSession,
        table: str,
        path: str,
        value: str
    ):
        """Query JSONB with proper index usage"""
        # Use operators for index-efficient queries
        query = select(table).where(
            text(f"data->>'{path}' = '{value}'")
        )
        return await session.execute(query)
```

---

## NoSQL Distributed Architectures

### MongoDB Sharding Configuration

```python
class MongoDBShardingStrategy:
    """Configure and manage MongoDB sharded cluster"""

    def design_shard_key(self, data_pattern: dict) -> str:
        """Design optimal shard key based on access patterns"""
        # Good: Shard on user_id for user-scoped queries
        return {"user_id": 1}

        # Avoid: Shard on timestamp only (hot shard problem)
        # return {"created_at": 1}  # DON'T DO THIS

        # Better: Compound key for range + distribution
        return {"tenant_id": 1, "created_at": -1}

    async def setup_sharding(self, client):
        """Enable sharding on MongoDB cluster"""
        # Create shard key index (required before sharding)
        db = client["production"]
        db["users"].create_index([("user_id", 1)])

        # Enable sharding
        client.admin.command("enableSharding", "production")
        client.admin.command(
            "shardCollection",
            "production.users",
            key={"user_id": 1}
        )
```

### Redis Clustering with Slot Distribution

```python
import redis.asyncio as redis

class RedisClusterStrategy:
    """Manage Redis cluster for distributed caching"""

    async def setup_cluster_slots(self, nodes: list):
        """Configure slot distribution across cluster nodes"""
        # Connect to cluster
        cluster = redis.asyncio.RedisCluster(
            startup_nodes=[
                {"host": node["host"], "port": node["port"]}
                for node in nodes
            ],
            skip_full_coverage_check=False
        )

        # Verify slot distribution (0-16383 slots)
        slots_info = await cluster.execute_command("CLUSTER", "SLOTS")

        for slot_info in slots_info:
            start_slot = slot_info[0]
            end_slot = slot_info[1]
            print(f"Slots {start_slot}-{end_slot}: {len(slot_info)-2} replicas")

        return cluster

    async def implement_cache_warming(self, cluster, key_patterns: list):
        """Pre-load critical data into cluster"""
        pipeline = cluster.pipeline()

        for pattern in key_patterns:
            # Pre-populate high-traffic keys
            cache_key = f"cache:{pattern}"
            cache_value = await self.fetch_from_source(pattern)
            pipeline.setex(cache_key, 3600, cache_value)  # 1-hour TTL

        await pipeline.execute()
```

---

## Transaction Patterns and Consistency

### Two-Phase Commit Protocol

```python
class DistributedTransactionCoordinator:
    """Implement 2PC for transactions across multiple databases"""

    async def execute_2pc_transaction(
        self,
        primary_db: AsyncSession,
        replica_dbs: list,
        operations: list
    ):
        """Execute transaction with prepare/commit phases"""

        # Phase 1: Prepare (vote phase)
        prepare_results = []

        # Primary database prepare
        try:
            await primary_db.execute(text("BEGIN TRANSACTION"))
            for op in operations:
                await primary_db.execute(text(op))
            # Don't commit yet
            prepare_results.append(("primary", True))
        except Exception as e:
            prepare_results.append(("primary", False, str(e)))

        # Replicas prepare
        for replica in replica_dbs:
            try:
                await replica.execute(text("BEGIN TRANSACTION"))
                for op in operations:
                    await replica.execute(text(op))
                prepare_results.append(("replica", True))
            except Exception as e:
                prepare_results.append(("replica", False, str(e)))

        # Phase 2: Commit (decision phase)
        if all(result[1] for result in prepare_results):
            # All prepared successfully - commit
            await primary_db.execute(text("COMMIT"))
            for replica in replica_dbs:
                await replica.execute(text("COMMIT"))
            return {"status": "committed"}
        else:
            # Rollback on any failure
            await primary_db.execute(text("ROLLBACK"))
            for replica in replica_dbs:
                await replica.execute(text("ROLLBACK"))
            return {"status": "aborted", "failures": prepare_results}
```

### Eventual Consistency Pattern

```python
class EventualConsistencyManager:
    """Manage eventual consistency across distributed stores"""

    async def publish_write_event(
        self,
        primary_db: AsyncSession,
        event_bus,
        entity_id: str,
        changes: dict
    ):
        """Write to primary and publish event for eventual consistency"""

        # Write to primary (strong consistency)
        await primary_db.execute(
            text(f"UPDATE entities SET data = data || :changes "
                 f"WHERE id = :id"),
            {"changes": json.dumps(changes), "id": entity_id}
        )
        await primary_db.commit()

        # Publish event for asynchronous replication
        event = {
            "entity_id": entity_id,
            "changes": changes,
            "timestamp": datetime.utcnow().isoformat(),
            "version": await self.get_entity_version(primary_db, entity_id)
        }

        # Event bus (Kafka, RabbitMQ, Redis Streams)
        await event_bus.publish("entity-updates", event)

        # Secondary databases catch up asynchronously
        return event
```

---

## Data Modeling for Scale

### Time-Series Data Optimization

```sql
-- Efficient time-series table design
CREATE TABLE metrics (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    host_id BIGINT NOT NULL,
    metric_name VARCHAR(50) NOT NULL,
    value FLOAT8,
    PRIMARY KEY (time DESC, host_id, metric_name)
) PARTITION BY RANGE (time);

-- Retention policy with automatic partitioning
CREATE POLICY auto_partition_metrics
    ON metrics
    FOR SELECT
    USING (time > NOW() - INTERVAL '90 days');
```

### Document Denormalization for Performance

```python
class DenormalizationStrategy:
    """Design denormalized documents for read performance"""

    @staticmethod
    def design_user_document():
        """Denormalized user profile for single-query access"""
        return {
            "user_id": "...",
            "username": "...",
            # Embed frequently accessed relationships
            "profile": {...},
            "settings": {...},
            "roles": ["admin", "editor"],  # Denormalized from roles table
            "stats": {
                "posts_count": 42,
                "followers_count": 1000
            },
            "last_login": "...",
            "created_at": "..."
        }

    @staticmethod
    async def update_denormalized_stats(
        db: AsyncSession,
        user_id: str
    ):
        """Update denormalized counts efficiently"""
        # Update in place instead of recalculating
        await db.execute(text(
            "UPDATE users_documents SET "
            "stats = stats || :new_stats "
            "WHERE user_id = :user_id"
        ), {
            "new_stats": {
                "posts_count": 43,
                "updated_at": datetime.utcnow().isoformat()
            },
            "user_id": user_id
        })
```

---

## Disaster Recovery Architecture

### Backup and Recovery Strategy

```python
class DisasterRecoveryManager:
    """Implement comprehensive backup and recovery procedures"""

    async def create_logical_backup(self, connection_string: str):
        """Create logical backup for point-in-time recovery"""
        backup_file = f"backup_{datetime.utcnow().isoformat()}.sql"

        # PostgreSQL logical backup with parallel jobs
        command = [
            "pg_dump",
            f"postgresql://{connection_string}",
            "--format=custom",
            "--jobs=4",
            "--compress=9",
            "--file=" + backup_file
        ]

        result = await asyncio.create_subprocess_exec(*command)
        return backup_file

    async def configure_wal_archiving(self, db_connection):
        """Enable WAL archiving for PITR (Point-In-Time Recovery)"""
        # Configure archiving in PostgreSQL
        await db_connection.execute(text(
            "ALTER SYSTEM SET archive_mode = on;"
        ))
        await db_connection.execute(text(
            "ALTER SYSTEM SET archive_command = "
            "'aws s3 cp %p s3://backups/wal/%f';"
        ))

        # Reload configuration
        await db_connection.execute(text("SELECT pg_reload_conf();"))
```

### Replication Architecture

```python
class ReplicationStrategy:
    """Configure streaming replication for HA"""

    async def setup_streaming_replication(self, primary_url: str):
        """Setup PostgreSQL streaming replication"""
        # On replica, initiate replication
        replication_slot = "replica_slot_1"

        # Create replication slot for ordered WAL delivery
        async with await asyncpg.connect(primary_url) as conn:
            await conn.execute(
                f"SELECT * FROM pg_create_physical_replication_slot("
                f"'{replication_slot}', true)"
            )

        # Configure replication parameters
        replication_config = """
        primary_conninfo = 'host=primary user=replicator password=secret'
        standby_mode = 'on'
        restore_command = 'cp /archive/%f "%p" 2>/dev/null || exit 1'
        trigger_file = '/var/lib/postgresql/failover.trigger'
        """

        return replication_config
```

---

**Best Practices**:
- Design shard keys carefully to avoid hot shards
- Implement idempotent operations for retry safety
- Monitor replication lag and apply rate
- Test disaster recovery procedures regularly
- Use connection pooling across shards
- Implement circuit breakers for distributed queries

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
