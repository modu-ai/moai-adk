# Neon Advanced Architecture Patterns

**Enterprise-grade patterns for Neon serverless PostgreSQL**

---

## Pattern 1: Multi-Tenant Architecture with RLS

**Use Case**: SaaS applications with tenant isolation

### Row-Level Security (RLS) Implementation

```sql
-- Enable RLS on tables
CREATE TABLE organizations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  email VARCHAR(255) UNIQUE NOT NULL,
  role VARCHAR(50) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organization_id UUID NOT NULL REFERENCES organizations(id),
  title VARCHAR(255) NOT NULL,
  content TEXT,
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMP DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Create policies
CREATE POLICY tenant_isolation ON documents
  USING (organization_id = current_setting('app.current_org_id')::UUID);

CREATE POLICY user_org_access ON users
  USING (organization_id = current_setting('app.current_org_id')::UUID);
```

### Application-Level Tenant Switching

```typescript
import { Pool } from '@neondatabase/serverless';

const pool = new Pool({ connectionString: process.env.DATABASE_URL! });

export async function queryWithTenant<T>(
  organizationId: string,
  queryFn: (client: any) => Promise<T>
): Promise<T> {
  const client = await pool.connect();

  try {
    // Set tenant context
    await client.query('BEGIN');
    await client.query(
      `SET LOCAL app.current_org_id = $1`,
      [organizationId]
    );

    // Execute query (RLS automatically enforced)
    const result = await queryFn(client);

    await client.query('COMMIT');
    return result;
  } catch (error) {
    await client.query('ROLLBACK');
    throw error;
  } finally {
    client.release();
  }
}

// Usage
export async function getDocuments(organizationId: string) {
  return await queryWithTenant(organizationId, async (client) => {
    const { rows } = await client.query(`
      SELECT id, title, content, created_at
      FROM documents
      ORDER BY created_at DESC
    `);
    return rows;
  });
}
```

---

## Pattern 2: Event Sourcing with PostgreSQL

**Use Case**: Audit trails, time-travel queries, CQRS

### Event Store Schema

```sql
CREATE TABLE events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  aggregate_id UUID NOT NULL,
  aggregate_type VARCHAR(100) NOT NULL,
  event_type VARCHAR(100) NOT NULL,
  event_data JSONB NOT NULL,
  metadata JSONB DEFAULT '{}',
  version INTEGER NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(aggregate_id, version)
);

CREATE INDEX idx_events_aggregate ON events(aggregate_id, version);
CREATE INDEX idx_events_type ON events(event_type);
CREATE INDEX idx_events_created ON events(created_at DESC);

-- Materialized view for current state
CREATE MATERIALIZED VIEW user_current_state AS
SELECT DISTINCT ON (aggregate_id)
  aggregate_id as user_id,
  event_data->>'name' as name,
  event_data->>'email' as email,
  event_data->>'status' as status,
  created_at as last_updated
FROM events
WHERE aggregate_type = 'User'
ORDER BY aggregate_id, version DESC;

CREATE UNIQUE INDEX ON user_current_state(user_id);
```

### Event Sourcing Implementation

```typescript
interface Event {
  aggregateId: string;
  aggregateType: string;
  eventType: string;
  eventData: Record<string, any>;
  metadata?: Record<string, any>;
  version: number;
}

export class EventStore {
  constructor(private pool: Pool) {}

  async appendEvent(event: Event): Promise<void> {
    await this.pool.query(
      `INSERT INTO events (aggregate_id, aggregate_type, event_type, event_data, metadata, version)
       VALUES ($1, $2, $3, $4, $5, $6)`,
      [
        event.aggregateId,
        event.aggregateType,
        event.eventType,
        JSON.stringify(event.eventData),
        JSON.stringify(event.metadata || {}),
        event.version,
      ]
    );
  }

  async getEvents(aggregateId: string): Promise<Event[]> {
    const { rows } = await this.pool.query(
      `SELECT * FROM events
       WHERE aggregate_id = $1
       ORDER BY version ASC`,
      [aggregateId]
    );

    return rows.map((row) => ({
      aggregateId: row.aggregate_id,
      aggregateType: row.aggregate_type,
      eventType: row.event_type,
      eventData: row.event_data,
      metadata: row.metadata,
      version: row.version,
    }));
  }

  async getCurrentState(aggregateId: string): Promise<any> {
    const { rows } = await this.pool.query(
      `SELECT * FROM user_current_state WHERE user_id = $1`,
      [aggregateId]
    );
    return rows[0];
  }

  async refreshMaterializedView(): Promise<void> {
    await this.pool.query('REFRESH MATERIALIZED VIEW CONCURRENTLY user_current_state');
  }
}
```

---

## Pattern 3: Read Replicas for Query Separation

**Use Case**: Read-heavy applications, analytics workloads

### Read/Write Connection Split

```typescript
import { Pool } from '@neondatabase/serverless';

// Primary (write)
const primaryPool = new Pool({
  connectionString: process.env.PRIMARY_DATABASE_URL!,
  max: 10,
});

// Read replica
const replicaPool = new Pool({
  connectionString: process.env.REPLICA_DATABASE_URL!,
  max: 20, // More connections for reads
});

export async function write<T>(queryFn: (pool: Pool) => Promise<T>): Promise<T> {
  return await queryFn(primaryPool);
}

export async function read<T>(queryFn: (pool: Pool) => Promise<T>): Promise<T> {
  return await queryFn(replicaPool);
}

// Usage examples
export async function createUser(name: string, email: string) {
  return await write(async (pool) => {
    const { rows } = await pool.query(
      'INSERT INTO users (name, email) VALUES ($1, $2) RETURNING *',
      [name, email]
    );
    return rows[0];
  });
}

export async function getUsers(limit: number = 100) {
  return await read(async (pool) => {
    const { rows } = await pool.query(
      'SELECT * FROM users ORDER BY created_at DESC LIMIT $1',
      [limit]
    );
    return rows;
  });
}
```

---

## Pattern 4: Database Sharding with Neon Branches

**Use Case**: Horizontal scaling, geographic distribution

### Shard Router Implementation

```typescript
interface ShardConfig {
  shardKey: string;
  connectionString: string;
}

export class ShardRouter {
  private shards: ShardConfig[];

  constructor(shards: ShardConfig[]) {
    this.shards = shards;
  }

  getShardForKey(key: string): Pool {
    // Simple hash-based sharding
    const hash = this.hashCode(key);
    const shardIndex = Math.abs(hash) % this.shards.length;
    const shard = this.shards[shardIndex];

    return new Pool({ connectionString: shard.connectionString });
  }

  private hashCode(str: string): number {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = (hash << 5) - hash + char;
      hash |= 0; // Convert to 32bit integer
    }
    return hash;
  }

  async queryAllShards<T>(queryFn: (pool: Pool) => Promise<T[]>): Promise<T[]> {
    const results = await Promise.all(
      this.shards.map(async (shard) => {
        const pool = new Pool({ connectionString: shard.connectionString });
        try {
          return await queryFn(pool);
        } finally {
          await pool.end();
        }
      })
    );

    return results.flat();
  }
}

// Usage
const router = new ShardRouter([
  { shardKey: 'shard-0', connectionString: process.env.SHARD_0_URL! },
  { shardKey: 'shard-1', connectionString: process.env.SHARD_1_URL! },
  { shardKey: 'shard-2', connectionString: process.env.SHARD_2_URL! },
]);

export async function getUserById(userId: string) {
  const pool = router.getShardForKey(userId);
  try {
    const { rows } = await pool.query('SELECT * FROM users WHERE id = $1', [userId]);
    return rows[0];
  } finally {
    await pool.end();
  }
}

export async function getAllUsers() {
  return await router.queryAllShards(async (pool) => {
    const { rows } = await pool.query('SELECT * FROM users');
    return rows;
  });
}
```

---

## Pattern 5: Change Data Capture (CDC) with Triggers

**Use Case**: Real-time data synchronization, cache invalidation

### PostgreSQL Trigger Setup

```sql
-- Audit log table
CREATE TABLE audit_log (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  table_name VARCHAR(100) NOT NULL,
  operation VARCHAR(10) NOT NULL,
  old_data JSONB,
  new_data JSONB,
  changed_by UUID,
  changed_at TIMESTAMP DEFAULT NOW()
);

-- Generic audit trigger function
CREATE OR REPLACE FUNCTION audit_trigger_func()
RETURNS TRIGGER AS $$
BEGIN
  IF (TG_OP = 'DELETE') THEN
    INSERT INTO audit_log (table_name, operation, old_data)
    VALUES (TG_TABLE_NAME, TG_OP, row_to_json(OLD));
    RETURN OLD;
  ELSIF (TG_OP = 'UPDATE') THEN
    INSERT INTO audit_log (table_name, operation, old_data, new_data)
    VALUES (TG_TABLE_NAME, TG_OP, row_to_json(OLD), row_to_json(NEW));
    RETURN NEW;
  ELSIF (TG_OP = 'INSERT') THEN
    INSERT INTO audit_log (table_name, operation, new_data)
    VALUES (TG_TABLE_NAME, TG_OP, row_to_json(NEW));
    RETURN NEW;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to tables
CREATE TRIGGER users_audit
AFTER INSERT OR UPDATE OR DELETE ON users
FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();
```

### CDC Consumer (Application Side)

```typescript
import { Pool } from '@neondatabase/serverless';

const pool = new Pool({ connectionString: process.env.DATABASE_URL! });

export async function getRecentChanges(
  tableName: string,
  since: Date,
  limit: number = 100
): Promise<any[]> {
  const { rows } = await pool.query(
    `SELECT * FROM audit_log
     WHERE table_name = $1 AND changed_at > $2
     ORDER BY changed_at DESC
     LIMIT $3`,
    [tableName, since, limit]
  );

  return rows;
}

// Real-time change listener (polling pattern)
export class ChangeListener {
  private interval: NodeJS.Timeout | null = null;
  private lastCheck: Date = new Date();

  constructor(
    private tableName: string,
    private onChanges: (changes: any[]) => void,
    private pollIntervalMs: number = 1000
  ) {}

  start(): void {
    this.interval = setInterval(async () => {
      const changes = await getRecentChanges(this.tableName, this.lastCheck);

      if (changes.length > 0) {
        this.onChanges(changes);
        this.lastCheck = new Date();
      }
    }, this.pollIntervalMs);
  }

  stop(): void {
    if (this.interval) {
      clearInterval(this.interval);
      this.interval = null;
    }
  }
}

// Usage
const listener = new ChangeListener('users', (changes) => {
  console.log('User changes detected:', changes);
  // Invalidate cache, trigger webhooks, etc.
});

listener.start();
```

---

## Pattern 6: Temporal Tables for Time-Travel Queries

**Use Case**: Historical data analysis, compliance, debugging

### Temporal Table Setup

```sql
CREATE TABLE products (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  price DECIMAL(10, 2) NOT NULL,
  valid_from TIMESTAMP NOT NULL DEFAULT NOW(),
  valid_to TIMESTAMP NOT NULL DEFAULT 'infinity'
);

CREATE INDEX idx_products_temporal ON products(id, valid_from, valid_to);

-- Insert with temporal tracking
CREATE OR REPLACE FUNCTION insert_product_version()
RETURNS TRIGGER AS $$
BEGIN
  -- Close previous version
  UPDATE products
  SET valid_to = NOW()
  WHERE id = NEW.id AND valid_to = 'infinity';

  -- Insert new version
  NEW.valid_from = NOW();
  NEW.valid_to = 'infinity';
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER product_versioning
BEFORE INSERT OR UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION insert_product_version();
```

### Time-Travel Queries

```typescript
export async function getProductAtTime(
  productId: string,
  asOfDate: Date
): Promise<any> {
  const { rows } = await pool.query(
    `SELECT * FROM products
     WHERE id = $1
       AND valid_from <= $2
       AND valid_to > $2
     LIMIT 1`,
    [productId, asOfDate]
  );

  return rows[0];
}

export async function getProductHistory(productId: string): Promise<any[]> {
  const { rows } = await pool.query(
    `SELECT * FROM products
     WHERE id = $1
     ORDER BY valid_from DESC`,
    [productId]
  );

  return rows;
}
```

---

## Summary: Advanced Pattern Selection

| Pattern | Use Case | Complexity | Scalability | Data Consistency |
|---------|----------|------------|-------------|------------------|
| Multi-Tenant RLS | SaaS apps | Medium | High | Strong |
| Event Sourcing | Audit trails | High | High | Eventual |
| Read Replicas | Read-heavy | Low | Very High | Eventual |
| Sharding | Massive scale | High | Extreme | Eventual |
| CDC | Real-time sync | Medium | High | Eventual |
| Temporal Tables | History tracking | Medium | Medium | Strong |

---

**Last Updated**: 2025-11-22
**Neon Version**: Latest
**Total Patterns**: 6 enterprise-grade architectures
