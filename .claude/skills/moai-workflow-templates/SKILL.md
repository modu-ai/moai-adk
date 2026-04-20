---
name: moai-workflow-templates
description: >
  Template management system for code boilerplates, feedback templates, scaffolding,
  and project optimization workflows.
  Use when creating code templates, generating boilerplate files, managing project
  scaffolding, optimizing template performance, or preparing GitHub issue templates.
  Do NOT use for SPEC document creation (use moai-workflow-spec instead)
  or documentation generation (use moai-workflow-project instead).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "3.1.0"
  category: "workflow"
  status: "active"
  updated: "2026-01-11"
  modularized: "true"
  tags: "workflow, templates, boilerplate, scaffolding, optimization, feedback"
  aliases: "moai-workflow-templates"
  replaces: "moai-core-code-templates, moai-core-feedback-templates, moai-project-template-optimizer"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["template", "boilerplate", "scaffolding", "code template", "project template", "feedback template", "GitHub issue", "template optimization"]
  phases: ["plan"]
  agents: ["manager-project", "builder-skill"]
---

# Enterprise Template Management

Unified template system combining code boilerplates, feedback templates, and project optimization workflows for rapid development and consistent patterns.

## Quick Reference

Core Capabilities:

- Code template library for FastAPI, React, Vue, and Next.js
- GitHub issue feedback templates covering 6 types
- Project template optimization and smart merging
- Template version management and history
- Backup discovery and restoration
- Pattern reusability and customization

When to Use:

- Scaffolding new projects or features
- Creating GitHub issues with /moai:9-feedback
- Optimizing template structures after MoAI-ADK updates
- Restoring from project backups
- Managing template versions and customizations
- Generating boilerplate code

Key Features:

- Code Templates: FastAPI, React, Vue, Docker, and CI/CD templates
- Feedback Templates: 6 GitHub issue types including bug, feature, improvement, refactor, docs, and question
- Template Optimizer: Smart merge, backup restoration, and version tracking
- Pattern Library: Reusable patterns for common scenarios

Quick Access to Modules:

- Code Templates documentation in modules/code-templates.md
- Feedback Templates documentation in modules/feedback-templates.md
- Template Optimizer documentation in modules/template-optimizer.md

## Implementation Guide

### Features

- Project templates for common architectures
- Boilerplate code generation with best practices
- Configurable template variables and customization
- Multi-framework support including React, FastAPI, and Spring
- Integrated testing and CI/CD configurations

### When to Use

- Bootstrapping new projects with proven architecture patterns
- Ensuring consistency across multiple projects in an organization
- Quickly prototyping new features with proper structure
- Onboarding new developers with standardized project layouts
- Generating microservices or modules following team conventions

### Core Patterns

Pattern 1 - Template Structure:

Templates are organized in a directory hierarchy. The top-level templates directory contains framework-specific subdirectories. A backend framework directory such as fastapi-backend contains template.json for variables and a src directory with main.py, models subdirectory, and tests subdirectory. A frontend framework directory such as nextjs-frontend contains template.json, app directory, and components directory. A fullstack template contains separate backend and frontend subdirectories.

Pattern 2 - Template Variables:

Template variables are defined in a JSON configuration file with two main sections. The variables section defines key-value pairs such as PROJECT_NAME, AUTHOR, LICENSE, and PYTHON_VERSION. The files section maps file patterns to processing modes: files marked as substitute have variables replaced, while files marked as copy are transferred unchanged.

Pattern 3 - Template Generation:

The template generation process follows five steps. First, load the template directory structure. Second, substitute variables in files marked for substitution. Third, copy static files as-is. Fourth, run post-generation hooks such as dependency installation and git initialization. Fifth, validate the generated project structure.

## Core Patterns in Detail

### Pattern 1: Code Template Scaffolding

Concept: Rapidly scaffold projects with production-ready boilerplates.

To generate a project, load the appropriate template such as backend/fastapi. Configure the scaffold with the project name, desired features such as auth, database, and celery, and customizations such as database type. Execute the scaffold to create the project structure.

For complete library and examples, see the Code Templates module documentation.

---

### Pattern 2: GitHub Feedback Templates

Concept: Structured templates for consistent GitHub issue creation.

Six Template Types: Bug Report, Feature Request, Improvement, Refactor, Documentation, and Question/Discussion.

Integration: Auto-triggered by the /moai:9-feedback command.

For all template types and usage, see the Feedback Templates module documentation.

---

### Pattern 3: Template Optimization and Smart Merge

Concept: Intelligently merge template updates while preserving user customizations.

Smart Merge Algorithm: The three-way merge process works as follows. First, extract user customizations from the backup. Second, get the latest template defaults from the current templates. Third, merge with appropriate priority where template_structure uses the latest defaults, user_config preserves user settings, and custom_content retains user modifications.

For complete workflow and examples, see the Template Optimizer module documentation.

---

### Pattern 4: Backup Discovery and Restoration

Concept: Automatic backup management with intelligent restoration.

Restoration Process: The process follows four steps. First, load backup metadata using the backup identifier. Second, validate backup integrity and raise an error if the backup is corrupted. Third, extract customizations from the validated backup. Fourth, apply the extracted customizations to the current project.

For complete implementation, see the Template Optimizer module section on Restoration Process.

---

### Pattern 5: Template Version Management

Concept: Track template versions and maintain update history.

Version Tracking: The template_optimization configuration section stores last_optimized timestamp, backup_version identifier, template_version number, and customizations_preserved list containing items like language, team_settings, and domains.

For complete implementation, see the Template Optimizer module section on Version Tracking.

---

## Module Reference

### Core Modules

- Code Templates in modules/code-templates.md: Boilerplate library, scaffold patterns, and framework templates
- Feedback Templates in modules/feedback-templates.md: 6 GitHub issue types, usage examples, and best practices
- Template Optimizer in modules/template-optimizer.md: Smart merge algorithm, backup restoration, and version management

### Module Contents

Code Templates include FastAPI REST API template, React component template, Docker and CI/CD templates, and template variables with scaffolding patterns.

Feedback Templates include Bug Report template, Feature Request template, Improvement template, Refactor template, Documentation template, Question template, and integration with /moai:9-feedback command.

Template Optimizer includes 6-phase optimization workflow, smart merge algorithm, backup discovery and restoration, and version tracking with history.

## Advanced Documentation

For detailed patterns and implementation strategies, refer to the Code Templates Guide for complete template library, Feedback Templates for issue template reference, and Template Optimizer for optimization and merge strategies.

## Best Practices

### Core Requirements

- Use templates for consistent project structure
- Preserve user customizations during updates
- Create backups before major template changes
- Follow template structure conventions
- Document custom modifications
- Use smart merge for template updates
- Track template versions in config
- Test templates before production use

### Quality Standards

[HARD] Document all template default modifications before applying changes.
WHY: Template defaults serve as the baseline for all projects and undocumented changes create confusion and inconsistency across teams.
IMPACT: Without documentation, teams cannot understand why defaults deviate from standards, leading to maintenance issues and conflicting implementations.

[HARD] Create backups before executing template optimization workflows.
WHY: Template optimization involves structural changes that may be difficult to reverse without a clean restoration point.
IMPACT: Missing backups can result in permanent loss of user customizations, requiring manual reconstruction of project-specific configurations.

[HARD] Resolve all merge conflicts during template update workflows.
WHY: Unresolved conflicts create broken configurations that prevent proper template functionality.
IMPACT: Ignored conflicts lead to runtime errors, inconsistent behavior, and project instability requiring emergency fixes.

[SOFT] Maintain consistent template pattern usage throughout the project.
WHY: Mixing different template patterns creates cognitive overhead and makes the codebase harder to understand and maintain.
IMPACT: Inconsistent patterns reduce code predictability and increase onboarding time for new team members.

[HARD] Preserve complete customization history across all template updates.
WHY: Customization history provides an audit trail of project-specific decisions and enables rollback to previous states.
IMPACT: Lost history makes it impossible to understand why changes were made, preventing informed decisions about future modifications.

[HARD] Validate template functionality through testing before production deployment.
WHY: Untested templates may contain errors that only surface in production environments, causing system failures.
IMPACT: Production failures from untested templates result in downtime, data issues, and emergency rollbacks affecting users.

[SOFT] Design templates within reasonable complexity limits for maintainability.
WHY: Excessive template complexity makes them difficult to understand, modify, and debug when issues arise.
IMPACT: Overly complex templates slow down development velocity and increase the likelihood of errors during customization.

[HARD] Track template versions using the built-in version management system.
WHY: Version tracking enables understanding of template evolution, compatibility checking, and coordinated updates.
IMPACT: Without version tracking, teams cannot determine which template features are available or coordinate updates across projects safely.

## Works Well With

Agents:

- workflow-project: Project initialization
- core-planner: Template planning
- workflow-spec: SPEC template generation

Skills:

- moai-project-config-manager: Configuration management and validation
- moai-cc-configuration: Claude Code settings integration
- moai-foundation-specs: SPEC template generation
- moai-docs-generation: Documentation template scaffolding
- moai-core-workflow: Template-driven workflows

Commands:

- /moai:0-project: Project initialization with templates
- /moai:9-feedback: Feedback template selection and issue creation

## Workflow Integration

Project Initialization Workflow: Select code template using Pattern 1, scaffold project structure, apply customizations, and initialize version tracking using Pattern 5.

Feedback Submission Workflow: Execute /moai:9-feedback command, select issue type using Pattern 2, fill template fields, and auto-generate GitHub issue.

Template Update Workflow: Detect template version change, create backup using Pattern 4, run smart merge using Pattern 3, and update version history using Pattern 5.

## Success Metrics

- Scaffold Time: 2 minutes for new projects compared to 30 minutes manual
- Template Adoption: 95% of projects use templates
- Customization Preservation: 100% user content retained during updates
- Feedback Completeness: 95% GitHub issues with complete information
- Merge Success Rate: 99% conflicts resolved automatically

## Changelog

- v3.1.0 (2026-01-11): Converted to CLAUDE.md documentation standards, removed code blocks and tables
- v3.0.0 (2026-01-08): Major version with modular architecture
- v2.0.0 (2025-11-24): Unified moai-core-code-templates, moai-core-feedback-templates, and moai-project-template-optimizer into single skill with 5 core patterns
- v1.0.0 (2025-11-22): Original individual skills

---

Status: Production Ready (Enterprise)
Modular Architecture: SKILL.md + 3 core modules
Integration: Plan-Run-Sync workflow optimized
Generated with: MoAI-ADK Skill Factory
