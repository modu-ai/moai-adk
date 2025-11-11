---
name: moai-alfred-agent-guide
version: 2.0.0
created: 2025-10-01
updated: 2025-11-11
status: active
description: "19-agent team structure, decision trees for agent selection, Haiku vs Sonnet model selection, and agent collaboration principles. Enhanced with research capabilities for agent performance analysis and optimization. Use when deciding which sub-agent to invoke, understanding team responsibilities, or learning multi-agent orchestration."
allowed-tools: "Read, Glob, Grep, TodoWrite"
tags: [agent, coordination, decision-tree, research, analysis, optimization, team-management, performance]
---

## What It Does

MoAI-ADK의 19개 Sub-agent 아키텍처, 어떤 agent를 선택할지 결정하는 트리, Haiku/Sonnet 모델 선택 기준을 정의합니다.

## When to Use

- ✅ 어떤 sub-agent를 invoke할지 불명확
- ✅ Agent 책임 범위 학습
- ✅ Haiku vs Sonnet 모델 선택 필요
- ✅ Multi-agent 협업 패턴 이해

## Agent Team at a Glance

### 10 Core Sub-agents (Sonnet)
- spec-builder: SPEC 작성
- tdd-implementer: TDD 구현 (RED → GREEN → REFACTOR)
- doc-syncer: 문서 동기화
- implementation-planner: 구현 전략
- debug-helper: 오류 분석
- quality-gate: TRUST 5 검증
- tag-agent: TAG 체인 검증
- git-manager: Git 워크플로우
- Explore: 코드베이스 탐색
- Plan: 작업 계획

### 4 Expert Agents (Sonnet - Proactively Triggered)
- **backend-expert**: Backend 아키텍처, API 설계, 데이터베이스
- **frontend-expert**: Frontend 아키텍처, 컴포넌트 설계, 상태 관리
- **devops-expert**: DevOps 전략, 배포, 인프라
- **ui-ux-expert**: UI/UX 설계, 접근성, 디자인 시스템 (Figma MCP)

### 6 Specialist Agents (Haiku)
- project-manager: 프로젝트 초기화
- skill-factory: Skill 생성/최적화
- cc-manager: Claude Code 설정
- cc-hooks: Hook 시스템
- cc-mcp-plugins: MCP 서버
- trust-checker: TRUST 검증

## Agent Selection Decision Tree

**CRITICAL**: Always invoke agents via `Task(subagent_type="agent-name")` - NEVER execute directly

```
Task Type?
├─ SPEC 작성/검증 → Task(subagent_type="spec-builder")
├─ TDD 구현 → Task(subagent_type="tdd-implementer")
├─ 문서 동기화 → Task(subagent_type="doc-syncer")
├─ 구현 계획 → Task(subagent_type="implementation-planner")
├─ 오류 분석 → Task(subagent_type="debug-helper")
├─ 품질 검증 → Task(subagent_type="quality-gate") + Skill("moai-foundation-trust")
├─ 코드베이스 탐색 → Task(subagent_type="Explore")
├─ Git 워크플로우 → Task(subagent_type="git-manager")
└─ 전체 프로젝트 계획 → Task(subagent_type="Plan")
```

## Model Selection

- **Sonnet**: Complex reasoning (spec-builder, tdd-implementer, implementation-planner)
- **Haiku**: Fast execution (project-manager, quality-gate, git-manager)

---

## Agent Delegation Patterns (v5.0.0)

### Commands → Agents → Skills Architecture

**CRITICAL RULES**:
1. **Commands NEVER execute directly** - Always orchestrate via agents
2. **Agents own domain expertise** - Handle complex reasoning and decisions
3. **Skills provide reusable knowledge** - Called by agents when needed

### Proper Delegation Templates

#### For Commands (Orchestration Only):
```bash
# ❌ WRONG: Direct execution
"Implement SPEC-001"

# ✅ CORRECT: Agent delegation
Task(
  subagent_type="tdd-implementer",
  description="Execute TDD implementation for SPEC-001",
  prompt="You are the tdd-implementer agent. Execute SPEC-001 using TDD cycle."
)
```

#### For Agents (Domain Execution):
```bash
# ❌ WRONG: Direct skill execution without context
Skill("moai-domain-backend")

# ✅ CORRECT: Skill loading with proper context
Skill("moai-domain-backend")  # Load domain knowledge
# Then apply to specific task with context
```

#### For Specialist Agent Activation:
```bash
# ❌ WRONG: Manual domain work
"Design backend API for user authentication"

# ✅ CORRECT: Delegate to domain expert
Task(
  subagent_type="backend-expert",
  description="Design and implement backend authentication system",
  prompt="You are the backend-expert agent. Design comprehensive authentication API."
)
```

### Agent Collaboration Protocols

#### Cross-Agent Coordination:
```bash
# Backend expert coordinates with frontend expert
Task(
  subagent_type="backend-expert",
  description="Create API contract for frontend integration",
  prompt="Coordinate with frontend-expert for API contract. Design endpoints that frontend can consume."
)
```

#### Sequential Agent Handoffs:
```bash
# 1. Plan agent creates strategy
Task(subagent_type="Plan", ...)
# 2. Implementation agent executes
Task(subagent_type="tdd-implementer", ...)
# 3. Quality agent validates
Task(subagent_type="quality-gate", ...)
```

### Anti-Patterns to Avoid

❌ **Direct Command Execution**: Commands implementing features directly
❌ **Agent Bypassing**: Using skills without proper agent context
❌ **Mixed Responsibilities**: Commands doing both orchestration AND implementation
❌ **Unclear Delegation**: Ambiguous handoffs between agents

### Best Practices

✅ **Clear Ownership**: Each task has one responsible agent
✅ **Proper Handoffs**: Explicit agent-to-agent communication
✅ **Skill Context**: Skills loaded within agent domain context
✅ **Traceable Work**: Every action traceable to responsible agent

---

Learn more in `reference.md` for complete agent responsibilities, collaboration patterns, and advanced orchestration strategies.

**Related Skills**: moai-alfred-rules, moai-alfred-practices
