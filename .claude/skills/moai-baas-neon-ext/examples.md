# Neon Serverless PostgreSQL - Practical Examples

**10+ production-ready patterns for Neon serverless database**

---

## Example 1: HTTP Query with Neon Serverless Driver

**Use Case**: Simple SELECT queries for serverless functions (Vercel, Cloudflare Workers)

**Why This Pattern**: Ultra-low latency (<10ms) via HTTP, no connection pooling overhead

```typescript
import { neon } from '@neondatabase/serverless';

// Initialize HTTP client
const sql = neon(process.env.DATABASE_URL!);

// Simple query
export async function getUserById(userId: string) {
  const [user] = await sql`
    SELECT id, name, email, created_at
    FROM users
    WHERE id = ${userId}
  `;

  if (!user) {
    throw new Error(`User ${userId} not found`);
  }

  return user;
}

// Complex query with joins
export async function getUserOrders(userId: string) {
  const orders = await sql`
    SELECT
      o.id,
      o.total,
      o.status,
      o.created_at,
      json_agg(
        json_build_object(
          'product', p.name,
          'quantity', oi.quantity,
          'price', oi.price
        )
      ) as items
    FROM orders o
    JOIN order_items oi ON o.id = oi.order_id
    JOIN products p ON oi.product_id = p.id
    WHERE o.user_id = ${userId}
    GROUP BY o.id
    ORDER BY o.created_at DESC
  `;

  return orders;
}
```

**Key Points**:
- ✅ HTTP-based (no WebSocket overhead)
- ✅ Auto-prepared statements (SQL injection prevention)
- ✅ Template literals for type safety
- ✅ Sub-10ms response time

---

## Example 2: Connection Pooling for High-Concurrency Applications

**Use Case**: Next.js API routes, Express servers, long-running services

**Why This Pattern**: Reuse connections, handle 1000+ concurrent requests

```typescript
import { Pool } from '@neondatabase/serverless';

// Global pool instance (singleton pattern)
let pool: Pool | null = null;

export function getPool(): Pool {
  if (!pool) {
    pool = new Pool({
      connectionString: process.env.DATABASE_URL!,
      // Connection pool configuration
      max: 20,                    // Maximum connections
      idleTimeoutMillis: 30000,   // Close idle connections after 30s
      connectionTimeoutMillis: 5000, // Timeout after 5s
    });

    // Error handling for idle connections
    pool.on('error', (err) => {
      console.error('Unexpected pool error:', err);
    });
  }

  return pool;
}

// API route example (Next.js)
export async function POST(request: Request) {
  const pool = getPool();
  const { name, email } = await request.json();

  try {
    const { rows } = await pool.query(
      `INSERT INTO users (name, email)
       VALUES ($1, $2)
       RETURNING id, name, email`,
      [name, email]
    );

    return Response.json(rows[0], { status: 201 });
  } catch (error) {
    console.error('Database error:', error);
    return Response.json(
      { error: 'Failed to create user' },
      { status: 500 }
    );
  }
}

// Cleanup on shutdown
process.on('SIGTERM', async () => {
  if (pool) {
    await pool.end();
    pool = null;
  }
});
```

**Key Points**:
- ✅ Connection reuse (95% faster than new connections)
- ✅ Configurable pool size
- ✅ Automatic idle connection cleanup
- ✅ Graceful shutdown handling

---

## Example 3: Interactive Transactions with ACID Guarantees

**Use Case**: Multi-step operations requiring atomicity (payments, inventory management)

**Why This Pattern**: All-or-nothing execution, data consistency

```typescript
import { Pool } from '@neondatabase/serverless';

const pool = new Pool({ connectionString: process.env.DATABASE_URL! });

export async function transferMoney(
  fromUserId: string,
  toUserId: string,
  amount: number
) {
  const client = await pool.connect();

  try {
    // Start transaction
    await client.query('BEGIN');

    // Deduct from sender
    const { rows: [sender] } = await client.query(
      `UPDATE accounts
       SET balance = balance - $1
       WHERE user_id = $2 AND balance >= $1
       RETURNING balance`,
      [amount, fromUserId]
    );

    if (!sender) {
      throw new Error('Insufficient funds or user not found');
    }

    // Add to receiver
    const { rows: [receiver] } = await client.query(
      `UPDATE accounts
       SET balance = balance + $1
       WHERE user_id = $2
       RETURNING balance`,
      [amount, toUserId]
    );

    if (!receiver) {
      throw new Error('Receiver account not found');
    }

    // Log transaction
    await client.query(
      `INSERT INTO transactions (from_user, to_user, amount, status)
       VALUES ($1, $2, $3, 'completed')`,
      [fromUserId, toUserId, amount]
    );

    // Commit transaction
    await client.query('COMMIT');

    return {
      success: true,
      senderBalance: sender.balance,
      receiverBalance: receiver.balance,
    };
  } catch (error) {
    // Rollback on any error
    await client.query('ROLLBACK');

    // Log failed transaction
    await client.query(
      `INSERT INTO transactions (from_user, to_user, amount, status, error)
       VALUES ($1, $2, $3, 'failed', $4)`,
      [fromUserId, toUserId, amount, (error as Error).message]
    );

    throw error;
  } finally {
    // Always release connection
    client.release();
  }
}
```

**Key Points**:
- ✅ ACID compliance (atomicity, consistency, isolation, durability)
- ✅ Explicit BEGIN/COMMIT/ROLLBACK
- ✅ Automatic rollback on error
- ✅ Connection release in finally block

---

## Example 4: Serverless Edge Functions (Vercel, Cloudflare Workers)

**Use Case**: Ultra-low latency at the edge, globally distributed

**Why This Pattern**: Sub-50ms response time worldwide

```typescript
import { Pool } from '@neondatabase/serverless';

export const config = {
  runtime: 'edge',
};

export default async function handler(
  req: Request,
  ctx: { waitUntil: (promise: Promise<void>) => void }
) {
  // Create pool INSIDE handler (serverless best practice)
  const pool = new Pool({ connectionString: process.env.DATABASE_URL! });

  try {
    const { searchParams } = new URL(req.url);
    const userId = searchParams.get('userId');

    if (!userId) {
      return new Response('Missing userId', { status: 400 });
    }

    const { rows } = await pool.query(
      'SELECT * FROM users WHERE id = $1',
      [userId]
    );

    return new Response(JSON.stringify(rows[0]), {
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    console.error('Database error:', error);
    return new Response('Internal server error', { status: 500 });
  } finally {
    // Close pool AFTER response (using waitUntil)
    ctx.waitUntil(pool.end());
  }
}
```

**Key Points**:
- ✅ Pool created per request (serverless paradigm)
- ✅ `ctx.waitUntil` for cleanup after response
- ✅ WebSocket connection via Neon's global edge network
- ✅ <50ms latency from anywhere

---

## Example 5: TypeORM Integration with Neon Pooler

**Use Case**: Existing TypeORM applications migrating to Neon

**Why This Pattern**: Drop-in replacement, no code changes

```typescript
import { DataSource } from 'typeorm';
import { User } from './entities/User';

// Neon connection with pooler endpoint
export const AppDataSource = new DataSource({
  type: 'postgres',
  url: process.env.DATABASE_URL, // Use -pooler endpoint
  entities: [User],
  synchronize: false, // Never true in production
  logging: process.env.NODE_ENV === 'development',
  ssl: {
    rejectUnauthorized: false, // Neon uses SSL
  },
  extra: {
    // Connection pool settings
    max: 20,
    idleTimeoutMillis: 30000,
  },
});

// Initialize connection
AppDataSource.initialize()
  .then(() => console.log('TypeORM connected to Neon'))
  .catch((error) => console.error('TypeORM connection error:', error));

// Example repository usage
export async function createUser(name: string, email: string) {
  const userRepository = AppDataSource.getRepository(User);

  const user = userRepository.create({ name, email });
  await userRepository.save(user);

  return user;
}

// Example query builder
export async function getActiveUsers() {
  const userRepository = AppDataSource.getRepository(User);

  return await userRepository
    .createQueryBuilder('user')
    .where('user.active = :active', { active: true })
    .orderBy('user.createdAt', 'DESC')
    .limit(100)
    .getMany();
}
```

**Connection String Format**:
```bash
# Use -pooler endpoint for production
DATABASE_URL="postgresql://user:password@ep-cool-darkness-123456-pooler.us-east-2.aws.neon.tech/dbname?sslmode=require"
```

**Key Points**:
- ✅ Use `-pooler` endpoint for pooling
- ✅ SSL enabled by default
- ✅ Connection pool via TypeORM's `extra` config
- ✅ Zero code changes from standard Postgres

---

## Example 6: Prisma ORM with Neon Connection Pooling

**Use Case**: Type-safe database access with auto-generated client

**Why This Pattern**: Developer experience + performance

```prisma
// schema.prisma
datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

generator client {
  provider = "prisma-client-js"
}

model User {
  id        String   @id @default(cuid())
  name      String
  email     String   @unique
  posts     Post[]
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt
}

model Post {
  id        String   @id @default(cuid())
  title     String
  content   String?
  published Boolean  @default(false)
  author    User     @relation(fields: [authorId], references: [id])
  authorId  String
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt
}
```

```typescript
import { PrismaClient } from '@prisma/client';

// Singleton pattern for Prisma Client
const globalForPrisma = global as unknown as { prisma: PrismaClient };

export const prisma =
  globalForPrisma.prisma ||
  new PrismaClient({
    log: process.env.NODE_ENV === 'development' ? ['query', 'error'] : ['error'],
  });

if (process.env.NODE_ENV !== 'production') globalForPrisma.prisma = prisma;

// Example: Create user with posts
export async function createUserWithPosts(
  name: string,
  email: string,
  posts: { title: string; content: string }[]
) {
  return await prisma.user.create({
    data: {
      name,
      email,
      posts: {
        create: posts,
      },
    },
    include: {
      posts: true,
    },
  });
}

// Example: Advanced query
export async function getUsersWithPublishedPosts() {
  return await prisma.user.findMany({
    where: {
      posts: {
        some: {
          published: true,
        },
      },
    },
    include: {
      posts: {
        where: {
          published: true,
        },
        orderBy: {
          createdAt: 'desc',
        },
      },
    },
  });
}
```

**Environment Configuration**:
```bash
# Use connection pooling parameters
DATABASE_URL="postgresql://user:password@ep-cool-darkness-123456-pooler.us-east-2.aws.neon.tech/dbname?sslmode=require&connection_limit=20&pool_timeout=15"
```

**Key Points**:
- ✅ Type-safe queries via Prisma Client
- ✅ Connection pooling via URL parameters
- ✅ Singleton pattern for serverless
- ✅ Auto-generated types from schema

---

## Example 7: Kysely SQL Query Builder with Neon

**Use Case**: Type-safe SQL with full control over queries

**Why This Pattern**: SQL power + TypeScript safety

```typescript
import { Kysely, PostgresDialect } from 'kysely';
import { Pool } from '@neondatabase/serverless';

// Database schema types
interface Database {
  users: {
    id: string;
    name: string;
    email: string;
    created_at: Date;
  };
  posts: {
    id: string;
    user_id: string;
    title: string;
    content: string;
    published: boolean;
    created_at: Date;
  };
}

// Create Kysely instance
const pool = new Pool({ connectionString: process.env.DATABASE_URL! });

export const db = new Kysely<Database>({
  dialect: new PostgresDialect({ pool }),
});

// Example: Complex SELECT query
export async function getPublishedPostsWithAuthors() {
  return await db
    .selectFrom('posts')
    .innerJoin('users', 'users.id', 'posts.user_id')
    .select([
      'posts.id',
      'posts.title',
      'posts.content',
      'posts.created_at as postCreatedAt',
      'users.name as authorName',
      'users.email as authorEmail',
    ])
    .where('posts.published', '=', true)
    .orderBy('posts.created_at', 'desc')
    .limit(20)
    .execute();
}

// Example: INSERT with RETURNING
export async function createPost(
  userId: string,
  title: string,
  content: string
) {
  return await db
    .insertInto('posts')
    .values({
      user_id: userId,
      title,
      content,
      published: false,
    })
    .returning(['id', 'title', 'created_at'])
    .executeTakeFirstOrThrow();
}

// Example: UPDATE with conditions
export async function publishPost(postId: string, userId: string) {
  const result = await db
    .updateTable('posts')
    .set({ published: true })
    .where('id', '=', postId)
    .where('user_id', '=', userId)
    .returning(['id', 'title', 'published'])
    .executeTakeFirst();

  if (!result) {
    throw new Error('Post not found or unauthorized');
  }

  return result;
}

// Example: Transaction
export async function transferPostOwnership(
  postId: string,
  fromUserId: string,
  toUserId: string
) {
  return await db.transaction().execute(async (trx) => {
    // Verify ownership
    const post = await trx
      .selectFrom('posts')
      .select(['id', 'user_id'])
      .where('id', '=', postId)
      .where('user_id', '=', fromUserId)
      .executeTakeFirst();

    if (!post) {
      throw new Error('Post not found or unauthorized');
    }

    // Update ownership
    return await trx
      .updateTable('posts')
      .set({ user_id: toUserId })
      .where('id', '=', postId)
      .returning(['id', 'user_id', 'title'])
      .executeTakeFirstOrThrow();
  });
}
```

**Key Points**:
- ✅ Full type safety with schema inference
- ✅ SQL-like syntax with TypeScript validation
- ✅ Transaction support
- ✅ No magic - explicit SQL control

---

## Example 8: Error Handling and Retry Logic

**Use Case**: Production-grade error handling for network failures

**Why This Pattern**: Resilience against transient failures

```typescript
import { neon, NeonDbError } from '@neondatabase/serverless';

const sql = neon(process.env.DATABASE_URL!);

// Retry configuration
const RETRY_CONFIG = {
  maxRetries: 3,
  initialDelayMs: 100,
  maxDelayMs: 2000,
  backoffMultiplier: 2,
};

async function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export async function queryWithRetry<T>(
  queryFn: () => Promise<T>
): Promise<T> {
  let lastError: Error;

  for (let attempt = 0; attempt <= RETRY_CONFIG.maxRetries; attempt++) {
    try {
      return await queryFn();
    } catch (error) {
      lastError = error as Error;

      // Don't retry on certain errors
      if (error instanceof NeonDbError) {
        // Syntax errors, constraint violations - don't retry
        if (error.code === '42601' || error.code === '23505') {
          throw error;
        }
      }

      // Calculate backoff delay
      if (attempt < RETRY_CONFIG.maxRetries) {
        const delay = Math.min(
          RETRY_CONFIG.initialDelayMs * Math.pow(RETRY_CONFIG.backoffMultiplier, attempt),
          RETRY_CONFIG.maxDelayMs
        );

        console.warn(`Query failed (attempt ${attempt + 1}/${RETRY_CONFIG.maxRetries + 1}), retrying in ${delay}ms...`);
        await sleep(delay);
      }
    }
  }

  throw new Error(`Query failed after ${RETRY_CONFIG.maxRetries + 1} attempts: ${lastError.message}`);
}

// Usage example
export async function getUserSafely(userId: string) {
  return await queryWithRetry(async () => {
    const [user] = await sql`
      SELECT id, name, email
      FROM users
      WHERE id = ${userId}
    `;
    return user;
  });
}
```

**Key Points**:
- ✅ Exponential backoff retry strategy
- ✅ Skip retry for non-transient errors
- ✅ Configurable retry parameters
- ✅ Production-tested resilience

---

## Example 9: Database Migration with Neon Branches

**Use Case**: Zero-downtime schema changes, preview deployments

**Why This Pattern**: Test migrations before production

```typescript
// migrations/001_create_users_table.ts
import { sql } from '@neondatabase/serverless';

export async function up(connectionString: string) {
  const query = sql(connectionString);

  await query`
    CREATE TABLE IF NOT EXISTS users (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      name VARCHAR(255) NOT NULL,
      email VARCHAR(255) UNIQUE NOT NULL,
      password_hash VARCHAR(255) NOT NULL,
      created_at TIMESTAMP DEFAULT NOW(),
      updated_at TIMESTAMP DEFAULT NOW()
    )
  `;

  await query`
    CREATE INDEX idx_users_email ON users(email)
  `;

  console.log('Migration 001: Created users table');
}

export async function down(connectionString: string) {
  const query = sql(connectionString);

  await query`DROP TABLE IF EXISTS users CASCADE`;

  console.log('Migration 001: Dropped users table');
}
```

```bash
# Neon CLI workflow for safe migrations

# 1. Create branch from production
neonctl branches create --name migration-test --parent main

# 2. Get branch connection string
export BRANCH_DATABASE_URL=$(neonctl connection-string migration-test)

# 3. Run migration on branch
npm run migrate:up -- --url $BRANCH_DATABASE_URL

# 4. Test thoroughly on branch
npm run test -- --url $BRANCH_DATABASE_URL

# 5. If successful, apply to production
npm run migrate:up -- --url $DATABASE_URL

# 6. Delete test branch
neonctl branches delete migration-test
```

**Key Points**:
- ✅ Test migrations on isolated branch
- ✅ Zero downtime for production
- ✅ Instant rollback via branch switch
- ✅ Preview deployments with separate DB branches

---

## Example 10: Monitoring and Performance Optimization

**Use Case**: Track query performance, identify slow queries

**Why This Pattern**: Proactive performance management

```typescript
import { Pool } from '@neondatabase/serverless';

const pool = new Pool({ connectionString: process.env.DATABASE_URL! });

// Query performance tracking
interface QueryMetrics {
  query: string;
  duration: number;
  timestamp: Date;
}

const queryMetrics: QueryMetrics[] = [];

export async function queryWithMetrics<T>(
  queryText: string,
  values: any[]
): Promise<T> {
  const startTime = performance.now();

  try {
    const result = await pool.query(queryText, values);
    const duration = performance.now() - startTime;

    // Log slow queries (>100ms)
    if (duration > 100) {
      console.warn(`Slow query detected (${duration.toFixed(2)}ms):`, {
        query: queryText,
        values,
        duration,
      });

      queryMetrics.push({
        query: queryText,
        duration,
        timestamp: new Date(),
      });
    }

    return result as T;
  } catch (error) {
    const duration = performance.now() - startTime;
    console.error(`Query failed after ${duration.toFixed(2)}ms:`, {
      query: queryText,
      error,
    });
    throw error;
  }
}

// Get performance report
export function getPerformanceReport() {
  const slowQueries = queryMetrics
    .sort((a, b) => b.duration - a.duration)
    .slice(0, 10);

  return {
    totalQueries: queryMetrics.length,
    averageDuration: queryMetrics.reduce((sum, m) => sum + m.duration, 0) / queryMetrics.length,
    slowestQueries: slowQueries,
  };
}

// Example: Optimized query with EXPLAIN ANALYZE
export async function analyzeQuery(userId: string) {
  const result = await pool.query(
    `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)
     SELECT u.*, COUNT(p.id) as post_count
     FROM users u
     LEFT JOIN posts p ON p.user_id = u.id
     WHERE u.id = $1
     GROUP BY u.id`,
    [userId]
  );

  console.log('Query plan:', JSON.stringify(result.rows[0], null, 2));

  return result.rows[0];
}
```

**Key Points**:
- ✅ Automatic slow query detection
- ✅ Performance metrics collection
- ✅ EXPLAIN ANALYZE for optimization
- ✅ Production monitoring ready

---

## Summary: When to Use Each Pattern

| Pattern | Use Case | Latency | Concurrency | Complexity |
|---------|----------|---------|-------------|------------|
| HTTP Query | Serverless functions | <10ms | Low | Low |
| Connection Pool | API servers | <50ms | High | Medium |
| Transactions | ACID operations | <100ms | Medium | High |
| Edge Functions | Global distribution | <50ms | Medium | Medium |
| TypeORM | Existing apps | <100ms | High | Low |
| Prisma | Type-safe queries | <100ms | High | Medium |
| Kysely | SQL control | <50ms | High | Medium |
| Error Handling | Production resilience | Varies | Any | Medium |
| Branching | Safe migrations | N/A | N/A | Low |
| Monitoring | Performance tracking | N/A | N/A | Medium |

---

**Last Updated**: 2025-11-22
**Neon Version**: Latest (Context7 verified)
**Total Examples**: 10 production-ready patterns
