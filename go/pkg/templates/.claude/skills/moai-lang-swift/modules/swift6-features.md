# Swift 6.0 Features

Comprehensive guide to Swift 6.0 language features including typed throws, complete concurrency, and data-race safety.

## Typed Throws

Swift 6 introduces typed throws allowing precise error type specification in function signatures.

### Basic Typed Throws

Error Type Specification:
```swift
enum NetworkError: Error {
    case invalidURL
    case requestFailed(statusCode: Int)
    case decodingFailed
}

// Function declares specific error type
func fetchData(from urlString: String) throws(NetworkError) -> Data {
    guard let url = URL(string: urlString) else {
        throw .invalidURL
    }
    // Implementation
}

// Caller has exhaustive error handling
do {
    let data = try fetchData(from: "https://api.example.com")
} catch .invalidURL {
    print("Invalid URL")
} catch .requestFailed(let code) {
    print("Request failed: \(code)")
} catch .decodingFailed {
    print("Decoding failed")
}
```

### Domain-Specific Error Types

Authentication Errors:
```swift
enum AuthError: Error, LocalizedError {
    case invalidCredentials
    case sessionExpired
    case networkError(underlying: Error)
    
    var errorDescription: String? {
        switch self {
        case .invalidCredentials: return "Invalid email or password"
        case .sessionExpired: return "Session expired. Please login again."
        case .networkError(let error): return "Network error: \(error.localizedDescription)"
        }
    }
}

protocol AuthServiceProtocol: Sendable {
    func login(email: String, password: String) async throws(AuthError) -> AuthTokens
    func refreshToken(_ token: String) async throws(AuthError) -> AuthTokens
}
```

## Complete Concurrency Checking

Swift 6 enforces complete data-race safety by default at compile time.

### Sendable Requirements

Value Types (Automatic Sendable):
```swift
// Structs with only Sendable properties are automatically Sendable
struct User: Codable, Identifiable, Sendable {
    let id: String
    let name: String
    let email: String
}
```

Reference Types (Explicit Sendable):
```swift
// Classes must explicitly conform and be final
final class ImmutableConfig: Sendable {
    let apiKey: String
    let baseURL: URL
    
    init(apiKey: String, baseURL: URL) {
        self.apiKey = apiKey
        self.baseURL = baseURL
    }
}

// Use @unchecked Sendable for thread-safe classes
final class ThreadSafeCache<Key: Hashable, Value>: @unchecked Sendable {
    private var cache: [Key: Value] = [:]
    private let lock = NSLock()
    
    func get(_ key: Key) -> Value? {
        lock.lock()
        defer { lock.unlock() }
        return cache[key]
    }
    
    func set(_ key: Key, value: Value) {
        lock.lock()
        defer { lock.unlock() }
        cache[key] = value
    }
}
```

### Actor Isolation

Protecting Mutable State:
```swift
actor UserCache {
    private var cache: [String: User] = [:]
    
    func get(_ id: String) -> User? { cache[id] }
    func set(_ id: String, user: User) { cache[id] = user }
    func clear() { cache.removeAll() }
    
    var isEmpty: Bool { cache.isEmpty }
}
```

### MainActor for UI Code

ViewModel with MainActor:
```swift
@MainActor
@Observable
final class ProfileViewModel {
    private(set) var user: User?
    private(set) var isLoading = false
    
    private let api: UserAPIProtocol
    
    func loadProfile(_ userId: String) async {
        isLoading = true
        defer { isLoading = false }
        user = try? await api.fetchUser(userId)
    }
}
```

### Nonisolated Functions

Escaping Actor Isolation:
```swift
actor NetworkMonitor {
    private var isConnected = true
    
    func updateStatus(connected: Bool) {
        isConnected = connected
    }
    
    // Nonisolated - can be called synchronously
    nonisolated func describeConnection() -> String {
        "Network monitor for connectivity tracking"
    }
    
    // Nonisolated with async for cross-actor reads
    nonisolated func getStatus() async -> Bool {
        await isConnected
    }
}
```

## Data-Race Safety Patterns

### Safe Alternatives

Using Actor:
```swift
actor SafeCounter {
    private var count = 0
    
    func increment() -> Int {
        count += 1
        return count
    }
}
```

Using AsyncStream for Events:
```swift
actor EventEmitter<Event: Sendable> {
    private var continuations: [UUID: AsyncStream<Event>.Continuation] = [:]
    
    func subscribe() -> AsyncStream<Event> {
        let id = UUID()
        return AsyncStream { continuation in
            continuations[id] = continuation
            continuation.onTermination = { [weak self] _ in
                Task { await self?.removeContinuation(id) }
            }
        }
    }
    
    private func removeContinuation(_ id: UUID) {
        continuations.removeValue(forKey: id)
    }
    
    func emit(_ event: Event) {
        for continuation in continuations.values {
            continuation.yield(event)
        }
    }
}
```

### Global Actor Pattern

Custom Global Actor:
```swift
@globalActor
actor DatabaseActor: GlobalActor {
    static let shared = DatabaseActor()
}

@DatabaseActor
final class DatabaseManager {
    private var connection: DatabaseConnection?
    
    func connect() async throws {
        connection = try await DatabaseConnection.open()
    }
}
```

## Migration Guidelines

### Enabling Strict Concurrency

Package.swift Settings:
```swift
targets: [
    .target(
        name: "MyApp",
        swiftSettings: [
            .enableExperimentalFeature("StrictConcurrency")
        ]
    )
]
```

### Common Migration Patterns

Callback to Async:
```swift
// Before: Callback-based
func fetchUser(id: String, completion: @escaping (Result<User, Error>) -> Void)

// After: Async/await
func fetchUser(id: String) async throws -> User {
    let (data, _) = try await URLSession.shared.data(from: url)
    return try JSONDecoder().decode(User.self, from: data)
}
```

---

Version: 2.0.0
Last Updated: 2026-01-06
Context7: /apple/swift, /apple/swift-evolution
