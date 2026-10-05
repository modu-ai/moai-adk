# SPEC-WORKTREE-KEY-WIRING-001 — Acceptance

## §A Discipline

Every criterion is binary and carries a judging command (or seam-injected
test selector) whose exit/status is recorded in progress.md §E.2. Each
judging command's swept count — the `go test` result line naming the
selected tests, or the grep verdict — is recorded verbatim in progress.md
§E.2; a PASS claimed without its recorded swept count is an unattributed
claim. The OFF
baseline criteria are characterization tests: captured against the pre-M1
tree behavior. Prefix observability: all auto-merge notices carry the new
distinct literal prefix (run-phase chooses the exact literal; distinctness
from `SessionExitCleanupNoticePrefix` and `PRMergeCleanupNoticePrefix` is
asserted by a test). "No window record" means
`kanban.ReadIntegrationLock(root).Held()` is false AND no lock-file mutation
occurred (seam assertion).

## §B Acceptance criteria

**AC-WKW-001** (REQ-001, OFF baseline)
Given `workflow.worktree.auto_merge: false` and a session worktree present,
when the session exits cleanly, then no merge invocation, no window record
write, and no new notice occur — session-exit observable output is
byte-identical to the pre-SPEC baseline (seams-never-called + captured
output comparison).
Judge: `go test ./internal/cli/ -run TestAutoMergeOffBaseline -count=1` → ok.

**AC-WKW-002** (REQ-002, happy path)
Given `auto_merge: true`, a clean exit, and the session branch one commit
ahead of the configured develop branch, when the session exits, then
`git merge --no-ff <branch>` ran with cmd.Dir at the develop worktree and
develop's HEAD is a merge commit with two parents naming the session branch.
Judge: `go test ./internal/cli/ -run TestAutoMergeHappyPath -count=1` → ok.

**AC-WKW-003** (REQ-002, non-clean exit)
Given `auto_merge: true` and a non-zero session exit, when the exit flow
runs, then no merge and no window mutation occur (clean-exit-only).
Judge: `go test ./internal/cli/ -run TestAutoMergeNonCleanExit -count=1` → ok.

**AC-WKW-004** (REQ-003, inert when unconfigured)
Given `auto_merge: true` and develop_branch empty (and, in a second case,
github-flow mode), when the trigger fires, then no merge runs and a
non-blocking notice naming `develop_branch` is emitted on stderr.
Judge: `go test ./internal/cli/ -run TestAutoMergeUnconfiguredTarget -count=1` → ok.

**AC-WKW-005** (REQ-004, busy window)
Given the window held — case 1 by a live session, case 2 by a stale
record — when the trigger fires, then the auto path writes no lock record,
performs no merge, and emits the skip notice; the existing hold is
untouched (ReadIntegrationLock returns the original holder).
Judge: `go test ./internal/cli/ -run TestAutoMergeBusyWindow -count=1` → ok.

**AC-WKW-006** (REQ-004/005, ceremony + zero-push)
Given a successful auto-merge, when it completes, then the lock record
transitioned acquire→release for this session (released at end), and the
seam log of every git invocation on the path contains no push, fetch, pull,
or remote-mutating subcommand — exactly one `merge` with `--no-ff`.
Judge: `go test ./internal/cli/ -run TestAutoMergeZeroPush -count=1` → ok.

**AC-WKW-007** (REQ-006, conflict)
Given a conflicting merge, when the merge fails, then `git merge --abort`
ran in the target worktree (no MERGE_HEAD remains), the window is released,
the skip notice is emitted, and the session worktree's contents (including
any uncommitted files) are byte-identical to pre-trigger.
Judge: `go test ./internal/cli/ -run TestAutoMergeConflict -count=1` → ok.

**AC-WKW-008** (REQ-007, source dirty)
Given uncommitted changes in the session worktree, when the trigger fires,
then the merge is skipped with the notice and the worktree is untouched;
with `auto_cleanup` also on, disposal then follows its own dirty guard.
Judge: `go test ./internal/cli/ -run TestAutoMergeSourceDirty -count=1` → ok.

**AC-WKW-009** (REQ-008, target guards)
Given develop configured but (case 1) no worktree holds it, (case 2) its
worktree dirty, when the trigger fires, then the merge is skipped with the
respective notice.
Judge: `go test ./internal/cli/ -run TestAutoMergeTargetGuards -count=1` → ok.

**AC-WKW-010** (REQ-010, no-op silent)
Given the session branch fully contained in develop, when the trigger
fires, then no window record, no merge invocation, and no notice occur.
Judge: `go test ./internal/cli/ -run TestAutoMergeNoOpSilent -count=1` → ok.

**AC-WKW-011** (REQ-011, auto_create truth)
Given `auto_create: true`, when the init/update/web advisory emits, then
the text matches the AC-WBG-009 observability regex AND contains no
auto-creation claim (negative pattern: no "is auto-creating", no "will be
created"); the `false` wording and the config-failure degradation are
unchanged.
Judge: `go test ./internal/cli/ -run TestWorktreeAdvisoryTruthful -count=1` → ok.

**AC-WKW-012** (REQ-012, honesty artifacts)
Given the change lands, when the verification batch runs, then: the
classification test passes with auto_merge direct-live via the new reader;
the inventory row for `workflow.worktree.auto_merge` is class W with no
`deprecate_after`; no "AutoMerge has no production reader" string remains in
types.go; the template workflow.yaml comment names consumers for all three
auto-* keys; `make build` ran after the template edit.
Judge: `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders
-count=1` → ok &&
`! grep -q "AutoMerge has no production reader" internal/config/types.go`
→ exit 0.

**AC-WKW-013** (REQ-009, prefix distinctness + non-blocking)
Given the auto-merge implementation, when any auto-merge failure path emits
a notice, then every notice line carries the auto-merge literal prefix,
that prefix differs from both `SessionExitCleanupNoticePrefix` and
`PRMergeCleanupNoticePrefix` (constant-level assertion), and a forced merge
failure leaves the session-exit flow's exit status unaffected.
Judge: `go test ./internal/cli/ -run TestAutoMergeNoticePrefixDistinct
-count=1` → ok.

**AC-WKW-014** (REQ-013, toggle independence)
Given all four combinations of `auto_merge` × `auto_cleanup`, when a clean
session exit runs with the merge preconditions satisfied, then merge occurs
if and only if `auto_merge` is on, removal occurs if and only if
`auto_cleanup` is on (subject to its own dirty guard), and no combination
implies the other.
Judge: `go test ./internal/cli/ -run TestAutoMergeToggleIndependence
-count=1` → ok.

## §C Falsification procedures

- **Vacuous OFF baseline**: AC-WKW-001 must fail if the trigger is wired but
  the gate inverted — mutant: flip the auto_merge check; the seam-never-called
  assertion must fail.
- **Ceremony theater**: AC-WKW-006 must fail if acquire is called but release
  is missing — mutant: drop the defer; the released-state assertion must fail.
- **Push leak**: any addition of a push/fetch seam call must fail
  AC-WKW-006's seam-log enumeration — mutant test reviewed in M1.
- **Prefix collision**: appending the auto-merge prefix to an existing
  prefix constant must fail the distinctness assertion.
- **Truthful-wording regression**: restoring the old "is auto-creating"
  string must fail AC-WKW-011's negative pattern.

## §D Definition of Done

1. AC-WKW-001..012 all PASS with recorded judging output (§E.2 evidence).
2. Cross-platform: `GOOS=windows GOARCH=amd64 go build ./...` and darwin
   build exit 0.
3. `golangci-lint run` clean on touched packages; coverage of the new path
   ≥ 85%.
4. Honesty artifacts agree with each other (types.go comment = inventory =
   template comment = reader test).
5. No regression in existing session-worktree tests
   (`go test ./internal/cli/ -run TestSessionWorktree -count=1` ok,
   `TestPRMergeCleanup` ok).
6. Quality gates: TRUST 5 — Tested/Readable/Unified/Secured/Trackable
   satisfied; conventional commits referencing SPEC-WORKTREE-KEY-WIRING-001.
