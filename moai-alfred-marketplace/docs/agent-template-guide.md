# Agent Template Guide

**How to write agent templates for Alfred Framework plugins**

## Overview

Agent templates define AI specialists that handle complex workflows within plugins. Each agent:
- Lives in `agents/{agent-name}.md`
- Describes agent responsibilities and tools
- Specifies interaction flow with users
- Documents skill dependencies
- Gets invoked by commands automatically

## Why Agents?

Agents enable:
- Complex multi-step workflows
- Decision-making and reasoning
- Skill orchestration
- Tool coordination
- Error handling and recovery

## Template Structure

```markdown
# Agent Name

One-line description of agent responsibility.

## Responsibilities

1. First responsibility
2. Second responsibility
3. Third responsibility

## Tools

- **Tool Name**: What this tool does

## Skills

- Skill reference for contextual guidance

## Interaction Flow

User action → Agent receives input → Agent invokes skills/tools → Output

## Examples

[Concrete examples]

## Error Handling

[Common error scenarios]
```

## Complete Example 1: PM Agent

File: `agents/pm-agent.md`

```markdown
# PM Agent

Specialist agent for project management automation and SPEC generation.

## Responsibilities

1. **Parse `/init-pm` Command Arguments**
   - Extract project name from user input
   - Validate project naming conventions
   - Determine template and options

2. **Invoke Skill-Based SPEC Generation**
   - Load `moai-foundation-ears` skill for requirement syntax
   - Invoke `moai-spec-authoring` for document templates
   - Generate spec.md with EARS format

3. **Generate Project Charter**
   - Create charter.md with governance section
   - Extract stakeholders from user input
   - Build stakeholder matrix (stakeholders.json)

4. **Create Risk Assessment Matrix**
   - Generate risk-matrix.json based on risk level
   - Identify key risks from project type
   - Create mitigation-plan.md with strategies

5. **Validate Generated SPEC**
   - Check EARS syntax compliance
   - Verify all required sections present
   - Validate @TAG chain creation

6. **Display Summary Report**
   - Show created files
   - Provide next steps
   - Report any warnings

## Tools

- **Read**: Access template files from plugin directory
  - Reads: `commands/init-pm.md`, skill templates

- **Write**: Create SPEC documents in `.moai/specs/`
  - Creates: spec.md, plan.md, acceptance.md
  - Creates: charter.md, risk-matrix.json

- **Edit**: Modify generated files based on user input
  - Updates: plan.md milestones, acceptance criteria

- **Bash**: Execute file operations
  - Creates: directories, validation scripts

## Skills

- **moai-foundation-ears**: EARS syntax reference (5 patterns)
- **moai-spec-authoring**: SPEC document templates
- **moai-foundation-specs**: @TAG validation (CODE-FIRST principle)
- **moai-plugin-scaffolding**: Plugin generation patterns

## Interaction Flow

```
User Input:
/init-pm my-awesome-project --template=enterprise --risk-level=high

                         ↓

Agent Receives:
command: "init-pm"
args: {
  projectName: "my-awesome-project",
  template: "enterprise",
  riskLevel: "high"
}

                         ↓

Agent Validates:
✅ Project name format valid
✅ Template "enterprise" supported
✅ Risk level in range (low/medium/high)

                         ↓

Agent Invokes Skills:
1. Skill("moai-foundation-ears")
   → Learn EARS patterns (Ubiquitous, Event-driven, State-driven, Optional, Unwanted)

2. Skill("moai-spec-authoring")
   → Load SPEC document templates

3. Skill("moai-plugin-scaffolding")
   → Get best practices for project structure

                         ↓

Agent Generates:
Write .moai/specs/SPEC-MY-AWESOME-001/spec.md
├─ EARS requirements with 5 patterns
├─ @TAG markers for traceability
├─ Architecture section
└─ Acceptance criteria

Write .moai/specs/SPEC-MY-AWESOME-001/plan.md
├─ Phase breakdown
├─ Timeline estimation
├─ Resource allocation
└─ Risk timeline

Write .moai/specs/SPEC-MY-AWESOME-001/acceptance.md
├─ Success criteria
├─ Quality metrics
├─ Testing requirements
└─ Deployment checklist

Write charter.md (for enterprise template)
├─ Governance structure
├─ Decision authority matrix
├─ Escalation procedures
└─ Stakeholder roles

Generate risk-matrix.json (for high risk level)
├─ 10+ identified risks
├─ Impact/Likelihood scores
├─ Mitigation strategies
└─ Contingency plans

                         ↓

Agent Validates:
✅ All files created
✅ EARS syntax validated
✅ @TAG chain verified
✅ File permissions correct

                         ↓

Output to User:
✅ Project initialization complete
📁 Created: .moai/specs/SPEC-MY-AWESOME-001/
📊 Files: spec.md, plan.md, acceptance.md, charter.md, risk-matrix.json
🚀 Next: /alfred:2-run SPEC-MY-AWESOME-001
```

## Examples

### Example 1: Basic SPEC Generation
```
User: /init-pm api-service
Agent:
  1. Creates .moai/specs/SPEC-API-SERVICE-001/
  2. Generates moai-spec template (standard)
  3. Sets risk-level to "medium" (default)
  4. Creates: spec.md, plan.md, acceptance.md
  5. Output: ✅ SPEC created, ready for /alfred:2-run
```

### Example 2: Enterprise with High Risk
```
User: /init-pm payment-system --template=enterprise --risk-level=high
Agent:
  1. Creates .moai/specs/SPEC-PAYMENT-SYSTEM-001/
  2. Loads enterprise template (governance + compliance)
  3. Generates comprehensive risk matrix (20+ risks)
  4. Creates: spec.md, plan.md, acceptance.md, charter.md, risk-matrix.json, mitigation-plan.md
  5. Output: ✅ Enterprise SPEC with governance, ready for compliance review
```

### Example 3: With Skip Charter
```
User: /init-pm monitoring-service --skip-charter
Agent:
  1. Creates .moai/specs/SPEC-MONITORING-001/
  2. Skips charter.md generation
  3. Creates: spec.md, plan.md, acceptance.md only
  4. Output: ✅ Minimal SPEC, suitable for small teams
```

## Error Handling

### Error 1: Invalid Project Name
```
User Input: /init-pm MyAwesomeProject

Agent Detection:
❌ Project name "MyAwesomeProject" invalid
   - Contains uppercase letters
   - Format must be: lowercase-with-hyphens

Agent Recovery:
🔧 Suggested fix: /init-pm my-awesome-project
   - Convert to lowercase: my-awesome-project
   - Ask user to confirm

User Confirmation: ✅ Yes, create with "my-awesome-project"

Agent Continues: Creates SPEC with corrected name
```

### Error 2: SPEC Already Exists
```
Agent Detection:
❌ SPEC-API-SERVICE-001 already exists
   Location: .moai/specs/SPEC-API-SERVICE-001/

Agent Recovery:
🔧 Options:
   1. Use different name: /init-pm api-service-v2
   2. Increment version: /init-pm api-service-v3 (creates SPEC-API-SERVICE-V3-001)
   3. Overwrite existing: /init-pm api-service --force

User Selection: Option 1 - Use api-service-v2

Agent Continues: Creates SPEC-API-SERVICE-V2-001
```

### Error 3: EARS Validation Failed
```
Agent Detection:
❌ Generated spec.md has EARS validation errors:
   - Missing "State-driven" behavior pattern
   - Optional behavior section incomplete
   - @TAG chain broken on line 125

Agent Recovery:
🔧 Regenerating with corrections:
   1. Add missing "State-driven" section
   2. Complete optional behavior documentation
   3. Reconnect @TAG chain

Agent Validation: ✅ All EARS patterns complete, @TAG chain verified

Agent Output: ✅ SPEC validated, ready for use
```

## Agent Configuration in plugin.json

```json
{
  "agents": [
    {
      "name": "pm-agent",
      "path": "agents/pm-agent.md",
      "type": "specialist",
      "description": "Project management specialist for SPEC generation"
    }
  ]
}
```

## Agent Best Practices

### 1. Clear Responsibilities

✅ **Good**:
- "Parse command arguments and validate input"
- "Invoke skills for SPEC generation"
- "Validate generated documents"

❌ **Bad**:
- "Do project management stuff"
- "Handle things"

### 2. Explicit Tools

✅ **Good**:
- List each tool with specific use case
- Show exact files accessed/created
- Document tool constraints

❌ **Bad**:
- "Use all available tools"
- Vague tool descriptions

### 3. Skill Integration

✅ **Good**:
- Reference specific skills by name
- Explain what each skill teaches
- Show skill usage in flow diagram

❌ **Bad**:
- Mention skills vaguely
- Don't link to skill content

### 4. Detailed Workflows

✅ **Good**:
- ASCII flow diagram showing steps
- Show intermediate states
- Document decision points

❌ **Bad**:
- Single paragraph description
- No visual representation

### 5. Error Recovery

✅ **Good**:
- Document 3-5 common errors
- Show detection strategy
- Provide recovery options

❌ **Bad**:
- "Errors will be handled"
- No specific scenarios

## Common Agent Patterns

### Pattern 1: Scaffolding Agent
```markdown
# Scaffolding Agent

Generate project structure from SPEC.

## Responsibilities
1. Parse scaffolding command
2. Validate SPEC document
3. Extract resource definitions
4. Generate directory structure
5. Create boilerplate code
6. Validate generated structure

## Tools
- Read: Load SPEC documents
- Write: Create project files
- Bash: Create directories, run generators

## Interaction
User: /generate-scaffold SPEC-001
  ↓
Agent: Load SPEC, extract resources
  ↓
Agent: Skill("moai-plugin-scaffolding")
  ↓
Agent: Create src/models/, src/routes/, tests/
  ↓
Output: ✅ Project structure created
```

### Pattern 2: Database Agent
```markdown
# Database Agent

Setup and manage database migrations.

## Responsibilities
1. Parse database setup command
2. Validate database configuration
3. Create SQLAlchemy models
4. Generate Alembic migrations
5. Execute migration scripts
6. Verify database state

## Tools
- Read: Load SPEC/templates
- Write: Create models and migrations
- Bash: Run migration commands

## Skills
- moai-domain-database (schema design)
- moai-lang-python (syntax)
- moai-plugin-scaffolding (patterns)
```

### Pattern 3: Configuration Agent
```markdown
# Configuration Agent

Setup and validate tool configurations.

## Responsibilities
1. Parse configuration command
2. Load configuration templates
3. Apply user settings
4. Validate configuration syntax
5. Test tool connectivity
6. Display configuration summary

## Tools
- Read: Load configuration templates
- Write: Create config files
- Edit: Modify existing configs
- Bash: Validate and test configs
```

## Linking Commands to Agents

**In plugin.json**:
```json
{
  "commands": [
    {
      "name": "init-pm",
      "path": "commands/init-pm.md",
      "description": "Initialize project management"
    }
  ],
  "agents": [
    {
      "name": "pm-agent",
      "path": "agents/pm-agent.md",
      "type": "specialist"
    }
  ]
}
```

**In command template (`commands/init-pm.md`)**:
```markdown
# /init-pm

[Command description]

## Related

- See [PM Agent](../agents/pm-agent.md) for implementation details
```

## See Also

- [Command Template Guide](./command-template-guide.md)
- [plugin.json Schema](./plugin-json-schema.md)
- [Contributing Guide](../CONTRIBUTING.md)
- [SPEC-CH08-001](../../.moai/specs/SPEC-CH08-001/spec.md)

---

**Version**: 1.0.0
**Last Updated**: 2025-10-30

🔗 Generated with [Claude Code](https://claude.com/claude-code)
