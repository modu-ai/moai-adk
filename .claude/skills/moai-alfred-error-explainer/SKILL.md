---
name: moai-alfred-error-explainer
description: Runtime error analysis with stack trace parsing, SPEC-based root cause detection, and actionable fix suggestions (user)
allowed-tools:
- Read
- Bash
- Grep
- Glob
- TodoWrite
---

# Alfred Error Explainer

## What it does

Comprehensive runtime error analysis with automatic stack trace parsing, SPEC-based root cause detection, and language-specific fix suggestions. **Auto-invoked by Alfred** when runtime errors are detected during execution.

## When to use

**Auto-invoked by Alfred**:
- When runtime errors occur during test execution
- When build/compilation fails with errors
- When `debug-helper` agent is triggered

**Manually invoked by users**:
- "이 에러 해결해줘"
- "TypeError 원인 분석"
- "스택 트레이스 설명"
- "왜 안 돼?"

## How it works

### Phase 1: Error Context Collection

**Stack Trace Parsing**:
```python
# Example Python traceback
Traceback (most recent call last):
  File "src/auth/service.py", line 142, in authenticate_user
    token = jwt.encode(payload, self.secret)
  File "/site-packages/jwt/__init__.py", line 67, in encode
    raise ExpiredSignatureError("Signature has expired")
jwt.exceptions.ExpiredSignatureError: Signature has expired

# Alfred parses:
📍 Error Location: src/auth/service.py:142
🔍 Error Type: jwt.exceptions.ExpiredSignatureError
📝 Error Message: "Signature has expired"
🎯 Function: authenticate_user
```

**TAG Chain Tracing**:
```bash
# Alfred automatically traces:
@CODE:AUTH-001 (src/auth/service.py:142)
  ↓
@TEST:AUTH-001 (tests/auth/test_service.py:45)
  ↓
@SPEC:AUTH-001 (.moai/specs/SPEC-AUTH-001/spec.md)
```

**Recent Changes**:
```bash
# Git blame to identify recent changes
$ git log -5 --oneline -- src/auth/service.py

71269d9 ♻️ REFACTOR: JWT token expiration handling
e324aac 📝 DOCS: Add token refresh documentation
93897a5 🎯 FEAT: Implement JWT authentication
```

---

### Phase 2: SPEC-Based Root Cause Analysis

**Step 1: Load Related SPEC**:
```yaml
# .moai/specs/SPEC-AUTH-001/spec.md
---
id: AUTH-001
version: 0.1.0
status: completed
---

## Event-driven Requirements
- WHEN 토큰이 만료되면, 시스템은 401 에러를 반환해야 한다
- WHEN 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급할 수 있다

## Constraints
- 액세스 토큰 만료시간은 15분을 초과하지 않아야 한다
```

**Step 2: Compare SPEC vs Implementation**:
```python
# Alfred Analysis:
⚠️ SPEC 요구사항: "WHEN 토큰이 만료되면, 401 에러를 반환"
❌ 현재 구현: ExpiredSignatureError 발생 (처리되지 않음)

💡 Gap Detected:
   - SPEC은 401 에러 반환을 요구
   - 구현은 예외를 catch하지 않음
   - 리프레시 토큰 로직 누락
```

---

### Phase 3: Common Error Pattern Matching

**Language-specific Error Patterns Database**:

**Python Errors**:
```python
# 1. AttributeError
error = "AttributeError: 'NoneType' object has no attribute 'name'"

# Alfred detects:
🔍 Pattern: None object access
💡 Fix: Add null check before access
✅ Solution:
   if user is not None:
       name = user.name
   # or
   name = user.name if user else None
   # or
   from typing import Optional
   name: Optional[str] = getattr(user, 'name', None)
```

```python
# 2. KeyError
error = "KeyError: 'user_id'"

# Alfred detects:
🔍 Pattern: Missing dictionary key
💡 Fix: Use .get() with default value
✅ Solution:
   user_id = data.get('user_id', None)
   # or with validation
   if 'user_id' not in data:
       raise ValueError("Missing required field: user_id")
```

```python
# 3. IndexError
error = "IndexError: list index out of range"

# Alfred detects:
🔍 Pattern: Out-of-bounds list access
💡 Fix: Add boundary check
✅ Solution:
   if 0 <= index < len(items):
       item = items[index]
   # or with try/except
   try:
       item = items[index]
   except IndexError:
       item = None
```

```python
# 4. TypeError
error = "TypeError: unsupported operand type(s) for +: 'int' and 'str'"

# Alfred detects:
🔍 Pattern: Type mismatch in operation
💡 Fix: Add type conversion or validation
✅ Solution:
   result = int(a) + int(b)
   # or with type hints
   def add(a: int, b: int) -> int:
       return a + b
```

---

**TypeScript Errors**:
```typescript
// 1. Cannot read property 'X' of undefined
error = "TypeError: Cannot read property 'name' of undefined"

// Alfred detects:
🔍 Pattern: Undefined object access
💡 Fix: Optional chaining
✅ Solution:
   const name = user?.name;
   // or null check
   if (user !== undefined && user !== null) {
       const name = user.name;
   }
   // or with default
   const name = user?.name ?? 'Unknown';
```

```typescript
// 2. Type 'X' is not assignable to type 'Y'
error = "Type 'string' is not assignable to type 'number'"

// Alfred detects:
🔍 Pattern: Type mismatch
💡 Fix: Type conversion or interface update
✅ Solution:
   const value: number = parseInt(input, 10);
   // or update interface
   interface Config {
       port: string | number;  // Union type
   }
```

```typescript
// 3. Promise rejection
error = "UnhandledPromiseRejectionWarning: Error: Connection timeout"

// Alfred detects:
🔍 Pattern: Unhandled promise rejection
💡 Fix: Add error handling
✅ Solution:
   try {
       await fetchData();
   } catch (error) {
       console.error('Failed to fetch:', error);
       // Handle error gracefully
   }
```

---

**Java Errors**:
```java
// 1. NullPointerException
error = "java.lang.NullPointerException at UserService.java:45"

// Alfred detects:
🔍 Pattern: Null reference access
💡 Fix: Use Optional or null check
✅ Solution:
   Optional<User> user = userRepository.findById(id);
   user.ifPresent(u -> {
       String name = u.getName();
   });
   // or
   if (user != null) {
       String name = user.getName();
   }
```

```java
// 2. ConcurrentModificationException
error = "java.util.ConcurrentModificationException"

// Alfred detects:
🔍 Pattern: Modifying collection during iteration
💡 Fix: Use Iterator or Stream
✅ Solution:
   Iterator<User> it = users.iterator();
   while (it.hasNext()) {
       if (shouldRemove(it.next())) {
           it.remove();
       }
   }
   // or
   users = users.stream()
       .filter(u -> !shouldRemove(u))
       .collect(Collectors.toList());
```

---

### Phase 4: Dependency & Environment Check

**Version Conflicts**:
```bash
# Alfred checks package versions
$ uv pip list | grep jwt
PyJWT==1.7.1  # Installed

# Alfred compares with requirements
$ cat requirements.txt
PyJWT>=2.0.0  # Required

# Alfred detects:
⚠️ Version mismatch: PyJWT 1.7.1 < 2.0.0
💡 Fix: Upgrade PyJWT
✅ Command: uv pip install --upgrade PyJWT
```

**Missing Dependencies**:
```bash
# Error
ModuleNotFoundError: No module named 'jwt'

# Alfred checks:
$ grep "jwt" requirements.txt
(not found)

# Alfred suggests:
💡 Missing dependency: PyJWT
✅ Command: uv pip install PyJWT
✅ Add to requirements.txt: PyJWT==2.8.0
```

---

### Phase 5: Fix Suggestions Generation

**Three-Level Fix Suggestions**:

**Level 1: Immediate Fix (Code)**:
```python
# Current code (src/auth/service.py:142)
def authenticate_user(self, username: str, password: str):
    payload = {'username': username}
    token = jwt.encode(payload, self.secret)  # ← Error here
    return token

# Alfred suggests:
💡 Immediate Fix:
   Add try/except to handle ExpiredSignatureError

✅ Fixed Code:
   def authenticate_user(self, username: str, password: str):
       payload = {'username': username}
       try:
           token = jwt.encode(payload, self.secret)
           return token
       except jwt.ExpiredSignatureError:
           raise AuthenticationError("Token has expired", status_code=401)
```

**Level 2: SPEC Alignment**:
```markdown
# Alfred checks SPEC
⚠️ SPEC Requirement Missing:
   - Refresh token logic not implemented
   - 401 error handling not specified

💡 SPEC Update Needed:
   1. Add refresh token endpoint to SPEC
   2. Specify error handling strategy
   3. Update @TEST:AUTH-001 to test expiration
```

**Level 3: Architecture Improvement**:
```python
# Alfred suggests refactoring
💡 Architecture Improvement:
   - Extract token management to separate TokenManager class
   - Implement TokenRefreshService
   - Add JWT configuration object

✅ Suggested Structure:
   src/auth/
   ├── service.py (authentication logic)
   ├── token_manager.py (JWT encoding/decoding) ← NEW
   ├── token_refresh.py (refresh logic) ← NEW
   └── exceptions.py (custom exceptions)
```

---

## Error Report Format

```markdown
## 🔴 Runtime Error Analysis Report

**Error Type**: jwt.exceptions.ExpiredSignatureError
**Location**: src/auth/service.py:142
**Function**: authenticate_user
**Timestamp**: 2025-10-20 14:32:15

---

### 📍 Stack Trace

```
Traceback (most recent call last):
  File "src/auth/service.py", line 142, in authenticate_user
    token = jwt.encode(payload, self.secret)
  File "/site-packages/jwt/__init__.py", line 67, in encode
    raise ExpiredSignatureError("Signature has expired")
jwt.exceptions.ExpiredSignatureError: Signature has expired
```

---

### 🔍 Root Cause Analysis

**SPEC Reference**: @SPEC:AUTH-001 (v0.1.0)

**SPEC Requirement**:
> WHEN 토큰이 만료되면, 시스템은 401 에러를 반환해야 한다

**Current Implementation**:
- ❌ ExpiredSignatureError 발생 (처리되지 않음)
- ❌ 401 에러 반환 로직 없음
- ❌ 리프레시 토큰 로직 미구현

**Gap Analysis**:
1. SPEC은 401 에러 반환을 요구하지만 구현 누락
2. 에러 처리 로직이 없어 예외가 상위로 전파됨
3. 리프레시 토큰 엔드포인트 미구현

---

### 💡 Fix Suggestions

#### Level 1: Immediate Fix (Code)

**Current Code**:
```python
def authenticate_user(self, username: str, password: str):
    payload = {'username': username}
    token = jwt.encode(payload, self.secret)  # ← Error here
    return token
```

**Fixed Code**:
```python
def authenticate_user(self, username: str, password: str):
    payload = {'username': username}
    try:
        token = jwt.encode(payload, self.secret, algorithm='HS256')
        return {
            'access_token': token,
            'token_type': 'Bearer',
            'expires_in': 900  # 15 minutes
        }
    except jwt.ExpiredSignatureError:
        raise AuthenticationError(
            "Token has expired. Please refresh your token.",
            status_code=401
        )
```

**Changes**:
1. ✅ Add try/except for ExpiredSignatureError
2. ✅ Return 401 status code (via custom exception)
3. ✅ Add expiration info in response

---

#### Level 2: SPEC Alignment

**SPEC Update Needed**:
```markdown
# Add to SPEC-AUTH-001/spec.md

## Event-driven Requirements (Updated)
- WHEN 토큰이 만료되면, 시스템은 401 에러를 반환해야 한다
- WHEN 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급해야 한다 ⭐ NEW

## Optional Features
- WHERE 리프레시 토큰이 유효하면, 시스템은 새 액세스 토큰을 발급할 수 있다 ⭐ NEW

## Constraints
- 리프레시 토큰 만료시간은 7일을 초과하지 않아야 한다 ⭐ NEW
```

**Test Update Needed**:
```python
# tests/auth/test_service.py
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

def test_expired_token_returns_401():
    """WHEN 토큰이 만료되면, 401 에러 반환"""
    service = AuthService()
    expired_token = generate_expired_token()

    with pytest.raises(AuthenticationError) as exc:
        service.validate_token(expired_token)

    assert exc.value.status_code == 401

def test_refresh_token_generates_new_access_token():  # ⭐ NEW
    """WHEN 리프레시 토큰 제공 시, 새 액세스 토큰 발급"""
    service = AuthService()
    refresh_token = "valid_refresh_token"

    result = service.refresh_access_token(refresh_token)

    assert result['access_token'] is not None
    assert result['expires_in'] == 900
```

---

#### Level 3: Architecture Improvement

**Refactoring Suggestion**:
```python
# Separate token management concerns

# src/auth/token_manager.py ⭐ NEW
class TokenManager:
    """Handles JWT encoding/decoding/validation"""

    def __init__(self, secret: str, algorithm: str = 'HS256'):
        self.secret = secret
        self.algorithm = algorithm

    def encode(self, payload: dict, expires_in: int = 900) -> str:
        """Encode payload with expiration"""
        payload['exp'] = datetime.utcnow() + timedelta(seconds=expires_in)
        return jwt.encode(payload, self.secret, algorithm=self.algorithm)

    def decode(self, token: str) -> dict:
        """Decode and validate token"""
        try:
            return jwt.decode(token, self.secret, algorithms=[self.algorithm])
        except jwt.ExpiredSignatureError:
            raise AuthenticationError("Token expired", status_code=401)
        except jwt.InvalidTokenError:
            raise AuthenticationError("Invalid token", status_code=401)

# src/auth/service.py (refactored)
class AuthService:
    def __init__(self, token_manager: TokenManager):
        self.token_manager = token_manager

    def authenticate_user(self, username: str, password: str):
        # Validation logic
        payload = {'username': username, 'sub': user.id}
        access_token = self.token_manager.encode(payload, expires_in=900)
        refresh_token = self.token_manager.encode(payload, expires_in=604800)

        return {
            'access_token': access_token,
            'refresh_token': refresh_token,
            'token_type': 'Bearer'
        }
```

---

### 🎯 Recommended Actions

**Priority 1: Immediate (Fix Production Issue)**
1. ✅ Add try/except to handle ExpiredSignatureError
2. ✅ Return proper 401 error response
3. ✅ Deploy hotfix

**Priority 2: Short-term (SPEC Alignment)**
1. ⚠️ Update SPEC-AUTH-001 with refresh token requirements
2. ⚠️ Add @TEST:AUTH-001 for expiration handling
3. ⚠️ Implement refresh token endpoint

**Priority 3: Long-term (Architecture)**
1. 💡 Extract TokenManager class
2. 💡 Implement TokenRefreshService
3. 💡 Add token rotation strategy

---

### 📚 Related Resources

**Documentation**:
- [PyJWT Documentation](https://pyjwt.readthedocs.io/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)

**Similar Issues**:
- Stack Overflow: ["How to handle JWT expiration in Python"](https://stackoverflow.com/q/...)

**Related SPEC**:
- @SPEC:AUTH-001 (JWT authentication)
- @SPEC:AUTH-002 (Token refresh)

---

**Alfred Recommendation**: 🟡 Immediate Fix Required + SPEC Update
```

---

## Integration with debug-helper Agent

**Workflow**:
```bash
# Runtime error occurs
pytest tests/auth/test_service.py::test_authenticate

# Error detected
FAILED tests/auth/test_service.py::test_authenticate
ExpiredSignatureError: Signature has expired

# debug-helper agent is triggered
Alfred: "오류가 발생했습니다. 분석 중..."

# Alfred invokes moai-alfred-error-explainer (AUTO) ⭐
⚙️ Alfred Error Explainer: Analyzing error...

# Phase 1: Context Collection
📍 Location: src/auth/service.py:142
🔍 Error: ExpiredSignatureError
🎯 Function: authenticate_user

# Phase 2: SPEC Analysis
📋 SPEC: @SPEC:AUTH-001
⚠️ Gap: 401 error handling missing

# Phase 3: Fix Suggestions
💡 Immediate Fix: Add try/except
💡 SPEC Update: Add refresh token requirements
💡 Architecture: Extract TokenManager

# Report Generated
📄 Report: .moai/reports/error-analysis-2025-10-20.md

Alfred: "에러 분석이 완료되었습니다. 3가지 해결 방법을 제안드립니다."
```

---

## Examples

### Example 1: Auto-invoked on runtime error

```bash
# User runs tests
$ pytest tests/auth/test_service.py

# Error occurs
FAILED - ExpiredSignatureError: Signature has expired

# Alfred automatically analyzes
⚙️ Alfred Error Explainer: Runtime error detected

📍 Error Location: src/auth/service.py:142
🔍 Root Cause: JWT token expired, no error handling
💡 SPEC Gap: 401 error requirement not implemented

✅ Fix Suggestions:
   1. Add try/except for ExpiredSignatureError
   2. Return 401 status code
   3. Update SPEC with refresh token logic

📋 Full Report: .moai/reports/error-analysis-2025-10-20.md
```

### Example 2: Manual invocation

```bash
# User pastes error
User: "이 에러 해결해줘"
User: (pastes stack trace)

Alfred: "에러를 분석하겠습니다."

# Alfred uses moai-alfred-error-explainer Skill
⚙️ Parsing stack trace...
📍 Error: NullPointerException at UserService.java:45

💡 Common Pattern: Null reference access
✅ Fix: Use Optional<User>

🔍 SPEC Check: @SPEC:USER-001
⚠️ SPEC doesn't specify null handling strategy

📋 Detailed analysis: .moai/reports/error-analysis-2025-10-20.md
```

---

## Works well with

**Skills**:
- moai-essentials-debug (manual debugging)
- moai-foundation-specs (SPEC compliance check)
- moai-foundation-tags (TAG chain tracing)

**Agents**:
- debug-helper (triggers auto-invocation)
- spec-builder (SPEC update)
- tdd-implementer (test update)

**Commands**:
- `/alfred:2-build` (TDD cycle)
- `/alfred:3-sync` (verify fixes)

---

## Configuration

**.moai/config.json**:
```json
{
  "error-explainer": {
    "auto-invoke": true,
    "trace-tag-chain": true,
    "check-spec-compliance": true,
    "suggest-levels": ["immediate", "spec", "architecture"],
    "language-patterns": {
      "python": true,
      "typescript": true,
      "java": true
    }
  }
}
```

---

## Differences from moai-essentials-debug

| Feature | moai-alfred-error-explainer | moai-essentials-debug |
|---------|----------------------------|----------------------|
| **Auto-invoked** | ✅ Yes (on runtime errors) | ❌ No (user only) |
| **SPEC Integration** | ✅ Yes (Gap analysis) | ❌ No |
| **TAG Tracing** | ✅ Yes (Auto) | ⚠️ Manual |
| **Fix Levels** | ✅ 3 levels (Code/SPEC/Arch) | ⚠️ Code only |
| **Pattern DB** | ✅ Language-specific | ⚠️ Generic |
| **Report Generation** | ✅ Markdown + JSON | ⚠️ Text only |

**Use Case**:
- `moai-alfred-error-explainer`: Automated error analysis (Alfred 워크플로우)
- `moai-essentials-debug`: Manual debugging (개발 중)

---

**Author**: Alfred SuperAgent
**Version**: 0.1.0
**License**: MIT
