# KV, Blob, and Postgres Storage

## Overview

Vercel provides managed storage solutions: KV (Redis-compatible key-value store), Blob (object storage for files), and Postgres (managed PostgreSQL database). All integrate seamlessly with Edge Functions and serverless workloads.

## Vercel KV

### Installation and Setup

```bash
npm install @vercel/kv
```

```bash
# Create KV store via CLI
vercel storage create kv my-kv-store

# Link to project
vercel link
vercel env pull
```

### Basic Operations

```typescript
// app/api/kv/route.ts
import { kv } from '@vercel/kv'

export async function GET() {
  // Get value
  const value = await kv.get('my-key')

  return Response.json({ value })
}

export async function POST(request: Request) {
  const { key, value, ttl } = await request.json()

  // Set value with optional TTL (seconds)
  if (ttl) {
    await kv.set(key, value, { ex: ttl })
  } else {
    await kv.set(key, value)
  }

  return Response.json({ success: true })
}
```

### Common Patterns

```typescript
// Session storage
async function getSession(sessionId: string) {
  return kv.get(`session:${sessionId}`)
}

async function setSession(sessionId: string, data: object) {
  await kv.set(`session:${sessionId}`, data, { ex: 3600 }) // 1 hour TTL
}

// Rate limiting
async function checkRateLimit(ip: string): Promise<boolean> {
  const key = `ratelimit:${ip}`
  const current = await kv.incr(key)

  if (current === 1) {
    await kv.expire(key, 60) // Reset after 60 seconds
  }

  return current <= 100 // 100 requests per minute
}

// Caching
async function getCachedData(key: string) {
  const cached = await kv.get(key)
  if (cached) return cached

  const fresh = await fetchFreshData()
  await kv.set(key, fresh, { ex: 300 }) // 5 minute cache
  return fresh
}
```

### Hash Operations

```typescript
// Store user data as hash
await kv.hset('user:123', {
  name: 'John Doe',
  email: 'john@example.com',
  plan: 'pro'
})

// Get single field
const name = await kv.hget('user:123', 'name')

// Get all fields
const user = await kv.hgetall('user:123')

// Increment numeric field
await kv.hincrby('user:123', 'loginCount', 1)
```

### List Operations

```typescript
// Add to queue
await kv.lpush('job-queue', JSON.stringify({ task: 'send-email', to: 'user@example.com' }))

// Process from queue
const job = await kv.rpop('job-queue')

// Get recent items
const recent = await kv.lrange('recent-views', 0, 9)
```

## Vercel Blob

### Installation and Setup

```bash
npm install @vercel/blob
```

```bash
# Create blob store via CLI
vercel storage create blob my-blob-store

# Link to project
vercel link
vercel env pull
```

### Upload Files

```typescript
// app/api/upload/route.ts
import { put } from '@vercel/blob'
import { NextRequest } from 'next/server'

export async function POST(request: NextRequest) {
  const formData = await request.formData()
  const file = formData.get('file') as File

  if (!file) {
    return Response.json({ error: 'No file provided' }, { status: 400 })
  }

  const blob = await put(file.name, file, {
    access: 'public',
    addRandomSuffix: true
  })

  return Response.json(blob)
}
```

### Client-Side Upload

```typescript
// app/api/upload/route.ts
import { handleUpload, type HandleUploadBody } from '@vercel/blob/client'
import { NextRequest } from 'next/server'

export async function POST(request: NextRequest) {
  const body = await request.json() as HandleUploadBody

  const jsonResponse = await handleUpload({
    body,
    request,
    onBeforeGenerateToken: async (pathname) => {
      // Validate upload before generating token
      return {
        allowedContentTypes: ['image/jpeg', 'image/png', 'image/webp'],
        maximumSizeInBytes: 5 * 1024 * 1024 // 5MB
      }
    },
    onUploadCompleted: async ({ blob }) => {
      // Update database with blob URL
      console.log('Upload completed:', blob.url)
    }
  })

  return Response.json(jsonResponse)
}
```

```typescript
// components/Uploader.tsx
'use client'

import { upload } from '@vercel/blob/client'
import { useState } from 'react'

export function Uploader() {
  const [uploading, setUploading] = useState(false)
  const [url, setUrl] = useState<string>()

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    setUploading(true)
    const blob = await upload(file.name, file, {
      access: 'public',
      handleUploadUrl: '/api/upload'
    })
    setUrl(blob.url)
    setUploading(false)
  }

  return (
    <div>
      <input type="file" onChange={handleUpload} disabled={uploading} />
      {url && <img src={url} alt="Uploaded" />}
    </div>
  )
}
```

### List and Delete

```typescript
// app/api/files/route.ts
import { list, del } from '@vercel/blob'

export async function GET() {
  const { blobs } = await list()
  return Response.json(blobs)
}

export async function DELETE(request: Request) {
  const { url } = await request.json()
  await del(url)
  return Response.json({ deleted: true })
}
```

## Vercel Postgres

### Installation and Setup

```bash
npm install @vercel/postgres
```

```bash
# Create postgres database via CLI
vercel storage create postgres my-database

# Link to project
vercel link
vercel env pull
```

### Basic Queries

```typescript
// app/api/users/route.ts
import { sql } from '@vercel/postgres'

export async function GET() {
  const { rows } = await sql`SELECT * FROM users LIMIT 10`
  return Response.json(rows)
}

export async function POST(request: Request) {
  const { name, email } = await request.json()

  const { rows } = await sql`
    INSERT INTO users (name, email)
    VALUES (${name}, ${email})
    RETURNING *
  `

  return Response.json(rows[0])
}
```

### Connection Pooling

```typescript
// lib/db.ts
import { createPool } from '@vercel/postgres'

const pool = createPool({
  connectionString: process.env.POSTGRES_URL
})

export async function query(text: string, params?: any[]) {
  const client = await pool.connect()
  try {
    return await client.query(text, params)
  } finally {
    client.release()
  }
}
```

### Transactions

```typescript
// app/api/transfer/route.ts
import { sql } from '@vercel/postgres'

export async function POST(request: Request) {
  const { from, to, amount } = await request.json()

  await sql`BEGIN`
  try {
    await sql`
      UPDATE accounts SET balance = balance - ${amount}
      WHERE id = ${from} AND balance >= ${amount}
    `
    await sql`
      UPDATE accounts SET balance = balance + ${amount}
      WHERE id = ${to}
    `
    await sql`COMMIT`
    return Response.json({ success: true })
  } catch (error) {
    await sql`ROLLBACK`
    return Response.json({ error: 'Transfer failed' }, { status: 500 })
  }
}
```

### Schema Migrations

```typescript
// scripts/migrate.ts
import { sql } from '@vercel/postgres'

async function migrate() {
  await sql`
    CREATE TABLE IF NOT EXISTS users (
      id SERIAL PRIMARY KEY,
      name VARCHAR(255) NOT NULL,
      email VARCHAR(255) UNIQUE NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )
  `

  await sql`
    CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)
  `

  console.log('Migration complete')
}

migrate()
```

## Edge-Compatible Usage

### KV at Edge

```typescript
// app/api/edge-kv/route.ts
import { kv } from '@vercel/kv'

export const runtime = 'edge'

export async function GET(request: Request) {
  const url = new URL(request.url)
  const key = url.searchParams.get('key')

  if (!key) {
    return Response.json({ error: 'Key required' }, { status: 400 })
  }

  const value = await kv.get(key)
  return Response.json({ value })
}
```

### Postgres at Edge

```typescript
// app/api/edge-db/route.ts
import { sql } from '@vercel/postgres'

export const runtime = 'edge'

export async function GET() {
  // Uses Neon serverless driver for edge compatibility
  const { rows } = await sql`SELECT NOW()`
  return Response.json({ time: rows[0].now })
}
```

## Context7 Integration

For latest storage documentation, use:
- Library: `/vercel/vercel`
- Topics: kv, blob, postgres, storage
- Token allocation: 8000-12000 for comprehensive coverage
