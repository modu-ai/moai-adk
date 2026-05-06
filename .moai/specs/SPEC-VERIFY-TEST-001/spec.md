---
id: SPEC-VERIFY-TEST-001
version: "0.1.0"
status: draft
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
priority: High
labels: [evaluator-active, verification, anti-rubber-stamping, test-execution, multi-language]
issue_number: null
wave: 1
tier: 0
scope:
  - .claude/agents/moai/evaluator-active.md
  - .moai/config/evaluator-profiles/default.md
  - .moai/config/evaluator-profiles/strict.md
  - .moai/config/evaluator-profiles/lenient.md
  - .moai/config/evaluator-profiles/frontend.md
  - .claude/rules/moai/quality/verification-test-matrix.md
blockedBy: []
dependents: []
---

# SPEC-VERIFY-TEST-001: Mandatory Test Execution for evaluator-active

## HISTORY

- 2026-04-30: Initial draft. Wave 1 — Tier 0. Closes the rubber-stamping gap in evaluator-active by mandating test suite execution before scoring. Source: Anthropic "Building Multi-Agent Systems" — concrete-criteria principle.

---

## 1. Overview

evaluator-active is MoAI-ADK's independent skeptical quality evaluator. Today, its body declares "do not award PASS without concrete evidence" but does not enumerate what evidence to gather. The Functionality dimension (40% weight) can be scored by reading the implementation and reasoning about acceptance criteria — without executing any test. Anthropic's empirical lesson (verbatim, see research.md §1.1) is that this configuration produces "rubber-stamping": rationalized passes that confirm the model's own conclusions rather than ground-truth correctness.

This SPEC operationalizes Anthropic's prescription. evaluator-active body is amended with a "Mandatory Test Execution" section. The default evaluator profile gains a `test_suite_passing` must-pass criterion. A 16-language test command matrix lives in `.claude/rules/moai/quality/verification-test-matrix.md`. evaluator-active runs the appropriate test command before scoring the Functionality dimension and reports passed/failed/skipped counts as concrete evidence.

---

## 2. Problem Statement

### 2.1 Current State

evaluator-active body (`.claude/agents/moai/evaluator-active.md:36-44`) declares HARD rules that include "Do NOT award PASS without concrete evidence (test output, verified behavior, specific file:line references)." However:

- No procedural step requires the agent to **execute** tests.
- No must-pass criterion in the default profile requires test suite passing.
- The {evidence} field in the output template is free-text; the agent's own discretion decides what qualifies.

### 2.2 Failure Mode (Rubber-Stamping)

Without enforcement:
1. Agent reads SPEC acceptance criteria.
2. Agent reads implementation.
3. Agent reasons: "the implementation appears to satisfy the criteria."
4. Agent awards Functionality PASS with evidence string like "implementation reviewed, logic appears sound."
5. No test was executed; broken changes pass through.

This is the precise failure Anthropic's blog post warns against. "Vague criteria let the model select for confirmation; concrete criteria force the model to confront reality."

### 2.3 Why now

Wave 1 — Tier 0. Highest immediate value: every evaluator-active invocation today is at risk of rubber-stamping. The cost of fixing this is small (one matrix file, one body section, four profile updates). The downstream impact — every SPEC closure relies on evaluator-active's verdict — is large. No upstream dependency on the advisor strategy SPEC (they are orthogonal).

### 2.4 Why this SPEC and not "just be more rigorous"

The current body already exhorts skepticism. Adding stronger language reproduces exhortation; we need procedural change. This SPEC introduces a procedural step (run the test command), an output contract (passed/failed/skipped counts), and a hard threshold (>5% failure = overall FAIL).

---

## 3. Requirements (EARS)

### 3.1 Ubiquitous

- **REQ-VTE-001** [Ubiquitous] THE EVALUATOR-ACTIVE AGENT BODY SHALL include a "Mandatory Test Execution" section enumerating: detection algorithm, test command per detected language, evidence format, and timeout policy.

- **REQ-VTE-002** [Ubiquitous] THE PROJECT SHALL maintain a test-command matrix at `.claude/rules/moai/quality/verification-test-matrix.md` covering all 16 supported languages (go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift) with: project marker, test command, pass exit code, optional coverage flag.

### 3.2 Event-Driven (Test Execution)

- **REQ-VTE-003** [Event-Driven] WHEN evaluator-active begins scoring the Functionality dimension, THE EVALUATOR SHALL detect the project's primary language(s) using project-marker presence (e.g., `go.mod` for go, `pyproject.toml`/`requirements.txt` for python).

- **REQ-VTE-004** [Event-Driven] WHEN one or more languages are detected, THE EVALUATOR SHALL execute the corresponding test command(s) from the matrix using the Bash tool (already declared in evaluator-active's `tools` field) with a 10-minute (600,000 ms) timeout per command.

- **REQ-VTE-005** [Event-Driven] WHEN the test command completes, THE EVALUATOR SHALL parse the output to extract: total tests, passed count, failed count, skipped count, and the file:line of each failure (when language toolchain provides this).

- **REQ-VTE-006** [Event-Driven] WHEN the test execution exceeds the 10-minute timeout, THE EVALUATOR SHALL emit a structured blocker report (per agent-common-protocol.md "Blocker Report Format") and return overall verdict UNVERIFIED for the Functionality dimension. The agent SHALL NOT auto-retry with a longer timeout.

### 3.3 State-Driven / Where (Multi-Language Projects)

- **REQ-VTE-007** [Where] WHERE multiple language markers are detected (e.g., a project with both `go.mod` and `package.json`), THE EVALUATOR SHALL execute test commands for each detected language sequentially AND aggregate results: total tests = sum across languages; failed = sum across languages.

- **REQ-VTE-008** [Where] WHERE no language marker is detected, THE EVALUATOR SHALL log `language_detected: none` AND apply the no-tests penalty (REQ-VTE-014) without attempting test execution.

### 3.4 State-Driven (Failure Thresholds)

- **REQ-VTE-009** [State-Driven] WHILE evaluating the Functionality dimension, IF any test fails (failed_count > 0), THEN THE FUNCTIONALITY DIMENSION SCORE SHALL be capped at 0.50 (regardless of other observations on that dimension).

- **REQ-VTE-010** [State-Driven] WHILE evaluating the Functionality dimension, IF the failure ratio (`failed_count / total_count`) exceeds 0.05 (5%), THEN THE OVERALL EVALUATION VERDICT SHALL be FAIL regardless of the weighted dimension scores.

- **REQ-VTE-011** [State-Driven] WHILE the failure ratio is between 0.00 (exclusive) and 0.05 (inclusive), THE EVALUATOR SHALL emit a `flaky_test_warning` in the report AND THE FUNCTIONALITY DIMENSION SHALL NOT exceed 0.50, but overall verdict is determined by aggregate scoring.

### 3.5 Evidence Reporting Contract

- **REQ-VTE-012** [Ubiquitous] THE FUNCTIONALITY DIMENSION'S {evidence} FIELD IN THE EVALUATION REPORT SHALL contain, at minimum: the test command(s) executed, exact passed/failed/skipped counts, language(s) detected, and per-failure file:line references for the first 10 failures. Counts SHALL NOT be approximated or rounded.

- **REQ-VTE-013** [Unwanted] THE EVALUATOR SHALL NOT summarize away test failures (e.g., "minor test issues" without specific counts and locations) AND SHALL NOT report a Functionality PASS without a corresponding test execution evidence block.

### 3.6 No-Tests-Present Penalty

- **REQ-VTE-014** [Where] WHERE the project has a recognizable language marker but no test directory or test files are present (heuristic: `find . -name '*_test.*' -o -name 'test_*.*' -o -name '*.test.*' | head` returns empty), THE EVALUATOR SHALL apply a Functionality dimension penalty of -0.20 AND log `no_tests_present: true` in the report.

### 3.7 Tool Availability and Graceful Skip

- **REQ-VTE-015** [Where] WHERE the test command for a detected language is not available on PATH (e.g., `go` binary missing in CI environment), THE EVALUATOR SHALL log `tool_skipped: <command>` AND mark the Functionality dimension as UNVERIFIED for that language. The evaluator SHALL NOT silently skip.

### 3.8 Profile Compatibility

- **REQ-VTE-016** [Ubiquitous] THE DEFAULT EVALUATOR PROFILE (`.moai/config/evaluator-profiles/default.md`) SHALL list `test_suite_passing` as a must-pass criterion alongside the existing Functionality / Security must-pass entries.

- **REQ-VTE-017** [Where] WHERE a project uses the `lenient` evaluator profile, the must-pass criterion may be downgraded to `test_suite_warning` (failure does not auto-FAIL but is reported). The `strict` profile SHALL retain the must-pass enforcement. The `frontend` profile SHALL inherit the default behavior unless explicitly overridden.

### 3.9 Permission Mode Interaction

- **REQ-VTE-018** [State-Driven] WHILE evaluator-active runs in its declared `permissionMode: plan` (read-only), IF the detected test command requires write permissions (e.g., creates test fixtures, runs a database), THEN THE EVALUATOR SHALL emit a `permission_required: write` warning in the report AND fall through to UNVERIFIED for the Functionality dimension. The evaluator SHALL NOT silently elevate its permission mode.

---

## 4. Acceptance Criteria

(Detailed Given-When-Then scenarios live in `acceptance.md`. This is the summary list.)

- **AC-1**: evaluator-active body contains the "Mandatory Test Execution" section with detection algorithm and reference to the matrix file.
- **AC-2**: `.claude/rules/moai/quality/verification-test-matrix.md` exists, lists all 16 languages, and is dry-run validated (each command parses and runs `--help` or equivalent on a controlled environment).
- **AC-3**: For a Go project with passing tests (the moai-adk-go repository itself), evaluator-active executes `go test ./...`, reports correct passed/failed counts, and the Functionality dimension scores reflect test outcome.
- **AC-4**: For a Go project with one synthetically failing test (failure ratio < 5%), evaluator-active reports the failure, caps Functionality at 0.50, but overall verdict is determined by aggregate; the report contains `flaky_test_warning`.
- **AC-5**: For a Go project with > 5% failure ratio, evaluator-active reports overall verdict FAIL regardless of other dimensions, with explicit `failure_ratio_exceeded` marker.
- **AC-6**: For a multi-language project (e.g., a directory with both `go.mod` and `package.json`), evaluator-active runs both `go test ./...` and `npm test`, aggregates results, and reports per-language breakdown.
- **AC-7**: For a project with no test files, evaluator-active applies the -0.20 penalty and logs `no_tests_present: true`. Verdict UNVERIFIED unless other dimensions compensate.
- **AC-8**: For a project where the test command is missing from PATH, evaluator-active logs `tool_skipped: <command>` and marks Functionality UNVERIFIED. The evaluator does not abort the entire evaluation.
- **AC-9**: For a test suite that exceeds the 10-minute timeout, evaluator-active emits a blocker report (UNVERIFIED) and does not retry.
- **AC-10**: All four evaluator profiles (default, strict, lenient, frontend) are reviewed and updated for compatibility with the new must-pass criterion.
- **AC-11**: The 16-language matrix is exercised by an automated dry-run test (in CI or in the verification step) that confirms each test command exists or is documented as optional.
- **AC-12**: Differential measurement: on a curated test set of 10 known-broken changes, the post-SPEC evaluator-active correctly reports FAIL on at least 8 (rate >= 80%); the pre-SPEC baseline is captured for comparison. Acceptance: improvement of at least 30 percentage points over baseline (e.g., baseline 40% → pilot 70%+).
- **AC-13**: For a project where the detected test command requires write permissions while evaluator-active runs in `permissionMode: plan` (read-only), evaluator-active emits a `permission_required: write` warning in the report, marks the Functionality dimension as UNVERIFIED, and SHALL NOT silently elevate its permission mode. Verifies REQ-VTE-018.

---

## 5. REQ-ID Matrix

| REQ-ID | Type | Priority | Verification | Acceptance Criterion |
|--------|------|----------|--------------|----------------------|
| REQ-VTE-001 | Ubiquitous | High | Body inspection | AC-1 |
| REQ-VTE-002 | Ubiquitous | High | File presence + content audit | AC-2, AC-11 |
| REQ-VTE-003 | Event-Driven | High | Trace inspection (language detection log) | AC-3, AC-6 |
| REQ-VTE-004 | Event-Driven | Critical | Trace inspection (Bash execution) | AC-3, AC-4 |
| REQ-VTE-005 | Event-Driven | High | Output parsing test | AC-3, AC-4, AC-5 |
| REQ-VTE-006 | Event-Driven | High | Timeout simulation test | AC-9 |
| REQ-VTE-007 | Where | Medium | Multi-language fixture test | AC-6 |
| REQ-VTE-008 | Where | Medium | No-marker fixture test | AC-7 |
| REQ-VTE-009 | State-Driven | Critical | Failing-test fixture test | AC-4 |
| REQ-VTE-010 | State-Driven | Critical | High-failure-ratio fixture test | AC-5 |
| REQ-VTE-011 | State-Driven | High | Low-failure-ratio fixture test | AC-4 |
| REQ-VTE-012 | Ubiquitous | High | Output schema audit | AC-3, AC-4, AC-5 |
| REQ-VTE-013 | Unwanted | Critical | Output scan for forbidden vagueness | AC-12 |
| REQ-VTE-014 | Where | Medium | No-test fixture test | AC-7 |
| REQ-VTE-015 | Where | High | Missing-tool fixture test | AC-8 |
| REQ-VTE-016 | Ubiquitous | Critical | Profile content audit | AC-10 |
| REQ-VTE-017 | Where | High | Profile content audit (per-profile) | AC-10 |
| REQ-VTE-018 | State-Driven | High | Permission warning trace | AC-13 |

**Total**: 18 requirements (4 Ubiquitous, 4 Event-Driven, 5 Where, 4 State-Driven, 1 Unwanted, 0 Optional). Distribution covers Ubiquitous, Event-Driven, State-Driven, Where, Unwanted (Optional N/A for this scope).

---

## 6. Out of Scope (Exclusions — What NOT to Build)

- **EX-1**: Coverage measurement is **not** mandated by this SPEC. The Craft dimension already references "test coverage >= 85%"; integrating a coverage tool runner into evaluator-active is a follow-up SPEC. This SPEC focuses on test PASS/FAIL counts, not coverage.
- **EX-2**: Rewriting tests, generating tests, or fixing failing tests is out of scope. evaluator-active is read-only (`permissionMode: plan`); it observes test outcomes and does not modify code.
- **EX-3**: Linter and static-analysis tool execution is out of scope. manager-quality already runs lint commands per CLAUDE.md §7. evaluator-active focuses on test execution specifically.
- **EX-4**: A new evaluator profile (e.g., `paranoid` with stricter thresholds) is out of scope. The four existing profiles are updated for compatibility, not extended.
- **EX-5**: Cross-session test result caching is out of scope. Each evaluator-active invocation runs tests fresh. Optimization deferred.
- **EX-6**: Test parallelism configuration is out of scope. evaluator-active uses each language's default test execution (e.g., `go test ./...` runs in parallel by default; `pytest` may need `-n auto` to parallelize). Tuning per project deferred.
- **EX-7**: Modifying manager-quality's behavior is out of scope. manager-quality continues to run its own quality gate independently. Cross-agent deduplication is not pursued in this SPEC.
- **EX-8**: Metrics dashboard / aggregated false-PASS rate dashboard is out of scope. AC-12 measurement is one-shot during pilot; continuous tracking deferred.

---

## 7. Open Questions

- **OQ-1**: Should `test_suite_passing` be a Functionality must-pass criterion (intra-dimension) or an evaluator-level must-pass criterion (overall)? REQ-VTE-010 currently makes >5% failure an overall FAIL hard threshold (matching Security FAIL), but REQ-VTE-016 lists it under default profile must-pass criteria as if it were a dimension constraint. Need to align the single specification location in plan phase.

- **OQ-2**: For multi-language projects, is the failure ratio computed per-language or aggregated? E.g., a project with `go test`: 100/100 pass and `npm test`: 5/10 pass — is failure ratio 5/110 = 4.5% (below threshold) or per-language 50% on JS (above threshold)? Resolution requires user input. Default proposal: per-language, with overall ratio as a secondary signal.

- **OQ-3**: How should evaluator-active handle test commands that produce non-machine-readable output (cpp `ctest` summary, scala `sbt test` raw text)? Specific parsers per language vs heuristic line-grep? Defer to implementation phase; document parser approach in the matrix file.

- **OQ-4**: For the AC-12 differential measurement, what counts as a "known-broken change"? Need a curated test fixture: 10 small commits where the change provably breaks at least one test. Curation effort deferred to plan phase.

- **OQ-5**: Should the no-tests penalty (-0.20, REQ-VTE-014) scale with project size (small library = harsher penalty; large monorepo with one tested module = lighter)? Default: flat -0.20. Revisit if pilot data suggests calibration.

- **OQ-6**: When the language detection finds Python via `requirements.txt` but no `pyproject.toml` exists, which test command takes precedence — `pytest`, `python -m unittest`, or `python -m pytest`? Decision deferred to matrix authoring; will probe each in priority order.

---

**Total lines**: ~245
**Status**: draft — awaiting plan-auditor review
