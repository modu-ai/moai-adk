# Verified Skill Inventory

**Phase 1 Documentation Audit - Extracted from Official Sources Only**

## Overview

This document provides a complete inventory of all verified skills that exist in the MoAI-ADK codebase, extracted exclusively from official sources. Only skills that have corresponding `.claude/skills/*/SKILL.md` files are included.

## Skill Categories

### Foundation Skills (Tier 1)

| Skill Name | Description | Purpose |
|------------|-------------|---------|
| `moai-foundation-specs` | SPEC YAML frontmatter validation (7 required fields) and HISTORY section enforcement | Validates SPEC document structure and metadata |
| `moai-foundation-ears` | EARS requirement authoring guide (Ubiquitous/Event-driven/State-driven/Optional/Unwanted Behaviors) with 5 official patterns | Provides EARS syntax guidance for requirement writing |
| `moai-foundation-tags` | TAG inventory management and orphan detection (CODE-FIRST principle) | Manages @TAG system integrity and traceability |
| `moai-foundation-trust` | TRUST 5-principles validation (Test First, Readable, Unified, Secured, Trackable) | Enforces quality standards and TRUST principles |
| `moai-foundation-langs` | Auto-detects project language from package.json, pyproject.toml, etc. | Detects project programming language automatically |
| `moai-foundation-git` | Git operations and version control guidance | Provides Git workflow guidance and best practices |

### Alfred Skills (Tier 2 - Workflow Orchestration)

| Skill Name | Description | Purpose |
|------------|-------------|---------|
| `moai-alfred-workflow` | Guide 4-step workflow execution with task tracking and quality gates | Orchestrates Alfred's 4-step workflow logic |
| `moai-alfred-ask-user-questions` | Guide Alfred sub-agents to actively invoke AskUserQuestion for ambiguous decisions | Enables interactive user clarification and decision making |
| `moai-alfred-agent-guide` | 19-agent team structure, decision trees for agent selection, Haiku vs Sonnet model selection, and agent collaboration principles | Provides agent selection guidance and collaboration patterns |
| `moai-alfred-personas` | Adaptive communication patterns and role selection based on user expertise level and request type | Manages Alfred's adaptive persona system |
| `moai-alfred-reporting` | Report generation standards, output formatting rules, and sub-agent report examples | Establishes reporting style guidelines |
| `moai-alfred-language-detection` | Auto-detects project language and framework from package.json, pyproject.toml, etc. | Detects project language and framework automatically |
| `moai-alfred-session-state` | Session state management, runtime state tracking, session handoff notes | Manages session state and handoff procedures |
| `moai-alfred-autofixes` | Safety protocol for automatic code fixes, merge conflicts, and user approval workflow | Provides automatic fix protocols and safety measures |
| `moai-alfred-trust-validation` | TRUST 5-principles validation (Test 85%+, Readable, Unified, Secured, Trackable) | Validates TRUST principle compliance |
| `moai-alfred-todowrite-pattern` | TodoWrite auto-initialization patterns from Plan agent, task tracking best practices, status management | Manages TodoWrite initialization and patterns |
| `moai-alfred-spec-metadata-validation` | SPEC YAML frontmatter validation with 7 required fields and HISTORY section compliance | Validates SPEC metadata and structure |
| `moai-alfred-spec-authoring` | SPEC document authoring guide - YAML metadata, EARS syntax, HISTORY section, and validation | Provides SPEC document creation guidance |
| `moai-alfred-spec-metadata-extended` | Standards for SPEC authoring including YAML metadata, EARS syntax, HISTORY section, and validation | Extended SPEC authoring standards |
| `moai-alfred-doc-management` | Internal documentation placement rules, forbidden patterns, and sub-agent output guidelines | Manages documentation placement and policies |
| `moai-alfred-expertise-detection` | Guide Alfred to detect user expertise level (Beginner/Intermediate/Expert) through in-session behavioral signals without memory file access | Detects user expertise level for adaptive responses |
| `moai-alfred-tag-scanning` | Scans @TAG markers from code and generates TAG inventory (CODE-FIRST principle) | Scans and manages TAG inventory |
| `moai-alfred-clone-pattern` | Master-Clone pattern implementation guide for complex multi-step tasks with full project context | Implements Master-Clone pattern for complex tasks |
| `moai-alfred-proactive-suggestions` | Guide Alfred to provide non-intrusive proactive suggestions based on risk detection, optimization patterns, and learning opportunities | Provides proactive suggestion system |
| `moai-alfred-context-budget` | Claude Code context window optimization strategies, JIT retrieval, progressive loading, memory file patterns, and cleanup practices | Manages context budget optimization |
| `moai-alfred-gitflow-policy` | Enforces MoAI-ADK GitFlow workflow for team and personal modes. Covers branch protection, PR creation rules, release process, and conflict resolution | Enforces GitFlow workflow policies |
| `moai-alfred-ears-authoring` | EARS (Easy Approach to Requirements Syntax) authoring with 5 statement patterns for clear requirements | Provides EARS authoring guidance |
| `moai-alfred-issue-labels` | GitHub issue labeling automation and management for MoAI-ADK workflow | Manages GitHub issue labeling |
| `moai-alfred-config-schema` | .moai/config.json official schema documentation, structure validation, project metadata, language settings, and configuration migration guide | Provides configuration schema guidance |
| `moai-alfred-practices` | Practical workflow examples, context engineering strategy, agent collaboration principles, and model selection guide | Provides practical workflow examples |
| `moai-alfred-rules` | Skill invocation rules, interactive question rules, Git commit message standard, @TAG lifecycle, TRUST 5 principles | Establishes core rules and standards |
| `moai-alfred-dev-guide` | Development guide for MoAI-ADK project setup, workflow understanding, and best practices | Provides development guidance |
| `moai-alfred-code-reviewer` | Code review guidance and standards for MoAI-ADK projects | Provides code review standards |

### Claude Code Configuration Skills (Tier 3 - Infrastructure)

| Skill Name | Description | Purpose |
|------------|-------------|---------|
| `moai-cc-agents` | Agent creation and management guidance for Claude Code | Provides agent creation guidance |
| `moai-cc-commands` | Command design and implementation for Claude Code | Provides command design guidance |
| `moai-cc-skills` | Creating and Optimizing Claude Code Skills. Design reusable knowledge capsules with progressive disclosure (metadata → content → resources). Apply freedom levels (high/medium/low), create examples, validate YAML | Manages skill creation and optimization |
| `moai-cc-skill-descriptions` | Skill description authoring standards, metadata format, freedom levels, and progressive disclosure patterns for Claude Code Skill development | Provides skill description standards |
| `moai-cc-skill-factory` | Create and maintain high-quality Claude Code Skills through interactive discovery, web research, and continuous updates | Factory for creating high-quality skills |
| `moai-cc-hooks` | Configuring Claude Code Hooks System. Design, implement, and manage PreToolUse/PostToolUse/SessionStart/Notification/Stop hooks | Manages hook system configuration |
| `moai-cc-settings` | Configuring Claude Code settings.json & Security. Set up permissions (allow/deny), permission modes, environment variables, tool restrictions | Provides settings configuration guidance |
| `moai-cc-memory` | Managing Claude Code Session Memory & Context. Understand session context limits, use just-in-time retrieval, cache insights, manage memory files | Manages session memory and context |
| `moai-cc-mcp-plugins` | Configuring MCP Servers & Plugins for Claude Code. Set up Model Context Protocol servers (GitHub, Filesystem, Brave Search, SQLite). Configure OAuth, manage permissions, validate MCP structure | Manages MCP plugin configuration |
| `moai-cc-guide` | Architecture decisions, workflow guidance, and roadmaps for Claude Code setup and usage | Provides architecture and workflow guidance |

### Domain-Specific Skills (Tier 4 - Technical Expertise)

| Skill Name | Description | Purpose |
|------------|-------------|---------|
| `moai-lang-python` | Python 3.13+ best practices with pytest 8.4.2, mypy 1.8.0, ruff 0.13.1, uv 0.9.3, and backend frameworks (FastAPI, Flask, Django) | Python language and ecosystem guidance |
| `moai-lang-shell` | Shell scripting best practices with bats-core 1.11, shellcheck 0.10, and POSIX compliance | Shell scripting guidance |
| `moai-lang-typescript` | TypeScript 5.7+ best practices with Vitest 2.1, Biome 1.9, strict typing, npm/pnpm/bun package management, and fullstack meta-frameworks (Next.js, Remix, etc.) | TypeScript language and ecosystem guidance |
| `moai-lang-go` | Go 1.24+ best practices with go test, golangci-lint, gofmt, standard library utilization, and web frameworks (Gin, Beego) | Go language and ecosystem guidance |
| `moai-lang-rust` | Rust best practices with cargo, rustfmt, clippy, and ecosystem patterns | Rust language and ecosystem guidance |
| `moai-lang-php` | PHP 8.4+ best practices with PHPUnit 11, Composer, PSR-12 standards, and web frameworks (Laravel, Symfony) | PHP language and ecosystem guidance |
| `moai-lang-kotlin` | Kotlin 2.1+ best practices with JUnit 5, Gradle, ktlint, coroutines, and extension functions | Kotlin language and ecosystem guidance |
| `moai-lang-ruby` | Ruby 3.4+ best practices with RSpec 4, RuboCop 2, Bundler, and Rails 8 patterns | Ruby language and ecosystem guidance |
| `moai-lang-r` | R 4.4+ best practices with testthat 3.2, lintr 3.2, and data analysis patterns | R language and ecosystem guidance |
| `moai-lang-javascript` | JavaScript best practices with modern ES2025+, testing frameworks, and ecosystem patterns | JavaScript language and ecosystem guidance |
| `moai-lang-java` | Java best practices with modern versions, testing frameworks, and ecosystem patterns | Java language and ecosystem guidance |
| `moai-lang-c` | C language best practices with modern standards, testing frameworks, and ecosystem patterns | C language and ecosystem guidance |
| `moai-lang-cpp` | C++ best practices with modern standards, testing frameworks, and ecosystem patterns | C++ language and ecosystem guidance |
| `moai-lang-csharp` | C# best practices with modern .NET, testing frameworks, and ecosystem patterns | C# language and ecosystem guidance |
| `moai-lang-scala` | Scala best practices with modern versions, testing frameworks, and ecosystem patterns | Scala language and ecosystem guidance |
| `moai-lang-swift` | Swift best practices with modern versions, testing frameworks, and ecosystem patterns | Swift language and ecosystem guidance |
| `moai-lang-dart` | Dart best practices with modern versions, testing frameworks, and ecosystem patterns | Dart language and ecosystem guidance |
| `moai-lang-sql` | SQL best practices across different databases and optimization patterns | SQL language and optimization guidance |
| `moai-domain-database` | Database design, schema optimization, indexing strategies, and migration management | Database design and optimization guidance |
| `moai-domain-cli-tool` | CLI tool development with argument parsing, POSIX compliance, and user-friendly help messages | CLI tool development guidance |
| `moai-domain-backend` | Provides backend architecture and scaling guidance; use when the project targets server-side APIs or infrastructure design decisions | Backend architecture guidance |
| `moai-domain-frontend` | React 19/Vue 3.5/Angular 19 with state management, performance optimization, accessibility, and meta-frameworks (Nuxt, SvelteKit, Astro, SolidJS) | Frontend development guidance |
| `moai-domain-security` | OWASP Top 10, SAST/DAST, dependency security, and secrets management | Security best practices guidance |
| `moai-domain-ml` | Machine learning model training, evaluation, deployment, and MLOps workflows | Machine learning workflow guidance |
| `moai-domain-mobile-app` | Flutter 3.27/React Native 0.76 with state management and native integration | Mobile app development guidance |
| `moai-domain-web-api` | REST API, GraphQL, and web service design patterns | Web API design guidance |
| `moai-domain-data-science` | Data analysis, visualization, and scientific computing patterns | Data science workflow guidance |
| `moai-domain-devops` | Deployment, CI/CD, infrastructure as code, and operations patterns | DevOps and deployment guidance |

### Project Management Skills (Tier 5 - Project Operations)

| Skill Name | Description | Purpose |
|------------|-------------|---------|
| `moai-project-config-manager` | Complete config.json CRUD operations with validation, merge strategy, and error recovery | Manages project configuration |
| `moai-project-language-initializer` | Handle comprehensive project language and user setup workflows including language selection, agent prompt configuration, user profiles, team settings, and domain selection | Handles project initialization workflows |
| `moai-project-batch-questions` | Standardize AskUserQuestion patterns and provide reusable question templates for batch optimization | Provides question templates and patterns |
| `moai-project-template-optimizer` | Optimize project templates for better performance and user experience | Optimizes project templates |
| `moai-project-documentation` | Project documentation standards and management | Manages project documentation standards |

### Essential Skills (Tier 6 - Core Practices)

| Skill Name | Description | Purpose |
|------------|-------------|---------|
| `moai-essentials-refactor` | Refactoring guidance with design patterns and code improvement strategies | Provides refactoring guidance |
| `moai-essentials-debug` | Advanced debugging with stack trace analysis, error pattern detection, and fix suggestions | Provides debugging strategies |
| `moai-essentials-perf` | Performance optimization with profiling, bottleneck detection, and tuning strategies | Provides performance optimization guidance |
| `moai-essentials-review` | Code review standards and best practices | Provides code review standards |

### Design and Specialized Skills (Tier 7 - Specialized Areas)

| Skill Name | Description | Purpose |
|------------|-------------|---------|
| `moai-design-systems` | Design system creation and management guidance | Design system development guidance |
| `moai-skill-factory` | Legacy redirect to moai-cc-skill-factory. Creating and Optimizing Claude Code Skills with progressive disclosure, freedom levels, and validation | Legacy skill factory redirect |

## Verified Skill Count

- **Total Skills**: 79
- **Foundation Skills**: 6
- **Alfred Skills**: 27
- **Claude Code Configuration Skills**: 9
- **Domain-Specific Skills**: 25
- **Project Management Skills**: 5
- **Essential Skills**: 4
- **Design and Specialized Skills**: 2
- **Legacy Redirects**: 1

## Important Notes

1. **Only Verified Skills**: This inventory only includes skills that exist in the `.claude/skills/` directory with corresponding `SKILL.md` files.

2. **No JIT Skills**: Skills mentioned in command files but not existing in the codebase (such as `moai-session-info`, `moai-jit-docs-enhanced`, `moai-streaming-ui`, etc.) are excluded as they don't exist.

3. **Categorized by Tier**: Skills are organized by their functional tier and purpose within the MoAI-ADK architecture.

4. **Verified from Official Sources**: All skill information is extracted from actual skill files in the repository.

---

**Generated**: 2025-11-05
**Source**: Official `.claude/skills/*/SKILL.md` files
**Audit Scope**: Phase 1 Documentation Audit