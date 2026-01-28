# Workflow Modes

Three-phase SPEC-First DDD workflow for quality-driven development.

## Plan Phase

Command: `/moai plan`
Agent: manager-spec
Purpose: Create comprehensive specification document

### Token Budget
- Allocation: 30,000 tokens
- Strategy: Minimal context
- Post-action: Execute /clear
- Saving: 45-50K tokens for implementation

### Input Requirements
- User request or feature description
- Project context and goals
- Stakeholder requirements
- Constraints and dependencies

### Output
- SPEC document at `.moai/specs/SPEC-XXX/spec.md`
- EARS format requirements
- Acceptance criteria
- Technical approach

### Success Criteria
- All requirements documented in EARS format
- Acceptance criteria clearly defined
- Technical approach feasible
- Dependencies identified
- Stakeholder approval obtained

## Run Phase

Command: `/moai run SPEC-XXX`
Agent: manager-ddd
Purpose: Implement specification with behavior preservation

### Token Budget
- Allocation: 180,000 tokens
- Strategy: Selective file loading
- Benefit: 70% larger implementations

### DDD Cycle

**ANALYZE**: Understand existing behavior and code structure
- Read existing code
- Identify dependencies
- Map domain boundaries

**PRESERVE**: Create characterization tests for existing behavior
- Write characterization tests
- Capture current behavior
- Verify test coverage

**IMPROVE**: Implement changes with behavior preservation
- Make small, incremental changes
- Run characterization tests after each change
- Refactor with test validation

### Success Criteria
- All SPEC requirements implemented
- Characterization tests passing
- 85%+ code coverage achieved
- TRUST 5 quality gates passed
- No behavior regressions

## Sync Phase

Command: `/moai sync SPEC-XXX`
Agent: manager-docs
Purpose: Generate documentation and prepare for deployment

### Token Budget
- Allocation: 40,000 tokens
- Strategy: Result caching
- Reduction: 60% fewer redundant file reads

### Output
- API documentation
- Updated README
- CHANGELOG entry
- Pull request with documentation

### Success Criteria
- API documentation complete
- README updated with usage examples
- CHANGELOG entry added
- Pull request created
- All links verified

## Phase Transitions

### Plan to Run
- Trigger: SPEC document approved
- Action: Execute /clear, then /moai run SPEC-XXX
- Handoff: SPEC document path, requirements summary

### Run to Sync
- Trigger: Implementation complete, tests passing
- Action: Execute /moai sync SPEC-XXX
- Handoff: SPEC reference, implementation summary, test results
