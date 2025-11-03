# Skill Metadata Authoring Guide

This guide details YAML frontmatter requirements for Claude Code Skills, with practical examples and validation rules.

---

## YAML Frontmatter Structure

Every SKILL.md requires a valid YAML frontmatter block:

```yaml
---
name: "Skill Name"
description: "What it does and when to use it"
allowed-tools: "Tool1, Tool2, Tool3"
---
```

### Frontmatter Rules

| Rule | Example | Reason |
|------|---------|--------|
| Start with `---` | `---` (3 dashes) | YAML document marker |
| No tabs, only spaces | Use 2 spaces for indent | YAML parsing requirements |
| Proper key-value pairs | `name: "value"` | Valid YAML syntax |
| String values in quotes | `"Skill Name"` | Required for special characters |
| End with `---` | `---` (3 dashes) | Marks end of metadata block |

---

## Field Specifications

### `name` (Required)

The display identifier for the Skill in Claude Code and documentation.

**Constraints**:
- Maximum 64 characters (strictly enforced)
- Gerund form (action verb): "Processing", "Debugging", "Analyzing"
- Domain-specific and discoverable
- Should NOT include "Skill", "Helper", or redundant prefixes

**Format**:
```
[Verb (Gerund)] [Domain/Context]
```

**Examples** ✅:
| Name | Character Count | Why It Works |
|------|-----------------|--------------|
| `Processing CSV Files with Python` | 36 | Action + specific domain |
| `Debugging Go Applications with Delve` | 42 | Action + language + tool |
| `Optimizing React Performance` | 31 | Action + framework |
| `Scanning Security Vulnerabilities` | 35 | Action + domain |

**Anti-examples** ❌:
| Name | Issue |
|------|-------|
| `CSV Helper` | Too generic, no verb |
| `Python Stuff` | Vague, not discoverable |
| `The Ultimate Skill for Processing CSV Files` | Too long (>64), verbose |
| `csv_processor_v2` | Underscore format, not gerund |

**Character Count Tool**:
```bash
echo -n "Your Skill Name" | wc -c
# Output: 16 characters
```

---

### `description` (Required)

The discovery and contextualization text. Claude uses this to decide when to activate the Skill.

**Constraints**:
- Maximum 1024 characters
- Third person perspective ("The Skill does X", NOT "I can do X")
- Includes both WHAT and WHEN
- Contains 3+ discoverable keywords
- Structured as: [Capabilities] + [Use cases/triggers]

**Format Template**:
```
"[Capability 1], [Capability 2], or [Capability 3].
Use when [trigger scenario 1], [trigger scenario 2],
or [trigger scenario 3]."
```

**Real-World Examples** ✅:

```yaml
# Example 1: PDF Processing
description: "Extract text and tables from PDF files,
create and edit presentations, or generate spreadsheets.
Use when working with PDF documents, PowerPoint files,
or when the user mentions document extraction or
data export."

# Example 2: Go Debugging
description: "Setup and use the Delve debugger for Go
applications with breakpoints, stack inspection, and
step-through execution. Use when debugging Go programs,
investigating runtime errors, or profiling performance
bottlenecks."

# Example 3: TypeScript Testing
description: "Write and execute Jest or Vitest unit tests
for TypeScript projects with mocking, snapshots, and
coverage reporting. Use when testing TypeScript code,
ensuring code quality, or implementing test-driven
development."
```

**Character Count**:
```bash
# Count description length
description="Your description text here"
echo -n "$description" | wc -c
```

**Trigger Keyword Audit**:

For each description, extract 3+ keywords Claude uses to decide activation:

```
Description: "Extract text and tables from PDF files..."
Keywords: ["PDF", "extract", "documents", "data export"]
           ↑ Would trigger on user mentioning these
```

### `allowed-tools` (Recommended)

Specifies which Claude Code tools this Skill can use. Restricts access for safety and clarity.

**Format**:
- Comma-separated list of tool names
- Can use wildcards for patterns: `Bash(python:*)`
- One tool per comma-separated item

**Tool Reference**:

| Tool Category | Examples | Use When |
|---------------|----------|----------|
| **File I/O** | `Read`, `Write`, `Edit`, `Glob` | Reading files, searching, writing output |
| **Search** | `Grep` | Pattern matching in files |
| **Execution** | `Bash(python:*)`, `Bash(git:*)` | Running specific commands |
| **Specialized** | `NotebookEdit` | Jupyter notebooks |
| **Meta** | `Bash(ls:*)`, `Bash(mkdir:*)` | Directory operations |

**Example Allowed-tools Lists**:

```yaml
# Analysis-only Skill
allowed-tools: "Read, Grep, Glob"

# File manipulation Skill
allowed-tools: "Read, Write, Edit, Bash(mkdir:*), Bash(touch:*)"

# Language-specific testing
allowed-tools: "Read, Bash(python:*), Bash(pytest:*), Bash(mypy:*)"

# Git workflow Skill
allowed-tools: "Read, Bash(git:*), Bash(gh:*)"

# Comprehensive (rare)
allowed-tools: "Read, Write, Edit, Glob, Grep, Bash(python:*), Bash(node:*)"
```

**Minimal Principle**:
- Include ONLY tools the Skill actually needs
- If a Skill only reads files, omit `Write` and `Edit`
- If it doesn't run code, omit `Bash(*)`
- Overly broad `allowed-tools` list reduces clarity

**Decision Matrix**:

```
Does the Skill need to create files?
├─ YES → Include Write, Edit
└─ NO → Omit them

Does the Skill need to run code?
├─ YES → Include Bash(language:*) for needed languages
└─ NO → Omit Bash

Does the Skill analyze text/patterns?
├─ YES → Include Grep, Read
└─ NO → Omit them
```

---

## Metadata Validation Checklist

Before publishing, verify all fields:

```
name:
  [ ] ≤ 64 characters
  [ ] Gerund format (action verb)
  [ ] Specific and discoverable (not generic)
  [ ] Matches skill domain

description:
  [ ] ≤ 1024 characters
  [ ] Includes what (capabilities)
  [ ] Includes when (triggers/use cases)
  [ ] Contains 3+ discoverable keywords
  [ ] Third person perspective
  [ ] No "I can", "we can", subjective language

allowed-tools:
  [ ] All tools listed are actually used
  [ ] No overly broad patterns (e.g., Bash(*) without specifics)
  [ ] Tools are comma-separated
  [ ] No typos in tool names
```

---

## Real-World Metadata Examples

### Example 1: High Freedom Skill (Principles)

```yaml
---
name: "Architecting Scalable Microservices"
description: "Design microservice architecture with patterns
for service decomposition, communication, and deployment. Use when
planning microservice strategies, discussing service boundaries,
analyzing scalability trade-offs, or implementing distributed systems."
allowed-tools: "Read, Glob"
---
```

**Analysis**:
- ✅ Name: Action verb + domain
- ✅ Description: Capabilities (design, patterns) + triggers (planning, discussing, analyzing)
- ✅ Keywords: "microservice", "architecture", "distributed", "scalability"
- ✅ Tools: Minimal (read-only, no execution)

### Example 2: Medium Freedom Skill (Patterns)

```yaml
---
name: "Writing Effective Unit Tests with Pytest"
description: "Write comprehensive unit tests using Pytest
with fixtures, mocking, and parametrization for Python projects.
Use when testing Python code, implementing test-driven development,
or improving test coverage."
allowed-tools: "Read, Bash(python:*), Bash(pytest:*)"
---
```

**Analysis**:
- ✅ Name: Action verb + framework
- ✅ Description: Specific capabilities (fixtures, mocking) + triggers (testing, TDD)
- ✅ Keywords: "testing", "Python", "Pytest", "unit tests"
- ✅ Tools: Specific to testing (python, pytest only)

### Example 3: Low Freedom Skill (Scripts)

```yaml
---
name: "Automating Deployment with Git Hooks"
description: "Set up and debug pre-commit, post-commit, and
pre-push Git hooks for automated linting, testing, and deployment.
Use when configuring CI/CD pipelines, preventing bad commits, or
automating deployment workflows."
allowed-tools: "Read, Write, Bash(git:*), Bash(bash:*)"
---
```

**Analysis**:
- ✅ Name: Action verb + mechanism
- ✅ Description: Concrete actions (setup, debug, configure) + use cases
- ✅ Keywords: "Git", "CI/CD", "automation", "deployment"
- ✅ Tools: Script execution for Git operations

---

## Common Metadata Mistakes & Fixes

| Mistake | Example | Fix |
|---------|---------|-----|
| **Name too generic** | `"Python Helper"` | `"Optimizing Python Code Performance"` |
| **Name too long** | `"The Complete Guide to Processing CSV Files in Python"` | `"Processing CSV Files with Python"` |
| **Description unclear** | `"Helps with stuff"` | `"Extract, validate, and transform CSV data. Use when..."` |
| **No triggers** | `"Database design patterns"` | `"...Use when designing databases, optimizing queries..."` |
| **Wrong perspective** | `"I can extract text from PDFs"` | `"Extracts text and tables from PDF files"` |
| **Over-scoped tools** | `allowed-tools: "Bash(*)"` | `allowed-tools: "Bash(python:*), Bash(pytest:*)"` |

---

## Metadata Interaction with Skill Discovery

Claude uses metadata to **decide whether to activate a Skill**:

```
User Request: "I need to extract tables from a PDF"
                ↓
Claude searches Skill descriptions for keywords:
├─ "PDF" → Found in moai-domain-pdf description ✓
├─ "extract" → Found in same description ✓
└─ "tables" → Found in same description ✓
                ↓
Result: Skill activates with ~90% confidence
```

**Weak Description** ❌:
```
"Processes documents"
User: "Extract PDF tables" → No match on "PDF", "extract", or "tables"
Result: Skill doesn't activate ✗
```

**Strong Description** ✅:
```
"Extract text and tables from PDF files, create presentations,
or generate spreadsheets. Use when working with PDF documents..."
User: "Extract PDF tables" → Matches "PDF", "extract", "tables"
Result: Skill activates ✓
```

---

## Metadata File Structure in Practice

### Correct Structure

```yaml
---
name: "Skill Name"
description: "Description here."
allowed-tools: "Tool1, Tool2"
---

# Heading (first line of content after metadata)
## Subheading
Content starts here...
```

### Common Errors

**❌ Missing starting dash**:
```yaml
name: "Skill Name"
description: "..."
---
```

**❌ Missing ending dash**:
```yaml
---
name: "Skill Name"
---
# Content here (metadata incomplete!)
```

**❌ Tabs instead of spaces**:
```yaml
---
	name: "Skill Name"  # TAB before name (invalid!)
---
```

**✅ Correct**:
```yaml
---
name: "Skill Name"
description: "Description text"
allowed-tools: "Tool1, Tool2"
---
```

---

## Multi-Model Metadata Compatibility

Metadata must work across Haiku, Sonnet, and Opus:

| Model | Consideration | Example |
|-------|---------------|---------|
| **Haiku** | Simpler keywords → trigger on broad patterns | Keywords: `["Python", "testing"]` |
| **Sonnet** | More nuanced keyword matching | Keywords: `["fixture-based testing", "parametrization"]` |
| **Opus** | Understands context and implicit triggers | Can infer "Go debugging" from mention of "breakpoints + delve" |

**Best Practice**: Write descriptions that work for all three models:
- Use common terms (accessible to Haiku)
- Add technical specificity (useful for Sonnet)
- Enable contextual understanding (leverages Opus)

---

## Version Control for Metadata

When updating a Skill's metadata:

```
Before:
name: "Processing CSV Files"

After:
name: "Processing CSV Files with Python"

Rationale: Added "with Python" for better discoverability
Updated: 2025-10-22
```

---

## Reference: Full Metadata Audit Template

```markdown
# Metadata Audit Checklist

## name
- [ ] Exactly 64 characters or fewer
- [ ] Begins with action verb (gerund)
- [ ] Specific to domain (not generic)
- [ ] No "Skill", "Helper", redundant prefixes
- [ ] Capitalized properly

## description
- [ ] Exactly 1024 characters or fewer
- [ ] Third person perspective
- [ ] Describes capabilities clearly
- [ ] Includes 3+ use case triggers
- [ ] Contains searchable keywords
- [ ] No subjective language ("great", "amazing")

## allowed-tools
- [ ] Comma-separated format
- [ ] Each tool actually used by Skill
- [ ] Specific patterns (not overly broad)
- [ ] No typos in tool names
- [ ] Minimal scope principle applied
```

---

## Quick Metadata Generator

Use this template to generate metadata for new Skills:

```yaml
---
name: "[VERB] [DOMAIN]"
description: "[CAPABILITY 1], [CAPABILITY 2], or [CAPABILITY 3].
Use when [TRIGGER 1], [TRIGGER 2], or [TRIGGER 3]."
allowed-tools: "[TOOL1], [TOOL2], [TOOL3]"
---
```

**Example Fills**:

```yaml
---
name: "Analyzing Performance Bottlenecks in Python"
description: "Profile Python applications using cProfile,
measure execution time, and identify memory leaks. Use when
optimizing performance, debugging slow functions, or analyzing
resource usage."
allowed-tools: "Read, Bash(python:*)"
---
```

---

**Version**: 0.3.0 (with Interactive Discovery & Web Research)
**Last Updated**: 2025-10-22
**Framework**: MoAI-ADK + Claude Code Skills + skill-factory
**Related Guides**:
- [INTERACTIVE-DISCOVERY.md](INTERACTIVE-DISCOVERY.md) — TUI survey patterns
- [WEB-RESEARCH.md](WEB-RESEARCH.md) — Web research strategies
- [SKILL-UPDATE-ADVISOR.md](SKILL-UPDATE-ADVISOR.md) — Skill analysis & updates
