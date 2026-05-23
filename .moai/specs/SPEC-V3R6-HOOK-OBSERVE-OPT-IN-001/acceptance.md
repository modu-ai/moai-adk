---
id: SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001
title: "Observability hook 3계열 opt-in — Acceptance Criteria (Tier S, 7 binary ACs)"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: Author Name
priority: P2
phase: "v3.0.0"
module: ".moai/config/sections, internal/hook, internal/template/templates"
lifecycle: spec-anchored
tags: "hook, observability, opt-in, acceptance, sprint-2"
tier: S
---

# SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 — Acceptance Criteria

## Overview

This document specifies binary, verifiable acceptance criteria (AC-HOI-001..007) for SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001. Each AC maps to one or more REQs (see spec.md § D Traceability Matrix) and provides a deterministic verification command.

**Revise note (2026-05-23)**: The original draft used `observability.yaml` `observability.enabled` as the toggle key. Per spec.md §A.2 collision discovery and user Q4=A decision, the toggle has been RELOCATED to `system.yaml` `hook.opt_in.enabled`. AC-HOI-001 and AC-HOI-007 have been substantively revised. AC-HOI-002 through AC-HOI-006 path/grep targets updated to the new key location. AC-HOI-007 has been reformulated from a template-mirror-parity check (now subsumed under AC-HOI-001 cohabitation invariant) to a dedicated 4-quadrant cohabitation regression test. See progress.md "Revise History" for full delta.

## Binary Acceptance Criteria

### AC-HOI-001 — system.yaml `hook.opt_in.enabled` key exists with cohabiting keys preserved

**REQ mapping**: REQ-HOI-001

**Given** the project has been initialized via `moai init` (or already exists with current MoAI templates),
**When** the user inspects `.moai/config/sections/system.yaml`,
**Then** the file shall contain `hook.opt_in.enabled: false` as a sub-block under the existing `hook:` block, AND the pre-existing `hook.observability_events: []` (SPEC-V3R2-RT-006 REQ-040) AND `hook.strict_mode: false` keys shall remain present unchanged.

**Verification commands**:
```bash
# Check 1 — NEW key exists with default false
grep -A1 '^\s*opt_in:' .moai/config/sections/system.yaml | grep -E '^\s*enabled:\s*false'
echo "check1_exit=$?"

# Check 2 — pre-existing observability_events still present (cohabitation invariant)
grep -E '^\s*observability_events:' .moai/config/sections/system.yaml
echo "check2_exit=$?"

# Check 3 — pre-existing strict_mode still present (cohabitation invariant)
grep -E '^\s*strict_mode:' .moai/config/sections/system.yaml
echo "check3_exit=$?"
```

**Expected**: All 3 exit codes equal 0.

**Pass criteria**: Binary — all 3 grep matches succeed (NEW key + 2 cohabiting keys preserved).

### AC-HOI-002 — moai init default is hook.opt_in.enabled: false

**REQ mapping**: REQ-HOI-004

**Given** a clean directory with no prior MoAI initialization,
**When** the user runs `moai init <project>` without any `--hook-opt-in` flag,
**Then** the rendered `system.yaml` shall have `hook.opt_in.enabled: false` (under the `hook:` block).

**Verification command**:
```bash
TEST_DIR=$(mktemp -d -t moai-init-XXXXXX)
moai init "$TEST_DIR/test-project" --quiet
grep -A1 '^\s*opt_in:' "$TEST_DIR/test-project/.moai/config/sections/system.yaml" | grep -E '^\s*enabled:\s*false'
echo "exit=$?"
```

**Expected**: Final exit code 0; grep matches exactly one `enabled: false` line under `opt_in:`.

**Pass criteria**: Binary — match exists with literal `false` value.

### AC-HOI-003 — Disabled path skips all 3 hook series

**REQ mapping**: REQ-HOI-002

**Given** `hook.opt_in.enabled: false` in the loaded config,
**When** an integration test fires `TaskCreated` + `Notification` + `harness-observe-stop` events,
**Then** zero JSONL appends, zero log lines, and zero telemetry side effects shall occur for these 3 events.

**Verification command**:
```bash
go test -run TestHookOptInDisabled ./internal/hook/...
```

**Expected**: Exit code 0, test passes with assertions:
- JSONL output file size unchanged after firing events
- Log output contains no `[harness-observe]` or `[task-created]` or `[notification]` markers
- No file writes to `.moai/logs/observability/` (or equivalent telemetry sink) caused by the 3 series specifically

**Pass criteria**: Binary — `go test` exits 0 for `TestHookOptInDisabled`.

### AC-HOI-004 — Enabled path executes all 3 hook series with unchanged payload

**REQ mapping**: REQ-HOI-003

**Given** `hook.opt_in.enabled: true` in the loaded config,
**When** an integration test fires `TaskCreated` + `Notification` + `harness-observe-stop` events,
**Then** each event shall produce exactly one JSONL entry whose payload schema is byte-equivalent to the pre-change baseline.

**Verification command**:
```bash
go test -run TestHookOptInEnabled ./internal/hook/...
```

**Expected**: Exit code 0, test passes with assertions:
- Exactly 3 JSONL entries appended (one per event)
- Each payload contains the expected fields (`event_type`, `timestamp`, `payload`, etc.)
- Schema diff against baseline fixture: 0 field additions, 0 field removals, 0 field renames

**Pass criteria**: Binary — `go test` exits 0 for `TestHookOptInEnabled` AND payload schema diff is empty.

### AC-HOI-005 — settings.json conditional render reflects flag

**REQ mapping**: REQ-HOI-004

**Given** a freshly initialized project with `hook.opt_in.enabled: false`,
**When** the user inspects the rendered `.claude/settings.json`,
**Then** the file shall contain zero references to `handle-harness-observe-stop`.

**And** **Given** the user flips the flag to `enabled: true` (under `hook.opt_in:` in `system.yaml`) and runs `moai update`,
**When** the user re-inspects `.claude/settings.json`,
**Then** the file shall contain at least one reference to `handle-harness-observe-stop`.

**Verification commands**:
```bash
# Phase 1: Default (disabled)
TEST_DIR=$(mktemp -d -t moai-render-XXXXXX)
moai init "$TEST_DIR/proj" --quiet
DISABLED_COUNT=$(grep -c handle-harness-observe-stop "$TEST_DIR/proj/.claude/settings.json")
echo "disabled_count=$DISABLED_COUNT"
[ "$DISABLED_COUNT" -eq 0 ] || exit 1

# Phase 2: Enabled — toggle hook.opt_in.enabled in system.yaml
# Note: yq or sed required to surgically flip the nested key
python3 -c "
import re, sys
p = '$TEST_DIR/proj/.moai/config/sections/system.yaml'
with open(p) as f: src = f.read()
# Replace 'opt_in:\n    enabled: false' with 'opt_in:\n    enabled: true'
out = re.sub(r'(opt_in:\s*\n\s*enabled:\s*)false', r'\1true', src)
with open(p, 'w') as f: f.write(out)
"
cd "$TEST_DIR/proj" && moai update --quiet
ENABLED_COUNT=$(grep -c handle-harness-observe-stop "$TEST_DIR/proj/.claude/settings.json")
echo "enabled_count=$ENABLED_COUNT"
[ "$ENABLED_COUNT" -ge 1 ] || exit 1
```

**Expected**: `disabled_count=0` AND `enabled_count>=1`, both phase exit codes 0.

**Pass criteria**: Binary — both counts match the expected values.

### AC-HOI-006 — moai doctor reports hook opt-in status

**REQ mapping**: REQ-HOI-005

**Given** any MoAI project (with or without the `hook.opt_in` sub-block in `system.yaml`),
**When** the user runs `moai doctor`,
**Then** the output shall contain a line matching the regex `Hook opt-in:\s*(enabled|disabled)`.

**Verification command**:
```bash
moai doctor 2>&1 | grep -E 'Hook opt-in:\s*(enabled|disabled)'
echo "exit=$?"
```

**Expected**: Exit code 0, at least one matching line. This line is DISTINCT from any pre-existing `Observability:` line that REQ-OBS-005 may produce — both lines may coexist.

**Pass criteria**: Binary — grep matches with one of the two enum values.

### AC-HOI-007 — Cohabitation invariant (3-key independence)

**REQ mapping**: REQ-HOI-001 (with cross-reference to REQ-OBS-005 and SPEC-V3R2-RT-006 REQ-040)

**Given** the three independent keys defined in spec.md §A.3:
- `hook.opt_in.enabled` (THIS SPEC, NEW) — 3 hook series master toggle
- `observability.enabled` in `observability.yaml` (REQ-OBS-005, pre-existing) — trace-logging master toggle
- `hook.observability_events` in `system.yaml` (SPEC-V3R2-RT-006 REQ-040, pre-existing) — per-event RETIRE-OBS-ONLY whitelist

**When** the integration test exercises all 4 quadrants of (HOI ∈ {false, true}) × (OBS ∈ {false, true}) while also varying the RT-006 whitelist between empty and non-empty,

**Then** each system shall function independently with no cross-system side effects:
- Quadrant 1 (HOI=false, OBS=false): 3 hook series SKIPPED, trace logging OFF, RT-006 events SKIPPED per whitelist
- Quadrant 2 (HOI=false, OBS=true): 3 hook series SKIPPED, trace logging ON, RT-006 events SKIPPED per whitelist
- Quadrant 3 (HOI=true, OBS=false): 3 hook series EXECUTED, trace logging OFF, RT-006 events follow whitelist
- Quadrant 4 (HOI=true, OBS=true): 3 hook series EXECUTED, trace logging ON, RT-006 events follow whitelist

**Verification command**:
```bash
go test -run TestHookOptInCohabitation ./internal/hook/...
```

**Expected**: Exit code 0, test passes all 4 quadrants with explicit assertions per quadrant:
- HOI 3-hook-series execution count matches expected (0 when HOI=false, ≥3 when HOI=true)
- Trace log file written or not based on OBS=true/false (REQ-OBS-005 behavior unchanged)
- `observabilityOptIn(cfg, "taskCreated")` returns based on RT-006 whitelist contents, INDEPENDENT of HOI/OBS values

Additionally, the run-phase MUST verify by inspection (or static-analysis test) that:
- `.moai/config/sections/observability.yaml` byte-content is UNCHANGED from pre-merge baseline (REQ-OBS-005 owner's file untouched)
- `internal/hook/observability.go` `observabilityOptIn()` function body is UNCHANGED (only file-top COHABITATION NOTE comment added per plan.md M3 deliverable #4)
- `internal/hook/coverage_table.go` `ResolutionRetireObsOnly` + `ObservabilityOptIn` fields are UNCHANGED
- `internal/hook/audit_test.go` `TestAuditObservabilityWhitelist` test body is UNCHANGED

**Pass criteria**: Binary — `go test` exits 0 for `TestHookOptInCohabitation` AND all 4 file-untouched assertions pass.

## Edge Cases (Verification Notes)

### Edge case 1: Legacy project without hook.opt_in sub-block

**Given** a pre-v3.0 project that has been upgraded to v3.0 binary but has not yet run `moai update`,
**When** the user invokes any MoAI command that loads config,
**Then** the loader shall treat missing `hook.opt_in` sub-block as `enabled: false` (Go zero-value default).

**Verification**: Implicit in AC-HOI-003 — the disabled-path test would also pass on a project missing the sub-block entirely (loader default kicks in). M3 milestone includes an explicit fixture test for this case: `TestHookOptInMissingKey_DefaultsDisabled`.

### Edge case 2: Hand-edited settings.json with hooks present but hook.opt_in.enabled is false

**Given** a project where `system.yaml` has `hook.opt_in.enabled: false` but `.claude/settings.json` has been hand-edited to include `handle-harness-observe-stop` references,
**When** the dispatcher routes a `Stop` event,
**Then** the runtime defense-in-depth gating shall still skip the `handle-harness-observe-stop` wrapper.

**Verification**: Covered by `TestHookOptInDisabled` (the test injects a settings.json with the wrappers present and asserts they are not executed). This is the runtime gating contract, complementing the template-side conditional render.

### Edge case 3: ConfigChange hot-reload

**Given** a running MoAI session with `hook.opt_in.enabled: false`,
**When** the user edits `system.yaml` to set `hook.opt_in.enabled: true` and the `ConfigChange` hook fires,
**Then** subsequent hook invocations within the same session shall respect the new value without process restart.

**Verification**: Manual smoke test documented in run-phase progress.md. Not part of automated AC set (would require multi-session orchestration). Documented here for completeness.

### Edge case 4: Naming collision regression (NEW post-revise)

**Given** a future PR proposes to rename `hook.opt_in.enabled` back to `observability.enabled` (in either `observability.yaml` or `system.yaml`),
**When** the AC-HOI-007 cohabitation test runs,
**Then** the test SHOULD fail because the unification would collapse 2 distinct semantics into 1 key.

**Verification**: `TestHookOptInCohabitation` 4-quadrant matrix relies on independent loader paths. A rename collapsing the keys would cause OBS=false + HOI=true (or vice versa) to become impossible to test, immediately surfacing the regression. Documented here as a permanent regression guard contract.

## Quality Gate Criteria

In addition to the 7 binary ACs above, the run-phase must satisfy:

1. **`go test ./...`** — exits 0, zero NEW test failures over the pre-change baseline.
2. **`golangci-lint run --timeout=2m`** — zero NEW lint issues over the pre-change baseline.
3. **`grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/`** — zero matches excluding `_test.go` files and comment lines (C-HRA-008 subagent boundary).
4. **Cross-platform CI** — darwin + linux + windows runners all green.
5. **Coverage** — `internal/hook/...` package coverage stays at or above the pre-change baseline (no coverage regression).
6. **Frontmatter lint** — `moai spec lint .moai/specs/SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001/spec.md` exits 0 (12-field canonical schema compliance).
7. **Cohabitation file-untouched check** (NEW post-revise) — verify via git diff that `.moai/config/sections/observability.yaml`, `internal/hook/observability.go` (function body only — file-top comment is the allowed exception), `internal/hook/coverage_table.go`, and `internal/hook/audit_test.go` are unchanged from pre-merge baseline. Implementation MAY be a CI script or a Go test using `git diff --quiet`.

## Definition of Done (DoD)

The SPEC is **DONE** when all of the following hold:

- [ ] AC-HOI-001 through AC-HOI-007: all 7 binary ACs PASS (note AC-HOI-007 reformulated as 4-quadrant cohabitation test post-revise)
- [ ] All 7 quality gate criteria above PASS (including NEW criterion #7 cohabitation file-untouched check)
- [ ] R1-R5 mitigations from spec.md § F are in place (release notes draft prepared for R1 with NEW key name `hook.opt_in.enabled` called out, ASYNC-EXPAND coordination noted for R2, doctor fallback verified for R3, fixture updates merged for R4, AC-HOI-007 permanent regression guard for R5)
- [ ] progress.md extended in run-phase with M1-M3 evidence (plan-phase revise iteration already documented per spec.md §A.2)
- [ ] spec.md frontmatter `status:` updated from `draft` to `implemented` (sync phase will move to `completed`)
- [ ] spec.md `version:` bumped `0.1.0 → 0.2.0` at sync time

## REQ ↔ AC Traceability (mirror of spec.md § D)

| REQ | Mapped ACs | Coverage Type |
|---|---|---|
| REQ-HOI-001 (system.yaml `hook.opt_in.enabled` + cohabitation) | AC-HOI-001, AC-HOI-007 | NEW key exists with cohabiting keys preserved + 4-quadrant independence test |
| REQ-HOI-002 (disabled skips) | AC-HOI-003 | Integration test — SKIP path |
| REQ-HOI-003 (enabled executes) | AC-HOI-004 | Integration test — EXECUTE path, schema preserved |
| REQ-HOI-004 (init default + render) | AC-HOI-002, AC-HOI-005 | Init default + settings.json conditional render |
| REQ-HOI-005 (doctor report) | AC-HOI-006 | Doctor diagnostic line distinct from any REQ-OBS-005 line |

Coverage: 5 REQs → 7 ACs, 100% bidirectional. Every REQ has at least one mapped AC; every AC traces back to at least one REQ. No orphaned criteria.

## Out of Scope (cross-reference spec.md § F)

The following are explicitly NOT verified by this AC set — they belong to separate SPECs or future work:

### Out of Scope: PostToolUse / SessionStart / PreToolUse opt-in verification
These hooks are not gated by the flag introduced here. No AC tests their always-active status.

### Out of Scope: Async transition behavior verification
Async execution semantics are owned by `SPEC-V3R6-HOOK-ASYNC-EXPAND-001`. This AC set only verifies ON/OFF toggling; it does not verify async vs sync execution timing.

### Out of Scope: Telemetry payload schema validation beyond byte-equivalence
AC-HOI-004 verifies payload schema is unchanged at the field level. It does NOT verify downstream harness consumers correctly parse the payload — that responsibility lies with harness integration tests, which are outside this SPEC's scope.

### Out of Scope: Migration / unification of REQ-OBS-005 or SPEC-V3R2-RT-006 keys
AC-HOI-007 verifies the 3 keys remain INDEPENDENT. It does NOT verify or enable any unification path. Future key unification requires a fresh SPEC with explicit migration plan.
