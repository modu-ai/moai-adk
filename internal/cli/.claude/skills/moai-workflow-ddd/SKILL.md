---
name: moai-workflow-ddd
description: >
  Domain-Driven Development workflow specialist using ANALYZE-PRESERVE-IMPROVE
  cycle for behavior-preserving code transformation.
  Use when refactoring legacy code, improving code structure without functional changes,
  reducing technical debt, or performing API migration with behavior preservation.
  Do NOT use for writing new tests (use moai-workflow-testing instead)
  or creating new features from scratch (use expert-backend or expert-frontend instead).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Bash(git:*) Bash(pytest:*) Bash(ruff:*) Bash(npm:*) Bash(npx:*) Bash(node:*) Bash(uv:*) Bash(make:*) Bash(cargo:*) Bash(go:*) Bash(mix:*) Bash(bundle:*) Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-01-16"
  modularized: "true"
  tags: "workflow, refactoring, ddd, domain-driven, behavior-preservation, ast-grep, characterization-tests"
  author: "MoAI-ADK Team"
  context: "fork"
  agent: "manager-ddd"
  related-skills: "moai-tool-ast-grep, moai-workflow-testing, moai-foundation-quality"
---

# Domain-Driven Development (DDD) Workflow

## Development Mode Configuration (CRITICAL)

[NOTE] This workflow is selected based on `.moai/config/sections/quality.yaml`:

```yaml
constitution:
  development_mode: hybrid    # or ddd, tdd
  hybrid_settings:
    new_features: tdd        # New code → use TDD
    legacy_refactoring: ddd  # Existing code → use DDD (this workflow)
```

**When to use this workflow**:
- `development_mode: ddd` → Always use DDD
- `development_mode: hybrid` + refactoring existing code → Use DDD
- `development_mode: hybrid` + new package/module → Use TDD instead (moai-workflow-tdd)

**Key distinction**:
- **New file/package** (doesn't exist yet) → TDD (RED-GREEN-REFACTOR)
- **Existing code** (file already exists) → DDD (ANALYZE-PRESERVE-IMPROVE)

## Quick Reference

Domain-Driven Development provides a systematic approach for refactoring existing codebases where behavior preservation is paramount. Unlike TDD which creates new functionality, DDD improves structure without changing behavior.

Core Cycle - ANALYZE-PRESERVE-IMPROVE:

- ANALYZE: Domain boundary identification, coupling metrics, AST structural analysis
- PRESERVE: Characterization tests, behavior snapshots, test safety net verification
- IMPROVE: Incremental structural changes with continuous behavior validation

When to Use DDD:

- Refactoring legacy code with existing tests
- Improving code structure without functional changes
- Technical debt reduction in production systems
- API migration and deprecation handling
- Code modernization projects
- When DDD is not applicable because code already exists
- Greenfield projects (with adapted cycle - see below)

When NOT to Use DDD:

- When behavior changes are required (modify SPEC first)

Greenfield Project Adaptation:

For new projects without existing code, DDD adapts its cycle:

- ANALYZE: Requirements analysis instead of code analysis
- PRESERVE: Define intended behavior through specification tests (test-first)
- IMPROVE: Implement code to satisfy the defined tests

This makes DDD a superset of TDD - it includes TDD's test-first approach while also supporting refactoring scenarios.

---

## Core Philosophy

### DDD vs TDD Comparison

TDD Approach (for new features):

- Cycle: RED-GREEN-REFACTOR
- Goal: Create new functionality through tests
- Starting Point: No code exists
- Test Type: Specification tests that define expected behavior
- Outcome: New working code with test coverage

DDD Approach (for refactoring):

- Cycle: ANALYZE-PRESERVE-IMPROVE
- Goal: Improve structure without behavior change
- Starting Point: Existing code with defined behavior
- Test Type: Characterization tests that capture current behavior
- Outcome: Better structured code with identical behavior

### Behavior Preservation Principle

The golden rule of DDD is that observable behavior must remain identical before and after refactoring. This means:

- All existing tests must pass unchanged
- API contracts remain identical
- Side effects remain identical
- Performance characteristics remain within acceptable bounds

---

## Implementation Guide

### Phase 1: ANALYZE

The analyze phase focuses on understanding the current codebase structure and identifying refactoring opportunities.

#### Domain Boundary Identification

Identify logical boundaries in the codebase by examining:

- Module dependencies and import patterns
- Data flow between components
- Shared state and coupling points
- Public API surfaces

Use AST-grep to analyze structural patterns. For Python, search for import patterns to understand module dependencies. For class hierarchies, analyze inheritance relationships and method distributions.

#### Coupling and Cohesion Metrics

Evaluate code quality metrics:

- Afferent Coupling (Ca): Number of classes depending on this module
- Efferent Coupling (Ce): Number of classes this module depends on
- Instability (I): Ce / (Ca + Ce) - higher means less stable
- Abstractness (A): Abstract classes / Total classes
- Distance from Main Sequence: |A + I - 1|

Low cohesion and high coupling indicate refactoring candidates.

#### Structural Analysis Patterns

Use AST-grep to identify problematic patterns:

- God classes with too many methods or responsibilities
- Feature envy where methods use other class data excessively
- Long parameter lists indicating missing abstractions
- Duplicate code patterns across modules

Create analysis reports documenting:

- Current architecture overview
- Identified problem areas with severity ratings
- Proposed refactoring targets with risk assessment
- Dependency graphs showing coupling relationships

### Phase 2: PRESERVE

The preserve phase establishes safety nets before making any changes.

#### Characterization Tests

Characterization tests capture existing behavior without assumptions about correctness. The goal is to document what the code actually does, not what it should do.

Steps for creating characterization tests:

- Step 1: Identify critical code paths through execution
- Step 2: Create tests that exercise these paths
- Step 3: Let tests fail initially to discover actual output
- Step 4: Update tests to expect actual output
- Step 5: Document any surprising behavior discovered

Characterization test naming convention: test*characterize*[component]\_[scenario]

#### Behavior Snapshots

For complex outputs, use snapshot testing to capture current behavior:

- API response snapshots
- Serialization output snapshots
- State transformation snapshots
- Error message snapshots

Snapshot files serve as behavior contracts during refactoring.

#### Test Safety Net Verification

Before proceeding to improvement phase, verify:

- All existing tests pass (100% green)
- New characterization tests cover refactoring targets
- Code coverage meets threshold for affected areas
- No flaky tests exist in the safety net

Run mutation testing if available to verify test effectiveness.

### Phase 3: IMPROVE

The improve phase makes structural changes while continuously validating behavior preservation.

#### Incremental Transformation Strategy

Never make large changes at once. Follow this pattern:

- Make smallest possible structural change
- Run full test suite
- If tests fail, revert immediately
- If tests pass, commit the change
- Repeat until refactoring goal achieved

#### Safe Refactoring Patterns

Extract Method: When a code block can be named and isolated. Use AST-grep to identify candidates by searching for repeated code blocks or long methods.

Extract Class: When a class has multiple responsibilities. Move related methods and fields to a new class while maintaining the original API through delegation.

Move Method: When a method uses data from another class more than its own. Relocate while preserving all call sites.

Inline Refactoring: When indirection adds complexity without benefit. Replace delegation with direct implementation.

Rename Refactoring: When names do not reflect current understanding. Update all references atomically using AST-grep rewrite.

#### AST-Grep Assisted Transformations

Use AST-grep for safe, semantic-aware transformations:

For method extraction, create a rule that identifies the code pattern and rewrites to the extracted form.

For API migration, create a rule that matches old API calls and rewrites to new API format.

For deprecation handling, create rules that identify deprecated patterns and suggest modern alternatives.

#### Continuous Validation Loop

After each transformation:

- Run unit tests (fast feedback)
- Run integration tests (behavior validation)
- Run characterization tests (snapshot comparison)
- Verify no new warnings or errors introduced
- Check performance benchmarks if applicable

---

## DDD Workflow Execution

### Standard DDD Session

When executing DDD through moai:2-run in DDD mode:

Step 1 - Initial Assessment:

- Read SPEC document for refactoring scope
- Identify affected files and components
- Assess current test coverage

Step 2 - Analyze Phase Execution:

- Run AST-grep analysis on target code
- Generate coupling and cohesion metrics
- Create domain boundary map
- Document refactoring opportunities

Step 3 - Preserve Phase Execution:

- Verify all existing tests pass
- Create characterization tests for uncovered paths
- Generate behavior snapshots
- Confirm safety net adequacy

Step 4 - Improve Phase Execution:

- Execute transformations incrementally
- Run tests after each change
- Commit successful changes immediately
- Document any discovered issues

Step 5 - Validation and Completion:

- Run full test suite
- Compare before/after metrics
- Verify all behavior snapshots match
- Generate refactoring report

### DDD Loop Pattern

For complex refactoring requiring multiple iterations:

- Set maximum loop iterations based on scope
- Each loop focuses on one refactoring target
- Exit conditions: all targets adddessed or iteration limit reached
- Progress tracking through TODO list updates

---

## Quality Metrics

### DDD Success Criteria

Behavior Preservation (Required):

- All pre-existing tests pass
- All characterization tests pass
- No API contract changes
- Performance within bounds

Structure Improvement (Goals):

- Reduced coupling metrics
- Improved cohesion scores
- Reduced code complexity
- Better separation of concerns

### DDD-Specific TRUST Validation

Apply TRUST 5 framework with DDD focus:

- Testability: Characterization test coverage adequate
- Readability: Naming and structure improvements verified
- Understandability: Domain boundaries clearer
- Security: No new vulnerabilities introduced
- Transparency: All changes documented and traceable

---

## Integration Points

### With AST-Grep Skill

DDD relies heavily on AST-grep for:

- Structural code analysis
- Pattern identification
- Safe code transformations
- Multi-file refactoring

Load moai-tool-ast-grep for detailed pattern syntax and rule creation.

### With Testing Workflow

DDD complements testing workflow:

- Uses characterization tests from testing patterns
- Integrates with mutation testing for safety net validation
- Leverages snapshot testing infrastructure

### With Quality Framework

DDD outputs feed into quality assessment:

- Before/after metrics comparison
- TRUST 5 validation for changes
- Technical debt tracking

---

## Troubleshooting

### Common Issues

Tests Fail After Transformation:

- Revert immediately to last known good state
- Analyze which tests failed and why
- Check if transformation changed behavior unintentionally
- Consider smaller transformation steps

Characterization Tests Are Flaky:

- Identify sources of non-determinism
- Mock external dependencies
- Fix time-dependent or order-dependent behavior
- Consider snapshot tolerance settings

Performance Degradation:

- Profile before and after
- Identify hot paths affected by changes
- Consider caching or optimization
- Document acceptable trade-offs

### Recovery Procedures

When DDD session encounters issues:

- Save current state with git stash
- Reset to last successful commit
- Review transformation that caused failure
- Plan alternative approach
- Resume from preserved state

---

Version: 1.0.0
Status: Active
Last Updated: 2026-01-16
