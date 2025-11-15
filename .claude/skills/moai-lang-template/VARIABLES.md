# Language Skill Template Variables

This document defines the template variables used for generating language-specific skills from the moai-lang-template.

## Core Variables

### Required Variables
- `{{LANGUAGE_NAME}}`: Full language name (e.g., "Java", "Kotlin", "Swift")
- `{{LANGUAGE_SLUG}}`: Language identifier for filenames (e.g., "java", "kotlin", "swift")
- `{{LANGUAGE_EXTENSION}}`: File extension (e.g., ".java", ".kt", ".swift")
- `{{LATEST_VERSION}}`: Current stable version (e.g., "21", "1.9.24", "5.9")
- `{{LTS_VERSION}}`: Long-term support version if available
- `{{PACKAGE_MANAGER}}`: Package manager name and commands
- `{{PACKAGE_MANAGER_VERSION}}`: Current version of package manager
- `{{RUNTIME_COMPILER}}`: Compiler or runtime name
- `{{RUNTIME_VERSION}}`: Runtime/compiler version

### Type System Variables
- `{{LANGUAGE_TYPE_SYSTEM}}`: Static/Dynamic/Gradual typing
- `{{PRIMARY_PARADIGMS}}`: Main programming paradigms (e.g., "OOP, Functional")

### Domain Variables
- `{{PRIMARY_DOMAIN}}`: Main application domain (e.g., "backend", "mobile", "systems", "data")
- `{{ECOSYSTEM_KEYWORDS}}`: Comma-separated keywords for skill discovery

## Conditional Sections

### Backend Focus (BACKEND_FOCUS=true)
- `{{WEB_FRAMEWORKS}}`: Main web frameworks
- `{{DATABASE_FRAMEWORKS}}`: Database integration frameworks
- `{{WEB_FRAMEWORKS_AVAILABLE}}`: Array of web framework objects
- `{{INTEGRATION_TESTING_AVAILABLE}}`: true if integration testing supported

### Mobile Focus (MOBILE_FOCUS=true)
- `{{MOBILE_FRAMEWORKS}}`: Mobile development frameworks
- `{{PLATFORM_SERVICES}}`: Platform-specific APIs and services

### Systems Focus (SYSTEMS_FOCUS=true)
- `{{PERFORMANCE_TOOLS}}`: Performance optimization tools
- `{{CROSS_PLATFORM_TOOLS}}`: Multi-platform build tools

### Data Focus (DATA_FOCUS=true)
- `{{DATA_FRAMEWORKS}}`: Data processing frameworks
- `{{STATISTICS_FRAMEWORKS}}`: Statistical computing frameworks
- `{{VISUALIZATION_FRAMEWORKS}}`: Data visualization libraries

## Tool Categories

### Testing Tools
- `{{TESTING_TOOLS}}`: Array of testing tool objects
- `{{TEST_FRAMEWORK}}`: Primary testing framework
- `{{TESTING_CONFIGURATION}}`: Configuration file content
- `{{TESTING_EXAMPLES}}`: Array of code examples
- `{{INTEGRATION_TESTING_EXAMPLE}}`: Integration testing code

### Code Quality
- `{{CODE_QUALITY_TOOLS}}`: Array of linting/formatting tools
- `{{PRECOMMIT_CONFIG}}`: Pre-commit configuration content

### Security
- `{{SECURITY_EXAMPLES}}`: Array of security code examples
- `{{AUTHENTICATION_PATTERNS}}`: Authentication code patterns
- `{{SECURITY_TOOLS}}`: Array of security scanning commands

## Code Examples

### Development Patterns
- `{{TYPE_EXAMPLES}}`: Array of type system examples
- `{{PARADIGM_EXAMPLES}}`: Array of programming paradigm examples
- `{{ASYNC_PATTERNS_AVAILABLE}}`: true if async patterns supported
- `{{ASYNC_PATTERN_EXAMPLE}}`: Async programming code example
- `{{DATA_STRUCTURE_EXAMPLES}}`: Array of data structure examples

### Performance
- `{{PERFORMANCE_EXAMPLES}}`: Array of performance optimization examples
- `{{PROFILING_COMMANDS}}`: Array of profiling commands

### Integration
- `{{INTEGRATION_EXAMPLES}}`: Array of integration pattern examples

### Domain-Specific
- `{{DOMAIN_SPECIFIC_PATTERNS}}`: Array of domain-specific patterns

## Project Configuration

- `{{PROJECT_STRUCTURE}}`: Directory structure template
- `{{PACKAGE_MANAGEMENT_COMMANDS}}`: Array of package management commands
- `{{PROJECT_CONFIGURATION}}`: Build configuration content
- `{{DOCKER_EXAMPLE}}`: Dockerfile example

## Template Processing Notes

The template uses Handlebars-style conditional logic:
- `{{#if CONDITION}}...{{/if}}` for conditional sections
- `{{#each ARRAY}}...{{/each}}` for array iteration
- `{{variable}}` for simple variable substitution

Each language skill should define all required variables and optionally include conditional sections based on the language's primary domain and capabilities.
