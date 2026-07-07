# Acceptance — SPEC-CLI-LINT-COMPLETED-001

Given-When-Then acceptance criteria for adding `completed` to the `terminalStatusEnum` map in
`internal/spec/lint.go`. Because this is a one-line production-code change to working code, the
acceptance model is CHARACTERIZATION + TDD-RED-PROOF: the new test must demonstrate the bug
exists (RED) before the enum entry greenlights it (GREEN).

## §A. Acceptance Criteria Summary

| AC ID | Description | Severity |
|-------|-------------|----------|
| AC-LC-001 | `completed` status produces 0 `StatusGitConsistency` findings | MUST-PASS |
| AC-LC-002 | Existing terminal states (`superseded` / `archived` / `rejected`) behavior unchanged | MUST-PASS |
| AC-LC-003 | New characterization test fails RED without the enum entry (TDD proof) | MUST-PASS |

## §B. Detailed Scenarios

### AC-LC-001 — `completed` status produces 0 drift findings

**Given** a `*SPECDoc` constructed with `status: completed` AND a git fixture producing a non-`completed` git-implied status (per the `setupGitFixture(t, []fixtureCommit{...})` + `os.Chdir(dir)` + `t.Cleanup(...)` pattern in `internal/spec/status_terminal_exemption_test.go` — same pattern as the existing Pattern D/E/F/G tests in that file)
**When** `StatusGitConsistencyRule.Check(doc, nil)` is invoked
**Then** the returned `[]Finding` slice has length 0 (no `StatusGitConsistency` finding emitted)

**Mechanism**: the L984 gate `if terminalStatusEnum[fm.Status] { return nil }` short-circuits before the drift-detection branch at L996 is reached.

**Evidence artifact**: characterization test in `internal/spec/status_terminal_exemption_test.go`. Suggested name: `TestStatusGitConsistency_CompletedIsTerminal_Implemented` (a fifth top-level function alongside the existing Pattern D/E/F/G tests in `status_terminal_exemption_test.go`).

### AC-LC-002 — Existing terminal states behavior unchanged

**Given** three `*SPECDoc` instances, each with valid frontmatter and one of `status: superseded`, `status: archived`, `status: rejected`
**When** `StatusGitConsistencyRule.Check(doc, nil)` is invoked for each
**Then** each returns a `[]Finding` slice of length 0 — identical to the pre-change behavior

**Evidence artifact**: existing characterization tests for `superseded` (Pattern D + E) and `archived` (Pattern F + G) continue to pass without modification — 4 existing top-level functions in `status_terminal_exemption_test.go` unchanged. A new test for `rejected` (`TestStatusGitConsistency_RejectedIsTerminal_Implemented`, Pattern I) is added in the same commit (currently NO existing `rejected` test — consistent with the L962 inline comment "future-proof: not currently used").

### AC-LC-003 — TDD RED proof

**Given** the characterization test from AC-LC-001 has been written (with the `setupGitFixture` + `os.Chdir` precondition) but the `"completed": true,` entry has NOT yet been added to `terminalStatusEnum`
**When** `go test -run TestStatusGitConsistency_CompletedIsTerminal_Implemented ./internal/spec/...` is executed
**Then** the new test FAILS — `StatusGitConsistencyRule.Check` reaches the L996 drift branch and emits a `StatusGitConsistency` finding (slice length ≥ 1). The `setupGitFixture` precondition is REQUIRED here too: without the git fixture, `getGitImpliedStatus` errors → L990-993 graceful-skip → 0 findings → RED proof becomes vacuous.

**Evidence artifact**: manager-develop captures the verbatim RED test failure output (including the emitted finding `Code: "StatusGitConsistency"`) and includes it in the run-phase §E.2 evidence. This is the TDD-discipline proof that the test has value (not a tautology) and that the bug genuinely existed.

## §C. Edge Cases

- **Empty frontmatter `status`**: handled by the L976-979 early-return guard. Unchanged by this SPEC.
- **Missing frontmatter `id`**: handled by the L976-979 early-return guard. Unchanged.
- **Git history unavailable** (e.g., `t.TempDir()` test fixture without `.git`): handled by the L990-993 graceful skip (`getGitImpliedStatus` returns error → `return nil`). Unchanged. Note: with the new enum entry, a `completed`-status doc never reaches L990 — the L984 gate returns first. This is a strict superset of the prior graceful-skip coverage.
- **`completed` status with concurrent map read**: Go `map[string]bool` reads are safe for concurrent access after initialization; the map is package-level and never mutated post-init. `go test -race ./internal/spec/...` is expected clean.
- **Interaction with `StatusEnum` rule**: `completed` is already an accepted status value per the 8-value enum; the schema rule does not flag it. No interaction.

## §D. AC Matrix

| AC ID | Severity | Traceability (spec.md) | Traceability (plan.md) | Verification Method |
|-------|----------|------------------------|------------------------|---------------------|
| AC-LC-001 | MUST-PASS | §A.2 (GEARS primary requirement) | §F M1 (characterization test) | `go test ./internal/spec/...` green; new subtest asserts 0 findings for `status: completed` |
| AC-LC-002 | MUST-PASS | §A.3 (supporting requirement) + Out of Scope (behavior preservation) | §F M1 Done criteria | `go test ./internal/spec/...` green; no existing test modified |
| AC-LC-003 | MUST-PASS | §A.2 (the requirement is only meaningful if the bug exists) | §F M1 TDD discipline (RED step) | RED output captured pre-fix; GREEN output captured post-fix |

## §E. Quality Gate Criteria

- `go test ./internal/spec/...` — exit 0.
- `go test -race ./internal/spec/...` — exit 0 (defensive; map is read-only post-init).
- `go test -cover ./internal/spec/...` — coverage for `internal/spec` package does not decrease (the new test adds coverage to `StatusGitConsistencyRule.Check`'s terminal-state path).
- `golangci-lint run ./internal/spec/...` — clean.
- `moai spec lint .moai/specs/SPEC-CLI-LINT-COMPLETED-001/` — 0 findings on this SPEC (the SPEC itself is well-formed; this is independent of the production-code fix).

## §F. Indirect Verification

The fix is mechanically minimal (one map entry). Indirect verification surfaces:

- **Catalog hash**: `internal/spec/catalog.yaml` does NOT hash `lint.go` production code; the catalog is unaffected. (If a future catalog version hashes `lint.go`, manager-docs reconciles at sync-phase.)
- **CLI smoke**: `go run ./cmd/moai spec lint .moai/specs/SPEC-CLI-LINT-COMPLETED-001/` after M1 should produce 0 findings (this is also a direct AC-LC-001 observable if the SPEC's own frontmatter is the test target, but the unit-test path is the canonical evidence).

## §G. Definition of Done

- [ ] Single-line production change merged to `internal/spec/lint.go` (`"completed": true,` in `terminalStatusEnum`).
- [ ] Characterization test for `status: completed` merged to `internal/spec/status_terminal_exemption_test.go` (Pattern H: `TestStatusGitConsistency_CompletedIsTerminal_Implemented`).
- [ ] `go test ./internal/spec/...` passes.
- [ ] `go test -race ./internal/spec/...` passes.
- [ ] `golangci-lint run ./internal/spec/...` clean.
- [ ] Coverage for `internal/spec` does not decrease.
- [ ] RED output captured and included in run-phase §E.2 evidence (AC-LC-003).
- [ ] Existing terminal-state tests pass without modification (AC-LC-002).
- [ ] Run-phase §E.3 audit-ready signal emitted by manager-develop.
