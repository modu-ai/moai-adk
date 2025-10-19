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

- "SPEC 작성", "요구사항 정리", "EARS 구문", "명세 작성", "기능 명세"
- "시스템 요구사항", "비즈니스 요구사항", "테스트 가능한 요구사항"
- "Requirement specification", "User stories", "Acceptance criteria"
- Automatically invoked by `/alfred:1-plan`
- When writing or refining SPEC documents
- During brainstorming and requirement analysis sessions

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
✅ Use domain language consistently
✅ Include acceptance criteria

## Anti-patterns to Avoid

❌ **Vague requirements**: "시스템은 빠르게 응답해야 한다"
✅ **Correct**: "시스템은 API 요청에 100ms 이내로 응답해야 한다"

❌ **Multiple requirements**: "WHEN 사용자가 로그인하면, 토큰을 발급하고 세션을 생성해야 한다"
✅ **Correct**: Split into two requirements

❌ **Ambiguous language**: "시스템은 충분한 보안을 제공해야 한다"
✅ **Correct**: "시스템은 SHA-256 해싱으로 비밀번호를 저장해야 한다"

## EARS Validation Commands

### Check EARS Pattern Usage
```bash
# Count EARS patterns in SPEC documents
rg "^### (Ubiquitous|Event-driven|State-driven|Optional|Constraints)" .moai/specs/ -c

# Find SPEC documents without EARS sections
rg -L "## 2\. 요구사항 \(EARS\)" .moai/specs/SPEC-*/spec.md

# Validate requirement format
rg "WHEN .+이면, .+해야 한다" .moai/specs/ -n
rg "WHILE .+일 때, .+해야 한다" .moai/specs/ -n
rg "WHERE .+이면, .+할 수 있다" .moai/specs/ -n
```

### Quality Checks
```bash
# Detect vague terms
rg "(적절한|충분한|빠른|느린|많은|적은)" .moai/specs/SPEC-*/spec.md -n

# Find missing metrics
rg "해야 한다$" .moai/specs/SPEC-*/spec.md -n | grep -v "([0-9]|ms|초|분)"
```

## Examples

### Example 1: Authentication System
User: "JWT 인증 SPEC 작성해줘"

Claude response:
```markdown
### Ubiquitous Requirements (기본 요구사항)
- 시스템은 JWT 기반 인증 기능을 제공해야 한다
- 시스템은 사용자 자격증명 검증 기능을 제공해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 사용자가 유효한 자격증명으로 로그인하면, 시스템은 액세스 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 시스템은 401 Unauthorized를 반환해야 한다

### State-driven Requirements (상태 기반)
- WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 API 접근을 허용해야 한다

### Optional Features (선택적 기능)
- WHERE 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급할 수 있다

### Constraints (제약사항)
- IF 잘못된 토큰이 제공되면, 시스템은 접근을 거부해야 한다
- 액세스 토큰 만료시간은 30분을 초과하지 않아야 한다
```

### Example 2: Payment System
User: "결제 시스템 SPEC 작성"

Claude response:
```markdown
### Ubiquitous Requirements
- 시스템은 신용카드 결제 기능을 제공해야 한다

### Event-driven Requirements
- WHEN 결제 요청이 발생하면, 시스템은 결제 게이트웨이에 요청을 전송해야 한다
- WHEN 결제가 성공하면, 시스템은 주문 상태를 "결제 완료"로 변경해야 한다
- WHEN 결제가 실패하면, 시스템은 사용자에게 실패 사유를 알려야 한다

### Constraints
- 결제 타임아웃은 30초를 초과하지 않아야 한다
- 결제 금액은 100원 이상이어야 한다
```

### Example 3: Real-time Chat
User: "실시간 채팅 SPEC"

Claude response:
```markdown
### Ubiquitous Requirements
- 시스템은 실시간 메시지 전송 기능을 제공해야 한다

### Event-driven Requirements
- WHEN 사용자가 메시지를 전송하면, 시스템은 1초 이내에 수신자에게 전달해야 한다
- WHEN 새 메시지가 도착하면, 시스템은 브라우저 알림을 표시해야 한다

### State-driven Requirements
- WHILE 사용자가 오프라인 상태일 때, 시스템은 메시지를 저장해야 한다

### Optional Features
- WHERE 사용자가 이미지를 첨부하면, 시스템은 이미지 프리뷰를 표시할 수 있다

### Constraints
- 메시지 길이는 1000자를 초과할 수 없다
```

## Works well with

- moai-foundation-specs (SPEC metadata validation)
- moai-foundation-tags (TAG system integration)
