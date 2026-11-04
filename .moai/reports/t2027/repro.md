# t2027 reproduction record (tree base 9fc89e246e, branch WT-internal-agenticfactory-)

Loads (uptime, 1/5/15 min): 56.81 57.30 61.58 (before) / 49.79 53.74 59.54 / 55.03 54.16 59.40 / 65.94 58.32 60.44 / 56.89 57.19 59.89 (after P3).

## F1 mergeq.go:284 - gate modifies a tracked file

Test (untracked, not committed): internal/agenticfactory/mergeq/gate_modifies_tracked_repro_test.go
Command: go test ./internal/agenticfactory/mergeq/ -count=1 -v -run TestLand_GateThatModifiesATrackedFile_Reproduction
Observed:
    outcome=landed reason="" trunkAfter=997f60d7... trunk=997f60d7...   (trunk == pre-gate cand.Result)
    status after gate: "M base.txt"
    base.txt on the trunk: "base"          (the gate's edit is not in what landed)
    check worktree still exists: false      (Discard removes it)
Disposition: behavior reproduces as described. Fix is a DESIGN DECISION (refuse the landing on a dirty tree vs commit the gate's changes into the merge result vs tolerate; needs a new Git interface method; gates such as `go mod tidy`/`go test` can legitimately touch go.sum). No product code changed.

## F2 location_memo.go:149 - symlink re-pointed

Plain case (two repos created back to back): NOT reproduced - config mtime differs, fingerprint changes, memo recomputes.
Equal-stamp case (repoB .git/config size+mtime == repoA's): REPRODUCED. memo answered repoA, cold answered repoB.
Fix: locationFingerprint now starts with filepath.EvalSymlinks(dir) ("=>" line). Tests: internal/homestate/location_memo_symlink_test.go.

## F3 af.go:219 - relative --db in the heartbeat binding

NOT reproduced. Claim verb with --db "ledger.db", "af/ledger.db", "./ledger.db" fails at ledger.Open:
    af: open the ledger ledger.db: ledger: SQL logic error: invalid uri authority: ledger.db?_pragma=... (1)
    af: open the ledger af/ledger.db: ... invalid uri authority: af (1)
    af: open the ledger ./ledger.db: ... invalid uri authority: . (1)
No binding file is written. (ledger.Open builds url.URL{Scheme:"file", Path:path}; a relative path becomes a URI authority.)
Latent only: if ledger.Open ever accepts relative paths, af.go:219 would store them as given. No change made.
Test (untracked, not committed): internal/cli/af_claim_binding_test.go (TestAFClaim_RelativeDBReachesNoBinding)
