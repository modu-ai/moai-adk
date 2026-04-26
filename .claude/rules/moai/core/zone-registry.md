# Zone Registry

MoAI-ADK 규칙 트리의 모든 HARD 조항을 열거하는 단일 진실 공급원(single source of truth).
각 엔트리에는 고유 ID, Zone 분류, 소스 파일, 앵커, verbatim clause, canary_gate 필드가 포함된다.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 1.0.0 | 2026-04-25 | SPEC-V3R2-CON-001 T-03 | 초기 생성 — 4개 load-bearing 파일 annotation pass 완료 |

## ID Allocation Policy

ID 형식: `CONST-V3R2-NNN` (zero-padded 3자리 이상)

할당 규칙 (SPEC-V3R2-CON-001 §7 Constraints + plan.md §7 OQ5 결정):
- 파일 순서 고정: `CLAUDE.md` → `.claude/rules/moai/core/moai-constitution.md` → `.claude/rules/moai/core/agent-common-protocol.md` → `.claude/rules/moai/design/constitution.md`
- 각 파일 내에서 `(anchor_line_number)` 오름차순으로 ID 할당
- 001-050: pre-existing 조항 (위 4개 파일에서 발견된 HARD 조항)
- 051-099: design constitution 미러 엔트리 (§2 + §3.1/§3.2/§3.3 [FROZEN] 조항)
- 100-149: design mirror overflow (auto-extend, doctor warning 발행)
- 150+: 향후 신규 추가용

CanaryGate 기본값 (plan.md §7 OQ6 결정):
- Frozen → `canary_gate: true`
- Evolvable → `canary_gate: false`

## Usage Guide

```bash
# 전체 registry 조회
moai constitution list

# Frozen zone 필터
moai constitution list --zone frozen

# 특정 파일의 조항만 조회
moai constitution list --file .claude/rules/moai/core/moai-constitution.md

# JSON 형식 출력
moai constitution list --format json
```

## Entries

```yaml
# ============================================================
# 001-010: CLAUDE.md HARD 조항 (§1 Hard Rules)
# ============================================================
- id: CONST-V3R2-001
  zone: Frozen
  file: .claude/rules/moai/workflow/spec-workflow.md
  anchor: "#phase-overview"
  clause: "SPEC+EARS format"
  canary_gate: true

- id: CONST-V3R2-002
  zone: Frozen
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#quality-gates"
  clause: "TRUST 5"
  canary_gate: true

- id: CONST-V3R2-003
  zone: Frozen
  file: .claude/rules/moai/workflow/mx-tag-protocol.md
  anchor: "#mx-tag-types"
  clause: "@MX TAG protocol"
  canary_gate: true

- id: CONST-V3R2-004
  zone: Frozen
  file: .claude/rules/moai/development/coding-standards.md
  anchor: "#language-policy"
  clause: "16-language neutrality"
  canary_gate: true

- id: CONST-V3R2-005
  zone: Frozen
  file: .claude/rules/moai/development/coding-standards.md
  anchor: "#thin-command-pattern"
  clause: "Template-First discipline"
  canary_gate: true

- id: CONST-V3R2-006
  zone: Frozen
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#user-interaction-boundary"
  clause: "AskUserQuestion monopoly"
  canary_gate: true

- id: CONST-V3R2-007
  zone: Frozen
  file: CLAUDE.md
  anchor: "#1-core-identity"
  clause: "Claude Code substrate"
  canary_gate: true

# ============================================================
# 008-020: CLAUDE.md HARD 조항 (§1 Hard Rules — 오케스트레이터 동작)
# ============================================================
- id: CONST-V3R2-008
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#1-hard-rules"
  clause: "Language-Aware Responses: All user-facing responses MUST be in user's conversation_language"
  canary_gate: false

- id: CONST-V3R2-009
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#1-hard-rules"
  clause: "Parallel Execution: Execute all independent tool calls in parallel when no dependencies exist"
  canary_gate: false

- id: CONST-V3R2-010
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#1-hard-rules"
  clause: "No XML in User Responses: Never display XML tags in user-facing responses"
  canary_gate: false

- id: CONST-V3R2-011
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#1-hard-rules"
  clause: "Markdown Output: Use Markdown for all user-facing communication"
  canary_gate: false

- id: CONST-V3R2-012
  zone: Frozen
  file: CLAUDE.md
  anchor: "#8-user-interaction-architecture"
  clause: "AskUserQuestion-Only Interaction: ALL questions directed at the user MUST go through AskUserQuestion"
  canary_gate: true

- id: CONST-V3R2-013
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "Context-First Discovery: Conduct Socratic interview via AskUserQuestion when context is insufficient before executing non-trivial tasks"
  canary_gate: false

- id: CONST-V3R2-014
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "Approach-First Development: Explain approach and get approval before writing code"
  canary_gate: false

- id: CONST-V3R2-015
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "Multi-File Decomposition: Split work when modifying 3+ files"
  canary_gate: false

- id: CONST-V3R2-016
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "Post-Implementation Review: List potential issues and suggest tests after coding"
  canary_gate: false

- id: CONST-V3R2-017
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "Reproduction-First Bug Fix: Write reproduction test before fixing bugs"
  canary_gate: false

- id: CONST-V3R2-018
  zone: Frozen
  file: CLAUDE.md
  anchor: "#8-user-interaction-architecture"
  clause: "Every question directed at the user MUST be asked via AskUserQuestion. Free-form prose questions in regular response text are prohibited."
  canary_gate: true

- id: CONST-V3R2-019
  zone: Frozen
  file: CLAUDE.md
  anchor: "#8-user-interaction-architecture"
  clause: "Deferred Tool Preload Requirement: AskUserQuestion is a deferred tool — its schema is NOT loaded at session start"
  canary_gate: true

# ============================================================
# 020-030: CLAUDE.md §14 Worktree Isolation Rules + §11 Background Agent
# ============================================================
- id: CONST-V3R2-020
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "Background subagents (run_in_background: true) auto-deny Write/Edit operations. Use run_in_background: false for agents that modify files."
  canary_gate: false

- id: CONST-V3R2-021
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "Implementation teammates in team mode (role_profiles: implementer, tester, designer) MUST use isolation: worktree when spawned via Agent()"
  canary_gate: false

- id: CONST-V3R2-022
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "Read-only teammates (role_profiles: researcher, analyst, reviewer) MUST NOT use isolation: worktree"
  canary_gate: false

- id: CONST-V3R2-023
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "One-shot sub-agents making cross-file changes SHOULD use isolation: worktree"
  canary_gate: false

- id: CONST-V3R2-024
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "GitHub workflow fixer agents MUST use isolation: worktree for branch isolation"
  canary_gate: false

# ============================================================
# 025-035: moai-constitution.md HARD 조항
# ============================================================
- id: CONST-V3R2-025
  zone: Frozen
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#moai-orchestrator"
  clause: "All user-facing questions MUST go through AskUserQuestion — no free-form prose questions in response text"
  canary_gate: true

- id: CONST-V3R2-026
  zone: Frozen
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#moai-orchestrator"
  clause: "AskUserQuestion is used ONLY by MoAI orchestrator; subagents must never prompt users"
  canary_gate: true

- id: CONST-V3R2-027
  zone: Frozen
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#moai-orchestrator"
  clause: "AskUserQuestion is a deferred tool — invoke ToolSearch(query: select:AskUserQuestion) immediately before each AskUserQuestion call"
  canary_gate: true

- id: CONST-V3R2-028
  zone: Evolvable
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#opus-47-prompt-philosophy"
  clause: "Principle 4 — Fewer subagents spawned by default: Opus 4.7 does not auto-spawn subagents."
  canary_gate: false

- id: CONST-V3R2-029
  zone: Evolvable
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#opus-47-prompt-philosophy"
  clause: "Principle 5 — Fewer tool calls by default, more reasoning: Opus 4.7 prefers reasoning over tool invocation."
  canary_gate: false

- id: CONST-V3R2-030
  zone: Evolvable
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Surface Assumptions: Before implementing anything non-trivial, list assumptions explicitly and wait for user confirmation."
  canary_gate: false

- id: CONST-V3R2-031
  zone: Evolvable
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Manage Confusion Actively: When encountering inconsistencies, STOP and surface the confusion before proceeding."
  canary_gate: false

- id: CONST-V3R2-032
  zone: Evolvable
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Push Back When Warranted: Point out issues directly when an approach has clear problems."
  canary_gate: false

- id: CONST-V3R2-033
  zone: Evolvable
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Enforce Simplicity: Actively resist overcomplexity."
  canary_gate: false

- id: CONST-V3R2-034
  zone: Evolvable
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Maintain Scope Discipline: Touch only what you were asked to touch."
  canary_gate: false

- id: CONST-V3R2-035
  zone: Evolvable
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Verify, Don't Assume: Every task requires evidence of completion."
  canary_gate: false

# ============================================================
# 036-045: agent-common-protocol.md HARD 조항
# ============================================================
- id: CONST-V3R2-036
  zone: Frozen
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#user-interaction-boundary"
  clause: "Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator."
  canary_gate: true

- id: CONST-V3R2-037
  zone: Frozen
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#user-interaction-boundary"
  clause: "The orchestrator MUST preload AskUserQuestion via ToolSearch(query: select:AskUserQuestion) before each call"
  canary_gate: true

- id: CONST-V3R2-038
  zone: Frozen
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#user-interaction-boundary"
  clause: "All user-facing questions MUST go through AskUserQuestion — free-form prose questions in response text are prohibited"
  canary_gate: true

- id: CONST-V3R2-039
  zone: Evolvable
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#language-handling"
  clause: "All agents receive and respond in user's configured conversation_language."
  canary_gate: false

- id: CONST-V3R2-040
  zone: Evolvable
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#output-format"
  clause: "User-Facing: Always use Markdown formatting. Never display XML tags to users."
  canary_gate: false

- id: CONST-V3R2-041
  zone: Evolvable
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#output-format"
  clause: "Internal Agent Data: XML tags are reserved for agent-to-agent data transfer only."
  canary_gate: false

- id: CONST-V3R2-042
  zone: Evolvable
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#mcp-fallback-strategy"
  clause: "Maintain effectiveness without MCP servers."
  canary_gate: false

- id: CONST-V3R2-043
  zone: Evolvable
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#agent-invocation-pattern"
  clause: "Agents are invoked through MoAI's natural language delegation pattern."
  canary_gate: false

- id: CONST-V3R2-044
  zone: Evolvable
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#background-agent-execution"
  clause: "Background subagents (run_in_background: true) MUST NOT perform Write/Edit operations."
  canary_gate: false

- id: CONST-V3R2-045
  zone: Evolvable
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#tool-usage-guidelines"
  clause: "Agents must follow tool usage patterns optimized for accuracy and efficiency."
  canary_gate: false

- id: CONST-V3R2-046
  zone: Evolvable
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#time-estimation"
  clause: "Never use time predictions in plans or reports."
  canary_gate: false

# ============================================================
# 051-099: design/constitution.md [FROZEN] 미러 엔트리 (§2 + §3.1/§3.2/§3.3)
# ============================================================
- id: CONST-V3R2-051
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] This constitution file (.claude/rules/moai/design/constitution.md)"
  canary_gate: true

- id: CONST-V3R2-052
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Section 3.1 Brand Context content"
  canary_gate: true

- id: CONST-V3R2-053
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Section 3.2 Design Brief content"
  canary_gate: true

- id: CONST-V3R2-054
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Section 3.3 Relationship rules"
  canary_gate: true

- id: CONST-V3R2-055
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Safety architecture (Section 5)"
  canary_gate: true

- id: CONST-V3R2-056
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] GAN Loop contract (Section 11)"
  canary_gate: true

- id: CONST-V3R2-057
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Evaluator leniency prevention mechanisms (Section 12)"
  canary_gate: true

- id: CONST-V3R2-058
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Pipeline phase ordering constraints (manager-spec always first, evaluator-active always last in loop)"
  canary_gate: true

- id: CONST-V3R2-059
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Pass threshold floor (minimum 0.60, cannot be lowered by evolution)"
  canary_gate: true

- id: CONST-V3R2-060
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Human approval requirement for evolution (require_approval in design.yaml)"
  canary_gate: true

- id: CONST-V3R2-061
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#31-brand-context-constitutional-parent"
  clause: "[HARD] manager-spec MUST load brand context before generating BRIEF documents"
  canary_gate: true

- id: CONST-V3R2-062
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#31-brand-context-constitutional-parent"
  clause: "[HARD] moai-domain-copywriting MUST adhere to brand voice, tone, and terminology from brand-voice.md"
  canary_gate: true

- id: CONST-V3R2-063
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#31-brand-context-constitutional-parent"
  clause: "[HARD] moai-domain-brand-design MUST use brand color palette, typography, and visual language from visual-identity.md"
  canary_gate: true

- id: CONST-V3R2-064
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#31-brand-context-constitutional-parent"
  clause: "[HARD] expert-frontend MUST implement design tokens derived from brand context"
  canary_gate: true

- id: CONST-V3R2-065
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#31-brand-context-constitutional-parent"
  clause: "[HARD] evaluator-active MUST score brand consistency as a must-pass criterion"
  canary_gate: true

- id: CONST-V3R2-066
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#32-design-brief-execution-scope"
  clause: "[HARD] /moai design MUST auto-load human-authored design documents when present and not _TBD_"
  canary_gate: true

- id: CONST-V3R2-067
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#32-design-brief-execution-scope"
  clause: "[HARD] Design briefs MUST NOT override brand context — brand remains the constitutional parent"
  canary_gate: true

- id: CONST-V3R2-068
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#32-design-brief-execution-scope"
  clause: "[HARD] moai-workflow-design-import continues to write machine-generated artifacts to .moai/design/"
  canary_gate: true

- id: CONST-V3R2-069
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#32-design-brief-execution-scope"
  clause: "[HARD] Reserved file paths (canonical list): tokens.json, components.json, assets/, import-warnings.json, brief/BRIEF-*.md"
  canary_gate: true

- id: CONST-V3R2-070
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#32-design-brief-execution-scope"
  clause: "[HARD] Token budget for auto-loading is bounded by design_docs.token_budget; when absent, MUST default to 20000"
  canary_gate: true

- id: CONST-V3R2-071
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#32-design-brief-execution-scope"
  clause: "[HARD] Priority order when truncation is needed: spec.md > system.md > research.md > pencil-plan.md"
  canary_gate: true

- id: CONST-V3R2-072
  zone: Frozen
  file: .claude/rules/moai/design/constitution.md
  anchor: "#33-relationship"
  clause: "Brand (.moai/project/brand/) = WHO the brand is; Design (.moai/design/) = WHAT each iteration produces; brand constraints win on conflict."
  canary_gate: true
```
