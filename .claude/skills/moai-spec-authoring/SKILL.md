---
name: moai-spec-authoring
version: 1.1.0
created: 2025-10-23
updated: 2025-10-27
status: active
description: SPEC 문서 작성 가이드 - YAML 메타데이터, EARS 문법, 검증 체크리스트 제공
keywords: ['spec', 'authoring', 'ears', 'metadata', 'requirements', 'tdd', 'planning']
allowed-tools:
  - Read
  - Bash
  - Glob
---

# SPEC Authoring Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-spec-authoring |
| **Version** | 1.1.0 (2025-10-27) |
| **Allowed tools** | Read, Bash, Glob |
| **Auto-load** | `/alfred:1-plan`, SPEC 작성 태스크 |
| **Tier** | Foundation |

---

## What It Does

MoAI-ADK SPEC 문서 작성을 위한 종합 가이드입니다. YAML 메타데이터 구조(7개 필수 + 9개 선택 필드), EARS 요구사항 문법(5개 패턴), 버전 관리 생애주기, TAG 통합, 검증 전략을 제공합니다.

**핵심 기능**:
- 단계별 SPEC 생성 워크플로우
- 완전한 메타데이터 필드 레퍼런스
- EARS 문법 템플릿과 실전 패턴
- 제출 전 검증 체크리스트
- 일반적인 함정 예방 가이드
- `/alfred:1-plan` 워크플로우 통합

---

## When to Use

**자동 트리거**:
- `/alfred:1-plan` 명령어 실행
- SPEC 문서 생성 요청
- 요구사항 명확화 논의
- 기능 계획 세션

**수동 호출**:
- SPEC 작성 모범 사례 학습
- 기존 SPEC 문서 검증
- 메타데이터 이슈 트러블슈팅
- EARS 문법 패턴 이해

---

## Quick Start: 5-Step SPEC Creation

### Step 1: SPEC 디렉토리 초기화

```bash
mkdir -p .moai/specs/SPEC-{DOMAIN}-{NUMBER}
# 예시: 인증 기능
mkdir -p .moai/specs/SPEC-AUTH-001
```

### Step 2: YAML Front Matter 작성

```yaml
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-23
updated: 2025-10-23
author: @YourGitHubHandle
priority: high
---
```

### Step 3: SPEC 타이틀 & HISTORY 추가

```markdown
# @SPEC:AUTH-001: JWT Authentication System

## HISTORY

### v0.0.1 (2025-10-23)
- **INITIAL**: JWT 인증 SPEC 초안 작성
- **AUTHOR**: @YourHandle
```

### Step 4: Environment & Assumptions 정의

```markdown
## Environment

**Runtime**: Node.js 20.x 이상
**Framework**: Express.js
**Database**: PostgreSQL 15+

## Assumptions

1. 사용자 인증 정보는 PostgreSQL에 저장
2. JWT 시크릿은 환경변수로 관리
3. 서버 시계는 NTP로 동기화
```

### Step 5: EARS 요구사항 작성

```markdown
## Requirements

### Ubiquitous Requirements
**UR-001**: 시스템은 JWT 기반 인증을 제공해야 한다.

### Event-driven Requirements
**ER-001**: WHEN 사용자가 유효한 인증 정보를 제출하면, 시스템은 15분 만료 시간을 가진 JWT 토큰을 발급해야 한다.

### State-driven Requirements
**SR-001**: WHILE 사용자가 인증된 상태이면, 시스템은 보호된 리소스에 대한 접근을 허용해야 한다.

### Optional Features
**OF-001**: WHERE 다중 인증이 활성화된 경우, 시스템은 비밀번호 확인 후 OTP 검증을 요구할 수 있다.

### Constraints
**C-001**: IF 토큰이 만료되었다면, 시스템은 접근을 거부하고 HTTP 401을 반환해야 한다.
```

---

## EARS 5가지 패턴 개요

| 패턴 | 키워드 | 용도 | 예시 |
|------|--------|------|------|
| **Ubiquitous** | shall | 항상 활성화된 핵심 기능 | "시스템은 로그인을 제공해야 한다" |
| **Event-driven** | WHEN | 특정 이벤트에 대한 응답 | "WHEN 로그인 실패 시, 에러 표시" |
| **State-driven** | WHILE | 상태 중 지속적 동작 | "WHILE 로그인 상태, 접근 허용" |
| **Optional** | WHERE | 기능 플래그 기반 조건부 | "WHERE 프리미엄이면, 기능 해제" |
| **Constraints** | IF-THEN | 품질 게이트, 비즈니스 규칙 | "IF 만료되었다면, 거부" |

---

## 필수 메타데이터 7개 필드

1. **id**: `<DOMAIN>-<NUMBER>` (예: `AUTH-001`) - 불변 식별자
2. **version**: `MAJOR.MINOR.PATCH` (예: `0.0.1`) - 시맨틱 버전
3. **status**: `draft` | `active` | `completed` | `deprecated`
4. **created**: `YYYY-MM-DD` - 최초 생성일
5. **updated**: `YYYY-MM-DD` - 최종 수정일
6. **author**: `@GitHubHandle` - 주 작성자 (@ 접두사 필수)
7. **priority**: `critical` | `high` | `medium` | `low`

**버전 생애주기**:
- `0.0.x` → draft (초안 작성 중)
- `0.1.0` → completed (구현 완료)
- `1.0.0` → stable (프로덕션 안정)

---

## 검증 체크리스트

### 메타데이터 검증
- [ ] 7개 필수 필드 모두 존재
- [ ] `author` 필드에 @ 접두사 포함
- [ ] `version` 형식이 `0.x.y` 형식
- [ ] `id`가 중복되지 않음 (`rg "@SPEC:AUTH-001" -n .moai/specs/`)

### 콘텐츠 검증
- [ ] YAML Front Matter 완성
- [ ] 타이틀에 `@SPEC:{ID}` TAG 블록
- [ ] HISTORY 섹션에 v0.0.1 INITIAL 엔트리
- [ ] Environment 섹션 정의
- [ ] Assumptions 섹션 정의 (최소 3개)
- [ ] Requirements 섹션에 EARS 패턴 사용
- [ ] Traceability 섹션에 TAG 체인 구조

### EARS 문법 검증
- [ ] Ubiquitous: "shall" + 능력 표현
- [ ] Event-driven: "WHEN [트리거]"로 시작
- [ ] State-driven: "WHILE [상태]"로 시작
- [ ] Optional: "WHERE [기능]"로 시작, "can" 사용
- [ ] Constraints: "IF-THEN" 또는 직접 제약 표현

---

## 일반적인 함정

1. ❌ **ID 변경**: 할당 후 SPEC ID 변경 → TAG 체인 파괴
2. ❌ **HISTORY 누락**: 콘텐츠 변경 시 HISTORY 업데이트 생략
3. ❌ **잘못된 버전 진행**: v0.0.1 → v1.0.0으로 건너뛰기
4. ❌ **모호한 요구사항**: "빠르고 사용자 친화적" 같은 측정 불가능한 표현
5. ❌ **@ 접두사 누락**: `author: Goos` 대신 `author: @Goos`
6. ❌ **EARS 패턴 혼합**: 하나의 요구사항에 여러 키워드 혼용

---

## Related Skills

- `moai-foundation-ears` - EARS 문법 패턴
- `moai-foundation-specs` - 메타데이터 검증
- `moai-foundation-tags` - TAG 시스템 통합
- `moai-alfred-spec-metadata-validation` - 자동화된 검증

---

## 상세 정보

- **메타데이터 레퍼런스**: [reference.md](./reference.md) 참조
- **실전 예제**: [examples.md](./examples.md) 참조

---

**Last Updated**: 2025-10-27
**Maintained By**: MoAI-ADK Team
**Support**: `/alfred:1-plan` 명령어로 가이드 받기
