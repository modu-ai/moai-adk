# Execution Rules and Constraints

Purpose: Security policies, execution constraints, and Git workflow strategies governing MoAI-ADK agent behavior.

Last Updated: 2025-11-25
Version: 2.0.0

---

## Quick Reference (30 seconds)

MoAI operates under strict execution rules:

Core Constraints:
- Agent-First: ALWAYS delegate via Task(), NEVER execute directly
- Allowed Tools: Task(), AskUserQuestion(), Skill(), MCP Servers
- Forbidden Tools: Read(), Write(), Edit(), Bash(), Grep(), Glob()

Security Sandbox:
- Protected paths: `.env*`, `.vercel/`, `.aws/`, `.github/workflows/secrets`
- Forbidden commands: `rm -rf`, `sudo`, `chmod 777`, `shutdown`

Git Strategy (3 Modes):
- Manual: Local Git only (no push)
- Personal: GitHub individual (auto-push)
- Team: GitHub team (Draft PR + review)

Quality Gates (TRUST 5):
- Test coverage ≥ 85%
- Security validation (OWASP)
- Compliance (GDPR, SOC 2, ISO 27001)

---

## Implementation Guide (5 minutes)

### Agent-First Mandate

Rule: Never execute tasks directly. Always delegate to specialized agents.

FORBIDDEN Pattern:
```python
# WRONG - Direct execution
def process_user_data():
 with open("data.txt") as f: # Direct file operation
 data = f.read()
 return process(data)
```

REQUIRED Pattern:
```python
# CORRECT - Agent delegation
result = await Task(
 subagent_type="code-backend",
 prompt="Process user data from data.txt",
 context={"file_path": "data.txt"}
)
```

Why Agent-First?:
- Security: Agents validate before execution
- Traceability: All actions logged via Task()
- Quality: Agents enforce TRUST 5 gates
- Scalability: Parallel/sequential delegation

---

### Tool Usage Restrictions

Allowed Tools (4 categories):

| Tool | Purpose | Example |
|------|---------|---------|
| `Task()` | Agent delegation | `Task("code-backend", "Implement API")` |
| `AskUserQuestion()` | User interaction | `AskUserQuestion(questions=[...])` |
| `Skill()` | Knowledge invocation | `Skill("moai-foundation-core")` |
| `MCP Servers` | External integrations | Context7, Playwright, Pencil |

Forbidden Tools (Why?):

| Tool | Reason |
|------|--------|
| `Read()`, `Write()`, `Edit()` | Use Task() for file operations (security validation) |
| `Bash()` | Use Task() for system operations (command validation) |
| `Grep()`, `Glob()` | Use Task() for file search (permission checks) |
| `TodoWrite()` | Use Task() for tracking (audit trail) |

Delegation Protocol:
```python
result = await Task(
 subagent_type="specialized_agent", # Agent name (lowercase, hyphenated)
 prompt="Clear, specific task description", # What to do
 context={ # Required information
 "relevant_data": "necessary_information",
 "constraints": "limitations_and_requirements"
 }
)
```

---

### Security Sandbox (Always Active)

Enabled Features:
- Tool usage validation and logging
- File access restrictions (deny list)
- Command execution limits (forbidden list)
- Permission enforcement (RBAC)

File Access Restrictions (Protected Paths):
```
DENIED ACCESS (Security Risk):
- .env* # Environment variables (secrets)
- .vercel/ # Vercel deployment config
- .netlify/ # Netlify deployment config
- .firebase/ # Firebase config
- .aws/ # AWS credentials
- .github/workflows/secrets # GitHub Actions secrets
- .kube/config # Kubernetes config
- ~/.ssh/ # SSH keys
```

Command Restrictions (Forbidden Commands):
```
FORBIDDEN (Destructive):
- rm -rf # Recursive force delete
- sudo # Privilege escalation
- chmod 777 # Insecure permissions
- dd # Disk operations
- mkfs # Filesystem formatting
- shutdown # System shutdown
- reboot # System reboot
```

Data Protection Rules:
- Never log passwords, API keys, or tokens
- Use environment variables for sensitive config
- Validate all user inputs before processing
- Encrypt sensitive data at rest

---

### Permission System (RBAC)

4 Permission Levels:

| Level | Name | Access | Use Case |
|-------|------|--------|----------|
| 1 | Read-only | File exploration, code analysis | `Explore`, `Plan` |
| 2 | Validated Write | File creation with validation | `workflow-ddd`, `workflow-docs` |
| 3 | System | Limited system operations | `infra-devops`, `core-git` |
| 4 | Security | Security analysis and enforcement | `security-expert`, `core-quality` |

Agent Permissions:
- Read Agents: File system exploration, code analysis
- Write Agents: File creation and modification (with validation)
- System Agents: Limited system operations (validated commands)
- Security Agents: Security analysis and validation

MCP Server Permissions:

| MCP Server | Permissions |
|------------|-------------|
| Context7 | Library documentation access, API reference resolution, version checking |
| Playwright | Browser automation, screenshot capture, UI simulation, E2E testing |
| Pencil | Design system access, .pen file editing, design-to-code, style guides, variables |

---

## Advanced Implementation (10+ minutes)

### Git Strategy 3-Mode System

MoAI automatically adjusts Git workflow based on `config.json` settings.

Key Configuration Fields:
- `git_strategy.mode`: Git mode selection (manual, personal, team)
- `git_strategy.branch_creation.prompt_always`: Prompt user for every SPEC (true/false)
- `git_strategy.branch_creation.auto_enabled`: Enable auto branch creation (true/false)

3-Mode System Overview (Two-Level Control):

| Configuration | Manual Mode | Personal/Team Mode | Effect |
|---------------|-------------|-------------------|--------|
| `prompt_always=true, auto_enabled=false` | Prompt each time | Prompt each time | Maximum control (default) |
| `prompt_always=false, auto_enabled=false` | Auto skip | Wait for approval | Manual=skip, Personal/Team=auto after approval |
| `prompt_always=false, auto_enabled=true` | Auto skip | Auto create | Full automation |

---

#### Mode 1: Manual (Local Git Only)

Configuration (default):
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

MoAI's Behavior (prompt_always=true):
1. When running `/moai:1-plan`, user prompted: "Create branch?"
 - Auto create → Creates feature/SPEC-001
 - Use current branch → Continues on current branch
2. All DDD commits saved locally only (automatic)
3. Push performed manually

Configuration (auto skip):
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

MoAI's Behavior (prompt_always=false):
- All SPECs automatically work on current branch (no branch creation)
- No user prompts

Use Case: Personal projects, GitHub not used, local Git only

---

#### Mode 2: Personal (GitHub Personal Project)

Configuration (default - prompt each time):
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

MoAI's Behavior (prompt_always=true):
1. When running `/moai:1-plan`, user prompted: "Create branch?"
 - Auto create → Creates feature/SPEC-002 + auto push
 - Use current branch → Commits directly on current branch
2. Running `/moai:2-run`: DDD commits + auto push
3. Running `/moai:3-sync`: Doc commits + suggest PR creation (user choice)

Configuration (auto after approval):
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

MoAI's Behavior (prompt_always=false, auto_enabled=false):
1. When running `/moai:1-plan`, user prompted once: "Enable automatic branch creation?"
 - Yes → Auto updates config.json with `auto_enabled=true` → Creates feature/SPEC
 - No → Works on current branch, no config change
2. From next SPEC: If `auto_enabled=true`, feature branches created automatically without prompts

Configuration (full automation):
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

MoAI's Behavior (prompt_always=false, auto_enabled=true):
- Automatically creates feature/SPEC-XXX branch for every SPEC
- No user prompts (full automation)
- All DDD and documentation commits auto-pushed to feature branch

Use Case: Personal GitHub projects, fast development speed needed

---

#### Mode 3: Team (GitHub Team Project)

Configuration (default - prompt each time):
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

MoAI's Behavior (prompt_always=true):
1. When running `/moai:1-plan`, user prompted: "Create branch?"
 - Auto create → Creates feature/SPEC-003 + auto create Draft PR
 - Use current branch → Proceeds on current branch (not recommended)
2. Running `/moai:2-run`: DDD commits + auto push (to feature branch)
3. Running `/moai:3-sync`: Doc commits + prepare PR
4. Team code review required (minimum 1 reviewer)
5. After approval: Merge (Squash or Merge)

Configuration (auto after approval):
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

MoAI's Behavior (prompt_always=false, auto_enabled=false):
1. When running `/moai:1-plan`, user prompted once: "Enable automatic branch creation and Draft PR creation?"
 - Yes → Auto updates config.json with `auto_enabled=true` → Creates feature/SPEC + Draft PR
 - No → Works on current branch, no config change
2. From next SPEC: If `auto_enabled=true`, feature branches + Draft PRs created automatically without prompts

Configuration (full automation):
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

MoAI's Behavior (prompt_always=false, auto_enabled=true):
- Automatically creates feature/SPEC-XXX branch + Draft PR for every SPEC
- No user prompts (full automation)
- All DDD and documentation commits auto-pushed to feature branch
- Maintains Draft PR status (until team review complete)

Use Case: Team projects, code review required, quality management needed

---

### `/clear` Execution Rule

Mandatory `/clear` After SPEC Generation:
Execute `/clear` after `/moai:1-plan` completion in ALL modes.

Why?:
- Saves 45-50K tokens (SPEC generation context)
- Prepares clean context for implementation
- Prevents token overflow

When to Execute `/clear`:
1. Immediately after `/moai:1-plan` (mandatory)
2. When context > 150K tokens
3. After 50+ conversation messages

---

### Quality Gates (TRUST 5 Framework)

Test-First:
- Every implementation must start with tests
- Test coverage must exceed 85%
- Tests must validate all requirements
- Failed tests must block deployment

Readable:
- Code must follow established style guidelines
- Variable and function names must be descriptive
- Complex logic must include comments
- Code structure must be maintainable

Unified:
- Consistent patterns across codebase
- Standardized naming conventions
- Uniform error handling
- Consistent documentation format

Secured:
- Security validation through security-expert
- OWASP compliance checking
- Input sanitization and validation
- Secure coding practices

Trackable:
- All changes must have clear origin
- Implementation must link to specifications
- Test coverage must be verifiable
- Quality metrics must be tracked

---

### Compliance Requirements

#### Legal and Regulatory

Data Privacy:
- GDPR compliance for user data
- CCPA compliance for California users
- Data minimization principles
- Right to deletion implementation

Security Standards:
- OWASP Top 10 compliance
- SOC 2 Type II controls
- ISO 27001 security management
- NIST Cybersecurity Framework

#### Industry Standards

Development Standards:
- ISO/IEC 27001 security management
- ISO/IEC 9126 software quality
- IEEE 730 software engineering standards
- Agile methodology compliance

Documentation Standards:
- IEEE 1016 documentation standards
- OpenAPI specification compliance
- Markdown formatting consistency
- Accessibility documentation (WCAG 2.1)

---

### Monitoring and Auditing

#### Activity Logging

Required Log Entries:
```python
{
 "timestamp": "2025-11-25T07:30:00Z",
 "agent": "security-expert",
 "action": "code_review",
 "files_accessed": ["src/auth.py", "tests/test_auth.py"],
 "token_usage": 5230,
 "duration_seconds": 12.5,
 "success": true,
 "quality_score": 0.95
}
```

Audit Trail Requirements:
- All agent delegations must be logged
- File access patterns must be tracked
- Security events must be recorded
- Quality metrics must be captured

#### Performance Monitoring

Key Metrics:
- Agent delegation success rate
- Average response time per task
- Token usage efficiency
- Quality gate pass rate
- Error recovery time

Alert Thresholds:
- Success rate < 95%
- Response time > 30 seconds
- Token usage > 90% of budget
- Quality gate failure rate > 5%

---

### Error Handling Protocols

#### Error Classification

Critical Errors (Immediate Stop):
- Security violations
- Data corruption risks
- System integrity threats
- Permission violations

Warning Errors (Continue with Monitoring):
- Performance degradation
- Resource limitations
- Quality gate failures
- Documentation gaps

Informational Errors (Log and Continue):
- Non-critical warnings
- Minor quality issues
- Optimization opportunities
- Style guide deviations

#### Recovery Procedures

Critical Error Recovery:
1. Immediately stop execution
2. Log error details securely
3. Notify system administrator
4. Rollback changes if possible
5. Analyze root cause

Warning Error Recovery:
1. Log warning with context
2. Continue execution with monitoring
3. Document issue for later resolution
4. Notify user of potential impact

Informational Error Recovery:
1. Log information
2. Continue execution normally
3. Add to improvement backlog
4. Monitor for patterns

---

### Resource Management

#### Token Budget Management

Phase-based Allocation:
- Planning Phase: 30K tokens maximum
- Implementation Phase: 180K tokens maximum
- Documentation Phase: 40K tokens maximum
- Total Budget: 250K tokens per feature

Optimization Requirements:
- Execute `/clear` immediately after SPEC creation
- Monitor token usage continuously
- Use efficient context loading
- Cache reusable results

#### File System Management

Allowed Operations:
- Read files within project directory
- Write files to designated locations
- Create documentation in `.moai/` directory
- Generate code in source directories

Prohibited Operations:
- Modify system files outside project
- Access sensitive configuration files
- Delete critical system resources
- Execute unauthorized system commands

#### Memory Management

Context Optimization:
- Load only necessary files for current task
- Use efficient data structures
- Clear context between major phases
- Cache frequently accessed information

Resource Limits:
- Maximum file size: 10MB per file
- Maximum concurrent operations: 5
- Maximum context size: 150K tokens
- Maximum execution time: 5 minutes per task

---

### Enforcement Mechanisms

#### Pre-execution Hooks

Required Validations:
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

#### Post-execution Hooks

Required Checks:
```python
def post_execution_hook(result, agent_type, task):
 validations = [
 validate_output_quality(result),
 validate_security_compliance(result),
 validate_documentation_completeness(result),
 validate_test_adequacy(result)
 ]

 issues = [v for v in validations if not v.passed]
 if issues:
 raise QualityGateError("Quality gate failures detected")

 return True
```

#### Automated Enforcement

Real-time Monitoring:
- Continuous validation of execution rules
- Automatic blocking of suspicious activities
- Real-time alert generation for violations
- Automated rollback of unsafe operations

Periodic Audits:
- Daily compliance checks
- Weekly performance reviews
- Monthly security assessments
- Quarterly quality audits

---

### Exception Handling

#### Security Exceptions

Emergency Override:
- Only available to authorized administrators
- Requires explicit approval and logging
- Temporary override with strict time limits
- Full audit trail required

Justified Exceptions:
- Documented business requirements
- Risk assessment and mitigation
- Alternative security controls
- Regular review and renewal

#### Performance Exceptions

Resource Optimization:
- Dynamic resource allocation
- Load balancing across agents
- Priority queue management
- Performance tuning procedures

Emergency Procedures:
- System overload protection
- Graceful degradation strategies
- User notification systems
- Recovery automation

---

## Works Well With

Skills:
- [moai-foundation-core](../SKILL.md) - Parent skill
- [moai-foundation-context](../../moai-foundation-context/SKILL.md) - Token budget and session state

Other Modules:
- [trust-5-framework.md](trust-5-framework.md) - Quality gates detail
- [token-optimization.md](token-optimization.md) - Token management strategies
- [agents-reference.md](agents-reference.md) - Agent permission levels

Agents:
- [security-expert](agents-reference.md#tier-3-domain-experts) - Security validation
- [core-quality](agents-reference.md#tier-2-orchestration--quality) - TRUST 5 enforcement
- [core-git](agents-reference.md#tier-2-orchestration--quality) - Git workflow management

---

This comprehensive set of execution rules ensures that MoAI-ADK operates securely, efficiently, and in compliance with industry standards while maintaining high quality output.

Maintained by: MoAI-ADK Team
Status: Production Ready
