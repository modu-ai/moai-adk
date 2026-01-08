---
name: "moai-platform-supabase"
description: "Supabase specialist covering PostgreSQL 16, pgvector, RLS, real-time subscriptions, and Edge Functions. Use when building full-stack apps with Supabase backend."
version: 2.0.0
category: "platform"
modularized: true
user-invocable: false
tags: ['supabase', 'postgresql', 'pgvector', 'realtime', 'rls', 'edge-functions']
context7-libraries: ['/supabase/supabase']
related-skills: ['moai-platform-neon', 'moai-lang-typescript']
updated: 2026-01-08
status: "active"
allowed-tools:
  - Read
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# moai-platform-supabase: Supabase Platform Specialist

## Quick Reference (30 seconds)

Supabase Full-Stack Platform: PostgreSQL 16 with pgvector for AI/vector search, Row-Level Security for multi-tenant apps, real-time subscriptions, Edge Functions with Deno runtime, and integrated Storage with transformations.

### Core Capabilities

PostgreSQL 16: Latest PostgreSQL with full SQL support, JSONB, and advanced features
pgvector Extension: AI embeddings storage with HNSW/IVFFlat indexes for similarity search
Row-Level Security: Automatic multi-tenant data isolation at database level
Real-time Subscriptions: Live data sync via Postgres Changes and Presence
Edge Functions: Serverless Deno functions at the edge
Storage: File storage with automatic image transformations
Auth: Built-in authentication with JWT integration

### When to Use Supabase

- Multi-tenant SaaS applications requiring data isolation
- AI/ML applications needing vector embeddings and similarity search
- Real-time collaborative features (presence, live updates)
- Full-stack applications needing auth, database, and storage
- Projects requiring PostgreSQL-specific features

### Context7 Documentation Access

For latest Supabase API documentation, use the Context7 MCP tools:

Step 1 - Resolve library ID:
Use mcp__context7__resolve-library-id with query "supabase" to get the Context7-compatible library ID

Step 2 - Fetch documentation:
Use mcp__context7__get-library-docs with the resolved library ID, specifying topic and token allocation

Example topics: "postgresql pgvector", "row-level-security policies", "realtime subscriptions presence", "edge-functions deno", "storage transformations", "auth jwt"

---

## Module Index

This skill uses progressive disclosure with specialized modules for detailed implementation patterns.

### Core Modules

**postgresql-pgvector** - PostgreSQL 16 with pgvector extension for AI embeddings and semantic search
- Vector storage with 1536-dimension OpenAI embeddings
- HNSW and IVFFlat index strategies
- Semantic search functions
- Hybrid search combining vector and full-text

**row-level-security** - RLS policies for multi-tenant data isolation
- Basic tenant isolation patterns
- Hierarchical organization access
- Role-based modification policies
- Service role bypass for server operations

**realtime-presence** - Real-time subscriptions and presence tracking
- Postgres Changes subscription patterns
- Filtered change listeners
- Presence state management
- Collaborative cursor and typing indicators

**edge-functions** - Serverless Deno functions at the edge
- Basic Edge Function with authentication
- CORS header configuration
- JWT token verification
- Rate limiting implementation

**storage-cdn** - File storage with image transformations
- File upload patterns
- Image transformation URLs
- Thumbnail generation
- Cache control configuration

**auth-integration** - Authentication patterns and JWT handling
- Server-side client creation
- Cookie-based session management
- Auth state synchronization
- Protected route patterns

**typescript-patterns** - TypeScript client patterns and service layers
- Server-side client for Next.js App Router
- Service layer abstraction pattern
- Subscription management
- Type-safe database operations

---

## Quick Start Patterns

### Database Setup

Enable pgvector extension and create embeddings table:

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  content TEXT NOT NULL,
  embedding vector(1536),
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_documents_embedding ON documents
USING hnsw (embedding vector_cosine_ops);
```

### Basic RLS Policy

```sql
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

CREATE POLICY "tenant_isolation" ON projects FOR ALL
  USING (tenant_id = (auth.jwt() ->> 'tenant_id')::UUID);
```

### Real-time Subscription

```typescript
const channel = supabase.channel('db-changes')
  .on('postgres_changes',
    { event: '*', schema: 'public', table: 'messages' },
    (payload) => console.log('Change:', payload)
  )
  .subscribe()
```

### Edge Function Template

```typescript
import { serve } from 'https://deno.land/std@0.168.0/http/server.ts'
import { createClient } from 'https://esm.sh/@supabase/supabase-js@2'

serve(async (req) => {
  const supabase = createClient(
    Deno.env.get('SUPABASE_URL')!,
    Deno.env.get('SUPABASE_SERVICE_ROLE_KEY')!
  )
  // Process request
  return new Response(JSON.stringify({ success: true }))
})
```

---

## Best Practices

Performance: Use HNSW indexes for vectors, Supavisor for connection pooling in serverless
Security: Always enable RLS, verify JWT tokens, use service_role only in Edge Functions
Migration: Use Supabase CLI (supabase migration new, supabase db push)

---

## Works Well With

- moai-platform-neon - Alternative PostgreSQL for specific use cases
- moai-lang-typescript - TypeScript patterns for Supabase client
- moai-domain-backend - Backend architecture integration
- moai-foundation-quality - Security and RLS best practices
- moai-workflow-testing - Test-driven development with Supabase

---

## Module References

For detailed implementation patterns, see the modules directory:

- modules/postgresql-pgvector.md - Complete vector search implementation
- modules/row-level-security.md - Multi-tenant RLS patterns
- modules/realtime-presence.md - Real-time collaboration features
- modules/edge-functions.md - Serverless function patterns
- modules/storage-cdn.md - File storage and transformations
- modules/auth-integration.md - Authentication patterns
- modules/typescript-patterns.md - TypeScript client architecture

For API reference summary, see reference.md
For full-stack templates, see examples.md

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v2.0
Last Updated: 2026-01-06
Version: 2.0.0 (Modularized)
Coverage: PostgreSQL 16, pgvector, RLS, Real-time, Edge Functions, Storage
