---
name: moai-lang-swift
description: Swift best practices with XCTest, SwiftLint, and iOS/macOS development
  patterns
allowed-tools:
- Read
- Bash
---

# Swift Expert

## What it does

Provides Swift-specific expertise for TDD development, including XCTest framework, SwiftLint linting, Swift Package Manager, and iOS/macOS platform patterns.

## When to use

- "Swift 테스트 작성", "XCTest 사용법", "iOS 패턴"
- Automatically invoked when working with Swift/iOS projects
- Swift SPEC implementation (`/alfred:2-build`)

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

### Example 1: TDD with XCTest
User: "/alfred:2-build AUTH-001"
Claude: (creates RED test with XCTest, GREEN implementation with optionals, REFACTOR)

### Example 2: SwiftLint check
User: "SwiftLint 실행"
Claude: (runs swiftlint and reports style violations)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Swift-specific review)
- mobile-app-expert (iOS app development)
