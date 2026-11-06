# t2137 progress (Class A, main red: home-guard)

## Claim
Two NEW home accesses cleared; TestHomeGuard_RepositoryHasNoNewHomeAccess passes.

## Evidence
Before (lane-reproduced, as dispatched; not re-run by me): exit 1, 2 new:
- internal/agenticfactory/core/sanitize_test.go:95 `"~/work/notes.md"`
- internal/cli/factory_fixture_home_test.go:36 message "...real ~/.moai"

After (this run, base 04d8c27a7e, fix commit 629bf3ea8b):
```
=== RUN   TestHomeGuard_RepositoryHasNoNewHomeAccess
    repo_guard_test.go:96: scanned 3927 test files: 160 violations, seed 184 entries, 0 new, 24 stale seed entries (stale entries do not fail; regenerate with TESTGUARD_WRITE_SEED=1)
--- PASS: TestHomeGuard_RepositoryHasNoNewHomeAccess (1.74s)
ok  	github.com/modu-ai/moai-adk/internal/testguard	1.842s
```
- `go test -count=1 -run TestSanitizeLeavesOrdinaryTextAlone ./internal/agenticfactory/core/` -> ok 0.262s
- `go test -count=1 -run TestFcFixtureKeepsMoaiHomeInsideTheTestSandbox ./internal/cli/` -> ok 1.479s
- `gofmt -l` on both files: no output.

Changes: marker line `// home-guard:allow deliberate input: ...` directly above the tilde entry; message reworded to "the real user home's .moai". Seed untouched.

## Gaps
- Pre-fix guard output taken from the dispatch, not re-run here.
- Only the three named tests were run, not the full packages.
- The 24 stale seed entries are untouched by design.

## Residual-risk
Merge with main may add other new violations.
