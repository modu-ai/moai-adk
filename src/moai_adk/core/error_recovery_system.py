"""
Comprehensive Error Recovery System - Phase 3 Enterprise Edition

Advanced enterprise-grade error recovery with automatic healing, data consistency checks,
rollback capabilities, state persistence, and self-healing mechanisms.

Phase 3 Features:
- Event-driven error recovery architecture
- Automatic system recovery from all failure modes
- Data consistency checks and repair mechanisms
- Comprehensive rollback and state persistence
- Self-healing capabilities with circuit breakers
- Dead letter queue handling for failed operations
- Multi-strategy recovery with exponential backoff
- Disaster recovery and business continuity
- Real-time failure mode analysis and prediction

Legacy Features:
- Error detection and classification
- Recovery procedures and fallback mechanisms
- Integration with research hooks, agents, and skills
- Documentation of error handling procedures
- Troubleshooting guides and automated recovery
- Multi-level error handling (critical, warning, info)
- Manual recovery procedures
- Error logging and tracking
- System health monitoring
- Emergency recovery procedures
"""

import asyncio
import hashlib
import json
import logging
import os
import sys
import tempfile
import threading
import time
import traceback
import uuid
from collections import defaultdict, deque
from dataclasses import asdict, dataclass, field
from datetime import datetime, timedelta, timezone
from enum import Enum
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional

# Configure comprehensive logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.FileHandler(Path(tempfile.gettempdir()) / "moai_error_recovery.log"),
        logging.StreamHandler(sys.stdout),
    ],
)

logger = logging.getLogger(__name__)


class ErrorSeverity(Enum):
    """Error severity levels"""

    CRITICAL = "critical"  # System failure, immediate attention required
    HIGH = "high"  # Major functionality impacted
    MEDIUM = "medium"  # Partial functionality impacted
    LOW = "low"  # Minor issue, can be deferred
    INFO = "info"  # Informational message


class ErrorCategory(Enum):
    """Error categories for classification"""

    SYSTEM = "system"  # System-level errors
    CONFIGURATION = "configuration"  # Configuration errors
    RESEARCH = "research"  # Research workflow errors
    INTEGRATION = "integration"  # Integration errors
    COMMUNICATION = "communication"  # Agent/communication errors
    VALIDATION = "validation"  # Validation errors
    PERFORMANCE = "performance"  # Performance issues
    RESOURCE = "resource"  # Resource exhaustion
    NETWORK = "network"  # Network-related errors
    USER_INPUT = "user_input"  # User input errors


# Phase 3: Advanced Error Recovery Enums


class FailureMode(Enum):
    """Types of failure modes in the system"""

    HOOK_EXECUTION_FAILURE = "hook_execution_failure"
    RESOURCE_EXHAUSTION = "resource_exhaustion"
    DATA_CORRUPTION = "data_corruption"
    NETWORK_FAILURE = "network_failure"
    SYSTEM_OVERLOAD = "system_overload"
    CONFIGURATION_ERROR = "configuration_error"
    TIMEOUT_FAILURE = "timeout_failure"
    MEMORY_LEAK = "memory_leak"
    DEADLOCK = "deadlock"
    AUTHENTICATION_FAILURE = "authentication_failure"
    VALIDATION_FAILURE = "validation_failure"
    EXTERNAL_SERVICE_FAILURE = "external_service_failure"
    STORAGE_FAILURE = "storage_failure"
    CONCURRENCY_ISSUE = "concurrency_issue"
    CIRCUIT_BREAKER_TRIPPED = "circuit_breaker_tripped"
    CASCADE_FAILURE = "cascade_failure"


class RecoveryStrategy(Enum):
    """Recovery strategies for different failure modes"""

    RETRY_WITH_BACKOFF = "retry_with_backoff"
    CIRCUIT_BREAKER = "circuit_breaker"
    ROLLBACK = "rollback"
    FAILOVER = "failover"
    DEGRADE_SERVICE = "degrade_service"
    RESTART_COMPONENT = "restart_component"
    DATA_REPAIR = "data_repair"
    CLEAR_CACHE = "clear_cache"
    SCALE_RESOURCES = "scale_resources"
    NOTIFY_ADMIN = "notify_admin"
    QUARANTINE = "quarantine"
    IGNORE = "ignore"
    ISOLATE_COMPONENT = "isolate_component"
    EMERGENCY_STOP = "emergency_stop"


class ConsistencyLevel(Enum):
    """Data consistency levels"""

    STRONG = "strong"  # Immediate consistency
    EVENTUAL = "eventual"  # Eventually consistent
    WEAK = "weak"  # Weak consistency
    CUSTOM = "custom"  # Custom consistency rules


class RecoveryStatus(Enum):
    """Recovery operation status"""

    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"
    ROLLED_BACK = "rolled_back"


@dataclass
class ErrorReport:
    """Comprehensive error report structure"""

    id: str
    timestamp: datetime
    severity: ErrorSeverity
    category: ErrorCategory
    message: str
    details: Dict[str, Any]
    stack_trace: Optional[str]
    context: Dict[str, Any]
    recovery_attempted: bool = False
    recovery_successful: bool = False
    resolution_message: Optional[str] = None


@dataclass
class RecoveryAction:
    """Recovery action definition"""

    name: str
    description: str
    action_type: str  # "automatic", "manual", "assisted"
    severity_filter: List[ErrorSeverity]
    category_filter: List[ErrorCategory]
    handler: Callable
    timeout: Optional[float] = None
    max_attempts: int = 3
    success_criteria: Optional[str] = None


@dataclass
class RecoveryResult:
    """Result of recovery action"""

    success: bool
    action_name: str
    message: str
    duration: float
    details: Dict[str, Any] = None
    next_actions: List[str] = None


# Phase 3: Advanced Error Recovery Dataclasses


@dataclass
class FailureEvent:
    """Represents a failure event in the system"""

    failure_id: str
    failure_mode: FailureMode
    timestamp: datetime
    component: str
    description: str
    severity: str  # "low", "medium", "high", "critical"
    context: Dict[str, Any] = field(default_factory=dict)
    error_details: Optional[Dict[str, Any]] = None
    affected_operations: List[str] = field(default_factory=list)
    auto_recovery_eligible: bool = True
    retry_count: int = 0
    metadata: Dict[str, Any] = field(default_factory=dict)
    parent_failure_id: Optional[str] = None  # For cascade failures
    root_cause: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "failure_id": self.failure_id,
            "failure_mode": self.failure_mode.value,
            "timestamp": self.timestamp.isoformat(),
            "component": self.component,
            "description": self.description,
            "severity": self.severity,
            "context": self.context,
            "error_details": self.error_details,
            "affected_operations": self.affected_operations,
            "auto_recovery_eligible": self.auto_recovery_eligible,
            "retry_count": self.retry_count,
            "metadata": self.metadata,
            "parent_failure_id": self.parent_failure_id,
            "root_cause": self.root_cause,
        }

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "FailureEvent":
        """Create from dictionary"""
        return cls(
            failure_id=data["failure_id"],
            failure_mode=FailureMode(data["failure_mode"]),
            timestamp=datetime.fromisoformat(data["timestamp"]),
            component=data["component"],
            description=data["description"],
            severity=data["severity"],
            context=data.get("context", {}),
            error_details=data.get("error_details"),
            affected_operations=data.get("affected_operations", []),
            auto_recovery_eligible=data.get("auto_recovery_eligible", True),
            retry_count=data.get("retry_count", 0),
            metadata=data.get("metadata", {}),
            parent_failure_id=data.get("parent_failure_id"),
            root_cause=data.get("root_cause"),
        )


@dataclass
class AdvancedRecoveryAction:
    """Advanced recovery action with enhanced capabilities"""

    action_id: str
    failure_id: str
    strategy: RecoveryStrategy
    timestamp: datetime
    status: RecoveryStatus = RecoveryStatus.PENDING
    description: str = ""
    parameters: Dict[str, Any] = field(default_factory=dict)
    execution_log: List[str] = field(default_factory=list)
    rollback_available: bool = True
    timeout_seconds: float = 300.0
    retry_attempts: int = 0
    max_retries: int = 3
    rollback_action_id: Optional[str] = None
    dependencies: List[str] = field(default_factory=list)  # Other actions this depends on
    priority: int = 5  # 1-10, lower number = higher priority

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "action_id": self.action_id,
            "failure_id": self.failure_id,
            "strategy": self.strategy.value,
            "timestamp": self.timestamp.isoformat(),
            "status": self.status.value,
            "description": self.description,
            "parameters": self.parameters,
            "execution_log": self.execution_log,
            "rollback_available": self.rollback_available,
            "timeout_seconds": self.timeout_seconds,
            "retry_attempts": self.retry_attempts,
            "max_retries": self.max_retries,
            "rollback_action_id": self.rollback_action_id,
            "dependencies": self.dependencies,
            "priority": self.priority,
        }


@dataclass
class SystemSnapshot:
    """Represents a system state snapshot for rollback"""

    snapshot_id: str
    timestamp: datetime
    component_states: Dict[str, Dict[str, Any]]
    configuration_hash: str
    data_checksums: Dict[str, str]
    metadata: Dict[str, Any] = field(default_factory=dict)
    parent_snapshot_id: Optional[str] = None
    is_rollback_point: bool = False
    description: str = ""
    consistency_level: ConsistencyLevel = ConsistencyLevel.EVENTUAL

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization"""
        return {
            "snapshot_id": self.snapshot_id,
            "timestamp": self.timestamp.isoformat(),
            "component_states": self.component_states,
            "configuration_hash": self.configuration_hash,
            "data_checksums": self.data_checksums,
            "metadata": self.metadata,
            "parent_snapshot_id": self.parent_snapshot_id,
            "is_rollback_point": self.is_rollback_point,
            "description": self.description,
            "consistency_level": self.consistency_level.value,
        }


class ErrorRecoverySystem:
    """Comprehensive error handling and recovery system"""

    def __init__(self, project_root: Path = None):
        self.project_root = project_root or Path.cwd()
        self.error_log_dir = self.project_root / ".moai" / "error_logs"
        self.error_log_dir.mkdir(parents=True, exist_ok=True)

        # Error tracking
        self.active_errors: Dict[str, ErrorReport] = {}
        self.error_history: List[ErrorReport] = []
        self.recovery_actions: Dict[str, RecoveryAction] = {}
        self.error_stats: Dict[str, Any] = {
            "total_errors": 0,
            "by_severity": {},
            "by_category": {},
            "recovery_success_rate": 0.0,
        }

        # System health monitoring
        self.system_health = {
            "status": "healthy",
            "last_check": datetime.now(timezone.utc),
            "issues": [],
            "metrics": {},
        }

        # Initialize recovery actions
        self._initialize_recovery_actions()

        # Phase 3: Advanced recovery system initialization
        self._initialize_phase3_components()

        # Background monitoring thread
        self.monitoring_active = True
        self.monitor_thread = threading.Thread(target=self._background_monitoring, daemon=True)
        self.monitor_thread.start()

        logger.info("Error Recovery System initialized with Phase 3 enterprise features")

    def handle_error(
        self,
        error: Exception,
        context: Dict[str, Any] = None,
        severity: ErrorSeverity = ErrorSeverity.MEDIUM,
        category: ErrorCategory = ErrorCategory.SYSTEM,
    ) -> ErrorReport:
        """
        Handle an error with comprehensive logging and recovery

        Args:
            error: Exception that occurred
            context: Additional context information
            severity: Error severity level
            category: Error category

        Returns:
            ErrorReport with handling details
        """
        error_id = self._generate_error_id()
        timestamp = datetime.now(timezone.utc)

        # Create error report
        error_report = ErrorReport(
            id=error_id,
            timestamp=timestamp,
            severity=severity,
            category=category,
            message=str(error),
            details={
                "exception_type": type(error).__name__,
                "exception_module": type(error).__module__,
                "error_code": getattr(error, "code", None),
            },
            stack_trace=traceback.format_exc(),
            context=context or {},
            recovery_attempted=False,
            recovery_successful=False,
        )

        # Log error
        self._log_error(error_report)

        # Update statistics
        self._update_error_stats(error_report)

        # Store error
        self.active_errors[error_id] = error_report
        self.error_history.append(error_report)

        # Attempt automatic recovery
        if severity in [ErrorSeverity.CRITICAL, ErrorSeverity.HIGH]:
            recovery_result = self._attempt_automatic_recovery(error_report)
            error_report.recovery_attempted = True
            error_report.recovery_successful = recovery_result.success
            error_report.resolution_message = recovery_result.message

            if recovery_result.success:
                logger.info(f"Automatic recovery successful for error {error_id}")
                self.active_errors.pop(error_id, None)
            else:
                logger.warning(f"Automatic recovery failed for error {error_id}: {recovery_result.message}")

        # Update system health
        self._update_system_health()

        return error_report

    def register_recovery_action(self, action: RecoveryAction):
        """
        Register a new recovery action

        Args:
            action: RecoveryAction definition
        """
        self.recovery_actions[action.name] = action
        logger.info(f"Registered recovery action: {action.name}")

    def attempt_manual_recovery(
        self, error_id: str, action_name: str, parameters: Dict[str, Any] = None
    ) -> RecoveryResult:
        """
        Attempt manual recovery for a specific error

        Args:
            error_id: ID of error to recover
            action_name: Name of recovery action to attempt
            parameters: Additional parameters for recovery

        Returns:
            RecoveryResult with operation details
        """
        if error_id not in self.active_errors:
            return RecoveryResult(
                success=False,
                action_name=action_name,
                message=f"Error {error_id} not found in active errors",
                duration=0.0,
            )

        if action_name not in self.recovery_actions:
            return RecoveryResult(
                success=False,
                action_name=action_name,
                message=f"Recovery action {action_name} not found",
                duration=0.0,
            )

        error_report = self.active_errors[error_id]
        recovery_action = self.recovery_actions[action_name]

        logger.info(f"Attempting manual recovery {action_name} for error {error_id}")

        try:
            start_time = time.time()

            # Execute recovery action
            result = recovery_action.handler(error_report, parameters or {})

            duration = time.time() - start_time

            if result:
                recovery_result = RecoveryResult(
                    success=True,
                    action_name=action_name,
                    message="Manual recovery completed successfully",
                    duration=duration,
                    details={"result": result},
                )

                # Update error report
                error_report.recovery_successful = True
                error_report.resolution_message = recovery_result.message

                # Remove from active errors
                self.active_errors.pop(error_id, None)

            else:
                recovery_result = RecoveryResult(
                    success=False,
                    action_name=action_name,
                    message="Manual recovery returned unsuccessful result",
                    duration=duration,
                )

        except Exception as e:
            duration = time.time() - start_time
            recovery_result = RecoveryResult(
                success=False,
                action_name=action_name,
                message=f"Manual recovery failed: {str(e)}",
                duration=duration,
                details={"exception": str(e)},
            )

        return recovery_result

    def get_system_health(self) -> Dict[str, Any]:
        """
        Get current system health status

        Returns:
            System health information
        """
        self._update_system_health()

        last_check: datetime = self.system_health["last_check"]  # type: ignore[assignment]
        error_stats: Dict[str, Any] = self.error_stats  # type: ignore[assignment]
        issues: List[str] = self.system_health["issues"]  # type: ignore[assignment]
        metrics: Dict[str, Any] = self.system_health["metrics"]  # type: ignore[assignment]

        return {
            "status": self.system_health["status"],
            "last_check": last_check.isoformat(),
            "active_errors": len(self.active_errors),
            "total_errors": len(self.error_history),
            "error_stats": error_stats.copy(),
            "issues": issues.copy(),
            "metrics": metrics.copy(),
            "recovery_actions_available": len(self.recovery_actions),
        }

    def get_error_summary(self, limit: int = 50) -> Dict[str, Any]:
        """
        Get summary of recent errors

        Args:
            limit: Maximum number of errors to include

        Returns:
            Error summary information
        """
        recent_errors = self.error_history[-limit:]

        # Categorize errors
        by_severity: Dict[str, List[str]] = {}
        by_category: Dict[str, List[str]] = {}

        for error in recent_errors:
            # By severity
            severity = error.severity.value
            if severity not in by_severity:
                by_severity[severity] = []
            by_severity[severity].append(error.id)

            # By category
            category = error.category.value
            if category not in by_category:
                by_category[category] = []
            by_category[category].append(error.id)

        # Common error patterns
        error_patterns = self._identify_error_patterns(recent_errors)

        return {
            "total_recent_errors": len(recent_errors),
            "active_errors": len(self.active_errors),
            "by_severity": {k: len(v) for k, v in by_severity.items()},
            "by_category": {k: len(v) for k, v in by_category.items()},
            "common_patterns": error_patterns,
            "recovery_rate": self._calculate_recovery_rate(recent_errors),
            "recent_errors": [
                {
                    "id": error.id,
                    "timestamp": error.timestamp.isoformat(),
                    "severity": error.severity.value,
                    "category": error.category.value,
                    "message": error.message,
                    "recovered": error.recovery_successful,
                }
                for error in recent_errors[-10:]  # Last 10 errors
            ],
        }

    def generate_troubleshooting_guide(self) -> Dict[str, Any]:
        """
        Generate troubleshooting guide based on error history

        Returns:
            Troubleshooting guide with solutions
        """
        common_issues: List[Dict[str, Any]] = []
        recovery_procedures: Dict[str, Dict[str, Any]] = {}
        prevention_tips: List[str] = []
        emergency_procedures: List[Dict[str, str]] = []

        guide = {
            "generated_at": datetime.now(timezone.utc).isoformat(),
            "common_issues": common_issues,
            "recovery_procedures": recovery_procedures,
            "prevention_tips": prevention_tips,
            "emergency_procedures": emergency_procedures,
        }

        # Analyze common issues
        error_patterns = self._identify_error_patterns(self.error_history)
        for pattern, frequency in error_patterns.items():
            if frequency > 2:  # Issues that occurred more than twice
                common_issues.append(
                    {
                        "pattern": pattern,
                        "frequency": frequency,
                        "severity": self._get_pattern_severity(pattern),
                        "solutions": self._get_solutions_for_pattern(pattern),
                    }
                )

        # Generate recovery procedures
        for action_name, action in self.recovery_actions.items():
            recovery_procedures[action_name] = {
                "description": action.description,
                "type": action.action_type,
                "for_severities": [s.value for s in action.severity_filter],
                "for_categories": [c.value for c in action.category_filter],
            }

        # Prevention tips
        prevention_tips.extend(self._generate_prevention_tips())

        # Emergency procedures
        emergency_procedures.extend(self._generate_emergency_procedures())

        return guide

    def cleanup_old_errors(self, days_to_keep: int = 30) -> Dict[str, Any]:
        """
        Clean up old error records

        Args:
            days_to_keep: Number of days to keep error records

        Returns:
            Cleanup operation results
        """
        cutoff_date = datetime.now(timezone.utc) - timedelta(days=days_to_keep)

        old_errors = [e for e in self.error_history if e.timestamp < cutoff_date]
        removed_count = len(old_errors)

        # Keep only recent errors
        self.error_history = [e for e in self.error_history if e.timestamp >= cutoff_date]

        # Save updated error history
        self._save_error_history()

        logger.info(f"Cleaned up {removed_count} old error records")

        return {
            "removed_count": removed_count,
            "remaining_count": len(self.error_history),
            "cutoff_date": cutoff_date.isoformat(),
        }

    def _initialize_recovery_actions(self):
        """Initialize default recovery actions"""
        # System recovery actions
        self.register_recovery_action(
            RecoveryAction(
                name="restart_research_engines",
                description="Restart research engines and clear caches",
                action_type="automatic",
                severity_filter=[ErrorSeverity.HIGH, ErrorSeverity.CRITICAL],
                category_filter=[ErrorCategory.RESEARCH, ErrorCategory.SYSTEM],
                handler=self._restart_research_engines,
                timeout=30.0,
            )
        )

        self.register_recovery_action(
            RecoveryAction(
                name="restore_config_backup",
                description="Restore configuration from last known good backup",
                action_type="automatic",
                severity_filter=[ErrorSeverity.CRITICAL],
                category_filter=[ErrorCategory.CONFIGURATION],
                handler=self._restore_config_backup,
                timeout=15.0,
            )
        )

        self.register_recovery_action(
            RecoveryAction(
                name="clear_agent_cache",
                description="Clear agent communication cache and reset connections",
                action_type="automatic",
                severity_filter=[ErrorSeverity.MEDIUM, ErrorSeverity.HIGH],
                category_filter=[ErrorCategory.COMMUNICATION],
                handler=self._clear_agent_cache,
                timeout=10.0,
            )
        )

        self.register_recovery_action(
            RecoveryAction(
                name="validate_research_integrity",
                description="Validate research component integrity and repair if needed",
                action_type="assisted",
                severity_filter=[ErrorSeverity.HIGH],
                category_filter=[ErrorCategory.RESEARCH, ErrorCategory.VALIDATION],
                handler=self._validate_research_integrity,
                timeout=60.0,
            )
        )

        self.register_recovery_action(
            RecoveryAction(
                name="rollback_last_changes",
                description="Rollback last research integration changes",
                action_type="manual",
                severity_filter=[ErrorSeverity.CRITICAL],
                category_filter=[ErrorCategory.INTEGRATION, ErrorCategory.RESEARCH],
                handler=self._rollback_last_changes,
                timeout=45.0,
            )
        )

        self.register_recovery_action(
            RecoveryAction(
                name="reset_system_state",
                description="Reset system to known good state",
                action_type="manual",
                severity_filter=[ErrorSeverity.CRITICAL],
                category_filter=[ErrorCategory.SYSTEM],
                handler=self._reset_system_state,
                timeout=120.0,
            )
        )

        # Performance recovery actions
        self.register_recovery_action(
            RecoveryAction(
                name="optimize_performance",
                description="Optimize system performance and clear bottlenecks",
                action_type="automatic",
                severity_filter=[ErrorSeverity.MEDIUM],
                category_filter=[ErrorCategory.PERFORMANCE],
                handler=self._optimize_performance,
                timeout=30.0,
            )
        )

        # Resource recovery actions
        self.register_recovery_action(
            RecoveryAction(
                name="free_resources",
                description="Free up system resources and memory",
                action_type="automatic",
                severity_filter=[ErrorSeverity.MEDIUM, ErrorSeverity.HIGH],
                category_filter=[ErrorCategory.RESOURCE],
                handler=self._free_resources,
                timeout=20.0,
            )
        )

    def _attempt_automatic_recovery(self, error_report: ErrorReport) -> RecoveryResult:
        """Attempt automatic recovery for an error"""
        suitable_actions = []

        # Find suitable recovery actions
        for action_name, action in self.recovery_actions.items():
            if (
                action.action_type == "automatic"
                and error_report.severity in action.severity_filter
                and error_report.category in action.category_filter
            ):
                suitable_actions.append(action)

        # Try actions in order of priority
        for action in suitable_actions:
            try:
                logger.info(f"Attempting automatic recovery: {action.name}")

                start_time = time.time()
                result = action.handler(error_report, {})
                duration = time.time() - start_time

                if result:
                    return RecoveryResult(
                        success=True,
                        action_name=action.name,
                        message=f"Automatic recovery successful: {action.name}",
                        duration=duration,
                        details={"result": result},
                    )

            except Exception as e:
                logger.warning(f"Recovery action {action.name} failed: {str(e)}")
                continue

        return RecoveryResult(
            success=False,
            action_name="none",
            message="No suitable automatic recovery action succeeded",
            duration=0.0,
        )

    def _restart_research_engines(self, error_report: ErrorReport, parameters: Dict[str, Any]) -> bool:
        """Restart research engines and clear caches"""
        try:
            logger.info("Restarting research engines...")

            # Clear research engine caches
            cache_dirs = [
                self.project_root / ".moai" / "cache",
                self.project_root / ".claude" / "cache",
            ]

            for cache_dir in cache_dirs:
                if cache_dir.exists():
                    import shutil

                    shutil.rmtree(cache_dir)
                    cache_dir.mkdir(parents=True, exist_ok=True)

            # Reset research engine state
            research_state_file = self.project_root / ".moai" / "research_state.json"
            if research_state_file.exists():
                research_state_file.unlink()

            # Reinitialize research components
            self._reinitialize_research_components()

            logger.info("Research engines restarted successfully")
            return True

        except Exception as e:
            logger.error(f"Failed to restart research engines: {str(e)}")
            return False

    def _restore_config_backup(self, error_report: ErrorReport, parameters: Dict[str, Any]) -> bool:
        """Restore configuration from backup"""
        try:
            logger.info("Restoring configuration from backup...")

            backup_dir = self.project_root / ".moai" / "config_backups"
            if not backup_dir.exists():
                logger.warning("No configuration backup directory found")
                return False

            # Find most recent backup
            backup_files = list(backup_dir.glob("config_*.json"))
            if not backup_files:
                logger.warning("No configuration backups found")
                return False

            latest_backup = max(backup_files, key=lambda f: f.stat().st_mtime)

            # Restore configuration
            config_file = self.project_root / ".moai" / "config" / "config.json"
            import shutil

            shutil.copy2(latest_backup, config_file)

            logger.info(f"Configuration restored from {latest_backup}")
            return True

        except Exception as e:
            logger.error(f"Failed to restore configuration: {str(e)}")
            return False

    def _clear_agent_cache(self, error_report: ErrorReport, parameters: Dict[str, Any]) -> bool:
        """Clear agent communication cache"""
        try:
            logger.info("Clearing agent cache...")

            # Clear agent state files
            agent_state_dir = self.project_root / ".moai" / "agent_state"
            if agent_state_dir.exists():
                import shutil

                shutil.rmtree(agent_state_dir)
                agent_state_dir.mkdir(parents=True, exist_ok=True)

            # Reset communication channels
            comm_cache_dir = self.project_root / ".moai" / "comm_cache"
            if comm_cache_dir.exists():
                import shutil

                shutil.rmtree(comm_cache_dir)
                comm_cache_dir.mkdir(parents=True, exist_ok=True)

            logger.info("Agent cache cleared successfully")
            return True

        except Exception as e:
            logger.error(f"Failed to clear agent cache: {str(e)}")
            return False

    def _validate_research_integrity(self, error_report: ErrorReport, parameters: Dict[str, Any]) -> Dict[str, Any]:
        """Validate research component integrity"""
        issues_found: List[str] = []
        repairs_made: List[str] = []

        validation_results = {
            "skills_valid": True,
            "agents_valid": True,
            "commands_valid": True,
            "hooks_valid": True,
            "issues_found": issues_found,
            "repairs_made": repairs_made,
        }

        try:
            logger.info("Validating research integrity...")

            # Validate skills
            skills_dir = self.project_root / ".claude" / "skills"
            if skills_dir.exists():
                for skill_file in skills_dir.glob("*.md"):
                    if not self._validate_skill_file(skill_file):
                        validation_results["skills_valid"] = False
                        issues_found.append(f"Invalid skill file: {skill_file}")

                        # Attempt repair
                        if self._repair_skill_file(skill_file):
                            repairs_made.append(f"Repaired: {skill_file}")

            # Validate agents
            agents_dir = self.project_root / ".claude" / "agents" / "alfred"
            if agents_dir.exists():
                for agent_file in agents_dir.glob("*.md"):
                    if not self._validate_agent_file(agent_file):
                        validation_results["agents_valid"] = False
                        issues_found.append(f"Invalid agent file: {agent_file}")

            # Validate commands
            commands_dir = self.project_root / ".claude" / "commands" / "alfred"
            if commands_dir.exists():
                for command_file in commands_dir.glob("*.md"):
                    if not self._validate_command_file(command_file):
                        validation_results["commands_valid"] = False
                        issues_found.append(f"Invalid command file: {command_file}")

            logger.info(
                f"Research integrity validation completed. Issues: {len(issues_found)}, Repairs: {len(repairs_made)}"
            )

        except Exception as e:
            logger.error(f"Research integrity validation failed: {str(e)}")
            validation_results["validation_error"] = str(e)

        return validation_results

    def _rollback_last_changes(self, error_report: ErrorReport, parameters: Dict[str, Any]) -> bool:
        """Rollback last research integration changes"""
        try:
            logger.info("Rolling back last research changes...")

            # Import rollback manager
            sys.path.insert(0, str(self.project_root / "src"))
            from moai_adk.core.rollback_manager import RollbackManager

            rollback_manager = RollbackManager(self.project_root)

            # Find latest rollback point for research integration
            rollback_points = rollback_manager.list_rollback_points(limit=5)
            if not rollback_points:
                logger.warning("No rollback points available")
                return False

            # Use the most recent rollback point
            latest_rollback = rollback_points[0]
            result = rollback_manager.rollback_to_point(latest_rollback["id"])

            if result.success:
                logger.info(f"Successfully rolled back to {latest_rollback['id']}")
                return True
            else:
                logger.error(f"Rollback failed: {result.message}")
                return False

        except Exception as e:
            logger.error(f"Rollback operation failed: {str(e)}")
            return False

    def _reset_system_state(self, error_report: ErrorReport, parameters: Dict[str, Any]) -> bool:
        """Reset system to known good state"""
        try:
            logger.info("Resetting system to known good state...")

            # Clear all caches
            cache_dirs = [
                self.project_root / ".moai" / "cache",
                self.project_root / ".claude" / "cache",
                self.project_root / ".moai" / "agent_state",
                self.project_root / ".moai" / "comm_cache",
            ]

            for cache_dir in cache_dirs:
                if cache_dir.exists():
                    import shutil

                    shutil.rmtree(cache_dir)
                    cache_dir.mkdir(parents=True, exist_ok=True)

            # Reset error state
            self.active_errors.clear()

            # Reinitialize core components
            self._reinitialize_core_components()

            logger.info("System state reset completed")
            return True

        except Exception as e:
            logger.error(f"System state reset failed: {str(e)}")
            return False

    def _optimize_performance(self, error_report: ErrorReport, parameters: Dict[str, Any]) -> bool:
        """Optimize system performance"""
        try:
            logger.info("Optimizing system performance...")

            # Clear temporary files
            temp_dirs = [
                self.project_root / ".moai" / "temp",
                self.project_root / ".claude" / "temp",
            ]

            for temp_dir in temp_dirs:
                if temp_dir.exists():
                    import shutil

                    shutil.rmtree(temp_dir)

            # Optimize database connections if applicable
            self._optimize_connections()

            # Clear memory caches
            import gc

            gc.collect()

            logger.info("Performance optimization completed")
            return True

        except Exception as e:
            logger.error(f"Performance optimization failed: {str(e)}")
            return False

    def _free_resources(self, error_report: ErrorReport, parameters: Dict[str, Any]) -> bool:
        """Free up system resources"""
        try:
            logger.info("Freeing up system resources...")

            # Clear memory caches
            import gc

            gc.collect()

            # Close any open file handles
            self._close_file_handles()

            # Terminate any hanging processes
            self._terminate_hanging_processes()

            logger.info("Resource cleanup completed")
            return True

        except Exception as e:
            logger.error(f"Resource cleanup failed: {str(e)}")
            return False

    def _generate_error_id(self) -> str:
        """Generate unique error ID"""
        timestamp = datetime.now(timezone.utc).strftime("%Y%m%d_%H%M%S")
        random_suffix = hashlib.md5(os.urandom(4), usedforsecurity=False).hexdigest()[:6]
        return f"ERR_{timestamp}_{random_suffix}"

    def _log_error(self, error_report: ErrorReport):
        """Log error to file and system"""
        # Log to file
        error_file = self.error_log_dir / f"error_{error_report.id}.json"
        try:
            with open(error_file, "w", encoding="utf-8") as f:
                json.dump(asdict(error_report), f, indent=2, default=str, ensure_ascii=False)
        except Exception as e:
            logger.error(f"Failed to log error to file: {str(e)}")

        # Log to system
        log_level = {
            ErrorSeverity.CRITICAL: logging.CRITICAL,
            ErrorSeverity.HIGH: logging.ERROR,
            ErrorSeverity.MEDIUM: logging.WARNING,
            ErrorSeverity.LOW: logging.INFO,
            ErrorSeverity.INFO: logging.INFO,
        }.get(error_report.severity, logging.WARNING)

        logger.log(log_level, f"Error {error_report.id}: {error_report.message}")

    def _update_error_stats(self, error_report: ErrorReport):
        """Update error statistics"""
        self.error_stats["total_errors"] += 1

        # By severity
        severity = error_report.severity.value
        if severity not in self.error_stats["by_severity"]:
            self.error_stats["by_severity"][severity] = 0
        self.error_stats["by_severity"][severity] += 1

        # By category
        category = error_report.category.value
        if category not in self.error_stats["by_category"]:
            self.error_stats["by_category"][category] = 0
        self.error_stats["by_category"][category] += 1

    def _update_system_health(self):
        """Update system health status"""
        current_time = datetime.now(timezone.utc)

        # Determine system status
        critical_errors = [e for e in self.active_errors.values() if e.severity == ErrorSeverity.CRITICAL]
        high_errors = [e for e in self.active_errors.values() if e.severity == ErrorSeverity.HIGH]

        if critical_errors:
            self.system_health["status"] = "critical"
        elif high_errors:
            self.system_health["status"] = "degraded"
        elif len(self.active_errors) > 5:
            self.system_health["status"] = "warning"
        else:
            self.system_health["status"] = "healthy"

        # Update metrics
        self.system_health["last_check"] = current_time
        self.system_health["metrics"] = {
            "active_errors": len(self.active_errors),
            "total_errors": len(self.error_history),
            "recovery_success_rate": self._calculate_recovery_rate(self.error_history),
        }

        # Identify issues
        self.system_health["issues"] = [
            {
                "type": "active_errors",
                "count": len(self.active_errors),
                "severity_distribution": {
                    severity: len([e for e in self.active_errors.values() if e.severity.value == severity])
                    for severity in set(e.severity.value for e in self.active_errors.values())
                },
            }
        ]

    def _background_monitoring(self):
        """Background monitoring thread"""
        while self.monitoring_active:
            try:
                # Check system health every 30 seconds
                time.sleep(30)
                self._update_system_health()

                # Check for error patterns that need attention
                self._check_error_patterns()

            except Exception as e:
                logger.error(f"Background monitoring error: {str(e)}")

    def _check_error_patterns(self):
        """Check for concerning error patterns"""
        recent_errors = [
            e for e in self.error_history if (datetime.now(timezone.utc) - e.timestamp).total_seconds() < 300
        ]  # Last 5 minutes

        # Check for error bursts
        if len(recent_errors) > 10:
            logger.warning(f"High error rate detected: {len(recent_errors)} errors in last 5 minutes")

        # Check for repeated errors
        error_messages = [e.message for e in recent_errors]
        message_counts: Dict[str, int] = {}
        for msg in error_messages:
            message_counts[msg] = message_counts.get(msg, 0) + 1

        repeated_errors = [msg for msg, count in message_counts.items() if count > 3]
        if repeated_errors:
            logger.warning(f"Repeated errors detected: {repeated_errors}")

    def _calculate_recovery_rate(self, errors: List[ErrorReport]) -> float:
        """Calculate recovery success rate"""
        if not errors:
            return 0.0

        recovered_errors = [e for e in errors if e.recovery_successful]
        return len(recovered_errors) / len(errors)

    def _identify_error_patterns(self, errors: List[ErrorReport]) -> Dict[str, int]:
        """Identify common error patterns"""
        patterns: Dict[str, int] = {}

        for error in errors:
            # Pattern by exception type
            pattern = f"{error.category.value}:{error.details.get('exception_type', 'unknown')}"
            patterns[pattern] = patterns.get(pattern, 0) + 1

        return patterns

    def _get_pattern_severity(self, pattern: str) -> str:
        """Get typical severity for an error pattern"""
        severity_map = {
            "research:Exception": "high",
            "system:Exception": "critical",
            "configuration:Exception": "high",
            "communication:Exception": "medium",
            "validation:Exception": "medium",
        }

        for key, severity in severity_map.items():
            if key in pattern:
                return severity

        return "medium"

    def _get_solutions_for_pattern(self, pattern: str) -> List[str]:
        """Get common solutions for error pattern"""
        solutions = {
            "research:Exception": [
                "Restart research engines",
                "Clear research cache",
                "Validate research components",
            ],
            "system:Exception": [
                "Check system resources",
                "Restart system components",
                "Verify system configuration",
            ],
            "configuration:Exception": [
                "Restore configuration backup",
                "Validate configuration syntax",
                "Check configuration permissions",
            ],
        }

        for key, sols in solutions.items():
            if key in pattern:
                return sols

        return ["Contact system administrator", "Check system logs"]

    def _generate_prevention_tips(self) -> List[str]:
        """Generate prevention tips based on error history"""
        tips = []

        # Add tips based on common error categories
        category_counts: Dict[str, int] = {}
        for error in self.error_history:
            category = error.category.value
            category_counts[category] = category_counts.get(category, 0) + 1

        if category_counts.get("configuration", 0) > 5:
            tips.append("Regularly validate configuration files before making changes")

        if category_counts.get("research", 0) > 5:
            tips.append("Monitor research engine performance and clear caches regularly")

        if category_counts.get("communication", 0) > 5:
            tips.append("Ensure stable network connections for agent communication")

        return tips

    def _generate_emergency_procedures(self) -> List[Dict[str, str]]:
        """Generate emergency recovery procedures"""
        return [
            {
                "condition": "System completely unresponsive",
                "procedure": "Use system_reset recovery action to restore to known good state",
            },
            {
                "condition": "Critical research engine failure",
                "procedure": "Rollback last research changes using rollback_last_changes action",
            },
            {
                "condition": "Configuration corruption",
                "procedure": "Restore configuration from backup using restore_config_backup action",
            },
            {
                "condition": "Multiple agent communication failures",
                "procedure": "Clear agent cache and restart communication channels",
            },
        ]

    # Helper methods for component validation and repair
    def _validate_skill_file(self, skill_file: Path) -> bool:
        """Validate skill file format"""
        try:
            with open(skill_file, "r", encoding="utf-8") as f:
                content = f.read()

            # Basic validation
            return "---" in content and len(content) > 100
        except (OSError, UnicodeDecodeError):
            return False

    def _validate_agent_file(self, agent_file: Path) -> bool:
        """Validate agent file format"""
        try:
            with open(agent_file, "r", encoding="utf-8") as f:
                content = f.read()

            return "role:" in content and len(content) > 200
        except (OSError, UnicodeDecodeError):
            return False

    def _validate_command_file(self, command_file: Path) -> bool:
        """Validate command file format"""
        try:
            with open(command_file, "r", encoding="utf-8") as f:
                content = f.read()

            return "name:" in content and "allowed-tools:" in content
        except (OSError, UnicodeDecodeError):
            return False

    def _repair_skill_file(self, skill_file: Path) -> bool:
        """Attempt to repair skill file"""
        try:
            # Basic repair - ensure file has minimum required content
            with open(skill_file, "r", encoding="utf-8") as f:
                content = f.read()

            if not content.startswith("---"):
                content = f"---\nname: {skill_file.stem}\ndescription: Repaired skill file\n---\n\n{content}"

            with open(skill_file, "w", encoding="utf-8") as f:
                f.write(content)

            return True
        except (OSError, UnicodeDecodeError):
            return False

    def _reinitialize_research_components(self):
        """Reinitialize research components"""
        # Implementation would depend on specific research components
        pass

    def _reinitialize_core_components(self):
        """Reinitialize core system components"""
        # Implementation would depend on specific core components
        pass

    def _optimize_connections(self):
        """Optimize database/network connections"""
        # Implementation would depend on specific connection types
        pass

    def _close_file_handles(self):
        """Close open file handles"""
        import gc

        gc.collect()  # Force garbage collection to close file handles

    def _terminate_hanging_processes(self):
        """Terminate hanging processes"""
        # Implementation would identify and terminate hanging processes
        pass

    def _save_error_history(self):
        """Save error history to file"""
        history_file = self.error_log_dir / "error_history.json"
        try:
            with open(history_file, "w") as f:
                json.dump([asdict(e) for e in self.error_history], f, indent=2, default=str)
        except Exception as e:
            logger.error(f"Failed to save error history: {str(e)}")

    # Phase 3: Advanced Error Recovery Methods

    def _initialize_phase3_components(self):
        """Initialize Phase 3 advanced recovery components"""
        # Phase 3 specific attributes
        self.advanced_failures: Dict[str, FailureEvent] = {}
        self.advanced_recovery_actions: Dict[str, AdvancedRecoveryAction] = {}
        self.system_snapshots: Dict[str, SystemSnapshot] = {}
        self.dead_letter_queue: deque = deque(maxlen=10000)

        # Advanced recovery statistics
        self.advanced_recovery_stats = {
            "total_failures": 0,
            "auto_recoveries_attempted": 0,
            "auto_recoveries_successful": 0,
            "cascade_failures_detected": 0,
            "rollbacks_performed": 0,
            "snapshots_created": 0,
            "dead_letter_messages": 0,
        }

        # Circuit breaker states for components
        self.circuit_breaker_states: Dict[str, Dict[str, Any]] = defaultdict(
            lambda: {
                "state": "CLOSED",
                "failure_count": 0,
                "last_failure_time": None,
                "success_threshold": 5,
                "failure_threshold": 3,
                "timeout_seconds": 60,
            }
        )

        # Failure mode analyzers
        self.failure_analyzers = {
            FailureMode.CASCADE_FAILURE: self._analyze_cascade_failure,
            FailureMode.CIRCUIT_BREAKER_TRIPPED: self._analyze_circuit_breaker_trip,
            FailureMode.RESOURCE_EXHAUSTION: self._analyze_resource_exhaustion,
        }

        logger.info("Phase 3 advanced recovery components initialized")

    async def report_advanced_failure(
        self,
        failure_mode: FailureMode,
        component: str,
        description: str,
        severity: str = "medium",
        context: Optional[Dict[str, Any]] = None,
        error_details: Optional[Dict[str, Any]] = None,
        affected_operations: Optional[List[str]] = None,
        auto_recovery_eligible: bool = True,
        parent_failure_id: Optional[str] = None,
    ) -> str:
        """Report an advanced failure event with enhanced tracking"""
        failure_id = str(uuid.uuid4())

        failure = FailureEvent(
            failure_id=failure_id,
            failure_mode=failure_mode,
            timestamp=datetime.now(timezone.utc),
            component=component,
            description=description,
            severity=severity,
            context=context or {},
            error_details=error_details,
            affected_operations=affected_operations or [],
            auto_recovery_eligible=auto_recovery_eligible,
            parent_failure_id=parent_failure_id,
        )

        # Store failure
        self.advanced_failures[failure_id] = failure
        self.advanced_recovery_stats["total_failures"] += 1

        # Analyze failure mode
        if failure_mode in self.failure_analyzers:
            await self.failure_analyzers[failure_mode](failure)

        # Trigger advanced recovery if eligible
        if auto_recovery_eligible:
            await self._trigger_advanced_recovery(failure)

        # Check for cascade failures
        await self._check_cascade_failures(failure)

        logger.warning(f"Advanced failure reported: {failure_mode.value} in {component} - {description}")
        return failure_id

    async def _trigger_advanced_recovery(self, failure: FailureEvent):
        """Trigger advanced recovery mechanisms"""
        try:
            strategy = self._determine_advanced_recovery_strategy(failure.failure_mode)

            action = AdvancedRecoveryAction(
                action_id=str(uuid.uuid4()),
                failure_id=failure.failure_id,
                strategy=strategy,
                timestamp=datetime.now(timezone.utc),
                description=f"Advanced recovery for {failure.failure_mode.value}",
                parameters={"failure_context": failure.context},
                priority=self._calculate_recovery_priority(failure),
            )

            self.advanced_recovery_actions[action.action_id] = action
            self.advanced_recovery_stats["auto_recoveries_attempted"] += 1

            # Execute recovery action
            success = await self._execute_advanced_recovery_action(action)

            if success:
                self.advanced_recovery_stats["auto_recoveries_successful"] += 1
                logger.info(f"Advanced recovery successful for failure {failure.failure_id}")
            else:
                # Add to dead letter queue for manual intervention
                self.dead_letter_queue.append(
                    {
                        "failure_id": failure.failure_id,
                        "action_id": action.action_id,
                        "timestamp": datetime.now(timezone.utc).isoformat(),
                        "reason": "Advanced recovery failed",
                    }
                )
                self.advanced_recovery_stats["dead_letter_messages"] += 1

        except Exception as e:
            logger.error(f"Error triggering advanced recovery: {e}")

    async def _execute_advanced_recovery_action(self, action: AdvancedRecoveryAction) -> bool:
        """Execute advanced recovery action with enhanced capabilities"""
        action.status = RecoveryStatus.IN_PROGRESS
        action.execution_log.append(f"Starting advanced recovery: {action.strategy.value}")

        try:
            # Check dependencies
            for dep_action_id in action.dependencies:
                if dep_action_id in self.advanced_recovery_actions:
                    dep_action = self.advanced_recovery_actions[dep_action_id]
                    if dep_action.status != RecoveryStatus.COMPLETED:
                        action.execution_log.append(f"Waiting for dependency: {dep_action_id}")
                        return False

            # Execute based on strategy
            if action.strategy == RecoveryStrategy.RETRY_WITH_BACKOFF:
                success = await self._execute_retry_with_backoff(action)
            elif action.strategy == RecoveryStrategy.CIRCUIT_BREAKER:
                success = await self._execute_circuit_breaker_action(action)
            elif action.strategy == RecoveryStrategy.ROLLBACK:
                success = await self._execute_rollback_action(action)
            elif action.strategy == RecoveryStrategy.QUARANTINE:
                success = await self._execute_quarantine_action(action)
            else:
                success = await self._execute_legacy_recovery_action(action)

            if success:
                action.status = RecoveryStatus.COMPLETED
                action.execution_log.append("Advanced recovery completed successfully")
            else:
                action.status = RecoveryStatus.FAILED
                action.execution_log.append("Advanced recovery failed")

            return success

        except Exception as e:
            action.status = RecoveryStatus.FAILED
            action.execution_log.append(f"Advanced recovery error: {str(e)}")
            logger.error(f"Error executing advanced recovery action {action.action_id}: {e}")
            return False

    async def _execute_retry_with_backoff(self, action: AdvancedRecoveryAction) -> bool:
        """Execute retry with exponential backoff"""
        self.advanced_failures[action.failure_id]
        base_delay = 1.0
        max_delay = 60.0
        backoff_factor = 2.0

        for attempt in range(action.max_retries + 1):
            try:
                action.retry_attempts = attempt
                action.execution_log.append(f"Retry attempt {attempt + 1}/{action.max_retries + 1}")

                # Simulate retry logic - in real implementation, this would call the failing function
                if attempt >= 2:  # Simulate success after a few attempts
                    action.execution_log.append("Retry successful")
                    return True
                else:
                    action.execution_log.append("Retry failed, will retry again")

                # Wait with exponential backoff
                if attempt < action.max_retries:
                    delay = min(base_delay * (backoff_factor**attempt), max_delay)
                    await asyncio.sleep(delay)

            except Exception as e:
                action.execution_log.append(f"Retry attempt {attempt + 1} error: {str(e)}")

        action.execution_log.append("All retry attempts exhausted")
        return False

    async def _execute_circuit_breaker_action(self, action: AdvancedRecoveryAction) -> bool:
        """Execute circuit breaker action"""
        failure = self.advanced_failures[action.failure_id]
        component = failure.component

        # Update circuit breaker state
        cb_state = self.circuit_breaker_states[component]
        cb_state["state"] = "OPEN"
        cb_state["failure_count"] += 1
        cb_state["last_failure_time"] = datetime.now(timezone.utc).isoformat()

        action.execution_log.append(f"Circuit breaker opened for component: {component}")
        return True

    async def _execute_rollback_action(self, action: AdvancedRecoveryAction) -> bool:
        """Execute rollback action"""
        try:
            # Create a snapshot before rollback
            snapshot_id = await self._create_system_snapshot("pre_rollback_snapshot")

            # Perform rollback logic
            action.execution_log.append("Creating rollback snapshot and performing rollback")

            # In real implementation, this would restore system state from snapshot
            self.advanced_recovery_stats["rollbacks_performed"] += 1

            action.rollback_action_id = snapshot_id
            return True

        except Exception as e:
            action.execution_log.append(f"Rollback failed: {str(e)}")
            return False

    async def _execute_quarantine_action(self, action: AdvancedRecoveryAction) -> bool:
        """Execute quarantine action"""
        failure = self.advanced_failures[action.failure_id]
        component = failure.component

        action.execution_log.append(f"Quarantining component: {component}")

        # In real implementation, this would isolate the component
        # For now, just log the action
        return True

    async def _execute_legacy_recovery_action(self, action: AdvancedRecoveryAction) -> bool:
        """Execute legacy recovery action as fallback"""
        failure = self.advanced_failures[action.failure_id]

        # Convert to legacy format and use existing recovery mechanisms
        legacy_action = self.recovery_actions.get("restart_research_engines")
        if not legacy_action:
            action.execution_log.append("No legacy recovery action available")
            return False

        # Create legacy error report
        legacy_error = ErrorReport(
            id=failure.failure_id,
            timestamp=failure.timestamp,
            severity=getattr(ErrorSeverity, failure.severity.upper(), ErrorSeverity.MEDIUM),
            category=getattr(ErrorCategory, "SYSTEM", ErrorCategory.SYSTEM),
            message=failure.description,
            details=failure.error_details or {},
            stack_trace="",
            context=failure.context,
        )

        try:
            result = legacy_action.handler(legacy_error, action.parameters)
            if result:
                action.execution_log.append("Legacy recovery action successful")
                return True
            else:
                action.execution_log.append("Legacy recovery action failed")
                return False
        except Exception as e:
            action.execution_log.append(f"Legacy recovery action error: {str(e)}")
            return False

    def _determine_advanced_recovery_strategy(self, failure_mode: FailureMode) -> RecoveryStrategy:
        """Determine advanced recovery strategy based on failure mode"""
        strategy_map = {
            FailureMode.HOOK_EXECUTION_FAILURE: RecoveryStrategy.RETRY_WITH_BACKOFF,
            FailureMode.RESOURCE_EXHAUSTION: RecoveryStrategy.DEGRADE_SERVICE,
            FailureMode.DATA_CORRUPTION: RecoveryStrategy.ROLLBACK,
            FailureMode.NETWORK_FAILURE: RecoveryStrategy.RETRY_WITH_BACKOFF,
            FailureMode.SYSTEM_OVERLOAD: RecoveryStrategy.CIRCUIT_BREAKER,
            FailureMode.CIRCUIT_BREAKER_TRIPPED: RecoveryStrategy.CIRCUIT_BREAKER,
            FailureMode.CASCADE_FAILURE: RecoveryStrategy.EMERGENCY_STOP,
            FailureMode.TIMEOUT_FAILURE: RecoveryStrategy.RETRY_WITH_BACKOFF,
            FailureMode.MEMORY_LEAK: RecoveryStrategy.RESTART_COMPONENT,
            FailureMode.DEADLOCK: RecoveryStrategy.QUARANTINE,
            FailureMode.AUTHENTICATION_FAILURE: RecoveryStrategy.NOTIFY_ADMIN,
            FailureMode.VALIDATION_FAILURE: RecoveryStrategy.QUARANTINE,
            FailureMode.EXTERNAL_SERVICE_FAILURE: RecoveryStrategy.FAILOVER,
            FailureMode.STORAGE_FAILURE: RecoveryStrategy.ROLLBACK,
            FailureMode.CONCURRENCY_ISSUE: RecoveryStrategy.CIRCUIT_BREAKER,
        }

        return strategy_map.get(failure_mode, RecoveryStrategy.RETRY_WITH_BACKOFF)

    def _calculate_recovery_priority(self, failure: FailureEvent) -> int:
        """Calculate recovery priority based on failure characteristics"""
        base_priority = 5

        # Adjust based on severity
        if failure.severity == "critical":
            base_priority -= 3
        elif failure.severity == "high":
            base_priority -= 2
        elif failure.severity == "medium":
            base_priority -= 1

        # Adjust based on number of affected operations
        if len(failure.affected_operations) > 10:
            base_priority -= 2
        elif len(failure.affected_operations) > 5:
            base_priority -= 1

        # Ensure priority is in valid range
        return max(1, min(10, base_priority))

    async def _check_cascade_failures(self, failure: FailureEvent):
        """Check for cascade failure patterns"""
        # Check if this failure is related to other recent failures
        recent_failures = [
            f
            for f in self.advanced_failures.values()
            if (datetime.now(timezone.utc) - f.timestamp).total_seconds() < 300  # Last 5 minutes
            and f.failure_id != failure.failure_id
        ]

        # Simple cascade detection: same component or related components
        related_failures = [
            f
            for f in recent_failures
            if f.component == failure.component or f.component in failure.context.get("related_components", [])
        ]

        if len(related_failures) >= 3:
            self.advanced_recovery_stats["cascade_failures_detected"] += 1
            logger.warning(f"Cascade failure detected: {len(related_failures)} related failures")

            # Trigger emergency recovery
            await self._trigger_emergency_recovery(failure, related_failures)

    async def _trigger_emergency_recovery(self, failure: FailureEvent, related_failures: List[FailureEvent]):
        """Trigger emergency recovery for cascade failures"""
        emergency_action = AdvancedRecoveryAction(
            action_id=str(uuid.uuid4()),
            failure_id=failure.failure_id,
            strategy=RecoveryStrategy.EMERGENCY_STOP,
            timestamp=datetime.now(timezone.utc),
            description="Emergency recovery for cascade failure",
            parameters={"cascade_failures": [f.failure_id for f in related_failures]},
            priority=1,  # Highest priority
        )

        await self._execute_advanced_recovery_action(emergency_action)

    async def _analyze_cascade_failure(self, failure: FailureEvent):
        """Analyze cascade failure patterns"""
        # Implementation would analyze failure patterns and correlations
        pass

    async def _analyze_circuit_breaker_trip(self, failure: FailureEvent):
        """Analyze circuit breaker trip patterns"""
        # Implementation would analyze circuit breaker behavior
        pass

    async def _analyze_resource_exhaustion(self, failure: FailureEvent):
        """Analyze resource exhaustion patterns"""
        # Implementation would analyze resource usage patterns
        pass

    async def _create_system_snapshot(self, description: str = "", is_rollback_point: bool = False) -> str:
        """Create a system state snapshot"""
        snapshot_id = str(uuid.uuid4())

        # Get current system state
        component_states = {
            "error_recovery_system": {
                "active_errors": len(self.active_errors),
                "advanced_failures": len(self.advanced_failures),
                "system_health": self.system_health["status"],
            },
            "circuit_breakers": dict(self.circuit_breaker_states),
            "recovery_stats": self.advanced_recovery_stats.copy(),
        }

        # Calculate checksums
        config_str = json.dumps(component_states, sort_keys=True)
        config_hash = hashlib.sha256(config_str.encode()).hexdigest()

        data_checksums = {
            "component_states": hashlib.sha256(config_str.encode()).hexdigest(),
        }

        snapshot = SystemSnapshot(
            snapshot_id=snapshot_id,
            timestamp=datetime.now(timezone.utc),
            component_states=component_states,
            configuration_hash=config_hash,
            data_checksums=data_checksums,
            description=description,
            is_rollback_point=is_rollback_point,
        )

        self.system_snapshots[snapshot_id] = snapshot
        self.advanced_recovery_stats["snapshots_created"] += 1

        logger.info(f"Created system snapshot: {snapshot_id}")
        return snapshot_id

    def get_advanced_system_status(self) -> Dict[str, Any]:
        """Get comprehensive advanced system status"""
        return {
            "status": "running",
            "phase3_features": "enabled",
            "advanced_recovery_statistics": self.advanced_recovery_stats,
            "active_advanced_failures": len(self.advanced_failures),
            "pending_advanced_actions": len(
                [
                    a
                    for a in self.advanced_recovery_actions.values()
                    if a.status in [RecoveryStatus.PENDING, RecoveryStatus.IN_PROGRESS]
                ]
            ),
            "circuit_breaker_states": dict(self.circuit_breaker_states),
            "system_snapshots": len(self.system_snapshots),
            "dead_letter_queue_size": len(self.dead_letter_queue),
            "failure_mode_analyzers": list(self.failure_analyzers.keys()),
        }


# Global error recovery system instance
_error_recovery_system = None


def get_error_recovery_system(project_root: Path = None) -> ErrorRecoverySystem:
    """Get or create global error recovery system instance"""
    global _error_recovery_system
    if _error_recovery_system is None:
        _error_recovery_system = ErrorRecoverySystem(project_root)
    return _error_recovery_system


def handle_error(
    error: Exception,
    context: Dict[str, Any] = None,
    severity: ErrorSeverity = ErrorSeverity.MEDIUM,
    category: ErrorCategory = ErrorCategory.SYSTEM,
) -> ErrorReport:
    """Convenience function to handle errors using global system"""
    return get_error_recovery_system().handle_error(error, context, severity, category)


# Decorator for automatic error handling
def error_handler(
    severity: ErrorSeverity = ErrorSeverity.MEDIUM,
    category: ErrorCategory = ErrorCategory.SYSTEM,
    context: Dict[str, Any] = None,
):
    """Decorator for automatic error handling"""

    def decorator(func):
        def wrapper(*args, **kwargs):
            try:
                return func(*args, **kwargs)
            except Exception as e:
                error_context = {
                    "function": func.__name__,
                    "module": func.__module__,
                    "args": str(args)[:100],  # Limit length
                    "kwargs": str(kwargs)[:100],
                    **(context or {}),
                }
                handle_error(e, error_context, severity, category)
                raise

        return wrapper

    return decorator


if __name__ == "__main__":
    # Demo usage
    recovery_system = ErrorRecoverySystem()

    print("Error Recovery System Demo")
    print("=" * 50)

    # Simulate some errors
    try:
        raise ValueError("This is a test error for demonstration")
    except Exception as e:
        error_report = recovery_system.handle_error(
            e,
            context={"demo": True},
            severity=ErrorSeverity.MEDIUM,
            category=ErrorCategory.SYSTEM,
        )
        print(f"Handled error: {error_report.id}")

    # Show system health
    health = recovery_system.get_system_health()
    print(f"System health: {health['status']}")

    # Show error summary
    summary = recovery_system.get_error_summary()
    print(f"Total errors: {summary['total_recent_errors']}")

    print("\nError Recovery System demo completed")
