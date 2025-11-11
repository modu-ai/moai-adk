---
name: moai-alfred-agent-guide
version: 2.0.0
created: 2025-10-22
updated: 2025-11-11
status: active
description: 19-agent team structure, decision trees for agent selection, Haiku vs Sonnet model selection, and agent collaboration principles. Use when deciding which sub-agent to invoke, understanding team responsibilities, or learning multi-agent orchestration.
keywords: ['agent', 'team', 'orchestration', 'selection', 'collaboration']
allowed-tools:
  - Read
  - Glob
  - Grep
---

# Alfred Agent Guide

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-agent-guide |
| **Version** | 2.0.0 (2025-11-11) |
| **Allowed tools** | Read, Glob, Grep |
| **Auto-load** | On demand when agent selection needed |
| **Tier** | Alfred (Orchestration) |

---

## What It Does

MoAI-ADK의 19개 Sub-agent 아키텍처, 어떤 agent를 선택할지 결정하는 트리, Haiku/Sonnet 모델 선택 기준을 정의합니다.

**Key capabilities**:
- ✅ 19-agent team structure guide
- ✅ Agent selection decision trees
- ✅ Haiku vs Sonnet model criteria
- ✅ Agent collaboration patterns
- ✅ Multi-agent orchestration principles

---

## When to Use

- ✅ 어떤 sub-agent를 invoke할지 불명확
- ✅ Agent 책임 범위 학습
- ✅ Haiku vs Sonnet 모델 선택 필요
- ✅ Multi-agent 협업 패턴 이해

---

## Agent Team at a Glance

### 10 Core Sub-agents (Sonnet)
- **spec-builder**: SPEC 작성
- **tdd-implementer**: TDD 구현 (RED → GREEN → REFACTOR)
- **doc-syncer**: 문서 동기화
- **implementation-planner**: 구현 전략
- **debug-helper**: 오류 분석
- **quality-gate**: TRUST 5 검증
- **tag-agent**: TAG 체인 검증
- **git-manager**: Git 워크플로우
- **Explore**: 코드베이스 탐색
- **Plan**: 작업 계획

### 4 Expert Agents (Sonnet - Proactively Triggered)
- **backend-expert**: Backend 아키텍처, API 설계, 데이터베이스
- **frontend-expert**: Frontend 아키텍처, 컴포넌트 설계, 상태 관리
- **devops-expert**: DevOps 전략, 배포, 인프라
- **ui-ux-expert**: UI/UX 설계, 접근성, 디자인 시스템 (Figma MCP)

### 6 Specialist Agents (Haiku)
- **project-manager**: 프로젝트 초기화
- **skill-factory**: Skill 생성/최적화
- **cc-manager**: Claude Code 설정
- **cc-hooks**: Hook 시스템
- **cc-mcp-plugins**: MCP 서버
- **trust-checker**: TRUST 검증

---

## Agent Selection Decision Tree

```
Task Type?
├─ SPEC 작성/검증 → spec-builder
├─ TDD 구현 → tdd-implementer
├─ 문서 동기화 → doc-syncer
├─ 구현 계획 → implementation-planner
├─ 오류 분석 → debug-helper
├─ 품질 검증 → quality-gate + Skill("moai-foundation-trust")
├─ 코드베이스 탐색 → Explore
├─ Git 워크플로우 → git-manager
└─ 전체 프로젝트 계획 → Plan
```

---

## Model Selection

- **Sonnet**: Complex reasoning (spec-builder, tdd-implementer, implementation-planner)
- **Haiku**: Fast execution (project-manager, quality-gate, git-manager)

---

## Inputs

- Task type and complexity
- Domain expertise requirements
- Context scope needed
- Time sensitivity

---

## Outputs

- Selected agent recommendation
- Model choice justification
- Collaboration patterns
- Handoff procedures

---

## Dependencies

- Integration with all 19 sub-agents
- Model selection criteria
- Task classification system
- Collaboration protocols

---

## Works Well With

- `moai-alfred-workflow` (4-step process)
- `moai-foundation-trust` (quality gates)
- `moai-alfred-ask-user-questions` (clarification)

---

## References

Learn more in `reference.md` for complete agent responsibilities, collaboration patterns, and advanced orchestration strategies.

**Related Skills**: moai-alfred-rules, moai-alfred-practices

---

## Changelog

- **v2.0.0** (2025-11-11): Added complete metadata, structured content, agent selection criteria
- **v1.0.0** (2025-10-22): Initial Skill release

---

**End of Skill** | Updated 2025-11-11
