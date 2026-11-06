# t2116 progress — owner re-lease of an expired lease on a staged card

Card t2116 (P0, Class B, no SPEC). Lane lane-21, run to2bbk. Tree `.moai/worktrees/projects/moai-adk-go/t2116`, branch `WT-plan-run-sync`. Created from `9cf8a1fc14`.

## Commit chain

| SHA | Content |
|---|---|
| (fast-forward) `27adc925ef` | t2091's tip absorbed as the leader asked (fast-forward from `9cf8a1fc14`) |
| `1cacc0634b` | `test(homestate,cli): RED` — two test files, red by design |
| `66ec369381` | `fix(homestate): the owner re-leases an expired lease on a staged card` — one predicate widened, one existing case moved |
| merge `5c9e2cb7b780c64075c057525b73edacf074b53f` | local `main` `b00c35d103` (t2098 landing) absorbed, so the branch descends from main |

## What this lands on main besides t2116 (leader requested the record)

Because the t2091 tip was absorbed, landing t2116 also puts these cards' content on main:

| Card | State in the record | Audit file in the tree | audited_sha | Verdict |
|---|---|---|---|---|
| t2071 | merge-ready | `.moai/reports/t2071/sync-audit.md` | `7131da04c1c0acfc9701d7b348675fed87b8a1fb` | PASS-WITH-DEBT 0.90 |
| t2091 | merge-ready | `.moai/reports/t2091/sync-audit.md` | `0ad78fd95ec7421e58d3c9bd5d24136ab823e339` | PASS 0.92 |

Also in the diff against `9cf8a1fc14`: `.moai/docs/af-flow-landed-card-close.md`, `internal/cli/factory_card_rebind.go`, `internal/cli/todo_landed.go` and their tests (t2091). The leader closes the two cards as landed after t2116 lands; this lane closes neither.

## Evidence

1. Defect reproduced in real terms: `factory next --card` on a card in `sync` whose lease expired answered `factory next: refused owned: the card is in flight at sync (holder lane-21)` (measured on t2065, 05:01Z lease). Source: `homestate.IsReacquirableExpiredLease` accepted state `leased` only.
2. RED: `go test ./internal/homestate/ -run T2116` — `TestIsReacquirableExpiredLeaseCoversStagedStates_T2116` and `TestReacquireExpiredLeaseStagedStates_T2116/{plan,plan-audit,run,sync,sync-audit,merge-ready}` failed at the intended assertions ("not reacquirable, want reacquirable", "reacquire … = (false, <nil>), want a re-lease"); the negative controls passed (`.moai/reports/t2116/red-homestate.txt`). `go test ./internal/cli/ -run TestT2116` failed with the incident's own text for all six states, `…LeavesStagedWorktreeUntouched` included (`red-cli.txt`). A first homestate run was a fixture error (UNIQUE constraint, a second insert of the same card) and is not counted as RED.
3. GREEN: homestate focused set 33 PASS / 0 FAIL (`green-homestate-focused.txt`); CLI narrow set `TestT2116*` + `TestT2013{Owner,Other,Second,Renew,Adopt}*` 32 PASS / 0 FAIL / 0 SKIP, `ok … 126.257s` (`green-cli-narrow.txt`).
4. At the merged HEAD `5c9e2cb7b7`: `go test ./internal/homestate/ -run 'T2116|Reacquire|IsReacquirable'` 22 PASS / 0 FAIL, `ok … 52.279s` (`merged-homestate.txt`).

## Gaps (not observed)

- The whole `internal/homestate` package and the broad `internal/cli` pattern (`T2013|T2116|Reacq|ExpiredLease|FactoryNext|NextCard`) both hit the test timeout (270 s / 280 s) with no failure line; they are not claimed as passes. The package-level proof is CI's.
- The CLI tests were run before the `main` absorption (`green-cli-narrow.txt`); at the merged HEAD only the homestate set was re-run. The merge touched `internal/cli/factory_release_route.go` and config files, not the re-lease path.
- No real-world run: no live expired staged card was re-leased with the new binary (the installed `moai` predates this change).
- Windows and the full multi-OS matrix: CI.

## Residual risk

- A staged card re-leased to `leased` keeps its stage; the lane must resume with `moai factory stage <card> <that stage>`. No test drives the `factory stage` verb after the CLI re-lease; the homestate test drives the transition edge.
- `merging` is excluded on purpose (expiry there means an interrupted merge, state `blocked`); a card stuck in `merging` still needs the leader.

## Audit (first audit, one `codex_audit` call)

- Claude sync-auditor at `5c9e2cb7b780c64075c057525b73edacf074b53f`: **PASS 0.92** (Functionality 95 / Security 95 / Craft 88 / Consistency 88), 0 blocking, 5 minor/notes. File `.moai/reports/t2116/sync-audit.md`, committed force-added on this branch after the audit; `audited_sha` stays `5c9e2cb7b7`, the code tree is unchanged by that commit.
- Codex receipt `rcpt-49ed7c1e5fb157b9fbd18305` (one call, native mode, no `delta_base`, merge base `b00c35d103`, head `5c9e2cb7b7`, codex result pass with no findings).
- The auditor's own re-runs at that SHA: homestate 22/22 PASS, CLI 32/32 PASS (14 tests + 18 subtests), no skips; four mutants (allow every state; drop the owner check; re-admit `merging`; the pre-fix `leased`-only predicate) were all caught.
- Minor findings, none blocking: (F1) `factory-dispatch-mechanics.md:30` and its template copy still describe a `leased` card only (card t2082's rewrite); (F2) a Codex lane on an expired `merge-ready` card is still refused without a write, but with refusal `backend-skip` instead of `owned`; (F3) the new state check in the predicate repeats what `LeaseExpired` already implies; (F4) no test covers a staged card through the MCP path or the `factory stage` resume after the CLI re-lease; (F5) the t2071 and t2091 verdict files are committed in this branch because they arrived with the absorbed t2091 tip.
- Gaps carried from the verdict: `audit_multi` not called (operator rule), no live run with a rebuilt binary, the worktree guard refused two auditor commands (re-run as plain commands).

## Stage record

The ledger shows `stage run` and `stage sync <5c9e2cb7b7>` for this card; the lease lapsed after that, and the leader lands the card by an operator fast-forward without further stage records, so there is no `sync-audit` or `merge-ready` stage record and none was backfilled.
