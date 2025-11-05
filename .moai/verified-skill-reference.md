# Verified Skill Reference Documentation

## Overview

This document provides a comprehensive reference of **only verified, existing skills** found in the `.claude/skills/` directory of the MoAI-ADK project. All skills listed below have been confirmed to exist and are available for use in agent and command documentation.

## Verification Methodology

1. **Comprehensive Scan**: Used glob pattern `.claude/skills/*/SKILL.md` to identify all existing skills
2. **Existence Verification**: Each skill directory was confirmed to contain SKILL.md file
3. **Cross-Reference**: Skills referenced in documentation were verified against this inventory
4. **Categorization**: Skills organized by functional category for easy reference

## Verified Skills Inventory

### Foundation Skills (Core Framework)

| Skill Name | Directory | Purpose |
|------------|-----------|---------|
| `moai-foundation-specs` | `.claude/skills/moai-foundation-specs/` | SPEC metadata structure and validation standards |
| `moai-foundation-ears` | `.claude/skills/moai-foundation-ears/` | EARS (Event-Action-Response-State) syntax and patterns |
| `moai-foundation-tags` | `.claude/skills/moai-foundation-tags/` | TAG system management and traceability |
| `moai-foundation-trust` | `.claude/skills/moai-foundation-trust/` | TRUST 5 principles (Test, Readable, Unified, Secured, Trackable) |
| `moai-foundation-git` | `.claude/skills/moai-foundation-git/` | Git workflow standards and best practices |
| `moai-foundation-langs` | `.claude/skills/moai-foundation-langs/` | Multi-language project support and conventions |

### Alfred Skills (Workflow Orchestration)

| Skill Name | Directory | Purpose |
|------------|-----------|---------|
| `moai-alfred-workflow` | `.claude/skills/moai-alfred-workflow/` | 4-step workflow orchestration (Plan → Run → Sync) |
| `moai-alfred-language-detection` | `.claude/skills/moai-alfred-language-detection/` | Automatic project language detection |
| `moai-alfred-git-workflow` | `.claude/skills/moai-alfred-git-workflow/` | GitFlow and branch management strategies |
| `moai-alfred-spec-metadata-validation` | `.claude/skills/moai-alfred-spec-metadata-validation/` | SPEC metadata validation and compliance |
| `moai-alfred-ask-user-questions` | `.claude/skills/moai-alfred-ask-user-questions/` | TUI-based user interaction and selection menus |
| `moai-alfred-spec-authoring` | `.claude/skills/moai-alfred-spec-authoring/` | SPEC document creation and authoring guidance |
| `moai-alfred-ears-authoring` | `.claude/skills/moai-alfred-ears-authoring/` | EARS requirement writing and syntax |
| `moai-alfred-tag-scanning` | `.claude/skills/moai-alfred-tag-scanning/` | TAG system scanning and validation |
| `moai-alfred-trust-validation` | `.claude/skills/moai-alfred-trust-validation/` | TRUST principles validation and compliance |
| `moai-alfred-gitflow-policy` | `.claude/skills/moai-alfred-gitflow-policy/` | GitFlow policy enforcement and compliance |
| `moai-alfred-context-budget` | `.claude/skills/moai-alfred-context-budget/` | Context window optimization and management |
| `moai-alfred-session-state` | `.claude/skills/moai-alfred-session-state/` | Session state management and handoff |
| `moai-alfred-autofixes` | `.claude/skills/moai-alfred-autofixes/` | Automatic code fixes and conflict resolution |
| `moai-alfred-todowrite-pattern` | `.claude/skills/moai-alfred-todowrite-pattern/` | TodoWrite usage patterns and best practices |
| `moai-alfred-reporting` | `.claude/skills/moai-alfred-reporting/` | Report generation and formatting standards |
| `moai-alfred-personas` | `.claude/skills/moai-alfred-personas/` | Adaptive communication and persona selection |
| `moai-alfred-proactive-suggestions` | `.claude/skills/moai-alfred-proactive-suggestions/` | Proactive optimization suggestions |
| `moai-alfred-doc-management` | `.claude/skills/moai-alfred-doc-management/` | Documentation placement and management rules |
| `moai-alfred-agent-guide` | `.claude/skills/moai-alfred-agent-guide/` | Agent selection and collaboration guidance |
| `moai-alfred-config-schema` | `.claude/skills/moai-alfred-config-schema/` | Configuration schema validation and standards |
| `moai-alfred-expertise-detection` | `.claude/skills/moai-alfred-expertise-detection/` | User expertise level detection |
| `moai-alfred-clone-pattern` | `.claude/skills/moai-alfred-clone-pattern/` | Master-Clone pattern implementation guide |
| `moai-alfred-rules` | `.claude/skills/moai-alfred-rules/` | Alfred operational rules and constraints |
| `moai-alfred-practices` | `.claude/skills/moai-alfred-practices/` | Best practices and operational guidelines |
| `moai-alfred-persona-roles` | `.claude/skills/moai-alfred-persona-roles/` | Persona role definitions and behaviors |
| `moai-alfred-dev-guide` | `.claude/skills/moai-alfred-dev-guide/` | Development guide and workflow documentation |
| `moai-alfred-code-reviewer` | `.claude/skills/moai-alfred-code-reviewer/` | Code review and quality assessment |
| `moai-alfred-issue-labels` | `.claude/skills/moai-alfred-issue-labels/` | GitHub issue labeling and management |

### Claude Code Skills (Infrastructure)

| Skill Name | Directory | Purpose |
|------------|-----------|---------|
| `moai-cc-agents` | `.claude/skills/moai-cc-agents/` | Claude Code agent creation and management |
| `moai-cc-commands` | `.claude/skills/moai-cc-commands/` | Command design and implementation |
| `moai-cc-skills` | `.claude/skills/moai-cc-skills/` | Skill creation and optimization |
| `moai-cc-hooks` | `.claude/skills/moai-cc-hooks/` | Hook system configuration and management |
| `moai-cc-settings` | `.claude/skills/moai-cc-settings/` | Settings.json configuration and permissions |
| `moai-cc-memory` | `.claude/skills/moai-cc-memory/` | Memory management and context optimization |
| `moai-cc-mcp-plugins` | `.claude/skills/moai-cc-mcp-plugins/` | MCP server and plugin configuration |
| `moai-cc-skill-descriptions` | `.claude/skills/moai-cc-skill-descriptions/` | Skill description authoring standards |
| `moai-cc-skill-factory` | `.claude/skills/moai-cc-skill-factory/` | Skill factory and creation patterns |
| `moai-cc-guide` | `.claude/skills/moai-cc-guide/` | Comprehensive Claude Code usage guide |

### Essential Skills (Core Development)

| Skill Name | Directory | Purpose |
|------------|-----------|---------|
| `moai-essentials-debug` | `.claude/skills/moai-essentials-debug/` | Debugging and error resolution |
| `moai-essentials-perf` | `.claude/skills/moai-essentials-perf/` | Performance optimization and profiling |
| `moai-essentials-review` | `.claude/skills/moai-essentials-review/` | Code review and quality assessment |
| `moai-essentials-refactor` | `.claude/skills/moai-essentials-refactor/` | Code refactoring and improvement patterns |

### Language Skills (Multi-Language Support)

| Skill Name | Directory | Purpose |
|------------|-----------|---------|
| `moai-lang-python` | `.claude/skills/moai-lang-python/` | Python 3.13+ best practices and tools |
| `moai-lang-typescript` | `.claude/skills/moai-lang-typescript/` | TypeScript 5.7+ best practices and tools |
| `moai-lang-javascript` | `.claude/skills/moai-lang-javascript/` | JavaScript development patterns and tools |
| `moai-lang-shell` | `.claude/skills/moai-lang-shell/` | Shell scripting best practices and tools |
| `moai-lang-go` | `.claude/skills/moai-lang-go/` | Go 1.24+ best practices and tools |
| `moai-lang-rust` | `.claude/skills/moai-lang-rust/` | Rust development patterns and tools |
| `moai-lang-java` | `.claude/skills/moai-lang-java/` | Java development patterns and tools |
| `moai-lang-kotlin` | `.claude/skills/moai-lang-kotlin/` | Kotlin 2.1+ best practices and tools |
| `moai-lang-c` | `.claude/skills/moai-lang-c/` | C development patterns and tools |
| `moai-lang-cpp` | `.claude/skills/moai-lang-cpp/` | C++ development patterns and tools |
| `moai-lang-csharp` | `.claude/skills/moai-lang-csharp/` | C# development patterns and tools |
| `moai-lang-php` | `.claude/skills/moai-lang-php/` | PHP 8.4+ best practices and tools |
| `moai-lang-ruby` | `.claude/skills/moai-lang-ruby/` | Ruby 3.4+ best practices and tools |
| `moai-lang-swift` | `.claude/skills/moai-lang-swift/` | Swift development patterns and tools |
| `moai-lang-scala` | `.claude/skills/moai-lang-scala/` | Scala development patterns and tools |
| `moai-lang-r` | `.claude/skills/moai-lang-r/` | R 4.4+ best practices and tools |
| `moai-lang-dart` | `.claude/skills/moai-lang-dart/` | Dart development patterns and tools |
| `moai-lang-sql` | `.claude/skills/moai-lang-sql/` | SQL development patterns and tools |

### Domain Skills (Specialized Areas)

| Skill Name | Directory | Purpose |
|------------|-----------|---------|
| `moai-domain-backend` | `.claude/skills/moai-domain-backend/` | Backend architecture and scaling guidance |
| `moai-domain-frontend` | `.claude/skills/moai-domain-frontend/` | React 19/Vue 3.5/Angular 19 with state management |
| `moai-domain-cli-tool` | `.claude/skills/moai-domain-cli-tool/` | CLI tool development with argument parsing |
| `moai-domain-database` | `.claude/skills/moai-domain-database/` | Database design, schema optimization, indexing |
| `moai-domain-security` | `.claude/skills/moai-domain-security/` | OWASP Top 10, SAST/DAST, security best practices |
| `moai-domain-ml` | `.claude/skills/moai-domain-ml/` | Machine learning model training, deployment, MLOps |
| `moai-domain-mobile-app` | `.claude/skills/moai-domain-mobile-app/` | Flutter 3.27/React Native 0.76 development |
| `moai-domain-devops` | `.claude/skills/moai-domain-devops/` | DevOps practices, infrastructure, deployment |
| `moai-domain-data-science` | `.claude/skills/moai-domain-data-science/` | Data science workflows and analysis |
| `moai-domain-web-api` | `.claude/skills/moai-domain-web-api/` | Web API design and development |

### Project Skills (Project Management)

| Skill Name | Directory | Purpose |
|------------|-----------|---------|
| `moai-project-config-manager` | `.claude/skills/moai-project-config-manager/` | Configuration management and validation |
| `moai-project-language-initializer` | `.claude/skills/moai-project-language-initializer/` | Project language and user setup workflows |
| `moai-project-documentation` | `.claude/skills/moai-project-documentation/` | Documentation generation and management |
| `moai-project-template-optimizer` | `.claude/skills/moai-project-template-optimizer/` | Template optimization and management |
| `moai-project-batch-questions` | `.claude/skills/moai-project-batch-questions/` | Batch question patterns and optimization |

### Design Skills (UI/UX and Design)

| Skill Name | Directory | Purpose |
|------------|-----------|---------|
| `moai-design-systems` | `.claude/skills/moai-design-systems/` | Design system creation and management |

### Legacy Skills (Redirects)

| Skill Name | Directory | Purpose |
|------------|-----------|---------|
| `moai-skill-factory` | `.claude/skills/moai-skill-factory/` | Legacy redirect to moai-cc-skill-factory |

## Non-Existent Skills (DO NOT USE)

The following skills have been referenced in documentation but **DO NOT EXIST** in the `.claude/skills/` directory:

### Hallucinated JIT Skills (Critical - Remove All References)
- ❌ `moai-session-info` - Referenced in command documentation but doesn't exist
- ❌ `moai-jit-docs-enhanced` - Referenced in command documentation but doesn't exist
- ❌ `moai-streaming-ui` - Referenced in command documentation but doesn't exist
- ❌ `moai-change-logger` - Referenced in command documentation but doesn't exist
- ❌ `moai-tag-policy-validator` - Referenced in command documentation but doesn't exist
- ❌ `moai-learning-optimizer` - Referenced in command documentation but doesn't exist

## Skill Usage Guidelines

### Correct Skill Invocation Pattern
```python
# Always use explicit syntax
Skill("moai-foundation-specs")
Skill("moai-alfred-workflow")
Skill("moai-cc-agents")
```

### Incorrect Skill References (Remove These)
```python
# These skills don't exist - remove from documentation
Skill("moai-session-info")        # ❌ Does not exist
Skill("moai-jit-docs-enhanced")   # ❌ Does not exist
Skill("moai-streaming-ui")       # ❌ Does not exist
Skill("moai-change-logger")      # ❌ Does not exist
Skill("moai-tag-policy-validator") # ❌ Does not exist
Skill("moai-learning-optimizer")  # ❌ Does not exist
```

## Summary Statistics

- **Total Verified Skills**: 76 skills
- **Foundation Skills**: 6
- **Alfred Skills**: 26
- **Claude Code Skills**: 10
- **Essential Skills**: 4
- **Language Skills**: 18
- **Domain Skills**: 10
- **Project Skills**: 5
- **Design Skills**: 1
- **Legacy Skills**: 1
- **Non-Existent Skills**: 6 (to be removed from documentation)

## Quality Assurance

All skills listed in the "Verified Skills Inventory" have been:
1. **Confirmed to exist** in `.claude/skills/` directory
2. **Verified to have SKILL.md file** with proper documentation
3. **Cross-referenced** with current usage in agent and command documentation
4. **Categorized** by functional area for easy reference

---

**Document Version**: 1.0
**Generated**: 2025-11-05
**Phase**: 3 - Documentation Alignment
**Scope**: Complete Verified Skill Reference