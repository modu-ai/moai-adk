# MoAI-ADK Core Rules & Architecture Covenant

**The immutable principles and architectural foundations that govern the MoAI-ADK system.**

> **Purpose**: This document defines the unchangeable rules and core architectural principles of MoAI-ADK. These rules take precedence over all other documentation and implementation details.

---

## 🏛️ Foundational Principles (Immutable)

### SPEC-First Development Covenant

**Rule 1**: Every feature MUST have a valid EARS-format SPEC document before implementation.

**EARS Format Mandate**:
- **E**valuate: Current state analysis and problem definition
- **A**nalyze: Requirements breakdown and constraints identification
- **R**ecommend: Solution architecture and implementation approach
- **S**ynthesize: Integrated solution with quality gates

**Implementation**: SPEC documents MUST be created using `/moai:1-plan` command ONLY. Direct file creation is strictly prohibited.

### TDD Enforcement Rule

**Rule 2**: All implementation MUST follow Red-Green-Refactor cycle.

**TDD Sequence**:
1. **RED**: Write failing tests first
2. **GREEN**: Implement minimum code to pass tests
3. **REFACTOR**: Optimize code quality and structure

**Quality Gates**: TRUST 5 compliance (Test-first, Readable, Unified, Secured, Trackable) MUST be achieved before proceeding.

### Agent Delegation Mandate

**Rule 3**: Direct tool usage is PROHIBITED. All work MUST be delegated through Task().

**Allowed Tools Only**:
- `Task()` - Agent delegation (PRIMARY)
- `AskUserQuestion()` - User interaction
- `Skill()` - Knowledge invocation
- MCP server tools - context7, github, filesystem

**Forbidden Tools**:
- `Read()`, `Write()`, `Edit()` → Must use Task() delegation
- `Bash()` → Must use Task() delegation
- `Grep()`, `Glob()` → Must use Task() delegation

---

## 🏗️ Architectural Covenant

### 4-Layer Architecture (Non-negotiable)

```
┌─────────────────────────────────────────────────────────────┐
│ COMMANDS (Workflow Orchestration)                           │
│ 6 core slash commands: /moai:0-project → /moai:99-release     │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ AGENTS (Domain Expertise)                                    │
│ 35 specialized agents with clear responsibilities             │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ SKILLS (Knowledge Library)                                   │
│ 135 reusable skills with Context7 integration                │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ HOOKS (Guardrails & Context)                                 │
│ 6 automated hooks for quality and security enforcement        │
└─────────────────────────────────────────────────────────────┘
```

### Agent Priority Stack (Strict Order)

**Priority 1: MoAI-ADK Agents (35 total)**
- MUST be used first for all domain-specific tasks
- SPEC-aware and production-ready
- Examples: `spec-builder`, `tdd-implementer`, `backend-expert`, `security-expert`

**Priority 2: MoAI-ADK Skills (135 total)**
- Reusable knowledge patterns
- Context7 integration for latest APIs
- Examples: `moai-lang-python`, `moai-domain-security`, `moai-foundation-trust`

**Priority 3: Claude Code Native Agents**
- Fallback only when Priority 1-2 unavailable
- Limited to `Explore`, `Plan`, `debug-helper`

---

## 🔄 Workflow Covenant

### Standard Development Workflow (Mandatory)

```
1. INITIALIZATION
   /moai:0-project → Project setup and detection

2. SPECIFICATION
   /moai:1-plan "feature description" → EARS SPEC creation
   ↳ MANDATORY: /clear (saves 45-50K tokens)

3. IMPLEMENTATION
   /moai:2-run SPEC-XXX → TDD implementation
   ↳ OPTIONAL: /clear if context > 150K tokens

4. SYNCHRONIZATION
   /moai:3-sync auto SPEC-XXX → Documentation and quality gates
```

### Token Management Rules

**Phase Budgets (Strict)**:
- SPEC Creation: 30K tokens maximum
- TDD Implementation: 180K tokens maximum (includes phases)
- Documentation Sync: 40K tokens maximum

**Required Clear Points**:
- **MANDATORY**: After SPEC creation (saves 45-50K tokens)
- **RECOMMENDED**: After implementation if context > 150K
- **BEST PRACTICE**: Every 50+ messages

### File Organization Covenant

**Directory Structure (Required)**:
```
.moai/
├── specs/          # SPEC documents (from /moai:1-plan ONLY)
├── docs/           # Generated documentation (from docs-manager)
├── reports/        # Analysis and completion reports
├── memory/         # Reference documentation (this file)
└── logs/           # Execution logs and transcripts
```

**Prohibited Patterns**:
- ❌ Documentation in project root
- ❌ Direct file creation with Task()
- ❌ SPEC documents outside .moai/specs/
- ❌ Mixed language in infrastructure files

---

## 🛡️ Security & Quality Covenant

### Sandbox Mode (Always Active)

**Security Rules**:
- Sandbox mode MUST always be enabled
- Dangerous commands blocked: `rm -rf`, `sudo`, `chmod 777`
- Protected files: `.env*`, `.vercel/`, `.aws/`, `.github/workflows/secrets`

**Permission Principles**:
- Principle of Least Privilege applied
- Tool access strictly controlled
- File access validated through hooks

### TRUST 5 Quality Gates (Non-negotiable)

**1. Test-first**: All code starts with tests
**2. Readable**: Clear naming and documentation
**3. Unified**: Consistent patterns and styles
**4. Secured**: Security validation passed
**5. Trackable**: Change history maintained

**Quality Metrics**:
- Test coverage: 85% minimum
- Code review: Security-expert validation required
- Documentation: Auto-sync with implementation
- Security: OWASP compliance mandatory

---

## 🔗 Integration Covenant

### MCP Server Integration Rules

**Required MCP Servers**:
- **Context7**: Documentation and library resolution
- **GitHub**: Issue and PR operations
- **Filesystem**: File navigation and search

**Integration Patterns**:
- MCP tools auto-available when configured
- Library resolution through `mcp__context7__resolve-library-id()`
- Documentation access via `mcp__context7__get-library-docs()`

### Context7 Usage Protocol

**Library Resolution**:
```python
# Pattern: Resolve → Get docs → Apply
library_id = await mcp__context7__resolve-library-id("React")
docs = await mcp__context7__get-library-docs(library_id)
# Then delegate to appropriate agent
```

**Documentation Integration**:
- Skills automatically integrate Context7 for latest APIs
- Agents receive up-to-date library documentation
- Version-specific documentation resolution supported

---

## 📋 Execution Covenant

### Model Selection Rules

**Sonnet 4.5 (High-Cost, High-Performance)**:
- SPEC creation and architecture decisions
- Security reviews and complex reasoning
- Multi-domain orchestration tasks

**Haiku 4.5 (70% Cost Savings)**:
- Code exploration and simple fixes
- Test execution and validation
- File search and navigation

### Agent Orchestration Rules

**Sequential Execution**:
- Output from previous agent becomes input to next
- Context preservation across agent chain
- Error handling and recovery patterns

**Parallel Execution**:
- Independent tasks run simultaneously
- Results aggregation and synthesis
- Resource optimization for speed

**Conditional Branching**:
- Decision-based workflow routing
- Complexity-based agent selection
- Dynamic task decomposition

---

## 🎯 Covenant Enforcement

### Hook System Validation

**6 Automated Hooks**:
1. **SessionStart**: Load project metadata and validate setup
2. **UserPromptSubmit**: Analyze complexity and route to agents
3. **SubagentStart**: Seed context and set constraints
4. **SubagentStop**: Validate output and handle errors
5. **PreToolUse**: Security validation and command checking
6. **SessionEnd**: Save metrics and perform cleanup

**Hook Failure Protocol**:
- Errors logged to `.moai/logs/`
- Graceful degradation with functionality preservation
- User notification for critical issues

### Compliance Validation

**Automated Checks**:
- SPEC document format validation
- TDD cycle compliance verification
- Agent delegation pattern enforcement
- Token budget monitoring
- Security rule validation

**Violation Handling**:
- Warning system for minor violations
- Block execution for critical violations
- Suggest corrective actions
- Log all violations for analysis

---

## 📜 Covenant Amendment Process

**Rule Change Requirements**:
- Core principles require 100% consensus
- Architectural changes need backward compatibility
- Workflow modifications must maintain efficiency
- Security rules require threat model review

**Version Control**:
- All changes tracked with semantic versioning
- Backward compatibility maintained for minor versions
- Major versions may include breaking changes
- Migration guides provided for all updates

---

**This covenant represents the fundamental laws of MoAI-ADK. Violation of these rules compromises system integrity, security, and efficiency. All agents, skills, and workflows must adhere to these principles without exception.**

---

*Document Status: Covenant v1.0 - Immutable Foundations*
*Last Updated: 2025-11-20*
*Authority: MoAI-ADK Architecture Board*
*Review Cycle: Annual (for additions only, never removals)*