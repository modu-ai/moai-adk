# Development Workflow Testing Modules

> Purpose: Advanced implementation modules for comprehensive development workflow management
> Structure: Modular architecture with focused, in-depth implementations
> Compliance: Progressive disclosure with main SKILL.md under 500 lines

## Module Overview

This modules directory contains detailed implementation modules for the moai-workflow-testing skill. Each module provides comprehensive code examples, advanced features, and integration patterns that extend the core functionality described in the main SKILL.md file.

## Available Modules

### Root Level Modules

#### [AI-Powered Debugging](./ai-debugging.md)
Complexity: Advanced | Time: 20+ minutes | Dependencies: the host language's async and stack-introspection facilities

- Intelligent error classification with documentation-sourced patterns
- AI-driven solution generation with confidence scoring
- Learning debugger that improves from previous fixes
- Real-time error pattern recognition and prevention strategies

Key Features:
- WebSearch / WebFetch integration for latest debugging patterns
- Error frequency tracking and analysis
- Automated solution generation with multiple approaches
- Performance-aware debugging with minimal overhead

### [Smart Refactoring](./smart-refactoring.md)
Complexity: Advanced | Time: 25+ minutes | Dependencies: AST/refactoring tool for your language

- Technical debt analysis with comprehensive code scanning
- Safe automated refactoring with risk assessment
- AI-driven refactoring suggestions grounded in official docs
- Dependency-aware refactoring with impact analysis

Key Features:
- Documentation-grounded refactoring patterns
- Safe transformation planning with rollback strategies
- Technical debt prioritization and quantification
- Project-aware refactoring with convention detection

### [Performance Optimization](./performance-optimization.md)
Complexity: Advanced | Time: 30+ minutes | Dependencies: CPU/memory profiler for your language

- Real-time performance monitoring with configurable sampling
- Bottleneck detection with AI-powered analysis
- Automated optimization plan generation
- Memory leak detection and optimization strategies

Key Features:
- Multi-dimensional performance analysis
- Intelligent optimization suggestions grounded in official docs
- Continuous monitoring with alerting capabilities
- Performance regression detection and prevention

### [Automated Code Review](./automated-code-review.md)
Complexity: Advanced | Time: 35+ minutes | Dependencies: static analyzers for your language

- TRUST 5 framework validation with AI analysis
- Multi-tool static analysis integration and aggregation
- Documentation-grounded security patterns and vulnerability detection
- Automated fix suggestions with diff generation

Key Features:
- Comprehensive TRUST 5 category scoring
- Documentation-grounded security and quality pattern integration
- Automated issue detection with prioritization
- Integration with CI/CD pipelines and quality gates

### Thematic Subdirectories

#### [Automated Code Review](./automated-code-review/)
Comprehensive code review workflows with TRUST 5 framework integration.
- `review-workflows.md` - Code review workflow patterns
- `trust5-framework.md` - TRUST 5 framework overview
- `trust5-framework/` - TRUST 5 sub-components directory
  - `relevance-analysis.md` - Relevance dimension analysis
  - `safety-analysis.md` - Safety dimension analysis
  - `scoring-algorithms.md` - Scoring algorithm details
  - `timeliness-analysis.md` - Timeliness dimension analysis
  - `truthfulness-analysis.md` - Truthfulness dimension analysis
  - `usability-analysis.md` - Usability dimension analysis

#### [Code Review Patterns](./code-review/)
Code review patterns and methodologies.
- `analysis-patterns.md` - Code analysis patterns
- `core-classes.md` - Core code review classes
- `tool-integration.md` - Tool integration patterns

#### [Debugging Workflows](./debugging/)
AI-powered debugging workflows.
- `debugging-workflows.md` - Debugging workflow processes
- `error-analysis.md` - Error analysis techniques

#### [Performance Optimization](./performance/)
Performance optimization strategies.
- `optimization-patterns.md` - Performance optimization patterns
- `profiling-techniques.md` - Profiling and measurement techniques

#### [Refactoring Patterns](./refactoring/)
AI-powered refactoring workflows.
- `ai-workflows.md` - AI refactoring workflows
- `patterns.md` - Refactoring patterns

#### [Core DDD](./ddd/)
Core DDD documentation.
- `core-classes.md` - Core DDD classes and patterns

## Module Integration

These modules are conceptual references, not an importable SDK. Apply each
module's workflow steps directly with your project's own testing, linting, and
profiling toolchain. Per-language tool inventories live in
[../references/multi-language-support.md](../references/multi-language-support.md).

### Using Individual Modules

Each module describes a self-contained workflow (debug, refactor, optimize,
review, DDD test). Run the relevant one against your codebase using the
language-appropriate toolchain:

- AI-Powered Debugging: capture error + context, classify, propose fix candidates
- Performance Optimization: profile (CPU/memory/IO), detect bottlenecks, plan optimization
- Smart Refactoring: scan technical debt, plan safe transforms, verify with tests
- Automated Code Review: run static analysis, score against TRUST 5, prioritize issues
- DDD: ANALYZE-PRESERVE-IMPROVE, characterization tests first

### Unified Workflow Integration

The modules compose into a single development cycle. Drive them in sequence
using your own tooling — debug, then refactor, then optimize, then review, then
test — and validate each stage against TRUST 5 before proceeding. There is no
SDK to import; each stage maps to commands in your language's ecosystem (see the
multi-language reference linked above).

## Module Dependencies

### Core Dependencies
- WebSearch / WebFetch: for latest pattern integration and AI assistance (optional; modules degrade gracefully using established best-practice patterns when documentation is unreachable)
- The project's own test runner, linter, and profiler (language-dependent — see multi-language reference)

### Module-Specific Tools (examples per language)
- Performance Optimization: a CPU/memory profiler for your language (cProfile/memory_profiler for Python, pprof for Go, flamegraph for Rust, Chrome DevTools for JS/TS)
- Smart Refactoring: a refactoring/AST tool for your language (Rope for Python, gopls for Go, rust-analyzer for Rust)
- Automated Code Review: static analyzers for your language (pylint/flake8/bandit/mypy for Python, staticcheck/gosec for Go, clippy for Rust, ESLint for JS/TS)
- DDD: the project's test runner + coverage tool (pytest/coverage for Python, go test -cover for Go, cargo test for Rust, Jest for JS/TS)

## Best Practices

### Module Selection
1. Start with main SKILL.md: Use the overview to understand capabilities
2. Progress to modules: Dive into specific modules as needed
3. Combine selectively: Use only the modules relevant to your workflow

### Integration Guidelines
1. Documentation Lookup: Use WebSearch / WebFetch to ground AI suggestions in current official docs
2. Performance Considerations: Monitor overhead of analysis tools
3. Quality Gates: Configure appropriate thresholds for your project

### Maintenance
1. Regular Updates: Keep documentation references current
2. Tool Versions: Maintain compatible static analysis tool versions
3. Pattern Evolution: Update patterns as best practices evolve

## Module Development

### Adding New Modules

To add a new module to this skill:

1. Create Module File: Use the established template pattern
2. Follow Structure: Include core implementation, advanced features, and best practices
3. Update References: Add module reference to main SKILL.md
4. Test Integration: Ensure compatibility with existing modules

### Module Template

```markdown
# Module Title

> Module: Brief module description
> Complexity: Basic|Intermediate|Advanced
> Time: X+ minutes
> Dependencies: List of required libraries

## Core Implementation

[Complete implementation with comprehensive examples]

## Advanced Features

[Extended functionality and integration patterns]

## Best Practices

[Guidelines for production use]

---

Module: `modules/module-name.md`
Related: [Other Module](./other-module.md) | [Related Module](./related-module.md)
```

## Quality Assurance

### Module Standards
- Comprehensive documentation with examples
- Error handling and edge case coverage
- Performance considerations and optimizations
- Documentation-grounded patterns where appropriate
- Cross-module compatibility testing

### Validation Checklist
- [ ] Module compiles and runs without errors
- [ ] Examples are functional and tested
- [ ] Documentation is complete and accurate
- [ ] Integration with other modules works
- [ ] Performance meets acceptable standards

## Support and Contributing

### Module Support
- Documentation: Each module includes comprehensive usage examples
- Integration: See main SKILL.md for integration patterns
- Dependencies: Check individual modules for specific requirements

### Contributing
When contributing to modules:

1. Follow Templates: Use established module structure
2. Test Thoroughly: Ensure compatibility with existing modules
3. Document Completely: Include comprehensive examples and use cases
4. Update References: Keep main SKILL.md and README current

---

Last Updated: 2026-01-06
Module Count: 11 root-level modules + 6 thematic subdirectories
Maintained by: MoAI-ADK Development Workflow Team
