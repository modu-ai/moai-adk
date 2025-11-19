# MoAI-ADK config.json Schema Documentation

> **Purpose**: This document defines the official schema structure for `.moai/config/config.json` file, which serves as the single source of truth for all MoAI-ADK project configuration.

---

## 📋 Schema Overview

The `.moai/config/config.json` file contains all project-level configuration for MoAI-ADK, including:
- Project metadata and settings
- User preferences (language, nickname)
- Git workflow strategies
- TAG system configuration
- Constitution principles

---

## 🏗️ Schema Structure

### Top-Level Sections

```json
{
  "_meta": { /* Metadata about the config file itself */ },
  "moai": { /* MoAI-ADK version info */ },
  "project": { /* Project-specific settings */ },
  "user": { /* User preferences */ },
  "constitution": { /* Development principles */ },
  "git_strategy": { /* Git workflow configuration */ },
  "pipeline": { /* Command pipeline settings */ },
  "tags": { /* TAG system configuration */ }
}
```

---

## 📖 Section Definitions

### 1. `_meta` Section

**Purpose**: Internal metadata for config structure tracking and TAG traceability.

**Schema**:
```json
{
  "_meta": {
  }
}
```

**Fields**:

---

### 2. `moai` Section

**Purpose**: Track MoAI-ADK package version for compatibility checks.

**Schema**:
```json
{
  "moai": {
    "version": "0.4.6"
  }
}
```

**Fields**:
- `version` (string, required): MoAI-ADK version in SemVer format (e.g., "0.4.6")

---

### 3. `project` Section

**Purpose**: Core project settings including name, mode, language preferences, and optimization status.

**Schema**:
```json
{
  "project": {
    "name": "MoAI-ADK",
    "description": "MoAI-Agentic Development Kit",
    "mode": "personal",
    "locale": "ko",
    "conversation_language": "ko",
    "conversation_language_name": "한국어",
    "language": "python",
    "moai_adk_version": "0.4.6",
    "initialized": true,
    "optimized": true,
    "optimized_at": "2025-10-22T12:45:00Z",
    "created_at": "2025-10-21 23:59:19"
  }
}
```

**Fields**:
- `name` (string, required): Project name
- `description` (string, optional): Brief project description
- `mode` (enum, required): Project mode — `"personal"` | `"team"`
- `locale` (string, required): System locale code (e.g., `"ko"`, `"en"`, `"ja"`)
- `conversation_language` (string, required): Language code for all Alfred dialogs and documentation (e.g., `"ko"`, `"en"`, `"ja"`, `"zh"`)
- `conversation_language_name` (string, required): Display name of conversation language (e.g., `"한국어"`, `"English"`)
- `language` (string, required): Primary codebase language (e.g., `"python"`, `"typescript"`)
- `moai_adk_version` (string, required): MoAI-ADK version this project was initialized with
- `initialized` (boolean, required): Whether project initialization completed
- `optimized` (boolean, required): Whether template optimization completed
- `optimized_at` (string, optional): ISO 8601 timestamp of optimization completion
- `created_at` (string, required): Project creation timestamp

**Notes**:
- `conversation_language` is set during `/alfred:0-project` STEP 0.1
- `language` is auto-detected during `/alfred:0-project` STEP 1 or can be manually specified

---

### 4. `user` Section (NEW in v0.4.6)

**Purpose**: User-specific preferences for personalized Alfred experience.

**Schema**:
```json
{
  "user": {
    "nickname": "GOOS오라버니"
  }
}
```

**Fields**:
- `nickname` (string, required): User's chosen nickname for personalized communication
  - Length: 1-50 characters
  - Allows emoji, spaces, and special characters
  - Set during `/alfred:0-project` STEP 0.2
  - Used by all sub-agents to address the user (e.g., "안녕하세요, GOOS오라버니님!")
  - Displayed in `CLAUDE.md` under "## 프로젝트 정보 | Project Information"

**Usage**:
- Alfred and all sub-agents receive `user_nickname` as a context parameter
- Language-specific honorifics are applied automatically (e.g., "님" in Korean)
- Enhances user experience by personalizing all interactions

---

### 5. `constitution` Section

**Purpose**: Define development principles and quality gates enforced by MoAI-ADK.

**Schema**:
```json
{
  "constitution": {
    "enforce_tdd": true,
    "require_tags": true,
    "test_coverage_target": 85,
    "simplicity_threshold": 5,
    "principles": {
      "simplicity": {
        "max_projects": 5,
        "notes": "Default recommendation. Adjust in .moai/config.json or via SPEC/ADR with documented rationale based on project size."
      }
    }
  }
}
```

**Fields**:
- `enforce_tdd` (boolean, required): Whether TDD workflow is mandatory
- `test_coverage_target` (integer, required): Minimum test coverage percentage (default: 85)
- `simplicity_threshold` (integer, required): Max number of concurrent projects (default: 5)
- `principles` (object, optional): Additional principle definitions

---

### 6. `git_strategy` Section

**Purpose**: Configure Git workflow automation based on project mode.

**Schema**:
```json
{
  "git_strategy": {
    "personal": {
      "auto_checkpoint": "disabled",
      "checkpoint_events": ["delete", "refactor", "merge", "script", "critical-file"],
      "checkpoint_type": "local-branch",
      "max_checkpoints": 10,
      "cleanup_days": 7,
      "push_to_remote": false,
      "auto_commit": true,
      "branch_prefix": "feature/",
      "develop_branch": "develop",
      "main_branch": "main"
    },
    "team": {
      "use_gitflow": true,
      "auto_pr": true,
      "draft_pr": true,
      "feature_prefix": "feature/SPEC-",
      "develop_branch": "develop",
      "main_branch": "main"
    }
  }
}
```

**Personal Mode Fields**:
- `auto_checkpoint` (enum): `"disabled"` | `"event-driven"` | `"manual"` (default: disabled)
- `checkpoint_events` (array): List of events triggering auto-checkpoint
- `checkpoint_type` (enum): `"local-branch"` | `"git-stash"` | `"backup-dir"`
- `max_checkpoints` (integer): Max number of checkpoints to retain
- `cleanup_days` (integer): Days before old checkpoints are removed
- `push_to_remote` (boolean): Whether to push checkpoints to remote
- `auto_commit` (boolean): Auto-commit changes after TDD cycles
- `branch_prefix` (string): Prefix for feature branches
- `develop_branch` (string): Name of develop branch
- `main_branch` (string): Name of main/production branch

**Team Mode Fields**:
- `use_gitflow` (boolean): Enable GitFlow branching strategy
- `auto_pr` (boolean): Auto-create PR after feature completion
- `draft_pr` (boolean): Create PRs in draft mode by default
- `feature_prefix` (string): Prefix for feature branches (includes SPEC ID)
- `develop_branch` (string): Name of develop branch
- `main_branch` (string): Name of main/production branch

---

### 7. `pipeline` Section

**Purpose**: Define available Alfred commands and track current workflow stage.

**Schema**:
```json
{
  "pipeline": {
    "available_commands": [
      "/alfred:0-project",
      "/alfred:1-plan",
      "/alfred:2-run",
      "/alfred:3-sync"
    ],
    "current_stage": "initialized"
  }
}
```

**Fields**:
- `available_commands` (array, required): List of available Alfred commands
- `current_stage` (string, required): Current workflow stage (e.g., `"initialized"`, `"planning"`, `"building"`, `"syncing"`)

---

### 8. `tags` Section

**Purpose**: Configure TAG system behavior and storage policy.

**Schema**:
```json
{
  "tags": {
    "storage_type": "code_scan",
    "auto_sync": true,
    "categories": ["REQ", "DESIGN", "TASK", "TEST", "FEATURE", "API", "UI", "DATA"],
    "code_scan_policy": {
      "no_intermediate_cache": true,
      "realtime_validation": true,
      "scan_tools": ["rg", "grep"],
      "philosophy": "The source of truth for TAGs lives in the code itself"
    }
  }
}
```

**Fields**:
- `storage_type` (enum, required): `"code_scan"` | `"registry_file"`
- `auto_sync` (boolean, required): Auto-sync TAGs during `/alfred:3-sync`
- `categories` (array, required): List of valid TAG categories
- `code_scan_policy` (object, required): Configuration for code-first TAG scanning
  - `no_intermediate_cache` (boolean): Never use intermediate TAG cache files
  - `realtime_validation` (boolean): Validate TAGs on every operation
  - `scan_tools` (array): Tools used for TAG scanning (e.g., `["rg", "grep"]`)
  - `scan_command` (string): Command to scan for TAGs
  - `philosophy` (string): Description of TAG philosophy

---

## 🔄 Schema Evolution

### Version History

| Version | Date       | Changes                                                |
| ------- | ---------- | ------------------------------------------------------ |
| v0.4.6  | 2025-10-23 | Added `user` section with `nickname` field             |
| v0.4.2  | 2025-10-22 | Added `conversation_language` and `language_name` fields |
| v0.4.0  | 2025-10-20 | Initial schema documentation                           |

### Migration Guide

**From v0.4.5 → v0.4.6**:
- Add `user` section with `nickname` field
- No breaking changes; existing configs remain compatible

**Example Migration**:
```json
// Before (v0.4.5)
{
  "project": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  }
}

// After (v0.4.6)
{
  "project": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  },
  "user": {
    "nickname": "GOOS오라버니"
  }
}
```

---

## 📚 Related Documentation

- **Template Processing**: `/src/moai_adk/cli/template_processor.py`
- **Project Initialization**: `/.claude/commands/alfred/0-project.md`
- **project-manager Agent**: `/.claude/agents/alfred/project-manager.md`
- **CLAUDE.md Template**: `/src/moai_adk/templates/CLAUDE.md`

---

## ✅ Validation Rules

### Required Fields per Section

**Minimum Valid Config**:
```json
{
  "moai": {
    "version": "0.4.6"
  },
  "project": {
    "name": "MyProject",
    "mode": "personal",
    "locale": "en",
    "conversation_language": "en",
    "conversation_language_name": "English",
    "language": "python",
    "initialized": true
  },
  "user": {
    "nickname": "Developer"
  },
  "constitution": {
    "enforce_tdd": true,
    "require_tags": true,
    "test_coverage_target": 85
  },
  "tags": {
    "storage_type": "code_scan",
    "auto_sync": true
  }
}
```

### Validation Checklist

- [ ] `moai.version` matches installed package version
- [ ] `project.mode` is either `"personal"` or `"team"`
- [ ] `project.conversation_language` is a valid ISO 639-1 code
- [ ] `user.nickname` length is between 1-50 characters
- [ ] `constitution.test_coverage_target` is between 0-100
- [ ] `tags.storage_type` is valid enum value
- [ ] All required fields are present

---

## 🛠️ Usage Examples

### Reading Config in Python

```python
import json
from pathlib import Path

def read_config():
    config_path = Path(".moai/config/config.json")
    with open(config_path, "r", encoding="utf-8") as f:
        return json.load(f)

config = read_config()
user_nickname = config.get("user", {}).get("nickname", "Developer")
conversation_language = config["project"]["conversation_language"]
```

### Updating User Nickname

```python
import json
from pathlib import Path

def update_user_nickname(new_nickname: str):
    config_path = Path(".moai/config/config.json")

    # Read existing config
    with open(config_path, "r", encoding="utf-8") as f:
        config = json.load(f)

    # Update user section
    if "user" not in config:
        config["user"] = {}
    config["user"]["nickname"] = new_nickname

    # Write back
    with open(config_path, "w", encoding="utf-8") as f:
        json.dump(config, f, indent=2, ensure_ascii=False)
```

---

## 📝 Notes

- **SSOT**: This config file is the single source of truth for all project settings
- **No Duplication**: Never duplicate config values in code or other files
- **Encoding**: Always use UTF-8 encoding to support international characters in nicknames and language names
- **Validation**: Always validate config structure before reading values
- **Migration**: When adding new fields, provide sensible defaults for backwards compatibility

---

**Last Updated**: 2025-10-23
**Version**: v0.4.6
**Maintainer**: cc-manager
