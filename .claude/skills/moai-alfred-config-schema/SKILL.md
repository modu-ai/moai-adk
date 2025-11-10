---
name: moai-alfred-config-schema
version: 2.0.0
created: 2025-11-02
updated: 2025-11-11
status: active
description: ".moai/config.json official schema documentation, structure validation, project metadata, language settings, and configuration migration guide. Use when setting up project configuration or understanding config.json structure."
keywords: ['config-schema', 'configuration-validation', 'project-metadata', 'language-settings', 'migration-guide', 'research', 'best-practices', 'configuration-optimization', 'schema-evolution']
allowed-tools: "Read, Grep, AskUserQuestion, TodoWrite"
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

## Research Integration

### Configuration Best Practices Research Capabilities

**Schema Evolution Research**:
- **Configuration pattern analysis**: Study how teams use different configuration options and identify common patterns
- **Migration path optimization**: Research on effective upgrade strategies and backward compatibility preservation
- **Validation effectiveness**: Research on optimal validation methods and error messaging for different configuration types
- **Schema versioning strategies**: Analysis of different versioning approaches and their impact on usability

**Best Practices Research Areas**:
- **Project configuration optimization**: Research on optimal configuration patterns for different project sizes and types
- **Language setting effectiveness**: Study on conversation language adoption and its impact on development efficiency
- **Git strategy optimization**: Research on optimal workflow strategies for different team sizes and collaboration styles
- **Report generation control**: Analysis of optimal report generation settings for different project requirements

**Research Methodology**:
- **Configuration adoption tracking**: Monitor how teams adopt different configuration options and identify trends
- **Migration success rate analysis**: Study the effectiveness of different migration strategies and their success rates
- **Configuration impact measurement**: Correlate configuration choices with project success metrics
- **User experience research**: Study the usability of different configuration approaches and identify improvement opportunities

### Configuration Research Framework

#### 1. Schema Optimization Research
- **Field usage patterns**: Research on which configuration fields are most/least commonly used
- **Validation improvement**: Research on optimal validation rules and error messaging
- **Schema evolution strategies**: Analysis of different schema evolution approaches and their impact
- **Configuration default optimization**: Research on optimal default values for different project types

#### 2. Migration Research
- **Migration effectiveness**: Study on different migration approaches and their success rates
- **Backward compatibility preservation**: Research on optimal strategies for preserving compatibility during upgrades
- **Migration automation potential**: Analysis of opportunities for automating migration processes
- **Version correlation studies**: Research on relationship between schema versions and project characteristics

#### 3. Project Type Research
```
Configuration Research Framework:
├── Usage Pattern Analysis
│   ├── Configuration adoption tracking
│   ├── Field popularity analysis
│   ├── Team size correlation
│   └── Project type patterns
├── Optimization Research
│   ├── Default value optimization
│   ├── Validation improvement
│   ├── Schema evolution
│   └── Migration strategies
└── Best Practices Development
        ├── Configuration templates
        ├── Migration guides
        ├── Validation rules
        └── Success metrics
```

**Current Research Focus Areas**:
- Configuration optimization for different team sizes and project types
- Migration path improvement strategies
- Validation method enhancement for better user experience
- Schema evolution patterns and best practices
- Configuration impact on development efficiency

---

## Integration with Research System

The configuration schema system integrates with MoAI-ADK's research framework by:

1. **Collecting configuration data**: Track how teams configure their projects and identify common patterns and best practices
2. **Validating schema evolution**: Provide real-world testing ground for new schema features and configuration options
3. **Documenting migration patterns**: Capture successful migration strategies and share them across the organization
4. **Benchmarking configuration approaches**: Measure the effectiveness of different configuration strategies and identify improvements

**Research Collaboration**:
- **Migration research team**: Share data on migration effectiveness and success patterns
- **Validation optimization team**: Provide insights on validation methods and error messaging improvements
- **Schema evolution team**: Collaborate on schema development and versioning strategies
- **User experience team**: Study configuration usability and identify improvement opportunities

---

Learn more in `reference.md` for complete schema reference, validation rules, and migration examples.

**Related Skills**: moai-foundation-git, moai-foundation-specs
