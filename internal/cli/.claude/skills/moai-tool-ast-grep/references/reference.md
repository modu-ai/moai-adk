# AST-Grep Tool - Reference Documentation

## Official Resources

### Core Documentation
- **AST-Grep Official Site**: https://ast-grep.github.io/
  - Main documentation hub
  - Getting started guide
  - Feature overview

- **GitHub Repository**: https://github.com/ast-grep/ast-grep
  - Source code
  - Issue tracker
  - Contributing guide
  - Release notes

- **Playground**: https://ast-grep.github.io/playground.html
  - Interactive pattern testing
  - Live rule validation
  - Multi-language support

### Installation

#### Package Managers
- **Homebrew (macOS/Linux)**: https://ast-grep.github.io/guide/install.html#homebrew
  ```bash
  brew install ast-grep
  ```

- **npm (Cross-platform)**: https://www.npmjs.com/package/@ast-grep/cli
  ```bash
  npm install -g @ast-grep/cli
  ```

- **Cargo (Rust)**: https://crates.io/crates/ast-grep
  ```bash
  cargo install ast-grep
  ```

- **Pre-built Binaries**: https://github.com/ast-grep/ast-grep/releases
  - Linux, macOS, Windows
  - No installation required

### Core Documentation

#### Getting Started
- **Quick Start**: https://ast-grep.github.io/guide/quick-start.html
  - Basic pattern search
  - First rule creation
  - Common use cases

- **Pattern Syntax**: https://ast-grep.github.io/reference/pattern.html
  - Meta-variables ($VAR)
  - Variadic variables ($$$ARGS)
  - Anonymous variables ($$_)
  - Wildcards and quantifiers

#### Rule Configuration
- **YAML Rule Reference**: https://ast-grep.github.io/reference/yaml.html
  - Rule structure
  - Composite rules (all/any/not)
  - Relational rules (inside/has/follows)
  - Fix patterns
  - Severity levels

- **Configuration Files**: https://ast-grep.github.io/reference/config.html
  - sgconfig.yml structure
  - Rule directory organization
  - Test configuration
  - Language glob patterns

### Advanced Features

#### Pattern Matching
- **Advanced Patterns**: https://ast-grep.github.io/guide/pattern-syntax.html
  - String matching
  - Regular expressions in patterns
  - Node type matching
  - Field constraints

#### Relational Rules
- **Inside Rule**: https://ast-grep.github.io/reference/rule.html#inside
  - Scoped search within parent nodes
  - Nested context matching

- **Has Rule**: https://ast-grep.github.io/reference/rule.html#has
  - Contains check
  - Descendant node matching

- **Follows/Precedes Rules**: https://ast-grep.github.io/reference/rule.html#follows-precedes
  - Sequential pattern matching
  - Order-based validation

#### Composite Rules
- **All/Any/Not**: https://ast-grep.github.io/reference/rule.html#all-any-not
  - Logical combinations
  - Complex conditionals
  - Negation patterns

### Language Support

#### Supported Languages
- **Full Language List**: https://ast-grep.github.io/languages.html
  - Python, JavaScript, TypeScript
  - Go, Rust, Java, Kotlin
  - C, C++, C#, Swift
  - Ruby, PHP, Scala, Elixir
  - HTML, CSS, Vue, Svelte
  - And 30+ more

#### Language-Specific Guides
- **Python Patterns**: https://ast-grep.github.io/languages/python.html
- **JavaScript Patterns**: https://ast-grep.github.io/languages/javascript.html
- **TypeScript Patterns**: https://ast-grep.github.io/languages/typescript.html
- **Go Patterns**: https://ast-grep.github.io/languages/go.html

### Security Scanning

#### Security Rules
- **Security Templates**: https://ast-grep.github.io/guide/security.html
  - SQL injection detection
  - XSS vulnerability scanning
  - Hardcoded credential detection
  - Insecure dependency checks

#### OWASP Integration
- **OWASP Top 10**: https://ast-grep.github.io/guide/security.html#owasp-top-10
  - Injection attacks
  - Broken authentication
  - Sensitive data exposure
  - Security misconfigurations

### Refactoring & Transformation

#### Code Transformation
- **Rewrite Rules**: https://ast-grep.github.io/reference/rule.html#fix
  - Pattern-to-pattern mapping
  - Variable substitution
  - Multi-file transformations

#### Refactoring Patterns
- **Common Refactorings**: https://ast-grep.github.io/guide/refactoring.html
  - API migration
  - Function renaming
  - Code modernization
  - Design pattern application

### Testing & Validation

#### Rule Testing
- **Test Framework**: https://ast-grep.github.io/guide/test.html
  - Snapshot testing
  - Inline test cases
  - Test organization

#### CI/CD Integration
- **GitHub Actions**: https://ast-grep.github.io/guide/ci.html#github-actions
  - Workflow examples
  - Automated scanning
  - PR integration

- **Pre-commit Hooks**: https://ast-grep.github.io/guide/ci.html#pre-commit
  - Local validation
  - Fast feedback
  - Configuration examples

### Context7 Integration

- **Library Resolution**: Use `mcp__context7__resolve-library-id` with query "ast-grep"
- **Documentation Fetch**: Use `mcp__context7__get-library-docs` for latest docs

### Module Organization

This skill contains 4 modules:

- **modules/pattern-syntax.md** - Complete pattern syntax reference
  - Meta-variable types
  - Wildcard patterns
  - String and regex matching
  - Language-specific syntax

- **modules/security-rules.md** - Security scanning rule templates
  - SQL injection detection
  - XSS vulnerability patterns
  - Hardcoded secret detection
  - Insecure dependency checks
  - OWASP Top 10 coverage

- **modules/refactoring-patterns.md** - Common refactoring patterns
  - API migration patterns
  - Function/method renaming
  - Code modernization
  - Design pattern applications
  - Multi-file transformations

- **modules/language-specific.md** - Language-specific patterns
  - Python patterns (decorators, context managers, etc.)
  - JavaScript/TypeScript patterns (React hooks, async/await, etc.)
  - Go patterns (error handling, interfaces, etc.)
  - Rust patterns (macros, traits, etc.)

### MoAI-ADK Integration

#### Tool Registry
- **Registration**: `internal/hook/registry.go` as AST_ANALYZER type
- **Permissions**: Auto-allowed for `Bash(sg:*)` and `Bash(ast-grep:*)`
- **Hooks**: PostToolUse hook for automatic security scanning

#### Running Scans
```bash
# Scan with MoAI-ADK rules
sg scan --config .claude/skills/moai-tool-ast-grep/rules/sgconfig.yml

# Scan specific directory
sg scan --config sgconfig.yml src/

# JSON output for CI/CD
sg scan --config sgconfig.yml --json > results.json
```

### Community & Support

- **GitHub Discussions**: https://github.com/ast-grep/ast-grep/discussions
  - Q&A threads
  - Pattern sharing
  - Community rules

- **Discord Server**: https://discord.gg/HuWxGYW6Pc
  - Real-time chat
  - Pattern help
  - Feature discussions

- **Twitter**: https://twitter.com/ast_grep
  - Updates and tips
  - Pattern examples
  - Community highlights

### Related Tools

#### Complementary Tools
- **ripgrep (rg)**: https://github.com/BurntSushi/ripgrep
  - Fast text search
  - Regex-based filtering

- **grep.app**: https://grep.app
  - Code search across GitHub
  - Pattern discovery

- **GitHub Code Search**: https://cs.github.com
  - Large-scale code search
  - Cross-repository patterns

### Related Skills

- **moai-workflow-testing** - DDD integration, test pattern detection
- **moai-foundation-quality** - TRUST 5 compliance, code quality gates
- **moai-domain-backend** - API pattern detection, security scanning
- **moai-domain-frontend** - React/Vue pattern optimization
- **moai-lang-python** - Python-specific security and style rules
- **moai-lang-typescript** - TypeScript type safety patterns

### Related Agents

- **expert-refactoring** - AST-based large-scale refactoring
- **expert-security** - Security vulnerability scanning
- **manager-quality** - Code complexity analysis
- **expert-debug** - Pattern-based debugging

### Books & Resources

- **Refactoring**: Martin Fowler
  - Classic refactoring patterns
  - Code smells identification

- **Clean Code**: Robert C. Martin
  - Code quality principles
  - Best practices

- **Design Patterns**: Gang of Four
  - Pattern implementations
  - Structural improvements

---

**Last Updated**: 2026-01-06
**Skill Version**: 1.0.0
**Total Modules**: 4
**Supported Languages**: 40+
