---
name: "Orchestrating SPEC-First TDD Development"
description: "Guides agents through SPEC-First TDD workflow with context engineering, TRUST principles, and @TAG traceability. Essential for /alfred:1-plan, /alfred:2-run, /alfred:3-sync commands. Covers EARS requirements, JIT context loading, TDD RED-GREEN-REFACTOR cycle, and TAG chain validation."
allowed-tools: "Read, Bash(rg:*), Bash(grep:*)"
---

# Alfred Development Guide Skill

## Core Workflow: SPEC → TEST → CODE → DOC

**No spec, no code. No tests, no implementation.**

### 1. SPEC Phase (`/alfred:1-plan`)
- Author detailed specifications first with `@SPEC:ID` tags
- Use EARS format (5 patterns: Ubiquitous, Event-driven, State-driven, Optional, Constraints)
- Store in `.moai/specs/SPEC-{ID}/spec.md`

### 2. TDD Phase (`/alfred:2-run`)
- **RED**: Write failing tests with `@TEST:ID` tags
- **GREEN**: Implement minimal code with `@CODE:ID` tags
- **REFACTOR**: Improve code while maintaining SPEC compliance
- Document with `@DOC:ID` tags

### 3. SYNC Phase (`/alfred:3-sync`)
- Verify @TAG chain integrity (SPEC→TEST→CODE→DOC)
- Synchronize documentation with implementation
- Generate sync report

## Key Principles

**Context Engineering**: Load only necessary documents at each phase
- `/alfred:1-plan` → product.md, structure.md, tech.md
- `/alfred:2-run` → SPEC-{ID}/spec.md, development-guide.md
- `/alfred:3-sync` → sync-report.md, TAG validation

**TRUST 5 Pillars**:
1. **T** – Test-driven (RED→GREEN→REFACTOR)
2. **R** – Readable (clear naming, documentation)
3. **U** – Unified (consistent patterns, language)
4. **S** – Secured (OWASP compliance, security reviews)
5. **E** – Evaluated (metrics, coverage ≥85%)

## Common Patterns

| Scenario | Action |
|----------|--------|
| Write tests first | RED phase: failing tests with @TEST:ID |
| Implement feature | GREEN phase: minimal code with @CODE:ID |
| Refactor safely | REFACTOR phase: improve code structure |
| Track changes | Always use @TAG system in code + docs |
| Validate completion | `/alfred:3-sync` verifies all links |
