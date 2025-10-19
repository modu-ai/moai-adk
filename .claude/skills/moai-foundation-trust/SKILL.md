---
name: moai-foundation-trust
description: Validates TRUST 5-principles (Test 85%+, Readable, Unified, Secured,
  Trackable)
allowed-tools:
- Read
- Write
- Edit
- Bash
- TodoWrite
---

# Foundation: TRUST Validation

## What it does

Validates MoAI-ADK's TRUST 5-principles compliance to ensure code quality, testability, security, and traceability.

## When to use

- "TRUST 원칙 확인", "품질 검증", "코드 품질 체크"
- Automatically invoked by `/alfred:3-sync`
- Before merging PR or releasing

## How it works

**T - Test First**:
- Checks test coverage ≥85% (pytest, vitest, go test, cargo test, etc.)
- Verifies TDD cycle compliance (RED → GREEN → REFACTOR)

**R - Readable**:
- File ≤300 LOC, Function ≤50 LOC, Parameters ≤5, Complexity ≤10

**U - Unified**:
- SPEC-driven architecture consistency, Clear module boundaries

**S - Secured**:
- Input validation, No hardcoded secrets, Access control

**T - Trackable**:
- TAG chain integrity (@SPEC → @TEST → @CODE → @DOC)

## Works well with

- moai-foundation-tags (TAG traceability)
- moai-foundation-specs (SPEC validation)
