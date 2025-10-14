# MoAI-ADK SPEC 메타데이터 완전 가이드

> **SPEC 메타데이터 표준 (SSOT - Single Source of Truth)**
>
> 모든 SPEC 문서는 이 메타데이터 구조를 따라야 합니다.

---

## 목차

- [1. 메타데이터 시스템 개요](#1-메타데이터-시스템-개요)
- [2. YAML Front Matter 구조](#2-yaml-front-matter-구조)
- [3. 필수 필드 상세 (7개)](#3-필수-필드-상세-7개)
- [4. 선택 필드 상세 (9개)](#4-선택-필드-상세-9개)
- [5. HISTORY 섹션 작성법](#5-history-섹션-작성법)
- [6. 버전 관리 체계](#6-버전-관리-체계)
- [7. 메타데이터 검증](#7-메타데이터-검증)
- [8. 의존성 그래프 관리](#8-의존성-그래프-관리)
- [9. 실전 예시](#9-실전-예시)
- [10. 마이그레이션 가이드](#10-마이그레이션-가이드)
- [11. 자동화 도구](#11-자동화-도구)
- [12. FAQ](#12-faq)

---

## 1. 메타데이터 시스템 개요

### 1.1. 메타데이터란?

**SPEC 메타데이터**는 SPEC 문서의 핵심 정보를 구조화된 형식으로 표현한 데이터입니다. YAML Front Matter 형식으로 작성되며, SPEC 문서의 첫 부분에 위치합니다.

### 1.2. 메타데이터의 목적

#### 자동화 가능성
- 파싱 가능한 구조화된 데이터
- 도구를 통한 자동 검증
- 의존성 그래프 자동 생성

#### 일관성 보장
- 모든 SPEC 동일한 구조
- 필수 정보 누락 방지
- 표준화된 형식

#### 추적성 제공
- SPEC 간 의존성 명시
- 변경 이력 자동 추적
- 영향 분석 가능

### 1.3. 메타데이터 구성 요소

MoAI-ADK SPEC 메타데이터는 **필수 필드 7개**와 **선택 필드 9개**로 구성됩니다.

#### 필수 필드 (7개)
1. `id` - SPEC 고유 ID
2. `version` - Semantic Version
3. `status` - 진행 상태
4. `created` - 생성일
5. `updated` - 최종 수정일
6. `author` - 작성자
7. `priority` - 우선순위

#### 선택 필드 (9개)
1. `category` - 변경 유형
2. `labels` - 분류 태그
3. `depends_on` - 의존 SPEC
4. `blocks` - 차단 SPEC
5. `related_specs` - 관련 SPEC
6. `related_issue` - 관련 GitHub Issue
7. `scope.packages` - 영향받는 패키지
8. `scope.files` - 핵심 파일
9. (추가 확장 가능)

---

## 2. YAML Front Matter 구조

### 2.1. 기본 구조

YAML Front Matter는 SPEC 문서의 맨 앞에 `---`로 감싸진 형태로 작성됩니다.

```yaml
---
# 필수 필드 (7개)
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-13
updated: 2025-10-13
author: @Goos
priority: high

# 선택 필드 - 분류/메타
category: security
labels:
  - authentication
  - jwt

# 선택 필드 - 관계
depends_on:
  - USER-001
blocks:
  - AUTH-002
related_specs:
  - TOKEN-002
related_issue: "https://github.com/modu-ai/moai-adk/issues/123"

# 선택 필드 - 범위
scope:
  packages:
    - src/core/auth
  files:
    - auth-service.py
    - jwt-manager.py
---
```

### 2.2. YAML 작성 규칙

#### 들여쓰기
- **2칸 공백** 사용 (탭 사용 금지)
- 일관된 들여쓰기 유지

```yaml
# 올바른 예시
scope:
  packages:
    - src/core/auth
  files:
    - auth-service.py

# 잘못된 예시 (4칸 들여쓰기)
scope:
    packages:
        - src/core/auth
```

#### 문자열 따옴표
- 공백이나 특수문자 포함 시 따옴표 사용
- URL은 항상 따옴표로 감싸기

```yaml
# 올바른 예시
author: @Goos
related_issue: "https://github.com/modu-ai/moai-adk/issues/123"

# 잘못된 예시
author: "@Goos"  # 불필요한 따옴표
related_issue: https://github.com/... # 따옴표 누락
```

#### 배열 표현
- 하이픈(`-`) 사용
- 각 항목은 새 줄에 작성

```yaml
# 올바른 예시
labels:
  - authentication
  - security
  - jwt

# 잘못된 예시
labels: [authentication, security, jwt]  # 인라인 형식 지양
```

#### 날짜 형식
- **YYYY-MM-DD** 형식 엄수
- 따옴표 없이 작성

```yaml
created: 2025-10-13  # 올바름
created: "2025-10-13"  # 불필요한 따옴표
created: 2025/10/13  # 잘못된 형식
created: Oct 13, 2025  # 잘못된 형식
```

### 2.3. 주석 사용

YAML Front Matter 내에서 주석을 사용할 수 있습니다.

```yaml
---
# 필수 필드 (7개)
id: AUTH-001                    # SPEC 고유 ID
version: 0.0.1                  # Semantic Version
status: draft                   # draft|active|completed|deprecated
created: 2025-10-13            # 생성일 (YYYY-MM-DD)
updated: 2025-10-13            # 최종 수정일
author: @Goos                   # 작성자 (GitHub ID)
priority: high                  # low|medium|high|critical
---
```

---

## 3. 필수 필드 상세 (7개)

### 3.1. `id` - SPEC 고유 ID

#### 정의
SPEC을 고유하게 식별하는 영구 불변 ID입니다.

#### 형식

```
<DOMAIN>-<NUMBER>
```

- **DOMAIN**: 대문자, 하이픈 가능
- **NUMBER**: 001부터 시작하는 3자리 숫자

#### 예시

```yaml
id: AUTH-001          # 인증 도메인 1번
id: PAYMENT-042       # 결제 도메인 42번
id: REFACTOR-001      # 리팩토링 1번
id: UPDATE-REFACTOR-001  # 복합 도메인
id: INSTALLER-SEC-001    # 설치 보안 1번
```

#### 규칙
- **영구 불변**: 한 번 부여된 ID는 절대 변경 금지
- **중복 금지**: 새 ID 생성 전 중복 확인 필수
- **도메인 선택**: 기능 영역에 맞는 도메인 사용
- **순차 번호**: 도메인 내에서 001부터 순차 증가

#### 중복 확인 방법

```bash
# 특정 ID 중복 확인
rg "@SPEC:AUTH-001" -n .moai/specs/

# 특정 도메인의 모든 ID 확인
rg "^id: AUTH" .moai/specs/SPEC-*/spec.md

# 사용 중인 모든 ID 목록
rg "^id: " .moai/specs/SPEC-*/spec.md | sort
```

#### 디렉토리 명명 규칙
SPEC ID가 `AUTH-001`이면, 디렉토리는 반드시 `SPEC-AUTH-001/`이어야 합니다.

```
.moai/specs/SPEC-{ID}/spec.md
```

✅ 올바른 예시:
- `.moai/specs/SPEC-AUTH-001/spec.md`
- `.moai/specs/SPEC-REFACTOR-001/spec.md`
- `.moai/specs/SPEC-UPDATE-REFACTOR-001/spec.md`

❌ 잘못된 예시:
- `.moai/specs/AUTH-001/spec.md` (SPEC- 접두사 누락)
- `.moai/specs/SPEC-001-auth/spec.md` (순서 잘못됨)
- `.moai/specs/SPEC-AUTH-001-jwt/spec.md` (추가 설명 불필요)

---

### 3.2. `version` - 버전

#### 정의
SPEC 문서의 버전을 Semantic Versioning 형식으로 표현합니다.

#### 형식

```
MAJOR.MINOR.PATCH
```

#### 기본값
모든 SPEC은 **v0.0.1**로 시작합니다.

```yaml
version: 0.0.1  # INITIAL 버전
```

#### 버전 체계 요약
- **v0.0.1**: INITIAL - SPEC 최초 작성 (status: draft)
- **v0.0.x**: Draft 수정/개선 (SPEC 문서 수정)
- **v0.1.0**: TDD 구현 완료 (status: completed)
- **v0.1.x**: 버그 수정, 문서 개선
- **v0.x.0**: 기능 추가, 주요 개선
- **v1.0.0**: 정식 안정화 버전 (프로덕션 준비)

#### 버전 증가 규칙
| 변경 유형 | 버전 증가 | 예시 |
|----------|----------|------|
| SPEC 문서 수정, 오타 수정 | Patch | 0.0.1 → 0.0.2 |
| 기능 추가, 요구사항 추가 | Minor | 0.0.2 → 0.1.0 |
| 하위 호환성 깨지는 변경 | Major | 0.9.5 → 1.0.0 |

상세 버전 체계는 [6. 버전 관리 체계](#6-버전-관리-체계) 참조

---

### 3.3. `status` - 진행 상태

#### 정의
SPEC의 현재 진행 상태를 나타냅니다.

#### 가능한 값

```yaml
status: draft        # 초안 작성 중
status: active       # 구현 진행 중
status: completed    # 구현 완료
status: deprecated   # 사용 중지 예정
```

#### 상태 전이 흐름

```
draft → active → completed → deprecated
```

#### 상태별 설명
| 상태 | 의미 | 버전 범위 | 전환 조건 |
|------|------|----------|----------|
| `draft` | 초안 작성 중 | 0.0.x | 리뷰 승인 후 active |
| `active` | 구현 진행 중 | 0.0.x ~ 0.y.z | TDD 완료 후 completed |
| `completed` | 구현 완료 | 0.1.0+ | 폐기 예정 시 deprecated |
| `deprecated` | 사용 중지 예정 | 유지 | 완전 제거 시 파일 삭제 |

#### 예시

```yaml
# 최초 작성
status: draft
version: 0.0.1

# 리뷰 승인 후 구현 시작
status: active
version: 0.0.2

# TDD 구현 완료
status: completed
version: 0.1.0

# 폐기 예정
status: deprecated
version: 0.1.0
```

---

### 3.4. `created` - 생성일

#### 정의
SPEC이 최초로 생성된 날짜입니다.

#### 형식

```yaml
created: YYYY-MM-DD
```

#### 규칙
- **변경 금지**: 최초 생성일은 절대 변경하지 않음
- **날짜 형식**: YYYY-MM-DD 엄수
- **시간 제외**: 날짜만 표시, 시간은 포함하지 않음

#### 예시

```yaml
created: 2025-10-13  # 올바름
created: "2025-10-13"  # 불필요한 따옴표
created: 2025/10/13  # 잘못된 형식
created: Oct 13, 2025  # 잘못된 형식
```

---

### 3.5. `updated` - 최종 수정일

#### 정의
SPEC이 마지막으로 수정된 날짜입니다.

#### 형식

```yaml
updated: YYYY-MM-DD
```

#### 규칙
- **자동 업데이트**: SPEC 내용 수정 시마다 업데이트
- **최초 작성 시**: `created`와 동일한 날짜
- **날짜 형식**: YYYY-MM-DD 엄수

#### 예시

```yaml
# 최초 작성 시
created: 2025-10-13
updated: 2025-10-13

# 이후 수정 시
created: 2025-10-13
updated: 2025-10-14  # 수정 날짜로 업데이트
```

---

### 3.6. `author` - 작성자

#### 정의
SPEC을 최초로 작성한 사람의 GitHub ID입니다.

#### 형식

```yaml
author: @{GitHub ID}
```

#### 규칙
- **단수형**: `authors` 배열이 아닌 단수 `author` 사용
- **@ 접두사**: GitHub ID 앞에 @ 필수
- **대문자 시작**: GitHub ID의 첫 글자는 대문자

#### 예시

```yaml
author: @Goos         # 올바름
author: @Alice        # 올바름
author: goos          # @ 누락 (잘못됨)
author: @goos         # 소문자 (권장하지 않음)
authors: ["@Goos"]    # 배열 형식 (잘못됨)
```

#### 복수 작성자 처리
복수 작성자는 `author` 필드에 최초 작성자만 명시하고, 나머지는 HISTORY 섹션에 기록합니다.

```yaml
author: @Goos
```

```markdown
## HISTORY
### v0.0.1 (2025-10-13)
- **INITIAL**: JWT 인증 시스템 명세 작성
- **AUTHOR**: @Goos

### v0.0.2 (2025-10-14)
- **CHANGED**: 토큰 만료 시간 변경
- **AUTHOR**: @Alice
- **REVIEW**: @Bob (approved)
```

---

### 3.7. `priority` - 우선순위

#### 정의
SPEC의 구현 우선순위를 나타냅니다.

#### 가능한 값

```yaml
priority: critical   # 즉시 처리 필요
priority: high       # 높은 우선순위
priority: medium     # 중간 우선순위
priority: low        # 낮은 우선순위
```

#### 우선순위 결정 기준
| 우선순위 | 설명 | 예시 |
|---------|------|------|
| `critical` | 즉시 처리 필요, 보안/중대 버그 | 보안 취약점, 서비스 중단 버그 |
| `high` | 높은 우선순위, 주요 기능 | 핵심 기능 구현, 중요 리팩토링 |
| `medium` | 중간 우선순위, 개선사항 | 성능 최적화, 코드 품질 개선 |
| `low` | 낮은 우선순위, 최적화/문서 | 문서 개선, 사소한 최적화 |

#### 예시

```yaml
# 보안 취약점 수정
priority: critical

# 핵심 인증 시스템 구현
priority: high

# 로깅 개선
priority: medium

# 코드 주석 추가
priority: low
```

---

## 4. 선택 필드 상세 (9개)

### 4.1. `category` - 변경 유형

#### 정의
SPEC이 어떤 유형의 변경인지 분류합니다.

#### 가능한 값

```yaml
category: feature    # 새 기능 추가
category: bugfix     # 버그 수정
category: refactor   # 리팩토링
category: security   # 보안 개선
category: docs       # 문서화
category: perf       # 성능 최적화
```

#### 사용 시점
- 프로젝트 관리 시 SPEC 분류
- 릴리스 노트 자동 생성
- 변경 유형별 통계 분석

#### 예시

```yaml
# 새 기능 구현
id: AUTH-001
category: feature

# 보안 취약점 수정
id: INSTALLER-SEC-001
category: security

# 코드 리팩토링
id: REFACTOR-001
category: refactor
```

---

### 4.2. `labels` - 분류 태그

#### 정의
SPEC을 검색, 필터링, 그루핑하기 위한 태그입니다.

#### 형식

```yaml
labels:
  - tag1
  - tag2
  - tag3
```

#### 사용 용도
- **검색**: 특정 태그로 SPEC 검색
- **필터링**: 태그별 SPEC 목록 조회
- **그루핑**: 관련 SPEC끼리 묶기

#### 예시

```yaml
# 인증 관련 SPEC
labels:
  - authentication
  - jwt
  - security

# 설치 관련 SPEC
labels:
  - installer
  - template
  - cli

# 리팩토링 SPEC
labels:
  - refactoring
  - code-quality
  - maintainability
```

#### 태그 명명 규칙
- 소문자 사용
- 하이픈으로 단어 구분
- 명확하고 간결한 이름

```yaml
# 올바른 예시
labels:
  - authentication
  - jwt-token
  - security-audit

# 잘못된 예시
labels:
  - Authentication  # 대문자 사용
  - jwt_token       # 언더스코어 사용
  - "security audit"  # 공백 사용
```

---

### 4.3. `depends_on` - 의존 SPEC

#### 정의
이 SPEC이 완료되려면 먼저 완료되어야 하는 SPEC 목록입니다.

#### 형식

```yaml
depends_on:
  - SPEC-ID-1
  - SPEC-ID-2
```

#### 의미
**"이 SPEC은 {depends_on}이 완료되어야 시작할 수 있다"**

#### 예시

```yaml
# AUTH-001은 USER-001에 의존
id: AUTH-001
depends_on:
  - USER-001

# PAYMENT-003은 AUTH-001과 USER-002에 의존
id: PAYMENT-003
depends_on:
  - AUTH-001
  - USER-002
```

#### 활용
- **작업 순서 결정**: 의존성 순서대로 구현
- **병렬 작업 판단**: 의존성 없는 SPEC은 병렬 작업 가능
- **영향 분석**: 의존 SPEC 변경 시 영향 범위 파악

#### 의존성 그래프 생성

```bash
# 모든 SPEC 의존성 출력
rg "^depends_on:" -A 5 .moai/specs/SPEC-*/spec.md
```

---

### 4.4. `blocks` - 차단 SPEC

#### 정의
이 SPEC으로 인해 차단된(시작할 수 없는) SPEC 목록입니다.

#### 형식

```yaml
blocks:
  - SPEC-ID-1
  - SPEC-ID-2
```

#### 의미
**"이 SPEC이 완료되어야 {blocks}를 시작할 수 있다"**

#### 예시

```yaml
# AUTH-001이 완료되어야 AUTH-002 시작 가능
id: AUTH-001
blocks:
  - AUTH-002
  - AUTH-003

# USER-001이 완료되어야 PROFILE-001 시작 가능
id: USER-001
blocks:
  - PROFILE-001
```

#### `depends_on`과의 관계

```yaml
# AUTH-001 관점
id: AUTH-001
blocks:
  - AUTH-002

# AUTH-002 관점
id: AUTH-002
depends_on:
  - AUTH-001
```

---

### 4.5. `related_specs` - 관련 SPEC

#### 정의
직접적 의존성은 없지만 관련 있는 SPEC 목록입니다.

#### 형식

```yaml
related_specs:
  - SPEC-ID-1
  - SPEC-ID-2
```

#### 의미
**"이 SPEC과 관련이 있지만, 의존성은 없는 SPEC"**

#### 예시

```yaml
# AUTH-001과 관련 있는 SPEC들
id: AUTH-001
related_specs:
  - TOKEN-002      # 토큰 관리 (관련 있지만 의존성 없음)
  - SESSION-001    # 세션 관리 (유사한 도메인)
  - AUDIT-003      # 감사 로깅 (인증 시 로깅 필요)
```

#### `depends_on`과의 차이
| 구분 | `depends_on` | `related_specs` |
|------|--------------|-----------------|
| 의존성 | 있음 (필수) | 없음 (참고용) |
| 구현 순서 | 순서 영향 | 순서 무관 |
| 차단 여부 | 차단함 | 차단 안 함 |

---

### 4.6. `related_issue` - 관련 GitHub Issue

#### 정의
이 SPEC과 관련된 GitHub Issue URL입니다.

#### 형식

```yaml
related_issue: "https://github.com/{org}/{repo}/issues/{number}"
```

#### 예시

```yaml
related_issue: "https://github.com/modu-ai/moai-adk/issues/123"
```

#### 규칙
- **전체 URL** 사용 (Issue 번호만 사용하지 않음)
- **따옴표** 필수
- **단수형**: 여러 Issue는 HISTORY에 기록

#### 활용
- Issue에서 SPEC으로 추적
- SPEC에서 Issue로 추적
- 요구사항 출처 명확화

---

### 4.7. `scope.packages` - 영향받는 패키지

#### 정의
이 SPEC이 영향을 주는 패키지/모듈 경로입니다.

#### 형식

```yaml
scope:
  packages:
    - path/to/package1
    - path/to/package2
```

#### 예시

```yaml
# Python 프로젝트
scope:
  packages:
    - src/core/auth
    - src/api/auth
    - src/utils/token

# TypeScript 프로젝트
scope:
  packages:
    - moai-adk-ts/src/core/installer
    - moai-adk-ts/src/core/git

# 모노레포
scope:
  packages:
    - packages/auth
    - packages/user
```

#### 활용
- **영향 범위 파악**: 변경 영향을 받는 패키지 확인
- **코드 리뷰 범위**: 리뷰어에게 검토 범위 안내
- **테스트 범위**: 영향받는 패키지의 테스트 실행

---

### 4.8. `scope.files` - 핵심 파일

#### 정의
이 SPEC과 직접 관련된 주요 파일 목록입니다.

#### 형식

```yaml
scope:
  files:
    - file1.py
    - file2.ts
```

#### 예시

```yaml
scope:
  packages:
    - src/core/auth
  files:
    - auth-service.py
    - jwt-manager.py
    - token-store.py
```

#### 규칙
- **상대 경로**: 패키지 루트 기준
- **핵심 파일만**: 모든 파일을 나열하지 말고 주요 파일만
- **선택적**: 필수 아님, 참고용

---

### 4.9. 추가 확장 필드

SPEC 메타데이터는 필요에 따라 추가 필드를 확장할 수 있습니다.

#### 예시: 예상 완료일

```yaml
estimated_completion: 2025-10-20
```

#### 예시: 할당자

```yaml
assignee: @Alice
```

#### 예시: 리뷰어

```yaml
reviewers:
  - @Bob
  - @Charlie
```

#### 주의사항
- 표준 필드 16개 우선 사용
- 확장 필드는 프로젝트 전체에서 일관성 유지
- 문서화 필수

---

## 5. HISTORY 섹션 작성법

### 5.1. HISTORY 섹션이란?

**HISTORY 섹션**은 SPEC 문서의 모든 변경 이력을 기록하는 필수 섹션입니다. Git 커밋 로그와 유사하지만, SPEC 문서 내에 직접 기록됩니다.

### 5.2. HISTORY 기본 구조

```markdown
## HISTORY

### v0.0.1 (2025-10-13)
- **INITIAL**: JWT 기반 인증 시스템 명세 작성
- **AUTHOR**: @Goos

### v0.0.2 (2025-10-14)
- **CHANGED**: 토큰 만료 시간 15분에서 30분으로 변경
- **AUTHOR**: @Alice
- **REVIEW**: @Bob (approved)
- **REASON**: 사용자 경험 개선 요청
- **RELATED**: #123
```

### 5.3. 변경 유형 태그

모든 HISTORY 항목은 변경 유형 태그로 시작합니다.

#### 변경 유형 태그 목록
| 태그 | 의미 | 버전 증가 | 예시 |
|------|------|----------|------|
| `INITIAL` | 최초 작성 | v0.0.1 | SPEC 최초 생성 |
| `ADDED` | 새 기능/요구사항 추가 | Minor | 새 API 추가 |
| `CHANGED` | 기존 내용 수정 | Patch | 파라미터 변경 |
| `FIXED` | 버그/오류 수정 | Patch | 오타 수정 |
| `REMOVED` | 기능/요구사항 제거 | Major | API 제거 |
| `BREAKING` | 하위 호환성 깨지는 변경 | Major | 인터페이스 변경 |
| `DEPRECATED` | 향후 제거 예정 표시 | Minor | 구 API 폐기 예고 |

#### 변경 유형별 예시

##### INITIAL (최초 작성)

```markdown
### v0.0.1 (2025-10-13)
- **INITIAL**: JWT 인증 시스템 명세 작성
- **AUTHOR**: @Goos
```

##### ADDED (추가)

```markdown
### v0.2.0 (2025-10-15)
- **ADDED**: 2FA 인증 요구사항 추가
- **AUTHOR**: @Alice
- **REASON**: 보안 강화 요청
```

##### CHANGED (수정)

```markdown
### v0.0.2 (2025-10-14)
- **CHANGED**: 토큰 만료 시간 15분 → 30분 변경
- **AUTHOR**: @Bob
- **REVIEW**: @Alice (approved)
```

##### FIXED (수정)

```markdown
### v0.0.3 (2025-10-14)
- **FIXED**: API 엔드포인트 경로 오타 수정 (/auth/loign → /auth/login)
- **AUTHOR**: @Charlie
```

##### REMOVED (제거)

```markdown
### v1.0.0 (2025-11-01)
- **REMOVED**: 구 세션 기반 인증 요구사항 제거
- **BREAKING**: 하위 호환성 깨짐
- **AUTHOR**: @Goos
```

##### DEPRECATED (폐기 예고)

```markdown
### v0.8.0 (2025-10-20)
- **DEPRECATED**: /auth/login-old 엔드포인트는 v1.0.0에서 제거 예정
- **AUTHOR**: @Alice
```

### 5.4. 필수 메타데이터

HISTORY 항목에는 다음 메타데이터를 포함할 수 있습니다.

#### AUTHOR (필수)
변경을 수행한 사람의 GitHub ID

```markdown
- **AUTHOR**: @Goos
```

#### REVIEW (선택)
리뷰어 및 승인 상태

```markdown
- **REVIEW**: @Alice (approved)
- **REVIEW**: @Bob (requested changes)
```

#### REASON (선택, 중요 변경 시 권장)
변경 이유

```markdown
- **REASON**: 사용자 경험 개선 요청
- **REASON**: 보안 취약점 발견으로 긴급 수정
```

#### RELATED (선택)
관련 이슈/PR 번호

```markdown
- **RELATED**: #123
- **RELATED**: #456, #789
```

#### SCOPE (선택)
변경 범위

```markdown
- **SCOPE**:
  - 토큰 만료 시간 변경
  - API 응답 형식 업데이트
```

### 5.5. HISTORY 작성 예시

#### 간단한 HISTORY

```markdown
## HISTORY

### v0.0.1 (2025-10-13)
- **INITIAL**: moai init 명령어 명세 작성
- **AUTHOR**: @Goos

### v0.0.2 (2025-10-14)
- **CHANGED**: 초기화 시간 제약 3초 → 5초로 완화
- **AUTHOR**: @Alice
- **REVIEW**: @Goos (approved)
```

#### 복잡한 HISTORY

```markdown
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

### v0.0.2 (2025-10-14)
- **CHANGED**: 토큰 만료 시간 15분 → 30분 변경
- **AUTHOR**: @Alice
- **REVIEW**: @Bob (approved)
- **REASON**: 사용자가 잦은 재로그인 불편 호소
- **RELATED**: #123

### v0.1.0 (2025-10-16)
- **ADDED**: 2FA 인증 요구사항 추가
- **AUTHOR**: @Charlie
- **REVIEW**: @Goos (approved), @Alice (approved)
- **REASON**: 보안 강화 필요
- **RELATED**: #145

### v0.2.0 (2025-10-18)
- **DEPRECATED**: /auth/login-old 엔드포인트는 v1.0.0에서 제거 예정
- **AUTHOR**: @Goos
- **REASON**: 새 /auth/login 엔드포인트로 통합

### v1.0.0 (2025-11-01)
- **REMOVED**: 구 세션 기반 인증 요구사항 제거
- **BREAKING**: 하위 호환성 깨짐 (세션 API 제거)
- **AUTHOR**: @Goos
- **REVIEW**: @Alice (approved), @Bob (approved), @Charlie (approved)
- **REASON**: JWT 전환 완료
- **RELATED**: #200
```

### 5.6. HISTORY 작성 모범 사례

#### 1. 최신 항목을 위에 작성

```markdown
## HISTORY

### v0.0.2 (2025-10-14)  ← 최신
- **CHANGED**: ...

### v0.0.1 (2025-10-13)  ← 이전
- **INITIAL**: ...
```

#### 2. 변경 유형 태그 명확히 사용

```markdown
✅ 좋은 예:
- **CHANGED**: 토큰 만료 시간 15분 → 30분 변경

❌ 나쁜 예:
- 토큰 만료 시간 변경됨
```

#### 3. AUTHOR는 항상 명시

```markdown
✅ 좋은 예:
- **AUTHOR**: @Alice

❌ 나쁜 예:
(AUTHOR 누락)
```

#### 4. 중요한 변경은 REASON 추가

```markdown
✅ 좋은 예:
- **CHANGED**: API 응답 형식 변경
- **REASON**: 프론트엔드 통합 요구사항 반영

❌ 나쁜 예:
- **CHANGED**: API 응답 형식 변경
(이유 불명확)
```

### 5.7. HISTORY 검색

#### 특정 TAG의 전체 변경 이력 조회

```bash
rg -A 20 "# @SPEC:AUTH-001" .moai/specs/SPEC-AUTH-001/spec.md
```

#### HISTORY 섹션만 추출

```bash
rg -A 50 "## HISTORY" .moai/specs/SPEC-AUTH-001/spec.md
```

#### 최근 변경 사항만 확인

```bash
rg "### v[0-9]" .moai/specs/SPEC-AUTH-001/spec.md | head -3
```

#### 특정 작성자의 변경 이력

```bash
rg "AUTHOR.*@Alice" .moai/specs/SPEC-*/spec.md
```

---

## 6. 버전 관리 체계

### 6.1. Semantic Versioning 개요

MoAI-ADK SPEC은 Semantic Versioning (SemVer) 2.0.0을 따릅니다.

```
MAJOR.MINOR.PATCH
```

- **MAJOR**: 하위 호환성 깨지는 변경
- **MINOR**: 기능 추가 (하위 호환)
- **PATCH**: 버그 수정 (하위 호환)

### 6.2. SPEC 버전 체계

#### v0.0.1 (INITIAL)
- **의미**: SPEC 최초 작성
- **status**: draft
- **작업**: 요구사항 작성, EARS 구문 적용

```yaml
version: 0.0.1
status: draft
```

```markdown
## HISTORY
### v0.0.1 (2025-10-13)
- **INITIAL**: JWT 인증 시스템 명세 작성
- **AUTHOR**: @Goos
```

#### v0.0.x (Draft 수정)
- **의미**: SPEC 문서 수정 및 개선
- **status**: draft
- **작업**: 오타 수정, 요구사항 명확화, 리뷰 반영

```yaml
version: 0.0.2
status: draft
```

```markdown
### v0.0.2 (2025-10-14)
- **CHANGED**: 토큰 만료 시간 명시 추가
- **FIXED**: API 경로 오타 수정
- **AUTHOR**: @Alice
```

#### v0.1.0 (TDD 구현 완료)
- **의미**: TDD 구현 완료, 테스트 통과
- **status**: completed
- **작업**: `/alfred:3-sync`가 자동으로 버전 업데이트

```yaml
version: 0.1.0
status: completed
```

```markdown
### v0.1.0 (2025-10-16)
- **ADDED**: TDD 구현 완료 (RED-GREEN-REFACTOR)
- **AUTHOR**: @Goos
- **TEST**: 테스트 커버리지 87%
- **CODE**: src/core/auth/ 구현 완료
```

#### v0.1.x (버그 수정)
- **의미**: 구현 후 버그 수정, 문서 개선
- **status**: completed
- **작업**: Patch 버전 증가

```yaml
version: 0.1.1
status: completed
```

```markdown
### v0.1.1 (2025-10-17)
- **FIXED**: 토큰 검증 로직 버그 수정
- **AUTHOR**: @Bob
```

#### v0.x.0 (기능 추가)
- **의미**: 새로운 기능 추가 (하위 호환 유지)
- **status**: completed
- **작업**: Minor 버전 증가

```yaml
version: 0.2.0
status: completed
```

```markdown
### v0.2.0 (2025-10-20)
- **ADDED**: 2FA 인증 기능 추가
- **AUTHOR**: @Alice
```

#### v1.0.0 (정식 버전)
- **의미**: 정식 안정화 버전, 프로덕션 준비
- **status**: completed
- **작업**: 사용자 명시적 승인 필수

```yaml
version: 1.0.0
status: completed
```

```markdown
### v1.0.0 (2025-11-01)
- **RELEASED**: 정식 안정화 버전
- **AUTHOR**: @Goos
- **REVIEW**: @Alice (approved), @Bob (approved)
```

### 6.3. 버전 증가 규칙 상세

#### Patch 증가 (x.y.Z)
다음 경우 Patch 버전을 증가합니다:

- SPEC 문서 오타 수정
- 요구사항 명확화 (의미 변경 없음)
- 코드 예시 오류 수정
- 주석 추가
- 버그 수정

```markdown
### v0.0.2 (2025-10-14)
- **FIXED**: API 경로 오타 수정 (/auth/loign → /auth/login)
- **CHANGED**: 요구사항 문구 명확화
```

#### Minor 증가 (x.Y.0)
다음 경우 Minor 버전을 증가합니다:

- 새로운 요구사항 추가
- 새로운 API 추가
- 선택적 기능 추가
- TDD 구현 완료 (v0.1.0)

```markdown
### v0.2.0 (2025-10-20)
- **ADDED**: 2FA 인증 요구사항 추가
- **ADDED**: /auth/2fa 엔드포인트 추가
```

#### Major 증가 (X.0.0)
다음 경우 Major 버전을 증가합니다:

- 하위 호환성 깨지는 변경
- 기존 요구사항 제거
- API 인터페이스 변경
- 정식 버전 릴리스 (v1.0.0)

```markdown
### v1.0.0 (2025-11-01)
- **REMOVED**: 구 세션 기반 인증 제거
- **BREAKING**: /auth/session-login 엔드포인트 제거
```

### 6.4. 버전 증가 예시 시나리오

#### 시나리오 1: SPEC 작성부터 구현 완료까지

```
v0.0.1 (INITIAL) → v0.0.2 (문서 수정) → v0.0.3 (리뷰 반영) → v0.1.0 (TDD 완료)
```

```markdown
## HISTORY

### v0.1.0 (2025-10-16)
- **ADDED**: TDD 구현 완료
- **AUTHOR**: @Goos

### v0.0.3 (2025-10-15)
- **CHANGED**: 리뷰 피드백 반영
- **AUTHOR**: @Alice

### v0.0.2 (2025-10-14)
- **FIXED**: 오타 수정
- **AUTHOR**: @Goos

### v0.0.1 (2025-10-13)
- **INITIAL**: JWT 인증 시스템 명세 작성
- **AUTHOR**: @Goos
```

#### 시나리오 2: 구현 후 기능 추가

```
v0.1.0 (TDD 완료) → v0.1.1 (버그 수정) → v0.2.0 (기능 추가) → v1.0.0 (정식 버전)
```

```markdown
## HISTORY

### v1.0.0 (2025-11-01)
- **RELEASED**: 정식 안정화 버전
- **AUTHOR**: @Goos

### v0.2.0 (2025-10-20)
- **ADDED**: 2FA 인증 추가
- **AUTHOR**: @Alice

### v0.1.1 (2025-10-17)
- **FIXED**: 토큰 검증 버그 수정
- **AUTHOR**: @Bob

### v0.1.0 (2025-10-16)
- **ADDED**: TDD 구현 완료
- **AUTHOR**: @Goos
```

---

## 7. 메타데이터 검증

### 7.1. 필수 필드 검증

#### 모든 SPEC에 필수 필드 있는지 확인

```bash
rg "^(id|version|status|created|updated|author|priority):" .moai/specs/SPEC-*/spec.md
```

#### 특정 필드 누락 확인

```bash
# priority 필드 누락 확인
rg -L "^priority:" .moai/specs/SPEC-*/spec.md

# author 필드 누락 확인
rg -L "^author:" .moai/specs/SPEC-*/spec.md
```

#### 출력 예시

```
.moai/specs/SPEC-AUTH-001/spec.md:id: AUTH-001
.moai/specs/SPEC-AUTH-001/spec.md:version: 0.0.1
.moai/specs/SPEC-AUTH-001/spec.md:status: draft
.moai/specs/SPEC-AUTH-001/spec.md:created: 2025-10-13
.moai/specs/SPEC-AUTH-001/spec.md:updated: 2025-10-13
.moai/specs/SPEC-AUTH-001/spec.md:author: @Goos
.moai/specs/SPEC-AUTH-001/spec.md:priority: high
```

### 7.2. 형식 검증

#### author 필드 형식 확인 (@Username)

```bash
rg "^author: @[A-Z]" .moai/specs/SPEC-*/spec.md
```

#### version 필드 형식 확인 (0.x.y)

```bash
rg "^version: 0\.\d+\.\d+" .moai/specs/SPEC-*/spec.md
```

#### 날짜 형식 확인 (YYYY-MM-DD)

```bash
rg "^created: \d{4}-\d{2}-\d{2}" .moai/specs/SPEC-*/spec.md
rg "^updated: \d{4}-\d{2}-\d{2}" .moai/specs/SPEC-*/spec.md
```

#### status 유효성 확인

```bash
rg "^status: (draft|active|completed|deprecated)" .moai/specs/SPEC-*/spec.md
```

#### priority 유효성 확인

```bash
rg "^priority: (critical|high|medium|low)" .moai/specs/SPEC-*/spec.md
```

### 7.3. HISTORY 섹션 검증

#### HISTORY 섹션 존재 확인

```bash
rg "## HISTORY" .moai/specs/SPEC-*/spec.md
```

#### INITIAL 항목 존재 확인

```bash
rg "INITIAL" .moai/specs/SPEC-*/spec.md
```

#### AUTHOR 메타데이터 존재 확인

```bash
rg "AUTHOR.*@" .moai/specs/SPEC-*/spec.md
```

### 7.4. 중복 ID 검증

#### 특정 ID 중복 확인

```bash
rg "@SPEC:AUTH-001" -n .moai/specs/
```

#### 모든 SPEC ID 목록

```bash
rg "^id: " .moai/specs/SPEC-*/spec.md | sort
```

#### 중복 ID 탐지

```bash
rg "^id: " .moai/specs/SPEC-*/spec.md | awk '{print $2}' | sort | uniq -d
```

### 7.5. 자동 검증 스크립트

#### validate-spec-metadata.sh

```bash
#!/bin/bash
# SPEC 메타데이터 검증 스크립트

echo "=== SPEC 메타데이터 검증 ==="

# 1. 필수 필드 검증
echo ""
echo "1. 필수 필드 누락 확인..."
for field in id version status created updated author priority; do
  missing=$(rg -L "^$field:" .moai/specs/SPEC-*/spec.md)
  if [ -n "$missing" ]; then
    echo "  ❌ $field 필드 누락: $missing"
  else
    echo "  ✅ $field 필드 모두 존재"
  fi
done

# 2. 형식 검증
echo ""
echo "2. 형식 검증..."

# author 형식 확인
invalid_author=$(rg "^author: " .moai/specs/SPEC-*/spec.md | grep -v "@")
if [ -n "$invalid_author" ]; then
  echo "  ❌ author 형식 오류:"
  echo "$invalid_author"
else
  echo "  ✅ author 형식 올바름"
fi

# version 형식 확인
invalid_version=$(rg "^version: " .moai/specs/SPEC-*/spec.md | grep -v -E "\d+\.\d+\.\d+")
if [ -n "$invalid_version" ]; then
  echo "  ❌ version 형식 오류:"
  echo "$invalid_version"
else
  echo "  ✅ version 형식 올바름"
fi

# 3. HISTORY 섹션 확인
echo ""
echo "3. HISTORY 섹션 확인..."
missing_history=$(rg -L "## HISTORY" .moai/specs/SPEC-*/spec.md)
if [ -n "$missing_history" ]; then
  echo "  ❌ HISTORY 섹션 누락: $missing_history"
else
  echo "  ✅ HISTORY 섹션 모두 존재"
fi

# 4. 중복 ID 확인
echo ""
echo "4. 중복 ID 확인..."
duplicates=$(rg "^id: " .moai/specs/SPEC-*/spec.md | awk '{print $2}' | sort | uniq -d)
if [ -n "$duplicates" ]; then
  echo "  ❌ 중복 ID 발견:"
  echo "$duplicates"
else
  echo "  ✅ 중복 ID 없음"
fi

echo ""
echo "=== 검증 완료 ==="
```

실행:

```bash
chmod +x validate-spec-metadata.sh
./validate-spec-metadata.sh
```

---

## 8. 의존성 그래프 관리

### 8.1. 의존성 종류

MoAI-ADK SPEC은 3가지 의존성 관계를 지원합니다:

1. **depends_on**: "이 SPEC은 {depends_on}이 완료되어야 시작 가능"
2. **blocks**: "이 SPEC이 완료되어야 {blocks}를 시작 가능"
3. **related_specs**: "이 SPEC과 관련 있지만 의존성 없음"

### 8.2. 의존성 추출

#### 모든 depends_on 추출

```bash
rg "^depends_on:" -A 5 .moai/specs/SPEC-*/spec.md
```

#### 모든 blocks 추출

```bash
rg "^blocks:" -A 5 .moai/specs/SPEC-*/spec.md
```

#### 모든 related_specs 추출

```bash
rg "^related_specs:" -A 5 .moai/specs/SPEC-*/spec.md
```

### 8.3. 의존성 그래프 시각화

#### Graphviz dot 파일 생성

```bash
#!/bin/bash
# generate-spec-dependency-graph.sh

echo "digraph SPEC_Dependencies {" > spec-deps.dot
echo "  rankdir=TB;" >> spec-deps.dot
echo "  node [shape=box];" >> spec-deps.dot
echo "" >> spec-deps.dot

# 모든 SPEC 노드 추가
rg "^id: " .moai/specs/SPEC-*/spec.md | awk '{print "  " $2 ";"}' >> spec-deps.dot

echo "" >> spec-deps.dot

# depends_on 관계 추가
for spec_file in .moai/specs/SPEC-*/spec.md; do
  spec_id=$(rg "^id: " "$spec_file" | awk '{print $2}')
  depends=$(rg "^depends_on:" -A 10 "$spec_file" | grep "  - " | awk '{print $2}')

  for dep in $depends; do
    echo "  $dep -> $spec_id [label=\"depends_on\"];" >> spec-deps.dot
  done
done

echo "}" >> spec-deps.dot
echo "Graphviz dot 파일 생성: spec-deps.dot"
```

#### PNG 이미지 생성

```bash
dot -Tpng spec-deps.dot -o spec-deps.png
```

### 8.4. 순환 의존성 탐지

#### 순환 의존성 검사 스크립트

```bash
#!/bin/bash
# detect-circular-dependencies.sh

# 간단한 순환 의존성 검사 (깊이 2)
echo "=== 순환 의존성 검사 ==="

for spec_file in .moai/specs/SPEC-*/spec.md; do
  spec_id=$(rg "^id: " "$spec_file" | awk '{print $2}')
  depends=$(rg "^depends_on:" -A 10 "$spec_file" | grep "  - " | awk '{print $2}')

  for dep in $depends; do
    # dep의 depends_on 확인
    dep_file=".moai/specs/SPEC-$dep/spec.md"
    if [ -f "$dep_file" ]; then
      dep_depends=$(rg "^depends_on:" -A 10 "$dep_file" | grep "  - " | awk '{print $2}')

      for dd in $dep_depends; do
        if [ "$dd" == "$spec_id" ]; then
          echo "  ❌ 순환 의존성 발견: $spec_id ↔ $dep"
        fi
      done
    fi
  done
done

echo "=== 검사 완료 ==="
```

---

## 9. 실전 예시

### 9.1. 간단한 SPEC 예시

```yaml
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
```

### 9.2. 복잡한 SPEC 예시

```yaml
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
related_issue: "https://github.com/modu-ai/moai-adk/issues/123"
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
```

### 9.3. 버전 업데이트 예시

#### v0.0.1 → v0.0.2 (Patch)

```yaml
version: 0.0.2
updated: 2025-10-14
```

```markdown
### v0.0.2 (2025-10-14)
- **CHANGED**: 토큰 만료 시간 15분 → 30분 변경
- **FIXED**: API 경로 오타 수정
- **AUTHOR**: @Alice
- **REVIEW**: @Bob (approved)
```

#### v0.0.2 → v0.1.0 (Minor, TDD 완료)

```yaml
version: 0.1.0
status: completed
updated: 2025-10-16
```

```markdown
### v0.1.0 (2025-10-16)
- **ADDED**: TDD 구현 완료 (RED-GREEN-REFACTOR)
- **AUTHOR**: @Goos
- **TEST**: 테스트 커버리지 87%
- **CODE**: src/core/auth/ 구현 완료
```

---

## 10. 마이그레이션 가이드

### 10.1. 기존 SPEC 업데이트

#### 1단계: priority 필드 추가
기존 SPEC에 priority 필드가 없다면 추가합니다.

```yaml
# 이전
id: AUTH-001
version: 0.0.1
status: draft

# 이후
id: AUTH-001
version: 0.0.1
status: draft
priority: high  # 추가
```

#### 2단계: author 필드 표준화
- `authors: ["@goos"]` → `author: @Goos`
- 소문자 → 대문자로 변경

```yaml
# 이전
authors: ["@goos"]

# 이후
author: @Goos
```

#### 3단계: 선택 필드 추가 (권장)

```yaml
category: refactor
labels:
  - code-quality
  - maintenance
```

#### 4단계: HISTORY 섹션 추가

```markdown
## HISTORY

### v0.0.1 (2025-10-13)
- **INITIAL**: [SPEC 제목] 명세 작성
- **AUTHOR**: @Goos
```

### 10.2. 일괄 업데이트 스크립트

```bash
#!/bin/bash
# migrate-spec-metadata.sh

echo "=== SPEC 메타데이터 마이그레이션 ==="

for spec_file in .moai/specs/SPEC-*/spec.md; do
  echo "처리 중: $spec_file"

  # priority 필드 누락 확인 및 추가
  if ! grep -q "^priority:" "$spec_file"; then
    echo "  → priority 필드 추가"
    # YAML Front Matter 끝나는 --- 앞에 priority 추가
    # (실제 구현 시 awk 또는 sed 사용)
  fi

  # HISTORY 섹션 누락 확인
  if ! grep -q "## HISTORY" "$spec_file"; then
    echo "  → HISTORY 섹션 추가 필요 (수동 작업)"
  fi
done

echo "=== 마이그레이션 완료 ==="
```

---

## 11. 자동화 도구

### 11.1. SPEC 템플릿 생성 스크립트

```bash
#!/bin/bash
# create-spec.sh - SPEC 템플릿 자동 생성

SPEC_ID=$1
AUTHOR=$2

if [ -z "$SPEC_ID" ] || [ -z "$AUTHOR" ]; then
  echo "Usage: ./create-spec.sh <SPEC-ID> <@Author>"
  exit 1
fi

# 중복 확인
if rg -q "@SPEC:$SPEC_ID" .moai/specs/; then
  echo "❌ SPEC ID 중복: $SPEC_ID"
  exit 1
fi

# 디렉토리 생성
SPEC_DIR=".moai/specs/SPEC-$SPEC_ID"
mkdir -p "$SPEC_DIR"

# 템플릿 생성
cat > "$SPEC_DIR/spec.md" <<EOF
---
id: $SPEC_ID
version: 0.0.1
status: draft
created: $(date +%Y-%m-%d)
updated: $(date +%Y-%m-%d)
author: $AUTHOR
priority: medium
category: feature
labels:
  - todo
---

# @SPEC:$SPEC_ID: [제목 작성]

## HISTORY

### v0.0.1 ($(date +%Y-%m-%d))
- **INITIAL**: [SPEC 제목] 명세 작성
- **AUTHOR**: $AUTHOR

---

## 개요
[1-2 문단으로 SPEC 설명]

---

## Environment (환경 및 전제조건)

### 기술 스택
- [기술 나열]

---

## Requirements (요구사항)

### Ubiquitous Requirements (기본 요구사항)
- 시스템은 [기능]을 제공해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN [조건]이면, 시스템은 [동작]해야 한다

### Constraints (제약사항)
- IF [조건]이면, 시스템은 [제약]해야 한다

---

## Specifications (상세 명세)

### 1. [컴포넌트명]
[구현 가이드]

---

## Traceability (추적성)
- **SPEC ID**: @SPEC:$SPEC_ID
- **TAG 체인**: @SPEC:$SPEC_ID → @TEST:$SPEC_ID → @CODE:$SPEC_ID
EOF

echo "✅ SPEC 생성 완료: $SPEC_DIR/spec.md"
```

사용:

```bash
./create-spec.sh AUTH-001 @Goos
```

### 11.2. SPEC 버전 업데이트 스크립트

```bash
#!/bin/bash
# update-spec-version.sh

SPEC_ID=$1
NEW_VERSION=$2
CHANGE_TYPE=$3
AUTHOR=$4

if [ -z "$SPEC_ID" ] || [ -z "$NEW_VERSION" ]; then
  echo "Usage: ./update-spec-version.sh <SPEC-ID> <new-version> <CHANGE_TYPE> <@Author>"
  exit 1
fi

SPEC_FILE=".moai/specs/SPEC-$SPEC_ID/spec.md"

if [ ! -f "$SPEC_FILE" ]; then
  echo "❌ SPEC 파일 없음: $SPEC_FILE"
  exit 1
fi

# 버전 업데이트
sed -i "" "s/^version: .*/version: $NEW_VERSION/" "$SPEC_FILE"

# updated 날짜 업데이트
sed -i "" "s/^updated: .*/updated: $(date +%Y-%m-%d)/" "$SPEC_FILE"

# HISTORY 추가 (수동 작업 필요)
echo "✅ 버전 업데이트 완료: $NEW_VERSION"
echo "⚠️ HISTORY 섹션을 수동으로 업데이트하세요"
```

---

## 12. FAQ

### Q1. 필수 필드를 모두 포함해야 하나요?

**A**: 네, 필수 필드 7개는 반드시 포함해야 합니다.

- id, version, status, created, updated, author, priority

누락 시 자동 검증 도구에서 에러가 발생합니다.

### Q2. version은 항상 0.0.1로 시작하나요?

**A**: 네, 모든 SPEC은 v0.0.1로 시작합니다.

- v0.0.1: INITIAL (최초 작성)
- v0.1.0: TDD 구현 완료
- v1.0.0: 정식 안정화 버전

### Q3. HISTORY 섹션은 필수인가요?

**A**: 네, HISTORY 섹션은 필수입니다.

- 최소 v0.0.1 INITIAL 항목 포함
- 변경 이력 추적 필수

### Q4. author는 단수형인가요, 복수형인가요?

**A**: 단수형 `author`를 사용합니다.

- ❌ `authors: ["@Goos"]` (잘못됨)
- ✅ `author: @Goos` (올바름)

복수 작성자는 HISTORY 섹션에 기록합니다.

### Q5. depends_on과 related_specs의 차이는?

**A**: 의존성 유무가 다릅니다.

- `depends_on`: 의존성 있음 (필수, 구현 순서 영향)
- `related_specs`: 의존성 없음 (참고용, 순서 무관)

### Q6. priority를 어떻게 결정하나요?

**A**: 다음 기준으로 결정합니다.

- `critical`: 보안, 중대 버그
- `high`: 핵심 기능
- `medium`: 개선사항
- `low`: 최적화, 문서

### Q7. YAML Front Matter에 주석을 달 수 있나요?

**A**: 네, `#`를 사용하여 주석을 달 수 있습니다.

```yaml
---
id: AUTH-001  # 인증 시스템
version: 0.0.1  # INITIAL 버전
---
```

### Q8. 날짜 형식은 왜 YYYY-MM-DD인가요?

**A**: ISO 8601 표준 형식으로, 정렬과 파싱이 용이합니다.

```yaml
created: 2025-10-13  # 올바름
created: 10/13/2025  # 잘못됨
```

### Q9. 선택 필드를 나중에 추가할 수 있나요?

**A**: 네, 언제든지 추가 가능합니다.

```yaml
# 최초
id: AUTH-001
version: 0.0.1

# 이후 추가
id: AUTH-001
version: 0.0.2
category: security  # 추가
labels:             # 추가
  - authentication
```

### Q10. 메타데이터 검증 도구가 있나요?

**A**: 네, `validate-spec-metadata.sh` 스크립트를 사용하세요.

```bash
./validate-spec-metadata.sh
```

---

## 부록

### A. 메타데이터 필드 요약표

| 필드 | 필수 | 타입 | 형식 | 기본값 |
|------|------|------|------|--------|
| `id` | ✅ | string | `<DOMAIN>-<NUMBER>` | - |
| `version` | ✅ | string | `MAJOR.MINOR.PATCH` | 0.0.1 |
| `status` | ✅ | enum | draft\|active\|completed\|deprecated | draft |
| `created` | ✅ | date | YYYY-MM-DD | - |
| `updated` | ✅ | date | YYYY-MM-DD | - |
| `author` | ✅ | string | @{GitHub ID} | - |
| `priority` | ✅ | enum | critical\|high\|medium\|low | - |
| `category` | ⚠️ | enum | feature\|bugfix\|refactor\|security\|docs\|perf | - |
| `labels` | ⚠️ | array | string[] | - |
| `depends_on` | ⚠️ | array | SPEC ID[] | - |
| `blocks` | ⚠️ | array | SPEC ID[] | - |
| `related_specs` | ⚠️ | array | SPEC ID[] | - |
| `related_issue` | ⚠️ | string | GitHub Issue URL | - |
| `scope.packages` | ⚠️ | array | path[] | - |
| `scope.files` | ⚠️ | array | filename[] | - |

✅ 필수, ⚠️ 선택

### B. HISTORY 변경 유형 태그 요약

| 태그 | 의미 | 버전 증가 |
|------|------|----------|
| `INITIAL` | 최초 작성 | v0.0.1 |
| `ADDED` | 기능/요구사항 추가 | Minor |
| `CHANGED` | 기존 내용 수정 | Patch |
| `FIXED` | 버그/오류 수정 | Patch |
| `REMOVED` | 기능/요구사항 제거 | Major |
| `BREAKING` | 하위 호환성 깨짐 | Major |
| `DEPRECATED` | 폐기 예정 | Minor |

### C. 검증 명령어 치트시트

```bash
# 필수 필드 검증
rg "^(id|version|status|created|updated|author|priority):" .moai/specs/SPEC-*/spec.md

# 형식 검증
rg "^author: @[A-Z]" .moai/specs/SPEC-*/spec.md
rg "^version: 0\.\d+\.\d+" .moai/specs/SPEC-*/spec.md
rg "^created: \d{4}-\d{2}-\d{2}" .moai/specs/SPEC-*/spec.md

# HISTORY 검증
rg "## HISTORY" .moai/specs/SPEC-*/spec.md
rg "INITIAL" .moai/specs/SPEC-*/spec.md

# 중복 ID 검증
rg "^id: " .moai/specs/SPEC-*/spec.md | awk '{print $2}' | sort | uniq -d

# 의존성 추출
rg "^depends_on:" -A 5 .moai/specs/SPEC-*/spec.md
```

### D. 참고 자료

- **Semantic Versioning**: https://semver.org/
- **YAML 1.2 Spec**: https://yaml.org/spec/1.2/spec.html
- **ISO 8601 (날짜 형식)**: https://www.iso.org/iso-8601-date-and-time-format.html
- **MoAI-ADK 개발 가이드**: `.moai/memory/development-guide.md`
- **SPEC 시스템 개요**: `docs/specs/overview.md`

---

**최종 업데이트**: 2025-10-14
**작성자**: @Alfred
**버전**: 1.0.0
