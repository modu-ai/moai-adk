# Existing Skill Combinations

**Phase 2 Documentation Standardization - Verified Skill Combinations Only**

## Overview

This document documents the actual skill combinations that exist in the MoAI-ADK codebase, extracted exclusively from verified agent and command implementations. All patterns documented here are based on real skill invocations found in official sources.

## Template 1: SPEC Creation Skill Combinations

### Core SPEC Skills (Always Required)
```python
def load_spec_creation_skills():
    """Actual SPEC creation skill combination from spec-builder.md"""

    # Foundation skills (always required)
    Skill("moai-foundation-ears")()           # EARS pattern maintenance
    Skill("moai-foundation-specs")()          # SPEC metadata structure

    # Authoring skills (conditional)
    if need_detailed_requirements:
        Skill("moai-alfred-ears-authoring")()     # Auto-expand detailed requests

    # Validation skills (conditional)
    if need_metadata_validation:
        Skill("moai-alfred-spec-metadata-validation")()  # Check ID/version/status

    if need_tag_traceability:
        Skill("moai-foundation-tags")()            # TAG chain traceability

    if need_quality_gates:
        Skill("moai-foundation-trust")()           # Preemptive quality gates

    # User interaction (always used for approvals)
    Skill("moai-alfred-ask-user-questions")()     # User approval/modification collection
```

### SPEC Creation Pattern Summary
- **Always Required**: 2 foundation skills (ears, specs)
- **Conditional Authoring**: 1 skill for detailed requirements
- **Conditional Validation**: Up to 2 skills for metadata and tags
- **User Interaction**: 1 skill for approval collection
- **Total Skills Used**: 2-6 depending on SPEC complexity

## Template 2: Agent Core Skills

### CC-Manager Core Skills
```python
def load_agent_core_skills():
    """Standard agent core skills from cc-manager.md"""

    # Always loaded for core agents
    Skill("moai-foundation-specs")()          # SPEC structure validation
    Skill("moai-alfred-workflow")()           # Decision trees & architecture
```

### Agent Core Pattern Summary
- **Always Loaded**: 2 core skills for all agent operations
- **Purpose**: Provide foundational validation and workflow guidance
- **Scope**: Used by cc-manager and other core agents

## Template 3: Conditional Agent Skills

### Language Detection Skills
```python
def load_language_skills():
    """Conditional language detection from cc-manager.md"""

    # Project language detection
    Skill("moai-alfred-language-detection")()  # Detect project language
    Skill("moai-foundation-langs")()           # Auto-detect from package files
```

### Validation and Quality Skills
```python
def load_validation_skills():
    """Conditional validation from cc-manager.md"""

    # TAG and validation systems
    if need_tag_validation:
        Skill("moai-foundation-tags")()        # Validate TAG chains

    if need_trust_validation:
        Skill("moai-foundation-trust")()       # TRUST 5 validation
```

### Git Workflow Skills
```python
def load_git_skills():
    """Conditional Git workflow from cc-manager.md"""

    if git_strategy_impact:
        Skill("moai-alfred-git-workflow")()    # Git strategy impact
```

### Domain-Specific Skills
```python
def load_domain_skills(detected_language, project_type):
    """Domain-specific skills from cc-manager.md"""

    # Language-specific skills (23 available)
    language_skills = {
        "python": "moai-lang-python",
        "typescript": "moai-lang-typescript",
        "go": "moai-lang-go",
        # ... 20+ other language skills
    }

    if detected_language in language_skills:
        Skill(language_skills[detected_language])()

    # Technical domain skills
    if project_type == "backend":
        Skill("moai-domain-backend")()
    elif project_type == "frontend":
        Skill("moai-domain-frontend")()
    elif project_type == "database":
        Skill("moai-domain-database")()
    elif project_type == "security":
        Skill("moai-domain-security")()
    # ... other domain skills
```

### User Interaction Skills
```python
def load_interaction_skills():
    """User interaction from cc-manager.md"""

    if user_clarification_needed:
        Skill("moai-alfred-ask-user-questions")()  # User clarification
```

## Template 4: Claude Code Configuration Skills

### CC-Manager Delegation Pattern
```python
def load_cc_configuration_skills():
    """Claude Code configuration from cc-manager.md"""

    # Claude Code configuration skills
    Skill("moai-cc-hooks")()                   # Hook configuration
    Skill("moai-cc-agents")()                  # Agent creation
    Skill("moai-cc-commands")()                # Command design
    Skill("moai-cc-skills")()                  # Skill creation
    Skill("moai-cc-settings")()                # Settings management
    Skill("moai-cc-mcp-plugins")()             # MCP plugin configuration

    # Performance and memory
    Skill("moai-cc-memory")()                  # Session memory management
```

## Template 5: Interactive Question Pattern

### Pre-Question Skill Loading
```python
def load_question_skills():
    """Always invoke before using AskUserQuestion"""

    # Required before any user interaction
    Skill("moai-alfred-ask-user-questions")()  # TUI-based interaction
```

## Verified Skill Combination Patterns

### Pattern 1: Minimal SPEC Creation
```python
# Basic SPEC with minimal validation
Skill("moai-foundation-ears")()
Skill("moai-foundation-specs")()
Skill("moai-alfred-ask-user-questions")()
# Total: 3 skills
```

### Pattern 2: Full SPEC Creation
```python
# Complete SPEC with all validations
Skill("moai-foundation-ears")()
Skill("moai-foundation-specs")()
Skill("moai-alfred-ears-authoring")()
Skill("moai-alfred-spec-metadata-validation")()
Skill("moai-foundation-tags")()
Skill("moai-foundation-trust")()
Skill("moai-alfred-ask-user-questions")()
# Total: 7 skills
```

### Pattern 3: Agent Core Operations
```python
# Standard agent operations
Skill("moai-foundation-specs")()
Skill("moai-alfred-workflow")()
# Total: 2 skills
```

### Pattern 4: Full Agent Delegation
```python
# Complete agent with all specializations
Skill("moai-foundation-specs")()
Skill("moai-alfred-workflow")()
Skill("moai-alfred-language-detection")()
Skill("moai-foundation-langs")()
Skill("moai-cc-hooks")()
Skill("moai-cc-agents")()
Skill("moai-cc-commands")()
Skill("moai-cc-skills")()
Skill("moai-cc-settings")()
Skill("moai-cc-mcp-plugins")()
Skill("moai-cc-memory")()
# Total: 11 skills
```

### Pattern 5: Domain-Specific Development
```python
# Project-specific development
Skill("moai-foundation-specs")()
Skill("moai-alfred-workflow")()
Skill("moai-alfred-language-detection")()
Skill("moai-foundation-langs")()
Skill("moai-lang-python")()                   # Example: Python project
Skill("moai-domain-backend")()                # Example: Backend project
Skill("moai-alfred-ask-user-questions")()
# Total: 7 skills
```

## Skill Loading Behavior

### Automatic Loading
- **Foundation Skills**: Automatically loaded by core agents
- **Workflow Skills**: Always loaded for agent orchestration
- **No JIT System**: All skills are explicitly invoked

### Conditional Loading
- **Language Skills**: Selected based on project detection
- **Domain Skills**: Loaded based on project requirements
- **Validation Skills**: Loaded when validation is needed
- **Interaction Skills**: Loaded when user ambiguity is detected

### Progressive Disclosure
- **Core Skills**: Provide foundational functionality
- **Specialized Skills**: Provide domain-specific expertise
- **Reference Skills**: Point to other skills for detailed information

## Important Constraints

### Only Verified Skills
- All skills documented here exist in `.claude/skills/*/SKILL.md`
- No references to non-existent JIT skills
- No hallucinated or planned features

### Real Patterns Only
- All patterns extracted from actual agent implementations
- No suggested improvements or optimizations
- Pure documentation of existing combinations

### No Performance Optimizations
- No caching mechanisms (don't exist in official docs)
- No shared utility functions (not in official architecture)
- No automatic loading based on keywords

---

**Generated**: 2025-11-05
**Source**: Official agent and command documentation analysis
**Scope**: Existing `.claude/skills/` directory only
**Phase**: Phase 2 Documentation Standardization