# Workflow: Project Setup with Claude Code (`/alfred:0-project`)

## Objective
Bootstrap a new MoAI-ADK project with Claude Code security, automation, and standards.

## Claude Code Components Involved

| Component | Purpose | Reference |
|-----------|---------|-----------|
| **settings.json** | Define basic permissions, restricted tools | [`moai-cc-settings`](../../../skills/moai-cc-settings/SKILL.md) |
| **CLAUDE.md** | Document project standards, SPEC-first principles | [`moai-cc-claude-md`](../../../skills/moai-cc-claude-md/SKILL.md) |
| **Hooks** | SessionStart hook to display project status | [`moai-cc-hooks`](../../../skills/moai-cc-hooks/SKILL.md) |
| **Skills** | Register domain-specific patterns (moai-foundation-*) | [`moai-cc-skills`](../../../skills/moai-cc-skills/SKILL.md) |

## Step-by-Step Setup

### Step 1: Initialize Project Metadata
```bash
/alfred:0-project
# Inputs: project name, tech stack, team size, MoAI mode
# Outputs: .moai/config.json, project structure
```

### Step 2: Configure Claude Code Security
```bash
@agent-cc-manager "Configure project-level settings.json with:
- allowedTools: Read(src/**, tests/**), Write(tests/**), Edit(src/**), Bash(pytest:*), Bash(npm run:*)
- deniedTools: Read(.env), Read(.env.*), Read(secrets/**), Bash(rm -rf:*), Bash(sudo:*)
- permissionMode: 'ask'
"
```

**What this does:**
- ✅ Restricts reads to source & test files
- ✅ Blocks access to environment secrets
- ✅ Allows safe testing commands
- ✅ Prompts before executing

### Step 3: Create Project CLAUDE.md
```bash
@agent-cc-manager "Create .claude/CLAUDE.md documenting:
- SPEC-first development workflow
- EARS requirement syntax
- TRUST 5 principles
- @TAG traceability rules
- Custom slash commands for this project
- Import Foundation tier Skills (moai-foundation-trust, moai-foundation-tags)
"
```

**What this does:**
- 📝 Centralizes project guidance
- 🔗 Imports TRUST/TAG/SPEC standards
- 📚 Provides AI context for all sessions
- 🎯 Establishes team conventions

### Step 4: Configure SessionStart Hook
```bash
@agent-cc-manager "Register SessionStart hook to:
- Display project name and version
- Show recent SPEC files (.moai/specs/)
- List active feature branches
- Indicate TRUST 5 compliance status
"
```

**What this does:**
- 👋 Greets user with project context
- 📊 Shows current workflow status
- 🎯 Reminds of SPEC-first workflow

### Step 5: Register Hooks for Code Quality
```bash
@agent-cc-manager "Add hooks:
1. PreToolUse (Edit/Write): Validate @TAG presence in all edits
2. PostToolUse (Edit): Auto-format code (black for Python, prettier for TS/JS)
3. PostToolUse (Edit): Run linter (ruff for Python, eslint for JS)
"
```

**What this does:**
- ✅ Enforces TAG traceability
- ✅ Auto-formats all code edits
- ✅ Catches style issues immediately

## Validation Checklist

- [ ] `.moai/config.json` created and valid
- [ ] `.claude/settings.json` exists with proper permissions
- [ ] `.claude/CLAUDE.md` documents SPEC-first principles
- [ ] Foundation tier Skills imported (@moai-foundation-*)
- [ ] SessionStart hook shows project status
- [ ] PreToolUse hook validates @TAG
- [ ] PostToolUse hooks auto-format and lint
- [ ] All team members can access `.claude/` files

## Troubleshooting

**Issue**: "Permission denied when running tests"
→ Add `Bash(pytest:*)` or `Bash(npm run:*)` to allowedTools

**Issue**: "SessionStart hook not running"
→ Check `.claude/settings.json` hooks section, ensure script path is absolute

**Issue**: "@TAG validation slowing down edits"
→ Adjust hook timeout or make validation conditional (warn, don't block)

## Next Steps
→ Move to `/alfred:1-plan` for SPEC authoring
