# Agent Skill Selection Patterns

**Phase 2 Documentation Standardization - Verified Agent Patterns Only**

## Overview

This document documents the actual agent skill selection patterns found in the MoAI-ADK codebase, extracted exclusively from verified agent implementations. All patterns documented here are based on real agent behavior and skill invocations.

## Agent Selection Pattern

### Core Selection Function
```python
def select_agent_skills(agent_type, task_context):
    """Actual skill selection from verified agents"""

    # SPEC Builder Pattern (from spec-builder.md)
    if agent_type == "spec-builder":
        return [
            "moai-foundation-ears",                # Always required
            "moai-foundation-specs",               # Always required
            "moai-alfred-ears-authoring",         # Conditional for detailed requests
            "moai-alfred-spec-metadata-validation",  # Conditional for validation
            "moai-foundation-tags",               # Conditional for traceability
            "moai-foundation-trust",              # Conditional for quality gates
            "moai-alfred-ask-user-questions"     # Always for user approval
        ]

    # CC-Manager Pattern (from cc-manager.md)
    elif agent_type == "cc-manager":
        core_skills = [
            "moai-foundation-specs",              # Always loaded
            "moai-alfred-workflow"                # Always loaded
        ]

        # Conditional additions based on context
        if task_context.get("language_detection_needed"):
            core_skills.extend([
                "moai-alfred-language-detection",
                "moai-foundation-langs"
            ])

        if task_context.get("validation_needed"):
            core_skills.extend([
                "moai-foundation-tags",
                "moai-foundation-trust"
            ])

        if task_context.get("git_workflow_needed"):
            core_skills.append("moai-alfred-git-workflow")

        if task_context.get("user_interaction_needed"):
            core_skills.append("moai-alfred-ask-user-questions")

        return core_skills

    return []
```

## SPEC-Builder Agent Pattern

### Skill Selection Logic
```python
def spec_builder_skill_selection(spec_context):
    """SPEC builder skill selection from spec-builder.md"""

    # Base foundation skills (always required)
    skills = [
        "moai-foundation-ears",         # EARS pattern maintenance
        "moai-foundation-specs"         # SPEC metadata structure
    ]

    # Authoring skills (conditional)
    if spec_context.get("detailed_requirements"):
        skills.append("moai-alfred-ears-authoring")

    # Validation skills (conditional)
    if spec_context.get("metadata_validation"):
        skills.append("moai-alfred-spec-metadata-validation")

    if spec_context.get("tag_traceability"):
        skills.append("moai-foundation-tags")

    if spec_context.get("quality_gates"):
        skills.append("moai-foundation-trust")

    # User interaction (always for approvals)
    skills.append("moai-alfred-ask-user-questions")

    return skills
```

### SPEC-Builder Skill Categories
```python
# Foundation Skills (Core)
SPEC_BUILDER_FOUNDATION = [
    "moai-foundation-ears",      # EARS syntax guidance
    "moai-foundation-specs"      # SPEC structure validation
]

# Authoring Skills (Conditional)
SPEC_BUILDER_AUTHORING = [
    "moai-alfred-ears-authoring"  # Detailed request expansion
]

# Validation Skills (Conditional)
SPEC_BUILDER_VALIDATION = [
    "moai-alfred-spec-metadata-validation",  # Metadata validation
    "moai-foundation-tags",                 # TAG traceability
    "moai-foundation-trust"                 # Quality gates
]

# Interaction Skills (Always)
SPEC_BUILDER_INTERACTION = [
    "moai-alfred-ask-user-questions"       # User approval
]
```

## CC-Manager Agent Pattern

### Skill Selection Logic
```python
def cc_manager_skill_selection(task_context):
    """CC manager skill selection from cc-manager.md"""

    # Core foundation skills (always loaded)
    skills = [
        "moai-foundation-specs",      # SPEC structure validation
        "moai-alfred-workflow"        # Decision trees & architecture
    ]

    # Language detection (conditional)
    if task_context.get("project_analysis"):
        skills.extend([
            "moai-alfred-language-detection",
            "moai-foundation-langs"
        ])

    # Domain-specific skills (conditional)
    detected_language = task_context.get("detected_language")
    if detected_language:
        language_skill = f"moai-lang-{detected_language}"
        if language_skill in VERIFIED_LANGUAGE_SKILLS:
            skills.append(language_skill)

    project_type = task_context.get("project_type")
    if project_type:
        domain_skill = f"moai-domain-{project_type}"
        if domain_skill in VERIFIED_DOMAIN_SKILLS:
            skills.append(domain_skill)

    # Validation systems (conditional)
    if task_context.get("validation_required"):
        skills.extend([
            "moai-foundation-tags",
            "moai-foundation-trust"
        ])

    # Claude Code configuration (conditional)
    if task_context.get("claude_code_setup"):
        skills.extend([
            "moai-cc-hooks",
            "moai-cc-agents",
            "moai-cc-commands",
            "moai-cc-skills",
            "moai-cc-settings",
            "moai-cc-mcp-plugins",
            "moai-cc-memory"
        ])

    # User interaction (conditional)
    if task_context.get("user_ambiguity"):
        skills.append("moai-alfred-ask-user-questions")

    return skills
```

### CC-Manager Skill Categories
```python
# Core Skills (Always Loaded)
CC_MANAGER_CORE = [
    "moai-foundation-specs",      # SPEC validation
    "moai-alfred-workflow"        # Workflow orchestration
]

# Language Skills (Conditional)
CC_MANAGER_LANGUAGE = [
    "moai-alfred-language-detection",  # Language detection
    "moai-foundation-langs"           # Package file detection
]

# Domain Skills (Conditional)
CC_MANAGER_DOMAIN = [
    # Language-specific skills (23 available)
    "moai-lang-python", "moai-lang-typescript", "moai-lang-go",
    # Domain-specific skills (25 available)
    "moai-domain-backend", "moai-domain-frontend", "moai-domain-database"
]

# Validation Skills (Conditional)
CC_MANAGER_VALIDATION = [
    "moai-foundation-tags",       # TAG validation
    "moai-foundation-trust"       # TRUST validation
]

# Configuration Skills (Conditional)
CC_MANAGER_CONFIG = [
    "moai-cc-hooks",              # Hook configuration
    "moai-cc-agents",             # Agent creation
    "moai-cc-commands",           # Command design
    "moai-cc-skills",             # Skill creation
    "moai-cc-settings",           # Settings management
    "moai-cc-mcp-plugins",        # MCP configuration
    "moai-cc-memory"              # Memory management
]

# Interaction Skills (Conditional)
CC_MANAGER_INTERACTION = [
    "moai-alfred-ask-user-questions"  # User clarification
]
```

## Agent Type Skill Mapping

### Complete Agent Skill Mapping
```python
AGENT_SKILL_MAPPING = {
    "spec-builder": {
        "core": [
            "moai-foundation-ears",
            "moai-foundation-specs"
        ],
        "conditional": [
            "moai-alfred-ears-authoring",
            "moai-alfred-spec-metadata-validation",
            "moai-foundation-tags",
            "moai-foundation-trust"
        ],
        "interaction": [
            "moai-alfred-ask-user-questions"
        ]
    },

    "cc-manager": {
        "core": [
            "moai-foundation-specs",
            "moai-alfred-workflow"
        ],
        "conditional": [
            "moai-alfred-language-detection",
            "moai-foundation-langs",
            "moai-foundation-tags",
            "moai-foundation-trust",
            "moai-alfred-git-workflow",
            # Language skills (23 available)
            # Domain skills (25 available)
            # CC configuration skills (7 available)
        ],
        "interaction": [
            "moai-alfred-ask-user-questions"
        ]
    }
}
```

## Conditional Skill Loading Patterns

### Language-Based Selection
```python
def select_language_skills(detected_language):
    """Language-based skill selection"""

    language_skill_map = {
        "python": "moai-lang-python",
        "typescript": "moai-lang-typescript",
        "javascript": "moai-lang-javascript",
        "go": "moai-lang-go",
        "rust": "moai-lang-rust",
        "java": "moai-lang-java",
        "php": "moai-lang-php",
        "kotlin": "moai-lang-kotlin",
        "ruby": "moai-lang-ruby",
        "r": "moai-lang-r",
        "c": "moai-lang-c",
        "cpp": "moai-lang-cpp",
        "csharp": "moai-lang-csharp",
        "scala": "moai-lang-scala",
        "swift": "moai-lang-swift",
        "dart": "moai-lang-dart",
        "sql": "moai-lang-sql"
    }

    skill = language_skill_map.get(detected_language)
    return [skill] if skill else []
```

### Domain-Based Selection
```python
def select_domain_skills(project_type):
    """Domain-based skill selection"""

    domain_skill_map = {
        "backend": "moai-domain-backend",
        "frontend": "moai-domain-frontend",
        "database": "moai-domain-database",
        "security": "moai-domain-security",
        "cli": "moai-domain-cli-tool",
        "ml": "moai-domain-ml",
        "mobile": "moai-domain-mobile-app",
        "api": "moai-domain-web-api",
        "data": "moai-domain-data-science",
        "devops": "moai-domain-devops"
    }

    skill = domain_skill_map.get(project_type)
    return [skill] if skill else []
```

### Context-Based Selection
```python
def select_contextual_skills(context):
    """Context-based skill selection"""

    skills = []

    # Validation context
    if context.get("validation_needed"):
        skills.extend([
            "moai-foundation-tags",
            "moai-foundation-trust"
        ])

    # Git context
    if context.get("git_workflow"):
        skills.append("moai-alfred-git-workflow")

    # User interaction context
    if context.get("user_ambiguity"):
        skills.append("moai-alfred-ask-user-questions")

    # Configuration context
    if context.get("claude_code_setup"):
        skills.extend([
            "moai-cc-hooks",
            "moai-cc-settings",
            "moai-cc-skills"
        ])

    return skills
```

## Agent Skill Loading Sequence

### Standard Loading Order
```python
def load_agent_skills_in_sequence(agent_type, context):
    """Standard agent skill loading sequence"""

    # 1. Core foundation skills (always first)
    core_skills = get_core_skills(agent_type)

    # 2. Language and detection skills
    detection_skills = get_detection_skills(context)

    # 3. Domain-specific skills
    domain_skills = get_domain_skills(context)

    # 4. Validation and quality skills
    validation_skills = get_validation_skills(context)

    # 5. Configuration skills (if needed)
    config_skills = get_config_skills(context)

    # 6. User interaction skills (always last)
    interaction_skills = get_interaction_skills(context)

    # Combine in order
    all_skills = (
        core_skills +
        detection_skills +
        domain_skills +
        validation_skills +
        config_skills +
        interaction_skills
    )

    return all_skills
```

## Verified Agent Behavior Patterns

### SPEC-Builder Behavior
1. **Always loads foundation skills** (ears, specs)
2. **Conditionally loads authoring skills** for detailed requirements
3. **Conditionally loads validation skills** for metadata and tags
4. **Always loads interaction skills** for user approval
5. **No automatic domain skill loading** (focus on SPEC creation)

### CC-Manager Behavior
1. **Always loads core foundation skills** (specs, workflow)
2. **Conditionally loads detection skills** for project analysis
3. **Conditionally loads domain skills** based on project context
4. **Conditionally loads configuration skills** for Claude Code setup
5. **Conditionally loads interaction skills** for user clarification
6. **Delegates knowledge to specialized skills** rather than containing it

## Important Constraints

### Only Verified Patterns
- All agent patterns documented here exist in official agent files
- No references to non-existent agent behaviors
- No hallucinated or planned agent features

### Real Agent Behavior Only
- All patterns extracted from actual agent implementations
- No suggested improvements or optimizations
- Pure documentation of existing agent behavior

### No Agent Creation Patterns
- No templates for creating new agents
- No agent optimization strategies
- Only documentation of existing agent selection logic

---

**Generated**: 2025-11-05
**Source**: Official agent documentation analysis
**Scope**: Existing `.claude/agents/` directory only
**Phase**: Phase 2 Documentation Standardization