---
name: moai-platform-supabase
description: Supabase specialist covering PostgreSQL 16, pgvector, RLS, real-time subscriptions, and Edge Functions. Use when building full-stack apps with Supabase backend.
version: 1.0.0
category: platform
tags: [supabase, postgresql, pgvector, realtime, rls, edge-functions]
context7-libraries: [/supabase/supabase]
related-skills: [moai-platform-neon, moai-lang-typescript]
updated: 2025-12-07
status: active
allowed-tools: Read, Write, Bash, Grep, Glob
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

```python
# Get latest Supabase documentation
docs = await mcp__context7__get_library_docs(
    context7CompatibleLibraryID="/supabase/supabase",
    topic="postgresql-16 pgvector rls edge-functions realtime storage auth",
    tokens=8000
)
```

---

## Implementation Guide

### PostgreSQL 16 + pgvector Setup

Enable Extensions and Create Embeddings Table:
```sql
-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector;

-- Create embeddings table for semantic search
CREATE TABLE documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  content TEXT NOT NULL,
  embedding vector(1536),  -- OpenAI ada-002 dimensions
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create HNSW index for fast similarity search (recommended)
CREATE INDEX idx_documents_embedding ON documents
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Alternative: IVFFlat index for large datasets (millions of rows)
-- CREATE INDEX idx_documents_ivf ON documents
-- USING ivfflat (embedding vector_cosine_ops)
-- WITH (lists = 100);
```

Semantic Search Function:
```sql
CREATE OR REPLACE FUNCTION search_documents(
  query_embedding vector(1536),
  match_threshold FLOAT DEFAULT 0.8,
  match_count INT DEFAULT 10
) RETURNS TABLE (id UUID, content TEXT, similarity FLOAT)
LANGUAGE plpgsql AS $$
BEGIN
  RETURN QUERY SELECT d.id, d.content,
    1 - (d.embedding <=> query_embedding) AS similarity
  FROM documents d
  WHERE 1 - (d.embedding <=> query_embedding) > match_threshold
  ORDER BY d.embedding <=> query_embedding
  LIMIT match_count;
END; $$;
```

Hybrid Search (Vector + Full-Text):
```sql
CREATE OR REPLACE FUNCTION hybrid_search(
  query_text TEXT,
  query_embedding vector(1536),
  match_count INT DEFAULT 10,
  full_text_weight FLOAT DEFAULT 0.3,
  semantic_weight FLOAT DEFAULT 0.7
) RETURNS TABLE (id UUID, content TEXT, score FLOAT) AS $$
BEGIN
  RETURN QUERY
  WITH semantic AS (
    SELECT e.id, e.content, 1 - (e.embedding <=> query_embedding) AS similarity
    FROM documents e ORDER BY e.embedding <=> query_embedding LIMIT match_count * 2
  ),
  full_text AS (
    SELECT e.id, e.content,
      ts_rank(to_tsvector('english', e.content), plainto_tsquery('english', query_text)) AS rank
    FROM documents e
    WHERE to_tsvector('english', e.content) @@ plainto_tsquery('english', query_text)
    LIMIT match_count * 2
  )
  SELECT COALESCE(s.id, f.id), COALESCE(s.content, f.content),
    (COALESCE(s.similarity, 0) * semantic_weight + COALESCE(f.rank, 0) * full_text_weight)
  FROM semantic s FULL OUTER JOIN full_text f ON s.id = f.id
  ORDER BY 3 DESC LIMIT match_count;
END; $$ LANGUAGE plpgsql;
```

### Row-Level Security (RLS) Patterns

Basic Tenant Isolation:
```sql
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

-- Policy based on JWT claims
CREATE POLICY "tenant_isolation" ON projects FOR ALL
  USING (tenant_id = (auth.jwt() ->> 'tenant_id')::UUID);

-- Owner-based access
CREATE POLICY "owner_access" ON projects FOR ALL
  USING (owner_id = auth.uid());
```

Multi-Tenant with Hierarchical Access:
```sql
-- Organization-based access
CREATE POLICY "org_member_select" ON organizations FOR SELECT
  USING (id IN (SELECT org_id FROM org_members WHERE user_id = auth.uid()));

-- Role-based modification
CREATE POLICY "org_admin_modify" ON organizations FOR UPDATE
  USING (id IN (
    SELECT org_id FROM org_members
    WHERE user_id = auth.uid() AND role IN ('owner', 'admin')
  ));

-- Cascading project access through organization membership
CREATE POLICY "project_access" ON projects FOR ALL
  USING (org_id IN (SELECT org_id FROM org_members WHERE user_id = auth.uid()));

-- Service role bypass for server-side operations
CREATE POLICY "service_bypass" ON organizations FOR ALL TO service_role USING (true);
```

### Real-time Subscriptions

Table Changes Subscription:
```typescript
import { createClient } from '@supabase/supabase-js'

const supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY)

// Subscribe to all changes on a table
const channel = supabase.channel('db-changes')
  .on('postgres_changes',
    { event: '*', schema: 'public', table: 'messages' },
    (payload) => console.log('Change:', payload)
  )
  .subscribe()

// Filter by specific conditions
supabase.channel('project-updates')
  .on('postgres_changes',
    { event: 'UPDATE', schema: 'public', table: 'projects', filter: `id=eq.${projectId}` },
    (payload) => handleProjectUpdate(payload.new)
  )
  .subscribe()
```

Presence Tracking:
```typescript
interface PresenceState {
  user_id: string
  online_at: string
  typing?: boolean
  cursor?: { x: number; y: number }
}

const channel = supabase.channel('room:collaborative-doc', {
  config: { presence: { key: userId } }
})

channel
  .on('presence', { event: 'sync' }, () => {
    const state = channel.presenceState<PresenceState>()
    console.log('Online users:', Object.keys(state))
  })
  .on('presence', { event: 'join' }, ({ key, newPresences }) => {
    console.log('User joined:', key, newPresences)
  })
  .on('presence', { event: 'leave' }, ({ key, leftPresences }) => {
    console.log('User left:', key, leftPresences)
  })
  .subscribe(async (status) => {
    if (status === 'SUBSCRIBED') {
      await channel.track({ user_id: userId, online_at: new Date().toISOString() })
    }
  })

// Update presence state
await channel.track({ typing: true })
await channel.track({ cursor: { x: 100, y: 200 } })
```

### Edge Functions

Basic Edge Function with Auth:
```typescript
// supabase/functions/api/index.ts
import { serve } from 'https://deno.land/std@0.168.0/http/server.ts'
import { createClient } from 'https://esm.sh/@supabase/supabase-js@2'

const corsHeaders = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type'
}

serve(async (req) => {
  if (req.method === 'OPTIONS') {
    return new Response('ok', { headers: corsHeaders })
  }

  const supabase = createClient(
    Deno.env.get('SUPABASE_URL')!,
    Deno.env.get('SUPABASE_SERVICE_ROLE_KEY')!
  )

  // Verify JWT token
  const authHeader = req.headers.get('authorization')
  if (!authHeader) {
    return new Response(JSON.stringify({ error: 'Unauthorized' }),
      { status: 401, headers: { ...corsHeaders, 'Content-Type': 'application/json' } })
  }

  const { data: { user }, error } = await supabase.auth.getUser(
    authHeader.replace('Bearer ', '')
  )

  if (error || !user) {
    return new Response(JSON.stringify({ error: 'Invalid token' }),
      { status: 401, headers: { ...corsHeaders, 'Content-Type': 'application/json' } })
  }

  // Process request
  const body = await req.json()
  return new Response(JSON.stringify({ success: true, user_id: user.id }),
    { headers: { ...corsHeaders, 'Content-Type': 'application/json' } })
})
```

Rate Limiting in Edge Functions:
```typescript
async function checkRateLimit(
  supabase: SupabaseClient, identifier: string, limit: number, windowSeconds: number
): Promise<boolean> {
  const windowStart = new Date(Date.now() - windowSeconds * 1000).toISOString()
  const { count } = await supabase
    .from('rate_limits')
    .select('*', { count: 'exact', head: true })
    .eq('identifier', identifier)
    .gte('created_at', windowStart)

  if (count && count >= limit) return false
  await supabase.from('rate_limits').insert({ identifier })
  return true
}
```

### Storage with Image Transformations

```typescript
const supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY)

async function uploadImage(file: File, userId: string) {
  const fileName = `${userId}/${Date.now()}-${file.name}`

  const { data, error } = await supabase.storage
    .from('images')
    .upload(fileName, file, { cacheControl: '3600', upsert: false })

  if (error) throw error

  // Get transformed URLs
  const { data: { publicUrl } } = supabase.storage
    .from('images')
    .getPublicUrl(fileName, {
      transform: { width: 800, height: 600, resize: 'contain' }
    })

  const { data: { publicUrl: thumbnailUrl } } = supabase.storage
    .from('images')
    .getPublicUrl(fileName, {
      transform: { width: 200, height: 200, resize: 'cover' }
    })

  return { originalPath: data.path, publicUrl, thumbnailUrl }
}
```

---

## Advanced Patterns

### Multi-Tenant SaaS Architecture

Complete Schema Setup:
```sql
-- Organizations (tenants)
CREATE TABLE organizations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name TEXT NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  plan TEXT DEFAULT 'free' CHECK (plan IN ('free', 'pro', 'enterprise')),
  settings JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Organization members with roles
CREATE TABLE organization_members (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
  user_id UUID NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'viewer')),
  joined_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(organization_id, user_id)
);

-- Projects within organizations
CREATE TABLE projects (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  owner_id UUID NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Enable RLS on all tables
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organization_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

-- Create comprehensive RLS policies
CREATE POLICY "org_member_select" ON organizations FOR SELECT
  USING (id IN (SELECT organization_id FROM organization_members WHERE user_id = auth.uid()));

CREATE POLICY "org_admin_update" ON organizations FOR UPDATE
  USING (id IN (SELECT organization_id FROM organization_members
    WHERE user_id = auth.uid() AND role IN ('owner', 'admin')));

CREATE POLICY "project_member_access" ON projects FOR ALL
  USING (organization_id IN (SELECT organization_id FROM organization_members WHERE user_id = auth.uid()));
```

### TypeScript Client Patterns

Server-Side Client (Next.js App Router):
```typescript
import { createServerClient } from '@supabase/ssr'
import { cookies } from 'next/headers'
import { Database } from './database.types'

export function createServerSupabase() {
  const cookieStore = cookies()
  return createServerClient<Database>(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
    {
      cookies: {
        get(name: string) { return cookieStore.get(name)?.value },
        set(name, value, options) { cookieStore.set({ name, value, ...options }) },
        remove(name, options) { cookieStore.set({ name, value: '', ...options }) }
      }
    }
  )
}
```

Service Layer Pattern:
```typescript
import { supabase } from './supabase/client'

export class DocumentService {
  async create(projectId: string, title: string, content: string) {
    const { data: { user } } = await supabase.auth.getUser()
    const { data, error } = await supabase
      .from('documents')
      .insert({ project_id: projectId, title, content, created_by: user!.id })
      .select().single()

    if (error) throw error

    // Generate embedding async
    await supabase.functions.invoke('generate-embedding',
      { body: { documentId: data.id, content } })

    return data
  }

  async semanticSearch(projectId: string, query: string) {
    const { data: embeddingData } = await supabase.functions.invoke(
      'get-embedding', { body: { text: query } })

    const { data, error } = await supabase.rpc('search_documents', {
      p_project_id: projectId,
      p_query_embedding: embeddingData.embedding,
      p_match_threshold: 0.7,
      p_match_count: 10
    })

    if (error) throw error
    return data
  }

  subscribeToChanges(projectId: string, callback: (payload: any) => void) {
    return supabase.channel(`documents:${projectId}`)
      .on('postgres_changes',
        { event: '*', schema: 'public', table: 'documents', filter: `project_id=eq.${projectId}` },
        callback)
      .subscribe()
  }
}
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
- moai-quality-security - Security and RLS best practices
- moai-workflow-tdd - Test-driven development with Supabase

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v2.0
Last Updated: 2025-12-07
Coverage: PostgreSQL 16, pgvector, RLS, Real-time, Edge Functions, Storage
