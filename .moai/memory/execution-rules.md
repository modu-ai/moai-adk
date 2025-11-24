# Execution Rules and Constraints

## Overview

This document defines the execution rules, constraints, and security policies that govern MoAI-ADK agent behavior and tool usage.

## Core Execution Principles

### Agent-First Mandate

**Rule**: Never execute tasks directly. Always delegate to specialized agents.

**Implementation**:

```python
# FORBIDDEN: Direct execution
def process_user_data():
    # This code should NOT exist
    pass

# REQUIRED: Agent delegation
result = await Task(
    subagent_type="security-expert",
    prompt="Process user data with security validation"
)
```

### Tool Usage Restrictions

**Allowed Tools Only**:

- `Task()`: Agent delegation (primary tool)
- `AskUserQuestion()`: User interaction and clarification
- `Skill()`: Knowledge invocation and expertise
- `MCP Servers`: Context7, Playwright, Figma integration

**Forbidden Tools**:

- `Read()`, `Write()`, `Edit()`: Use `Task()` for file operations
- `Bash()`: Use `Task()` for system operations
- `Grep()`, `Glob()`: Use `Task()` for file search
- `TodoWrite()`: Use `Task()` for tracking

### Delegation Protocol

**Standard Delegation Pattern**:

```python
result = await Task(
    subagent_type="specialized_agent",
    prompt="Clear, specific task description",
    context={
        "relevant_data": "necessary_information",
        "constraints": "limitations_and_requirements"
    }
)
```

## Security Constraints

### Sandbox Mode (Always Active)

**Enabled Features**:

- Tool usage validation and logging
- File access restrictions
- Command execution limits
- Permission enforcement

**File Access Restrictions**:

```
PROTECTED PATHS (DENIED ACCESS):
- .env*
- .vercel/
- .netlify/
- .firebase/
- .aws/
- .github/workflows/secrets
- .kube/config
- ~/.ssh/
```

**Command Restrictions**:

```
FORBIDDEN COMMANDS:
- rm -rf
- sudo
- chmod 777
- dd
- mkfs
- shutdown
- reboot
```

### Data Protection Rules

**Sensitive Data Handling**:

- Never log or expose passwords, API keys, or tokens
- Use environment variables for sensitive configuration
- Validate all user inputs before processing
- Encrypt sensitive data at rest

**Input Validation**:

```python
# Required validation before processing
def validate_input(user_input, context=None):
    # Always validate inputs before delegation
    if not is_safe_input(user_input):
        raise SecurityValidationError("Unsafe input detected")
    return sanitize_input(user_input)
```

## Permission System

### Role-Based Access Control

**Agent Permissions**:

- **Read Agents**: File system exploration, code analysis
- **Write Agents**: File creation and modification (with validation)
- **System Agents**: Limited system operations (validated commands)
- **Security Agents**: Security analysis and validation

**Permission Levels**:

1. **Level 1** (Read-only): Code exploration, analysis
2. **Level 2** (Validated Write): File creation with validation
3. **Level 3** (System): Limited system operations
4. **Level 4** (Security): Security analysis and enforcement

### MCP Server Permissions

**Context7**:

- Library documentation access
- API reference resolution
- Version compatibility checking

**Playwright**:

- Browser automation for testing
- Screenshot capture
- UI interaction simulation
- E2E test execution

**Figma**:

- Design system access
- Component extraction
- Design-to-code conversion
- Style guide generation

## Quality Gates

### TRUST 5 Framework

**Test-First**:

- Every implementation must start with tests
- Test coverage must exceed 85%
- Tests must validate all requirements
- Failed tests must block deployment

**Readable**:

- Code must follow established style guidelines
- Variable and function names must be descriptive
- Complex logic must include comments
- Code structure must be maintainable

**Unified**:

- Consistent patterns across codebase
- Standardized naming conventions
- Uniform error handling
- Consistent documentation format

**Secured**:

- Security validation through security-expert
- OWASP compliance checking
- Input sanitization and validation
- Secure coding practices

**Trackable**:

- All changes must have clear origin
- Implementation must link to specifications
- Test coverage must be verifiable
- Quality metrics must be tracked

### Automated Validation

**Pre-execution Validation**:

```python
# Required before any task execution
def validate_execution_requirements(task, context):
    validations = [
        validate_security_clearance(task),
        validate_resource_availability(context),
        validate_quality_standards(task),
        validate_permission_compliance(task)
    ]

    return all(validations)
```

**Post-execution Validation**:

```python
# Required after task completion
def validate_execution_results(result, task):
    validations = [
        validate_output_quality(result),
        validate_security_compliance(result),
        validate_test_coverage(result),
        validate_documentation_completeness(result)
    ]

    return all(validations)
```

## Error Handling Protocols

### Error Classification

**Critical Errors** (Immediate Stop):

- Security violations
- Data corruption risks
- System integrity threats
- Permission violations

**Warning Errors** (Continue with Monitoring):

- Performance degradation
- Resource limitations
- Quality gate failures
- Documentation gaps

**Informational Errors** (Log and Continue):

- Non-critical warnings
- Minor quality issues
- Optimization opportunities
- Style guide deviations

### Recovery Procedures

**Critical Error Recovery**:

1. Immediately stop execution
2. Log error details securely
3. Notify system administrator
4. Rollback changes if possible
5. Analyze root cause

**Warning Error Recovery**:

1. Log warning with context
2. Continue execution with monitoring
3. Document issue for later resolution
4. Notify user of potential impact

**Informational Error Recovery**:

1. Log information
2. Continue execution normally
3. Add to improvement backlog
4. Monitor for patterns

## Resource Management

### Token Budget Management

**Phase-based Allocation**:

- **Planning Phase**: 30K tokens maximum
- **Implementation Phase**: 180K tokens maximum
- **Documentation Phase**: 40K tokens maximum
- **Total Budget**: 250K tokens per feature

**Optimization Requirements**:

- Execute `/clear` immediately after SPEC creation
- Monitor token usage continuously
- Use efficient context loading
- Cache reusable results

### File System Management

**Allowed Operations**:

- Read files within project directory
- Write files to designated locations
- Create documentation in `.moai/` directory
- Generate code in source directories

**Prohibited Operations**:

- Modify system files outside project
- Access sensitive configuration files
- Delete critical system resources
- Execute unauthorized system commands

### Memory Management

**Context Optimization**:

- Load only necessary files for current task
- Use efficient data structures
- Clear context between major phases
- Cache frequently accessed information

**Resource Limits**:

- Maximum file size: 10MB per file
- Maximum concurrent operations: 5
- Maximum context size: 150K tokens
- Maximum execution time: 5 minutes per task

## Git Strategy Guide

### Configuration-Based Git Behavior (GitHub Flow 3-Mode System)

Alfred automatically adjusts Git workflow based on git configuration in `.moai/config/config.json`.

**Key Configuration Fields**:

- `git_strategy.mode`: Git mode selection (manual, personal, team)
- `git_strategy.branch_creation.prompt_always`: Whether to prompt user for every SPEC (true/false)
- `git_strategy.branch_creation.auto_enabled`: Enable automatic branch creation in Personal/Team modes (true/false)

**3-Mode System Overview (Two-Level Control)**:

| Configuration                               | Manual Mode           | Personal/Team Mode     | Effect                                              |
| ------------------------------------------- | --------------------- | ---------------------- | --------------------------------------------------- |
| **prompt_always=true, auto_enabled=false**  | Prompt user each time | Prompt user each time  | Maximum control (default)                           |
| **prompt_always=false, auto_enabled=false** | Auto skip             | Wait for user approval | Manual=auto skip, Personal/Team=auto after approval |
| **prompt_always=false, auto_enabled=true**  | Auto skip             | Auto create            | Full automation                                     |

### Common Branch Creation Setting: Two-Level Control System

**Level 1: `branch_creation.prompt_always`**

**Default (true)**:

- User prompted on every SPEC generation: "Create branch?"
- User choices: "Auto create" or "Use current branch"
- Consistent UX across all modes

**Custom (false)**:

- **Manual Mode**: Automatically skip branch creation (work on current branch)
- **Personal/Team Mode**: Behavior determined by `auto_enabled` value

**Level 2: `branch_creation.auto_enabled` (Valid only in Personal/Team Mode)**

**Default (false)**:

- Only active in Personal/Team mode when `prompt_always=false`
- User prompted for approval: "Enable automatic branch creation?"
- On approval: config.json automatically updated with `auto_enabled=true`
- From next SPEC: Branches created automatically without user prompt

**Custom (true)**:

- Automatically create branches for every SPEC in Personal/Team mode
- No user prompts (full automation)

### Mode 1: Manual (Local Git Only)

**Configuration (default)**:

```json
{
  "git_strategy": {
    "mode": "manual",
    "branch_creation": {
      "prompt_always": true,
      "auto_enabled": false
    }
  }
}
```

**Alfred's Behavior** (prompt_always=true):

1. When running `/moai:1-plan`, user prompted: "Create branch?"
   - Auto create option → Creates feature/SPEC-001
   - Use current branch option → Continues on current branch
2. All TDD commits saved locally only (automatic)
3. Push performed manually

**Configuration (auto skip)**:

```json
{
  "git_strategy": {
    "mode": "manual",
    "branch_creation": {
      "prompt_always": false,
      "auto_enabled": false
    }
  }
}
```

**Alfred's Behavior** (prompt_always=false):

- All SPECs automatically work on current branch (no branch creation)
- No user prompts

**Use Case**: Personal projects, GitHub not used, local Git only

### Mode 2: Personal (GitHub Personal Project)

**Configuration (default - prompt each time)**:

```json
{
  "git_strategy": {
    "mode": "personal",
    "branch_creation": {
      "prompt_always": true,
      "auto_enabled": false
    }
  }
}
```

**Alfred's Behavior** (prompt_always=true):

1. When running `/moai:1-plan`, user prompted: "Create branch?"
   - Auto create option → Creates feature/SPEC-002 + auto push
   - Use current branch option → Commits directly on current branch
2. Running `/moai:2-run`: TDD commits + auto push
3. Running `/moai:3-sync`: Doc commits + suggest PR creation (user choice)

**Configuration (auto after approval)**:

```json
{
  "git_strategy": {
    "mode": "personal",
    "branch_creation": {
      "prompt_always": false,
      "auto_enabled": false
    }
  }
}
```

**Alfred's Behavior** (prompt_always=false, auto_enabled=false):

1. When running `/moai:1-plan`, user prompted once: "Enable automatic branch creation?"
   - **Yes** option → Auto updates config.json with `auto_enabled=true` → Creates feature/SPEC
   - **No** option → Works on current branch, no config change
2. From next SPEC: If `auto_enabled=true`, feature branches created automatically without prompts

**Configuration (full automation)**:

```json
{
  "git_strategy": {
    "mode": "personal",
    "branch_creation": {
      "prompt_always": false,
      "auto_enabled": true
    }
  }
}
```

**Alfred's Behavior** (prompt_always=false, auto_enabled=true):

- Automatically creates feature/SPEC-XXX branch for every SPEC
- No user prompts (full automation)
- All TDD and documentation commits auto-pushed to feature branch

**Use Case**: Personal GitHub projects, fast development speed needed

### Mode 3: Team (GitHub Team Project)

**Configuration (default - prompt each time)**:

```json
{
  "git_strategy": {
    "mode": "team",
    "branch_creation": {
      "prompt_always": true,
      "auto_enabled": false
    }
  }
}
```

**Alfred's Behavior** (prompt_always=true):

1. When running `/moai:1-plan`, user prompted: "Create branch?"
   - Auto create option → Creates feature/SPEC-003 + auto create Draft PR
   - Use current branch option → Proceeds on current branch (not recommended)
2. Running `/moai:2-run`: TDD commits + auto push (to feature branch)
3. Running `/moai:3-sync`: Doc commits + prepare PR
4. Team code review required (minimum 1 reviewer)
5. After approval: Merge (Squash or Merge)

**Configuration (auto after approval)**:

```json
{
  "git_strategy": {
    "mode": "team",
    "branch_creation": {
      "prompt_always": false,
      "auto_enabled": false
    }
  }
}
```

**Alfred's Behavior** (prompt_always=false, auto_enabled=false):

1. When running `/moai:1-plan`, user prompted once: "Enable automatic branch creation and Draft PR creation?"
   - **Yes** option → Auto updates config.json with `auto_enabled=true` → Creates feature/SPEC + Draft PR
   - **No** option → Works on current branch, no config change
2. From next SPEC: If `auto_enabled=true`, feature branches + Draft PRs created automatically without prompts

**Configuration (full automation)**:

```json
{
  "git_strategy": {
    "mode": "team",
    "branch_creation": {
      "prompt_always": false,
      "auto_enabled": true
    }
  }
}
```

**Alfred's Behavior** (prompt_always=false, auto_enabled=true):

- Automatically creates feature/SPEC-XXX branch + Draft PR for every SPEC
- No user prompts (full automation)
- All TDD and documentation commits auto-pushed to feature branch
- Maintains Draft PR status (until team review complete)

**Use Case**: Team projects, code review required, quality management needed

### Configuration Flexibility

**When `prompt_always: false` is configured**:

```json
{
  "git_strategy": {
    "mode": "manual",
    "branch_creation": { "prompt_always": false }
  }
}
```

- **Manual Mode**: Automatically skip branch creation (no user prompt)
- **Personal/Team Mode**: Automatically create feature branches (no user choice)

### Mandatory `/clear` Execution After SPEC Generation

Execute `/clear` after `/moai:1-plan` completion in all modes.

---

## Compliance Requirements

### Legal and Regulatory

**Data Privacy**:

- GDPR compliance for user data
- CCPA compliance for California users
- Data minimization principles
- Right to deletion implementation

**Security Standards**:

- OWASP Top 10 compliance
- SOC 2 Type II controls
- ISO 27001 security management
- NIST Cybersecurity Framework

### Industry Standards

**Development Standards**:

- ISO/IEC 27001 security management
- ISO/IEC 9126 software quality
- IEEE 730 software engineering standards
- Agile methodology compliance

**Documentation Standards**:

- IEEE 1016 documentation standards
- OpenAPI specification compliance
- Markdown formatting consistency
- Accessibility documentation (WCAG 2.1)

## Monitoring and Auditing

### Activity Logging

**Required Log Entries**:

```python
{
    "timestamp": "2025-11-20T07:30:00Z",
    "agent": "security-expert",
    "action": "code_review",
    "files_accessed": ["src/auth.py", "tests/test_auth.py"],
    "token_usage": 5230,
    "duration_seconds": 12.5,
    "success": true,
    "quality_score": 0.95
}
```

**Audit Trail Requirements**:

- All agent delegations must be logged
- File access patterns must be tracked
- Security events must be recorded
- Quality metrics must be captured

### Performance Monitoring

**Key Metrics**:

- Agent delegation success rate
- Average response time per task
- Token usage efficiency
- Quality gate pass rate
- Error recovery time

**Alert Thresholds**:

- Success rate < 95%
- Response time > 30 seconds
- Token usage > 90% of budget
- Quality gate failure rate > 5%

## Enforcement Mechanisms

### Pre-execution Hooks

**Required Validations**:

```python
def pre_execution_hook(agent_type, prompt, context):
    validations = [
        validate_agent_permissions(agent_type),
        validate_prompt_safety(prompt),
        validate_context_integrity(context),
        validate_resource_availability()
    ]

    for validation in validations:
        if not validation.passed:
            raise ValidationError(validation.message)

    return True
```

### Post-execution Hooks

**Required Checks**:

```python
def post_execution_hook(result, agent_type, task):
    validations = [
        validate_output_quality(result),
        validate_security_compliance(result),
        validate_documentation_completeness(result),
        validate_test_adequacy(result)
    ]

    issues = [v for v in violations if not v.passed]
    if issues:
        raise QualityGateError("Quality gate failures detected")

    return True
```

### Automated Enforcement

**Real-time Monitoring**:

- Continuous validation of execution rules
- Automatic blocking of suspicious activities
- Real-time alert generation for violations
- Automated rollback of unsafe operations

**Periodic Audits**:

- Daily compliance checks
- Weekly performance reviews
- Monthly security assessments
- Quarterly quality audits

## Exception Handling

### Security Exceptions

**Emergency Override**:

- Only available to authorized administrators
- Requires explicit approval and logging
- Temporary override with strict time limits
- Full audit trail required

**Justified Exceptions**:

- Documented business requirements
- Risk assessment and mitigation
- Alternative security controls
- Regular review and renewal

### Performance Exceptions

**Resource Optimization**:

- Dynamic resource allocation
- Load balancing across agents
- Priority queue management
- Performance tuning procedures

**Emergency Procedures**:

- System overload protection
- Graceful degradation strategies
- User notification systems
- Recovery automation

This comprehensive set of execution rules ensures that MoAI-ADK operates securely, efficiently, and in compliance with industry standards while maintaining high quality output.
