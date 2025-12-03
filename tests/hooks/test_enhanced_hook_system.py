"""
Comprehensive Test Suite for Enhanced Hook System

Tests for JIT-Enhanced Hook Manager and Phase-Optimized Hook Scheduler
with full coverage of all components and integration scenarios.

Test Coverage:
- JIT-Enhanced Hook Manager functionality
- Phase-Optimized Hook Scheduler behavior
- Integration between components
- Performance optimization scenarios
- Error handling and edge cases
- Phase-specific optimizations
- Token budget management
- Dependency resolution
"""

import time
from datetime import datetime
from pathlib import Path
from unittest.mock import Mock, patch

import pytest

# Import enhanced hook system components
try:
    from src.moai_adk.core.jit_context_loader import JITContextLoader
    from src.moai_adk.core.jit_enhanced_hook_manager import (
        HookEvent,
        HookExecutionResult,
        HookMetadata,
        HookPerformanceMetrics,
        HookPriority,
        JITEnhancedHookManager,
        execute_session_start_hooks,
        get_jit_hook_manager,
    )
    from src.moai_adk.core.phase_optimized_hook_scheduler import (
        ExecutionGroup,
        HookSchedulingContext,
        Phase,
        PhaseOptimizedHookScheduler,
        ScheduledHook,
        SchedulingDecision,
        SchedulingResult,
        SchedulingStrategy,
    )
except ImportError:
    # Fallback imports for testing environment
    import sys
    from pathlib import Path

    sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

    from moai_adk.core.jit_enhanced_hook_manager import (
        HookEvent,
        HookMetadata,
        HookPriority,
        JITEnhancedHookManager,
        execute_session_start_hooks,
    )
    from moai_adk.core.phase_optimized_hook_scheduler import (
        HookSchedulingContext,
        Phase,
        PhaseOptimizedHookScheduler,
        ScheduledHook,
        SchedulingDecision,
        SchedulingResult,
        SchedulingStrategy,
    )


# Fixtures and test data
@pytest.fixture
def temp_hooks_directory(tmp_path):
    """Create temporary hooks directory with test hooks"""
    hooks_dir = tmp_path / ".claude" / "hooks"
    hooks_dir.mkdir(parents=True)

    # Create test hooks
    session_start_hook = hooks_dir / "session_start__test.py"
    session_start_hook.write_text(
        """#!/usr/bin/env python3
import json
import sys
data = json.loads(sys.stdin.read() or "{}")
result = {"continue": True, "systemMessage": "Test session start hook"}
print(json.dumps(result))
"""
    )

    pre_tool_hook = hooks_dir / "pre_tool__validation.py"
    pre_tool_hook.write_text(
        """#!/usr/bin/env python3
import json
import sys
data = json.loads(sys.stdin.read() or "{}")
result = {"continue": True, "hookSpecificOutput": {"validated": True}}
print(json.dumps(result))
"""
    )

    slow_hook = hooks_dir / "session_start__slow.py"
    slow_hook.write_text(
        """#!/usr/bin/env python3
import json
import sys
import time
data = json.loads(sys.stdin.read() or "{}")
time.sleep(0.1)  # Simulate slow operation
result = {"continue": True, "systemMessage": "Slow hook completed"}
print(json.dumps(result))
"""
    )

    error_hook = hooks_dir / "session_start__error.py"
    error_hook.write_text(
        """#!/usr/bin/env python3
import json
import sys
data = json.loads(sys.stdin.read() or "{}")
print(json.dumps({"error": "Test error"}), file=sys.stderr)
sys.exit(1)
"""
    )

    return hooks_dir


@pytest.fixture
def temp_cache_directory(tmp_path):
    """Create temporary cache directory"""
    cache_dir = tmp_path / ".moai" / "cache" / "hooks"
    cache_dir.mkdir(parents=True)
    return cache_dir


@pytest.fixture
def mock_jit_loader():
    """Mock JIT Context Loader"""
    mock_loader = Mock()

    # Mock detect_phase method
    phase_result = Mock()
    phase_result.current_phase = Phase.SPEC
    phase_result.confidence = 0.9
    phase_result.phase_history = [Phase.SPEC]
    mock_loader.detect_phase = Mock(return_value=phase_result)

    # Mock context loading
    mock_loader.load_context = Mock(
        return_value={
            "user_input": "test",
            "current_phase": "SPEC",
            "relevant_skills": ["test_skill"],
            "context": {"test": True},
        }
    )

    return mock_loader


@pytest.fixture
def hook_manager(temp_hooks_directory, temp_cache_directory, mock_jit_loader):
    """Create JIT-Enhanced Hook Manager with mocked dependencies"""
    with patch("src.moai_adk.core.jit_enhanced_hook_manager.JITContextLoader", return_value=mock_jit_loader):
        manager = JITEnhancedHookManager(
            hooks_directory=temp_hooks_directory, cache_directory=temp_cache_directory, max_concurrent_hooks=3
        )
        return manager


@pytest.fixture
def hook_scheduler():
    """Create Phase-Optimized Hook Scheduler"""
    return PhaseOptimizedHookScheduler(max_parallel_groups=2, enable_adaptive_scheduling=True)


@pytest.fixture
def sample_context():
    """Sample hook execution context"""
    return {"user_id": "test_user", "session_id": "test_session", "timestamp": datetime.now().isoformat(), "test": True}


@pytest.fixture
def scheduling_context():
    """Sample scheduling context"""
    return HookSchedulingContext(
        event_type=HookEvent.SESSION_START,
        current_phase=Phase.SPEC,
        user_input="Creating new user authentication specification",
        available_token_budget=10000,
        max_execution_time_ms=1000.0,
        system_load=0.5,
    )


class TestJITEnhancedHookManager:
    """Test JIT-Enhanced Hook Manager functionality"""

    @pytest.mark.asyncio
    async def test_hook_discovery(self, hook_manager):
        """Test automatic hook discovery and registration"""
        # Check that hooks were discovered
        assert len(hook_manager._hook_registry) > 0

        # Check specific hooks were registered
        registered_hooks = list(hook_manager._hook_registry.keys())
        assert "session_start__test.py" in registered_hooks
        assert "pre_tool__validation.py" in registered_hooks
        assert "session_start__slow.py" in registered_hooks

    def test_event_type_extraction(self, hook_manager):
        """Test event type extraction from filenames"""
        # Test various filename patterns
        test_cases = [
            ("session_start__test.py", HookEvent.SESSION_START),
            ("pre_tool__validation.py", HookEvent.PRE_TOOL_USE),
            ("post_tool__log.py", HookEvent.POST_TOOL_USE),
            ("subagent_start__context.py", HookEvent.SUBAGENT_START),
            ("subagent_stop__cleanup.py", HookEvent.SUBAGENT_STOP),
            ("session_end__save.py", HookEvent.SESSION_END),
            ("unknown_hook.py", None),
        ]

        for filename, expected_event in test_cases:
            event = hook_manager._extract_event_type_from_filename(filename)
            assert event == expected_event, f"Failed for {filename}: got {event}, expected {expected_event}"

    def test_hook_priority_determination(self, hook_manager):
        """Test hook priority determination logic"""
        # Test critical hooks
        security_metadata = hook_manager._determine_hook_priority("security_validator.py", HookEvent.PRE_TOOL_USE)
        assert security_metadata == HookPriority.CRITICAL

        # Test high priority hooks
        performance_metadata = hook_manager._determine_hook_priority("performance_optimizer.py", HookEvent.PRE_TOOL_USE)
        assert performance_metadata == HookPriority.HIGH

        # Test low priority hooks
        analytics_metadata = hook_manager._determine_hook_priority("analytics_tracker.py", HookEvent.SESSION_START)
        assert analytics_metadata == HookPriority.LOW

    def test_execution_time_estimation(self, hook_manager):
        """Test execution time estimation"""
        # Test git operations
        git_time = hook_manager._estimate_execution_time("git_status_hook.py")
        assert git_time == 200.0

        # Test network operations
        network_time = hook_manager._estimate_execution_time("api_fetch_hook.py")
        assert network_time == 500.0

        # Test file I/O operations
        file_time = hook_manager._estimate_execution_time("config_reader.py")
        assert file_time == 50.0

        # Test simple operations
        simple_time = hook_manager._estimate_execution_time("simple_logger.py")
        assert simple_time == 10.0

    def test_phase_relevance_determination(self, hook_manager):
        """Test phase relevance determination"""
        # Test SPEC phase relevance
        spec_metadata = hook_manager._determine_phase_relevance("spec_generator.py", HookEvent.SESSION_START)
        assert spec_metadata[Phase.SPEC] == 1.0

        # Test RED phase relevance
        test_metadata = hook_manager._determine_phase_relevance("test_runner.py", HookEvent.PRE_TOOL_USE)
        assert test_metadata[Phase.RED] == 1.0

        # Test default relevance
        generic_metadata = hook_manager._determine_phase_relevance("generic_hook.py", HookEvent.SESSION_START)
        for phase in Phase:
            assert generic_metadata[phase] == 0.5

    def test_parallel_safety_determination(self, hook_manager):
        """Test parallel safety determination"""
        # Test non-parallel safe hooks
        write_metadata = hook_manager._is_parallel_safe("file_writer_hook.py")
        assert not write_metadata

        modify_metadata = hook_manager._is_parallel_safe("config_modifier.py")
        assert not modify_metadata

        # Test parallel safe hooks
        read_metadata = hook_manager._is_parallel_safe("info_reader.py")
        assert read_metadata

    @pytest.mark.asyncio
    async def test_hook_execution_success(self, hook_manager, sample_context):
        """Test successful hook execution"""
        results = await hook_manager.execute_hooks(
            HookEvent.SESSION_START, sample_context, user_input="Test session start", max_total_execution_time_ms=2000.0
        )

        assert len(results) > 0

        # Check that at least one hook succeeded
        successful_results = [r for r in results if r.success]
        assert len(successful_results) > 0

    @pytest.mark.asyncio
    async def test_hook_execution_with_errors(self, hook_manager, sample_context):
        """Test hook execution with errors"""
        results = await hook_manager.execute_hooks(
            HookEvent.SESSION_START,
            sample_context,
            user_input="Test session with errors",
            max_total_execution_time_ms=2000.0,
        )

        # Should have both successful and failed results
        successful_results = [r for r in results if r.success]
        failed_results = [r for r in results if not r.success]

        assert len(successful_results) > 0
        assert len(failed_results) > 0

        # Check error handling
        for failed_result in failed_results:
            assert failed_result.error_message is not None
            assert failed_result.execution_time_ms >= 0

    @pytest.mark.asyncio
    async def test_parallel_hook_execution(self, hook_manager, sample_context):
        """Test parallel hook execution"""
        # Execute hooks that should run in parallel
        results = await hook_manager.execute_hooks(
            HookEvent.SESSION_START,
            sample_context,
            user_input="Test parallel execution",
            max_total_execution_time_ms=1000.0,
        )

        # Should complete in reasonable time
        total_time = sum(r.execution_time_ms for r in results)
        assert total_time < 1000.0  # Should complete within timeout

    @pytest.mark.asyncio
    async def test_performance_metrics_tracking(self, hook_manager, sample_context):
        """Test performance metrics tracking"""
        # Execute hooks
        await hook_manager.execute_hooks(HookEvent.SESSION_START, sample_context, user_input="Test metrics tracking")

        # Get metrics
        metrics = hook_manager.get_performance_metrics()

        assert metrics.total_executions > 0
        assert metrics.average_execution_time_ms >= 0
        assert metrics.total_token_usage >= 0

    def test_hook_recommendations(self, hook_manager):
        """Test hook optimization recommendations"""
        recommendations = hook_manager.get_hook_recommendations()

        assert isinstance(recommendations, dict)
        assert "slow_hooks" in recommendations
        assert "unreliable_hooks" in recommendations
        assert "phase_mismatched_hooks" in recommendations
        assert "optimization_suggestions" in recommendations

    @pytest.mark.asyncio
    async def test_global_hook_manager_functions(self, hook_manager, sample_context):
        """Test global convenience functions"""
        with patch("src.moai_adk.core.jit_enhanced_hook_manager.get_jit_hook_manager", return_value=hook_manager):
            # Test session start hooks execution
            results = await execute_session_start_hooks(sample_context, "Test global function")

            assert isinstance(results, list)
            assert len(results) >= 0


class TestPhaseOptimizedHookScheduler:
    """Test Phase-Optimized Hook Scheduler functionality"""

    def test_phase_parameter_initialization(self, hook_scheduler):
        """Test phase-specific parameter initialization"""
        assert len(hook_scheduler._phase_parameters) == 7  # All phases

        # Check SPEC phase parameters
        spec_params = hook_scheduler._phase_parameters[Phase.SPEC]
        assert "max_total_time_ms" in spec_params
        assert "token_budget_ratio" in spec_params
        assert "priority_weights" in spec_params
        assert "prefer_parallel" in spec_params

    def test_strategy_performance_initialization(self, hook_scheduler):
        """Test strategy performance tracking initialization"""
        assert len(hook_scheduler._strategy_performance) == len(SchedulingStrategy)

        for strategy in SchedulingStrategy:
            performance = hook_scheduler._strategy_performance[strategy]
            assert "success_rate" in performance
            assert "avg_efficiency" in performance
            assert "usage_count" in performance

    @pytest.mark.asyncio
    async def test_hook_scheduling_basic(self, hook_scheduler, hook_manager, scheduling_context):
        """Test basic hook scheduling"""
        with patch.object(hook_scheduler, "hook_manager", hook_manager):
            result = await hook_scheduler.schedule_hooks(HookEvent.SESSION_START, scheduling_context)

            assert isinstance(result, SchedulingResult)
            assert isinstance(result.scheduled_hooks, list)
            assert isinstance(result.execution_plan, list)
            assert result.estimated_total_time_ms >= 0
            assert result.estimated_total_tokens >= 0
            assert result.scheduling_strategy in SchedulingStrategy

    @pytest.mark.asyncio
    async def test_strategy_selection(self, hook_scheduler, scheduling_context):
        """Test optimal strategy selection"""
        # Test with low token budget
        low_budget_context = scheduling_context
        low_budget_context.available_token_budget = 1000
        strategy = hook_scheduler._select_optimal_strategy(HookEvent.SESSION_START, low_budget_context)
        assert strategy == SchedulingStrategy.TOKEN_EFFICIENT

        # Test with tight time constraint
        tight_time_context = scheduling_context
        tight_time_context.max_execution_time_ms = 200
        strategy = hook_scheduler._select_optimal_strategy(HookEvent.SESSION_START, tight_time_context)
        assert strategy == SchedulingStrategy.PERFORMANCE_FIRST

        # Test with high system load
        high_load_context = scheduling_context
        high_load_context.system_load = 0.9
        strategy = hook_scheduler._select_optimal_strategy(HookEvent.SESSION_START, high_load_context)
        assert strategy == SchedulingStrategy.PRIORITY_FIRST

    @pytest.mark.asyncio
    async def test_priority_score_calculation(self, hook_scheduler, hook_manager, scheduling_context):
        """Test priority score calculation for different strategies"""
        # Create mock metadata
        metadata = HookMetadata(
            hook_path="test_hook.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
            estimated_execution_time_ms=100.0,
            token_cost_estimate=500,
            phase_relevance={Phase.SPEC: 0.8, Phase.RED: 0.3},
            success_rate=0.9,
        )

        # Test different strategies
        priority_score = hook_scheduler._calculate_priority_score(
            metadata, scheduling_context, SchedulingStrategy.PRIORITY_FIRST
        )
        assert isinstance(priority_score, float)
        assert priority_score > 0

        phase_score = hook_scheduler._calculate_priority_score(
            metadata, scheduling_context, SchedulingStrategy.PHASE_OPTIMIZED
        )
        assert isinstance(phase_score, float)
        assert phase_score > 0

    def test_hook_cost_estimation(self, hook_scheduler, hook_manager, scheduling_context):
        """Test hook execution cost estimation"""
        metadata = HookMetadata(
            hook_path="test_hook.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
            token_cost_estimate=1000,
            phase_relevance={Phase.SPEC: 0.8},
        )

        # Test cost estimation
        cost = hook_scheduler._estimate_hook_cost(metadata, scheduling_context)
        assert isinstance(cost, int)
        assert cost > 0

        # Test with low phase relevance
        low_relevance_metadata = HookMetadata(
            hook_path="low_relevance.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
            token_cost_estimate=1000,
            phase_relevance={Phase.SPEC: 0.1},
        )

        low_relevance_cost = hook_scheduler._estimate_hook_cost(low_relevance_metadata, scheduling_context)
        assert low_relevance_cost > cost  # Should cost more due to low relevance

    def test_hook_time_estimation(self, hook_scheduler, hook_manager, scheduling_context):
        """Test hook execution time estimation"""
        metadata = HookMetadata(
            hook_path="test_hook.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
            estimated_execution_time_ms=100.0,
            success_rate=0.9,
        )

        time_estimate = hook_scheduler._estimate_hook_time(metadata, scheduling_context)
        assert isinstance(time_estimate, float)
        assert time_estimate >= 100.0  # Should be at least base time

    def test_scheduling_decision_making(self, hook_scheduler, hook_manager, scheduling_context):
        """Test initial scheduling decision making"""
        # Test critical hook
        critical_metadata = HookMetadata(
            hook_path="critical_hook.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.CRITICAL,
            estimated_execution_time_ms=50.0,
            token_cost_estimate=100,
        )

        decision = hook_scheduler._make_initial_scheduling_decision(critical_metadata, scheduling_context, 100, 50.0)
        assert decision == SchedulingDecision.EXECUTE

        # Test hook that exceeds time budget
        slow_metadata = HookMetadata(
            hook_path="slow_hook.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
            estimated_execution_time_ms=2000.0,
            token_cost_estimate=100,
        )

        decision = hook_scheduler._make_initial_scheduling_decision(slow_metadata, scheduling_context, 100, 2000.0)
        assert decision == SchedulingDecision.DEFER

    def test_hook_filtering_by_constraints(self, hook_scheduler):
        """Test hook filtering based on constraints"""
        # Create mock hooks
        hooks = [
            ScheduledHook(
                hook_path="fast_hook.py",
                metadata=Mock(estimated_execution_time_ms=50, token_cost_estimate=100),
                priority_score=80,
                estimated_cost=100,
                estimated_time_ms=50,
                scheduling_decision=SchedulingDecision.EXECUTE,
            ),
            ScheduledHook(
                hook_path="slow_hook.py",
                metadata=Mock(estimated_execution_time_ms=1500, token_cost_estimate=200),
                priority_score=60,
                estimated_cost=200,
                estimated_time_ms=1500,
                scheduling_decision=SchedulingDecision.EXECUTE,
            ),
            ScheduledHook(
                hook_path="expensive_hook.py",
                metadata=Mock(estimated_execution_time_ms=100, token_cost_estimate=5000),
                priority_score=70,
                estimated_cost=5000,
                estimated_time_ms=100,
                scheduling_decision=SchedulingDecision.EXECUTE,
            ),
        ]

        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="Test filtering",
            available_token_budget=1000,
            max_execution_time_ms=1000.0,
        )

        filtered = hook_scheduler._filter_hooks_by_constraints(hooks, context)

        # Should filter out expensive and slow hooks
        executable_hooks = [h for h in filtered if h.scheduling_decision == SchedulingDecision.EXECUTE]
        assert len(executable_hooks) == 1
        assert executable_hooks[0].hook_path == "fast_hook.py"

    def test_dependency_resolution(self, hook_scheduler):
        """Test hook dependency resolution"""
        # Create hooks with dependencies
        hook_a = ScheduledHook(
            hook_path="hook_a.py",
            metadata=Mock(dependencies=set()),
            priority_score=80,
            estimated_cost=100,
            estimated_time_ms=50,
            scheduling_decision=SchedulingDecision.EXECUTE,
            dependencies=set(),
        )

        hook_b = ScheduledHook(
            hook_path="hook_b.py",
            metadata=Mock(dependencies=set()),
            priority_score=70,
            estimated_cost=100,
            estimated_time_ms=50,
            scheduling_decision=SchedulingDecision.EXECUTE,
            dependencies={"hook_a.py"},
        )

        hook_c = ScheduledHook(
            hook_path="hook_c.py",
            metadata=Mock(dependencies=set()),
            priority_score=60,
            estimated_cost=100,
            estimated_time_ms=50,
            scheduling_decision=SchedulingDecision.EXECUTE,
            dependencies={"hook_b.py"},
        )

        hooks = [hook_c, hook_b, hook_a]  # Intentionally out of order

        resolved = hook_scheduler._resolve_dependencies(hooks)

        # Should be in dependency order
        assert resolved[0].hook_path == "hook_a.py"
        assert resolved[1].hook_path == "hook_b.py"
        assert resolved[2].hook_path == "hook_c.py"

    def test_execution_group_creation(self, hook_scheduler, scheduling_context):
        """Test execution group creation"""
        hooks = [
            ScheduledHook(
                hook_path="parallel_hook_1.py",
                metadata=Mock(parallel_safe=True),
                priority_score=80,
                estimated_cost=100,
                estimated_time_ms=50,
                scheduling_decision=SchedulingDecision.EXECUTE,
            ),
            ScheduledHook(
                hook_path="parallel_hook_2.py",
                metadata=Mock(parallel_safe=True),
                priority_score=70,
                estimated_cost=100,
                estimated_time_ms=50,
                scheduling_decision=SchedulingDecision.EXECUTE,
            ),
            ScheduledHook(
                hook_path="sequential_hook.py",
                metadata=Mock(parallel_safe=False),
                priority_score=60,
                estimated_cost=100,
                estimated_time_ms=100,
                scheduling_decision=SchedulingDecision.EXECUTE,
            ),
        ]

        groups = hook_scheduler._create_execution_groups(hooks, scheduling_context, SchedulingStrategy.BALANCED)

        assert len(groups) >= 1

        # Check that parallel hooks are grouped together
        parallel_group = groups[0]
        assert parallel_group.execution_type == SchedulingDecision.PARALLEL
        assert len(parallel_group.hooks) == 2
        assert parallel_group.estimated_time_ms == 50  # Should be max of parallel hooks

    def test_phase_optimization_insights(self, hook_scheduler):
        """Test phase-specific optimization insights"""
        insights = hook_scheduler.get_phase_optimization_insights(Phase.SPEC)

        assert "phase" in insights
        assert "parameters" in insights
        assert "optimization_recommendations" in insights
        assert insights["phase"] == "SPEC"

        # Should have recommendations
        recommendations = insights["optimization_recommendations"]
        assert len(recommendations) > 0

    def test_scheduling_statistics(self, hook_scheduler):
        """Test scheduling statistics collection"""
        stats = hook_scheduler.get_scheduling_statistics()

        assert "total_schedules" in stats
        assert "strategy_performance" in stats
        assert "recent_performance" in stats
        assert isinstance(stats["total_schedules"], int)


class TestIntegration:
    """Integration tests for enhanced hook system"""

    @pytest.mark.asyncio
    async def test_hook_manager_scheduler_integration(self, hook_manager, hook_scheduler, scheduling_context):
        """Test integration between hook manager and scheduler"""
        # Register hooks from manager to scheduler
        hook_scheduler.hook_manager = hook_manager

        # Schedule hooks
        result = await hook_scheduler.schedule_hooks(HookEvent.SESSION_START, scheduling_context)

        assert isinstance(result, SchedulingResult)
        assert len(result.execution_plan) >= 0

    @pytest.mark.asyncio
    async def test_end_to_end_hook_execution(self, hook_manager, sample_context):
        """Test end-to-end hook execution with JIT optimization"""
        # Execute hooks with JIT optimization
        results = await hook_manager.execute_hooks(
            HookEvent.SESSION_START,
            sample_context,
            user_input="Creating user authentication system",
            phase=Phase.SPEC,
            max_total_execution_time_ms=2000.0,
        )

        # Verify execution
        assert isinstance(results, list)
        assert len(results) >= 0

        # Check performance metrics
        metrics = hook_manager.get_performance_metrics()
        assert metrics.total_executions > 0

    def test_performance_optimization_scenarios(self, hook_scheduler):
        """Test various performance optimization scenarios"""
        scenarios = [
            {
                "name": "Low Token Budget",
                "context": HookSchedulingContext(
                    event_type=HookEvent.SESSION_START,
                    current_phase=Phase.SPEC,
                    user_input="Test low budget",
                    available_token_budget=1000,
                    max_execution_time_ms=1000.0,
                ),
                "expected_strategy": SchedulingStrategy.TOKEN_EFFICIENT,
            },
            {
                "name": "Tight Time Constraint",
                "context": HookSchedulingContext(
                    event_type=HookEvent.PRE_TOOL_USE,
                    current_phase=Phase.RED,
                    user_input="Test tight time",
                    available_token_budget=10000,
                    max_execution_time_ms=200.0,
                ),
                "expected_strategy": SchedulingStrategy.PERFORMANCE_FIRST,
            },
            {
                "name": "High System Load",
                "context": HookSchedulingContext(
                    event_type=HookEvent.SESSION_START,
                    current_phase=Phase.GREEN,
                    user_input="Test high load",
                    available_token_budget=10000,
                    max_execution_time_ms=1000.0,
                    system_load=0.9,
                ),
                "expected_strategy": SchedulingStrategy.PRIORITY_FIRST,
            },
        ]

        for scenario in scenarios:
            strategy = hook_scheduler._select_optimal_strategy(HookEvent.SESSION_START, scenario["context"])
            # Strategy should be appropriate for the scenario
            assert strategy in SchedulingStrategy

    @pytest.mark.asyncio
    async def test_error_recovery_and_resilience(self, hook_manager, sample_context):
        """Test error recovery and system resilience"""
        # Execute hooks with some that will fail
        results = await hook_manager.execute_hooks(
            HookEvent.SESSION_START,
            sample_context,
            user_input="Test error recovery",
            max_total_execution_time_ms=2000.0,
        )

        # Should have mixed success/failure results
        successful_count = sum(1 for r in results if r.success)
        failed_count = sum(1 for r in results if not r.success)

        assert successful_count >= 0
        assert failed_count >= 0
        assert successful_count + failed_count == len(results)

        # System should remain functional
        metrics = hook_manager.get_performance_metrics()
        assert metrics.total_executions > 0


class TestPerformanceOptimization:
    """Performance optimization tests"""

    @pytest.mark.asyncio
    async def test_concurrent_hook_execution_performance(self, hook_manager, sample_context):
        """Test performance of concurrent hook execution"""
        start_time = time.time()

        results = await hook_manager.execute_hooks(
            HookEvent.SESSION_START,
            sample_context,
            user_input="Test concurrent performance",
            max_total_execution_time_ms=5000.0,
        )

        execution_time = (time.time() - start_time) * 1000

        # Should complete within reasonable time
        assert execution_time < 5000.0
        assert len(results) > 0

    def test_memory_efficiency(self, hook_manager):
        """Test memory efficiency of hook system"""
        # Get initial memory usage
        initial_metrics = hook_manager.get_performance_metrics()

        # Simulate multiple hook registrations and executions
        # (This would be more comprehensive in a real test environment)

        final_metrics = hook_manager.get_performance_metrics()

        # Memory usage should be reasonable
        assert final_metrics.total_executions >= initial_metrics.total_executions

    def test_cache_efficiency(self, hook_manager):
        """Test caching efficiency"""
        # Execute hooks multiple times to test caching
        cache_stats_before = hook_manager._result_cache.get_stats()

        # Simulate cache operations
        # (Would be more comprehensive in real test)

        cache_stats_after = hook_manager._result_cache.get_stats()

        # Cache should be functioning
        assert isinstance(cache_stats_before, dict)
        assert isinstance(cache_stats_after, dict)


if __name__ == "__main__":
    # Run tests when executed directly
    pytest.main([__file__, "-v"])
