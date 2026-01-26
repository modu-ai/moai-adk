---
paths:
  - "**/*.swift"
  - "**/Package.swift"
---

# Swift Rules

Version: Swift 6 / 6.2

## Tooling

- Build: Swift Package Manager or Xcode
- Testing: XCTest with coverage >= 85%
- Linting: SwiftLint

## Best Practices (2026)

- Use "Approachable Concurrency" philosophy (Swift 6.2)
- Use `@MainActor` for UI code - guarantees main thread execution
- Use `.task {}` modifier - auto-cancels when view disappears
- Use actors for mutable state isolation
- Mark types as `Sendable` for cross-concurrency-domain safety

## Swift 6.2 Concurrency

```swift
// Default MainActor isolation for modules
// Enable: -default-isolation MainActor

// Explicit concurrent execution
@concurrent
func processData() async -> Data {
    // Runs in parallel, not on MainActor
}

// Nonisolated async inheritance
// Inherits caller's actor instead of own isolation
nonisolated func helper() async {
    // Inherits isolation from caller
}
```

## SwiftUI Patterns

- `@State` for local state
- `@Binding` for parent-child communication
- `@Observable` (macro) for external state (iOS 17+)
- Use `NavigationStack` with `.navigationDestination` for type-safe navigation

## Progressive Disclosure

Swift 6.2 principle: Only learn concurrency concepts as you need them:
1. Sequential code first
2. Add async/await for suspending APIs
3. Add actors/Sendable only when introducing parallelism

## MoAI Integration

- Use Skill("moai-lang-swift") for detailed patterns
- Follow TRUST 5 quality gates
