# t1475 verdict — schema guard validates raw wire error frames

Class B (test-guard defect). Branch WT-codex-error-frame-raw-schema-guard, fix commit cfaabc0c3, base 2b9e4a4d0 (= develop).

## Claim
TestManagedServerRequestPolicyMatchesCodexSchema re-assembled error frames from the decoded struct, so a wire frame without error.message passed JSONRPCError.json. It now validates the wire bytes (hardenFrame.Raw).

## Evidence
RED (before fix, `go test -count=1 -run TestManagedSchemaGuardSeesMessagelessErrorFrame ./internal/cli/`):
```
--- FAIL: TestManagedSchemaGuardSeesMessagelessErrorFrame (0.00s)
    managed_hardening_test.go:586: wire frame {"id":7,"error":{"code":-32601}} lacks error.message yet the guard accepts it
    managed_hardening_test.go:586: wire frame {"id":"7","jsonrpc":"2.0","error":{"code":-32601}} lacks error.message yet the guard accepts it
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.215s
```
GREEN:
```
--- PASS: TestManagedSchemaGuardSeesMessagelessErrorFrame (0.00s)
--- PASS: TestManagedServerRequestPolicyMatchesCodexSchema (0.51s)
go test -count=1 -race -run TestManaged ./internal/cli/ -> ok  github.com/modu-ai/moai-adk/internal/cli 27.766s
go vet ./internal/cli/ -> clean
golangci-lint v2.1.6 run ./internal/cli/ -> 0 issues.
```
Full owning package `go test -count=1 -timeout 30m ./internal/cli/...`: all subpackages ok; internal/cli FAIL 1801.750s with two timing failures:
```
--- FAIL: TestStopChainMemberCostWithinBudget — member 2 (sync-phase quality gate) observed max 1.812582166s exceeds its declared budget 1s
--- FAIL: TestCodexTaskBackgroundHandshakeHonorsTaskBound — background handshake outlived the 100ms task bound
```
Isolated re-run of both: `ok github.com/modu-ai/moai-adk/internal/cli 25.025s`.

## Baseline-attribution
This run, this tree (HEAD cfaabc0c3 over base 2b9e4a4d0).

## Gaps
- The two load-dependent timing failures were not re-run on the base tree; "unrelated" is inferred from their isolated pass and their subject (stop-hook timing, codex task handshake). CI is the verdict.
- No live codex session was observed emitting a message-less frame; the regression test uses hand-written wire frames.
- `.moai/reports/t1409/sync-audit-delta.md` is not present in this tree; debt text was read from SPEC-FACTORY-MANAGED-HARDEN-001 progress.md:525.

## Residual-risk
Production writer (managed_codex_factory.go) unchanged; the guard would now catch a message-less regression there.
