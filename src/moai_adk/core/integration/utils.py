"""
Integration testing utilities and helper functions.

This module contains utility functions for component discovery,
dependency resolution, and test result analysis.
"""

import shutil
import tempfile
from pathlib import Path
from typing import Any, Dict, List, Optional

from .models import IntegrationTestResult, TestComponent


class ComponentDiscovery:
    """
    Component discovery utilities for integration testing.
    """

    @staticmethod
    def discover_components(base_path: str) -> List[TestComponent]:
        """
        Discover testable components in the given path.

        Args:
            base_path: Base path to search for components

        Returns:
            List of discovered components
        """
        components: List[TestComponent] = []
        base_dir = Path(base_path)

        if not base_dir.exists():
            return components

        # Look for Python modules
        for py_file in base_dir.rglob("*.py"):
            if py_file.name.startswith("__"):
                continue

            relative_path = py_file.relative_to(base_dir)
            module_name = str(relative_path.with_suffix("")).replace("/", ".")

            component = TestComponent(
                name=module_name,
                component_type="python_module",
                version="1.0.0",  # Default version
            )
            components.append(component)

        return components

    @staticmethod
    def resolve_dependencies(components: List[TestComponent]) -> Dict[str, List[str]]:
        """
        Resolve dependencies between components.

        Args:
            components: List of components to analyze

        Returns:
            Dictionary mapping component names to their dependencies
        """
        dependency_map = {}

        for component in components:
            try:
                # Try to import and analyze dependencies
                dependencies = ComponentDiscovery._analyze_imports(component)
                dependency_map[component.name] = dependencies
            except Exception:
                dependency_map[component.name] = component.dependencies

        return dependency_map

    @staticmethod
    def _analyze_imports(component: TestComponent) -> List[str]:
        """
        Analyze imports for a component to determine dependencies.

        Args:
            component: Component to analyze

        Returns:
            List of dependency names
        """
        # This is a simplified implementation
        # In a real scenario, you would parse the source code
        return component.dependencies


class TestResultAnalyzer:
    """
    Test result analysis utilities.
    """

    @staticmethod
    def calculate_success_rate(results: List[IntegrationTestResult]) -> float:
        """
        Calculate the success rate of test results.

        Args:
            results: List of test results

        Returns:
            Success rate as percentage (0-100)
        """
        if not results:
            return 0.0

        passed = sum(1 for result in results if result.passed)
        return (passed / len(results)) * 100

    @staticmethod
    def get_failed_tests(
        results: List[IntegrationTestResult],
    ) -> List[IntegrationTestResult]:
        """
        Get list of failed tests.

        Args:
            results: List of test results

        Returns:
            List of failed test results
        """
        return [result for result in results if not result.passed]

    @staticmethod
    def get_execution_stats(results: List[IntegrationTestResult]) -> Dict[str, Any]:
        """
        Get execution statistics for test results.

        Args:
            results: List of test results

        Returns:
            Dictionary with execution statistics
        """
        if not results:
            return {
                "total": 0,
                "passed": 0,
                "failed": 0,
                "success_rate": 0.0,
                "total_time": 0.0,
                "avg_time": 0.0,
            }

        passed = sum(1 for result in results if result.passed)
        failed = len(results) - passed
        total_time = sum(result.execution_time for result in results)

        return {
            "total": len(results),
            "passed": passed,
            "failed": failed,
            "success_rate": TestResultAnalyzer.calculate_success_rate(results),
            "total_time": total_time,
            "avg_time": total_time / len(results),
        }


class TestEnvironment:
    """
    Test environment management utilities.
    """

    def __init__(self, temp_dir: Optional[str] = None):
        """
        Initialize test environment.

        Args:
            temp_dir: Optional temporary directory path
        """
        self.temp_dir = temp_dir
        self.created_temp = False

        if self.temp_dir is None:
            self.temp_dir = tempfile.mkdtemp(prefix="moai_integration_test_")
            self.created_temp = True

    def cleanup(self):
        """Clean up test environment."""
        if self.created_temp and self.temp_dir:
            try:
                shutil.rmtree(self.temp_dir)
            except Exception:
                pass  # Ignore cleanup errors

    def get_temp_path(self, name: str) -> str:
        """
        Get a temporary file path.

        Args:
            name: File name

        Returns:
            Full path to temporary file
        """
        return str(Path(self.temp_dir) / name)

    def __enter__(self):
        """Context manager entry."""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.cleanup()
