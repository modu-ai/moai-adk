---
name: moai-alfred-rules
description: "Mandatory rules for Skill invocation, AskUserQuestion usage, TRUST principles, TAG validation, and TDD workflow. Use when validating workflow compliance, checking quality gates, enforcing MoAI-ADK standards, or verifying rule adherence."
allowed-tools: "Read, Glob, Grep, Bash"
---

## Skill Metadata

| Field | Value |
| ----- | ----- |
| Version | 1.0.0 |
| Tier | Alfred |
| Auto-load | When validating rules or quality gates |
| Keywords | skill-invocation, ask-user-question, trust, tag, tdd, quality-gates, workflow-compliance |

## What It Does

MoAI-ADK의 10가지 필수 Skill 호출 규칙, 5가지 AskUserQuestion 시나리오, TRUST 5 품질 게이트, TAG 체인 규칙, TDD 워크플로우를 정의합니다.

## When to Use

- ✅ Skill() 호출이 mandatory인지 optional인지 판단 필요
- ✅ 사용자 질문이 ambiguous할 때 AskUserQuestion 사용 여부 결정
- ✅ 코드/커밋이 TRUST 5를 준수하는지 확인
- ✅ TAG 체인 무결성 검증
- ✅ 커밋 메시지 형식 확인
- ✅ 품질 게이트(quality gate) 검증

## Core Rules at a Glance

### 10 Mandatory Skill Invocations

| User Request | Skill | Pattern |
|---|---|---|
| TRUST validation, quality check | `moai-foundation-trust` | `Skill("moai-foundation-trust")` |
| TAG validation, orphan detection | `moai-foundation-tags` | `Skill("moai-foundation-tags")` |
| SPEC authoring, spec validation | `moai-foundation-specs` | `Skill("moai-foundation-specs")` |
| EARS syntax, requirement formatting | `moai-foundation-ears` | `Skill("moai-foundation-ears")` |
| Git workflow, branch management | `moai-foundation-git` | `Skill("moai-foundation-git")` |
| Language detection, stack detection | `moai-foundation-langs` | `Skill("moai-foundation-langs")` |
| Debugging, error analysis | `moai-essentials-debug` | `Skill("moai-essentials-debug")` |
| Refactoring, code improvement | `moai-essentials-refactor` | `Skill("moai-essentials-refactor")` |
| Performance optimization | `moai-essentials-perf` | `Skill("moai-essentials-perf")` |
| Code review, quality review | `moai-essentials-review` | `Skill("moai-essentials-review")` |

### 5 AskUserQuestion Scenarios

Use `AskUserQuestion` when:
1. Tech stack choice unclear (multiple frameworks/languages)
2. Architecture decision needed (monolith vs microservices)
3. User intent ambiguous (multiple valid interpretations)
4. Existing component impacts unknown (breaking changes)
5. Resource constraints unclear (budget, timeline)

### TRUST 5 Quality Gates

- **Test**: 85%+ coverage required
- **Readable**: No code smells, SOLID principles
- **Unified**: Consistent patterns, no duplicate logic
- **Secured**: OWASP Top 10 checks, no secrets
- **Trackable**: @TAG chain intact (SPEC→TEST→CODE→DOC)

### TAG Chain Integrity Rules

- Assign as `<DOMAIN>-<###>` (e.g., `AUTH-003`)
- Create `@TEST` before `@CODE`
- Document in HISTORY section
- Never have orphan TAGs (TAG without corresponding code)

## Progressive Disclosure

Learn more in `reference.md` for complete rules, decision trees, and validation methods.

---

**Version**: 1.0.0
**Related Skills**: moai-foundation-trust, moai-foundation-tags, moai-alfred-practices