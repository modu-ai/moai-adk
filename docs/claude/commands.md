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

## @DOC:MISSION-001 핵심 미션
[프로젝트의 목표와 가치]

## @SPEC:USER-001 주요 사용자층
- **대상**: [타겟 사용자]
- **핵심 니즈**: [해결할 문제]

## @SPEC:PROBLEM-001 해결하는 핵심 문제
1. [주요 문제 1]
2. [주요 문제 2]

## @SPEC:SUCCESS-001 성공 지표
- [측정 가능한 KPI]
```

#### 2. structure.md (구조 설계)

```markdown
# my-project Structure Design

## @DOC:ARCHITECTURE-001 시스템 아키텍처
```
Project Architecture
├── Frontend Layer    # 사용자 인터페이스
├── API Layer        # 비즈니스 로직
├── Data Layer       # 데이터 저장
└── External APIs    # 외부 통합
```

## @DOC:MODULES-001 모듈별 책임 구분
[각 모듈의 역할과 인터페이스]
```

#### 3. tech.md (기술 스택)

```markdown
# my-project Technology Stack

## @DOC:STACK-001 언어 & 런타임
- **주 언어**: TypeScript 5.9.2+
- **런타임**: Node.js 18+
- **패키지 매니저**: Bun 1.2.19

## @DOC:FRAMEWORK-001 핵심 프레임워크
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

기능 요구사항을 체계적으로 작성하고, TAG BLOCK을 생성합니다.

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

## Traceability

TAG BLOCK을 통한 추적성 확보:

```markdown
# @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 ->  -> @CODE:AUTH-001 -> @TEST:AUTH-001
# Related: @CODE:AUTH-001:API

# SPEC-AUTH-001: 사용자 이메일/비밀번호 인증
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
1. src/auth/service.ts        (@CODE:AUTH-001)
2. src/auth/types.ts           (@CODE:AUTH-001:DATA)
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
// @CODE:AUTH-001: 인증 서비스 구현
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
// @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 →  → @CODE:AUTH-001 → @TEST:AUTH-001
// Related: @CODE:AUTH-001:INFRA, @CODE:AUTH-001

export class AuthService {
  constructor(
    private userRepository: UserRepository,
    private tokenService: TokenService,
    private passwordService: PasswordService
  ) {}

  async authenticate(email: string, password: string): Promise<AuthResult> {
    // @CODE:AUTH-001:INFRA: 입력값 검증
    this.validateInput(email, password);

    // @CODE:AUTH-001: 사용자 조회
    const user = await this.userRepository.findByEmail(email);
    if (!user) {
      return this.failureResponse();
    }

    // @CODE:AUTH-001:INFRA: bcrypt 비밀번호 검증
    const isValid = await this.passwordService.verify(
      password,
      user.passwordHash
    );
    if (!isValid) {
      return this.failureResponse();
    }

    // @CODE:AUTH-001: JWT 토큰 발급
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

✓ TAG 체인 검증
  - 완결 체인: 32개
  - 불완전 체인: 2개
  - 고아 TAG: 0개

⚠️ 불완전한 체인
  - NOTIFICATION-004: @TEST 누락
  - REPORT-005: @SPEC 누락
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
- @SPEC:AUTH-001
- 
- @CODE:AUTH-001
- @TEST:AUTH-001
- @CODE:AUTH-001
- @CODE:AUTH-001:INFRA

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

## 커스터마이징 가이드

MoAI-ADK의 슬래시 명령어는 프로젝트 특성에 맞게 커스터마이징할 수 있습니다.

### 명령어 파일 구조

```
.claude/
└── commands/
    └── moai/
        ├── 0-init.md       # /moai:0-init
        ├── 1-spec.md       # /moai:1-spec
        ├── 2-build.md      # /moai:2-build
        ├── 3-sync.md       # /moai:3-sync
        └── 8-project.md    # /moai:8-project
```

### 명령어 커스터마이징 방법

#### 1. 기존 명령어 수정

**예시: /moai:1-spec에 커스텀 템플릿 추가**

```markdown
<!-- .claude/commands/moai/1-spec.md -->

# /moai:1-spec - SPEC 작성

## 추가 옵션

### --template [type]

사용 가능한 템플릿:
- `api`: API 전용 SPEC
- `ui`: UI 컴포넌트 SPEC
- `data`: 데이터 모델 SPEC
- `microservice`: 마이크로서비스 SPEC (커스텀)

## 커스텀 템플릿: microservice

마이크로서비스 아키텍처를 위한 SPEC 템플릿:

```yaml
# SPEC-MS-001: 결제 서비스
type: microservice
domain: payment
port: 8080

## Service Contract
- REST API: /api/v1/payments
- gRPC: payment.PaymentService
- Event Bus: payment.events

## Dependencies
- User Service (HTTP)
- Notification Service (Event)
- Database: PostgreSQL
```
```

#### 2. 새로운 명령어 추가

**예시: /moai:4-deploy 추가**

```bash
# 1. 명령어 파일 생성
touch .claude/commands/moai/4-deploy.md
```

```markdown
<!-- .claude/commands/moai/4-deploy.md -->

# /moai:4-deploy - 배포 자동화

## 목적

SPEC 기반 배포 자동화 및 롤백 관리

## 사용법

```bash
/moai:4-deploy [SPEC-ID] [환경]
```

## 예시

```bash
# 스테이징 배포
/moai:4-deploy SPEC-AUTH-001 staging

# 프로덕션 배포 (승인 필요)
/moai:4-deploy SPEC-AUTH-001 production

# 롤백
/moai:4-deploy rollback SPEC-AUTH-001
```

## 배포 체크리스트

- [ ] 모든 테스트 통과 (커버리지 ≥85%)
- [ ] TRUST 원칙 준수 (≥80%)
- [ ] TAG 체인 완결성 검증
- [ ] 문서 동기화 완료
- [ ] 보안 스캔 통과
```

#### 3. settings.json 업데이트

```json
{
  "customCommands": {
    "moai:4-deploy": {
      "enabled": true,
      "requiresApproval": true,
      "environments": ["staging", "production"]
    }
  }
}
```

### EARS 템플릿 커스터마이징

프로젝트 도메인에 맞는 EARS 예시를 추가할 수 있습니다.

**예시: 금융 도메인**

```markdown
<!-- templates/ears/financial.md -->

### Ubiquitous Requirements (금융)
- 시스템은 PCI-DSS 표준을 준수해야 한다
- 시스템은 이중 인증을 제공해야 한다

### Event-driven Requirements (금융)
- WHEN 거래 금액이 $10,000 초과하면, 시스템은 매니저 승인을 요청해야 한다
- WHEN 비정상 패턴이 감지되면, 시스템은 계정을 일시 정지해야 한다

### Constraints (금융)
- 모든 거래는 감사 로그에 기록되어야 한다
- 개인정보는 AES-256으로 암호화되어야 한다
```

### TAG 시스템 커스터마이징

**예시: 도메인별 TAG 추가**

```typescript
// .moai/config/tags.ts

export const customTags = {
  // 기본 
  primary: ['@SPEC', '@SPEC', '@CODE', '@TEST'],
  implementation: ['@CODE', '@CODE', '@CODE', '@CODE'],

  // 커스텀 TAG (금융 도메인)
  finance: [
    'COMPLIANCE',  // 규제 준수
    'AUDIT',       // 감사 추적
    'FRAUD',       // 사기 탐지
    'RISK'         // 리스크 관리
  ]
};
```

### 워크플로우 훅 커스터마이징

**예시: /moai:1-spec 실행 전 검증**

```typescript
// .claude/hooks/pre-spec.ts

export async function preSpecHook(context) {
  // 중복 SPEC 체크
  const existingSpecs = await scanSpecs(context.title);

  if (existingSpecs.length > 0) {
    return {
      block: true,
      reason: `유사한 SPEC 발견: ${existingSpecs.join(', ')}`
    };
  }

  // 프로젝트 문서 존재 확인
  const hasProjectDocs = await checkProjectDocs();

  if (!hasProjectDocs) {
    return {
      block: true,
      reason: '/moai:8-project를 먼저 실행해주세요'
    };
  }

  return { block: false };
}
```

### 명령어 별칭 설정

**예시: 자주 사용하는 명령어 단축**

```json
// .claude/settings.json

{
  "commandAliases": {
    "/s": "/moai:1-spec",
    "/b": "/moai:2-build",
    "/sync": "/moai:3-sync",
    "/p": "/moai:8-project"
  }
}
```

사용 예시:
```bash
/s "사용자 인증"        # /moai:1-spec "사용자 인증"과 동일
/b SPEC-001            # /moai:2-build SPEC-001과 동일
```

### 언어별 TDD 도구 커스터마이징

**예시: Python 프로젝트에 ruff 추가**

```json
// .moai/config.json

{
  "languages": {
    "python": {
      "test": "pytest",
      "type_check": "mypy",
      "formatter": "black",
      "linter": "ruff",  // 커스텀 추가
      "coverage_threshold": 90  // 기본값 85에서 변경
    }
  }
}
```

## 고급 옵션 및 플래그

각 명령어는 고급 사용자를 위한 추가 옵션을 제공합니다.

### /moai:8-project 고급 옵션

```bash
# 기본 사용
/moai:8-project

# 특정 템플릿으로 생성
/moai:8-project --template microservice

# 대화형 모드 스킵 (기본값 사용)
/moai:8-project --skip-interactive

# 기존 문서 업데이트만
/moai:8-project --update

# 특정 문서만 생성
/moai:8-project --only product
/moai:8-project --only structure
/moai:8-project --only tech
```

#### 사용 가능한 템플릿

| 템플릿 | 설명 | 적용 사례 |
|--------|------|-----------|
| `default` | 범용 애플리케이션 | 웹앱, CLI 도구 |
| `microservice` | 마이크로서비스 | 분산 시스템 |
| `library` | 라이브러리/SDK | npm 패키지, pip 모듈 |
| `mobile` | 모바일 앱 | React Native, Flutter |
| `data-pipeline` | 데이터 파이프라인 | ETL, ML 파이프라인 |

**예시:**
```bash
/moai:8-project --template microservice

# 생성되는 추가 섹션:
# - Service Contract
# - API Gateway 설정
# - 서비스 간 통신
# - 배포 전략
```

### /moai:1-spec 고급 옵션

```bash
# 기본 사용
/moai:1-spec "기능 제목"

# 우선순위 지정
/moai:1-spec "사용자 인증" --priority high

# 특정 EARS 구문만 사용
/moai:1-spec "결제 시스템" --ears ubiquitous,event-driven

# 템플릿 지정
/moai:1-spec "API 엔드포인트" --template api

# 의존성 명시
/moai:1-spec "알림 서비스" --depends-on SPEC-AUTH-001,SPEC-USER-002

# 담당자 지정 (Team 모드)
/moai:1-spec "UI 개선" --assignee @username

# 마일스톤 연결 (Team 모드)
/moai:1-spec "대시보드" --milestone v2.0

# 라벨 추가 (Team 모드)
/moai:1-spec "성능 개선" --labels performance,optimization

# Draft 스킵하고 바로 Ready
/moai:1-spec "긴급 수정" --ready

# 브랜치 생성 스킵
/moai:1-spec "문서만 작성" --no-branch
```

#### EARS 구문 선택

```bash
# 기본 기능만
/moai:1-spec "CRUD API" --ears ubiquitous

# 이벤트 중심 시스템
/moai:1-spec "알림 시스템" --ears event-driven,state-driven

# 제약사항 중심
/moai:1-spec "보안 강화" --ears constraints

# 모두 포함 (기본값)
/moai:1-spec "사용자 관리" --ears all
```

### /moai:2-build 고급 옵션

```bash
# 기본 사용
/moai:2-build SPEC-001

# 특정 Phase만 실행
/moai:2-build SPEC-001 --phase red
/moai:2-build SPEC-001 --phase green
/moai:2-build SPEC-001 --phase refactor

# TDD 사이클 스킵 (위험!)
/moai:2-build SPEC-001 --skip-tdd

# 테스트 커버리지 임계값 지정
/moai:2-build SPEC-001 --coverage 90

# TRUST 검증 스킵 (비권장)
/moai:2-build SPEC-001 --skip-trust

# 병렬 구현 (여러 SPEC)
/moai:2-build SPEC-001 SPEC-002 SPEC-003 --parallel

# 드라이런 (계획만 확인)
/moai:2-build SPEC-001 --dry-run

# 특정 언어 도구 강제
/moai:2-build SPEC-001 --test-framework jest

# 자동 승인 (CI/CD용)
/moai:2-build SPEC-001 --auto-approve

# 실패 시 롤백
/moai:2-build SPEC-001 --rollback-on-fail
```

#### Phase별 실행

**RED Phase만:**
```bash
/moai:2-build SPEC-AUTH-001 --phase red

# 수행 작업:
# 1. SPEC 분석
# 2. 테스트 파일 생성
# 3. 실패하는 테스트 작성
# 4. 테스트 실행 (실패 확인)
```

**GREEN Phase만:**
```bash
/moai:2-build SPEC-AUTH-001 --phase green

# 수행 작업:
# 1. 최소 구현 작성
# 2. 테스트 실행 (통과 확인)
```

**REFACTOR Phase만:**
```bash
/moai:2-build SPEC-AUTH-001 --phase refactor

# 수행 작업:
# 1. 코드 개선
# 2. TAG 추가
# 3. TRUST 검증
# 4. 문서화
```

#### 병렬 구현

```bash
# 여러 SPEC 동시 구현
/moai:2-build SPEC-001 SPEC-002 SPEC-003 --parallel

# 최대 워커 수 지정
/moai:2-build SPEC-* --parallel --max-workers 4

# 의존성 있는 SPEC 순차 처리
/moai:2-build SPEC-001 SPEC-002 --parallel --respect-deps
```

### /moai:3-sync 고급 옵션

```bash
# 기본 사용
/moai:3-sync

# 특정 모드
/moai:3-sync --mode full        # 전체 동기화 (기본값)
/moai:3-sync --mode tags        # TAG 검증만
/moai:3-sync --mode docs        # 문서 갱신만
/moai:3-sync --mode status      # 상태만 확인

# 특정 경로만
/moai:3-sync --path src/auth
/moai:3-sync --path "src/**/*.ts"

# 특정 TAG만 검증
/moai:3-sync --tags @SPEC,@SPEC

# 자동 수정 활성화
/moai:3-sync --auto-fix

# Living Document 생성 스킵
/moai:3-sync --no-docs

# PR 전환 스킵
/moai:3-sync --no-pr

# 강제 동기화 (충돌 무시)
/moai:3-sync --force

# 드라이런 (미리보기)
/moai:3-sync --dry-run

# 리포트 형식 지정
/moai:3-sync --format json
/moai:3-sync --format markdown
/moai:3-sync --format html
```

#### 모드별 동작

**tags 모드:**
```bash
/moai:3-sync --mode tags

# 수행 작업:
# ✓ 코드 전체 스캔
# ✓ TAG 체인 검증
# ✓ 고아 TAG 탐지
# ✓ sync-report.md 생성
# ✗ Living Document 갱신 안 함
# ✗ PR 상태 변경 안 함
```

**docs 모드:**
```bash
/moai:3-sync --mode docs

# 수행 작업:
# ✗ TAG 검증 안 함
# ✓ README.md 갱신
# ✓ API 문서 자동 생성
# ✓ CHANGELOG.md 업데이트
# ✗ PR 상태 변경 안 함
```

**status 모드:**
```bash
/moai:3-sync --mode status

# 수행 작업:
# ✓ 현재 상태만 출력
# ✗ 파일 변경 안 함
# ✗ PR 변경 안 함

# 출력 예시:
# TAG 상태: 149개 (완결 32개, 불완전 2개)
# 문서 상태: 최신 (3일 전 동기화)
# PR 상태: Draft
```

#### 자동 수정 옵션

```bash
/moai:3-sync --auto-fix

# 자동 수정 항목:
# - 불완전한 TAG 체인 완성
# - 고아 TAG 제거 또는 연결
# - 오래된 문서 자동 갱신
# - 중복 TAG ID 통합
```

### 옵션 조합 예시

#### 빠른 TAG 검증

```bash
/moai:3-sync --mode tags --path src/auth --dry-run

# 출력만 보고 파일 변경 없음
```

#### CI/CD 자동화

```bash
# SPEC 작성 (자동 승인)
/moai:1-spec "자동 배포" --ready --no-branch

# TDD 구현 (자동 승인)
/moai:2-build SPEC-001 --auto-approve --coverage 90

# 동기화 (PR 전환)
/moai:3-sync --auto-fix --no-pr
```

#### 대규모 리팩토링

```bash
# 1. 현재 상태 확인
/moai:3-sync --mode status

# 2. TAG 검증 (드라이런)
/moai:3-sync --mode tags --dry-run

# 3. 자동 수정
/moai:3-sync --mode tags --auto-fix

# 4. 문서 동기화
/moai:3-sync --mode docs

# 5. 전체 검증
/moai:3-sync --mode full
```

#### 핫픽스 워크플로우

```bash
# 1. 긴급 SPEC (브랜치 없이)
/moai:1-spec "긴급 보안 수정" --priority critical --ready --no-branch

# 2. 빠른 구현 (TDD 스킵, 비권장!)
/moai:2-build SPEC-HOTFIX-001 --skip-tdd --skip-trust

# 3. 최소 동기화
/moai:3-sync --mode tags --no-pr
```

## 문제 해결 (Troubleshooting)

각 명령어 사용 시 발생할 수 있는 문제와 해결 방법을 안내합니다.

### /moai:8-project 문제 해결

#### 문제 1: 프로젝트 문서가 이미 존재

**증상:**
```
Error: .moai/project/ 디렉토리가 이미 존재합니다.
```

**해결 방법:**

```bash
# 방법 1: 업데이트 모드 사용
/moai:8-project --update

# 방법 2: 기존 문서 백업 후 재생성
mv .moai/project .moai/project.backup
/moai:8-project

# 방법 3: 특정 문서만 재생성
/moai:8-project --only tech
```

#### 문제 2: 대화형 입력 타임아웃

**증상:**
```
Error: 사용자 입력 타임아웃 (60초)
```

**해결 방법:**

```bash
# 기본값으로 생성 (대화형 스킵)
/moai:8-project --skip-interactive

# 또는 설정 파일 수정
# .moai/config.json
{
  "prompts": {
    "timeout": 300  # 5분으로 연장
  }
}
```

#### 문제 3: 템플릿을 찾을 수 없음

**증상:**
```
Error: Template 'microservice' not found
```

**해결 방법:**

```bash
# 사용 가능한 템플릿 확인
moai status --verbose

# 템플릿 업데이트
moai update --templates

# 커스텀 템플릿 생성
mkdir -p .moai/templates/custom
# 템플릿 파일 추가...
```

### /moai:1-spec 문제 해결

#### 문제 1: 중복된 SPEC ID

**증상:**
```
Error: SPEC-AUTH-001 already exists
```

**해결 방법:**

```bash
# 기존 SPEC 확인
rg "SPEC-AUTH-001" .moai/specs/

# 기존 SPEC 수정
/moai:1-spec SPEC-AUTH-001 "추가 요구사항"

# 새 SPEC으로 생성
/moai:1-spec "사용자 인증 v2"  # 자동으로 SPEC-AUTH-002 할당
```

#### 문제 2: 브랜치 생성 실패

**증상:**
```
Error: Failed to create branch feature/spec-auth-001
fatal: A branch named 'feature/spec-auth-001' already exists.
```

**해결 방법:**

```bash
# 방법 1: 기존 브랜치로 전환
git checkout feature/spec-auth-001
/moai:1-spec SPEC-AUTH-001 "업데이트"

# 방법 2: 기존 브랜치 삭제 후 재생성
git branch -D feature/spec-auth-001
/moai:1-spec "사용자 인증"

# 방법 3: 브랜치 생성 스킵
/moai:1-spec "사용자 인증" --no-branch
```

#### 문제 3: GitHub API 권한 오류 (Team 모드)

**증상:**
```
Error: GitHub API authentication failed
403 Forbidden
```

**해결 방법:**

```bash
# GitHub 토큰 확인
gh auth status

# 재인증
gh auth login

# 권한 범위 확인 (repo, issues, pull_requests 필요)
gh auth refresh -s repo -s issues

# Personal 모드로 전환 (임시)
/moai:1-spec "기능" --mode personal
```

#### 문제 4: EARS 구문 자동 생성 실패

**증상:**
```
Warning: EARS requirements are incomplete
```

**해결 방법:**

```bash
# 더 상세한 컨텍스트 제공
/moai:1-spec "사용자가 이메일과 비밀번호로 로그인하고, JWT 토큰을 받아 인증된 상태로 API를 사용할 수 있어야 함"

# 프로젝트 문서 먼저 생성
/moai:8-project
/moai:1-spec "사용자 인증"

# 수동으로 EARS 구문 추가 (SPEC 파일 편집)
```

### /moai:2-build 문제 해결

#### 문제 1: SPEC을 찾을 수 없음

**증상:**
```
Error: SPEC-001 not found
```

**해결 방법:**

```bash
# SPEC 목록 확인
moai status

# SPEC 파일 존재 확인
ls -la .moai/specs/

# SPEC 재생성
/moai:1-spec "기능 제목"
```

#### 문제 2: 테스트 실패 (RED Phase)

**증상:**
```
Error: Tests are already passing
Expected: FAIL
Actual: PASS
```

**해결 방법:**

```bash
# 기존 구현 제거
rm src/auth/service.ts

# RED Phase 재실행
/moai:2-build SPEC-AUTH-001 --phase red

# 또는 강제 실행
/moai:2-build SPEC-AUTH-001 --phase red --force
```

#### 문제 3: 테스트 통과 실패 (GREEN Phase)

**증상:**
```
FAIL  __tests__/auth/service.test.ts
  ✗ should authenticate valid user
    Expected: true
    Received: false
```

**해결 방법:**

```bash
# 1. 테스트 로그 확인
npm test -- --verbose

# 2. debug-helper 에이전트 호출
@agent-debug-helper "테스트 실패: should authenticate valid user"

# 3. 수동 디버깅
node --inspect-brk node_modules/.bin/vitest run

# 4. Phase 재실행
/moai:2-build SPEC-AUTH-001 --phase green --retry
```

#### 문제 4: 커버리지 임계값 미달

**증상:**
```
Error: Coverage 78% < threshold 85%
```

**해결 방법:**

```bash
# 커버리지 리포트 확인
npm test -- --coverage

# 누락된 케이스 추가 (debug-helper 활용)
@agent-debug-helper "커버리지 분석"

# 임시로 임계값 낮춤 (비권장)
/moai:2-build SPEC-001 --coverage 75

# 설정 파일 업데이트
# .moai/config.json
{
  "quality": {
    "coverage_threshold": 75
  }
}
```

#### 문제 5: TRUST 검증 실패

**증상:**
```
Error: TRUST score 68% < threshold 80%
- Test First: 60% ❌
- Readable: 100% ✓
- Unified: 80% ✓
- Secured: 50% ❌
- Trackable: 90% ✓
```

**해결 방법:**

```bash
# trust-checker 에이전트로 상세 분석
@agent-trust-checker "SPEC-AUTH-001 검증"

# 개별 원칙 개선
# T: 테스트 추가
npm test -- --coverage

# S: 보안 검증 추가
# - 입력 검증
# - 에러 처리
# - 민감정보 마스킹

# 재검증
/moai:2-build SPEC-001 --phase refactor
```

#### 문제 6: 언어 감지 실패

**증상:**
```
Error: Unable to detect project language
```

**해결 방법:**

```bash
# 수동으로 언어 지정
/moai:2-build SPEC-001 --language typescript

# 언어 감지 설정 업데이트
# .moai/config.json
{
  "language": {
    "primary": "typescript",
    "test_framework": "vitest"
  }
}

# 프로젝트 구조 확인
moai doctor
```

#### 문제 7: 병렬 구현 충돌

**증상:**
```
Error: Dependency conflict detected
SPEC-002 depends on SPEC-001 (not completed)
```

**해결 방법:**

```bash
# 의존성 순서대로 실행
/moai:2-build SPEC-001 SPEC-002 --parallel --respect-deps

# 또는 순차 실행
/moai:2-build SPEC-001
/moai:2-build SPEC-002

# 의존성 그래프 확인
moai status --graph
```

### /moai:3-sync 문제 해결

#### 문제 1: TAG 체인 불완전

**증상:**
```
⚠️ Incomplete TAG chains detected:
- AUTH-001: Missing @TEST
- USER-002: Missing @SPEC
```

**해결 방법:**

```bash
# 자동 수정 시도
/moai:3-sync --auto-fix

# 또는 수동 수정
# 1. 누락된 TAG 추가
rg "@CODE:AUTH-001" -l  # 파일 찾기
# 파일에 @TEST:AUTH-001 추가

# 2. 재검증
/moai:3-sync --mode tags

# 3. tag-agent로 상세 분석
@agent-tag-agent "TAG 체인 검증"
```

#### 문제 2: 고아 TAG 발견

**증상:**
```
⚠️ Orphan TAGs detected:
- @SPEC:PAYMENT-005 (no related TAGs)
```

**해결 방법:**

```bash
# 고아 TAG 파일 찾기
rg "@SPEC:PAYMENT-005" -n

# 옵션 1: TAG 체인 완성
# 파일에 @SPEC, @CODE, @TEST 추가

# 옵션 2: TAG 제거 (더 이상 사용 안 함)
# @SPEC:PAYMENT-005 삭제

# 옵션 3: DEPRECATED 표시
# @SPEC:PAYMENT-005:DEPRECATED

# 재검증
/moai:3-sync --mode tags
```

#### 문제 3: 문서 동기화 충돌

**증상:**
```
Error: Merge conflict in README.md
```

**해결 방법:**

```bash
# 충돌 파일 확인
git status

# 수동 해결
vi README.md
# 충돌 해결 후...
git add README.md

# 동기화 재시도
/moai:3-sync --mode docs

# 또는 강제 동기화 (주의!)
/moai:3-sync --force
```

#### 문제 4: PR 전환 실패

**증상:**
```
Error: Cannot transition PR to Ready
- Incomplete TAG chains: 2
- TRUST score below threshold: 75%
```

**해결 방법:**

```bash
# 1. TAG 체인 완성
/moai:3-sync --mode tags --auto-fix

# 2. TRUST 개선
@agent-trust-checker "전체 검증"

# 3. 재시도
/moai:3-sync

# 4. 강제 전환 (비권장)
/moai:3-sync --force-ready
```

#### 문제 5: Living Document 생성 오류

**증상:**
```
Error: Failed to update CHANGELOG.md
Template parsing error
```

**해결 방법:**

```bash
# 템플릿 확인
cat .moai/templates/changelog.md

# 템플릿 복구
moai restore --templates

# 수동 생성 스킵
/moai:3-sync --no-docs

# 또는 doc-syncer 에이전트 직접 호출
@agent-doc-syncer "CHANGELOG 수동 갱신"
```

### 공통 문제 해결

#### 문제: 명령어가 인식되지 않음

**증상:**
```
Unknown command: /moai:1-spec
```

**해결 방법:**

```bash
# Claude Code 설정 확인
cat .claude/settings.json

# moai init 재실행
moai init --force

# 또는 수동 설정
# .claude/settings.json에 명령어 등록 확인
```

#### 문제: 권한 오류

**증상:**
```
Error: Permission denied
EACCES: permission denied
```

**해결 방법:**

```bash
# 파일 권한 확인
ls -la .moai/

# 권한 수정
chmod -R u+w .moai/

# Git 권한 확인
git config --list | grep user
```

#### 문제: 성능 저하 (느린 실행)

**증상:**
```
/moai:3-sync takes 5+ minutes
```

**해결 방법:**

```bash
# 1. 특정 경로만 스캔
/moai:3-sync --path src/

# 2. 캐시 활성화
# .moai/config.json
{
  "performance": {
    "enable_cache": true,
    "cache_ttl": 300
  }
}

# 3. 병렬 처리
/moai:3-sync --parallel --max-workers 4

# 4. 불필요한 파일 제외
# .moai/config.json
{
  "sync": {
    "exclude": ["dist/", "node_modules/", "*.log"]
  }
}
```

### 디버깅 도구

#### 로그 레벨 조정

```bash
# 상세 로그 활성화
export MOAI_LOG_LEVEL=debug
/moai:1-spec "테스트"

# 또는 명령어 옵션
/moai:1-spec "테스트" --verbose

# 로그 파일 확인
tail -f .moai/logs/moai.log
```

#### 드라이런 모드

```bash
# 실제 변경 없이 미리보기
/moai:1-spec "기능" --dry-run
/moai:2-build SPEC-001 --dry-run
/moai:3-sync --dry-run
```

#### 시스템 진단

```bash
# 전체 진단
moai doctor

# 고급 진단
moai doctor --advanced

# 특정 항목만
moai doctor --check git,node,typescript
```

### 긴급 복구

#### 전체 롤백

```bash
# 최근 백업으로 복구
moai restore

# 특정 시점으로 복구
moai restore --timestamp 2024-01-15-14-30

# Git 히스토리로 복구
git reflog
git reset --hard HEAD@{5}
```

#### 설정 초기화

```bash
# 설정 백업
cp -r .moai .moai.backup

# 초기화
rm -rf .moai
moai init

# 또는
moai init --reset
```

## 실전 시나리오

실제 프로젝트에서 자주 발생하는 복합적인 상황과 해결 방법을 소개합니다.

### 시나리오 1: 새 프로젝트 시작 (0→100)

**상황:** 완전히 새로운 프로젝트를 MoAI-ADK로 시작

**전체 워크플로우:**

```bash
# 1. MoAI 초기화
cd my-new-project
moai init

# 2. 시스템 진단
moai doctor

# 3. 프로젝트 비전 수립
/moai:8-project

# (대화형으로 프로젝트 정보 입력)

# 4. 첫 번째 SPEC 작성
/moai:1-spec "사용자 인증"

# (SPEC 검토 및 승인)

# 5. TDD 구현
/moai:2-build SPEC-AUTH-001

# (RED → GREEN → REFACTOR)

# 6. 문서 동기화
/moai:3-sync

# 7. 반복
/moai:1-spec "사용자 프로필 관리"
/moai:2-build SPEC-USER-001
/moai:3-sync
```

**예상 소요 시간:**
- 초기 설정: 10분
- SPEC 작성: 15분/기능
- TDD 구현: 30-60분/기능
- 동기화: 5분

### 시나리오 2: 기존 프로젝트에 MoAI 도입

**상황:** 이미 진행 중인 프로젝트에 MoAI-ADK 적용

**단계별 마이그레이션:**

```bash
# Phase 1: 초기 설정 (위험도 낮음)
moai init --existing
moai doctor

# Phase 2: 프로젝트 문서 작성
/moai:8-project --update

# Phase 3: 기존 코드 분석 및 TAG 추가
# 기존 기능 1개 선정
/moai:1-spec "기존 인증 시스템 문서화"

# Phase 4: TAG 체인 구축
# 기존 코드에 수동으로 TAG 추가
# src/auth/service.ts

# @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 ->  -> @CODE:AUTH-001 -> @TEST:AUTH-001
# Related: @CODE:AUTH-001:API
class AuthService {
  // 기존 코드...
}

# Phase 5: 검증
/moai:3-sync --mode tags --dry-run
/moai:3-sync --mode tags

# Phase 6: 새 기능부터 정식 워크플로우 적용
/moai:1-spec "OAuth2 소셜 로그인"
/moai:2-build SPEC-AUTH-002
/moai:3-sync
```

**주의사항:**
- 한 번에 모든 코드에 TAG 추가하지 말 것
- 새 기능부터 정식 워크플로우 적용
- 기존 기능은 점진적으로 TAG 추가

### 시나리오 3: 멀티 SPEC 병렬 개발

**상황:** 여러 팀원이 동시에 다른 기능 개발

**Team 모드 워크플로우:**

```bash
# 팀원 A: 인증 기능
/moai:1-spec "JWT 인증" --assignee @teamA --milestone v1.0
# → SPEC-AUTH-001, feature/spec-auth-001 브랜치, GitHub Issue 생성

# 팀원 B: 결제 기능
/moai:1-spec "결제 시스템" --assignee @teamB --milestone v1.0
# → SPEC-PAY-001, feature/spec-pay-001 브랜치, GitHub Issue 생성

# 팀원 C: 알림 기능
/moai:1-spec "이메일 알림" --assignee @teamC --milestone v1.0 --depends-on SPEC-AUTH-001
# → SPEC-NOTIF-001, 의존성 명시

# 병렬 구현 (각 브랜치에서)
# 팀원 A:
/moai:2-build SPEC-AUTH-001
/moai:3-sync

# 팀원 B:
/moai:2-build SPEC-PAY-001
/moai:3-sync

# 팀원 C (AUTH-001 완료 후):
/moai:2-build SPEC-NOTIF-001
/moai:3-sync

# 통합 (develop 브랜치)
git checkout develop
git merge feature/spec-auth-001
git merge feature/spec-pay-001
git merge feature/spec-notif-001

# 전체 TAG 검증
/moai:3-sync --mode tags
```

**충돌 방지 전략:**
- SPEC 작성 시 의존성 명시
- 브랜치별 독립적인 모듈 개발
- 통합 전 TAG 체인 검증

### 시나리오 4: 긴급 핫픽스 (우회 워크플로우)

**상황:** 프로덕션 심각한 버그, 즉시 수정 필요

**빠른 핫픽스 워크플로우:**

```bash
# 1. 긴급 브랜치 생성
git checkout -b hotfix/security-fix main

# 2. 최소 SPEC (브랜치 생성 스킵)
/moai:1-spec "XSS 취약점 수정" --priority critical --ready --no-branch

# 3. 빠른 구현 (TDD 스킵 가능, 단 테스트는 필수!)
/moai:2-build SPEC-HOTFIX-001 --phase green --skip-trust

# 4. 최소 검증
npm test
npm run lint

# 5. 수동 TAG 추가
# 파일에 @CODE:HOTFIX-001, @TEST:HOTFIX-001 추가

# 6. 즉시 머지
git add .
git commit -m "hotfix: XSS vulnerability fix (SPEC-HOTFIX-001)"
git push origin hotfix/security-fix

# 7. PR 생성 및 긴급 머지
gh pr create --title "hotfix: XSS fix" --base main
gh pr merge --auto --squash

# 8. 나중에 보완
# develop 브랜치에서 SPEC 보완
git checkout develop
/moai:1-spec SPEC-HOTFIX-001 "테스트 추가 및 리팩토링"
/moai:2-build SPEC-HOTFIX-001 --phase refactor
/moai:3-sync
```

**주의사항:**
- 긴급 상황에만 사용
- 반드시 나중에 정식 워크플로우로 보완
- TRUST 원칙 최소한 준수 (특히 T, S)

### 시나리오 5: 대규모 리팩토링

**상황:** 레거시 코드베이스 전체 리팩토링

**단계별 리팩토링 전략:**

```bash
# 1. 현재 상태 스냅샷
moai doctor --advanced > refactoring-baseline.txt
/moai:3-sync --mode status > tag-baseline.txt
git tag refactor-start

# 2. 리팩토링 SPEC 작성
/moai:1-spec "아키텍처 개선: 레이어 분리"
/moai:1-spec "타입 안전성 강화"
/moai:1-spec "테스트 커버리지 85% 달성"

# 3. 순차적 리팩토링 (기능 단위)
# 3-1. 인증 모듈
/moai:2-build SPEC-REFACTOR-001 --phase refactor
/moai:3-sync --mode tags

# 3-2. 사용자 모듈
/moai:2-build SPEC-REFACTOR-002 --phase refactor
/moai:3-sync --mode tags

# 3-3. 결제 모듈
/moai:2-build SPEC-REFACTOR-003 --phase refactor
/moai:3-sync --mode tags

# 4. 중간 검증
npm test
npm run build
/moai:3-sync --mode full

# 5. TRUST 원칙 검증
@agent-trust-checker "전체 프로젝트 검증"

# 6. 비교 리포트
moai doctor --advanced > refactoring-after.txt
diff refactoring-baseline.txt refactoring-after.txt

# 7. 롤백 포인트 생성
git tag refactor-complete
```

**성공 기준:**
- 테스트 커버리지 ≥85%
- TRUST 점수 ≥80%
- TAG 체인 완결성 100%
- 빌드 시간 변화 ±10% 이내

### 시나리오 6: CI/CD 통합

**상황:** GitHub Actions에 MoAI 워크플로우 통합

**.github/workflows/moai-ci.yml:**

```yaml
name: MoAI CI/CD

on:
  pull_request:
    branches: [develop, main]
  push:
    branches: [develop]

jobs:
  spec-validation:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: oven-sh/setup-bun@v1

      - name: Install MoAI-ADK
        run: bun install -g moai-adk

      - name: System Check
        run: moai doctor

      - name: SPEC Validation
        run: moai status --format json > spec-status.json

      - name: Upload SPEC Status
        uses: actions/upload-artifact@v3
        with:
          name: spec-status
          path: spec-status.json

  tdd-implementation:
    needs: spec-validation
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: oven-sh/setup-bun@v1

      - name: Install Dependencies
        run: bun install

      - name: Auto Build (if SPEC is Draft)
        run: |
          DRAFT_SPECS=$(moai status --format json | jq -r '.specs[] | select(.status=="draft") | .id')
          for SPEC in $DRAFT_SPECS; do
            /moai:2-build $SPEC --auto-approve --coverage 85
          done

      - name: Run Tests
        run: bun test --coverage

      - name: TRUST Validation
        run: |
          # trust-checker 검증
          bun run trust-check

  tag-sync:
    needs: tdd-implementation
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: TAG Chain Validation
        run: /moai:3-sync --mode tags --dry-run

      - name: Auto Fix TAG Issues
        run: /moai:3-sync --mode tags --auto-fix

      - name: Generate Sync Report
        run: /moai:3-sync --format json > sync-report.json

      - name: Comment PR
        uses: actions/github-script@v6
        with:
          script: |
            const report = require('./sync-report.json');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `## MoAI Sync Report\n\n${report.summary}`
            })
```

**로컬 개발 워크플로우:**

```bash
# 1. SPEC 작성 (로컬)
/moai:1-spec "새 기능"

# 2. 로컬 구현
/moai:2-build SPEC-001 --dry-run  # 계획 확인
/moai:2-build SPEC-001

# 3. 로컬 검증
npm test
/moai:3-sync --dry-run

# 4. PR 생성
git push origin feature/spec-001
gh pr create --title "feat: 새 기능 (SPEC-001)"

# 5. CI 자동 실행
# - SPEC 검증
# - TAG 체인 검증
# - TRUST 검증
# - PR에 리포트 자동 댓글

# 6. 승인 후 머지
# CI 통과 → 리뷰어 승인 → 자동 머지
```

### 시나리오 7: 멀티 언어 프로젝트

**상황:** TypeScript + Python + Rust 혼합 프로젝트

**프로젝트 구조:**

```
my-project/
├── frontend/          # TypeScript (React)
├── backend/           # Python (FastAPI)
├── engine/            # Rust (성능 크리티컬)
└── .moai/
    └── config.json
```

**설정 파일 (.moai/config.json):**

```json
{
  "languages": {
    "primary": "typescript",
    "secondary": ["python", "rust"]
  },
  "workspaces": {
    "frontend": {
      "language": "typescript",
      "test": "vitest",
      "path": "frontend/"
    },
    "backend": {
      "language": "python",
      "test": "pytest",
      "path": "backend/"
    },
    "engine": {
      "language": "rust",
      "test": "cargo test",
      "path": "engine/"
    }
  }
}
```

**워크플로우:**

```bash
# 1. 프론트엔드 기능
/moai:1-spec "사용자 대시보드 UI" --workspace frontend
/moai:2-build SPEC-UI-001 --workspace frontend
# → Vitest, Biome 자동 사용

# 2. 백엔드 API
/moai:1-spec "대시보드 데이터 API" --workspace backend
/moai:2-build SPEC-API-001 --workspace backend
# → pytest, mypy, ruff 자동 사용

# 3. 성능 엔진
/moai:1-spec "데이터 처리 엔진" --workspace engine
/moai:2-build SPEC-ENGINE-001 --workspace engine
# → cargo test, clippy 자동 사용

# 4. 통합 동기화
/moai:3-sync --all-workspaces

# TAG 체인이 언어 경계를 넘어 연결됨:
# @SPEC:DASHBOARD-001 (공통)
#   ├── @CODE:DASHBOARD-001:UI (TypeScript)
#   ├── @CODE:DASHBOARD-001:API (Python)
#   └── @CODE:DASHBOARD-001:DATA (Rust)
```

### 시나리오 8: 롤백 및 복구

**상황:** /moai:2-build 실행 후 문제 발생, 이전 상태로 복구 필요

**복구 전략:**

```bash
# 상황: SPEC-015 구현 중 심각한 문제 발견

# 방법 1: Git 롤백
git status
git log --oneline -5
git reset --hard HEAD~3  # 3개 커밋 롤백
git push origin feature/spec-015 --force

# 방법 2: MoAI 백업 복구
moai restore --list
# 출력:
# 1. 2024-01-15-14-30 (before SPEC-015)
# 2. 2024-01-15-12-00 (before SPEC-014)

moai restore --timestamp 2024-01-15-14-30

# 방법 3: SPEC 재작성
/moai:1-spec SPEC-015 "요구사항 재정의"
# 기존 SPEC-015를 수정

# 방법 4: 새로운 접근
/moai:1-spec "대체 접근: [기능명]"
# SPEC-016으로 새로 시작
```

**예방책:**

```bash
# 중요 작업 전 체크포인트 생성
git tag checkpoint-before-spec-015
moai doctor > health-check.txt

# 드라이런으로 미리 확인
/moai:2-build SPEC-015 --dry-run

# Phase별 점진적 구현
/moai:2-build SPEC-015 --phase red
# 확인 후...
/moai:2-build SPEC-015 --phase green
# 확인 후...
/moai:2-build SPEC-015 --phase refactor
```

## Personal vs Team 모드 상세 비교

MoAI-ADK는 두 가지 모드를 제공하며, 프로젝트 규모와 협업 방식에 따라 선택할 수 있습니다.

### 모드 비교표

| 항목 | Personal 모드 | Team 모드 |
|------|---------------|-----------|
| **SPEC 저장** | 로컬 `.moai/specs/` | GitHub Issues + 로컬 |
| **브랜치 관리** | 로컬 Git | GitHub 브랜치 + PR |
| **가시성** | 로컬만 | 팀 전체 |
| **이슈 추적** | 로컬 파일 | GitHub Issues |
| **PR 연동** | 수동 | 자동 생성 |
| **담당자 지정** | 불가 | 가능 |
| **마일스톤** | 불가 | 가능 |
| **라벨** | 불가 | 가능 |
| **알림** | 없음 | GitHub 알림 |
| **코드 리뷰** | 수동 | GitHub PR |
| **CI/CD 통합** | 수동 | 자동 |
| **설정 복잡도** | 낮음 | 중간 |
| **GitHub 토큰** | 불필요 | 필수 |

### Personal 모드 상세

#### 언제 사용하나?

✅ **권장 상황:**
- 개인 프로젝트
- 프로토타이핑
- 학습 및 실험
- 1-2명 소규모 프로젝트
- 오프라인 작업

❌ **비권장 상황:**
- 3명 이상 팀 협업
- 공식 제품 개발
- 복잡한 의존성 관리 필요

#### 워크플로우 예시

```bash
# 1. SPEC 작성 (로컬 저장)
/moai:1-spec "사용자 인증"
# 저장 위치: .moai/specs/SPEC-AUTH-001/

# 2. 로컬 브랜치 생성
# 자동: feature/spec-auth-001

# 3. TDD 구현
/moai:2-build SPEC-AUTH-001

# 4. 동기화 (로컬 문서만)
/moai:3-sync

# 5. 수동 PR 생성
git push origin feature/spec-auth-001
gh pr create --title "feat: 사용자 인증 (SPEC-AUTH-001)"
```

#### 설정 (.moai/config.json)

```json
{
  "mode": "personal",
  "git_strategy": {
    "branch_prefix": "feature/",
    "auto_branch": true,
    "auto_pr": false
  },
  "storage": {
    "specs": "local",
    "sync": "local"
  }
}
```

#### 장점

- **빠른 시작**: GitHub 설정 불필요
- **오프라인 작업**: 인터넷 없이 사용
- **단순함**: 복잡한 설정 없음
- **프라이버시**: 로컬에만 저장

#### 단점

- **팀 가시성 부족**: 다른 팀원이 SPEC 볼 수 없음
- **수동 PR**: PR 수동 생성 필요
- **이슈 추적 어려움**: 로컬 파일만 의존
- **알림 없음**: SPEC 변경 알림 없음

### Team 모드 상세

#### 언제 사용하나?

✅ **권장 상황:**
- 3명 이상 팀 프로젝트
- 공식 제품/서비스 개발
- 복잡한 의존성 관리
- 코드 리뷰 프로세스 필수
- CI/CD 파이프라인 통합
- 원격 근무 환경

❌ **비권장 상황:**
- 개인 학습 프로젝트
- 빠른 프로토타이핑
- 오프라인 작업
- GitHub 접근 불가

#### 워크플로우 예시

```bash
# 1. SPEC 작성 (GitHub Issue + 로컬 저장)
/moai:1-spec "사용자 인증" --assignee @john --milestone v2.0 --labels feature,high-priority

# 생성 결과:
# - GitHub Issue #42
# - 로컬 .moai/specs/SPEC-AUTH-001/
# - GitHub 브랜치 feature/spec-auth-001
# - Draft PR #43

# 2. TDD 구현
/moai:2-build SPEC-AUTH-001

# 3. 동기화 (GitHub + 로컬)
/moai:3-sync

# 자동 수행:
# - Draft PR → Ready for Review
# - 라벨 업데이트
# - 리뷰어 알림
# - CI/CD 트리거
```

#### 설정 (.moai/config.json)

```json
{
  "mode": "team",
  "git_strategy": {
    "branch_prefix": "feature/",
    "auto_branch": true,
    "auto_pr": true,
    "pr_template": ".moai/templates/pr.md",
    "reviewers": ["@team-lead", "@senior-dev"],
    "labels": ["moai", "spec-based"]
  },
  "storage": {
    "specs": "github",
    "sync": "github",
    "issues": true
  },
  "github": {
    "repo": "org/project",
    "token_env": "GITHUB_TOKEN"
  }
}
```

#### GitHub 토큰 설정

```bash
# 1. GitHub Personal Access Token 생성
# https://github.com/settings/tokens
# 권한: repo, issues, pull_requests

# 2. 환경 변수 설정
export GITHUB_TOKEN=ghp_xxxxxxxxxxxxx

# 또는 .env 파일
echo "GITHUB_TOKEN=ghp_xxxxxxxxxxxxx" >> .env

# 3. 검증
gh auth status
```

#### 장점

- **팀 가시성**: 모든 SPEC을 GitHub에서 확인
- **자동화**: PR, 라벨, 알림 자동 처리
- **이슈 추적**: GitHub Issues 통합
- **코드 리뷰**: PR 기반 리뷰 프로세스
- **CI/CD**: 자동 파이프라인 트리거
- **협업**: 담당자, 마일스톤, 라벨 관리

#### 단점

- **복잡성**: 초기 설정 필요
- **GitHub 의존**: 인터넷 필수
- **토큰 관리**: GitHub 토큰 보안 관리
- **비용**: Private 리포지토리는 유료 (팀 규모에 따라)

### 모드 전환

#### Personal → Team 전환

```bash
# 1. 현재 상태 백업
moai doctor > personal-baseline.txt
cp -r .moai .moai.personal.backup

# 2. GitHub 토큰 설정
export GITHUB_TOKEN=ghp_xxxxxxxxxxxxx
gh auth status

# 3. 설정 파일 업데이트
# .moai/config.json
{
  "mode": "team",
  "github": {
    "repo": "org/project",
    "token_env": "GITHUB_TOKEN"
  }
}

# 4. 기존 SPEC을 GitHub로 마이그레이션
moai migrate --from personal --to team

# 생성 결과:
# - 각 SPEC → GitHub Issue
# - 로컬 SPEC 파일 유지
# - 메타데이터 동기화

# 5. 검증
moai status --verbose
```

#### Team → Personal 전환

```bash
# 1. GitHub 연동 해제 확인
# (주의: GitHub Issues는 삭제되지 않음, 로컬만 전환)

# 2. 설정 파일 업데이트
# .moai/config.json
{
  "mode": "personal",
  "storage": {
    "specs": "local"
  }
}

# 3. GitHub 메타데이터 로컬 복사
moai migrate --from team --to personal

# 4. 검증
moai status --verbose

# 5. 주의: 기존 GitHub Issues/PR은 수동 관리 필요
```

### 하이브리드 모드 (고급)

**상황:** 팀 프로젝트지만 일부 SPEC은 로컬만 관리하고 싶음

```bash
# 기본은 Team 모드
# .moai/config.json
{
  "mode": "team"
}

# 특정 SPEC만 Personal 모드
/moai:1-spec "실험적 기능" --mode personal --no-branch

# 또는 Private SPEC (팀원 중 일부만)
/moai:1-spec "내부 리팩토링" --private --assignee @me
```

### 모드 선택 가이드

#### 프로젝트 규모별

**1명 프로젝트:**
- Personal 모드 권장
- 간단하고 빠른 시작

**2-3명 프로젝트:**
- Personal 모드: 비공식 협업
- Team 모드: 공식 프로젝트

**4명 이상 프로젝트:**
- Team 모드 필수
- GitHub 이슈 추적 필요

#### 프로젝트 단계별

**프로토타입 단계:**
```bash
# Personal 모드로 빠르게 시작
moai init --mode personal
```

**MVP 단계:**
```bash
# Team 모드로 전환
moai migrate --to team
```

**프로덕션 단계:**
```bash
# Team 모드 + CI/CD 통합
# .github/workflows/moai-ci.yml 설정
```

### 비용 및 리소스 비교

#### Personal 모드

- **비용**: 무료
- **GitHub**: 불필요
- **설정 시간**: 5분
- **유지 관리**: 최소

#### Team 모드

- **비용**: GitHub 유료 플랜 필요 (Private 리포지토리)
  - Free: Public 리포지토리만
  - Team: $4/user/month
  - Enterprise: $21/user/month

- **GitHub 토큰**: 필수
- **설정 시간**: 20-30분
- **유지 관리**: 중간 (토큰, 권한 관리)

### 실전 권장사항

#### 개인 개발자

```bash
# 1. Personal 모드로 시작
moai init --mode personal

# 2. 프로젝트 성장 시 Team 전환
# - GitHub 리포지토리 Public으로 변경
# - 기여자 3명 이상 도달
# - 이슈 추적 필요
moai migrate --to team
```

#### 스타트업 팀

```bash
# 처음부터 Team 모드
moai init --mode team --github org/project

# CI/CD 조기 통합
# .github/workflows/moai-ci.yml 설정
```

#### 대기업 조직

```bash
# Enterprise GitHub + Team 모드
moai init --mode team --github enterprise/project

# 추가 설정:
# - SSO 통합
# - Advanced Security
# - Compliance 정책
```

### 자주 묻는 질문 (FAQ)

**Q: Personal에서 Team으로 전환 시 데이터 손실 있나요?**

A: 없습니다. 로컬 SPEC 파일은 유지되며, GitHub에 추가로 동기화됩니다.

**Q: Team 모드에서 GitHub 토큰 만료 시?**

A: `gh auth refresh` 명령어로 갱신하거나, 새 토큰을 발급하여 환경 변수를 업데이트합니다.

**Q: Personal 모드에서도 PR 자동 생성 가능한가요?**

A: 아니요. PR 자동 생성은 Team 모드 전용입니다. Personal 모드는 수동 PR 생성이 필요합니다.

**Q: 프로젝트마다 다른 모드 사용 가능한가요?**

A: 네. 각 프로젝트의 `.moai/config.json`에서 독립적으로 설정 가능합니다.

**Q: Team 모드 비용이 부담스러운데 대안은?**

A: Public 리포지토리는 무료이므로, 오픈소스 프로젝트라면 Team 모드를 무료로 사용 가능합니다.

## 참고 자료

- **명령어 소스**: `.claude/commands/moai/`
- **설정 파일**: `.claude/settings.json`
- **커스터마이징**: [고급 가이드](/advanced/custom-commands)