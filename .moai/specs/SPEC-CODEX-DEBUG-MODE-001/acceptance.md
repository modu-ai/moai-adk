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

All scenarios run against stubbed launch seams (`codexDirectLaunchFn`-style) — no real `codex`/`claude`/`glm` process is launched in any test. Verification commands use plain, file-anchored forms (worktree-guard safe) and `-run '^Name$'` anchored patterns.

### AC-001 — Codex head `-d` is stripped, never forwarded (release-blocking)

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

### AC-004 — Debug token on a readout verb refuses with a named diagnostic (release-blocking)

- **Given** the invocation `status -d` (readout verb, pre-`--` token)
- **When** the launcher routes the verb
- **Then** it prints the named refusal diagnostic constant (the `codexSpawnReadoutDiag` shape) and returns exit code 1, starting nothing.
- Verify: `go test -count=1 -run '^TestCodexDebugTokenRefusedOnReadoutVerb$' ./internal/cli/...`

### AC-005 — cc launch is observe-only: child keeps `-d`, launcher traces (release-blocking)

- **Given** a cc launch invocation `["-d", "--", "extra"]` with the launch seam stubbed
- **When** the cc launcher runs the launch path
- **Then** the child argv still contains `-d` (unmodified), and stderr carries the launcher debug trace lines.
- Verify: `go test -count=1 -run '^TestCCDebugTokenObserveOnly$' ./internal/cli/...`

### AC-006 — glm launch is observe-only under the same contract (release-blocking)

- **Given** a glm launch invocation `["-d"]` with the launch seam stubbed
- **When** the glm launcher runs the launch path
- **Then** the child argv still contains `-d` and stderr carries the launcher debug trace lines.
- Verify: `go test -count=1 -run '^TestGLMDebugTokenObserveOnly$' ./internal/cli/...`

### AC-007 — Three-runner uniformity matrix (release-blocking)

- **Given** the cc, glm, and codex launchers, each invoked once with `-d` and once with `--debug` (pre-`--`), seams stubbed
- **When** the matrix runs over all six combinations
- **Then** all three accept both spellings pre-`--`, all three emit trace lines to stderr under the same prefix constant, and the per-backend child linkage holds (cc/glm forward the token; codex strips it and injects `RUST_LOG=debug`).
- Verify: `go test -count=1 -run '^TestThreeRunnerDebugUniformityMatrix$' ./internal/cli/...`

### AC-008 — One trace line per executed pre-exec step (release-blocking)

- **Given** a codex launch under debug mode with the seam stubbed
- **When** the launch completes its pre-exec path
- **Then** stderr carries exactly one line per executed step from the named vocabulary (binary resolution, project-root resolution, init-offer gate, local-instruction load, lane claim when a factory lane, worktree materialization when `-w`, child-env assembly, exec handoff), each line under the single prefix constant.
- Verify: `go test -count=1 -run '^TestCodexDebugTraceStepLines$' ./internal/cli/...`

### AC-009 — Env keys named, env values never emitted (release-blocking)

- **Given** a traced launch with a sentinel value (e.g. `MOAI_KANBAN_LEAD_ADDR=leader-secret-sentinel`) present in the environment, all lane axes pinned via `t.Setenv`
- **When** the child-env assembly step is traced
- **Then** the trace output contains the key name `MOAI_KANBAN_LEAD_ADDR` with its presence, and the sentinel value string appears nowhere in the captured stderr.
- Verify: `go test -count=1 -run '^TestCodexDebugTraceEnvKeysOnly$' ./internal/cli/...`

### AC-010 — Worktree entry traced with resolved dir and lock outcome (release-blocking)

- **Given** an existing worktree fixture under `t.TempDir()` and a codex launch `["-w", "<fixture>", "-d"]`
- **When** the worktree materialization step runs
- **Then** the trace names the resolved directory, the writer-check outcome, and the anchor-lock outcome.
- Verify: `go test -count=1 -run '^TestCodexDebugTraceWorktreeEntry$' ./internal/cli/...`

### AC-011 — RUST_LOG injected when absent, preserved when present (release-blocking)

- **Given** (a) a traced launch with `RUST_LOG` unset, and (b) a traced launch with `RUST_LOG=info` set (pinned via `t.Setenv`)
- **When** the child env assembles in each scenario
- **Then** (a) the child env contains `RUST_LOG=debug`, and (b) the child env keeps `RUST_LOG=info` unmodified.
- Verify: `go test -count=1 -run '^TestCodexDebugRustLogInjection$' ./internal/cli/...`

### AC-012 — Debug lane launch prints step timings regardless of threshold (release-blocking)

- **Given** a codex lane launch (`-f lane`) under debug mode with all pre-exec steps stubbed to complete in well under the slow-launch threshold
- **When** the pre-exec path reaches the seam
- **Then** the per-step timing lines print unconditionally (the threshold does not gate them), and the lines come from the shared collector's recorded steps — a single recording site.
- Verify: `go test -count=1 -run '^TestCodexDebugSupersedesLaunchThreshold$' ./internal/cli/...`

### AC-013 — Debug-off behavior of REQ-012 is frozen (regression-guard)

- **Given** debug tokens absent, (a) a fast lane launch and (b) a lane launch exceeding the threshold (threshold pinned via `t.Setenv` on `MOAI_FACTORY_SLOW_LAUNCH_MS`)
- **When** each launch reaches the seam
- **Then** (a) prints no timing lines and no trace lines, and (b) prints exactly the REQ-012 threshold report — and `internal/cli/factory_launch_timing_test.go` passes with no modification.
- Verify: `go test -count=1 -run '^TestCodexDebugOffKeepsTimingReportFrozen$' ./internal/cli/... && go test -count=1 -run '^(TestFactoryLaunchTimingNilSafe|TestFactoryLaunchTimingMeasuresElapsed|TestCodexLaneLaunchTimingNamesSteps|TestCodexLaneLaunchTimingQuietUnderThreshold)$' ./internal/cli/...`

### AC-014 — MOAI_LOG_LEVEL does not gate the trace; package re-measurement green (regression-guard)

- **Given** a traced codex launch with `MOAI_LOG_LEVEL=error` pinned via `t.Setenv`
- **When** the launch runs
- **Then** the trace lines still print (level does not suppress them), and the whole-package re-measurement is green.
- Verify: `go test -count=1 -run '^TestCodexDebugTraceIgnoresLogLevel$' ./internal/cli/... && go test -count=1 -timeout 30m ./internal/cli/...`

## § Evidence Ledger (RED-now observations, executed on the pre-implementation tree `16424e9b4`)

- **RED-1** (AC-001/AC-002 premise): the codex launcher registers exactly one flag — `codexCmd.Flags().Bool("spawn", ...)` at `internal/cli/codex_launcher.go:652`; grep for `-d`/`--debug` handling over the launcher returns nothing. Observed this run.
- **RED-2** (AC-011 premise): `grep -rn RUST_LOG internal/ pkg/ cmd/` → zero hits (exit 1). No `RUST_LOG` handling exists. Observed this run.
- **RED-3** (AC-005/AC-006 premise): `grep '"-d"\|--debug\|debug' internal/cli/cc.go internal/cli/glm.go` → zero hits; both launchers are `DisableFlagParsing` (`cc.go:133`, `glm.go:137`) and consume only `--help/-h`, `--spawn`, `-p`, `-k/-f`, `-w` shapes, so `-d` flows to the child untouched today. Observed this run.
- **RED-4** (AC-012 premise): the timing collector exists and is threshold-gated — `factory_launch_timing.go:87-101` (`reportSlow` returns early when `phaseElapsed() < threshold`); no unconditional debug dump exists. Observed this run.

## §D.1 — Edge cases

- `-d` as a `-w` VALUE (`moai codex -w -d`): the worktree value parser treats `-`-prefixed tokens as flags — the debug scan must not consume the value position of another flag (token-shape parity with `stripSpawnFlag`).
- Repeated debug tokens (`-d -d`): idempotent — one activation, no duplicate refusal.
- `--debug` on the `app` verb: launch verb — traced like `cli`.
- Debug token combined with `-f lane`: lane-claim step traces the claimed label and run id (KEY names only for env).
- Windows door: the trace prints ahead of the non-exec Windows launch path identically (build-tagged).
- Empty head with only `--` and `-d` after it: passthrough, no activation (AC-003 shape).

## §D.2 — Quality gates

- TRUST 5: Tested (AC-001..AC-014 + 85% package coverage), Readable (named constants, golangci clean), Unified (gofmt), Secured (REQ-006 keys-only assertion is a gate, not a hope — AC-009 asserts sentinel-value absence), Trackable (Conventional Commits, card t1380 in every commit message).
- plan-auditor: Tier M threshold 0.80.

## §D.3 — Definition of Done

1. All 14 ACs green with verbatim outputs recorded in `progress.md` §E.2.
2. REQ-012's existing tests pass unmodified (AC-013 second clause).
3. `GOOS=windows go build ./...` exit 0.
4. `golangci-lint run ./internal/cli/...` clean.
5. `RUST_LOG` constant present in `internal/config/envkeys.go`; no inline env-name literals added outside it.
6. Sync-phase docs updated (launcher `--help` text is code-adjacent and covered in M1; CHANGELOG owned by manager-docs).
