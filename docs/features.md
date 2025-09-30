---
title: 주요 특징
description: MoAI-ADK의 핵심 기능과 차별점
---

# 주요 특징

MoAI-ADK는 SPEC-First TDD 개발을 위한 완전한 자동화 프레임워크입니다. 다음 5가지 핵심 기능으로 현대적인 개발 경험을 제공합니다.

## 1. SPEC-First 개발

### EARS 방법론 지원

**EARS (Easy Approach to Requirements Syntax)**: 체계적인 요구사항 작성 방법론

```markdown
### Ubiquitous Requirements (기본 기능)
- 시스템은 결제 처리 기능을 제공해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 사용자가 결제 버튼을 클릭하면, 시스템은 결제 프로세스를 시작해야 한다
- WHEN 결제가 완료되면, 시스템은 주문 확인 이메일을 발송해야 한다

### State-driven Requirements (상태 기반)
- WHILE 결제가 진행 중일 때, 시스템은 로딩 화면을 표시해야 한다
- WHILE 재고가 부족할 때, 시스템은 구매 버튼을 비활성화해야 한다

### Optional Features (선택적 기능)
- WHERE 프리미엄 회원이면, 시스템은 무료 배송을 제공할 수 있다

### Constraints (제약사항)
- IF 결제 금액이 0원이면, 시스템은 결제를 거부해야 한다
- 결제 처리 시간은 3초를 초과하지 않아야 한다
```

### 자동 SPEC 템플릿 생성

`/moai:1-spec "기능 제목"` 실행 시 자동으로 생성되는 구조:

```markdown
# SPEC-001: [기능 제목]

## Metadata
- **ID**: SPEC-001
- **생성일**: 2024-01-15
- **상태**: Draft
- **담당자**: @user

## Background
[배경 설명]

## Problem Statement
[해결할 문제]

## Requirements

### Ubiquitous Requirements
- [기본 기능 요구사항]

### Event-driven Requirements
- WHEN [조건], [동작]

### State-driven Requirements
- WHILE [상태], [동작]

### Constraints
- [제약사항]

## Acceptance Criteria
- [ ] [검수 기준 1]
- [ ] [검수 기준 2]

## @TAG Catalog
| Chain | TAG | 설명 | 연관 산출물 |
|-------|-----|------|------------|
| Primary | @REQ:SPEC-001 | 요구사항 | 이 문서 |
| Primary | @DESIGN:SPEC-001 | 설계 | design/ |
| Primary | @TASK:SPEC-001 | 구현 | src/ |
| Primary | @TEST:SPEC-001 | 테스트 | __tests__/ |
```

### GitHub Issues 통합 (Team 모드)

Personal 모드와 Team 모드 차이:

| 기능 | Personal 모드 | Team 모드 |
|------|--------------|-----------|
| SPEC 저장 | 로컬 `.moai/specs/` | GitHub Issues |
| 브랜치 생성 | 사용자 확인 필수 | 사용자 확인 필수 |
| PR 생성 | Git 저장소 | GitHub API |
| 협업 기능 | 제한적 | 전체 팀 협업 |

Team 모드 활성화:

```bash
moai init my-project --team
```

## 2. TDD 자동화

### Red-Green-Refactor 강제

MoAI-ADK는 엄격한 TDD 사이클을 강제합니다:

#### Red Phase: 실패하는 테스트

```typescript
// @TEST:AUTH-001: 유효한 사용자 인증 테스트
import { describe, test, expect } from 'vitest';
import { AuthService } from '../src/auth/service';

describe('AuthService', () => {
  test('should authenticate valid user', async () => {
    const service = new AuthService();
    const result = await service.authenticate('user@example.com', 'password');

    expect(result.success).toBe(true);
    expect(result.token).toBeDefined();
    expect(result.user.email).toBe('user@example.com');
  });

  test('should reject invalid credentials', async () => {
    const service = new AuthService();
    const result = await service.authenticate('user@example.com', 'wrong');

    expect(result.success).toBe(false);
    expect(result.error).toBe('Invalid credentials');
  });
});
```

#### Green Phase: 최소 구현

```typescript
// @FEATURE:AUTH-001 | Chain: @REQ → @DESIGN → @TASK → @TEST
export class AuthService {
  async authenticate(email: string, password: string) {
    // 최소한의 구현으로 테스트 통과
    if (email === 'user@example.com' && password === 'password') {
      return {
        success: true,
        token: 'jwt-token',
        user: { email }
      };
    }

    return {
      success: false,
      error: 'Invalid credentials'
    };
  }
}
```

#### Refactor Phase: 품질 개선

```typescript
// @FEATURE:AUTH-001 | Chain: @REQ → @DESIGN → @TASK → @TEST
// Related: @SEC:AUTH-001, @PERF:AUTH-001

export class AuthService {
  constructor(
    private userRepository: UserRepository,
    private tokenService: TokenService,
    private passwordService: PasswordService
  ) {}

  async authenticate(email: string, password: string): Promise<AuthResult> {
    // @SEC:AUTH-001: 입력값 검증
    this.validateInput(email, password);

    // @TASK:AUTH-001: 사용자 조회
    const user = await this.userRepository.findByEmail(email);
    if (!user) {
      return this.failureResponse('Invalid credentials');
    }

    // @SEC:AUTH-001: 비밀번호 검증
    const isValid = await this.passwordService.verify(password, user.passwordHash);
    if (!isValid) {
      return this.failureResponse('Invalid credentials');
    }

    // @TASK:AUTH-001: 토큰 발급
    const token = await this.tokenService.generate(user);

    return {
      success: true,
      token,
      user: { email: user.email }
    };
  }

  private validateInput(email: string, password: string): void {
    if (!email || !password) {
      throw new ValidationError('Email and password are required');
    }
  }

  private failureResponse(error: string): AuthResult {
    return { success: false, error };
  }
}
```

### 언어별 자동 도구 선택

MoAI-ADK는 프로젝트 언어를 자동 감지하고 최적의 도구를 선택합니다:

#### TypeScript 프로젝트

```bash
# 자동 감지: package.json, tsconfig.json
/moai:2-build SPEC-001

# 자동 사용 도구:
# - Vitest: 테스트 프레임워크
# - Biome: 린터 + 포맷터
# - TypeScript: 타입 검사
```

#### Python 프로젝트

```bash
# 자동 감지: requirements.txt, pyproject.toml
/moai:2-build SPEC-001

# 자동 사용 도구:
# - pytest: 테스트 프레임워크
# - mypy: 타입 검사
# - black: 포맷터
# - ruff: 린터
```

#### Java 프로젝트

```bash
# 자동 감지: pom.xml, build.gradle
/moai:2-build SPEC-001

# 자동 사용 도구:
# - JUnit: 테스트 프레임워크
# - Maven/Gradle: 빌드 도구
```

### 92.9% 테스트 성공률

MoAI-ADK v0.0.1 달성 지표:

- **Vitest 테스트**: 92.9% 성공률
- **테스트 속도**: 평균 45ms (Bun 최적화)
- **커버리지**: 85% 이상 목표
- **타입 안전성**: TypeScript strict 모드 100%

## 3. @TAG 추적성

### Primary Chain: 4단계 필수 체인

모든 기능은 다음 4단계 체인을 따릅니다:

```
@REQ → @DESIGN → @TASK → @TEST
```

#### 실제 예시: 사용자 인증

**1. @REQ (요구사항)**

```markdown
# SPEC-AUTH-001

## @REQ:AUTH-001
- 시스템은 이메일/비밀번호 기반 인증을 제공해야 한다
- WHEN 유효한 자격증명이 제공되면, JWT 토큰을 발급해야 한다
- IF 자격증명이 틀리면, 접근을 거부해야 한다
```

**2. @DESIGN (설계)**

```markdown
## @DESIGN:AUTH-001

### 시퀀스 다이어그램
User → AuthService → UserRepository → PasswordService → TokenService

### 인터페이스 설계
interface AuthService {
  authenticate(email: string, password: string): Promise<AuthResult>
}
```

**3. @TASK (구현)**

```typescript
// @TASK:AUTH-001: 인증 서비스 구현
export class AuthService {
  async authenticate(email: string, password: string): Promise<AuthResult> {
    // 구현...
  }
}
```

**4. @TEST (검증)**

```typescript
// @TEST:AUTH-001: 인증 서비스 테스트
describe('AuthService', () => {
  test('should authenticate valid user', async () => {
    // 테스트...
  });
});
```

### 코드 직접 스캔 방식

MoAI-ADK는 중간 캐시 없이 코드를 직접 스캔하여 TAG 추적성을 보장합니다:

```bash
# TAG 검색 (ripgrep 권장)
rg "@REQ:AUTH-001" -n          # 요구사항 TAG 검색
rg "@TASK:AUTH-001" -n         # 구현 TAG 검색
rg "AUTH-001" -n               # 모든 관련 TAG 검색

# TAG 체인 검증
/moai:3-sync                   # 전체 코드 스캔 및 검증
```

### 실시간 무결성 검증

`/moai:3-sync` 실행 시 자동으로 검증되는 항목:

- ✅ **Primary Chain 완결성**: @REQ → @DESIGN → @TASK → @TEST 연결 확인
- ✅ **고아 TAG 감지**: 참조되지 않는 TAG 식별
- ✅ **끊어진 링크**: 중간 단계 누락 확인
- ✅ **중복 TAG**: 동일 ID 중복 사용 검사

검증 리포트 예시:

```markdown
# TAG 검증 리포트

## 완결된 체인 (3개)
- AUTH-001: ✅ 완전 (REQ → DESIGN → TASK → TEST)
- PAYMENT-002: ✅ 완전
- PROFILE-003: ✅ 완전

## 불완전한 체인 (1개)
- NOTIFICATION-004: ⚠️ @TEST 누락

## 고아 TAG (0개)

## 중복 TAG (0개)
```

## 4. 범용 언어 지원

### 지원 언어 목록

MoAI-ADK는 다음 언어를 공식 지원합니다:

| 언어 | 테스트 도구 | 린터 | 포맷터 | 빌드 도구 |
|------|-------------|------|--------|-----------|
| **TypeScript** | Vitest | Biome | Biome | tsup |
| **Python** | pytest | ruff | black | pip |
| **Java** | JUnit | Checkstyle | Google Java Format | Maven/Gradle |
| **Go** | go test | golangci-lint | gofmt | go build |
| **Rust** | cargo test | clippy | rustfmt | cargo |
| **C++** | GoogleTest | clang-tidy | clang-format | CMake |
| **C#** | xUnit | Roslyn | dotnet format | MSBuild |
| **PHP** | PHPUnit | PHPStan | PHP-CS-Fixer | Composer |

### 자동 언어 감지

프로젝트를 분석하여 자동으로 언어를 감지합니다:

```bash
moai doctor

# 출력:
✓ Language Detection
  - JavaScript/TypeScript: 65% (package.json, tsconfig.json detected)
  - Python: 25% (requirements.txt detected)
  - Go: 10% (go.mod detected)

✓ Language-Specific Requirements
  - npm: ✓ (v10.2.3)
  - TypeScript: ✓ (v5.9.2)
  - pytest: ✓ (v8.0.0)
```

### 도구 자동 매핑

각 언어별로 최적의 도구를 자동 선택합니다:

```json
{
  "languageSupport": {
    "typescript": {
      "testRunner": "vitest",
      "linter": "biome",
      "formatter": "biome"
    },
    "python": {
      "testRunner": "pytest",
      "linter": "ruff",
      "formatter": "black"
    },
    "java": {
      "testRunner": "junit",
      "buildTool": "maven"
    }
  }
}
```

## 5. Claude Code 통합

### 7개 전문 에이전트

각 단계를 전담하는 전문 에이전트:

#### 1. spec-builder

```
@agent-spec-builder "사용자 인증 기능 SPEC 작성"
```

- SPEC 작성 전담
- EARS 요구사항 생성
- @TAG Catalog 자동 생성
- 브랜치 생성 (사용자 확인 후)

#### 2. code-builder

```
@agent-code-builder "SPEC-001 구현 계획 수립"
```

- TDD 구현 전담
- 범용 언어 지원
- Red-Green-Refactor 자동화
- @TAG 자동 삽입

#### 3. doc-syncer

```
@agent-doc-syncer "문서 동기화 수행"
```

- 문서 동기화 전담
- TAG 체인 검증
- Living Document 업데이트
- PR 상태 전환 (사용자 확인 후)

#### 4. cc-manager

```
@agent-cc-manager "Claude Code 설정 최적화"
```

- Claude Code 설정 전담
- 권한 최적화
- 설정 표준화

#### 5. debug-helper

```
@agent-debug-helper "TypeError 오류 분석"
```

- 온디맨드 디버깅
- 시스템 진단
- 개발 가이드 검사
- TAG 무결성 검증

#### 6. git-manager

```
@agent-git-manager "feature 브랜치 생성"
```

- Git 작업 전담
- 브랜치/머지 관리 (사용자 확인 후)
- 커밋 자동화

#### 7. trust-checker

```
@agent-trust-checker "TRUST 원칙 검증"
```

- 품질 검증 통합
- TRUST 5원칙 검사
- 코드 품질 분석

### 5개 워크플로우 명령어

```bash
/moai:0-project    # (선택) 프로젝트 비전 수립
/moai:1-spec      # SPEC 작성
/moai:2-build     # TDD 구현
/moai:3-sync      # 문서 동기화
/moai:help        # 도움말
```

### 8개 이벤트 훅

자동으로 실행되는 이벤트 훅:

1. **file-monitor**: 파일 변경 모니터링
2. **language-detector**: 언어 자동 감지
3. **policy-block**: 정책 위반 차단
4. **pre-write-guard**: 파일 쓰기 전 검증
5. **session-notice**: 세션 시작 알림
6. **steering-guard**: 개발 방향 가이드
7. **run-tests-and-report**: 테스트 자동 실행
8. **claude-code-monitor**: Claude Code 상태 감시

## 실습 예제: 사용자 인증 구현

### 1단계: SPEC 작성

```bash
/moai:1-spec "사용자 이메일/비밀번호 인증"
```

생성된 SPEC-AUTH-001:

```markdown
## Requirements

### Ubiquitous Requirements
- 시스템은 이메일/비밀번호 기반 인증을 제공해야 한다

### Event-driven Requirements
- WHEN 유효한 자격증명으로 로그인하면, JWT 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 401 에러를 반환해야 한다

### Constraints
- 토큰 만료시간은 15분을 초과하지 않아야 한다
```

### 2단계: TDD 구현

```bash
/moai:2-build SPEC-AUTH-001
```

자동 생성되는 파일:

```
src/
  auth/
    service.ts        # @TASK:AUTH-001
    types.ts          # @DATA:AUTH-001
__tests__/
  auth/
    service.test.ts   # @TEST:AUTH-001
```

### 3단계: 문서 동기화

```bash
/moai:3-sync
```

업데이트되는 항목:

- TAG 체인 검증
- API 문서 자동 생성
- README 업데이트
- sync-report.md 생성

## 다음 단계

### 시작하기

- **[설치](/getting-started/installation)**: 시스템 요구사항 및 설치
- **[빠른 시작](/getting-started/quick-start)**: 5분 튜토리얼
- **[프로젝트 초기화](/getting-started/project-setup)**: 프로젝트 구조 이해

### 심화 학습

- **[SPEC-First TDD](/concepts/spec-first-tdd)**: 방법론 완전 가이드
- **[TAG 시스템](/concepts/tag-system)**: 추적성 관리
- **[TRUST 원칙](/concepts/trust-principles)**: 품질 기준

### 실전 활용

- **[TypeScript 가이드](/languages/typescript)**: TypeScript 프로젝트
- **[Python 가이드](/languages/python)**: Python 프로젝트
- **[CLI 레퍼런스](/cli/init)**: 명령어 상세

## 참고 자료

- **GitHub**: [MoAI-ADK Repository](https://github.com/modu-ai/moai-adk)
- **NPM**: [@moai/adk](https://www.npmjs.com/package/@moai/adk)
- **커뮤니티**: [Discussions](https://github.com/modu-ai/moai-adk/discussions)