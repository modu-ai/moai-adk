# MoAI-ADK SPEC 시스템 완전 가이드

> **SPEC-First TDD 방법론의 핵심 문서**
>
> MoAI-ADK의 모든 개발 작업은 SPEC 문서로부터 시작됩니다.

---

## 목차

- [1 SPEC 시스템이란?](#1-spec-시스템이란)
- [2 SPEC-First TDD 철학](#2-spec-first-tdd-철학)
- [3 SPEC 워크플로우](#3-spec-워크플로우)
- [4 @TAG 시스템 통합](#4-tag-시스템-통합)
- [5 EARS 요구사항 작성법](#5-ears-요구사항-작성법)
- [6 SPEC 문서 구조](#6-spec-문서-구조)
- [7 실전 SPEC 작성 예시](#7-실전-spec-작성-예시)
- [8 SPEC 품질 기준](#8-spec-품질-기준)
- [9 SPEC 검증 및 추적](#9-spec-검증-및-추적)
- [10 SPEC 라이프사이클](#10-spec-라이프사이클)
- [11 자주 묻는 질문 (FAQ)](#11-자주-묻는-질문-faq)

---

## 1 SPEC 시스템이란?

### 1.1. 정의

**SPEC (Specification)**은 MoAI-ADK에서 사용하는 표준화된 요구사항 명세 문서입니다. 모든 코드 구현은 SPEC 문서로부터 시작되며, SPEC 없이는 코드를 작성하지 않습니다.

### 1.2. SPEC 시스템의 목적

#### 명확한 요구사항 정의
- 구현 전에 무엇을 만들지 명확히 정의
- 모호함 제거, 구현 범위 명확화
- 이해관계자 간 합의 도구

#### 추적성 보장
- 요구사항 → 테스트 → 코드 → 문서 간 완벽한 연결
- @TAG 시스템으로 모든 변경 이력 추적
- 코드의 존재 이유를 항상 설명 가능

#### TDD 기반 구현 가이드
- RED (테스트 작성) 단계의 명확한 기준 제공
- GREEN (구현) 단계에서 참조할 명세 제공
- REFACTOR (개선) 단계에서 준수할 제약 조건 명시

#### 문서화 자동화
- Living Document의 기준점
- API 문서 자동 생성의 기초
- 변경 이력 자동 추적

### 1.3. SPEC 시스템의 핵심 원칙

#### 원칙 1: CODE-FIRST
모든 SPEC은 코드와 직접 연결됩니다. 중간 캐시나 데이터베이스 없이 코드를 직접 스캔하여 추적합니다.

```bash
# TAG 스캔 예시
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

#### 원칙 2: 영구 불변 ID
SPEC ID는 한 번 부여되면 절대 변경되지 않습니다. 내용은 자유롭게 수정 가능하지만, ID는 영구적입니다.

```yaml
id: AUTH-001  # 영구 불변, 절대 변경 금지
```

#### 원칙 3: HISTORY 기반 버전 관리
모든 변경 사항은 HISTORY 섹션에 기록됩니다. 이를 통해 SPEC의 진화 과정을 완전히 추적할 수 있습니다.

```markdown
## HISTORY
### v0.0.1 (2025-10-13)
- **INITIAL**: JWT 인증 시스템 명세 작성
- **AUTHOR**: @Goos

### v0.1.0 (2025-10-14)
- **CHANGED**: 토큰 만료 시간 15분에서 30분으로 변경
- **AUTHOR**: @Alice
- **REVIEW**: @Bob (approved)
- **REASON**: 사용자 경험 개선 요청
```

#### 원칙 4: EARS 구문 준수
모든 요구사항은 EARS (Easy Approach to Requirements Syntax) 방법론을 따릅니다.

- **Ubiquitous**: 시스템은 [기능]을 제공해야 한다
- **Event-driven**: WHEN [조건]이면, 시스템은 [동작]해야 한다
- **State-driven**: WHILE [상태]일 때, 시스템은 [동작]해야 한다
- **Optional**: WHERE [조건]이면, 시스템은 [동작]할 수 있다
- **Constraints**: IF [조건]이면, 시스템은 [제약]해야 한다

---

## 2 SPEC-First TDD 철학

### 2.1. "명세 없으면 코드 없다"

MoAI-ADK의 핵심 철학은 **"명세 없으면 코드 없다"**입니다. 이는 다음을 의미합니다:

#### 모든 코드는 SPEC에서 시작

```
아이디어 → SPEC 작성 → 테스트 작성 → 코드 구현
```

#### 즉흥적인 코드 작성 금지
SPEC 없이 작성된 코드는 다음 문제를 야기합니다:
- 요구사항 불명확
- 테스트 기준 부재
- 추적성 상실
- 문서화 불가능

#### SPEC은 팀의 계약서
SPEC은 개발자, 리뷰어, 사용자 간의 명확한 계약입니다. 구현 전에 합의를 얻고, 구현 후에 검증합니다.

### 2.2. SPEC-First TDD 사이클

```
1. SPEC 작성 (/alfred:1-spec)
   ↓
   [요구사항 명확화, EARS 구문 작성]
   ↓
2. RED: 테스트 작성 (/alfred:2-build)
   ↓
   [SPEC 기반 실패하는 테스트]
   ↓
3. GREEN: 구현
   ↓
   [최소한의 코드로 테스트 통과]
   ↓
4. REFACTOR: 개선
   ↓
   [코드 품질 향상, SPEC 준수 확인]
   ↓
5. 문서 동기화 (/alfred:3-sync)
   ↓
   [Living Document 생성, TAG 검증]
   ↓
다음 SPEC으로 반복
```

### 2.3. SPEC의 역할

#### 설계 도구
- 구현 전에 아키텍처 설계
- 의존성 및 인터페이스 정의
- 복잡도 분석 및 제약 조건 명시

#### 소통 도구
- 개발자 간 명확한 의사소통
- 리뷰어에게 구현 의도 전달
- 사용자에게 기능 설명

#### 검증 도구
- 테스트 케이스의 기준점
- 코드 리뷰의 체크리스트
- 완료 기준 (Definition of Done)

---

## 3 SPEC 워크플로우

### 3.1. SPEC 생성 워크플로우

#### 1단계: SPEC ID 결정

```bash
# 중복 확인 (필수)
rg "@SPEC:AUTH" -n .moai/specs/

# 새 ID 결정
# 형식: <DOMAIN>-<3자리 숫자>
# 예시: AUTH-001, PAYMENT-002, REFACTOR-001
```

**ID 명명 규칙**:
- 도메인은 대문자
- 숫자는 001부터 시작
- 하이픈으로 구분
- 복합 도메인 가능 (예: `UPDATE-REFACTOR-001`)

**디렉토리 명명 규칙** (필수):

```
.moai/specs/SPEC-{ID}/
```

✅ 올바른 예시:
- `.moai/specs/SPEC-AUTH-001/`
- `.moai/specs/SPEC-REFACTOR-001/`
- `.moai/specs/SPEC-UPDATE-REFACTOR-001/`

❌ 잘못된 예시:
- `.moai/specs/AUTH-001/` (SPEC- 접두사 누락)
- `.moai/specs/SPEC-001-auth/` (순서 잘못됨)
- `.moai/specs/SPEC-AUTH-001-jwt/` (추가 설명 불필요)

#### 2단계: SPEC 파일 생성

```bash
# 디렉토리 생성
mkdir -p .moai/specs/SPEC-AUTH-001/

# SPEC 파일 생성
touch .moai/specs/SPEC-AUTH-001/spec.md
```

#### 3단계: YAML Front Matter 작성

```yaml
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-13
updated: 2025-10-13
author: @Goos
priority: high
category: feature
labels:
  - authentication
  - security
depends_on:
  - USER-001
scope:
  packages:
    - src/core/auth
  files:
    - auth-service.py
    - jwt-manager.py
---
```

#### 4단계: SPEC 본문 작성

```markdown
# @SPEC:AUTH-001: JWT 인증 시스템

## HISTORY
### v0.0.1 (2025-10-13)
- **INITIAL**: JWT 기반 인증 시스템 명세 작성
- **AUTHOR**: @Goos

## 개요
JWT 토큰 기반 사용자 인증 시스템을 구현한다.

## Environment (환경 및 전제조건)
- Python 3.14+
- PyJWT 2.8+
- Redis (토큰 저장)

## Requirements (요구사항)
### Ubiquitous Requirements
- 시스템은 JWT 토큰 발급 기능을 제공해야 한다

### Event-driven Requirements
- WHEN 유효한 자격증명으로 로그인하면, 시스템은 JWT 토큰을 발급해야 한다

...
```

### 3.2. SPEC 브랜치 전략

#### Personal 모드

```bash
# main에서 직접 작업
git checkout main
```

#### Team 모드 (권장)

```bash
# develop에서 feature 브랜치 생성
git checkout develop
git checkout -b feature/SPEC-AUTH-001

# Draft PR 생성
gh pr create --title "SPEC-AUTH-001: JWT 인증 시스템" \
  --body "JWT 기반 인증 구현" \
  --base develop \
  --draft
```

### 3.3. SPEC 리뷰 프로세스

#### 리뷰 체크리스트
- [ ] SPEC ID 중복 확인
- [ ] YAML Front Matter 필수 필드 7개 포함
- [ ] HISTORY 섹션 존재
- [ ] EARS 구문 준수
- [ ] 요구사항 명확성
- [ ] 기술적 타당성
- [ ] 의존성 확인
- [ ] 성공 기준 명시

#### 리뷰 승인 후

```bash
# SPEC 승인 상태 변경
status: active

# HISTORY에 리뷰 기록
### v0.0.2 (2025-10-14)
- **CHANGED**: 토큰 만료 시간 명시 추가
- **AUTHOR**: @Goos
- **REVIEW**: @Alice (approved)
```

---

## 4 @TAG 시스템 통합

### 4.1. TAG 체계 개요

MoAI-ADK는 4개의 TAG로 전체 개발 라이프사이클을 추적합니다:

```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

### 4.2. TAG 역할

#### @SPEC:ID (사전 준비)
- **위치**: `.moai/specs/SPEC-{ID}/spec.md`
- **역할**: 요구사항 명세
- **작성 시점**: `/alfred:1-spec` 실행 시
- **필수 여부**: ✅ 필수

```markdown
# @SPEC:AUTH-001: JWT 인증 시스템
```

#### @TEST:ID (RED 단계)
- **위치**: `tests/`
- **역할**: 테스트 케이스
- **작성 시점**: `/alfred:2-build` RED 단계
- **필수 여부**: ✅ 필수

```python
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

def test_jwt_token_generation():
    """JWT 토큰 발급 테스트"""
    pass
```

#### @CODE:ID (GREEN + REFACTOR 단계)
- **위치**: `src/`
- **역할**: 구현 코드
- **작성 시점**: `/alfred:2-build` GREEN/REFACTOR 단계
- **필수 여부**: ✅ 필수

```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/test_service.py

class JWTService:
    """JWT 토큰 관리 서비스"""
    pass
```

#### @DOC:ID (문서화)
- **위치**: `docs/`
- **역할**: Living Document
- **작성 시점**: `/alfred:3-sync` 실행 시
- **필수 여부**: ⚠️ 조건부 (API/CLI 프로젝트는 필수)

```markdown
# @DOC:AUTH-001 | SPEC: SPEC-AUTH-001.md

## JWT 인증 API
...
```

### 4.3. @CODE 서브 카테고리

구현 세부사항은 주석으로 표기합니다:

```python
# @CODE:AUTH-001:API
# REST API 엔드포인트

# @CODE:AUTH-001:DOMAIN
# 비즈니스 로직

# @CODE:AUTH-001:DATA
# 데이터 모델

# @CODE:AUTH-001:UI
# UI 컴포넌트

# @CODE:AUTH-001:INFRA
# 인프라 코드
```

### 4.4. TAG 검증

#### 전체 TAG 스캔

```bash
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

#### 특정 ID 검증

```bash
rg "@SPEC:AUTH-001" -n .moai/specs/
rg "@TEST:AUTH-001" -n tests/
rg "@CODE:AUTH-001" -n src/
rg "@DOC:AUTH-001" -n docs/
```

#### 고아 TAG 탐지

```bash
# CODE는 있는데 SPEC이 없으면 고아
rg '@CODE:AUTH-001' -n src/
rg '@SPEC:AUTH-001' -n .moai/specs/
```

---

## 5 EARS 요구사항 작성법

### 5.1. EARS 방법론 소개

**EARS (Easy Approach to Requirements Syntax)**는 체계적이고 명확한 요구사항 작성을 위한 방법론입니다. 5가지 구문 패턴으로 모든 요구사항을 표현할 수 있습니다.

### 5.2. EARS 5가지 구문

#### 1. Ubiquitous (기본 요구사항)

**형식**: 시스템은 [기능]을 제공해야 한다

**언제 사용**: 항상 참인 필수 기능

**예시**:

```markdown
### Ubiquitous Requirements
- 시스템은 사용자 인증 기능을 제공해야 한다
- 시스템은 JWT 토큰 발급 기능을 제공해야 한다
- 시스템은 토큰 검증 기능을 제공해야 한다
- 시스템은 사용자 권한 확인 기능을 제공해야 한다
```

#### 2. Event-driven (이벤트 기반)

**형식**: WHEN [조건]이면, 시스템은 [동작]해야 한다

**언제 사용**: 특정 이벤트 발생 시 동작

**예시**:

```markdown
### Event-driven Requirements
- WHEN 사용자가 유효한 자격증명으로 로그인하면, 시스템은 JWT 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 시스템은 401 에러를 반환해야 한다
- WHEN 잘못된 토큰이 제공되면, 시스템은 접근을 거부해야 한다
- WHEN 로그아웃 요청이 오면, 시스템은 토큰을 무효화해야 한다
```

#### 3. State-driven (상태 기반)

**형식**: WHILE [상태]일 때, 시스템은 [동작]해야 한다

**언제 사용**: 특정 상태가 유지되는 동안의 동작

**예시**:

```markdown
### State-driven Requirements
- WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 리소스 접근을 허용해야 한다
- WHILE 토큰이 유효한 상태일 때, 시스템은 API 요청을 처리해야 한다
- WHILE 세션이 활성화된 상태일 때, 시스템은 사용자 활동을 추적해야 한다
```

#### 4. Optional (선택적 기능)

**형식**: WHERE [조건]이면, 시스템은 [동작]할 수 있다

**언제 사용**: 필수가 아닌 선택적 기능

**예시**:

```markdown
### Optional Features
- WHERE 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급할 수 있다
- WHERE 2FA가 활성화되면, 시스템은 추가 인증을 요청할 수 있다
- WHERE 소셜 로그인이 설정되면, 시스템은 OAuth 인증을 지원할 수 있다
```

#### 5. Constraints (제약사항)

**형식**: IF [조건]이면, 시스템은 [제약]해야 한다

**언제 사용**: 시스템의 제한 사항 및 경계 조건

**예시**:

```markdown
### Constraints
- IF 잘못된 토큰이 제공되면, 시스템은 접근을 거부해야 한다
- 액세스 토큰 만료시간은 15분을 초과하지 않아야 한다
- 리프레시 토큰 만료시간은 7일을 초과하지 않아야 한다
- 동시 세션 수는 5개를 초과할 수 없다
- 패스워드 길이는 8자 이상이어야 한다
```

### 5.3. EARS 작성 모범 사례

#### 명확성 우선

```markdown
❌ 나쁜 예:
- 시스템은 안전해야 한다

✅ 좋은 예:
- 시스템은 모든 API 요청에 HTTPS를 강제해야 한다
- 시스템은 패스워드를 bcrypt로 해싱해야 한다
```

#### 검증 가능성

```markdown
❌ 나쁜 예:
- 시스템은 빠르게 응답해야 한다

✅ 좋은 예:
- 시스템은 API 요청에 200ms 이내에 응답해야 한다
```

#### 모호함 제거

```markdown
❌ 나쁜 예:
- 시스템은 적절한 에러를 반환해야 한다

✅ 좋은 예:
- WHEN 인증 실패 시, 시스템은 401 상태 코드와 "Invalid credentials" 메시지를 반환해야 한다
```

### 5.4. EARS 구문 선택 가이드

| 상황 | 사용할 구문 | 예시 키워드 |
|------|------------|------------|
| 항상 참인 기능 | Ubiquitous | "제공해야 한다", "지원해야 한다" |
| 이벤트 발생 시 | Event-driven | "WHEN...하면", "요청이 오면" |
| 상태 유지 중 | State-driven | "WHILE...일 때", "동안" |
| 조건부 기능 | Optional | "WHERE...이면", "할 수 있다" |
| 제한/경계 | Constraints | "IF...이면", "초과하지 않아야" |

---

## 6 SPEC 문서 구조

### 6.1. 표준 SPEC 템플릿

```markdown
---
# YAML Front Matter (필수)
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-13
updated: 2025-10-13
author: @Goos
priority: high
category: feature
labels:
  - authentication
  - security
depends_on:
  - USER-001
scope:
  packages:
    - src/core/auth
  files:
    - auth-service.py
---

# @SPEC:AUTH-001: JWT 인증 시스템

## HISTORY
### v0.0.1 (2025-10-13)
- **INITIAL**: JWT 기반 인증 시스템 명세 작성
- **AUTHOR**: @Goos

---

## 개요
[1-2 문단으로 SPEC의 목적과 범위를 설명]

---

## Environment (환경 및 전제조건)

### 기술 스택
- [사용할 기술 나열]

### 기존 시스템
- [현재 시스템 상태 설명]

---

## Requirements (요구사항)

### Ubiquitous Requirements (기본 요구사항)
- [필수 기능 나열]

### Event-driven Requirements (이벤트 기반)
- WHEN [조건]이면, 시스템은 [동작]해야 한다

### State-driven Requirements (상태 기반)
- WHILE [상태]일 때, 시스템은 [동작]해야 한다

### Optional Features (선택적 기능)
- WHERE [조건]이면, 시스템은 [동작]할 수 있다

### Constraints (제약사항)
- IF [조건]이면, 시스템은 [제약]해야 한다

---

## Specifications (상세 명세)

### 1. [주요 컴포넌트]
[코드 예시 및 설명]

### 2. [인터페이스 정의]
[API/함수 시그니처]

---

**Traceability (추적성)**

- **SPEC ID**: @SPEC:AUTH-001
- **Depends on**: USER-001
- **TAG 체인**: @SPEC:AUTH-001 → @TEST:AUTH-001 → @CODE:AUTH-001
```

### 6.2. 섹션별 상세 가이드

#### 개요 섹션
- **목적**: 1-2 문단으로 SPEC의 핵심을 설명
- **포함 내용**: 무엇을, 왜, 어떻게
- **길이**: 100-200 단어

```markdown
## 개요
JWT 토큰 기반 사용자 인증 시스템을 구현한다. 기존 세션 기반 인증을 대체하여
무상태(stateless) 인증을 제공하고, 마이크로서비스 아키텍처에 적합한 인증
메커니즘을 구축한다.
```

#### Environment 섹션
- **목적**: 구현 환경 및 전제 조건 명시
- **포함 내용**: 기술 스택, 의존성, 기존 시스템
- **중요성**: 구현 가능성 판단 기준

```markdown
## Environment (환경 및 전제조건)

### 기술 스택
- **Language**: Python 3.14+
- **Framework**: FastAPI 0.100+
- **Library**: PyJWT 2.8+
- **Storage**: Redis 7.0+

### 기존 시스템
- 세션 기반 인증 (Flask-Session)
- 단일 서버 아키텍처
```

#### Requirements 섹션
- **목적**: EARS 구문으로 요구사항 명세
- **포함 내용**: 5가지 EARS 구문
- **검증 가능성**: 각 요구사항은 테스트 가능해야 함

#### Specifications 섹션
- **목적**: 구현 가이드 제공
- **포함 내용**: 클래스 구조, 함수 시그니처, 알고리즘
- **코드 예시**: 가상 코드 또는 의사 코드

### 6.3. 선택적 섹션

#### Assumptions (가정)

```markdown
## Assumptions (가정)
1. Redis는 항상 사용 가능하다고 가정
2. 토큰 갱신 요청은 드물다고 가정
3. 사용자는 동시에 최대 3개 디바이스에서 로그인
```

#### Success Criteria (성공 기준)

```markdown
## Success Criteria (성공 기준)

### 기능 완성도
- [ ] JWT 토큰 발급 구현
- [ ] 토큰 검증 구현
- [ ] 토큰 갱신 구현

### 품질 기준
- [ ] 테스트 커버리지 85% 이상
- [ ] API 응답 시간 < 100ms

### 문서화
- [ ] API 문서 자동 생성
- [ ] 사용 예시 3개 이상
```

#### Risk Analysis (리스크 분석)

```markdown
## Risk Analysis (리스크 분석)

### 높은 리스크
1. **Redis 장애 시 인증 불가**
   - 완화: Redis Cluster 구성, 백업 저장소 준비

### 중간 리스크
1. **토큰 탈취 위험**
   - 완화: HTTPS 강제, 짧은 만료 시간
```

---

## 7 실전 SPEC 작성 예시

### 7.1. 간단한 SPEC 예시: CLI 명령어

```markdown
---
id: CLI-INIT-001
version: 0.0.1
status: draft
created: 2025-10-13
updated: 2025-10-13
author: @Goos
priority: high
category: feature
labels:
  - cli
  - initialization
scope:
  packages:
    - src/cli/commands
  files:
    - init.py
---

# @SPEC:CLI-INIT-001: moai init 명령어

## HISTORY
### v0.0.1 (2025-10-13)
- **INITIAL**: moai init 명령어 명세 작성
- **AUTHOR**: @Goos

---

## 개요
`moai init .` 명령어를 구현하여 새로운 MoAI-ADK 프로젝트를 초기화한다.
`.moai/` 디렉토리 구조를 생성하고, 기본 설정 파일을 배치한다.

---

## Environment (환경 및 전제조건)

### 기술 스택
- **CLI Framework**: Click 8.1+
- **Terminal UI**: Rich 13.0+

---

## Requirements (요구사항)

### Ubiquitous Requirements
- 시스템은 `moai init <path>` 명령어를 제공해야 한다
- 시스템은 `.moai/` 디렉토리를 생성해야 한다

### Event-driven Requirements
- WHEN `moai init .` 명령이 실행되면, 시스템은 현재 디렉토리를 초기화해야 한다
- WHEN `.moai/`가 이미 존재하면, 시스템은 에러를 반환해야 한다

### Constraints
- 초기화 시간은 3초 이내여야 한다
- `.moai/` 디렉토리는 읽기/쓰기 권한이어야 한다

---

## Specifications (상세 명세)

### 1. 명령어 인터페이스

```python
@cli.command()
@click.argument('path', type=click.Path(), default='.')
def init(path: str):
    """Initialize a new MoAI-ADK project"""
    pass
```

### 2. 생성할 디렉토리 구조

```
.moai/
├── config.json
├── specs/
├── memory/
└── reports/
```

---

**Traceability (추적성)**

- **SPEC ID**: @SPEC:CLI-INIT-001
- **TAG 체인**: @SPEC:CLI-INIT-001 → @TEST:CLI-INIT-001 → @CODE:CLI-INIT-001

```

### 7.2. 복잡한 SPEC 예시: 인증 시스템

```markdown
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-13
updated: 2025-10-13
author: @Goos
priority: critical
category: security
labels:
  - authentication
  - jwt
  - security
depends_on:
  - USER-001
  - SESSION-001
blocks:
  - AUTH-002
related_specs:
  - TOKEN-001
scope:
  packages:
    - src/core/auth
    - src/api/auth
  files:
    - auth-service.py
    - jwt-manager.py
    - token-store.py
---

# @SPEC:AUTH-001: JWT 기반 인증 시스템

## HISTORY
### v0.0.1 (2025-10-13)
- **INITIAL**: JWT 인증 시스템 명세 작성
- **AUTHOR**: @Goos
- **SCOPE**:
  - JWT 토큰 발급/검증
  - 토큰 갱신 메커니즘
  - Redis 기반 토큰 저장소
- **BACKGROUND**:
  - 기존 세션 기반 인증을 JWT로 전환
  - 마이크로서비스 아키텍처 대응

---

## 개요
JWT (JSON Web Token) 기반 사용자 인증 시스템을 구현한다.
기존 세션 기반 인증을 대체하여 무상태(stateless) 인증을 제공하고,
마이크로서비스 아키텍처에 적합한 인증 메커니즘을 구축한다.

---

## Environment (환경 및 전제조건)

### 기술 스택
- **Language**: Python 3.14+
- **Framework**: FastAPI 0.100+
- **JWT Library**: PyJWT 2.8+
- **Storage**: Redis 7.0+
- **Validation**: Pydantic 2.0+

### 기존 시스템
- Flask-Session 기반 세션 관리
- 단일 서버 아키텍처
- 메모리 기반 세션 저장

### 전제 조건
- Redis 서버 가용성
- HTTPS 통신 환경
- USER-001 (사용자 관리) 구현 완료

---

## Assumptions (가정)

1. **Redis 가용성**: Redis는 99.9% 가용성을 보장한다
2. **토큰 갱신 빈도**: 평균 사용자는 하루 1회 토큰 갱신
3. **동시 세션**: 사용자당 최대 3개 디바이스 동시 로그인
4. **보안**: HTTPS 환경에서만 운영
5. **시간 동기화**: 모든 서버는 NTP로 시간 동기화

---

## Requirements (요구사항)

### Ubiquitous Requirements (기본 요구사항)
- 시스템은 JWT 토큰 발급 기능을 제공해야 한다
- 시스템은 JWT 토큰 검증 기능을 제공해야 한다
- 시스템은 토큰 갱신(refresh) 기능을 제공해야 한다
- 시스템은 토큰 무효화 기능을 제공해야 한다
- 시스템은 사용자 권한 확인 기능을 제공해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 유효한 자격증명으로 로그인하면, 시스템은 액세스 토큰과 리프레시 토큰을 발급해야 한다
- WHEN 액세스 토큰이 만료되면, 시스템은 401 에러를 반환해야 한다
- WHEN 리프레시 토큰으로 갱신 요청하면, 시스템은 새로운 액세스 토큰을 발급해야 한다
- WHEN 잘못된 토큰이 제공되면, 시스템은 접근을 거부해야 한다
- WHEN 로그아웃 요청이 오면, 시스템은 토큰을 블랙리스트에 추가해야 한다
- WHEN 3회 연속 인증 실패 시, 시스템은 계정을 5분간 잠가야 한다

### State-driven Requirements (상태 기반)
- WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 리소스 접근을 허용해야 한다
- WHILE 토큰이 유효한 상태일 때, 시스템은 API 요청을 처리해야 한다
- WHILE 세션이 활성화된 상태일 때, 시스템은 사용자 활동을 추적해야 한다
- WHILE 계정이 잠긴 상태일 때, 시스템은 로그인 시도를 거부해야 한다

### Optional Features (선택적 기능)
- WHERE 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급할 수 있다
- WHERE 2FA가 활성화되면, 시스템은 추가 인증을 요청할 수 있다
- WHERE 소셜 로그인이 설정되면, 시스템은 OAuth 인증을 지원할 수 있다
- WHERE 디바이스 추적이 활성화되면, 시스템은 로그인 디바이스 정보를 저장할 수 있다

### Constraints (제약사항)
- IF 잘못된 토큰이 제공되면, 시스템은 접근을 거부해야 한다
- 액세스 토큰 만료시간은 15분을 초과하지 않아야 한다
- 리프레시 토큰 만료시간은 7일을 초과하지 않아야 한다
- 동시 세션 수는 5개를 초과할 수 없다
- 패스워드 길이는 8자 이상 128자 이하여야 한다
- JWT 페이로드 크기는 1KB를 초과하지 않아야 한다
- 인증 API 응답 시간은 100ms 이내여야 한다

---

## Specifications (상세 명세)

### 1. JWTService 클래스

```python
# @CODE:AUTH-001:DOMAIN | SPEC: SPEC-AUTH-001.md

from datetime import datetime, timedelta
import jwt
from redis import Redis
from pydantic import BaseModel

class TokenPair(BaseModel):
    access_token: str
    refresh_token: str
    expires_in: int

class JWTService:
    def __init__(self, redis_client: Redis, secret_key: str):
        self.redis = redis_client
        self.secret_key = secret_key
        self.algorithm = "HS256"
        self.access_token_expire = timedelta(minutes=15)
        self.refresh_token_expire = timedelta(days=7)

    def generate_token_pair(self, user_id: str) -> TokenPair:
        """액세스 토큰 + 리프레시 토큰 생성"""
        pass

    def verify_access_token(self, token: str) -> dict:
        """액세스 토큰 검증"""
        pass

    def refresh_access_token(self, refresh_token: str) -> str:
        """리프레시 토큰으로 새 액세스 토큰 발급"""
        pass

    def revoke_token(self, token: str) -> None:
        """토큰 무효화 (블랙리스트 추가)"""
        pass

    def is_token_blacklisted(self, token: str) -> bool:
        """토큰 블랙리스트 확인"""
        pass
```

### 2. 토큰 구조

```python
# 액세스 토큰 페이로드
{
    "user_id": "user123",
    "email": "user@example.com",
    "role": "admin",
    "exp": 1697123456,  # 만료 시간
    "iat": 1697122556,  # 발급 시간
    "jti": "token-id"   # 토큰 고유 ID
}

# 리프레시 토큰 페이로드
{
    "user_id": "user123",
    "exp": 1697728356,
    "iat": 1697122556,
    "jti": "refresh-token-id"
}
```

### 3. Redis 키 전략

```python
# 블랙리스트 키 (TTL = 토큰 만료 시간)
blacklist:{jti} = "1"

# 리프레시 토큰 저장 (TTL = 7일)
refresh:{user_id}:{jti} = "refresh_token_value"

# 실패 시도 카운터 (TTL = 5분)
login_attempts:{user_id} = 3
```

### 4. API 엔드포인트

```python
# @CODE:AUTH-001:API | SPEC: SPEC-AUTH-001.md

from fastapi import APIRouter, Depends, HTTPException
from fastapi.security import HTTPBearer

router = APIRouter(prefix="/auth", tags=["authentication"])
security = HTTPBearer()

@router.post("/login")
async def login(credentials: LoginRequest) -> TokenPair:
    """로그인 및 토큰 발급"""
    pass

@router.post("/refresh")
async def refresh(refresh_token: str) -> dict:
    """토큰 갱신"""
    pass

@router.post("/logout")
async def logout(token: str = Depends(security)) -> dict:
    """로그아웃 및 토큰 무효화"""
    pass

@router.get("/verify")
async def verify(token: str = Depends(security)) -> dict:
    """토큰 검증"""
    pass
```

### 5. 에러 처리

```python
class AuthenticationError(Exception):
    """인증 실패 예외"""
    pass

class TokenExpiredError(AuthenticationError):
    """토큰 만료 예외"""
    pass

class InvalidTokenError(AuthenticationError):
    """잘못된 토큰 예외"""
    pass

class AccountLockedError(AuthenticationError):
    """계정 잠김 예외"""
    pass
```

---

**Traceability (추적성)**

- **SPEC ID**: @SPEC:AUTH-001
- **Depends on**: USER-001, SESSION-001
- **Blocks**: AUTH-002
- **Related**: TOKEN-001
- **TAG 체인**: @SPEC:AUTH-001 → @TEST:AUTH-001 → @CODE:AUTH-001 → @DOC:AUTH-001

---

**Success Criteria (성공 기준)**

### 기능 완성도
- [ ] JWT 토큰 발급 구현
- [ ] 토큰 검증 구현
- [ ] 토큰 갱신 구현
- [ ] 토큰 무효화 구현
- [ ] 블랙리스트 관리 구현

### 품질 기준
- [ ] 테스트 커버리지 85% 이상
- [ ] API 응답 시간 < 100ms
- [ ] 코드 복잡도 ≤ 10
- [ ] 보안 스캔 통과 (Bandit)

### 성능 기준
- [ ] 초당 1000 토큰 발급 가능
- [ ] Redis 연결 풀 관리
- [ ] 메모리 사용량 < 100MB

### 문서화
- [ ] API 문서 자동 생성 (FastAPI Swagger)
- [ ] 사용 예시 5개 이상
- [ ] 보안 가이드 작성

---

**Risk Analysis (리스크 분석)**

### 높은 리스크
1. **Redis 장애 시 인증 불가**
   - **완화**: Redis Cluster 구성, Sentinel 모니터링
   - **대안**: 로컬 캐시 fallback

2. **토큰 탈취 위험**
   - **완화**: HTTPS 강제, 짧은 만료 시간, HttpOnly 쿠키
   - **대안**: 디바이스 핑거프린팅 추가

### 중간 리스크
1. **시간 동기화 문제**
   - **완화**: NTP 설정, 서버 시간 모니터링
   - **대안**: 시간 오차 허용 범위 설정

2. **블랙리스트 크기 증가**
   - **완화**: TTL 기반 자동 삭제, 주기적 정리
   - **대안**: Bloom Filter 적용

---

**Implementation Notes (구현 참고사항)**

### 단계별 구현
1. **Phase 1**: JWTService 기본 구현 (발급/검증)
2. **Phase 2**: Redis 연동 및 블랙리스트
3. **Phase 3**: API 엔드포인트 구현
4. **Phase 4**: 에러 처리 및 로깅
5. **Phase 5**: 테스트 작성 (85% 커버리지)
6. **Phase 6**: 성능 최적화 및 보안 강화

### 테스트 전략
- **단위 테스트**: 각 메서드 독립 테스트
- **통합 테스트**: Redis 연동 테스트
- **E2E 테스트**: 전체 인증 플로우
- **부하 테스트**: 초당 1000 요청 처리 확인
- **보안 테스트**: 토큰 탈취 시나리오

### 코드 리뷰 포인트
- 토큰 검증 로직 정확성
- Redis 연결 에러 처리
- 시크릿 키 보안 관리
- 로깅 민감 정보 제외
- 타입 힌트 완전성

```

---

## 8 SPEC 품질 기준

### 8.1. 필수 품질 체크리스트

#### YAML Front Matter
- [ ] 필수 필드 7개 모두 포함 (id, version, status, created, updated, author, priority)
- [ ] version은 0.0.1로 시작
- [ ] status는 draft로 시작
- [ ] author는 @{GitHubID} 형식
- [ ] priority는 low|medium|high|critical 중 하나

#### HISTORY 섹션
- [ ] HISTORY 섹션 존재
- [ ] v0.0.1 INITIAL 항목 포함
- [ ] AUTHOR 명시
- [ ] 변경 유형 태그 사용 (INITIAL, ADDED, CHANGED, etc.)

#### Requirements 섹션
- [ ] EARS 5가지 구문 중 최소 3개 사용
- [ ] Ubiquitous Requirements 필수 포함
- [ ] 각 요구사항은 검증 가능
- [ ] 모호한 표현 없음

#### Traceability 섹션
- [ ] @SPEC:ID 명시
- [ ] TAG 체인 명시
- [ ] 의존성 명시 (있는 경우)

### 8.2. 권장 품질 기준

#### 명확성
- 기술 용어 정의 명확
- 약어 첫 사용 시 전체 이름 명시
- 예시 코드 포함

#### 완전성
- Environment 섹션 상세
- Specifications 섹션에 구현 가이드
- Success Criteria 명시

#### 검증 가능성
- 모든 요구사항은 테스트 가능
- 정량적 기준 제시 (응답 시간, 커버리지 등)
- 경계 조건 명시

### 8.3. SPEC 길이 가이드

| SPEC 복잡도 | 추천 길이 | 예시 |
|------------|----------|------|
| 간단 | 100-300 줄 | CLI 명령어, 유틸리티 함수 |
| 보통 | 300-600 줄 | API 엔드포인트, 서비스 클래스 |
| 복잡 | 600-1000 줄 | 인증 시스템, 결제 모듈 |
| 매우 복잡 | 1000+ 줄 | 분산 시스템, 복잡한 워크플로우 |

**참고**: 1000줄 초과 시 SPEC 분리 고려

---

## 9 SPEC 검증 및 추적

### 9.1. SPEC 검증 명령어

#### 필수 필드 검증
```bash
# 모든 SPEC에 필수 필드 있는지 확인
rg "^(id|version|status|created|updated|author|priority):" .moai/specs/SPEC-*/spec.md

# priority 필드 누락 확인
rg -L "^priority:" .moai/specs/SPEC-*/spec.md
```

#### 형식 검증

```bash
# author 필드 형식 확인 (@Username)
rg "^author: @[A-Z]" .moai/specs/SPEC-*/spec.md

# version 필드 형식 확인 (0.x.y)
rg "^version: 0\.\d+\.\d+" .moai/specs/SPEC-*/spec.md

# HISTORY 섹션 존재 확인
rg "## HISTORY" .moai/specs/SPEC-*/spec.md
```

#### 중복 ID 확인

```bash
# 특정 ID 중복 확인
rg "@SPEC:AUTH-001" -n .moai/specs/

# 모든 SPEC ID 목록
rg "^id: " .moai/specs/SPEC-*/spec.md | sort
```

### 9.2. TAG 체인 검증

#### 전체 TAG 스캔

```bash
# 모든 TAG 출력
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

#### SPEC별 TAG 추적

```bash
# AUTH-001 TAG 전체 추적
echo "=== @SPEC:AUTH-001 ==="
rg '@SPEC:AUTH-001' -n .moai/specs/

echo "=== @TEST:AUTH-001 ==="
rg '@TEST:AUTH-001' -n tests/

echo "=== @CODE:AUTH-001 ==="
rg '@CODE:AUTH-001' -n src/

echo "=== @DOC:AUTH-001 ==="
rg '@DOC:AUTH-001' -n docs/
```

#### 고아 TAG 탐지

```bash
# CODE는 있는데 SPEC이 없는 경우
for code_tag in $(rg '@CODE:' -no-filename src/ | sort -u); do
  spec_tag=$(echo $code_tag | sed 's/@CODE:/@SPEC:/')
  if ! rg -q "$spec_tag" .moai/specs/; then
    echo "Orphan: $code_tag (no matching SPEC)"
  fi
done
```

### 9.3. SPEC 의존성 그래프

#### depends_on 추출

```bash
# 모든 SPEC 의존성 출력
rg "^depends_on:" -A 5 .moai/specs/SPEC-*/spec.md
```

#### 의존성 시각화 (dot 파일 생성)

```bash
# SPEC 의존성 그래프 생성 (수동)
# 1. 모든 SPEC ID와 depends_on 추출
# 2. Graphviz dot 파일 생성
# 3. dot -Tpng deps.dot -o deps.png
```

---

## 10 SPEC 라이프사이클

### 10.1. SPEC 상태 전이

```
draft → active → completed → deprecated
```

#### draft (초안)
- **의미**: 작성 중, 아직 구현 시작 안 함
- **version**: 0.0.x
- **전환 조건**: 리뷰 승인 후 active로 전환

#### active (활성)
- **의미**: 구현 진행 중
- **version**: 0.0.x ~ 0.y.z
- **전환 조건**: TDD 구현 완료 후 completed로 전환

#### completed (완료)
- **의미**: 구현 완료, 테스트 통과, 문서화 완료
- **version**: 0.1.0+
- **전환 조건**: 더 이상 사용 안 함 시 deprecated로 전환

#### deprecated (폐기 예정)
- **의미**: 향후 제거 예정
- **version**: 유지
- **전환 조건**: 완전 제거 시 파일 삭제

### 10.2. 버전 관리 전략

#### Semantic Versioning
- **v0.0.1**: INITIAL - SPEC 최초 작성 (status: draft)
- **v0.0.x**: Draft 수정/개선
- **v0.1.0**: TDD 구현 완료 (status: completed)
- **v0.1.x**: 버그 수정, 문서 개선
- **v0.x.0**: 기능 추가, 주요 개선
- **v1.0.0**: 정식 안정화 버전 (프로덕션 준비)

#### 버전 증가 규칙
- **Patch (0.0.x)**: SPEC 문서 수정, 오타 수정, 명확화
- **Minor (0.x.0)**: 기능 추가, 요구사항 추가 (하위 호환)
- **Major (x.0.0)**: 하위 호환성 깨지는 변경, 요구사항 제거

### 10.3. SPEC 업데이트 프로세스

#### 1단계: 변경 필요성 확인

```bash
# 관련 SPEC 읽기
cat .moai/specs/SPEC-AUTH-001/spec.md

# HISTORY 확인
rg -A 20 "## HISTORY" .moai/specs/SPEC-AUTH-001/spec.md
```

#### 2단계: SPEC 수정

```markdown
# updated 날짜 변경
updated: 2025-10-14

# version 증가
version: 0.0.2

# HISTORY 추가
### v0.0.2 (2025-10-14)
- **CHANGED**: 토큰 만료 시간 15분 → 30분 변경
- **AUTHOR**: @Goos
- **REVIEW**: @Alice (approved)
- **REASON**: 사용자 경험 개선 요청
- **RELATED**: #123
```

#### 3단계: 영향 분석

```bash
# 관련 코드 확인
rg '@CODE:AUTH-001' -n src/

# 관련 테스트 확인
rg '@TEST:AUTH-001' -n tests/
```

#### 4단계: 코드/테스트 업데이트
- SPEC 변경에 따라 코드 수정
- 테스트 케이스 업데이트
- TDD 사이클 재실행

#### 5단계: 문서 동기화

```bash
# /alfred:3-sync 실행
# Living Document 자동 업데이트
# TAG 체인 무결성 검증
```

---

## 11 자주 묻는 질문 (FAQ)

### Q1. SPEC ID는 어떻게 결정하나요?

**A**: SPEC ID는 `<DOMAIN>-<NUMBER>` 형식입니다.

- **도메인**: 기능 영역 (AUTH, USER, PAYMENT 등)
- **번호**: 001부터 시작하는 3자리 숫자

**중복 확인 필수**:

```bash
rg "@SPEC:AUTH" -n .moai/specs/
```

### Q2. SPEC 작성 시 가장 중요한 것은?

**A**: **명확성과 검증 가능성**입니다.

- 모호한 표현 금지
- 각 요구사항은 테스트 가능해야 함
- EARS 구문 준수

### Q3. SPEC 없이 코드를 작성해도 되나요?

**A**: **절대 안 됩니다**. MoAI-ADK의 핵심 원칙은 **"명세 없으면 코드 없다"**입니다.

SPEC 없는 코드는:
- 추적 불가능
- 테스트 기준 부재
- 문서화 불가능
- 리뷰 불가능

### Q4. SPEC이 너무 길어지면 어떻게 하나요?

**A**: **SPEC을 분리**하세요.

- 1000줄 초과 시 분리 고려
- 도메인별로 분리
- 의존성 관계 명시

예시:

```
AUTH-001: JWT 토큰 발급
AUTH-002: 토큰 검증
AUTH-003: 토큰 갱신
```

### Q5. SPEC 수정 시 버전은 어떻게 증가하나요?

**A**: Semantic Versioning을 따릅니다.

- **Patch (0.0.x)**: SPEC 문서 수정, 오타 수정
- **Minor (0.x.0)**: 기능 추가, 요구사항 추가
- **Major (x.0.0)**: 하위 호환성 깨지는 변경

### Q6. TAG 체인이 끊어졌을 때 어떻게 하나요?

**A**: `/alfred:3-sync` 명령으로 TAG 무결성을 검증하고 복구하세요.

```bash
# TAG 체인 검증
rg '@(SPEC|TEST|CODE|DOC):AUTH-001' -n

# 누락된 TAG 추가
# 예: @TEST:AUTH-001이 없다면 테스트 파일에 추가
```

### Q7. EARS 구문을 모두 사용해야 하나요?

**A**: **최소 3개 이상 사용**을 권장합니다.

- Ubiquitous (필수 기능) - 필수
- Event-driven (이벤트 기반) - 권장
- Constraints (제약사항) - 권장
- State-driven, Optional - 필요 시 사용

### Q8. SPEC 리뷰는 누가 하나요?

**A**: **Team 모드**에서는 PR 리뷰어가 SPEC을 리뷰합니다.

리뷰 체크리스트:
- [ ] SPEC ID 중복 없음
- [ ] 필수 필드 모두 포함
- [ ] EARS 구문 준수
- [ ] 요구사항 명확성
- [ ] 기술적 타당성

### Q9. SPEC 작성에 얼마나 시간이 걸리나요?

**A**: SPEC 복잡도에 따라 다릅니다.

- **간단한 SPEC**: 30분 ~ 1시간
- **보통 SPEC**: 1 ~ 2시간
- **복잡한 SPEC**: 2 ~ 4시간

**팁**: 구현 시간의 20-30%를 SPEC 작성에 할애하세요.

### Q10. SPEC 템플릿은 어디에 있나요?

**A**: 이 문서의 [6. SPEC 문서 구조](#6-spec-문서-구조)를 참조하세요.

또는 기존 SPEC 예시를 복사하여 시작:

```bash
cp .moai/specs/SPEC-CLI-001/spec.md .moai/specs/SPEC-NEW-001/spec.md
```

---

## 부록

### A. SPEC 작성 체크리스트

#### 작성 전
- [ ] SPEC ID 중복 확인
- [ ] 의존성 SPEC 확인
- [ ] 기존 시스템 이해

#### 작성 중
- [ ] YAML Front Matter 작성
- [ ] HISTORY 섹션 작성
- [ ] EARS 구문으로 요구사항 작성
- [ ] 코드 예시 포함
- [ ] TAG 체인 명시

#### 작성 후
- [ ] 필수 필드 검증
- [ ] EARS 구문 검증
- [ ] 요구사항 명확성 확인
- [ ] 리뷰 요청
- [ ] 승인 후 status 변경

### B. SPEC 용어집

| 용어 | 정의 |
|------|------|
| SPEC | Specification의 약자, 요구사항 명세 문서 |
| EARS | Easy Approach to Requirements Syntax |
| TAG | 코드-문서 추적을 위한 마커 (@SPEC, @TEST, @CODE, @DOC) |
| CODE-FIRST | 코드를 직접 스캔하여 TAG를 추적하는 방식 |
| Living Document | 코드 변경에 따라 자동으로 업데이트되는 문서 |
| TDD | Test-Driven Development, 테스트 주도 개발 |
| TRUST | Test, Readable, Unified, Secured, Trackable (5원칙) |

### C. 참고 자료

- **EARS 방법론**: [IEEE Guide for Software Requirements Specifications](https://standards.ieee.org/standard/830-1998.html)
- **Semantic Versioning**: https://semver.org/
- **MoAI-ADK 개발 가이드**: `.moai/memory/development-guide.md`
- **SPEC 메타데이터 표준**: `.moai/memory/spec-metadata.md`

---

**최종 업데이트**: 2025-10-14
**작성자**: @Alfred
**버전**: 1.0.0
