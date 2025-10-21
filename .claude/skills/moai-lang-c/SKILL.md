---

name: moai-lang-c
description: C best practices with Unity test framework, cppcheck, and Make build system. Use when writing or reviewing C code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# C Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | C code discussions, framework guidance, or file extensions such as .c/.h. |
| Tier | 3 |

## What it does

Provides C-specific expertise for TDD development, including Unity test framework, cppcheck static analysis, and Make build system for system programming.

## When to use

- Engages when the conversation references C work, frameworks, or files like .c/.h.
- "Writing C tests", "Unity test framework", "Embedded C"
- Automatically invoked when working with C projects
- C SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **Unity**: Lightweight C test framework
- **CMock**: Mocking framework for C
- **Ceedling**: Build automation for C
- Test coverage with gcov

**Build Tools**:
- **Make**: Standard build automation
- **CMake**: Modern build system
- **GCC/Clang**: C compilers

**Code Quality**:
- **cppcheck**: Static code analysis
- **Valgrind**: Memory leak detection
- **splint**: Secure programming lint

**C Patterns**:
- **Opaque pointers**: Information hiding
- **Function pointers**: Callback mechanisms
- **Error codes**: Integer return values
- **Manual memory management**: malloc/free discipline

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Always check malloc return values
- Free every malloc
- Avoid buffer overflows (use strncpy, snprintf)
- Use const for read-only parameters
- Initialize all variables

## Examples
```bash
make test && cppcheck src
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
- ISO. "ISO/IEC 9899:2018 Programming Languages — C." (accessed 2025-03-29).
- Cppcheck. "Cppcheck Manual." http://cppcheck.sourceforge.net/manual.pdf (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (C-specific review)
- alfred-debugger-pro (C debugging)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
