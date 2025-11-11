---
title: "Supabase 완벽 가이드"
description: "Enterprise PostgreSQL BaaS - RLS, Realtime, Edge Functions으로 30분 내 Production 배포"
---

# Supabase 완벽 가이드

> **Open-source PostgreSQL BaaS**: Row Level Security, Real-time Subscriptions, Edge Functions로 엔터프라이즈급 Multi-tenant SaaS를 30분 내 구축

## 1. Supabase란?

### 개요
```yaml
supabase:
  tagline: "The Open Source Firebase Alternative"
  category: "PostgreSQL-based BaaS"
  license: "Apache 2.0 (Open Source)"

  core_stack:
    database: "PostgreSQL 16+"
    api: "PostgREST (Auto-generated REST API)"
    auth: "GoTrue (JWT-based)"
    realtime: "Realtime Server (WebSocket)"
    storage: "S3-compatible Object Storage"
    functions: "Deno Edge Functions"

  managed_by: "Supabase Inc."
  pricing: "Free tier → Pro ($25/mo) → Team → Enterprise"
```

### 왜 Supabase인가?

**전통적인 Firebase 문제점**:
```yaml
firebase_limitations:
  database: "NoSQL only (Firestore)"
  queries: "제한적 (복잡한 JOIN 불가)"
  vendor_lock_in: "Google Cloud 종속"
  sql: "SQL 사용 불가"
  self_hosting: "불가능"
```

**Supabase 솔루션**:
```yaml
supabase_advantages:
  database: "PostgreSQL 16 with extensions"
  queries: "Full SQL power (JOIN, subqueries, CTEs)"
  vendor_freedom: "Self-hosting 가능"
  sql: "Native SQL support"
  open_source: "Apache 2.0 license"
  migration: "Easy import from other databases"
```

## 2. 핵심 기능

### 2.1 PostgreSQL Database

#### PostgreSQL 16 with Extensions
```sql
-- pgvector: Vector similarity search (AI embeddings)
CREATE EXTENSION vector;

CREATE TABLE documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  content TEXT,
  embedding vector(1536)  -- OpenAI embeddings
);

-- Cosine similarity search
SELECT * FROM documents
ORDER BY embedding <=> $1
LIMIT 10;

-- pg_cron: Scheduled jobs
SELECT cron.schedule(
  'cleanup-old-data',
  '0 2 * * *',  -- Daily at 2am
  $$DELETE FROM logs WHERE created_at < NOW() - INTERVAL '30 days'$$
);

-- PostGIS: Geospatial queries
CREATE EXTENSION postgis;

SELECT * FROM locations
WHERE ST_DWithin(
  geography(geom),
  'POINT(-122.4194 37.7749)'::geography,
  5000  -- 5km radius
);
```

#### Row Level Security (RLS)
```sql
-- Multi-tenant SaaS 핵심 기능

-- 1. 테이블 생성
CREATE TABLE projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  name TEXT NOT NULL,
  created_by UUID NOT NULL
);

-- 2. RLS 활성화
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

-- 3. 정책 설정: 사용자는 자신의 테넌트 데이터만 조회
CREATE POLICY tenant_isolation ON projects
  FOR ALL
  USING (tenant_id = (SELECT tenant_id FROM users WHERE id = auth.uid()));

-- 4. 관리자는 모든 데이터 조회
CREATE POLICY admin_all ON projects
  FOR ALL
  USING (
    EXISTS (
      SELECT 1 FROM user_roles
      WHERE user_id = auth.uid()
      AND role = 'admin'
    )
  );
```

**RLS 자동 적용**:
```typescript
// 클라이언트 코드는 RLS 정책을 신경 쓸 필요 없음
const { data: projects } = await supabase
  .from('projects')
  .select('*')
// → 자동으로 현재 사용자의 tenant_id 데이터만 반환
```

### 2.2 Realtime Subscriptions

#### Database Changes 구독
```typescript
// 1. 테이블 변경사항 실시간 구독
const channel = supabase
  .channel('todos-channel')
  .on(
    'postgres_changes',
    {
      event: '*',  // INSERT, UPDATE, DELETE
      schema: 'public',
      table: 'todos'
    },
    (payload) => {
      console.log('Change received!', payload)
      // payload.eventType: 'INSERT' | 'UPDATE' | 'DELETE'
      // payload.new: 새 데이터
      // payload.old: 이전 데이터
    }
  )
  .subscribe()

// 2. 특정 필터 적용
const userChannel = supabase
  .channel('user-todos')
  .on(
    'postgres_changes',
    {
      event: 'INSERT',
      schema: 'public',
      table: 'todos',
      filter: `user_id=eq.${userId}`  // 특정 사용자만
    },
    (payload) => {
      console.log('New todo for user:', payload.new)
    }
  )
  .subscribe()
```

#### Presence (사용자 현황)
```typescript
// 실시간 협업 도구용
const channel = supabase.channel('room:123')

// 현재 사용자 상태 전송
await channel.track({
  online_at: new Date().toISOString(),
  user_id: userId,
  cursor: { x: 100, y: 200 }
})

// 다른 사용자 상태 수신
channel.on('presence', { event: 'sync' }, () => {
  const state = channel.presenceState()
  // { 'user1': { online_at, cursor }, 'user2': ... }
})

// 사용자 입장/퇴장 이벤트
channel.on('presence', { event: 'join' }, ({ key, newPresences }) => {
  console.log(`${key} joined`)
})
```

#### Broadcast (메시지 전송)
```typescript
// 클라이언트 간 직접 메시지 전송 (DB 거치지 않음)
const channel = supabase.channel('chat:room123')

// 메시지 전송
await channel.send({
  type: 'broadcast',
  event: 'message',
  payload: { text: 'Hello!' }
})

// 메시지 수신
channel.on('broadcast', { event: 'message' }, (payload) => {
  console.log('Message:', payload.text)
})
```

### 2.3 Authentication

#### Email/Password
```typescript
// 회원가입
const { data, error } = await supabase.auth.signUp({
  email: 'user@example.com',
  password: 'password',
  options: {
    data: {
      name: 'John Doe',
      avatar_url: 'https://...'
    }
  }
})

// 로그인
const { data, error } = await supabase.auth.signInWithPassword({
  email: 'user@example.com',
  password: 'password'
})

// 현재 세션 확인
const { data: { session } } = await supabase.auth.getSession()
```

#### OAuth Providers (50+)
```typescript
// Google 로그인
await supabase.auth.signInWithOAuth({
  provider: 'google',
  options: {
    redirectTo: 'https://yourapp.com/auth/callback'
  }
})

// GitHub, Apple, Facebook, Discord 등 50+ providers
```

#### Magic Link (비밀번호 없는 로그인)
```typescript
await supabase.auth.signInWithOtp({
  email: 'user@example.com',
  options: {
    emailRedirectTo: 'https://yourapp.com/welcome'
  }
})
```

#### MFA (Multi-Factor Authentication)
```typescript
// MFA 등록
const { data, error } = await supabase.auth.mfa.enroll({
  factorType: 'totp'
})

// QR 코드 표시
const qrCode = data.totp.qr_code

// MFA 검증
await supabase.auth.mfa.verify({
  factorId: data.id,
  code: '123456'  // User's TOTP code
})
```

### 2.4 Edge Functions

#### Deno Runtime
```typescript
// supabase/functions/hello/index.ts

import { serve } from 'https://deno.land/std@0.168.0/http/server.ts'
import { createClient } from 'https://esm.sh/@supabase/supabase-js@2'

serve(async (req) => {
  // 1. CORS 헤더 자동 처리
  if (req.method === 'OPTIONS') {
    return new Response('ok', { headers: corsHeaders })
  }

  // 2. JWT 자동 검증
  const authHeader = req.headers.get('Authorization')!
  const supabase = createClient(
    Deno.env.get('SUPABASE_URL')!,
    Deno.env.get('SUPABASE_ANON_KEY')!,
    { global: { headers: { Authorization: authHeader } } }
  )

  // 3. 데이터베이스 접근 (RLS 자동 적용)
  const { data: user } = await supabase.auth.getUser()

  const { data } = await supabase
    .from('todos')
    .select('*')

  // 4. 응답 반환
  return new Response(
    JSON.stringify({ user, todos: data }),
    { headers: { 'Content-Type': 'application/json' } }
  )
})
```

#### 배포
```bash
# 로컬 개발
supabase functions serve hello

# Production 배포
supabase functions deploy hello

# 호출
curl -X POST https://project-ref.supabase.co/functions/v1/hello \
  -H "Authorization: Bearer YOUR_ANON_KEY"
```

### 2.5 Storage

#### File Upload
```typescript
// 파일 업로드
const { data, error } = await supabase.storage
  .from('avatars')
  .upload('user123/avatar.png', file, {
    cacheControl: '3600',
    upsert: false
  })

// Public URL 생성
const { data: { publicUrl } } = supabase.storage
  .from('avatars')
  .getPublicUrl('user123/avatar.png')

// Signed URL (만료 시간 지정)
const { data: { signedUrl } } = await supabase.storage
  .from('private-files')
  .createSignedUrl('document.pdf', 60 * 60)  // 1 hour
```

#### 이미지 변환
```typescript
// 자동 이미지 리사이즈/변환
const { data: { publicUrl } } = supabase.storage
  .from('images')
  .getPublicUrl('photo.jpg', {
    transform: {
      width: 800,
      height: 600,
      resize: 'cover',
      format: 'webp',
      quality: 80
    }
  })
```

#### RLS로 Storage 보호
```sql
-- Storage 객체에도 RLS 적용 가능
CREATE POLICY user_avatar ON storage.objects
  FOR ALL
  USING (
    bucket_id = 'avatars' AND
    auth.uid()::text = (storage.foldername(name))[1]
  );
```

## 3. Architecture Patterns

### Pattern A: Multi-tenant SaaS

#### 스키마 설계
```sql
-- 테넌트 테이블
CREATE TABLE tenants (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 사용자 테이블 (테넌트 연결)
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID REFERENCES tenants(id),
  email TEXT UNIQUE NOT NULL,
  role TEXT NOT NULL DEFAULT 'member'
);

-- 비즈니스 데이터 (테넌트별 격리)
CREATE TABLE projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id),
  name TEXT NOT NULL,
  created_by UUID NOT NULL REFERENCES users(id)
);

-- RLS 정책
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON projects
  FOR ALL
  USING (
    tenant_id = (
      SELECT tenant_id FROM users
      WHERE id = auth.uid()
    )
  );
```

#### 클라이언트 코드
```typescript
// Middleware: JWT에 tenant_id 추가
export async function middleware(req: NextRequest) {
  const supabase = createMiddlewareClient({ req, res })

  const { data: { user } } = await supabase.auth.getUser()

  if (user) {
    // 현재 사용자의 tenant_id 조회
    const { data: userData } = await supabase
      .from('users')
      .select('tenant_id')
      .eq('id', user.id)
      .single()

    // JWT에 tenant_id 주입
    await supabase.auth.updateUser({
      data: { tenant_id: userData.tenant_id }
    })
  }
}

// 페이지: RLS 자동 적용으로 안전한 데이터 조회
export default async function ProjectsPage() {
  const supabase = createServerClient()

  const { data: projects } = await supabase
    .from('projects')
    .select('*')
  // → 자동으로 현재 테넌트 데이터만 반환

  return <ProjectsList projects={projects} />
}
```

### Pattern D: Real-time Collaboration

#### 스키마
```sql
-- 공유 문서
CREATE TABLE documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title TEXT NOT NULL,
  content JSONB,  -- Y.js CRDT state
  version INTEGER DEFAULT 0
);

-- 커서 위치 추적
CREATE TABLE cursor_positions (
  user_id UUID,
  document_id UUID,
  position JSONB,
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (user_id, document_id)
);
```

#### 실시간 협업 구현
```typescript
import { useEffect, useState } from 'react'
import { createClient } from '@supabase/supabase-js'

function CollaborativeEditor({ documentId }) {
  const [users, setUsers] = useState([])
  const supabase = createClient(...)

  useEffect(() => {
    // 1. Presence: 현재 온라인 사용자
    const channel = supabase.channel(`doc:${documentId}`)

    channel
      .on('presence', { event: 'sync' }, () => {
        const state = channel.presenceState()
        setUsers(Object.values(state).flat())
      })
      .subscribe(async (status) => {
        if (status === 'SUBSCRIBED') {
          // 현재 사용자 등록
          await channel.track({
            user_id: userId,
            name: userName,
            cursor: null
          })
        }
      })

    // 2. Broadcast: 커서 이동 공유
    const handleCursorMove = (position) => {
      channel.send({
        type: 'broadcast',
        event: 'cursor',
        payload: { user_id: userId, position }
      })
    }

    channel.on('broadcast', { event: 'cursor' }, ({ payload }) => {
      updateCursor(payload.user_id, payload.position)
    })

    // 3. Database Changes: 문서 변경 감지
    const docChannel = supabase
      .channel('document-changes')
      .on(
        'postgres_changes',
        {
          event: 'UPDATE',
          schema: 'public',
          table: 'documents',
          filter: `id=eq.${documentId}`
        },
        (payload) => {
          if (payload.new.version > localVersion) {
            // 새 버전 적용
            applyRemoteChanges(payload.new.content)
          }
        }
      )
      .subscribe()

    return () => {
      channel.unsubscribe()
      docChannel.unsubscribe()
    }
  }, [documentId])

  return (
    <div>
      <OnlineUsers users={users} />
      <Editor />
    </div>
  )
}
```

## 4. Performance Optimization

### Database Optimization

#### Indexes
```sql
-- 1. B-tree index (일반 조회)
CREATE INDEX idx_projects_tenant ON projects(tenant_id);

-- 2. Partial index (조건부)
CREATE INDEX idx_active_projects ON projects(tenant_id)
WHERE status = 'active';

-- 3. Covering index (Index-only scan)
CREATE INDEX idx_projects_cover ON projects(tenant_id, name, created_at);

-- 4. GiST index (Full-text search)
CREATE INDEX idx_projects_search ON projects
USING GiST (to_tsvector('english', name || ' ' || description));

SELECT * FROM projects
WHERE to_tsvector('english', name || ' ' || description) @@ to_tsquery('postgres');
```

#### Connection Pooling
```typescript
// Supabase는 자동으로 PgBouncer 사용
// 최대 10,000+ 동시 연결 지원

const supabase = createClient(
  process.env.SUPABASE_URL!,
  process.env.SUPABASE_ANON_KEY!,
  {
    db: {
      schema: 'public'
    },
    auth: {
      autoRefreshToken: true,
      persistSession: true
    }
  }
)
```

#### Query Optimization
```typescript
// 1. Select specific columns (not *)
const { data } = await supabase
  .from('projects')
  .select('id, name, created_at')  // Good
  // .select('*')  // Bad: 불필요한 데이터 포함

// 2. Use pagination
const { data } = await supabase
  .from('projects')
  .select('*')
  .range(0, 9)  // First 10 items
  .order('created_at', { ascending: false })

// 3. Limit joined data
const { data } = await supabase
  .from('projects')
  .select(`
    *,
    tasks!inner(*)
  `)
  .limit(10)  // Projects
  .limit(5, { foreignTable: 'tasks' })  // Tasks per project
```

### Realtime Optimization

#### Channels 최적화
```typescript
// Bad: 모든 변경사항 구독
const channel = supabase
  .channel('all-changes')
  .on('postgres_changes', { event: '*', schema: 'public', table: 'todos' }, ...)

// Good: 필터 사용
const channel = supabase
  .channel('my-todos')
  .on(
    'postgres_changes',
    {
      event: 'INSERT',
      schema: 'public',
      table: 'todos',
      filter: `user_id=eq.${userId}`  // 현재 사용자만
    },
    ...
  )
```

## 5. Security Best Practices

### RLS 패턴
```sql
-- 1. 읽기 전용 접근
CREATE POLICY public_read ON posts
  FOR SELECT
  USING (published = true);

-- 2. 소유자만 수정
CREATE POLICY owner_update ON posts
  FOR UPDATE
  USING (auth.uid() = user_id)
  WITH CHECK (auth.uid() = user_id);

-- 3. Role-based 접근
CREATE POLICY admin_all ON sensitive_data
  FOR ALL
  USING (
    EXISTS (
      SELECT 1 FROM user_roles
      WHERE user_id = auth.uid()
      AND role IN ('admin', 'superadmin')
    )
  );

-- 4. Time-based 접근
CREATE POLICY scheduled_access ON scheduled_content
  FOR SELECT
  USING (
    publish_at <= NOW() AND
    (expire_at IS NULL OR expire_at > NOW())
  );
```

### API Key 관리
```typescript
// .env.local
NEXT_PUBLIC_SUPABASE_URL=https://xxx.supabase.co
NEXT_PUBLIC_SUPABASE_ANON_KEY=eyJhb...  # Public (frontend)
SUPABASE_SERVICE_ROLE_KEY=eyJhb...  # Secret (backend only)

// Frontend: anon key 사용 (RLS 적용)
const supabase = createBrowserClient(
  process.env.NEXT_PUBLIC_SUPABASE_URL!,
  process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!
)

// Backend: service role key (RLS 우회 가능 - 주의!)
const supabaseAdmin = createClient(
  process.env.SUPABASE_URL!,
  process.env.SUPABASE_SERVICE_ROLE_KEY!
)
```

## 6. 배포 및 운영

### Local Development
```bash
# Supabase CLI 설치
npm install -g supabase

# 프로젝트 초기화
supabase init

# 로컬 Supabase 실행 (Docker)
supabase start
# → PostgreSQL, GoTrue, PostgREST, Realtime, Storage 모두 실행

# Database 마이그레이션
supabase db reset
```

### Database Migrations
```bash
# 새 마이그레이션 생성
supabase migration new add_projects_table

# supabase/migrations/20250111000000_add_projects_table.sql
CREATE TABLE projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL
);

# 마이그레이션 적용
supabase db push
```

### Production Deployment
```bash
# Supabase 프로젝트 연결
supabase link --project-ref your-project-ref

# Database 변경사항 배포
supabase db push

# Edge Functions 배포
supabase functions deploy hello

# 환경 변수 설정
supabase secrets set STRIPE_API_KEY=sk_live_...
```

### Monitoring
```typescript
// Supabase Dashboard:
// - Real-time queries/sec
// - Connection pool usage
// - Database size
// - Bandwidth usage
// - Error logs

// Custom monitoring
const { data: metrics } = await supabase
  .from('pg_stat_statements')
  .select('*')
  .order('total_exec_time', { ascending: false })
  .limit(10)
// → 가장 느린 쿼리 Top 10
```

## 7. 비용 최적화

### 무료 티어 (Free)
```yaml
free_tier:
  database: "500MB"
  bandwidth: "5GB/month"
  storage: "1GB"
  auth_users: "Unlimited"
  edge_functions: "500K invocations/month"
  projects: "2"

  ideal_for: "Hobby projects, MVPs"
```

### Pro 티어 ($25/month)
```yaml
pro_tier:
  database: "8GB (+ $0.125/GB)"
  bandwidth: "50GB (+ $0.09/GB)"
  storage: "100GB (+ $0.021/GB)"
  auth_users: "100,000 MAU"
  edge_functions: "2M invocations/month"
  projects: "Unlimited"

  extras:
    - "7-day Point-in-time recovery"
    - "1-day database backups"
    - "Email support"
```

### 비용 절감 팁
```typescript
// 1. Selective column 조회
const { data } = await supabase
  .from('large_table')
  .select('id, name')  // 필요한 컬럼만

// 2. Pagination으로 대량 데이터 방지
const { data } = await supabase
  .from('posts')
  .select('*')
  .range(0, 49)  // 50개씩

// 3. Edge Caching
const { data, error } = await supabase
  .from('static_content')
  .select('*')
  .eq('id', contentId)
// → CDN 캐싱으로 DB 조회 최소화

// 4. Database Cleanup
-- 오래된 데이터 자동 삭제
SELECT cron.schedule(
  'cleanup',
  '0 0 * * *',
  $$DELETE FROM logs WHERE created_at < NOW() - INTERVAL '30 days'$$
);
```

## 8. 마이그레이션 가이드

### From Firebase
```typescript
// 1. Firestore → PostgreSQL schema 변환
// Firestore collection:
{
  users: {
    user123: {
      name: "John",
      email: "john@example.com",
      posts: {
        post1: { title: "Hello" }
      }
    }
  }
}

// PostgreSQL schema:
CREATE TABLE users (
  id UUID PRIMARY KEY,
  name TEXT,
  email TEXT
);

CREATE TABLE posts (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  title TEXT
);

// 2. 데이터 마이그레이션
const firestore = getFirestore()
const users = await firestore.collection('users').get()

for (const doc of users.docs) {
  const userData = doc.data()

  await supabase.from('users').insert({
    id: doc.id,
    name: userData.name,
    email: userData.email
  })
}
```

### From Traditional Backend
```bash
# 1. 기존 PostgreSQL 데이터 덤프
pg_dump -h old-host -U user -d mydb > backup.sql

# 2. Supabase에 복원
psql -h db.xxx.supabase.co -U postgres -d postgres < backup.sql

# 3. RLS 정책 추가
ALTER TABLE existing_table ENABLE ROW LEVEL SECURITY;
```

## 9. 다음 단계

- [BaaS Ecosystem 개요](../baas-ecosystem) - 9개 플랫폼 비교
- [Firebase 가이드](./firebase) - NoSQL 기반 BaaS
- [Vercel 가이드](./vercel) - Edge Platform
- [Advanced Skills](../advanced-skills) - Context7, MCP Builder
