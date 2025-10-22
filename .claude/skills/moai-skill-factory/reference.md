# Skill Factory Reference

> Complete guide to creating high-quality Claude Code Skills

_Last updated: 2025-10-22_

---

## Skill Creation Workflow

### Phase 1: Discovery (15 min)

**Objective**: Define scope and gather requirements

**Activities**:
1. Identify the gap or need
2. Research official documentation
3. Collect examples and best practices
4. Define target audience and trigger scenarios

**Deliverables**:
- Problem statement (1-2 sentences)
- List of 5+ trigger keywords
- 3-5 reference URLs
- Target skill tier (Foundation/Essentials/Alfred/Domain/Language)

---

### Phase 2: Design (20 min)

**Objective**: Structure the skill content and metadata

**Activities**:
1. Write skill name (gerund + domain, ≤64 chars)
2. Draft description (capabilities + triggers, ≤1024 chars)
3. List allowed tools (minimal set)
4. Outline content structure (High/Medium/Low freedom)

**Deliverables**:
- Complete YAML frontmatter
- Content outline with sections
- Freedom level breakdown (% allocation)

---

### Phase 3: Production (60 min)

**Objective**: Write the skill content

**Activities**:
1. Write SKILL.md (main instructions, ≤500 lines)
2. Create reference.md (detailed docs)
3. Write examples.md (3-4 real-world scenarios)
4. Add scripts/ and templates/ (if needed)

**Deliverables**:
- Complete skill package
- All documentation files
- Supporting resources

---

### Phase 4: Validation (20 min)

**Objective**: Ensure quality and consistency

**Activities**:
1. Run through CHECKLIST.md
2. Test with multiple Claude models (Haiku/Sonnet)
3. Verify metadata completeness
4. Check for anti-patterns

**Deliverables**:
- Validation report
- Test results
- Quality gate confirmation

---

## Skill Metadata Design

### Name Format

**Pattern**: `<gerund-verb> <domain>`

**Examples**:
```
✅ Debugging Go Applications with Delve
✅ Processing CSV Files with Python
✅ Managing Docker Containers
✅ Implementing OAuth2 Authentication

❌ Python Helper  # Too generic
❌ Go Debug Tool  # Not gerund form
❌ CSV Processing Utility with Multiple Format Support  # Too long
```

### Description Template

```
[Primary capability], [secondary capability], or [tertiary capability].
Use when [trigger scenario 1], [trigger scenario 2], or [user mentions keywords].
```

**Real Example**:
```
Extract text and tables from PDF files, create and edit spreadsheets,
generate business reports, or merge documents. Use when working with
PDF documents, Excel files, CSV data, or when the user mentions document
extraction, data export, or report generation.
```

**Keyword Density**: Aim for 10-15 discoverable keywords in description.

---

## Content Structure Guidelines

### Freedom Level Framework

| Level | Allocation | Content Type | Example |
|-------|-----------|--------------|---------|
| **High** | 30-40% | Principles, trade-offs, considerations | Architecture decisions |
| **Medium** | 40-50% | Patterns, pseudocode, annotated code | Data validation workflows |
| **Low** | 10-20% | Ready-to-run scripts, specific commands | Bash operations, file handling |

### Section Templates

#### High Freedom Section

```markdown
## Architecture Trade-offs

| Pattern | Pros | Cons | When to Use |
|---------|------|------|-------------|
| Monolith | Simple deployment | Hard to scale | MVP, single team |
| Microservices | Independent scaling | Complex networking | 10+ teams |
| Serverless | Zero ops | Cold starts | Event-driven workloads |

**Decision Framework**:
1. Assess team size and structure
2. Consider deployment frequency
3. Evaluate scaling requirements
4. Analyze operational complexity
```

#### Medium Freedom Section

```markdown
## Data Validation Pattern

**Pseudocode**:
```pseudocode
1. Load data from source
2. For each field:
   a. Check type matches schema
   b. Validate constraints (range, format)
   c. Flag violations with location
3. If all checks pass:
   → Proceed to next stage
4. Else:
   → Report errors with suggested fixes
```

**Example (Python)**:
```python
def validate_user_input(data: dict) -> ValidationResult:
    errors = []
    if not isinstance(data.get("age"), int):
        errors.append("Age must be integer")
    if data.get("age", 0) < 0:
        errors.append("Age must be non-negative")
    return ValidationResult(valid=len(errors) == 0, errors=errors)
```
```

#### Low Freedom Section

```markdown
## Safe File Copy Script

```bash
#!/bin/bash
set -euo pipefail  # Exit on error, undefined var, pipe fail

SOURCE="$1"
DEST="$2"

# Validation
if [[ ! -f "$SOURCE" ]]; then
  echo "ERROR: Source file not found: $SOURCE" >&2
  exit 1
fi

if [[ ! -d "$DEST" ]]; then
  echo "ERROR: Destination directory not found: $DEST" >&2
  exit 1
fi

# Execute with confirmation
cp "$SOURCE" "$DEST/"
echo "✓ File copied: $SOURCE → $DEST"
exit 0
```

**Usage**:
```bash
./safe-copy.sh /path/to/source.txt /path/to/dest/
```
```

---

## Web Research Integration

### Research Strategy

**Step 1: Identify Official Sources**
```
Priority 1: Official documentation (project websites, GitHub repos)
Priority 2: Framework authors' blogs/articles
Priority 3: Industry standards (IEEE, IETF, W3C)
Priority 4: Reputable developer platforms (MDN, Microsoft Docs)
```

**Step 2: Validate Currency**
```
- Check publication/update dates (prefer 2024-2025)
- Verify version numbers match current stable releases
- Cross-reference multiple sources for consistency
- Flag deprecated patterns or outdated recommendations
```

**Step 3: Extract Patterns**
```
- Copy canonical examples
- Note best practices explicitly stated
- Document anti-patterns warned against
- Capture tool version requirements
```

### Search Query Templates

```
# Official documentation
"<technology> official documentation 2024 2025"
"<technology> best practices guide latest"

# Tool versions
"<tool> latest stable version 2025"
"<framework> migration guide <old-version> to <new-version>"

# Best practices
"<domain> industry standards 2024"
"<technology> production best practices"
```

---

## Example Coverage Requirements

### Minimum Requirements

- **3-4 complete examples** per skill
- Each example must include:
  - Context (what problem it solves)
  - Code/configuration snippets
  - Expected output
  - Common pitfalls

### Example Template

```markdown
## Example 1: [Scenario Name]

### Context
Brief description of the problem and solution approach.

### Prerequisites
- Dependency 1
- Dependency 2

### Implementation

```language
# Code example with inline comments
def example_function():
    """Docstring explaining purpose."""
    # Implementation details
    pass
```

### Expected Output
```
Sample output demonstrating success
```

### Common Pitfalls
- ❌ **Problem**: Description of common mistake
- ✅ **Solution**: How to avoid it
```

---

## Anti-Patterns to Avoid

### Time-Sensitive Information

❌ **DON'T**:
```markdown
Today's date is 2025-01-15, so use the latest version released last week.
```

✅ **DO**:
```markdown
Use the latest stable version (check official releases page).
```

### Vague Instructions

❌ **DON'T**:
```markdown
Make it fast and efficient.
```

✅ **DO**:
```markdown
Profile with `py-spy` to identify bottlenecks. Target: <10ms latency for 95th percentile.
```

### Nested File References

❌ **DON'T**:
```markdown
See docs/api/v2/reference/examples/advanced.md
```

✅ **DO**:
```markdown
See [reference.md](reference.md) for detailed API documentation.
```

---

## Progressive Disclosure Pattern

### Three-Level Information Architecture

**Level 1: SKILL.md** (Always loaded)
- Quick start (5-10 min read)
- Core concepts (high-level)
- Common patterns
- Link to level 2

**Level 2: reference.md** (On-demand)
- Detailed specifications
- Complete API reference
- Edge cases and constraints
- Link to level 3

**Level 3: examples.md + scripts/** (As needed)
- Real-world scenarios
- Production-ready templates
- Utility scripts
- Troubleshooting guide

### Linking Strategy

```markdown
<!-- SKILL.md -->
For detailed API reference, see [reference.md](reference.md).

<!-- reference.md -->
For working examples, see [examples.md](examples.md).

<!-- examples.md -->
Utility scripts available in [scripts/](scripts/) directory.
```

---

## Quality Validation Checklist

### Metadata Validation

- [ ] Name ≤64 characters, gerund format
- [ ] Description ≤1024 characters, includes 5+ keywords
- [ ] `allowed-tools` is minimal and justified
- [ ] YAML frontmatter is valid (proper indentation, no tabs)

### Content Quality

- [ ] SKILL.md ≤500 lines
- [ ] All major concepts have examples
- [ ] Terminology is consistent (glossary provided)
- [ ] No time-sensitive information
- [ ] Relative paths used throughout
- [ ] Freedom level breakdown documented

### Specificity & Discoverability

- [ ] Name is specific (not "Helper" or "Utility")
- [ ] Description includes problem domain keywords
- [ ] Triggers match likely user search terms
- [ ] Scope is bounded (1-3 related domains)

### Multi-Model Testing

- [ ] Haiku: Can understand concise examples
- [ ] Sonnet: Exploits full skill capabilities
- [ ] All models: Activate on correct triggers

---

## Skill Versioning Strategy

### Version Numbering

Follow Semantic Versioning:
- **Major** (X.0.0): Breaking changes, restructure
- **Minor** (0.X.0): New sections, significant additions
- **Patch** (0.0.X): Typo fixes, clarifications

### Changelog Format

```markdown
## Changelog

### v2.0.0 (2025-10-22)
- **BREAKING**: Restructure freedom levels
- **FEATURE**: Add web research integration
- **FEATURE**: Include 5 new examples

### v1.1.0 (2025-08-15)
- **FEATURE**: Add Go examples
- **FIX**: Clarify pseudocode notation

### v1.0.0 (2025-06-01)
- **INITIAL**: First production release
```

---

## Resources

**Official Claude Skills Documentation**:
- Claude Code Skills guide (official Anthropic docs)

**Skill Factory Components**:
- [METADATA.md](METADATA.md) — Metadata authoring
- [STRUCTURE.md](STRUCTURE.md) — File organization
- [EXAMPLES.md](EXAMPLES.md) — Real-world case studies
- [CHECKLIST.md](CHECKLIST.md) — Quality validation
- [WEB-RESEARCH.md](WEB-RESEARCH.md) — Research strategies

**External Resources**:
- Semantic Versioning: https://semver.org/
- YAML Specification: https://yaml.org/spec/
- Markdown Guide: https://www.markdownguide.org/

---

**Last Updated**: 2025-10-22
**Version**: 2.0.0
**Maintained by**: MoAI-ADK Foundation Team
