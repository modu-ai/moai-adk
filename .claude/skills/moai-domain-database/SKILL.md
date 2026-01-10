---
name: "moai-domain-database"
description: "Database specialist covering PostgreSQL, MongoDB, Redis, and advanced data patterns for modern applications"
version: 1.0.0
category: "domain"
modularized: true
user-invocable: false
tags: ['database', 'postgresql', 'mongodb', 'redis', 'data-patterns', 'performance']
updated: 2026-01-11
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
status: "active"
author: "MoAI-ADK Team"
---

# Database Domain Specialist

## Quick Reference

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

---

## Advanced Patterns

### Multi-Database Architecture

Polyglot Persistence Pattern:

Create a DataRouter class that initializes connections to PostgreSQL, MongoDB, and Redis. Implement get_user_profile method that retrieves structured user data from PostgreSQL, flexible profile data from MongoDB, and real-time status from Redis, then merges all data sources. Implement update_user_data method that routes structured data updates to PostgreSQL, profile data updates to MongoDB, and real-time data updates to Redis, followed by cache invalidation.

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
- Migration tools including Alembic and Flyway
- Monitoring with pg_stat_statements and MongoDB Atlas
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
