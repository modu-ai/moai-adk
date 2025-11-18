---
name: moai-docs-generation
version: 4.0.0
created: 2025-11-12
updated: '2025-11-18'
status: stable
tier: specialization
description: Enhanced docs generation with AI-powered features. Enhanced with Context7
  MCP for up-to-date documentation.
allowed-tools: Read, Glob, Grep, WebSearch, WebFetch, mcp__context7__resolve-library-id,
  mcp__context7__get-library-docs
primary-agent: doc-syncer
secondary-agents:
- alfred
keywords:
- docs
- generation
- test
- api
- spec
tags:
- documentation
orchestration: null
can_resume: true
typical_chain_position: terminal
depends_on: []
stability: stable
---


# moai-docs-generation

**Docs Generation**

> **Primary Agent**: doc-syncer  
> **Secondary Agents**: alfred  
> **Version**: 4.0.0  
> **Keywords**: docs, generation, test, api, spec

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

Overview
Brief description of what this guide covers.

---

### Level 2: Practical Implementation (Common Patterns)

Metadata

```yaml
skill_id: moai-docs-generation
skill_name: Documentation Generation & Template Management
version: 1.0.0
created_date: 2025-11-10
updated_date: 2025-11-10
language: english
word_count: 1400
triggers:
  - keywords: [documentation generation, doc template, scaffold, generate docs, api documentation, readme generation]
  - contexts: [docs-generation, @docs:generate, documentation-template, doc-scaffold]
agents:
  - docs-manager
  - spec-builder
  - frontend-expert
  - backend-expert
freedom_level: high
context7_references:
  - url: "https://www.typescriptlang.org/docs/handbook/"
    topic: "TypeScript Documentation Pattern"
  - url: "https://github.com/prettier/prettier"
    topic: "Code Formatting Standards"
```

---

Step-by-Step Tutorial
### Step 1: [Action]
Detailed explanation...

```code-example```

### Step 2: [Next Action]
...

---

Usage

### Method: [methodName]

**Signature**:
\`\`\`typescript
function methodName(param1: Type1, param2: Type2): ReturnType
\`\`\`

**Parameters**:
| Name | Type | Description |
|------|------|-------------|
| param1 | Type1 | What it does |

**Returns**: Description of return value

**Example**:
\`\`\`typescript
const result = methodName(arg1, arg2);
\`\`\`

**Throws**: Possible exceptions

---

Examples

### Example 1: Basic Usage
...

### Example 2: Advanced Usage
...

---

Core Concepts

### Concept 1: [Name]
Explanation with examples.

### Concept 2: [Name]
Explanation with examples.

---

License
[License Type](LICENSE)
```

### Section 3: Scaffold Generation

**Directory Structure Generation**:

```python
class DocumentationScaffold:
    def __init__(self, project_name: str):
        self.project_name = project_name

    def create_guide_structure(self, guide_name: str) -> None:
        """Create guide directory and template files"""
        guide_dir = Path(f"docs/guides/{guide_name}")
        guide_dir.mkdir(parents=True, exist_ok=True)

        # Create index.md with guide template
        index_path = guide_dir / "index.md"
        index_path.write_text(GUIDE_TEMPLATE)

        # Create subdirectories
        (guide_dir / "examples").mkdir(exist_ok=True)
        (guide_dir / "images").mkdir(exist_ok=True)
        (guide_dir / "code-samples").mkdir(exist_ok=True)

    def create_api_docs(self, module_name: str) -> None:
        """Generate API documentation structure"""
        api_dir = Path(f"docs/api/{module_name}")
        api_dir.mkdir(parents=True, exist_ok=True)

        # Create main API doc
        api_path = api_dir / "index.md"
        api_path.write_text(API_TEMPLATE)

    def create_multilingual_structure(self, doc_name: str) -> None:
        """Create docs in ko/, en/, ja/, zh/"""
        for lang in ["ko", "en", "ja", "zh"]:
            doc_dir = Path(f"docs/src/{lang}/{doc_name}")
            doc_dir.mkdir(parents=True, exist_ok=True)

            doc_path = doc_dir / "index.md"
            doc_path.write_text(self._get_template_for_lang(lang))
```

### Section 4: Auto-Documentation from Code

**TypeScript/JavaScript**:

```typescript
/**
 * Calculate sum of two numbers
 * @param a First number
 * @param b Second number
 * @returns Sum of a and b
 * @example
 * const result = sum(2, 3);  // Returns 5
 */
function sum(a: number, b: number): number {
    return a + b;
}
```

Generate documentation:
```markdown
### Function: sum

Calculate sum of two numbers

**Signature**:
```typescript
function sum(a: number, b: number): number
```

**Parameters**:
- `a`: First number
- `b`: Second number

**Returns**: Sum of a and b

**Example**:
```typescript
const result = sum(2, 3);  // Returns 5
```
```

**Python**:

```python
def calculate_mean(numbers: List[float]) -> float:
    """
    Calculate arithmetic mean of numbers.

    Args:
        numbers: List of numerical values

    Returns:
        Arithmetic mean of the values

    Raises:
        ValueError: If list is empty

    Example:
        >>> calculate_mean([1, 2, 3])
        2.0
    """
    if not numbers:
        raise ValueError("Cannot calculate mean of empty list")
    return sum(numbers) / len(numbers)
```

### Section 5: Batch Generation

**Generate all project documentation**:

```bash
# Generate README for each module
/docs:generate --type readme --scope all

# Generate API documentation from code
/docs:generate --type api --language typescript

# Create guides for new features
/docs:generate --type guide --feature-spec SPEC-001

# Generate multilingual structure
/docs:generate --type i18n --languages ko,en,ja,zh
```

---

✅ Validation Checklist

- [x] Template library comprehensive
- [x] Scaffold generation patterns included
- [x] Auto-documentation examples provided
- [x] Multilingual support documented
- [x] Code examples included
- [x] Integration patterns shown
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

**When to use moai-docs-generation:**

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
skill_id: moai-docs-generation
skill_name: Documentation Generation & Template Management
version: 1.0.0
created_date: 2025-11-10
updated_date: 2025-11-10
language: english
word_count: 1400
triggers:
  - keywords: [documentation generation, doc template, scaffold, generate docs, api documentation, readme generation]
  - contexts: [docs-generation, @docs:generate, documentation-template, doc-scaffold]
agents:
  - docs-manager
  - spec-builder
  - frontend-expert
  - backend-expert
freedom_level: high
context7_references:
  - url: "https://www.typescriptlang.org/docs/handbook/"
    topic: "TypeScript Documentation Pattern"
  - url: "https://github.com/prettier/prettier"
    topic: "Code Formatting Standards"
```

---

## 📈 Version History

** .0** (2025-11-12)
- ✨ Context7 MCP integration
- ✨ Progressive Disclosure structure
- ✨ 10+ code examples
- ✨ Primary/secondary agents defined
- ✨ Best practices checklist
- ✨ Decision tree
- ✨ Official references



---

**Generated with**: MoAI-ADK Skill Factory    
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (doc-syncer)
