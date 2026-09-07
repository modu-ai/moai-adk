# SPEC-CODEX-COVER-RESIDUAL-001 — Acceptance Criteria

## A. Format contract

Every criterion below is a Given-When-Then scenario with a binary, mechanically-checkable outcome. Each carries:

- a **class** — `RB` (release-blocking) or `RG` (regression-guard), per `.claude/rules/moai/development/verification-completeness.md` §2.1;
- an **adoption cell** — either a RED-now observation (the criterion measured red on the pre-work tree) or, for an absence guard, a named mutant, because an absence guard is satisfied for free when the feature is missing and RED-now proves nothing about it;
- a **verification command** plus the expected observation.

Every RED-now command is a single read-only invocation with no pipe, redirection, or chaining, pinned to tree `bf779ecf2`.

### A.1 The empty-sweep hazard, stated once

A Go test selector that matches zero tests **exits 0 and prints `ok`**. It is byte-indistinguishable from a run in which everything passed. Consequently:

- No criterion here is adopted on the premise "the test does not exist yet, so the suite is red". That premise is false, and it is the exact defect `verification-completeness.md` §2.1 records (nine release-blocking criteria resting on it).
- AC-CCR-009 counts the tests that actually ran and requires an exact non-zero count, so an empty sweep fails rather than passing quietly.
- Every per-test criterion is adopted on a **mutant**, never on absence.

## B. AC matrix

| AC | Requirement | Class | Adoption basis |
|---|---|---|---|
| AC-CCR-001 | REQ-CCR-001 | RB | mutant M1 |
| AC-CCR-002 | REQ-CCR-001 | RB | mutant M1b |
| AC-CCR-003 | REQ-CCR-002 | RB | mutant M2 |
| AC-CCR-004 | REQ-CCR-003 | RB | mutant M3a / M3b |
| AC-CCR-005 | REQ-CCR-004 | RB | mutant M4 |
| AC-CCR-006 | REQ-CCR-005 | RB | mutant M5a / M5b |
| AC-CCR-007 | REQ-CCR-007 | RB | mutant M6 (absence guard) |
| AC-CCR-008 | REQ-CCR-009 | RB | RED-now (measured 0.0% / 66.7%) |
| AC-CCR-009 | REQ-CCR-008 | RB | exact-count sweep |
| AC-CCR-010 | REQ-CCR-008 | RB | per-item mutant ledger |
| AC-CCR-011 | REQ-CCR-010 | RG | quality gate |
| AC-CCR-012 | REQ-CCR-007 | RG | undecidable disposition (post-close only) |

Requirement coverage: REQ-CCR-001 → AC-001/002 · REQ-CCR-002 → AC-003 · REQ-CCR-003 → AC-004 · REQ-CCR-004 → AC-005 · REQ-CCR-005 → AC-006 · REQ-CCR-007 → AC-007 · REQ-CCR-008 → AC-009/010 · REQ-CCR-009 → AC-008 · REQ-CCR-010 → AC-011. REQ-CCR-006 (the documented-skip record) is satisfied by spec.md §D existing and is verified in the Definition of Done (§F), not by a runtime criterion.

## C. Criteria

### AC-CCR-001 — malformed stdin fails open [RB]

**Given** a throwaway cobra command wired to `runCodexReviewGate` with in-memory streams,
**When** it is executed with stdin `{not json`,
**Then** `Execute()` returns nil, stdout decodes as JSON with no `decision: "block"`, and stderr contains `codex-review-gate:`.

- Verify: `go test -count=1 -run 'TestRunCodexReviewGate_InvalidStdinFailsOpen' -v ./internal/cli/`
- Expect: one `--- PASS: TestRunCodexReviewGate_InvalidStdinFailsOpen` line, `ok` on the package line.
- Adopted via mutant **M1** (§D).

### AC-CCR-002 — empty stdin fails open [RB]

**Given** the same command harness,
**When** it is executed with empty stdin,
**Then** `Execute()` returns nil and stdout decodes to an ALLOW.

- Verify: `go test -count=1 -run 'TestRunCodexReviewGate_EmptyStdinFailsOpen' -v ./internal/cli/`
- Expect: one `--- PASS:` line for that test.
- Adopted via mutant **M1** (§D).

### AC-CCR-003 — happy path allows without consulting codex [RB]

**Given** a temp project written by `writeWorkflowYAML` carrying `workflow:\n  codex:\n    review_gate:\n      enabled: false`, `withChangeDetector(t, true)`, and `codexLookPath` swapped via `withCodexLookPath` to a function that calls `t.Fatal` if reached,
**When** the command is executed with a well-formed payload `{"session_id":…,"project_dir":<dir>}`,
**Then** `Execute()` returns nil, stdout decodes to an ALLOW, and the `t.Fatal` guard is never reached.

`withChangeDetector(t, true)` is not decoration and must not be dropped as redundant. It is inert on the green path — with the gate disabled, `HandleCodexReviewGate` returns at step 1 (codex_review_gate.go:69) long before the detector is consulted at step 3 — but it is what makes mutant M2 able to fire. Without it, the production detector runs `git status` against the `t.TempDir()` that `writeWorkflowYAML` returns, which is not a git repository, so it returns false (codex_review_gate.go:129-132) and the mutated handler ALLOWs at step 3 before ever reaching `codexLookPath`. The precedent pairs them for exactly this reason, with the comment "even with changes present…" (codex_review_gate_test.go:41), and the in-repo assertion that a non-git dir yields false is codex_review_gate_test.go:312-314.

- Verify: `go test -count=1 -run 'TestRunCodexReviewGate_HappyPathAllow' -v ./internal/cli/`
- Expect: one `--- PASS:` line for that test.
- Adopted via mutant **M2** (§D).

### AC-CCR-004 — handler error fails open, with the reason on stderr [RB]

**Given** the gate enabled in the temp project's `workflow.yaml`, `withChangeDetector(t, true)`, and **all three** codex seams swapped together under one `t.Cleanup` — `codexRunner = stubCodexRunner{}`, `codexLookPath = func(string) (string, error) { return "/fake/codex", nil }`, `codexSession = &fakeCodexSession{startErr: errFakeCodexCrash}` — copied verbatim from `TestReviewGate_FailOpenOnCodexError` (codex_review_gate_test.go:154-158),
**When** the command is executed with a well-formed payload naming that project,
**Then** `Execute()` returns nil, stdout decodes to an ALLOW, and stderr contains `codex-review-gate: error:`.

The `codexLookPath` swap is mandatory, not one of three interchangeable seams. The production default is `var codexLookPath = exec.LookPath` (mcp_codex.go:368) and `HandleCodexReviewGate` consults it at step 4 (codex_review_gate.go:78) **before** the session is started, so an unswapped fixture performs a real PATH lookup for a `codex` binary. On a host that has one the lookup succeeds and the test passes; on a host that does not, step 4 returns ALLOW with a nil error (codex_review_gate.go:80), `gateErr` is nil, the RunE takes the success path, no `codex-review-gate: error:` reaches stderr, and the test fails. That is a local-green / CI-red split, and it would also make the M3a/M3b RED evidence recorded on a developer machine unreproducible in CI.

The stderr assertion is load-bearing, not decorative: both the error branch and the success branch write byte-identical `{}` to stdout, so a stdout-only test passes under the naive mutant and asserts nothing about this arm (spec.md §D.2).

- Verify: `go test -count=1 -run 'TestRunCodexReviewGate_HandlerErrorFailsOpen' -v ./internal/cli/`
- Expect: one `--- PASS:` line for that test.
- Adopted via mutants **M3a** and **M3b** (§D). The known-vacuous mutant is recorded in §D as M3-vac and is not counted as adoption evidence.

### AC-CCR-005 — BLOCK propagates through the RunE [RB]

**Given** the gate enabled, `withChangeDetector(t, true)`, and `withCodexSession(t, codexSessionScript("- [P1] found issues"))`,
**When** the command is executed with a well-formed payload naming that project,
**Then** stdout decodes to an object whose `decision` is `"block"` and whose `reason` is a non-empty string.

- Verify: `go test -count=1 -run 'TestRunCodexReviewGate_BlockVerdictPropagates' -v ./internal/cli/`
- Expect: one `--- PASS:` line for that test.
- Adopted via mutant **M4** (§D).

### AC-CCR-006 — the pid nil-guard covers all three arms [RB]

**Given** direct in-package construction of `codexSessionHandle`,
**When** `pid()` is called on a nil receiver, on `&codexSessionHandle{}`, and on `&codexSessionHandle{conn: &fakeCodexConn{}}`,
**Then** the results are `0`, `0`, and `fakeCodexConnPID` (424242) respectively, with no panic on the nil receiver.

- Verify: `go test -count=1 -run 'TestCodexSessionHandlePid' -v ./internal/cli/`
- Expect: one `--- PASS:` line for that test.
- Adopted via mutants **M5a** and **M5b** (§D).

### AC-CCR-007 — zero non-test production diffs [RB, absence guard]

**Given** the base tree `bf779ecf2`,
**When** the union of committed and uncommitted change is enumerated at run-phase close and again at sync-phase close,
**Then** no path in that union is a `.go` file outside `*_test.go`.

Two single-invocation commands, read together as the union:

- `git diff --name-only bf779ecf2..HEAD`
- `git status --short`

Expected observation: every `.go` path appearing in either output ends in `_test.go`. Non-`.go` paths (SPEC artifacts, evidence files) are expected and are not a failure.

**Adoption — mutant M6, not RED-now.** On the pre-work tree both commands are empty, so this criterion is GREEN at arrival and RED-now would prove nothing: an absence guard is satisfied for free precisely when nothing has been done. It is adopted by observing it fire — append a newline to `internal/cli/codex_review_gate.go`, run `git status --short`, confirm the file is listed and classified as a non-test `.go` violation, then `git checkout -- internal/cli/codex_review_gate.go`. Record the observed output of the fired state; a guard never seen firing is not adopted.

### AC-CCR-008 — per-function coverage on the final tree [RB]

**Given** a fresh coverprofile over `./internal/cli/` on the final tree, taken with the same env-scrubbed command as spec.md §B.1,
**When** `go tool cover -func` is read for the two target rows,
**Then** `runCodexReviewGate` reports ≥ 90.0% and `(codexSessionHandle).pid` (mcp_codex.go:701) reports 100.0%.

The package-wide figure is recorded alongside but is not part of this criterion (spec.md §F).

**RED-now cell — axis 1.**

- Command: `grep runCodexReviewGate .moai/reports/t519/coverage-baseline-targets.txt`
- Verbatim stdout: `github.com/modu-ai/moai-adk/internal/cli/codex_review_gate.go:183:		runCodexReviewGate			0.0%`
- Exit code: `0`
- Tree: `bf779ecf2`
- Red for the stated reason: the measured value is 0.0%, below the 90.0% threshold. The command exits 0 because the grep matched; the redness is in the value, not the exit code.

**RED-now cell — axis 2.**

- Command: `grep mcp_codex.go:701 .moai/reports/t519/coverage-baseline-targets.txt`
- Verbatim stdout: `github.com/modu-ai/moai-adk/internal/cli/mcp_codex.go:701:			pid					66.7%`
- Exit code: `0`
- Tree: `bf779ecf2`
- Red for the stated reason: 66.7% is below the 100.0% threshold; the uncovered third of the three statements is the `return 0` guard arm.

**Green path.** M1 flips axis 1 (the five RunE tests reach every statement except S1); M2 flips axis 2 (the three-arm pid test reaches `return 0`). Neither depends on any file this SPEC does not touch, so neither green path runs through "someone fixes something unrelated".

### AC-CCR-009 — the new tests actually ran [RB]

**Given** the final tree,
**When** the six new test functions are run by name,
**Then** exactly six top-level PASS lines are observed and none of `[no tests to run]`, `[no test files]`, or a `FAIL` line appears.

- Verify: `go test -count=1 -run 'TestRunCodexReviewGate_InvalidStdinFailsOpen|TestRunCodexReviewGate_EmptyStdinFailsOpen|TestRunCodexReviewGate_HappyPathAllow|TestRunCodexReviewGate_HandlerErrorFailsOpen|TestRunCodexReviewGate_BlockVerdictPropagates|TestCodexSessionHandlePid' -v ./internal/cli/`
- Expect: six lines matching `^--- PASS: Test`, and the package line `ok`.
- The count is the criterion. A selector typo yields fewer than six PASS lines while still exiting 0, so reading the exit code alone would pass a run that swept nothing. Match on `^--- PASS: Test` with no end anchor — a Go PASS line ends in ` (0.00s)`, so a `$` anchor matches nothing and silently reports zero.

### AC-CCR-010 — every test item carries a mutant, RED then GREEN [RB]

**Given** the mutant ledger in §D,
**When** each row is worked,
**Then** the run-phase evidence records, per row: the mutant's exact edit, the verbatim failing output observed under it, the passing output observed after reverting it, and a clean `git status --short` for production sources at the commit that lands the test.

- Verify: `.moai/specs/SPEC-CODEX-COVER-RESIDUAL-001/progress.md` §E.2 carries one evidence block per ledger row, each with a verbatim `--- FAIL:` line.
- A row whose recorded mutant turned out not to change the verdict is recorded as vacuous with a replacement mutant named, per REQ-CCR-008. A vacuous mutant left in the ledger unreplaced fails this criterion.
- **M5a is the one row whose RED does not look like the others.** Deleting the nil guard makes the nil-receiver call dereference a nil pointer, so the failure arrives as a runtime panic that also aborts the remaining tests in the same binary. Go still emits the `--- FAIL:` line ahead of the panic trace, so the criterion is satisfiable as written — but run M5a in isolation (`-run 'TestCodexSessionHandlePid'`) and record the panic trace alongside the `--- FAIL:` line, so the evidence block is not mistaken for a truncated or corrupted capture.

### AC-CCR-011 — quality gate clean [RG]

**Given** the final tree,
**When** the toolchain is run over the touched package,
**Then** `go vet ./internal/cli/` exits 0, `golangci-lint run` reports no new finding attributable to the touched files, and `gofmt -l internal/cli/` lists none of them.

Classified `RG` rather than `RB`: the lint baseline is inherited and this SPEC cannot establish a clean pre-work RED for findings it did not create. Pre-existing findings are reported as baseline, never as this SPEC's failure.

### AC-CCR-012 — sync close is the last write [RG]

**Given** the sync-phase close commit,
**When** the tree is inspected after the close,
**Then** `git diff --name-only <sync-close-sha>..HEAD` lists no `.go` file.

Classified `RG` with the **undecidable disposition** (`verification-completeness.md` §2.1): it cannot be exercised before the close exists, so it has no reproducible RED cell and is not release-blocking. It is a guard against the specific failure of landing code after documenting the tree as closed.

## D. Mutant ledger

Each mutant is an edit to a production source file, applied only long enough to observe the failure, then reverted. No mutant is present in any commit (REQ-CCR-007, REQ-CCR-008).

| ID | Target | Edit | Expected effect |
|---|---|---|---|
| M1 | codex_review_gate.go:186 | delete the `fmt.Fprintf` stderr diagnostic line | AC-CCR-001's stderr assertion fails. Does **not** fire AC-CCR-002, which asserts only a nil `Execute()` error and an ALLOW on stdout — hence M1b. |
| M1b | codex_review_gate.go:188 | replace `return emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{})` with `return err` | `Execute()` returns non-nil on both the malformed and the empty payload, so AC-CCR-001 **and** AC-CCR-002 fail. `err` is in scope inside the `if err != nil` block, so the edit compiles. This is AC-CCR-002's adoption basis. |
| M2 | codex_review_gate.go:191 | replace `enabled := readCodexReviewGateEnabled(projectDir)` with `enabled := true` | the codex path is entered, the `t.Fatal` guard in `withCodexLookPath` fires, AC-CCR-003 fails. **Detectable only because AC-CCR-003's fixture carries `withChangeDetector(t, true)`** — without it the mutated handler ALLOWs at step 3 on the non-git temp dir and this mutant is vacuous. Confirm the RED is the `t.Fatal` guard message, not a passing run. |
| M3a | codex_review_gate.go:195 | delete the `fmt.Fprintln` stderr diagnostic line | AC-CCR-004's stderr assertion fails |
| M3b | codex_review_gate.go:196 | replace `return emitHookOutput(...)` with `return gateErr` | `Execute()` returns an error, AC-CCR-004's fail-open assertion fails |
| M3-vac | codex_review_gate.go:196 | emit `out` instead of `&hook.HookOutput{}` | **known vacuous — do not use as adoption evidence.** `HandleCodexReviewGate` returns its empty `allow` value alongside the error, so both forms serialize to `{}` and no stdout assertion can separate them (spec.md §D.2). |
| M4 | codex_review_gate.go:201 | replace `emitHookOutput(cmd.OutOrStdout(), out)` with `emitHookOutput(cmd.OutOrStdout(), &hook.HookOutput{})` | the BLOCK is replaced by an ALLOW, AC-CCR-005 fails |
| M5a | mcp_codex.go:702-704 | delete the `if h == nil \|\| h.conn == nil { return 0 }` guard | the nil-receiver arm panics, AC-CCR-006 fails |
| M5b | mcp_codex.go:703 | replace `return 0` with `return -1` | both zero-arms mismatch, AC-CCR-006 fails |
| M6 | internal/cli/codex_review_gate.go | append a trailing newline | `git status --short` lists a non-test `.go` path, AC-CCR-007 fires |

**A mutant that fails to produce RED is itself a finding.** It means the test does not assert what it appears to assert. Record the non-firing mutant, name a replacement, and do not quietly drop the row — M3-vac above is the worked example of this, recorded in advance so it is not rediscovered as a surprise.

## E. Quality gates

| Gate | Threshold | Blocking |
|---|---|---|
| New tests pass | 6/6 top-level PASS (AC-CCR-009) | yes |
| `runCodexReviewGate` coverage | ≥ 90.0% (AC-CCR-008) | yes |
| `(codexSessionHandle).pid` coverage | 100.0% (AC-CCR-008) | yes |
| Non-test `.go` diff | empty union (AC-CCR-007) | yes |
| Mutant ledger complete | every row RED-then-GREEN (AC-CCR-010) | yes |
| `go test ./internal/cli/` | pass, timeout ≥ 600s | yes |
| `go vet` / `golangci-lint` / `gofmt` | no new finding (AC-CCR-011) | no — baseline-inherited |
| Package coverage | recorded, not gated (spec.md §F) | no |

## F. Definition of Done

- [ ] `internal/cli/codex_review_gate_wiring_test.go` exists, carrying the five `TestRunCodexReviewGate_*` functions and its own command constructor under a name distinct from `newGateCmd` (constraint E.4).
- [ ] `internal/cli/mcp_codex_test.go` carries `TestCodexSessionHandlePid`.
- [ ] AC-CCR-001 through AC-CCR-010 observed PASS on the final tree, with commands and verbatim outputs recorded in progress.md §E.2.
- [ ] Every §D ledger row has recorded RED and GREEN evidence; any vacuous mutant is recorded as such with a named replacement.
- [ ] The union non-test `.go` diff is empty at run-phase close and again at sync-phase close.
- [ ] spec.md §D (the documented-skip record) is present and names S1 with its reason — satisfying REQ-CCR-006.
- [ ] AC-CCR-011 reported with baseline attribution; pre-existing lint findings named as inherited, not as this SPEC's.
- [ ] No production `.go` file modified, and no new production seam introduced.
