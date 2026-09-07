---
id: SPEC-CODEX-TEST-GAPS-001
title: "Codex uncovered-surface test reinforcement — terminateCodexProcess, codexIDMatches, awaitCodexResponse cancel arm, HTML-comment import fixtures, error arms, delegation contract"
version: "0.2.0"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/cli
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "codex, testing, coverage, test-gaps, internal-cli, characterization"
related_specs: [SPEC-CODEX-WIRING-001, SPEC-CODEX-SESSION-MSG-001]
---

# SPEC-CODEX-TEST-GAPS-001 — Codex uncovered-surface test reinforcement

## A. History

- 2026-09-07 — v0.1.0 — Card t501 plan-phase. Authored from a mechanically-refuted investigation: the card's original premise ("6 uncovered clusters") was confirmed on the name-citation axis and refuted on the execution axis. SPEC narrowed to the genuine residue (7 test items + 1 documented-skip record). Evidence: `.moai/reports/t501/{namegrep-counts,coverage-perfunc,coverage-run.log}`.
- 2026-09-07 — v0.2.0 — Plan-audit iteration 1 fix round (score 0.875, FAIL on 5 blocking instrument defects). Tier S→M (artifact set + REQ/AC counts exceeded the Tier S ceiling). `(realCodexConn).pid` removed from the documented-skip record — the rationale was refuted by direct read of mcp_codex.go:494-499 (a 3-branch field read, same-package constructible); an 8th test item (REQ-CTG-012) covers it. Completion AC retargeted to ZERO remaining 0.0% functions. AC verification verbs hardened against empty-sweep green; zero-diff gate rewritten as a base-SHA-pinned union check. Quality-gate constraint promoted to REQ-CTG-011 (was an orphaned AC). Census denominators clarified (114 unique names vs 118 coverprofile rows).

## B. Refuted-Premise Narrative (audit trail — why this SPEC is small)

The card originated from a function-name-based citation scan claiming 6 clusters of "uncovered" codex surfaces. The lane refuted/refined the claim mechanically before dispatch. Both axes are recorded here so the next investigator does not re-derive them and does not re-flag the closed surfaces.

### B.1 Axis 1 — name-citation: CONFIRMED TRUE

62 of 114 functions in the 6 target files are never named in any `*_test.go` (scan scope closed: `internal/`, `pkg/`, `cmd/` test files only; no codex surface exists in `test/integration`, `e2e/`, or `scripts/`).

- Evidence: `.moai/reports/t501/namegrep-counts.txt`

### B.2 Axis 2 — actual execution: REFUTED AT SCALE

Census denominators (stated once, used throughout): **114** is the count of unique function names in the 6 target files — the name-citation denominator (§B.1). **118** is the count of function rows in the coverprofile extract (`.moai/reports/t501/coverage-perfunc.txt`) — the execution denominator; the difference of 4 is method names duplicated across receiver types, which the profile counts once per receiver.

Per-function coverage was measured with the domain's mechanical tool:

- Command: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -coverprofile=/tmp/t501_cover.out ./internal/cli/`
- Observed output (verbatim): `ok  github.com/modu-ai/moai-adk/internal/cli  515.612s  coverage: 80.7% of statements`, rc=0
- Per-function extract (118 functions across the 6 files): `.moai/reports/t501/coverage-perfunc.txt`; run log: `.moai/reports/t501/coverage-run.log`

Measured residue:

- Only 4 of 118 functions are at 0.0%: `terminateCodexProcess` (codex_job_control.go:94 — structurally shadowed: tests swap the `codexTerminateProcess` seam var, so the real body never runs via cancel-path tests), `(realCodexConn).pid` (mcp_codex.go:494), `(codexSessionError).Error` (:602), `(codexSessionError).Unwrap` (:603)
- Only 1 function is below 60%: `codexIDMatches` 55.6% (mcp_codex.go:912)
- All 16 card-named functions measured at or above 75% except `terminateCodexProcess` (0.0): writeCodexRequest 80.0, awaitCodexResponse 92.9, writeCodexEnvelope 75.0, terminateCodexProcess 0.0, codexAwaitJobExit 85.7, loadCodexJobFor 100.0, codexChildArgs 100.0, codexChildEnv 100.0, splitCodexDashDash 100.0, wrapCodexDetail 91.7, codexOverflowMarker 100.0, classifyCodexSkillPath 100.0, codexCountExecutingImports 77.3, codexGateReport 100.0, codexInitOfferGate 100.0, codexGatePrintf 75.0

### B.3 Consequence

The six "clusters" are behaviorally tested via existing idioms (interface fakes `fakeCodexConn` + canned NDJSON scripts in `mcp_codex_test.go`; gated conns asserting transmitted responses in `codex_protocol_liveness_test.go` incl. the AC-CX2-017 deny + unknown-method arms; panel/width-band tests in `doctor_codex_test.go`; seam injections in `codex_init_test.go`; an independent reimplementation cross-checker `codexTestExecImports` for the import counter). Therefore this SPEC mandates NO blanket per-cluster test files. The requirements below cover only the genuine residue.

## C. Requirements (GEARS)

### REQ-CTG-001 — terminateCodexProcess direct test [U3, Priority High]

**When** a test invokes `terminateCodexProcess` directly (codex_job_control.go:94), the test suite shall cover both the `pid <= 0` refusal arm and the successful-kill arm, and **While** the successful-kill arm runs, the test shall guarantee child-process cleanup by registering the kill via `t.Cleanup` before any assertion that could fail.

**Where** the repo gates platform-specific behavior, the test shall be cross-platform (a helper-process re-exec pattern OR `runtime.GOOS` gating; the chosen approach is fixed in plan.md §F M1 and its rationale recorded there).

### REQ-CTG-002 — codexIDMatches table test [U2, Priority High]

**When** a table-driven test enumerates `codexIDMatches` (mcp_codex.go:912), the test suite shall exercise all 5 arms: empty raw, int match, int mismatch, string-id match (the strconv arm — never hit in production tests because codex sends integer ids), and malformed input (both unmarshals fail).

### REQ-CTG-003 — awaitCodexResponse ctx-cancel arm [U2, Priority High]

**When** the context passed to `awaitCodexResponse` (mcp_codex.go:888-892) is already canceled and the fake conn yields non-matching noise lines, the test suite shall observe the canceled-context return taken between reads.

### REQ-CTG-004 — codexCountExecutingImports HTML-comment fixtures [U6, Priority Medium]

**When** fixture content carries an HTML comment in either the multi-line shape (directive after the `-->` close) or the inline single-line shape (`<!-- ... -->` on one line), the test suite shall assert `codexCountExecutingImports` (codex_contract.go:186-194) counts the directive correctly, and the fixtures shall match the fixture style of the existing cross-checker `codexTestExecImports` (codex_init_test.go).

### REQ-CTG-005 — codexGatePrintf erroring-writer arm [U7, Priority Low]

**When** the `io.Writer` passed to `codexGatePrintf` (codex_init.go:56-58) returns an error from its write, the function shall return silently with no panic, and the test suite shall assert this contract.

### REQ-CTG-006 — defaultCodexInitGenerator error arm [U7, Priority Low]

**When** `codexwiring.Wire` fails inside `defaultCodexInitGenerator` (codex_init.go:136-138), the failure shall surface as a wrapped error, and the test suite shall assert the wrap contract (not a bare pass-through).

### REQ-CTG-007 — codexSessionError delegation contract [Priority Medium]

**When** a `codexSessionError` (mcp_codex.go:602-603, constructed in production fail-open paths at :609 and :646) is examined, `Error()` shall delegate to `cause.Error()` and `Unwrap()` shall return `cause`, and the test suite shall assert both delegation arms.

### REQ-CTG-008 — documented-skip record (no test)

**The SPEC artifact itself** (this file, §D) shall carry the skip record for surfaces deliberately left untested, so the next investigator does not re-flag them. No test is written for these surfaces in this SPEC's run phase.

### REQ-CTG-009 — zero production diffs [HARD, Priority High]

**While** the run phase executes, the working tree shall produce zero diffs to non-test `.go` files: every change lands in `*_test.go` files under `internal/cli/` only.

### REQ-CTG-010 — completion re-measurement [Priority High]

**When** the run phase completes, a fresh coverprofile over `./internal/cli/` shall show ZERO remaining 0.0% functions among the 118 measured function rows: all four originally-0.0% surfaces — `terminateCodexProcess`, `(realCodexConn).pid` (REQ-CTG-012), `(codexSessionError).Error`, `(codexSessionError).Unwrap` (REQ-CTG-007) — are covered by this SPEC's test items.

### REQ-CTG-011 — quality-gate constraint [Priority High]

**While** the run phase executes, the touched test files shall pass `go vet`, `golangci-lint run`, and `gofmt` clean, and the full diff shall remain tests-only (consistent with REQ-CTG-009).

### REQ-CTG-012 — realCodexConn.pid direct test [Priority Medium]

**When** a direct method-call test constructs `realCodexConn` in the same package (mcp_codex.go:494-499 — a 3-branch field read: `cmd == nil` → 0; `cmd.Process == nil` → 0; else `cmd.Process.Pid`), the test suite shall assert all 3 branches — `&realCodexConn{}`, `&realCodexConn{cmd: &exec.Cmd{}}`, and a `cmd.Process` obtained via `os.FindProcess(os.Getpid())` — with no subprocess spawn.

## D. Documented-Skip Record

These surfaces are deliberately NOT tested by this SPEC. The record is the deliverable (REQ-CTG-008); a future investigator citing them as "uncovered" should find this section first.

| Surface | Location | Reason |
|---|---|---|
| `writeCodexRequest` / `writeCodexEnvelope` marshal-error arms | mcp_codex.go | Dead-in-practice: envelopes are built from JSON-safe literals; the marshal error path cannot be reached with values the production callers construct. |
| `terminateCodexProcess` FindProcess-error arm | codex_job_control.go:94 | Platform-unreachable on darwin/linux for integer pids: `os.FindProcess` never fails for a valid integer pid on these platforms. |

> Note: `(realCodexConn).pid` was removed from this record in v0.2.0 — its former "live-session-gated" rationale was refuted by direct read of the source (same-package constructible; no subprocess needed). It is now a test item (REQ-CTG-012).

## E. Constraints

1. Tests only — zero diffs to non-test `.go` files (REQ-CTG-009).
2. All code and comments in English (repo standard).
3. `go vet` + `golangci-lint run` clean on touched files.
4. Verification recipe is scoped: `go test ./internal/cli/` only, timeout budget >= 600s (measured 515.6s). NEVER `go test ./...` locally — CI runs the full suite.
5. File placement per plan.md §F — extend existing sibling test files, matching repo style.

## F. Out of Scope

### Out of Scope — functions measured above 60% not named in REQ-CTG-001..007

- No new tests for any function measured above 60% that is not named in REQ-CTG-001 through REQ-CTG-007. This includes the `writeCodexRequest` / `writeCodexEnvelope` marshal arms (see §D skip record).

### Out of Scope — per-cluster test files

- No blanket per-cluster test files duplicating the behaviorally-tested cluster paths (B.3): the `fakeCodexConn` NDJSON-script paths, the gated-conn liveness paths, the doctor panel/width-band paths, the init seam-injection paths, and the `codexTestExecImports` cross-checker path are already covered by existing idioms.

### Out of Scope — production code changes

- No production code changes of any kind, including refactors that would make a surface more testable. If a test reveals a production defect, it is reported as a blocker/finding, not fixed in this SPEC's run phase.

### Out of Scope — new test infrastructure

- No new shared test helpers or frameworks beyond what the existing sibling test files already provide (`fakeCodexConn`, seam vars, `codexTestExecImports`).

## G. Cross-References

- Evidence: `.moai/reports/t501/namegrep-counts.txt` (Axis 1), `.moai/reports/t501/coverage-perfunc.txt` (per-function extract), `.moai/reports/t501/coverage-run.log` (run log)
- Related SPECs: SPEC-CODEX-WIRING-001 (codex wiring, source of `codexwiring.Wire`), SPEC-CODEX-SESSION-MSG-001 (codex broker path)
