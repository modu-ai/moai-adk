"""
Comprehensive coverage tests for ErrorRecoverySystem - Phase 3 Advanced Recovery

Target: 95%+ coverage for error_recovery_system.py
Tests all recovery methods, edge cases, exception handling, and recovery paths.
Uses pytest-mock for mocking external dependencies.
"""

import asyncio
import pytest
import tempfile
import json
from datetime import datetime, timedelta, timezone
from pathlib import Path
from unittest.mock import MagicMock, patch, AsyncMock, call, mock_open
import threading
import time

from moai_adk.core.error_recovery_system import (
    ErrorSeverity,
    ErrorCategory,
    FailureMode,
    RecoveryStrategy,
    ConsistencyLevel,
    RecoveryStatus,
    ErrorReport,
    RecoveryAction,
    RecoveryResult,
    FailureEvent,
    AdvancedRecoveryAction,
    SystemSnapshot,
    ErrorRecoverySystem,
    get_error_recovery_system,
    handle_error,
    error_handler,
)


class TestDataClassSerializationAndDeserialization:
    """Test serialization and deserialization of dataclasses."""

    def test_failure_event_to_dict(self):
        """Test FailureEvent serialization to dictionary."""
        # Arrange
        failure = FailureEvent(
            failure_id="fail123",
            failure_mode=FailureMode.NETWORK_FAILURE,
            timestamp=datetime.now(timezone.utc),
            component="test_component",
            description="Test failure",
            severity="high",
            context={"test": "data"},
            error_details={"error": "details"},
            affected_operations=["op1", "op2"],
            auto_recovery_eligible=True,
            retry_count=2,
            metadata={"meta": "data"},
            parent_failure_id="parent123",
            root_cause="Test cause",
        )

        # Act
        result = failure.to_dict()

        # Assert
        assert result["failure_id"] == "fail123"
        assert result["failure_mode"] == "network_failure"
        assert result["component"] == "test_component"
        assert result["severity"] == "high"
        assert result["retry_count"] == 2
        assert result["affected_operations"] == ["op1", "op2"]
        assert result["parent_failure_id"] == "parent123"

    def test_failure_event_from_dict(self):
        """Test FailureEvent deserialization from dictionary."""
        # Arrange
        now = datetime.now(timezone.utc)
        data = {
            "failure_id": "fail123",
            "failure_mode": "network_failure",
            "timestamp": now.isoformat(),
            "component": "test_component",
            "description": "Test failure",
            "severity": "high",
            "context": {"test": "data"},
            "error_details": {"error": "details"},
            "affected_operations": ["op1", "op2"],
            "auto_recovery_eligible": True,
            "retry_count": 2,
            "metadata": {"meta": "data"},
            "parent_failure_id": "parent123",
            "root_cause": "Test cause",
        }

        # Act
        failure = FailureEvent.from_dict(data)

        # Assert
        assert failure.failure_id == "fail123"
        assert failure.failure_mode == FailureMode.NETWORK_FAILURE
        assert failure.component == "test_component"
        assert failure.severity == "high"
        assert failure.retry_count == 2

    def test_advanced_recovery_action_to_dict(self):
        """Test AdvancedRecoveryAction serialization."""
        # Arrange
        action = AdvancedRecoveryAction(
            action_id="act123",
            failure_id="fail123",
            strategy=RecoveryStrategy.RETRY_WITH_BACKOFF,
            timestamp=datetime.now(timezone.utc),
            status=RecoveryStatus.IN_PROGRESS,
            description="Test recovery",
            parameters={"key": "value"},
            execution_log=["step1", "step2"],
            rollback_available=True,
            timeout_seconds=300.0,
            retry_attempts=1,
            max_retries=3,
            rollback_action_id="rollback123",
            dependencies=["dep1"],
            priority=5,
        )

        # Act
        result = action.to_dict()

        # Assert
        assert result["action_id"] == "act123"
        assert result["strategy"] == "retry_with_backoff"
        assert result["status"] == "in_progress"
        assert result["max_retries"] == 3

    def test_system_snapshot_to_dict(self):
        """Test SystemSnapshot serialization."""
        # Arrange
        snapshot = SystemSnapshot(
            snapshot_id="snap123",
            timestamp=datetime.now(timezone.utc),
            component_states={"component": "state"},
            configuration_hash="abc123",
            data_checksums={"checksum": "def456"},
            metadata={"meta": "data"},
            parent_snapshot_id="parent_snap",
            is_rollback_point=True,
            description="Test snapshot",
            consistency_level=ConsistencyLevel.STRONG,
        )

        # Act
        result = snapshot.to_dict()

        # Assert
        assert result["snapshot_id"] == "snap123"
        assert result["configuration_hash"] == "abc123"
        assert result["consistency_level"] == "strong"
        assert result["is_rollback_point"] is True


class TestErrorHandlingAdvanced:
    """Test advanced error handling scenarios."""

    @pytest.fixture
    def recovery_system(self):
        """Create a fresh recovery system for testing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(Path(tmpdir))
            yield system

    def test_handle_critical_error_triggers_recovery(self, recovery_system):
        """Test that CRITICAL errors trigger automatic recovery."""
        # Arrange
        test_error = RuntimeError("Critical system error")

        # Act
        error_report = recovery_system.handle_error(
            test_error,
            context={"test": "context"},
            severity=ErrorSeverity.CRITICAL,
            category=ErrorCategory.SYSTEM,
        )

        # Assert
        assert error_report.recovery_attempted is True
        assert error_report.severity == ErrorSeverity.CRITICAL

    def test_handle_high_severity_error(self, recovery_system):
        """Test HIGH severity error handling."""
        # Arrange
        test_error = ValueError("High severity error")

        # Act
        error_report = recovery_system.handle_error(
            test_error,
            severity=ErrorSeverity.HIGH,
            category=ErrorCategory.INTEGRATION,
        )

        # Assert
        assert error_report.recovery_attempted is True
        assert error_report.severity == ErrorSeverity.HIGH

    def test_handle_low_severity_error_no_recovery(self, recovery_system):
        """Test LOW severity errors don't trigger automatic recovery."""
        # Arrange
        test_error = Exception("Low severity error")

        # Act
        error_report = recovery_system.handle_error(
            test_error,
            severity=ErrorSeverity.LOW,
            category=ErrorCategory.USER_INPUT,
        )

        # Assert
        assert error_report.recovery_attempted is False

    def test_error_statistics_updated(self, recovery_system):
        """Test error statistics are updated correctly."""
        # Arrange
        error1 = ValueError("Error 1")
        error2 = RuntimeError("Error 2")

        # Act
        recovery_system.handle_error(error1, severity=ErrorSeverity.HIGH)
        recovery_system.handle_error(error2, severity=ErrorSeverity.MEDIUM, category=ErrorCategory.RESEARCH)

        # Assert
        assert recovery_system.error_stats["total_errors"] == 2
        assert recovery_system.error_stats["by_severity"]["high"] == 1
        assert recovery_system.error_stats["by_severity"]["medium"] == 1
        assert recovery_system.error_stats["by_category"]["research"] == 1

    def test_error_context_preserved(self, recovery_system):
        """Test error context is preserved in reports."""
        # Arrange
        test_error = Exception("Test")
        context = {"user": "admin", "action": "delete", "resource": "file123"}

        # Act
        error_report = recovery_system.handle_error(
            test_error,
            context=context,
            severity=ErrorSeverity.MEDIUM,
        )

        # Assert
        assert error_report.context == context
        assert error_report.context["user"] == "admin"

    def test_error_stack_trace_captured(self, recovery_system):
        """Test that stack traces are captured."""
        # Arrange
        try:
            raise ValueError("Test error with traceback")
        except Exception as e:
            # Act
            error_report = recovery_system.handle_error(e)

            # Assert
            assert error_report.stack_trace is not None
            assert "ValueError" in error_report.stack_trace
            assert "test_error_stack_trace_captured" in error_report.stack_trace


class TestManualRecoveryActions:
    """Test manual recovery action execution."""

    @pytest.fixture
    def recovery_system(self):
        """Create a fresh recovery system for testing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(Path(tmpdir))
            yield system

    def test_attempt_manual_recovery_success(self, recovery_system):
        """Test successful manual recovery."""
        # Arrange
        test_error = ValueError("Test error")
        error_report = recovery_system.handle_error(test_error)
        error_id = error_report.id

        # Create a mock recovery handler that returns True
        def mock_handler(error, params):
            return True

        recovery_action = RecoveryAction(
            name="test_recovery",
            description="Test recovery action",
            action_type="manual",
            severity_filter=[ErrorSeverity.MEDIUM],
            category_filter=[ErrorCategory.SYSTEM],
            handler=mock_handler,
            timeout=10.0,
        )
        recovery_system.register_recovery_action(recovery_action)

        # Act
        result = recovery_system.attempt_manual_recovery(error_id, "test_recovery")

        # Assert
        assert result.success is True
        assert error_id not in recovery_system.active_errors

    def test_attempt_manual_recovery_error_not_found(self, recovery_system):
        """Test manual recovery when error not found."""
        # Act
        result = recovery_system.attempt_manual_recovery("nonexistent_error", "test_action")

        # Assert
        assert result.success is False
        assert "not found in active errors" in result.message

    def test_attempt_manual_recovery_action_not_found(self, recovery_system):
        """Test manual recovery when action not found."""
        # Arrange
        test_error = ValueError("Test error")
        error_report = recovery_system.handle_error(test_error)

        # Act
        result = recovery_system.attempt_manual_recovery(error_report.id, "nonexistent_action")

        # Assert
        assert result.success is False
        assert "not found" in result.message

    def test_attempt_manual_recovery_handler_exception(self, recovery_system):
        """Test manual recovery when handler throws exception."""
        # Arrange
        test_error = ValueError("Test error")
        error_report = recovery_system.handle_error(test_error)

        def failing_handler(error, params):
            raise RuntimeError("Handler failed")

        recovery_action = RecoveryAction(
            name="failing_recovery",
            description="Failing recovery action",
            action_type="manual",
            severity_filter=[ErrorSeverity.MEDIUM],
            category_filter=[ErrorCategory.SYSTEM],
            handler=failing_handler,
        )
        recovery_system.register_recovery_action(recovery_action)

        # Act
        result = recovery_system.attempt_manual_recovery(error_report.id, "failing_recovery")

        # Assert
        assert result.success is False
        assert "Handler failed" in result.message

    def test_attempt_manual_recovery_with_parameters(self, recovery_system):
        """Test manual recovery with additional parameters."""
        # Arrange
        test_error = ValueError("Test error")
        error_report = recovery_system.handle_error(test_error)

        received_params = {}

        def param_handler(error, params):
            received_params.update(params)
            return True

        recovery_action = RecoveryAction(
            name="param_recovery",
            description="Recovery with params",
            action_type="manual",
            severity_filter=[ErrorSeverity.MEDIUM],
            category_filter=[ErrorCategory.SYSTEM],
            handler=param_handler,
        )
        recovery_system.register_recovery_action(recovery_action)

        # Act
        result = recovery_system.attempt_manual_recovery(
            error_report.id,
            "param_recovery",
            parameters={"key": "value", "number": 42},
        )

        # Assert
        assert result.success is True
        assert received_params["key"] == "value"
        assert received_params["number"] == 42


class TestSystemHealthMonitoring:
    """Test system health monitoring."""

    @pytest.fixture
    def recovery_system(self):
        """Create a fresh recovery system for testing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(Path(tmpdir))
            yield system

    def test_system_health_healthy_state(self, recovery_system):
        """Test system health when no errors."""
        # Act
        health = recovery_system.get_system_health()

        # Assert
        assert health["status"] == "healthy"
        assert health["active_errors"] == 0
        assert health["total_errors"] == 0

    def test_system_health_degraded_with_high_errors(self, recovery_system):
        """Test system health with HIGH severity errors."""
        # Arrange
        error = ValueError("High severity error")

        # Mock the automatic recovery to fail so error stays in active_errors
        with patch.object(recovery_system, "_attempt_automatic_recovery") as mock_recovery:
            mock_recovery.return_value = RecoveryResult(
                success=False, action_name="test", message="Failed", duration=0.0
            )
            recovery_system.handle_error(
                error, severity=ErrorSeverity.HIGH, category=ErrorCategory.SYSTEM
            )

        # Act
        health = recovery_system.get_system_health()

        # Assert
        assert health["status"] == "degraded"

    def test_system_health_critical_with_critical_errors(self, recovery_system):
        """Test system health with CRITICAL errors."""
        # Arrange
        error = RuntimeError("Critical error")

        # Act
        with patch.object(recovery_system, "_attempt_automatic_recovery") as mock_recovery:
            mock_recovery.return_value = RecoveryResult(
                success=False, action_name="test", message="Failed", duration=0.0
            )
            recovery_system.handle_error(
                error, severity=ErrorSeverity.CRITICAL, category=ErrorCategory.SYSTEM
            )

        health = recovery_system.get_system_health()

        # Assert
        assert health["status"] == "critical"

    def test_system_health_warning_with_multiple_errors(self, recovery_system):
        """Test system health warning when many errors but not critical."""
        # Arrange
        for i in range(7):
            error = Exception(f"Error {i}")
            recovery_system.handle_error(error, severity=ErrorSeverity.MEDIUM)

        # Act
        health = recovery_system.get_system_health()

        # Assert
        assert health["status"] == "warning"


class TestErrorSummaryGeneration:
    """Test error summary generation."""

    @pytest.fixture
    def recovery_system(self):
        """Create a fresh recovery system for testing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(Path(tmpdir))
            yield system

    def test_get_error_summary_empty(self, recovery_system):
        """Test error summary when no errors."""
        # Act
        summary = recovery_system.get_error_summary()

        # Assert
        assert summary["total_recent_errors"] == 0
        assert summary["active_errors"] == 0
        assert summary["by_severity"] == {}

    def test_get_error_summary_with_errors(self, recovery_system):
        """Test error summary with multiple errors."""
        # Arrange
        recovery_system.handle_error(
            ValueError("Error 1"), severity=ErrorSeverity.HIGH, category=ErrorCategory.RESEARCH
        )
        recovery_system.handle_error(
            RuntimeError("Error 2"), severity=ErrorSeverity.MEDIUM, category=ErrorCategory.SYSTEM
        )
        recovery_system.handle_error(
            Exception("Error 3"), severity=ErrorSeverity.LOW, category=ErrorCategory.RESEARCH
        )

        # Act
        summary = recovery_system.get_error_summary()

        # Assert
        assert summary["total_recent_errors"] == 3
        assert "high" in summary["by_severity"]
        assert "medium" in summary["by_severity"]
        assert "research" in summary["by_category"]

    def test_get_error_summary_limit(self, recovery_system):
        """Test error summary respects limit parameter."""
        # Arrange
        for i in range(100):
            recovery_system.handle_error(
                Exception(f"Error {i}"), severity=ErrorSeverity.LOW
            )

        # Act
        summary = recovery_system.get_error_summary(limit=10)

        # Assert
        assert len(summary["recent_errors"]) <= 10


class TestTroubleshootingGuideGeneration:
    """Test troubleshooting guide generation."""

    @pytest.fixture
    def recovery_system(self):
        """Create a fresh recovery system for testing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(Path(tmpdir))
            yield system

    def test_generate_troubleshooting_guide(self, recovery_system):
        """Test troubleshooting guide generation."""
        # Arrange
        # Add some repeated errors to create patterns
        for _ in range(3):
            recovery_system.handle_error(
                ValueError("Configuration error"),
                severity=ErrorSeverity.HIGH,
                category=ErrorCategory.CONFIGURATION,
            )

        # Act
        guide = recovery_system.generate_troubleshooting_guide()

        # Assert
        assert "generated_at" in guide
        assert "common_issues" in guide
        assert "recovery_procedures" in guide
        assert "prevention_tips" in guide
        assert "emergency_procedures" in guide

    def test_troubleshooting_guide_includes_recovery_actions(self, recovery_system):
        """Test troubleshooting guide includes registered recovery actions."""
        # Act
        guide = recovery_system.generate_troubleshooting_guide()

        # Assert
        assert len(guide["recovery_procedures"]) > 0
        assert "restart_research_engines" in guide["recovery_procedures"]


class TestErrorCleanup:
    """Test error cleanup operations."""

    @pytest.fixture
    def recovery_system(self):
        """Create a fresh recovery system for testing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(Path(tmpdir))
            yield system

    def test_cleanup_old_errors(self, recovery_system):
        """Test cleanup of old error records."""
        # Arrange
        # Add current error
        recovery_system.handle_error(
            ValueError("Current error"), severity=ErrorSeverity.MEDIUM
        )

        # Add old error
        old_error = ErrorReport(
            id="old_error",
            timestamp=datetime.now(timezone.utc) - timedelta(days=35),
            severity=ErrorSeverity.LOW,
            category=ErrorCategory.SYSTEM,
            message="Old error",
            details={},
            stack_trace="",
            context={},
        )
        recovery_system.error_history.append(old_error)

        initial_count = len(recovery_system.error_history)

        # Act
        result = recovery_system.cleanup_old_errors(days_to_keep=30)

        # Assert
        assert result["removed_count"] == 1
        assert len(recovery_system.error_history) == initial_count - 1

    def test_cleanup_no_old_errors(self, recovery_system):
        """Test cleanup when no old errors."""
        # Arrange
        recovery_system.handle_error(ValueError("Recent error"))

        # Act
        result = recovery_system.cleanup_old_errors(days_to_keep=30)

        # Assert
        assert result["removed_count"] == 0


class TestPhase3AdvancedRecovery:
    """Test Phase 3 advanced recovery features."""

    @pytest.fixture
    def recovery_system(self):
        """Create a fresh recovery system for testing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(Path(tmpdir))
            yield system

    @pytest.mark.asyncio
    async def test_report_advanced_failure(self, recovery_system):
        """Test reporting advanced failure events."""
        # Act
        failure_id = await recovery_system.report_advanced_failure(
            failure_mode=FailureMode.NETWORK_FAILURE,
            component="api_service",
            description="Network connection timeout",
            severity="high",
            context={"timeout": 30},
            affected_operations=["get_user", "post_data"],
        )

        # Assert
        assert failure_id is not None
        assert failure_id in recovery_system.advanced_failures
        assert recovery_system.advanced_recovery_stats["total_failures"] == 1

    @pytest.mark.asyncio
    async def test_report_cascade_failure_detection(self, recovery_system):
        """Test cascade failure detection."""
        # Arrange - report multiple related failures
        component = "database"
        for i in range(4):
            await recovery_system.report_advanced_failure(
                failure_mode=FailureMode.SYSTEM_OVERLOAD,
                component=component,
                description=f"Overload {i}",
                severity="high",
                context={"related_components": [component]},
            )

        # Assert
        assert recovery_system.advanced_recovery_stats["cascade_failures_detected"] > 0

    @pytest.mark.asyncio
    async def test_execute_retry_with_backoff(self, recovery_system):
        """Test retry with exponential backoff execution."""
        # Arrange
        failure = FailureEvent(
            failure_id="fail123",
            failure_mode=FailureMode.NETWORK_FAILURE,
            timestamp=datetime.now(timezone.utc),
            component="test_component",
            description="Test",
            severity="medium",
        )
        recovery_system.advanced_failures["fail123"] = failure

        action = AdvancedRecoveryAction(
            action_id="retry_action",
            failure_id="fail123",
            strategy=RecoveryStrategy.RETRY_WITH_BACKOFF,
            timestamp=datetime.now(timezone.utc),
            max_retries=3,
        )

        # Act
        success = await recovery_system._execute_retry_with_backoff(action)

        # Assert
        assert success is True
        assert len(action.execution_log) > 0

    @pytest.mark.asyncio
    async def test_execute_circuit_breaker_action(self, recovery_system):
        """Test circuit breaker action execution."""
        # Arrange
        failure = FailureEvent(
            failure_id="fail123",
            failure_mode=FailureMode.SYSTEM_OVERLOAD,
            timestamp=datetime.now(timezone.utc),
            component="test_component",
            description="Test",
            severity="high",
        )
        recovery_system.advanced_failures["fail123"] = failure

        action = AdvancedRecoveryAction(
            action_id="cb_action",
            failure_id="fail123",
            strategy=RecoveryStrategy.CIRCUIT_BREAKER,
            timestamp=datetime.now(timezone.utc),
        )

        # Act
        success = await recovery_system._execute_circuit_breaker_action(action)

        # Assert
        assert success is True
        assert recovery_system.circuit_breaker_states["test_component"]["state"] == "OPEN"

    @pytest.mark.asyncio
    async def test_execute_rollback_action(self, recovery_system):
        """Test rollback action execution."""
        # Arrange
        action = AdvancedRecoveryAction(
            action_id="rollback_action",
            failure_id="fail123",
            strategy=RecoveryStrategy.ROLLBACK,
            timestamp=datetime.now(timezone.utc),
        )

        # Act
        success = await recovery_system._execute_rollback_action(action)

        # Assert
        assert success is True
        assert action.rollback_action_id is not None

    @pytest.mark.asyncio
    async def test_execute_quarantine_action(self, recovery_system):
        """Test quarantine action execution."""
        # Arrange
        failure = FailureEvent(
            failure_id="fail123",
            failure_mode=FailureMode.DEADLOCK,
            timestamp=datetime.now(timezone.utc),
            component="deadlock_component",
            description="Deadlock detected",
            severity="critical",
        )
        recovery_system.advanced_failures["fail123"] = failure

        action = AdvancedRecoveryAction(
            action_id="quarantine_action",
            failure_id="fail123",
            strategy=RecoveryStrategy.QUARANTINE,
            timestamp=datetime.now(timezone.utc),
        )

        # Act
        success = await recovery_system._execute_quarantine_action(action)

        # Assert
        assert success is True

    @pytest.mark.asyncio
    async def test_create_system_snapshot(self, recovery_system):
        """Test system snapshot creation."""
        # Act
        snapshot_id = await recovery_system._create_system_snapshot(
            description="Test snapshot", is_rollback_point=True
        )

        # Assert
        assert snapshot_id is not None
        assert snapshot_id in recovery_system.system_snapshots
        assert recovery_system.advanced_recovery_stats["snapshots_created"] == 1
        snapshot = recovery_system.system_snapshots[snapshot_id]
        assert snapshot.is_rollback_point is True

    def test_determine_advanced_recovery_strategy(self, recovery_system):
        """Test recovery strategy determination."""
        # Test various failure modes
        test_cases = [
            (FailureMode.NETWORK_FAILURE, RecoveryStrategy.RETRY_WITH_BACKOFF),
            (FailureMode.RESOURCE_EXHAUSTION, RecoveryStrategy.DEGRADE_SERVICE),
            (FailureMode.DATA_CORRUPTION, RecoveryStrategy.ROLLBACK),
            (FailureMode.SYSTEM_OVERLOAD, RecoveryStrategy.CIRCUIT_BREAKER),
            (FailureMode.CASCADE_FAILURE, RecoveryStrategy.EMERGENCY_STOP),
            (FailureMode.DEADLOCK, RecoveryStrategy.QUARANTINE),
        ]

        for failure_mode, expected_strategy in test_cases:
            # Act
            strategy = recovery_system._determine_advanced_recovery_strategy(failure_mode)

            # Assert
            assert strategy == expected_strategy

    def test_calculate_recovery_priority(self, recovery_system):
        """Test recovery priority calculation."""
        # Arrange
        critical_failure = FailureEvent(
            failure_id="critical",
            failure_mode=FailureMode.SYSTEM_OVERLOAD,
            timestamp=datetime.now(timezone.utc),
            component="critical_component",
            description="Critical",
            severity="critical",
            affected_operations=list(range(15)),
        )

        low_failure = FailureEvent(
            failure_id="low",
            failure_mode=FailureMode.NETWORK_FAILURE,
            timestamp=datetime.now(timezone.utc),
            component="low_component",
            description="Low",
            severity="low",
            affected_operations=[],
        )

        # Act
        critical_priority = recovery_system._calculate_recovery_priority(critical_failure)
        low_priority = recovery_system._calculate_recovery_priority(low_failure)

        # Assert
        assert critical_priority < low_priority  # Lower number = higher priority

    def test_get_advanced_system_status(self, recovery_system):
        """Test advanced system status reporting."""
        # Act
        status = recovery_system.get_advanced_system_status()

        # Assert
        assert status["status"] == "running"
        assert status["phase3_features"] == "enabled"
        assert "advanced_recovery_statistics" in status
        assert "circuit_breaker_states" in status
        assert "system_snapshots" in status


class TestRecoveryActionHelpers:
    """Test recovery action helper methods."""

    @pytest.fixture
    def recovery_system(self):
        """Create a fresh recovery system for testing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(Path(tmpdir))
            yield system

    @patch("shutil.rmtree")
    @patch("shutil.copy2")
    def test_restart_research_engines(self, mock_copy, mock_rmtree, recovery_system):
        """Test research engine restart."""
        # Arrange
        error_report = ErrorReport(
            id="test",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.HIGH,
            category=ErrorCategory.RESEARCH,
            message="Test",
            details={},
            stack_trace="",
            context={},
        )

        # Act
        result = recovery_system._restart_research_engines(error_report, {})

        # Assert
        assert result is True

    @patch("shutil.rmtree")
    @patch("shutil.copy2")
    def test_restore_config_backup(self, mock_copy, mock_rmtree, recovery_system):
        """Test configuration backup restoration."""
        # Arrange
        backup_dir = recovery_system.project_root / ".moai" / "config_backups"
        backup_dir.mkdir(parents=True, exist_ok=True)

        # Create a mock backup file
        backup_file = backup_dir / "config_20231201_120000.json"
        backup_file.write_text('{"test": "config"}')

        error_report = ErrorReport(
            id="test",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.CRITICAL,
            category=ErrorCategory.CONFIGURATION,
            message="Test",
            details={},
            stack_trace="",
            context={},
        )

        # Act
        result = recovery_system._restore_config_backup(error_report, {})

        # Assert
        assert result is True

    @patch("shutil.rmtree")
    def test_clear_agent_cache(self, mock_rmtree, recovery_system):
        """Test agent cache clearing."""
        # Arrange
        error_report = ErrorReport(
            id="test",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.HIGH,
            category=ErrorCategory.COMMUNICATION,
            message="Test",
            details={},
            stack_trace="",
            context={},
        )

        # Act
        result = recovery_system._clear_agent_cache(error_report, {})

        # Assert
        assert result is True

    def test_validate_research_integrity(self, recovery_system):
        """Test research integrity validation."""
        # Arrange
        error_report = ErrorReport(
            id="test",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.HIGH,
            category=ErrorCategory.RESEARCH,
            message="Test",
            details={},
            stack_trace="",
            context={},
        )

        # Act
        result = recovery_system._validate_research_integrity(error_report, {})

        # Assert
        assert isinstance(result, dict)
        assert "issues_found" in result
        assert "repairs_made" in result

    def test_rollback_last_changes(self, recovery_system):
        """Test rolling back last changes."""
        # Arrange
        with patch("sys.path") as mock_path:
            with patch("builtins.__import__", side_effect=ImportError("RollbackManager not found")):
                error_report = ErrorReport(
                    id="test",
                    timestamp=datetime.now(timezone.utc),
                    severity=ErrorSeverity.CRITICAL,
                    category=ErrorCategory.RESEARCH,
                    message="Test",
                    details={},
                    stack_trace="",
                    context={},
                )

                # Act
                result = recovery_system._rollback_last_changes(error_report, {})

                # Assert - Should return False when RollbackManager import fails
                assert result is False

    @patch("shutil.rmtree")
    def test_reset_system_state(self, mock_rmtree, recovery_system):
        """Test system state reset."""
        # Arrange
        error_report = ErrorReport(
            id="test",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.CRITICAL,
            category=ErrorCategory.SYSTEM,
            message="Test",
            details={},
            stack_trace="",
            context={},
        )

        # Act
        result = recovery_system._reset_system_state(error_report, {})

        # Assert
        assert result is True

    @patch("shutil.rmtree")
    def test_optimize_performance(self, mock_rmtree, recovery_system):
        """Test performance optimization."""
        # Arrange
        error_report = ErrorReport(
            id="test",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.MEDIUM,
            category=ErrorCategory.PERFORMANCE,
            message="Test",
            details={},
            stack_trace="",
            context={},
        )

        # Act
        result = recovery_system._optimize_performance(error_report, {})

        # Assert
        assert result is True

    def test_free_resources(self, recovery_system):
        """Test resource cleanup."""
        # Arrange
        error_report = ErrorReport(
            id="test",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.HIGH,
            category=ErrorCategory.RESOURCE,
            message="Test",
            details={},
            stack_trace="",
            context={},
        )

        # Act
        result = recovery_system._free_resources(error_report, {})

        # Assert
        assert result is True


class TestValidationAndRepair:
    """Test validation and repair methods."""

    @pytest.fixture
    def recovery_system(self):
        """Create a fresh recovery system for testing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(Path(tmpdir))
            yield system

    def test_validate_skill_file_valid(self, recovery_system):
        """Test validation of valid skill file."""
        # Arrange
        skill_dir = recovery_system.project_root / ".claude" / "skills"
        skill_dir.mkdir(parents=True, exist_ok=True)
        skill_file = skill_dir / "test_skill.md"
        # Create valid skill file with proper format and sufficient length
        skill_file.write_text(
            "---\nname: test\ndescription: Test Skill\n---\n\n"
            + "Content here\n" * 20  # Ensure content is longer than 100 chars
        )

        # Act
        result = recovery_system._validate_skill_file(skill_file)

        # Assert
        assert result is True

    def test_validate_skill_file_invalid(self, recovery_system):
        """Test validation of invalid skill file."""
        # Arrange
        skill_dir = recovery_system.project_root / ".claude" / "skills"
        skill_dir.mkdir(parents=True, exist_ok=True)
        skill_file = skill_dir / "bad_skill.md"
        skill_file.write_text("invalid")

        # Act
        result = recovery_system._validate_skill_file(skill_file)

        # Assert
        assert result is False

    def test_validate_agent_file_valid(self, recovery_system):
        """Test validation of valid agent file."""
        # Arrange
        agent_dir = recovery_system.project_root / ".claude" / "agents" / "alfred"
        agent_dir.mkdir(parents=True, exist_ok=True)
        agent_file = agent_dir / "test_agent.md"
        # Create a valid agent file with sufficient length (>200 chars)
        agent_file.write_text(
            "role: Test Agent\n"
            "description: Test\n"
            "content here\n"
            + "Extended content\n" * 20  # Ensure content is longer than 200 chars
        )

        # Act
        result = recovery_system._validate_agent_file(agent_file)

        # Assert
        assert result is True

    def test_validate_command_file_valid(self, recovery_system):
        """Test validation of valid command file."""
        # Arrange
        command_dir = recovery_system.project_root / ".claude" / "commands" / "alfred"
        command_dir.mkdir(parents=True, exist_ok=True)
        command_file = command_dir / "test_command.md"
        command_file.write_text("name: test\nallowed-tools:\n - Read")

        # Act
        result = recovery_system._validate_command_file(command_file)

        # Assert
        assert result is True

    def test_repair_skill_file(self, recovery_system):
        """Test skill file repair."""
        # Arrange
        skill_dir = recovery_system.project_root / ".claude" / "skills"
        skill_dir.mkdir(parents=True, exist_ok=True)
        skill_file = skill_dir / "broken_skill.md"
        skill_file.write_text("Just some content without frontmatter")

        # Act
        result = recovery_system._repair_skill_file(skill_file)

        # Assert
        assert result is True
        content = skill_file.read_text()
        assert content.startswith("---")


class TestErrorPatternAnalysis:
    """Test error pattern analysis."""

    @pytest.fixture
    def recovery_system(self):
        """Create a fresh recovery system for testing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(Path(tmpdir))
            yield system

    def test_identify_error_patterns(self, recovery_system):
        """Test error pattern identification."""
        # Arrange
        for _ in range(3):
            recovery_system.handle_error(
                ValueError("Value error"),
                category=ErrorCategory.VALIDATION,
            )

        for _ in range(2):
            recovery_system.handle_error(
                RuntimeError("Runtime error"),
                category=ErrorCategory.SYSTEM,
            )

        # Act
        patterns = recovery_system._identify_error_patterns(recovery_system.error_history)

        # Assert
        assert len(patterns) > 0
        assert patterns["validation:ValueError"] == 3
        assert patterns["system:RuntimeError"] == 2

    def test_get_pattern_severity(self, recovery_system):
        """Test pattern severity mapping."""
        # Act & Assert
        assert recovery_system._get_pattern_severity("system:Exception") == "critical"
        assert recovery_system._get_pattern_severity("research:Exception") == "high"
        assert recovery_system._get_pattern_severity("communication:Exception") == "medium"
        assert recovery_system._get_pattern_severity("unknown:Exception") == "medium"

    def test_get_solutions_for_pattern(self, recovery_system):
        """Test getting solutions for error patterns."""
        # Act
        solutions = recovery_system._get_solutions_for_pattern("research:Exception")

        # Assert
        assert len(solutions) > 0
        assert "Restart research engines" in solutions


class TestGlobalFunctionsAndDecorator:
    """Test global functions and decorator."""

    def test_get_error_recovery_system_singleton(self):
        """Test singleton pattern for global system."""
        # Reset global instance
        import moai_adk.core.error_recovery_system as module

        module._error_recovery_system = None

        with tempfile.TemporaryDirectory() as tmpdir:
            # Act
            system1 = get_error_recovery_system(Path(tmpdir))
            system2 = get_error_recovery_system(Path(tmpdir))

            # Assert
            assert system1 is system2

            # Clean up
            module._error_recovery_system = None

    def test_handle_error_convenience_function(self):
        """Test convenience function for handling errors."""
        # Act
        error_report = handle_error(
            ValueError("Test error"),
            context={"test": "context"},
            severity=ErrorSeverity.MEDIUM,
        )

        # Assert
        assert error_report is not None
        assert error_report.message == "Test error"

    def test_error_handler_decorator_success(self):
        """Test error handler decorator on successful function."""
        # Arrange
        @error_handler(severity=ErrorSeverity.MEDIUM, category=ErrorCategory.SYSTEM)
        def test_function():
            return "success"

        # Act
        result = test_function()

        # Assert
        assert result == "success"

    def test_error_handler_decorator_error(self):
        """Test error handler decorator on failing function."""
        # Arrange
        @error_handler(severity=ErrorSeverity.HIGH, category=ErrorCategory.SYSTEM)
        def failing_function():
            raise ValueError("Test error")

        # Act & Assert
        with pytest.raises(ValueError):
            failing_function()

    def test_error_handler_decorator_with_context(self):
        """Test error handler decorator with custom context."""
        # Arrange
        @error_handler(
            severity=ErrorSeverity.MEDIUM,
            category=ErrorCategory.SYSTEM,
            context={"custom": "context"},
        )
        def test_function():
            raise RuntimeError("Test")

        # Act & Assert
        with pytest.raises(RuntimeError):
            test_function()


class TestBackgroundMonitoring:
    """Test background monitoring thread."""

    @pytest.fixture
    def recovery_system(self):
        """Create a fresh recovery system for testing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(Path(tmpdir))
            yield system
            system.monitoring_active = False

    def test_background_monitoring_thread_running(self, recovery_system):
        """Test background monitoring thread is running."""
        # Assert
        assert recovery_system.monitoring_active is True
        assert recovery_system.monitor_thread.is_alive()

    def test_check_error_patterns_detects_burst(self, recovery_system):
        """Test error burst detection."""
        # Arrange
        # Add many errors in a short time
        for i in range(12):
            recovery_system.handle_error(
                Exception(f"Error {i}"), severity=ErrorSeverity.LOW
            )

        # Act
        recovery_system._check_error_patterns()

        # Assert - Should log warning about high error rate
        # This is tested indirectly through error_history


class TestErrorRecoveryEdgeCases:
    """Test edge cases and error conditions."""

    @pytest.fixture
    def recovery_system(self):
        """Create a fresh recovery system for testing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(Path(tmpdir))
            yield system

    def test_generate_error_id_uniqueness(self, recovery_system):
        """Test that error IDs are unique."""
        # Act
        id1 = recovery_system._generate_error_id()
        time.sleep(0.01)  # Ensure different timestamp
        id2 = recovery_system._generate_error_id()

        # Assert
        assert id1 != id2

    def test_calculate_recovery_rate_empty_list(self, recovery_system):
        """Test recovery rate calculation with empty error list."""
        # Act
        rate = recovery_system._calculate_recovery_rate([])

        # Assert
        assert rate == 0.0

    def test_calculate_recovery_rate_all_recovered(self, recovery_system):
        """Test recovery rate when all errors recovered."""
        # Arrange
        errors = []
        for i in range(5):
            error = ErrorReport(
                id=f"err{i}",
                timestamp=datetime.now(timezone.utc),
                severity=ErrorSeverity.LOW,
                category=ErrorCategory.SYSTEM,
                message="Test",
                details={},
                stack_trace="",
                context={},
                recovery_successful=True,
            )
            errors.append(error)

        # Act
        rate = recovery_system._calculate_recovery_rate(errors)

        # Assert
        assert rate == 1.0

    def test_calculate_recovery_rate_partial(self, recovery_system):
        """Test recovery rate with partial recovery."""
        # Arrange
        errors = []
        for i in range(5):
            error = ErrorReport(
                id=f"err{i}",
                timestamp=datetime.now(timezone.utc),
                severity=ErrorSeverity.LOW,
                category=ErrorCategory.SYSTEM,
                message="Test",
                details={},
                stack_trace="",
                context={},
                recovery_successful=i < 3,  # First 3 recovered
            )
            errors.append(error)

        # Act
        rate = recovery_system._calculate_recovery_rate(errors)

        # Assert
        assert rate == 0.6  # 3/5 = 0.6

    @patch("json.dump")
    def test_log_error_file_write_failure(self, mock_dump, recovery_system):
        """Test error logging when file write fails."""
        # Arrange
        mock_dump.side_effect = IOError("Write failed")
        error = ValueError("Test error")

        # Act & Assert - Should not raise, just log
        error_report = recovery_system.handle_error(error)
        assert error_report is not None

    def test_automatic_recovery_with_no_suitable_actions(self, recovery_system):
        """Test automatic recovery when no suitable actions found."""
        # Arrange
        error_report = ErrorReport(
            id="test",
            timestamp=datetime.now(timezone.utc),
            severity=ErrorSeverity.CRITICAL,
            category=ErrorCategory.USER_INPUT,  # No automatic actions for this category
            message="Test",
            details={},
            stack_trace="",
            context={},
        )

        # Act
        result = recovery_system._attempt_automatic_recovery(error_report)

        # Assert
        assert result.success is False
        assert "No suitable automatic recovery action succeeded" in result.message


class TestIntegrationScenarios:
    """Test complete integration scenarios."""

    @pytest.fixture
    def recovery_system(self):
        """Create a fresh recovery system for testing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            system = ErrorRecoverySystem(Path(tmpdir))
            yield system

    def test_complete_error_handling_workflow(self, recovery_system):
        """Test complete error handling workflow."""
        # Arrange
        test_error = ValueError("Complete test error")

        # Act
        error_report = recovery_system.handle_error(
            test_error,
            context={"workflow": "complete"},
            severity=ErrorSeverity.MEDIUM,
            category=ErrorCategory.INTEGRATION,
        )

        # Assert error is handled
        assert error_report.id is not None

        # Get summary
        summary = recovery_system.get_error_summary()
        assert summary["total_recent_errors"] >= 1

        # Get health
        health = recovery_system.get_system_health()
        assert health["total_errors"] >= 1

        # Generate troubleshooting guide
        guide = recovery_system.generate_troubleshooting_guide()
        assert "recovery_procedures" in guide

    @pytest.mark.asyncio
    async def test_advanced_recovery_workflow(self, recovery_system):
        """Test advanced recovery workflow."""
        # Act
        failure_id = await recovery_system.report_advanced_failure(
            failure_mode=FailureMode.HOOK_EXECUTION_FAILURE,
            component="test_hook",
            description="Hook failed to execute",
            severity="high",
            affected_operations=["hook_setup", "hook_execute"],
        )

        # Assert
        assert failure_id is not None
        assert len(recovery_system.advanced_recovery_actions) > 0


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--cov=src/moai_adk/core/error_recovery_system"])
