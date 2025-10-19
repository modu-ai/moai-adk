---
name: moai-lang-lua
description: Lua best practices with busted, luacheck, and embedded scripting patterns
allowed-tools:
- Read
- Bash
---

# Lua Expert

## What it does

Provides Lua-specific expertise for TDD development, including busted testing framework, luacheck linting, and embedded scripting patterns for game development and system configuration.

## When to use

- "Lua 테스트 작성", "busted 사용법", "임베디드 스크립팅"
- Automatically invoked when working with Lua projects
- Lua SPEC implementation (`/alfred:2-build`)

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

### Example 1: TDD with busted
User: "/alfred:2-build CONFIG-001"
Claude: (creates RED test with busted, GREEN implementation, REFACTOR with metatables)

### Example 2: Linting check
User: "luacheck 실행"
Claude: (runs luacheck and reports style violations)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Lua-specific review)
- cli-tool-expert (Lua scripting utilities)
