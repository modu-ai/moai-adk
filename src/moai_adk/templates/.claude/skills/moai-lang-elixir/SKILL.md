---

name: moai-lang-elixir
description: Elixir best practices with ExUnit, Mix, and OTP patterns. Use when writing or reviewing Elixir code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Elixir Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Elixir code discussions, framework guidance, or file extensions such as .ex/.exs. |
| Tier | 3 |

## What it does

Provides Elixir-specific expertise for TDD development, including ExUnit testing, Mix build tool, and OTP (Open Telecom Platform) patterns for concurrent systems.

## When to use

- Engages when the conversation references Elixir work, frameworks, or files like .ex/.exs.
- "Writing Elixir tests", "How to use ExUnit", "OTP patterns"
- Automatically invoked when working with Elixir/Phoenix projects
- Elixir SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **ExUnit**: Built-in test framework
- **Mox**: Mocking library
- **StreamData**: Property-based testing
- Test coverage with `mix test --cover`

**Build Tools**:
- **Mix**: Build tool and project manager
- **mix.exs**: Project configuration
- **Hex**: Package manager

**Code Quality**:
- **Credo**: Static code analysis
- **Dialyzer**: Type checking
- **mix format**: Code formatting

**OTP Patterns**:
- **GenServer**: Generic server behavior
- **Supervisor**: Process supervision
- **Application**: Application behavior
- **Task**: Async/await operations

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Pattern matching over conditionals
- Pipe operator (|>) for data transformations
- Immutable data structures
- "Let it crash" philosophy with supervisors

## Examples
```bash
mix test && mix credo --strict
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
- Elixir Lang. "Getting Started." https://elixir-lang.org/getting-started/introduction.html (accessed 2025-03-29).
- Credo. "Credo — The Elixir Linter." https://hexdocs.pm/credo/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Elixir-specific review)
- web-api-expert (Phoenix API development)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
