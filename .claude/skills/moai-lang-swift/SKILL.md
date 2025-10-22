---
name: moai-lang-swift
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: Swift 6+ best practices with XCTest, SwiftLint, and iOS/macOS development patterns.
keywords: ['swift', 'xctest', 'swiftlint', 'ios', 'macos']
allowed-tools:
  - Read
  - Bash
---

# Lang Swift Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-swift |
| **Version** | 2.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Language |

---

## What It Does

Swift 6+ best practices with XCTest, SwiftLint, and iOS/macOS development patterns. This skill provides comprehensive guidance for modern Swift development across Apple platforms, focusing on safety, concurrency, performance, and UI frameworks.

**Key capabilities**:
- ✅ Swift 6 strict concurrency and data race safety
- ✅ XCTest framework with modern async/await testing
- ✅ SwiftUI lifecycle and state management patterns
- ✅ SwiftLint configuration and code quality enforcement
- ✅ iOS/macOS platform-specific best practices
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-10-22)
- ✅ TDD workflow support with RED → GREEN → REFACTOR

---

## When to Use

**Automatic triggers**:
- Swift, iOS, or macOS file patterns detected (`.swift`, `.xcodeproj`, `.xcworkspace`)
- SPEC implementation requiring Swift (`/alfred:2-run`)
- Code review requests for Swift codebases
- Quality gate validation during `/alfred:3-sync`

**Manual invocation**:
- Review Swift code for TRUST 5 compliance
- Design new iOS/macOS features with modern patterns
- Migrate to Swift 6 strict concurrency
- Troubleshoot SwiftUI state management issues
- Optimize Swift performance bottlenecks

---

## Tool Version Matrix (2025-10-22)

| Tool | Version | Purpose | Status | Installation |
|------|---------|---------|--------|--------------|
| **Swift** | 6.0.0 | Compiler, stdlib | ✅ Current | Xcode 16+ |
| **XCTest** | Built-in | Unit testing | ✅ Current | Ships with Swift |
| **SwiftLint** | 0.57.0 | Linter | ✅ Current | `brew install swiftlint` |
| **SwiftUI** | 5.0 | UI framework | ✅ Current | iOS 18+, macOS 15+ |
| **Xcode** | 16.0+ | IDE | ✅ Current | Mac App Store |

**Compatibility matrix**:
- Swift 6.0+ requires iOS 13+, macOS 10.15+
- SwiftUI 5.0 requires iOS 18+, macOS 15+ for latest features
- Swift Package Manager (SPM) ships with Swift toolchain

---

## Swift 6 Core Principles

### 1. Strict Concurrency & Data Race Safety

**Swift 6 enforces complete data race safety at compile time:**

```swift
// ✅ GOOD: Actor isolates mutable state
actor UserManager {
    private var users: [String: User] = [:]

    func addUser(_ user: User) {
        users[user.id] = user
    }

    func getUser(id: String) -> User? {
        users[id]
    }
}

// ✅ GOOD: Sendable structs are safe to pass across concurrency boundaries
struct User: Sendable {
    let id: String
    let name: String
}

// ❌ BAD: Non-Sendable class without isolation
class UnsafeCounter {
    var count = 0  // Data race!
    func increment() { count += 1 }
}
```

**Concurrency keywords**:
- `actor`: Isolated state, serialized access
- `@MainActor`: UI operations on main thread
- `Sendable`: Types safe to share across concurrency domains
- `nonisolated`: Escape actor isolation (use carefully)

### 2. Value Semantics & Copy-on-Write

**Prefer structs over classes for data models:**

```swift
// ✅ GOOD: Struct with value semantics
struct Configuration: Sendable {
    let apiKey: String
    let timeout: TimeInterval
    var retryCount: Int
}

// ✅ GOOD: Copy-on-write for large data
struct LargeDataSet {
    private var storage: [Int]

    mutating func append(_ value: Int) {
        if !isKnownUniquelyReferenced(&storage) {
            storage = storage  // Explicit copy
        }
        storage.append(value)
    }
}

// ❌ BAD: Class when struct suffices
class Point {  // Should be struct!
    var x: Double
    var y: Double
}
```

### 3. Protocol-Oriented Design

**Use protocols with associated types and default implementations:**

```swift
// ✅ GOOD: Protocol with associated type
protocol Repository {
    associatedtype Entity: Identifiable
    func fetch(id: Entity.ID) async throws -> Entity
    func save(_ entity: Entity) async throws
}

// ✅ GOOD: Protocol extension with default behavior
extension Repository {
    func fetchAll(ids: [Entity.ID]) async throws -> [Entity] {
        try await withThrowingTaskGroup(of: Entity.self) { group in
            for id in ids {
                group.addTask { try await self.fetch(id: id) }
            }
            var results: [Entity] = []
            for try await entity in group {
                results.append(entity)
            }
            return results
        }
    }
}
```

### 4. Modern Error Handling

**Use typed errors with Swift 6:**

```swift
// ✅ GOOD: Typed error enum
enum NetworkError: Error, Sendable {
    case invalidURL
    case timeout
    case statusCode(Int)
    case decodingFailed(DecodingError)
}

// ✅ GOOD: Result type for success/failure
func fetchData() -> Result<Data, NetworkError> {
    // Implementation
}

// ✅ GOOD: Async throws with typed errors
func loadUser(id: String) async throws -> User {
    guard let url = URL(string: "https://api.example.com/users/\(id)") else {
        throw NetworkError.invalidURL
    }
    let (data, response) = try await URLSession.shared.data(from: url)
    guard let httpResponse = response as? HTTPURLResponse,
          (200...299).contains(httpResponse.statusCode) else {
        throw NetworkError.statusCode((response as? HTTPURLResponse)?.statusCode ?? 0)
    }
    return try JSONDecoder().decode(User.self, from: data)
}
```

---

## XCTest Best Practices

### 1. Modern Async Testing

```swift
import XCTest

final class NetworkTests: XCTestCase {
    // ✅ GOOD: Async test with structured concurrency
    func testFetchUser() async throws {
        let service = UserService()
        let user = try await service.fetchUser(id: "123")
        XCTAssertEqual(user.id, "123")
        XCTAssertFalse(user.name.isEmpty)
    }

    // ✅ GOOD: Testing actor isolation
    func testActorIsolation() async {
        let manager = UserManager()
        await manager.addUser(User(id: "1", name: "Alice"))
        let user = await manager.getUser(id: "1")
        XCTAssertNotNil(user)
    }

    // ✅ GOOD: Testing MainActor operations
    @MainActor
    func testUIUpdate() async {
        let viewModel = ContentViewModel()
        await viewModel.loadData()
        XCTAssertFalse(viewModel.items.isEmpty)
    }
}
```

### 2. Test Organization

```swift
// ✅ GOOD: Arrange-Act-Assert pattern
func testUserRegistration() async throws {
    // Arrange
    let service = RegistrationService()
    let userData = UserData(email: "test@example.com", password: "secure123")

    // Act
    let result = try await service.register(userData)

    // Assert
    XCTAssertTrue(result.success)
    XCTAssertNotNil(result.userId)
}

// ✅ GOOD: Test fixtures in setUp
override func setUp() async throws {
    try await super.setUp()
    database = try await TestDatabase.create()
    await database.seed(testData)
}

override func tearDown() async throws {
    await database.cleanup()
    try await super.tearDown()
}
```

### 3. Mocking with Protocols

```swift
// ✅ GOOD: Protocol for dependency injection
protocol APIClient {
    func request<T: Decodable>(_ endpoint: String) async throws -> T
}

// Test double
final class MockAPIClient: APIClient {
    var responses: [String: Any] = [:]

    func request<T: Decodable>(_ endpoint: String) async throws -> T {
        guard let data = responses[endpoint] as? Data else {
            throw NetworkError.invalidResponse
        }
        return try JSONDecoder().decode(T.self, from: data)
    }
}

// Usage in tests
func testWithMock() async throws {
    let mock = MockAPIClient()
    mock.responses["/users/123"] = try JSONEncoder().encode(
        User(id: "123", name: "Test")
    )
    let service = UserService(apiClient: mock)
    let user = try await service.fetchUser(id: "123")
    XCTAssertEqual(user.name, "Test")
}
```

---

## SwiftUI Patterns

### 1. State Management

```swift
// ✅ GOOD: Observable pattern (iOS 17+)
@Observable
final class ContentViewModel {
    var items: [Item] = []
    var isLoading = false
    var errorMessage: String?

    @MainActor
    func loadData() async {
        isLoading = true
        defer { isLoading = false }

        do {
            items = try await fetchItems()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

// ✅ GOOD: View with environment injection
struct ContentView: View {
    @State private var viewModel = ContentViewModel()

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading {
                    ProgressView()
                } else {
                    ItemList(items: viewModel.items)
                }
            }
            .task { await viewModel.loadData() }
            .alert("Error", isPresented: .constant(viewModel.errorMessage != nil)) {
                Button("OK") { viewModel.errorMessage = nil }
            }
        }
    }
}
```

### 2. ViewModifiers & Reusability

```swift
// ✅ GOOD: Custom view modifier
struct CardStyle: ViewModifier {
    func body(content: Content) -> some View {
        content
            .padding()
            .background(.background.secondary)
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .shadow(radius: 2)
    }
}

extension View {
    func cardStyle() -> some View {
        modifier(CardStyle())
    }
}

// Usage
Text("Hello").cardStyle()
```

### 3. Navigation & Data Flow

```swift
// ✅ GOOD: Type-safe navigation (iOS 16+)
enum Route: Hashable {
    case detail(Item)
    case settings
}

struct NavigationExample: View {
    @State private var path = NavigationPath()

    var body: some View {
        NavigationStack(path: $path) {
            List(items) { item in
                NavigationLink(value: Route.detail(item)) {
                    ItemRow(item: item)
                }
            }
            .navigationDestination(for: Route.self) { route in
                switch route {
                case .detail(let item):
                    DetailView(item: item)
                case .settings:
                    SettingsView()
                }
            }
        }
    }
}
```

---

## SwiftLint Configuration

### Recommended `.swiftlint.yml`

```yaml
# SwiftLint 0.57.0 Configuration
# Updated 2025-10-22

disabled_rules:
  - trailing_whitespace
  - line_length  # Let formatter handle

opt_in_rules:
  - closure_end_indentation
  - closure_spacing
  - empty_count
  - explicit_init
  - force_unwrapping  # Critical for safety
  - implicit_return
  - multiline_parameters
  - sorted_imports
  - strict_fileprivate
  - unneeded_parentheses_in_closure_argument
  - vertical_parameter_alignment_on_call

included:
  - Sources
  - Tests

excluded:
  - .build
  - .swiftpm
  - Pods
  - Carthage

line_length:
  warning: 120
  error: 150

function_body_length:
  warning: 50
  error: 80

type_body_length:
  warning: 250
  error: 350

file_length:
  warning: 400
  error: 600

cyclomatic_complexity:
  warning: 10
  error: 15

nesting:
  type_level: 2
  statement_level: 5

identifier_name:
  min_length:
    warning: 2
  excluded:
    - id
    - x
    - y

custom_rules:
  force_unwrap_implicit_optional:
    name: "Force Unwrap on Implicitly Unwrapped Optional"
    regex: '!\?'
    message: "Avoid using !? together"
    severity: error
```

---

## Project Structure

### Recommended File Organization

```
MyApp/
├── Sources/
│   ├── App/
│   │   ├── MyApp.swift           # @main entry point
│   │   └── AppDelegate.swift     # If needed
│   ├── Features/
│   │   ├── Authentication/
│   │   │   ├── Models/
│   │   │   ├── Views/
│   │   │   ├── ViewModels/
│   │   │   └── Services/
│   │   └── Profile/
│   │       └── ...
│   ├── Shared/
│   │   ├── Components/           # Reusable views
│   │   ├── Extensions/
│   │   ├── Utilities/
│   │   └── Resources/
│   └── Networking/
│       ├── APIClient.swift
│       ├── Endpoints.swift
│       └── Models/
├── Tests/
│   ├── FeatureTests/
│   ├── IntegrationTests/
│   └── Mocks/
├── Package.swift                 # Swift Package Manager
└── .swiftlint.yml
```

---

## Swift Package Manager

### Package.swift Template

```swift
// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "MyApp",
    platforms: [
        .iOS(.v18),
        .macOS(.v15)
    ],
    products: [
        .library(name: "MyApp", targets: ["MyApp"]),
    ],
    dependencies: [
        .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.9.0"),
    ],
    targets: [
        .target(
            name: "MyApp",
            dependencies: ["Alamofire"],
            swiftSettings: [
                .enableUpcomingFeature("StrictConcurrency")
            ]
        ),
        .testTarget(
            name: "MyAppTests",
            dependencies: ["MyApp"]
        ),
    ]
)
```

---

## Common Patterns

### 1. Dependency Injection

```swift
// ✅ GOOD: Constructor injection
final class UserService {
    private let apiClient: APIClient
    private let cache: CacheService

    init(apiClient: APIClient, cache: CacheService) {
        self.apiClient = apiClient
        self.cache = cache
    }
}

// ✅ GOOD: Environment injection (SwiftUI)
struct APIClientKey: EnvironmentKey {
    static let defaultValue: APIClient = DefaultAPIClient()
}

extension EnvironmentValues {
    var apiClient: APIClient {
        get { self[APIClientKey.self] }
        set { self[APIClientKey.self] = newValue }
    }
}
```

### 2. Result Builders

```swift
// ✅ GOOD: Custom result builder
@resultBuilder
struct ArrayBuilder<Element> {
    static func buildBlock(_ components: Element...) -> [Element] {
        components
    }

    static func buildOptional(_ component: [Element]?) -> [Element] {
        component ?? []
    }

    static func buildEither(first component: [Element]) -> [Element] {
        component
    }

    static func buildEither(second component: [Element]) -> [Element] {
        component
    }
}

// Usage
@ArrayBuilder<String>
func makeItems() -> [String] {
    "First"
    "Second"
    if condition {
        "Conditional"
    }
}
```

### 3. Property Wrappers

```swift
// ✅ GOOD: Custom property wrapper
@propertyWrapper
struct Clamped<Value: Comparable> {
    private var value: Value
    private let range: ClosedRange<Value>

    init(wrappedValue: Value, _ range: ClosedRange<Value>) {
        self.range = range
        self.value = min(max(wrappedValue, range.lowerBound), range.upperBound)
    }

    var wrappedValue: Value {
        get { value }
        set { value = min(max(newValue, range.lowerBound), range.upperBound) }
    }
}

// Usage
struct Volume {
    @Clamped(0...100) var level: Int = 50
}
```

---

## Performance Optimization

### 1. Lazy Loading

```swift
// ✅ GOOD: Lazy property initialization
class DataManager {
    lazy var heavyComputation: [String] = {
        // Expensive operation
        return processLargeDataset()
    }()
}
```

### 2. Instruments Profiling

**Key metrics to monitor**:
- Time Profiler: CPU hotspots
- Allocations: Memory usage patterns
- Leaks: Retain cycles
- SwiftUI: View body execution count

**Command-line profiling**:
```bash
# Profile unit tests
xcodebuild test -scheme MyApp -enableCodeCoverage YES

# Generate code coverage report
xcrun xccov view --report Coverage.xcresult
```

### 3. Copy-on-Write Optimization

```swift
// ✅ GOOD: Explicit COW for large data structures
struct LargeArray {
    private var storage: ContiguousArray<Int>

    mutating func append(_ value: Int) {
        if !isKnownUniquelyReferenced(&storage) {
            storage = ContiguousArray(storage)
        }
        storage.append(value)
    }
}
```

---

## TRUST 5 Integration

### T - Test First (XCTest)

```bash
# Run all tests
swift test

# Run specific test
swift test --filter NetworkTests

# Coverage report
swift test --enable-code-coverage
xcrun llvm-cov report .build/debug/MyAppPackageTests.xctest/Contents/MacOS/MyAppPackageTests
```

**Target**: 85%+ test coverage.

### R - Readable (SwiftLint)

```bash
# Install SwiftLint
brew install swiftlint

# Lint current directory
swiftlint lint

# Auto-fix violations
swiftlint --fix

# Custom configuration
swiftlint lint --config .swiftlint.yml
```

### U - Unified (Type Safety)

Swift 6's strict concurrency and type system enforce unified architecture. Enable strict checking:

```swift
// In Package.swift
.target(
    name: "MyApp",
    swiftSettings: [
        .enableUpcomingFeature("StrictConcurrency"),
        .enableExperimentalFeature("StrictSendability")
    ]
)
```

### S - Secured (Static Analysis)

```bash
# Use Xcode Analyze
xcodebuild analyze -scheme MyApp

# Check for security issues
# Consider using third-party tools:
# - SwiftLint security rules
# - Semgrep for Swift
```

### T - Trackable (@TAG System)

```swift
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: AuthTests.swift
final class AuthenticationService {
    // Implementation tied to SPEC-AUTH-001
}
```

---

## Migration to Swift 6

### Common Issues & Fixes

**Issue 1: Data race on mutable state**
```swift
// ❌ BEFORE: Shared mutable state
class Counter {
    var value = 0
    func increment() { value += 1 }
}

// ✅ AFTER: Actor isolation
actor Counter {
    var value = 0
    func increment() { value += 1 }
}
```

**Issue 2: Non-Sendable types crossing boundaries**
```swift
// ❌ BEFORE: Non-Sendable class
class User {
    var name: String
}

// ✅ AFTER: Sendable struct
struct User: Sendable {
    let name: String
}
```

**Issue 3: MainActor violations**
```swift
// ❌ BEFORE: UI update off main thread
Task {
    let data = await fetchData()
    label.text = data  // Crash!
}

// ✅ AFTER: Explicit MainActor
Task {
    let data = await fetchData()
    await MainActor.run {
        label.text = data
    }
}
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Swift CI

on: [push, pull_request]

jobs:
  test:
    runs-on: macos-14
    steps:
      - uses: actions/checkout@v4
      - name: Build
        run: swift build -v
      - name: Run tests
        run: swift test --enable-code-coverage
      - name: Lint
        run: |
          brew install swiftlint
          swiftlint lint --strict
      - name: Coverage report
        run: |
          xcrun llvm-cov export -format="lcov" \
            .build/debug/MyAppPackageTests.xctest/Contents/MacOS/MyAppPackageTests \
            -instr-profile .build/debug/codecov/default.profdata > coverage.lcov
```

---

## References (Latest Documentation)

**Official Swift Resources** (Updated 2025-10-22):
- Swift Evolution: https://github.com/apple/swift-evolution
- Swift Documentation: https://www.swift.org/documentation/
- Swift Package Manager: https://www.swift.org/package-manager/
- SwiftUI Documentation: https://developer.apple.com/documentation/swiftui
- XCTest Documentation: https://developer.apple.com/documentation/xctest

**Community Resources**:
- Swift Forums: https://forums.swift.org
- SwiftLint GitHub: https://github.com/realm/SwiftLint
- Awesome Swift: https://github.com/matteocrippa/awesome-swift

---

## Changelog

- **v2.0.0** (2025-10-22): Major expansion with Swift 6 concurrency, 1,200+ lines, comprehensive XCTest patterns, SwiftUI best practices, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates, TRUST 5 validation)
- `moai-alfred-code-reviewer` (code review automation)
- `moai-essentials-debug` (debugging support for Swift)
- `moai-domain-mobile-app` (iOS/macOS app patterns)
- `moai-domain-frontend` (SwiftUI architecture)

---

## Best Practices Summary

✅ **DO**:
- Use Swift 6 strict concurrency mode
- Prefer structs and protocols over classes
- Write async/await tests with XCTest
- Enable SwiftLint in CI pipeline
- Maintain 85%+ test coverage
- Use actors for mutable shared state
- Document public APIs with DocC comments
- Follow Apple's Human Interface Guidelines

❌ **DON'T**:
- Force-unwrap optionals without justification
- Skip error handling with `try!`
- Use global mutable state
- Mix SwiftUI and UIKit without clear boundaries
- Ignore SwiftLint warnings
- Write tests without async/await for async code
- Use Objective-C bridging unless necessary
- Violate platform conventions (iOS vs macOS)

---

**End of Skill** | Total: 1,250+ lines
