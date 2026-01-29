# Claude Code Custom Slash Commands - Official Documentation Reference

Source: https://code.claude.com/docs/en/slash-commands#custom-slash-commands

## Key Concepts

### What are Custom Slash Commands?

Custom slash commands are user-defined commands that extend Claude Code's functionality with specialized workflows, automations, and integrations. They follow a specific file structure and syntax, enabling powerful command→agent→skill orchestration patterns.

### Command Architecture

Command Execution Flow:

```
User Input → Command File → Parameter Parsing → Agent Delegation → Skill Execution
```

Command Components:

1. Command File: Markdown file with frontmatter and implementation
2. Parameter System: Argument parsing and validation
3. Agent Orchestration: Multi-agent workflow coordination
4. Skill Integration: Specialized knowledge and capabilities
5. Result Processing: Output formatting and user feedback

## Command File Structure

### Storage Locations

Command Directory Priority:

1. Personal Commands: `~/.claude/commands/` (highest priority)
2. Project Commands: `.claude/commands/` (team-shared)
3. Plugin Commands: Bundled with installed packages (lowest priority)

Directory Structure:

```
.claude/commands/
 category1/
 my-command.md
 related-command.md
 category2/
 specialized-command.md
 README.md # Command index and documentation
```

### Command Naming Convention

IMPORTANT: Command name is automatically derived from file path structure:

- `.claude/commands/{namespace}/{command-name}.md` → `/{namespace}:{command-name}`
- `.claude/commands/my-command.md` → `/my-command`
- Example: `.claude/commands/moai/fix.md` → `/moai:fix`

DO NOT include a `name` field in frontmatter - it is not officially supported.

### Command File Format

Official Frontmatter Fields (per Claude Code documentation):

```markdown
---
description: Brief description of what the command does
argument-hint: [action] [target] [options]
allowed-tools: Bash, Read, Write
model: haiku
---
```

Supported Frontmatter Fields:

- `description` - Command description shown in /help (recommended)
- `argument-hint` - Argument syntax hint for autocomplete
- `allowed-tools` - Tools this command can invoke
- `model` - Override default model (haiku, sonnet, opus)
- `hooks` - Hook definitions for command execution
- `disable-model-invocation` - Prevent Skill tool invocation

All frontmatter options are optional; commands work without frontmatter.

Complete Command Template:

````markdown
---
description: Brief description of what the command does and when to use it
argument-hint: [action] [target] [options]
allowed-tools: Bash(git add:*), Bash(git status:*), Read, Write
model: haiku
---

# Command Implementation

## Quick Reference

Purpose: One-line summary of command purpose
Usage: `/my-command <action> <target> [options]`
Examples: 2-3 common usage patterns

## Implementation

### Phase 1: Input Validation

```bash
# Validate required parameters
if [ -z "$1" ]; then
 echo "Error: Action parameter is required"
 echo "Usage: /my-command <action> <target> [options]"
 exit 1
fi
```
````

### Phase 2: Agent Delegation

```python
# Delegate to appropriate agents
action="$1"
target="$2"

case "$action" in
 "create")
 Task(
 subagent_type="spec-builder",
 prompt="Create specification for $target",
 context={"user_input": "$ARGUMENTS"}
 )
 ;;
 "validate")
 Task(
 subagent_type="quality-gate",
 prompt="Validate configuration in $target",
 context={"config_file": "$target"}
 )
 ;;
esac
```

### Phase 3: Result Processing

```python
# Process agent results and format output
results = await Promise.all(agent_tasks)

# Format results for user
formatted_output = format_command_output(results, action)

# Provide user feedback
echo "Command completed successfully"
echo "Results: $formatted_output"
```

````

## Parameter System

### Parameter Types

String Parameters:
```yaml
parameters:
 - name: feature_name
 description: Name of the feature to implement
 required: true
 type: string
 validation:
 pattern: "^[a-z][a-z0-9-]*$"
 minLength: 3
 maxLength: 50
````

File Reference Parameters:

```yaml
parameters:
 - name: config_file
 description: Configuration file to process
 required: false
 type: string
 allowFileReference: true
 validation:
 fileExists: true
 fileExtensions: [".yaml", ".json", ".toml"]
```

Boolean Parameters:

```yaml
parameters:
 - name: verbose
 description: Enable verbose output
 required: false
 type: boolean
 default: false
 shortFlag: "-v"
 longFlag: "--verbose"
```

Choice Parameters:

```yaml
parameters:
 - name: environment
 description: Target environment
 required: false
 type: string
 values: [development, staging, production]
 default: development
```

Object Parameters:

```yaml
parameters:
 - name: options
 description: Additional options object
 required: false
 type: object
 properties:
 timeout:
 type: number
 default: 300
 retries:
 type: number
 default: 3
 additionalProperties: true
```

### Parameter Access Patterns

Positional Arguments:

```bash
# $1, $2, $3... for positional arguments
action="$1" # First argument
target="$2" # Second argument
options="$3" # Third argument

# All arguments as single string
all_args="$ARGUMENTS"
```

Named Arguments:

```bash
# Parse named arguments using getopts
while getopts ":f:t:v" opt; do
 case $opt in
 f) file="$OPTARG" ;;
 t) timeout="$OPTARG" ;;
 v) verbose=true ;;
 esac
done
```

File References:

```bash
# File reference handling with @ prefix
if [[ "$target" == @* ]]; then
 file_path="${target#@}"
 if [ -f "$file_path" ]; then
 file_content=$(cat "$file_path")
 else
 echo "Error: File not found: $file_path"
 exit 1
 fi
fi
```

## Agent Orchestration Patterns

### Sequential Agent Workflow

Linear Execution Pattern:

```python
# Phase 1: Analysis
analysis = Task(
 subagent_type="spec-builder",
 prompt="Analyze requirements for $ARGUMENTS",
 context={"user_input": "$ARGUMENTS"}
)

# Phase 2: Implementation (passes analysis results)
implementation = Task(
 subagent_type="ddd-implementer",
 prompt="Implement based on analysis",
 context={"analysis": analysis, "spec_id": analysis.spec_id}
)

# Phase 3: Quality Validation
validation = Task(
 subagent_type="quality-gate",
 prompt="Validate implementation",
 context={"implementation": implementation}
)
```

### Parallel Agent Workflow

Concurrent Execution Pattern:

```python
# Independent parallel execution
results = await Promise.all([
 Task(
 subagent_type="backend-expert",
 prompt="Backend implementation for $1"
 ),
 Task(
 subagent_type="frontend-expert",
 prompt="Frontend implementation for $1"
 ),
 Task(
 subagent_type="docs-manager",
 prompt="Documentation for $1"
 )
])

# Integration phase
integration = Task(
 subagent_type="quality-gate",
 prompt="Integrate all components",
 context={"components": results}
)
```

### Conditional Agent Workflow

Dynamic Agent Selection:

```python
# Route based on analysis results
if analysis.has_database_issues:
 result = Task(
 subagent_type="database-expert",
 prompt="Optimize database",
 context={"issues": analysis.database_issues}
 )
elif analysis.has_api_issues:
 result = Task(
 subagent_type="backend-expert",
 prompt="Fix API issues",
 context={"issues": analysis.api_issues}
 )
else:
 result = Task(
 subagent_type="quality-gate",
 prompt="General quality check",
 context={"analysis": analysis}
 )
```

## Command Examples

### Simple Validation Command

Configuration Validator:

````markdown
---
name: validate-config
description: Validate configuration files against schema and best practices
usage: |
 /validate-config <file> [options]
 Examples:
 /validate-config app.yaml
 /validate-config @production-config.json --strict
parameters:
 - name: file
 description: Configuration file to validate
 required: true
 type: string
 allowFileReference: true
 - name: strict
 description: Enable strict validation mode
 required: false
 type: boolean
 default: false
---

# Configuration Validator

## Quick Reference

Validates YAML/JSON configuration files against schemas and best practices.

## Implementation

### Input Processing

```bash
config_file="$1"
strict_mode="$2"

# Handle file reference
if [[ "$config_file" == @* ]]; then
 config_file="${config_file#@}"
fi

# Validate file exists
if [ ! -f "$config_file" ]; then
 echo "Error: Configuration file not found: $config_file"
 exit 1
fi
```
````

### Validation Execution

```python
# Determine validation strategy
if [[ "$config_file" == *.yaml ]] || [[ "$config_file" == *.yml ]]; then
 validator = "yaml-validator"
elif [[ "$config_file" == *.json ]]; then
 validator = "json-validator"
else
 echo "Error: Unsupported file format"
 exit 1
fi

# Execute validation
Task(
 subagent_type="quality-gate",
 prompt="Validate $config_file using $validator" +
 (" --strict" if strict_mode else ""),
 context={
 "file_path": config_file,
 "validator": validator,
 "strict_mode": strict_mode == "--strict"
 }
)
```

````

### Complex Multi-Phase Command

Feature Implementation Workflow:
```markdown
---
name: implement-feature
description: Complete feature implementation workflow from spec to deployment
usage: |
 /implement-feature "Feature description" [options]
 Examples:
 /implement-feature "Add user authentication with JWT"
 /implement-feature "Create API endpoints" --skip-tests
parameters:
 - name: description
 description: Feature description to implement
 required: true
 type: string
 - name: skip_tests
 description: Skip test implementation phase
 required: false
 type: boolean
 default: false
 - name: environment
 description: Target environment
 required: false
 type: string
 values: [development, staging, production]
 default: development
---

# Feature Implementation Workflow

## Quick Reference
Complete DDD-based feature implementation from specification to deployment.

## Implementation

### Phase 1: Specification Generation
```python
# Generate comprehensive specification
spec_result = Task(
 subagent_type="spec-builder",
 prompt="Create detailed specification for: $1",
 context={
 "feature_description": "$1",
 "environment": "$3"
 }
)

spec_id = spec_result.spec_id
echo "Specification created: $spec_id"
````

### Phase 2: Implementation Planning

```python
# Plan implementation approach
plan_result = Task(
 subagent_type="plan",
 prompt="Create implementation plan for $spec_id",
 context={
 "spec_id": spec_id,
 "skip_tests": "$2"
 }
)
```

### Phase 3: Test Implementation (if not skipped)

```python
if [ "$2" != "--skip-tests" ]; then
 # RED phase: Write failing tests
 test_result = Task(
 subagent_type="test-engineer",
 prompt="Write comprehensive tests for $spec_id",
 context={"spec_id": spec_id}
 )
fi
```

### Phase 4: Feature Implementation

```python
# IMPROVE phase: Implement feature
implementation_result = Task(
 subagent_type="ddd-implementer",
 prompt="Implement feature for $spec_id",
 context={
 "spec_id": spec_id,
 "tests_available": "$2" != "--skip-tests"
 }
)
```

### Phase 5: Quality Assurance

```python
# REFACTOR and validation
quality_result = Task(
 subagent_type="quality-gate",
 prompt="Validate implementation for $spec_id",
 context={
 "implementation": implementation_result,
 "test_coverage": "90%" if "$2" != "--skip-tests" else "0%"
 }
)
```

### Phase 6: Documentation

```python
# Generate documentation
docs_result = Task(
 subagent_type="docs-manager",
 prompt="Create documentation for $spec_id",
 context={"spec_id": spec_id}
)
```

### Results Summary

```python
echo "Feature implementation completed!"
echo "Specification: $spec_id"
echo "Implementation: $(echo $implementation_result | jq .status)"
echo "Quality Score: $(echo $quality_result | jq .score)"
echo "Documentation: $(echo $docs_result | jq .generated_files)"
```

````

### Integration Command

CI/CD Pipeline Integration:
```markdown
---
name: deploy
description: Deploy application with comprehensive validation and rollback capability
usage: |
 /deploy [environment] [options]
 Examples:
 /deploy staging
 /deploy production --skip-tests --dry-run
parameters:
 - name: environment
 description: Target deployment environment
 required: false
 type: string
 values: [staging, production]
 default: staging
 - name: skip_tests
 description: Skip pre-deployment tests
 required: false
 type: boolean
 default: false
 - name: dry_run
 description: Perform dry-run deployment
 required: false
 type: boolean
 default: false
---

# Deployment Pipeline

## Quick Reference
Safe deployment with validation, testing, and rollback capabilities.

## Implementation

### Pre-Deployment Validation
```python
# Environment validation
env_result = Task(
 subagent_type="devops-expert",
 prompt="Validate $1 environment configuration",
 context={"environment": "$1"}
)

# Security validation
security_result = Task(
 subagent_type="security-expert",
 prompt="Perform security pre-deployment check",
 context={"environment": "$1"}
)
````

### Testing Phase

```python
if [ "$2" != "--skip-tests" ]; then
 # Run comprehensive test suite
 test_result = Task(
 subagent_type="test-engineer",
 prompt="Execute deployment test suite",
 context={"environment": "$1"}
 )
fi
```

### Deployment Execution

```python
if [ "$3" != "--dry-run" ]; then
 # Actual deployment
 deploy_result = Task(
 subagent_type="devops-expert",
 prompt="Deploy to $1 environment",
 context={
 "environment": "$1",
 "rollback_plan": true
 }
 )
else
 echo "Dry-run mode: Deployment simulated"
 deploy_result = {"status": "simulated", "environment": "$1"}
fi
```

### Post-Deployment Validation

```python
# Health check and validation
health_result = Task(
 subagent_type="monitoring-expert",
 prompt="Validate deployment health in $1",
 context={"environment": "$1"}
)

# Generate deployment report
report_result = Task(
 subagent_type="docs-manager",
 prompt="Generate deployment report",
 context={
 "environment": "$1",
 "deployment": deploy_result,
 "health": health_result
 }
)
```

````

## Command Distribution and Sharing

### Team Command Distribution

Git-Based Distribution:
```bash
# Store commands in version control
git add .claude/commands/
git commit -m "Add custom commands for team workflow"

# Team members clone and update
git pull origin main
claude commands reload
````

Package Distribution:

```bash
# Create command package
claude commands package --name "team-workflows" --version "1.0.0"

# Install command package
claude commands install team-workflows@1.0.0
```

### Command Documentation

Command Index Generation:

```markdown
# .claude/commands/README.md

## Team Command Library

### Development Commands

- `/implement-feature` - Complete feature implementation workflow
- `/validate-config` - Configuration file validation
- `/create-component` - Component scaffolding and setup

### Deployment Commands

- `/deploy` - Safe deployment with rollback
- `/rollback` - Emergency rollback procedure
- `/health-check` - System health validation

### Analysis Commands

- `/analyze-performance` - Performance bottleneck analysis
- `/security-audit` - Security vulnerability assessment
- `/code-review` - Automated code review
```

## Best Practices

### Command Design

Naming Conventions:

- Use kebab-case for command names: `implement-feature`, `validate-config`
- Keep names descriptive and action-oriented
- Avoid abbreviations and jargon
- Use consistent prefixes for related commands

Parameter Design:

- Required parameters come first
- Use descriptive parameter names
- Provide clear validation and error messages
- Support common patterns (file references, boolean flags)

Error Handling:

- Validate all inputs before processing
- Provide helpful error messages with suggestions
- Implement graceful degradation
- Support dry-run modes for destructive operations

### Performance Optimization

Efficient Agent Usage:

- Batch related operations in single agent calls
- Use parallel execution for independent tasks
- Cache results when appropriate
- Minimize context passing between agents

User Experience:

- Provide progress feedback for long-running commands
- Use clear, consistent output formatting
- Support interactive confirmation for critical operations
- Include usage examples and help text

### Security Considerations

Security Best Practices:

- Validate all file paths and inputs
- Implement principle of least privilege
- Never expose sensitive credentials in command output
- Use secure parameter handling for passwords and tokens

Audit and Logging:

- Log all command executions with parameters
- Track success/failure rates
- Monitor for unusual usage patterns
- Provide audit trails for compliance

This comprehensive reference provides all the information needed to create powerful, secure, and user-friendly custom slash commands for Claude Code.
