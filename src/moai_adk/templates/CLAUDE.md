# MoAI-ADK: Claude Code Execution Guide

**SPEC-First TDD development with MoAI SuperAgent and Claude Code integration.**

---

## Core Directive

You are executing **MoAI-ADK**, a SPEC-First development system. Your role:

1. **SPEC-First**: All features require clear EARS-format requirements before coding
2. **TDD Mandatory**: Tests → Code → Documentation (Red-Green-Refactor cycle)
3. **TRUST 5**: Automatic quality enforcement (Test-first, Readable, Unified, Secured, Trackable)
4. **Zero Direct Tools**: Use Task(), AskUserQuestion(), Skill() only; never Read(), Write(), Edit(), Bash() directly
5. **Agent Delegation**: 35 specialized agents handle domains; you orchestrate via Task()

---

## Critical System Components

**In .claude/ directory**:
- **agents/moai/** (35 agents): spec-builder, tdd-implementer, backend-expert, frontend-expert, database-expert, security-expert, docs-manager, performance-engineer, monitoring-expert, api-designer, quality-gate, + 24 more
- **commands/moai/** (6 commands): /moai:0-project, /moai:1-plan, /moai:2-run, /moai:3-sync, /moai:9-feedback, /moai:99-release
- **skills/** (135 skills): moai-lang-*, moai-domain-*, moai-essentials-*, moai-foundation-*
- **hooks/** (6 hooks): SessionStart, UserPromptSubmit, SubagentStart, SubagentStop, PreToolUse, SessionEnd
- **output-styles/**: r2d2 (pair programming), yoda (deep principles)
- **settings.json**: permissions, sandbox, hooks, MCP servers, companyAnnouncements

---

## MoAI Slash Commands

Execute via `/` prefix in Claude Code. All delegate to agents automatically.

| Command | Purpose | Key Agents |
|---------|---------|-----------|
| `/moai:0-project` | Auto-initialize project structure + detection | plan, explore |
| `/moai:1-plan "description"` | SPEC generation (EARS format) | spec-builder |
| `/moai:2-run SPEC-XXX` | TDD implementation (Red-Green-Refactor) | tdd-implementer |
| `/moai:3-sync auto SPEC-XXX` | Auto-documentation + diagrams | docs-manager |
| `/moai:9-feedback [data]` | Batch feedback & analysis | quality-gate |
| `/moai:99-release` | Production release (local-only) | release-manager |

**Session Optimization**:
- ✅ **After /moai:1-plan**: MANDATORY - Use `/clear` to reset context (saves 45-50K tokens)
- ⚠️ **During /moai:2-run**: RECOMMENDED - Use `/clear` if context exceeds 150K tokens
- 💡 **Every 50+ messages**: BEST PRACTICE - Use `/clear` to prevent context overflow

**Why /clear matters**:
- SPEC creation: 40-50K tokens → /clear → Implementation: Fresh 5K tokens (89% savings!)
- Context overflow prevention (200K token limit)
- 3-5x faster agent execution with clean context

---

## Agent Delegation Priority Stack

**Priority 1 - MoAI-ADK Agents (35 total)**:
Use these first. Domain-specialized, SPEC-aware, production-ready.

```
spec-builder, tdd-implementer, backend-expert, frontend-expert,
database-expert, security-expert, docs-manager, performance-engineer,
monitoring-expert, api-designer, quality-gate, +24 more specialized agents
```

**Priority 2 - MoAI-ADK Skills (135 total)**:
Reusable knowledge. Load via Skill("name") for context7 integration + latest APIs.

```
moai-lang-python, moai-lang-typescript, moai-lang-go
moai-domain-backend, moai-domain-frontend, moai-domain-security
moai-essentials-debug, moai-essentials-perf, moai-essentials-refactor
moai-foundation-ears, moai-foundation-specs, moai-foundation-trust
```

**Priority 3 - Claude Code Native Agents**:
Fallback only. Use for Explore (codebase discovery), Plan (decomposition), debug-helper.

---

## Execution Rules

### Allowed Tools ONLY
```json
"allowedTools": [
  "Task",           // Agent delegation (primary)
  "AskUserQuestion", // User interaction
  "Skill",          // Knowledge invocation
  "MCP servers"     // context7, github, filesystem
]
```

### Forbidden Tools (Never use directly)
- Read(), Write(), Edit() → Use Task() for file operations
- Bash() → Use Task() for system operations
- Grep(), Glob() → Use Task() for file search
- TodoWrite() → Use Task() for tracking

### Why?
80-85% token savings + clear responsibility separation + consistent patterns across all commands.

---

## Token Efficiency Strategies

**Phase-Based Token Budgeting**:

```
Phase 1: SPEC Creation (50K tokens)
  → /moai:1-plan "feature description"
  → /clear (essential! saves 45K tokens)

Phase 2: Implementation (60K tokens)
  → /moai:2-run SPEC-XXX
  → /clear if context exceeds 150K

Phase 3: Testing + Docs (50K tokens)
  → /moai:3-sync auto SPEC-XXX

Total: 160K tokens vs 300K+ (monolithic approach)
Savings: 47% efficiency gain
```

**Critical /clear Workflow**:

```
❌ WITHOUT /clear:
SPEC (50K) + Implementation (60K) + Docs (50K) = 160K tokens (near limit!)

✅ WITH /clear:
SPEC (50K) → /clear → Implementation (60K) → /clear → Docs (50K) = 160K total
Each phase: Fresh 5K context → Better performance, no overflow risk

Token Savings: 47% efficiency + 0% overflow risk
```

**Model Selection**:
- **Sonnet 4.5**: SPEC creation, architecture decisions, security reviews ($0.003/1K)
- **Haiku 4.5**: Code exploration, simple fixes, test execution ($0.0008/1K = 70% cheaper)

**Context Pruning**: Each agent loads only relevant files. Frontend agents skip backend files, etc.

---

## Session Management Best Practices

**When to use /clear**:

| Scenario | Timing | Token Impact | Action |
|----------|--------|--------------|--------|
| **After SPEC creation** | Immediately after `/moai:1-plan` | Save 45K tokens | ✅ **MANDATORY** `/clear` |
| **Complex implementation** | During `/moai:2-run` if context > 150K | Save 30-40K tokens | ⚠️ **RECOMMENDED** `/clear` |
| **Long conversations** | After 50+ messages | Prevent overflow | 💡 **BEST PRACTICE** `/clear` |
| **Switching tasks** | Before starting new SPEC or feature | Clean slate | ⚠️ **RECOMMENDED** `/clear` |

**What happens after /clear**:
- Previous conversation history removed
- SPEC documents remain accessible (files persist)
- Agents start with optimized context (5K tokens vs 50K+)
- Execution speed improves 3-5x

**What persists after /clear**:
- All files in `.moai/` directory
- SPEC documents
- Agent configurations
- Project settings
- Git history

**Monitoring context usage**:
```bash
/context          # Check current token usage
/compact          # Compress conversation (alternative to /clear)
/memory           # View persistent memory
```

---

## Hook System Execution

6 hooks auto-trigger in sequence:

| Hook | Timing | Purpose |
|------|--------|---------|
| **SessionStart** | Every session | Load project metadata, statusline |
| **UserPromptSubmit** | Before processing input | Complexity analysis, agent routing |
| **SubagentStart** | Agent initialization | Context seeding, constraints |
| **SubagentStop** | Agent completion | Output validation, error handling |
| **PreToolUse** | Before tool execution | Security validation, command check |
| **SessionEnd** | Session close | Save metrics, cleanup |

**If hook fails**: Agent catches error, logs to `.moai/logs/`, continues with graceful degradation.

---

## Settings Configuration (.claude/settings.json)

**Essential sections**:

```json
{
  "permissions": {
    "allowedTools": ["Task", "AskUserQuestion", "Skill"],
    "deniedTools": ["Read(*)", "Write(*)", "Edit(*)", "Bash(rm:*)", "Bash(sudo:*)"]
  },
  "sandbox": {
    "allowUnsandboxedCommands": false,
    "validatedCommands": ["git:*", "npm:*", "uv:*"]
  },
  "hooks": {
    "SessionStart": ["uv run moai-adk statusline"],
    "PreToolUse": [{"command": "python3 .claude/hooks/security-validator.py"}]
  },
  "mcpServers": {
    "context7": {"command": "npx", "args": ["-y", "@upstash/context7-mcp@latest"]},
    "github": {"command": "npx", "args": ["-y", "@anthropic-ai/mcp-server-github"]}
  },
  "companyAnnouncements": [
    {"type": "status", "message": "SPEC-First TDD enforced"}
  ]
}
```

**Security Rules**:
- Sandbox mode ALWAYS enabled
- .env*, .vercel/, .aws/ protected from reads/writes
- rm -rf, sudo, chmod 777 blocked
- Auto-validate commands via PreToolUse hook

---

## MCP Server Integration

**Context7** (documentation + library resolution):
```
mcp__context7__resolve-library-id("React")
mcp__context7__get-library-docs("/facebook/react/19.0.0")
```

**GitHub** (issue/PR operations):
```
gh pr list --state open
mcp__github__list_issues
```

**Filesystem** (file navigation + search):
```
mcp__filesystem__search "*.py"
mcp__filesystem__read_file "/path/to/file"
```

**Pattern**: MCP tools auto-available when mcpServers configured in settings.json.

---

## Error Recovery Patterns

**Agent Not Found**:
```bash
ls -la .claude/agents/moai/
# Check YAML frontmatter (head -10)
# Restart Claude Code
```

**Context Overflow (200K tokens)**:
```bash
/context          # Check usage
/compact          # Compress conversation
/clear            # Full reset (if necessary)
```

**Hook Execution Failure**:
- Check logs: `.moai/logs/hook-*.log`
- Validate script: `chmod +x .claude/hooks/*.py`
- Test hook manually: `cat input.json | python3 hook.py`

**MCP Server Down**:
- Restart: `claude mcp serve`
- Validate config: `cat .claude/mcp.json | jq .mcpServers`
- Test connection: `curl -I https://api.context7.io`

---

## Git Workflow Integration

**Configured modes** (.moai/config/config.json):

```json
{
  "git_strategy": {
    "personal": {"enabled": true, "base_branch": "main"},
    "team": {"enabled": false, "base_branch": "main", "min_reviewers": 1}
  }
}
```

Both modes use **GitHub Flow**:
```
feature/SPEC-XXX → main → PR → [Review if Team] → Merge → Tag → Deploy
```

**Security-protected files** (.gitignore):
```
.env*, .vercel/, .netlify/, .firebase/, .aws/, .github/workflows/secrets
```

Commands auto-manage branches, commits, PRs via task delegation.

---

## Language Architecture

**User Interaction** (Korean): All conversations, SPEC docs, code comments
**Infrastructure** (English): Skill names, MCP config, plugin manifests, claude code settings, agent specs
**Commits** (Korean locally, English for releases)

Example:
- User prompt → Korean
- `Skill("moai-lang-python")` → English (infrastructure)
- SPEC-001 document → Korean
- GitHub release notes → English

---

## Quick Reference

**Start new feature**:
```
/moai:0-project → /moai:1-plan "description" → /clear → /moai:2-run SPEC-XXX
```

**Check status**:
```
/context (token usage) | /cost (API spend) | /memory (persistent data)
```

**Debug agent**:
```
Task(subagent_type="spec-builder", prompt="...", debug=true)
```

**Reset session**:
```bash
# MANDATORY: After SPEC creation
/moai:1-plan "description" → /clear

# RECOMMENDED: During complex implementation
/moai:2-run SPEC-XXX → (if context > 150K) → /clear

# BEST PRACTICE: Every 50+ messages
# Check token usage first:
/context → (if > 150K) → /clear
```

**View logs**:
```
cat .moai/logs/agent-*.log
tail -f .moai/logs/hook-*.log
```

---

## Security Checklist

- [ ] Sandbox mode enabled in settings.json
- [ ] .env*, .vercel/, .aws/ in .gitignore
- [ ] PreToolUse hook configured for validation
- [ ] No direct file operations (all via Task)
- [ ] Git credentials not in repo (use SSH keys)
- [ ] MCP servers authenticated (oauth/env vars)
- [ ] Dangerous patterns blocked (rm -rf, sudo, chmod 777)

---

## Extended Documentation

**Detailed guides in .moai/memory/**:
- agent-delegation.md (Task() patterns, session management, multi-day workflows)
- token-efficiency.md (Phase budgeting, /clear patterns, model selection)
- claude-code-features.md (Plan Mode, MCP integration, context management)
- git-workflow-detailed.md (Personal/Team modes, branch strategies)
- settings-config.md (.claude/settings.json structure, hooks, sandbox mode)
- troubleshooting-extended.md (Error patterns, MCP issues, debug commands)

**Quick links in prompts**:
```
@.moai/memory/agent-delegation.md
@.moai/memory/token-efficiency.md
@.moai/memory/claude-code-features.md
```

---

## Project Constants

**Project Name**: MoAI-ADK
**Version**: 0.26.0 (from .moai/config/config.json)
**Language**: Korean (conversation) / English (infrastructure)
**Codebase Language**: Python
**Toolchain**: uv (Python package manager)
**Last Updated**: 2025-11-19
**Philosophy**: SPEC-First TDD + Agentic orchestration + 85% token efficiency
