---
id: SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001
title: "Hook cwd leak audit + resolveProjectRoot consistency — Acceptance Criteria"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
tier: S
---

# SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001 — Acceptance Criteria

## Overview

This document defines 7 binary acceptance criteria (ACs) that must all PASS for the SPEC to be marked `status: implemented`. Each AC has a single, verifiable command and a single binary pass/fail outcome. Ambiguity is intentionally minimized.

---

## AC-HCWA-001 — Zero `os.Getwd()` leaks in production hook code

**Requirement linkage**: REQ-HCWA-001, REQ-HCWA-009

**Given** the working tree is at the post-M3 commit
**When** the audit grep command is executed
**Then** the count of `os.Getwd()` matches in non-test files under `internal/hook/` (excluding the two helper definition files `path_resolve.go` and `quality/path_resolve.go` which legitimately call `os.Getwd()` as the documented last-resort fallback) shall be exactly 0

**Verification command**:
```bash
test "$(grep -rn 'os\.Getwd' internal/hook/ | grep -v '_test.go' | grep -v 'path_resolve\.go' | grep -v '^[^:]*:[0-9]*:[ \t]*//' | wc -l)" -eq 0 && echo "AC-HCWA-001 PASS" || echo "AC-HCWA-001 FAIL"
```

**Pass condition**: Output is `AC-HCWA-001 PASS`
**Fail condition**: Output is `AC-HCWA-001 FAIL` OR any remaining match in non-helper production source

**Exclusion rationale** (refined post-M3):
1. **Helper definitions** (`path_resolve.go` × 2): REQ-HCWA-001 mandates "Direct `os.Getwd()` calls outside these helpers shall not exist." The two helper files (`internal/hook/path_resolve.go` and `internal/hook/quality/path_resolve.go`) MUST contain exactly one `os.Getwd()` call each as the documented last-resort fallback after env-var and (for quality) cfg.ProjectDir resolution. These are the contractual implementation of REQ-HCWA-001 itself, NOT instances of the leak class being audited.
2. **Comments referencing the fallback behavior**: After refactoring, several call-site comments still describe the env-var → `os.Getwd()` fallback chain as the documented contract (e.g., `// env var → os.Getwd() fallback with slog.Warn cwd_fallback:true marker`). These are documentation references, not active call sites, and the `^[^:]*:[0-9]*:[ \t]*//` filter excludes them.

---

## AC-HCWA-002 — Helper call site coverage

**Requirement linkage**: REQ-HCWA-001, REQ-HCWA-003, REQ-HCWA-004, REQ-HCWA-005, REQ-HCWA-006, REQ-HCWA-007

**Given** the working tree is at the post-M3 commit
**When** the helper-call-site grep is executed
**Then**:
- `resolveProjectRootFromEnv` OR `resolveProjectRootFromInputOrEnv` shall be called from at least 5 sites in `internal/hook/*.go` (non-test files)
- `resolveQualityProjectDir` shall be called from at least 4 sites in `internal/hook/quality/*.go` (non-test files)

**Verification command**:
```bash
ENV_COUNT=$(grep -rn "resolveProjectRootFromEnv\|resolveProjectRootFromInputOrEnv" internal/hook/ | grep -v "_test.go" | grep -v "func resolveProject" | wc -l)
QUAL_COUNT=$(grep -rn "resolveQualityProjectDir" internal/hook/quality/ | grep -v "_test.go" | grep -v "func resolveQual" | wc -l)
test "$ENV_COUNT" -ge 5 && test "$QUAL_COUNT" -ge 4 && echo "AC-HCWA-002 PASS ($ENV_COUNT env / $QUAL_COUNT quality)" || echo "AC-HCWA-002 FAIL ($ENV_COUNT env / $QUAL_COUNT quality)"
```

**Pass condition**: Output is `AC-HCWA-002 PASS (5 env / 4 quality)` or higher
**Fail condition**: Either count below threshold

**Notes**:
- The `grep -v "func resolveProject"` and `grep -v "func resolveQual"` filters exclude the helper definition lines themselves so only call sites are counted.
- M1 contributes 2 sites (subagent_start.go × 2), M2 contributes 3 sites (pre_tool.go × 2 + observability_master.go × 1) → total ≥5 for env helpers.
- M3 contributes 4 sites (quality/gate.go × 4) for quality helper.

---

## AC-HCWA-003 — HOI cohabitation guard test continues to pass

**Requirement linkage**: REQ-HCWA-010 (PRESERVE)

**Given** the working tree is at the post-M3 commit
**When** the cohabitation guard test is executed
**Then** all test cases shall pass without any modification to `cohabitation_guard_test.go`

**Verification command**:
```bash
# Verify the file is byte-identical to baseline
EXPECTED_HASH=$(grep "cohabitation_guard_test.go" /tmp/preserve-baseline.sha256 | awk '{print $1}')
ACTUAL_HASH=$(sha256sum internal/hook/cohabitation_guard_test.go | awk '{print $1}')
test "$EXPECTED_HASH" = "$ACTUAL_HASH" || { echo "AC-HCWA-003 FAIL — file modified"; exit 1; }

# Run the test
go test -run TestCohabitationGuard -count=1 ./internal/hook/... && echo "AC-HCWA-003 PASS" || echo "AC-HCWA-003 FAIL"
```

**Pass condition**: Output is `AC-HCWA-003 PASS` AND file hash unchanged
**Fail condition**: File hash changed OR test failure

---

## AC-HCWA-004 — Race detector clean

**Requirement linkage**: REQ-HCWA-008 (logging safety)

**Given** the working tree is at the post-M3 commit
**When** the full hook test suite is run with race detection
**Then** the test run shall exit with code 0 and no new race detector warnings

**Verification command**:
```bash
go test -race -count=1 ./internal/hook/... 2>&1 | tee /tmp/m3-race.log
EXIT=${PIPESTATUS[0]}
RACE_COUNT=$(grep -c "WARNING: DATA RACE" /tmp/m3-race.log || echo 0)
test "$EXIT" -eq 0 && test "$RACE_COUNT" -eq 0 && echo "AC-HCWA-004 PASS" || echo "AC-HCWA-004 FAIL (exit=$EXIT race=$RACE_COUNT)"
```

**Pass condition**: Output is `AC-HCWA-004 PASS`, exit 0, 0 race warnings
**Fail condition**: Non-zero exit OR any new race warning

---

## AC-HCWA-005 — Lint baseline 0 NEW issues

**Requirement linkage**: Quality gate per `.moai/config/sections/quality.yaml`

**Given** the M0 lint baseline was captured at plan-phase entry to `/tmp/m0-lint-baseline.txt`
**When** golangci-lint is re-run on the post-M3 working tree
**Then** the count of WARNING/ERROR lines shall not exceed the M0 baseline count

**Verification command**:
```bash
golangci-lint run --timeout=2m ./internal/hook/... 2>&1 | tee /tmp/m3-lint-final.txt
M0_COUNT=$(grep -cE "^(internal/hook|.*\.go:[0-9]+:[0-9]+:)" /tmp/m0-lint-baseline.txt 2>/dev/null || echo 0)
M3_COUNT=$(grep -cE "^(internal/hook|.*\.go:[0-9]+:[0-9]+:)" /tmp/m3-lint-final.txt 2>/dev/null || echo 0)
test "$M3_COUNT" -le "$M0_COUNT" && echo "AC-HCWA-005 PASS (M0=$M0_COUNT M3=$M3_COUNT)" || echo "AC-HCWA-005 FAIL (M0=$M0_COUNT M3=$M3_COUNT — $((M3_COUNT - M0_COUNT)) NEW)"
```

**Pass condition**: Output is `AC-HCWA-005 PASS (M0=N M3=N or less)`
**Fail condition**: M3 issue count > M0 baseline

**Notes**:
- "0 NEW" is interpreted relative to the M0 baseline captured at plan-phase entry, NOT absolute zero. The codebase has pre-existing lint findings unrelated to this SPEC.
- If M0 baseline cannot be loaded (`/tmp/m0-lint-baseline.txt` missing), the AC defaults to FAIL and the implementer must re-capture baseline.

---

## AC-HCWA-006 — `subagent_stop.go` PRESERVE (byte-identical)

**Requirement linkage**: REQ-HCWA-010

**Given** `internal/hook/subagent_stop.go` was hashed at plan-phase entry
**When** the file is hashed at post-M3
**Then** the SHA-256 hash shall be unchanged

**Verification command**:
```bash
EXPECTED_HASH=$(grep "subagent_stop.go" /tmp/preserve-baseline.sha256 | awk '{print $1}')
ACTUAL_HASH=$(sha256sum internal/hook/subagent_stop.go | awk '{print $1}')
test "$EXPECTED_HASH" = "$ACTUAL_HASH" && echo "AC-HCWA-006 PASS" || echo "AC-HCWA-006 FAIL (expected=$EXPECTED_HASH actual=$ACTUAL_HASH)"
```

**Pass condition**: Output is `AC-HCWA-006 PASS`
**Fail condition**: Hash mismatch

---

## AC-HCWA-007 — `post_tool_metrics.go` `resolveProjectRoot` function PRESERVE

**Requirement linkage**: REQ-HCWA-002, REQ-HCWA-010

**Given** the `resolveProjectRoot` function in `internal/hook/post_tool_metrics.go` was the reference pattern at plan-phase entry
**When** the function body is extracted at post-M3
**Then** the function signature, body, and `.moai/` existence guard shall be unchanged

**Verification command**:
```bash
# Extract the function body (from "func resolveProjectRoot" to next "func " or EOF)
awk '/^func resolveProjectRoot/{flag=1} flag{print} /^}/{if(flag){flag=0}}' internal/hook/post_tool_metrics.go > /tmp/m3-resolveProjectRoot.txt
EXPECTED_LINES=$(grep -c "" /tmp/m3-resolveProjectRoot.txt)
test "$EXPECTED_LINES" -ge 15 && test "$EXPECTED_LINES" -le 17 || { echo "AC-HCWA-007 FAIL (line count $EXPECTED_LINES not in 15..17)"; exit 1; }
grep -q 'os.Getenv(config.EnvClaudeProjectDir)' /tmp/m3-resolveProjectRoot.txt && \
grep -q 'os.Stat(filepath.Join(root, ".moai"))' /tmp/m3-resolveProjectRoot.txt && \
echo "AC-HCWA-007 PASS" || echo "AC-HCWA-007 FAIL — body signature drifted"
```

**Pass condition**: Output is `AC-HCWA-007 PASS`; function still has 15-17 lines, contains the env-var lookup and the `.moai/` Stat guard
**Fail condition**: Function modified (line count drift, missing env-var lookup, or missing guard)

**Notes**:
- AC-HCWA-007 deliberately tests the function body shape rather than a strict hash, because the reference pattern's location in the file (line number 98) may shift if M1/M2 add helpers above or below it. The function shape is what's frozen.

---

## Summary Matrix

| AC | Requirement | Verification | Trigger |
|----|-------------|--------------|---------|
| AC-HCWA-001 | REQ-HCWA-001, 009 | `grep -rn os.Getwd internal/hook/ | grep -v _test.go | wc -l == 0` | M3 final |
| AC-HCWA-002 | REQ-HCWA-001, 003-007 | Helper call site count >= 5+4 | M3 final |
| AC-HCWA-003 | REQ-HCWA-010 | Cohabitation guard test PASS + file hash unchanged | M3 final |
| AC-HCWA-004 | REQ-HCWA-008 | `go test -race ./internal/hook/...` exit 0, 0 races | M3 final |
| AC-HCWA-005 | Quality gate | golangci-lint <= M0 baseline | M3 final |
| AC-HCWA-006 | REQ-HCWA-010 | subagent_stop.go SHA-256 unchanged | M3 final |
| AC-HCWA-007 | REQ-HCWA-002, 010 | resolveProjectRoot signature + .moai/ guard preserved | M3 final |

**All 7 ACs must PASS for status: implemented transition.**

---

## Definition of Done

The SPEC is considered DONE when:

1. All 7 ACs above produce PASS output
2. `spec.md` frontmatter is updated: `status: draft → implemented`, `version: "0.1.0" → "0.2.0"`
3. `progress.md` tracks M1, M2, M3 as completed with commit hashes
4. `test-getwd-inventory.txt` is written to SPEC directory (REQ-HCWA-011 deliverable)
5. M3 final commit message includes the SPEC ID, all 3 milestones, and the AC summary line
6. orchestrator-driven `git push origin main` (or PR) succeeds (manager-develop B9 boundary: subagent does NOT push; orchestrator pushes)

---

Version: 0.2.0
Status: implemented
Last Updated: 2026-05-23
