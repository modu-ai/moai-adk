// Package quality provides automatic code formatting and linting for PostToolUse hooks.
//
// This package implements SPEC-HOOK-002: Code Quality Automation, which provides
// automatic code formatting and linting after Write/Edit operations. It supports
// 16+ programming languages with automatic tool detection and graceful fallback.
//
// # Architecture
//
// The package consists of four main components:
//
//   - ToolRegistry: Manages tool registration and execution for multiple languages
//   - ChangeDetector: Provides hash-based file change detection
//   - Formatter: Handles automatic code formatting via PostToolUse hooks
//   - Linter: Handles automatic code linting via PostToolUse hooks
//
// # Supported Languages
//
// Python, JavaScript, TypeScript, Go, Rust, Java, Kotlin, Swift, C/C++,
// Ruby, PHP, Elixir, Scala, R, Dart, C#, Markdown
//
// # Usage
//
// The quality handlers are registered with the hook Registry and automatically
// process files after Write/Edit operations:
//
//	import (
//	    "github.com/modu-ai/moai-adk-go/internal/hook"
//	    "github.com/modu-ai/moai-adk-go/internal/hook/quality"
//	)
//
//	func init() {
//	    registry := hook.NewRegistry(cfg)
//	    registry.Register(quality.NewFormatter())
//	    registry.Register(quality.NewLinter())
//	}
//
// # Cross-Platform Support
//
// All tools are executed using exec.Command with proper path validation and
// timeout handling. The package gracefully degrades when tools are unavailable.
package quality
