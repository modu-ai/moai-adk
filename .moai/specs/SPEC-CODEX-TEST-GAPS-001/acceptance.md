# SPEC-CODEX-TEST-GAPS-001 — Acceptance Criteria

Verification layer. Each AC is binary-testable. Requirement layer lives in spec.md §C (GEARS).

## D. AC Matrix

| AC | Requirement | Given | When | Then | Priority | Verification |
|---|---|---|---|---|---|---|
| AC-CTG-001 | REQ-CTG-001 | `codex_job_control_test.go` extended | The new direct tests for `terminateCodexProcess` run | The `pid <= 0` refusal arm returns an error for pids `-1` and `0`, and the success arm kills a real re-exec helper child and observes its exit; the child kill is registered via `t.Cleanup` before the first assertion | High | `go test -run TestTerminateCodexProcess ./internal/cli/` rc=0 |
| AC-CTG-002 | REQ-CTG-002 | `mcp_codex_test.go` extended | The `codexIDMatches` table test runs | All 5 arms assert: empty raw → miss, int match → hit, int mismatch → miss, string-id match → hit, malformed → miss | High | `go test -run TestCodexIDMatches ./internal/cli/` rc=0; table contains exactly the 5 named arms |
| AC-CTG-003 | REQ-CTG-003 | `mcp_codex_test.go` extended | `awaitCodexResponse` runs with a canceled context and a fake conn yielding non-matching noise lines | The context-cancel return is observed (context error returned, not a match, not an EOF) | High | `go test -run TestAwaitCodexResponse ./internal/cli/` rc=0 |
| AC-CTG-004 | REQ-CTG-004 | `codex_contract_test.go` extended | The HTML-comment fixtures run through `codexCountExecutingImports` | The multi-line shape (directive after `-->`) and the inline single-line shape (`<!-- ... -->`) both produce the asserted count | Medium | `go test -run TestCodexCountExecutingImports ./internal/cli/` rc=0 |
| AC-CTG-005 | REQ-CTG-005 | `codex_init_test.go` extended | `codexGatePrintf` writes through an `io.Writer` whose write returns an error | The call returns silently; no panic escapes the call | Low | `go test -run TestCodexGatePrintf ./internal/cli/` rc=0 |
| AC-CTG-006 | REQ-CTG-006 | `codex_init_test.go` extended | `defaultCodexInitGenerator` runs with an injected `codexwiring.Wire` failure | The returned error wraps the cause (`errors.Is` reaches the sentinel/cause), not a bare pass-through | Low | `go test -run TestDefaultCodexInitGenerator ./internal/cli/` rc=0 |
| AC-CTG-007 | REQ-CTG-007 | `mcp_codex_test.go` extended | A `codexSessionError` constructed per the production fail-open shape is examined | `Error()` equals `cause.Error()` and `Unwrap()` returns `cause`; `errors.Is`/`errors.As` reach the cause through the wrap | Medium | `go test -run TestCodexSessionError ./internal/cli/` rc=0 |
| AC-CTG-008 | REQ-CTG-008 | spec.md §D exists | A reader consults the skip record | The record names `(realCodexConn).pid` (live-gated), the `writeCodexRequest`/`writeCodexEnvelope` marshal arms (dead-in-practice), and the `terminateCodexProcess` FindProcess arm (platform-unreachable), each with its reason | High | spec.md §D table contains all three rows |
| AC-CTG-009 | REQ-CTG-009 | Run phase complete | `git diff --name-only` (and `git status --short`) is read | Every changed/added path matches `*_test.go` — zero non-test `.go` diffs | High | `git diff --name-only | grep -v '_test\.go$'` prints nothing |
| AC-CTG-010 | REQ-CTG-010 | Fresh coverprofile over `./internal/cli/` | `go tool cover -func` is extracted | Exactly ONE 0.0% function among the 118 — `(realCodexConn).pid` (mcp_codex.go:494); `codexIDMatches` no longer below 60%; the 7 tested functions each above their pre-SPEC measured values | High | Verbatim extract line for `(realCodexConn).pid` cited; count of 0.0% functions == 1 |
| AC-CTG-011 | spec.md §E (Constraints 2, 3) | Touched files | `go vet ./internal/cli/` and `golangci-lint run` execute | Both clean on touched files; `gofmt -l` prints nothing for touched files | High | Verbatim tool outputs cited in run-phase §E |

## D.1 Edge cases (covered inside the ACs above)

- M1 success-arm child exits slowly or is already dead before the kill — the cleanup ordering makes the outcome deterministic (kill may report an already-exited child; the test asserts the call contract, not the child's liveness timing).
- `codexIDMatches` malformed arm: input that fails BOTH unmarshal attempts (neither number nor string).
- `awaitCodexResponse` noise lines that match neither id nor method, so the loop keeps reading until the canceled context is observed between reads.

## D.2 Quality gates

- TRUST 5 Tested: AC-CTG-001..007 all pass; AC-CTG-010 completion measurement.
- TRUST 5 Readable/Unified: AC-CTG-011 (gofmt/vet/golangci-lint).
- TRUST 5 Secured: child-process handling in AC-CTG-001 is cleanup-guaranteed (no leaked processes; `t.Cleanup` before first assertion).
- TRUST 5 Trackable: conventional commits referencing SPEC-CODEX-TEST-GAPS-001 and card t501.

## D.3 Definition of Done

1. AC-CTG-001..011 all PASS with verbatim command output cited in run-phase §E.
2. Zero production diffs (AC-CTG-009).
3. Completion measurement shows exactly one remaining 0.0% function (AC-CTG-010), matching spec.md §D's documented skip record.
4. Evidence persisted under `.moai/specs/SPEC-CODEX-TEST-GAPS-001/` (§E.2) or a resolvable evidence path, not `/tmp` alone.
