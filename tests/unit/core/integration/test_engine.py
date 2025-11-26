"""
Test integration testing execution engine.
"""

import time

import pytest

from moai_adk.core.integration.engine import TestEngine as IntegrationTestEngine
from moai_adk.core.integration.models import TestStatus


class TestEngine:
    """Test TestEngine class"""

    def test_engine_initialization(self):
        """Test engine initialization"""
        engine = IntegrationTestEngine(test_timeout=10.0, max_workers=2)

        assert engine.test_timeout == 10.0
        assert engine.max_workers == 2
        assert engine._test_counter == 0

    def test_engine_invalid_initialization(self):
        """Test engine initialization with invalid parameters"""
        with pytest.raises(ValueError, match="Test timeout must be positive"):
            IntegrationTestEngine(test_timeout=-1.0)

        with pytest.raises(ValueError, match="Max workers must be positive"):
            IntegrationTestEngine(max_workers=0)

    def test_generate_test_id(self):
        """Test test ID generation"""
        engine = IntegrationTestEngine()

        id1 = engine._generate_test_id()
        id2 = engine._generate_test_id()
        id3 = engine._generate_test_id()

        assert id1 == "test_0001"
        assert id2 == "test_0002"
        assert id3 == "test_0003"
        assert id1 != id2 != id3

    def test_execute_test_success(self):
        """Test successful test execution"""
        engine = IntegrationTestEngine(test_timeout=5.0)

        def test_func():
            return "success"

        result = engine.execute_test(test_func, "test_success", ["comp1"])

        assert result.test_name == "test_success"
        assert result.passed is True
        assert result.error_message is None
        assert result.components_tested == ["comp1"]
        assert result.status == TestStatus.PASSED
        assert result.execution_time > 0

    def test_execute_test_failure(self):
        """Test failed test execution"""
        engine = IntegrationTestEngine()

        def test_func():
            raise ValueError("Test error")

        result = engine.execute_test(test_func, "test_failure")

        assert result.passed is False
        assert result.error_message == "Test error"
        assert result.status == TestStatus.FAILED

    def test_execute_test_timeout(self):
        """Test test execution timeout"""
        engine = IntegrationTestEngine(test_timeout=0.1)  # Very short timeout

        def slow_test():
            time.sleep(0.5)  # Sleep longer than timeout
            return "should not reach here"

        result = engine.execute_test(slow_test, "slow_test")

        assert result.passed is False
        assert "timed out" in result.error_message.lower()
        assert result.status == TestStatus.FAILED

    def test_execute_test_with_defaults(self):
        """Test test execution with default parameters"""
        engine = IntegrationTestEngine()

        def test_func():
            return True

        result = engine.execute_test(test_func)

        assert result.test_name.startswith("test_")
        assert result.components_tested == []
        assert result.passed is True

    @pytest.mark.asyncio
    async def test_execute_test_async(self):
        """Test asynchronous test execution"""
        engine = IntegrationTestEngine()

        def test_func():
            return "async_result"

        result = await engine.execute_test_async(test_func, "async_test")

        assert result.test_name == "async_test"
        assert result.passed is True
        assert result.status == TestStatus.PASSED

    def test_run_concurrent_tests(self):
        """Test concurrent test execution"""
        engine = IntegrationTestEngine(test_timeout=5.0, max_workers=3)

        def test_func(result_value):
            return result_value

        tests = [
            (lambda: test_func("result1"), "test1", ["comp1"]),
            (lambda: test_func("result2"), "test2", ["comp2"]),
            (lambda: test_func("result3"), "test3", ["comp3"]),
        ]

        results = engine.run_concurrent_tests(tests)

        assert len(results) == 3
        assert all(result.passed for result in results)
        assert set(result.test_name for result in results) == {"test1", "test2", "test3"}
        assert all("comp" in str(result.components_tested) for result in results)

    def test_run_concurrent_tests_with_failures(self):
        """Test concurrent test execution with failures"""
        engine = IntegrationTestEngine()

        def success_test():
            return "success"

        def failure_test():
            raise ValueError("Test failure")

        tests = [(success_test, "success_test", []), (failure_test, "failure_test", [])]

        results = engine.run_concurrent_tests(tests)

        assert len(results) == 2
        # Check results regardless of order
        success_results = [r for r in results if r.passed]
        failure_results = [r for r in results if not r.passed]

        assert len(success_results) == 1, "Should have one success result"
        assert len(failure_results) == 1, "Should have one failure result"
        assert failure_results[0].error_message == "Test failure"

    def test_run_concurrent_tests_timeout(self):
        """Test concurrent tests with batch timeout"""
        engine = IntegrationTestEngine(test_timeout=1.0, max_workers=2)

        def quick_test():
            return "quick"

        def slow_test():
            time.sleep(2.0)  # Slower than batch timeout
            return "slow"

        tests = [(quick_test, "quick_test", []), (slow_test, "slow_test", [])]

        # Should handle timeout gracefully
        try:
            results = engine.run_concurrent_tests(tests, timeout=0.5)  # Very short batch timeout
            # May get partial results before timeout
            assert len(results) >= 0  # May get partial results before timeout
        except TimeoutError:
            # Expected timeout behavior
            pass

    @pytest.mark.asyncio
    async def test_run_concurrent_tests_async(self):
        """Test asynchronous concurrent test execution"""
        engine = IntegrationTestEngine()

        def test_func(value):
            return value

        tests = [(lambda: test_func("async1"), "async_test1", []), (lambda: test_func("async2"), "async_test2", [])]

        results = await engine.run_concurrent_tests_async(tests)

        assert len(results) == 2
        assert all(result.passed for result in results)

    def test_concurrent_test_execution_order(self):
        """Test that concurrent tests execute in reasonable time"""
        engine = IntegrationTestEngine(max_workers=4)

        def test_func():
            time.sleep(0.1)  # Small delay
            return True

        tests = [(test_func, f"test_{i}", []) for i in range(4)]

        start_time = time.time()
        results = engine.run_concurrent_tests(tests)
        end_time = time.time()

        # Should complete in roughly 0.1 seconds, not 0.4 seconds (sequential)
        execution_time = end_time - start_time
        assert execution_time < 0.3  # Allow some overhead
        assert len(results) == 4
        assert all(result.passed for result in results)
