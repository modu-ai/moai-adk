# Closing a card that landed in local main (af-flow)

Card t2071. Applies to a card the leader integrated into LOCAL `main` whose lane never walked
`factory stage` (factory row `state=leased`, `stage=-`), and which has not reached `origin/main`
yet because af-flow pushes only in the leader's batch.

## Why the three verbs refuse

All three refusals trace to one missing fact — no landing record on the card:

| Verb | Refusal | Cause |
|---|---|---|
| `moai factory rebind-evidence` | `cannot infer the audit phase from stage "" (state leased)` | no stage to read a verdict for, and no `todo landed` record to fall back on |
| `moai factory approve` | `records no evidence hash to bind` | the row's `evidence_sha` is empty until rebind succeeds |
| `moai todo done` | `leader approval receipt absent` | the receipt is bound to that evidence hash |

A bare `moai todo landed <card> --sha <merge>` is refused as unreachable, because its default ref
is `origin/main` and the merge sits on local `main`.

## Batch landing is the usual way a card gets here

Under af-flow the default landing is the leader's batch: one audited
candidate chain the operator fast-forwards into `main`
(`.claude/rules/moai/workflow/factory-dispatch-mechanics.md` § Batch landing —
the af-flow default). The batch manifest
(`.moai/reports/af-batch-<date>/manifest.json`) names each landed card's merge
SHA; that SHA is the `--sha` of the chain below, and the chain runs once per
landed row after `git merge-base --is-ancestor <tip> main` holds.

## Batch landing is the usual way a card gets here

Under af-flow the default landing is the leader's batch: one audited
candidate chain the operator fast-forwards into `main`
(`.claude/rules/moai/workflow/factory-dispatch-mechanics.md` § Batch landing —
the af-flow default). The batch manifest
(`.moai/reports/af-batch-<date>/manifest.json`) names each landed card's merge
SHA; that SHA is the `--sha` of the chain below, and the chain runs once per
landed row after `git merge-base --is-ancestor <tip> main` holds.

## The formal chain (leader path; no store edits)

A lane cannot run steps 3, 5 and 6: `todo landed`, `factory approve` and `todo done` are queue or
leader-path acts. Run them from the leader session in the primary checkout.

1. Find the delivering commit (read-only): `git log --oneline --grep 'merge: <card>' main`.
2. Check it is on local `main` (`git merge-base --is-ancestor <sha> main` exits 0) and not yet on
   origin (the same command against `origin/main` exits 1).
3. Record the landing against the local branch: `moai todo landed <card> --sha <sha> --ref main`.
4. Re-bind the evidence from that record: `moai factory rebind-evidence <card> --run <run> --reason "<why>"`.
   The recorded commit is accepted only when it is reachable from the record's ref.
5. Issue the receipt: `moai factory approve <card> --run <run>`. The run must be the one the done
   gate reads (`moai factory status` shows it); a receipt for another run is refused at `todo done`.
6. Close: `moai todo done <card>`.

## When `rebind-evidence` still refuses after a landing is recorded

- `recorded landing refused: observed <time>, before this run engaged the card (<time>) — it belongs
  to an earlier issuance`: the landing record is stored per card id, not per run, so a reissued card
  still carries the previous issuance's record, and a record observed before this run first engaged
  the card is refused. Record the landing again (step 3): `todo landed` replaces the record with a
  fresh observation time, and `rebind-evidence` then accepts it. This is pinned by
  `TestAFFlowEarlierIssuanceRecordIsRefusedAndRecordingAgainHeals`.
- `the card's landing record names no delivering commit`: a record exists but was made without
  `--sha`; record it again with `--sha <delivering commit>` (step 3).
- `recorded landing commit refused: ...`: the commit is not reachable from the record's ref; for a
  card merged into local `main` the ref must be `main` (step 3), not the default `origin/main`.

## Worked example (read-only facts, not executed by the lane)

Card t2055, run `to0ck1`, state `leased`, stage `-` (`moai factory status`). Its local-main merge is
`7f1ab8d42e` (`[t2055] merge: lane landing closes the card at the local main merge ...`);
`git merge-base --is-ancestor 7f1ab8d42e main` succeeds and the same check against `origin/main`
fails. So the leader runs:

```text
moai todo landed t2055 --sha 7f1ab8d42e --ref main
moai factory rebind-evidence t2055 --run to0ck1 --reason "af-flow: merged into local main, no stage walked"
moai factory approve t2055 --run to0ck1
moai todo done t2055
```

## Verification

`TestAFFlowLocalMainCardClosesThroughFormalVerbs` in `internal/cli/factory_af_flow_close_test.go`
runs steps 3-6 on an isolated fixture (leased, stage-less row; merge on local `main`, origin one
commit behind) and passes at the candidate head. `TestAFFlowRefusalsNameTheLandingStep` pins the two
refusals that now name the missing step. The real t2055 chain above was not executed by this lane.
