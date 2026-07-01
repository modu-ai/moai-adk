---
description: "Sync Phase 0.5~0.7 — Quality Verification, Security Scan, MX Tag Validation, and Coverage Analysis with Test Generation."
user-invocable: false
metadata:
  parent: moai-workflow-sync
  phase: "Phase 0.5~0.7: Quality Verification and Coverage"
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->
<!-- Emits one line per Phase entry/exit to stderr in format: [trace] /moai sync Phase <N> <enter|exit> -->

### Phase 0.5: Quality Verification

Purpose: Detect project language and run language-specific diagnostics (tests, linter, type checker) in parallel, followed by code review.

#### Step 0.5.1: Language Detection

Check indicator files in priority order (first match wins):

- Python: pyproject.toml, setup.py, requirements.txt, .python-version, Pipfile
- TypeScript: tsconfig.json, package.json with typescript dependency
- JavaScript: package.json without tsconfig
- Go: go.mod, go.sum
- Rust: Cargo.toml, Cargo.lock
- Ruby: Gemfile, .ruby-version, Rakefile
- Java: pom.xml, build.gradle, build.gradle.kts
- PHP: composer.json, composer.lock
- Kotlin: build.gradle.kts with kotlin plugin
- Swift: Package.swift, .xcodeproj, .xcworkspace
- C#/.NET: .csproj, .sln, .fsproj
- C++: CMakeLists.txt, Makefile with C++ content
- Elixir: mix.exs
- R: DESCRIPTION (R package), .Rproj, renv.lock
- Flutter/Dart: pubspec.yaml
- Scala: build.sbt, build.sc
- Fallback: unknown (skip language-specific tools, proceed to code review)

#### Step 0.5.2: Execute Diagnostics in Parallel

Launch three background tasks simultaneously:

- Test Runner: Language-specific test command (pytest, npm test, go test, cargo test, etc.)
- Linter: Language-specific lint command (ruff, eslint, golangci-lint, clippy, etc.)
- Type Checker: Language-specific type check (mypy, tsc --noEmit, go vet, etc.)

Collect all results with timeouts (180s for tests, 120s for others). Handle partial failures gracefully.

#### Step 0.5.3: Handle Test Failures

If any tests fail, use AskUserQuestion:

- Continue: Proceed with sync despite failures
- Abort: Stop sync, fix tests first (exit to Phase 4 graceful exit)

#### Step 0.5.4: Deep Code Review with Auto-Fix

Agent: sync-auditor subagent (independent quality scoring per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 2)

Invoke regardless of project language. Execute multi-perspective code review beyond basic TRUST 5 validation:

Review Perspectives:
- Security: OWASP Top 10 compliance, injection risks, secrets exposure, dependency vulnerabilities
- Performance: Algorithmic complexity, query efficiency (N+1), memory patterns, concurrency safety
- Quality: TRUST 5 compliance, error handling completeness, naming conventions, code consistency
- UX: User flow integrity, error states, accessibility (WCAG/ARIA), breaking changes in public interfaces

Auto-Fix Behavior:
- If critical issues found: Delegate auto-fix to manager-develop or a per-spawn `Agent(general-purpose)` domain specialist (per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C)
- Re-run review after fix to verify resolution
- Maximum 3 auto-fix iterations for critical issues before escalating to user
- Warnings and suggestions are logged in report but do not block pipeline

Output:
- Review report with findings by severity (critical, warning, suggestion)
- @MX tag compliance status (integrated with Phase 0.6)
- Auto-fix log if corrections were applied

#### LSP Quality Gates

The sync phase enforces LSP-based quality gates as configured in quality.yaml:
- Zero errors required (lsp_quality_gates.sync.max_errors: 0)
- Maximum 10 warnings allowed (lsp_quality_gates.sync.max_warnings: 10)
- Clean LSP state required (lsp_quality_gates.sync.require_clean_lsp: true)

#### Step 0.5.5: Generate Quality Report

Aggregate all results into a quality report showing status for test-runner, linter, type-checker, and code-review. Determine overall status (PASS or WARN).

### Phase 0.55: Security Scan (Conditional)

Purpose: Run a targeted security audit on changed files before PR creation. Catches security vulnerabilities that code review alone may miss.

**Activation condition**: Execute this phase ONLY when changed files match security-sensitive patterns:
- Authentication/authorization files (auth, login, session, token, permission, role)
- Database interaction files (query, model, migration, schema, repository, dao)
- API endpoint files (handler, controller, route, endpoint, middleware)
- User input handling files (form, input, validation, sanitize)
- Configuration files with secrets (.env, config with credentials)

**Skip condition**: If no changed files match security-sensitive patterns, skip to Phase 0.6. Log: "Security scan skipped: no security-sensitive files changed."

#### Step 0.55.1: Security Analysis

Agent: per-spawn `Agent(general-purpose)` security reviewer (security whitelist per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 9; OR the Stop hook dependency-manifest audit `.claude/hooks/moai/sync-phase-quality-gate.sh`)

Delegate to a per-spawn `Agent(general-purpose)` security reviewer loading the retained `moai-ref-owasp-checklist` / `moai-ref-secops` skills (the documented security replacement path) in inline mode:
- Only CRITICAL findings block the sync pipeline
- HIGH findings are reported as warnings in PR description
- MEDIUM and LOW findings are logged in sync report

**Dependency manifest audit (always runs, regardless of whether manifest files changed in this SPEC)**:

Audit ALL of the following manifest files present at project root — dependency surface must be checked at every sync to detect drift from transitive vulnerability changes unrelated to this SPEC:
`go.mod`, `package.json`, `requirements.txt`, `Cargo.toml`, `pyproject.toml`, `Gemfile`, `composer.json`, `mix.exs`, `Package.swift`, `pubspec.yaml`.

When any manifest is detected, run a dependency vulnerability scan of that manifest via the Stop hook dependency-manifest audit (`.claude/hooks/moai/sync-phase-quality-gate.sh`) OR a per-spawn `Agent(general-purpose)` security reviewer.
Rationale: a transitive vulnerability may have been introduced by an unrelated dependency update since the last sync, even if no manifest file was modified in the current SPEC.

#### Step 0.55.2: Security Gate Decision

If CRITICAL findings exist:
- Present findings via AskUserQuestion:
  - Fix now (Recommended): Delegate to a per-spawn `Agent(general-purpose)` security reviewer for auto-fix, then re-scan
  - Continue with warning: Proceed to Phase 0.6 with security warnings embedded in PR description
  - Abort: Exit sync workflow

If no CRITICAL findings: Proceed to Phase 0.6. Include any HIGH/MEDIUM findings in the sync report.

### Phase 0.6: MX Tag Validation (Multi-Language)

Purpose: Ensure code has appropriate @MX annotations for AI agent context. Supports all 16 MoAI-ADK languages.

**[HARD] P1/P2 violations BLOCK sync.** If any P1 (missing @MX:ANCHOR on fan_in >= 3 function) or P2 (missing @MX:WARN on goroutine pattern) violations are found, sync is halted and the user must resolve them before proceeding.

- P1 (Blocking): exported function with fan_in >= 3 missing @MX:ANCHOR
- P2 (Blocking): goroutine/async pattern missing @MX:WARN
- P3 (Advisory): long exported function missing @MX:NOTE — warning only, sync continues
- P4 (Advisory): untested public function missing @MX:TODO — warning only, sync continues

When P1/P2 violations are detected:
1. Display full violation report with file:line references
2. Show message: "Run /moai run to add missing tags, or use --skip-mx to bypass"
3. Halt sync — do NOT proceed to Phase 0.7+

Skip if `--skip-mx` flag is provided. When skipped, log: "MX validation skipped by user flag" in sync report.

#### Step 0.6.1: Language Detection for Modified Files

Detect languages present in modified files:

| Language | Indicator Files | File Patterns | Comment Prefix |
|----------|----------------|---------------|----------------|
| Go | go.mod | *.go | `//` |
| Python | pyproject.toml | *.py | `#` |
| TypeScript | tsconfig.json | *.ts, *.tsx | `//` |
| JavaScript | package.json | *.js, *.jsx | `//` |
| Rust | Cargo.toml | *.rs | `//` |
| Java | pom.xml | *.java | `//` |
| Kotlin | build.gradle.kts | *.kt | `//` |
| C# | .csproj | *.cs | `//` |
| Ruby | Gemfile | *.rb | `#` |
| PHP | composer.json | *.php | `//` |
| Elixir | mix.exs | *.ex, *.exs | `#` |
| C++ | CMakeLists.txt | *.cpp, *.h | `//` |
| Scala | build.sbt | *.scala | `//` |
| R | DESCRIPTION | *.R, *.r | `#` |
| Flutter | pubspec.yaml | *.dart | `//` |
| Swift | Package.swift | *.swift | `//` |

#### Step 0.6.2: Scan Modified Files

- Get list of files changed since last sync (git diff)
- For each modified source file, check for @MX tags
- Identify functions/code blocks that should have tags but don't

#### Step 0.6.3: Add Missing Tags (Language-Aware)

For modified files missing @MX tags, use language-specific patterns:

**Backend Languages (Go, Python, Rust, Java, Kotlin, C#, Ruby, PHP, Elixir, C++, Scala)**:
1. **fan_in >= 3**: Add `@MX:ANCHOR` for functions/methods with many callers
2. **Language-specific WARN patterns**:
   - Go: `go func`, `go ` (goroutines without context)
   - Python: `async def`, `threading` (async/threading patterns)
   - Rust: `async fn`, `unsafe ` (async/unsafe blocks)
   - Java: `new Thread`, `Executor` (thread usage)
   - Kotlin: `GlobalScope`, `runBlocking` (coroutine issues)
   - C#: `Task.Run`, `Thread.` (async/threading)
   - Ruby: `Thread.new` (thread creation)
   - PHP: `async ` (async patterns)
   - Elixir: `Task.async`, `spawn` (async/process)
   - C++: `std::thread`, `new ` (thread/memory)
   - Scala: `Future.`, `new Thread` (async/thread)
3. **magic constants**: Add `@MX:NOTE` for unexplained values
4. **missing tests**: Add `@MX:TODO` for untested public functions

**Frontend Languages (TypeScript, JavaScript)**:
1. **fan_in >= 3**: Add `@MX:ANCHOR` for functions with many callers
2. **Promise chains**: Add `@MX:WARN` for Promise.all without error handling
3. **async/await**: Add `@MX:WARN` for async functions without try/catch
4. **magic constants**: Add `@MX:NOTE` for unexplained values
5. **missing tests**: Add `@MX:TODO` for untested functions

**Data Science Languages (R, Flutter/Dart)**:
1. **fan_in >= 3**: Add `@MX:ANCHOR` for functions with many callers
2. **Language-specific WARN patterns**:
   - R: `parallel::` (parallel processing)
   - Flutter: `Isolate.`, `Future.` (async/isolate patterns)
3. **magic constants**: Add `@MX:NOTE` for unexplained values
4. **missing tests**: Add `@MX:TODO` for untested functions

**Mobile (Swift)**:
1. **fan_in >= 3**: Add `@MX:ANCHOR` for functions with many callers
2. **Swift-specific WARN**: `Task.`, `DispatchQueue` (async/concurrency)
3. **magic constants**: Add `@MX:NOTE` for unexplained values
4. **missing tests**: Add `@MX:TODO` for untested functions

#### Step 0.6.4: Generate Tag Report

Include in sync report:
- Files scanned: N (by language)
- Tags added: N (by type, by language)
- Files requiring attention (high complexity, missing documentation)

#### MX Tag Integration

When MX tags are added during sync:
- Changes are included in the same commit as documentation updates
- Tag additions are noted in the PR description
- Report summarizes tag changes by category

Status mode early exit: If mode is "status", display quality report and exit. No further phases execute.

### Phase 0.7: Coverage Analysis and Test Generation

Purpose: Measure test coverage, identify gaps, and generate missing tests to meet coverage targets before documentation sync.

#### Step 0.7.1: Coverage Measurement

Agent: manager-develop subagent

Measure current coverage using language-specific tools:
- Go: `go test -coverprofile=coverage.out -covermode=atomic ./...` then `go tool cover -func=coverage.out`
- Python: `pytest --cov --cov-report=json`
- TypeScript/JavaScript: `vitest run --coverage` or `jest --coverage --json`
- Rust: `cargo llvm-cov --json`

Output: Overall coverage percentage, per-file coverage, per-function data.

#### Step 0.7.2: Gap Analysis

Agent: manager-develop subagent

Identify files below the coverage target (from quality.yaml test_coverage_target, default 85%).

Prioritize gaps by risk:
- P1 (Critical): Public API functions, high fan_in (>=3), functions with @MX:ANCHOR
- P2 (High): Business logic, error handling paths
- P3 (Medium): Internal utilities, helper functions
- P4 (Low): Generated code, configuration, trivial getters/setters

#### Step 0.7.3: Test Generation

Agent: manager-develop subagent

Generate missing tests for P1 and P2 gaps:
- Follow development_mode for test style (TDD: table-driven tests, DDD: characterization tests)
- Include edge cases and error scenarios
- Follow existing test patterns in the codebase
- Respect file naming conventions (*_test.go, *.test.ts, test_*.py)

#### Step 0.7.4: Verification

After test generation:
- Run the full test suite to ensure no regressions
- Re-measure coverage to confirm improvement
- Compare before/after coverage percentages

Behavior:
- If coverage target met: Proceed to Phase 1
- If coverage target not met after test generation: Log remaining gaps and proceed (do not block pipeline)

#### Step 0.7.5: Coverage Report

Include in sync quality report:
- Before/after coverage percentages
- Tests generated (count and file list)
- Remaining gaps if target not fully met
- Coverage by package/module breakdown
