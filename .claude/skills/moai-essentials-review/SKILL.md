---
name: moai-essentials-review
description: Automated code review with SOLID principles, code smells, and language-specific best practices
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
tier: 1
auto-load: "true"
---

# Alfred Code Reviewer

## What it does

Automated code review with language-specific best practices, SOLID principles verification, and code smell detection.

## When to use

- "코드 리뷰해줘", "이 코드 개선점은?", "코드 품질 확인", "리뷰 부탁해", "문제점 찾아줘", "개선 제안"
- "SOLID 원칙", "베스트 프랙티스", "코드 스멜", "안티패턴", "보안 취약점", "디자인 패턴"
- "Code review", "Quality check", "Best practices", "Security audit", "SOLID principles", "Code smells"
- Optionally invoked after `/alfred:3-sync`
- Before merging PR or releasing
- During peer code review

## How it works

**Code Constraints Check**:
- File ≤300 LOC
- Function ≤50 LOC
- Parameters ≤5
- Cyclomatic complexity ≤10

**SOLID Principles**:
- Single Responsibility
- Open/Closed
- Liskov Substitution
- Interface Segregation
- Dependency Inversion

**Code Smell Detection**:
- Long Method
- Large Class
- Duplicate Code
- Dead Code
- Magic Numbers

**Language-specific Best Practices**:
- Python: List comprehension, type hints, PEP 8
- TypeScript: Strict typing, async/await, error handling
- Java: Streams API, Optional, Design patterns

**Review Report**:
```markdown
## Code Review Report

### 🔴 Critical Issues (3)
1. **src/auth/service.py:45** - Function too long (85 > 50 LOC)
2. **src/api/handler.ts:120** - Missing error handling
3. **src/db/repository.java:200** - Magic number

### ⚠️ Warnings (5)
1. **src/utils/helper.py:30** - Unused import

### ✅ Good Practices Found
- Test coverage: 92%
- Consistent naming
```

## Examples

### Example 1: Function Too Long (> 50 LOC) → Extract Method

**❌ Before (Code Smell: Long Method)**:
```python
# @CODE:USER-SERVICE-001: 길이 85 LOC - 코드 냄새
def process_user_registration(email, password, phone):
    """사용자 등록 프로세스 (너무 길음)"""
    # 유효성 검증
    if not email or '@' not in email:
        raise ValueError("Invalid email")
    if len(password) < 8:
        raise ValueError("Password too short")
    if not phone or len(phone) < 10:
        raise ValueError("Invalid phone")

    # 데이터베이스 조회 (6줄)
    existing = db.query("SELECT * FROM users WHERE email = %s", email)
    if existing:
        raise ValueError("Email already registered")

    # 비밀번호 해싱 (3줄)
    salt = generate_salt()
    hashed = bcrypt.hash(password, salt)

    # 이메일 검증 토큰 생성 (3줄)
    token = secrets.token_urlsafe(32)
    token_hash = hash_token(token)

    # 데이터베이스 저장 (5줄)
    user = {
        'email': email,
        'password_hash': hashed,
        'phone': phone,
        'email_verified': False,
        'token_hash': token_hash
    }
    user_id = db.insert("INSERT INTO users VALUES (...)", user)

    # 이메일 전송 (5줄)
    send_email(
        email,
        subject="Verify your email",
        body=f"Click: {BASE_URL}/verify?token={token}"
    )

    # 로그 기록 (3줄)
    logger.info(f"User registered: {email}")
    audit_log.insert(user_id, "registration")

    return {"user_id": user_id, "message": "Check your email"}

# 문제점:
# ❌ 함수가 너무 길어서 테스트 어려움
# ❌ 여러 책임 혼재 (검증, 해싱, DB, 이메일)
# ❌ 각 단계 수정 시 전체 함수 영향
```

**✅ After (Clean: Extracted Methods)**:
```python
# @CODE:USER-SERVICE-001: 각 메서드 ≤50 LOC
def process_user_registration(email: str, password: str, phone: str) -> Dict:
    """사용자 등록 프로세스 (조율만 담당)"""
    # 단일 책임: 조율
    validate_registration_input(email, password, phone)

    if user_exists(email):
        raise ValueError("Email already registered")

    password_hash = hash_password(password)
    verification_token = generate_verification_token()

    user_id = create_user(email, password_hash, phone, verification_token)

    send_verification_email(email, verification_token)
    log_registration(user_id, email)

    return {"user_id": user_id, "message": "Check your email"}

# 추출된 메서드들 (각 ≤30 LOC):
def validate_registration_input(email: str, password: str, phone: str) -> None:
    """@CODE:USER-SERVICE-001:VALIDATION"""
    if not email or '@' not in email:
        raise ValueError("Invalid email")
    if len(password) < 8:
        raise ValueError("Password too short")
    if not phone or len(phone) < 10:
        raise ValueError("Invalid phone")

def hash_password(password: str) -> str:
    """@CODE:USER-SERVICE-001:SECURITY"""
    salt = generate_salt()
    return bcrypt.hash(password, salt)

def create_user(email: str, password_hash: str, phone: str, token: str) -> int:
    """@CODE:USER-SERVICE-001:PERSISTENCE"""
    user = {
        'email': email,
        'password_hash': password_hash,
        'phone': phone,
        'email_verified': False,
        'token_hash': hash_token(token)
    }
    return db.insert("INSERT INTO users VALUES (...)", user)

# 개선점:
# ✅ 각 메서드 ≤30 LOC
# ✅ 단일 책임 원칙 (SRP)
# ✅ 각 부분 독립적 테스트 가능
# ✅ 유지보수 용이
```

**코드 리뷰 결과**:
```
Before:  85 LOC (1 function) → 테스트 어려움, 복잡도 높음
After:  28 + 10 + 12 + 8 = 58 LOC (5 functions) → 명확, 테스트 용이

Complexity: 12 → 3 per function ✅
Testability: 1 test case → 5 independent tests ✅
Maintainability: 고 → 낮 ✅
```

### Example 2: SOLID Violation - Single Responsibility Principle

**❌ Before (SRP Violation)**:
```typescript
// @CODE:REPORT-GENERATOR-001: 여러 책임 혼재
class ReportGenerator {
    generateReport(userId: number): void {
        // 책임 1: DB 조회
        const user = database.findUser(userId);
        const sales = database.findSales(userId);

        // 책임 2: 계산
        const total = sales.reduce((sum, s) => sum + s.amount, 0);
        const average = total / sales.length;
        const tax = total * 0.1;

        // 책임 3: 포맷팅
        let report = `USER REPORT\n`;
        report += `Name: ${user.name}\n`;
        report += `Total Sales: $${total}\n`;
        report += `Average: $${average}\n`;
        report += `Tax: $${tax}\n`;

        // 책임 4: 파일 저장
        fs.writeFileSync(`reports/${userId}.txt`, report);

        // 책임 5: 이메일 발송
        emailService.send(user.email, report);

        // 책임 6: 로깅
        logger.info(`Report generated for user ${userId}`);
    }
}

// 문제점:
// ❌ 5가지 이상의 책임 혼재
// ❌ DB 변경 시 영향
// ❌ 포맷 변경 시 영향
// ❌ 단위 테스트 불가능 (모의객체 5개 필요)
```

**✅ After (SRP Adherence)**:
```typescript
// @CODE:REPORT-GENERATOR-001: 각 클래스 = 1책임

// 책임 1: 데이터 조회
interface IUserRepository {
    findUser(userId: number): User;
    findSales(userId: number): Sale[];
}

// 책임 2: 계산
class SalesCalculator {
    calculateTotal(sales: Sale[]): number {
        return sales.reduce((sum, s) => sum + s.amount, 0);
    }
    calculateTax(total: number): number {
        return total * 0.1;
    }
}

// 책임 3: 포맷팅
class ReportFormatter {
    format(user: User, stats: SalesStats): string {
        return `USER REPORT\n...`;
    }
}

// 책임 4-5: 저장/전송
interface IReportDelivery {
    save(userId: number, report: string): void;
    send(email: string, report: string): void;
}

// 조율 클래스 (조합만)
class ReportGenerator {
    constructor(
        private repository: IUserRepository,
        private calculator: SalesCalculator,
        private formatter: ReportFormatter,
        private delivery: IReportDelivery
    ) {}

    generateReport(userId: number): void {
        const user = this.repository.findUser(userId);
        const sales = this.repository.findSales(userId);

        const total = this.calculator.calculateTotal(sales);
        const tax = this.calculator.calculateTax(total);

        const report = this.formatter.format(user, { total, tax });

        this.delivery.save(userId, report);
        this.delivery.send(user.email, report);
    }
}

// 개선점:
// ✅ 각 클래스 = 1 책임
// ✅ 모의객체 쉬움
// ✅ 각 부분 독립 테스트:
//   - SalesCalculator 테스트 (계산 로직만)
//   - ReportFormatter 테스트 (포맷 로직만)
//   - 의존성 주입으로 테스트 용이
```

### Example 3: Magic Numbers → Named Constants

**❌ Before (Code Smell)**:
```java
// @CODE:PRICING-001: 매직 숫자
public class PricingEngine {
    public double calculateDiscount(int quantity, double price) {
        if (quantity >= 100) {
            return price * 0.15;  // ❓ 0.15 = 15%? 뭐?
        }
        if (quantity >= 50) {
            return price * 0.10;  // ❓ 0.10 = 10%?
        }
        if (quantity >= 10) {
            return price * 0.05;  // ❓ 0.05 = 5%?
        }
        return price;
    }

    public double calculateShipping(double weight) {
        return weight * 2.5;  // ❓ 2.5 = $per kg?
    }
}

// 문제점:
// ❌ 의도 불명확
// ❌ 수정 시 어떤 값인지 모름
// ❌ 여러 곳에 분산되면 일관성 문제
```

**✅ After (Named Constants)**:
```java
// @CODE:PRICING-001: 명확한 상수
public class PricingEngine {
    // 수량 할인 기준
    private static final int BULK_TIER_1_QUANTITY = 10;
    private static final double BULK_TIER_1_DISCOUNT = 0.05;

    private static final int BULK_TIER_2_QUANTITY = 50;
    private static final double BULK_TIER_2_DISCOUNT = 0.10;

    private static final int BULK_TIER_3_QUANTITY = 100;
    private static final double BULK_TIER_3_DISCOUNT = 0.15;

    // 배송료
    private static final double SHIPPING_RATE_PER_KG = 2.5;

    public double calculateDiscount(int quantity, double price) {
        if (quantity >= BULK_TIER_3_QUANTITY) {
            return price * BULK_TIER_3_DISCOUNT;  // ✅ 의도 명확
        }
        if (quantity >= BULK_TIER_2_QUANTITY) {
            return price * BULK_TIER_2_DISCOUNT;
        }
        if (quantity >= BULK_TIER_1_QUANTITY) {
            return price * BULK_TIER_1_DISCOUNT;
        }
        return price;
    }

    public double calculateShipping(double weight) {
        return weight * SHIPPING_RATE_PER_KG;  // ✅ 의도 명확
    }
}

// 개선점:
// ✅ 의도 명확 (상수명으로 설명)
// ✅ 수정 시 한 곳만 변경
// ✅ 단위 테스트 용이 (상수 주입 가능)
```

### Example 4: Code Review Report (Complete Example)

**코드 리뷰 보고서**:
```markdown
## Code Review Report: auth-service.ts

### 🔴 Critical Issues (2)

**1. Function Too Long**
- **File**: src/auth/service.ts:42
- **Function**: authenticateUser() - 87 LOC (limit: 50)
- **Fix**: Extract validateCredentials(), hashPassword() methods

**2. Unused Import**
- **File**: src/auth/service.ts:3
- **Import**: `import * as crypto from 'crypto'` (unused, bcrypt 사용함)
- **Fix**: Remove import

### ⚠️ Warnings (3)

**1. Magic Number**
- **File**: src/auth/service.ts:65
- **Code**: `if (attempts > 5) { ... }`
- **Fix**: Use constant `const MAX_LOGIN_ATTEMPTS = 5`

**2. Missing Error Handling**
- **File**: src/auth/service.ts:78
- **Code**: `const token = jwt.sign(...)` without try-catch
- **Fix**: Wrap in try-catch or use .catch()

**3. Type Any Usage**
- **File**: src/auth/service.ts:45
- **Code**: `function verify(token: any): boolean`
- **Fix**: Use specific type `Token | string`

### ✅ Good Practices Found

- ✅ Test coverage: 94% (target: 85%+)
- ✅ Consistent naming convention
- ✅ Input validation present
- ✅ Error messages descriptive
- ✅ Logging appropriate

### 📊 Metrics

| Metric | Current | Target | Status |
|--------|---------|--------|--------|
| LOC per file | 287 | ≤300 | ✅ |
| LOC per function | 87 | ≤50 | ❌ |
| Complexity | 8 | ≤10 | ✅ |
| Test coverage | 94% | ≥85% | ✅ |
| Unused imports | 1 | 0 | ⚠️ |

### 🔧 Recommended Actions

1. **Immediate (Day 1)**
   - [ ] Extract functions: authenticateUser() → validateCredentials() + authenticateUser()
   - [ ] Remove unused crypto import

2. **Soon (This Sprint)**
   - [ ] Define MAX_LOGIN_ATTEMPTS constant
   - [ ] Add try-catch for JWT operations
   - [ ] Update types (any → specific)

3. **Follow-up**
   - [ ] Re-run review after fixes
   - [ ] Verify test coverage remains ≥94%
```

## Keywords

"코드 리뷰", "SOLID 원칙", "코드 냄새", "함수 길이", "복잡도", "매직 숫자", "타입 검사", "테스트 가능성", "code smell detection", "best practices", "code quality metrics"

## Reference

- SOLID Principles guide: `.moai/memory/development-guide.md#SOLID-원칙`
- Code constraints: CLAUDE.md#코드-제약
- Language-specific style: moai-lang-* skills

## Works well with

- moai-foundation-specs (요구사항 검증)
- moai-essentials-refactor (리팩토링 지원)
- moai-essentials-perf (성능 검증)
