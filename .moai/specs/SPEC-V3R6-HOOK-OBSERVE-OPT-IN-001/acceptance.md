---
id: SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001
title: "Observability hook 3계열 opt-in — Acceptance Criteria (Tier S, 7 binary ACs)"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".moai/config/sections, internal/hook, internal/template/templates"
lifecycle: spec-anchored
tags: "hook, observability, opt-in, acceptance, sprint-2"
tier: S
---

# SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 — Acceptance Criteria

## Overview

This document specifies binary, verifiable acceptance criteria (AC-HOI-001..007) for SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001. Each AC maps to one or more REQs (see spec.md § C Traceability Matrix) and provides a deterministic verification command.

## Binary Acceptance Criteria

### AC-HOI-001 — observability.yaml exists with enabled: key

**REQ mapping**: REQ-HOI-001

**Given** the project has been initialized via `moai init` (or already exists with current MoAI templates),
**When** the user inspects `.moai/config/sections/observability.yaml`,
**Then** the file shall exist and contain a top-level `enabled:` key.

**Verification command**:
```bash
grep -E '^enabled:' .moai/config/sections/observability.yaml
```

**Expected**: Exit code 0, at least one line of output matching `enabled: <value>`.

**Pass criteria**: Binary — file exists AND grep matches.

### AC-HOI-002 — moai init default is enabled: false

**REQ mapping**: REQ-HOI-004

**Given** a clean directory with no prior MoAI initialization,
**When** the user runs `moai init <project>` without any `--observability` flag,
**Then** the rendered `observability.yaml` shall have `enabled: false`.

**Verification command**:
```bash
TEST_DIR=$(mktemp -d -t moai-init-XXXXXX)
moai init "$TEST_DIR/test-project" --quiet
grep -E '^enabled:\s*false' "$TEST_DIR/test-project/.moai/config/sections/observability.yaml"
echo "exit=$?"
```

**Expected**: Final exit code 0; grep matches exactly one line `enabled: false`.

**Pass criteria**: Binary — match exists with literal `false` value.

### AC-HOI-003 — Disabled path skips all 3 hook series

**REQ mapping**: REQ-HOI-002

**Given** `observability.enabled: false` in the loaded config,
**When** an integration test fires `TaskCreated` + `Notification` + `harness-observe-stop` events,
**Then** zero JSONL appends, zero log lines, and zero telemetry side effects shall occur for these 3 events.

**Verification command**:
```bash
go test -run TestObservabilityDisabled ./internal/hook/...
```

**Expected**: Exit code 0, test passes with assertions:
- JSONL output file size unchanged after firing events
- Log output contains no `[observability]` markers
- No file writes to `.moai/logs/observability/` (or equivalent telemetry sink)

**Pass criteria**: Binary — `go test` exits 0 for `TestObservabilityDisabled`.

### AC-HOI-004 — Enabled path executes all 3 hook series with unchanged payload

**REQ mapping**: REQ-HOI-003

**Given** `observability.enabled: true` in the loaded config,
**When** an integration test fires `TaskCreated` + `Notification` + `harness-observe-stop` events,
**Then** each event shall produce exactly one JSONL entry whose payload schema is byte-equivalent to the pre-change baseline.

**Verification command**:
```bash
go test -run TestObservabilityEnabled ./internal/hook/...
```

**Expected**: Exit code 0, test passes with assertions:
- Exactly 3 JSONL entries appended (one per event)
- Each payload contains the expected fields (`event_type`, `timestamp`, `payload`, etc.)
- Schema diff against baseline fixture: 0 field additions, 0 field removals, 0 field renames

**Pass criteria**: Binary — `go test` exits 0 for `TestObservabilityEnabled` AND payload schema diff is empty.

### AC-HOI-005 — settings.json conditional render reflects flag

**REQ mapping**: REQ-HOI-004

**Given** a freshly initialized project with `observability.enabled: false`,
**When** the user inspects the rendered `.claude/settings.json`,
**Then** the file shall contain zero references to `handle-harness-observe-stop`.

**And** **Given** the user flips the flag to `enabled: true` and runs `moai update`,
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

# Phase 2: Enabled
sed -i.bak 's/^enabled:.*/enabled: true/' "$TEST_DIR/proj/.moai/config/sections/observability.yaml"
cd "$TEST_DIR/proj" && moai update --quiet
ENABLED_COUNT=$(grep -c handle-harness-observe-stop "$TEST_DIR/proj/.claude/settings.json")
echo "enabled_count=$ENABLED_COUNT"
[ "$ENABLED_COUNT" -ge 1 ] || exit 1
```

**Expected**: `disabled_count=0` AND `enabled_count>=1`, both phase exit codes 0.

**Pass criteria**: Binary — both counts match the expected values.

### AC-HOI-006 — moai doctor reports observability status

**REQ mapping**: REQ-HOI-005

**Given** any MoAI project (with or without `observability.yaml`),
**When** the user runs `moai doctor`,
**Then** the output shall contain a line matching the regex `Observability:\s*(enabled|disabled)`.

**Verification command**:
```bash
moai doctor 2>&1 | grep -E 'Observability:\s*(enabled|disabled)'
echo "exit=$?"
```

**Expected**: Exit code 0, at least one matching line.

**Pass criteria**: Binary — grep matches with one of the two enum values.

### AC-HOI-007 — Template mirror parity

**REQ mapping**: REQ-HOI-001

**Given** the source-of-truth template at `internal/template/templates/.moai/config/sections/observability.yaml`,
**When** the user runs `moai init` and compares the rendered output to the template,
**Then** the rendered file shall be byte-equivalent to the template (modulo expected template variable expansion, which observability.yaml has none of in this SPEC's design).

**Verification command**:
```bash
TEST_DIR=$(mktemp -d -t moai-mirror-XXXXXX)
moai init "$TEST_DIR/proj" --quiet
diff internal/template/templates/.moai/config/sections/observability.yaml \
     "$TEST_DIR/proj/.moai/config/sections/observability.yaml"
echo "exit=$?"
```

**Expected**: Exit code 0, zero diff lines.

**Pass criteria**: Binary — `diff` reports no differences.

## Edge Cases (Verification Notes)

### Edge case 1: Legacy project without observability.yaml

**Given** a pre-v3.0 project that has been upgraded to v3.0 binary but has not yet run `moai update`,
**When** the user invokes any MoAI command that loads config,
**Then** the loader shall treat missing `observability.yaml` as `enabled: false` (zero-value default).

**Verification**: Implicit in AC-HOI-003 — the disabled-path test would also pass on a project missing the file entirely (loader default kicks in). M3 milestone includes an explicit fixture test for this case: `TestObservabilityMissingFile_DefaultsDisabled`.

### Edge case 2: Hand-edited settings.json with hooks present but observability.yaml says disabled

**Given** a project where `observability.yaml` has `enabled: false` but `.claude/settings.json` has been hand-edited to include `handle-harness-observe-stop` references,
**When** the dispatcher routes a `Stop` event,
**Then** the runtime defense-in-depth gating shall still skip the `handle-harness-observe-stop` wrapper.

**Verification**: Covered by `TestObservabilityDisabled` (the test injects a settings.json with the wrappers present and asserts they are not executed). This is the runtime gating contract, complementing the template-side conditional render.

### Edge case 3: ConfigChange hot-reload

**Given** a running MoAI session with `observability.enabled: false`,
**When** the user edits `observability.yaml` to `enabled: true` and the `ConfigChange` hook fires,
**Then** subsequent hook invocations within the same session shall respect the new value without process restart.

**Verification**: Manual smoke test documented in run-phase progress.md. Not part of automated AC set (would require multi-session orchestration). Documented here for completeness.

## Quality Gate Criteria

In addition to the 7 binary ACs above, the run-phase must satisfy:

1. **`go test ./...`** — exits 0, zero NEW test failures over the pre-change baseline.
2. **`golangci-lint run --timeout=2m`** — zero NEW lint issues over the pre-change baseline.
3. **`grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/`** — zero matches (C-HRA-008 subagent boundary).
4. **Cross-platform CI** — darwin + linux + windows runners all green.
5. **Coverage** — `internal/hook/...` package coverage stays at or above the pre-change baseline (no coverage regression).
6. **Frontmatter lint** — `moai spec lint .moai/specs/SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001/spec.md` exits 0 (12-field canonical schema compliance).

## Definition of Done (DoD)

The SPEC is **DONE** when all of the following hold:

- [ ] AC-HOI-001 through AC-HOI-007: all 7 binary ACs PASS
- [ ] All 6 quality gate criteria above PASS
- [ ] R1-R4 mitigations from spec.md § E are in place (release notes draft prepared for R1, ASYNC-EXPAND coordination noted for R2, doctor fallback verified for R3, fixture updates merged for R4)
- [ ] progress.md created in this SPEC directory documenting M1-M3 evidence
- [ ] spec.md frontmatter `status:` updated from `draft` to `implemented` (sync phase will move to `completed`)
- [ ] spec.md `version:` bumped `0.1.0 → 0.2.0` at sync time

## REQ ↔ AC Traceability (mirror of spec.md § C)

| REQ | Mapped ACs | Coverage Type |
|---|---|---|
| REQ-HOI-001 (yaml + key) | AC-HOI-001, AC-HOI-007 | File existence + template mirror parity |
| REQ-HOI-002 (disabled skips) | AC-HOI-003 | Integration test — SKIP path |
| REQ-HOI-003 (enabled executes) | AC-HOI-004 | Integration test — EXECUTE path, schema preserved |
| REQ-HOI-004 (init default + render) | AC-HOI-002, AC-HOI-005 | Init default + settings.json conditional render |
| REQ-HOI-005 (doctor report) | AC-HOI-006 | Doctor diagnostic line |

Coverage: 5 REQs → 7 ACs, 100% bidirectional. Every REQ has at least one mapped AC; every AC traces back to at least one REQ. No orphaned criteria.

## Out of Scope (cross-reference spec.md § E)

The following are explicitly NOT verified by this AC set — they belong to separate SPECs or future work:

### Out of Scope: PostToolUse / SessionStart / PreToolUse opt-in verification
These hooks are not gated by the flag introduced here. No AC tests their always-active status.

### Out of Scope: Async transition behavior verification
Async execution semantics are owned by `SPEC-V3R6-HOOK-ASYNC-EXPAND-001`. This AC set only verifies ON/OFF toggling; it does not verify async vs sync execution timing.

### Out of Scope: Telemetry payload schema validation beyond byte-equivalence
AC-HOI-004 verifies payload schema is unchanged at the field level. It does NOT verify downstream harness consumers correctly parse the payload — that responsibility lies with harness integration tests, which are outside this SPEC's scope.
