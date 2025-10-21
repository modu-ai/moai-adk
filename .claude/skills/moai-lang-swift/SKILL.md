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
- 언어별 소스 디렉터리(e.g. `src/`, `app/`).
- 언어별 빌드/테스트 설정 파일(예: `package.json`, `pyproject.toml`, `go.mod`).
- 관련 테스트 스위트 및 샘플 데이터.

## Outputs
- 선택된 언어에 맞춘 테스트/린트 실행 계획.
- 주요 언어 관용구와 리뷰 체크포인트 목록.

## Failure Modes
- 언어 런타임이나 패키지 매니저가 설치되지 않았을 때.
- 다중 언어 프로젝트에서 주 언어를 판별하지 못했을 때.

## Dependencies
- Read/Grep 도구로 프로젝트 파일 접근이 필요합니다.
- `Skill("moai-foundation-langs")`와 함께 사용하면 교차 언어 규약 공유가 용이합니다.

## References
- Apple. "Swift Programming Language Guide." https://docs.swift.org/swift-book/ (accessed 2025-03-29).
- Apple. "Swift Package Manager." https://developer.apple.com/documentation/swift_packages (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Swift-specific review)
- mobile-app-expert (iOS app development)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
