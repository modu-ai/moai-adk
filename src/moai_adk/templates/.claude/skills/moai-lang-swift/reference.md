# moai-lang-swift - CLI Reference & API Documentation

_Last updated: 2025-11-12_

## Quick Reference

### Installation

```bash
# Install Swift (macOS)
brew install swift

# Install Xcode (macOS)
xcode-select --install

# Verify installation
swift --version
swift package --version
```

### Common Commands

```bash
# Create new Swift package
swift package init --type executable
swift package init --type library

# Build project
swift build
swift build -c release

# Run executable
swift run

# Test project
swift test
swift test --verbose

# Format code
swift format --in-place -r .

# Lint code
swiftlint lint
swiftlint autocorrect

# Generate documentation
swift package generate-documentation
```

## Tool Versions (2025-11-12)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **Swift** | 6.0.1 | Core language | ✅ Current |
| **Xcode** | 16.2 | IDE | ✅ Current |
| **SwiftLint** | 0.57.0 | Linting | ✅ Current |
| **SwiftFormat** | 0.54.0 | Code formatting | ✅ Current |
| **XCTest** | Built-in | Unit testing | ✅ Current |
| **Swift Testing** | 0.10.0 | Modern testing | ✅ Current |

---

## Swift Version Information

| Version | Release Date | Status | Key Features |
|---------|--------------|--------|--------------|
| **Swift 6.0** | 2024-11 | Current | Strict concurrency, async/await by default |
| **Swift 5.9** | 2023-09 | Maintenance | Macro support, explicit concurrency checking |
| **Swift 5.8** | 2023-03 | Deprecated | Async/await, actors, document dependencies |
| **Swift 5.7** | 2022-09 | EOL | Basic async/await, Protocol improvements |

---

## Framework Compatibility Matrix

### SwiftUI Version Support

| Feature | iOS 13+ | iOS 14+ | iOS 15+ | iOS 16+ | iOS 17+ |
|---------|---------|---------|---------|---------|---------|
| View | ✅ | ✅ | ✅ | ✅ | ✅ |
| @State | ✅ | ✅ | ✅ | ✅ | ✅ |
| @StateObject | ❌ | ✅ | ✅ | ✅ | ✅ |
| @EnvironmentObject | ✅ | ✅ | ✅ | ✅ | ✅ |
| NavigationStack | ❌ | ❌ | ❌ | ✅ | ✅ |
| .task modifier | ❌ | ❌ | ✅ | ✅ | ✅ |
| AsyncImage | ❌ | ❌ | ✅ | ✅ | ✅ |

### Combine Framework Version Support

| Feature | iOS 13+ | iOS 14+ | iOS 15+ | iOS 16+ |
|---------|---------|---------|---------|---------|
| Publisher | ✅ | ✅ | ✅ | ✅ |
| Subject | ✅ | ✅ | ✅ | ✅ |
| Subscriber | ✅ | ✅ | ✅ | ✅ |
| .task publisher | ❌ | ❌ | ✅ | ✅ |
| Debounce | ✅ | ✅ | ✅ | ✅ |

### Vapor Framework Version Support

| Vapor | Swift | Release | Status | LTS |
|-------|-------|---------|--------|-----|
| **4.x** | 5.3+ | 2021-02 | Current | Until 2026 |
| **3.x** | 4.0+ | 2018-02 | EOL | Ended 2020 |
| **2.x** | 3.0+ | 2016-05 | EOL | Ended 2019 |

---

## API Reference

### Swift Concurrency APIs

**Core async/await**:
- `async` - Mark function as asynchronous
- `await` - Suspend function until result available
- `throws` - Propagate errors from async functions
- `async throws` - Combine both patterns

**Task Management**:
- `Task {}` - Create unstructured task
- `Task.detached {}` - Create detached task
- `Task.sleep(nanoseconds:)` - Sleep asynchronously
- `Task.isCancelled` - Check cancellation status

**Structured Concurrency**:
- `withTaskGroup()` - Create task group
- `withThrowingTaskGroup()` - Task group with error handling
- `group.addTask {}` - Add task to group

**Actor Isolation**:
- `actor` - Define actor type
- `@MainActor` - Mark for main thread execution
- `distributed actor` - Cross-process actors
- `nonisolated` - Non-isolated context

### SwiftUI View APIs

**View Containers**:
- `VStack` - Vertical stack container
- `HStack` - Horizontal stack container
- `ZStack` - Z-order stack (layered)
- `ForEach` - Iterate over collections
- `Group` - Logical grouping
- `ScrollView` - Scrollable content

**State Management**:
- `@State` - Local view state
- `@Binding` - Two-way binding
- `@StateObject` - Lifecycle-managed object
- `@EnvironmentObject` - Global environment state
- `@ObservedObject` - Observed external object
- `@Environment` - Access environment values

**Modifiers**:
- `.padding()` - Add padding
- `.frame()` - Set size constraints
- `.background()` - Set background color
- `.foregroundColor()` - Set text color
- `.onTapGesture {}` - Handle tap
- `.task {}` - Run async task on appear

### Combine Framework APIs

**Publishers**:
- `PassthroughSubject` - Manual value publisher
- `CurrentValueSubject` - Publisher with current value
- `Future` - Delayed value publisher
- `Timer.publish` - Timer-based publisher
- `URLSession.dataTaskPublisher` - Network publisher

**Operators**:
- `.map` - Transform values
- `.filter` - Filter values
- `.debounce` - Debounce emissions
- `.flatMap` - Transform to nested publishers
- `.zip` - Combine multiple publishers
- `.removeDuplicates` - Filter duplicate values
- `.catch` - Handle errors

**Subscribers**:
- `.sink` - Simple subscriber
- `.assign` - Assign to property
- `.store(in:)` - Store cancellable

---

## Best Practices

### Swift Concurrency

1. **Prefer async/await over Combine**: More readable, type-safe
2. **Use actors for shared mutable state**: Prevents race conditions
3. **Mark UI updates with @MainActor**: Ensure main thread execution
4. **Avoid blocking operations**: Use async alternatives
5. **Handle task cancellation**: Check `Task.isCancelled`

### SwiftUI Development

1. **Keep state minimal**: Only store necessary state
2. **Use @StateObject for ViewModels**: Proper lifecycle management
3. **Separate concerns**: Model, ViewModel, View layers
4. **Avoid deep nesting**: Use view extraction for readability
5. **Optimize animations**: Use `.animation()` modifier strategically

### Combine Programming

1. **Combine with async/await**: Use when interoperability needed
2. **Manage subscriptions**: Use `AnyCancellable` collection
3. **Use appropriate operators**: Choose most efficient operators
4. **Error handling**: Always handle publisher errors
5. **Memory leaks prevention**: Properly store subscriptions

### Server-Side Swift (Vapor)

1. **Use async/await routes**: Modern async route handlers
2. **Validate input data**: Always validate user input
3. **Handle errors properly**: Implement error middleware
4. **Use dependency injection**: Inject services into controllers
5. **Connection pooling**: Configure database connection pools

---

## Environment Setup

### macOS Development

```bash
# Install Xcode
xcode-select --install

# Install Swift toolchain (if needed)
brew install swift

# Install development tools
brew install swift-format swiftlint

# Configure git pre-commit hooks
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
swiftlint lint
EOF
chmod +x .git/hooks/pre-commit
```

### Linux Development

```bash
# Install Swift (Ubuntu 20.04+)
curl https://swift.org/Key.asc | sudo apt-key add -
echo "deb https://apt.swift.org/ubuntu/focal/swift-focal main" | sudo tee /etc/apt/sources.list.d/swift.list
sudo apt update
sudo apt install swift-lang

# Verify installation
swift --version
```

---

## Troubleshooting

### Problem: Sendable Conformance Error

**Symptom**: "type X cannot conform to Sendable"

**Solution**:
```swift
// Add explicit conformance
extension MyType: Sendable {}

// Or use @Sendable closure
func processData(_ callback: @Sendable @escaping () -> Void) {}
```

### Problem: Actor Isolation Violation

**Symptom**: "actor-isolated property cannot be mutated"

**Solution**: Use `nonisolated` or proper await calls
```swift
actor MyActor {
    nonisolated var publicProperty: String { "value" }
    var privateProperty: String = "private"
}
```

### Problem: Memory Leaks in Closures

**Symptom**: Reference cycles in task closures

**Solution**: Use weak self
```swift
Task { [weak self] in
    guard let self = self else { return }
    // Safe to use self
}
```

### Problem: SwiftUI State Not Updating

**Symptom**: View doesn't update when @State changes

**Solution**: Ensure state changes on main thread
```swift
@MainActor
class ViewModel: ObservableObject {
    @Published var data: [Item] = []
}
```

---

## Performance Benchmarks

| Operation | Performance | Latency | Memory | Notes |
|-----------|-------------|---------|--------|-------|
| **Async function call** | 1M ops/sec | <1µs | 50MB | Very fast overhead |
| **Actor method call** | 800K ops/sec | <2µs | 60MB | Isolation cost |
| **SwiftUI render** | 60 FPS | 16.67ms | 100MB | Limited by screen |
| **Combine pipeline** | 500K ops/sec | <5µs | 80MB | Operator chain cost |
| **Vapor request** | 50K req/sec | 20ms | 150MB | Network-bound |

---

## Related Skills & Resources

- Skill("moai-context7-lang-integration"): Get latest Swift docs via Context7
- Skill("moai-foundation-testing"): Swift testing best practices
- Skill("moai-foundation-security"): Security patterns for Swift apps
- Skill("moai-essentials-debug"): Debugging Swift applications

---

## Official Documentation

| Resource | Link | Coverage |
|----------|------|----------|
| **Swift.org** | https://swift.org/documentation | Language, foundation |
| **Apple Developer** | https://developer.apple.com/documentation | All frameworks |
| **SwiftUI Tutorials** | https://developer.apple.com/tutorials/swiftui | UI development |
| **Combine Framework** | https://developer.apple.com/documentation/combine | Reactive programming |
| **Vapor Docs** | https://docs.vapor.codes | Server-side Swift |
| **Swift Forums** | https://forums.swift.org | Community discussion |
| **Vapor Community** | https://vapor.community | Vapor support |

---

## Code Examples by Category

### Async/Await Examples
- Simple function call
- Error handling
- Timeout handling
- Task groups
- Cancellation

### SwiftUI Examples
- State management
- Binding
- View composition
- Navigation
- Custom layouts

### Combine Examples
- Publishers
- Operators
- Error handling
- Multi-publisher combinations

### Server-Side Examples
- Route definition
- Middleware
- Database integration
- Authentication
- Error handling

---

## Glossary

**Actor**: Type that protects mutable state with serial isolation

**Async/Await**: Modern concurrency syntax for asynchronous operations

**@MainActor**: Attribute marking code to run on main thread

**Publisher**: Emit values that subscribers can receive

**Sendable**: Protocol indicating type is thread-safe

**Structured Concurrency**: System where tasks have clear parent-child relationships

**SwiftUI**: Apple's declarative UI framework for iOS/macOS

**Vapor**: Server-side Swift web framework

---

_For working examples, see examples.md_
