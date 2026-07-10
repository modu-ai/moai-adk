# Advanced DDD Features with Context7

> Module: AI-powered comprehensive test suite generation and analysis
> Complexity: Expert
> Time: 25+ minutes
> Dependencies: the project's test runner, Context7 MCP, source parser (AST)

## Enhanced Test Generator

```text
class EnhancedTestGenerator(TestGenerator):
    generate_comprehensive_test_suite(function_code, context7_patterns):
        # Analyze the function via the host language's parser
        function_analysis = analyze_function_code(function_code)

        test_cases = []
        test_cases.extend(generate_happy_path_tests   (function_analysis, context7_patterns))
        test_cases.extend(generate_edge_case_tests    (function_analysis, context7_patterns))
        test_cases.extend(generate_error_handling_tests(function_analysis, context7_patterns))
        if is_performance_critical(function_analysis):
            test_cases.extend(generate_performance_tests(function_analysis))
        return test_cases

    analyze_function_code(code):
        try:
            tree = parse(code)     # host language's parser
            analysis = { functions: [], parameters: [], return_statements: [], exceptions: [], external_calls: [] }
            for node in walk(tree):
                if node is a FunctionDecl:
                    analysis.functions.append({ name: node.name, args: node.params, decorators: node.decorators })
                if node is a Throw/Raise:
                    analysis.exceptions.append({ type: node.exception_type, message: node.message })
                if node is a Call:
                    analysis.external_calls.append(callee_name(node))
            return analysis
        except e:
            log("Error analyzing function code: " + e)
            return {}

    generate_happy_path_tests(analysis, context7_patterns):
        tests = []
        for func in analysis.functions:
            tests.append(test_template(
                name="test_" + func.name + "_happy_path",
                given="valid input parameters",
                when=func.name + " is called",
                then="the expected result is returned",
                arrange="setup using parameters: " + join(func.args),
                act="result = " + func.name + "(...args)",
                assert="result is not none"))
        return tests

    generate_edge_case_tests(analysis, context7_patterns):
        tests = []
        edge_cases = [
            ("empty_input",    "Test with empty input"),
            ("null_input",     "Test with null/none input"),
            ("boundary_value", "Test with boundary values"),
            ("max_input",      "Test with maximum allowed input"),
            ("min_input",      "Test with minimum allowed input")]
        for func in analysis.functions:
            for (case_name, description) in edge_cases:
                tests.append(test_template(
                    name="test_" + func.name + "_" + case_name,
                    given="edge case input (" + case_name + ")",
                    when=func.name + " is called",
                    then="function handles the edge case appropriately"))
        return tests

    generate_error_handling_tests(analysis, context7_patterns):
        tests = []
        for exc in analysis.exceptions:
            tests.append(test_template(
                name="test_error_handling_" + lowercase(exc.type),
                given="invalid input or error condition",
                when="the function is called with invalid input",
                then=exc.type + " is raised",
                assert="assert raises(" + exc.type + ")"))
        return tests

    generate_performance_tests(analysis):
        tests = []
        for func in analysis.functions:
            tests.append(test_template(
                name="test_" + func.name + "_performance",
                given="a large input dataset",
                when=func.name + " is called",
                then="the function completes within an acceptable time",
                act="measure execution time on large input",
                assert="execution_time < threshold"))
        return tests

    is_performance_critical(analysis):
        performance_keywords = ["process", "calculate", "compute", "parse", "transform"]
        names = [f.name for f in analysis.functions]
        return any keyword appears in any name (case-insensitive)
```

## Context7-Enhanced Testing

```text
class Context7EnhancedTesting(context7):
    get_intelligent_test_suggestions(codebase_context):
        if context7 is none: return rule_based_suggestions()
        try:
            advanced_patterns = context7.get_library_docs("<testing/advanced>",
                                    topic="property-based testing mutation testing", tokens=5000)
            quality_patterns  = context7.get_library_docs("<testing/quality>",
                                    topic="test quality analysis coverage gaps", tokens=3000)
            return {
                advanced_patterns, quality_patterns,
                suggestions: generate_intelligent_suggestions(advanced_patterns, quality_patterns, codebase_context)
            }
        except e:
            log("Context7 test suggestions failed: " + e)
            return rule_based_suggestions()

    generate_intelligent_suggestions(advanced_patterns, quality_patterns, context):
        suggestions = []
        coverage = context.coverage_percentage default 0
        if coverage < 80: suggestions.append("Increase test coverage to at least 80%")
        test_types = context.test_types default []
        if "integration" not in test_types: suggestions.append("Add integration tests for component interactions")
        if "performance"  not in test_types: suggestions.append("Add performance tests for critical paths")
        if "security"     not in test_types: suggestions.append("Add security tests for authentication and authorization")
        return suggestions

    rule_based_suggestions():
        return { suggestions: [
            "Analyze existing behavior before refactoring (DDD)",
            "Aim for high test coverage (80%+)",
            "Test both positive and negative cases",
            "Use mocking for external dependencies",
            "Parameterize / table-drive tests for multiple scenarios",
            "Add performance tests for critical functions",
            "Implement property-based testing for data validation",
            "Use mutation testing to verify test quality"
        ]}
```

## Property-Based Testing

Property-based testing generates many inputs from a strategy and asserts an invariant holds. The idiom exists in most languages (Python Hypothesis, Go `testing/quick` + `rapid`, Rust `proptest`, JS `fast-check`, Haskell QuickCheck).

```text
# Conceptual property tests (syntax is library/language-specific):
property "addition is commutative"      (a, b integers):  assert add(a, b) == add(b, a)
property "sort is idempotent"           (lst list<int>):  assert sort(sort(lst)) == sort(lst)
property "reverse is its own inverse"   (s string):       assert reverse(reverse(s)) == s
property "square is non-negative"       (x int in [0,1000]): assert square(x) >= 0
```

## Mutation Testing

Mutation testing verifies test-suite quality by injecting small faults and checking whether the suite catches them. Per-language tools: Python `mutmut`/`cosmic-ray`, Go `go-mutesting`, JS `stryker`, Rust `mutagen`.

```text
class MutationTesting(project_path):
    run_mutation_tests():
        try:
            output = exec([mutation_tool_cmd(), "--paths-to-mutate", "src"],
                          cwd=project_path, capture=stdout)
            return parse_mutation_results(output)
        except e:
            return { error: text(e), killed_mutants: 0, survived_mutants: 0 }

    parse_mutation_results(output):
        results = { total_mutations: 0, killed_mutants: 0, survived_mutants: 0, mutation_score: 0.0 }
        for line in lines(output):
            if "killed"   in lowercase(line): results.killed_mutants   = parse_count(line)
            if "survived" in lowercase(line): results.survived_mutants = parse_count(line)
        results.total_mutations = results.killed_mutants + results.survived_mutants
        if results.total_mutations > 0:
            results.mutation_score = results.killed_mutants / results.total_mutations * 100
        return results
```

## Continuous Testing Integration

```text
class ContinuousTesting(project_path):
    start_watch_mode():
        # Watch source files and re-run the test suite on change.
        # Idiom per language: pytest-watch, Go `air`/`entr`, JS jest --watch, Rust cargo-watch.
        try:
            exec([watch_test_cmd(), "--", project_path], cwd=project_path)
        except e: log("Watch mode error: " + e)

    run_parallel_tests(num_workers = 4):
        # Run the suite across N workers for faster feedback.
        try:
            result = exec([test_runner_cmd(), "--workers", num_workers, project_path],
                          cwd=project_path, capture=stdout)
            return { output: result.stdout, success: result.exit_code == 0 }
        except e:
            return { error: text(e), success: false }
```

## Best Practices

1. Comprehensive Testing: Use Context7 patterns to ensure complete test coverage
2. Property-Based Testing: Add property-based tests for data validation functions
3. Mutation Testing: Use mutation testing to verify test suite quality
4. Continuous Testing: Implement watch mode for immediate feedback
5. Performance Testing: Add performance tests for critical paths
6. Security Testing: Include security tests for authentication and authorization
7. Integration Testing: Test component interactions thoroughly
8. Test Documentation: Document test intent and expected behavior
9. Context7 Integration: Leverage Context7 for latest testing patterns and practices
10. Automated Analysis: Use AI-powered test suggestions to identify gaps

---

Related: [ANALYZE-PRESERVE-IMPROVE](./analyze-preserve-improve.md) | [Test Generation](./test-generation.md) | [Test Patterns](./test-patterns.md)
