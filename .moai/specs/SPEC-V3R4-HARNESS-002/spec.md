---
id: SPEC-V3R4-HARNESS-002
version: "0.2.1"
status: draft
created_at: 2026-05-14
updated_at: 2026-05-14
author: manager-spec
priority: P1
labels: [harness, self-evolution, multi-event-observer, v3r4, downstream, observer-expansion]
issue_number: null
phase: "v3.0.0 R4 — Phase B — Observer Expansion"
module: ".claude/hooks/moai/, internal/cli/hook.go, internal/harness/observer.go, internal/harness/types.go, .moai/harness/usage-log.jsonl (schema extension), .claude/settings.json (hook registration template)"
dependencies:
  - SPEC-V3R4-HARNESS-001
supersedes: []
related_specs:
  - SPEC-V3R4-HARNESS-001
breaking: false
bc_id: []
lifecycle: spec-anchored
related_theme: "Self-Evolving Harness v2 — Observer Expansion Wave"
target_release: v2.21.0
---

# SPEC-V3R4-HARNESS-002 — Multi-Event Observer Expansion

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.2.1   | 2026-05-14 | manager-spec | plan-auditor iteration 1 mechanical fixes — removed deprecated `title:` frontmatter field (D1), normalized standalone AC leaves to canonical `(maps REQ-...)` tail format (D2). No content semantic changes. |
| 0.2.0   | 2026-05-14 | manager-spec | Plan-phase expansion. Replaces §3 Requirements placeholder with full EARS-format enumeration (REQ-HRN-OBS-001 ~ REQ-HRN-OBS-018). Replaces §4 Acceptance Criteria placeholder with hierarchical AC enumeration (AC-HRN-OBS-001 ~ AC-HRN-OBS-013) mapped 1:N to REQs. Adds §1.1 Constitutional Contract Preservation referencing the four V3R4-001 contracts (REQ-HRN-FND-005/009/010/011/015). Renames frontmatter fields to canonical 9-field schema (`created_at` / `updated_at` / `labels`). Bumps version 0.1.0 → 0.2.0. Adds research.md, plan.md, acceptance.md, tasks.md sibling files. |
| 0.1.0   | 2026-05-14 | manager-spec | Initial plan-phase entry seed. Created as the first downstream SPEC after SPEC-V3R4-HARNESS-001 (foundation) merged. Scope: extend the PostToolUse-only observer baseline established by V3R4-HARNESS-001 to cover Stop, SubagentStop, and UserPromptSubmit Claude Code hook events with a unified observation schema. Detailed EARS-format requirements and acceptance criteria are reserved for the next plan-phase session. |

---

## 1. Goal

The V3R4 self-evolving harness observation surface MUST expand from the current PostToolUse-only baseline (per `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-010) to a multi-event observer covering Claude Code's `Stop`, `SubagentStop`, and `UserPromptSubmit` hook events in addition to `PostToolUse`. All four event types MUST share a unified observation schema (extended JSONL append-only format under `.moai/harness/usage-log.jsonl`) so downstream classifiers (frequency-count today, embedding-cluster in `SPEC-V3R4-HARNESS-003`) can aggregate cross-event patterns without per-event schema branching. The `learning.enabled` gate established in `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-009 MUST extend to ALL new event hooks (no-op when disabled). The 5-Layer Safety architecture (REQ-HRN-FND-005), the 4-tier ladder (REQ-HRN-FND-011), and the subagent AskUserQuestion prohibition (REQ-HRN-FND-015) MUST remain intact and unmodified by V3R4-002.

### 1.1 Background

- `SPEC-V3R4-HARNESS-001` shipped the PostToolUse-only observer baseline with a 4-field schema (`timestamp` / `event_type` / `subject` / `context_hash`). It explicitly deferred multi-event expansion to this SPEC (`spec.md` §1.3 Non-Goals item 1).
- Claude Code v2.1.110+ exposes four primary lifecycle hooks the harness can observe: `PostToolUse` (already wired), `Stop` (session end signal), `SubagentStop` (Agent() teammate exit), `UserPromptSubmit` (user input received). Each provides distinct learning signals: PostToolUse → tool usage frequency, Stop → session-level completion patterns, SubagentStop → agent invocation outcomes, UserPromptSubmit → user intent patterns.
- The existing observer code path (`internal/cli/hook.go::runHarnessObserve` + `internal/harness/observer.go`) was designed for a single event type. Multi-event support requires schema extension (event-specific payload fields) and routing logic for the additional `moai hook` subcommands.

### 1.2 Constitutional Contract Preservation (verbatim citation of V3R4-001)

This SPEC is bound by, and MUST preserve, the following load-bearing contracts shipped by SPEC-V3R4-HARNESS-001 (PR #909/#910/#911, merged commit `bb80ea0f4`). Any V3R4-002 design that weakens these contracts is a scope violation:

| V3R4-001 REQ | Contract preserved by V3R4-002 |
|--------------|--------------------------------|
| REQ-HRN-FND-005 | 5-Layer Safety architecture (Frozen Guard / Canary Check / Contradiction Detector / Rate Limiter / Human Oversight) preserved unchanged. V3R4-002 introduces NO new bypass. See §6 REQ-HRN-OBS-007. |
| REQ-HRN-FND-009 | `learning.enabled` gate is a fail-open boolean read from `.moai/config/sections/harness.yaml`. V3R4-002 reuses the SAME `isHarnessLearningEnabled` function for all four event hooks. See §6 REQ-HRN-OBS-008. |
| REQ-HRN-FND-010 | Baseline JSONL schema (`timestamp` / `event_type` / `subject` / `context_hash`) holds. V3R4-002 adds OPTIONAL fields with `omitempty` JSON tags. Old entries lacking optional fields remain valid. See §6 REQ-HRN-OBS-009. |
| REQ-HRN-FND-011 | 4-tier ladder (Observation 1× / Heuristic 3× / Rule 5× / AutoUpdate 10×) preserved unchanged. V3R4-002 does NOT modify the classifier. See §6 REQ-HRN-OBS-010. |
| REQ-HRN-FND-015 | Subagent AskUserQuestion prohibition. V3R4-002 introduces no new AskUserQuestion call sites and adds no subagent definition. See §6 REQ-HRN-OBS-011. |

### 1.3 Scope (preliminary — full enumeration in next plan session)

**In Scope**:

- Add `moai hook harness-observe-stop`, `moai hook harness-observe-subagent-stop`, `moai hook harness-observe-user-prompt-submit` cobra subcommands that wire the harness observer for the respective events.
- Extend the `Event` struct in `internal/harness/types.go` with optional per-event payload fields (all tagged `omitempty`) while preserving the four baseline fields from REQ-HRN-FND-010.
- Extend the `EventType` enum in `internal/harness/types.go` with three new SEMANTIC values: `session_stop`, `subagent_stop`, `user_prompt`. The existing four enum values (`moai_subcommand`, `agent_invocation`, `spec_reference`, `feedback`) are preserved.
- Reuse `isHarnessLearningEnabled` gate (REQ-HRN-FND-009) so all four new event hooks become complete no-ops when `learning.enabled: false`.
- Create per-event Claude Code hook wrapper scripts: `handle-harness-observe-stop.sh`, `handle-harness-observe-subagent-stop.sh`, `handle-harness-observe-user-prompt-submit.sh` (template-first authored under `internal/template/templates/.claude/hooks/moai/`).
- Update `internal/template/templates/.claude/settings.json.tmpl` to register the new hook entries under their respective event slots (additive, Strategy WIRE-A per research.md §2.5).
- Adopt **Strategy A (SHA-256 hash + length + language)** as the default PII handling policy for UserPromptSubmit; provide opt-in keys `learning.user_prompt_content: hash|preview|full|none` for richer or weaker telemetry; reject unknown values with fail-open to Strategy A.

**Out of Scope (deferred to downstream V3R4 SPECs)**:

- Embedding-cluster classifier replacing frequency-count tier ladder — `SPEC-V3R4-HARNESS-003`.
- Reflexion self-critique loop — `SPEC-V3R4-HARNESS-004`.
- Principle-based scoring — `SPEC-V3R4-HARNESS-005`.
- Multi-objective effectiveness measurement — `SPEC-V3R4-HARNESS-006`.
- Voyager skill library — `SPEC-V3R4-HARNESS-007`.
- Cross-project federation — `SPEC-V3R4-HARNESS-008`.

### 1.4 Non-Goals

This SPEC is the OBSERVER EXPANSION. The following capabilities are explicitly OUT OF SCOPE and are deferred to the named downstream SPECs:

- Frequency-count classifier replacement — deferred to `SPEC-V3R4-HARNESS-003`.
- Verbal self-critique loop — deferred to `SPEC-V3R4-HARNESS-004`.
- Principle-based scoring rubric — deferred to `SPEC-V3R4-HARNESS-005`.
- Multi-objective scoring tuple — deferred to `SPEC-V3R4-HARNESS-006`.
- Skill library auto-organization — deferred to `SPEC-V3R4-HARNESS-007`.
- Cross-project federation — deferred to `SPEC-V3R4-HARNESS-008`.
- Any migration of historical `usage-log.jsonl` entries that lack the extended schema. Old entries remain valid under the baseline 4-field contract from REQ-HRN-FND-010; new entries MAY include extended fields.
- Retroactively wiring the PostToolUse hook entry in `settings.json.tmpl` to invoke `handle-harness-observe.sh`. This is a V3R4-001 follow-up gap (per research.md §2.5) and remains out of scope for this SPEC.
- Networking, telemetry upload, or any cross-machine data exchange.
- GUI / dashboard for inspecting per-event observation logs.

---

## 2. Stakeholders

| Role | Interest |
|------|----------|
| MoAI-ADK maintainer | Single source of truth for harness observation surface; uniform `learning.enabled` gate enforcement across all four event types; settings.json.tmpl additive change minimizes conflict risk. |
| Solo developer / power user (GOOS-style) | Richer learning signals without manual intervention; preserved per-project no-op via `learning.enabled: false`; PII-safe default (Strategy A) requires no opt-in for privacy. |
| Privacy-conscious user | Strategy A default ensures no raw user content is recorded; opt-in mechanism (`learning.user_prompt_content`) provides explicit consent gate for richer telemetry. |
| `manager-spec` (this SPEC's author for downstream sessions) | Clean enumeration of REQ IDs and acceptance criteria for plan-auditor cycle. |
| `plan-auditor` subagent | Single SPEC to audit; clear `dependencies` relationship for impact analysis; REQ↔AC matrix complete; V3R4-001 contract preservation verifiable via REQ-HRN-OBS-007/008/009/010/011. |
| `evaluator-active` subagent | Unchanged binding-gate role; new observation entries feed downstream classifier and Tier-4 proposal generation; richer signal improves classifier accuracy. |
| Downstream SPEC authors (V3R4-003 through V3R4-008) | This SPEC's multi-event observation stream is the prerequisite input for embedding-cluster classifier (SPEC-V3R4-HARNESS-003); the schema extension contract is binding. |
| `manager-git` subagent | Owns the V3R4-002 plan PR squash-merge; no breaking change ID required (breaking: false). |

---

## 3. Exclusions (What NOT to Build)

[HARD] This SPEC explicitly EXCLUDES the following — building any of these within this PR is a scope violation:

1. Any embedding model, embedding-cluster algorithm, or vector index — deferred to SPEC-V3R4-HARNESS-003.
2. Any Reflexion-style self-critique loop, episodic memory, or iteration-cap mechanism — deferred to SPEC-V3R4-HARNESS-004.
3. Any principle-based scoring rubric or constitution-parsing logic — deferred to SPEC-V3R4-HARNESS-005.
4. Any multi-objective scoring tuple or auto-rollback-on-regression mechanism — deferred to SPEC-V3R4-HARNESS-006.
5. Any embedding-indexed skill library or top-K retrieval — deferred to SPEC-V3R4-HARNESS-007.
6. Any cross-project lesson sharing, anonymization layer, or federation — deferred to SPEC-V3R4-HARNESS-008.
7. Migration tooling that converts pre-existing `.moai/harness/usage-log.jsonl` entries into the new schema. Users continue with whatever state they have on disk.
8. Modification of `.claude/rules/moai/design/constitution.md` (FROZEN file).
9. Retroactive wiring of `handle-harness-observe.sh` into the PostToolUse slot of `settings.json.tmpl` (V3R4-001 follow-up gap, separately tracked).
10. Modification of `.claude/agents/moai/**`, `.claude/skills/moai-*/**`, `.claude/rules/moai/**`, or `.moai/project/brand/**` (FROZEN per REQ-HRN-FND-006).
11. Modification of the `SchemaVersion` field on the `Event` struct. Additive optional fields do not require a version bump; the field remains `v1`.
12. New AskUserQuestion call sites in any subagent (preserved REQ-HRN-FND-015).
13. Network call, telemetry upload, or external API integration.

---

## 4. Dependencies

| SPEC | Relationship | Notes |
|------|--------------|-------|
| `SPEC-V3R4-HARNESS-001` | Hard dependency (foundation) | Establishes the PostToolUse observer baseline, `learning.enabled` gate (REQ-HRN-FND-009), FROZEN zone (REQ-HRN-FND-006), 5-Layer Safety (REQ-HRN-FND-005), and `usage-log.jsonl` schema (REQ-HRN-FND-010) this SPEC extends. The 4-tier ladder (REQ-HRN-FND-011) is preserved unchanged by this SPEC. |
| `SPEC-V3R4-HARNESS-003` | Downstream (blocked by this SPEC) | Embedding-cluster classifier consumes the multi-event observation stream produced by this SPEC. The new EventType enum values introduced here are the input vocabulary for the cluster algorithm. |
| `SPEC-V3R4-HARNESS-004` through `SPEC-V3R4-HARNESS-008` | Downstream (indirectly blocked) | All build on the V3R4 architecture established by V3R4-HARNESS-001 and the observation surface expanded by this SPEC. |

---

## 5. Constraints

[HARD] Language: All SPEC artifact content body in English where possible. Conversation-language Korean is acceptable per `.moai/config/sections/language.yaml` `conversation_language: ko`. EARS keywords (WHEN / WHILE / WHERE / IF / SHALL) remain English. Code identifiers (function names, file paths, REQ IDs, AC IDs, EventType enum values) remain English.

[HARD] FROZEN zone preservation: `.claude/rules/moai/design/constitution.md` §2 and §5 are NOT modified by this SPEC. The four V3R4-001 contracts (REQ-HRN-FND-005/009/010/011/015) are referenced verbatim in §1.2 above.

[HARD] EARS format mandatory for all REQs. Every REQ in §6 uses one of the five EARS patterns: Ubiquitous, Event-Driven, State-Driven, Optional, Unwanted.

[HARD] No tech-stack implementation assumptions in this spec.md. Requirements describe contracts and capabilities. Implementation decisions (file paths, Go struct layout, exact function signatures) belong to plan.md.

[HARD] Conventional Commits format for the plan PR commit message.

[HARD] No emojis in user-facing output (per `.claude/rules/moai/development/coding-standards.md` § Content Restrictions, where applicable).

[HARD] No time estimates or duration predictions in any artifact (per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation). Use priority labels (P0 / P1 / P2 / P3) and phase ordering.

[HARD] Template-First discipline (per CLAUDE.local.md §2): all template changes happen in `internal/template/templates/` first; `make build` regenerates `internal/template/embedded.go`. Local `.claude/` and `.moai/` files are NOT directly edited.

[HARD] Schema additivity rule (per REQ-HRN-OBS-009): all new fields in the `Event` struct are tagged `omitempty`. Old PostToolUse-only entries remain valid under the baseline 4-field contract.

---

## 6. Requirements (EARS format)

Each requirement is identified by `REQ-HRN-OBS-NNN` and uses one of the five EARS patterns: Ubiquitous (system shall always), Event-Driven (when X then system shall Y), State-Driven (while X the system shall Y), Optional Feature (where X exists the system shall Y), Unwanted Behavior (if X then system shall not / shall reject Y). Most lifecycle rules use Event-Driven because each new hook is an event-triggered observation point.

### REQ-HRN-OBS-001 (Ubiquitous — cobra subcommand coverage for Stop)

The system **shall** expose a cobra subcommand `moai hook harness-observe-stop` registered under `hookCmd` in `internal/cli/hook.go`. The subcommand **shall** be reachable from the corresponding Claude Code hook wrapper script.

### REQ-HRN-OBS-002 (Ubiquitous — cobra subcommand coverage for SubagentStop)

The system **shall** expose a cobra subcommand `moai hook harness-observe-subagent-stop` registered under `hookCmd` in `internal/cli/hook.go`. The subcommand **shall** be reachable from the corresponding Claude Code hook wrapper script.

### REQ-HRN-OBS-003 (Ubiquitous — cobra subcommand coverage for UserPromptSubmit)

The system **shall** expose a cobra subcommand `moai hook harness-observe-user-prompt-submit` registered under `hookCmd` in `internal/cli/hook.go`. The subcommand **shall** be reachable from the corresponding Claude Code hook wrapper script.

### REQ-HRN-OBS-004 (Event-Driven — Stop observation entry)

**When** the `Stop` hook fires from Claude Code AND `learning.enabled` resolves to true, the system **shall** append exactly one JSONL entry to `.moai/harness/usage-log.jsonl` containing:
- the four baseline fields from REQ-HRN-FND-010 (`timestamp`, `event_type`, `subject`, `context_hash`)
- `event_type` set to `session_stop`
- the optional extended fields `session_id`, `last_assistant_message_hash`, `last_assistant_message_len` when their source values are present in the hook stdin
- absent extended fields **shall** be omitted from the JSONL line (`omitempty` serialization)

### REQ-HRN-OBS-005 (Event-Driven — SubagentStop observation entry)

**When** the `SubagentStop` hook fires from Claude Code AND `learning.enabled` resolves to true, the system **shall** append exactly one JSONL entry to `.moai/harness/usage-log.jsonl` containing:
- the four baseline fields from REQ-HRN-FND-010
- `event_type` set to `subagent_stop`
- `subject` set to the `agentName` value from the hook stdin (e.g., `manager-spec`, `expert-backend`)
- the optional extended fields `agent_name`, `agent_type`, `agent_id`, `parent_session_id` when their source values are present in the hook stdin

### REQ-HRN-OBS-006 (Event-Driven — UserPromptSubmit observation entry)

**When** the `UserPromptSubmit` hook fires from Claude Code AND `learning.enabled` resolves to true AND `learning.user_prompt_content` is absent or set to `hash` (the default Strategy A), the system **shall** append exactly one JSONL entry to `.moai/harness/usage-log.jsonl` containing:
- the four baseline fields from REQ-HRN-FND-010
- `event_type` set to `user_prompt`
- the optional extended fields `prompt_hash` (first 16 hex chars of SHA-256(prompt)), `prompt_len` (byte length), `prompt_lang` (ISO-639-1 code from Unicode-block heuristic, empty if undetectable)
- the raw `prompt` text **shall NOT** appear anywhere in the JSONL entry

### REQ-HRN-OBS-007 (Ubiquitous — 5-Layer Safety preservation)

The system **shall** preserve the 5-Layer Safety architecture defined in `.claude/rules/moai/design/constitution.md` §5 and re-asserted in `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-005 in its current form. The expanded observation surface introduced by this SPEC **shall not** remove, bypass, or weaken any of the five layers (L1 Frozen Guard, L2 Canary Check, L3 Contradiction Detector, L4 Rate Limiter, L5 Human Oversight).

### REQ-HRN-OBS-008 (State-Driven — unified gate for all four event hooks)

**While** the configuration key `learning.enabled` in `.moai/config/sections/harness.yaml` resolves to `false`, ALL four event hooks (PostToolUse, Stop, SubagentStop, UserPromptSubmit) **shall** be complete no-ops: each handler **shall not** read, write, or append to `.moai/harness/usage-log.jsonl` nor invoke any tier-classification or evolution logic. Disabling **shall not** delete existing log entries. The gate function used by all four handlers **shall** be the shared `isHarnessLearningEnabled` introduced by SPEC-V3R4-HARNESS-001.

### REQ-HRN-OBS-009 (Ubiquitous — schema additivity preserved)

The system **shall** preserve the four baseline fields (`timestamp`, `event_type`, `subject`, `context_hash`) defined by `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-010 in every JSONL entry produced by any of the four event hooks. New fields introduced by this SPEC **shall** be Go-struct-tagged with `omitempty` so they are absent from the serialized JSONL line when their source value is empty. Pre-existing JSONL entries (PostToolUse-only baseline) **shall** remain valid under the contract; no migration occurs.

### REQ-HRN-OBS-010 (Ubiquitous — 4-tier ladder preservation)

The system **shall** preserve the existing 4-tier observation ladder with thresholds 1 (Observation), 3 (Heuristic), 5 (Rule), 10 (Auto-update) as defined in `SPEC-V3R3-HARNESS-LEARNING-001` REQ-HL-002 and re-asserted in `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-011. Replacement of the frequency-count classifier with an embedding-cluster classifier is deferred to `SPEC-V3R4-HARNESS-003`; this SPEC only expands the observation INPUT stream, not the classifier.

### REQ-HRN-OBS-011 (Ubiquitous — subagent AskUserQuestion prohibition preserved)

The system **shall** enforce the orchestrator-only AskUserQuestion contract from `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary and `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-015 across every handler introduced by this SPEC. The new event handlers (`runHarnessObserveStop`, `runHarnessObserveSubagentStop`, `runHarnessObserveUserPromptSubmit`) **shall not** invoke `AskUserQuestion` and **shall not** spawn any subagent that invokes `AskUserQuestion`. This SPEC introduces no new subagent definitions.

### REQ-HRN-OBS-012 (Ubiquitous — PII privacy default, Strategy A)

The system **shall** default to Strategy A (SHA-256 hash + length + language heuristic) for UserPromptSubmit observation. The raw `prompt` content **shall not** be recorded in `.moai/harness/usage-log.jsonl` under the default configuration. The system **shall** apply Strategy A when `learning.user_prompt_content` is absent in `.moai/config/sections/harness.yaml`.

### REQ-HRN-OBS-013 (Optional Feature — UserPromptSubmit opt-in strategies)

**Where** the configuration key `learning.user_prompt_content` is explicitly set to `preview`, the system **shall** record the first 64 bytes of the prompt as `prompt_preview` (Strategy B) in addition to the Strategy A fields.

**Where** the configuration key is explicitly set to `full`, the system **shall** record the entire prompt as `prompt_content` (Strategy C) in addition to the Strategy A fields.

**Where** the configuration key is explicitly set to `none`, the UserPromptSubmit hook **shall** become a complete no-op for the observer (the gate from REQ-HRN-OBS-008 is unaffected).

### REQ-HRN-OBS-014 (Unwanted — invalid PII config fails open to Strategy A)

**If** the configuration key `learning.user_prompt_content` is present in `.moai/config/sections/harness.yaml` but contains a value not in the set `{hash, preview, full, none}`, **then** the system **shall** fall back to Strategy A (hash) and **shall not** record raw content under any circumstance. Fail-open to strongest privacy is the safety invariant.

### REQ-HRN-OBS-015 (Ubiquitous — EventType enum extension)

The system **shall** extend the `EventType` enum in `internal/harness/types.go` with three new SEMANTIC values: `session_stop`, `subagent_stop`, `user_prompt`. The existing four values (`moai_subcommand`, `agent_invocation`, `spec_reference`, `feedback`) **shall** remain unchanged. Downstream classifiers MAY treat the four new values as belonging to a distinct cluster for cross-event aggregation.

### REQ-HRN-OBS-016 (Ubiquitous — wrapper script template-first authoring)

The system **shall** ship Claude Code hook wrapper scripts (`handle-harness-observe-stop.sh`, `handle-harness-observe-subagent-stop.sh`, `handle-harness-observe-user-prompt-submit.sh`) authored first under `internal/template/templates/.claude/hooks/moai/` as `.sh.tmpl` files. The wrappers **shall** follow the existing 33-line pattern: stderr log path, log rotation at 10MB, binary search order (PATH → `~/go/bin/moai` → `~/.local/bin/moai`), `exec moai hook <subcommand>`, silent exit 0 if binary not found.

### REQ-HRN-OBS-017 (Ubiquitous — settings.json.tmpl additive registration)

The system **shall** register the three new wrapper scripts in `internal/template/templates/.claude/settings.json.tmpl` as ADDITIVE entries under their respective event slots (`Stop`, `SubagentStop`, `UserPromptSubmit`). The existing entries (`handle-stop.sh`, `handle-subagent-stop.sh`, `handle-user-prompt-submit.sh`) **shall** remain unchanged. The new entries **shall** carry timeout 5 seconds, type `command`, and platform-conditional path quoting matching the existing pattern.

### REQ-HRN-OBS-018 (Unwanted — hook latency budget)

**If** any of the four event observer handlers (`runHarnessObserve`, `runHarnessObserveStop`, `runHarnessObserveSubagentStop`, `runHarnessObserveUserPromptSubmit`) exceeds 100ms wall-clock time per invocation under typical workloads (a single `Observer.RecordEvent` call, no retention pruning), **then** the implementation **shall** be considered out of scope for direct merge and **shall** require a follow-up optimization SPEC. The budget mirrors the existing REQ-HL-001 latency bound for the PostToolUse observer.

---

## 7. Acceptance Coverage Map

The acceptance criteria are defined in `acceptance.md` (sibling file). The coverage map below shows every REQ from §6 mapped to at least one AC. Full Given-When-Then scenarios are in `acceptance.md`.

| AC ID | Covers REQ IDs |
|-------|----------------|
| AC-HRN-OBS-001 | REQ-HRN-OBS-001, REQ-HRN-OBS-002, REQ-HRN-OBS-003 |
| AC-HRN-OBS-002 | REQ-HRN-OBS-004 |
| AC-HRN-OBS-003 | REQ-HRN-OBS-005 |
| AC-HRN-OBS-004 | REQ-HRN-OBS-006 |
| AC-HRN-OBS-005 | REQ-HRN-OBS-008 |
| AC-HRN-OBS-006 | REQ-HRN-OBS-009 |
| AC-HRN-OBS-007 | REQ-HRN-OBS-012 |
| AC-HRN-OBS-008 | REQ-HRN-OBS-013 |
| AC-HRN-OBS-009 | REQ-HRN-OBS-014 |
| AC-HRN-OBS-010 | REQ-HRN-OBS-015 |
| AC-HRN-OBS-011 | REQ-HRN-OBS-016, REQ-HRN-OBS-017 |
| AC-HRN-OBS-012 | REQ-HRN-OBS-007, REQ-HRN-OBS-010, REQ-HRN-OBS-011 |
| AC-HRN-OBS-013 | REQ-HRN-OBS-018 |

Coverage: 18 REQs ↔ 13 ACs, every REQ appears in at least one AC.

---

## 8. Risks

| Risk | Likelihood | Severity | Mitigation |
|------|------------|----------|------------|
| PII leakage if Strategy A is bypassed by a misconfiguration | Low | High | REQ-HRN-OBS-014 enforces fail-open to Strategy A on unknown values. Tests verify the fail-open path explicitly. |
| Performance regression from three additional JSONL writes per turn | Medium | Low | REQ-HRN-OBS-018 enforces 100ms latency budget per handler. Each Observer.RecordEvent is bounded; lazy retention pruning is non-blocking. |
| Settings.json.tmpl conflict with concurrent SPECs editing the same template | Medium | Medium | Strategy WIRE-A (additive) per research.md §2.5. Concurrent SPECs that don't touch Stop / SubagentStop / UserPromptSubmit slots cannot conflict. |
| Test isolation — three new observer code paths each need NoOp / Records / PreservesExisting test triple | Low | Low | Pattern reuse from existing `hook_harness_observe_test.go`. Table-driven pattern keeps LOC low. |
| Template-first discipline violation if developer edits `.claude/hooks/moai/` directly without updating `internal/template/templates/` | Low | Medium | CI guard (existing in `internal/template/embedded_drift_test.go` or equivalent) detects drift. Documented in CLAUDE.local.md §2. |
| EventType enum extension breaks an unknown downstream consumer | Low | Medium | The enum is a Go string type; new values cannot break existing parsers that treat unknown values as a default category. Downstream classifier code is unchanged by this SPEC (REQ-HRN-OBS-010 preservation). |

---

## 9. References

- `SPEC-V3R4-HARNESS-001/spec.md` §1.3 Non-Goals item 1 (this SPEC's scope source).
- `SPEC-V3R4-HARNESS-001/spec.md` REQ-HRN-FND-005, REQ-HRN-FND-009, REQ-HRN-FND-010, REQ-HRN-FND-011, REQ-HRN-FND-015 (preserved contracts).
- `SPEC-V3R4-HARNESS-001/acceptance.md` AC-HRN-FND-007, AC-HRN-FND-008 (pattern source for V3R4-002 ACs).
- `internal/cli/hook.go::runHarnessObserve` (baseline implementation; pattern to clone).
- `internal/cli/hook.go::isHarnessLearningEnabled` (shared gate function; reused unchanged).
- `internal/harness/observer.go::Observer.RecordEvent` (JSONL append point; reused unchanged).
- `internal/harness/types.go::Event`, `EventType` (struct + enum extension targets).
- `internal/cli/hook_harness_observe_test.go` (baseline test suite; pattern to clone).
- `internal/template/templates/.claude/hooks/moai/handle-harness-observe.sh.tmpl` (wrapper template; pattern to clone).
- `internal/template/templates/.claude/settings.json.tmpl` (hook registration template; additive edit target).
- `.claude/rules/moai/core/hooks-system.md` § Hook Event stdin/stdout Reference (Claude Code event payload reference).
- `.claude/rules/moai/design/constitution.md` §2 (Frozen vs Evolvable), §5 (5-Layer Safety).
- `.moai/specs/SPEC-V3R4-HARNESS-002/research.md` (this SPEC's research artifact).
- `.moai/specs/SPEC-V3R4-HARNESS-002/plan.md` (this SPEC's implementation plan).
- `.moai/specs/SPEC-V3R4-HARNESS-002/acceptance.md` (this SPEC's acceptance criteria).
- `.moai/specs/SPEC-V3R4-HARNESS-002/tasks.md` (this SPEC's task breakdown).

---

End of spec.md.
