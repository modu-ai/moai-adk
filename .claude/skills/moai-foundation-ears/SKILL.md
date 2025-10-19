---
name: moai-foundation-ears
description: EARS requirement authoring guide (Ubiquitous/Event/State/Optional/Constraints)
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
tier: 0
auto-load: "true"
---

# Alfred EARS Authoring Guide

## What it does

EARS (Easy Approach to Requirements Syntax) authoring guide for writing clear, testable requirements using 5 statement patterns.

## When to use

- "SPEC 작성", "요구사항 정리", "EARS 구문", "요구사항 분석", "명세 작성", "기능 정의"
- "Requirements", "Specification", "EARS syntax", "Feature requirements", "User story"
- Automatically invoked by `/alfred:1-plan`
- When writing or refining SPEC documents

## How it works

EARS provides 5 statement patterns for structured requirements:

### 1. Ubiquitous (기본 요구사항)
**Format**: 시스템은 [기능]을 제공해야 한다
**Example**: 시스템은 사용자 인증 기능을 제공해야 한다

### 2. Event-driven (이벤트 기반)
**Format**: WHEN [조건]이면, 시스템은 [동작]해야 한다
**Example**: WHEN 사용자가 로그인하면, 시스템은 JWT 토큰을 발급해야 한다

### 3. State-driven (상태 기반)
**Format**: WHILE [상태]일 때, 시스템은 [동작]해야 한다
**Example**: WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 리소스 접근을 허용해야 한다

### 4. Optional (선택적 기능)
**Format**: WHERE [조건]이면, 시스템은 [동작]할 수 있다
**Example**: WHERE 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급할 수 있다

### 5. Constraints (제약사항)
**Format**: IF [조건]이면, 시스템은 [제약]해야 한다
**Example**: IF 잘못된 토큰이 제공되면, 시스템은 접근을 거부해야 한다

## Writing Tips

✅ Be specific and measurable
✅ Avoid vague terms ("적절한", "충분한", "빠른")
✅ One requirement per statement
✅ Make it testable

## How to Structure EARS in SPEC Document

```markdown
## Requirements (EARS)

### Ubiquitous Requirements (기본 요구사항)
- 시스템은 [기능]을 제공해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN [조건]이면, 시스템은 [동작]해야 한다

### State-driven Requirements (상태 기반)
- WHILE [상태]일 때, 시스템은 [동작]해야 한다

### Optional Features (선택적 기능)
- WHERE [조건]이면, 시스템은 [동작]할 수 있다

### Constraints (제약사항)
- IF [조건]이면, 시스템은 [제약]해야 한다
```

## Examples

### Example 1: Complete JWT Authentication SPEC
User: "JWT 인증 SPEC을 EARS 방식으로 작성해줘"
Claude:
```markdown
## Requirements (EARS)

### Ubiquitous Requirements
- 시스템은 JWT 기반 인증 기능을 제공해야 한다

### Event-driven Requirements
- WHEN 사용자가 유효한 자격증명(username/password)으로 로그인하면, 시스템은 액세스 토큰(JWT)을 발급해야 한다
- WHEN 토큰이 만료되면, 시스템은 401 Unauthorized 에러를 반환해야 한다
- WHEN 리프레시 토큰으로 새 액세스 토큰을 요청하면, 시스템은 새 토큰을 발급해야 한다

### State-driven Requirements
- WHILE 유효한 토큰으로 인증된 상태일 때, 시스템은 보호된 API 엔드포인트에 접근을 허용해야 한다
- WHILE 토큰이 만료된 상태일 때, 시스템은 모든 보호된 리소스 접근을 거부해야 한다

### Optional Features
- WHERE 사용자가 2FA(2-Factor Authentication)를 활성화하면, 시스템은 인증 후 추가 검증 코드를 요청할 수 있다

### Constraints
- IF 잘못된 자격증명이 제공되면, 시스템은 접근을 거부하고 403 Forbidden을 반환해야 한다
- IF 토큰에 서명이 변조되었다면, 시스템은 토큰을 무효화해야 한다
- 액세스 토큰 만료시간은 15분을 초과하지 않아야 한다
- 리프레시 토큰 만료시간은 7일을 초과하지 않아야 한다
- 암호는 bcrypt로 해싱되어 평문으로 저장되지 않아야 한다
```

### Example 2: Anti-patterns (피해야 할 패턴)

**❌ 모호한 표현**:
```
- "시스템은 적절한 속도로 응답해야 한다"
  → 문제: "적절한"의 기준이 불명확, 테스트 불가능

✅ 개선:
- "시스템은 API 요청에 200ms 이내 응답해야 한다"
```

**❌ 구현 세부사항 포함**:
```
- "시스템은 PostgreSQL을 사용하여 데이터를 저장해야 한다"
  → 문제: SPEC은 WHAT(무엇)에만 집중, HOW(어떻게)는 설계 단계에서

✅ 개선:
- "시스템은 사용자 데이터를 영구 저장해야 한다"
```

**❌ 복합 요구사항**:
```
- "시스템은 로그인하고 토큰을 발급하고 로그를 남겨야 한다"
  → 문제: 3가지 요구사항을 하나로 표현

✅ 개선 (각각 분리):
- "시스템은 로그인 기능을 제공해야 한다"
- "WHEN 로그인 성공 시, 시스템은 JWT 토큰을 발급해야 한다"
- "WHEN 로그인 시도 시, 시스템은 감사 로그를 남겨야 한다"
```

### Example 3: Payment Processing SPEC
User: "결제 시스템 SPEC을 EARS로 구성해줘"
Claude:
```markdown
## Requirements (EARS)

### Ubiquitous Requirements
- 시스템은 신용카드 기반 결제 기능을 제공해야 한다

### Event-driven Requirements
- WHEN 사용자가 결제 버튼을 클릭하면, 시스템은 결제 게이트웨이에 결제 요청을 전송해야 한다
- WHEN 결제가 성공하면, 시스템은 주문 상태를 "결제완료"로 변경해야 한다
- WHEN 결제가 실패하면, 시스템은 사용자에게 오류 메시지를 표시해야 한다

### State-driven Requirements
- WHILE 결제 처리 중일 때, 시스템은 중복 결제 요청을 거부해야 한다

### Optional Features
- WHERE 사용자가 구독 서비스를 선택하면, 시스템은 자동 갱신 설정을 활성화할 수 있다

### Constraints
- IF 결제 금액이 0 이하이면, 시스템은 결제를 진행하지 않아야 한다
- 민감한 카드 정보는 평문으로 저장되지 않아야 한다
- 모든 결제 거래는 암호화되어야 한다
```

## Writing Tips

✅ **구체성**: 측정 가능한 기준 포함
- ❌ "빠른 응답"
- ✅ "200ms 이내 응답"

✅ **테스트 가능성**: 검증 가능한 표현
- ❌ "적절한 보안"
- ✅ "bcrypt로 해싱된 암호"

✅ **한 문장 한 요구사항**:
- ❌ "로그인하고 토큰을 발급하고 로그를 남겨야 한다"
- ✅ "로그인 기능을 제공해야 한다" (별도 문장으로 분리)

✅ **도메인 용어 사용**: 팀에서 합의된 용어
- ❌ "사용자가 버튼을 누르면"
- ✅ "WHEN 결제 버튼을 클릭하면"

## Keywords

"SPEC 작성", "요구사항 정리", "EARS 구문", "요구사항 분석", "명세 작성", "기능 정의", "feature requirements"

## Reference

- EARS 상세 가이드: `.moai/memory/development-guide.md#EARS-요구사항-작성법`
- SPEC 메타데이터: `.moai/memory/spec-metadata.md`
- TDD 워크플로우: CLAUDE.md#3단계-개발-워크플로우

## Works well with

- moai-foundation-specs (SPEC 메타데이터 검증)
- moai-foundation-tags (TAG 추적성)
- moai-foundation-trust (요구사항 검증)
