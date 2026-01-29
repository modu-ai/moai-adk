# Database Domain Modules

This directory contains specialized modules for advanced database patterns and implementations across multiple database systems.

## Module Structure

### Core Database Systems

- postgresql.md - Advanced PostgreSQL patterns, optimization, and scaling
- mongodb.md - NoSQL document modeling, aggregation, and performance tuning
- redis.md - In-memory caching, real-time data structures, and distributed systems
- oracle.md - Enterprise Oracle patterns, PL/SQL, partitioning, and hierarchical queries

### Usage Patterns

1. Relational Database: Use `postgresql.md` for structured data and complex queries
2. Document Database: Use `mongodb.md` for flexible schemas and rapid development
3. In-Memory Store: Use `redis.md` for caching, sessions, and real-time features
4. Enterprise Database: Use `oracle.md` for Oracle-specific patterns and PL/SQL
5. Multi-Database: Combine modules for hybrid data architectures

### Integration Guidelines

Each module provides comprehensive patterns for:

```python
# PostgreSQL integration
from moai_domain_database.modules.postgresql import AdvancedPostgreSQL

# MongoDB integration
from moai_domain_database.modules.mongodb import MongoAggregation

# Redis integration
from moai_domain_database.modules.redis import CacheManager

# Oracle integration
from moai_domain_database.modules.oracle import OracleEnterprise

# Hybrid database architecture
def setup_database_stack():
    postgresql = AdvancedPostgreSQL()
    mongodb = MongoAggregation()
    redis = CacheManager()
    oracle = OracleEnterprise()
    return DatabaseStack(postgresql, mongodb, redis, oracle)
```

### Progressive Disclosure

- Quick Start: Use individual database modules for specific needs
- Implementation: Combine modules for comprehensive data solutions
- Advanced: Custom implementations based on module patterns

### Dependencies

- PostgreSQL: asyncpg, SQLAlchemy, Psycopg2
- MongoDB: pymongo, motor, mongoose
- Redis: redis-py, ioredis, node-redis
- Oracle: python-oracledb, SQLAlchemy Oracle dialect
- General: Connection pooling, ORM frameworks, migration tools
