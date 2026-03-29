---
name: moai-workflow-gate
description: >
  Lightweight pre-commit quality gate running lint, format, type-check, and tests
  in parallel. Fast validation (<30s) without full code review or coverage analysis.
  Use before any commit for quick quality assurance.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-03-29"
  tags: "gate, quality, lint, format, test, pre-commit, check"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000

# MoAI Extension: Triggers
triggers:
  keywords: ["gate", "check", "pre-commit", "lint", "format", "quality gate"]
  agents: ["manager-quality"]
  phases: ["gate"]
---

# Gate Workflow Orchestration

## Purpose

Lightweight pre-commit quality gate. Runs lint, format check, type check, and tests in parallel for fast validation. Designed to complete in under 30 seconds for typical projects.

## Difference from Other Workflows

| Workflow | Scope | Speed | Use Case |
|----------|-------|-------|----------|
| `/moai gate` | lint + format + type-check + test | Fast (<30s) | Before every commit |
| `/moai review` | 4-perspective deep code review | Medium (2-5min) | Before PR, design review |
| sync Phase 0.5 | Full quality + code review + coverage | Slow (5-10min) | Part of sync pipeline |

## Input

- $ARGUMENTS: Optional flags
  - --fix: Auto-fix lint and format issues (default: report only)
  - --staged: Only check staged files (`git diff --staged`)
  - --file PATH: Check specific file(s) only

## Phase 1: Language Detection

Check indicator files in priority order (first match wins):

- Go: go.mod → `go vet`, `golangci-lint run`, `go test -race`
- Python: pyproject.toml → `ruff check`, `ruff format --check`, `mypy`, `pytest`
- TypeScript: tsconfig.json → `tsc --noEmit`, `eslint`, `vitest run` or `jest`
- JavaScript: package.json → `eslint`, `prettier --check`, `npm test`
- Rust: Cargo.toml → `cargo check`, `cargo clippy`, `cargo test`
- Ruby: Gemfile → `rubocop`, `bundle exec rspec`
- Java: pom.xml → `mvn compile`, `mvn test`
- PHP: composer.json → `php-cs-fixer`, `phpstan`, `phpunit`
- Kotlin: build.gradle.kts → `ktlint`, `gradle test`
- Swift: Package.swift → `swiftlint`, `swift test`
- C#: .csproj → `dotnet build`, `dotnet test`
- C++: CMakeLists.txt → `cmake --build`, `ctest`
- Elixir: mix.exs → `mix format --check-formatted`, `mix credo`, `mix test`
- R: DESCRIPTION → `lintr::lint_package()`, `testthat::test_package()`
- Flutter: pubspec.yaml → `dart analyze`, `flutter test`
- Scala: build.sbt → `sbt compile`, `sbt test`
- Fallback: Skip language-specific checks, report "unknown language"

## Phase 2: Execute Checks in Parallel

Launch all checks simultaneously using Bash with background execution:

### Check Categories

**Lint Check**: Language-specific linter
- Reports style violations, unused imports, dead code
- With --fix flag: Auto-corrects fixable issues

**Format Check**: Code formatting verification
- Reports unformatted files
- With --fix flag: Auto-formats files

**Type Check**: Static type analysis
- Reports type errors and missing annotations
- No auto-fix (requires manual intervention)

**Test Check**: Run test suite
- Reports test failures
- No auto-fix (failing tests need investigation)

### Timeout

- Individual check timeout: 60 seconds
- Overall gate timeout: 90 seconds
- If any check times out: Report as WARNING, do not block

### Scoping

When --staged flag is provided:
- Only check files in `git diff --staged --name-only`
- Pass file list to lint/format/type-check commands where supported
- Tests always run full suite (scoped tests may miss regressions)

When --file flag is provided:
- Only check specified file(s)
- Tests run only matching test files if discoverable

## Phase 3: Report Results

### Pass (All Green)

```markdown
## Quality Gate: PASS
| Check  | Status | Time  |
|--------|--------|-------|
| Lint   | PASS   | 2.1s  |
| Format | PASS   | 0.8s  |
| Type   | PASS   | 3.2s  |
| Test   | PASS   | 12.4s |
Total: 18.5s
```

### Fail (Issues Found)

```markdown
## Quality Gate: FAIL
| Check  | Status | Issues | Time  |
|--------|--------|--------|-------|
| Lint   | FAIL   | 3      | 2.1s  |
| Format | WARN   | 1      | 0.8s  |
| Type   | PASS   | 0      | 3.2s  |
| Test   | FAIL   | 2      | 12.4s |

### Lint Issues
- src/auth.go:45: unused variable 'token'
- src/handler.go:12: missing error check

### Test Failures
- TestLogin: expected 200, got 401
```

### Auto-Fix Results (--fix flag)

```markdown
## Quality Gate: FIXED
| Check  | Before | After | Fixed |
|--------|--------|-------|-------|
| Lint   | 3      | 0     | 3     |
| Format | 1      | 0     | 1     |
| Type   | 0      | 0     | -     |
| Test   | 2      | 2     | 0     |
Note: 2 test failures require manual investigation.
```

## Phase 4: Next Steps

If all checks pass: Display success message, ready to commit.

If checks fail without --fix:
- Present via AskUserQuestion:
  - Auto-fix (Recommended): Re-run with --fix to auto-correct lint and format issues
  - Run /moai fix: Use full fix workflow for deeper issue resolution
  - Ignore and continue: Proceed despite issues (not recommended)

If checks fail after --fix (remaining unfixable issues):
- Display remaining issues
- Suggest manual investigation for test failures and type errors

## Completion Criteria

- All 4 check categories executed (or timed out with warning)
- Results displayed in table format
- Auto-fix applied if --fix flag was set
- Next steps presented if issues found

---

Version: 1.0.0
Updated: 2026-03-29
