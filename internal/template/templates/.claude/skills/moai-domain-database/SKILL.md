---
name: moai-domain-database
description: >
  Database specialist covering PostgreSQL, MongoDB, Redis, Oracle, and advanced data
  patterns for modern applications. Use for database schema design, query optimization,
  indexing strategies, or data modeling.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Bash(psql:*), Bash(mysql:*), Bash(sqlite3:*), Bash(mongosh:*), Bash(redis-cli:*), Bash(npm:*), Bash(npx:*), Bash(prisma:*), Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-01-11"
  modularized: "true"
  tags: "database, postgresql, mongodb, redis, oracle, data-patterns, performance"
  author: "MoAI-ADK Team"

# MoAI Extension: Triggers
triggers:
  keywords: ["database", "PostgreSQL", "MongoDB", "Redis", "Oracle", "SQL", "NoSQL", "PL/SQL", "query", "schema", "migration", "indexing", "ORM", "ODM", "SQLAlchemy", "Mongoose", "Prisma", "Drizzle", "python-oracledb", "cx_Oracle", "connection pool", "transaction", "data modeling", "aggregation", "partitioning", "hierarchical query"]
---

# Database Domain Specialist

## Quick Reference

Enterprise Database Expertise - Comprehensive database patterns and implementations covering PostgreSQL, MongoDB, Redis, Oracle, and advanced data management for scalable modern applications.

Core Capabilities:

- PostgreSQL: Advanced relational patterns, optimization, and scaling
- MongoDB: Document modeling, aggregation, and NoSQL performance tuning
- Redis: In-memory caching, real-time analytics, and distributed systems
- Oracle: Enterprise patterns, PL/SQL, partitioning, and hierarchical queries
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

## Implementation Guide

### Quick Start Workflow

Database Stack Initialization:

Create a DatabaseManager instance and configure multiple database connections. Set up PostgreSQL with connection string, pool size of 20, and query logging enabled. Configure MongoDB with connection string, database name, and sharding enabled. Configure Redis with connection string, max connections of 50, and clustering enabled. Use the unified interface to query user data with profile and analytics across all database types.

Single Database Operations:

Run PostgreSQL schema migrations using the migration command with the database type and migration file path. Execute MongoDB aggregation pipelines by specifying the collection name and pipeline JSON file. Warm Redis cache by specifying key patterns and TTL values.

### Core Components

PostgreSQL Module:

- Advanced schema design and constraints
- Complex query optimization and indexing
- Window functions and CTEs
- Partitioning and materialized views
- Connection pooling and performance tuning

MongoDB Module:

- Document modeling and schema design
- Aggregation pipelines for analytics
- Indexing strategies and performance
- Sharding and scaling patterns
- Data consistency and validation

Redis Module:

- Multi-layer caching strategies
- Real-time analytics and counting
- Distributed locking and coordination
- Pub/sub messaging and streams
- Advanced data structures including HyperLogLog and Geo

Oracle Module:

- Hierarchical and recursive query patterns (CONNECT BY)
- PL/SQL procedures, packages, and batch operations
- Partitioning strategies (range, list, hash, composite)
- Enterprise features and statement caching
- LOB handling and large data processing

---

## Advanced Patterns

### Multi-Database Architecture

Polyglot Persistence Pattern:

Create a DataRouter class that initializes connections to PostgreSQL, MongoDB, Redis, and Oracle. Implement get_user_profile method that retrieves structured user data from PostgreSQL or Oracle, flexible profile data from MongoDB, and real-time status from Redis, then merges all data sources. Implement update_user_data method that routes structured data updates to PostgreSQL/Oracle, profile data updates to MongoDB, and real-time data updates to Redis, followed by cache invalidation.

Data Synchronization:

Create a DataSyncManager class that synchronizes user data across databases. Implement sync_user_data method that retrieves user from PostgreSQL, creates a search document for MongoDB, upserts to the MongoDB search collection, creates cache data, and updates Redis cache with TTL.

### Performance Optimization

Query Performance Analysis:

For PostgreSQL, execute EXPLAIN ANALYZE BUFFERS on queries and use a QueryAnalyzer to generate optimization suggestions. For MongoDB, create an AggregationOptimizer to analyze and optimize aggregation pipelines. For Redis, retrieve info metrics and use a PerformanceAnalyzer to generate recommendations.

Scaling Strategies:

Configure PostgreSQL read replicas by providing replica connection URLs. Set up MongoDB sharding with shard key and number of shards. Configure Redis clustering by providing node URLs for the cluster.

---

## Works Well With

Complementary Skills:

- moai-domain-backend - API integration and business logic
- moai-foundation-core - Database migration and schema management
- moai-workflow-project - Database project setup and configuration
- moai-platform-supabase - Supabase database integration patterns
- moai-platform-neon - Neon database integration patterns
- moai-platform-firestore - Firestore database integration patterns

Technology Integration:

- ORMs and ODMs including SQLAlchemy, Mongoose, and TypeORM
- Connection pooling with PgBouncer and connection pools
- Migration tools including Alembic, Flyway, and Data Pump
- Monitoring with pg_stat_statements, MongoDB Atlas, and Oracle AWR
- python-oracledb for Oracle connectivity and PL/SQL execution
- Cache invalidation and synchronization

---

## Technology Stack

Relational Database:

- PostgreSQL 14+ as primary database
- MySQL 8.0+ as alternative
- Connection pooling with PgBouncer and SQLAlchemy

NoSQL Database:

- MongoDB 6.0+ as primary document store
- Document modeling and validation
- Aggregation framework
- Sharding and replication

In-Memory Database:

- Redis 7.0+ as primary cache
- Redis Stack for advanced features
- Clustering and high availability
- Advanced data structures

Enterprise Database:

- Oracle 19c+ / 21c+ for enterprise workloads
- python-oracledb (successor to cx_Oracle)
- PL/SQL procedures and packages
- Partitioning and advanced analytics

Supporting Tools:

- Migration tools including Alembic and Flyway
- Monitoring with Prometheus and Grafana
- ORMs and ODMs including SQLAlchemy and Mongoose
- Connection management utilities

Performance Features:

- Query optimization and analysis
- Index management and strategies
- Caching layers and invalidation
- Load balancing and failover

---

## Resources

For working code examples, see [examples.md](examples.md).

For detailed implementation patterns and database-specific optimizations, see the modules directory.

Status: Production Ready
Last Updated: 2026-01-11
Maintained by: MoAI-ADK Database Team

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "I do not need an index, the table is small" | Tables grow. The missing index that is invisible at 1K rows becomes a production incident at 1M rows. |
| "I will add the migration later" | Schema changes without migrations are unreproducible. Every change must have a reversible migration script. |
| "This query works fine in development" | Development databases have tiny datasets. Production query plans differ dramatically at scale. Explain analyze first. |
| "NoSQL does not need schema design" | Schemaless does not mean designless. Document structure decisions affect every query and index. |
| "I will just add a column, it is non-breaking" | Adding a NOT NULL column without a default breaks existing inserts. Column additions need default values or migration backfills. |
| "Connection pooling is handled by the framework" | Framework defaults are generic. Pool size, timeout, and idle limits must be tuned to the workload. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Schema change committed without a corresponding migration file
- Query uses SELECT * in production code instead of explicit column list
- No index exists for columns used in WHERE, JOIN, or ORDER BY clauses
- Connection string hardcoded in source instead of environment variable
- Transaction scope spans user-facing HTTP request duration (long-held locks)
- No EXPLAIN ANALYZE output for new queries touching large tables

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Migration file exists for every schema change (show migration file list)
- [ ] Indexes exist for frequently queried columns (show index definitions)
- [ ] EXPLAIN ANALYZE run for new queries on representative data (show output)
- [ ] Connection credentials sourced from environment variables
- [ ] Transaction scopes are minimal and do not span I/O waits
- [ ] Backup and restore procedure documented and tested
- [ ] Connection pool settings configured with explicit size and timeout

<!-- moai:evolvable-end -->
