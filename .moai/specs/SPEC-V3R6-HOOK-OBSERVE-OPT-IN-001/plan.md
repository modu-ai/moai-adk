---
id: SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001
title: "Observability hook 3계열 opt-in — Implementation Plan (Tier S, M1-M3)"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".moai/config/sections, internal/hook, internal/template/templates"
lifecycle: spec-anchored
tags: "hook, observability, opt-in, plan, sprint-2"
tier: S
---

# SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 — Implementation Plan

## Overview

This plan operationalizes the conversion of 3 observability hook series from always-active to opt-in. The change introduces a single configuration boolean (`.moai/config/sections/observability.yaml` `enabled:`, default `false`) that gates execution of:

1. `TaskCreated` event hook
2. `Notification` event hook
3. `handle-harness-observe-*` wrapper trio (Stop / SubagentStop / UserPromptSubmit secondary wrappers)

Tier: **S** (minimal — markdown-only spec artifacts in plan-phase, Section A-E minimal scope, M1-M3 milestones).

## Goal Alignment

**Design doc objective** (`.moai/research/v3.0-design-2026-05-22.md` § 4 Layer 4): "매 turn당 hook 호출 30% 감소" — per-turn hook invocation count reduces by 30% when the flag is off.

**This SPEC's contribution**: provides the toggle mechanism. The actual 30% reduction realizes when downstream users (or `moai init` default) leave `enabled: false`. The KPI is measurable: hook call count per session ↓ ≥ 25% with default config.

## Milestones (Tier S — M1, M2, M3)

### M1 — observability.yaml Schema + Config Loader

**Scope**: Introduce the configuration file and Go-side loader.

**Deliverables**:

1. New file: `.moai/config/sections/observability.yaml` with schema:
   ```yaml
   # Observability hook opt-in toggle
   # Controls TaskCreated, Notification, and handle-harness-observe-* hook execution.
   # Default: false (hooks are skipped, reduces per-turn overhead by ~25-30%)
   enabled: false
   ```
2. Template mirror: `internal/template/templates/.moai/config/sections/observability.yaml` (byte-equivalent).
3. Go config loader extension — add `ObservabilitySection` struct with `Enabled bool` field, wire into existing config aggregation in `internal/config/...`.
4. Loader contract: when file is absent (legacy project), `Enabled` defaults to `false` — no error, no warning.

**Verification**: `AC-HOI-001` (file + key exist), `AC-HOI-007` (template mirror parity).

**Estimated scope**: ~50-80 LOC (yaml file 5-10 lines, Go struct + loader 40-70 lines).

### M2 — settings.json.tmpl Conditional Render + Hook Loader Gating

**Scope**: Two coordinated changes — template-side conditional inclusion AND runtime-side execution gating.

**Deliverables**:

1. `internal/template/templates/.claude/settings.json.tmpl` — wrap the 3 hook stanzas in a Go template conditional:
   - `hooks.TaskCreated` array entries
   - `hooks.Notification` array entries
   - `hooks.Stop` / `hooks.SubagentStop` / `hooks.UserPromptSubmit` — the `handle-harness-observe-*` SECONDARY wrappers only (primary wrappers stay always-rendered)
   - Wrap pattern: `{{ if .Observability.Enabled }}...{{ end }}`
2. `internal/hook/` runtime gating — at the dispatcher level (where hook events are routed to wrappers), add an `observability.enabled` check that early-returns for the 3 hook series when disabled. This is a defense-in-depth measure: even if `settings.json` was hand-edited to include the hooks, the runtime layer still respects the flag.
3. Integration tests in `internal/hook/observability_test.go` (or equivalent path):
   - `TestObservabilityDisabled` — set flag `false`, fire `TaskCreated` + `Notification` + `harness-observe-stop`, assert zero JSONL appends and zero log lines.
   - `TestObservabilityEnabled` — set flag `true`, fire same 3 events, assert one JSONL entry per event with payload schema unchanged from baseline.

**Verification**: `AC-HOI-003` (disabled path), `AC-HOI-004` (enabled path), `AC-HOI-005` (settings.json conditional render).

**Estimated scope**: ~150-250 LOC (template ~30 lines added, runtime gating ~50 lines, tests ~100-150 lines).

### M3 — moai doctor Report + Test Fixture Refactor + moai init Default

**Scope**: User-visible diagnostic surface and clean-up of pre-existing test fixtures.

**Deliverables**:

1. `moai doctor` extension — add observability status section to its output. Pattern:
   ```
   Observability: disabled (set .moai/config/sections/observability.yaml `enabled: true` to enable)
   ```
   or when enabled:
   ```
   Observability: enabled
   ```
   When file is absent on a legacy project, doctor reports `disabled` and adds a tip to run `moai update` to materialize the file.
2. `moai init` default — confirm and test that fresh init renders `observability.yaml` with `enabled: false` (no flag required). The `--observability` flag MAY be added as an opt-in shortcut to flip the default to `true` at init time. Decision: this SPEC keeps `--observability` flag as OUT OF SCOPE for now; default behavior is sufficient.
3. Test fixture refactor — identify any tests in `internal/hook/...` or `internal/cli/...` that assume observability hooks are always-active. Update fixtures to either:
   - explicitly set `enabled: true` in the test setup if they need the hooks, OR
   - remove the assertion if it was incidental.
4. Cross-platform verification — run `go test ./...` on darwin, linux, windows runners (CI must pass).

**Verification**: `AC-HOI-002` (init default), `AC-HOI-006` (doctor line).

**Estimated scope**: ~100-150 LOC (doctor extension ~30 lines, fixture updates ~50-100 lines, init test ~20 lines).

## Cross-SPEC Coordination

### R2 Resolution — HOI merges BEFORE HOOK-ASYNC-EXPAND-001

Confirmed ordering:

1. **HOI (this SPEC) merges first**. After merge, `observability.enabled: false` is the default and the 3 hook series are skipped at runtime.
2. **HOOK-ASYNC-EXPAND-001 merges second**. Its plan MUST make the `TaskCreated` and `Notification` async stanzas conditional on `observability.enabled == true` (since they have no effect when disabled). The `handle-harness-observe-*` async transition is independent of the opt-in toggle ONLY for the `harness-observe` wrapper trio scope — async transition still requires the flag to enable.

This ordering avoids two conflicting refactors on the same hook stanzas. HOOK-ASYNC-EXPAND-001's spec.md will reference this SPEC as a precondition.

### Sprint 2 partner independence

`SPEC-V3R6-AGENT-MODEL-ROUTING-001` and `SPEC-V3R6-PROMPT-CACHE-001` — neither touches `internal/hook/` or `.moai/config/sections/observability.yaml`. No coordination required; these may merge in any order relative to HOI.

### Wave 0 regression guard

`SPEC-V3R6-HOOK-CONTRACT-FIX-001` — fixes `WorktreeCreate` hook contract. Same `internal/hook/` module surface, but completely orthogonal hook event. No coordination required.

## Technical Approach

### Configuration loading sequence

1. `moai` binary boot → load `.moai/config/sections/*.yaml` via existing loader
2. Loader assembles `Config` struct including new `Observability` section
3. When file is absent: `Observability.Enabled` zero-value (`false`) is the correct default — no special case needed
4. When file exists with malformed YAML: existing loader error path applies (no special handling)

### Settings.json template rendering

The template uses Go's text/template package. The render context already exposes the loaded config. Extend the context with:

```go
type SettingsContext struct {
    // ... existing fields ...
    Observability ObservabilityContext
}

type ObservabilityContext struct {
    Enabled bool
}
```

Template conditional:

```
{{ if .Observability.Enabled }}
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
  {{ if .Observability.Enabled }},
  { "hooks": [{ "command": "... handle-harness-observe-stop.sh ..." }] }
  {{ end }}
],
```

### Runtime defense-in-depth gating

In `internal/hook/dispatcher.go` (or equivalent), the dispatcher function that routes events to wrappers checks `cfg.Observability.Enabled` before invoking any of the 3 series. If `enabled: false`, the dispatcher returns `HookOutput{}` immediately (zero side effects). This ensures:

- Hand-edited `settings.json` cannot bypass the flag
- Hot-reload via `ConfigChange` hook applies the new value without process restart
- Tests can deterministically toggle the flag via fixture setup

## Risks & Mitigations (cross-reference spec.md § E)

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| R1 — Silent telemetry loss on upgrade | Low | Medium | v3.0 major bump, release notes call out toggle, migration guide in Wave 5 docs SPEC |
| R2 — Conflict with HOOK-ASYNC-EXPAND-001 | Medium | Low | HOI merges first; ASYNC-EXPAND made conditional on `enabled: true` |
| R3 — Legacy project doctor regression | Low | Low | Doctor defaults to `disabled` report when file absent; never errors |
| R4 — Test fixture breakage | Low | Low | M3 explicitly enumerates fixture updates; CI cross-platform run gates merge |

## Definition of Done

- All 5 REQs implemented and verified via mapped ACs
- All 7 ACs PASS on darwin + linux + windows (CI green)
- 0 NEW golangci-lint issues over baseline
- 0 NEW `go test ./...` failures over baseline
- C-HRA-008 subagent-boundary grep: 0 matches (no inadvertent `AskUserQuestion` references in `internal/hook/`)
- Template mirror parity: `internal/template/templates/.moai/config/sections/observability.yaml` byte-equivalent to runtime default
- Spec frontmatter: status field updates to `implemented` only after all ACs PASS; version bumps `0.1.0 → 0.2.0` at sync time
- progress.md created during run-phase documenting M1-M3 evidence

## Out of Scope (cross-reference spec.md § E)

See `spec.md` § E for the 3 h3-sectioned Out of Scope entries:

### Out of Scope: PostToolUse / SessionStart / PreToolUse opt-in
Synchronous critical-path hooks stay always-active.

### Out of Scope: Async expansion of observability hooks
Owned by sibling `SPEC-V3R6-HOOK-ASYNC-EXPAND-001`.

### Out of Scope: Telemetry data schema redesign
JSONL payload format preserved exactly when `enabled: true`.
