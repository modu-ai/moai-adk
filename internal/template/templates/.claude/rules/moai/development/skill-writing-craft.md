---
description: "Skill writing guidelines covering description craft, body structure, and schema"
paths: ".claude/skills/**/SKILL.md"
---

<!-- Source: revfactory/harness — Apache License 2.0 — see .claude/rules/moai/NOTICE.md -->

# Skill Writing Craft

Guidelines for writing high-quality skills: description triggering, body structure, and schema validation.

---

## Part 1: Description Craft

The description field is **read-only metadata** that Claude Code uses to decide whether to load your skill into context. It's not displayed to users.

### Writing a Good Description

**Purpose**: Summarize when the skill should trigger, in a single sentence. Optimize for Claude Code's context matcher.

**Template**:
```
{Capability Summary}: {Domain Keywords} for {Use Case}
```

**Examples**:

| Skill | Good Description | Why |
|-------|------------------|-----|
| moai-domain-frontend | "React 19 / Next.js 16 component development with modern patterns" | Clear domain (React/Next) + capability (components) + keywords |
| moai-ref-testing-pyramid | "Test strategy and pyramid patterns for unit/integration/E2E coverage" | Clear domain (testing) + specific patterns + coverage levels |
| moai-library-shadcn | "shadcn/ui component library for React with TypeScript integration" | Clear library name + framework + integration |

**Bad Descriptions** (too vague, won't trigger):
- "General programming help" (too broad)
- "Code quality" (ambiguous — which kind?)
- "Documentation tool" (doesn't specify what documentation)

### When to Trigger vs When to Skip

**The skill should trigger when the user is**:
- Explicitly asking for the domain (e.g., "help with React hooks")
- Working in the domain (e.g., modifying a `.tsx` file)
- Referencing the specific capability (e.g., "create component library")

**The skill should NOT trigger when**:
- User is asking about a different domain (e.g., "help with Python" to a React skill)
- The capability is handled by general knowledge (e.g., "basic syntax" for moai-library-shadcn)
- The skill would duplicate general MoAI capabilities (e.g., moai-domain-backend shouldn't trigger for "debug this Go error")

### Paths Frontmatter for Smart Triggering

Use the `paths` field to control when your skill is auto-loaded by Claude Code:

```yaml
---
description: "Skill description"
paths: ".claude/skills/**/*.md,**/hooks/**/*.ts"
---
```

**Path Patterns** (glob syntax):
- `**/*.tsx` — Any TypeScript React file
- `**/*_test.go` — Any Go test file
- `.moai/config/**/*.yaml` — Configuration files
- `src/api/**/*` — API directory

**Strategy**:
- Use `paths` for automatic domain detection
- Use `description` as fallback for manual triggering
- Combine: paths + description = high-precision triggering

---

## Part 2: Body Structure — Progressive Disclosure

The skill body should follow a three-level progressive disclosure structure:

### Level 1: Quick Reference (100-200 lines)

**Contents**:
- One-line summary
- Key concepts (5-7 bullet points)
- When to use this skill
- Typical workflow (3-5 steps)
- Links to deeper content

**Token Cost**: ~1,500 tokens

**Example Structure**:
```markdown
# Skill Name

One-line summary: what this skill provides

## Quick Concepts
- Concept A: explanation
- Concept B: explanation
- Concept C: explanation

## When to Use
- Scenario 1
- Scenario 2

## Typical Workflow
1. Step A
2. Step B
3. Step C

## See Also
- [Deep dive topic] — modules/topic.md
- [Examples] — examples.md
```

### Level 2: Implementation Guide (500-1,500 lines)

**Contents**:
- Detailed workflows with code snippets
- Real-world patterns
- Integration examples
- Configuration reference
- Common patterns and their implementation

**Token Cost**: ~3,000-5,000 tokens

**Structure**:
```markdown
## Implementation Guide

### Pattern 1: Name
- Description
- When to apply
- Example code (pseudo-code or multi-language)
- Integration with MoAI

### Pattern 2: Name
...

### Configuration
- Key settings
- Defaults
- Override mechanisms
```

### Level 3: Advanced & Reference (Unlimited)

**Contents**:
- Deep technical dives
- Edge cases and workarounds
- Performance optimization
- Troubleshooting guides
- Full API reference

**Token Cost**: On-demand (user explicitly requests)

**Structure**:
```markdown
## Advanced Topics

### Topic 1: Name
[Full technical treatment]

### Topic 2: Name
[Extended examples]

## Reference
- [External documentation]
- [Related specs]
- [Tool configuration]
```

### File Organization for Large Skills

When body exceeds 500 lines, split into modules:

```
.claude/skills/my-skill/
├── SKILL.md (Quick Reference + Implementation, ≤500 lines)
├── modules/
│   ├── advanced-patterns.md
│   ├── troubleshooting.md
│   └── reference.md
├── examples.md
└── reference.md
```

**Linking Between Levels**:
- SKILL.md references → `See also: [Topic] — modules/topic.md`
- Examples → `Examples: [Name] — examples.md#section`
- External resources → `Reference: [Tool] — reference.md#api`

---

## Part 3: Frontmatter Schema

Every skill MUST have complete frontmatter:

```yaml
---
name: "Display Name of Skill"
description: "One-line description for context matching"
type: skill
paths: "**/*.tsx,**/__tests__/**"
domains: ["frontend", "testing"]
model: "default"
effort: "high"
tools: ["WebSearch", "WebFetch", "Bash", "Read", "Write", "Edit", "Grep", "Glob"]
allowed-tools: "WebSearch,WebFetch,Bash,Read,Write,Edit,Grep,Glob"
---
```

### Field Reference

| Field | Required | Type | Notes |
|-------|----------|------|-------|
| `name` | ✅ | string | Display name, 30-50 chars |
| `description` | ✅ | string | Single sentence, ≤80 chars, context-matching optimized |
| `type` | ✅ | string | Always `"skill"` |
| `paths` | Optional | string | Glob pattern CSV (no YAML array) |
| `domains` | Optional | array | Topic categories for organization |
| `model` | Optional | string | `"default"`, `"opus"`, `"sonnet"` |
| `effort` | Optional | string | `"low"`, `"medium"`, `"high"`, `"xhigh"` |
| `tools` | Optional | array | Full tool array (if providing access) |
| `allowed-tools` | Optional | string | CSV string of tool names |

### Validation Rules

**Rule 1: `paths` is CSV string, NOT YAML array**
```yaml
# WRONG
paths:
  - "**/*.tsx"
  - "**/__tests__/**"

# CORRECT
paths: "**/*.tsx,**/__tests__/**"
```

**Rule 2: Single-sentence description**
```yaml
# WRONG
description: "Frontend development. React and Next.js. Components and hooks."

# CORRECT
description: "React 19 and Next.js 16 component development with hooks"
```

**Rule 3: tool vs allowed-tools**
```yaml
# Use EITHER tools (array) OR allowed-tools (CSV string), not both

# OPTION A: Array form
tools:
  - WebSearch
  - Bash
  - Read

# OPTION B: CSV form
allowed-tools: "WebSearch,Bash,Read"
```

**Rule 4: Accurate domains**
```yaml
domains: ["frontend", "testing"]  # Correct
domains: ["frontend"]  # Missing testing even though skill covers both

# Use domains that match skill content
```

### Common Frontmatter Patterns

**API/CLI Reference Skill**:
```yaml
---
name: "Tool Name Reference"
description: "Tool Name CLI commands and API reference for common use cases"
type: skill
paths: "**/*.ts,**/*.py,README.md"
domains: ["reference", "api"]
effort: "low"
allowed-tools: "WebFetch,Read,Grep"
---
```

**Code Implementation Skill**:
```yaml
---
name: "Framework Feature"
description: "Framework Feature implementation patterns with examples"
type: skill
paths: "src/**/*.tsx,**/*_test.tsx"
domains: ["framework", "implementation"]
effort: "high"
allowed-tools: "WebSearch,WebFetch,Bash,Read,Write,Edit,Grep,Glob"
---
```

**Analysis/Research Skill**:
```yaml
---
name: "Topic Analysis"
description: "Topic Analysis and research methodology with patterns"
type: skill
paths: "docs/**,**/*.md,SPEC-*"
domains: ["research", "analysis"]
effort: "medium"
allowed-tools: "WebSearch,WebFetch,Read,Grep"
---
```

---

## Schema Validation

Before committing a skill, verify:

- [ ] Frontmatter fields all present and valid
- [ ] `name` matches actual skill capability
- [ ] `description` is single sentence, ≤80 characters
- [ ] `paths` is CSV string (not YAML array)
- [ ] `domains` accurately reflect skill scope
- [ ] `allowed-tools` matches actual tool usage in body
- [ ] No tool used without being listed in frontmatter
- [ ] Body follows progressive disclosure (Quick/Implementation/Advanced if >500 lines)
- [ ] Cross-references in body match actual file paths
- [ ] Examples are copy-paste ready with comments
- [ ] External references are verified (URLs valid)

---

## Skill Quality Checklist

**Before marking a skill complete**:

- [ ] Description triggers correctly (tested with Claude Code)
- [ ] Quick Reference is scannable (≤200 lines, clear structure)
- [ ] Implementation examples are language-neutral (pseudo-code or multi-language)
- [ ] No hardcoded framework/language bias
- [ ] Advanced section covers edge cases
- [ ] Examples run without modification (or clearly noted limitations)
- [ ] All external links verified and current
- [ ] Frontmatter complete and valid
- [ ] Domains reflect actual coverage
- [ ] No duplicate content with other skills

This craft guide ensures skills are discoverable, usable, and maintainable.
