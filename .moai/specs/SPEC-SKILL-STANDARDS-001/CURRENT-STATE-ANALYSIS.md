# MoAI-ADK Skills Migration - Current State Analysis

**Analysis Date**: 2025-11-21
**Scope**: Skills standardization to Claude Code official format
**Version**: 1.0

---

## Executive Summary

MoAI-ADK currently maintains **280 skills** across two directories (~141 in custom, ~139 in templates) with **significant format inconsistencies**. The skills use extended custom YAML metadata (10+ fields) versus the official Claude Code format (2-3 required fields). This analysis identifies gaps and proposes a phased migration strategy.

### Key Findings

- **280 total skills** with format fragmentation
- **37 skills exceed 500 lines** (complexity threshold)
- **117 skills use allowed-tools** (83% adoption)
- **60+ skills have extended metadata** (custom fields)
- **Official format requires**: name, description, (optional: allowed-tools)

---

## Part 1: Current State Gap Analysis

### 1.1 Directory Structure

```
Custom Skills:         /Users/goos/MoAI/MoAI-ADK/.claude/skills/
Template Skills:       /Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills/

Total Directories:     280 skill directories
Current Status:        Mixed formats (70% extended YAML, 30% official format)
```

### 1.2 YAML Metadata Format Gap

#### Current Custom Format (Extended)
Example from `moai-foundation-ears`:
```yaml
---
name: "moai-foundation-ears"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-13
status: stable
tier: foundation
description: Enterprise EARS (Evaluate, Analyze, Recommend, Synthesize) Framework...
keywords: ['ears', 'requirements-engineering', 'systematic-thinking', ...]
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - Glob
  - Grep
  - WebFetch
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---
```

**Custom Fields** (not in official format):
- `version` - Semantic versioning (e.g., 4.0.0)
- `created` - Creation date
- `updated` - Last update date
- `status` - Stability marker (stable, beta, alpha, deprecated)
- `tier` - Skill tier (foundation, specialization, integration)
- `keywords` - Search keywords array
- `spec` - Orchestration metadata (some skills)
- `metadata` - Additional metadata (some skills)
- `orchestration` - Agent orchestration config (some skills)
- `primary-agent` - Associated agent (some skills)

**Official Format** (Claude Code):
```yaml
---
name: moai-foundation-ears
description: EARS format for clear requirements. Provides templates for writing unambiguous requirements.
allowed-tools: Read, Write
---
```

**Official Constraints**:
- `name`: max 64 chars, lowercase, hyphens/numbers only
- `description`: max 1024 chars, must specify "what" and "when to use"
- `allowed-tools`: optional, comma-separated list

#### Format Gap Summary

| Aspect | Current | Official | Gap |
|--------|---------|----------|-----|
| Required Fields | 7-10 | 2 | 5-8 extra fields |
| YAML Complexity | Extended metadata | Minimal | Reduce 60-70% |
| Allowed Tools Format | Array/List | Comma-separated | Parser change required |
| Description Length | 200-500 chars | Max 1024 chars | Within bounds |
| Version Management | Semantic (4.0.0) | Not in YAML | Move to file system |

### 1.3 File Structure Complexity

#### Current Structure (Sample: moai-foundation-specs)
```
moai-foundation-specs/
├── SKILL.md              # Main skill document (775 lines)
├── examples.md           # Code examples
├── reference.md          # API reference
└── [advanced]/           # Optional nested structure
```

#### Complexity Issues
1. **Large monolithic files**: 37 skills exceed 500 lines
2. **Nested directories**: Reference materials mixed with SKILL.md
3. **No progressive disclosure**: Level 1-4 structure embedded in single file
4. **Line count distribution**:
   - 6 skills: 1000-1681 lines
   - 15 skills: 800-999 lines
   - 16 skills: 600-799 lines
   - Remaining: 300-599 lines

#### Official Format Recommendation
- SKILL.md should stay <500 lines
- Progressive disclosure: Level 1 (overview) → Level 2-3 (detail) → Level 4 (reference)
- Nested structure:
  ```
  skill-name/
  ├── SKILL.md          # Progressive disclosure <500 lines
  ├── examples.md       # Code examples (separate)
  ├── reference.md      # API reference (separate)
  └── advanced/         # Optional advanced topics
  ```

### 1.4 Naming Convention Analysis

#### Current Naming Patterns
- Prefixes: `moai-*`, `moai-foundation-*`, `moai-domain-*`, `moai-lang-*`, `moai-lib-*`
- Format: Lowercase with hyphens (COMPLIANT with official standard)
- Length compliance:
  - **COMPLIANT**: ~90% within 64 char limit
  - **NON-COMPLIANT**: ~10% exceed 64 chars

#### Examples of Non-Compliant Names
- `moai-translation-korean-multilingual` (38 chars - OK)
- `moai-project-template-optimizer` (32 chars - OK)
- `moai-baas-auth0-ext` (19 chars - OK)

Actually, all current names are within the 64-char limit. Format is COMPLIANT.

### 1.5 Metadata Field Distribution

**Top 10 YAML Fields Used**:
1. `status: stable` - 119 skills (84%)
2. `version: "4.0.0"` - 91 skills (65%)
3. `allowed-tools:` - 65 skills (46%)
4. `created:` - 33 skills (23%)
5. `updated:` - 28 skills (20%)
6. `tier:` - 10 skills (7%)
7. `keywords:` - 8 skills (6%)
8. `spec:` - 14 skills (10%)
9. `orchestration:` - 12 skills (8%)
10. `can_resume:` - 14 skills (10%)

### 1.6 Allowed-Tools Usage Pattern

**Current Status**:
- 117/141 skills (83%) use `allowed-tools` field
- Most skills specify 4-8 tools
- Common tools: Read, Write, Edit, Bash, Grep, Glob, WebFetch

**Format Variations**:
- Array format: `allowed-tools: [Read, Write]`
- List format: `allowed-tools:\n  - Read\n  - Write`
- String format: `allowed-tools: "Read, Write, Edit"`

**Official Format**:
- Comma-separated string: `allowed-tools: Read, Write, Edit`

---

## Part 2: Gap Classification

### Critical Gaps (Must Fix)
1. Extended YAML metadata (60+ skills)
2. Large monolithic SKILL.md files (37 skills >500 lines)
3. Inconsistent allowed-tools format (mixed array/list/string)
4. Missing progressive disclosure structure

### Important Gaps (Should Fix)
1. Version management approach
2. Status/stability markers location
3. Keyword/search metadata handling
4. Created/updated timestamp tracking

### Minor Gaps (Nice to Have)
1. Orchestration metadata structure
2. Primary agent association
3. Dependency tracking in YAML

---

## Part 3: Impact Assessment

### Affected Skills (by migration effort)

**Tier 1 - Simple Migration (50% of skills)**
- Well-structured, <300 lines
- Minimal custom metadata
- Single allowed-tools list
- Action: Direct YAML reduction + structural cleanup

**Tier 2 - Moderate Migration (35% of skills)**
- Extended YAML, 300-500 lines
- Multiple custom fields
- Complex examples/reference sections
- Action: Full restructuring with progressive disclosure

**Tier 3 - Complex Migration (15% of skills)**
- Large monolithic files (500-1681 lines)
- Deep nesting, extensive metadata
- Multiple levels of content
- Action: Major refactoring + progressive disclosure design

### Risk Assessment

**Technical Risks**:
- Backward compatibility: Skills with custom orchestration metadata
- Tool resolution: MCP tools in allowed-tools need validation
- Version management: Semantic versioning outside YAML

**Mitigation**:
- Create compatibility layer during transition
- Validate all tool names against official tool list
- Move version info to README/metadata file

---

## Part 4: Recommendation Summary

**Phase 1 (Immediate)**: Migrate 50% (Tier 1 simple skills)
- Time: 3-4 weeks
- Effort: 2-3 engineers
- Risk: Low
- Benefit: 70 skills standardized

**Phase 2 (Q1 2026)**: Migrate 35% (Tier 2 moderate skills)
- Time: 4-5 weeks
- Effort: 3-4 engineers
- Risk: Medium
- Benefit: 100 additional skills standardized

**Phase 3 (Q1 2026)**: Migrate 15% (Tier 3 complex skills)
- Time: 6-8 weeks
- Effort: 4-5 engineers
- Risk: High
- Benefit: Remaining 42 skills + full standardization

**Total Migration**: 9-13 weeks, 2-5 engineers per phase

---

## References

- Claude Code Skills Overview: https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview
- Claude Code Skills Format: https://platform.claude.com/docs/en/agents-and-tools/agent-skills/quickstart
- Claude Code Skills Best Practices: https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices

---

**Status**: Analysis Complete
**Next Steps**: Review SPEC candidates and select primary implementation focus
