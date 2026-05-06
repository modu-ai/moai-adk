# Research: SPEC-VERIFY-TEST-001 (Verification Test Suite Enforcement)

> **SPEC**: SPEC-VERIFY-TEST-001
> **Wave**: 1 — Tier 0
> **Author**: manager-spec
> **Date**: 2026-04-30

---

## 1. Source Documents

### 1.1 Primary Source — Anthropic "Building Multi-Agent Systems"

**URL**: https://anthropic.com/engineering/built-multi-agent-research-system

**Verbatim quotations** (foundational claims that motivate this SPEC):

> "Use concrete criteria like 'Run the full test suite and report all failures' rather than vague requirements."

> "You MUST run the complete test suite before marking as passed. Only mark as PASSED if ALL tests pass with no failures."

**Key insight**: Anthropic's empirical lesson is that LLM evaluators trend toward **rationalized passes** ("rubber-stamping") when the verification criterion is implicit or aspirational. The mitigation is not "ask the model to be more rigorous" but to make verification **mechanical**: run the actual test command, parse the actual output, and let the result determine the verdict. Vague criteria let the model select for confirmation; concrete criteria force the model to confront reality.

### 1.2 Supporting Source — Anthropic "Best Practices for Opus 4.7"

> "Opus 4.7 prefers reasoning over tool invocation. When tool use is expected, specify when and why to use each tool."

**Inference**: Without an explicit "MUST run the test suite" directive in the agent body, Opus 4.7 may rationalize that test-running is unnecessary because "the implementation logic is sound." This is exactly the failure mode the verbatim quotation in §1.1 warns against.

### 1.3 Supporting Source — Karpathy Coding Principles (NOTICE.md)

`.claude/rules/moai/NOTICE.md` documents an absorbed anti-pattern: "Claiming Without Evidence."

> "Every task requires evidence of completion."
> ".claude/skills/moai/references/anti-patterns.md" (cf. CLAUDE.md §6 Verify, Don't Assume)

**Inference**: This SPEC operationalizes that already-absorbed principle for one specific agent (evaluator-active) on one specific evidence type (test execution output).

---

## 2. Codebase Analysis

### 2.1 Current evaluator-active Configuration

`.claude/agents/moai/evaluator-active.md`:

| Field | Value | File:Line |
|-------|-------|-----------|
| name | evaluator-active | :2 |
| model | sonnet | :13 |
| effort | high | :14 |
| permissionMode | plan | :15 |
| tools | Read, Grep, Glob, **Bash**, mcp__sequential-thinking__sequentialthinking | :12 |

**Critical finding**: evaluator-active already has `Bash` tool access (line 12). It CAN run tests today; it is not configured to actually do so.

### 2.2 Skeptical Evaluation Mandate (lines 36-44)

The agent body declares hard rules including:

> "Do NOT award PASS without concrete evidence (test output, verified behavior, specific file:line references)."

> "If you cannot verify a criterion, mark it as UNVERIFIED, not PASS."

**Problem**: "Concrete evidence" is mentioned but **what counts as evidence is not enumerated**. The agent body never explicitly says "run `go test ./...` (or language equivalent) before scoring the Functionality dimension." Without this, the agent's own discretion decides what evidence to gather, producing the rationalization gap §1.1 warns about.

### 2.3 Evaluation Dimensions (lines 46-55)

```
| Functionality | 40% | All SPEC acceptance criteria met | Any criterion FAIL |
| Security | 25% | OWASP Top 10 compliance | Any Critical/High finding |
| Craft | 20% | Test coverage >= 85%, error handling | Coverage below threshold |
| Consistency | 15% | Codebase pattern adherence | Major pattern violations |
```

The Craft dimension says "Test coverage >= 85%" but neither "running tests" nor "measuring coverage" appears as a procedural step. The dimension reads as a target, not a procedure.

### 2.4 Default Evaluator Profile

`.moai/config/evaluator-profiles/default.md` (lines 14-18):

```
## Must-Pass Criteria
- Functionality: All SPEC acceptance criteria must be met (no partial credit)
- Security: No Critical or High severity findings (FAIL overrides overall score)
```

**Gap**: `test_suite_passing` is NOT a must-pass criterion. An agent could PASS Functionality by reading the implementation and reasoning that "the acceptance criteria appear to be addressed by this code" without ever running a test. This is the rubber-stamping risk in measurable form.

### 2.5 Existing Language Toolchain Detection in moai-adk-go

CLAUDE.md §7 (Language-Specific Guidelines):

> "The quality gate auto-detects the project language and runs the appropriate toolchain:
> - **Go**: `go vet` → `golangci-lint` → `go test`
> - **Node.js**: `eslint` → `npm test`
> - **Python**: `ruff` → `pytest`
> - **Rust**: `cargo clippy` → `cargo test`
> Tools that are not installed are skipped gracefully. Projects with no recognized language marker pass the gate silently."

**Reference precedent**: The existing `manager-quality` quality gate does exactly the kind of toolchain detection this SPEC needs. We can borrow its detection table and extend coverage to all 16 supported languages.

CLAUDE.local.md §15 lists the 16 supported languages:

```
go, python, typescript, javascript, rust, java, kotlin, csharp,
ruby, php, elixir, cpp, scala, r, flutter, swift
```

### 2.6 Test Command Matrix (16 Languages)

Survey of standard test invocations per language:

| Language | Project marker | Test command | Pass exit | Coverage flag |
|----------|---------------|--------------|-----------|---------------|
| go | go.mod | `go test ./...` | 0 | `-cover -coverprofile=...` |
| python | pyproject.toml, requirements.txt | `pytest` (or `python -m pytest`) | 0 | `--cov=` |
| typescript | package.json + tsconfig.json | `npm test` (or `pnpm test`, `yarn test`) | 0 | `--coverage` |
| javascript | package.json | `npm test` | 0 | `--coverage` |
| rust | Cargo.toml | `cargo test` | 0 | `cargo tarpaulin` (separate) |
| java | pom.xml or build.gradle | `mvn test` / `gradle test` | 0 | jacoco plugin |
| kotlin | build.gradle.kts | `gradle test` | 0 | jacoco plugin |
| csharp | *.csproj | `dotnet test` | 0 | `--collect:"XPlat Code Coverage"` |
| ruby | Gemfile | `bundle exec rake test` or `rspec` | 0 | simplecov gem |
| php | composer.json | `vendor/bin/phpunit` | 0 | `--coverage-html` |
| elixir | mix.exs | `mix test` | 0 | `--cover` |
| cpp | CMakeLists.txt or Makefile | `ctest` (or `make test`) | 0 | gcov / lcov (separate) |
| scala | build.sbt | `sbt test` | 0 | scoverage plugin |
| r | DESCRIPTION + tests/ | `Rscript -e 'devtools::test()'` | 0 | covr package |
| flutter | pubspec.yaml + flutter dependency | `flutter test` | 0 | `--coverage` |
| swift | Package.swift | `swift test` | 0 | `--enable-code-coverage` |

**Verification needed in plan phase**: each command exists on PATH OR project README documents an alternative. Skip-with-warning behavior aligns with CLAUDE.md §7 ("Tools that are not installed are skipped gracefully").

### 2.7 Failure Reporting Granularity

evaluator-active output format (.claude/agents/moai/evaluator-active.md:60-77):

```
| Functionality (40%) | {n}/100 | PASS/FAIL/UNVERIFIED | {evidence} |
```

The {evidence} field is currently free-text. We need a stricter contract for the Functionality dimension: when a test suite was run, the evidence field MUST include passed/failed/skipped counts and at least one file:line reference per failure (no aggregation away).

### 2.8 Hard Threshold Mechanism Already Exists

`evaluator-active.md:55`: "HARD THRESHOLD: Security dimension FAIL = Overall FAIL (regardless of other scores)."

This pattern (a single criterion that overrides aggregate) is already in place. Adding "test suite > 5% failure rate = Overall FAIL" follows the same pattern; no architectural change needed.

---

## 3. Alternative Approaches

### 3.1 Alternative A — Extend manager-quality only (no evaluator-active change)

manager-quality already runs the quality gate per CLAUDE.md §7. Push test execution there exclusively; evaluator-active does not run tests, only reads manager-quality output.

- **Pros**: No new test execution path; single ownership of toolchain detection; no risk of evaluator-active hitting test infrastructure (CI agents).
- **Cons**: evaluator-active and manager-quality run in different phases (run-end vs separate). evaluator-active may report PASS based on stale manager-quality data. The "rubber-stamping" failure mode this SPEC targets is *specifically* in evaluator-active (skeptical evaluator dimension), not manager-quality (mechanical lint+test runner). Delegating away misses the point.
- **Verdict**: Rejected. The SPEC's whole motivation is closing the evaluator's evidence gap.

### 3.2 Alternative B — evaluator-active runs tests inline (preferred)

evaluator-active body is amended with a "Mandatory Test Execution" section. Bash tool is already declared. Test command matrix is referenced from a shared rule.

- **Pros**: Closes the rubber-stamping gap directly; evaluator owns its evidence chain; no cross-agent dependency.
- **Cons**: evaluator-active execution time increases (test suite runs are slow on large projects). Mitigation: 10-minute timeout (REQ in spec.md); long-running suites trigger blocker report.
- **Verdict**: Preferred.

### 3.3 Alternative C — Hybrid: manager-quality runs tests, evaluator-active *verifies the manager-quality run was recent*

Add a freshness check: evaluator-active reads manager-quality's last execution timestamp; if > 5 minutes old, evaluator runs tests itself; otherwise reuses manager-quality's output.

- **Pros**: Avoids redundant test runs.
- **Cons**: Adds session-state coupling that doesn't exist today; freshness threshold is arbitrary; manager-quality's output format is not currently stable enough for cross-agent reuse without a contract SPEC.
- **Verdict**: Defer to follow-up. Complexity-to-value ratio is poor for Wave 1.

### 3.4 Alternative D — Use harness profile to gate enforcement

Make test execution mandatory only when harness level is `thorough`; standard/minimal evaluations skip it.

- **Pros**: Fast feedback for trivial changes.
- **Cons**: The rubber-stamping risk is highest at standard level (most common evaluations), not thorough. Gating by harness reverses the protection where it's needed most.
- **Verdict**: Rejected at gating axis. May still tune *timeout* by harness level (thorough gets more time).

### 3.5 Decision Matrix

| Criterion (weight) | A (manager-quality) | B (evaluator inline) | C (hybrid freshness) | D (harness-gated) |
|--------------------|--------------------|----------------------|-----------------------|--------------------|
| Closes rubber-stamping gap (40%) | 4/10 | 9/10 | 7/10 | 5/10 |
| Implementation cost (20%) | 6/10 | 8/10 | 4/10 | 7/10 |
| Maintenance burden (15%) | 8/10 | 7/10 | 4/10 | 6/10 |
| Consistency with existing patterns (15%) | 7/10 | 8/10 | 5/10 | 6/10 |
| Performance impact (10%) | 9/10 | 6/10 | 7/10 | 8/10 |
| **Weighted total** | **6.05** | **8.00** | **5.65** | **6.05** |

Path B wins clearly.

---

## 4. Decision Rationale

### 4.1 Why concretization, not exhortation

A common but failing approach is to add stronger language to the agent body: "REALLY make sure to run tests" / "Do NOT skip test execution." This is the same exhortation already present in lines 36-44 of evaluator-active.md (the "Skeptical Evaluation Mandate"). The Anthropic blog's prescription is opposite: replace exhortation with mechanical procedure. We add procedural steps with concrete commands, not stronger adjectives.

### 4.2 Why per-language matrix vs single command

A single command (`make test` or similar) does not exist in moai-adk-go's 16-language template universe. Each language has idiomatic test commands. The matrix is finite (16 entries), can be tabulated once, and matches existing CLAUDE.md §7 precedent.

### 4.3 Why 5% failure cap on overall PASS

The threshold is calibrated against typical engineering tolerance:
- 0% failure = "all green" — the gold standard.
- 1-5% failure = flaky test territory — common in mature codebases; doesn't necessarily indicate the new change broke something. Allows PASS at the dimension level (capped at 0.50) but does not force overall FAIL.
- > 5% failure = something is broken. Force overall FAIL regardless of other dimensions.

This mirrors the Security FAIL hard threshold already in place: a single quality dimension can override aggregate.

### 4.4 Why -0.20 penalty for "no_tests_present"

If a project ships without tests, evaluator-active cannot rubber-stamp and cannot verify either. The penalty discourages no-test projects from achieving high Functionality scores while not making them impossible to evaluate (penalty caps the dimension, not the project). Magnitude (-0.20) is calibrated so that a no-tests project can still PASS Functionality if all other evidence is exceptionally strong (1.00 - 0.20 = 0.80, still above 0.75 threshold for "primary criteria pass").

### 4.5 Why 10-minute timeout

Aligns with the Bash tool's hard ceiling (CLAUDE.local.md §coding-standards.md and agent-authoring.md: "Maximum: 600,000ms (10 minutes)"). Above this, the runtime kills the process. Setting evaluator-active's test timeout at the runtime maximum extracts the most signal possible without exceeding platform constraints. Tests longer than 10 minutes route to a blocker report, requiring human triage of which subset to run.

### 4.6 Why detection is mandatory and 100% accurate is the bar

A wrong language detection (e.g., flagging a Python project as Ruby) means running a wrong command, getting 0 results, and false-passing. Detection accuracy must be 100% on the 16 official markers. Detection is structural (file presence checks); 100% is achievable.

---

## 5. Risks & Mitigations

| # | Risk | Severity | Likelihood | Mitigation |
|---|------|----------|------------|------------|
| R1 | Test suite is genuinely flaky; PASS becomes inconsistent | Medium | High | 1-5% failure tolerance allows PASS-with-warning; >5% forces FAIL |
| R2 | Test command not on PATH (CI environment) | Medium | High | Detection step verifies command presence; if absent, dimension marked UNVERIFIED, not FAIL |
| R3 | Test suite takes > 10 minutes | High | Medium | Timeout triggers blocker report; user can subset tests; do not auto-skip |
| R4 | Multi-language project (Go + TypeScript) | Medium | Medium | Detection runs all matchers; runs tests for each detected language; aggregates failures |
| R5 | Tests have side effects (DB writes, API calls) | High | Medium | evaluator-active runs in `permissionMode: plan` — read-only by default. Test execution requires elevation. Mitigation: agent emits warning when test run requires write permissions; user explicitly grants |
| R6 | False-PASS rate is hard to measure (we'd need a parallel ground-truth evaluator) | Medium | High | Acceptance metric in spec.md is differential: pre-pilot vs post-pilot self-reported PASS rate on a curated sample of known-broken changes. Imperfect but actionable |
| R7 | Output format change breaks downstream consumers (manager-docs, plan-auditor) | Medium | Low | Breaking change documented in HISTORY; downstream consumers audited in M1 |
| R8 | "Skip gracefully" semantics may mask missing tools | Medium | Medium | Agent emits explicit `tool_skipped: <name>` log entry; no silent skip |

---

## 6. References

### Anthropic Sources
- Building Multi-Agent Systems: https://anthropic.com/engineering/built-multi-agent-research-system
- Best Practices for Opus 4.7: https://platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-7

### Codebase References
- `.claude/agents/moai/evaluator-active.md:12,13,36-44,46-55,60-77` — primary modification target
- `.moai/config/evaluator-profiles/default.md:14-18` — must-pass criteria extension target
- `.moai/config/evaluator-profiles/strict.md` — secondary profile, must inherit new must-pass criterion
- `.moai/config/evaluator-profiles/lenient.md` — must allow opt-out OR inherit lite version
- `.moai/config/evaluator-profiles/frontend.md` — language-specific dimension extensions, must remain compatible
- `.claude/agents/moai/manager-quality.md` — toolchain-detection precedent
- `CLAUDE.md §7 Language-Specific Guidelines` — language matrix precedent
- `CLAUDE.local.md §15` — official 16-language list
- `.claude/rules/moai/NOTICE.md § Karpathy Coding Principles` — Verify, Don't Assume principle

### Related SPECs
- SPEC-CORE-BEHAV-001 — Six Agent Core Behaviors; Behavior 6 ("Verify, Don't Assume") is the upstream principle
- SPEC-AGENT-002 — Agent body minimization; new evaluator-active body must respect minimum-body discipline
- SPEC-ASKUSER-ENFORCE-001 — pattern reference for [HARD] enforcement specs

---

**Total lines**: ~220
**Status**: Ready for SPEC drafting
