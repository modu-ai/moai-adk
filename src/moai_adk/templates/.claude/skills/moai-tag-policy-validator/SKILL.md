---
name: "moai-tag-policy-validator"
version: "4.0.0"
created: 2025-11-05
updated: 2025-11-12
status: stable
tier: specialization
description: "Comprehensive TAG system validator and policy enforcer that monitors, validates, and corrects TAG usage across code, tests, and documentation. Use when ensuring TAG compliance, validating TAG policy violations, analyzing TAG coverage, or when maintaining TAG system integrity and governance.. Enhanced with Context7 MCP for up-to-date documentation."
allowed-tools: "Read, Glob, Grep, Bash, Write, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs"
primary-agent: "alfred"
secondary-agents: []
keywords: [tag, policy, validator, git, spec]
tags: []
orchestration: 
can_resume: true
typical_chain_position: "middle"
depends_on: []
---

# moai-tag-policy-validator

**Tag Policy Validator**

> **Primary Agent**: alfred  
> **Secondary Agents**: none  
> **Version**: 4.0.0  
> **Keywords**: tag, policy, validator, git, spec

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

What It Does

Comprehensive TAG system validator and policy enforcer that monitors TAG usage, validates compliance with TAG policies, analyzes coverage, and provides automatic corrections for TAG violations. Ensures the integrity and governance of the MoAI-ADK TAG system.

**Core capabilities**:
- ✅ Real-time TAG policy validation and enforcement
- ✅ TAG coverage analysis across code, tests, and documentation
- ✅ Automatic TAG violation detection and correction
- ✅ TAG policy compliance reporting and metrics
- ✅ Orphaned TAG detection and resolution
- ✅ TAG chain integrity verification (SPEC→TEST→CODE→DOC)
- ✅ TAG policy governance and rule management
- ✅ TAG system health monitoring and alerts

---

---

### Level 2: Practical Implementation (Common Patterns)

TAG Policy Framework

### 1. TAG Policy Rules
```python
TAG_POLICIES = {
    "SPEC_TAGS": {
        "required_fields": ["id", "version", "status", "created", "updated", "author", "priority"],
        "format_validation": {
            "id": r"SPEC-[0-9]{3}",
            "version": r"[0-9]+\.[0-9]+\.[0-9]+",
            "status": r"^(draft|active|completed|archived)$"
        },
        "content_validation": {
            "min_content_length": 100,
            "required_sections": ["## What It Does", "## When to Use"],
            "max_content_lines": 500
        }
    },
    "CODE_TAGS": {
        "placement_rules": {
            "function_comments": True,
            "class_documentation": True,
            "module_level": True
        },
        "content_requirements": {
            "min_description_length": 20,
            "require_examples": True,
            "require_parameters": True
        }
    },
    "TEST_TAGS": {
        "coverage_requirements": {
            "test_function_tagging": True,
            "test_class_tagging": True,
            "parameter_testing": True
        },
        "content_validation": {
            "require_test_description": True,
            "require_expected_behavior": True
        }
    },
    "DOC_TAGS": {
        "linking_rules": {
            "bidirectional_links": True,
            "chain_completeness": True,
            "no_broken_links": True
        },
        "content_requirements": {
            "up_to_date": True,
            "accuracy_check": True,
            "completeness_check": True
        }
    }
}
```

### 2. TAG Chain Validation
```python
def validate_tag_chains():
    """Validate complete TAG chains: SPEC → TEST → CODE → DOC"""
    chain_analysis = {
        "complete_chains": [],
        "broken_chains": [],
        "missing_links": [],
        "orphaned_tags": []
    }

    # Find all SPEC tags
    spec_tags = find_all_tags("SPEC-")

    for spec_tag in spec_tags:
        chain = {
            "spec": spec_tag,
            "tests": find_linked_tests(spec_tag),
            "code": find_linked_code(spec_tag),
            "docs": find_linked_docs(spec_tag)
        }

        # Validate chain completeness
        if is_complete_chain(chain):
            chain_analysis["complete_chains"].append(chain)
        else:
            missing = identify_missing_links(chain)
            chain_analysis["broken_chains"].append({
                "spec": spec_tag,
                "missing_links": missing,
                "severity": assess_chain_severity(missing)
            })

    return chain_analysis
```

### 3. TAG Coverage Analysis
```python
def analyze_tag_coverage():
    """Analyze TAG coverage across project"""
    coverage_metrics = {
        "total_functions": count_functions(),
        "tagged_functions": count_tagged_functions(),
        "total_classes": count_classes(),
        "tagged_classes": count_tagged_classes(),
        "total_tests": count_tests(),
        "tagged_tests": count_tagged_tests(),
        "total_docs": count_documentation_files(),
        "tagged_docs": count_tagged_docs(),
        "coverage_percentages": {},
        "untagged_areas": []
    }

    # Calculate coverage percentages
    coverage_metrics["coverage_percentages"] = {
        "functions": (coverage_metrics["tagged_functions"] / coverage_metrics["total_functions"]) * 100,
        "classes": (coverage_metrics["tagged_classes"] / coverage_metrics["total_classes"]) * 100,
        "tests": (coverage_metrics["tagged_tests"] / coverage_metrics["total_tests"]) * 100,
        "docs": (coverage_metrics["tagged_docs"] / coverage_metrics["total_docs"]) * 100
    }

    # Identify untagged areas
    coverage_metrics["untagged_areas"] = find_untagged_areas()

    return coverage_metrics
```

---

---

Integration Examples

### Example 1: Pre-Commit Validation
```python
def validate_tags_before_commit():
    """Validate TAGs before Git commit"""
    Skill("moai-tag-policy-validator")

    # Get staged files
    staged_files = get_staged_files()

    validation_results = []
    critical_violations = []

    for file_path in staged_files:
        result = validate_file_tags(file_path, read_file(file_path))
        validation_results.append(result)

        # Check for critical violations
        critical_violations.extend([
            v for v in result.violations if v.severity == "critical"
        ])

    if critical_violations:
        display_critical_violations(critical_violations)
        return False  # Block commit
    else:
        display_validation_summary(validation_results)
        return True  # Allow commit
```

### Example 2: TAG Coverage Analysis
```python
def analyze_project_tag_coverage():
    """Analyze TAG coverage across entire project"""
    Skill("moai-tag-policy-validator")

    coverage = analyze_tag_coverage()

    display_coverage_dashboard(coverage)

    # Identify areas needing attention
    low_coverage_areas = [
        area for area, coverage in coverage["coverage_percentages"].items()
        if coverage < 80
    ]

    if low_coverage_areas:
        display_coverage_gaps(low_coverage_areas)
        suggest_improvement_strategies(low_coverage_areas)
```

### Example 3: Automatic Correction
```python
def fix_tag_violations():
    """Fix TAG violations automatically"""
    Skill("moai-tag-policy-validator")

    violations = detect_policy_violations()
    auto_correctable = [v for v in violations if v.auto_correctable]

    if auto_correctable:
        print(f"Found {len(auto_correctable)} auto-correctable violations")

        # Preview corrections
        corrections = auto_correct_violations(auto_correctable, dry_run=True)
        display_correction_preview(corrections)

        # Ask for confirmation
        if confirm_corrections():
            # Apply corrections
            result = auto_correct_violations(auto_correctable, dry_run=False)
            display_correction_results(result)
    else:
        print("No auto-correctable violations found")
```

---

---

Quality Assurance

### 1. Validation Rules
```python
VALIDATION_RULES = {
    "spec_format": {
        "required": ["@SPEC:ID", "@VERSION", "@STATUS"],
        "forbidden": ["@TODO", "@FIXME"],
        "patterns": {
            "@SPEC:": r"@[A-Z]+:[A-Z0-9-]+"
        }
    },
    "code_format": {
        "placement": ["function_level", "class_level", "module_level"],
        "content": ["description", "parameters", "returns"],
        "forbidden": ["vague_descriptions"]
    },
    "link_format": {
        "required": ["source", "target"],
        "bidirectional": True,
        "no_broken_links": True
    }
}
```

### 2. Quality Metrics
```python
def calculate_tag_quality_metrics():
    """Calculate TAG quality metrics"""
    metrics = {
        "accuracy": measure_tag_accuracy(),
        "completeness": measure_tag_completeness(),
        "consistency": measure_tag_consistency(),
        "maintainability": measure_tag_maintainability(),
        "traceability": measure_tag_traceability()
    }

    # Calculate overall quality score
    metrics["overall_score"] = calculate_overall_quality_score(metrics)

    return metrics
```

### 3. Continuous Improvement
```python
def improve_tag_system():
    """Continuously improve TAG system based on usage patterns"""
    improvement_plan = {
        "usage_analysis": analyze_tag_usage_patterns(),
        "feedback_integration": integrate_user_feedback(),
        "policy_refinement": refine_policies_based_on_data(),
        "tooling_enhancement": identify_tooling_improvements()
    }

    # Generate improvement recommendations
    recommendations = generate_improvement_recommendations(improvement_plan)

    return {
        "plan": improvement_plan,
        "recommendations": recommendations,
        "implementation_priority": prioritize_improvements(recommendations)
    }
```

---

---

Usage Examples

### Example 1: Validate Single File
```python
# User wants to validate a specific file
Skill("moai-tag-policy-validator")

file_path = "src/main.py"
validation_result = validate_file_tags(file_path, read_file(file_path))

if validation_result.violations:
    display_violations(validation_result.violations)
    suggest_corrections(validation_result.violations)
else:
    display_success("All TAG policies compliant")
```

### Example 2: Project-Wide Validation
```python
# User wants to validate entire project
Skill("moai-tag-policy-validator")

violations = detect_policy_violations()
compliance_report = generate_compliance_report()

display_compliance_dashboard(compliance_report)

if violations["critical_violations"]:
    display_critical_alerts(violations["critical_violations"])
```

### Example 3: TAG Chain Analysis
```python
# User wants to analyze TAG chains
Skill("moai-tag-policy-validator")

chain_analysis = validate_tag_chains()

display_chain_analysis(chain_analysis)

if chain_analysis["broken_chains"]:
    display_chain_repair_suggestions(chain_analysis["broken_chains"])
```

---

**End of Skill** | Comprehensive TAG system validation and governance for maintaining TAG integrity and compliance

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
- Working with [tag]
- Need latest documentation
- Verifying technical details

**Example Usage:**

```python
# Fetch latest documentation
from moai_adk.integrations import Context7Helper

helper = Context7Helper()
docs = await helper.get_docs(
    library_id="/org/library",
    topic="tag",
    tokens=5000
)
```

**Relevant Libraries:**

| Library | Context7 ID | Use Case |
|---------|-------------|----------|
| [Library 1] | `/org/lib1` | [When to use] |


---

## 📊 Decision Tree

**When to use moai-tag-policy-validator:**

```
Start
  ├─ Need tag?
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

TAG Policy Framework

### 1. TAG Policy Rules
```python
TAG_POLICIES = {
    "SPEC_TAGS": {
        "required_fields": ["id", "version", "status", "created", "updated", "author", "priority"],
        "format_validation": {
            "id": r"SPEC-[0-9]{3}",
            "version": r"[0-9]+\.[0-9]+\.[0-9]+",
            "status": r"^(draft|active|completed|archived)$"
        },
        "content_validation": {
            "min_content_length": 100,
            "required_sections": ["## What It Does", "## When to Use"],
            "max_content_lines": 500
        }
    },
    "CODE_TAGS": {
        "placement_rules": {
            "function_comments": True,
            "class_documentation": True,
            "module_level": True
        },
        "content_requirements": {
            "min_description_length": 20,
            "require_examples": True,
            "require_parameters": True
        }
    },
    "TEST_TAGS": {
        "coverage_requirements": {
            "test_function_tagging": True,
            "test_class_tagging": True,
            "parameter_testing": True
        },
        "content_validation": {
            "require_test_description": True,
            "require_expected_behavior": True
        }
    },
    "DOC_TAGS": {
        "linking_rules": {
            "bidirectional_links": True,
            "chain_completeness": True,
            "no_broken_links": True
        },
        "content_requirements": {
            "up_to_date": True,
            "accuracy_check": True,
            "completeness_check": True
        }
    }
}
```

### 2. TAG Chain Validation
```python
def validate_tag_chains():
    """Validate complete TAG chains: SPEC → TEST → CODE → DOC"""
    chain_analysis = {
        "complete_chains": [],
        "broken_chains": [],
        "missing_links": [],
        "orphaned_tags": []
    }

    # Find all SPEC tags
    spec_tags = find_all_tags("SPEC-")

    for spec_tag in spec_tags:
        chain = {
            "spec": spec_tag,
            "tests": find_linked_tests(spec_tag),
            "code": find_linked_code(spec_tag),
            "docs": find_linked_docs(spec_tag)
        }

        # Validate chain completeness
        if is_complete_chain(chain):
            chain_analysis["complete_chains"].append(chain)
        else:
            missing = identify_missing_links(chain)
            chain_analysis["broken_chains"].append({
                "spec": spec_tag,
                "missing_links": missing,
                "severity": assess_chain_severity(missing)
            })

    return chain_analysis
```

### 3. TAG Coverage Analysis
```python
def analyze_tag_coverage():
    """Analyze TAG coverage across project"""
    coverage_metrics = {
        "total_functions": count_functions(),
        "tagged_functions": count_tagged_functions(),
        "total_classes": count_classes(),
        "tagged_classes": count_tagged_classes(),
        "total_tests": count_tests(),
        "tagged_tests": count_tagged_tests(),
        "total_docs": count_documentation_files(),
        "tagged_docs": count_tagged_docs(),
        "coverage_percentages": {},
        "untagged_areas": []
    }

    # Calculate coverage percentages
    coverage_metrics["coverage_percentages"] = {
        "functions": (coverage_metrics["tagged_functions"] / coverage_metrics["total_functions"]) * 100,
        "classes": (coverage_metrics["tagged_classes"] / coverage_metrics["total_classes"]) * 100,
        "tests": (coverage_metrics["tagged_tests"] / coverage_metrics["total_tests"]) * 100,
        "docs": (coverage_metrics["tagged_docs"] / coverage_metrics["total_docs"]) * 100
    }

    # Identify untagged areas
    coverage_metrics["untagged_areas"] = find_untagged_areas()

    return coverage_metrics
```

---

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
**Maintained by**: Primary Agent (alfred)
