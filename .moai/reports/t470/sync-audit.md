# Sync-phase Audit — SPEC-QUEUE-UPGRADE-PROOF-001 (card t470)

Independent, adversarial. Every figure below was measured in THIS run, in this
worktree, against this HEAD. Nothing is carried over from `verdict.md`; where a
recorded figure is cited it is cited as a comparison against my own measurement,
never as a substitute for one.

**Verdict: PASS-WITH-DEBT — 91.6** (weighted harmonic mean; Tier M).
Must-pass firewall (Functionality, Security) clears independently.
One blocking finding, remedy one line. No code defect.

| Dimension | Weight | Score | Verdict |
|---|---|---|---|
| Functionality | 40% | 92 | PASS |
| Security | 25% | 96 | PASS |
| Craft | 20% | 84 | PASS |
| Consistency | 15% | 95 | PASS |

Weighted harmonic mean = 1 / (0.40/92 + 0.25/96 + 0.20/84 + 0.15/95) = **91.64**.
(The arithmetic mean would have been 91.7 — they nearly coincide because no
dimension is an outlier; the harmonic form is used because it is the rule, not
because it changed the answer.)

---

## Claim

1. The GREEN is falsifiable. The AC-QUP-010 mutation, applied by me from
   scratch, produces RED with the symptom set the record names.
2. The test enters the composed path through the CLI, and its isolation is real
   — structurally and observationally.
3. AC-QUP-002 is carried by the sentinel limb and the negative-existence limb;
   the `backlog.db` limb does not discriminate. The implementer's report of
   this is accurate, and the two remaining limbs are sufficient.
4. The committed mutation seam is acceptable as delivered, with a recorded
   craft finding: one of its two parameters fails SILENTLY if flipped alone.
5. Fixture F1 is faithful to what v3.1.2 actually writes — verified against
   that tag's source, not against the comment asserting it. F2 is correctly
   quarantined as forward-compatibility-only.
6. AC-QUP-008's Gap is honest and conservative — I reproduced the controlled
   window independently and it is clean. AC-QUP-009's baseline substitution is
   a legitimate correction. **But AC-QUP-009's path allowlist is violated at
   final HEAD by `CHANGELOG.md`, and the verdict asserts otherwise.**

---

## Evidence

### E1 — GREEN, reproduced

```
$ go test ./internal/cli/ -run 'TestTodoComposedUpgrade' -count=1 -v -timeout 300s
=== RUN   TestTodoComposedUpgrade_FromLegacyV312Layout
--- PASS: TestTodoComposedUpgrade_FromLegacyV312Layout (1.37s)
=== RUN   TestTodoComposedUpgrade_ForwardCompatibleFieldsSurvive
--- PASS: TestTodoComposedUpgrade_ForwardCompatibleFieldsSurvive (1.14s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	4.322s
```

Selector matched 2, both RUN and PASS. Non-empty swept set (AC-QUP-005).

### E2 — RED, re-established by me, not accepted on the log

I did not read `red-mutation.log` as evidence. I backed the file up, applied the
mutation at the two committed literals myself, ran, observed, and reverted.

```
$ sed -i '' '221s/, false)/, true)/; 223s/, false)/, true)/' internal/cli/todo_composed_upgrade_test.go
$ sed -n '221,224p' internal/cli/todo_composed_upgrade_test.go
	legacyDir, currentDir := seedLegacyV312Layout(t, root, f1LegacyBacklogJSON, true)

	assertPreUpgradeState(t, legacyDir, currentDir, true)
	assertQueueRootIsolated(t, root)

$ go test ./internal/cli/ -run 'TestTodoComposedUpgrade_FromLegacyV312Layout' -count=1 -v -timeout 300s
=== RUN   TestTodoComposedUpgrade_FromLegacyV312Layout
    todo_composed_upgrade_test.go:236: relocation sentinel must be readable under the current name: open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoComposedUpgrade_FromLegacyV312Layout2228780865/002/.moai/state/todo/companions.json: no such file or directory
    todo_composed_upgrade_test.go:236: legacy state dir "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoComposedUpgrade_FromLegacyV312Layout2228780865/002/.moai/state/kanban" must no longer exist after the upgrade (stat err = <nil>)
    todo_composed_upgrade_test.go:236: quarantined legacy document ".../002/.moai/state/todo/backlog.json.migrated" must exist: ... no such file or directory
    todo_composed_upgrade_test.go:249: composed upgrade yielded 0 items, want 3: []
--- FAIL: TestTodoComposedUpgrade_FromLegacyV312Layout (0.40s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.530s
```

Then reverted from the backup:

```
$ cp <scratch>/orig_test.go.bak internal/cli/todo_composed_upgrade_test.go
$ git status --short && git diff --stat
(no output — tree clean, byte-identical to HEAD)
```

**Match against the record**: four failing assertions, at lines 236/236/236/249,
naming the same four symptoms. The recorded RED is what it says it is.

**The mutation reaches the criterion it is supposed to reach.** The symptom
AC-QUP-010 names — "the legacy directory STILL EXISTS after the command" — is
present verbatim as the second failure. The GREEN is falsifiable at the named
place, not merely somewhere.

### E3 — the composed-path entrance and the isolation

Read, not assumed. The chain, verified in source at this HEAD:

- `internal/cli/todo.go:72` `resolveTodoQueueRoot()` → `kanban.ResolveTodoQueueRootAdopting(resolveProjectDir())`
- `internal/cli/todo.go:57` `todoBacklogPath()` → `kanban.BacklogPathForRootAdopting(root)` — reached from `newTodoStore()` at `:79`, i.e. INSIDE the command's run, not from the test helpers.
- `internal/kanban/state_dir.go:81-88` stale-copy branch; `:100-103` relocation branch. The cited coordinates are correct.

Three isolation layers, all present:

- `todoFixture` (`internal/cli/todo_test.go:61`) = `t.TempDir()` + `t.Setenv("CLAUDE_PROJECT_DIR", root)` + `initGitRepo`. The git repository is load-bearing: `primaryCheckoutRoot` (`internal/kanban/todo_root.go:96-101`) is the early return, so without it the resolution falls through to the home branch.
- `composedUpgradeFixture` additionally redirects `userHomeDirFn` to a second temp directory, closing that fallback.
- `assertQueueRootIsolated` asserts, BEFORE any command, `sameDir(resolved, root)` and `queueRootInsideTemp(resolved)`; `runTodo`'s t422 guard fails the test outright if the root ever names the live checkout.

**One contamination path I probed and found closed.** `composedUpgradeFixture`
calls `todoFixture`, which constructs a store through the ADOPTING
`todoBacklogPath(root)` — before the legacy layout is seeded. Had that call
created the current-name directory, the whole proof would be severed exactly the
way the mutation severs it. It does not: at that moment neither directory
exists, `resolveStateDir` returns the current name without a mkdir, and
`assertPreUpgradeState`'s "current state dir must NOT exist" precondition is the
empirical proof of it — it passes in GREEN. The order is safe, and the test
already guards its own ordering.

`assertQueueRootIsolated` itself calls only `resolveTodoQueueRoot()`, which on a
git-repository base returns from `primaryCheckoutRoot` before
`adoptLocalTodoQueue` is reached. The assertion performs no relocation, so the
relocation observed after the command is attributable to the command.

### E4 — which limbs of AC-QUP-002 carry it

Confirmed from the RED output above, by which assertions did and did not fire:

| AC-QUP-002 limb | Fired under mutation? |
|---|---|
| current state dir exists | no |
| `backlog.db` exists and non-empty | **no** — the store creates an empty DB at the resolved path |
| relocation sentinel byte-identical under the new name | **yes** |
| legacy state dir no longer exists | **yes** |

The implementer's claim is accurate. **The two remaining limbs are sufficient**,
and are the right two: one positive (contents arrived, byte-identical) and one
negative (the source is gone). Together they exclude both failure shapes — a
relocation that never happened, and one that lost the directory's contents. The
`backlog.db` limb is not load-bearing and should not be read as if it were.

### E5 — fixture fidelity (AC-QUP-006), verified at the tag

Not taken from the comment asserting it:

```
$ git show v3.1.2:internal/kanban/backlog_store.go | grep -n "type BacklogRecord" -A 5
75:type BacklogRecord struct {
76-	Version int           `json:"version"`
77-	LastSeq int           `json:"last_seq"`
78-	Items   []BacklogItem `json:"items"`
79-}

$ git show v3.1.2:internal/kanban/backlog_store.go | grep -n "type BacklogItem" -A 7
64:type BacklogItem struct {
65-	ID      string       `json:"id"`
66-	Text    string       `json:"text"`
67-	AddedAt string       `json:"added_at"`
68-	SpecID  *string      `json:"spec_id"`
69-	State   BacklogState `json:"state"`
70-}

$ ... | grep 'BacklogState[A-Za-z]* BacklogState = '
54:	BacklogStateQueued BacklogState = "queued"
56:	BacklogStatePicked BacklogState = "picked"
58:	BacklogStateDropped BacklogState = "dropped"
```

F1 carries `version`, `last_seq`, `items` and nothing else; every item field and
every state value it uses exists at v3.1.2. **F1 is faithful.** F2 carries
`findings` and `archived`, is used only by the optional M2 test, and its own
comment states the reachability caveat in the right terms. The two are not
confused: `f1LegacyBacklogJSON` and `f2ForwardCompatibleJSON` are distinct
constants passed to distinct tests, and only F1 backs the user-facing criteria.

### E6 — AC-QUP-008, re-measured independently

Path derived, not hardcoded, and the `git rev-parse` issued as its own command
per the criterion's worktree note:

```
$ git rev-parse --path-format=absolute --git-common-dir
/Users/goos/MoAI/moai-adk-go/.git
```
→ `QUEUE_DB=/Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db`
(absolute, resolution succeeded — the criterion's resolution-failure branch did
not fire).

```
$ shasum -a 256 "$QUEUE_DB"   # before
ecefa722b3b1181a8864e02ff0257301ed368d3f652b439f35b20521004d802f
$ stat -f '%m %z' "$QUEUE_DB" # before
1788422246 368640
$ go test ./internal/cli/ -run 'TestTodoComposedUpgrade' -count=1 -timeout 300s
ok  	github.com/modu-ai/moai-adk/internal/cli	2.369s
$ shasum -a 256 "$QUEUE_DB"   # after
ecefa722b3b1181a8864e02ff0257301ed368d3f652b439f35b20521004d802f
$ stat -f '%m %z' "$QUEUE_DB" # after
1788422246 368640
```

Byte-identical, same mtime, same size — a **third** independent controlled
window, taken by an actor with no stake in the close. `git status --short
.moai/state/` was not used (`.gitignore` ignores that tree; the check cannot
fail and therefore asserts nothing).

### E7 — AC-QUP-009, ancestry and the diff at FINAL HEAD

The substitution's premise, verified:

```
$ git merge-base --is-ancestor 4e4607abe 6765a75c0 ; echo rc=$?
rc=0
$ git merge-base --is-ancestor 6765a75c0 HEAD ; echo rc=$?
rc=0
$ git rev-list --count 4e4607abe..6765a75c0
37
```

The `acceptance.md`-named pin is indeed an ancestor of the absorbed tip, and the
named diff would span 37 commits of unrelated develop work. The substitution is
a legitimate baseline correction: both endpoints are pinned SHAs, and the
substituted one is the tightest ancestor that isolates this card's authorship.

The diff at final HEAD `077da90b8` — the measurement the verdict does NOT carry:

```
$ git diff --stat 6765a75c0..HEAD
 .moai/reports/t470/cli_suite.log                   |   2 +
 .moai/reports/t470/cli_suite_start.txt             |   2 +
 .moai/reports/t470/green.log                       |   6 +
 .moai/reports/t470/live_queue_ctrl_after.txt       |   2 +
 .moai/reports/t470/live_queue_ctrl_before.txt      |   2 +
 .moai/reports/t470/plan-audit-iter2.md             | 480 +++++
 .moai/reports/t470/plan-audit.md                   | 388 +++++
 .moai/reports/t470/red-mutation.log                |   9 +
 .moai/reports/t470/verdict.md                      | 333 +++++
 .../SPEC-QUEUE-UPGRADE-PROOF-001/acceptance.md     | 248 +++++
 .moai/specs/SPEC-QUEUE-UPGRADE-PROOF-001/plan.md   | 279 +++++
 .../specs/SPEC-QUEUE-UPGRADE-PROOF-001/progress.md | 293 +++++
 .moai/specs/SPEC-QUEUE-UPGRADE-PROOF-001/spec.md   | 290 +++++
 CHANGELOG.md                                       |   1 +
 internal/cli/todo_composed_upgrade_test.go         | 321 +++++
 15 files changed, 2656 insertions(+)

$ git diff --name-only 6765a75c0..HEAD -- internal/ | grep -v '_test\.go$' | wc -l
       0
```

**`CHANGELOG.md` is outside the criterion's allowlist.** See F1 below.

### E8 — quality gate, run by me

```
$ gofmt -l internal/cli/todo_composed_upgrade_test.go   → (no output)
$ go vet ./internal/cli/ ./internal/kanban/...          → (no output, exit 0)
$ go test ./internal/kanban/... -count=1 -timeout 500s
ok  	github.com/modu-ai/moai-adk/internal/kanban	146.737s
$ go test ./internal/cli/ -run 'TestTodoComposedUpgrade' -count=1 -race -timeout 300s
ok  	github.com/modu-ai/moai-adk/internal/cli	3.899s
```

The race run is an addition to the delivered gate: the test mutates the
package-level `userHomeDirFn`, so a race was worth excluding rather than
assuming. It is clean.

`golangci-lint` — recorded as a Gap in the verdict — run here on the axis that
matters for F2 below:

```
$ golangci-lint run --default=none -E unparam,unused ./internal/cli/...
48 issues:
* unparam: 48
$ ... | grep todo_composed_upgrade
NO FINDINGS in todo_composed_upgrade_test.go
```

The new file adds zero issues to a package already carrying 48. Notably `unparam`
does **not** flag the always-`false` mutation parameters, which independently
confirms the verdict's residual-risk statement that no mechanical guard exists.

### E9 — the excluded gaps, spot-checked

The G3 correction is the one exclusion that makes a factual claim about existing
code, so it is the one worth checking:

```
$ sed -n '135,139p' internal/kanban/backlog_concurrency_test.go
func TestConcurrencyStress(t *testing.T) {
	t.Parallel()
	const attempted = stressWriters * stressAddsPerWriter
	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
```

Exactly as claimed: line 135, in-process goroutines, ONE store object. The
correction ("the gap is cross-PROCESS, not 'no concurrency test at all'") is
accurate, and the exclusion reason (C-4 forbids spawning a second process) is
consistent with the constraints this card was dispatched under. G4 and G5 are
design/feature decisions correctly scoped out of a proof card. G2's absorption
is verifiable — its content maps onto AC-QUP-001a/001b/002/003/004/006, all
present. **No exclusion is a defect dressed as a decision.**

---

## Findings

### F1 [MINOR] [blocking] `.moai/reports/t470/verdict.md` §E6 + `acceptance.md` AC-QUP-009 — the path allowlist is violated at final HEAD, and the report asserts it is not

`acceptance.md:169-171` requires that **every** changed path be a `_test.go`
file, a file under `.moai/specs/SPEC-QUEUE-UPGRADE-PROOF-001/`, or under
`.moai/reports/t470/`. At final HEAD, `CHANGELOG.md` is none of those (E7).

`verdict.md` §E6 states: *"Every changed path is a `_test.go` file, a file under
`.moai/specs/…`, or under `.moai/reports/t470/`."* That sentence is **false at
HEAD**, and the diff it cites was taken *before* the M1 commit and before the
sync commit — so it does not contain the paths it generalizes over. That is the
unobserved-claim shape (`verification-claim-integrity.md` §1.1 surface 2): the
assertion outran its own measurement.

Two mitigations keep this MINOR rather than serious. The requirement the
criterion serves (`REQ-QUP-009`, no production behavior changed) **is** met, and
I measured it directly — `git diff --name-only 6765a75c0..HEAD -- internal/ |
grep -v '_test\.go$'` returns 0. And the `CHANGELOG.md` edit is deliberate and
argued in `progress.md:138`, not an accident. What is wrong is the record, not
the work.

**Required fix (one line, no code):** amend `verdict.md` §E6 to state the
measurement at final HEAD and name `CHANGELOG.md` as a fourth permitted path
(the sync-phase artifact set), OR amend `acceptance.md` AC-QUP-009's allowlist to
include it. Do not leave the current sentence standing — a committed verdict
carrying a claim falsified in one command is exactly what a later reader will
trust.

### F2 [MINOR] [optional] `internal/cli/todo_composed_upgrade_test.go:69,106` — the mutation seam has one parameter that fails silently

Judged as asked. **Acceptable as delivered, with a correction to the reason given
for keeping it.**

The two parameters do not behave alike under a stray flip:

- `preCreateCurrentDir=true` alone → the test goes **RED, loudly**. Self-revealing.
- `currentDirPreCreated=true` alone → the test stays **GREEN**, and the
  "current state dir must NOT exist before the upgrade" precondition is silently
  skipped (`:115-117` early-returns). Nothing observable changes in the committed
  configuration, so the disarming would survive review, CI, and `unparam` (E8).

Its harm is latent rather than immediate — it only bites if a later editor ALSO
pre-creates the directory — which is why this is optional and not blocking. But
it is a real silent-degradation seam, and "documented in place" is a weaker
guard than the residual-risk section implies.

**The stated justification for not repairing it is wrong.** `verdict.md`
Residual-risk and `progress.md §H` say repairing it "would add production-shaped
scaffolding to a card whose REQ-QUP-009 forbids exactly that." REQ-QUP-009
forbids changing **production** files; both parameters live in a `_test.go` file
and any guard would too. The honest reason to keep the seam is that reproduction
becomes a two-literal flip — which is a real benefit and a sufficient one. The
recommendation is to correct the reason, not necessarily the code.

**Optional fix:** delete both parameters and apply the mutation as a temporary
local edit (which is what both the implementer and I actually did — E2), or add
one line to `assertPreUpgradeState` refusing `currentDirPreCreated` when the
directory does not in fact exist.

### F3 [MINOR] [optional] `todo_composed_upgrade_test.go` AC-QUP-004 block — discriminates nothing and duplicates AC-QUP-002's check

The AC-QUP-004 block re-runs `os.Stat(dbPath)` plus the size check, byte-for-byte
the same assertion already made ~15 lines above inside the AC-QUP-002 block.
Neither copy fires under the named mutation (E4), because the store creates a
database at the resolved path either way. So AC-QUP-004 passes identically on the
intact and the severed path, and adds no discriminating power beyond
AC-QUP-001a's content assertions.

This is disclosed in the CHANGELOG entry as a property of AC-QUP-002's limbs. It
is **not** disclosed as a property of AC-QUP-004 itself, which is where it
matters more: a criterion that cannot distinguish success from failure carries no
weight. **Optional fix:** record in `acceptance.md` that AC-QUP-004 is a presence
check, not a composition check, so no later reader counts it as evidence the
composition worked.

### F4 [INFO] [optional] the delivered gate omitted `golangci-lint`

Recorded as a Gap in the verdict, correctly. Closed here (E8): clean on this
file. No action.

---

## Baseline-attribution

- **Tree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t470`, confirmed by
  `git rev-parse --show-toplevel`. Branch `WT-queue-upgrade-proof`.
- **HEAD**: `077da90b8984da979387c06865f05b4c274b2bfa`. Working tree clean
  before and after my mutation round (`git status --short` → no output both
  times; the revert was verified by `git diff --stat` → no output).
- **Diff baseline**: `6765a75c0` — the same pinned SHA the verdict substituted,
  whose ancestry relation to `4e4607abe` and to HEAD I verified myself (E7),
  rather than accepting the substitution's premise.
- **v3.1.2 fixture baseline**: the `v3.1.2` tag object in this repository, read
  via `git show` (E5).
- **Live-queue baseline**: derived at run time from
  `git rev-parse --path-format=absolute --git-common-dir` (E6), never hardcoded.
- Every command above was run by me, in this run, against this tree. No figure
  is reused from `verdict.md`, `green.log`, `red-mutation.log`, or
  `cli_suite.log`; where those files are mentioned it is as a *comparison target*
  for a measurement I took independently.
- Traceability carriers, checked: card id `t470` in all 6 card commit messages;
  branch is `WT-` + descriptive slug carrying no card id; evidence path
  `.moai/reports/t470/`. All three present.

---

## Gaps — what I did NOT observe

- **`./internal/cli/` full package suite.** Not re-run. It takes ~935s by the
  delivered record, past the Bash ceiling, and running it would have put a second
  concurrent full-package load on a machine already carrying lanes. I ran the
  targeted selector (twice, plus once under `-race`) and the whole
  `./internal/kanban/...` package. **The full-`internal/cli` verdict in this
  report rests on the implementer's `cli_suite.log`, not on a run of mine** —
  stated so it is not mistaken for a measurement of mine.
- **`go test ./...`** — prohibited by the dispatch and not attempted.
- **Cross-platform.** No `GOOS=windows` build. The file is test-only with no
  build tags, but I did not verify it compiles under another GOOS.
- **The causal reproduction behind AC-QUP-008's run-wide Gap.** I did not attempt
  to reproduce the lead's three `moai todo` writes and observe the digest change.
  My controlled window is clean; the run-wide attribution remains inference,
  exactly as `progress.md §H` records.
- **Coverage.** Not measured. No production line changed, so a package coverage
  delta would attribute a number to code this card did not author, and the 85%
  package gate has nothing to bind to on a test-only diff. A deliberate omission
  — but an omission, so the Craft score does not rest on a coverage figure.
- **The remaining `moai todo` verbs against a legacy layout** (`next`, `done`,
  `drop`, `relate`, `export-json`). Not exercised, by me or by the card. Already
  named in the verdict's Gaps.
- **`plan.md`, `plan-audit.md`, `plan-audit-iter2.md`** were not audited. The
  plan-phase verdict is the plan-auditor's, and re-litigating it is outside this
  audit's scope.
- **Sibling-lane interference during my runs.** I did not check for concurrent
  lanes while measuring. My timings (4.3s / 1.5s / 146.7s) are wall-clock figures
  under unknown load; no verdict of mine depends on a timing.

This Gaps section is intended to be complete. Anything not listed here, I
observed.

---

## Residual-risk

- **F1's remedy is a text edit, so it is the kind that gets deferred and then
  forgotten.** The verdict is committed. Once the card reaches `done`, nothing
  schedules the correction, and the false sentence outlives the card — which is
  precisely why F1 is classified blocking despite being one line.
- **A green here still does not prove the mechanism handles every legacy shape**,
  and my re-run does not extend it. Both of us exercised the same two fixtures. A
  partially-written document, an in-flight marker, or a directory the process
  cannot rename takes a branch neither of us entered.
- **My RED reproduction shares the implementer's mutation.** I applied the
  mutation `acceptance.md` names, so I proved the same thing they did, more
  strongly (independently, from the committed state). I did **not** search for a
  *second* mutation that ought to go red and does not — so the possibility of a
  criterion vacuous against some other severing remains open. F3 is one instance
  found by inspection rather than by systematic mutation.
- **F2's seam remains live.** If the lead accepts it as delivered, a future
  editor can silently disarm the AC-QUP-002 precondition and nothing in the
  toolchain will say so — confirmed, not assumed (`unparam` is silent on it).
- **The branch is unpushed and its worktree holds the only copy.** Not disposed
  by me; nothing committed by me. Disposal before the remote merge lands would
  destroy the work.

---

*Audited by `sync-auditor`. Read-only throughout: the one file I modified (the
mutation) was restored from backup and byte-identity to HEAD verified. No commit,
no push, no production edit.*
