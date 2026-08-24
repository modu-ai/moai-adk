# acceptance.md — SPEC-WORKTREE-REAPER-001

26 acceptance criteria: M1 seven (AC-WR-001…007), M2 twelve (AC-WR-008…016 plus
AC-WR-024…026), M3 four (AC-WR-017…020), cross-cutting three (AC-WR-021…023).

---

## §0 — Criterion form, and its falsifiability baseline

**The falsifiability defect this section closes.** A criterion of the form
`go test ./pkg/ -run <TestName>` with expected observation "`ok`, exit 0"
**cannot fail**, because Go exits 0 with `ok … [no tests to run]` when the
`-run` pattern matches nothing. Measured on this tree:

```
$ go test ./internal/cli/ -run TestPRMergeCleanup_GhNoAnswerConsultsGitFallback -count=1
ok  github.com/modu-ai/moai-adk/internal/cli  1.111s [no tests to run]   # exit 0
```

Every test-shaped criterion below therefore uses a form that observes the named
test **ran and passed**:

```
go test <pkg> -run '^<TestName>$' -v -count=1 2>&1 | grep -c '^--- PASS: <TestName>'
```

**Form verified in both directions on this tree, 2026-08-24:**

```
$ go test ./internal/cli/ -run '^TestParseWorktreeList_BranchExtraction$' -v -count=1 2>&1 \
    | grep -c '^--- PASS: TestParseWorktreeList_BranchExtraction'
1                                    # existing, passing test → 1

$ go test ./internal/cli/ -run '^TestDoesNotExistAtAll$' -v -count=1 2>&1 \
    | grep -c '^--- PASS: TestDoesNotExistAtAll'
0                                    # grep exit 1 → criterion FAILS
```

**Pre-implementation baseline for every new-test criterion = `0`.** Established
mechanically, not per-criterion by hand: none of the test names introduced below
exists in the tree.

```
$ grep -rn "func TestPRMergeCleanup_GhNoAnswerConsultsGitFallback\|func TestPRMergeCleanup_GhOpenSkipsGitFallback\
\|func TestPRMergeCleanup_UndeterminedMergePreserves\|func TestPRMergeCleanup_ZeroUniqueCommitPreserved\
\|func TestParseWorktreeList_CapturesLockReason\|func TestLockAnchor_\|func TestAnchorDecision_\
\|func TestPRMergeCleanup_PorcelainFailureRemovesNothing\|func TestPRMergeCleanup_T207SamplePreservedByLock\
\|func TestPRMergeCleanup_RefusalClassNamesCause\|func TestCleanMergedOnly_\|func TestCleanStale_" internal/
→ 7 matches, ALL of them pre-existing TestCleanStale_* tests with different suffixes
  (KeepsAnchoredWorktree, KeepsDirtyWorktree, KeepsUnmergedWorktree, PreviewsByDefault,
   RemovesWithYes, SkipsProtectedAndDetached, RejectsMergedOnlyCombination).
  Zero matches for any name this file introduces.
```

Each criterion carries a **`Pre-impl observed:`** line recording what the exact
command returns on this tree today, and a **`Covers:`** line naming the
requirements it exercises.

**Grep-token baseline** (unchanged discipline, re-measured 2026-08-24):
`mergeStateUndetermined` → **0** hits in `internal/`, admissible as a token.
`lockReason` → **22** hits, NOT admissible and not used as a criterion token.

---

## §A — M1: merge-detection repair

### AC-WR-001 — gh MERGED still removes (regression guard)

**Given** the sweep enabled, `gh` present, the gh seam answering `MERGED` for a
`WT-*` branch on a clean unanchored tree,
**When** `prMergeCleanup` runs,
**Then** the tree is removed with the `removed by PR-merge cleanup:` notice.

```
go test ./internal/cli/ -run '^TestPRMergeCleanup_GhPresentMergedRemoves$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_GhPresentMergedRemoves'
```
Expected: `1`. **Pre-impl observed: `1`** (existing M8 test; must stay green
across the seam-signature change).
**Covers:** REQ-WR-001.

### AC-WR-002 — gh no-answer consults the git fallback and removes

**Given** the gh seam answering *no answer* for `WT-forge-counts` **and** the
git-merged seam returning a list containing it,
**When** `prMergeCleanup` runs,
**Then** the tree is removed.

```
go test ./internal/cli/ -run '^TestPRMergeCleanup_GhNoAnswerConsultsGitFallback$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_GhNoAnswerConsultsGitFallback'
```
Expected: `1`. **Pre-impl observed: `0`** (test absent; and the behaviour is
absent — today the gh path returns `""` and the tree is preserved).
**Covers:** REQ-WR-002.

### AC-WR-003 — a determinate gh negative does NOT consult git

**Given** the gh seam answering `OPEN` determinately,
**When** `prMergeCleanup` runs,
**Then** the tree is preserved **and** the git-merged seam was invoked zero
times (call counter in the seam closure).

```
go test ./internal/cli/ -run '^TestPRMergeCleanup_GhOpenSkipsGitFallback$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_GhOpenSkipsGitFallback'
```
Expected: `1`. **Pre-impl observed: `0`**.
**Covers:** REQ-WR-004.

### AC-WR-004 — both sources indeterminate preserves, with a notice

**Given** the gh seam answering *no answer* **and** the git-merged seam
returning an error,
**When** `prMergeCleanup` runs,
**Then** the tree is preserved and stdout names the path and an undetermined
merge state.

```
go test ./internal/cli/ -run '^TestPRMergeCleanup_UndeterminedMergePreserves$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_UndeterminedMergePreserves'
```
Expected: `1`. **Pre-impl observed: `0`**.
**Covers:** REQ-WR-003, REQ-WR-017.

### AC-WR-005 — gh-absent path: fallback is the sole source, notice fires once

**Given** the gh-lookpath seam reporting `false`,
**When** `prMergeCleanup` runs,
**Then** the git fallback decides and the squash-blindness notice appears
exactly once.

```
go test ./internal/cli/ -run '^TestPRMergeCleanup_GhAbsentBranchMergedRemoves$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_GhAbsentBranchMergedRemoves'
go test ./internal/cli/ -run '^TestPRMergeCleanup_GhAbsentEmitsBlindnessNoticeOnce$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_GhAbsentEmitsBlindnessNoticeOnce'
```
Expected: `1` and `1`. **Pre-impl observed: `1` and `1`** (existing M8 tests —
these are the real names; the earlier draft's `TestPRMergeCleanup_GhAbsentFallback`
does not exist and matched nothing).
**Covers:** REQ-WR-005.

### AC-WR-006 — the squash-merge property is not lost

**Given** a squash-merged branch: gh answers `MERGED`, the git-merged seam does
**not** list it,
**When** `prMergeCleanup` runs,
**Then** the tree is removed — git never overrides a determinate gh answer.

```
go test ./internal/cli/ -run '^TestPRMergeCleanup_GhPresentSeesSquashMerge$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_GhPresentSeesSquashMerge'
```
Expected: `1`. **Pre-impl observed: `1`** (existing M8 test).
**Covers:** REQ-WR-001, REQ-WR-004.

### AC-WR-007 — the zero-unique-commit removal class is bounded by the dirty guard

**Given** a `WT-*` branch with zero commits of its own (`git rev-list --count
origin/main..<branch>` = 0), no PR, gh answering *no answer*, and the
git-merged seam listing it — in two variants: (a) the tree is clean; (b) the
tree carries an untracked file,
**When** `prMergeCleanup` runs,
**Then** (a) is removed and (b) is preserved by the dirty guard.

```
go test ./internal/cli/ -run '^TestPRMergeCleanup_ZeroUniqueCommitPreserved$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_ZeroUniqueCommitPreserved'
```
Expected: `1`. **Pre-impl observed: `0`**.
**Covers:** REQ-WR-018, REQ-WR-017.

This is the criterion that pins the removal class M1 newly reaches. Live
instance of the class on this tree: `WT-worktree-reaper`, this SPEC's own
branch — `git rev-list --count origin/main..HEAD` → `0`, and it appears in
`git branch --merged origin/main`. See `design.md` §A.4 for why the class is
accepted rather than excluded, and EC-8 / EC-9 below for its boundaries.

---

## §B — M2: anchor-guard repair

### AC-WR-008 — the porcelain parser captures the stored lock reason

**Given** a porcelain body carrying `locked claude session t207 (pid 36912
start Sun Aug 23 07:26:09 2026)` for one entry, and a bare `locked` line
(reason-less lock) for another,
**When** `parseWorktreeList` parses it,
**Then** the first entry's stored reason is `claude session t207 (pid 36912
start Sun Aug 23 07:26:09 2026)` — git's own `locked ` prefix stripped — and
the second entry is marked locked with an empty reason.

```
go test ./internal/cli/ -run '^TestParseWorktreeList_CapturesLockReason$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestParseWorktreeList_CapturesLockReason'
```
Expected: `1`. **Pre-impl observed: `0`** (the parser's switch has no `locked`
case today).
**Covers:** REQ-WR-006.

### AC-WR-009 — a live locked PID is anchored

**Given** a stored lock reason naming `os.Getpid()`,
**When** the anchor decision runs,
**Then** the verdict is anchored and the reported source is the lock.

```
go test ./internal/session/ -run '^TestLockAnchor_LivePidAnchored$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestLockAnchor_LivePidAnchored'
```
Expected: `1`. **Pre-impl observed: `0`**.
**Covers:** REQ-WR-007, REQ-WR-011.

### AC-WR-010 — a confirmed-dead PID is not anchored by the lock source

**Given** a stored lock reason naming a PID the liveness seam returns
`(alive=false, determined=true)` for,
**When** the anchor decision runs,
**Then** the lock source returns not-anchored.

```
go test ./internal/session/ -run '^TestLockAnchor_DeadPidNotAnchored$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestLockAnchor_DeadPidNotAnchored'
```
Expected: `1`. **Pre-impl observed: `0`**.
**Covers:** REQ-WR-007.

### AC-WR-011 — every indeterminate lock shape is anchored

**Given** each of: a reason with no `pid` token; a reason with a non-integer
pid; a locked entry with an empty reason; a probe returning
`(alive=false, determined=false)`,
**When** the anchor decision runs,
**Then** all four yield anchored.

```
go test ./internal/session/ -run '^TestLockAnchor_IndeterminateIsAnchored$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestLockAnchor_IndeterminateIsAnchored'
```
Expected: `1`. **Pre-impl observed: `0`**.
**Covers:** REQ-WR-008.

The fourth case is reachable only because the probe seam is respecified as
`(alive bool, determined bool)`; today's `session.isProcessAlive` returns a bare
`bool` and cannot express it (`design.md` §B.5).

### AC-WR-012 — a confirmed-dead lock does not produce a removal attempt

**Given** a worktree whose lock names a confirmed-dead PID, merged and clean,
**When** `prMergeCleanup` runs,
**Then** no removal is attempted and the tree is reported as locked-but-inert —
`git worktree remove` refuses a locked tree regardless of PID liveness, so an
attempt would emit a failure notice on every sweep, forever.

```
go test ./internal/session/ -run '^TestLockAnchor_DeadLockRemovalIsInert$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestLockAnchor_DeadLockRemovalIsInert'
```
Expected: `1`. **Pre-impl observed: `0`**.
**Covers:** REQ-WR-021.

### AC-WR-013 — the registry-only path still anchors (the union's other half)

**Given** a worktree with **no** lock line but a live registry entry whose
`cwd` is the tree,
**When** `prMergeCleanup` runs,
**Then** the tree is preserved and the notice names the registry as the source.

```
go test ./internal/cli/ -run '^TestPRMergeCleanup_AnchoredSessionSkipsRemoval$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_AnchoredSessionSkipsRemoval'
go test ./internal/session/ -run '^TestAnchorDecision_RegistryOnlyPathStillAnchors$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestAnchorDecision_RegistryOnlyPathStillAnchors'
```
Expected: `1` and `1`. **Pre-impl observed: `1` and `0`** (the first is the
existing t73 guard, which must stay green — that is what proves the registry
source was retained rather than replaced).
**Covers:** REQ-WR-009, REQ-WR-010, REQ-WR-011, REQ-WR-020.

The no-lock fixture is exactly the unlocked-anchor shape REQ-WR-020 records: the lock source has *no opinion*, not a negative, and the tree survives only because the registry happened to see it. The criterion pins the union's behaviour; it does not establish that the registry is reliable — measured, it is 1-of-5 — which is why REQ-WR-020 is a recorded residual rather than a closed case.

### AC-WR-014 — the t207 sample is preserved by the lock source alone

**Given** a fixture reproducing t207 — branch `WT-web-live-todo`, clean tree,
gh answering *no answer*, the git-merged seam **not** listing the branch, the
session registry **empty**, and the porcelain entry carrying the verbatim
measured lock reason with a live PID,
**When** `prMergeCleanup` runs,
**Then** the tree is not removed and the preserve notice names the lock source.

```
go test ./internal/cli/ -run '^TestPRMergeCleanup_T207SamplePreservedByLock$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_T207SamplePreservedByLock'
```
Expected: `1`. **Pre-impl observed: `0`**.
**Covers:** REQ-WR-006, REQ-WR-007, REQ-WR-011.

**Mandatory sample criterion.** The empty registry is what makes it exercise
the lock guard: t207 is the one tree the registry can see, so a criterion that
let the registry answer would still be 4-of-5 blind on t209 / t210 / t212 /
t213.

### AC-WR-015 — `moai worktree clean --stale` gains the same lock guard

**Given** a worktree that is clean, merged, and carries a lock naming a live
PID, with **no** registry entry,
**When** `cleanStaleWorktrees` runs,
**Then** it is reported as kept with a live-anchor keep-reason and is not
removed even with `--yes`.

```
go test ./internal/cli/worktree/ -run '^TestCleanStale_LockAnchoredWorktreeKept$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestCleanStale_LockAnchoredWorktreeKept'
go test ./internal/cli/worktree/ -run '^TestCleanStale_KeepsAnchoredWorktree$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestCleanStale_KeepsAnchoredWorktree'
```
Expected: `1` and `1`. **Pre-impl observed: `0` and `1`**.
**Covers:** REQ-WR-019.

This is the criterion that makes the repair reach the consumer with the wider
blast radius: `cleanStaleWorktrees` sweeps **all** registered trees (its
provider is `git worktree list --porcelain`), not just `WT-*`, and consumes the
same blind `LiveAnchoredSessions` today (`internal/cli/worktree/clean.go:162`).

### AC-WR-016 — a porcelain parse failure removes nothing

**Given** the worktree-list seam returning an error,
**When** `prMergeCleanup` runs,
**Then** no removal is attempted and a notice is emitted.

```
go test ./internal/cli/ -run '^TestPRMergeCleanup_PorcelainFailureRemovesNothing$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_PorcelainFailureRemovesNothing'
go test ./internal/cli/ -run '^TestPRMergeCleanup_WorktreeListErrorFailOpen$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_WorktreeListErrorFailOpen'
```
Expected: `1` and `1`. **Pre-impl observed: `0` and `1`** (the second is the
existing fail-open guard: the sweep must still not abort its caller).
**Covers:** REQ-WR-008, REQ-WR-016.

### AC-WR-024 — a refusal-class tree is pre-detected, and its notice names the cause

**Given** a merged, porcelain-clean, registry-unanchored candidate carrying a
lock whose PID the probe confirms dead,
**When** `prMergeCleanup` runs,
**Then** no removal is attempted, the tree is preserved, and the preserve notice
**contains the cause-specific token `locked`** — not merely a string that differs
from some other notice.

```
go test ./internal/cli/ -run '^TestPRMergeCleanup_RefusalClassNamesCause$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_RefusalClassNamesCause'
```
Expected: `1`. **Pre-impl observed: `0`** (no such test; and today the candidate
reaches `git worktree remove` and produces the `cannot remove a locked working
tree` failure notice observed for t212 in `.moai/reports/t209/investigation.md`).
**Covers:** REQ-WR-021, REQ-WR-023.

Rebuilt at v0.3.0 on two counts. Its v0.2.0 second fixture — an ignored-only tree
as a second *refusal* cause — models a condition measured not to exist (EC-9), so
it is gone. And its declared assertion was string **inequality**, which any two
distinct literals satisfy including ones that name nothing; REQ-WR-023 requires
the notice to *name* the cause, so the assertion is now a positive token match.

### AC-WR-025 — the ignored-content prevalence measurement (M1 precondition)

**Given** the repository's registered worktrees,
**When** the measurement below is run **from outside every worktree a session
occupies** — a constraint the v1 EC-9 failure established (`design.md` §A.6),
**Then** its result is recorded and `design.md` §A.7's decision rule selects the
ignored-content policy.

```
# from the primary checkout, for each registered worktree path:
git -C <tree> status --porcelain --ignored | grep -c '^!!'
# cross-referenced against the M1-unblocked set (merged AND porcelain-clean
# AND unanchored), recording per tree: any ignored entry? all under a
# regenerable path?
```
Expected observation: a count of M1-unblocked trees that carry ignored content,
and of those, how many carry ignored content **outside** regenerable paths.
Decision: `design.md` §A.7 — P1 if ignored content is present in at most half the
unblocked set, otherwise choose P2 or P3.

**Pre-impl observed: not run — and not runnable from here.** The
worktree-isolation guard refuses `cd` and `git -C` into sibling trees, so this
session observes only its own tree (`git status --porcelain --ignored` → 7
entries, 5 of them `!!`, including `.moai/state/config-cache.json` and
`.moai/state/context-usage.json`). That single data point is consistent with the
hypothesis and does not test it.

[HARD] **This is a gate on M1, not a report.** M1 does not land before this
measurement is taken and §A.7's fork closed. If ignored content proves near
universal, policy P1 preserves nearly the whole population and M2's guard cancels
M1's unblocking of the ~98 merged trees.
**Covers:** REQ-WR-024.

### AC-WR-026 — the `--merged-only` sweep gains the same anchor decision

**Given** a merged worktree carrying a lock naming a live PID and **no** registry
entry,
**When** the `--merged-only` path runs,
**Then** it is reported as kept and not removed.

```
go test ./internal/cli/worktree/ -run '^TestCleanMergedOnly_LockAnchoredWorktreeKept$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestCleanMergedOnly_LockAnchoredWorktreeKept'
```
Expected: `1`. **Pre-impl observed: `0`.**
**Covers:** REQ-WR-019.

This is the third consumer, unnamed until v0.3.0 and the most exposed of the
three: `internal/cli/worktree/clean.go:95`, whose own comment records that
`--merged-only` "has no dirty guard of its own, so this is the only protection
between the sweep and a live lane's tree". Verified:
`grep -n 'LiveAnchoredSessions' internal/cli/worktree/clean.go` → `95:`, `163:`.

---

## §C — M3: non-`WT-*` coverage via `worktree clean --stale`

The M3 decision is **extend the shipped `moai worktree clean --stale`**, not
build a parallel inventory surface (`design.md` §C, option O3-d). These criteria
pin the three measured gaps that extension closes, and guard the two properties
it already has.

### AC-WR-017 — machine-readable output covering every tree

**Given** a fixture worktree set containing `WT-*` and non-`WT-*` entries,
**When** `clean --stale --json` runs,
**Then** stdout is valid JSON with one object per non-protected tree, each
carrying path, branch, keep-reason (empty when removable), dirty state, merge
state, and anchor state — and the non-`WT-*` entries are present.

```
go test ./internal/cli/worktree/ -run '^TestCleanStale_JSONEmitsAllTreesWithStates$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestCleanStale_JSONEmitsAllTreesWithStates'
```
Expected: `1`. **Pre-impl observed: `0`** (`clean` has no `--json` flag today:
its flag set is `--merged-only`, `--stale`, `--yes`, `--base`).
**Covers:** REQ-WR-012.

### AC-WR-018 — preview stays the default; removal still requires `--yes`

**Given** a fixture set with removable trees,
**When** `clean --stale` runs without `--yes`,
**Then** nothing is removed and the preview text is emitted; **and when** `--yes`
is passed, removal proceeds.

```
go test ./internal/cli/worktree/ -run '^TestCleanStale_PreviewsByDefault$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestCleanStale_PreviewsByDefault'
go test ./internal/cli/worktree/ -run '^TestCleanStale_RemovesWithYes$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestCleanStale_RemovesWithYes'
```
Expected: `1` and `1`. **Pre-impl observed: `1` and `1`**.
**Covers:** REQ-WR-013, REQ-WR-014.

These two pre-existing tests are why REQ-WR-014's "forward constraint" needs no
new mechanism: the shipped command already previews by default and gates removal
behind an explicit flag. The criterion's job is to stop the extension regressing
it.

### AC-WR-019 — the base ref resolves to `origin/main`, not local `main`

**Given** a repository whose local `main` is behind `origin/main`,
**When** `clean --stale` runs with no `--base`,
**Then** the merge comparison is against `origin/main`.

```
go test ./internal/cli/worktree/ -run '^TestCleanStale_BaseDefaultsToOriginMain$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestCleanStale_BaseDefaultsToOriginMain'
```
Expected: `1`. **Pre-impl observed: `0`** (`cmd.Flags().String("base", "main", …)`
— the default is the local branch, which diverges from what `prMergeCleanup`
compares against).
**Covers:** REQ-WR-022.

### AC-WR-020 — no interactive prompt on the changed surface

**Given** the changed `internal/cli/worktree` surface,
**When** the package's no-prompt guard runs,
**Then** it reports no `AskUserQuestion` reference.

```
go test ./internal/cli/worktree/ -run '^TestCleanStale_NoAskUserQuestion$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestCleanStale_NoAskUserQuestion'
```
Expected: `1`. **Pre-impl observed: `0`.** The name is exact and anchored, because an
unanchored `-run 'NoAskUserQuestion'` matches every existing guard and would
bind nothing about this change. Measured:
`grep -rn 'func Test.*NoAskUserQuestion' internal/cli/ | wc -l` → **31**.
**Covers:** subagent-boundary craft constraint (not a REQ).

---

## §D — Cross-cutting

### AC-WR-021 — toggle-off is inert

**Given** `auto_cleanup: false`,
**When** `prMergeCleanup` runs,
**Then** the worktree-list seam is not invoked and stdout is empty.

```
go test ./internal/cli/ -run '^TestPRMergeCleanup_ToggleOffNoOp$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_ToggleOffNoOp'
go test ./internal/cli/ -run '^TestPRMergeCleanup_NilCfgNoOp$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_NilCfgNoOp'
```
Expected: `1` and `1`. **Pre-impl observed: `1` and `1`.** Scope note: the
existing `ToggleOffNoOp` asserts *`wtList` was not called* plus empty output —
not "no git seam is invoked" across all seams. The criterion is stated to match
what the test actually asserts; extending it is not required and is not claimed.
**Covers:** REQ-WR-015.

### AC-WR-022 — affected packages build, vet, and test clean

```
go build ./... ; echo "build=$?"
go vet ./internal/cli/... ./internal/session/... ; echo "vet=$?"
go test ./internal/cli/... ./internal/session/... -count=1 -timeout 600s ; echo "test=$?"
GOOS=windows go vet ./internal/cli/... ./internal/session/... ; echo "winvet=$?"
```
Expected: all four `=0`. **Pre-impl observed: not run as a criterion** — it is a
completion gate, not a discriminator. The `internal/cli` 600s floor is a repo
rule (the package measures ~336s standalone). `GOOS=windows go vet` proves
cross-platform **compilation only**; it is not evidence of Windows behaviour,
which matters here because the Windows liveness probe returns `true`
unconditionally (`design.md` §B.5).
**Covers:** all (build/lint gate).

### AC-WR-023 — the real t207 tree survives a real sweep

Two parts. **Part (a) is mandatory and non-destructive. Part (b) executes only
under explicit operator authorisation**, because running the repaired sweep in
this repository with `auto_cleanup: true` performs the mass removal that
`spec.md` §G declares out of scope.

**(a) — inert-sweep check.** With `workflow.worktree.auto_cleanup` temporarily
set to `false`, using the **tree-local** binary (the installed `~/go/bin/moai`
may predate the change):

```
go build -o ./bin/moai ./cmd/moai && echo "build=$?"
./bin/moai session list --json > .moai/state/verify/t209/sweep-a.log 2>&1
grep -c 'PR-merge cleanup' .moai/state/verify/t209/sweep-a.log      # expected 0
test -d .claude/worktrees/t207 && echo PRESERVED                     # expected PRESERVED
```

**(b) — authorised-sweep check.** Only after the operator has enumerated the
trees they intend to remove into `expected-removals.txt`:

```
./bin/moai session list --json > .moai/state/verify/t209/sweep-b.log 2>&1
grep -o 'worktree [^ ]*' .moai/state/verify/t209/sweep-b.log | sed 's/^worktree //' \
  | grep 'removed by PR-merge cleanup' -c                            # count of removals
comm -13 <(sort expected-removals.txt) <(grep 'removed by PR-merge cleanup' \
  .moai/state/verify/t209/sweep-b.log | sed 's/.*worktree \([^ ]*\).*/\1/' | sort)
```
Expected: the `comm` output is **empty** — every tree actually removed was in
the pre-enumerated expected set — and `.claude/worktrees/t207` is absent from
that set and still on disk.

**Pre-impl observed: not run** (the repair does not exist; running it today
sweeps under the blind guard). Evidence is persisted under
`.moai/state/verify/t209/` rather than `/tmp`, which the OS clears.
**Covers:** REQ-WR-006, REQ-WR-007, REQ-WR-017 (observational end-to-end).

---

## §E — Edge cases

| # | Case | Expected |
|---|---|---|
| EC-1 | gh returns valid JSON with an unrecognised `state` value | determinate, not merged; git not consulted |
| EC-2 | gh exits 0 with empty stdout | no answer → git consulted |
| EC-3 | branch merged per git but the PR was closed unmerged | gh answers `CLOSED` determinately → preserved (AC-WR-003) |
| EC-4 | lock line on a detached-HEAD worktree | anchored; the prefix filter skips it in `prMergeCleanup` anyway, and `clean --stale` already keeps detached trees |
| EC-5 | the same PID locks two worktrees | both anchored |
| EC-6 | lock present **and** tree dirty | dirty guard fires first; preserved either way |
| EC-7 | `origin/main` ref absent (fresh clone) | git-merged seam errors → undetermined → preserved (AC-WR-004) |
| EC-8 | branch with **zero commits of its own** | git reports it merged (it is an ancestor of the base). Clean ⇒ removable, and nothing **committed, tracked, or untracked** is lost; dirty/untracked ⇒ preserved (AC-WR-007). **Ignored content is NOT protected** — see EC-9. This is the class M1 newly reaches (`design.md` §A.4) |
| EC-9 | tree whose only content is **gitignored** files | **Removed today, and the ignored content is destroyed.** Measured (`.moai/reports/t209/ec9-measurement.md` **v2** §Q1 — v1 claimed the opposite and was wrong): `git status --porcelain` → 0, `git status --porcelain --ignored` → 1, `git worktree remove` → **exit 0**, tree gone. `git status --porcelain` and `git worktree remove` agree in disregarding ignored files, so the dirty guard has no backstop for this class. Whether the sweep preserves such a tree is the **open decision** at `design.md` §A.7, gated on AC-WR-025 |
| EC-10 | `origin/main` **stale** (never fetched by the sweep) | a behind ref yields fewer ancestors, hence fewer removals — the safe direction. The one exception is a force-pushed / rewritten `origin/main`, where a branch could appear merged that is not. The sweep does not fetch; this is recorded as a bounded residual, not a claim of safety |
| EC-11 | several trees are preserved for **different** reasons in one sweep | Each preserve notice carries a cause-specific token (REQ-WR-023): refusal-class, dirty, anchored-by-lock, anchored-by-registry, undetermined-merge. Re-derived at v0.3.0: the v0.2.0 form paired a locked tree with an ignored-only tree as two refusal causes, but the second was measured not to be a refusal at all (EC-9) |
| EC-12 | populated **submodule** in a merged clean tree | Non-forced `git worktree remove` refuses (exit 128, `working trees containing submodules cannot be moved or removed`, not curable by `--force`) with a clean porcelain — a real member of REQ-WR-021's class that the pre-detection set does not cover. Falls through to fail-open: git refuses, a cause-naming notice is emitted, nothing is lost. No live instance here (`.gitmodules` absent) — held out of scope with that reason |
| EC-13 | a candidate goes **dirty between the guard and the removal** | For tracked or untracked content the race is benign because git re-checks at removal time and refuses — a preserved tree plus a notice. For ignored content there is no second observation and therefore no protection (`design.md` §B.11). The v1 EC-9 fixture is a live instance: the statusline wrote two untracked files into the window |

## §F — Definition of Done

- [ ] All 26 criteria run with the exact commands above; each `grep -c` output
      cited verbatim
- [ ] Every new-test criterion moved from its recorded pre-impl `0` to `1`
- [ ] AC-WR-023(a) executed; (b) executed only with explicit operator
      authorisation and an enumerated expected-removal set
- [ ] `.claude/worktrees/t207` on disk; `WT-lint-heading` untouched; no tree
      backing an open PR removed
- [ ] Every requirement REQ-WR-001…024 named by at least one `Covers:` line
- [ ] The lock-based anchor decision reached **all three** consumers (AC-WR-015, AC-WR-026)
- [ ] AC-WR-025 run from outside every worktree, and `design.md` §A.7's fork closed, BEFORE M1 lands
- [ ] `design.md` §C decision honoured: `clean --stale` extended, no parallel
      inventory surface added
- [ ] English code, comments, godoc; `t.TempDir()` isolation; existing
      function-variable seams extended, not replaced
