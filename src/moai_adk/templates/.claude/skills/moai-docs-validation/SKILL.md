---
name: "moai-docs-validation"
version: "4.0.0"
created: 2025-11-12
updated: 2025-11-12
status: stable
tier: specialization
description: "Enhanced docs validation with AI-powered features. Enhanced with Context7 MCP for up-to-date documentation."
allowed-tools: "Read, Glob, Grep, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs"
primary-agent: "doc-syncer"
secondary-agents: [alfred]
keywords: [docs, validation, auth, cd, test]
tags: [documentation]
orchestration: 
can_resume: true
typical_chain_position: "terminal"
depends_on: []
---

# moai-docs-validation

**Docs Validation**

> **Primary Agent**: doc-syncer  
> **Secondary Agents**: alfred  
> **Version**: 4.0.0  
> **Keywords**: docs, validation, auth, cd, test

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

📚 Content

### Section 1: Validation Framework Overview

Documentation validation ensures content accuracy, completeness, and compliance with MoAI-ADK standards. This skill covers comprehensive validation strategies for:

- **SPEC Compliance**: Verify documentation references valid SPECs and TAG chains
- **Content Accuracy**: Validate code examples, API signatures, configuration patterns
- **Completeness Checking**: Ensure all required sections present and filled
- **Quality Metrics**: Measure readability, coverage, translation quality
- **Multilingual Consistency**: Verify structure and content alignment across languages

**Key Benefits**:
- Catch inaccurate documentation before publication
- Ensure SPEC-documentation traceability
- Maintain quality standards across all documents
- Automate quality gate enforcement
- Enable data-driven documentation improvements

### Section 2: Validation Rules & Standards

#### SPEC Compliance Validation

```yaml
Rules:
  - Requirement Coverage: All SPEC requirements addressed in documentation
```

**Example - Good**:
```markdown
# Feature: User Authentication


---

### Level 2: Practical Implementation (Common Patterns)

📚 Content

### Section 1: Validation Framework Overview

Documentation validation ensures content accuracy, completeness, and compliance with MoAI-ADK standards. This skill covers comprehensive validation strategies for:

- **SPEC Compliance**: Verify documentation references valid SPECs and TAG chains
- **Content Accuracy**: Validate code examples, API signatures, configuration patterns
- **Completeness Checking**: Ensure all required sections present and filled
- **Quality Metrics**: Measure readability, coverage, translation quality
- **Multilingual Consistency**: Verify structure and content alignment across languages

**Key Benefits**:
- Catch inaccurate documentation before publication
- Ensure SPEC-documentation traceability
- Maintain quality standards across all documents
- Automate quality gate enforcement
- Enable data-driven documentation improvements

### Section 2: Validation Rules & Standards

#### SPEC Compliance Validation

```yaml
Rules:
  - Requirement Coverage: All SPEC requirements addressed in documentation
```

**Example - Good**:
```markdown
# Feature: User Authentication


---

Implementation

```

**Example - Bad**:
```markdown
# User Authentication

---

How to Authenticate
[No SPEC reference]
[No TAG chain]
[No testing guidance]
```

#### Content Accuracy Validation

```yaml
Rules:
  - Code Examples: Tested, executable, syntax-correct
  - API Signatures: Match actual implementation
  - Parameter Types: Accurate type annotations
  - Return Values: Documented behavior matches actual behavior
  - Configuration: All config examples work in practice
```

**Validation Pattern**:
```python
def validate_code_examples(self, doc_path: Path) -> List[ValidationError]:
    """Verify code examples are syntactically correct"""
    errors = []

    # Extract all ```language code blocks
    code_blocks = self._extract_code_blocks(doc_path)

    for block in code_blocks:
        # Verify syntax
        syntax_errors = self._check_syntax(block.code, block.language)
        if syntax_errors:
            errors.append(ValidationError(
                file=doc_path,
                line=block.line_number,
                message=f"Invalid {block.language} syntax",
                details=syntax_errors
            ))

    return errors
```

#### Quality Metrics Validation

```yaml
Metrics:
  - Readability Score: 60-100 (Flesch-Kincaid readability)
  - Coverage: 80%+ of specification requirements
  - Code Example Ratio: 1 example per 300 words (target)
  - Link Validity: 100% of internal/external links valid
  - Translation Completeness: 100% structure alignment across languages
  - Image Optimization: All images <500KB, proper alt text
```

**Quality Score Calculation**:
```python
def calculate_quality_score(self, doc_path: Path) -> float:
    """Calculate documentation quality (0-100)"""
    scores = {
        'spec_compliance': self._check_spec_compliance(doc_path),      # 25%
        'content_accuracy': self._validate_content_accuracy(doc_path),  # 25%
        'completeness': self._check_completeness(doc_path),             # 20%
        'readability': self._calculate_readability(doc_path),           # 15%
        'formatting': self._check_formatting(doc_path),                 # 15%
    }

    weights = {
        'spec_compliance': 0.25,
        'content_accuracy': 0.25,
        'completeness': 0.20,
        'readability': 0.15,
        'formatting': 0.15,
    }

    total_score = sum(scores[k] * weights[k] for k in scores)
    return round(total_score, 1)
```

### Section 3: TAG Verification System

**TAG Chain Validation**:

```yaml
Valid Chains:

Chain Rules:
  - No broken links (referenced TAGs must exist)
```

**Verification Script Pattern**:
```python
class TAGVerifier:
    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.spec_docs = {}
        self.test_docs = {}
        self.code_refs = {}
        self.doc_refs = {}

    def verify_chain(self, spec_id: str) -> ValidationResult:
        result = ValidationResult(spec_id)

        # Check SPEC exists
        if spec_id not in self.spec_docs:
            return result

        # Check TEST exists
        if spec_id not in self.test_docs:

        # Check CODE references
        if spec_id not in self.code_refs:

        # Check DOC exists
        if spec_id not in self.doc_refs:

        return result
```

### Section 4: Multilingual Validation

**Translation Consistency Checks**:

```python
class MultilingualValidator:
    def validate_structure_consistency(self) -> List[str]:
        """Ensure all languages have same document structure"""
        issues = []

        # Get Korean file structure (source)
        ko_structure = self._get_file_structure("docs/src/ko")

        # Compare with other languages
        for lang in ["en", "ja", "zh"]:
            lang_structure = self._get_file_structure(f"docs/src/{lang}")

            if ko_structure != lang_structure:
                missing = set(ko_structure) - set(lang_structure)
                extra = set(lang_structure) - set(ko_structure)

                if missing:
                    issues.append(f"[{lang}] Missing files: {missing}")
                if extra:
                    issues.append(f"[{lang}] Extra files: {extra}")

        return issues

    def validate_translation_quality(self, lang: str) -> float:
        """Score translation completeness (0-100)"""
        ko_files = set(self._list_files("docs/src/ko", "*.md"))
        lang_files = set(self._list_files(f"docs/src/{lang}", "*.md"))

        translated = len(ko_files & lang_files)
        translation_ratio = (translated / len(ko_files)) * 100

        return round(translation_ratio, 1)
```

### Section 5: Automation & CI/CD Integration

**GitHub Actions Integration Pattern**:

```bash
# Pre-commit validation
python3 .moai/scripts/validate_docs.py --mode pre-commit

# Pull request validation
python3 .moai/scripts/validate_docs.py --mode pr --files-changed

# Full documentation audit
python3 .moai/scripts/validate_docs.py --mode full --report comprehensive
```

**Quality Gate Configuration**:

```yaml
# .moai/quality-gates.yml
documentation:
  spec_compliance:
    min_score: 90
    required: true
    action: block_merge

  content_accuracy:
    min_score: 85
    required: true
    action: block_merge

  link_validity:
    broken_links_allowed: 0
    required: true
    action: block_merge

  multilingual_consistency:
    max_missing_translations: 0
    required: true
    action: warning
```

**Automated Reports**:

```python
def generate_validation_report(self, output_format: str = "markdown") -> str:
    """Generate comprehensive validation report"""
    report = []

    # Summary
    report.append("# Documentation Validation Report")
    report.append(f"Generated: {datetime.now()}")
    report.append("")

    # Quality Scores
    report.append("## Quality Metrics")
    for doc, score in self.quality_scores.items():
        status = "✅" if score >= 85 else "⚠️" if score >= 70 else "❌"
        report.append(f"{status} {doc}: {score}/100")

    # Issues Summary
    report.append("## Issues Found")
    report.append(f"- Errors: {len(self.errors)}")
    report.append(f"- Warnings: {len(self.warnings)}")

    # Detailed Issues
    for error in self.errors:
        report.append(f"❌ {error.file}:{error.line} - {error.message}")

    return "\n".join(report)
```

---

✅ Validation Checklist

- [x] Comprehensive validation rules documented
- [x] SPEC compliance patterns included
- [x] TAG verification system explained
- [x] Quality metrics calculation patterns provided
- [x] Python script patterns included
- [x] CI/CD integration examples shown
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

**When to use moai-docs-validation:**

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
skill_id: moai-docs-validation
skill_name: Documentation Validation & Quality Assurance
version: 1.0.0
created_date: 2025-11-10
updated_date: 2025-11-10
language: english
word_count: 1400
triggers:
  - keywords: [documentation validation, content verification, quality assurance, spec compliance, tag verification, documentation audit, quality metrics]
  - contexts: [docs-validation, @docs:validate, quality-audit, spec-compliance]
agents:
  - docs-auditor
  - quality-gate
  - spec-builder
freedom_level: high
context7_references:
  - url: "https://en.wikipedia.org/wiki/Software_quality"
    topic: "Software Quality Metrics"
  - url: "https://github.com/moai-adk/moai-adk"
    topic: "MoAI-ADK SPEC Standards"
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
