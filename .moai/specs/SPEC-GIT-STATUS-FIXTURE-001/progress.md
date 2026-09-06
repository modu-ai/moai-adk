# progress.md — SPEC-GIT-STATUS-FIXTURE-001

Card: t474 · Branch: `WT-git-status-fixture` · Base: `25a3212a9` (origin/develop tip)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-GIT-STATUS-FIXTURE-001
card: t474
branch: WT-git-status-fixture
base: 25a3212a9
status: draft
verified_facts:
  - "Two internal/core/git tests (TestStatusAheadBehindFromHeader :269, TestStatusBranchHeaderShapes :122/:128) are deterministically red on develop CI (dual-job: Test (ubuntu-latest) + Race Test) and green locally."
  - "Single root cause, measured: unpinned 'git init -q --bare' at status_branch_test.go:99 and :235 leaves the bare remote HEAD at ambient init.defaultBranch (main on dev machines via Xcode system gitconfig, master on CI) → unborn-HEAD clones → wrong-branch push exiting 0 → all three CI signatures reproduced with byte-identical fatal signature."
  - "Identity hypothesis eliminated: gitFixture injects GIT_AUTHOR_*/GIT_COMMITTER_* env (status_branch_test.go:22-27)."
  - "Repair precedent in-repo: helpers_test.go:42 'init --bare -b main'."
  - "Local RED not achievable in lane (worktree-scoped config forcing measured ineffective; system/global/env forms rejected for lane safety) — RED-now evidence is the L4/L5 mechanism probe + live CI red; the deciding fix verdict is the merged-tree CI run (external dependency)."
evidence: .moai/reports/t474/repro-probe.md
plan_status: audit-ready
plan_complete_at: "2026-09-03T19:04:13Z"
```

Artifacts: spec.md (4 REQ, 6 AC — Tier S inline), plan.md (M1 kickoff gate on optional :315 pin → M2 repair+local sweep → M3 external CI confirmation).

## §E.2 Run-phase Evidence

Full command + verbatim output record: `.moai/reports/t474/run-evidence.md` (tracked export).

### M1 — Kickoff decision (verbatim from the dispatch)

> The operator ACCEPTED the optional `:315` consistency pin. So the repair touches THREE sites: `:99`, `:235`, AND `:315` — all in `internal/core/git/status_branch_test.go`, and that file only. Rationale recorded: `:315` is harmless today (nothing clones from it) but the same trap remains for the next fixture copied from it.

Implementation Kickoff Approval: APPROVED by the operator (relayed via the lead).

### AC PASS/FAIL matrix

| AC | Status | Command (a) | Verbatim output (b) | Baseline attribution (c) |
|----|--------|-------------|---------------------|--------------------------|
| AC-GSF-001 | PASS | `grep -n '"init", "-q", "--bare", "-b", "main"' internal/core/git/status_branch_test.go` | `99:` / `235:` / `315:` all three sites carry `"-b", "main"` (count form `grep -c` → `3`) | this run, this worktree, pre-repair HEAD `45a089bee` for the RED-now baseline (all 3 unpinned); post-repair working tree for the green cell |
| AC-GSF-002 | PASS (regression check, not fix proof) | `go test ./internal/core/git/ -count=1` + `-v 2>&1 \| grep -c '^--- PASS'` | `ok github.com/modu-ai/moai-adk/internal/core/git 71.171s` exit=0; census `107` | this run, this worktree, post-repair working tree |
| AC-GSF-003 | PENDING-EXTERNAL | (external) CI on merged develop tree after lead's push | — not runnable from the lane | deciding verdict is `Test (ubuntu-latest)` + `Race Test` on the merged tree |
| AC-GSF-004 | PASS | `git diff --name-only 25a3212a9...HEAD` + `git diff --name-only` | committed range: 3 SPEC `.md` + probe `.md`; code change: only `internal/core/git/status_branch_test.go`; no non-test `.go` file | this run, base `25a3212a9` (pinned SHA) |
| AC-GSF-005 | PASS | `gofmt -l .` | (empty — 0 rows) exit=0 | this run, this worktree, post-repair working tree |
| AC-GSF-006 | PASS | (decision record) | kickoff **accept** branch taken: `:315` pinned (visible in AC-GSF-001 grep), not byte-unchanged | dispatch record + AC-GSF-001 grep, this run |

Mutant probe (plan §E.1): `:99` pin reverted → pinned-form grep shows 2/3 sites (line 99 absent) → restored → count re-measured 3. Note: `grep -n` exits 0 at 2 matches, so the criterion's asserting form is the match count, not the exit code.

Gaps: AC-GSF-003 pending-external; the pinned-form grep was not run as a standalone command pre-repair (the `"--bare"` baseline grep establishes the same RED-now fact); no pre-repair gofmt baseline (formatting is not the defect).

## §E.3 Run-phase Audit-Ready Signal

```yaml
phase: run
spec: SPEC-GIT-STATUS-FIXTURE-001
card: t474
branch: WT-git-status-fixture
run_complete_at: "2026-09-03T19:16:06Z"
run_commit_sha: "93dad8a115e3ddf00075beb5477ad820368937b0"   # backfilled in the follow-up commit (self-referential SHA, D3 backfill window)
run_status: run-complete-local            # local ACs 001/002/004/005/006 PASS; AC-GSF-003 pending-external (merged-tree CI, lead-owned)
ac_pass_count: 5
ac_fail_count: 0
ac_pending_external: 1                   # AC-GSF-003 — card does not close before this verdict
preserve_list_post_run_count: 0          # no PRESERVE-surface files modified
new_warnings_or_lints_introduced: 0      # gofmt -l . empty; single-line insertions in one _test.go file
total_run_phase_files: 1                 # internal/core/git/status_branch_test.go
m1_to_mN_commit_strategy: "M1 kickoff-decision recorded (no commit of its own, per Tier S minimal form) → M2 single repair commit (code + spec frontmatter draft→in-progress + progress.md + run-evidence.md) → SHA backfill follow-up commit"
```


## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
