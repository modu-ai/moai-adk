# Neon Platform Code Examples

## Basic Serverless Driver

### Simple Query Execution

```typescript
import { neon } from '@neondatabase/serverless'

const sql = neon(process.env.DATABASE_URL!)

// Simple select
const users = await sql`SELECT * FROM users WHERE active = true`

// Parameterized query (SQL injection safe)
const userId = 'user-123'
const user = await sql`SELECT * FROM users WHERE id = ${userId}`

// Insert with returning
const newUser = await sql`
  INSERT INTO users (email, name)
  VALUES (${email}, ${name})
  RETURNING *
`

// Update with conditions
const updated = await sql`
  UPDATE users
  SET last_login = NOW()
  WHERE id = ${userId}
  RETURNING *
`
```

### Transaction Support

```typescript
import { neon } from '@neondatabase/serverless'

const sql = neon(process.env.DATABASE_URL!)

// Transaction with multiple statements
const result = await sql.transaction([
  sql`UPDATE accounts SET balance = balance - 100 WHERE id = ${fromId}`,
  sql`UPDATE accounts SET balance = balance + 100 WHERE id = ${toId}`,
  sql`INSERT INTO transfers (from_id, to_id, amount) VALUES (${fromId}, ${toId}, 100)`
])
```

---

## WebSocket Connection

### Session-Based Operations

```typescript
import { Pool, neonConfig } from '@neondatabase/serverless'
import ws from 'ws'

// Required for Node.js environments
neonConfig.webSocketConstructor = ws

const pool = new Pool({ connectionString: process.env.DATABASE_URL })

// Use pool for session-based operations
const client = await pool.connect()
try {
  await client.query('BEGIN')
  await client.query('INSERT INTO logs (message) VALUES ($1)', ['Action started'])
  await client.query('UPDATE counters SET value = value + 1 WHERE name = $1', ['actions'])
  await client.query('INSERT INTO logs (message) VALUES ($1)', ['Action completed'])
  await client.query('COMMIT')
} catch (error) {
  await client.query('ROLLBACK')
  throw error
} finally {
  client.release()
}
```

---

## Drizzle ORM Integration

### Complete Schema Definition

```typescript
// schema.ts
import { pgTable, uuid, text, timestamp, boolean, jsonb, integer } from 'drizzle-orm/pg-core'
import { relations } from 'drizzle-orm'

export const users = pgTable('users', {
  id: uuid('id').primaryKey().defaultRandom(),
  email: text('email').notNull().unique(),
  name: text('name'),
  avatarUrl: text('avatar_url'),
  createdAt: timestamp('created_at').defaultNow(),
  updatedAt: timestamp('updated_at').defaultNow(),
  metadata: jsonb('metadata')
})

export const projects = pgTable('projects', {
  id: uuid('id').primaryKey().defaultRandom(),
  name: text('name').notNull(),
  description: text('description'),
  ownerId: uuid('owner_id').references(() => users.id, { onDelete: 'cascade' }),
  isPublic: boolean('is_public').default(false),
  starCount: integer('star_count').default(0),
  createdAt: timestamp('created_at').defaultNow(),
  updatedAt: timestamp('updated_at').defaultNow()
})

export const usersRelations = relations(users, ({ many }) => ({
  projects: many(projects)
}))

export const projectsRelations = relations(projects, ({ one }) => ({
  owner: one(users, {
    fields: [projects.ownerId],
    references: [users.id]
  })
}))
```

### Drizzle Client with Queries

```typescript
// db.ts
import { neon } from '@neondatabase/serverless'
import { drizzle } from 'drizzle-orm/neon-http'
import { eq, desc, and, like } from 'drizzle-orm'
import * as schema from './schema'

const sql = neon(process.env.DATABASE_URL!)
export const db = drizzle(sql, { schema })

// Query examples
async function getUserWithProjects(userId: string) {
  return db.query.users.findFirst({
    where: eq(schema.users.id, userId),
    with: {
      projects: {
        orderBy: desc(schema.projects.createdAt)
      }
    }
  })
}

async function searchProjects(query: string, limit = 10) {
  return db
    .select()
    .from(schema.projects)
    .where(
      and(
        eq(schema.projects.isPublic, true),
        like(schema.projects.name, `%${query}%`)
      )
    )
    .orderBy(desc(schema.projects.starCount))
    .limit(limit)
}

async function createProject(data: { name: string; ownerId: string; description?: string }) {
  const [project] = await db
    .insert(schema.projects)
    .values(data)
    .returning()
  return project
}
```

---

## Prisma ORM Integration

### Prisma Schema

```prisma
// schema.prisma
generator client {
  provider        = "prisma-client-js"
  previewFeatures = ["driverAdapters"]
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id        String    @id @default(uuid())
  email     String    @unique
  name      String?
  avatarUrl String?   @map("avatar_url")
  projects  Project[]
  createdAt DateTime  @default(now()) @map("created_at")
  updatedAt DateTime  @updatedAt @map("updated_at")

  @@map("users")
}

model Project {
  id          String   @id @default(uuid())
  name        String
  description String?
  owner       User     @relation(fields: [ownerId], references: [id], onDelete: Cascade)
  ownerId     String   @map("owner_id")
  isPublic    Boolean  @default(false) @map("is_public")
  starCount   Int      @default(0) @map("star_count")
  createdAt   DateTime @default(now()) @map("created_at")
  updatedAt   DateTime @updatedAt @map("updated_at")

  @@map("projects")
}
```

### Prisma Client Setup

```typescript
// db.ts
import { Pool, neonConfig } from '@neondatabase/serverless'
import { PrismaNeon } from '@prisma/adapter-neon'
import { PrismaClient } from '@prisma/client'

neonConfig.webSocketConstructor = require('ws')

const pool = new Pool({ connectionString: process.env.DATABASE_URL })
const adapter = new PrismaNeon(pool)
export const prisma = new PrismaClient({ adapter })

// Query examples
async function getUserWithProjects(userId: string) {
  return prisma.user.findUnique({
    where: { id: userId },
    include: {
      projects: {
        orderBy: { createdAt: 'desc' }
      }
    }
  })
}

async function searchProjects(query: string, limit = 10) {
  return prisma.project.findMany({
    where: {
      isPublic: true,
      name: { contains: query, mode: 'insensitive' }
    },
    orderBy: { starCount: 'desc' },
    take: limit
  })
}
```

---

## Branch Management

### Complete Branch Manager Class

```typescript
class NeonBranchManager {
  private apiKey: string
  private projectId: string
  private baseUrl = 'https://console.neon.tech/api/v2'

  constructor(apiKey: string, projectId: string) {
    this.apiKey = apiKey
    this.projectId = projectId
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json',
        ...options.headers
      }
    })
    if (!response.ok) {
      const error = await response.text()
      throw new Error(`Neon API error: ${response.status} - ${error}`)
    }
    return response.json()
  }

  async createBranch(name: string, parentId: string = 'main') {
    return this.request<{ branch: Branch; endpoints: Endpoint[] }>(
      `/projects/${this.projectId}/branches`,
      {
        method: 'POST',
        body: JSON.stringify({
          branch: { name, parent_id: parentId }
        })
      }
    )
  }

  async createBranchAtTimestamp(name: string, timestamp: Date, parentId: string = 'main') {
    return this.request<{ branch: Branch; endpoints: Endpoint[] }>(
      `/projects/${this.projectId}/branches`,
      {
        method: 'POST',
        body: JSON.stringify({
          branch: {
            name,
            parent_id: parentId,
            parent_timestamp: timestamp.toISOString()
          }
        })
      }
    )
  }

  async deleteBranch(branchId: string) {
    return this.request<{ branch: Branch }>(
      `/projects/${this.projectId}/branches/${branchId}`,
      { method: 'DELETE' }
    )
  }

  async listBranches() {
    return this.request<{ branches: Branch[] }>(
      `/projects/${this.projectId}/branches`
    )
  }

  async getBranchEndpoints(branchId: string) {
    return this.request<{ endpoints: Endpoint[] }>(
      `/projects/${this.projectId}/branches/${branchId}/endpoints`
    )
  }

  async getBranchConnectionString(branchId: string): Promise<string | undefined> {
    const { endpoints } = await this.getBranchEndpoints(branchId)
    return endpoints[0]?.connection_uri
  }
}

interface Branch {
  id: string
  name: string
  project_id: string
  parent_id: string
  created_at: string
  current_state: string
}

interface Endpoint {
  id: string
  branch_id: string
  host: string
  connection_uri: string
}
```

---

## Edge Function Integration

### Vercel Edge Function

```typescript
import { neon } from '@neondatabase/serverless'

export const config = {
  runtime: 'edge'
}

export default async function handler(request: Request) {
  const sql = neon(process.env.DATABASE_URL_POOLED!)

  const { searchParams } = new URL(request.url)
  const limit = parseInt(searchParams.get('limit') || '10')

  const users = await sql`
    SELECT id, name, email, created_at
    FROM users
    WHERE active = true
    ORDER BY created_at DESC
    LIMIT ${limit}
  `

  return Response.json(users, {
    headers: {
      'Cache-Control': 'public, s-maxage=60, stale-while-revalidate=300'
    }
  })
}
```

### Next.js App Router

```typescript
// app/api/users/route.ts
import { neon } from '@neondatabase/serverless'
import { NextResponse } from 'next/server'

export const runtime = 'edge'

export async function GET(request: Request) {
  const sql = neon(process.env.DATABASE_URL_POOLED!)

  const users = await sql`SELECT * FROM users WHERE active = true LIMIT 100`

  return NextResponse.json(users)
}

export async function POST(request: Request) {
  const sql = neon(process.env.DATABASE_URL_POOLED!)
  const { email, name } = await request.json()

  const [user] = await sql`
    INSERT INTO users (email, name)
    VALUES (${email}, ${name})
    RETURNING *
  `

  return NextResponse.json(user, { status: 201 })
}
```

---

## Migration Workflow

### Drizzle Migrations

```typescript
// migrate.ts
import { migrate } from 'drizzle-orm/neon-http/migrator'
import { neon } from '@neondatabase/serverless'
import { drizzle } from 'drizzle-orm/neon-http'

async function runMigrations() {
  // Use direct connection for migrations (not pooled)
  const sql = neon(process.env.DATABASE_URL!)
  const db = drizzle(sql)

  console.log('Running migrations...')

  await migrate(db, { migrationsFolder: './drizzle' })

  console.log('Migrations completed successfully')
}

runMigrations().catch(console.error)
```

### Drizzle Config

```typescript
// drizzle.config.ts
import type { Config } from 'drizzle-kit'

export default {
  schema: './src/db/schema.ts',
  out: './drizzle',
  driver: 'pg',
  dbCredentials: {
    connectionString: process.env.DATABASE_URL!
  }
} satisfies Config
```

---

Version: 2.0.0
Last Updated: 2026-01-06
