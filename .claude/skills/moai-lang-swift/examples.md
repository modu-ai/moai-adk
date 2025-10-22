# Swift 6.0 Code Examples

Production-ready examples for modern Swift development with XCTest 6.0, SwiftLint 0.57.0, Swift 6 concurrency, and iOS/macOS patterns.

---

## Example 1: XCTest 6.0 with setUp/tearDown and Async Tests

### Test File: `Tests/UserServiceTests.swift`

```swift
// @TEST:USER-001 | SPEC: SPEC-USER-001.md
// XCTest demonstration with modern Swift 6 patterns
import XCTest
@testable import MyApp

final class UserServiceTests: XCTestCase {
    // MARK: - Properties

    var sut: UserService!
    var mockDatabase: MockDatabase!

    // MARK: - Lifecycle

    override func setUp() {
        super.setUp()
        mockDatabase = MockDatabase()
        sut = UserService(database: mockDatabase)
    }

    override func tearDown() {
        sut = nil
        mockDatabase = nil
        super.tearDown()
    }

    // MARK: - Tests

    func testUserCreation_ValidData_ReturnsUser() throws {
        // Given
        let email = "alice@example.com"
        let username = "alice"

        // When
        let user = try sut.createUser(email: email, username: username)

        // Then
        XCTAssertEqual(user.email, email)
        XCTAssertEqual(user.username, username)
        XCTAssertGreaterThan(user.id, 0)
    }

    func testUserCreation_InvalidEmail_ThrowsError() {
        // Given
        let invalidEmail = "invalid-email"
        let username = "alice"

        // When/Then
        XCTAssertThrowsError(try sut.createUser(email: invalidEmail, username: username)) { error in
            XCTAssertEqual(error as? UserServiceError, .invalidEmail)
        }
    }

    func testFetchUser_ExistingUser_ReturnsUser() async throws {
        // Given
        let existingUser = User(id: 1, email: "alice@example.com", username: "alice")
        mockDatabase.users = [existingUser]

        // When
        let user = try await sut.fetchUser(id: 1)

        // Then
        XCTAssertEqual(user, existingUser)
    }

    func testFetchUser_NonExistentUser_ThrowsError() async {
        // Given
        mockDatabase.users = []

        // When/Then
        do {
            _ = try await sut.fetchUser(id: 999)
            XCTFail("Expected userNotFound error")
        } catch {
            XCTAssertEqual(error as? UserServiceError, .userNotFound)
        }
    }

    func testConcurrentUserFetches_MultipleUsers_AllSucceed() async throws {
        // Given
        let users = (1...10).map { User(id: $0, email: "user\($0)@example.com", username: "user\($0)") }
        mockDatabase.users = users

        // When
        try await withThrowingTaskGroup(of: User.self) { group in
            for userId in 1...10 {
                group.addTask {
                    try await self.sut.fetchUser(id: userId)
                }
            }

            // Then
            var fetchedUsers: [User] = []
            for try await user in group {
                fetchedUsers.append(user)
            }

            XCTAssertEqual(fetchedUsers.count, 10)
        }
    }
}
```

**Key Features**:
- ✅ `setUp()` and `tearDown()` for test isolation
- ✅ `throws` marking for cleaner error handling
- ✅ `async` tests with `await` for asynchronous operations
- ✅ `withThrowingTaskGroup` for concurrent testing
- ✅ Given-When-Then structure for clarity

**Run Commands**:
```bash
# Run all tests
swift test

# Run specific test case
swift test --filter UserServiceTests

# Generate code coverage
swift test --enable-code-coverage

# View coverage report
xcrun llvm-cov show .build/debug/MyAppPackageTests.xctest/Contents/MacOS/MyAppPackageTests \
  --instr-profile .build/debug/codecov/default.profdata \
  --use-color --format=html > coverage.html
```

---

## Example 2: SwiftLint 0.57.0 Configuration

### Project Configuration: `.swiftlint.yml`

```yaml
# SwiftLint 0.57.0 Configuration

included:
  - Sources
  - Tests

excluded:
  - Pods
  - .build
  - DerivedData

disabled_rules:
  - trailing_whitespace
  - todo

opt_in_rules:
  - empty_count
  - explicit_init
  - force_unwrapping
  - sorted_imports

line_length:
  warning: 120
  error: 150

file_length:
  warning: 400
  error: 500

function_body_length:
  warning: 50
  error: 100

cyclomatic_complexity:
  warning: 10
  error: 20

custom_rules:
  no_print:
    regex: 'print\('
    message: "Use Logger instead of print()"
    severity: warning
```

### Test Configuration: `Tests/.swiftlint.yml`

```yaml
# Disable strict rules for test files
disabled_rules:
  - force_unwrapping
  - force_try
  - function_body_length

line_length:
  warning: 150
  error: 200
```

**Run Commands**:
```bash
# Lint all files
swiftlint

# Auto-fix violations
swiftlint --fix

# Lint specific file
swiftlint lint --path Sources/UserService.swift

# Generate HTML report
swiftlint lint --reporter html > report.html

# Strict mode (warnings as errors)
swiftlint lint --strict
```

---

## Example 3: Swift 6 Concurrency with Actors

### Implementation: `Sources/BankAccount.swift`

```swift
// @CODE:BANK-001 | SPEC: SPEC-BANK-001.md | TEST: Tests/BankAccountTests.swift

import Foundation

actor BankAccount {
    private(set) var balance: Decimal
    private(set) var transactions: [Transaction] = []
    let accountNumber: String

    init(accountNumber: String, initialBalance: Decimal = 0) {
        self.accountNumber = accountNumber
        self.balance = initialBalance
    }

    func deposit(amount: Decimal) throws {
        guard amount > 0 else {
            throw BankAccountError.invalidAmount
        }

        balance += amount

        let transaction = Transaction(
            type: .deposit,
            amount: amount,
            timestamp: Date(),
            balanceAfter: balance
        )
        transactions.append(transaction)
    }

    func withdraw(amount: Decimal) throws {
        guard amount > 0 else {
            throw BankAccountError.invalidAmount
        }

        guard balance >= amount else {
            throw BankAccountError.insufficientFunds
        }

        balance -= amount

        let transaction = Transaction(
            type: .withdrawal,
            amount: amount,
            timestamp: Date(),
            balanceAfter: balance
        )
        transactions.append(transaction)
    }

    func transfer(amount: Decimal, to destination: BankAccount) async throws {
        try withdraw(amount: amount)
        try await destination.deposit(amount: amount)
    }
}

struct Transaction: Sendable, Identifiable {
    let id = UUID()
    let type: TransactionType
    let amount: Decimal
    let timestamp: Date
    let balanceAfter: Decimal
}

enum TransactionType: String, Sendable {
    case deposit
    case withdrawal
}

enum BankAccountError: Error {
    case invalidAmount
    case insufficientFunds
}
```

### Test File: `Tests/BankAccountTests.swift`

```swift
// @TEST:BANK-001 | SPEC: SPEC-BANK-001.md

import XCTest
@testable import MyApp

final class BankAccountTests: XCTestCase {
    func testDeposit_ValidAmount_IncreasesBalance() async throws {
        let account = BankAccount(accountNumber: "12345", initialBalance: 100)
        try await account.deposit(amount: 50)
        let balance = await account.balance
        XCTAssertEqual(balance, 150)
    }

    func testConcurrentDeposits_100Deposits_AllSucceed() async throws {
        let account = BankAccount(accountNumber: "12345", initialBalance: 0)

        await withTaskGroup(of: Void.self) { group in
            for _ in 1...100 {
                group.addTask {
                    try? await account.deposit(amount: 10)
                }
            }
        }

        let balance = await account.balance
        XCTAssertEqual(balance, 1000)
    }
}
```

**Key Features**:
- ✅ `actor` for thread-safe mutable state
- ✅ `Sendable` protocol for safe cross-actor communication
- ✅ `async`/`await` for asynchronous operations
- ✅ No data races (Swift 6 compiler verified)

---

## Example 4: TDD Workflow (RED → GREEN → REFACTOR)

### RED Phase: Write Failing Test

```swift
// @TEST:VALIDATOR-001 | SPEC: SPEC-VALIDATOR-001.md

import XCTest
@testable import MyApp

final class EmailValidatorTests: XCTestCase {
    var sut: EmailValidator!

    override func setUp() {
        super.setUp()
        sut = EmailValidator()
    }

    func testValidEmail_StandardFormat_ReturnsTrue() {
        XCTAssertTrue(sut.isValid("alice@example.com"))
    }

    func testInvalidEmail_MissingAt_ReturnsFalse() {
        XCTAssertFalse(sut.isValid("invalid-email"))
    }

    func testInvalidEmail_EmptyString_ReturnsFalse() {
        XCTAssertFalse(sut.isValid(""))
    }
}
```

**Run (should FAIL)**:
```bash
swift test --filter EmailValidatorTests
```

### GREEN Phase: Implement Feature

```swift
// @CODE:VALIDATOR-001 | SPEC: SPEC-VALIDATOR-001.md | TEST: Tests/EmailValidatorTests.swift

import Foundation

struct EmailValidator {
    func isValid(_ email: String) -> Bool {
        let emailRegex = "^[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}$"
        let emailPredicate = NSPredicate(format: "SELF MATCHES[c] %@", emailRegex)
        return emailPredicate.evaluate(with: email)
    }
}
```

**Run (should PASS)**:
```bash
swift test --filter EmailValidatorTests
```

### REFACTOR Phase: Improve Code

```swift
// @CODE:VALIDATOR-001 | SPEC: SPEC-VALIDATOR-001.md | TEST: Tests/EmailValidatorTests.swift

import Foundation

struct EmailValidator {
    private let emailRegex: NSRegularExpression

    init() {
        let pattern = "^[A-Z0-9a-z._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,64}$"
        // swiftlint:disable:next force_try
        self.emailRegex = try! NSRegularExpression(pattern: pattern, options: [.caseInsensitive])
    }

    func isValid(_ email: String) -> Bool {
        guard !email.isEmpty else { return false }
        let range = NSRange(location: 0, length: email.utf16.count)
        return emailRegex.firstMatch(in: email, options: [], range: range) != nil
    }

    func validate(_ email: String) -> ValidationResult {
        guard !email.isEmpty else {
            return .invalid(reason: "Email cannot be empty")
        }

        guard email.count <= 254 else {
            return .invalid(reason: "Email exceeds maximum length")
        }

        let range = NSRange(location: 0, length: email.utf16.count)
        guard emailRegex.firstMatch(in: email, options: [], range: range) != nil else {
            return .invalid(reason: "Email format is invalid")
        }

        return .valid
    }
}

enum ValidationResult {
    case valid
    case invalid(reason: String)
}
```

**Key Improvements**:
- ✅ Pre-compiled regex (performance)
- ✅ Detailed validation results
- ✅ Length validation
- ✅ Better documentation

---

## Example 5: CI/CD Integration

### GitHub Actions: `.github/workflows/swift-ci.yml`

```yaml
name: Swift CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install SwiftLint
        run: brew install swiftlint

      - name: Run SwiftLint
        run: swiftlint lint --strict

  test:
    runs-on: macos-latest
    strategy:
      matrix:
        platform: ['iOS', 'macOS']
    steps:
      - uses: actions/checkout@v4

      - name: Build and Test
        run: |
          if [ "${{ matrix.platform }}" == "iOS" ]; then
            xcodebuild test \
              -scheme MyApp \
              -destination 'platform=iOS Simulator,name=iPhone 16 Pro' \
              -enableCodeCoverage YES
          else
            swift test --enable-code-coverage
          fi

      - name: Upload Coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.lcov
```

---

## Example 6: Package.swift Configuration

```swift
// swift-tools-version: 6.0

import PackageDescription

let package = Package(
    name: "MyApp",
    platforms: [
        .iOS(.v17),
        .macOS(.v14)
    ],
    products: [
        .library(name: "MyApp", targets: ["MyApp"])
    ],
    dependencies: [
        .package(url: "https://github.com/apple/swift-log.git", from: "1.5.0")
    ],
    targets: [
        .target(
            name: "MyApp",
            dependencies: [
                .product(name: "Logging", package: "swift-log")
            ],
            swiftSettings: [
                .enableExperimentalFeature("StrictConcurrency")
            ]
        ),
        .testTarget(
            name: "MyAppTests",
            dependencies: ["MyApp"]
        )
    ]
)
```

---

## Summary

These examples demonstrate:
- ✅ **XCTest 6.0**: Modern testing with setUp/tearDown and async tests
- ✅ **SwiftLint 0.57.0**: Linting configuration and enforcement
- ✅ **Swift 6 Concurrency**: Actors, Sendable, async/await
- ✅ **TDD Workflow**: RED → GREEN → REFACTOR cycle
- ✅ **CI/CD Integration**: Automated testing and linting

**Next Steps**:
1. Review `reference.md` for detailed CLI commands
2. Integrate XCTest and SwiftLint into CI/CD
3. Enable strict concurrency checking in Swift 6
4. Follow TDD workflow for all new features

---

_For detailed reference documentation, see [reference.md](reference.md)_
_Last updated: 2025-10-22_
