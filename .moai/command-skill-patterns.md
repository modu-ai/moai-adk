# Command Skill Patterns

**Phase 2 Documentation Standardization - Verified Command Patterns Only**

## Overview

This document documents the actual command skill patterns found in the MoAI-ADK codebase, extracted exclusively from verified command implementations. All patterns documented here are based on real command references and skill invocations.

## Command Pattern Analysis

### Important: Removed Non-Existent JIT Skills

The following skills are mentioned in command documentation but **do not exist** in the codebase and have been removed from this analysis:

- ❌ `moai-session-info` - Not found in `.claude/skills/`
- ❌ `moai-jit-docs-enhanced` - Not found in `.claude/skills/`
- ❌ `moai-streaming-ui` - Not found in `.claude/skills/`
- ❌ `moai-change-logger` - Not found in `.claude/skills/`
- ❌ `moai-tag-policy-validator` - Not found in `.claude/skills/`
- ❌ `moai-learning-optimizer` - Not found in `.claude/skills/`

These were hallucinated or planned features that don't exist in the actual implementation.

## Verified Command Skill Patterns

### 1. Interactive Question Pattern

#### Pattern Description
All commands that use `AskUserQuestion` must first invoke the question skill.

```python
def load_question_skills():
    """Always invoke before using AskUserQuestion tool"""

    # Required before any user interaction
    Skill("moai-alfred-ask-user-questions")()
```

#### Command Usage
From `/alfred:1-plan.md`:
```markdown
> **Critical Note**: ALWAYS invoke `Skill("moai-alfred-ask-user-questions")` before using `AskUserQuestion` tool. This skill provides up-to-date best practices, field specifications, and validation rules for interactive prompts.
```

#### Implementation Pattern
- **Trigger**: Any command using `AskUserQuestion` tool
- **Required Skill**: `moai-alfred-ask-user-questions`
- **Purpose**: Provides TUI-based interaction guidance
- **Timing**: Must be called BEFORE using `AskUserQuestion`

### 2. SPEC Creation Reference Pattern

#### Pattern Description
Commands reference skills for SPEC creation guidance without direct invocation.

```python
def spec_creation_references():
    """Skills referenced for SPEC creation guidance"""

    # Reference skills (not directly called by commands)
    references = [
        "moai-foundation-ears",           # EARS syntax guidance
        "moai-foundation-specs",          # SPEC metadata validation
        "moai-alfred-spec-metadata-validation",  # Metadata validation
        "moai-foundation-tags",          # TAG chain traceability
    ]

    return references  # Used for documentation/guidance only
```

#### Command Usage
From `/alfred:1-plan.md`:
```markdown
스킬 호출:
필요 시 명시적 Skill() 호출 사용:
- Skill("moai-foundation-specs") - SPEC 구조 가이드
- Skill("moai-foundation-ears") - EARS 문법 요구사항
- Skill("moai-alfred-spec-metadata-validation") - 메타데이터 검증
- Skill("moai-foundation-tags") - TAG 체인 참조
```

#### Implementation Pattern
- **Trigger**: SPEC creation phases in commands
- **Reference Skills**: EARS, SPEC validation, TAG management
- **Purpose**: Provide guidance to sub-agents
- **Usage**: Referenced in command documentation, not called directly

### 3. Agent Skill Reference Pattern

#### Pattern Description
Commands reference agent selection and workflow skills for guidance.

```python
def agent_workflow_references():
    """Skills referenced for agent workflow guidance"""

    # Reference skills (not directly called by commands)
    references = [
        "moai-alfred-agent-guide",        # Agent selection guidance
        "moai-alfred-workflow",          # 4-step workflow
        "moai-alfred-rules",             # Core rules and standards
        "moai-alfred-dev-guide",         # Development guidance
    ]

    return references  # Used for documentation/guidance only
```

#### Command Usage
From `/alfred:1-plan.md`:
```markdown
For complete EARS syntax and examples, invoke: `Skill("moai-foundation-ears")`

For complete metadata field descriptions, validation rules, and version system guide, invoke: `Skill("moai-foundation-specs")`

For complete context engineering strategy, invoke: `Skill("moai-alfred-dev-guide")`
```

#### Implementation Pattern
- **Trigger**: Command execution phases
- **Reference Skills**: Agent guidance, workflow, development patterns
- **Purpose**: Provide operational guidance to command executors
- **Usage**: Referenced in command documentation for detailed information

## Command Skill Loading Patterns

### Pattern 1: Pre-Execution Skill Loading
```python
def pre_execution_skills():
    """Skills loaded before command execution"""

    # Only one verified pattern exists
    if user_interaction_expected:
        Skill("moai-alfred-ask-user-questions")()
```

### Pattern 2: Documentation Reference Skills
```python
def documentation_references():
    """Skills referenced in command documentation"""

    # SPEC creation references
    spec_references = [
        "moai-foundation-ears",
        "moai-foundation-specs",
        "moai-alfred-spec-metadata-validation",
        "moai-foundation-tags"
    ]

    # Workflow guidance references
    workflow_references = [
        "moai-alfred-workflow",
        "moai-alfred-agent-guide",
        "moai-alfred-rules",
        "moai-alfred-dev-guide"
    ]

    return {
        "spec": spec_references,
        "workflow": workflow_references
    }
```

### Pattern 3: Agent Delegation References
```python
def agent_delegation_references():
    """Skills referenced when delegating to agents"""

    # Commands don't call skills directly
    # They delegate to agents which then call skills

    agent_skill_mappings = {
        "spec-builder": [
            "moai-foundation-ears",
            "moai-foundation-specs",
            "moai-alfred-spec-metadata-validation",
            "moai-foundation-tags"
        ],
        "git-manager": [
            # No specific skill references found in git commands
            # Git operations handled directly by agent
        ]
    }

    return agent_skill_mappings
```

## Verified Command-to-Skill Relationships

### /alfred:1-plan Command
```python
def alfred_1_plan_skills():
    """Verified skill patterns for /alfred:1-plan"""

    # Direct skill invocation (removed non-existent JIT skills)
    # Original: Skill("moai-session-info") - REMOVED (doesn't exist)
    # Original: Skill("moai-jit-docs-enhanced") - REMOVED (doesn't exist)

    # Actual pattern: No direct skill invocation
    # Command delegates to agents which handle skill loading

    # Reference skills (for guidance only)
    reference_skills = [
        "moai-foundation-ears",
        "moai-foundation-specs",
        "moai-alfred-spec-metadata-validation",
        "moai-foundation-tags",
        "moai-alfred-dev-guide"
    ]

    # Interactive skill (required for user questions)
    interaction_skills = [
        "moai-alfred-ask-user-questions"
    ]

    return {
        "direct_calls": [],  # No direct skill calls
        "references": reference_skills,
        "interaction": interaction_skills
    }
```

### /alfred:3-sync Command
```python
def alfred_3_sync_skills():
    """Verified skill patterns for /alfred:3-sync"""

    # Based on command documentation analysis
    # Similar pattern to /alfred:1-plan

    # Reference skills (for guidance only)
    reference_skills = [
        "moai-foundation-tags",      # TAG validation
        "moai-foundation-trust",     # TRUST validation
        "moai-foundation-specs"      # SPEC validation
    ]

    # Interactive skill (required for user questions)
    interaction_skills = [
        "moai-alfred-ask-user-questions"
    ]

    return {
        "direct_calls": [],  # No direct skill calls
        "references": reference_skills,
        "interaction": interaction_skills
    }
```

## Command Skill Behavior Summary

### Verified Command Behaviors

#### 1. Command Delegation Pattern
- **Commands do NOT call skills directly**
- **Commands delegate to specialized agents**
- **Agents handle skill loading and execution**
- **Commands provide skill references for guidance**

#### 2. Reference-Only Pattern
- **Skills are referenced in command documentation**
- **References provide detailed information sources**
- **No automatic skill loading by commands**
- **Users must manually invoke skills if needed**

#### 3. Interactive Pattern
- **Single verified pattern**: `moai-alfred-ask-user-questions`
- **Required before using `AskUserQuestion` tool**
- **Provides TUI-based interaction guidance**
- **Only skill directly called by commands**

#### 4. Agent-Centric Pattern
- **Commands orchestrate agent execution**
- **Agents contain the actual skill invocation logic**
- **Clear separation of concerns**
- **No command-level skill management**

## Command Skill Categories

### Direct Invocation Skills (Very Limited)
```python
# Only ONE verified pattern
DIRECT_INVOCATION = [
    "moai-alfred-ask-user-questions"  # For user interaction
]
```

### Reference Skills (Documentation Only)
```python
# Skills referenced for guidance
REFERENCE_SKILLS = {
    "spec_creation": [
        "moai-foundation-ears",
        "moai-foundation-specs",
        "moai-alfred-spec-metadata-validation",
        "moai-foundation-tags"
    ],
    "workflow_guidance": [
        "moai-alfred-workflow",
        "moai-alfred-agent-guide",
        "moai-alfred-rules",
        "moai-alfred-dev-guide"
    ],
    "validation": [
        "moai-foundation-trust",
        "moai-foundation-tags",
        "moai-foundation-specs"
    ]
}
```

### Agent-Managed Skills
```python
# Skills managed by agents, not commands
AGENT_MANAGED = {
    "spec-builder": [
        "moai-foundation-ears",
        "moai-foundation-specs",
        "moai-alfred-ears-authoring",
        "moai-alfred-spec-metadata-validation",
        "moai-foundation-tags",
        "moai-foundation-trust",
        "moai-alfred-ask-user-questions"
    ],
    "cc-manager": [
        "moai-foundation-specs",
        "moai-alfred-workflow",
        # ... other skills as documented in agent patterns
    ]
}
```

## Important Constraints

### Only Verified Patterns
- All command patterns documented here exist in official command files
- No references to non-existent command behaviors
- No hallucinated or planned command features

### Real Command Behavior Only
- All patterns extracted from actual command implementations
- No suggested improvements or optimizations
- Pure documentation of existing command behavior

### No Direct Skill Loading
- Commands do NOT load skills directly (except one exception)
- Commands delegate to agents for skill management
- No JIT loading system in commands
- No automatic skill selection in commands

### Reference-Only Documentation
- Most skill references are for documentation purposes
- Commands point to skills for detailed information
- No automatic skill invocation by commands
- Users must manually invoke referenced skills if needed

---

**Generated**: 2025-11-05
**Source**: Official command documentation analysis
**Scope**: Existing `.claude/commands/` directory only
**Phase**: Phase 2 Documentation Standardization