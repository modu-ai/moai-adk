# Acceptance: SPEC-VERIFY-TEST-001 (Mandatory Test Execution)

> Companion to `spec.md`. Detailed Given-When-Then scenarios, edge cases, quality gates.

---

## Scenario 1 — Go Project, All Tests Pass (REQ-VTE-003, REQ-VTE-004, AC-3)

**Given** evaluator-active is invoked on a SPEC implementation that lives in a Go project
**And** the project has `go.mod` at its root
**And** running `go test ./...` produces 100/100 PASS, 0 FAIL, exit 0
**When** evaluator-active begins scoring the Functionality dimension
**Then** the agent SHALL detect language `go` via `go.mod` presence
**And** the agent SHALL execute `go test ./...` with timeout 600000 ms via the Bash tool
**And** the agent SHALL parse the output: `total=100, passed=100, failed=0, skipped=0`
**And** the Functionality dimension SHALL be eligible for scores up to 1.00
**And** the {evidence} field for Functionality SHALL contain the executed command, the exact counts, and `language_detected: go`.

**Edge case 1a**: Go project has no test files at all (`go test ./...` reports `no test files`). Agent SHALL apply the `no_tests_present: true` flag and -0.20 penalty per REQ-VTE-014.

**Edge case 1b**: Go project has a build error (test compilation fails). Agent SHALL log `build_error: true`, mark Functionality UNVERIFIED, and report the build error file:line in the evidence field.

---

## Scenario 2 — Flaky Test (Failure Ratio ≤ 5%) (REQ-VTE-009, REQ-VTE-011, AC-4)

**Given** a Go project where `go test ./...` produces 30/30 tests run, 1 failed (3.3% failure ratio), exit non-zero
**When** evaluator-active scores Functionality
**Then** the Functionality dimension SHALL be capped at 0.50 (REQ-VTE-009 — any failure caps Functionality)
**And** the agent SHALL emit `flaky_test_warning: true` in the report
**And** the overall verdict SHALL NOT auto-FAIL based on this dimension alone (failure ratio ≤ 5%)
**And** the evidence field SHALL include the failing test's file:line.

**Edge case 2a**: Failure is intermittent (the same test run twice produces different outcomes). The agent does not re-run; the first run's outcome is canonical. Re-execution would be a different SPEC.

**Edge case 2b**: 30/30 with 0 failed but 1 skipped. Skipped is not a failure. Functionality eligible up to 1.00. The evidence field SHALL still report the skipped count.

---

## Scenario 3 — High Failure Ratio (> 5%) (REQ-VTE-010, AC-5)

**Given** a Go project where `go test ./...` produces 30/30 tests run, 5 failed (16.7% failure ratio)
**When** evaluator-active scores any dimension
**Then** the overall evaluation verdict SHALL be FAIL regardless of any other dimension scores
**And** the report SHALL contain `failure_ratio_exceeded: true` with the computed ratio
**And** the Functionality dimension score SHALL be capped at 0.50
**And** the evidence field SHALL list the first 10 failures' file:line references.

**Edge case 3a**: Exactly 5% failure (e.g., 1 fail in 20 tests). Threshold is `> 0.05`, so this is NOT an overall FAIL but the Functionality cap at 0.50 still applies (REQ-VTE-009). Boundary semantics: REQ-VTE-010 says "exceeds 0.05"; 0.05 itself is not exceeding.

**Edge case 3b**: 100% failure (every test fails). Same outcome as Scenario 3 — overall FAIL — but the report SHALL clearly indicate `total_failure: true` to differentiate "broken commit" from "broken test infrastructure."

---

## Scenario 4 — Multi-Language Project (REQ-VTE-007, AC-6)

**Given** a project with both `go.mod` and `package.json` in the root
**And** running `go test ./...` produces 50/50 PASS
**And** running `npm test` produces 20/20 PASS (or whatever the project's test script returns)
**When** evaluator-active scores Functionality
**Then** the agent SHALL detect both languages
**And** the agent SHALL execute both test commands sequentially
**And** the agent SHALL aggregate: `total=70, passed=70, failed=0, languages=[go, typescript]`
**And** the evidence field SHALL include per-language breakdown.

**Edge case 4a**: Go tests pass (50/50) but JS tests have 5/10 failures. Per-language failure ratio: go=0%, js=50%. Aggregate ratio: 5/60 = 8.3%. Per REQ-VTE-007 + REQ-VTE-010, the **per-language max** is the gating value: js=50% > 5% → overall FAIL. Evidence field clearly reports per-language ratios.

**Edge case 4b**: Project has go.mod and pyproject.toml but no actual Python code (orphan marker). `pytest` produces "no tests collected." Apply no-tests handling per Edge case 1a, but only for the Python detection. Go portion proceeds normally.

---

## Scenario 5 — No Tests Present (REQ-VTE-014, AC-7)

**Given** a project with `go.mod` but no `*_test.go` files anywhere in the tree
**When** evaluator-active scores Functionality
**Then** the agent SHALL log `no_tests_present: true` and `language_detected: go`
**And** the agent SHALL apply a -0.20 penalty to the Functionality dimension's pre-penalty score
**And** the agent SHALL NOT execute `go test ./...` (avoids spurious "no test files" outputs across modules)
**And** the evidence field SHALL state explicitly that the project has no tests.

**Edge case 5a**: Project has tests in non-standard naming (e.g., `*_check.go` instead of `*_test.go`). The detection heuristic in REQ-VTE-014 (`find . -name '*_test.*' -o -name 'test_*.*' -o -name '*.test.*'`) is conservative; non-standard names are reported as no_tests_present even if "tests" exist by intent. User can rename files to match conventions or document this as a known limitation.

**Edge case 5b**: Project has tests only under `examples/` directory (typically excluded from `go test ./...`). Detection SHALL find them (heuristic matches filenames anywhere); execution SHALL proceed; but `go test ./...` will not run them. The evidence field reports the discrepancy.

---

## Scenario 6 — Tool Missing from PATH (REQ-VTE-015, AC-8)

**Given** a project with `Cargo.toml` (Rust)
**And** the `cargo` binary is not installed in the evaluator's environment
**When** evaluator-active scores Functionality
**Then** the agent SHALL detect `language_detected: rust`
**And** the agent SHALL attempt `which cargo` (or the matrix-prescribed presence check)
**And** the agent SHALL log `tool_skipped: cargo`
**And** the agent SHALL mark the Functionality dimension as UNVERIFIED for Rust
**And** the agent SHALL continue to other detected languages if any.

**Edge case 6a**: All detected languages have missing tools. Functionality SHALL be UNVERIFIED for the entire dimension. Other dimensions (Security, Craft, Consistency) proceed normally. Overall verdict UNVERIFIED unless other dimensions force FAIL.

**Edge case 6b**: Tool is on PATH but produces version mismatch error (e.g., `pyproject.toml` requires Python 3.12 but only 3.10 is available). Treated as `tool_skipped` with the error message captured in evidence. Same outcome as Edge case 6a.

---

## Scenario 7 — Test Suite Timeout (REQ-VTE-006, AC-9)

**Given** a project whose test suite exceeds 600 seconds (10 minutes)
**When** evaluator-active executes the test command
**Then** the Bash tool runtime SHALL terminate the process at the 10-minute ceiling
**And** the agent SHALL emit a structured blocker report:
  ```
  ## Missing Inputs
  | Parameter | Type | Expected Values | Rationale |
  |-----------|------|-----------------|-----------|
  | test_subset | string | comma-separated test names or "skip" | Test suite exceeded 10-minute timeout; user must select subset |
  ```
**And** the Functionality dimension SHALL be marked UNVERIFIED with reason `timeout_exceeded`
**And** the agent SHALL NOT auto-retry with a longer timeout.

**Edge case 7a**: Timeout is approached but the test process exits exactly at 10 minutes with output buffered. The Bash tool's behavior is implementation-dependent; the agent treats any exit with `signal_killed: true` or partial output as timeout per REQ-VTE-006. Evidence captures whatever was emitted before termination.

---

## Scenario 8 — Permission Elevation Required (REQ-VTE-018, AC-13)

**Given** a project whose tests write to disk (e.g., create a fixture database, generate snapshot files)
**And** evaluator-active is running in `permissionMode: plan` (read-only)
**When** the test command attempts a write operation
**Then** the test SHALL fail with permission denied OR the agent's pre-flight check SHALL detect the write requirement (e.g., `test:db` script in package.json)
**And** the agent SHALL log `permission_required: write` with the specific path/operation
**And** the Functionality dimension SHALL be marked UNVERIFIED
**And** the agent SHALL NOT silently elevate to a write-capable permission mode.

**Edge case 8a**: Test writes only to `/tmp` or system-allowed temporary locations. The test passes (system-level allow). The agent does not detect `permission_required: write` because the writes succeeded under plan mode's actual capability set. This is acceptable — the SPEC's intent is to prevent silent escalation, not to over-detect.

---

## Scenario 9 — Evidence Field Discipline (REQ-VTE-012, REQ-VTE-013, AC-12)

**Given** evaluator-active completes a Functionality dimension scoring with one or more failures
**When** the report is rendered
**Then** the {evidence} field SHALL contain at minimum:
  - The exact test command(s) executed
  - Numeric counts: `total: N, passed: P, failed: F, skipped: S` (no rounding, no approximation)
  - Language(s) detected: a list, not summary
  - For the first 10 failures: `file:line - test_name - error excerpt (max 200 chars)`
**And** the {evidence} field SHALL NOT contain phrases that summarize away failures, such as:
  - "minor test issues"
  - "a few tests failed"
  - "tests mostly pass"
**And** an automated lint over the evidence field SHALL flag forbidden phrases for human review.

**Edge case 9a**: There are more than 10 failures. Evidence reports the first 10 by file path order; an additional summary line states `... and N more failures (full list in test output)`. The verdict still uses the full count, not the truncated 10.

---

## Scenario 10 — Profile Compatibility (REQ-VTE-016, REQ-VTE-017, AC-10)

**Given** evaluator-active is invoked under each of the four profiles (default, strict, lenient, frontend)
**When** the same test scenario runs (e.g., 5/30 failures, 16.7% ratio)
**Then** under `default` profile: overall verdict FAIL per REQ-VTE-010
**And** under `strict` profile: overall verdict FAIL per REQ-VTE-010 with stricter language in the report ("must-pass non-negotiable")
**And** under `lenient` profile: overall verdict NOT auto-FAIL; emit `test_suite_warning` instead of overall FAIL; Functionality still capped at 0.50
**And** under `frontend` profile: same as default unless explicitly overridden in profile metadata.

**Edge case 10a**: Profile file is missing entirely (e.g., user typo in SPEC frontmatter `evaluator_profile: nonexistent`). evaluator-active falls back to built-in default per existing behavior (`evaluator-active.md:80-89`). New must-pass criterion still applies via the built-in default.

---

## Scenario 11 — Differential Measurement (AC-12)

**Given** a curated fixture set of 10 historical commits known to break at least one test
**And** the pre-SPEC evaluator-active body (rolled back)
**When** the pre-SPEC evaluator-active is run on each of the 10 commits
**Then** record `baseline_caught: B` (number of commits where verdict was FAIL).
**And** when the post-SPEC evaluator-active is run on the same 10 commits
**Then** record `pilot_caught: P` (number of commits where verdict is FAIL).
**And** the improvement `(P - B) / 10` SHALL be at least 30 percentage points.

**Edge case 11a**: All 10 commits are caught by both baseline and pilot (B=10, P=10, improvement = 0). The differential metric is trivially 0%, not failure. Re-curate fixtures: select commits where the breakage is subtle (e.g., changes that pass type-check but break runtime tests). The fixture curation, not the SPEC, is the limiting factor.

**Edge case 11b**: Pilot catches 8/10 baseline catches 6/10 (improvement 20 percentage points, below 30 threshold). Document the result; do not auto-rollback. User decision: accept the smaller improvement, or extend the fixture curation, or revisit the trigger thresholds.

---

## Scenario 12 — Matrix Dry-Run (AC-2, AC-11)

**Given** the verification-test-matrix.md file is committed
**When** the audit test in `internal/template/verification_matrix_audit_test.go` runs
**Then** the test SHALL parse the matrix and assert: 16 entries (one per official language), no duplicate language names, every command's first token is a known binary identifier (e.g., `go`, `pytest`, `npm`, `cargo`).
**And** the dry-run availability matrix in `.moai/reports/verify-test-dry-run-{DATE}/availability.md` SHALL be present and timestamped.

**Edge case 12a**: A new language is added to the official 16-language list (e.g., Zig). The audit test SHALL detect 17 entries (or 15 if removed) and FAIL until the matrix is updated. This is a guard against drift between the matrix and the supported-language list.

---

## Quality Gates (TRUST 5 Mapping)

- **Tested**: Scenarios 1-11 each have an automated or semi-automated verification (fixture corpus in M5; audit test in M1/M2; differential measurement in M5).
- **Readable**: New "Mandatory Test Execution" section in evaluator-active body is procedural and concrete (no exhortation). Matrix file is structured tabular data.
- **Unified**: Hard threshold pattern (>5% failure = overall FAIL) follows the same shape as the existing Security FAIL hard threshold. Profile updates are uniformly structured.
- **Secured**: REQ-VTE-018 prevents silent permission escalation. Evidence field discipline (Scenario 9) prevents detail loss that could mask security-relevant test failures.
- **Trackable**: Every test execution emits structured log fields (`test_command`, `failure_ratio`, `flaky_test_warning`, `failure_ratio_exceeded`, `tool_skipped`, `permission_required`, etc.) consumable by the existing observability log.

---

## Definition of Done

- [ ] M0 — Open questions resolved; user approved decisions
- [ ] M1 — verification-test-matrix.md authored and audit test passing
- [ ] M2 — Dry-run availability matrix recorded
- [ ] M3 — evaluator-active body updated with Mandatory Test Execution section
- [ ] M4 — All four evaluator profiles updated and consistent
- [ ] M5 — Seven fixture scenarios verified; differential measurement showing >= 30 percentage points improvement
- [ ] M6 — Cross-references in zone-registry, CLAUDE.md §7, boundary-verification.md
- [ ] All Scenarios 1-12 satisfied
- [ ] plan-auditor verdict on this SPEC: PASS
- [ ] User approval (AP-4) on differential measurement results

---

**Status**: draft — pending plan-auditor review and pilot execution
