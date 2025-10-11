# spec-builder Agent Guide

<!-- @CODE:DOCS-002 | SPEC: .moai/specs/SPEC-DOCS-002/spec.md -->

> "명세 없으면 코드 없다. 모든 훌륭한 개발은 명확한 SPEC에서 시작된다."

**spec-builder**는 MoAI-ADK의 핵심 에이전트로, SPEC-First TDD 워크플로우의 시작점을 담당합니다. 비즈니스 요구사항을 체계적인 EARS 명세로 변환하고, 프로젝트 전체의 개발 방향을 설정하는 시스템 아키텍트 역할을 수행합니다.

---

## 목차

1. [에이전트 개요](#에이전트-개요)
2. [핵심 책임과 역할](#핵심-책임과-역할)
3. [워크플로우 상세](#워크플로우-상세)
4. [EARS 명세 작성법](#ears-명세-작성법)
5. [Personal vs Team 모드](#personal-vs-team-모드)
6. [실전 시나리오](#실전-시나리오)
7. [Best Practices](#best-practices)
8. [안티 패턴](#안티-패턴)
9. [문제 해결 가이드](#문제-해결-가이드)
10. [관련 문서](#관련-문서)

---

## 에이전트 개요

### 기본 정보

| 항목 | 내용 |
|------|------|
| **아이콘** | 🏗️ |
| **이름** | spec-builder |
| **페르소나** | 시스템 아키텍트 (System Architect) |
| **전문 영역** | 요구사항 분석, EARS 명세, 아키텍처 설계 |
| **커맨드** | `/alfred:1-spec` |
| **모델** | Sonnet |

### 전문가 특성

**spec-builder**는 다음과 같은 특성을 가진 시스템 아키텍트입니다:

- **사고 방식**: 비즈니스 요구사항을 체계적인 EARS 구문과 아키텍처 패턴으로 구조화
- **의사결정 기준**: 명확성, 완전성, 추적성, 확장성이 모든 설계 결정의 기준
- **커뮤니케이션 스타일**: 정확하고 구조화된 질문을 통해 요구사항과 제약사항을 명확히 도출
- **전문 분야**: EARS 방법론, 시스템 아키텍처, 요구사항 공학

### 위치와 역할

```
MoAI-ADK 3단계 워크플로우

┌─────────────────────────────────────────────────────────────┐
│  1단계: SPEC 작성 (spec-builder) 🏗️                           │
│  → 비즈니스 요구사항을 EARS 명세로 변환                         │
│  → SPEC 문서 작성 (spec.md, plan.md, acceptance.md)          │
│  → Git 브랜치 생성 (feature/SPEC-{ID})                        │
│  → Draft PR 생성 (Team 모드)                                  │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  2단계: TDD 구현 (code-builder) 💎                             │
│  → /alfred:2-build SPEC-{ID}                                 │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│  3단계: 문서 동기화 (doc-syncer) 📖                            │
│  → /alfred:3-sync                                            │
└─────────────────────────────────────────────────────────────┘
```

**spec-builder**는 모든 개발의 시작점입니다. 명확한 SPEC이 없으면 코드 작성이 불가능하며, 이후 모든 단계는 SPEC을 기준으로 진행됩니다.

---

## 핵심 책임과 역할

### 전담 영역

spec-builder는 다음 작업만 담당합니다:

1. **프로젝트 문서 분석**
   - `.moai/project/product.md` - 비즈니스 요구사항, 사용자 스토리
   - `.moai/project/structure.md` - 아키텍처 설계 참고
   - `.moai/project/tech.md` - 기술 스택 참고

2. **SPEC 문서 작성**
   - EARS 방식 요구사항 명세
   - SPEC 메타데이터 관리 (id, version, status 등)
   - 구현 계획 및 수락 기준 초기화

3. **산출물 생성**
   - **Personal 모드**: 3개 파일 (spec.md, plan.md, acceptance.md)
   - **Team 모드**: GitHub Issue + SPEC 문서

4. **품질 검증**
   - EARS 구문 준수 검증
   - SPEC 메타데이터 완전성 확인
   - TAG ID 중복 검사
   - 프로젝트 문서와 일관성 검증

### 위임하는 영역 (git-manager 담당)

spec-builder는 다음 작업을 **git-manager 에이전트에게 위임**합니다:

- Git 브랜치 생성 (feature/SPEC-{ID})
- GitHub Issue/PR 생성 (Team 모드)
- 커밋 생성 및 원격 동기화
- 모든 Git 관련 작업

**중요**: spec-builder는 git-manager를 직접 호출하지 않습니다. Alfred가 순차적으로 조율합니다.

---

## 워크플로우 상세

### 실행 방법

```bash
# 자동 제안 방식 (프로젝트 문서 기반)
/alfred:1-spec

# 수동 지정 방식 (특정 기능 명시)
/alfred:1-spec "JWT 인증 시스템"
/alfred:1-spec "파일 업로드 기능" "결제 처리 시스템"

# 기존 SPEC 수정
/alfred:1-spec SPEC-AUTH-001 "리프레시 토큰 만료시간 변경"
```

### Phase 1: 분석 및 계획 수립

**목표**: 프로젝트 상태를 분석하고 SPEC 작성 계획을 수립합니다.

#### 1.1 프로젝트 문서 확인

spec-builder는 다음 순서로 문서를 로드합니다:

**필수 로드** (항상):
```bash
.moai/config.json           # 프로젝트 모드 확인 (Personal/Team)
.moai/project/product.md    # 비즈니스 요구사항, 사용자 스토리
.moai/memory/spec-metadata.md  # SPEC 메타데이터 표준 (필수/선택 필드)
```

**조건부 로드** (필요 시):
```bash
.moai/project/structure.md  # 아키텍처 설계가 필요한 경우
.moai/project/tech.md       # 기술 스택 선정/변경이 필요한 경우
.moai/specs/SPEC-*/spec.md  # 유사 기능 참조가 필요한 경우
```

**참조 문서** (SPEC 작성 중 필요 시):
```bash
.moai/memory/development-guide.md  # EARS 템플릿, TAG 규칙
```

#### 1.2 기능 후보 도출

**자동 제안 방식** (`/alfred:1-spec`):

spec-builder가 프로젝트 문서를 분석하여 후보를 제안합니다:

```markdown
## 📋 SPEC 후보 분석 결과

### product.md 분석
- ✅ 사용자 인증 기능 (JWT 기반)
- ✅ 파일 업로드 기능 (멀티파트)
- ⚠️ 결제 처리 시스템 (외부 API 의존성)

### 우선순위 제안
1. **AUTH-001**: JWT 인증 시스템 (HIGH - 다른 기능의 선행 요구사항)
2. **UPLOAD-001**: 파일 업로드 기능 (MEDIUM - 독립적 구현 가능)
3. **PAYMENT-001**: 결제 처리 시스템 (LOW - AUTH-001 의존)

### 다음 단계
어떤 기능부터 시작하시겠습니까?
- "AUTH-001 진행"
- "UPLOAD-001 진행"
- "수정 [내용]"
```

**수동 지정 방식** (`/alfred:1-spec "기능명"`):

사용자가 지정한 기능을 바로 분석합니다:

```markdown
## 📋 SPEC 작성 계획

### 요청 기능
- JWT 인증 시스템

### 분석 결과
- **도메인**: AUTH
- **ID 제안**: AUTH-001
- **우선순위**: HIGH
- **의존성**: 없음

### 생성 파일
- `.moai/specs/SPEC-AUTH-001/spec.md`
- `.moai/specs/SPEC-AUTH-001/plan.md`
- `.moai/specs/SPEC-AUTH-001/acceptance.md`

진행하시겠습니까? (진행/수정/중단)
```

#### 1.3 ID 중복 검사

**필수 단계**: SPEC ID가 이미 존재하는지 확인합니다.

```bash
# ID 중복 확인
rg "@SPEC:AUTH-001" -n .moai/specs/

# 디렉토리 중복 확인
ls .moai/specs/ | grep "SPEC-AUTH-001"
```

**중복 발견 시**:
```markdown
⚠️ 중복 ID 발견

기존 SPEC이 존재합니다:
- `.moai/specs/SPEC-AUTH-001/spec.md`
- 버전: v0.1.0
- 상태: completed

옵션:
1. "AUTH-002로 변경" (새로운 ID 사용)
2. "SPEC-AUTH-001 수정" (기존 SPEC 업데이트)
3. "중단"
```

#### 1.4 계획 보고서 생성

spec-builder는 최종 계획을 사용자에게 확인받습니다:

```markdown
## 📋 SPEC 작성 최종 계획

### SPEC 정보
- **ID**: AUTH-001
- **제목**: JWT 인증 시스템
- **도메인**: 인증/보안
- **우선순위**: HIGH
- **모드**: Personal

### 생성 파일 (3개)
1. `.moai/specs/SPEC-AUTH-001/spec.md`
   - EARS 방식 요구사항 명세
   - YAML Front Matter (필수 필드 7개)
   - HISTORY 섹션

2. `.moai/specs/SPEC-AUTH-001/plan.md`
   - 구현 계획 및 우선순위
   - 기술적 접근 방법
   - 아키텍처 설계 방향

3. `.moai/specs/SPEC-AUTH-001/acceptance.md`
   - Given-When-Then 테스트 시나리오
   - 품질 게이트 기준
   - 완료 조건 (Definition of Done)

### 검증 완료
- ✅ ID 중복 없음
- ✅ 디렉토리명 형식 올바름 (SPEC-AUTH-001)
- ✅ 필수 문서 로드 완료

### 다음 단계
"진행" 입력 시:
1. 3개 파일 동시 생성 (MultiEdit)
2. Alfred가 git-manager 호출 (브랜치 생성, 커밋)
3. /alfred:2-build AUTH-001 안내

계속 진행하시겠습니까? (진행/수정/중단)
```

### Phase 2: 실행 (사용자 승인 후)

사용자가 **"진행"**을 입력하면 Phase 2가 시작됩니다.

#### 2.1 SPEC 문서 작성

**중요**: Personal 모드에서는 **반드시 MultiEdit 도구 사용**하여 3개 파일을 동시에 생성합니다.

**올바른 방법** (✅):
```json
{
  "files": [
    {
      "path": ".moai/specs/SPEC-AUTH-001/spec.md",
      "content": "..."
    },
    {
      "path": ".moai/specs/SPEC-AUTH-001/plan.md",
      "content": "..."
    },
    {
      "path": ".moai/specs/SPEC-AUTH-001/acceptance.md",
      "content": "..."
    }
  ]
}
```

**잘못된 방법** (❌):
```bash
# Write 도구로 순차 생성 (비효율적)
Write spec.md
Write plan.md
Write acceptance.md
```

#### 2.2 파일별 상세 내용

##### spec.md (EARS 명세)

```markdown
---
# 필수 필드 (7개)
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-11
updated: 2025-10-11
author: @Goos
priority: high

# 선택 필드
category: feature
labels:
  - authentication
  - jwt
  - security
---

# @SPEC:AUTH-001: JWT 인증 시스템

## HISTORY
### v0.0.1 (2025-10-11)
- **INITIAL**: JWT 기반 인증 시스템 명세 작성
- **AUTHOR**: @Goos
- **REASON**: 사용자 인증 및 세션 관리 필요

---

## 개요

사용자 인증을 위한 JWT 기반 시스템을 구현합니다.

### 목표
- 안전한 사용자 인증 제공
- 액세스 토큰 및 리프레시 토큰 발급
- 토큰 기반 API 접근 제어

---

## Requirements (기능 요구사항)

### Ubiquitous Requirements (기본 요구사항)
- 시스템은 JWT 기반 인증 기능을 제공해야 한다
- 시스템은 토큰 갱신 기능을 제공해야 한다
- 시스템은 사용자 로그아웃 기능을 제공해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 사용자가 유효한 이메일/비밀번호로 로그인하면, 시스템은 액세스 토큰과 리프레시 토큰을 발급해야 한다
- WHEN 액세스 토큰이 만료되면, 시스템은 401 Unauthorized 에러를 반환해야 한다
- WHEN 리프레시 토큰으로 갱신 요청이 오면, 시스템은 새로운 액세스 토큰을 발급해야 한다

### State-driven Requirements (상태 기반)
- WHILE 사용자가 인증된 상태일 때, 시스템은 API 요청에 사용자 정보를 포함해야 한다
- WHILE 토큰이 유효한 동안, 시스템은 자동 재로그인 없이 API 접근을 허용해야 한다

### Optional Features (선택적 기능)
- WHERE 2FA(이중 인증)가 활성화되어 있으면, 시스템은 추가 인증 코드를 요구할 수 있다

### Constraints (제약사항)
- IF 잘못된 자격증명이 5회 연속 입력되면, 시스템은 계정을 15분간 잠가야 한다
- 액세스 토큰 만료시간은 15분을 초과하지 않아야 한다
- 리프레시 토큰 만료시간은 7일을 초과하지 않아야 한다

---

## Specifications (상세 명세)

### 인증 흐름
1. 사용자가 이메일/비밀번호 입력
2. 서버가 자격증명 검증
3. 성공 시 JWT 토큰 발급 (액세스 + 리프레시)
4. 클라이언트는 액세스 토큰을 API 요청에 포함
5. 만료 시 리프레시 토큰으로 갱신

### 토큰 구조
- **액세스 토큰**: { userId, email, exp: 15m }
- **리프레시 토큰**: { userId, exp: 7d }

### 보안 요구사항
- 비밀번호는 bcrypt (cost 12)로 해싱
- JWT 시크릿은 환경변수로 관리
- HTTPS 필수

---

## Traceability (추적성)

### TAG 체인
- @SPEC:AUTH-001 (본 문서)
- @TEST:AUTH-001 (tests/auth/)
- @CODE:AUTH-001 (src/auth/)
- @DOC:AUTH-001 (docs/features/)

### 관련 문서
- [EARS 가이드](../../../guides/concepts/ears-guide.md)
- [TRUST 원칙](../../../guides/concepts/trust-principles.md)
```

##### plan.md (구현 계획)

```markdown
# AUTH-001 구현 계획

## 우선순위별 작업

### 1차 목표 (필수 구현)
- [ ] JWT 토큰 생성 유틸리티
- [ ] 로그인 엔드포인트 (/auth/login)
- [ ] 토큰 검증 미들웨어
- [ ] 리프레시 토큰 엔드포인트 (/auth/refresh)

### 2차 목표 (보안 강화)
- [ ] 비밀번호 해싱 (bcrypt)
- [ ] 토큰 블랙리스트 (로그아웃)
- [ ] Rate limiting (5회 실패 시 잠금)

### 3차 목표 (최적화)
- [ ] 토큰 캐싱 (Redis)
- [ ] 2FA 통합 준비

## 기술적 접근

### 기술 스택
- **언어**: TypeScript
- **프레임워크**: Express.js
- **JWT 라이브러리**: jsonwebtoken
- **해싱**: bcrypt
- **데이터베이스**: PostgreSQL

### 아키텍처 설계
```
src/auth/
├── services/
│   ├── auth-service.ts       # 비즈니스 로직
│   └── token-service.ts      # JWT 관리
├── middlewares/
│   └── auth-middleware.ts    # 토큰 검증
├── routes/
│   └── auth-routes.ts        # API 엔드포인트
└── utils/
    └── password-hasher.ts    # 비밀번호 해싱
```

### 의존성
- 없음 (독립적 구현 가능)

## 리스크 및 대응

### 리스크 1: JWT 시크릿 노출
- **대응**: 환경변수 필수, .env 파일 gitignore 추가

### 리스크 2: 토큰 탈취
- **대응**: HTTPS 필수, 리프레시 토큰 Rotation 구현

---

**참고**: 시간 예측은 포함하지 않습니다. 우선순위 기반으로 진행합니다.
```

##### acceptance.md (수락 기준)

```markdown
# AUTH-001 수락 기준

## Given-When-Then 테스트 시나리오

### 시나리오 1: 성공적인 로그인
```gherkin
Given 사용자가 유효한 계정을 가지고 있고
When 올바른 이메일/비밀번호로 로그인하면
Then 액세스 토큰과 리프레시 토큰을 받아야 한다
And HTTP 상태 코드는 200이어야 한다
```

### 시나리오 2: 잘못된 자격증명
```gherkin
Given 사용자가 잘못된 비밀번호를 입력하고
When 로그인을 시도하면
Then 인증 실패 메시지를 받아야 한다
And HTTP 상태 코드는 401이어야 한다
```

### 시나리오 3: 토큰 갱신
```gherkin
Given 액세스 토큰이 만료되었고
And 유효한 리프레시 토큰을 가지고 있으면
When 토큰 갱신을 요청하면
Then 새로운 액세스 토큰을 받아야 한다
```

### 시나리오 4: 계정 잠금
```gherkin
Given 사용자가 5회 연속 로그인 실패했을 때
When 6번째 로그인을 시도하면
Then 계정 잠금 메시지를 받아야 한다
And 15분 동안 로그인이 차단되어야 한다
```

## 품질 게이트

### TRUST 원칙
- **T - Test**: 테스트 커버리지 ≥85%
- **R - Readable**: 린터 0 issues
- **U - Unified**: TypeScript strict 모드 통과
- **S - Secured**: 0 security vulnerabilities
- **T - Trackable**: TAG 체인 완전성 확인

### 성능 기준
- 로그인 응답 시간: 평균 200ms 이하
- 토큰 검증: 50ms 이하

### 보안 기준
- bcrypt cost factor: 12
- JWT 시크릿: 최소 256비트
- HTTPS 필수

## 완료 조건 (Definition of Done)

- [ ] 모든 EARS 요구사항 구현 완료
- [ ] 테스트 커버리지 85% 이상
- [ ] 모든 테스트 통과
- [ ] 린터 0 issues
- [ ] 보안 스캔 통과
- [ ] TAG 체인 완전성 확인
- [ ] Living Document 생성 완료
- [ ] PR 머지 (Team 모드)
```

#### 2.3 Git 작업 안내

spec-builder는 SPEC 문서 작성 후 **Alfred에게 보고**합니다. Alfred는 이어서 **git-manager**를 호출하여 Git 작업을 수행합니다.

```markdown
## ✅ SPEC 작성 완료

### 생성 파일
- `.moai/specs/SPEC-AUTH-001/spec.md` (EARS 명세)
- `.moai/specs/SPEC-AUTH-001/plan.md` (구현 계획)
- `.moai/specs/SPEC-AUTH-001/acceptance.md` (수락 기준)

### 다음 단계 (Alfred가 자동 진행)
1. git-manager 호출
   - 브랜치 생성: feature/SPEC-AUTH-001
   - 커밋: "🔴 RED: SPEC-AUTH-001 명세 작성"
   - Draft PR 생성 (Team 모드)

2. 구현 시작 안내
   - /alfred:2-build AUTH-001

Alfred가 나머지 작업을 진행합니다...
```

---

## EARS 명세 작성법

### EARS란?

**EARS (Easy Approach to Requirements Syntax)** 는 체계적이고 명확한 요구사항 작성을 위한 방법론입니다.

### 5가지 구문 유형

#### 1. Ubiquitous (기본 요구사항)

**형식**: 시스템은 [기능]을 제공해야 한다

**예시**:
```markdown
- 시스템은 사용자 인증 기능을 제공해야 한다
- 시스템은 파일 업로드 기능을 제공해야 한다
- 시스템은 실시간 알림 기능을 제공해야 한다
```

#### 2. Event-driven (이벤트 기반)

**형식**: WHEN [조건]이면, 시스템은 [동작]해야 한다

**예시**:
```markdown
- WHEN 사용자가 유효한 자격증명으로 로그인하면, 시스템은 JWT 토큰을 발급해야 한다
- WHEN 파일 업로드가 완료되면, 시스템은 성공 메시지를 표시해야 한다
- WHEN 결제가 실패하면, 시스템은 실패 사유를 사용자에게 표시해야 한다
```

#### 3. State-driven (상태 기반)

**형식**: WHILE [상태]일 때, 시스템은 [동작]해야 한다

**예시**:
```markdown
- WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 리소스 접근을 허용해야 한다
- WHILE 파일 업로드가 진행 중일 때, 시스템은 진행률을 실시간으로 표시해야 한다
- WHILE 결제 처리가 진행 중일 때, 시스템은 중복 요청을 차단해야 한다
```

#### 4. Optional (선택적 기능)

**형식**: WHERE [조건]이면, 시스템은 [동작]할 수 있다

**예시**:
```markdown
- WHERE 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급할 수 있다
- WHERE 프리미엄 사용자이면, 시스템은 5GB 이상의 파일 업로드를 허용할 수 있다
- WHERE 쿠폰 코드가 입력되면, 시스템은 할인을 적용할 수 있다
```

#### 5. Constraints (제약사항)

**형식**: IF [조건]이면, 시스템은 [제약]해야 한다

**예시**:
```markdown
- IF 잘못된 토큰이 제공되면, 시스템은 접근을 거부해야 한다
- IF 파일 크기가 100MB를 초과하면, 시스템은 업로드를 거부해야 한다
- IF 일일 결제 시도가 3회를 초과하면, 시스템은 계정을 일시 정지해야 한다
```

### EARS 작성 체크리스트

- [ ] 모든 요구사항이 5가지 구문 중 하나에 해당하는가?
- [ ] 주체가 명확한가? (시스템, 사용자 등)
- [ ] 측정 가능한 기준이 있는가? (숫자, 시간, 상태)
- [ ] 조건과 동작이 명확히 구분되는가?
- [ ] 모호한 표현(예: "빠르게", "사용자 친화적")을 피했는가?

**더 자세한 EARS 가이드**: [EARS 요구사항 작성 가이드](../concepts/ears-guide.md)

---

## Personal vs Team 모드

### Personal 모드

**특징**:
- 개인 프로젝트 또는 소규모 팀
- 로컬 파일 기반 SPEC 관리
- GitHub Issue/PR 없이 진행

**산출물**:
```
.moai/specs/SPEC-{ID}/
├── spec.md           # EARS 명세
├── plan.md           # 구현 계획
└── acceptance.md     # 수락 기준
```

**Git 워크플로우**:
```bash
# 1. SPEC 작성
/alfred:1-spec "새 기능"
→ feature/SPEC-{ID} 브랜치 생성
→ 커밋: "🔴 RED: SPEC-{ID} 명세 작성"

# 2. TDD 구현
/alfred:2-build SPEC-{ID}
→ RED, GREEN, REFACTOR 커밋

# 3. 문서 동기화 + 로컬 머지
/alfred:3-sync
→ 문서 동기화
→ develop 브랜치로 머지
→ feature 브랜치 삭제
```

### Team 모드

**특징**:
- 팀 협업 프로젝트
- GitHub Issue/PR 기반 워크플로우
- Draft PR → Ready → Auto-merge

**산출물**:
```
.moai/specs/SPEC-{ID}/
├── spec.md           # EARS 명세
├── plan.md           # 구현 계획
└── acceptance.md     # 수락 기준

+ GitHub Issue: "[SPEC-{ID}] 기능명"
+ Draft PR: "feature/SPEC-{ID} → develop"
```

**Git 워크플로우**:
```bash
# 1. SPEC 작성
/alfred:1-spec "새 기능"
→ GitHub Issue 생성 (SPEC 본문 포함)
→ feature/SPEC-{ID} 브랜치 생성
→ Draft PR 생성

# 2. TDD 구현
/alfred:2-build SPEC-{ID}
→ RED, GREEN, REFACTOR 커밋

# 3. 문서 동기화 + 자동 머지
/alfred:3-sync --auto-merge
→ 문서 동기화
→ PR Ready 전환
→ CI/CD 확인
→ PR 자동 머지 (squash)
→ develop 체크아웃
```

### 모드 선택 가이드

| 항목 | Personal | Team |
|------|----------|------|
| **프로젝트 규모** | 개인/소규모 | 팀 협업 |
| **Issue 트래킹** | 로컬 파일 | GitHub Issue |
| **PR 리뷰** | 불필요 | 필요 (Draft → Ready) |
| **CI/CD** | 선택적 | 필수 |
| **자동 머지** | 로컬 머지 | PR 자동 머지 |
| **권장 대상** | 개인, 프로토타입 | 팀, 프로덕션 |

---

## 실전 시나리오

### 시나리오 1: 새 프로젝트 시작

**상황**: 처음 MoAI-ADK를 도입한 프로젝트

**단계**:

1. **프로젝트 초기화**
   ```bash
   moai init .
   /alfred:8-project
   ```

2. **product.md 작성**
   - 비즈니스 요구사항 명시
   - 주요 기능 목록
   - 사용자 스토리

3. **SPEC 작성**
   ```bash
   /alfred:1-spec
   # Alfred가 product.md를 분석하여 후보 제안
   ```

4. **첫 번째 SPEC 선택**
   ```
   "AUTH-001 진행"
   ```

### 시나리오 2: 기존 기능 확장

**상황**: 이미 AUTH-001이 완료되어 있고, 2FA 기능 추가

**단계**:

1. **새 SPEC 작성**
   ```bash
   /alfred:1-spec "2FA 이중 인증"
   ```

2. **의존성 명시**
   spec-builder가 자동으로 의존성 탐지:
   ```yaml
   depends_on:
     - AUTH-001
   ```

3. **관련 파일 참조**
   ```yaml
   scope:
     packages:
       - src/auth
     files:
       - auth-service.ts
   ```

### 시나리오 3: 도메인 복합 SPEC

**상황**: 결제 시스템 리팩토링 + 보안 강화

**단계**:

1. **복합 도메인 ID 사용**
   ```bash
   /alfred:1-spec "결제 시스템 리팩토링 + 보안 강화"
   # ID 제안: PAYMENT-SECURITY-001
   ```

2. **경고 확인**
   ```markdown
   ⚠️ 복합 도메인 경고
   - ID: PAYMENT-SECURITY-001
   - 하이픈 2개 사용
   - 단순화 권장: PAYMENT-001 (category: security)
   ```

3. **단순화 적용**
   ```yaml
   id: PAYMENT-001
   category: security
   labels:
     - payment
     - refactor
     - security
   ```

### 시나리오 4: SPEC 수정

**상황**: SPEC-AUTH-001의 토큰 만료시간 변경

**단계**:

1. **기존 SPEC 수정 요청**
   ```bash
   /alfred:1-spec SPEC-AUTH-001 "액세스 토큰 만료시간 30분으로 변경"
   ```

2. **HISTORY 업데이트**
   ```markdown
   ## HISTORY
   ### v0.0.2 (2025-10-12)
   - **CHANGED**: 액세스 토큰 만료시간 15분 → 30분 변경
   - **AUTHOR**: @Goos
   - **REASON**: 사용자 경험 개선 (재로그인 빈도 감소)
   - **REVIEW**: @TeamLead (승인)

   ### v0.0.1 (2025-10-11)
   - **INITIAL**: JWT 기반 인증 시스템 명세 작성
   ```

3. **updated 필드 갱신**
   ```yaml
   version: 0.0.2
   updated: 2025-10-12
   ```

### 시나리오 5: 대규모 기능 (멀티 SPEC)

**상황**: 전자상거래 결제 시스템 (3개 SPEC 필요)

**단계**:

1. **기능 분해**
   ```bash
   /alfred:1-spec "결제 시스템"
   # spec-builder가 분해 제안
   ```

2. **SPEC 후보 제안**
   ```markdown
   ## 📋 대규모 기능 분해 제안

   ### 결제 시스템 → 3개 SPEC
   1. **PAYMENT-001**: 결제 인터페이스 (HIGH)
      - 결제 요청/응답 API
      - 결제 상태 관리

   2. **PAYMENT-002**: PG 연동 (HIGH)
      - 외부 PG사 API 통합
      - 의존성: PAYMENT-001

   3. **PAYMENT-003**: 결제 이력 관리 (MEDIUM)
      - 결제 이력 저장/조회
      - 의존성: PAYMENT-001, PAYMENT-002

   ### 권장 순서
   1. PAYMENT-001 (독립 구현)
   2. PAYMENT-002 (PAYMENT-001 완료 후)
   3. PAYMENT-003 (병렬 가능)

   어떻게 진행하시겠습니까?
   - "PAYMENT-001부터 시작"
   - "전체 승인 (3개 순차 생성)"
   ```

---

## Best Practices

### 1. 프로젝트 문서 먼저 작성

**권장사항**:
```bash
# 프로젝트 초기화 및 문서 작성
/alfred:8-project

# product.md, structure.md, tech.md 완성 후
/alfred:1-spec
```

**이유**:
- spec-builder는 프로젝트 문서를 기반으로 분석
- 문서가 없으면 자동 제안 불가능
- 명확한 문서 = 명확한 SPEC

### 2. SPEC ID 중복 확인 습관

**권장사항**:
```bash
# SPEC 작성 전 수동 확인
rg "@SPEC:AUTH-001" -n .moai/specs/

# 디렉토리 확인
ls .moai/specs/ | grep "SPEC-AUTH-001"
```

**spec-builder가 자동으로 확인하지만**, 사용자가 먼저 확인하면 더 빠릅니다.

### 3. 도메인 네이밍 일관성

**권장사항**:
```
✅ 좋은 예:
- AUTH-001, AUTH-002, AUTH-003
- UPLOAD-001, UPLOAD-002
- PAYMENT-001, PAYMENT-002

❌ 나쁜 예:
- auth-001 (소문자)
- Authentication-001 (너무 긴 도메인)
- AUTH-1 (3자리 숫자 미사용)
```

**이유**:
- 일관된 네이밍으로 검색 용이
- TAG 체인 추적 간소화

### 4. EARS 구문 엄격히 준수

**권장사항**:
```markdown
✅ 좋은 예:
- WHEN 사용자가 로그인하면, 시스템은 JWT 토큰을 발급해야 한다

❌ 나쁜 예:
- 로그인 시 토큰 발급 (주체 불명확, 조건 누락)
```

**spec-builder의 검증**:
- EARS 구문 패턴 검증
- 주체 명확성 확인
- 측정 가능성 검증

### 5. HISTORY 섹션 상세 작성

**권장사항**:
```markdown
## HISTORY
### v0.0.2 (2025-10-12)
- **CHANGED**: 액세스 토큰 만료시간 15분 → 30분 변경
- **AUTHOR**: @Goos
- **REASON**: 사용자 경험 개선 (재로그인 빈도 감소)
- **REVIEW**: @TeamLead (승인)
- **RELATED**: https://github.com/org/repo/issues/42
```

**이유**:
- 변경 이력 추적성
- 의사결정 기록
- 팀원 간 컨텍스트 공유

### 6. MultiEdit 활용 (Personal 모드)

**권장사항**:
```bash
# 3개 파일 동시 생성
MultiEdit: spec.md, plan.md, acceptance.md
```

**성능 향상**:
- 3회 파일 생성 → 1회 일괄 생성
- 60% 시간 단축

### 7. 의존성 명시

**권장사항**:
```yaml
depends_on:
  - USER-001      # 사용자 관리 선행 필요
  - CONFIG-001    # 설정 시스템 필요
```

**이유**:
- 작업 순서 명확화
- 병렬 작업 가능 여부 판단
- 영향도 분석 용이

---

## 안티 패턴

### 안티 패턴 1: 프로젝트 문서 없이 SPEC 작성

❌ **잘못된 방법**:
```bash
# product.md, structure.md 없이
/alfred:1-spec "새 기능"
```

**문제점**:
- spec-builder가 컨텍스트 없이 작성
- 프로젝트 목표와 불일치 가능
- 기존 아키텍처 무시

✅ **올바른 방법**:
```bash
# 프로젝트 문서 먼저 작성
/alfred:8-project

# product.md에 비즈니스 요구사항 명시 후
/alfred:1-spec
```

### 안티 패턴 2: ID 중복 검사 생략

❌ **잘못된 방법**:
```bash
# 기존 AUTH-001이 있는데
/alfred:1-spec "새 인증 기능"
# → AUTH-001로 생성 시도 → 충돌
```

**문제점**:
- SPEC 덮어쓰기 위험
- TAG 체인 충돌
- 이전 작업 손실

✅ **올바른 방법**:
```bash
# 기존 SPEC 확인
rg "@SPEC:AUTH" -n .moai/specs/

# 다음 ID 사용
/alfred:1-spec "OAuth 연동"
# → AUTH-002로 생성
```

### 안티 패턴 3: 모호한 EARS 구문

❌ **잘못된 예**:
```markdown
- 시스템은 빠르게 응답해야 한다
- 시스템은 사용자 친화적이어야 한다
- 시스템은 안전해야 한다
```

**문제점**:
- 측정 불가능
- 검증 불가능
- 구현 모호

✅ **올바른 예**:
```markdown
- WHEN 사용자가 로그인 버튼을 클릭하면, 시스템은 2초 이내에 응답해야 한다
- 시스템은 주요 기능에 3번의 클릭 이내로 접근할 수 있어야 한다
- 시스템은 HTTPS를 사용하여 모든 데이터 전송을 암호화해야 한다
```

### 안티 패턴 4: 시간 예측 포함

❌ **잘못된 예**:
```markdown
## 구현 계획
- 1주일: JWT 토큰 생성
- 3일: 로그인 엔드포인트
- 2일: 토큰 검증 미들웨어
```

**문제점**:
- 예측 불가능성
- TRUST 원칙 위반 (Trackable)
- 불필요한 압박

✅ **올바른 예**:
```markdown
## 구현 계획 (우선순위 기반)

### 1차 목표 (필수)
- JWT 토큰 생성
- 로그인 엔드포인트
- 토큰 검증 미들웨어

### 2차 목표 (보안 강화)
- 비밀번호 해싱
- Rate limiting
```

### 안티 패턴 5: Git 작업 직접 수행

❌ **잘못된 방법**:
```markdown
# spec-builder가 직접 Git 작업
git checkout -b feature/SPEC-AUTH-001
git add .moai/specs/SPEC-AUTH-001/
git commit -m "SPEC 작성"
```

**문제점**:
- 단일 책임 원칙 위반
- git-manager 역할 침범
- Alfred 오케스트레이션 무시

✅ **올바른 방법**:
```markdown
# spec-builder는 SPEC 작성만
# Alfred가 git-manager 호출
```

### 안티 패턴 6: 하이픈 3개 이상 도메인

❌ **잘못된 예**:
```
PAYMENT-REFACTOR-SECURITY-FIX-001
```

**문제점**:
- 복잡성 증가
- 검색 어려움
- 가독성 저하

✅ **올바른 예**:
```yaml
id: PAYMENT-001
category: security
labels:
  - payment
  - refactor
  - fix
```

---

## 문제 해결 가이드

### 문제 1: "프로젝트가 초기화되지 않았습니다"

**증상**:
```bash
$ /alfred:1-spec
❌ 프로젝트가 초기화되지 않았습니다
  → moai init . 실행 후 재시도
```

**원인**:
- `.moai/` 디렉토리 없음
- `config.json` 없음

**해결**:
```bash
# 프로젝트 초기화
moai init .

# 또는 Alfred로 초기화
/alfred:8-project
```

### 문제 2: "product.md를 찾을 수 없습니다"

**증상**:
```bash
⚠️ product.md를 찾을 수 없습니다
  → 자동 제안 기능을 사용할 수 없습니다
  → 수동으로 기능명을 지정하세요
```

**원인**:
- `.moai/project/product.md` 없음
- 프로젝트 문서 미작성

**해결**:
```bash
# 프로젝트 문서 작성
/alfred:8-project

# 또는 수동 지정
/alfred:1-spec "기능명"
```

### 문제 3: ID 중복 충돌

**증상**:
```bash
❌ ID 중복: SPEC-AUTH-001이 이미 존재합니다
  → .moai/specs/SPEC-AUTH-001/spec.md
  → AUTH-002 사용 또는 기존 SPEC 수정
```

**원인**:
- 동일 ID로 SPEC 생성 시도

**해결**:
```bash
# 옵션 1: 다음 ID 사용
/alfred:1-spec "새 인증 기능"
# → AUTH-002로 생성

# 옵션 2: 기존 SPEC 수정
/alfred:1-spec SPEC-AUTH-001 "수정 내용"
```

### 문제 4: MultiEdit 실패 (Personal 모드)

**증상**:
```bash
❌ 파일 생성 실패: .moai/specs/SPEC-AUTH-001/
  → 디렉토리가 이미 존재합니다
```

**원인**:
- 디렉토리 사전 생성
- 권한 문제

**해결**:
```bash
# 기존 디렉토리 삭제
rm -rf .moai/specs/SPEC-AUTH-001/

# 또는 다른 ID 사용
/alfred:1-spec "새 기능"
```

### 문제 5: EARS 구문 검증 실패

**증상**:
```bash
⚠️ EARS 구문 검증 실패
- "시스템은 빠르게 동작해야 한다" (측정 불가능)
- "사용자 친화적이어야 한다" (모호함)
```

**원인**:
- 모호한 표현 사용
- 측정 기준 누락

**해결**:
```markdown
# 수정 전
- 시스템은 빠르게 동작해야 한다

# 수정 후
- WHEN 사용자가 로그인 버튼을 클릭하면, 시스템은 2초 이내에 응답해야 한다
```

### 문제 6: Team 모드 GitHub Issue 생성 실패

**증상**:
```bash
❌ GitHub Issue 생성 실패
  → gh: command not found
```

**원인**:
- GitHub CLI 미설치
- 인증 실패

**해결**:
```bash
# GitHub CLI 설치
brew install gh  # macOS
sudo apt install gh  # Linux

# 인증
gh auth login

# 재시도
/alfred:1-spec "새 기능"
```

### 문제 7: 디렉토리명 형식 오류

**증상**:
```bash
❌ 디렉토리명 형식 오류
  → .moai/specs/AUTH-001/ (잘못됨)
  → .moai/specs/SPEC-AUTH-001/ (올바름)
```

**원인**:
- `SPEC-` 접두어 누락

**해결**:
```bash
# 잘못된 디렉토리 삭제
rm -rf .moai/specs/AUTH-001/

# spec-builder가 자동으로 올바른 형식 사용
/alfred:1-spec "새 기능"
# → .moai/specs/SPEC-AUTH-001/ 생성
```

---

## 관련 문서

### 핵심 개념
- **[EARS 요구사항 작성 가이드](../concepts/ears-guide.md)** - EARS 5가지 구문 상세 설명
- **[SPEC-First TDD 워크플로우](../concepts/spec-first-tdd.md)** - SPEC-First 개발 방법론
- **[TAG 시스템 가이드](../concepts/tag-system.md)** - @TAG 추적성 시스템
- **[TRUST 원칙](../concepts/trust-principles.md)** - 품질 5대 원칙

### 워크플로우
- **[Stage 1: SPEC Writing](/guides/workflow/1-spec)** - SPEC 작성 단계 상세
- **[Stage 2: TDD Implementation](/guides/workflow/2-build)** - TDD 구현 단계
- **[Stage 3: Document Sync](/guides/workflow/3-sync)** - 문서 동기화 단계

### 에이전트
- **[Alfred Agent Ecosystem](./overview.md)** - 9개 에이전트 생태계 개요
- **[code-builder Agent](./code-builder.md)** - TDD 구현 에이전트
- **[doc-syncer Agent](./doc-syncer.md)** - 문서 동기화 에이전트

### 메타데이터 표준
- **[SPEC 메타데이터 가이드](../../.moai/memory/spec-metadata.md)** - 필수/선택 필드 16개 상세
- **[개발 가이드](../../.moai/memory/development-guide.md)** - SPEC-First TDD 통합 가드레일

---

## 부록: SPEC 메타데이터 빠른 참조

### 필수 필드 (7개)

| 필드 | 타입 | 형식 | 예시 |
|------|------|------|------|
| `id` | string | `<DOMAIN>-<NUMBER>` | `AUTH-001` |
| `version` | string | `MAJOR.MINOR.PATCH` | `0.0.1` |
| `status` | enum | draft\|active\|completed\|deprecated | `draft` |
| `created` | date | `YYYY-MM-DD` | `2025-10-11` |
| `updated` | date | `YYYY-MM-DD` | `2025-10-11` |
| `author` | string | `@{GitHub ID}` | `@Goos` |
| `priority` | enum | critical\|high\|medium\|low | `high` |

### 선택 필드 (9개)

| 필드 | 타입 | 예시 |
|------|------|------|
| `category` | enum | `feature`, `bugfix`, `security` |
| `labels` | array | `["authentication", "jwt"]` |
| `depends_on` | array | `["USER-001"]` |
| `blocks` | array | `["AUTH-002"]` |
| `related_specs` | array | `["TOKEN-002"]` |
| `related_issue` | string | GitHub Issue URL |
| `scope.packages` | array | `["src/auth"]` |
| `scope.files` | array | `["auth-service.ts"]` |

### HISTORY 변경 유형 태그

| 태그 | 의미 | 버전 영향 |
|------|------|----------|
| `INITIAL` | 최초 작성 | v1.0.0 |
| `ADDED` | 새 기능 추가 | Minor ↑ |
| `CHANGED` | 내용 수정 | Patch ↑ |
| `FIXED` | 버그 수정 | Patch ↑ |
| `REMOVED` | 기능 제거 | Major ↑ |
| `BREAKING` | 하위 호환 깨짐 | Major ↑ |
| `DEPRECATED` | 향후 제거 예정 | - |

---

## 부록: 도메인 예시

### 일반적인 도메인

| 도메인 | 설명 | 예시 |
|--------|------|------|
| `AUTH` | 인증/인가 | `AUTH-001`: JWT 인증 |
| `USER` | 사용자 관리 | `USER-001`: 회원가입 |
| `UPLOAD` | 파일 업로드 | `UPLOAD-001`: 멀티파트 업로드 |
| `PAYMENT` | 결제 처리 | `PAYMENT-001`: PG 연동 |
| `NOTIFICATION` | 알림 시스템 | `NOTIFICATION-001`: 실시간 알림 |
| `SEARCH` | 검색 기능 | `SEARCH-001`: 전문 검색 |
| `ADMIN` | 관리자 기능 | `ADMIN-001`: 대시보드 |

### 기술적 도메인

| 도메인 | 설명 | 예시 |
|--------|------|------|
| `REFACTOR` | 리팩토링 | `REFACTOR-001`: 인증 로직 개선 |
| `SECURITY` | 보안 강화 | `SECURITY-001`: XSS 방어 |
| `PERF` | 성능 최적화 | `PERF-001`: 쿼리 최적화 |
| `DOCS` | 문서화 | `DOCS-001`: API 문서 작성 |
| `TEST` | 테스트 개선 | `TEST-001`: E2E 테스트 추가 |

---

<div style="text-align: center; margin-top: 40px; padding: 20px; background-color: #f8f9fa; border-radius: 8px;">
  <h3>🏗️ spec-builder: 완벽한 SPEC의 시작</h3>
  <p><strong>"명세 없으면 코드 없다"</strong></p>
  <p>모든 훌륭한 개발은 명확한 SPEC에서 시작됩니다.</p>
  <p>spec-builder가 당신의 아이디어를 체계적인 명세로 변환합니다.</p>
</div>

---

**작성일**: 2025-10-11
**버전**: v1.0.0
**TAG**: @CODE:DOCS-002
**작성자**: Alfred (MoAI SuperAgent)
