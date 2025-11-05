# 1단계: 계획 (Plan)

`/alfred:1-plan`은 MoAI-ADK 워크플로우의 첫 번째 단계로, 명확한 요구사항을 정의하고 실행 계획을 수립하는 과정입니다. Alfred의 spec-builder가 EARS 문법을 사용하여 전문적인 SPEC을 자동으로 생성합니다.

## 🎯 Plan 단계 개요

Plan 단계는 다음 활동들을 포함합니다:

```mermaid
%%{init: {'theme':'neutral'}}%%
graph TD
    Start([/alfred:1-plan 실행]) --> Analyze[요구사항 분석]
    Analyze --> Expert[전문가 상담]
    Expert --> EARS[EARS 문법 적용]
    EARS --> Board[Plan Board 생성]
    Board --> Spec[SPEC 문서 작성]
    Spec --> Tag[@TAG 할당]
    Tag --> Branch[Feature 브랜치 생성]
    Branch --> End([Plan 완료])

    subgraph "핵심 산출물"
        Spec
        Board
        Tag
        Branch
    end
```

### Plan 단계의 목표

✅ **명확한 요구사항 정의**: 모호함 없는 구체적인 요구사항
✅ **기술적 계획 수립**: 아키텍처와 구현 전략
✅ **전문가 검증**: 도메인 전문가의 기술적 조언
✅ **추적 가능성**: @TAG를 통한 요구사항 추적
✅ **실행 준비**: 개발을 바로 시작할 수 있는 상태

## 🔧 명령어 사용법

### 기본 형식

```bash
/alfred:1-plan "기능에 대한 간단한 설명"
```

### 실제 사용 예시

```bash
# 기본 예시
/alfred:1-plan "사용자 인증 기능"

# 상세한 설명
/alfred:1-plan "JWT 기반 사용자 인증 API - 이메일/비밀번호 로그인, 토큰 리프레시"

# 복잡한 기능
/alfred:1-plan "사용자 관리 시스템 - 회원가입, 로그인, 프로필 관리, 권한 제어"
```

### Alfred의 응답 구조

Alfred는 Plan 단계에서 다음 정보들을 제공합니다:

```
🎯 Plan 단계를 시작하겠습니다.

1️⃣ 요구사항 분석 중...
   - 기능: 사용자 인증
   - 복잡도: 중간
   - 예상 개발 시간: 2-3일

2️⃣ 전문가 상담...
   - backend-expert 활성화 (인증 키워드 감지)
   - 아키텍처 추천: JWT + bcrypt

3️⃣ SPEC 작성 중...
   - SPEC ID: AUTH-001
   - EARS 문법 적용 완료
   - Plan Board 생성 완료

4️⃣ 결과:
   ✅ .moai/specs/SPEC-AUTH-001/spec.md
   ✅ .moai/specs/SPEC-AUTH-001/plan.md
   ✅ feature/SPEC-AUTH-001 브랜치

다음 단계: /alfred:2-run AUTH-001
```

## 📋 Plan 단계 상세 과정

### 1단계: 요구사항 분석

Alfred는 사용자의 입력을 분석하여 요구사항의 복잡성과 범위를 결정합니다.

#### 분석 항목

| 분석 항목 | 설명 | Alfred의 처리 |
|----------|------|---------------|
| **기능 유형** | API, UI, 데이터 처리 등 | 구현 전략 결정 |
| **복잡도** | 단순, 중간, 복잡 | 개발 시간 예측 |
| **도메인** | 인증, 데이터, UI 등 | 전문가 활성화 |
| **의존성** | 다른 기능과의 관계 | 선행 조건 확인 |
| **위험 요소** | 기술적, 비즈니스적 리스크 | 완화 전략 수립 |

#### 복잡도별 특징

| 복잡도 | 특징 | 예상 시간 | 전문가 필요 |
|--------|------|-----------|------------|
| **단순** | 단일 기능, 외부 의존성 적음 | 1-2시간 | 선택적 |
| **중간** | 여러 컴포넌트, 데이터베이스 연동 | 0.5-1일 | 권장 |
| **복잡** | 여러 시스템 연동, 보안 고려사항 | 2-5일 | 필수 |

### 2단계: 전문가 상담

Alfred는 요구사항의 도메인을 분석하여 적절한 전문가 에이전트를 자동으로 활성화합니다.

#### 전문가 활성화 규칙

| 키워드 | 활성화 전문가 | 제공 내용 |
|--------|----------------|-----------|
| '인증', '로그인', '보안' | backend-expert, security-expert | JWT, OAuth, bcrypt 추천 |
| '데이터', '데이터베이스' | backend-expert, database-expert | 스키마 설계, 최적화 |
| 'UI', '프론트엔드', '디자인' | frontend-expert, ui-ux-expert | 컴포넌트 구조, 접근성 |
| '배포', 'Docker', '클라우드' | devops-expert | 배포 전략, CI/CD |

#### 전문가 상담 예시

```
🔧 backend-expert 의견:
- JWT 토큰 사용 추천 (무상태, 확장성)
- bcrypt로 비밀번호 해싱 (보안)
- 15분 토큰 만료 (보안 + 사용자 경험)
- 리프레시 토큰 구현 (UX 향상)

🔒 security-expert 의견:
- 비밀번호 정책: 8자 이상, 특수문자 포함
- 로그인 실패 시 5회 제한 (Brute Force 방지)
- HTTPS 필수 (전송 암호화)
- 환경변수로 시크릿 키 관리
```

### 3단계: EARS 문법 적용

Alfred는 EARS(Easy Approach to Requirements Syntax) 문법을 사용하여 명확한 요구사항을 작성합니다. EARS는 요구사항 작성을 위한 표준화된 구문으로, 모호함을 제거하고 테스트 가능한 명세를 만듭니다.

#### EARS 5가지 패턴 상세 설명

EARS는 5가지 패턴을 통해 모든 유형의 요구사항을 체계적으로 표현합니다:

##### 1. Ubiquitous Requirements (보편적 요구사항)
**목적**: 시스템의 기본적인 기능과 책임을 정의
**형식**: "시스템은 [기능]을 제공해야 한다"
**특징**: 조건 없이 항상 참이어야 하는 기본 요구사항

**예시**:
```yaml
## Ubiquitous Requirements
- 시스템은 JWT 기반 인증을 제공해야 한다
- 시스템은 비밀번호 해싱을 지원해야 한다
- 시스템은 사용자 등록을 지원해야 한다
- 시스템은 RESTful API 형식을 제공해야 한다
```

##### 2. Event-driven Requirements (이벤트 기반 요구사항)
**목적**: 특정 이벤트가 발생했을 때의 시스템 반응을 정의
**형식**: "**WHEN** [조건]이/가 발생하면, 시스템은 [행동]을 해야 한다"
**특징**: 트리거-액션 관계를 명확하게 정의

**예시**:
```yaml
## Event-driven Requirements
- **WHEN** 유효한 이메일과 비밀번호가 제공되면, 시스템은 JWT 토큰을 발급해야 한다
- **WHEN** 만료된 토큰이 제공되면, 시스템은 401 Unauthorized 에러를 반환해야 한다
- **WHEN** 리프레시 토큰이 유효하면, 시스템은 새로운 액세스 토큰을 발급해야 한다
- **WHEN** 존재하지 않는 사용자 ID로 조회하면, 시스템은 404 Not Found 에러를 반환해야 한다
```

##### 3. State-driven Requirements (상태 기반 요구사항)
**목적**: 특정 상태가 유지되는 동안의 시스템 동작을 정의
**형식**: "**WHILE** [상태]일 때, 시스템은 [행동]을 해야 한다"
**특징**: 지속적인 상태 조건에서의 동작을 정의

**예시**:
```yaml
## State-driven Requirements
- **WHILE** 사용자가 인증된 상태일 때, 시스템은 보호된 리소스에 접근을 허용해야 한다
- **WHILE** 토큰이 유효한 상태일 때, 시스템은 토큰 갱신을 허용해야 한다
- **WHILE** 사용자 계정이 잠긴 상태일 때, 시스템은 모든 로그인 시도를 거부해야 한다
- **WHILE** 시스템이 유지보수 모드일 때, 시스템은 읽기 전용 동작만 허용해야 한다
```

##### 4. Optional Requirements (선택적 요구사항)
**목적**: 특정 조건에서만 적용되는 부가적인 기능을 정의
**형식**: "**WHERE** [조건]이면, 시스템은 [기능]을 할 수 있다"
**특징**: 필수가 아닌 부가 기능이나 조건부 기능을 정의

**예시**:
```yaml
## Optional Requirements
- **WHERE** OAuth2 제공자가 구성되면, 시스템은 소셜 로그인을 지원할 수 있다
- **WHERE** 2FA가 활성화되면, 시스템은 추가 인증을 요구할 수 있다
- **WHERE** 관리자 권한이 있으면, 시스템은 사용자 관리 기능을 제공할 수 있다
- **WHERE** 분산 캐시가 구성되면, 시스템은 토큰 캐싱을 지원할 수 있다
```

##### 5. Unwanted Behaviors (바람직하지 않은 동작)
**목적**: 시스템이 해서는 안 되는 행동이나 제약 조건을 정의
**형식**: "[기능]은/는 [제약]을 해서는 안 된다"
**특징**: 부정적 제약을 통해 시스템 경계를 명확히 정의

**예시**:
```yaml
## Unwanted Behaviors
- 비밀번호는 평문으로 저장되어서는 안 된다
- 토큰 만료 시간은 15분을 초과하지 않아야 한다
- 로그인 실패는 5회로 제한되어야 한다
- 시스템은 사용자 비밀번호를 노출해서는 안 된다
- 동일한 이메일로 중복 가입을 허용해서는 안 된다
```

#### Alfred의 EARS 자동 변환 예시

**입력**: "JWT 기반 사용자 인증 API - 이메일/비밀번호 로그인, 토큰 리프레시"

**Alfred의 EARS 적용 결과**:

```yaml
# .moai/specs/SPEC-AUTH-001/spec.md

## EARS Requirements Analysis

### 🎯 기능 분석 결과
- **주요 기능**: 사용자 인증
- **핵심 플로우**: 로그인 → 토큰 발급 → 리소스 접근
- **복잡도**: 중간 (보안 고려사항 포함)
- **도메인**: 인증 및 보안

### 📋 EARS 패턴 적용

## Ubiquitous Requirements (보편적 요구사항)
- 시스템은 JWT 기반 인증을 제공해야 한다
- 시스템은 비밀번호 해싱을 지원해야 한다
- 시스템은 사용자 등록을 지원해야 한다
- 시스템은 토큰 갱신 메커니즘을 제공해야 한다

## Event-driven Requirements (이벤트 기반 요구사항)
- **WHEN** 유효한 이메일과 비밀번호가 제공되면, 시스템은 JWT 액세스 토큰과 리프레시 토큰을 발급해야 한다
- **WHEN** 만료된 액세스 토큰이 제공되면, 시스템은 401 Unauthorized 에러를 반환해야 한다
- **WHEN** 유효한 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급해야 한다
- **WHEN** 잘못된 자격증명이 제공되면, 시스템은 401 Unauthorized 에러를 반환해야 한다
- **WHEN** 5회 연속 로그인 실패 시, 시스템은 계정을 잠가야 한다

## State-driven Requirements (상태 기반 요구사항)
- **WHILE** 사용자가 인증된 상태일 때, 시스템은 보호된 API 엔드포인트에 접근을 허용해야 한다
- **WHILE** 액세스 토큰이 유효한 상태일 때, 시스템은 사용자 세션을 유지해야 한다
- **WHILE** 계정이 잠긴 상태일 때, 시스템은 모든 로그인 시도를 차단해야 한다
- **WHILE** 리프레시 토큰이 유효한 상태일 때, 시스템은 액세스 토큰 갱신을 허용해야 한다

## Optional Requirements (선택적 요구사항)
- **WHERE** Google OAuth2가 구성되면, 시스템은 구글 로그인을 지원할 수 있다
- **WHERE** 2FA가 활성화되면, 시스템은 TOTP 인증을 요구할 수 있다
- **WHERE** Redis가 구성되면, 시스템은 토큰 블랙리스팅을 지원할 수 있다
- **WHERE** 이메일 서비스가 구성되면, 시스템은 비밀번호 재설정 이메일을 발송할 수 있다

## Unwanted Behaviors (바람직하지 않은 동작)
- 비밀번호는 평문으로 저장되어서는 안 된다
- 액세스 토큰 만료 시간은 15분을 초과하지 않아야 한다
- 리프레시 토큰 만료 시간은 7일을 초과하지 않아야 한다
- 시스템은 JWT 시크릿 키를 로그에 기록해서는 안 된다
- 동일 IP에서 5회 이상 로그인 실패 시 계정을 잠그지 않으면 안 된다
```

#### EARS 문법의 장점

1. **명확성**: 모호한 표현을 제거하고 명확한 요구사항 정의
2. **테스트 가능성**: 각 요구사항이 자동으로 테스트 케이스로 변환
3. **일관성**: 표준화된 형식으로 일관된 명세 작성
4. **완전성**: 5가지 패턴으로 모든 유형의 요구사항 표현
5. **추적성**: 각 요구사항이 구현 및 테스트와 직접 연결

#### Alfred의 EARS 자동화 기능

Alfred는 다음과 같은 EARS 자동화 기능을 제공합니다:

1. **자동 패턴 식별**: 사용자 입력을 분석하여 적절한 EARS 패턴으로 자동 분류
2. **요구사항 확장**: 단순한 설명을 완전한 EARS 요구사항으로 자동 확장
3. **검증 규칙 적용**: EARS 문법 규칙에 맞지 않는 표현 자동 수정
4. **테스트 케이스 생성**: 각 EARS 요구사항에서 자동으로 테스트 케이스 생성
5. **커버리지 보장**: 모든 요구사항이 테스트 되도록 자동으로 커버리지 계획 수립

### 4단계: Plan Board 생성

Plan Board는 구현을 위한 상세한 계획 문서입니다.

#### Plan Board 구조

```markdown
# .moai/specs/SPEC-AUTH-001/plan.md

# Plan Board: 사용자 인증 시스템

## 개요
JWT 기반 사용자 인증 시스템 구현 계획

## 기술 스택
- **프레임워크**: FastAPI (Python)
- **데이터베이스**: PostgreSQL
- **인증**: JWT + bcrypt
- **테스트**: pytest

## 아키텍처 설계

### 컴포넌트 구조
```
auth/
├── models.py      # User 모델
├── service.py     # AuthService
├── routes.py      # API 엔드포인트
└── utils.py       # JWT 유틸리티
```

### 데이터 모델
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## 구현 계획

### Phase 1: 기본 인증 (우선순위: 높음)
- [ ] 사용자 모델 생성
- [ ] 비밀번호 해싱 구현
- [ ] JWT 토큰 생성/검증
- [ ] 로그인 API 엔드포인트

### Phase 2: 토큰 관리 (우선순위: 중간)
- [ ] 토큰 리프레시 구현
- [ ] 토큰 블랙리스팅
- [ ] 토큰 만료 처리

### Phase 3: 보안 강화 (우선순위: 중간)
- [ ] 로그인 실패 제한
- [ ] 비밀번호 정책
- [ ] 보안 헤더 설정

## 위험 요소 및 완화 전략

| 위험 요소 | 영향도 | 완화 전략 |
|-----------|--------|-----------|
| JWT 시크릿 키 노출 | 높음 | 환경변수 사용, 정기적 키 교체 |
| 비밀번호 해싱 취약점 | 높음 | bcrypt 사용, 솔트 적용 |
| 토큰 재사용 공격 | 중간 | 짧은 만료 시간, 리프레시 토큰 |

## 성공 기준
- [ ] 모든 테스트 통과 (85%+ 커버리지)
- [ ] TRUST 5원칙 준수
- [ ] 성능: 로그인 200ms 이내
- [ ] 보안: OWASP Top 10 준수

## @TAG 계획
- @SPEC:EX-AUTH-001 ✅ 할당됨
- @TEST:EX-AUTH-001 (예정)
- @CODE:EX-AUTH-001:MODEL (예정)
- @CODE:EX-AUTH-001:SERVICE (예정)
- @CODE:EX-AUTH-001:ROUTES (예정)
- @DOC:EX-AUTH-001 (예정)
```

### 5단계: @TAG 할당

Alfred는 모든 산출물에 추적 가능한 @TAG를 자동으로 할당합니다.

#### @TAG 할당 규칙

| 산출물 | TAG 형식 | 예시 |
|--------|----------|------|
| **SPEC 문서** | `@SPEC:EX-{DOMAIN}-{ID}` | `@SPEC:EX-AUTH-001` |
| **Plan Board** | `@PLAN:EX-{DOMAIN}-{ID}` | `@PLAN:EX-AUTH-001` |
| **향상 테스트** | `@TEST:EX-{DOMAIN}-{ID}` | `@TEST:EX-AUTH-001` |
| **향상 코드** | `@CODE:EX-{DOMAIN}-{ID}:{TYPE}` | `@CODE:EX-AUTH-001:SERVICE` |

#### TAG 체인 계획

```
@SPEC:EX-AUTH-001 (현재 단계)
    ↓ (다음 단계에서 할당 예정)
@TEST:EX-AUTH-001 (RED 단계)
    ↓
@CODE:EX-AUTH-001:MODEL (GREEN 단계)
    ↓
@CODE:EX-AUTH-001:SERVICE (GREEN 단계)
    ↓
@CODE:EX-AUTH-001:ROUTES (GREEN 단계)
    ↓
@DOC:EX-AUTH-001 (SYNC 단계)
```

### 6단계: Feature 브랜치 생성

Alfred는 개발을 위한 feature 브랜치를 자동으로 생성합니다.

#### 브랜치 명명 규칙

| 팀 모드 | 브랜치 형식 | 예시 |
|---------|-------------|------|
| **Personal** | `feature/SPEC-{ID}` | `feature/SPEC-AUTH-001` |
| **Team** | `feature/{description}` | `feature/user-authentication` |

#### 브랜치 생성 프로세스

```bash
# Alfred가 자동으로 실행
git checkout -b feature/SPEC-AUTH-001

# 초기 커밋 생성
git add .moai/specs/SPEC-AUTH-001/
git commit -m "📋 plan(AUTH-001): create user authentication specification

- EARS requirements with 5 patterns
- Technical architecture plan
- Risk assessment and mitigation strategies
- @SPEC:EX-AUTH-001 assigned

Co-Authored-By: 🎩 Alfred@MoAI"
```

## 🎯 Plan 단계 완료 기준

### 필수 완료 조건

✅ **SPEC 문서**: EARS 문법으로 작성된 명확한 요구사항
✅ **Plan Board**: 상세한 구현 계획과 아키텍처
✅ **@TAG 할당**: 추적 가능한 TAG 시스템
✅ **전문가 검증**: 도메인 전문가의 기술적 조언
✅ **브랜치 준비**: 개발 시작을 위한 Git 브랜치

### 품질 검증 체크리스트

```bash
# Alfred가 자동으로 검증
✅ SPEC 형식 검증 통과
✅ EARS 문법 준수
✅ Plan Board 완성도 100%
✅ @TAG 형식 올바름
✅ Git 브랜치 생성 완료
✅ 의존성 분석 완료
```

## 📝 Plan 단계 산출물 상세

### 1. SPEC 문서 (.moai/specs/SPEC-{ID}/spec.md)

```yaml
---
id: AUTH-001
version: 1.0.0
status: draft
priority: high
created: 2025-11-06
updated: 2025-11-06
author: @developer
domain: authentication
complexity: medium
estimated_hours: 16
---

# `@SPEC:EX-AUTH-001: 사용자 인증 시스템

## 개요
JWT 기반 사용자 인증 시스템으로 이메일/비밀번호 로그인과 토큰 관리를 제공합니다.

## Ubiquitous Requirements
- 시스템은 JWT 기반 인증을 제공해야 한다
- 시스템은 비밀번호 해싱을 지원해야 한다
- 시스템은 사용자 등록을 지원해야 한다

## Event-driven Requirements
- **WHEN** 유효한 이메일과 비밀번호가 제공되면, 시스템은 JWT 토큰을 발급해야 한다
- **WHEN** 만료된 토큰이 제공되면, 시스템은 401 Unauthorized 에러를 반환해야 한다
- **WHEN** 유효한 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급해야 한다

## State-driven Requirements
- **WHILE** 사용자가 인증된 상태일 때, 시스템은 보호된 리소스에 접근을 허용해야 한다
- **WHILE** 토큰이 유효한 상태일 때, 시스템은 토큰 갱신을 허용해야 한다

## Optional Requirements
- **WHERE** OAuth2 제공자가 구성되면, 시스템은 소셜 로그인을 지원할 수 있다
- **WHERE** 2FA가 활성화되면, 시스템은 추가 인증을 요구할 수 있다

## Unwanted Behaviors
- 비밀번호는 평문으로 저장되어서는 안 된다
- 토큰 만료 시간은 15분을 초과하지 않아야 한다
- 동일 IP에서 5회 이상 로그인 실패 시 계정을 잠가야 한다
- 이메일은 중복 가입을 허용하지 않아야 한다

## Acceptance Criteria
- 사용자는 이메일과 비밀번호로 로그인할 수 있다
- 시스템은 JWT 액세스 토큰(15분)과 리프레시 토큰(7일)을 발급한다
- 비밀번호는 bcrypt로 해싱된다
- 로그인 실패는 5회로 제한된다
- 모든 API 엔드포인트는 인증이 필요하다

## Dependencies
- 데이터베이스 시스템 (PostgreSQL 권장)
- 이메일 서비스 (가입 확인용)
- 환경변수 설정 (JWT 시크릿 키)

## History
- v1.0.0 - 초기 SPEC 작성 (2025-11-06)
```

### 2. Plan Board (.moai/specs/SPEC-{ID}/plan.md)

이미 위에서 상세하게 다룬 내용으로, 구현 계획, 아키텍처, 위험 분석 등을 포함합니다.

### 3. Git 상태

```bash
# 브랜치 확인
git branch
# * feature/SPEC-AUTH-001

# 커밋 히스토리
git log --oneline -1
# 📋 plan(AUTH-001): create user authentication specification
```

## 🚀 첫 10분 실습: Hello World API

이제 직접 Alfred를 사용하여 Plan 단계를 경험해보겠습니다.

### 실습 목표

- 간단한 Hello World API SPEC 작성
- EARS 문법 적용
- Plan Board 생성
- @TAG 시스템 이해

### 실습 과정

#### 1단계: Claude Code 실행

```bash
# 프로젝트 디렉토리에서 Claude Code 실행
claude
```

#### 2단계: Alfred Plan 명령 실행

```
/alfred:1-plan "GET /hello 엔드포인트 - 쿼리 파라미터 name을 받아서 인사말 반환"
```

#### 3단계: Alfred의 처리 과정 관찰

Alfred는 다음 단계들을 자동으로 처리합니다:

```
🎯 Plan 단계를 시작하겠습니다.

1️⃣ 요구사항 분석...
   - 기능: Hello World API
   - 복잡도: 단순
   - 예상 개발 시간: 30분

2️⃣ 전문가 상담...
   - backend-expert 활성화 (API 키워드 감지)
   - FastAPI 프레임워크 추천

3️⃣ SPEC 작성 중...
   - SPEC ID: HELLO-001
   - EARS 문법 적용 완료

4️⃣ 결과:
   ✅ .moai/specs/SPEC-HELLO-001/spec.md
   ✅ .moai/specs/SPEC-HELLO-001/plan.md
   ✅ feature/SPEC-HELLO-001 브랜치

다음 단계: /alfred:2-run HELLO-001
```

#### 4단계: 생성된 SPEC 확인

```bash
# SPEC 파일 확인
cat .moai/specs/SPEC-HELLO-001/spec.md

# Plan Board 확인
cat .moai/specs/SPEC-HELLO-001/plan.md
```

#### 5단계: 브랜치 상태 확인

```bash
# 현재 브랜치 확인
git branch
# * feature/SPEC-HELLO-001

# 커밋 확인
git log --oneline -1
# 📋 plan(HELLO-001): create hello world API specification
```

### 실습 결과 검증

✅ **SPEC 문서**: EARS 문법으로 작성된 요구사항
✅ **Plan Board**: FastAPI 기반 구현 계획
✅ **@TAG 할당**: `@SPEC:EX-HELLO-001`
✅ **브랜치**: `feature/SPEC-HELLO-001`
✅ **준비 완료**: 다음 단계(`/alfred:2-run`) 준비됨

## 🎯 Plan 단계 성공 팁

### 좋은 SPEC 작성 팁

1. **구체적 작성**: "사용자 기능" → "JWT 기반 사용자 인증"
2. **단일 책임**: 하나의 SPEC은 하나의 기능에 집중
3. **테스트 가능**: 각 요구사항이 테스트 가능해야 함
4. **명확한 조건**: WHEN/WHERE/WHILE을 명확하게 사용

### 전문가 활용 팁

1. **키워드 포함**: 도메인별 키워드를 포함하여 전문가 활성화
2. **구체적 질문**: 전문가에게 구체적인 기술적 질문 포함
3. **위험 노출**: 잠재적 위험 요소를 명시적으로 언급

### Plan Board 작성 팁

1. **단계별 계획**: 복잡한 기능은 Phase로 나누기
2. **성공 기준**: 명확한 완료 조건 정의
3. **위험 관리**: 각 위험에 대한 구체적인 완화 전략

## 🔍 문제 해결

### 일반적인 문제들

**문제 1**: SPEC이 너무 복잡하게 작성됨

**원인**: 요구사항이 너무 광범위함

**해결책**:
```bash
# 더 작은 단위로 분리
/alfred:1-plan "사용자 로그인 기능"
/alfred:1-plan "비밀번호 재설정 기능"
/alfred:1-plan "프로필 관리 기능"
```

**문제 2**: 전문가가 활성화되지 않음

**원인**: 적절한 키워드가 포함되지 않음

**해결책**:
```bash
# 도메인 키워드 포함
/alfred:1-plan "데이터베이스 연동 사용자 인증 API (backend, security, database)"
```

**문제 3**: Plan Board가 너무 단순함

**원인**: 복잡도가 낮게 평가됨

**해결책**:
```bash
# 복잡성을 명시적으로 지정
/alfred:1-plan "대규모 사용자 관리 시스템 - 복잡한 권한 제어와 보안 고려사항 포함"
```

## 🚀 다음 단계

Plan 단계가 완료되면 다음 단계로 진행할 수 있습니다:

- **[2단계: 실행 (Run)](2-run.md)** - TDD 개발 사이클 시작
- **[SPEC 작성 기초](../specs/basics.md)** - SPEC 문서 작성 심화
- **[EARS 문법 상세](../specs/ears.md)** - EARS 문법 마스터

## 💡 Plan 단계 핵심 요약

1. **명령어**: `/alfred:1-plan "기능 설명"`
2. **핵심 산출물**: SPEC 문서, Plan Board, @TAG, Feature 브랜치
3. **처리 시간**: 2-5분 (복잡도에 따라)
4. **품질 보증**: EARS 문법, 전문가 검증, TRUST 준수
5. **성공 기준**: 명확한 요구사항과 실행 계획 수립

---

**Plan 단계를 완료하면 개발의 방향이 완벽하게 설정됩니다!** 🎯