---
id: SPEC-CODEX-DEBUG-MODE-001
title: "Acceptance criteria — codex launcher debug mode"
version: "0.1.0"
created: 2026-10-01
updated: 2026-10-01
author: manager-spec
module: "internal/cli"
tier: M
---

# SPEC-CODEX-DEBUG-MODE-001 — Acceptance

## §D — AC Matrix

All scenarios run against stubbed launch seams (`codexDirectLaunchFn`-style) — no real `codex`/`claude`/`glm` process is launched in any test. Verification commands use plain, file-anchored forms (worktree-guard safe) and `-run '^Name$'` anchored patterns. Each release-blocking AC names its RED-now cell in the § Evidence Ledger.

### AC-001 — Codex head `-d` is stripped, never forwarded (release-blocking, RED: RED-1)

- **Given** the codex launcher parse over the head `["cli", "-d"]` with no `--`
- **When** the launch assembles child arguments through the stubbed seam
- **Then** the child argv contains no `-d` token, and the debug trace is active.
- Verify: `go test -count=1 -run '^TestCodexDebugTokenStrippedFromChildArgs$' ./internal/cli/...`

### AC-002 — `--debug` spelling strips identically (regression-guard)

- **Given** the codex launcher parse over the head `["-w", "wt-a", "--debug"]`
- **When** the launch assembles child arguments
- **Then** the child argv contains no `--debug` token and the trace is active.
- Verify: `go test -count=1 -run '^TestCodexDebugTokenStrippedFromChildArgs$' ./internal/cli/...` (same table, `--debug` row)

### AC-003 — Post-`--` debug token is verbatim codex args (regression-guard)

- **Given** the invocation head `["cli", "--", "-d"]`
- **When** the launch assembles child arguments
- **Then** the child argv carries `-d` after `--` verbatim and the trace is NOT activated.
- Verify: `go test -count=1 -run '^TestCodexDebugTokenAfterDashDashForwarded$' ./internal/cli/...`

### AC-004 — Debug token on a readout verb refuses with a named diagnostic (release-blocking, RED: RED-1)

- **Given** the invocation `status -d` (readout verb, pre-`--` token)
- **When** the launcher routes the verb
- **Then** it prints the named refusal diagnostic constant (the `codexSpawnReadoutDiag` shape) and returns exit code 1, starting nothing.
- Verify: `go test -count=1 -run '^TestCodexDebugTokenRefusedOnReadoutVerb$' ./internal/cli/...`

### AC-005 — cc launch is observe-only: child keeps `-d`, launcher traces (release-blocking, RED: RED-3)

- **Given** a cc launch invocation `["-d", "--", "extra"]` with the launch seam stubbed
- **When** the cc launcher runs the launch path
- **Then** the child argv still contains `-d` (unmodified), and stderr carries the launcher debug trace lines.
- Verify: `go test -count=1 -run '^TestCCDebugTokenObserveOnly$' ./internal/cli/...`

### AC-006 — glm launch is observe-only under the same contract (release-blocking, RED: RED-3)

- **Given** a glm launch invocation `["-d"]` with the launch seam stubbed
- **When** the glm launcher runs the launch path
- **Then** the child argv still contains `-d` and stderr carries the launcher debug trace lines.
- Verify: `go test -count=1 -run '^TestGLMDebugTokenObserveOnly$' ./internal/cli/...`

### AC-007 — Three-runner uniformity matrix (release-blocking, RED: RED-5)

- **Given** the cc, glm, and codex launchers, each invoked once with `-d` and once with `--debug` (pre-`--`), seams stubbed
- **When** the matrix runs over all six combinations
- **Then** all three accept both spellings pre-`--`, all three emit trace lines to stderr under the same prefix constant, and the per-backend child linkage holds (cc/glm forward the token; codex strips it and injects `RUST_LOG=debug`).
- Verify: `go test -count=1 -run '^TestThreeRunnerDebugUniformityMatrix$' ./internal/cli/...`

### AC-008 — One trace line per executed pre-exec step (release-blocking, RED: RED-6)

- **Given** a codex launch under debug mode with the seam stubbed
- **When** the launch completes its pre-exec path
- **Then** stderr carries exactly one line per executed step from the named vocabulary (binary resolution, project-root resolution, init-offer gate, local-instruction load, lane claim when a factory lane, worktree materialization when `-w`, child-env assembly, exec handoff), each line under the single prefix constant.
- Verify: `go test -count=1 -run '^TestCodexDebugTraceStepLines$' ./internal/cli/...`

### AC-009 — Env keys named, env values never emitted (release-blocking, RED: RED-7)

- **Given** a traced launch with a sentinel value (e.g. `MOAI_KANBAN_LEAD_ADDR=leader-secret-sentinel`) present in the environment, all lane axes pinned via `t.Setenv`
- **When** the child-env assembly step is traced
- **Then** the trace output contains the key name `MOAI_KANBAN_LEAD_ADDR` with its presence, and the sentinel value string appears nowhere in the captured stderr.
- Verify: `go test -count=1 -run '^TestCodexDebugTraceEnvKeysOnly$' ./internal/cli/...`

### AC-010 — Worktree entry traced with resolved dir and lock outcome (release-blocking, RED: RED-8)

- **Given** an existing worktree fixture under `t.TempDir()` and a codex launch `["-w", "<fixture>", "-d"]`
- **When** the worktree materialization step runs
- **Then** the trace names the resolved directory, the writer-check outcome, and the anchor-lock outcome.
- Verify: `go test -count=1 -run '^TestCodexDebugTraceWorktreeEntry$' ./internal/cli/...`

### AC-011 — RUST_LOG injected when absent, preserved when present (release-blocking, RED: RED-2)

- **Given** (a) a traced launch with `RUST_LOG` unset, and (b) a traced launch with `RUST_LOG=info` set (pinned via `t.Setenv`)
- **When** the child env assembles in each scenario
- **Then** (a) the child env contains `RUST_LOG=debug`, and (b) the child env keeps `RUST_LOG=info` unmodified.
- Verify: `go test -count=1 -run '^TestCodexDebugRustLogInjection$' ./internal/cli/...`

### AC-012 — Debug lane launch prints step timings regardless of threshold (release-blocking, RED: RED-4)

- **Given** a codex lane launch (`-f lane`) under debug mode with all pre-exec steps stubbed to complete in well under the slow-launch threshold
- **When** the pre-exec path reaches the seam
- **Then** the per-step timing lines print unconditionally (the threshold does not gate them), and the lines come from the shared collector's recorded steps — a single recording site.
- Verify: `go test -count=1 -run '^TestCodexDebugSupersedesLaunchThreshold$' ./internal/cli/...`

### AC-013 — Debug-off behavior of t1378 REQ-012 is frozen (regression-guard)

- **Given** debug tokens absent, (a) a fast lane launch and (b) a lane launch exceeding the threshold (threshold pinned via `t.Setenv` on `MOAI_FACTORY_SLOW_LAUNCH_MS`)
- **When** each launch reaches the seam
- **Then** (a) prints no timing lines and no trace lines, and (b) prints exactly the t1378 REQ-012 (SPEC-CODEX-LANE-SLOTS-001) threshold report — and `internal/cli/factory_launch_timing_test.go` passes with no modification.
- Verify (part 1 — frozen behavior): `go test -count=1 -run '^TestCodexDebugOffKeepsTimingReportFrozen$' ./internal/cli/...`
- Verify (part 2 — existing tests unmodified): `go test -count=1 -run '^(TestFactoryLaunchTimingNilSafe|TestFactoryLaunchTimingMeasuresElapsed|TestCodexLaneLaunchTimingNamesSteps|TestCodexLaneLaunchTimingQuietUnderThreshold)$' ./internal/cli/...`

### AC-014 — MOAI_LOG_LEVEL does not gate the trace (regression-guard)

- **Given** a traced codex launch with `MOAI_LOG_LEVEL=error` pinned via `t.Setenv`
- **When** the launch runs
- **Then** the trace lines still print (level does not suppress them).
- Verify: `go test -count=1 -run '^TestCodexDebugTraceIgnoresLogLevel$' ./internal/cli/...`

### AC-015 — Debug trace completes before the exec seam on both doors (release-blocking, RED: RED-9)

- **Given** debug mode on, the direct door selected, and the platform launch seam stubbed to record the relative order of (a) the trace sink's writes and (b) the seam invocation
- **When** the launch runs through the pre-exec path
- **Then** every trace line is written before the seam function is invoked on the POSIX door (`codexDirectLaunchFn` → `syscall.Exec`, `internal/cli/codex_direct_posix.go:53`), and the Windows door's stubbed launch entry asserts the same already-flushed ordering.
- Verify: `go test -count=1 -run '^TestCodexDebugTracePrecedesExecSeam$' ./internal/cli/...`

## § Evidence Ledger (RED-now observations, measured at `49a42c2fc`, branch `WT-codex-debug-mode`)

Carrier form (uniform per cell): **Command** / **Verbatim output** (fenced) / **Exit** / **Measured at** / **Positive control** (mandatory for every zero-hit observation — a zero without a positive control is unmeasured, t1133) / **Red-reason**. No probe file is cited: every observation below is an invocation, a grep, or an anchored test-absence selector, so the ledger carries no scratch-and-delete history. The anchored test-absence selectors are single-invocation and re-executable — after implementation the identical command runs the real test — so their ACs keep release-blocking classification (a green run of an absent selector would be vacuous; the red-reason states this per cell).

### RED-1 — AC-001 + AC-004: the codex launcher has no debug-token recognition (invocation RED)

- Command: `go run ./cmd/moai codex -d status`
- Verbatim output:
```
unknown verb - usage: moai codex [cli] [-w <worktree>] [-f [lane|lane-<n>]] [--factory-run <id>] [-- codex-args...] | moai codex status | moai codex app
exit status 1
```
- Exit: 1 (`go run` wrapper reports `exit status 1`; raw binary exit is 1 via the launcher's `exitCodeError` discipline)
- Measured at: `49a42c2fc`
- Companion grep (token absence in the launcher source): `grep -n '"-d"\|--debug' internal/cli/codex_launcher.go` → no output, exit 1. Positive control: `grep -n "codexSpawnReadoutDiag" internal/cli/codex_launcher.go` → hits at `:70`, `:72`, `:721`, exit 0.
- Red-reason: `codexCmd` carries `DisableFlagParsing: true` (`internal/cli/codex_launcher.go:641`), so the head token `-d` reaches `runCodex` raw and dies at the closed-set verb lookup (`codexUsageDiag`, `:60`) BEFORE any launch-path side effect — the probe starts nothing and writes nothing, which is what makes the invocation probe the honest RED form. The observed refusal is the GENERIC unknown-verb usage line, not a named debug-token diagnostic: REQ-001 (token recognition + strip) and REQ-003 (named readout refusal) are not implemented.

### RED-2 — AC-011: no RUST_LOG handling exists anywhere in the repo

- Command: `grep -rn "RUST_LOG" internal/ pkg/ cmd/`
- Verbatim output: (none — zero lines)
- Exit: 1
- Measured at: `49a42c2fc`
- Positive control: `grep -rn "CODEX_HOME" internal/cli/mcp_codex.go` → `internal/cli/mcp_codex.go:2091:// Stage 1 reads <CODEX_HOME>/auth.json ...`, exit 0 (same command form over a literal known to be present).
- Red-reason: REQ-010's injection and REQ-011's override guard have no existing site — the `codexChildEnv` posture (`codex_launcher.go:575`) carries no `RUST_LOG` branch.

### RED-3 — AC-005 + AC-006: cc/glm launchers have no `-d`/debug handling (observe-only is the today-behavior)

- Command: `grep -n '"-d"\|--debug\|debug' internal/cli/cc.go internal/cli/glm.go`
- Verbatim output: (none — zero lines across both files)
- Exit: 1
- Measured at: `49a42c2fc`
- Positive control: `grep -n "debug" internal/cli/logging.go` → hits at `:13`, `:18`, `:66`, exit 0 (same command form over a file known to contain the pattern).
- Red-reason: cc/glm consume only help/spawn/profile/entry tokens (`DisableFlagParsing` at `cc.go:133`, `glm.go:137`), so `-d` flows to the child untouched today and NO launcher-side trace exists — REQ-004's observe-only activation and the uniform emission of REQ-013 are absent. Note for run phase: the launcher-side change must leave the child-visible `-d` behavior byte-identical (that is the AC-005/006 assertion).

### RED-4 — AC-012: the only timing print path is threshold-gated; the debug test does not exist

- Command: `go test -count=1 -run '^TestCodexDebugSupersedesLaunchThreshold$' ./internal/cli/`
- Verbatim output:
```
ok  	github.com/modu-ai/moai-adk/internal/cli	1.040s [no tests to run]
```
- Exit: 0
- Measured at: `49a42c2fc`
- Structural evidence: `internal/cli/factory_launch_timing.go:92` — `if t.phaseElapsed() < threshold { return }` is the sole gate on `reportSlow` (`:87`); no unconditional dump path exists.
- Red-reason: the intended test is absent — a green run of an absent selector is vacuous, and the collector's only print path is threshold-gated, so REQ-014's unconditional debug print does not exist. Single-invocation and re-executable: the identical command runs the real test after implementation.

### RED-5 — AC-007: the uniformity matrix test does not exist

- Command: `go test -count=1 -run '^TestThreeRunnerDebugUniformityMatrix$' ./internal/cli/`
- Verbatim output:
```
ok  	github.com/modu-ai/moai-adk/internal/cli	1.347s [no tests to run]
```
- Exit: 0
- Measured at: `49a42c2fc`
- Red-reason: the intended test is absent; a green run of an absent selector is vacuous. Single-invocation and re-executable (release-blocking retained).

### RED-6 — AC-008: the step-trace test does not exist

- Command: `go test -count=1 -run '^TestCodexDebugTraceStepLines$' ./internal/cli/`
- Verbatim output:
```
ok  	github.com/modu-ai/moai-adk/internal/cli	1.098s [no tests to run]
```
- Exit: 0
- Measured at: `49a42c2fc`
- Red-reason: the intended test is absent; the named step vocabulary and prefix constant do not exist. Single-invocation and re-executable (release-blocking retained).

### RED-7 — AC-009: the keys-only secrecy test does not exist

- Command: `go test -count=1 -run '^TestCodexDebugTraceEnvKeysOnly$' ./internal/cli/`
- Verbatim output:
```
ok  	github.com/modu-ai/moai-adk/internal/cli	0.820s [no tests to run]
```
- Exit: 0
- Measured at: `49a42c2fc`
- Red-reason: the intended test is absent; REQ-006's keys-only rule has no assertion anywhere. Single-invocation and re-executable (release-blocking retained).

### RED-8 — AC-010: the worktree trace test does not exist

- Command: `go test -count=1 -run '^TestCodexDebugTraceWorktreeEntry$' ./internal/cli/`
- Verbatim output:
```
ok  	github.com/modu-ai/moai-adk/internal/cli	0.780s [no tests to run]
```
- Exit: 0
- Measured at: `49a42c2fc`
- Red-reason: the intended test is absent; REQ-007's worktree trace has no site. Single-invocation and re-executable (release-blocking retained).

### RED-9 — AC-015: the pre-seam ordering test does not exist

- Command: `go test -count=1 -run '^TestCodexDebugTracePrecedesExecSeam$' ./internal/cli/`
- Verbatim output:
```
ok  	github.com/modu-ai/moai-adk/internal/cli	0.972s [no tests to run]
```
- Exit: 0
- Measured at: `49a42c2fc`
- Structural evidence: the pre-seam discipline exists today only for the threshold report (t1378 REQ-012; MX:NOTE at `factory_launch_timing.go:85`); no ordering assertion covers a debug trace, and no test pins trace-vs-seam order on either door.
- Red-reason: the intended test is absent; REQ-009's both-doors ordering is unasserted. Single-invocation and re-executable (release-blocking retained).

## §D.1 — Edge cases

- `-d` as a `-w` VALUE (`moai codex -w -d`): the worktree value parser treats `-`-prefixed tokens as flags — the debug scan must not consume the value position of another flag (token-shape parity with `stripSpawnFlag`).
- Repeated debug tokens (`-d -d`): idempotent — one activation, no duplicate refusal.
- `--debug` on the `app` verb: launch verb — traced like `cli`.
- Debug token combined with `-f lane`: lane-claim step traces the claimed label and run id (KEY names only for env).
- Windows door: the trace prints ahead of the non-exec Windows launch path identically (build-tagged; asserted by AC-015's Windows clause).
- Empty head with only `--` and `-d` after it: passthrough, no activation (AC-003 shape).

## §D.2 — Quality gates

- TRUST 5: Tested (AC-001..AC-015 + 85% package coverage), Readable (named constants, golangci clean), Unified (gofmt), Secured (REQ-006 keys-only assertion is a gate, not a hope — AC-009 asserts sentinel-value absence), Trackable (Conventional Commits, card t1380 in every commit message).
- plan-auditor: Tier M threshold 0.80.

## §D.3 — Definition of Done

1. All 15 ACs green with verbatim outputs recorded in `progress.md` §E.2.
2. t1378 REQ-012's existing tests pass unmodified (AC-013 part 2).
3. `GOOS=windows go build ./...` exit 0.
4. `golangci-lint run ./internal/cli/...` clean.
5. `RUST_LOG` constant present in `internal/config/envkeys.go`; no inline env-name literals added outside it.
6. Sync-phase docs updated (launcher `--help` text is code-adjacent and covered in M1; CHANGELOG owned by manager-docs).
