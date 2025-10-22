---
name: "Claude Code Components Guide for MoAI-ADK"
description: "Navigate Claude Code architecture, map components to MoAI-ADK workflows, make decisions about Agents vs Commands vs Skills vs Hooks vs Plugins. Use when setting up Claude Code for MoAI-ADK projects or understanding which component to use for your task."
allowed-tools: "Read, Glob, Grep"
---

# Claude Code Components Guide for MoAI-ADK

Master Claude Code's four-layer architecture and integrate it seamlessly with MoAI-ADK's SPEC → TDD → Sync workflow.

## 🎯 Quick Decision Tree

**What do you need to do?**

```
┌─ Start Here ─────────────────────────────────────────┐
│                                                       │
├─ "Run workflow/automation at command level"         │
│  → Use: /alfred:* Commands                            │
│  → See: moai-cc-commands                              │
│                                                       │
├─ "Enforce safety checks, auto-format, run linters"  │
│  → Use: Hooks (PreToolUse, PostToolUse, SessionStart) │
│  → See: moai-cc-hooks                                 │
│                                                       │
├─ "Create specialized expert for analysis/debug"     │
│  → Use: Sub-agents                                    │
│  → See: moai-cc-agents                                │
│                                                       │
├─ "Restrict tool access, security settings"          │
│  → Use: settings.json                                 │
│  → See: moai-cc-settings                              │
│                                                       │
├─ "Integrate external APIs (GitHub, Filesystem)"     │
│  → Use: Plugins/MCP                                   │
│  → See: moai-cc-mcp-plugins                           │
│                                                       │
├─ "Share reusable domain knowledge/patterns"         │
│  → Use: Skills                                        │
│  → See: moai-cc-skills                                │
│                                                       │
├─ "Document project standards and AI guidance"       │
│  → Use: CLAUDE.md                                     │
│  → See: moai-cc-claude-md                             │
│                                                       │
└─ "Optimize context usage, manage session memory"    │
   → Use: Memory strategies                             │
   → See: moai-cc-memory                                │
```

---

## 📚 Seven Claude Code Components

### 1. **Commands** (`moai-cc-commands`)
**When**: Workflow orchestration, user-facing shortcuts, parameter passing
**Examples**: `/alfred:0-project`, `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync`
**Key features**: `$ARGUMENTS`, `$1`/`$2` positional args, Agent orchestration
**File location**: `.claude/commands/`

👉 **[Read Full Guide: moai-cc-commands](./../../skills/moai-cc-commands/SKILL.md)**

---

### 2. **Hooks** (`moai-cc-hooks`)
**When**: Enforce guardrails, auto-format code, run linters, trigger workflows
**Events**: PreToolUse, PostToolUse, SessionStart, SessionEnd, Stop, Notification
**Key features**: Exit codes (0=allow, 2=block), <100ms latency, script-based
**File location**: `.claude/settings.json` → hooks section

👉 **[Read Full Guide: moai-cc-hooks](./../../skills/moai-cc-hooks/SKILL.md)**

---

### 3. **Sub-agents** (`moai-cc-agents`)
**When**: Parallel work, specialized analysis, autonomous tasks, debugging
**Key features**: Independent context, tool restrictions, Proactive Triggers
**File location**: `.claude/agents/`

👉 **[Read Full Guide: moai-cc-agents](./../../skills/moai-cc-agents/SKILL.md)**

---

### 4. **Skills** (`moai-cc-skills`)
**When**: Reusable patterns, domain knowledge, team practices
**Key features**: Progressive Disclosure, allowed-tools, supporting files
**File location**: `.claude/skills/`

👉 **[Read Full Guide: moai-cc-skills](./../../skills/moai-cc-skills/SKILL.md)**

---

### 5. **settings.json** (`moai-cc-settings`)
**When**: Restrict tool access, manage permissions, configure MCP servers
**Key features**: allowedTools/deniedTools, permissionMode, hooks integration
**File location**: `.claude/settings.json`

👉 **[Read Full Guide: moai-cc-settings](./../../skills/moai-cc-settings/SKILL.md)**

---

### 6. **Plugins/MCP** (`moai-cc-mcp-plugins`)
**When**: External tool integration (GitHub APIs, Filesystem access, Web search)
**Key features**: OAuth configuration, marketplace, security restrictions
**File location**: `.claude/settings.json` → mcpServers section

👉 **[Read Full Guide: moai-cc-mcp-plugins](./../../skills/moai-cc-mcp-plugins/SKILL.md)**

---

### 7. **CLAUDE.md** (`moai-cc-claude-md`)
**When**: Document project standards, AI guidance, architecture patterns
**Key features**: 4-level hierarchy, imports (@path syntax), discovery
**File location**: Project root, ~/.claude/CLAUDE.md, or .claude/CLAUDE.md

👉 **[Read Full Guide: moai-cc-claude-md](./../../skills/moai-cc-claude-md/SKILL.md)**

---

### 8. **Memory Management** (`moai-cc-memory`)
**When**: Optimize context usage, handle large projects, implement caching
**Key features**: JIT retrieval, session context, layered strategies
**Managed by**: CLAUDE.md + Hook scripts

👉 **[Read Full Guide: moai-cc-memory](./../../skills/moai-cc-memory/SKILL.md)**

---

## 🎬 MoAI-ADK Workflow Integration

### `/alfred:0-project` — Project Initialization

**Claude Code Components Used:**
- **settings.json** (`moai-cc-settings`) - Create basic security permissions
- **CLAUDE.md** (`moai-cc-claude-md`) - Document project guidelines
- **Skills** (`moai-cc-skills`) - Set up domain-specific patterns
- **Hooks** (`moai-cc-hooks`) - SessionStart hook for project summary

**Typical setup:**
```bash
# 1. Initialize project metadata
/alfred:0-project

# 2. Configure Claude Code security
@agent-cc-manager "Set up project-level settings.json with TRUST permissions"

# 3. Create project CLAUDE.md
@agent-cc-manager "Create CLAUDE.md documenting SPEC-first principles"

# 4. Register custom Hooks
@agent-cc-manager "Configure SessionStart hook to show project status"
```

**Relevant guides:**
- 📖 [Project Setup Workflow](./workflows/alfred-0-project-setup.md)

---

### `/alfred:1-plan` — SPEC Authoring

**Claude Code Components Used:**
- **CLAUDE.md** (`moai-cc-claude-md`) - Reference EARS patterns, project standards
- **Skills** (`moai-cc-skills`) - EARS writing patterns (e.g., moai-foundation-ears)
- **Slash Commands** (`moai-cc-commands`) - `/alfred:1-plan` entry point

**Typical flow:**
```bash
# 1. Invoke planning command
/alfred:1-plan "JWT authentication system"

# 2. spec-builder agent loads:
#    - moai-foundation-specs Skill (SPEC metadata)
#    - moai-foundation-ears Skill (EARS patterns)
#    - CLAUDE.md for project context

# 3. Output: .moai/specs/SPEC-AUTH-001/spec.md
```

**Relevant guides:**
- 📖 [Planning Phase Workflow](./workflows/alfred-1-plan-flow.md)

---

### `/alfred:2-run` — TDD Implementation

**Claude Code Components Used:**
- **Hooks** (`moai-cc-hooks`) - PreToolUse (validate edits), PostToolUse (auto-format)
- **Sub-agents** (`moai-cc-agents`) - code-builder pipeline (RED → GREEN → REFACTOR)
- **settings.json** (`moai-cc-settings`) - Restrict edits to src/, tests/
- **Memory** (`moai-cc-memory`) - Cache test results across iterations
- **CLAUDE.md** - Reference TRUST 5 principles

**Typical flow:**
```bash
# 1. Start TDD loop
/alfred:2-run AUTH-001

# 2. PreToolUse Hook validates:
#    - Are edits in src/? tests/?
#    - Do edits include @TAG:AUTH-001?
#    - Follow formatting rules?

# 3. code-builder runs RED → GREEN → REFACTOR with:
#    - Test validation via pytest/npm test
#    - Auto-format via PostToolUse Hook
#    - TRUST 5 principle checks (trust-checker agent)

# 4. Output: Passing tests + formatted code + Git commits
```

**Relevant guides:**
- 📖 [Implementation Phase Workflow](./workflows/alfred-2-run-flow.md)

---

### `/alfred:3-sync` — Document Synchronization

**Claude Code Components Used:**
- **Sub-agents** (`moai-cc-agents`) - tag-agent (verify @TAG chains)
- **Skills** (`moai-cc-skills`) - moai-essentials-review (code review)
- **Hooks** (`moai-cc-hooks`) - PostToolUse (auto-generate docs)
- **Plugins/MCP** (`moai-cc-mcp-plugins`) - GitHub MCP (create PR)
- **Memory** (`moai-cc-memory`) - Aggregate changes for Living Docs

**Typical flow:**
```bash
# 1. Start sync
/alfred:3-sync

# 2. tag-agent verifies:
#    - @SPEC:AUTH-001 exists
#    - @TEST:AUTH-001 exists
#    - @CODE:AUTH-001 exists
#    - No orphan TAGs

# 3. doc-syncer generates:
#    - Living Docs from code
#    - README updates
#    - CHANGELOG entries

# 4. GitHub MCP creates:
#    - PR with all changes
#    - PR description from @TAG metadata

# 5. Output: Draft → Ready PR ready for review
```

**Relevant guides:**
- 📖 [Synchronization Phase Workflow](./workflows/alfred-3-sync-flow.md)

---

## 🛠️ Common Setup Tasks

### Task 1: Secure a Project
**Goal**: Restrict edits to specific paths, prevent dangerous commands

**Components to configure:**
1. `moai-cc-settings` - Set allowedTools/deniedTools
2. `moai-cc-hooks` - Add PreToolUse validation

**How-to:**
```bash
@agent-cc-manager "Configure .claude/settings.json to allow edits only in src/ and tests/"
@agent-cc-manager "Add PreToolUse hook to block rm -rf and sudo commands"
```

---

### Task 2: Add Custom Validation
**Goal**: Auto-format code, run linters, enforce TRUST 5

**Components to configure:**
1. `moai-cc-hooks` - Add PostToolUse hook
2. `moai-cc-settings` - Register hook in settings.json

**How-to:**
```bash
@agent-cc-manager "Create PostToolUse hook to auto-format Python code with black"
@agent-cc-manager "Create PreToolUse hook to validate @TAG presence in all edits"
```

---

### Task 3: Integrate GitHub
**Goal**: Auto-create PRs, manage issues

**Components to configure:**
1. `moai-cc-mcp-plugins` - Set up GitHub MCP
2. `moai-cc-settings` - Register MCP in settings.json
3. `moai-cc-commands` - Create `/alfred:3-sync` to trigger PR

**How-to:**
```bash
@agent-cc-manager "Configure GitHub MCP with GITHUB_TOKEN"
@agent-cc-manager "Update /alfred:3-sync to create PR via GitHub MCP"
```

---

### Task 4: Document Project Standards
**Goal**: Create CLAUDE.md with team guidelines, EARS patterns, architectural decisions

**Components to configure:**
1. `moai-cc-claude-md` - Create project CLAUDE.md
2. `moai-cc-skills` - Reference domain-specific patterns
3. `moai-cc-memory` - Implement layered imports

**How-to:**
```bash
@agent-cc-manager "Create CLAUDE.md documenting SPEC-first, EARS, and TRUST principles"
@agent-cc-manager "Add imports to Foundation tier Skills (moai-foundation-trust, moai-foundation-tags)"
```

---

## 📊 Component Interaction Matrix

| Phase | Command | Key Hooks | Required Skills | Agents | MCP Used |
|-------|---------|-----------|-----------------|--------|----------|
| Init | `0-project` | SessionStart | moai-foundation-langs | project-manager | None |
| Plan | `1-plan` | SessionStart | moai-foundation-specs, moai-foundation-ears | spec-builder | None |
| Run | `2-run` | PreToolUse, PostToolUse | moai-essentials-refactor, moai-foundation-trust | code-builder pipeline | None |
| Sync | `3-sync` | PostToolUse | moai-essentials-review, moai-foundation-tags | tag-agent, doc-syncer | GitHub MCP |

---

## 🔗 Complete Knowledge Base

**All 8 Components:**
1. 📖 Commands: [`moai-cc-commands`](./../../skills/moai-cc-commands/SKILL.md)
2. 🔐 Hooks: [`moai-cc-hooks`](./../../skills/moai-cc-hooks/SKILL.md)
3. 👥 Sub-agents: [`moai-cc-agents`](./../../skills/moai-cc-agents/SKILL.md)
4. 🎓 Skills: [`moai-cc-skills`](./../../skills/moai-cc-skills/SKILL.md)
5. ⚙️ settings.json: [`moai-cc-settings`](./../../skills/moai-cc-settings/SKILL.md)
6. 🔌 Plugins/MCP: [`moai-cc-mcp-plugins`](./../../skills/moai-cc-mcp-plugins/SKILL.md)
7. 📝 CLAUDE.md: [`moai-cc-claude-md`](./../../skills/moai-cc-claude-md/SKILL.md)
8. 💾 Memory: [`moai-cc-memory`](./../../skills/moai-cc-memory/SKILL.md)

**Workflow Guides:**
- 📖 [Project Setup: `/alfred:0-project`](./workflows/alfred-0-project-setup.md)
- 📖 [Planning: `/alfred:1-plan`](./workflows/alfred-1-plan-flow.md)
- 📖 [Implementation: `/alfred:2-run`](./workflows/alfred-2-run-flow.md)
- 📖 [Synchronization: `/alfred:3-sync`](./workflows/alfred-3-sync-flow.md)

---

## ✅ When to Use This Skill

**Use moai-cc-guide when:**
- ✅ Setting up a new Claude Code environment for MoAI-ADK
- ✅ Unsure which Claude Code component to use (Agents? Hooks? Plugins?)
- ✅ Integrating Claude Code with MoAI-ADK workflows
- ✅ Troubleshooting: "Why isn't this working?"
- ✅ Planning security or automation improvements

**Delegate to specific Skills when:**
- 🔗 Need detailed Hooks configuration → `moai-cc-hooks`
- 🔗 Creating new Agents → `moai-cc-agents`
- 🔗 Setting up commands → `moai-cc-commands`
- 🔗 Configuring MCP plugins → `moai-cc-mcp-plugins`
- 🔗 Writing CLAUDE.md → `moai-cc-claude-md`

---

## 📞 Next Steps

1. **Choose your workflow**: [Project Setup](./workflows/alfred-0-project-setup.md) | [Planning](./workflows/alfred-1-plan-flow.md) | [Implementation](./workflows/alfred-2-run-flow.md) | [Synchronization](./workflows/alfred-3-sync-flow.md)

2. **Reference specific components**: Pick from the 8-component knowledge base above

3. **Ask for help**: `@agent-cc-manager "I need to [task]"` → Will reference this guide and suggest components

---

**Reference**: MoAI-ADK Claude Code Integration Guide
**Version**: 3.0.0
**Last Updated**: 2025-10-23
**Claude Code Version**: 0.6.0
**MoAI-ADK Mode**: SPEC-First Learning
