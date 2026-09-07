# SPEC-CODEX-TEST-GAPS-001 — Implementation Plan

Tier S | cycle_type recommendation: ddd (ANALYZE-PRESERVE-IMPROVE — characterization-test work over existing, behaviorally-frozen code; zero production diffs mandated by REQ-CTG-009) | harness: minimal (single package, tests only, scoped verification recipe)

## A. Context

Card t501. The card's original premise (6 uncovered codex clusters from a name-citation scan) was mechanically refuted on the execution axis before dispatch; spec.md §B records both axes and the measured residue. This plan covers only that residue: 7 test items in `internal/cli/`, no production changes.

Existing idioms to reuse (verified present in the worktree at authoring time):

- `fakeCodexConn` + canned NDJSON scripts — `mcp_codex_test.go`
- `codexTerminateProcess` seam var — `codex_job_control_test.go` (tests currently swap it; REQ-CTG-001 tests the real body directly instead)
- `codexTestExecImports` cross-checker + fixture style — `codex_init_test.go`
- Existing `codex_contract_test.go` fixtures for comment-shape style reference

## B. Known Issues

- `internal/cli` full-package test run measured 515.6s (spec.md §B.2). The verification recipe (§E below) must carry a timeout budget >= 600s.
- `terminateCodexProcess` success arm spawns a real child process — a leaked child on a failed assertion is a hazard; cleanup ordering is a HARD requirement of REQ-CTG-001.

## C. Pre-flight

1. Confirm branch `WT-codex-uncovered` and base `origin/develop @ ace1c5440` (`git rev-parse --short HEAD`, `git branch --show-current`).
2. Confirm evidence files resolve: `.moai/reports/t501/{namegrep-counts,coverage-perfunc,coverage-run.log}`.
3. Confirm no foreign session shares the tree before first edit (per AGENTS.md §2; this is a card worktree, but re-check before any commit).

## D. Constraints

- Tests only (REQ-CTG-009): every diff lands in `*_test.go`.
- English code/comments; `gofmt` clean; `go vet` + `golangci-lint run` clean on touched files.
- Scoped verification only (§E). No `go test ./...` locally.
- No production-code refactors to "make testing easier" — out of scope (spec.md §F).

## E. Self-Verification (run phase)

Verification recipe (scoped; run from the worktree root):

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -timeout 700s -coverprofile=/tmp/t501_cover_after.out ./internal/cli/
```

- Bash tool timeout: set to the 600,000ms ceiling and run in foreground; if the package exceeds the wall-time budget, split by test pattern (`-run 'TestTerminateCodexProcessDirect|TestCodexIDMatches|...'`) for the iterating loop, and run the full-package scoped command once at the end for the completion measurement (REQ-CTG-010).
- E1: full scoped command exits rc=0 with the package `ok` line observed.
- E2: `gofmt -l internal/cli/` prints nothing for touched files; `go vet ./internal/cli/` exits 0; `golangci-lint run internal/cli/...` clean on touched files.
- E3: extract per-function coverage from the fresh profile (`go tool cover -func=/tmp/t501_cover_after.out`) and confirm exactly ONE 0.0% function among the 118 — `(realCodexConn).pid` (REQ-CTG-010). Cite the verbatim extract line.
- E4: `git diff --name-only` shows only `*_test.go` paths (REQ-CTG-009).

## F. Milestones (priority-ordered per the lead's directive: U2/U3 first; ordered by decision-reversibility — the cross-platform approach decision lands in M1 because it is the least reversible choice in this SPEC)

### M1 — [U3] terminateCodexProcess direct test (REQ-CTG-001) — Priority High

- File: extend `codex_job_control_test.go`.
- **Decision to lock first (least reversible): cross-platform strategy = helper-process re-exec pattern.** Rationale: the success arm must kill a real child; re-exec of the test binary under a sentinel env var (standard `TestHelperProcess` pattern) gives a genuinely killable child on darwin, linux, and windows alike, whereas `runtime.GOOS` gating would either skip the success arm on some CI platforms or duplicate platform-specific child-spawning code. The re-exec helper is itself gated so it only runs when the sentinel env var is set, keeping normal test runs unaffected. (This decision is recorded here so a reviewer can override it before run-phase; changing it later means rewriting the test.)
- `pid <= 0` refusal arm: pure table entries (`-1`, `0`), assert error returned, no side effects.
- Success arm: spawn the helper child, register `t.Cleanup(func() { _ = terminate kill; wait })` BEFORE the first assertion, call `terminateCodexProcess(pid)`, assert nil error, then assert the child has exited.
- Never spawn background load; the child is short-lived and reaped by the cleanup.

### M2 — [U2] codexIDMatches table test (REQ-CTG-002) — Priority High

- File: extend `mcp_codex_test.go`.
- Justification for extending vs a new file: `fakeCodexConn` and the existing frame-level helpers live in `mcp_codex_test.go`; `codexIDMatches` and `awaitCodexResponse` are frame helpers consumed by those paths — colocating keeps the fake and its consumers' tests in one file. A new small file would split a cohesive seam.
- Table: empty raw (miss), int match (hit), int mismatch (miss), string-id match (hit — the strconv arm), malformed (both unmarshals fail → miss).

### M3 — [U2] awaitCodexResponse ctx-cancel arm (REQ-CTG-003) — Priority High

- File: extend `mcp_codex_test.go` (same rationale as M2).
- Canceled context + fake conn yielding non-matching noise lines; assert the ctx-cancel return is observed between reads (the read loop returns the context error, not a match).

### M4 — [U6] codexCountExecutingImports HTML-comment fixtures (REQ-CTG-004) — Priority Medium

- File: extend `codex_contract_test.go`.
- Two fixture shapes: multi-line comment with directive after `-->` close; inline single-line `<!-- ... -->`. Assert counts; style-match the `codexTestExecImports` fixture conventions.

### M5 — codexSessionError delegation contract (REQ-CTG-007) — Priority Medium

- File: extend `mcp_codex_test.go`.
- Construct via the production fail-open shape (`codexHandshakeFailure` or direct struct literal), assert `Error()` equals `cause.Error()` and `Unwrap()` returns `cause` (assert both `errors.Is`/`errors.As` reachability through the wrap).

### M6 — [U7] error arms: codexGatePrintf + defaultCodexInitGenerator (REQ-CTG-005, REQ-CTG-006) — Priority Low

- File: extend `codex_init_test.go` (seam-injection idioms already live there).
- `codexGatePrintf`: an `io.Writer` whose write returns an error → silent return, no panic (assert via recovery-free contract: the call returns and any panic fails the test naturally).
- `defaultCodexInitGenerator`: inject a `codexwiring.Wire` failure path per the existing seam; assert the returned error wraps the cause (`errors.Is` reaches the sentinel), not a bare pass-through.

### M7 — Completion re-measurement + skip-record verification (REQ-CTG-008, REQ-CTG-010) — Priority High

- Run the §E full scoped command; extract per-function coverage; confirm exactly ONE remaining 0.0% function (`(realCodexConn).pid`).
- Confirm spec.md §D skip record is intact and matches the final profile (no new 0.0% functions appeared).

## G. Anti-Patterns (must NOT do)

- Do NOT write per-cluster test files for the behaviorally-tested paths (spec.md §F).
- Do NOT touch non-test `.go` files, including "small" refactors.
- Do NOT register cleanup after the first assertion in M1's success arm.
- Do NOT run `go test ./...` locally, and do not run the scoped package command repeatedly in a fix loop — use `-run` pattern scoping while iterating (§E).
- Do NOT use `t.Setenv` with OTEL variables (repo CLAUDE.local.md §6 [WARN]).
- Do NOT spawn background load for the M1 child process; re-exec helper + `t.Cleanup` only.

## H. Cross-References

- spec.md §B (refuted-premise narrative + evidence paths), §D (skip record), §F (out of scope)
- acceptance.md §D AC matrix
- `.moai/reports/t501/` — the three evidence files cited throughout
