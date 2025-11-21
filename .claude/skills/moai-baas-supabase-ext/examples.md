# Supabase Platform Examples & Patterns

**Complete production-ready examples for Supabase PostgreSQL, Auth, Storage, Realtime, and Edge Functions.**

**Research Base**: Official Supabase documentation, Context7 `/websites/supabase` patterns
**Last Updated**: 2025-11-22
**Version**: 1.0.0

---

## 📚 Table of Contents

1. [Authentication Examples](#1-authentication-examples)
2. [Row-Level Security (RLS)](#2-row-level-security-rls)
3. [Realtime Subscriptions](#3-realtime-subscriptions)
4. [Storage Operations](#4-storage-operations)
5. [Database Operations](#5-database-operations)
6. [Edge Functions](#6-edge-functions)
7. [OAuth2 Integration](#7-oauth2-integration)
8. [Multi-Tenant Architecture](#8-multi-tenant-architecture)
9. [Real-time Multiplayer](#9-real-time-multiplayer)
10. [Performance Monitoring](#10-performance-monitoring)

---

## 1. Authentication Examples

### Example 1.1: Email/Password Authentication

**Complete authentication flow with TypeScript**:

```typescript
import { createClient } from '@supabase/supabase-js'

// Initialize Supabase client
const supabase = createClient(
  process.env.SUPABASE_URL!,
  process.env.SUPABASE_ANON_KEY!
)

// Sign up new user
async function signUpUser(email: string, password: string, userData: any) {
  const { data, error } = await supabase.auth.signUp({
    email,
    password,
    options: {
      data: {
        full_name: userData.full_name,
        avatar_url: userData.avatar_url,
        // Custom user metadata
      },
      emailRedirectTo: 'https://yourapp.com/auth/callback'
    }
  })

  if (error) {
    throw new Error(`Sign up failed: ${error.message}`)
  }

  return data
}

// Sign in existing user
async function signInUser(email: string, password: string) {
  const { data, error } = await supabase.auth.signInWithPassword({
    email,
    password
  })

  if (error) {
    throw new Error(`Sign in failed: ${error.message}`)
  }

  return data
}

// Sign out user
async function signOutUser() {
  const { error } = await supabase.auth.signOut()

  if (error) {
    throw new Error(`Sign out failed: ${error.message}`)
  }
}

// Get current session
async function getCurrentSession() {
  const { data: { session }, error } = await supabase.auth.getSession()

  if (error) {
    throw new Error(`Session retrieval failed: ${error.message}`)
  }

  return session
}

// Listen to auth state changes
supabase.auth.onAuthStateChange((event, session) => {
  console.log('Auth event:', event)
  console.log('Session:', session)

  switch (event) {
    case 'SIGNED_IN':
      console.log('User signed in:', session?.user)
      break
    case 'SIGNED_OUT':
      console.log('User signed out')
      break
    case 'TOKEN_REFRESHED':
      console.log('Token refreshed')
      break
    case 'USER_UPDATED':
      console.log('User updated:', session?.user)
      break
  }
})
```

**Usage**:
```typescript
// Sign up
try {
  const user = await signUpUser(
    'user@example.com',
    'SecurePass123!',
    { full_name: 'John Doe', avatar_url: 'https://...' }
  )
  console.log('User created:', user)
} catch (error) {
  console.error('Sign up error:', error)
}

// Sign in
try {
  const session = await signInUser('user@example.com', 'SecurePass123!')
  console.log('Session:', session)
} catch (error) {
  console.error('Sign in error:', error)
}
```

---

## 2. Row-Level Security (RLS)

### Example 2.1: Multi-Tenant RLS Policies

**Complete RLS setup for multi-tenant SaaS**:

```sql
-- Enable RLS on tables
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE tasks ENABLE ROW LEVEL SECURITY;

-- Organizations: Users can only see their own organizations
CREATE POLICY "Users can view their own organizations"
  ON organizations FOR SELECT
  USING (
    auth.uid() IN (
      SELECT user_id FROM organization_members
      WHERE organization_id = organizations.id
    )
  );

CREATE POLICY "Users can insert organizations"
  ON organizations FOR INSERT
  WITH CHECK (auth.uid() = owner_id);

CREATE POLICY "Only owners can update organizations"
  ON organizations FOR UPDATE
  USING (auth.uid() = owner_id)
  WITH CHECK (auth.uid() = owner_id);

-- Projects: Tenant isolation
CREATE POLICY "Users can view organization projects"
  ON projects FOR SELECT
  USING (
    organization_id IN (
      SELECT organization_id FROM organization_members
      WHERE user_id = auth.uid()
    )
  );

CREATE POLICY "Members can create projects"
  ON projects FOR INSERT
  WITH CHECK (
    organization_id IN (
      SELECT organization_id FROM organization_members
      WHERE user_id = auth.uid()
      AND role IN ('admin', 'member')
    )
  );

-- Tasks: Role-based access
CREATE POLICY "Users can view accessible tasks"
  ON tasks FOR SELECT
  USING (
    project_id IN (
      SELECT p.id FROM projects p
      JOIN organization_members om ON om.organization_id = p.organization_id
      WHERE om.user_id = auth.uid()
    )
  );

CREATE POLICY "Only assigned users and admins can update tasks"
  ON tasks FOR UPDATE
  USING (
    assigned_to = auth.uid()
    OR
    project_id IN (
      SELECT p.id FROM projects p
      JOIN organization_members om ON om.organization_id = p.organization_id
      WHERE om.user_id = auth.uid() AND om.role = 'admin'
    )
  );

-- Security definer functions for admin operations
CREATE OR REPLACE FUNCTION assign_task_to_user(
  task_id UUID,
  user_id UUID
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
  -- Verify caller is admin
  IF NOT EXISTS (
    SELECT 1 FROM organization_members om
    JOIN tasks t ON t.project_id IN (
      SELECT id FROM projects WHERE organization_id = om.organization_id
    )
    WHERE om.user_id = auth.uid()
    AND om.role = 'admin'
    AND t.id = task_id
  ) THEN
    RAISE EXCEPTION 'Unauthorized: Admin role required';
  END IF;

  -- Assign task
  UPDATE tasks
  SET assigned_to = user_id,
      updated_at = NOW()
  WHERE id = task_id;
END;
$$;
```

**TypeScript client usage**:

```typescript
// Query with RLS automatically applied
async function getOrganizationProjects(orgId: string) {
  const { data, error } = await supabase
    .from('projects')
    .select('*')
    .eq('organization_id', orgId)

  // RLS ensures user only sees projects they have access to
  if (error) throw error
  return data
}

// Create project with RLS check
async function createProject(orgId: string, name: string) {
  const { data, error } = await supabase
    .from('projects')
    .insert({
      organization_id: orgId,
      name: name,
      created_by: (await supabase.auth.getUser()).data.user?.id
    })
    .select()

  // RLS ensures user is a member of the organization
  if (error) throw error
  return data
}
```

---

## 3. Realtime Subscriptions

### Example 3.1: Real-time Data Sync

**Complete realtime subscription patterns**:

```typescript
import { RealtimeChannel } from '@supabase/supabase-js'

// Subscribe to table changes
function subscribeToTaskUpdates(projectId: string, callback: (payload: any) => void) {
  const channel = supabase
    .channel(`project-${projectId}-tasks`)
    .on(
      'postgres_changes',
      {
        event: '*',  // INSERT, UPDATE, DELETE
        schema: 'public',
        table: 'tasks',
        filter: `project_id=eq.${projectId}`
      },
      (payload) => {
        console.log('Task change:', payload)
        callback(payload)
      }
    )
    .subscribe()

  return channel
}

// Broadcast messages
async function broadcastMessage(channel: RealtimeChannel, message: any) {
  await channel.send({
    type: 'broadcast',
    event: 'message',
    payload: message
  })
}

// Presence tracking
function setupPresence(channel: RealtimeChannel, userId: string) {
  channel
    .on('presence', { event: 'sync' }, () => {
      const state = channel.presenceState()
      console.log('Online users:', state)
    })
    .on('presence', { event: 'join' }, ({ key, newPresences }) => {
      console.log('User joined:', key, newPresences)
    })
    .on('presence', { event: 'leave' }, ({ key, leftPresences }) => {
      console.log('User left:', key, leftPresences)
    })
    .subscribe(async (status) => {
      if (status === 'SUBSCRIBED') {
        await channel.track({
          user_id: userId,
          online_at: new Date().toISOString()
        })
      }
    })

  return channel
}
```

**React hook example**:

```typescript
import { useEffect, useState } from 'react'
import { RealtimeChannel } from '@supabase/supabase-js'

function useRealtimeTasks(projectId: string) {
  const [tasks, setTasks] = useState<any[]>([])
  const [channel, setChannel] = useState<RealtimeChannel | null>(null)

  useEffect(() => {
    // Fetch initial data
    async function fetchTasks() {
      const { data } = await supabase
        .from('tasks')
        .select('*')
        .eq('project_id', projectId)

      if (data) setTasks(data)
    }

    fetchTasks()

    // Subscribe to changes
    const newChannel = subscribeToTaskUpdates(projectId, (payload) => {
      if (payload.eventType === 'INSERT') {
        setTasks((prev) => [...prev, payload.new])
      } else if (payload.eventType === 'UPDATE') {
        setTasks((prev) =>
          prev.map((task) =>
            task.id === payload.new.id ? payload.new : task
          )
        )
      } else if (payload.eventType === 'DELETE') {
        setTasks((prev) =>
          prev.filter((task) => task.id !== payload.old.id)
        )
      }
    })

    setChannel(newChannel)

    // Cleanup
    return () => {
      newChannel.unsubscribe()
    }
  }, [projectId])

  return { tasks, channel }
}
```

---

## 4. Storage Operations

### Example 4.1: File Upload with Progress

**Complete file storage implementation**:

```typescript
// Upload file with progress tracking
async function uploadFileWithProgress(
  bucket: string,
  path: string,
  file: File,
  onProgress?: (progress: number) => void
) {
  const { data, error } = await supabase.storage
    .from(bucket)
    .upload(path, file, {
      cacheControl: '3600',
      upsert: false,
      onUploadProgress: (progress) => {
        const percent = (progress.loaded / progress.total) * 100
        onProgress?.call(null, percent)
      }
    })

  if (error) throw error
  return data
}

// Generate signed URL
async function getSignedUrl(bucket: string, path: string, expiresIn = 60) {
  const { data, error } = await supabase.storage
    .from(bucket)
    .createSignedUrl(path, expiresIn)

  if (error) throw error
  return data.signedUrl
}

// Download file
async function downloadFile(bucket: string, path: string) {
  const { data, error } = await supabase.storage
    .from(bucket)
    .download(path)

  if (error) throw error
  return data
}

// Delete file
async function deleteFile(bucket: string, path: string) {
  const { error } = await supabase.storage
    .from(bucket)
    .remove([path])

  if (error) throw error
}

// List files
async function listFiles(bucket: string, folder?: string) {
  const { data, error } = await supabase.storage
    .from(bucket)
    .list(folder, {
      limit: 100,
      offset: 0,
      sortBy: { column: 'name', order: 'asc' }
    })

  if (error) throw error
  return data
}
```

**React component example**:

```tsx
import { useState } from 'react'

function FileUploader() {
  const [uploading, setUploading] = useState(false)
  const [progress, setProgress] = useState(0)

  async function handleUpload(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    if (!file) return

    setUploading(true)

    try {
      const filePath = `${Date.now()}-${file.name}`

      await uploadFileWithProgress(
        'avatars',
        filePath,
        file,
        (percent) => setProgress(percent)
      )

      // Get public URL
      const { data: { publicUrl } } = supabase.storage
        .from('avatars')
        .getPublicUrl(filePath)

      console.log('File uploaded:', publicUrl)
    } catch (error) {
      console.error('Upload error:', error)
    } finally {
      setUploading(false)
      setProgress(0)
    }
  }

  return (
    <div>
      <input type="file" onChange={handleUpload} disabled={uploading} />
      {uploading && <progress value={progress} max="100" />}
    </div>
  )
}
```

---

## 5. Database Operations

### Example 5.1: Advanced Queries with Joins

**Complex database queries**:

```typescript
// Query with joins and filters
async function getProjectsWithTasks(userId: string) {
  const { data, error } = await supabase
    .from('projects')
    .select(`
      *,
      organization:organizations(*),
      tasks(
        id,
        title,
        status,
        assigned_to,
        assignee:users(id, email, full_name)
      )
    `)
    .eq('tasks.assigned_to', userId)
    .order('created_at', { ascending: false })

  if (error) throw error
  return data
}

// Aggregation query
async function getProjectStats(projectId: string) {
  const { data, error } = await supabase.rpc('get_project_stats', {
    project_id: projectId
  })

  if (error) throw error
  return data
}

// Database function for stats
/*
CREATE OR REPLACE FUNCTION get_project_stats(project_id UUID)
RETURNS TABLE(
  total_tasks BIGINT,
  completed_tasks BIGINT,
  in_progress_tasks BIGINT,
  completion_rate NUMERIC
)
LANGUAGE plpgsql
AS $$
BEGIN
  RETURN QUERY
  SELECT
    COUNT(*)::BIGINT AS total_tasks,
    COUNT(*) FILTER (WHERE status = 'completed')::BIGINT AS completed_tasks,
    COUNT(*) FILTER (WHERE status = 'in_progress')::BIGINT AS in_progress_tasks,
    CASE
      WHEN COUNT(*) > 0 THEN
        ROUND((COUNT(*) FILTER (WHERE status = 'completed')::NUMERIC / COUNT(*)::NUMERIC) * 100, 2)
      ELSE 0
    END AS completion_rate
  FROM tasks
  WHERE tasks.project_id = get_project_stats.project_id;
END;
$$;
*/

// Batch insert
async function batchCreateTasks(projectId: string, tasks: any[]) {
  const { data, error } = await supabase
    .from('tasks')
    .insert(
      tasks.map(task => ({
        ...task,
        project_id: projectId
      }))
    )
    .select()

  if (error) throw error
  return data
}

// Transaction using RPC
async function transferTaskOwnership(taskId: string, newOwnerId: string) {
  const { data, error } = await supabase.rpc('transfer_task', {
    task_id: taskId,
    new_owner_id: newOwnerId
  })

  if (error) throw error
  return data
}
```

---

## 6. Edge Functions

### Example 6.1: Serverless API Endpoint

**Deno Edge Function example**:

```typescript
// supabase/functions/send-notification/index.ts
import { serve } from 'https://deno.land/std@0.168.0/http/server.ts'
import { createClient } from 'https://esm.sh/@supabase/supabase-js@2'

serve(async (req) => {
  try {
    // CORS headers
    if (req.method === 'OPTIONS') {
      return new Response('ok', {
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type',
        }
      })
    }

    // Authenticate request
    const authHeader = req.headers.get('Authorization')
    if (!authHeader) {
      throw new Error('Missing authorization header')
    }

    const supabase = createClient(
      Deno.env.get('SUPABASE_URL') ?? '',
      Deno.env.get('SUPABASE_ANON_KEY') ?? '',
      {
        global: {
          headers: { Authorization: authHeader }
        }
      }
    )

    // Verify user
    const { data: { user }, error: userError } = await supabase.auth.getUser()
    if (userError || !user) {
      throw new Error('Unauthorized')
    }

    // Parse request body
    const { taskId, message } = await req.json()

    // Business logic: Send notification
    const { data: task, error: taskError } = await supabase
      .from('tasks')
      .select('*, assigned_to')
      .eq('id', taskId)
      .single()

    if (taskError) throw taskError

    // Insert notification
    const { error: notificationError } = await supabase
      .from('notifications')
      .insert({
        user_id: task.assigned_to,
        message: message,
        task_id: taskId,
        created_by: user.id
      })

    if (notificationError) throw notificationError

    return new Response(
      JSON.stringify({ success: true }),
      {
        headers: {
          'Content-Type': 'application/json',
          'Access-Control-Allow-Origin': '*'
        }
      }
    )
  } catch (error) {
    return new Response(
      JSON.stringify({ error: error.message }),
      {
        status: 400,
        headers: {
          'Content-Type': 'application/json',
          'Access-Control-Allow-Origin': '*'
        }
      }
    )
  }
})
```

**Invoke from client**:

```typescript
async function sendNotification(taskId: string, message: string) {
  const { data, error } = await supabase.functions.invoke('send-notification', {
    body: { taskId, message }
  })

  if (error) throw error
  return data
}
```

---

## 7. OAuth2 Integration

### Example 7.1: Google OAuth

**OAuth provider setup**:

```typescript
// Sign in with Google
async function signInWithGoogle() {
  const { data, error } = await supabase.auth.signInWithOAuth({
    provider: 'google',
    options: {
      redirectTo: 'https://yourapp.com/auth/callback',
      scopes: 'email profile',
      queryParams: {
        access_type: 'offline',
        prompt: 'consent'
      }
    }
  })

  if (error) throw error
  return data
}

// Handle OAuth callback
async function handleOAuthCallback() {
  const { data: { session }, error } = await supabase.auth.getSession()

  if (error) throw error

  if (session) {
    // User authenticated successfully
    console.log('User:', session.user)

    // Store user profile
    const { error: profileError } = await supabase
      .from('profiles')
      .upsert({
        id: session.user.id,
        email: session.user.email,
        avatar_url: session.user.user_metadata.avatar_url,
        full_name: session.user.user_metadata.full_name,
        updated_at: new Date().toISOString()
      })

    if (profileError) throw profileError
  }
}

// Link additional OAuth provider
async function linkGitHubProvider() {
  const { data, error } = await supabase.auth.linkIdentity({
    provider: 'github'
  })

  if (error) throw error
  return data
}
```

---

## 8. Multi-Tenant Architecture

### Example 8.1: Tenant Isolation

**Complete multi-tenant setup**:

```sql
-- Tenant context function
CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS UUID
LANGUAGE sql
STABLE
AS $$
  SELECT COALESCE(
    current_setting('app.current_tenant_id', TRUE)::UUID,
    (SELECT tenant_id FROM user_tenants WHERE user_id = auth.uid() LIMIT 1)
  );
$$;

-- Automatically set tenant_id on insert
CREATE OR REPLACE FUNCTION set_tenant_id()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.tenant_id := current_tenant_id();
  RETURN NEW;
END;
$$;

-- Apply trigger to all tenant tables
CREATE TRIGGER set_tenant_id_trigger
  BEFORE INSERT ON projects
  FOR EACH ROW
  EXECUTE FUNCTION set_tenant_id();
```

**TypeScript middleware**:

```typescript
// Set tenant context for session
async function setTenantContext(tenantId: string) {
  const { error } = await supabase.rpc('set_config', {
    setting: 'app.current_tenant_id',
    value: tenantId
  })

  if (error) throw error
}

// Query with tenant isolation
async function getTenantProjects() {
  // Tenant ID automatically applied via RLS
  const { data, error } = await supabase
    .from('projects')
    .select('*')

  if (error) throw error
  return data
}
```

---

## 9. Real-time Multiplayer

### Example 9.1: Collaborative Editing

**Real-time collaboration pattern**:

```typescript
interface CursorPosition {
  userId: string
  x: number
  y: number
  color: string
}

function setupCollaborativeEditor(documentId: string) {
  const channel = supabase.channel(`document-${documentId}`)

  // Track cursor positions
  const cursors = new Map<string, CursorPosition>()

  // Broadcast cursor movement
  function broadcastCursor(x: number, y: number) {
    channel.send({
      type: 'broadcast',
      event: 'cursor',
      payload: { x, y }
    })
  }

  // Listen to cursor updates
  channel
    .on('broadcast', { event: 'cursor' }, ({ payload }) => {
      cursors.set(payload.userId, payload)
      renderCursors(cursors)
    })
    .on('presence', { event: 'sync' }, () => {
      const state = channel.presenceState()
      updateOnlineUsers(state)
    })
    .subscribe()

  return { channel, broadcastCursor }
}
```

---

## 10. Performance Monitoring

### Example 10.1: Query Performance

**Monitor slow queries**:

```sql
-- Enable pg_stat_statements
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Query performance stats
SELECT
  query,
  calls,
  total_exec_time,
  mean_exec_time,
  max_exec_time
FROM pg_stat_statements
WHERE query NOT LIKE '%pg_stat_statements%'
ORDER BY mean_exec_time DESC
LIMIT 10;
```

**Client-side monitoring**:

```typescript
// Performance wrapper
async function monitoredQuery<T>(
  queryFn: () => Promise<T>,
  queryName: string
): Promise<T> {
  const start = performance.now()

  try {
    const result = await queryFn()
    const duration = performance.now() - start

    console.log(`Query ${queryName}: ${duration}ms`)

    // Send to monitoring service
    if (duration > 1000) {
      console.warn(`Slow query detected: ${queryName} (${duration}ms)`)
    }

    return result
  } catch (error) {
    console.error(`Query ${queryName} failed:`, error)
    throw error
  }
}

// Usage
const projects = await monitoredQuery(
  () => supabase.from('projects').select('*'),
  'fetch_projects'
)
```

---

## Best Practices

**DO**:
- ✅ Enable RLS on all tables
- ✅ Use parameterized queries
- ✅ Implement proper error handling
- ✅ Monitor query performance
- ✅ Use Edge Functions for complex logic
- ✅ Leverage realtime for live updates

**DON'T**:
- ❌ Expose service role key in client
- ❌ Skip RLS policies
- ❌ Store sensitive data unencrypted
- ❌ Ignore query performance
- ❌ Over-fetch data

---

**Context7 Reference**: `/websites/supabase` (latest API v2.38+)
**Last Updated**: 2025-11-22
**Status**: Production Ready
