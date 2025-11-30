# PostgreSQL Advanced Patterns

## Overview
Comprehensive PostgreSQL patterns covering advanced schema design, query optimization, indexing strategies, partitioning, and performance tuning for modern applications.

## Quick Implementation

### Advanced Schema Design

```sql
-- User table with proper constraints and indexes
CREATE TABLE users (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 email VARCHAR(255) UNIQUE NOT NULL,
 username VARCHAR(50) UNIQUE,
 password_hash VARCHAR(255) NOT NULL,
 email_verified BOOLEAN DEFAULT FALSE,
 created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
 updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
 last_login_at TIMESTAMP WITH TIME ZONE
);

-- Optimized indexes
CREATE INDEX idx_users_email ON users (email) WHERE email_verified = TRUE;
CREATE INDEX idx_users_created_at ON users (created_at DESC);
CREATE INDEX idx_users_username_trgm ON users USING gin (username gin_trgm_ops);

-- User profiles with JSONB
CREATE TABLE user_profiles (
 user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
 display_name VARCHAR(100),
 bio TEXT,
 avatar_url VARCHAR(500),
 preferences JSONB DEFAULT '{}',
 metadata JSONB DEFAULT '{}',
 created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
 updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- GIN index for JSONB queries
CREATE INDEX idx_user_profiles_preferences ON user_profiles USING gin (preferences);
```

### Advanced Query Patterns

```sql
-- Efficient pagination with cursor-based approach
WITH ranked_users AS (
 SELECT
 u.id,
 u.username,
 u.email,
 up.display_name,
 ROW_NUMBER() OVER (ORDER BY u.created_at DESC) as rn
 FROM users u
 LEFT JOIN user_profiles up ON u.id = up.user_id
 WHERE u.created_at < :cursor OR :cursor IS NULL
)
SELECT * FROM ranked_users
WHERE rn > :offset
ORDER BY created_at DESC
LIMIT :limit;

-- Full-text search with trigrams
SELECT
 u.id,
 u.username,
 up.display_name,
 similarity(u.username, :query) as username_sim,
 ts_rank_cd(search_vector, plainto_tsquery('english', :query)) as search_rank
FROM users u
LEFT JOIN user_profiles up ON u.id = up.user_id,
 to_tsvector('english', u.username || ' ' || COALESCE(up.display_name, '')) search_vector
WHERE u.username % :query
 OR plainto_tsquery('english', :query) @@ search_vector
ORDER BY username_sim DESC, search_rank DESC
LIMIT 20;

-- Window functions for analytics
SELECT
 DATE_TRUNC('month', created_at) as month,
 COUNT(*) as new_users,
 SUM(COUNT(*)) OVER (ORDER BY DATE_TRUNC('month', created_at)) as cumulative_users,
 COUNT(*) - LAG(COUNT(*)) OVER (ORDER BY DATE_TRUNC('month', created_at)) as month_over_month_change
FROM users
WHERE created_at >= NOW() - INTERVAL '1 year'
GROUP BY DATE_TRUNC('month', created_at)
ORDER BY month;
```

### Performance Optimization

```sql
-- Partitioning for time-series data
CREATE TABLE events (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id UUID NOT NULL REFERENCES users(id),
 event_type VARCHAR(50) NOT NULL,
 data JSONB,
 created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- Monthly partitions
CREATE TABLE events_2024_01 PARTITION OF events
 FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE events_2024_02 PARTITION OF events
 FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Materialized views for analytics
CREATE MATERIALIZED VIEW user_stats AS
SELECT
 u.id,
 u.email,
 u.created_at,
 COUNT(DISTINCT e.id) as event_count,
 MAX(e.created_at) as last_event_at,
 COUNT(DISTINCT CASE WHEN e.event_type = 'login' THEN e.id END) as login_count
FROM users u
LEFT JOIN events e ON u.id = e.user_id
GROUP BY u.id, u.email, u.created_at;

-- Refresh materialized view
CREATE OR REPLACE FUNCTION refresh_user_stats()
RETURNS void AS $$
BEGIN
 REFRESH MATERIALIZED VIEW CONCURRENTLY user_stats;
END;
$$ LANGUAGE plpgsql;
```

## Key Features

### 1. Advanced Indexing
- Partial indexes for filtered data
- GIN indexes for full-text search and JSONB
- Expression indexes for computed values
- Composite indexes for multi-column queries

### 2. Query Optimization
- Cursor-based pagination
- Window functions for analytics
- CTE (Common Table Expressions)
- Materialized views for performance

### 3. Data Modeling
- JSONB for flexible schemas
- Array types for multi-valued data
- Range types for temporal data
- Custom types for domain-specific data

### 4. Performance Tuning
- Table partitioning
- Connection pooling
- Query execution plans
- Vacuum and analyze strategies

## Best Practices
- Use appropriate data types and constraints
- Implement proper indexing strategies
- Monitor query performance
- Use transactions for data consistency
- Implement backup and recovery procedures

## Integration Points
- ORMs (SQLAlchemy, Django ORM)
- Connection pools (PgBouncer, connection poolers)
- Monitoring tools (pg_stat_statements)
- Migration tools (Alembic, Flyway)
