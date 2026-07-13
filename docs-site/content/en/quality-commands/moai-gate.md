---
title: /moai gate
weight: 15
draft: false
---

A lightweight pre-commit quality-gate command that runs lint, format, type-check, and test in parallel. It is designed to complete within 30 seconds and is used for a fast verification right before every commit.

{{< callout type="info" >}}
**Slash command**: In Claude Code, type `/moai gate` to run this command directly.
{{< /callout >}}

## Overview

`/moai gate` is a lightweight quality gate used right before a commit. It runs four verifications — lint + format check + type-check + test — in parallel, designed to finish within 30 seconds. It does not perform heavy analysis like code review (`/moai review`) or sync Phase 0.5; it serves as the quick safety net you use before committing in your regular workflow.

The v3 tokenomics principle of "verification depth matched to the size of the work" is this command's design rationale — instead of running a 5-10-minute full quality pipeline on every commit, a 30-second gate filters out most defects before the commit.

## Command Format

```bash
/moai gate [--fix] [--staged] [--file PATH]
```

- With no arguments, the 4 verifications run in parallel over the whole project.
- The `--mode pipeline` argument triggers a `MODE_PIPELINE_ONLY_UTILITY` error (`/moai gate` is not a multi-agent class).

## Options

### `--fix`

Directly fixes lint/format items that can be auto-fixed. The default behavior only prints a report and makes no changes.

- **Recommended timing**: use once right after writing new code to clean up style defects up front.
- **Caution**: after an auto-fix, always review the changes with `git diff`.

### `--staged`

Verifies only the staged files identified by `git diff --staged`.

- Can further shorten pre-commit verification time in a large monorepo.

### `--file PATH`

Verifies only the specified single file (or glob). Useful while debugging.

## The 4 Checks Run in Parallel

`/moai gate` runs the following four verifications **simultaneously** (completion time is determined by the slowest check).

| Check | Role | Main tools (auto-detected) |
|-------|------|-----------------------|
| Lint | Reports style violations, unused imports, dead code | `golangci-lint`, `ruff`, `eslint`, `clippy`, `rubocop`, `mvn compile`, `php-cs-fixer`, `ktlint`, `swiftlint`, `dotnet build`, `cmake --build`, `mix credo`, `lintr`, `dart analyze`, `sbt compile` |
| Format check | Detects formatting violations (auto-fix requires `--fix`) | `gofmt`, `ruff format --check`, `prettier --check`, `cargo fmt --check`, `rubocop`, `php-cs-fixer`, `ktlint`, `swift-format`, `dotnet format --verify-no-changes`, `clang-format`, `mix format --check-formatted` |
| Type check | Static type verification | `go vet`, `mypy`, `tsc --noEmit`, `cargo check`, `phpstan`, `dotnet build`, `cmake` |
| Test | Runs unit/integration tests | `go test -race`, `pytest`, `vitest`/`jest`, `cargo test`, `bundle exec rspec`, `mvn test`, `phpunit`, `gradle test`, `swift test`, `ctest`, `mix test`, `testthat`, `flutter test`, `sbt test` |

## 16-Language Auto-Detection

`/moai gate` checks indicator files at the project root in priority order and uses the toolchain of the first match. All 16 languages are supported equally.

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

If no indicator matches, the language checks are skipped and the report says `unknown language`.

## /moai gate vs /moai review vs sync Phase 0.5

The verification tools are divided into three depths, and the principle is to choose the depth matching the situation.

| Workflow | Scope | Speed | When to use |
|----------|------|------|-----------|
| `/moai gate` | lint + format + type-check + test | Fast (<30 s) | Right before every commit |
| `/moai review` | 4-perspective in-depth code review | Medium (2-5 min) | Right before a PR, design review |
| sync Phase 0.5 | Full quality + code review + coverage | Slow (5-10 min) | Part of the `/moai sync` pipeline |

## Usage Examples

```bash
# 1) Quick verification right before a commit
/moai gate

# 2) Auto-fix lint/format, then re-verify
/moai gate --fix

# 3) Verify staged files only (recommended for large monorepos)
/moai gate --staged

# 4) Verify a specific file only
/moai gate --file internal/cli/run.go
```

## Related Resources

- [`.claude/skills/moai/workflows/gate.md`](https://github.com/modu-ai/moai-adk) — workflow body SSOT
- [`/moai review`](/en/quality-commands/moai-review) — 4-perspective code review
- [`/moai sync`](/en/workflow-commands/moai-sync) — includes sync Phase 0.5 quality verification
- [`/moai fix`](/en/utility-commands/moai-fix) — the auto-fix pipeline
