# Claude Code collection of practical examples

> **Actual working example code**
>
> Proven examples that can be copied and used immediately

---

## 📋 Table of Contents

1. [Agent example](#1-agent-example)
2. [Command example](#2-command-example)
3. [Skill example](#3-skill-example)
4. [Plugin example](#4-plugin-example)
5. [Settings example](#5-settings-example)

---

## 1. Agent example

### Example 1: spec-builder (MoAI-ADK)

**File**: `.claude/agents/alfred/spec-builder.md`

```markdown
---
name: spec-builder
description: "Use when: When writing SPEC, EARS specification, and requirements analysis are necessary."
tools: Read, Write, Edit, MultiEdit, Grep, Glob, Bash, TodoWrite
model: sonnet
---

# SPEC Builder 🏗️ - System Architect

**MoAI-ADK SPEC writing expert**

## 🎭 Agent Persona

**Icon**: 🏗️
**Job**: System Architect
**Area of ​​Expertise**: SPEC writing, EARS specification, requirements analysis
**Role**: Convert business requirements into structured SPECs
**Goal**: Create clear and testable SPEC documents.

## 🎯 Key Role

### 1. Create SPEC document
- Apply EARS 5 syntax
- YAML Front Matter 7 required fields
- HISTORY section management

### 2. Auto-execution conditions
- When executing the `/alfred:1-plan` command
- When a new function requirement occurs
- When requesting improvement of the existing SPEC

## 📐 Workflow

### STEP 1: Project document analysis
\`\`\`bash
# read product.md
Read .moai/project/product.md

# Check existing SPEC
ls .moai/specs/SPEC-*/spec.md
\`\`\`

### STEP 2: Draft SPEC
\`\`\`bash
Write .moai/specs/SPEC-{ID}/spec.md
\`\`\`

## 💡 User Guide
\`\`\`bash
@agent-spec-builder "Create a user authentication function SPEC"
\`\`\`
```

### Example 2: code-analyzer (simple analysis agent)

**File**: `.claude/agents/code-analyzer.md`

```markdown
---
name: code-analyzer
description: Use when: code quality analysis is needed
tools: Read, Grep, Bash(rg:*)
model: haiku
---

# Code Analyzer 🔍 - Quality Manager

Analyze code quality metrics and provide improvement suggestions.

## 🎭 Agent Persona

**Icon**: 🔍
**Job**: Quality Manager
**Area of ​​Expertise**: Code quality analysis, linting, test coverage
**Role**: Measure code quality metrics and suggest improvements
**Goal**: Maintain high code quality

## 🎯 Key Role

### 1. Collect quality indicators
- Check test coverage
- Detect linting errors
- Complexity analysis

### 2. Conditions for automatic execution
- When quality verification is requested before commit
- Automatic verification before creating PR

## 📐 Workflow

### STEP 1: Scan code
\`\`\`bash
rg "@TEST:" -n tests/
pytest --cov
\`\`\`

### STEP 2: Generate report
- Coverage: {percentage}%
- Linting errors: {count}

## 💡 User Guide
\`\`\`bash
@agent-code-analyzer "Analyze current project quality"
\`\`\`
```

---

## 2. Command example

### Example 1: /alfred:1-plan (MoAI-ADK)

**File**: `.claude/commands/alfred/1-spec.md`

```markdown
---
name: alfred:1-spec
description: Project document analysis and SPEC creation (EARS method)
argument-hint: [feature-description]
allowed-tools:
  - Read
  - Write
  - Edit
  - Task
  - Grep
  - Glob
  - TodoWrite
  - Bash(git:*)
---

# 📋 MoAI-ADK Step 1: Write SPEC

## 🎯 Command Purpose

Analyze project documents, create SPEC documents using the EARS method, and create Git branches.

## 💡 Example of use
\`\`\`bash
/alfred:1-plan "User authentication function"
\`\`\`

## 📋 Execution flow

1. **Phase 1**: Analyze project document
 - Read product.md
 - Check existing SPEC

2. **Phase 2**: SPEC creation
 - EARS method specification
 - Git branch creation

## 🔗 Associated Agent

- **Primary**: spec-builder (🏗️ System Architect)
- **Secondary**: git-manager (🚀 Release Engineer)

## ⚠️ Precautions

- SPEC ID duplication check required
- Compliance with EARS 5 phrases

## 📋 Next steps

- `/alfred:2-run SPEC-{ID}` - Start TDD implementation
```

### Example 2: /deploy-api (deploy command)

**File**: `.claude/commands/deploy-api.md`

```markdown
---
name: deploy-api
description: Deploy API server to production environment
argument-hint: [environment]
allowed-tools:
  - Read
  - Bash(git:*)
  - Bash(npm:*)
---

# 🚀 API deployment command

Deploys the API server to the specified environment.

## 🎯 Command Purpose

Automate Git tag creation and deployment by environment

## 💡 Example of use
\`\`\`bash
/deploy-api production
/deploy-api staging
\`\`\`

## 📋 Execution flow

1. **Phase 1**: Check Git status
 - Check current branch (main required)
 - Create version tag

2. **Phase 2**: Deployment execution
   - npm run build
- Execute deployment script for each environment

## ⚠️ Precautions

- Can only be run from the main branch
- Must pass all tests
```

---

## 3. Skill example

### Example 1: moai-alfred-tag-scanning

**File**: `.claude/skills/moai-alfred-tag-scanning/SKILL.md`

```markdown
---
name: moai-alfred-tag-scanning
description: Directly scan TAG markers and create inventory (CODE-FIRST principle)
model: haiku
allowed-tools:
  - Grep
  - Read
---

# TAG scanner

> CODE-FIRST principle: scan code directly without intermediate cache

## 🎯 Purpose

Scan `@SPEC`, `@TEST`, `@CODE`, and `@DOC` TAGs directly from your code.

## 💡 How to use

“Check AUTH domain TAG list”

## 📋 How to scan

\`\`\`bash
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
\`\`\`

## ✅ Verification

- Does every `@CODE` TAG have a corresponding `@SPEC`?
- No orphan TAGs
```

### Example 2: moai-alfred-feature-selector

**File**: `.claude/skills/moai-alfred-feature-selector/SKILL.md`

```markdown
---
name: moai-alfred-feature-selector
description: Select optimal functions for each project type (37 skills → automatically filter 3 to 5)
model: haiku
allowed-tools:
  - Read
---

# MoAI Alfred Feature Selector

> Automatic selection of MoAI-ADK functions suitable for project characteristics

## 🎯 Purpose

Analyze your project type and select only the features you need.

## 📋 Project classification

### By language
- **Python**: pytest, mypy, ruff
- **TypeScript**: Vitest, Biome

### Domain-specific
- **CLI Tool**: Argument parsing, POSIX compliance
- **Web API**: REST/GraphQL, authentication

## 💡 How to use

"Automatically called when running /alfred:0-project"
```

---

## 4. Plugin example

### Example 1: GitHub + Filesystem (default)

**FILE**: `.claude/settings.json` (mcpServers section)

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "filesystem": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
        "${CLAUDE_PROJECT_DIR}/src",
        "${CLAUDE_PROJECT_DIR}/tests"
      ]
    }
  }
}
```

### Example 2: Fully configured MoAI-ADK

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "moai-filesystem": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
        "${CLAUDE_PROJECT_DIR}/.moai",
        "${CLAUDE_PROJECT_DIR}/src",
        "${CLAUDE_PROJECT_DIR}/tests",
        "${CLAUDE_PROJECT_DIR}/docs"
      ]
    },
    "brave-search": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "env": {
        "BRAVE_API_KEY": "${BRAVE_API_KEY}"
      }
    }
  }
}
```

---

## 5. Settings example

### Example 1: Python project

**File**: `.claude/settings.json`

```json
{
  "permissions": {
    "defaultMode": "default",
    "allow": [
      "Read",
      "Write",
      "Edit",
      "Grep",
      "Glob",
      "Bash(git:*)",
      "Bash(python:*)",
      "Bash(pytest:*)",
      "Bash(mypy:*)",
      "Bash(ruff:*)"
    ],
    "ask": [
      "Bash(pip install:*)",
      "Bash(git push:*)"
    ],
    "deny": [
      "Read(./.env)",
      "Bash(sudo:*)",
      "Bash(rm -rf:*)"
    ]
  }
}
```

### Example 2: TypeScript project

```json
{
  "permissions": {
    "defaultMode": "default",
    "allow": [
      "Read",
      "Write",
      "Edit",
      "Grep",
      "Glob",
      "Bash(git:*)",
      "Bash(node:*)",
      "Bash(npm:*)",
      "Bash(pnpm:*)"
    ],
    "ask": [
      "Bash(npm install:*)",
      "Bash(pnpm install:*)",
      "Bash(git push:*)"
    ],
    "deny": [
      "Read(./.env)",
      "Read(./secrets/**)",
      "Bash(sudo:*)",
      "Bash(rm -rf:*)"
    ]
  }
}
```

### Example 3: MoAI-ADK project (with hooks)

```json
{
  "permissions": {
    "defaultMode": "default",
    "allow": [
      "Task",
      "Read",
      "Write",
      "Edit",
      "MultiEdit",
      "Grep",
      "Glob",
      "TodoWrite",
      "Bash(git:*)",
      "Bash(python:*)",
      "Bash(pytest:*)",
      "Bash(mypy:*)",
      "Bash(ruff:*)",
      "Bash(moai-adk:*)",
      "Bash(alfred:*)"
    ],
    "ask": [
      "Bash(git push:*)",
      "Bash(pip install:*)"
    ],
    "deny": [
      "Read(./.env)",
      "Read(./secrets/**)",
      "Bash(sudo:*)",
      "Bash(rm -rf:*)"
    ]
  },
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
    ],
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

---

**Last update**: 2025-10-19
**Author**: @Alfred
