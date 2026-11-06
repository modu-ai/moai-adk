# A card left with an expired lease in a dead run (GitHub #1938)

Two refusals sit in front of an operator who finds a card in a dead run. The
first is `factory assign --reassign --run <dead>`: `resolveFactoryCardRun`
(`internal/cli/factory_card.go`) refuses an explicit `--run` whose row is not
`active`, and `factoryAssignReassign` / `ReassignCard` keep the move inside one
run id, so neither can move a row out of the run it sits in. The formal path for
that case is the leader/operator verb:

```
moai factory runs --retire <dead-run> --rebind-to <live-run>
```

Measured (isolated sandboxes; tests `internal/cli/factory_runs_rebind_expired_t2104_test.go`):

| Situation | Result |
|---|---|
| Dead run holds `t1` leased with an expired lease; live run holds no `t1` row | `t1` moves intact (state, owner, lease fields) to the live run; source run is `retired`; one row remains |
| Live run already holds an `assigned` `t1` row (assign-first) | the source row is abandoned with a `card.rebound` pointer event; the live row is untouched; the source run is `retired` |
| Two dead runs hold `t1` leased, each expired at a different time | the first rebind moves `t1`; the second abandons its row with a pointer; both dead runs are `retired`; one live row remains |
| The source row also differs from the live row on any compared column other than the dead lease (for example `stage`) | refused, naming the column; nothing changes |

Why the lease alone does not refuse: the consolidation has already established
that the source row holds no VALID lease (a live lease refuses the whole
operation first), so `lease_holder` / `lease_expires_at` on the source are dead
data. A source row whose only differences from the live row are those two
columns duplicates the live row, and the abandon-with-pointer keeps nothing
live behind. Before this change that divergence refused, which left the second
dead run active and unretirable (measured: `runs --retire <dead>` alone is
refused while the run holds an in-progress card).

## What an unconsolidated two-row state does (measured before the change)

State: expired `leased` `t1` in a dead run, `assigned` `t1` in the live run,
`card_dispatch` bound to the live run.

| Surface | Observed |
|---|---|
| `moai factory status` | prints both rows (`run=run-src state=leased ... lease=lane-9 expired ...` and `run=run-tgt state=assigned ...`); the card counts twice |
| completion gate row resolution (`RecordedCardRowReadonly`, via the `card_dispatch` binding) | resolves the live-run row (`run=run-tgt state=assigned`); the leftover row is never read |
| `moai factory runs --retire <dead>` | refused: the run holds one in-progress card |
| `factorymsg.ResolveActiveRun` without `--run` (what run-less verbs use) | `AMBIGUOUS_FACTORY`: the dead run is still `active` |

So the leftover row does not corrupt the completion gate; its harm is that the
dead run stays active, so every run-less verb is ambiguous and `status` shows
the card twice. With the change above one `--rebind-to` call removes it.

Hand-off for a state that predates this change (a dead run still active with a
leftover row): as the operator (`MOAI_FACTORY_ROLE=operator`), run
`moai factory decide <card> --choice abandon --run <dead>` (an expired-lease card
needs the call twice), then `moai factory runs --retire <dead>`. This is
irreversible; the `--rebind-to` route is preferred.

Where a landed card's rows are closed, the AF-Flow chain in
`af-flow-landed-card-close.md` applies; the replacement behavior (AF-18,
`f8ba9857bc`, local main only) is tracked under #1948 and is not restated here.
