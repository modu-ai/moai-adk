# Database Performance Optimization

> **Version**: 4.0.0 Enterprise
> **Last Updated**: 2025-11-22
> **Focus**: Query optimization, indexing strategies, performance tuning, scaling techniques

---

## Query Optimization Techniques

### Index Strategy Framework

```python
class IndexOptimizationStrategy:
    """Design and implement optimal indexing for performance"""

    @staticmethod
    def design_indexes(table_name: str, access_patterns: list) -> list:
        """Design indexes based on query access patterns"""
        indexes = []

        # B-tree indexes for equality and range queries
        indexes.append({
            "name": f"idx_{table_name}_created_at",
            "type": "BTREE",
            "columns": ["created_at"],
            "use_case": "Range queries on timestamps"
        })

        # Partial indexes for filtered queries
        indexes.append({
            "name": f"idx_{table_name}_active_users",
            "type": "BTREE",
            "columns": ["user_id", "email"],
            "where": "is_active = true",
            "use_case": "Filter active users only"
        })

        # BRIN indexes for large sequential tables
        indexes.append({
            "name": f"idx_{table_name}_time_brin",
            "type": "BRIN",
            "columns": ["event_time"],
            "use_case": "Time-series data with sequential writes"
        })

        # GiST or GIN for full-text search
        indexes.append({
            "name": f"idx_{table_name}_search",
            "type": "GIN",
            "columns": ["to_tsvector('english', content)"],
            "use_case": "Full-text search"
        })

        return indexes

    @staticmethod
    async def create_covering_indexes(session: AsyncSession, table: str):
        """Create covering indexes to enable index-only scans"""
        # PostgreSQL 11+ supports INCLUDE clause
        await session.execute(text(f"""
            CREATE INDEX idx_{table}_covering
            ON {table} (user_id)
            INCLUDE (email, created_at, status)
        """))

        # Verify index is used for index-only scans
        await session.execute(text(f"""
            EXPLAIN (ANALYZE, BUFFERS)
            SELECT user_id, email, status
            FROM {table}
            WHERE user_id = 123
        """))
```

### Query Execution Plan Analysis

```python
class QueryAnalyzer:
    """Analyze and optimize query execution plans"""

    async def analyze_slow_queries(self, session: AsyncSession, threshold_ms: float = 100):
        """Identify and analyze slow queries"""
        # Enable query statistics collection
        await session.execute(text("CREATE EXTENSION IF NOT EXISTS pg_stat_statements"))

        # Find slow queries
        slow_queries = await session.execute(text(f"""
            SELECT query, calls, total_time, mean_time, rows
            FROM pg_stat_statements
            WHERE mean_time > {threshold_ms}
            ORDER BY total_time DESC
            LIMIT 20
        """))

        results = []
        for query in slow_queries:
            # Analyze execution plan
            plan = await self.get_execution_plan(session, query.query)
            results.append({
                "query": query.query[:100],
                "mean_time_ms": query.mean_time,
                "calls": query.calls,
                "plan": plan,
                "recommendations": self._analyze_plan(plan)
            })

        return results

    async def get_execution_plan(self, session: AsyncSession, query: str):
        """Get detailed execution plan"""
        result = await session.execute(text(f"""
            EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)
            {query}
        """))
        return result.scalar()[0]

    @staticmethod
    def _analyze_plan(plan: dict) -> list:
        """Generate optimization recommendations"""
        recommendations = []

        # Check for sequential scans
        for node in plan.get("Plan", {}).get("Plans", []):
            if node.get("Node Type") == "Seq Scan":
                recommendations.append("Add index for sequential scan")

            # Check for nested loops
            if node.get("Node Type") == "Nested Loop":
                recommendations.append("Consider hash join or merge join")

        return recommendations
```

### Query Rewriting Patterns

```python
class QueryOptimization:
    """Rewrite queries for optimal performance"""

    @staticmethod
    async def optimize_n_plus_one_queries(session: AsyncSession):
        """Convert N+1 queries to single joined query"""
        # Inefficient: N+1 queries
        # users = session.query(User).all()
        # for user in users:
        #     posts = session.query(Post).filter(Post.user_id == user.id)

        # Optimized: Single query with joins
        query = select(User).join(Post).options(
            selectinload(User.posts)
        )
        return await session.execute(query)

    @staticmethod
    async def optimize_subqueries(session: AsyncSession):
        """Convert correlated subqueries to joins"""
        # Inefficient correlated subquery
        # SELECT * FROM users WHERE id IN (
        #     SELECT DISTINCT user_id FROM posts
        # )

        # Optimized: Use JOIN instead
        subquery = select(func.distinct(Post.user_id)).subquery()
        query = select(User).join(subquery, User.id == subquery.c.user_id)
        return await session.execute(query)
```

---

## Caching Optimization

### Multi-Level Caching Strategy

```python
class CachingStrategy:
    """Implement efficient multi-level caching"""

    async def setup_cache_layers(self, db: AsyncSession, redis_client):
        """Setup L1 (application), L2 (Redis), L3 (database)"""

        class CacheManager:
            def __init__(self, db, redis):
                self.db = db
                self.redis = redis
                self.local_cache = {}  # L1: Application memory

            async def get_user(self, user_id: int):
                # L1: Check local cache
                cache_key = f"user:{user_id}"
                if cache_key in self.local_cache:
                    return self.local_cache[cache_key]

                # L2: Check Redis
                cached = await self.redis.get(cache_key)
                if cached:
                    user = json.loads(cached)
                    self.local_cache[cache_key] = user
                    return user

                # L3: Query database
                user = await self.db.get(User, user_id)
                if user:
                    # Cache in Redis (5 min TTL)
                    await self.redis.setex(
                        cache_key, 300, json.dumps(user.to_dict())
                    )
                    self.local_cache[cache_key] = user

                return user

            async def invalidate_user(self, user_id: int):
                """Invalidate user across all cache layers"""
                cache_key = f"user:{user_id}"
                self.local_cache.pop(cache_key, None)
                await self.redis.delete(cache_key)

        return CacheManager(db, redis_client)
```

---

## Connection Pooling Optimization

### Dynamic Pool Size Configuration

```python
class ConnectionPoolOptimization:
    """Optimize connection pool parameters"""

    @staticmethod
    def calculate_optimal_pool_size(
        max_connections_db: int,
        concurrent_users: int,
        avg_query_time_ms: float = 50
    ) -> dict:
        """Calculate optimal connection pool parameters"""
        # Formula: pool_size = (concurrent_requests * query_duration) / 1000
        pool_size = max(
            5,  # minimum
            min(
                max_connections_db // 2,  # Use 50% of DB limit
                int((concurrent_users * avg_query_time_ms) / 1000)
            )
        )

        return {
            "pool_size": pool_size,
            "max_overflow": max(10, pool_size // 2),
            "pool_timeout": 30,
            "pool_recycle": 3600,
            "pool_pre_ping": True
        }

    async def setup_optimized_engine(self, db_url: str, concurrent_users: int):
        """Create engine with optimal pool configuration"""
        params = self.calculate_optimal_pool_size(
            max_connections_db=100,
            concurrent_users=concurrent_users
        )

        engine = create_async_engine(
            db_url,
            **params,
            echo=False
        )

        return engine
```

---

## Replication and Sharding Optimization

### Replication Lag Management

```python
class ReplicationMonitoring:
    """Monitor and optimize replication lag"""

    async def monitor_replication_lag(self, replica_connection):
        """Track replication lag to primary"""
        # PostgreSQL replication lag check
        lag = await replica_connection.fetchval(
            "SELECT EXTRACT(EPOCH FROM (NOW() - pg_last_xact_replay_timestamp()))"
        )

        # MySQL replication lag check
        # lag = await replica_connection.fetchval(
        #     "SELECT EXTRACT(EPOCH FROM (NOW() - max_replication_timestamp))"
        # )

        return {
            "lag_seconds": lag,
            "acceptable": lag < 5.0,  # Threshold: 5 seconds
            "status": "healthy" if lag < 5 else "degraded"
        }

    async def adaptive_read_routing(self, lag_seconds: float):
        """Route reads based on replication lag"""
        if lag_seconds < 1:
            return "replica"  # Use replica for reads
        elif lag_seconds < 5:
            return "replica_with_cache"  # Cache reads from replica
        else:
            return "primary"  # Failover to primary
```

### Shard-Aware Query Routing

```python
class ShardRouter:
    """Route queries to correct shards"""

    @staticmethod
    def determine_shard(entity_id: str, shard_count: int) -> int:
        """Consistent hashing for shard selection"""
        import hashlib
        hash_value = int(hashlib.md5(str(entity_id).encode()).hexdigest(), 16)
        return hash_value % shard_count

    async def route_query(self, entity_id: str, query: str):
        """Route query to appropriate shard"""
        shard_id = self.determine_shard(entity_id, shard_count=4)
        shard_connection = self.get_shard_connection(shard_id)

        result = await shard_connection.execute(query)
        return result

    def get_shard_connection(self, shard_id: int):
        """Get connection for specific shard"""
        shards = {
            0: "postgresql://user:pass@shard0.example.com/db",
            1: "postgresql://user:pass@shard1.example.com/db",
            2: "postgresql://user:pass@shard2.example.com/db",
            3: "postgresql://user:pass@shard3.example.com/db"
        }
        return create_async_engine(shards[shard_id])
```

---

## Best Practices

- **Index selectively**: Not every column needs an index
- **Monitor query statistics**: Use pg_stat_statements regularly
- **Test performance**: Profile under realistic load
- **Batch operations**: Use bulk insert/update for efficiency
- **Archive old data**: Partition time-series data for performance
- **Monitor replication lag**: Keep secondaries within SLA
- **Use EXPLAIN ANALYZE**: Before and after optimization
- **Denormalize strategically**: Balance normalization with read performance

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
