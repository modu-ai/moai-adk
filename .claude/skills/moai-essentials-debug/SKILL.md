---
name: moai-essentials-debug
description: Advanced debugging with stack trace analysis, error pattern detection, and fix suggestions
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
tier: 1
auto-load: "true"
---

# Alfred Debugger Pro

## What it does

Advanced debugging support with stack trace analysis, common error pattern detection, and actionable fix suggestions.

## When to use

- "에러 해결해줘", "이 오류 원인은?", "스택 트레이스 분석", "버그 찾아줘", "디버깅 도와줘"
- "왜 안 돼?", "NullPointerException 해결", "런타임 에러", "예외 처리"
- "Error analysis", "Exception handling", "Bug fixing", "Root cause analysis"
- Automatically invoked on runtime errors (via debug-helper sub-agent)
- When tests fail or unexpected behavior occurs

## How it works

**Stack Trace Analysis**:
```python
# Error example
jwt.exceptions.ExpiredSignatureError: Signature has expired

# Alfred Analysis
📍 Error Location: src/auth/service.py:142
🔍 Root Cause: JWT token has expired
💡 Fix Suggestion:
   1. Implement token refresh logic
   2. Check expiration before validation
   3. Handle ExpiredSignatureError gracefully
```

**Common Error Patterns**:
- `NullPointerException` → Optional usage, guard clauses
- `IndexError` → Boundary checks
- `KeyError` → `.get()` with defaults
- `TypeError` → Type hints, input validation
- `ConnectionError` → Retry logic, timeouts

**Debugging Checklist**:
- [ ] Reproducible?
- [ ] Log messages?
- [ ] Input data?
- [ ] Recent changes?
- [ ] Dependency versions?

**Language-specific Tips**:
- **Python**: Logging, type guards, pdb debugger
- **TypeScript**: Type guards, null checks, debugger statement
- **Java**: Optional, try-with-resources, logging
- **Go**: panic/recover, error wrapping
- **Rust**: Result/Option types, ? operator

## Debugging Commands

### Python
```bash
# Run with debugger
python -m pdb script.py

# Enable verbose logging
export PYTHONVERBOSE=1

# Check stack trace
python -m traceback script.py

# Memory profiling
python -m memory_profiler script.py
```

### TypeScript/JavaScript
```bash
# Run with inspector
node --inspect script.js

# Enable source maps
tsc --sourceMap

# Check with debugger
node --inspect-brk script.js
```

### Go
```bash
# Run with race detector
go run -race main.go

# Debug with delve
dlv debug main.go
```

### Rust
```bash
# Run with backtrace
RUST_BACKTRACE=1 cargo run

# Full backtrace
RUST_BACKTRACE=full cargo run
```

## Common Error Patterns & Fixes

### NullPointerException / AttributeError
```python
# ❌ Bad
user = get_user(id)
print(user.name)  # Crash if user is None

# ✅ Good
user = get_user(id)
if user:
    print(user.name)
# Or
print(user.name if user else "Unknown")
```

### IndexError / ArrayIndexOutOfBoundsException
```python
# ❌ Bad
items[5]  # Crash if len(items) < 6

# ✅ Good
if len(items) > 5:
    print(items[5])
# Or
print(items[5] if len(items) > 5 else None)
```

### KeyError / Property not found
```python
# ❌ Bad
data["key"]  # Crash if key doesn't exist

# ✅ Good
data.get("key", default_value)
```

### TypeError / Type mismatch
```python
# ❌ Bad
def add(a, b):
    return a + b
add("5", 3)  # TypeError

# ✅ Good
def add(a: int, b: int) -> int:
    if not isinstance(a, int) or not isinstance(b, int):
        raise TypeError("Both arguments must be integers")
    return a + b
```

### ConnectionError / TimeoutError
```python
# ❌ Bad
response = requests.get(url)

# ✅ Good
try:
    response = requests.get(url, timeout=5)
except requests.exceptions.Timeout:
    logger.error("Request timed out")
    raise
except requests.exceptions.ConnectionError:
    logger.error("Connection failed")
    raise
```

## Debugging Workflow

**1. Reproduce the Error**:
```bash
# Capture exact input
echo '{"user_id": 123}' | python script.py

# Save error output
python script.py 2> error.log
```

**2. Analyze Stack Trace**:
```bash
# Python: Last line = root cause
Traceback (most recent call last):
  File "service.py", line 142, in authenticate
    jwt.decode(token, secret)
jwt.exceptions.ExpiredSignatureError: Signature has expired
```

**3. Add Logging**:
```python
import logging
logging.basicConfig(level=logging.DEBUG)

def authenticate(token):
    logger.debug(f"Authenticating token: {token[:10]}...")
    # ... rest of code
```

**4. Hypothesis Testing**:
```python
# Test assumption
assert token is not None, "Token is None!"
assert isinstance(token, str), f"Token is {type(token)}, not str!"
```

**5. Fix & Verify**:
```bash
# Run tests
pytest tests/test_auth.py -v

# Verify fix in production-like env
docker-compose up -d
curl -X POST http://localhost:8000/auth
```

## Examples

### Example 1: JWT ExpiredSignatureError
User: "JWT ExpiredSignatureError 해결해줘"

Alfred analyzes:
```
📍 Error Location: src/auth/service.py:142
🔍 Root Cause: JWT token expired (issued 35 minutes ago, TTL=30min)
💡 Fix Suggestions:
  1. Implement token refresh endpoint
  2. Add expiration check before validation:
     if token_age > MAX_AGE:
         raise TokenExpired()
  3. Handle gracefully with 401 response
```

### Example 2: NullPointerException in Java
User: "NullPointerException 해결"

Alfred provides:
```java
// Before (crashes)
User user = userService.findById(id);
return user.getName();

// After (safe)
User user = userService.findById(id);
return Optional.ofNullable(user)
    .map(User::getName)
    .orElse("Unknown");
```

### Example 3: N+1 Query Debugging
User: "왜 느려?"

Alfred detects:
```sql
-- N+1 Query Problem
SELECT * FROM users;  -- 1 query
SELECT * FROM posts WHERE user_id = 1;  -- N queries
SELECT * FROM posts WHERE user_id = 2;
...

-- Fix: Eager Loading
SELECT users.*, posts.*
FROM users
LEFT JOIN posts ON users.id = posts.user_id;
```

### Example 4: Memory Leak Detection
User: "메모리 계속 증가해"

Alfred profiles:
```bash
# Python memory profiling
python -m memory_profiler script.py

Result:
Line 45: +500MB (event listeners not removed)
Line 78: +200MB (cache never cleared)

Fix:
1. Remove event listeners on cleanup
2. Implement LRU cache with max size
```

### Example 5: Race Condition
User: "가끔 틀린 결과 나와"

Alfred identifies:
```go
// Race condition detected
var counter int
go func() { counter++ }()
go func() { counter++ }()
fmt.Println(counter) // Sometimes 0, 1, or 2!

// Fix: Use mutex
var mu sync.Mutex
var counter int
go func() { mu.Lock(); counter++; mu.Unlock() }()
```

## Works well with

- moai-essentials-refactor (Clean up after fixing bugs)
- moai-foundation-trust (Prevent bugs with TRUST principles)
