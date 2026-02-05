---
paths:
  - "**/*.rb"
  - "**/Gemfile"
  - "**/Gemfile.lock"
  - "**/*.gemspec"
---

# Ruby Rules

Version: Ruby 3.3+

## Tooling

- Package management: Bundler
- Linting: RuboCop
- Testing: RSpec or Minitest
- Coverage: SimpleCov >= 85%

## MUST

- Use frozen string literals (# frozen_string_literal: true)
- Use keyword arguments for methods with 3+ params
- Use pattern matching for complex conditionals
- Use Sorbet or RBS for type annotations
- Handle exceptions with specific classes
- Follow Rails conventions for Rails projects

## MUST NOT

- Monkey-patch core classes without good reason
- Use eval or send with user input
- Ignore RuboCop offenses without justification
- Use global variables ($var)
- Leave rescue blocks empty
- Use return nil explicitly

## File Conventions

- *_spec.rb or *_test.rb for test files
- Use snake_case for files and methods
- Use PascalCase for classes and modules
- Use SCREAMING_CASE for constants
- Group related classes in modules

## Testing

- Use RSpec with let and describe blocks
- Use FactoryBot for test data
- Use VCR for HTTP recording
- Use database_cleaner for isolation

## MoAI Integration

- Use Skill("moai-lang-ruby") for detailed patterns
- Follow TRUST 5 quality gates
