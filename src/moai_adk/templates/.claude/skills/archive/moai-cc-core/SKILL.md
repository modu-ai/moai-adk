---
name: moai-cc-core
description: Unified Claude Code integration hub with enterprise-grade command orchestration, skill management, memory optimization, configuration patterns, and hooks system for production workflows.
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
version: 1.0.0
modularized: true
last_updated: 2025-11-24
compliance_score: 95%
auto_trigger_keywords: claude-code, cc, core, enterprise, integration
tags:
  - enterprise
  - integration
  - orchestration
  - claude-code
---

# Claude Code Core Integration Hub

## Quick Reference (30 seconds)

**Enterprise Claude Code Orchestration Platform** - Unified hub for command execution, skill management, memory optimization, configuration patterns, and intelligent hooks system.

**Core Capabilities**:
- ✅ **Command System**: Advanced workflow orchestration with parameter validation
- ✅ **Skill Factory**: AI-powered skill creation and optimization
- ✅ **Memory Management**: Three-layer architecture with token optimization
- ✅ **Configuration Management**: Enterprise-grade config and secret handling
- ✅ **Hooks Orchestration**: AI-enhanced workflow automation
- ✅ **Integration Patterns**: Context7-powered latest standards

**When to Use**:
- Setting up enterprise Claude Code workflows
- Creating custom skills and commands
- Optimizing token usage and memory
- Managing configuration across environments
- Implementing automated hooks and workflows

---

## Core Integration Patterns

### 1. Command Architecture & Workflow Orchestration

**Command Structure**:
```
/moai:N-action [parameters] [options]

Core Commands:
- /moai:0-project     # Project initialization
- /moai:1-plan        # SPEC generation with EARS format
- /moai:2-run SPEC-ID # TDD implementation (RED-GREEN-REFACTOR)
- /moai:3-sync SPEC-ID # Documentation synchronization
- /moai:9-feedback    # Continuous improvement
```

**Advanced Command Features**:
- Pre-execution context with `!bash` commands
- File references with `@file.path` inclusion
- Dynamic arguments and variable expansion
- Model selection optimization (Sonnet vs Haiku)

### 2. Skill Factory & Management

**Skill Creation Workflow**:
```
Discovery → Architecture → Generation → Validation → Deployment
```

**Progressive Disclosure Structure**:
- **SKILL.md** (≤500 lines): Quick reference + core information
- **reference.md**: Documentation links and specifications
- **examples.md**: 10+ real-world working examples
- **scripts/**: Utility scripts and automation tools
- **templates/**: Reusable templates and patterns

### 3. Memory Management Architecture

**Three-Layer System**:
```
Working Memory (Active Context) - 50K tokens
├─ Current session state
├─ Active agent contexts
└─ Lifecycle: Current session

Long-Term Memory (Persistent Storage) - Unlimited
├─ Skill knowledge capsules
├─ Agent configurations
└─ Lifecycle: Persistent across sessions

Context Cache (Intelligent Buffer) - 20K tokens
├─ Frequently accessed patterns
├─ Precomputed embeddings
└─ Lifecycle: Session + TTL
```

**Token Optimization**:
- Phase-based allocation (Plan: 30K, Implement: 180K, Docs: 40K)
- Context seeding and progressive loading
- Strategic compression and consolidation

### 4. Configuration Management

**Enterprise Config Patterns**:
- Environment-based loading (dev/staging/prod)
- HashiCorp Vault integration for secrets
- Configuration validation with schema checking
- Kubernetes ConfigMap/Secret integration
- Git workflow configuration (3-mode system)

### 5. AI-Enhanced Hooks System

**Hook Categories**:
- **Pre-Tool Hooks**: Security validation, input analysis
- **Post-Tool Hooks**: Code optimization, formatting
- **Session Hooks**: Context management, workflow orchestration
- **AI Integration**: Context7 pattern matching, predictive optimization

---

## Implementation Workflows

### Workflow 1: Project Initialization
```bash
/moai:0-project                    # Initialize .moai/ structure
  ↓
Generate config.json               # Configure git strategy, language
  ↓
Set up skills and agents           # Default enterprise setup
  ↓
Ready for SPEC generation
```

### Workflow 2: SPEC-First Development
```bash
/moai:1-plan "Feature description" # Generate EARS format SPEC
  ↓
/clear                             # Reset context (mandatory)
  ↓
/moai:2-run SPEC-001              # TDD implementation
  ↓
/moai:3-sync SPEC-001             # Documentation sync
  ↓
/moai:9-feedback                   # Continuous improvement
```

### Workflow 3: Skill Creation
```bash
Skill Factory → Define requirements
  ↓
Architecture Design → 7-section structure
  ↓
Content Generation → Context7 integration
  ↓
Quality Assurance → TRUST 5 validation
  ↓
Deployment → Production-ready skill
```

---

## Enterprise Integration Patterns

### Configuration Management
```yaml
# Git Strategy (3-Mode System)
git_strategy:
  mode: manual | personal | team
  branch_creation:
    prompt_always: true      # Ask on every SPEC
    auto_enabled: false      # Auto-creation after approval

# Example: Personal mode with automation
{
  "mode": "personal",
  "branch_creation": {
    "prompt_always": false,
    "auto_enabled": true
  }
}
```

### Memory Optimization
```python
# Token Budget Allocation
budget_allocations = {
    "system_context": 0.10,      # 10% for system prompts
    "working_memory": 0.30,      # 30% for active context
    "knowledge_base": 0.20,      # 20% for skills/docs
    "agent_context": 0.15,       # 15% for agent states
    "interaction_buffer": 0.25   # 25% for conversation
}
```

### Hooks Integration
```json
{
  "ai_enterprise_hooks": {
    "version": "4.0.0",
    "ai_orchestration": true,
    "context7_integration": true,
    "hooks": {
      "pre_tools": [
        {
          "type": "ai_security_validator",
          "features": ["ml_threat_detection", "context7_compliance"]
        }
      ],
      "post_tools": [
        {
          "type": "ai_auto_optimizer",
          "features": ["intelligent_formatting", "security_hardening"]
        }
      ]
    }
  }
}
```

---

## Best Practices

### ✅ DO
- Use Context7 integration for latest patterns and standards
- Apply TRUST 5 framework (Test-first, Readable, Unified, Secured, Trackable)
- Implement three-layer memory architecture for optimization
- Use skill factory for consistent, production-ready skills
- Leverage AI-enhanced hooks for workflow automation
- Follow enterprise configuration patterns with Git 3-mode system

### ❌ DON'T
- Exceed 500 lines in SKILL.md (use modular structure)
- Skip Context7 validation for latest standards
- Ignore token budget management (causes overflow)
- Apply AI-generated content without validation
- Mix responsibilities in single skills (follow single responsibility)
- Skip security validation in hooks and configurations

---

## Works Well With

- `moai-context7-integration` - Latest documentation and standards
- `moai-foundation-trust` - TRUST 5 quality framework
- `moai-essentials-debug` - Advanced debugging patterns
- `moai-essentials-perf` - Performance optimization
- `moai-domain-*` skills - Domain-specific expertise

---

## Quality Gates (TRUST 5)

**Test-First**: ≥85% test coverage with comprehensive scenarios
**Readable**: Functions <50 lines, clear naming, complexity <10
**Unified**: Consistent patterns across all components
**Secured**: OWASP compliance, no vulnerabilities
**Trackable**: SPEC-linked, test-traced implementations

---

## Success Metrics

- **Automation**: 95% automated command execution
- **Performance**: 40-60% token reduction with optimization
- **Quality**: 90% success rate for AI-generated skills
- **Integration**: 95% Context7 pattern application
- **Enterprise**: 100% compliance with security standards

---

**Status**: Production Ready (Enterprise)
**Version**: 1.0.0 (Unified Core)
**Enhanced with**: Context7 integration, AI optimization, TRUST 5 framework

---

## Advanced Documentation

For detailed implementation patterns:
- **reference.md**: Complete specifications and Context7 patterns
- **examples.md**: 10+ real-world implementation examples
- **scripts/**: Utility scripts for automation and management
- **templates/**: Reusable enterprise templates

**End of Core Skill** | Modular architecture with progressive disclosure