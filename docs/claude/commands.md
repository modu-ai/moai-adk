---
title: 워크플로우 명령어
description: 4개 슬래시 명령어로 개발 사이클 자동화
---

# 워크플로우 명령어

MoAI-ADK는 SPEC-First TDD 개발을 위한 **4개 슬래시 명령어**를 제공합니다. 각 명령어는 개발 사이클의 특정 단계를 자동화합니다.

## 명령어 개요

### 전체 명령어 목록

| 명령어 | 단계 | 주요 기능 | 자동화 |
|--------|------|-----------|--------|
| `/moai:8-project` | 준비 (선택) | 프로젝트 비전 수립 | 3대 문서 생성 |
| `/moai:1-spec` | SPEC 작성 | EARS 요구사항 작성 | 사용자 확인 후 브랜치 |
| `/moai:2-build` | TDD 구현 | Red-Green-Refactor | 범용 언어 자동 지원 |
| `/moai:3-sync` | 문서 동기화 | TAG 검증, PR 전환 | 사용자 확인 후 머지 |

### 기본 사용법

```bash
# Claude Code에서 입력
/moai:1-spec "기능 제목"
/moai:2-build SPEC-ID
/moai:3-sync
```

## /moai:8-project (선택)

### 목적

**프로젝트 비전 수립 및 초기 문서 생성**

새 프로젝트 시작 시 또는 프로젝트 방향 재정의 시 사용합니다.

### 생성되는 문서

#### 1. product.md (제품 정의)

```markdown
# my-project Product Definition

## @VISION:MISSION-001 핵심 미션
[프로젝트의 목표와 가치]

## @REQ:USER-001 주요 사용자층
- **대상**: [타겟 사용자]
- **핵심 니즈**: [해결할 문제]

## @REQ:PROBLEM-001 해결하는 핵심 문제
1. [주요 문제 1]
2. [주요 문제 2]

## @REQ:SUCCESS-001 성공 지표
- [측정 가능한 KPI]
```

#### 2. structure.md (구조 설계)

```markdown
# my-project Structure Design

## @STRUCT:ARCHITECTURE-001 시스템 아키텍처
```
Project Architecture
├── Frontend Layer    # 사용자 인터페이스
├── API Layer        # 비즈니스 로직
├── Data Layer       # 데이터 저장
└── External APIs    # 외부 통합
```

## @STRUCT:MODULES-001 모듈별 책임 구분
[각 모듈의 역할과 인터페이스]
```

#### 3. tech.md (기술 스택)

```markdown
# my-project Technology Stack

## @TECH:STACK-001 언어 & 런타임
- **주 언어**: TypeScript 5.9.2+
- **런타임**: Node.js 18+
- **패키지 매니저**: Bun 1.2.19

## @TECH:FRAMEWORK-001 핵심 프레임워크
- **웹 프레임워크**: Express.js
- **테스트**: Vitest
- **린터**: Biome
```

### 사용 예시

```bash
# 기본 사용
/moai:8-project

# 대화형 생성
에이전트: "프로젝트 이름은 무엇인가요?"
사용자: "todo-app"

에이전트: "주요 사용자층은 누구인가요?"
사용자: "개인 생산성을 높이고 싶은 사용자"

에이전트: "핵심 기능 3가지를 알려주세요"
사용자: "할일 관리, 우선순위 설정, 데드라인 알림"
```

### 언제 사용하나?

- ✅ 새 프로젝트 시작
- ✅ 프로젝트 방향 재정의
- ✅ 팀 온보딩 문서 필요
- ❌ 기존 프로젝트에 기능 추가만 할 때

## /moai:1-spec

### 목적

**EARS 방법론 기반 SPEC 작성**

기능 요구사항을 체계적으로 작성하고, @TAG Catalog를 생성합니다.

### 기본 사용법

```bash
# 새 SPEC 작성
/moai:1-spec "기능 제목"

# 복수 SPEC 작성
/moai:1-spec "기능1" "기능2" "기능3"

# 기존 SPEC 수정
/moai:1-spec SPEC-001 "수정 내용"
```

### 실행 프로세스

#### 1단계: SPEC 초안 생성

```markdown
# SPEC-AUTH-001: 사용자 이메일/비밀번호 인증

## Metadata
- **ID**: SPEC-AUTH-001
- **생성일**: 2024-01-15
- **상태**: Draft
- **담당자**: @username

## Background
사용자가 시스템에 안전하게 접근하기 위한 인증 메커니즘이 필요합니다.

## Problem Statement
현재 시스템에는 사용자 인증 기능이 없어, 누구나 모든 리소스에 접근할 수 있습니다.

## Requirements

### Ubiquitous Requirements
- 시스템은 이메일/비밀번호 기반 인증을 제공해야 한다
- 시스템은 JWT 토큰 기반 세션 관리를 제공해야 한다

### Event-driven Requirements
- WHEN 유효한 자격증명으로 로그인하면, 시스템은 JWT 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 시스템은 401 Unauthorized 에러를 반환해야 한다
- WHEN 잘못된 자격증명이 제공되면, 시스템은 로그인을 거부해야 한다

### State-driven Requirements
- WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 리소스 접근을 허용해야 한다
- WHILE 토큰이 유효할 때, 시스템은 자동 재인증 없이 요청을 처리해야 한다

### Optional Features
- WHERE 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급할 수 있다

### Constraints
- 액세스 토큰 만료시간은 15분을 초과하지 않아야 한다
- 비밀번호는 bcrypt (cost factor: 12)로 해싱해야 한다
- 토큰은 HS256 알고리즘으로 서명해야 한다

## Acceptance Criteria
- [ ] 유효한 자격증명으로 로그인 성공
- [ ] 잘못된 자격증명으로 로그인 실패
- [ ] JWT 토큰 발급 및 검증
- [ ] 토큰 만료 시 401 에러 반환
- [ ] 비밀번호 bcrypt 해싱 적용

## @TAG Catalog
| Chain | TAG | 설명 | 연관 산출물 |
|-------|-----|------|------------|
| Primary | @REQ:AUTH-001 | 인증 요구사항 | 이 문서 |
| Primary | @DESIGN:AUTH-001 | 인증 설계 | design/auth-sequence.md |
| Primary | @TASK:AUTH-001 | 인증 구현 | src/auth/service.ts |
| Primary | @TEST:AUTH-001 | 인증 테스트 | __tests__/auth/service.test.ts |
| Implementation | @FEATURE:AUTH-001 | 인증 서비스 | src/auth/ |
| Implementation | @API:AUTH-001 | 인증 API | src/auth/controller.ts |
```

#### 2단계: 사용자 확인 및 수정

```
에이전트: "SPEC-AUTH-001 초안이 생성되었습니다. 검토해주세요.

          수정이 필요한 부분이 있나요? (y/n)"

사용자: y

사용자: "리프레시 토큰 요구사항을 필수로 변경해주세요"

에이전트: "수정 완료했습니다. 다시 검토해주세요."
```

#### 3단계: 브랜치 생성 (사용자 확인)

```
에이전트: "SPEC이 확정되었습니다.

          feature/spec-auth-001-authentication 브랜치를 생성하시겠습니까?

          - 베이스 브랜치: develop
          - 브랜치명: feature/spec-auth-001-authentication

          진행하시겠습니까? (y/n)"

사용자: y

에이전트: "브랜치 생성 완료.
          Git 브랜치: feature/spec-auth-001-authentication

          다음 단계: /moai:2-build SPEC-AUTH-001"
```

### Personal vs Team 모드

#### Personal 모드

```bash
/moai:1-spec "사용자 인증"

# 저장 위치: .moai/specs/SPEC-AUTH-001/
# 브랜치: 로컬 Git 브랜치 생성
```

#### Team 모드

```bash
/moai:1-spec "사용자 인증"

# 저장 위치: GitHub Issues + .moai/specs/
# 브랜치: GitHub 브랜치 + PR 생성
# 팀 가시성: GitHub Issues에서 확인 가능
```

### 사용 예시

#### 기본 사용

```bash
# 단일 기능
/moai:1-spec "사용자 인증"

# 관련 기능들
/moai:1-spec "회원가입" "로그인" "비밀번호 재설정"

# 상세 컨텍스트
/moai:1-spec "OAuth2 소셜 로그인 (Google, GitHub, Apple 지원)"
```

#### 고급 사용

```bash
# 기존 SPEC 수정
/moai:1-spec SPEC-AUTH-001 "2FA 추가"

# 특정 EARS 구문만
/moai:1-spec "결제 시스템" --ears event-driven

# API 전용 SPEC
/moai:1-spec "REST API 인증 엔드포인트" --template api
```

## /moai:2-build

### 목적

**TDD 기반 기능 구현**

Red-Green-Refactor 사이클로 SPEC을 코드로 전환합니다.

### 기본 사용법

```bash
# 특정 SPEC 구현
/moai:2-build SPEC-001

# 여러 SPEC 구현
/moai:2-build SPEC-001 SPEC-002 SPEC-003

# 모든 Draft SPEC 구현
/moai:2-build all
```

### 실행 프로세스

#### 1단계: 분석 및 계획

```
에이전트: "SPEC-AUTH-001 분석 중..."

## 구현 계획

### 언어 감지
- 프로젝트 언어: TypeScript
- 테스트 도구: Vitest
- 린터: Biome

### 필요한 파일
1. src/auth/service.ts        (@TASK:AUTH-001)
2. src/auth/types.ts           (@DATA:AUTH-001)
3. __tests__/auth/service.test.ts (@TEST:AUTH-001)

### 의존성
- bcrypt: 비밀번호 해싱
- jsonwebtoken: JWT 생성/검증
- @types/bcrypt, @types/jsonwebtoken

### 우선순위
- 1차 목표: 핵심 인증 로직
- 2차 목표: 토큰 검증
- 최종 목표: 통합 테스트

이 계획으로 진행하시겠습니까? (y/n)
```

#### 2단계: RED Phase (사용자 승인 후)

```typescript
// __tests__/auth/service.test.ts
// @TEST:AUTH-001: 인증 서비스 테스트
import { describe, test, expect, beforeEach } from 'vitest';
import { AuthService } from '../../src/auth/service';

describe('AuthService - SPEC-AUTH-001', () => {
  let service: AuthService;

  beforeEach(() => {
    service = new AuthService();
  });

  test('should authenticate valid credentials', async () => {
    const result = await service.authenticate(
      'user@example.com',
      'password123'
    );

    expect(result.success).toBe(true);
    expect(result.token).toBeDefined();
  });

  test('should reject invalid credentials', async () => {
    const result = await service.authenticate(
      'user@example.com',
      'wrongpassword'
    );

    expect(result.success).toBe(false);
    expect(result.error).toBe('Invalid credentials');
  });
});
```

테스트 실행 → **실패 확인**:

```
FAIL  __tests__/auth/service.test.ts
  ✗ should authenticate valid credentials
    TypeError: service.authenticate is not a function
```

#### 3단계: GREEN Phase

```typescript
// src/auth/service.ts
// @TASK:AUTH-001: 인증 서비스 구현
export class AuthService {
  async authenticate(email: string, password: string) {
    // 최소 구현으로 테스트 통과
    if (email === 'user@example.com' && password === 'password123') {
      return {
        success: true,
        token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
      };
    }

    return {
      success: false,
      error: 'Invalid credentials'
    };
  }
}
```

테스트 실행 → **통과 확인**:

```
PASS  __tests__/auth/service.test.ts
  ✓ should authenticate valid credentials (15ms)
  ✓ should reject invalid credentials (8ms)
```

#### 4단계: REFACTOR Phase

```typescript
// @FEATURE:AUTH-001 | Chain: @REQ:AUTH-001 → @DESIGN:AUTH-001 → @TASK:AUTH-001 → @TEST:AUTH-001
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
      return this.failureResponse();
    }

    // @SEC:AUTH-001: bcrypt 비밀번호 검증
    const isValid = await this.passwordService.verify(
      password,
      user.passwordHash
    );
    if (!isValid) {
      return this.failureResponse();
    }

    // @TASK:AUTH-001: JWT 토큰 발급
    const token = await this.tokenService.generate(user);

    return {
      success: true,
      token
    };
  }

  private validateInput(email: string, password: string): void {
    if (!email || !password) {
      throw new ValidationError('Email and password required');
    }
  }

  private failureResponse(): AuthResult {
    return {
      success: false,
      error: 'Invalid credentials'
    };
  }
}
```

#### 5단계: TRUST 검증

```
에이전트: "구현 완료. TRUST 원칙 검증 중..."

✓ Test First: 100% (테스트 먼저 작성)
✓ Readable: 100% (함수 ≤50 LOC)
✓ Unified: 100% (SPEC 기반 설계)
✓ Secured: 100% (입력 검증, bcrypt)
✓ Trackable: 100% (TAG 적용)

TRUST 준수율: 100%

다음 단계: /moai:3-sync
```

### 언어별 자동 지원

#### TypeScript

```bash
/moai:2-build SPEC-001

# 자동 사용:
# - Vitest (테스트)
# - Biome (린터)
# - tsup (빌드)
```

#### Python

```bash
/moai:2-build SPEC-001

# 자동 사용:
# - pytest (테스트)
# - mypy (타입 검사)
# - black (포맷터)
# - ruff (린터)
```

### 사용 예시

```bash
# 기본 TDD 구현
/moai:2-build SPEC-AUTH-001

# 특정 단계만
/moai:2-build SPEC-AUTH-001 --phase red
/moai:2-build SPEC-AUTH-001 --phase refactor

# 여러 SPEC 일괄 구현
/moai:2-build SPEC-001 SPEC-002 SPEC-003

# 모든 Draft SPEC
/moai:2-build all --filter status:draft
```

## /moai:3-sync

### 목적

**문서 동기화 및 TAG 검증**

코드와 문서를 동기화하고, @TAG 체인 무결성을 검증합니다.

### 기본 사용법

```bash
# 전체 동기화
/moai:3-sync

# 특정 모드
/moai:3-sync full        # 전체 동기화
/moai:3-sync tags-only   # TAG 검증만
/moai:3-sync docs-only   # 문서만

# 특정 경로
/moai:3-sync --path src/auth
```

### 실행 프로세스

#### 1단계: TAG 체인 검증

```
에이전트: "코드 전체 스캔 중..."

✓ TAG 스캔 완료
  - 총 TAG: 149개
  - 파일: 122개
  - 스캔 시간: 45ms

✓ Primary Chain 검증
  - 완결 체인: 32개
  - 불완전 체인: 2개
  - 고아 TAG: 0개

⚠️ 불완전한 체인
  - NOTIFICATION-004: @TEST 누락
  - REPORT-005: @DESIGN 누락
```

#### 2단계: Living Document 업데이트

```
✓ README.md 갱신
  - 새 기능 추가: 사용자 인증
  - 사용법 업데이트

✓ API 문서 자동 생성
  - src/auth/service.ts → docs/api/auth.md

✓ CHANGELOG.md 업데이트
  - v0.0.2: feat(auth): 인증 시스템 추가
```

#### 3단계: sync-report.md 생성

```markdown
# 동기화 리포트 (2024-01-15 14:30)

## TAG 통계
- 총 TAG: 149개
- 완결 체인: 32개
- 불완전 체인: 2개

## 새로 추가된 TAG
- @REQ:AUTH-001
- @DESIGN:AUTH-001
- @TASK:AUTH-001
- @TEST:AUTH-001
- @FEATURE:AUTH-001
- @SEC:AUTH-001

## 권장 사항
1. NOTIFICATION-004: __tests__/notification/service.test.ts 추가
2. REPORT-005: design/ 폴더에 설계 문서 추가
```

#### 4단계: PR 상태 전환 (사용자 확인)

```
에이전트: "문서 동기화가 완료되었습니다.

          다음 작업을 수행하시겠습니까?

          1. PR 상태: Draft → Ready for Review
          2. develop 브랜치로 머지 요청

          진행하시겠습니까? (y/n)"

사용자: y

에이전트: "PR 상태 변경 완료.
          리뷰어에게 알림이 전송되었습니다."
```

### 사용 예시

```bash
# 전체 동기화
/moai:3-sync

# TAG 검증만
/moai:3-sync tags-only

# 특정 문서만
/moai:3-sync docs-only --target api

# 특정 경로만
/moai:3-sync --path src/auth
```

## 워크플로우 완전 예시

### 사용자 인증 기능 구현

```bash
# 1단계: 프로젝트 준비 (선택)
/moai:8-project

# 2단계: SPEC 작성
/moai:1-spec "사용자 이메일/비밀번호 인증"

# (사용자 확인 및 브랜치 생성)

# 3단계: TDD 구현
/moai:2-build SPEC-AUTH-001

# (사용자 승인 및 구현)

# 4단계: 문서 동기화
/moai:3-sync

# (사용자 확인 및 PR 전환)

# 완료!
```

## 다음 단계

### 에이전트 활용

- **[에이전트 가이드](/claude/agents)**: 8개 에이전트 상세
- **[훅 시스템](/claude/hooks)**: 자동화 훅

### 실전 활용

- **[빠른 시작](/getting-started/quick-start)**: 5분 튜토리얼
- **[워크플로우](/concepts/workflow)**: 상세 개발 사이클

## 참고 자료

- **명령어 소스**: `.claude/commands/moai/`
- **설정 파일**: `.claude/settings.json`
- **커스터마이징**: [고급 가이드](/advanced/custom-commands)