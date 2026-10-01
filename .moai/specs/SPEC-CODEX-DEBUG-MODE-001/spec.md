---
id: SPEC-CODEX-DEBUG-MODE-001
title: "Codex launcher debug mode: -d/--debug flag parity across the cc/glm/codex launchers, launcher-side pre-exec trace, and Codex CLI debug linkage"
version: "0.1.0"
status: in-progress
created: 2026-10-01
updated: 2026-10-01
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "codex,launcher,debug,flag,parity,cc,glm,factory,lane,stderr,rust-log,t1378"
tier: M
---

# SPEC-CODEX-DEBUG-MODE-001

## §A — History

- 2026-10-01: Card t1380 opened from the operator's launcher-parity observation. The cc and glm launchers reach debug today by passing `-d` through to the sub-CLI (claude CLI documents `-d, --debug [filter]` — verified locally), while the codex launcher's only flag is `--spawn` (`internal/cli/codex_launcher.go:652`; command defined at `:600-601`). The codex CLI itself rejects `-d` (verified: codex-cli 0.159.3, `error: unexpected argument '-d' found`), so the cc/glm passthrough form cannot carry over — a launcher-owned debug surface is required.
- 2026-10-01: Operator rule stated on the card: "-f lane is the shared lane mode across all three runners, so debug must be uniform across the three too" (3러너 동형). The uniformity contract is fixed in §B.4/§E: uniform flag spellings, uniform destination and line-prefix vocabulary, per-backend child linkage.
- 2026-10-01: Composition constraint stated on the card: the per-step pre-exec launch timing (REQ-012 of SPEC-CODEX-LANE-SLOTS-001, card t1378, `internal/cli/factory_launch_timing.go`) already exists and must not be re-implemented or broken; the t1378 slot/launch-delay defect class is the first validation case for this mode.
- 2026-10-01: Plan-phase premise verification executed against this tree (HEAD `16424e9b4`, branch `WT-codex-debug-mode`); all axis premises verified, the codex CLI support question closed with an honest verification record (see `progress.md` § Premise Verification V1–V12).

## §B — Requirements

### §B.1 Debug flag surface (axis 3 — flag shape)

- REQ-001 (Event-driven): **When** a `-d` or `--debug` token appears in the pre-`--` argument head of a codex launch invocation, the codex launcher shall strip it from the launch arguments so it never reaches the codex child command line.
- REQ-002 (Event-driven): **When** a `-d` or `--debug` token appears after the `--` separator of a codex invocation, the codex launcher shall treat it as a verbatim codex argument — no stripping and no debug-trace activation.
- REQ-003 (Event-driven): **When** a debug token accompanies a readout verb (`moai codex status`), the codex launcher shall print a named diagnostic constant — mirroring the `--spawn` readout-refusal discipline (`internal/cli/codex_launcher.go:70-72`) — and exit non-zero.
- REQ-004 (Event-driven): **When** a `-d` or `--debug` token appears in the pre-`--` head of a cc or glm launch invocation, the cc/glm launcher shall activate its own debug trace and shall leave the token unmodified in the child arguments (observe-only: the child's native `-d` handling is preserved — verified on the claude CLI).

### §B.2 Launcher-side pre-exec trace (axis 1)

- REQ-005 (Event-driven): **When** debug mode is on, the launcher shall print one debug line per pre-exec step it executes — binary resolution, project-root resolution, init-offer gate, local-instruction load, lane claim, worktree materialization, child-env assembly, exec handoff — to stderr, every line carrying a single named prefix constant.
- REQ-006 (Ubiquitous): The debug trace shall name environment-variable keys and their presence or absence only; it shall never emit an environment-variable value.
- REQ-007 (Event-driven): **When** a `-w` worktree entry is resolved under debug mode, the launcher shall print the resolved directory, the writer-check outcome, and the anchor-lock outcome.
- REQ-008 (Ubiquitous): The launcher debug trace shall be gated by the debug tokens alone; `MOAI_LOG_LEVEL` shall neither enable nor suppress it (the level governs the slog default handler per `internal/cli/logging.go`, a separate axis).
- REQ-009 (State-driven): **While** the direct door replaces the process (`syscall.Exec`, `internal/cli/codex_direct_posix.go:53`), the debug trace shall be printed in full before the exec seam — the same pre-seam discipline REQ-012 of SPEC-CODEX-LANE-SLOTS-001 fixed for the timing report.

### §B.3 Codex CLI debug linkage (axis 2)

- REQ-010 (Event-driven): **When** debug mode is on and the inherited environment carries no `RUST_LOG`, the codex launcher shall append `RUST_LOG=debug` to the child environment (last-wins append posture of `codexChildEnv`, `internal/cli/codex_launcher.go:575`).
- REQ-011 (Ubiquitous): The codex launcher shall never modify an operator-supplied `RUST_LOG` value.

### §B.4 Three-runner uniformity contract (axis 3 — 3러너 동형)

- REQ-012 (Ubiquitous): The cc, glm, and codex launchers shall accept the same debug token spellings (`-d`, `--debug`) under the same pre-`--` scoping discipline.
- REQ-013 (Ubiquitous): The three launchers shall emit their launcher-side debug trace to the same destination (stderr) under the same line-prefix vocabulary. The child linkage is per-backend by design: cc/glm preserve the native `-d` passthrough (REQ-004), codex injects `RUST_LOG=debug` (REQ-010) because its CLI rejects `-d`.

### §B.5 Composition with the t1378 pre-exec timing report

Citation convention: within this SPEC, `t1378 REQ-012` always names SPEC-CODEX-LANE-SLOTS-001's REQ-012 — this SPEC also has its own §B.4 REQ-012 (uniform spelling), and the bare token collides in prose.

- REQ-014 (Event-driven): **When** debug mode is on for a codex lane launch, the launcher shall print the per-step pre-exec timing lines unconditionally — the slow-launch threshold does not gate them — reusing the t1378 REQ-012 collector (`internal/cli/factory_launch_timing.go`), which records every step regardless of threshold; no second instrumentation site shall be created.
- REQ-015 (Ubiquitous): With debug mode off, the launcher shall preserve t1378 REQ-012's threshold-gated report behavior unchanged; the debug print path adds output only when the debug tokens are present.

## §C — Acceptance Criteria (summary)

| REQ | AC | Classification |
|-----|----|----------------|
| REQ-001 | AC-001 (release-blocking), AC-002 (regression-guard), AC-003 (regression-guard) | flag strip + scoping |
| REQ-002 | AC-003 | post-`--` verbatim |
| REQ-003 | AC-004 (release-blocking) | readout refusal |
| REQ-004 | AC-005 (release-blocking), AC-006 (release-blocking) | cc/glm observe-only |
| REQ-005 | AC-008 (release-blocking) | step trace |
| REQ-006 | AC-009 (release-blocking) | keys-only secrecy |
| REQ-007 | AC-010 (release-blocking) | worktree trace |
| REQ-008 | AC-014 (regression-guard) | MOAI_LOG_LEVEL non-gating |
| REQ-009 | AC-015 (release-blocking) | pre-seam ordering, both doors |
| REQ-010, REQ-011 | AC-011 (release-blocking) | RUST_LOG injection + override guard |
| REQ-012, REQ-013 | AC-007 (release-blocking) | 3-runner uniformity matrix |
| REQ-014 | AC-012 (release-blocking) | debug supersedes threshold |
| REQ-015 | AC-013 (regression-guard) | debug-off behavior frozen |

Full Given-When-Then scenarios: `acceptance.md`.

## §D — Constraints

- New environment-variable names (`RUST_LOG`) enter `internal/config/envkeys.go` as constants; no inline `os.Getenv("RUST_LOG")` literals elsewhere (CLAUDE.local.md §14).
- All diagnostics (refusal text, trace prefix) are named constants in the launcher discipline (`internal/cli/codex_launcher.go:60-72` pattern); error wrapping `fmt.Errorf("operation: %w", err)`.
- The trace never emits environment-variable values (REQ-006; lane keys carry leader addresses and identity — Secured).
- Tests: `t.TempDir()` for every temporary directory; no OTEL `t.Setenv`; tests that read lane environment variables pin all axes via `t.Setenv` (lane env falsifies env-reading guard tests — recurring lesson, t1350).
- Windows parity: the trace and any pre-seam print must sit ahead of BOTH exec doors (POSIX `syscall.Exec` / `internal/cli/codex_direct_windows.go`); verified with `GOOS=windows go build`.
- t1378 REQ-012's threshold report and its tests (`internal/cli/factory_launch_timing_test.go`) are frozen when debug mode is off (REQ-015).
- No time estimates in plans or reports; priority labels only.

## §E — Design Notes (verified facts from plan-phase diagnosis)

- **Launch path today** (all steps pre-exec, before the seam): binary resolution (`codex_launcher.go:1002`), timing gate for lane launches only (`:1012-1014`), project-root resolution (`:1023-1027`), init-offer gate (`:1032`), local-instruction load (`:1033`), worktree resolve + writer check (`:1044-1055`, `resolveCodexWorktreeDir` at `:428`), child-args assembly (`:1056`), factory entry — join gate then lane claim (`internal/cli/codex_factory.go:135-151`, `kanban.ClaimFactoryLaneWithin`), env assembly (`codexChildEnv` `:575` + `codexFactoryEnv` `internal/cli/codex_factory.go:206`), anchor lock + exec handoff (`:1084-1106`), direct door (`internal/cli/codex_direct_posix.go:25-58`, `syscall.Exec` at `:53`).
- **The timing collector is the composition seam**: `factoryLaunchTiming` is nil-safe and records every `begin()` step regardless of threshold; the threshold only gates `reportSlow` (`internal/cli/factory_launch_timing.go:46-48, 87-101`). The cc/glm twins currently pass nil. Debug mode = instantiate the collector on every traced launch (any backend, lane or not) and add an unconditional pre-seam dump; t1378 REQ-012's threshold report is untouched.
- **Destination decision — stderr, not a `.moai/logs/` file.** Three precedents converge: (1) the launcher's own diagnostics already go to `cmd.ErrOrStderr()` (install hint `:1004`, advisory `:1049`, timing report `:1093/:1105`); (2) the repo-wide slog decision routes non-hook subcommands to stderr (`internal/cli/logging.go:78-83`) and reserves `.moai/logs/` writers for background processes that own no terminal (hook sink, `config.log` in `internal/config/log.go`, codex-adapter diagnostics `internal/codexadapter/diagnostics.go:16`); (3) the direct door replaces the process — a file sink would need open+flush before the seam on every traced launch, exactly the constraint REQ-012 solved by printing pre-seam.
- **Codex CLI verification record (honest)**: codex-cli 0.159.3 installed at `/Users/goos/.local/bin/codex`. Verified: `codex -d` → `error: unexpected argument '-d' found`; top-level `--help` lists no `--debug` option; `debug` exists only as a subcommand group (models / app-server / prompt-input); `codex exec --help` exposes no logging flag. NOT verified: `RUST_LOG` is undocumented in all local help surfaces and the upstream `docs/config.md` fetch returned non-codex content (inconclusive). Supporting-but-not-conclusive: the installed binary contains the `RUST_LOG` literal. Resolution: the `RUST_LOG=debug` injection is a best-effort linkage (an unknown env var is harmless to the child), NOT a load-bearing contract — the launcher-side trace (§B.2) is the primary debug surface. Residual recorded in plan.md §B.
- **claude CLI verified**: `claude --help` → `-d, --debug [filter]   Enable debug mode with optional category` plus `--debug-file <path>`. glm CLI: binary not installed locally; its native `-d` support is unverified and deliberately non-load-bearing — the cc/glm launcher change is observe-only, so child behavior is unchanged whatever the child supports.
- **Uniformity contract, precisely**: uniform = (a) flag spellings accepted pre-`--`, (b) trace destination and prefix vocabulary, (c) the semantic "debug mode = launcher pre-exec trace + maximal child debug linkage". Non-uniform by external necessity = the child linkage mechanism (cc/glm native passthrough vs codex env injection).
- **No existing debug surface**: `grep -rn RUST_LOG internal/ pkg/ cmd/` → zero hits; no `-d`/`--debug` handling in cc.go/glm.go/codex_launcher.go; `MOAI_LOG_LEVEL`/`MOAI_LOG_FORMAT` (`internal/config/envkeys.go:40,43`) gate only the slog handler.

## §F — Out of Scope

### Out of Scope — child-CLI internals

- The codex child's own debug output format or verbosity: external binary, `RUST_LOG` effect on it unverified (recorded residual, not repair scope).
- Any change to claude/glm child debug behavior (observe-only keeps it byte-identical).

### Out of Scope — per-axis debug granularity

- Separate per-axis flags (`-d=timing`, `-d=env`, …): debug mode is single-level in v1; the step vocabulary already separates axes per line.

### Out of Scope — persistent debug log files

- A `.moai/logs/` file sink for the launcher trace: stderr-only in v1 (§E destination decision); the hook sink and `config.log` writers are untouched.

### Out of Scope — factory/kanban subcommand debug surfaces

- Tracing inside `moai factory …` / `moai gtd …` subcommands: this SPEC covers the three launcher binaries' launch path only.

## §G — Cross-References

- SPEC-CODEX-LANE-SLOTS-001 — REQ-012 pre-exec timing report (composition target, card t1378); REQ-010 shared lane join ("one implementation for cc, glm, and the codex twin") is the uniformity precedent this SPEC extends.
- `internal/cli/codex_launcher.go:60-72` — the named-constant readout-refusal discipline REQ-003 mirrors.
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier M classification basis.
- Card t1380; audit evidence dir `.moai/reports/t1380/`.
