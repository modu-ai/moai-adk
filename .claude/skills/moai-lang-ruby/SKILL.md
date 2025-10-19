---
name: moai-lang-ruby
description: Ruby best practices with RSpec, RuboCop, Bundler, and Rails patterns
allowed-tools:
- Read
- Bash
---

# Ruby Expert

## What it does

Provides Ruby-specific expertise for TDD development, including RSpec BDD testing, RuboCop linting, Bundler package management, and Rails framework patterns.

## When to use

- "Ruby 테스트 작성", "RSpec 사용법", "Rails 패턴"
- Automatically invoked when working with Ruby/Rails projects
- Ruby SPEC implementation (`/alfred:2-build`)

## How it works

**TDD Framework**:
- **RSpec**: Behavior-driven development (describe, context, it)
- **FactoryBot**: Test data factories
- **Capybara**: Integration testing for web apps
- Test coverage ≥85% with SimpleCov

**Code Quality**:
- **RuboCop**: Ruby linter and formatter
- **Reek**: Code smell detection
- **Brakeman**: Security vulnerability scanner (Rails)

**Package Management**:
- **Bundler**: Dependency management with Gemfile
- **RubyGems**: Package distribution
- Semantic versioning in gemspec

**Rails Patterns**:
- MVC architecture (Model-View-Controller)
- ActiveRecord for database interactions
- RESTful routing conventions
- Service objects for business logic
- Strong parameters for security

**Best Practices**:
- File ≤300 LOC, method ≤50 LOC
- Prefer symbols over strings for hash keys
- Use blocks and yielding for abstraction
- Duck typing over explicit type checking

## Examples

### Example 1: TDD with RSpec
User: "/alfred:2-build USER-001"
Claude: (creates RED test with RSpec BDD style, GREEN implementation, REFACTOR)

### Example 2: RuboCop check
User: "RuboCop 실행"
Claude: (runs rubocop and reports style violations)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Ruby-specific review)
- web-api-expert (Rails API development)
