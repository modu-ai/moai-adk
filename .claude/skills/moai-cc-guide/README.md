# Claude Code Guide for MoAI-ADK — Quick Start & FAQ

Welcome! This guide helps you navigate Claude Code's seven components and integrate them with MoAI-ADK's SPEC → TDD → Sync workflow.

---

## ⚡ Quick Start (5 minutes)

### I want to...

#### 1️⃣ **Set up a new MoAI-ADK project with Claude Code**
```bash
/alfred:0-project
```
Then follow: **[Project Setup Workflow](./workflows/alfred-0-project-setup.md)**
- ✅ Creates `.claude/settings.json` with security
- ✅ Generates `.claude/CLAUDE.md` with standards
- ✅ Registers Hooks for quality gates

**Skill to reference:** `Skill("moai-cc-settings")` + `Skill("moai-cc-hooks")`

---

#### 2️⃣ **Write my first SPEC**
```bash
/alfred:1-plan "User authentication system"
```
Then follow: **[Planning Phase Workflow](./workflows/alfred-1-plan-flow.md)**
- ✅ Loads EARS patterns from `moai-foundation-ears` Skill
- ✅ Validates SPEC metadata
- ✅ Assigns unique @SPEC:ID

**Skill to reference:** `Skill("moai-cc-guide")` (SKILL.md section 7️⃣)

---

#### 3️⃣ **Implement the SPEC using TDD**
```bash
/alfred:2-run AUTH-001
```
Then follow: **[Implementation Phase Workflow](./workflows/alfred-2-run-flow.md)**
- ✅ RED: Write failing tests with @TEST:ID
- ✅ GREEN: Implement with @CODE:ID
- ✅ REFACTOR: Improve per TRUST 5 principles
- ✅ Hooks auto-format code after each edit

**Skill to reference:** `Skill("moai-cc-hooks")` + `Skill("moai-cc-settings")`

---

#### 4️⃣ **Sync docs and create PR**
```bash
/alfred:3-sync
```
Then follow: **[Synchronization Phase Workflow](./workflows/alfred-3-sync-flow.md)**
- ✅ Verify @TAG chains (@SPEC → @TEST → @CODE → @DOC)
- ✅ Generate Living Docs
- ✅ Create GitHub PR
- ✅ Validate TRUST 5

**Skill to reference:** `Skill("moai-cc-mcp-plugins")` (GitHub MCP)

---

#### 5️⃣ **Configure Hooks for my project**
```bash
@agent-cc-manager "Configure PreToolUse hook to validate @TAG in all edits"
```
**Skill to reference:** `Skill("moai-cc-hooks")`

---

#### 6️⃣ **Create a Sub-agent**
```bash
@agent-cc-manager "Create a code-reviewer agent for quality checks"
```
**Skill to reference:** `Skill("moai-cc-agents")`

---

#### 7️⃣ **Integrate GitHub for PR automation**
```bash
@agent-cc-manager "Set up GitHub MCP for /alfred:3-sync PR creation"
```
**Skill to reference:** `Skill("moai-cc-mcp-plugins")`

---

#### 8️⃣ **Decide: Agents vs Commands vs Skills vs Hooks**
Read: **SKILL.md → Quick Decision Tree**

Or ask: `@agent-cc-manager "Should I use an Agent or a Command for X?"`

**Skill to reference:** `Skill("moai-cc-guide")` (SKILL.md section 🎯)

---

## ❓ Frequently Asked Questions (FAQ)

### General Claude Code Questions

**Q: What's the difference between Agents and Commands?**
A:
- **Agents**: Run in isolated context, handle complex analysis independently
- **Commands**: Orchestrate workflow steps, pass data between stages
- **Decision**: Use Agents for analysis, Commands for orchestration
- **Read**: SKILL.md → Quick Decision Tree or `Skill("moai-cc-guide")`

---

**Q: Do I need all 8 Claude Code components?**
A:
- **Minimum**: settings.json (security) + CLAUDE.md (guidance)
- **Typical**: Add Hooks (quality), Agents (debugging), Commands (workflow)
- **Advanced**: Add Skills (patterns), MCP (GitHub), Memory (optimization)
- **Read**: SKILL.md → Component Overview

---

**Q: How do I restrict what Claude Code can do?**
A:
- Use `settings.json` → `permissions` → `allowedTools` / `deniedTools`
- Specific paths: `Read(src/**)`, `Edit(src/**/*.py)`
- Specific commands: `Bash(pytest:*)` (not `Bash(*)`)
- **Read**: `Skill("moai-cc-settings")`

---

**Q: What are Hooks and when do I need them?**
A:
- **Purpose**: Auto-run scripts on events (PreToolUse, PostToolUse, SessionStart)
- **Examples**: Auto-format code, validate @TAGs, show project status
- **When**: Adding quality checks, enforcing standards, automation
- **Read**: `Skill("moai-cc-hooks")` + workflows/alfred-2-run-flow.md

---

### MoAI-ADK Integration Questions

**Q: How do I integrate Claude Code with /alfred:0-project?**
A:
- Step 1: Create `settings.json` with permissions (moai-cc-settings)
- Step 2: Create `CLAUDE.md` documenting SPEC-first (moai-cc-claude-md)
- Step 3: Register Hooks for project status (moai-cc-hooks)
- Step 4: Import Foundation tier Skills (moai-cc-guide)
- **Read**: workflows/alfred-0-project-setup.md

---

**Q: Which Claude Code components are used in /alfred:2-run?**
A:
- **PreToolUse Hook**: Validate @TAG before edits
- **PostToolUse Hook**: Auto-format code after edits
- **Sub-agents**: code-builder pipeline (RED/GREEN/REFACTOR)
- **settings.json**: Restrict edits to src/, tests/
- **CLAUDE.md**: Reference TRUST 5 principles
- **Read**: workflows/alfred-2-run-flow.md

---

**Q: How does /alfred:3-sync use Claude Code?**
A:
- **tag-agent Sub-agent**: Verify @TAG chains
- **doc-syncer Sub-agent**: Generate Living Docs
- **GitHub MCP**: Create/update PR
- **Hooks**: Auto-generate README/CHANGELOG
- **Read**: workflows/alfred-3-sync-flow.md

---

**Q: Should I use Claude Code Hooks or Alfred Hooks?**
A:
- **Claude Code Hooks** (settings.json): General events (Edit, Bash, SessionStart)
- **Alfred Hooks** (cc-manager): MoAI-ADK specific (RED/GREEN, SPEC validation)
- **Use both**: Claude Code for general validation, Alfred for MoAI logic
- **Read**: `Skill("moai-cc-hooks")` + cc-manager.md

---

### Configuration Questions

**Q: My @TAG validation Hook is too slow. How do I speed it up?**
A:
- **Check 1**: Reduce regex complexity (avoid full-file scans)
- **Check 2**: Cache recent @TAGs in memory
- **Check 3**: Only validate on Edit/Write (not on Bash)
- **Check 4**: Set Hook timeout: `"timeout": 30` (milliseconds)
- **Read**: `Skill("moai-cc-hooks")` → Best Practices section

---

**Q: How do I prevent Claude from editing .env or secrets?**
A:
```json
{
  "permissions": {
    "deniedTools": [
      "Read(.env)",
      "Read(.env.*)",
      "Read(secrets/**)",
      "Edit(.env)"
    ]
  }
}
```
- **Read**: `Skill("moai-cc-settings")` → Dangerous Tools section

---

**Q: Can I use multiple MCP servers at once?**
A:
- Yes! Register in `settings.json` → `mcpServers` as separate objects
- **Example**: GitHub + Filesystem + Brave Search simultaneously
- **Important**: Each MCP needs proper OAuth/env vars
- **Read**: `Skill("moai-cc-mcp-plugins")`

---

**Q: How do I write CLAUDE.md for my project?**
A:
- **Section 1**: Project mission & standards
- **Section 2**: SPEC-first principles (EARS patterns)
- **Section 3**: TRUST 5 principles & quality gates
- **Section 4**: @TAG traceability rules
- **Section 5**: Import Foundation tier Skills
- **Read**: `Skill("moai-cc-claude-md")` + workflows/alfred-0-project-setup.md

---

## 🔧 Troubleshooting

### Problem: "Permission denied when running pytest"

**Diagnosis:**
1. Check `.claude/settings.json` permissions
2. Look for `Bash(pytest:*)` in `allowedTools`

**Solution:**
```json
{
  "permissions": {
    "allowedTools": ["Bash(pytest:*)"]
  }
}
```

**Learn more:** `Skill("moai-cc-settings")` → Tool Permission Patterns

---

### Problem: "PreToolUse Hook not blocking dangerous command"

**Diagnosis:**
1. Check Hook script logic (is pattern correct?)
2. Verify Hook exit code (should be 2 to block)
3. Check Hook path in settings.json

**Solution (example):**
```bash
#!/bin/bash
if [[ "$COMMAND" =~ "rm -rf /" ]]; then
  echo "🔴 Blocked: rm -rf detected" >&2
  exit 2  # Must use exit code 2 to block
fi
exit 0
```

**Learn more:** `Skill("moai-cc-hooks")` → Hook Exit Codes

---

### Problem: "SessionStart Hook not running"

**Diagnosis:**
1. Check `.claude/settings.json` hooks section
2. Verify script path is absolute (not relative)
3. Check script is executable: `chmod +x hook.sh`

**Solution:**
```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "bash /Users/username/.claude/hooks/session-status.sh"
          }
        ]
      }
    ]
  }
}
```

**Learn more:** `Skill("moai-cc-hooks")` → Hook Configuration

---

### Problem: "GitHub MCP not creating PR"

**Diagnosis:**
1. Check GITHUB_TOKEN is set: `echo $GITHUB_TOKEN`
2. Verify MCP config in settings.json
3. Check gh CLI is installed: `which gh`

**Solution:**
```bash
# Set environment variable
export GITHUB_TOKEN="your_token_here"

# Test gh CLI
gh auth status

# Then try /alfred:3-sync again
```

**Learn more:** `Skill("moai-cc-mcp-plugins")` → Plugin Installation Guide

---

### Problem: "Skill not loading or unrecognized"

**Diagnosis:**
1. Check Skill file exists: `ls .claude/skills/moai-cc-*/`
2. Verify YAML frontmatter syntax
3. Check filename is kebab-case

**Solution:**
```bash
# Verify Skill exists
ls -la .claude/skills/moai-cc-hooks/

# Check YAML validity
head -10 .claude/skills/moai-cc-hooks/SKILL.md

# Restart Claude Code for changes to take effect
```

**Learn more:** `Skill("moai-cc-skills")` → Skill Registration

---

### Problem: "Custom Agent not working"

**Diagnosis:**
1. Check agent YAML frontmatter (name, description, tools, model)
2. Verify description includes "Use PROACTIVELY for"
3. Check tool restrictions (avoid `Bash(*)`)

**Solution (example):**
```yaml
---
name: code-reviewer
description: Use PROACTIVELY for code review requests, quality audits
tools: Read, Glob, Grep, Bash(git:*)
model: sonnet
---

# Code Reviewer — Quality Expert
[rest of agent content]
```

**Learn more:** `Skill("moai-cc-agents")` → Agent Structure

---

## 📞 Getting Help

### Step 1: Identify the component
Use the **Quick Decision Tree** in SKILL.md or ask:
```bash
@agent-cc-manager "Should I use a Hook, Agent, or Command for X?"
```

### Step 2: Find the relevant Skill
- Architecture question? → `Skill("moai-cc-guide")`
- Hooks setup? → `Skill("moai-cc-hooks")`
- Agents creation? → `Skill("moai-cc-agents")`
- Settings/security? → `Skill("moai-cc-settings")`
- MCP integration? → `Skill("moai-cc-mcp-plugins")`
- Workflow integration? → Skill + workflows/ guides

### Step 3: Check the workflow guide
- Starting project? → workflows/alfred-0-project-setup.md
- Writing SPEC? → workflows/alfred-1-plan-flow.md
- Implementing code? → workflows/alfred-2-run-flow.md
- Syncing docs? → workflows/alfred-3-sync-flow.md

### Step 4: Ask cc-manager
```bash
@agent-cc-manager "[Your specific question with context]"
```

---

## 📚 Complete Knowledge Base

### Orchestrator Guide (Start Here)
- **SKILL.md** - Decision trees, component overview, MoAI integration
- **README.md** (this file) - Quick start, FAQ, troubleshooting

### MoAI-ADK Workflows
- **workflows/alfred-0-project-setup.md** - Initial Claude Code setup
- **workflows/alfred-1-plan-flow.md** - SPEC authoring with Claude Code
- **workflows/alfred-2-run-flow.md** - TDD with Hooks & validation
- **workflows/alfred-3-sync-flow.md** - Docs & PR automation

### Specialized Skills (Choose What You Need)
| Skill | Use When | File |
|-------|----------|------|
| `moai-cc-agents` | Creating Sub-agents | `./.../moai-cc-agents/SKILL.md` |
| `moai-cc-commands` | Designing Commands | `./.../moai-cc-commands/SKILL.md` |
| `moai-cc-skills` | Building Skills | `./.../moai-cc-skills/SKILL.md` |
| `moai-cc-hooks` | Configuring Hooks | `./.../moai-cc-hooks/SKILL.md` |
| `moai-cc-settings` | Managing security/permissions | `./.../moai-cc-settings/SKILL.md` |
| `moai-cc-mcp-plugins` | Integrating GitHub/Filesystem | `./.../moai-cc-mcp-plugins/SKILL.md` |
| `moai-cc-claude-md` | Writing project instructions | `./.../moai-cc-claude-md/SKILL.md` |
| `moai-cc-memory` | Optimizing context usage | `./.../moai-cc-memory/SKILL.md` |

---

## 🎓 Learning Path

**Beginner** (30 minutes)
1. Read: SKILL.md → Quick Decision Tree
2. Follow: workflows/alfred-0-project-setup.md
3. Done! You can now set up Claude Code for a project

**Intermediate** (1 hour)
1. Follow all 4 workflow guides (0-3)
2. Read: `Skill("moai-cc-settings")` & `Skill("moai-cc-hooks")`
3. Done! You can configure security & automation

**Advanced** (2+ hours)
1. Deep-dive into specialized Skills (agents, commands, MCP)
2. Read: `Skill("moai-cc-memory")` for context optimization
3. Experiment: Create custom Hooks, Agents, Skills
4. Done! You're a Claude Code expert

---

## 📈 v3.0.0 Changelog

**v3.0.0 (2025-10-23)**
- ✨ Created moai-cc-guide orchestrator
- ✨ Added 4 MoAI-ADK workflow guides
- ✨ Decoupled knowledge from cc-manager agent
- 📚 Organized into 9 specialized Skills
- 🎯 Added Quick Decision Tree
- 🔗 Cross-referenced all components

**v2.0.0 (2025-10-22)**
- 📦 Initial 8 Claude Code Skills created
- 🛠️ cc-manager agent templates

**v1.0.0 (2025-10-21)**
- 🚀 Project kickoff

---

## 💬 Feedback & Questions

- **Question about this guide?** Ask in the main chat
- **Bug or issue?** Reference the skill name + symptom
- **Suggestion?** Describe what would help you

---

**Happy coding with Claude Code + MoAI-ADK! 🚀**

For the authoritative guide, start with:
- 📖 **SKILL.md** (architecture & decisions)
- 📖 **workflows/** (MoAI-ADK integration)
- 📖 **Specialized Skills** (deep dives)
