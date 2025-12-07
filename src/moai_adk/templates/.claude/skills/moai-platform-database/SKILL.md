---
name: moai-platform-database
description: Database platform specialist covering Supabase, Neon, Convex, and Firebase Firestore with PostgreSQL 16, real-time features, serverless patterns, and AI extensions. Use when implementing database architecture, real-time sync, Row-Level Security, or database branching.
version: 1.0.0
category: platform
updated: 2025-12-07
status: active
tags:
  - platform
  - database
  - supabase
  - neon
  - convex
  - firestore
  - postgresql
  - realtime
allowed-tools: Read, Write, Bash, Grep, Glob
---

# moai-platform-database: Database Platform Specialist

## Quick Reference (30 seconds)

Modern Database Platform Expertise: Specialized knowledge for Supabase, Neon, Convex, and Firebase Firestore covering PostgreSQL 16, real-time subscriptions, serverless patterns, and AI-powered database features.

### Provider Selection Matrix

Supabase: PostgreSQL 16 with pgvector, RLS, Edge Functions, real-time subscriptions
Neon: Serverless PostgreSQL with auto-scaling, branching, 30-day PITR
Convex: Real-time reactive backend with TypeScript-first design, optimistic updates
Firestore: Real-time sync with offline caching, mobile-first SDKs, Cloud Functions

### Quick Provider Selection

- Need PostgreSQL with AI/vector search? Supabase (pgvector)
- Need serverless scaling with branching? Neon
- Need real-time collaborative features? Convex
- Need mobile-first with offline sync? Firestore

### Context7 Library Mappings

Supabase: /supabase/supabase
Neon: /neondatabase/neon
Convex: /get-convex/convex
Firestore: /firebase/firebase-docs

---

## Implementation Guide

### Supabase: PostgreSQL 16 Powerhouse

Core Features:
- PostgreSQL 16 with pgvector for AI embeddings and vector search
- Row-Level Security (RLS) for multi-tenant isolation
- Real-time subscriptions via Postgres Changes
- Edge Functions with Deno runtime
- Storage with image transformations
- Auto-generated REST and GraphQL APIs

PostgreSQL 16 + pgvector Setup:
```sql
-- Enable pgvector extension for AI embeddings
CREATE EXTENSION IF NOT EXISTS vector;

-- Create embeddings table for semantic search
CREATE TABLE documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  content TEXT NOT NULL,
  embedding vector(1536),
  metadata JSONB DEFAULT '{}'
);

-- Create HNSW index for fast similarity search
CREATE INDEX ON documents USING hnsw (embedding vector_cosine_ops);

-- Semantic search function
CREATE OR REPLACE FUNCTION search_documents(
  query_embedding vector(1536),
  match_threshold FLOAT DEFAULT 0.8
) RETURNS TABLE (id UUID, content TEXT, similarity FLOAT)
LANGUAGE plpgsql AS $$
BEGIN
  RETURN QUERY SELECT d.id, d.content,
    1 - (d.embedding <=> query_embedding) AS similarity
  FROM documents d
  WHERE 1 - (d.embedding <=> query_embedding) > match_threshold
  ORDER BY d.embedding <=> query_embedding LIMIT 10;
END; $$;
```

Row-Level Security for Multi-Tenant:
```sql
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

CREATE POLICY "tenant_isolation" ON projects FOR ALL
  USING (tenant_id = (auth.jwt() ->> 'tenant_id')::UUID);

CREATE POLICY "role_based_access" ON projects FOR SELECT
  USING (owner_id = auth.uid() OR EXISTS (
    SELECT 1 FROM project_members WHERE project_id = projects.id AND user_id = auth.uid()
  ));
```

Real-time Subscriptions:
```typescript
const supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY)

// Subscribe to table changes
supabase.channel('changes').on('postgres_changes',
  { event: '*', schema: 'public', table: 'messages' },
  (payload) => console.log('Change:', payload)
).subscribe()

// Presence tracking
const channel = supabase.channel('room:1')
channel.on('presence', { event: 'sync' }, () => {
  console.log('Online:', Object.keys(channel.presenceState()))
}).subscribe(async (status) => {
  if (status === 'SUBSCRIBED') await channel.track({ user_id: 'user-1' })
})
```

### Neon: Serverless PostgreSQL

Core Features:
- Auto-scaling serverless compute (scale to zero)
- Instant database branching for dev/preview
- 30-day Point-in-Time Recovery
- Connection pooling for edge compatibility
- PostgreSQL 16 compatibility

Database Branching Workflow:
```typescript
import { neon } from '@neondatabase/serverless'

class NeonBranchManager {
  async createBranch(name: string, parentId: string = 'main') {
    const response = await fetch(
      `https://console.neon.tech/api/v2/projects/${this.projectId}/branches`,
      {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${this.apiKey}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({ branch: { name, parent_id: parentId } })
      }
    )
    return response.json()
  }

  async createPreviewBranch(prNumber: number) {
    return this.createBranch(`pr-${prNumber}`, 'main')
  }
}

// Serverless query with connection pooling
const sql = neon(process.env.DATABASE_URL!)
const users = await sql`SELECT * FROM users WHERE active = true`
```

Point-in-Time Recovery:
```typescript
async function restoreToPoint(branchId: string, timestamp: Date) {
  return await fetch(`https://console.neon.tech/api/v2/projects/${PROJECT_ID}/branches`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${API_KEY}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({
      branch: { name: `restore-${timestamp.toISOString()}`, parent_id: branchId, parent_timestamp: timestamp.toISOString() }
    })
  }).then(r => r.json())
}
```

### Convex: Real-time Reactive Backend

Core Features:
- Real-time reactive queries (automatic UI updates)
- Optimistic updates with conflict resolution
- Database branching for instant environments
- TypeScript-first with full type safety
- Built-in caching and automatic optimization

Schema Definition:
```typescript
import { defineSchema, defineTable } from 'convex/server'
import { v } from 'convex/values'

export default defineSchema({
  documents: defineTable({
    title: v.string(),
    content: v.string(),
    ownerId: v.string(),
    isPublic: v.boolean(),
    lastModified: v.number()
  }).index('by_owner', ['ownerId']).searchIndex('search_content', { searchField: 'content' }),

  collaborators: defineTable({
    documentId: v.id('documents'),
    userId: v.string(),
    permission: v.union(v.literal('read'), v.literal('write'))
  }).index('by_document', ['documentId']).index('by_user', ['userId'])
})
```

Reactive Queries and Mutations:
```typescript
import { mutation, query } from './_generated/server'
import { v } from 'convex/values'

export const list = query({
  args: { ownerId: v.string() },
  handler: async (ctx, args) => await ctx.db.query('documents')
    .withIndex('by_owner', (q) => q.eq('ownerId', args.ownerId)).order('desc').collect()
})

export const update = mutation({
  args: { id: v.id('documents'), content: v.string() },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new Error('Unauthorized')
    await ctx.db.patch(args.id, { content: args.content, lastModified: Date.now() })
  }
})
```

### Firebase Firestore: Mobile-First Real-time

Core Features:
- Real-time synchronization across clients
- Offline caching with automatic sync
- Security Rules for field-level access
- Composite indexes for complex queries
- Cloud Functions triggers

Security Rules:
```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    match /users/{userId} {
      allow read, write: if request.auth.uid == userId;
    }
    match /documents/{docId} {
      allow read: if resource.data.isPublic == true
        || request.auth.uid == resource.data.ownerId
        || request.auth.uid in resource.data.collaborators;
      allow create: if request.auth != null && request.resource.data.ownerId == request.auth.uid;
      allow update: if request.auth.uid == resource.data.ownerId;
      allow delete: if request.auth.uid == resource.data.ownerId;
    }
  }
}
```

Real-time Listeners with Offline Support:
```typescript
import { getFirestore, collection, query, where, onSnapshot, enableIndexedDbPersistence } from 'firebase/firestore'

const db = getFirestore(app)
enableIndexedDbPersistence(db).catch(console.warn)

export function subscribeToDocuments(userId: string, callback: (docs: Doc[]) => void) {
  const q = query(collection(db, 'documents'), where('collaborators', 'array-contains', userId))
  return onSnapshot(q, { includeMetadataChanges: true }, (snapshot) => {
    callback(snapshot.docs.map((doc) => ({
      id: doc.id, ...doc.data(),
      _pending: doc.metadata.hasPendingWrites, _fromCache: doc.metadata.fromCache
    })))
  })
}
```

Cloud Functions Triggers:
```typescript
import { onDocumentCreated } from 'firebase-functions/v2/firestore'
import { getFirestore } from 'firebase-admin/firestore'

export const onDocumentCreate = onDocumentCreated('documents/{docId}', async (event) => {
  const data = event.data?.data()
  if (!data) return
  const db = getFirestore()
  await db.collection('activity').add({
    type: 'document_created', documentId: event.params.docId,
    userId: data.ownerId, timestamp: new Date()
  })
})
```

---

## Advanced Patterns

For detailed implementation patterns including:
- Multi-tenant SaaS with RLS
- Preview environments with Neon branching
- Real-time collaboration with Convex
- Mobile offline sync with Firestore

See [reference.md](reference.md) and [examples.md](examples.md)

---

## Provider Decision Matrix

### By Use Case

Multi-Tenant SaaS: Supabase (RLS provides automatic tenant isolation)
AI/ML Applications: Supabase (pgvector for embeddings and similarity search)
Preview Environments: Neon (instant database branching per PR)
Real-time Collaboration: Convex (reactive queries with optimistic updates)
Mobile Apps: Firestore (offline-first with automatic sync)
Edge Computing: Neon or Supabase Edge Functions

### By Technical Requirements

PostgreSQL Required: Supabase or Neon
NoSQL Flexible Schema: Firestore or Convex
TypeScript-First: Convex
Serverless Scale-to-Zero: Neon
Google Cloud Integration: Firestore

### Cost Comparison (2024)

Supabase Free: 500MB DB, 1GB storage, 50K MAU
Neon Free: 3GB storage, 100 compute hours/month
Convex Free: 500MB storage, 1M function calls
Firestore Free: 1GB storage, 50K reads/day

---

## Works Well With

- moai-platform-baas - Cross-provider integration patterns
- moai-domain-backend - Backend architecture with database integration
- moai-lang-typescript - TypeScript patterns for Convex and Supabase
- moai-quality-security - Database security and RLS best practices
- moai-context7-integration - Latest documentation access

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v2.0
Last Updated: 2025-12-07
Providers Covered: Supabase, Neon, Convex, Firebase Firestore
