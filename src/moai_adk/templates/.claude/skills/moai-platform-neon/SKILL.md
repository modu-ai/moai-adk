---
name: "moai-platform-neon"
description: "Neon serverless PostgreSQL specialist covering auto-scaling, database branching, PITR, and connection pooling. Use when building serverless apps needing PostgreSQL, implementing preview environments, or optimizing database costs."
version: 2.0.0
category: "platform"
modularized: true
user-invocable: false
tags: ['neon', 'postgresql', 'serverless', 'branching', 'auto-scaling']
context7-libraries: ['/neondatabase/neon']
related-skills: ['moai-platform-supabase', 'moai-lang-typescript', 'moai-domain-database']
updated: 2026-01-08
status: "active"
allowed-tools:
  - Read
  - Write
  - Bash
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# moai-platform-neon: Neon Serverless PostgreSQL Specialist

## Quick Reference

Neon Serverless PostgreSQL Expertise: Specialized knowledge for Neon serverless PostgreSQL covering auto-scaling, scale-to-zero compute, database branching, Point-in-Time Recovery, and modern ORM integration.

### Core Capabilities

Serverless Compute: Auto-scaling PostgreSQL with scale-to-zero for cost optimization
Database Branching: Instant copy-on-write branches for dev, staging, and preview environments
Point-in-Time Recovery: 30-day PITR with instant restore to any timestamp
Connection Pooling: Built-in connection pooler for serverless and edge compatibility
PostgreSQL 16: Full PostgreSQL 16 compatibility with extensions support

### Quick Decision Guide

Need serverless PostgreSQL with auto-scaling: Use Neon
Need database branching for CI/CD: Use Neon branching
Need edge-compatible database: Use Neon with connection pooling
Need instant preview environments: Use Neon branch per PR
Need vector search: Consider Supabase with pgvector instead

### Context7 Documentation Access

To fetch the latest Neon documentation:

Step 1: Resolve the library ID using mcp__context7__resolve-library-id with library name "neondatabase/neon"

Step 2: Fetch documentation using mcp__context7__get-library-docs with the resolved Context7 ID, specifying topics like "branching", "connection pooling", or "auto-scaling"

---

## Module Index

This skill is organized into focused modules for progressive disclosure:

### Core Modules

[Database Branching](modules/branching-workflows.md): Copy-on-write branches for development, preview environments, and CI/CD integration with GitHub Actions

[Auto-Scaling and Compute](modules/auto-scaling.md): Compute unit configuration, scale-to-zero settings, and cost optimization strategies

[Connection Pooling](modules/connection-pooling.md): Serverless connection pooling for edge runtimes, WebSocket configuration, and pool sizing

[PITR and Backups](modules/pitr-backups.md): Point-in-time recovery, branch restoration, and backup strategies

### Supporting Files

[Reference Guide](reference.md): API reference, environment configuration, and provider comparison

[Code Examples](examples.md): Complete working examples for common integration patterns

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

### Drizzle ORM Integration

Schema Definition:
```typescript
import { pgTable, uuid, text, timestamp, boolean, jsonb } from 'drizzle-orm/pg-core'

export const users = pgTable('users', {
  id: uuid('id').primaryKey().defaultRandom(),
  email: text('email').notNull().unique(),
  name: text('name'),
  createdAt: timestamp('created_at').defaultNow(),
  metadata: jsonb('metadata')
})
```

Drizzle Client Setup:
```typescript
import { neon } from '@neondatabase/serverless'
import { drizzle } from 'drizzle-orm/neon-http'
import * as schema from './schema'

const sql = neon(process.env.DATABASE_URL!)
export const db = drizzle(sql, { schema })

// Query examples
const allUsers = await db.select().from(schema.users)
```

### Prisma ORM Integration

Prisma with Neon Serverless Driver:
```typescript
import { Pool, neonConfig } from '@neondatabase/serverless'
import { PrismaNeon } from '@prisma/adapter-neon'
import { PrismaClient } from '@prisma/client'

neonConfig.webSocketConstructor = require('ws')

const pool = new Pool({ connectionString: process.env.DATABASE_URL })
const adapter = new PrismaNeon(pool)
export const prisma = new PrismaClient({ adapter })
```

---

## Provider Decision Guide

### When to Use Neon

Serverless Applications: Auto-scaling and scale-to-zero reduce costs significantly
Preview Environments: Instant branching enables per-PR databases with production data
Edge Deployment: Connection pooling provides edge runtime compatibility
Development Workflow: Branch from production for realistic development data
Cost Optimization: Pay only for active compute time with scale-to-zero

### When to Consider Alternatives

Need Vector Search: Consider Supabase with pgvector or dedicated vector database
Need Real-time Subscriptions: Consider Supabase or Convex for real-time features
Need NoSQL Flexibility: Consider Firestore or Convex for document storage
Need Built-in Auth: Consider Supabase for integrated authentication

### Pricing Reference

Free Tier: 3GB storage, 100 compute hours per month
Pro Tier: Usage-based pricing with additional storage and compute
Scale-to-Zero: No charges during idle periods

---

## Works Well With

- moai-platform-supabase - Alternative when RLS or pgvector needed
- moai-lang-typescript - TypeScript patterns for Drizzle and Prisma
- moai-domain-backend - Backend architecture with database integration
- moai-domain-database - General database patterns and optimization

---

Status: Production Ready
Version: 2.0.0
Generated with: MoAI-ADK Skill Factory v2.0
Last Updated: 2026-01-06
Technology: Neon Serverless PostgreSQL
