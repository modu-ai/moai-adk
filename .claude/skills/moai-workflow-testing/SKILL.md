---
name: moai-workflow-testing
description: >
  Comprehensive testing and development workflow specialist combining DDD
  testing, characterization tests, performance profiling, and quality
  assurance. Use when writing tests or measuring coverage.

when_to_use: >
  Use for comprehensive testing and QA: DDD domain-driven testing,
  characterization tests, behavior preservation, performance profiling,
  code and PR review, CI/CD, and TRUST 5 quality-assurance validation.

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Bash(pytest:*), Bash(ruff:*), Bash(npm:*), Bash(npx:*), Bash(node:*), Bash(jest:*), Bash(vitest:*), Bash(go:*), Bash(cargo:*), Bash(mix:*), Bash(uv:*), Bash(bundle:*), Bash(php:*), Bash(phpunit:*), Grep, Glob
user-invocable: false
metadata:
  version: "2.4.0"
  category: "workflow"
  status: "active"
  updated: "2026-01-21"
  modularized: "true"
  tags: "workflow, ddd, testing, debugging, performance, quality, review, pr-review"
  author: "MoAI-ADK Team"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
---

# Development Workflow Specialist

## Quick Reference

Unified development workflow combining DDD (domain-driven development) testing, debugging guidance, performance optimization, automated code review, and CI/CD quality gates. Emphasizes behavior preservation during refactoring through characterization tests.

Core Capabilities:

- DDD Testing: Characterization tests (legacy) + specification tests (greenfield) + behavior snapshots
- AI-Powered Debugging: Error analysis, classification, solution candidates
- Performance Optimization: Profiling, bottleneck detection, optimization recommendations
- Automated Code Review: TRUST 5 framework validation
- PR Code Review: Multi-agent pattern (Haiku eligibility + 5 Sonnet parallel reviewers)
- Quality Assurance: CI/CD integration with quality gates

Workflow Progression: Debug → Refactor → Optimize → Review → Test → Profile

When to Use:

- Complete development lifecycle management
- Quality assurance and CI/CD integration
- Multi-language and performance-critical projects
- Technical debt reduction
- PR code review automation

---

## Implementation Guide

### Core Concepts

Five integrated components form the workflow:

- AI-Powered Debugging: Error classification (syntax/runtime/logic/integration/performance) and solution candidates ranked by likelihood
- Smart Refactoring: Technical debt analysis with complexity metrics and risk assessment
- Performance Optimization: CPU/memory/IO/network profiling with optimization strategies
- DDD Testing Management: Characterization tests (PRESERVE phase) for legacy + specification tests for greenfield + TRUST 5 validation
- Automated Code Review: TRUST 5 framework with actionable recommendations

### TRUST 5 Framework

Quality assessment model with five dimensions:

- Testability: pure functions, injectable dependencies, modular design
- Readability: descriptive names, logical structure, documented complexity
- Understandability: clear business logic, appropriate abstractions, conceptual clarity
- Security: input validation, secret management, OWASP compliance (injection/XSS/CSRF)
- Transparency: comprehensive error handling, structured logs, traceable issues

Overall score: weighted average with critical-dimension override (security/testability cannot be masked).

See [TRUST 5 detailed dimensions and scoring](references/trust5-framework.md) for full assessment rubric.

### DDD Testing Process

Legacy Code (PRESERVE phase):

1. Write characterization tests documenting current behavior (not aspirational)
2. Organize tests by domain concepts to surface domain boundaries
3. Use behavior snapshots as regression safeguards for complex scenarios
4. Verify baseline: all characterization tests PASS before any change
5. Apply refactoring with continuous test execution
6. Run TRUST 5 validation post-refactor

Greenfield Development:

1. Derive specification tests from domain requirements (each test = business rule)
2. Organize tests by aggregates, entities, value objects (DDD ubiquitous language)
3. Specify behavior in business language, not implementation details
4. Implement to satisfy specifications
5. Verify with integration tests (domain interactions + invariants)
6. Apply TRUST 5 validation

### Debugging, Refactoring, Performance Workflows

All three follow a 6-step pattern: capture/analyze → classify → identify candidates → apply → verify → document.

See [debugging/refactoring/performance step-by-step walkthroughs](references/workflow-processes.md) for detailed process tables.

### Code Review Process

1. Scan codebase for review targets
2. Apply TRUST 5 framework per file
3. Identify critical issues
4. Calculate per-file + aggregate scores
5. Generate prioritized recommendations
6. Create summary report with improvement roadmap

### PR Code Review (Multi-Agent Pattern)

5-step multi-agent pipeline:

1. Eligibility Check (Haiku): skip closed/draft/already-reviewed/trivial PRs
2. Context Gathering: find CLAUDE.md per modified dir + summarize PR
3. Parallel Review (5 Sonnet agents): CLAUDE.md compliance / obvious bugs / git blame / previous comments / code comment compliance
4. Confidence Scoring (0-100): 0=false positive, 25=somewhat, 50=moderate, 75=high, 100=certain
5. Filter & Report: drop issues <80 confidence, post via gh CLI with file/line/commit links

See [PR review multi-agent architecture and output format](references/pr-review-multi-agent.md) for agent role detail and example output.

### Multi-Language Support

Per-language toolchain mappings (Python pytest+ruff+bandit, JS/TS Jest+ESLint+npm audit, Go go test+staticcheck+gosec, Rust cargo test+clippy+gosec equivalents).

See [multi-language toolchain reference](references/multi-language-support.md) for per-language testing/lint/security/perf tool inventory.

---

## Advanced Features

### Quality Gate Configuration

Three strictness modes:

- Strict: all TRUST dimensions ≥ threshold, zero critical issues, full coverage
- Standard: average score ≥ threshold, no critical issues blocking, warnings allowed
- Lenient: only critical blockers prevent progression

Gate config: per-dimension thresholds, max issues by severity, coverage targets, perf benchmarks.

### CI/CD Integration

Four-stage pipeline: Code Quality → Testing → Performance → Security. Each stage terminates pipeline on failure with stage-specific failure report.

See [CI/CD integration patterns (GitHub Actions + Docker)](references/integration-patterns.md) for job configuration walkthroughs.

---

## Works Well With

- moai-domain-backend: Backend testing patterns
- moai-domain-frontend: Frontend UI testing
- moai-foundation-core: SPEC system integration
- moai-platform-supabase / moai-platform-vercel / moai-platform-firebase-auth: Platform-specific testing
- moai-workflow-project: Project management workflows

---

Status: Production Ready
Last Updated: 2026-05-23
Maintained by: MoAI-ADK Development Workflow Team
Version: 2.4.0 (skill body compression pass)

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "This code is already covered by integration tests" | Integration tests catch different bugs than unit tests. The testing pyramid exists for a reason. |
| "Mocking the database is too hard, I will skip that test" | If the test is hard to write because of coupling, the code needs a better abstraction boundary. |
| "80% coverage is good enough" | Coverage targets are floors, not ceilings. The missing 20% often contains the error handling paths. |
| "These are just utility functions, they do not need tests" | Utility functions are the most reused code. A bug in a utility propagates everywhere. |
| "I ran the tests locally, CI will pass" | Environment differences cause CI-only failures. Trust CI output, not local runs. |
| "Flaky tests are normal, just re-run" | Flaky tests hide real failures. Fix the flakiness or quarantine the test explicitly. |

**Shift Left**: Find and fix defects as early as possible. Every test that runs in CI instead of locally adds latency. Every test that could have been a unit test but is an E2E test adds fragility.

**Beyonce Rule**: If you liked it, you should have put a test on it. Untested behavior is unspecified behavior.

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Coverage report shows decreased coverage after a feature addition
- Test file contains `t.Skip()` or `skip` without an accompanying issue tracker link
- Test names are auto-generated (test_1, test_2) instead of behavior-descriptive
- No test touches the error/failure branch of a new function
- Test file imports the concrete implementation instead of the interface

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Test suite passes with zero failures (paste command output)
- [ ] Coverage report generated and meets the 85% threshold for changed packages
- [ ] Error paths have dedicated test cases (not just happy path)
- [ ] No flaky tests introduced (run with -count=3 to verify stability)
- [ ] Test isolation confirmed: each test uses its own fixtures or t.TempDir()
- [ ] Race detector passed for concurrent code (go test -race or equivalent)

<!-- moai:evolvable-end -->
