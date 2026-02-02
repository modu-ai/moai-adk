---
name: moai-workflow-project
description: >
  Integrated project management system with documentation, language
  initialization, and template optimization modules. Use when setting up
  projects, generating documentation, configuring multilingual support,
  or optimizing templates.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Bash(git:*) Bash(npm:*) Bash(npx:*) Bash(uv:*) Bash(pip:*) Bash(ls:*) Bash(mkdir:*) Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-01-08"
  modularized: "true"
  tags: "workflow, project, documentation, initialization, templates"
  aliases: "moai-workflow-project"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["project setup", "initialization", "project initialization", "project configuration", "project documentation", "multilingual", "language initialization"]
  phases: ["plan", "run"]
  agents: ["manager-project", "manager-docs"]
---

# MoAI Workflow Project - Integrated Project Management System

Purpose: Comprehensive project management system that integrates documentation generation, multilingual support, and template optimization into unified architecture with intelligent automation and Claude Code integration.

Scope: Consolidates documentation management, language initialization, and template optimization into single cohesive system supporting complete project lifecycle from initialization to maintenance.

Target: Claude Code agents for project setup, documentation generation, multilingual support, and performance optimization.

---

## Quick Reference

Core Capabilities:

- Documentation Management: Template-based documentation generation with multilingual support
- Language Initialization: Language detection, configuration, and localization management
- Template Optimization: Advanced template analysis and performance optimization
- Unified Interface: Single entry point integrating all capabilities

Key Features:

- Automatic project type detection and template selection
- Multilingual documentation generation supporting English, Korean, Japanese, and Chinese
- Intelligent template optimization with performance benchmarking
- SPEC-driven documentation updates
- Multi-format export including Markdown, HTML, and PDF

Supported Project Types:

- Web applications
- Mobile applications
- Command-line interface tools
- Libraries and packages
- Machine learning projects

---

## Implementation Guide

### Module Architecture

Documentation Management Capabilities:

- Template-based documentation generation
- Project type detection for web, mobile, CLI, library, and ML projects
- Multilingual support with localized content
- SPEC data integration for automatic updates
- Multi-format export capabilities

Language Initialization Capabilities:

- Automatic language detection from project content
- Comprehensive language configuration management
- Agent prompt localization with cost optimization
- Domain-specific language support
- Locale management and cultural adaptation

Template Optimization Capabilities:

- Advanced template analysis with complexity metrics
- Performance optimization with size reduction
- Intelligent backup and recovery system
- Benchmarking and performance tracking
- Automated optimization recommendations

### Core Workflows

Complete Project Initialization Workflow:

Step 1: Initialize the project management system by specifying the project directory path

Step 2: Execute complete setup with the following configuration parameters:

- Language setting: Specify the primary language code such as "en" for English or "ko" for Korean
- User name: Provide the developer or team name for personalization
- Domains: List the project domains such as backend, frontend, and mobile
- Project type: Specify the project type such as web_application
- Optimization enabled: Set to true to enable template optimization during initialization

Step 3: Review initialization results which include:

- Language configuration with token cost analysis
- Documentation structure creation status
- Template analysis and optimization report
- Multilingual documentation setup confirmation

Documentation Generation from SPEC Workflow:

Step 1: Prepare SPEC data with the following structure:

- Identifier: Unique SPEC ID such as SPEC-001
- Title: Feature or component title
- Description: Brief description of the implementation
- Requirements: List of specific requirements
- Status: Current status such as Planned, In Progress, or Complete
- Priority: Priority level such as High, Medium, or Low
- API Endpoints: List of endpoint definitions including path, method, and description

Step 2: Generate comprehensive documentation from the SPEC data

Step 3: Review generated documentation which includes:

- Feature documentation with requirements
- API documentation with endpoint details
- Updated project documentation files
- Multilingual versions if configured

Template Performance Optimization Workflow:

Step 1: Analyze current templates to gather metrics

Step 2: Configure optimization options:

- Backup first: Set to true to create backup before optimization
- Apply size optimizations: Enable to reduce file sizes
- Apply performance optimizations: Enable to improve loading times
- Apply complexity optimizations: Enable to simplify template structures
- Preserve functionality: Ensure all features remain intact

Step 3: Execute optimization and review results:

- Size reduction percentage achieved
- Performance improvement metrics
- Backup creation confirmation
- Detailed optimization report

### Language and Localization

Automatic Language Detection Process:

The system analyzes the project for language indicators using the following methods:

- File content analysis examining comments and strings
- Configuration file examination for locale settings
- System locale detection
- Directory structure patterns

Multilingual Documentation Structure:

When creating documentation for multiple languages, the system generates:

- Language-specific documentation directories such as docs/ko for Korean and docs/en for English
- Language negotiation configuration
- Automatic redirection setup between language versions

Agent Prompt Localization:

The localization system provides:

- Language-specific instructions for agents
- Cultural context adaptations
- Token cost optimization recommendations for multilingual prompts

### Template Optimization

Performance Analysis Process:

Template analysis provides the following metrics:

- File size and complexity measurements
- Performance bottleneck identification
- Optimization opportunity scoring
- Resource usage patterns
- Backup recommendations

Intelligent Optimization Process:

The optimization system applies the following techniques:

- Whitespace and redundancy reduction
- Template structure optimization
- Complexity reduction techniques
- Performance caching improvements

### Configuration Management

Integrated Configuration Status:

The project status includes:

- Project metadata and type classification
- Language configuration and associated costs
- Documentation completion status
- Template optimization results
- Module initialization states

Language Settings Update Process:

When updating language settings, configure the following parameters:

- Conversation language: The language for user-facing responses
- Agent prompt language: The language for internal agent instructions, often kept as English for cost optimization
- Documentation language: The language for generated documentation

Updates trigger the following automatic changes:

- Configuration file modifications
- Documentation structure updates
- Template localization adjustments

---

## Advanced Implementation

For advanced patterns including custom template development, performance optimization strategies, and integration workflows, see reference.md which covers:

- Custom Templates: Creating project-type-specific documentation templates
- Performance Caching: Implementing template caching for improved performance
- Batch Processing: Efficiently processing multiple templates simultaneously
- Integration Workflows: Complete project lifecycle and multilingual management

---

## Resources

### Configuration Structure

Project Configuration Fields:

- Project name: The display name for the project
- Project type: Classification such as web_application, mobile_application, or cli_tool
- Initialization timestamp: When the project was initialized
- Language configuration: Conversation, agent prompt, and documentation language settings
- System version: The project management system version
- Initialization status: Whether all modules are fully initialized

Language Configuration Fields for Each Supported Language:

- Name: Display name in English
- Native name: Display name in the native language
- Code: Language code such as en, ko, ja, or zh
- Locale: System locale string
- Agent prompt language: Whether to use English or localized prompts
- Token cost impact: Percentage increase in token usage for non-English prompts

### Performance Metrics

Documentation Generation Performance:

- Complete documentation generation typically completes within 2 to 5 seconds
- Language detection analysis completes within approximately 500 milliseconds
- Template optimization duration varies from 10 to 30 seconds depending on project size
- Configuration updates complete within approximately 100 milliseconds

Memory Usage Characteristics:

- Base system requires approximately 50MB RAM
- Large projects may require an additional 10 to 50MB depending on template count
- Optimization cache uses approximately 5 to 20MB for performance improvements

File Size Impact:

- Documentation files range from 50 to 200KB per project
- Optimization backups match the size of original templates
- Configuration files range from 5 to 10KB for complete project setup

---

## Works Well With

- moai-foundation-core: Core execution patterns and SPEC-driven development workflows
- moai-foundation-claude: Claude Code integration and configuration
- moai-workflow-docs: Unified documentation management
- moai-workflow-templates: Template optimization strategies
- moai-library-nextra: Advanced documentation architecture
