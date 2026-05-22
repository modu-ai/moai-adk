---
id: SPEC-V3R6-HOOK-ASYNC-EXPAND-001
title: "Hook async 확대 — Acceptance Criteria (Tier M, 8 binary ACs)"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/hook, internal/hook/testutil"
lifecycle: spec-anchored
tags: "hook, async, acceptance, sprint-2"
tier: M
---

# Acceptance Criteria — SPEC-V3R6-HOOK-ASYNC-EXPAND-001

## Overview

This document defines 8 binary acceptance criteria (AC-HAE-001..008) for the hook async expansion. Each AC is binary (PASS/FAIL), reproducible via the listed verification command, and traceable to one or more EARS requirements in `spec.md` § 3.

Every AC MUST PASS before the run-phase PR is eligible for merge.

## Traceability Matrix (REQ ↔ AC)

| REQ | Coverage ACs | Notes |
|-----|--------------|-------|
| REQ-HAE-001 (FileChanged async) | AC-HAE-001, AC-HAE-002 | Goroutine launch + benchmark |
| REQ-HAE-002 (ConfigChange async) | AC-HAE-001, AC-HAE-003 | Goroutine launch + benchmark |
| REQ-HAE-003 (TaskCreated conditional) | AC-HAE-001, AC-HAE-004 | Goroutine launch + benchmark + observability matrix |
| REQ-HAE-004 (Notification conditional) | AC-HAE-001, AC-HAE-005 | Goroutine launch + benchmark + observability matrix |
| REQ-HAE-005 (deadline enforcement) | AC-HAE-007 | Goroutine leak verification (zero leaks confirms self-cancel) |
| REQ-HAE-006 (WaitForAsync helper) | AC-HAE-008 | Helper exists + usage count ≥ 4 |
| (cross-cutting safety) | AC-HAE-006 | Race detection |

100% REQ coverage. 100% AC has at least one REQ ancestor.

## Given-When-Then Scenarios (canonical)

### Scenario 1: FileChanged responds within 100 ms regardless of MX scan duration

**Given** the `FileChanged` hook handler is registered and a Go source file change is reported  
**When** the handler is invoked with a workload that would synchronously take 500 ms+ for MX delta scan  
**Then** the handler returns `{"continue": true, "exitCode": 0}` within ≤ 100 ms (p95 across 10 concurrent invocations)  
**And** the MX delta scan completes asynchronously in a background goroutine within the 5-second deadline.

### Scenario 2: TaskCreated zero-overhead when observability disabled

**Given** `.moai/config/sections/observability.yaml` sets `observability.enabled: false` (v3.0 default)  
**When** the `TaskCreated` hook handler is invoked  
**Then** the handler returns the empty payload immediately  
**And** no goroutine is launched (verified via `goleak` baseline assertion)  
**And** no entry is written to `.moai/state/task-events.jsonl`.

### Scenario 3: Goroutine deadline self-cancellation logs warning

**Given** an async side-effect is artificially delayed beyond the 5-second deadline (test injection)  
**When** the goroutine's `context.WithTimeout` triggers cancellation  
**Then** a structured warning entry is appended to `.moai/state/hook-async-warnings.jsonl`  
**And** the goroutine self-terminates without leaking  
**And** the main handler response (already returned at ≤ 100 ms) is unaffected.

## Binary Acceptance Criteria (8)

### AC-HAE-001 — Goroutine launch in all 4 target handlers

**Verification command**:
```bash
grep -E 'go func\(\)' internal/hook/file_changed.go internal/hook/config_change.go internal/hook/task_created.go internal/hook/notification.go | wc -l
```

**Expected output**: integer ≥ 4 (one `go func()` per handler; TaskCreated/Notification may have the launch behind an `if IsObservabilityEnabled()` block).

**PASS criterion**: stdout ≥ 4.

**Traceability**: REQ-HAE-001, REQ-HAE-002, REQ-HAE-003, REQ-HAE-004.

### AC-HAE-002 — FileChanged handler p95 ≤ 100 ms under 10-concurrent load

**Verification command**:
```bash
go test -run '^$' -bench 'BenchmarkFileChanged_AsyncReturn' -benchtime 10x -count 5 ./internal/hook/
```

**Expected output**: benchmark report shows `b.ReportMetric` for p95 ≤ 100 ms (100000000 ns) under 10-concurrent goroutine invocation. Benchmark MUST register the metric via `b.ReportMetric(p95Ms, "p95-ms")` so the value is grep-able.

**PASS criterion**: parsed `p95-ms` metric is ≤ 100.

**Traceability**: REQ-HAE-001.

### AC-HAE-003 — ConfigChange handler p95 ≤ 100 ms under 10-concurrent load

**Verification command**:
```bash
go test -run '^$' -bench 'BenchmarkConfigChange_AsyncReturn' -benchtime 10x -count 5 ./internal/hook/
```

**Expected output**: parsed `p95-ms` metric ≤ 100.

**PASS criterion**: stdout `p95-ms` field ≤ 100.

**Traceability**: REQ-HAE-002.

### AC-HAE-004 — TaskCreated handler p95 ≤ 100 ms with observability.enabled=true

**Verification command**:
```bash
MOAI_TEST_OBSERVABILITY_ENABLED=true go test -run '^$' -bench 'BenchmarkTaskCreated_AsyncReturn' -benchtime 10x -count 5 ./internal/hook/
```

The benchmark MUST use a test fixture that sets `observability.enabled=true` (either via the env var above + test setup, or a `t.Setenv`-style equivalent inside `BenchmarkTaskCreated_AsyncReturn`'s `b.B.Setenv` if Go version permits).

**Expected output**: parsed `p95-ms` metric ≤ 100.

**PASS criterion**: stdout `p95-ms` field ≤ 100.

**Bonus verification (separate test, not benchmark)**: `TestTaskCreated_ObservabilityDisabled_NoGoroutine` — when observability is false, no goroutine is spawned. Verified via test-local goleak snapshot delta.

**Traceability**: REQ-HAE-003.

### AC-HAE-005 — Notification handler p95 ≤ 100 ms with observability.enabled=true

**Verification command**:
```bash
MOAI_TEST_OBSERVABILITY_ENABLED=true go test -run '^$' -bench 'BenchmarkNotification_AsyncReturn' -benchtime 10x -count 5 ./internal/hook/
```

**Expected output**: parsed `p95-ms` metric ≤ 100.

**PASS criterion**: stdout `p95-ms` field ≤ 100.

**Bonus verification**: `TestNotification_ObservabilityDisabled_NoGoroutine` mirrors AC-HAE-004 bonus.

**Traceability**: REQ-HAE-004.

### AC-HAE-006 — Race detection clean across all hook tests

**Verification command**:
```bash
go test -race -count=1 ./internal/hook/...
```

**Expected output**: final line `ok  github.com/modu-ai/moai-adk/internal/hook ...`, no `DATA RACE` warnings anywhere in stdout/stderr, exit 0.

**PASS criterion**: exit 0 AND zero `DATA RACE` occurrences in output.

**Traceability**: REQ-HAE-001..004 (cross-cutting safety for all four async handlers).

### AC-HAE-007 — Zero goroutine leaks via goleak.VerifyTestMain

**Verification command**:
```bash
go test -count=1 -timeout=30s ./internal/hook/...
```

**Expected output**: `goleak.VerifyTestMain(m)` in `internal/hook/main_test.go` reports no leaks. If a leak occurs, goleak prints the leaking goroutine stack and the test binary exits non-zero.

**PASS criterion**: exit 0 AND no `goroutine leak detected` stderr output.

**Traceability**: REQ-HAE-005 (deadline self-cancellation prevents leaks; goleak is the verification).

### AC-HAE-008 — WaitForAsync helper exists and used by ≥ 4 test files

**Verification commands** (two parts, both MUST PASS):

Part 1 — Helper file exists:
```bash
test -f internal/hook/testutil/wait_async.go && echo "EXISTS" || echo "MISSING"
```

Expected: `EXISTS`.

Part 2 — Used by ≥ 4 test files:
```bash
grep -l 'testutil.WaitForAsync\|WaitForAsync(' internal/hook/file_changed_test.go internal/hook/config_change_test.go internal/hook/task_created_test.go internal/hook/notification_test.go | wc -l
```

Expected: integer ≥ 4.

**PASS criterion**: Part 1 `EXISTS` AND Part 2 stdout ≥ 4.

**Traceability**: REQ-HAE-006.

## Edge Cases

### EC1: Goroutine that completes within 100 ms (no deadline triggering)

When the side-effect work is fast (e.g., MX scan on a 5-line file), the goroutine completes well within 5 seconds. AC-HAE-007 PASS even with the goroutine fully completing — goleak verifies absence of LEFTOVER goroutines, not absence of goroutines entirely.

### EC2: Multiple rapid invocations of the same hook event

If 100 `FileChanged` events arrive in 1 second, 100 goroutines spawn in parallel. AC-HAE-002 covers 10-concurrent; for higher concurrency, the OS thread pool may saturate but AC-HAE-006 + AC-HAE-007 still PASS because each goroutine self-cancels via its own context. No global mutex contention.

### EC3: `observability.yaml` missing or unparseable (REQ-HAE-003/004 fallback)

When the config file is missing or YAML-invalid: `IsObservabilityEnabled()` returns `false` (safe default — zero overhead). The handler returns the empty payload immediately. AC-HAE-001 grep STILL matches (the `go func()` source line is present in the file even if conditionally skipped at runtime).

### EC4: Goroutine encounters panic in side-effect

The goroutine MUST `defer recover()` and log to `.moai/state/hook-async-warnings.jsonl` (similar to REQ-HAE-005 deadline log but with `"reason":"panic"`). Panics MUST NOT propagate to the main handler's response. This is implied by REQ-HAE-005's design intent but not explicitly verified by a binary AC — non-binary edge case acceptable here; recommended for future enhancement.

## Quality Gate Criteria

Beyond the 8 binary ACs above, the run-phase must satisfy:

- TRUST 5 framework: all 5 dimensions PASS for changed files.
- Coverage: `internal/hook/` package coverage ≥ 85% after this SPEC's changes (verify via Section E.3 in plan.md).
- Lint: 0 NEW `golangci-lint` issues (pre-existing baseline preserved).
- Cross-platform build: PASS on darwin, linux, windows.
- Subagent boundary (C-HRA-008): zero `AskUserQuestion` / `mcp__askuser` matches in `internal/hook/` production code.

## Definition of Done

The SPEC is DONE when:

1. All 8 binary ACs PASS (AC-HAE-001..008) with reproducible evidence captured in `progress.md`.
2. Quality gates above PASS.
3. All 5 milestones (M1..M5 in plan.md § F) complete with commit SHA recorded.
4. PR merged to `main` (Hybrid Trunk Tier M doctrine: feat-branch + squash merge).
5. `spec.md` frontmatter updated: `status: implemented`, `version: 0.2.0`.
6. `progress.md` written with M1~M5 evidence + AC PASS matrix + cross-Wave caveats noted (e.g., HOOK-OBSERVE-OPT-IN-001 merge confirmation, baseline lint drift if any).
7. Dependency partner SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 is in MERGED state (verified via `gh pr view ... --json state`).

## Out of Scope (mirrors spec.md § 6)

### Out of Scope: Single mega-dispatcher consolidation
Verification of `moai hook dispatch <event>` entry-point unification is NOT part of this SPEC's AC set.

### Out of Scope: External queue / message broker
No AC verifies Redis/Kafka/disk-queue integration. In-process goroutine + `context.WithTimeout` only.

### Out of Scope: Async transition for other hook events
ACs do not verify Stop/SubagentStop/UserPromptSubmit/PreToolUse/SessionStart/WorktreeCreate/WorktreeRemove behavior — those events retain their current sync/async profile and any verification of their behavior is owned by their respective SPECs (e.g., SPEC-V3R6-HOOK-CONTRACT-FIX-001 for WorktreeCreate/Remove).
