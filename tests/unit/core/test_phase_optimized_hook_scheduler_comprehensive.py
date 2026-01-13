"""
Comprehensive TDD test suite for phase_optimized_hook_scheduler.py

This test suite covers:
- All enums (SchedulingStrategy, SchedulingDecision)
- Data classes (HookSchedulingContext, ScheduledHook, SchedulingResult, ExecutionGroup)
- PhaseOptimizedHookScheduler main class with all methods
- Hook scheduling logic
- Phase parameter initialization
- Priority score calculation
- Dependency resolution
- Execution group creation
- Strategy selection and optimization
- Statistics and insights generation
- Edge cases and error conditions

Coverage Target: 100%
"""

import asyncio
import time
from datetime import datetime
from unittest.mock import AsyncMock, Mock, patch
import pytest

from moai_adk.core.phase_optimized_hook_scheduler import (
    SchedulingStrategy,
    SchedulingDecision,
    HookSchedulingContext,
    ScheduledHook,
    SchedulingResult,
    ExecutionGroup,
    PhaseOptimizedHookScheduler,
)
from moai_adk.core.jit_enhanced_hook_manager import (
    HookEvent,
    HookMetadata,
    HookPriority,
    Phase,
)


# =============================================================================
# ENUM TESTS
# =============================================================================


class TestSchedulingStrategy:
    """Test SchedulingStrategy enum"""

    def test_all_strategies(self):
        """Test all SchedulingStrategy values"""
        assert SchedulingStrategy.PRIORITY_FIRST.value == "priority_first"
        assert SchedulingStrategy.PERFORMANCE_FIRST.value == "performance_first"
        assert SchedulingStrategy.PHASE_OPTIMIZED.value == "phase_optimized"
        assert SchedulingStrategy.TOKEN_EFFICIENT.value == "token_efficient"
        assert SchedulingStrategy.BALANCED.value == "balanced"


class TestSchedulingDecision:
    """Test SchedulingDecision enum"""

    def test_all_decisions(self):
        """Test all SchedulingDecision values"""
        assert SchedulingDecision.EXECUTE.value == "execute"
        assert SchedulingDecision.DEFER.value == "defer"
        assert SchedulingDecision.SKIP.value == "skip"
        assert SchedulingDecision.PARALLEL.value == "parallel"
        assert SchedulingDecision.SEQUENTIAL.value == "sequential"


# =============================================================================
# DATA CLASS TESTS
# =============================================================================


class TestHookSchedulingContext:
    """Test HookSchedulingContext dataclass"""

    def test_initialization(self):
        """Test HookSchedulingContext initialization"""
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.RED,
            user_input="test input",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )
        assert context.event_type == HookEvent.SESSION_START
        assert context.current_phase == Phase.RED
        assert context.user_input == "test input"
        assert context.available_token_budget == 10000
        assert context.max_execution_time_ms == 1000.0
        assert context.system_load == 0.5
        assert isinstance(context.recent_performance, dict)
        assert isinstance(context.active_dependencies, set)

    def test_with_custom_values(self):
        """Test with custom values"""
        perf = {"hook1": 100.0, "hook2": 200.0}
        deps = {"/dep1", "/dep2"}

        context = HookSchedulingContext(
            event_type=HookEvent.PRE_TOOL_USE,
            current_phase=Phase.GREEN,
            user_input="input",
            available_token_budget=15000,
            max_execution_time_ms=500.0,
            current_time=datetime.now(),
            system_load=0.8,
            recent_performance=perf,
            active_dependencies=deps,
        )
        assert context.system_load == 0.8
        assert context.recent_performance == perf
        assert context.active_dependencies == deps


class TestScheduledHook:
    """Test ScheduledHook dataclass"""

    def test_initialization(self):
        """Test ScheduledHook initialization"""
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
        assert hook.estimated_cost == 100
        assert hook.estimated_time_ms == 150.0
        assert hook.scheduling_decision == SchedulingDecision.EXECUTE
        assert hook.execution_group is None
        assert hook.dependencies == set()
        assert hook.dependents == set()
        assert hook.retry_count == 0
        assert hook.max_retries == 2

    def test_with_dependencies(self):
        """Test with dependencies"""
        metadata = HookMetadata(
            hook_path="/test/hook",
            event_type=HookEvent.POST_TOOL_USE,
            priority=HookPriority.HIGH,
            dependencies={"/dep1", "/dep2"},
        )
        hook = ScheduledHook(
            hook_path="/test/hook",
            metadata=metadata,
            priority_score=0.9,
            estimated_cost=200,
            estimated_time_ms=100.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
            dependencies={"/dep1", "/dep2"},
            dependents={"/dependent"},
            retry_count=1,
            max_retries=3,
        )
        assert "/dep1" in hook.dependencies
        assert "/dep2" in hook.dependencies
        assert "/dependent" in hook.dependents
        assert hook.retry_count == 1


class TestSchedulingResult:
    """Test SchedulingResult dataclass"""

    def test_initialization(self):
        """Test SchedulingResult initialization"""
        result = SchedulingResult(
            scheduled_hooks=[],
            execution_plan=[],
            estimated_total_time_ms=1000.0,
            estimated_total_tokens=5000,
            skipped_hooks=[],
            deferred_hooks=[],
            scheduling_strategy=SchedulingStrategy.PHASE_OPTIMIZED,
        )
        assert result.scheduled_hooks == []
        assert result.execution_plan == []
        assert result.estimated_total_time_ms == 1000.0
        assert result.estimated_total_tokens == 5000
        assert result.skipped_hooks == []
        assert result.deferred_hooks == []
        assert result.scheduling_strategy == SchedulingStrategy.PHASE_OPTIMIZED

    def test_with_hooks(self):
        """Test with scheduled hooks"""
        metadata = HookMetadata(
            hook_path="/test/hook",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
        )
        hook = ScheduledHook(
            hook_path="/test/hook",
            metadata=metadata,
            priority_score=0.7,
            estimated_cost=100,
            estimated_time_ms=150.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )
        skipped = ScheduledHook(
            hook_path="/skipped",
            metadata=metadata,
            priority_score=0.3,
            estimated_cost=1000,
            estimated_time_ms=500.0,
            scheduling_decision=SchedulingDecision.SKIP,
        )

        result = SchedulingResult(
            scheduled_hooks=[hook, skipped],
            execution_plan=[[hook]],
            estimated_total_time_ms=150.0,
            estimated_total_tokens=100,
            skipped_hooks=[skipped],
            deferred_hooks=[],
            scheduling_strategy=SchedulingStrategy.PRIORITY_FIRST,
        )
        assert len(result.scheduled_hooks) == 2
        assert len(result.execution_plan) == 1
        assert len(result.skipped_hooks) == 1


class TestExecutionGroup:
    """Test ExecutionGroup dataclass"""

    def test_initialization(self):
        """Test ExecutionGroup initialization"""
        metadata = HookMetadata(
            hook_path="/test/hook",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
        )
        hook = ScheduledHook(
            hook_path="/test/hook",
            metadata=metadata,
            priority_score=0.7,
            estimated_cost=100,
            estimated_time_ms=150.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )

        group = ExecutionGroup(
            group_id=1,
            execution_type=SchedulingDecision.PARALLEL,
            hooks=[hook],
            estimated_time_ms=150.0,
            estimated_tokens=100,
            max_wait_time_ms=150.0,
        )
        assert group.group_id == 1
        assert group.execution_type == SchedulingDecision.PARALLEL
        assert len(group.hooks) == 1
        assert group.estimated_time_ms == 150.0
        assert group.estimated_tokens == 100
        assert group.max_wait_time_ms == 150.0
        assert group.dependencies == set()


# =============================================================================
# PHASE OPTIMIZED HOOK SCHEDULER TESTS
# =============================================================================


class TestPhaseOptimizedHookScheduler:
    """Test PhaseOptimizedHookScheduler class"""

    @pytest.fixture
    def scheduler(self):
        """Create scheduler instance"""
        return PhaseOptimizedHookScheduler(
            default_strategy=SchedulingStrategy.PHASE_OPTIMIZED,
            max_parallel_groups=3,
            enable_adaptive_scheduling=True,
        )

    @pytest.fixture
    def sample_metadata(self):
        """Create sample hook metadata"""
        return {
            "/hook1.py": HookMetadata(
                hook_path="/hook1.py",
                event_type=HookEvent.SESSION_START,
                priority=HookPriority.CRITICAL,
                estimated_execution_time_ms=100.0,
                token_cost_estimate=500,
                phase_relevance={
                    Phase.SPEC: 0.9,
                    Phase.RED: 0.7,
                    Phase.GREEN: 0.5,
                },
                parallel_safe=True,
            ),
            "/hook2.py": HookMetadata(
                hook_path="/hook2.py",
                event_type=HookEvent.SESSION_START,
                priority=HookPriority.NORMAL,
                estimated_execution_time_ms=200.0,
                token_cost_estimate=1000,
                phase_relevance={
                    Phase.SPEC: 0.5,
                    Phase.RED: 0.9,
                    Phase.GREEN: 0.7,
                },
                parallel_safe=False,
            ),
            "/hook3.py": HookMetadata(
                hook_path="/hook3.py",
                event_type=HookEvent.SESSION_START,
                priority=HookPriority.HIGH,
                estimated_execution_time_ms=150.0,
                token_cost_estimate=750,
                phase_relevance={
                    Phase.SPEC: 0.8,
                    Phase.RED: 0.6,
                    Phase.GREEN: 0.9,
                },
                parallel_safe=True,
                dependencies={"/hook1.py"},
            ),
        }

    def test_initialization(self, scheduler):
        """Test scheduler initialization"""
        assert scheduler.default_strategy == SchedulingStrategy.PHASE_OPTIMIZED
        assert scheduler.max_parallel_groups == 3
        assert scheduler.enable_adaptive_scheduling is True
        assert scheduler.hook_manager is not None
        assert scheduler._scheduling_history == []
        assert isinstance(scheduler._phase_parameters, dict)

    def test_initialize_phase_parameters(self, scheduler):
        """Test phase parameter initialization"""
        params = scheduler._phase_parameters
        assert Phase.SPEC in params
        assert Phase.RED in params
        assert Phase.GREEN in params
        assert Phase.REFACTOR in params
        assert Phase.SYNC in params
        assert Phase.DEBUG in params
        assert Phase.PLANNING in params

        # Check structure
        spec_params = params[Phase.SPEC]
        assert "max_total_time_ms" in spec_params
        assert "token_budget_ratio" in spec_params
        assert "priority_weights" in spec_params
        assert "prefer_parallel" in spec_params

    def test_select_optimal_strategy_low_token_budget(self, scheduler):
        """Test strategy selection with low token budget"""
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.RED,
            user_input="test",
            available_token_budget=3000,  # Low budget
            max_execution_time_ms=1000.0,
        )
        strategy = scheduler._select_optimal_strategy(HookEvent.SESSION_START, context)
        assert strategy == SchedulingStrategy.TOKEN_EFFICIENT

    def test_select_optimal_strategy_tight_time(self, scheduler):
        """Test strategy selection with tight time constraint"""
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.RED,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=300,  # Tight time
        )
        strategy = scheduler._select_optimal_strategy(HookEvent.SESSION_START, context)
        assert strategy == SchedulingStrategy.PERFORMANCE_FIRST

    def test_select_optimal_strategy_high_load(self, scheduler):
        """Test strategy selection with high system load"""
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.GREEN,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
            system_load=0.9,  # High load
        )
        strategy = scheduler._select_optimal_strategy(HookEvent.SESSION_START, context)
        assert strategy == SchedulingStrategy.PRIORITY_FIRST

    def test_select_optimal_strategy_sync_phase(self, scheduler):
        """Test strategy selection for SYNC phase"""
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SYNC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )
        strategy = scheduler._select_optimal_strategy(HookEvent.SESSION_START, context)
        assert strategy == SchedulingStrategy.PHASE_OPTIMIZED

    def test_calculate_priority_score_priority_first(self, scheduler, sample_metadata):
        """Test priority score calculation for PRIORITY_FIRST strategy"""
        metadata = sample_metadata["/hook1.py"]
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )
        score = scheduler._calculate_priority_score(metadata, context, SchedulingStrategy.PRIORITY_FIRST)
        # CRITICAL priority should get high score
        assert score > 0

    def test_calculate_priority_score_performance_first(self, scheduler, sample_metadata):
        """Test priority score calculation for PERFORMANCE_FIRST strategy"""
        metadata = sample_metadata["/hook1.py"]
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )
        score = scheduler._calculate_priority_score(metadata, context, SchedulingStrategy.PERFORMANCE_FIRST)
        # Faster hooks get higher score
        assert score >= 0

    def test_calculate_priority_score_phase_optimized(self, scheduler, sample_metadata):
        """Test priority score calculation for PHASE_OPTIMIZED strategy"""
        metadata = sample_metadata["/hook1.py"]
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )
        score = scheduler._calculate_priority_score(metadata, context, SchedulingStrategy.PHASE_OPTIMIZED)
        # Should consider phase relevance
        assert score >= 0

    def test_estimate_hook_cost(self, scheduler, sample_metadata):
        """Test hook cost estimation"""
        metadata = sample_metadata["/hook1.py"]
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )
        cost = scheduler._estimate_hook_cost(metadata, context)
        assert cost >= metadata.token_cost_estimate

    def test_estimate_hook_cost_high_relevance(self, scheduler, sample_metadata):
        """Test cost estimation with high phase relevance"""
        metadata = sample_metadata["/hook1.py"]
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,  # High relevance (0.9)
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )
        cost = scheduler._estimate_hook_cost(metadata, context)
        # High relevance should reduce cost
        assert cost < metadata.token_cost_estimate

    def test_estimate_hook_time(self, scheduler, sample_metadata):
        """Test hook time estimation"""
        metadata = sample_metadata["/hook1.py"]
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
            system_load=0.5,
        )
        time_ms = scheduler._estimate_hook_time(metadata, context)
        # Time should be at least base time
        assert time_ms >= metadata.estimated_execution_time_ms

    def test_make_initial_scheduling_decision_critical(self, scheduler, sample_metadata):
        """Test scheduling decision for CRITICAL priority hook"""
        metadata = sample_metadata["/hook1.py"]  # CRITICAL
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )
        decision = scheduler._make_initial_scheduling_decision(metadata, context, 500, 100.0)
        assert decision == SchedulingDecision.EXECUTE

    def test_make_initial_scheduling_decision_over_budget(self, scheduler, sample_metadata):
        """Test scheduling decision when over budget"""
        metadata = sample_metadata["/hook2.py"]  # NORMAL
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=100,  # Very low budget
            max_execution_time_ms=1000.0,
        )
        decision = scheduler._make_initial_scheduling_decision(
            metadata,
            context,
            1000,
            100.0,  # Cost exceeds budget
        )
        assert decision == SchedulingDecision.SKIP

    def test_make_initial_scheduling_decision_over_time(self, scheduler, sample_metadata):
        """Test scheduling decision when over time limit"""
        metadata = sample_metadata["/hook2.py"]
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=50.0,  # Very low time
        )
        decision = scheduler._make_initial_scheduling_decision(
            metadata,
            context,
            500,
            200.0,  # Time exceeds limit
        )
        assert decision == SchedulingDecision.DEFER

    def test_make_initial_scheduling_decision_low_relevance(self, scheduler, sample_metadata):
        """Test scheduling decision with low phase relevance"""
        # Create metadata with low relevance
        metadata = HookMetadata(
            hook_path="/low_relev.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.LOW,
            phase_relevance={Phase.SPEC: 0.1},  # Low relevance
        )
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )
        decision = scheduler._make_initial_scheduling_decision(metadata, context, 500, 100.0)
        assert decision == SchedulingDecision.SKIP

    def test_filter_hooks_by_constraints(self, scheduler, sample_metadata):
        """Test filtering hooks by constraints"""
        # Create scheduled hooks
        hooks = []
        for path, metadata in sample_metadata.items():
            hook = ScheduledHook(
                hook_path=path,
                metadata=metadata,
                priority_score=0.7,
                estimated_cost=metadata.token_cost_estimate,
                estimated_time_ms=metadata.estimated_execution_time_ms,
                scheduling_decision=SchedulingDecision.EXECUTE,
            )
            hooks.append(hook)

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.RED,
            user_input="test",
            available_token_budget=1000,  # Limited budget
            max_execution_time_ms=500.0,  # Limited time
        )

        filtered = scheduler._filter_hooks_by_constraints(hooks, context)
        # Some hooks should be filtered or deferred
        assert len(filtered) <= len(hooks)

    def test_prioritize_hooks(self, scheduler):
        """Test hook prioritization"""
        metadata_list = [
            HookMetadata(
                hook_path=f"/hook{i}.py",
                event_type=HookEvent.SESSION_START,
                priority=HookPriority.LOW if i == 0 else HookPriority.HIGH,
            )
            for i in range(3)
        ]

        hooks = []
        for i, metadata in enumerate(metadata_list):
            hook = ScheduledHook(
                hook_path=f"/hook{i}.py",
                metadata=metadata,
                priority_score=0.3 + (i * 0.2),
                estimated_cost=100,
                estimated_time_ms=100.0,
                scheduling_decision=SchedulingDecision.EXECUTE,
            )
            hooks.append(hook)

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.RED,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        prioritized = scheduler._prioritize_hooks(hooks, context, SchedulingStrategy.PRIORITY_FIRST)
        # EXECUTE hooks should come first, sorted by priority
        execute_hooks = [h for h in prioritized if h.scheduling_decision == SchedulingDecision.EXECUTE]
        assert len(execute_hooks) > 0

    def test_resolve_dependencies(self, scheduler, sample_metadata):
        """Test dependency resolution"""
        # Create hooks with dependencies
        hooks = []
        for path, metadata in sample_metadata.items():
            hook = ScheduledHook(
                hook_path=path,
                metadata=metadata,
                priority_score=0.7,
                estimated_cost=metadata.token_cost_estimate,
                estimated_time_ms=metadata.estimated_execution_time_ms,
                scheduling_decision=SchedulingDecision.EXECUTE,
                dependencies=metadata.dependencies.copy(),
            )
            hooks.append(hook)

        resolved = scheduler._resolve_dependencies(hooks)
        # hook3 depends on hook1, so hook1 should come first
        hook3_idx = next(i for i, h in enumerate(resolved) if h.hook_path == "/hook3.py")
        hook1_idx = next(i for i, h in enumerate(resolved) if h.hook_path == "/hook1.py")
        assert hook1_idx < hook3_idx

    def test_create_execution_groups_parallel(self, scheduler, sample_metadata):
        """Test creating parallel execution groups"""
        hooks = []
        for path, metadata in sample_metadata.items():
            if metadata.parallel_safe:
                hook = ScheduledHook(
                    hook_path=path,
                    metadata=metadata,
                    priority_score=0.7,
                    estimated_cost=metadata.token_cost_estimate,
                    estimated_time_ms=metadata.estimated_execution_time_ms,
                    scheduling_decision=SchedulingDecision.EXECUTE,
                )
                hooks.append(hook)

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.RED,  # Prefers parallel
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        groups = scheduler._create_execution_groups(hooks, context, SchedulingStrategy.PHASE_OPTIMIZED)
        assert len(groups) > 0
        # At least one group should be parallel
        has_parallel = any(g.execution_type == SchedulingDecision.PARALLEL for g in groups)
        assert has_parallel

    def test_create_execution_groups_sequential(self, scheduler):
        """Test creating sequential execution groups"""
        # Create hooks that are not parallel safe
        hooks = []
        for i in range(3):
            metadata = HookMetadata(
                hook_path=f"/hook{i}.py",
                event_type=HookEvent.SESSION_START,
                priority=HookPriority.NORMAL,
                parallel_safe=False,  # Not parallel safe
            )
            hook = ScheduledHook(
                hook_path=f"/hook{i}.py",
                metadata=metadata,
                priority_score=0.7,
                estimated_cost=100,
                estimated_time_ms=100.0,
                scheduling_decision=SchedulingDecision.EXECUTE,
            )
            hooks.append(hook)

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.REFACTOR,  # Prefers sequential
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        groups = scheduler._create_execution_groups(hooks, context, SchedulingStrategy.PHASE_OPTIMIZED)
        assert len(groups) > 0
        # Groups should be sequential
        for group in groups:
            assert group.execution_type == SchedulingDecision.SEQUENTIAL

    def test_optimize_execution_order(self, scheduler):
        """Test execution order optimization"""
        # Create execution groups
        metadata = HookMetadata(
            hook_path="/test.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.HIGH,
        )
        hook = ScheduledHook(
            hook_path="/test.py",
            metadata=metadata,
            priority_score=0.9,
            estimated_cost=100,
            estimated_time_ms=100.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )

        groups = [
            ExecutionGroup(
                group_id=1,
                execution_type=SchedulingDecision.SEQUENTIAL,
                hooks=[hook],
                estimated_time_ms=100.0,
                estimated_tokens=100,
                max_wait_time_ms=100.0,
            ),
            ExecutionGroup(
                group_id=2,
                execution_type=SchedulingDecision.PARALLEL,
                hooks=[hook],
                estimated_time_ms=50.0,
                estimated_tokens=50,
                max_wait_time_ms=50.0,
            ),
        ]

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.RED,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        optimized = scheduler._optimize_execution_order(groups, context)
        # Should return groups sorted by score
        assert len(optimized) == len(groups)

    @pytest.mark.asyncio
    async def test_schedule_hooks_full_workflow(self, scheduler, sample_metadata):
        """Test complete hook scheduling workflow"""
        # Register hooks
        for path, metadata in sample_metadata.items():
            scheduler.hook_manager._hook_registry[path] = metadata
            scheduler.hook_manager._hooks_by_event[HookEvent.SESSION_START] = [path]

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.RED,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        result = await scheduler.schedule_hooks(HookEvent.SESSION_START, context)
        assert isinstance(result, SchedulingResult)
        assert isinstance(result.scheduled_hooks, list)
        assert isinstance(result.execution_plan, list)
        assert result.estimated_total_time_ms >= 0
        assert result.estimated_total_tokens >= 0
        assert result.scheduling_strategy in SchedulingStrategy

    def test_get_scheduling_statistics(self, scheduler):
        """Test getting scheduling statistics"""
        # Perform some scheduling operations
        stats = scheduler.get_scheduling_statistics()
        assert "total_schedules" in stats
        assert "strategy_performance" in stats
        assert "recent_performance" in stats

    def test_get_phase_optimization_insights(self, scheduler):
        """Test getting phase optimization insights"""
        insights = scheduler.get_phase_optimization_insights(Phase.RED)
        assert "phase" in insights
        assert "parameters" in insights
        assert "historical_schedules" in insights
        assert "optimization_recommendations" in insights

    def test_update_scheduling_history(self, scheduler):
        """Test scheduling history update"""
        result = SchedulingResult(
            scheduled_hooks=[],
            execution_plan=[],
            estimated_total_time_ms=100.0,
            estimated_total_tokens=50,
            skipped_hooks=[],
            deferred_hooks=[],
            scheduling_strategy=SchedulingStrategy.PHASE_OPTIMIZED,
        )
        scheduler._update_scheduling_history(result, 0.05)
        assert len(scheduler._scheduling_history) == 1

        entry = scheduler._scheduling_history[0]
        assert "timestamp" in entry
        assert "strategy" in entry
        assert "planning_time_ms" in entry

    def test_update_strategy_performance(self, scheduler):
        """Test strategy performance update"""
        result = SchedulingResult(
            scheduled_hooks=[],
            execution_plan=[],
            estimated_total_time_ms=100.0,
            estimated_total_tokens=50,
            skipped_hooks=[],
            deferred_hooks=[],
            scheduling_strategy=SchedulingStrategy.PRIORITY_FIRST,
        )
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.RED,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )
        scheduler._update_strategy_performance(SchedulingStrategy.PRIORITY_FIRST, result, context)
        # Should update performance metrics
        perf = scheduler._strategy_performance[SchedulingStrategy.PRIORITY_FIRST]
        assert perf["usage_count"] > 0

    def test_phase_parameters_for_all_phases(self, scheduler):
        """Test phase parameters exist for all phases"""
        for phase in Phase:
            params = scheduler._phase_parameters.get(phase)
            assert params is not None
            assert "max_total_time_ms" in params
            assert "token_budget_ratio" in params
            assert "priority_weights" in params


# =============================================================================
# EDGE CASES AND ERROR CONDITIONS
# =============================================================================


class TestEdgeCases:
    """Test edge cases and error conditions"""

    @pytest.fixture
    def scheduler(self):
        """Create scheduler instance"""
        return PhaseOptimizedHookScheduler()

    def test_empty_hook_list(self, scheduler):
        """Test scheduling with no hooks"""
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.RED,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        # Mock empty hook list
        scheduler.hook_manager._hooks_by_event[HookEvent.SESSION_START] = []

        async def run_test():
            result = await scheduler.schedule_hooks(HookEvent.SESSION_START, context)
            assert len(result.scheduled_hooks) == 0
            assert len(result.execution_plan) == 0

        asyncio.run(run_test())

    def test_all_hooks_skipped(self, scheduler):
        """Test when all hooks are skipped due to constraints"""
        # Create hooks with high costs
        hooks = []
        for i in range(3):
            metadata = HookMetadata(
                hook_path=f"/expensive_hook{i}.py",
                event_type=HookEvent.SESSION_START,
                priority=HookPriority.NORMAL,
                token_cost_estimate=100000,  # Very high
            )
            hook = ScheduledHook(
                hook_path=f"/expensive_hook{i}.py",
                metadata=metadata,
                priority_score=0.5,
                estimated_cost=100000,
                estimated_time_ms=100.0,
                scheduling_decision=SchedulingDecision.EXECUTE,
            )
            hooks.append(hook)

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.RED,
            user_input="test",
            available_token_budget=100,  # Very low budget
            max_execution_time_ms=1000.0,
        )

        filtered = scheduler._filter_hooks_by_constraints(hooks, context)
        # All should be deferred or skipped
        assert len(filtered) == len(hooks)

    def test_circular_dependencies(self, scheduler):
        """Test handling of circular dependencies"""
        # Create hooks with circular dependencies
        hooks = []
        for i in range(3):
            metadata = HookMetadata(
                hook_path=f"/hook{i}.py",
                event_type=HookEvent.SESSION_START,
                priority=HookPriority.NORMAL,
            )
            # Circular: 0->1, 1->2, 2->0
            next_hook = (i + 1) % 3
            hook = ScheduledHook(
                hook_path=f"/hook{i}.py",
                metadata=metadata,
                priority_score=0.7,
                estimated_cost=100,
                estimated_time_ms=100.0,
                scheduling_decision=SchedulingDecision.EXECUTE,
                dependencies={f"/hook{next_hook}.py"},
            )
            hooks.append(hook)

        # Should resolve circular dependency by priority
        resolved = scheduler._resolve_dependencies(hooks)
        assert len(resolved) == len(hooks)

    def test_strategy_performance_initialization(self, scheduler):
        """Test strategy performance initialization"""
        for strategy in SchedulingStrategy:
            perf = scheduler._strategy_performance[strategy]
            assert "success_rate" in perf
            assert "avg_efficiency" in perf
            assert "usage_count" in perf
            assert perf["success_rate"] == 1.0
            assert perf["avg_efficiency"] == 0.8
            assert perf["usage_count"] == 0


# =============================================================================
# CONVENIENCE FUNCTION TESTS
# =============================================================================


class TestConvenienceFunctions:
    """Test module-level convenience functions"""

    @pytest.mark.asyncio
    async def test_schedule_session_start_hooks(self):
        """Test schedule_session_start_hooks convenience function"""
        from moai_adk.core.phase_optimized_hook_scheduler import (
            schedule_session_start_hooks,
        )

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        result = await schedule_session_start_hooks(context)
        assert isinstance(result, SchedulingResult)

    @pytest.mark.asyncio
    async def test_schedule_pre_tool_hooks(self):
        """Test schedule_pre_tool_hooks convenience function"""
        from moai_adk.core.phase_optimized_hook_scheduler import schedule_pre_tool_hooks

        context = HookSchedulingContext(
            event_type=HookEvent.PRE_TOOL_USE,
            current_phase=Phase.RED,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        result = await schedule_pre_tool_hooks(context)
        assert isinstance(result, SchedulingResult)

    def test_get_hook_scheduling_insights(self):
        """Test get_hook_scheduling_insights convenience function"""
        from moai_adk.core.phase_optimized_hook_scheduler import (
            get_hook_scheduling_insights,
        )

        insights = get_hook_scheduling_insights(Phase.GREEN)
        assert "phase" in insights
        assert "parameters" in insights


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
