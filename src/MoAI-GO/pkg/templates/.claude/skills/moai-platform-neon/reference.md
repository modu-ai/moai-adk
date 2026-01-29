# Neon Platform Reference Guide

## API Reference

### Neon Management API

Base URL: https://console.neon.tech/api/v2

Authentication: Bearer token via NEON_API_KEY header

### Common Endpoints

Project Endpoints:
- GET /projects - List all projects
- GET /projects/{project_id} - Get project details
- PATCH /projects/{project_id} - Update project settings

Branch Endpoints:
- GET /projects/{project_id}/branches - List branches
- POST /projects/{project_id}/branches - Create branch
- DELETE /projects/{project_id}/branches/{branch_id} - Delete branch
- GET /projects/{project_id}/branches/{branch_id}/endpoints - Get branch endpoints

Endpoint Endpoints:
- GET /projects/{project_id}/endpoints - List compute endpoints
- PATCH /projects/{project_id}/endpoints/{endpoint_id} - Update endpoint settings

### API Response Formats

Branch Creation Response:
```json
{
  "branch": {
    "id": "br-xxx",
    "name": "feature-branch",
    "project_id": "project-xxx",
    "parent_id": "br-main",
    "created_at": "2024-01-01T00:00:00Z",
    "current_state": "ready"
  },
  "endpoints": [
    {
      "id": "ep-xxx",
      "branch_id": "br-xxx",
      "host": "ep-xxx.region.neon.tech",
      "connection_uri": "postgresql://..."
    }
  ]
}
```

Endpoint Configuration Response:
```json
{
  "endpoint": {
    "id": "ep-xxx",
    "autoscaling_limit_min_cu": 0.25,
    "autoscaling_limit_max_cu": 4,
    "suspend_timeout_seconds": 300,
    "current_state": "active"
  }
}
```

---

## Environment Configuration

### Connection String Formats

Direct Connection (for migrations and admin tasks):
```
postgresql://user:password@ep-xxx.region.neon.tech/dbname?sslmode=require
```

Pooled Connection (for serverless and edge):
```
postgresql://user:password@ep-xxx-pooler.region.neon.tech/dbname?sslmode=require
```

### Environment Variables

Required Variables:
- DATABASE_URL: Primary connection string for migrations
- DATABASE_URL_POOLED: Pooled connection for application use
- NEON_API_KEY: API key for branch management
- NEON_PROJECT_ID: Project identifier for API calls

Optional Variables:
- NEON_BRANCH_ID: Specific branch identifier
- NEON_ENDPOINT_ID: Specific endpoint identifier

### Connection String Parameters

SSL Mode Options:
- sslmode=require: Require SSL (recommended)
- sslmode=verify-full: Verify SSL certificate

Connection Pool Options:
- connection_limit: Maximum connections (pooled only)
- pool_timeout: Connection timeout in seconds

---

## Compute Unit Reference

### Compute Unit (CU) Specifications

0.25 CU: 0.25 vCPU, 1 GB RAM - Development and testing
0.5 CU: 0.5 vCPU, 2 GB RAM - Light production workloads
1 CU: 1 vCPU, 4 GB RAM - Standard production
2 CU: 2 vCPU, 8 GB RAM - Medium workloads
4 CU: 4 vCPU, 16 GB RAM - Heavy workloads
8 CU: 8 vCPU, 32 GB RAM - High-performance requirements

### Auto-Scaling Configuration

Minimum CU: Lowest compute level (0.25 for scale-to-zero)
Maximum CU: Highest compute level for peak load
Suspend Timeout: Seconds of inactivity before scaling to zero

Recommended Settings by Use Case:

Development: min 0.25, max 0.5, timeout 300
Staging: min 0.25, max 1, timeout 600
Production: min 0.5, max 4, timeout 3600
High-Traffic: min 1, max 8, timeout 0 (never suspend)

---

## PostgreSQL Extensions

### Supported Extensions

Core Extensions:
- pg_stat_statements: Query performance statistics
- pg_trgm: Trigram text similarity
- uuid-ossp: UUID generation
- hstore: Key-value storage
- pgcrypto: Cryptographic functions

Spatial Extensions:
- postgis: Spatial and geographic objects
- postgis_topology: Topology support

Full-Text Search:
- pg_search: Full-text search improvements
- unaccent: Remove accents from text

JSON Processing:
- jsonb_plperl: Perl JSON functions
- jsonb_plpython3u: Python JSON functions

### Extension Installation

Enable Extension:
```sql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
CREATE EXTENSION IF NOT EXISTS uuid-ossp;
CREATE EXTENSION IF NOT EXISTS postgis;
```

---

## Provider Comparison Matrix

### Feature Comparison

Serverless Compute:
- Neon: Full auto-scaling with scale-to-zero
- Supabase: Fixed compute tiers
- PlanetScale: MySQL-based serverless

Database Branching:
- Neon: Instant copy-on-write branches
- Supabase: Manual database cloning
- PlanetScale: Schema-only branching

Point-in-Time Recovery:
- Neon: 30-day PITR with instant restore
- Supabase: 7-day backups (Pro tier)
- PlanetScale: Continuous backups

Connection Pooling:
- Neon: Built-in HTTP and WebSocket pooler
- Supabase: PgBouncer integration
- PlanetScale: Native connection handling

Edge Compatibility:
- Neon: Full edge runtime support
- Supabase: Edge functions with pooler
- PlanetScale: Edge-compatible driver

### Pricing Comparison

Free Tier Storage:
- Neon: 3 GB
- Supabase: 500 MB
- PlanetScale: 5 GB

Free Tier Compute:
- Neon: 100 compute hours
- Supabase: Fixed 500 MB RAM
- PlanetScale: 1 billion row reads

---

## Error Codes and Troubleshooting

### Common Error Codes

Connection Errors:
- ECONNREFUSED: Endpoint suspended, will wake on next request
- SSL_CERTIFICATE_REQUIRED: Missing sslmode=require parameter
- TOO_MANY_CONNECTIONS: Pool exhausted, use pooled connection

Branch Errors:
- BRANCH_NOT_FOUND: Invalid branch ID
- BRANCH_LIMIT_EXCEEDED: Project branch limit reached
- PARENT_BRANCH_NOT_FOUND: Invalid parent for branching

API Errors:
- 401 Unauthorized: Invalid or expired API key
- 403 Forbidden: Insufficient permissions
- 429 Too Many Requests: Rate limit exceeded

### Troubleshooting Guide

Slow Cold Starts:
- Increase minimum compute units
- Reduce suspend timeout
- Use connection pooling

Connection Timeouts:
- Switch to pooled connection string
- Verify network connectivity
- Check endpoint status via API

Branch Creation Failures:
- Verify parent branch exists
- Check project branch limits
- Ensure API key has write permissions

---

## Security Best Practices

### Connection Security

Always Use SSL: Set sslmode=require in connection strings
Rotate Credentials: Regularly rotate database passwords and API keys
Use Environment Variables: Never hardcode credentials in source code
Restrict IP Access: Configure IP allow lists for production

### API Key Management

Scoped Keys: Create keys with minimal required permissions
Key Rotation: Rotate API keys on regular schedule
Audit Access: Monitor API key usage via Neon console
Revoke Unused: Remove keys no longer in use

### Database Security

Row-Level Security: Implement RLS for multi-tenant applications
Prepared Statements: Always use parameterized queries
Audit Logging: Enable pg_stat_statements for query monitoring
Backup Verification: Regularly test PITR restoration

---

Version: 2.0.0
Last Updated: 2026-01-06
