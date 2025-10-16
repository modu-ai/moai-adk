---
name: MoAI ADK Learning
description: MoAI-ADK의 개념과 워크플로우를 쉽게 배우는 학습 모드
---

# MoAI ADK Learning

**대상**: MoAI-ADK를 처음 사용하는 개발자

MoAI-ADK의 핵심 개념과 3단계 워크플로우를 친절하게 설명하여 빠르게 익힐 수 있도록 돕는 학습 모드입니다.

## MoAI-ADK란?

**핵심 철학**: "명세 없으면 코드 없다, 테스트 없으면 구현 없다"

MoAI-ADK는 3가지 핵심 개념으로 구성됩니다:
1. **SPEC-First**: 코드 작성 전 명세를 먼저 작성
2. **@TAG 추적성**: 모든 코드를 SPEC과 연결
3. **TRUST 품질**: 5가지 원칙으로 코드 품질 보장

이 3가지 개념이 어떻게 작동하는지 하나씩 배워봅시다!

---

## 핵심 개념 1: SPEC-First

### SPEC이란?

**간단히 말하면**:
- 무엇을 만들지 미리 적어놓는 설계도
- 요리 레시피처럼 단계와 재료를 명확히 정의

**왜 필요한가요?**
- 개발 전 요구사항을 명확히 정리
- 팀원과 소통할 때 기준이 됨
- 나중에 변경 사항을 추적 가능
- "이 코드는 왜 만들었지?"라는 질문에 답할 수 있음

### EARS 구문: 요구사항 작성법

EARS는 요구사항을 5가지 패턴으로 작성하는 방법입니다:

#### 1. Ubiquitous (기본 기능)
```markdown
시스템은 [기능]을 제공해야 한다

예시:
- 시스템은 JWT 기반 인증을 제공해야 한다
```

#### 2. Event-driven (조건부 동작)
```markdown
WHEN [조건]이면, 시스템은 [동작]해야 한다

예시:
- WHEN 유효한 자격증명 제공 시, 시스템은 JWT 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 시스템은 401 에러를 반환해야 한다
```

#### 3. State-driven (상태 기반 동작)
```markdown
WHILE [상태]일 때, 시스템은 [동작]해야 한다

예시:
- WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 리소스 접근을 허용해야 한다
```

#### 4. Optional (선택적 기능)
```markdown
WHERE [조건]이면, 시스템은 [동작]할 수 있다

예시:
- WHERE 리프레시 토큰이 제공되면, 시스템은 새 액세스 토큰을 발급할 수 있다
```

#### 5. Constraints (제약사항)
```markdown
IF [조건]이면, 시스템은 [제약]해야 한다

예시:
- 토큰 만료시간은 15분을 초과해서는 안 된다
- 비밀번호는 최소 8자 이상이어야 한다
```

### 실제 예시: 로그인 기능 SPEC

```markdown
# @SPEC:AUTH-001: JWT 인증 시스템

## Ubiquitous Requirements (기본 기능)
- 시스템은 JWT 기반 인증을 제공해야 한다

## Event-driven Requirements (조건부 동작)
- WHEN 유효한 자격증명 제공 시, 시스템은 JWT 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 시스템은 401 에러를 반환해야 한다
- WHEN 잘못된 토큰이 제공되면, 시스템은 접근을 거부해야 한다

## State-driven Requirements (상태 기반)
- WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 리소스 접근을 허용해야 한다

## Optional Features (선택적 기능)
- WHERE 리프레시 토큰이 제공되면, 시스템은 새 액세스 토큰을 발급할 수 있다

## Constraints (제약사항)
- 액세스 토큰 만료시간은 15분을 초과해서는 안 된다
- 리프레시 토큰 만료시간은 7일을 초과해서는 안 된다
```

---

## 핵심 개념 2: @TAG 추적성

### TAG란?

**간단히 말하면**:
- 코드 조각마다 붙이는 이름표
- SPEC → TEST → CODE → DOC을 연결하는 끈
- 나중에 코드를 찾을 때 SPEC 번호로 검색 가능

**왜 TAG가 필요한가요?**
- 나중에 코드를 찾을 때 SPEC 번호로 검색 가능
- SPEC이 변경되면 어떤 코드를 수정할지 명확
- 코드 리뷰 시 "이 코드는 어떤 요구사항인가?" 즉시 파악
- 버그 발생 시 관련된 모든 파일을 빠르게 찾을 수 있음

### TAG 체계

MoAI-ADK는 4가지 TAG를 사용합니다:

```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

| TAG | 의미 | 위치 | 예시 |
|-----|------|------|------|
| `@SPEC:ID` | 요구사항 명세 | `.moai/specs/` | @SPEC:AUTH-001 |
| `@TEST:ID` | 테스트 코드 | `tests/` | @TEST:AUTH-001 |
| `@CODE:ID` | 구현 코드 | `src/` | @CODE:AUTH-001 |
| `@DOC:ID` | 문서 | `docs/` | @DOC:AUTH-001 |

### TAG ID 규칙

**형식**: `<도메인>-<3자리 숫자>`

**예시**:
- `AUTH-001`: 인증 관련 첫 번째 기능
- `USER-002`: 사용자 관련 두 번째 기능
- `PAYMENT-015`: 결제 관련 15번째 기능

**중요**: TAG ID는 한 번 할당되면 절대 변경하지 않습니다!

### 실제 예시: TAG 사용법

#### SPEC 파일 (`.moai/specs/SPEC-AUTH-001/spec.md`)
```yaml
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-16
updated: 2025-10-16
author: @YourName
priority: high
---

# @SPEC:AUTH-001: JWT 인증 시스템

[요구사항 내용...]
```

#### 테스트 파일 (`tests/auth/service.test.ts`)
```typescript
// @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

test('@TEST:AUTH-001: JWT 발급 on valid credentials', async () => {
  const service = new AuthService();
  const result = await service.authenticate('user', 'pass');
  expect(result.token).toBeDefined();
});
```

#### 구현 파일 (`src/auth/service.ts`)
```typescript
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts

export class AuthService {
  async authenticate(username: string, password: string): Promise<AuthResult> {
    // 구현
  }
}
```

#### 문서 파일 (`docs/api/auth.md`)
```markdown
# @DOC:AUTH-001: 인증 API 문서

## POST /auth/login
[API 설명...]
```

### TAG 검색하기

**특정 TAG 찾기**:
```bash
# AUTH-001 관련 모든 파일 찾기
rg "AUTH-001" -n
```

**TAG 체인 검증**:
```bash
# 모든 TAG 확인
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

---

## 핵심 개념 3: TRUST 5원칙

좋은 코드를 만드는 5가지 원칙을 비유로 설명합니다:

### 1. 🧪 Test (테스트)

**비유**: 요리하기 전에 맛을 상상하는 것

**의미**:
- 코드 작성 전에 테스트를 먼저 작성
- 테스트가 통과하면 코드가 제대로 작동하는 것

**기준**:
- 테스트 커버리지 ≥85%
- SPEC → Test → Code 순서 엄수

**예시**:
```typescript
// 먼저 테스트 작성 (RED)
test('should add two numbers', () => {
  expect(add(2, 3)).toBe(5);  // 아직 add 함수가 없어서 실패
});

// 그 다음 코드 작성 (GREEN)
function add(a: number, b: number): number {
  return a + b;  // 테스트 통과!
}
```

### 2. 📖 Readable (읽기 쉬움)

**비유**: 깔끔한 글씨로 쓰기

**의미**:
- 다른 사람이 읽어도 이해할 수 있는 코드
- 나중에 다시 봐도 이해하기 쉬운 코드

**기준**:
- 함수 ≤50줄
- 파일 ≤300줄
- 복잡도 ≤10
- 매개변수 ≤5개
- 의미 있는 이름 사용

**예시**:
```typescript
// ❌ 나쁜 예: 이해하기 어려움
function f(x, y) {
  return x + y;
}

// ✅ 좋은 예: 명확한 이름
function calculateTotal(price: number, tax: number): number {
  return price + tax;
}
```

### 3. 🎯 Unified (통일성)

**비유**: 같은 방식 사용하기

**의미**:
- 같은 패턴을 일관되게 적용
- 한 가지 방법만 익히면 모든 곳에 적용 가능

**기준**:
- SPEC 기반 아키텍처
- 타입 안전성 또는 런타임 검증
- 일관된 코딩 스타일

**예시**:
```typescript
// ✅ 통일성: 모든 API가 같은 패턴
async function getUser(id: string): Promise<User> { ... }
async function getPost(id: string): Promise<Post> { ... }
async function getComment(id: string): Promise<Comment> { ... }
```

### 4. 🔒 Secured (보안)

**비유**: 집에 나갈 때 문 잠그기

**의미**:
- 해커나 나쁜 사람이 코드를 악용하지 못하게 보호
- 사용자 데이터를 안전하게 보호

**기준**:
- 입력 검증
- SQL Injection 방어
- XSS/CSRF 방어
- 비밀번호 해싱
- 민감 데이터 보호

**예시**:
```typescript
// ❌ 나쁜 예: SQL Injection 위험
const query = `SELECT * FROM users WHERE id = '${userId}'`;

// ✅ 좋은 예: Prepared Statement 사용
const query = 'SELECT * FROM users WHERE id = ?';
const user = await db.execute(query, [userId]);

// ❌ 나쁜 예: 비밀번호 평문 저장
user.password = password;

// ✅ 좋은 예: 비밀번호 해싱
user.password = await bcrypt.hash(password, 10);
```

### 5. 🔗 Trackable (추적 가능)

**비유**: 옷장 정리할 때 상자마다 이름표 붙이기

**의미**:
- 나중에 필요한 코드를 빠르게 찾을 수 있음
- 코드 변경 이력을 추적 가능

**기준**:
- @TAG 시스템 사용
- SPEC과 코드 연결
- Git 커밋 메시지에 TAG 포함

**예시**:
```typescript
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts
export class AuthService {
  // 구현
}

// Git 커밋 메시지
// 🟢 feat(AUTH-001): implement JWT authentication
```

---

## 핵심 개념 4: Alfred와 9개 에이전트

### Alfred란?

**간단히 말하면**:
- MoAI-ADK의 중앙 오케스트레이터 (지휘자)
- 9개 전문 에이전트를 조율하여 최적의 도움 제공
- 사용자 요청을 분석하고 적절한 에이전트에게 작업 위임

**비유**: 오케스트라 지휘자처럼 여러 전문가를 조율

### 주요 에이전트 (간단 소개)

| 에이전트 | 역할 | 언제 사용 |
|---------|------|----------|
| 🏗️ spec-builder | SPEC 작성 전문가 | `/alfred:1-spec` 명령어 |
| 💎 code-builder | TDD 구현 전문가 | `/alfred:2-build` 명령어 |
| 📖 doc-syncer | 문서 동기화 전문가 | `/alfred:3-sync` 명령어 |
| 🔬 debug-helper | 디버깅 전문가 | 에러 발생 시 자동 호출 |
| ✅ trust-checker | 품질 검증 전문가 | 코드 품질 확인 시 |
| 🏷️ tag-agent | TAG 관리 전문가 | TAG 검증 시 |

**전체 에이전트 목록**: `AGENTS.md` 파일을 참조하세요

### Alfred가 작동하는 방식

```
사용자 요청
    ↓
Alfred가 요청 분석
    ↓
적절한 전문 에이전트에게 위임
    ↓
에이전트가 작업 수행
    ↓
Alfred가 결과 통합하여 보고
```

---

## 3단계 워크플로우 익히기

MoAI-ADK의 핵심은 이 3단계 워크플로우입니다:

```
/alfred:1-spec → /alfred:2-build → /alfred:3-sync
```

### 1단계: SPEC 작성 (`/alfred:1-spec`)

**무엇을 하나요?**
- 요구사항을 EARS 구문으로 작성
- `.moai/specs/SPEC-{ID}/spec.md` 파일 생성
- @SPEC:ID TAG 자동 할당
- Git 브랜치 생성 (옵션)

**사용 예시**:
```bash
/alfred:1-spec "JWT 인증 시스템"
```

**Alfred가 자동으로 수행**:
1. 중복 확인: "AUTH-001이 이미 존재하나요?"
2. SPEC 파일 생성: `.moai/specs/SPEC-AUTH-001/spec.md`
3. YAML 메타데이터 추가:
   ```yaml
   ---
   id: AUTH-001
   version: 0.0.1
   status: draft
   created: 2025-10-16
   updated: 2025-10-16
   author: @YourName
   priority: high
   ---
   ```
4. EARS 구문 템플릿 제공
5. @SPEC:AUTH-001 TAG 할당

**결과물 예시**:
```yaml
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-16
updated: 2025-10-16
author: @YourName
priority: high
---

# @SPEC:AUTH-001: JWT 인증 시스템

## Ubiquitous Requirements
- 시스템은 JWT 기반 인증을 제공해야 한다

## Event-driven Requirements
- WHEN 유효한 자격증명 제공 시, 시스템은 JWT 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 시스템은 401 에러를 반환해야 한다

## Constraints
- 토큰 만료시간은 15분을 초과해서는 안 된다
```

### 2단계: TDD 구현 (`/alfred:2-build`)

**무엇을 하나요?**
- **RED**: 실패하는 테스트 작성 (`@TEST:ID`)
- **GREEN**: 최소 구현으로 테스트 통과 (`@CODE:ID`)
- **REFACTOR**: 코드 품질 개선 (TRUST 5원칙 적용)

**사용 예시**:
```bash
/alfred:2-build AUTH-001
```

**Alfred가 자동으로 수행**:

#### 🔴 RED: 실패하는 테스트 작성
```typescript
// tests/auth/service.test.ts
// @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

test('@TEST:AUTH-001: JWT 발급 on valid credentials', async () => {
  const service = new AuthService();
  const result = await service.authenticate('user', 'pass');
  expect(result.token).toBeDefined();
  expect(result.expiresIn).toBeLessThanOrEqual(900); // 15분
});
```

**테스트 실행**: ❌ FAIL (AuthService가 아직 없음)

#### 🟢 GREEN: 최소 구현
```typescript
// src/auth/service.ts
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts

export class AuthService {
  async authenticate(username: string, password: string): Promise<AuthResult> {
    return {
      token: jwt.sign({ username }, SECRET, { expiresIn: '15m' }),
      expiresIn: 900
    };
  }
}
```

**테스트 실행**: ✅ PASS

#### ♻️ REFACTOR: 품질 개선
```typescript
// 입력 검증 추가 (Secured)
if (!username || !password) {
  throw new ValidationError('Username and password required');
}

// 함수 분리 (Readable)
private generateToken(username: string): string {
  return jwt.sign({ username }, SECRET, { expiresIn: '15m' });
}

// 복잡도 감소 (Unified)
```

**테스트 실행**: ✅ PASS
**TRUST 검증**: ✅ 모두 통과

**Git 커밋**:
```bash
1. 🔴 RED: test(AUTH-001): add failing auth service test
2. 🟢 GREEN: feat(AUTH-001): implement minimal auth service
3. ♻️ REFACTOR: refactor(AUTH-001): improve code quality per TRUST
```

### 3단계: 문서 동기화 (`/alfred:3-sync`)

**무엇을 하나요?**
- TAG 체인 검증 (@SPEC → @TEST → @CODE → @DOC)
- 고아 TAG 탐지 (SPEC 없는 CODE 등)
- Living Document 자동 생성
- PR 상태 업데이트 (Draft → Ready)

**사용 예시**:
```bash
/alfred:3-sync
```

**Alfred가 자동으로 수행**:

1. **TAG 체인 검증**:
```bash
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

2. **검증 결과**:
```
✓ @SPEC:AUTH-001 → .moai/specs/SPEC-AUTH-001.md
✓ @TEST:AUTH-001 → tests/auth/service.test.ts
✓ @CODE:AUTH-001 → src/auth/service.ts
✓ @DOC:AUTH-001 → docs/api/auth.md

TAG 체인 완전성: 100%
고아 TAG: 없음
SPEC 버전 일치: v0.0.1
```

3. **Living Document 생성**:
```markdown
# @DOC:AUTH-001: 인증 API 문서

## POST /auth/login

**요구사항**: @SPEC:AUTH-001
**구현**: @CODE:AUTH-001
**테스트**: @TEST:AUTH-001

[자동 생성된 API 문서...]
```

4. **PR 상태 전환 제안**:
```
PR #123: feature/spec-auth-001
현재 상태: Draft
제안: Ready for Review

SPEC 요구사항: ✅ 모두 충족
TDD 이력: ✅ RED → GREEN → REFACTOR
TRUST 검증: ✅ 모두 통과

PR Ready 전환? (y/n)
```

---

## 실전 예시: 간단한 계산기 만들기

3단계 워크플로우를 실제로 사용해봅시다!

### 1️⃣ SPEC 작성
```bash
/alfred:1-spec "두 숫자를 더하는 계산기"

# Alfred가 생성: .moai/specs/SPEC-CALC-001/spec.md
```

**생성된 SPEC**:
```yaml
---
id: CALC-001
version: 0.0.1
status: draft
created: 2025-10-16
updated: 2025-10-16
author: @YourName
priority: medium
---

# @SPEC:CALC-001: 계산기 - 덧셈 기능

## Ubiquitous Requirements
- 시스템은 두 숫자의 덧셈을 제공해야 한다

## Event-driven Requirements
- WHEN 두 숫자가 입력되면, 시스템은 합계를 반환해야 한다

## Constraints
- 입력은 숫자여야 한다
- 결과는 정확해야 한다
```

### 2️⃣ TDD 구현
```bash
/alfred:2-build CALC-001

# Alfred가 Red-Green-Refactor 자동 수행
```

**생성된 코드**:
```typescript
// tests/calc.test.ts
// @TEST:CALC-001 | SPEC: SPEC-CALC-001.md
test('@TEST:CALC-001: should add two numbers', () => {
  expect(add(2, 3)).toBe(5);
  expect(add(10, 20)).toBe(30);
});

// src/calc.ts
// @CODE:CALC-001 | SPEC: SPEC-CALC-001.md | TEST: tests/calc.test.ts
export function add(a: number, b: number): number {
  if (typeof a !== 'number' || typeof b !== 'number') {
    throw new TypeError('Both arguments must be numbers');
  }
  return a + b;
}
```

### 3️⃣ 문서 동기화
```bash
/alfred:3-sync

# TAG 검증 및 문서 생성
```

**결과**:
```
✓ @SPEC:CALC-001
✓ @TEST:CALC-001
✓ @CODE:CALC-001
✓ @DOC:CALC-001

완성! 3개 명령어로 SPEC → TEST → CODE → DOC 완성!
```

---

## 다음 단계

### MoAI-ADK를 익혔다면

이제 다른 스타일로 전환하여 실무에 적용해봅시다:

| 다음 목표 | 권장 스타일 | 전환 명령어 |
|----------|------------|------------|
| 실무 프로젝트 개발 | **agentic-coding** | `/output-style agentic-coding` |
| 새로운 언어/프레임워크 학습 | **study-with-alfred** | `/output-style study-with-alfred` |

### 더 배우기

**상세 가이드**:
- `.moai/memory/development-guide.md` - 개발 상세 가이드
- `.moai/project/structure.md` - 프로젝트 구조
- `.moai/memory/spec-metadata.md` - SPEC 메타데이터 표준

**에이전트 문서**:
- `AGENTS.md` - 9개 전문 에이전트 상세 설명

---

**MoAI ADK Learning**: MoAI-ADK의 핵심 개념과 워크플로우를 쉽게 배워 빠르게 익힐 수 있도록 돕는 친절한 학습 모드입니다.
