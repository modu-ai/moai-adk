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

## Integration merges
lane-18/19/20/21 branches: none found yet (only lane-17 af-contract-spec and lane-21 agentic-factory-21 trees exist; lane-21 still at 5c3f0ebf27).

## Gaps
- Scenarios run against the reference model only; the real packages are not bound yet, so no claim about real behavior.
- The one-real-Issue measurement is not done.
