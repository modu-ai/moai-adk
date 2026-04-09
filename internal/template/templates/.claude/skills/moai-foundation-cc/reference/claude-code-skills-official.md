# Claude Code Skills - Official Documentation Reference

Source: https://code.claude.com/docs/en/skills
Related: https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview
Updated: 2026-01-06

## What are Agent Skills?

Agent Skills are modular extensions that expand Claude's capabilities. They consist of a SKILL.md file with YAML frontmatter and Markdown instructions, plus optional supporting files (scripts, templates, documentation).

Key Characteristic: Skills are model-invoked, meaning Claude autonomously decides when to use them based on user requests and skill descriptions. This differs from slash commands which are user-invoked.

## Skill Types

Three categories of Skills exist:

1. Personal Skills: Located at `~/.claude/skills/skill-name/`, available across all projects
2. Project Skills: Located at `.claude/skills/skill-name/`, shared via git with team members
3. Plugin Skills: Bundled within Claude Code plugins

## Progressive Disclosure Architecture

Skills leverage Claude's VM environment with a three-level loading system that optimizes context window usage:

### Level 1: Metadata (Always Loaded)

The Skill's YAML frontmatter provides discovery information and is pre-loaded into the system prompt at startup. This lightweight approach means many Skills can be installed without context penalty.

Content: `name` and `description` fields from YAML frontmatter
Token Cost: Approximately 100 tokens per Skill

### Level 2: Instructions (Loaded When Triggered)

The main body of SKILL.md contains procedural knowledge including workflows, best practices, and guidance. When a request matches a Skill's description, Claude reads SKILL.md from the filesystem via bash, only then loading this content into the context window.

Content: SKILL.md body with instructions and guidance
Token Cost: Under 5K tokens recommended

### Level 3: Resources and Code (Loaded As Needed)

Skills can bundle additional materials that Claude accesses only when referenced:

- Instructions: Additional markdown files (FORMS.md, REFERENCE.md) containing specialized guidance
- Code: Executable scripts (fill_form.py, validate.py) that Claude runs via bash
- Resources: Reference materials like database schemas, API documentation, templates, or examples

Content: Bundled files executed via bash without loading contents into context
Token Cost: Effectively unlimited since they are accessed on-demand

## SKILL.md Structure and Format

### Directory Organization

skill-name/
- SKILL.md (required, main file, 500 lines or less)
- reference.md (optional, extended documentation)
- examples.md (optional, code examples)
- scripts/ (optional, utility scripts)
- templates/ (optional, file templates)

### YAML Frontmatter Requirements

Required Fields:

- name: Skill identifier (max 64 characters, lowercase letters, numbers, and hyphens only, no XML tags, no reserved words like "anthropic" or "claude")

- description: What the Skill does and when to use it (max 1024 characters, non-empty, no XML tags)

Optional Fields:

- allowed-tools: Tool names to restrict access. Supports comma-separated string or YAML list format. If not specified, Claude follows standard permission model.

- model: Model to use when Skill is active (e.g., `claude-sonnet-4-20250514`). Defaults to the current model.

- context: Set to `fork` to run Skill in isolated sub-agent context with separate conversation history.

- agent: Agent type when `context: fork` is set. Options: `Explore`, `Plan`, `general-purpose`. Defaults to `general-purpose`.

- hooks: Define lifecycle hooks (PreToolUse, PostToolUse, Stop) scoped to the Skill. See Hooks section below.

- user-invocable: Boolean to control slash command menu visibility. Default is `true`. Set to `false` to hide internal Skills from the menu.

### Advanced Frontmatter Examples (2026-01)

#### allowed-tools as YAML List

```yaml
---
name: reading-files-safely
description: Read files without making changes. Use for read-only file access.
allowed-tools:
  - Read
  - Grep
  - Glob
---
```

#### Forked Context with Agent Type

```yaml
---
name: code-analysis
description: Analyze code quality and generate detailed reports. Use for comprehensive code review.
context: fork
agent: Explore
allowed-tools:
  - Read
  - Grep
  - Glob
---
```

#### With Lifecycle Hooks

```yaml
---
name: secure-operations
description: Perform operations with additional security checks.
hooks:
  PreToolUse:
    - matcher: "Bash"
      hooks:
        - type: command
          command: "./scripts/security-check.sh $TOOL_INPUT"
          once: true
  PostToolUse:
    - matcher: "Write|Edit"
      hooks:
        - type: command
          command: "./scripts/verify-write.sh"
---
```

Hook Configuration Fields:
- type: "command" (bash) or "prompt" (LLM evaluation)
- command: Bash command to execute (for type: command)
- prompt: LLM prompt for evaluation (for type: prompt)
- timeout: Timeout in seconds (default: 60)
- matcher: Pattern to match tool names (regex supported)
- once: Boolean, run hook only once per session (Skills only)

#### Hidden from Menu

```yaml
---
name: internal-helper
description: Internal Skill used by other Skills. Not for direct user invocation.
user-invocable: false
allowed-tools:
  - Read
  - Grep
---
```

### Example SKILL.md Structure

```yaml
---
name: your-skill-name
description: Brief description of what this Skill does and when to use it. Include both what it does AND specific triggers for when Claude should use it.
allowed-tools: Read, Grep, Glob
---

# Your Skill Name

## Instructions
Clear, step-by-step guidance for Claude to follow.

## Examples
Concrete examples of using this Skill.
```

## Tool Restrictions with allowed-tools

The `allowed-tools` field restricts which tools Claude can use when a skill is active.

Use Cases for Tool Restrictions:

- Read-only Skills that should not modify files (allowed-tools: Read, Grep, Glob)
- Limited-scope Skills for data analysis only
- Security-sensitive workflows

If `allowed-tools` is not specified, Claude follows the standard permission model and may request tool access as needed.

## Writing Effective Descriptions

The description field enables Skill discovery and should include both what the Skill does and when to use it.

Critical Rules:

- Always write in third person. The description is injected into the system prompt, and inconsistent point-of-view can cause discovery problems.
- Good: "Processes Excel files and generates reports"
- Avoid: "I can help you process Excel files"
- Avoid: "You can use this to process Excel files"

Be Specific and Include Key Terms:

Effective examples:

- PDF Processing: "Extract text and tables from PDF files, fill forms, merge documents. Use when working with PDF files or when the user mentions PDFs, forms, or document extraction."

- Git Commit Helper: "Generate descriptive commit messages by analyzing git diffs. Use when the user asks for help writing commit messages or reviewing staged changes."

Avoid vague descriptions:

- "Helps with documents" (too vague)
- "Processes data" (not specific)
- "Does stuff with files" (unclear triggers)

## Naming Conventions

Recommended format uses gerund form (verb + -ing) for Skill names as this clearly describes the activity or capability:

Good Naming Examples:

- processing-pdfs
- analyzing-spreadsheets
- managing-databases
- testing-code
- writing-documentation

Acceptable Alternatives:

- Noun phrases: pdf-processing, spreadsheet-analysis
- Action-oriented: process-pdfs, analyze-spreadsheets

Avoid:

- Vague names: helper, utils, tools
- Overly generic: documents, data, files
- Reserved words: anthropic-helper, claude-tools
- Inconsistent patterns within skill collection

## Best Practices

### Core Principle: Concise is Key

The context window is a shared resource. Your Skill competes with system prompt, conversation history, other Skills' metadata, and the actual request.

Default Assumption: Claude is already very smart. Only add context Claude does not already have. Challenge each piece of information by asking:

- Does Claude really need this explanation?
- Can I assume Claude knows this?
- Does this paragraph justify its token cost?

### Set Appropriate Degrees of Freedom

Match the level of specificity to the task's fragility and variability.

High Freedom (Text-based instructions):

Use when multiple approaches are valid, decisions depend on context, or heuristics guide the approach.

Medium Freedom (Pseudocode or scripts with parameters):

Use when a preferred pattern exists, some variation is acceptable, or configuration affects behavior.

Low Freedom (Specific scripts, few or no parameters):

Use when operations are fragile and error-prone, consistency is critical, or a specific sequence must be followed.

### Test With All Models You Plan to Use

Skills act as additions to models, so effectiveness depends on the underlying model:

- Claude Haiku (fast, economical): Does the Skill provide enough guidance?
- Claude Sonnet (balanced): Is the Skill clear and efficient?
- Claude Opus (powerful reasoning): Does the Skill avoid over-explaining?

### Build Evaluations First

Create evaluations BEFORE writing extensive documentation to ensure your Skill solves real problems:

1. Identify gaps: Run Claude on representative tasks without a Skill, document specific failures
2. Create evaluations: Build three scenarios that test these gaps
3. Establish baseline: Measure Claude's performance without the Skill
4. Write minimal instructions: Create just enough content to adddess gaps and pass evaluations
5. Iterate: Execute evaluations, compare against baseline, refine

### Develop Skills Iteratively with Claude

Work with one instance of Claude ("Claude A") to create a Skill that will be used by other instances ("Claude B"):

1. Complete a task without a Skill using normal prompting
2. Identify the reusable pattern from the context you provided
3. Ask Claude A to create a Skill capturing that pattern
4. Review for conciseness
5. Improve information architecture
6. Test on similar tasks with Claude B
7. Iterate based on observation

## Progressive Disclosure Patterns

### Pattern 1: High-level Guide with References

Keep SKILL.md as overview pointing Claude to detailed materials:

```markdown
# PDF Processing

## Quick start
Extract text with pdfplumber (brief example)

## Advanced features
**Form filling**: See [FORMS.md](FORMS.md) for complete guide
**API reference**: See [REFERENCE.md](REFERENCE.md) for all methods
**Examples**: See [EXAMPLES.md](EXAMPLES.md) for common patterns
```

Claude loads additional files only when needed.

### Pattern 2: Domain-specific Organization

For Skills with multiple domains, organize content by domain to avoid loading irrelevant context:

```
bigquery-skill/
- SKILL.md (overview and navigation)
- reference/
  - finance.md (revenue metrics)
  - sales.md (pipeline data)
  - product.md (usage analytics)
```

When user asks about revenue, Claude reads only reference/finance.md.

### Pattern 3: Conditional Details

Show basic content, link to advanced content:

```markdown
## Creating documents
Use docx-js for new documents. See [DOCX-JS.md](DOCX-JS.md).

## Editing documents
For simple edits, modify the XML directly.
**For tracked changes**: See [REDLINING.md](REDLINING.md)
```

### Important: Avoid Deeply Nested References

Keep references one level deep from SKILL.md. Claude may partially read files when they are referenced from other referenced files, resulting in incomplete information.

Bad: SKILL.md references advanced.md which references details.md
Good: SKILL.md directly references all files (advanced.md, reference.md, examples.md)

## Where Skills Work

Skills are available across Claude's agent products with different behaviors:

### Claude Code

Custom Skills only. Create Skills as directories with SKILL.md files. Claude discovers and uses them automatically. Skills are filesystem-based and do not require API uploads.

### Claude API

Supports both pre-built Agent Skills and custom Skills. Specify the relevant `skill_id` in the `container` parameter. Custom Skills are shared organization-wide.

### Claude.ai

Supports both pre-built Agent Skills and custom Skills. Upload custom Skills as zip files through Settings, Features. Custom Skills are individual to each user and not shared organization-wide.

### Claude Agent SDK

Supports custom Skills through filesystem-based configuration. Create Skills in `.claude/skills/` and enable by including "Skill" in `allowed_tools` configuration.

## Security Considerations

We strongly recommend using Skills only from trusted sources: those you created yourself or obtained from Anthropic. Skills provide Claude with new capabilities through instructions and code, and a malicious Skill can direct Claude to invoke tools or execute code in ways that do not match the Skill's stated purpose.

Key Security Considerations:

- Audit thoroughly: Review all files bundled in the Skill including SKILL.md, scripts, images, and other resources
- External sources are risky: Skills that fetch data from external URLs pose particular risk
- Tool misuse: Malicious Skills can invoke tools in harmful ways
- Data exposure: Skills with access to sensitive data could leak information to external systems
- Treat like installing software: Only use Skills from trusted sources

## Managing Skills

### View Available Skills

Ask Claude directly: "What Skills are available?"

Or check file system:

- Personal Skills: ls ~/.claude/skills/
- Project Skills: ls .claude/skills/

### Update a Skill

Edit SKILL.md directly. Changes apply on next Claude Code startup.

### Remove a Skill

Personal: rm -rf ~/.claude/skills/my-skill
Project: rm -rf .claude/skills/my-skill && git commit -m "Remove unused Skill"

## Debugging Skills

### Claude Not Using the Skill

Check if description is specific enough:

- Include what it does AND when to use it
- Add key trigger terms users will mention

Check YAML syntax validity:

- Opening and closing --- markers
- Proper indentation
- No tabs (use spaces)

Check correct file location:

- Personal: ~/.claude/skills/*/SKILL.md
- Project: .claude/skills/*/SKILL.md

### Multiple Skills Conflicting

Use distinct trigger terms in descriptions:

Instead of two skills both having "For data analysis" and "For analyzing data", use specific triggers:

- Skill 1: "Analyze sales data in Excel files and CRM exports. Use for sales reports, pipeline analysis, and revenue tracking."
- Skill 2: "Analyze log files and system metrics data. Use for performance monitoring, debugging, and system diagnostics."

## Anti-patterns to Avoid

### Avoid Windows-style Paths

Always use forward slashes in file paths, even on Windows:

- Good: scripts/helper.py, reference/guide.md
- Avoid: scripts\helper.py, reference\guide.md

### Avoid Offering Too Many Options

Do not present multiple approaches unless necessary. Provide a default with an escape hatch for special cases.

### Avoid Time-sensitive Information

Do not include information that will become outdated. Use "old patterns" section for deprecated approaches instead of date-based conditions.

## Checklist for Effective Skills

Before sharing a Skill, verify:

Core Quality:

- Description is specific and includes key terms
- Description includes both what the Skill does and when to use it
- SKILL.md body is under 500 lines
- Additional details are in separate files if needed
- No time-sensitive information
- Consistent terminology throughout
- Examples are concrete, not abstract
- File references are one level deep
- Progressive disclosure used appropriately
- Workflows have clear steps

Testing:

- At least three evaluations created
- Tested with Haiku, Sonnet, and Opus
- Tested with real usage scenarios
- Team feedback incorporated if applicable
