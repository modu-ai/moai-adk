---
name: moai-workflow-testing
description: >
  Comprehensive testing and development workflow specialist combining DDD testing,
  characterization tests, performance profiling, code review, and quality assurance.
  Use when writing tests, measuring coverage, creating characterization tests,
  performing TDD, running CI/CD quality checks, or reviewing pull requests.
  Do NOT use for debugging runtime errors (use expert-debug agent instead)
  or code refactoring (use moai-workflow-ddd instead).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Bash(pytest:*) Bash(ruff:*) Bash(npm:*) Bash(npx:*) Bash(node:*) Bash(jest:*) Bash(vitest:*) Bash(go:*) Bash(cargo:*) Bash(mix:*) Bash(uv:*) Bash(bundle:*) Bash(php:*) Bash(phpunit:*) Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.4.0"
  category: "workflow"
  status: "active"
  updated: "2026-01-21"
  modularized: "true"
  tags: "workflow, ddd, testing, debugging, performance, quality, review, pr-review"
  author: "MoAI-ADK Team"
  context: "fork"
  agent: "manager-ddd"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["DDD", "domain-driven development", "characterization tests", "behavior preservation", "debugging", "performance optimization", "code review", "PR review", "quality assurance", "testing", "CI/CD", "TRUST 5"]
  phases: ["run", "sync"]
  agents: ["manager-ddd", "expert-testing", "expert-debug", "expert-performance", "manager-quality"]
---

# Development Workflow Specialist

## Quick Reference

Unified Development Workflow provides comprehensive development lifecycle management combining DDD testing, AI-powered debugging, performance optimization, automated code review, and quality assurance into integrated workflows.

Core Capabilities:

- DDD Testing: Characterization tests for legacy code, specification tests for greenfield projects, behavior snapshots
- AI-Powered Debugging: Intelligent error analysis and solution recommendations
- Performance Optimization: Profiling and bottleneck detection guidance
- Automated Code Review: TRUST 5 validation framework for quality analysis
- PR Code Review: Multi-agent pattern with Haiku eligibility check and Sonnet parallel review
- Quality Assurance: Comprehensive testing and CI/CD integration patterns
- Workflow Orchestration: End-to-end development process guidance

Workflow Progression: Debug stage leads to Refactor stage, which leads to Optimize stage, then Review stage, followed by Test stage, and finally Profile stage. Each stage benefits from AI-powered analysis and recommendations.

When to Use:

- Complete development lifecycle management
- Enterprise-grade quality assurance implementation
- Multi-language development projects
- Performance-critical applications
- Technical debt reduction initiatives
- Automated testing and CI/CD integration
- Pull request code review automation

---

## Implementation Guide

### Core Concepts

Unified Development Philosophy:

- Integrates all aspects of development into cohesive workflow
- AI-powered assistance for complex decision-making
- Industry best practices integration for optimal patterns
- Continuous feedback loops between workflow stages
- Automated quality gates and validation

Workflow Components:

Component 1 - AI-Powered Debugging: The debugging component provides intelligent error classification and solution recommendations. When an error occurs, the system analyzes the error type, stack trace, and surrounding context to identify root causes and suggest appropriate fixes. The debugger references current best practices and common error resolution patterns.

Component 2 - Smart Refactoring: The refactoring component performs technical debt analysis and identifies safe automated transformation opportunities. It evaluates code complexity, duplication, and maintainability metrics to recommend specific refactoring actions with risk assessments.

Component 3 - Performance Optimization: The performance component provides real-time monitoring guidance and bottleneck detection. It helps identify CPU-intensive operations, memory leaks, and I/O bottlenecks, then recommends specific optimization strategies based on the identified issues.

Component 4 - DDD Testing Management: The DDD testing component provides specialized testing approaches aligned with domain-driven development. For legacy code, it uses characterization tests to capture current behavior before refactoring (PRESERVE phase). For greenfield projects, it uses specification tests derived from domain requirements. Behavior snapshots ensure behavioral consistency during transformations, and TRUST 5 validation ensures quality standards are maintained.

Component 5 - Automated Code Review: The code review component applies TRUST 5 framework validation with AI-powered quality analysis. It evaluates code against five trust dimensions and provides actionable improvement recommendations.

### TRUST 5 Framework

The TRUST 5 framework is a conceptual quality assessment model with five dimensions. This framework provides guidance for evaluating code quality, not an implemented module.

Dimension 1 - Testability: Evaluate whether the code can be effectively tested. Consider whether functions are pure and deterministic, whether dependencies are injectable, and whether the code is modular enough for unit testing. Scoring ranges from low testability requiring significant refactoring to high testability with excellent test coverage support.

Dimension 2 - Readability: Assess how easily the code can be understood by others. Consider whether variable and function names are descriptive, whether the code structure is logical, and whether complex operations are documented. Scoring evaluates naming conventions, code organization, and documentation quality.

Dimension 3 - Understandability: Evaluate the conceptual clarity of the implementation. Consider whether the business logic is clearly expressed, whether abstractions are appropriate, and whether a new developer can understand the code quickly. This goes beyond surface readability to assess architectural clarity.

Dimension 4 - Security: Assess security posture and vulnerability exposure. Consider whether inputs are validated, whether secrets are properly managed, and whether common vulnerability patterns are avoided including injection, XSS, and CSRF. Scoring evaluates adherence to security best practices.

Dimension 5 - Transparency: Evaluate operational visibility and debuggability. Consider whether error handling is comprehensive, whether logs are meaningful and structured, and whether issues can be traced through the system. Scoring assesses observability and troubleshooting capabilities.

Overall TRUST Score Calculation: The overall TRUST score combines all five dimensions using weighted averaging. Critical issues in any dimension can override the average, ensuring security or testability problems are not masked by high scores elsewhere. A passing score typically requires minimum thresholds in each dimension plus an acceptable weighted average.

### Basic Workflow Implementation

Debugging Workflow Process:

- Step 1: Capture the error with full context including stack trace, environment, and recent code changes
- Step 2: Classify the error type as syntax, runtime, logic, integration, or performance
- Step 3: Analyze the error pattern against known issue databases and best practices
- Step 4: Generate solution candidates ranked by likelihood of success
- Step 5: Apply the recommended fix and verify resolution
- Step 6: Document the issue and solution for future reference

Refactoring Workflow Process:

- Step 1: Analyze the target codebase for code smells and technical debt indicators
- Step 2: Calculate complexity metrics including cyclomatic complexity and coupling
- Step 3: Identify refactoring opportunities with associated risk levels
- Step 4: Generate a refactoring plan with prioritized actions
- Step 5: Apply refactoring transformations in safe increments
- Step 6: Verify behavior preservation through test execution

Performance Optimization Process:

- Step 1: Configure profiling for target metrics including CPU, memory, I/O, and network
- Step 2: Execute profiling runs under representative load conditions
- Step 3: Analyze profiling results to identify bottlenecks
- Step 4: Generate optimization recommendations with expected impact estimates
- Step 5: Apply optimizations in isolation to measure individual effects
- Step 6: Validate overall performance improvement

DDD Testing Process:

Legacy Code Testing (PRESERVE Phase):

- Step 1: Write characterization tests that capture the current behavior of the system. These tests document what the code actually does, not what it should do.
- Step 2: Organize characterization tests by domain concepts and business rules. Group related behaviors to identify potential domain boundaries.
- Step 3: Use behavior snapshots to record input-output pairs for complex scenarios. These serve as regression safeguards during refactoring.
- Step 4: Verify that all characterization tests pass before making any changes. This establishes a baseline of current behavior.
- Step 5: Apply refactoring transformations while continuously running characterization tests to ensure behavior preservation.
- Step 6: After refactoring, run TRUST 5 validation to ensure code quality standards are maintained.

Greenfield Development Testing:

- Step 1: Derive specification tests directly from domain requirements and use cases. Each test should express a business rule or domain invariant.
- Step 2: Organize tests by domain concepts (aggregates, entities, value objects) following DDD principles.
- Step 3: Write tests that specify domain behavior in business language, avoiding implementation details.
- Step 4: Implement domain logic to satisfy specification tests while maintaining ubiquitous language.
- Step 5: Verify behavior with integration tests that validate domain interactions and invariants.
- Step 6: Apply TRUST 5 validation to ensure testability, readability, and understandability of domain code.

Code Review Process:

- Step 1: Scan the codebase to identify files requiring review
- Step 2: Apply TRUST 5 framework analysis to each file
- Step 3: Identify critical issues requiring immediate attention
- Step 4: Calculate per-file and aggregate quality scores
- Step 5: Generate actionable recommendations with priority rankings
- Step 6: Create a summary report with improvement roadmap

PR Code Review Process:

- Step 1: Eligibility Check using Haiku agent to filter PRs, skipping closed, draft, already reviewed, and trivial changes
- Step 2: Gather Context by finding CLAUDE.md files in modified directories and summarizing PR changes
- Step 3: Parallel Review Agents using five Sonnet agents for independent analysis covering CLAUDE.md compliance, obvious bugs, git blame context, previous comments, and code comment compliance
- Step 4: Confidence Scoring from 0 to 100 for each detected issue where 0 indicates false positive, 25 indicates somewhat confident, 50 indicates moderately confident, 75 indicates highly confident, and 100 indicates absolutely certain
- Step 5: Filter and Report by removing issues below 80 confidence threshold and posting via gh CLI

### Common Use Cases

Enterprise Development Workflow: For enterprise applications, the workflow integrates quality gates at each stage. Before deployment, the code must pass minimum TRUST score thresholds, have zero critical issues identified, and meet required test coverage percentages. The quality gates configuration specifies minimum trust scores (typically 0.85), maximum allowed critical issues (typically zero), and required coverage levels (typically 80 percent).

Performance-Critical Applications: For performance-sensitive systems, the workflow emphasizes profiling and optimization stages. Performance thresholds define maximum acceptable response times, memory usage limits, and minimum throughput requirements. The workflow provides percentage improvement tracking and specific optimization recommendations.

---

## Advanced Features

### Workflow Integration Patterns

Continuous Integration Integration: The workflow integrates with CI/CD pipelines through a multi-stage validation process.

Stage 1 - Code Quality Validation: Run the code review component and verify results meet quality standards. If the quality check fails, the pipeline terminates with a quality failure report.

Stage 2 - Testing Validation: Execute the full test suite including unit, integration, and end-to-end tests. If any tests fail, the pipeline terminates with a test failure report.

Stage 3 - Performance Validation: Run performance tests and compare results against defined thresholds. If performance standards are not met, the pipeline terminates with a performance failure report.

Stage 4 - Security Validation: Execute security analysis including static analysis and dependency scanning. If critical vulnerabilities are found, the pipeline terminates with a security failure report.

Upon passing all stages, the pipeline generates a success report and proceeds to deployment.

### Quality Gate Configuration

Quality gates define the criteria that must be met at each workflow stage. Gates can be configured with different strictness levels.

Strict Mode: All quality dimensions must meet or exceed thresholds. Any critical issue blocks progression. Full test coverage requirements apply.

Standard Mode: Average quality score must meet threshold. Critical issues block progression, but warnings are allowed. Standard coverage requirements apply.

Lenient Mode: Only critical blocking issues prevent progression. Quality scores generate warnings but do not block. Reduced coverage requirements apply.

Gate configuration includes threshold values for each TRUST dimension, maximum allowed issues by severity, required test coverage levels, and performance benchmark targets.

### Multi-Language Support

The workflow supports development across multiple programming languages with language-specific adaptations.

Python Projects: Integration with pytest for testing, pylint and flake8 for static analysis, bandit for security scanning, and cProfile or memory_profiler for performance analysis.

JavaScript and TypeScript Projects: Integration with Jest or Vitest for testing, ESLint for static analysis, npm audit for security scanning, and Chrome DevTools or lighthouse for performance analysis.

Go Projects: Integration with go test for testing, golint and staticcheck for static analysis, gosec for security scanning, and pprof for performance analysis.

Rust Projects: Integration with cargo test for testing, clippy for static analysis, cargo audit for security scanning, and flamegraph for performance analysis.

### PR Code Review Multi-Agent Pattern

The PR Code Review process uses a multi-agent architecture following the official Claude Code plugin pattern.

Eligibility Check Agent using Haiku: The Haiku agent performs lightweight filtering to avoid unnecessary reviews. It checks the PR state and metadata to determine if review is warranted. Skip conditions include closed PRs, draft PRs, PRs already reviewed by bot, trivial changes like typo fixes, and automated dependency updates.

Context Gathering: Before launching review agents, the system gathers relevant context by finding CLAUDE.md files in directories containing modified code to understand project-specific coding standards. It also generates a concise summary of PR changes including files modified, lines added or removed, and overall impact assessment.

Parallel Review Agents using five Sonnet instances: Five Sonnet agents run in parallel, each focusing on a specific review dimension. Agent 1 audits CLAUDE.md compliance checking for violations of documented coding standards and conventions. Agent 2 scans for obvious bugs including logic errors, null reference risks, and resource leaks. Agent 3 provides git blame and history context to identify recent changes and potential patterns. Agent 4 checks previous PR comments for recurring issues and unresolved feedback. Agent 5 validates code comment compliance ensuring comments are accurate and helpful.

Confidence Scoring System: Each detected issue receives a confidence score from 0 to 100. A score of 0 indicates a false positive with no confidence. A score of 25 means somewhat confident but might be real. A score of 50 indicates moderately confident that the issue is real but minor. A score of 75 means highly confident that the issue is very likely real. A score of 100 indicates absolutely certain that the issue is definitely real.

Filter and Report Stage: Issues below the 80 confidence threshold are filtered out to reduce noise. Remaining issues are formatted and posted to the PR using the GitHub CLI. The output format follows a standardized markdown structure with issue count, numbered list of issues with descriptions, and direct links to code with specific commit SHA and line range.

Example PR Review Output: The review output begins with a Code review header, followed by the count of found issues. Each issue is numbered and includes a description of the problem with reference to the relevant CLAUDE.md rule, followed by a link to the specific file, line range, and commit SHA in the pull request.

---

## Works Well With

- moai-domain-backend: Backend development workflows and API testing patterns
- moai-domain-frontend: Frontend development workflows and UI testing strategies
- moai-foundation-core: Core SPEC system and workflow management integration
- moai-platform-supabase: Supabase-specific testing patterns and database testing
- moai-platform-vercel: Vercel deployment testing and edge function validation
- moai-platform-firebase-auth: Firebase authentication testing patterns
- moai-workflow-project: Project management and documentation workflows

---

## Technology Stack Reference

The workflow leverages industry-standard tools for each capability area.

Analysis Libraries: cProfile provides Python profiling and performance analysis. memory_profiler enables memory usage analysis and optimization. psutil supports system resource monitoring. line_profiler offers line-by-line performance profiling.

Static Analysis Tools: pylint performs comprehensive code analysis and quality checks. flake8 enforces style guide compliance and error detection. bandit scans for security vulnerabilities. mypy validates static types.

Testing Frameworks: pytest provides advanced testing with fixtures and plugins. unittest offers standard library testing capabilities. coverage measures code coverage and identifies untested paths.

---

## Integration Patterns

### GitHub Actions Integration

The workflow integrates with GitHub Actions through a multi-step job configuration.

Job Configuration Steps:

- Step 1: Check out the repository using actions/checkout
- Step 2: Set up the Python environment using actions/setup-python with the target Python version
- Step 3: Install project dependencies including testing and analysis tools
- Step 4: Execute the quality validation workflow with strict quality gates
- Step 5: Run the test suite with coverage reporting
- Step 6: Perform performance benchmarking against baseline metrics
- Step 7: Execute security scanning and vulnerability detection
- Step 8: Upload workflow results as job artifacts for review

The job can be configured to run on push and pull request events, with matrix testing across multiple Python versions if needed.

### Docker Integration

For containerized environments, the workflow executes within Docker containers.

Container Configuration: Base the image on a Python slim variant for minimal size. Install project dependencies from requirements file. Copy project source code into the container. Configure entrypoint to execute the complete workflow sequence. Mount volumes for result output if persistent storage is needed.

The containerized workflow ensures consistent execution environments across development, testing, and production systems.

---

Status: Production Ready
Last Updated: 2026-01-21
Maintained by: MoAI-ADK Development Workflow Team
Version: 2.4.0 (DDD Testing Methodology)
