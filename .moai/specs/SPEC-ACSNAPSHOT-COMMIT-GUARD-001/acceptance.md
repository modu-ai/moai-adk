---
id: SPEC-ACSNAPSHOT-COMMIT-GUARD-001
title: "Acceptance — commit-time AC-snapshot guard"
version: "0.2.0"
created: 2026-09-24
author: manager-spec (card t1150)
---

# Acceptance — commit-time AC-snapshot guard

This file declares its criteria under the default counter prefix, so it carries no prefix declaration. It deliberately cites no other SPEC's criterion identifiers, no fixture criterion identifiers, and no SPEC ID containing a hyphenated `AC` segment: the counter has no left token boundary and would count them. Fixture criteria below are described by role ("identifier X"), never spelled.

Document-level measurement pin: tree `60017eb83` (binds every RED-now cell without its own pin). Every green cell names the run-phase command whose verbatim output and exit code must be recorded in progress.md §E.2 against the run tree.

Test selector used throughout: `go test ./internal/spec -run 'TestACBaselineCommitGuard' -count=1 -v`. A green reading requires the named subtest to appear as a `--- PASS` line; `[no tests to run]`, a missing subtest name, or a `--- SKIP` on the run machine is an empty sweep, not a pass.

Common Given (unless stated): a throwaway repo in `t.TempDir()` whose `HEAD` holds the counter carrier copied from the tracked `.claude/agents/moai/manager-docs.md`, `.moai/specs/SPEC-X-001/acceptance.md` with two live criteria, and a baseline recording that file as `COUNT 2 live=2 excluded=0 ambiguous=0`.

## §1 Requirements → Criteria

| REQ | AC |
|---|---|
| REQ-ABG-001 | AC-ABG-001, AC-ABG-009, AC-ABG-010 |
| REQ-ABG-002 | AC-ABG-002, AC-ABG-003, AC-ABG-006 |
| REQ-ABG-003 | AC-ABG-001 |
| REQ-ABG-004 | AC-ABG-004 |
| REQ-ABG-005 | AC-ABG-012 |
| REQ-ABG-006 | AC-ABG-005 |
| REQ-ABG-007 | AC-ABG-007, AC-ABG-009, AC-ABG-010, AC-ABG-016 |
| REQ-ABG-008 | AC-ABG-007, AC-ABG-008 |
| REQ-ABG-009 | AC-ABG-002, AC-ABG-003 |
| REQ-ABG-010 | AC-ABG-014 |
| REQ-ABG-011, REQ-ABG-012 | AC-ABG-013 |
| REQ-ABG-013 | AC-ABG-013 |
| REQ-ABG-014 | AC-ABG-014 |
| REQ-ABG-015 | AC-ABG-011 |
| REQ-ABG-016 | AC-ABG-015 |

## §2 Criteria

### AC-ABG-001 — True positive: count-moving amendment without baseline is rejected (release-blocking)

**Given** the common repo, **When** a third criterion is added to the recorded file, only that file is staged, and `scripts/ac-baseline/check-staged.sh` runs, **Then** it exits non-zero and stderr contains the path, the structured tokens `live=2` (recorded) and `live=3` (staged), the regeneration command asserted against the package constant `acRegenerateCommand`, and `.moai/docs/ac-count-baseline-refresh.md`.
- RED-now: ledger E1. Green path: M1; subtest `reject_count_move`.

### AC-ABG-002 — Pass: in-place edit leaving the count unchanged (release-blocking)

**Given** the common repo, **When** prose in the recorded file is edited without changing its criteria and the file is staged, **Then** the checker exits 0 and stderr is exactly one line beginning `ac-baseline-guard: checked 1`.
- RED-now: ledger E1. Green path: M1; subtest `pass_count_unchanged`.

### AC-ABG-003 — Pass: new or unrecorded acceptance.md (release-blocking)

**Given** the common repo, **When** a new `.moai/specs/SPEC-Y-001/acceptance.md` (index status `A`) is staged, **Then** the checker exits 0 with empty stdout and stderr. **When** instead an `M` file with no baseline record is staged, once counting and once halting (identifier X marked on one occurrence only), **Then** both exit 0 and the `checked` line reports that file's staged `COUNT` or `HALT` value.
- RED-now: ledger E1. Green path: M1; subtests `pass_new_file`, `pass_unrecorded_counts`, `pass_unrecorded_halts`.

### AC-ABG-004 — Pass: amendment plus regenerated baseline in the same commit (release-blocking)

**Given** the AC-ABG-001 amendment, **When** the baseline line is changed to `COUNT 3 live=3 excluded=0 ambiguous=0` and both files are staged, **Then** the checker exits 0. **When** the baseline is edited in the working tree but NOT staged, **Then** it exits non-zero.
- RED-now: ledger E1. Green path: M1; subtests `pass_with_staged_baseline`, `reject_unstaged_baseline`.

### AC-ABG-005 — No-op: no qualifying file is silent and does no work

**Given** a repo whose staged counter carrier has its sentinel pair deleted, **When** only an unrelated file is staged, and separately an `_archive/<dir>/acceptance.md` with status `M`, and separately a depth-2 `.moai/specs/a/b/acceptance.md` with status `M`, **Then** each run exits 0 with empty stdout and empty stderr — the broken counter source is never reached, proving extraction was skipped.
- RED-now: ledger E1. Green path: M1; subtests `noop_unrelated`, `noop_archive`, `noop_depth2`.

### AC-ABG-006 — Comparison parity with the corpus test, every row

**Given** fixture pairs built from EVERY row of the case table of `TestACBaselineComparisonTransitions` (ten rows at `60017eb83`, iterated from the table itself rather than restated), plus one extra row where the staged file's ambiguous identifiers first appear in reverse sorted order and form the same set as a recorded HALT, **When** the checker runs on each, **Then** its pass/reject outcome equals `problem != ""` from `acComparison` for the same inputs (called directly in the test), and the extra row passes.
- RED-now: ledger E1. Green path: M1; subtests `parity/<row-name>` for every table row and `parity/halt_ids_unsorted`.

### AC-ABG-007 — Tool fault fails open, loudly

**Given** a qualifying staged amendment, **When** each fault is injected in turn — sentinel pair absent; duplicated; END before BEGIN; empty counter body; baseline absent from the index; that file's baseline line malformed; staged counter body ending in `exit 2` — **Then** each run exits 0 and stderr has a line beginning `ac-baseline-guard: NOT CHECKED` naming the fault and the file.
- RED-now: ledger E1. Green path: M1; subtests `fault/<kind>` for all seven kinds.

### AC-ABG-008 — Mismatch takes precedence over a fault (release-blocking)

**Given** two qualifying staged files in one commit — file A amended to move its count without a baseline change, file B whose baseline line is malformed, **When** the checker runs, **Then** it exits non-zero, stderr carries the rejection for A AND a `NOT CHECKED` line for B.
- RED-now: ledger E1. Green path: M1; subtest `mixed/mismatch_plus_fault`.

### AC-ABG-009 — Only the index is judged (release-blocking)

**Given** the common repo, **When** (a) the staged blob is count-unchanged while the working tree adds a criterion (`MM`), **Then** exit 0 with `checked 1`; **When** (b) the staged blob adds a criterion while the working tree is reverted to the `HEAD` content, **Then** exit non-zero; **When** (c) a counter carrier whose default prefix is changed is staged while the working tree AND `HEAD` keep the unmutated carrier, **Then** the checker's measured count follows the staged carrier.
- RED-now: ledger E1. Green path: M1; subtests `index_only/mm_unstaged_criterion`, `index_only/staged_criterion_reverted_tree`, `index_only/counter_from_index`.

### AC-ABG-010 — Hostile path text reaches no shell and no pattern (release-blocking)

**Given** a qualifying directory whose single segment contains the substring `$(touch pwned)`, a space, and the regex metacharacters `.` and `[`, with a matching baseline record and a second record whose path differs only in that metacharacter's position, **When** the file is amended to move its count and the checker runs, **Then** it rejects naming that exact path and the right record's value, and no file named `pwned` exists anywhere under the temp repo afterwards.
- RED-now: ledger E1. Green path: M1; subtest `hostile_path`.

### AC-ABG-011 — Real `git commit` enforcement, with and without the hookdir (release-blocking)

**Given** a throwaway repo on git ≥ 2.54 with `scripts/ac-baseline/install-hook.sh` applied and an executable `.git/hooks/pre-commit` that writes a marker file, **When** `git commit -m x` runs on the AC-ABG-001 scenario, **Then** exit non-zero and `git rev-parse HEAD` unchanged; **When** it runs on the AC-ABG-004 scenario, **Then** a new commit exists; the test logs whether the marker was written in each case. **When** the same rejecting scenario runs with `core.hooksPath=/dev/null`, and again from a linked worktree of the repo, **Then** each exits non-zero with `HEAD` unchanged.
- RED-now: ledger E2. Green path: M0 (premise) + M2; subtests `commit/premise_probe`, `commit/hookdir_present`, `commit/hookspath_devnull`, `commit/linked_worktree`.

### AC-ABG-012 — Alternate indexes (`-a`, pathspec-only)

**Given** an installed throwaway repo, **When** the AC-ABG-001 amendment is left UNSTAGED and `git commit -a -m x` runs, **Then** exit non-zero and `HEAD` unchanged. **When** the amendment and a matching regenerated baseline are both staged in the real index and `git commit -m x .moai/specs/SPEC-X-001/acceptance.md` runs (the temporary index omits the baseline), **Then** exit non-zero and `HEAD` unchanged.
- RED-now: ledger E2. Green path: M2; subtests `commit/all_flag`, `commit/pathspec_only`.

### AC-ABG-013 — Installer idempotency, old-git refusal, missing-script tolerance

**Given** a throwaway repo, **When** the installer runs twice, **Then** `git config --get-all hook.ac-baseline-guard.event` prints exactly `pre-commit` once and `git config --get-all hook.ac-baseline-guard.command` prints exactly one line. **When** it runs with a stub `git` first on `PATH` reporting `git version 2.53.0`, **Then** it exits non-zero, stderr names `2.53.0` and `2.54`, and `git config --get-regexp '^hook\.'` stays empty. **When** the installed command runs in a repo without `scripts/ac-baseline/check-staged.sh`, **Then** it exits 0 with a stderr line beginning `ac-baseline-guard: NOT CHECKED`.
- RED-now: ledger E2. Green path: M2; subtests `install/idempotent`, `install/old_git`, `install/missing_script`.

### AC-ABG-014 — Nothing managed or tracked is written

**Given** an installed throwaway repo with a hookdir `pre-commit` and a `.moai-pre-commit.sha256` file, **When** the installer runs and a rejected plus a passing commit are attempted, **Then** the sha256 of `.git/hooks/pre-commit` and of `.git/hooks/.moai-pre-commit.sha256` and the value of `core.hooksPath` are identical before and after, and the checker leaves the baseline file's index and working-tree content byte-unchanged.
- RED-now: ledger E2. Green path: M2; subtest `install/managed_untouched`.

### AC-ABG-015 — Documentation, and the lead's live-install evidence

**Then** `grep -c 'hook.ac-baseline-guard' .moai/docs/ac-count-baseline-refresh.md` prints a value ≥ 1, and the doc names `NOT CHECKED`, `--no-verify`, `SKIP_MOAI_PRECOMMIT`, the lead as install owner, and the completion-report quoting obligation. After the lead's install (spec.md §A.4-3 — not a lane step): `git config --get-regexp '^hook\.ac-baseline-guard\.'` prints exactly two lines, and one rejected plus one passing commit in a `develop`-absorbed tree are recorded.
- RED-now: ledger E3 (doc), E4 (config). Green path: M3 (doc); lead install after the develop merge (live evidence).

### AC-ABG-016 — Swept count, single counter source, placement, corpus gate intact

**Then** the selector in the header lists every subtest named above as `--- PASS` (none `--- SKIP` on the run machine, git 2.54.0, darwin); `grep -c 'moai-ac-prefix' scripts/ac-baseline/check-staged.sh` prints `0`; `git diff --name-only 60017eb83 -- internal/template/templates .github/workflows` on the run tree prints nothing; and `go test ./internal/spec -run 'TestACCounterFullCorpusMatchesBaseline|TestACBaselineComparisonTransitions' -count=1 -v` shows both as `--- PASS`.
- RED-now: ledger E5 (checker absent), E6 (empty sweep today). Green path: M4.

## §3 RED-now evidence ledger (tree 60017eb83, 2026-09-24)

| Id | Command | Verbatim stdout/stderr | Exit | Why red |
|---|---|---|---|---|
| E1 | `ls scripts/ac-baseline` | `ls: scripts/ac-baseline: No such file or directory` | 1 | checker absent — every checker-level criterion cannot run |
| E2 | `git config --get-regexp '^hook\.'` | (empty) | 1 | no config-defined hook installed; installer absent (E1) |
| E3 | `grep -c 'hook\.' .moai/docs/ac-count-baseline-refresh.md` | `0` | 1 | doc carries no mechanical-gate section |
| E4 | `git config --get-regexp '^hook\.'` | (empty) | 1 | live repository not armed |
| E5 | `ls scripts/ac-baseline` | `ls: scripts/ac-baseline: No such file or directory` | 1 | nothing to grep yet; measurable at M1 |
| E6 | `go test ./internal/spec -run 'TestACBaselineCommitGuard' -count=1 -v` | `testing: warning: no tests to run` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/spec	0.324s [no tests to run]` | 0 | empty sweep reads as `PASS` — exactly why AC-ABG-016 requires named `--- PASS` lines |

## §4 Edge cases

- Rename of an `acceptance.md` within `.moai/specs/`: with renames disabled it appears as `D` + `A`, both out of scope (spec.md §C).
- Merge commit without conflicts: not guarded (spec.md §C); a conflicted merge concluded with `git commit` is guarded like any commit.
- Trees without the checker (unabsorbed lanes, primary checkout on `main`): `NOT CHECKED` on every commit, never blocked (spec.md §A.4-3; AC-ABG-013 third case).

## §5 Definition of Done

- Release-blocking criteria (001, 002, 003, 004, 008, 009, 010, 011) green with recorded command, verbatim output, exit code, and run-tree SHA in progress.md §E.2.
- AC-ABG-016 swept count shows no empty sweep and no skip on the run machine.
- `go vet ./internal/spec/...` and `golangci-lint run ./internal/spec/...` clean.
- AC-ABG-015 live evidence recorded by the lead after install.
