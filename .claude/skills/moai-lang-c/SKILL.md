---
name: moai-lang-c
description: C best practices with Unity test framework, cppcheck, and Make build
  system
allowed-tools:
- Read
- Bash
---

# C Expert

## What it does

Provides C-specific expertise for TDD development, including Unity test framework, cppcheck static analysis, and Make build system for system programming.

## When to use

- "C 테스트 작성", "Unity 테스트 프레임워크", "임베디드 C"
- Automatically invoked when working with C projects
- C SPEC implementation (`/alfred:2-build`)

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

### Example 1: TDD with Unity
User: "/alfred:2-build DRIVER-001"
Claude: (creates RED test with Unity, GREEN implementation, REFACTOR with error handling)

### Example 2: Memory leak check
User: "Valgrind 메모리 체크"
Claude: (runs valgrind --leak-check=full and reports leaks)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (C-specific review)
- alfred-debugger-pro (C debugging)
