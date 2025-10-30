---
title: Supabase PostgreSQL & Auth Best Practices
description: Master Supabase for PostgreSQL database, authentication, real-time subscriptions, and schema migrations
freedom_level: high
tier: saas
updated: 2025-10-31
---

# Supabase PostgreSQL & Auth Best Practices

## Overview

Supabase is an open-source Firebase alternative built on PostgreSQL, offering database, authentication, real-time subscriptions, and storage. This skill covers schema design, Row-Level Security (RLS), authentication patterns, migration strategies, and production best practices for 2025.

## Key Patterns

### 1. Row-Level Security (RLS) for Multi-Tenancy

**Pattern**: Use PostgreSQL RLS to enforce data isolation at database level.

```sql
-- Enable RLS on tables
ALTER TABLE posts ENABLE ROW LEVEL SECURITY;

-- Policy: Users can only see their own posts
CREATE POLICY "Users can view own posts"
ON posts
FOR SELECT
USING (auth.uid() = user_id);

-- Policy: Users can insert posts with their own user_id
CREATE POLICY "Users can create posts"
ON posts
FOR INSERT
WITH CHECK (auth.uid() = user_id);

-- Policy: Users can update their own posts
CREATE POLICY "Users can update own posts"
ON posts
FOR UPDATE
USING (auth.uid() = user_id)
WITH CHECK (auth.uid() = user_id);

-- Policy: Public posts visible to everyone
CREATE POLICY "Public posts are viewable by everyone"
ON posts
FOR SELECT
USING (is_public = true);
```

**CRITICAL**: Always enable RLS on tables with user data. Queries bypass RLS when using service role key.

### 2. Schema Migrations with Supabase CLI

**Pattern**: Manage schema changes via migrations, not UI.

```bash
# Initialize Supabase locally
supabase init

# Start local development
supabase start

# Create migration
supabase migration new create_posts_table

# Write migration SQL
```

```sql
-- supabase/migrations/20231031000001_create_posts_table.sql
CREATE TABLE posts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  content TEXT,
  is_public BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Add indexes
CREATE INDEX posts_user_id_idx ON posts(user_id);
CREATE INDEX posts_created_at_idx ON posts(created_at DESC);

-- Enable RLS
ALTER TABLE posts ENABLE ROW LEVEL SECURITY;

-- Create policies (as shown in Pattern 1)
```

```bash
# Apply migration locally
supabase db reset

# Push to production
supabase db push
```

### 3. Authentication Patterns

**Pattern**: Use Supabase Auth with proper session management.

```typescript
// lib/supabase.ts
import { createClient } from '@supabase/supabase-js'

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL!
const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!

export const supabase = createClient(supabaseUrl, supabaseAnonKey, {
  auth: {
    persistSession: true,
    autoRefreshToken: true,
    detectSessionInUrl: true
  }
})

// app/auth/sign-in/page.tsx
'use client'

import { useState } from 'react'
import { supabase } from '@/lib/supabase'

export default function SignIn() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  
  const handleSignIn = async (e: React.FormEvent) => {
    e.preventDefault()
    
    const { data, error } = await supabase.auth.signInWithPassword({
      email,
      password
    })
    
    if (error) {
      console.error('Error:', error.message)
    } else {
      // Redirect to dashboard
      window.location.href = '/dashboard'
    }
  }
  
  return (
    <form onSubmit={handleSignIn}>
      <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
      <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
      <button type="submit">Sign In</button>
    </form>
  )
}
```

**Session Refresh Pattern**:
```typescript
// app/layout.tsx
'use client'

import { useEffect } from 'react'
import { supabase } from '@/lib/supabase'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    // Listen for auth state changes
    const { data: { subscription } } = supabase.auth.onAuthStateChange(
      async (event, session) => {
        if (event === 'SIGNED_OUT') {
          window.location.href = '/auth/sign-in'
        }
        
        if (event === 'TOKEN_REFRESHED') {
          console.log('Token refreshed:', session)
        }
      }
    )
    
    return () => {
      subscription.unsubscribe()
    }
  }, [])
  
  return <html><body>{children}</body></html>
}
```

### 4. Real-Time Subscriptions

**Pattern**: Use Postgres LISTEN/NOTIFY for real-time updates.

```typescript
// Enable real-time on table (in Supabase dashboard or SQL)
ALTER PUBLICATION supabase_realtime ADD TABLE posts;

// Subscribe to changes
'use client'

import { useEffect, useState } from 'react'
import { supabase } from '@/lib/supabase'

type Post = {
  id: string
  title: string
  content: string
  created_at: string
}

export function RealtimePosts() {
  const [posts, setPosts] = useState<Post[]>([])
  
  useEffect(() => {
    // Fetch initial data
    supabase
      .from('posts')
      .select('*')
      .order('created_at', { ascending: false })
      .then(({ data }) => setPosts(data || []))
    
    // Subscribe to inserts
    const subscription = supabase
      .channel('posts-channel')
      .on(
        'postgres_changes',
        {
          event: 'INSERT',
          schema: 'public',
          table: 'posts'
        },
        (payload) => {
          setPosts((prev) => [payload.new as Post, ...prev])
        }
      )
      .on(
        'postgres_changes',
        {
          event: 'UPDATE',
          schema: 'public',
          table: 'posts'
        },
        (payload) => {
          setPosts((prev) =>
            prev.map((post) =>
              post.id === payload.new.id ? (payload.new as Post) : post
            )
          )
        }
      )
      .subscribe()
    
    return () => {
      subscription.unsubscribe()
    }
  }, [])
  
  return (
    <div>
      {posts.map((post) => (
        <div key={post.id}>{post.title}</div>
      ))}
    </div>
  )
}
```

### 5. Database Functions for Business Logic

**Pattern**: Encapsulate complex logic in PostgreSQL functions.

```sql
-- supabase/migrations/20231031000002_create_like_post_function.sql
CREATE OR REPLACE FUNCTION like_post(post_id UUID)
RETURNS VOID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
  -- Insert like record
  INSERT INTO post_likes (post_id, user_id)
  VALUES (post_id, auth.uid())
  ON CONFLICT (post_id, user_id) DO NOTHING;
  
  -- Increment like count
  UPDATE posts
  SET like_count = like_count + 1
  WHERE id = post_id;
END;
$$;

-- Grant execute permission
GRANT EXECUTE ON FUNCTION like_post TO authenticated;
```

```typescript
// Call from TypeScript
const { data, error } = await supabase.rpc('like_post', {
  post_id: '123e4567-e89b-12d3-a456-426614174000'
})
```

### 6. Service Role vs Anon Key Usage

**Pattern**: Use appropriate keys for different contexts.

```typescript
// ✅ GOOD: Client-side with anon key (respects RLS)
const supabaseClient = createClient(
  process.env.NEXT_PUBLIC_SUPABASE_URL!,
  process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!
)

// ✅ GOOD: Server-side admin operations (bypasses RLS)
const supabaseAdmin = createClient(
  process.env.NEXT_PUBLIC_SUPABASE_URL!,
  process.env.SUPABASE_SERVICE_ROLE_KEY!
)

// Use cases for service role key:
// - Bulk operations
// - Admin dashboard queries
// - Scheduled jobs
// - Data migrations

// ❌ NEVER expose service role key to client
// ❌ NEVER use in browser JavaScript
```

### 7. Connection Pooling for Production

**Pattern**: Use connection pooling for serverless environments.

```typescript
// For serverless functions (Vercel, Netlify, etc.)
import { createClient } from '@supabase/supabase-js'

const supabase = createClient(
  process.env.NEXT_PUBLIC_SUPABASE_URL!,
  process.env.SUPABASE_SERVICE_ROLE_KEY!,
  {
    db: {
      schema: 'public'
    },
    auth: {
      persistSession: false  // Don't persist in serverless
    },
    global: {
      headers: {
        'x-connection-pooling': 'true'  // Enable pooling
      }
    }
  }
)
```

**Direct Connection vs Pooler**:
- **Direct Connection** (Port 5432): Max 100 connections, use for long-running servers
- **Connection Pooler** (Port 6543): Use for serverless, Vercel Edge Functions, high-concurrency

## Checklist

- [ ] Enable RLS on ALL tables with user data
- [ ] Use Supabase CLI for migrations (not UI-based schema changes)
- [ ] Store service role key in server environment only (NEVER in client)
- [ ] Implement JWT refresh logic (`supabase.auth.onAuthStateChange`)
- [ ] Use connection pooler (port 6543) for serverless deployments
- [ ] Create database indexes on frequently queried columns
- [ ] Test RLS policies: verify users can't access others' data
- [ ] Enable real-time only on necessary tables (reduce overhead)
- [ ] Use Postgres functions for complex multi-step operations
- [ ] Set up daily backups (automatic in Supabase for paid plans)

## Resources

- **Official Supabase Docs**: https://supabase.com/docs
- **Row-Level Security Guide**: https://supabase.com/docs/guides/auth/row-level-security
- **CLI Documentation**: https://supabase.com/docs/guides/local-development/overview
- **Migration Best Practices**: https://supabase.com/docs/guides/platform/migrating-to-supabase/postgres
- **Production Security (2025)**: https://medium.com/@firmanbrilian/best-practices-for-securing-and-scaling-supabase-for-production-data-workloads-bdd726313177
- **Common Mistakes**: https://medium.com/@lior_amsalem/3-biggest-mistakes-using-supabase-854fe45712e3

---

**Version**: 1.0.0  
**Last Updated**: 2025-10-31  
**Model Recommendation**: Sonnet (deep reasoning for RLS policies and database architecture)
