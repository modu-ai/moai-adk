# F04 verdict

## Claim

F04 is closed by the existing sync quality-gate state machine and a new
end-to-end regression fixture. A failed check is recorded as `fail` with its
payload and is re-delivered on the next call for the same HEAD; it is not
mistaken for a silent successful once-per-commit result.

## Evidence

The fixture creates a temporary Git repository, makes `go vet` exit 7 while
`go build` succeeds, invokes the real hook twice, and asserts both invocations
emit the blocking JSON plus the persisted `<HEAD> fail` state.

Command:

```text
.claude/hooks/tests/test-sync-phase-quality-gate.sh
PASS: failed gate is re-delivered for the same HEAD
```

The inspected implementation is at
`.claude/hooks/moai/sync-phase-quality-gate.sh`; the latest pre-card source
commit is `1d7af6f21` (t624), already present on local `develop`.

## Baseline-attribution

The fixture ran in `WT-workflow-audit-f04` after fast-forwarding the worktree
to local `develop` commit `f6d25b2bf`.

## Gaps

This fixture exercises a completed failed run, not a process killed during the
`running` interval. The hook's stale-running recovery remains a separate path.

## Residual-risk

The Stop runtime may suppress output when its own `stop_hook_active` guard is
set; that suppression is intentional and covered only by the hook's branch
logic, not by this fixture.
