# A card left with an expired lease in a dead run (GitHub #1938)

`factory assign --reassign` is scoped to the card's current dispatch run
(`internal/cli/factory_card.go`, `factoryAssignReassign`), so it cannot move a
row out of a run that is no longer the active one. The formal path for that case
already exists and is the leader/operator verb:

```
moai factory runs --retire <dead-run> --rebind-to <live-run>
```

Measured (isolated sandbox, tests `factory_runs_rebind_expired_t2104_test.go`):

| Situation | Result |
|---|---|
| Dead run holds `t1` leased with an expired lease; live run holds no `t1` row | `t1` moves intact (state, owner, lease fields) to the live run; source run is `retired`; one row remains |
| Same, but `assign --run <live>` already created an `assigned` row for `t1` in the live run | consolidation is refused, naming the diverging columns (`lease_holder`, `lease_expires_at`); nothing changes; the source row stays |

So: use the rebind verb BEFORE assigning the card into the live run. Assigning
first is the step that creates the second row; the consolidation then refuses by
design (a live row is never silently discarded). A card whose live-run row was
already created needs the operator's decision on which row wins.

Not measured here: whether the two rows change `factory complete` or status
counting (see the card's progress record, Gaps). Where a landed card's rows are
closed, the AF-Flow chain in `af-flow-landed-card-close.md` applies; the
replacement behavior (AF-18, `f8ba9857bc`, local main only) is tracked under #1948
and is not restated here.
