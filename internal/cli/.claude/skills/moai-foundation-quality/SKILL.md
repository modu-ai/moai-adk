---
name: moai-foundation-quality
description: >
  Code quality orchestrator enforcing TRUST 5 validation, proactive code analysis,
  linting standards, and automated best practices.
  Use when performing code review, quality gate checks, lint configuration,
  TRUST 5 compliance validation, or establishing coding standards.
  Do NOT use for writing tests (use moai-workflow-testing instead)
  or debugging runtime errors (use expert-debug agent instead).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.2.0"
  category: "foundation"
  status: "active"
  updated: "2026-01-11"
  modularized: "true"
  tags: "foundation, quality, testing, validation, trust-5, best-practices, code-review"
  aliases: "moai-foundation-quality"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords:
    - "quality"
    - "testing"
    - "test"
    - "validation"
    - "trust-5"
    - "best practice"
    - "code review"
    - "linting"
    - "coverage"
    - "pytest"
    - "security"
    - "ci/cd"
    - "quality gate"
    - "proactive"
    - "code smell"
    - "technical debt"
    - "refactoring"
  agents:
    - "manager-quality"
    - "manager-ddd"
    - "expert-testing"
    - "expert-security"
    - "expert-refactoring"
  phases:
    - "run"
    - "sync"
  languages:
    - "python"
    - "javascript"
    - "typescript"
    - "java"
    - "go"
    - "rust"
    - "cpp"
    - "csharp"
---

# Enterprise Code Quality Orchestrator

Enterprise-grade code quality management system that combines systematic code review, proactive improvement suggestions, and automated best practices enforcement. Provides comprehensive quality assurance through TRUST 5 framework validation with Context7 integration for real-time best practices.

## Quick Reference (30 seconds)

Core Capabilities:

- TRUST 5 Validation: Testable, Readable, Unified, Secured, Trackable quality gates
- Proactive Analysis: Automated issue detection and improvement suggestions
- Best Practices Enforcement: Context7-powered real-time standards validation
- Multi-Language Support: 25+ programming languages with specialized rules
- Enterprise Integration: CI/CD pipelines, quality metrics, reporting

Key Patterns:

- Quality Gate Pipeline: Automated validation with configurable thresholds
- Proactive Scanner: Continuous analysis with improvement recommendations
- Best Practices Engine: Context7-driven standards enforcement
- Quality Metrics Dashboard: Comprehensive reporting and trend analysis

When to Use:

- Code review automation and quality gate enforcement
- Proactive code quality improvement and technical debt reduction
- Enterprise coding standards enforcement and compliance validation
- CI/CD pipeline integration with automated quality checks

Quick Access:

- TRUST 5 Framework: See [trust5-validation.md](modules/trust5-validation.md)
- Proactive Analysis: See [proactive-analysis.md](modules/proactive-analysis.md)
- Best Practices: See [best-practices.md](modules/best-practices.md)
- Integration Patterns: See [integration-patterns.md](modules/integration-patterns.md)

## Implementation Guide

### Getting Started

Basic Quality Validation: Initialize QualityOrchestrator with trust5_enabled, proactive_analysis, best_practices_enforcement, and context7_integration all set to True. Call analyze_codebase method with path parameter set to source directory, languages list including python, javascript, and typescript, and quality_threshold of 0.85. The method returns comprehensive quality results.

For quality gate validation with TRUST 5, create QualityGate instance and call validate_trust5 with codebase_path, test_coverage_threshold of 0.90, and complexity_threshold of 10.

Proactive Quality Analysis: Initialize ProactiveQualityScanner with context7_client and BestPracticesEngine rule_engine. Call scan_codebase with path and scan_types list including security, performance, maintainability, and testing. Generate recommendations by calling generate_recommendations with issues, priority set to high, and auto_fix enabled.

### Core Components

#### Quality Orchestration Engine

The QualityOrchestrator class provides enterprise quality orchestration with TRUST 5 framework. Initialize with QualityConfig and create instances of TRUST5Validator, ProactiveScanner, BestPracticesEngine, Context7Client, and QualityMetricsCollector.

The analyze_codebase method performs comprehensive analysis in four phases. Phase 1 runs TRUST 5 validation on the codebase with specified thresholds. Phase 2 performs proactive analysis scanning focus areas. Phase 3 checks best practices for specified languages with Context7 docs enabled. Phase 4 collects comprehensive metrics from all analysis results.

The method returns QualityResult containing trust5_validation, proactive_analysis, best_practices, metrics, and overall_score calculated from all results.

Detailed implementations available in modules:

- TRUST 5 Validator Implementation in [trust5-validation.md](modules/trust5-validation.md)
- Proactive Scanner Implementation in [proactive-analysis.md](modules/proactive-analysis.md)
- Best Practices Engine Implementation in [best-practices.md](modules/best-practices.md)

### Configuration and Customization

Quality Configuration: Create quality-config.yaml with quality_orchestration section.

Under trust5_framework, set enabled to true with thresholds for overall (0.85), testable (0.90), readable (0.80), unified (0.85), secured (0.90), and trackable (0.80).

Under proactive_analysis, set enabled true, scan_frequency to daily, and focus_areas list including performance, security, maintainability, and technical_debt.

Under auto_fix, set enabled true, severity_threshold to medium, and confirmation_required to true.

Under best_practices, set enabled true, context7_integration true, auto_update_standards true, and compliance_target to 0.85.

Under language_rules, configure python with pep8 style_guide, black formatter, ruff linter, and mypy type_checker. Configure javascript with airbnb style_guide, prettier formatter, and eslint linter. Configure typescript with google style_guide, prettier formatter, and eslint linter.

Under reporting, set enabled true, metrics_retention_days to 90, trend_analysis true, and executive_dashboard true.

Under notifications, enable quality_degradation, security_vulnerabilities, and technical_debt_increase.

Integration Examples: See [Integration Patterns](modules/integration-patterns.md) for CI/CD Pipeline Integration, GitHub Actions Integration, Quality-as-Service REST API, and Cross-Project Benchmarking.

## Advanced Patterns

### Custom Quality Rules

Create CustomQualityRule class with name, validator callable, and severity defaulting to medium. The validate async method executes the validator on codebase, wrapping in try-except. On success, return RuleResult with rule_name, passed status, severity, details, and recommendations. On exception, return RuleResult with passed false, severity error, error details, and fix recommendation.

See [Best Practices - Custom Rules](modules/best-practices.md#custom-quality-rules) for complete examples.

### Machine Learning Quality Prediction

ML-powered quality issue prediction using code feature extraction and predictive models. See [Proactive Analysis - ML Prediction](modules/proactive-analysis.md#machine-learning-quality-prediction) for implementation details.

### Real-time Quality Monitoring

Continuous quality monitoring with automated alerting for quality degradation and security vulnerabilities. See [Proactive Analysis - Real-time Monitoring](modules/proactive-analysis.md#real-time-quality-monitoring) for implementation details.

### Cross-Project Quality Benchmarking

Compare project quality metrics against similar projects in your industry. See [Integration Patterns - Benchmarking](modules/integration-patterns.md#cross-project-quality-benchmarking) for implementation details.

## Module Reference

### Core Modules

- [TRUST 5 Validation](modules/trust5-validation.md) - Comprehensive quality framework validation
- [Proactive Analysis](modules/proactive-analysis.md) - Automated issue detection and improvements
- [Best Practices](modules/best-practices.md) - Context7-powered standards enforcement
- [Integration Patterns](modules/integration-patterns.md) - CI/CD and enterprise integrations

### Key Components by Module

TRUST 5 Validation: TRUST5Validator for five-pillar quality validation, TestableValidator for test coverage and quality, SecuredValidator for security and OWASP compliance, and quality gate pipeline integration.

Proactive Analysis: ProactiveQualityScanner for automated issue detection, QualityPredictionEngine for ML-powered predictions, RealTimeQualityMonitor for continuous monitoring, and performance and maintainability analysis.

Best Practices: BestPracticesEngine for standards validation, Context7 integration for latest docs, custom quality rules, and language-specific validators.

Integration Patterns: CI/CD pipeline integration, GitHub Actions workflows, Quality-as-Service REST API, and cross-project benchmarking.

## Context7 Library Mappings

Essential library mappings for quality analysis tools and frameworks. See [Best Practices - Library Mappings](modules/best-practices.md#context7-library-mappings) for complete list.

## Works Well With

Agents:

- core-planner - Quality requirements planning
- workflow-ddd - DDD implementation validation
- security-expert - Security vulnerability analysis
- code-backend - Backend code quality
- code-frontend - Frontend code quality

Skills:

- moai-foundation-core - TRUST 5 framework reference
- moai-workflow-ddd - DDD workflow validation
- moai-security-owasp - Security compliance
- moai-context7-integration - Context7 best practices
- moai-performance-optimization - Performance analysis

Commands:

- /moai:2-run - DDD validation integration
- /moai:3-sync - Documentation quality checks
- /moai:9-feedback - Quality improvement feedback

## Quick Reference Summary

Core Capabilities: TRUST 5 validation, proactive scanning, Context7-powered best practices, multi-language support, enterprise integration

Key Classes: QualityOrchestrator, TRUST5Validator, ProactiveQualityScanner, BestPracticesEngine, QualityMetricsCollector

Essential Methods: analyze_codebase(), validate_trust5(), scan_for_issues(), validate_best_practices(), generate_quality_report()

Integration Ready: CI/CD pipelines, GitHub Actions, REST APIs, real-time monitoring, cross-project benchmarking

Enterprise Features: Custom rules, ML prediction, real-time monitoring, benchmarking, comprehensive reporting

Quality Standards: OWASP compliance, TRUST 5 framework, Context7 integration, automated improvement recommendations
