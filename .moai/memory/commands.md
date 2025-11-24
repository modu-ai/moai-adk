# MoAI-ADK Commands Reference

Six core MoAI-ADK commands used by Alfred. Essential tools for SPEC-First TDD execution.

## `/moai:0-project` - Project Initialization

**Purpose**: Initialize project structure and generate configuration

**Delegation**: `project-manager`

**Usage**:

```
/moai:0-project
/moai:0-project --with-git
```

**Output**: `.moai/` directory + config.json

**Next Step**: Ready for SPEC generation

---

## `/moai:1-plan` - SPEC Generation

**Purpose**: Generate SPEC document in EARS format

**Delegation**: `spec-builder`

**Usage**:

```
/moai:1-plan "Implement user authentication endpoint (JWT)"
```

**Output**: `.moai/specs/SPEC-001/spec.md` (EARS format document)

**Required**: Execute `/clear` after completion (saves 45-50K tokens)

---

## `/moai:2-run` - TDD Implementation

**Purpose**: Execute RED-GREEN-REFACTOR cycle

**Delegation**: `tdd-implementer`

**Usage**:

```
/moai:2-run SPEC-001
```

**Process**:

1. RED: Write failing tests
2. GREEN: Pass with minimal code
3. REFACTOR: Optimize and clean up

**Output**: Implemented code + tests + quality report

**Requirement**: Test coverage ≥ 85% (TRUST 5)

---

## `/moai:3-sync` - Documentation Synchronization

**Purpose**: Auto-generate API documentation and project artifacts

**Delegation**: `docs-manager`

**Usage**:

```
/moai:3-sync SPEC-001
```

**Output**:

- API documentation (OpenAPI format)
- Architecture diagrams
- Project report

---

## `/moai:9-feedback` - Improvement Feedback Collection

**Purpose**: Error analysis and improvement suggestions

**Delegation**: `quality-gate`

**Usage**:

```
/moai:9-feedback
/moai:9-feedback --analyze SPEC-001
```

**Purpose**: Continuous MoAI-ADK improvement, error recovery

---

## Required Workflow

```
1. /moai:0-project              # Initialize project
2. /moai:1-plan "description"   # Generate SPEC
3. /clear                       # Initialize context (required)
4. /moai:2-run SPEC-001         # TDD implementation
5. /moai:3-sync SPEC-001        # Generate documentation
6. /moai:9-feedback             # Collect feedback
7. /moai:99-release             # Production deployment
```

---

## Context Initialization Rules

- Execute `/clear` **after** `/moai:1-plan` (mandatory)
- Execute `/clear` when context > 150K tokens
- Execute `/clear` after 50+ conversation messages

Refer to CLAUDE.md for detailed command usage information.
