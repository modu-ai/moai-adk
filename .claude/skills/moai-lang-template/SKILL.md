---
name: moai-lang-template
version: 1.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: Parameterized template for 13 additional languages with Context7 MCP integration and standardized best practices.
keywords: ['template', 'language', 'parameterized', 'context7', 'mcp']
allowed-tools:
  - Read
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Lang Template Skill - Parameterized Language Support

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-template |
| **Version** | 1.0.0 (2025-11-06) |
| **Allowed tools** | Read, Context7 MCP integration |
| **Auto-load** | On demand for supported languages |
| **Tier** | Language |

---

## What It Does

Parameterized template supporting 13 programming languages with Context7 MCP integration for up-to-date documentation and standardized best practices.

## Supported Languages

| Language | Version Range | Primary Use Cases |
| -------- | ------------- | ---------------- |
| Java | 17+ | Enterprise applications, Spring Boot |
| Kotlin | 1.9+ | Android, JVM backend, coroutines |
| Swift | 5.9+ | iOS, macOS applications |
| C# | 11+ | .NET applications, Unity |
| Dart | 3.0+ | Flutter cross-platform |
| C++ | 17/20 | Systems programming, performance |
| C | C11/C17 | Embedded systems, low-level |
| Ruby | 3.2+ | Rails, web applications |
| PHP | 8.2+ | Web backend, Laravel |
| Scala | 3.x | Functional programming, Spark |
| SQL | ANSI SQL | Database queries, migrations |
| Shell | POSIX | DevOps, automation |
| R | 4.4+ | Data analysis, statistics |

---

## Template Variables

The skill uses parameterized variables that adapt to each language:

### Core Variables
```
{{LANGUAGE_NAME}} - Full language name (e.g., "Python", "Java")
{{LANGUAGE_VERSION}} - Current stable version
{{TEST_FRAMEWORK}} - Primary testing framework
{{BUILD_TOOL}} - Build/package management
{{LINTER}} - Code quality tool
{{PACKAGE_MANAGER}} - Dependency management
```

### Context7 Integration
- **Auto-documentation fetch**: Uses Context7 MCP to retrieve latest documentation
- **Version-specific guidance**: Always current with latest releases
- **Official source prioritization**: Primary documentation over tutorials

---

## Usage Pattern

```python
# For any supported language
Skill("moai-lang-template")

# Skill automatically:
# 1. Detects target language from context
# 2. Loads language-specific parameters
# 3. Fetches latest docs via Context7
# 4. Provides current best practices
```

---

## Language-Specific Examples

### Example Usage (Java)
```python
# User asks about Java testing
Skill("moai-lang-template")

# Skill provides:
# - JUnit 5 best practices
# - Maven/Gradle configuration
# - Test coverage strategies
# - Context7-fetched latest docs
```

### Example Usage (TypeScript)
```python
# User asks about TypeScript patterns
Skill("moai-lang-template")

# Skill provides:
# - TypeScript 5.x strict typing
# - Node.js integration patterns
# - E2E type safety
# - Context7-fetched latest docs
```

---

## Context7 MCP Integration

This skill integrates with Context7 MCP for always-current documentation:

### Documentation Sources
- **Official language docs** (python.org, nodejs.org, etc.)
- **Framework documentation** (Spring, Rails, Laravel, etc.)
- **Community standards** (style guides, conventions)

### Auto-Update Process
1. **Language detection**: Automatically identify target language
2. **Documentation fetch**: Use Context7 to get latest official docs
3. **Best practices extraction**: Extract current recommended patterns
4. **Version awareness**: Always provide version-specific guidance

---

## Best Practices By Category

### Testing Strategies
- **Unit testing**: Language-specific testing frameworks
- **Integration testing**: Database and API testing patterns
- **E2E testing**: Full application testing approaches

### Code Quality
- **Linting**: Language-specific linters and formatters
- **Static analysis**: Type checking and security scanning
- **Code review**: Language-agnostic review patterns

### Performance Optimization
- **Profiling**: Language-specific profiling tools
- **Memory management**: Garbage collection and memory patterns
- **Concurrency**: Multi-threading and async patterns

### Security
- **Input validation**: Language-specific validation libraries
- **Authentication**: JWT, OAuth2 implementation patterns
- **Data protection**: Encryption and secure storage

---

## Progressive Disclosure Structure

### High Freedom (10-15 tokens)
Quick guidance and pattern identification:
```python
# Example: Testing setup
"Use pytest with fixtures and parameterized tests for Python testing."
```

### Medium Freedom (30-40 tokens)
Detailed guidance with framework-specific recommendations:
```python
# Example: Java testing strategy
"Set up JUnit 5 with Jupiter, use @ParameterizedTest for data-driven tests,
and integrate with Mockito for mocking in Spring Boot applications."
```

### Low Freedom (60-80 tokens)
Comprehensive guidance with latest Context7 documentation:
```python
# Example: Complete testing setup
"Configure pytest.ini with testpaths and python_files,
create conftest.py with database fixtures,
use pytest-cov for coverage reporting,
and integrate with GitHub Actions for CI/CD.
Latest pytest 8.0 features include async fixtures and improved parametrization."
```

---

## Works Well With

### Core Skills
- `Skill("moai-domain-backend")` - Backend architecture patterns
- `Skill("moai-domain-frontend")` - Frontend integration patterns
- `Skill("moai-foundation-trust")` - TRUST 5 compliance

### Language-Specific Skills
- `Skill("moai-lang-python")` - Python expertise (premium)
- `Skill("moai-lang-typescript")` - TypeScript expertise (premium)
- `Skill("moai-lang-go")` - Go expertise (premium)
- `Skill("moai-lang-rust")` - Rust expertise (premium)
- `Skill("moai-lang-javascript")` - JavaScript expertise (premium)

### Tools Integration
- **Context7 MCP**: Latest documentation retrieval
- **GitHub Actions**: CI/CD pipeline setup
- **Docker**: Containerization patterns

---

## Model Recommendations

- **Haiku**: Quick pattern identification and basic guidance
- **Sonnet**: Detailed best practices with Context7 integration
- **Opus**: Complex architectural decisions and migration strategies

---

## Quality Assurance

### Validation Checklist
- [ ] Language correctly detected from context
- [ ] Context7 documentation successfully retrieved
- [ ] Version-specific guidance provided
- [ ] Best practices are current and official
- [ ] Examples are functional and tested

### Self-Correction
- If Context7 fails → Use cached documentation with version disclaimer
- If language ambiguous → Ask for clarification with supported options
- If outdated patterns detected → Flag and provide current alternatives

---

**Last Updated**: 2025-11-06
**Version**: 1.0.0 (Parameterized template system)
**Context7 Integration**: Enabled for all 13 supported languages
**Status**: Production Ready