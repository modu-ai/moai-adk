# Skill: moai-baas-supabase-ext

## 메타데이터

```yaml
skill_id: moai-baas-supabase-ext
skill_name: Supabase 심화 가이드 (RLS, Migrations, Realtime)
version: 1.0.0
created_date: 2025-11-09
language: korean
triggers:
  - keywords: ["Supabase", "RLS", "Row Level Security", "PostgreSQL", "마이그레이션", "Realtime"]
  - contexts: ["supabase-detected", "pattern-a", "pattern-d"]
agents:
  - backend-expert
  - database-expert
  - security-expert
freedom_level: high
word_count: 1000
context7_references:
  - url: "https://supabase.com/docs/guides/database/postgres/row-level-security"
    topic: "RLS 정책 작성"
  - url: "https://supabase.com/docs/guides/database/migrations"
    topic: "마이그레이션 안전성"
  - url: "https://supabase.com/docs/guides/realtime"
    topic: "Realtime 구독"
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

---

## 📚 내용

### 1. Supabase 아키텍처 (150 words)

**Supabase**는 PostgreSQL 기반의 오픈소스 Firebase 대체제입니다.

**핵심 구성요소**:
```
┌─────────────────────────────────┐
│ Supabase                        │
├─────────────────────────────────┤
│ 1. PostgreSQL Database          │
│    └─ Tables, Functions, Triggers
│                                  │
│ 2. Authentication               │
│    └─ Email, Magic Link, OAuth  │
│                                  │
│ 3. Row Level Security (RLS)     │
│    └─ Policy-based access       │
│                                  │
│ 4. Real-time Subscriptions      │
│    └─ Broadcast, Postgres Changes
│                                  │
│ 5. Storage                       │
│    └─ File buckets, CDN         │
│                                  │
│ 6. Edge Functions               │
│    └─ Serverless PostgreSQL Funcs
└─────────────────────────────────┘
```

**Edge Functions vs Database Functions**:

| 기능 | Edge Functions | Database Functions |
|-----|---|---|
| 언어 | TypeScript/JavaScript | PL/pgSQL, Python |
| 실행 위치 | 엣지 (고속) | 데이터베이스 내부 |
| 사용 시기 | HTTP 요청 응답 | 데이터 변경 트리거 |
| 성능 | 매우 빠름 | 제한적 |

---

### 2. RLS (Row Level Security) 심화 (300 words)

**RLS란**: 사용자의 역할과 정책에 따라 행 단위로 데이터 접근을 제어하는 PostgreSQL 기능.

**기본 개념**:
```sql
-- Example: users 테이블
-- Rule: 자기 자신의 데이터만 조회 가능

ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view their own data"
ON users FOR SELECT
USING (auth.uid() = id);

CREATE POLICY "Users can update their own data"
ON users FOR UPDATE
USING (auth.uid() = id)
WITH CHECK (auth.uid() = id);
```

**Policy 작성 패턴**:

**Pattern 1: 자신의 데이터만 (Most Common)**
```sql
CREATE POLICY "Self access"
ON profiles FOR ALL
USING (auth.uid() = user_id);
```

**Pattern 2: 역할 기반 (Role-based)**
```sql
CREATE POLICY "Admin or owner can delete"
ON posts FOR DELETE
USING (
  auth.uid() = user_id
  OR auth.jwt()->>'role' = 'admin'
);
```

**Pattern 3: 공유 데이터 (Shared)**
```sql
CREATE POLICY "Shared with me"
ON documents FOR SELECT
USING (
  user_id = auth.uid()
  OR shared_with @> jsonb_build_array(auth.uid()::text)
);
```

**500 에러 디버깅**:

```
현상: "new row violates row-level security policy"
원인: 쓰기 작업 후 SELECT 정책 확인 부족

해결:
1. Supabase 대시보드 → SQL Editor
2. 로그 확인: SELECT * FROM auth.logs
3. Policy 검증:
   SELECT * FROM pg_policies WHERE schemaname='public';
```

**Policy 테스트 (pgTAP)**:

```sql
-- pgTAP을 사용한 정책 검증
CREATE OR REPLACE FUNCTION test_rls()
RETURNS void AS $$
DECLARE
  user_id uuid := 'xxx';
BEGIN
  -- User는 자신의 데이터만 보임
  ASSERT (
    SELECT COUNT(*) FROM profiles
    WHERE user_id = auth.uid()
  ) = 1;
END;
$$ LANGUAGE plpgsql;
```

**보안 Best Practices**:
- ✅ 모든 테이블에 RLS 활성화
- ✅ 각 테이블마다 SELECT, INSERT, UPDATE, DELETE 정책 정의
- ✅ auth.uid()를 항상 포함 (인증 확인)
- ✅ JWT claims 검증 (`auth.jwt()->>'role'`)
- ❌ 서비스 역할(Service Role) 토큰 노출 금지

---

### 3. Database Functions (200 words)

**Database Functions**: PostgreSQL 함수를 RPC(Remote Procedure Call)로 노출.

**사용 시나리오**:
- 복잡한 비즈니스 로직
- 원자성 보장 필요
- 다중 테이블 변경

**예제: 트윗 생성 (좋아요 카운트 업데이트)**

```sql
CREATE OR REPLACE FUNCTION create_tweet(
  p_content TEXT,
  p_user_id UUID
)
RETURNS tweets AS $$
DECLARE
  v_tweet tweets;
BEGIN
  -- 트윗 삽입
  INSERT INTO tweets (content, user_id, created_at)
  VALUES (p_content, p_user_id, NOW())
  RETURNING * INTO v_tweet;

  -- 사용자의 트윗 카운트 증가 (한 번의 트랜잭션)
  UPDATE users
  SET tweet_count = tweet_count + 1
  WHERE id = p_user_id;

  RETURN v_tweet;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

**클라이언트에서 호출**:
```typescript
const { data, error } = await supabase.rpc('create_tweet', {
  p_content: 'Hello World',
  p_user_id: userId
});
```

**Triggers**: 자동 실행되는 함수

```sql
CREATE OR REPLACE FUNCTION update_user_stats()
RETURNS TRIGGER AS $$
BEGIN
  -- 새로운 트윗이 생성될 때마다
  UPDATE users
  SET tweet_count = tweet_count + 1
  WHERE id = NEW.user_id;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_tweet_created
AFTER INSERT ON tweets
FOR EACH ROW
EXECUTE FUNCTION update_user_stats();
```

---

### 4. Migrations (200 words)

**마이그레이션**: 데이터베이스 스키마의 버전 관리.

**전략 1: Migration-first (추천)**

```bash
# 1. 마이그레이션 생성
supabase migration new add_user_table

# 2. SQL 작성
cat supabase/migrations/20250101120000_add_user_table.sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email TEXT UNIQUE NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

# 3. 로컬에서 테스트
supabase db reset

# 4. 프로덕션에 배포
supabase db push
```

**전략 2: Dashboard-first (피해야 함)**

```
Supabase 대시보드에서 직접 테이블 생성
→ 마이그레이션 파일이 없음
→ 다른 개발자와 동기화 불가
→ 프로덕션 배포 불가능
```

**안전한 마이그레이션**:

```sql
-- ❌ 위험: 데이터 손실 가능
ALTER TABLE users DROP COLUMN email;

-- ✅ 안전: 단계적 변경
-- Step 1: 새 컬럼 추가
ALTER TABLE users ADD COLUMN email_new TEXT;

-- Step 2: 데이터 마이그레이션
UPDATE users SET email_new = email;

-- Step 3: 기존 컬럼 제거 (다음 배포)
ALTER TABLE users DROP COLUMN email;
```

**Rollback 전략**:
```sql
-- 이전 마이그레이션으로 되돌리기
supabase db push --version 20250101110000
```

---

### 5. Realtime (100 words)

**Realtime**: WebSocket을 통한 실시간 데이터 동기화.

**두 가지 모드**:

**Mode 1: Broadcast** (메시지 전송)
```typescript
// 사용자 1: 메시지 브로드캐스트
supabase.realtime.channel('game').send({
  type: 'broadcast',
  event: 'player_moved',
  payload: { x: 100, y: 200 }
});

// 사용자 2: 메시지 수신
channel.on('broadcast', { event: 'player_moved' }, (payload) => {
  console.log('Player moved:', payload);
});
```

**Mode 2: Postgres Changes** (DB 변경 감지)
```typescript
supabase
  .channel('public:messages')
  .on(
    'postgres_changes',
    { event: 'INSERT', schema: 'public', table: 'messages' },
    (payload) => {
      console.log('New message:', payload.new);
    }
  )
  .subscribe();
```

**성능**: 1000+ 동시 연결 지원, RLS 자동 적용.

---

### 6. Common Issues & Solutions (50 words)

| 문제 | 원인 | 해결 |
|-----|------|-----|
| Auth 토큰 만료 | 1시간 유효기간 | Refresh token 사용 |
| RLS 500 에러 | 정책 누락 | `INSERT INTO` 후 `SELECT` 정책 확인 |
| 느린 쿼리 | 인덱스 미생성 | `CREATE INDEX` 추가 |
| Realtime 연결 안됨 | Replication 비활성화 | 대시보드에서 활성화 |

---

## 🎯 사용 방법

### Agent에서 호출

```python
# database-expert, security-expert에서
Skill("moai-baas-supabase-ext")

# Supabase 패턴 감지 시 자동 로드
```

### Context7 자동 로딩

Supabase 감지 시 다음 문서 자동 로드:
- RLS 정책 작성 가이드
- 마이그레이션 베스트 프랙티스
- Realtime 구독 방법

---

## 📚 참고 자료

- [Supabase RLS 공식 문서](https://supabase.com/docs/guides/database/postgres/row-level-security)
- [마이그레이션 가이드](https://supabase.com/docs/guides/database/migrations)
- [Realtime](https://supabase.com/docs/guides/realtime)

---

## ✅ 검증

- [x] 아키텍처 설명
- [x] RLS 심화 가이드
- [x] Database Functions
- [x] Migrations
- [x] Realtime
- [x] 1000 단어 목표
