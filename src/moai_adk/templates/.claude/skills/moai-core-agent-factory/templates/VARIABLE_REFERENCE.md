# Agent Template Variable Reference

## Overview

This document defines all variables used in agent templates and provides guidance for substitution logic.

## Template Variable Categories

### 1. Agent Identity Variables

| Variable | Description | Example | Template Level |
|----------|-------------|---------|-----------------|
| `{{AGENT_NAME}}` | kebab-case agent identifier | `backend-expert`, `security-auditor` | All |
| `{{AGENT_TITLE}}` | Formal agent title | `Backend Expert - API & Database Specialist` | All |
| `{{AGENT_SUBTITLE}}` | One-line specialization | `Senior Backend Architect` | All |
| `{{AGENT_ICON}}` | Emoji representing agent | `🔧`, `🔐`, `📚` | All |
| `{{AGENT_JOB}}` | Professional job title | `Senior Backend Architect` | All |
| `{{AGENT_TAGLINE}}` | One-sentence agent purpose | `Comprehensive backend architecture and API design specialist` | All |

### 2. Trigger & Context Variables

| Variable | Description | Example | Template Level |
|----------|-------------|---------|-----------------|
| `{{PROACTIVE_TRIGGERS}}` | When agent should be auto-invoked | `Backend architecture, API design, server implementation` | All |
| `{{COMMAND_CONTEXT}}` | Command or workflow calling agent | `/alfred:2-run SPEC-BACKEND-001` | All |
| `{{TASK_DESCRIPTION}}` | Task summary for Task() invocation | `Design REST API for user authentication` | Standard, Complex |
| `{{CORE_MISSION}}` | Primary agent mission statement | `Design scalable, secure backend architectures` | Standard, Complex |
| `{{WRONG_USAGE}}` | Example of incorrect usage | `"Design a backend API"` | Standard, Complex |

### 3. Expertise & Role Variables

| Variable | Description | Example | Template Level |
|----------|-------------|---------|-----------------|
| `{{EXPERTISE_AREA}}` | Domain of specialization | `REST/GraphQL API design, microservices, database modeling` | All |
| `{{AGENT_ROLE}}` | Agent's role in team | `Architect who translates requirements into scalable implementations` | All |
| `{{AGENT_GOAL}}` | What agent aims to deliver | `Production-ready backend with 85%+ test coverage` | All |
| `{{DOMAIN}}` | Primary domain | `backend`, `frontend`, `security`, `testing` | All |

### 4. Tool & Permission Variables

| Variable | Description | Example | Template Level |
|----------|-------------|---------|-----------------|
| `{{TOOLS_LIST}}` | Comma-separated tool permissions | `Read, Write, Edit, Bash, WebFetch` | All |
| `{{TOOL_COUNT}}` | Number of tool permissions | `11` | Complex |
| `{{JUSTIFICATION_PRINCIPLE}}` | Why this tool set | `Least privilege + research needs` | Complex |

### 5. Skill Variables

| Variable | Description | Example | Template Level |
|----------|-------------|---------|-----------------|
| `{{AUTO_SKILLS}}` | Bullet list of auto-loaded skills | `- Skill("moai-domain-backend")\n- Skill("moai-lang-python")` | All |
| `{{CONDITIONAL_SKILLS}}` | Conditional skill loading logic | `- Language-specific: Based on project\n- Conditional: Load on-demand` | All |
| `{{SKILL_COUNT}}` | Total number of skills | `8 (2 auto + 6 conditional)` | Complex |
| `{{SKILL_LOADING_STRATEGY}}` | How skills are loaded | `Core immediate, conditional on-demand` | Complex |

### 6. Responsibility Variables

| Variable | Description | Example | Template Level |
|----------|-------------|---------|-----------------|
| `{{RESPONSIBILITIES_DO}}` | Bulleted list of DO items | `- Design REST/GraphQL APIs\n- Recommend SQL/NoSQL solutions` | All |
| `{{RESPONSIBILITIES_DONT}}` | Bulleted list of DO NOT items with delegation | `- Implement code (→ tdd-implementer)\n- Security review (→ security-expert)` | All |

### 7. Workflow Variables

| Variable | Description | Example | Template Level |
|----------|-------------|---------|-----------------|
| `{{WORKFLOW_SUMMARY}}` | High-level workflow description | `Analyze requirements → Design architecture → Recommend implementation` | All |
| `{{WORKFLOW_STEPS}}` | Detailed step-by-step guide | `### Step 1: Requirement Analysis\n...` | All |
| `{{PIPELINE_DIAGRAM}}` | ASCII art pipeline visualization | `User Input → Analysis → Design → Output` | Standard, Complex |

### 8. Collaboration Variables

| Variable | Description | Example | Template Level |
|----------|-------------|---------|-----------------|
| `{{COLLABORATION_PATTERNS}}` | How agent works with others | `Consults security-expert for auth patterns` | Standard |
| `{{COLLABORATION_PATTERNS}}` | Multi-agent orchestration map | `graph showing agent relationships` | Complex |
| `{{DELEGATION_TARGETS}}` | List of agents to delegate to | `security-expert, database-expert, quality-gate` | All |
| `{{DELEGATION_TRIGGER}}` | When to delegate | `Security concerns arise` | Standard |
| `{{DELEGATION_TARGET}}` | Which agent to delegate to | `security-expert` | Standard |
| `{{DELEGATION_EXAMPLE_SITUATION}}` | Example delegation scenario | `Designing authentication system` | Standard |
| `{{DELEGATION_RESULT}}` | Expected outcome | `Secure auth patterns returned for integration` | Standard |
| `{{DELEGATION_LOGIC}}` | Complex delegation decision tree | `IF security_critical → delegate to security-expert` | Complex |
| `{{DEPENDENT_AGENTS}}` | Agents this one depends on | `backend-expert, database-expert, security-expert` | Complex |
| `{{ORCHESTRATION_SEQUENCE}}` | Sequence for multi-agent work | `1. backend-expert → 2. security-expert → 3. quality-gate` | Complex |

### 9. Research & MCP Variables

| Variable | Description | Example | Template Level |
|----------|-------------|---------|-----------------|
| `{{RESEARCH_CAPABILITY}}` | What research this agent does | `Extensive documentation research via Context7 MCP` | Standard |
| `{{RESEARCH_PATTERNS}}` | Research methodology details | `Query patterns, documentation quality checks` | Standard |
| `{{RESEARCH_APPROACH}}` | High-level research strategy | `Evidence-based recommendations from official documentation` | Complex |
| `{{RESEARCH_LIBRARIES}}` | Key libraries/frameworks researched | `FastAPI, Express, NestJS, Spring Boot` | Complex |
| `{{RESEARCH_BEST_PRACTICES}}` | Extracted best practices | `- Use async/await for I/O operations\n- Implement rate limiting` | Complex |
| `{{MCP_LIBRARIES}}` | Context7 MCP research targets | `/tiangolo/fastapi, /expressjs/express` | Complex |
| `{{MCP_PATTERN}}` | MCP integration pattern | `Resolve library → Fetch docs → Extract patterns` | Complex |
| `{{MCP_QUALITY_THRESHOLD}}` | Minimum quality score | `≥70% documentation coverage` | Complex |
| `{{MCP_TOOLS}}` | Array of MCP tools used | `["context7"]` | Complex |
| `{{FALLBACK_STRATEGY}}` | What to do if MCP fails | `Use WebFetch for framework docs + established patterns` | Complex |

### 10. Model & Performance Variables

| Variable | Description | Example | Template Level |
|----------|-------------|---------|-----------------|
| `{{MODEL_SELECTION}}` | Claude model assignment | `sonnet`, `haiku`, `inherit` | All |
| `{{MODEL_JUSTIFICATION}}` | Why this model selected | `Complex architecture design requires deep reasoning` | Standard, Complex |
| `{{EXECUTION_TIME}}` | Estimated execution duration | `5-15 minutes` | All |
| `{{EXECUTION_TIME_SECONDS}}` | Execution time in seconds | `600` (10 minutes) | Complex |
| `{{CONTEXT_USAGE}}` | Context window usage | `~8000 tokens`, `Medium` | All |
| `{{CONTEXT_HEAVY}}` | Is agent context-heavy? | `true`, `false` | Complex |
| `{{MCP_INTEGRATION}}` | MCP servers used | `Context7 for documentation research` | Standard, Complex |

### 11. Orchestration Variables (Complex Only)

| Variable | Description | Example |
|----------|-------------|---------|
| `{{CAN_RESUME}}` | Can agent resume iteratively? | `true` |
| `{{CHAIN_POSITION}}` | Position in workflow | `initial`, `middle`, `final` |
| `{{DEPENDENCIES}}` | Agent dependencies | `spec-builder, backend-expert` |
| `{{RESUME_PATTERN}}` | Resume pattern | `single-session`, `multi-day` |
| `{{PARALLEL_SAFE}}` | Can run in parallel? | `true`, `false` |
| `{{REQUIRES_APPROVAL}}` | Needs user approval? | `true`, `false` |

### 12. Feature & Advanced Variables (Complex Only)

| Variable | Description | Example |
|----------|-------------|---------|
| `{{FEATURE_1_NAME}}` | Feature 1 name | `Framework-Agnostic Design` |
| `{{FEATURE_1_DESCRIPTION}}` | Feature 1 description | `Supports 13+ backend frameworks` |
| `{{FEATURE_2_NAME}}` | Feature 2 name | `Optimization Recommendations` |
| `{{FEATURE_2_DESCRIPTION}}` | Feature 2 description | `Performance-focused suggestions` |
| `{{FEATURE_3_NAME}}` | Feature 3 name | `MCP Fallback Strategy` |
| `{{FEATURE_3_DESCRIPTION}}` | Feature 3 description | `Works without MCP servers` |

### 13. Quality & Standards Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `{{DESIGN_PRINCIPLES}}` | Core design principles | `- Framework-agnostic\n- Security-first\n- Testability-focused` |
| `{{TRUST_5_COMPLIANCE}}` | TRUST 5 compliance notes | `Test-first ✓, Readable ✓, Unified ✓, Secured ✓, Trackable ✓` |
| `{{QUALITY_ASSURANCE}}` | QA practices | `- Automated testing\n- Code review patterns\n- Performance benchmarks` |
| `{{SUCCESS_CRITERIA}}` | Measurable success metrics | `- Passes all test cases\n- <15 min execution\n- 90% accuracy` |
| `{{OPTIMIZATION_NOTES}}` | Performance optimization notes | `Context loading: 40% reduction with conditional skills` |
| `{{KEY_PRINCIPLES}}` | Guiding principles | `- Clarity over brevity\n- Security-first approach` |

### 14. Integration Variables (Complex Only)

| Variable | Description | Example |
|----------|-------------|---------|
| `{{PHASE_1_INTEGRATION}}` | Phase 1 integration | `Generate backend-specific SPECs` |
| `{{PHASE_2_INTEGRATION}}` | Phase 2 integration | `Design implementation from SPEC` |
| `{{PHASE_3_INTEGRATION}}` | Phase 3 integration | `Document generated architecture` |
| `{{AGENT_ECOSYSTEM}}` | Agent ecosystem integration | `Works with spec-builder, quality-gate, security-expert` |
| `{{SKILL_ECOSYSTEM}}` | Skill ecosystem integration | `Leverages all 128+ MoAI-ADK skills` |
| `{{COMMAND_ECOSYSTEM}}` | Command ecosystem integration | `Compatible with all Alfred commands` |

### 15. Example & Usage Variables (Complex Only)

| Variable | Description | Example |
|----------|-------------|---------|
| `{{EXAMPLE_1_TITLE}}` | Example 1 title | `REST API with Authentication` |
| `{{EXAMPLE_1_INPUT}}` | Example 1 input | `Design login endpoint with JWT` |
| `{{EXAMPLE_1_OUTPUT}}` | Example 1 output | `Complete API specification` |
| `{{EXAMPLE_2_TITLE}}` | Example 2 title | `GraphQL Optimization` |
| `{{EXAMPLE_2_INPUT}}` | Example 2 input | `Optimize slow GraphQL query` |
| `{{EXAMPLE_2_OUTPUT}}` | Example 2 output | `Optimized schema with indexes` |
| `{{EXAMPLE_3_TITLE}}` | Example 3 title | `Microservices Architecture` |
| `{{EXAMPLE_3_INPUT}}` | Example 3 input | `Design scalable microservices` |
| `{{EXAMPLE_3_OUTPUT}}` | Example 3 output | `Complete microservices blueprint` |

### 16. Knowledge & Planning Variables (Complex Only)

| Variable | Description | Example |
|----------|-------------|---------|
| `{{KNOWLEDGE_BASE_REFERENCES}}` | Relevant documentation links | `- Official Framework Docs\n- API Design Patterns` |
| `{{LEARNING_RESOURCES}}` | Educational resources | `- REST API Best Practices\n- GraphQL Fundamentals` |
| `{{PLANNED_FEATURES}}` | Future planned features | `- GraphQL subscriptions\n- WebSocket support` |
| `{{RESEARCH_OPPORTUNITIES}}` | Research areas | `- AI-powered optimization\n- Multi-language support` |

### 17. Metadata Variables

| Variable | Description | Example | Template Level |
|----------|-------------|---------|-----------------|
| `{{CREATION_DATE}}` | Generation date | `2025-11-15` | All |

---

## Variable Substitution Rules

### 1. Formatting Rules

**Bullet Lists**:
```
{{RESPONSIBILITIES_DO}}
→ - Responsibility 1
  - Responsibility 2
  - Responsibility 3
```

**Code Blocks**:
```
{{WORKFLOW_STEPS}}
→ ### Step 1: Description
   **Actions**:
   1. Action one
   2. Action two
```

**Arrays/Lists**:
```
{{TOOLS_LIST}}
→ Read, Write, Edit, Bash, WebFetch

{{DELEGATION_TARGETS}}
→ ["security-expert", "database-expert"]
```

### 2. Conditional Substitution

**Only substitute if value exists**:
```
{{OPTIONAL_VARIABLE}}
→ If empty, remove entire section
→ If present, include with full context
```

**Multi-value lists**:
```
{{DEPENDENCIES}}
→ If none: []
→ If one: ["agent-name"]
→ If multiple: ["agent1", "agent2", "agent3"]
```

### 3. Markdown Formatting

**Bold emphasis**:
```
**ALWAYS USE BOLD** for critical terms:
- Agent names: **{{AGENT_NAME}}**
- Keywords: **Claude Code**, **Task()**
```

**Code inline**:
```
Inline code for technical terms:
- Agent names in code context: `{{AGENT_NAME}}`
- File paths: `.claude/agents/alfred/`
- Function calls: `Task(subagent_type="...")`
```

---

## Generation Workflow

### Step 1: Collect Requirements
- User provides agent description
- Extract domain, capabilities, complexity
- Detect research needs

### Step 2: Determine Template Tier
- Complexity score → Template selection
- LOW → Tier 1 (Simple)
- MEDIUM → Tier 2 (Standard)
- HIGH → Tier 3 (Complex)

### Step 3: Calculate Variables
- Generate all variable values
- Validate against requirements
- Check for missing critical variables

### Step 4: Substitute Into Template
- Replace each {{VARIABLE}} with value
- Preserve formatting and structure
- Validate result for completeness

### Step 5: Validate Output
- Syntax check (YAML, Markdown)
- Structure verification
- Content review
- Quality gate approval

---

## Example: Variable Substitution

**Input Request**:
```
"Create an agent that designs REST APIs with performance optimization"
```

**Extracted Variables**:
```
AGENT_NAME = api-designer
DOMAIN = backend
COMPLEXITY = MEDIUM → TEMPLATE = Standard (Tier 2)
MODEL_SELECTION = inherit
TOOLS_LIST = Read, Write, Edit, WebFetch, AskUserQuestion
AUTO_SKILLS = Skill("moai-domain-backend")
RESPONSIBILITIES_DO = - Design REST API specifications with proper error handling
                      - Recommend pagination and caching strategies
```

**Substituted Output**:
```markdown
---
name: api-designer
description: "Use PROACTIVELY when: REST API design, API contract definition..."
tools: Read, Write, Edit, WebFetch, AskUserQuestion
model: inherit
---

# API Designer - REST API Specialist

...

## 🧰 Required Skills

**Automatic Core Skills**:
- `Skill("moai-domain-backend")` – REST API design patterns

...

## 🎯 Core Responsibilities

✅ **DOES**:
- Design REST API specifications with proper error handling
- Recommend pagination and caching strategies

...
```

---

## Variable Validation

### Critical Variables (Must Not Be Empty)

- {{AGENT_NAME}}
- {{DOMAIN}}
- {{AGENT_ICON}}
- {{TOOLS_LIST}}
- {{MODEL_SELECTION}}
- {{AUTO_SKILLS}}
- {{RESPONSIBILITIES_DO}}
- {{RESPONSIBILITIES_DONT}}

### Optional Variables (Can Be Empty)

- {{MCP_INTEGRATION}} (defaults to none)
- {{FEATURE_1_NAME}} (only for Complex)
- {{PLANNED_FEATURES}} (future-looking)

### Dependent Variables (Requires Validation)

- {{RESPONSIBILITIES_DONT}}: Each must have {{DELEGATION_TARGET}}
- {{DELEGATION_TARGETS}}: Each agent must exist in MoAI-ADK
- {{SKILL_LIST}}: Each skill must exist in repository

---

## Quality Checklist

- [ ] All critical variables substituted
- [ ] No untouched {{VARIABLE}} strings remain
- [ ] Bullet lists properly formatted
- [ ] Code blocks properly indented
- [ ] YAML frontmatter valid
- [ ] No duplicate information
- [ ] Links and references valid
- [ ] Markdown heading hierarchy correct
- [ ] All agent names in kebab-case
- [ ] All Skill() calls valid

---

**Version**: 1.0.0
**Last Updated**: 2025-11-15
