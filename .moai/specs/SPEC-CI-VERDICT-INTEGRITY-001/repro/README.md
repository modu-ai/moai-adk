# repro/ — evidence-regeneration scripts for plan.md §B ledger entries

Canonical, committed copies of every RED/observation probe the ledger cites.
The first measurements ran from `/tmp/t1534-probe/` working copies; those were
ephemeral — these committed copies are the citable ground. `run-gate-cancelled`,
`run-checks-loop`, and `run-install-summary-cancelled` EXTRACT the workflow step
body from the live workflow file at run time (awk slice by step-name anchor +
template substitution) so a re-run after the repo repair exercises the repaired
logic, never a frozen copy.

| Script | Ledger entry | RED expectation (pre-repair) |
|--------|--------------|------------------------------|
| `run-gate-cancelled.sh` | E2 | gate prints PASSED, exit 0, on a cancelled matrix |
| `run-checks-loop.sh` (+ `stubbin/gh`) | E4 | lookup failure/empty output → `should_merge=true` |
| `run-install-summary-cancelled.sh` | E7 | 19-case matrix: full-success control exit 0 + ALL 18 single-dependency failures (`cancelled`/`timed_out`/`skipped` × 6 needs) exit 0 |
| `run-malformed.sh` (+ `malformed/`) | E11 | malformed SSoT → vacuous pass, exit 0 |
| `run-phantom.sh` (+ `phantom/`) | E24 | valid YAML with a phantom required context → silent pass, exit 0 |
| `run-phantom-control.sh` (+ `phantom-control/`) | E27 | identical tree, context swapped to a published name → exit 0 (stays 0 post-M2; the two-directional pair with E24) |
| `run-deadline-mutant.sh` | E26 | merge-step condition carries no deadline term; merge proceeds at deadline+1, exit 0 (behavioral: pre/post clock runs must yield proceeds-then-held to count as repaired) |

All scripts run from the repository root. Post-repair, each flips per the
green-path cell of its acceptance criterion.
