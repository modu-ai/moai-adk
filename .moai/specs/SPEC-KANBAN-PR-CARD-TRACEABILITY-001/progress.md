# Progress — SPEC-KANBAN-PR-CARD-TRACEABILITY-001

## §E.1 Plan-phase Audit-Ready Signal

### Artifacts

Tier S set: `spec.md`, `plan.md`. Acceptance criteria are inline in `spec.md`
§D per the Tier S contract. `progress.md` is emitted at every tier and is not
counted in the Tier artifact total.

### Provenance

Split out of `SPEC-KANBAN-QUEUE-PR-SYNC-001` per audit finding D14
(`.moai/reports/t210/verdict.md`, iteration 1): that SPEC carried 19 leaf
requirements against a Tier M ceiling of 16, and its four doctrine requirements
are code-free and land on a different schedule from the tooling.

The split also makes an ordering explicit that was previously implicit inside a
single milestone sequence: the [HARD] cross-check clause is live and satisfied
by hand for the interval between this SPEC landing and the sibling shipping its
tooling. That interval is stated in `spec.md` §A rather than left for a reader
to discover.

Audit findings carried into this SPEC: D3 (REQ-002's report-and-re-decide
wording, which is the control against a de-facto-veto reading), D12 (each
acceptance criterion split into a mechanical half and a reviewer-judgement
half), D15 (mirror parity).

### SPEC ID check

```
$ ID="SPEC-KANBAN-PR-CARD-TRACEABILITY-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
PASS
```

No collision in `.moai/specs/`.

### Budget

```
$ grep -cE '^\*\*REQ-[0-9]{3}\*\*' .moai/specs/SPEC-KANBAN-PR-CARD-TRACEABILITY-001/spec.md
5
$ grep -cE '^### AC-[0-9]{3}' .moai/specs/SPEC-KANBAN-PR-CARD-TRACEABILITY-001/spec.md
4
```

5 leaf requirements and 4 acceptance criteria against a Tier S ceiling of 8 each.

### Audit iteration 2 — PASS (0.775 against the Tier S threshold of 0.75)

`.moai/reports/t210/verdict-2.md`. Two blocking findings, both lead-verified
rather than sent through a third round (the iteration ceiling is reached):

- **N3 — the split dropped a requirement.** The pre-split SPEC's REQ-3.4 (the
  template-mirror obligation) did not survive into this SPEC's requirement
  layer, which left AC-004 exercising nothing. Restored as **REQ-005**, with
  AC-004 mapped to it and a traceability table added. Tier S is at 5 of 8, so
  there was no budget obstacle — unlike the sibling, where the structurally
  correct fix for N1 would have breached 16/16.
- **N4 — a judgement inside a `**Mechanical.**` block.** AC-003's mechanical
  half asserted *"and the same section names all four traceability carriers"*
  while showing a grep covering only the first clause. Removed; the claim was
  already stated correctly in the judgement half, so nothing was lost. It was
  flagged blocking because it is a relapse of the D12 defect inside D12's own
  remedy — the reason this SPEC uses the split-AC pattern at all.

### Lint — verbatim

```
$ moai spec lint .moai/specs/SPEC-KANBAN-PR-CARD-TRACEABILITY-001/spec.md --json
[]
EXIT=0
```

### Zero baselines observed in this worktree (not relayed)

Each grep-shaped acceptance criterion is non-vacuous only against a zero
baseline. All three confirm at 0:

```
$ grep -c 'pre-dispatch PR cross-check' .claude/rules/moai/workflow/kanban-dispatch.md
0
$ grep -c 'confirms or withdraws' .claude/rules/moai/workflow/kanban-dispatch.md
0
$ grep -c 'PR title MUST carry the delivering card id' .claude/rules/moai/workflow/kanban-dispatch.md
0
```

The first and third were independently confirmed at 0 during the t210 audit;
`plan.md` §B requires re-running all three against the tree being edited rather
than citing these, because a baseline measured earlier proves nothing about the
tree at edit time.

### Gaps

- The reviewer-judgement halves of AC-001, AC-002, and AC-003 have no command.
  They are recorded as judgements and are not claimed as mechanical coverage.
- No CI check enforces the PR-title convention (`spec.md` §E). The obligation is
  doctrinal; mechanical enforcement is named as a plausible follow-up and is not
  specified.
- The `[HARD]`-marker grep in AC-001 is order-sensitive — it matches only when
  the marker precedes the phrase on the same line. `plan.md` §B records this so
  a correct clause is not failed by a grep that assumed the other order.

### Signal

`status: draft`. Awaiting audit, then Implementation Kickoff Approval. Lands
**before** `SPEC-KANBAN-QUEUE-PR-SYNC-001`.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
