# t1074 run-phase verdict draft

## Claim

AC-FMH-001 through AC-FMH-009 and AC-FMH-015 have recorded local evidence. AC-FMH-010 proves a real Codex nonce/receipt exchange, but production launcher-to-hook-to-MCP identity binding remains **UNVERIFIED**. The overall SPEC verdict remains **FAIL**; the recorded Claude live attempts also have not satisfied AC-FMH-011 through AC-FMH-014.

## Evidence

- Baseline: implementation began from `0314801c200fb9a8d29363ef35261d695acd2eb1`; audited SPEC commit `8d8e101bf`; M1 `cb099897a`; M2/M3 `6bde8412c`; M4 harness `45285bf1b` plus the current worktree repair diff.
- Exact gates: `.moai/reports/t1074/ac01.jsonl` through `ac09.jsonl` each contain the named test `Action=pass`; the acceptance JQ predicates each printed `true`.
- Scoped regression: `GOCACHE=/tmp/t1074-final-target go test ./internal/factorymsg ./internal/cli ./internal/hook ./internal/template ./internal/mcp -run '^(TestFactory...|TestMoaiMCP.*|TestMoaiMCPTools_.*|TestFactoryStopHookSynchronousParity)$' -count=1 -timeout 90s` returned `ok` for all five packages.
- Strict lint: `GOCACHE=/tmp/t1074-lint-cache go run ./cmd/moai spec lint SPEC-FACTORY-MIXED-HOOK-001 --strict --json` exited 0 with one informational `OwnershipTransitionUnmeasured` item and no error.
- Codex live log: `ac10.jsonl` records separate real Codex contexts, manually registered fixture sleep-process owner identities, the nonce/receipt exchange with payload omitted here, `acknowledged=2`, and a passing exact JQ predicate. It does not exercise production owner registration.
- Claude live log: `ac11.jsonl` records a successful Codex lead send followed by Claude's `OAuth session expired and could not be refreshed`. A separate project-managed-profile probe reached the subscription service but returned HTTP 429 weekly-limit exhaustion. AC-FMH-012 through AC-FMH-014 were not rerun on this baseline.
- Benchmark log: `ac15.jsonl` records all requested queue/session cells, empty p50 `31.216917ms`, empty p95 `33.143167ms`, and contention `57.022917ms`; the exact predicate printed `true` without threshold changes.
- `git diff --check` returned no output and exit 0 after the final scoped verification.

## Per-AC verdict

| AC | Verdict | Evidence |
|---|---|---|
| AC-FMH-001..009 | PASS | `ac01.jsonl`..`ac09.jsonl` |
| AC-FMH-010 | UNVERIFIED — production identity axis | `ac10.jsonl`: two acknowledgements and exact gate `true`, but injected sleep-owner PIDs bypass production registration |
| AC-FMH-011 | FAIL | `ac11.jsonl`: Codex send succeeded; Claude OAuth refresh failed |
| AC-FMH-012..014 | FAIL | Not rerun: required Claude subscription turns remain unavailable |
| AC-FMH-015 | PASS | `ac15.jsonl`: empty p95 `33.143167ms`, all inspection cells below 200ms, exact gate `true` |

## Gaps

- Independent delta audit at `99d77dd25` found that `factory_live_test.go:125-220` manually registers sleep-process identities and stamps `MOAI_SESSION_PID`. A process-faithful `moai codex -f` → SessionStart → MCP live gate is still required. No production runtime defect was reproduced by that audit.
- Operator report: in another project (`mo.ai.kr`), a lead launched with `moai codex -f` could not inspect two workers launched with `moai codex -f agent`. The screenshot is a reproduction report, not a completed diagnostic. At inspection, the installed binary was built from `cd99336bf`; `git merge-base --is-ancestor 99d77dd25 develop` returned exit 1. This card's current `Store.Status` returns message counts, not a live lane roster. The added lane-status contract and a separate-project lead-plus-two-workers live check remain pending.
- No successful real Claude model round trip was observed on the repaired baseline.
- Cross-platform builds and repository-wide tests were not completed.
- Separate full `internal/hook` and `internal/cli` runs both reached the explicit 180-second local timeout amid broad unrelated test activity. This is retained as a gap, not claimed as either pass or t1074 regression.

## Residual risk

- Claude-side MCP approval/config behavior remains unverified until subscription capacity permits AC-FMH-011 through AC-FMH-014 to run.
- The Codex MCP env allowlist and launcher attribution paths are covered locally, but cross-platform process-fingerprint behavior outside Darwin/Linux remains unmeasured.
- Worktree creation/cwd transition/BOUND lifecycle, idle wake, and `CLAUDE.local.md` loading remain explicitly owned by t1082, t1075, and t1078 respectively.
