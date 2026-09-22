# SPEC-WORKTREE-DONE-TIER-001 — Progress

Card: t1073 | Branch: WT-done-tier-claim | Base: develop 0314801c2 | Tier: M

## Status

- Plan phase: COMPLETE (2026-09-22) — spec.md / plan.md / acceptance.md authored.
- Run phase: COMPLETE (2026-09-22) — M1-M4 landed as commits 2cde18a7b / 1a5ff7780 / 7795226c2 (+ M4 progress record); AC-001..AC-010 green with verbatim evidence in §E.2; see §E.7-level divergence notes for the inherited mirror-parity red.
- RED-now pinning (M1) executed as the FIRST run-phase action, before any guard code landed (REQ-006) — verbatim record below.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-22
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
spec_id_check: "Bash regex PASS (verbatim in plan-phase transcript)"
baseline_tree: 0314801c2
```

## §E.2 Run-phase Evidence

### Run provenance

- Branch `WT-done-tier-claim`, entry HEAD `e0f910705` (absorbed develop `fcb193626` + handoff residue `9a35f217d`), worktree `.claude/worktrees/t1073`. Card base re-derived at verification: `git merge-base develop HEAD` → `fcb193626`.
- Commits (per-milestone, direct on the WT branch — NOT pushed; lead batch-pushes develop):
  - M1 `2cde18a7b` — test(SPEC-WORKTREE-DONE-TIER-001): RED-now reproduction tests + spec.md frontmatter flip draft→in-progress
  - M2 `1a5ff7780` — feat(SPEC-WORKTREE-DONE-TIER-001): L1 tier guard (predicate + target-derived resolver + shared refusal site, wired into both `runDone` and `runDoneWorktreeCleanup` before the anchor guard and before any `Remove` call; `--force` never reaches the refusal as a parameter)
  - M3 `7795226c2` — docs(SPEC-WORKTREE-DONE-TIER-001): doctrine precision edit, both copies
  - M4 this progress record.
- Plan-provenance interpretation (on record per dispatch): plan.md §D said "Do not commit (orchestrator commits after audit)". The audit (iter-2 PASS 0.9375) has passed, so that sequencing condition is satisfied; per this lane's standing gitflow pattern the run commits per-milestone directly on `WT-done-tier-claim`. Recorded as an intentional divergence from a literal reading of that plan line.
- Final verification HEAD: `7795226c2` (all §E items below measured at this tree).

### E8 — RED-now evidence (pre-fix tree pinned)

- Pre-fix tree SHA: `2cde18a7b` (M1 commit; guard code not yet present).
- Command: `go test ./internal/cli/worktree/... -run 'TestDoneL1TierGuard' -v`
- Exit code: **1**
- Verbatim output:

```text
=== RUN   TestDoneL1TierGuard_RefusesCleanL1
    done_l1_tier_guard_test.go:144: done must refuse an L1 session worktree (nil error = removal proceeded)
--- FAIL: TestDoneL1TierGuard_RefusesCleanL1 (0.53s)
=== RUN   TestDoneL1TierGuard_RefusesDirtyL1ByTier
    done_l1_tier_guard_test.go:161: refusal must carry the L1_SESSION_WORKTREE sentinel, got: remove worktree: remove worktree at "/private/var/folders/.../main/.claude/worktrees/tier-a2": git: worktree has uncommitted changes
--- FAIL: TestDoneL1TierGuard_RefusesDirtyL1ByTier (0.55s)
=== RUN   TestDoneL1TierGuard_ForceDoesNotBypassL1
    done_l1_tier_guard_test.go:179: done must refuse an L1 session worktree (nil error = removal proceeded)
--- FAIL: TestDoneL1TierGuard_ForceDoesNotBypassL1 (0.51s)
=== RUN   TestDoneL1TierGuard_RefusesIgnoredOnlyDirtyL1
    done_l1_tier_guard_test.go:214: done must refuse an L1 session worktree (nil error = removal proceeded)
--- FAIL: TestDoneL1TierGuard_RefusesIgnoredOnlyDirtyL1 (0.55s)
=== RUN   TestDoneL1TierGuard_DeleteBranchRefusedPreRemoval
    done_l1_tier_guard_test.go:236: done must refuse an L1 session worktree (nil error = removal proceeded)
--- FAIL: TestDoneL1TierGuard_DeleteBranchRefusedPreRemoval (0.68s)
=== RUN   TestDoneL1TierGuard_L2FlowUnchanged
--- PASS: TestDoneL1TierGuard_L2FlowUnchanged (0.62s)
=== RUN   TestDoneL1TierGuard_RefusalMessageBothModes
=== RUN   TestDoneL1TierGuard_RefusalMessageBothModes/interactive
    done_l1_tier_guard_test.go:286: both modes must exit non-zero on an L1 target
=== RUN   TestDoneL1TierGuard_RefusalMessageBothModes/auto
    done_l1_tier_guard_test.go:286: both modes must exit non-zero on an L1 target
--- FAIL: TestDoneL1TierGuard_RefusalMessageBothModes (1.08s)
    --- FAIL: TestDoneL1TierGuard_RefusalMessageBothModes/interactive (0.53s)
    --- FAIL: TestDoneL1TierGuard_RefusalMessageBothModes/auto (0.55s)
=== RUN   TestDoneL1TierGuard_CWDIndependent
    done_l1_tier_guard_test.go:331: done must refuse an L1 session worktree (nil error = removal proceeded)
--- FAIL: TestDoneL1TierGuard_CWDIndependent (0.67s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	5.585s
FAIL
```

Reading: 8 refusal cells RED for the right stated reason (no guard — silent removal or git-dirty provenance); AC-006 (B-direction) PASS as its RED cell documents ("deployed legit behavior — must stay PASS"); the path-truncation `...` above is this record's abbreviation of one `/private/var/folders` temp path, not the runner's. RED observed BEFORE any guard code was written (M2 landed after).

### E1 — AC matrix (all cells green at HEAD 7795226c2)

Verifying command for all rows: `go test ./internal/cli/worktree/... -run 'TestDoneL1TierGuard' -v` → exit 0, `ok github.com/modu-ai/moai-adk/internal/cli/worktree`.

| AC | Scenario | Status | Evidence (test) |
|----|----------|--------|-----------------|
| AC-001 | A1 clean L1 refused | PASS | `--- PASS: TestDoneL1TierGuard_RefusesCleanL1` |
| AC-002 | A2 dirty L1 refused BY TIER | PASS | `--- PASS: TestDoneL1TierGuard_RefusesDirtyL1ByTier` |
| AC-003 | A3 `--force` does not bypass | PASS | `--- PASS: TestDoneL1TierGuard_ForceDoesNotBypassL1` (uncommitted content intact asserted in-test) |
| AC-004 | C ignored-only dirtiness kept | PASS | `--- PASS: TestDoneL1TierGuard_RefusesIgnoredOnlyDirtyL1` |
| AC-005 | A4 `--delete-branch` refused pre-removal | PASS | `--- PASS: TestDoneL1TierGuard_DeleteBranchRefusedPreRemoval` (branch survival asserted in-test) |
| AC-006 | B L2 flow unchanged (both flags) | PASS | `--- PASS: TestDoneL1TierGuard_L2FlowUnchanged` (`done SPEC-TIER-B --auto --delete-branch` removes tree + merged branch, exit 0) |
| AC-007 | Refusal message + exit codes, both modes | PASS | `--- PASS: TestDoneL1TierGuard_RefusalMessageBothModes` (interactive + auto subtests; message carries path, `L1_SESSION_WORKTREE`, `session worktree`, `session-scoped`, `session-end`, `git worktree unlock <path>`, `git worktree remove <path>`) |
| AC-008 | Doctrine precision edit | PASS-WITH-DEBT | Sentence updated in both copies, byte-identical wording: `sed -n 52p` of both files piped to `cmp` → `LINE52_BYTE_IDENTICAL` (exit 0). DEBT: whole-file `cmp` cannot be byte-identical — pre-existing divergence (below), not introduced by this SPEC. `kanban-dispatch.md` "closes L2 trees only" claims verified factually true post-guard — no edit. |
| AC-009 | Scope confinement | PASS | `git diff --name-only fcb193626..HEAD -- internal/ .claude/rules/... worktree-integration.md mirror` → exactly: `internal/cli/worktree/done.go`, `internal/cli/worktree/done_l1_tier_guard_test.go`, `.claude/rules/moai/workflow/worktree-integration.md`, `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` |
| AC-010 | CWD-independence (D1/D6 mutation) | PASS | `--- PASS: TestDoneL1TierGuard_CWDIndependent` — process CWD inside a second linked worktree, DEFAULT target-derived resolver (no seam override anywhere in the file), real git commands exercised |

### E2 — Regression / vet / lint (baseline vs NEW)

- `go test ./internal/cli/worktree/...` → `ok ... (cached)` after an uncached first run; exit 0. No regression.
- `go vet ./internal/cli/...` → silent, exit 0 (`VET_OK`).
- `golangci-lint run internal/cli/worktree/...` → baseline (pre-change): `0 issues.`; post-change: `0 issues.` — **NEW issues: 0**.
- Coverage: `go test ./internal/cli/worktree/... -cover` → `coverage: 87.2% of statements` (project gate 85%).

### E3 — Cross-platform

- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (`WINDOWS_OK`), measured at final HEAD.
- `GOOS=darwin GOARCH=amd64 go build ./...` → exit 0 (`DARWIN_OK`).
- Predicate is `filepath`-based (`filepath.Dir`, `filepath.EvalSymlinks`, `filepath.Join`, `os.PathSeparator`) — no hardcoded `/`.

### E4 — Doctrine parity

- Edited sentence (line 52) byte-identical across both copies: `cmp` of `sed -n 52p` outputs → exit 0 (`LINE52_BYTE_IDENTICAL`).
- Whole-file `diff` post-edit shows ONLY the pre-existing divergence blocks (lines 610-612, 653-665): the local copy carries card t1067 internal-state records the template mirror must not mirror (template content-neutrality boundary, CLAUDE.local.md §2.1). That divergence is present at the develop base: `git show fcb193626:<local>` vs `git show fcb193626:<mirror>` → `differ: char 54628, line 610` (exit 1).
- Consequence: `TestRuleTemplateMirrorDrift/worktree-integration.md` is RED — **inherited at the develop base, not introduced** (verified: the failure exists with identical coordinates at `fcb193626` before any of this SPEC's edits). Uncached `go test ./internal/template/ -count=1` → exactly ONE failing subtest, this one; no other failures. Fixing it would require either mirroring internal development state into the distributed template (neutrality violation) or stripping local dogfood records (out of scope) — left to the owning card/orchestrator.

### E7 — Divergences and refusals on record

1. AC-008 whole-file byte-identity was unachievable before this SPEC began (pre-existing template-neutrality divergence, see E4). REQ-007's actual obligation — byte-identical EDIT WORDING — is met and proven at line-52 granularity.
2. Worktree-session guard refusals during M3 (recorded per VCI §3.1; both resolved by falling back to the Edit tool, no measurement lost): (a) a compound command with `<(...)` process substitution; (b) a `python3` heredoc whose payload named git paths. No command was executed in refused form.
3. `make build` / `make agents-emit` deliberately NOT run (plan §D): the edited files are markdown rules, not agent definitions; `catalog.yaml` carries no `worktree-integration` entry (grep 0 rows) so no catalog-hash cascade applies; embed correctness for Go code is proven by the build + template-package test run above.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-22
run_commit_sha: "7795226c2"
run_status: complete
ac_pass_count: 9
ac_fail_count: 0
ac_pass_with_debt_count: 1   # AC-008 — whole-file mirror parity debt, inherited at develop base
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "n/a (git-flow lane — merge-base develop HEAD re-derived per §8 gitflow-lane-protocol)"
l44_post_push_fetch: "n/a — NOT pushed (lead batch-pushes develop; operator directive 2026-09-01)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: pass
  windows: pass
total_run_phase_files: 4   # done.go, done_l1_tier_guard_test.go, worktree-integration.md (local), worktree-integration.md (mirror)
m1_to_mN_commit_strategy: per-milestone commits on WT-done-tier-claim (M1 2cde18a7b test-first RED, M2 1a5ff7780 guard, M3 7795226c2 doctrine, M4 this record); no push
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_close_at: 2026-09-22
sync_commit_sha: "b33c5788f"   # backfilled (D3 window) — the sync commit that carried the 3-phase close
sync_status: complete
tier: M
ac_pass_count: 9
ac_fail_count: 0
ac_pass_with_debt_count: 1   # AC-008 — whole-file mirror parity debt, inherited at the develop base (see §E.2 E1 + §E.2 E4)
```

## §F Phase 4 Mode Selection

```yaml
mode_selection_at: 2026-09-22
selected_by: lane orchestrator (agent-27, re-dispatch from lead; original lane agent-28 on t1084)
kickoff: Implementation Kickoff Approval PASSED — operator "전부 승인" (approve all), 2026-09-22
plan_audit_skip: TAKEN — all three conditions hold per spec-workflow.md § Phase 1 skip policy
  1. verdict: PASS (iter-2/2, score 0.9375)
  2. score >= Tier M threshold: 0.9375 >= 0.80
  3. artifact-hash unchanged since the audit (plan artifacts as committed at 99629d498; progress.md is outside the hash subject set)
inputs:
  tier: M
  scope_files: 3 (done.go guard site + done_l1_tier_guard_test.go + worktree-integration.md local/template pair)
  domain_count: 1 (internal/cli/worktree + its doctrine pair)
  file_language_mix: Go + markdown
  concurrency_benefit: LOW (coding-heavy, single package)
mode_evaluation:
  direct: not selected — semantic guard implementation, not a typo fix
  serial: SELECTED — one manager-develop spawn carries M1-M4 in order with per-milestone commits
  fanout: not selected — coding-heavy single-package work (Anthropic coding-task parallelism caveat)
  sweep: not selected — semantic work, tiny scope (3 files), far below the ~30-file mechanical threshold
decision: serial
justification: >
  Single-package guard implementation with tight sequencing (M1 RED tests must land before M2
  code; M2f mutation test depends on the shipped default resolver) — sequential single-spawn is
  both the Anthropic coding-task default and the only shape that preserves the M1->M2 ordering
  the two-cell discipline requires. No independent fan-out surface exists in this scope.
```

