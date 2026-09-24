---
id: SPEC-ACSNAPSHOT-COMMIT-GUARD-001
title: "Plan — commit-time AC-snapshot guard (git config-defined pre-commit hook)"
version: "0.2.0"
created: 2026-09-24
author: manager-spec (card t1150)
---

# Plan — SPEC-ACSNAPSHOT-COMMIT-GUARD-001

Tier M · cycle_type: tdd · local-only dev tooling (`scripts/ac-baseline/`) + one repository-local Go test file in `internal/spec/` + one tracked procedure doc. 16 requirements / 16 acceptance criteria (Tier M ceilings).

## §1 Decisions ordered by reversibility

The decisions most likely to change are listed first; mechanical steps are at the bottom (§5).

| # | Decision | Chosen | Reversal cost | Grounds |
|---|---|---|---|---|
| D1 | Enforcement mechanism | git config-defined `pre-commit` hook | high (re-plan) | Operator decision (spec.md §A.4-1) |
| D2 | Scope | index status `M`, depth-1, `_archive/` excluded | medium | Operator decision (spec.md §A.4-2) |
| D3 | Installation owner | the lead, once, after the develop merge; lane never touches shared config | medium | Operator decision (spec.md §A.4-3) |
| D4 | Comparison source | counter, acceptance blob, and baseline all read from the INDEX | medium | Makes "baseline staged in the same commit" a pass by construction; judges exactly what will be committed (REQ-ABG-001/004); pinned by AC-ABG-009 |
| D5 | Tool-fault policy | fail open, loud `NOT CHECKED` line; a mismatch elsewhere in the commit still rejects | low | spec.md §A.5 — CI backstop unchanged, shared config, agents cannot bypass (F9) |
| D6 | Missing-script policy (unabsorbed trees, primary checkout) | exit 0 with `NOT CHECKED` | low | Shared config reaches trees whose branch predates the script (F11, §A.4-3) |
| D7 | File-derived text handling | NUL-delimited enumeration; only counter body + `AC_FILE` reach `sh`; exact-string baseline lookup | low | spec.md §A.6 trust boundary; pinned by AC-ABG-010 |
| D8 | Implementation language | POSIX `sh` + `awk` + `git` | medium | Counter is already `sh`+`awk`; no Go build step in a hook; no product binary change |
| D9 | Comparison-logic twin | shell reimplementation of `acComparison`, pinned by a parity subtest that iterates EVERY row of `TestACBaselineComparisonTransitions`' case table plus an out-of-order HALT-identifier row | medium | The counter itself is never copied (REQ-ABG-007); drift in the small comparison is caught mechanically (AC-ABG-006) |
| D10 | Hook name | `ac-baseline-guard` | low | Short, names the artifact guarded |

### Alternatives considered (recorded, operator-decided)

| Candidate | What it would do | Why not chosen |
|---|---|---|
| **git config-defined pre-commit hook** (CHOSEN) | Rejects the commit at creation time, in every worktree, for every committer (human or agent) | — Operator decision 1. Coexists with the managed hookdir file per `man git-hook` (F6); one install covers all worktrees (F11). |
| Claude Code PreToolUse hook | Deny a `git commit` Bash call whose index carries the defect | Covers only Claude sessions, not terminal commits; the PreToolUse hook has a 10 s budget and a history of being killed by it (`pre_tool.go` gate-relocation comment); would ship into the product hook binary. |
| `moai spec lint` rule | Report snapshot mismatch as a lint finding | Lint is not run at commit time by lanes, and is a product surface (distributed); it reports rather than blocks — the same "documentation-shaped" control that failed twice. |
| Automatic regeneration in the hook | Regenerate and stage the snapshot on commit | Forbidden by `ac-count-baseline-refresh.md` §6: regeneration is a human-reviewed measurement, and silent absorption would bless an unintended count move. |
| (not a candidate) extend the managed `.git/hooks/pre-commit` | — | The managed file is rewritten by `moai hook install` and provenance-checked (F10); the installer ships no `pre-commit.local`. Editing it is lost on the next install and trips provenance. |

## §2 File inventory (run phase)

| File | Action | Purpose |
|---|---|---|
| `scripts/ac-baseline/check-staged.sh` | create | The guard: NUL-delimited qualifying-file selection, counter extraction from the staged `manager-docs.md`, staged-blob measurement, exact-string staged-baseline lookup, comparison with mismatch-over-fault precedence, messages. |
| `scripts/ac-baseline/install-hook.sh` | create | Idempotent `git config` writer for `hook.ac-baseline-guard.{event,command}`; git ≥ 2.54 check; never touches `.git/hooks/*` or `core.hooksPath`. Run by the lead after the develop merge, never by the lane against the live repo. |
| `internal/spec/ac_baseline_commit_guard_test.go` | create | Repository-local tests: throwaway repos in `t.TempDir()`, all pass/reject/fault/no-op/precedence/hostile-path paths, parity rows, installer idempotency and old-git refusal, real `git commit` coexistence. |
| `internal/spec/testdata/ac_baseline_guard/` | create | acceptance and baseline fixtures only. It carries NO `manager-docs.md`: the test writes the counter carrier into each temp repo at runtime by copying the tracked `.claude/agents/moai/manager-docs.md` (R4). |
| `.moai/docs/ac-count-baseline-refresh.md` | modify | New section per REQ-ABG-016. |

NOT touched: `internal/spec/ac_count_clause_test.go` semantics (the new test only calls `acComparison` and reads the case table), the snapshot, any `acceptance.md`, `internal/template/templates/**`, `internal/cli/hook_install_precommit.go`, `.git/hooks/*`, `.github/workflows/**`, the shared `.git/config`, `CLAUDE.local.md` (sync-phase note only).

To let the parity subtest iterate the same rows, the run phase may lift the anonymous case slice of `TestACBaselineComparisonTransitions` into a package-level test variable in the same file — a test-only refactor with no change to any assertion (verified by that test still passing unchanged).

## §3 Milestones (priority order)

### M0 — Priority High: premise probe (before any implementation)

In a `t.TempDir()` repo (the first subtest written, `commit/premise_probe`): install a config hook whose command prints a marker and exits 1, set `core.hooksPath=/dev/null`, run `git commit`, and assert exit non-zero with `HEAD` unchanged; add a linked worktree and repeat with `git commit -a`. This re-measures the auditor's scratch-repo observations (spec.md F6) on the run machine before M1 builds on them. **Stop condition:** if the config hook does not fire or does not abort, stop and return a blocker report — do not modify `core.hooksPath`.

### M1 — Priority High: guard predicate and comparison (REQ-ABG-001..010)

RED first: `ac_baseline_commit_guard_test.go` builds a throwaway repo per subtest (`git init` in `t.TempDir()`, env `GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_NOSYSTEM=1`, explicit author identity), seeds `HEAD` with the copied counter carrier, one recorded `acceptance.md`, and a baseline, then stages the scenario and invokes `scripts/ac-baseline/check-staged.sh` directly (cwd = temp repo root). Subtests fail before the script exists.

GREEN: write `check-staged.sh`:
- enumerate `git diff --cached --name-status -z --no-renames HEAD`; keep status `M` whose path is exactly `.moai/specs/<one segment>/acceptance.md` with a segment other than `_archive` (F12 — do not trust pathspec globbing);
- zero qualifying → `exit 0`, no output, before any extraction;
- extract the counter from `git show :.claude/agents/moai/manager-docs.md` between the sentinels, requiring exactly one pair, END after BEGIN, non-empty body;
- write each staged blob (`git show ":$path"`, path passed as one quoted argument) to a `mktemp` file removed by `trap`; run `sh -c "$counter"` with `AC_FILE` set to the temp file; classify exit 0 / 3 / other;
- sort HALT identifiers before comparing;
- look up the record in `git show :.moai/reports/t338/ac-count-baseline.txt` by exact first-field equality (e.g. `awk -v p="$path" '$1 == p'`, never a pattern built from the path); compare per REQ-ABG-002;
- accumulate per-file results, then decide: any mismatch → exit non-zero (fault lines still printed); else faults → exit 0 with `NOT CHECKED`; else `checked` line;
- reject text reuses the regeneration command and doc path verbatim (F3).

Verification: `go test ./internal/spec -run 'TestACBaselineCommitGuard' -count=1 -v` — every named subtest listed as `--- PASS`, no `[no tests to run]`.

### M2 — Priority High: installer and real-commit coexistence (REQ-ABG-011..015)

RED first: installer subtests (idempotency, old-git refusal via a stub `git` earlier on `PATH` printing `git version 2.53.0`, missing-script tolerance, managed-file byte-identity) and real-commit subtests: install into the temp repo, place an executable hookdir `pre-commit` that writes a marker file, then `git commit` rejecting and passing scenarios, with and without `core.hooksPath=/dev/null`, plus `git commit -a` and `git commit <path>`. These subtests `t.Skip` with an explicit message on git < 2.54 and on Windows (D15 — see R7); the run-phase evidence MUST show them executed, not skipped, on this machine (git 2.54.0, F5).

GREEN: write `install-hook.sh`; the configured command runs the checker if present and otherwise prints the REQ-ABG-013 line and exits 0.

### M3 — Priority Medium: documentation and install hand-off (REQ-ABG-016)

Update `.moai/docs/ac-count-baseline-refresh.md`. Record in progress.md the exact install command for the lead and the post-install observation the lead records as AC-ABG-015 evidence.

### M4 — Priority Low: mechanical closure

`go vet ./internal/spec/...`, `golangci-lint run ./internal/spec/...`, the existing `TestACCounterFullCorpusMatchesBaseline` and `TestACBaselineComparisonTransitions` still exit 0 on the SPEC's own tree (this SPEC's `acceptance.md` appears as an absent-from-snapshot report, not a failure), evidence exported under `.moai/reports/t1150/`.

## §4 Risks

| # | Risk | Mitigation |
|---|---|---|
| R1 | Config hooks might not fire under `core.hooksPath=/dev/null` on the run machine | Auditor observed they do (F6); M0 re-measures first with a stop condition; AC-ABG-011 is release-blocking |
| R2 | Multi-hook ordering with a hookdir file present is unobserved on the live repo | Recorded (not asserted) by AC-ABG-011's marker; the guard only needs "a guard failure aborts the commit" |
| R3 | Shell comparison drifts from `acComparison` | Parity subtest iterates the full case table plus an out-of-order row (AC-ABG-006) |
| R4 | A committed fixture copy of the counter would be the second instrument REQ-ABG-007 forbids | testdata carries no `manager-docs.md`; tests copy the tracked file into the temp repo at runtime |
| R5 | False positive stalls a lane with no agent-side bypass (F9) | Pass-path ACs 002–004 and 009 are release-blocking; tool faults fail open |
| R6 | Trees without the script print `NOT CHECKED` on every commit (primary checkout until release) | Accepted and recorded with the install decision (spec.md §A.4-3) |
| R7 | Windows: the ubuntu `test` job runs M1 subtests; the release-time Windows leg runs them too, and would run M2 whenever its git is ≥ 2.54 | M1 subtests invoke the checker through `sh` exactly as the existing `runCounter` does in this package; M2 real-commit subtests skip on `runtime.GOOS == "windows"` explicitly (hook-command shell semantics under Git for Windows are unmeasured), not only on the version gate |
| R8 | Pathspec `*` crossing `/` admits `_archive/` files (F12) | Explicit depth/segment filter; `_archive` and depth-2 fixtures (AC-ABG-005) |
| R9 | Hostile path text (`$(…)`, spaces, regex metacharacters) reaching a shell or regex | §A.6 trust boundary; AC-ABG-010 fixture asserts right record compared and no side-effect file created |

## §5 Mechanical steps (least likely to change)

- `chmod +x` both scripts; explicit fault branches rather than `set -e` (a fault must not surface as an accidental non-zero exit).
- The checker must not change directory before its git calls: `GIT_INDEX_FILE` may be relative (`.git/index`, spec.md F6) and is resolved against the tree root the hook starts in.
- Test helpers: `runGit`, `stage`, `commit` with isolated env; `t.TempDir()` only (CLAUDE.local.md §6).
- Every commit on this branch carries the card id `t1150` and the `Authored-By-Agent: <agent>` trailer (the `OwnershipTransitionRule` WHO signal).
- Doc section in Korean, matching the existing doc's register.

## §6 Open items

- None. The plan-phase clarification (installation owner and timing) is resolved by operator decision, recorded in spec.md §A.4-3.
