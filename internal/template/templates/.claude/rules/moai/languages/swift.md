---
paths:
  - "**/*.swift"
  - "**/Package.swift"
  - "**/*.xcodeproj/**"
---

# Swift Rules

Version: Swift 6+

## Tooling

- Build: Xcode or Swift Package Manager
- Linting: SwiftLint
- Testing: XCTest or Swift Testing
- Formatting: swift-format

## MUST

- Use Swift Concurrency (async/await, actors)
- Use Codable for JSON serialization
- Use property wrappers appropriately (@State, @Binding)
- Handle errors with do-catch or Result
- Use guard for early returns
- Document public APIs with documentation comments

## MUST NOT

- Force unwrap optionals (!) without safety check
- Use implicitly unwrapped optionals unless required
- Block the main actor with synchronous calls
- Ignore compiler warnings
- Use stringly-typed APIs
- Create retain cycles in closures (use [weak self])

## File Conventions

- *Tests.swift for test files
- One type per file for public types
- Use PascalCase for types and protocols
- Use camelCase for properties and methods
- Group related functionality with extensions

## Testing

- Use XCTest or Swift Testing framework
- Use async tests for concurrent code
- Mock dependencies with protocols
- Use snapshot testing for UI

## MoAI Integration

- Use Skill("moai-lang-swift") for detailed patterns
- Follow TRUST 5 quality gates
