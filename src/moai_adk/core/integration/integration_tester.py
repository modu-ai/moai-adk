"""
Integration Tester - Main Interface

Provides a high-level interface for comprehensive integration testing
of MoAI-ADK components. This module serves as the main entry point
for integration testing functionality.
"""

from typing import Any, Callable, Dict, List, Optional, Union

from .engine import TestEngine
from .models import IntegrationTestResult, TestComponent, TestSuite
from .utils import ComponentDiscovery, TestEnvironment, TestResultAnalyzer


class IntegrationTester:
    """
    Comprehensive integration tester for MoAI-ADK components.

    This class provides a high-level interface for testing multiple components
    together to ensure they work correctly in integration scenarios.
    """

    def __init__(self, test_timeout: float = 30.0, max_workers: int = 4):
        """
        Initialize the integration tester.

        Args:
            test_timeout: Maximum time (in seconds) for each test
            max_workers: Maximum number of concurrent workers
        """
        self.engine = TestEngine(test_timeout, max_workers)
        self.test_results: List[IntegrationTestResult] = []
        self.discovery = ComponentDiscovery()
        self.analyzer = TestResultAnalyzer()

    def add_test_result(self, result: IntegrationTestResult):
        """
        Add a test result to the results list.

        Args:
            result: Test result to add
        """
        self.test_results.append(result)

    def clear_results(self):
        """Clear all test results."""
        self.test_results.clear()

    def get_success_rate(self) -> float:
        """
        Get the success rate of all tests.

        Returns:
            Success rate as percentage (0-100)
        """
        return self.analyzer.calculate_success_rate(self.test_results)

    def get_test_stats(self) -> Dict[str, Any]:
        """
        Get comprehensive test statistics.

        Returns:
            Dictionary with test statistics
        """
        return self.analyzer.get_execution_stats(self.test_results)

    def run_test(
        self, test_func: Callable, test_name: str = None, components: List[str] = None
    ) -> IntegrationTestResult:
        """
        Run a single integration test.

        Args:
            test_func: Test function to execute
            test_name: Optional test name
            components: List of components being tested

        Returns:
            IntegrationTestResult: Test execution result
        """
        result = self.engine.execute_test(test_func, test_name, components)
        self.add_test_result(result)
        return result

    async def run_test_async(
        self, test_func: Callable, test_name: str = None, components: List[str] = None
    ) -> IntegrationTestResult:
        """
        Run a single integration test asynchronously.

        Args:
            test_func: Test function to execute
            test_name: Optional test name
            components: List of components being tested

        Returns:
            IntegrationTestResult: Test execution result
        """
        result = await self.engine.execute_test_async(test_func, test_name, components)
        self.add_test_result(result)
        return result

    def run_test_suite(self, test_suite: TestSuite) -> List[IntegrationTestResult]:
        """
        Run a complete test suite.

        Args:
            test_suite: Test suite to run

        Returns:
            List of test results
        """
        results = []

        def placeholder_test():
            """Placeholder test function"""
            return True

        for test_case_name in test_suite.test_cases:
            # This is a simplified implementation
            # In practice, you would map test case names to actual test functions
            result = self.run_test(
                placeholder_test,
                test_case_name,
                [c.name for c in test_suite.components],
            )
            results.append(result)

        return results

    def run_concurrent_tests(self, tests: List[tuple], timeout: Optional[float] = None) -> List[IntegrationTestResult]:
        """
        Run multiple tests concurrently.

        Args:
            tests: List of (test_func, test_name, components) tuples
            timeout: Optional timeout for entire batch

        Returns:
            List of test results
        """
        results = self.engine.run_concurrent_tests(tests, timeout)
        self.test_results.extend(results)
        return results

    async def run_concurrent_tests_async(
        self, tests: List[tuple], timeout: Optional[float] = None
    ) -> List[IntegrationTestResult]:
        """
        Run multiple tests concurrently asynchronously.

        Args:
            tests: List of (test_func, test_name, components) tuples
            timeout: Optional timeout for entire batch

        Returns:
            List of test results
        """
        results = await self.engine.run_concurrent_tests_async(tests, timeout)
        self.test_results.extend(results)
        return results

    def discover_components(self, base_path: str) -> List[TestComponent]:
        """
        Discover testable components in the given path.

        Args:
            base_path: Base path to search for components

        Returns:
            List of discovered components
        """
        return self.discovery.discover_components(base_path)

    def create_test_environment(self, temp_dir: Optional[str] = None) -> TestEnvironment:
        """
        Create a test environment for integration testing.

        Args:
            temp_dir: Optional temporary directory path

        Returns:
            TestEnvironment instance
        """
        return TestEnvironment(temp_dir)

    def export_results(self, format: str = "dict") -> Union[Dict, str]:
        """
        Export test results in specified format.

        Args:
            format: Export format ("dict", "json", "summary")

        Returns:
            Exported results
        """
        if format == "dict":
            return [vars(result) for result in self.test_results]  # type: ignore[return-value]
        elif format == "summary":
            return {
                "stats": self.get_test_stats(),
                "failed_tests": self.analyzer.get_failed_tests(self.test_results),
            }
        else:
            raise ValueError(f"Unsupported format: {format}")

    def validate_test_environment(self) -> List[str]:
        """
        Validate the test environment.

        Returns:
            List of validation warnings
        """
        warnings = []

        # Check if we have any test results
        if not self.test_results:
            warnings.append("No test results found")

        # Check success rate
        success_rate = self.get_success_rate()
        if success_rate < 80.0:
            warnings.append(f"Low success rate: {success_rate:.1f}%")

        return warnings
