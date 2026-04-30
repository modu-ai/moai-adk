# Plan: SPEC-VERIFY-TEST-001 (Mandatory Test Execution)

> Companion to `spec.md`. Implementation plan with milestones, technical approach, risks.

---

## Implementation Plan

### Milestone M0 — OQ Resolution (Priority: Critical)

**Goal**: Resolve open questions OQ-1 through OQ-6 before code changes.

**Tasks**:
1. **OQ-1 resolution**: Decide single source of truth for `test_suite_passing` enforcement. Recommend: hard threshold lives in evaluator-active body (REQ-VTE-010 — analogous to Security FAIL); default profile lists it as a "must-pass referenced from agent body" entry. Confirm with user via AskUserQuestion.
2. **OQ-2 resolution**: Specify per-language failure ratio with secondary aggregate. Document the formula in matrix file.
3. **OQ-3 resolution**: Tabulate parser strategy per language: regex-based for go/python/rust (machine-readable summary), heuristic line-grep for cpp/scala (last "passed/failed" line).
4. **OQ-4 resolution**: Curate `.moai/reports/verify-test-fixtures-{DATE}/` with 10 known-broken commits sampled from git history.
5. **OQ-5 resolution**: Confirm flat -0.20 penalty for v1; defer scaling.
6. **OQ-6 resolution**: Define Python test command priority: `pytest` (if importable), then `python -m pytest`, then `python -m unittest discover`. Probe order baked into matrix.

**Exit criteria**: All 6 OQs answered in `.moai/reports/verify-test-oq-{DATE}/decisions.md`; user approves via AskUserQuestion.

---

### Milestone M1 — Test Command Matrix (Priority: Critical)

**Goal**: REQ-VTE-002 — author the 16-language matrix.

**Tasks**:
1. Create `.claude/rules/moai/quality/verification-test-matrix.md` with the table from research.md §2.6.
2. For each language, document:
   - Project marker (file presence check)
   - Primary test command (string)
   - Fallback commands (priority list)
   - Pass exit code (always 0)
   - Output parser approach (regex pattern or heuristic)
   - Optional coverage flag (informational only — not used by this SPEC)
3. Mirror to `internal/template/templates/.claude/rules/moai/quality/verification-test-matrix.md` (Template-First).
4. Run `make build`.
5. Add validation test (`internal/template/verification_matrix_audit_test.go`) that parses the matrix and asserts: 16 entries, no duplicate language names, every command starts with a known binary name.

**Exit criteria**: Matrix file exists; template-mirrored; audit test passes.

---

### Milestone M2 — Matrix Dry-Run Validation (Priority: High)

**Goal**: REQ-VTE-002 + AC-2, AC-11 — verify matrix commands are syntactically valid and binaries exist where expected.

**Tasks**:
1. For each of the 16 languages, attempt:
   - `which <test-binary>` (e.g., `which go`, `which pytest`) — record availability.
   - `<test-binary> --help` (or `--version`) — confirm it executes and is the expected tool.
2. Document availability matrix in `.moai/reports/verify-test-dry-run-{DATE}/availability.md`.
3. For unavailable binaries, document expected install instructions per OS (macOS, Linux, Windows).
4. The dry-run is run on the developer machine and on a CI runner (Ubuntu) to catch environment-specific gaps.

**Exit criteria**: Availability matrix shows at least the local-developer machine has go, python, typescript, rust toolchains. Other languages are documented but not blocking.

---

### Milestone M3 — evaluator-active Body Update (Priority: Critical)

**Goal**: REQ-VTE-001, REQ-VTE-003 through REQ-VTE-015 — add the "Mandatory Test Execution" section.

**Tasks**:
1. Update `.claude/agents/moai/evaluator-active.md` body. Add new section (between "Skeptical Evaluation Mandate" and "Evaluation Dimensions" — placement preserves logical flow):

   ```markdown
   ## Mandatory Test Execution

   Before scoring the Functionality dimension, you MUST execute the project's test suite using the language toolchain detected from project markers. See `.claude/rules/moai/quality/verification-test-matrix.md` for the per-language command table.

   Procedure:
   1. Detect language(s) by checking marker files (e.g., go.mod, pyproject.toml).
   2. For each detected language, run the matrix-prescribed test command via Bash with timeout 600000ms (10 minutes maximum, runtime ceiling).
   3. Parse the output for total/passed/failed/skipped counts.
   4. Apply scoring rules:
      - failed_count = 0 → Functionality may score up to 1.00
      - failed_count > 0 AND failure_ratio <= 0.05 → Functionality capped at 0.50, emit flaky_test_warning
      - failure_ratio > 0.05 → Overall verdict FAIL, regardless of other dimensions
   5. If no language marker detected → log no_tests_present, apply -0.20 penalty.
   6. If timeout exceeded → emit blocker report (UNVERIFIED), do not retry.
   7. If test binary missing from PATH → log tool_skipped:<command>, mark Functionality UNVERIFIED.
   8. If permission elevation required (test writes to disk) → log permission_required:write, fall through to UNVERIFIED.

   Evidence requirement: The Functionality dimension's evidence field MUST contain the executed command, exact passed/failed/skipped counts, language(s) detected, and per-failure file:line for the first 10 failures. Approximation, summarization-away, and rounding are prohibited.
   ```

2. Update the "Evaluation Dimensions" table to add a footnote clarifying that Functionality dimension is gated by test execution (REQ-VTE-009).

3. Update the "Skeptical Evaluation Mandate" HARD RULES to add: "Functionality PASS without test execution evidence is a hard violation and MUST be reported as UNVERIFIED, not PASS."

4. Mirror to `internal/template/templates/.claude/agents/moai/evaluator-active.md`. Run `make build`.

**Exit criteria**: Body updated; mirrored; rebuilt; no regressions in existing evaluator-active tests.

---

### Milestone M4 — Profile Updates (Priority: High)

**Goal**: REQ-VTE-016, REQ-VTE-017 — update all four evaluator profiles.

**Tasks**:
1. `.moai/config/evaluator-profiles/default.md`: Add `test_suite_passing` to Must-Pass Criteria section. Add reference to the matrix file.
2. `.moai/config/evaluator-profiles/strict.md`: Inherit default behavior; add stricter language ("test_suite_passing is non-negotiable; no force-pass override").
3. `.moai/config/evaluator-profiles/lenient.md`: Downgrade to `test_suite_warning`. Document opt-out semantics.
4. `.moai/config/evaluator-profiles/frontend.md`: Inherit default; add note that frontend projects often have additional E2E test suites that may exceed the 10-minute timeout — recommend running unit tests only at evaluator-active phase, deferring E2E to manager-quality.
5. Mirror all four to `internal/template/templates/.moai/config/evaluator-profiles/`. Run `make build`.

**Exit criteria**: Four profile files updated and consistent with each other; mirrored; rebuilt.

---

### Milestone M5 — Fixture and Differential Measurement (Priority: High)

**Goal**: REQ-VTE-009 through REQ-VTE-015 functional verification + AC-12 differential measurement.

**Tasks**:
1. Create test fixture directories under `.moai/reports/verify-test-fixtures-{DATE}/`:
   - `passing-go/` — small Go project, all tests pass.
   - `flaky-go/` — Go project with 1 failing test out of 30 (3.3% failure ratio).
   - `broken-go/` — Go project with 5 failing tests out of 30 (16.7% failure ratio).
   - `multi-lang/` — directory with go.mod + package.json, both with tests.
   - `no-tests/` — Go project with no _test.go files.
   - `missing-tool/` — Project flagged for a language whose binary is not installed (synthetic).
   - `slow-tests/` — Project with a `time.Sleep(11 * time.Minute)` in tests (timeout trigger).
2. Run evaluator-active against each fixture; verify expected outcomes per acceptance.md scenarios.
3. AC-12 differential measurement:
   - **Baseline**: Capture evaluator-active output on 10 historical known-broken commits with the **pre-SPEC** evaluator-active body (revert temporarily).
   - **Pilot**: Apply M3 changes; re-run evaluator-active on the same 10 commits.
   - Compute: how many of the 10 broken commits did the baseline catch (FAIL verdict)? How many did the pilot catch?
4. Document in `.moai/reports/verify-test-pilot-results-{DATE}/`.

**Exit criteria**: All fixture scenarios produce expected outcomes. Differential improvement >= 30 percentage points (pilot - baseline).

---

### Milestone M6 — Documentation and Cross-References (Priority: Medium)

**Goal**: Surface the new behavior in user-facing rule documents.

**Tasks**:
1. Update `.claude/rules/moai/core/zone-registry.md` with new HARD entries:
   - REQ-VTE-010 (>5% failure overall FAIL)
   - REQ-VTE-013 (no summary-away of failures)
   - REQ-VTE-016 (default profile must-pass)
2. Update CLAUDE.md §7 Language-Specific Guidelines to cross-reference the new matrix file (no content duplication; reference only).
3. Update `.claude/rules/moai/quality/boundary-verification.md` (if it exists) with a new section "Test Execution Boundary."

**Exit criteria**: Rule files cross-referenced; no duplicated content; zone-registry has new entries.

---

## Technical Approach

### Decision: Single matrix file vs per-language rule files

We choose a single matrix file (`verification-test-matrix.md`) over per-language rule files because:
- The matrix is structured tabular data; markdown table is appropriate.
- Per-language rule files (e.g., `.claude/rules/moai/language/go.md`) would explode count and dilute existing language rules.
- Single-file edits are easier to audit (M2 dry-run).

### Decision: 10-minute timeout matches runtime ceiling

The Bash tool's runtime maximum is 600,000 ms (CLAUDE.local.md, agent-authoring.md). We set the SPEC timeout exactly at this maximum to extract maximum signal. Tests that genuinely exceed 10 minutes route to a blocker report — which is the correct semantic, since the user (not the agent) needs to decide what subset to evaluate.

### Decision: Per-language failure ratio + aggregate

Per-language ratio is the primary metric (matches developer mental model: "JavaScript tests are broken"). Aggregate ratio is reported alongside but does not gate the verdict. This avoids the multi-language masking failure mode (10000 Go tests passing dilute 5/10 JavaScript failures to 0.05% aggregate).

### Decision: Output parser approach is per-language regex

Each language's test runner has a deterministic summary line:
- Go: `PASS\nok ...` or `FAIL ... [N tests, M failures]`
- Python (pytest): `==== N passed, M failed in T seconds ====`
- Rust: `test result: ok. N passed; M failed; ...`
- TypeScript (Jest): `Tests: N passed, M failed, total ...`

A regex per language captures these. For languages with non-machine-readable output (older toolchains), the matrix documents the heuristic. This is fragile but acceptable — false ratios are detectable downstream by manual audit if needed.

### Code-Level Architecture Notes (informational, deferred to Run phase)

- Whether the matrix is parsed by Go code (e.g., `internal/quality/test_matrix.go`) or interpreted directly by the agent body via Bash heuristics is a Run-phase decision. spec.md REQ-VTE-002 mandates the matrix file location; the consumer of the matrix is open.
- The fixture corpus (M5) may live under `internal/template/templates/.moai/reports/.fixtures/` or under `testdata/`; placement deferred to Run phase per existing test-data conventions.
- The dry-run automation (M2) may be a Bash script or a Go test; deferred.

---

## Risks

| # | Risk | Severity | Likelihood | Mitigation | Tracked in |
|---|------|----------|------------|------------|-----------|
| R1 | False positives on flaky tests increase noise | Medium | High | 5% tolerance gives a buffer; flaky_test_warning lets users see the flake | M5 fixture: flaky-go |
| R2 | Test execution increases evaluator-active wall-clock time substantially | Medium | High | 10-minute ceiling; user-visible warning when test run is the dominant time | M3 body: timeout doc |
| R3 | Language detection misclassifies polyglot project (false language reported) | Medium | Medium | Multi-language detection (REQ-VTE-007) runs all matchers; aggregation rule prevents single false detection from dominating | M5 fixture: multi-lang |
| R4 | Test command requires write permissions; evaluator-active blocks under permissionMode: plan | High | Medium | REQ-VTE-018 emits explicit warning; user can re-run evaluator-active with elevated permission OR fix the test to be read-only | M5 fixture: write-side-effect |
| R5 | New must-pass criterion breaks projects whose tests were never green | High | Low | Lenient profile downgrade (REQ-VTE-017); user can opt out at profile level | M4 |
| R6 | Output parser regression when toolchain output format changes (e.g., go test new format) | Medium | Medium | Add explicit toolchain version range to matrix entry; document upgrade procedure | M1 |
| R7 | AC-12 differential measurement is hard to perform repeatably | Medium | High | Pilot uses 10 fixed historical commits; result is one-shot, not a recurring metric. Acknowledged in EX-8 | M5 |
| R8 | Existing CI workflows that invoke evaluator-active break under longer wall-clock | Medium | Medium | M5 includes CI integration check; if breaking, file follow-up SPEC for CI tuning | M5, M6 |

---

## Approval Points

- **AP-1**: After M0 — user approval on OQ-1 through OQ-6 resolutions.
- **AP-2**: After M2 — user review of dry-run availability matrix; confirm pilot languages.
- **AP-3**: After M3 — user review of evaluator-active body diff before mirroring to template.
- **AP-4**: After M5 — user review of differential measurement results.

---

## Handover to Run Phase

When this SPEC moves to /moai run:

- **Manager**: manager-tdd (test fixtures and differential measurement are test-shaped work).
- **Expert delegation**: expert-testing for fixture authoring; expert-backend if Go-side parser/audit code is added.
- **Skills to load**: `moai-foundation-core`, `moai-foundation-quality`, `moai-workflow-spec`.
- **Quality gate**: Standard harness.
- **Critical path file order**:
  1. `.moai/reports/verify-test-oq-{DATE}/decisions.md` (M0)
  2. `internal/template/templates/.claude/rules/moai/quality/verification-test-matrix.md` (M1)
  3. `internal/template/verification_matrix_audit_test.go` (M1 audit)
  4. `internal/template/templates/.claude/agents/moai/evaluator-active.md` (M3)
  5. `internal/template/templates/.moai/config/evaluator-profiles/*.md` (M4)
  6. Fixture corpus under `.moai/reports/verify-test-fixtures-{DATE}/` (M5)
  7. `internal/template/templates/.claude/rules/moai/core/zone-registry.md` (M6)

---

**Status**: draft — review with plan-auditor and stakeholder before promotion.
