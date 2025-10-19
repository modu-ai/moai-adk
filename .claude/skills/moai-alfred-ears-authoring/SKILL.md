---
name: moai-alfred-ears-authoring
description: EARS (Easy Approach to Requirements Syntax) authoring guide with 5 statement patterns for clear, testable requirements
version: 0.1.0
author: MoAI Skill Factory
license: MIT
tags:
  - spec
  - ears
  - requirements
  - authoring
---

# Alfred EARS Authoring Guide

## What it does

EARS (Easy Approach to Requirements Syntax) authoring guide for writing clear, testable requirements using 5 statement patterns.

## When to use

- "SPEC 작성", "요구사항 정리", "EARS 구문"
- Automatically invoked by `/alfred:1-plan`
- When writing or refining SPEC documents

## How it works

<!-- @CODE:UPDATE-004:PHASE2 -->

EARS (Easy Approach to Requirements Syntax)는 체계적인 요구사항 작성 방법론으로, 5가지 구문 패턴을 제공합니다.

### 1. Ubiquitous (기본 요구사항)

**구문**: 시스템은 [기능]을 제공해야 한다

**특징**:
- 항상 참이며 조건이 없는 기본 요구사항
- 시스템이 제공해야 하는 핵심 기능 정의
- "must", "shall", "해야 한다" 등의 의무 표현 사용

**예시**:
```markdown
### Ubiquitous Requirements (기본 요구사항)
- 시스템은 사용자 인증 기능을 제공해야 한다
- 시스템은 데이터 암호화 기능을 제공해야 한다
- 시스템은 로그 기록 기능을 제공해야 한다
```

### 2. Event-driven (이벤트 기반)

**구문**: WHEN [조건]이면, 시스템은 [동작]해야 한다

**특징**:
- 특정 이벤트 발생 시의 시스템 동작 정의
- 트리거 조건과 결과 동작을 명확히 분리
- 비동기 처리, API 호출, 사용자 액션 등에 적합

**예시**:
```markdown
### Event-driven Requirements (이벤트 기반)
- WHEN 사용자가 유효한 자격증명으로 로그인하면, 시스템은 JWT 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 시스템은 401 에러를 반환해야 한다
- WHEN 파일 업로드가 완료되면, 시스템은 업로드 성공 알림을 전송해야 한다
```

### 3. State-driven (상태 기반)

**구문**: WHILE [상태]일 때, 시스템은 [동작]해야 한다

**특징**:
- 특정 상태가 유지되는 동안의 시스템 동작 정의
- 지속적인 조건 체크와 상태 관리
- 인증 세션, 연결 상태, 작업 진행 상태 등에 적합

**예시**:
```markdown
### State-driven Requirements (상태 기반)
- WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 리소스 접근을 허용해야 한다
- WHILE 네트워크가 연결된 상태일 때, 시스템은 실시간 데이터 동기화를 수행해야 한다
- WHILE 파일이 업로드 중일 때, 시스템은 진행률 표시줄을 업데이트해야 한다
```

### 4. Optional (선택적 기능)

**구문**: WHERE [조건]이면, 시스템은 [동작]할 수 있다

**특징**:
- 선택적으로 제공되는 기능 정의
- "may", "can", "할 수 있다" 등의 허용 표현 사용
- 확장 기능, 최적화, 옵션 기능 등에 적합

**예시**:
```markdown
### Optional Features (선택적 기능)
- WHERE 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급할 수 있다
- WHERE 관리자 권한이 있으면, 시스템은 상세 로그를 제공할 수 있다
- WHERE 고해상도 이미지가 요청되면, 시스템은 원본 크기로 반환할 수 있다
```

### 5. Constraints (제약사항)

**구문**: IF [조건]이면, 시스템은 [제약]해야 한다

**특징**:
- 시스템의 제한사항과 보안 정책 정의
- 금지 행위, 최대/최소 값, 유효성 검증 규칙 등
- "must not", "shall not", "해야 한다" 등의 강제 표현

**예시**:
```markdown
### Constraints (제약사항)
- IF 잘못된 토큰이 제공되면, 시스템은 접근을 거부해야 한다
- 액세스 토큰 만료시간은 15분을 초과하지 않아야 한다
- 비밀번호는 8자 이상이어야 하며, 특수문자를 포함해야 한다
- 파일 크기는 10MB를 초과할 수 없다
```

## Writing Tips

<!-- @CODE:UPDATE-004:PHASE2 -->

### EARS 작성 원칙

✅ **구체적이고 측정 가능하게 작성**
- ❌ "시스템은 빠르게 응답해야 한다"
- ✅ "시스템은 API 요청에 대해 500ms 이내에 응답해야 한다"

✅ **모호한 표현 지양**
- ❌ "적절한", "충분한", "빠른", "많은" 등의 주관적 표현
- ✅ 숫자, 단위, 명확한 기준 사용

✅ **하나의 문장에 하나의 요구사항**
- ❌ "시스템은 로그인하고 토큰을 발급하고 세션을 생성해야 한다"
- ✅ 각각을 별도 문장으로 분리

✅ **테스트 가능하게 작성**
- 모든 요구사항은 자동화 테스트로 검증 가능해야 함
- Given-When-Then 형식으로 변환 가능해야 함

✅ **부정문보다 긍정문 사용**
- ❌ "시스템은 유효하지 않은 입력을 허용하지 않아야 한다"
- ✅ "시스템은 유효한 입력만 허용해야 한다"

### 구문별 키워드

| 구문 | 키워드 | 용도 |
|------|--------|------|
| Ubiquitous | 시스템은 [기능]을 제공해야 한다 | 핵심 기능 |
| Event-driven | WHEN [조건], 시스템은 [동작]해야 한다 | 이벤트 처리 |
| State-driven | WHILE [상태], 시스템은 [동작]해야 한다 | 상태 유지 |
| Optional | WHERE [조건], 시스템은 [동작]할 수 있다 | 선택 기능 |
| Constraints | IF [조건], 시스템은 [제약]해야 한다 | 제약사항 |

## Examples

<!-- @CODE:UPDATE-004:PHASE2 -->

### 실제 JWT 인증 SPEC 작성 예시

**사용자 요청**:
```
"JWT 인증 SPEC 작성해줘"
```

**EARS 적용 결과**:

```markdown
### Ubiquitous Requirements (기본 요구사항)
- 시스템은 JWT 기반 인증 기능을 제공해야 한다
- 시스템은 토큰 발급 API를 제공해야 한다
- 시스템은 토큰 검증 미들웨어를 제공해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 사용자가 유효한 자격증명으로 로그인하면, 시스템은 액세스 토큰과 리프레시 토큰을 발급해야 한다
- WHEN 액세스 토큰이 만료되면, 시스템은 401 Unauthorized 에러를 반환해야 한다
- WHEN 유효하지 않은 토큰으로 요청하면, 시스템은 403 Forbidden 에러를 반환해야 한다
- WHEN 리프레시 토큰으로 갱신 요청하면, 시스템은 새로운 액세스 토큰을 발급해야 한다

### State-driven Requirements (상태 기반)
- WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 엔드포인트 접근을 허용해야 한다
- WHILE 토큰이 유효한 상태일 때, 시스템은 사용자 컨텍스트를 요청 객체에 주입해야 한다

### Optional Features (선택적 기능)
- WHERE 관리자 권한이 있으면, 시스템은 다른 사용자의 토큰을 무효화할 수 있다
- WHERE 토큰 갱신 옵션이 활성화되면, 시스템은 자동 토큰 갱신을 제공할 수 있다

### Constraints (제약사항)
- 액세스 토큰 만료시간은 15분을 초과하지 않아야 한다
- 리프레시 토큰 만료시간은 7일을 초과하지 않아야 한다
- IF 동일 사용자의 동시 로그인 수가 5개를 초과하면, 시스템은 가장 오래된 세션을 무효화해야 한다
- 토큰 페이로드는 개인정보(비밀번호, 주민번호 등)를 포함해서는 안 된다
```

### 파일 업로드 기능 SPEC 예시

**사용자 요청**:
```
"파일 업로드 기능 명세 작성"
```

**EARS 적용 결과**:

```markdown
### Ubiquitous Requirements (기본 요구사항)
- 시스템은 멀티파트 파일 업로드 기능을 제공해야 한다
- 시스템은 업로드 진행률 조회 API를 제공해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 파일 업로드가 시작되면, 시스템은 고유한 업로드 ID를 생성해야 한다
- WHEN 파일 업로드가 완료되면, 시스템은 파일 메타데이터를 데이터베이스에 저장해야 한다
- WHEN 업로드 중 오류가 발생하면, 시스템은 부분 업로드 파일을 삭제해야 한다

### State-driven Requirements (상태 기반)
- WHILE 파일이 업로드 중일 때, 시스템은 진행률을 1초마다 업데이트해야 한다
- WHILE 업로드가 일시정지 상태일 때, 시스템은 임시 파일을 유지해야 한다

### Optional Features (선택적 기능)
- WHERE 대용량 파일(100MB 이상)이면, 시스템은 청크 업로드를 사용할 수 있다
- WHERE 이미지 파일이면, 시스템은 썸네일을 자동 생성할 수 있다

### Constraints (제약사항)
- 파일 크기는 500MB를 초과할 수 없다
- IF 허용되지 않은 확장자(exe, bat, sh)이면, 시스템은 업로드를 거부해야 한다
- 파일명은 255자를 초과할 수 없다
- 동시 업로드 파일 수는 10개를 초과할 수 없다
```

## Works well with

- alfred-spec-metadata-validation
- alfred-trust-validation

## Reference

`.moai/memory/development-guide.md#ears-요구사항-작성법`
