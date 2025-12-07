# Database Platform Production Examples

## Example 1: Multi-Tenant SaaS with Supabase

Complete multi-tenant application with RLS, real-time, and AI features.

### Database Schema
```sql
-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";

-- Organizations (tenants)
CREATE TABLE organizations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name TEXT NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  plan TEXT DEFAULT 'free' CHECK (plan IN ('free', 'pro', 'enterprise')),
  settings JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Organization members
CREATE TABLE organization_members (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
  user_id UUID NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'viewer')),
  invited_by UUID,
  joined_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(organization_id, user_id)
);

-- Projects within organizations
CREATE TABLE projects (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  description TEXT,
  status TEXT DEFAULT 'active' CHECK (status IN ('active', 'archived', 'deleted')),
  owner_id UUID NOT NULL,
  settings JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Documents with AI embeddings
CREATE TABLE documents (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  embedding vector(1536),
  metadata JSONB DEFAULT '{}',
  created_by UUID NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organization_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

-- RLS Policies
CREATE POLICY "org_member_select" ON organizations
  FOR SELECT USING (
    id IN (
      SELECT organization_id FROM organization_members
      WHERE user_id = auth.uid()
    )
  );

CREATE POLICY "org_admin_update" ON organizations
  FOR UPDATE USING (
    id IN (
      SELECT organization_id FROM organization_members
      WHERE user_id = auth.uid() AND role IN ('owner', 'admin')
    )
  );

CREATE POLICY "project_member_access" ON projects
  FOR ALL USING (
    organization_id IN (
      SELECT organization_id FROM organization_members
      WHERE user_id = auth.uid()
    )
  );

CREATE POLICY "document_project_access" ON documents
  FOR ALL USING (
    project_id IN (
      SELECT p.id FROM projects p
      JOIN organization_members om ON p.organization_id = om.organization_id
      WHERE om.user_id = auth.uid()
    )
  );

-- Semantic search function
CREATE OR REPLACE FUNCTION search_documents(
  p_project_id UUID,
  p_query_embedding vector(1536),
  p_match_threshold FLOAT DEFAULT 0.7,
  p_match_count INT DEFAULT 10
)
RETURNS TABLE (
  id UUID,
  title TEXT,
  content TEXT,
  similarity FLOAT
) LANGUAGE plpgsql SECURITY DEFINER AS $$
BEGIN
  RETURN QUERY
  SELECT
    d.id,
    d.title,
    d.content,
    1 - (d.embedding <=> p_query_embedding) AS similarity
  FROM documents d
  WHERE d.project_id = p_project_id
    AND d.embedding IS NOT NULL
    AND 1 - (d.embedding <=> p_query_embedding) > p_match_threshold
  ORDER BY d.embedding <=> p_query_embedding
  LIMIT p_match_count;
END;
$$;

-- Create indexes
CREATE INDEX idx_documents_project ON documents(project_id);
CREATE INDEX idx_documents_embedding ON documents
  USING hnsw (embedding vector_cosine_ops);
CREATE INDEX idx_org_members_user ON organization_members(user_id);
CREATE INDEX idx_projects_org ON projects(organization_id);
```

### TypeScript Client
```typescript
// lib/supabase/client.ts
import { createClient } from '@supabase/supabase-js'
import { Database } from './database.types'

export const supabase = createClient<Database>(
  process.env.NEXT_PUBLIC_SUPABASE_URL!,
  process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!
)

// lib/supabase/server.ts
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
        get(name: string) {
          return cookieStore.get(name)?.value
        },
        set(name: string, value: string, options) {
          cookieStore.set({ name, value, ...options })
        },
        remove(name: string, options) {
          cookieStore.set({ name, value: '', ...options })
        }
      }
    }
  )
}

// lib/organizations.ts
import { supabase } from './supabase/client'

export class OrganizationService {
  async create(name: string, slug: string) {
    const { data: org, error: orgError } = await supabase
      .from('organizations')
      .insert({ name, slug })
      .select()
      .single()

    if (orgError) throw orgError

    // Add creator as owner
    const { data: { user } } = await supabase.auth.getUser()

    const { error: memberError } = await supabase
      .from('organization_members')
      .insert({
        organization_id: org.id,
        user_id: user!.id,
        role: 'owner'
      })

    if (memberError) throw memberError

    return org
  }

  async invite(organizationId: string, email: string, role: string) {
    const { data, error } = await supabase.functions.invoke('invite-member', {
      body: { organizationId, email, role }
    })

    if (error) throw error
    return data
  }

  async getMembers(organizationId: string) {
    const { data, error } = await supabase
      .from('organization_members')
      .select(`
        *,
        user:user_id (
          id,
          email,
          raw_user_meta_data
        )
      `)
      .eq('organization_id', organizationId)

    if (error) throw error
    return data
  }
}

// lib/documents.ts
import { supabase } from './supabase/client'

export class DocumentService {
  async create(projectId: string, title: string, content: string) {
    const { data: { user } } = await supabase.auth.getUser()

    const { data, error } = await supabase
      .from('documents')
      .insert({
        project_id: projectId,
        title,
        content,
        created_by: user!.id
      })
      .select()
      .single()

    if (error) throw error

    // Generate embedding async
    await supabase.functions.invoke('generate-embedding', {
      body: { documentId: data.id, content }
    })

    return data
  }

  async semanticSearch(projectId: string, query: string) {
    // Get embedding for query
    const { data: embeddingData } = await supabase.functions.invoke(
      'get-embedding',
      { body: { text: query } }
    )

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
    return supabase
      .channel(`documents:${projectId}`)
      .on(
        'postgres_changes',
        {
          event: '*',
          schema: 'public',
          table: 'documents',
          filter: `project_id=eq.${projectId}`
        },
        callback
      )
      .subscribe()
  }
}
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

  const { documentId, content } = await req.json()

  // Generate embedding via OpenAI
  const embeddingResponse = await fetch(
    'https://api.openai.com/v1/embeddings',
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${Deno.env.get('OPENAI_API_KEY')}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        model: 'text-embedding-ada-002',
        input: content.substring(0, 8000) // Limit input
      })
    }
  )

  const embeddingData = await embeddingResponse.json()
  const embedding = embeddingData.data[0].embedding

  // Store embedding
  const supabase = createClient(
    Deno.env.get('SUPABASE_URL')!,
    Deno.env.get('SUPABASE_SERVICE_ROLE_KEY')!
  )

  const { error } = await supabase
    .from('documents')
    .update({ embedding })
    .eq('id', documentId)

  return new Response(
    JSON.stringify({ success: !error }),
    { headers: { ...corsHeaders, 'Content-Type': 'application/json' } }
  )
})
```

---

## Example 2: Preview Environments with Neon

GitHub Actions workflow for database branching.

### GitHub Action
```yaml
# .github/workflows/preview.yml
name: Preview Environment

on:
  pull_request:
    types: [opened, synchronize, reopened, closed]

env:
  NEON_API_KEY: ${{ secrets.NEON_API_KEY }}
  NEON_PROJECT_ID: ${{ secrets.NEON_PROJECT_ID }}

jobs:
  create-preview:
    if: github.event.action != 'closed'
    runs-on: ubuntu-latest
    outputs:
      database_url: ${{ steps.create-branch.outputs.database_url }}

    steps:
      - uses: actions/checkout@v4

      - name: Create Neon Branch
        id: create-branch
        uses: neondatabase/create-branch-action@v4
        with:
          project_id: ${{ env.NEON_PROJECT_ID }}
          branch_name: pr-${{ github.event.number }}
          api_key: ${{ env.NEON_API_KEY }}

      - name: Run Migrations
        run: |
          npm ci
          npx prisma migrate deploy
        env:
          DATABASE_URL: ${{ steps.create-branch.outputs.db_url_with_pooler }}

      - name: Deploy Preview
        uses: vercel/actions/deploy@v2
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
        env:
          DATABASE_URL: ${{ steps.create-branch.outputs.db_url_with_pooler }}

      - name: Comment on PR
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `Preview environment deployed!\n\nDatabase branch: pr-${{ github.event.number }}`
            })

  cleanup-preview:
    if: github.event.action == 'closed'
    runs-on: ubuntu-latest

    steps:
      - name: Delete Neon Branch
        uses: neondatabase/delete-branch-action@v3
        with:
          project_id: ${{ env.NEON_PROJECT_ID }}
          branch: pr-${{ github.event.number }}
          api_key: ${{ env.NEON_API_KEY }}
```

### Neon Client Library
```typescript
// lib/neon/client.ts
import { neon, NeonQueryFunction } from '@neondatabase/serverless'

let sql: NeonQueryFunction<boolean, boolean>

export function getDB() {
  if (!sql) {
    sql = neon(process.env.DATABASE_URL!)
  }
  return sql
}

// lib/neon/users.ts
import { getDB } from './client'

export interface User {
  id: string
  email: string
  name: string
  created_at: Date
}

export async function createUser(email: string, name: string): Promise<User> {
  const sql = getDB()
  const [user] = await sql`
    INSERT INTO users (email, name)
    VALUES (${email}, ${name})
    RETURNING *
  `
  return user as User
}

export async function getUserByEmail(email: string): Promise<User | null> {
  const sql = getDB()
  const [user] = await sql`
    SELECT * FROM users WHERE email = ${email}
  `
  return user as User || null
}

export async function updateUser(
  id: string,
  data: Partial<Pick<User, 'name' | 'email'>>
): Promise<User> {
  const sql = getDB()
  const updates: string[] = []
  const values: any[] = []

  if (data.name) {
    updates.push('name = $' + (values.length + 2))
    values.push(data.name)
  }
  if (data.email) {
    updates.push('email = $' + (values.length + 2))
    values.push(data.email)
  }

  const [user] = await sql`
    UPDATE users
    SET ${sql.unsafe(updates.join(', '))}, updated_at = NOW()
    WHERE id = ${id}
    RETURNING *
  `
  return user as User
}

// lib/neon/transactions.ts
import { getDB } from './client'

export async function transferFunds(
  fromAccountId: string,
  toAccountId: string,
  amount: number
): Promise<{ success: boolean; transactionId: string }> {
  const sql = getDB()

  // Neon supports transactions
  const results = await sql.transaction([
    sql`
      UPDATE accounts
      SET balance = balance - ${amount}
      WHERE id = ${fromAccountId} AND balance >= ${amount}
      RETURNING id
    `,
    sql`
      UPDATE accounts
      SET balance = balance + ${amount}
      WHERE id = ${toAccountId}
      RETURNING id
    `,
    sql`
      INSERT INTO transactions (from_account, to_account, amount)
      VALUES (${fromAccountId}, ${toAccountId}, ${amount})
      RETURNING id
    `
  ])

  if (!results[0].length) {
    throw new Error('Insufficient funds')
  }

  return {
    success: true,
    transactionId: results[2][0].id
  }
}
```

---

## Example 3: Real-time Collaboration with Convex

Full collaborative document editor with presence.

### Convex Backend
```typescript
// convex/schema.ts
import { defineSchema, defineTable } from 'convex/server'
import { v } from 'convex/values'

export default defineSchema({
  documents: defineTable({
    title: v.string(),
    content: v.string(),
    ownerId: v.string(),
    collaborators: v.array(v.string()),
    lastModified: v.number(),
    version: v.number()
  })
    .index('by_owner', ['ownerId'])
    .index('by_collaborator', ['collaborators']),

  documentVersions: defineTable({
    documentId: v.id('documents'),
    content: v.string(),
    version: v.number(),
    createdBy: v.string(),
    createdAt: v.number()
  })
    .index('by_document', ['documentId', 'version']),

  presence: defineTable({
    documentId: v.id('documents'),
    oderId: v.string(),
    cursor: v.optional(v.object({
      line: v.number(),
      column: v.number()
    })),
    selection: v.optional(v.object({
      start: v.object({ line: v.number(), column: v.number() }),
      end: v.object({ line: v.number(), column: v.number() })
    })),
    lastSeen: v.number()
  })
    .index('by_document', ['documentId'])
    .index('by_user', ['userId'])
})

// convex/documents.ts
import { mutation, query } from './_generated/server'
import { v } from 'convex/values'

export const get = query({
  args: { id: v.id('documents') },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new Error('Unauthorized')

    const doc = await ctx.db.get(args.id)
    if (!doc) return null

    // Check access
    if (
      doc.ownerId !== identity.subject &&
      !doc.collaborators.includes(identity.subject)
    ) {
      throw new Error('Access denied')
    }

    return doc
  }
})

export const update = mutation({
  args: {
    id: v.id('documents'),
    content: v.string(),
    expectedVersion: v.number()
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new Error('Unauthorized')

    const doc = await ctx.db.get(args.id)
    if (!doc) throw new Error('Document not found')

    // Optimistic concurrency control
    if (doc.version !== args.expectedVersion) {
      return { conflict: true, currentVersion: doc.version }
    }

    const newVersion = doc.version + 1

    // Save version history
    await ctx.db.insert('documentVersions', {
      documentId: args.id,
      content: doc.content,
      version: doc.version,
      createdBy: identity.subject,
      createdAt: Date.now()
    })

    // Update document
    await ctx.db.patch(args.id, {
      content: args.content,
      version: newVersion,
      lastModified: Date.now()
    })

    return { success: true, version: newVersion }
  }
})

export const addCollaborator = mutation({
  args: {
    documentId: v.id('documents'),
    userId: v.string()
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new Error('Unauthorized')

    const doc = await ctx.db.get(args.documentId)
    if (!doc) throw new Error('Document not found')

    if (doc.ownerId !== identity.subject) {
      throw new Error('Only owner can add collaborators')
    }

    if (!doc.collaborators.includes(args.userId)) {
      await ctx.db.patch(args.documentId, {
        collaborators: [...doc.collaborators, args.userId]
      })
    }
  }
})

// convex/presence.ts
import { mutation, query } from './_generated/server'
import { v } from 'convex/values'

export const updatePresence = mutation({
  args: {
    documentId: v.id('documents'),
    cursor: v.optional(v.object({
      line: v.number(),
      column: v.number()
    })),
    selection: v.optional(v.object({
      start: v.object({ line: v.number(), column: v.number() }),
      end: v.object({ line: v.number(), column: v.number() })
    }))
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) return

    const existing = await ctx.db
      .query('presence')
      .withIndex('by_document', (q) => q.eq('documentId', args.documentId))
      .filter((q) => q.eq(q.field('userId'), identity.subject))
      .unique()

    if (existing) {
      await ctx.db.patch(existing._id, {
        cursor: args.cursor,
        selection: args.selection,
        lastSeen: Date.now()
      })
    } else {
      await ctx.db.insert('presence', {
        documentId: args.documentId,
        userId: identity.subject,
        cursor: args.cursor,
        selection: args.selection,
        lastSeen: Date.now()
      })
    }
  }
})

export const getActiveUsers = query({
  args: { documentId: v.id('documents') },
  handler: async (ctx, args) => {
    const fiveMinutesAgo = Date.now() - 5 * 60 * 1000

    return await ctx.db
      .query('presence')
      .withIndex('by_document', (q) => q.eq('documentId', args.documentId))
      .filter((q) => q.gt(q.field('lastSeen'), fiveMinutesAgo))
      .collect()
  }
})
```

### React Components
```typescript
// components/CollaborativeEditor.tsx
'use client'

import { useQuery, useMutation } from 'convex/react'
import { api } from '@/convex/_generated/api'
import { Id } from '@/convex/_generated/dataModel'
import { useEffect, useState, useCallback, useRef } from 'react'
import { useUser } from '@clerk/nextjs'
import debounce from 'lodash/debounce'

interface EditorProps {
  documentId: Id<'documents'>
}

export function CollaborativeEditor({ documentId }: EditorProps) {
  const { user } = useUser()
  const document = useQuery(api.documents.get, { id: documentId })
  const activeUsers = useQuery(api.presence.getActiveUsers, { documentId })
  const updateDocument = useMutation(api.documents.update)
  const updatePresence = useMutation(api.presence.updatePresence)

  const [localContent, setLocalContent] = useState('')
  const [version, setVersion] = useState(0)
  const [saving, setSaving] = useState(false)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  // Sync document content
  useEffect(() => {
    if (document && document.version !== version) {
      setLocalContent(document.content)
      setVersion(document.version)
    }
  }, [document])

  // Debounced save
  const debouncedSave = useCallback(
    debounce(async (content: string, currentVersion: number) => {
      setSaving(true)
      const result = await updateDocument({
        id: documentId,
        content,
        expectedVersion: currentVersion
      })

      if (result.conflict) {
        // Handle conflict - reload document
        console.warn('Version conflict detected')
      } else if (result.success) {
        setVersion(result.version)
      }
      setSaving(false)
    }, 500),
    [documentId]
  )

  // Update presence on cursor change
  const handleSelectionChange = useCallback(() => {
    if (!textareaRef.current) return

    const start = textareaRef.current.selectionStart
    const end = textareaRef.current.selectionEnd
    const content = textareaRef.current.value

    const getPosition = (index: number) => {
      const lines = content.substring(0, index).split('\n')
      return {
        line: lines.length - 1,
        column: lines[lines.length - 1].length
      }
    }

    updatePresence({
      documentId,
      cursor: getPosition(end),
      selection: start !== end ? {
        start: getPosition(start),
        end: getPosition(end)
      } : undefined
    })
  }, [documentId])

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const content = e.target.value
    setLocalContent(content)
    debouncedSave(content, version)
  }

  if (!document) {
    return <div className="animate-pulse">Loading...</div>
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header with active users */}
      <div className="flex items-center justify-between p-4 border-b">
        <div>
          <h1 className="text-xl font-semibold">{document.title}</h1>
          <p className="text-sm text-gray-500">
            {saving ? 'Saving...' : `Version ${version}`}
          </p>
        </div>

        <div className="flex items-center gap-2">
          {activeUsers?.filter(u => u.userId !== user?.id).map((presence) => (
            <div
              key={presence._id}
              className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white text-sm"
              title={presence.userId}
            >
              {presence.userId.charAt(0).toUpperCase()}
            </div>
          ))}
        </div>
      </div>

      {/* Editor */}
      <div className="flex-1 p-4">
        <textarea
          ref={textareaRef}
          value={localContent}
          onChange={handleChange}
          onSelect={handleSelectionChange}
          className="w-full h-full p-4 font-mono text-sm border rounded-lg
                     resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Start typing..."
        />
      </div>

      {/* Cursor indicators from other users would be rendered here */}
    </div>
  )
}
```

---

## Example 4: Mobile App with Firestore Offline Sync

React Native app with offline-first architecture.

### Firestore Setup
```typescript
// lib/firebase.ts
import { initializeApp } from 'firebase/app'
import {
  getFirestore,
  initializeFirestore,
  persistentLocalCache,
  persistentMultipleTabManager,
  CACHE_SIZE_UNLIMITED
} from 'firebase/firestore'

const firebaseConfig = {
  apiKey: process.env.EXPO_PUBLIC_FIREBASE_API_KEY,
  authDomain: process.env.EXPO_PUBLIC_FIREBASE_AUTH_DOMAIN,
  projectId: process.env.EXPO_PUBLIC_FIREBASE_PROJECT_ID
}

const app = initializeApp(firebaseConfig)

// Initialize with offline persistence
export const db = initializeFirestore(app, {
  localCache: persistentLocalCache({
    tabManager: persistentMultipleTabManager(),
    cacheSizeBytes: CACHE_SIZE_UNLIMITED
  })
})

// lib/hooks/useTasks.ts
import { useEffect, useState } from 'react'
import {
  collection,
  query,
  where,
  orderBy,
  onSnapshot,
  addDoc,
  updateDoc,
  deleteDoc,
  doc,
  serverTimestamp
} from 'firebase/firestore'
import { db } from '../firebase'
import { useAuth } from './useAuth'

export interface Task {
  id: string
  title: string
  completed: boolean
  createdAt: Date
  updatedAt: Date
  _pending?: boolean
  _fromCache?: boolean
}

export function useTasks() {
  const { user } = useAuth()
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    if (!user) {
      setTasks([])
      setLoading(false)
      return
    }

    const q = query(
      collection(db, 'tasks'),
      where('userId', '==', user.uid),
      orderBy('createdAt', 'desc')
    )

    const unsubscribe = onSnapshot(
      q,
      { includeMetadataChanges: true },
      (snapshot) => {
        const newTasks = snapshot.docs.map((doc) => ({
          id: doc.id,
          ...doc.data(),
          createdAt: doc.data().createdAt?.toDate(),
          updatedAt: doc.data().updatedAt?.toDate(),
          _pending: doc.metadata.hasPendingWrites,
          _fromCache: doc.metadata.fromCache
        })) as Task[]

        setTasks(newTasks)
        setLoading(false)
      },
      (err) => {
        setError(err)
        setLoading(false)
      }
    )

    return unsubscribe
  }, [user])

  const addTask = async (title: string) => {
    if (!user) throw new Error('Not authenticated')

    await addDoc(collection(db, 'tasks'), {
      title,
      completed: false,
      userId: user.uid,
      createdAt: serverTimestamp(),
      updatedAt: serverTimestamp()
    })
  }

  const toggleTask = async (taskId: string, completed: boolean) => {
    await updateDoc(doc(db, 'tasks', taskId), {
      completed,
      updatedAt: serverTimestamp()
    })
  }

  const deleteTask = async (taskId: string) => {
    await deleteDoc(doc(db, 'tasks', taskId))
  }

  return {
    tasks,
    loading,
    error,
    addTask,
    toggleTask,
    deleteTask
  }
}
```

### React Native Component
```typescript
// screens/TasksScreen.tsx
import React, { useState } from 'react'
import {
  View,
  Text,
  FlatList,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator
} from 'react-native'
import { useTasks, Task } from '../lib/hooks/useTasks'
import { useNetInfo } from '@react-native-community/netinfo'

export function TasksScreen() {
  const { tasks, loading, addTask, toggleTask, deleteTask } = useTasks()
  const [newTaskTitle, setNewTaskTitle] = useState('')
  const netInfo = useNetInfo()

  const handleAddTask = async () => {
    if (!newTaskTitle.trim()) return

    await addTask(newTaskTitle.trim())
    setNewTaskTitle('')
  }

  const renderTask = ({ item }: { item: Task }) => (
    <View style={[styles.taskItem, item._pending && styles.pendingTask]}>
      <TouchableOpacity
        style={styles.checkbox}
        onPress={() => toggleTask(item.id, !item.completed)}
      >
        <View style={[
          styles.checkboxInner,
          item.completed && styles.checkboxChecked
        ]} />
      </TouchableOpacity>

      <View style={styles.taskContent}>
        <Text style={[
          styles.taskTitle,
          item.completed && styles.taskCompleted
        ]}>
          {item.title}
        </Text>
        {item._pending && (
          <Text style={styles.pendingLabel}>Syncing...</Text>
        )}
        {item._fromCache && !item._pending && (
          <Text style={styles.cacheLabel}>Cached</Text>
        )}
      </View>

      <TouchableOpacity
        style={styles.deleteButton}
        onPress={() => deleteTask(item.id)}
      >
        <Text style={styles.deleteButtonText}>Delete</Text>
      </TouchableOpacity>
    </View>
  )

  if (loading) {
    return (
      <View style={styles.centered}>
        <ActivityIndicator size="large" />
      </View>
    )
  }

  return (
    <View style={styles.container}>
      {/* Offline indicator */}
      {!netInfo.isConnected && (
        <View style={styles.offlineBanner}>
          <Text style={styles.offlineText}>
            Offline - Changes will sync when connected
          </Text>
        </View>
      )}

      {/* Add task input */}
      <View style={styles.inputContainer}>
        <TextInput
          style={styles.input}
          value={newTaskTitle}
          onChangeText={setNewTaskTitle}
          placeholder="New task..."
          onSubmitEditing={handleAddTask}
        />
        <TouchableOpacity style={styles.addButton} onPress={handleAddTask}>
          <Text style={styles.addButtonText}>Add</Text>
        </TouchableOpacity>
      </View>

      {/* Task list */}
      <FlatList
        data={tasks}
        renderItem={renderTask}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.listContent}
      />
    </View>
  )
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#fff' },
  centered: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  offlineBanner: { backgroundColor: '#f59e0b', padding: 8 },
  offlineText: { color: '#fff', textAlign: 'center', fontSize: 12 },
  inputContainer: { flexDirection: 'row', padding: 16, borderBottomWidth: 1, borderBottomColor: '#e5e7eb' },
  input: { flex: 1, borderWidth: 1, borderColor: '#d1d5db', borderRadius: 8, paddingHorizontal: 12, paddingVertical: 8, marginRight: 8 },
  addButton: { backgroundColor: '#3b82f6', paddingHorizontal: 16, paddingVertical: 8, borderRadius: 8, justifyContent: 'center' },
  addButtonText: { color: '#fff', fontWeight: '600' },
  listContent: { padding: 16 },
  taskItem: { flexDirection: 'row', alignItems: 'center', paddingVertical: 12, borderBottomWidth: 1, borderBottomColor: '#e5e7eb' },
  pendingTask: { opacity: 0.7 },
  checkbox: { width: 24, height: 24, borderWidth: 2, borderColor: '#d1d5db', borderRadius: 4, marginRight: 12, justifyContent: 'center', alignItems: 'center' },
  checkboxInner: { width: 14, height: 14, borderRadius: 2 },
  checkboxChecked: { backgroundColor: '#3b82f6' },
  taskContent: { flex: 1 },
  taskTitle: { fontSize: 16 },
  taskCompleted: { textDecorationLine: 'line-through', color: '#9ca3af' },
  pendingLabel: { fontSize: 10, color: '#f59e0b', marginTop: 2 },
  cacheLabel: { fontSize: 10, color: '#6b7280', marginTop: 2 },
  deleteButton: { paddingHorizontal: 12, paddingVertical: 6 },
  deleteButtonText: { color: '#ef4444', fontSize: 12 }
})
```

---

## Migration Scripts

### Supabase to Neon
```python
# migrate_supabase_to_neon.py
import os
import subprocess
from neon_api import NeonClient

class SupabaseToNeonMigration:
    def __init__(self):
        self.neon = NeonClient(api_key=os.getenv('NEON_API_KEY'))
        self.supabase_url = os.getenv('SUPABASE_DB_URL')
        self.neon_project_id = os.getenv('NEON_PROJECT_ID')

    def export_schema(self) -> str:
        result = subprocess.run(
            ['pg_dump', '--schema-only', '--no-owner', self.supabase_url],
            capture_output=True,
            text=True
        )
        return result.stdout

    def export_data(self) -> str:
        result = subprocess.run(
            ['pg_dump', '--data-only', '--no-owner', self.supabase_url],
            capture_output=True,
            text=True
        )
        return result.stdout

    def create_neon_branch(self, name: str):
        return self.neon.create_branch(
            project_id=self.neon_project_id,
            branch_name=name
        )

    def import_to_neon(self, connection_string: str, schema: str, data: str):
        # Import schema
        subprocess.run(
            ['psql', connection_string],
            input=schema,
            text=True
        )

        # Import data
        subprocess.run(
            ['psql', connection_string],
            input=data,
            text=True
        )

    def run(self):
        print("Exporting from Supabase...")
        schema = self.export_schema()
        data = self.export_data()

        print("Creating Neon branch...")
        branch = self.create_neon_branch('migration')
        connection_string = branch.connection_uri

        print("Importing to Neon...")
        self.import_to_neon(connection_string, schema, data)

        print("Migration complete!")

if __name__ == '__main__':
    migration = SupabaseToNeonMigration()
    migration.run()
```

---

Last Updated: 2025-12-07
Examples: Production-ready implementations
Status: Production Ready
