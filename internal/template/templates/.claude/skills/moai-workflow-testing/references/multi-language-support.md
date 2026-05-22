# Multi-Language Toolchain Reference

Per-language testing/lint/security/performance tool inventory. The workflow adapts to each language's standard ecosystem.

## Python Projects

| Concern | Tool |
|---------|------|
| Testing | pytest |
| Static analysis | pylint, flake8 |
| Security | bandit |
| Performance | cProfile, memory_profiler, psutil, line_profiler |
| Type checking | mypy |

## JavaScript / TypeScript Projects

| Concern | Tool |
|---------|------|
| Testing | Jest, Vitest |
| Static analysis | ESLint |
| Security | npm audit |
| Performance | Chrome DevTools, Lighthouse |

## Go Projects

| Concern | Tool |
|---------|------|
| Testing | go test |
| Static analysis | golint, staticcheck |
| Security | gosec |
| Performance | pprof |

## Rust Projects

| Concern | Tool |
|---------|------|
| Testing | cargo test |
| Static analysis | clippy |
| Security | cargo audit |
| Performance | flamegraph |

## Technology Stack Reference

Standard libraries the workflow leverages across languages:

- Analysis Libraries: cProfile (Python profiling), memory_profiler (memory analysis), psutil (system monitoring), line_profiler (line-level profiling)
- Static Analysis Tools: pylint (Python code analysis), flake8 (style guide), bandit (security scan), mypy (static types)
- Testing Frameworks: pytest (Python fixtures and plugins), unittest (standard library), coverage (coverage measurement)
