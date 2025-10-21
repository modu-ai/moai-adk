---

name: moai-lang-julia
description: Julia best practices with Test stdlib, Pkg manager, and scientific computing patterns. Use when writing or reviewing Julia code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Julia Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Julia code discussions, framework guidance, or file extensions such as .jl. |
| Tier | 3 |

## What it does

Provides Julia-specific expertise for TDD development, including Test standard library, Pkg package manager, and high-performance scientific computing patterns.

## When to use

- Engages when the conversation references Julia work, frameworks, or files like .jl.
- "Writing Julia tests", "How to use Test stdlib", "Scientific computing"
- Automatically invoked when working with Julia projects
- Julia SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **Test**: Built-in testing library (@test, @testset)
- **Coverage.jl**: Test coverage analysis
- **BenchmarkTools.jl**: Performance benchmarking

**Package Management**:
- **Pkg**: Built-in package manager
- **Project.toml**: Package configuration
- **Manifest.toml**: Dependency lock file

**Code Quality**:
- **JuliaFormatter.jl**: Code formatting
- **Lint.jl**: Static analysis
- **JET.jl**: Type inference analysis

**Scientific Computing**:
- **Multiple dispatch**: Method specialization on argument types
- **Type stability**: Performance optimization
- **Broadcasting**: Element-wise operations (. syntax)
- **Linear algebra**: Built-in BLAS/LAPACK

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Type annotations for performance-critical code
- Prefer abstract types for function arguments
- Use @inbounds for performance (after bounds checking)
- Profile before optimizing

## Examples
```bash
julia --project -e 'using Pkg; Pkg.test()'
```

## Inputs
- 언어별 소스 디렉터리(e.g. `src/`, `app/`).
- 언어별 빌드/테스트 설정 파일(예: `package.json`, `pyproject.toml`, `go.mod`).
- 관련 테스트 스위트 및 샘플 데이터.

## Outputs
- 선택된 언어에 맞춘 테스트/린트 실행 계획.
- 주요 언어 관용구와 리뷰 체크포인트 목록.

## Failure Modes
- 언어 런타임이나 패키지 매니저가 설치되지 않았을 때.
- 다중 언어 프로젝트에서 주 언어를 판별하지 못했을 때.

## Dependencies
- Read/Grep 도구로 프로젝트 파일 접근이 필요합니다.
- `Skill("moai-foundation-langs")`와 함께 사용하면 교차 언어 규약 공유가 용이합니다.

## References
- Julia Language. "Documentation." https://docs.julialang.org/en/v1/ (accessed 2025-03-29).
- JuliaFormatter.jl. "JuliaFormatter Documentation." https://domluna.github.io/JuliaFormatter.jl/stable/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Julia-specific review)
- alfred-performance-optimizer (Julia profiling)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
