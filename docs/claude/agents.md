---
title: Claude Code 에이전트 가이드
description: 7개 전문 에이전트 활용법
---

# Claude Code 에이전트 가이드

MoAI-ADK는 SPEC-First TDD 개발을 자동화하는 **7개 전문 에이전트**를 제공합니다. 각 에이전트는 특정 역할에 전문화되어 있으며, 3단계 워크플로우를 지원합니다.

## 에이전트 개요

### 전체 에이전트 목록

| 에이전트 | 역할 | 주요 기능 | 자동화 수준 |
|---------|------|-----------|-------------|
| **spec-builder** | SPEC 작성 | EARS 요구사항 생성 | 사용자 확인 후 브랜치 |
| **code-builder** | TDD 구현 | Red-Green-Refactor | 자동 (범용 언어) |
| **doc-syncer** | 문서 동기화 | TAG 검증, PR 전환 | 사용자 확인 후 머지 |
| **cc-manager** | Claude Code 설정 | 권한 최적화 | 자동 |
| **debug-helper** | 디버깅 지원 | 시스템 진단 | 온디맨드 |
| **git-manager** | Git 작업 | 브랜치/커밋/머지 | 사용자 확인 필수 |
| **trust-checker** | 품질 검증 | TRUST 5원칙 검사 | 자동 |

### 호출 방법

```bash
# 기본 호출
@agent-{name} "요청 내용"

# 예시
@agent-spec-builder "사용자 인증 기능 SPEC 작성"
@agent-code-builder "SPEC-001 구현 계획 수립"
@agent-debug-helper "TypeError 오류 분석"
```

## 1. spec-builder

### 역할

**SPEC 작성 전담 에이전트**

- EARS 방법론 기반 요구사항 작성
- @TAG Catalog 자동 생성
- 브랜치 생성 (사용자 확인 후)

### 주요 기능

#### 1. 새 SPEC 작성

```bash
@agent-spec-builder "사용자 이메일/비밀번호 인증 기능"
```

자동 생성되는 SPEC:

```markdown
# SPEC-AUTH-001: 사용자 이메일/비밀번호 인증

## Metadata
- **ID**: SPEC-AUTH-001
- **생성일**: 2024-01-15
- **상태**: Draft

## Background
사용자가 시스템에 안전하게 접근하기 위한 인증 시스템이 필요합니다.

## Requirements

### Ubiquitous Requirements
- 시스템은 이메일/비밀번호 기반 인증을 제공해야 한다

### Event-driven Requirements
- WHEN 유효한 자격증명으로 로그인하면, JWT 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 401 에러를 반환해야 한다

### Constraints
- 토큰 만료시간은 15분을 초과하지 않아야 한다
- 비밀번호는 bcrypt로 해싱해야 한다

## @TAG Catalog
| Chain | TAG | 설명 | 연관 산출물 |
|-------|-----|------|------------|
| Primary | @REQ:AUTH-001 | 인증 요구사항 | 이 문서 |
| Primary | @DESIGN:AUTH-001 | 인증 설계 | design/ |
| Primary | @TASK:AUTH-001 | 인증 구현 | src/auth/ |
| Primary | @TEST:AUTH-001 | 인증 테스트 | __tests__/auth/ |
```

#### 2. 기존 SPEC 수정

```bash
@agent-spec-builder "SPEC-AUTH-001에 2FA 요구사항 추가"
```

자동 업데이트:

```markdown
### Additional Requirements
- WHERE 2FA 활성화 시, 시스템은 OTP 검증을 요구할 수 있다
```

#### 3. 브랜치 생성 (사용자 확인)

SPEC 작성 완료 후:

```
에이전트: "SPEC-AUTH-001 작성이 완료되었습니다.
          feature/spec-auth-001-authentication 브랜치를 생성하시겠습니까? (y/n)"

사용자: y

에이전트: "브랜치 생성 완료. Git 브랜치: feature/spec-auth-001-authentication"
```

### 사용 예시

#### 기본 사용

```bash
# 단일 SPEC
@agent-spec-builder "결제 시스템 구현"

# 복수 SPEC
@agent-spec-builder "회원가입" "로그인" "비밀번호 재설정"

# 상세 컨텍스트 제공
@agent-spec-builder "OAuth2 소셜 로그인 (Google, GitHub, Apple)"
```

#### 고급 사용

```bash
# SPEC 수정
@agent-spec-builder "SPEC-PAYMENT-002 결제 금액 제한 추가"

# SPEC 검증
@agent-spec-builder "SPEC-PAYMENT-002 EARS 구문 검증"

# SPEC 템플릿 커스터마이징
@agent-spec-builder "API 전용 SPEC 템플릿으로 작성"
```

## 2. code-builder

### 역할

**TDD 구현 전담 에이전트**

- 범용 언어 지원 (TypeScript, Python, Java, Go, Rust 등)
- Red-Green-Refactor 자동화
- @TAG 자동 삽입

### 주요 기능

#### 1. 분석 단계 (계획 수립)

```bash
@agent-code-builder "SPEC-AUTH-001 분석 및 구현 계획 수립"
```

출력:

```markdown
## 구현 계획: SPEC-AUTH-001

### 분석 결과
- 언어: TypeScript
- 테스트 도구: Vitest
- 의존성: bcrypt, jsonwebtoken

### 구현 순서
1. UserRepository 인터페이스 정의
2. PasswordService 구현 (bcrypt)
3. TokenService 구현 (JWT)
4. AuthService 구현
5. 통합 테스트

### 예상 소요 시간
- RED: 30분
- GREEN: 1시간
- REFACTOR: 30분
총: 2시간

승인하시겠습니까? (y/n)
```

#### 2. TDD 구현 (사용자 승인 후)

```bash
사용자: y

@agent-code-builder "승인된 계획으로 구현 시작"
```

자동 생성:

1. **RED Phase**: 실패하는 테스트

```typescript
// __tests__/auth/service.test.ts
// @TEST:AUTH-001
describe('AuthService', () => {
  test('should authenticate valid user', async () => {
    // 실패하는 테스트
  });
});
```

2. **GREEN Phase**: 최소 구현

```typescript
// src/auth/service.ts
// @TASK:AUTH-001
export class AuthService {
  async authenticate(email: string, password: string) {
    // 최소 구현
  }
}
```

3. **REFACTOR Phase**: 품질 개선

```typescript
// @FEATURE:AUTH-001 | Chain: @REQ → @DESIGN → @TASK → @TEST
export class AuthService {
  constructor(
    private userRepository: UserRepository,
    private tokenService: TokenService
  ) {}
  // 리팩토링된 구현
}
```

### 언어별 자동 감지

#### TypeScript 프로젝트

```bash
@agent-code-builder "SPEC-001 구현"

# 자동 감지:
# - package.json, tsconfig.json 발견
# - Vitest + Biome 사용
# - strict 타입 체크
```

#### Python 프로젝트

```bash
@agent-code-builder "SPEC-001 구현"

# 자동 감지:
# - requirements.txt, pyproject.toml 발견
# - pytest + mypy + black + ruff 사용
# - Type hints 적용
```

### 사용 예시

```bash
# 기본 TDD 구현
@agent-code-builder "SPEC-AUTH-001 TDD 구현"

# 특정 단계만
@agent-code-builder "SPEC-AUTH-001 RED 단계만 작성"
@agent-code-builder "SPEC-AUTH-001 REFACTOR 수행"

# 여러 SPEC 일괄 구현
@agent-code-builder "SPEC-001 SPEC-002 SPEC-003 구현"
```

## 3. doc-syncer

### 역할

**문서 동기화 전담 에이전트**

- TAG 체인 검증
- Living Document 업데이트
- PR 상태 전환 (사용자 확인 후)

### 주요 기능

#### 1. 문서 동기화

```bash
@agent-doc-syncer "전체 문서 동기화"
```

실행 항목:

```
✓ TAG 체인 검증
  - 코드 전체 스캔
  - Primary Chain 완결성 확인
  - 고아 TAG 감지

✓ Living Document 업데이트
  - README.md 갱신
  - API 문서 자동 생성
  - CHANGELOG.md 업데이트

✓ sync-report.md 생성
  - TAG 통계
  - 불완전한 체인 목록
  - 권장 사항
```

#### 2. TAG 체인 검증

```bash
@agent-doc-syncer "TAG 무결성 검사"
```

검증 리포트:

```markdown
# TAG 검증 리포트

## 완결된 체인 (32개)
✅ AUTH-001: REQ → DESIGN → TASK → TEST
✅ PAYMENT-002: REQ → DESIGN → TASK → TEST

## 불완전한 체인 (2개)
⚠️ NOTIFICATION-004: @TEST 누락
  - 파일: src/notification/service.ts
  - 권장: __tests__/notification/service.test.ts 생성

⚠️ REPORT-005: @DESIGN 누락
  - 파일: .moai/specs/SPEC-005/
  - 권장: design/ 폴더에 설계 문서 추가

## 고아 TAG (0개)
```

#### 3. PR 상태 전환 (사용자 확인)

```bash
@agent-doc-syncer "PR 상태 변경"

# 출력:
문서 동기화가 완료되었습니다.
다음 작업을 수행하시겠습니까?

1. PR을 Draft → Ready로 전환
2. develop 브랜치로 머지 요청

진행하시겠습니까? (y/n)
```

### 사용 예시

```bash
# 전체 동기화
@agent-doc-syncer "문서 동기화 수행"

# 특정 문서만
@agent-doc-syncer "API 문서만 갱신"
@agent-doc-syncer "README 업데이트"

# TAG 검증만
@agent-doc-syncer "TAG 체인 검증만 수행"
```

## 4. cc-manager

### 역할

**Claude Code 설정 전담 에이전트**

- 권한 최적화
- 설정 표준화
- 훅 시스템 관리

### 주요 기능

#### 1. Claude Code 설정 최적화

```bash
@agent-cc-manager "Claude Code 설정 최적화"
```

최적화 항목:

```json
// .claude/settings.json
{
  "version": "1.0",
  "project": "my-project",
  "agents": {
    "enabled": true,
    "path": "agents/moai"
  },
  "hooks": {
    "enabled": true,
    "path": "hooks/moai",
    "autoRun": true
  },
  "permissions": {
    "fileWrite": "confirm",      // 파일 쓰기 확인 필요
    "gitCommit": "confirm",      // Git 커밋 확인 필요
    "branchCreate": "confirm"    // 브랜치 생성 확인 필요
  }
}
```

#### 2. 권한 설정

```bash
@agent-cc-manager "파일 쓰기 권한 조정"
```

#### 3. 훅 시스템 관리

```bash
@agent-cc-manager "pre-write-guard 훅 활성화"
```

## 5. debug-helper

### 역할

**온디맨드 디버깅 에이전트**

- 오류 분석
- 시스템 진단
- TAG 무결성 검증
- 개발 가이드 검사

### 주요 기능

#### 1. 오류 분석

```bash
@agent-debug-helper "TypeError: Cannot read property 'name' of undefined"
```

분석 결과:

```markdown
## 오류 분석

### 오류 유형
TypeError: null/undefined 참조

### 발생 위치
src/user/service.ts:45

### 원인
user 객체가 null일 가능성

### 해결 방법
1. Optional chaining 사용: user?.name
2. Guard clause 추가
3. Type narrowing 적용

### 수정 코드
```typescript
// Before
function getUserName(user: User) {
  return user.name; // ❌
}

// After
function getUserName(user: User | null) {
  if (!user) {
    throw new ValidationError('User not found');
  }
  return user.name; // ✅
}
```
```

#### 2. 시스템 진단

```bash
@agent-debug-helper "시스템 진단 수행"
```

진단 리포트:

```
✓ System Diagnosis
  ✓ Node.js: v18.19.0
  ✓ Git: v2.39.2
  ✓ npm: v10.2.3
  ✓ TypeScript: v5.9.2
  ✓ Vitest: v3.2.4
  ⚠️ Git LFS: 미설치 (선택사항)

✓ Language Detection
  - TypeScript: 65%
  - Python: 25%
  - Go: 10%

✓ Project Health
  - Test Coverage: 92.5%
  - TRUST Score: 92%
  - Code Quality: Excellent

⚠️ Warnings
  - 3개 파일이 300 LOC 초과
  - 2개 함수가 50 LOC 초과
```

#### 3. TAG 무결성 검증

```bash
@agent-debug-helper "TAG 체인 검증"
```

#### 4. 개발 가이드 검사

```bash
@agent-debug-helper "개발 가이드 준수 확인"
```

TRUST 5원칙 검증:

```markdown
## 개발 가이드 준수 검사

### TRUST 준수율: 92%

✓ Test First: 80%
✓ Readable: 100%
✓ Unified: 90%
✓ Secured: 100%
✓ Trackable: 90%

### 권장 사항
- Red-Green-Refactor 사이클 강화
- TAG 체인 완결성 개선
```

## 6. git-manager

### 역할

**Git 작업 전담 에이전트**

- 브랜치/머지 관리 (사용자 확인 필수)
- 커밋 자동화
- PR 생성 및 관리

### 주요 기능

#### 1. 브랜치 생성 (사용자 확인)

```bash
@agent-git-manager "feature 브랜치 생성"

# 출력:
새 브랜치를 생성하시겠습니까?
- 브랜치명: feature/auth-implementation
- 베이스: develop

진행하시겠습니까? (y/n)
```

#### 2. 커밋 자동화

```bash
@agent-git-manager "변경사항 커밋"

# 자동 실행:
git add .
git commit -m "feat(auth): implement authentication service

- Add AuthService with email/password authentication
- Add JWT token generation
- Add input validation
- @TASK:AUTH-001"
```

#### 3. PR 생성 (사용자 확인)

```bash
@agent-git-manager "PR 생성"

# 출력:
Pull Request를 생성하시겠습니까?
- 제목: feat(auth): Authentication system
- 베이스: develop ← feature/auth-implementation
- 상태: Draft

진행하시겠습니까? (y/n)
```

## 7. trust-checker

### 역할

**품질 검증 통합 에이전트**

- TRUST 5원칙 검사
- 코드 품질 분석
- 자동 리포트 생성

### 주요 기능

#### 1. TRUST 원칙 검증

```bash
@agent-trust-checker "TRUST 원칙 검증"
```

검증 리포트:

```markdown
# TRUST 준수율: 92%

## T - Test First: 80%
✓ Test Coverage: 92.5%
✓ TDD Cycle: 85%
⚠️ Red-Green-Refactor: 75%

## R - Readable: 100%
✓ Function Size: 100%
✓ File Size: 100%
✓ Complexity: 98%

## U - Unified: 90%
✓ SPEC-driven: 95%
✓ Consistency: 85%

## S - Secured: 100%
✓ Input Validation: 100%
✓ Winston Logger: 97.92%
✓ Secret Management: 100%

## T - Trackable: 90%
✓ TAG Chain: 94%
✓ SPEC-Code Link: 88%
```

#### 2. 코드 품질 분석

```bash
@agent-trust-checker "코드 품질 분석"
```

#### 3. 개선 제안

```bash
@agent-trust-checker "품질 개선 제안"
```

## 에이전트 연계 사용

### 3단계 워크플로우

```bash
# 1단계: SPEC 작성
@agent-spec-builder "사용자 인증 기능"

# 2단계: TDD 구현
@agent-code-builder "SPEC-AUTH-001 구현"

# 3단계: 문서 동기화
@agent-doc-syncer "문서 동기화 수행"
```

### 디버깅 포함

```bash
# 오류 발생 시
@agent-debug-helper "TypeError 분석"

# 수정 후
@agent-code-builder "SPEC-AUTH-001 재구현"

# 검증
@agent-trust-checker "품질 검증"
```

## 다음 단계

### 워크플로우 명령어

- **[/moai:1-spec](/claude/commands)**: SPEC 작성 자동화
- **[/moai:2-build](/claude/commands)**: TDD 구현 자동화
- **[/moai:3-sync](/claude/commands)**: 문서 동기화 자동화

### 이벤트 훅

- **[Hooks 가이드](/claude/hooks)**: 8개 이벤트 훅 활용

### 실전 활용

- **[빠른 시작](/getting-started/quick-start)**: 에이전트 실습
- **[워크플로우](/concepts/workflow)**: 3단계 사이클

## 참고 자료

- **에이전트 소스**: `.claude/agents/moai/`
- **설정 파일**: `.claude/settings.json`
- **커스터마이징**: [고급 가이드](/advanced/custom-agents)