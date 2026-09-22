# t1074 run-phase verdict draft

## Claim

The M1-M3 factory-only logical-lane/current-endpoint broker implementation is locally verified by AC-FMH-001 through AC-FMH-009. The overall SPEC verdict is **FAIL** because all five live criteria and the fixed performance criterion are not passing.

## Evidence

- Baseline: implementation began from `0314801c200fb9a8d29363ef35261d695acd2eb1`; audited SPEC commit `8d8e101bf`; M1 `cb099897a`; M2/M3 `6bde8412c`.
- Exact gates: `.moai/reports/t1074/ac01.jsonl` through `ac09.jsonl` each contain the named test `Action=pass`; the acceptance JQ predicates each printed `true`.
- Scoped regression: `GOCACHE=/tmp/t1074-final-target go test ./internal/factorymsg ./internal/cli ./internal/hook ./internal/template ./internal/mcp -run '^(TestFactory...|TestMoaiMCP.*|TestMoaiMCPTools_.*|TestFactoryStopHookSynchronousParity)$' -count=1 -timeout 90s` returned `ok` for all five packages.
- Strict lint: `GOCACHE=/tmp/t1074-lint-cache go run ./cmd/moai spec lint SPEC-FACTORY-MIXED-HOOK-001 --strict --json` exited 0 with one informational `OwnershipTransitionUnmeasured` item and no error.
- Live logs: `ac10.jsonl`..`ac12.jsonl` record process-start identity `indeterminate`; `ac13.jsonl` and `ac14.jsonl` record Claude subscription HTTP 401.
- Benchmark log: `ac15.jsonl` records all requested queue/session cells and the fixed threshold failure. The final real-worktree empty p95 was `370.799917ms`, above 50ms; every matrix cell also exceeded 200ms.
- `git diff --check` returned no output and exit 0 after the final scoped verification.

## Per-AC verdict

| AC | Verdict | Evidence |
|---|---|---|
| AC-FMH-001..009 | PASS | `ac01.jsonl`..`ac09.jsonl` |
| AC-FMH-010..012 | FAIL | Codex owner fingerprint unavailable on this host |
| AC-FMH-013 | FAIL | Claude OAuth revoked |
| AC-FMH-014 | FAIL | Claude OAuth revoked before boundary receipt |
| AC-FMH-015 | FAIL | Empty p95 and real-worktree inspection exceeded fixed budgets |

## Gaps

- No successful real Codex or Claude model round trip was observed.
- Cross-platform builds and repository-wide tests were not completed.
- A broad local package run was not a clean baseline: environment-sensitive CLI tests failed under the isolated `MOAI_HOME`, an opt-in live Codex test attempted a handshake, and `internal/hook` reached its 120-second timeout. Those failures are not claimed as t1074 regressions.

## Residual risk

- Hook cold-path canonical project resolution dominates latency and currently violates the fixed performance target.
- Real MCP approval/config behavior remains unverified until valid Claude authentication and a deterministic Codex owner fingerprint are available.
- Worktree creation/cwd transition/BOUND lifecycle, idle wake, and `CLAUDE.local.md` loading remain explicitly owned by t1082, t1075, and t1078 respectively.
