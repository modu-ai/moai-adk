---
id: SPEC-ACSNAPSHOT-COMMIT-GUARD-001
title: "Plan — commit-time AC-snapshot guard (git config-defined pre-commit hook)"
version: "0.1.0"
created: 2026-09-24
author: manager-spec (card t1150)
---

# Plan — SPEC-ACSNAPSHOT-COMMIT-GUARD-001

Tier M · cycle_type: tdd · local-only dev tooling (`scripts/ac-baseline/`) + one repository-local Go test file in `internal/spec/` + one tracked procedure doc.

## §1 Decisions ordered by reversibility

The decisions most likely to change are listed first; mechanical steps are at the bottom (§5).

| # | Decision | Chosen | Reversal cost | Grounds |
|---|---|---|---|---|
| D1 | Enforcement mechanism | git config-defined `pre-commit` hook | high (re-plan) | Operator decision (spec.md §A.4-1) |
| D2 | Scope | index status `M`, depth-1, `_archive/` excluded | medium | Operator decision (spec.md §A.4-2) |
| D3 | Comparison source | both counter and baseline read from the INDEX | medium | Makes "baseline staged in the same commit" a pass by construction; judges exactly what will be committed (spec.md REQ-ABG-001/004) |
| D4 | Tool-fault policy | fail open, loud `NOT CHECKED` line | low (flip a branch) | spec.md §A.5 — CI backstop unchanged, shared config, agents cannot bypass (F9) |
| D5 | Missing-script policy in old worktrees | exit 0 with `NOT CHECKED` | low | Shared config reaches worktrees whose branch predates the script (F11) |
| D6 | Implementation language | POSIX `sh` + `awk` + `git` | medium | Counter is already `sh`+`awk`; no Go build step in a hook; no product binary change (Exclusions) |
| D7 | Comparison-logic twin | shell reimplementation of `acComparison`, pinned by a parity test over the same transition rows as `TestACBaselineComparisonTransitions` | medium | The counter itself is never copied (REQ-ABG-007); the comparison is small, and drift is caught mechanically (AC-ABG-006) |
| D8 | Hook name | `ac-baseline-guard` | low | Short, names the artifact guarded |

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
| `scripts/ac-baseline/check-staged.sh` | create | The guard: qualifying-file selection, counter extraction from the staged `manager-docs.md`, staged-blob measurement, staged-baseline lookup, comparison, messages. |
| `scripts/ac-baseline/install-hook.sh` | create | Idempotent `git config` writer for `hook.ac-baseline-guard.{event,command}`; git ≥ 2.54 check; never touches `.git/hooks/*` or `core.hooksPath`. |
| `internal/spec/ac_baseline_commit_guard_test.go` | create | Repository-local tests: throwaway repos in `t.TempDir()`, all pass/reject/fault/no-op paths, parity rows, installer idempotency and old-git refusal, real `git commit` coexistence. |
| `internal/spec/testdata/ac_baseline_guard/` | create | Minimal fixtures: a counter-carrier `manager-docs.md` (the real sentinel block copied in by the test at runtime from the tracked file, NOT committed as a second copy — see §4 R4), acceptance fixtures, baseline fixtures. |
| `.moai/docs/ac-count-baseline-refresh.md` | modify | New section: the mechanical gate, install command, `NOT CHECKED` semantics, bypass and its agent-side unavailability (REQ-ABG-017). |

NOT touched: `internal/spec/ac_count_clause_test.go` semantics, the snapshot, any `acceptance.md`, `internal/template/templates/**`, `internal/cli/hook_install_precommit.go`, `.git/hooks/*`, `.github/workflows/**`, `CLAUDE.local.md` (sync-phase note only).

## §3 Milestones (priority order)

### M1 — Priority High: guard predicate and comparison (REQ-ABG-001..010)

RED first: `ac_baseline_commit_guard_test.go` builds a throwaway repo per subtest (`git init` in `t.TempDir()`, env `GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_NOSYSTEM=1`, explicit author identity), seeds `HEAD` with a counter-carrier, one recorded `acceptance.md`, and a baseline, then stages the scenario and invokes `scripts/ac-baseline/check-staged.sh` directly (cwd = temp repo). Subtests fail before the script exists.

GREEN: write `check-staged.sh`:
- select `git diff --cached --name-status --no-renames HEAD`; keep status `M` whose path matches exactly `.moai/specs/<one segment>/acceptance.md` and whose segment is not `_archive` (F12 — do not trust pathspec globbing);
- zero qualifying → `exit 0`, no output, before any extraction;
- extract the counter from `git show :.claude/agents/moai/manager-docs.md` between the sentinels, requiring exactly one pair and a non-empty body;
- write each staged blob (`git show :<path>`) to a `mktemp` file removed by `trap`, run the extracted counter with `AC_FILE` set, classify exit 0 / 3 / other;
- look up the file's line in `git show :.moai/reports/t338/ac-count-baseline.txt`, compare per REQ-ABG-002;
- print reject / checked / NOT CHECKED lines per REQ-ABG-003/008/009, the reject text reusing the regeneration command and doc path verbatim from F3.

Verification: `go test ./internal/spec -run 'TestACBaselineCommitGuard' -count=1 -v` — every named subtest listed as `--- PASS`, no `[no tests to run]`.

### M2 — Priority High: installer and real-commit coexistence (REQ-ABG-011..015)

RED first: installer subtests (idempotency, old-git refusal via a stub `git` earlier on `PATH` printing `git version 2.53.0`, managed-file byte-identity) and real-commit subtests: install into the temp repo, place an executable hookdir `pre-commit` that writes a marker file, then `git commit` a rejecting scenario (HEAD unchanged, exit non-zero) and a passing one (new commit; record whether the marker ran). Repeat with `core.hooksPath=/dev/null` (the live repository's state, F7). These subtests `t.Skip` with an explicit message on git < 2.54; the run-phase evidence MUST show them executed, not skipped, on this machine (git 2.54.0, F5).

GREEN: write `install-hook.sh`; command value runs the checker if present and otherwise prints the REQ-ABG-013 line and exits 0.

**Stop condition:** if the config hook does not fire under `core.hooksPath=/dev/null`, stop and return a blocker report — do not modify `core.hooksPath`; the operator decides.

### M3 — Priority Medium: documentation and live install hand-off (REQ-ABG-016/017)

Update `.moai/docs/ac-count-baseline-refresh.md`. Record in progress.md the exact install command for the lead. Live installation into the shared `.git/config` is a shared-state mutation that arms the guard for every lane; see §6 clarification.

### M4 — Priority Low: mechanical closure

`go vet ./internal/spec/...`, `golangci-lint run ./internal/spec/...`, the existing `TestACCounterFullCorpusMatchesBaseline` still exits 0 on the SPEC's own tree (this SPEC's `acceptance.md` appears as an absent-from-snapshot report, not a failure), evidence exported under `.moai/reports/t1150/`.

## §4 Risks

| # | Risk | Mitigation |
|---|---|---|
| R1 | Config hooks may not fire under `core.hooksPath=/dev/null` (F7) | Release-blocking criterion AC-ABG-008b; stop condition in M2 |
| R2 | Multi-hook failure semantics (does a failing config hook still let the hookdir hook run? does any failure abort?) are unobserved (F6) | Measured and recorded in AC-ABG-008; the guard only needs "a guard failure aborts the commit" |
| R3 | Shell comparison drifts from `acComparison` | Parity subtests over the same transition rows (AC-ABG-006) |
| R4 | A committed fixture copy of the counter would be the second instrument REQ-ABG-007 forbids | Tests copy the sentinel block from the tracked `manager-docs.md` into the temp repo at runtime |
| R5 | False positive stalls a lane with no agent-side bypass (F9) | Pass-path ACs 002–004 are release-blocking; tool faults fail open |
| R6 | Worktrees on older branches lack the script | REQ-ABG-013 `NOT CHECKED` line; disappears once lanes absorb develop |
| R7 | Windows CI runner `sh` availability | The existing `runCounter` already runs `sh -c` in this package; real-commit subtests skip where `git` < 2.54 |
| R8 | Pathspec `*` crossing `/` admits `_archive/` files (F12) | Explicit depth/segment filter; fixture with an `_archive/` M file |

## §5 Mechanical steps (least likely to change)

- `chmod +x` both scripts; `set -eu`-style strictness with explicit fault branches (a fault must not surface as a `set -e` non-zero exit).
- Test helpers: `runGit`, `stage`, `commit` with isolated env; `t.TempDir()` only (CLAUDE.local.md §6).
- Doc section in Korean, matching the existing doc's register.

## §6 Open items

- [NEEDS CLARIFICATION: who runs `scripts/ac-baseline/install-hook.sh` against the shared `.git/config`, and when] — installing arms the guard for every worktree at once (F11). Recommended: the lead, after this card's branch is merged into local develop, with `git config --get-regexp '^hook\.ac-baseline-guard\.'` recorded as AC-ABG-012 evidence. The lane does not install on the live repository during run.
