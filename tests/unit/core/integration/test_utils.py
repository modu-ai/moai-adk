"""
Test integration testing utilities.
"""

import tempfile
from pathlib import Path
from unittest.mock import patch

from moai_adk.core.integration.models import IntegrationTestResult, TestComponent
from moai_adk.core.integration.utils import ComponentDiscovery, TestEnvironment
from moai_adk.core.integration.utils import TestResultAnalyzer as IntegrationTestResultAnalyzer


class TestComponentDiscovery:
    """Test ComponentDiscovery class"""

    def test_discover_components_empty_dir(self):
        """Test component discovery in empty directory"""
        with tempfile.TemporaryDirectory() as temp_dir:
            discovery = ComponentDiscovery()
            components = discovery.discover_components(temp_dir)

            assert components == []

    def test_discover_components_with_python_files(self):
        """Test component discovery with Python files"""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create some Python files
            Path(temp_dir, "module1.py").touch()
            Path(temp_dir, "module2.py").touch()
            Path(temp_dir, "subdir").mkdir()
            Path(temp_dir, "subdir", "nested.py").touch()
            Path(temp_dir, "__init__.py").touch()  # Should be ignored

            discovery = ComponentDiscovery()
            components = discovery.discover_components(temp_dir)

            assert len(components) == 3  # module1, module2, nested

            # Check component names
            component_names = [c.name for c in components]
            assert "module1" in component_names
            assert "module2" in component_names
            assert "subdir.nested" in component_names

    def test_discover_components_nonexistent_dir(self):
        """Test component discovery with non-existent directory"""
        discovery = ComponentDiscovery()
        components = discovery.discover_components("/nonexistent/path")

        assert components == []

    def test_resolve_dependencies(self):
        """Test dependency resolution"""
        components = [
            TestComponent("comp1", "type1", "1.0.0", ["dep1", "dep2"]),
            TestComponent("comp2", "type2", "2.0.0", ["dep3"]),
            TestComponent("comp3", "type3", "3.0.0", []),
        ]

        discovery = ComponentDiscovery()
        dependency_map = discovery.resolve_dependencies(components)

        assert dependency_map["comp1"] == ["dep1", "dep2"]
        assert dependency_map["comp2"] == ["dep3"]
        assert dependency_map["comp3"] == []


class TestResultAnalyzer:
    """Test TestResultAnalyzer class"""

    def test_calculate_success_rate_empty(self):
        """Test success rate calculation with empty results"""
        rate = IntegrationTestResultAnalyzer.calculate_success_rate([])

        assert rate == 0.0

    def test_calculate_success_rate_all_passed(self):
        """Test success rate calculation with all passed tests"""
        results = [
            IntegrationTestResult("test1", True),
            IntegrationTestResult("test2", True),
            IntegrationTestResult("test3", True),
        ]

        rate = IntegrationTestResultAnalyzer.calculate_success_rate(results)

        assert rate == 100.0

    def test_calculate_success_rate_mixed(self):
        """Test success rate calculation with mixed results"""
        results = [
            IntegrationTestResult("test1", True),
            IntegrationTestResult("test2", False),
            IntegrationTestResult("test3", True),
            IntegrationTestResult("test4", False),
        ]

        rate = IntegrationTestResultAnalyzer.calculate_success_rate(results)

        assert rate == 50.0  # 2 out of 4 passed

    def test_get_failed_tests(self):
        """Test getting failed tests"""
        results = [
            IntegrationTestResult("test1", True),
            IntegrationTestResult("test2", False, "Error 1"),
            IntegrationTestResult("test3", False, "Error 2"),
            IntegrationTestResult("test4", True),
        ]

        failed = IntegrationTestResultAnalyzer.get_failed_tests(results)

        assert len(failed) == 2
        assert failed[0].test_name == "test2"
        assert failed[1].test_name == "test3"
        assert failed[0].error_message == "Error 1"
        assert failed[1].error_message == "Error 2"

    def test_get_execution_stats_empty(self):
        """Test execution stats with empty results"""
        stats = IntegrationTestResultAnalyzer.get_execution_stats([])

        expected = {"total": 0, "passed": 0, "failed": 0, "success_rate": 0.0, "total_time": 0.0, "avg_time": 0.0}

        assert stats == expected

    def test_get_execution_stats_with_results(self):
        """Test execution stats with test results"""
        results = [
            IntegrationTestResult("test1", True, execution_time=1.0),
            IntegrationTestResult("test2", False, execution_time=2.0),
            IntegrationTestResult("test3", True, execution_time=3.0),
        ]

        stats = IntegrationTestResultAnalyzer.get_execution_stats(results)

        assert stats["total"] == 3
        assert stats["passed"] == 2
        assert stats["failed"] == 1
        assert stats["success_rate"] == 66.66666666666666  # 2/3 * 100
        assert stats["total_time"] == 6.0
        assert stats["avg_time"] == 2.0


class TestEnvironment:
    """Test TestEnvironment class"""

    def test_environment_with_auto_temp_dir(self):
        """Test environment with auto-created temp directory"""
        env = TestEnvironment()

        assert env.temp_dir is not None
        assert env.created_temp is True
        assert Path(env.temp_dir).exists()

        # Test getting temp path
        temp_path = env.get_temp_path("test_file.txt")
        assert temp_path.endswith("test_file.txt")
        assert Path(temp_path).parent == Path(env.temp_dir)

        # Cleanup
        env.cleanup()
        assert not Path(env.temp_dir).exists()

    def test_environment_with_custom_temp_dir(self):
        """Test environment with custom temp directory"""
        with tempfile.TemporaryDirectory() as custom_dir:
            env = TestEnvironment(custom_dir)

            assert env.temp_dir == custom_dir
            assert env.created_temp is False

            # Cleanup should not remove custom directory
            env.cleanup()
            assert Path(custom_dir).exists()

    def test_environment_context_manager(self):
        """Test environment as context manager"""
        with TestEnvironment() as env:
            assert env.temp_dir is not None
            assert Path(env.temp_dir).exists()

            # Create a test file
            test_file = Path(env.temp_dir) / "test.txt"
            test_file.write_text("test content")
            assert test_file.exists()

        # Directory should be cleaned up after context
        assert not Path(env.temp_dir).exists()

    def test_environment_cleanup_error_handling(self):
        """Test environment cleanup error handling"""
        env = TestEnvironment()

        # Mock rmtree to raise an exception
        with patch("shutil.rmtree", side_effect=OSError("Permission denied")):
            # Should not raise exception
            env.cleanup()

        # Environment should still be considered cleaned up
        assert True  # If we reach here, cleanup didn't crash

    def test_multiple_temp_paths(self):
        """Test getting multiple temp paths"""
        with TestEnvironment() as env:
            path1 = env.get_temp_path("file1.txt")
            path2 = env.get_temp_path("file2.txt")
            path3 = env.get_temp_path("subdir/file3.txt")

            assert path1 != path2
            assert path1.endswith("file1.txt")
            assert path2.endswith("file2.txt")
            assert "subdir" in path3
            assert path3.endswith("file3.txt")

            # All paths should be within the temp directory
            assert path1.startswith(env.temp_dir)
            assert path2.startswith(env.temp_dir)
            assert path3.startswith(env.temp_dir)
