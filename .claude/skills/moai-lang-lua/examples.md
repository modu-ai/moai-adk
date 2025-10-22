# moai-lang-lua - Working Examples

_Last updated: 2025-10-22_

## Example 1: Project Setup with LuaRocks & busted

```bash
# Install Lua (via package manager)
# macOS
brew install lua luarocks

# Ubuntu/Debian
sudo apt-get install lua5.4 luarocks

# Install busted (testing framework)
luarocks install busted

# Install luacheck (linter)
luarocks install luacheck

# Install luacov (coverage)
luarocks install luacov

# Project structure
mkdir -p my-project/src my-project/spec
cd my-project

# Create .busted configuration
cat > .busted << 'EOF'
return {
  default = {
    verbose = true,
    coverage = true,
    lpath = 'src/?.lua;src/?/init.lua',
    pattern = '_spec'
  }
}
EOF

# Create .luacheckrc configuration
cat > .luacheckrc << 'EOF'
std = "lua54"
ignore = {"212"} -- unused arguments
exclude_files = {"spec/**/*_spec.lua"}
EOF
```

**Project structure**:
```
my-project/
├── .busted
├── .luacheckrc
├── src/
│   └── calculator.lua
├── spec/
│   └── calculator_spec.lua
└── luacov.stats.out (generated after tests)
```

## Example 2: TDD Workflow with busted

**RED: Write failing test**
```lua
-- spec/user_service_spec.lua
describe("UserService", function()
  -- @TEST:USER-001

  local UserService = require("user_service")

  describe("createUser", function()
    it("should create user with valid email", function()
      local service = UserService.new()
      local user = service:createUser("test@example.com")

      assert.is_not_nil(user)
      assert.equals("test@example.com", user.email)
      assert.is_not_nil(user.id)
    end)

    it("should reject invalid email", function()
      local service = UserService.new()

      assert.has_error(function()
        service:createUser("invalid-email")
      end, "Invalid email format")
    end)

    it("should reject duplicate email", function()
      local service = UserService.new()

      service:createUser("test@example.com")

      assert.has_error(function()
        service:createUser("test@example.com")
      end, "User already exists")
    end)
  end)

  describe("findUserByEmail", function()
    it("should return nil for non-existent user", function()
      local service = UserService.new()
      local user = service:findUserByEmail("nonexistent@example.com")

      assert.is_nil(user)
    end)

    it("should return existing user", function()
      local service = UserService.new()
      service:createUser("test@example.com")

      local user = service:findUserByEmail("test@example.com")

      assert.is_not_nil(user)
      assert.equals("test@example.com", user.email)
    end)
  end)
end)
```

**GREEN: Implement feature**
```lua
-- src/user_service.lua
-- @CODE:USER-001 | SPEC: SPEC-USER-001.md | TEST: spec/user_service_spec.lua

local UserService = {}
UserService.__index = UserService

function UserService.new()
  local self = setmetatable({}, UserService)
  self.users = {}
  return self
end

function UserService:createUser(email)
  if not self:isValidEmail(email) then
    error("Invalid email format")
  end

  if self:findUserByEmail(email) then
    error("User already exists")
  end

  local user = {
    id = self:generateId(),
    email = email
  }

  table.insert(self.users, user)
  return user
end

function UserService:findUserByEmail(email)
  for _, user in ipairs(self.users) do
    if user.email == email then
      return user
    end
  end
  return nil
end

function UserService:isValidEmail(email)
  return email:match("^[%w+%.%-_]+@[%w+%.%-_]+%.%a%a+$") ~= nil
end

function UserService:generateId()
  return string.format("%s-%d", "user", os.time())
end

return UserService
```

**REFACTOR: Improve with better error handling and validation**
```lua
-- src/user_service.lua
-- @CODE:USER-001 | SPEC: SPEC-USER-001.md | TEST: spec/user_service_spec.lua

local UserService = {}
UserService.__index = UserService

-- Constants
local EMAIL_PATTERN = "^[%w+%.%-_]+@[%w+%.%-_]+%.%a%a+$"

function UserService.new()
  local self = setmetatable({}, UserService)
  self.users = {}
  self.emailIndex = {}
  return self
end

function UserService:createUser(email)
  -- Validate email format
  assert(type(email) == "string", "Email must be a string")
  assert(self:isValidEmail(email), "Invalid email format")

  -- Check for duplicates using index
  assert(not self.emailIndex[email], "User already exists")

  -- Create user
  local user = {
    id = self:generateId(),
    email = email,
    createdAt = os.time()
  }

  -- Store user with index
  table.insert(self.users, user)
  self.emailIndex[email] = user

  return user
end

function UserService:findUserByEmail(email)
  return self.emailIndex[email]
end

function UserService:isValidEmail(email)
  return email:match(EMAIL_PATTERN) ~= nil
end

function UserService:generateId()
  return string.format("user-%s", os.time() * 1000 + math.random(1000))
end

function UserService:getUserCount()
  return #self.users
end

return UserService
```

## Example 3: Metatables & OOP Patterns

```lua
-- spec/shape_spec.lua
describe("Shape OOP", function()
  -- @TEST:SHAPE-001

  local Shape = require("shape")
  local Circle = require("circle")
  local Rectangle = require("rectangle")

  describe("Circle", function()
    it("should calculate area correctly", function()
      local circle = Circle.new(5)
      local area = circle:area()

      assert.is_near(78.54, area, 0.01)
    end)

    it("should inherit from Shape", function()
      local circle = Circle.new(5)
      local description = circle:describe()

      assert.equals("This is a circle", description)
    end)
  end)

  describe("Rectangle", function()
    it("should calculate area correctly", function()
      local rect = Rectangle.new(4, 5)
      local area = rect:area()

      assert.equals(20, area)
    end)

    it("should calculate perimeter", function()
      local rect = Rectangle.new(4, 5)
      local perimeter = rect:perimeter()

      assert.equals(18, perimeter)
    end)
  end)
end)

-- src/shape.lua
-- @CODE:SHAPE-001 | SPEC: SPEC-SHAPE-001.md | TEST: spec/shape_spec.lua

local Shape = {}
Shape.__index = Shape

function Shape.new()
  local self = setmetatable({}, Shape)
  return self
end

function Shape:describe()
  return "This is a shape"
end

return Shape

-- src/circle.lua
local Shape = require("shape")

local Circle = setmetatable({}, {__index = Shape})
Circle.__index = Circle

function Circle.new(radius)
  local self = setmetatable(Shape.new(), Circle)
  self.radius = radius
  return self
end

function Circle:area()
  return math.pi * self.radius * self.radius
end

function Circle:describe()
  return "This is a circle"
end

return Circle

-- src/rectangle.lua
local Shape = require("shape")

local Rectangle = setmetatable({}, {__index = Shape})
Rectangle.__index = Rectangle

function Rectangle.new(width, height)
  local self = setmetatable(Shape.new(), Rectangle)
  self.width = width
  self.height = height
  return self
end

function Rectangle:area()
  return self.width * self.height
end

function Rectangle:perimeter()
  return 2 * (self.width + self.height)
end

function Rectangle:describe()
  return "This is a rectangle"
end

return Rectangle
```

## Example 4: Quality Gate Check

```bash
# Run all tests
busted

# Run tests with coverage
busted --coverage

# Generate coverage report
luacov
cat luacov.report.out

# Run luacheck (linter)
luacheck src/

# Fix common issues
luacheck src/ --formatter plain

# TRUST 5 validation
echo "T - Test coverage"
busted --coverage
luacov
# Check luacov.report.out for ≥85% coverage

echo "R - Readable code"
luacheck src/ --no-color

echo "U - Unified types"
# Lua is dynamically typed, use assert() for runtime checks

echo "S - Security scan"
# Manual code review for security issues

echo "T - Trackable with @TAG"
rg '@(CODE|TEST|SPEC):' -n src/ spec/ --type lua
```

**busted configuration** (`.busted`):
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
    defer_print = false,
    Xoutput = nil,
    Xhelper = {},
    ['coverage-config-file'] = '.luacov'
  }
}
```

**luacov configuration** (`.luacov`):
```lua
return {
  statsfile = "luacov.stats.out",
  reportfile = "luacov.report.out",
  include = {"src"},
  exclude = {"spec"},
  tick = false,
  savestepsize = 100
}
```

## Example 5: Coroutines & Async Patterns

```lua
-- spec/async_spec.lua
describe("Async Operations", function()
  -- @TEST:ASYNC-001

  local async = require("async")

  describe("fetchData", function()
    it("should fetch data asynchronously", function()
      local co = coroutine.create(function()
        local data = async.fetchData("user123")
        assert.equals("Data for user123", data)
      end)

      coroutine.resume(co)
    end)

    it("should handle multiple concurrent fetches", function()
      local results = {}

      local co1 = coroutine.create(function()
        results[1] = async.fetchData("user1")
      end)

      local co2 = coroutine.create(function()
        results[2] = async.fetchData("user2")
      end)

      coroutine.resume(co1)
      coroutine.resume(co2)

      assert.equals(2, #results)
    end)
  end)
end)

-- src/async.lua
-- @CODE:ASYNC-001 | SPEC: SPEC-ASYNC-001.md | TEST: spec/async_spec.lua

local async = {}

function async.fetchData(id)
  -- Simulate async operation
  coroutine.yield()

  if not id then
    error("ID is required")
  end

  return "Data for " .. id
end

function async.runConcurrent(tasks)
  local coroutines = {}
  local results = {}

  -- Create coroutines for each task
  for i, task in ipairs(tasks) do
    coroutines[i] = coroutine.create(function()
      results[i] = task()
    end)
  end

  -- Resume all coroutines
  local allDone = false
  while not allDone do
    allDone = true
    for _, co in ipairs(coroutines) do
      if coroutine.status(co) ~= "dead" then
        coroutine.resume(co)
        allDone = false
      end
    end
  end

  return results
end

return async
```

---

## TRUST 5 Integration

### Test Coverage (≥85%)
```bash
busted --coverage
luacov
# Check: luacov.report.out
```

### Readable Code
```bash
luacheck src/
# Auto-fix (manual)
```

### Unified Types
- Use `assert()` for runtime type checks
- Document expected types in comments
- Use metatables for structured data

### Security
```bash
# Manual security review
# Check for:
# - Unsafe string operations
# - Unvalidated user input
# - Unsafe require() calls
```

### Trackable with @TAG
```bash
rg '@(CODE|TEST|SPEC):' -n src/ spec/ --type lua
```

---

_For detailed CLI reference, see [reference.md](reference.md)_
