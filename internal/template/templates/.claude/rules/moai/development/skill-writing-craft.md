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
| moai-domain-frontend | "frontend development specialist for React 19 and Next.js 16" | Clear domain + framework + scope |

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
- The capability is handled by general knowledge (e.g., "basic syntax" for moai-domain-frontend)
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

The three-level progressive disclosure structure (Level 1 metadata / Level 2 body / Level 3 bundled, with token budgets and skill-listing / post-compaction budget) is canonical in `.claude/rules/moai/development/skill-authoring.md` § Progressive Disclosure; this section provides the body-structure templates that fit each level.

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

> Canonical: the skill frontmatter schema lives in `.claude/rules/moai/development/skill-authoring.md` § Standard Fields (agentskills.io) / § Key Format Rules / § Schema Example. This file owns only the prose craft (Parts 1-2 above: description craft + body structure); it does not restate the schema table.

Schema and validation rules live in `skill-authoring.md` § Key Format Rules; common frontmatter patterns live in `skill-authoring.md` § Schema Example. This section does not restate them — Parts 1-2 above cover the unique description and body-structure craft.

---

## Schema Validation

Before committing a skill, verify:

- [ ] Frontmatter fields all present and valid
- [ ] `name` matches actual skill capability
- [ ] `description` is a concise trigger summary within the 1,536-char cap (combined `description` + `when_to_use`)
- [ ] `paths` is CSV string (not YAML array)
- [ ] `description` trigger keywords accurately reflect skill scope
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
