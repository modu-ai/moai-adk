"""
Test integration testing models and data structures.
"""

import pytest

from moai_adk.core.integration.models import (
    ComponentNotFoundError,
    IntegrationTestError,
    IntegrationTestResult,
    TestTimeoutError,
)
from moai_adk.core.integration.models import (
    TestComponent as TestComponentClass,
)
from moai_adk.core.integration.models import (
    TestStatus as TestStatusEnum,
)
from moai_adk.core.integration.models import (
    TestSuite as TestSuiteClass,
)


class TestIntegrationTestResult:
    """Test IntegrationTestResult class"""

    def test_result_creation(self):
        """Test basic result creation"""
        result = IntegrationTestResult(
            test_name="test_example", passed=True, execution_time=1.5, components_tested=["comp1", "comp2"]
        )

        assert result.test_name == "test_example"
        assert result.passed is True
        assert result.execution_time == 1.5
        assert result.components_tested == ["comp1", "comp2"]
        assert result.status == TestStatusEnum.PASSED

    def test_result_with_error(self):
        """Test result with error"""
        result = IntegrationTestResult(test_name="test_failed", passed=False, error_message="Something went wrong")

        assert result.passed is False
        assert result.error_message == "Something went wrong"
        assert result.status == TestStatusEnum.FAILED

    def test_result_defaults(self):
        """Test result with default values"""
        result = IntegrationTestResult(test_name="test_defaults", passed=True)

        assert result.error_message is None
        assert result.execution_time == 0.0
        assert result.components_tested == []
        assert result.status == TestStatusEnum.PASSED


class TestComponent:
    """Test TestComponent class"""

    def test_component_creation(self):
        """Test basic component creation"""
        component = TestComponentClass(
            name="test_component", component_type="python_module", version="1.0.0", dependencies=["dep1", "dep2"]
        )

        assert component.name == "test_component"
        assert component.component_type == "python_module"
        assert component.version == "1.0.0"
        assert component.dependencies == ["dep1", "dep2"]

    def test_component_defaults(self):
        """Test component with default dependencies"""
        component = TestComponentClass(name="test_comp", component_type="module", version="2.0.0")

        assert component.dependencies == []


class TestSuite:
    """Test TestSuite class"""

    def test_suite_creation(self):
        """Test basic suite creation"""
        comp1 = TestComponentClass("comp1", "type1", "1.0.0")
        comp2 = TestComponentClass("comp2", "type2", "2.0.0")

        suite = TestSuiteClass(
            name="test_suite",
            description="Test suite description",
            components=[comp1, comp2],
            test_cases=["test1", "test2"],
        )

        assert suite.name == "test_suite"
        assert suite.description == "Test suite description"
        assert len(suite.components) == 2
        assert suite.test_cases == ["test1", "test2"]

    def test_suite_defaults(self):
        """Test suite with default test cases"""
        component = TestComponentClass("comp", "type", "1.0.0")
        suite = TestSuiteClass(name="empty_suite", description="Empty test suite", components=[component])

        assert suite.test_cases == []


class TestExceptions:
    """Test custom exceptions"""

    def test_integration_test_error(self):
        """Test base exception"""
        with pytest.raises(IntegrationTestError) as exc_info:
            raise IntegrationTestError("Test error")

        assert str(exc_info.value) == "Test error"

    def test_test_timeout_error(self):
        """Test timeout exception"""
        with pytest.raises(TestTimeoutError) as exc_info:
            raise TestTimeoutError("Test timed out")

        assert str(exc_info.value) == "Test timed out"
        assert isinstance(exc_info.value, IntegrationTestError)

    def test_component_not_found_error(self):
        """Test component not found exception"""
        with pytest.raises(ComponentNotFoundError) as exc_info:
            raise ComponentNotFoundError("Component not found")

        assert str(exc_info.value) == "Component not found"
        assert isinstance(exc_info.value, IntegrationTestError)


class TestStatus:
    """Test TestStatus enum"""

    def test_status_values(self):
        """Test status enum values"""
        assert TestStatusEnum.PENDING.value == "pending"
        assert TestStatusEnum.RUNNING.value == "running"
        assert TestStatusEnum.PASSED.value == "passed"
        assert TestStatusEnum.FAILED.value == "failed"
        assert TestStatusEnum.SKIPPED.value == "skipped"

    def test_status_comparison(self):
        """Test status comparison"""
        status1 = TestStatusEnum.PASSED
        status2 = TestStatusEnum.PASSED
        status3 = TestStatusEnum.FAILED

        assert status1 == status2
        assert status1 != status3
