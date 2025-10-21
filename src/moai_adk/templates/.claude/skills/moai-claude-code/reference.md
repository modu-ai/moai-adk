# Claude Code component writing guide

> **A complete guide to writing 5 components**
>
> Agent, Command, Skill, Plugin, Settings

---

## 📋 Table of Contents

1. [Agent writing guide](#1-agent-writing-guide)
2. [Command writing guide](#2-command-writing-guide)
3. [Skill writing guide](#3-skill-writing-guide)
4. [Plugin setting guide](#4-plugin-setting-guide)
5. [Settings settings guide](#5-settings-settings-guide)

---

## 1. Agent writing guide

### 📐 File Structure

**Location**: `.claude/agents/{agent-name}.md`

**YAML Frontmatter** (required):
```yaml
---
name: {agent-name}              # kebab-case
description: "Use when: {trigger}" # "Use when:" pattern required
tools: Read, Write, Edit # Only necessary tools
model: sonnet                   # sonnet|haiku
---
```

### 🎭 Agent Persona

**Required elements**:
- **Icon**: Visual identifier (emoji)
- **Duties**: IT professional duties (System Architect, QA Lead, etc.)
- **Area of ​​expertise**: Specific area of ​​expertise
- **Role**: Agent responsibilities
- **Goals**: what you want to achieve

**example**:
```markdown
## 🎭 Agent Persona

**Icon**: 🏗️
**Job**: System Architect
**Area of ​​Expertise**: SPEC writing, EARS specification, requirements analysis
**Role**: Convert business requirements into systematic SPEC
**Goal**: Clear and testable SPEC Write a document
```

### ⚙️ Model selection guide

| model | When to use | Example |
|------|----------|------|
| **sonnet** | Complex judgment, design, creativity | SPEC writing, TDD strategy, debugging |
| **haiku** | Fast processing, pattern-based operation | Document synchronization, TAG scanning, linting |

### 🛠️ Tool Selection Guide

| Job type | Essential Tools |
|----------|----------|
| **Analysis** | Read, Grep, Glob |
| **Create Document** | Read, Write, Edit |
| **Code Implementation** | Read, Write, Edit, MultiEdit |
| **Git Operations** | Read, Bash(git:*) |
| **Verification** | Read, Grep, Bash |

### ✅ Verification Checklist

- [ ] YAML frontmatter exists
- [ ] `name`: kebab-case
- [ ] `description`: Contains the “Use when:” pattern
- [ ] `tools`: Only the tools you need
- [ ] `model`: sonnet or haiku
- [ ] Contains the agent persona section
- [ ] Contains workflow specific steps

---

## 2. Command writing guide

### 📐 File Structure

**Location**: `.claude/commands/{command-name}.md`

**YAML Frontmatter** (required):
```yaml
---
name: {command-name}            # kebab-case
description: {one-line description} # Clear purpose
argument-hint: [{param}] # Optional
allowed-tools: # Only the tools you need
  - Read
  - Write
  - Task
---
```

### 🔧 Naming Conventions

- Use **kebab-case**
- **Start with a verb** (run, check, deploy, create)
- **Clear and specific**

**Correct Example**:
- ✅ `deploy-production`
- ✅ `run-tests`
- ✅ `alfred:1-spec`

**Incorrect example**:
- ❌ `doSomething` (camelCase)
- ❌ `cmd1` (unclear)

### 📋 Standard section structure

```markdown
# {Command Title}

{Brief description}

## 🎯 Command Purpose
{Detailed purpose}

## 💡 Example of use
\`\`\`bash
/{command-name} {example-args}
\`\`\`

## 📋 Execution flow
1. **Phase 1**: {Planning}
2. **Phase 2**: {Execution}

## 🔗 Associated Agent
- **Primary**: {agent-name} - {role}

## ⚠️ Precautions
- {Warning 1}

## 📋 Next steps
- {Next step}
```

### ✅ Verification Checklist

- [ ] YAML frontmatter exists
- [ ] `name`: kebab-case
- [ ] `description`: One-line description
- [ ] `allowed-tools`: Array format
- [ ] Specific patterns when using Bash tools (`Bash(git:*)`)
- [ ] Include usage examples
- [ ] Specify execution flow

---

## 3. Skill creation guide

### 📐 File Structure

**Location**: `.claude/skills/{skill-name}/SKILL.md`

**YAML Frontmatter** (required fields + optional `allowed-tools`):
```yaml
---
name: {skill-name}              # kebab-case, ≤64 chars
description: {What it does + when to use (≤1024 chars)}
allowed-tools:
  - Read                       # optional, enforce least privilege
  - Bash
---
```

### 🎯 How to write a description

**Important**: Key field that determines when Claude will call the skill (<=1024 chars; aim for ≤200 for clarity)

**Good example**:
- ✅ "Directly scan TAG markers and create inventory (CODE-FIRST principle)"
- ✅ "Select optimal features by project type (37 skills → automatically filter 3-5)"

**Bad example**:
- ❌ “This skill does something” (too vague)
- ❌ “This skill does something” (non-specific)

### 📁 Optional file structure

```
.claude/skills/{skill-name}/
├── SKILL.md (required)
├── reference.md (optional)
├── examples.md (optional)
├── scripts/ (optional)
│   └── helper.py
└── templates/ (optional)
    └── template.txt
```

### ✅ Verification Checklist

- [ ] YAML frontmatter exists
- [ ] `name`: kebab-case, ≤64 chars, gerund style preferred
- [ ] `description`: Explains capability + trigger keywords
- [ ] `allowed-tools`: Lists only tools required by the workflow
- [ ] Title (# {Skill Title}) exists
- [ ] Include purpose section

### ✨ Anthropic best practices (2024-12)

- **Stay concise**: Keep SKILL.md lean so Claude reads only what it needs. Move lengthy procedures into referenced files and keep body <500 lines.  
  Source: [Skill authoring best practices](https://docs.claude.com/en/docs/agents-and-tools/agent-skills/best-practices)
- **Progressive disclosure**: Link reference files (reference.md, examples.md, scripts/) directly from SKILL.md and avoid multi-hop chains. Claude loads each file on demand.
- **Right-sized guidance**: Match specificity to risk—high-level checklists for flexible tasks, prescriptive scripts for fragile flows.
- **Consistent naming**: Use gerund or action-oriented names (“Processing PDFs”) and include trigger phrases in descriptions to improve discovery.
- **Test across models**: Validate behavior with the models you plan to run (Haiku, Sonnet, Opus) to ensure instructions are neither too sparse nor verbose.
- **Security posture**: Audit bundled scripts, restrict `allowed-tools`, and document any prerequisites or packages.

### 📂 Discovery rules

- Personal skills live in `~/.claude/skills/{skill-name}/SKILL.md`
- Project skills live in `.claude/skills/{skill-name}/SKILL.md` and should be committed to git
- Claude expects each skill as a first-level directory under `skills/`; nested categories like `.claude/skills/domain/backend/SKILL.md` are **not** auto-discovered (per [Agent Skills - Claude Docs](https://docs.claude.com/en/docs/claude-code/skills))

---

## 4. Plugin setup guide

### 📐 File Structure

**Location**: `.claude/settings.json` (mcpServers section)

**Basic structure**:
```json
{
  "mcpServers": {
    "{plugin-name}": {
      "command": "npx",
      "args": ["-y", "{plugin-package}"],
      "env": {
        "API_KEY": "${API_KEY}"
      }
    }
  }
}
```

### 🌟 Recommended Plugin

| Plugin | Use | Utilizing MoAI-ADK |
|--------|------|--------------|
| **@modelcontextprotocol/server-github** | GitHub API | Automatically generate PR/Issue |
| **@modelcontextprotocol/server-filesystem** | file system | `.moai/` safe access |
| **@modelcontextprotocol/server-brave-search** | web search | See technical documentation |

### 🔒 Security Principles

#### Essential checklist
- [ ] Use of environment variables (no hardcoding)
- [ ] Restrict paths (Filesystem MCP)
- [ ] Minimum privileges
- [ ] Block sensitive information (`.env`, `secrets/`)
- [ ] Source reliability (official or verified plugin)

#### Secure settings
```json
{
  "mcpServers": {
    "github": {
      "env": {
"GITHUB_TOKEN": "${GITHUB_TOKEN}" // ✅ Environment variable
      }
    },
    "filesystem": {
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
"${CLAUDE_PROJECT_DIR}/.moai", // ✅ Restricted path
        "${CLAUDE_PROJECT_DIR}/src"
      ]
    }
  }
}
```

### ✅ Verification Checklist

- [ ] No JSON syntax errors
- [ ] Use of environment variables (no hard coding)
- [ ] Restrictions on file system path
- [ ] Completed setting of necessary environment variables

---

## 5. Settings Setting Guide

### 📐 File Structure

**Location**: `.claude/settings.json`

**Main sections**:
```json
{
  "permissions": {
    "defaultMode": "default",
    "allow": [...],
    "ask": [...],
    "deny": [...]
  },
  "hooks": {
    "SessionStart": [...],
    "PreToolUse": [...]
  },
  "mcpServers": {...}
}
```

### 🔒 3-level permission management

#### 1. `allow` - Automatic approval
Allow only safe and essential tools:
```json
"allow": [
  "Read",
  "Write",
  "Edit",
  "Grep",
  "Glob",
  "Bash(git:*)",
  "Bash(pytest:*)"
]
```

#### 2. `ask` - User confirmation
Important or potentially changeable actions:
```json
"ask": [
  "Bash(git push:*)",
  "Bash(pip install:*)",
  "Bash(rm:*)"
]
```

#### 3. `deny` - Absolutely prohibited
Block dangerous or sensitive operations:
```json
"deny": [
  "Read(./.env)",
  "Read(./secrets/**)",
  "Bash(sudo:*)",
  "Bash(rm -rf:*)"
]
```

### 🪝 Hook system

#### SessionStart hook
Display project information when session starts:
```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "command": "node $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/session-notice.cjs",
            "type": "command"
          }
        ],
        "matcher": "*"
      }
    ]
  }
}
```

#### PreToolUse hook
Verify and block before executing the tool:
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "hooks": [
          {
            "command": "node $CLAUDE_PROJECT_DIR/.claude/hooks/alfred/pre-write-guard.cjs",
            "type": "command"
          }
        ],
        "matcher": "Edit|Write|MultiEdit"
      }
    ]
  }
}
```

### ✅ Verification Checklist

- [ ] No JSON syntax errors
- [ ] `allow`: Only essential tools
- [ ] `ask`: Critical tasks
- [ ] `deny`: Sensitive files/commands
- [ ] Bash pattern refinement (`Bash(git:*)`)
- [ ] Hook file existence and execution permissions

---

## 📊 Best Practices

### Common principles

1. **Principle of least privilege**
 - Specify only necessary tools
 - Specific patterns when using Bash

2. **Compliant with official standards**
 - YAML frontmatter required fields
 - Filename convention (kebab-case)

3. **Security priority**
 - Manage sensitive information as environmental variable
 - Block dangerous operations

4. **Documentation**
 - Clear explanation
 - Concrete examples
 - Verification methods

---

**Last update**: 2025-10-19
**Author**: @Alfred
