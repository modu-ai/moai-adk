# SPEC-AGENT-002 Compact

## Requirements

- REQ-001: 공통 Agent Protocol을 rule 파일로 추출 (Language Handling, Output Format, MCP Fallback, Essential Reference)
- REQ-002: Agent Body 80-120줄로 최소화 (워크플로우 단계 100% 보존, 비기능 콘텐츠 제거)
- REQ-003: permissionMode 정렬 (strategy→plan, quality→plan, debug/perf→Write/Edit 제거)
- REQ-004: maxTurns 제거 (22개 에이전트)
- REQ-005: tools 최소 권한 원칙 적용 (분석 전용 에이전트에서 Write/Edit/Agent 제거)
- REQ-006: Hook matcher에 MultiEdit 추가 (manager-ddd, manager-tdd)
- REQ-007: Stale reference 정리 (존재하지 않는 규칙/에이전트/스킬 참조 제거)

## Files to Modify

- NEW: internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md
- MODIFY: internal/template/templates/.claude/agents/moai/expert-backend.md
- MODIFY: internal/template/templates/.claude/agents/moai/expert-frontend.md
- MODIFY: internal/template/templates/.claude/agents/moai/expert-security.md
- MODIFY: internal/template/templates/.claude/agents/moai/expert-debug.md
- MODIFY: internal/template/templates/.claude/agents/moai/expert-devops.md
- MODIFY: internal/template/templates/.claude/agents/moai/expert-performance.md
- MODIFY: internal/template/templates/.claude/agents/moai/expert-refactoring.md
- MODIFY: internal/template/templates/.claude/agents/moai/expert-testing.md
- MODIFY: internal/template/templates/.claude/agents/moai/manager-ddd.md
- MODIFY: internal/template/templates/.claude/agents/moai/manager-tdd.md
- MODIFY: internal/template/templates/.claude/agents/moai/manager-spec.md
- MODIFY: internal/template/templates/.claude/agents/moai/manager-docs.md
- MODIFY: internal/template/templates/.claude/agents/moai/manager-git.md
- MODIFY: internal/template/templates/.claude/agents/moai/manager-quality.md
- MODIFY: internal/template/templates/.claude/agents/moai/manager-project.md
- MODIFY: internal/template/templates/.claude/agents/moai/manager-strategy.md
- MODIFY: internal/template/templates/.claude/agents/moai/builder-agent.md
- MODIFY: internal/template/templates/.claude/agents/moai/builder-plugin.md
- MODIFY: internal/template/templates/.claude/agents/moai/builder-skill.md
- MODIFY: internal/template/templates/.claude/agents/moai/evaluator-active.md

## Exclusions (What NOT to Build)

- Shall NOT modify Agency agents (.claude/agents/agency/*.md)
- Shall NOT change agent naming conventions
- Shall NOT modify CLAUDE.md Section 4
- Shall NOT create new per-agent workflow skills
- Shall NOT change model assignments
- Shall NOT modify hook handler Go code

## Acceptance Criteria

- AC-001: Common rule exists and loads for all agents
- AC-002: Workflow Step count preserved 100% (before/after identical)
- AC-003: permissionMode aligned for 4 agents
- AC-004: maxTurns removed from all 22 agents
- AC-005: Average body < 120 lines, max < 200 lines
- AC-006: Hook matchers include MultiEdit
- AC-007: Zero stale references (grep verification)
- AC-008: make build + go test ./... pass
- AC-009: Delegation Protocol preserved per agent
- AC-010: Scope Boundaries preserved per agent
