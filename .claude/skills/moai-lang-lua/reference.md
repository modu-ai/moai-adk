# moai-lang-lua - CLI Reference

_Last updated: 2025-10-22_

## Tool Versions (2025-10-22)

| Tool | Version | Official Link |
|------|---------|---------------|
| **Lua** | 5.4.7 | https://www.lua.org/manual/5.4/ |
| **busted** | 2.2.0 | https://olivinelabs.com/busted/ |
| **luacheck** | 1.2.0 | https://github.com/mpeterv/luacheck |
| **luacov** | 0.15.0 | https://keplerproject.github.io/luacov/ |

## Quick Reference

### Project Initialization

```bash
# Install Lua via package manager
# macOS
brew install lua luarocks

# Ubuntu/Debian
sudo apt-get install lua5.4 luarocks

# Windows (LuaRocks)
# Download from https://github.com/luarocks/luarocks/releases

# Install testing and quality tools
luarocks install busted
luarocks install luacheck
luarocks install luacov
```

### Common Commands

```bash
# Run tests
busted

# Run tests with verbose output
busted --verbose

# Run tests with coverage
busted --coverage

# Generate coverage report
luacov
cat luacov.report.out

# Run linter
luacheck src/

# Check specific file
luacheck src/module.lua

# Run with strict mode
luacheck src/ --std lua54

# List all issues
luacheck src/ --formatter plain
```

### Testing Commands

```bash
# Run all tests
busted

# Run specific test file
busted spec/user_service_spec.lua

# Run tests matching pattern
busted --filter="UserService"

# Run with tags
busted --tags=unit

# Exclude tags
busted --exclude-tags=slow

# Run with coverage
busted --coverage

# Watch mode (requires entr or similar)
find spec/ -name '*_spec.lua' | entr busted
```

### Code Quality Commands

```bash
# Run luacheck
luacheck src/

# Check with specific standard
luacheck src/ --std lua54

# Ignore specific warnings
luacheck src/ --ignore 212

# Format output
luacheck src/ --formatter plain

# Generate report
luacheck src/ --formatter JUnit > luacheck-report.xml
```

---

## Configuration Files

### .busted Configuration

```lua
return {
  default = {
    verbose = true,
    coverage = true,
    lpath = 'src/?.lua;src/?/init.lua',
    cpath = '',
    pattern = '_spec',
    ROOT = {'.'},
    recursive = true,
    tags = {},
    output = 'utfTerminal',
    randomize = false,
    seed = os.time(),
    defer_print = false
  }
}
```

### .luacheckrc Configuration

```lua
std = "lua54"
ignore = {
  "212", -- unused argument
  "213"  -- unused loop variable
}
exclude_files = {
  "spec/**/*_spec.lua",
  ".luarocks/**"
}
globals = {
  "describe",
  "it",
  "before_each",
  "after_each",
  "assert"
}
max_line_length = 120
```

### .luacov Configuration

```lua
return {
  statsfile = "luacov.stats.out",
  reportfile = "luacov.report.out",
  
  include = {
    "src"
  },
  
  exclude = {
    "spec",
    "test"
  },
  
  tick = false,
  savestepsize = 100
}
```

---

## Testing Patterns

### Basic Test Structure

```lua
describe("Module", function()
  local Module = require("module")

  before_each(function()
    -- Setup before each test
  end)

  after_each(function()
    -- Cleanup after each test
  end)

  describe("functionName", function()
    it("should perform expected behavior", function()
      local result = Module.functionName()
      assert.equals("expected", result)
    end)
  end)
end)
```

### Assertions

```lua
-- Equality
assert.equals(expected, actual)
assert.same({1, 2}, {1, 2})  -- deep equality

-- Type checks
assert.is_true(value)
assert.is_false(value)
assert.is_nil(value)
assert.is_not_nil(value)

-- Strings
assert.matches("pattern", "string")

-- Numbers
assert.is_near(3.14, 3.1415, 0.01)

-- Errors
assert.has_error(function() error("fail") end)
assert.has_error(function() error("fail") end, "fail")

-- Tables
assert.is_table(value)
assert.is_empty({})
```

### Mocking with luassert

```lua
local spy = require("luassert.spy")
local stub = require("luassert.stub")

describe("Mocking", function()
  it("should spy on function calls", function()
    local callback = spy.new(function() end)
    
    callback("arg1", "arg2")
    
    assert.spy(callback).was.called()
    assert.spy(callback).was.called_with("arg1", "arg2")
  end)

  it("should stub function behavior", function()
    local obj = {method = function() return "original" end}
    stub(obj, "method")
    obj.method.returns("stubbed")
    
    local result = obj.method()
    
    assert.equals("stubbed", result)
    assert.stub(obj.method).was.called()
  end)
end)
```

---

## Lua Best Practices

### Module Pattern

```lua
-- src/module.lua
local Module = {}
Module.__index = Module

-- Constructor
function Module.new(config)
  local self = setmetatable({}, Module)
  self.config = config or {}
  return self
end

-- Instance method
function Module:doSomething()
  return "result"
end

-- Module function
function Module.utilityFunction()
  return "utility"
end

return Module
```

### Error Handling

```lua
-- Using pcall (protected call)
local success, result = pcall(function()
  return risky_operation()
end)

if success then
  print("Result:", result)
else
  print("Error:", result)
end

-- Using assert
local function divide(a, b)
  assert(b ~= 0, "Division by zero")
  return a / b
end

-- Custom error handling
local function safeOperation()
  if not valid_state() then
    error("Invalid state", 2)  -- level 2 = caller's context
  end
end
```

### Metatables and Inheritance

```lua
-- Base class
local Base = {}
Base.__index = Base

function Base.new()
  return setmetatable({}, Base)
end

function Base:baseMethod()
  return "base"
end

-- Derived class
local Derived = setmetatable({}, {__index = Base})
Derived.__index = Derived

function Derived.new()
  local self = setmetatable(Base.new(), Derived)
  return self
end

function Derived:derivedMethod()
  return "derived"
end

-- Usage
local obj = Derived.new()
print(obj:baseMethod())     -- "base"
print(obj:derivedMethod())  -- "derived"
```

---

## TRUST 5 Checklist

### T - Test Coverage (≥85%)

```bash
busted --coverage
luacov
# Check: luacov.report.out
```

### R - Readable Code

```bash
luacheck src/
# Manual review for code style
```

### U - Unified Types

- Use `assert()` for runtime type validation
- Document expected types in comments
- Use metatables for structured data

### S - Security

```bash
# Manual security review
# Check for:
# - Unsafe string operations (string.format injection)
# - Unvalidated user input
# - Unsafe require() or loadstring() calls
# - File system access without validation
```

### T - Trackable with @TAG

```bash
rg '@(CODE|TEST|SPEC):' -n src/ spec/ --type lua
```

---

## Official Resources

- **Lua Manual**: https://www.lua.org/manual/5.4/
- **Lua Style Guide**: https://github.com/luarocks/lua-style-guide
- **busted Documentation**: https://olivinelabs.com/busted/
- **luacheck Documentation**: https://luacheck.readthedocs.io/
- **LuaRocks**: https://luarocks.org/
- **Lua Users Wiki**: http://lua-users.org/wiki/

---

_For working examples, see [examples.md](examples.md)_
