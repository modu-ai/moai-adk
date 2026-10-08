# acceptance.md — SPEC-DISPATCH-INTEGRITY-001 (card t1595)

Stateless acceptance artifact. Every AC is binary-testable and carries the
two-cell structure (`verification-completeness.md` §2): a RED-now /
baseline-now cell observed on THIS tree (command + verbatim outcome + tree
SHA), and the milestone that flips or holds it. An AC whose defect measures
not-reproducible at the baseline is classified **regression-guard** and is
NOT recorded as a pass — the observation is the closure evidence.

Measurement protocol (every AC): env-scrubbed compound invocation
(`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …`)
as the protocol form, `-run` selector scoped to the named tests, `-count=1`
unless the AC says otherwise, tree SHA recorded beside the outcome. A
selector matching zero tests is a failed measurement, not a pass.

Release-blocking and regression-guard baseline cells additionally carry the
verification-completeness §2.1 four elements — a single-invocation command
(the recorded command form; pipes, `&&`, `;`, and subshells live outside it),
its raw verbatim stdout, the exit code as its own field, and the tree SHA —
carried in the evidence ledger below and cited from the ACs. The ledger's
capture history (including the worktree-guard refusal of the `env -u … go
test` single-invocation substitute) is recorded at the ledger head.

Plan-phase baseline (tree 81786284e, 2026-10-08): (4) PASS, (5) PASS,
(8a) FAIL, (8b) FAIL, control PASS — verbatim in `plan.md` §B; the
four-element cells live in the evidence ledger below.

## AC-DI-001 — Baseline classification currency (gate for all fixes)

**Given** the card tree at HEAD H with the overlay set dropped per plan M0,
**When** the env-scrubbed targeted run executes per package,
**Then** `progress.md` §E.2 records a per-defect classification row
(defect, test, command, verbatim outcome, H) that matches THIS run; any
row whose cited tree is older than H's merge-base with the absorbed
integration ref is re-measured before any fix lands.
Green path: M0. Baseline cell: the 2026-10-08 §B table (81786284e).

## AC-DI-002 — Defect (3): duplicate member refused

**Given** a queue holding picked card `t2`,
**When** a bundle load runs with member list `t2 t2`,
**Then** the load exits non-zero with a duplicate-member refusal and the
factory record contains no bundle row for `t2`, no after relation
(self-dependency included), and no lane assignment.
Two-cell: RED-now observed in M1 on the current tree (expected LIVE — no
dedup exists at `runFactoryBundleLocked`); green path M1(b). If it measures
GREEN at baseline, reclassify regression-guard with the observation
recorded. Trace: REQ-DISPATCH-003.

## AC-DI-003 — Defect (1): member not lease-eligible past unmerged hub sharers (multi-hub)

**Given** two open recorded cards `t1` (files crossing hub path X) and `t2`
(files crossing hub path Y, disjoint from X), and a recorded bundle member
`mk` whose files cross BOTH X and Y,
**When** `factory next` selects for `mk`'s lane,
**Then** `mk` is not leased in every partially-blocked state — both
unmerged, `t1` merged with `t2` unmerged, AND `t2` merged with `t1`
unmerged (each sharer independently blocks; a configuration that consults
only one hub's predecessor fails this AC) — and `mk` becomes
lease-eligible only once BOTH `t1` and `t2` reach a merged state
(merged-local or merged-pr).
Two-cell: RED-now observed in M1; green path M1(b). Trace:
REQ-DISPATCH-001.

## AC-DI-004 — Defect (2): first member carries the hub constraint

**Given** a bundle head `h` whose files cross hub path X shared with open
recorded non-member `s`,
**When** the bundle load records `h`,
**Then** the load either refuses or records `h` with a hub-predecessor
constraint naming `s` — never an empty after that leaves `s` unchecked —
and selection honors the constraint.
Two-cell: RED-now observed in M1; green path M1(b). Trace:
REQ-DISPATCH-002.

## AC-DI-005 — Defect (4): nomination preserves stored after (regression guard)

**Given** `t3` picked with `HintAfter="t1"` (`t1` picked, unmerged),
**When** `factory next --card t3` runs,
**Then** the verb either refuses on the unmerged explicit predecessor or
leases `t3` with `HintAfter` still `"t1"` — never a recomputed hint.
Committed-test body duty (D13): the regression guard observes BOTH the
`runFactory` error AND the resulting card state — a two-arm assertion:
(A) the refusal NAMES the unmerged predecessor (`t1`) and `t3` stays
picked with `HintAfter="t1"` — a nil error, or any refusal that does not
name the predecessor (the injected `nomination unavailable` mutant class),
fails, because the dependency check provably did not run; (B) the
positive control `TestReviewFindingNominatedLeasesAfterPredecessorMerges`
places `t1` at merged-pr and requires the nomination to actually lease
`t3` with `HintAfter="t1"`, proving the refusal comes from the dependency
check and not a dead path.
Baseline cell: EL-004 (observed PASS at 81786284e/544462a8d, original
body). The strengthened nomination body is a GREEN-at-adoption regression
guard, NOT a RED-now body: the codex gate executed it against HEAD and
observed PASS — that observation is recorded at its M0 ledger intake.
The RED-now classification belongs to the strengthened fold bodies and
the positive control, whose M0 RED/GREEN status is genuinely undetermined
until first compile+run. Green path: stays green through close (M2
commits them as regression guards). Trace: REQ-DISPATCH-004.

## AC-DI-006 — Defect (5): merged-pr predecessor releases successor (regression guard)

**Given** `t1` at the merged-pr state and `t2` picked with `HintAfter="t1"`,
**When** `factory next` runs for a registered lane,
**Then** `t2` is leased.
Baseline cell: observed PASS at 81786284e (`leased="t2"`); green path: stays
green through close (M2 commits it as a regression guard). Trace:
REQ-DISPATCH-005.

## AC-DI-007 — Defect (6): no-record arm skips without aborting the pass

**Given** queue-picked card `nb` with no record row whose hub path is
shared with unmerged `s`, AND a second ready card `r` for the lane,
**When** `factory next` runs,
**Then** selection skips `nb` (no record created, no claim) and progresses
`r` — the verb does not error the whole pass.
Two-cell: RED-now observed in M2; green path M2 (expected regression
guard — pre-filter present). Trace: REQ-DISPATCH-006.

## AC-DI-008 — Defect (7): retry validates the lease before any remote mutation

**Given** card `c` at the merging state with `LeaseHolder` = another lane
(or an expired lease held by the caller),
**When** the delivery retry runs from the ineligible lane,
**Then** the verb refuses naming the holder/expiry, and the fixture remote
observes zero mutation — no push, no pull request, no auto-merge request.
Two-cell: RED-now observed in M3; green path M3 (expected regression
guard). Trace: REQ-DISPATCH-007.

## AC-DI-009 — Defect (8a): concurrent write detected by the comparison that follows the probe

**Given** the fold seam injects a NON-cooperating concurrent author write
(plain `os.WriteFile` — it honors no lock) at the pinned
`orderProbe("bytes-done")` call site — which TODAY sits between the
current final byte comparison and the rename, and AFTER the fix sits
between that comparison and the ADDED final comparison. The post-fix
write geometry is: byte comparison (cmp1) → seam probe → NEW final byte
comparison → rename —
**When** the fold write completes,
**Then** the concurrent change injected at the seam is detected by the
byte comparison that FOLLOWS the injection — the new last comparison
before the rename — the fold returns an error, and the file carries the
concurrent author's bytes (no silent overwrite observable at that
boundary). "Final byte comparison" is a ROLE: in every acceptable
geometry, a byte comparison follows the probe and is the last check
before the rename. (`TestReviewFindingFoldConcurrentWrite`; evidence
ledger EL-001.)

Committed-test body duty (D12): the committed body asserts BOTH clauses
of the Then — `err != nil` AND the final file bytes equal the concurrent
author's content (`bytes.Equal` against `concurrent author's new
memory\n`) — plus, ideally, that the error is the change-detection error
rather than an unrelated failure. `err != nil` alone is mutant-passable
(a mutant that detects the change, overwrites with the fold output
anyway, and returns an error passes it). The mirror at
`owned-tests/owned_red_tests.go.txt` carries the strengthened body; the
strengthened body's RED-now is captured at M0.

Guarantee scope — residual risk, stated explicitly rather than
absolutized: the guarantee extends to the comparison that follows the
probe, not to the rename itself. An irreducible TOCTOU tail remains
between that comparison and the rename against a writer that bypasses
every protocol — the codex gate's standalone reproduction (a held
exclusive flock, with the non-cooperating write still overwritten by the
rename) is the empirical record of that tail. Eliminating it entirely
would require the concurrent writer to honor the shared write protocol,
which nothing can force on a raw `os.WriteFile` author. The cross-process
lock (REQ-DISPATCH-008) closes the tail for cooperating writers; the
follow-the-probe detection is the defense for the rest. Baseline cell:
EL-001 (RED, exit 1, trees 81786284e/544462a8d); green path: M4 flips it
green. Trace: REQ-DISPATCH-008.

## AC-DI-010 — Defect (8b): a waiting fold completes without losing a completed fold's line

**Given** fold A running against the store holds the cross-process store
lock across its whole write transaction, and fold B — a SEPARATE process
with its own lock acquisition — starts for a different card while A holds
the lock (the committed test observes B's wait and completion from B's own
process result; B never writes inside A's window),
**When** A's transaction completes and releases the lock,
**Then** B acquires the lock, completes normally against the post-A store,
and after BOTH folds terminate every completed fold's index line is
present in `MEMORY.md` or the archive exactly once — no line lost, no
duplicate.

Committed-test note: `TestReviewFindingFoldInterleavedArchiveLoss` is
RE-AUTHORED to this serialized shape — M0 owns the re-authoring and its
RED-now measurement (the canonical source is otherwise still the
synchronous in-callback form; M4 delivers the lock that flips it green).
The re-authored body carries the D12 strengthened-assertion duty — every
completing fold's retention claim is asserted on the resulting file
CONTENT, not on an error value alone. The
original body ran B synchronously inside A's seam window — under the
mandated lock B can only wait or refuse there, so that criterion was
impossible under the design it accompanies (verification-completeness §2,
the impossible-red direction) and pinned the DEFECTIVE concurrency
semantics. Do NOT weaken the lock span to make the old body pass.
Two-cell: the original-body RED — the archive-preceding snapshot
overwrite, B's completed line vanishing into neither index — is EL-002
(RED, exit 1, trees 81786284e/544462a8d); the RE-AUTHORED body's RED-now
(under today's unlocked code B does not wait and the loss still occurs)
is captured in M0 as a four-element ledger entry, and M4 flips it green.
Trace: REQ-DISPATCH-008.

## AC-DI-011 — Defect (8) concurrency discipline: race detector, repeated

**Given** M4 landed,
**When** the env-scrubbed `go test -count=5 -race` runs over the memory-fold
family (selector recorded in `progress.md` §E.2 BEFORE the run; the family
sweep covers `memory_fold_test.go`'s 14 tests plus
`memory_fold_wiring_test.go`),
**Then** all 5 iterations exit 0 with no data race reported. A single green
run does not satisfy this AC. Trace: REQ-DISPATCH-008.

## AC-DI-012 — Re-measurement scope: every touched function's full test family

**Given** any production fix in this SPEC,
**When** the fix lands,
**Then** the owning package's full test family containing that function's
tests — and the families housing that function's CALLERS — re-runs green
in the same verification pass: `factory_bundle_test.go` (13 tests),
`factory_card_test.go` (10), `factory_nominate_test.go` (the
nomination-path family, `TestFactoryNextNominateLeasesNominee`),
`factory_card_pr_test.go` (15) + `factory_card_pr_guard_test.go`,
`memory_fold_test.go` (14) + `memory_fold_wiring_test.go` — per family
actually entered by the change, with the family list recorded per
milestone in §E.2 BEFORE the run (the list is re-derived from the touched
functions' callers at fix time, not only from this enumeration).

## AC-DI-013 — Methodology control stays green

**Given** any verification pass of this card,
**When** `TestReviewFindingZoneExistingDotDot` runs,
**Then** it passes (committed as the control; baseline PASS at
81786284e). A control regression means the verification methodology
itself broke — stop and re-derive before trusting any other AC result.
Baseline cell: EL-005 — **PENDING M0 capture** (the 2026-10-08 PASS
observation cannot be honestly completed to the §2.1 four elements; see
the ledger entry). Measurement input (D18): tracked at
`owned-tests/zone_control_test.go.txt` — added in the v0.1.5 closing
commit, the revision that binds; drop-in procedure
(`owned-tests/README.md`, target package `internal/hook`).

## Evidence ledger — baseline cells

Carrier for the §2.1 four-element baseline cells (single-invocation
command, raw verbatim stdout, exit code as its own field, tree SHA).
Capture history: the `env -u MOAI_KANBAN_ID … go test` single-invocation
scrub substitute was REFUSED by the worktree guard ("cannot be shown not
to be git") in both the round-1 audit and this repair session, so the
cells record the plain single-invocation form. Deviation, named not
silent: the session env carried the three scrub-target variables
(MOAI_KANBAN_ID, MOAI_KANBAN_LEAD_ADDR, MOAI_KANBAN_SETTINGS_INJECTED);
the four tests pin isolated fixture roots/tempdirs, route through no home
queue, and their outputs are byte-identical to the env-scrubbed records of
2026-10-08 (plan §B). Entries for the M1–M3 characterization tests and the
re-authored AC-DI-010 body are added at their first run (M0/M1–M3), each
in this four-element form.

Measurement-input binding (round-2, D10): the tree SHA alone does not make
a cell replayable — the four owned test sources were untracked drop-ins at
measurement time, and the codex gate demonstrated the hazard (`no tests to
run`, exit 0, replaying the selectors against the cited revision without
them). The four owned sources are tracked at `owned-tests/owned_red_tests.go.txt`
(drop-in procedure: `owned-tests/README.md`); every cell binds to the
tree SHA AND that input, and M0's re-measurement from the canonical
committed drop-in is the durable re-pin. Selector anchoring (D11): new
ledger captures use anchored selectors (`-run '^TestName$'`); the
EL-001..004 commands below remain verbatim as measured (unanchored —
over-selection only, and each recorded output shows exactly one test ran).

M0 capture form (run-phase, 2026-10-09): every EL-005..011 command is a
plain single invocation recorded verbatim; the output was captured by
redirecting to a file with `> file 2>&1` (the capture mechanism, not part
of the recorded command), so each verbatim block is the merged
stdout+stderr of that invocation and the exit code is the command's own.
Deviation, named not silent: as with the EL-001..004 captures, the session
env carried the three scrub-target variables; the compound
`unset … && go test` form ran for every M0 measurement and the plain form
is what is recorded. The EL-006..011 input binds to
`internal/cli/review_observation_test.go` as introduced by this SPEC's M0
commit — at measurement time (tree 5a91b5758) the file was the untracked
working-tree form whose bytes the M0 commit carries; replaying these
selectors against a checkout of that commit or later is the durable replay
path, against 5a91b5758 itself it is not (the D10 hazard, self-fulfilled
for the M0-introduced inputs).

### EL-001 — TestReviewFindingFoldConcurrentWrite (RED)

- tree: 544462a8d (Go bytes identical to 81786284e — the plan commit
  touched only `.moai/specs`, verified via `git show --stat`)
- input: `owned-tests/owned_red_tests.go.txt` as of e725633e0 — the
  original overlay bodies these cells executed (the mirror's first
  tracked revision); the strengthened bodies now in the mirror are
  post-M0 intake forms (README § Assertion strength)
- command: `go test ./internal/cli -run 'TestReviewFindingFoldConcurrentWrite' -count=1 -v`
- exit code: 1
- stdout (verbatim):

```
=== RUN   TestReviewFindingFoldConcurrentWrite
    review_observation_test.go:31: err=<nil> final bytes="fold output\n"
    review_observation_test.go:32: concurrent update between recheck and rename lost without refusal
--- FAIL: TestReviewFindingFoldConcurrentWrite (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.295s
FAIL
```

### EL-002 — TestReviewFindingFoldInterleavedArchiveLoss (RED — original body)

- tree: 544462a8d (as EL-001)
- input: the mirror as of e725633e0 (as EL-001 — original bodies)
- command: `go test ./internal/cli -run 'TestReviewFindingFoldInterleavedArchiveLoss' -count=1 -v`
- exit code: 1
- stdout (verbatim):

```
=== RUN   TestReviewFindingFoldInterleavedArchiveLoss
    review_observation_test.go:46: fold A err=memory fold: MEMORY.md changed since the plan was computed — aborting without writing; completed fold B line in MEMORY=false archive=false
    review_observation_test.go:47: completed fold B's line disappeared from both indexes after fold A resumed
--- FAIL: TestReviewFindingFoldInterleavedArchiveLoss (0.05s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.684s
FAIL
```

Note: this cell pins the ORIGINAL test body (AC-DI-010's committed-test
note). The re-authored body's RED-now entry is captured in M0.

### EL-003 — TestReviewFindingMergedPRPredecessor (PASS — regression guard)

- tree: 544462a8d (as EL-001)
- input: the mirror as of e725633e0 (as EL-001 — original bodies)
- command: `go test ./internal/cli -run 'TestReviewFindingMergedPRPredecessor' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingMergedPRPredecessor
    review_observation_test.go:9: predecessor=merged-pr; factory next leased="t2"
--- PASS: TestReviewFindingMergedPRPredecessor (3.74s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	4.887s
```

### EL-004 — TestReviewFindingNominatedOverwritesDependency (PASS — regression guard)

- tree: 544462a8d (as EL-001)
- input: the mirror as of e725633e0 (as EL-001 — original bodies)
- command: `go test ./internal/cli -run 'TestReviewFindingNominatedOverwritesDependency' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingNominatedOverwritesDependency
    review_observation_test.go:82: err=factory next: predecessor card not merged: t1 has not reached merged-local (git-flow) or merged-pr (github-flow) t3 state=picked after="t1"; original t1 state=picked
--- PASS: TestReviewFindingNominatedOverwritesDependency (3.06s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	4.245s
```

### EL-005 — TestReviewFindingZoneExistingDotDot (baseline PASS — measured M0)

- tree: 5a91b5758 (M0 intake, run-phase; the pending placeholder this
  entry replaces was stated AS pending, never as a pass — the 2026-10-08
  PASS observation behind plan §B's interim pointer could not satisfy
  §2.1's four elements, and this cell is the measured replacement)
- input: `owned-tests/zone_control_test.go.txt` as of e7c0e4791 (the
  revision that added the tracked mirror — the named commit, literal SHA
  stamped per the iter-7 advisory); the committed drop-in
  `internal/hook/review_observation_test.go` enters the branch in this
  SPEC's M0 commit and is the replayable form from that commit onward
- command: `go test ./internal/hook -run '^TestReviewFindingZoneExistingDotDot$' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingZoneExistingDotDot
    review_observation_test.go:35: OS physical target contents="safe"
    review_observation_test.go:42: tool=Write decision="deny" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=zone_dir/secret.md"
    review_observation_test.go:42: tool=Bash decision="deny" reason="HARNESS_FROZEN_PROTECTED_ZONE_VIOLATION: harness-learner category=probe_zone route=human next=return-blocker-report path=secret.md"
--- PASS: TestReviewFindingZoneExistingDotDot (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.642s
```

### EL-006 — TestReviewFindingNominatedOverwritesDependency (M0 intake — strengthened body, GREEN-at-adoption)

- tree: 5a91b5758 (M0 intake; the strengthened two-arm body's first
  compile+run on THIS tree — re-affirms the codex gate's observed PASS
  at HEAD, AC-DI-005's recorded classification)
- input: `internal/cli/review_observation_test.go` (this SPEC's M0
  commit — the strengthened D13/D15 body, gofmt-normalized from the
  mirror's post-e725633e0 forms)
- command: `go test ./internal/cli -run '^TestReviewFindingNominatedOverwritesDependency$' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingNominatedOverwritesDependency
    review_observation_test.go:62: err=factory next: predecessor card not merged: t1 has not reached merged-local (git-flow) or merged-pr (github-flow) t3 state=picked after="t1"; original t1 state=picked
--- PASS: TestReviewFindingNominatedOverwritesDependency (1.84s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.733s
```

### EL-007 — TestReviewFindingNominatedLeasesAfterPredecessorMerges (M0 intake — positive control, GREEN-at-adoption)

- tree: 5a91b5758 (M0 intake; the D13 positive control's FIRST
  compile+run — its M0 RED/GREEN status was genuinely undetermined until
  this capture, per AC-DI-005)
- input: `internal/cli/review_observation_test.go` (this SPEC's M0 commit)
- command: `go test ./internal/cli -run '^TestReviewFindingNominatedLeasesAfterPredecessorMerges$' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingNominatedLeasesAfterPredecessorMerges
    review_observation_test.go:99: positive control: err=<nil> t3 state=leased after="t1"
--- PASS: TestReviewFindingNominatedLeasesAfterPredecessorMerges (2.07s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.094s
```

### EL-008 — TestReviewFindingFoldConcurrentWrite (RED — strengthened body, this tree)

- tree: 5a91b5758 (M0 intake; the strengthened D12 body's first compile+run
  on THIS tree — re-affirms the original-body RED of EL-001 and extends it:
  all three clauses fire)
- input: `internal/cli/review_observation_test.go` (this SPEC's M0 commit)
- command: `go test ./internal/cli -run '^TestReviewFindingFoldConcurrentWrite$' -count=1 -v`
- exit code: 1
- stdout (verbatim):

```
=== RUN   TestReviewFindingFoldConcurrentWrite
    review_observation_test.go:126: err=<nil> final bytes="fold output\n"
    review_observation_test.go:132: concurrent update between recheck and rename lost without refusal
    review_observation_test.go:135: concurrent author's bytes were not preserved after refusal; final = "fold output\n"
--- FAIL: TestReviewFindingFoldConcurrentWrite (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.762s
FAIL
```

### EL-009 — TestReviewFindingFoldInterleavedArchiveLoss (RED — RE-AUTHORED serialized body)

- tree: 5a91b5758 (M0; the re-authored AC-DI-010 body's RED-now —
  fold B a separate OS process completing inside fold A's held-open
  transaction window; B's completed line lands in neither index and
  A's line is DUPLICATED across both — the loss plus the duplicate the
  exactly-once clause forbids; the ordering clause shows B's process
  finished before A's transaction closed)
- input: `internal/cli/review_observation_test.go` (this SPEC's M0 commit)
- command: `go test ./internal/cli -run '^TestReviewFindingFoldInterleavedArchiveLoss$' -count=1 -v`
- exit code: 1
- stdout (verbatim):

```
=== RUN   TestReviewFindingFoldInterleavedArchiveLoss
    review_observation_test.go:215: fold A err=memory fold: MEMORY.md changed since the plan was computed — aborting without writing; fold B err=<nil> in-window=true; line t9003 in MEMORY=false archive=false; line t9001 in MEMORY=true archive=true
    review_observation_test.go:225: fold B's process finished at 2026-10-09 04:29:35.891342 +0900 KST m=+0.146701959, before fold A's transaction closed at 2026-10-09 04:29:35.893391 +0900 KST m=+0.148751584 — it did not wait for the store lock
    review_observation_test.go:228: completed fold B's line is present 0 times across the indexes after both folds terminated, want exactly 1
    review_observation_test.go:231: completed fold A's line is present 2 times across the indexes after both folds terminated, want exactly 1
--- FAIL: TestReviewFindingFoldInterleavedArchiveLoss (0.06s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.359s
FAIL
```

### EL-010 — TestReviewFindingFoldArchiveConcurrentWrite (RED — arch-coverage instrument, archive write path)

- tree: 5a91b5758 (M0; the iter-7 arch-coverage debt's archive-write
  instrument: the fold's archive append — the guard=nil write — renamed
  over a non-cooperating concurrent author's bytes with no refusal)
- input: `internal/cli/review_observation_test.go` (this SPEC's M0 commit)
- command: `go test ./internal/cli -run '^TestReviewFindingFoldArchiveConcurrentWrite$' -count=1 -v`
- exit code: 1
- stdout (verbatim):

```
=== RUN   TestReviewFindingFoldArchiveConcurrentWrite
    review_observation_test.go:267: fold A err=<nil> archive="---\nname: a\ndescription: d\ntype: reference\n---\n- [a](feedback_a.md) — one\n- [b](feedback_b.md) — two\n- [c](feedback_c.md) — three\n- [t9001 alpha card — done, merged](project_card_t9001_alpha.md) — merged; detail lives in the topic file\n"
    review_observation_test.go:269: the archive rename published over a concurrent author's bytes without refusal
    review_observation_test.go:272: concurrent author's archive bytes were not preserved; archive = "---\nname: a\ndescription: d\ntype: reference\n---\n- [a](feedback_a.md) — one\n- [b](feedback_b.md) — two\n- [c](feedback_c.md) — three\n- [t9001 alpha card — done, merged](project_card_t9001_alpha.md) — merged; detail lives in the topic file\n"
--- FAIL: TestReviewFindingFoldArchiveConcurrentWrite (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.834s
FAIL
```

### EL-011 — TestReviewFindingFoldGuardArchiveChange (RED — arch-coverage instrument, guard post-probe window)

- tree: 5a91b5758 (M0; the iter-7 arch-coverage debt's guard instrument —
  the codex gate's data-loss mutant reproduced through the fold verb: a
  non-cooperating author stripped the appended line from the archive after
  the guard's byte comparison; MEMORY.md renamed over it and the line is
  present in NEITHER index)
- input: `internal/cli/review_observation_test.go` (this SPEC's M0 commit)
- command: `go test ./internal/cli -run '^TestReviewFindingFoldGuardArchiveChange$' -count=1 -v`
- exit code: 1
- stdout (verbatim):

```
=== RUN   TestReviewFindingFoldGuardArchiveChange
    review_observation_test.go:321: fold A err=<nil>; card line in MEMORY=false archive=false
    review_observation_test.go:323: MEMORY.md renamed over a concurrent archive change the guard no longer sees
    review_observation_test.go:329: the folded card's line is present 0 times across the indexes after the concurrent archive change, want exactly 1
--- FAIL: TestReviewFindingFoldGuardArchiveChange (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.748s
FAIL
```

### EL-012 — TestReviewFindingBundleDuplicateMemberRefused (RED — defect 3 LIVE)

- tree: 345eb6483 (M1; the M0 commit's production bytes, the M1
  characterization uncommitted in the working tree — the test enters the
  branch in this SPEC's M1 commit)
- command: `go test ./internal/cli -run '^TestReviewFindingBundleDuplicateMemberRefused$' -count=1 -v`
- exit code: 1
- stdout (verbatim):

```
=== RUN   TestReviewFindingBundleDuplicateMemberRefused
    review_observation_test.go:401: duplicate member load: err=stale card version: card t1 is at version 2, request expected 1
    review_observation_test.go:406: refusal does not name the duplicate member: stale card version: card t1 is at version 2, request expected 1
--- FAIL: TestReviewFindingBundleDuplicateMemberRefused (1.26s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.559s
FAIL
```

### EL-013 — TestReviewFindingBundleHeadHubConstraint (baseline NOT-REPRODUCED — defect 2 regression guard)

- tree: 345eb6483 (as EL-012)
- command: `go test ./internal/cli -run '^TestReviewFindingBundleHeadHubConstraint$' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingBundleHeadHubConstraint
    review_observation_test.go:439: head hub load: err=predecessor card not merged: t2 has not reached merged-local (git-flow) or merged-pr (github-flow)
    review_observation_test.go:441: the load refused — AC-DI-004 admits refusal as the constraint
--- PASS: TestReviewFindingBundleHeadHubConstraint (0.92s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.881s
```

### EL-014 — TestReviewFindingBundleMultiHubMemberWaits (baseline NOT-REPRODUCED — defect 1 regression guard)

- tree: 345eb6483 (as EL-012)
- command: `go test ./internal/cli -run '^TestReviewFindingBundleMultiHubMemberWaits$' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingBundleMultiHubMemberWaits
=== RUN   TestReviewFindingBundleMultiHubMemberWaits/both_unmerged
    review_observation_test.go:497: member lease: ""
=== RUN   TestReviewFindingBundleMultiHubMemberWaits/t1_merged_t2_unmerged
    review_observation_test.go:497: member lease: ""
=== RUN   TestReviewFindingBundleMultiHubMemberWaits/t2_merged_t1_unmerged
    review_observation_test.go:497: member lease: ""
=== RUN   TestReviewFindingBundleMultiHubMemberWaits/both_merged
    review_observation_test.go:497: member lease: "t3"
--- PASS: TestReviewFindingBundleMultiHubMemberWaits (8.01s)
    --- PASS: TestReviewFindingBundleMultiHubMemberWaits/both_unmerged (2.01s)
    --- PASS: TestReviewFindingBundleMultiHubMemberWaits/t1_merged_t2_unmerged (1.63s)
    --- PASS: TestReviewFindingBundleMultiHubMemberWaits/t2_merged_t1_unmerged (1.58s)
    --- PASS: TestReviewFindingBundleMultiHubMemberWaits/both_merged (2.78s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	8.994s
```

### EL-015 — TestReviewFindingBundleDuplicateMemberRefused (GREEN post-fix — defect 3 closed)

- tree: 345eb6483 working tree with the M1 fix applied uncommitted; the fix
  and this cell land together in the M1 commit (the replayable form from
  that commit onward)
- command: `go test ./internal/cli -run '^TestReviewFindingBundleDuplicateMemberRefused$' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingBundleDuplicateMemberRefused
    review_observation_test.go:401: duplicate member load: err=duplicate member t1 in the bundle member list
--- PASS: TestReviewFindingBundleDuplicateMemberRefused (0.81s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.775s
```

### EL-016 — TestReviewFindingNoRecordArmSkipsBlockedCandidate (baseline NOT-REPRODUCED — defect 6 regression guard)

- tree: 96f392d06 (M1 commit's production bytes, the M2 characterization
  uncommitted — the test enters the branch in the M2 commit)
- command: `go test ./internal/cli -run '^TestReviewFindingNoRecordArmSkipsBlockedCandidate$' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingNoRecordArmSkipsBlockedCandidate
    review_observation_test.go:529: lane lease: "t3"
--- PASS: TestReviewFindingNoRecordArmSkipsBlockedCandidate (2.26s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.424s
```

### EL-017 — TestReviewFindingNominatedOverwritesDependency + TestReviewFindingMergedPRPredecessor (M2 re-affirm — defects 4/5 regression guards)

- tree: 96f392d06 (M2)
- command: `go test ./internal/cli -run '^TestReviewFindingNominatedOverwritesDependency$|^TestReviewFindingMergedPRPredecessor$' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingMergedPRPredecessor
    review_observation_test.go:42: predecessor=merged-pr; factory next leased="t2"
--- PASS: TestReviewFindingMergedPRPredecessor (2.51s)
=== RUN   TestReviewFindingNominatedOverwritesDependency
    review_observation_test.go:62: err=factory next: predecessor card not merged: t1 has not reached merged-local (git-flow) or merged-pr (github-flow) t3 state=picked after="t1"; original t1 state=picked
--- PASS: TestReviewFindingNominatedOverwritesDependency (1.94s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	5.630s
```

### EL-018 — TestReviewFindingMergingRetryValidatesLeaseBeforeRemote (baseline NOT-REPRODUCED — defect 7 regression guard)

- tree: 271d71ab9 (M2 commit's production bytes, the M3 characterization
  uncommitted — the test enters the branch in the M3 commit)
- command: `go test ./internal/cli -run '^TestReviewFindingMergingRetryValidatesLeaseBeforeRemote$' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingMergingRetryValidatesLeaseBeforeRemote
=== RUN   TestReviewFindingMergingRetryValidatesLeaseBeforeRemote/foreign_lane_holder
    review_observation_test.go:567: merging retry: lane=lane-2 err=factory complete: refused — card t1 belongs to lane-1 (lease lane-1), not lane-2; a lane completes only its own card
=== RUN   TestReviewFindingMergingRetryValidatesLeaseBeforeRemote/expired_caller_lease
    review_observation_test.go:567: merging retry: lane=lane-1 err=factory complete: refused — card t1's merging lease held by lane-1 expired at 2026-09-01T00:00:00Z; the expiry must be collected before the delivery is retried
--- PASS: TestReviewFindingMergingRetryValidatesLeaseBeforeRemote (3.89s)
    --- PASS: TestReviewFindingMergingRetryValidatesLeaseBeforeRemote/foreign_lane_holder (1.83s)
    --- PASS: TestReviewFindingMergingRetryValidatesLeaseBeforeRemote/expired_caller_lease (2.06s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	5.161s
```

### EL-019 — TestReviewFindingFoldConcurrentWrite (GREEN post-M4 — defect 8a closed)

- tree: c0a0d7cbd working tree with the M4 fix applied uncommitted; the fix
  and this cell land together in the M4 commit (the replayable form from
  that commit onward)
- command: `go test ./internal/cli -run '^TestReviewFindingFoldConcurrentWrite$' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingFoldConcurrentWrite
    review_observation_test.go:126: err=memory fold: MEMORY.md changed since the plan was computed — aborting without writing final bytes="concurrent author's new memory\n"
--- PASS: TestReviewFindingFoldConcurrentWrite (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.137s
```

### EL-020 — TestReviewFindingFoldArchiveConcurrentWrite (GREEN post-M4 — arch-coverage debt, archive write path)

- tree: c0a0d7cbd working tree with the M4 fix applied uncommitted (as
  EL-019)
- command: `go test ./internal/cli -run '^TestReviewFindingFoldArchiveConcurrentWrite$' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingFoldArchiveConcurrentWrite
    review_observation_test.go:267: fold A err=memory fold: project_card_archive_2026_10.md changed since the plan was computed — aborting without writing archive="concurrent author's archive bytes\n"
--- PASS: TestReviewFindingFoldArchiveConcurrentWrite (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.137s
```

### EL-021 — TestReviewFindingFoldGuardArchiveChange (GREEN post-M4 — arch-coverage debt, guard post-probe recheck)

- tree: c0a0d7cbd working tree with the M4 fix applied uncommitted (as
  EL-019)
- command: `go test ./internal/cli -run '^TestReviewFindingFoldGuardArchiveChange$' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingFoldGuardArchiveChange
    review_observation_test.go:321: fold A err=memory fold: project_card_archive_2026_10.md changed since the plan was computed — aborting without writing; card line in MEMORY=true archive=false
--- PASS: TestReviewFindingFoldGuardArchiveChange (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.137s
```

### EL-022 — TestReviewFindingFoldInterleavedArchiveLoss (GREEN post-M4 — defect 8b closed, cross-process serialization)

- tree: c0a0d7cbd working tree with the M4 fix applied uncommitted (as
  EL-019); fold B's process made NO progress inside A's window (the bounded
  wait expired with in-window=false — the lock held it), then completed
  normally against the post-A store after A released; both completed folds'
  lines are present exactly once and B's exit was observed after A's
  transaction closed
- command: `go test ./internal/cli -run '^TestReviewFindingFoldInterleavedArchiveLoss$' -count=1 -v`
- exit code: 0
- stdout (verbatim):

```
=== RUN   TestReviewFindingFoldInterleavedArchiveLoss
    review_observation_test.go:215: fold A err=<nil>; fold B err=<nil> in-window=false; line t9003 in MEMORY=false archive=true; line t9001 in MEMORY=false archive=true
--- PASS: TestReviewFindingFoldInterleavedArchiveLoss (60.02s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	60.993s
```

### EL-023 — TestReviewFindingStoreLockIndependentOfTempDir (RED → GREEN — gate finding 1)

- RED tree: the M4 working tree at HEAD c0a0d7cbd with the temp-dir lock
  applied, the test uncommitted; GREEN tree: the same working tree with the
  in-store lock fix applied (the fix and the test land together in the M4
  commit)
- command (RED observation): `go test ./internal/cli -run '^TestReviewFindingStoreLockIndependentOfTempDir$' -count=1 -v`
- RED exit code: 1 — stdout (verbatim):

```
=== RUN   TestReviewFindingStoreLockIndependentOfTempDir
    review_observation_test.go:408: a different-TMPDIR locker entered the store's critical section while it was held (err=<nil>) — the lock path follows the process temp dir, not the store
--- FAIL: TestReviewFindingStoreLockIndependentOfTempDir (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	4.986s
```

- command (GREEN): the same anchored selector; exit code 0 — stdout
  (verbatim, the combined run with the finding-2 test):

```
=== RUN   TestReviewFindingStoreLockIndependentOfTempDir
--- PASS: TestReviewFindingStoreLockIndependentOfTempDir (0.70s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.132s
```

### EL-024 — TestReviewFindingAbandonedFoldReleasesStoreLock (RED → GREEN — gate finding 2, unix-only)

- trees: as EL-023 (RED against the lock without the abandonment-polled
  read; GREEN after `snapshotStoreBounded`)
- command (RED observation): `go test ./internal/cli -run '^TestReviewFindingAbandonedFoldReleasesStoreLock$' -count=1 -v`
- RED exit code: 1 — stdout (verbatim, the combined RED run with the
  finding-1 test; the fold's own stderr line carried in the merged
  capture):

```
=== RUN   TestReviewFindingAbandonedFoldReleasesStoreLock
memory fold-on-done: t9001: abandoned after 300ms — the store did not answer in time; the step will begin no write
    review_observation_fifo_unix_test.go:60: the store lock stayed held after the bounded fold was abandoned — the blocked worker pins the lock past its caller's timeout
--- FAIL: TestReviewFindingAbandonedFoldReleasesStoreLock (3.53s)
```

- command (GREEN): the same anchored selector; exit code 0 — stdout
  (verbatim, from the combined GREEN run):

```
=== RUN   TestReviewFindingAbandonedFoldReleasesStoreLock
memory fold-on-done: t9001: abandoned after 300ms — the store did not answer in time; the step will begin no write
--- PASS: TestReviewFindingAbandonedFoldReleasesStoreLock (0.57s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.132s
```

### EL-025 — TestReviewFindingAbandonedFoldReleasesLockOnApplyReads (RED → GREEN — post-report finding 1, unix-only)

- trees: RED against HEAD 2f000f923's bytes (only the initial snapshot
  cancellation-guarded); GREEN after `readFileBounded` threading; the test
  and the fix land together in the post-report repair commit
- command (RED observation): `go test ./internal/cli -run '^TestReviewFindingAbandonedFoldReleasesLockOnApplyReads$' -count=1 -v`
- RED exit code: 1 — stdout (verbatim, the combined RED run with the
  preview test):

```
=== RUN   TestReviewFindingAbandonedFoldReleasesLockOnApplyReads
memory fold-on-done: t9001: abandoned after 300ms — the store did not answer in time; the step will begin no write
    review_observation_fifo_unix_test.go:76: the store lock stayed held after the bounded fold was abandoned — a blocking read inside the apply pins the lock past its caller's timeout
--- FAIL: TestReviewFindingAbandonedFoldReleasesLockOnApplyReads (3.56s)
```

- command (GREEN): the same anchored selector; exit code 0 — stdout
  (verbatim, the combined GREEN run):

```
=== RUN   TestReviewFindingAbandonedFoldReleasesLockOnApplyReads
memory fold-on-done: t9001: abandoned after 300ms — the store did not answer in time; the step will begin no write
--- PASS: TestReviewFindingAbandonedFoldReleasesLockOnApplyReads (0.60s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.745s
```

### EL-026 — TestReviewFindingPreviewNeedsNoWriteAccess (RED → GREEN — post-report finding 2)

- trees: as EL-025
- command (RED observation): `go test ./internal/cli -run '^TestReviewFindingPreviewNeedsNoWriteAccess$' -count=1 -v`
- RED exit code: 1 — stdout (verbatim; long fixture paths elided with …
  for width, the refusal text verbatim):

```
=== RUN   TestReviewFindingPreviewNeedsNoWriteAccess
    review_observation_test.go:394: memory fold --card t9001 --dir …/001 exited with error: memory fold: open the store lock …/001/.moai-store-lock: open …/001/.moai-store-lock: permission denied
        stderr: Error: memory fold: open the store lock …/001/.moai-store-lock: open …/001/.moai-store-lock: permission denied
--- FAIL: TestReviewFindingPreviewNeedsNoWriteAccess (0.00s)
```

- command (GREEN): the same anchored selector; exit code 0 — stdout
  (verbatim):

```
=== RUN   TestReviewFindingPreviewNeedsNoWriteAccess
    review_observation_test.go:397: preview on a read-only store: store: …/001 (--dir)
        card: t9001
        archive index: project_card_archive_2026_10.md
        would file 1 line(s)
--- PASS: TestReviewFindingPreviewNeedsNoWriteAccess (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.955s
```

### EL-027 — TestReviewFindingAbandonedFoldExitsLockWait (RED → GREEN — post-report finding 3, the abandonment trio's third member)

- trees: as EL-025/EL-026; the RED state is the compile refusal against
  the one-arg `acquireFoldStoreLock` signature (the abandonment-aware
  acquisition the fix introduces did not exist): `go vet ./internal/cli`
  → `too many arguments in call to acquireFoldStoreLock have (string, nil) want (string)` — the test cannot compile, therefore cannot pass, on the
  pre-fix tree; the green path is the fix itself
- command (GREEN): `go test ./internal/cli -run '^TestReviewFindingAbandonedFoldExitsLockWait$' -count=1 -v`
- GREEN exit code: 0 — stdout (verbatim, the combined trio run):

```
=== RUN   TestReviewFindingAbandonedFoldExitsLockWait
--- PASS: TestReviewFindingAbandonedFoldExitsLockWait (0.53s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.379s
```

## Quality gates and closure

- TRUST 5: Tested (every AC above; 85%+ on touched packages per repo
  standard), Readable/Unified (gofmt, go vet clean on touched files),
  Secured (no new trust boundaries; the fold lock touches no credentials),
  Trackable (Conventional Commits carrying card id t1595 + SPEC id).
- Closure gates: all 8 defects carry a two-cell record; owned tests
  committed on the card branch; foreign-card tests absent from the branch;
  `progress.md` §E.2/§E.3 populated by run-phase; the completion report
  carries Claim/Evidence/Baseline-attribution/Gaps/Residual-risk with no
  unobserved claims.
