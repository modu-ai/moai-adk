---
id: SPEC-CODEX-DEBUG-MODE-001
title: "Implementation plan — codex launcher debug mode"
version: "0.1.0"
created: 2026-10-01
updated: 2026-10-01
author: manager-spec
module: "internal/cli"
tier: M
---

# SPEC-CODEX-DEBUG-MODE-001 — Plan

## §A — Context

Card t1380: the codex launcher has no debug surface while cc/glm reach debug via `-d` passthrough to their sub-CLIs. The codex CLI rejects `-d` (verified 0.159.3), so parity cannot be passthrough — the launcher needs its own debug mode that (1) traces its pre-exec path, (2) links the codex child into debug via env injection, and (3) presents the same flag surface as cc/glm (operator rule: 3러너 동형). Card t1378's pre-exec timing collector (`internal/cli/factory_launch_timing.go`, REQ-012 of SPEC-CODEX-LANE-SLOTS-001) is the composition seam: nil-safe, records every step regardless of threshold, shared by the join gate.

Primary touch points: `internal/cli/codex_launcher.go` (flag parse, trace emission, env assembly), `internal/cli/codex_factory.go` (lane-claim trace), `internal/cli/codex_direct_{posix,windows}.go` (pre-seam ordering), `internal/cli/cc.go` / `internal/cli/glm.go` (observe-only scan + trace), `internal/config/envkeys.go` (`RUST_LOG` constant), `internal/cli/factory_launch_timing.go` (debug dump next to `reportSlow` — read-mostly), tests across the same packages.

## §B — Known Issues / Verified-vs-Residual

- **Verified**: codex flag surface is `--spawn` only (`codex_launcher.go:652`); codex CLI 0.159.3 rejects `-d` and documents no `--debug`/logging flag; claude CLI documents `-d, --debug [filter]`; no `RUST_LOG` handling exists anywhere in the repo (grep exit 1).
- **Residual 1 (accepted, non-blocking)**: codex's `RUST_LOG` support is undocumented in every locally available surface; the upstream config-docs fetch was inconclusive. The injection is best-effort (unknown env vars are harmless to the child); the launcher trace is the primary debug surface. If a future codex version documents a richer form, the injection site (`codexChildEnv`) is the single place to change.
- **Residual 2 (accepted, non-blocking)**: the glm binary is not installed locally, so its native `-d` support is unverified. Non-load-bearing: the cc/glm change is observe-only, so child behavior is identical whether or not the child understands `-d`.
- **No `[NEEDS CLARIFICATION]` markers**: both residuals above are recorded decisions that do not gate run-phase entry; they are escalated to the operator as decision-index rows Q3 (EVIDENCE-NEEDED) and via the frozen-behavior guarantees instead.

## §C — Pre-flight

1. Baseline: `go vet ./internal/cli/...` and `golangci-lint run ./internal/cli/...` clean on the untouched tree.
2. Baseline: `go test -count=1 -run '^(TestFactoryLaunchTimingNilSafe|TestFactoryLaunchTimingMeasuresElapsed|TestCodexLaneLaunchTimingNamesSteps|TestCodexLaneLaunchTimingQuietUnderThreshold)$' ./internal/cli/...` green (REQ-015 freeze reference — all four existing timing tests, exact-name anchored).
3. Confirm worktree HEAD equals local develop tip lineage before first commit (lane discipline; this plan performs no git operations — the lane commits after audit).

## §D — Constraints

- Env names as constants: add `EnvRustLog = "RUST_LOG"` to `internal/config/envkeys.go`; never inline (§14).
- All new diagnostics and the trace prefix are named constants (launcher pattern `codex_launcher.go:60-72`).
- Errors wrapped `fmt.Errorf("operation: %w", err)`; exit-code discipline via `exitCodeError` (launcher precedent).
- Tests: `t.TempDir()`; no OTEL `t.Setenv`; every test reading lane env vars pins all axes with `t.Setenv` (t1350 lesson); launch seams (`codexDirectLaunchFn`-style) stub the child — no real `codex`/`claude`/`glm` process in any test.
- REQ-006 is a Secured constraint: assert absence of sentinel env VALUES in trace output tests, not just presence of key names.
- Pre-seam ordering holds on both doors: POSIX `syscall.Exec` and the Windows stub; verify `GOOS=windows go build ./...`.
- No changes to `internal/kanban` (the claim engine is traced from the launcher side, not modified); no changes to t1378 REQ-012 (SPEC-CODEX-LANE-SLOTS-001) threshold semantics.

## §E — Self-Verification (run-phase exit evidence)

- E1: AC-001..AC-015 re-run green, outputs captured into `progress.md` §E.2.
- E2: `GOOS=windows go build ./...` exit 0.
- E3: `go test -cover ./internal/cli/...` at or above the package target (85%).
- E4: grep guard — no AskUserQuestion/interactive-prompt additions in the diff.
- E5: `golangci-lint run ./internal/cli/...` clean.
- E6: push state — lane-owned (lane pushes; this plan asserts nothing about remotes).
- E7: lessons check — lane-env pinning applied to every new env-reading test.

## §F — Milestones (decision-reversibility order — most likely to change first)

- M1 (Priority High) — Debug flag surface and parse discipline. Debug-token constants; codex head strip (mirror the `stripSpawnFlag` / `stripCodexWorktreeFlag` token-shape pattern); post-`--` scoping; readout-verb refusal with a named diagnostic; `codexCmd` Long/help text updated. ACs: AC-001, AC-002, AC-003, AC-004. (Parser shape and flag semantics are the decisions most likely to be revisited — they are the user-facing contract.)
- M2 (Priority High) — Launcher trace engine + child-env linkage. Step-recorder wiring: instantiate the shared collector on every traced codex launch (not lane-only); a debug dump function beside `reportSlow` (pre-seam, unconditional under debug); per-step lines for binary/root/init/instruction/claim/worktree/env/handoff; keys-only env reporting; worktree trace; `RUST_LOG=debug` injection with operator-override guard in `codexChildEnv` posture. ACs: AC-008, AC-009, AC-010, AC-011, AC-015. (New trace vocabulary and the env-injection rule are the second-most-likely-to-change surface; AC-015's ordering assertion lives with the trace emission it orders.)
- M3 (Priority Medium) — Three-runner uniformity wiring. cc/glm observe-only debug-token scan (post-`--`-aware, pre-subcommand-routing discipline identical to the `--help` scan); shared trace emission from the cc/glm pre-exec phases (entry parse, settings prep, lane join, worktree, launch handoff); the uniformity matrix test iterating all three launchers. ACs: AC-005, AC-006, AC-007.
- M4 (Priority Medium) — t1378 composition + regression pins. Debug-supersedes-threshold print path reusing the collector's recorded steps; this SPEC's REQ-015 freeze: existing `factory_launch_timing_test.go` passes unmodified (t1378 REQ-012 behavior); debug-off byte-absence of trace lines. ACs: AC-012, AC-013.
- M5 (Priority Low) — Mechanical sweep. `MOAI_LOG_LEVEL` non-gating pin; whole-package `internal/cli` re-measurement; `GOOS=windows` build; lint clean; MX annotations per protocol. ACs: AC-014.

## §G — Anti-Patterns

- Do NOT forward `-d`/`--debug` to the codex child (the CLI rejects them — verified).
- Do NOT consume `-d` on cc/glm by stripping it (that would silently change the child's behavior — REQ-004).
- Do NOT emit environment values in trace lines (REQ-006; lane keys carry identity and addresses).
- Do NOT add a second timing instrumentation site (REQ-014 names the collector as the only recorder).
- Do NOT gate the trace on `MOAI_LOG_LEVEL` (REQ-008 — two separate axes).
- Do NOT place trace prints after the exec seam (REQ-009 — they would never run on the direct door).

## §H — Cross-References

- `../spec.md` §E Design Notes — verified launch-path enumeration and the codex CLI verification record.
- `../acceptance.md` § Evidence Ledger — RED-now observations on the pre-implementation tree.
- `../decision-index.md` — Q1–Q6 decision routing (output destination, flag spelling, linkage evidence, granularity, readout refusal, tier).
- SPEC-CODEX-LANE-SLOTS-001 — REQ-012 collector; REQ-010 shared lane join precedent.
