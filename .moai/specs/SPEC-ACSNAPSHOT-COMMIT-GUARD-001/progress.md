# progress.md — SPEC-ACSNAPSHOT-COMMIT-GUARD-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored 2026-09-24 by manager-spec (card t1150), Tier M, 4 files (spec.md, plan.md, acceptance.md, progress.md), worktree `.claude/worktrees/t1150`, branch `WT-ac-snapshot-guard`, base `60017eb83`.
- SPEC ID pre-write regex check: `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` on `SPEC-ACSNAPSHOT-COMMIT-GUARD-001` → `PASS`. No existing directory matched `*COMMIT-GUARD*` in `.moai/specs/`.
- ID choice: no hyphenated `AC` segment, so this SPEC's own `acceptance.md` is not contaminated by the counter's boundary-less pattern (spec.md header note, §A.3 F4).
- Operator decisions recorded as binding (spec.md §A.4): git config-defined pre-commit hook; in-place `M` amendments only.
- New measured fact surfaced in plan: the shared `.git/config` sets `core.hooksPath=/dev/null` (spec.md §A.3 F7) — the managed hookdir `pre-commit` does not currently run; config-hook firing under that setting is release-blocking AC-ABG-011 (auditor observed it fires in a scratch repo; M0 re-measures first).
- Plan-audit iteration 1 (2026-09-24): FAIL 0.80 (MP-7), report `.moai/reports/plan-audit/SPEC-ACSNAPSHOT-COMMIT-GUARD-001-review-1.md`. Repaired in 0.2.0: D1–D8 closed; optional D9–D16 applied (D16: trailer on the revision commit).
- Installation owner resolved by operator decision (spec.md §A.4-3): the lead installs once after the develop merge; this lane never touches the shared git config.
- Open items: 0 clarification markers.
- Plan-audit iteration 2 (2026-09-24): FAIL 0.92 on N1 only (D1–D16 confirmed closed), report `.moai/reports/plan-audit/SPEC-ACSNAPSHOT-COMMIT-GUARD-001-review-2.md`. Repaired in 0.2.1: N1 whitespace-free hostile path, N2 `ENVIRON`-based lookup example, N3 live-install observation demoted to a post-merge lead observation.
- plan_status: audit-ready (pending re-audit)
- The baseline snapshot was NOT regenerated; this SPEC's `acceptance.md` is an absent-from-snapshot report row until the next reviewed regeneration.

## §E.2 Run-phase Evidence

Run by manager-develop (cycle_type=tdd), 2026-09-24, worktree `.claude/worktrees/t1150`, branch `WT-ac-snapshot-guard`, base `60017eb83`. Machine: darwin, `git version 2.54.0 (Apple Git-157)`. Code measured: the tree committed as `00ad53a9e` (scripts, tests, fixtures unchanged after that commit; only the doc and this file changed afterwards). The shared repository git config was never written; every real-commit test ran in a `t.TempDir()` repo with `GIT_*` stripped, `GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_NOSYSTEM=1`, per-test `HOME` on the `exec.Cmd`.

Raw outputs are in `.moai/reports/t1150/` (gitignored by `.gitignore:235 .moai/reports/*`, so they live only in this worktree; the lines below are quoted verbatim from them).

### M0 — premise probe (before any implementation)

Command: `go test ./internal/spec -run 'TestACBaselineCommitGuard/commit/premise_probe' -count=1 -v` → exit 0 (`m0-premise-probe.txt`):

```
    ac_baseline_commit_guard_test.go:174: leg1 hooksPath=/dev/null: exit=1 stderr="PROBE-FIRED\n"
    ac_baseline_commit_guard_test.go:188: leg2 hookdir present: exit=1 hookdir_marker_written=true stderr="PROBE-FIRED\n"
    ac_baseline_commit_guard_test.go:197: leg3 linked worktree: exit=1 stderr="PROBE-FIRED\n"
    --- PASS: TestACBaselineCommitGuard/commit/premise_probe (1.12s)
```

A config-defined pre-commit hook fired and aborted the commit (HEAD unchanged, asserted) under `core.hooksPath=/dev/null`, alongside an executable hookdir `pre-commit` (which ALSO ran — marker written), and from a linked worktree via `git commit -a`. Premise holds; no blocker.

### RED (scripts absent)

Command: `go test ./internal/spec -run 'TestACBaselineCommitGuard' -count=1 -v` → exit 1 (`red-run.txt`): 42 subtests `--- FAIL`, only `commit/premise_probe` (a git probe, not the guard) passing. Representative verbatim lines:

```
    ac_baseline_commit_guard_test.go:339: want exit 0 and exactly one "ac-baseline-guard: checked 1" line; exit=127 stderr="sh: scripts/ac-baseline/check-staged.sh: No such file or directory\n"
    ac_baseline_commit_guard_test.go:387: baseline edited but not staged must still reject; exit=127 stderr="sh: scripts/ac-baseline/check-staged.sh: No such file or directory\n"
    --- FAIL: TestACBaselineCommitGuard/reject_count_move (0.48s)
FAIL	github.com/modu-ai/moai-adk/internal/spec	32.533s
```

The first RED run showed two reject subtests passing vacuously on the exit-127 of the absent script; they were tightened to require the guard's own `ac-baseline-guard: REJECT` line before GREEN (above is the re-run).

### Mutation probes (after GREEN, each reverted; `cmp` confirmed the restore)

| Mutant in `check-staged.sh` | Subtests that went red |
|---|---|
| exact lookup `($1 "") == want` → regex `$1 ~ want` | `hostile_path` |
| fail-open `exit 0` → `exit 1` | all seven `fault/*` |
| malformed-record `continue` → `exit 0` (fault swallows a later mismatch) | `mixed/mismatch_plus_fault` |
| `git show ":$p"` → `cat "$p"` (working tree instead of index) | `index_only/mm_unstaged_criterion`, `index_only/staged_criterion_reverted_tree` |
| `git show ":$carrier"` → `cat "$carrier"` | `index_only/counter_from_index` |

The regex mutant also exposed a real gap: an erroring lookup `awk` left `rec` empty and the file counted as checked with no problem. Fixed before commit — any lookup result other than `COUNT …`/`HALT …`/`ABSENT`/`MALFORMED` is now a `NOT CHECKED (snapshot lookup failed)` fault.

### GREEN — AC matrix

Selector command (AC-ABG-001..014, 016): `go test ./internal/spec -run 'TestACBaselineCommitGuard' -count=1 -v` → exit 0, `ok  	github.com/modu-ai/moai-adk/internal/spec	23.808s`; 43 `    --- PASS` lines, 0 `--- FAIL`, 0 `--- SKIP`, 0 `no tests to run` (`ac-selector-run.txt`).

| AC | Actual Output (verbatim `--- PASS` lines) | Status |
|---|---|---|
| AC-ABG-001 | `--- PASS: TestACBaselineCommitGuard/reject_count_move (0.40s)` | PASS |
| AC-ABG-002 | `--- PASS: TestACBaselineCommitGuard/pass_count_unchanged (0.39s)` | PASS |
| AC-ABG-003 | `pass_new_file (0.34s)`, `pass_unrecorded_counts (0.71s)`, `pass_unrecorded_halts (0.52s)` — all `--- PASS` | PASS |
| AC-ABG-004 | `pass_with_staged_baseline (0.60s)`, `reject_unstaged_baseline (0.48s)` — both `--- PASS` | PASS |
| AC-ABG-005 | `noop_unrelated (0.53s)`, `noop_archive (0.40s)`, `noop_depth2 (0.46s)` — all `--- PASS` | PASS |
| AC-ABG-006 | 10 table rows `parity/count-stable`, `parity/count-moved`, `parity/count-state-moved-live-to-excluded`, `parity/b:_count-to-halt`, `parity/c:_halt-to-count`, `parity/e:_halting-id-set-moved`, `parity/halt-stable`, `parity/absent-and-counts:_report,_do_not_fail`, `parity/absent-and-halts:_report,_do_not_fail_(§3.5_rule_4)`, `parity/d:_recorded-COUNT_starts_halting` + `parity/halt_ids_unsorted` — all `--- PASS` | PASS |
| AC-ABG-007 | `fault/sentinel_absent`, `fault/sentinel_duplicated`, `fault/end_before_begin`, `fault/empty_body`, `fault/baseline_absent`, `fault/baseline_line_malformed`, `fault/counter_exit_2` — all `--- PASS`; e.g. logged `ac-baseline-guard: NOT CHECKED (counter exited 2): .moai/specs/SPEC-X-001/acceptance.md` | PASS |
| AC-ABG-008 | `--- PASS: TestACBaselineCommitGuard/mixed/mismatch_plus_fault (0.69s)` | PASS |
| AC-ABG-009 | `index_only/mm_unstaged_criterion (0.77s)`, `index_only/staged_criterion_reverted_tree (0.60s)`, `index_only/counter_from_index (0.48s)` — all `--- PASS` | PASS |
| AC-ABG-010 | `--- PASS: TestACBaselineCommitGuard/hostile_path (0.46s)` (no `pwned` under the temp root; right record `live=2` named) | PASS |
| AC-ABG-011 | `commit/premise_probe (0.81s)`, `commit/hookdir_present (1.11s)`, `commit/hookspath_devnull (0.74s)`, `commit/linked_worktree (0.63s)` — all `--- PASS`; logged `reject case: hookdir marker written=true`, `pass case: exit=0 hookdir marker written=true` | PASS |
| AC-ABG-012 | `commit/all_flag (0.56s)`, `commit/pathspec_only (0.86s)` — both `--- PASS` | PASS |
| AC-ABG-013 | `install/idempotent (0.55s)`, `install/old_git (0.59s)`, `install/missing_script (0.54s)` — all `--- PASS` | PASS |
| AC-ABG-014 | `--- PASS: TestACBaselineCommitGuard/install/managed_untouched (1.28s)` | PASS |
| AC-ABG-015 | `grep -c 'hook.ac-baseline-guard' .moai/docs/ac-count-baseline-refresh.md` → `2` (exit 0); §7 names `NOT CHECKED`, `--no-verify`, `SKIP_MOAI_PRECOMMIT`, the lead as install owner (§7.2), the completion-report quoting obligation (§7.3), and `sh scripts/ac-baseline/install-hook.sh` | PASS |
| AC-ABG-016 | selector sweep above (43 PASS, 0 SKIP); `grep -c 'moai-ac-prefix' scripts/ac-baseline/check-staged.sh` → `0`; `git diff --name-only 60017eb83 -- internal/template/templates .github/workflows` → empty; corpus gate: `--- PASS: TestACCounterFullCorpusMatchesBaseline (12.52s)` and `--- PASS: TestACBaselineComparisonTransitions (0.00s)` | PASS (form deviation below) |

AC-ABG-016 form deviation: the worktree-session guard refused the single `-run 'TestACCounterFullCorpusMatchesBaseline|TestACBaselineComparisonTransitions'` invocation (it cannot statically verify a `|` inside the argument), so the two tests ran as two separate `go test ./internal/spec -run <Name> -count=1 -v` invocations, each exit 0 (`corpus-gate-run.txt`, `transitions-run.txt`). The corpus run reports this SPEC's own file as `absent-from-snapshot .moai/specs/SPEC-ACSNAPSHOT-COMMIT-GUARD-001/acceptance.md: COUNT 16` — a report, not a failure, as planned.

### Invariants and package-level checks

| Check | Command | Output | Status |
|---|---|---|---|
| Whole affected package | `go test ./internal/spec/... -count=1` | `ok  	github.com/modu-ai/moai-adk/internal/spec	169.829s` (exit 0) | PASS |
| vet | `go vet ./internal/spec/...` | (empty, exit 0) | PASS |
| lint | `golangci-lint run ./internal/spec/...` | `0 issues.` (exit 0) | PASS |
| Repo-tree write guard | `TestNoTestWritesRepoTree` | first package run flagged 5 textual false positives (temp-repo writes via a `.moai/`-tainted identifier); routed through a parameter-scoped `writeTempFile` helper → `--- PASS: TestNoTestWritesRepoTree (0.03s)` | PASS |
| Transition-table refactor | `TestACBaselineComparisonTransitions` | `--- PASS` — table lifted to `acComparisonTransitionCases`, no assertion changed | PASS |
| Shared repo config untouched | no `git config` without `-C <tempdir>` was run against the real repo | — | PASS (by construction) |

### Hand-off to the lead (post-merge, not a close gate — spec.md §A.4-3)

After this branch is merged into local `develop`, from any tree of this repository:

```
sh scripts/ac-baseline/install-hook.sh
git config --get-regexp '^hook\.ac-baseline-guard\.'
```

The second command must print two lines (`event pre-commit` and the `command`). Then observe one rejected and one passing commit in a develop-absorbed tree and record it as a post-merge observation.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-24
run_commit_sha: 00ad53a9e, f173cb0b3   # 00ad53a9e = M1-M2 code+tests; f173cb0b3 = M3-M4 doc + evidence
run_status: complete
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 0   # no existing test deleted or weakened; one table lifted verbatim
l44_pre_commit_fetch: not-run     # lane pushes nothing; lead batches the develop push
l44_post_push_fetch: not-applicable
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: measured (this machine)
  linux: not-measured (CI)
  windows: not-measured; real-commit and installer subtests skip on windows by design (plan.md R7)
total_run_phase_files: 12   # 2 scripts, 1 new test, 1 refactored test, 5 fixtures, spec.md, cascade doc, progress.md
m1_to_mN_commit_strategy: two commits (code+tests+status, then doc+evidence)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-24
sync_commit_sha: pending-backfill-sync   # this commit cannot cite its own hash
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-ACSNAPSHOT-COMMIT-GUARD-001' CHANGELOG.md (pre-emission) -> 0, no duplicate"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-ACSNAPSHOT-COMMIT-GUARD-001/acceptance.md | sort -u | wc -l -> 16, matches CHANGELOG entry's stated AC count"
b12_self_test_c: "ls scripts/ac-baseline/check-staged.sh scripts/ac-baseline/install-hook.sh internal/spec/ac_baseline_commit_guard_test.go .moai/docs/ac-count-baseline-refresh.md -> all exist"
changelog_entry_position: "top of [Unreleased] > Added, above SPEC-DUAL-HARNESS-RECOVERY-001"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (status field only; updated already read 2026-09-24)"
  plan_md: "no status field present (frontmatter carries id/title/version/created/author only) - no transition applicable"
  acceptance_md: "no status field present (frontmatter carries id/title/version/created/author only) - no transition applicable"
canary_compliance_check: "not applicable - this SPEC defines no forward-looking policy that its own sync tests"
```

No README or docs-site change: this SPEC ships a repository-local dev-only commit-time guard (`scripts/ac-baseline/`, installed into the shared `.git/config` by the lead after the develop merge, per spec.md §A.4-3), with no user-facing surface. `grep -c 'ac-baseline\|ACSNAPSHOT-COMMIT-GUARD' README.md README.ko.md` → 0; `scripts/ac-baseline` and this SPEC ID do not appear under `docs-site/content`.
