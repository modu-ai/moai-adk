---
name: moai-workflow-project
description: >
  Integrated project management system covering documentation, language initialization,
  template optimization, docs generation, and JIT document loading. Absorbed from
  moai-workflow-templates, moai-docs-generation, and moai-workflow-jit-docs.

when_to_use: >
  Use for integrated project management: documentation scaffolding
  (product/structure/tech.md), multilingual language initialization,
  template and boilerplate optimization, docs generation (Sphinx, MkDocs,
  TypeDoc, OpenAPI), and JIT document loading.

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Bash(git:*), Bash(npm:*), Bash(npx:*), Bash(uv:*), Bash(pip:*), Bash(ls:*), Bash(mkdir:*), Grep, Glob, WebFetch, WebSearch
user-invocable: false
metadata:
  version: "3.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-04-25"
  modularized: "true"
  tags: "workflow, project, documentation, initialization, templates, boilerplate, scaffolding, jit-docs, docs-generation"
  aliases: "moai-workflow-project"
  related-skills: "moai-workflow-templates, moai-docs-generation, moai-workflow-jit-docs"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
---

# MoAI Workflow Project - Integrated Project Management System

Comprehensive project management system integrating documentation generation, multilingual support, and template optimization into a unified architecture with intelligent automation and Claude Code integration.

Scope: Consolidates documentation management, language initialization, and template optimization into a single cohesive system supporting complete project lifecycle from initialization to maintenance.

Target: Claude Code agents for project setup, documentation generation, multilingual support, and performance optimization.

---

## Quick Reference

Core Capabilities:

- Documentation Management: Template-based generation with multilingual support
- Language Initialization: Detection, configuration, localization
- Template Optimization: Analysis with performance optimization
- Unified Interface: Single entry point for all capabilities

Key Features:

- Automatic project type detection and template selection
- Multilingual documentation (English, Korean, Japanese, Chinese)
- Intelligent template optimization with benchmarking
- SPEC-driven documentation updates
- Multi-format export (Markdown, HTML, PDF)

Supported Project Types: web applications, mobile applications, CLI tools, libraries, ML projects.

---

## Implementation Guide

### Module Architecture

Three capability areas:

- Documentation Management: template-based generation, project type detection, multilingual support, SPEC integration, multi-format export
- Language Initialization: automatic detection, configuration management, agent prompt localization, locale management
- Template Optimization: complexity analysis, performance optimization, backup/recovery, benchmarking

### Core Workflows

Three workflows: project initialization, documentation generation from SPEC, template performance optimization. Each follows a 3-step pattern (configure → execute → review results).

See [core workflow walkthroughs](references/workflows.md) for detailed step-by-step procedures.

### Language and Localization

Automatic Language Detection: analyzes file content, configuration files, system locale, and directory structure.

Multilingual Documentation: language-specific directories (e.g., `docs/ko`, `docs/en`), language negotiation, automatic redirection.

Agent Prompt Localization: language-specific instructions, cultural context, token cost optimization.

See [language and localization detail](references/language-localization.md) for token cost analysis and locale configuration.

### Template Optimization

Performance Analysis: file size, complexity, performance bottlenecks, optimization opportunities, resource usage, backup recommendations.

Optimization Techniques: whitespace reduction, structure optimization, complexity reduction, performance caching.

### Configuration Management

Integrated configuration covers project metadata, language settings + costs, documentation status, template optimization results, module initialization states.

Language Settings: conversation_language (user-facing), agent_prompt_language (internal, often English for cost), documentation language (per language).

Updates trigger configuration file modifications, documentation structure updates, template localization.

See [configuration schema and language fields](references/configuration.md) for full field reference and supported language metadata.

---

## Advanced Implementation

For advanced patterns (custom templates, performance caching, batch processing, integration workflows), see [references/reference.md](references/reference.md) and [references/examples.md](references/examples.md).

---

## Resources

### Performance Metrics

| Operation | Typical Duration |
|-----------|------------------|
| Complete documentation generation | 2-5 seconds |
| Language detection analysis | ~500 ms |
| Template optimization | 10-30 seconds |
| Configuration updates | ~100 ms |

Memory: base ~50MB, large projects +10-50MB, optimization cache 5-20MB.

File sizes: documentation 50-200KB per project, optimization backups match originals, configuration 5-10KB.

---

## Works Well With

- moai-foundation-core: Core execution patterns and SPEC-driven workflows
- moai-foundation-cc: Claude Code integration and configuration
- moai-workflow-docs: Unified documentation management
- moai-workflow-templates: Template optimization strategies
- moai-library-nextra: Documentation architecture

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "project docs are a one-time setup, no need to keep them current" | Stale product.md and tech.md mislead every SPEC written against them. They are living documents. |
| "structure.md will stay accurate since the codebase does not change much" | Every PR that adds a directory invalidates structure.md. Sync on every /moai project invocation. |
| "I know the tech stack, I do not need tech.md" | tech.md is not for you. It is for every agent that consults the project context before acting. |
| "codemaps take too long to generate" | codemaps are the only artifact that gives agents file-level awareness without reading every file. The cost pays for itself. |
| "I will write project docs after the feature is done" | Post-hoc docs capture what was built, not what was intended. Pre-feature docs guide the build. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- product.md references features that do not exist in the codebase
- tech.md lists a framework version that differs from the actual dependency file
- structure.md missing directories that exist on disk
- codemaps/ directory is empty or missing while the project has 10+ source files
- /moai project last ran more than 30 days ago (check file modification dates)

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] product.md exists and describes the current product scope (compare with README)
- [ ] tech.md lists dependencies that match the actual lock file (package.json, go.mod, etc.)
- [ ] structure.md top-level directories match `ls` output of the project root
- [ ] codemaps/ contains at least one codemap file per major package or module
- [ ] All three docs updated within the current session or since last structural change
- [ ] No placeholder text ("TODO", "TBD") remains in generated documents

<!-- moai:evolvable-end -->

---

## Template Optimization (absorbed from moai-workflow-templates)

Code boilerplates, feedback templates, scaffolding, and project template optimization.

### Core Capabilities

- Code template library: FastAPI, React, Vue, Next.js boilerplates
- GitHub issue feedback templates: 6 types (bug, feature, question, docs, perf, security)
- Project template optimization: size reduction, complexity analysis, smart merging
- Template version management, backup discovery and restoration

### Template Application Workflow

1. Identify template category: code boilerplate, feedback template, or project scaffold
2. Select template variant matching the project stack and language
3. Apply customization variables (project name, author, license, framework version)
4. Validate rendered output against schema or existing conventions
5. Optionally run template optimizer to reduce redundancy

### Template Optimization Process

Analysis metrics: file size, complexity score, redundancy ratio, load performance.
Optimization techniques: whitespace reduction, deduplication, structure simplification.
Always create backup before applying optimization (`backup: true`).

---

## Documentation Generation (absorbed from moai-docs-generation)

Technical docs generation using Sphinx, MkDocs, TypeDoc, OpenAPI, and Nextra. Covers project configuration discovery, how to integrate with existing docs sites, and best practices for keeping framework documentation and technology guide content current.

### Supported Generators

| Generator | Use Case | Primary Format |
|-----------|----------|----------------|
| Sphinx | Python projects, API docs | RST / Markdown |
| MkDocs | General projects | Markdown |
| TypeDoc | TypeScript libraries | TypeScript JSDoc |
| OpenAPI / Swagger | REST APIs | YAML / JSON |
| Nextra | Next.js docs sites | MDX |

### Generation Workflow

1. Detect project type and select appropriate generator
2. Extract documentation sources: docstrings, JSDoc, OpenAPI specs, SPEC documents
3. Apply project language and branding from `.moai/config/sections/language.yaml`
4. Generate output in configured format (Markdown, HTML, PDF)
5. Update `/moai sync` artifacts: README, CHANGELOG, API reference

---

## JIT Document Loading (absorbed from moai-workflow-jit-docs)

JIT docs (just in time docs) — on-demand documentation discovery and loading based on user intent and conversation context.

### Primary Tools

- WebFetch / WebSearch: Latest online documentation and official library docs
- Read, Grep, Glob: Local project documentation

### Trigger Patterns

- User asks specific technical questions about a library or framework
- Technology keyword detected (library name, framework name, API name)
- Domain expertise required (authentication, database, deployment)
- Implementation guidance needed during run phase

### Loading Priority

1. Local project docs (`.moai/`, README, SPEC documents)
2. WebSearch + WebFetch (official documentation and latest online resources)

Token budget: 5000 tokens per JIT load. Summarize if source exceeds budget.
