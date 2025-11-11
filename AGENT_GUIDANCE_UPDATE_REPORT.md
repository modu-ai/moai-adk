# MoAI-ADK Agent-Based Workflow Update Report (v5.0.0)

## 목표 및 배경

CLAUDE.md에 명시된 새로운 에이전트-기반 워크플로우(v5.0.0)를 실제 에이전트, 커맨드, 스킬 파일들에 적용하여 **Commands → Agents → Skills** 계층 구조를 강화하고 모든 위임 패턴을 표준화합니다.

## 핵심 변경 원칙

### 1. 에이전트-우선 원칙 (Agent-First Principle)

**규칙 #1**: 항상 에이전트에게 위임 - 절대 직접 실행 금지

```bash
# ❌ WRONG: Command doing domain work
"Design REST API for user management"

# ✅ CORRECT: Delegate to domain expert
Task(
  subagent_type="backend-expert",
  description="Design REST API for user management",
  prompt="You are the backend-expert agent. Design comprehensive user management API."
)
```

### 2. 아키텍처 강화 규칙

1. **Commands**: 오케스트레이션 ONLY - 직접 기능 구현 금지
2. **Agents**: 도메인 전문성 소유 - 복잡한 추론 및 의사결정 처리
3. **Skills**: 재사용 가능한 지식 캡슐 - 에이전트가 필요할 때 호출

## 업데이트된 파일 목록

### 핵심 스킬 파일

#### 1. `.claude/skills/moai-alfred-agent-guide/SKILL.md`

**주요 변경사항**:
- 에이전트 선택 트리에 `Task()` 호출 패턴 명시
- 새로운 "Agent Delegation Patterns (v5.0.0)" 섹션 추가
- Proper Delegation Templates, Anti-Patterns, Best Practices 명시
- 크로스-에이전트 협업 프로토콜 포함

**새로 추가된 내용**:
```markdown
**CRITICAL**: Always invoke agents via `Task(subagent_type="agent-name")` - NEVER execute directly
```

#### 2. `.claude/skills/moai-alfred-workflow/SKILL.md`

**주요 변경사항**:
- Step 2 (Plan Creation)와 Step 3 (Task Execution)에 강력한 에이전트 위임 규칙 추가
- "Forbidden Direct Execution"과 "Required Agent Delegation" 섹션 추가
- 모든 작업 실행은 반드시 에이전트를 통해 위임하도록 명시

#### 3. `.claude/skills/moai-alfred-rules/SKILL.md`

**주요 변경사항**:
- "AGENT-FIRST PRINCIPLE (v5.0.0)" 섹션을 최상단에 추가
- Architecture Enforcement Rules 명시
- 모든 규칙의 최우선 원칙으로 에이전트 위임 강조

### 에이전트 파일

#### 1. `.claude/agents/alfred/backend-expert.md`

**주요 변경사항**:
- description에 "CRITICAL: This agent MUST be invoked via Task() - NEVER executed directly" 추가
- "🚨 CRITICAL: AGENT INVOCATION RULE" 섹션 추가
- 올바른 호출 패턴과 잘못된 패턴 예시 제공

#### 2. `.claude/agents/alfred/doc-syncer.md`

**주요 변경사항**:
- description에 강력한 호출 규칙 추가
- 에이전트 호출 규칙 섹션 추가
- Commands → Agents → Skills 아키텍처 명시

#### 3. `.claude/agents/alfred/tdd-implementer.md`

**주요 변경사항**:
- description에 호출 규칙 추가
- "🚨 CRITICAL: AGENT INVOCATION RULE" 섹션 추가
- TDD 전문 에이전트로서의 역할과 호출 패턴 명시

#### 4. `.claude/agents/alfred/frontend-expert.md`

**주요 변경사항**:
- description에 호출 규칙 추가
- 에이전트 호출 규칙 섹션 추가
- 프론트엔드 도메인 전문성과 위임 패턴 명시

### 커맨드 파일

#### 1. `.claude/commands/alfred/2-run.md`

**주요 변경사항**:
- "CRITICAL: This command orchestrates ONLY - never implements directly" 명시
- Associated Skills & Agents 테이블에 "Delegation Pattern" 컬럼 추가
- Command Responsibility, Agent Responsibility, Skill Responsibility 명시

#### 2. `.claude/commands/alfred/3-sync.md`

**주요 변경사항**:
- "CRITICAL: This command orchestrates ONLY - delegates all sync work to doc-syncer agent" 명시
- Agent Delegation Pattern 섹션 추가
- 올바른 위임과 잘못된 직접 실행 예시 제공

## 적용된 핵심 패턴

### 1. 표준 에이전트 호출 템플릿

```bash
# 모든 에이전트 파일에 추가된 표준 패턴
Task(
  subagent_type="[agent-name]",
  description="[clear task description]",
  prompt="You are the [agent-name] agent. [specific instructions]"
)
```

### 2. 안티-패턴 명시

```bash
# ❌ WRONG: Direct execution
"Design backend API"
"Update documentation"
"Write tests and implementation"

# ✅ CORRECT: Agent delegation
Task(subagent_type="backend-expert", ...)
Task(subagent_type="doc-syncer", ...)
Task(subagent_type="tdd-implementer", ...)
```

### 3. 아키텍처 경계 강화

- **Commands**: 오케스트레이션만 담당, 절대 직접 구현 금지
- **Agents**: 도메인 전문성 소유, 복잡한 추론 처리
- **Skills**: 지식 캡슐화, 에이전트가 필요할 때만 호출

## 예상 효과

### 1. 이론과 실제의 일치
- CLAUDE.md의 v5.0.0 아키텍처가 실제 파일에 반영됨
- 모든 사용자가 명확한 가이드라인 따를 수 있음

### 2. 책임 소재 명확화
- 각 계층의 역할이 명확히 구분됨
- 혼란 없는 워크플로우 실행 가능

### 3. 에이전트 시스템의 체계적 활용
- 19개 에이전트 팀 구조가 효과적으로 활용됨
- 도메인 전문성이 적절히 위임됨

### 4. 일관성 있는 워크플로우 실행
- 모든 커맨드와 에이전트가 동일한 패턴 따름
- 예측 가능하고 안정적인 시스템 동작

## 추후 적용이 필요한 파일

현재 핵심 파일들이 업데이트되었으나, 남아있는 에이전트들도 동일한 패턴으로 업데이트 권장:

- `debug-helper.md`
- `implementation-planner.md`
- `git-manager.md`
- `quality-gate.md`
- `spec-builder.md`
- 기타 전문가 에이전트 (ui-ux-expert, devops-expert 등)

## 결론

MoAI-ADK 에이전트 지침 체계적 업데이트가 성공적으로 완료되었습니다. **Commands → Agents → Skills** 계층 구조가 강화되고 모든 위임 패턴이 표준화되어, 이론과 실제의 간극이 해소되었습니다.

이제 사용자들은 명확하고 일관된 가이드라인에 따라 에이전트 시스템을 체계적으로 활용할 수 있게 되었습니다.

---

**업데이트 버전**: v5.0.0
**적용 일자**: 2025-11-11
**핵심 원칙**: Agent-First, Commands-Orchestrate-Only, Clear-Ownership-Boundaries