# Advanced DDD Features with Context7

> Module: AI-powered comprehensive test suite generation and analysis
> Complexity: Expert
> Time: 25+ minutes
> Dependencies: Python 3.8+, pytest, Context7 MCP, AST analysis, asyncio

## Enhanced Test Generator

```python
import ast
import inspect

class EnhancedTestGenerator(TestGenerator):
    """Enhanced test generator with advanced Context7 integration."""

    async def generate_comprehensive_test_suite(
        self, function_code: str,
        context7_patterns: Dict[str, Any]
    ) -> List[str]:
        """Generate comprehensive test suite from function code."""

        # Analyze function using AST
        function_analysis = self._analyze_function_code(function_code)

        # Generate tests for different scenarios
        test_cases = []

        # Happy path tests
        happy_path_tests = await self._generate_happy_path_tests(
            function_analysis, context7_patterns
        )
        test_cases.extend(happy_path_tests)

        # Edge case tests
        edge_case_tests = await self._generate_edge_case_tests(
            function_analysis, context7_patterns
        )
        test_cases.extend(edge_case_tests)

        # Error handling tests
        error_tests = await self._generate_error_handling_tests(
            function_analysis, context7_patterns
        )
        test_cases.extend(error_tests)

        # Performance tests for critical functions
        if self._is_performance_critical(function_analysis):
            perf_tests = await self._generate_performance_tests(function_analysis)
            test_cases.extend(perf_tests)

        return test_cases

    def _analyze_function_code(self, code: str) -> Dict[str, Any]:
        """Analyze function code to extract test requirements."""

        try:
            tree = ast.parse(code)

            analysis = {
                'functions': [],
                'parameters': [],
                'return_statements': [],
                'exceptions': [],
                'external_calls': []
            }

            for node in ast.walk(tree):
                if isinstance(node, ast.FunctionDef):
                    analysis['functions'].append({
                        'name': node.name,
                        'args': [arg.arg for arg in node.args.args],
                        'decorators': [d.id if isinstance(d, ast.Name) else str(d)
                                      for d in node.decorator_list]
                    })

                elif isinstance(node, ast.Raise):
                    analysis['exceptions'].append({
                        'type': node.exc.func.id if node.exc and hasattr(node.exc, 'func') else 'Exception',
                        'message': node.exc.msg if node.exc and hasattr(node.exc, 'msg') else None
                    })

                elif isinstance(node, ast.Call):
                    if isinstance(node.func, ast.Attribute):
                        analysis['external_calls'].append(f"{node.func.value.id}.{node.func.attr}")
                    elif isinstance(node.func, ast.Name):
                        analysis['external_calls'].append(node.func.id)

            return analysis

        except Exception as e:
            print(f"Error analyzing function code: {e}")
            return {}

    async def _generate_happy_path_tests(
        self, analysis: Dict[str, Any],
        context7_patterns: Dict[str, Any]
    ) -> List[str]:
        """Generate happy path test cases."""

        tests = []

        for func in analysis.get('functions', []):
            # Generate test for normal operation
            test_code = f"""
def test_{func['name']}_happy_path():
    '''
    Test {func['name']} with valid inputs.

    Given: Valid input parameters
    When: {func['name']} is called
    Then: Expected result is returned
    '''
    # Arrange
    # Add setup code based on parameters: {', '.join(func['args'])}

    # Act
    # result = {func['name']}(*args)

    # Assert
    # assert result is not None
"""
            tests.append(test_code)

        return tests

    async def _generate_edge_case_tests(
        self, analysis: Dict[str, Any],
        context7_patterns: Dict[str, Any]
    ) -> List[str]:
        """Generate edge case test scenarios."""

        tests = []

        for func in analysis.get('functions', []):
            # Generate tests for edge cases
            edge_cases = [
                ("empty_input", "Test with empty input"),
                ("null_input", "Test with None/null input"),
                ("boundary_value", "Test with boundary values"),
                ("max_input", "Test with maximum allowed input"),
                ("min_input", "Test with minimum allowed input")
            ]

            for case_name, description in edge_cases:
                test_code = f"""
def test_{func['name']}_{case_name}():
    '''
    {description}

    Given: Edge case input ({case_name})
    When: {func['name']} is called
    Then: Function handles edge case appropriately
    '''
    # Arrange
    # Setup edge case input

    # Act
    # result = {func['name']}(*edge_case_args)

    # Assert
    # Verify function handles edge case
"""
                tests.append(test_code)

        return tests

    async def _generate_error_handling_tests(
        self, analysis: Dict[str, Any],
        context7_patterns: Dict[str, Any]
    ) -> List[str]:
        """Generate error handling test cases."""

        tests = []

        for exc in analysis.get('exceptions', []):
            exc_type = exc.get('type', 'Exception')
            test_code = f"""
def test_error_handling_{exc_type.lower()}():
    '''
    Test {exc_type} error handling.

    Given: Invalid input or error condition
    When: Function is called with invalid input
    Then: Appropriate exception is raised
    '''
    # Arrange
    # Setup invalid input

    # Act & Assert
    with pytest.raises({exc_type}):
        # function_call()
        pass
"""
            tests.append(test_code)

        return tests

    async def _generate_performance_tests(
        self, analysis: Dict[str, Any]
    ) -> List[str]:
        """Generate performance test cases."""

        tests = []

        for func in analysis.get('functions', []):
            test_code = f"""
def test_{func['name']}_performance():
    '''
    Test {func['name']} performance characteristics.

    Given: Large input dataset
    When: {func['name']} is called
    Then: Function completes within acceptable time
    '''
    # Arrange
    import time
    large_input = list(range(10000))

    # Act
    start_time = time.time()
    # result = {func['name']}(large_input)
    execution_time = time.time() - start_time

    # Assert
    assert execution_time < 1.0, f"Function too slow: {{execution_time}}s"
"""
            tests.append(test_code)

        return tests

    def _is_performance_critical(self, analysis: Dict[str, Any]) -> bool:
        """Determine if function is performance-critical."""

        # Check for performance indicators
        func_names = [f['name'] for f in analysis.get('functions', [])]

        performance_keywords = ['process', 'calculate', 'compute', 'parse', 'transform']

        return any(
            any(keyword in name.lower() for keyword in performance_keywords)
            for name in func_names
        )
```

## Context7-Enhanced Testing

```python
class Context7EnhancedTesting:
    """Advanced testing capabilities with Context7 integration."""

    def __init__(self, context7_client=None):
        self.context7 = context7_client

    async def get_intelligent_test_suggestions(
        self, codebase_context: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Get AI-powered test suggestions using Context7."""

        if not self.context7:
            return self._get_rule_based_suggestions()

        try:
            # Get advanced testing patterns
            advanced_patterns = await self.context7.get_library_docs(
                context7_library_id="/testing/advanced",
                topic="property-based testing mutation testing 2025",
                tokens=5000
            )

            # Get test quality metrics
            quality_patterns = await self.context7.get_library_docs(
                context7_library_id="/testing/quality",
                topic="test quality analysis coverage gaps 2025",
                tokens=3000
            )

            return {
                'advanced_patterns': advanced_patterns,
                'quality_metrics': quality_patterns,
                'suggestions': self._generate_intelligent_suggestions(
                    advanced_patterns, quality_patterns, codebase_context
                )
            }

        except Exception as e:
            print(f"Context7 test suggestions failed: {e}")
            return self._get_rule_based_suggestions()

    def _generate_intelligent_suggestions(
        self, advanced_patterns: Dict, quality_patterns: Dict, context: Dict
    ) -> List[str]:
        """Generate intelligent test suggestions."""

        suggestions = []

        # Analyze codebase context
        coverage = context.get('coverage_percentage', 0)

        if coverage < 80:
            suggestions.append("Increase test coverage to at least 80%")

        # Check for missing test types
        test_types = context.get('test_types', [])
        if 'integration' not in test_types:
            suggestions.append("Add integration tests for component interactions")

        if 'performance' not in test_types:
            suggestions.append("Add performance tests for critical paths")

        if 'security' not in test_types:
            suggestions.append("Add security tests for authentication and authorization")

        return suggestions

    def _get_rule_based_suggestions(self) -> Dict[str, Any]:
        """Get rule-based testing suggestions."""

        return {
            'suggestions': [
                "Analyze existing behavior before refactoring (DDD)",
                "Aim for high test coverage (80%+)",
                "Test both positive and negative cases",
                "Use mocking for external dependencies",
                "Parameterize tests for multiple scenarios",
                "Add performance tests for critical functions",
                "Implement property-based testing for data validation",
                "Use mutation testing to verify test quality"
            ]
        }
```

## Property-Based Testing

```python
from hypothesis import given, strategies as st

class PropertyBasedTests:
    """Property-based testing with Hypothesis."""

    @given(st.integers(), st.integers())
    def test_addition_commutative(self, a, b):
        """Test that addition is commutative."""
        assert add(a, b) == add(b, a)

    @given(st.lists(st.integers()))
    def test_sort_idempotent(self, lst):
        """Test that sorting is idempotent."""
        result = sort(lst)
        assert sort(result) == result

    @given(st.text())
    def test_reverse_inverse(self, text):
        """Test that reverse is its own inverse."""
        assert reverse(reverse(text)) == text

    @given(st.integers(min_value=0, max_value=1000))
    def test_square_non_negative(self, x):
        """Test that square of any number is non-negative."""
        assert square(x) >= 0
```

## Mutation Testing

```python
class MutationTesting:
    """Mutation testing to verify test quality."""

    def __init__(self, project_path: str):
        self.project_path = Path(project_path)

    async def run_mutation_tests(self) -> Dict[str, Any]:
        """Run mutation tests to check test suite quality."""

        try:
            # Use mutmut (Python mutation testing tool)
            result = subprocess.run(
                [
                    sys.executable, '-m', 'mutmut',
                    'run',
                    '--paths-to-mutate', 'src'
                ],
                capture_output=True,
                text=True,
                cwd=str(self.project_path)
            )

            return self._parse_mutation_results(result.stdout)

        except Exception as e:
            return {'error': str(e), 'killed_mutants': 0, 'survived_mutants': 0}

    def _parse_mutation_results(self, output: str) -> Dict[str, Any]:
        """Parse mutation testing results."""

        # Parse output to extract mutation statistics
        lines = output.split('\n')

        results = {
            'total_mutations': 0,
            'killed_mutants': 0,
            'survived_mutants': 0,
            'mutation_score': 0.0
        }

        for line in lines:
            if 'killed' in line.lower():
                parts = line.split()
                if len(parts) >= 2 and parts[0].isdigit():
                    results['killed_mutants'] = int(parts[0])

            elif 'survived' in line.lower():
                parts = line.split()
                if len(parts) >= 2 and parts[0].isdigit():
                    results['survived_mutants'] = int(parts[0])

        # Calculate mutation score
        results['total_mutations'] = results['killed_mutants'] + results['survived_mutants']

        if results['total_mutations'] > 0:
            results['mutation_score'] = (
                results['killed_mutants'] / results['total_mutations']
            ) * 100

        return results
```

## Continuous Testing Integration

```python
class ContinuousTesting:
    """Continuous testing integration for DDD workflow."""

    def __init__(self, project_path: str):
        self.project_path = Path(project_path)
        self.test_watcher = None

    async def start_watch_mode(self):
        """Start watching files for changes and run tests automatically."""

        try:
            # Use pytest-watch or pytest-xdist for continuous testing
            result = subprocess.run(
                [
                    sys.executable, '-m', 'pytest_watch',
                    '--', str(self.project_path)
                ],
                capture_output=False,
                cwd=str(self.project_path)
            )

        except Exception as e:
            print(f"Watch mode error: {e}")

    async def run_parallel_tests(self, num_workers: int = 4) -> Dict[str, Any]:
        """Run tests in parallel for faster feedback."""

        try:
            result = subprocess.run(
                [
                    sys.executable, '-m', 'pytest',
                    '-n', str(num_workers),
                    str(self.project_path)
                ],
                capture_output=True,
                text=True,
                cwd=str(self.project_path)
            )

            return {
                'output': result.stdout,
                'success': result.returncode == 0
            }

        except Exception as e:
            return {'error': str(e), 'success': False}
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
