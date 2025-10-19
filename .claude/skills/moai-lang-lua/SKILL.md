---
name: moai-lang-lua
description: Lua best practices with busted, luacheck, and embedded scripting patterns
allowed-tools:
  - Read
  - Bash
tier: 3
auto-load: "true"
---

# Lua Expert

## What it does

Provides Lua-specific expertise for TDD development, including busted testing framework, luacheck linting, and embedded scripting patterns for game development and system configuration.

## When to use

- "Lua 테스트 작성", "busted 사용법", "임베디드 스크립팅", "게임 개발", "Neovim 플러그인", "시스템 설정"
- "Love2D", "Corona SDK", "Roblox", "Redis Lua", "Nginx Lua"
- "게임 로직", "설정 파일", "확장 스크립팅"
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

## Modern Lua (5.4+)

**Recommended Version**: Lua 5.4+ for production, Lua 5.1+ for LuaJIT compatibility

**Modern Features**:
- **Const variables** (5.4+): `local x <const> = 10`
- **To-be-closed variables** (5.4+): Automatic resource cleanup
- **Bitwise operators** (5.3+): `&`, `|`, `~`, `<<`, `>>`
- **Integer subtype** (5.3+): Separate int and float
- **UTF-8 library** (5.3+): Unicode support
- **table.move** (5.3+): Efficient table operations

**Version Check**:
```bash
lua -v
luajit -v
```

## Package Management Commands

### Using LuaRocks
```bash
# Install LuaRocks
# macOS: brew install luarocks
# Ubuntu: sudo apt install luarocks

# Install packages
luarocks install busted
luarocks install luacheck
luarocks install penlight

# Install specific version
luarocks install lapis 1.7.0

# Search packages
luarocks search http
luarocks show busted

# List installed
luarocks list

# Remove package
luarocks remove busted

# Update
luarocks install busted --force
```

**rockspec Example**:
```lua
-- mylib-1.0-1.rockspec
package = "mylib"
version = "1.0-1"
source = {
   url = "git://github.com/user/mylib.git",
   tag = "v1.0"
}
description = {
   summary = "My Lua library",
   detailed = "A detailed description",
   homepage = "https://github.com/user/mylib",
   license = "MIT"
}
dependencies = {
   "lua >= 5.1, < 5.5",
   "penlight >= 1.5"
}
build = {
   type = "builtin",
   modules = {
      mylib = "src/mylib.lua",
      ["mylib.utils"] = "src/utils.lua"
   }
}
```

### Common Development Commands
```bash
# Run Lua script
lua script.lua
luajit script.lua  # LuaJIT (faster)

# Interactive REPL
lua
luajit

# Run tests with busted
busted
busted spec/  # Specific directory
busted -v  # Verbose output
busted --coverage  # Coverage report

# Lint with luacheck
luacheck .
luacheck src/ --globals love  # Game frameworks
luacheck --std=ngx_lua nginx.lua  # Nginx Lua

# Format with StyLua
stylua .
stylua --check .
stylua --config-path .stylua.toml .

# Create package
luarocks pack mylib-1.0-1.rockspec
luarocks upload mylib-1.0-1.rockspec
```

### Project Structure
```
my-lua-project/
├── rockspec/
│   └── mylib-1.0-1.rockspec
├── src/
│   ├── mylib.lua
│   └── utils.lua
├── spec/
│   ├── mylib_spec.lua
│   └── utils_spec.lua
├── .luacheckrc
├── .stylua.toml
└── README.md
```

## Examples

### Example 1: TDD with busted
User: "/alfred:2-run CONFIG-001"
Claude: (creates RED test with busted, GREEN implementation, REFACTOR with metatables)

### Example 2: Linting check
User: "luacheck 실행"
Claude: (runs luacheck and reports style violations)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Lua-specific review)
- cli-tool-expert (Lua scripting utilities)
