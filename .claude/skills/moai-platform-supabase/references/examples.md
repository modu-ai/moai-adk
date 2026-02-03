---
name: supabase-examples
description: Full-stack templates and working examples for Supabase applications
parent-skill: moai-platform-supabase
version: 1.0.0
updated: 2026-01-06
---

# Supabase Full-Stack Examples

## Multi-Tenant SaaS Application

### Database Schema

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
  description TEXT,
  owner_id UUID NOT NULL,
  status TEXT DEFAULT 'active',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organization_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

-- Policies
CREATE POLICY "org_member_select" ON organizations FOR SELECT
  USING (id IN (SELECT organization_id FROM organization_members WHERE user_id = auth.uid()));

CREATE POLICY "org_admin_update" ON organizations FOR UPDATE
  USING (id IN (SELECT organization_id FROM organization_members
    WHERE user_id = auth.uid() AND role IN ('owner', 'admin')));

CREATE POLICY "member_view" ON organization_members FOR SELECT
  USING (organization_id IN (SELECT organization_id FROM organization_members WHERE user_id = auth.uid()));

CREATE POLICY "project_access" ON projects FOR ALL
  USING (organization_id IN (SELECT organization_id FROM organization_members WHERE user_id = auth.uid()));

-- Indexes
CREATE INDEX idx_org_members_user ON organization_members(user_id);
CREATE INDEX idx_org_members_org ON organization_members(organization_id);
CREATE INDEX idx_projects_org ON projects(organization_id);
```

### TypeScript Service Layer

```typescript
// services/organization-service.ts
import { supabase } from '@/lib/supabase/client'
import type { Database } from '@/types/database'

type Organization = Database['public']['Tables']['organizations']['Row']
type OrganizationMember = Database['public']['Tables']['organization_members']['Row']

export class OrganizationService {
  async create(name: string, slug: string): Promise<Organization> {
    const { data: { user } } = await supabase.auth.getUser()
    if (!user) throw new Error('Not authenticated')

    // Create organization
    const { data: org, error: orgError } = await supabase
      .from('organizations')
      .insert({ name, slug })
      .select()
      .single()

    if (orgError) throw orgError

    // Add creator as owner
    const { error: memberError } = await supabase
      .from('organization_members')
      .insert({
        organization_id: org.id,
        user_id: user.id,
        role: 'owner'
      })

    if (memberError) {
      // Rollback org creation
      await supabase.from('organizations').delete().eq('id', org.id)
      throw memberError
    }

    return org
  }

  async getMyOrganizations(): Promise<Organization[]> {
    const { data, error } = await supabase
      .from('organizations')
      .select('*, organization_members!inner(role)')
      .order('name')

    if (error) throw error
    return data
  }

  async getMembers(orgId: string): Promise<OrganizationMember[]> {
    const { data, error } = await supabase
      .from('organization_members')
      .select('*, user:profiles(*)')
      .eq('organization_id', orgId)
      .order('joined_at')

    if (error) throw error
    return data
  }

  async inviteMember(orgId: string, email: string, role: string): Promise<void> {
    const { error } = await supabase.functions.invoke('invite-member', {
      body: { organizationId: orgId, email, role }
    })

    if (error) throw error
  }
}

export const organizationService = new OrganizationService()
```

## AI Document Search Application

### Database Schema

```sql
-- Enable extensions
CREATE EXTENSION IF NOT EXISTS vector;

-- Documents with embeddings
CREATE TABLE documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  embedding vector(1536),
  metadata JSONB DEFAULT '{}',
  created_by UUID NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- HNSW index for fast similarity search
CREATE INDEX idx_documents_embedding ON documents
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Full-text search index
CREATE INDEX idx_documents_content_fts ON documents
USING gin(to_tsvector('english', content));

-- Enable RLS
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

CREATE POLICY "document_access" ON documents FOR ALL
  USING (project_id IN (
    SELECT p.id FROM projects p
    JOIN organization_members om ON p.organization_id = om.organization_id
    WHERE om.user_id = auth.uid()
  ));

-- Semantic search function
CREATE OR REPLACE FUNCTION search_documents(
  p_project_id UUID,
  p_query_embedding vector(1536),
  p_match_threshold FLOAT DEFAULT 0.7,
  p_match_count INT DEFAULT 10
) RETURNS TABLE (
  id UUID,
  title TEXT,
  content TEXT,
  similarity FLOAT
) LANGUAGE plpgsql AS $$
BEGIN
  RETURN QUERY
  SELECT d.id, d.title, d.content,
    1 - (d.embedding <=> p_query_embedding) AS similarity
  FROM documents d
  WHERE d.project_id = p_project_id
    AND 1 - (d.embedding <=> p_query_embedding) > p_match_threshold
  ORDER BY d.embedding <=> p_query_embedding
  LIMIT p_match_count;
END; $$;
```

### Edge Function for Embeddings

```typescript
// supabase/functions/generate-embedding/index.ts
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

  try {
    const supabase = createClient(
      Deno.env.get('SUPABASE_URL')!,
      Deno.env.get('SUPABASE_SERVICE_ROLE_KEY')!
    )

    const { documentId, content } = await req.json()

    // Generate embedding using OpenAI
    const embeddingResponse = await fetch('https://api.openai.com/v1/embeddings', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${Deno.env.get('OPENAI_API_KEY')}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        model: 'text-embedding-ada-002',
        input: content.slice(0, 8000)
      })
    })

    const embeddingData = await embeddingResponse.json()
    const embedding = embeddingData.data[0].embedding

    // Update document with embedding
    const { error } = await supabase
      .from('documents')
      .update({ embedding })
      .eq('id', documentId)

    if (error) throw error

    return new Response(
      JSON.stringify({ success: true }),
      { headers: { ...corsHeaders, 'Content-Type': 'application/json' } }
    )
  } catch (error) {
    return new Response(
      JSON.stringify({ error: error.message }),
      { status: 500, headers: { ...corsHeaders, 'Content-Type': 'application/json' } }
    )
  }
})
```

### React Search Component

```typescript
// components/DocumentSearch.tsx
'use client'

import { useState, useEffect } from 'react'
import { useDebounce } from '@/hooks/useDebounce'
import { documentService } from '@/services/document-service'

interface SearchResult {
  id: string
  title: string
  content: string
  similarity: number
}

export function DocumentSearch({ projectId }: { projectId: string }) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<SearchResult[]>([])
  const [loading, setLoading] = useState(false)

  const debouncedQuery = useDebounce(query, 300)

  useEffect(() => {
    if (debouncedQuery.length < 3) {
      setResults([])
      return
    }

    async function search() {
      setLoading(true)
      try {
        const data = await documentService.semanticSearch(projectId, debouncedQuery)
        setResults(data)
      } catch (error) {
        console.error('Search failed:', error)
      } finally {
        setLoading(false)
      }
    }

    search()
  }, [debouncedQuery, projectId])

  return (
    <div className="space-y-4">
      <input
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search documents..."
        className="w-full px-4 py-2 border rounded-lg"
      />

      {loading && <div>Searching...</div>}

      <div className="space-y-2">
        {results.map((result) => (
          <div key={result.id} className="p-4 border rounded-lg">
            <div className="flex justify-between">
              <h3 className="font-medium">{result.title}</h3>
              <span className="text-sm text-gray-500">
                {(result.similarity * 100).toFixed(1)}% match
              </span>
            </div>
            <p className="mt-2 text-gray-600 line-clamp-2">{result.content}</p>
          </div>
        ))}
      </div>
    </div>
  )
}
```

## Real-Time Collaboration

### Collaborative Editor with Presence

```typescript
// components/CollaborativeEditor.tsx
'use client'

import { useEffect, useState, useCallback } from 'react'
import { supabase } from '@/lib/supabase/client'

interface User {
  id: string
  name: string
  color: string
}

interface PresenceState {
  user: User
  cursor: { x: number; y: number } | null
  selection: { start: number; end: number } | null
}

export function CollaborativeEditor({
  documentId,
  currentUser
}: {
  documentId: string
  currentUser: User
}) {
  const [content, setContent] = useState('')
  const [otherUsers, setOtherUsers] = useState<PresenceState[]>([])
  const [channel, setChannel] = useState<ReturnType<typeof supabase.channel> | null>(null)

  useEffect(() => {
    const ch = supabase.channel(`doc:${documentId}`, {
      config: { presence: { key: currentUser.id } }
    })

    ch.on('presence', { event: 'sync' }, () => {
      const state = ch.presenceState<PresenceState>()
      const users = Object.values(state)
        .flat()
        .filter((p) => p.user.id !== currentUser.id)
      setOtherUsers(users)
    })

    ch.on('broadcast', { event: 'content-update' }, ({ payload }) => {
      if (payload.userId !== currentUser.id) {
        setContent(payload.content)
      }
    })

    ch.subscribe(async (status) => {
      if (status === 'SUBSCRIBED') {
        await ch.track({
          user: currentUser,
          cursor: null,
          selection: null
        })
      }
    })

    setChannel(ch)

    return () => {
      supabase.removeChannel(ch)
    }
  }, [documentId, currentUser])

  const handleContentChange = useCallback(
    async (newContent: string) => {
      setContent(newContent)
      if (channel) {
        await channel.send({
          type: 'broadcast',
          event: 'content-update',
          payload: { userId: currentUser.id, content: newContent }
        })
      }
    },
    [channel, currentUser.id]
  )

  return (
    <div className="relative">
      <textarea
        value={content}
        onChange={(e) => handleContentChange(e.target.value)}
        className="w-full h-96 p-4 border rounded-lg"
      />
      <div className="mt-4 flex gap-2">
        <span className="text-sm text-gray-500">Active:</span>
        {otherUsers.map((presence) => (
          <span
            key={presence.user.id}
            className="px-2 py-1 text-xs rounded-full"
            style={{ backgroundColor: presence.user.color + '20', color: presence.user.color }}
          >
            {presence.user.name}
          </span>
        ))}
      </div>
    </div>
  )
}
```

## Project Structure Template

```
my-supabase-app/
├── supabase/
│   ├── functions/
│   │   ├── generate-embedding/
│   │   │   └── index.ts
│   │   └── invite-member/
│   │       └── index.ts
│   ├── migrations/
│   │   ├── 20240101000000_initial_schema.sql
│   │   └── 20240101000001_add_embeddings.sql
│   └── config.toml
├── src/
│   ├── lib/
│   │   └── supabase/
│   │       ├── client.ts
│   │       └── server.ts
│   ├── services/
│   │   ├── organization-service.ts
│   │   ├── project-service.ts
│   │   └── document-service.ts
│   ├── types/
│   │   └── database.ts
│   └── components/
│       ├── DocumentSearch.tsx
│       ├── CollaborativeEditor.tsx
│       └── FileUploader.tsx
├── .env.local
└── package.json
```

## Environment Configuration

```bash
# .env.local
NEXT_PUBLIC_SUPABASE_URL=https://xxx.supabase.co
NEXT_PUBLIC_SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Server-side only
SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Edge Functions secrets (set via CLI)
# supabase secrets set OPENAI_API_KEY=sk-...
```
