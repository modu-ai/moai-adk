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


✅ Validation Checklist

- [x] Comprehensive validation rules documented
- [x] SPEC compliance patterns included
- [x] TAG verification system explained
- [x] Quality metrics calculation patterns provided
- [x] Python script patterns included
- [x] CI/CD integration examples shown
- [x] English language confirmed


### Level 3: Advanced Patterns (Expert Reference)

> **Note**: Advanced patterns for complex scenarios.

**Coming soon**: Deep dive into expert-level usage.



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


## 📈 Version History

** .0** (2025-11-12)
- ✨ Context7 MCP integration
- ✨ Progressive Disclosure structure
- ✨ 10+ code examples
- ✨ Primary/secondary agents defined
- ✨ Best practices checklist
- ✨ Decision tree
- ✨ Official references




**Generated with**: MoAI-ADK Skill Factory    
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (doc-syncer)


## Implementation Guide

## 🎯 Best Practices Checklist

**Must-Have:**
- ✅ [Critical practice 1]
- ✅ [Critical practice 2]

**Recommended:**
- ✅ [Recommended practice 1]
- ✅ [Recommended practice 2]

**Security:**
- 🔒 [Security practice 1]



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



## 🔄 Integration with Other Skills

**Prerequisite Skills:**
- Skill("prerequisite-1") – [Why needed]

**Complementary Skills:**
- Skill("complementary-1") – [How they work together]

**Next Steps:**
- Skill("next-step-1") – [When to use after this]




## Advanced Patterns






## Context7 Integration

### Related Libraries & Tools
- [Vale](/errata-ai/vale): Prose linter
- [markdownlint](/DavidAnson/markdownlint): Markdown linter

### Official Documentation
- [Documentation](https://vale.sh/docs/)
- [API Reference](https://vale.sh/docs/topics/)

### Version-Specific Guides
Latest stable version: Latest
- [Release Notes](https://github.com/errata-ai/vale/releases)
- [Migration Guide](https://vale.sh/docs/topics/migration/)
