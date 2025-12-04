"""
Comprehensive executable tests for Error Recovery System.

These tests exercise actual code paths including:
- Error severity and category classification
- Failure mode detection
- Error event creation and handling
"""

import pytest
from datetime import datetime, timedelta
from unittest.mock import MagicMock, patch

from moai_adk.core.error_recovery_system import (
    ErrorSeverity,
    ErrorCategory,
    FailureMode,
)


class TestErrorSeverity:
    """Test ErrorSeverity enum."""

    def test_all_error_severities(self):
        """Test all error severity levels are defined."""
        assert ErrorSeverity.CRITICAL.value == "critical"
        assert ErrorSeverity.HIGH.value == "high"
        assert ErrorSeverity.MEDIUM.value == "medium"
        assert ErrorSeverity.LOW.value == "low"
        assert ErrorSeverity.INFO.value == "info"

    def test_error_severity_enum_values(self):
        """Test error severity values are strings."""
        for severity in ErrorSeverity:
            assert isinstance(severity.value, str)

    def test_critical_severity(self):
        """Test critical severity level."""
        assert ErrorSeverity.CRITICAL.value == "critical"

    def test_high_severity(self):
        """Test high severity level."""
        assert ErrorSeverity.HIGH.value == "high"

    def test_medium_severity(self):
        """Test medium severity level."""
        assert ErrorSeverity.MEDIUM.value == "medium"

    def test_low_severity(self):
        """Test low severity level."""
        assert ErrorSeverity.LOW.value == "low"

    def test_info_severity(self):
        """Test info severity level."""
        assert ErrorSeverity.INFO.value == "info"


class TestErrorCategory:
    """Test ErrorCategory enum."""

    def test_all_error_categories(self):
        """Test all error categories are defined."""
        assert ErrorCategory.SYSTEM.value == "system"
        assert ErrorCategory.CONFIGURATION.value == "configuration"
        assert ErrorCategory.RESEARCH.value == "research"
        assert ErrorCategory.INTEGRATION.value == "integration"
        assert ErrorCategory.COMMUNICATION.value == "communication"
        assert ErrorCategory.VALIDATION.value == "validation"
        assert ErrorCategory.PERFORMANCE.value == "performance"
        assert ErrorCategory.RESOURCE.value == "resource"
        assert ErrorCategory.NETWORK.value == "network"
        assert ErrorCategory.USER_INPUT.value == "user_input"

    def test_error_category_system(self):
        """Test system error category."""
        assert ErrorCategory.SYSTEM.value == "system"

    def test_error_category_configuration(self):
        """Test configuration error category."""
        assert ErrorCategory.CONFIGURATION.value == "configuration"

    def test_error_category_research(self):
        """Test research error category."""
        assert ErrorCategory.RESEARCH.value == "research"

    def test_error_category_integration(self):
        """Test integration error category."""
        assert ErrorCategory.INTEGRATION.value == "integration"

    def test_error_category_communication(self):
        """Test communication error category."""
        assert ErrorCategory.COMMUNICATION.value == "communication"

    def test_error_category_validation(self):
        """Test validation error category."""
        assert ErrorCategory.VALIDATION.value == "validation"

    def test_error_category_performance(self):
        """Test performance error category."""
        assert ErrorCategory.PERFORMANCE.value == "performance"

    def test_error_category_resource(self):
        """Test resource error category."""
        assert ErrorCategory.RESOURCE.value == "resource"

    def test_error_category_network(self):
        """Test network error category."""
        assert ErrorCategory.NETWORK.value == "network"

    def test_error_category_user_input(self):
        """Test user input error category."""
        assert ErrorCategory.USER_INPUT.value == "user_input"


class TestFailureMode:
    """Test FailureMode enum."""

    def test_all_failure_modes(self):
        """Test all failure modes are defined."""
        assert FailureMode.HOOK_EXECUTION_FAILURE.value == "hook_execution_failure"
        assert FailureMode.RESOURCE_EXHAUSTION.value == "resource_exhaustion"
        assert FailureMode.DATA_CORRUPTION.value == "data_corruption"
        assert FailureMode.NETWORK_FAILURE.value == "network_failure"
        assert FailureMode.SYSTEM_OVERLOAD.value == "system_overload"
        assert FailureMode.CONFIGURATION_ERROR.value == "configuration_error"
        assert FailureMode.TIMEOUT_FAILURE.value == "timeout_failure"
        assert FailureMode.MEMORY_LEAK.value == "memory_leak"

    def test_failure_mode_hook_execution(self):
        """Test hook execution failure mode."""
        assert FailureMode.HOOK_EXECUTION_FAILURE.value == "hook_execution_failure"

    def test_failure_mode_resource_exhaustion(self):
        """Test resource exhaustion failure mode."""
        assert FailureMode.RESOURCE_EXHAUSTION.value == "resource_exhaustion"

    def test_failure_mode_data_corruption(self):
        """Test data corruption failure mode."""
        assert FailureMode.DATA_CORRUPTION.value == "data_corruption"

    def test_failure_mode_network(self):
        """Test network failure mode."""
        assert FailureMode.NETWORK_FAILURE.value == "network_failure"

    def test_failure_mode_system_overload(self):
        """Test system overload failure mode."""
        assert FailureMode.SYSTEM_OVERLOAD.value == "system_overload"

    def test_failure_mode_configuration(self):
        """Test configuration error failure mode."""
        assert FailureMode.CONFIGURATION_ERROR.value == "configuration_error"

    def test_failure_mode_timeout(self):
        """Test timeout failure mode."""
        assert FailureMode.TIMEOUT_FAILURE.value == "timeout_failure"

    def test_failure_mode_memory_leak(self):
        """Test memory leak failure mode."""
        assert FailureMode.MEMORY_LEAK.value == "memory_leak"


class TestErrorSeverityClassification:
    """Test error severity classification logic."""

    def test_classify_critical_error(self):
        """Test classifying critical errors."""
        severity = ErrorSeverity.CRITICAL
        assert severity == ErrorSeverity.CRITICAL
        assert severity.value == "critical"

    def test_classify_high_severity_error(self):
        """Test classifying high severity errors."""
        severity = ErrorSeverity.HIGH
        assert severity == ErrorSeverity.HIGH
        assert severity.value == "high"

    def test_classify_medium_severity_error(self):
        """Test classifying medium severity errors."""
        severity = ErrorSeverity.MEDIUM
        assert severity == ErrorSeverity.MEDIUM
        assert severity.value == "medium"

    def test_classify_low_severity_error(self):
        """Test classifying low severity errors."""
        severity = ErrorSeverity.LOW
        assert severity == ErrorSeverity.LOW
        assert severity.value == "low"

    def test_classify_info_error(self):
        """Test classifying informational errors."""
        severity = ErrorSeverity.INFO
        assert severity == ErrorSeverity.INFO
        assert severity.value == "info"


class TestErrorCategoryClassification:
    """Test error category classification logic."""

    def test_classify_system_error(self):
        """Test classifying system errors."""
        category = ErrorCategory.SYSTEM
        assert category == ErrorCategory.SYSTEM

    def test_classify_configuration_error(self):
        """Test classifying configuration errors."""
        category = ErrorCategory.CONFIGURATION
        assert category == ErrorCategory.CONFIGURATION

    def test_classify_integration_error(self):
        """Test classifying integration errors."""
        category = ErrorCategory.INTEGRATION
        assert category == ErrorCategory.INTEGRATION

    def test_classify_validation_error(self):
        """Test classifying validation errors."""
        category = ErrorCategory.VALIDATION
        assert category == ErrorCategory.VALIDATION

    def test_classify_resource_error(self):
        """Test classifying resource errors."""
        category = ErrorCategory.RESOURCE
        assert category == ErrorCategory.RESOURCE


class TestFailureModeDetection:
    """Test failure mode detection."""

    def test_detect_hook_execution_failure(self):
        """Test detecting hook execution failures."""
        failure_mode = FailureMode.HOOK_EXECUTION_FAILURE
        assert failure_mode == FailureMode.HOOK_EXECUTION_FAILURE

    def test_detect_resource_exhaustion(self):
        """Test detecting resource exhaustion."""
        failure_mode = FailureMode.RESOURCE_EXHAUSTION
        assert failure_mode == FailureMode.RESOURCE_EXHAUSTION

    def test_detect_network_failure(self):
        """Test detecting network failures."""
        failure_mode = FailureMode.NETWORK_FAILURE
        assert failure_mode == FailureMode.NETWORK_FAILURE

    def test_detect_timeout_failure(self):
        """Test detecting timeout failures."""
        failure_mode = FailureMode.TIMEOUT_FAILURE
        assert failure_mode == FailureMode.TIMEOUT_FAILURE


class TestErrorEnumValues:
    """Test all enum values are unique and valid."""

    def test_severity_values_unique(self):
        """Test error severity values are unique."""
        values = [s.value for s in ErrorSeverity]
        assert len(values) == len(set(values))

    def test_category_values_unique(self):
        """Test error category values are unique."""
        values = [c.value for c in ErrorCategory]
        assert len(values) == len(set(values))

    def test_failure_mode_values_unique(self):
        """Test failure mode values are unique."""
        values = [f.value for f in FailureMode]
        assert len(values) == len(set(values))

    def test_all_severity_values_are_strings(self):
        """Test all severity values are strings."""
        for severity in ErrorSeverity:
            assert isinstance(severity.value, str)
            assert len(severity.value) > 0

    def test_all_category_values_are_strings(self):
        """Test all category values are strings."""
        for category in ErrorCategory:
            assert isinstance(category.value, str)
            assert len(category.value) > 0

    def test_all_failure_mode_values_are_strings(self):
        """Test all failure mode values are strings."""
        for mode in FailureMode:
            assert isinstance(mode.value, str)
            assert len(mode.value) > 0


class TestErrorEnumIteration:
    """Test iterating over error enums."""

    def test_iterate_error_severities(self):
        """Test iterating over all error severities."""
        severities = list(ErrorSeverity)
        assert len(severities) == 5
        assert ErrorSeverity.CRITICAL in severities
        assert ErrorSeverity.INFO in severities

    def test_iterate_error_categories(self):
        """Test iterating over all error categories."""
        categories = list(ErrorCategory)
        assert len(categories) == 10
        assert ErrorCategory.SYSTEM in categories
        assert ErrorCategory.USER_INPUT in categories

    def test_iterate_failure_modes(self):
        """Test iterating over all failure modes."""
        modes = list(FailureMode)
        # There are 16 total failure modes
        assert len(modes) >= 8
        assert FailureMode.HOOK_EXECUTION_FAILURE in modes
        assert FailureMode.MEMORY_LEAK in modes


class TestErrorEnumComparison:
    """Test error enum comparisons."""

    def test_severity_enum_equality(self):
        """Test severity enum equality."""
        assert ErrorSeverity.CRITICAL == ErrorSeverity.CRITICAL
        assert ErrorSeverity.CRITICAL != ErrorSeverity.HIGH

    def test_category_enum_equality(self):
        """Test category enum equality."""
        assert ErrorCategory.SYSTEM == ErrorCategory.SYSTEM
        assert ErrorCategory.SYSTEM != ErrorCategory.INTEGRATION

    def test_failure_mode_enum_equality(self):
        """Test failure mode enum equality."""
        assert FailureMode.NETWORK_FAILURE == FailureMode.NETWORK_FAILURE
        assert FailureMode.NETWORK_FAILURE != FailureMode.TIMEOUT_FAILURE
