---
id: SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001
title: "Observability hook 3계열 opt-in — Implementation Plan (Tier S, M1-M3)"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: Author Name
priority: P2
phase: "v3.0.0"
module: ".moai/config/sections, internal/hook, internal/template/templates"
lifecycle: spec-anchored
tags: "hook, observability, opt-in, plan, sprint-2"
tier: S
---

# SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 — Implementation Plan

## Overview

This plan operationalizes the conversion of 3 observability hook series from always-active to opt-in. The change introduces a single configuration boolean (`.moai/config/sections/system.yaml` `hook.opt_in.enabled:`, default `false`) that gates execution of:

1. `TaskCreated` event hook
2. `Notification` event hook
3. `handle-harness-observe-*` wrapper trio (Stop / SubagentStop / UserPromptSubmit secondary wrappers)

Tier: **S** (minimal — markdown-only spec artifacts in plan-phase, Section A-E minimal scope, M1-M3 milestones).

**Revise note (2026-05-23)**: The original plan-phase first-PASS proposed `observability.yaml` `observability.enabled` as the toggle key. Per spec.md §A.2 collision discovery (the key is byte-identical to pre-existing REQ-OBS-005 trace-logging master toggle) and user decision Q4=A, the toggle is RELOCATED to `system.yaml` `hook.opt_in.enabled` (NEW sub-block under existing `hook:` block). This plan reflects the post-revise mechanism. See spec.md §A for the full pre-existing state survey.

## Goal Alignment

**Design doc objective** (`.moai/research/v3.0-design-2026-05-22.md` § 4 Layer 4): "매 turn당 hook 호출 30% 감소" — per-turn hook invocation count reduces by 30% when the flag is off.

**This SPEC's contribution**: provides the toggle mechanism. The actual 30% reduction realizes when downstream users (or `moai init` default) leave `hook.opt_in.enabled: false`. The KPI is measurable: hook call count per session ↓ ≥ 25% with default config.

## Milestones (Tier S — M1, M2, M3)

### M1 — system.yaml `hook.opt_in.enabled` Schema + Config Loader Extension

**Scope**: Extend existing `system.yaml` `hook:` block with NEW `opt_in.enabled` key and wire the Go loader.

**Deliverables**:

1. Edit existing file: `.moai/config/sections/system.yaml` — extend the existing `hook:` block (currently containing `observability_events: []` and `strict_mode: false`) with a NEW `opt_in:` sub-block:
   ```yaml
   # Hook observability configuration (SPEC-V3R2-RT-006 REQ-004)
   hook:
     # List of retired events to re-enable as observability taps.
     # Empty list (default) means all retired events are silently no-op.
     # Example: [notification, elicitation, elicitationResult, taskCreated]
     observability_events: []
     # When true, retired events in strict mode still succeed silently.
     strict_mode: false
     # Hook opt-in master toggle (SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 REQ-HOI-001)
     # Controls TaskCreated, Notification, and handle-harness-observe-* hook execution.
     # Default: false (hooks are skipped, reduces per-turn overhead by ~25-30%).
     # NOTE: This key is INDEPENDENT of `observability.enabled` in observability.yaml
     # (REQ-OBS-005 trace-logging master toggle). See SPEC §A.3 cohabitation contract.
     opt_in:
       enabled: false
   ```
2. Template mirror: `internal/template/templates/.moai/config/sections/system.yaml` (byte-equivalent to the runtime default, modulo template variable expansion which this section has none of).
3. Go config loader extension — extend the existing `SystemHookConfig` struct (where `ObservabilityEvents` and `StrictMode` already live, per `internal/hook/observability.go` line 32 reference `cfg.Get().System.Hook.ObservabilityEvents`) with a NEW `OptIn` sub-struct:
   ```go
   type SystemHookConfig struct {
       ObservabilityEvents []string          `yaml:"observability_events"`
       StrictMode          bool              `yaml:"strict_mode"`
       OptIn               HookOptInConfig   `yaml:"opt_in"`
   }
   type HookOptInConfig struct {
       Enabled bool `yaml:"enabled"`
   }
   ```
4. Loader contract: when the `opt_in:` sub-block is absent (legacy project), `OptIn.Enabled` defaults to `false` (Go zero-value) — no error, no warning. This matches the §A.3 cohabitation contract and R3 mitigation.
5. **§A.3 cohabitation invariant verification**: `.moai/config/sections/observability.yaml` (REQ-OBS-005 trace-logging master toggle) is NOT touched in M1. `hook.observability_events` (SPEC-V3R2-RT-006 REQ-040) is NOT touched in M1. Only the NEW `hook.opt_in.enabled` sub-block is added.

**Verification**: `AC-HOI-001` (system.yaml key exists with cohabiting keys preserved), `AC-HOI-007` (cohabitation invariant — REQ-OBS-005 + RT-006 untouched).

**Estimated scope**: ~30-50 LOC (system.yaml addition 8-10 lines, Go struct + loader 20-40 lines). Smaller than the original M1 estimate because we extend an existing file/struct instead of creating a new one.

### M2 — settings.json.tmpl Conditional Render + Hook Loader Gating

**Scope**: Two coordinated changes — template-side conditional inclusion AND runtime-side execution gating. Mechanism choice: **NEW dispatcher gate** reading `cfg.System.Hook.OptIn.Enabled` (does NOT reuse `observabilityOptIn()` because that function reads a different key — `ObservabilityEvents` list — and serves RT-006 per-event semantics; see Section "M2/M3 Mechanism Rationale" below).

**Deliverables**:

1. `internal/template/templates/.claude/settings.json.tmpl` — wrap the 3 hook stanzas in a Go template conditional:
   - `hooks.TaskCreated` array entries
   - `hooks.Notification` array entries
   - `hooks.Stop` / `hooks.SubagentStop` / `hooks.UserPromptSubmit` — the `handle-harness-observe-*` SECONDARY wrappers only (primary wrappers stay always-rendered)
   - Wrap pattern: `{{ if .HookOptIn.Enabled }}...{{ end }}`
   - Template context extension required (see Technical Approach §A below).
2. `internal/hook/` runtime gating — introduce a NEW helper alongside existing `observabilityOptIn()`:
   ```go
   // hookOptInEnabled reports whether the 3 observability-only hook series
   // (TaskCreated, Notification, handle-harness-observe-*) are enabled.
   // Reads system.yaml hook.opt_in.enabled (SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 REQ-HOI-001).
   //
   // INDEPENDENT of observabilityOptIn() which reads hook.observability_events
   // (SPEC-V3R2-RT-006 REQ-040 per-event whitelist) — different semantics, do NOT unify
   // without a fresh SPEC. See SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 §A.3.
   func hookOptInEnabled(cfg ConfigProvider) bool {
       if cfg == nil || cfg.Get() == nil {
           return false
       }
       return cfg.Get().System.Hook.OptIn.Enabled
   }
   ```
   This new helper is placed in a NEW file `internal/hook/hook_opt_in.go` (NOT in `observability.go` to keep the cohabitation contract textually visible — separate file = separate concern).
3. Dispatcher integration — at the dispatcher level (where hook events are routed to wrappers), add an `hookOptInEnabled(cfg)` early-return check for the 3 hook series when disabled. This is a defense-in-depth measure: even if `settings.json` was hand-edited to include the hooks, the runtime layer still respects the flag. Pattern A semantics (silent return `HookOutput{}`) per `observability.go` line 18 convention.
4. Integration tests in `internal/hook/hook_opt_in_test.go` (NEW file, parallel to existing `audit_test.go`):
   - `TestHookOptInDisabled` — set flag `false`, fire `TaskCreated` + `Notification` + `harness-observe-stop`, assert zero JSONL appends and zero log lines.
   - `TestHookOptInEnabled` — set flag `true`, fire same 3 events, assert one JSONL entry per event with payload schema unchanged from baseline.
   - `TestHookOptInCohabitation` — covers AC-HOI-007 — 4-quadrant test (HOI on/off × OBS on/off) verifying all 3 systems function independently.

**Verification**: `AC-HOI-003` (disabled path), `AC-HOI-004` (enabled path), `AC-HOI-005` (settings.json conditional render), `AC-HOI-007` (cohabitation 4-quadrant).

**Estimated scope**: ~200-300 LOC (template ~30 lines added, runtime gating helper ~30 lines new file, dispatcher integration ~20 lines, tests ~150-200 lines including 4-quadrant cohabitation matrix).

### M3 — moai doctor Report + Test Fixture Refactor + moai init Default

**Scope**: User-visible diagnostic surface and clean-up of pre-existing test fixtures.

**Deliverables**:

1. `moai doctor` extension — add hook opt-in status section to its output. Pattern:
   ```
   Hook opt-in: disabled (set .moai/config/sections/system.yaml `hook.opt_in.enabled: true` to enable)
   ```
   or when enabled:
   ```
   Hook opt-in: enabled
   ```
   When the `opt_in:` sub-block is absent on a legacy project, doctor reports `disabled` and adds a tip to run `moai update` to materialize the schema. This line is DISTINCT from any existing `Observability:` line that REQ-OBS-005 may already produce (cohabitation invariant per REQ-HOI-005).
2. `moai init` default — confirm and test that fresh init renders `system.yaml` with `hook.opt_in.enabled: false` (no flag required). The `--hook-opt-in` flag MAY be added as an opt-in shortcut to flip the default to `true` at init time. Decision: this SPEC keeps `--hook-opt-in` flag as OUT OF SCOPE for now; default behavior is sufficient.
3. Test fixture refactor — identify any tests in `internal/hook/...` or `internal/cli/...` that assume the 3 observability hook series are always-active. Update fixtures to either:
   - explicitly set `hook.opt_in.enabled: true` in the test setup if they need the hooks, OR
   - remove the assertion if it was incidental.
4. `internal/hook/observability.go` annotation — add a top-of-file comment block referencing SPEC §A.3 cohabitation contract:
   ```go
   // COHABITATION NOTE (SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 §A.3):
   // observabilityOptIn() (this file) reads system.yaml hook.observability_events
   // — SPEC-V3R2-RT-006 REQ-040 per-event RETIRE-OBS-ONLY whitelist.
   //
   // hookOptInEnabled() (hook_opt_in.go) reads system.yaml hook.opt_in.enabled
   // — SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 REQ-HOI-001 master toggle for 3 hook series.
   //
   // observability.yaml `enabled:` reads from cfg.Observability.Enabled
   // — REQ-OBS-005 trace-logging master toggle (DIFFERENT FILE, untouched here).
   //
   // ALL 3 KEYS ARE INDEPENDENT. Do NOT unify without a fresh SPEC.
   ```
   This implements R5 mitigation (cohabitation regression guard).
5. Cross-platform verification — run `go test ./...` on darwin, linux, windows runners (CI must pass).

**Verification**: `AC-HOI-002` (init default), `AC-HOI-006` (doctor line).

**Estimated scope**: ~100-150 LOC (doctor extension ~30 lines, fixture updates ~50-80 lines, init test ~20 lines, observability.go cohabitation comment ~15 lines).

## M2/M3 Mechanism Rationale

### Why NOT reuse `observabilityOptIn()` for the M2 gating

A natural-seeming design simplification would be to reuse the existing `observabilityOptIn(cfg, eventName)` helper (see spec.md §A.1.3) for the 3 HOI hook series. **This was considered and rejected** for the following reasons:

1. **Semantic mismatch**: `observabilityOptIn()` reads `hook.observability_events` (a per-event whitelist) and serves SPEC-V3R2-RT-006 REQ-040 (RETIRE-OBS-ONLY tap re-enablement for retired events). HOI is a master toggle (single boolean) covering 3 always-active hook series — different shape, different semantics.
2. **AC-HOI-007 cohabitation invariant**: The 4-quadrant test explicitly verifies that the 3 systems (HOI, OBS, RT-006) function independently. Reusing the same gate would collapse 2 of the 3 systems and break AC-HOI-007.
3. **Future evolution risk**: A future RT-006 evolution (e.g., per-event opt-in granularity) would inadvertently affect HOI's 3 hook series if they shared the gate. Separate gates preserve independent evolution paths.
4. **Audit clarity**: M3 deliverable #4 adds a comment block to `observability.go` explicitly warning against unification — this comment is meaningful only if the two helpers exist in separate files.

### Why M2 uses defense-in-depth (template conditional + runtime gate)

Template-side conditional alone is insufficient because users may hand-edit `settings.json`. Runtime gate alone is insufficient because rendering the 3 hooks in `settings.json` (even if gated at runtime) bloats config and confuses users inspecting their settings. Both layers are required:

- **Template conditional** (M2 deliverable #1) — clean `settings.json` on default render, no hook references for disabled state
- **Runtime gate** (M2 deliverable #3) — defense against hand-edited `settings.json`, supports hot-reload via `ConfigChange` hook without process restart

## Cross-SPEC Coordination

### R2 Resolution — HOI merges BEFORE HOOK-ASYNC-EXPAND-001

Confirmed ordering:

1. **HOI (this SPEC) merges first**. After merge, `hook.opt_in.enabled: false` is the default and the 3 hook series are skipped at runtime.
2. **HOOK-ASYNC-EXPAND-001 merges second**. Its plan MUST make the `TaskCreated` and `Notification` async stanzas conditional on `hook.opt_in.enabled == true` (since they have no effect when disabled). The `handle-harness-observe-*` async transition is independent of the opt-in toggle ONLY for the `harness-observe` wrapper trio scope — async transition still requires the flag to enable.

This ordering avoids two conflicting refactors on the same hook stanzas. HOOK-ASYNC-EXPAND-001's spec.md will reference this SPEC as a precondition. The key-name change (from original `observability.enabled` to revised `hook.opt_in.enabled`) MUST be propagated to HOOK-ASYNC-EXPAND-001's plan when it enters run-phase.

### Sprint 2 partner independence

`SPEC-V3R6-AGENT-MODEL-ROUTING-001` and `SPEC-V3R6-PROMPT-CACHE-001` — neither touches `internal/hook/` or `.moai/config/sections/system.yaml` `hook:` block. No coordination required; these may merge in any order relative to HOI.

### Wave 0 regression guard

`SPEC-V3R6-HOOK-CONTRACT-FIX-001` — fixes `WorktreeCreate` hook contract. Same `internal/hook/` module surface, but completely orthogonal hook event. No coordination required.

### Cohabitation invariant (NEW post-revise) — REQ-OBS-005 + SPEC-V3R2-RT-006

[ZONE:Frozen for this SPEC] [HARD] HOI's run-phase MUST NOT modify:
- `.moai/config/sections/observability.yaml` (REQ-OBS-005 trace-logging owner)
- `.moai/config/sections/system.yaml` `hook.observability_events` (SPEC-V3R2-RT-006 REQ-040)
- `.moai/config/sections/system.yaml` `hook.strict_mode` (SPEC-V3R2-RT-006 sibling)
- `internal/hook/observability.go` `observabilityOptIn()` function body (only the file-top COHABITATION NOTE comment is added per M3 deliverable #4)
- `internal/hook/coverage_table.go` `ResolutionRetireObsOnly` + `ObservabilityOptIn` fields
- `internal/hook/audit_test.go` `TestAuditObservabilityWhitelist` test body

AC-HOI-007 4-quadrant test enforces this invariant as a permanent regression guard.

## Technical Approach

### A. system.yaml schema extension and config loading sequence

1. `moai` binary boot → load `.moai/config/sections/*.yaml` via existing loader
2. Loader assembles `Config` struct. The existing `System.Hook` field (`SystemHookConfig`) gains a NEW `OptIn` sub-struct (`HookOptInConfig` with `Enabled bool`).
3. When the `opt_in:` sub-block is absent (legacy project pre-update): Go zero-value `false` is the correct default — no special case needed, matches R3 mitigation
4. When the `opt_in:` sub-block exists with malformed YAML: existing loader error path applies (no special handling)
5. Pre-existing `observability_events []string` and `strict_mode bool` fields remain in `SystemHookConfig` — schema is purely additive

### B. Settings.json template rendering

The template uses Go's text/template package. The render context (`SettingsContext`) already exposes the loaded config. Extend the context with:

```go
type SettingsContext struct {
    // ... existing fields ...
    HookOptIn HookOptInContext
}

type HookOptInContext struct {
    Enabled bool
}
```

The `HookOptIn` field is populated from `cfg.System.Hook.OptIn` during template context construction (NOT from `cfg.Observability` which is REQ-OBS-005's namespace).

Template conditional:

```
{{ if .HookOptIn.Enabled }}
"TaskCreated": [
  { "hooks": [{ "command": "... handle-task-created.sh ..." }] }
],
"Notification": [
  { "hooks": [{ "command": "... handle-notification.sh ..." }] }
],
{{ end }}
```

For the `Stop` / `SubagentStop` / `UserPromptSubmit` wrappers, the primary wrapper stays always-rendered; the secondary `handle-harness-observe-*` wrapper is wrapped:

```
"Stop": [
  { "hooks": [{ "command": "... handle-stop.sh ..." }] }
  {{ if .HookOptIn.Enabled }},
  { "hooks": [{ "command": "... handle-harness-observe-stop.sh ..." }] }
  {{ end }}
],
```

### C. Runtime defense-in-depth gating

In `internal/hook/dispatcher.go` (or equivalent), the dispatcher function that routes events to wrappers checks `hookOptInEnabled(cfg)` before invoking any of the 3 series. If returns `false`, the dispatcher returns `HookOutput{}` immediately (zero side effects). This ensures:

- Hand-edited `settings.json` cannot bypass the flag
- Hot-reload via `ConfigChange` hook applies the new value without process restart
- Tests can deterministically toggle the flag via fixture setup

The NEW helper lives in `internal/hook/hook_opt_in.go` (separate from `observability.go` per M2 deliverable #2 rationale).

## Risks & Mitigations (cross-reference spec.md § F)

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| R1 — Silent telemetry loss on upgrade | Low | Medium | v3.0 major bump, release notes call out toggle (NEW key name `hook.opt_in.enabled`), migration guide in Wave 5 docs SPEC |
| R2 — Conflict with HOOK-ASYNC-EXPAND-001 | Medium | Low | HOI merges first; ASYNC-EXPAND made conditional on `hook.opt_in.enabled == true`; key name change propagated |
| R3 — Legacy project doctor regression | Low | Low | Doctor defaults to `disabled` report when `opt_in:` sub-block absent; never errors |
| R4 — Test fixture breakage | Low | Low | M3 explicitly enumerates fixture updates; CI cross-platform run gates merge |
| R5 — Cohabitation regression (NEW) | Low | Medium | AC-HOI-007 4-quadrant test permanent guard; M3 deliverable #4 file-top COHABITATION NOTE in `observability.go` |

## Definition of Done

- All 5 REQs implemented and verified via mapped ACs
- All 7 ACs PASS on darwin + linux + windows (CI green)
- 0 NEW golangci-lint issues over baseline
- 0 NEW `go test ./...` failures over baseline
- C-HRA-008 subagent-boundary grep: 0 matches (no inadvertent `AskUserQuestion` references in `internal/hook/`)
- Template mirror parity: `internal/template/templates/.moai/config/sections/system.yaml` byte-equivalent to runtime default for the `hook.opt_in.enabled` sub-block
- §A.3 cohabitation invariant verified: `observability.yaml` untouched + `hook.observability_events`/`hook.strict_mode` untouched + `observabilityOptIn()` function body untouched (only file-top comment added)
- Spec frontmatter: status field updates to `implemented` only after all ACs PASS; version bumps `0.1.0 → 0.2.0` at sync time
- progress.md created during plan-revise (this iteration) documenting collision discovery + Q4=A relocation, extended in run-phase with M1-M3 evidence

## Out of Scope (cross-reference spec.md § F)

See `spec.md` § F for the 4 h3-sectioned Out of Scope entries:

### Out of Scope: PostToolUse / SessionStart / PreToolUse opt-in
Synchronous critical-path hooks stay always-active.

### Out of Scope: Async expansion of observability hooks
Owned by sibling `SPEC-V3R6-HOOK-ASYNC-EXPAND-001`.

### Out of Scope: Telemetry data schema redesign
JSONL payload format preserved exactly when `hook.opt_in.enabled: true`.

### Out of Scope: Migration of REQ-OBS-005 or SPEC-V3R2-RT-006 key namespaces
Both pre-existing keys preserved untouched; AC-HOI-007 enforces.
