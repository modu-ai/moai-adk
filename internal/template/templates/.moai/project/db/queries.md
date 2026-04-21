# Query Patterns

_TBD — Document recurring query patterns, aggregations, and report queries as they are developed.
These serve as a team reference and input for query optimization._

---

## Common Queries

Frequently used queries for core application flows.

### _TBD_ Query Name

```sql
-- Purpose: _TBD_
-- Parameters: _TBD_
-- Returns: _TBD_

SELECT *
FROM _TBD_
WHERE _TBD_
LIMIT 20;
```

<!--
Example:
### Paginated user posts

-- Purpose: Fetch a page of posts for a given user, ordered by recency
-- Parameters: $1 = user_id (uuid), $2 = limit (int), $3 = offset (int)
-- Returns: posts with author info

SELECT
  p.id,
  p.title,
  p.status,
  p.created_at,
  u.name AS author_name
FROM posts p
JOIN users u ON u.id = p.user_id
WHERE p.user_id = $1
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;
-->

---

## Aggregations

Summary queries used for analytics and business logic.

### _TBD_ Aggregation Name

```sql
-- Purpose: _TBD_
-- Frequency: _TBD_ (e.g., on-demand, scheduled nightly)
-- Performance note: _TBD_

SELECT
  _TBD_
FROM _TBD_
GROUP BY _TBD_;
```

<!--
Example:
### Posts per user (last 30 days)

-- Purpose: Count how many posts each user published in the last 30 days
-- Frequency: On-demand (profile page)
-- Performance note: Index on (user_id, created_at) required

SELECT
  u.id,
  u.name,
  COUNT(p.id) AS post_count
FROM users u
LEFT JOIN posts p ON p.user_id = u.id
  AND p.created_at >= NOW() - INTERVAL '30 days'
GROUP BY u.id, u.name
ORDER BY post_count DESC;
-->

---

## Reports

Complex queries that power dashboards and exports.

### _TBD_ Report Name

```sql
-- Purpose: _TBD_
-- Used by: _TBD_ (e.g., admin dashboard, scheduled export)
-- Estimated runtime: _TBD_

SELECT
  _TBD_
FROM _TBD_;
```

<!--
Example:
### Monthly active users (MAU) by cohort

-- Purpose: Track MAU broken down by signup cohort month
-- Used by: Growth dashboard
-- Estimated runtime: < 2s with materialized view

SELECT
  DATE_TRUNC('month', u.created_at) AS cohort_month,
  DATE_TRUNC('month', a.activity_at) AS active_month,
  COUNT(DISTINCT u.id) AS mau
FROM users u
JOIN user_activity a ON a.user_id = u.id
GROUP BY cohort_month, active_month
ORDER BY cohort_month, active_month;
-->
