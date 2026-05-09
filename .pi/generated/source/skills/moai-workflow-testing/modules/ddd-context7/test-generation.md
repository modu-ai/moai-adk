# AI-Powered Test Generation

> Module: Context7-enhanced test case generation and specifications
> Complexity: Advanced
> Time: 15+ minutes
> Dependencies: Python 3.8+, pytest, Context7 MCP, AST analysis

## Test Generator Class

```python
class TestGenerator:
    """AI-powered test case generation based on specifications."""

    def __init__(self, context7_client=None):
        self.context7 = context7_client
        self.templates = self._load_test_templates()

    def _load_test_templates(self) -> Dict[str, str]:
        """Load test templates for different scenarios."""
        return {
            'unit_function': '''
def test_{function_name}_{scenario}():
    """
    Test {description}

    Given: {preconditions}
    When: {action}
    Then: {expected_outcome}
    """
    # Arrange
    {setup_code}

    # Act
    result = {function_call}

    # Assert
    assert result == {expected_value}, f"Expected {expected_value}, got {result}"
''',
            'exception_test': '''
def test_{function_name}_raises_{exception}_{scenario}():
    """
    Test that {function_name} raises {exception} when {condition}
    """
    # Arrange
    {setup_code}

    # Act & Assert
    with pytest.raises({exception}) as exc_info:
        {function_call}

    assert "{expected_message}" in str(exc_info.value)
''',
            'parameterized_test': '''
@pytest.mark.parametrize("{param_names}", {test_values})
def test_{function_name}_{scenario}({param_names}):
    """
    Test {function_name} with different inputs: {description}
    """
    # Arrange
    {setup_code}

    # Act
    result = {function_call}

    # Assert
    assert result == {expected_value}, f"For {param_names}={{param_names}}, expected {expected_value}, got {{result}}"
'''
        }

    async def generate_test_case(
        self, specification: TestSpecification,
        context7_patterns: Dict[str, Any] = None
    ) -> str:
        """Generate test code based on specification."""

        if self.context7 and context7_patterns:
            try:
                # Use Context7 to enhance test generation
                enhanced_spec = await self._enhance_specification_with_context7(
                    specification, context7_patterns
                )
                return self._generate_test_from_enhanced_spec(enhanced_spec)
            except Exception as e:
                print(f"Context7 test generation failed: {e}")

        return self._generate_test_from_specification(specification)

    async def _enhance_specification_with_context7(
        self, specification: TestSpecification,
        context7_patterns: Dict[str, Any]
    ) -> TestSpecification:
        """Enhance test specification using Context7 patterns."""

        # Add additional edge cases based on Context7 patterns
        additional_edge_cases = []

        testing_patterns = context7_patterns.get('python_testing', {})
        if testing_patterns:
            # Add common edge cases for different data types
            if any('number' in str(req).lower() for req in specification.requirements):
                additional_edge_cases.extend([
                    "Test with zero value",
                    "Test with negative value",
                    "Test with maximum/minimum values",
                    "Test with floating point edge cases"
                ])

            if any('string' in str(req).lower() for req in specification.requirements):
                additional_edge_cases.extend([
                    "Test with empty string",
                    "Test with very long string",
                    "Test with special characters",
                    "Test with unicode characters"
                ])

            if any('list' in str(req).lower() or 'array' in str(req).lower()
                   for req in specification.requirements):
                additional_edge_cases.extend([
                    "Test with empty list",
                    "Test with single element",
                    "Test with large list",
                    "Test with duplicate elements"
                ])

        # Combine original and additional edge cases
        combined_edge_cases = list(set(specification.edge_cases + additional_edge_cases))

        return TestSpecification(
            name=specification.name,
            description=specification.description,
            test_type=specification.test_type,
            requirements=specification.requirements,
            acceptance_criteria=specification.acceptance_criteria,
            edge_cases=combined_edge_cases,
            preconditions=specification.preconditions,
            postconditions=specification.postconditions,
            dependencies=specification.dependencies,
            mock_requirements=specification.mock_requirements
        )

    def _generate_test_from_enhanced_spec(self, spec: TestSpecification) -> str:
        """Generate test code from enhanced specification."""
        return self._generate_test_from_specification(spec)

    def _generate_test_from_specification(self, spec: TestSpecification) -> str:
        """Generate test code from specification."""

        # Determine appropriate template based on test type
        if spec.test_type == TestType.UNIT:
            return self._generate_unit_test(spec)
        elif spec.test_type == TestType.INTEGRATION:
            return self._generate_integration_test(spec)
        else:
            return self._generate_generic_test(spec)

    def _generate_unit_test(self, spec: TestSpecification) -> str:
        """Generate unit test code."""

        function_name = spec.name.lower().replace('test_', '').replace('_test', '')

        # Check if this is an exception test
        if any('error' in criterion.lower() or 'exception' in criterion.lower()
               for criterion in spec.acceptance_criteria):
            return self._generate_exception_test(spec, function_name)

        # Check if this requires parameterization
        if len(spec.acceptance_criteria) > 1 or len(spec.edge_cases) > 2:
            return self._generate_parameterized_test(spec, function_name)

        # Generate standard unit test
        return self._generate_standard_unit_test(spec, function_name)

    def _generate_standard_unit_test(self, spec: TestSpecification, function_name: str) -> str:
        """Generate standard unit test."""

        template = self.templates['unit_function']

        setup_code = self._generate_setup_code(spec)
        function_call = self._generate_function_call(function_name, spec)
        assertions = self._generate_assertions(spec)

        return template.format(
            function_name=function_name,
            scenario=self._extract_scenario(spec),
            description=spec.description,
            preconditions=', '.join(spec.preconditions),
            action=self._describe_action(spec),
            expected_outcome=spec.acceptance_criteria[0] if spec.acceptance_criteria else "expected behavior",
            setup_code=setup_code,
            function_call=function_call,
            expected_value=self._extract_expected_value(spec),
            assertions=assertions
        )

    def _generate_exception_test(self, spec: TestSpecification, function_name: str) -> str:
        """Generate exception test."""

        template = self.templates['exception_test']

        # Extract expected exception and message
        exception_type = "Exception" # Default
        expected_message = "Error occurred"

        for criterion in spec.acceptance_criteria:
            if 'raise' in criterion.lower() or 'exception' in criterion.lower():
                # Try to extract exception type
                if 'valueerror' in criterion.lower():
                    exception_type = "ValueError"
                elif 'typeerror' in criterion.lower():
                    exception_type = "TypeError"
                elif 'attributeerror' in criterion.lower():
                    exception_type = "AttributeError"
                elif 'keyerror' in criterion.lower():
                    exception_type = "KeyError"

                # Try to extract expected message
                if 'message:' in criterion.lower():
                    parts = criterion.split('message:')
                    if len(parts) > 1:
                        expected_message = parts[1].strip().strip('"\'')
                        break

        return template.format(
            function_name=function_name,
            exception=exception_type,
            scenario=self._extract_scenario(spec),
            condition=self._describe_condition(spec),
            setup_code=self._generate_setup_code(spec),
            function_call=self._generate_function_call(function_name, spec),
            expected_message=expected_message
        )

    def _generate_parameterized_test(self, spec: TestSpecification, function_name: str) -> str:
        """Generate parameterized test."""

        template = self.templates['parameterized_test']

        # Generate test parameters and values
        param_names, test_values = self._generate_test_parameters(spec)

        return template.format(
            function_name=function_name,
            scenario=self._extract_scenario(spec),
            description=spec.description,
            param_names=', '.join(param_names),
            test_values=test_values,
            setup_code=self._generate_setup_code(spec),
            function_call=self._generate_function_call(function_name, spec),
            expected_value=self._extract_expected_value(spec)
        )

    def _generate_integration_test(self, spec: TestSpecification) -> str:
        """Generate integration test."""
        return f'''
def test_{spec.name.replace(' ', '_').lower()}():
    """Test: {spec.description}"""
    # Arrange: {', '.join(spec.preconditions[:2])}
    # Act: Call integration function
    # Assert: Verify integration behavior
'''

    def _generate_generic_test(self, spec: TestSpecification) -> str:
        """Generate generic test code."""
        return f'''
def test_{spec.name.replace(' ', '_').lower()}():
    """Test: {spec.description}"""
    # TODO: Implement based on specification
    # Requirements: {len(spec.requirements)} items
    # Acceptance Criteria: {len(spec.acceptance_criteria)} items
'''

    def _extract_scenario(self, spec: TestSpecification) -> str:
        """Extract scenario name from specification."""
        if '_' in spec.name:
            parts = spec.name.split('_')
            if len(parts) > 1:
                return '_'.join(parts[1:])
        return 'default'

    def _describe_action(self, spec: TestSpecification) -> str:
        """Describe the action being tested."""
        return f"Call {spec.name}"

    def _describe_condition(self, spec: TestSpecification) -> str:
        """Describe condition for exception test."""
        return spec.requirements[0] if spec.requirements else "invalid input"

    def _generate_setup_code(self, spec: TestSpecification) -> str:
        """Generate setup code based on specification."""
        setup_lines = []

        # Add mock requirements
        for mock_name, mock_config in spec.mock_requirements.items():
            if isinstance(mock_config, dict) and 'return_value' in mock_config:
                setup_lines.append(f"{mock_name} = Mock(return_value={mock_config['return_value']})")
            else:
                setup_lines.append(f"{mock_name} = Mock()")

        # Add preconditions as setup
        for condition in spec.preconditions:
            setup_lines.append(f"# {condition}")

        return '\n '.join(setup_lines) if setup_lines else "pass"

    def _generate_function_call(self, function_name: str, spec: TestSpecification) -> str:
        """Generate function call with arguments."""

        # Extract arguments from mock requirements or requirements
        args = []

        if spec.mock_requirements:
            args.extend(spec.mock_requirements.keys())

        if not args:
            # Add placeholder arguments based on requirements
            for req in spec.requirements[:3]: # Limit to first 3 requirements
                if 'input' in req.lower() or 'parameter' in req.lower():
                    args.append("test_input")
                    break

        return f"{function_name}({', '.join(args)})" if args else f"{function_name}()"

    def _generate_assertions(self, spec: TestSpecification) -> str:
        """Generate assertions based on acceptance criteria."""
        assertions = []

        for criterion in spec.acceptance_criteria[:3]: # Limit to first 3 criteria
            if 'returns' in criterion.lower() or 'result' in criterion.lower():
                assertions.append("assert result is not None")
            elif 'equals' in criterion.lower() or 'equal' in criterion.lower():
                assertions.append("assert result == expected_value")
            elif 'length' in criterion.lower():
                assertions.append("assert len(result) > 0")
            else:
                assertions.append(f"# {criterion}")

        return '\n '.join(assertions) if assertions else "assert True # Add specific assertions"

    def _extract_expected_value(self, spec: TestSpecification) -> str:
        """Extract expected value from acceptance criteria."""
        for criterion in spec.acceptance_criteria:
            if 'returns' in criterion.lower():
                # Try to extract expected value
                if 'true' in criterion.lower():
                    return "True"
                elif 'false' in criterion.lower():
                    return "False"
                elif 'none' in criterion.lower():
                    return "None"
                elif 'empty' in criterion.lower():
                    return "[]"
                else:
                    return "expected_result"
        return "expected_result"

    def _generate_test_parameters(self, spec: TestSpecification) -> tuple:
        """Generate parameters and values for parameterized tests."""

        # Create test cases from acceptance criteria and edge cases
        test_cases = []

        # Add acceptance criteria as test cases
        for criterion in spec.acceptance_criteria:
            if 'input' in criterion.lower():
                # Extract input values
                if 'valid' in criterion.lower():
                    test_cases.append(('valid_input', 'expected_output'))
                elif 'invalid' in criterion.lower():
                    test_cases.append(('invalid_input', 'exception'))

        # Add edge cases
        for edge_case in spec.edge_cases:
            if 'zero' in edge_case.lower():
                test_cases.append((0, 'zero_result'))
            elif 'empty' in edge_case.lower():
                test_cases.append(('', 'empty_result'))
            elif 'null' in edge_case.lower() or 'none' in edge_case.lower():
                test_cases.append((None, 'none_result'))

        # Convert to pytest format
        if test_cases:
            param_names = ['test_input', 'expected_output']
            test_values = str(test_cases).replace("'", '"')
            return param_names, test_values

        # Fallback
        return ['test_input', 'expected_output'], '[("test", "expected")]'
```

## Context7 Integration

```python
class Context7TestIntegration:
    """Integration with Context7 for test generation patterns."""

    def __init__(self, context7_client=None):
        self.context7 = context7_client
        self.pattern_cache = {}

    async def load_test_generation_patterns(
        self, language: str = "python"
    ) -> Dict[str, Any]:
        """Load test generation patterns from Context7."""

        cache_key = f"test_gen_patterns_{language}"
        if cache_key in self.pattern_cache:
            return self.pattern_cache[cache_key]

        patterns = {}

        if self.context7:
            try:
                # Load test generation patterns
                gen_patterns = await self.context7.get_library_docs(
                    context7_library_id="/testing/pytest",
                    topic="test generation patterns automation 2025",
                    tokens=3000
                )
                patterns['generation'] = gen_patterns

                # Load edge case patterns
                edge_patterns = await self.context7.get_library_docs(
                    context7_library_id="/testing/edge-cases",
                    topic="edge case generation boundary testing 2025",
                    tokens=2000
                )
                patterns['edge_cases'] = edge_patterns

            except Exception as e:
                print(f"Failed to load Context7 patterns: {e}")
                patterns = self._get_default_patterns()
        else:
            patterns = self._get_default_patterns()

        self.pattern_cache[cache_key] = patterns
        return patterns

    def _get_default_patterns(self) -> Dict[str, Any]:
        """Get default test generation patterns."""
        return {
            'generation': {
                'strategies': [
                    "Generate tests from specifications",
                    "Analyze code to identify missing test cases",
                    "Create parameterized tests for multiple scenarios",
                    "Generate exception tests for error conditions"
                ]
            },
            'edge_cases': {
                'categories': [
                    "Boundary values (min, max, just above/below)",
                    "Empty/null inputs",
                    "Invalid data types",
                    "Special characters and unicode",
                    "Large inputs (performance testing)"
                ]
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
