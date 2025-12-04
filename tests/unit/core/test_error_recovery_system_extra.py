"""
Extended tests for error_recovery_system module.

Tests enums, dataclasses, error handling, recovery actions, and system health monitoring.
"""

import asyncio
import tempfile
from datetime import datetime, timedelta, timezone
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.core.error_recovery_system import (
    AdvancedRecoveryAction,
    ConsistencyLevel,
    ErrorCategory,
    ErrorRecoverySystem,
    ErrorReport,
    ErrorSeverity,
    FailureEvent,
    FailureMode,
    RecoveryAction,
    RecoveryResult,
    RecoveryStatus,
    RecoveryStrategy,
    SystemSnapshot,
    error_handler,
    get_error_recovery_system,
    handle_error,
)


class TestEnums:
    """Test error recovery enums."""

    def test_error_severity_values(self):
        """Test ErrorSeverity enum values."""
        assert ErrorSeverity.CRITICAL.value == "critical"
        assert ErrorSeverity.HIGH.value == "high"
        assert ErrorSeverity.MEDIUM.value == "medium"
        assert ErrorSeverity.LOW.value == "low"
        assert ErrorSeverity.INFO.value == "info"

    def test_error_category_values(self):
        """Test ErrorCategory enum values."""
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

    def test_failure_mode_values(self):
        """Test FailureMode enum values."""
        assert FailureMode.HOOK_EXECUTION_FAILURE.value == "hook_execution_failure"
        assert FailureMode.RESOURCE_EXHAUSTION.value == "resource_exhaustion"
        assert FailureMode.DATA_CORRUPTION.value == "data_corruption"
        assert FailureMode.CASCADE_FAILURE.value == "cascade_failure"

    def test_recovery_strategy_values(self):
        """Test RecoveryStrategy enum values."""
        assert RecoveryStrategy.RETRY_WITH_BACKOFF.value == "retry_with_backoff"
        assert RecoveryStrategy.CIRCUIT_BREAKER.value == "circuit_breaker"
        assert RecoveryStrategy.ROLLBACK.value == "rollback"
        assert RecoveryStrategy.FAILOVER.value == "failover"

    def test_recovery_status_values(self):
        """Test RecoveryStatus enum values."""
        assert RecoveryStatus.PENDING.value == "pending"
        assert RecoveryStatus.IN_PROGRESS.value == "in_progress"
        assert RecoveryStatus.COMPLETED.value == "completed"
        assert RecoveryStatus.FAILED.value == "failed"

    def test_consistency_level_values(self):
        """Test ConsistencyLevel enum values."""
        assert ConsistencyLevel.STRONG.value == "strong"
        assert ConsistencyLevel.EVENTUAL.value == "eventual"
        assert ConsistencyLevel.WEAK.value == "weak"
        assert ConsistencyLevel.CUSTOM.value == "custom"


class TestErrorReportDataclass:
    """Test ErrorReport dataclass."""

    def test_error_report_creation(self):
        """Test ErrorReport creation."""
        report = ErrorReport(
            id="ERR_001",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.HIGH,
            category=ErrorCategory.SYSTEM,
            message="Test error",
            details={"code": "TEST_001"},
            stack_trace="traceback...",
            context={"module": "test"},
            recovery_attempted=False,
            recovery_successful=False,
        )
        assert report.id == "ERR_001"
        assert report.severity == ErrorSeverity.HIGH

    def test_error_report_defaults(self):
        """Test ErrorReport default values."""
        report = ErrorReport(
            id="ERR_001",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.MEDIUM,
            category=ErrorCategory.SYSTEM,
            message="Test",
            details={},
            stack_trace=None,
            context={},
        )
        assert report.recovery_attempted is False
        assert report.recovery_successful is False
        assert report.resolution_message is None


class TestRecoveryActionDataclass:
    """Test RecoveryAction dataclass."""

    def test_recovery_action_creation(self):
        """Test RecoveryAction creation."""
        handler = lambda e, p: True
        action = RecoveryAction(
            name="test_recovery",
            description="Test recovery action",
            action_type="automatic",
            severity_filter=[ErrorSeverity.HIGH, ErrorSeverity.CRITICAL],
            category_filter=[ErrorCategory.SYSTEM],
            handler=handler,
            timeout=30.0,
            max_attempts=3,
        )
        assert action.name == "test_recovery"
        assert action.timeout == 30.0


class TestFailureEventDataclass:
    """Test FailureEvent dataclass."""

    def test_failure_event_creation(self):
        """Test FailureEvent creation."""
        event = FailureEvent(
            failure_id="FAIL_001",
            failure_mode=FailureMode.RESOURCE_EXHAUSTION,
            timestamp=datetime.now(timezone.utc),
            component="test_component",
            description="Test failure",
            severity="high",
            context={"data": "value"},
        )
        assert event.failure_id == "FAIL_001"
        assert event.failure_mode == FailureMode.RESOURCE_EXHAUSTION

    def test_failure_event_to_dict(self):
        """Test FailureEvent serialization."""
        event = FailureEvent(
            failure_id="FAIL_001",
            failure_mode=FailureMode.TIMEOUT_FAILURE,
            timestamp=datetime.now(timezone.utc),
            component="test",
            description="Test",
            severity="medium",
        )
        result = event.to_dict()
        assert result["failure_id"] == "FAIL_001"
        assert isinstance(result["timestamp"], str)

    def test_failure_event_from_dict(self):
        """Test FailureEvent deserialization."""
        data = {
            "failure_id": "FAIL_001",
            "failure_mode": "resource_exhaustion",
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "component": "test",
            "description": "Test",
            "severity": "high",
            "context": {},
            "error_details": None,
            "affected_operations": [],
            "auto_recovery_eligible": True,
            "retry_count": 0,
            "metadata": {},
            "parent_failure_id": None,
            "root_cause": None,
        }
        event = FailureEvent.from_dict(data)
        assert event.failure_id == "FAIL_001"


class TestAdvancedRecoveryActionDataclass:
    """Test AdvancedRecoveryAction dataclass."""

    def test_advanced_recovery_action_creation(self):
        """Test AdvancedRecoveryAction creation."""
        action = AdvancedRecoveryAction(
            action_id="ACT_001",
            failure_id="FAIL_001",
            strategy=RecoveryStrategy.RETRY_WITH_BACKOFF,
            timestamp=datetime.now(timezone.utc),
            description="Test action",
            priority=5,
        )
        assert action.action_id == "ACT_001"
        assert action.status == RecoveryStatus.PENDING

    def test_advanced_recovery_action_to_dict(self):
        """Test AdvancedRecoveryAction serialization."""
        action = AdvancedRecoveryAction(
            action_id="ACT_001",
            failure_id="FAIL_001",
            strategy=RecoveryStrategy.CIRCUIT_BREAKER,
            timestamp=datetime.now(timezone.utc),
        )
        result = action.to_dict()
        assert result["action_id"] == "ACT_001"
        assert result["strategy"] == "circuit_breaker"


class TestSystemSnapshotDataclass:
    """Test SystemSnapshot dataclass."""

    def test_system_snapshot_creation(self):
        """Test SystemSnapshot creation."""
        snapshot = SystemSnapshot(
            snapshot_id="SNAP_001",
            timestamp=datetime.now(timezone.utc),
            component_states={"component1": {"state": "active"}},
            configuration_hash="abc123",
            data_checksums={"data1": "hash1"},
            description="Test snapshot",
        )
        assert snapshot.snapshot_id == "SNAP_001"
        assert snapshot.is_rollback_point is False

    def test_system_snapshot_to_dict(self):
        """Test SystemSnapshot serialization."""
        snapshot = SystemSnapshot(
            snapshot_id="SNAP_001",
            timestamp=datetime.now(timezone.utc),
            component_states={},
            configuration_hash="hash",
            data_checksums={},
        )
        result = snapshot.to_dict()
        assert result["snapshot_id"] == "SNAP_001"
        assert isinstance(result["timestamp"], str)


class TestErrorRecoverySystemInit:
    """Test ErrorRecoverySystem initialization."""

    def test_system_init_default(self):
        """Test system initialization with default path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            assert system.error_log_dir.exists()
            assert isinstance(system.active_errors, dict)
            assert isinstance(system.error_history, list)

    def test_system_init_creates_directories(self):
        """Test system creates necessary directories."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            assert (Path(tmpdir) / ".moai" / "error_logs").exists()

    def test_system_monitoring_thread(self):
        """Test monitoring thread starts."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            assert system.monitoring_active is True
            assert system.monitor_thread.is_alive()


class TestErrorHandling:
    """Test error handling functionality."""

    def test_handle_error_basic(self):
        """Test basic error handling."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            error = ValueError("Test error")
            report = system.handle_error(error)

            assert report.id is not None
            assert report.severity == ErrorSeverity.MEDIUM
            assert report.message == "Test error"

    def test_handle_error_with_context(self):
        """Test error handling with context."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            error = RuntimeError("Test")
            report = system.handle_error(
                error,
                context={"module": "test"},
                severity=ErrorSeverity.HIGH,
                category=ErrorCategory.SYSTEM,
            )

            assert report.context["module"] == "test"
            assert report.severity == ErrorSeverity.HIGH

    def test_handle_critical_error_attempts_recovery(self):
        """Test critical errors trigger recovery."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            error = RuntimeError("Critical error")
            report = system.handle_error(
                error,
                severity=ErrorSeverity.CRITICAL,
                category=ErrorCategory.SYSTEM,
            )

            assert report.recovery_attempted is True


class TestRecoveryActions:
    """Test recovery action functionality."""

    def test_register_recovery_action(self):
        """Test registering recovery action."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            handler = lambda e, p: True
            action = RecoveryAction(
                name="test_action",
                description="Test",
                action_type="automatic",
                severity_filter=[ErrorSeverity.HIGH],
                category_filter=[ErrorCategory.SYSTEM],
                handler=handler,
            )
            system.register_recovery_action(action)

            assert "test_action" in system.recovery_actions

    def test_attempt_manual_recovery(self):
        """Test manual recovery attempt."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            error = ValueError("Test")
            report = system.handle_error(error)

            handler = lambda e, p: True
            action = RecoveryAction(
                name="test_action",
                description="Test",
                action_type="manual",
                severity_filter=[ErrorSeverity.MEDIUM],
                category_filter=[ErrorCategory.SYSTEM],
                handler=handler,
            )
            system.register_recovery_action(action)

            result = system.attempt_manual_recovery(report.id, "test_action")
            assert isinstance(result, RecoveryResult)


class TestSystemHealth:
    """Test system health monitoring."""

    def test_get_system_health(self):
        """Test getting system health status."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            health = system.get_system_health()

            assert "status" in health
            assert "active_errors" in health
            assert "error_stats" in health

    def test_system_health_status_healthy(self):
        """Test healthy system status."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            health = system.get_system_health()
            assert health["status"] == "healthy"

    def test_system_health_status_degraded(self):
        """Test degraded system status."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            # Add critical error
            error = ValueError("Critical")
            system.handle_error(error, severity=ErrorSeverity.CRITICAL)

            health = system.get_system_health()
            assert health["status"] in ["critical", "degraded"]


class TestErrorAnalysis:
    """Test error analysis functionality."""

    def test_get_error_summary(self):
        """Test getting error summary."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            for i in range(3):
                error = ValueError(f"Error {i}")
                system.handle_error(error)

            summary = system.get_error_summary()
            assert "by_severity" in summary
            assert "by_category" in summary

    def test_get_error_summary_limit(self):
        """Test error summary with limit."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            for i in range(100):
                error = ValueError(f"Error {i}")
                system.handle_error(error)

            summary = system.get_error_summary(limit=10)
            assert len(summary["recent_errors"]) <= 10


class TestTroubleshooting:
    """Test troubleshooting guide generation."""

    def test_generate_troubleshooting_guide(self):
        """Test troubleshooting guide generation."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            for i in range(3):
                error = ValueError("Repeated error")
                system.handle_error(error)

            guide = system.generate_troubleshooting_guide()
            assert "generated_at" in guide
            assert "common_issues" in guide
            assert "recovery_procedures" in guide

    def test_cleanup_old_errors(self):
        """Test cleaning up old error records."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            error = ValueError("Old error")
            system.handle_error(error)

            result = system.cleanup_old_errors(days_to_keep=0)
            assert "removed_count" in result


class TestStatistics:
    """Test error statistics."""

    def test_error_stats_tracking(self):
        """Test error statistics are tracked."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            initial_count = system.error_stats["total_errors"]

            error = ValueError("Test")
            system.handle_error(error, severity=ErrorSeverity.HIGH)

            assert system.error_stats["total_errors"] == initial_count + 1

    def test_stats_by_severity(self):
        """Test stats breakdown by severity."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            error = ValueError("Test")
            system.handle_error(error, severity=ErrorSeverity.CRITICAL)

            assert "critical" in system.error_stats["by_severity"]

    def test_stats_by_category(self):
        """Test stats breakdown by category."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            error = ValueError("Test")
            system.handle_error(error, category=ErrorCategory.SYSTEM)

            assert "system" in system.error_stats["by_category"]


class TestPhase3Features:
    """Test Phase 3 advanced recovery features."""

    @pytest.mark.asyncio
    async def test_report_advanced_failure(self):
        """Test reporting advanced failure."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            failure_id = await system.report_advanced_failure(
                failure_mode=FailureMode.RESOURCE_EXHAUSTION,
                component="test_component",
                description="Test failure",
                severity="high",
            )

            assert failure_id is not None
            assert failure_id in system.advanced_failures

    @pytest.mark.asyncio
    async def test_create_system_snapshot(self):
        """Test creating system snapshot."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            snapshot_id = await system._create_system_snapshot("Test snapshot")

            assert snapshot_id is not None
            assert snapshot_id in system.system_snapshots

    def test_get_advanced_system_status(self):
        """Test getting advanced system status."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            status = system.get_advanced_system_status()

            assert "status" in status
            assert "phase3_features" in status
            assert "advanced_recovery_statistics" in status


class TestGlobalFunctions:
    """Test module-level convenience functions."""

    def test_get_error_recovery_system_singleton(self):
        """Test get_error_recovery_system returns singleton."""
        system1 = get_error_recovery_system()
        system2 = get_error_recovery_system()
        # Should be same instance
        assert isinstance(system1, ErrorRecoverySystem)

    def test_handle_error_function(self):
        """Test handle_error convenience function."""
        error = ValueError("Test")
        report = handle_error(error)
        assert isinstance(report, ErrorReport)

    def test_error_handler_decorator(self):
        """Test error_handler decorator."""

        @error_handler(severity=ErrorSeverity.HIGH)
        def test_function():
            raise ValueError("Test error")

        with pytest.raises(ValueError):
            test_function()


class TestEdgeCases:
    """Test edge cases and error conditions."""

    def test_handle_error_with_exception_hierarchy(self):
        """Test handling errors from exception hierarchy."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))

            # Test different exception types
            exceptions = [
                ValueError("Value"),
                RuntimeError("Runtime"),
                Exception("Generic"),
                KeyError("Key"),
            ]

            for exc in exceptions:
                report = system.handle_error(exc)
                assert report.id is not None

    def test_very_large_error_history(self):
        """Test handling very large error history."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))

            # Add many errors
            for i in range(200):
                error = ValueError(f"Error {i}")
                system.handle_error(error)

            assert len(system.error_history) == 200

    def test_empty_error_context(self):
        """Test with empty context."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            error = ValueError("Test")
            report = system.handle_error(error, context={})
            assert report.context == {}

    def test_none_stack_trace(self):
        """Test error with no stack trace."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(project_root=Path(tmpdir))
            error = Exception("Test")
            # Handle outside try/except to test None stack trace handling
            report = system.handle_error(error)
            assert report.stack_trace is not None  # Still captured


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
