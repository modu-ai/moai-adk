# SPEC-REFACTOR-001: MoAI-ADK Template Comprehensive Refactoring

## Overview

Based on analysis of Claude Code official documentation (14 docs + 12 engineering blogs), this SPEC defines the comprehensive refactoring plan for all MoAI-ADK templates.

## Current State Analysis

### Template Inventory

| Category | Count | Location |
|----------|-------|----------|
| Agents | 27 | .claude/agents/moai/ |
| Skills | 47+ | .claude/skills/ |
| Commands | 5 | .claude/commands/moai/ |
| Hooks | 5 | .claude/hooks/moai/ |
| Output Styles | 2 | .claude/output-styles/moai/ |
| Config Files | 16 | .moai/config/ |

### Updated Reference Documents (v4.0.0)

moai-foundation-claude skill now includes:
- claude-code-skills-official.md (Progressive Disclosure 3-level)
- claude-code-sub-agents-official.md (/agents command, resumable)
- claude-code-plugins-official.md (NEW)
- claude-code-sandboxing-official.md (NEW)
- claude-code-headless-official.md (NEW)
- claude-code-statusline-official.md (NEW)
- claude-code-devcontainers-official.md (NEW)
- claude-code-cli-reference-official.md (NEW)
- advanced-agent-patterns.md (NEW - from engineering blogs)

## Refactoring Phases

### Phase 2: CLAUDE.md Refactoring

Owner: manager-docs subagent

Tasks:
- Update version to 9.0.0
- Add new sections for Plugins, Sandboxing, Headless
- Update Agent Invocation Patterns with /agents command
- Add Resumable Agents section
- Update Tool Access with new categories
- Add Advanced Agent Patterns from engineering blogs
- Optimize token usage (target: under 700 lines)

### Phase 3: Agents Definition Update (27 agents)

Owner: builder-agent subagent

Priority Groups:
1. Core Managers (7): manager-tdd, manager-spec, manager-git, manager-quality, manager-project, manager-docs, manager-strategy, manager-claude-code
2. Domain Experts (8): expert-backend, expert-frontend, expert-database, expert-security, expert-testing, expert-performance, expert-debug, expert-devops, expert-uiux
3. Builders (4): builder-agent, builder-skill, builder-command, builder-plugin
4. MCP Integrations (5): mcp-context7, mcp-playwright, mcp-figma, mcp-notion, mcp-sequential-thinking
5. Specialized (3): ai-nano-banana

Update Requirements:
- Add YAML frontmatter format validation
- Include "use PROACTIVELY" in descriptions for auto-delegation
- Verify tool restrictions follow least-privilege principle
- Update model selection (sonnet/opus/haiku/inherit)
- Remove deprecated permissionMode values
- Add skills field where applicable

### Phase 4: Skills Validation and Update (47+ skills)

Owner: builder-skill subagent

Validation Checklist:
- SKILL.md under 500 lines
- Description in third person with trigger terms
- Progressive disclosure with reference files
- No deeply nested references (one level from SKILL.md)
- Consistent naming (gerund form preferred)

Priority Skills for Update:
1. moai-foundation-core (align with v4.0.0 claude patterns)
2. moai-foundation-philosopher (integrate with CLAUDE.md)
3. moai-workflow-spec (EARS format alignment)
4. moai-workflow-project (config integration)
5. moai-plugin-builder (align with new plugins guide)

### Phase 5: Commands Update (5 commands)

Owner: builder-command subagent

Commands:
- /moai:0-project - Project configuration
- /moai:1-plan - SPEC generation
- /moai:2-run - TDD implementation
- /moai:3-sync - Documentation sync
- /moai:9-feedback - Feedback submission

Update Requirements:
- Verify $ARGUMENTS, $1, $2 parameter usage
- Check @file references
- Update AskUserQuestion usage (max 4 options, no emoji)
- Ensure agent delegation follows new patterns

### Phase 6: Hooks Validation and Optimization (5 hooks)

Owner: expert-backend subagent

Hooks:
- session_start__show_project_info.py
- session_end__auto_cleanup.py
- pre_tool__security_guard.py
- post_tool__code_formatter.py
- post_tool__linter.py

Validation:
- Check timeout settings (PreToolUse/PostToolUse events)
- Verify error handling and fallback
- Optimize performance
- Update lib/ dependencies

### Phase 7: settings.json Optimization

Owner: manager-claude-code subagent

Updates:
- Verify hooks configuration format
- Update permissions structure
- Check statusLine configuration
- Verify env variables
- Update companyAnnouncements for v4.0.0

### Phase 8: Output Styles Update (2 styles)

Owner: manager-docs subagent

Styles:
- r2d2.md - Pair programming partner
- yoda.md - Technical wisdom master

Updates:
- Align with AskUserQuestion constraints
- Update version references
- Verify language handling

### Phase 9: .moai/config Structure Review

Owner: manager-project subagent

Files:
- config.yaml - Main configuration
- statusline-config.yaml - Statusline settings
- sections/*.yaml - Modular config sections
- questions/*.yaml - AskUserQuestion templates

Validation:
- Schema consistency
- Default value verification
- Template variable placeholders

### Phase 10: Template Synchronization

Owner: manager-git subagent

Tasks:
- Sync templates to local .claude/ directory
- Validate all changes
- Create comprehensive changelog
- Prepare release notes

## Success Criteria

1. All agents follow new YAML frontmatter format
2. All skills under 500 lines with progressive disclosure
3. Commands properly delegate to agents
4. Hooks optimized with proper timeouts
5. settings.json uses correct permission structure
6. CLAUDE.md under 700 lines
7. All templates sync successfully

## Timeline

Phase 2-3: Agents and CLAUDE.md (parallel)
Phase 4: Skills validation (bulk processing)
Phase 5-6: Commands and Hooks (parallel)
Phase 7-9: Configuration (sequential)
Phase 10: Synchronization and validation

## Version

SPEC Version: 1.0.0
Target MoAI-ADK Version: 0.37.0
Created: 2026-01-06
