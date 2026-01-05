"""
Minimal import and instantiation tests for Phase Optimized Hook Scheduler.

These tests verify that the module can be imported and basic classes
can be instantiated without errors.
"""


from moai_adk.core.jit_enhanced_hook_manager import (
    HookEvent,
    HookMetadata,
    HookPriority,
    Phase,
)
from moai_adk.core.phase_optimized_hook_scheduler import (
    ExecutionGroup,
    HookSchedulingContext,
    ScheduledHook,
    SchedulingDecision,
    SchedulingResult,
    SchedulingStrategy,
)


class TestImports:
    """Test that all enums and classes can be imported."""

    def test_scheduling_strategy_enum_exists(self):
        """Test SchedulingStrategy enum is importable."""
        assert SchedulingStrategy is not None
        assert hasattr(SchedulingStrategy, "PRIORITY_FIRST")

    def test_scheduling_decision_enum_exists(self):
        """Test SchedulingDecision enum is importable."""
        assert SchedulingDecision is not None
        assert hasattr(SchedulingDecision, "EXECUTE")

    def test_hook_scheduling_context_exists(self):
        """Test HookSchedulingContext class is importable."""
        assert HookSchedulingContext is not None

    def test_scheduled_hook_exists(self):
        """Test ScheduledHook class is importable."""
        assert ScheduledHook is not None

    def test_scheduling_result_exists(self):
        """Test SchedulingResult class is importable."""
        assert SchedulingResult is not None

    def test_execution_group_exists(self):
        """Test ExecutionGroup class is importable."""
        assert ExecutionGroup is not None


class TestSchedulingStrategyEnum:
    """Test SchedulingStrategy enum values."""

    def test_scheduling_strategy_priority_first(self):
        """Test SchedulingStrategy has PRIORITY_FIRST."""
        assert hasattr(SchedulingStrategy, "PRIORITY_FIRST")
        assert SchedulingStrategy.PRIORITY_FIRST.value == "priority_first"

    def test_scheduling_strategy_performance_first(self):
        """Test SchedulingStrategy has PERFORMANCE_FIRST."""
        assert hasattr(SchedulingStrategy, "PERFORMANCE_FIRST")

    def test_scheduling_strategy_phase_optimized(self):
        """Test SchedulingStrategy has PHASE_OPTIMIZED."""
        assert hasattr(SchedulingStrategy, "PHASE_OPTIMIZED")

    def test_scheduling_strategy_token_efficient(self):
        """Test SchedulingStrategy has TOKEN_EFFICIENT."""
        assert hasattr(SchedulingStrategy, "TOKEN_EFFICIENT")

    def test_scheduling_strategy_balanced(self):
        """Test SchedulingStrategy has BALANCED."""
        assert hasattr(SchedulingStrategy, "BALANCED")


class TestSchedulingDecisionEnum:
    """Test SchedulingDecision enum values."""

    def test_scheduling_decision_execute(self):
        """Test SchedulingDecision has EXECUTE."""
        assert hasattr(SchedulingDecision, "EXECUTE")
        assert SchedulingDecision.EXECUTE.value == "execute"

    def test_scheduling_decision_defer(self):
        """Test SchedulingDecision has DEFER."""
        assert hasattr(SchedulingDecision, "DEFER")
        assert SchedulingDecision.DEFER.value == "defer"

    def test_scheduling_decision_skip(self):
        """Test SchedulingDecision has SKIP."""
        assert hasattr(SchedulingDecision, "SKIP")
        assert SchedulingDecision.SKIP.value == "skip"

    def test_scheduling_decision_parallel(self):
        """Test SchedulingDecision has PARALLEL."""
        assert hasattr(SchedulingDecision, "PARALLEL")

    def test_scheduling_decision_sequential(self):
        """Test SchedulingDecision has SEQUENTIAL."""
        assert hasattr(SchedulingDecision, "SEQUENTIAL")


class TestHookSchedulingContextInstantiation:
    """Test HookSchedulingContext dataclass instantiation."""

    def test_hook_scheduling_context_basic_init(self):
        """Test HookSchedulingContext can be instantiated."""
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.RED,
            user_input="test input",
            available_token_budget=50000,
            max_execution_time_ms=5000,
        )
        assert context.event_type == HookEvent.SESSION_START
        assert context.current_phase == Phase.RED
        assert context.available_token_budget == 50000

    def test_hook_scheduling_context_defaults(self):
        """Test HookSchedulingContext respects default values."""
        context = HookSchedulingContext(
            event_type=HookEvent.PRE_TOOL_USE,
            current_phase=Phase.GREEN,
            user_input="input",
            available_token_budget=30000,
            max_execution_time_ms=3000,
        )
        assert context.system_load == 0.5
        assert isinstance(context.recent_performance, dict)
        assert isinstance(context.active_dependencies, set)

    def test_hook_scheduling_context_with_custom_values(self):
        """Test HookSchedulingContext with custom values."""
        context = HookSchedulingContext(
            event_type=HookEvent.POST_TOOL_USE,
            current_phase=Phase.REFACTOR,
            user_input="input",
            available_token_budget=40000,
            max_execution_time_ms=4000,
            system_load=0.8,
            recent_performance={"hook1": 150.0, "hook2": 200.0},
        )
        assert context.system_load == 0.8
        assert "hook1" in context.recent_performance
        assert context.recent_performance["hook1"] == 150.0


class TestScheduledHookInstantiation:
    """Test ScheduledHook dataclass instantiation."""

    def test_scheduled_hook_basic_init(self):
        """Test ScheduledHook can be instantiated."""
        metadata = HookMetadata(
            hook_path="/test/hook",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
        )
        hook = ScheduledHook(
            hook_path="/test/hook",
            metadata=metadata,
            priority_score=0.75,
            estimated_cost=100,
            estimated_time_ms=150.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )
        assert hook.hook_path == "/test/hook"
        assert hook.priority_score == 0.75

    def test_scheduled_hook_defaults(self):
        """Test ScheduledHook respects default values."""
        metadata = HookMetadata(
            hook_path="/test",
            event_type=HookEvent.PRE_TOOL_USE,
            priority=HookPriority.HIGH,
        )
        hook = ScheduledHook(
            hook_path="/test",
            metadata=metadata,
            priority_score=0.9,
            estimated_cost=200,
            estimated_time_ms=100.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )
        assert hook.execution_group is None
        assert isinstance(hook.dependencies, set)
        assert isinstance(hook.dependents, set)
        assert hook.retry_count == 0
        assert hook.max_retries == 2

    def test_scheduled_hook_with_dependencies(self):
        """Test ScheduledHook with dependencies."""
        metadata = HookMetadata(
            hook_path="/hook2",
            event_type=HookEvent.POST_TOOL_USE,
            priority=HookPriority.NORMAL,
        )
        hook = ScheduledHook(
            hook_path="/hook2",
            metadata=metadata,
            priority_score=0.5,
            estimated_cost=100,
            estimated_time_ms=200.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
            dependencies={"/hook1"},
            dependents={"/hook3"},
        )
        assert "/hook1" in hook.dependencies
        assert "/hook3" in hook.dependents


class TestSchedulingResultInstantiation:
    """Test SchedulingResult dataclass instantiation."""

    def test_scheduling_result_basic_init(self):
        """Test SchedulingResult can be instantiated."""
        result = SchedulingResult(
            scheduled_hooks=[],
            execution_plan=[],
            estimated_total_time_ms=1000.0,
            estimated_total_tokens=5000,
            skipped_hooks=[],
            deferred_hooks=[],
            scheduling_strategy=SchedulingStrategy.BALANCED,
        )
        assert result.estimated_total_time_ms == 1000.0
        assert result.estimated_total_tokens == 5000
        assert result.scheduling_strategy == SchedulingStrategy.BALANCED

    def test_scheduling_result_with_hooks(self):
        """Test SchedulingResult with scheduled hooks."""
        metadata = HookMetadata(
            hook_path="/hook",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
        )
        hook = ScheduledHook(
            hook_path="/hook",
            metadata=metadata,
            priority_score=0.7,
            estimated_cost=100,
            estimated_time_ms=150.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )
        result = SchedulingResult(
            scheduled_hooks=[hook],
            execution_plan=[[hook]],
            estimated_total_time_ms=150.0,
            estimated_total_tokens=100,
            skipped_hooks=[],
            deferred_hooks=[],
            scheduling_strategy=SchedulingStrategy.PRIORITY_FIRST,
        )
        assert len(result.scheduled_hooks) == 1
        assert len(result.execution_plan) == 1

    def test_scheduling_result_different_strategies(self):
        """Test SchedulingResult with different strategies."""
        for strategy in SchedulingStrategy:
            result = SchedulingResult(
                scheduled_hooks=[],
                execution_plan=[],
                estimated_total_time_ms=0.0,
                estimated_total_tokens=0,
                skipped_hooks=[],
                deferred_hooks=[],
                scheduling_strategy=strategy,
            )
            assert result.scheduling_strategy == strategy


class TestExecutionGroupInstantiation:
    """Test ExecutionGroup dataclass instantiation."""

    def test_execution_group_basic_init(self):
        """Test ExecutionGroup can be instantiated."""
        group = ExecutionGroup(
            group_id=1,
            execution_type=SchedulingDecision.PARALLEL,
            hooks=[],
            estimated_time_ms=100.0,
            estimated_tokens=500,
            max_wait_time_ms=5000.0,
        )
        assert group.group_id == 1
        assert group.execution_type == SchedulingDecision.PARALLEL

    def test_execution_group_with_sequential_type(self):
        """Test ExecutionGroup can be instantiated with sequential type."""
        # ExecutionGroup should have group_id and execution_type
        group = ExecutionGroup(
            group_id=2,
            execution_type=SchedulingDecision.SEQUENTIAL,
            hooks=[],
            estimated_time_ms=200.0,
            estimated_tokens=1000,
            max_wait_time_ms=10000.0,
        )
        assert group.group_id == 2
        assert group.execution_type == SchedulingDecision.SEQUENTIAL
        assert isinstance(group.hooks, list)


class TestEnumValues:
    """Test enum value types and formats."""

    def test_scheduling_strategy_values_are_strings(self):
        """Test all SchedulingStrategy values are strings."""
        for strategy in SchedulingStrategy:
            assert isinstance(strategy.value, str)

    def test_scheduling_decision_values_are_strings(self):
        """Test all SchedulingDecision values are strings."""
        for decision in SchedulingDecision:
            assert isinstance(decision.value, str)
