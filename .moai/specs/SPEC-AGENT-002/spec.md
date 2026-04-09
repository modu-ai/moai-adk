---
id: SPEC-AGENT-002
version: 1.0.0
status: draft
created: 2026-04-09
updated: 2026-04-09
author: GOOS
priority: high
issue_number: 0
---

# SPEC-AGENT-002: Agent Definition Optimization - Token Efficiency with Workflow Preservation

## HISTORY

- 2026-04-09: Initial draft based on Claude Code Subagents blog analysis + 26 agent audit

## Overview

MoAI-ADK의 16개 MoAI 에이전트 정의 본문을 최적화하여 토큰 소비를 90% 절감하되, 기존 워크플로우의 기능을 100% 보존한다. Claude Code 공식 Subagents 블로그의 권장 패턴을 채택하고, Agency 에이전트(33~49줄)의 간결한 패턴을 MoAI 에이전트에 적용한다.

## Environment

- Claude Code v2.1.97+
- moai-adk-go template system (internal/template/templates/)
- 16 MoAI agent definitions (.claude/agents/moai/*.md)
- 6 Agency agent definitions (.claude/agents/agency/*.md) - 참조 모델, 수정 대상 아님
- Go embed system (make build required after template changes)

## Assumptions

- Agent body content는 subagent 호출 시 전체 컨텍스트에 주입됨
- Skills는 frontmatter에 명시되면 agent 컨텍스트에 주입됨
- Rules는 paths 패턴 매칭 또는 @import로 로드됨
- maxTurns는 v2.1.69+에서 deprecated되었으며 maxContextSize로 대체 가능

## Requirements

### REQ-001: Common Agent Protocol Rule 추출 (Event-Driven)

WHEN 에이전트가 호출될 때 THEN 공통 프로토콜(Language Handling, Output Format, MCP Fallback Strategy)이 rule에서 자동 로드되어야 한다.

- 새 rule 파일 생성: `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`
- 16개 에이전트 본문에서 반복되는 Language Handling 섹션 통합
- 16개 에이전트 본문에서 반복되는 Output Format Rules 섹션 통합
- 3개 에이전트에서 반복되는 MCP Fallback Strategy 통합
- "Essential Reference" 섹션 통합 (CLAUDE.md 자동 로드 중복 제거)

### REQ-002: Agent Body 최소화 (State-Driven)

WHILE 에이전트 본문이 로드된 상태에서 THEN 본문은 역할 정의, 행동 계약, 범위, 위임 프로토콜, 핵심 워크플로우만 포함해야 한다.

각 에이전트 본문에서 제거 대상:
- "Essential Reference" 섹션 (REQ-001 rule로 이동)
- Language Handling 섹션 (REQ-001 rule로 이동)
- Output Format Rules + 예시 (REQ-001 rule로 이동)
- MCP Fallback Strategy (REQ-001 rule로 이동)
- Research Integration & Continuous Learning 섹션 (비기능적 가상 프로세스 - 삭제)
- Orchestration Metadata 블록 (Claude Code 미인식 커스텀 필드 - 삭제)
- Agent Persona "Icon:" 빈 필드 (삭제)
- Team Collaboration Patterns 코드 블록 예시 (삭제, 패턴 자체는 Delegation Protocol에 요약 유지)
- "Works Well With" 섹션 (삭제 - CLAUDE.md Agent Catalog에서 관리)
- 존재하지 않는 Rule 참조 (삭제: "Rule 1: 8-Step...", "Rule 3: Behavioral Constraints" 등)
- 존재하지 않는 에이전트/스킬 참조 (수정 또는 삭제)

각 에이전트 본문에서 보존 대상:
- Primary Mission (1-2문장)
- Behavioral Contract / SEMAP (해당 에이전트만)
- Core Capabilities (간결 목록)
- Scope Boundaries (IN/OUT)
- Delegation Protocol (위임 규칙)
- Workflow Steps (핵심 단계만 - WHY/IMPACT 주석은 축약하되 단계 자체는 100% 보존)
- Framework Detection Logic (해당 에이전트만)
- Success Criteria (간결 체크리스트)

목표: 에이전트당 평균 700줄 → 80-120줄

### REQ-003: Permission Mode 정렬 (Ubiquitous)

시스템은 항상 에이전트의 permissionMode가 실제 역할과 일치해야 한다.

| Agent | Current | Target | Reason |
|-------|---------|--------|--------|
| manager-strategy | bypassPermissions | plan | Read-Only Analysis Mode 선언됨 |
| manager-quality | bypassPermissions | plan | "never modify source code" 선언됨 |
| expert-debug | bypassPermissions | bypassPermissions | 유지 - Write/Edit tools만 제거 |
| expert-performance | bypassPermissions | bypassPermissions | 유지 - Write/Edit tools만 제거 |

### REQ-004: Deprecated maxTurns 제거 (Ubiquitous)

시스템의 모든 에이전트 정의는 deprecated된 maxTurns 필드를 사용하지 않아야 한다.

- 16개 MoAI 에이전트 + 6개 Agency 에이전트에서 maxTurns 필드 제거
- 필요 시 maxContextSize로 대체 (현재 대부분 기본값 사용 가능)

### REQ-005: Tools 최소 권한 원칙 적용 (Ubiquitous)

시스템의 모든 에이전트는 역할에 필요한 최소한의 tools만 보유해야 한다.

| Category | Tools to Remove | Affected Agents |
|----------|----------------|-----------------|
| Analysis-only | Write, Edit, Agent | manager-strategy |
| Read-only evaluator | Write, Edit, Agent, WebFetch, WebSearch | manager-quality |
| Diagnosis-only | Write, Edit | expert-debug |
| Analysis-only | Write, Edit | expert-performance |
| Implementation agents | Agent (spawns_subagents:false) | expert-backend, expert-frontend, expert-testing, expert-devops, expert-security |

### REQ-006: Hook Matcher MultiEdit 포함 (Event-Driven)

WHEN Write, Edit, 또는 MultiEdit 도구가 사용될 때 THEN 해당 에이전트의 hook이 트리거되어야 한다.

- manager-ddd, manager-tdd: tools에 MultiEdit 포함하지만 hooks matcher가 `Write|Edit`만 매칭
- 수정: matcher를 `Write|Edit|MultiEdit`로 변경

### REQ-007: Stale Reference 정리 (Ubiquitous)

시스템의 모든 에이전트 정의는 존재하지 않는 자원을 참조하지 않아야 한다.

정리 대상:
- 존재하지 않는 CLAUDE.md 규칙: "Rule 1: 8-Step User Request Analysis Process", "Rule 3: Behavioral Constraints", "Rule 5: Agent Delegation Guide (7-Tier hierarchy)", "Rule 6: Foundation Knowledge Access"
- 존재하지 않는 에이전트: support-claude, workflow-docs, expert-database, design-uiux, core-planner, mcp-context7, mcp-sequential-thinking
- 존재하지 않는 스킬: moai-core-trust-validation, moai-essentials-review, moai-essentials-perf, moai-core-tag-scanning, moai-manager-spec, moai-language-support, moai-ai-nano-banana
- 존재하지 않는 hook 이벤트: Stop (SubagentStop으로 변경 필요 여부 확인)

## Exclusions (What NOT to Build)

- Shall NOT modify Agency agent definitions (.claude/agents/agency/*.md) - 이미 33~49줄로 최적
- Shall NOT change agent naming conventions or SPEC ID format
- Shall NOT modify CLAUDE.md Section 4 (Agent Catalog) - 별도 SPEC
- Shall NOT restructure .claude/skills/ directory layout
- Shall NOT change hook handler Go implementations (internal/hook/)
- Shall NOT modify agent description multilingual keywords - 라우팅에 필요
- Shall NOT create new per-agent workflow skills - 워크플로우는 본문에 간결하게 유지
- Will NOT change model assignments (model_policy.go 관리)
