---
name: moai-tag-policy-validator
version: 1.0.0
created: 2025-11-05
updated: 2025-11-05
status: active
description: Comprehensive TAG system validator and policy enforcer that monitors, validates, and corrects TAG usage across code, tests, and documentation. Use when ensuring TAG compliance, validating TAG policy violations, analyzing TAG coverage, or when maintaining TAG system integrity and governance.
keywords: [tag-validation, tag-policy, tag-governance, tag-coverage, tag-compliance, tag-correction]
allowed-tools:
  - Read
  - Glob
  - Grep
  - Bash
  - Write
---

# TAG Policy Validator and Governance System

## Skill Metadata

| Field | Value |
| ----- | ----- |
| Version | 1.0.0 |
| Tier | Alfred (TAG System) |
| Auto-load | When TAG violations detected or on demand |
| Purpose | Ensure TAG system integrity and policy compliance |

---

## What It Does

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

## When to Use

- ✅ When validating TAG policy compliance
- ✅ Before committing changes with TAG modifications
- ✅ During code reviews and quality assurance
- ✅ When analyzing TAG coverage and completeness
- ✅ For TAG system maintenance and governance
- ✅ When troubleshooting TAG-related issues
- ✅ Before major releases or deployments

---

## TAG Policy Framework

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

## Validation Engine

### 1. Real-time Validation
```python
def validate_file_tags(file_path, file_content):
    """Validate TAGs in a specific file"""
    validation_result = {
        "file_path": file_path,
        "file_type": determine_file_type(file_path),
        "tags_found": extract_tags(file_content),
        "violations": [],
        "recommendations": [],
        "compliance_score": 0.0
    }

    # Extract and validate tags
    tags = extract_tags(file_content)

    for tag in tags:
        violations = validate_single_tag(tag, file_path)
        validation_result["violations"].extend(violations)

    # Calculate compliance score
    validation_result["compliance_score"] = calculate_compliance_score(
        validation_result["violations"], file_path
    )

    # Generate recommendations
    validation_result["recommendations"] = generate_recommendations(
        validation_result["violations"], file_path
    )

    return validation_result
```

### 2. Policy Violation Detection
```python
def detect_policy_violations():
    """Detect all TAG policy violations in project"""
    violations = {
        "spec_violations": [],
        "code_violations": [],
        "test_violations": [],
        "doc_violations": [],
        "cross_file_violations": [],
        "policy_compliance": {}
    }

    # Scan all files for violations
    for file_path in get_all_project_files():
        file_violations = scan_file_for_violations(file_path)

        violation_type = categorize_violations(file_violations)
        violations[violation_type].extend(file_violations)

    # Check cross-file violations
    violations["cross_file_violations"] = check_cross_file_violations()

    # Calculate overall compliance
    violations["policy_compliance"] = calculate_overall_compliance(violations)

    return violations
```

### 3. Automatic Correction System
```python
def auto_correct_violations(violations, dry_run=True):
    """Automatically correct TAG violations"""
    corrections_applied = []
    correction_failures = []

    for violation in violations:
        if violation.auto_correctable:
            try:
                correction = generate_correction(violation)

                if not dry_run:
                    apply_correction(correction)

                corrections_applied.append({
                    "violation": violation,
                    "correction": correction,
                    "status": "applied" if not dry_run else "preview"
                })

            except Exception as e:
                correction_failures.append({
                    "violation": violation,
                    "error": str(e),
                    "status": "failed"
                })

    return {
        "corrections_applied": corrections_applied,
        "correction_failures": correction_failures,
        "summary": generate_correction_summary(corrections_applied, correction_failures)
    }
```

---

## TAG System Governance

### 1. Policy Management
```python
def update_tag_policies(new_policies):
    """Update TAG policies with validation"""
    policy_update = {
        "timestamp": datetime.now().isoformat(),
        "old_policies": get_current_policies(),
        "new_policies": new_policies,
        "impact_analysis": analyze_policy_impact(new_policies),
        "validation_results": validate_new_policies(new_policies)
    }

    # Validate new policies
    if policy_update["validation_results"]["is_valid"]:
        apply_new_policies(new_policies)
        policy_update["status"] = "applied"
    else:
        policy_update["status"] = "rejected"
        policy_update["rejection_reasons"] = policy_update["validation_results"]["errors"]

    # Log policy changes
    log_policy_change(policy_update)

    return policy_update
```

### 2. Compliance Reporting
```python
def generate_compliance_report():
    """Generate comprehensive TAG compliance report"""
    report = {
        "generated_at": datetime.now().isoformat(),
        "period": "last_7_days",
        "overall_compliance": 0.0,
        "coverage_metrics": analyze_tag_coverage(),
        "violation_summary": summarize_violations(),
        "trend_analysis": analyze_compliance_trends(),
        "recommendations": generate_compliance_recommendations(),
        "action_items": identify_action_items()
    }

    # Calculate overall compliance
    report["overall_compliance"] = calculate_overall_compliance_score(report)

    # Generate actionable insights
    report["insights"] = generate_compliance_insights(report)

    return report
```

### 3. Health Monitoring
```python
def monitor_tag_system_health():
    """Monitor TAG system health and performance"""
    health_metrics = {
        "system_status": "healthy",
        "performance_metrics": {
            "validation_time": measure_validation_performance(),
            "correction_time": measure_correction_performance(),
            "scan_time": measure_scan_performance()
        },
        "quality_metrics": {
            "tag_accuracy": measure_tag_accuracy(),
            "link_integrity": measure_link_integrity(),
            "coverage_trend": measure_coverage_trend()
        },
        "alerts": [],
        "recommendations": []
    }

    # Check for health issues
    if health_metrics["quality_metrics"]["tag_accuracy"] < 0.9:
        health_metrics["alerts"].append({
            "severity": "warning",
            "message": "TAG accuracy below 90%",
            "recommendation": "Run validation and correction"
        })

    if health_metrics["quality_metrics"]["link_integrity"] < 0.85:
        health_metrics["alerts"].append({
            "severity": "error",
            "message": "TAG link integrity compromised",
            "recommendation": "Run TAG chain validation"
        })

    # Update system status
    health_metrics["system_status"] = determine_system_status(health_metrics["alerts"])

    return health_metrics
```

---

## Integration Examples

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

## Quality Assurance

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

## Usage Examples

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