---
name: moai-alfred-config-schema
description: ".moai/config.json official schema documentation, structure validation, project metadata, language settings, and configuration migration guide. Use when setting up project configuration or understanding config.json structure."
allowed-tools: "Read, Grep"
---

## What It Does

`.moai/config.json` 파일의 공식 스키마와 각 필드의 목적, 유효한 값, 마이그레이션 규칙을 정의합니다.

## When to Use

- ✅ Project 초기화 후 config.json 설정
- ✅ config.json 스키마 이해
- ✅ Language settings, git strategy, TAG configuration 변경
- ✅ Legacy config 마이그레이션

## Schema Overview

```json
{
  "version": "0.17.0",
  "project": {
    "name": "ProjectName",
    "codebase_language": "python",
    "conversation_language": "ko",
    "conversation_language_name": "Korean"
  },
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "Korean"
  },
  "git": {
    "strategy": "github-pr",
    "main_branch": "main",
    "protected": true,
    "auto_delete_branches": true,
    "spec_git_workflow": "feature_branch"
  },
  "report_generation": {
    "enabled": true,
    "auto_create": false,
    "warn_user": true
  },
  "tag": {
    "prefix_style": "DOMAIN-###"
  }
}
```

## Top-Level Sections

- **version**: Configuration version (do not edit)
- **project**: Name, codebase language, conversation language
- **language**: Multi-language support settings
- **git**: GitHub workflow strategy, branch auto-cleanup, SPEC workflow mode
- **report_generation**: Document generation frequency control (v0.17.0+)
- **tag**: TAG system configuration

---

## v0.17.0 New Features

### 1. Report Generation Control

**Section**: `report_generation` (new in v0.17.0)

Controls automatic document generation frequency to manage token usage:

```json
{
  "report_generation": {
    "enabled": true,
    "auto_create": false,
    "warn_user": true,
    "user_choice": "Minimal",
    "allowed_locations": [
      ".moai/docs/",
      ".moai/reports/",
      ".moai/analysis/",
      ".moai/specs/SPEC-*/"
    ]
  }
}
```

**Fields**:
- `enabled` (boolean): Turn report generation on/off (0 tokens when disabled)
- `auto_create` (boolean): Auto-create full reports vs essential only
- `warn_user` (boolean): Show token/time warnings in surveys
- `user_choice` (string): User's selected level (Enable/Minimal/Disable)
- `allowed_locations` (array): Where auto-generated reports are placed

**Token Savings**:
- Enable (full reports): 50-60 tokens per report
- Minimal (essential only): 20-30 tokens per report
- Disable (no reports): 0 tokens (100% savings)

### 2. SPEC Git Workflow Mode (Team Mode)

**Section**: `git.spec_git_workflow` (new in v0.17.0)

Controls how features are developed in team mode:

```json
{
  "git": {
    "spec_git_workflow": "feature_branch"
  }
}
```

**Valid Values**:
1. **`feature_branch`** (recommended for team mode)
   - Create feature/SPEC-{ID} branch
   - PR to develop with review
   - Auto-merge after approval

2. **`develop_direct`** (recommended for rapid development)
   - Skip branching, commit directly to develop
   - No PR review needed
   - Fastest path to integration

3. **`per_spec`** (maximum flexibility)
   - Ask user per SPEC which workflow to use
   - Combine both approaches in single project

---

Learn more in `reference.md` for complete schema reference, validation rules, and migration examples.

**Related Skills**: moai-foundation-git, moai-foundation-specs
