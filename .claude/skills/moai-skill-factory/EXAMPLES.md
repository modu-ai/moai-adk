# moai-skill-factory - Working Examples

_Last updated: 2025-10-22_

## Example 1: Creating a New Language Skill

```bash
# Phase 1: Interactive Discovery
User: "Create a new Skill for TypeScript development"
→ skill-factory activates moai-alfred-tui-survey
→ User answers questions about scope, tools, patterns
→ Skill metadata drafted

# Phase 2: Web Research
→ skill-factory uses WebSearch for latest TypeScript versions
→ Fetches official docs: https://typescriptlang.org/docs
→ Validates tool versions (TypeScript 5.7, Vitest 2.1, etc.)

# Phase 3: Skill Generation
→ Creates SKILL.md with v2.0.0 structure
→ Generates examples.md with working code samples
→ Creates reference.md with CLI commands
→ Dual deployment to .claude/skills and templates/
```

## Example 2: Updating Existing Skills

```bash
# Batch update workflow
User: "Update all Python Skills to latest versions"
→ skill-factory analyzes current SKILL.md
→ WebSearch for latest tool versions (pytest 8.4.2, ruff 0.13.1)
→ Detects outdated patterns (black → ruff migration)
→ Generates update report with changelog
→ Applies updates across all affected Skills
```

## Example 3: Skill Quality Validation

```bash
# Pre-publication checklist
→ Validates YAML frontmatter (v2.0.0 schema)
→ Checks keywords field exists
→ Verifies tool version matrix present
→ Ensures examples.md and reference.md exist
→ Confirms dual deployment sync
→ Reports validation results
```

---

_For more examples, see SKILL.md and supporting documentation_
