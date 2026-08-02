# Progress — SPEC-WORKTREE-SQUASH-MERGE-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-02
plan_revision: 4          # iteration-4 revision after plan-audit FAIL (0.79), user-override scoped to N5
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
requirements: 14   # REQ-WSM-001 .. REQ-WSM-014        (Tier M cap 16)
acceptance_criteria: 16   # AC-WSM-001 .. AC-WSM-016   (Tier M cap 16)
open_clarifications: 0
```

Scope decisions resolved during authoring, each from executed git experiments rather than
reasoning: the false-positive boundary (`spec.md` §5 decision 1), synthetic-object handling
(decision 2), the `--merged-only` co-change (decision 3), and the preserved safety
constraints (decision 4). Four falsified alternatives are recorded in `spec.md` §2 so they
are not re-proposed: plain per-commit `git cherry` as the squash detector; history-only
patch-id equivalence with no state check; enumerating S5's paths with rename detection on;
and splitting the enumeration's output on newline rather than NUL.

### Iteration-2 revision (plan-audit D1-D5 + D6-D8)

- **D2** — REQ-WSM-007 narrowed with a measured per-probe verdict-code table; the earlier
  wording made S3/S4 unreachable because `git diff --quiet` uses rc 1 as a verdict.
- **D1** — added the S5 state conjunct (REQ-WSM-014): S3/S4 are no longer sufficient alone.
  Matrix re-run at 9 scenarios × 5 signals; SC-2 (rebase) verified not to regress.
- **D3** — AC-WSM-009's judge repaired (`-r` dropped, exclusion anchored on the literal
  argument pair) and its falsification restated using `git prune`, then executed.
- **D4** — the clean-worktree-with-unpushed-commits removal case recorded in §5 decision 3,
  accepted with the branch-retention mitigation, and listed in §4 as out of scope.
- **D5** — AC-WSM-015 (REQ-001 + REQ-014) and AC-WSM-016 (REQ-011 + REQ-012) added;
  every requirement now has a binding criterion beyond the AC-WSM-014 catch-all.
- **D6** — AC-WSM-012 rewritten to match on the `false` argument and pin both call sites.
- **D7** — the unrelated-histories case folded into AC-WSM-007 as a second row.
- **D8** — §2's no-false-positive claim explicitly scoped to the constructed matrix.

Two residual risks the audit flagged are addressed rather than carried: `plan.md` §F's
`git cherry` cost claim is corrected with measurement (the bound holds for neither S3 nor
S4), and AC-WSM-008 now records the inputs that produce its hash.

### Iteration-3 revision (plan-audit N1-N4)

- **N1** — the rename-fold false positive closed: S5's enumeration takes `--no-renames`.
  Matrix extended to 11 scenarios with SC-10 and SC-11; SC-1..SC-9 unchanged.
- **N2** — non-vacuity guards added to AC-WSM-001, AC-WSM-010, AC-WSM-015.
- **N3** — SC-9's construction pinned to one recipe, with the SHA-collision degeneration
  recorded so the cherry-pick phrasing is not substituted.
- **N4** — the S5 shell snippet rewritten as an array; `plan.md` §D gained the
  `[]string`-not-a-joined-string subsection.

### Iteration-4 revision (plan-audit N5-N6, user override)

Scoped by explicit user override of the three-iteration ceiling to **one finding**. No
resolved finding was re-opened and no structural change was made beyond the two below.

- **N5** — the path-encoding false positive closed: S5's enumeration takes `-z` and its
  output is split on NUL. `git diff --name-only` C-quotes any path containing a backslash,
  tab, double quote, or control character — and, under the documented default
  `core.quotePath=true`, any non-ASCII path — and the quoted rendering matches nothing,
  which by §2 finding 5 reads as merged. Matrix extended to 12 scenarios with SC-12;
  SC-1..SC-11 re-executed and cell-for-cell unchanged. §C.1's mutation grid re-run at
  6 × 12; the new `newline-split` mutation flips exactly SC-12, disjoint from
  `folded-names`, which is the evidence that `--no-renames` and `-z` are separate guards.
  AC-WSM-006 was extended to carry both falsifications rather than a seventeenth criterion
  being added, keeping the SPEC at the Tier M ceiling of 16.
  The load-bearing part of this repair is a **correction**, not an addition: `plan.md` §D
  previously stated that a path list built with `strings.Split(out, "\n")` is *safe*. That
  claim was false in the unsafe direction — it endorsed the exact construction that
  produces the SC-12 false positive — and leaving it in place was the stated reason the
  override was authorized.
- **N6** — run-phase judging commands added to AC-WSM-002 through AC-WSM-008, so §F's
  "the judging command and its observed output recorded" requirement now holds for every
  criterion as written rather than for a subset. All twelve `-run` selectors were executed
  against the pre-implementation tree and their baseline `--- PASS:` counts recorded in §A
  (nine at 0, and 3 / 1 / 2 for the three pre-existing tests).

## §E.2 Run-phase Evidence

### M1 — predicate S1-S5 composition + guards + 17-scenario suite

Implementation: `internal/core/git/worktree.go` `IsBranchMerged` rewritten as
ordered OR over S1 (reachability, retained verbatim incl. `trimBranchListMarker`),
S2 (empty diff), (S3 rebase-merge ∧ S5), (S4 squash-merge ∧ S5). merge-base and
`names` computed once. `internal/core/git/manager.go` gains `execGitExit`
(exit-code-aware executor + the per-call env hook mechanism for M2); `execGit`
unchanged. Eight guards verbatim (plan.md §D).

RED captured before GREEN (TDD, acceptance.md §A non-vacuity discipline). Against
the pre-implementation S1-only tree, the new selectors that should FAIL did fail:
```
--- FAIL: TestIsBranchMerged_SquashMerged    (false, want true — S4+S5 gap)
--- FAIL: TestIsBranchMerged_RebaseMerged    (false, want true — S3+S5 gap)
--- FAIL: TestIsBranchMerged_EmptyDiffBranch (false, want true — S2 gap)
--- FAIL: TestIsBranchMerged_Superset        (false, want true — S3+S4+S5 gap)
--- FAIL: TestIsBranchMerged_ProbeExit_Table/EmptyDiffMerged (false, want true)
```
SC-6/7/8/10-15/P3c/P4c passed coincidentally under S1-only (expected false), and
the reachability rows (SC-3/SC-4) passed (S1 handles them). Verbatim RED output
recorded above.

GREEN — full 17-scenario suite + ProbeExit table after implementation:
```
$ go test ./internal/core/git/ -run 'TestIsBranchMerged_|TestSyntheticCommit' -count=1 -v
--- PASS: TestIsBranchMerged_SquashMerged (1.10s)
--- PASS: TestIsBranchMerged_RebaseMerged (1.02s)
--- PASS: TestIsBranchMerged_Reachability_TrueMergeCommit (1.11s)
--- PASS: TestIsBranchMerged_Reachability_StrictlyBehind (0.17s)
--- PASS: TestIsBranchMerged_EmptyDiffBranch (0.23s)
--- PASS: TestIsBranchMerged_PartiallyApplied (0.50s)
--- PASS: TestIsBranchMerged_UnmergedControl (0.46s)
--- PASS: TestIsBranchMerged_RevertRemovedFromBase (0.32s)
--- PASS: TestIsBranchMerged_Superset (0.33s)
--- PASS: TestIsBranchMerged_RenameReAddSquash (1.37s)
--- PASS: TestIsBranchMerged_RenameReAddCherryPick (0.42s)
--- PASS: TestIsBranchMerged_NonASCIIPathRemoved (1.41s)
--- PASS: TestIsBranchMerged_ColonPathRemovedFromBase (0.41s)
--- PASS: TestIsBranchMerged_TextconvDriverCollapsesChange (0.51s)
--- PASS: TestIsBranchMerged_SubmoduleIgnoreDirective (2.86s)
--- PASS: TestIsBranchMerged_ModeOnlyDivergence (0.59s)
--- PASS: TestIsBranchMerged_SymlinkFileOIDCollision (0.83s)
--- PASS: TestIsBranchMerged_ProbeExit_Table (0.93s)
    --- PASS: .../UnknownBase (0.35s)
    --- PASS: .../UnrelatedHistories (0.16s)
    --- PASS: .../NonEmptyDiffNoError (0.26s)
    --- PASS: .../EmptyDiffMerged (0.16s)
PASS
ok  github.com/modu-ai/moai-adk/internal/core/git  15.104s
```

AC-WSM-013 pre-existing predicate tests unchanged:
```
$ go test ./internal/core/git/ -run 'TestWorktreeIsBranchMerged' -count=1 -v
--- PASS: TestWorktreeIsBranchMerged_BranchCheckedOutInWorktree (0.49s)
--- PASS: TestWorktreeIsBranchMerged (0.34s)
--- PASS: TestWorktreeIsBranchMerged_NotMerged (0.70s)
PASS
```

Shell-judge ACs (acceptance.md §D):
- AC-WSM-009 `grep ... "prune|gc" | wc -l` → `0`
- AC-WSM-011 `git diff --exit-code -- types.go` → exit 0 (interface unchanged)
- AC-WSM-012 `grep -cE 'Remove\([A-Za-z_.]+, false\)' clean.go` → `2`

Coverage + build + vet + lint (acceptance.md §E):
```
$ go test -cover ./internal/core/git/...   → coverage: 85.7% of statements
$ go build ./...                            → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...  → exit 0
$ go vet ./internal/core/git/               → exit 0
$ golangci-lint run ./internal/core/git/... → 0 issues. (baseline was 0; no NEW)
```

E4 no-shell grep (plan.md §D) — no match:
```
$ grep -nE 'exec\.Command(Context)?\((ctx, )?"(sh|bash|zsh)"|"-c"' internal/core/git/*.go
(exit 1, no match)
```

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-02      # M1 complete; M2-M5 pending
run_commit_sha: pending-backfill # M1 commit SHA backfilled after push (self-referential)
run_status: m1-green             # M1 GREEN; M2..M5 not yet run
ac_pass_count: 12                # AC-WSM-001..006, 007, 009, 011, 012, 013, 017 measured PASS at M1
ac_fail_count: 0
preserve_list_post_run_count: 0  # types.go unmodified; clean.go call sites untouched
l44_pre_commit_fetch: n/a        # 1-person OSS, branch-local
l44_post_push_fetch: pending     # backfilled after M5 push
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues == baseline 0
cross_platform_build:
  go_build_all: exit_0
  windows_amd64: exit_0
total_run_phase_files: 3         # worktree.go, manager.go, worktree_squash_merge_test.go
m1_to_mN_commit_strategy: per-milestone Conventional-Commit + 🗿 MoAI trailer; branch push only (repo-local PR-mandatory)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

```yaml
# Phase 4 Mode Selection — orchestrator decision log (SPEC-WORKTREE-SQUASH-MERGE-001)
tier: M
scope_files: 4            # internal/core/git/worktree.go (IsBranchMerged), manager.go (execGit env hook), new scenario test file, internal/cli/worktree/clean.go (help text M5)
domain_count: 1           # Go internal/core/git — single domain
language_mix: "100% Go"
concurrency_benefit: LOW  # coding-heavy, cohesive single-function rewrite
agent_teams_prereqs: n/a  # Mode 3 retired
```

Mode evaluation:
- Mode 1 trivial — not selected: safety-critical predicate rewrite + 17-scenario matrix; not a typo/single-line.
- Mode 2 background — not selected: write-capable implementation, not read-only.
- Mode 3 agent-team — RETIRED.
- Mode 4 parallel — not selected: single domain + coding-heavy → Anthropic coding-task parallelism caveat favors sequential.
- Mode 5 sub-agent — SELECTED: coding-heavy cohesive function; sequential TDD cycles in one agent context.
- Mode 6 workflow — not selected: <30 files, semantic new-logic (not a mechanical uniform transform).

Decision: sub-agent (Mode 5)
cycle_type: tdd (per `.moai/config/sections/quality.yaml` constitution.development_mode: tdd)

Justification: Anthropic's coding-task parallelism caveat — most coding tasks involve fewer truly parallelizable tasks than research. This SPEC is one cohesive predicate function (`IsBranchMerged`) plus its falsifiable scenario tests; the TDD RED-GREEN-REFACTOR cycles (M1→M5) share state and build on each other, so a single sequential `manager-develop` agent in one context is correct. Implementation Kickoff Approval passed (user approved run entry 2026-08-02 after Phase 1 plan-audit PASS 0.95, independent fresh-eyes verdict in `.moai/reports/plan-audit/SPEC-WORKTREE-SQUASH-MERGE-001-2026-08-02.md`).
