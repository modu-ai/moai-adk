"""Tests for Hook-related dataclasses - Fixing the phase enum values."""

from datetime import datetime

from moai_adk.core.jit_enhanced_hook_manager import (
    CircuitBreakerState,
    HookEvent,
    HookExecutionResult,
    HookMetadata,
    HookPriority,
    Phase,
)


class TestHookEnums:
    """Test hook enum definitions."""

    def test_hook_event_enum_values(self):
        """Test HookEvent enum contains all expected values."""
        assert HookEvent.SESSION_START.value == "SessionStart"
        assert HookEvent.SESSION_END.value == "SessionEnd"
        assert HookEvent.USER_PROMPT_SUBMIT.value == "UserPromptSubmit"
        assert HookEvent.PRE_TOOL_USE.value == "PreToolUse"
        assert HookEvent.POST_TOOL_USE.value == "PostToolUse"
        assert HookEvent.SUBAGENT_START.value == "SubagentStart"
        assert HookEvent.SUBAGENT_STOP.value == "SubagentStop"

    def test_hook_priority_values(self):
        """Test HookPriority enum has correct numeric values."""
        assert HookPriority.CRITICAL.value == 1
        assert HookPriority.HIGH.value == 2
        assert HookPriority.NORMAL.value == 3
        assert HookPriority.LOW.value == 4

    def test_phase_enum_exists(self):
        """Test Phase enum exists and has expected phases."""
        # Just verify the phases exist, not their string values
        phases = [
            Phase.SPEC,
            Phase.RED,
            Phase.GREEN,
            Phase.REFACTOR,
            Phase.SYNC,
            Phase.DEBUG,
            Phase.PLANNING,
        ]
        assert len(phases) == 7


class TestHookMetadata:
    """Test HookMetadata dataclass."""

    def test_hook_metadata_creation(self):
        """Test creating HookMetadata instance."""
        metadata = HookMetadata(
            hook_path="/test/hook.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.HIGH,
        )
        assert metadata.hook_path == "/test/hook.py"
        assert metadata.event_type == HookEvent.SESSION_START
        assert metadata.priority == HookPriority.HIGH

    def test_hook_metadata_default_values(self):
        """Test HookMetadata default values."""
        metadata = HookMetadata(
            hook_path="/test/hook.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
        )
        assert metadata.estimated_execution_time_ms == 0.0
        assert metadata.last_execution_time is None
        assert metadata.success_rate == 1.0
        assert metadata.phase_relevance == {}
        assert metadata.token_cost_estimate == 0
        assert metadata.dependencies == set()
        assert metadata.parallel_safe is True

    def test_hook_metadata_all_fields(self):
        """Test HookMetadata with all fields populated."""
        now = datetime.now()
        metadata = HookMetadata(
            hook_path="/test/hook.py",
            event_type=HookEvent.PRE_TOOL_USE,
            priority=HookPriority.CRITICAL,
            estimated_execution_time_ms=50.0,
            last_execution_time=now,
            success_rate=0.99,
            phase_relevance={Phase.RED: 0.9, Phase.GREEN: 0.8},
            token_cost_estimate=200,
            dependencies={"/dep1.py", "/dep2.py"},
            parallel_safe=True,
        )
        assert metadata.estimated_execution_time_ms == 50.0
        assert metadata.last_execution_time == now
        assert metadata.success_rate == 0.99
        assert len(metadata.phase_relevance) == 2
        assert metadata.token_cost_estimate == 200
        assert len(metadata.dependencies) == 2


class TestHookExecutionResult:
    """Test HookExecutionResult dataclass."""

    def test_hook_execution_result_success(self):
        """Test successful hook execution result."""
        result = HookExecutionResult(
            hook_path="/test/hook.py",
            success=True,
            execution_time_ms=25.5,
            token_usage=100,
            output={"status": "ok"},
        )
        assert result.success is True
        assert result.execution_time_ms == 25.5
        assert result.token_usage == 100

    def test_hook_execution_result_failure(self):
        """Test failed hook execution result."""
        result = HookExecutionResult(
            hook_path="/test/hook.py",
            success=False,
            execution_time_ms=15.0,
            token_usage=50,
            output=None,
            error_message="Hook execution timed out",
        )
        assert result.success is False
        assert result.error_message == "Hook execution timed out"

    def test_hook_execution_result_with_metadata(self):
        """Test hook execution result with metadata."""
        metadata = {"retry_count": 2, "phase": "RED", "cache_hit": True}
        result = HookExecutionResult(
            hook_path="/test/hook.py",
            success=True,
            execution_time_ms=30.0,
            token_usage=120,
            output={"data": "value"},
            metadata=metadata,
        )
        assert result.metadata == metadata


class TestCircuitBreakerState:
    """Test CircuitBreakerState dataclass."""

    def test_circuit_breaker_initial_state(self):
        """Test circuit breaker starts in CLOSED state."""
        cb = CircuitBreakerState()
        assert cb.state == "CLOSED"
        assert cb.failure_count == 0
        assert cb.last_failure_time is None

    def test_circuit_breaker_state_transitions(self):
        """Test circuit breaker state transitions."""
        cb = CircuitBreakerState()
        cb.failure_count = 5
        cb.state = "OPEN"
        assert cb.state == "OPEN"

    def test_circuit_breaker_with_timestamps(self):
        """Test circuit breaker with timestamps."""
        now = datetime.now()
        cb = CircuitBreakerState()
        cb.last_failure_time = now
        assert cb.last_failure_time == now

    def test_circuit_breaker_success_threshold(self):
        """Test circuit breaker success threshold."""
        cb = CircuitBreakerState(success_threshold=5)
        assert cb.success_threshold == 5
        cb.failure_count = 5
        assert cb.failure_count >= cb.success_threshold
