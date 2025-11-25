# MoAI-ADK Agents Reference

Alfred's agent delegation reference. Each agent follows the `{domain}-{role}` naming convention and is optimized for specific tasks.

## Naming Convention: `{domain}-{role}`

All MoAI-ADK agents follow a consistent naming pattern:

| Domain | Purpose | Examples |
|--------|---------|----------|
| `workflow` | Core workflow command processors | workflow-spec, workflow-tdd |
| `core` | Orchestration & quality management | core-planner, core-quality |
| `code` | Code implementation experts | code-backend, code-frontend |
| `data` | Data-related experts | data-database |
| `infra` | Infrastructure/DevOps experts | infra-devops |
| `design` | Design/UX experts | design-uiux |
| `security` | Security experts | security-expert |
| `mcp` | MCP server integrations | mcp-context7, mcp-ultrathink |
| `factory` | Meta-generation agents | factory-agent, factory-skill |
| `support` | Support services | support-debug, support-claude |
| `ai` | AI model integrations | ai-codex, ai-gemini |

---

## Tier 1: Command Processors (Essential - Always Active)

Core command processors directly bound to MoAI commands.

| Agent | Command | Purpose |
|-------|---------|---------|
| `workflow-project` | `/moai:0-project` | Project initialization and setup |
| `workflow-spec` | `/moai:1-plan` | EARS SPEC generation and planning |
| `workflow-tdd` | `/moai:2-run` | TDD RED-GREEN-REFACTOR execution |
| `workflow-docs` | `/moai:3-sync` | Documentation generation and synchronization *(merged: doc-syncer)* |

---

## Tier 2: Orchestration & Quality (Auto-triggered)

Orchestration and quality management agents.

| Agent | Trigger | Purpose |
|-------|---------|---------|
| `core-planner` | `/moai:2-run` Phase 1 | SPEC analysis and execution strategy |
| `core-quality` | Post-implementation | TRUST 5 validation *(merged: trust-checker)* |
| `core-git` | Git operations | Branch, commit, and PR management |

---

## Tier 3: Domain Experts (Lazy-loaded)

Domain-specific implementation experts.

| Agent | Domain | Purpose |
|-------|--------|---------|
| `code-backend` | Backend | Backend architecture and API design *(merged: api-designer)* |
| `code-frontend` | Frontend | Frontend UI/UX implementation |
| `data-database` | Data | Database schema design and migration *(merged: migration-expert)* |
| `infra-devops` | Infrastructure | DevOps, monitoring, and performance *(merged: monitoring-expert, performance-engineer)* |
| `security-expert` | Security | Security analysis and OWASP validation |
| `design-uiux` | Design | UI/UX, components, and accessibility *(merged: component-designer, accessibility-expert)* |

---

## Tier 4: MCP Integrators (Resume-enabled)

External MCP server integrations with context continuity support.

| Agent | MCP Server | Purpose |
|-------|------------|---------|
| `mcp-context7` | Context7 | Documentation research and API reference |
| `mcp-figma` | Figma | Design system integration |
| `mcp-notion` | Notion | Knowledge base integration |
| `mcp-playwright` | Playwright | Browser automation and E2E testing |
| `mcp-ultrathink` | Sequential-Thinking | Complex reasoning and strategic analysis |

**Resume Pattern**:
All MCP integrators support agent resume for context continuity:
```python
# Initial call
result = Task(subagent_type="mcp-context7", prompt="...")
agent_id = result.agent_id

# Resume with context
result = Task(subagent_type="mcp-context7", prompt="Continue...", resume=agent_id)
```

---

## Tier 5: Factory Agents (Meta-development)

Meta-generation agents for MoAI-ADK development.

| Agent | Purpose |
|-------|---------|
| `factory-agent` | New agent creation and configuration |
| `factory-skill` | Skill definition creation and management |
| `factory-command` | Custom slash command creation and optimization |

---

## Tier 6: Support (On-demand)

Support and utility services.

| Agent | Purpose |
|-------|---------|
| `support-debug` | Error analysis and diagnostic support |
| `support-claude` | Claude Code configuration management |

---

## Tier 7: AI & Specialized

AI model integrations and specialized services.

| Agent | Purpose |
|-------|---------|
| `ai-codex` | OpenAI Codex CLI integration |
| `ai-gemini` | Google Gemini API integration |
| `ai-banana` | Gemini 3 image generation |

---

## System Agents

Built-in system agents for codebase exploration.

| Agent | Purpose |
|-------|---------|
| `Explore` | Codebase exploration and file system analysis |
| `Plan` | Strategic decomposition and planning |

---

## Delegation Principles

1. **Agent-First**: Alfred always delegates to specialized agents via Task()
2. **Naming Consistency**: All agents follow `{domain}-{role}` pattern
3. **Context Passing**: Pass each agent's results as context to the next agent
4. **Sequential vs Parallel**: Analyze dependencies to determine execution strategy

## Agent Selection Criteria

| Task Complexity | Files | Agents | Strategy |
|----------------|-------|--------|----------|
| Simple | 1 file | 1-2 agents | Sequential |
| Medium | 3-5 files | 2-3 agents | Sequential |
| Complex | 10+ files | 5+ agents | Mixed parallel/sequential |

---

## Merged Agents (Historical Reference)

The following agents were merged to reduce complexity:

| Old Agent | Merged Into | Reason |
|-----------|-------------|--------|
| doc-syncer | workflow-docs | Documentation consolidation |
| trust-checker | core-quality | Quality gate unification |
| api-designer | code-backend | Backend expertise consolidation |
| migration-expert | data-database | Data operations unification |
| monitoring-expert | infra-devops | Infrastructure consolidation |
| performance-engineer | infra-devops | Infrastructure consolidation |
| component-designer | design-uiux | Design system unification |
| accessibility-expert | design-uiux | Design system unification |

## Removed Agents

The following agents were removed:

| Agent | Reason |
|-------|--------|
| format-expert | Replaced by direct linter usage (ruff, prettier) |
| sync-manager | Redundant with workflow-docs |

---

## Skill Consolidation Reference

The following legacy skills have been consolidated into unified skills:

| Legacy Skills (Removed) | Unified Skill (Current) | Reason |
|------------------------|------------------------|--------|
| moai-foundation-specs, moai-foundation-ears, moai-foundation-trust, moai-foundation-git, moai-foundation-langs | moai-foundation-core | Core principles consolidation |
| moai-lang-python, moai-lang-typescript, moai-lang-sql | moai-lang-unified | Language unification |
| moai-essentials-debug, moai-essentials-perf, moai-essentials-refactor | moai-essentials-unified | Development tools unification |
| moai-cc-claude-md, moai-cc-configuration, moai-cc-hooks, moai-cc-claude-settings | moai-core-claude-code | Claude Code features consolidation |
| moai-domain-backend, moai-domain-frontend | moai-lang-unified | Domain expertise integration |
| moai-domain-database, moai-domain-devops | moai-baas-unified | Infrastructure consolidation |
| moai-domain-security, moai-security-owasp | moai-universal-ultimate | Security consolidation |
| moai-core-spec-authoring, moai-core-todowrite-pattern | moai-foundation-core | Core workflow patterns |
| moai-core-context-budget | moai-context-manager | Token budget management |
| moai-quality-validation | moai-core-quality | Quality gate consolidation |

**Note**: All agent_skills_mapping references have been updated to use unified skills. Legacy skill names are no longer valid.

---

**Total Agents**: 26 (down from 35, -26% reduction)

Refer to CLAUDE.md for detailed agent descriptions and usage guidelines.
