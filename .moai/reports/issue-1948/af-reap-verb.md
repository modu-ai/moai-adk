# moai af reap (lane-20)

`moai af reap` takes back the cards whose holder process is gone, using the ledger's `Store.Reap` with the judge
`afAlive` = `flow.ProcessProbe(afProbe)` (homestate.ClassifyOwnerWith: pid>0 and fingerprint present, same
rule the claim presence uses; positively dead → Dead, anything undecided → Unknown, never reclaimed). Output:
`{"reclaimed":[...],"unknown_expired":[...]}` — the second lists expired leases without a usable holder identity
so a lane that really died holding one is visible to the operator.

## Claim 1 — RED (tree 5997cbb66e)
- Stub verb that reaps nothing. `go test -p 4 -count=1 -run TestAFReap ./internal/cli/` →
  `--- FAIL: TestAFReap_DeadSessionLosesItsCard ... reclaimed = [], want t1 once` (the live-session test passed, as it should).
- Processes are real: a `sleep 120` child stands for the session (killed in t.Cleanup), the claim runs through the
  cobra verb with the real fingerprint, the death is a real Kill+Wait.

## Claim 2 — GREEN
- Same command → `TestAFReap_LiveSessionKeepsItsCard` PASS, `TestAFReap_DeadSessionLosesItsCard` PASS (card t1
  reclaimed once, lease gone); `-run TestAF` ok; internal/session full ok 14.9s; retirement test ok 11.3s.
- lint `--new-from-rev=4452d85a5d` on cli, session, agenticfactory: 0 issues.

## lane-19 31st review items (all in this change)
1. wrapperProcessNames += timeout, gtimeout, nohup, sudo, xargs, script, npx, nice, stdbuf (not `node`: the
   session itself can be a node process). RED test `TestAncestorSessionPID_SkipsCommonProcessWrappers`
   (commit before the fix) failed on `stdbuf: ancestorSessionPID = 5300, want 5100`.
2. `afHolder`: an empty fingerprint records pid 0 (test `TestAFHolder_NoFingerprintIsUnknown`).
3. reap lists expired unknown leases (`afUnknownExpired`, test `TestAFUnknownExpired_ListsOnlyExpiredUnknownLeases`);
   it never reclaims them. (Correction of af-lease-holder-pid.md: an unknown holder is NOT replaced by
   "heartbeat + expiry" reclaim; the ledger only writes a stalled note after expiry.)
4. claim's local `session` renamed `sessionID`.
5. AC-AFL-015: the import-edge control is asserted per GOOS (retirement test).

## Gaps
- The claim process-tree test uses the `afOwnerPID` seam to name the real child as the session; the ancestry walk
  itself (session.ResolveOwnerPID) is covered by its fake-tree tests, not by a real wrapper chain.
- Item 1's RED was run and observed only for the first failing wrapper in the list (stdbuf was the last line of the
  failure output; all nine share one assertion).
- `moai af reap` is not wired to run on a schedule or at lane start; it is a verb.
