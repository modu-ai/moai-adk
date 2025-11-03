---
name: moai-alfred-config-advanced
description: Advanced Claude Code settings and configuration guide for MoAI-ADK
category: alfred-config
freedom_level: 3
---

<!-- @DOC:CLAUDE-PHILOSOPHY-001 -->

# Advanced Claude Code Configuration Guide

Comprehensive guide for configuring Claude Code settings in MoAI-ADK projects.

## Overview

This Skill documents the **three configuration files** that control MoAI-ADK behavior:

1. **`.claude/settings.json`**: Claude Code hooks, permissions, and environment
2. **`.moai/config.json`**: MoAI-ADK project settings
3. **`src/moai_adk/templates/`**: Package templates for new projects

## 1. .claude/settings.json (Local)

### Role
Claude Code's hook execution, permission rules, and environment configuration.

### Key Sections

#### 1.1 Hooks Configuration

```json
{
  "hooks": {
    "sessionStart": {
      "enabled": true,
      "timeout": 5000,
      "script": ".claude/hooks/session-start.py"
    },
    "preToolUse": {
      "enabled": true,
      "timeout": 5000,
      "script": ".claude/hooks/pre-tool-use.py"
    },
    "userPromptSubmit": {
      "enabled": true,
      "timeout": 5000,
      "script": ".claude/hooks/user-prompt-submit.py"
    },
    "postToolUse": {
      "enabled": true,
      "timeout": 5000,
      "script": ".claude/hooks/post-tool-use.py"
    },
    "sessionEnd": {
      "enabled": true,
      "timeout": 5000,
      "script": ".claude/hooks/session-end.py"
    }
  }
}
```

**Hook Execution Flow**:
```
Session Start
    ↓
[sessionStart hook]
    ↓
User Submits Prompt
    ↓
[userPromptSubmit hook]
    ↓
Before Tool Execution
    ↓
[preToolUse hook]
    ↓
Tool Executes
    ↓
[postToolUse hook]
    ↓
Session End
    ↓
[sessionEnd hook]
```

#### 1.2 Permissions Configuration

```json
{
  "permissions": {
    "allow": [
      "Bash(git status:*)",
      "Bash(git log:*)",
      "Bash(git diff:*)",
      "Bash(git show:*)",
      "Bash(git branch:*)",
      "Bash(git tag:*)",
      "Bash(git config:*)",
      "Bash(ls:*)",
      "Bash(pwd:*)",
      "Bash(echo:*)",
      "Bash(cat:*)",
      "Bash(which:*)",
      "Bash(python --version:*)",
      "Bash(python -m pytest:*)",
      "Bash(uv --version:*)",
      "Bash(ruff check:*)",
      "Bash(ruff format --check:*)",
      "Bash(wc -l:*)"
    ],
    "ask": [
      "Bash(git add:*)",
      "Bash(git commit:*)",
      "Bash(git checkout:*)",
      "Bash(git merge:*)",
      "Bash(git reset:*)",
      "Bash(git push:*)",
      "Bash(git pull:*)",
      "Bash(uv add:*)",
      "Bash(uv remove:*)",
      "Bash(pip install:*)",
      "Bash(rm:*)",
      "Bash(rm -rf:*)",
      "Bash(gh pr merge:*)"
    ],
    "deny": [
      "Bash(git push --force:*)",
      "Bash(git reset --hard:*)",
      "Bash(rm -rf /:*)",
      "Bash(chmod -R 777:*)",
      "Bash(dd:*)",
      "Bash(mkfs:*)",
      "Bash(format:*)",
      "Bash(shutdown:*)",
      "Bash(reboot:*)",
      "Read(.env:*)",
      "Write(.env:*)",
      "Read(secrets.json:*)",
      "Write(secrets.json:*)"
    ]
  }
}
```

**Permission Priority**:
```
1. deny (highest) - Always blocked
2. ask (middle) - User confirmation required
3. allow (lowest) - Automatic approval
```

**Pattern Specificity Rule**:
- More specific patterns override general patterns
- Example: `Bash(git push --force:*)` in deny overrides `Bash(git push:*)` in ask

### Recommendations

#### Best Practices

1. **Granular Git Commands**
   - ✅ Use specific commands: `Bash(git status:*)`, `Bash(git log:*)`
   - ❌ Avoid wildcard: `Bash(git:*)` (too broad)

2. **Dangerous Commands in Deny**
   - ✅ `git push --force`, `git reset --hard`, `rm -rf /`
   - Prevents accidental destructive operations

3. **Read-Only in Allow**
   - ✅ `git status`, `git log`, `git diff`, `ls`, `pwd`
   - Safe, frequently used, no side effects

4. **Modifications in Ask**
   - ✅ `git commit`, `git push`, `uv add`, `pip install`
   - Requires user confirmation before execution

5. **Secrets in Deny**
   - ✅ `.env`, `secrets.json`, `credentials.yaml`
   - Prevents accidental exposure

#### Common Patterns

**Development Workflow**:
```json
{
  "allow": [
    "Bash(git status:*)",
    "Bash(git diff:*)",
    "Bash(python -m pytest:*)",
    "Bash(ruff check:*)"
  ],
  "ask": [
    "Bash(git add:*)",
    "Bash(git commit:*)",
    "Bash(git push:*)"
  ]
}
```

**Package Management**:
```json
{
  "allow": [
    "Bash(uv --version:*)",
    "Bash(pip list:*)"
  ],
  "ask": [
    "Bash(uv add:*)",
    "Bash(uv remove:*)",
    "Bash(pip install:*)"
  ]
}
```

## 2. .moai/config.json (Local)

### Role
MoAI-ADK project settings, GitFlow strategy, and TRUST 5 principles.

### Key Sections

#### 2.1 Language Configuration

```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  }
}
```

**Purpose**:
- Controls user-facing content language
- Not a Claude Code native setting (MoAI-ADK custom field)
- Read by hooks and passed to sub-agents

**Usage Example**:
```python
import json
from pathlib import Path

config = json.loads(Path(".moai/config.json").read_text())
lang = config["language"]["conversation_language"]  # "ko"
```

#### 2.2 Project Metadata

```json
{
  "project": {
    "name": "MoAI-ADK",
    "version": "0.15.2",
    "mode": "team",
    "codebase_language": "python",
    "description": "SPEC-First TDD Development Framework with Alfred Super-Agent"
  }
}
```

#### 2.3 Git Strategy

```json
{
  "git_strategy": {
    "workflow": "gitflow",
    "main_branch": "main",
    "develop_branch": "develop",
    "feature_prefix": "feature/",
    "hotfix_prefix": "hotfix/",
    "release_prefix": "release/"
  }
}
```

#### 2.4 Hooks Configuration

```json
{
  "hooks": {
    "timeout": 5,
    "enabled": {
      "session_start": true,
      "pre_tool_use": true,
      "user_prompt_submit": true,
      "post_tool_use": true,
      "session_end": true
    }
  }
}
```

#### 2.5 TAG Scan Policy

```json
{
  "tags": {
    "scan_enabled": true,
    "scan_patterns": [
      "@SPEC:*",
      "@TEST:*",
      "@CODE:*",
      "@DOC:*"
    ],
    "scan_directories": [
      ".moai/specs/",
      "src/",
      "tests/",
      "docs/"
    ]
  }
}
```

#### 2.6 TRUST 5 Principles

```json
{
  "constitution": {
    "trust_5": {
      "test_first": true,
      "readable": true,
      "unified": true,
      "secured": true,
      "trackable": true
    },
    "coverage_target": 80,
    "quality_gate": {
      "tests_pass": true,
      "coverage_check": true,
      "linting_pass": true,
      "type_check": true
    }
  }
}
```

### Configuration Example

**Complete `.moai/config.json`**:
```json
{
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  },
  "project": {
    "name": "MoAI-ADK",
    "version": "0.15.2",
    "mode": "team",
    "codebase_language": "python",
    "description": "SPEC-First TDD Development Framework"
  },
  "git_strategy": {
    "workflow": "gitflow",
    "main_branch": "main",
    "develop_branch": "develop",
    "feature_prefix": "feature/",
    "hotfix_prefix": "hotfix/",
    "release_prefix": "release/"
  },
  "hooks": {
    "timeout": 5,
    "enabled": {
      "session_start": true,
      "pre_tool_use": true,
      "user_prompt_submit": true,
      "post_tool_use": true,
      "session_end": true
    }
  },
  "tags": {
    "scan_enabled": true,
    "scan_patterns": ["@SPEC:*", "@TEST:*", "@CODE:*", "@DOC:*"],
    "scan_directories": [".moai/specs/", "src/", "tests/", "docs/"]
  },
  "constitution": {
    "trust_5": {
      "test_first": true,
      "readable": true,
      "unified": true,
      "secured": true,
      "trackable": true
    },
    "coverage_target": 80,
    "quality_gate": {
      "tests_pass": true,
      "coverage_check": true,
      "linting_pass": true,
      "type_check": true
    }
  }
}
```

## 3. src/moai_adk/templates/ (Package Templates)

### Role
Templates used when creating new projects with `moai init`.

### Key Files

#### 3.1 .claude/settings.json Template

**Location**: `src/moai_adk/templates/.claude/settings.json`

**Purpose**: Hook and permission template for new projects

**Variables**:
- No template variables (static configuration)
- Copied as-is to new projects

#### 3.2 .moai/config.json Template

**Location**: `src/moai_adk/templates/.moai/config.json`

**Purpose**: Project settings template

**Variables**:
```json
{
  "language": {
    "conversation_language": "{{CONVERSATION_LANGUAGE}}",
    "conversation_language_name": "{{CONVERSATION_LANGUAGE_NAME}}"
  },
  "project": {
    "name": "{{PROJECT_NAME}}",
    "version": "{{PROJECT_VERSION}}",
    "mode": "{{PROJECT_MODE}}"
  }
}
```

**Substitution Example**:
```
{{CONVERSATION_LANGUAGE}} → "ko"
{{CONVERSATION_LANGUAGE_NAME}} → "한국어"
{{PROJECT_NAME}} → "my-project"
{{PROJECT_VERSION}} → "0.1.0"
{{PROJECT_MODE}} → "solo" or "team"
```

#### 3.3 CLAUDE.md Template

**Location**: `src/moai_adk/templates/CLAUDE.md`

**Purpose**: Project instruction template (English)

**Variables**:
```
{{PROJECT_NAME}}
{{CONVERSATION_LANGUAGE}}
{{CONVERSATION_LANGUAGE_NAME}}
```

### Synchronization Strategy

**Important**: Package templates must be updated when local settings change.

#### When to Sync

✅ **Sync Required**:
- New hook added to `.claude/settings.json`
- Permission rules updated significantly
- New TRUST 5 principle added
- New Skill or agent added

❌ **Sync NOT Required**:
- Local project metadata changes
- Project-specific CLAUDE.md edits (Korean version)
- Temporary settings adjustments

#### How to Sync

**Step 1: Identify Changes**
```bash
# Compare local and template
diff .claude/settings.json src/moai_adk/templates/.claude/settings.json
```

**Step 2: Update Template**
```bash
# Copy local to template (if intentional improvement)
cp .claude/settings.json src/moai_adk/templates/.claude/settings.json
```

**Step 3: Test Template**
```bash
# Create test project
moai init test-project --language ko

# Verify settings applied correctly
cat test-project/.claude/settings.json
```

**Step 4: Commit Changes**
```bash
git add src/moai_adk/templates/
git commit -m "sync: Update package templates with improved settings"
```

## Troubleshooting

### Common Issues

#### Issue 1: Hook Timeout

**Symptom**: "Hook execution timed out after 5000ms"

**Solution**:
```json
{
  "hooks": {
    "sessionStart": {
      "timeout": 10000  // Increase to 10s
    }
  }
}
```

#### Issue 2: Permission Denied

**Symptom**: "Permission denied: Bash(git push:*)"

**Solution**:
```json
{
  "permissions": {
    "ask": [
      "Bash(git push:*)"  // Move from deny to ask
    ]
  }
}
```

#### Issue 3: conversation_language Not Working

**Symptom**: Alfred responds in English instead of Korean

**Solution**:
```json
{
  "language": {
    "conversation_language": "ko",  // Ensure correct
    "conversation_language_name": "한국어"
  }
}
```

**Verify**:
```bash
# Check config
cat .moai/config.json | jq '.language'

# Restart session
# conversation_language is read on SessionStart hook
```

## Integration with Alfred Workflow

### When to Invoke This Skill

Alfred should invoke `Skill("moai-alfred-config-advanced")` when:

1. **User asks about settings**: "How do I configure permissions?", "What's in .moai/config.json?"
2. **Permission errors**: Recurring permission denied issues
3. **Hook failures**: Debugging hook timeout or execution errors
4. **Template updates**: Syncing local changes to package templates

### Example Invocation

```markdown
User: "I want to allow git push without confirmation"

Alfred: Let me guide you through updating permissions.
→ Skill("moai-alfred-config-advanced")
→ Explains: Move "Bash(git push:*)" from ask to allow
→ Shows: Example .claude/settings.json update
→ Warns: Security implications of auto-allowing push
```

## Related Skills

- `Skill("moai-alfred-session-analytics")`: Session log analysis for settings optimization
- `Skill("moai-alfred-autofixes")`: Automatic error fixing protocols
- `Skill("moai-alfred-dev-guide")`: Development guidelines

## References

- **Settings Location**: `.claude/settings.json`
- **Config Location**: `.moai/config.json`
- **Templates**: `src/moai_adk/templates/`
- **Hook Scripts**: `.claude/hooks/*.py`

---

**Note**: Always sync package templates after improving local settings to ensure new projects benefit from improvements.
