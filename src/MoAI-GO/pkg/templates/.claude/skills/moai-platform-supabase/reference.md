---
name: supabase-reference
description: API reference summary for Supabase platform
parent-skill: moai-platform-supabase
version: 1.0.0
updated: 2026-01-06
---

# Supabase API Reference

## Client Methods

### Database Operations

```typescript
// Select
const { data, error } = await supabase.from('table').select('*')
const { data } = await supabase.from('table').select('id, name, relation(*)').eq('id', 1)

// Insert
const { data, error } = await supabase.from('table').insert({ column: 'value' }).select()
const { data } = await supabase.from('table').insert([...items]).select()

// Update
const { data, error } = await supabase.from('table').update({ column: 'value' }).eq('id', 1).select()

// Upsert
const { data, error } = await supabase.from('table').upsert({ id: 1, column: 'value' }).select()

// Delete
const { error } = await supabase.from('table').delete().eq('id', 1)
```

### Query Filters

```typescript
.eq('column', 'value')         // Equal
.neq('column', 'value')        // Not equal
.gt('column', 0)               // Greater than
.gte('column', 0)              // Greater than or equal
.lt('column', 100)             // Less than
.lte('column', 100)            // Less than or equal
.like('column', '%pattern%')   // LIKE
.ilike('column', '%pattern%')  // ILIKE (case insensitive)
.is('column', null)            // IS NULL
.in('column', ['a', 'b'])      // IN
.contains('array_col', ['a'])  // Array contains
.containedBy('col', ['a','b']) // Array contained by
.range('col', '[1,10)')        // Range
.textSearch('col', 'query')    // Full-text search
.filter('col', 'op', 'val')    // Generic filter
```

### Query Modifiers

```typescript
.order('column', { ascending: false })
.limit(10)
.range(0, 9)              // Pagination
.single()                 // Expect exactly one row
.maybeSingle()           // Expect zero or one row
.count('exact', { head: true })  // Count only
```

### RPC (Remote Procedure Call)

```typescript
const { data, error } = await supabase.rpc('function_name', {
  arg1: 'value1',
  arg2: 'value2'
})
```

## Auth Methods

```typescript
// Sign up
await supabase.auth.signUp({ email, password })

// Sign in
await supabase.auth.signInWithPassword({ email, password })
await supabase.auth.signInWithOAuth({ provider: 'google' })
await supabase.auth.signInWithOtp({ email })

// Session
await supabase.auth.getUser()
await supabase.auth.getSession()
await supabase.auth.refreshSession()

// Sign out
await supabase.auth.signOut()

// Password
await supabase.auth.resetPasswordForEmail(email)
await supabase.auth.updateUser({ password: newPassword })

// Listener
supabase.auth.onAuthStateChange((event, session) => {})
```

## Real-time Methods

```typescript
// Subscribe to changes
const channel = supabase.channel('channel-name')
  .on('postgres_changes',
    { event: '*', schema: 'public', table: 'table_name' },
    (payload) => {}
  )
  .subscribe()

// Presence
channel.on('presence', { event: 'sync' }, () => {
  const state = channel.presenceState()
})
await channel.track({ user_id: 'id', online_at: new Date() })

// Broadcast
await channel.send({ type: 'broadcast', event: 'name', payload: {} })
channel.on('broadcast', { event: 'name' }, ({ payload }) => {})

// Unsubscribe
await supabase.removeChannel(channel)
await supabase.removeAllChannels()
```

## Storage Methods

```typescript
// Upload
await supabase.storage.from('bucket').upload('path/file.ext', file, { cacheControl: '3600' })

// Download
await supabase.storage.from('bucket').download('path/file.ext')

// Get URL
supabase.storage.from('bucket').getPublicUrl('path/file.ext', {
  transform: { width: 800, height: 600, resize: 'cover' }
})

// Signed URL
await supabase.storage.from('bucket').createSignedUrl('path/file.ext', 3600)

// List
await supabase.storage.from('bucket').list('folder', { limit: 100 })

// Delete
await supabase.storage.from('bucket').remove(['path/file.ext'])

// Move
await supabase.storage.from('bucket').move('old/path', 'new/path')
```

## Edge Functions

```typescript
// Invoke
const { data, error } = await supabase.functions.invoke('function-name', {
  body: { key: 'value' },
  headers: { 'Custom-Header': 'value' }
})
```

## SQL Quick Reference

### pgvector

```sql
CREATE EXTENSION vector;

-- Create table with vector column
CREATE TABLE items (
  id UUID PRIMARY KEY,
  embedding vector(1536)
);

-- HNSW index (recommended)
CREATE INDEX ON items USING hnsw (embedding vector_cosine_ops);

-- IVFFlat index (large datasets)
CREATE INDEX ON items USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Distance operators
<->  -- Euclidean distance
<=>  -- Cosine distance
<#>  -- Negative inner product
```

### Row-Level Security

```sql
-- Enable RLS
ALTER TABLE table_name ENABLE ROW LEVEL SECURITY;

-- Create policy
CREATE POLICY "policy_name" ON table_name
  FOR SELECT | INSERT | UPDATE | DELETE | ALL
  TO role_name
  USING (expression)
  WITH CHECK (expression);

-- Auth functions
auth.uid()                    -- Current user ID
auth.jwt() ->> 'claim'        -- JWT claim value
auth.role()                   -- Current role
```

### Useful Functions

```sql
gen_random_uuid()             -- Generate UUID
uuid_generate_v4()            -- Generate UUID (requires uuid-ossp)
NOW()                         -- Current timestamp
CURRENT_TIMESTAMP            -- Current timestamp
to_tsvector('english', text)  -- Full-text search vector
plainto_tsquery('query')      -- Full-text search query
ts_rank(vector, query)        -- Full-text search rank
```

## Environment Variables

```bash
# Public (safe for client)
NEXT_PUBLIC_SUPABASE_URL=https://xxx.supabase.co
NEXT_PUBLIC_SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Private (server-only)
SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
SUPABASE_JWT_SECRET=your-jwt-secret
```

## CLI Commands

```bash
# Project
supabase init
supabase start
supabase stop
supabase status

# Database
supabase db diff
supabase db push
supabase db reset
supabase migration new migration_name
supabase migration list

# Types
supabase gen types typescript --project-id xxx > database.types.ts

# Functions
supabase functions new function-name
supabase functions serve function-name
supabase functions deploy function-name
supabase functions list

# Secrets
supabase secrets set KEY=value
supabase secrets list
```

## Context7 Documentation Access

For detailed API documentation, use Context7 MCP tools:

```
Step 1: Resolve library ID
mcp__context7__resolve-library-id with query "supabase"

Step 2: Fetch documentation
mcp__context7__get-library-docs with:
- context7CompatibleLibraryID: resolved ID
- topic: "specific topic"
- tokens: 5000-10000
```

Common topics:
- "javascript client select insert update"
- "auth signIn signUp oauth"
- "realtime postgres_changes presence"
- "storage upload download transform"
- "edge-functions deploy invoke"
- "row-level-security policies"
- "pgvector similarity search"
