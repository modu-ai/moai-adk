---
description: "Constitution zone registry — CONST-* clause records consumed by moai constitution CLI and zone audits"
paths: "**/zone-registry.md,**/.moai/config/sections/constitution.yaml"
---

# Zone Registry

Single source of truth enumerating every HARD clause in the MoAI-ADK rules tree.
Each entry carries a unique ID, Zone classification, source file, anchor, verbatim clause, and canary_gate field.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 1.0.0 | (initial) | maintainer | Initial creation — annotation pass over 4 load-bearing source files |
| 1.1.0 | (later)   | maintainer | Coverage gap closure — CONST-V3R5-001..041 added (parallel namespace), zone_class 4-classification introduced (retroactive on all 115 entries) |

## ID Allocation Policy

ID format: `CONST-V3R2-NNN` (initial namespace) or `CONST-V3R5-NNN` (parallel namespace)

Allocation rules:
- Fixed file order: `CLAUDE.md` → `.claude/rules/moai/core/moai-constitution.md` → `.claude/rules/moai/core/agent-common-protocol.md` → `.claude/rules/moai/design/constitution.md`
- Within each file, assign IDs in ascending `(anchor_line_number)` order
- 001-050: pre-existing clauses (HARD clauses found in the 4 files above)
- 051-099: design constitution mirror entries (§2 + §3.1/§3.2/§3.3 [FROZEN] clauses)
- 100-149: design mirror overflow (auto-extend, emits doctor warning)
- 150+: reserved for future additions (V3R2 namespace)

V3R5 namespace policy:
- New entries use the parallel namespace starting at `CONST-V3R5-001`
- The 3 internal V3R2 gaps (047/048/050) are NOT filled — preserved as historical record
- `zone_class` field (4-enum): `frozen-canonical` | `frozen-safety` | `evolvable-tuning` | `evolvable-experimental`

CanaryGate defaults (plan.md §7 OQ6 decision):
- Frozen → `canary_gate: true`
- Evolvable → `canary_gate: false`

## Retiring an Entry

A clause that is no longer in force is **retired, not deleted** — deleting it destroys the record that the clause once existed and was withdrawn.

To retire an entry, prefix its `clause` with a `[SUPERSEDED …]` marker naming what replaced it:

```text
clause: "[SUPERSEDED by <replacement>] <the original clause text>"
```

(The fence above is deliberately tagged `text`, not `yaml`. The registry loader reads the **first** yaml-tagged fence in this file as the entry list, so a yaml-tagged example placed before `## Entries` would be parsed as the registry and fail to load — and so would that fence marker written out literally in prose.)

`moai constitution validate` then counts the entry as retired and skips its drift, canary-gate, and source-file checks — the source text is gone by definition, so those checks could only ever fail. The entry still appears in `moai constitution list`, and the retired total is reported (`retired_count` in JSON output).

Two boundaries:

- The marker is a **prefix**, not a substring. A live clause that merely mentions `[SUPERSEDED …]` in its own text stays fully checked.
- `canary_gate: false` is **not** a retirement marker — it is the documented default for every Evolvable entry. A retired Frozen entry may carry `canary_gate: false` because it is retired, not the other way round.

`moai constitution validate --strict` ignores the marker and checks retired entries verbatim, for auditing what they still hold.

## Usage Guide

```bash
# List the entire registry
moai constitution list

# Filter by Frozen zone
moai constitution list --zone frozen

# List clauses from a specific file only
moai constitution list --file .claude/rules/moai/core/moai-constitution.md

# JSON-format output
moai constitution list --format json
```

## Entries

```yaml
# ============================================================
# 001-010: CLAUDE.md HARD clauses (§1 Hard Rules)
# ============================================================
- id: CONST-V3R2-001
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/spec-workflow.md
  anchor: "#plan-phase"
  clause: "Create comprehensive specification using EARS format."
  canary_gate: true

- id: CONST-V3R2-002
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#quality-gates"
  clause: "All code changes must pass TRUST 5 validation"
  canary_gate: true

- id: CONST-V3R2-003
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/mx-tag-protocol.md
  anchor: "#scope"
  clause: "This rule applies to all agents working with source code in the supported programming languages"
  canary_gate: true

- id: CONST-V3R2-004
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/development/coding-standards.md
  anchor: "#language-policy"
  clause: "All instruction documents must be in English:"
  canary_gate: true

- id: CONST-V3R2-005
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/development/coding-standards.md
  anchor: "#thin-command-pattern"
  clause: "All slash command files MUST be thin routing wrappers (under 20 LOC body)."
  canary_gate: true

- id: CONST-V3R2-006
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#user-interaction-boundary"
  clause: "`AskUserQuestion` is the **only** user-facing question channel"
  canary_gate: true

- id: CONST-V3R2-007
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#1-core-identity"
  clause: "MoAI is the Strategic Orchestrator for Claude Code."
  canary_gate: true

# ============================================================
# 008-020: CLAUDE.md HARD clauses (§1 Hard Rules — orchestrator behavior)
# ============================================================
- id: CONST-V3R2-008
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#response-language"
  clause: "All user-facing responses MUST be in the user's conversation_language."
  canary_gate: false

- id: CONST-V3R2-009
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#parallel-execution"
  clause: "Execute all independent tool calls in parallel when no dependencies exist."
  canary_gate: false

- id: CONST-V3R2-010
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#output-format"
  clause: "XML tags are reserved for agent-to-agent data transfer"
  canary_gate: false

- id: CONST-V3R2-011
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#output-format"
  clause: "Use Markdown for all user-facing communication"
  canary_gate: false

- id: CONST-V3R2-012
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#8-user-interaction-architecture"
  clause: "Every question directed at the user MUST be asked via AskUserQuestion."
  canary_gate: true

- id: CONST-V3R2-013
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "When intent is unclear, conduct a Socratic interview before execution"
  canary_gate: false

- id: CONST-V3R2-014
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "Before non-trivial code, explain the approach + which files change + why; get user approval"
  canary_gate: false

- id: CONST-V3R2-015
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "When modifying 3+ files, split into logical units (TodoList), execute file-by-file, analyze dependencies before parallel execution, report progress per unit"
  canary_gate: false

- id: CONST-V3R2-016
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "After coding, provide potential-issue list (edge cases, error/concurrency scenarios), suggested test cases, known limitations/assumptions, additional-validation recommendations"
  canary_gate: false

- id: CONST-V3R2-017
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#7-safe-development-protocol"
  clause: "Write a failing reproduction test first; confirm it fails; challenge the diagnosed root cause once"
  canary_gate: false

- id: CONST-V3R2-018
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#8-user-interaction-architecture"
  clause: "Every question directed at the user MUST be asked via AskUserQuestion. Free-form prose questions in response text are prohibited."
  canary_gate: true

- id: CONST-V3R2-019
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#8-user-interaction-architecture"
  clause: "`AskUserQuestion`, `TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet` are **deferred tools** — schemas NOT loaded at session start"
  canary_gate: true

# ============================================================
# 020-030: CLAUDE.md §14 Worktree Isolation Rules + §11 Background Agent
# ============================================================
- id: CONST-V3R2-020
  zone: Evolvable
  zone_class: frozen-safety
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "subagents run in the background by default (the runtime chooses foreground only when it needs the result; every permission prompt still surfaces in the main session); MoAI does not set `background:` — the retained safeguard is concurrency, not backgrounding"
  canary_gate: false

- id: CONST-V3R2-021
  zone: Evolvable
  zone_class: evolvable-experimental
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "[SUPERSEDED by worktree-opt-in policy — see CLAUDE.md §14 + worktree-integration.md § Terminology Glossary] Implementation teammates in team mode (role_profiles: implementer, tester, designer) MUST use isolation: worktree when spawned via Agent()"
  canary_gate: false

- id: CONST-V3R2-022
  zone: Evolvable
  zone_class: evolvable-experimental
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "[SUPERSEDED by worktree-opt-in policy — see CLAUDE.md §14 + worktree-integration.md § Terminology Glossary] Read-only teammates (role_profiles: researcher, analyst, reviewer) MUST NOT use isolation: worktree"
  canary_gate: false

- id: CONST-V3R2-023
  zone: Evolvable
  zone_class: evolvable-experimental
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "[SUPERSEDED by worktree-opt-in policy — see CLAUDE.md §14 + worktree-integration.md § Terminology Glossary] One-shot sub-agents making cross-file changes SHOULD use isolation: worktree"
  canary_gate: false

- id: CONST-V3R2-024
  zone: Evolvable
  zone_class: evolvable-experimental
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "[SUPERSEDED by worktree-opt-in policy — see CLAUDE.md §14 + worktree-integration.md § Terminology Glossary] GitHub workflow fixer agents MUST use isolation: worktree for branch isolation"
  canary_gate: false

# ============================================================
# 025-035: moai-constitution.md HARD clauses
# ============================================================
- id: CONST-V3R2-025
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#moai-orchestrator"
  clause: "AskUserQuestion is the sole user-facing question channel"
  canary_gate: true

- id: CONST-V3R2-026
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#moai-orchestrator"
  clause: "used ONLY by the MoAI orchestrator (subagents must never prompt users)"
  canary_gate: true

- id: CONST-V3R2-027
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#moai-orchestrator"
  clause: "Canonical reference: `.claude/rules/moai/core/askuser-protocol.md` § Channel Monopoly / § ToolSearch Preload Procedure / § Socratic Interview Structure / § Option Description Standards"
  canary_gate: true

- id: CONST-V3R2-028
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#opus-5-48-prompt-philosophy"
  clause: "Principle 4 — fewer subagents by default**: 4.7+ does not auto-spawn"
  canary_gate: false

- id: CONST-V3R2-029
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#opus-5-48-prompt-philosophy"
  clause: "Principle 5 — fewer tool calls by default**: specify when and why each"
  canary_gate: false

- id: CONST-V3R2-030
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Before implementing anything non-trivial, list assumptions explicitly and wait for user confirmation"
  canary_gate: false

- id: CONST-V3R2-031
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "When encountering inconsistencies, conflicting requirements, or unclear specifications, STOP and surface the confusion before proceeding"
  canary_gate: false

- id: CONST-V3R2-032
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Point out issues directly when an approach has clear problems. Sycophancy is a failure mode."
  canary_gate: false

- id: CONST-V3R2-033
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Actively resist overcomplexity. The natural tendency of code generation is toward over-engineering. Resist it."
  canary_gate: false

- id: CONST-V3R2-034
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Touch only what you were asked to touch. Drive-by refactors create noise and risk regressions."
  canary_gate: false

- id: CONST-V3R2-035
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#agent-core-behaviors"
  clause: "Every task requires evidence of completion."
  canary_gate: false

# ============================================================
# 036-045: agent-common-protocol.md HARD clauses
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
  clause: "Preload `AskUserQuestion` via `ToolSearch(query:"
  canary_gate: true

- id: CONST-V3R2-038
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#user-interaction-boundary"
  clause: "AskUserQuestion is reserved exclusively for the MoAI orchestrator"
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
  clause: "Agents are invoked through MoAI's natural language delegation pattern"
  canary_gate: false

- id: CONST-V3R2-044
  zone: Evolvable
  zone_class: frozen-safety
  file: .claude/rules/moai/core/agent-common-protocol.md
  anchor: "#background-agent-execution"
  clause: "The retained safeguard is **concurrency, not backgrounding**"
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
  clause: "The reviewer mode operates as a fresh-judgment auditor"
  canary_gate: false

# ============================================================
# 051-099: design/constitution.md [FROZEN] mirror entries (§2 + §3.1/§3.2/§3.3)
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
  clause: "moai-domain-copywriting MUST adhere to brand voice, tone, and terminology from brand-voice.md"
  canary_gate: true

- id: CONST-V3R2-063
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#31-brand-context-constitutional-parent"
  clause: "moai-domain-brand-design MUST use brand color palette, typography, and visual language from visual-identity.md"
  canary_gate: true

- id: CONST-V3R2-064
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#31-brand-context-constitutional-parent"
  clause: "expert-frontend MUST implement design tokens derived from brand context"
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
  clause: "MUST auto-load human-authored design documents (research.md, system.md, spec.md) when present and not _TBD_"
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
  clause: "`moai-workflow-design` continues to write machine-generated artifacts to `.moai/design/`"
  canary_gate: true

- id: CONST-V3R2-069
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#32-design-brief-execution-scope"
  clause: "Reserved file paths (canonical list): `tokens.json`, `components.json`, `assets/`, `import-warnings.json`, `brief/BRIEF-*.md`"
  canary_gate: true

- id: CONST-V3R2-070
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/design/constitution.md
  anchor: "#32-design-brief-execution-scope"
  clause: "Token budget for auto-loading is bounded by `.moai/config/sections/design.yaml` `design_docs.token_budget`; when the key is absent, the system MUST default to 20000"
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
  clause: "When both are present, brand constraints win on conflict."
  canary_gate: true

# ============================================================
# 150-159: session-handoff.md HARD clauses (new workflow rules;
#          model-specific threshold revision:
#          Trigger #1 = 1M context 50% / 200K context 90%; 5 triggers retained)
# ============================================================
- id: CONST-V3R2-150
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/session-handoff.md
  anchor: "#when-to-generate-5-triggers"
  clause: "The orchestrator MUST emit a paste-ready resume message when ANY of these conditions activate"
  canary_gate: false

- id: CONST-V3R2-151
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/session-handoff.md
  anchor: "#canonical-format-verbatim-spec"
  clause: "Resume message MUST follow this exact 6-block structure, **bounded by cut-line markers**"
  canary_gate: false

- id: CONST-V3R2-152
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/session-handoff.md
  anchor: "#auto-memory-integration-mandatory"
  clause: "Save the message to a memory project entry. Filename pattern: `project_<epic>_<spec>_<status>.md`"
  canary_gate: false

- id: CONST-V3R2-153
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/session-handoff.md
  anchor: "#canonical-format-verbatim-spec"
  clause: "`✂` symbol (U+2702 BLACK SCISSORS) is **preserved verbatim across all locales** — never translate or substitute"
  canary_gate: false

# ============================================================
# CONST-V3R5-001..041: new parallel namespace
# Completes coverage of unmapped [HARD] rules — 11 source files newly registered
# ============================================================
- id: CONST-V3R5-001
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/core/askuser-protocol.md
  anchor: "#orchestratorsubagent-boundary"
  clause: "Subagents MUST NOT invoke `AskUserQuestion`"
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
  anchor: "#entry-condition"
  clause: "The CI auto-fix loop MUST be entered ONLY when the orchestrator hands off"
  canary_gate: true

- id: CONST-V3R5-005
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#iteration-cap"
  clause: "The auto-fix loop MUST attempt at most **3 iterations**"
  canary_gate: true

- id: CONST-V3R5-006
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#iteration-cap"
  clause: "The AskUserQuestion at iteration > 3 MUST be a blocking call"
  canary_gate: true

- id: CONST-V3R5-007
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#patch-commit-rule-no-force-push"
  clause: "Every auto-fix patch MUST be applied as a **new commit** on the PR branch"
  canary_gate: true

- id: CONST-V3R5-008
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#askuserquestion-boundary"
  clause: "AskUserQuestion is the **exclusive user interaction channel**"
  canary_gate: true

- id: CONST-V3R5-009
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#askuserquestion-boundary"
  clause: "The orchestrator MUST preload AskUserQuestion via"
  canary_gate: true

- id: CONST-V3R5-010
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#semantic-failure-no-auto-patch"
  clause: "Semantic failures (data race, deadlock, panic, test assertion failure) MUST"
  canary_gate: true

- id: CONST-V3R5-011
  zone: Frozen
  zone_class: frozen-safety
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#secrets-and-credentials-protection"
  clause: "The auto-fix loop MUST NOT modify `.env`, `.env.*`, credentials files"
  canary_gate: true

- id: CONST-V3R5-012
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#audit-log-requirement"
  clause: "Every auto-fix iteration MUST be logged to"
  canary_gate: true

- id: CONST-V3R5-013
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/ci-autofix-protocol.md
  anchor: "#ci-infrastructure-preservation"
  clause: "The auto-fix loop MUST NOT modify CI watch infrastructure scripts or"
  canary_gate: true

# --- context-window-management.md (5 entries: V3R5-022..026) ---
- id: CONST-V3R5-022
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/context-window-management.md
  anchor: "#context-window-targets"
  clause: "Operational threshold is **model-specific**. Larger windows tolerate higher percentage utilization before stall risk dominates"
  canary_gate: false

- id: CONST-V3R5-023
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/context-window-management.md
  anchor: "#user-responsibilities"
  clause: "When usage crosses the model-specific threshold:"
  canary_gate: false

- id: CONST-V3R5-024
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/context-window-management.md
  anchor: "#user-responsibilities"
  clause: "The next action MUST be `/clear` — no further large work in the current session"
  canary_gate: false

- id: CONST-V3R5-025
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/context-window-management.md
  anchor: "#orchestrator-responsibilities"
  clause: "Pre-clear announcement: When the orchestrator detects accumulated context (input + output) approaching the model-specific threshold"
  canary_gate: false

- id: CONST-V3R5-026
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/context-window-management.md
  anchor: "#orchestrator-responsibilities"
  clause: "Resume message format: include all of the following so the next session is self-sufficient"
  canary_gate: false

# --- spec-workflow.md (2 new entries: V3R5-027..028; CONST-V3R2-001 covers the third) ---
- id: CONST-V3R5-027
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/spec-workflow.md
  anchor: "#spec-phase-discipline"
  clause: "Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step"
  canary_gate: true

- id: CONST-V3R5-028
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/spec-workflow.md
  anchor: "#spec-phase-discipline"
  clause: "Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged"
  canary_gate: true

# --- worktree-state-guard.md (1 entry: V3R5-029) ---
- id: CONST-V3R5-029
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/workflow/worktree-state-guard.md
  anchor: "#escalation-path"
  clause: "AskUserQuestion is invoked by the **orchestrator only**."
  canary_gate: true

# --- branch-origin-protocol.md (1 entry: V3R5-035) ---

- id: CONST-V3R5-035
  zone: Frozen
  zone_class: frozen-canonical
  file: .claude/rules/moai/development/branch-origin-protocol.md
  anchor: "#hard-rules"
  clause: "Skill body BODP gate MUST follow the askuser-protocol Socratic structure: `(권장)` first, ≤4 options, conversation_language match"
  canary_gate: true

# --- agent-authoring.md (1 entry: V3R5-037) ---
- id: CONST-V3R5-037
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/development/agent-authoring.md
  anchor: "#frontmatter-format-rules"
  clause: "Comma-separated string ONLY (`tools: Read, Write, Edit`). YAML arrays NOT supported"
  canary_gate: false

# --- skill-authoring.md (1 entry: V3R5-038) ---
- id: CONST-V3R5-038
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/development/skill-authoring.md
  anchor: "#key-format-rules"
  clause: "allowed-tools format: [ZONE:Evolvable] [HARD] Comma-separated string ONLY. Space-separated values are PROHIBITED"
  canary_gate: false

# --- session-handoff.md supplementary (1 entry: V3R5-039; covers worktree-anchored resume) ---
- id: CONST-V3R5-039
  zone: Evolvable
  zone_class: evolvable-tuning
  file: .claude/rules/moai/workflow/session-handoff.md
  anchor: "#worktree-anchored-resume-pattern"
  clause: "When the work happened inside a worktree, the resume message MUST prepend **Block 0 (cwd anchoring)** before the standard 6-block structure"
  canary_gate: false

# --- glm-web-tooling.md (2 entries: V3R5-040 mandate + V3R5-041 prohibition) ---
- id: CONST-V3R5-040
  zone: Frozen
  zone_class: frozen-safety
  file: .claude/rules/moai/core/glm-web-tooling.md
  anchor: "#hard-routing-table"
  clause: "They SHALL route web search to `mcp__web_search_prime__webSearchPrime`, web fetch to `mcp__web_reader__webReader`, and image reading to a `mcp__zai-mcp-server__*` vision tool"
  canary_gate: true

- id: CONST-V3R5-041
  zone: Frozen
  zone_class: frozen-safety
  file: .claude/rules/moai/core/glm-web-tooling.md
  anchor: "#hard-routing-table"
  clause: "While a session is GLM-backed, the built-in `WebSearch` / `WebFetch` tools and `Read`-on-an-image-file are **PROHIBITED** because they route through the 529-prone `api.z.ai/api/anthropic` gateway and the base64→422 image path"
  canary_gate: true

# ============================================================
# CONST-V3R6-NNN: V3R6 modern-era parallel namespace
# (first V3R6 entry: a runtime-recovery predecessor SPEC, M3)
# ============================================================
# --- runtime-recovery-doctrine.md (1 entry: V3R6-001 anti-death-spiral) ---
- id: CONST-V3R6-001
  zone: Evolvable
  zone_class: frozen-safety
  file: .claude/rules/moai/workflow/runtime-recovery-doctrine.md
  anchor: "#4-anti-death-spiral-hook-carve-out-documentation-only-policy"
  clause: "Stop/PostToolUse hooks SHOULD exit 0 (allow the turn to end / the tool call to proceed) rather than exit 2 (block), so that recovery turns are NOT placed into the `error → stop-hook-blocks → retry → error` loop"
  canary_gate: true
```
