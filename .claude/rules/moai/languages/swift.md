---
paths: "**/*.swift,**/Package.swift,**/*.xcodeproj/**"
---

# Swift Development Guide


---


# Swift 6+ Development Specialist

Swift 6.0+ development expert for iOS/macOS with SwiftUI, Combine, and Swift Concurrency.


Core Capabilities:

- Swift 6.0: Typed throws, complete concurrency, data-race safety by default
- SwiftUI 6: @Observable macro, NavigationStack, modern declarative UI
- Combine: Reactive programming with publishers and subscribers
- Swift Concurrency: async/await, actors, TaskGroup, isolation
- XCTest: Unit testing, UI testing, async test support
- Swift Package Manager: Dependency management

### Version Requirements

- Swift: 6.0+
- Xcode: 16.0+
- iOS: 18.0+ (recommended), minimum 16.0
- macOS: 15.0+ (recommended)

### Project Setup

Package.swift Configuration: Begin with swift-tools-version comment set to 6.0. Import PackageDescription. Define let package with Package initializer. Set name, platforms array with .iOS and .macOS minimum versions, products array with library definitions, dependencies array with package URLs and version requirements, and targets array with target and testTarget entries including dependencies.

### Essential Patterns

Basic @Observable ViewModel: Import Observation framework. Apply @Observable and @MainActor attributes to final class. Declare private(set) var properties for state. Create async functions that set isLoading true, use defer to set false, and assign fetched data with try? await and nil coalescing.

Basic SwiftUI View: Define struct conforming to View. Declare @State private var for viewModel. In body computed property, use NavigationStack containing List iterating over viewModel items. Add .task modifier calling await on viewModel.load and .refreshable modifier for pull-to-refresh.

Basic Actor for Thread Safety: Define actor type with private dictionary for cache. Create get function returning optional Data for key lookup. Create set function taking key and data parameters for cache storage.

## Coverage Areas

This guide is self-contained. Use the sections below as the primary reference for:

### Swift 6 Features

- Typed throws for precise error handling
- Complete concurrency checking
- Data-race safety by default
- Sendable conformance requirements

### SwiftUI Patterns

- @Observable macro and state management
- NavigationStack and navigation patterns
- View lifecycle and .task modifier
- Environment and dependency injection

### Swift Concurrency

- async/await fundamentals
- Actor isolation and @MainActor
- TaskGroup for parallel execution
- Custom executors and structured concurrency

### Combine Framework

- Publishers and Subscribers
- Operators and transformations
- async/await bridge patterns
- Integration with SwiftUI

## Context7 Integration

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


- `.claude/rules/moai/languages/kotlin.md` - Android counterpart for cross-platform projects
- `.claude/rules/moai/languages/flutter.md` - Flutter/Dart for cross-platform mobile
- `moai-domain-backend` - API integration and backend communication
- `moai-foundation-quality` - iOS security best practices
- `moai-workflow-testing` - Xcode debugging and profiling

## Resources

For architecture patterns, network-layer design, SwiftData, and production-ready code examples, use the Coverage Areas and Context7 Integration sections above; this guide is self-contained.

---

## Troubleshooting

Common Issues:

Xcode Build Clean: Run xcodebuild clean to remove build artifacts derived from the current project. For a deeper clean, delete the DerivedData folder at ~/Library/Developer/Xcode/DerivedData (affects all Xcode projects on this machine). Close Xcode before deleting the directory.

SwiftPM Resolution Errors: Run swift package resolve to re-resolve dependencies. For cache corruption, delete .build/ and Package.resolved then run swift package update. Verify Package.swift version constraints are compatible across dependencies.

Xcode "No such module" Errors: Run xcodebuild clean, then build again. Ensure the framework search paths and import paths are correctly configured in Build Settings. For SwiftPM dependencies, verify they are linked in the target's Frameworks, Libraries, and Embedded Content section.

Simulator Issues: Run xcrun simctl list devices to enumerate available simulators. Reset a simulator with xcrun simctl erase all (erases all simulators) or xcrun simctl erase <device-id>. If simulator fails to boot, run xcrun simctl shutdown all then boot again.

Code Signing Issues: Run security find-identity -v -p codesigning to list available signing identities. Verify the development team is set in Signing & Capabilities. For local development, use "Automatically manage signing" with a personal team.

---

