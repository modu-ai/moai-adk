---
name: moai-import-agent-skills
description: Import and convert agent-skills format to MoAI-ADK skill format. Use when importing skills from agent-skills repository, converting third-party skills to MoAI-ADK format, or parsing agent-skills structure with YAML frontmatter and metadata.json files.
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
version: 1.0.0
status: active
updated: 2026-01-15
author: MoAI-ADK Team
category: workflow
tags:
  - import
  - conversion
  - agent-skills
  - skills
  - migration
---

# Agent-Skills Import and Conversion

## Quick Reference

Agent-skills format import and conversion specialist for MoAI-ADK compatibility.

Purpose: Convert agent-skills repository format to MoAI-ADK skill format with progressive disclosure architecture and official standards compliance.

Agent-skills Format Structure:

- SKILL.md: Main skill file with YAML frontmatter (name, description, argument-hint)
- metadata.json: Additional metadata (version, organization, abstract, references)
- rules/ or references/: Supporting documentation files
- AGENTS.md: Full compiled document (optional)

MoAI-ADK Format Requirements:

- SKILL.md: Main skill file with comprehensive YAML frontmatter and progressive disclosure
- reference.md: Extended documentation (optional, for overflow from 500-line limit)
- examples.md: Working code examples (optional, for demonstration)
- scripts/: Utility scripts (optional, for automation)
- templates/: Reusable templates (optional)

When to Use:

- Importing skills from agent-skills repository
- Converting third-party skills to MoAI-ADK format
- Migrating existing skill repositories to MoAI-ADK structure
- Validating agent-skills format compliance

---

## Implementation Guide

### Reading Agent-Skills Files

Parse YAML Frontmatter from SKILL.md:

Read the SKILL.md file from agent-skills directory. Extract YAML frontmatter between the first two triple-dash delimiters. The frontmatter typically contains name, description, and optionally argument-hint fields. The name field uses kebab-case format and may include organization prefix like vercel-react-best-practices. The description field explains when the skill should be invoked and what trigger terms activate it.

Read metadata.json for Additional Context:

Parse the metadata.json file to extract version, organization, date, abstract, and references array. The version follows semantic versioning like 0.1.0. The organization field identifies the maintaining entity like Vercel Engineering. The abstract provides a comprehensive summary of the skill purpose and capabilities. The references array contains URLs to official documentation and related resources.

Read Supporting Documentation:

Check for rules/ directory containing individual rule files with markdown format. Each rule file typically contains frontmatter with title, impact, impactDescription, and tags. The body includes brief explanations, incorrect and correct code examples, and additional context. Check for references/ directory as alternative structure. Check for AGENTS.md which contains the full compiled document with all rules expanded.

### Converting to MoAI-ADK Format

Map Frontmatter Fields:

Convert agent-skills name to MoAI-ADK format by adding moai- prefix if not already present. Convert agent-skills description to MoAI-ADK format which must include both the function and trigger scenarios. The description format should be "[Function verb] [target domain]. Use when [trigger 1], [trigger 2], or [trigger 3]." Maximum 1024 characters allowed. Add allowed-tools field based on skill requirements using comma-separated format without brackets. Add version field from metadata.json or default to 1.0.0. Add status field set to active. Add updated field with current date in YYYY-MM-DD format. Add category field like workflow, domain, language, platform, library, or foundation. Add optional author field set to MoAI-ADK Team for converted skills. Add tags array with relevant keywords for discoverability.

Determine Tool Permissions:

Apply least privilege access principle by granting only tools required for skill function. For read-only information gathering skills, include Read, Grep, and Glob. For documentation research skills, add WebFetch and WebSearch if needed. For system operations requiring file modification, add Write, Edit, and Bash when absolutely necessary. For external documentation access, add mcp__context7__resolve-library-id and mcp__context7__get-library-docs. Never include tools that enable unintended modifications or unauthorized operations.

Implement Progressive Disclosure Structure:

Level 1 Quick Reference section provides immediate value in approximately 30 seconds. Include core capabilities list with 4 to 6 bullet points. Include when to use scenarios with 4 to 6 specific triggers. Include key features or quick commands if applicable. Keep this section between 80 to 120 lines.

Level 2 Implementation Guide section provides step-by-step guidance for common tasks. Organize into logical subsections with clear headings. Include working code examples demonstrating each major capability. Use narrative text format for explaining workflows and logic. Keep this section between 180 to 250 lines.

Level 3 Advanced Patterns section provides expert-level knowledge for complex scenarios. Include optimization strategies, edge case handling, and integration patterns. Reference external documentation or extended guides in reference.md when needed. Keep this section between 80 to 140 lines.

Enforce 500-Line Limit:

Count total lines in SKILL.md including frontmatter and all sections. If the file exceeds 500 lines, extract overflow content to reference.md. Move advanced patterns or detailed explanations to reference.md. Add cross-references between SKILL.md and reference.md using relative paths. Keep only core content in SKILL.md for optimal loading performance.

### Handling Rules and References

Process Rules Directory:

List all markdown files in the rules/ directory. Parse frontmatter from each rule file to extract title, impact level, impact description, and tags. Group rules by impact priority if available such as critical, high, medium, or low. Extract code examples from each rule showing incorrect and correct implementations. Create summary tables or categorized lists for Quick Reference section. Include detailed rule explanations in Implementation Guide or Advanced Patterns sections.

Process References Directory:

List all files in the references/ directory. Read documentation files to extract key concepts and patterns. Organize content by topic or category for progressive disclosure. Create brief summaries for Quick Reference section. Include detailed explanations in Implementation Guide section with cross-references to original reference files.

Generate MoAI-ADK Output:

Create new skill directory at src/moai_adk/templates/.claude/skills/skill-name/. Write SKILL.md with converted frontmatter and progressive disclosure content. Create reference.md for overflow content if SKILL.md exceeds 500 lines. Create examples.md with working code examples demonstrating skill capabilities. Create scripts/ directory if automation scripts are needed. Create templates/ directory if reusable templates are available.

### Example Conversion

React Best Practices Conversion:

Original agent-skills format has name vercel-react-best-practices with description explaining React and Next.js performance optimization guidelines. The metadata.json contains version 0.1.0 from Vercel Engineering with comprehensive abstract and references array. The rules/ directory contains 45 rule files organized by priority categories.

Convert to MoAI-ADK format with name moai-domain-react (or similar kebab-case with moai- prefix). The description becomes "React and Next.js performance specialist covering waterfalls elimination, bundle optimization, server-side rendering, and client-side data fetching patterns. Use when optimizing React components, Next.js pages, data fetching, bundle size, or performance improvements." Add allowed-tools including Read, Write, Edit, Grep, Glob, and Context7 tools for documentation access.

Create Quick Reference section with 8 rule categories by priority including eliminating waterfalls as critical, bundle size optimization as critical, server-side performance as high, client-side data fetching as medium-high, re-render optimization as medium, rendering performance as medium, JavaScript performance as low-medium, and advanced patterns as low.

Create Implementation Guide section with detailed workflow for each rule category. Include code examples showing incorrect patterns with sequential promises, barrel imports, or missing Suspense boundaries. Include correct patterns with parallel promises, dynamic imports, and proper caching.

Create Advanced Patterns section covering performance monitoring, profiling strategies, and measurement techniques. Reference official React and Next.js documentation via Context7 integration.

Web Design Guidelines Conversion:

Original agent-skills format has name web-design-guidelines with description explaining Web Interface Guidelines compliance review. The skill fetches guidelines from external URL and reviews files for compliance.

Convert to MoAI-ADK format with name moai-domain-web-design or moai-workflow-design-review. The description becomes "Web interface guidelines compliance checker covering accessibility, responsive design, performance, and user experience patterns. Use when reviewing UI code, checking accessibility compliance, auditing design patterns, or validating UX best practices." Add allowed-tools including Read, Grep, Glob, and WebFetch for remote guidelines retrieval.

Create Quick Reference section explaining the compliance review workflow. List supported guideline categories including accessibility standards, responsive design patterns, performance metrics, and user experience heuristics.

Create Implementation Guide section with step-by-step process for fetching guidelines, reading files, applying rules, and generating reports. Include output format specification with file:line terse format for findings.

Create Advanced Patterns section covering integration with CI/CD pipelines, automated compliance reporting, and custom rule creation.

---

## Advanced Patterns

### Tool Permission Determination

Analyze Skill Requirements:

Examine the skill purpose and intended operations to determine required tools. For skills that only read and analyze code without modification, grant Read, Grep, and Glob for file access. For skills that generate documentation or reports, add Write tool for creating output files. For skills that execute build or test commands, add Bash tool for process execution. For skills requiring official documentation access, add Context7 tools for library resolution and documentation retrieval.

Apply Security Principles:

Follow least privilege access principle by granting minimum required tools. Exclude tools that enable dangerous operations like unrestricted Bash or Write. Use specific tool patterns when available such as Bash(git:*) for limiting to git commands only. Validate that granted tools align with skill scope and audience. Review tool permissions against security requirements for production environments.

### Validation and Quality Assurance

Validate Frontmatter Compliance:

Verify that name follows kebab-case format with maximum 64 characters. Confirm that description includes both function and trigger scenarios within 1024 characters. Check that allowed-tools uses comma-separated format without brackets. Ensure that version follows semantic versioning format. Verify that status field is set to active or appropriate value. Confirm that updated field uses YYYY-MM-DD date format.

Validate Content Structure:

Verify that progressive disclosure is implemented with all three levels. Check that Quick Reference section provides immediate value with core capabilities. Confirm that Implementation Guide section includes step-by-step workflows. Verify that Advanced Patterns section includes expert-level content. Check that all sections use narrative text format instead of code examples for explaining logic.

Validate File Structure:

Verify that SKILL.md is under 500 lines including frontmatter. Confirm that reference.md exists for overflow content if needed. Check that examples.md includes working code examples. Verify that directory structure matches MoAI-ADK standards. Confirm that all cross-references use relative paths correctly.

Validate Standards Compliance:

Verify that skill follows MoAI-ADK naming conventions with moai- prefix. Confirm that skill follows progressive disclosure architecture. Check that skill works well with related skills and agents. Verify that tool permissions follow least privilege principle. Confirm that skill content is in English as required by MoAI-ADK standards.

### Integration Patterns

Works Well With Section:

Identify complementary skills that enhance functionality. List related domain skills for broader coverage. Identify workflow skills for process integration. List platform or library skills for specific technology support. Identify foundation skills for core capabilities integration.

Example Integration:

React skills work well with moai-domain-frontend for UI patterns, moai-platform-vercel for deployment, moai-workflow-testing for test coverage, and moai-foundation-claude for Claude Code integration. Web design skills work well with moai-domain-uiux for design systems, moai-lang-typescript for type safety, and moai-workflow-docs for documentation generation.

---

## Works Well With

- moai-foundation-claude: Claude Code authoring and integration patterns
- moai-foundation-core: SPEC system and execution workflows
- moai-workflow-project: Project management and initialization
- moai-workflow-docs: Documentation generation and management
- builder-skill: Skill creation and validation workflows

---

## Resources

### Agent-Skills Repository

Source Format: https://github.com/vercel/agent-skills

Example Skills:

- vercel-react-best-practices: React and Next.js performance guidelines
- web-design-guidelines: Web interface guidelines compliance checker
- vercel-deploy-claimable: Vercel deployment automation

### MoAI-ADK Documentation

Skill Standards: .claude/skills/moai-foundation-claude/SKILL.md

Format Requirements: Progressive disclosure architecture, 500-line limit, English content

Integration Patterns: Works well with related skills and agents

### Tool Reference

Read: Read file contents with absolute path support

Grep: Search file contents with regex patterns and output modes

Glob: Find files by name patterns with recursive search

WebFetch: Fetch URLs for remote documentation and guidelines

WebSearch: Search web for latest information and resources

Context7: Resolve library IDs and fetch official documentation
