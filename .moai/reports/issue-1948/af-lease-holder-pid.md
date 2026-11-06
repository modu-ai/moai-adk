# af claim lease holder (lane-20)

Problem (lane-22 residual risk, lane-19 30th review): `moai af claim` recorded `os.Getppid()` as the lease holder.
Run through the Bash tool the parent is the per-command `zsh -c` wrapper, which exits with the command, so the
holder is dead on arrival; the first reap that judges liveness by pid+fingerprint would take the card from a live
session (third reclaim parks it). `ProbeProcessIdentity`'s state was also discarded, leaving an empty fingerprint.

## Claim 1 — RED (tree 150b686136)
- Command: `go test -p 4 -count=1 -run TestAFHolder ./internal/cli/`
- Observed: `--- FAIL: TestAFHolder_SkipsTheShortLivedWrapper ... holder pid = 9001, want the session's 4242` and
  `--- FAIL: TestAFHolder_UnresolvedSessionIsUnknownNotTheWrapper ... holder = (9001, "start-of-1"), want (0, "")`.
- Baseline: the old behaviour extracted unchanged into `afHolder`; the tests drive it through seams
  (the wrapper is the parent and is never asked in the fixed code).

## Claim 2 — GREEN (tree 68cdd89aeb)
- Fix: `afHolder` uses `session.ResolveOwnerPID` (nearest non-shell ancestor, MOAI_SESSION_PID checked against the
  own session) and probes THAT pid for the fingerprint; unresolved → pid 0 / no fingerprint. The ledger judges
  a lease by pid only when pid > 0 (`claim.go:63`), so an unknown holder is never reclaimed by liveness (I5) and
  falls back to heartbeat + expiry.
- Command: `go test -p 4 -count=1 -v -run TestAFHolder ./internal/cli/` → both PASS; `-run 'TestAFHolder|TestAFCommand|AF'` ok.

## Gaps / residual
- Not observed: a real process tree (child claim, parent exits, real probe, Reap) — the tests drive the seams.
  ResolveOwnerPID's own ancestry walk has its tests in internal/session.
- `Indeterminate` from the probe still yields an empty fingerprint with a real pid: ClassifyOwnerWith reads that
  as Indeterminate → Unknown → alive, so it is safe, but the fingerprint is then absent.
- No Alive judge and no reap verb exist yet (`Store.Reap` has no non-test caller): a really dead lane is not
  reclaimed through `moai af` — the other half of the -l zombie root fix, not done here.

## AC-AFL-015 P3 follow-up (68cdd89aeb)
- go list now runs for the host, windows, linux and darwin, with a 3-minute timeout per call; the pre-retirement
  tree still fails (76 lines carry a GOOS tag, windows lines included), the retired tree passes (11.9s).
