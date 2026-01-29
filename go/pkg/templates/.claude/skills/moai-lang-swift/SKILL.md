---
name: moai-lang-swift
description: >
  Swift 6+ development specialist covering SwiftUI, Combine, Swift
  Concurrency, and iOS patterns. Use when building iOS apps, macOS apps, or
  Apple platform applications.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.1.0"
  category: "language"
  status: "active"
  updated: "2026-01-11"
  modularized: "true"
  tags: "swift, swiftui, ios, macos, combine, concurrency"
  context7-libraries: "/apple/swift, /apple/swift-evolution"
  related-skills: "moai-lang-kotlin, moai-lang-flutter"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["Swift", "SwiftUI", "Combine", "iOS", "macOS", "async", "await", "Actor", "@Observable", ".swift", "Xcode"]
  languages: ["swift"]
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

Package.swift Configuration: Begin with swift-tools-version comment set to 6.0. Import PackageDescription. Define let package with Package initializer. Set name, platforms array with .iOS and .macOS minimum versions, products array with library definitions, dependencies array with package URLs and version requirements, and targets array with target and testTarget entries including dependencies.

### Essential Patterns

Basic @Observable ViewModel: Import Observation framework. Apply @Observable and @MainActor attributes to final class. Declare private(set) var properties for state. Create async functions that set isLoading true, use defer to set false, and assign fetched data with try? await and nil coalescing.

Basic SwiftUI View: Define struct conforming to View. Declare @State private var for viewModel. In body computed property, use NavigationStack containing List iterating over viewModel items. Add .task modifier calling await on viewModel.load and .refreshable modifier for pull-to-refresh.

Basic Actor for Thread Safety: Define actor type with private dictionary for cache. Create get function returning optional Data for key lookup. Create set function taking key and data parameters for cache storage.

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

Async Test with MainActor: Apply @MainActor attribute to test class extending XCTestCase. Define test function with async throws. Create mock API and set mock data. Initialize system under test with mock. Call await on async method. Use XCTAssertEqual for count verification and XCTAssertFalse for boolean state checks.

## Works Well With

- `moai-lang-kotlin` - Android counterpart for cross-platform projects
- `moai-lang-flutter` - Flutter/Dart for cross-platform mobile
- `moai-domain-backend` - API integration and backend communication
- `moai-foundation-quality` - iOS security best practices
- `moai-workflow-testing` - Xcode debugging and profiling

## Resources

- [reference.md](reference.md) - Architecture patterns, network layer, SwiftData
- [examples.md](examples.md) - Production-ready code examples
