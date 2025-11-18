---
name: moai-lang-swift
version: 4.0.0
updated: '2025-11-19'
status: stable
stability: stable
description: Swift 6.0 enterprise development with async/await, SwiftUI, Combine,
allowed-tools:
- Read
- Bash
- WebSearch
- WebFetch
---



# Swift - Enterprise 

## Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-swift |
| **Version** | 4.0.0 (2025-11-12) |
| **Allowed tools** | Read, Bash, Context7 MCP |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Language Enterprise |
| **Context7 Integration** | ✅ Swift/SwiftUI/Vapor/Combine |

---

## What It Does

Swift 6.0 enterprise development featuring modern concurrency with async/await, SwiftUI for declarative UI, Combine for reactive programming, server-side Swift with Vapor, and enterprise-grade patterns for scalable, performant applications. Context7 MCP integration provides real-time access to official Swift and ecosystem documentation.

**Key capabilities**:
- ✅ Swift 6.0 with strict concurrency and actor isolation
- ✅ Advanced async/await patterns and structured concurrency
- ✅ SwiftUI 6.0 for declarative UI development
- ✅ Combine framework for reactive programming
- ✅ Server-side Swift with Vapor 4.x
- ✅ Enterprise architecture patterns (MVVM, TCA, Clean Architecture)
- ✅ Context7 MCP integration for real-time docs
- ✅ Performance optimization and memory management
- ✅ Testing strategies with XCTest and Swift Testing
- ✅ Swift Concurrency with actors and distributed actors

---

## When to Use

**Automatic triggers**:
- Swift 6.0 development discussions
- SwiftUI and iOS/macOS app development
- Async/await and concurrency patterns
- Combine reactive programming
- Server-side Swift and Vapor development
- Enterprise mobile application architecture

**Manual invocation**:
- Design iOS/macOS application architecture
- Implement async/await patterns
- Optimize performance and memory usage
- Review enterprise Swift code
- Implement reactive UI with Combine
- Troubleshoot concurrency issues

---

## Technology Stack (2025-11-12)

| Component | Version | Purpose | Status |
|-----------|---------|---------|--------|
| **Swift** | 6.0.1 | Core language | ✅ Current |
| **SwiftUI** | 6.0 | Declarative UI | ✅ Current |
| **Combine** | 6.0 | Reactive programming | ✅ Current |
| **Vapor** | 4.102.0 | Server-side framework | ✅ Current |
| **Xcode** | 16.2 | Development environment | ✅ Current |
| **Swift Concurrency** | 6.0 | Async/await & actors | ✅ Current |
| **Swift Testing** | 0.10.0 | Modern testing framework | ✅ Current |

---

## Quick Start: Hello Async/Await

```swift
import Foundation

// Swift 6.0 with async/await
actor GreeterService {
    func greet(name: String) -> String {
        "Hello, \(name)!"
    }
}

// Usage
Task {
    let service = GreeterService()
    let greeting = await service.greet(name: "Swift")
    print(greeting)
}
```

---

## Level 1: Quick Reference

### Core Concepts

1. **Async/Await** - Modern concurrency without callbacks
   - Function marked with `async` - Suspends for I/O
   - Caller uses `await` - Waits for result
   - Native error handling with `throws`
   - Replaces callbacks and completion handlers

2. **SwiftUI** - Declarative UI framework
   - State-driven views update automatically
   - `@State` for local state
   - `@StateObject` for ViewModels
   - Composable views with modifiers

3. **Combine** - Reactive programming
   - Publishers emit values
   - Operators transform pipelines
   - Subscribers receive results
   - Error handling with `.catch`

4. **Actors** - Thread-safe state isolation
   - Protect mutable state automatically
   - Replace locks and semaphores
   - `@MainActor` for UI thread
   - Distributed actors for RPC

5. **Vapor** - Server-side Swift
   - Async route handlers
   - Database integration (Fluent)
   - Middleware for cross-cutting concerns
   - Type-safe API responses

### Project Structure

```
MyApp/
├── Sources/
│   ├── App.swift                 # Entry point
│   ├── Models/                   # Data types
│   ├── Services/                 # Business logic
│   ├── ViewModels/               # UI state management
│   └── Views/                    # SwiftUI components
├── Tests/
│   ├── UnitTests/
│   └── IntegrationTests/
└── Package.swift                 # Dependencies
```

---

## Level 2: Implementation Patterns

### Async/Await Pattern

```swift
import Foundation

// Structured async function
func fetchData() async throws -> String {
    let url = URL(string: "https://api.example.com/data")!
    let (data, _) = try await URLSession.shared.data(from: url)
    return String(data: data, encoding: .utf8) ?? ""
}

// Concurrent operations with TaskGroup
func loadMultipleResources() async throws -> (String, String) {
    try await withThrowingTaskGroup(of: (String, String).self) { group in
        group.addTask { ("users", try await fetchUsers()) }
        group.addTask { ("posts", try await fetchPosts()) }

        var results: [String: String] = [:]
        for try await (key, value) in group {
            results[key] = value
        }
        return (results["users"] ?? "", results["posts"] ?? "")
    }
}
```

### SwiftUI State Management

```swift
import SwiftUI

@MainActor
class ContentViewModel: ObservableObject {
    @Published var items: [String] = []
    @Published var isLoading = false

    func loadItems() async {
        isLoading = true
        defer { isLoading = false }

        do {
            items = try await fetchItems()
        } catch {
            items = []
        }
    }
}

struct ContentView: View {
    @StateObject private var viewModel = ContentViewModel()

    var body: some View {
        NavigationView {
            VStack {
                if viewModel.isLoading {
                    ProgressView()
                } else {
                    List(viewModel.items, id: \.self) { item in
                        Text(item)
                    }
                }
            }
            .navigationTitle("Items")
            .task {
                await viewModel.loadItems()
            }
        }
    }
}
```

### Actor Isolation Pattern

```swift
// Thread-safe counter
actor CounterService {
    private var count: Int = 0

    func increment() { count += 1 }
    func decrement() { count -= 1 }
    func getCount() -> Int { count }
}

// Usage (automatically thread-safe)
Task {
    let counter = CounterService()
    await counter.increment()
    let value = await counter.getCount()
}
```

### Vapor Server Route

```swift
import Vapor

func routes(_ app: Application) throws {
    // GET /api/users
    app.get("api", "users") { req async -> [String: String] in
        return ["status": "success"]
    }

    // POST /api/users
    app.post("api", "users") { req async -> HTTPStatus in
        // Save user
        return .created
    }
}
```

---

## Level 3: Advanced Topics

### Concurrency Best Practices

1. **Prefer async/await** over Combine for sequential operations
2. **Use actors** for mutable shared state (not locks)
3. **Mark UI code @MainActor** to ensure main thread
4. **Handle cancellation** properly in long-running tasks
5. **Avoid blocking operations** (no sleep, no synchronous I/O)

### Performance Optimization

- **Memory**: Use value types (struct) by default
- **CPU**: Profile with Xcode Instruments
- **Rendering**: Keep SwiftUI view body pure
- **Networking**: Implement request caching
- **Database**: Use connection pooling in Vapor

### Security Patterns

- **Input validation**: Always validate user input
- **Error handling**: Don't expose internal errors to users
- **Encryption**: Use CryptoKit for sensitive data
- **Authentication**: Implement JWT or OAuth2
- **SQL injection prevention**: Use parameterized queries

### Testing Strategy

- **Unit tests**: Pure functions with XCTest
- **Integration tests**: Database and API tests
- **UI tests**: SwiftUI view behavior
- **Mocking**: Use protocols for dependency injection

---

## Context7 MCP Integration

**Get latest Swift documentation on-demand:**

```python
# Access Swift documentation via Context7
from context7 import resolve_library_id, get_library_docs

# Swift Language Documentation
swift_id = resolve_library_id("swift")
docs = get_library_docs(
    context7_compatible_library_id=swift_id,
    topic="structured-concurrency",
    tokens=5000
)

# SwiftUI Documentation
swiftui_id = resolve_library_id("swiftui")
swiftui_docs = get_library_docs(
    context7_compatible_library_id=swiftui_id,
    topic="state-management",
    tokens=4000
)

# Vapor Framework Documentation
vapor_id = resolve_library_id("vapor")
vapor_docs = get_library_docs(
    context7_compatible_library_id=vapor_id,
    topic="routing",
    tokens=3000
)
```

---

## Related Skills & Resources

**Language Integration**:
- Skill("moai-context7-lang-integration"): Latest Swift/Vapor documentation

**Quality & Testing**:
- Skill("moai-foundation-testing"): Swift testing best practices
- Skill("moai-foundation-trust"): TRUST 5 principles application

**Security & Performance**:
- Skill("moai-foundation-security"): Security patterns for Swift
- Skill("moai-essentials-debug"): Swift debugging techniques

**Official Resources**:
- [Swift.org Documentation](https://swift.org/documentation)
- [Apple Developer](https://developer.apple.com/documentation)
- [SwiftUI Tutorials](https://developer.apple.com/tutorials/swiftui)
- [Vapor Documentation](https://docs.vapor.codes)
- [Swift Forums](https://forums.swift.org)

---

## Troubleshooting

**Problem**: Sendable conformance error
**Solution**: Implement `Sendable` protocol or use `@Sendable` closure

**Problem**: Actor isolation violation
**Solution**: Use `nonisolated` for safe properties or proper await calls

**Problem**: Memory leaks in closures
**Solution**: Capture `[weak self]` to break retain cycles

**Problem**: SwiftUI view not updating
**Solution**: Ensure state changes happen on `@MainActor`

---

## Changelog

- ** .0** (2025-11-12): Enterprise upgrade - Progressive Disclosure structure, 90% content reduction, Context7 integration
- **v3.0.0** (2025-03-15): SwiftUI 5.0 and Combine 6.0 patterns
- **v2.0.0** (2025-01-10): Basic Swift 5.x patterns
- **v1.0.0** (2024-12-01): Initial release

---

## Resources

**For working examples**: See `examples.md`

**For API reference**: See `reference.md`

**For advanced patterns**: See full SKILL.md in documentation archive

---

_Last updated: 2025-11-18 | Maintained by moai-adk team_
