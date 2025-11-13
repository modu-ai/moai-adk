# Agent Template Frontmatter Updates
**MoAI-ADK Agent Definition Standards**
**Version**: 1.0.0
**Date**: 2025-11-12

---

## 개요

MoAI-ADK의 29개 agent 정의 파일을 공식 Claude Code sub-agent 표준에 맞게 업데이트합니다.

**목표**:

1. Agent orchestration 메타데이터 추가
2. Resume 가능 여부 명시
3. Workflow chain에서의 위치 정의
4. Agent 간 의존성 문서화

---

## 새로운 Frontmatter 구조

### 기존 구조

```yaml
---
name: agent-name
description: "Use PROACTIVELY when..."
tools: [Read, Write, Edit, ...]
model: sonnet/haiku
---
```

### 확장된 구조 (v1.0.0)

```yaml
---
name: agent-name
description: "Use PROACTIVELY when... Called in /alfred:X-xxx Phase Y."
tools: [Read, Write, Edit, ...]
model: sonnet/haiku

# 🆕 Orchestration metadata (공식 문서 기반)
orchestration:
  can_resume: true/false
  typical_chain_position: "initial|middle|final|consultation|support"
  depends_on: [list-of-parent-agents]
  resume_pattern: "pattern_description"
  session_strategy: "new|resumable|independent"

# 🆕 Agent coordination
coordination:
  returns_to_alfred: true
  spawns_subagents: false  # Always false (공식 제약)
  requires_approval: true/false
  parallel_safe: true/false

# 🆕 Performance hints
performance:
  avg_execution_time_ms: estimated_time
  token_intensive: true/false
  cache_friendly: true/false
---
```

---

## 필드 설명

### `orchestration` 섹션

#### `can_resume: boolean`

**의미**: 이 agent가 resume 메커니즘을 활용할 수 있는가?

**결정 기준**:

- ✅ `true`: 연속 작업, 반복 수행, context 누적이 이점인 경우
- ❌ `false`: 독립 실행, 검증/분석, 상태 없는 작업

**예시**:

```yaml
# tdd-implementer: TAG 단위 연속 구현
can_resume: true

# quality-gate: 매번 독립 검증
can_resume: false
```

---

#### `typical_chain_position: string`

**의미**: Workflow chain에서 이 agent의 일반적 위치

**옵션**:

- `initial`: Workflow 시작 단계 (예: spec-builder, implementation-planner)
- `middle`: 중간 실행 단계 (예: tdd-implementer, quality-gate)
- `final`: 마지막 단계 (예: git-manager, doc-syncer)
- `consultation`: 자문 역할 (예: backend-expert, security-expert)
- `support`: 지원 도구 (예: debug-helper, mcp-integrators)

---

#### `depends_on: list[string]`

**의미**: 이 agent가 의존하는 선행 agent 목록

**규칙**:

- Alfred가 선행 agent 완료 후 이 agent 호출
- 빈 리스트 `[]`는 독립 실행 가능 의미
- 순환 의존성 금지

**예시**:

```yaml
# implementation-planner는 spec-builder 결과 필요
depends_on: ["spec-builder"]

# backend-expert는 독립 자문 가능
depends_on: []
```

---

#### `resume_pattern: string`

**의미**: 이 agent가 resume를 사용하는 전형적 패턴 설명

**예시**:

```yaml
# tdd-implementer
resume_pattern: "sequential_tag_implementation"
# TAG-001 → TAG-002 → TAG-003 연속 구현

# doc-syncer
resume_pattern: "multi_document_sync"
# product.md → structure.md → tech.md 순차 업데이트

# quality-gate
resume_pattern: "independent_validation"
# 매번 새로운 검증 실행
```

---

#### `session_strategy: string`

**의미**: Session 관리 전략

**옵션**:

- `new`: 항상 새 session 시작
- `resumable`: Resume 가능, Alfred가 결정
- `independent`: 다른 agent와 독립 실행

---

### `coordination` 섹션

#### `returns_to_alfred: boolean`

**의미**: 결과를 Alfred에게 반환하는가?

**값**: 항상 `true` (공식 문서 요구사항)

---

#### `spawns_subagents: boolean`

**의미**: 이 agent가 다른 agent를 호출하는가?

**값**: 항상 `false` (공식 문서 제약)

> "Sub-agents CANNOT spawn other sub-agents"

---

#### `requires_approval: boolean`

**의미**: 실행 전/후 사용자 승인이 필요한가?

**예시**:

```yaml
# git-manager: Commit 전 승인 필요
requires_approval: true

# doc-syncer: 자동 실행 가능
requires_approval: false
```

---

#### `parallel_safe: boolean`

**의미**: 다른 agent와 병렬 실행이 안전한가?

**결정 기준**:

- ✅ `true`: 읽기 전용, 독립 분석, 상태 없음
- ❌ `false`: 파일 수정, Git 작업, 상태 변경

**예시**:

```yaml
# backend-expert: 병렬 자문 가능
parallel_safe: true

# tdd-implementer: 파일 수정 (순차 실행 필요)
parallel_safe: false
```

---

### `performance` 섹션

#### `avg_execution_time_ms: integer`

**의미**: 평균 실행 시간 (밀리초)

**용도**: Alfred의 timeout 및 병렬 실행 결정

**예시**:

```yaml
# spec-builder: 문서 생성 (빠름)
avg_execution_time_ms: 5000

# tdd-implementer: TDD cycle (느림)
avg_execution_time_ms: 30000
```

---

#### `token_intensive: boolean`

**의미**: Token을 많이 소비하는 작업인가?

**용도**: Context budget 관리

---

#### `cache_friendly: boolean`

**의미**: Prompt caching으로 최적화 가능한가?

**용도**: 성능 최적화 힌트

---

## 29개 Agent 업데이트 매핑

### Category 1: Core Planning & Design

#### spec-builder

```yaml
---
name: spec-builder
description: "Use when: When you need to create an EARS-style SPEC document. Called from the /alfred:1-plan command."
tools: [Read, Write, Edit, MultiEdit, Bash, Glob, Grep, TodoWrite, WebFetch, AskUserQuestion, mcp__sequential_thinking_think, mcp__context7__resolve-library-id, mcp__context7__get-library-docs]
model: inherit

orchestration:
  can_resume: true
  typical_chain_position: "initial"
  depends_on: []
  resume_pattern: "multi_spec_creation"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true

performance:
  avg_execution_time_ms: 8000
  token_intensive: true
  cache_friendly: true
---
```

---

#### implementation-planner

```yaml
---
name: implementation-planner
description: "Use PROACTIVELY when detailed implementation planning is needed. Called in /alfred:1-plan Phase 2."
tools: [Read, Write, Edit, Bash, Glob, Grep, TodoWrite, WebFetch, AskUserQuestion, mcp__context7__resolve-library-id, mcp__context7__get-library-docs, mcp__sequential_thinking_think]
model: sonnet

orchestration:
  can_resume: true
  typical_chain_position: "initial"
  depends_on: ["spec-builder"]
  resume_pattern: "plan_refinement"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: true
  parallel_safe: true

performance:
  avg_execution_time_ms: 10000
  token_intensive: true
  cache_friendly: true
---
```

---

### Category 2: Core Implementation

#### tdd-implementer

```yaml
---
name: tdd-implementer
description: "Use PROACTIVELY when TDD RED-GREEN-REFACTOR implementation is needed. Called in /alfred:2-run Phase 2."
tools: [Read, Write, Edit, MultiEdit, Bash, Grep, Glob, TodoWrite, AskUserQuestion, mcp__context7__resolve-library-id, mcp__context7__get-library-docs, mcp__sequential_thinking_think]
model: haiku

orchestration:
  can_resume: true
  typical_chain_position: "middle"
  depends_on: ["implementation-planner"]
  resume_pattern: "sequential_tag_implementation"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: false  # File modifications

performance:
  avg_execution_time_ms: 35000
  token_intensive: true
  cache_friendly: false
---
```

---

#### quality-gate

```yaml
---
name: quality-gate
description: "Use PROACTIVELY when code quality validation is needed. Called in /alfred:2-run Phase 3."
tools: [Read, Bash, Grep, Glob, TodoWrite]
model: haiku

orchestration:
  can_resume: false  # Independent validation each time
  typical_chain_position: "middle"
  depends_on: ["tdd-implementer"]
  resume_pattern: "independent_validation"
  session_strategy: "new"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true  # Read-only validation

performance:
  avg_execution_time_ms: 15000
  token_intensive: false
  cache_friendly: true
---
```

---

### Category 3: Documentation & Sync

#### doc-syncer

```yaml
---
name: doc-syncer
description: "Use PROACTIVELY when documentation synchronization is needed. Called in /alfred:3-sync."
tools: [Read, Write, Edit, MultiEdit, Bash, Grep, Glob, TodoWrite]
model: haiku

orchestration:
  can_resume: true
  typical_chain_position: "final"
  depends_on: ["tdd-implementer", "quality-gate"]
  resume_pattern: "multi_document_sync"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: false  # File modifications

performance:
  avg_execution_time_ms: 12000
  token_intensive: true
  cache_friendly: true
---
```

---

#### tag-agent

```yaml
---
name: tag-agent
description: "Use PROACTIVELY when TAG validation or scanning is needed."
tools: [Read, Bash, Grep, Glob]
model: haiku

orchestration:
  can_resume: false  # Independent scan each time
  typical_chain_position: "middle"
  depends_on: []
  resume_pattern: "independent_scan"
  session_strategy: "new"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true  # Read-only scan

performance:
  avg_execution_time_ms: 5000
  token_intensive: false
  cache_friendly: true
---
```

---

### Category 4: Git & Version Control

#### git-manager

```yaml
---
name: git-manager
description: "Use PROACTIVELY when Git operations are needed. Handles commits, branches, PRs."
tools: [Read, Bash, Grep, Glob, TodoWrite, AskUserQuestion]
model: haiku

orchestration:
  can_resume: true
  typical_chain_position: "final"
  depends_on: ["quality-gate"]
  resume_pattern: "commit_cycle"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: true  # Commit confirmation
  parallel_safe: false  # Git state modification

performance:
  avg_execution_time_ms: 8000
  token_intensive: false
  cache_friendly: false
---
```

---

### Category 5: Domain Specialists

#### backend-expert

```yaml
---
name: backend-expert
description: "Use PROACTIVELY when backend architecture consultation is needed."
tools: [Read, Bash, Grep, Glob, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs, mcp__sequential_thinking_think]
model: sonnet

orchestration:
  can_resume: true
  typical_chain_position: "consultation"
  depends_on: []
  resume_pattern: "architecture_review"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true  # Consultation only

performance:
  avg_execution_time_ms: 12000
  token_intensive: true
  cache_friendly: true
---
```

---

#### frontend-expert

```yaml
---
name: frontend-expert
description: "Use PROACTIVELY when frontend/UI consultation is needed."
tools: [Read, Bash, Grep, Glob, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs]
model: sonnet

orchestration:
  can_resume: true
  typical_chain_position: "consultation"
  depends_on: []
  resume_pattern: "component_review"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true

performance:
  avg_execution_time_ms: 12000
  token_intensive: true
  cache_friendly: true
---
```

---

#### devops-expert

```yaml
---
name: devops-expert
description: "Use PROACTIVELY when DevOps/deployment consultation is needed."
tools: [Read, Bash, Grep, Glob, WebFetch, mcp__context7__get-library-docs]
model: sonnet

orchestration:
  can_resume: true
  typical_chain_position: "consultation"
  depends_on: []
  resume_pattern: "deployment_strategy"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true

performance:
  avg_execution_time_ms: 15000
  token_intensive: true
  cache_friendly: true
---
```

---

#### security-expert

```yaml
---
name: security-expert
description: "Use PROACTIVELY when security audit or consultation is needed."
tools: [Read, Bash, Grep, Glob, WebFetch]
model: sonnet

orchestration:
  can_resume: false  # Independent audit
  typical_chain_position: "consultation"
  depends_on: []
  resume_pattern: "security_audit"
  session_strategy: "independent"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true

performance:
  avg_execution_time_ms: 18000
  token_intensive: true
  cache_friendly: true
---
```

---

#### database-expert

```yaml
---
name: database-expert
description: "Use PROACTIVELY when database design consultation is needed."
tools: [Read, Bash, Grep, Glob, WebFetch, mcp__context7__get-library-docs]
model: sonnet

orchestration:
  can_resume: true
  typical_chain_position: "consultation"
  depends_on: []
  resume_pattern: "schema_design"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true

performance:
  avg_execution_time_ms: 14000
  token_intensive: true
  cache_friendly: true
---
```

---

#### ui-ux-expert

```yaml
---
name: ui-ux-expert
description: "Use PROACTIVELY when UI/UX design consultation is needed."
tools: [Read, Bash, Grep, Glob, WebFetch, mcp__context7__get-library-docs]
model: sonnet

orchestration:
  can_resume: true
  typical_chain_position: "consultation"
  depends_on: []
  resume_pattern: "design_system_review"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true

performance:
  avg_execution_time_ms: 13000
  token_intensive: true
  cache_friendly: true
---
```

---

#### performance-engineer

```yaml
---
name: performance-engineer
description: "Use PROACTIVELY when performance analysis is needed."
tools: [Read, Bash, Grep, Glob]
model: sonnet

orchestration:
  can_resume: false  # Independent analysis
  typical_chain_position: "consultation"
  depends_on: []
  resume_pattern: "performance_audit"
  session_strategy: "independent"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true

performance:
  avg_execution_time_ms: 20000
  token_intensive: true
  cache_friendly: true
---
```

---

### Category 6: Utility & Integration

#### debug-helper

```yaml
---
name: debug-helper
description: "Use PROACTIVELY when debugging or error analysis is needed."
tools: [Read, Bash, Grep, Glob, mcp__sequential_thinking_think]
model: sonnet

orchestration:
  can_resume: true
  typical_chain_position: "support"
  depends_on: []
  resume_pattern: "iterative_debugging"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true

performance:
  avg_execution_time_ms: 15000
  token_intensive: true
  cache_friendly: false
---
```

---

#### mcp-context7-integrator

```yaml
---
name: mcp-context7-integrator
description: "Use PROACTIVELY when Context7 library documentation is needed."
tools: [mcp__context7__resolve-library-id, mcp__context7__get-library-docs]
model: haiku

orchestration:
  can_resume: false
  typical_chain_position: "support"
  depends_on: []
  resume_pattern: "independent_lookup"
  session_strategy: "independent"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true

performance:
  avg_execution_time_ms: 3000
  token_intensive: false
  cache_friendly: true
---
```

---

#### mcp-playwright-integrator

```yaml
---
name: mcp-playwright-integrator
description: "Use PROACTIVELY when E2E test automation with Playwright is needed."
tools: [Read, Write, Bash, mcp__playwright__*]
model: haiku

orchestration:
  can_resume: true
  typical_chain_position: "support"
  depends_on: []
  resume_pattern: "test_scenario"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: false  # Browser interaction

performance:
  avg_execution_time_ms: 25000
  token_intensive: false
  cache_friendly: false
---
```

---

#### mcp-sequential-thinking-integrator

```yaml
---
name: mcp-sequential-thinking-integrator
description: "Use PROACTIVELY when deep analytical thinking is needed."
tools: [mcp__sequential_thinking_think]
model: sonnet

orchestration:
  can_resume: true
  typical_chain_position: "support"
  depends_on: []
  resume_pattern: "deep_analysis"
  session_strategy: "resumable"

coordination:
  returns_to_alfred: true
  spawns_subagents: false
  requires_approval: false
  parallel_safe: true

performance:
  avg_execution_time_ms: 18000
  token_intensive: true
  cache_friendly: false
---
```

---

## 나머지 Agent 목록 (간략 매핑)

### Specialists

#### accessibility-expert

```yaml
orchestration:
  can_resume: true
  typical_chain_position: "consultation"
  depends_on: []
  resume_pattern: "a11y_audit"
  session_strategy: "resumable"
```

---

#### api-designer

```yaml
orchestration:
  can_resume: true
  typical_chain_position: "consultation"
  depends_on: []
  resume_pattern: "api_design_review"
  session_strategy: "resumable"
```

---

#### component-designer

```yaml
orchestration:
  can_resume: true
  typical_chain_position: "consultation"
  depends_on: []
  resume_pattern: "component_architecture"
  session_strategy: "resumable"
```

---

#### migration-expert

```yaml
orchestration:
  can_resume: true
  typical_chain_position: "consultation"
  depends_on: []
  resume_pattern: "migration_strategy"
  session_strategy: "resumable"
```

---

#### monitoring-expert

```yaml
orchestration:
  can_resume: true
  typical_chain_position: "consultation"
  depends_on: []
  resume_pattern: "observability_setup"
  session_strategy: "resumable"
```

---

#### format-expert

```yaml
orchestration:
  can_resume: false
  typical_chain_position: "support"
  depends_on: []
  resume_pattern: "code_formatting"
  session_strategy: "independent"
```

---

### Management Agents

#### project-manager

```yaml
orchestration:
  can_resume: true
  typical_chain_position: "initial"
  depends_on: []
  resume_pattern: "project_planning"
  session_strategy: "resumable"
```

---

#### docs-manager

```yaml
orchestration:
  can_resume: true
  typical_chain_position: "final"
  depends_on: []
  resume_pattern: "documentation_workflow"
  session_strategy: "resumable"
```

---

#### cc-manager

```yaml
orchestration:
  can_resume: true
  typical_chain_position: "support"
  depends_on: []
  resume_pattern: "claude_code_config"
  session_strategy: "resumable"
```

---

#### trust-checker

```yaml
orchestration:
  can_resume: false
  typical_chain_position: "middle"
  depends_on: []
  resume_pattern: "trust_audit"
  session_strategy: "independent"
```

---

#### skill-factory

```yaml
orchestration:
  can_resume: true
  typical_chain_position: "support"
  depends_on: []
  resume_pattern: "skill_creation"
  session_strategy: "resumable"
```

---

## 업데이트 실행 계획

### Phase 1: Core Agents (우선순위 높음)

1. spec-builder
2. implementation-planner
3. tdd-implementer
4. quality-gate
5. git-manager
6. doc-syncer

**방법**:

```bash
# 각 agent 파일에 orchestration 섹션 추가
for agent in spec-builder implementation-planner tdd-implementer quality-gate git-manager doc-syncer; do
  # Edit .claude/agents/alfred/$agent.md
  # Add orchestration, coordination, performance sections
done
```

---

### Phase 2: Domain Experts

7. backend-expert
8. frontend-expert
9. devops-expert
10. security-expert
11. database-expert
12. ui-ux-expert
13. performance-engineer

---

### Phase 3: Utility & Support

14. debug-helper
15. mcp-context7-integrator
16. mcp-playwright-integrator
17. mcp-sequential-thinking-integrator
18. tag-agent
19. format-expert

---

### Phase 4: Remaining Specialists

20-29. (accessibility-expert, api-designer, component-designer, migration-expert, monitoring-expert, project-manager, docs-manager, cc-manager, trust-checker, skill-factory)

---

## 검증 체크리스트

업데이트 후 각 agent 파일에서 확인:

- [ ] `orchestration` 섹션 추가됨
- [ ] `can_resume` 값이 agent 특성에 맞게 설정됨
- [ ] `typical_chain_position` 값이 올바름
- [ ] `depends_on` 목록이 실제 의존성 반영
- [ ] `coordination.spawns_subagents` 값이 `false`
- [ ] `coordination.returns_to_alfred` 값이 `true`
- [ ] `performance` 힌트 추가됨

---

## 자동화 스크립트

### Python 스크립트로 일괄 업데이트

```python
import yaml
from pathlib import Path

def update_agent_frontmatter(agent_file: Path, orchestration_config: dict):
    """Agent 파일의 frontmatter에 orchestration 메타데이터 추가"""

    # Read file
    content = agent_file.read_text()

    # Extract frontmatter
    if not content.startswith("---"):
        print(f"❌ {agent_file.name}: No frontmatter found")
        return

    # Parse YAML frontmatter
    parts = content.split("---", 2)
    if len(parts) < 3:
        print(f"❌ {agent_file.name}: Invalid frontmatter")
        return

    frontmatter_str = parts[1]
    body = parts[2]

    frontmatter = yaml.safe_load(frontmatter_str)

    # Add orchestration metadata
    frontmatter["orchestration"] = orchestration_config["orchestration"]
    frontmatter["coordination"] = orchestration_config["coordination"]
    frontmatter["performance"] = orchestration_config["performance"]

    # Write back
    new_content = "---\n" + yaml.dump(frontmatter, allow_unicode=True, sort_keys=False) + "---" + body
    agent_file.write_text(new_content)

    print(f"✅ {agent_file.name}: Updated")

# 사용 예시
agent_configs = {
    "spec-builder": {
        "orchestration": {
            "can_resume": True,
            "typical_chain_position": "initial",
            "depends_on": [],
            "resume_pattern": "multi_spec_creation",
            "session_strategy": "resumable"
        },
        "coordination": {
            "returns_to_alfred": True,
            "spawns_subagents": False,
            "requires_approval": False,
            "parallel_safe": True
        },
        "performance": {
            "avg_execution_time_ms": 8000,
            "token_intensive": True,
            "cache_friendly": True
        }
    },
    # ... 나머지 agent 설정
}

# 실행
agents_dir = Path(".claude/agents/alfred")
for agent_name, config in agent_configs.items():
    agent_file = agents_dir / f"{agent_name}.md"
    if agent_file.exists():
        update_agent_frontmatter(agent_file, config)
```

---

## 참고 자료

- **Alfred Orchestration**: `.moai/config/alfred-orchestration.yaml`
- **Invocation Standards**: `.moai/guidelines/agent-invocation.md`
- **Official Docs**: https://code.claude.com/docs/en/sub-agents

---

**Last Updated**: 2025-11-12
**Version**: 1.0.0
