# SPEC-SKILL-STANDARDS-001: Claude Code Skills Standardization

**Environment:**
- Project: MoAI-ADK
- Date: 2025-11-21
- Scope: All skill files in `.claude/skills/` directories
- Standard: Claude Code Official Standards v1.0

**Assumptions:**
- Current skills use custom format with extensive metadata
- Official Claude Code standards require minimal YAML frontmatter
- Skills must be compatible with both Claude Code VS Code extension and CLI
- Progressive disclosure structure preferred for long skills
- Need to maintain functionality while simplifying format

**Requirements:**

### R1: Format Compliance
- All skills MUST use official Claude Code YAML frontmatter format
- YAML frontmatter MUST contain only required fields: `name`, `description`
- Optional field: `allowed-tools` for tool permission control
- Maximum skill name length: 64 characters, lowercase with hyphens only
- Maximum description length: 1024 characters, no XML tags

### R2: Content Structure
- Skills MUST follow progressive disclosure structure (Level 1-2-3)
- Content MUST be under 500 lines total
- File structure MUST use `SKILL.md` in skill directory
- Third-person descriptions required
- Examples MUST be concrete and actionable

### R3: Naming Conventions
- Skill names MUST use gerund form (verb-ing)
- Names MUST be lowercase with hyphens only
- Names MUST be descriptive and specific (avoid "helper", "tools")
- Names MUST not contain reserved words: "anthropic", "claude"

### R4: Tool Usage
- Skills MUST explicitly handle errors
- Skills MUST solve problems completely
- Required packages MUST be listed
- File paths MUST use forward slashes
- Script permissions MUST be set with `chmod +x`

**Specifications:**

### S1: YAML Frontmatter Standard
```yaml
---
name: skill-identifier (max 64 chars, lowercase-hyphens)
description: Brief description what this skill does and when to use it (max 1024 chars)
allowed-tools: Read, Grep, Glob, Bash (optional, comma-separated)
---
```

### S2: Content Organization Standard
```markdown
# Skill Title

## Instructions
Clear, step-by-step guidance for Claude to follow

## Examples
Concrete examples of using this skill

[Optional additional sections as needed]
```

### S3: Directory Structure Standard
```
.claude/skills/
├── skill-name/
│   ├── SKILL.md (required)
│   ├── examples.md (optional)
│   ├── reference.md (optional)
│   └── scripts/ (optional, executable files)
```

### S4: Progressive Disclosure Pattern
- **Level 1**: Quick reference (30-second value)
- **Level 2**: Implementation guide (common patterns)
- **Level 3**: Advanced patterns (expert reference)

### S5: Quality Standards
- Skills MUST be under 500 lines
- Examples MUST be specific and actionable
- Feedback loops MUST be included for workflows
- Checklists MUST be provided for multi-step processes
- Context7 integration SHOULD be documented when applicable

**Traceability:**
- TAG: SKILL-STANDARDS-001
- Related: [Documentation Standards], [Code Quality Standards]
- Dependencies: None
- Conflicts: Current custom skill format

**Acceptance Criteria:**
1. All 144 skills in main directory converted to standard format
2. All 378 skills in template directory converted to standard format
3. YAML frontmatter compliance 100%
4. Content structure compliance 100%
5. Naming convention compliance 100%
6. All skills under 500 lines
7. All skills have concrete examples
8. Tool permissions properly specified