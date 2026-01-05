# Connection Pooling for Serverless

## Overview

Neon provides built-in connection pooling optimized for serverless environments, eliminating connection overhead and enabling edge runtime compatibility.

---

## Connection Types

### Direct Connection

Purpose: Database migrations, admin operations, session-based work
Format: postgresql://user:pass@ep-xxx.region.neon.tech/dbname?sslmode=require
Characteristics: Full PostgreSQL protocol, session persistence, limited connections

### Pooled Connection

Purpose: Application queries, serverless functions, edge runtimes
Format: postgresql://user:pass@ep-xxx-pooler.region.neon.tech/dbname?sslmode=require
Characteristics: Connection multiplexing, HTTP-based driver support, high concurrency

---

## Environment Configuration

### Dual Connection Setup

```env
# Direct connection for migrations and admin tasks
DATABASE_URL=postgresql://user:pass@ep-xxx.region.neon.tech/dbname?sslmode=require

# Pooled connection for application queries
DATABASE_URL_POOLED=postgresql://user:pass@ep-xxx-pooler.region.neon.tech/dbname?sslmode=require
```

### Connection Selection Logic

```typescript
// Use appropriate connection based on context
function getConnectionString(context: 'migration' | 'application' | 'edge'): string {
  switch (context) {
    case 'migration':
      // Direct connection for schema changes
      return process.env.DATABASE_URL!
    case 'application':
      // Pooled connection for general queries
      return process.env.DATABASE_URL_POOLED!
    case 'edge':
      // Pooled connection required for edge runtimes
      return process.env.DATABASE_URL_POOLED!
    default:
      return process.env.DATABASE_URL_POOLED!
  }
}
```

---

## Serverless Driver Configuration

### HTTP-Based Driver

```typescript
import { neon } from '@neondatabase/serverless'

// HTTP driver - ideal for serverless and edge
const sql = neon(process.env.DATABASE_URL_POOLED!)

// Simple query
const users = await sql`SELECT * FROM users LIMIT 10`

// Parameterized query
const userId = 'user-123'
const user = await sql`SELECT * FROM users WHERE id = ${userId}`
```

### WebSocket-Based Driver

```typescript
import { Pool, neonConfig } from '@neondatabase/serverless'
import ws from 'ws'

// Required for Node.js environments
neonConfig.webSocketConstructor = ws

// WebSocket pool for session-based operations
const pool = new Pool({
  connectionString: process.env.DATABASE_URL_POOLED!,
  max: 10  // Maximum connections in local pool
})

const client = await pool.connect()
try {
  await client.query('BEGIN')
  // Transaction operations
  await client.query('COMMIT')
} finally {
  client.release()
}
```

---

## Edge Runtime Integration

### Vercel Edge Functions

```typescript
import { neon } from '@neondatabase/serverless'

export const config = {
  runtime: 'edge'
}

export default async function handler(request: Request) {
  // Must use pooled connection for edge
  const sql = neon(process.env.DATABASE_URL_POOLED!)

  const { searchParams } = new URL(request.url)
  const limit = parseInt(searchParams.get('limit') || '10')

  const users = await sql`
    SELECT id, name, email
    FROM users
    WHERE active = true
    ORDER BY created_at DESC
    LIMIT ${limit}
  `

  return Response.json(users, {
    headers: {
      'Cache-Control': 'public, s-maxage=60'
    }
  })
}
```

### Cloudflare Workers

```typescript
import { neon } from '@neondatabase/serverless'

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const sql = neon(env.DATABASE_URL_POOLED)

    const data = await sql`SELECT COUNT(*) as count FROM users`

    return Response.json(data)
  }
}
```

### Next.js App Router

```typescript
// app/api/users/route.ts
import { neon } from '@neondatabase/serverless'
import { NextResponse } from 'next/server'

export const runtime = 'edge'

export async function GET() {
  const sql = neon(process.env.DATABASE_URL_POOLED!)

  const users = await sql`
    SELECT id, name, email, created_at
    FROM users
    WHERE active = true
    LIMIT 100
  `

  return NextResponse.json(users)
}
```

---

## Pool Configuration

### Pool Sizing Guidelines

Development Environment:
- Max Connections: 5-10
- Reason: Limited concurrent requests, cost optimization

Staging Environment:
- Max Connections: 10-25
- Reason: Moderate testing load, simulates production

Production Environment:
- Max Connections: 25-100
- Reason: Handle concurrent requests, account for connection overhead

### WebSocket Pool Configuration

```typescript
import { Pool, neonConfig } from '@neondatabase/serverless'
import ws from 'ws'

neonConfig.webSocketConstructor = ws

const pool = new Pool({
  connectionString: process.env.DATABASE_URL_POOLED!,
  max: 25,                    // Maximum connections
  idleTimeoutMillis: 30000,   // Close idle connections after 30s
  connectionTimeoutMillis: 5000  // Connection timeout
})

// Graceful shutdown
process.on('SIGTERM', async () => {
  await pool.end()
  process.exit(0)
})
```

---

## Connection Pooling with ORMs

### Drizzle with Pooled Connection

```typescript
import { neon } from '@neondatabase/serverless'
import { drizzle } from 'drizzle-orm/neon-http'
import * as schema from './schema'

// HTTP driver with pooled connection
const sql = neon(process.env.DATABASE_URL_POOLED!)
export const db = drizzle(sql, { schema })

// All queries go through pooled connection
const users = await db.select().from(schema.users)
```

### Prisma with Pooled Connection

```typescript
import { Pool, neonConfig } from '@neondatabase/serverless'
import { PrismaNeon } from '@prisma/adapter-neon'
import { PrismaClient } from '@prisma/client'
import ws from 'ws'

neonConfig.webSocketConstructor = ws

const pool = new Pool({ connectionString: process.env.DATABASE_URL_POOLED! })
const adapter = new PrismaNeon(pool)

export const prisma = new PrismaClient({
  adapter,
  log: process.env.NODE_ENV === 'development' ? ['query'] : []
})
```

---

## Transaction Handling

### HTTP Driver Transactions

```typescript
import { neon } from '@neondatabase/serverless'

const sql = neon(process.env.DATABASE_URL_POOLED!)

// Transaction with array of statements
const result = await sql.transaction([
  sql`INSERT INTO orders (user_id, total) VALUES (${userId}, ${total}) RETURNING id`,
  sql`UPDATE users SET order_count = order_count + 1 WHERE id = ${userId}`,
  sql`INSERT INTO order_events (order_id, event) VALUES (${orderId}, 'created')`
])
```

### WebSocket Pool Transactions

```typescript
import { Pool, neonConfig } from '@neondatabase/serverless'
import ws from 'ws'

neonConfig.webSocketConstructor = ws
const pool = new Pool({ connectionString: process.env.DATABASE_URL_POOLED! })

async function transferFunds(fromId: string, toId: string, amount: number) {
  const client = await pool.connect()

  try {
    await client.query('BEGIN')

    // Check balance
    const { rows } = await client.query(
      'SELECT balance FROM accounts WHERE id = $1 FOR UPDATE',
      [fromId]
    )

    if (rows[0].balance < amount) {
      throw new Error('Insufficient funds')
    }

    // Perform transfer
    await client.query(
      'UPDATE accounts SET balance = balance - $1 WHERE id = $2',
      [amount, fromId]
    )
    await client.query(
      'UPDATE accounts SET balance = balance + $1 WHERE id = $2',
      [amount, toId]
    )

    await client.query('COMMIT')
  } catch (error) {
    await client.query('ROLLBACK')
    throw error
  } finally {
    client.release()
  }
}
```

---

## Error Handling

### Connection Error Handling

```typescript
import { neon } from '@neondatabase/serverless'

const sql = neon(process.env.DATABASE_URL_POOLED!)

async function queryWithRetry<T>(query: () => Promise<T>, maxRetries = 3): Promise<T> {
  let lastError: Error | null = null

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return await query()
    } catch (error) {
      lastError = error as Error

      // Retry on connection errors
      if (error instanceof Error && error.message.includes('ECONNREFUSED')) {
        console.log(`Connection attempt ${attempt} failed, retrying...`)
        await new Promise(resolve => setTimeout(resolve, 1000 * attempt))
        continue
      }

      // Don't retry on other errors
      throw error
    }
  }

  throw lastError
}

// Usage
const users = await queryWithRetry(() => sql`SELECT * FROM users`)
```

### Pool Exhaustion Handling

```typescript
import { Pool, neonConfig } from '@neondatabase/serverless'
import ws from 'ws'

neonConfig.webSocketConstructor = ws

const pool = new Pool({
  connectionString: process.env.DATABASE_URL_POOLED!,
  max: 25,
  connectionTimeoutMillis: 5000
})

async function queryWithPoolMonitoring<T>(query: () => Promise<T>): Promise<T> {
  const waitingCount = pool.waitingCount
  const totalCount = pool.totalCount

  if (waitingCount > 10) {
    console.warn(`High pool contention: ${waitingCount} waiting, ${totalCount} total`)
  }

  return query()
}
```

---

## Best Practices

### Connection Selection

Always Use Pooled for Application Code: Reduces connection overhead
Use Direct for Migrations: Requires full PostgreSQL protocol
Edge Runtimes Require Pooled: Direct connections not supported

### Performance Optimization

Reuse Connections: Create pool once, reuse across requests
Avoid Connection Leaks: Always release connections in finally blocks
Monitor Pool Metrics: Track waiting count and total connections
Set Appropriate Timeouts: Balance between availability and resource usage

### Security Considerations

SSL Always Required: Neon enforces sslmode=require
Credential Rotation: Rotate database credentials periodically
Environment Variables: Never hardcode connection strings
IP Allowlists: Configure IP restrictions for production

---

Version: 2.0.0
Last Updated: 2026-01-06
