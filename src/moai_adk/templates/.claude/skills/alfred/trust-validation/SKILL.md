---
name: alfred-trust-validation
description: Validates TRUST 5-principles compliance (Test coverage 85%+, Code constraints, Architecture unity, Security, TAG trackability)
version: 0.1.0
author: MoAI Skill Factory
license: MIT
tags:
  - trust
  - quality
  - validation
  - tdd
---

# Alfred TRUST Validation

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
- File ≤300 LOC
- Function ≤50 LOC
- Parameters ≤5
- Cyclomatic complexity ≤10

**U - Unified**:
- SPEC-driven architecture consistency
- Clear module boundaries
- Language-specific standard structures

**S - Secured**:
- Input validation implementation
- No hardcoded secrets
- Access control applied

**T - Trackable**:
- TAG chain integrity (@SPEC → @TEST → @CODE → @DOC)
- No orphaned TAGs
- No duplicate SPEC IDs

## Examples

### Example 1: Quality gate check
User: "/alfred:3-sync"
Claude: (validates TRUST 5-principles and generates quality report)

### Example 2: Manual validation
User: "TRUST 원칙 준수도 확인해줘"
Claude: (scans codebase and reports compliance status)

## Works well with

- alfred-tag-scanning (TAG traceability)
- alfred-code-reviewer (code quality analysis)

## Files included

- templates/trust-report-template.md
