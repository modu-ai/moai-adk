---
name: moai-domain-database
description: Database specialist covering PostgreSQL, MongoDB, Redis, and advanced data patterns for modern applications
version: 1.0.0
category: domain
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
tags:
 - database
 - postgresql
 - mongodb
 - redis
 - data-patterns
 - performance
updated: 2025-12-06
status: active
author: MoAI-ADK Team
---

# Database Domain Specialist

## Quick Reference (30 seconds)

Enterprise Database Expertise - Comprehensive database patterns and implementations covering PostgreSQL, MongoDB, Redis, and advanced data management for scalable modern applications.

Core Capabilities:
- PostgreSQL: Advanced relational patterns, optimization, and scaling
- MongoDB: Document modeling, aggregation, and NoSQL performance tuning
- Redis: In-memory caching, real-time analytics, and distributed systems
- Multi-Database: Hybrid architectures and data integration patterns
- Performance: Query optimization, indexing strategies, and scaling
- Operations: Connection management, migrations, and monitoring

When to Use:
- Designing database schemas and data models
- Implementing caching strategies and performance optimization
- Building scalable data architectures
- Working with multi-database systems
- Optimizing database queries and performance

---

## Implementation Guide (5 minutes)

### Quick Start Workflow

Database Stack Initialization:
```python
from moai_domain_database import DatabaseManager

# Initialize multi-database stack
db_manager = DatabaseManager()

# Configure PostgreSQL for relational data
postgresql = db_manager.setup_postgresql(
 connection_string="postgresql://...",
 connection_pool_size=20,
 enable_query_logging=True
)

# Configure MongoDB for document storage
mongodb = db_manager.setup_mongodb(
 connection_string="mongodb://...",
 database_name="app_data",
 enable_sharding=True
)

# Configure Redis for caching and real-time features
redis = db_manager.setup_redis(
 connection_string="redis://...",
 max_connections=50,
 enable_clustering=True
)

# Use unified database interface
user_data = db_manager.get_user_with_profile(user_id)
analytics = db_manager.get_user_analytics(user_id, time_range="30d")
```

Single Database Operations:
```bash
# PostgreSQL schema migration
moai db:migrate --database postgresql --migration-file schema_v2.sql

# MongoDB aggregation pipeline
moai db:aggregate --collection users --pipeline analytics_pipeline.json

# Redis cache warming
moai db:cache:warm --pattern "user:*" --ttl 3600
```

### Core Components

1. PostgreSQL (`modules/postgresql.md`)
- Advanced schema design and constraints
- Complex query optimization and indexing
- Window functions and CTEs
- Partitioning and materialized views
- Connection pooling and performance tuning

2. MongoDB (`modules/mongodb.md`)
- Document modeling and schema design
- Aggregation pipelines for analytics
- Indexing strategies and performance
- Sharding and scaling patterns
- Data consistency and validation

3. Redis (`modules/redis.md`)
- Multi-layer caching strategies
- Real-time analytics and counting
- Distributed locking and coordination
- Pub/sub messaging and streams
- Advanced data structures (HyperLogLog, Geo)

---

## Advanced Patterns (10+ minutes)

### Multi-Database Architecture

Polyglot Persistence Pattern:
```python
class DataRouter:
 def __init__(self):
 self.postgresql = PostgreSQLConnection()
 self.mongodb = MongoDBConnection()
 self.redis = RedisConnection()

 def get_user_profile(self, user_id):
 # Get structured user data from PostgreSQL
 user = self.postgresql.get_user(user_id)

 # Get flexible profile data from MongoDB
 profile = self.mongodb.get_user_profile(user_id)

 # Get real-time status from Redis
 status = self.redis.get_user_status(user_id)

 return self.merge_user_data(user, profile, status)

 def update_user_data(self, user_id, data):
 # Route different data types to appropriate databases
 if 'structured_data' in data:
 self.postgresql.update_user(user_id, data['structured_data'])

 if 'profile_data' in data:
 self.mongodb.update_user_profile(user_id, data['profile_data'])

 if 'real_time_data' in data:
 self.redis.set_user_status(user_id, data['real_time_data'])

 # Invalidate cache across databases
 self.invalidate_user_cache(user_id)
```

Data Synchronization:
```python
class DataSyncManager:
 def sync_user_data(self, user_id):
 # Sync from PostgreSQL to MongoDB for search
 pg_user = self.postgresql.get_user(user_id)
 search_document = self.create_search_document(pg_user)
 self.mongodb.upsert_user_search(user_id, search_document)

 # Update cache in Redis
 cache_data = self.create_cache_document(pg_user)
 self.redis.set_user_cache(user_id, cache_data, ttl=3600)
```

### Performance Optimization

Query Performance Analysis:
```python
# PostgreSQL query optimization
def analyze_query_performance(query):
 explain_result = postgresql.execute(f"EXPLAIN (ANALYZE, BUFFERS) {query}")
 return QueryAnalyzer(explain_result).get_optimization_suggestions()

# MongoDB aggregation optimization
def optimize_aggregation_pipeline(pipeline):
 optimizer = AggregationOptimizer()
 return optimizer.optimize_pipeline(pipeline)

# Redis performance monitoring
def monitor_redis_performance():
 metrics = redis.info()
 return PerformanceAnalyzer(metrics).get_recommendations()
```

Scaling Strategies:
```python
# Read replicas for PostgreSQL
read_replicas = postgresql.setup_read_replicas([
 "postgresql://replica1...",
 "postgresql://replica2..."
])

# Sharding for MongoDB
mongodb.setup_sharding(
 shard_key="user_id",
 num_shards=4
)

# Redis clustering
redis.setup_cluster([
 "redis://node1:7000",
 "redis://node2:7000",
 "redis://node3:7000"
])
```

---

## Works Well With

Complementary Skills:
- `moai-domain-backend` - API integration and business logic
- `moai-foundation-core` - Database migration and schema management
- `moai-workflow-project` - Database project setup and configuration
- `moai-platform-supabase` - Supabase database integration patterns
- `moai-platform-neon` - Neon database integration patterns
- `moai-platform-firestore` - Firestore database integration patterns

Technology Integration:
- ORMs and ODMs (SQLAlchemy, Mongoose, TypeORM)
- Connection pooling (PgBouncer, connection pools)
- Migration tools (Alembic, Flyway)
- Monitoring (pg_stat_statements, MongoDB Atlas)
- Cache invalidation and synchronization

---

## Usage Examples

### Database Operations
```python
# PostgreSQL advanced queries
users = postgresql.query(
 "SELECT * FROM users WHERE created_at > %s ORDER BY activity_score DESC LIMIT 100",
 [datetime.now() - timedelta(days=30)]
)

# MongoDB analytics
analytics = mongodb.aggregate('events', [
 {"$match": {"timestamp": {"$gte": start_date}}},
 {"$group": {"_id": "$type", "count": {"$sum": 1}}},
 {"$sort": {"count": -1}}
])

# Redis caching operations
async def get_user_data(user_id):
 cache_key = f"user:{user_id}"
 data = await redis.get(cache_key)

 if not data:
 data = fetch_from_database(user_id)
 await redis.setex(cache_key, 3600, json.dumps(data))

 return json.loads(data)
```

### Multi-Database Transactions
```python
async def create_user_with_profile(user_data, profile_data):
 try:
 # Start transaction across databases
 async with transaction_manager():
 # Create user in PostgreSQL
 user_id = await postgresql.insert_user(user_data)

 # Create profile in MongoDB
 await mongodb.insert_user_profile(user_id, profile_data)

 # Set initial cache in Redis
 await redis.set_user_cache(user_id, {
 "id": user_id,
 "status": "active",
 "created_at": datetime.now().isoformat()
 })

 return user_id

 except Exception as e:
 # Automatic rollback across databases
 logger.error(f"User creation failed: {e}")
 raise
```

---

## Technology Stack

Relational Database:
- PostgreSQL 14+ (primary)
- MySQL 8.0+ (alternative)
- Connection pooling (PgBouncer, SQLAlchemy)

NoSQL Database:
- MongoDB 6.0+ (primary)
- Document modeling and validation
- Aggregation framework
- Sharding and replication

In-Memory Database:
- Redis 7.0+ (primary)
- Redis Stack for advanced features
- Clustering and high availability
- Advanced data structures

Supporting Tools:
- Migration tools (Alembic, Flyway)
- Monitoring (Prometheus, Grafana)
- ORMs/ODMs (SQLAlchemy, Mongoose)
- Connection management

Performance Features:
- Query optimization and analysis
- Index management and strategies
- Caching layers and invalidation
- Load balancing and failover

---

*For detailed implementation patterns and database-specific optimizations, see the `modules/` directory.*
