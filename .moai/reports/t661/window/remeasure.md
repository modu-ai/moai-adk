# t661 window re-measure (merged tree c005ec1f1)

Tree: HEAD c005ec1f1 (absorb merge; ^1 e09900ef5, ^2 local develop 84e5666d9), clean before the run (tree-head.txt).
Pre-check: pgrep for go test/cli.test/go vet/go build exit 1 (none); control pgrep -x zsh count 42. Load before `{ 11.01 10.96 15.63 }`, after `{ 14.38 11.94 15.73 }`.
Command: `go test ./internal/cli/ -count=1 -timeout 600s -v -run ^(TestProfileBaseDirIsSandboxed|TestUserHomeDirFnSandboxesRealHome|TestProfileSandbox_HelperSubprocessLeavesNoBaseDir|TestRunUpdate_V3ProjectWithAgencyDir_MigratesIndependently|TestReproduction_NonProjectDirectoryPollution_Issue1086|TestRunUpdate_ThreeRunIdempotency_V3Project|TestUpdateSubsystem_HomeSeamReach|TestSkipSyncNoArchive)$` > card-tests.txt 2>&1
Result: exit 0; 8 top-level RUN, 8 `--- PASS`, `ok internal/cli 1.956s`; C redirect log lines 1.
Home fingerprint (same command set as full-run, stderr separate): cmp exit 0 for all 8 streams.
Delta judgment: delta-judgment.md. Full suite on the merged tree is left to CI after the lead push (lead ruling).
