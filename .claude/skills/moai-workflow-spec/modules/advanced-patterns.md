# Advanced SPEC Patterns

## Custom SPEC Templates

This module provides advanced patterns for SPEC workflow management, including template customization, validation automation, and workflow optimization strategies.

---

## SPEC Template Customization

### Template Inheritance Pattern

Create base templates that can be extended for specific use cases:

Base Template Structure:
- Header section with standard metadata fields
- Requirements section with placeholder EARS patterns
- Constraints section with technical and business categories
- Success criteria section with functional, performance, and security categories
- Test scenarios section with category organization

Extension Process:
- Copy base template to new SPEC file
- Fill in specific requirements for your feature
- Add domain-specific constraints
- Define measurable success criteria
- Create comprehensive test scenarios

### Domain-Specific Templates

Backend API Template:
- Includes API specification section
- Response schema definitions
- Rate limiting and authentication requirements
- Performance targets for endpoint response times
- Database schema considerations

Frontend Component Template:
- Includes component API section
- Props interface definition
- Accessibility requirements (WCAG compliance)
- Styling constraints and design tokens
- Storybook documentation requirements

Workflow Template:
- Includes preconditions and side effects sections
- State machine definitions
- Rollback strategy documentation
- Compensating transaction patterns
- Multi-step orchestration requirements

---

## Validation Automation

### Automated Quality Checks

EARS Pattern Validation:
- Verify each requirement uses exactly one EARS pattern
- Check for pattern-appropriate keywords (WHEN/THEN, IF/THEN, etc.)
- Detect mixed patterns within single requirements
- Flag ambiguous language that needs clarification

Test Coverage Validation:
- Map each requirement to corresponding test scenarios
- Identify requirements without test coverage
- Detect orphan test scenarios without linked requirements
- Calculate coverage percentage for SPEC completeness

Success Criteria Validation:
- Verify all criteria are quantifiable (contain numeric targets)
- Check for measurable metrics (response time, coverage percentage, etc.)
- Flag qualitative criteria that cannot be objectively verified
- Ensure all criteria have verification methods defined

### Continuous Validation Integration

Pre-Commit Validation:
- Run SPEC linter before allowing commits
- Block commits with validation errors
- Generate validation reports for review
- Suggest corrections for common issues

CI/CD Integration:
- Validate SPECs in pull request checks
- Generate coverage reports for SPEC requirements
- Track SPEC quality metrics over time
- Alert on quality regression

---

## Workflow Optimization Strategies

### Token Budget Optimization

PLAN Phase Optimization:
- Focus on requirement extraction efficiency
- Use structured prompts for faster clarification
- Minimize back-and-forth dialogue through comprehensive questions
- Save SPEC document before context limit reached

RUN Phase Optimization:
- Load only relevant SPEC sections for current task
- Use targeted test execution for faster feedback
- Implement incremental DDD for large features
- Clear context between major implementation phases

SYNC Phase Optimization:
- Generate documentation from structured SPEC data
- Use templates for consistent documentation format
- Batch multiple SPEC updates in single sync session
- Automate changelog generation from SPEC metadata

### Parallel Development Patterns

Feature Independence Analysis:
- Identify truly independent features for parallel work
- Map shared dependencies that require coordination
- Define integration points and contracts upfront
- Establish merge order based on dependency graph

Worktree Coordination:
- Create separate worktrees for independent SPECs
- Use feature flags for parallel integration
- Define clear ownership boundaries per worktree
- Schedule regular sync points for dependency updates

Conflict Prevention:
- Define interface contracts before parallel implementation
- Use lock files for shared configuration changes
- Implement automated conflict detection in CI
- Document merge strategies for common conflict patterns

---

## Advanced Quality Patterns

### SPEC Review Process

Review Checklist:
- All EARS patterns correctly applied
- No ambiguous language present
- All error cases documented
- Performance targets quantified
- Security requirements OWASP-compliant
- Test scenarios cover all requirements
- Success criteria measurable and verifiable

Review Workflow:
- Author creates SPEC draft
- Reviewer validates against checklist
- Author adddesses review feedback
- Final approval before implementation
- Post-implementation verification

### Quality Metrics Tracking

SPEC Quality Indicators:
- Requirement clarity score based on language analysis
- Test coverage percentage for requirements
- Constraint completeness assessment
- Success criteria measurability rating

Trend Analysis:
- Track quality metrics over time
- Identify common quality issues
- Measure improvement from process changes
- Generate team quality reports

---

## Integration Patterns

### External System Integration

Third-Party API Integration:
- Define contract requirements in SPEC
- Document expected response schemas
- Include error handling for external failures
- Specify timeout and retry requirements

Database Schema Integration:
- Document schema requirements in constraints
- Define migration strategy for schema changes
- Include rollback procedures for failed migrations
- Specify data validation requirements

Message Queue Integration:
- Define message schemas in SPEC
- Document delivery guarantees (at-least-once, exactly-once)
- Include dead letter queue handling
- Specify retry and timeout policies

### Cross-SPEC Dependencies

Dependency Documentation:
- Use Related SPECs field for explicit dependencies
- Document blocking and blocked-by relationships
- Define integration contracts between SPECs
- Specify integration testing requirements

Dependency Management:
- Create dependency graph for complex features
- Identify critical path for prioritization
- Track dependency completion status
- Alert on blocked SPEC resolution

---

## Troubleshooting Advanced Issues

### Complex Requirement Decomposition

Problem: Requirement too complex for single EARS pattern
Solution: Break into multiple atomic requirements, each with single pattern

Problem: Requirements have implicit dependencies
Solution: Make dependencies explicit with precondition documentation

Problem: Requirements conflict with existing system behavior
Solution: Document conflict and propose resolution strategy

### Performance Optimization Issues

Problem: SPEC processing takes too long
Solution: Use progressive disclosure, load sections on demand

Problem: Too many SPECs to manage effectively
Solution: Group SPECs into Epics, use automated tracking

Problem: SPEC drift from implementation
Solution: Regular sync validation, automated drift detection

---

Version: 1.0.0
Last Updated: 2025-12-07
