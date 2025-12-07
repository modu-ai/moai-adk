# Database Platform Reference Documentation

## Supabase Deep Dive

### Context7 Integration
```python
async def get_supabase_docs():
    """Get latest Supabase documentation."""
    docs = await mcp__context7__get_library_docs(
        context7CompatibleLibraryID="/supabase/supabase",
        topic="postgresql-16 pgvector rls edge-functions realtime storage",
        tokens=8000
    )
    return docs
```

### PostgreSQL 16 Advanced Features

JSON Functions and Operators:
```sql
-- JSONB containment and existence
SELECT * FROM products
WHERE metadata @> '{"category": "electronics"}'
  AND metadata ? 'brand';

-- JSONB path queries (PostgreSQL 12+)
SELECT * FROM orders
WHERE jsonb_path_exists(
  items,
  '$.items[*] ? (@.quantity > 5)'
);

-- Aggregate JSON data
SELECT
  user_id,
  jsonb_agg(jsonb_build_object(
    'product', product_name,
    'quantity', quantity
  )) AS order_items
FROM order_items
GROUP BY user_id;
```

pgvector Advanced Operations:
```sql
-- Create embedding with specific dimensions
CREATE TABLE embeddings (
  id SERIAL PRIMARY KEY,
  content TEXT,
  embedding vector(1536),  -- OpenAI ada-002
  metadata JSONB
);

-- IVFFlat index for large datasets (millions of rows)
CREATE INDEX embeddings_ivf_idx ON embeddings
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- HNSW index for faster queries (recommended)
CREATE INDEX embeddings_hnsw_idx ON embeddings
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Hybrid search: vector + full-text
CREATE OR REPLACE FUNCTION hybrid_search(
  query_text TEXT,
  query_embedding vector(1536),
  match_count INT DEFAULT 10,
  full_text_weight FLOAT DEFAULT 0.3,
  semantic_weight FLOAT DEFAULT 0.7
)
RETURNS TABLE (id INT, content TEXT, score FLOAT) AS $$
BEGIN
  RETURN QUERY
  WITH semantic AS (
    SELECT
      embeddings.id,
      embeddings.content,
      1 - (embeddings.embedding <=> query_embedding) AS similarity
    FROM embeddings
    ORDER BY embeddings.embedding <=> query_embedding
    LIMIT match_count * 2
  ),
  full_text AS (
    SELECT
      embeddings.id,
      embeddings.content,
      ts_rank(
        to_tsvector('english', embeddings.content),
        plainto_tsquery('english', query_text)
      ) AS rank
    FROM embeddings
    WHERE to_tsvector('english', embeddings.content) @@
          plainto_tsquery('english', query_text)
    LIMIT match_count * 2
  )
  SELECT
    COALESCE(s.id, f.id) AS id,
    COALESCE(s.content, f.content) AS content,
    (COALESCE(s.similarity, 0) * semantic_weight +
     COALESCE(f.rank, 0) * full_text_weight) AS score
  FROM semantic s
  FULL OUTER JOIN full_text f ON s.id = f.id
  ORDER BY score DESC
  LIMIT match_count;
END;
$$ LANGUAGE plpgsql;
```

### Row-Level Security Patterns

Multi-Tenant with Hierarchical Access:
```sql
-- Enable RLS
ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE tasks ENABLE ROW LEVEL SECURITY;

-- Organization level policy
CREATE POLICY "org_member_access" ON organizations
FOR ALL USING (
  id IN (
    SELECT org_id FROM org_members
    WHERE user_id = auth.uid()
  )
);

-- Project level with role checking
CREATE POLICY "project_access" ON projects
FOR SELECT USING (
  org_id IN (
    SELECT org_id FROM org_members
    WHERE user_id = auth.uid()
  )
);

CREATE POLICY "project_admin_modify" ON projects
FOR ALL USING (
  org_id IN (
    SELECT org_id FROM org_members
    WHERE user_id = auth.uid()
    AND role IN ('admin', 'owner')
  )
);

-- Task level with project membership
CREATE POLICY "task_access" ON tasks
FOR ALL USING (
  project_id IN (
    SELECT p.id FROM projects p
    JOIN org_members om ON p.org_id = om.org_id
    WHERE om.user_id = auth.uid()
  )
  OR assignee_id = auth.uid()
);

-- Bypass RLS for service role
CREATE POLICY "service_role_bypass" ON organizations
FOR ALL TO service_role USING (true);
```

### Real-time Advanced Patterns

Presence and Typing Indicators:
```typescript
// realtime-presence.ts
import { createClient, RealtimeChannel } from '@supabase/supabase-js'

interface PresenceState {
  user_id: string
  email: string
  online_at: string
  typing?: boolean
  cursor_position?: { x: number; y: number }
}

class RealtimePresence {
  private channel: RealtimeChannel
  private supabase: SupabaseClient

  constructor(roomId: string) {
    this.supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY)
    this.channel = this.supabase.channel(`room:${roomId}`, {
      config: { presence: { key: 'user-id' } }
    })
  }

  async join(userInfo: PresenceState) {
    this.channel
      .on('presence', { event: 'sync' }, () => {
        const state = this.channel.presenceState<PresenceState>()
        this.onPresenceSync(state)
      })
      .on('presence', { event: 'join' }, ({ key, newPresences }) => {
        this.onUserJoin(key, newPresences)
      })
      .on('presence', { event: 'leave' }, ({ key, leftPresences }) => {
        this.onUserLeave(key, leftPresences)
      })
      .subscribe(async (status) => {
        if (status === 'SUBSCRIBED') {
          await this.channel.track(userInfo)
        }
      })
  }

  async updateTyping(isTyping: boolean) {
    await this.channel.track({ typing: isTyping })
  }

  async updateCursor(position: { x: number; y: number }) {
    await this.channel.track({ cursor_position: position })
  }

  private onPresenceSync(state: Record<string, PresenceState[]>) {
    console.log('Current users:', Object.keys(state))
  }

  private onUserJoin(key: string, presences: PresenceState[]) {
    console.log('User joined:', key, presences)
  }

  private onUserLeave(key: string, presences: PresenceState[]) {
    console.log('User left:', key, presences)
  }

  leave() {
    this.channel.unsubscribe()
  }
}
```

### Edge Functions Best Practices

Rate Limiting and Authentication:
```typescript
// supabase/functions/api/index.ts
import { serve } from 'https://deno.land/std@0.168.0/http/server.ts'
import { createClient } from 'https://esm.sh/@supabase/supabase-js@2'

const corsHeaders = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type'
}

// Rate limiting using Supabase
async function checkRateLimit(
  supabase: SupabaseClient,
  identifier: string,
  limit: number,
  windowSeconds: number
): Promise<boolean> {
  const windowStart = new Date(Date.now() - windowSeconds * 1000).toISOString()

  const { count } = await supabase
    .from('rate_limits')
    .select('*', { count: 'exact', head: true })
    .eq('identifier', identifier)
    .gte('created_at', windowStart)

  if (count && count >= limit) {
    return false
  }

  await supabase
    .from('rate_limits')
    .insert({ identifier })

  return true
}

serve(async (req) => {
  if (req.method === 'OPTIONS') {
    return new Response('ok', { headers: corsHeaders })
  }

  const supabase = createClient(
    Deno.env.get('SUPABASE_URL')!,
    Deno.env.get('SUPABASE_SERVICE_ROLE_KEY')!
  )

  // Verify JWT
  const authHeader = req.headers.get('authorization')
  if (!authHeader) {
    return new Response(
      JSON.stringify({ error: 'Missing authorization header' }),
      { status: 401, headers: { ...corsHeaders, 'Content-Type': 'application/json' } }
    )
  }

  const { data: { user }, error } = await supabase.auth.getUser(
    authHeader.replace('Bearer ', '')
  )

  if (error || !user) {
    return new Response(
      JSON.stringify({ error: 'Invalid token' }),
      { status: 401, headers: { ...corsHeaders, 'Content-Type': 'application/json' } }
    )
  }

  // Check rate limit
  const allowed = await checkRateLimit(supabase, user.id, 100, 60)
  if (!allowed) {
    return new Response(
      JSON.stringify({ error: 'Rate limit exceeded' }),
      { status: 429, headers: { ...corsHeaders, 'Content-Type': 'application/json' } }
    )
  }

  // Process request
  const body = await req.json()

  return new Response(
    JSON.stringify({ success: true, user_id: user.id }),
    { headers: { ...corsHeaders, 'Content-Type': 'application/json' } }
  )
})
```

### Storage with Transformations

```typescript
// supabase-storage.ts
import { createClient } from '@supabase/supabase-js'

const supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY)

async function uploadImage(file: File, userId: string) {
  const fileName = `${userId}/${Date.now()}-${file.name}`

  const { data, error } = await supabase.storage
    .from('images')
    .upload(fileName, file, {
      cacheControl: '3600',
      upsert: false
    })

  if (error) throw error

  // Get transformed URLs
  const { data: { publicUrl } } = supabase.storage
    .from('images')
    .getPublicUrl(fileName, {
      transform: {
        width: 800,
        height: 600,
        resize: 'contain'
      }
    })

  const { data: { publicUrl: thumbnailUrl } } = supabase.storage
    .from('images')
    .getPublicUrl(fileName, {
      transform: {
        width: 200,
        height: 200,
        resize: 'cover'
      }
    })

  return { originalPath: data.path, publicUrl, thumbnailUrl }
}
```

---

## Neon Deep Dive

### Context7 Integration
```python
async def get_neon_docs():
    """Get latest Neon documentation."""
    docs = await mcp__context7__get_library_docs(
        context7CompatibleLibraryID="/neondatabase/neon",
        topic="serverless branching pitr auto-scaling connection-pooling",
        tokens=6000
    )
    return docs
```

### Branching Strategies

Preview Branch Automation:
```typescript
// neon-github-action.ts
import { NeonClient } from '@neondatabase/api-client'

interface GitHubContext {
  eventName: string
  prNumber?: number
  ref: string
  sha: string
}

class NeonPreviewBranches {
  private neon: NeonClient
  private projectId: string

  constructor(apiKey: string, projectId: string) {
    this.neon = new NeonClient({ apiKey })
    this.projectId = projectId
  }

  async handlePullRequest(context: GitHubContext) {
    if (context.eventName === 'pull_request') {
      return this.createPreviewBranch(context.prNumber!, context.sha)
    } else if (context.eventName === 'pull_request_closed') {
      return this.deletePreviewBranch(context.prNumber!)
    }
  }

  async createPreviewBranch(prNumber: number, sha: string) {
    const branchName = `pr-${prNumber}`

    // Check if branch exists
    const branches = await this.neon.listProjectBranches(this.projectId)
    const existing = branches.branches.find(b => b.name === branchName)

    if (existing) {
      // Reset to latest main
      await this.neon.restoreProjectBranch(this.projectId, existing.id, {
        source_branch_id: 'main'
      })
      return existing
    }

    // Create new branch from main
    const branch = await this.neon.createProjectBranch(this.projectId, {
      branch: {
        name: branchName,
        parent_id: 'main'
      },
      endpoints: [{ type: 'read_write' }]
    })

    return branch
  }

  async deletePreviewBranch(prNumber: number) {
    const branchName = `pr-${prNumber}`
    const branches = await this.neon.listProjectBranches(this.projectId)
    const branch = branches.branches.find(b => b.name === branchName)

    if (branch) {
      await this.neon.deleteProjectBranch(this.projectId, branch.id)
    }
  }

  async getConnectionString(branchName: string): Promise<string> {
    const branches = await this.neon.listProjectBranches(this.projectId)
    const branch = branches.branches.find(b => b.name === branchName)

    if (!branch) throw new Error(`Branch ${branchName} not found`)

    const endpoints = await this.neon.listProjectBranchEndpoints(
      this.projectId,
      branch.id
    )

    return endpoints.endpoints[0].connection_uri
  }
}
```

### Connection Pooling for Edge

```typescript
// neon-edge.ts
import { neon, neonConfig } from '@neondatabase/serverless'

// Configure for edge environments
neonConfig.fetchConnectionCache = true

// Pooled connection for serverless
const sql = neon(process.env.DATABASE_URL!, {
  fetchOptions: {
    cache: 'no-store'
  }
})

// Transaction support
export async function transferFunds(
  fromAccount: string,
  toAccount: string,
  amount: number
) {
  const results = await sql.transaction([
    sql`UPDATE accounts SET balance = balance - ${amount} WHERE id = ${fromAccount}`,
    sql`UPDATE accounts SET balance = balance + ${amount} WHERE id = ${toAccount}`,
    sql`INSERT INTO transfers (from_account, to_account, amount) VALUES (${fromAccount}, ${toAccount}, ${amount})`
  ])

  return results
}

// Batch queries for efficiency
export async function batchInsert(records: any[]) {
  const values = records.map(r => sql`(${r.name}, ${r.email}, ${r.created_at})`)

  await sql`
    INSERT INTO users (name, email, created_at)
    VALUES ${sql.join(values, sql`, `)}
    ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name
  `
}
```

---

## Convex Deep Dive

### Context7 Integration
```python
async def get_convex_docs():
    """Get latest Convex documentation."""
    docs = await mcp__context7__get_library_docs(
        context7CompatibleLibraryID="/get-convex/convex",
        topic="reactive-queries mutations optimistic-updates scheduling actions",
        tokens=6000
    )
    return docs
```

### Advanced Reactive Patterns

Paginated Queries:
```typescript
// convex/documents.ts
import { query, mutation } from './_generated/server'
import { v } from 'convex/values'
import { paginationOptsValidator } from 'convex/server'

export const listPaginated = query({
  args: {
    ownerId: v.string(),
    paginationOpts: paginationOptsValidator
  },
  handler: async (ctx, args) => {
    return await ctx.db
      .query('documents')
      .withIndex('by_owner', (q) => q.eq('ownerId', args.ownerId))
      .order('desc')
      .paginate(args.paginationOpts)
  }
})

// Search with full-text
export const search = query({
  args: { query: v.string(), limit: v.optional(v.number()) },
  handler: async (ctx, args) => {
    return await ctx.db
      .query('documents')
      .withSearchIndex('search_content', (q) =>
        q.search('content', args.query)
      )
      .take(args.limit ?? 10)
  }
})
```

Actions for External APIs:
```typescript
// convex/actions/ai.ts
import { action } from '../_generated/server'
import { v } from 'convex/values'
import { internal } from '../_generated/api'

export const generateEmbedding = action({
  args: { text: v.string(), documentId: v.id('documents') },
  handler: async (ctx, args) => {
    const response = await fetch('https://api.openai.com/v1/embeddings', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${process.env.OPENAI_API_KEY}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        model: 'text-embedding-ada-002',
        input: args.text
      })
    })

    const result = await response.json()
    const embedding = result.data[0].embedding

    // Store embedding using internal mutation
    await ctx.runMutation(internal.documents.storeEmbedding, {
      documentId: args.documentId,
      embedding
    })

    return { success: true }
  }
})
```

Scheduled Jobs:
```typescript
// convex/crons.ts
import { cronJobs } from 'convex/server'
import { internal } from './_generated/api'

const crons = cronJobs()

// Daily cleanup at midnight UTC
crons.daily(
  'cleanup-old-documents',
  { hourUTC: 0, minuteUTC: 0 },
  internal.maintenance.cleanupOldDocuments
)

// Every 5 minutes
crons.interval(
  'sync-external-data',
  { minutes: 5 },
  internal.sync.syncExternalData
)

export default crons
```

---

## Firebase Firestore Deep Dive

### Context7 Integration
```python
async def get_firestore_docs():
    """Get latest Firestore documentation."""
    docs = await mcp__context7__get_library_docs(
        context7CompatibleLibraryID="/firebase/firebase-docs",
        topic="firestore security-rules offline-persistence cloud-functions indexes",
        tokens=6000
    )
    return docs
```

### Advanced Security Rules

Role-Based Access with Custom Claims:
```javascript
// firestore.rules
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Helper functions
    function isSignedIn() {
      return request.auth != null;
    }

    function isAdmin() {
      return request.auth.token.admin == true;
    }

    function hasRole(role) {
      return request.auth.token.roles[role] == true;
    }

    function isOwner(userId) {
      return request.auth.uid == userId;
    }

    // Organizations with team hierarchy
    match /organizations/{orgId} {
      function isMember() {
        return exists(/databases/$(database)/documents/organizations/$(orgId)/members/$(request.auth.uid));
      }

      function getMemberRole() {
        return get(/databases/$(database)/documents/organizations/$(orgId)/members/$(request.auth.uid)).data.role;
      }

      function isOrgAdmin() {
        return isMember() && getMemberRole() in ['admin', 'owner'];
      }

      allow read: if isSignedIn() && isMember();
      allow update: if isOrgAdmin();
      allow delete: if getMemberRole() == 'owner';

      // Nested members collection
      match /members/{memberId} {
        allow read: if isMember();
        allow write: if isOrgAdmin();
      }

      // Projects within organization
      match /projects/{projectId} {
        allow read: if isMember();
        allow create: if isMember() && getMemberRole() in ['admin', 'owner', 'editor'];
        allow update, delete: if isOrgAdmin() ||
          resource.data.createdBy == request.auth.uid;
      }
    }

    // Rate limiting using timestamps
    match /rateLimits/{limitId} {
      allow read: if false;
      allow create: if isSignedIn()
        && request.resource.data.timestamp == request.time
        && request.resource.data.userId == request.auth.uid;
    }
  }
}
```

### Composite Indexes

```json
// firestore.indexes.json
{
  "indexes": [
    {
      "collectionGroup": "documents",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "organizationId", "order": "ASCENDING" },
        { "fieldPath": "createdAt", "order": "DESCENDING" }
      ]
    },
    {
      "collectionGroup": "documents",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "tags", "arrayConfig": "CONTAINS" },
        { "fieldPath": "createdAt", "order": "DESCENDING" }
      ]
    },
    {
      "collectionGroup": "comments",
      "queryScope": "COLLECTION_GROUP",
      "fields": [
        { "fieldPath": "authorId", "order": "ASCENDING" },
        { "fieldPath": "createdAt", "order": "DESCENDING" }
      ]
    }
  ],
  "fieldOverrides": [
    {
      "collectionGroup": "documents",
      "fieldPath": "content",
      "indexes": [
        { "order": "ASCENDING", "queryScope": "COLLECTION" }
      ]
    }
  ]
}
```

### Cloud Functions V2

```typescript
// functions/src/index.ts
import { onDocumentCreated, onDocumentUpdated } from 'firebase-functions/v2/firestore'
import { onCall, HttpsError } from 'firebase-functions/v2/https'
import { onSchedule } from 'firebase-functions/v2/scheduler'
import { getFirestore, FieldValue } from 'firebase-admin/firestore'
import { getMessaging } from 'firebase-admin/messaging'

const db = getFirestore()

// Document trigger with batched writes
export const onDocumentUpdate = onDocumentUpdated(
  {
    document: 'documents/{docId}',
    region: 'us-central1'
  },
  async (event) => {
    const before = event.data?.before.data()
    const after = event.data?.after.data()

    if (!before || !after) return

    // Track changes
    const batch = db.batch()

    // Create change log
    const changeRef = db.collection('changes').doc()
    batch.set(changeRef, {
      documentId: event.params.docId,
      before,
      after,
      changedBy: after.lastModifiedBy,
      changedAt: FieldValue.serverTimestamp()
    })

    // Update document stats
    const statsRef = db.doc(`stats/documents`)
    batch.update(statsRef, {
      totalModifications: FieldValue.increment(1)
    })

    await batch.commit()
  }
)

// Callable function with validation
export const inviteToOrganization = onCall(
  { region: 'us-central1' },
  async (request) => {
    if (!request.auth) {
      throw new HttpsError('unauthenticated', 'Must be signed in')
    }

    const { organizationId, email, role } = request.data

    // Verify caller is org admin
    const memberDoc = await db
      .doc(`organizations/${organizationId}/members/${request.auth.uid}`)
      .get()

    if (!memberDoc.exists || !['admin', 'owner'].includes(memberDoc.data()?.role)) {
      throw new HttpsError('permission-denied', 'Must be organization admin')
    }

    // Create invitation
    const invitation = await db.collection('invitations').add({
      organizationId,
      email,
      role,
      invitedBy: request.auth.uid,
      createdAt: FieldValue.serverTimestamp(),
      status: 'pending'
    })

    return { invitationId: invitation.id }
  }
)

// Scheduled function
export const dailyCleanup = onSchedule(
  {
    schedule: '0 0 * * *',
    timeZone: 'UTC',
    region: 'us-central1'
  },
  async () => {
    const thirtyDaysAgo = new Date()
    thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30)

    const oldDocs = await db
      .collection('tempFiles')
      .where('createdAt', '<', thirtyDaysAgo)
      .limit(500)
      .get()

    const batch = db.batch()
    oldDocs.docs.forEach((doc) => batch.delete(doc.ref))

    await batch.commit()
    console.log(`Deleted ${oldDocs.size} old temp files`)
  }
)
```

---

## Cross-Provider Comparison

### Performance Characteristics

Supabase:
- Read latency: 10-50ms (PostgreSQL performance)
- Write latency: 15-80ms
- Real-time: WebSocket, 100ms propagation
- Scaling: Vertical (compute size) + Read replicas

Neon:
- Read latency: 10-30ms (serverless overhead minimal)
- Write latency: 20-60ms
- Cold start: 500ms-2s (scale to zero)
- Scaling: Automatic, pay-per-use

Convex:
- Read latency: 5-20ms (optimized reactive cache)
- Write latency: 10-40ms
- Real-time: Automatic, sub-100ms
- Scaling: Automatic, serverless

Firestore:
- Read latency: 50-200ms (varies by region)
- Write latency: 100-300ms
- Real-time: WebSocket, 100-500ms
- Scaling: Automatic, unlimited

### Pricing Comparison (as of 2024)

Supabase Free Tier:
- 500MB database
- 1GB storage
- 2GB bandwidth
- 50,000 monthly active users

Neon Free Tier:
- 3GB storage
- 1 project, unlimited branches
- 100 compute hours/month

Convex Free Tier:
- 500MB storage
- 1M function calls
- 1M bandwidth

Firestore Free Tier:
- 1GB storage
- 50,000 reads/day
- 20,000 writes/day

---

Last Updated: 2025-12-07
Context7 Mappings: All 4 providers mapped
Status: Production Ready
