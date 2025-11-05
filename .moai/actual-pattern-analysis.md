# Actual Pattern Analysis

**Phase 1 Documentation Audit - Real Skill Invocation Patterns Only**

## Overview

This document analyzes actual skill invocation patterns found in official MoAI-ADK sources, removing any references to non-existent JIT skills or suggested implementations not present in the official codebase.

## Verified Skill Invocation Patterns

### 1. Automatic Skills (Always Loaded)

From `cc-manager.md`, these skills are automatically loaded:

```python
# Core foundation skills
Skill("moai-foundation-specs")     # SPEC structure validation
Skill("moai-alfred-workflow")     # Decision trees & architecture
```

**Pattern**: Always loaded when cc-manager agent is invoked, providing foundational validation and workflow guidance.

### 2. Conditional Skills (Based on Request)

From `cc-manager.md`, these skills are loaded conditionally:

```python
# Language and framework detection
Skill("moai-alfred-language-detection")  # Detect project language
Skill("moai-foundation-langs")           # Auto-detect from package files

# TAG and validation systems
Skill("moai-foundation-tags")            # Validate TAG chains
Skill("moai-foundation-trust")           # TRUST 5 validation

# Git workflow
Skill("moai-alfred-git-workflow")        # Git strategy impact

# Domain-specific skills (loaded when relevant)
# Language skills (23 available) - Based on detected language
# Technical domain skills - Based on project type

# User interaction
Skill("moai-alfred-ask-user-questions") # User clarification
```

**Pattern**: Loaded on-demand based on specific trigger conditions or project requirements.

### 3. SPEC Builder Skill Patterns

From `spec-builder.md`, these skills are used during SPEC creation:

#### Core SPEC Skills (Always Used)
```python
# EARS and SPEC structure
Skill("moai-foundation-ears")            # EARS pattern maintenance
Skill("moai-foundation-specs")           # SPEC metadata structure
```

#### Conditional SPEC Skills
```python
# Authoring and validation
Skill("moai-alfred-ears-authoring")     # Auto-expand detailed requests
Skill("moai-alfred-spec-metadata-validation")  # Check ID/version/status
Skill("moai-foundation-tags")            # TAG chain traceability
Skill("moai-foundation-trust")           # Preemptive quality gates
```

#### User Interaction
```python
# Interactive prompts
Skill("moai-alfred-ask-user-questions") # User approval/modification collection
```

**Pattern**: SPEC creation uses foundation skills for structure, conditional skills for validation, and user interaction skills for approval.

### 4. Agent-Specific Skill Patterns

From agent documentation, verified skill usage patterns:

#### CC-Manager Agent Delegation Pattern
```python
# Architecture decisions
Skill("moai-alfred-workflow")

# Claude Code configuration
Skill("moai-cc-hooks")
Skill("moai-cc-agents")
Skill("moai-cc-commands")
Skill("moai-cc-skills")
Skill("moai-cc-settings")
Skill("moai-cc-mcp-plugins")

# Documentation
Skill("moai-cc-claude-md")

# Performance
Skill("moai-cc-memory")
```

**Pattern**: CC-manager delegates knowledge to specialized skills rather than containing knowledge itself.

### 5. Command Integration Patterns

From command files, legitimate skill invocations:

#### Interactive Question Pattern
```python
# Always invoke before using AskUserQuestion
Skill("moai-alfred-ask-user-questions")
```

#### Reference Pattern (Commands → Skills)
Commands reference skills for detailed guidance but don't call them directly:
- `Skill("moai-foundation-ears")` for EARS syntax
- `Skill("moai-foundation-specs")` for SPEC metadata
- `Skill("moai-alfred-agent-guide")` for agent selection
- `Skill("moai-alfred-rules")` for rules and standards

## Non-Existent JIT Skills (Removed)

The following skills are mentioned in command documentation but **do not exist** in the codebase and have been removed from this analysis:

- ❌ `moai-session-info` - Not found in `.claude/skills/`
- ❌ `moai-jit-docs-enhanced` - Not found in `.claude/skills/`
- ❌ `moai-streaming-ui` - Not found in `.claude/skills/`
- ❌ `moai-change-logger` - Not found in `.claude/skills/`
- ❌ `moai-tag-policy-validator` - Not found in `.claude/skills/`
- ❌ `moai-learning-optimizer` - Not found in `.claude/skills/`

These were hallucinated or planned features that don't exist in the actual implementation.

## Real Skill Categories and Their Patterns

### Foundation Skills Pattern
```python
# Always available for core functionality
Skill("moai-foundation-specs")      # SPEC validation
Skill("moai-foundation-ears")       # EARS patterns
Skill("moai-foundation-tags")       # TAG management
Skill("moai-foundation-trust")      # Quality gates
Skill("moai-foundation-langs")      # Language detection
Skill("moai-foundation-git")        # Git operations
```

### Alfred Workflow Skills Pattern
```python
# Workflow orchestration
Skill("moai-alfred-workflow")       # 4-step workflow
Skill("moai-alfred-ask-user-questions")  # User interaction
Skill("moai-alfred-agent-guide")    # Agent selection
Skill("moai-alfred-personas")       # Adaptive behavior
Skill("moai-alfred-reporting")      # Report standards
```

### Domain-Specific Skills Pattern
```python
# Language-specific (loaded based on detection)
Skill("moai-lang-python")           # For Python projects
Skill("moai-lang-typescript")       # For TypeScript projects
Skill("moai-lang-go")               # For Go projects
# ... (20+ other language skills)

# Domain-specific (loaded based on project type)
Skill("moai-domain-backend")        # For backend development
Skill("moai-domain-frontend")       # For frontend development
Skill("moai-domain-database")       # For database projects
Skill("moai-domain-security")       # For security concerns
# ... (other domain skills)
```

### Configuration Skills Pattern
```python
# Claude Code configuration
Skill("moai-cc-settings")           # Settings management
Skill("moai-cc-hooks")              # Hook configuration
Skill("moai-cc-skills")             # Skill creation
Skill("moai-cc-agents")             # Agent creation
Skill("moai-cc-commands")           # Command design
```

## Actual Skill Loading Behavior

### Automatic Loading
- Foundation skills are loaded automatically by agents like cc-manager
- No JIT (Just-In-Time) loading system exists - skills are explicitly invoked

### Conditional Loading
- Language skills are selected based on project detection
- Domain skills are loaded based on project requirements
- User interaction skills are loaded when ambiguity is detected

### Progressive Disclosure
- Skills provide different levels of detail based on user expertise
- Reference skills point to other skills for detailed information
- No automatic loading based on keywords - explicit invocation required

## Verified Agent-Skill Interaction Patterns

### 1. Agent Delegation Pattern (cc-manager)
```python
# cc-manager delegates to specialized skills
# Agent validates/creates files → Skills provide knowledge
```

### 2. SPEC Creation Pattern (spec-builder)
```python
# spec-builder uses foundation skills for structure
# Foundation skills provide EARS and SPEC standards
# User interaction skills collect approval
```

### 3. Interactive Pattern (All Agents)
```python
# All agents use moai-alfred-ask-user-questions
# Provides TUI-based interaction for ambiguous decisions
```

## Command-to-Skill Reference Patterns

Commands reference skills for guidance but don't call them directly:

- `/alfred:1-plan` references EARS and SPEC skills for guidance
- `/alfred:3-sync` references TAG and validation skills
- Commands act as orchestrators, not skill callers

## Summary of Actual Patterns

1. **Explicit Invocation Only**: All skill usage is through explicit `Skill("name")` calls
2. **No JIT Loading**: No just-in-time loading system exists
3. **Agent Delegation**: Agents delegate knowledge to skills rather than containing it
4. **Foundation First**: Core functionality provided by foundation skills
5. **Domain Specialization**: Language and domain skills provide specialized guidance
6. **Interactive Layer**: User interaction handled by dedicated skills
7. **Configuration Separation**: Claude Code configuration managed by separate skills

This analysis reflects only the actual, existing skill patterns in the MoAI-ADK codebase, excluding any non-existent or planned features.

---

**Generated**: 2025-11-05
**Source**: Official agent and command documentation analysis
**Scope**: Existing `.claude/skills/` directory only