# Neon Platform Reference (2025)

## Official Documentation
- [Neon Platform](https://neon.tech/docs)
- [PostgreSQL 16 Features](https://www.postgresql.org/docs/16/)
- [Neon API Reference](https://neon.tech/docs/reference/api)
- [Neon CLI](https://neon.tech/docs/reference/cli)

## Latest Features (2025)

### Autoscaling 2.0
- Dynamic compute adjustment based on load
- Sub-second scaling response time
- Predictive scaling with ML models
- Cost optimization with automatic scale-to-zero

### Enhanced Branching
- Instant branch creation (< 1 second)
- Copy-on-write for efficient storage
- Branch-per-PR workflows
- Automated branch lifecycle management

### PgBouncer Improvements
- Connection pooling optimization
- Transaction-level pooling support
- Session-level pooling improvements
- Automatic failover handling

### Performance Monitoring
- Real-time query performance analytics
- Slow query detection and alerts
- Connection pool metrics
- Storage usage optimization

## Context7 Integration

Access latest Neon documentation:
```python
docs = await context7.get_library_docs(
    context7_library_id="/neon/docs",
    topic="branching autoscaling performance 2025"
)
```

## Connection Examples

### Node.js with Neon
```javascript
const { Pool } = require('pg');

const pool = new Pool({
  connectionString: process.env.NEON_DATABASE_URL,
  ssl: { rejectUnauthorized: false },
  max: 20,
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 2000,
});
```

### Python with Neon
```python
import psycopg2
from psycopg2.pool import ThreadedConnectionPool

pool = ThreadedConnectionPool(
    minconn=5,
    maxconn=20,
    dsn=os.environ['NEON_DATABASE_URL']
)
```

## Best Practices

1. **Use connection pooling** - PgBouncer or application-level pooling
2. **Implement branch-per-PR** - Isolated testing environments
3. **Monitor query performance** - Identify and optimize slow queries
4. **Leverage autoscaling** - Let Neon handle compute resources
5. **Regular backups** - Use point-in-time recovery features

---

**Last Updated**: 2025-11-22
