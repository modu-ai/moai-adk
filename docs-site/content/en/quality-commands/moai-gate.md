---
title: /moai gate
weight: 15
draft: false
---

A lightweight pre-commit quality gate that runs lint, format check, type-check, and tests in parallel. Designed to complete within 30 seconds, ideal as the last fast safety net before every commit.

{{< callout type="info" >}}
**Slash command**: Type `/moai gate` in Claude Code to invoke this command directly.
{{< /callout >}}

## Overview

`/moai gate` is the fast pre-commit quality gate. It runs lint + format check + type-check + test in parallel, completing in under 30 seconds for typical projects. It is intentionally lighter than `/moai review` and sync Phase 0.5: it focuses on catching mechanical defects, not deep design review.

## Command Syntax

```bash
/moai gate [--fix] [--staged] [--file PATH]
```

- An empty argument set runs all four checks against the entire project.
- The `--mode pipeline` flag triggers a `MODE_PIPELINE_ONLY_UTILITY` error — `/moai gate` is not a multi-agent class command.

## Options

### `--fix`

Auto-corrects fixable lint and format issues. The default behavior only reports.

- **When to use**: Right after authoring new code, to clean up style defects in one pass.
- **Note**: Always review changes with `git diff` after running with `--fix`.

### `--staged`

Restricts checks to files identified by `git diff --staged`.

- Useful in large monorepos to keep pre-commit checks fast.

### `--file PATH`

Checks only the specified file (or glob). Handy during debugging.

## Four Checks Run in Parallel

`/moai gate` runs four checks **concurrently**; the wall-clock time equals the slowest check.

| Check | Role | Primary tools (auto-detected) |
|-------|------|-------------------------------|
| Lint | Style violations, unused imports, dead code | `golangci-lint`, `ruff`, `eslint`, `clippy`, `rubocop`, `mvn compile`, `php-cs-fixer`, `ktlint`, `swiftlint`, `dotnet build`, `cmake --build`, `mix credo`, `lintr`, `dart analyze`, `sbt compile` |
| Format check | Detects formatting violations (auto-fix requires `--fix`) | `gofmt`, `ruff format --check`, `prettier --check`, `cargo fmt --check`, `rubocop`, `php-cs-fixer`, `ktlint`, `swift-format`, `dotnet format --verify-no-changes`, `clang-format`, `mix format --check-formatted` |
| Type check | Static type validation | `go vet`, `mypy`, `tsc --noEmit`, `cargo check`, `phpstan`, `dotnet build`, `cmake` |
| Test | Unit and integration tests | `go test -race`, `pytest`, `vitest`/`jest`, `cargo test`, `bundle exec rspec`, `mvn test`, `phpunit`, `gradle test`, `swift test`, `ctest`, `mix test`, `testthat`, `flutter test`, `sbt test` |

## 16-Language Auto-detection

`/moai gate` checks indicator files at the project root in priority order; the first match selects the toolchain.

1. Go: `go.mod`
2. Python: `pyproject.toml`
3. TypeScript: `tsconfig.json`
4. JavaScript: `package.json`
5. Rust: `Cargo.toml`
6. Ruby: `Gemfile`
7. Java: `pom.xml`
8. PHP: `composer.json`
9. Kotlin: `build.gradle.kts`
10. Swift: `Package.swift`
11. C#: `.csproj`
12. C++: `CMakeLists.txt`
13. Elixir: `mix.exs`
14. R: `DESCRIPTION`
15. Flutter: `pubspec.yaml`
16. Scala: `build.sbt`

If no indicator matches, language-specific checks are skipped and the run reports `unknown language`.

## /moai gate vs /moai review vs sync Phase 0.5

| Workflow | Scope | Speed | When to use |
|----------|-------|-------|-------------|
| `/moai gate` | lint + format + type-check + test | Fast (<30 s) | Before every commit |
| `/moai review` | 4-perspective deep code review | Medium (2-5 min) | Before PR, design review |
| sync Phase 0.5 | Full quality + code review + coverage | Slow (5-10 min) | Part of the `/moai sync` pipeline |

## Examples

```bash
# 1) Fast pre-commit verification
/moai gate

# 2) Auto-fix lint/format then re-check
/moai gate --fix

# 3) Restrict to staged files (recommended for monorepos)
/moai gate --staged

# 4) Check a single file
/moai gate --file internal/cli/run.go
```

## References

- [`.claude/skills/moai/workflows/gate.md`](https://github.com/modu-ai/moai-adk) — workflow body SSOT
- [`/moai review`](/en/quality-commands/moai-review) — 4-perspective code review
- [`/moai sync`](/en/workflow-commands/moai-sync) — includes the sync Phase 0.5 quality run
- [`/moai fix`](/en/utility-commands/moai-fix) — automated repair pipeline
