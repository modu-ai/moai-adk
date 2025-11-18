---
name: moai-lang-template
version: 4.0.0
updated: '2025-11-19'
status: stable
stability: stable
description: {{LANGUAGE_NAME}} best practices with modern frameworks, {{PRIMARY_DOMAIN}},
allowed-tools:
- Read
- Bash
- WebSearch
- WebFetch
---



# {{LANGUAGE_NAME}} Development Mastery

**Modern {{LANGUAGE_NAME}} Development with 2025 Best Practices**

> Comprehensive {{LANGUAGE_NAME}} development guidance covering {{PRIMARY_DOMAIN}} applications, {{LANGUAGE_PARADIGMS}} patterns, and production-ready development using the latest tools and frameworks.

## What It Does

{{#if BACKEND_FOCUS}}
- **Backend API Development**: {{WEB_FRAMEWORKS}} applications with modern patterns
- **Database Integration**: {{DATABASE_FRAMEWORKS}} with query optimization
- **API Development**: REST and GraphQL services with authentication
- **Testing & Quality**: {{TEST_FRAMEWORK}} with mocking and integration tests
- **Performance Optimization**: Caching, async patterns, memory management
- **Production Deployment**: Docker, CI/CD, monitoring, scaling
- **Security Best Practices**: Input validation, authentication, dependency scanning
- **Code Quality**: Static analysis, formatting, pre-commit hooks
{{/if}}

{{#if MOBILE_FOCUS}}
- **Mobile App Development**: {{MOBILE_FRAMEWORKS}} applications with platform-specific patterns
- **UI/UX Integration**: Modern UI frameworks and design systems
- **Platform Services**: Native APIs, push notifications, app store integration
- **Testing & Quality**: Unit tests, UI tests, device testing
- **Performance Optimization**: Memory management, battery optimization
- **Production Deployment**: App store releases, CI/CD, analytics
- **Security Best Practices**: Data encryption, authentication, secure storage
- **Code Quality**: Linting, formatting, modular architecture
{{/if}}

{{#if SYSTEMS_FOCUS}}
- **Systems Programming**: Low-level development with performance optimization
- **Memory Management**: Smart pointers, RAII, resource management
- **Performance Engineering**: Profiling, optimization, benchmarks
- **Cross-Platform Development**: Multi-platform builds and toolchains
- **Testing & Quality**: Unit tests, integration tests, property-based testing
- **Library Development**: API design, ABI compatibility, versioning
- **Security Best Practices**: Memory safety, input validation, secure coding
- **Code Quality**: Static analysis, formatting, code review
{{/if}}

{{#if DATA_FOCUS}}
- **Data Processing**: {{DATA_FRAMEWORKS}} for analytics and manipulation
- **Statistical Computing**: {{STATISTICS_FRAMEWORKS}} with modern methods
- **Data Visualization**: {{VISUALIZATION_FRAMEWORKS}} for insights and reporting
- **Database Integration**: SQL and NoSQL database connectivity
- **Testing & Quality**: Unit tests, data validation, reproducible research
- **Performance Optimization**: Vectorized operations, parallel processing
- **Production Deployment**: Scheduled jobs, monitoring, data pipelines
- **Code Quality**: Documentation, reproducible workflows, version control
{{/if}}

## When to Use

### Perfect Scenarios
{{#if BACKEND_FOCUS}}
- **Building REST APIs and microservices**
- **Developing web applications and services**
- **Creating database-driven applications**
- **Integrating with third-party services**
- **Automating backend workflows**
- **Deploying scalable server applications**
- **Enterprise application development**
{{/if}}

{{#if MOBILE_FOCUS}}
- **Building mobile applications**
- **Creating platform-specific user interfaces**
- **Integrating with device hardware and sensors**
- **Developing cross-platform mobile apps**
- **Implementing push notifications and background tasks**
- **Publishing to app stores**
- **Mobile-first application development**
{{/if}}

{{#if SYSTEMS_FOCUS}}
- **Performance-critical applications**
- **Systems programming and embedded development**
- **Game development and graphics programming**
- **Library and framework development**
- **Cross-platform application development**
- **Memory-sensitive applications**
- **High-performance computing**
{{/if}}

{{#if DATA_FOCUS}}
- **Data analysis and statistical computing**
- **Machine learning and data science workflows**
- **Business intelligence and reporting**
- **Data visualization and dashboarding**
- **Research and academic computing**
- **Big data processing and analytics**
- **Statistical modeling and forecasting**
{{/if}}

### Common Triggers
- "Create {{LANGUAGE_NAME}} {{PRIMARY_DOMAIN}}"
- "Set up {{LANGUAGE_NAME}} project"
- "Optimize {{LANGUAGE_NAME}} performance"
- "Test {{LANGUAGE_NAME}} code"
- "Deploy {{LANGUAGE_NAME}} application"
- "{{LANGUAGE_NAME}} best practices"
- "{{LANGUAGE_NAME}} {{WEB_FRAMEWORKS}} application"

## Tool Version Matrix (2025-11-06)

### Core {{LANGUAGE_NAME}}
- **{{LANGUAGE_NAME}}**: {{LATEST_VERSION}} (current) / {{LTS_VERSION}} (LTS)
- **Package Manager**: {{PACKAGE_MANAGER}} {{PACKAGE_MANAGER_VERSION}}
- **Runtime/Compiler**: {{RUNTIME_COMPILER}} {{RUNTIME_VERSION}}

{{#if WEB_FRAMEWORKS_AVAILABLE}}
### Web Frameworks
{{#each WEB_FRAMEWORKS_AVAILABLE}}
- **{{name}}**: {{version}} - {{description}}
{{/each}}
{{/if}}

{{#if TESTING_TOOLS}}
### Testing Tools
{{#each TESTING_TOOLS}}
- **{{name}}**: {{version}} - {{description}}
{{/each}}
{{/if}}

{{#if CODE_QUALITY_TOOLS}}
### Code Quality
{{#each CODE_QUALITY_TOOLS}}
- **{{name}}**: {{version}} - {{description}}
{{/each}}
{{/if}}

{{#if SPECIALIZED_TOOLS}}
### Specialized Tools
{{#each SPECIALIZED_TOOLS}}
- **{{name}}**: {{version}} - {{description}}
{{/each}}
{{/if}}

## Ecosystem Overview

### Package Management

{{#if PACKAGE_MANAGEMENT_COMMANDS}}
```bash
{{#each PACKAGE_MANAGEMENT_COMMANDS}}
{{this}}
{{/each}}
```
{{/if}}

### Project Structure (2025 Best Practice)

```
{{PROJECT_STRUCTURE}}
```

## Modern Development Patterns

### {{LANGUAGE_TYPE_SYSTEM}} Best Practices

{{#if TYPE_EXAMPLES}}
{{#each TYPE_EXAMPLES}}
```{{../LANGUAGE_EXTENSION}}
{{this}}
```
{{/each}}
{{/if}}

### {{PRIMARY_PARADIGMS}} Patterns

{{#if PARADIGM_EXAMPLES}}
{{#each PARADIGM_EXAMPLES}}
```{{../LANGUAGE_EXTENSION}}
{{this}}
```
{{/each}}
{{/if}}

{{#if ASYNC_PATTERNS_AVAILABLE}}
### Asynchronous Programming Patterns

```{{LANGUAGE_EXTENSION}}
{{ASYNC_PATTERN_EXAMPLE}}
```
{{/if}}

### Data Structures and Patterns

{{#if DATA_STRUCTURE_EXAMPLES}}
{{#each DATA_STRUCTURE_EXAMPLES}}
```{{../LANGUAGE_EXTENSION}}
{{this}}
```
{{/each}}
{{/if}}

## Performance Considerations

{{#if PERFORMANCE_EXAMPLES}}
{{#each PERFORMANCE_EXAMPLES}}
### {{title}}

```{{../LANGUAGE_EXTENSION}}
{{code}}
```
{{/each}}
{{/if}}

### Profiling and Monitoring

{{#if PROFILING_COMMANDS}}
```bash
{{#each PROFILING_COMMANDS}}
{{this}}
{{/each}}
```
{{/if}}

## Testing Strategy

{{#if TESTING_CONFIGURATION}}
### {{TEST_FRAMEWORK}} Configuration

{{TESTING_CONFIGURATION}}
{{/if}}

### Modern Testing Patterns

{{#if TESTING_EXAMPLES}}
{{#each TESTING_EXAMPLES}}
```{{../LANGUAGE_EXTENSION}}
{{this}}
```
{{/each}}
{{/if}}

{{#if INTEGRATION_TESTING_AVAILABLE}}
### Integration Testing

```{{LANGUAGE_EXTENSION}}
{{INTEGRATION_TESTING_EXAMPLE}}
```
{{/if}}

## Security Best Practices

### Input Validation and Sanitization

{{#if SECURITY_EXAMPLES}}
{{#each SECURITY_EXAMPLES}}
```{{../LANGUAGE_EXTENSION}}
{{this}}
```
{{/each}}
{{/if}}

{{#if AUTHENTICATION_PATTERNS}}
### Authentication and Authorization

```{{LANGUAGE_EXTENSION}}
{{AUTHENTICATION_PATTERNS}}
```
{{/if}}

{{#if SECURITY_TOOLS}}
### Dependency Security Scanning

```bash
{{#each SECURITY_TOOLS}}
{{this}}
{{/each}}
```
{{/if}}

## Integration Patterns

{{#if INTEGRATION_EXAMPLES}}
{{#each INTEGRATION_EXAMPLES}}
### {{title}}

```{{../LANGUAGE_EXTENSION}}
{{code}}
```
{{/each}}
{{/if}}

## Modern Development Workflow

{{#if PROJECT_CONFIGURATION}}
### Project Configuration

{{PROJECT_CONFIGURATION}}
{{/if}}

{{#if PRECOMMIT_CONFIG}}
### Pre-commit Configuration

```yaml
{{PRECOMMIT_CONFIG}}
```
{{/if}}

{{#if DOCKER_EXAMPLE}}
### Docker Best Practices

```dockerfile
{{DOCKER_EXAMPLE}}
```
{{/if}}

{{#if DOMAIN_SPECIFIC_PATTERNS}}
## {{PRIMARY_DOMAIN}} Development

{{#each DOMAIN_SPECIFIC_PATTERNS}}
### {{title}}

```{{../LANGUAGE_EXTENSION}}
{{code}}
```
{{/each}}
{{/if}}

---

**Created by**: MoAI Language Skill Factory  
**Last Updated**: 2025-11-06  
**Version**: 2.0.0  
**{{LANGUAGE_NAME}} Target**: {{LATEST_VERSION}} with modern {{PRIMARY_PARADIGMS}} features  

This skill provides comprehensive {{LANGUAGE_NAME}} development guidance with 2025 best practices, covering everything from basic project setup to advanced {{PRIMARY_DOMAIN}} integration and production deployment patterns.
