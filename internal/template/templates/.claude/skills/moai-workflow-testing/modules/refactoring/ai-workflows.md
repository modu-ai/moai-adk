# AI-Powered Refactoring Workflows

> Sub-module: AI-assisted refactoring workflows with Documentation integration
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: WebSearch/WebFetch, source parser (AST), the host language's refactoring tool

## Overview

This module provides advanced AI-assisted refactoring workflows that leverage WebSearch/WebFetch for real-time access to latest refactoring patterns, best practices, and framework-specific guidance.

---

## Context-Aware Refactoring

### Project Convention Detection

Detect and respect project-specific conventions for intelligent refactoring. Walk the host language's syntax tree to collect identifier names, then infer the dominant naming convention.

```text
class ContextAwareRefactorer(AIRefactorer):
    project_conventions = {}
    api_boundaries = set()

    analyze_project_context(codebase_path):
        detect_naming_conventions(codebase_path)
        identify_api_boundaries(codebase_path)
        analyze_architecture_patterns(codebase_path)

    detect_naming_conventions(codebase_path):
        naming = { variables: [], functions: [], classes: [], constants: [] }
        for file_path in sample(find_source_files(codebase_path), 50):
            try:
                tree = parse(read(file_path))
                # Collect identifier names by symbol kind, using the host
                # language's parser visitor / traversal API.
                for node in walk(tree):
                    if node is a VariableDecl: naming.variables.append(node.name)
                    if node is a FunctionDecl: naming.functions.append(node.name)
                    if node is a ClassDecl:    naming.classes.append(node.name)
                    if node is a ConstantDecl:naming.constants.append(node.name)
            except e:
                log("Error analyzing " + file_path + ": " + e)
        project_conventions = analyze_naming_patterns(naming)

    analyze_naming_patterns(patterns):
        conventions = {}
        # Pick the dominant case style among variables and among functions
        snake_count = count(names in patterns.variables matching snake_case)
        camel_count = count(names in patterns.variables matching camelCase)
        conventions.variable_naming = "snake_case" if snake_count > camel_count else "camelCase"
        # (repeat for functions)
        return conventions
```

### API Boundary Detection

Identify public vs internal APIs for safer refactoring:

```text
identify_api_boundaries(codebase_path):
    for file_path in find_source_files(codebase_path):
        # Heuristic: files under an api/ or public/ path are public API
        if "api" in path_segments(file_path) or "public" in path_segments(file_path):
            api_boundaries.add(file_path)
        # Heuristic: a file declaring an explicit export list (Python __all__,
        # Go exported identifiers, JS/TS named exports, Rust pub items) is public API
        if declares_explicit_exports(read(file_path)):
            api_boundaries.add(file_path)
```

---

## Documentation Integration

### Real-Time Pattern Retrieval

Fetch latest refactoring patterns from Documentation:

```text
get_docs_refactoring_patterns():
    patterns = {}
    if docs available:
        try:
            patterns.tool = docs.get_library_docs(
                "<host-language-refactor-tool>",   # e.g. python-rope, gopls, rust-analyzer
                topic="safe refactoring patterns technical debt",
                tokens=4000)
            patterns.general = docs.get_library_docs(
                "<refactoring/guru>",
                topic="code refactoring best practices design patterns",
                tokens=3000)
        except e:
            log("Failed to get Documentation patterns: " + e)
    return patterns
```

### Framework-Specific Patterns

Get refactoring patterns specific to a framework. Populate the table with the project's actual frameworks (the entries below are illustrative).

```text
get_framework_patterns(framework):
    # Map each framework to its Documentation library id + topic. Examples:
    #   web (python):  django / fastapi / flask
    #   web (go):      gin / echo / chi
    #   web (node):    express / nestjs / fastify
    #   web (rust):    axum / actix
    framework_patterns = {
        # "<framework>": { library_id: "<org/repo>", topic: "<framework> refactoring best practices" }
    }
    if framework not in framework_patterns: return {}
    config = framework_patterns[framework]
    try:
        return docs.get_library_docs(config.library_id, config.topic, tokens=5000)
    except e:
        log("Failed to get " + framework + " patterns: " + e)
        return {}
```

---

## Intelligent Refactoring Pipeline

### End-to-End Workflow

Complete AI-assisted refactoring pipeline:

```text
intelligent_refactoring_pipeline(codebase_path, framework = none, max_risk_level = "medium"):
    # Step 1: analyze project context
    analyze_project_context(codebase_path)

    # Step 2: analyze technical debt
    debt_items = technical_debt_analyzer.analyze(codebase_path)

    # Step 3: fetch Documentation patterns (+ framework-specific if requested)
    docs_patterns = get_docs_refactoring_patterns()
    if framework:
        docs_patterns.framework = get_framework_patterns(framework)

    # Step 4: identify opportunities with AI
    opportunities = identify_refactor_opportunities(codebase_path, debt_items, docs_patterns)

    # Step 5: filter by risk and conventions
    filtered = filter_by_conventions_and_risk(opportunities, max_risk_level)

    # Step 6: create a safe execution plan
    return create_safe_refactor_plan(filtered, debt_items, docs_patterns)

filter_by_conventions_and_risk(opportunities, max_risk_level):
    risk_order = { low: 1, medium: 2, high: 3 }
    max_risk_value = risk_order[max_risk_level] default 2
    filtered = []
    for opp in opportunities:
        if risk_order[opp.risk_level] default 3 > max_risk_value: continue
        # Don't refactor public APIs at above-low risk
        if opp.file_path in api_boundaries and opp.risk_level != "low": continue
        filtered.append(opp)
    return filtered
```

---

## Safe Refactoring Execution

### Pre-Refactoring Checklist

Verify conditions before starting refactoring:

```text
pre_refactoring_checklist(codebase_path):
    checks = {
        has_tests:           len(find_test_files(codebase_path)) > 0,
        has_version_control: exists(codebase_path / ".git"),
        tests_passing:       false,
        coverage_sufficient: false
    }
    if checks.has_tests:
        result = run_tests(codebase_path)
        checks.tests_passing = result.passed == result.total
    if checks.has_tests and checks.tests_passing:
        checks.coverage_sufficient = calculate_coverage(codebase_path) >= 0.8
    return checks

run_tests(codebase_path):
    # Invoke the project's test runner (go test, pytest, jest, cargo test, dotnet test...)
    try:
        output = exec(test_runner_cmd(), cwd=codebase_path, timeout=300)
        return parse_test_summary(output) default { total: 0, passed: 0 }
    except e:
        log("Error running tests: " + e)
        return { total: 0, passed: 0 }

calculate_coverage(codebase_path):
    # Invoke the project's coverage tool and parse the TOTAL coverage line.
    try:
        output = exec(coverage_cmd(), cwd=codebase_path, timeout=300)
        return parse_total_coverage(output) / 100 default 0.0
    except e:
        log("Error calculating coverage: " + e)
        return 0.0
```

### Incremental Refactoring Execution

Execute refactoring in safe, incremental steps:

```text
execute_refactoring_plan(refactor_plan, codebase_path):
    results = { successful: [], failed: [], skipped: [] }

    checks = pre_refactoring_checklist(codebase_path)
    if not all(checks.values()):
        log("Pre-refactoring checks failed:")
        for (check, passed) in checks: if not passed: log("  - " + check + ": FAILED")
        return results

    for (i, opp_index) in enumerate(refactor_plan.execution_order, from=1):
        opportunity = refactor_plan.opportunities[opp_index]
        log("[" + i + "/" + len(refactor_plan.execution_order) + "] " + opportunity.description)
        try:
            git_commit(codebase_path, "Before refactoring: " + opportunity.description)
            success = execute_single_refactoring(opportunity, codebase_path)
            if success:
                test_result = run_tests(codebase_path)
                if test_result.passed == test_result.total:
                    results.successful.append(opportunity)
                    git_commit(codebase_path, "After refactoring: " + opportunity.description)
                else:
                    results.failed.append(opportunity)
                    git_revert(codebase_path)        # revert on test failure
            else:
                results.skipped.append(opportunity)
        except e:
            log("Error executing refactoring: " + e)
            results.failed.append(opportunity)
            git_revert(codebase_path)
    return results

execute_single_refactoring(opportunity, codebase_path):
    try:
        switch opportunity.type:
            case EXTRACT_METHOD:     return execute_extract_method(opportunity, codebase_path)
            case REORGANIZE_IMPORTS: return execute_reorganize_imports(opportunity, codebase_path)
            case RENAME:             return execute_rename(opportunity, codebase_path)
            # ... other refactoring types
            default: log("Unsupported refactoring type: " + opportunity.type); return false
    except e:
        log("Error executing " + opportunity.type + ": " + e); return false

git_commit(codebase_path, message):
    try:
        exec(["git", "add", "."], cwd=codebase_path)
        exec(["git", "commit", "-m", message], cwd=codebase_path)
    except e: log("Error creating git commit: " + e)

git_revert(codebase_path):
    try: exec(["git", "revert", "--no-commit", "HEAD"], cwd=codebase_path)
    except e: log("Error reverting git commit: " + e)
```

---

## Post-Refactoring Analysis

### Impact Assessment

Analyze the impact of refactoring changes:

```text
post_refactoring_analysis(codebase_path, execution_results):
    analysis = {
        debt_reduction:      0,
        complexity_reduction:0,
        lines_changed:       0,
        files_modified:      len(execution_results.successful),
        test_results:        {}
    }
    new_debt_items = technical_debt_analyzer.analyze(codebase_path)
    original_debt_count = <tracked from before>   # capture the pre-refactor count
    analysis.debt_reduction = (original_debt_count - len(new_debt_items)) / original_debt_count
    analysis.test_results = run_tests(codebase_path)

    # Count lines changed from git
    try:
        output = exec(["git", "diff", "--shortstat"], cwd=codebase_path)
        analysis.lines_changed = parse_insertions_deletions(output)
    except e: log("Error counting lines changed: " + e)
    return analysis
```

---

## Best Practices

1. Always run pre-refactoring checks before starting
2. Create git commits before each operation
3. Run tests after every change
4. Revert immediately on test failure
5. Document all changes in commit messages
6. Use Documentation for latest patterns and practices
7. Respect project conventions when refactoring
8. Prioritize low-risk, high-impact changes first
9. Monitor performance impact of changes
10. Keep refactoring sessions short and focused

---

## Resources

### WebSearch/WebFetch Libraries

- Host-language refactor tools: rope, gopls, rust-analyzer, TS language server
- General: `/refactoring/guru`
- Frameworks: the project's actual framework docs (django/fastapi/flask, gin/echo, express/nestjs, axum/actix, etc.)

### Tools

- Refactoring tool: the host language's refactor library/LSP
- Git: Version control and rollback
- Test runner: the project's test framework
- Coverage tool: the project's coverage measurement tool

---

Sub-module: `modules/refactoring/ai-workflows.md`
Related: [patterns.md](./patterns.md) | [../smart-refactoring.md](../smart-refactoring.md)
