---

name: moai-lang-ruby
description: Ruby best practices with RSpec, RuboCop, Bundler, and Rails patterns. Use when writing or reviewing Ruby code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Ruby Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Ruby code discussions, framework guidance, or file extensions such as .rb. |
| Tier | 3 |

## What it does

Provides Ruby-specific expertise for TDD development, including RSpec BDD testing, RuboCop linting, Bundler package management, and Rails framework patterns.

## When to use

- Engages when the conversation references Ruby work, frameworks, or files like .rb.
- “Writing Ruby tests”, “How to use RSpec”, “Rails patterns”
- Automatically invoked when working with Ruby/Rails projects
- Ruby SPEC implementation (`/alfred:2-run`)

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
```bash
bundle exec rubocop && bundle exec rspec
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
- Ruby Lang. "Ruby Programming Language." https://www.ruby-lang.org/en/documentation/ (accessed 2025-03-29).
- RuboCop. "RuboCop Documentation." https://docs.rubocop.org/rubocop/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Ruby-specific review)
- web-api-expert (Rails API development)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
