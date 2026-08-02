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

### M3 — guard falsifiability verification (acceptance.md §C.1 grid reproduced)

Each guard's removal was applied as a temporary mutation to `worktree.go`,
run against the targeted scenarios, captured, then reverted via
`git checkout HEAD --`. Every mutation flipped EXACTLY its §C.1 blast radius
and nothing else — the measured-disjoint property holds for this implementation.

| Mutation (criterion) | Flipped rows (observed) | Disjoint? |
|---|---|---|
| `folded-names` — drop `--no-renames`, keep `-z` (AC-WSM-006.1) | SC-10, SC-11 | yes — SC-12-15 PASS |
| `newline-split` — drop `-z`, split on `\n` (AC-WSM-006.2) | SC-12 | yes — SC-10/11/13-15 PASS |
| `no-literal` — drop `--literal-pathspecs` (AC-WSM-006.3) | SC-13 | yes — SC-10-12/14/15 PASS |
| `no-textconv` — drop `--no-textconv` (AC-WSM-006.4) | SC-14 | yes — SC-10-13/15 PASS |
| `no-ignoresub-cmp` — drop comparison-side `--ignore-submodules=none` (AC-WSM-006.5) | SC-15 | yes — SC-10-14 PASS |
| `no-ignoresub-enum` — drop enumeration-side `--ignore-submodules=none` | SC-15 | both stages load-bearing (each alone flips SC-15) |
| `no-state` — drop S5 conjunct from S3+S4 (AC-WSM-015) | SC-8, SC-10-15, P3c, P4c | wide — FORBIDDEN as sole falsification for AC-WSM-005/006/017 |
| `both` — drop S5 + accept any non-empty cherry (AC-WSM-005) | SC-6, SC-7 (+ no-state set) | the unique mutation that flips SC-6 AND SC-7 |
| `oid-only-cmp` — replace comparison with mode-blind `ls-tree` OID maps (AC-WSM-017) | P3c, P4c | yes — SC-6-15 PASS; `git diff` is mode-sensitive |
| unpin `GIT_COMMITTER_DATE` (AC-WSM-008) | determinism test FAIL (delta 0→2) | yes — pinned stays delta 1 |
| inject bare `git prune` (AC-WSM-009) | grep judge 0→1 | yes — restored to 0 |

Notable findings during verification:
- The determinism test was STRENGTHENED with a >1s gap between the two
  evaluations, because two rapid un-pinned calls land in the same wall-clock
  second and produce the same object (delta 1) — the test as first written did
  NOT falsify the unpin. With the gap, un-pinned → delta 2 (FAIL), pinned →
  delta 1 (PASS). This is the load-bearing fix that makes AC-WSM-008 falsifiable.
- The `oid-only-cmp` mutation helper initially used `ls-tree` WITHOUT `-z`, which
  C-quoted the non-ASCII path and incidentally also flipped SC-12 — a quoting
  artifact of the mutation, not a property of oid-only-cmp. Fixed by adding `-z`
  to the helper; SC-12 then correctly stays at keep. This is itself evidence of
  the encoding axis the production `-z` guard exists to close.

Full suite after M3 (determinism test now sleeps 1.1s, total ~79s):
```
$ go test ./internal/core/git/ -count=1   → ok (PASS)
$ go vet ./internal/core/git/             → clean
$ golangci-lint run ./internal/core/git/  → 0 issues
```
Permanent M3 code change: the determinism-test gap + `time` import (9 LOC). All
mutations were reverted; `worktree.go` matches the M2 baseline byte-for-byte.

### M4 — full-suite + interface-stability verification

```
$ go test ./... -count=1   → exit 0 (105 packages ok, 0 FAIL/panic)
$ go vet ./...             → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...  → exit 0
```

Interface-stability evidence (the premise of the shared-predicate decision,
spec.md §5 decision 3 — no test double, no mock, no call site needed editing):
```
$ git diff --exit-code 1dbf1f47c -- internal/core/git/types.go   → exit 0 (UNCHANGED)
$ git diff --stat 1dbf1f47c -- internal/cli/worktree/clean.go    → (empty, unmodified)
$ grep -cE 'Remove\([A-Za-z_.]+, false\)' internal/cli/worktree/clean.go   → 2 (both call sites)
```
The `--stale` and `--merged-only` CLI tests pass unchanged — they stub the
predicate (`mockIsBranchMergedFunc`) so they cannot exercise the real one, but
their continued green is the evidence that the interface the predicate satisfies
did not change shape:
```
$ go test ./internal/cli/worktree/ -run 'TestCleanStale_(KeepsDirtyWorktree|PreviewsByDefault|RemovesWithYes)' -count=1 -v
--- PASS: TestCleanStale_KeepsDirtyWorktree (0.00s)
--- PASS: TestCleanStale_PreviewsByDefault (0.00s)
--- PASS: TestCleanStale_RemovesWithYes (0.00s)
```
No code change in M4; this milestone is pure verification.

### M5 — mechanical follow-through

**Help text: NO CHANGE.** The `clean` command's `Long` text and the `--stale` /
`--merged-only` flag descriptions are stated generally — "worktrees whose
branches have been merged into the base branch" and "no commits of its own
beyond the base branch". Read naturally, squash-merged branches were always
WITHIN that wording; the pre-fix implementation under-delivered against it
(missed squash merges), and this SPEC makes the behaviour converge with the
wording rather than diverge. The plan.md §E M5 bar is "update only if the
observed behaviour no longer matches its wording" — it matches, so no edit.

**`@MX:ANCHOR` on the predicate: present** (added in M1, lines 231-232 of
`worktree.go`). It records the non-obvious invariant the plan §D anchoring note
names — why reachability (S1) is retained alongside the patch-id probes (S3/S4)
and conjoined with the state check (S5) — with a `@MX:REASON` naming the
specific scenarios each guard protects (SC-3/SC-4 for S1; SC-8..SC-15/P3c/P4c
for S5 via the `no-state` mutation). A future edit that drops S1 or S5 is the
break this anchor exists to flag.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-02      # M1-M5 complete
run_commit_sha: 4966e6720        # run-phase terminal commit (M5); backfilled post-push per D3 self-referential exemption
run_status: run-green            # all five milestones GREEN; 17/17 AC measured PASS
ac_pass_count: 17                # AC-WSM-001..017 all measured PASS (M1-M5)
ac_fail_count: 0
preserve_list_post_run_count: 0  # types.go unmodified; clean.go call sites untouched
l44_pre_commit_fetch: n/a        # 1-person OSS, branch-local
l44_post_push_fetch: pending     # backfilled after branch push
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues == baseline 0
cross_platform_build:
  go_build_all: exit_0
  windows_amd64: exit_0
total_run_phase_files: 3         # worktree.go, manager.go, worktree_squash_merge_test.go (clean.go help text unchanged, types.go unchanged)
m1_to_mN_commit_strategy: per-milestone Conventional-Commit + 🗿 MoAI trailer; branch push only (repo-local PR-mandatory, no main direct push)
run_phase_commits:               # M1..M5 commits on fix/clean-stale-squash-detect
  m1: 20858cfc6
  m2: b22584c7c
  m3: 50b711eb9
  m4: d107d40db
  m5: 4966e6720                  # run-phase terminal commit (M5)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: COMPLETE
sync_complete_at: 2026-08-03
sync_commit_sha: e17653bbf            # backfilled (was pending-backfill-sync placeholder in the sync commit PR #1297 — D3 self-referential-hazard exemption); sync squash-merge subject: docs(SPEC-WORKTREE-SQUASH-MERGE-001): sync-phase close (3-phase plan→run→sync)
run_commit_sha: 4966e6720                # run-phase terminal commit (M5); origin's squash-merge is d7506e9e4 (PR #1293)
tier: M
changelog_entry_position: "### Fixed 최상단 (CHANGELOG.md) — squash-merge detection is a false-negative bug fix, not a new feature"
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed (3-phase close merged into this single sync commit; manager-docs owns the transition per Status Transition Ownership Matrix)"
  plan_md: "n/a — markdown header convention, no YAML frontmatter"
  acceptance_md: "n/a — markdown header convention, no YAML frontmatter"
  progress_md: "n/a — markdown header convention, no YAML frontmatter (§E.4 yaml block is a fenced code block, not frontmatter)"
canary_compliance_check:
  b12_pre_emission_duplicate: "grep -c 'SPEC-WORKTREE-SQUASH-MERGE-001' CHANGELOG.md → 0 (사전) → 1 (커밋 후)"
  b12_ac_count_match: "CHANGELOG 엔트리는 user-facing 요약 — AC/REQ 카운트는 명시하지 않는다 (CHANGELOG ≠ spec audit; acceptance.md §A 17 AC SSOT 참조)"
  b12_file_paths_verified: "ls internal/core/git/worktree.go internal/core/git/manager.go internal/cli/worktree/clean.go → 3 files (M5 help-text no-change confirmed; types.go unchanged confirmed)"
docs_site_assessment: "no doc-body contradiction found — docs-site content/*/worktree/{guide,faq,examples}.md describe --merged-only generically ('remove only Worktrees whose branch is merged into base') with no claim about squash-merge detection either way; the fix makes behaviour converge to the existing wording. README has no worktree-clean content. No doc edit required."
```

### Claim (주장)

이 커밋은 SPEC-WORKTREE-SQUASH-MERGE-001의 **3단계 종료(sync-phase close)** 를 수행한다: `spec.md` 의 `status:` 를 `in-progress → implemented → completed` 로 전이시키고 (Status Transition Ownership Matrix가 규정한 대로 단일 sync commit에 3-phase close가 병합 — 별도 Mx chore commit 없음), 본 `progress.md §E.4` 를 채우며, `CHANGELOG.md [Unreleased] → ### Fixed` 최상단에 엔트리를 추가한다. **코드 변경은 없다** — run-phase 코드는 PR #1293 (squash merge `d7506e9e4`) 로 이미 `origin/main` 에 착지했다. 이 커밋은 산출물 전이 + §E.4 시그널 + CHANGELOG 엔트리만 운반한다.

### Evidence (증거)

동기화 커밋이 존재하기 전에는 자기 SHA를 알 수 없으므로 (`git commit` 이 SHA를 할당하는 시점이 커밋 생성 이후), `sync_commit_sha` 필드는 `pending-backfill-sync` 자리표시자로 두고 후속 backfill 커밋에서 채운다 (`9ec8d8464` "docs(SPEC-HOOK-TRACE-FLUSH-001): backfill sync_commit_sha 9d976c95b" 및 직전 `cdb692135` "chore(SPEC-PROFILE-MEMORY-001): backfill sync_commit_sha 24d7d3eb1" 의 정준 패턴 — D3 self-referential-hazard exemption).

frontmatter 전이 (Edit 전/후):
- `status: in-progress` → `status: completed` (spec.md 유일한 YAML-frontmatter 산출물 — Tier M: spec/plan/acceptance/progress 중 spec.md만 frontmatter 보유)
- `updated: 2026-08-02` → `updated: 2026-08-03` (오늘 날짜; 본 sync commit이 spec.md에 마지막으로 손대는 시점)
- 그 외 frontmatter 필드 + 본문(§HISTORY 포함) 전체 = 불변 (Forbidden ownership crossings 준수)

CHANGELOG B12 사전 검증 (커밋 전):
```
$ grep -c 'SPEC-WORKTREE-SQUASH-MERGE-001' CHANGELOG.md
0   # 중복 없음 — parallel BATCH-SYNC 세션 충돌 없음
```

CHANGELOG 삽입 위치: `### Fixed` 최상단 (squat-merge detection은 false-negative 버그 수정 — "Added" 신규 기능이 아님). 기존 `SPEC-OBS-ENABLED-GATE-001` 엔트리 위에 삽입.

### Baseline-attribution (baseline 귀속)

- **run-phase baseline**: PR #1293 squash-merge `d7506e9e4` (`feat(SPEC-WORKTREE-SQUASH-MERGE-001): squash-merge detection for worktree --stale cleanup (#1293)`). 17/17 AC measured PASS (M1-M5), coverage 85.7%, cross-platform build exit 0, golangci-lint 0 issues — `progress.md §E.2 / §E.3` 에 verbatim 인용.
- **sync-phase baseline (본 커밋)**: `origin/main d7506e9e4` 위에서 `chore/SPEC-WORKTREE-SQUASH-MERGE-001-sync` 브랜치 생성 → CHANGELOG + spec.md frontmatter + progress.md §E.4 편집 → 단일 sync commit. 측정은 커밋 후 backfill PR에서 수행.
- **MX tag cross-cutting**: M1에서 `internal/core/git/worktree.go` `IsBranchMerged` 에 `@MX:ANCHOR` + `@MX:REASON` 추가됨 (lines 231-232, §E.2 M5 evidence 인용). 별도 Mx commit 없음 — 3-phase close 계약에 따라 sync sub-step으로 validate 완료.

### Gaps (미검증)

- **sync_commit_sha 본 커밋 미검증**: 위에서 명시한 대로 자기-참조 위험(self-referential hazard) — 후속 backfill PR에서 실제 merge SHA로 교체할 때까지 `pending-backfill-sync` 로 둔다.
- **CHANGELOG build/lint verification 본 커밋 미검증**: markdown 문서 편집이므로 `go build` / `go test` / `golangci-lint` 는 sync PR의 CI에서 검증 (run-phase에서 이미 exit 0 확인됨 — 코드 변경 없으므로 regression 없음).
- **docs-site build 본 커밋 미검증**: docs-site 편집이 없으므로 hugo build 영향 없음 — 검증 불필요 (docs_site_assessment 참조).

### Residual-risk (잔여 위험)

- **backfill PR 누락 위험**: `pending-backfill-sync` 가 후속 커밋으로 교체되지 않으면 era classifier (H-4 predicate) 가 `sync_commit_sha` empty 값을 보고 V3R5로 misclassify할 수 있다. 완화: 본 커밋 merge 후 즉시 backfill PR 생성 (task step 7).
- **auto-merge bot 지연**: repo의 auto-merge bot이 CI green 후에도 merge를 지연시킬 수 있다 — 수동 self-merge fallback 준비.

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
