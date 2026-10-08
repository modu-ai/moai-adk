# acceptance.md — SPEC-DISPATCH-INTEGRITY-001 (card t1595)

Stateless acceptance artifact. Every AC is binary-testable and carries the
two-cell structure (`verification-completeness.md` §2): a RED-now /
baseline-now cell observed on THIS tree (command + verbatim outcome + tree
SHA), and the milestone that flips or holds it. An AC whose defect measures
not-reproducible at the baseline is classified **regression-guard** and is
NOT recorded as a pass — the observation is the closure evidence.

Measurement protocol (every AC): env-scrubbed compound invocation
(`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …`),
`-run` selector scoped to the named tests, `-count=1` unless the AC says
otherwise, tree SHA recorded beside the outcome. A selector matching zero
tests is a failed measurement, not a pass.

Plan-phase baseline (tree 81786284e, 2026-10-08): (4) PASS, (5) PASS,
(8a) FAIL, (8b) FAIL, control PASS — verbatim in `plan.md` §B.

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

## AC-DI-003 — Defect (1): member not lease-eligible past unmerged hub sharer

**Given** recorded card `t1` (open, files crossing hub path X) and a
recorded bundle member `mk` whose files cross X, chained behind earlier
members,
**When** `factory next` selects for `mk`'s lane while `t1` is unmerged,
**Then** `mk` is not leased (skip or a refusal naming `t1`); after `t1`
reaches merged-local or merged-pr, `mk` becomes lease-eligible.
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
Baseline cell: observed PASS at 81786284e (`after="t1"`, refusal observed);
green path: stays green through close (M2 commits it as a regression
guard). Trace: REQ-DISPATCH-004.

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

## AC-DI-009 — Defect (8a): concurrent write between verification and rename refused

**Given** the fold seam injects a concurrent author write at the last
verification boundary,
**When** the fold write completes,
**Then** it returns an error and the file carries the concurrent author's
bytes — no silent overwrite. (`TestReviewFindingFoldConcurrentWrite`.)
Baseline cell: FAIL at 81786284e (`err=<nil>`); green path: M4 flips it
green. Trace: REQ-DISPATCH-008.

## AC-DI-010 — Defect (8b): a completed fold's index line survives

**Given** fold A paused at its archive-write boundary while fold B completes
fully on the same store,
**When** fold A resumes and terminates (complete or refuse),
**Then** B's completed index line is present in `MEMORY.md` or the archive —
never in neither. (`TestReviewFindingFoldInterleavedArchiveLoss`.)
Baseline cell: FAIL at 81786284e (`MEMORY=false archive=false`); green path:
M4 flips it green. Trace: REQ-DISPATCH-008.

## AC-DI-011 — Defect (8) concurrency discipline: race detector, repeated

**Given** M4 landed,
**When** the env-scrubbed `go test -count=5 -race` runs over the memory-fold
family (selector recorded in `progress.md` §E.2 BEFORE the run; the family
sweep covers `memory_fold_test.go`'s 14 tests plus
`memory_fold_wiring_test.go`),
**Then** all 5 iterations exit 0 with no data race reported. A single green
run does not satisfy this AC.

## AC-DI-012 — Re-measurement scope: every touched function's full test family

**Given** any production fix in this SPEC,
**When** the fix lands,
**Then** the owning package's full test family containing that function's
tests re-runs green in the same verification pass: `factory_bundle_test.go`
(13 tests), `factory_card_test.go` (10), `factory_card_pr_test.go` (15) +
`factory_card_pr_guard_test.go`, `memory_fold_test.go` (14) +
`memory_fold_wiring_test.go` — per family actually entered by the change,
with the family list recorded per milestone in §E.2.

## AC-DI-013 — Methodology control stays green

**Given** any verification pass of this card,
**When** `TestReviewFindingZoneExistingDotDot` runs,
**Then** it passes (committed as the control; baseline PASS at 81786284e).
A control regression means the verification methodology itself broke —
stop and re-derive before trusting any other AC result.

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
