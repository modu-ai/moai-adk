# lane-18 progress — claim / runner (Issue #1948)

Basis: SPEC-AGENTIC-FACTORY-001 0.4.0 (8dc835756b), 0.5.0 (3275706cc9) names unchanged for state/ledger. Provisional until plan-auditor PASS.
Tree: .moai/worktrees/af-core-claim-runner, branch WT-agentic-factory-core-claim-runner, base cb30e05b3e. Not pushed.

## M0 core/ — done
- Command: go test -cover ./internal/agenticfactory/core/
- Observed: ok ... coverage: 95.3% of statements (gofmt -l and go vet clean)
- Contents: State (9), BlockReason (18, operator-hold on-hold only), Delivery (4, ordered), RunID/AttemptID/HeadSHA/IssueRef, error sentinels (Held != Busy), Ledger/Evidence/Binding.
- Gap: transition table is the `state` scope, not here. Identity (attributed/unattributable) is IDENTITY-001, only Binding carried.

## Next
- M1 claim/: atomic claim keyed by (issue), HEAD bound in record, duplicate dispatch rejected, write refused on ownership failure.
- M2 runner/: waiting -> claimed -> implementing -> verifying -> ready -> done, blocked/on-hold separate.
