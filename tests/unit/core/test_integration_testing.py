"""
Integration Testing Framework Tests

Test cases for integration testing functionality.
"""

import tempfile
from pathlib import Path

import pytest

from moai_adk.core.integration import IntegrationTester, IntegrationTestResult, TestComponent, TestSuite


class TestIntegrationTester:
    """Test suite for integration tester functionality."""

    def test_integration_tester_creation(self):
        """Test that integration tester can be created successfully."""
        tester = IntegrationTester()

        assert tester is not None
        assert tester.engine.test_timeout == 30.0
        assert tester.engine.max_workers == 4
        assert len(tester.test_results) == 0

    def test_integration_tester_with_custom_params(self):
        """Test integration tester with custom parameters."""
        tester = IntegrationTester(test_timeout=60.0, max_workers=8)

        assert tester.engine.test_timeout == 60.0
        assert tester.engine.max_workers == 8
        assert len(tester.test_results) == 0

    def test_add_test_result(self):
        """Test adding test results."""
        tester = IntegrationTester()
        result = IntegrationTestResult("test1", True)

        tester.add_test_result(result)

        assert len(tester.test_results) == 1
        assert tester.test_results[0] == result

    def test_clear_results(self):
        """Test clearing test results."""
        tester = IntegrationTester()
        tester.add_test_result(IntegrationTestResult("test1", True))
        tester.add_test_result(IntegrationTestResult("test2", False))

        assert len(tester.test_results) == 2

        tester.clear_results()

        assert len(tester.test_results) == 0

    def test_get_success_rate(self):
        """Test success rate calculation."""
        tester = IntegrationTester()

        # Empty results
        assert tester.get_success_rate() == 0.0

        # Add some results
        tester.add_test_result(IntegrationTestResult("test1", True))
        tester.add_test_result(IntegrationTestResult("test2", False))
        tester.add_test_result(IntegrationTestResult("test3", True))

        # 2 out of 3 passed = 66.67%
        success_rate = tester.get_success_rate()
        assert abs(success_rate - 66.67) < 0.1

    def test_get_test_stats(self):
        """Test getting test statistics."""
        tester = IntegrationTester()

        # Add test results with different execution times
        tester.add_test_result(IntegrationTestResult("test1", True, execution_time=1.0))
        tester.add_test_result(IntegrationTestResult("test2", False, execution_time=2.0))
        tester.add_test_result(IntegrationTestResult("test3", True, execution_time=3.0))

        stats = tester.get_test_stats()

        assert stats["total"] == 3
        assert stats["passed"] == 2
        assert stats["failed"] == 1
        assert stats["success_rate"] == 66.66666666666666
        assert stats["total_time"] == 6.0
        assert stats["avg_time"] == 2.0

    def test_run_test_success(self):
        """Test running a successful test."""
        tester = IntegrationTester()

        def test_func():
            return "success"

        result = tester.run_test(test_func, "test_success", ["comp1"])

        assert result.passed is True
        assert result.test_name == "test_success"
        assert result.components_tested == ["comp1"]
        assert len(tester.test_results) == 1

    def test_run_test_failure(self):
        """Test running a failing test."""
        tester = IntegrationTester()

        def test_func():
            raise ValueError("Test error")

        result = tester.run_test(test_func, "test_failure")

        assert result.passed is False
        assert result.error_message == "Test error"
        assert len(tester.test_results) == 1

    @pytest.mark.asyncio
    async def test_run_test_async(self):
        """Test running a test asynchronously."""
        tester = IntegrationTester()

        def test_func():
            return "async_result"

        result = await tester.run_test_async(test_func, "async_test", ["comp1"])

        assert result.passed is True
        assert result.test_name == "async_test"
        assert result.components_tested == ["comp1"]
        assert len(tester.test_results) == 1

    def test_run_test_suite(self):
        """Test running a test suite."""
        tester = IntegrationTester()

        # Create test suite
        comp1 = TestComponent("comp1", "type1", "1.0.0")
        comp2 = TestComponent("comp2", "type2", "2.0.0")
        suite = TestSuite(
            name="test_suite", description="Test suite", components=[comp1, comp2], test_cases=["test1", "test2"]
        )

        results = tester.run_test_suite(suite)

        assert len(results) == 2
        assert len(tester.test_results) == 2
        assert all(result.passed for result in results)  # All tests return True

    def test_run_concurrent_tests(self):
        """Test running multiple tests concurrently."""
        tester = IntegrationTester()

        def test_func(value):
            return value

        tests = [
            (lambda: test_func("result1"), "test1", ["comp1"]),
            (lambda: test_func("result2"), "test2", ["comp2"]),
            (lambda: test_func("result3"), "test3", ["comp3"]),
        ]

        results = tester.run_concurrent_tests(tests)

        assert len(results) == 3
        assert len(tester.test_results) == 3
        assert all(result.passed for result in results)

    @pytest.mark.asyncio
    async def test_run_concurrent_tests_async(self):
        """Test running multiple tests concurrently asynchronously."""
        tester = IntegrationTester()

        def test_func(value):
            return value

        tests = [(lambda: test_func("async1"), "async_test1", []), (lambda: test_func("async2"), "async_test2", [])]

        results = await tester.run_concurrent_tests_async(tests)

        assert len(results) == 2
        assert len(tester.test_results) == 2
        assert all(result.passed for result in results)

    def test_discover_components(self):
        """Test component discovery."""
        tester = IntegrationTester()

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create some Python files
            Path(temp_dir, "module1.py").touch()
            Path(temp_dir, "module2.py").touch()

            components = tester.discover_components(temp_dir)

            assert len(components) == 2
            component_names = [c.name for c in components]
            assert "module1" in component_names
            assert "module2" in component_names

    def test_create_test_environment(self):
        """Test creating test environment."""
        tester = IntegrationTester()

        env = tester.create_test_environment()

        assert env is not None
        assert env.temp_dir is not None
        assert Path(env.temp_dir).exists()

        # Cleanup
        env.cleanup()

    def test_create_test_environment_with_custom_dir(self):
        """Test creating test environment with custom directory."""
        tester = IntegrationTester()

        with tempfile.TemporaryDirectory() as custom_dir:
            env = tester.create_test_environment(custom_dir)

            assert env.temp_dir == custom_dir
            assert not env.created_temp  # Should not create new dir

            env.cleanup()  # Should not remove custom dir
            assert Path(custom_dir).exists()

    def test_export_results_dict(self):
        """Test exporting results as dictionary."""
        tester = IntegrationTester()

        # Add test results
        tester.add_test_result(IntegrationTestResult("test1", True, execution_time=1.0))
        tester.add_test_result(IntegrationTestResult("test2", False, error_message="Error"))

        exported = tester.export_results("dict")

        assert len(exported) == 2
        assert exported[0]["test_name"] == "test1"
        assert exported[0]["passed"] is True
        assert exported[1]["test_name"] == "test2"
        assert exported[1]["passed"] is False

    def test_export_results_summary(self):
        """Test exporting results as summary."""
        tester = IntegrationTester()

        # Add test results
        tester.add_test_result(IntegrationTestResult("test1", True))
        tester.add_test_result(IntegrationTestResult("test2", False, error_message="Error"))

        exported = tester.export_results("summary")

        assert "stats" in exported
        assert "failed_tests" in exported
        assert exported["stats"]["total"] == 2
        assert exported["stats"]["passed"] == 1
        assert exported["stats"]["failed"] == 1
        assert len(exported["failed_tests"]) == 1

    def test_export_results_invalid_format(self):
        """Test exporting results with invalid format."""
        tester = IntegrationTester()

        with pytest.raises(ValueError, match="Unsupported format"):
            tester.export_results("invalid_format")

    def test_validate_test_environment_empty(self):
        """Test validating empty test environment."""
        tester = IntegrationTester()
        warnings = tester.validate_test_environment()

        assert len(warnings) == 1
        assert "No test results found" in warnings[0]

    def test_validate_test_environment_low_success_rate(self):
        """Test validating environment with low success rate."""
        tester = IntegrationTester()

        # Add failing tests to get low success rate
        tester.add_test_result(IntegrationTestResult("test1", False))
        tester.add_test_result(IntegrationTestResult("test2", False))
        tester.add_test_result(IntegrationTestResult("test3", True))  # Only 33% success rate

        warnings = tester.validate_test_environment()

        assert len(warnings) == 1
        assert "Low success rate" in warnings[0]

    def test_validate_test_environment_good(self):
        """Test validating good test environment."""
        tester = IntegrationTester()

        # Add mostly successful tests
        tester.add_test_result(IntegrationTestResult("test1", True))
        tester.add_test_result(IntegrationTestResult("test2", True))
        tester.add_test_result(IntegrationTestResult("test3", True))
        tester.add_test_result(IntegrationTestResult("test4", False))  # 75% success rate

        warnings = tester.validate_test_environment()

        assert len(warnings) == 0  # No warnings for good environment
