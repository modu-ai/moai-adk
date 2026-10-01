---
id: SPEC-SYNC-GATE-SKIP-SUBSHELL-001
title: "Sync gate skipped-tool pipe-subshell loss repair (card t1395)"
version: "1.0.0"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh"
lifecycle: spec-anchored
tier: M
tags: "hooks, sync-gate, subshell, skipped-tools, template-first"
---

# SPEC-SYNC-GATE-SKIP-SUBSHELL-001

## HISTORY

- 2026-10-02: Authored at plan phase for card t1395 (factory lane, worktree `WT-gate-skipped-tools`, base `develop` @ `89e3164af`). Card reports the 16th reproduction of the sync-gate defect family: four language checks (csharp, elixir, flutter, swift) run `run_step` inside a pipeline subshell, so the `SKIPPED_TOOLS` update performed at `run_step`'s absent-tool branch is lost. Premises verified against this tree; observational RED executed and confirmed (evidence: `.moai/reports/t1395/`).

## A. Problem Statement

When a user project carries one of the four language markers (csharp / elixir / flutter / swift) and the corresponding toolchain (`dotnet` / `mix` / `dart` / `swift`) is absent from PATH, the sync-phase quality gate records the skip in a pipeline subshell and the record is lost. Observed chain (all four checks share the shape):

1. `run_step <tool> c1 <cmd> 2>&1 | head -N || true` — the left side of `|` runs in a subshell.
2. Inside `run_step`, the absent-tool branch executes `SKIPPED_TOOLS="$SKIPPED_TOOLS $tool"` (line 693 of the gate script) — this mutates the subshell's copy.
3. The main shell's `SKIPPED_TOOLS` stays empty → the advisory notice at lines 999-1001 is never emitted (stdout empty), and the audit line at line 1019 reads `skipped_tools=` (empty).
4. The allow path at line 1011 writes a `pass` record to `.moai/state/sync-quality-gate.last`; on the next Stop turn the record short-circuits at lines 545-547 with a silent `exit 0` — the pass is cached despite the missing checker.

Net effect: the user is never told a checker did not run, and the gate reads as verified. This is a user-project quality-gate reliability defect; this repository itself carries none of the four markers, so the defect is invisible to this repo's own dogfooding.

## B. Verified Premises (evidence, this tree @ `89e3164af`)

- **P1 — four checks, two byte-identical surfaces.** `cmp .claude/hooks/moai/sync-phase-quality-gate.sh internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh` → identical. The four pipeline shapes (working-tree line numbers, post-t1392): csharp `dotnet` (807), elixir `mix` (823), flutter `dart` (882), swift `swift` (886). A precise static discriminator — `grep -cE "run_step [a-zA-Z]+ c[12] [^']*\|"` — returns exactly 4 on both the working copy and `git show 89e3164af:<template path>`; the other 7 `run_step` lines carrying pipes have them inside the `sh -c '...'` argument string and are safe (no subshell wraps `run_step` itself).
- **P2 — `run_step` owns its own log redirection.** Line 681: `"$@" >> "$GATE_TMPDIR/$prefix.log" 2>&1`. All checker output reaches the per-check log without any external pipe; removing `| head -N` loses nothing observable (`run_step` emits nothing to stdout in either branch, and the slot exit rides the `$prefix.exit` file, which survives subshells).
- **P3 — subshell loss is mechanically the code shape.** Write: line 693 inside the function. Reads: notice gate 999-1000, audit line 1019. Cache write: line 1011 (`pass` record); cache silent reuse: lines 545-547. A pipeline's components run in subshells (POSIX shell semantics); a variable assignment in a subshell cannot reach the parent.
- **P4 — the existing C# fixture does NOT catch this defect.** `internal/template/hook_gate_exit_status_test.go` `gateExitCases()` (line 54 for csharp) places tool STUBS on PATH (lines 80-84), so `command -v "$tool"` succeeds and the absent-tool branch — the only writer of `SKIPPED_TOOLS` — never executes. The fixture covers the t602 exit-status family only. The observational RED therefore required a new tool-absent harness, committed as a plan-phase artifact at `red-now-t1395.sh` (this SPEC directory; byte-identical copy of the first run's `.moai/reports/t1395/red-now.sh`, which remains historical lane evidence — `.moai/reports/*` is gitignored per `.gitignore:235`, so the committed copy is the citation target). Executed 2026-10-02 against the pre-implementation tree, FROM the committed path (`bash .moai/specs/SPEC-SYNC-GATE-SKIP-SUBSHELL-001/red-now-t1395.sh "$PWD/.claude/hooks/moai/sync-phase-quality-gate.sh" "$PWD/.moai/reports/t1395/reexec"`): harness exit 0; `git rev-parse --short HEAD` = `89e3164af` observed immediately before and after; CELL1 stdout EMPTY, CELL2 `skipped_tools=` empty, CELL3 `pass` record written; positive controls PC1 (dotnet unresolvable in scrubbed env) / PC2 (`language=csharp` in audit) / PC3 (`skipped_tools=` token present) all PASS. Gate script sha256 prefix `19180598f67db114` (recomputed after the re-execution — unchanged; the harness path does not influence it). First-run and re-execution observed the identical worktree content id (`cf131089db9b7300b2acb5cfc18618bc5e08e1204abc26be7da7e38878afc8ff`); the temp fixture's own git commit SHA differs per run by design (fresh `mktemp` fixture each run). Full verbatim output quoted in `acceptance.md` §D.1.
- **P5 — t1392 already merged into this file.** Commits `f3530d80e` (line-based GO_ROOTS), `195d20ca9` (deleted-file module resolution), `b0a70f03b` (worktree/heavy-dir exclusion), `79aa807f0` (untracked-others exclusion walk), merged via `151dc980f`. All premise line numbers above are read from the post-t1392 tree. Line numbers are moving coordinates: run-phase re-anchors by the quoted content patterns, not by number.
- **P6 — marker gating keeps these paths out of this repo's own runs.** `detect_languages` (line 90) derives `GATE_LANG_CANDIDATES` (line 237) from project markers (`f.csproj`/`*.csproj`, `mix.exs`, `pubspec.yaml`, `Package.swift`); `CHANGED_LANGS` (lines 314-326) intersects with the changed-file delta; the case loop at line 753 runs per changed language. This repo carries none of the four markers — every AC is fixture-driven, never dependent on an installed toolchain.

## C. Requirements (GEARS)

- **REQ-001** — **When** a `run_step` invocation executes inside the sync-phase quality gate, the gate shall run it outside any pipeline subshell, so that the `SKIPPED_TOOLS` update performed by the absent-tool branch lands in the gate's main shell.
- **REQ-002** — **When** the gate's decision is `allow` and at least one checker was skipped because its tool was absent from PATH, the gate shall emit the skip advisory through the notice channel (`emit_gate_notice` appended to `GATE_OUTPUT_FILE`) so the skip reaches the user on stdout.
- **REQ-003** — **When** the gate writes its audit log line, the line shall carry a non-empty `skipped_tools=` value whenever a skip occurred in that run.
- **REQ-004** — The sync-phase quality gate script shall apply the repair to all four affected language checks (csharp `dotnet`, elixir `mix`, flutter `dart`, swift `swift`) — no check may retain the pipeline-wrapping-`run_step` shape.
- **REQ-005** — The gate script's two surfaces shall stay in contract: the template original (`internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh`) is the edit surface of record, the deployed local copy (`.claude/hooks/moai/sync-phase-quality-gate.sh`) shall be byte-identical after the repair, and `make build` shall regenerate the embedded binary copy.
- **REQ-006** — **When** the repair lands, the gate shall preserve every neighboring contract that shares the file, each verified by its named acceptance criterion: a failing checker still blocks (t602 exit-status shape → AC-005), the pass record is still written on allow-with-skips and checker-tool presence still rides the worktree content key (t1385 gate-cache contract → AC-006), the t1392 multi-language routing and aggregation behaviors stay unregressed (→ AC-007), and the cpp case-table family stays unregressed (→ AC-008).

## D. Non-Functional Constraints

- Shell-script-only hook pattern per `.moai/docs/hook-development.md` (no jq dependency in this gate — git/grep/awk only; 60s Stop registration timeout unchanged; `set -e` convention preserved).
- The new behavioral test drives the gate with a scrubbed, explicitly constructed environment (PATH built from known dirs with the target tool excluded) — lane env leakage falsifies env-reading gate tests; every env axis the test depends on is pinned in the test body (pattern: `gateExitRun`, `hook_gate_exit_status_test.go:110-117`).
- No git operations in the repair itself beyond what the gate script already performs; the lane commits after audit per factory protocol.
- Simplicity: the minimal repair is the deletion of the external pipeline on four lines. Any reshaping beyond that (e.g. journal-derived skip reconstruction) is out of scope unless the FOUNDER decision row (see `decision-index.md`) is settled otherwise.

## E. Out of Scope

### Out of Scope — t1396 sync-gate hook regression (exit 2 + empty stdout losing block JSON)
- Card t1396 records a separate sync-gate defect family (exit-2 stdout discard). This SPEC does not touch exit-code semantics; the gate keeps `exit 0` with stdout JSON as the blocking channel (lines 1022-1026).

### Out of Scope — t1385 gate-cache contract changes
- The pass-record-on-skip behavior, the stale window, the retry marker, and checker-tool-presence-in-key (lines 449-453, 519-636) are t1385 territory and stay byte-for-byte behaviorally identical.

### Out of Scope — internally-piped checker scripts (javac / kotlinc / scalac / ruby / php / g++ / R)
- Seven `run_step` lines carry pipes INSIDE their `sh -c '...'` argument strings. Those pipes do not wrap `run_step`, lose no variable state, and were already adjudicated by the t602 exit-status fix. Not touched.

### Out of Scope — Codex-side Go receipt producer (`internal/cli/codex_sync_gate.go`)
- The Codex Stop chain runs its checks in Go (`produceSyncGateReceipt`), not through the shell script; no subshell exists there. Separate surface, separate defect family if any.

### Out of Scope — notice wording and advisory-channel design
- The skip notice (line 1000) exists and is correct; this SPEC restores its reachability, not its text.

## F. Acceptance Criteria

Acceptance criteria are enumerated in `acceptance.md` (§D AC Matrix). Summary: one observational two-cell RED/GREEN AC, one static four-check sweep AC with positive control, one behavioral four-language sweep test AC, one dual-surface byte-identity + `make build` AC, and regression ACs for the t1392 behaviors and the t1385/t602 contracts that share the file.

## G. Risks

- The repair must not disturb the seven internally-piped checkers — the static AC discriminator (`run_step [a-zA-Z]+ c[12] [^']*\|`) separates the two classes and is proven on the pre-fix commit (4 hits) and post-fix target (0 hits).
- Line numbers in this SPEC are observations at `89e3164af`; run-phase re-anchors by quoted content.
- The behavioral test must exclude the target tool from PATH without breaking the gate's own dependencies (git/find/head/date/awk/...); the RED harness's `/usr/bin:/bin` scrub is the proven recipe on macOS.

## H. Cross-References

- Card t1395 (dispatch intent, verbatim in lane record) — origin.
- Card t1396 — sibling sync-gate regression, explicitly out of scope.
- Card t1385 / t1388 — gate-cache and stored-advisory contracts sharing the file (regression-protected here).
- Card t602 / H07 hooks audit — the exit-status shape family this file's tests already guard.
- `.moai/reports/t1395/` — RED-now harness, stdout, audit log, pass record, evidence ledger.
