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
dependencies:
  - SPEC-V3R4-HARNESS-001
target_release: v2.21.0
breaking: false
generated_from: spec.md v0.2.1 (auto-extracted for run-phase token efficiency)
---

# SPEC-V3R4-HARNESS-002 — Compact (Run-Phase)

> This file is auto-generated from `spec.md`. It contains REQ enumeration, AC coverage map, Files-to-modify, and Exclusions. The full SPEC (background, stakeholders, constraints, risks, references) lives in `spec.md`. Run-phase agents read `spec-compact.md` + `acceptance.md` + `plan.md` + `tasks.md`.

---

## Files to Modify

Extracted from spec.md §1.3 In Scope:

- `internal/cli/hook.go` — register 3 new cobra subcommands (`harness-observe-stop`, `harness-observe-subagent-stop`, `harness-observe-user-prompt-submit`); reuse `isHarnessLearningEnabled` gate
- `internal/harness/types.go` — extend `EventType` enum with 3 new values (`session_stop`, `subagent_stop`, `user_prompt`); extend `Event` struct with optional per-event fields (`omitempty`)
- `internal/harness/observer.go` — no signature change; consumes extended `Event` struct
- `internal/cli/hook_harness_observe_test.go` — extend test suite for 3 new event types (NoOp / Records / PreservesExisting patterns)
- `internal/template/templates/.claude/hooks/moai/handle-harness-observe-stop.sh.tmpl` — NEW wrapper script template
- `internal/template/templates/.claude/hooks/moai/handle-harness-observe-subagent-stop.sh.tmpl` — NEW wrapper script template
- `internal/template/templates/.claude/hooks/moai/handle-harness-observe-user-prompt-submit.sh.tmpl` — NEW wrapper script template
- `internal/template/templates/.claude/settings.json.tmpl` — ADDITIVE entries under `Stop`, `SubagentStop`, `UserPromptSubmit` slots
- `.moai/config/sections/harness.yaml` — add `learning.user_prompt_content` key (default absent → Strategy A)

Template-First discipline: all template changes happen first in `internal/template/templates/`; `make build` regenerates `internal/template/embedded.go`.

---

## Requirements (EARS)

18 REQs covering cobra subcommand wiring, event-driven observation entries, V3R4-001 contract preservation, PII handling, schema additivity, EventType extension, template-first authoring, and latency budget.

### REQ-HRN-OBS-001 (Ubiquitous)

The system **shall** expose a cobra subcommand `moai hook harness-observe-stop` registered under `hookCmd` in `internal/cli/hook.go`. The subcommand **shall** be reachable from the corresponding Claude Code hook wrapper script.

### REQ-HRN-OBS-002 (Ubiquitous)

The system **shall** expose a cobra subcommand `moai hook harness-observe-subagent-stop` registered under `hookCmd` in `internal/cli/hook.go`. The subcommand **shall** be reachable from the corresponding Claude Code hook wrapper script.

### REQ-HRN-OBS-003 (Ubiquitous)

The system **shall** expose a cobra subcommand `moai hook harness-observe-user-prompt-submit` registered under `hookCmd` in `internal/cli/hook.go`. The subcommand **shall** be reachable from the corresponding Claude Code hook wrapper script.

### REQ-HRN-OBS-004 (Event-Driven — Stop)

**When** the `Stop` hook fires from Claude Code AND `learning.enabled` resolves to true, the system **shall** append exactly one JSONL entry to `.moai/harness/usage-log.jsonl` containing: baseline 4 fields (REQ-HRN-FND-010) + `event_type: session_stop` + optional extended fields `session_id`, `last_assistant_message_hash`, `last_assistant_message_len` (omitempty when source absent).

### REQ-HRN-OBS-005 (Event-Driven — SubagentStop)

**When** the `SubagentStop` hook fires from Claude Code AND `learning.enabled` resolves to true, the system **shall** append exactly one JSONL entry containing: baseline 4 fields + `event_type: subagent_stop` + `subject: agentName from stdin` + optional extended fields `agent_name`, `agent_type`, `agent_id`, `parent_session_id` (omitempty when source absent).

### REQ-HRN-OBS-006 (Event-Driven — UserPromptSubmit)

**When** the `UserPromptSubmit` hook fires from Claude Code AND `learning.enabled` resolves to true AND `learning.user_prompt_content` is absent or set to `hash` (Strategy A), the system **shall** append exactly one JSONL entry containing: baseline 4 fields + `event_type: user_prompt` + optional extended fields `prompt_hash` (first 16 hex chars of SHA-256(prompt)), `prompt_len` (byte length), `prompt_lang` (ISO-639-1 code from Unicode-block heuristic). Raw `prompt` text **shall NOT** appear in the JSONL entry.

### REQ-HRN-OBS-007 (Ubiquitous — 5-Layer Safety preserved)

The system **shall** preserve the 5-Layer Safety architecture (REQ-HRN-FND-005, constitution §5) in its current form. The expanded observation surface **shall not** remove, bypass, or weaken any layer (L1 Frozen Guard, L2 Canary Check, L3 Contradiction Detector, L4 Rate Limiter, L5 Human Oversight).

### REQ-HRN-OBS-008 (State-Driven — unified gate)

**While** `learning.enabled` resolves to `false`, ALL four event hooks (PostToolUse, Stop, SubagentStop, UserPromptSubmit) **shall** be complete no-ops: each handler **shall not** read, write, or append to `.moai/harness/usage-log.jsonl` nor invoke any tier-classification logic. Disabling **shall not** delete existing entries. Shared gate function `isHarnessLearningEnabled` from SPEC-V3R4-HARNESS-001.

### REQ-HRN-OBS-009 (Ubiquitous — schema additivity)

The system **shall** preserve the four baseline fields (`timestamp`, `event_type`, `subject`, `context_hash`) from REQ-HRN-FND-010 in every JSONL entry produced by any of the four event hooks. New fields **shall** be Go-struct-tagged `omitempty`. Pre-existing PostToolUse-only entries **shall** remain valid; no migration occurs.

### REQ-HRN-OBS-010 (Ubiquitous — 4-tier ladder preserved)

The system **shall** preserve the existing 4-tier ladder with thresholds 1 / 3 / 5 / 10 (Observation / Heuristic / Rule / Auto-update) per REQ-HRN-FND-011. Classifier replacement is deferred to SPEC-V3R4-HARNESS-003.

### REQ-HRN-OBS-011 (Ubiquitous — subagent AskUserQuestion prohibition preserved)

The system **shall** enforce orchestrator-only AskUserQuestion (REQ-HRN-FND-015, agent-common-protocol §User Interaction Boundary) across every new handler. New event handlers (`runHarnessObserveStop`, `runHarnessObserveSubagentStop`, `runHarnessObserveUserPromptSubmit`) **shall not** invoke `AskUserQuestion` and **shall not** spawn subagents that invoke it. No new subagent definitions are introduced.

### REQ-HRN-OBS-012 (Ubiquitous — PII default Strategy A)

The system **shall** default to Strategy A (SHA-256 hash + length + language heuristic) for UserPromptSubmit observation. Raw `prompt` content **shall not** be recorded under default configuration. Strategy A applies when `learning.user_prompt_content` is absent in `harness.yaml`.

### REQ-HRN-OBS-013 (Optional Feature — UserPromptSubmit opt-in)

**Where** `learning.user_prompt_content` is set to `preview`, the system **shall** record the first 64 bytes as `prompt_preview` (Strategy B) in addition to Strategy A fields.

**Where** set to `full`, the system **shall** record the entire prompt as `prompt_content` (Strategy C).

**Where** set to `none`, the UserPromptSubmit hook **shall** be a complete no-op for the observer (gate from REQ-HRN-OBS-008 unaffected).

### REQ-HRN-OBS-014 (Unwanted — fail-open invariant)

**If** `learning.user_prompt_content` contains a value not in `{hash, preview, full, none}`, **then** the system **shall** fall back to Strategy A and **shall not** record raw content under any circumstance. Fail-open to strongest privacy.

### REQ-HRN-OBS-015 (Ubiquitous — EventType enum extension)

The system **shall** extend `EventType` in `internal/harness/types.go` with 3 new SEMANTIC values: `session_stop`, `subagent_stop`, `user_prompt`. Existing 4 values (`moai_subcommand`, `agent_invocation`, `spec_reference`, `feedback`) **shall** remain unchanged.

### REQ-HRN-OBS-016 (Ubiquitous — wrapper script template-first)

The system **shall** ship 3 hook wrapper scripts authored first under `internal/template/templates/.claude/hooks/moai/` as `.sh.tmpl` files. Wrappers **shall** follow the existing 33-line pattern: stderr log path, log rotation at 10MB, binary search order (PATH → `~/go/bin/moai` → `~/.local/bin/moai`), `exec moai hook <subcommand>`, silent exit 0 if binary not found.

### REQ-HRN-OBS-017 (Ubiquitous — settings.json.tmpl additive registration)

The system **shall** register the 3 new wrapper scripts in `internal/template/templates/.claude/settings.json.tmpl` as ADDITIVE entries under `Stop`, `SubagentStop`, `UserPromptSubmit` slots. Existing entries (`handle-stop.sh`, `handle-subagent-stop.sh`, `handle-user-prompt-submit.sh`) **shall** remain unchanged. New entries **shall** carry `timeout: 5`, `type: command`, platform-conditional path quoting matching existing pattern.

### REQ-HRN-OBS-018 (Unwanted — latency budget)

**If** any of the four observer handlers exceeds 100ms wall-clock per invocation under typical workload (single `Observer.RecordEvent`, no retention pruning), **then** the implementation **shall** be out of scope for direct merge and **shall** require a follow-up optimization SPEC. Budget mirrors REQ-HL-001.

---

## Acceptance Coverage Map

Full Given-When-Then scenarios in `acceptance.md`. Coverage map below ensures every REQ has at least one AC.

| AC ID | Covers REQ IDs |
|-------|----------------|
| AC-HRN-OBS-001 (.a/.b/.c) | REQ-HRN-OBS-001, 002, 003 |
| AC-HRN-OBS-002 | REQ-HRN-OBS-004 |
| AC-HRN-OBS-003 | REQ-HRN-OBS-005 |
| AC-HRN-OBS-004 | REQ-HRN-OBS-006 |
| AC-HRN-OBS-005 | REQ-HRN-OBS-008 |
| AC-HRN-OBS-006 | REQ-HRN-OBS-009 |
| AC-HRN-OBS-007 | REQ-HRN-OBS-012 |
| AC-HRN-OBS-008 (.a/.b/.c) | REQ-HRN-OBS-013 |
| AC-HRN-OBS-009 | REQ-HRN-OBS-014 |
| AC-HRN-OBS-010 | REQ-HRN-OBS-015 |
| AC-HRN-OBS-011 (.a/.b) | REQ-HRN-OBS-016, 017 |
| AC-HRN-OBS-012 | REQ-HRN-OBS-007, 010, 011 |
| AC-HRN-OBS-013 | REQ-HRN-OBS-018 |

Coverage: 18 REQs ↔ 13 AC parents (21 leaves with hierarchical children), every REQ covered.

---

## Exclusions (What NOT to Build)

[HARD] Building any of the following within this PR is a scope violation:

1. Any embedding model, embedding-cluster algorithm, or vector index — deferred to SPEC-V3R4-HARNESS-003.
2. Any Reflexion-style self-critique loop, episodic memory, or iteration-cap mechanism — deferred to SPEC-V3R4-HARNESS-004.
3. Any principle-based scoring rubric or constitution-parsing logic — deferred to SPEC-V3R4-HARNESS-005.
4. Any multi-objective scoring tuple or auto-rollback-on-regression mechanism — deferred to SPEC-V3R4-HARNESS-006.
5. Any embedding-indexed skill library or top-K retrieval — deferred to SPEC-V3R4-HARNESS-007.
6. Any cross-project lesson sharing, anonymization layer, or federation — deferred to SPEC-V3R4-HARNESS-008.
7. Migration tooling that converts pre-existing `.moai/harness/usage-log.jsonl` entries into the new schema.
8. Modification of `.claude/rules/moai/design/constitution.md` (FROZEN file).
9. Retroactive wiring of `handle-harness-observe.sh` into PostToolUse slot of `settings.json.tmpl` (V3R4-001 follow-up gap, separately tracked).
10. Modification of `.claude/agents/moai/**`, `.claude/skills/moai-*/**`, `.claude/rules/moai/**`, or `.moai/project/brand/**` (FROZEN per REQ-HRN-FND-006).
11. Modification of `SchemaVersion` field on `Event` struct. Additive optional fields do not require a version bump; field remains `v1`.
12. New AskUserQuestion call sites in any subagent (preserved REQ-HRN-FND-015).
13. Network call, telemetry upload, or external API integration.

---

End of spec-compact.md (auto-generated from spec.md v0.2.1).
