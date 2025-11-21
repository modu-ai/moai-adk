# Supabase Advanced Patterns

**Enterprise-grade patterns for complex Supabase architectures.**

**Version**: 1.0.0
**Last Updated**: 2025-11-22
**Research Base**: Context7 `/websites/supabase`, production patterns

---

## 📚 Table of Contents

1. [Multi-Tenant SaaS Architecture](#1-multi-tenant-saas-architecture)
2. [Advanced RLS Patterns](#2-advanced-rls-patterns)
3. [Realtime Broadcast Channels](#3-realtime-broadcast-channels)
4. [Session Management](#4-session-management)
5. [Vector Search Integration](#5-vector-search-integration)
6. [Webhook Integration](#6-webhook-integration)
7. [Real-time Presence Tracking](#7-real-time-presence-tracking)
8. [CI/CD Deployment](#8-cicd-deployment)

---

## 1. Multi-Tenant SaaS Architecture

### Pattern 1.1: Tenant Isolation with RLS

**Complete multi-tenant architecture**:

```sql
-- Tenant management tables
CREATE TABLE tenants (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  subscription_tier TEXT DEFAULT 'free',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE tenant_members (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
  user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'viewer')),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(tenant_id, user_id)
);

-- Helper function: Get current user's tenants
CREATE OR REPLACE FUNCTION user_tenants()
RETURNS SETOF UUID
LANGUAGE sql
STABLE
AS $$
  SELECT tenant_id FROM tenant_members WHERE user_id = auth.uid();
$$;

-- Helper function: Check tenant access
CREATE OR REPLACE FUNCTION has_tenant_access(tenant_id UUID)
RETURNS BOOLEAN
LANGUAGE sql
STABLE
AS $$
  SELECT EXISTS (
    SELECT 1 FROM tenant_members
    WHERE tenant_members.tenant_id = has_tenant_access.tenant_id
    AND user_id = auth.uid()
  );
$$;

-- Helper function: Check role
CREATE OR REPLACE FUNCTION has_role(tenant_id UUID, required_role TEXT)
RETURNS BOOLEAN
LANGUAGE sql
STABLE
AS $$
  SELECT EXISTS (
    SELECT 1 FROM tenant_members
    WHERE tenant_members.tenant_id = has_role.tenant_id
    AND user_id = auth.uid()
    AND role = required_role
  );
$$;

-- Tenant data table with RLS
CREATE TABLE tenant_data (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  data JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE tenant_data ENABLE ROW LEVEL SECURITY;

-- RLS policies
CREATE POLICY "Users can view tenant data"
  ON tenant_data FOR SELECT
  USING (has_tenant_access(tenant_id));

CREATE POLICY "Members can insert tenant data"
  ON tenant_data FOR INSERT
  WITH CHECK (
    has_tenant_access(tenant_id)
    AND has_role(tenant_id, 'member')
  );

CREATE POLICY "Admins can update tenant data"
  ON tenant_data FOR UPDATE
  USING (has_role(tenant_id, 'admin'))
  WITH CHECK (has_role(tenant_id, 'admin'));

CREATE POLICY "Owners can delete tenant data"
  ON tenant_data FOR DELETE
  USING (has_role(tenant_id, 'owner'));
```

**TypeScript client**:

```typescript
// Tenant context manager
class TenantContext {
  private currentTenantId: string | null = null

  async setCurrentTenant(tenantId: string) {
    const { data, error } = await supabase
      .from('tenant_members')
      .select('tenant_id')
      .eq('tenant_id', tenantId)
      .eq('user_id', (await supabase.auth.getUser()).data.user?.id)
      .single()

    if (error || !data) {
      throw new Error('Unauthorized tenant access')
    }

    this.currentTenantId = tenantId
  }

  getCurrentTenantId(): string {
    if (!this.currentTenantId) {
      throw new Error('No tenant context set')
    }
    return this.currentTenantId
  }

  async getTenantData() {
    const tenantId = this.getCurrentTenantId()

    const { data, error } = await supabase
      .from('tenant_data')
      .select('*')
      .eq('tenant_id', tenantId)

    if (error) throw error
    return data
  }
}

const tenantContext = new TenantContext()
```

---

## 2. Advanced RLS Patterns

### Pattern 2.1: Dynamic Column-Level Security

**Column-level access control**:

```sql
-- User permissions table
CREATE TABLE user_permissions (
  user_id UUID REFERENCES auth.users(id),
  table_name TEXT,
  column_name TEXT,
  can_read BOOLEAN DEFAULT FALSE,
  can_write BOOLEAN DEFAULT FALSE,
  PRIMARY KEY (user_id, table_name, column_name)
);

-- Secure view with column filtering
CREATE OR REPLACE FUNCTION create_secure_view(table_name TEXT)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
  allowed_columns TEXT;
BEGIN
  -- Get columns user has read access to
  SELECT string_agg(column_name, ', ')
  INTO allowed_columns
  FROM user_permissions
  WHERE user_permissions.table_name = create_secure_view.table_name
  AND user_id = auth.uid()
  AND can_read = TRUE;

  -- Create dynamic view
  EXECUTE format(
    'CREATE OR REPLACE VIEW %I_secure AS SELECT %s FROM %I',
    table_name,
    allowed_columns,
    table_name
  );
END;
$$;

-- Usage example
SELECT create_secure_view('sensitive_data');
```

### Pattern 2.2: Time-Based Access Control

**Temporal RLS policies**:

```sql
-- Access schedules
CREATE TABLE access_schedules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES auth.users(id),
  resource_type TEXT,
  resource_id UUID,
  valid_from TIMESTAMPTZ NOT NULL,
  valid_until TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Time-based RLS policy
CREATE POLICY "Time-based access"
  ON protected_resources FOR SELECT
  USING (
    id IN (
      SELECT resource_id FROM access_schedules
      WHERE user_id = auth.uid()
      AND resource_type = 'protected_resource'
      AND NOW() BETWEEN valid_from AND valid_until
    )
  );
```

---

## 3. Realtime Broadcast Channels

### Pattern 3.1: Custom Realtime Events

**Advanced broadcast patterns**:

```typescript
// Event bus for realtime communication
class RealtimeEventBus {
  private channels = new Map<string, RealtimeChannel>()

  async createChannel(channelName: string) {
    const channel = supabase.channel(channelName, {
      config: {
        broadcast: { self: true, ack: true },
        presence: { key: 'user_id' }
      }
    })

    this.channels.set(channelName, channel)
    await channel.subscribe()

    return channel
  }

  async broadcast(channelName: string, event: string, payload: any) {
    const channel = this.channels.get(channelName)
    if (!channel) {
      throw new Error(`Channel ${channelName} not found`)
    }

    const { error } = await channel.send({
      type: 'broadcast',
      event,
      payload
    })

    if (error) throw error
  }

  on(channelName: string, event: string, callback: (payload: any) => void) {
    const channel = this.channels.get(channelName)
    if (!channel) {
      throw new Error(`Channel ${channelName} not found`)
    }

    channel.on('broadcast', { event }, ({ payload }) => {
      callback(payload)
    })
  }

  async cleanup() {
    for (const [name, channel] of this.channels) {
      await channel.unsubscribe()
    }
    this.channels.clear()
  }
}

// Usage
const eventBus = new RealtimeEventBus()

await eventBus.createChannel('game-room-123')

eventBus.on('game-room-123', 'player-move', (payload) => {
  console.log('Player moved:', payload)
})

await eventBus.broadcast('game-room-123', 'player-move', {
  playerId: 'user-456',
  x: 100,
  y: 200
})
```

---

## 4. Session Management

### Pattern 4.1: Advanced Session Handling

**Custom session management**:

```typescript
// Session manager with refresh logic
class SessionManager {
  private refreshTimer: NodeJS.Timeout | null = null

  async initialize() {
    // Get current session
    const { data: { session } } = await supabase.auth.getSession()

    if (session) {
      this.setupAutoRefresh(session.expires_at!)
    }

    // Listen to auth changes
    supabase.auth.onAuthStateChange((event, session) => {
      if (event === 'SIGNED_IN' && session) {
        this.setupAutoRefresh(session.expires_at!)
      } else if (event === 'SIGNED_OUT') {
        this.cleanup()
      }
    })
  }

  private setupAutoRefresh(expiresAt: number) {
    // Clear existing timer
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer)
    }

    // Refresh 5 minutes before expiry
    const refreshTime = (expiresAt * 1000) - Date.now() - (5 * 60 * 1000)

    this.refreshTimer = setTimeout(async () => {
      const { error } = await supabase.auth.refreshSession()
      if (error) {
        console.error('Session refresh failed:', error)
        // Redirect to login
        window.location.href = '/login'
      }
    }, refreshTime)
  }

  private cleanup() {
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer)
      this.refreshTimer = null
    }
  }
}

const sessionManager = new SessionManager()
sessionManager.initialize()
```

---

## 5. Vector Search Integration

### Pattern 5.1: pgvector for Semantic Search

**Vector embeddings with Supabase**:

```sql
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create embeddings table
CREATE TABLE document_embeddings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  document_id UUID NOT NULL,
  content TEXT NOT NULL,
  embedding vector(1536),  -- OpenAI ada-002 dimension
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create index for fast similarity search
CREATE INDEX ON document_embeddings
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- Similarity search function
CREATE OR REPLACE FUNCTION match_documents(
  query_embedding vector(1536),
  match_threshold FLOAT,
  match_count INT
)
RETURNS TABLE (
  id UUID,
  document_id UUID,
  content TEXT,
  similarity FLOAT
)
LANGUAGE sql
STABLE
AS $$
  SELECT
    id,
    document_id,
    content,
    1 - (embedding <=> query_embedding) AS similarity
  FROM document_embeddings
  WHERE 1 - (embedding <=> query_embedding) > match_threshold
  ORDER BY embedding <=> query_embedding
  LIMIT match_count;
$$;
```

**TypeScript client**:

```typescript
import { OpenAI } from 'openai'

const openai = new OpenAI({ apiKey: process.env.OPENAI_API_KEY })

async function semanticSearch(query: string, threshold = 0.7, limit = 10) {
  // Generate embedding for query
  const embeddingResponse = await openai.embeddings.create({
    model: 'text-embedding-ada-002',
    input: query
  })

  const queryEmbedding = embeddingResponse.data[0].embedding

  // Search similar documents
  const { data, error } = await supabase.rpc('match_documents', {
    query_embedding: queryEmbedding,
    match_threshold: threshold,
    match_count: limit
  })

  if (error) throw error
  return data
}
```

---

## 6. Webhook Integration

### Pattern 6.1: Database Webhooks

**Event-driven webhooks**:

```sql
-- Webhook configuration
CREATE TABLE webhooks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  url TEXT NOT NULL,
  events TEXT[] NOT NULL,
  secret TEXT NOT NULL,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Trigger function for webhooks
CREATE OR REPLACE FUNCTION notify_webhook()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
  webhook_record RECORD;
  payload JSONB;
BEGIN
  -- Build payload
  payload := jsonb_build_object(
    'event', TG_OP,
    'table', TG_TABLE_NAME,
    'old', to_jsonb(OLD),
    'new', to_jsonb(NEW)
  );

  -- Notify all matching webhooks
  FOR webhook_record IN
    SELECT * FROM webhooks
    WHERE is_active = TRUE
    AND TG_OP = ANY(events)
  LOOP
    PERFORM http_post(
      webhook_record.url,
      payload::TEXT,
      'application/json',
      ARRAY[
        http_header('X-Webhook-Signature', encode(hmac(payload::TEXT, webhook_record.secret, 'sha256'), 'hex'))
      ]
    );
  END LOOP;

  RETURN NEW;
END;
$$;

-- Apply trigger
CREATE TRIGGER webhook_trigger
  AFTER INSERT OR UPDATE OR DELETE ON important_table
  FOR EACH ROW
  EXECUTE FUNCTION notify_webhook();
```

---

## 7. Real-time Presence Tracking

### Pattern 7.1: Collaborative Presence

**Multi-user presence system**:

```typescript
class PresenceTracker {
  private channel: RealtimeChannel
  private users = new Map<string, any>()

  constructor(roomId: string) {
    this.channel = supabase.channel(`presence-${roomId}`, {
      config: {
        presence: {
          key: 'user_id'
        }
      }
    })

    this.setupPresenceHandlers()
  }

  private setupPresenceHandlers() {
    this.channel
      .on('presence', { event: 'sync' }, () => {
        const state = this.channel.presenceState()
        this.users.clear()

        for (const [userId, presences] of Object.entries(state)) {
          this.users.set(userId, presences[0])
        }

        this.onUsersChange(Array.from(this.users.values()))
      })
      .on('presence', { event: 'join' }, ({ key, newPresences }) => {
        console.log('User joined:', key)
        this.onUserJoin(newPresences[0])
      })
      .on('presence', { event: 'leave' }, ({ key, leftPresences }) => {
        console.log('User left:', key)
        this.onUserLeave(leftPresences[0])
      })
  }

  async join(userData: any) {
    await this.channel.subscribe(async (status) => {
      if (status === 'SUBSCRIBED') {
        await this.channel.track({
          ...userData,
          online_at: new Date().toISOString()
        })
      }
    })
  }

  async updatePresence(updates: any) {
    await this.channel.track(updates)
  }

  async leave() {
    await this.channel.untrack()
    await this.channel.unsubscribe()
  }

  // Override these methods
  protected onUsersChange(users: any[]) {}
  protected onUserJoin(user: any) {}
  protected onUserLeave(user: any) {}
}

// Usage
const presence = new PresenceTracker('room-123')
presence.onUsersChange = (users) => {
  console.log('Online users:', users)
}

await presence.join({
  user_id: 'user-456',
  name: 'John Doe',
  avatar: 'https://...'
})
```

---

## 8. CI/CD Deployment

### Pattern 8.1: GitHub Actions Integration

**Automated migration deployment**:

```yaml
# .github/workflows/supabase-deploy.yml
name: Supabase Deploy

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  SUPABASE_ACCESS_TOKEN: ${{ secrets.SUPABASE_ACCESS_TOKEN }}
  SUPABASE_PROJECT_ID: ${{ secrets.SUPABASE_PROJECT_ID }}

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Supabase CLI
        uses: supabase/setup-cli@v1
        with:
          version: latest

      - name: Link Supabase project
        run: |
          supabase link --project-ref $SUPABASE_PROJECT_ID

      - name: Run migrations
        run: |
          supabase db push

      - name: Deploy Edge Functions
        run: |
          supabase functions deploy --project-ref $SUPABASE_PROJECT_ID

      - name: Run tests
        run: |
          supabase test db
```

**Migration testing**:

```sql
-- supabase/tests/migration_test.sql
BEGIN;

-- Test RLS policies
SELECT plan(3);

-- Test 1: User can view own data
SELECT lives_ok(
  $$
    SET LOCAL role TO authenticated;
    SET LOCAL request.jwt.claim.sub TO 'user-123';
    SELECT * FROM tenant_data WHERE tenant_id = 'tenant-456';
  $$,
  'User can view tenant data'
);

-- Test 2: User cannot view other tenant data
SELECT throws_ok(
  $$
    SET LOCAL role TO authenticated;
    SET LOCAL request.jwt.claim.sub TO 'user-789';
    SELECT * FROM tenant_data WHERE tenant_id = 'tenant-456';
  $$,
  'User cannot view other tenant data'
);

-- Test 3: RLS is enabled
SELECT is(
  (SELECT relrowsecurity FROM pg_class WHERE relname = 'tenant_data'),
  TRUE,
  'RLS is enabled on tenant_data'
);

SELECT * FROM finish();
ROLLBACK;
```

---

## Best Practices

**DO**:
- ✅ Use RLS for tenant isolation
- ✅ Implement proper session management
- ✅ Test RLS policies thoroughly
- ✅ Use vector search for semantic queries
- ✅ Monitor webhook delivery
- ✅ Implement presence tracking for collaboration

**DON'T**:
- ❌ Skip RLS testing
- ❌ Hardcode tenant IDs
- ❌ Ignore session refresh
- ❌ Over-complicate RLS policies
- ❌ Miss webhook signature verification

---

**Context7 Reference**: `/websites/supabase` (latest API v2.38+)
**Last Updated**: 2025-11-22
**Status**: Production Ready
