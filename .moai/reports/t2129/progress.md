# t2129 progress — PR #1969 Codex P2: RegisterAgent leaks the duplicate agent number in the authority (Class B)

Base: local main 04d8c27a7e. Branch WT-pr-codex-repair-5. Lane: lane-22. Finding: PR #1969 review comment on internal/agenticfactory/claim/identity.go:114 (commit 04d8c27a7e).

## Reproduction (RED, before the fix)

Test `internal/agenticfactory/claim/identity_duplicate_number_t2129_test.go` (TestRegisterAgentReturnsADuplicateNumberToTheAuthority): two stores share one `MemoryRef`; the first registers session 1 (agent/1). The second runs on a ref wrapper whose first `Keys("agent/")` listing comes back empty, which stands for the race window in which the other registration is not yet visible. It passes the "already registered" check, loses agent/1 on the compare-and-swap, wins agent/2, then sees the lower number and gives its own back.
Observed on the unfixed tree, verbatim: `authority keys after the race = [agent/1 agent/2], want only agent/1 (the duplicate's number leaked)`. The returned number (1) was already right; only the authority kept agent/2, because the give-back was `os.Remove` of the local cache file only.
The race is simulated deterministically with the listing wrapper; a real concurrent interleaving was not reproduced separately.

## Fix

`RegisterAgent` now calls the new `giveBackAgent(n)` instead of `_ = os.Remove(...)`. With an attached authority it reads `agent/<n>` for its revision and deletes it conditionally on that revision through `ControlRef.Delete` (an already-gone key counts as given back), then removes the cache file; without an authority it behaves as before except that a failing cache removal is no longer ignored when the file exists. A failed give-back is returned as an error, never swallowed.

## Verification (this run, card worktree, uncommitted at the time of the run)

- `go test -count=1 -timeout 100s ./internal/agenticfactory/claim/` → ok (87.4s), whole package.
- `go test -count=1 -run 'TestRegisterAgentReturnsADuplicateNumberToTheAuthority|TestAgentNumbersAreUniqueAcrossCaches|TestConcurrentRegistrationNeverDuplicatesANumber' -v` → all three PASS (the new test passes after failing before).
- `gofmt -l internal/agenticfactory/claim/` → empty; `go vet ./internal/agenticfactory/claim/` → no output.

## Gaps

- Delete is conditional on the revision read just before; if another caller changes `agent/<n>` in between (a different session taking the number after our release), the delete is refused with a revision conflict, which now surfaces as an error from RegisterAgent. Not exercised by a test here.
- Real concurrent interleaving, race detector, other OS, golangci-lint: not run. No codex card-review yet.
