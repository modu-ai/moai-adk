// Package models provides shared data models and types for MoAI-ADK.
//
// This package contains configuration types, enums, and data structures
// that are used across multiple packages in the MoAI-ADK codebase.
//
// # Development Modes
//
// The package defines two development methodologies:
//   - DDD (Domain-Driven Development): ANALYZE-PRESERVE-IMPROVE cycle for existing codebases
//   - TDD (Test-Driven Development): RED-GREEN-REFACTOR cycle for new projects and feature development (default)
//
// Use [DevelopmentMode] type and its constants:
//
//	mode := models.ModeTDD
//	if mode.IsValid() {
//	    fmt.Println("Valid mode:", mode)
//	}
//
// # Project Types
//
// Projects can be classified into four types:
//   - WebApp: Web applications with frontend and backend
//   - API: Backend API services
//   - CLI: Command-line interface tools
//   - Library: Reusable code libraries
//
// # Configuration Types
//
// The package provides structured configuration types:
//   - [UserConfig]: User preferences and identity
//   - [LanguageConfig]: Language and localization settings
//   - [QualityConfig]: Code quality and testing standards
//   - [ProjectConfig]: Project metadata and settings
//
// # Language Support
//
// Supported conversation languages can be queried:
//
//	langs := models.SupportedLanguages() // ["ko", "en", "ja", "zh"]
//	name := models.GetLanguageName("ko") // "Korean (한국어)"
package models
