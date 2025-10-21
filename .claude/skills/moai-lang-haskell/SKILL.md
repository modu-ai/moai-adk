---

name: moai-lang-haskell
description: Haskell best practices with HUnit, Stack/Cabal, and pure functional programming. Use when writing or reviewing Haskell code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Haskell Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Haskell code discussions, framework guidance, or file extensions such as .hs. |
| Tier | 3 |

## What it does

Provides Haskell-specific expertise for TDD development, including HUnit testing, Stack/Cabal build tools, and pure functional programming with strong type safety.

## When to use

- Engages when the conversation references Haskell work, frameworks, or files like .hs.
- “Writing Haskell tests”, “How to use HUnit”, “Pure functional programming”
- Automatically invoked when working with Haskell projects
- Haskell SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **HUnit**: Unit testing framework
- **QuickCheck**: Property-based testing
- **Hspec**: BDD-style testing
- Test coverage with hpc

**Build Tools**:
- **Stack**: Reproducible builds, dependency resolution
- **Cabal**: Haskell package system
- **hpack**: Alternative package description

**Code Quality**:
- **hlint**: Haskell linter
- **stylish-haskell**: Code formatting
- **GHC warnings**: Compiler-level checks

**Functional Programming**:
- **Pure functions**: No side effects
- **Monads**: IO, Maybe, Either, State
- **Functors/Applicatives**: Abstraction patterns
- **Type classes**: Polymorphism
- **Lazy evaluation**: Infinite data structures

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Prefer total functions (avoid partial)
- Type-driven development
- Point-free style (when readable)
- Avoid do-notation overuse

## Examples
```bash
cabal test && hlint src
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
- Haskell.org. "Haskell Language Documentation." https://www.haskell.org/documentation/ (accessed 2025-03-29).
- GitHub. "HLint." https://github.com/ndmitchell/hlint (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Haskell-specific review)
- alfred-refactoring-coach (functional refactoring)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
