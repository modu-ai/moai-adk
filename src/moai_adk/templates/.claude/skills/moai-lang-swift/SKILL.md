---
name: "moai-lang-swift"
description: "Swift 6+ development specialist covering SwiftUI, Combine, Swift Concurrency, and iOS patterns. Use when building iOS apps, macOS apps, or Apple platform applications."
version: 2.0.0
category: "language"
modularized: true
user-invocable: false
tags: ['swift', 'swiftui', 'ios', 'macos', 'combine', 'concurrency']
allowed-tools:
  - Read
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
context7-libraries: ['/apple/swift', '/apple/swift-evolution']
related-skills: ['moai-lang-kotlin', 'moai-lang-flutter']
updated: 2026-01-08
status: "active"
---

# Swift 6+ Development Specialist

Swift 6.0+ development expert for iOS/macOS with SwiftUI, Combine, and Swift Concurrency.

Auto-Triggers: Swift files (`.swift`), iOS/macOS projects, Xcode workspaces

## Quick Reference

### Core Capabilities

- Swift 6.0: Typed throws, complete concurrency, data-race safety by default
- SwiftUI 6: @Observable macro, NavigationStack, modern declarative UI
- Combine: Reactive programming with publishers and subscribers
- Swift Concurrency: async/await, actors, TaskGroup, isolation
- XCTest: Unit testing, UI testing, async test support
- Swift Package Manager: Dependency management

### Version Requirements

- Swift: 6.0+
- Xcode: 16.0+
- iOS: 17.0+ (recommended), minimum 15.0
- macOS: 14.0+ (recommended)

### Project Setup

Package.swift Configuration:
```swift
// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "MyApp",
    platforms: [.iOS(.v17), .macOS(.v14)],
    products: [
        .library(name: "MyAppCore", targets: ["MyAppCore"])
    ],
    dependencies: [
        .package(url: "https://github.com/Alamofire/Alamofire.git", from: "5.9.0"),
        .package(url: "https://github.com/pointfreeco/swift-composable-architecture", from: "1.15.0")
    ],
    targets: [
        .target(name: "MyAppCore", dependencies: ["Alamofire"]),
        .testTarget(name: "MyAppCoreTests", dependencies: ["MyAppCore"])
    ]
)
```

### Essential Patterns

Basic @Observable ViewModel:
```swift
import Observation

@Observable
@MainActor
final class ContentViewModel {
    private(set) var items: [Item] = []
    private(set) var isLoading = false
    
    func load() async {
        isLoading = true
        defer { isLoading = false }
        items = try? await api.fetchItems() ?? []
    }
}
```

Basic SwiftUI View:
```swift
struct ContentView: View {
    @State private var viewModel = ContentViewModel()
    
    var body: some View {
        NavigationStack {
            List(viewModel.items) { item in
                Text(item.title)
            }
            .task { await viewModel.load() }
            .refreshable { await viewModel.load() }
        }
    }
}
```

Basic Actor for Thread Safety:
```swift
actor DataCache {
    private var cache: [String: Data] = [:]
    
    func get(_ key: String) -> Data? { cache[key] }
    func set(_ key: String, data: Data) { cache[key] = data }
}
```

## Module Index

### Swift 6 Features
[modules/swift6-features.md](modules/swift6-features.md)
- Typed throws for precise error handling
- Complete concurrency checking
- Data-race safety by default
- Sendable conformance requirements

### SwiftUI Patterns
[modules/swiftui-patterns.md](modules/swiftui-patterns.md)
- @Observable macro and state management
- NavigationStack and navigation patterns
- View lifecycle and .task modifier
- Environment and dependency injection

### Swift Concurrency
[modules/concurrency.md](modules/concurrency.md)
- async/await fundamentals
- Actor isolation and @MainActor
- TaskGroup for parallel execution
- Custom executors and structured concurrency

### Combine Framework
[modules/combine-reactive.md](modules/combine-reactive.md)
- Publishers and Subscribers
- Operators and transformations
- async/await bridge patterns
- Integration with SwiftUI

## Context7 Library Mappings

### Core Swift
- `/apple/swift` - Swift language and standard library
- `/apple/swift-evolution` - Swift evolution proposals
- `/apple/swift-package-manager` - SwiftPM documentation
- `/apple/swift-async-algorithms` - Async sequence algorithms

### Popular Libraries
- `/Alamofire/Alamofire` - HTTP networking
- `/onevcat/Kingfisher` - Image downloading and caching
- `/realm/realm-swift` - Mobile database
- `/pointfreeco/swift-composable-architecture` - TCA architecture
- `/Quick/Quick` - BDD testing framework
- `/Quick/Nimble` - Matcher framework

## Testing Quick Start

Async Test with MainActor:
```swift
@MainActor
final class ViewModelTests: XCTestCase {
    func testLoadSuccess() async throws {
        let mockAPI = MockAPI()
        mockAPI.mockItems = [Item(id: "1", title: "Test")]
        let sut = ContentViewModel(api: mockAPI)
        
        await sut.load()
        
        XCTAssertEqual(sut.items.count, 1)
        XCTAssertFalse(sut.isLoading)
    }
}
```

## Works Well With

- `moai-lang-kotlin` - Android counterpart for cross-platform projects
- `moai-lang-flutter` - Flutter/Dart for cross-platform mobile
- `moai-domain-backend` - API integration and backend communication
- `moai-foundation-quality` - iOS security best practices
- `moai-workflow-testing` - Xcode debugging and profiling

## Resources

- [reference.md](reference.md) - Architecture patterns, network layer, SwiftData
- [examples.md](examples.md) - Production-ready code examples
