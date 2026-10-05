# SPEC-HARNESS-EVIDENCE-WRITE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec_id: SPEC-HARNESS-EVIDENCE-WRITE-001
phase: plan
plan_status: audit-ready
revision: "1.0.1"  # delta per plan-audit review-1 (D1-D3 fixed: REQ-007 pre-run override + no-clobber; AC env pins + -count=1; AC-008 added)
tier: S
artifacts: [spec.md, plan.md, progress.md]
card: t569
grounding: lane-10 premise re-verification @ 3ac58b5a1 (2026-09-08)
req_count: 7
ac_count: 8
open_blockers: none  # delta re-audit PASSED (review-2, 1.0) — D1-D3 closed on substance
```

## §F Phase 4 Mode Selection

Decision: serial — one manager-develop spawn, sequential milestones.

| Mode | Verdict | Rationale |
|------|---------|-----------|
| direct | not selected | multi-file implementation (4 test files + guard), not a typo-level change |
| serial | SELECTED | coding-heavy single-package work — Anthropic coding-task parallelism caveat |
| fanout | not selected | single domain (internal/spec), no research fan-out |
| sweep | not selected | <30 files, semantic harness change, not mechanical-uniform |

Kickoff note: Implementation Kickoff Approval satisfied by the operator's factory dispatch (lane-10, card t569) — the lead's dispatch carries the operator-approved chain plan → run → sync for this card; per-card pre-authorization in Factory Mode, not a gate bypass. plan-audit record: iter1 FAIL 0.875 (review-1, D1-D3 blocking) → delta revision v1.0.1 → iter2 PASS 1.0 (review-2).

## §E.2 Run-phase Evidence

Run phase executed 2026-09-08 in worktree `.claude/worktrees/t569`, branch `WT-harness-evidence-write`, base HEAD `fccc31d7e` (plan-phase commit). Raw evidence files under `/tmp/t569/` (verbatim outputs quoted below).

### E.2.1 TDD cycle — RED / GREEN / MUTANT (E8, verbatim)

**RED — guard vs CURRENT tree (before any fix), exit 1** (`/tmp/t569/red-guard-vs-current-tree.txt`; command: `unset MOAI_T362_CORPUS_SCAN T528_PROBE_OUT && go test ./internal/spec/ -run TestNoTestWritesRepoTree -count=1 -v`):

```
=== RUN   TestNoTestWritesRepoTree
    guard_no_repo_tree_write_test.go:67: lint_req_widen_corpus_test.go:240: os.MkdirAll writes via "out" (repo-anchored)
    guard_no_repo_tree_write_test.go:67: lint_req_widen_corpus_test.go:243: os.WriteFile writes via "out" (repo-anchored)
    guard_no_repo_tree_write_test.go:67: lint_req_widen_decompose_test.go:542: os.MkdirAll writes via "out" (repo-anchored)
    guard_no_repo_tree_write_test.go:67: lint_req_widen_decompose_test.go:545: os.WriteFile writes via "out" (repo-anchored)
    guard_no_repo_tree_write_test.go:67: zz_t528_anchor_probe_test.go:61: os.MkdirAll writes via "dir" (repo-anchored)
    guard_no_repo_tree_write_test.go:67: zz_t528_anchor_probe_test.go:73: os.WriteFile writes via "dir" (repo-anchored)
--- FAIL: TestNoTestWritesRepoTree (0.02s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/spec	0.429s
FAIL
```

**GREEN — guard after fix, exit 0** (`/tmp/t569/green-guard-after-fix.txt`; same command):

```
=== RUN   TestNoTestWritesRepoTree
--- PASS: TestNoTestWritesRepoTree (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/spec	0.376s
```

**MUTANT RED — AC-006 (scratch test with `os.WriteFile(filepath.Join(root, ".moai/reports/t569-mutant-probe.txt"), ...)` re-added; mutant test never executed, only scanned), exit 1** (`/tmp/t569/red-guard-mutant.txt`):

```
=== RUN   TestNoTestWritesRepoTree
    guard_no_repo_tree_write_test.go:67: zz_t569_mutant_scratch_test.go:15: os.WriteFile writes a repo-anchored path (direct .moai/ literal)
--- FAIL: TestNoTestWritesRepoTree (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/spec	0.385s
FAIL
```

**MUTANT REVERTED — guard GREEN again, exit 0** (`/tmp/t569/green-guard-mutant-reverted.txt`): `--- PASS: TestNoTestWritesRepoTree (0.01s)` / `ok ... 0.213s`.

### E.2.2 AC binary matrix (E1)

| AC | Status | Verification command | Actual output |
|----|--------|---------------------|---------------|
| AC-001 | PASS | `unset MOAI_T362_CORPUS_SCAN T528_PROBE_OUT && MOAI_T362_CORPUS_SCAN=1 go test ./internal/spec/ -run TestCorpusREQWideningMeasurement -count=1 -v` | exit 0; `lint_req_widen_corpus_test.go:251: measurement written to /var/folders/.../TestCorpusREQWideningMeasurement386710612/001/m1-corpus-measurement.txt`; `--- PASS`; `git status --porcelain .moai/reports/t362/` printed nothing |
| AC-002 | PASS | same form, `-run TestCorpusRejectedREQIDDecomposition` | exit 0; `.../TestCorpusRejectedREQIDDecomposition4050208730/001/m2-gate0-decomposition.txt`; `--- PASS`; t362 tree clean |
| AC-003 | PASS | `unset ... && go test ./internal/spec/ -run TestT528Anchor -count=1 -v` | exit 0; `OUTDIR = /var/folders/.../TestT528Anchor2984576082/001`; `--- PASS`; `git status --porcelain .moai/reports/t528/` printed nothing |
| AC-004 | PASS | `T528_PROBE_OUT=/tmp/t569-probe-out-override go test ./internal/spec/ -run TestT528Anchor -count=1 -v` | exit 0; `OUTDIR = /tmp/t569-probe-out-override`; all 8 probe files present in that dir |
| AC-005 | PASS | guard on fixed tree (E.2.1 GREEN) | exit 0, PASS |
| AC-006 | PASS | mutant probe (E.2.1 MUTANT RED/REVERTED) | exit 1 on mutant naming file+line; exit 0 after revert |
| AC-007 | PASS | `git status --porcelain internal/spec/ .moai/specs/.../ && git diff --stat .moai/reports/` | `.moai/reports/` diff EMPTY; `ac_count_clause_test.go` / `zz_t528_overacceptance_test.go` absent from modified list (unmodified); before-image `git ls-tree HEAD .moai/reports/t362/` unchanged (12 tracked files) |
| AC-008a | PASS | `MOAI_T362_CORPUS_SCAN=1 MOAI_T362_EVIDENCE_OUT=/tmp/t569-evidence-out go test ... -run TestCorpusREQWideningMeasurement -count=1 -v` | exit 0; report landed at `/tmp/t569-evidence-out/m1-corpus-measurement.txt`; `--- PASS` |
| AC-008b | PASS | identical invocation re-run against same dir | exit 1; `MOAI_T362_EVIDENCE_OUT no-clobber: target ... already exists; refusing to overwrite` (test file line 246); sha256 `234dc16585962e12...f9a93` BEFORE=AFTER (`SHASUM_IDENTICAL`) |
| AC-008c | PASS | `head -6` of both produced reports | both headers carry `# output location: a per-run t.TempDir() directory` + `# durable capture: set MOAI_T362_EVIDENCE_OUT=<dir> BEFORE the run` lines; zero `.moai/reports` mentions (grep count 0) |

Gate-preservation (non-AC check): ungated pair run → 2 SKIPs, exit 0 (`/tmp/t569/gate-preserved-skip.txt`). Guard boundary documented in the guard file's doc comment (dynamically built paths, parameter-carried anchors, multi-line statements, self-scan exclusion, non-test sources).

### E.2.4 Post-verification polish (lane LSP-diagnostic follow-up, 2026-09-08)

1. **Adopted — scanner error check** (`guard_no_repo_tree_write_test.go`): the `bufio.Scanner` loop now surfaces a mid-file read error as a finding (post-loop `sc.Err()` → finding → `t.Errorf`), closing a truncated-file vacuous-green path in the guard itself. Post-fix guard run: exit 0, `--- PASS: TestNoTestWritesRepoTree (0.01s)` / `ok ... 0.373s` (`/tmp/t569/green-guard-after-polish.txt`).
2. **Rejected with evidence — "t362ReportPath unused"** (`t362_evidence_out_test.go:27`): the diagnostic's premise is false on the committed tip. Verified: `grep -n t362ReportPath` shows two LIVE call sites (`lint_req_widen_corpus_test.go:246`, `lint_req_widen_decompose_test.go:547`), the helper's no-clobber branch was observed executing (AC-008b's `Fatalf` message is emitted by this helper), and `golangci-lint run ./internal/spec/...` on the tip reports only the pre-existing `zz_t528_overacceptance_test.go:100` errcheck — no `unused` finding. Deleting a live, exercised helper would have broken the build; the diagnostic is treated as a stale LSP read against the pre-edit buffer. Nothing changed in `t362_evidence_out_test.go`.

### E.2.3 Self-verification E2-E6

- **E2 build**: `go build ./...` → `BUILD_NATIVE_OK`; `GOOS=windows GOARCH=amd64 go build ./...` → `BUILD_WINDOWS_OK` (both exit 0).
- **E3 coverage**: `go test -cover ./internal/spec/` → `coverage: 90.5% of statements` (target 85% — met; test-only change, coverage delta from added guard/helper code is in test files which do not count toward statement coverage).
- **E4 boundary grep**: N/A — test-only change inside `internal/spec`; no non-test source modified (verified via git status above).
- **E5 lint**: `golangci-lint run ./internal/spec/...` → 1 issue remaining: `zz_t528_overacceptance_test.go:100 errcheck` (PRE-EXISTING baseline, file is on the PRESERVE list and untouched). NEW issues: 0 (the one introduced during the run — guard file errcheck — was fixed and re-verified to 0).
- **E6 commits/push**: see §E.3; NO push performed (git-flow lane protocol §4 — lead batch-pushes develop).

## §E.3 Run-phase Audit-Ready Signal

```yaml
spec_id: SPEC-HARNESS-EVIDENCE-WRITE-001
card: t569
phase: run
run_status: complete
run_complete_at: 2026-09-08
run_commit_sha: ada562362  # D3 backfilled — run-phase tip (the polish commit) immediately before the sync commit
ac_pass_count: 10
ac_fail_count: 0
ac_matrix: AC-001..008c all PASS (see §E.2.2)
preserve_list_post_run_count: 4  # ac_count_clause_test.go, zz_t528_overacceptance_test.go, findRepoRoot helper, .moai/reports/t362/ + t528/ pinned evidence — all verified untouched
new_warnings_or_lints_introduced: 0
l44_pre_commit_fetch: not-applicable  # worktree-isolated card branch; no shared-checkout commit
l44_post_push_fetch: not-applicable  # NO push performed — lane protocol §4, lead batch-pushes develop
cross_platform_build.native: PASS
cross_platform_build.windows_amd64: PASS
total_run_phase_files: 5  # 3 modified harness tests + 2 new test files (+ spec.md frontmatter, progress.md)
m1_to_mN_commit_strategy: per-milestone commits on WT-harness-evidence-write (M1+M2 fix, M3 guard test, M4 evidence)
tdd_red_evidence: /tmp/t569/red-guard-vs-current-tree.txt (verbatim in §E.2.1)
mutant_evidence: /tmp/t569/red-guard-mutant.txt + /tmp/t569/green-guard-mutant-reverted.txt (verbatim in §E.2.1)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
spec_id: SPEC-HARNESS-EVIDENCE-WRITE-001
card: t569
phase: sync
sync_status: complete
sync_complete_at: 2026-09-08
sync_commit_sha: 3db94543e  # D3 backfilled — written as pending-backfill-sync inside the sync commit itself (a commit cannot cite its own SHA), resolved here
sync_branch: WT-harness-evidence-write
frontmatter_status_transitions:
  implemented_to_completed: carried-by-sync-commit  # status + updated only, spec.md frontmatter
changelog_entry: added  # [Unreleased] §Fixed — developer-visible harness behavior change (t.TempDir() defaults + MOAI_T362_EVIDENCE_OUT override); entry-worthiness judged against the repo's existing internal-behavior entries
b12_self_test_a: changelog_grep_count_0  # pre-emission `grep -c SPEC-HARNESS-EVIDENCE-WRITE-001 CHANGELOG.md` = 0 — no duplicate emission
b12_self_test_b: ac_count_match_10_vs_10  # spec.md §3 (Tier S AC SSOT) distinct AC identifiers = 10 (AC-001..007, AC-008a/b/c); CHANGELOG entry references the same 10; progress.md's count not used (it counts deferred ACs)
b12_self_test_c: all_claimed_paths_exist  # 5 implementation files + .moai/specs/SPEC-HARNESS-EVIDENCE-WRITE-001/spec.md verified via ls before entry authoring
canary_compliance_check:
  single_sync_commit: true  # transition + §E.4 + CHANGELOG in ONE commit, no separate Mx chore
  no_push: true  # lane protocol §4 — lead batch-pushes develop
  docs_scope: none  # test-only change — README / docs-site / .moai/docs/ untouched (no user-facing doc surface)
  codemaps: not-needed  # no non-test source changed — no codemap regeneration
mx_tag_delta: 0  # test-only change; no new exported symbols; scan of the 5 changed test files found no @MX tags added or removed
ac_summary: 10 PASS / 0 FAIL (§E.2.2 matrix)
read_only_consumers_untouched: true  # ac_count_clause_test.go, zz_t528_overacceptance_test.go unmodified (AC-007)
pinned_evidence_diff: zero  # .moai/reports/t362/ + t528/probe/ byte-unchanged
```

