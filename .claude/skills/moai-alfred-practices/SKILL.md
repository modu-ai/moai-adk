---
name: moai-alfred-practices
description: "Practical workflows, context engineering strategies, and real-world execution examples for MoAI-ADK. Use when learning workflow patterns, optimizing context management, debugging issues, or implementing features end-to-end."
allowed-tools: "Read, Glob, Grep, Bash"
---

## Skill Metadata

| Field | Value |
| ----- | ----- |
| Version | 1.0.0 |
| Tier | Alfred |
| Auto-load | When practical guidance is needed |
| Keywords | workflow-examples, context-engineering, jit-retrieval, agent-usage, debugging-patterns, feature-implementation |

## What It Does

JIT (Just-in-Time) context 관리 전략, Explore agent 효율적 사용법, SPEC → TDD → Sync 실행 순서, 자주 발생하는 문제 해결책을 제공합니다.

## When to Use

- ✅ 실제 task 실행 시 구체적 단계 필요
- ✅ Context 관리 최적화 필요 (큰 프로젝트)
- ✅ Explore agent 효율적 활용 방법 학습
- ✅ SPEC → TDD → Sync 실행 패턴 학습
- ✅ 자주 발생하는 문제 해결 방법 찾기

## Core Practices at a Glance

### 1. Context Engineering Strategy

#### JIT (Just-in-Time) Retrieval
- 필요한 context만 즉시 pull
- Explore로 manual file hunting 대체
- Task thread에서 결과 cache하여 재사용

#### Efficient Use of Explore
- Call graphs/dependency maps for core module changes
- Similar features 검색으로 구현 참고
- SPEC references나 TAG metadata로 변경사항 anchor

### 2. Context Layering

```
High-level brief      → Purpose, stakeholders, success criteria
        ↓
Technical core        → Entry points, domain models, utilities
        ↓
Edge cases           → Known bugs, constraints, SLAs
```

### 3. Practical Workflow Commands

```bash
/alfred:1-plan "Feature name"
  → Skill("moai-alfred-spec-metadata-extended") validation
  → SPEC 생성

/alfred:2-run SPEC-ID
  → TDD RED → GREEN → REFACTOR
  → Tests + Implementation

/alfred:3-sync
  → Documentation auto-update
  → TAG chain validation
  → PR ready
```

## 5 Practical Scenarios

1. **Feature Implementation**: New feature from SPEC to production
2. **Debugging & Triage**: Error analysis with fix-forward recommendations
3. **TAG System Management**: ID assignment, HISTORY updates
4. **Backup Management**: Automatic safety snapshots before risky actions
5. **Multi-Agent Collaboration**: Coordinate between debug-helper, spec-builder, tdd-implementer

## Key Principles

- ✅ **Context minimization**: Load only what's needed now
- ✅ **Explore-first**: Use Explore agent for large searches
- ✅ **Living documentation**: Sync after significant changes
- ✅ **Problem diagnosis**: Use debug-helper for error triage
- ✅ **Reproducibility**: Record rationale for SPEC deviations

---

**Learn More**: See `reference.md` for step-by-step examples, full workflow sequences, and advanced patterns.

**Related Skills**: moai-alfred-rules, moai-alfred-agent-guide, moai-essentials-debug
