# Claude Code Core Integration Reference

## Context7 Integration Patterns

### Official Documentation Sources
- **Claude Code Official Documentation**: https://code.claude.com/docs/en/skills
- **Skills Creation Guide**: https://code.claude.com/docs/en/sub-agents
- **MCP Integration Patterns**: https://code.claude.com/docs/en/mcp

### Context7 Library Mappings
```python
CONTEXT7_LIBRARY_MAPPINGS = {
    # Frontend Libraries
    "react": "/facebook/react",
    "nextjs": "/vercel/next.js",
    "typescript": "/microsoft/TypeScript",
    "vitest": "/vitest-dev/vitest",

    # Backend Libraries
    "fastapi": "/tiangolo/fastapi",
    "django": "/django/django",
    "pydantic": "/pydantic/pydantic",
    "pytest": "/pytest-dev/pytest",

    # Development Tools
    "debugpy": "/microsoft/debugpy",
    "playwright": "/microsoft/playwright",
    "nodejs": "/nodejs/node",

    # Security
    "owasp": "/owasp/top-ten",
    "bcrypt": "/pyca/bcrypt"
}
```

## Command System Reference

### Core MoAI Commands
| Command | Purpose | Delegate | Token Budget |
|---------|---------|----------|-------------|
| `/moai:0-project` | Project initialization | project-manager | 15K |
| `/moai:1-plan` | SPEC generation | spec-builder | 30K |
| `/moai:2-run` | TDD implementation | tdd-implementer | 180K |
| `/moai:3-sync` | Documentation sync | docs-manager | 40K |
| `/moai:9-feedback` | Feedback collection | quality-gate | 10K |

### Command Frontmatter Fields

#### Required Fields
```yaml
name: command-name                    # Auto-generated from filename
description: "Brief command description"
```

#### Optional Fields
```yaml
argument-hint: "arg1 arg2 --option value"  # Parameter hints
allowed-tools:                          # Tool restrictions
  - Task
  - AskUserQuestion
  - Skill
model: "sonnet"                         # Model selection
skills:                                  # Auto-loaded skills
  - relevant-skill-1
  - relevant-skill-2
```

### Advanced Command Features

#### Pre-execution Context (`!` prefix)
```yaml
# Execute bash commands before main execution
!git status --porcelain
!git branch --show-current
!find .moai/specs -name "*.md"
```

#### File References (`@` prefix)
```yaml
# Include file contents automatically
@.moai/config/config.json
@.moai/specs/SPEC-001/spec.md
@src/main.py
```

#### Dynamic Arguments
```yaml
# Positional arguments
$1      # First argument
$ARGUMENTS  # All arguments

# Variable expansion
{{project-root}}     # Project root path
{{semantic-version}} # Version from config
```

## Skill Architecture Reference

### Official Skill Structure
```
skill-name/
├── SKILL.md              # Required (≤500 lines)
├── reference.md          # Optional documentation
├── examples.md           # Optional examples
├── scripts/              # Optional utilities
│   └── helper.py
└── templates/            # Optional templates
    └── template.md
```

### Skill Loading Priority
1. **Project Skills**: `.claude/skills/` (version-controlled)
2. **Personal Skills**: `~/.claude/skills/` (individual)
3. **Plugin Skills**: Bundled with Claude Code plugins

### Progressive Disclosure Pattern
```
Level 1: SKILL.md (always loaded, <5K tokens)
  ├── Quick Reference (30-second value)
  ├── Implementation Guide (step-by-step)
  └── Advanced Patterns (expert-level)

Level 2: reference.md (loaded if referenced)
Level 3: scripts/, templates/ (loaded on demand)
```

## Memory Management Reference

### Three-Layer Architecture

#### Working Memory (Active Context)
- **Size**: 50K tokens maximum
- **Lifetime**: Current session only
- **Contents**: Session state, active contexts, recent files
- **Optimization**: Context seeding, progressive loading

#### Long-Term Memory (Persistent Storage)
- **Size**: Unlimited (compressed)
- **Lifetime**: Persistent across sessions
- **Contents**: Skill knowledge, agent configs, history
- **Optimization**: Semantic compression, hierarchical summarization

#### Context Cache (Intelligent Buffer)
- **Size**: 20K tokens
- **Lifetime**: Session + TTL
- **Contents**: Frequent patterns, compiled knowledge
- **Optimization**: LRU eviction, TTL-based invalidation

### Token Budget Allocation
```python
BUDGET_ALLOCATION = {
    "system_context": 0.10,      # 10% system prompts
    "working_memory": 0.30,      # 30% active context
    "knowledge_base": 0.20,      # 20% skills/docs
    "agent_context": 0.15,       # 15% agent states
    "interaction_buffer": 0.25   # 25% conversation
}
```

## Configuration Management Reference

### Git Strategy 3-Mode System

#### Mode 1: Manual (Local Git Only)
```json
{
  "git_strategy": {
    "mode": "manual",
    "github_integration": false,
    "auto_branch": false,
    "auto_push": false
  }
}
```

#### Mode 2: Personal (GitHub Individual)
```json
{
  "git_strategy": {
    "mode": "personal",
    "github_integration": true,
    "auto_branch": true,
    "auto_commit": true,
    "auto_push": true
  }
}
```

#### Mode 3: Team (GitHub Team)
```json
{
  "git_strategy": {
    "mode": "team",
    "github_integration": true,
    "auto_branch": true,
    "auto_commit": true,
    "auto_push": true,
    "auto_pr": true,
    "draft_pr": true,
    "required_reviews": 1
  }
}
```

### Branch Creation Control
```json
{
  "branch_creation": {
    "prompt_always": true,     # Ask on every SPEC
    "auto_enabled": false      # Auto-create after approval
  }
}
```

### Configuration Precedence
1. **Base Config**: `.moai/config/config.json`
2. **Environment Config**: `.moai/config/{env}.json`
3. **Secrets**: Environment variables, Vault, `.env` file
4. **Runtime Overrides**: Session-specific settings

## Hooks System Reference

### Hook Categories
```yaml
hooks:
  pre_tools:
    - matcher: "Bash"
      hooks:
        - type: security_validator
          command: python ~/.claude/hooks/validate_command.py
    - matcher: "Edit|Write"
      hooks:
        - type: code_analyzer
          command: python ~/.claude/hooks/analyze_code.py

  post_tools:
    - matcher: "Edit"
      hooks:
        - type: auto_formatter
          command: python ~/.claude/hooks/format_code.py
    - matcher: "Bash"
      hooks:
        - type: performance_monitor
          command: python ~/.claude/hooks/monitor_performance.py

  session_management:
    - matcher: "*"
      hooks:
        - type: context_optimizer
          command: python ~/.claude/hooks/optimize_context.py
```

### AI-Enhanced Hook Features
- **ML Threat Detection**: Pattern-based security validation
- **Behavioral Analysis**: Usage pattern monitoring
- **Predictive Optimization**: Performance tuning based on history
- **Context7 Integration**: Latest pattern application

## Quality Gates (TRUST 5)

### Test-First Requirements
```yaml
test_requirements:
  coverage_threshold: 85
  critical_path_coverage: 100
  test_categories:
    - unit_tests
    - integration_tests
    - acceptance_tests
    - edge_cases
```

### Readability Standards
```yaml
readability_metrics:
  max_function_length: 50
  max_complexity: 10
  max_nesting_depth: 3
  comment_ratio_min: 0.15
  comment_ratio_max: 0.25
```

### Unified Patterns
```yaml
consistency_requirements:
  naming_convention: consistent
  error_handling: unified
  logging_strategy: standardized
  documentation_format: consistent
```

### Security Requirements
```yaml
security_standards:
  owasp_compliance: true
  input_validation: required
  secret_management: enforced
  dependency_scanning: automatic
```

### Trackability Requirements
```yaml
traceability_standards:
  spec_linking: required
  test_tracing: mandatory
  change_origin: documented
  quality_metrics: tracked
```

## Troubleshooting Guide

### Common Issues

#### Skill Not Loading
**Symptoms**: Skill not discovered by Claude
**Causes**: Invalid YAML frontmatter, file structure issues
**Solutions**:
1. Validate YAML syntax with `python -c "import yaml; yaml.safe_load(open('SKILL.md'))"`
2. Check file structure matches official template
3. Verify skill name follows kebab-case pattern

#### Memory Overflow
**Symptoms**: Context limit exceeded, slow performance
**Causes**: Large file loading, insufficient cleanup
**Solutions**:
1. Run `python scripts/memory-optimizer.py --optimize`
2. Use selective file loading with @file references
3. Execute `/clear` between major phases

#### Command Failures
**Symptoms**: Command not found, execution errors
**Causes**: Invalid syntax, missing dependencies, permission issues
**Solutions**:
1. Validate command frontmatter structure
2. Check allowed-tools permissions
3. Verify script dependencies and paths

### Debug Tools
```bash
# Validate skills
python scripts/skill-validator.py --validate-all

# Optimize memory
python scripts/memory-optimizer.py --analyze

# Manage configuration
python scripts/config-manager.py --validate

# Check Claude Code debugging
claude --debug
```

## Performance Optimization

### Token Optimization Strategies
1. **Phase Separation**: Use `/clear` between major phases
2. **Selective Loading**: Load only necessary files
3. **Result Caching**: Preserve reusable information
4. **Concise Prompts**: Remove redundant explanations
5. **Model Selection**: Use Haiku for simple tasks

### Memory Optimization Strategies
1. **Three-Layer Architecture**: Separate working/LT memory
2. **Context Seeding**: Initialize with essential data only
3. **Progressive Loading**: Load additional data on-demand
4. **Cache Management**: Implement TTL and LRU eviction
5. **Compression**: Use semantic compression for storage

## External Integrations

### Context7 MCP Integration
```python
# Library resolution
library_id = await mcp__context7__resolve-library_id("react")

# Documentation access
docs = await mcp__context7__get-library_docs(
    context7CompatibleLibraryID=library_id,
    topic="hooks and patterns",
    tokens=3000
)
```

### MCP Server Requirements
```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["@upstash/context7-mcp@latest"]
    },
    "sequential-thinking": {
      "command": "npx",
      "args": ["@modelcontextprotocol/server-sequential-thinking@latest"]
    }
  }
}
```

## Version History

### moai-cc-core v1.0.0
- **Consolidated** 7 moai-cc* skills into unified core
- **Implemented** three-layer memory architecture
- **Added** AI-enhanced hooks system
- **Integrated** Context7 MCP patterns
- **Standardized** TRUST 5 quality framework
- **Created** enterprise configuration management
- **Provided** utility scripts for automation

### Merged Skills
- `moai-cc-commands`: Command system and workflow orchestration
- `moai-cc-configuration`: Enterprise configuration patterns
- `moai-cc-hooks`: AI-enhanced hooks system
- `moai-cc-memory`: Memory management and optimization
- `moai-cc-skills-guide`: Skill creation and management
- `moai-cc-skill-factory`: AI-powered skill generation
- `moai-cc-claude-md`: Claude Code markdown standards

---

**Last Updated**: 2025-11-24
**Version**: 1.0.0
**Status**: Production Ready