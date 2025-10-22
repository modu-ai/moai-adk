# Swift CLI Reference

Quick reference for Swift 6.0, XCTest 6.0, SwiftLint 0.57.0, and iOS/macOS development tools.

---

## Tool Versions (2025-10-22)

| Tool | Version | Release Date | Purpose |
|------|---------|--------------|---------|
| **Swift** | 6.0.0 | 2024-09-16 | Primary language |
| **XCTest** | 6.0.0 | 2024-09-16 | Unit testing framework |
| **SwiftLint** | 0.57.0 | 2024-10-21 | Linting and style enforcement |
| **Xcode** | 16.0 | 2024-09-16 | IDE and build system |

---

## Swift 6.0

### Installation

```bash
# macOS (via Xcode)
xcode-select --install

# Linux (Ubuntu)
wget https://download.swift.org/swift-6.0-release/ubuntu2204/swift-6.0-RELEASE/swift-6.0-RELEASE-ubuntu22.04.tar.gz
tar xzf swift-6.0-RELEASE-ubuntu22.04.tar.gz
export PATH=/path/to/swift-6.0-RELEASE-ubuntu22.04/usr/bin:$PATH

# Verify installation
swift --version
# Expected: Swift version 6.0
```

### Common Commands

```bash
# Build project
swift build

# Build for release
swift build -c release

# Run executable
swift run

# Run specific target
swift run MyAppCLI

# Clean build artifacts
swift package clean

# Update dependencies
swift package update

# Resolve dependencies
swift package resolve

# Show dependencies
swift package show-dependencies
```

### Swift Package Manager

```bash
# Initialize new package
swift package init --type library
swift package init --type executable

# Add dependency
# Edit Package.swift manually

# Generate Xcode project
swift package generate-xcodeproj

# Describe package
swift package describe --type json
```

---

## XCTest 6.0

### Running Tests

```bash
# Run all tests
swift test

# Run specific test case
swift test --filter UserServiceTests

# Run specific test method
swift test --filter UserServiceTests/testUserCreation_ValidData_ReturnsUser

# Run tests with parallel execution
swift test --parallel

# Run tests with verbose output
swift test --verbose

# Generate code coverage
swift test --enable-code-coverage

# View coverage report
xcrun llvm-cov show .build/debug/MyAppPackageTests.xctest/Contents/MacOS/MyAppPackageTests \
  --instr-profile .build/debug/codecov/default.profdata \
  --use-color

# Export coverage to HTML
xcrun llvm-cov show .build/debug/MyAppPackageTests.xctest/Contents/MacOS/MyAppPackageTests \
  --instr-profile .build/debug/codecov/default.profdata \
  --format=html > coverage.html

# Export coverage to LCOV format
xcrun llvm-cov export .build/debug/MyAppPackageTests.xctest/Contents/MacOS/MyAppPackageTests \
  --instr-profile .build/debug/codecov/default.profdata \
  --format=lcov > coverage.lcov
```

### Test Assertions

```swift
// Equality
XCTAssertEqual(actual, expected)
XCTAssertNotEqual(actual, expected)

// Boolean
XCTAssertTrue(condition)
XCTAssertFalse(condition)

// Nil checking
XCTAssertNil(value)
XCTAssertNotNil(value)

// Comparison
XCTAssertGreaterThan(value1, value2)
XCTAssertLessThan(value1, value2)
XCTAssertGreaterThanOrEqual(value1, value2)
XCTAssertLessThanOrEqual(value1, value2)

// Error handling
XCTAssertThrowsError(try someFunction())
XCTAssertNoThrow(try someFunction())

// Failure
XCTFail("Explicit failure message")

// Async assertions
await XCTAssertAsyncEqual(await asyncValue(), expectedValue)
```

---

## SwiftLint 0.57.0

### Installation

```bash
# Homebrew
brew install swiftlint

# CocoaPods (add to Podfile)
pod 'SwiftLint'

# Mint
mint install realm/SwiftLint

# Download binary
curl -L https://github.com/realm/SwiftLint/releases/download/0.57.0/swiftlint_darwin.zip -o swiftlint.zip
unzip swiftlint.zip
mv swiftlint /usr/local/bin/

# Verify installation
swiftlint version
# Expected: 0.57.0
```

### Common Commands

```bash
# Lint all files
swiftlint

# Lint specific file
swiftlint lint --path Sources/UserService.swift

# Lint with auto-fix
swiftlint --fix

# Lint with strict mode (warnings as errors)
swiftlint lint --strict

# Generate configuration file
swiftlint generate-docs > .swiftlint.yml

# Show rules
swiftlint rules

# Lint with specific configuration
swiftlint lint --config custom-config.yml

# Lint and output to file
swiftlint lint > swiftlint-output.txt

# Lint with reporter
swiftlint lint --reporter json
swiftlint lint --reporter html > report.html
swiftlint lint --reporter checkstyle > checkstyle.xml
```

### Configuration File: `.swiftlint.yml`

```yaml
# SwiftLint 0.57.0 Configuration

included:
  - Sources
  - Tests

excluded:
  - Pods
  - .build
  - DerivedData
  - .swiftpm

disabled_rules:
  - trailing_whitespace
  - todo
  - line_length

opt_in_rules:
  - empty_count
  - explicit_init
  - file_header
  - first_where
  - force_unwrapping
  - sorted_imports
  - vertical_whitespace_closing_braces
  - yoda_condition

line_length:
  warning: 120
  error: 150
  ignores_comments: true
  ignores_urls: true

file_length:
  warning: 400
  error: 500

function_body_length:
  warning: 50
  error: 100

type_body_length:
  warning: 300
  error: 500

cyclomatic_complexity:
  warning: 10
  error: 20

identifier_name:
  min_length:
    warning: 3
  max_length:
    warning: 40
    error: 50
  excluded:
    - id
    - url
    - db

custom_rules:
  no_print:
    regex: 'print\('
    message: "Use Logger instead of print()"
    severity: warning
```

### Common Rules

| Rule | Description | Auto-fix |
|------|-------------|----------|
| **force_cast** | Avoid force casts (`as!`) | ✅ |
| **force_unwrapping** | Avoid force unwrapping (`!`) | ❌ |
| **line_length** | Line length limit (default 120) | ❌ |
| **trailing_whitespace** | Remove trailing whitespace | ✅ |
| **opening_brace** | Opening brace spacing | ✅ |
| **colon** | Colon spacing | ✅ |
| **comma** | Comma spacing | ✅ |
| **empty_count** | Use `isEmpty` instead of `count == 0` | ✅ |
| **sorted_imports** | Alphabetically sort imports | ✅ |
| **unused_optional_binding** | Remove unused optional bindings | ✅ |

---

## Xcode Build Commands

### Command Line Build

```bash
# Build iOS app
xcodebuild build \
  -scheme MyApp \
  -destination 'platform=iOS Simulator,name=iPhone 16 Pro'

# Run tests
xcodebuild test \
  -scheme MyApp \
  -destination 'platform=iOS Simulator,name=iPhone 16 Pro' \
  -enableCodeCoverage YES

# Build macOS app
xcodebuild build \
  -scheme MyApp \
  -destination 'platform=macOS'

# Archive for distribution
xcodebuild archive \
  -scheme MyApp \
  -archivePath ./build/MyApp.xcarchive

# Export archive
xcodebuild -exportArchive \
  -archivePath ./build/MyApp.xcarchive \
  -exportPath ./build/export \
  -exportOptionsPlist ExportOptions.plist

# Clean build folder
xcodebuild clean \
  -scheme MyApp
```

### Available Simulators

```bash
# List available simulators
xcrun simctl list devices

# Boot simulator
xcrun simctl boot "iPhone 16 Pro"

# Install app on simulator
xcrun simctl install booted ./build/MyApp.app

# Launch app on simulator
xcrun simctl launch booted com.example.MyApp

# Uninstall app from simulator
xcrun simctl uninstall booted com.example.MyApp

# Erase simulator
xcrun simctl erase "iPhone 16 Pro"
```

---

## Swift 6 Concurrency

### Actor Syntax

```swift
// Define actor
actor BankAccount {
    private(set) var balance: Decimal

    func deposit(amount: Decimal) {
        balance += amount
    }

    func withdraw(amount: Decimal) throws {
        guard balance >= amount else {
            throw BankAccountError.insufficientFunds
        }
        balance -= amount
    }
}

// Use actor
let account = BankAccount(initialBalance: 1000)
try await account.withdraw(amount: 100)
let balance = await account.balance
```

### Sendable Protocol

```swift
// Sendable structs (value types)
struct User: Sendable {
    let id: Int
    let name: String
}

// Sendable actors
actor DataStore: Sendable {
    // ...
}

// Sendable closures
let closure: @Sendable () -> Void = {
    print("Thread-safe closure")
}
```

### Async/Await

```swift
// Async function
func fetchUser(id: Int) async throws -> User {
    // Async operation
}

// Call async function
let user = try await fetchUser(id: 1)

// Concurrent execution
async let user1 = fetchUser(id: 1)
async let user2 = fetchUser(id: 2)
let users = try await [user1, user2]
```

### Task Groups

```swift
// Structured concurrency
try await withThrowingTaskGroup(of: User.self) { group in
    for userId in 1...10 {
        group.addTask {
            try await fetchUser(id: userId)
        }
    }

    var users: [User] = []
    for try await user in group {
        users.append(user)
    }
    return users
}
```

---

## TRUST 5 Principles for Swift

### T - Test First

```swift
// Write tests before implementation (XCTest)
func testUserCreation_ValidData_ReturnsUser() throws {
    let sut = UserService()
    let user = try sut.createUser(email: "alice@example.com", username: "alice")
    XCTAssertEqual(user.email, "alice@example.com")
}
```

### R - Readable

```swift
// Use SwiftLint for consistent formatting
swiftlint --fix

// Clear naming conventions
class UserService { }
func fetchUser(id: Int) -> User? { }
let userEmail = "alice@example.com"

// MARK comments for organization
// MARK: - Properties
// MARK: - Initialization
// MARK: - Public Methods
```

### U - Unified

```swift
// Consistent project structure
Sources/
  MyApp/
    Models/
    Services/
    Views/
Tests/
  MyAppTests/
    ModelTests/
    ServiceTests/

// Type safety with protocols
protocol DatabaseProtocol {
    func insert(_ user: User) throws
}
```

### S - Secured

```swift
// Avoid force unwrapping
// Bad: let value = optional!
// Good: guard let value = optional else { return }

// Input validation
func createUser(email: String) throws -> User {
    guard !email.isEmpty else {
        throw ValidationError.emptyEmail
    }
    // ...
}

// Sensitive data handling
// Use Keychain for storing credentials
```

### T - Trackable

```swift
// @TAG markers in code
// @CODE:USER-001 | SPEC: SPEC-USER-001.md | TEST: Tests/UserServiceTests.swift

// Version tracking in Git
// Follow semantic versioning

// Documentation
/// Creates a new user with the provided email
/// - Parameter email: User's email address
/// - Returns: Created user instance
/// - Throws: `UserServiceError` if validation fails
func createUser(email: String) throws -> User { }
```

---

## Best Practices

### Code Style

```swift
// Use guard for early returns
func process(_ user: User?) {
    guard let user = user else { return }
    // ...
}

// Prefer map/filter/reduce over loops
let emails = users.map { $0.email }
let active = users.filter { $0.isActive }

// Use trailing closures
UIView.animate(withDuration: 0.3) {
    view.alpha = 0
}

// Avoid force unwrapping
// Bad: let value = optional!
// Good: if let value = optional { }
```

### Memory Management

```swift
// Use weak references in closures
class ViewController {
    func setup() {
        service.fetch { [weak self] result in
            self?.updateUI(with: result)
        }
    }
}

// Avoid retain cycles with unowned
class Node {
    unowned let parent: Node
}
```

### Error Handling

```swift
// Define custom errors
enum UserServiceError: Error {
    case invalidEmail
    case userNotFound
}

// Use throws
func createUser() throws -> User { }

// Handle errors
do {
    let user = try createUser()
} catch UserServiceError.invalidEmail {
    // Handle invalid email
} catch {
    // Handle other errors
}
```

---

## Troubleshooting

### Common Issues

```bash
# Build failures
# Solution: Clean build folder
swift package clean
rm -rf .build

# Dependency resolution issues
# Solution: Reset package cache
swift package reset
swift package update

# SwiftLint crashes
# Solution: Update to latest version
brew upgrade swiftlint

# Xcode build hangs
# Solution: Restart Xcode and clean derived data
rm -rf ~/Library/Developer/Xcode/DerivedData
```

### Debugging Commands

```bash
# Print build settings
xcodebuild -showBuildSettings

# Check Xcode version
xcodebuild -version

# List schemes
xcodebuild -list

# Verbose build
swift build -v

# Check package dependencies
swift package show-dependencies
```

---

## Additional Resources

- **Swift Docs**: https://docs.swift.org/swift-book/
- **Swift Evolution**: https://github.com/apple/swift-evolution
- **SwiftLint**: https://github.com/realm/SwiftLint
- **Apple Developer**: https://developer.apple.com/documentation/xcode
- **Swift Forums**: https://forums.swift.org/

---

_Last updated: 2025-10-22_
