---
description: "Sync Phase 7~10 — Quality Verification, Security Scan, MX Tag Validation, and Coverage Analysis with Test Generation."
user-invocable: false
metadata:
  parent: moai-workflow-sync
  phase: "Phase 7~10: Quality Verification and Coverage"
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->
<!-- Emits one line per Phase entry/exit to stderr in format: [trace] /moai sync Phase <N> <enter|exit> -->

### Phase 7: Quality Verification

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

Snapshot consumption: before launching, query the shared diagnostic snapshot with `moai verify check --key-current`. **Where** a fresh snapshot — key equality AND within the TTL — already covers one of the three categories below (recorded by the run-phase pre-review gate at run Phase 15, or by sync Phase 1 on the unchanged tree), consume the recorded result instead of re-executing that category; the quality report cites the snapshot path, key, original command, and recorded exit code as that category's evidence (per `.claude/rules/moai/core/verification-claim-integrity.md` §2 — the snapshot is the observed evidence, and the freshness rule is what keeps the attribution valid). A stale snapshot is never cited: on key mismatch or TTL expiry, execute the check as below and record the fresh result via `moai verify record`.

**Shared-snapshot wiring.** The snapshot is keyed by HEAD SHA (HEAD + porcelain-v2 + diff hash); a new commit invalidates the prior snapshot. Three sync-phase consumers — the `sync-auditor` Evidence cells, the `.claude/hooks/moai/sync-phase-quality-gate.sh` Stop hook, and the `.claude/workflows/sync-audit-4dim.js` 4-dimension judges — all consume this single snapshot keyed by HEAD SHA rather than each independently re-executing `go test` / `golangci-lint` / `go vet` / `go test -cover`. Concurrent recording requests for the SAME HEAD SHA are serialized via the per-key claim/lock mechanism (in-process mutex + cross-process `O_EXCL` claim-stamp with staleness reclaim), so exactly one consumer's recording per dimension lands and the rest read — last-writer-wins never silently drops a dimension.

Launch three background tasks simultaneously:

- Test Runner: Language-specific test command (pytest, npm test, go test, cargo test, etc.)
- Linter: Language-specific lint command (ruff, eslint, golangci-lint, clippy, etc.)
- Type Checker: Language-specific type check (mypy, tsc --noEmit, go vet, etc.)

Collect all results with timeouts (180s for tests, 120s for others). Handle partial failures gracefully.

#### Step 0.5.3: Handle Test Failures

If any tests fail, use AskUserQuestion:

- Continue: Proceed with sync despite failures
- Abort: Stop sync, fix tests first (exit to Phase 14 graceful exit)

#### Step 0.5.4: Deep Code Review with Auto-Fix

Agent: sync-auditor subagent (independent quality scoring), gated by harness level:

- `minimal` harness level (`harness.yaml` `levels.minimal.evaluator: false`): skip the sync-auditor invocation; rely on the orchestrator verification batch (lint + test + coverage) instead
- `standard` / `thorough` harness level (`evaluator: true`): invoke sync-auditor for independent quality scoring

Execute multi-perspective code review beyond basic TRUST 5 validation, using the canonical sync-auditor rubric (`.claude/agents/moai/sync-auditor.md`):

Evaluation Dimensions:
- Functionality (40%): All SPEC acceptance criteria met
- Security (25%): OWASP Top 10 compliance, injection risks, secrets exposure, dependency vulnerabilities — HARD THRESHOLD: any Critical/High finding causes overall FAIL regardless of other scores
- Craft (20%): Test coverage >= 85%, error handling completeness, naming conventions, algorithmic complexity, concurrency safety
- Consistency (15%): Codebase pattern adherence, code style consistency

Web-output conditional: when the changed files emit pages a search engine or automated reader will fetch (markup templates, static-site content, server-rendered views), inject `At start, invoke Skill("moai-ref-seo") for the canonical-address, per-page metadata, and structured-data pre-ship baseline.` into the reviewing spawn (per `.moai/config/sections/delegation.yaml` domain_skills.frontend; skill-routing.md §1).

Auto-Fix Behavior:
- If critical issues found: Delegate auto-fix to manager-develop or a per-spawn `Agent(general-purpose)` domain specialist (per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C) — inject the cycle_type skill (`moai-workflow-ddd`|`moai-workflow-tdd`) plus 0-3 domain `moai-ref-*` skills per the mission domain (`.moai/config/sections/delegation.yaml`; per skill-routing.md §1)
- Re-run review after fix to verify resolution
- Maximum 3 auto-fix iterations for critical issues before escalating to user
- Warnings and suggestions are logged in report but do not block pipeline

Output:
- Review report with findings by severity (critical, warning, suggestion)
- @MX tag compliance status (integrated with Phase 9)
- Auto-fix log if corrections were applied

#### LSP Quality Gates

The sync phase enforces LSP-based quality gates as configured in quality.yaml:
- Zero errors required (lsp_quality_gates.sync.max_errors: 0)
- Maximum 10 warnings allowed (lsp_quality_gates.sync.max_warnings: 10)
- Clean LSP state required (lsp_quality_gates.sync.require_clean_lsp: true)

#### Step 0.5.5: Generate Quality Report

Aggregate all results into a quality report showing status for test-runner, linter, type-checker, and code-review. Determine overall status (PASS or WARN).

### Phase 8: Security Scan (Conditional)

Purpose: Run a targeted security audit on changed files before PR creation. Catches security vulnerabilities that code review alone may miss.

**Activation condition**: Execute this phase ONLY when changed files match security-sensitive patterns:
- Authentication/authorization files (auth, login, session, token, permission, role)
- Database interaction files (query, model, migration, schema, repository, dao)
- API endpoint files (handler, controller, route, endpoint, middleware)
- User input handling files (form, input, validation, sanitize)
- Configuration files with secrets (.env, config with credentials)

**Skip condition**: If no changed files match security-sensitive patterns, skip to Phase 9. Log: "Security scan skipped: no security-sensitive files changed."

#### Step 0.55.0: Factory dedup gate (conditional suppression of Step 0.55.1)

Applies only to a sync entered from a factory chain, whose run-phase verify stage already ran a whole-repository deep security scan. The gate decides one thing: whether that scan's evidence may stand in for the Step 0.55.1 analysis below.

**Scope of the suppression — Step 0.55.1 and nothing else.** A passing gate suppresses only the agent-invoked security analysis of Step 0.55.1. The dependency manifest audit below is a separate mechanism serving a separate purpose — detecting transitive-vulnerability drift unrelated to the current SPEC — and a source-code deep scan does not substitute for it, so it continues to run unconditionally whether or not this gate passes. Skipping the whole of Phase 8 would remove the only check for that drift.

**Procedure.** Derive both runtime inputs at sync entry rather than judging them:

- head SHA — the stdout of `git rev-parse HEAD`.
- tree-dirty flag — `git status --porcelain` reports at least one line after excluding `.moai/state/`, `.moai/reports/`, `.moai/cache/`, `.moai/logs/`, and the results directory itself.

Then call the **revision-match predicate** with the results directory recorded for the session plus those two derived inputs. The predicate returns TRUE only when: the results directory exists; its `findings.jsonl` exists and every line parses as JSON; its `revision.json` exists, is readable, and parses; the recorded `scanned_commit` equals the derived head SHA; the recorded scope is the whole repository; and — where the tree is dirty — the scan included the working tree. A clean tree makes the working-tree question irrelevant, because there were no uncommitted edits for the scan to have missed.

**Fail-safe default — the gate defaults to RUN.** Every conjunct that fails yields FALSE, and FALSE means Step 0.55.1 runs. An absent results directory, an absent or unreadable `revision.json`, malformed JSON, a commit mismatch, a narrower scope, an absent `findings.jsonl`, and a `findings.jsonl` line that does not parse are each FALSE. Absence is never a match. The one absence-shaped input that does match is a zero-line `findings.jsonl`: a clean scan writes that file empty, while an aborted scan characteristically never writes it at all — which is why the completeness check consults `findings.jsonl` rather than `revision.json`, whose schema carries no completion field and therefore cannot distinguish an aborted run from a finished one.

**Rung allow-list.** A passing predicate is only half the condition. Suppression additionally requires the verify stage's rigor rung to have been **recorded** on the session state record and to equal `PRIMARY` or `FALLBACK`. Every other value yields no suppression: `DEGRADED` (a single-pass, rigor-reduced scan, which must never suppress the independent adversarial analysis of the same surface), an unrecognized value, an empty value, and a rung that was never recorded at all. State this as an allow-list, never as a "not `DEGRADED`" exclusion — the state record is best-effort and its fields land independently, so a record carrying a results directory but no rung is reachable, and a deny-list would read that absence as permission to suppress. The scan's effort tier is deliberately not part of this condition: an effort level is not a rigor rung, so a maximum-effort single-pass scan would clear an effort floor while still being rigor-reduced.

**Disclosure.** When the gate suppresses Step 0.55.1, the sync report MUST record the inheritance explicitly — naming the results directory and the matched `scanned_commit`, and stating that the findings were inherited from the factory verify stage. A suppressed scan must never be indistinguishable from a clean one.

#### Step 0.55.1: Security Analysis

Agent: per-spawn `Agent(general-purpose)` security reviewer (security whitelist per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 9).

Delegate to a per-spawn `Agent(general-purpose)` security reviewer loading the retained `moai-ref-owasp-checklist` / `moai-ref-secops` skills (the documented security replacement path) in inline mode:
- Only CRITICAL findings block the sync pipeline
- HIGH findings are reported as warnings in PR description
- MEDIUM and LOW findings are logged in sync report

**Dependency manifest audit (always runs, regardless of whether manifest files changed in this SPEC)** — a SEPARATE, automatic mechanism distinct from the agent-invoked security analysis above:

Audit ALL of the following manifest files present at project root — dependency surface must be checked at every sync to detect drift from transitive vulnerability changes unrelated to this SPEC:
`go.mod`, `package.json`, `requirements.txt`, `Cargo.toml`, `pyproject.toml`, `Gemfile`, `composer.json`, `mix.exs`, `Package.swift`, `pubspec.yaml`.

When any manifest is detected, the dependency vulnerability scan runs automatically via the Stop hook (`.claude/hooks/moai/sync-phase-quality-gate.sh`) as a mechanical check outside agent delegation. If the Stop hook path is unavailable, a per-spawn `Agent(general-purpose)` security reviewer MAY perform the same scan as a fallback — inject `At start, invoke Skill("moai-ref-supply-chain") for the dependency / transitive-vulnerability baseline.` (per skill-routing.md §1).
Rationale: a transitive vulnerability may have been introduced by an unrelated dependency update since the last sync, even if no manifest file was modified in the current SPEC.

#### Step 0.55.2: Security Gate Decision

If CRITICAL findings exist:
- Present findings via AskUserQuestion:
  - Fix now (Recommended): Delegate to a per-spawn `Agent(general-purpose)` security reviewer for auto-fix, then re-scan
  - Continue with warning: Proceed to Phase 9 with security warnings embedded in PR description
  - Abort: Exit sync workflow

If no CRITICAL findings: Proceed to Phase 9. Include any HIGH/MEDIUM findings in the sync report.

### Phase 9: MX Tag Validation (Multi-Language)

Purpose: Ensure code has appropriate @MX annotations for AI agent context. Supports all 16 MoAI-ADK languages.

**[HARD] P1/P2 violations BLOCK sync.** If any P1 (missing @MX:ANCHOR on fan_in >= 3 function) or P2 (missing @MX:WARN on goroutine pattern) violations are found, sync is halted and the user must resolve them before proceeding.

- P1 (Blocking): exported function with fan_in >= 3 missing @MX:ANCHOR
- P2 (Blocking): goroutine/async pattern missing @MX:WARN
- P3 (Advisory): long exported function missing @MX:NOTE — warning only, sync continues
- P4 (Advisory): untested public function missing @MX:TODO — warning only, sync continues

When P1/P2 violations are detected:
1. Display full violation report with file:line references
2. Show message: "Run /moai run to add missing tags, or use --skip-mx to bypass"
3. Halt sync — do NOT proceed to Phase 10+

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

Sharding (`FO-SYNC-2`, read-only): **Where** the modified files span several languages or packages, the orchestrator shall shard this scan — one read-only `Agent()` per shard in a single turn, 3-5 concurrent per the Mode 4 ceiling (`.claude/rules/moai/workflow/orchestration-mode-selection.md` §C.2), each returning its findings as text and writing nothing. A shard missing a required input returns a structured blocker report (`.claude/rules/moai/core/agent-common-protocol.md` § Blocker Report Format) rather than prompting the user. Step 0.6.3 tag application stays single-writer, and the blocking P1/P2 determination above is made on the merged result so no shard can clear the gate alone. The orchestrator launches the shards itself — scaling, not subagent nesting. **Where** the changeset sits in a single language and package, or the concurrency ceiling leaves no room to shard, the scan runs as one pass with identical output — no error, no warning.

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

### Phase 10: Coverage Analysis and Test Generation

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

Per-package fan-out (`FO-SYNC-3`, read-only drafting): **Where** the gaps span several independent packages, the orchestrator shall draft their tests in parallel — one read-only `Agent()` per package in a single turn, 3-5 concurrent per the Mode 4 ceiling (`.claude/rules/moai/workflow/orchestration-mode-selection.md` §C.2). Each drafter reads its package plus that package's gap list and returns test source as text; it writes no file and never prompts the user, returning a structured blocker report (`.claude/rules/moai/core/agent-common-protocol.md` § Blocker Report Format) when an input is missing. The single `manager-develop` subagent applies the drafts and runs Step 0.7.4 — it remains the only writer, so two write-capable agents never run at once. The orchestrator launches the drafters itself — scaling, not subagent nesting. **Where** the gaps sit in one package, the existing serial path runs unchanged.

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
