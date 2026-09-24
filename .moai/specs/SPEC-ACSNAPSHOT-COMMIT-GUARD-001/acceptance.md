---
id: SPEC-ACSNAPSHOT-COMMIT-GUARD-001
title: "Acceptance — commit-time AC-snapshot guard"
version: "0.1.0"
created: 2026-09-24
author: manager-spec (card t1150)
---

# Acceptance — commit-time AC-snapshot guard

This file declares its criteria under the default counter prefix, so it carries no prefix declaration. It deliberately cites no other SPEC's criterion identifiers and no SPEC ID containing a hyphenated `AC` segment: the counter has no left token boundary and would count them.

Document-level measurement pin: tree `60017eb83` (binds every RED-now cell without its own pin). Every green cell names the run-phase command whose verbatim output and exit code must be recorded in progress.md §E.2 against the run tree.

Test selector used throughout: `go test ./internal/spec -run 'TestACBaselineCommitGuard' -count=1 -v`. A green reading requires the named subtest to appear as a `--- PASS` line; `[no tests to run]` or a missing subtest name is an empty sweep, not a pass.

## §1 Requirements → Criteria

| REQ | AC |
|---|---|
| REQ-ABG-001, 002, 003 | AC-ABG-001, AC-ABG-006 |
| REQ-ABG-004 | AC-ABG-004 |
| REQ-ABG-002 (unrecorded file) | AC-ABG-003 |
| REQ-ABG-002 (unchanged count) | AC-ABG-002 |
| REQ-ABG-005 | AC-ABG-010 |
| REQ-ABG-006 | AC-ABG-005 |
| REQ-ABG-007 | AC-ABG-007, AC-ABG-013 |
| REQ-ABG-008 | AC-ABG-007 |
| REQ-ABG-009 | AC-ABG-002 |
| REQ-ABG-010 | AC-ABG-011 |
| REQ-ABG-011, 012 | AC-ABG-009 |
| REQ-ABG-013 | AC-ABG-009 |
| REQ-ABG-014 | AC-ABG-011 |
| REQ-ABG-015 | AC-ABG-008a, AC-ABG-008b |
| REQ-ABG-016 | AC-ABG-014 |
| REQ-ABG-017 | AC-ABG-012 |

## §2 Criteria

### AC-ABG-001 — True positive: count-moving amendment without baseline is rejected (release-blocking)

**Given** a throwaway repo whose `HEAD` records `.moai/specs/SPEC-X-001/acceptance.md` as `COUNT 2 live=2 excluded=0 ambiguous=0` in the baseline, **When** a third criterion is added to that file and only the file is staged, and `scripts/ac-baseline/check-staged.sh` runs, **Then** it exits non-zero and stderr contains the path, `2`, `3`, the literal `MOAI_AC_BASELINE_REGENERATE=1 go test ./internal/spec -run TestACCounterBaselineRegenerate -count=1`, and `.moai/docs/ac-count-baseline-refresh.md`. The test asserts the regeneration text against the package constant `acRegenerateCommand`, not a restated literal.
- RED-now: ledger E1. Green path: M1; subtest `TestACBaselineCommitGuard/reject_count_move` `--- PASS`.

### AC-ABG-002 — Pass: in-place edit leaving the count unchanged (release-blocking)

**Given** the same repo, **When** prose in the recorded file is edited without changing its criteria and the file is staged, **Then** the checker exits 0 and stderr has exactly one line beginning `ac-baseline-guard: checked 1`.
- RED-now: ledger E1. Green path: M1; subtest `pass_count_unchanged`.

### AC-ABG-003 — Pass: new acceptance.md absent from the baseline (release-blocking)

**Given** the same repo, **When** a new `.moai/specs/SPEC-Y-001/acceptance.md` (index status `A`) is staged, **Then** the checker exits 0 and prints nothing (status `A` is not qualifying). A second case: an `M` file with no baseline record exits 0 and its `checked` line reports that file's staged count.
- RED-now: ledger E1. Green path: M1; subtests `pass_new_file`, `pass_unrecorded_modified`.

### AC-ABG-004 — Pass: amendment plus regenerated baseline in the same commit (release-blocking)

**Given** the AC-ABG-001 amendment, **When** the baseline line for that file is changed to `COUNT 3 live=3 excluded=0 ambiguous=0` and both files are staged, **Then** the checker exits 0. A mutant variant — baseline edited in the working tree but NOT staged — exits non-zero (proves the baseline is read from the index).
- RED-now: ledger E1. Green path: M1; subtests `pass_with_staged_baseline`, `reject_unstaged_baseline`.

### AC-ABG-005 — No-op: no qualifying file is silent and does no work

**Given** a repo whose staged `manager-docs.md` has its sentinel pair deleted, **When** only an unrelated file (and separately an `_archive/<dir>/acceptance.md` with status `M`, and a depth-2 `.moai/specs/a/b/acceptance.md`) is staged, **Then** the checker exits 0 with empty stdout and empty stderr — a broken counter source that is never reached produces no `NOT CHECKED` line, which proves extraction was skipped.
- RED-now: ledger E1. Green path: M1; subtests `noop_unrelated`, `noop_archive`, `noop_depth2`.

### AC-ABG-006 — Comparison parity with the corpus test

**Given** fixture pairs for each transition — COUNT→COUNT different `live`; COUNT→COUNT different `excluded` only; recorded COUNT→staged HALT; recorded HALT→staged COUNT; recorded HALT→staged HALT with a moved identifier set; recorded HALT→same HALT set — **When** the checker runs on each, **Then** its pass/reject outcome equals the `problem != ""` outcome of `acComparison` for the same inputs, asserted in the same test by calling `acComparison` directly.
- RED-now: ledger E1. Green path: M1; subtest `parity/<row>` for all six rows.

### AC-ABG-007 — Tool fault fails open, loudly

**Given** a qualifying staged amendment, **When** each fault is injected in turn — sentinel pair absent, sentinel pair duplicated, empty counter body, baseline absent from the index, that file's baseline line malformed — **Then** the checker exits 0 and stderr has a line beginning `ac-baseline-guard: NOT CHECKED` naming the fault and the file.
- RED-now: ledger E1. Green path: M1; subtests `fault/<kind>` for all five kinds.

### AC-ABG-008a — Real `git commit` enforcement with a hookdir hook present (release-blocking)

**Given** a throwaway repo on git ≥ 2.54 with `scripts/ac-baseline/install-hook.sh` applied and an executable `.git/hooks/pre-commit` that writes a marker file, **When** `git commit -m x` runs on the AC-ABG-001 scenario, **Then** it exits non-zero and `git rev-parse HEAD` is unchanged; **When** it runs on the AC-ABG-004 scenario, **Then** a new commit exists. The test logs whether the marker file was written in each case (hookdir coexistence order recorded, not assumed).
- RED-now: ledger E2. Green path: M2; subtest `commit/hookdir_present` `--- PASS` (a `--- SKIP` does not satisfy this criterion).

### AC-ABG-008b — Real `git commit` enforcement under `core.hooksPath=/dev/null` (release-blocking)

**Given** the AC-ABG-008a repo with `core.hooksPath` set to `/dev/null` (the live repository's measured setting), **When** `git commit -m x` runs on the AC-ABG-001 scenario, **Then** it exits non-zero and `HEAD` is unchanged. If instead the commit succeeds, run halts with a blocker report (plan.md M2 stop condition).
- RED-now: ledger E2. Green path: M2; subtest `commit/hookspath_devnull`.

### AC-ABG-009 — Installer idempotency, old-git refusal, missing-script tolerance

**Given** a throwaway repo, **When** the installer runs twice, **Then** `git config --get-all hook.ac-baseline-guard.event` prints exactly `pre-commit` once and `git config --get-all hook.ac-baseline-guard.command` prints exactly one line. **When** it runs with a stub `git` first on `PATH` reporting `git version 2.53.0`, **Then** it exits non-zero, stderr names `2.53.0` and `2.54`, and `git config --get-regexp '^hook\.'` stays empty. **When** the installed command runs in a repo without `scripts/ac-baseline/check-staged.sh`, **Then** it exits 0 with a stderr line beginning `ac-baseline-guard: NOT CHECKED`.
- RED-now: ledger E2. Green path: M2; subtests `install/idempotent`, `install/old_git`, `install/missing_script`.

### AC-ABG-010 — Alternate index (`git commit -a`)

**Given** the AC-ABG-001 amendment left UNSTAGED in the working tree of an installed throwaway repo, **When** `git commit -a -m x` runs, **Then** it exits non-zero and `HEAD` is unchanged.
- RED-now: ledger E2. Green path: M2; subtest `commit/all_flag`.

### AC-ABG-011 — Nothing managed is written

**Given** an installed throwaway repo with a hookdir `pre-commit` and a `.moai-pre-commit.sha256` file, **When** the installer runs and a rejected plus a passing commit are attempted, **Then** the sha256 of `.git/hooks/pre-commit`, of `.git/hooks/.moai-pre-commit.sha256`, and the value of `core.hooksPath` are identical before and after, and `git status --porcelain` shows no change to the baseline file made by the checker.
- RED-now: ledger E2. Green path: M2; subtest `install/managed_untouched`.

### AC-ABG-012 — Documentation and live install evidence

**Then** `grep -c 'hook.ac-baseline-guard' .moai/docs/ac-count-baseline-refresh.md` prints a value ≥ 1 and the doc names `NOT CHECKED` and `--no-verify`. After the live install (plan.md §6), `git config --get-regexp '^hook\.ac-baseline-guard\.'` prints exactly two lines (`event pre-commit`, `command …`).
- RED-now: ledger E3 (doc), E4 (config). Green path: M3 (doc); live install per plan.md §6.

### AC-ABG-013 — No counter copy in the checker

**Then** `grep -c 'moai-ac-prefix' scripts/ac-baseline/check-staged.sh` prints `0` with exit 1 (the counter's unique declaration token is absent — the program is extracted, not embedded), and a mutant subtest that changes the staged counter's default prefix changes the checker's measured count (proves the staged carrier is what runs).
- RED-now: ledger E5. Green path: M1; subtest `counter_from_index`.

### AC-ABG-014 — Local-only placement, CI untouched

**Then** `git diff --name-only 60017eb83 -- internal/template/templates .github/workflows` on the run tree prints nothing, and `ls internal/template/templates/scripts/ac-baseline` fails with exit 1.
- RED-now: not applicable — a preservation guard (regression-guard class, not release-blocking). Green path: M4.

### AC-ABG-015 — Swept count and corpus gate intact

**Then** the selector in the header lists every subtest named above as `--- PASS` (none as `--- SKIP` on the run machine, git 2.54.0), and `go test ./internal/spec -run TestACCounterFullCorpusMatchesBaseline -count=1` exits 0 on the run tree.
- RED-now: ledger E6 (empty sweep today). Green path: M4.

## §3 RED-now evidence ledger (tree 60017eb83, 2026-09-24)

| Id | Command | Verbatim stdout/stderr | Exit | Why red |
|---|---|---|---|---|
| E1 | `ls scripts/ac-baseline` | `ls: scripts/ac-baseline: No such file or directory` | 1 | checker absent — every checker-level criterion cannot run |
| E2 | `git config --get-regexp '^hook\.'` | (empty) | 1 | no config-defined hook installed; installer absent (E1) |
| E3 | `grep -c 'hook\.' .moai/docs/ac-count-baseline-refresh.md` | `0` | 1 | doc carries no mechanical-gate section |
| E4 | `git config --get-regexp '^hook\.'` | (empty) | 1 | live repository not armed |
| E5 | `ls scripts/ac-baseline` | `ls: scripts/ac-baseline: No such file or directory` | 1 | nothing to grep yet; criterion becomes measurable at M1 |
| E6 | `go test ./internal/spec -run 'TestACBaselineCommitGuard' -count=1 -v` | `testing: warning: no tests to run` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/spec	0.324s [no tests to run]` | 0 | empty sweep reads as `PASS` — exactly why AC-ABG-015 requires named `--- PASS` lines |

## §4 Edge cases

- Partially staged file (`MM`): the staged blob is judged, not the working tree (AC-ABG-004 mutant variant covers the baseline side).
- Rename of an `acceptance.md` within `.moai/specs/`: with renames disabled it appears as `D` + `A`, both out of scope (spec.md §C).
- Merge commit without conflicts: not guarded (spec.md §C); a conflicted merge concluded with `git commit` is guarded like any commit.

## §5 Definition of Done

- All release-blocking criteria (001, 002, 003, 004, 008a, 008b) green with recorded command, verbatim output, exit code, and run-tree SHA in progress.md §E.2.
- AC-ABG-015 swept count shows no empty sweep and no skip.
- `go vet ./internal/spec/...` and `golangci-lint run ./internal/spec/...` clean.
- Live install decision (plan.md §6) resolved and its evidence recorded.
