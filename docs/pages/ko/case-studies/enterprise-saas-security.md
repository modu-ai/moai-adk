---
title: "사례 연구: Enterprise SaaS 보안 구현"
description: "SOC 2 Type 2 준수, Multi-tenant 아키텍처, Zero-trust 보안 모델"
---

# 사례 연구: Enterprise SaaS 보안 구현

## Executive Summary

**프로젝트**: B2B SaaS 플랫폼 엔터프라이즈급 보안 시스템 구축
**기간**: 3개월 (보안 강화 및 규정 준수)
**팀 규모**: 8명 (백엔드 3명, DevOps 2명, 보안 전문가 1명, QA 2명)
**기술 스택**: Node.js, Supabase, Auth0, PostgreSQL, AWS

### 핵심 성과

- ✅ **SOC 2 Type 2 준수** (첫 감사 통과)
- ✅ **Multi-tenant 완벽 격리** (100% RLS 커버리지)
- ✅ **Zero-trust 보안 모델** 구현
- ✅ **보안 사고 제로** (6개월 무사고 운영)
- ✅ **엔터프라이즈 고객 300% 증가**
- ✅ **성능 영향 < 5ms** (보안 기능 추가 후)

---

## 📋 프로젝트 배경

### 비즈니스 상황

**DataFlow**는 데이터 분석 SaaS 플랫폼으로, 중소기업 고객을 대상으로 빠르게 성장했습니다. 그러나 엔터프라이즈 고객 확보를 위해서는 엄격한 보안 요구사항을 충족해야 했습니다.

**기존 시스템의 한계**:
- 기본적인 비밀번호 인증만 존재
- 조직 간 데이터 격리가 애플리케이션 레벨에만 의존
- 감사 로그 시스템 부재
- 보안 규정 준수 불가 (SOC 2, GDPR, HIPAA)

### 엔터프라이즈 고객 요구사항

**Fortune 500 기업의 보안 체크리스트**:

| 요구사항 | 현재 상태 | 필요 조치 |
|---------|----------|----------|
| **SSO/SAML** | ❌ 없음 | Auth0 Enterprise 통합 |
| **MFA** | ❌ 없음 | 필수 적용 |
| **데이터 격리** | ⚠️ App 레벨 | Database RLS |
| **감사 로그** | ❌ 없음 | 완전한 추적 시스템 |
| **암호화** | ⚠️ 부분적 | 저장/전송 모두 |
| **SOC 2** | ❌ 없음 | Type 2 인증 필요 |
| **침투 테스트** | ❌ 없음 | 연 2회 필수 |
| **재해 복구** | ⚠️ 백업만 | RPO/RTO 정의 |

### 도전 과제

**1. 기술적 도전**:
- 기존 시스템 중단 없이 보안 강화
- 성능 영향 최소화 (< 10ms 추가 레이턴시)
- 레거시 코드와 신규 보안 시스템 통합

**2. 비즈니스 도전**:
- 빠른 시장 진입 (3개월 내 SOC 2 감사)
- 제한된 보안 전문 인력 (1명)
- 기존 고객 서비스 지속

---

## 💡 솔루션: MoAI-ADK Security Framework

### 왜 MoAI-ADK를 선택했는가?

**1. security-expert 에이전트**
- OWASP Top 10 자동 체크
- 보안 Best Practices 제안
- 취약점 자동 탐지

**2. Senior Engineer Thinking**
- Auth0, Okta 등 Identity Provider 연구
- Row Level Security 패턴 분석
- Zero-trust 아키텍처 설계

**3. SPEC-First 보안 요구사항**
- 명확한 보안 정책 문서화
- 감사 추적 가능성
- 규정 준수 증거 자료

**4. 자동화된 보안 테스트**
- TDD로 보안 정책 검증
- 침투 테스트 시나리오 자동화
- 지속적 보안 모니터링

---

## 🚀 구현 과정

### Phase 1: 인증 강화 (4주)

#### SPEC-SEC-001: Enterprise Authentication

```markdown
# SPEC-SEC-001: Enterprise Authentication

@TAG:SPEC-SEC-001

## 요구사항 (EARS 형식)

**UBIQUITOUS**:
- 시스템은 SAML 2.0 기반 SSO를 지원해야 한다
- 시스템은 Multi-Factor Authentication(MFA)을 강제해야 한다

**EVENT-DRIVEN**:
- WHEN 사용자가 로그인을 시도하면
- THEN 시스템은 조직의 Identity Provider로 리디렉션해야 한다

**STATE-DRIVEN**:
- WHILE 사용자 세션이 활성화된 동안
- THEN 시스템은 JWT 토큰을 검증하고 갱신해야 한다

**UNWANTED BEHAVIOR**:
- IF MFA를 완료하지 않은 사용자가 접근하려고 하면
- THEN 시스템은 즉시 MFA 설정 페이지로 리디렉션해야 한다

## 인수 기준

1. ✅ SAML SSO 지원 (Okta, Azure AD, Google Workspace)
2. ✅ MFA 필수 (TOTP, SMS, Hardware Token)
3. ✅ Session 관리 (30분 idle timeout)
4. ✅ JWT 토큰 검증 및 갱신
5. ✅ 로그인 이력 기록 (IP, 디바이스, 시간)

## 기술 제약사항

- Auth0 Enterprise Plan
- JWT 기반 토큰 (RS256 알고리즘)
- Refresh Token 순환 (Rotation)
- 보안 헤더 (HSTS, CSP, X-Frame-Options)
```

#### Auth0 Enterprise 통합

```typescript
// @TAG:CODE-SEC-001:AUTH
// lib/auth/auth0-config.ts

import { Auth0Client } from '@auth0/auth0-spa-js'

export const auth0Config = {
  domain: process.env.AUTH0_DOMAIN!,
  clientId: process.env.AUTH0_CLIENT_ID!,

  // Enterprise 기능
  enterprise: {
    // SAML SSO 지원
    connections: [
      'google-workspace',
      'azure-ad',
      'okta',
      'onelogin'
    ],

    // MFA 필수
    mfa: {
      required: true,
      methods: ['totp', 'sms', 'recovery-code'],
      allowRememberBrowser: false
    },

    // Session 정책
    session: {
      idleTimeout: 1800, // 30분
      absoluteTimeout: 43200, // 12시간
      rolling: true
    }
  },

  // JWT 설정
  jwt: {
    algorithm: 'RS256',
    expiresIn: '1h',
    issuer: `https://${process.env.AUTH0_DOMAIN}/`,
    audience: process.env.AUTH0_AUDIENCE
  },

  // 보안 옵션
  security: {
    // Refresh Token Rotation
    useRefreshTokens: true,
    rotateRefreshTokens: true,

    // PKCE (Proof Key for Code Exchange)
    usePKCE: true,

    // 로그인 이력
    trackLoginHistory: true
  }
}

/**
 * Auth0 클라이언트 초기화
 * @TAG:SEC-001
 */
export const auth0 = new Auth0Client({
  domain: auth0Config.domain,
  clientId: auth0Config.clientId,
  authorizationParams: {
    audience: auth0Config.jwt.audience,
    redirect_uri: window.location.origin
  },
  useRefreshTokens: auth0Config.security.useRefreshTokens,
  cacheLocation: 'memory' // XSS 방어
})
```

#### 테스트: 인증 플로우

```typescript
// @TAG:TEST-SEC-001
// tests/auth/enterprise-auth.test.ts

import { describe, it, expect, beforeEach } from 'vitest'
import { auth0, loginWithSSO, verifyMFA } from '@/lib/auth'

describe('SEC-001: Enterprise Authentication', () => {
  describe('SAML SSO 로그인', () => {
    it('Google Workspace SSO가 정상 작동한다', async () => {
      // Given
      const connection = 'google-workspace'
      const email = 'user@company.com'

      // When
      const result = await loginWithSSO(connection, email)

      // Then
      expect(result.success).toBe(true)
      expect(result.requiresMFA).toBe(true)
      expect(result.user.email).toBe(email)
    })

    it('지원하지 않는 도메인은 거부된다', async () => {
      // Given
      const connection = 'google-workspace'
      const email = 'user@gmail.com' // 개인 이메일

      // When & Then
      await expect(
        loginWithSSO(connection, email)
      ).rejects.toThrow('허용되지 않은 도메인입니다')
    })
  })

  describe('MFA 검증', () => {
    it('TOTP 코드가 정확하면 인증된다', async () => {
      // Given
      const userId = 'user-123'
      const totpCode = '123456'

      // When
      const result = await verifyMFA(userId, totpCode, 'totp')

      // Then
      expect(result.verified).toBe(true)
      expect(result.accessToken).toBeDefined()
    })

    it('잘못된 TOTP 코드는 거부된다', async () => {
      // Given
      const userId = 'user-123'
      const invalidCode = '999999'

      // When & Then
      await expect(
        verifyMFA(userId, invalidCode, 'totp')
      ).rejects.toThrow('잘못된 인증 코드입니다')
    })

    it('MFA 없이는 보호된 리소스에 접근할 수 없다', async () => {
      // Given
      const accessTokenWithoutMFA = 'token-without-mfa'

      // When & Then
      await expect(
        fetchProtectedResource(accessTokenWithoutMFA)
      ).rejects.toThrow('MFA 인증이 필요합니다')
    })
  })

  describe('Session 관리', () => {
    it('30분 idle 후 자동 로그아웃된다', async () => {
      // Given
      const session = await createTestSession()

      // When
      await wait(31 * 60 * 1000) // 31분 대기

      // Then
      await expect(
        verifySession(session.id)
      ).rejects.toThrow('세션이 만료되었습니다')
    })

    it('JWT 토큰이 자동 갱신된다', async () => {
      // Given
      const initialToken = await getAccessToken()
      await wait(55 * 60 * 1000) // 55분 대기 (만료 5분 전)

      // When
      const refreshedToken = await getAccessToken()

      // Then
      expect(refreshedToken).not.toBe(initialToken)
      expect(decodeJWT(refreshedToken).exp).toBeGreaterThan(
        decodeJWT(initialToken).exp
      )
    })
  })
})
```

---

### Phase 2: Multi-tenant Row Level Security (6주)

#### SPEC-SEC-002: Data Isolation

```markdown
# SPEC-SEC-002: Multi-tenant Data Isolation

@TAG:SPEC-SEC-002

## 요구사항

**UBIQUITOUS**:
- 시스템은 조직 간 데이터를 완벽히 격리해야 한다
- 모든 데이터베이스 쿼리는 Row Level Security를 통과해야 한다

**STATE-DRIVEN**:
- WHILE 사용자가 조직 A에 속해 있을 때
- THEN 시스템은 조직 A의 데이터만 반환해야 한다

**UNWANTED BEHAVIOR**:
- IF 악의적 사용자가 다른 조직의 tenant_id로 쿼리하려고 하면
- THEN 시스템은 데이터베이스 레벨에서 차단해야 한다

## 인수 기준

1. ✅ 모든 테이블에 tenant_id 컬럼 존재
2. ✅ 100% RLS 정책 적용
3. ✅ Application 레벨 필터 제거 (DB 레벨만)
4. ✅ 성능 영향 < 5ms
5. ✅ Cross-tenant 쿼리 불가능 증명

## 기술 제약사항

- PostgreSQL Row Level Security
- JWT 토큰에 tenant_id 포함
- 인덱스 최적화 (tenant_id 컬럼)
```

#### Database Schema with RLS

```sql
-- @TAG:CODE-SEC-002:DB
-- supabase/migrations/002_rls.sql

-- 1. 모든 테이블에 tenant_id 추가
ALTER TABLE documents ADD COLUMN tenant_id UUID NOT NULL;
ALTER TABLE projects ADD COLUMN tenant_id UUID NOT NULL;
ALTER TABLE users ADD COLUMN tenant_id UUID NOT NULL;
ALTER TABLE analytics ADD COLUMN tenant_id UUID NOT NULL;

-- 2. 인덱스 생성 (성능 최적화)
CREATE INDEX idx_documents_tenant_id ON documents(tenant_id);
CREATE INDEX idx_projects_tenant_id ON projects(tenant_id);
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_analytics_tenant_id ON analytics(tenant_id);

-- 3. Row Level Security 활성화
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE analytics ENABLE ROW LEVEL SECURITY;

-- 4. RLS 정책 정의

-- Documents 테이블 정책
CREATE POLICY "tenant_isolation_documents_select"
  ON documents FOR SELECT
  USING (tenant_id = auth.jwt() ->> 'tenant_id'::UUID);

CREATE POLICY "tenant_isolation_documents_insert"
  ON documents FOR INSERT
  WITH CHECK (tenant_id = auth.jwt() ->> 'tenant_id'::UUID);

CREATE POLICY "tenant_isolation_documents_update"
  ON documents FOR UPDATE
  USING (tenant_id = auth.jwt() ->> 'tenant_id'::UUID)
  WITH CHECK (tenant_id = auth.jwt() ->> 'tenant_id'::UUID);

CREATE POLICY "tenant_isolation_documents_delete"
  ON documents FOR DELETE
  USING (tenant_id = auth.jwt() ->> 'tenant_id'::UUID);

-- Projects 테이블 정책
CREATE POLICY "tenant_isolation_projects"
  ON projects FOR ALL
  USING (tenant_id = auth.jwt() ->> 'tenant_id'::UUID)
  WITH CHECK (tenant_id = auth.jwt() ->> 'tenant_id'::UUID);

-- Users 테이블 정책 (추가 권한 체크)
CREATE POLICY "tenant_isolation_users_select"
  ON users FOR SELECT
  USING (
    tenant_id = auth.jwt() ->> 'tenant_id'::UUID
    OR auth.jwt() ->> 'role' = 'super_admin'
  );

CREATE POLICY "tenant_isolation_users_update"
  ON users FOR UPDATE
  USING (
    tenant_id = auth.jwt() ->> 'tenant_id'::UUID
    AND (
      id = auth.uid()  -- 자기 자신
      OR auth.jwt() ->> 'role' = 'org_admin'  -- 조직 관리자
    )
  );

-- Analytics 테이블 정책
CREATE POLICY "tenant_isolation_analytics"
  ON analytics FOR ALL
  USING (tenant_id = auth.jwt() ->> 'tenant_id'::UUID);

-- 5. 함수: Cross-tenant 쿼리 감지 및 차단
CREATE OR REPLACE FUNCTION enforce_tenant_isolation()
RETURNS TRIGGER AS $$
BEGIN
  -- 새 레코드의 tenant_id가 현재 사용자의 tenant_id와 다른 경우
  IF NEW.tenant_id::TEXT != auth.jwt() ->> 'tenant_id' THEN
    RAISE EXCEPTION 'Cross-tenant access denied: % != %',
      NEW.tenant_id,
      auth.jwt() ->> 'tenant_id';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 6. 트리거: 모든 INSERT/UPDATE에 적용
CREATE TRIGGER enforce_tenant_isolation_documents
  BEFORE INSERT OR UPDATE ON documents
  FOR EACH ROW
  EXECUTE FUNCTION enforce_tenant_isolation();

CREATE TRIGGER enforce_tenant_isolation_projects
  BEFORE INSERT OR UPDATE ON projects
  FOR EACH ROW
  EXECUTE FUNCTION enforce_tenant_isolation();
```

#### 테스트: RLS 검증

```typescript
// @TAG:TEST-SEC-002
// tests/security/rls.test.ts

import { describe, it, expect } from 'vitest'
import { supabase } from '@/lib/supabase'

describe('SEC-002: Row Level Security', () => {
  describe('데이터 격리', () => {
    it('사용자는 자신의 조직 데이터만 조회할 수 있다', async () => {
      // Given
      const orgAUser = await createTestUser({ orgId: 'org-a' })
      const orgADoc = await createTestDocument({ orgId: 'org-a' })
      const orgBDoc = await createTestDocument({ orgId: 'org-b' })

      // When: Org A 사용자로 로그인
      await supabase.auth.signInWithPassword({
        email: orgAUser.email,
        password: 'test123'
      })

      const { data } = await supabase
        .from('documents')
        .select('*')

      // Then: Org A 문서만 반환됨
      expect(data).toHaveLength(1)
      expect(data[0].id).toBe(orgADoc.id)
      expect(data.find(d => d.id === orgBDoc.id)).toBeUndefined()
    })

    it('악의적 tenant_id 변조 시도는 차단된다', async () => {
      // Given
      const orgAUser = await createTestUser({ orgId: 'org-a' })
      await supabase.auth.signInWithPassword({
        email: orgAUser.email,
        password: 'test123'
      })

      // When: 다른 조직 ID로 데이터 삽입 시도
      const { error } = await supabase
        .from('documents')
        .insert({
          tenant_id: 'org-b',  // 악의적 시도!
          title: 'Malicious Doc'
        })

      // Then: 데이터베이스 레벨에서 차단
      expect(error).toBeDefined()
      expect(error.message).toContain('Cross-tenant access denied')
    })

    it('SQL Injection 공격은 방어된다', async () => {
      // Given
      const orgAUser = await createTestUser({ orgId: 'org-a' })
      await supabase.auth.signInWithPassword({
        email: orgAUser.email,
        password: 'test123'
      })

      // When: SQL Injection 시도
      const maliciousQuery = "' OR '1'='1' --"
      const { data, error } = await supabase
        .from('documents')
        .select('*')
        .eq('title', maliciousQuery)

      // Then: 안전하게 처리됨 (RLS가 여전히 적용됨)
      expect(error).toBeNull()
      expect(data).toHaveLength(0) // RLS로 인해 다른 조직 데이터 접근 불가
    })
  })

  describe('성능', () => {
    it('RLS 적용 후에도 쿼리 성능이 유지된다', async () => {
      // Given
      const orgAUser = await createTestUser({ orgId: 'org-a' })
      await createTestDocuments(1000, { orgId: 'org-a' })
      await supabase.auth.signInWithPassword({
        email: orgAUser.email,
        password: 'test123'
      })

      // When
      const startTime = performance.now()
      await supabase
        .from('documents')
        .select('*')
        .order('created_at', { ascending: false })
        .limit(20)
      const endTime = performance.now()

      // Then: 5ms 이하 오버헤드
      const queryTime = endTime - startTime
      expect(queryTime).toBeLessThan(50) // 총 50ms 이하
    })
  })

  describe('관리자 권한', () => {
    it('Super Admin은 모든 조직 데이터를 조회할 수 있다', async () => {
      // Given
      const superAdmin = await createTestUser({
        role: 'super_admin'
      })
      await createTestDocuments(10, { orgId: 'org-a' })
      await createTestDocuments(10, { orgId: 'org-b' })

      // When
      await supabase.auth.signInWithPassword({
        email: superAdmin.email,
        password: 'test123'
      })

      const { data } = await supabase
        .from('documents')
        .select('*')

      // Then: 모든 조직 데이터 반환
      expect(data).toHaveLength(20)
    })

    it('Org Admin은 자신의 조직만 관리할 수 있다', async () => {
      // Given
      const orgAdmin = await createTestUser({
        orgId: 'org-a',
        role: 'org_admin'
      })
      const orgAUser = await createTestUser({ orgId: 'org-a' })
      const orgBUser = await createTestUser({ orgId: 'org-b' })

      // When: Org A Admin으로 로그인
      await supabase.auth.signInWithPassword({
        email: orgAdmin.email,
        password: 'test123'
      })

      // Then: Org A 사용자는 수정 가능
      const { error: errorA } = await supabase
        .from('users')
        .update({ name: 'Updated' })
        .eq('id', orgAUser.id)
      expect(errorA).toBeNull()

      // Then: Org B 사용자는 수정 불가
      const { error: errorB } = await supabase
        .from('users')
        .update({ name: 'Updated' })
        .eq('id', orgBUser.id)
      expect(errorB).toBeDefined()
    })
  })
})
```

---

### Phase 3: 감사 로그 시스템 (2주)

#### SPEC-SEC-003: Audit Logging

```markdown
# SPEC-SEC-003: Comprehensive Audit Logging

@TAG:SPEC-SEC-003

## 요구사항

**UBIQUITOUS**:
- 시스템은 모든 중요 작업을 감사 로그에 기록해야 한다

**EVENT-DRIVEN**:
- WHEN 사용자가 데이터를 생성/수정/삭제하면
- THEN 시스템은 누가, 언제, 무엇을, 어떻게 변경했는지 기록해야 한다

**STATE-DRIVEN**:
- WHILE 감사 로그가 저장되는 동안
- THEN 시스템은 로그의 불변성을 보장해야 한다 (수정/삭제 불가)

## 인수 기준

1. ✅ 100% 작업 추적 (CRUD 모두)
2. ✅ 로그 불변성 (Append-only)
3. ✅ 7년 보관 (규정 준수)
4. ✅ 실시간 알림 (이상 행위 감지)
5. ✅ 검색 및 필터링 (감사 조사 지원)

## 기록 대상

- 로그인/로그아웃
- 데이터 생성/수정/삭제
- 권한 변경
- 설정 변경
- API 호출 (실패 포함)
- 파일 다운로드
```

#### Audit Log 구현

```typescript
// @TAG:CODE-SEC-003:LIB
// lib/audit/audit-logger.ts

import { supabase } from '@/lib/supabase'

export interface AuditLogEntry {
  id?: string
  userId: string
  userName: string
  tenantId: string
  action: AuditAction
  resourceType: string
  resourceId: string
  changes?: Record<string, any>
  metadata: {
    ipAddress: string
    userAgent: string
    timestamp: Date
    requestId: string
  }
  status: 'success' | 'failure'
  errorMessage?: string
}

export enum AuditAction {
  // 인증
  LOGIN = 'auth.login',
  LOGOUT = 'auth.logout',
  MFA_VERIFY = 'auth.mfa_verify',

  // 데이터 작업
  CREATE = 'data.create',
  READ = 'data.read',
  UPDATE = 'data.update',
  DELETE = 'data.delete',

  // 권한
  PERMISSION_GRANT = 'permission.grant',
  PERMISSION_REVOKE = 'permission.revoke',
  ROLE_CHANGE = 'permission.role_change',

  // 파일
  FILE_UPLOAD = 'file.upload',
  FILE_DOWNLOAD = 'file.download',
  FILE_DELETE = 'file.delete',

  // 설정
  SETTINGS_CHANGE = 'settings.change',
  INTEGRATION_ADD = 'integration.add',
  INTEGRATION_REMOVE = 'integration.remove'
}

/**
 * 감사 로그를 기록합니다
 * @TAG:SEC-003
 */
export async function logAudit(entry: AuditLogEntry): Promise<void> {
  try {
    // 1. 변경 사항 diff 생성
    const diff = entry.changes
      ? generateDiff(entry.changes)
      : null

    // 2. 민감 정보 마스킹
    const sanitized = maskSensitiveData(entry)

    // 3. 데이터베이스에 저장 (Append-only)
    const { error } = await supabase
      .from('audit_logs')
      .insert({
        ...sanitized,
        changes_diff: diff,
        created_at: new Date().toISOString()
      })

    if (error) {
      // 감사 로그 실패는 심각한 문제
      console.error('[CRITICAL] Audit log failed:', error)
      await alertSecurityTeam('Audit log failure', entry)
    }

    // 4. 이상 행위 감지
    await detectAnomalies(entry)

  } catch (error) {
    // 감사 로그 시스템은 절대 실패하면 안됨
    console.error('[CRITICAL] Audit system error:', error)
    await alertSecurityTeam('Audit system error', error)
  }
}

/**
 * 변경 사항 diff 생성
 */
function generateDiff(changes: Record<string, any>): string {
  const { before, after } = changes

  if (!before || !after) return JSON.stringify(changes)

  const diff: Record<string, any> = {}

  for (const key of Object.keys(after)) {
    if (before[key] !== after[key]) {
      diff[key] = {
        before: before[key],
        after: after[key]
      }
    }
  }

  return JSON.stringify(diff)
}

/**
 * 민감 정보 마스킹
 */
function maskSensitiveData(entry: AuditLogEntry): AuditLogEntry {
  const masked = { ...entry }

  // 비밀번호, 토큰 등 마스킹
  const sensitiveFields = ['password', 'token', 'secret', 'api_key']

  if (masked.changes) {
    for (const field of sensitiveFields) {
      if (field in masked.changes) {
        masked.changes[field] = '***MASKED***'
      }
    }
  }

  return masked
}

/**
 * 이상 행위 감지
 */
async function detectAnomalies(entry: AuditLogEntry): Promise<void> {
  // 1. 단기간 다량 요청 (DDoS, Brute Force)
  const recentLogs = await getRecentLogs(entry.userId, 60) // 1분
  if (recentLogs.length > 100) {
    await alertSecurityTeam('Potential DDoS attack', entry)
  }

  // 2. 비정상적 시간 접근 (새벽 3시 로그인 등)
  const hour = new Date(entry.metadata.timestamp).getHours()
  if (hour >= 2 && hour <= 5) {
    await alertSecurityTeam('Unusual login time', entry)
  }

  // 3. 다량 데이터 다운로드
  if (entry.action === AuditAction.FILE_DOWNLOAD) {
    const downloadsToday = await countTodayDownloads(entry.userId)
    if (downloadsToday > 100) {
      await alertSecurityTeam('Excessive file downloads', entry)
    }
  }

  // 4. 권한 변경 (특히 민감한 작업)
  if (entry.action === AuditAction.PERMISSION_GRANT) {
    await alertSecurityTeam('Permission granted', entry, 'info')
  }
}

/**
 * 보안 팀 알림
 */
async function alertSecurityTeam(
  alertType: string,
  data: any,
  severity: 'info' | 'warning' | 'critical' = 'warning'
): Promise<void> {
  // Slack, PagerDuty 등으로 알림
  await sendSlackAlert({
    channel: '#security-alerts',
    severity,
    message: `[${severity.toUpperCase()}] ${alertType}`,
    data
  })

  if (severity === 'critical') {
    await sendPagerDutyAlert({
      service: 'security',
      incident: alertType,
      data
    })
  }
}
```

#### Audit Log Database Schema

```sql
-- @TAG:CODE-SEC-003:DB
-- supabase/migrations/003_audit_logs.sql

-- Audit Logs 테이블 (Append-only)
CREATE TABLE audit_logs (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,

  -- 사용자 정보
  user_id UUID NOT NULL,
  user_name TEXT NOT NULL,
  tenant_id UUID NOT NULL,

  -- 작업 정보
  action TEXT NOT NULL,
  resource_type TEXT NOT NULL,
  resource_id TEXT NOT NULL,

  -- 변경 내역
  changes_diff JSONB,

  -- 메타데이터
  ip_address INET NOT NULL,
  user_agent TEXT NOT NULL,
  request_id TEXT NOT NULL,

  -- 결과
  status TEXT NOT NULL CHECK (status IN ('success', 'failure')),
  error_message TEXT,

  -- 타임스탬프 (불변)
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  -- 인덱스용 컬럼
  action_category TEXT GENERATED ALWAYS AS (
    split_part(action, '.', 1)
  ) STORED
);

-- 인덱스 (빠른 검색)
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- Row Level Security
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

-- 정책: 자신의 조직 로그만 조회 가능
CREATE POLICY "tenant_isolation_audit_logs"
  ON audit_logs FOR SELECT
  USING (
    tenant_id = auth.jwt() ->> 'tenant_id'::UUID
    OR auth.jwt() ->> 'role' = 'super_admin'
  );

-- 정책: 삽입만 가능 (수정/삭제 불가)
CREATE POLICY "audit_logs_insert_only"
  ON audit_logs FOR INSERT
  WITH CHECK (true);

-- 트리거: 수정/삭제 방지 (불변성 보장)
CREATE OR REPLACE FUNCTION prevent_audit_log_modification()
RETURNS TRIGGER AS $$
BEGIN
  RAISE EXCEPTION 'Audit logs are immutable and cannot be modified or deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_audit_log_update
  BEFORE UPDATE ON audit_logs
  FOR EACH ROW
  EXECUTE FUNCTION prevent_audit_log_modification();

CREATE TRIGGER prevent_audit_log_delete
  BEFORE DELETE ON audit_logs
  FOR EACH ROW
  EXECUTE FUNCTION prevent_audit_log_modification();

-- 파티셔닝 (7년 보관, 성능 최적화)
CREATE TABLE audit_logs_y2024m01 PARTITION OF audit_logs
  FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- 자동 파티션 생성 함수
CREATE OR REPLACE FUNCTION create_monthly_audit_log_partition()
RETURNS void AS $$
DECLARE
  start_date DATE;
  end_date DATE;
  table_name TEXT;
BEGIN
  start_date := date_trunc('month', NOW());
  end_date := start_date + INTERVAL '1 month';
  table_name := 'audit_logs_y' || to_char(start_date, 'YYYYmMM');

  EXECUTE format(
    'CREATE TABLE IF NOT EXISTS %I PARTITION OF audit_logs
     FOR VALUES FROM (%L) TO (%L)',
    table_name, start_date, end_date
  );
END;
$$ LANGUAGE plpgsql;

-- 매월 1일에 자동 실행
-- (pg_cron 확장 필요: CREATE EXTENSION pg_cron;)
SELECT cron.schedule(
  'create-monthly-audit-partition',
  '0 0 1 * *',  -- 매월 1일 00:00
  'SELECT create_monthly_audit_log_partition()'
);
```

---

## 📊 성과 및 결과

### 정량적 성과

| 지표 | Before | After | 개선 |
|------|--------|-------|------|
| **SOC 2 준수** | ❌ 없음 | ✅ Type 2 인증 | 100% |
| **보안 사고** | 2건/월 | 0건 (6개월) | 100% 감소 |
| **엔터프라이즈 고객** | 5개사 | 20개사 | 300% 증가 |
| **ARR** | $500K | $2.1M | 320% 성장 |
| **RLS 커버리지** | 0% (App 레벨) | 100% (DB 레벨) | - |
| **감사 로그** | 없음 | 100% 추적 | - |
| **성능 영향** | - | < 5ms | 목표 달성 |
| **MFA 적용률** | 0% | 100% (강제) | - |

### 정성적 성과

**1. 규정 준수**
- SOC 2 Type 2 첫 감사 통과 (3개월 내)
- GDPR, HIPAA ready 인프라
- 연간 침투 테스트 통과

**2. 고객 신뢰**
- Fortune 500 기업 계약 체결
- 보안 문의 응답 시간 90% 단축
- 계약 전환율 40% → 75%

**3. 운영 효율**
- 보안 사고 조사 시간 85% 단축 (감사 로그 덕분)
- 수동 보안 리뷰 불필요 (자동화)
- 온콜 알림 70% 감소 (이상 탐지 정확도)

---

## 💡 배운 교훈

### 1. Database-Level 보안의 중요성

**실제 사례**:
Week 5에 애플리케이션 버그로 tenant_id 필터가 누락된 쿼리가 배포되었습니다.

```typescript
// ❌ 버그: tenant_id 필터 누락
async function getDocuments() {
  return await db.documents.findMany()  // 모든 조직 데이터 반환!
}
```

**RLS가 없었다면**: 모든 조직의 데이터 유출 (심각한 보안 사고)

**RLS 덕분에**: 데이터베이스 레벨에서 자동 차단, 피해 없음

**교훈**: Application 레벨 보안은 부족하다. Database-Level 보안이 필수!

---

### 2. 감사 로그의 비즈니스 가치

단순히 규정 준수용이 아닌 **비즈니스 인텔리전스** 도구로 활용:

**사례 1: 사용 패턴 분석**
- 고객이 어떤 기능을 가장 많이 사용하는지 파악
- 제품 로드맵 우선순위 결정

**사례 2: 고객 지원 개선**
- "이상한 일이 생겼어요" → 감사 로그로 즉시 원인 파악
- 고객 지원 만족도 30% 상승

**사례 3: 내부 감사**
- 직원의 고객 데이터 접근 이력 추적
- 내부 규정 준수 강화

---

### 3. MoAI-ADK Security Framework의 위력

**Before MoAI-ADK**:
- 보안 Best Practices 연구에 주당 10시간 소요
- OWASP Top 10 수동 체크
- 보안 전문가 1명이 병목

**After MoAI-ADK**:
- security-expert 에이전트가 자동 체크
- Senior Engineer Thinking으로 최신 패턴 학습
- 모든 개발자가 보안 Best Practices 적용

**구체적 사례**:

```bash
# Alfred가 자동으로 보안 취약점 감지 및 제안
/alfred:2-run SEC-001

# Alfred의 제안:
# 1. CORS 설정 강화
# 2. Rate Limiting 추가
# 3. Input Validation 강화
# 4. SQL Injection 방어 확인
# 5. XSS 방어 확인
```

---

## 🎯 권장 사항

### Enterprise SaaS 보안 체크리스트

#### 1. 인증 및 권한 (필수)

- [ ] **SSO/SAML 지원** (Okta, Azure AD, Google Workspace)
- [ ] **MFA 강제** (TOTP, SMS, Hardware Token)
- [ ] **Session 관리** (Idle timeout, Absolute timeout)
- [ ] **JWT 보안** (RS256, Refresh Token Rotation)
- [ ] **비밀번호 정책** (복잡도, 재사용 방지, 정기 변경)

#### 2. 데이터 보안 (필수)

- [ ] **Row Level Security** (데이터베이스 레벨 격리)
- [ ] **암호화** (저장: AES-256, 전송: TLS 1.3)
- [ ] **백업 암호화** (7년 보관)
- [ ] **PII 마스킹** (로그, 오류 메시지)
- [ ] **GDPR 준수** (Right to be forgotten)

#### 3. 감사 및 모니터링 (필수)

- [ ] **감사 로그** (100% 추적, Append-only)
- [ ] **이상 탐지** (Brute Force, DDoS, 내부자 위협)
- [ ] **실시간 알림** (보안 사고 즉시 통지)
- [ ] **로그 보관** (7년, 파티셔닝)
- [ ] **침투 테스트** (연 2회)

#### 4. 인프라 보안 (권장)

- [ ] **WAF** (Web Application Firewall)
- [ ] **DDoS 방어** (Cloudflare, AWS Shield)
- [ ] **Rate Limiting** (API 남용 방지)
- [ ] **보안 헤더** (HSTS, CSP, X-Frame-Options)
- [ ] **Dependency 스캔** (Snyk, Dependabot)

#### 5. 규정 준수 (엔터프라이즈)

- [ ] **SOC 2 Type 2** (연간 감사)
- [ ] **GDPR** (유럽 고객)
- [ ] **HIPAA** (의료 데이터)
- [ ] **ISO 27001** (정보 보안 관리)
- [ ] **PCI DSS** (결제 정보 처리)

---

### MoAI-ADK로 시작하기

```bash
# 1. 보안 감사 실행
/alfred:1-plan "Security audit for enterprise SaaS"

# 2. security-expert 에이전트 활용
/alfred:2-run SEC-001  # Enterprise Authentication
/alfred:2-run SEC-002  # Multi-tenant RLS
/alfred:2-run SEC-003  # Audit Logging

# 3. 자동 보안 테스트
npm run test:security

# 4. 문서 동기화
/alfred:3-sync auto ALL
```

---

## 📚 관련 자료

- [MoAI-ADK 시작하기](/ko/getting-started)
- [security-expert 에이전트](/ko/agents/security-expert)
- [Supabase RLS 가이드](/ko/skills/baas/supabase-rls)
- [Auth0 통합](/ko/skills/baas/auth0)
- [SOC 2 준비 가이드](/ko/guides/soc2-compliance)

---

## 💬 질문이 있으신가요?

이 사례 연구에 대해 궁금한 점이 있으시면:

- **GitHub Discussions**: [질문하기](https://github.com/modu-ai/moai-adk/discussions)
- **Discord**: [#security 채널](https://discord.gg/moai-adk)
- **이메일**: security@moai-adk.com

---

**다음 사례 연구**: [Microservices 아키텍처 전환 →](/ko/case-studies/microservices-migration)
