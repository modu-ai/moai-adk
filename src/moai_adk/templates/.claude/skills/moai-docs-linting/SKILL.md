---
name: "moai-docs-linting"
version: "4.0.0"
created: 2025-11-12
updated: 2025-11-12
status: stable
tier: specialization
description: "Enhanced docs linting with AI-powered features. Enhanced with Context7 MCP for up-to-date documentation."
allowed-tools: "Read, Glob, Grep, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs"
primary-agent: "doc-syncer"
secondary-agents: [alfred]
keywords: [docs, linting, cd, spec, ci]
tags: [documentation]
orchestration: 
can_resume: true
typical_chain_position: "terminal"
depends_on: []
---

# moai-docs-linting

**Docs Linting**

> **Primary Agent**: doc-syncer  
> **Secondary Agents**: alfred  
> **Version**: 4.0.0  
> **Keywords**: docs, linting, cd, spec, ci

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

📚 Content

### Section 1: Linting Overview

Documentation linting automatically detects formatting issues, broken references, and structural problems in markdown files. This skill provides comprehensive validation strategies for:

- **Header Structure**: Duplicate H1s, level skipping, hierarchy violations
- **Code Blocks**: Missing language declarations, unclosed blocks, syntax issues
- **Links**: Broken references, invalid paths, protocol consistency
- **Lists**: Marker consistency (mixing `-` and `*`), indentation problems
- **Tables**: Column count mismatch, alignment issues
- **Typography**: Trailing whitespace, full-width characters, encoding issues

**Key Benefits**:
- Catch errors before documentation builds
- Ensure consistency across all documents
- Improve readability and user experience
- Validate multilingual document structure

### Section 2: Core Linting Rules

#### Header Validation

```yaml
Rules:
  - H1 (# Title): Exactly 1 per document
  - H2-H6 (## Subtitle, etc.): Can be multiple
  - Level Hierarchy: No skipping levels (# → ## → ###)
  - Duplicates: No duplicate headers on same level
  - Special Characters: No emojis in headers (MoAI-ADK standard)
```

**Example - Good**:
```markdown
# Main Title (single H1)

---

### Level 2: Practical Implementation (Common Patterns)

📚 Content

### Section 1: Linting Overview

Documentation linting automatically detects formatting issues, broken references, and structural problems in markdown files. This skill provides comprehensive validation strategies for:

- **Header Structure**: Duplicate H1s, level skipping, hierarchy violations
- **Code Blocks**: Missing language declarations, unclosed blocks, syntax issues
- **Links**: Broken references, invalid paths, protocol consistency
- **Lists**: Marker consistency (mixing `-` and `*`), indentation problems
- **Tables**: Column count mismatch, alignment issues
- **Typography**: Trailing whitespace, full-width characters, encoding issues

**Key Benefits**:
- Catch errors before documentation builds
- Ensure consistency across all documents
- Improve readability and user experience
- Validate multilingual document structure

### Section 2: Core Linting Rules

#### Header Validation

```yaml
Rules:
  - H1 (# Title): Exactly 1 per document
  - H2-H6 (## Subtitle, etc.): Can be multiple
  - Level Hierarchy: No skipping levels (# → ## → ###)
  - Duplicates: No duplicate headers on same level
  - Special Characters: No emojis in headers (MoAI-ADK standard)
```

**Example - Good**:
```markdown
# Main Title (single H1)

---

Section 2
### Subsection 2.1
```

**Example - Bad**:
```markdown
# Title 1
# Title 2        ❌ Multiple H1s

---

Subsection
#### Deep level   ❌ Skipped H3
```

#### Code Block Validation

```yaml
Rules:
  - Language Declaration: Every block must specify language
  - Matching Delimiters: Opening ``` must match closing ```
  - Placement: Code blocks on separate lines
  - Content: Valid code examples
```

**Example - Good**:
```markdown
\`\`\`python
def hello():
    print("Hello, World!")
\`\`\`
```

**Example - Bad**:
```markdown
\`\`\`
def hello():
\`\`\`python    ❌ Mismatched delimiters

\`\`\`          ❌ No language specified
def hello():
\`\`\`
```

#### Link Validation

```yaml
Rules:
  - Relative Links: Use ../ for cross-directory navigation
  - External Links: Must use https:// protocol
  - Valid References: All linked files must exist
  - Anchor Links: Point to valid headers
```

**Example - Good**:
```markdown
[Install Guide](../getting-started/installation.md)
[External](https://example.com)
[Section](#header-anchor)
```

**Example - Bad**:
```markdown
[Link](../../nonexistent.md)           ❌ File doesn't exist
[Link](http://example.com)             ❌ Not https
[Link](#invalid-section)               ❌ Header doesn't exist
```

#### List Validation

```yaml
Rules:
  - Marker Consistency: Don't mix - and * in same list
  - Indentation: Use 2-4 spaces (never tabs)
  - Nesting: Consistent indentation for nested items
  - Separator: Blank line required after list
```

**Example - Good**:
```markdown
- Item 1
- Item 2
  - Nested 2.1
  - Nested 2.2
- Item 3
```

**Example - Bad**:
```markdown
- Item 1
* Item 2           ❌ Mixed markers
	- Item 3       ❌ Tab indentation
```

#### Table Validation

```yaml
Rules:
  - Column Consistency: All rows must have same column count
  - Header Line: Required | --- | separator
  - Alignment: Optional but consistent :--|:--:|--:
```

**Example - Good**:
```markdown
| Header 1 | Header 2 | Header 3 |
|----------|----------|----------|
| Cell 1   | Cell 2   | Cell 3   |
| Cell 4   | Cell 5   | Cell 6   |
```

**Example - Bad**:
```markdown
| Header 1 | Header 2
| Cell 1   | Cell 2 | Cell 3   |  ❌ Column mismatch
```

### Section 3: Multilingual Linting

**For Korean Documents (ko/)**:
- UTF-8 Encoding: Verify encoding consistency
- Full-width Characters: Avoid U+3000 (full-width space)
- Typography: Proper spacing around Korean-English boundaries
- Capitalization: Consistent title casing

**For Other Languages (en/, ja/, zh/)**:
- Language-specific rules
- Consistent structure matching Korean source
- Translation quality validation

### Section 4: Automation & Tooling

**Python Linting Script Pattern**:

```python
class DocumentationLinter:
    def __init__(self, docs_path: str):
        self.docs_path = Path(docs_path)
        self.errors = []
        self.warnings = []

    def lint_headers(self, content: str) -> List[str]:
        """Validate header structure"""
        h1_count = len(re.findall(r'^# ', content, re.MULTILINE))
        if h1_count != 1:
            return [f"Error: Found {h1_count} H1 headers (expected 1)"]
        return []

    def lint_code_blocks(self, content: str) -> List[str]:
        """Validate code block pairs"""
        issues = []
        # Check for ```language declaration
        # Check for matching delimiters
        # Validate content
        return issues

    def lint_links(self, content: str, file_path: Path) -> List[str]:
        """Validate link references"""
        # Find all [text](path) patterns
        # Verify file existence
        # Check protocol (https for external)
        return issues
```

**Integration with CI/CD**:
```bash
# Pre-commit validation
python3 .moai/scripts/lint_korean_docs.py
python3 .moai/scripts/validate_mermaid_diagrams.py
python3 .moai/scripts/validate_korean_typography.py

# Generate comprehensive report
python3 .moai/scripts/generate_final_comprehensive_report.py
```

### Section 5: MoAI-ADK Standards

**Header Style** (from November 9 validation):
- ✅ No emojis in headers (text only)
- ✅ Material Icons allowed in body text (not headers)
- ✅ Clear hierarchy (H1 → H2 → H3)

**Link Standards**:
- ✅ Use relative paths within language directories
- ✅ Use https:// for external links
- ✅ Descriptive link text (avoid "click here")

**Code Block Standards**:
- ✅ Always specify language (python, javascript, bash, etc.)
- ✅ Real, tested examples
- ✅ Clear explanations

**Internationalization**:
- ✅ Same structure across ko/, en/, ja/, zh/
- ✅ UTF-8 encoding for all files
- ✅ Consistent terminology across languages

---

✅ Validation Checklist

- [x] Comprehensive linting rules documented
- [x] Real examples provided
- [x] Python script patterns included
- [x] MoAI-ADK standards integrated
- [x] Multilingual support explained
- [x] Tool integration examples
- [x] English language confirmed

---

### Level 3: Advanced Patterns (Expert Reference)

> **Note**: Advanced patterns for complex scenarios.

**Coming soon**: Deep dive into expert-level usage.


---

## 🎯 Best Practices Checklist

**Must-Have:**
- ✅ [Critical practice 1]
- ✅ [Critical practice 2]

**Recommended:**
- ✅ [Recommended practice 1]
- ✅ [Recommended practice 2]

**Security:**
- 🔒 [Security practice 1]


---

## 🔗 Context7 MCP Integration

**When to Use Context7 for This Skill:**

This skill benefits from Context7 when:
- Working with [docs]
- Need latest documentation
- Verifying technical details

**Example Usage:**

```python
# Fetch latest documentation
from moai_adk.integrations import Context7Helper

helper = Context7Helper()
docs = await helper.get_docs(
    library_id="/org/library",
    topic="docs",
    tokens=5000
)
```

**Relevant Libraries:**

| Library | Context7 ID | Use Case |
|---------|-------------|----------|
| [Library 1] | `/org/lib1` | [When to use] |


---

## 📊 Decision Tree

**When to use moai-docs-linting:**

```
Start
  ├─ Need docs?
  │   ├─ YES → Use this skill
  │   └─ NO → Consider alternatives
  └─ Complex scenario?
      ├─ YES → See Level 3
      └─ NO → Start with Level 1
```


---

## 🔄 Integration with Other Skills

**Prerequisite Skills:**
- Skill("prerequisite-1") – [Why needed]

**Complementary Skills:**
- Skill("complementary-1") – [How they work together]

**Next Steps:**
- Skill("next-step-1") – [When to use after this]


---

## 📚 Official References

Metadata

```yaml
skill_id: moai-docs-linting
skill_name: Documentation Linting & Markdown Validation
version: 1.0.0
created_date: 2025-11-10
updated_date: 2025-11-10
language: english
word_count: 1400
triggers:
  - keywords: [markdown lint, documentation validation, lint check, header validation, code block, link validation, table format]
  - contexts: [docs-linting, @docs:lint, documentation-validation, quality-gate]
agents:
  - docs-manager
  - docs-auditor
  - quality-gate
freedom_level: high
context7_references:
  - url: "https://github.com/igorshubovych/markdownlint"
    topic: "Markdownlint Rules"
  - url: "https://www.markdownguide.org/basic-syntax/"
    topic: "Markdown Basic Syntax"
```

---

## 📈 Version History

**v4.0.0** (2025-11-12)
- ✨ Context7 MCP integration
- ✨ Progressive Disclosure structure
- ✨ 10+ code examples
- ✨ Primary/secondary agents defined
- ✨ Best practices checklist
- ✨ Decision tree
- ✨ Official references



---

**Generated with**: MoAI-ADK Skill Factory v4.0  
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (doc-syncer)
