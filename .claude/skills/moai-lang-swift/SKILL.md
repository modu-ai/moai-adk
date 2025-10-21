---

name: moai-lang-swift
description: Swift best practices with XCTest, SwiftLint, and iOS/macOS development patterns. Use when writing or reviewing Swift code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Swift Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Swift code discussions, framework guidance, or file extensions such as .swift. |
| Tier | 3 |

## What it does

Provides Swift-specific expertise for TDD development, including XCTest framework, SwiftLint linting, Swift Package Manager, and iOS/macOS platform patterns.

## When to use

- Engages when the conversation references Swift work, frameworks, or files like .swift.
- “Writing Swift tests”, “How to use XCTest”, “iOS patterns”
- Automatically invoked when working with Swift/iOS projects
- Swift SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **XCTest**: Apple's native testing framework
- **Quick/Nimble**: BDD-style testing (alternative)
- **XCUITest**: UI testing for iOS/macOS apps
- Test coverage with Xcode Code Coverage

**Code Quality**:
- **SwiftLint**: Swift linter and style checker
- **SwiftFormat**: Code formatting tool
- **Xcode Analyzer**: Static code analysis

**Package Management**:
- **Swift Package Manager (SPM)**: Dependency management
- **CocoaPods**: Alternative package manager (legacy)
- **Carthage**: Decentralized dependency manager

**Swift Patterns**:
- **Optionals**: Safe handling of nil values (?, !)
- **Guard statements**: Early exit patterns
- **Protocol-oriented programming**: Protocols over inheritance
- **Value types**: Prefer structs over classes
- **Closures**: First-class functions

**iOS/macOS Patterns**:
- **SwiftUI**: Declarative UI framework
- **Combine**: Reactive programming
- **UIKit/AppKit**: Traditional UI frameworks
- **MVVM/MVC**: Architecture patterns

## Examples
```bash
swift test && swift-format --lint --recursive Sources
```

## Inputs
- Language-specific source directories (e.g. `src/`, `app/`).
- Language-specific build/test configuration files (e.g. `package.json`, `pyproject.toml`, `go.mod`).
- Relevant test suites and sample data.

## Outputs
- Test/lint execution plan tailored to the selected language.
- List of key language idioms and review checkpoints.

## Failure Modes
- When the language runtime or package manager is not installed.
- When the main language cannot be determined in a multilingual project.

## Dependencies
- Access to the project file is required using the Read/Grep tool.
- When used with `Skill("moai-foundation-langs")`, it is easy to share cross-language conventions.

## References
- Apple. "Swift Programming Language Guide." https://docs.swift.org/swift-book/ (accessed 2025-03-29).
- Apple. "Swift Package Manager." https://developer.apple.com/documentation/swift_packages (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Swift-specific review)
- mobile-app-expert (iOS app development)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
