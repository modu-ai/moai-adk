# MoAI-ADK Commands Reference

## Overview

This document defines the 6 core MoAI-ADK slash commands for SPEC-First TDD execution. Each command delegates to specialized agents and follows strict execution patterns.

## Command Architecture

### Execution Flow
1. User input → Command parsing → Agent delegation → Task execution
2. Context optimization between commands using `/clear`
3. Automatic quality gate enforcement
4. Integration with MCP servers for enhanced capabilities

### Core Principles
1. **SPEC-First**: All features start with specification
2. **TDD Mandatory**: Test-driven development cycle enforcement
3. **Agent Delegation**: Never execute directly, always delegate
4. **Context Optimization**: Use `/clear` for token efficiency

## Core Commands

### `/moai:0-project` - Project Initialization

**Purpose**: Initialize new project structure and detection

**Delegation**: `project-manager`, `plan`, `explore`

**Execution Process**:
1. Analyze current directory structure
2. Detect existing configuration files
3. Initialize `.moai/` directory structure
4. Create basic configuration files
5. Set up Git repository if needed
6. Generate project metadata

**Usage**:
```
/moai:0-project
/moai:0-project --with-git    # Initialize with Git
/moai:0-project --force      # Force reinitialization
```

**Outputs**:
- `.moai/` directory structure
- `.moai/config/config.json` with project metadata
- Git repository (if requested)
- Project analysis report

**Post-Execution**: Ready for SPEC creation

### `/moai:1-plan` - Specification Generation

**Purpose**: Create EARS-format specifications from user requirements

**Delegation**: `spec-builder`

**Execution Process**:
1. Parse user requirement description
2. Apply EARS format structure
3. Generate comprehensive specification
4. Create specification document structure
5. Validate specification completeness
6. Save to `.moai/specs/SPEC-XXX/spec.md`

**Usage**:
```
/moai:1-plan "User authentication system with JWT tokens"
/moai:1-plan "E-commerce product catalog with search"
```

**Outputs**:
- SPEC document in EARS format
- Specification file: `.moai/specs/SPEC-XXX/spec.md`
- Test cases outline
- Implementation requirements

**Critical Post-Execution**:
- MUST execute `/clear` immediately after completion
- Saves 45-50K tokens
- Prepares clean context for implementation phase

### `/moai:2-run` - TDD Implementation

**Purpose**: Execute Red-Green-Refactor TDD cycle for specifications

**Delegation**: `tdd-implementer`

**Execution Process**:
1. Load specification document
2. Phase 1: RED - Write failing tests
3. Phase 2: GREEN - Implement minimum passing code
4. Phase 3: REFACTOR - Optimize for quality
5. Quality gate validation
6. Generate implementation report

**Usage**:
```
/moai:2-run SPEC-001
/moai:2-run SPEC-001 --verbose    # Detailed output
```

**TDD Phases**:

#### RED Phase (Test Writing)
- Write comprehensive failing tests
- Cover all specification requirements
- Include edge cases and error conditions
- Validate test structure and coverage

#### GREEN Phase (Implementation)
- Implement minimum code to pass tests
- Focus on functionality over optimization
- Ensure all tests pass
- Document implementation decisions

#### REFACTOR Phase (Quality)
- Optimize code for readability and performance
- Apply design patterns and best practices
- Ensure test coverage meets standards (85%+)
- Validate code quality metrics

**Outputs**:
- Implemented code with tests
- Test coverage report
- Quality metrics analysis
- Implementation documentation

**Post-Execution**:
- Execute `/clear` if context > 150K tokens
- Ready for documentation synchronization

### `/moai:3-sync` - Documentation Synchronization

**Purpose**: Auto-generate documentation and create project artifacts

**Delegation**: `docs-manager`

**Execution Process**:
1. Analyze implementation and test results
2. Generate API documentation
3. Create architecture diagrams
4. Sync with external systems (Notion, GitHub)
5. Validate documentation completeness
6. Create project completion reports

**Usage**:
```
/moai:3-sync SPEC-001
/moai:3-sync SPEC-001 --force     # Force full sync
/moai:3-sync SPEC-001 status      # Check sync status
```

**Documentation Types**:
- API documentation (OpenAPI format)
- Architecture diagrams
- User guides and tutorials
- Developer documentation
- Project completion reports

**Outputs**:
- Generated documentation in `.moai/docs/`
- Project reports in `.moai/reports/`
- Quality gate validation results
- Synchronization status

**Post-Execution**:
- Execute `/clear` for clean next phase
- Ready for feature completion or iteration

### `/moai:9-feedback` - Feedback Analysis

**Purpose**: Analyze batch feedback and provide improvement recommendations

**Delegation**: `quality-gate`

**Execution Process**:
1. Collect feedback data from various sources
2. Analyze patterns and trends
3. Identify improvement opportunities
4. Generate actionable recommendations
5. Create improvement action plans
6. Track implementation progress

**Usage**:
```
/moai:9-feedback
/moai:9-feedback --data feedback.json    # With specific data
/moai:9-feedback --analyze SPEC-001       # Analyze specific spec
```

**Feedback Sources**:
- Code review comments
- Test results and coverage
- Performance metrics
- User acceptance testing
- Quality gate results

**Outputs**:
- Feedback analysis report
- Improvement recommendations
- Action item prioritization
- Progress tracking dashboard

### `/moai:99-release` - Production Release

**Purpose**: Create production-ready release with validation

**Delegation**: `release-manager`

**Execution Process**:
1. Validate all quality gates
2. Generate release artifacts
3. Create release notes
4. Tag and version repository
5. Prepare deployment packages
6. Generate compliance reports

**Usage**:
```
/moai:99-release
/moai:99-release --patch    # Patch release
/moai:99-release --minor    # Minor release
/moai:99-release --major    # Major release
```

**Release Validation**:
- All quality gates passed
- Test coverage > 85%
- Security scan passed
- Documentation complete
- Performance benchmarks met

**Outputs**:
- Release package
- Version tags
- Release notes
- Deployment instructions
- Compliance reports

## Command Integration Patterns

### Complete Feature Development
```
# 1. Initialize project
/moai:0-project

# 2. Create specification
/moai:1-plan "User authentication system"
/clear    # Critical: Reset context

# 3. Implement with TDD
/moai:2-run SPEC-001
/clear    # Optional: If context > 150K

# 4. Sync documentation
/moai:3-sync SPEC-001

# 5. Release
/moai:99-release
```

### Iterative Development
```
# Base feature
/moai:1-plan "Core authentication"
/clear
/moai:2-run SPEC-001

# Enhancement
/moai:1-plan "Add social login"
/clear
/moai:2-run SPEC-002

# Integration
/moai:3-sync SPEC-001 SPEC-002
```

### Bug Fix Workflow
```
# Analyze issue
/moai:9-feedback --analyze bug-report.json

# Create fix specification
/moai:1-plan "Fix authentication token expiry"
/clear

# Implement fix
/moai:2-run SPEC-003

# Validate and release
/moai:3-sync SPEC-003
/moai:99-release --patch
```

## Context Management Rules

### When to Use `/clear`

**Mandatory**:
- Immediately after `/moai:1-plan` completion
- Saves 45-50K tokens
- Prevents context overflow

**Recommended**:
- After `/moai:2-run` if context > 150K tokens
- Before starting new complex feature
- Every 50+ messages in conversation

**Process**:
1. Check token usage: `/context`
2. If > 150K: Execute `/clear`
3. Continue with next command

### Context Budgeting

**Phase-based Token Allocation**:
- SPEC Creation: 30K tokens maximum
- TDD Implementation: 180K tokens maximum
- Documentation: 40K tokens maximum
- Total: 250K tokens per feature

**Optimization Strategy**:
- Load only relevant files for each phase
- Use `Task()` context passing between phases
- Clear context between major phases

## Error Handling

### Command Execution Failures

**Common Issues**:
- Agent not found or unavailable
- File permission errors
- Token limit exceeded
- Network connectivity issues

**Recovery Process**:
1. Use `debug-helper` to analyze error
2. Check system status with relevant agents
3. Restart failed command with corrected context
4. Validate success with quality gates

### Context Overflow Recovery

**Symptoms**:
- Slow response times
- Token limit errors
- Context truncation

**Recovery Steps**:
1. Execute `/clear` immediately
2. Restart from last successful checkpoint
3. Use minimal context for retry
4. Monitor token usage closely

## Quality Gates

### Automatic Validation

All commands enforce TRUST 5 principles:
- **Test-first**: Every implementation starts with tests
- **Readable**: Code clarity and documentation
- **Unified**: Consistent patterns and conventions
- **Secured**: Security validation and best practices
- **Trackable**: Change history and provenance

### Quality Metrics

**Code Quality**:
- Test coverage > 85%
- Code review passed
- Security scan passed
- Performance benchmarks met

**Documentation Quality**:
- API documentation complete
- User guides validated
- Developer guides comprehensive
- Architecture diagrams accurate

**Process Quality**:
- SPEC-First compliance
- TDD cycle completion
- Agent delegation followed
- Context optimization applied

## Integration with MCP Servers

### Context7 Integration
- Resolve library documentation
- Get latest API references
- Validate implementation patterns

### Playwright Integration
- E2E test automation
- Browser-based validation
- UI component testing

### Figma Integration
- Design system validation
- Component library sync
- Design-to-code verification

## Command Customization

### Environment Variables
- `MOAI_DEBUG`: Enable detailed logging
- `MOAI_DRY_RUN`: Simulate execution without changes
- `MOAI_VERBOSE`: Show detailed execution output

### Configuration Files
- `.moai/config/config.json`: Project configuration
- `.claude/settings.json`: Claude Code settings
- `.mcp.json`: MCP server configuration

### Hooks Integration
- Pre-command validation
- Post-command cleanup
- Error handling and recovery
- Progress tracking and reporting