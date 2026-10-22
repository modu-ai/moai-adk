# t1894 — lane-8 disk wait (2026-10-22)

Card t1894 re-leased by lane-8 on run tn8rpj (`moai factory next --card t1894 --run tn8rpj`); worktree re-entered at `3a163bf413`, branch `WT-fail-safe-serial`, tree clean.

## State read from disk

- `plan-hold-final.md`: the 8108 exception allowed exactly ONE correction round; iteration 3 returned FAIL 0.89 with blocker D28. By that contract FAIL = hold, and the ceiling disposition (another exception round / split per `plan-hold.md` option B / accept with codex-gate waiver) is operator-held.
- The leader's 2026-10-22 dispatch carries no disposition: `cmd` is "card body as written", and the closing line is the generic repair-test-push-draft-PR cycle.

## Wait

| Field | Value |
|---|---|
| Reason | Ceiling disposition for D28 is a keep-set (operator-held) call; the dispatch does not name one. A repair round without it would exceed the 8108 contract. |
| Whom | leader (relaying the operator) |
| Question | Which disposition: (i) one more exception round, (ii) split, or (iii) accept with waiver? |
| Recheck point | On the next awaken: re-read this directory and the leader's reply; if the disposition is named, proceed with plan repair via `manager-spec` (D28) then a fresh `plan-auditor`; Q7 still blocks only the autonomous Kickoff. |

No repair, spawn, or push was performed in this wake.

## Disposition received (2026-10-22, lane-8 wake)

Source: leader message (relaying operator delegation; jev MCP unavailable, leader judged directly). Decision: **(i) one additional exception round approved** — `manager-spec` repairs D28 (a sixth fence-pair-vs-base comparison number, OR disclosure of the joint-move limit in the three statements: acceptance.md AC-MJ-020 prose, design.md D-10 row + Unit bullet, checker docstring), then one fresh `plan-auditor`.
Condition: if a new D29+ blocker appears, NO further exception — switch to the split (`plan-hold.md` option B).
Reason given: small, precise candidate fix; independently reproduced by lane and auditor; same-day precedents of the same shape.
Lane action: proceed to plan repair; wait is closed.
