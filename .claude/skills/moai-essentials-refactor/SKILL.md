---
name: moai-essentials-refactor
description: Refactoring guidance with design patterns and code improvement strategies
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
tier: 1
auto-load: "true"
---

# Alfred Refactoring Coach

## What it does

Refactoring guidance with design pattern recommendations, code smell detection, and step-by-step improvement plans.

## When to use

- "리팩토링 도와줘", "이 코드 개선 방법은?", "디자인 패턴 적용", "코드 정리", "구조 개선", "중복 제거"
- "함수 분리", "클래스 분리", "코드 스멜 제거", "복잡도 낮추기", "패턴 적용", "3회 반복 규칙"
- "Refactoring", "Design patterns", "Code cleanup", "Extract method", "DRY principle", "3-strike rule"
- When code becomes hard to maintain
- Before adding new features to legacy code

## How it works

**Refactoring Techniques**:
- **Extract Method**: 긴 메서드 분리
- **Replace Conditional with Polymorphism**: 조건문 제거
- **Introduce Parameter Object**: 매개변수 그룹화
- **Extract Class**: 거대한 클래스 분리

**Design Pattern Recommendations**:
- Complex object creation → **Builder Pattern**
- Type-specific behavior → **Strategy Pattern**
- Global state → **Singleton Pattern**
- Incompatible interfaces → **Adapter Pattern**
- Delayed object creation → **Factory Pattern**

**3-Strike Rule**:
```
1st occurrence: Just implement
2nd occurrence: Notice similarity (leave as-is)
3rd occurrence: Pattern confirmed → Refactor! 🔧
```

**Refactoring Checklist**:
- [ ] All tests passing before refactoring
- [ ] Code smells identified
- [ ] Refactoring goal clear
- [ ] Change one thing at a time
- [ ] Run tests after each change
- [ ] Commit frequently

## Common Refactoring Patterns by Language

| Pattern | Problem | Solution | Language |
|---------|---------|----------|----------|
| **Extract Method** | Long function | Break into smaller functions | All |
| **Replace Conditional** | Complex if-else | Use polymorphism/strategy | OOP languages |
| **Consolidate Duplicates** | Repeated code | Extract shared logic | All |
| **Replace Magic Number** | Unclear constants | Name them | All |
| **Extract Class** | Large class (>300 LOC) | Split responsibilities | OOP languages |
| **Replace Parameter Object** | Too many parameters (>5) | Group into object | All |
| **Introduce Strategy** | If-else type checking | Use interfaces/traits | OOP/FP languages |
| **Builder Pattern** | Complex construction | Builder class | Java, TypeScript, Rust |

## Examples

### Example 1: Extract Method → Duplicate Code Removal

**❌ Before (Duplicate Logic)**:
```python
# @CODE:REFACTOR-USER-001: 중복 코드
class UserService:
    def create_admin(self, email, name):
        # 검증 1
        if not email or '@' not in email:
            raise ValueError("Invalid email")
        if len(name) < 2:
            raise ValueError("Invalid name")

        # DB 저장
        user = {'email': email, 'name': name, 'role': 'admin'}
        return db.insert("INSERT INTO users VALUES (...)", user)

    def create_user(self, email, name):
        # 검증 1 (중복!)
        if not email or '@' not in email:
            raise ValueError("Invalid email")
        if len(name) < 2:
            raise ValueError("Invalid name")

        # DB 저장
        user = {'email': email, 'name': name, 'role': 'user'}
        return db.insert("INSERT INTO users VALUES (...)", user)

    def update_user(self, user_id, email, name):
        # 검증 1 (중복!)
        if not email or '@' not in email:
            raise ValueError("Invalid email")
        if len(name) < 2:
            raise ValueError("Invalid name")

        # DB 업데이트
        return db.update(f"UPDATE users SET email=?, name=? WHERE id=?",
                        email, name, user_id)

# 문제:
# - 검증 로직 3곳에 중복
# - 수정 시 모두 변경 필요
# - 테스트 어려움
```

**✅ After (Extract Method)**:
```python
# @CODE:REFACTOR-USER-001: 추출된 메서드
class UserService:
    def _validate_user_input(self, email: str, name: str) -> None:
        """사용자 입력 검증 (공통 로직)"""
        if not email or '@' not in email:
            raise ValueError("Invalid email")
        if len(name) < 2:
            raise ValueError("Invalid name")

    def create_admin(self, email: str, name: str) -> int:
        """관리자 계정 생성"""
        self._validate_user_input(email, name)
        return self._insert_user(email, name, 'admin')

    def create_user(self, email: str, name: str) -> int:
        """일반 사용자 계정 생성"""
        self._validate_user_input(email, name)
        return self._insert_user(email, name, 'user')

    def update_user(self, user_id: int, email: str, name: str) -> None:
        """사용자 정보 수정"""
        self._validate_user_input(email, name)
        db.update("UPDATE users SET email=?, name=? WHERE id=?",
                 email, name, user_id)

    def _insert_user(self, email: str, name: str, role: str) -> int:
        """DB에 사용자 저장"""
        user = {'email': email, 'name': name, 'role': role}
        return db.insert("INSERT INTO users VALUES (...)", user)

# 개선:
# ✅ 검증 로직 1곳에만 존재
# ✅ 수정 시 한 곳만 변경
# ✅ 테스트 용이 (_validate_user_input 독립 테스트)
# ✅ DRY (Don't Repeat Yourself) 준수
```

### Example 2: Replace Conditional with Strategy Pattern

**❌ Before (If-Else Hell)**:
```java
// @CODE:REFACTOR-PAYMENT-001: 복잡한 조건문
public class PaymentProcessor {
    public void processPayment(String paymentType, BigDecimal amount) {
        if (paymentType.equals("CREDIT_CARD")) {
            // 신용카드 처리
            String cardToken = getCardToken();
            verifyCardToken(cardToken);
            chargeCard(cardToken, amount);
            logTransaction("CREDIT_CARD", amount);

        } else if (paymentType.equals("PAYPAL")) {
            // 페이팔 처리
            String ppEmail = getPayPalEmail();
            authorizePayPal(ppEmail);
            chargePayPal(ppEmail, amount);
            logTransaction("PAYPAL", amount);

        } else if (paymentType.equals("BANK_TRANSFER")) {
            // 계좌 이체 처리
            String bankAccount = getBankAccount();
            validateBankAccount(bankAccount);
            transferBank(bankAccount, amount);
            logTransaction("BANK_TRANSFER", amount);

        } else {
            throw new IllegalArgumentException("Unknown payment type");
        }
    }
}

// 문제:
// ❌ 메서드가 너무 길고 복잡
// ❌ 새 결제 수단 추가 시 메서드 수정 필요
// ❌ Open/Closed 원칙 위반
// ❌ 테스트 어려움 (모든 분기 테스트 필요)
```

**✅ After (Strategy Pattern)**:
```java
// @CODE:REFACTOR-PAYMENT-001: Strategy 패턴 적용

// 1️⃣ Strategy 인터페이스
public interface PaymentStrategy {
    void process(BigDecimal amount);
}

// 2️⃣ 구체적 전략 구현
public class CreditCardPayment implements PaymentStrategy {
    @Override
    public void process(BigDecimal amount) {
        String cardToken = getCardToken();
        verifyCardToken(cardToken);
        chargeCard(cardToken, amount);
    }
}

public class PayPalPayment implements PaymentStrategy {
    @Override
    public void process(BigDecimal amount) {
        String ppEmail = getPayPalEmail();
        authorizePayPal(ppEmail);
        chargePayPal(ppEmail, amount);
    }
}

public class BankTransferPayment implements PaymentStrategy {
    @Override
    public void process(BigDecimal amount) {
        String bankAccount = getBankAccount();
        validateBankAccount(bankAccount);
        transferBank(bankAccount, amount);
    }
}

// 3️⃣ 간단한 Processor
public class PaymentProcessor {
    private final Map<String, PaymentStrategy> strategies = new HashMap<>();

    public PaymentProcessor() {
        strategies.put("CREDIT_CARD", new CreditCardPayment());
        strategies.put("PAYPAL", new PayPalPayment());
        strategies.put("BANK_TRANSFER", new BankTransferPayment());
    }

    public void processPayment(String type, BigDecimal amount) {
        PaymentStrategy strategy = strategies.get(type);
        if (strategy == null) {
            throw new IllegalArgumentException("Unknown type: " + type);
        }

        strategy.process(amount);
        logTransaction(type, amount);
    }

    // 새 결제 수단 추가 (메서드 수정 없음!)
    public void registerPaymentStrategy(String type, PaymentStrategy strategy) {
        strategies.put(type, strategy);
    }
}

// 개선:
// ✅ 메서드 간결함
// ✅ 새 전략 추가 시 기존 코드 수정 안 함 (Open/Closed)
// ✅ 각 전략 독립 테스트 가능
// ✅ 유지보수 용이
```

### Example 3: Replace Magic Numbers with Named Constants

**❌ Before**:
```go
// @CODE:REFACTOR-AUTH-001: 매직 숫자
func authenticate(password string, attempts int) bool {
    if len(password) < 8 {  // ❓ 8 = 최소 길이?
        return false
    }

    if attempts > 5 {  // ❓ 5 = 최대 시도?
        return false
    }

    if passwordAge > 90 {  // ❓ 90 = 최대 일수?
        return false
    }

    return verifyPassword(password)
}
```

**✅ After**:
```go
// @CODE:REFACTOR-AUTH-001: 명확한 상수
const (
    MIN_PASSWORD_LENGTH = 8
    MAX_LOGIN_ATTEMPTS = 5
    PASSWORD_EXPIRY_DAYS = 90
)

func authenticate(password string, attempts int) bool {
    if len(password) < MIN_PASSWORD_LENGTH {  // ✅ 의도 명확
        return false
    }

    if attempts > MAX_LOGIN_ATTEMPTS {
        return false
    }

    if passwordAge > PASSWORD_EXPIRY_DAYS {
        return false
    }

    return verifyPassword(password)
}
```

### Example 4: Extract Class → Single Responsibility

**❌ Before (God Class)**:
```typescript
// @CODE:REFACTOR-USER-001: 너무 많은 책임 (400+ LOC)
class User {
    id: number;
    email: string;
    password: string;

    // 책임 1: 사용자 정보 관리
    updateProfile(name, bio) { /* ... */ }
    getProfile() { /* ... */ }

    // 책임 2: 이메일 검증
    validateEmail() { /* ... */ }
    sendVerificationEmail() { /* ... */ }
    confirmEmailVerification(token) { /* ... */ }

    // 책임 3: 비밀번호 관리
    hashPassword() { /* ... */ }
    verifyPassword(input) { /* ... */ }
    resetPassword(token) { /* ... */ }

    // 책임 4: 보안 감사
    logLoginAttempt() { /* ... */ }
    checkBruteForceAttempt() { /* ... */ }
    requirePasswordReset() { /* ... */ }

    // 책임 5: 알림
    sendPasswordResetEmail() { /* ... */ }
    sendLoginAlert() { /* ... */ }
}

// 문제:
// ❌ 5가지 책임 혼재
// ❌ 한 책임 수정 시 다른 부분 영향
// ❌ 테스트 어려움
// ❌ 재사용 불가
```

**✅ After (Separated Responsibilities)**:
```typescript
// @CODE:REFACTOR-USER-001: 책임 분리

// 클래스 1: 사용자 정보 (책임 1)
class User {
    id: number;
    email: string;
    passwordHash: string;

    updateProfile(name: string, bio: string) { /* ... */ }
    getProfile() { /* ... */ }
}

// 클래스 2: 이메일 검증 (책임 2)
class EmailVerificationService {
    validateEmail(email: string): boolean { /* ... */ }
    sendVerificationEmail(user: User) { /* ... */ }
    confirmVerification(token: string) { /* ... */ }
}

// 클래스 3: 비밀번호 관리 (책임 3)
class PasswordService {
    hashPassword(password: string): string { /* ... */ }
    verifyPassword(input: string, hash: string): boolean { /* ... */ }
    resetPassword(user: User, token: string) { /* ... */ }
}

// 클래스 4: 보안 감사 (책임 4)
class SecurityAuditService {
    logLoginAttempt(userId: number) { /* ... */ }
    checkBruteForceAttempt(userId: number): boolean { /* ... */ }
    requirePasswordReset(user: User) { /* ... */ }
}

// 클래스 5: 알림 (책임 5)
class NotificationService {
    sendPasswordResetEmail(user: User) { /* ... */ }
    sendLoginAlert(user: User) { /* ... */ }
}

// 조율 클래스 (조합만)
class UserAuthenticationManager {
    constructor(
        private userService: UserService,
        private emailService: EmailVerificationService,
        private passwordService: PasswordService,
        private auditService: SecurityAuditService,
        private notificationService: NotificationService
    ) {}

    authenticate(email: string, password: string): User {
        const user = this.userService.findByEmail(email);
        this.auditService.logLoginAttempt(user.id);

        if (this.auditService.checkBruteForceAttempt(user.id)) {
            throw new Error("Too many attempts");
        }

        if (!this.passwordService.verifyPassword(password, user.passwordHash)) {
            throw new Error("Invalid password");
        }

        this.notificationService.sendLoginAlert(user);
        return user;
    }
}

// 개선:
// ✅ 각 클래스 = 1 책임
// ✅ 각 부분 독립 테스트
// ✅ 재사용 가능 (다른 서비스에서 PasswordService 재사용)
// ✅ 수정 시 영향 범위 최소화
```

## Refactoring Workflow

```
1️⃣ 코드 냄새 식별
   └─ Long method, duplicate, magic numbers, ...

2️⃣ 테스트 작성
   └─ 기존 동작 보장

3️⃣ 작은 변경 수행
   └─ 한 번에 한 가지만

4️⃣ 테스트 실행
   └─ 모든 테스트 통과 확인

5️⃣ 반복
   └─ 다음 냄새로 이동

6️⃣ 커밋
   └─ 각 단계별 커밋
```

## Keywords

"리팩토링", "코드 정리", "디자인 패턴", "Extract Method", "Strategy", "중복 제거", "복잡도 감소", "코드 냄새", "refactoring techniques", "design patterns", "SOLID principles"

## Reference

- Refactoring techniques: `.moai/memory/development-guide.md#리팩토링-기법`
- Design patterns: CLAUDE.md#디자인-패턴
- Code improvement: `.moai/memory/development-guide.md#코드-개선-전략`

## Works well with

- moai-essentials-review (품질 검증)
- moai-essentials-debug (오류 분석)
