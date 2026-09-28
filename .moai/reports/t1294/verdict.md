# t1294 — Codex factory session and slot regression

## Claim

The installed rc.16 incident supplied by the operator is a regression scenario for the F2 factory implementation. This checkout cannot yet validate or repair that scenario: Codex factory entry is currently retired, and t1240 owns its reintroduction.

## Evidence

- Operator-provided prior observation: installed `moai` identified commit `a8a9b9376`; `moai codex -f agent` in run `tm1saz` registered workers numbered 10–17 while the lead expected worker 1. The worker bindings remained `launch_pending` without session UUIDs. Treat this as incident input, not a fresh live measurement.
- `git show a8a9b9376:internal/kanban/bootstrap.go` defines `NextFactoryWorkerNumber` as `highest + 1` after pruning dead claims. `git show a8a9b9376:internal/cli/codex_factory.go` documents `-f agent` as a legacy alias for worker join.
- Current checkout `d5df9457c`, branch `WT-codex-factory-binding`: `go test ./internal/cli -run '^TestCodexFactoryEntryIsRefused$' -count=1 -v -timeout 90s` returned:

  ```text
  === RUN   TestCodexFactoryEntryIsRefused
  --- PASS: TestCodexFactoryEntryIsRefused (0.00s)
  PASS
  ok  github.com/modu-ai/moai-adk/internal/cli  0.822s
  ```

- The official SQLite todo card `t1240` specifies `moai cc|glm|codex -f agent` as a new self-dispatch role. This differs from rc.16's legacy worker alias. Card `t1292` moves MoAI-created L1 worktrees to `.moai/worktrees/`; its isolated implementation commit is `8c23dca9f` and awaits the integration window.

## Baseline attribution

The passing refusal test above was measured against this checkout's `d5df9457c` tree in this run. The rc.16 incident and `tm1saz` binding state were supplied by the operator; their live state was not remeasured here. The historical code was read from commit `a8a9b9376` in this run.

## Gaps

- t1240 has not landed in this checkout, so F2 `-f agent` launch, self-dispatch, and Codex session binding cannot yet be tested.
- The integration window is held by worker-63 for t1256. No merge to `develop` or remote push is claimed.
- The live `tm1saz` run and its Codex windows were not changed or rechecked.

## Residual risk and completion gate

After t1240 and t1292 integrate, test `-f agent` separately from worker joins. Prove it does not claim a worker number, then prove explicit worker joins use only the selected run's vacant allowed slots with no state mutation on overflow. Verify the state/help text explains `launch_pending` until the first binding hook and that a bound entry carries the selected run, role, and session UUID. Record the exact argv, state transition, and scoped test output here before marking t1294 complete.
