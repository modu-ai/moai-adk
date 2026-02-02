---
name: moai-workflow-spec
description: >
  SPEC workflow orchestration with EARS format requirements, acceptance criteria,
  and Plan-Run-Sync integration for MoAI-ADK development methodology.
  Use when creating SPEC documents, writing EARS requirements, defining acceptance
  criteria, planning features, or orchestrating the /moai plan phase.
  Do NOT use for implementation (use moai-workflow-ddd instead)
  or documentation generation (use moai-workflow-project instead).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Bash(git:*) Bash(ls:*) Bash(wc:*) Bash(mkdir:*) Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.2.0"
  category: "workflow"
  status: "active"
  updated: "2026-01-08"
  modularized: "true"
  tags: "workflow, spec, ears, requirements, moai-adk, planning"
  author: "MoAI-ADK Team"
  context: "fork"
  agent: "Plan"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords:
    [
      "SPEC",
      "requirement",
      "EARS",
      "acceptance criteria",
      "user story",
      "planning",
      "specification",
      "requirements gathering",
    ]
  phases: ["plan"]
  agents: ["manager-spec", "manager-strategy", "Plan"]
---

# SPEC Workflow Management

## Quick Reference (30 seconds)

SPEC Workflow Orchestration - Comprehensive specification management using EARS format for systematic requirement definition and Plan-Run-Sync workflow integration.

Core Capabilities:

- EARS Format Specifications: Five requirement patterns for clarity
- Requirement Clarification: Four-step systematic process
- SPEC Document Templates: Standardized structure for consistency
- Plan-Run-Sync Integration: Seamless workflow connection
- Parallel Development: Git Worktree-based SPEC isolation
- Quality Gates: TRUST 5 framework validation

EARS Five Patterns:

- Ubiquitous: The system shall always perform action - Always active
- Event-Driven: WHEN event occurs THEN action executes - Trigger-response
- State-Driven: IF condition is true THEN action executes - Conditional behavior
- Unwanted: The system shall not perform action - Prohibition
- Optional: Where possible, provide feature - Nice-to-have

When to Use:

- Feature planning and requirement definition
- SPEC document creation and maintenance
- Parallel feature development coordination
- Quality assurance and validation planning

Quick Commands:

- Create new SPEC: /moai:1-plan "user authentication system"
- Create parallel SPECs with Worktrees: /moai:1-plan "login feature" "signup feature" --worktree
- Create SPEC with new branch: /moai:1-plan "payment processing" --branch
- Update existing SPEC: /moai:1-plan SPEC-001 "add OAuth support"

---

## Implementation Guide (5 minutes)

### Core Concepts

SPEC-First Development Philosophy:

- EARS format ensures unambiguous requirements
- Requirement clarification prevents scope creep
- Systematic validation through test scenarios
- Integration with DDD workflow for implementation
- Quality gates enforce completion criteria
- Constitution reference ensures project-wide consistency

### Constitution Reference (SDD 2025 Standard)

Constitution defines the project DNA that all SPECs must respect. Before creating any SPEC, verify alignment with project constitution defined in `.moai/project/tech.md`.

Constitution Components:

- Technology Stack: Required versions and frameworks
- Naming Conventions: Variable, function, and file naming standards
- Forbidden Libraries: Libraries explicitly prohibited with alternatives
- Architectural Patterns: Layering rules and dependency directions
- Security Standards: Authentication patterns and encryption requirements
- Logging Standards: Log format and structured logging requirements

Constitution Verification:

- All SPEC technology choices must align with Constitution stack versions
- No SPEC may introduce forbidden libraries or patterns
- SPEC must follow naming conventions defined in Constitution
- SPEC must respect architectural boundaries and layering

WHY: Constitution prevents architectural drift and ensures maintainability
IMPACT: SPECs aligned with Constitution reduce integration conflicts significantly

### SPEC Workflow Stages

Stage 1 - User Input Analysis: Parse natural language feature description
Stage 2 - Requirement Clarification: Four-step systematic process
Stage 3 - EARS Pattern Application: Structure requirements using five patterns
Stage 4 - Success Criteria Definition: Establish completion metrics
Stage 5 - Test Scenario Generation: Create verification test cases
Stage 6 - SPEC Document Generation: Produce standardized markdown output

### EARS Format Deep Dive

Ubiquitous Requirements - Always Active:

- Use case: System-wide quality attributes
- Examples: Logging, input validation, error handling
- Test strategy: Include in all feature test suites as common verification

Event-Driven Requirements - Trigger-Response:

- Use case: User interactions and inter-system communication
- Examples: Button clicks, file uploads, payment completions
- Test strategy: Event simulation with expected response verification

State-Driven Requirements - Conditional Behavior:

- Use case: Access control, state machines, conditional business logic
- Examples: Account status checks, inventory verification, permission checks
- Test strategy: State setup with conditional behavior verification

Unwanted Requirements - Prohibited Actions:

- Use case: Security vulnerabilities, data integrity protection
- Examples: No plaintext passwords, no unauthorized access, no PII in logs
- Test strategy: Negative test cases with prohibited behavior verification

Optional Requirements - Enhancement Features:

- Use case: MVP scope definition, feature prioritization
- Examples: OAuth login, dark mode, offline mode
- Test strategy: Conditional test execution based on implementation status

### Requirement Clarification Process

Step 0 - Assumption Analysis (Philosopher Framework):

Before defining scope, surface and validate underlying assumptions using AskUserQuestion.

Assumption Categories:

- Technical Assumptions: Technology capabilities, API availability, performance characteristics
- Business Assumptions: User behavior, market requirements, timeline feasibility
- Team Assumptions: Skill availability, resource allocation, knowledge gaps
- Integration Assumptions: Third-party service reliability, compatibility expectations

Assumption Documentation:

- Assumption Statement: Clear description of what is assumed
- Confidence Level: High, Medium, or Low based on evidence
- Evidence Basis: What supports this assumption
- Risk if Wrong: Consequence if assumption proves false
- Validation Method: How to verify before committing significant effort

Step 0.5 - Root Cause Analysis:

For feature requests or problem-driven SPECs, apply Five Whys:

- Surface Problem: What is the user observing or requesting?
- First Why: What immediate need drives this request?
- Second Why: What underlying problem creates that need?
- Third Why: What systemic factor contributes?
- Root Cause: What fundamental issue must the solution adddess?

Step 1 - Scope Definition:

- Identify supported authentication methods
- Define validation rules and constraints
- Determine failure handling strategy
- Establish session management approach

Step 2 - Constraint Extraction:

- Performance Requirements: Response time targets
- Security Requirements: OWASP compliance, encryption standards
- Compatibility Requirements: Supported browsers and devices
- Scalability Requirements: Concurrent user targets

Step 3 - Success Criteria Definition:

- Test Coverage: Minimum percentage target
- Response Time: Percentile targets (P50, P95, P99)
- Functional Completion: All normal scenarios pass verification
- Quality Gates: Zero linter warnings, zero security vulnerabilities

Step 4 - Test Scenario Creation:

- Normal Cases: Valid inputs with expected outputs
- Error Cases: Invalid inputs with error handling
- Edge Cases: Boundary conditions and corner cases
- Security Cases: Injection attacks, privilege escalation attempts

### Plan-Run-Sync Workflow Integration

PLAN Phase (/moai:1-plan):

- manager-spec agent analyzes user input
- EARS format requirements generation
- Requirement clarification with user interaction
- SPEC document creation in .moai/specs/ directory
- Git branch creation (optional --branch flag)
- Git Worktree setup (optional --worktree flag)

RUN Phase (/moai:2-run):

- manager-ddd agent loads SPEC document
- ANALYZE-PRESERVE-IMPROVE DDD cycle execution
- moai-workflow-testing skill reference for test patterns
- Domain Expert agent delegation (expert-backend, expert-frontend, etc.)
- Quality validation through manager-quality agent

SYNC Phase (/moai:3-sync):

- manager-docs agent synchronizes documentation
- API documentation generation from SPEC
- README and architecture document updates
- CHANGELOG entry creation
- Version control commit with SPEC reference

### Parallel Development with Git Worktree

Worktree Concept:

- Independent working directories for multiple branches
- Each SPEC gets isolated development environment
- No branch switching needed for parallel work
- Reduced merge conflicts through feature isolation

Worktree Creation:

- Command /moai:1-plan "login feature" "signup feature" --worktree creates multiple SPECs
- Result creates project-worktrees directory with SPEC-specific subdirectories

Worktree Benefits:

- Parallel Development: Multiple features developed simultaneously
- Team Collaboration: Clear ownership boundaries per SPEC
- Dependency Isolation: Different library versions per feature
- Risk Reduction: Unstable code does not affect other features

---

## Advanced Implementation (10+ minutes)

For advanced patterns including SPEC templates, validation automation, and workflow optimization, see:

- [Advanced Patterns](modules/advanced-patterns.md): Custom SPEC templates, validation automation
- [Reference Guide](reference.md): SPEC metadata schema, integration examples
- [Examples](examples.md): Real-world SPEC documents, workflow scenarios

## Resources

### SPEC File Organization

Directory Structure (Standard 3-File Format):

- .moai/specs/SPEC-{ID}/: SPEC document directory containing 3 required files
  - spec.md: EARS format specification (Environment, Assumptions, Requirements, Specifications)
  - plan.md: Implementation plan, milestones, technical approach
  - acceptance.md: Detailed acceptance criteria, test scenarios (Given-When-Then format)
- .moai/memory/: Session state files (last-session-state.json)
- .moai/docs/: Generated documentation (api-documentation.md)

[HARD] Required File Set:
Every SPEC directory MUST contain all 3 files (spec.md, plan.md, acceptance.md)
WHY: Complete SPEC structure ensures traceability, implementation guidance, and quality validation
IMPACT: Missing files create incomplete requirements and prevent proper workflow execution

### SPEC Metadata Schema

Required Fields:

- SPEC ID: Sequential number (SPEC-001, SPEC-002, etc.)
- Title: Feature name in English
- Created: ISO 8601 timestamp
- Status: Planned, In Progress, Completed, Blocked
- Priority: High, Medium, Low
- Assigned: Agent responsible for implementation

Optional Fields:

- Related SPECs: Dependencies and related features
- Epic: Parent feature group
- Estimated Effort: Time estimate in hours or story points
- Labels: Tags for categorization

### SPEC Lifecycle Management (SDD 2025 Standard)

Lifecycle Level Field:

Level 1 - spec-first:

- Description: SPEC written before implementation, discarded after completion
- Use Case: One-time features, prototypes, experiments
- Maintenance Policy: No maintenance required after implementation

Level 2 - spec-anchored:

- Description: SPEC maintained alongside implementation for evolution
- Use Case: Core features, API contracts, integration points
- Maintenance Policy: Quarterly review, update when implementation changes

Level 3 - spec-as-source:

- Description: SPEC is the single source of truth; only SPEC is edited by humans
- Use Case: Critical systems, regulated environments, code generation workflows
- Maintenance Policy: SPEC changes trigger implementation regeneration

Lifecycle Transition Rules:

- spec-first to spec-anchored: When feature becomes production-critical
- spec-anchored to spec-as-source: When compliance or regeneration workflow required
- Downgrade allowed but requires explicit justification in SPEC history

### Quality Metrics

SPEC Quality Indicators:

- Requirement Clarity: All EARS patterns used appropriately
- Test Coverage: All requirements have corresponding test scenarios
- Constraint Completeness: Technical and business constraints defined
- Success Criteria Measurability: Quantifiable completion metrics

Validation Checklist:

- All EARS requirements testable
- No ambiguous language (should, might, usually)
- All error cases documented
- Performance targets quantified
- Security requirements OWASP-compliant

### Works Well With

- moai-foundation-core: SPEC-First DDD methodology and TRUST 5 framework
- moai-workflow-testing: DDD implementation and test automation
- moai-workflow-project: Project initialization and configuration
- moai-workflow-worktree: Git Worktree management for parallel development
- manager-spec: SPEC creation and requirement analysis agent
- manager-ddd: DDD implementation based on SPEC requirements
- manager-quality: TRUST 5 quality validation and gate enforcement

### Integration Examples

Sequential Workflow:

- Step 1 PLAN: /moai:1-plan "user authentication system"
- Step 2 RUN: /moai:2-run SPEC-001
- Step 3 SYNC: /moai:3-sync SPEC-001

Parallel Workflow:

- Create multiple SPECs: /moai:1-plan "backend API" "frontend UI" "database schema" --worktree
- Session 1: /moai:2-run SPEC-001 (backend API)
- Session 2: /moai:2-run SPEC-002 (frontend UI)
- Session 3: /moai:2-run SPEC-003 (database schema)

### Token Management

Session Strategy:

- PLAN phase uses approximately 30% of session tokens
- RUN phase uses approximately 60% of session tokens
- SYNC phase uses approximately 10% of session tokens

Context Optimization:

- SPEC document persists in .moai/specs/ directory
- Session memory in .moai/memory/ for cross-session context
- Minimal context transfer through SPEC ID reference
- Agent delegation reduces token overhead

---

## SPEC Scope and Classification (NEW)

### What Belongs in .moai/specs/

The `.moai/specs/` directory is EXCLUSIVELY for SPEC documents that define features to be implemented.

Valid SPEC Content:

- Feature requirements in EARS format
- Implementation plans with milestones
- Acceptance criteria with Given/When/Then scenarios
- Technical specifications for new functionality
- User stories with clear deliverables

SPEC Characteristics:

- Forward-looking: Describes what WILL be built
- Actionable: Contains implementation guidance
- Testable: Includes acceptance criteria
- Structured: Uses EARS format patterns

### What Does NOT Belong in .moai/specs/

| Document Type         | Why Not SPEC                  | Correct Location                          |
| --------------------- | ----------------------------- | ----------------------------------------- |
| Security Audit        | Analyzes existing code        | `.moai/reports/security-audit-{DATE}/`    |
| Performance Report    | Documents current metrics     | `.moai/reports/performance-{DATE}/`       |
| Dependency Analysis   | Reviews existing dependencies | `.moai/reports/dependency-review-{DATE}/` |
| Architecture Overview | Documents current state       | `.moai/docs/architecture.md`              |
| API Reference         | Documents existing APIs       | `.moai/docs/api-reference.md`             |
| Meeting Notes         | Records decisions made        | `.moai/reports/meeting-{DATE}/`           |
| Retrospective         | Analyzes past work            | `.moai/reports/retro-{DATE}/`             |

### Exclusion Rules

[HARD] Report vs SPEC Distinction:

Reports analyze what EXISTS → `.moai/reports/`
SPECs define what will be BUILT → `.moai/specs/`

[HARD] Documentation vs SPEC Distinction:

Documentation explains HOW TO USE → `.moai/docs/`
SPECs define WHAT TO BUILD → `.moai/specs/`

---

## Migration Guide for Legacy Files

### Scenario 1: Flat SPEC File → Directory Conversion

Problem: `.moai/specs/SPEC-AUTH-001.md` exists as single file

Solution Steps:

1. Create directory: `mkdir -p .moai/specs/SPEC-AUTH-001/`
2. Move content: `mv .moai/specs/SPEC-AUTH-001.md .moai/specs/SPEC-AUTH-001/spec.md`
3. Create missing files:
   - Extract implementation plan → `plan.md`
   - Extract acceptance criteria → `acceptance.md`
4. Verify structure: All 3 files present
5. Commit: `git add . && git commit -m "refactor(spec): Convert SPEC-AUTH-001 to directory structure"`

Validation Command:

```bash
# Check for flat SPEC files (should return empty)
find .moai/specs -maxdepth 1 -name "SPEC-*.md" -type f
```

### Scenario 2: Unnumbered SPEC ID → Number Assignment

Problem: `SPEC-REDESIGN` or `SPEC-SDK-INTEGRATION` without number

Solution Steps:

1. Find next available number:
   ```bash
   ls -d .moai/specs/SPEC-*-[0-9][0-9][0-9] 2>/dev/null | sort -t- -k3 -n | tail -1
   ```
2. Assign number: `SPEC-REDESIGN` → `SPEC-REDESIGN-001`
3. Rename directory:
   ```bash
   mv .moai/specs/SPEC-REDESIGN .moai/specs/SPEC-REDESIGN-001
   ```
4. Update internal references in spec.md frontmatter
5. Commit: `git commit -m "refactor(spec): Assign number to SPEC-REDESIGN → SPEC-REDESIGN-001"`

### Scenario 3: Report in SPEC Directory → Separation

Problem: Analysis/audit document in `.moai/specs/`

Solution Steps:

1. Identify document type from content
2. Create reports directory:
   ```bash
   mkdir -p .moai/reports/security-audit-2025-01/
   ```
3. Move content:
   ```bash
   mv .moai/specs/SPEC-SECURITY-AUDIT/* .moai/reports/security-audit-2025-01/
   rmdir .moai/specs/SPEC-SECURITY-AUDIT
   ```
4. Rename main file to report.md if needed
5. Commit: `git commit -m "refactor: Move security audit from specs to reports"`

### Scenario 4: Duplicate SPEC ID → Resolution

Problem: Two directories with same SPEC ID

Solution Steps:

1. Compare creation dates:
   ```bash
   ls -la .moai/specs/ | grep SPEC-AUTH-001
   ```
2. Determine which is canonical (usually older one)
3. Renumber newer one to next available:
   ```bash
   mv .moai/specs/SPEC-AUTH-001-duplicate .moai/specs/SPEC-AUTH-002
   ```
4. Update internal references
5. Commit: `git commit -m "fix(spec): Resolve duplicate SPEC-AUTH-001 → SPEC-AUTH-002"`

### Validation Script

Run this script to identify SPEC organization issues:

```bash
#!/bin/bash
# SPEC Organization Validator

echo "=== SPEC Organization Check ==="

# Check 1: Flat files in specs root
echo -e "\n[Check 1] Flat SPEC files (should be empty):"
find .moai/specs -maxdepth 1 -name "SPEC-*.md" -type f

# Check 2: Directories without required files
echo -e "\n[Check 2] SPEC directories missing required files:"
for dir in .moai/specs/SPEC-*/; do
  if [ -d "$dir" ]; then
    missing=""
    [ ! -f "${dir}spec.md" ] && missing="${missing}spec.md "
    [ ! -f "${dir}plan.md" ] && missing="${missing}plan.md "
    [ ! -f "${dir}acceptance.md" ] && missing="${missing}acceptance.md "
    [ -n "$missing" ] && echo "$dir: Missing $missing"
  fi
done

# Check 3: SPECs without numbers
echo -e "\n[Check 3] SPECs without proper numbering:"
ls -d .moai/specs/SPEC-*/ 2>/dev/null | grep -v -E 'SPEC-[A-Z]+-[0-9]{3}'

# Check 4: Potential reports in specs
echo -e "\n[Check 4] Potential reports in specs (check manually):"
grep -l -r "findings\|recommendations\|audit\|analysis" .moai/specs/*/spec.md 2>/dev/null

echo -e "\n=== Check Complete ==="
```

---

Version: 1.3.0 (SDD 2025 Standard Integration + SPEC Scope Classification)
Last Updated: 2026-01-21
Integration Status: Complete - Full Plan-Run-Sync workflow with SDD 2025 features and Migration Guide
