# Multi-Language Toolchain Reference

Per-language testing / lint / security / performance tool inventory. This is the
single home for language-specific tooling in this skill — module headers name a
concern (e.g. "static analyzers for your language") and defer the concrete tool
list to this file. The workflow adapts to each language's standard ecosystem;
all 16 supported languages are treated equally.

## Python

| Concern | Tool |
|---------|------|
| Testing | pytest, unittest |
| Coverage | coverage.py, pytest-cov |
| Static analysis / style | pylint, flake8, ruff |
| Type checking | mypy, pyright |
| Security | bandit |
| Performance | cProfile, memory_profiler, psutil, line_profiler, tracemalloc |

## JavaScript / TypeScript

| Concern | Tool |
|---------|------|
| Testing | Jest, Vitest, Mocha |
| Coverage | c8, nyc, Jest --coverage |
| Static analysis / style | ESLint, Biome, Prettier |
| Type checking | tsc (TypeScript) |
| Security | npm audit, eslint-plugin-security |
| Performance | Chrome DevTools, Lighthouse, clinic.js |

## Go

| Concern | Tool |
|---------|------|
| Testing | go test (built-in), testify |
| Coverage | go test -cover, go test -coverprofile |
| Static analysis / style | gofmt, go vet, golangci-lint, staticcheck |
| Security | gosec, govulncheck |
| Performance | pprof (built-in), benchstat |

## Rust

| Concern | Tool |
|---------|------|
| Testing | cargo test |
| Coverage | tarpaulin, llvm-cov |
| Static analysis / style | rustfmt, clippy |
| Security | cargo audit, cargo deny |
| Performance | flamegraph, cargo flamechart |

## Java / Kotlin

| Concern | Tool |
|---------|------|
| Testing | JUnit 5, TestNG, Kotest (Kotlin) |
| Coverage | JaCoCo |
| Static analysis / style | Checkstyle, SpotBugs, detekt (Kotlin), ktlint (Kotlin) |
| Security | SpotBugs with Find Security Bugs |
| Performance | async-profiler, JMH |

## C# / .NET

| Concern | Tool |
|---------|------|
| Testing | xUnit, NUnit, MSTest |
| Coverage | coverlet, dotnet-coverage |
| Static analysis / style | dotnet format, Roslyn analyzers, SonarAnalyzer |
| Security | dotnet-format security rules, Snyk |
| Performance | dotnet-trace, dotnet-counters, PerfView |

## Ruby

| Concern | Tool |
|---------|------|
| Testing | RSpec, Minitest |
| Coverage | SimpleCov |
| Static analysis / style | RuboCop, Standard |
| Security | Brakeman, bundler-audit |
| Performance | stackprof, rbspy |

## PHP

| Concern | Tool |
|---------|------|
| Testing | PHPUnit, Pest |
| Coverage | Xdebug, phpunit --coverage |
| Static analysis / style | PHP_CodeSniffer, PHP-CS-Fixer |
| Type / static analysis | PHPStan, Psalm |
| Security | Composer audit, Psalm security checks |
| Performance | Xdebug profiler, Tideways |

## Swift

| Concern | Tool |
|---------|------|
| Testing | XCTest, Swift Testing |
| Coverage | xccov, Slather |
| Static analysis / style | SwiftFormat, SwiftLint |
| Security | Apple sanitizer address/thread |
| Performance | Instruments (Time Profiler, Allocations) |

## Flutter / Dart

| Concern | Tool |
|---------|------|
| Testing | flutter test, integration_test |
| Coverage | flutter test --coverage, lcov |
| Static analysis / style | dart format, flutter analyze (dartanalyzer) |
| Security | dependency security scan via pub |
| Performance | Flutter DevTools (CPU/memory profiler) |

## Elixir

| Concern | Tool |
|---------|------|
| Testing | ExUnit (built-in) |
| Coverage | ExCoveralls |
| Static analysis / style | mix format, Credo |
| Security | mix sobelow |
| Performance | :fprof, :eprof, Benchee (benchmarking) |

## C / C++

| Concern | Tool |
|---------|------|
| Testing | Google Test, Catch2, doctest |
| Coverage | gcov / lcov, llvm-cov |
| Static analysis / style | clang-format, clang-tidy, cppcheck |
| Security | clang static analyzer, Coverity |
| Performance | perf, Valgrind/Callgrind, gprof |

## Scala

| Concern | Tool |
|---------:|------|
| Testing | ScalaTest, munit |
| Coverage | scoverage |
| Static analysis / style | scalafmt, scalafix |
| Security | dependency-check |
| Performance | async-profiler (JVM) |

## R

| Concern | Tool |
|---------|------|
| Testing | testthat |
| Coverage | covr |
| Static analysis / style | lintr, styler |
| Security | security checks via oysteR |
| Performance | Rprof (built-in), profvis |

## Choosing Tools for a Language Not Listed

Map each concern — testing, coverage, static analysis/style, security,
performance — to the standard tool for the target language. When unsure, use
WebSearch/WebFetch (`mcp__docs__resolve-library-id`) to look up the current
recommended testing/profiling library for the language. If a recognized tool is
not installed, the quality gate skips it gracefully (per CLAUDE.md §7
Language-Specific Guidelines).
