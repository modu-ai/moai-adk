"""
Integration testing execution engine.

This module contains the core test execution logic for running
integration tests with timeout handling and concurrent execution.
"""

import asyncio
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from typing import Callable, List, Optional

from .models import IntegrationTestResult, TestStatus


class TestEngine:
    """
    Core test execution engine for integration testing.

    Handles test execution, timeout management, and concurrent execution.
    """

    def __init__(self, test_timeout: float = 30.0, max_workers: int = 4):
        """
        Initialize the test engine.

        Args:
            test_timeout: Maximum time (in seconds) for each test
            max_workers: Maximum number of concurrent workers
        """
        if test_timeout <= 0:
            raise ValueError("Test timeout must be positive")
        if max_workers <= 0:
            raise ValueError("Max workers must be positive")

        self.test_timeout = test_timeout
        self.max_workers = max_workers
        self._test_counter = 0

    def _generate_test_id(self) -> str:
        """Generate a unique test ID."""
        self._test_counter += 1
        return f"test_{self._test_counter:04d}"

    def execute_test(
        self, test_func: Callable, test_name: str = None, components: List[str] = None
    ) -> IntegrationTestResult:
        """
        Execute a single integration test.

        Args:
            test_func: Test function to execute
            test_name: Optional test name
            components: List of components being tested

        Returns:
            IntegrationTestResult: Test execution result
        """
        if test_name is None:
            test_name = self._generate_test_id()

        if components is None:
            components = []

        start_time = time.time()
        result = IntegrationTestResult(test_name=test_name, passed=False, components_tested=components)

        try:
            # Execute test with timeout
            with ThreadPoolExecutor(max_workers=1) as executor:
                future = executor.submit(test_func)
                try:
                    test_result = future.result(timeout=self.test_timeout)
                    result.passed = bool(test_result)
                    result.status = TestStatus.PASSED if result.passed else TestStatus.FAILED

                except TimeoutError:
                    result.error_message = f"Test timed out after {self.test_timeout}s"
                    result.status = TestStatus.FAILED

        except Exception as e:
            result.error_message = str(e)
            result.status = TestStatus.FAILED

        finally:
            result.execution_time = time.time() - start_time

        return result

    async def execute_test_async(
        self, test_func: Callable, test_name: str = None, components: List[str] = None
    ) -> IntegrationTestResult:
        """
        Execute a single integration test asynchronously.

        Args:
            test_func: Test function to execute
            test_name: Optional test name
            components: List of components being tested

        Returns:
            IntegrationTestResult: Test execution result
        """
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(None, self.execute_test, test_func, test_name, components)

    def run_concurrent_tests(self, tests: List[tuple], timeout: Optional[float] = None) -> List[IntegrationTestResult]:
        """
        Run multiple tests concurrently.

        Args:
            tests: List of (test_func, test_name, components) tuples
            timeout: Optional timeout for entire batch

        Returns:
            List of test results
        """
        if timeout is None:
            timeout = self.test_timeout * 2

        results = []

        with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
            # Submit all tests
            future_to_test = {executor.submit(self.execute_test, *test): test for test in tests}

            # Collect results as they complete
            for future in as_completed(future_to_test, timeout=timeout):
                try:
                    result = future.result()
                    results.append(result)
                except Exception as e:
                    test_info = future_to_test[future]
                    error_result = IntegrationTestResult(
                        test_name=test_info[1] if len(test_info) > 1 else "unknown",
                        passed=False,
                        error_message=f"Execution error: {str(e)}",
                    )
                    results.append(error_result)

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
        loop = asyncio.get_event_loop()
        return await loop.run_in_executor(None, self.run_concurrent_tests, tests, timeout)
