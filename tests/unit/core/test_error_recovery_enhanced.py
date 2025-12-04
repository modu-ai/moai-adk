"""
Enhanced tests for Error Recovery System - targeting 70%+ coverage.

Focus on actual API and implementation that exists.
"""

import pytest
import tempfile
from datetime import datetime, timezone
from pathlib import Path
from unittest.mock import MagicMock, patch

from moai_adk.core.error_recovery_system import (
    ErrorSeverity,
    ErrorCategory,
    FailureMode,
    RecoveryStrategy,
    ErrorRecoverySystem,
    get_error_recovery_system,
    handle_error,
)


class TestErrorSeverityValues:
    """Test error severity enum."""

    def test_all_severity_values(self):
        """Test all severity values exist."""
        assert ErrorSeverity.CRITICAL.value == "critical"
        assert ErrorSeverity.HIGH.value == "high"
        assert ErrorSeverity.MEDIUM.value == "medium"
        assert ErrorSeverity.LOW.value == "low"
        assert ErrorSeverity.INFO.value == "info"


class TestErrorCategoryValues:
    """Test error category enum."""

    def test_all_categories_exist(self):
        """Test all categories are defined."""
        assert ErrorCategory.SYSTEM.value == "system"
        assert ErrorCategory.INTEGRATION.value == "integration"
        assert ErrorCategory.VALIDATION.value == "validation"
        assert ErrorCategory.NETWORK.value == "network"


class TestFailureModes:
    """Test failure mode enumeration."""

    def test_failure_modes_defined(self):
        """Test failure modes exist."""
        assert FailureMode.HOOK_EXECUTION_FAILURE.value == "hook_execution_failure"
        assert FailureMode.RESOURCE_EXHAUSTION.value == "resource_exhaustion"
        assert FailureMode.NETWORK_FAILURE.value == "network_failure"


class TestRecoveryStrategies:
    """Test recovery strategy enumeration."""

    def test_strategies_defined(self):
        """Test all recovery strategies are defined."""
        strategies = [s.value for s in RecoveryStrategy]
        assert len(strategies) > 0


@patch("pathlib.Path.mkdir")
def test_error_recovery_system_singleton(mock_mkdir):
    """Test singleton pattern."""
    system1 = get_error_recovery_system()
    system2 = get_error_recovery_system()
    assert system1 is system2


def test_handle_error_convenience_function():
    """Test handle_error convenience function."""
    result = handle_error("Test", ErrorSeverity.MEDIUM, ErrorCategory.INTEGRATION)
    assert result is not None


@patch("pathlib.Path.mkdir")
def test_system_health_status(mock_mkdir):
    """Test system health monitoring."""
    system = ErrorRecoverySystem()
    health = system.get_system_health()
    assert "status" in health
    assert "active_errors" in health


@patch("pathlib.Path.mkdir")
def test_error_summary(mock_mkdir):
    """Test error summary generation."""
    system = ErrorRecoverySystem()
    summary = system.get_error_summary()
    assert "total_recent_errors" in summary
    assert "by_severity" in summary


@patch("pathlib.Path.mkdir")
def test_troubleshooting_guide(mock_mkdir):
    """Test troubleshooting guide generation."""
    system = ErrorRecoverySystem()
    guide = system.generate_troubleshooting_guide()
    assert "generated_at" in guide
    assert "common_issues" in guide


@patch("pathlib.Path.mkdir")
def test_error_cleanup(mock_mkdir):
    """Test error cleanup functionality."""
    system = ErrorRecoverySystem()
    result = system.cleanup_old_errors(days_to_keep=30)
    assert result is not None


@patch("pathlib.Path.mkdir")
def test_pattern_identification(mock_mkdir):
    """Test error pattern identification."""
    system = ErrorRecoverySystem()
    patterns = system._identify_error_patterns(system.error_history)
    assert isinstance(patterns, dict)


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
