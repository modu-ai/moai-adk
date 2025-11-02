# Skill File Organization Guide

This guide covers directory structure, file naming conventions, Progressive Disclosure patterns, and best practices for organizing Skill packages.

---

## Recommended File Structure

### Complete Skill Package (All Components)

```
skill-name/
├── SKILL.md                         # Main instructions & framework
├── METADATA.md                      # Metadata authoring guide (optional)
├── STRUCTURE.md                     # File organization guide (optional)
├── reference.md                     # Detailed reference documentation
├── examples.md                      # Real-world usage examples
├── CHECKLIST.md                     # Quality validation (optional)
├── scripts/                         # Executable utilities
│   ├── helper1.sh
│   ├── helper2.py
│   └── validator.js
└── templates/                       # Reusable templates
    ├── template-main.md
    ├── template-config.yaml
    └── template-script.sh
```

### Minimal Skill Package (Essentials Only)

```
skill-name/
├── SKILL.md                         # Main content (required)
└── examples.md                      # Usage examples (optional)
```

### Medium Skill Package (Balanced)

```
skill-name/
├── SKILL.md                         # Main content
├── reference.md                     # Supporting documentation
├── examples.md                      # Real-world examples
└── scripts/                         # Utilities
    └── helper.sh
```

---

## File Organization Principles

### Rule 1: One-Level Deep

**Principle**: Keep all files one level below the Skill root. No nested subdirectories.

```
✅ GOOD:
skill-name/
├── SKILL.md
├── reference.md
└── scripts/helper.sh

❌ BAD:
skill-name/
├── docs/
│   ├── reference.md
│   └── api/
│       └── details.md
└── src/
    └── scripts/
        └── helper.sh
```

**Why**: Progressive Disclosure requires flat structure. Claude loads file references incrementally; nested structures complicate discovery.

### Rule 2: Relative Path References

**Principle**: Use relative paths for all cross-file links. No absolute paths or external URLs (except documentation).

```
✅ GOOD (SKILL.md):
See [reference.md](reference.md) for detailed API docs.
For examples, visit [examples.md](examples.md).
Scripts available in scripts/helper.sh.

❌ BAD:
See /Users/goos/MoAI/MoAI-ADK/.claude/skills/skill-name/reference.md
See /home/user/project/.claude/skills/skill-name/reference.md
See https://example.com/api-docs.md
```

**Platform Compatibility**: Forward slashes work on macOS, Linux, and Windows Git Bash.

### Rule 3: Forward Slashes

**Principle**: Always use forward slashes (`/`) for path separators, never backslashes (`\`).

```
✅ GOOD:
scripts/helper.sh
templates/config-template.yaml

❌ BAD:
scripts\helper.sh
templates\config-template.yaml
```

**Why**: Unix convention, Git-compatible, works across all platforms.

### Rule 4: Clear Naming Convention

**Principle**: File names should be self-documenting. Use kebab-case for multi-word files.

```
✅ GOOD:
├── SKILL.md
├── reference.md
├── examples.md
├── scripts/
│   ├── validate-schema.sh
│   ├── deploy-service.py
│   └── test-integration.js
└── templates/
    ├── config-template.yaml
    ├── deployment-script.sh
    └── docker-template
```

**Naming Patterns**:
| Type | Format | Example |
|------|--------|---------|
| Main docs | `UPPERCASE.md` | `SKILL.md`, `CHECKLIST.md` |
| Supporting | `lowercase.md` | `reference.md`, `examples.md` |
| Scripts | `kebab-case.ext` | `validate-schema.sh` |
| Templates | `word-template.ext` or `template-word.ext` | `config-template.yaml` |

### Rule 5: Manageable File Size

**Principle**: Keep SKILL.md under 500 lines. Split large content into supporting files.

```
SKILL.md size guide:
├── 0-200 lines → Concise Skill (OK)
├── 200-400 lines → Standard Skill (Recommended)
├── 400-500 lines → Large Skill (Split if possible)
└── 500+ lines → Too large, needs restructuring
```

**Strategy for Large Skills**:
```
Original (SKILL.md: 800 lines):
├── Overview
├── Framework (200 lines)
├── Patterns (300 lines)
├── Advanced Topics (200 lines)
└── Reference (100 lines)

Restructured:
SKILL.md (400 lines):
├── Overview
├── Framework (200 lines)
├── Patterns (summary, 80 lines)
└── "For advanced topics, see reference.md"

reference.md (300 lines):
├── Advanced Topics (300 lines)

examples.md (200 lines):
├── Pattern Examples
└── Advanced Use Cases
```

---

## Progressive Disclosure Pattern

Progressive Disclosure enables efficient context loading across three levels:

### Level 1: Metadata (Always Active ~100 tokens)

```yaml
---
name: "Skill Name"
description: "What it does. Use when..."
allowed-tools: "Tool1, Tool2"
---
```

**When Loaded**: Session startup
**Context Cost**: ~100 tokens total across all Skills
**Purpose**: Claude learns when to activate each Skill

### Level 2: Main Instructions (On-Demand ~1000-2000 tokens)

```markdown
# Main Skill Content

## Overview
## Quick Start
## Framework (High/Medium/Low Freedom)
## Examples
## Links to supporting docs
```

**When Loaded**: When Claude recognizes request relevance
**Context Cost**: ~1000-2000 tokens for active Skill
**Purpose**: Detailed guidance, decision trees, pseudocode

### Level 3: Supporting Resources (As-Needed 0-3000 tokens)

```
├── reference.md (detailed specs)
├── examples.md (real-world cases)
├── scripts/ (executable utilities)
└── templates/ (reusable scaffolding)
```

**When Loaded**: Only when explicitly referenced or needed
**Context Cost**: Only consumed when accessed
**Purpose**: Deep details, working examples, ready-to-use code

### Example Progressive Disclosure Flow

```
User: "Debug my Python app"
    ↓
Claude searches Skill descriptions (L1 metadata active)
    ↓
Finds "Debugging Python Applications"
    ↓
Loads SKILL.md (L2) for main framework
    ↓
User: "Show me how to inspect variables"
    ↓
Claude references [examples.md](examples.md) (L3)
    ↓
Loads examples.md with concrete code
    ↓
User: "I need the delve reference"
    ↓
Claude could reference [reference.md](reference.md) (L3)
```

---

## File Reference Patterns

### Linking from SKILL.md

```markdown
# Main Skill

See [reference.md](reference.md) for:
- API reference
- Configuration options
- Edge cases

[examples.md](examples.md) demonstrates:
- Basic usage
- Advanced patterns
- Real-world scenarios

Scripts in `scripts/` directory:
- `helper.sh` for automation
- `validator.py` for validation
```

### Linking from reference.md to examples.md

```markdown
# Detailed Reference

## Configuration

See practical examples in [examples.md](examples.md).
```

### Linking Back to Main

```markdown
# Examples

For conceptual framework, see [SKILL.md](SKILL.md).
For API details, see [reference.md](reference.md).
```

### Linking to Scripts

```markdown
# Automation

Run `scripts/deploy-service.sh` to automate deployment.

The script includes:
- Error handling
- Validation checks
- Status logging
```

---

## Supporting File Templates

### reference.md Structure

```markdown
# Reference: [Skill Domain]

## API Reference
- Function signatures
- Parameters
- Return values
- Examples

## Configuration
- Config file format
- All available options
- Defaults

## Edge Cases
- Known limitations
- Workarounds
- Performance considerations

## Glossary
- Key terms defined
- Acronyms explained
```

### examples.md Structure

```markdown
# Examples: [Skill Domain]

## Example 1: Basic Usage
[Concrete scenario]
[Code example]
[Expected output]

## Example 2: Intermediate Pattern
[Scenario description]
[Code example]
[Expected output]

## Example 3: Advanced Use Case
[Complex scenario]
[Code example]
[Expected output]

## Example 4: Edge Case
[Unusual scenario]
[Code example]
[Expected output]
```

### scripts/ Organization

```
scripts/
├── helper1.sh           # Bash utility
├── helper2.py           # Python utility
├── validator.js         # JavaScript utility
└── README.md            # Scripts documentation (optional)
```

**Script Requirements**:
- Shebang line: `#!/bin/bash`, `#!/usr/bin/env python3`, etc.
- Error handling: `set -euo pipefail` for Bash
- Exit codes: 0 for success, non-zero for failure
- Usage comments: Document what script does and how to run it

### templates/ Organization

```
templates/
├── template-main.md              # Markdown template
├── config-template.yaml          # YAML config template
├── deployment-script.sh          # Bash script template
└── docker-template               # Docker configuration
```

**Template Naming**:
- For `.md` files: `template-name.md` or `name-template.md`
- For configs: `template.yaml`, `template.json`
- For scripts: `template-action.sh`

---

## File Size Guidelines

| File | Recommended Size | Limit | Action |
|------|------------------|-------|--------|
| SKILL.md | 300-400 lines | 500 lines | Split into reference.md if exceeds |
| reference.md | 200-300 lines | 500 lines | Create multiple reference files if needed |
| examples.md | 150-250 lines | 400 lines | Split into use-case-specific examples |
| Individual script | 50-100 lines | 200 lines | Split or create helper library |
| Individual template | Variable | N/A | Keep readable, consider documentation |

---

## Version Control for File Organization

### Tracking Structure Changes

```
CHANGELOG Entry:
Added: reference.md (new file)
Moved: advanced-patterns from SKILL.md to reference.md
Updated: example links in SKILL.md to reference.md

Rationale: Reduced SKILL.md from 520 to 380 lines,
improved Progressive Disclosure structure.
```

### Git Best Practices

```bash
# Good commit messages for structure changes
git add .claude/skills/skill-name/
git commit -m "refactor: organize skill documentation into reference.md"

git add .claude/skills/skill-name/scripts/
git commit -m "feat: add validation script for skill testing"
```

---

## Directory Creation Template

### Bash Script to Create Standard Structure

```bash
#!/bin/bash
SKILL_NAME=$1

mkdir -p "$SKILL_NAME/scripts"
mkdir -p "$SKILL_NAME/templates"

# Create basic files
touch "$SKILL_NAME/SKILL.md"
touch "$SKILL_NAME/reference.md"
touch "$SKILL_NAME/examples.md"
touch "$SKILL_NAME/scripts/.gitkeep"
touch "$SKILL_NAME/templates/.gitkeep"

echo "✓ Skill structure created for: $SKILL_NAME"
```

**Usage**:
```bash
./create-skill.sh my-new-skill
```

---

## Common Structure Anti-Patterns

| Anti-Pattern | Example | Issue | Fix |
|--------------|---------|-------|-----|
| **Over-nesting** | `docs/api/v2/reference/details.md` | Breaks Progressive Disclosure | Use flat: `reference.md` |
| **Absolute paths** | `/Users/name/.claude/skills/...` | Not portable, breaks on different systems | Use relative: `scripts/helper.sh` |
| **Windows paths** | `scripts\helper.sh` | Not Git-compatible | Use: `scripts/helper.sh` |
| **One giant file** | SKILL.md with 2000 lines | Exceeds context, breaks Progressive Disclosure | Split into reference.md, examples.md |
| **Unclear naming** | `s1.sh`, `ref.md`, `ex.txt` | Not self-documenting | Use: `validate-schema.sh`, `reference.md` |
| **Circular links** | A → B → A | Confusing navigation | Establish clear hierarchy: Main → Supporting |

---

## Real-World Structure Examples

### Example 1: Minimal Skill

```
moai-skill-basics/
├── SKILL.md (280 lines, complete guidance)
└── examples.md (simple examples)
```

### Example 2: Balanced Skill

```
moai-skill-testing/
├── SKILL.md (380 lines, framework + patterns)
├── reference.md (220 lines, API reference)
├── examples.md (150 lines, 3 examples)
└── scripts/
    └── generate-tests.sh (automation)
```

### Example 3: Comprehensive Skill

```
moai-skill-deployment/
├── SKILL.md (400 lines, main framework)
├── reference.md (300 lines, detailed specs)
├── examples.md (200 lines, 4 scenarios)
├── CHECKLIST.md (validation)
├── scripts/
│   ├── validate-config.sh
│   ├── pre-deploy-check.py
│   └── rollback.sh
└── templates/
    ├── deployment-config-template.yaml
    ├── deployment-script.sh
    └── health-check-template.sh
```

---

## Migration: Restructuring Existing Skills

### Scenario: SKILL.md Too Large (850 lines)

**Before**:
```
skill-name/
└── SKILL.md (850 lines: overview + patterns + advanced + reference)
```

**Analysis**:
- Overview: 50 lines (keep in SKILL.md)
- Patterns: 300 lines (medium freedom content)
- Advanced topics: 200 lines (less common use)
- Reference: 300 lines (detailed specs)

**After**:
```
skill-name/
├── SKILL.md (350 lines: overview + patterns only)
├── reference.md (300 lines: specs + API)
├── examples.md (100 lines: advanced use cases)
└── "For advanced topics, see reference.md"
```

**Update Links**:
```markdown
# In SKILL.md
For detailed API reference, see [reference.md](reference.md).
Advanced scenarios are in [examples.md](examples.md).
```

---

## Validation: Structure Checklist

```
Organization:
  [ ] All files one level below Skill root
  [ ] No nested subdirectories deeper than 1 level
  [ ] Scripts in scripts/ directory
  [ ] Templates in templates/ directory

File Naming:
  [ ] SKILL.md in root
  [ ] Supporting files use lowercase.md
  [ ] Scripts use kebab-case.ext
  [ ] Templates use template-name.ext

Path References:
  [ ] All links use relative paths
  [ ] Forward slashes used (/)
  [ ] No Windows backslashes
  [ ] No absolute paths

File Size:
  [ ] SKILL.md ≤ 500 lines
  [ ] reference.md ≤ 500 lines
  [ ] examples.md ≤ 400 lines
  [ ] Each script ≤ 200 lines

Progressive Disclosure:
  [ ] Metadata complete
  [ ] SKILL.md stands alone
  [ ] Links to reference.md, examples.md provided
  [ ] Scripts documented but not required
```

---

**Version**: 0.3.0 (with Interactive Discovery & Web Research)
**Last Updated**: 2025-10-22
**Framework**: MoAI-ADK + Claude Code Skills + skill-factory
**Related Guides**: [SKILL.md](SKILL.md), [METADATA.md](METADATA.md), [EXAMPLES.md](EXAMPLES.md), [WEB-RESEARCH.md](WEB-RESEARCH.md)
