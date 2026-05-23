---
id: SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001
title: "Observability hook 3계열 opt-in 전환 (TaskCreated + Notification + harness-observe)"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: Author Name
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

## Section A — Pre-existing State Survey (codebase audit prior to scoping)

This section enumerates pre-existing infrastructure facts discovered during plan-phase revise (2026-05-23) after the first plan-auditor PASS missed a key-name semantic collision. Every fact below was verified against working-tree HEAD `ef95739c6` on `main`.

### A.1 — Pre-existing infrastructure facts (5)

1. **`.moai/config/sections/observability.yaml` exists with `observability.enabled: true`** (10 lines total). The `enabled:` key here is the master toggle for **trace-logging + hook_metrics** infrastructure (REQ-OBS-005 from a prior SPEC, inferred). The file also defines `hook_metrics.output_path`, `hook_metrics.slow_hook_threshold_ms`, `max_file_size_mb`, `report_dir`, `retention_days`, `trace_dir`. This file is NOT empty and NOT new — it has shipping semantics.

2. **`.moai/config/sections/system.yaml` `hook.observability_events: []`** with explicit header comment `# Hook observability configuration (SPEC-V3R2-RT-006 REQ-004)`. The key is a per-event RETIRE-OBS-ONLY whitelist (default empty = all retired events silently no-op). Sibling key `hook.strict_mode: false` controls strict-mode behavior for retired events.

3. **`internal/hook/observability.go`** (44 lines) implements `observabilityOptIn(cfg ConfigProvider, eventName string) bool` which reads `cfg.Get().System.Hook.ObservabilityEvents` (the per-event list from §A.1.2). Annotated `@MX:ANCHOR observabilityOptIn guards all RETIRE-OBS-ONLY handler entry paths` with `@MX:REASON: fan_in=4, called by notification/elicitation/elicitationResult/taskCreated handlers`. Pattern A semantics: callers MUST silently return `HookOutput{}` when this returns false.

4. **`internal/hook/coverage_table.go`** declares `ResolutionRetireObsOnly` + `ObservabilityOptIn` cohort indexing fields for RT-006 (verified by orchestrator via grep; not re-read in full).

5. **`internal/hook/audit_test.go`** contains `TestAuditObservabilityWhitelist` (T-RT006-02 RED + T-RT006-08 GREEN) covering `observabilityOptIn` behavior with whitelist semantics + case-insensitivity (verified by orchestrator via grep; not re-read in full).

### A.2 — Collision discovered post-first-PASS

The original draft of this SPEC (plan-auditor iter-1 PASS 0.8775, Tier S 0.75 +0.1275) introduced `observability.enabled` in `.moai/config/sections/observability.yaml` as the HOI master toggle. This key name is **byte-identical** to the pre-existing key in §A.1.1 but carries **different semantics** (HOI = 3 hook series gating; REQ-OBS-005 = trace-logging master switch). Plan-auditor did not detect this because its audit scope is artifact self-consistency, not codebase pre-existing state survey.

### A.3 — Collision-zero verification (3-key cohabitation contract)

Per user decision Q4=A (2026-05-23), the HOI master toggle is relocated to a NEW key under existing `hook:` block in `system.yaml`:

| Key | File | Owner SPEC | Semantics | Default |
|---|---|---|---|---|
| `observability.enabled` | `.moai/config/sections/observability.yaml` | REQ-OBS-005 (pre-existing, untouched) | Trace-logging + hook_metrics master switch | `true` (unchanged) |
| `hook.observability_events` | `.moai/config/sections/system.yaml` | SPEC-V3R2-RT-006 REQ-040 (pre-existing, untouched) | Per-event RETIRE-OBS-ONLY whitelist | `[]` (unchanged) |
| **`hook.opt_in.enabled`** | `.moai/config/sections/system.yaml` | **SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 (this SPEC, NEW)** | 3 hook series (TaskCreated + Notification + handle-harness-observe-*) opt-in master toggle | `false` (NEW) |

All 3 keys are **independent** — no read/write overlap, no shared loader path, no test fixture conflict. HOI's runtime gating reads ONLY `cfg.System.Hook.OptIn.Enabled`. REQ-OBS-005's trace logger reads ONLY `cfg.Observability.Enabled`. RT-006's `observabilityOptIn(cfg, eventName)` reads ONLY `cfg.System.Hook.ObservabilityEvents`. Cohabitation is safe.

### A.4 — Mitigation contract (plan-auditor codebase state blindspot)

This §A section implements mitigation #1 from the 2026-05-23 lesson on plan-auditor's pre-existing state blindspot. Every future SPEC scoped against `.moai/config/sections/*.yaml` SHOULD include a §A "Pre-existing State Survey" enumerating shipping keys that share or neighbor the proposed key namespace. Mitigations #2/#3/#4 (manager-develop Section C grep extension, plan-auditor checklist extension, state-snapshot.md option) are out of scope for this revise and tracked as separate rule-revise items.

## Section B — Source of Truth (Design Doc Verbatim)

From `.moai/research/v3.0-design-2026-05-22.md` § 4 Layer 4 first item (lines 245-247):

> **Observe Hook 기본 비활성화**:
> - `TaskCreated`, `Notification`, `handle-harness-observe-*` 3계열 → opt-in (`.moai/config/sections/observability.yaml` flag)
> - 효과: 매 turn당 hook 호출 30% 감소

**Revise note (2026-05-23)**: The design doc cites `observability.yaml flag` as the toggle location. Per §A.2 collision discovery and user Q4=A decision, the toggle location is **relocated** to `system.yaml` `hook.opt_in.enabled`. The design doc text is retained verbatim above for traceability; the implementation diverges in toggle location only (semantics + 30% reduction target unchanged).

Baseline state from `.moai/research/moai-adk-current-state-2026-05-22.md` § 6.3 (lines 304-310):

> `Stop`/`SubagentStop`/`UserPromptSubmit`은 **2개 wrapper 병렬 실행**:
> 1. 표준 wrapper (`handle-stop.sh`) — 품질 gate, blocking 가능
> 2. Harness observe (`handle-harness-observe-stop.sh`) — 학습 신호 수집, non-blocking
>
> 이 분리는 **차단 로직과 관찰 로직의 결합도 분리**가 목적이지만, hook 호출 횟수 2배 발생.

And § 6.2 hook event table (line 302):

> `TaskCreated/Notification` — 관찰성 tap (opt-in)

This SPEC operationalizes the opt-in transition by introducing a single boolean toggle (`hook.opt_in.enabled` in `system.yaml`, per §A.3) that gates all 3 hook series.

## Section C — Goal & KPI

**Goal**: Convert observability-only hooks from always-active to opt-in, defaulting to disabled, to reduce per-turn hook invocation overhead.

**KPI**: Hook call count per session decreases by at least 25% when `hook.opt_in.enabled: false` (default), measured against the pre-change baseline. Design doc target is 30%; the 25% threshold provides a 5pp implementation margin.

**Secondary KPI**: Zero functional regression in the orchestrator → existing telemetry consumers MUST continue to receive identical payloads when the flag is flipped to `true`. Additionally, REQ-OBS-005 (`observability.enabled`) trace logging MUST continue to function unchanged regardless of `hook.opt_in.enabled` value (cohabitation invariant).

## Section D — Requirements (GEARS Notation) & Acceptance Criteria

### Scope (IN SCOPE — exactly 3 hook series)

1. **`TaskCreated`** event hook (currently always-active, configured in `settings.json` `hooks.TaskCreated`)
2. **`Notification`** event hook (currently always-active, configured in `settings.json` `hooks.Notification`)
3. **`handle-harness-observe-*` wrapper trio** — the secondary wrappers attached to `Stop`, `SubagentStop`, `UserPromptSubmit` (currently always-active alongside the primary `handle-stop.sh` / `handle-subagent-stop.sh` / `handle-user-prompt-submit.sh` wrappers)

These 3 series are controlled by a single boolean: `.moai/config/sections/system.yaml` `hook.opt_in.enabled:` (default `false`).

### Requirements (REQ-HOI-001..005)

- **REQ-HOI-001 (Ubiquitous)**: The system shall expose a configuration key at `.moai/config/sections/system.yaml` under the existing `hook:` block as `hook.opt_in.enabled`, a boolean whose default value is `false`. This key is NEW (per §A.3 collision-zero contract) and MUST NOT collide with pre-existing `observability.enabled` (REQ-OBS-005, different file + different semantics) or `hook.observability_events` (SPEC-V3R2-RT-006 REQ-040, same file but different semantics — per-event whitelist).

- **REQ-HOI-002 (When)**: When `hook.opt_in.enabled == false`, the orchestrator shall not execute any of the 3 hook series defined in scope (`TaskCreated`, `Notification`, `handle-harness-observe-*`).

- **REQ-HOI-003 (When)**: When `hook.opt_in.enabled == true`, the orchestrator shall execute all 3 hook series with their existing payload schema and side effects unchanged.

- **REQ-HOI-004 (Where)**: Where `moai init` runs without an explicit `--hook-opt-in` flag, the system shall render `system.yaml` with `hook.opt_in.enabled: false` as the default initial state.

- **REQ-HOI-005 (Ubiquitous)**: The `moai doctor` command shall report current hook opt-in status with a line matching the pattern `Hook opt-in:\s*(enabled|disabled)` in its diagnostic output. This is distinct from any existing `Observability:` line that REQ-OBS-005 may already produce (cohabitation invariant).

### Binary Acceptance Criteria (AC-HOI-001..007)

- **AC-HOI-001**: `.moai/config/sections/system.yaml` contains `hook.opt_in.enabled` key with default `false`. Verification: `grep -A1 '^\s*opt_in:' .moai/config/sections/system.yaml | grep -E 'enabled:\s*false'` exits 0 with at least one match. Pre-existing `hook.observability_events` and `hook.strict_mode` MUST remain present (cohabitation invariant).

- **AC-HOI-002**: `moai init` in a clean directory (`/tmp/test-init-$(date +%s)`) produces `system.yaml` with `hook.opt_in.enabled: false`. Verification: post-`moai init`, the rendered file passes the same grep pattern as AC-HOI-001.

- **AC-HOI-003**: Integration test verifies that with `hook.opt_in.enabled: false`, calls to `TaskCreated` / `Notification` / `handle-harness-observe-*` are SKIPPED — no JSONL append, no log line, no telemetry side effect. Verification: `go test ./internal/hook/... -run TestHookOptInDisabled` exits 0.

- **AC-HOI-004**: Integration test verifies that with `hook.opt_in.enabled: true`, the same 3 hooks DO execute and produce expected outputs (one JSONL entry per event, payload schema unchanged from baseline). Verification: `go test ./internal/hook/... -run TestHookOptInEnabled` exits 0.

- **AC-HOI-005**: Rendered `settings.json` conditionally includes the 3 hook entries based on flag. Verification: after `moai init` with `hook.opt_in.enabled: false`, `grep -c handle-harness-observe-stop <project>/.claude/settings.json` returns `0`; after toggling to `true` and running `moai update`, the same grep returns `>=1`.

- **AC-HOI-006**: `moai doctor` output includes a hook opt-in status line. Verification: `moai doctor 2>&1 | grep -E 'Hook opt-in:\s*(enabled|disabled)'` exits 0.

- **AC-HOI-007**: Cohabitation invariant — `hook.opt_in.enabled` does NOT interfere with REQ-OBS-005 `observability.enabled` (trace logging) or SPEC-V3R2-RT-006 REQ-040 `hook.observability_events` (per-event whitelist). Verification: dedicated integration test `TestHookOptInCohabitation` exercises all 4 quadrants (HOI on/off × OBS on/off) and asserts all 3 systems function independently (HOI gates 3 hook series, OBS gates trace logging, RT-006 gates retired event taps). All 4 quadrants MUST pass with no cross-system side effects.

### Traceability Matrix (REQ ↔ AC, 100% coverage)

| Requirement | Mapped ACs | Coverage |
|---|---|---|
| REQ-HOI-001 | AC-HOI-001, AC-HOI-007 | system.yaml key exists + cohabitation invariant verified |
| REQ-HOI-002 | AC-HOI-003 | Integration test proves SKIP path when disabled |
| REQ-HOI-003 | AC-HOI-004 | Integration test proves EXECUTE path when enabled, payload unchanged |
| REQ-HOI-004 | AC-HOI-002, AC-HOI-005 | `moai init` default + conditional `settings.json` render |
| REQ-HOI-005 | AC-HOI-006 | Doctor diagnostic line |

All 5 REQs map to at least one AC; all 7 ACs trace to at least one REQ. No orphans.

## Section E — Implementation Approach (See plan.md for milestones)

Three milestones (M1-M3) cover system.yaml schema extension, loader gating, and CLI/doctor surface. See `plan.md` § Milestones for the ordered work breakdown and §A.3 collision-zero contract enforcement details.

## Section F — Risks & Exclusions

### Risks (R1-R5)

- **R1 (Likelihood: Low / Impact: Medium)**: Existing users with harness learning pipelines depending on observability telemetry experience silent data loss after upgrade. Mitigation: opt-in default `false` is intentional shift; v3.0 major version bump signals breaking change. Release notes + migration guide (Wave 5 docs SPEC) MUST call out the toggle explicitly. Users requiring the prior behavior set `hook.opt_in.enabled: true` post-upgrade.

- **R2 (Likelihood: Medium / Impact: Low)**: Cross-SPEC conflict with sibling `SPEC-V3R6-HOOK-ASYNC-EXPAND-001` (Sprint 2 partner) — both touch `TaskCreated` and `Notification` hook stanzas. If `HOOK-OBSERVE-OPT-IN-001` opts the hooks out, the async-transition work in `HOOK-ASYNC-EXPAND-001` becomes conditional. **Resolution**: this SPEC (HOI) merges first; `HOOK-ASYNC-EXPAND-001` then makes its `TaskCreated`/`Notification` async stanzas conditional on `hook.opt_in.enabled == true`. See plan.md § Cross-SPEC Coordination for the ordering contract.

- **R3 (Likelihood: Low / Impact: Low)**: `moai doctor` regression on legacy projects that lack the `hook.opt_in` sub-block. Mitigation: doctor MUST default to reporting `disabled` when the key is absent — never error. Loader contract per plan.md §A.3 uses zero-value Go default (`false`) when key is missing.

- **R4 (Likelihood: Low / Impact: Low)**: Test fixtures across `internal/hook/...` expect always-active observability and break after the gating logic lands. Mitigation: M3 milestone explicitly enumerates fixture updates and the integration tests are added in M3 alongside the gating change.

- **R5 (Likelihood: Low / Impact: Medium — NEW post-revise)**: Cohabitation regression — a future refactor unifies `hook.opt_in.enabled` with `observability.enabled` (REQ-OBS-005) and inadvertently breaks one of the two semantics. Mitigation: AC-HOI-007 dedicated cohabitation integration test (`TestHookOptInCohabitation`) is a permanent regression guard; M3 milestone includes a comment in `internal/hook/observability.go` (or new sibling file) cross-referencing §A.3 of this SPEC and warning future maintainers against key unification without a fresh SPEC.

## Out of Scope

### Out of Scope: PostToolUse / SessionStart / PreToolUse opt-in

- `PostToolUse` / `SessionStart` / `PreToolUse` STAY always-active — they carry MX validation + LSP gating (PostToolUse), GLM setup + skill discovery + memory load (SessionStart), and security blocker (PreToolUse) responsibilities. All three are synchronous critical-path hooks; making them opt-in requires a separate SPEC with its own threat model.

`PostToolUse` carries MX validation and LSP gating responsibilities; `SessionStart` carries GLM setup, skill discovery, and memory load responsibilities; `PreToolUse` carries security blocker responsibilities. All three are synchronous critical-path hooks. Making them opt-in is explicitly outside this SPEC's scope — they STAY always-active. Any future opt-in proposal for these hooks requires a separate SPEC with its own threat model.

### Out of Scope: Async expansion of observability hooks

- Async transition of the same 3 hook series is owned by sibling `SPEC-V3R6-HOOK-ASYNC-EXPAND-001` (Sprint 2 partner) — this SPEC handles ON/OFF toggling only. The opt-in flag and the async transition are independent dimensions.

Async transition of the same 3 hook series (changing them from synchronous to background execution) is the responsibility of sibling SPEC `SPEC-V3R6-HOOK-ASYNC-EXPAND-001` (Sprint 2 partner). This SPEC handles ON/OFF toggling only. The opt-in flag and the async transition are independent dimensions: when `hook.opt_in.enabled: true`, the hooks may run sync (current behavior) or async (after HOOK-ASYNC-EXPAND-001 merges). HOI does not pre-empt ASYNC-EXPAND's design space.

### Out of Scope: Telemetry data schema redesign

- JSONL payload format emitted by observability hooks (when `hook.opt_in.enabled: true`) stays unchanged in this SPEC — field names, ordering, semantics preserved exactly. Schema evolution requires a separate SPEC and a backward-compatibility plan for harness consumers.

The JSONL payload format emitted by observability hooks (when `hook.opt_in.enabled: true`) stays unchanged in this SPEC. Field names, ordering, and semantics are preserved exactly. Only the ON/OFF toggle is in scope. Schema evolution of telemetry payloads requires a separate SPEC and a backward-compatibility plan for harness consumers.

### Out of Scope: Migration of REQ-OBS-005 or SPEC-V3R2-RT-006 key namespaces

- The pre-existing `observability.enabled` (REQ-OBS-005) and `hook.observability_events` (SPEC-V3R2-RT-006 REQ-040) keys are PRESERVED untouched. This SPEC does NOT migrate, deprecate, or rename either key. Any future unification of these 3 keys into a single namespace requires a separate SPEC with explicit migration plan + backward-compatibility window. AC-HOI-007 enforces this preservation as a binary acceptance criterion (cohabitation invariant).

## Cross-References

- **Pre-existing state**: §A.1 enumerates 5 codebase facts verified at HEAD `ef95739c6` (2026-05-23). Future SPECs touching `.moai/config/sections/system.yaml` or `internal/hook/observability.go` SHOULD re-verify §A.1 facts at their plan-phase HEAD.
- **Sibling (R2 partner)**: `SPEC-V3R6-HOOK-ASYNC-EXPAND-001` — Sprint 2 async-expansion of the same 3 hook series. Merges after this SPEC.
- **Sprint 2 independents**: `SPEC-V3R6-AGENT-MODEL-ROUTING-001`, `SPEC-V3R6-PROMPT-CACHE-001` — no overlap with this SPEC.
- **Wave 0 regression guard**: `SPEC-V3R6-HOOK-CONTRACT-FIX-001` — fixes `WorktreeCreate` hook contract; orthogonal to this SPEC but shares the `internal/hook/` module surface.
- **Cohabiting SPECs (DO NOT BREAK)**: REQ-OBS-005 (trace logging master toggle, owns `observability.yaml`) + SPEC-V3R2-RT-006 REQ-040 (RETIRE-OBS-ONLY per-event whitelist, owns `system.yaml` `hook.observability_events`). HOI cohabits with both per §A.3 contract.
- **Design doc**: `.moai/research/v3.0-design-2026-05-22.md` § 4 Layer 4 (cited toggle location since superseded by user Q4=A 2026-05-23).
- **Baseline**: `.moai/research/moai-adk-current-state-2026-05-22.md` § 6.2, § 6.3, § 6.4.
- **Frontmatter SSOT**: `.claude/rules/moai/development/spec-frontmatter-schema.md`.
- **GEARS notation**: PR #1046 merged 2026-05-22 — REQs in this SPEC use GEARS WHEN/WHERE notation (zero legacy IF/THEN).
- **Sprint naming SSOT**: `.claude/rules/moai/development/sprint-round-naming.md` (v2.0.0; legacy filename `sprint-wave-naming.md` retired per AP-SRN-004) — this SPEC is one of 4 in Sprint 2.
- **Revise lesson**: 2026-05-23 plan-auditor codebase state blindspot — see progress.md "Revise History" + lesson candidate `feedback-plan-auditor-codebase-state-blindspot` (mitigation #1 implemented in §A of this SPEC).
