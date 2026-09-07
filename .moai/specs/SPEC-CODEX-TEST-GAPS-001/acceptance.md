# SPEC-CODEX-TEST-GAPS-001 — Acceptance Criteria

Verification layer. Each AC is binary-testable. Requirement layer lives in spec.md §C (GEARS).

## D. AC Matrix

| AC | Requirement | Given | When | Then | Priority | Verification |
|---|---|---|---|---|---|---|
| AC-CTG-001 | REQ-CTG-001 | `codex_job_control_test.go` extended | The new direct tests for `terminateCodexProcess` run | The `pid <= 0` refusal arm returns an error for pids `-1` and `0`, and the success arm kills a real re-exec helper child and observes its exit; the child kill is registered via `t.Cleanup` before the first assertion | High | `go test -run TestTerminateCodexProcess -v ./internal/cli/` exits rc=0 AND the output contains at least one observed `--- PASS: TestTerminateCodexProcess...` line; an empty-swept run (`[no tests to run]` or zero `--- PASS` lines) is a FAIL |
| AC-CTG-002 | REQ-CTG-002 | `mcp_codex_test.go` extended | The `codexIDMatches` table test runs | All 5 arms assert: empty raw → miss, int match → hit, int mismatch → miss, string-id match → hit, malformed → miss | High | `go test -run TestCodexIDMatches -v ./internal/cli/` exits rc=0 AND at least one observed `--- PASS: TestCodexIDMatches...` line in the output; `[no tests to run]` is a FAIL |
| AC-CTG-003 | REQ-CTG-003 | `mcp_codex_test.go` extended | `awaitCodexResponse` runs with a canceled context and a fake conn yielding non-matching noise lines | The context-cancel return is observed (context error returned, not a match, not an EOF) | High | `go test -run TestAwaitCodexResponse -v ./internal/cli/` exits rc=0 AND at least one observed `--- PASS: TestAwaitCodexResponse...` line; `[no tests to run]` is a FAIL |
| AC-CTG-004 | REQ-CTG-004 | `codex_contract_test.go` extended | The HTML-comment fixtures run through `codexCountExecutingImports` | The multi-line shape (directive after `-->`) and the inline single-line shape (`<!-- ... -->`) both produce the asserted count | Medium | `go test -run TestCodexCountExecutingImports -v ./internal/cli/` exits rc=0 AND at least one observed `--- PASS: TestCodexCountExecutingImports...` line; `[no tests to run]` is a FAIL |
| AC-CTG-005 | REQ-CTG-005 | `codex_init_test.go` extended | `codexGatePrintf` writes through an `io.Writer` whose write returns an error | The call returns silently; no panic escapes the call | Low | `go test -run TestCodexGatePrintf -v ./internal/cli/` exits rc=0 AND at least one observed `--- PASS: TestCodexGatePrintf...` line; `[no tests to run]` is a FAIL |
| AC-CTG-006 | REQ-CTG-006 | `codex_init_test.go` extended | `defaultCodexInitGenerator` runs with an injected `codexwiring.Wire` failure | The returned error wraps the cause (`errors.Is` reaches the sentinel/cause), not a bare pass-through | Low | `go test -run TestDefaultCodexInitGenerator -v ./internal/cli/` exits rc=0 AND at least one observed `--- PASS: TestDefaultCodexInitGenerator...` line; `[no tests to run]` is a FAIL |
| AC-CTG-007 | REQ-CTG-007 | `mcp_codex_test.go` extended | A `codexSessionError` constructed per the production fail-open shape is examined | `Error()` equals `cause.Error()` and `Unwrap()` returns `cause`; `errors.Is`/`errors.As` reach the cause through the wrap | Medium | `go test -run TestCodexSessionError -v ./internal/cli/` exits rc=0 AND at least one observed `--- PASS: TestCodexSessionError...` line; `[no tests to run]` is a FAIL |
| AC-CTG-008 | REQ-CTG-008 | spec.md §D exists | A reader consults the skip record | The record names the `writeCodexRequest`/`writeCodexEnvelope` marshal arms (dead-in-practice) and the `terminateCodexProcess` FindProcess arm (platform-unreachable), each with its reason, and carries the note that `(realCodexConn).pid` was removed in v0.2.0 | High | spec.md §D table contains exactly the two rows and the v0.2.0 removal note |
| AC-CTG-009 | REQ-CTG-009 | Fix-round commit landed as `<FIX-ROUND-BASE-SHA>` (the lane fills the literal SHA when committing this fix round) | The zero-production-diff gate is evaluated as a base-pinned UNION check | The union of `git diff --name-only <FIX-ROUND-BASE-SHA>..HEAD` and `git status --short`, filtered to paths matching non-test `.go` (i.e. ending `.go` but not `_test.go`), is EMPTY — committed and unstaged production changes are both caught | High | `git diff --name-only <FIX-ROUND-BASE-SHA>..HEAD` UNION `git status --short` piped through the non-test-`.go` filter prints nothing; the base SHA used is cited verbatim in run-phase §E |
| AC-CTG-010 | REQ-CTG-010 | Fresh coverprofile over `./internal/cli/` | `go tool cover -func` is extracted | ZERO 0.0% functions remain among the 118 measured function rows — `terminateCodexProcess` covered by AC-CTG-001, `(realCodexConn).pid` by AC-CTG-012, `(codexSessionError).Error`/`Unwrap` by AC-CTG-007; `codexIDMatches` no longer below 60% | High | Verbatim extract cited; count of 0.0% function rows == 0 |
| AC-CTG-011 | REQ-CTG-011 | Touched files | `go vet ./internal/cli/` and `golangci-lint run` execute | Both clean on touched files; `gofmt -l` prints nothing for touched files | High | Verbatim tool outputs cited in run-phase §E |
| AC-CTG-012 | REQ-CTG-012 | `mcp_codex_test.go` extended | `TestRealCodexConnPid` runs over the 3 constructed receivers | `&realCodexConn{}` → 0, `&realCodexConn{cmd: &exec.Cmd{}}` → 0, and a `cmd.Process` from `os.FindProcess(os.Getpid())` → that pid; no subprocess spawned | Medium | `go test -run TestRealCodexConnPid -v ./internal/cli/` exits rc=0 AND at least one observed `--- PASS: TestRealCodexConnPid` line; `[no tests to run]` is a FAIL |

## D.1 Edge cases (covered inside the ACs above)

- M1 success-arm child exits slowly or is already dead before the kill — the cleanup ordering makes the outcome deterministic (kill may report an already-exited child; the test asserts the call contract, not the child's liveness timing).
- `codexIDMatches` malformed arm: input that fails BOTH unmarshal attempts (neither number nor string).
- `awaitCodexResponse` noise lines that match neither id nor method, so the loop keeps reading until the canceled context is observed between reads.
- `realCodexConn.pid` nil-arm ordering: `cmd == nil` must short-circuit before `cmd.Process` is dereferenced (the first two branches are a disjunction).

## D.2 Quality gates

- TRUST 5 Tested: AC-CTG-001..007 and AC-CTG-012 all pass with observed `--- PASS` lines; AC-CTG-010 completion measurement.
- TRUST 5 Readable/Unified: AC-CTG-011 (gofmt/vet/golangci-lint).
- TRUST 5 Secured: child-process handling in AC-CTG-001 is cleanup-guaranteed (no leaked processes; `t.Cleanup` before first assertion).
- TRUST 5 Trackable: conventional commits referencing SPEC-CODEX-TEST-GAPS-001 and card t501.

## D.3 Definition of Done

1. AC-CTG-001..012 all PASS with verbatim command output — including at least one observed `--- PASS` line per targeted selector — cited in run-phase §E.
2. Zero production diffs by the base-pinned union check (AC-CTG-009), with the base SHA cited.
3. Completion measurement shows ZERO remaining 0.0% functions among the 118 (AC-CTG-010).
4. Evidence persisted under `.moai/specs/SPEC-CODEX-TEST-GAPS-001/` (§E.2) or a resolvable evidence path, not `/tmp` alone.
