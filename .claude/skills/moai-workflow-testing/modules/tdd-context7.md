# Test-Driven Development with Context7 Integration

> Module: RED-GREEN-REFACTOR TDD cycle with Context7 patterns and AI-powered testing
> Complexity: Advanced
> Time: 25+ minutes
> Dependencies: Python 3.8+, pytest, Context7 MCP, unittest, asyncio

## Core Implementation

### TDDManager Class

```python
import pytest
import unittest
import asyncio
import subprocess
import os
import sys
import time
from typing import Dict, List, Optional, Any, Union, Callable
from dataclasses import dataclass, field
from enum import Enum
import ast
import inspect
import json
from pathlib import Path

class TDDPhase(Enum):
 """TDD cycle phases."""
 RED = "red" # Write failing test
 GREEN = "green" # Make test pass
 REFACTOR = "refactor" # Improve code while keeping tests green
 REVIEW = "review" # Review and commit changes

class TestType(Enum):
 """Types of tests in TDD."""
 UNIT = "unit"
 INTEGRATION = "integration"
 ACCEPTANCE = "acceptance"
 PERFORMANCE = "performance"
 SECURITY = "security"
 REGRESSION = "regression"

class TestStatus(Enum):
 """Test execution status."""
 PENDING = "pending"
 RUNNING = "running"
 PASSED = "passed"
 FAILED = "failed"
 SKIPPED = "skipped"
 ERROR = "error"

@dataclass
class TestSpecification:
 """Specification for a TDD test."""
 name: str
 description: str
 test_type: TestType
 requirements: List[str]
 acceptance_criteria: List[str]
 edge_cases: List[str]
 preconditions: List[str] = field(default_factory=list)
 postconditions: List[str] = field(default_factory=list)
 dependencies: List[str] = field(default_factory=list)
 mock_requirements: Dict[str, Any] = field(default_factory=dict)

@dataclass
class TestCase:
 """Individual test case with metadata."""
 id: str
 name: str
 file_path: str
 line_number: int
 specification: TestSpecification
 status: TestStatus
 execution_time: float
 error_message: Optional[str] = None
 coverage_data: Dict[str, Any] = field(default_factory=dict)

@dataclass
class TDDSession:
 """TDD development session with cycle tracking."""
 id: str
 project_path: str
 current_phase: TDDPhase
 test_cases: List[TestCase]
 start_time: float
 context7_patterns: Dict[str, Any] = field(default_factory=dict)
 metrics: Dict[str, Any] = field(default_factory=dict)

class Context7TDDIntegration:
 """Integration with Context7 for TDD patterns and best practices."""

 def __init__(self, context7_client=None):
 self.context7 = context7_client
 self.pattern_cache = {}
 self.test_generators = {}

 async def load_tdd_patterns(self, language: str = "python") -> Dict[str, Any]:
 """Load TDD patterns and best practices from Context7."""

 cache_key = f"tdd_patterns_{language}"
 if cache_key in self.pattern_cache:
 return self.pattern_cache[cache_key]

 patterns = {}

 if self.context7:
 try:
 # Load TDD best practices
 tdd_patterns = await self.context7.get_library_docs(
 context7_library_id="/testing/pytest",
 topic="TDD RED-GREEN-REFACTOR patterns best practices 2025",
 tokens=4000
 )
 patterns['tdd_best_practices'] = tdd_patterns

 # Load test patterns for specific language
 if language == "python":
 python_patterns = await self.context7.get_library_docs(
 context7_library_id="/python/pytest",
 topic="advanced testing patterns mocking fixtures 2025",
 tokens=3000
 )
 patterns['python_testing'] = python_patterns

 # Load assertion patterns
 assertion_patterns = await self.context7.get_library_docs(
 context7_library_id="/testing/assertions",
 topic="assertion patterns error messages test design 2025",
 tokens=2000
 )
 patterns['assertions'] = assertion_patterns

 # Load mocking patterns
 mocking_patterns = await self.context7.get_library_docs(
 context7_library_id="/python/unittest-mock",
 topic="mocking strategies test doubles isolation patterns 2025",
 tokens=3000
 )
 patterns['mocking'] = mocking_patterns

 except Exception as e:
 print(f"Failed to load Context7 patterns: {e}")
 patterns = self._get_default_patterns()
 else:
 patterns = self._get_default_patterns()

 self.pattern_cache[cache_key] = patterns
 return patterns

 def _get_default_patterns(self) -> Dict[str, Any]:
 """Get default TDD patterns when Context7 is unavailable."""
 return {
 'tdd_best_practices': {
 'red_phase': [
 "Write the simplest possible failing test",
 "Test one specific behavior or requirement",
 "Ensure test clearly expresses intent",
 "Make test fail for the right reason"
 ],
 'green_phase': [
 "Write the simplest code to make test pass",
 "Don't worry about code quality yet",
 "Focus on making the test green quickly",
 "Avoid premature optimization"
 ],
 'refactor_phase': [
 "Improve code design while keeping tests green",
 "Remove duplication and improve readability",
 "Apply design patterns appropriately",
 "Ensure all tests still pass"
 ]
 },
 'python_testing': {
 'pytest_features': [
 "Parametrized tests for multiple scenarios",
 "Fixtures for test setup and teardown",
 "Markers for categorizing tests",
 "Plugins for enhanced functionality"
 ],
 'assertions': [
 "Use pytest's assert statements",
 "Provide clear error messages",
 "Test expected exceptions with pytest.raises",
 "Use pytest.approx for floating point comparisons"
 ]
 },
 'assertions': {
 'best_practices': [
 "One assertion per test when possible",
 "Clear and descriptive assertion messages",
 "Test both positive and negative cases",
 "Use appropriate assertion methods"
 ]
 },
 'mocking': {
 'strategies': [
 "Mock external dependencies",
 "Use dependency injection for testability",
 "Create test doubles for complex objects",
 "Verify interactions with mocks"
 ]
 }
 }

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
 'unit_method': '''
class Test{ClassName}:
 def test_{method_name}_{scenario}(self):
 """
 Test {description}
 """
 # Arrange
 {setup_code}
 instance = {ClassName}({constructor_args})

 # Act
 result = instance.{method_name}({method_args})

 # Assert
 {assertions}
''',
 'integration_test': '''
def test_{feature_name}_{scenario}():
 """
 Test integration: {description}
 """
 # Arrange
 {setup_code}

 # Act
 result = {action}

 # Assert
 {assertions}

 # Cleanup
 {cleanup_code}
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

 if any('list' in str(req).lower() or 'array' in str(req).lower() for req in specification.requirements):
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

 # Determine appropriate template based on test type and requirements
 if spec.test_type == TestType.UNIT:
 return self._generate_unit_test(spec)
 elif spec.test_type == TestType.INTEGRATION:
 return self._generate_integration_test(spec)
 else:
 return self._generate_generic_test(spec)

 def _generate_unit_test(self, spec: TestSpecification) -> str:
 """Generate unit test code."""

 # Extract function/class information from specification name
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

 # Fill template with specification details
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

 template = self.templates['integration_test']

 feature_name = spec.name.lower().replace('test_', '').replace('_test', '')

 return template.format(
 feature_name=feature_name,
 scenario=self._extract_scenario(spec),
 description=spec.description,
 setup_code=self._generate_setup_code(spec),
 action=self._describe_action(spec),
 assertions=self._generate_assertions(spec),
 cleanup_code=self._generate_cleanup_code(spec)
 )

 def _generate_generic_test(self, spec: TestSpecification) -> str:
 """Generate generic test code."""

 test_code = f'''
def test_{spec.name.replace(' ', '_').lower()}():
 """
 Test: {spec.description}

 Requirements:
 {chr(10).join(f" - {req}" for req in spec.requirements)}

 Acceptance Criteria:
 {chr(10).join(f" - {crit}" for crit in spec.acceptance_criteria)}
 """
 # TODO: Implement test based on specification
 # This is a generated template - fill in with actual test logic

 # Arrange
 # {chr(10).join(f" # {req}" for req in spec.preconditions)}

 # Act
 # result = function_to_test()

 # Assert
 # assert result is not None
 # Add specific assertions based on acceptance criteria
'''

 return test_code

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

 def _generate_cleanup_code(self, spec: TestSpecification) -> str:
 """Generate cleanup code for integration tests."""
 cleanup_lines = []

 # Add cleanup for any resources mentioned in postconditions
 for condition in spec.postconditions:
 if 'close' in condition.lower():
 cleanup_lines.append("# Close connections")
 elif 'delete' in condition.lower():
 cleanup_lines.append("# Delete temporary resources")
 else:
 cleanup_lines.append(f"# {condition}")

 return '\n '.join(cleanup_lines) if cleanup_lines else "pass"

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

class TDDManager:
 """Main TDD workflow manager with Context7 integration."""

 def __init__(self, project_path: str, context7_client=None):
 self.project_path = Path(project_path)
 self.context7 = context7_client
 self.context7_integration = Context7TDDIntegration(context7_client)
 self.test_generator = TestGenerator(context7_client)
 self.current_session = None
 self.test_history = []

 async def start_tdd_session(
 self, feature_name: str,
 test_types: List[TestType] = None
 ) -> TDDSession:
 """Start a new TDD development session."""

 if test_types is None:
 test_types = [TestType.UNIT, TestType.INTEGRATION]

 # Load Context7 patterns
 context7_patterns = await self.context7_integration.load_tdd_patterns()

 # Create session
 session = TDDSession(
 id=f"tdd_{feature_name}_{int(time.time())}",
 project_path=str(self.project_path),
 current_phase=TDDPhase.RED,
 test_cases=[],
 start_time=time.time(),
 context7_patterns=context7_patterns,
 metrics={
 'tests_written': 0,
 'tests_passing': 0,
 'tests_failing': 0,
 'coverage_percentage': 0.0
 }
 )

 self.current_session = session
 return session

 async def write_failing_test(
 self, specification: TestSpecification,
 file_path: str = None
 ) -> str:
 """RED phase: Write a failing test based on specification."""

 if not self.current_session:
 raise ValueError("No active TDD session")

 # Generate test code
 test_code = await self.test_generator.generate_test_case(
 specification, self.current_session.context7_patterns
 )

 # Determine file path
 if file_path is None:
 file_path = self._determine_test_file_path(specification)

 # Write test file
 full_path = self.project_path / file_path
 full_path.parent.mkdir(parents=True, exist_ok=True)

 with open(full_path, 'a', encoding='utf-8') as f:
 f.write('\n\n' + test_code)

 # Create test case object
 test_case = TestCase(
 id=f"test_{int(time.time())}",
 name=specification.name,
 file_path=str(full_path),
 line_number=self._count_lines_in_file(str(full_path)),
 specification=specification,
 status=TestStatus.PENDING,
 execution_time=0.0
 )

 self.current_session.test_cases.append(test_case)
 self.current_session.metrics['tests_written'] += 1

 return str(full_path)

 async def run_tests_and_verify_failure(self) -> Dict[str, Any]:
 """RED phase: Run tests and verify they fail as expected."""

 if not self.current_session:
 raise ValueError("No active TDD session")

 # Run tests using pytest
 test_results = await self._run_pytest()

 # Update session metrics
 failing_tests = test_results.get('failures', 0)
 self.current_session.metrics['tests_failing'] = failing_tests

 # Verify at least one test is failing (RED phase requirement)
 if failing_tests == 0:
 raise ValueError("RED phase violation: No tests are failing. Write a proper failing test first.")

 return test_results

 async def implement_minimum_code_to_pass(self, target_function: str) -> Dict[str, Any]:
 """GREEN phase: Implement minimum code to make tests pass."""

 if not self.current_session:
 raise ValueError("No active TDD session")

 # Generate minimum implementation
 implementation = await self._generate_minimum_implementation(target_function)

 # Write implementation
 impl_path = self._determine_implementation_file_path(target_function)
 full_path = self.project_path / impl_path
 full_path.parent.mkdir(parents=True, exist_ok=True)

 with open(full_path, 'w', encoding='utf-8') as f:
 f.write(implementation)

 # Run tests to verify they pass
 test_results = await self._run_pytest()

 # Update session metrics
 passing_tests = test_results.get('passed', 0)
 self.current_session.metrics['tests_passing'] = passing_tests

 if test_results.get('failures', 0) > 0:
 raise ValueError("GREEN phase violation: Tests are still failing. Implementation needs adjustment.")

 self.current_session.current_phase = TDDPhase.GREEN
 return test_results

 async def refactor_while_maintaining_tests(self) -> Dict[str, Any]:
 """REFACTOR phase: Improve code while keeping tests green."""

 if not self.current_session:
 raise ValueError("No active TDD session")

 if self.current_session.current_phase != TDDPhase.GREEN:
 raise ValueError("Must complete GREEN phase before REFACTOR")

 # Get Context7 refactoring patterns
 refactoring_patterns = await self._get_refactoring_patterns()

 # Generate refactoring suggestions
 suggestions = await self._generate_refactoring_suggestions(refactoring_patterns)

 # Apply refactoring suggestions
 for suggestion in suggestions:
 await self._apply_refactoring_suggestion(suggestion)

 # Verify tests still pass after each refactoring
 test_results = await self._run_pytest()
 if test_results.get('failures', 0) > 0:
 raise ValueError(f"REFACTOR phase violation: Tests failed after applying suggestion: {suggestion['description']}")

 self.current_session.current_phase = TDDPhase.REFACTOR
 return suggestions

 async def run_full_tdd_cycle(
 self, specification: TestSpecification,
 target_function: str = None
 ) -> Dict[str, Any]:
 """Run complete RED-GREEN-REFACTOR TDD cycle."""

 cycle_results = {}

 # RED phase
 print(" RED Phase: Writing failing test...")
 test_file = await self.write_failing_test(specification)
 red_results = await self.run_tests_and_verify_failure()
 cycle_results['red'] = {'test_file': test_file, 'results': red_results}

 # GREEN phase
 print("🟢 GREEN Phase: Implementing minimum code...")
 if target_function is None:
 target_function = specification.name.replace('test_', '').replace('_test', '')
 green_results = await self.implement_minimum_code_to_pass(target_function)
 cycle_results['green'] = {'results': green_results}

 # REFACTOR phase
 print(" REFACTOR Phase: Improving code quality...")
 refactor_results = await self.refactor_while_maintaining_tests()
 cycle_results['refactor'] = {'suggestions': refactor_results}

 # REVIEW phase
 print("🟣 REVIEW Phase: Final verification...")
 coverage_results = await self._run_coverage_analysis()
 cycle_results['review'] = {'coverage': coverage_results}

 self.current_session.current_phase = TDDPhase.REVIEW
 return cycle_results

 async def _run_pytest(self) -> Dict[str, Any]:
 """Run pytest and return results."""

 try:
 # Run pytest with JSON output for parsing
 result = subprocess.run(
 [
 sys.executable, '-m', 'pytest',
 str(self.project_path),
 '--tb=short',
 '--json-report',
 '--json-report-file=/tmp/pytest_results.json'
 ],
 capture_output=True,
 text=True,
 cwd=str(self.project_path)
 )

 # Parse results
 try:
 with open('/tmp/pytest_results.json', 'r') as f:
 results_data = json.load(f)

 return {
 'summary': results_data.get('summary', {}),
 'tests': results_data.get('tests', []),
 'passed': results_data.get('summary', {}).get('passed', 0),
 'failed': results_data.get('summary', {}).get('failed', 0),
 'skipped': results_data.get('summary', {}).get('skipped', 0),
 'total': results_data.get('summary', {}).get('total', 0),
 'duration': results_data.get('summary', {}).get('duration', 0)
 }
 except (FileNotFoundError, json.JSONDecodeError):
 # Fallback to parsing stdout
 return self._parse_pytest_output(result.stdout)

 except Exception as e:
 print(f"Error running pytest: {e}")
 return {'error': str(e), 'passed': 0, 'failed': 0, 'skipped': 0, 'total': 0}

 def _parse_pytest_output(self, output: str) -> Dict[str, Any]:
 """Parse pytest output when JSON is not available."""

 lines = output.split('\n')
 results = {'passed': 0, 'failed': 0, 'skipped': 0, 'total': 0, 'duration': 0}

 for line in lines:
 if ' passed in ' in line:
 parts = line.split()
 if parts and parts[0].isdigit():
 results['passed'] = int(parts[0])
 results['total'] = int(parts[0])
 elif ' passed' in line and ' failed' in line:
 # Parse format like "2 passed, 1 failed"
 passed_part = line.split(' passed')[0]
 if passed_part.strip().isdigit():
 results['passed'] = int(passed_part.strip())

 if ' failed' in line:
 failed_part = line.split(' failed')[0].split(', ')[-1]
 if failed_part.strip().isdigit():
 results['failed'] = int(failed_part.strip())

 results['total'] = results['passed'] + results['failed']

 return results

 def _determine_test_file_path(self, specification: TestSpecification) -> str:
 """Determine appropriate test file path based on specification."""

 # Create tests directory structure
 test_dir = "tests"

 # Determine subdirectory based on test type
 if specification.test_type == TestType.INTEGRATION:
 test_dir += "/integration"
 elif specification.test_type == TestType.ACCEPTANCE:
 test_dir += "/acceptance"
 elif specification.test_type == TestType.PERFORMANCE:
 test_dir += "/performance"

 # Generate filename
 feature_name = specification.name.replace('test_', '').replace('_test', '')
 filename = f"test_{feature_name}.py"

 return f"{test_dir}/{filename}"

 def _determine_implementation_file_path(self, function_name: str) -> str:
 """Determine implementation file path for target function."""

 # Simple heuristic - place in src directory
 # In real implementation, this would be more sophisticated
 return f"src/{function_name}.py"

 def _count_lines_in_file(self, file_path: str) -> int:
 """Count lines in a file."""
 try:
 with open(file_path, 'r', encoding='utf-8') as f:
 return len(f.readlines())
 except FileNotFoundError:
 return 0

 async def _generate_minimum_implementation(self, target_function: str) -> str:
 """Generate minimum implementation to make tests pass."""

 # Get Context7 patterns for implementation
 patterns = self.current_session.context7_patterns

 # Simple implementation template
 implementation = f'''
def {target_function}(*args, kwargs):
 """
 Minimum implementation for {target_function}.

 This is a TDD-generated implementation that will be
 refined during the REFACTOR phase.
 """
 # TODO: Implement proper logic based on test requirements
 # This is a placeholder to make tests pass initially

 if args:
 return args[0] # Return first argument as default

 return None # Default return value
'''

 return implementation

 async def _get_refactoring_patterns(self) -> Dict[str, Any]:
 """Get refactoring patterns from Context7."""

 if self.context7:
 try:
 patterns = await self.context7.get_library_docs(
 context7_library_id="/refactoring/guru",
 topic="code refactoring patterns clean code 2025",
 tokens=3000
 )
 return patterns
 except Exception as e:
 print(f"Failed to get refactoring patterns: {e}")

 return {}

 async def _generate_refactoring_suggestions(self, patterns: Dict[str, Any]) -> List[Dict[str, Any]]:
 """Generate refactoring suggestions based on current implementation."""

 suggestions = []

 # Add general refactoring suggestions
 suggestions.extend([
 {
 'type': 'extract_method',
 'description': 'Extract complex logic into separate methods',
 'priority': 'medium',
 'applicable_to': ['large_functions']
 },
 {
 'type': 'improve_naming',
 'description': 'Improve variable and function naming for clarity',
 'priority': 'high',
 'applicable_to': ['all_functions']
 },
 {
 'type': 'add_documentation',
 'description': 'Add comprehensive docstrings and comments',
 'priority': 'medium',
 'applicable_to': ['all_functions']
 },
 {
 'type': 'handle_edge_cases',
 'description': 'Add proper error handling and edge case management',
 'priority': 'high',
 'applicable_to': ['public_functions']
 }
 ])

 return suggestions

 async def _apply_refactoring_suggestion(self, suggestion: Dict[str, Any]) -> bool:
 """Apply a refactoring suggestion to the codebase."""

 # In a real implementation, this would analyze the code
 # and apply automated refactoring using tools like Rope

 print(f"Applying refactoring: {suggestion['description']}")

 # Simulate refactoring application
 await asyncio.sleep(0.1) # Simulate work

 return True

 async def _run_coverage_analysis(self) -> Dict[str, Any]:
 """Run test coverage analysis."""

 try:
 # Run coverage analysis
 result = subprocess.run(
 [
 sys.executable, '-m', 'pytest',
 str(self.project_path),
 '--cov=src',
 '--cov-report=json',
 '--cov-report=term-missing'
 ],
 capture_output=True,
 text=True,
 cwd=str(self.project_path)
 )

 # Parse coverage results
 try:
 # Look for coverage file
 import glob
 coverage_files = glob.glob(str(self.project_path) + '/coverage.json')
 if coverage_files:
 with open(coverage_files[0], 'r') as f:
 coverage_data = json.load(f)

 total_coverage = coverage_data.get('totals', {}).get('percent_covered', 0)
 self.current_session.metrics['coverage_percentage'] = total_coverage

 return {
 'total_coverage': total_coverage,
 'files': coverage_data.get('files', {}),
 'totals': coverage_data.get('totals', {})
 }
 except Exception:
 pass

 return {'total_coverage': 0, 'message': 'Coverage analysis failed'}

 except Exception as e:
 return {'error': str(e), 'total_coverage': 0}

 def get_session_summary(self) -> Dict[str, Any]:
 """Get summary of current TDD session."""

 if not self.current_session:
 return {}

 duration = time.time() - self.current_session.start_time

 return {
 'session_id': self.current_session.id,
 'phase': self.current_session.current_phase.value,
 'duration_seconds': duration,
 'duration_formatted': f"{duration:.1f} seconds",
 'metrics': self.current_session.metrics,
 'test_cases_count': len(self.current_session.test_cases)
 }

# Usage Examples
"""
# Initialize TDD Manager
tdd_manager = TDDManager(
 project_path="/path/to/project",
 context7_client=context7
)

# Start TDD session
session = await tdd_manager.start_tdd_session("user_authentication")

# Create test specification
test_spec = TestSpecification(
 name="test_user_login_valid_credentials",
 description="Test that user can login with valid credentials",
 test_type=TestType.UNIT,
 requirements=[
 "User must provide valid email and password",
 "System should authenticate user credentials",
 "Successful login should return user token"
 ],
 acceptance_criteria=[
 "Valid credentials return user token",
 "Invalid credentials raise AuthenticationError",
 "Empty credentials raise ValidationError"
 ],
 edge_cases=[
 "Test with empty email",
 "Test with empty password",
 "Test with malformed email",
 "Test with very long password"
 ],
 mock_requirements={
 "user_database": {"return_value": "mock_user"},
 "token_generator": {"return_value": "mock_token"}
 }
)

# Run complete TDD cycle
cycle_results = await tdd_manager.run_full_tdd_cycle(
 specification=test_spec,
 target_function="authenticate_user"
)

print("TDD Cycle Results:")
print(f" RED phase: Test written to {cycle_results['red']['test_file']}")
print(f" GREEN phase: {cycle_results['green']['results']['passed']} tests passing")
print(f" REFACTOR phase: {len(cycle_results['refactor']['suggestions'])} improvements made")
print(f" REVIEW phase: {cycle_results['review']['coverage']['total_coverage']:.1f}% coverage")

# Get session summary
summary = tdd_manager.get_session_summary()
print(f"Session completed in {summary['duration_formatted']}")
print(f"Final phase: {summary['phase']}")
"""
```

## Advanced Features

### Context7-Enhanced Test Generation

AI-Powered Test Case Generation:
```python
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
 'decorators': [d.id if isinstance(d, ast.Name) else str(d) for d in node.decorator_list]
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
```

## Best Practices

1. RED Phase: Always write failing tests first, ensure they fail for the right reason
2. GREEN Phase: Write minimum code to pass tests, avoid premature optimization
3. REFACTOR Phase: Improve design while keeping tests green, apply small incremental changes
4. Test Coverage: Aim for high coverage but focus on meaningful tests
5. Context Integration: Leverage Context7 patterns for industry-standard testing approaches

---

Module: `modules/tdd-context7.md`
Related: [AI Debugging](./ai-debugging.md) | [Performance Optimization](./performance-optimization.md)
