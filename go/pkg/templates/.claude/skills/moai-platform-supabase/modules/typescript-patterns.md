---
name: typescript-patterns
description: TypeScript client patterns and service layer architecture for Supabase
parent-skill: moai-platform-supabase
version: 1.0.0
updated: 2026-01-06
---

# TypeScript Patterns Module

## Overview

Type-safe Supabase client patterns for building maintainable full-stack applications with TypeScript.

## Type Generation

### Generate Types from Database

```bash
supabase gen types typescript --project-id your-project-id > database.types.ts
```

### Database Types Structure

```typescript
// database.types.ts
export type Database = {
  public: {
    Tables: {
      projects: {
        Row: {
          id: string
          name: string
          organization_id: string
          owner_id: string
          created_at: string
        }
        Insert: {
          id?: string
          name: string
          organization_id: string
          owner_id: string
          created_at?: string
        }
        Update: {
          id?: string
          name?: string
          organization_id?: string
          owner_id?: string
          created_at?: string
        }
      }
      // ... other tables
    }
    Functions: {
      search_documents: {
        Args: {
          query_embedding: number[]
          match_threshold: number
          match_count: number
        }
        Returns: { id: string; content: string; similarity: number }[]
      }
    }
  }
}
```

## Client Configuration

### Browser Client with Types

```typescript
import { createClient } from '@supabase/supabase-js'
import { Database } from './database.types'

export const supabase = createClient<Database>(
  process.env.NEXT_PUBLIC_SUPABASE_URL!,
  process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!
)
```

### Server Client (Next.js App Router)

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
        get(name: string) {
          return cookieStore.get(name)?.value
        },
        set(name, value, options) {
          cookieStore.set({ name, value, ...options })
        },
        remove(name, options) {
          cookieStore.set({ name, value: '', ...options })
        }
      }
    }
  )
}
```

## Service Layer Pattern

### Base Service

```typescript
import { supabase } from './supabase/client'
import { Database } from './database.types'

type Tables = Database['public']['Tables']

export abstract class BaseService<T extends keyof Tables> {
  constructor(protected tableName: T) {}

  async findAll() {
    const { data, error } = await supabase
      .from(this.tableName)
      .select('*')

    if (error) throw error
    return data as Tables[T]['Row'][]
  }

  async findById(id: string) {
    const { data, error } = await supabase
      .from(this.tableName)
      .select('*')
      .eq('id', id)
      .single()

    if (error) throw error
    return data as Tables[T]['Row']
  }

  async create(item: Tables[T]['Insert']) {
    const { data, error } = await supabase
      .from(this.tableName)
      .insert(item)
      .select()
      .single()

    if (error) throw error
    return data as Tables[T]['Row']
  }

  async update(id: string, item: Tables[T]['Update']) {
    const { data, error } = await supabase
      .from(this.tableName)
      .update(item)
      .eq('id', id)
      .select()
      .single()

    if (error) throw error
    return data as Tables[T]['Row']
  }

  async delete(id: string) {
    const { error } = await supabase
      .from(this.tableName)
      .delete()
      .eq('id', id)

    if (error) throw error
  }
}
```

### Document Service with Embeddings

```typescript
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

    // Generate embedding asynchronously
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

    // Search using RPC
    const { data, error } = await supabase.rpc('search_documents', {
      p_project_id: projectId,
      p_query_embedding: embeddingData.embedding,
      p_match_threshold: 0.7,
      p_match_count: 10
    })

    if (error) throw error
    return data
  }

  async findByProject(projectId: string) {
    const { data, error } = await supabase
      .from('documents')
      .select('*, created_by_user:profiles!created_by(*)')
      .eq('project_id', projectId)
      .order('created_at', { ascending: false })

    if (error) throw error
    return data
  }

  subscribeToChanges(projectId: string, callback: (payload: any) => void) {
    return supabase.channel(`documents:${projectId}`)
      .on('postgres_changes',
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

export const documentService = new DocumentService()
```

## React Query Integration

### Query Keys

```typescript
export const queryKeys = {
  projects: {
    all: ['projects'] as const,
    list: (filters?: ProjectFilters) => [...queryKeys.projects.all, 'list', filters] as const,
    detail: (id: string) => [...queryKeys.projects.all, 'detail', id] as const
  },
  documents: {
    all: ['documents'] as const,
    list: (projectId: string) => [...queryKeys.documents.all, 'list', projectId] as const,
    search: (projectId: string, query: string) =>
      [...queryKeys.documents.all, 'search', projectId, query] as const
  }
}
```

### Custom Hooks

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { documentService } from '@/services/document-service'

export function useDocuments(projectId: string) {
  return useQuery({
    queryKey: queryKeys.documents.list(projectId),
    queryFn: () => documentService.findByProject(projectId)
  })
}

export function useSemanticSearch(projectId: string, query: string) {
  return useQuery({
    queryKey: queryKeys.documents.search(projectId, query),
    queryFn: () => documentService.semanticSearch(projectId, query),
    enabled: query.length > 2
  })
}

export function useCreateDocument(projectId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ title, content }: { title: string; content: string }) =>
      documentService.create(projectId, title, content),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.documents.list(projectId)
      })
    }
  })
}
```

## Real-time with React

### Subscription Hook

```typescript
import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { supabase } from '@/lib/supabase/client'

export function useRealtimeDocuments(projectId: string) {
  const queryClient = useQueryClient()

  useEffect(() => {
    const channel = supabase
      .channel(`documents:${projectId}`)
      .on('postgres_changes',
        {
          event: '*',
          schema: 'public',
          table: 'documents',
          filter: `project_id=eq.${projectId}`
        },
        () => {
          queryClient.invalidateQueries({
            queryKey: queryKeys.documents.list(projectId)
          })
        }
      )
      .subscribe()

    return () => {
      supabase.removeChannel(channel)
    }
  }, [projectId, queryClient])
}
```

### Optimistic Updates

```typescript
export function useUpdateDocument() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, updates }: { id: string; updates: Partial<Document> }) => {
      const { data, error } = await supabase
        .from('documents')
        .update(updates)
        .eq('id', id)
        .select()
        .single()

      if (error) throw error
      return data
    },
    onMutate: async ({ id, updates }) => {
      await queryClient.cancelQueries({ queryKey: ['documents'] })

      const previousDocuments = queryClient.getQueryData(['documents'])

      queryClient.setQueryData(['documents'], (old: Document[]) =>
        old.map(doc => doc.id === id ? { ...doc, ...updates } : doc)
      )

      return { previousDocuments }
    },
    onError: (err, variables, context) => {
      if (context?.previousDocuments) {
        queryClient.setQueryData(['documents'], context.previousDocuments)
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['documents'] })
    }
  })
}
```

## Error Handling

### Custom Error Types

```typescript
export class SupabaseError extends Error {
  constructor(
    message: string,
    public code: string,
    public details?: unknown
  ) {
    super(message)
    this.name = 'SupabaseError'
  }
}

export function handleSupabaseError(error: PostgrestError): never {
  switch (error.code) {
    case '23505':
      throw new SupabaseError('Resource already exists', 'DUPLICATE', error)
    case '23503':
      throw new SupabaseError('Referenced resource not found', 'NOT_FOUND', error)
    case 'PGRST116':
      throw new SupabaseError('Resource not found', 'NOT_FOUND', error)
    default:
      throw new SupabaseError(error.message, error.code, error)
  }
}
```

### Service with Error Handling

```typescript
async findById(id: string) {
  const { data, error } = await supabase
    .from(this.tableName)
    .select('*')
    .eq('id', id)
    .single()

  if (error) {
    handleSupabaseError(error)
  }

  return data
}
```

## Context7 Query Examples

For latest client documentation:

Topic: "supabase-js typescript client"
Topic: "supabase ssr next.js app router"
Topic: "supabase realtime subscription"

---

Related Modules:
- auth-integration.md - Auth patterns
- realtime-presence.md - Real-time subscriptions
- postgresql-pgvector.md - Database operations
