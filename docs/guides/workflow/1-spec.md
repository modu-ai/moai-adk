# /alfred:1-spec

EARS 방식의 체계적인 요구사항 명세를 작성합니다.

## Overview

SPEC Writing은 MoAI-ADK 3단계 워크플로우의 첫 번째 단계입니다. **"명세 없이는 코드 없음"** 원칙을 따라 모든 개발은 SPEC 작성부터 시작합니다.

### 담당 에이전트

- **spec-builder** 🏗️: 시스템 아키텍트
- **역할**: EARS 방식 요구사항 작성, SPEC 문서 생성, Git 브랜치 생성

---

## When to Use

다음과 같은 경우 `/alfred:1-spec`을 사용합니다:

- ✅ 새로운 기능을 추가할 때
- ✅ 기존 기능을 수정/개선할 때
- ✅ 버그 수정 전 원인 분석이 필요할 때
- ✅ 아키텍처 변경을 계획할 때

---

## Command Syntax

### Basic Usage

```bash
/alfred:1-spec "기능 설명"
```

### Examples

```bash
# 단일 기능
/alfred:1-spec "사용자 로그인 기능"

# 복합 기능 (여러 제목)
/alfred:1-spec "로그인" "로그아웃" "세션 관리"

# 기존 SPEC 수정
/alfred:1-spec AUTH-001 "토큰 만료 시간 변경"
```

---

## Workflow (2단계)

### Phase 1: 분석 및 계획 수립

Alfred가 다음 작업을 수행합니다:

1. **프로젝트 상태 분석**

   ```bash
   # Git 상태 확인
   git status
   git branch

   # 기존 SPEC 조회
   ls .moai/specs/
   ```

2. **SPEC ID 생성**
   - 도메인 추출: "사용자 로그인" → `AUTH`
   - 번호 할당: `AUTH-001`, `AUTH-002`, ...
   - 중복 확인: `rg "@SPEC:AUTH-001" -n`

3. **계획 보고서 생성**

   ```markdown
   📋 SPEC 작성 계획

   다음 SPEC을 작성합니다:
   - SPEC ID: AUTH-001
   - 제목: 사용자 로그인 기능
   - 브랜치: feature/SPEC-AUTH-001
   - Draft PR: 생성 예정

   진행하시겠습니까? (진행/수정/중단)
   ```

4. **사용자 확인 대기**
   - **"진행"**: Phase 2로 이동
   - **"수정 [내용]"**: 계획 재수립
   - **"중단"**: 작업 취소

### Phase 2: SPEC 문서 작성 및 Git 작업

사용자가 "진행"하면 Alfred가 실행합니다:

1. **Git 브랜치 생성**

   ```bash
   # Team mode
   git checkout develop
   git pull origin develop
   git checkout -b feature/SPEC-AUTH-001

   # Personal mode
   git checkout main
   git pull origin main
   git checkout -b feature/SPEC-AUTH-001
   ```

2. **SPEC 문서 생성**

   **파일**: `.moai/specs/SPEC-AUTH-001/spec.md`

   ```markdown
   ---
   id: AUTH-001
   version: 0.0.1
   status: draft
   created: 2025-10-11
   updated: 2025-10-11
   author: @YourName
   priority: high
   category: feature
   ---

   # @SPEC:AUTH-001: 사용자 로그인 기능

   ## HISTORY
   ### v0.0.1 (2025-10-11)
   - **INITIAL**: 사용자 로그인 기능 명세 작성

   ## Overview
   사용자가 이메일과 비밀번호로 로그인할 수 있는 기능을 제공합니다.

   ## EARS Requirements

   ### Ubiquitous Requirements
   - 시스템은 사용자 로그인 기능을 제공해야 한다

   ### Event-driven Requirements
   - WHEN 사용자가 유효한 자격증명을 입력하면, 시스템은 JWT 토큰을 발급해야 한다
   - WHEN 토큰이 만료되면, 시스템은 401 에러를 반환해야 한다

   ### State-driven Requirements
   - WHILE 사용자가 로그인 상태일 때, 시스템은 보호된 리소스 접근을 허용해야 한다

   ### Optional Features
   - WHERE 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급할 수 있다

   ### Constraints
   - IF 잘못된 자격증명이 제공되면, 시스템은 접근을 거부해야 한다
   - 비밀번호는 bcrypt로 해시되어야 한다
   - 액세스 토큰 만료시간은 15분을 초과하지 않아야 한다

   ## Acceptance Criteria
   - [ ] 이메일과 비밀번호 입력 폼이 제공된다
   - [ ] 유효한 자격증명으로 로그인 시 JWT 토큰이 발급된다
   - [ ] 잘못된 자격증명으로 로그인 시 에러 메시지가 표시된다
   - [ ] 토큰 만료 시 자동으로 로그아웃된다

   ## Technical Design

   ### API Endpoints
   ```typescript
   POST /api/auth/login
   Request: { email: string, password: string }
   Response: { token: string, refreshToken: string, expiresIn: number }
   ```

   ### Data Model

   ```typescript
   interface User {
     id: string
     email: string
     passwordHash: string
     createdAt: Date
     updatedAt: Date
   }

   interface Session {
     userId: string
     token: string
     refreshToken: string
     expiresAt: Date
   }
   ```

   ### Dependencies

   - bcryptjs: 비밀번호 해싱
   - jsonwebtoken: JWT 생성/검증
   - express-validator: 입력 검증

   ## Test Plan

   - Unit tests: 로그인 로직, 토큰 생성
   - Integration tests: API 엔드포인트
   - E2E tests: 로그인 플로우

   ## Related SPECs

   - SPEC-AUTH-002: 사용자 등록 기능
   - SPEC-AUTH-003: 비밀번호 재설정

   ```

3. **Git 커밋 (Locale 기반)**

   `.moai/config.json`의 `locale` 설정에 따라 커밋 메시지 생성:

   ```bash
   # 한국어 (ko)
   git add .moai/specs/SPEC-AUTH-001/
   git commit -m "🔴 RED: SPEC-AUTH-001 명세 작성

   @TAG:SPEC-AUTH-001-spec"

   # 영어 (en)
   git commit -m "🔴 RED: SPEC-AUTH-001 specification written

   @TAG:SPEC-AUTH-001-spec"
   ```

4. **Draft PR 생성** (Team mode only)

   ```bash
   # GitHub에 푸시
   git push -u origin feature/SPEC-AUTH-001

   # Draft PR 생성
   gh pr create \
     --title "SPEC-AUTH-001: 사용자 로그인 기능" \
     --body "$(cat .moai/specs/SPEC-AUTH-001/spec.md)" \
     --draft \
     --base develop
   ```

5. **완료 보고**

   ```markdown
   ✅ SPEC 작성 완료

   생성된 파일:
   - .moai/specs/SPEC-AUTH-001/spec.md

   Git 작업:
   - 브랜치: feature/SPEC-AUTH-001
   - 커밋: 🔴 RED: SPEC-AUTH-001 명세 작성
   - PR: #42 (Draft) → develop

   다음 단계:
   /alfred:2-build AUTH-001
   ```

---

## EARS Requirements 작성 가이드

### 1. Ubiquitous (기본 요구사항)

**패턴**: "시스템은 [기능]을 제공해야 한다"

```markdown
### Ubiquitous Requirements
- 시스템은 사용자 인증 기능을 제공해야 한다
- 시스템은 데이터 암호화를 제공해야 한다
```

### 2. Event-driven (이벤트 기반)

**패턴**: "WHEN [조건]이면, 시스템은 [동작]해야 한다"

```markdown
### Event-driven Requirements
- WHEN 사용자가 로그인 버튼을 클릭하면, 시스템은 자격증명을 검증해야 한다
- WHEN 검증이 성공하면, 시스템은 홈 화면으로 리다이렉트해야 한다
```

### 3. State-driven (상태 기반)

**패턴**: "WHILE [상태]일 때, 시스템은 [동작]해야 한다"

```markdown
### State-driven Requirements
- WHILE 사용자가 로그인 상태일 때, 시스템은 개인화된 콘텐츠를 표시해야 한다
- WHILE 네트워크가 오프라인일 때, 시스템은 캐시된 데이터를 사용해야 한다
```

### 4. Optional (선택적 기능)

**패턴**: "WHERE [조건]이면, 시스템은 [동작]할 수 있다"

```markdown
### Optional Features
- WHERE 생체 인증이 가능하면, 시스템은 지문 로그인을 제공할 수 있다
- WHERE GPS가 활성화되면, 시스템은 위치 기반 서비스를 제공할 수 있다
```

### 5. Constraints (제약사항)

**패턴**: "IF [조건]이면, 시스템은 [제약]해야 한다"

```markdown
### Constraints
- IF 비밀번호 길이가 8자 미만이면, 시스템은 입력을 거부해야 한다
- 동시 세션 수는 5개를 초과하지 않아야 한다
```

---

## SPEC Metadata 필드

### 필수 필드 (7개)

| 필드 | 설명 | 예시 |
|------|------|------|
| `id` | SPEC 고유 ID | `AUTH-001` |
| `version` | Semantic Version | `0.0.1` |
| `status` | 상태 | `draft`, `active`, `deprecated` |
| `created` | 생성일 | `2025-10-11` |
| `updated` | 최종 수정일 | `2025-10-11` |
| `author` | 작성자 | `@YourName` |
| `priority` | 우선순위 | `high`, `medium`, `low` |

### 선택 필드 (9개)

| 필드 | 설명 | 예시 |
|------|------|------|
| `category` | 카테고리 | `feature`, `bugfix`, `refactor` |
| `labels` | 태그 | `authentication`, `security` |
| `depends_on` | 의존 SPEC | `AUTH-002` |
| `blocks` | 차단 SPEC | `USER-001` |
| `related_specs` | 관련 SPEC | `AUTH-003`, `SESSION-001` |
| `related_issue` | 관련 이슈 | `#42` |
| `scope` | 영향 범위 | `frontend`, `backend`, `full-stack` |

---

## Best Practices

### 1. 명확한 제목

```markdown
# ✅ Good
@SPEC:AUTH-001: JWT 기반 사용자 인증 시스템

# ❌ Bad
@SPEC:AUTH-001: 로그인
```

### 2. 구체적인 요구사항

```markdown
# ✅ Good
WHEN 사용자가 유효한 이메일과 비밀번호를 입력하면, 시스템은 15분 유효기간의 JWT 토큰을 발급해야 한다

# ❌ Bad
WHEN 로그인하면 토큰을 준다
```

### 3. 측정 가능한 Acceptance Criteria

```markdown
# ✅ Good
- [ ] 로그인 API 응답 시간 < 200ms (p95)
- [ ] 비밀번호 해싱 시간 < 100ms

# ❌ Bad
- [ ] 빠르게 동작한다
```

### 4. 기술 설계 포함

```typescript
// ✅ Good - 구체적인 타입 정의
interface LoginRequest {
  email: string  // RFC 5322 형식
  password: string  // 8-128 characters
  rememberMe?: boolean
}

// ❌ Bad - 추상적
interface LoginData {
  data: any
}
```

---

## Common Pitfalls

### ❌ SPEC 없이 구현

```bash
# Bad
직접 코드 작성 → 나중에 SPEC 추가
```

### ✅ SPEC 우선

```bash
# Good
/alfred:1-spec → /alfred:2-build → /alfred:3-sync
```

### ❌ 너무 추상적인 SPEC

```markdown
# Bad
시스템은 사용자를 관리한다
```

### ✅ 구체적인 SPEC

```markdown
# Good
시스템은 이메일 기반 사용자 등록, 로그인, 로그아웃, 비밀번호 재설정 기능을 제공해야 한다
```

---

## Next Steps

SPEC 작성이 완료되면 다음 단계로 진행합니다:

1. **[Stage 2: TDD Implementation](/guides/workflow/2-build)** - `/alfred:2-build` 실행
2. **[EARS Guide](/guides/concepts/ears-guide)** - EARS 작성법 상세
3. **[SPEC Metadata](/guides/concepts/spec-metadata)** - 메타데이터 필드 상세

---

<div style="text-align: center; margin-top: 40px;">
  <p><strong>명세 없이는 코드 없음</strong> 📋</p>
  <p>SPEC-First로 시작하세요!</p>
</div>
