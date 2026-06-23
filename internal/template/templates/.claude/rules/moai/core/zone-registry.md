---
description: "Zone Registry — MoAI-ADK HARD 조항 SSOT. rules/agents 디렉토리 수정 시에만 로드 (registry는 moai constitution CLI가 직접 읽으므로 SPEC authoring turn에는 컨텍스트 로드 불필요)."
paths: ".claude/rules/**,.claude/agents/**"
---

# Zone Registry

MoAI-ADK 규칙 트리의 모든 HARD 조항을 열거하는 단일 진실 공급원(single source of truth).
각 엔트리에는 고유 ID, Zone 분류, 소스 파일, 앵커, verbatim clause, canary_gate 필드가 포함된다.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 1.0.0 | (initial) | maintainer | Initial creation — annotation pass over 4 load-bearing source files |
| 1.1.0 | (later)   | maintainer | Coverage gap closure — CONST-V3R5-001..039 added (parallel namespace), zone_class 4-classification introduced (retroactive on all 111 entries) |

## ID Allocation Policy

ID 형식: `CONST-V3R2-NNN` (초기 namespace) 또는 `CONST-V3R5-NNN` (parallel namespace)

할당 규칙:
- 파일 순서 고정: `CLAUDE.md` → `.claude/rules/moai/core/moai-constitution.md` → `.claude/rules/moai/core/agent-common-protocol.md` → `.claude/rules/moai/design/constitution.md`
- 각 파일 내에서 `(anchor_line_number)` 오름차순으로 ID 할당
- 001-050: pre-existing 조항 (위 4개 파일에서 발견된 HARD 조항)
- 051-099: design constitution 미러 엔트리 (§2 + §3.1/§3.2/§3.3 [FROZEN] 조항)
- 100-149: design mirror overflow (auto-extend, doctor warning 발행)
- 150+: 향후 신규 추가용 (V3R2 namespace)

V3R5 namespace 정책:
- 신규 entries 는 `CONST-V3R5-001` 부터 시작하는 parallel namespace 사용
- V3R2 의 3 internal gaps (047/048/050) 은 fill 하지 않고 역사적 기록으로 보존
- `zone_class` 필드 (4-enum): `frozen-canonical` | `frozen-safety` | `evolvable-tuning` | `evolvable-experimental`

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
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/spec-workflow.md
  anchor: "#phase-overview"
  clause: "SPEC+EARS format"
  canary_gate: true

- id: CONST-V3R2-002
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#quality-gates"
  clause: "TRUST 5"
  canary_gate: true

- id: CONST-V3R2-003
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/mx-tag-protocol.md
  anchor: "#mx-tag-types"
  clause: "@MX TAG protocol"
  canary_gate: true

- id: CONST-V3R2-004
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/development/coding-standards.md
  anchor: "#language-policy"
  clause: "16-language neutrality"
  canary_gate: true

- id: CONST-V3R2-005
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/development/coding-standards.md
  anchor: "#thin-command-pattern"
  clause: "Template-First discipline"
  canary_gate: true

- id: CONST-V3R2-006
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#user-interaction-boundary"
  clause: "AskUserQuestion monopoly"
  canary_gate: true

- id: CONST-V3R2-007
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#1-core-identity"
  clause: "Claude Code substrate"
  canary_gate: true

# ============================================================
# 008-020: CLAUDE.md HARD 조항 (§1 Hard Rules — 오케스트레이터 동작)
# ============================================================
- id: CONST-V3R2-008
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#1-hard-rules"
  clause: "Language-Aware Responses: All user-facing responses MUST be in user's conversation_language"
  canary_gate: false

- id: CONST-V3R2-009
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#1-hard-rules"
  clause: "Parallel Execution: Execute all independent tool calls in parallel when no dependencies exist"
  canary_gate: false

- id: CONST-V3R2-010
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#1-hard-rules"
  clause: "User Response Format: Use plain Markdown for all user-facing responses (XML tags are reserved for internal agent-to-agent data transfer)"
  canary_gate: false

- id: CONST-V3R2-011
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#1-hard-rules"
  clause: "Markdown Output: Use Markdown for all user-facing communication"
  canary_gate: false

- id: CONST-V3R2-012
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#8-user-interaction-architecture"
  clause: "AskUserQuestion-Only Interaction: ALL questions directed at the user MUST go through AskUserQuestion"
  canary_gate: true

- id: CONST-V3R2-013
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "Context-First Discovery: Conduct Socratic interview via AskUserQuestion when context is insufficient before executing non-trivial tasks"
  canary_gate: false

- id: CONST-V3R2-014
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "Approach-First Development: Explain approach and get approval before writing code"
  canary_gate: false

- id: CONST-V3R2-015
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "Multi-File Decomposition: Split work when modifying 3+ files"
  canary_gate: false

- id: CONST-V3R2-016
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "Post-Implementation Review: List potential issues and suggest tests after coding"
  canary_gate: false

- id: CONST-V3R2-017
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "Reproduction-First Bug Fix: Write reproduction test before fixing bugs"
  canary_gate: false

- id: CONST-V3R2-018
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#8-user-interaction-architecture"
  clause: "Every question directed at the user MUST be asked via AskUserQuestion. Free-form prose questions in regular response text are prohibited."
  canary_gate: true

- id: CONST-V3R2-019
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#8-user-interaction-architecture"
  clause: "Deferred Tool Preload Requirement: AskUserQuestion is a deferred tool — its schema is NOT loaded at session start"
  canary_gate: true

# ============================================================
# 020-030: CLAUDE.md §14 Worktree Isolation Rules + §11 Background Agent
# ============================================================
- id: CONST-V3R2-020
  zone: Evolvable
  zone_class: frozen-safety
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "As of CC v2.1.186 background subagents (run_in_background: true) surface permission prompts in the main session rather than auto-denying; MoAI keeps run_in_background: false for agents that modify files as a conservative default."
  canary_gate: false

- id: CONST-V3R2-021
  zone: Evolvable
  zone_class: evolvable-experimental
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "Implementation teammates in team mode (role_profiles: implementer, tester, designer) MUST use isolation: worktree when spawned via Agent()"
  canary_gate: false

- id: CONST-V3R2-022
  zone: Evolvable
  zone_class: evolvable-experimental
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "Read-only teammates (role_profiles: researcher, analyst, reviewer) MUST NOT use isolation: worktree"
  canary_gate: false

- id: CONST-V3R2-023
  zone: Evolvable
  zone_class: evolvable-experimental
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "One-shot sub-agents making cross-file changes SHOULD use isolation: worktree"
  canary_gate: false

- id: CONST-V3R2-024
  zone: Evolvable
  zone_class: evolvable-experimental
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "GitHub workflow fixer agents MUST use isolation: worktree for branch isolation"
  canary_gate: false

# ============================================================
# 025-035: moai-constitution.md HARD 조항
# ============================================================
- id: CONST-V3R2-025
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#moai-orchestrator"
  clause: "All user-facing questions MUST go through AskUserQuestion — no free-form prose questions in response text"
  canary_gate: true

- id: CONST-V3R2-026
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#moai-orchestrator"
  clause: "AskUserQuestion is used ONLY by MoAI orchestrator; subagents must never prompt users"
  canary_gate: true

- id: CONST-V3R2-027
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#moai-orchestrator"
  clause: "AskUserQuestion is a deferred tool — invoke ToolSearch(query: select:AskUserQuestion) immediately before each AskUserQuestion call"
  canary_gate: true

- id: CONST-V3R2-028
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#opus-47-prompt-philosophy"
  clause: "Principle 4 — Fewer subagents spawned by default: Opus 4.7+ / 4.8 does not auto-spawn subagents."
  canary_gate: false

- id: CONST-V3R2-029
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#opus-47-prompt-philosophy"
  clause: "Principle 5 — Fewer tool calls by default, more reasoning: Opus 4.7+ / 4.8 prefers reasoning over tool invocation."
  canary_gate: false

- id: CONST-V3R2-030
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Surface Assumptions: Before implementing anything non-trivial, list assumptions explicitly and wait for user confirmation."
  canary_gate: false

- id: CONST-V3R2-031
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Manage Confusion Actively: When encountering inconsistencies, STOP and surface the confusion before proceeding."
  canary_gate: false

- id: CONST-V3R2-032
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Push Back When Warranted: Point out issues directly when an approach has clear problems."
  canary_gate: false

- id: CONST-V3R2-033
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Enforce Simplicity: Actively resist overcomplexity."
  canary_gate: false

- id: CONST-V3R2-034
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Maintain Scope Discipline: Touch only what you were asked to touch."
  canary_gate: false

- id: CONST-V3R2-035
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Verify, Don't Assume: Every task requires evidence of completion."
  canary_gate: false

# ============================================================
# 036-045: agent-common-protocol.md HARD 조항
# ============================================================
- id: CONST-V3R2-036
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#user-interaction-boundary"
  clause: "Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator."
  canary_gate: true

- id: CONST-V3R2-037
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#user-interaction-boundary"
  clause: "The orchestrator MUST preload AskUserQuestion via ToolSearch(query: select:AskUserQuestion) before each call"
  canary_gate: true

- id: CONST-V3R2-038
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#user-interaction-boundary"
  clause: "All user-facing questions MUST go through AskUserQuestion — free-form prose questions in response text are prohibited"
  canary_gate: true

- id: CONST-V3R2-039
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#language-handling"
  clause: "All agents receive and respond in user's configured conversation_language."
  canary_gate: false

- id: CONST-V3R2-040
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#output-format"
  clause: "User-Facing: Always use Markdown formatting. Never display XML tags to users."
  canary_gate: false

- id: CONST-V3R2-041
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#output-format"
  clause: "Internal Agent Data: XML tags are reserved for agent-to-agent data transfer only."
  canary_gate: false

- id: CONST-V3R2-042
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#mcp-fallback-strategy"
  clause: "Maintain effectiveness without MCP servers."
  canary_gate: false

- id: CONST-V3R2-043
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#agent-invocation-pattern"
  clause: "Agents are invoked through MoAI's natural language delegation pattern."
  canary_gate: false

- id: CONST-V3R2-044
  zone: Evolvable
  zone_class: frozen-safety
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#background-agent-execution"
  clause: "Background subagents (run_in_background: true) MUST NOT perform Write/Edit operations."
  canary_gate: false

- id: CONST-V3R2-045
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#tool-usage-guidelines"
  clause: "Agents must follow tool usage patterns optimized for accuracy and efficiency."
  canary_gate: false

- id: CONST-V3R2-046
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#time-estimation"
  clause: "Never use time predictions in plans or reports."
  canary_gate: false

- id: CONST-V3R2-049
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#skeptical-evaluation-stance"
  clause: "Skeptical Evaluation Stance: reviewer operates as fresh-judgment auditor — treat claims as suspect until evidence shown, demand reproducible verification, reject when must-pass criteria fail."
  canary_gate: false

# ============================================================
# 051-099: design/constitution.md [FROZEN] 미러 엔트리 (§2 + §3.1/§3.2/§3.3)
# ============================================================
- id: CONST-V3R2-051
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] This constitution file (.claude/rules/moai/design/constitution.md)"
  canary_gate: true

- id: CONST-V3R2-052
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Section 3.1 Brand Context content"
  canary_gate: true

- id: CONST-V3R2-053
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Section 3.2 Design Brief content"
  canary_gate: true

- id: CONST-V3R2-054
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Section 3.3 Relationship rules"
  canary_gate: true

- id: CONST-V3R2-055
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Safety architecture (Section 5)"
  canary_gate: true

- id: CONST-V3R2-056
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] GAN Loop contract (Section 11)"
  canary_gate: true

- id: CONST-V3R2-057
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Evaluator leniency prevention mechanisms (Section 12)"
  canary_gate: true

- id: CONST-V3R2-058
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Pipeline phase ordering constraints (manager-spec always first, sync-auditor always last in loop)"
  canary_gate: true

- id: CONST-V3R2-059
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Pass threshold floor (minimum 0.60, cannot be lowered by evolution)"
  canary_gate: true

- id: CONST-V3R2-060
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#2-frozen-vs-evolvable-zones"
  clause: "[FROZEN] Human approval requirement for evolution (require_approval in design.yaml)"
  canary_gate: true

- id: CONST-V3R2-061
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#31-brand-context-constitutional-parent"
  clause: "[HARD] manager-spec MUST load brand context before generating BRIEF documents"
  canary_gate: true

- id: CONST-V3R2-062
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#31-brand-context-constitutional-parent"
  clause: "[HARD] moai-domain-copywriting MUST adhere to brand voice, tone, and terminology from brand-voice.md"
  canary_gate: true

- id: CONST-V3R2-063
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#31-brand-context-constitutional-parent"
  clause: "[HARD] moai-domain-brand-design MUST use brand color palette, typography, and visual language from visual-identity.md"
  canary_gate: true

- id: CONST-V3R2-064
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#31-brand-context-constitutional-parent"
  clause: "[HARD] expert-frontend MUST implement design tokens derived from brand context (archived name — resolves to Agent(general-purpose) with frontend whitelist per archived-agent-rejection.md §C; see design/constitution.md carve-out note)"
  canary_gate: true

- id: CONST-V3R2-065
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#31-brand-context-constitutional-parent"
  clause: "[HARD] sync-auditor MUST score brand consistency as a must-pass criterion"
  canary_gate: true

- id: CONST-V3R2-066
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#32-design-brief-execution-scope"
  clause: "[HARD] /moai design MUST auto-load human-authored design documents when present and not _TBD_"
  canary_gate: true

- id: CONST-V3R2-067
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#32-design-brief-execution-scope"
  clause: "[HARD] Design briefs MUST NOT override brand context — brand remains the constitutional parent"
  canary_gate: true

- id: CONST-V3R2-068
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#32-design-brief-execution-scope"
  clause: "[HARD] moai-workflow-design continues to write machine-generated artifacts to .moai/design/"
  canary_gate: true

- id: CONST-V3R2-069
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#32-design-brief-execution-scope"
  clause: "[HARD] Reserved file paths (canonical list): tokens.json, components.json, assets/, import-warnings.json, brief/BRIEF-*.md"
  canary_gate: true

- id: CONST-V3R2-070
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#32-design-brief-execution-scope"
  clause: "[HARD] Token budget for auto-loading is bounded by design_docs.token_budget; when absent, MUST default to 20000"
  canary_gate: true

- id: CONST-V3R2-071
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#32-design-brief-execution-scope"
  clause: "[HARD] Priority order when truncation is needed: spec.md > system.md > research.md"
  canary_gate: true

- id: CONST-V3R2-072
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#33-relationship"
  clause: "Brand (.moai/project/brand/) = WHO the brand is; Design (.moai/design/) = WHAT each iteration produces; brand constraints win on conflict."
  canary_gate: true

# ============================================================
# 150-159: session-handoff.md HARD 조항 (신규 워크플로우 규칙, 2026-05-04;
#          2026-05-09 model-specific threshold revision:
#          Trigger #1 = 1M context 50% / 200K context 90%; 5 triggers retained)
# ============================================================
- id: CONST-V3R2-150
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/session-handoff.md
  anchor: "#when-to-generate-5-triggers"
  clause: "[HARD] The orchestrator MUST emit a paste-ready resume message when ANY of the 5 trigger conditions activate (model-specific context threshold per context-window-management.md § Context Window Targets / SPEC phase complete / user session-end request / PR creation success with pending SPECs / multi-milestone checkpoint)"
  canary_gate: false

- id: CONST-V3R2-151
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/session-handoff.md
  anchor: "#canonical-format-verbatim-spec"
  clause: "[HARD] Resume message MUST follow the exact 6-block structure (Block 1 ultrathink + SPEC-ID + phase, Block 2 applied lessons, Block 3 separator + 전제 검증/Preconditions header, Block 4 numbered preconditions, Block 5 separator + 실행/run, Block 6 separator + 머지 후/after-merge)"
  canary_gate: false

- id: CONST-V3R2-152
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/session-handoff.md
  anchor: "#auto-memory-integration-mandatory"
  clause: "[HARD] When generating a resume message, the orchestrator MUST also persist it to a memory project entry (project_<wave>_<spec>_<status>.md), include verbatim under '## 다음 세션 시작점 (paste-ready resume message)' heading, update MEMORY.md index, and mark superseded entries with [SUPERSEDED by <new-file>] prefix"
  canary_gate: false

- id: CONST-V3R2-153
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/session-handoff.md
  anchor: "#canonical-format-verbatim-spec"
  clause: "[HARD] Resume message fenced text block MUST be bounded by cut-line markers: top marker '✂──── 여기부터 복사 ────✂' before Block 1 (or Block 0 if L3 worktree), bottom marker '✂──── 여기까지 복사 ────✂' after Block 6. ✂ symbol (U+2702 BLACK SCISSORS) and ─ (U+2500) preserved verbatim across all locales; only the marker text translates per conversation_language. Markers sit inside the fenced block alongside content so they are copied verbatim with the message, providing an unambiguous copy boundary in long terminal scrollback."
  canary_gate: false

# ============================================================
# CONST-V3R5-001..039: 신규 parallel namespace
# 미매핑 [HARD] 규칙 coverage 완성 — 11개 소스 파일 신규 등록
# ============================================================
- id: CONST-V3R5-001
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/askuser-protocol.md
  anchor: "#orchestratorsubagent-boundary"
  clause: "Subagents MUST NOT invoke AskUserQuestion"
  canary_gate: true

- id: CONST-V3R5-002
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/askuser-protocol.md
  anchor: "#orchestratorsubagent-boundary"
  clause: "Subagents MUST NOT output free-form prose questions directed at the user"
  canary_gate: true

- id: CONST-V3R5-003
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/askuser-protocol.md
  anchor: "#orchestratorsubagent-boundary"
  clause: "Subagents MUST NOT embed AskUserQuestion call syntax in their response body"
  canary_gate: true

# --- ci-autofix-protocol.md (10 entries: V3R5-004..013) ---
- id: CONST-V3R5-004
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#ci-auto-fix-loop-entry-condition"
  clause: "The CI auto-fix loop MUST be entered ONLY when scripts/ci-watch/run.sh detects a failing required check"
  canary_gate: true

- id: CONST-V3R5-005
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#iteration-limit"
  clause: "The auto-fix loop MUST attempt at most 3 iterations. The iteration counter is per-PR-push, not per-session"
  canary_gate: true

- id: CONST-V3R5-006
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#escalation-at-iteration-3"
  clause: "The AskUserQuestion at iteration > 3 MUST be a blocking call with no auto-resume timeout"
  canary_gate: true

- id: CONST-V3R5-007
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#commit-strategy"
  clause: "Every auto-fix patch MUST be applied as a new commit on the PR branch. Force-pushing or amending are prohibited"
  canary_gate: true

- id: CONST-V3R5-008
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#user-interaction-channel"
  clause: "AskUserQuestion is the exclusive user interaction channel for the auto-fix loop"
  canary_gate: true

- id: CONST-V3R5-009
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#user-interaction-channel"
  clause: "The orchestrator MUST preload AskUserQuestion via ToolSearch before each call in the auto-fix loop"
  canary_gate: true

- id: CONST-V3R5-010
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#semantic-failure-handling"
  clause: "Semantic failures (data race, deadlock, panic, test assertion failure) MUST NOT be auto-fixed without human approval"
  canary_gate: true

- id: CONST-V3R5-011
  zone: Frozen
  zone_class: frozen-safety
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#protected-files"
  clause: "The auto-fix loop MUST NOT modify .env, .env.*, credentials files, or secrets"
  canary_gate: true

- id: CONST-V3R5-012
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#audit-log"
  clause: "Every auto-fix iteration MUST be logged to .moai/logs/ci-autofix/ with timestamp, patch summary, and CI result"
  canary_gate: true

- id: CONST-V3R5-013
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#protected-files"
  clause: "The auto-fix loop MUST NOT modify scripts/ci-watch/run.sh or any Wave 2 infrastructure scripts"
  canary_gate: true

# --- ci-watch-protocol.md (8 entries: V3R5-014..021) ---
- id: CONST-V3R5-014
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-watch-protocol.md
  anchor: "#watch-loop-entry"
  clause: "The orchestrator MUST invoke the CI watch loop after /moai sync Phase 4 completes and a PR is open"
  canary_gate: true

- id: CONST-V3R5-015
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-watch-protocol.md
  anchor: "#poll-interval"
  clause: "Poll interval MUST be 30 seconds minimum. GitHub Actions API rate limits require respectful polling"
  canary_gate: true

- id: CONST-V3R5-016
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-watch-protocol.md
  anchor: "#timeout"
  clause: "The watch loop MUST exit with code 3 after 30 minutes wall-clock time if required checks have not completed"
  canary_gate: true

- id: CONST-V3R5-017
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-watch-protocol.md
  anchor: "#required-checks-ssot"
  clause: "Required checks are defined ONLY in .github/required-checks.yml. Hard-coding check names is prohibited"
  canary_gate: true

- id: CONST-V3R5-018
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-watch-protocol.md
  anchor: "#required-checks-ssot"
  clause: "Auxiliary checks listed under auxiliary: in .github/required-checks.yml MUST NOT block merge decisions"
  canary_gate: true

- id: CONST-V3R5-019
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-watch-protocol.md
  anchor: "#emit-ready-to-merge-report"
  clause: "The CLI (moai pr watch, EmitReadyToMergeReport) MUST NOT call AskUserQuestion — it emits a report and exits"
  canary_gate: true

- id: CONST-V3R5-020
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-watch-protocol.md
  anchor: "#failed-checks-reporting"
  clause: "Only required failures appear in failedChecks. Auxiliary failures are reported separately as warnings"
  canary_gate: true

- id: CONST-V3R5-021
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-watch-protocol.md
  anchor: "#protected-files"
  clause: "Wave 2 watch loop MUST NOT modify .github/required-checks.yml (Wave 1 SSoT)"
  canary_gate: true

# --- context-window-management.md (5 entries: V3R5-022..026) ---
- id: CONST-V3R5-022
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/context-window-management.md
  anchor: "#context-window-targets"
  clause: "Operational threshold is model-specific: 1M context (Opus 4.7) = 50%, 200K context (Sonnet/Opus/Haiku) = 90%"
  canary_gate: false

- id: CONST-V3R5-023
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/context-window-management.md
  anchor: "#user-responsibilities"
  clause: "When usage crosses the model-specific threshold, user MUST save in-flight state, run /clear, then paste resume message"
  canary_gate: false

- id: CONST-V3R5-024
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/context-window-management.md
  anchor: "#user-responsibilities"
  clause: "When usage crosses 95% on any model, the next action MUST be /clear — no further large work in current session"
  canary_gate: false

- id: CONST-V3R5-025
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/context-window-management.md
  anchor: "#orchestrator-responsibilities"
  clause: "Pre-clear announcement: When orchestrator detects context approaching model-specific threshold, it MUST stop large tool calls, persist progress, emit resume message, and recommend /clear"
  canary_gate: false

- id: CONST-V3R5-026
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/context-window-management.md
  anchor: "#orchestrator-responsibilities"
  clause: "Resume message format must include all fields so next session is self-sufficient"
  canary_gate: false

# --- spec-workflow.md (2 new entries: V3R5-027..028; CONST-V3R2-001 covers the third) ---
- id: CONST-V3R5-027
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/spec-workflow.md
  anchor: "#spec-phase-discipline"
  clause: "Step ordering: Step 1 (plan) MUST execute in main checkout. NO L2/L3 worktree at this step"
  canary_gate: true

- id: CONST-V3R5-028
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/spec-workflow.md
  anchor: "#spec-phase-discipline"
  clause: "Step ordering rules: Step 2 run, Step 3 sync, and Step 4 cleanup sequencing constraints"
  canary_gate: true

# --- worktree-state-guard.md (1 entry: V3R5-029) ---
- id: CONST-V3R5-029
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/worktree-state-guard.md
  anchor: "#escalation-path"
  clause: "AskUserQuestion is invoked by the orchestrator only. The Go CLI returns exit codes and JSON; the orchestrator translates these into user-facing prompts"
  canary_gate: true

# --- branch-origin-protocol.md (7 entries: V3R5-030..036) ---
- id: CONST-V3R5-030
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/development/branch-origin-protocol.md
  anchor: "#hard-rules"
  clause: "CLI path (moai worktree new) MUST NOT invoke AskUserQuestion — orchestrator-only HARD"
  canary_gate: true

- id: CONST-V3R5-031
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/development/branch-origin-protocol.md
  anchor: "#hard-rules"
  clause: "Default base for moai worktree new is origin/main (from internal/bodp.DefaultBase)"
  canary_gate: true

- id: CONST-V3R5-032
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/development/branch-origin-protocol.md
  anchor: "#hard-rules"
  clause: "--base main is the explicit opt-in for solo workflows where the user has committed locally to main without pushing"
  canary_gate: true

- id: CONST-V3R5-033
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/development/branch-origin-protocol.md
  anchor: "#hard-rules"
  clause: "--base and --from-current are mutually exclusive flags on moai worktree new"
  canary_gate: true

- id: CONST-V3R5-034
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/development/branch-origin-protocol.md
  anchor: "#hard-rules"
  clause: "Every BODP decision (skill or CLI) MUST be persisted to .moai/branches/decisions/<normalized-branch>.md via bodp.WriteDecision"
  canary_gate: true

- id: CONST-V3R5-035
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/development/branch-origin-protocol.md
  anchor: "#hard-rules"
  clause: "Skill body BODP gate MUST follow the askuser-protocol Socratic structure: (권장) first, ≤4 options, conversation_language match, Other auto-appended"
  canary_gate: true

- id: CONST-V3R5-036
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/development/branch-origin-protocol.md
  anchor: "#hard-rules"
  clause: "bodp.HasAuditTrail MUST return false when the audit directory itself is absent (fresh project)"
  canary_gate: true

# --- agent-authoring.md (1 entry: V3R5-037) ---
- id: CONST-V3R5-037
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/development/agent-authoring.md
  anchor: "#frontmatter-format-rules"
  clause: "Field format constraints: tools Comma-separated string ONLY. YAML arrays NOT supported for tools/disallowedTools"
  canary_gate: false

# --- skill-authoring.md (1 entry: V3R5-038) ---
- id: CONST-V3R5-038
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/development/skill-authoring.md
  anchor: "#key-format-rules"
  clause: "allowed-tools format: Comma-separated string ONLY. Space-separated values are PROHIBITED"
  canary_gate: false

# --- session-handoff.md supplementary (1 entry: V3R5-039; covers worktree-anchored resume) ---
- id: CONST-V3R5-039
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/session-handoff.md
  anchor: "#worktree-anchored-resume-pattern"
  clause: "When SPEC was initialized via L3 /moai plan --worktree, the resume message MUST include Block 0 (cwd anchoring) prepended before the standard 6-block structure"
  canary_gate: false

# --- glm-web-tooling.md (2 entries: V3R5-040 mandate + V3R5-041 prohibition) ---
- id: CONST-V3R5-040
  zone: Frozen
  zone_class: frozen-safety
  file: .claude/rules/moai/core/glm-web-tooling.md
  anchor: "#hard-routing-table"
  clause: "While a session is GLM-backed, MoAI agents and the orchestrator SHALL route web search to mcp__web_search_prime__webSearchPrime, web fetch to mcp__web_reader__webReader, and image reading to a mcp__zai-mcp-server__* vision tool instead of the built-in WebSearch / WebFetch / Read-on-image"
  canary_gate: true

- id: CONST-V3R5-041
  zone: Frozen
  zone_class: frozen-safety
  file: .claude/rules/moai/core/glm-web-tooling.md
  anchor: "#hard-routing-table"
  clause: "While a session is GLM-backed, MoAI agents and the orchestrator SHALL NOT invoke the built-in WebSearch or WebFetch, nor Read on an image file, because those route through the 529-prone api.z.ai/api/anthropic gateway and the base64->422 image path; the moai cg leader pane (Claude backend) is exempt"
  canary_gate: true
```
