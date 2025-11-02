---
name: moai-alfred-dev-guide
description: "SPEC-First TDD development workflow, 3-step cycle (SPEC → TDD → Sync), Skills index, TRUST principles, and TAG system overview. Use as the foundational development guide for MoAI-ADK workflow."
allowed-tools: "Read, Glob, Grep"
---

## What It Does

MoAI-ADK의 핵심 개발 철학: "No spec, no code. No tests, no implementation." SPEC-First TDD 3-step 사이클과 55개 Skill 전체 인덱스를 제공합니다.

## When to Use

- ✅ MoAI-ADK 개발 흐름 이해
- ✅ SPEC → TDD → Sync 3-step 사이클
- ✅ TRUST 5 원칙 학습
- ✅ TAG 시스템 개요
- ✅ 사용 가능한 Skills 전체 목록 검색

## Core Development Loop

```bash
Step 1: /alfred:1-plan "Feature name"
   → SPEC 작성 (no code without spec)

Step 2: /alfred:2-run SPEC-ID
   → TDD: RED → GREEN → REFACTOR (no implementation without tests)

Step 3: /alfred:3-sync
   → Documentation update (no completion without traceability)
```

## TRUST 5 Principles

- **Test**: 85%+ coverage
- **Readable**: SOLID principles, no code smells
- **Unified**: Consistent patterns
- **Secured**: OWASP Top 10 safe
- **Trackable**: @TAG chain SPEC→TEST→CODE→DOC

## 55 Skills by Tier

| Tier | Count | Purpose |
|------|-------|---------|
| Foundation | 6 | Core principles (TRUST, TAG, SPEC, EARS, Git, Lang) |
| Essentials | 4 | Debug, Refactor, Perf, Review |
| Alfred | 11 | Workflow orchestration |
| Domain | 10 | Backend, Frontend, API, DB, Security, DevOps, ML, Mobile, CLI, Data Science |
| Language | 23 | Python, TypeScript, Go, Rust, Java, and 18+ more |
| Ops | 1 | Claude Code session management |

---

Learn more in `reference.md` for complete workflow details, all 55 Skill descriptions, and implementation patterns.

**Related Skills**: moai-alfred-rules, moai-foundation-trust, moai-foundation-specs
