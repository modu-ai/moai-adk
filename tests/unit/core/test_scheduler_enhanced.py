"""
Enhanced tests for Phase-Optimized Hook Scheduler - targeting 60%+ coverage.

Focus on actual API and enumerations.
"""

from datetime import datetime

import pytest

from moai_adk.core.phase_optimized_hook_scheduler import (
    ExecutionGroup,
    HookSchedulingContext,
    PhaseOptimizedHookScheduler,
    SchedulingDecision,
    SchedulingResult,
    SchedulingStrategy,
)


class TestSchedulingStrategy:
    """Test scheduling strategy enumeration."""

    def test_all_strategies_defined(self):
        """Test all scheduling strategies."""
        assert SchedulingStrategy.PRIORITY_FIRST.value == "priority_first"
        assert SchedulingStrategy.PERFORMANCE_FIRST.value == "performance_first"
        assert SchedulingStrategy.PHASE_OPTIMIZED.value == "phase_optimized"
        assert SchedulingStrategy.TOKEN_EFFICIENT.value == "token_efficient"
        assert SchedulingStrategy.BALANCED.value == "balanced"


class TestSchedulingDecision:
    """Test scheduling decision enumeration."""

    def test_all_decisions_defined(self):
        """Test all decisions exist."""
        assert SchedulingDecision.EXECUTE.value == "execute"
        assert SchedulingDecision.DEFER.value == "defer"
        assert SchedulingDecision.SKIP.value == "skip"
        assert SchedulingDecision.PARALLEL.value == "parallel"
        assert SchedulingDecision.SEQUENTIAL.value == "sequential"


class TestHookSchedulingContext:
    """Test scheduling context dataclass."""

    def test_context_creation(self):
        """Test creating scheduling context."""
        from moai_adk.core.jit_enhanced_hook_manager import HookEvent, Phase

        now = datetime.now()
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.PLANNING,
            user_input="test",
            available_token_budget=50000,
            max_execution_time_ms=5000.0,
            current_time=now,
        )
        assert context.available_token_budget == 50000
        assert context.max_execution_time_ms == 5000.0


class TestSchedulingResult:
    """Test scheduling result dataclass."""

    def test_result_creation(self):
        """Test creating scheduling result."""
        result = SchedulingResult(
            scheduled_hooks=[],
            execution_plan=[],
            estimated_total_time_ms=500.0,
            estimated_total_tokens=1000,
            skipped_hooks=[],
            deferred_hooks=[],
            scheduling_strategy=SchedulingStrategy.BALANCED,
        )
        assert result.estimated_total_time_ms == 500.0
        assert result.estimated_total_tokens == 1000


class TestExecutionGroup:
    """Test execution group dataclass."""

    def test_execution_group_creation(self):
        """Test creating execution group."""
        group = ExecutionGroup(
            group_id=1,
            hooks=[],
            execution_type="parallel",
            estimated_time_ms=100.0,
            estimated_tokens=500,
            max_wait_time_ms=5000.0,
        )
        assert group.group_id == 1


class TestSchedulerInitialization:
    """Test scheduler initialization."""

    def test_scheduler_creation(self):
        """Test creating scheduler."""
        scheduler = PhaseOptimizedHookScheduler()
        assert scheduler is not None


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
