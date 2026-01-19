# Claude Code Sub-agents Formatting Guide

Complete formatting reference for creating Claude Code Sub-agents that comply with official standards and deliver specialized task execution capabilities.

Purpose: Standardized formatting guide for sub-agent creation and validation
Target: Sub-agent creators and maintainers
Last Updated: 2025-11-25
Version: 2.0.0

---

## Quick Reference (30 seconds)

Core Format: YAML frontmatter + system prompt with clear domain focus. Naming: kebab-case, unique, max 64 chars. Constraints: No sub-agent nesting, Task() delegation only, isolated context windows. Key Features: Specific domain expertise, clear trigger scenarios, proper tool permissions.

---

## Complete Sub-agent Template

```yaml
---
name: subagent-name # Required: kebab-case, unique within project
description: Use PROACTIVELY when: [specific trigger scenarios]. Called from [context/workflow]. CRITICAL: This agent MUST be invoked via Task(subagent_type='subagent-name') - NEVER executed directly.
tools: Read, Write, Edit, Bash, Grep, Glob # Optional: comma-separated, principle of least privilege
model: sonnet # Optional: sonnet/opus/haiku/inherit (default: inherit)
permissionMode: default # Optional: default/acceptEdits/dontAsk (default: default)
skills: skill1, skill2, skill3 # Optional: comma-separated skill list (auto-loaded)
---

# Sub-agent System Prompt

[Clear statement of agent's specialized role and expertise with specific domain focus.]

## Core Responsibilities

Primary Domain: [Specific domain of expertise]
Key Capabilities: [List of 3-5 core capabilities]
Focus Areas: [Specific areas within the domain]

## Workflow Process

### Phase 1: [Specific phase name]
[Clear description of what happens in this phase]
- [Specific action 1]
- [Specific action 2]
- [Expected outcome]

### Phase 2: [Specific phase name]
[Clear description of what happens in this phase]
- [Specific action 1]
- [Specific action 2]
- [Expected outcome]

### Phase 3: [Specific phase name]
[Clear description of what happens in this phase]
- [Specific action 1]
- [Specific action 2]
- [Expected outcome]

## Critical Constraints

- No sub-agent spawning: This agent CANNOT create other sub-agents. Use Task() delegation for complex workflows.
- Domain focus: Stay within defined domain expertise. Delegate to other agents for outside-domain tasks.
- Tool permissions: Only use tools explicitly granted in the frontmatter.
- Quality standards: All outputs must meet [specific quality standards]

## Example Workflows

Example 1: [Specific task type]
```
Input: [Example input scenario]
Process: [Step-by-step processing approach]
Output: [Expected output format and content]
```

Example 2: [Another task type]
```
Input: [Example input scenario]
Process: [Step-by-step processing approach]
Output: [Expected output format and content]
```

## Integration Patterns

When to Use: Clear scenarios for invoking this sub-agent
- [Trigger scenario 1]
- [Trigger scenario 2]
- [Trigger scenario 3]

Delegation Targets: When to delegate to other sub-agents
- [Other agent name] for [specific task type]
- [Another agent name] for [specific task type]

## Quality Standards

- [Standard 1]: Specific quality requirement
- [Standard 2]: Specific quality requirement
- [Standard 3]: Specific quality requirement
```

---

## Frontmatter Field Specifications

### Required Fields

#### `name` (String)
Format: kebab-case (lowercase, numbers, hyphens only)
Length: Maximum 64 characters
Uniqueness: Must be unique within project
Pattern: `[domain]-[specialization]` or `[function]-expert`
Examples:
- `code-backend` (backend domain specialization)
- `frontend-developer` (frontend specialization)
- `api-designer` (API design specialization)
- `security-auditor` (security specialization)
- `MyAgent` (uppercase, spaces)
- `agent_v2` (underscore)
- `this-name-is-way-too-long-and-exceeds-the-sixty-four-character-limit`

#### `description` (String)
Format: Natural language with specific components
Required Components:
1. "Use PROACTIVELY when:" clause with specific trigger scenarios
2. "Called from" clause indicating workflow context
3. "CRITICAL: This agent MUST be invoked via Task(subagent_type='...')" clause

Examples:
- `Use PROACTIVELY for backend architecture, API design, server implementation, database integration, or microservices architecture. Called from /moai:1-plan and task delegation workflows. CRITICAL: This agent MUST be invoked via Task(subagent_type='code-backend') - NEVER executed directly.`
- `Backend development agent` (too vague, missing required clauses)
- `Helps with backend stuff` (unprofessional, missing trigger scenarios)

### Optional Fields

#### `tools` (String List)
Format: Comma-separated list, no brackets
Purpose: Principle of least privilege
Default: All available tools if omitted
Examples:
```yaml
# CORRECT: Minimal specific tools
tools: Read, Write, Edit, Bash

# CORRECT: Analysis-focused tools
tools: Read, Grep, Glob, WebFetch

# CORRECT: Documentation tools with MCP
tools: Read, Write, mcp__context7__resolve-library-id, mcp__context7__get-library-docs

# WRONG: YAML array format
tools: [Read, Write, Bash]

# WRONG: Overly permissive
tools: Read, Write, Edit, Bash, Grep, Glob, WebFetch, MultiEdit, TodoWrite, AskUserQuestion
```

#### `model` (String)
Options: `sonnet`, `opus`, `haiku`, `inherit`
Default: `inherit`
Recommendations:
- `sonnet`: Complex reasoning, architecture, research
- `opus`: Maximum quality for critical tasks
- `haiku`: Fast execution, well-defined tasks
- `inherit`: Let context decide (default)

Examples:
```yaml
# Appropriate for complex reasoning
model: sonnet

# Appropriate for fast, well-defined tasks
model: haiku

# Default behavior
# model: inherit (not specified)
```

#### `permissionMode` (String)
Options: `default`, `acceptEdits`, `dontAsk`
Default: `default`
Purpose: Control tool permission prompts
Examples:
```yaml
# Default behavior
permissionMode: default

# Accept file edits without asking
permissionMode: acceptEdits

# Never ask for permissions (risky)
permissionMode: dontAsk
```

#### `skills` (String List)
Format: Comma-separated list of skill names
Purpose: Auto-load specific skills when agent starts
Loading: Skills available automatically, no explicit invocation needed
Examples:
```yaml
# Load language and domain skills
skills: moai-lang-python, moai-domain-backend, moai-context7-integration

# Load quality and documentation skills
skills: moai-foundation-quality, moai-docs-generation, moai-cc-claude-code
```

---

## System Prompt Structure

### 1. Agent Identity and Role

Clear Role Definition:
```markdown
# Backend Expert 

You are a specialized backend architecture expert focused on designing and implementing scalable, secure, and maintainable backend systems.
```

Domain Expertise Statement:
```markdown
## Core Responsibilities

Primary Domain: Backend architecture and API development
Key Capabilities: REST/GraphQL API design, microservices architecture, database optimization, security implementation
Focus Areas: Scalability, security, performance optimization
```

### 2. Workflow Process Definition

Phase-based Structure:
```markdown
## Workflow Process

### Phase 1: Requirements Analysis
1. Parse user requirements to extract technical specifications
2. Identify performance and scalability requirements
3. Assess security and compliance needs
4. Determine technology stack constraints

### Phase 2: Architecture Design
1. Design API schemas and data models
2. Plan database architecture and relationships
3. Define service boundaries and interfaces
4. Establish security and authentication patterns

### Phase 3: Implementation Planning
1. Create implementation roadmap with milestones
2. Specify required dependencies and frameworks
3. Define testing strategy and quality gates
4. Plan deployment and monitoring approach
```

### 3. Constraints and Boundaries

Critical Constraints Section:
```markdown
## Critical Constraints

- No sub-agent spawning: This agent CANNOT create other sub-agents. Use Task() delegation only.
- Domain focus: Stay within backend domain. Delegate frontend tasks to code-frontend.
- Security-first: All designs must pass OWASP validation.
- Performance-aware: Include scalability and optimization considerations.
```

### 4. Example Workflows

Concrete Examples:
```markdown
## Example Workflows

REST API Design:
```
Input: "Design user management API"
Process:
1. Extract entities: User, Profile, Authentication
2. Design endpoints: /users, /auth, /profiles
3. Define data models and validation rules
4. Specify authentication and authorization flows
5. Document error handling and status codes
6. Include rate limiting and security measures
Output: Complete API specification with:
- Endpoint definitions (/users, /auth, /profiles)
- Data models and validation rules
- Authentication and authorization flows
- Error handling and status codes
- Rate limiting and security measures
```
```

### 5. Integration Patterns

When to Use Section:
```markdown
## Integration Patterns

When to Use:
- Designing new backend APIs and services
- Architecting microservices systems
- Optimizing database performance and queries
- Implementing authentication and authorization
- Conducting backend security audits

Delegation Targets:
- `data-database` for complex database schema design
- `security-expert` for advanced security analysis
- `performance-engineer` for performance optimization
- `api-designer` for detailed API specification
```

### 6. Quality Standards

Specific Quality Requirements:
```markdown
## Quality Standards

- API Documentation: All APIs must include comprehensive OpenAPI specifications
- Security Compliance: All designs must pass OWASP Top 10 validation
- Performance: Include benchmarks and optimization strategies
- Testing: Specify unit and integration testing requirements
- Monitoring: Define observability and logging patterns
```

---

## Common Sub-agent Patterns

### 1. Domain Expert Pattern

Purpose: Deep expertise in specific technical domain
Structure: Domain-focused responsibilities, specialized workflows
Examples: `code-backend`, `code-frontend`, `data-database`

```yaml
---
name: code-backend
description: Use PROACTIVELY for backend architecture, API design, server implementation, database integration, or microservices architecture. Called from /moai:1-plan and task delegation workflows.
tools: Read, Write, Edit, Bash, WebFetch, Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
model: sonnet
skills: moai-domain-backend, moai-essentials-perf, moai-context7-integration
---

# Backend Expert 

You are a specialized backend architecture expert focused on designing and implementing scalable, secure, and maintainable backend systems.

## Core Responsibilities

Primary Domain: Backend architecture and API development
Key Capabilities: REST/GraphQL API design, microservices architecture, database optimization, security implementation
Focus Areas: Scalability, security, performance optimization
```

### 2. Tool Specialist Pattern

Purpose: Expertise in specific tools or technologies
Structure: Tool-focused workflows, integration patterns
Examples: `format-expert`, `support-debug`, `workflow-docs`

```yaml
---
name: format-expert
description: Use PROACTIVELY for code formatting, style consistency, linting configuration, and automated code quality improvements. Called from /moai:2-run quality gates and task delegation workflows.
tools: Read, Write, Edit, Bash, Grep, Glob
model: haiku
skills: moai-code-quality, moai-cc-configuration
---

# Code Format Expert

You are a code formatting and style consistency expert specializing in automated code quality improvements and standardized formatting.

## Core Responsibilities

Primary Domain: Code formatting and style consistency
Key Capabilities: Multi-language formatting, linting configuration, style guide enforcement, automated quality improvements
Focus Areas: Code readability, consistency, maintainability
```

### 3. Process Orchestrator Pattern

Purpose: Manage complex multi-step processes
Structure: Phase-based workflows, coordination patterns
Examples: `workflow-ddd`, `agent-factory`, `skill-factory`

```yaml
---
name: workflow-ddd
description: Execute ANALYZE-PRESERVE-IMPROVE DDD cycle for implementing features with behavior preservation. Called from /moai:2-run SPEC implementation and task delegation workflows.
tools: Read, Write, Edit, Bash, Grep, Glob, MultiEdit, TodoWrite
model: sonnet
skills: moai-lang-python, moai-domain-testing, moai-foundation-quality
---

# DDD Implementation Expert

You are a Domain-Driven Development implementation expert specializing in the ANALYZE-PRESERVE-IMPROVE cycle for robust feature development.

## Core Responsibilities

Primary Domain: DDD implementation and behavior preservation
Key Capabilities: ANALYZE-PRESERVE-IMPROVE cycle, characterization tests, coverage optimization, quality gates
Focus Areas: Behavior-first development, comprehensive coverage, code quality
```

### 4. Quality Assurance Pattern

Purpose: Validate and improve quality of work products
Structure: Quality criteria, validation workflows, improvement recommendations
Examples: `core-quality`, `security-expert`, `core-quality`

```yaml
---
name: core-quality
description: Validate code quality against TRUST 5 framework (Testable, Readable, Unified, Secured, Trackable). Called from /moai:2-run quality validation and task delegation workflows.
tools: Read, Grep, Glob, Bash, Write, Edit
model: sonnet
skills: moai-foundation-trust, moai-code-quality, moai-security-expert
---

# Quality Gate Validator

You are a quality assurance expert specializing in comprehensive code validation using the TRUST 5 framework.

## Core Responsibilities

Primary Domain: Code quality validation and improvement
Key Capabilities: TRUST 5 framework validation, security assessment, performance analysis, maintainability review
Focus Areas: Quality standards compliance, security validation, performance optimization
```

---

## Sub-agent Creation Process

### 1. Domain Analysis

Identify Specialization Need:
- What specific domain expertise is missing?
- What tasks require specialized knowledge?
- What workflows would benefit from automation?
- What quality gaps exist in current processes?

Define Domain Boundaries:
- Clear scope of expertise
- Boundaries with other domains
- Integration points with other agents
- Delegation triggers and patterns

### 2. Capability Definition

Core Capabilities:
- List 3-5 primary capabilities
- Define measurable outcomes
- Specify tools and resources needed
- Identify integration patterns

Workflow Design:
- Phase-based process definition
- Clear decision points
- Quality validation steps
- Error handling strategies

### 3. Constraint Specification

Technical Constraints:
- Tool permissions and limitations
- No sub-agent nesting rule
- Context window isolation
- Resource usage boundaries

Quality Constraints:
- Domain-specific quality standards
- Output format requirements
- Integration compatibility
- Performance expectations

### 4. Implementation Guidelines

Naming Convention:
- Follow kebab-case format
- Include domain or function indicator
- Ensure uniqueness within project
- Keep under 64 characters

Description Writing:
- Include PROACTIVELY clause
- Specify called-from contexts
- Include Task() invocation requirement
- Provide specific trigger scenarios

System Prompt Development:
- Clear role definition
- Structured workflow process
- Specific constraints and boundaries
- Concrete example workflows

---

## Integration and Coordination Patterns

### 1. Sequential Delegation

Pattern: Agent A completes → Agent B continues
Use Case: Multi-phase workflows with dependencies
Example: `workflow-spec` → `workflow-ddd` → `workflow-docs`

```python
# Sequential delegation example
# Phase 1: Specification
spec_result = Task(
 subagent_type="workflow-spec",
 prompt="Create specification for user authentication system"
)

# Phase 2: Implementation (passes spec as context)
implementation_result = Task(
 subagent_type="workflow-ddd",
 prompt="Implement user authentication from specification",
 context={"specification": spec_result}
)

# Phase 3: Documentation (passes implementation as context)
documentation_result = Task(
 subagent_type="workflow-docs",
 prompt="Generate API documentation",
 context={"implementation": implementation_result}
)
```

### 2. Parallel Delegation

Pattern: Multiple agents work simultaneously
Use Case: Independent tasks that can be processed in parallel
Example: `code-backend` + `code-frontend` + `data-database`

```python
# Parallel delegation example
results = await Promise.all([
 Task(
 subagent_type="code-backend",
 prompt="Design backend API for user management"
 ),
 Task(
 subagent_type="code-frontend",
 prompt="Design frontend user interface for user management"
 ),
 Task(
 subagent_type="data-database",
 prompt="Design database schema for user management"
 )
])

# Integration phase
integrated_result = Task(
 subagent_type="integration-specialist",
 prompt="Integrate backend, frontend, and database designs",
 context={"results": results}
)
```

### 3. Conditional Delegation

Pattern: Route to different agents based on analysis
Use Case: Task classification and routing
Example: `security-expert` vs `performance-engineer`

```python
# Conditional delegation example
analysis_result = Task(
 subagent_type="analysis-expert",
 prompt="Analyze code issue and classify problem type"
)

if analysis_result.type == "security":
 result = Task(
 subagent_type="security-expert",
 prompt="Adddess security vulnerability",
 context={"analysis": analysis_result}
 )
elif analysis_result.type == "performance":
 result = Task(
 subagent_type="performance-engineer",
 prompt="Optimize performance issue",
 context={"analysis": analysis_result}
 )
```

---

## Error Handling and Recovery

### 1. Agent-Level Error Handling

Error Classification:
```python
# Error types and handling strategies
error_types = {
 "permission_denied": {
 "severity": "high",
 "action": "check tool permissions",
 "recovery": "request permission adjustment"
 },
 "resource_unavailable": {
 "severity": "medium",
 "action": "check resource availability",
 "recovery": "use alternative resource"
 },
 "domain_violation": {
 "severity": "high",
 "action": "delegate to appropriate agent",
 "recovery": "task redirection"
 }
}
```

### 2. Workflow Error Recovery

Recovery Strategies:
```markdown
## Error Handling Protocol

### Type 1: Tool Permission Errors
- Detection: Tool access denied
- Recovery: Log error, suggest permission adjustment, use alternative tools
- Escalation: Report to system administrator if critical

### Type 2: Domain Boundary Violations
- Detection: Request outside agent expertise
- Recovery: Delegate to appropriate specialized agent
- Documentation: Log delegation with reasoning

### Type 3: Resource Constraints
- Detection: Memory, time, or resource limits exceeded
- Recovery: Implement progressive processing, use caching
- Optimization: Suggest workflow improvements
```

### 3. Quality Assurance

Output Validation:
```python
# Quality validation checkpoints
def validate_agent_output(output, agent_type):
 """Validate agent output meets quality standards."""
 validations = [
 check_completeness(output),
 check_accuracy(output),
 check_formatting(output),
 check_domain_compliance(output, agent_type)
 ]

 return all(validations)
```

---

## Performance Optimization

### 1. Context Management

Context Window Optimization:
```markdown
## Context Optimization Strategy

### Input Context Management
- Load only essential information for task execution
- Use progressive disclosure for complex scenarios
- Implement context caching for repeated patterns

### Output Context Control
- Provide concise, focused responses
- Use structured output formats
- Implement result summarization
```

### 2. Tool Usage Optimization

Efficient Tool Patterns:
```python
# Optimized tool usage patterns
class EfficientToolUser:
 def __init__(self):
 self.tool_cache = {}
 self.batch_size = 10

 def batch_file_operations(self, file_operations):
 """Process multiple file operations efficiently."""
 # Group operations by type and location
 batches = self.group_operations(file_operations)

 # Process each batch efficiently
 for batch in batches:
 self.execute_batch(batch)

 def cache_frequently_used_data(self, data_key, data):
 """Cache frequently accessed data."""
 self.tool_cache[data_key] = data
 return data
```

### 3. Model Selection Optimization

Model Choice Guidelines:
```yaml
# Model selection optimization guidelines
model_selection:
 haiku:
 use_cases:
 - Simple, well-defined tasks
 - Fast execution required
 - Token efficiency critical
 examples:
 - Code formatting
 - Simple analysis
 - Data validation

 sonnet:
 use_cases:
 - Complex reasoning required
 - Architecture design
 - Multi-step workflows
 examples:
 - System design
 - Complex problem solving
 - Quality analysis

 opus:
 use_cases:
 - Maximum quality required
 - Critical decision making
 - Complex research tasks
 examples:
 - Security analysis
 - Research synthesis
 - Complex debugging
```

---

## Quality Assurance Framework

### 1. Pre-Publication Validation

Technical Validation:
```markdown
## Pre-Publication Checklist

### Frontmatter Validation
- [ ] Name uses kebab-case and is unique
- [ ] Description includes all required clauses
- [ ] Tool permissions follow principle of least privilege
- [ ] Model selection appropriate for task complexity

### System Prompt Validation
- [ ] Clear role definition and domain focus
- [ ] Structured workflow process defined
- [ ] Critical constraints specified
- [ ] Example workflows provided

### Integration Validation
- [ ] Delegation patterns clearly defined
- [ ] Error handling strategies documented
- [ ] Quality standards specified
- [ ] Performance considerations adddessed
```

### 2. Runtime Quality Monitoring

Performance Metrics:
```python
# Performance monitoring for sub-agents
class AgentPerformanceMonitor:
 def __init__(self):
 self.metrics = {
 'execution_time': [],
 'token_usage': [],
 'success_rate': 0.0,
 'error_patterns': {}
 }

 def record_execution(self, agent_type, execution_time, tokens, success, error=None):
 """Record execution metrics for analysis."""
 self.metrics['execution_time'].append(execution_time)
 self.metrics['token_usage'].append(tokens)

 if success:
 self.update_success_rate(True)
 else:
 self.update_success_rate(False)
 self.record_error_pattern(agent_type, error)

 def generate_performance_report(self):
 """Generate comprehensive performance report."""
 return {
 'avg_execution_time': sum(self.metrics['execution_time']) / len(self.metrics['execution_time']),
 'avg_token_usage': sum(self.metrics['token_usage']) / len(self.metrics['token_usage']),
 'success_rate': self.metrics['success_rate'],
 'common_errors': self.get_common_errors()
 }
```

### 3. Continuous Improvement

Feedback Integration:
```markdown
## Continuous Improvement Process

### User Feedback Collection
- Collect success rates and user satisfaction
- Monitor common error patterns and resolutions
- Track performance metrics and optimization opportunities
- Analyze usage patterns for improvement insights

### Iterative Enhancement
- Regular review of agent performance and accuracy
- Update workflows based on user feedback and metrics
- Optimize tool usage and model selection
- Enhance error handling and recovery mechanisms

### Quality Gate Updates
- Incorporate lessons learned into quality standards
- Update validation checklists based on new requirements
- Refine integration patterns with other agents
- Improve documentation and example workflows
```

---

## Security and Compliance

### 1. Security Constraints

Tool Permission Security:
```markdown
## Security Guidelines

### Tool Permission Principles
- Principle of Least Privilege: Only grant tools essential for agent's domain
- Regular Permission Reviews: Periodically audit and update tool permissions
- Security Impact Assessment: Consider security implications of each tool
- Secure Default Configurations: Use secure defaults for all permissions

### High-Risk Tool Management
- Bash tool: Restrict to essential system operations only
- WebFetch tool: Validate URLs and implement content sanitization
- Write/Edit tools: Implement path validation and content restrictions
- MultiEdit tool: Use with caution and implement proper validation
```

### 2. Data Protection

Privacy Considerations:
```python
# Data protection patterns
class SecureDataHandler:
 def __init__(self):
 self.sensitive_patterns = [
 r'password\s*=\s*["\'][^"\']+["\']',
 r'api_key\s*=\s*["\'][^"\']+["\']',
 r'token\s*=\s*["\'][^"\']+["\']'
 ]

 def sanitize_output(self, text):
 """Remove sensitive information from agent output."""
 for pattern in self.sensitive_patterns:
 text = re.sub(pattern, '[REDACTED]', text, flags=re.IGNORECASE)
 return text

 def validate_input(self, user_input):
 """Validate user input for security concerns."""
 security_checks = [
 self.check_for_injection_attempts(user_input),
 self.check_for_privilege_escalation(user_input),
 self.check_for_system_abuse(user_input)
 ]

 return all(security_checks)
```

---

## Advanced Sub-agent Patterns

### 1. Multi-Modal Agents

Multi-capability Design:
```yaml
---
name: full-stack-developer
description: Use PROACTIVELY for complete application development including frontend, backend, database, and deployment. Called from /moai:2-run comprehensive implementation and task delegation workflows.
tools: Read, Write, Edit, Bash, Grep, Glob, WebFetch, MultiEdit, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
model: sonnet
skills: moai-domain-backend, moai-domain-frontend, moai-domain-database, moai-devops-expert, moai-security-expert
---

# Full-Stack Developer

You are a comprehensive full-stack developer with expertise across all application layers, capable of end-to-end development from frontend to deployment.

## Core Responsibilities

Primary Domain: End-to-end application development
Key Capabilities: Frontend development, backend APIs, database design, deployment automation, security implementation
Focus Areas: Complete application lifecycle, technology integration, performance optimization
```

### 2. Adaptive Agents

Context-Aware Behavior:
```markdown
## Adaptive Behavior Patterns

### Model Selection Logic
```python
def select_optimal_model(task_complexity, time_constraints, quality_requirements):
 """Select optimal model based on task characteristics."""

 if time_constraints == "critical" and task_complexity < 5:
 return "haiku" # Fast execution for simple tasks

 if quality_requirements == "maximum":
 return "opus" # Maximum quality for critical tasks

 if task_complexity > 8 or requires_research:
 return "sonnet" # Complex reasoning and analysis

 return "inherit" # Let context decide
```

### Dynamic Tool Allocation
```python
def allocate_tools_by_task(task_type, security_level, performance_requirements):
 """Dynamically allocate tools based on task requirements."""

 base_tools = ["Read", "Grep", "Glob"]

 if task_type == "development":
 base_tools.extend(["Write", "Edit", "Bash"])

 if security_level == "high":
 base_tools.extend(["security-scanner", "vulnerability-checker"])

 if performance_requirements == "optimization":
 base_tools.extend(["profiler", "benchmark-tools"])

 return base_tools
```

### 3. Learning Agents

Knowledge Accumulation:
```markdown
## Learning and Adaptation

### Pattern Recognition
- Common Task Patterns: Identify frequent user request patterns
- Solution Templates: Develop reusable solution templates
- Error Pattern Analysis: Learn from common errors and solutions
- Performance Optimization: Continuously improve based on metrics

### Adaptive Workflows
```python
class AdaptiveWorkflow:
 def __init__(self):
 self.workflow_history = []
 self.success_patterns = {}
 self.optimization_suggestions = []

 def learn_from_execution(self, workflow, success, execution_metrics):
 """Learn from workflow execution outcomes."""
 self.workflow_history.append({
 'workflow': workflow,
 'success': success,
 'metrics': execution_metrics
 })

 if success:
 self.identify_success_pattern(workflow, execution_metrics)
 else:
 self.identify_failure_pattern(workflow, execution_metrics)

 def suggest_optimization(self, current_workflow):
 """Suggest optimizations based on learned patterns."""
 suggestions = []

 for pattern in self.success_patterns:
 if self.is_similar_workflow(current_workflow, pattern):
 suggestions.extend(pattern['optimizations'])

 return suggestions
```

---

## Maintenance and Updates

### 1. Regular Maintenance Schedule

Monthly Reviews:
```markdown
## Monthly Maintenance Checklist

### Performance Review
- [ ] Analyze execution metrics and performance trends
- [ ] Identify bottlenecks and optimization opportunities
- [ ] Update tool permissions based on usage patterns
- [ ] Optimize model selection based on success rates

### Quality Assurance
- [ ] Review error patterns and success rates
- [ ] Update example workflows based on user feedback
- [ ] Validate integration with other agents
- [ ] Test compatibility with latest Claude Code version

### Documentation Updates
- [ ] Update system prompt based on lessons learned
- [ ] Refresh example workflows and use cases
- [ ] Update integration patterns and delegation targets
- [ ] Document known limitations and workarounds
```

### 2. Version Management

Semantic Versioning:
```yaml
# Version update guidelines
version_updates:
 major_changes:
 - Breaking changes to agent interface or workflow
 - Significant changes to domain expertise
 - Removal of core capabilities
 - Changes to required tool permissions

 minor_changes:
 - Addition of new capabilities within domain
 - Enhanced error handling and recovery
 - Performance optimizations
 - Integration improvements

 patch_changes:
 - Bug fixes and error corrections
 - Documentation improvements
 - Minor workflow enhancements
 - Security updates and patches
```

### 3. Continuous Monitoring

Real-time Monitoring:
```python
# Agent monitoring system
class SubAgentMonitor:
 def __init__(self):
 self.active_agents = {}
 self.performance_metrics = {}
 self.error_rates = {}

 def track_agent_execution(self, agent_name, execution_data):
 """Track real-time agent execution metrics."""
 self.update_performance_metrics(agent_name, execution_data)
 self.update_error_rates(agent_name, execution_data)

 # Alert on performance degradation
 if self.is_performance_degraded(agent_name):
 self.send_performance_alert(agent_name)

 def generate_health_report(self):
 """Generate comprehensive health report for all agents."""
 return {
 'active_agents': len(self.active_agents),
 'overall_performance': self.calculate_overall_performance(),
 'error_trends': self.analyze_error_trends(),
 'optimization_opportunities': self.identify_optimizations()
 }
```

---

Version: 2.0.0
Compliance: Claude Code Official Standards
Last Updated: 2025-11-25
Quality Gates: Technical + Integration + Security
Pattern Library: 6+ proven sub-agent patterns

Generated with Claude Code using official documentation and best practices.
