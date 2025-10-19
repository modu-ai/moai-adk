---
name: moai-foundation-ears
description: EARS requirement authoring guide (Ubiquitous/Event/State/Optional/Constraints)
allowed-tools:
- Read
- Bash
- Write
- Edit
- TodoWrite
---

# Alfred EARS Authoring Guide

## What it does

EARS (Easy Approach to Requirements Syntax) authoring guide for writing clear, testable requirements using 5 statement patterns.

## When to use

- "SPEC 작성", "요구사항 정리", "EARS 구문"
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

## Examples

User: "JWT 인증 SPEC 작성해줘"
Claude: (applies EARS patterns to structure authentication requirements)
## Works well with

- moai-foundation-specs
