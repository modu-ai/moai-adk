---
name: moai-alfred-code-reviewer
description: Automated code review with language-specific best practices, SOLID principles, and actionable improvement suggestions (user)
tier: 4
---

# Alfred Code Reviewer

## What it does

Automated code review with language-specific best practices, SOLID principles verification, and code smell detection. **Auto-invoked by Alfred** after `/alfred:3-sync` to ensure code quality before PR creation.

## When to use

**Auto-invoked by Alfred**:
- After `/alfred:3-sync` completes
- Before PR status changes to "Ready for Review"
- When TRUST 5-principles validation is needed

**Manually invoked by users**:
- "코드 리뷰해줘"
- "이 코드 개선점은?"
- "SOLID 원칙 확인"
- Before merging PR

## How it works

### Phase 1: Code Constraints Check

**File-level checks**:
- File ≤300 LOC
- No duplicate code across files
- Consistent naming conventions

**Function-level checks**:
- Function ≤50 LOC
- Parameters ≤5
- Cyclomatic complexity ≤10
- Single Responsibility Principle

**Example violation**:
```python
# ❌ BAD: Function too long (85 LOC)
def process_user_data(user_id, data, config, options, flags):
    # 85 lines of code...
    pass

# ✅ GOOD: Refactored
def process_user_data(user_id: str, data: dict) -> Result:
    validated = _validate_data(data)  # 10 LOC
    processed = _transform_data(validated)  # 15 LOC
    return _save_data(user_id, processed)  # 12 LOC
```

---

### Phase 2: SOLID Principles Verification

**S - Single Responsibility**:
```python
# ❌ BAD: Multiple responsibilities
class UserManager:
    def create_user(self): ...
    def send_email(self): ...  # ← Email sending responsibility
    def generate_report(self): ...  # ← Reporting responsibility

# ✅ GOOD: Separated
class UserService:
    def create_user(self): ...

class EmailService:
    def send_email(self): ...

class ReportGenerator:
    def generate_report(self): ...
```

**O - Open/Closed**:
```typescript
// ❌ BAD: Modify existing code
function calculateDiscount(type: string, amount: number) {
  if (type === 'student') return amount * 0.9;
  if (type === 'senior') return amount * 0.8;
  // Adding new type requires modifying this function
}

// ✅ GOOD: Extension without modification
interface DiscountStrategy {
  calculate(amount: number): number;
}

class StudentDiscount implements DiscountStrategy {
  calculate(amount: number) { return amount * 0.9; }
}

class SeniorDiscount implements DiscountStrategy {
  calculate(amount: number) { return amount * 0.8; }
}
```

**L - Liskov Substitution**:
```java
// ❌ BAD: Subclass breaks parent contract
class Rectangle {
    void setWidth(int w) { ... }
    void setHeight(int h) { ... }
}

class Square extends Rectangle {
    void setWidth(int w) {
        super.setWidth(w);
        super.setHeight(w);  // ← Breaks LSP
    }
}

// ✅ GOOD: Separate abstractions
interface Shape {
    int getArea();
}

class Rectangle implements Shape { ... }
class Square implements Shape { ... }
```

**I - Interface Segregation**:
```typescript
// ❌ BAD: Fat interface
interface Worker {
  code(): void;
  test(): void;
  deploy(): void;  // ← Not all workers deploy
}

// ✅ GOOD: Segregated interfaces
interface Developer {
  code(): void;
  test(): void;
}

interface DevOpsEngineer extends Developer {
  deploy(): void;
}
```

**D - Dependency Inversion**:
```python
# ❌ BAD: High-level depends on low-level
class OrderService:
    def __init__(self):
        self.db = MySQLDatabase()  # ← Concrete dependency

# ✅ GOOD: Both depend on abstraction
class OrderService:
    def __init__(self, db: DatabaseInterface):
        self.db = db  # ← Abstract dependency
```

---

### Phase 3: Code Smell Detection

**Long Method** (>50 LOC):
```python
# Alfred detects:
⚠️ src/auth/service.py:45 - Function 'authenticate_user' is 85 LOC (max 50)

💡 Suggestion: Extract methods
  - _validate_credentials (15 LOC)
  - _check_permissions (20 LOC)
  - _generate_token (18 LOC)
```

**Large Class** (>300 LOC):
```typescript
// Alfred detects:
⚠️ src/api/handler.ts:1 - Class 'ApiHandler' is 450 LOC (max 300)

💡 Suggestion: Split into
  - UserApiHandler (120 LOC)
  - ProductApiHandler (150 LOC)
  - OrderApiHandler (180 LOC)
```

**Duplicate Code**:
```java
// Alfred detects:
⚠️ Duplicate code detected (15 lines)
  - src/service/UserService.java:45-60
  - src/service/AdminService.java:78-93

💡 Suggestion: Extract common method to BaseService
```

**Dead Code**:
```python
# Alfred detects:
⚠️ src/utils/helper.py:120 - Unused function 'legacy_processor'

💡 Suggestion: Remove or document why it's kept
```

**Magic Numbers**:
```typescript
// Alfred detects:
⚠️ src/config/settings.ts:42 - Magic number 86400

💡 Suggestion:
const SECONDS_IN_DAY = 86400;
const timeout = SECONDS_IN_DAY;
```

---

### Phase 4: Language-specific Best Practices

**Python**:
- ✅ List comprehension over loops
- ✅ Type hints for function signatures
- ✅ PEP 8 compliance
- ✅ `with` statements for resource management

**TypeScript**:
- ✅ Strict typing enabled
- ✅ `async/await` over callbacks
- ✅ Error handling with `try/catch`
- ✅ `const` over `let` when possible

**Java**:
- ✅ Streams API over imperative loops
- ✅ `Optional` instead of null
- ✅ Design patterns (Builder, Factory)
- ✅ Try-with-resources

---

### Phase 5: TRUST 5-Principles Integration

**T - Test First**:
```bash
# Alfred checks:
✅ Test coverage: 92% (target: 85%)
⚠️ Missing tests for:
  - src/auth/service.py::refresh_token (line 120)
```

**R - Readable**:
```bash
# Alfred checks:
✅ All files ≤300 LOC
⚠️ src/api/handler.ts:45 - Function too long (62 LOC)
```

**U - Unified**:
```bash
# Alfred checks:
✅ SPEC compliance verified
⚠️ @CODE:AUTH-001 missing SPEC reference
```

**S - Secured**:
```bash
# Alfred checks:
⚠️ Potential SQL injection: src/db/query.py:78
⚠️ Hardcoded secret: src/config/settings.py:12
```

**T - Trackable**:
```bash
# Alfred checks:
✅ TAG chain complete: @SPEC → @TEST → @CODE → @DOC
⚠️ Orphan TAG: @CODE:PAYMENT-005 (no corresponding @SPEC)
```

---

## Review Report Format

```markdown
## 📊 Code Review Report

**Project**: MoAI-ADK
**Review Date**: 2025-10-20
**Reviewed Files**: 15 files
**Total Issues**: 8 (3 Critical, 5 Warnings)

---

### 🔴 Critical Issues (3) - Must Fix

1. **src/auth/service.py:45**
   - **Issue**: Function too long (85 > 50 LOC)
   - **Principle**: Single Responsibility (SOLID)
   - **Fix**: Extract 3 helper methods
   - **Priority**: High

2. **src/api/handler.ts:120**
   - **Issue**: Missing error handling
   - **Principle**: Secured (TRUST)
   - **Fix**: Add try/catch block
   - **Priority**: Critical

3. **src/db/repository.java:200**
   - **Issue**: Magic number 86400
   - **Principle**: Readable (TRUST)
   - **Fix**: Extract constant SECONDS_IN_DAY
   - **Priority**: Medium

---

### ⚠️ Warnings (5) - Consider Fixing

1. **src/utils/helper.py:30**
   - **Issue**: Unused import 'datetime'
   - **Fix**: Remove unused import

2. **src/config/settings.ts:50**
   - **Issue**: Using 'any' type
   - **Fix**: Define proper interface

3. **src/service/UserService.java:75**
   - **Issue**: Imperative loop (can use Stream API)
   - **Fix**: Refactor to Stream API

4. **src/auth/jwt.py:120**
   - **Issue**: Duplicate code with src/auth/oauth.py:85
   - **Fix**: Extract common validation logic

5. **src/api/routes.ts:200**
   - **Issue**: Function has 6 parameters (max 5)
   - **Fix**: Use options object

---

### ✅ Good Practices Found (10)

1. ✅ Consistent naming conventions across all files
2. ✅ Test coverage: 92% (target: 85%)
3. ✅ All @TAG chains are complete
4. ✅ SPEC compliance verified
5. ✅ Type hints used in Python code
6. ✅ Strict typing enabled in TypeScript
7. ✅ Error handling implemented
8. ✅ No hardcoded secrets detected
9. ✅ Dependency Inversion followed
10. ✅ Interface Segregation applied

---

## 🎯 Next Steps

**Before PR Ready**:
1. Fix 3 Critical Issues
2. Consider 5 Warnings
3. Re-run Alfred Code Reviewer

**Commands**:
```bash
# Fix issues and re-review
moai-alfred-code-reviewer

# Or integrate with sync
/alfred:3-sync
```

---

**Alfred Recommendation**: 🟡 Review Needed (3 Critical Issues)
```

---

## Integration with /alfred:3-sync

**Workflow**:
```bash
# User runs sync
/alfred:3-sync

# Phase 1: TAG 체인 검증
✅ TAG chain complete

# Phase 2: 문서 동기화
✅ Living Document updated

# Phase 3: Alfred Code Reviewer (AUTO) ⭐
⚙️ Running code review...
⚠️ Found 3 Critical Issues
📋 Report: .moai/reports/code-review-2025-10-20.md

# Phase 4: PR Ready 전환 (Conditional)
❌ BLOCKED: Fix Critical Issues before PR Ready
```

---

## Examples

### Example 1: Auto-invoked after sync

```bash
# User completes implementation
/alfred:3-sync

# Alfred automatically runs code review
⚙️ Alfred Code Reviewer: Analyzing code...
✅ Code Constraints: Pass
⚠️ SOLID Principles: 2 violations
✅ Code Smells: None
⚠️ TRUST 5-Principles: 1 issue

📋 Review Report: .moai/reports/code-review-2025-10-20.md

🟡 Recommendation: Review Needed
   - Fix 2 SOLID violations
   - Fix 1 TRUST issue
   - Then re-run /alfred:3-sync
```

### Example 2: Manual invocation

```bash
# User wants code review before commit
User: "이 코드 리뷰해줘"

Alfred: "코드 리뷰를 시작하겠습니다."

# Alfred uses moai-alfred-code-reviewer Skill
⚙️ Analyzing src/auth/service.py...
⚠️ Function 'authenticate_user' is 85 LOC (max 50)

💡 Suggestion:
   - Extract _validate_credentials (15 LOC)
   - Extract _check_permissions (20 LOC)
   - Extract _generate_token (18 LOC)

✅ After refactoring, all functions ≤50 LOC
```

---

## Works well with

**Skills**:
- moai-foundation-specs (SPEC 준수 검증)
- moai-foundation-trust (TRUST 5원칙 검증)
- moai-essentials-refactor (리팩토링 제안)

**Commands**:
- `/alfred:3-sync` (자동 호출)
- `/alfred:2-build` (TDD 완료 후)

**Agents**:
- trust-checker (TRUST 검증 위임)
- tag-agent (TAG 체인 검증 위임)

---

## Configuration

**.moai/config.json**:
```json
{
  "code-review": {
    "auto-invoke": true,
    "block-pr-on-critical": true,
    "severity-threshold": "warning",
    "checks": {
      "file-loc": 300,
      "function-loc": 50,
      "parameters": 5,
      "complexity": 10
    }
  }
}
```

---

## Differences from moai-essentials-review

| Feature | moai-alfred-code-reviewer | moai-essentials-review |
|---------|---------------------------|------------------------|
| **Auto-invoked** | ✅ Yes (by Alfred) | ❌ No (user only) |
| **TRUST Integration** | ✅ Yes | ⚠️ Partial |
| **SPEC Compliance** | ✅ Yes | ❌ No |
| **TAG Validation** | ✅ Yes | ❌ No |
| **PR Blocking** | ✅ Yes (Critical Issues) | ❌ No |
| **Report Generation** | ✅ Markdown + JSON | ⚠️ Text only |

**Use Case**:
- `moai-alfred-code-reviewer`: Automated quality gate (Alfred 워크플로우)
- `moai-essentials-review`: Quick manual review (개발 중)

---

**Author**: Alfred SuperAgent
**Version**: 0.1.0
**License**: MIT
