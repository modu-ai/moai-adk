---

name: moai-lang-lua
description: Lua best practices with busted, luacheck, and embedded scripting patterns. Use when writing or reviewing Lua code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Lua Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Lua code discussions, framework guidance, or file extensions such as .lua. |
| Tier | 3 |

## What it does

Provides Lua-specific expertise for TDD development, including busted testing framework, luacheck linting, and embedded scripting patterns for game development and system configuration.

## When to use

- Engages when the conversation references Lua work, frameworks, or files like .lua.
- "Writing Lua tests", "How to use busted", "Embedded scripting"
- Automatically invoked when working with Lua projects
- Lua SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **busted**: Elegant Lua testing framework
- **luassert**: Assertion library
- **lua-coveralls**: Coverage reporting
- BDD-style test writing

**Code Quality**:
- **luacheck**: Lua linter and static analyzer
- **StyLua**: Code formatting
- **luadoc**: Documentation generation

**Package Management**:
- **LuaRocks**: Package manager
- **rockspec**: Package specification

**Lua Patterns**:
- **Tables**: Versatile data structure
- **Metatables**: Operator overloading
- **Closures**: Function factories
- **Coroutines**: Cooperative multitasking

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Use `local` for all variables
- Prefer tables over multiple return values
- Document public APIs
- Avoid global variables

## Examples
```bash
luacheck src && busted
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
- Lua.org. "Programming in Lua." https://www.lua.org/pil/contents.html (accessed 2025-03-29).
- Olivine Labs. "busted." https://olivinelabs.com/busted/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Lua-specific review)
- cli-tool-expert (Lua scripting utilities)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
