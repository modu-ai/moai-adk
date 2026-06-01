---
id: SPEC-V3R6-CI-FLAKY-STABILIZE-002
title: "CI Flaky Test Stabilization — supervisor Watch timeout + quality-gate dotnet timeout (contention-driven)"
version: "0.1.0"
status: in-progress
created: 2026-06-01
updated: 2026-06-01
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/lsp/subprocess, internal/hook/quality"
lifecycle: spec-anchored
tags: "ci, flaky-test, contention, test-artifact, stabilization, timeout"
tier: S
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-01 | manager-spec | Initial plan-phase draft (Tier S LEAN 3-artifact). Follow-up to SPEC-V3R6-CI-FLAKY-STABILIZE-001 (whose §C Out-of-Scope deferred the broader flaky refactor). Two contention-driven flaky tests, both empirically reproduced (7/40 each) and adversarially root-caused as TEST-ARTIFACTS with ZERO production-code change. |

## §A — Context and Problem Statement

This SPEC is the planned follow-up to `SPEC-V3R6-CI-FLAKY-STABILIZE-001`. That predecessor's §C Exclusions explicitly deferred "the 3 pre-existing flaky tests' broader refactor" (verbatim: *"The 3 pre-existing flaky tests' broader refactor (test-isolation overhaul, fixture restructuring beyond the two diagnosed failures) is out of scope."*). This SPEC covers **2 of those 3** residual flaky tests. The 3rd residual flaky remains out of scope (see §C).

Both target tests share a single failure mechanism: each binds its assertion to an **arbitrary 5-second wall-clock** that is unrelated to what the test actually verifies. Under whole-suite CPU/scheduler contention the per-test subprocess cannot be fork/exec'd-and-run within 5s, the 5s deadline fires, and the test fails — while the same test passes deterministically in isolation. Both were empirically reproduced under a CPU-saturation harness (7/40 failures each) and adversarially root-caused: the production code (`supervisor.go`, `client.go`, `gate.go`) is correct; the defect lives entirely in the `*_test.go` scaffolding.

The fix is localized to two `*_test.go` files with **zero production-code change**. Neither test failure was introduced by recent work — both predate it and only manifest under CI whole-suite contention.

### FLAKY-1 — `internal/lsp/subprocess` Supervisor.Watch timeout (CPU-contention)

`supervisor.go` `Watch()` (lines 75-91) runs a goroutine with a two-arm select. The `<-s.doneCh` arm SENDS the real `ExitEvent` (`ch <- s.exitEv`, line 84). The `<-ctx.Done()` arm CLOSES `ch` EMPTY without sending (lines 85-87). The godoc contract (lines 70-72) documents this: *"If ctx is cancelled before the process exits, the channel is closed without an ExitEvent."*

`TestSupervisor_NonZeroExit` arms a fixed 5s context (line 119) and does a BARE receive `ev := <-ch` (line 124), then asserts `ev.ExitCode != 1` (line 125). Under heavy fork+CPU contention the per-test `/bin/sh` `exit 1` stub cannot be scheduled to run-and-exit within 5s, so `ctx.Done()` wins the select, `ch` is closed empty, and the bare receive returns the Go ZERO VALUE `ExitEvent{ExitCode:0}` (built by `buildExitEvent(nil)`, `supervisor.go` lines 107-109). The assertion then observes `ExitCode=0 != 1` and fails at line 125-126 with "ExitCode = 0, want 1". The sibling `TestSupervisor_NormalExit` (lines 93-109) has the same bare-receive pattern (`ev := <-ch` at line 105) and is exposed to the same mechanism.

**Smoking gun**: every failing instance took ~5.04s, exactly equal to the 5s `context.WithTimeout`. Reproduced 7/40 under 16× `yes` CPU burners + concurrent `go build ./...`.

**Classification**: TEST-ARTIFACT. NOT a data race — `s.exitEv` write (`supervisor.go:64`) happens-before `close(s.doneCh)` (`:65`); the read (`:84`) only after `<-s.doneCh` (`:81`); `go test -race` is clean. NOT a production defect — the sole production caller (`internal/lsp/.../client.go` ~lines 359-365) does `case <-exitCh:` and NEVER reads the `ExitEvent` value/`ExitCode`, so a zero-value-on-timeout is harmless in production. `supervisor.go` was never touched by any prior flaky fix; the prior `SPEC-LSP-FLAKY-001/002` fixes targeted launch-time ETXTBSY, an orthogonal mechanism.

### FLAKY-2 — `internal/hook/quality` dotnet-format timeout (CPU-contention)

`TestQualityGate_RunsDotnetFormatWhenCSharpStaged` (lines 670-729) stages a `.cs` file then calls `g.executeStep(ctx, step, 5*time.Second)` (line 724). `executeStep → runStep` wraps the fake `dotnet` shell-script (`#!/bin/sh; exit 0`) subprocess in `context.WithTimeout(ctx, 5s)` (`gate.go:488`) and `cmd.Run()` (`gate.go:496`). Under full-machine CPU saturation + `GOMAXPROCS=1` scheduler starvation, fork/exec of the fake-dotnet subprocess cannot complete within 5s, so `context.DeadlineExceeded` fires (`gate.go:506`), `runStep` returns `(false, "quality gate timed out: dotnet format exceeded 5s")`, and the assertion at line 727 fails ("dotnet format must pass when .cs files are staged").

**Mechanism scope**: only the SINGLE fake-dotnet subprocess runs inside the 5s window; the `git diff --cached --name-only` query (`gate.go:442`) runs EARLIER under the test's outer `context.Background()` with NO deadline, so it is NOT inside the 5s bound. The 4 test-setup git calls (init / config ×2 / add via `runGitCmd`, `gate_test.go:681-702`) have no timeout and run before `executeStep`.

The 5s comes from the explicit timeout ARGUMENT at `gate_test.go:724`; the `GateConfig.LintTimeout: 5*time.Second` at `gate_test.go:720` is INERT (`executeStep` takes timeout as a parameter, never reads `g.config.LintTimeout`). The test verifies ROUTING logic (`changedExts=[".cs"]` filter causes dotnet to execute), NOT timeout/perf behavior, so 5s is arbitrary scaffolding.

The sibling `TestQualityGate_SkipsDotnetFormatWhenNoCSharpStaged` (the no-C# variant, returns early before spawning the dotnet subprocess) is structurally IMMUNE and serves as the immunity control — it MUST NOT be modified.

**Smoking gun**: failure durations cluster 5.06-5.19s. Reproduced 7/40 under 32× `yes` burners + `GOMAXPROCS=1` + concurrent `go test ./internal/...`. Isolation: Runs variant costs ~0.21s (24× headroom).

**Classification**: TEST-ARTIFACT. `gate.go` production code is correct (the `DeadlineExceeded` distinction and routing filter behave as designed). Production already applies a 60s budget to this step (`DefaultGateConfig.LintTimeout = 60s`, `gate.go:46`). The 5s literals + these tests were introduced together (issue #667/#668, 2026-04-17) and never changed; the commit intent is routing verification, not timeout behavior.

## §B — GEARS Requirements

### FLAKY-1 — Supervisor.Watch timeout-vs-event disambiguation (test-only)

**REQ-CFS2-001** (Event-driven): When `TestSupervisor_NonZeroExit` receives from the `Watch` channel, the test shall use a comma-ok receive (`ev, ok := <-ch`) so it can distinguish a delivered `ExitEvent` (`ok == true`) from a context-deadline empty-close (`ok == false`).

**REQ-CFS2-002** (Event-driven): When the `Watch` channel is closed without an `ExitEvent` (`ok == false`, i.e. the context deadline fired before the stub exited under load), `TestSupervisor_NonZeroExit` shall skip via `t.Skip(...)` carrying a load-attribution reason that names the context-deadline-fired-before-stub-exit cause and the no-production-impact note (client.go ignores the `ExitEvent` value).

**REQ-CFS2-003** (Ubiquitous): `TestSupervisor_NormalExit` (the sibling carrying the identical bare-receive pattern) shall apply the same comma-ok receive + load-induced-timeout skip as `TestSupervisor_NonZeroExit`.

**REQ-CFS2-004** (State-driven): While a `Watch` receive delivers a real `ExitEvent` (`ok == true`), `TestSupervisor_NonZeroExit` shall retain its deterministic `ExitCode` assertion so that a genuine wrong exit code (e.g. a stub that really exits 0 when 1 is expected) still fails the test deterministically with "ExitCode = 0, want 1".

**REQ-CFS2-005** (Ubiquitous): The `supervisor.go` production file shall NOT be modified by this SPEC — the disambiguation is read entirely from the comma-ok signal that `Watch()` already emits.

### FLAKY-2 — quality-gate dotnet timeout SLA alignment (test-only)

**REQ-CFS2-006** (Event-driven): When `TestQualityGate_RunsDotnetFormatWhenCSharpStaged` calls `executeStep`, the test shall pass a timeout argument mirroring the production `DefaultGateConfig.LintTimeout` SLA so the test no longer binds its routing assertion to an under-spec wall-clock bound, giving the fake exit-0 subprocess scheduling headroom under whole-suite CPU contention. (The concrete value and the `gate.go` SLA line reference are recorded in plan.md §F M2 as run-phase edit guidance.)

**REQ-CFS2-007** (Ubiquitous): The inert `GateConfig.LintTimeout` field in `TestQualityGate_RunsDotnetFormatWhenCSharpStaged` shall be updated from `5*time.Second` to `60*time.Second` so the test config mirrors the production SLA and does not leave a misleading 5s literal in the test body.

**REQ-CFS2-008** (Ubiquitous): The structurally-immune sibling `TestQualityGate_SkipsDotnetFormatWhenNoCSharpStaged` shall NOT be modified — it returns early before spawning the dotnet subprocess and serves as the immunity control.

**REQ-CFS2-009** (State-driven): While `TestQualityGate_RunsDotnetFormatWhenCSharpStaged` exercises the `.cs`-staged routing path, the test shall retain its deterministic `passed == true` routing assertion so that a genuine routing regression (dotnet NOT executed when `.cs` is staged) still fails the test.

**REQ-CFS2-010** (Ubiquitous): The `gate.go` production file shall NOT be modified by this SPEC.

### Cross-cutting test-only + green gate

**REQ-CFS2-011** (Ubiquitous): This SPEC shall modify ONLY `internal/lsp/subprocess/supervisor_test.go` and `internal/hook/quality/gate_test.go`; `supervisor.go`, `client.go`, and `gate.go` shall remain byte-identical to HEAD.

**REQ-CFS2-012** (Event-driven): When `go test ./...` runs locally (darwin), the two affected packages shall pass deterministically (at minimum under `-count=5`), establishing the local green gate.

## §C — Exclusions (What NOT to Build)

### Out of Scope — the 3rd residual flaky test
This SPEC covers 2 of the 3 pre-existing flaky tests deferred by `SPEC-V3R6-CI-FLAKY-STABILIZE-001` §C. The **3rd residual flaky test remains out of scope** here and is deferred to a future SPEC. This SPEC MUST NOT expand to stabilize the 3rd flaky.

### Out of Scope — production-code change
This SPEC MUST NOT modify `supervisor.go`, `client.go`, or `gate.go`. The diagnosed defects are test-artifacts; the production code is correct. Any "fix" touching production code would be out of scope and would contradict the root-cause analysis.

### Out of Scope — the immune sibling test
`TestQualityGate_SkipsDotnetFormatWhenNoCSharpStaged` is structurally immune (returns early before spawning the dotnet subprocess) and MUST NOT be modified. It is cited only as the immunity control.

### Out of Scope — broader test-isolation overhaul
Restructuring the `t.Parallel()` topology of `supervisor_test.go` or `gate_test.go`, introducing a CPU-contention test harness into CI, fixture restructuring, or a generalized timeout-config refactor are all out of scope. This SPEC applies the smallest robust change to the two diagnosed tests only.

### Out of Scope — predecessor's two failures
`SPEC-V3R6-CI-FLAKY-STABILIZE-001` (FLAKY-1 internal/spec git-add race, FLAKY-2 Windows merge-TUI hang) is `status: completed`. This SPEC MUST NOT re-touch `internal/spec`, `internal/cli`, or `internal/merge`.

### Out of Scope — implementation details
Per SPEC scope discipline, this document specifies observable behaviors and constraints, not exact diff line counts or helper-function names. The precise edits are decided at run-phase (recommended targets are recorded in plan.md as guidance).
