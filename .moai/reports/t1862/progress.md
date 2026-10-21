# t1862 — progress record (Class B defect, cause known)

Tree: `.moai/worktrees/t1862`, branch `WT-factory-cardroot-align` (renamed in place from the
lease-time `WT-10-18-t1827`), base `2178966bb` (== local main at lease).
Scope: internal/cli (`factory_handoff_recover.go`) + the t1827 SPEC-wording reversion after
t1827 lands. Implementation is HELD for t1827 (lane-29, `WT-db-paths-where`) landing per the
leader dispatch of 10-18; this record currently carries cause evidence + RED only.

## 1. Cause evidence (mechanism + code sites)

**Mechanism.** Both verbs open the factory at the raw working directory:

- `internal/cli/factory_handoff_recover.go:29-36` (`factory handoff recover-resume`):
  `root, err := os.Getwd()` → `homestate.OpenFactory(root)`, and the same raw `root` feeds
  `homestate.ReadRuntimeCensus(root)`.
- `internal/cli/factory_handoff_recover.go:86-95` (`factory runs`): the same `os.Getwd()` root
  feeds `homestate.OpenFactory(root)` and `factorymsg.ReconcileOptionsFor(root)`.

The two rules every other factory surface applies:

- moai-paths root: `resolveProjectDir()` (internal/cli/session.go:315-323) —
  `CLAUDE_PROJECT_DIR` first (`env:CLAUDE_PROJECT_DIR`), then the cwd (`server-cwd`).
- factory-root rule (SPEC-FACTORY-LEASE-RENEW-001, sync-audit F2, card t1663;
  internal/cli/factory_card.go:86-92): "every surface that opens the factory for a launch tree
  applies this rule" — `factoryRootOf(launch) = factory.ResolveTodoQueueRoot(launch)`, which
  maps a linked worktree to the primary checkout.
- `factoryCardRoot()` (internal/cli/factory_card.go:84) composes both; `factory next`, `stage`,
  `renew`, `complete`, `decide`, `bundle` all route through it.

Neither verb applies either rule, so a cwd naming a different repository than moai paths (a
different `CLAUDE_PROJECT_DIR`, or a linked worktree) reads that other repository's factory
DB. `OpenFactory` additionally "reads the lease keys from the root the factory is opened at"
(internal/homestate/factory.go:171-190), so lease state is judged on the same wrong root.

t1827's unmerged branch (`WT-db-paths-where`) documents the divergence in `moai paths` help
("moai factory runs and moai factory handoff recover-resume fix the root from the working
directory instead — SPEC-MOAI-DB-PATHS-001") and narrowed that SPEC's wording to declare the
divergence pre-existing; this card reverts that narrowing once t1827 lands (the divergence is
then gone). lane-29 reproduced the two-repository read during the t1827 sync audit.

## 2. RED — verbatim, before the fix

Reproduction: `internal/cli/factory_root_alignment_test.go` (new file, this commit). Boundary
as the card names it: cwd names project B while `CLAUDE_PROJECT_DIR` names project A; both
verbs must read A's factory, never B's.

Command: `go test -run 'TestFactoryRunsOpensMoaiPathsRootNotCwd|TestRecoverResumeOpensMoaiPathsRootNotCwd' -count=1 -timeout 8m ./internal/cli/`

```
--- FAIL: TestFactoryRunsOpensMoaiPathsRootNotCwd (0.58s)
    factory_root_alignment_test.go:56: factory runs read the cwd factory, not the moai-paths root (card t1862): output "RUN ID  STATUS  OWNER  PID  CREATED  UPDATED\n"
Error: legacy resume recovery CAS precondition: sql: no rows in result set
--- FAIL: TestRecoverResumeOpensMoaiPathsRootNotCwd (0.48s)
    factory_root_alignment_test.go:85: recover-resume read the cwd factory, not the moai-paths root (card t1862): legacy resume recovery CAS precondition: sql: no rows in result set
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.408s
```

Failure modes are exactly the defect: `runs` lists the cwd (B) factory — empty, header only —
and `recover-resume` finds no row because B's factory never held A's claim.

## 3. GREEN (t1827 absorbed; fix + SPEC-wording reversion)

**Absorb.** t1827 landed on local main `026f48658f` (merge of `WT-db-paths-where`;
`f403f5224` reachable — measured `git merge-base --is-ancestor` exit 0, 06:32Z). Absorbed:
`git merge main --no-edit` → merge commit `26ea85135f`, no conflicts. Lease had expired
mid-wait; recovered through the standing 8859 sequence (`factory stage` expiry handler →
`factory next --card t1862` re-lease → `factory stage t1862 run` v7 with RED evidence).

**Fix** (internal/cli/factory_handoff_recover.go): both verbs now open the factory through
`factoryCardRoot()` —

- `recover-resume` RunE: `root, err := os.Getwd()` → `root := factoryCardRoot()` (the
  `ReadRuntimeCensus` closure keeps the same root);
- `factory runs` RunE: the same replacement.

**SPEC-wording reversion** (card scope: restore the narrowed wording to the original
meaning now that the divergence is gone):

- `.moai/specs/SPEC-MOAI-DB-PATHS-001/spec.md` — REQ-DBP-002 divergence tail removed;
  § H rewritten from "Known divergence (debt)" to the alignment record naming t1862.
- `acceptance.md` AC-DBP-002 — the "roots at the working directory and agree only when…"
  clause replaced with the shared-root statement.
- `plan.md` §A.2 — divergence sentence removed; risk-table row and the §E open item mark
  the follow-up RESOLVED by card t1862.
- `internal/cli/paths.go` — the resolver-seams comment and the `moai paths` Long help no
  longer claim the working-directory divergence (no test asserted that wording).

**GREEN** (this tree, after the fix) — `go test -run 'TestFactoryRunsOpensMoaiPathsRootNotCwd|TestRecoverResumeOpensMoaiPathsRootNotCwd' -count=1 -timeout 8m ./internal/cli/`:

```
ok  	github.com/modu-ai/moai-adk/internal/cli	4.150s
```

**Scoped regression verification** (rule 35e9 — `-run` narrowed; full suite is CI's):

| Check | Command | Result |
|---|---|---|
| recover/resume + abandon-lane | `go test -run 'TestResumeLegacy\|TestFactoryRecoverResume\|TestAbandonLane' -count=1 ./internal/cli/` | `ok … 4.181s` |
| factory runs + retire guard | `go test -run 'TestFactoryRuns\|TestRetireRun' -count=1 ./internal/cli/` | `ok … 17.764s` |
| paths command (Long text changed) | `go test -run 'TestPaths' -count=1 ./internal/cli/` | `ok … 28.167s` |
| template db-paths surfaces | `go test -run 'TestDBPaths' -count=1 ./internal/template/` | `ok … 0.490s` |
| build | `go build ./internal/...` | exit 0 |
| vet | `go vet ./internal/cli/` | exit 0 |
| gofmt | `gofmt -l <3 changed files>` | empty |

No test was skipped, weakened, or given a larger timeout. Evidence mtime and the commit
below are the baseline; the card-review and sync audit re-read from disk.

## Hold (RESOLVED 06:32Z)

- t1827 landed on local main `026f48658f`; hold lifted, implementation landed above.

## Landing queue wait (07:07Z, leader msg)

Leader confirmed merge-ready (sync-audit PASS, receipt `rcpt-887168f33117a07a6c6f2161`) and
queued this card at position 10: `t1834(진행 중) → t1682 → t1688 → t1703 → t1866 → t1859 →
t1708 → t1680 → t1821 → t1862 → t1709 → t1860`. On my turn: absorb latest main → scoped
re-measure (rule 35e9) → `factory complete`. Pre-verified here: the audited tree for the
sync-audit verdict was HEAD `e707bdf806` (fix + tests); the verdict-file commit `0a752d2c9a`
adds reports only, so the audited code is unchanged by the record commit. Lease renewal
continues every cron fire until the turn arrives.

## Handoff memo — draft PR landing (board c58f/14a6, 2026-10-19)

- Draft PR: **#1833** (https://github.com/modu-ai/moai-adk/pull/1833), head
  `WT-factory-cardroot-align` @ `54be77d2fd`, base `release/main-batch-20261019`
  (`a3fcc98114` at creation). Preservation push done at PR creation (first push of the
  branch); no re-push — the same-PR repeat-push ban holds.
- Next (after the leader merges): card-tree ff to the merged state → scoped re-measure
  (rule 35e9) → `factory_complete` (MCP, in-session) → CodeRabbit commit-status + Merge
  Risk read → landing record.
- t1641 stays HOLD (round-3 FAIL, operator disposition pending) — not affected by this
  landing flow change.

## PR #1833 Codex-review P2 repairs (2026-10-19, head `74134e24f1`)

Leader relayed two P2 findings from the PR's Codex code_review (merge on hold):

- **P2-1** (`factory_handoff_recover.go`, both verbs): with CLAUDE_PROJECT_DIR unset and the
  working directory deleted, `resolveProjectDir()` returned "" and my earlier
  `factoryCardRoot()` swap fed `OpenFactory("")` — a stray factory DB under the home-state
  layout. Measured deeper than the review's premise: on darwin `os.Getwd` can SUCCEED with
  a deleted cwd, so `OpenFactory` recreated the deleted tree and the verb "succeeded" with
  an empty listing (the RED run shows exactly that). Fix: a shared `verifyWorkingDir(dir)`
  helper — Stat the resolved dir; on failure return the underlying working-directory error
  as-is (the review's ask), covering both the "" branch and the deleted-path branch.
- **P2-2** (`factory_relaunch.go` `relaunchRetireFromRun`): still rooted at the raw working
  directory (os.Getwd → FactoryDBPath → ProbeRunStateAt → OpenFactory), contradicting the
  new `moai paths` help guarantee. Fix: the same resolution + `verifyWorkingDir` +
  `factoryRootOf(dir)`.

Verification: RED for both (`TestFactoryRunsAndRecoverRefuseUnresolvableRoot` — runs
succeeded with an unresolvable root; `TestRelaunchRetireFromRunOpensMoaiPathsRoot` —
"unknown (no such run)"), then GREEN (both + the original boundary pair, ok 3.444s), then
on the absorbed tree `4bbd2ed113` with slot go-test-heavy: recover/resume ok 2.021s,
runs+retire ok 6.744s, paths ok 12.509s, template db-paths ok 0.851s, gofmt clean, vet
exit 0, build exit 0. Admission: load 11.56, mem 82%, one Go call at a time.

Pushed `4bbd2ed113..74134e24f1` → PR #1833 head. Slot released immediately.

## Session closure (2026-10-19, leader tn564n operator directive)

lane-38 closes. State at close: HEAD `a3d25aef1f` on `WT-factory-cardroot-align` (pushed,
PR #1833 OPEN, head observed by the leader), factory record merge-ready — merge is HELD on
the operator constraint (@codex review re-request + merge queued by the leader). The
post-merge record (PR number · merge commit · head SHA) will be written by the LEADER per
board 21f2; the lane does not act further. Card work complete: both Codex-review P2 repairs
landed and verified; sync-audit PASS stands (receipt rcpt-887168f33117a07a6c6f2161).
