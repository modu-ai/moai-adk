# Agent-Skills Import Reference

Extended documentation for agent-skills format conversion and validation.

## Agent-Skills Format Specification

### YAML Frontmatter Fields

Required Fields:

- name: Skill identifier in kebab-case format, may include organization prefix
- description: Brief explanation of skill purpose and trigger scenarios, maximum 1024 characters

Optional Fields:

- argument-hint: Hint for command-line argument format like <file-or-pattern>

Example Frontmatter:

```yaml
---
name: vercel-react-best-practices
description: React and Next.js performance optimization guidelines from Vercel Engineering. This skill should be used when writing, reviewing, or refactoring React/Next.js code to ensure optimal performance patterns.
argument-hint: <file-or-pattern>
---
```

### metadata.json Fields

Required Fields:

- version: Semantic version string like 0.1.0 or 1.0.0
- organization: Maintaining entity name like Vercel Engineering
- date: Publication date in Month YYYY format like January 2026
- abstract: Comprehensive summary of skill purpose and capabilities
- references: Array of URL strings to official documentation

Example metadata.json:

```json
{
  "version": "0.1.0",
  "organization": "Vercel Engineering",
  "date": "January 2026",
  "abstract": "Comprehensive performance optimization guide for React and Next.js applications, designed for AI agents and LLMs.",
  "references": [
    "https://react.dev",
    "https://nextjs.org",
    "https://swr.vercel.app"
  ]
}
```

### Rule File Format

Frontmatter Fields:

- title: Rule title in human-readable format
- impact: Impact level like CRITICAL, HIGH, MEDIUM, or LOW
- impactDescription: Performance improvement description like 2-10x improvement
- tags: Array of relevant keywords for categorization

Body Structure:

- Brief explanation of why the rule matters
- Incorrect code example with explanation
- Correct code example with explanation
- Additional context and references

Example Rule File:

```markdown
---
title: Promise.all() for Independent Operations
impact: CRITICAL
impactDescription: 2-10x improvement
tags: async, parallelization, promises, waterfalls
---

## Promise.all() for Independent Operations

When async operations have no interdependencies, execute them concurrently using Promise.all().

**Incorrect (sequential execution, 3 round trips):**

```typescript
const user = await fetchUser()
const posts = await fetchPosts()
const comments = await fetchComments()
```

**Correct (parallel execution, 1 round trip):**

```typescript
const [user, posts, comments] = await Promise.all([
  fetchUser(),
  fetchPosts(),
  fetchComments()
])
```
```

## MoAI-ADK Format Specification

### Comprehensive Frontmatter

Required Fields:

- name: Skill identifier with moai- prefix, kebab-case format, maximum 64 characters
- description: Function and trigger scenarios, maximum 1024 characters, format "[Function verb] [target domain]. Use when [trigger 1], [trigger 2], or [trigger 3]."
- allowed-tools: Comma-separated tool list without brackets
- version: Semantic version string
- status: Active status like active or deprecated
- updated: Last update date in YYYY-MM-DD format

Optional Fields:

- category: Skill category like workflow, domain, language, platform, library, or foundation
- author: Maintainer name like MoAI-ADK Team
- tags: Array of relevant keywords for discoverability
- modularized: Boolean indicating if skill is modular
- user-invocable: Boolean indicating if users can invoke directly

Example Frontmatter:

```yaml
---
name: moai-domain-react
description: React and Next.js performance specialist covering waterfalls elimination, bundle optimization, server-side rendering, and client-side data fetching patterns. Use when optimizing React components, Next.js pages, data fetching, bundle size, or performance improvements.
allowed-tools: Read, Write, Edit, Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
version: 1.0.0
status: active
updated: 2026-01-15
category: domain
author: MoAI-ADK Team
tags:
  - react
  - nextjs
  - performance
  - optimization
---
```

### Progressive Disclosure Structure

Level 1: Quick Reference (80-120 lines)

Purpose: Provide immediate value within 30 seconds

Content Requirements:

- Core capabilities list with 4 to 6 bullet points
- When to use scenarios with 4 to 6 specific triggers
- Key features or quick commands if applicable
- Brief overview of main functionality

Level 2: Implementation Guide (180-250 lines)

Purpose: Provide step-by-step guidance for common tasks

Content Requirements:

- Logical subsections with clear headings
- Working code examples demonstrating each major capability
- Narrative text format for explaining workflows and logic
- Detailed procedures for common use cases

Level 3: Advanced Patterns (80-140 lines)

Purpose: Provide expert-level knowledge for complex scenarios

Content Requirements:

- Optimization strategies and best practices
- Edge case handling and error recovery
- Integration patterns with related tools and services
- References to external documentation or extended guides

## Conversion Workflow

### Phase 1: Analysis

Step 1: Read agent-skills SKILL.md file to extract frontmatter and content structure

Step 2: Read metadata.json to extract version, organization, abstract, and references

Step 3: List files in rules/ or references/ directories to understand supporting documentation

Step 4: Analyze skill purpose and scope to determine appropriate MoAI-ADK category

Step 5: Identify required tools based on skill functionality and intended operations

### Phase 2: Frontmatter Conversion

Step 1: Convert name field by adding moai- prefix if not already present

Step 2: Convert description field to include both function and trigger scenarios

Step 3: Determine allowed-tools based on skill requirements and least privilege principle

Step 4: Add version from metadata.json or default to 1.0.0

Step 5: Add status field set to active

Step 6: Add updated field with current date in YYYY-MM-DD format

Step 7: Add category field based on skill domain and purpose

Step 8: Add author field set to MoAI-ADK Team for converted skills

Step 9: Add tags array with relevant keywords for discoverability

### Phase 3: Content Organization

Step 1: Create Quick Reference section with core capabilities and when to use scenarios

Step 2: Extract key features from original content for immediate value

Step 3: Create Implementation Guide section with step-by-step workflows

Step 4: Include working code examples demonstrating major capabilities

Step 5: Create Advanced Patterns section for expert-level knowledge

Step 6: Reference external documentation or extended guides when needed

### Phase 4: File Structure

Step 1: Create skill directory at src/moai_adk/templates/.claude/skills/skill-name/

Step 2: Write SKILL.md with converted frontmatter and progressive disclosure content

Step 3: Create reference.md for overflow content if SKILL.md exceeds 500 lines

Step 4: Create examples.md with working code examples demonstrating skill capabilities

Step 5: Create scripts/ directory if automation scripts are needed

Step 6: Create templates/ directory if reusable templates are available

### Phase 5: Validation

Step 1: Verify that name follows kebab-case format with maximum 64 characters

Step 2: Confirm that description includes both function and trigger scenarios within 1024 characters

Step 3: Check that allowed-tools uses comma-separated format without brackets

Step 4: Ensure that version follows semantic versioning format

Step 5: Verify that status field is set to active or appropriate value

Step 6: Confirm that updated field uses YYYY-MM-DD date format

Step 7: Verify that SKILL.md is under 500 lines including frontmatter

Step 8: Confirm that progressive disclosure is implemented with all three levels

Step 9: Check that skill works well with related skills and agents

Step 10: Verify that tool permissions follow least privilege principle

## Tool Permission Guide

### Read-Only Skills

Purpose: Analyze code, extract patterns, provide insights

Allowed Tools:

- Read: Read file contents with absolute path support
- Grep: Search file contents with regex patterns
- Glob: Find files by name patterns with recursive search

Use Cases:

- Code analysis and review
- Pattern detection and validation
- Documentation generation
- Compliance checking

### Documentation Research Skills

Purpose: Access official documentation, retrieve latest information

Allowed Tools:

- Read: Read local documentation files
- Grep: Search documentation content
- Glob: Find documentation files
- WebFetch: Fetch remote documentation and guidelines
- WebSearch: Search web for latest information
- mcp__context7__resolve-library-id: Resolve library names to Context7 IDs
- mcp__context7__get-library-docs: Fetch official documentation via Context7

Use Cases:

- Official documentation access
- Latest version guidance
- API reference retrieval
- Best practice research

### File Modification Skills

Purpose: Generate code, refactor files, create documentation

Allowed Tools:

- Read: Read existing file contents
- Write: Create new files or overwrite existing files
- Edit: Make specific string replacements in files
- Grep: Search file contents
- Glob: Find files by patterns

Use Cases:

- Code generation and refactoring
- Documentation creation
- Configuration file updates
- Template instantiation

### System Operations Skills

Purpose: Execute commands, run tests, build projects

Allowed Tools:

- Bash: Execute shell commands with timeout and error handling
- Read: Read file contents
- Grep: Search file contents
- Glob: Find files

Use Cases:

- Build and test execution
- Dependency installation
- Process management
- System integration

## Error Handling

### Missing metadata.json

Error: metadata.json file not found in agent-skills directory

Resolution: Use default values for version (1.0.0), organization (Unknown), and references (empty array). Log warning about missing metadata file.

### Invalid Frontmatter

Error: YAML frontmatter parsing fails or missing required fields

Resolution: Use sensible defaults based on file name and content structure. Log specific parsing errors for manual review.

### No Rules Directory

Error: rules/ or references/ directory not found

Resolution: Extract content directly from SKILL.md body. Check for AGENTS.md as alternative source of compiled content.

### Exceeds 500-Line Limit

Error: SKILL.md content exceeds 500 lines after conversion

Resolution: Extract overflow content to reference.md. Add cross-references between files. Prioritize core content in SKILL.md.

### Tool Permission Uncertainty

Error: Unable to determine appropriate tool permissions for converted skill

Resolution: Apply least privilege principle. Start with Read, Grep, Glob for read-only skills. Add Write, Edit, Bash only when absolutely necessary. Review and adjust based on skill requirements.
