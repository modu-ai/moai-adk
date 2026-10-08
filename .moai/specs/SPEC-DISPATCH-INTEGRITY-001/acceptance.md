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
body); the strengthened bodies' baselines are captured at M0; green path:
stays green through close (M2 commits them as regression guards). Trace:
REQ-DISPATCH-004.

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
**Then** it passes (committed as the control; baseline PASS at 81786284e).
A control regression means the verification methodology itself broke —
stop and re-derive before trusting any other AC result.

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

### EL-001 — TestReviewFindingFoldConcurrentWrite (RED)

- tree: 544462a8d (Go bytes identical to 81786284e — the plan commit
  touched only `.moai/specs`, verified via `git show --stat`)
- input: `owned-tests/owned_red_tests.go.txt` as of 5ae7d6ebc — the
  original overlay bodies these cells executed; the strengthened bodies
  now in the mirror are post-M0 intake forms (README § Assertion
  strength)
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
- input: the mirror as of 5ae7d6ebc (as EL-001 — original bodies)
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
- input: the mirror as of 5ae7d6ebc (as EL-001 — original bodies)
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
- input: the mirror as of 5ae7d6ebc (as EL-001 — original bodies)
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
