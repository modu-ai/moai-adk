# AI-Powered Test Generation

> Module: Context7-enhanced test case generation and specifications
> Complexity: Advanced
> Time: 15+ minutes
> Dependencies: the project's test runner, Context7 MCP, source parser (AST)

## Test Generator Class

The generator emits test code from a TestSpecification. The templates below are language-neutral skeletons — render them in the host language's test syntax (Go table-driven tests, Python pytest, JS/TS jest, Rust #[test], JUnit).

```text
class TestGenerator:
    context7
    templates = load_test_templates()

    load_test_templates():
        return {
          unit_function: """
            test "<function_name>_<scenario>":
                # Test: {description}
                # Given: {preconditions}
                # When:  {action}
                # Then:  {expected_outcome}
                # Arrange
                {setup_code}
                # Act
                result = {function_call}
                # Assert
                assert result == {expected_value}
          """,
          exception_test: """
            test "<function_name>_raises_<exception>_<scenario>":
                # Test that {function_name} raises {exception} when {condition}
                # Arrange
                {setup_code}
                # Act & Assert — assert that {exception} is raised
                assert raises({exception}): {function_call}
                    message contains "{expected_message}"
          """,
          parameterized_test: """
            # table-driven / parameterized test
            cases = {test_values}
            for ({param_names}) in cases:
                test "<function_name>_<scenario>":
                    # Test {function_name} with different inputs: {description}
                    # Arrange
                    {setup_code}
                    # Act
                    result = {function_call}
                    # Assert
                    assert result == {expected_value}
          """
        }

    generate_test_case(specification, context7_patterns = none):
        if context7 available and context7_patterns:
            try:
                enhanced = enhance_specification_with_context7(specification, context7_patterns)
                return generate_test_from_specification(enhanced)
            except e:
                log("Context7 test generation failed: " + e)
        return generate_test_from_specification(specification)

    enhance_specification_with_context7(specification, context7_patterns):
        additional_edge_cases = []
        testing_patterns = context7_patterns.testing default {}
        if testing_patterns:
            reqs = [lowercase(text(r)) for r in specification.requirements]
            if any("number" in r for r in reqs):
                additional_edge_cases += ["zero value", "negative value",
                                          "maximum/minimum values", "floating-point edge cases"]
            if any("string" in r for r in reqs):
                additional_edge_cases += ["empty string", "very long string",
                                          "special characters", "unicode characters"]
            if any("list" in r or "array" in r for r in reqs):
                additional_edge_cases += ["empty list", "single element",
                                          "large list", "duplicate elements"]
        combined = distinct(specification.edge_cases + additional_edge_cases)
        return specification.with(edge_cases=combined)

    generate_test_from_specification(spec):
        switch spec.test_type:
            case UNIT:        return generate_unit_test(spec)
            case INTEGRATION: return generate_integration_test(spec)
            default:          return generate_generic_test(spec)

    generate_unit_test(spec):
        function_name = strip_test_prefix(spec.name)
        if any("error" in c or "exception" in c for c in spec.acceptance_criteria):
            return generate_exception_test(spec, function_name)
        if len(spec.acceptance_criteria) > 1 or len(spec.edge_cases) > 2:
            return generate_parameterized_test(spec, function_name)
        return generate_standard_unit_test(spec, function_name)

    # generate_standard_unit_test / generate_exception_test / generate_parameterized_test /
    # generate_integration_test / generate_generic_test all select a template and fill its
    # placeholders from the spec (setup code, function call, assertions, expected value,
    # scenario name). The assertion helper maps acceptance-criterion phrasing
    # ("returns X", "equals Y", "length N") to assertion idioms; the parameter helper
    # builds (input, expected) rows from acceptance criteria and edge cases.
```

## Context7 Integration

```text
class Context7TestIntegration:
    context7
    pattern_cache = {}

    load_test_generation_patterns(language = "python"):
        cache_key = "test_gen_patterns_" + language
        if cache_key in pattern_cache: return pattern_cache[cache_key]

        patterns = {}
        if context7 available:
            try:
                patterns.generation = context7.get_library_docs(
                    "<testing-lib>", topic="test generation patterns automation", tokens=3000)
                patterns.edge_cases = context7.get_library_docs(
                    "<testing/edge-cases>", topic="edge case generation boundary testing", tokens=2000)
            except e:
                log("Failed to load Context7 patterns: " + e)
                patterns = default_patterns()
        else:
            patterns = default_patterns()
        pattern_cache[cache_key] = patterns
        return patterns

    default_patterns():
        return {
          generation: {
            strategies: [
                "Generate tests from specifications",
                "Analyze code to identify missing test cases",
                "Create parameterized / table-driven tests for multiple scenarios",
                "Generate exception tests for error conditions"]
          },
          edge_cases: {
            categories: [
                "Boundary values (min, max, just above/below)",
                "Empty/null inputs",
                "Invalid data types",
                "Special characters and unicode",
                "Large inputs (performance testing)"]
          }
        }
```

## Best Practices

1. Specification-Driven: Always generate tests from clear specifications
2. Edge Case Coverage: Use Context7 patterns to ensure comprehensive edge case testing
3. Readable Tests: Generate tests that clearly express intent
4. Maintainable: Keep generated tests simple and focused
5. Context-Aware: Leverage Context7 for language-specific and framework-specific patterns

---

Related: [ANALYZE-PRESERVE-IMPROVE](./analyze-preserve-improve.md) | [Test Patterns](./test-patterns.md)
