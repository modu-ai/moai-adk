---
name: "moai-cc-permission-mode"
version: "2.0.43"
created: 2025-11-18
updated: 2025-11-18
status: stable
tier: specialization
description: "Claude Code v2.0.43 Agent Permission Mode strategy (auto vs ask) for 32 agents. Complete guide for permission classification, workflow integration, and token optimization."
allowed-tools: "Read, Edit, Bash"
primary-agent: "implementation-planner"
secondary-agents: ["spec-builder", "tdd-implementer", "quality-gate"]
keywords: ["claude-code", "agents", "permissions", "auto-mode", "ask-mode", "security"]
tags: ["claude-code-v2.0.43", "advanced"]
orchestration:
  multi_agent: true
  supports_chaining: false
can_resume: false
typical_chain_position: "planning"
depends_on: ["moai-cc-hooks"]
---

# moai-cc-permission-mode

**Claude Code v2.0.43 Agent Permission Mode Strategy**

> **Primary Agent**: implementation-planner
> **Secondary Agents**: spec-builder, tdd-implementer, quality-gate
> **Version**: 2.0.43
> **Keywords**: agents, permissions, security, workflow-integration, token-optimization

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (50 lines)

**Core Purpose**: Classify 32 agents into two permission modes (auto/ask) to balance execution speed and user control.

**Two Permission Modes**:

#### Auto Mode (11 Agents)
- **Execution**: Automatic, no user approval needed
- **Use Case**: Safe, non-destructive operations
- **Examples**: SPEC generation, documentation, validation
- **Token Cost**: Lower (faster execution)

```
Agents: spec-builder, docs-manager, quality-gate, project-manager,
agent-factory, skill-factory, doc-syncer, cc-manager, format-expert,
sync-manager, trust-checker
```

#### Ask Mode (21 Agents)
- **Execution**: Requires user approval via AskUserQuestion
- **Use Case**: Code changes, infrastructure, critical decisions
- **Examples**: Implementation, architecture, security audit
- **Token Cost**: Higher (user interaction required)

```
Agents: tdd-implementer, backend-expert, frontend-expert,
performance-engineer, devops-expert, api-designer, mcp-context7-integrator,
debug-helper, implementation-planner, monitoring-expert, migration-expert,
ui-ux-expert, accessibility-expert, component-designer, trust-checker,
mcp-sequential-thinking-integrator, mcp-playwright-integrator,
mcp-notion-integrator, figma-expert, database-expert, security-expert, git-manager
```

**Quick Decision Matrix**:

| Operation Type | Mode | Reason | Examples |
|---|---|---|---|
| **Document Generation** | Auto | No code changes | SPEC docs, API docs, README |
| **Validation/Linting** | Auto | Verification only | Test coverage, code quality checks |
| **Code Implementation** | Ask | Changes codebase | Feature development, bug fixes |
| **Architecture Changes** | Ask | Critical impact | API design, database schema |
| **Infrastructure Changes** | Ask | Deployment risk | DevOps, deployment scripts |
| **Security Operations** | Ask | High risk | Encryption setup, auth changes |

---

### Level 2: Core Implementation (150 lines)

**11 Auto-Mode Agents**

#### 1. spec-builder
- **Purpose**: Generate SPEC documents in EARS format
- **Operations**: Read requirements, write SPEC files, validate format
- **Safety**: No code changes, purely documentation
- **Execution**: Automatic trigger on `/alfred:1-plan`
- **Frontmatter**:
  ```yaml
  permissionMode: auto
  skills:
    - moai-foundation-specs
  ```

#### 2. docs-manager
- **Purpose**: Generate living documentation
- **Operations**: Read code, generate markdown, create diagrams
- **Safety**: Documentation-only, no code execution
- **Execution**: Automatic trigger on `/alfred:3-sync`
- **Frontmatter**:
  ```yaml
  permissionMode: auto
  skills:
    - moai-docs-generation
  ```

#### 3. quality-gate
- **Purpose**: Validate TRUST 5 principles
- **Operations**: Run tests, check coverage, validate code quality
- **Safety**: Validation-only, no destructive changes
- **Execution**: Automatic on code completion
- **Frontmatter**:
  ```yaml
  permissionMode: auto
  skills:
    - moai-docs-validation
  ```

#### 4. project-manager
- **Purpose**: Project initialization and metadata
- **Operations**: Read project structure, create config files
- **Safety**: Setup-only, controlled changes
- **Execution**: Automatic on `/alfred:0-project`
- **Frontmatter**:
  ```yaml
  permissionMode: auto
  skills:
    - moai-project-config-manager
  ```

#### 5. agent-factory
- **Purpose**: Create new agent definitions
- **Operations**: Generate agent template files
- **Safety**: Template generation only
- **Execution**: Automatic on agent creation request
- **Frontmatter**:
  ```yaml
  permissionMode: auto
  skills:
    - moai-alfred-agent-factory
  ```

#### 6. skill-factory
- **Purpose**: Create new Skill definitions
- **Operations**: Generate Skill template files
- **Safety**: Template generation only
- **Execution**: Automatic on Skill creation request
- **Frontmatter**:
  ```yaml
  permissionMode: auto
  skills:
    - skill-factory
  ```

#### 7. doc-syncer
- **Purpose**: Auto-synchronize documentation with code
- **Operations**: Read code, update docs automatically
- **Safety**: Documentation sync only
- **Execution**: Automatic on code changes
- **Frontmatter**:
  ```yaml
  permissionMode: auto
  skills:
    - moai-docs-generation
  ```

#### 8. cc-manager
- **Purpose**: Manage Claude Code configuration
- **Operations**: Read/write settings files, validate config
- **Safety**: Configuration-only changes
- **Execution**: Automatic on config updates
- **Frontmatter**:
  ```yaml
  permissionMode: auto
  skills:
    - moai-cc-configuration
  ```

#### 9. format-expert
- **Purpose**: Apply code formatting and style
- **Operations**: Run formatters, apply linting rules
- **Safety**: Style-only changes, no logic changes
- **Execution**: Automatic on formatting request
- **Frontmatter**:
  ```yaml
  permissionMode: auto
  skills:
    - moai-essentials-refactor
  ```

#### 10. sync-manager
- **Purpose**: Orchestrate /alfred:3-sync workflow
- **Operations**: Coordinate docs sync and validation
- **Safety**: Workflow orchestration only
- **Execution**: Automatic trigger on `/alfred:3-sync`
- **Frontmatter**:
  ```yaml
  permissionMode: auto
  skills:
    - moai-alfred-workflow
  ```

#### 11. trust-checker
- **Purpose**: Verify TRUST 5 compliance
- **Operations**: Validate code quality, test coverage, security
- **Safety**: Validation-only, no changes
- **Execution**: Automatic on compliance check
- **Frontmatter**:
  ```yaml
  permissionMode: auto
  skills:
    - moai-docs-validation
  ```

---

**21 Ask-Mode Agents**

#### Code Implementation Agents (Require User Approval)

1. **tdd-implementer**
   - **Purpose**: TDD Red-Green-Refactor cycle
   - **Requires**: User approval for code changes
   - **Frontmatter**: `permissionMode: ask`

2. **backend-expert**
   - **Purpose**: API and microservice design
   - **Requires**: User approval for architecture decisions
   - **Frontmatter**: `permissionMode: ask`

3. **frontend-expert**
   - **Purpose**: React/Vue/Angular component design
   - **Requires**: User approval for UI implementation
   - **Frontmatter**: `permissionMode: ask`

4. **database-expert**
   - **Purpose**: Schema design and optimization
   - **Requires**: User approval for schema changes
   - **Frontmatter**: `permissionMode: ask`

5. **api-designer**
   - **Purpose**: REST/GraphQL API design
   - **Requires**: User approval for API contracts
   - **Frontmatter**: `permissionMode: ask`

#### Infrastructure & DevOps Agents

6. **devops-expert**
   - **Purpose**: Deployment and infrastructure
   - **Requires**: User approval for infra changes
   - **Frontmatter**: `permissionMode: ask`

7. **migration-expert**
   - **Purpose**: Database migrations and schema evolution
   - **Requires**: User approval for data changes
   - **Frontmatter**: `permissionMode: ask`

#### Architecture & Planning Agents

8. **implementation-planner**
   - **Purpose**: Strategy and approach planning
   - **Requires**: User approval for implementation strategy
   - **Frontmatter**: `permissionMode: ask`

9. **ui-ux-expert**
   - **Purpose**: UI/UX design and accessibility
   - **Requires**: User approval for design decisions
   - **Frontmatter**: `permissionMode: ask`

10. **component-designer**
    - **Purpose**: Component architecture and design systems
    - **Requires**: User approval for component design
    - **Frontmatter**: `permissionMode: ask`

#### Monitoring & Performance Agents

11. **performance-engineer**
    - **Purpose**: Performance optimization
    - **Requires**: User approval for optimization strategy
    - **Frontmatter**: `permissionMode: ask`

12. **monitoring-expert**
    - **Purpose**: Observability and alerting
    - **Requires**: User approval for monitoring setup
    - **Frontmatter**: `permissionMode: ask`

#### Security & Compliance Agents

13. **security-expert**
    - **Purpose**: Security analysis and hardening
    - **Requires**: User approval for security changes
    - **Frontmatter**: `permissionMode: ask`

14. **accessibility-expert**
    - **Purpose**: WCAG compliance and accessibility
    - **Requires**: User approval for a11y implementation
    - **Frontmatter**: `permissionMode: ask`

#### Integration & Specialized Agents

15. **mcp-context7-integrator**
    - **Purpose**: Context7 MCP integration
    - **Requires**: User approval for MCP setup
    - **Frontmatter**: `permissionMode: ask`

16. **mcp-playwright-integrator**
    - **Purpose**: Playwright MCP integration
    - **Requires**: User approval for test setup
    - **Frontmatter**: `permissionMode: ask`

17. **mcp-notion-integrator**
    - **Purpose**: Notion MCP integration
    - **Requires**: User approval for Notion sync
    - **Frontmatter**: `permissionMode: ask`

18. **figma-expert**
    - **Purpose**: Figma design-to-code conversion
    - **Requires**: User approval for code generation
    - **Frontmatter**: `permissionMode: ask`

#### Advanced Analysis Agents

19. **mcp-sequential-thinking-integrator**
    - **Purpose**: Multi-step reasoning with thinking mode
    - **Requires**: User approval for strategic thinking
    - **Frontmatter**: `permissionMode: ask`

20. **debug-helper**
    - **Purpose**: Runtime error diagnosis
    - **Requires**: User approval for fix strategy
    - **Frontmatter**: `permissionMode: ask`

21. **git-manager**
    - **Purpose**: Git operations and PR management
    - **Requires**: User approval for git changes
    - **Frontmatter**: `permissionMode: ask`

---

### Level 3: Integration Patterns (200+ lines)

**How Permission Modes Integrate with Workflow Commands**

#### /alfred:0-project Workflow

```
/alfred:0-project
    ↓
[Auto] project-manager
  └─ Task: Initialize project structure
  └─ Permission: Auto (setup operation)
  └─ Operations: Create config files, detect language
    ↓
[Auto] cc-manager
  └─ Task: Configure Claude Code settings
  └─ Permission: Auto (configuration-only)
  └─ Operations: Create settings.json, mcp.json
    ↓
✅ Project initialized automatically
   (No user approval needed for setup)
```

#### /alfred:1-plan Workflow

```
/alfred:1-plan "user authentication"
    ↓
[Auto] spec-builder
  └─ Task: Generate SPEC-AUTH-001
  └─ Permission: Auto (documentation)
  └─ Operations: Create SPEC file, validate EARS format
    ↓
✅ SPEC created automatically
   (No user approval needed for SPEC)
```

#### /alfred:2-run SPEC-001 Workflow

```
/alfred:2-run SPEC-001
    ↓
❓ [Ask Mode] implementation-planner
  └─ Task: Plan implementation approach
  └─ Permission: Ask (strategic decision)
  └─ User Input: "Approve implementation approach?"
    ↓
✅ User approves
    ↓
[Ask Mode] tdd-implementer
  └─ Task: Implement feature
  └─ Permission: Ask (code changes)
  └─ User Input: "Ready to implement?"
    ↓
✅ User approves
    ↓
Implementation proceeds (Red → Green → Refactor)
    ↓
[Auto] quality-gate
  └─ Task: Validate TRUST 5
  └─ Permission: Auto (validation-only)
  └─ Operations: Run tests, check coverage, lint code
    ↓
✅ Implementation + Validation complete
```

#### /alfred:3-sync Workflow

```
/alfred:3-sync auto SPEC-001
    ↓
[Auto] docs-manager
  └─ Task: Generate documentation
  └─ Permission: Auto (documentation)
  └─ Operations: Create API docs, README updates
    ↓
[Auto] doc-syncer
  └─ Task: Sync docs with code
  └─ Permission: Auto (sync operation)
  └─ Operations: Update diagrams, examples
    ↓
[Auto] quality-gate
  └─ Task: Validate documentation
  └─ Permission: Auto (validation)
  └─ Operations: Check links, validate markdown
    ↓
[Ask Mode] git-manager (if commit needed)
  └─ Task: Create commit + PR
  └─ Permission: Ask (git changes)
  └─ User Input: "Create commit and PR?"
    ↓
✅ Documentation synced + PR created
```

---

**Token Cost Analysis**

#### Auto Mode Cost

```
spec-builder → docs-manager → quality-gate
12,000 tokens   +  8,000 tokens   +  5,000 tokens  = 25,000 tokens
(No AskUserQuestion overhead)
```

#### Ask Mode Cost (with User Interaction)

```
implementation-planner → AskUserQuestion: 3,000 tokens (user input gathering)
    ↓
tdd-implementer → Implementation: 30,000 tokens
    ↓
AskUserQuestion: 2,000 tokens (confirmation)
    ↓
Total: 35,000 tokens
(AskUserQuestion adds 5,000 tokens overhead)
```

#### Optimization: When to Use Each Mode

| Scenario | Mode | Reasoning | Tokens Saved |
|---|---|---|---|
| Document generation | Auto | Safe, fast | -0 (baseline) |
| Code implementation | Ask | User control critical | +5,000 (user input cost) |
| Parallel docs + impl | Mix | Docs auto, code ask | +2,500 (reduced overhead) |
| Validation gates | Auto | Verification-only | -0 (no overhead) |

---

## 🎯 Real-World Examples

### Example 1: Full Feature Lifecycle

```yaml
# User: "Implement user authentication"

# Phase 1: SPEC Generation (Auto)
/alfred:1-plan "user authentication"
  → spec-builder (Auto) ✅ No approval needed
  → SPEC-AUTH-001 generated
  → Tokens: 15,000

# Phase 2: Implementation Planning (Ask)
/alfred:2-run SPEC-AUTH-001
  → implementation-planner (Ask) ❓ Requires approval
    └─ AskUserQuestion: "Approve implementation approach?"
    └─ User selects: "JWT + Refresh Tokens"
  → Tokens: 5,000 (approval) + 3,000 (planning) = 8,000

# Phase 3: Code Implementation (Ask)
  → tdd-implementer (Ask) ❓ Requires approval
    └─ AskUserQuestion: "Ready to implement?"
    └─ User: "Yes"
  → Implements Red → Green → Refactor
  → Tokens: 2,000 (approval) + 45,000 (implementation) = 47,000

# Phase 4: Validation (Auto)
  → quality-gate (Auto) ✅ No approval needed
  → Checks: Tests (95% coverage), Types, Security
  → Tokens: 8,000

# Phase 5: Documentation (Auto)
/alfred:3-sync auto SPEC-AUTH-001
  → docs-manager (Auto) ✅ No approval needed
  → doc-syncer (Auto) ✅ No approval needed
  → Tokens: 12,000

# Total Tokens: 90,000
# Without Ask Mode: Would need approval 5+ times!
# With Ask Mode: Strategic approval points only
```

### Example 2: Permission Mode Distribution

```
Project with 20 SPEC documents:

Auto-Mode Operations: 20 × 5 agents/SPEC × 10,000 tokens = 1,000,000 tokens
├─ spec-builder (generation)
├─ docs-manager (documentation)
├─ quality-gate (validation)
├─ format-expert (styling)
└─ doc-syncer (synchronization)

Ask-Mode Operations: 20 × 2 agents/SPEC × 15,000 tokens = 600,000 tokens
├─ implementation-planner (with 1 AskUserQuestion)
└─ tdd-implementer (with 1 AskUserQuestion)

Total: 1,600,000 tokens
→ Only 2 approval points per SPEC (vs 10+ without permission modes)
→ 60% reduction in user interaction overhead
```

---

## 🔧 Configuration & Setup

### Agent Frontmatter Configuration

Each agent YAML should include permissionMode:

```yaml
# .claude/agents/alfred/spec-builder.md
---
name: "spec-builder"
permissionMode: auto
skills:
  - moai-foundation-specs
---

# .claude/agents/alfred/tdd-implementer.md
---
name: "tdd-implementer"
permissionMode: ask
skills:
  - moai-lang-python
  - moai-essentials-debug
---
```

### config.json Agent Permissions

```json
{
  "agent_permissions": {
    "permission_modes": {
      "auto": {
        "description": "Auto-executed (no approval)",
        "agents": [
          "spec-builder", "docs-manager", "quality-gate",
          "project-manager", "agent-factory", "skill-factory"
        ]
      },
      "ask": {
        "description": "Requires user approval",
        "agents": [
          "tdd-implementer", "backend-expert", "frontend-expert",
          "database-expert", "security-expert"
        ]
      }
    }
  }
}
```

---

## 📊 Best Practices

### ✅ Do's

- ✅ Use Auto mode for documentation and validation
- ✅ Use Ask mode for code and infrastructure changes
- ✅ Place Ask mode agents at critical decision points
- ✅ Combine Auto + Ask for efficient workflows
- ✅ Monitor token usage for Ask mode overhead
- ✅ Use AskUserQuestion for strategic approvals

### ❌ Don'ts

- ❌ Use Ask mode for every operation (token waste)
- ❌ Use Auto mode for critical infrastructure changes
- ❌ Skip user approval for security operations
- ❌ Chain too many Ask-mode agents (approval fatigue)
- ❌ Assume all agents should have same permission level

---

## 🔗 Related Skills

- **moai-cc-subagent-lifecycle** - Agent lifecycle management
- **moai-cc-hook-model-strategy** - Hook model optimization
- **moai-cc-hooks** - General hook architecture
- **moai-alfred-workflow** - Workflow orchestration patterns

---

**Last Updated**: 2025-11-18
**Version**: 2.0.43
**Status**: Production Ready
