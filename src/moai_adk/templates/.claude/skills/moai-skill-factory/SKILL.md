---
name: moai-skill-factory
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: Create and maintain high-quality Claude Code Skills through interactive discovery, web research, and continuous updates. Use when building new Skills, researching latest best practices, updating existing Skills with current information, or generating Skill packages backed by official documentation and real-world examples.
keywords: [skill-creation, claude-skills, best-practices, web-research, interactive-discovery, skill-updates]
allowed-tools:
  - Read
  - Write
  - Glob
  - Bash
  - WebFetch
  - WebSearch
---
# Generating High-Quality Claude Code Skills

> **Quick Reference**: See `METADATA.md` for metadata authoring, `STRUCTURE.md` for file organization, `EXAMPLES.md` for real-world case studies, and `CHECKLIST.md` for validation.

---

## What Are Skills?

**Skills** are modular, reusable knowledge capsules that extend Claude's capabilities in Claude Code. They combine strategic guidance, best practices, templates, and automation scripts into organized filesystem packages.

### Three-Level Progressive Disclosure Model

Skills load in stages to minimize context overhead:

```
Level 1: Metadata (Always Active)
├── name, description, allowed-tools
└── ~100 tokens total (minimal overhead)
    ↓
Level 2: Instructions (Triggered On Demand)
├── SKILL.md main body (workflows, patterns, best practices)
└── Loads when Claude recognizes relevance to the request
    ↓
Level 3: Resources (As Needed)
├── Supporting files (reference.md, examples.md)
├── Scripts, templates
└── Consumed only when explicitly referenced or accessed
```

**Key Benefit**: Many Skills loaded at session start without context penalty. Full content loads only when relevant.

---

## Core Principles for Skill Creation

### 1. Conciseness Over Completeness

```
❌ Don't: Explain how algorithms work in detail (assume Claude knows)
✅ Do: Reference patterns, provide decision trees, link to examples
```

SKILL.md should stay **under 500 lines**. Trust that Claude is already intelligent—provide guidance, not lectures.

### 2. Appropriate Freedom Levels

Match specificity to task fragility:

| Freedom Level | When to Use                | Example                        | Content Style                          |
| ------------- | -------------------------- | ------------------------------ | -------------------------------------- |
| **High**      | Flexible, creative work    | Architecture design, strategy  | Principles, trade-offs, considerations |
| **Medium**    | Standard patterns exist    | Data validation workflows      | Pseudocode, flowcharts, annotated code |
| **Low**       | Deterministic, error-prone | Bash operations, file handling | Specific scripts with error handling   |

### 3. Multi-Model Testing

Verify Skill works across:
- **Haiku**: Can it understand concise examples? Does it activate correctly?
- **Sonnet**: Does it exploit the full Skill? Synthesize across sections?
- **Opus**: Can it extend patterns beyond examples?

### 4. Consistent Terminology

```
✅ Good: Use "migration", "index", "constraint" consistently
❌ Bad: Switch between "migration", "schema change", "version bump"
```

Maintain a glossary for domain-specific terms.

### 5. Progressive Disclosure Through Linking

```
Main SKILL.md:
├── Overview + Quick Start
└── "For detailed reference, see [reference.md](reference.md)"

reference.md:
├── Detailed specs, API docs, edge cases
└── "Example implementations in [examples.md](examples.md)"

examples.md:
├── 3-4 real-world scenarios
└── "Utility scripts available in scripts/"
```

Never create deeply nested structures—one level of indirection is enough.

---

## SKILL Package File Organization

### Recommended Structure

```
skill-name/
├── SKILL.md                    # Main instructions (~400 lines)
├── METADATA.md                 # Metadata authoring guide (optional)
├── STRUCTURE.md                # File organization reference (optional)
├── reference.md                # Detailed reference docs (optional)
├── examples.md                 # 3-4 real-world examples (optional)
├── CHECKLIST.md                # Quality validation (optional)
├── scripts/                    # Utility scripts
│   ├── helper_script.sh
│   └── validator.py
└── templates/                  # Reusable templates
    ├── template_main.md
    └── config-template.json
```

### Rules for File Organization

| Rule                | Rationale                    | Example                                               |
| ------------------- | ---------------------------- | ----------------------------------------------------- |
| One level deep      | Avoids nested discovery      | ✅ `reference.md` not `docs/reference/api.md`          |
| Relative paths      | Cross-platform compatibility | ✅ `[see reference.md](reference.md)` not `./docs/...` |
| Forward slashes     | Unix convention              | ✅ `scripts/helper.sh` not `scripts\helper.sh`         |
| Clear naming        | Self-documenting             | ✅ `validate-skill.sh` not `v.sh`                      |
| <500 lines per file | Manageable context           | Split if SKILL.md exceeds 500 lines                   |

---

## YAML Metadata Requirements

### Frontmatter Structure

```yaml
---
name: "Skill Name"
description: "What it does and when to use it"
allowed-tools: "Tool1, Tool2, Tool3"
---
```

### Field Specifications

#### `name` (Required)
- **Max 64 characters**
- **Recommended Format**: Gerund (action verb) + domain
- **Examples**:
  - ✅ "Processing CSV Files with Python"
  - ✅ "Debugging Go Applications with Delve"
  - ❌ "CSV Processing" (too generic)
  - ❌ "Python CSV Helper Tool" (too descriptive)

#### `description` (Required)
- **Max 1024 characters**
- **Format**: Third person, action-oriented
- **Must include**:
  - What the Skill does (capabilities)
  - When to use it (trigger scenarios)
  - 3+ discoverable keywords

**Template**:
```
"[Capability 1], [Capability 2], or [Capability 3].
Use when [trigger scenario 1], [trigger scenario 2],
or [trigger scenario 3]."
```

**Example**:
```
"Extract text and tables from PDF files, create and edit spreadsheets,
generate business reports, or merge documents. Use when working with
PDF documents, Excel files, CSV data, or when the user mentions document
extraction, data export, or report generation."
```

**Trigger Keywords** (examples):
- Problem domain: PDF, Excel, CSV, JSON, XML
- Operation: parsing, validation, transformation, cleaning
- Tech stack: Python, TypeScript, Go, Rust
- Pattern: algorithm, design pattern, architecture

#### `allowed-tools` (Recommended)
- **Minimal principle**: Only include tools the Skill actually uses
- **Format**: Comma-separated list of tool names and patterns
- **Examples**:
  - `"Read, Grep, Glob"` (read-only analysis)
  - `"Read, Write, Edit, Bash(git:*)"` (Git operations)
  - `"Bash(python:*), Bash(pytest:*), Read"` (Python testing)
  - `"Read, Bash(curl:*), Bash(jq:*)"` (API interaction)

---

## Content Structure: Freedom Level Framework

### High Freedom: Principles & Guidance (30-40% of SKILL.md)

For creative, flexible tasks where direction matters more than prescription.

**Content Type**:
- Core principles (3-5 key ideas)
- Trade-off analysis
- Decision matrices
- Consideration frameworks

**Example: Architecture Decisions**
```
## Architecture Trade-offs

| Pattern       | Pros                       | Cons                        | When to use                 |
| ------------- | -------------------------- | --------------------------- | --------------------------- |
| Monolith      | Simple to start            | Scales poorly               | MVP, <10 microservices      |
| Microservices | Scales, independent deploy | Complex networking          | 10+ teams, different stacks |
| Serverless    | Zero ops, elastic          | Cold starts, vendor lock-in | Event-driven, variable load |
```

### Medium Freedom: Patterns & Pseudocode (40-50% of SKILL.md)

For standard patterns where a reference implementation aids understanding.

**Content Type**:
- Pseudocode with annotations
- Flowcharts/decision trees
- Pattern examples (not production code)
- Configuration templates

**Example: Data Validation Workflow**
```
## Pseudocode Pattern: CSV Validation

```pseudocode
1. Load CSV file
2. For each column:
   a. Check data type matches schema
   b. Validate range/constraints
   c. Flag missing values
3. If all checks pass:
   ✓ Proceed to transformation
4. Else:
   ✗ Report errors with row/column numbers
   ✗ Suggest fixes
```

### Low Freedom: Specific Scripts & Commands (10-20% of SKILL.md)

For deterministic, error-prone operations where exact execution matters.

**Content Type**:
- Ready-to-run Bash/Python scripts
- Error handling & validation
- Exit codes & logging
- Hardened templates

**Example: Safe File Operation**
```bash
#!/bin/bash
set -euo pipefail  # Exit on error, undefined var, pipe fail

SOURCE_FILE="$1"
DEST_DIR="$2"

# Validation
if [[ ! -f "$SOURCE_FILE" ]]; then
  echo "ERROR: Source file not found: $SOURCE_FILE" >&2
  exit 1
fi

if [[ ! -d "$DEST_DIR" ]]; then
  echo "ERROR: Destination directory not found: $DEST_DIR" >&2
  exit 1
fi

# Safe copy with backup
cp "$SOURCE_FILE" "$DEST_DIR/"
echo "✓ File copied successfully to $DEST_DIR"
exit 0
```

---

## Writing Style Guidelines

### Terminology Consistency

**Audit Checklist**:
- [ ] Define each domain term once (glossary section)
- [ ] Use the same term consistently throughout (no synonyms)
- [ ] Italicize or bold terms on first use: *term*, **term**

**Example Glossary**:
```markdown
### Glossary

- **Index**: B-tree data structure for search optimization
- **Constraint**: Data integrity rule (NOT "validation rule")
- **Migration**: Schema version change (NOT "upgrade", "deploy")
```

### Example Coverage

Every major concept needs at least one example:

```
✅ GOOD:
## Pseudocode Pattern
[pattern description]
[pseudocode or flowchart]

### Example 1: Simple Case
[code example]

### Example 2: Edge Case
[code example]

❌ BAD:
## Pseudocode Pattern
[pattern description only, no examples]
```

### Anti-Patterns to Avoid

| Anti-Pattern        | Example                             | Fix                                              |
| ------------------- | ----------------------------------- | ------------------------------------------------ |
| Time-sensitive info | "Today's date is 2025-01-15"        | "Use current system date"                        |
| Vague instructions  | "Make it fast"                      | "Profile with [tool], measure [metric]"          |
| Too many options    | 10+ approaches listed               | Pick 3 main approaches, others in "Alternatives" |
| Nested references   | `docs/api/v2/reference/examples.md` | Keep to one level: `reference.md`                |
| Windows paths       | `docs\api\reference.md`             | Use `/`: `docs/api/reference.md`                 |

---

## Quality Validation Checklist

Before publishing, audit your Skill:

### Metadata Completeness
- [ ] `name` ≤ 64 characters, gerund format
- [ ] `description` ≤ 1024 chars, includes 3+ triggers
- [ ] `allowed-tools` is minimal and justified
- [ ] YAML frontmatter is valid (no tabs, proper indentation)

### Content Quality
- [ ] SKILL.md ≤ 500 lines
- [ ] All major concepts have examples
- [ ] Terminology is consistent throughout
- [ ] Anti-patterns avoided (no time-sensitive data, no nested paths)
- [ ] Progressive Disclosure pattern applied
- [ ] Relative paths used throughout

### Specificity & Discoverability
- [ ] Name is specific (not "Python Helper")
- [ ] Description includes problem domain keywords
- [ ] Triggers match likely user search terms
- [ ] Scope is bounded (1-3 related domains)

### Multi-Model Compatibility
- [ ] Haiku can understand concise examples
- [ ] Sonnet exploits full Skill capabilities
- [ ] Opus can extend patterns beyond examples

### Security Posture
- [ ] No credentials, API keys, or secrets in examples
- [ ] No system state assumptions (e.g., "assuming ~/config exists")
- [ ] Scripts include error handling & exit codes
- [ ] Dangerous operations (rm, mv) use explicit confirmations

---

## Freedom Level Decision Framework

Use this tree to determine high/medium/low freedom mix:

```
1. Is the task deterministic?
   ├─ YES: Well-defined algorithm, fixed steps
   │   └─→ Low Freedom: Specific script with error handling
   │       (Example: Bash file operations, deployment steps)
   │
   └─ NO: Flexible, context-dependent, multiple valid approaches
       └─ Does the task have standard patterns?
           ├─ YES: Established best practice (e.g., validation)
           │   └─→ Medium Freedom: Pseudocode + examples
           │       (Example: Data validation, testing patterns)
           │
           └─ NO: Novel or creative work
               └─→ High Freedom: Principles + considerations
                   (Example: Architecture design, code organization)
```

---

## Real-World Skill Examples

See `[examples.md](examples.md)` for three complete case studies:
1. **High Freedom Example**: Architecture Design Skill
2. **Medium Freedom Example**: Testing Patterns Skill
3. **Low Freedom Example**: Deployment Automation Skill

Each demonstrates complete metadata, content structure, and freedom levels.

---

## File Templates & Scaffolding

### Creating a New Skill

1. **Generate structure** (see `STRUCTURE.md`):
   ```
   skill-name/
   ├── SKILL.md
   ├── reference.md
   ├── examples.md
   ├── scripts/
   └── templates/
   ```

2. **Add frontmatter** to SKILL.md:
   ```yaml
   ---
   name: "[Gerund] [Domain]"
   description: "[Capability]. Use when [trigger 1], [trigger 2]."
   allowed-tools: "Tool1, Tool2"
   ---
   ```

3. **Structure content** (High + Medium + Low):
   - Principles/overview (High)
   - Patterns & pseudocode (Medium)
   - Ready-to-run scripts (Low)

4. **Validate** against CHECKLIST.md

5. **Test** across Haiku, Sonnet, Opus

### Template Files

See `templates/` directory:
- `SKILL_TEMPLATE.md`: Base Skill structure
- `reference-template.md`: Supporting docs template
- `examples-template.md`: Example scenarios template
- `scripts-template.sh`: Error-handling Bash template

---

## Common Failure Modes & Fixes

| Issue                    | Root Cause                   | Fix                                               |
| ------------------------ | ---------------------------- | ------------------------------------------------- |
| **Not activating**       | Description too generic      | Add 5+ specific keywords, mention common triggers |
| **Haiku ignores it**     | Examples too complex         | Simplify pseudocode, add short examples           |
| **Over-specifying**      | Too much low-freedom content | Reduce scripts, increase principles               |
| **Scope creep**          | Covers too many domains      | Split into 2-3 focused Skills                     |
| **Terminology mismatch** | Terms used inconsistently    | Audit and standardize with glossary               |
| **File too large**       | SKILL.md > 500 lines         | Move content to reference.md or examples.md       |

---

## Integration with moai-alfred-skill-generator

This Skill is invoked by the **moai-alfred-skill-generator** Sub-Agent during Skill creation phases:

```
User: "/alfred:1-plan Create Skill for X"
    ↓
skill-generator Sub-Agent (ANALYZE → DESIGN → ASSURE phases)
    ↓
moai-skill-generator Skill (PRODUCE phase)
    ↓
Complete Skill package created
```

---

## Quick Start: Creating Your First Skill

### 5-Step Process

1. **Define Problem** (10 min):
   - What gap does this Skill fill?
   - Who uses it? What are trigger scenarios?

2. **Design Metadata** (10 min):
   - Name (gerund + domain)
   - Description (capabilities + triggers)
   - Allowed-tools (minimal list)

3. **Structure Content** (30 min):
   - High freedom: Principles (20%)
   - Medium freedom: Patterns (50%)
   - Low freedom: Scripts (30%)

4. **Add Examples & References** (30 min):
   - 3-4 concrete examples in examples.md
   - Detailed reference in reference.md
   - Reusable templates in templates/

5. **Validate & Test** (20 min):
   - Run through CHECKLIST.md
   - Test with Haiku, Sonnet, Opus
   - Peer review (if available)

**Total Time**: ~2 hours for comprehensive Skill

---

---

## Advanced Features: Interactive Discovery & Web Research

See [INTERACTIVE-DISCOVERY.md](INTERACTIVE-DISCOVERY.md) for detailed guidance on:
- Using TUI surveys for user engagement
- Structuring effective questions
- Clarifying vague requirements

See [WEB-RESEARCH.md](WEB-RESEARCH.md) for detailed guidance on:
- Web search strategies for latest information
- Source prioritization and validation
- Official documentation integration
- Fact-checking and deprecation detection

See [SKILL-UPDATE-ADVISOR.md](SKILL-UPDATE-ADVISOR.md) for detailed guidance on:
- Analyzing existing Skills for updates
- Detecting outdated patterns and versions
- Proposing data-driven improvements
- Generating migration guides

---

## Key References

- [METADATA.md](METADATA.md) — Metadata authoring guide
- [STRUCTURE.md](STRUCTURE.md) — File organization patterns
- [EXAMPLES.md](EXAMPLES.md) — Real-world Skill case studies
- [CHECKLIST.md](CHECKLIST.md) — Pre-publication validation
- [INTERACTIVE-DISCOVERY.md](INTERACTIVE-DISCOVERY.md) — TUI survey patterns (NEW)
- [WEB-RESEARCH.md](WEB-RESEARCH.md) — Web research strategies (NEW)
- [SKILL-UPDATE-ADVISOR.md](SKILL-UPDATE-ADVISOR.md) — Skill analysis & updates (NEW)

---

**Version**: 0.3.0 (with Interactive Discovery, Web Research, & Update Advisor)
**Last Updated**: 2025-10-22
**Framework**: MoAI-ADK + Claude Skills + skill-factory Sub-Agent
**Status**: Production-ready
**Key Features**:
- ✅ Interactive user surveys via moai-alfred-interactive-questions
- ✅ Web research integration for latest information
- ✅ Skill analysis & update recommendations
- ✅ Official documentation validation
- ✅ Progressive Disclosure pattern
- ✅ Freedom level framework
- ✅ Multi-model compatibility testing
