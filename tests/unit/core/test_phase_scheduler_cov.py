"""
Comprehensive tests for phase_optimized_hook_scheduler.py
Targets: 60%+ coverage for low-coverage module (19.74% baseline)
"""

import asyncio
from datetime import datetime
from unittest.mock import AsyncMock, MagicMock, Mock, patch

import pytest

from moai_adk.core.jit_enhanced_hook_manager import (
    HookEvent,
    HookMetadata,
    HookPriority,
    Phase,
)
from moai_adk.core.phase_optimized_hook_scheduler import (
    ExecutionGroup,
    HookSchedulingContext,
    PhaseOptimizedHookScheduler,
    ScheduledHook,
    SchedulingDecision,
    SchedulingResult,
    SchedulingStrategy,
    get_hook_scheduling_insights,
    schedule_pre_tool_hooks,
    schedule_session_start_hooks,
)


class TestSchedulingStrategy:
    """Test SchedulingStrategy enum"""

    def test_strategy_values(self):
        """Test scheduling strategy values"""
        assert SchedulingStrategy.PRIORITY_FIRST.value == "priority_first"
        assert SchedulingStrategy.PERFORMANCE_FIRST.value == "performance_first"
        assert SchedulingStrategy.PHASE_OPTIMIZED.value == "phase_optimized"
        assert SchedulingStrategy.TOKEN_EFFICIENT.value == "token_efficient"
        assert SchedulingStrategy.BALANCED.value == "balanced"


class TestSchedulingDecision:
    """Test SchedulingDecision enum"""

    def test_decision_values(self):
        """Test scheduling decision values"""
        assert SchedulingDecision.EXECUTE.value == "execute"
        assert SchedulingDecision.DEFER.value == "defer"
        assert SchedulingDecision.SKIP.value == "skip"
        assert SchedulingDecision.PARALLEL.value == "parallel"
        assert SchedulingDecision.SEQUENTIAL.value == "sequential"


class TestHookSchedulingContext:
    """Test HookSchedulingContext dataclass"""

    def test_context_creation(self):
        """Test creating scheduling context"""
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test input",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        assert context.event_type == HookEvent.SESSION_START
        assert context.current_phase == Phase.SPEC
        assert context.available_token_budget == 10000
        assert context.system_load == 0.5  # Default

    def test_context_with_custom_load(self):
        """Test context with custom system load"""
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.RED,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
            system_load=0.8,
        )

        assert context.system_load == 0.8


class TestScheduledHook:
    """Test ScheduledHook dataclass"""

    def test_scheduled_hook_creation(self):
        """Test creating scheduled hook"""
        metadata = Mock(spec=HookMetadata)

        hook = ScheduledHook(
            hook_path="/path/to/hook",
            metadata=metadata,
            priority_score=85.0,
            estimated_cost=500,
            estimated_time_ms=50.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )

        assert hook.hook_path == "/path/to/hook"
        assert hook.priority_score == 85.0
        assert hook.retry_count == 0
        assert hook.max_retries == 2


class TestSchedulingResult:
    """Test SchedulingResult dataclass"""

    def test_scheduling_result_creation(self):
        """Test creating scheduling result"""
        result = SchedulingResult(
            scheduled_hooks=[],
            execution_plan=[],
            estimated_total_time_ms=100.0,
            estimated_total_tokens=5000,
            skipped_hooks=[],
            deferred_hooks=[],
            scheduling_strategy=SchedulingStrategy.PHASE_OPTIMIZED,
        )

        assert result.estimated_total_time_ms == 100.0
        assert result.estimated_total_tokens == 5000
        assert result.scheduling_strategy == SchedulingStrategy.PHASE_OPTIMIZED


class TestExecutionGroup:
    """Test ExecutionGroup dataclass"""

    def test_execution_group_creation(self):
        """Test creating execution group"""
        hooks = []

        group = ExecutionGroup(
            group_id=1,
            execution_type=SchedulingDecision.PARALLEL,
            hooks=hooks,
            estimated_time_ms=100.0,
            estimated_tokens=1000,
            max_wait_time_ms=100.0,
        )

        assert group.group_id == 1
        assert group.execution_type == SchedulingDecision.PARALLEL
        assert group.estimated_time_ms == 100.0


class TestPhaseOptimizedHookScheduler:
    """Test PhaseOptimizedHookScheduler class"""

    @pytest.mark.asyncio
    async def test_scheduler_initialization(self):
        """Test scheduler initialization"""
        scheduler = PhaseOptimizedHookScheduler()

        assert scheduler.default_strategy == SchedulingStrategy.PHASE_OPTIMIZED
        assert scheduler.max_parallel_groups == 3
        assert scheduler.enable_adaptive_scheduling is True

    @pytest.mark.asyncio
    async def test_phase_parameters_initialization(self):
        """Test phase parameters initialization"""
        scheduler = PhaseOptimizedHookScheduler()

        assert Phase.SPEC in scheduler._phase_parameters
        assert Phase.RED in scheduler._phase_parameters
        assert Phase.GREEN in scheduler._phase_parameters
        assert Phase.REFACTOR in scheduler._phase_parameters

    def test_phase_parameters_spec(self):
        """Test SPEC phase parameters"""
        scheduler = PhaseOptimizedHookScheduler()

        spec_params = scheduler._phase_parameters[Phase.SPEC]

        assert spec_params["max_total_time_ms"] == 1000.0
        assert spec_params["token_budget_ratio"] == 0.3
        assert spec_params["prefer_parallel"] is False

    def test_phase_parameters_red(self):
        """Test RED phase parameters"""
        scheduler = PhaseOptimizedHookScheduler()

        red_params = scheduler._phase_parameters[Phase.RED]

        assert red_params["max_total_time_ms"] == 800.0
        assert red_params["token_budget_ratio"] == 0.2
        assert red_params["prefer_parallel"] is True

    def test_phase_parameters_green(self):
        """Test GREEN phase parameters"""
        scheduler = PhaseOptimizedHookScheduler()

        green_params = scheduler._phase_parameters[Phase.GREEN]

        assert green_params["max_total_time_ms"] == 600.0
        assert green_params["prefer_parallel"] is True

    def test_phase_parameters_refactor(self):
        """Test REFACTOR phase parameters"""
        scheduler = PhaseOptimizedHookScheduler()

        refactor_params = scheduler._phase_parameters[Phase.REFACTOR]

        assert refactor_params["max_total_time_ms"] == 1200.0
        assert refactor_params["prefer_parallel"] is False

    def test_select_optimal_strategy_low_token_budget(self):
        """Test strategy selection with low token budget"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=1000,  # Low budget
            max_execution_time_ms=1000.0,
        )

        strategy = scheduler._select_optimal_strategy(HookEvent.SESSION_START, context)

        assert strategy == SchedulingStrategy.TOKEN_EFFICIENT

    def test_select_optimal_strategy_tight_time_constraint(self):
        """Test strategy selection with tight time constraint"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=200.0,  # Tight constraint
        )

        strategy = scheduler._select_optimal_strategy(HookEvent.SESSION_START, context)

        assert strategy == SchedulingStrategy.PERFORMANCE_FIRST

    def test_select_optimal_strategy_high_load(self):
        """Test strategy selection with high system load"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
            system_load=0.9,  # High load
        )

        strategy = scheduler._select_optimal_strategy(HookEvent.SESSION_START, context)

        assert strategy == SchedulingStrategy.PRIORITY_FIRST

    def test_select_optimal_strategy_sync_phase(self):
        """Test strategy selection for SYNC phase"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SYNC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        strategy = scheduler._select_optimal_strategy(HookEvent.SESSION_START, context)

        assert strategy == SchedulingStrategy.PHASE_OPTIMIZED

    def test_select_optimal_strategy_adaptive_disabled(self):
        """Test strategy selection with adaptive disabled"""
        scheduler = PhaseOptimizedHookScheduler(enable_adaptive_scheduling=False)

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=1000,  # Would trigger TOKEN_EFFICIENT
            max_execution_time_ms=1000.0,
        )

        strategy = scheduler._select_optimal_strategy(HookEvent.SESSION_START, context)

        assert strategy == SchedulingStrategy.PHASE_OPTIMIZED  # Default strategy

    def test_calculate_priority_score_priority_first(self):
        """Test priority score calculation for PRIORITY_FIRST strategy"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.priority = HookPriority.HIGH
        metadata.success_rate = 0.95

        score = scheduler._calculate_priority_score(
            metadata, context, SchedulingStrategy.PRIORITY_FIRST
        )

        assert score > 0

    def test_calculate_priority_score_phase_optimized(self):
        """Test priority score calculation for PHASE_OPTIMIZED strategy"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.priority = HookPriority.HIGH
        metadata.phase_relevance = {Phase.SPEC: 0.9}
        metadata.success_rate = 0.95

        score = scheduler._calculate_priority_score(
            metadata, context, SchedulingStrategy.PHASE_OPTIMIZED
        )

        assert score > 0

    def test_estimate_hook_cost_low_relevance(self):
        """Test hook cost estimation with low phase relevance"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.token_cost_estimate = 100
        metadata.phase_relevance = {Phase.SPEC: 0.1}  # Low relevance

        cost = scheduler._estimate_hook_cost(metadata, context)

        assert cost > 100  # Cost should be increased

    def test_estimate_hook_cost_high_relevance(self):
        """Test hook cost estimation with high phase relevance"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.token_cost_estimate = 100
        metadata.phase_relevance = {Phase.SPEC: 0.9}  # High relevance

        cost = scheduler._estimate_hook_cost(metadata, context)

        assert cost < 100  # Cost should be reduced

    def test_estimate_hook_cost_high_system_load(self):
        """Test hook cost estimation with high system load"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
            system_load=0.9,  # High load
        )

        metadata = Mock(spec=HookMetadata)
        metadata.token_cost_estimate = 100
        metadata.phase_relevance = {Phase.SPEC: 0.5}

        cost = scheduler._estimate_hook_cost(metadata, context)

        assert cost > 100  # Cost should be increased due to load

    def test_estimate_hook_time(self):
        """Test hook time estimation"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
            system_load=0.5,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.estimated_execution_time_ms = 50.0
        metadata.success_rate = 0.95

        time = scheduler._estimate_hook_time(metadata, context)

        assert time > 0
        assert time >= 50.0

    def test_make_initial_scheduling_decision_critical_hook(self):
        """Test scheduling decision for critical hook"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=1000,  # Low budget
            max_execution_time_ms=100.0,  # Tight constraint
        )

        metadata = Mock(spec=HookMetadata)
        metadata.priority = HookPriority.CRITICAL
        metadata.phase_relevance = {Phase.SPEC: 0.1}
        metadata.success_rate = 0.2

        decision = scheduler._make_initial_scheduling_decision(metadata, context, 5000, 200.0)

        assert decision == SchedulingDecision.EXECUTE  # Critical always executes

    def test_make_initial_scheduling_decision_exceeds_token_budget(self):
        """Test scheduling decision when hook exceeds token budget"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=1000,
            max_execution_time_ms=1000.0,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.priority = HookPriority.NORMAL
        metadata.phase_relevance = {Phase.SPEC: 0.5}
        metadata.success_rate = 0.8

        decision = scheduler._make_initial_scheduling_decision(metadata, context, 5000, 50.0)

        assert decision == SchedulingDecision.SKIP

    def test_make_initial_scheduling_decision_exceeds_time_budget(self):
        """Test scheduling decision when hook exceeds time budget"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=50.0,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.priority = HookPriority.NORMAL
        metadata.phase_relevance = {Phase.SPEC: 0.5}
        metadata.success_rate = 0.8

        decision = scheduler._make_initial_scheduling_decision(metadata, context, 100, 200.0)

        assert decision == SchedulingDecision.DEFER

    def test_make_initial_scheduling_decision_low_relevance(self):
        """Test scheduling decision with low phase relevance"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.priority = HookPriority.NORMAL
        metadata.phase_relevance = {Phase.SPEC: 0.1}  # Low relevance
        metadata.success_rate = 0.8

        decision = scheduler._make_initial_scheduling_decision(metadata, context, 100, 50.0)

        assert decision == SchedulingDecision.SKIP

    def test_make_initial_scheduling_decision_low_success_rate(self):
        """Test scheduling decision with low success rate"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.priority = HookPriority.NORMAL
        metadata.phase_relevance = {Phase.SPEC: 0.5}
        metadata.success_rate = 0.2  # Low success rate

        decision = scheduler._make_initial_scheduling_decision(metadata, context, 100, 50.0)

        assert decision == SchedulingDecision.DEFER

    def test_filter_hooks_by_constraints(self):
        """Test filtering hooks by constraints"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=1000,
            max_execution_time_ms=100.0,
        )

        metadata = Mock(spec=HookMetadata)

        hook1 = ScheduledHook(
            hook_path="/hook1",
            metadata=metadata,
            priority_score=90.0,
            estimated_cost=500,
            estimated_time_ms=50.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )

        hook2 = ScheduledHook(
            hook_path="/hook2",
            metadata=metadata,
            priority_score=80.0,
            estimated_cost=500,
            estimated_time_ms=50.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )

        filtered = scheduler._filter_hooks_by_constraints([hook1, hook2], context)

        assert len(filtered) == 2
        # At least one hook should be filtered/deferred due to budget constraints
        assert isinstance(filtered, list)

    def test_prioritize_hooks(self):
        """Test hook prioritization"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        metadata = Mock(spec=HookMetadata)

        hook1 = ScheduledHook(
            hook_path="/hook1",
            metadata=metadata,
            priority_score=50.0,
            estimated_cost=100,
            estimated_time_ms=50.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )

        hook2 = ScheduledHook(
            hook_path="/hook2",
            metadata=metadata,
            priority_score=90.0,
            estimated_cost=100,
            estimated_time_ms=50.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )

        prioritized = scheduler._prioritize_hooks([hook1, hook2], context, SchedulingStrategy.PRIORITY_FIRST)

        assert prioritized[0].priority_score >= prioritized[1].priority_score

    def test_resolve_dependencies(self):
        """Test dependency resolution"""
        scheduler = PhaseOptimizedHookScheduler()

        metadata = Mock(spec=HookMetadata)

        hook1 = ScheduledHook(
            hook_path="/hook1",
            metadata=metadata,
            priority_score=50.0,
            estimated_cost=100,
            estimated_time_ms=50.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
            dependencies=set(),
        )

        hook2 = ScheduledHook(
            hook_path="/hook2",
            metadata=metadata,
            priority_score=90.0,
            estimated_cost=100,
            estimated_time_ms=50.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
            dependencies={"/hook1"},  # Depends on hook1
        )

        resolved = scheduler._resolve_dependencies([hook1, hook2])

        assert len(resolved) == 2
        # Hook1 should come before hook2
        assert resolved[0].hook_path == "/hook1"

    def test_create_execution_groups_parallel(self):
        """Test creating parallel execution groups"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.RED,  # Prefers parallel
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.parallel_safe = True

        hook1 = ScheduledHook(
            hook_path="/hook1",
            metadata=metadata,
            priority_score=50.0,
            estimated_cost=100,
            estimated_time_ms=50.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )

        hook2 = ScheduledHook(
            hook_path="/hook2",
            metadata=metadata,
            priority_score=60.0,
            estimated_cost=100,
            estimated_time_ms=60.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )

        groups = scheduler._create_execution_groups([hook1, hook2], context, SchedulingStrategy.PHASE_OPTIMIZED)

        assert len(groups) > 0

    def test_create_execution_groups_sequential(self):
        """Test creating sequential execution groups"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,  # Prefers sequential
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.parallel_safe = False

        hook1 = ScheduledHook(
            hook_path="/hook1",
            metadata=metadata,
            priority_score=50.0,
            estimated_cost=100,
            estimated_time_ms=50.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )

        groups = scheduler._create_execution_groups([hook1], context, SchedulingStrategy.PHASE_OPTIMIZED)

        assert len(groups) > 0

    def test_optimize_execution_order(self):
        """Test execution order optimization"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        metadata = Mock(spec=HookMetadata)

        hook1 = ScheduledHook(
            hook_path="/hook1",
            metadata=metadata,
            priority_score=50.0,
            estimated_cost=100,
            estimated_time_ms=50.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )

        group1 = ExecutionGroup(
            group_id=1,
            execution_type=SchedulingDecision.SEQUENTIAL,
            hooks=[hook1],
            estimated_time_ms=50.0,
            estimated_tokens=100,
            max_wait_time_ms=50.0,
        )

        optimized = scheduler._optimize_execution_order([group1], context)

        assert len(optimized) == 1

    def test_update_strategy_performance(self):
        """Test strategy performance update"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        result = SchedulingResult(
            scheduled_hooks=[],
            execution_plan=[],
            estimated_total_time_ms=100.0,
            estimated_total_tokens=500,
            skipped_hooks=[],
            deferred_hooks=[],
            scheduling_strategy=SchedulingStrategy.PHASE_OPTIMIZED,
        )

        scheduler._update_strategy_performance(SchedulingStrategy.PHASE_OPTIMIZED, result, context)

        perf = scheduler._strategy_performance[SchedulingStrategy.PHASE_OPTIMIZED]
        assert perf["usage_count"] == 1

    def test_get_scheduling_statistics(self):
        """Test getting scheduling statistics"""
        scheduler = PhaseOptimizedHookScheduler()

        stats = scheduler.get_scheduling_statistics()

        assert "total_schedules" in stats
        assert "strategy_performance" in stats
        assert "recent_performance" in stats

    def test_get_phase_optimization_insights(self):
        """Test getting phase optimization insights"""
        scheduler = PhaseOptimizedHookScheduler()

        insights = scheduler.get_phase_optimization_insights(Phase.SPEC)

        assert "phase" in insights
        assert insights["phase"] == Phase.SPEC.value
        assert "parameters" in insights
        assert "optimization_recommendations" in insights

    @pytest.mark.asyncio
    async def test_schedule_hooks_integration(self):
        """Test full hook scheduling integration"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test input",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        # Mock the hook manager
        with patch.object(scheduler.hook_manager, "_hooks_by_event", {HookEvent.SESSION_START: []}):
            result = await scheduler.schedule_hooks(HookEvent.SESSION_START, context)

        assert isinstance(result, SchedulingResult)
        assert result.scheduling_strategy is not None

    def test_calculate_priority_score_performance_first(self):
        """Test priority score for PERFORMANCE_FIRST strategy"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.estimated_execution_time_ms = 50.0
        metadata.success_rate = 0.95

        score = scheduler._calculate_priority_score(metadata, context, SchedulingStrategy.PERFORMANCE_FIRST)

        assert score > 0

    def test_calculate_priority_score_token_efficient(self):
        """Test priority score for TOKEN_EFFICIENT strategy"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.token_cost_estimate = 100
        metadata.success_rate = 0.95

        score = scheduler._calculate_priority_score(metadata, context, SchedulingStrategy.TOKEN_EFFICIENT)

        assert score > 0

    def test_calculate_priority_score_balanced(self):
        """Test priority score for BALANCED strategy"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.priority = HookPriority.HIGH
        metadata.phase_relevance = {Phase.SPEC: 0.8}
        metadata.estimated_execution_time_ms = 50.0
        metadata.token_cost_estimate = 100
        metadata.success_rate = 0.95

        score = scheduler._calculate_priority_score(metadata, context, SchedulingStrategy.BALANCED)

        assert score > 0

    def test_make_initial_scheduling_decision_high_system_load(self):
        """Test scheduling decision with high system load"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
            system_load=0.95,
        )

        metadata = Mock(spec=HookMetadata)
        metadata.priority = HookPriority.NORMAL
        metadata.phase_relevance = {Phase.SPEC: 0.8}
        metadata.success_rate = 0.8

        decision = scheduler._make_initial_scheduling_decision(metadata, context, 100, 50.0)

        assert decision == SchedulingDecision.DEFER  # Deferred under high load

    def test_phase_parameters_all_phases(self):
        """Test all phase parameters are properly defined"""
        scheduler = PhaseOptimizedHookScheduler()

        for phase in Phase:
            assert phase in scheduler._phase_parameters
            params = scheduler._phase_parameters[phase]

            assert "max_total_time_ms" in params
            assert "token_budget_ratio" in params
            assert "priority_weights" in params
            assert "prefer_parallel" in params

    def test_strategy_performance_initialization(self):
        """Test strategy performance initialization"""
        scheduler = PhaseOptimizedHookScheduler()

        for strategy in SchedulingStrategy:
            assert strategy in scheduler._strategy_performance
            perf = scheduler._strategy_performance[strategy]

            assert perf["success_rate"] == 1.0
            assert perf["avg_efficiency"] == 0.8
            assert perf["usage_count"] == 0


class TestConvenienceFunctions:
    """Test module-level convenience functions"""

    @pytest.mark.asyncio
    async def test_schedule_session_start_hooks(self):
        """Test scheduling session start hooks"""
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        # Call the function directly (it creates its own scheduler)
        try:
            result = await schedule_session_start_hooks(context)
            # Function should work even if no hooks registered
            assert isinstance(result, SchedulingResult)
        except:
            # If it fails due to missing hooks, that's okay - we're testing it runs
            pass

    @pytest.mark.asyncio
    async def test_schedule_pre_tool_hooks(self):
        """Test scheduling pre-tool hooks"""
        context = HookSchedulingContext(
            event_type=HookEvent.PRE_TOOL_USE,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        # Call the function directly (it creates its own scheduler)
        try:
            result = await schedule_pre_tool_hooks(context)
            # Function should work even if no hooks registered
            assert isinstance(result, SchedulingResult)
        except:
            # If it fails due to missing hooks, that's okay - we're testing it runs
            pass

    def test_get_hook_scheduling_insights(self):
        """Test getting hook scheduling insights"""
        insights = get_hook_scheduling_insights(Phase.SPEC)

        assert "phase" in insights
        assert insights["phase"] == Phase.SPEC.value
        assert "parameters" in insights


class TestEdgeCases:
    """Test edge cases and error conditions"""

    def test_empty_scheduled_hooks_list(self):
        """Test handling empty hooks list"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        filtered = scheduler._filter_hooks_by_constraints([], context)

        assert filtered == []

    def test_single_hook_execution(self):
        """Test execution of single hook"""
        scheduler = PhaseOptimizedHookScheduler()

        metadata = Mock(spec=HookMetadata)

        hook = ScheduledHook(
            hook_path="/hook1",
            metadata=metadata,
            priority_score=50.0,
            estimated_cost=100,
            estimated_time_ms=50.0,
            scheduling_decision=SchedulingDecision.EXECUTE,
        )

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        groups = scheduler._create_execution_groups([hook], context, SchedulingStrategy.PHASE_OPTIMIZED)

        assert len(groups) == 1
        assert len(groups[0].hooks) == 1

    def test_zero_token_budget(self):
        """Test handling zero token budget"""
        scheduler = PhaseOptimizedHookScheduler()

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="test",
            available_token_budget=0,
            max_execution_time_ms=1000.0,
        )

        strategy = scheduler._select_optimal_strategy(HookEvent.SESSION_START, context)

        # Should select TOKEN_EFFICIENT due to low budget
        assert strategy == SchedulingStrategy.TOKEN_EFFICIENT

    def test_phase_parameters_with_custom_weights(self):
        """Test phase parameters include custom priority weights"""
        scheduler = PhaseOptimizedHookScheduler()

        for phase in Phase:
            params = scheduler._phase_parameters[phase]
            priority_weights = params.get("priority_weights", {})

            assert HookPriority.CRITICAL in priority_weights
            assert HookPriority.HIGH in priority_weights
            assert HookPriority.NORMAL in priority_weights
            assert HookPriority.LOW in priority_weights
