# Relevance Analysis - TRUST 5 Framework

> Module: Relevance category deep dive with requirements traceability and dead code detection
> Parent: [TRUST 5 Framework](./trust5-framework.md)
> Complexity: Advanced
> Time: 10+ minutes
> Dependencies: source parser (AST), requirements mapping

## Overview

Relevance (20% weight) validates requirements fulfillment and purpose alignment through feature completeness checks, requirements traceability, dead code identification, and purpose alignment validation.

## Requirements Traceability

### Missing Requirement Detection

```text
validate_requirements_traceability(file_path, content, requirements):
    issues = []
    for (req_id, requirement) in requirements:
        if req_id not in content:    # requirement identifier not referenced in code
            issues.append(CodeIssue(
                id="missing_requirement_" + req_id,
                category=TrustCategory.RELEVANCE, severity="medium",
                issue_type="documentation_issue", title="Missing Requirement Implementation",
                description="Requirement " + req_id + " not found in code",
                file_path=file_path, line_number=1, column_number=1,
                code_snippet="# TODO: Implement " + req_id + ": " + requirement,
                suggested_fix="Implement requirement " + req_id + ": " + requirement,
                confidence=0.8, rule_violated="MISSING_REQUIREMENT"))
    return issues
```

## Dead Code Detection

### Unused Function Detection

```text
detect_dead_code(file_path, tree):
    issues = []
    defined = set of names from walk(tree, matching=FunctionDecl)
    called  = set of names from walk(tree, matching=Call)

    unused = defined - called
    # Exclude private/internal symbols (e.g. leading underscore) and known
    # entry points (main, test functions, exported public symbols).
    unused = [n for n in unused if not is_private_or_entry_point(n)]

    for name in unused:
        node = find_definition(tree, name)
        issues.append(CodeIssue(
            id="dead_code_" + node.line,
            category=TrustCategory.RELEVANCE, severity="low",
            issue_type="code_smell", title="Dead Code",
            description="Function '" + name + "' is never called",
            file_path=file_path, line_number=node.line, column_number=1,
            code_snippet=signature(node),
            suggested_fix="Remove unused function '" + name + "' or update references",
            confidence=0.6, rule_violated="DEAD_CODE"))
    return issues
```

## Detection Patterns

### Common Relevance Issues

1. **Missing Requirements**: Specified features not implemented
2. **Unused Functions**: Functions defined but never called
3. **Unused Imports**: Imported modules not referenced
4. **Dead Classes**: Classes defined but never instantiated
5. **Commented Code**: Large blocks of commented-out code
6. **TODO/FIXME**: Unresolved development markers
7. **Feature Creep**: Code beyond original requirements

### Purpose Alignment Analysis

```text
check_purpose_alignment(file_path, content, purpose):
    issues = []
    complexity = calculate_complexity(content)
    if complexity > expected_complexity(purpose):
        issues.append(CodeIssue(
            id="over_engineered",
            category=TrustCategory.RELEVANCE, severity="low",
            issue_type="code_smell", title="Over-Engineered Solution",
            description="Code complexity exceeds what the stated purpose requires",
            file_path=file_path,
            suggested_fix="Simplify the implementation to match its purpose",
            confidence=0.7, rule_violated="PURPOSE_MISALIGNMENT"))
    return issues
```

## Best Practices

1. Requirements Mapping: Maintain explicit requirement-to-code traceability
2. Regular Cleanup: Remove dead code during refactoring
3. Purpose Documentation: Document module/class/function purposes
4. Complexity Monitoring: Monitor complexity growth vs requirements
5. Impact Analysis: Analyze impact before removing dead code

---

Version: 1.0.0
Last Updated: 2026-01-06
