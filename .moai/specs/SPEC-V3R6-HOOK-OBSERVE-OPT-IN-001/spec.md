---
id: SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001
title: "Observability hook 3계열 opt-in 전환 (TaskCreated + Notification + harness-observe)"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: GOOS행님
priority: P2
phase: "v3.0.0"
module: ".moai/config/sections, internal/hook, internal/template/templates"
lifecycle: spec-anchored
tags: "hook, observability, opt-in, config, cost-optimization, sprint-2, v3.0"
tier: S
issue_number: null
depends_on: []
related_specs: [SPEC-V3R6-HOOK-ASYNC-EXPAND-001, SPEC-V3R6-HOOK-CONTRACT-FIX-001]
---

# SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001: Observability Hook 3계열 Opt-In 전환

## Section A — Source of Truth (Design Doc Verbatim)

From `.moai/research/v3.0-design-2026-05-22.md` § 4 Layer 4 first item (lines 245-247):

> **Observe Hook 기본 비활성화**:
> - `TaskCreated`, `Notification`, `handle-harness-observe-*` 3계열 → opt-in (`.moai/config/sections/observability.yaml` flag)
> - 효과: 매 turn당 hook 호출 30% 감소

Baseline state from `.moai/research/moai-adk-current-state-2026-05-22.md` § 6.3 (lines 304-310):

> `Stop`/`SubagentStop`/`UserPromptSubmit`은 **2개 wrapper 병렬 실행**:
> 1. 표준 wrapper (`handle-stop.sh`) — 품질 gate, blocking 가능
> 2. Harness observe (`handle-harness-observe-stop.sh`) — 학습 신호 수집, non-blocking
>
> 이 분리는 **차단 로직과 관찰 로직의 결합도 분리**가 목적이지만, hook 호출 횟수 2배 발생.

And § 6.2 hook event table (line 302):

> `TaskCreated/Notification` — 관찰성 tap (opt-in)

This SPEC operationalizes the opt-in transition by introducing a single boolean toggle (`observability.enabled`) that gates all 3 hook series.

## Section B — Goal & KPI

**Goal**: Convert observability-only hooks from always-active to opt-in, defaulting to disabled, to reduce per-turn hook invocation overhead.

**KPI**: Hook call count per session decreases by at least 25% when `observability.enabled: false` (default), measured against the pre-change baseline. Design doc target is 30%; the 25% threshold provides a 5pp implementation margin.

**Secondary KPI**: Zero functional regression in the orchestrator → existing telemetry consumers MUST continue to receive identical payloads when the flag is flipped to `true`.

## Section C — Requirements (GEARS Notation) & Acceptance Criteria

### Scope (IN SCOPE — exactly 3 hook series)

1. **`TaskCreated`** event hook (currently always-active, configured in `settings.json` `hooks.TaskCreated`)
2. **`Notification`** event hook (currently always-active, configured in `settings.json` `hooks.Notification`)
3. **`handle-harness-observe-*` wrapper trio** — the secondary wrappers attached to `Stop`, `SubagentStop`, `UserPromptSubmit` (currently always-active alongside the primary `handle-stop.sh` / `handle-subagent-stop.sh` / `handle-user-prompt-submit.sh` wrappers)

These 3 series are controlled by a single boolean: `.moai/config/sections/observability.yaml` `enabled:` (default `false`).

### Requirements (REQ-HOI-001..005)

- **REQ-HOI-001 (Ubiquitous)**: The system shall expose a configuration file at `.moai/config/sections/observability.yaml` with a top-level `enabled:` boolean key, whose default value is `false`.

- **REQ-HOI-002 (When)**: When `observability.enabled == false`, the orchestrator shall not execute any of the 3 hook series defined in scope (`TaskCreated`, `Notification`, `handle-harness-observe-*`).

- **REQ-HOI-003 (When)**: When `observability.enabled == true`, the orchestrator shall execute all 3 hook series with their existing payload schema and side effects unchanged.

- **REQ-HOI-004 (Where)**: Where `moai init` runs without an explicit `--observability` flag, the system shall render `observability.yaml` with `enabled: false` as the default initial state.

- **REQ-HOI-005 (Ubiquitous)**: The `moai doctor` command shall report current observability status with a line matching the pattern `Observability:.*(enabled|disabled)` in its diagnostic output.

### Binary Acceptance Criteria (AC-HOI-001..007)

- **AC-HOI-001**: `.moai/config/sections/observability.yaml` exists with `enabled:` key. Verification: `grep -E '^enabled:' .moai/config/sections/observability.yaml` exits 0 with at least one match.

- **AC-HOI-002**: `moai init` in a clean directory (`/tmp/test-init-$(date +%s)`) produces `observability.yaml` with `enabled: false`. Verification: post-`moai init`, `grep -E '^enabled:\s*false' <project>/.moai/config/sections/observability.yaml` exits 0.

- **AC-HOI-003**: Integration test verifies that with `enabled: false`, calls to `TaskCreated` / `Notification` / `handle-harness-observe-*` are SKIPPED — no JSONL append, no log line, no telemetry side effect. Verification: `go test ./internal/hook/... -run TestObservabilityDisabled` exits 0.

- **AC-HOI-004**: Integration test verifies that with `enabled: true`, the same 3 hooks DO execute and produce expected outputs (one JSONL entry per event, payload schema unchanged from baseline). Verification: `go test ./internal/hook/... -run TestObservabilityEnabled` exits 0.

- **AC-HOI-005**: Rendered `settings.json` conditionally includes the 3 hook entries based on flag. Verification: after `moai init` with `observability.enabled: false`, `grep -c handle-harness-observe-stop <project>/.claude/settings.json` returns `0`; after toggling to `true` and running `moai update`, the same grep returns `>=1`.

- **AC-HOI-006**: `moai doctor` output includes an observability status line. Verification: `moai doctor 2>&1 | grep -E 'Observability:\s*(enabled|disabled)'` exits 0.

- **AC-HOI-007**: Template mirror matches rendered output. Verification: `internal/template/templates/.moai/config/sections/observability.yaml` exists with `enabled: false`, and after `moai init` the rendered file is byte-equivalent (modulo template variable expansion).

### Traceability Matrix (REQ ↔ AC, 100% coverage)

| Requirement | Mapped ACs | Coverage |
|---|---|---|
| REQ-HOI-001 | AC-HOI-001, AC-HOI-007 | YAML file exists with `enabled:` key + template mirror parity |
| REQ-HOI-002 | AC-HOI-003 | Integration test proves SKIP path when disabled |
| REQ-HOI-003 | AC-HOI-004 | Integration test proves EXECUTE path when enabled, payload unchanged |
| REQ-HOI-004 | AC-HOI-002, AC-HOI-005 | `moai init` default + conditional `settings.json` render |
| REQ-HOI-005 | AC-HOI-006 | Doctor diagnostic line |

All 5 REQs map to at least one AC; all 7 ACs trace to at least one REQ. No orphans.

## Section D — Implementation Approach (See plan.md for milestones)

Three milestones (M1-M3) cover schema, loader gating, and CLI/doctor surface. See `plan.md` § Milestones for the ordered work breakdown.

## Section E — Risks & Exclusions

### Risks (R1-R4)

- **R1 (Likelihood: Low / Impact: Medium)**: Existing users with harness learning pipelines depending on observability telemetry experience silent data loss after upgrade. Mitigation: opt-in default `false` is intentional shift; v3.0 major version bump signals breaking change. Release notes + migration guide (Wave 5 docs SPEC) MUST call out the toggle explicitly. Users requiring the prior behavior set `enabled: true` post-upgrade.

- **R2 (Likelihood: Medium / Impact: Low)**: Cross-SPEC conflict with sibling `SPEC-V3R6-HOOK-ASYNC-EXPAND-001` (Sprint 2 partner) — both touch `TaskCreated` and `Notification` hook stanzas. If `HOOK-OBSERVE-OPT-IN-001` opts the hooks out, the async-transition work in `HOOK-ASYNC-EXPAND-001` becomes conditional. **Resolution**: this SPEC (HOI) merges first; `HOOK-ASYNC-EXPAND-001` then makes its `TaskCreated`/`Notification` async stanzas conditional on `observability.enabled == true`. See plan.md § Cross-SPEC Coordination for the ordering contract.

- **R3 (Likelihood: Low / Impact: Low)**: `moai doctor` regression on legacy projects that lack `observability.yaml`. Mitigation: doctor MUST default to reporting `disabled` when the file is absent — never error.

- **R4 (Likelihood: Low / Impact: Low)**: Test fixtures across `internal/hook/...` expect always-active observability and break after the gating logic lands. Mitigation: M3 milestone explicitly enumerates fixture updates and the integration tests are added in M3 alongside the gating change.

## Out of Scope

### Out of Scope: PostToolUse / SessionStart / PreToolUse opt-in

- `PostToolUse` / `SessionStart` / `PreToolUse` STAY always-active — they carry MX validation + LSP gating (PostToolUse), GLM setup + skill discovery + memory load (SessionStart), and security blocker (PreToolUse) responsibilities. All three are synchronous critical-path hooks; making them opt-in requires a separate SPEC with its own threat model.

`PostToolUse` carries MX validation and LSP gating responsibilities; `SessionStart` carries GLM setup, skill discovery, and memory load responsibilities; `PreToolUse` carries security blocker responsibilities. All three are synchronous critical-path hooks. Making them opt-in is explicitly outside this SPEC's scope — they STAY always-active. Any future opt-in proposal for these hooks requires a separate SPEC with its own threat model.

### Out of Scope: Async expansion of observability hooks

- Async transition of the same 3 hook series is owned by sibling `SPEC-V3R6-HOOK-ASYNC-EXPAND-001` (Sprint 2 partner) — this SPEC handles ON/OFF toggling only. The opt-in flag and the async transition are independent dimensions.

Async transition of the same 3 hook series (changing them from synchronous to background execution) is the responsibility of sibling SPEC `SPEC-V3R6-HOOK-ASYNC-EXPAND-001` (Sprint 2 partner). This SPEC handles ON/OFF toggling only. The opt-in flag and the async transition are independent dimensions: when `enabled: true`, the hooks may run sync (current behavior) or async (after HOOK-ASYNC-EXPAND-001 merges). HOI does not pre-empt ASYNC-EXPAND's design space.

### Out of Scope: Telemetry data schema redesign

- JSONL payload format emitted by observability hooks (when `enabled: true`) stays unchanged in this SPEC — field names, ordering, semantics preserved exactly. Schema evolution requires a separate SPEC and a backward-compatibility plan for harness consumers.

The JSONL payload format emitted by observability hooks (when `enabled: true`) stays unchanged in this SPEC. Field names, ordering, and semantics are preserved exactly. Only the ON/OFF toggle is in scope. Schema evolution of telemetry payloads requires a separate SPEC and a backward-compatibility plan for harness consumers.

## Cross-References

- **Sibling (R2 partner)**: `SPEC-V3R6-HOOK-ASYNC-EXPAND-001` — Sprint 2 async-expansion of the same 3 hook series. Merges after this SPEC.
- **Sprint 2 independents**: `SPEC-V3R6-AGENT-MODEL-ROUTING-001`, `SPEC-V3R6-PROMPT-CACHE-001` — no overlap with this SPEC.
- **Wave 0 regression guard**: `SPEC-V3R6-HOOK-CONTRACT-FIX-001` — fixes `WorktreeCreate` hook contract; orthogonal to this SPEC but shares the `internal/hook/` module surface.
- **Design doc**: `.moai/research/v3.0-design-2026-05-22.md` § 4 Layer 4.
- **Baseline**: `.moai/research/moai-adk-current-state-2026-05-22.md` § 6.2, § 6.3, § 6.4.
- **Frontmatter SSOT**: `.claude/rules/moai/development/spec-frontmatter-schema.md`.
- **GEARS notation**: PR #1046 merged 2026-05-22 — REQs in this SPEC use GEARS WHEN/WHERE notation (zero legacy IF/THEN).
- **Sprint naming SSOT**: `.claude/rules/moai/development/sprint-round-naming.md` (v2.0.0; legacy filename `sprint-wave-naming.md` retired per AP-SRN-004) — this SPEC is one of 4 in Sprint 2.
