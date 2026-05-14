---
id: SPEC-V3R4-HARNESS-002
version: "0.1.0"
status: draft
created: 2026-05-14
updated: 2026-05-14
author: manager-spec
priority: P1
tags: "harness, self-evolution, multi-event-observer, v3r4, downstream, observer-expansion"
issue_number: null
title: Multi-Event Observer Expansion (Stop / SubagentStop / UserPromptSubmit hook integration)
phase: "v3.0.0 R4 — Phase B — Observer Expansion"
module: ".claude/hooks/moai/, internal/cli/hook.go, internal/harness/observer.go, .moai/harness/usage-log.jsonl (schema extension)"
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

> **STATUS: DRAFT — plan-phase entry seed only.** Full Plan workflow (research.md, acceptance.md, plan.md, tasks.md, EARS-format REQ enumeration, plan-auditor cycle) to be executed in a follow-up `/moai plan SPEC-V3R4-HARNESS-002` session per the spec-workflow.md Plan Phase contract. This file establishes the directory, frontmatter, and goal context so the next plan session has a starting point.

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-05-14 | manager-spec | Initial plan-phase entry seed. Created as the first downstream SPEC after SPEC-V3R4-HARNESS-001 (foundation) merged. Scope: extend the PostToolUse-only observer baseline established by V3R4-HARNESS-001 to cover Stop, SubagentStop, and UserPromptSubmit Claude Code hook events with a unified observation schema. Detailed EARS-format requirements and acceptance criteria are reserved for the next plan-phase session. |

---

## 1. Goal

The V3R4 self-evolving harness observation surface MUST expand from the current PostToolUse-only baseline (per `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-010) to a multi-event observer covering Claude Code's `Stop`, `SubagentStop`, and `UserPromptSubmit` hook events in addition to `PostToolUse`. All four event types MUST share a unified observation schema (extended JSONL append-only format under `.moai/harness/usage-log.jsonl`) so downstream classifiers (frequency-count today, embedding-cluster in `SPEC-V3R4-HARNESS-003`) can aggregate cross-event patterns without per-event schema branching. The `learning.enabled` gate established in `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-009 MUST extend to ALL new event hooks (no-op when disabled).

### 1.1 Background

- `SPEC-V3R4-HARNESS-001` shipped the PostToolUse-only observer baseline with a 4-field schema (`timestamp` / `event_type` / `subject` / `context_hash`). It explicitly deferred multi-event expansion to this SPEC (`spec.md` §1.3 Non-Goals item 1).
- Claude Code v2.1.110+ exposes four primary lifecycle hooks the harness can observe: `PostToolUse` (already wired), `Stop` (session end signal), `SubagentStop` (Agent() teammate exit), `UserPromptSubmit` (user input received). Each provides distinct learning signals: PostToolUse → tool usage frequency, Stop → session-level completion patterns, SubagentStop → agent invocation outcomes, UserPromptSubmit → user intent patterns.
- The existing observer code path (`internal/cli/hook.go::runHarnessObserve` + `internal/harness/observer.go`) was designed for a single event type. Multi-event support requires schema extension (event-specific payload fields) and routing logic for the additional `moai hook` subcommands.

### 1.2 Scope (preliminary — full enumeration in next plan session)

**In Scope (preliminary)**:

- Add `moai hook stop`, `moai hook subagent-stop`, `moai hook user-prompt-submit` cobra subcommands that wire the harness observer for the respective events.
- Extend `.moai/harness/usage-log.jsonl` schema with optional per-event fields while preserving the 4 baseline fields from REQ-HRN-FND-010.
- Extend `isHarnessLearningEnabled` gate (REQ-HRN-FND-009) to apply to all new event hooks identically.
- Extend `handle-*-observe.sh` Claude Code hook wrappers to route to the corresponding `moai hook` subcommand for each event type.
- Update `internal/template/templates/.claude/settings.json.tmpl` to register the new hook handlers.

**Out of Scope (preliminary — deferred to downstream V3R4 SPECs)**:

- Embedding-cluster classifier replacing frequency-count tier ladder — `SPEC-V3R4-HARNESS-003`.
- Reflexion self-critique loop — `SPEC-V3R4-HARNESS-004`.
- Principle-based scoring — `SPEC-V3R4-HARNESS-005`.

### 1.3 Non-Goals

This SPEC is the OBSERVER EXPANSION. The following capabilities are explicitly OUT OF SCOPE and are deferred to the named downstream SPECs:

- Frequency-count classifier replacement — deferred to `SPEC-V3R4-HARNESS-003`.
- Verbal self-critique loop — deferred to `SPEC-V3R4-HARNESS-004`.
- Principle-based scoring rubric — deferred to `SPEC-V3R4-HARNESS-005`.
- Multi-objective scoring tuple — deferred to `SPEC-V3R4-HARNESS-006`.
- Skill library auto-organization — deferred to `SPEC-V3R4-HARNESS-007`.
- Cross-project federation — deferred to `SPEC-V3R4-HARNESS-008`.
- Any migration of historical `usage-log.jsonl` entries that lack the extended schema. Old entries remain valid under the baseline 4-field contract from REQ-HRN-FND-010; new entries MAY include extended fields.

---

## 2. Stakeholders (preliminary — full enumeration in next plan session)

| Role | Interest |
|------|----------|
| MoAI-ADK maintainer | Single source of truth for harness observation surface; uniform `learning.enabled` gate enforcement across all four event types. |
| Solo developer / power user | Richer learning signals without manual intervention; preserved per-project no-op via `learning.enabled: false`. |
| `manager-spec` (this SPEC's author for downstream sessions) | Clean enumeration of REQ IDs and acceptance criteria for plan-auditor cycle. |
| `evaluator-active` | Unchanged binding-gate role; new observation entries feed downstream classifier and Tier-4 proposal generation. |

---

## 3. Requirements (preliminary — placeholder)

> **NOTE**: Full EARS-format REQ enumeration is reserved for the next `/moai plan` session. The IDs reserved so far:

- `REQ-HRN-OBS-001` ~ `REQ-HRN-OBS-NNN` (to be assigned during plan workflow).

The plan-auditor cycle MUST verify:
- Each new event type has at least one REQ describing observer behavior under `learning.enabled: true` and one describing no-op under `false`.
- The unified schema's extended fields are explicitly documented per event type.
- `SPEC-V3R4-HARNESS-001` REQ-HRN-FND-005 (5-Layer Safety preservation) and REQ-HRN-FND-009 (observer no-op gate) MUST remain intact (no weakening).

---

## 4. Acceptance Criteria (preliminary — placeholder)

> Full Given-When-Then enumeration reserved for next plan session. Skeleton:

- `AC-HRN-OBS-001` — when `learning.enabled: true` and a `Stop` event fires, the observer appends one JSONL entry with the baseline 4 fields + the event-specific extended fields.
- `AC-HRN-OBS-002` — when `learning.enabled: false`, ALL four event hooks become complete no-ops (preserves AC-HRN-FND-007 semantics from V3R4-HARNESS-001).
- `AC-HRN-OBS-003` — the new cobra subcommands (`stop`, `subagent-stop`, `user-prompt-submit`) MUST exist on `hookCmd.Commands()` and be invokable from the corresponding `handle-*-observe.sh` wrapper.
- (full set to be enumerated in next plan session)

---

## 5. Dependencies

| SPEC | Relationship | Notes |
|------|--------------|-------|
| `SPEC-V3R4-HARNESS-001` | Hard dependency (foundation) | Establishes the PostToolUse observer baseline, `learning.enabled` gate, FROZEN zone, 5-Layer Safety, and `usage-log.jsonl` schema this SPEC extends. The 4-tier ladder (REQ-HRN-FND-011) is preserved unchanged by this SPEC. |
| `SPEC-V3R4-HARNESS-003` | Downstream (blocked by this SPEC) | Embedding-cluster classifier consumes the multi-event observation stream produced by this SPEC. |

---

## 6. References

- `SPEC-V3R4-HARNESS-001/spec.md` §1.3 Non-Goals item 1 (this SPEC's scope source).
- `SPEC-V3R4-HARNESS-001/spec.md` REQ-HRN-FND-009 / REQ-HRN-FND-010 (gate + baseline schema preserved here).
- `internal/cli/hook.go::runHarnessObserve` (baseline implementation; extends here).
- `internal/cli/hook.go::isHarnessLearningEnabled` (gate function; reused here).
- `internal/cli/hook_harness_observe_test.go` (baseline test suite; extends here).
- Claude Code hook events: PostToolUse / Stop / SubagentStop / UserPromptSubmit (Anthropic Claude Code v2.1.110+ documentation).

---

> **Next plan-phase actions** (reserved for next session):
>
> 1. `/moai plan SPEC-V3R4-HARNESS-002` invokes `manager-spec` to perform research (research.md), expand §2-§5 above with full EARS-format REQ enumeration, write `acceptance.md` with Given-When-Then scenarios, `plan.md` (implementation strategy), `tasks.md` (Wave breakdown).
> 2. `plan-auditor` runs the iterative audit cycle (1-6 iterations) on the four artifacts.
> 3. plan PR squash-merged into main under conventional commit `plan(SPEC-V3R4-HARNESS-002): ...`.
