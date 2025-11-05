# Phase 2: Documentation Standardization (1-2 Weeks)

## Objective
Create standardized documentation based ONLY on existing, verified patterns from the official MoAI-ADK codebase.

## Tasks

### Task 1: Document Existing Skill Combinations
**Based ONLY on verified patterns from official sources**:

```python
# Template 1: SPEC Creation (from spec-builder.md)
def load_spec_creation_skills():
    """Actual SPEC creation skill combination"""
    Skill("moai-foundation-ears")()   # Always required

    # Conditional additions (from spec-builder.md)
    if need_metadata_validation:
        Skill("moai-alfred-spec-metadata-validation")()
    if need_tag_references:
        Skill("moai-foundation-tags")()
    if need_new_spec_structure:
        Skill("moai-foundation-specs")()

# Template 2: Agent Core Skills (from cc-manager.md)
def load_agent_core_skills():
    """Standard agent core skills"""
    Skill("moai-foundation-specs")()      # Always loaded for core agents
    Skill("moai-alfred-workflow")()       # Decision trees & architecture

# Template 3: Conditional Agent Skills (from cc-manager.md)
def load_conditional_skills():
    """Conditional skills based on request"""
    if language_detection_needed:
        Skill("moai-alfred-language-detection")()
    if validation_needed:
        Skill("moai-foundation-trust")()
    if tag_validation_needed:
        Skill("moai-foundation-tags")()
```

**Deliverable**: `existing-skill-combinations.md`

### Task 2: Document Agent Skill Selection Patterns
**From actual agent implementations**:

```python
# Agent Selection Pattern (from cc-manager.md and spec-builder.md)
def select_agent_skills(agent_type, task_context):
    """Actual skill selection from verified agents"""

    # spec-builder pattern (from spec-builder.md)
    if agent_type == "spec-builder":
        return [
            "moai-foundation-ears",  # Always required
            "moai-alfred-spec-metadata-validation",  # Conditional
            "moai-foundation-tags",  # Conditional
            "moai-foundation-specs",  # Conditional
            "moai-foundation-trust"   # Conditional
        ]

    # cc-manager pattern (from cc-manager.md)
    elif agent_type == "cc-manager":
        return [
            "moai-foundation-specs",  # Always loaded
            "moai-alfred-workflow"    # Always loaded
        ]
```

**Deliverable**: `agent-skill-patterns.md`

### Task 3: Document Command Skill Patterns
**From actual command implementations (alfred:1-plan.md)**:

```python
# Command Pattern (from alfred:1-plan.md)
def load_command_skills(command_name, context):
    """Actual command skill loading"""

    if command_name == "alfred:1-plan":
        return [
            "moai-foundation-specs",  # From Phase 1
            "moai-foundation-ears",   # From Phase 1
            # Note: Remove non-existent JIT skills
        ]
```

**Important**: Remove references to non-existent JIT skills:
- ❌ `moai-session-info`
- ❌ `moai-jit-docs-enhanced`
- ❌ `moai-streaming-ui`
- ❌ `moai-change-logger`
- ❌ `moai-tag-policy-validator`
- ❌ `moai-learning-optimizer`

**Deliverable**: `command-skill-patterns.md`

## Expected Outcomes
- Documentation of ONLY existing skill combinations
- Clear patterns from actual agent implementations
- Command patterns without hallucinated JIT skills
- No suggested improvements or optimizations

## Success Criteria
- All documented patterns exist in official codebase
- No references to non-existent skills
- No performance optimizations (not in official docs)
- Pure documentation of existing patterns

## Exclusions (What NOT to include)
- ❌ Caching mechanisms (don't exist in official docs)
- ❌ Performance optimizations (not documented)
- ❌ Shared utility functions (not in official architecture)
- ❌ New skill templates (only document existing)
- ❌ Decision trees not found in official agents