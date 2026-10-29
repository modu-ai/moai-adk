# lane-22 progress — Issue #1948

Basis: SPEC-AGENTIC-FACTORY-001 0.4.0 (8dc835756b), provisional. Independent audit and merge wait for contract PASS.
Tree: WT-agentic-factory-integration. No push.

## e2e scenarios (internal/agenticfactory/e2e)
| # | Scenario | State |
|---|---|---|
| 1 | competing claim | GREEN (model driver, -race) |
| 2 | duplicate dispatch | GREEN (model); mutant "dedupe off" killed |
| 3 | ACK loss | GREEN (model), 4 sub-cases |
| 4 | interruption recovery | GREEN (model) |
| 5 | audit-budget exhaustion | GREEN (model) |
| 6 | unknown merge result | GREEN (model); lookup-failure block reason `connection-failed` is my mapping, contract says "해당하는 값" |
| 7 | closure + next dispatch after remote landing | GREEN (model); "next dispatch not allowed before done" deliberately not asserted (contract silent) |

## Bound to real lane-18 code (real_claim_test.go)
- Scenario 1 real: 8 concurrent `claim.Store.Acquire` -> exactly 1 winner, epoch 1, late claimant ErrHeld. `go test -race -count=5` ok. Losers may be ErrHeld or ErrBusy (lock retries exhausted); the test accepts both and does not report which occurred.
- Scenario 2 real: repeated same dispatch id -> ErrDuplicateDispatch, record epoch/seq unchanged.
- Scenarios 3-7 on real code: not bound. lane-18 claim/runner exposes no ACK-timeout probe, no dead-owner resume, no audit-round budget; reconcile/effect (lane-20) not yet examined for 4, 6, 7.

## Integration merges
Merged --no-ff locally (no push): lane-18 WT-agentic-factory-core-claim-runner @ca414f1f35, lane-19 WT-agentic-factory-review @9f058f45b7, lane-20 WT-agentic-monitor-effect-reconcile @3c255a16ca.
After the merges: `go build ./...` ok; `go vet ./internal/agenticfactory/...` ok; `go test -race -count=1 ./internal/agenticfactory/...` ok (8 packages).
lane-21 WT-browser-contracts-memo still at 5c3f0ebf27 (nothing to merge). Lane-17 WT-agentic-factory-contract (docs only, 3275706cc9) not merged: SPEC docs belong to lane 17.
CLI: hidden `agentic-factory` skeleton registered in internal/cli/root.go (72264f3db5); name provisional.

## Gaps
- Scenarios run against the reference model only; the real packages are not bound yet, so no claim about real behavior.
- The one-real-Issue measurement is not done.
