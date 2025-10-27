# Auto-Delegation Test Scenario Set (20 Scenarios)

## Overview
This document defines 20 test scenarios to validate the effectiveness of improved agent descriptions for auto-delegation in MoAI-ADK.

**Baseline Metrics**: Current ~70% auto-delegation success rate → Target 90%+ after improvements

---

## Category 1: Error Handling (5 Scenarios)

### Scenario 1: TypeError in Authentication Code
**Trigger**: "There's a TypeError in my authentication code"
- **Expected Agent**: debug-helper
- **Keyword Trigger**: "errors", "TypeError"
- **Before**: May not trigger debug-helper reliably
- **After**: Immediate debug-helper selection with "Actively use when errors occur"

### Scenario 2: Test Failure in JWT Validation
**Trigger**: "Tests are failing after I added JWT validation: AssertionError: token_expiry must be 30 minutes"
- **Expected Agent**: debug-helper → tdd-implementer (chain)
- **Keyword Trigger**: "test failures"
- **Before**: Ambiguous between multiple agents
- **After**: Clear chain: debug-helper diagnoses, then tdd-implementer fixes

### Scenario 3: Git Push Rejection
**Trigger**: "Git push rejected: non-fast-forward"
- **Expected Agent**: debug-helper
- **Keyword Trigger**: "Git errors"
- **Before**: May route to git-manager incorrectly
- **After**: debug-helper handles with "Must use for...Git errors"

### Scenario 4: ImportError in Module
**Trigger**: "ImportError: cannot import name 'Validator' from mymodule"
- **Expected Agent**: debug-helper
- **Keyword Trigger**: "ImportError", "errors"
- **Before**: ~60% accuracy
- **After**: >95% with "Must use for TypeError, ImportError"

### Scenario 5: Configuration Permission Error
**Trigger**: "PermissionError: [Errno 13] Permission denied: '.moai/config.json'"
- **Expected Agent**: debug-helper → cc-manager (possible chain)
- **Keyword Trigger**: "Configuration issues", "errors"
- **Before**: Unclear routing
- **After**: Clear: debug-helper for diagnosis, cc-manager for config fixes

---

## Category 2: Implementation (5 Scenarios)

### Scenario 6: Implement User Authentication with JWT
**Trigger**: "Implement user authentication with JWT tokens, 30-minute expiry, email+password login following SPEC-AUTH-001"
- **Expected Agent Chain**: spec-builder (if SPEC needed) → implementation-planner → tdd-implementer
- **Keyword Trigger**: "SPEC", "implementation", "TDD"
- **Before**: ~65% successful routing
- **After**: >90% with "Essential when code implementation follows SPEC requirements"

### Scenario 7: Write Failing Tests for Payment Module
**Trigger**: "Write failing tests for the payment processing module to start the RED phase"
- **Expected Agent**: tdd-implementer
- **Keyword Trigger**: "tests", "RED phase"
- **Before**: ~70% accuracy
- **After**: >95% with "Use proactively for RED-GREEN-REFACTOR workflows"

### Scenario 8: Refactor Database Connection Logic
**Trigger**: "Refactor the database connection logic to improve maintainability without changing functionality"
- **Expected Agent**: tdd-implementer
- **Keyword Trigger**: "refactor", "REFACTOR phase"
- **Before**: ~60% (might route to quality-gate)
- **After**: Clear with "refactoring for quality" in tdd-implementer description

### Scenario 9: Create New SPEC for Password Reset
**Trigger**: "Create a new SPEC for password reset feature with EARS syntax and acceptance criteria"
- **Expected Agent**: spec-builder
- **Keyword Trigger**: "SPEC", "EARS", "feature planning"
- **Before**: ~75% accuracy
- **After**: >95% with "Essential when starting new features"

### Scenario 10: Plan Implementation for SPEC-AUTH-001
**Trigger**: "Analyze SPEC-AUTH-001 and create implementation strategy with TAG chain and library versions"
- **Expected Agent**: implementation-planner
- **Keyword Trigger**: "implementation strategy", "TAG chain"
- **Before**: ~70% success
- **After**: Clear with improved description focusing on "architecture, selecting libraries, designing TAG chains"

---

## Category 3: Quality & Verification (5 Scenarios)

### Scenario 11: Check TRUST Principles Compliance
**Trigger**: "Check if my code follows TRUST 5 principles and show detailed report"
- **Expected Agent**: quality-gate (→ trust-checker internally)
- **Keyword Trigger**: "TRUST principles", "code quality"
- **Before**: ~65% (might go to trust-checker directly)
- **After**: >95% with "Actively use for verifying code quality against TRUST 5 principles"

### Scenario 12: Verify TAG Chain Integrity
**Trigger**: "Verify TAG chain integrity and detect orphan TAGs in the codebase"
- **Expected Agent**: tag-agent
- **Keyword Trigger**: "TAG", "traceability", "orphan"
- **Before**: ~80% accuracy (already good)
- **After**: >95% with improved description

### Scenario 13: Check Test Coverage Sufficiency
**Trigger**: "Is my test coverage sufficient? I need at least 80% for passing quality gate"
- **Expected Agent**: quality-gate
- **Keyword Trigger**: "test coverage"
- **Before**: ~75% (might go to tdd-implementer)
- **After**: >95% with "Must use for test coverage verification"

### Scenario 14: Code Quality Review Before Commit
**Trigger**: "Review code quality and identify issues before I commit these changes"
- **Expected Agent**: quality-gate
- **Keyword Trigger**: "before commits", "quality verification"
- **Before**: ~70% accuracy
- **After**: >90% with "Use proactively before commits"

### Scenario 15: Performance Regression Analysis
**Trigger**: "Check for performance regressions after optimizing the data processing pipeline"
- **Expected Agent**: quality-gate (→ trust-checker with performance analysis)
- **Keyword Trigger**: "performance", "quality"
- **Before**: ~60% (might route to debug-helper)
- **After**: >85% with improved trust-checker linkage description

---

## Category 4: Documentation & Git (5 Scenarios)

### Scenario 16: Sync Documentation with Code Changes
**Trigger**: "Synchronize documentation with recent code changes and update API reference"
- **Expected Agent**: doc-syncer
- **Keyword Trigger**: "documentation", "API references"
- **Before**: ~70% accuracy
- **After**: >95% with "Essential for keeping README, architecture docs, and API references current"

### Scenario 17: Create Feature Branch for SPEC-001
**Trigger**: "Create a feature branch for SPEC-AUTH-001 and set up Draft PR for team review"
- **Expected Agent**: git-manager
- **Keyword Trigger**: "Git", "branch creation", "PR"
- **Before**: ~85% (already good)
- **After**: >95% with "Actively use for all Git operations"

### Scenario 18: Update README with New API Endpoints
**Trigger**: "Update README with new API endpoints and integration examples"
- **Expected Agent**: doc-syncer
- **Keyword Trigger**: "documentation", "Living Document"
- **Before**: ~65% (might go to tdd-implementer)
- **After**: >90% with improved doc-syncer description

### Scenario 19: Convert PR to Ready for Review (Team Mode)
**Trigger**: "The implementation is complete, convert this Draft PR to Ready for Review"
- **Expected Agent**: git-manager
- **Keyword Trigger**: "Git", "PR", "GitFlow"
- **Before**: ~80%
- **After**: >95% with "Use proactively...managing PR workflows, handling PR transitions (Draft → Ready → Merge)"

### Scenario 20: Initialize New MoAI-ADK Project
**Trigger**: "Set up a new MoAI-ADK project with product documentation and project structure"
- **Expected Agent**: project-manager
- **Keyword Trigger**: "project initialization", "project setup"
- **Before**: ~75%
- **After**: >90% with improved project-manager description

---

## Validation Metrics & Scoring

### Success Rate Calculation
```
Success Rate = (Correct Agent Selected / Total Scenarios) × 100

Passing Thresholds:
- Phase 1 Target: 85% (17/20 correct)
- Phase 1 Success: ≥17/20
- Phase 2 Target: 90% (18/20 correct)
- Phase 3 Target: 95% (19/20 correct)
```

### Individual Agent Success Rate
Track success rate per agent:

| Agent | Baseline | Phase 1 Target | Current | Status |
|-------|----------|---|---|---|
| tdd-implementer | 70% | 85% | pending | ⏳ |
| debug-helper | 65% | 90% | pending | ⏳ |
| spec-builder | 75% | 95% | pending | ⏳ |
| quality-gate | 65% | 90% | pending | ⏳ |
| git-manager | 85% | 95% | pending | ⏳ |
| doc-syncer | 70% | 90% | pending | ⏳ |
| tag-agent | 80% | 95% | pending | ⏳ |
| project-manager | 75% | 90% | pending | ⏳ |

---

## Testing Procedure

### Step 1: Baseline Testing (Before Improvements)
1. Document current agent selection accuracy for each scenario
2. Record actual agent selected vs. expected agent
3. Calculate baseline success rate
4. Document reasons for failures

### Step 2: Apply Improvements
1. Update agent descriptions (Phase 1)
2. Implement chaining patterns (Phase 2)
3. Apply proactive triggers (Phase 3)

### Step 3: Retest (After Improvements)
1. Run same 20 scenarios
2. Measure new success rate
3. Compare before/after
4. Identify remaining issues

### Step 4: Iterate
- For failures in Phase 1 retest, refine descriptions
- Repeat testing until 85%+ success rate achieved

---

## Documentation of Results

### Phase 1 Results Template
```markdown
## Phase 1 Auto-Delegation Test Results

### Overall Metrics
- Success Rate: X/20 (X%)
- Target Met: ✅ YES / ❌ NO

### Agent-by-Agent Results
- tdd-implementer: Y/Z (%)
- debug-helper: Y/Z (%)
- spec-builder: Y/Z (%)
- quality-gate: Y/Z (%)

### Failed Scenarios
List scenarios that didn't route to expected agent:
- Scenario #: Expected X, Got Y (Reason: ...)

### Recommendations for Phase 2
- Description refinements needed
- Chaining patterns to implement
- Keywords to emphasize
```

---

## Notes

- **Token Impact**: Improved descriptions (slightly longer) but better auto-delegation accuracy
- **Risk Mitigation**: Descriptions follow official Claude Code documentation patterns
- **Backward Compatibility**: No breaking changes, only enhancement of existing descriptions
- **Multilingual Support**: English-only descriptions maintain language boundary principle

---

**Test Scenario Set Created**: 2025-10-28
**Last Updated**: 2025-10-28
**Status**: Ready for Phase 1 validation
