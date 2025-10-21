---

name: moai-lang-cpp
description: C++ best practices with Google Test, clang-format, and modern C++ (C++17/20). Use when writing or reviewing C++ code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# C++ Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | C++ code discussions, framework guidance, or file extensions such as .cpp/.hpp. |
| Tier | 3 |

## What it does

Provides C++-specific expertise for TDD development, including Google Test framework, clang-format formatting, and modern C++ (C++17/20) features.

## When to use

- Engages when the conversation references C++ work, frameworks, or files like .cpp/.hpp.
- "Writing C++ tests", "How to use Google Test", "Modern C++"
- Automatically invoked when working with C++ projects
- C++ SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **Google Test (gtest)**: Unit testing framework
- **Google Mock (gmock)**: Mocking framework
- **Catch2**: Alternative testing framework
- Test coverage with gcov/lcov

**Build Tools**:
- **CMake**: Cross-platform build system
- **Make**: Traditional build tool
- **Conan/vcpkg**: Package managers

**Code Quality**:
- **clang-format**: Code formatting
- **clang-tidy**: Static analysis
- **cppcheck**: Additional static analysis

**Modern C++ Features**:
- **Smart pointers**: unique_ptr, shared_ptr, weak_ptr
- **Move semantics**: std::move, rvalue references
- **Lambda expressions**: Inline functions
- **auto keyword**: Type inference
- **constexpr**: Compile-time evaluation
- **std::optional**: Nullable types (C++17)
- **Concepts**: Type constraints (C++20)

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- RAII (Resource Acquisition Is Initialization)
- Rule of Five (destructor, copy/move constructors/assignments)
- Prefer stack allocation over heap
- Const correctness

## Examples
```bash
cmake --build build --target test && clang-tidy src/*.cpp
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
- ISO. "ISO/IEC 14882:2020(E) Programming Language C++." (accessed 2025-03-29).
- LLVM Project. "clang-tidy Documentation." https://clang.llvm.org/extra/clang-tidy/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (C++-specific review)
- alfred-performance-optimizer (C++ profiling)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
