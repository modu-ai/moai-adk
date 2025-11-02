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
  "version": "0.7.0",
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
    "protected": true
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
- **git**: GitHub workflow strategy
- **tag**: TAG system configuration

---

Learn more in `reference.md` for complete schema reference, validation rules, and migration examples.

**Related Skills**: moai-alfred-dev-guide, moai-foundation-specs
