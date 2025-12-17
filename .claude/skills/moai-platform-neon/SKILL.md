---
name: moai-platform-neon
description: Neon serverless PostgreSQL specialist covering auto-scaling, database branching, PITR, and connection pooling. Use when building serverless apps needing PostgreSQL.
version: 1.0.0
category: platform
tags: [neon, postgresql, serverless, branching, auto-scaling]
context7-libraries: [/neondatabase/neon]
related-skills: [moai-platform-supabase, moai-lang-typescript]
allowed-tools: Read, Write, Bash, Grep, Glob
---

# moai-platform-neon: Neon Serverless PostgreSQL Specialist

## Quick Reference (30 seconds)

Neon Serverless PostgreSQL Expertise: Specialized knowledge for Neon serverless PostgreSQL covering auto-scaling, scale-to-zero compute, database branching, Point-in-Time Recovery, and modern ORM integration.

### Core Capabilities

Serverless Compute: Auto-scaling PostgreSQL with scale-to-zero for cost optimization
Database Branching: Instant copy-on-write branches for dev, staging, and preview environments
Point-in-Time Recovery: 30-day PITR with instant restore to any timestamp
Connection Pooling: Built-in connection pooler for serverless and edge compatibility
PostgreSQL 16: Full PostgreSQL 16 compatibility with extensions support

### Quick Decision Guide

- Need serverless PostgreSQL with auto-scaling? Neon
- Need database branching for CI/CD? Neon branching
- Need edge-compatible database? Neon with connection pooling
- Need instant preview environments? Neon branch per PR

### Context7 Library Mapping

Neon: /neondatabase/neon

---

## Implementation Guide

### Setup and Configuration

Package Installation:
```bash
npm install @neondatabase/serverless
npm install drizzle-orm  # Optional: Drizzle ORM
npm install @prisma/client prisma  # Optional: Prisma ORM
```

Environment Configuration:
```env
# Direct connection (for migrations)
DATABASE_URL=postgresql://user:pass@ep-xxx.region.neon.tech/dbname?sslmode=require

# Pooled connection (for serverless/edge)
DATABASE_URL_POOLED=postgresql://user:pass@ep-xxx-pooler.region.neon.tech/dbname?sslmode=require

# Neon API for branching
NEON_API_KEY=neon_api_key_xxx
NEON_PROJECT_ID=project-xxx
```

### Serverless Driver Usage

Basic Query Execution:
```typescript
import { neon } from '@neondatabase/serverless'

const sql = neon(process.env.DATABASE_URL!)

// Simple query
const users = await sql`SELECT * FROM users WHERE active = true`

// Parameterized query (SQL injection safe)
const userId = 'user-123'
const user = await sql`SELECT * FROM users WHERE id = ${userId}`

// Transaction support
const result = await sql.transaction([
  sql`UPDATE accounts SET balance = balance - 100 WHERE id = ${fromId}`,
  sql`UPDATE accounts SET balance = balance + 100 WHERE id = ${toId}`
])
```

WebSocket Connection for Session Persistence:
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
  await client.query('INSERT INTO logs (message) VALUES ($1)', ['Action'])
  await client.query('COMMIT')
} finally {
  client.release()
}
```

### Database Branching

Branch Management API:
```typescript
class NeonBranchManager {
  private apiKey: string
  private projectId: string
  private baseUrl = 'https://console.neon.tech/api/v2'

  constructor(apiKey: string, projectId: string) {
    this.apiKey = apiKey
    this.projectId = projectId
  }

  private async request(path: string, options: RequestInit = {}) {
    const response = await fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json',
        ...options.headers
      }
    })
    if (!response.ok) throw new Error(`Neon API error: ${response.statusText}`)
    return response.json()
  }

  async createBranch(name: string, parentId: string = 'main') {
    return this.request(`/projects/${this.projectId}/branches`, {
      method: 'POST',
      body: JSON.stringify({
        branch: { name, parent_id: parentId }
      })
    })
  }

  async deleteBranch(branchId: string) {
    return this.request(`/projects/${this.projectId}/branches/${branchId}`, {
      method: 'DELETE'
    })
  }

  async listBranches() {
    return this.request(`/projects/${this.projectId}/branches`)
  }

  async getBranchConnectionString(branchId: string) {
    const endpoints = await this.request(
      `/projects/${this.projectId}/branches/${branchId}/endpoints`
    )
    return endpoints.endpoints[0]?.connection_uri
  }
}
```

Preview Branch for Pull Requests:
```typescript
async function createPreviewEnvironment(prNumber: number) {
  const branchManager = new NeonBranchManager(
    process.env.NEON_API_KEY!,
    process.env.NEON_PROJECT_ID!
  )

  // Create branch from main
  const branch = await branchManager.createBranch(`pr-${prNumber}`, 'main')

  // Get connection string
  const connectionString = await branchManager.getBranchConnectionString(branch.branch.id)

  return {
    branchId: branch.branch.id,
    branchName: branch.branch.name,
    connectionString
  }
}

async function cleanupPreviewEnvironment(branchId: string) {
  const branchManager = new NeonBranchManager(
    process.env.NEON_API_KEY!,
    process.env.NEON_PROJECT_ID!
  )
  await branchManager.deleteBranch(branchId)
}
```

### Point-in-Time Recovery

Restore to Specific Timestamp:
```typescript
async function restoreToPoint(timestamp: Date) {
  const branchManager = new NeonBranchManager(
    process.env.NEON_API_KEY!,
    process.env.NEON_PROJECT_ID!
  )

  const response = await fetch(
    `https://console.neon.tech/api/v2/projects/${process.env.NEON_PROJECT_ID}/branches`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${process.env.NEON_API_KEY}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        branch: {
          name: `restore-${timestamp.toISOString().replace(/[:.]/g, '-')}`,
          parent_id: 'main',
          parent_timestamp: timestamp.toISOString()
        }
      })
    }
  )

  return response.json()
}

// Usage: Restore to 1 hour ago
const oneHourAgo = new Date(Date.now() - 60 * 60 * 1000)
const restoredBranch = await restoreToPoint(oneHourAgo)
```

### Drizzle ORM Integration

Schema Definition:
```typescript
// schema.ts
import { pgTable, uuid, text, timestamp, boolean, jsonb } from 'drizzle-orm/pg-core'

export const users = pgTable('users', {
  id: uuid('id').primaryKey().defaultRandom(),
  email: text('email').notNull().unique(),
  name: text('name'),
  createdAt: timestamp('created_at').defaultNow(),
  metadata: jsonb('metadata')
})

export const projects = pgTable('projects', {
  id: uuid('id').primaryKey().defaultRandom(),
  name: text('name').notNull(),
  ownerId: uuid('owner_id').references(() => users.id),
  isPublic: boolean('is_public').default(false),
  createdAt: timestamp('created_at').defaultNow()
})
```

Drizzle Client Setup:
```typescript
// db.ts
import { neon } from '@neondatabase/serverless'
import { drizzle } from 'drizzle-orm/neon-http'
import * as schema from './schema'

const sql = neon(process.env.DATABASE_URL!)
export const db = drizzle(sql, { schema })

// Query examples
const allUsers = await db.select().from(schema.users)

const userProjects = await db
  .select()
  .from(schema.projects)
  .where(eq(schema.projects.ownerId, userId))
  .orderBy(desc(schema.projects.createdAt))
```

### Prisma ORM Integration

Prisma Schema:
```prisma
// schema.prisma
generator client {
  provider = "prisma-client-js"
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
  projects  Project[]
  createdAt DateTime  @default(now())
}

model Project {
  id        String   @id @default(uuid())
  name      String
  owner     User     @relation(fields: [ownerId], references: [id])
  ownerId   String
  isPublic  Boolean  @default(false)
  createdAt DateTime @default(now())
}
```

Prisma with Neon Serverless Driver:
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
const users = await prisma.user.findMany({
  include: { projects: true }
})
```

---

## Advanced Patterns

### Connection Pooling for Edge

Edge Function Configuration:
```typescript
import { neon } from '@neondatabase/serverless'

// Use pooled connection for edge environments
const sql = neon(process.env.DATABASE_URL_POOLED!)

export const config = {
  runtime: 'edge'
}

export default async function handler(request: Request) {
  const users = await sql`SELECT id, name FROM users LIMIT 10`
  return Response.json(users)
}
```

### CI/CD Branch Automation

GitHub Actions Integration:
```yaml
name: Preview Environment

on:
  pull_request:
    types: [opened, synchronize, closed]

jobs:
  create-preview:
    if: github.event.action != 'closed'
    runs-on: ubuntu-latest
    steps:
      - name: Create Neon Branch
        id: create-branch
        run: |
          BRANCH=$(curl -s -X POST \
            -H "Authorization: Bearer ${{ secrets.NEON_API_KEY }}" \
            -H "Content-Type: application/json" \
            -d '{"branch":{"name":"pr-${{ github.event.number }}"}}' \
            "https://console.neon.tech/api/v2/projects/${{ secrets.NEON_PROJECT_ID }}/branches")
          echo "branch_id=$(echo $BRANCH | jq -r '.branch.id')" >> $GITHUB_OUTPUT

  cleanup-preview:
    if: github.event.action == 'closed'
    runs-on: ubuntu-latest
    steps:
      - name: Delete Neon Branch
        run: |
          curl -X DELETE \
            -H "Authorization: Bearer ${{ secrets.NEON_API_KEY }}" \
            "https://console.neon.tech/api/v2/projects/${{ secrets.NEON_PROJECT_ID }}/branches/pr-${{ github.event.number }}"
```

### Auto-Scaling Configuration

Compute Settings via API:
```typescript
async function configureAutoScaling(endpointId: string) {
  const response = await fetch(
    `https://console.neon.tech/api/v2/projects/${process.env.NEON_PROJECT_ID}/endpoints/${endpointId}`,
    {
      method: 'PATCH',
      headers: {
        'Authorization': `Bearer ${process.env.NEON_API_KEY}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        endpoint: {
          autoscaling_limit_min_cu: 0.25,  // Scale to zero
          autoscaling_limit_max_cu: 4,     // Max 4 compute units
          suspend_timeout_seconds: 300     // Suspend after 5 min idle
        }
      })
    }
  )
  return response.json()
}
```

### Migration Workflow

Development to Production:
```typescript
// Run migrations on direct connection (not pooled)
import { migrate } from 'drizzle-orm/neon-http/migrator'
import { neon } from '@neondatabase/serverless'
import { drizzle } from 'drizzle-orm/neon-http'

async function runMigrations() {
  // Use direct connection for migrations
  const sql = neon(process.env.DATABASE_URL!)
  const db = drizzle(sql)

  await migrate(db, { migrationsFolder: './drizzle' })
  console.log('Migrations completed')
}
```

---

## Provider Decision Guide

### When to Use Neon

Serverless Applications: Auto-scaling and scale-to-zero reduce costs
Preview Environments: Instant branching enables per-PR databases
Edge Deployment: Connection pooling works with edge runtimes
Development Workflow: Branch from production for realistic dev data
Cost Optimization: Pay only for active compute time

### When to Consider Alternatives

Need Vector Search: Consider Supabase with pgvector
Need Real-time Subscriptions: Consider Supabase or Convex
Need NoSQL Flexibility: Consider Firestore or Convex
Need Built-in Auth: Consider Supabase

### Pricing Reference (2024)

Free Tier: 3GB storage, 100 compute hours per month
Pro Tier: Usage-based pricing, additional storage and compute
Scale-to-Zero: No charges during idle periods

---

## Works Well With

- moai-platform-supabase - Alternative when RLS or pgvector needed
- moai-lang-typescript - TypeScript patterns for Drizzle and Prisma
- moai-domain-backend - Backend architecture with database integration
- moai-workflow-cicd - CI/CD pipeline integration patterns
- moai-context7-integration - Latest Neon documentation access

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v2.0
Last Updated: 2025-12-07
Technology: Neon Serverless PostgreSQL
