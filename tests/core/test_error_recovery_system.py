"""
Comprehensive Test Suite for Error Recovery System

Tests cover:
- Error severity and category classification
- Error report generation and storage
- Recovery action registration and execution
- Automatic recovery mechanisms
- Manual recovery procedures
- System health monitoring
- Error pattern detection and analysis
- Error history management and cleanup
- Troubleshooting guide generation
"""

import os
import tempfile
from datetime import datetime, timedelta, timezone
from pathlib import Path
from unittest.mock import Mock, mock_open, patch

import pytest

try:
    from moai_adk.core.error_recovery_system import (
        ErrorCategory,
        ErrorRecoverySystem,
        ErrorReport,
        ErrorSeverity,
        RecoveryAction,
        RecoveryResult,
        error_handler,
        get_error_recovery_system,
        handle_error,
    )
except ImportError:
    pytest.skip("error_recovery_system not available", allow_module_level=True)


@pytest.fixture
def temp_project_dir():
    """Create a temporary project directory structure"""
    with tempfile.TemporaryDirectory() as tmpdir:
        project_root = Path(tmpdir)

        # Create necessary directory structure
        (project_root / ".moai" / "cache").mkdir(parents=True, exist_ok=True)
        (project_root / ".moai" / "agent_state").mkdir(parents=True, exist_ok=True)
        (project_root / ".moai" / "comm_cache").mkdir(parents=True, exist_ok=True)
        (project_root / ".moai" / "config_backups").mkdir(parents=True, exist_ok=True)
        (project_root / ".moai" / "error_logs").mkdir(parents=True, exist_ok=True)
        (project_root / ".claude" / "cache").mkdir(parents=True, exist_ok=True)
        (project_root / ".claude" / "skills").mkdir(parents=True, exist_ok=True)
        (project_root / ".claude" / "agents" / "alfred").mkdir(parents=True, exist_ok=True)
        (project_root / ".claude" / "commands" / "alfred").mkdir(parents=True, exist_ok=True)

        yield project_root


@pytest.fixture
def error_recovery_system(temp_project_dir):
    """Create an ErrorRecoverySystem instance for testing"""
    system = ErrorRecoverySystem(project_root=temp_project_dir)
    yield system
    # Cleanup
    system.monitoring_active = False


class TestErrorSeverityEnum:
    """Tests for ErrorSeverity enumeration"""

    def test_error_severity_values(self):
        """Test that all error severity levels are defined"""
        assert ErrorSeverity.CRITICAL.value == "critical"
        assert ErrorSeverity.HIGH.value == "high"
        assert ErrorSeverity.MEDIUM.value == "medium"
        assert ErrorSeverity.LOW.value == "low"
        assert ErrorSeverity.INFO.value == "info"

    def test_error_severity_count(self):
        """Test correct number of severity levels"""
        severities = list(ErrorSeverity)
        assert len(severities) == 5


class TestErrorCategoryEnum:
    """Tests for ErrorCategory enumeration"""

    def test_error_category_values(self):
        """Test that all error categories are defined"""
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

    def test_error_category_count(self):
        """Test correct number of error categories"""
        categories = list(ErrorCategory)
        assert len(categories) == 10


class TestErrorReport:
    """Tests for ErrorReport dataclass"""

    def test_error_report_creation(self):
        """Test ErrorReport creation with all fields"""
        timestamp = datetime.now(timezone.utc)
        report = ErrorReport(
            id="ERR_20240101_000000_123abc",
            timestamp=timestamp,
            severity=ErrorSeverity.CRITICAL,
            category=ErrorCategory.SYSTEM,
            message="Test error message",
            details={"key": "value"},
            stack_trace="traceback info",
            context={"context_key": "context_value"},
            recovery_attempted=True,
            recovery_successful=True,
            resolution_message="Error resolved",
        )

        assert report.id == "ERR_20240101_000000_123abc"
        assert report.timestamp == timestamp
        assert report.severity == ErrorSeverity.CRITICAL
        assert report.category == ErrorCategory.SYSTEM
        assert report.message == "Test error message"
        assert report.details == {"key": "value"}
        assert report.stack_trace == "traceback info"
        assert report.context == {"context_key": "context_value"}
        assert report.recovery_attempted is True
        assert report.recovery_successful is True
        assert report.resolution_message == "Error resolved"

    def test_error_report_defaults(self):
        """Test ErrorReport default field values"""
        timestamp = datetime.now(timezone.utc)
        report = ErrorReport(
            id="ERR_test",
            timestamp=timestamp,
            severity=ErrorSeverity.LOW,
            category=ErrorCategory.USER_INPUT,
            message="Test",
            details={},
            stack_trace=None,
            context={},
        )

        assert report.recovery_attempted is False
        assert report.recovery_successful is False
        assert report.resolution_message is None


class TestRecoveryAction:
    """Tests for RecoveryAction dataclass"""

    def test_recovery_action_creation(self):
        """Test RecoveryAction creation with all fields"""
        handler = Mock(return_value=True)
        action = RecoveryAction(
            name="test_action",
            description="Test recovery action",
            action_type="automatic",
            severity_filter=[ErrorSeverity.CRITICAL, ErrorSeverity.HIGH],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler,
            timeout=30.0,
            max_attempts=3,
            success_criteria="System responsive",
        )

        assert action.name == "test_action"
        assert action.description == "Test recovery action"
        assert action.action_type == "automatic"
        assert action.severity_filter == [ErrorSeverity.CRITICAL, ErrorSeverity.HIGH]
        assert action.category_filter == [ErrorCategory.SYSTEM]
        assert action.handler == handler
        assert action.timeout == 30.0
        assert action.max_attempts == 3
        assert action.success_criteria == "System responsive"

    def test_recovery_action_defaults(self):
        """Test RecoveryAction default field values"""
        handler = Mock()
        action = RecoveryAction(
            name="test",
            description="Test",
            action_type="manual",
            severity_filter=[ErrorSeverity.MEDIUM],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler,
        )

        assert action.timeout is None
        assert action.max_attempts == 3
        assert action.success_criteria is None


class TestRecoveryResult:
    """Tests for RecoveryResult dataclass"""

    def test_recovery_result_creation_success(self):
        """Test RecoveryResult creation for successful recovery"""
        result = RecoveryResult(
            success=True,
            action_name="test_action",
            message="Recovery successful",
            duration=1.5,
            details={"recovered": True},
            next_actions=["validate", "monitor"],
        )

        assert result.success is True
        assert result.action_name == "test_action"
        assert result.message == "Recovery successful"
        assert result.duration == 1.5
        assert result.details == {"recovered": True}
        assert result.next_actions == ["validate", "monitor"]

    def test_recovery_result_creation_failure(self):
        """Test RecoveryResult creation for failed recovery"""
        result = RecoveryResult(
            success=False,
            action_name="test_action",
            message="Recovery failed: timeout",
            duration=30.0,
        )

        assert result.success is False
        assert result.action_name == "test_action"
        assert result.message == "Recovery failed: timeout"
        assert result.duration == 30.0
        assert result.details is None
        assert result.next_actions is None


class TestErrorRecoverySystemInitialization:
    """Tests for ErrorRecoverySystem initialization"""

    def test_system_initialization(self, temp_project_dir):
        """Test ErrorRecoverySystem initialization"""
        system = ErrorRecoverySystem(project_root=temp_project_dir)

        assert system.project_root == temp_project_dir
        assert system.error_log_dir == temp_project_dir / ".moai" / "error_logs"
        assert system.error_log_dir.exists()
        assert isinstance(system.active_errors, dict)
        assert isinstance(system.error_history, list)
        assert isinstance(system.recovery_actions, dict)
        assert system.monitoring_active is True
        assert system.monitor_thread.is_alive()

        # Cleanup
        system.monitoring_active = False

    def test_system_initialization_creates_log_directory(self, temp_project_dir):
        """Test that initialization creates error log directory"""
        system = ErrorRecoverySystem(project_root=temp_project_dir)
        assert (temp_project_dir / ".moai" / "error_logs").exists()
        system.monitoring_active = False

    def test_default_recovery_actions_registered(self, error_recovery_system):
        """Test that default recovery actions are registered"""
        action_names = list(error_recovery_system.recovery_actions.keys())

        assert "restart_research_engines" in action_names
        assert "restore_config_backup" in action_names
        assert "clear_agent_cache" in action_names
        assert "validate_research_integrity" in action_names
        assert "rollback_last_changes" in action_names
        assert "reset_system_state" in action_names
        assert "optimize_performance" in action_names
        assert "free_resources" in action_names

    def test_system_health_initialized(self, error_recovery_system):
        """Test initial system health status"""
        health = error_recovery_system.system_health

        assert health["status"] == "healthy"
        assert "last_check" in health
        assert isinstance(health["issues"], list)
        assert isinstance(health["metrics"], dict)

    def test_error_stats_initialized(self, error_recovery_system):
        """Test initial error statistics"""
        stats = error_recovery_system.error_stats

        assert stats["total_errors"] == 0
        assert isinstance(stats["by_severity"], dict)
        assert isinstance(stats["by_category"], dict)
        assert stats["recovery_success_rate"] == 0.0


class TestErrorHandling:
    """Tests for error handling functionality"""

    def test_handle_error_basic(self, error_recovery_system):
        """Test basic error handling"""
        error = ValueError("Test error")
        report = error_recovery_system.handle_error(
            error,
            severity=ErrorSeverity.MEDIUM,
            category=ErrorCategory.VALIDATION,
        )

        assert report.id.startswith("ERR_")
        assert report.message == "Test error"
        assert report.severity == ErrorSeverity.MEDIUM
        assert report.category == ErrorCategory.VALIDATION
        assert report.stack_trace is not None
        assert len(error_recovery_system.error_history) == 1

    def test_handle_error_with_context(self, error_recovery_system):
        """Test error handling with additional context"""
        error = RuntimeError("Test runtime error")
        context = {"user": "test_user", "action": "test_action"}

        report = error_recovery_system.handle_error(
            error,
            context=context,
            severity=ErrorSeverity.HIGH,
            category=ErrorCategory.SYSTEM,
        )

        assert report.context == context
        assert report.severity == ErrorSeverity.HIGH
        assert report.category == ErrorCategory.SYSTEM

    def test_handle_error_updates_stats(self, error_recovery_system):
        """Test that handling error updates statistics"""
        error1 = ValueError("Error 1")
        error2 = RuntimeError("Error 2")

        error_recovery_system.handle_error(error1, severity=ErrorSeverity.CRITICAL)
        error_recovery_system.handle_error(error2, severity=ErrorSeverity.HIGH)

        stats = error_recovery_system.error_stats
        assert stats["total_errors"] == 2
        assert stats["by_severity"]["critical"] == 1
        assert stats["by_severity"]["high"] == 1

    def test_handle_error_stores_in_history(self, error_recovery_system):
        """Test that handled errors are stored in history"""
        error = ValueError("Test error")
        report = error_recovery_system.handle_error(error)

        assert report in error_recovery_system.error_history
        assert len(error_recovery_system.error_history) == 1
        assert error_recovery_system.error_history[0].message == "Test error"

    def test_handle_error_with_exception_attributes(self, error_recovery_system):
        """Test error handling captures exception attributes"""
        error = ValueError("Test with code")
        error.code = "VAL_001"

        report = error_recovery_system.handle_error(error)

        assert report.details["exception_type"] == "ValueError"
        assert report.details["exception_module"] == "builtins"
        assert report.details["error_code"] == "VAL_001"

    def test_handle_critical_error_attempts_recovery(self, error_recovery_system):
        """Test that critical errors trigger automatic recovery"""
        error = RuntimeError("Critical error")

        with patch.object(error_recovery_system, "_attempt_automatic_recovery") as mock_recovery:
            mock_recovery.return_value = RecoveryResult(
                success=True,
                action_name="test",
                message="Recovered",
                duration=1.0,
            )

            report = error_recovery_system.handle_error(
                error, severity=ErrorSeverity.CRITICAL, category=ErrorCategory.SYSTEM
            )

            assert report.recovery_attempted is True
            assert report.recovery_successful is True
            mock_recovery.assert_called_once()

    def test_handle_high_error_attempts_recovery(self, error_recovery_system):
        """Test that high severity errors trigger automatic recovery"""
        error = RuntimeError("High severity error")

        with patch.object(error_recovery_system, "_attempt_automatic_recovery") as mock_recovery:
            mock_recovery.return_value = RecoveryResult(
                success=False,
                action_name="test",
                message="Recovery failed",
                duration=2.0,
            )

            report = error_recovery_system.handle_error(
                error, severity=ErrorSeverity.HIGH, category=ErrorCategory.RESEARCH
            )

            assert report.recovery_attempted is True
            assert report.recovery_successful is False

    def test_handle_low_error_no_automatic_recovery(self, error_recovery_system):
        """Test that low severity errors don't trigger automatic recovery"""
        error = RuntimeError("Low severity error")

        with patch.object(error_recovery_system, "_attempt_automatic_recovery") as mock_recovery:
            report = error_recovery_system.handle_error(
                error, severity=ErrorSeverity.LOW, category=ErrorCategory.USER_INPUT
            )

            assert report.recovery_attempted is False
            mock_recovery.assert_not_called()

    def test_error_logged_to_file(self, error_recovery_system):
        """Test that errors are logged to file"""
        error = ValueError("Test error for logging")

        with patch("builtins.open", mock_open()) as mock_file:
            error_recovery_system.handle_error(error)

            # File should be opened for writing
            mock_file.assert_called()

    def test_multiple_errors_stored_correctly(self, error_recovery_system):
        """Test storing multiple errors"""
        errors = [
            ValueError("Error 1"),
            RuntimeError("Error 2"),
            TypeError("Error 3"),
        ]

        for error in errors:
            error_recovery_system.handle_error(error)

        assert len(error_recovery_system.error_history) == 3
        assert len(error_recovery_system.active_errors) == 3


class TestRecoveryActionManagement:
    """Tests for recovery action registration and management"""

    def test_register_recovery_action(self, error_recovery_system):
        """Test registering a new recovery action"""
        handler = Mock(return_value=True)
        action = RecoveryAction(
            name="custom_recovery",
            description="Custom recovery action",
            action_type="manual",
            severity_filter=[ErrorSeverity.CRITICAL],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler,
        )

        error_recovery_system.register_recovery_action(action)

        assert "custom_recovery" in error_recovery_system.recovery_actions
        assert error_recovery_system.recovery_actions["custom_recovery"] == action

    def test_register_multiple_actions(self, error_recovery_system):
        """Test registering multiple recovery actions"""
        initial_count = len(error_recovery_system.recovery_actions)

        for i in range(3):
            action = RecoveryAction(
                name=f"action_{i}",
                description=f"Action {i}",
                action_type="automatic",
                severity_filter=[ErrorSeverity.MEDIUM],
                category_filter=[ErrorCategory.SYSTEM],
                handler=Mock(return_value=True),
            )
            error_recovery_system.register_recovery_action(action)

        assert len(error_recovery_system.recovery_actions) == initial_count + 3

    def test_register_action_overwrites_existing(self, error_recovery_system):
        """Test that registering action with same name overwrites"""
        handler1 = Mock(return_value=True)
        action1 = RecoveryAction(
            name="overwrite_test",
            description="First action",
            action_type="automatic",
            severity_filter=[ErrorSeverity.HIGH],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler1,
        )

        handler2 = Mock(return_value=False)
        action2 = RecoveryAction(
            name="overwrite_test",
            description="Second action",
            action_type="manual",
            severity_filter=[ErrorSeverity.CRITICAL],
            category_filter=[ErrorCategory.INTEGRATION],
            handler=handler2,
        )

        error_recovery_system.register_recovery_action(action1)
        error_recovery_system.register_recovery_action(action2)

        assert error_recovery_system.recovery_actions["overwrite_test"].description == "Second action"
        assert error_recovery_system.recovery_actions["overwrite_test"].handler == handler2


class TestManualRecovery:
    """Tests for manual recovery operations"""

    def test_attempt_manual_recovery_success(self, error_recovery_system):
        """Test successful manual recovery"""
        # Create error
        error = ValueError("Test error")
        report = error_recovery_system.handle_error(error)
        error_id = report.id

        # Register recovery action
        handler = Mock(return_value={"status": "recovered"})
        action = RecoveryAction(
            name="test_recovery",
            description="Test recovery",
            action_type="manual",
            severity_filter=[ErrorSeverity.MEDIUM],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler,
        )
        error_recovery_system.register_recovery_action(action)

        # Attempt recovery
        result = error_recovery_system.attempt_manual_recovery(error_id, "test_recovery")

        assert result.success is True
        assert result.action_name == "test_recovery"
        assert error_id not in error_recovery_system.active_errors
        assert handler.call_count == 1

    def test_attempt_manual_recovery_nonexistent_error(self, error_recovery_system):
        """Test manual recovery on non-existent error"""
        result = error_recovery_system.attempt_manual_recovery("ERR_nonexistent", "some_action")

        assert result.success is False
        assert "not found in active errors" in result.message

    def test_attempt_manual_recovery_nonexistent_action(self, error_recovery_system):
        """Test manual recovery with non-existent action"""
        # Create error
        error = ValueError("Test error")
        report = error_recovery_system.handle_error(error)
        error_id = report.id

        # Attempt with non-existent action
        result = error_recovery_system.attempt_manual_recovery(error_id, "nonexistent_action")

        assert result.success is False
        assert "not found" in result.message

    def test_manual_recovery_with_parameters(self, error_recovery_system):
        """Test manual recovery passing parameters"""
        # Create error
        error = ValueError("Test error")
        report = error_recovery_system.handle_error(error)
        error_id = report.id

        # Register recovery action
        handler = Mock(return_value=True)
        action = RecoveryAction(
            name="parameterized_recovery",
            description="Recovery with parameters",
            action_type="manual",
            severity_filter=[ErrorSeverity.MEDIUM],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler,
        )
        error_recovery_system.register_recovery_action(action)

        # Attempt recovery with parameters
        params = {"retry": True, "timeout": 30}
        result = error_recovery_system.attempt_manual_recovery(error_id, "parameterized_recovery", params)

        assert result.success is True
        handler.assert_called_once()
        # Check parameters passed to handler
        call_args = handler.call_args
        assert call_args[0][1] == params

    def test_manual_recovery_handler_exception(self, error_recovery_system):
        """Test manual recovery when handler raises exception"""
        # Create error
        error = ValueError("Test error")
        report = error_recovery_system.handle_error(error)
        error_id = report.id

        # Register recovery action with failing handler
        handler = Mock(side_effect=RuntimeError("Handler failed"))
        action = RecoveryAction(
            name="failing_recovery",
            description="Recovery that fails",
            action_type="manual",
            severity_filter=[ErrorSeverity.MEDIUM],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler,
        )
        error_recovery_system.register_recovery_action(action)

        # Attempt recovery
        result = error_recovery_system.attempt_manual_recovery(error_id, "failing_recovery")

        assert result.success is False
        assert "Handler failed" in result.message

    def test_manual_recovery_updates_error_report(self, error_recovery_system):
        """Test that manual recovery updates error report"""
        # Create error
        error = ValueError("Test error")
        report = error_recovery_system.handle_error(error)
        error_id = report.id

        # Register recovery action
        handler = Mock(return_value={"status": "fixed"})
        action = RecoveryAction(
            name="updating_recovery",
            description="Recovery that updates error",
            action_type="manual",
            severity_filter=[ErrorSeverity.MEDIUM],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler,
        )
        error_recovery_system.register_recovery_action(action)

        # Attempt recovery
        error_recovery_system.attempt_manual_recovery(error_id, "updating_recovery")

        # Check error report was updated
        updated_report = error_recovery_system.error_history[0]
        assert updated_report.recovery_successful is True
        assert "completed successfully" in updated_report.resolution_message


class TestSystemHealth:
    """Tests for system health monitoring"""

    def test_get_system_health(self, error_recovery_system):
        """Test getting system health status"""
        health = error_recovery_system.get_system_health()

        assert "status" in health
        assert "last_check" in health
        assert "active_errors" in health
        assert "total_errors" in health
        assert "error_stats" in health
        assert "issues" in health
        assert "metrics" in health
        assert "recovery_actions_available" in health

    def test_system_health_healthy_status(self, error_recovery_system):
        """Test system health is healthy with no errors"""
        health = error_recovery_system.get_system_health()
        assert health["status"] == "healthy"
        assert health["active_errors"] == 0
        assert health["total_errors"] == 0

    def test_system_health_degraded_with_high_errors(self, error_recovery_system):
        """Test system health becomes degraded with high severity errors"""
        # Patch automatic recovery to fail so error remains in active_errors
        with patch.object(error_recovery_system, "_attempt_automatic_recovery") as mock_recovery:
            mock_recovery.return_value = RecoveryResult(
                success=False,
                action_name="test",
                message="Recovery failed",
                duration=1.0,
            )
            error = RuntimeError("High severity error")
            error_recovery_system.handle_error(error, severity=ErrorSeverity.HIGH)

        health = error_recovery_system.get_system_health()
        assert health["status"] == "degraded"

    def test_system_health_critical_with_critical_errors(self, error_recovery_system):
        """Test system health becomes critical with critical errors"""
        # Patch automatic recovery to fail so error remains in active_errors
        with patch.object(error_recovery_system, "_attempt_automatic_recovery") as mock_recovery:
            mock_recovery.return_value = RecoveryResult(
                success=False,
                action_name="test",
                message="Recovery failed",
                duration=1.0,
            )
            error = RuntimeError("Critical error")
            error_recovery_system.handle_error(error, severity=ErrorSeverity.CRITICAL)

        health = error_recovery_system.get_system_health()
        assert health["status"] == "critical"

    def test_system_health_warning_with_many_errors(self, error_recovery_system):
        """Test system health shows warning with many errors"""
        for i in range(6):
            error = ValueError(f"Error {i}")
            error_recovery_system.handle_error(error, severity=ErrorSeverity.MEDIUM)

        health = error_recovery_system.get_system_health()
        assert health["status"] == "warning"

    def test_system_health_recovery_success_rate(self, error_recovery_system):
        """Test system health includes recovery success rate"""
        # Create errors
        error1 = ValueError("Error 1")
        error_recovery_system.handle_error(error1)

        # Simulate successful recovery
        error_recovery_system.error_history[0].recovery_successful = True

        health = error_recovery_system.get_system_health()
        assert health["metrics"]["recovery_success_rate"] > 0


class TestErrorSummary:
    """Tests for error summary functionality"""

    def test_get_error_summary_empty(self, error_recovery_system):
        """Test error summary with no errors"""
        summary = error_recovery_system.get_error_summary()

        assert summary["total_recent_errors"] == 0
        assert summary["active_errors"] == 0
        assert isinstance(summary["by_severity"], dict)
        assert isinstance(summary["by_category"], dict)
        assert summary["common_patterns"] == {}
        assert summary["recovery_rate"] == 0.0

    def test_get_error_summary_with_errors(self, error_recovery_system):
        """Test error summary with multiple errors"""
        # Patch automatic recovery to fail so errors remain in active_errors
        with patch.object(error_recovery_system, "_attempt_automatic_recovery") as mock_recovery:
            mock_recovery.return_value = RecoveryResult(
                success=False,
                action_name="test",
                message="Recovery failed",
                duration=1.0,
            )
            errors = [
                (ValueError("Error 1"), ErrorSeverity.HIGH, ErrorCategory.VALIDATION),
                (RuntimeError("Error 2"), ErrorSeverity.CRITICAL, ErrorCategory.SYSTEM),
                (TypeError("Error 3"), ErrorSeverity.MEDIUM, ErrorCategory.RESEARCH),
            ]

            for error, severity, category in errors:
                error_recovery_system.handle_error(error, severity=severity, category=category)

        summary = error_recovery_system.get_error_summary()

        assert summary["total_recent_errors"] == 3
        assert summary["active_errors"] == 3
        assert len(summary["recent_errors"]) == 3

    def test_error_summary_by_severity(self, error_recovery_system):
        """Test error summary categorization by severity"""
        error_recovery_system.handle_error(ValueError("Error 1"), severity=ErrorSeverity.CRITICAL)
        error_recovery_system.handle_error(RuntimeError("Error 2"), severity=ErrorSeverity.CRITICAL)
        error_recovery_system.handle_error(TypeError("Error 3"), severity=ErrorSeverity.HIGH)

        summary = error_recovery_system.get_error_summary()

        assert summary["by_severity"]["critical"] == 2
        assert summary["by_severity"]["high"] == 1

    def test_error_summary_by_category(self, error_recovery_system):
        """Test error summary categorization by category"""
        error_recovery_system.handle_error(ValueError("Error 1"), category=ErrorCategory.SYSTEM)
        error_recovery_system.handle_error(RuntimeError("Error 2"), category=ErrorCategory.RESEARCH)
        error_recovery_system.handle_error(TypeError("Error 3"), category=ErrorCategory.SYSTEM)

        summary = error_recovery_system.get_error_summary()

        assert summary["by_category"]["system"] == 2
        assert summary["by_category"]["research"] == 1

    def test_error_summary_limit(self, error_recovery_system):
        """Test error summary respects limit parameter"""
        for i in range(100):
            error_recovery_system.handle_error(ValueError(f"Error {i}"))

        summary = error_recovery_system.get_error_summary(limit=30)

        assert summary["total_recent_errors"] == 30

    def test_error_summary_recent_errors_list(self, error_recovery_system):
        """Test error summary includes recent errors list"""
        errors = [
            (ValueError("Error 1"), ErrorSeverity.HIGH),
            (RuntimeError("Error 2"), ErrorSeverity.CRITICAL),
            (TypeError("Error 3"), ErrorSeverity.MEDIUM),
        ]

        for error, severity in errors:
            error_recovery_system.handle_error(error, severity=severity)

        summary = error_recovery_system.get_error_summary()

        assert len(summary["recent_errors"]) == 3
        assert all("id" in e and "timestamp" in e for e in summary["recent_errors"])
        assert all("severity" in e and "category" in e for e in summary["recent_errors"])

    def test_error_summary_recovery_rate(self, error_recovery_system):
        """Test error summary calculates recovery rate"""
        error1 = ValueError("Error 1")
        report1 = error_recovery_system.handle_error(error1)
        report1.recovery_successful = True

        error2 = RuntimeError("Error 2")
        error_recovery_system.handle_error(error2)

        summary = error_recovery_system.get_error_summary()

        # 1 out of 2 recovered
        assert summary["recovery_rate"] == 0.5


class TestTroubleshootingGuide:
    """Tests for troubleshooting guide generation"""

    def test_generate_troubleshooting_guide_empty(self, error_recovery_system):
        """Test generating guide with no error history"""
        guide = error_recovery_system.generate_troubleshooting_guide()

        assert "generated_at" in guide
        assert "common_issues" in guide
        assert "recovery_procedures" in guide
        assert "prevention_tips" in guide
        assert "emergency_procedures" in guide
        assert isinstance(guide["common_issues"], list)
        assert isinstance(guide["recovery_procedures"], dict)
        assert isinstance(guide["prevention_tips"], list)
        assert isinstance(guide["emergency_procedures"], list)

    def test_generate_troubleshooting_guide_with_errors(self, error_recovery_system):
        """Test guide generation with error history"""
        # Create multiple errors of same type
        for i in range(3):
            error_recovery_system.handle_error(ValueError(f"Validation error {i}"), category=ErrorCategory.VALIDATION)

        guide = error_recovery_system.generate_troubleshooting_guide()

        # Should have recovery procedures
        assert len(guide["recovery_procedures"]) > 0

    def test_troubleshooting_guide_emergency_procedures(self, error_recovery_system):
        """Test that guide includes emergency procedures"""
        guide = error_recovery_system.generate_troubleshooting_guide()

        emergency_procs = guide["emergency_procedures"]
        assert len(emergency_procs) > 0
        assert all("condition" in p and "procedure" in p for p in emergency_procs)

    def test_troubleshooting_guide_recovery_procedures(self, error_recovery_system):
        """Test that guide includes all recovery procedures"""
        guide = error_recovery_system.generate_troubleshooting_guide()

        procedures = guide["recovery_procedures"]
        # Should have procedures from registered actions
        assert "restart_research_engines" in procedures
        assert "restore_config_backup" in procedures

    def test_troubleshooting_guide_prevention_tips(self, error_recovery_system):
        """Test prevention tips generation"""
        # Create multiple configuration errors
        for i in range(6):
            error_recovery_system.handle_error(ValueError(f"Config error {i}"), category=ErrorCategory.CONFIGURATION)

        guide = error_recovery_system.generate_troubleshooting_guide()

        # Should have prevention tip for configuration
        tips = guide["prevention_tips"]
        assert any("configuration" in tip.lower() for tip in tips)


class TestErrorCleanup:
    """Tests for error history cleanup"""

    def test_cleanup_old_errors(self, error_recovery_system):
        """Test cleaning up old error records"""
        # Add old error
        old_time = datetime.now(timezone.utc) - timedelta(days=40)
        old_error = ErrorReport(
            id="ERR_old_001",
            timestamp=old_time,
            severity=ErrorSeverity.LOW,
            category=ErrorCategory.USER_INPUT,
            message="Old error",
            details={},
            stack_trace=None,
            context={},
        )
        error_recovery_system.error_history.append(old_error)

        # Add recent error
        recent_error = ValueError("Recent error")
        error_recovery_system.handle_error(recent_error)

        assert len(error_recovery_system.error_history) == 2

        # Cleanup
        result = error_recovery_system.cleanup_old_errors(days_to_keep=30)

        assert result["removed_count"] == 1
        assert result["remaining_count"] == 1
        assert len(error_recovery_system.error_history) == 1

    def test_cleanup_returns_cutoff_date(self, error_recovery_system):
        """Test cleanup returns correct cutoff date"""
        result = error_recovery_system.cleanup_old_errors(days_to_keep=30)

        assert "cutoff_date" in result
        assert "removed_count" in result
        assert "remaining_count" in result

    def test_cleanup_no_old_errors(self, error_recovery_system):
        """Test cleanup when no old errors exist"""
        error_recovery_system.handle_error(ValueError("Recent error"))

        result = error_recovery_system.cleanup_old_errors(days_to_keep=30)

        assert result["removed_count"] == 0
        assert result["remaining_count"] == 1

    @patch("builtins.open", new_callable=mock_open)
    def test_cleanup_saves_history(self, mock_file, error_recovery_system):
        """Test that cleanup saves updated history"""
        error_recovery_system.handle_error(ValueError("Test error"))

        with patch.object(error_recovery_system, "_save_error_history") as mock_save:
            error_recovery_system.cleanup_old_errors(days_to_keep=30)
            mock_save.assert_called_once()


class TestPatternDetection:
    """Tests for error pattern detection"""

    def test_identify_error_patterns_empty(self, error_recovery_system):
        """Test pattern detection with no errors"""
        patterns = error_recovery_system._identify_error_patterns([])
        assert patterns == {}

    def test_identify_error_patterns_single_type(self, error_recovery_system):
        """Test pattern detection with same error type"""
        error1 = ValueError("Error 1")
        error2 = ValueError("Error 2")

        error_recovery_system.handle_error(error1, category=ErrorCategory.VALIDATION)
        error_recovery_system.handle_error(error2, category=ErrorCategory.VALIDATION)

        patterns = error_recovery_system._identify_error_patterns(error_recovery_system.error_history)

        assert len(patterns) > 0
        assert any("validation" in pattern for pattern in patterns.keys())

    def test_identify_error_patterns_multiple_types(self, error_recovery_system):
        """Test pattern detection with different error types"""
        errors = [
            (ValueError("Val 1"), ErrorCategory.VALIDATION),
            (RuntimeError("Run 1"), ErrorCategory.SYSTEM),
            (ValueError("Val 2"), ErrorCategory.VALIDATION),
        ]

        for error, category in errors:
            error_recovery_system.handle_error(error, category=category)

        patterns = error_recovery_system._identify_error_patterns(error_recovery_system.error_history)

        # Should have multiple patterns
        assert len(patterns) > 1

    def test_pattern_frequency_count(self, error_recovery_system):
        """Test that patterns track frequency"""
        # Create multiple same-type errors
        for i in range(3):
            error_recovery_system.handle_error(ValueError(f"Validation error {i}"), category=ErrorCategory.VALIDATION)

        patterns = error_recovery_system._identify_error_patterns(error_recovery_system.error_history)

        # Should have pattern with frequency 3
        assert any(count == 3 for count in patterns.values())


class TestAutomaticRecovery:
    """Tests for automatic recovery mechanisms"""

    def test_attempt_automatic_recovery_no_suitable_actions(self, error_recovery_system):
        """Test automatic recovery when no suitable actions exist"""
        error = ValueError("Test error")
        report = error_recovery_system.handle_error(
            error, severity=ErrorSeverity.LOW, category=ErrorCategory.USER_INPUT
        )

        result = error_recovery_system._attempt_automatic_recovery(report)

        assert result.success is False
        assert "no suitable" in result.message.lower()

    def test_attempt_automatic_recovery_finds_matching_action(self, error_recovery_system):
        """Test automatic recovery finds matching action"""
        # Remove default recovery action first to test our custom one
        error_recovery_system.recovery_actions.clear()

        handler = Mock(return_value=True)
        action = RecoveryAction(
            name="auto_test",
            description="Test automatic recovery",
            action_type="automatic",
            severity_filter=[ErrorSeverity.CRITICAL],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler,
        )
        error_recovery_system.register_recovery_action(action)

        RuntimeError("Critical system error")
        report = ErrorReport(
            id="ERR_test",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.CRITICAL,
            category=ErrorCategory.SYSTEM,
            message="Test error",
            details={},
            stack_trace=None,
            context={},
        )

        result = error_recovery_system._attempt_automatic_recovery(report)

        assert result.success is True
        assert result.action_name == "auto_test"

    def test_automatic_recovery_skips_manual_actions(self, error_recovery_system):
        """Test that automatic recovery skips manual actions"""
        # Remove default recovery actions first
        error_recovery_system.recovery_actions.clear()

        handler = Mock(return_value=True)
        action = RecoveryAction(
            name="manual_action",
            description="Manual recovery",
            action_type="manual",
            severity_filter=[ErrorSeverity.CRITICAL],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler,
        )
        error_recovery_system.register_recovery_action(action)

        report = ErrorReport(
            id="ERR_test",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.CRITICAL,
            category=ErrorCategory.SYSTEM,
            message="Critical error",
            details={},
            stack_trace=None,
            context={},
        )

        result = error_recovery_system._attempt_automatic_recovery(report)

        # Should not call manual action
        handler.assert_not_called()
        assert result.success is False

    def test_automatic_recovery_continues_on_action_failure(self, error_recovery_system):
        """Test that automatic recovery tries next action if one fails"""
        # Remove default recovery actions first
        error_recovery_system.recovery_actions.clear()

        handler1 = Mock(side_effect=RuntimeError("Action 1 failed"))
        action1 = RecoveryAction(
            name="failing_action",
            description="Fails",
            action_type="automatic",
            severity_filter=[ErrorSeverity.CRITICAL],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler1,
        )

        handler2 = Mock(return_value=True)
        action2 = RecoveryAction(
            name="success_action",
            description="Succeeds",
            action_type="automatic",
            severity_filter=[ErrorSeverity.CRITICAL],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler2,
        )

        error_recovery_system.register_recovery_action(action1)
        error_recovery_system.register_recovery_action(action2)

        report = ErrorReport(
            id="ERR_test",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.CRITICAL,
            category=ErrorCategory.SYSTEM,
            message="Critical error",
            details={},
            stack_trace=None,
            context={},
        )

        error_recovery_system._attempt_automatic_recovery(report)

        # Should try both actions, second one succeeds
        assert handler1.call_count == 1
        assert handler2.call_count == 1


class TestHelperMethods:
    """Tests for helper and utility methods"""

    def test_generate_error_id_format(self, error_recovery_system):
        """Test error ID generation format"""
        error_id = error_recovery_system._generate_error_id()

        assert error_id.startswith("ERR_")
        parts = error_id.split("_")
        assert len(parts) == 4  # ERR, date, time, random
        assert len(parts[3]) == 6  # Random suffix is 6 chars

    def test_generate_error_id_uniqueness(self, error_recovery_system):
        """Test that generated error IDs are unique"""
        ids = set()
        for _ in range(100):
            error_id = error_recovery_system._generate_error_id()
            assert error_id not in ids
            ids.add(error_id)

    def test_calculate_recovery_rate_no_errors(self, error_recovery_system):
        """Test recovery rate with no errors"""
        rate = error_recovery_system._calculate_recovery_rate([])
        assert rate == 0.0

    def test_calculate_recovery_rate_all_recovered(self, error_recovery_system):
        """Test recovery rate when all errors recovered"""
        error_recovery_system.handle_error(ValueError("Error 1"))
        error_recovery_system.handle_error(ValueError("Error 2"))
        error_recovery_system.handle_error(ValueError("Error 3"))

        # Mark all as recovered
        for error in error_recovery_system.error_history:
            error.recovery_successful = True

        rate = error_recovery_system._calculate_recovery_rate(error_recovery_system.error_history)
        assert rate == 1.0

    def test_calculate_recovery_rate_partial(self, error_recovery_system):
        """Test recovery rate with partial recovery"""
        for i in range(4):
            error_recovery_system.handle_error(ValueError(f"Error {i}"))

        # Mark 2 as recovered
        error_recovery_system.error_history[0].recovery_successful = True
        error_recovery_system.error_history[1].recovery_successful = True

        rate = error_recovery_system._calculate_recovery_rate(error_recovery_system.error_history)
        assert rate == 0.5

    def test_get_pattern_severity_mapping(self, error_recovery_system):
        """Test pattern severity mapping"""
        test_cases = [
            ("research:Exception", "high"),
            ("system:Exception", "critical"),
            ("configuration:Exception", "high"),
            ("communication:Exception", "medium"),
            ("unknown:Exception", "medium"),  # Default
        ]

        for pattern, expected_severity in test_cases:
            severity = error_recovery_system._get_pattern_severity(pattern)
            assert severity == expected_severity

    def test_get_solutions_for_pattern(self, error_recovery_system):
        """Test getting solutions for error patterns"""
        solutions = error_recovery_system._get_solutions_for_pattern("research:Exception")

        assert isinstance(solutions, list)
        assert len(solutions) > 0
        assert all(isinstance(s, str) for s in solutions)

    def test_generate_prevention_tips(self, error_recovery_system):
        """Test prevention tips generation"""
        # Create multiple configuration errors
        for i in range(6):
            error_recovery_system.handle_error(ValueError(f"Config error {i}"), category=ErrorCategory.CONFIGURATION)

        tips = error_recovery_system._generate_prevention_tips()

        assert isinstance(tips, list)

    def test_generate_emergency_procedures(self, error_recovery_system):
        """Test emergency procedures generation"""
        procedures = error_recovery_system._generate_emergency_procedures()

        assert isinstance(procedures, list)
        assert len(procedures) > 0
        assert all("condition" in p and "procedure" in p for p in procedures)


class TestConvenienceFunctions:
    """Tests for module-level convenience functions"""

    @patch.dict(os.environ, {"MOAI_PROJECT_ROOT": "/test/root"})
    def test_get_error_recovery_system_singleton(self, temp_project_dir):
        """Test that get_error_recovery_system returns singleton"""
        # Reset global instance for testing
        import moai_adk.core.error_recovery_system as ers_module

        ers_module._error_recovery_system = None

        system1 = get_error_recovery_system(temp_project_dir)
        system2 = get_error_recovery_system(temp_project_dir)

        assert system1 is system2

        # Cleanup
        system1.monitoring_active = False

    def test_handle_error_convenience_function(self, temp_project_dir):
        """Test convenience handle_error function"""
        import moai_adk.core.error_recovery_system as ers_module

        ers_module._error_recovery_system = None

        error = ValueError("Test error")
        report = handle_error(error, severity=ErrorSeverity.HIGH)

        assert report.severity == ErrorSeverity.HIGH
        assert report.message == "Test error"

        # Cleanup
        system = get_error_recovery_system()
        system.monitoring_active = False

    def test_error_handler_decorator(self, temp_project_dir):
        """Test error_handler decorator"""
        import moai_adk.core.error_recovery_system as ers_module

        ers_module._error_recovery_system = None

        @error_handler(severity=ErrorSeverity.HIGH, category=ErrorCategory.RESEARCH)
        def failing_function():
            raise ValueError("Decorated error")

        with pytest.raises(ValueError):
            failing_function()

        # Error should be logged
        system = get_error_recovery_system()
        assert len(system.error_history) > 0
        assert system.error_history[0].severity == ErrorSeverity.HIGH

        # Cleanup
        system.monitoring_active = False

    def test_error_handler_decorator_with_context(self, temp_project_dir):
        """Test error_handler decorator with context"""
        import moai_adk.core.error_recovery_system as ers_module

        ers_module._error_recovery_system = None

        @error_handler(severity=ErrorSeverity.CRITICAL, context={"operation": "test_operation"})
        def context_function():
            raise RuntimeError("Context error")

        with pytest.raises(RuntimeError):
            context_function()

        system = get_error_recovery_system()
        assert system.error_history[0].context["operation"] == "test_operation"

        # Cleanup
        system.monitoring_active = False

    def test_error_handler_decorator_success(self, temp_project_dir):
        """Test error_handler decorator with successful function"""
        import moai_adk.core.error_recovery_system as ers_module

        ers_module._error_recovery_system = None

        @error_handler()
        def successful_function(x, y):
            return x + y

        result = successful_function(5, 3)
        assert result == 8

        system = get_error_recovery_system()
        assert len(system.error_history) == 0

        # Cleanup
        system.monitoring_active = False


class TestValidationMethods:
    """Tests for file validation and repair methods"""

    def test_validate_skill_file_valid(self, temp_project_dir):
        """Test validating a valid skill file"""
        skill_file = temp_project_dir / ".claude" / "skills" / "test_skill.md"
        skill_file.write_text("---\nname: test\n---\n\nContent here" * 10)

        system = ErrorRecoverySystem(project_root=temp_project_dir)
        assert system._validate_skill_file(skill_file) is True
        system.monitoring_active = False

    def test_validate_skill_file_invalid(self, temp_project_dir):
        """Test validating an invalid skill file"""
        skill_file = temp_project_dir / ".claude" / "skills" / "bad_skill.md"
        skill_file.write_text("Invalid skill")

        system = ErrorRecoverySystem(project_root=temp_project_dir)
        assert system._validate_skill_file(skill_file) is False
        system.monitoring_active = False

    def test_validate_agent_file_valid(self, temp_project_dir):
        """Test validating a valid agent file"""
        agent_file = temp_project_dir / ".claude" / "agents" / "alfred" / "test_agent.md"
        agent_file.write_text("role: test_role\n" + "content" * 50)

        system = ErrorRecoverySystem(project_root=temp_project_dir)
        assert system._validate_agent_file(agent_file) is True
        system.monitoring_active = False

    def test_validate_agent_file_invalid(self, temp_project_dir):
        """Test validating an invalid agent file"""
        agent_file = temp_project_dir / ".claude" / "agents" / "alfred" / "bad_agent.md"
        agent_file.write_text("Invalid")

        system = ErrorRecoverySystem(project_root=temp_project_dir)
        assert system._validate_agent_file(agent_file) is False
        system.monitoring_active = False

    def test_validate_command_file_valid(self, temp_project_dir):
        """Test validating a valid command file"""
        cmd_file = temp_project_dir / ".claude" / "commands" / "alfred" / "test_cmd.md"
        cmd_file.write_text("name: test_command\nallowed-tools: [Read, Write]")

        system = ErrorRecoverySystem(project_root=temp_project_dir)
        assert system._validate_command_file(cmd_file) is True
        system.monitoring_active = False

    def test_validate_command_file_invalid(self, temp_project_dir):
        """Test validating an invalid command file"""
        cmd_file = temp_project_dir / ".claude" / "commands" / "alfred" / "bad_cmd.md"
        cmd_file.write_text("Invalid command")

        system = ErrorRecoverySystem(project_root=temp_project_dir)
        assert system._validate_command_file(cmd_file) is False
        system.monitoring_active = False

    def test_repair_skill_file(self, temp_project_dir):
        """Test repairing a skill file"""
        skill_file = temp_project_dir / ".claude" / "skills" / "repair_test.md"
        skill_file.write_text("Content without header")

        system = ErrorRecoverySystem(project_root=temp_project_dir)
        result = system._repair_skill_file(skill_file)

        assert result is True
        content = skill_file.read_text()
        assert content.startswith("---")
        system.monitoring_active = False


class TestBackgroundMonitoring:
    """Tests for background monitoring functionality"""

    def test_background_monitoring_thread_runs(self, error_recovery_system):
        """Test that background monitoring thread starts"""
        assert error_recovery_system.monitor_thread.is_alive()

    def test_check_error_patterns_high_rate(self, error_recovery_system):
        """Test checking for high error rate"""
        # Create many recent errors
        current_time = datetime.now(timezone.utc)
        for i in range(12):
            error_time = current_time - timedelta(seconds=i * 10)
            error = ErrorReport(
                id=f"ERR_test_{i}",
                timestamp=error_time,
                severity=ErrorSeverity.MEDIUM,
                category=ErrorCategory.SYSTEM,
                message=f"Error {i}",
                details={},
                stack_trace=None,
                context={},
            )
            error_recovery_system.error_history.append(error)

        # Should detect high error rate
        with patch("logging.Logger.warning") as mock_warn:
            error_recovery_system._check_error_patterns()
            # Check if warning was called for high error rate
            assert mock_warn.call_count > 0

    def test_check_repeated_errors(self, error_recovery_system):
        """Test detection of repeated error messages"""
        current_time = datetime.now(timezone.utc)
        error_message = "Repeated error"

        for i in range(4):
            error_time = current_time - timedelta(seconds=i * 10)
            error = ErrorReport(
                id=f"ERR_repeat_{i}",
                timestamp=error_time,
                severity=ErrorSeverity.MEDIUM,
                category=ErrorCategory.SYSTEM,
                message=error_message,
                details={},
                stack_trace=None,
                context={},
            )
            error_recovery_system.error_history.append(error)

        # Should detect repeated errors
        with patch("logging.Logger.warning") as mock_warn:
            error_recovery_system._check_error_patterns()
            assert mock_warn.call_count > 0


class TestEdgeCases:
    """Tests for edge cases and boundary conditions"""

    def test_handle_error_with_none_context(self, error_recovery_system):
        """Test error handling with None context"""
        error = ValueError("Test error")
        report = error_recovery_system.handle_error(error, context=None)

        assert report.context == {}

    def test_handle_error_empty_message(self, error_recovery_system):
        """Test error handling with empty message"""
        error = ValueError("")
        report = error_recovery_system.handle_error(error)

        assert report.message == ""

    def test_recovery_result_none_duration(self, error_recovery_system):
        """Test recovery result with zero duration"""
        result = RecoveryResult(
            success=False,
            action_name="test",
            message="Failed",
            duration=0.0,
        )

        assert result.duration == 0.0

    def test_multiple_concurrent_errors(self, error_recovery_system):
        """Test handling multiple errors concurrently"""
        import concurrent.futures

        def create_error(i):
            error = ValueError(f"Concurrent error {i}")
            return error_recovery_system.handle_error(error)

        with concurrent.futures.ThreadPoolExecutor(max_workers=5) as executor:
            futures = [executor.submit(create_error, i) for i in range(10)]
            results = [f.result() for f in concurrent.futures.as_completed(futures)]

        assert len(error_recovery_system.error_history) == 10
        assert len(results) == 10

    def test_very_long_error_message(self, error_recovery_system):
        """Test handling very long error messages"""
        long_message = "Error: " + "x" * 10000
        error = ValueError(long_message)
        report = error_recovery_system.handle_error(error)

        assert report.message == long_message

    def test_error_with_special_characters(self, error_recovery_system):
        """Test error handling with special characters"""
        special_message = "Error with special chars: 中文, 한글, العربية, 🔥"
        error = ValueError(special_message)
        report = error_recovery_system.handle_error(error)

        assert report.message == special_message

    def test_recovery_with_large_parameters(self, error_recovery_system):
        """Test manual recovery with large parameters"""
        error = ValueError("Test error")
        report = error_recovery_system.handle_error(error)
        error_id = report.id

        handler = Mock(return_value=True)
        action = RecoveryAction(
            name="large_params",
            description="Test",
            action_type="manual",
            severity_filter=[ErrorSeverity.MEDIUM],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler,
        )
        error_recovery_system.register_recovery_action(action)

        # Large parameters
        large_params = {"data": "x" * 10000}
        result = error_recovery_system.attempt_manual_recovery(error_id, "large_params", large_params)

        assert result.success is True


class TestIntegration:
    """Integration tests for multiple components"""

    def test_full_error_lifecycle(self, error_recovery_system):
        """Test complete error lifecycle"""
        # 1. Create and handle error (patch recovery so error stays in active_errors)
        with patch.object(error_recovery_system, "_attempt_automatic_recovery") as mock_recovery:
            mock_recovery.return_value = RecoveryResult(
                success=False,
                action_name="test",
                message="Recovery failed",
                duration=1.0,
            )
            error = RuntimeError("Lifecycle test error")
            report = error_recovery_system.handle_error(
                error, severity=ErrorSeverity.HIGH, category=ErrorCategory.SYSTEM
            )
            error_id = report.id

        # 2. Verify error is tracked
        assert error_id in error_recovery_system.active_errors
        assert len(error_recovery_system.error_history) == 1

        # 3. Register and attempt recovery
        handler = Mock(return_value=True)
        action = RecoveryAction(
            name="lifecycle_recovery",
            description="Lifecycle recovery",
            action_type="manual",
            severity_filter=[ErrorSeverity.HIGH],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler,
        )
        error_recovery_system.register_recovery_action(action)

        result = error_recovery_system.attempt_manual_recovery(error_id, "lifecycle_recovery")

        # 4. Verify recovery
        assert result.success is True
        assert error_id not in error_recovery_system.active_errors
        assert error_recovery_system.error_history[0].recovery_successful is True

        # 5. Verify system health updated
        health = error_recovery_system.get_system_health()
        assert health["status"] == "healthy"

    def test_multiple_errors_recovery_flow(self, error_recovery_system):
        """Test handling multiple errors with recovery"""
        # Handle multiple errors
        errors = [
            (ValueError("Error 1"), ErrorSeverity.CRITICAL, ErrorCategory.SYSTEM),
            (RuntimeError("Error 2"), ErrorSeverity.HIGH, ErrorCategory.RESEARCH),
            (TypeError("Error 3"), ErrorSeverity.MEDIUM, ErrorCategory.VALIDATION),
        ]

        reports = []
        for error, severity, category in errors:
            with patch.object(error_recovery_system, "_attempt_automatic_recovery") as mock_recovery:
                mock_recovery.return_value = RecoveryResult(
                    success=True,
                    action_name="auto_recovery",
                    message="Auto recovered",
                    duration=1.0,
                )
                report = error_recovery_system.handle_error(error, severity=severity, category=category)
                reports.append(report)

        # Verify all handled
        assert len(error_recovery_system.error_history) == 3

        # Check summary
        summary = error_recovery_system.get_error_summary()
        assert summary["total_recent_errors"] == 3
        assert summary["recovery_rate"] > 0

    def test_system_degradation_detection(self, error_recovery_system):
        """Test system health degradation detection"""
        # Start healthy
        health = error_recovery_system.get_system_health()
        assert health["status"] == "healthy"

        # Add high severity error (patch recovery to fail so error stays)
        with patch.object(error_recovery_system, "_attempt_automatic_recovery") as mock_recovery:
            mock_recovery.return_value = RecoveryResult(
                success=False,
                action_name="test",
                message="Recovery failed",
                duration=1.0,
            )
            error_recovery_system.handle_error(RuntimeError("High error"), severity=ErrorSeverity.HIGH)

        health = error_recovery_system.get_system_health()
        assert health["status"] == "degraded"

        # Add critical error
        with patch.object(error_recovery_system, "_attempt_automatic_recovery") as mock_recovery:
            mock_recovery.return_value = RecoveryResult(
                success=False,
                action_name="test",
                message="Recovery failed",
                duration=1.0,
            )
            error_recovery_system.handle_error(RuntimeError("Critical error"), severity=ErrorSeverity.CRITICAL)

        health = error_recovery_system.get_system_health()
        assert health["status"] == "critical"
