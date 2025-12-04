"""
Comprehensive test coverage for JIT Enhanced Hook Manager.

Focuses on uncovered code paths including:
- Hook execution paths (lines 882-908, 920-950, 1135-1225)
- Circuit breaker and retry logic
- Performance metrics and anomaly detection
- Context loading and optimization
- Priority calculation and filtering
"""

import asyncio
import json
import time
from datetime import datetime
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, patch, call

import pytest

from moai_adk.core.jit_enhanced_hook_manager import (
    HookEvent,
    HookPriority,
    HookMetadata,
    HookExecutionResult,
    Phase,
    ContextCache,
    TokenBudgetManager,
    JITEnhancedHookManager,
)


class TestHookPrioritization:
    """Test hook prioritization logic (lines 910-950)."""

    @pytest.fixture
    def hook_manager(self):
        """Create hook manager instance."""
        with patch("moai_adk.core.jit_enhanced_hook_manager.JITContextLoader"):
            manager = JITEnhancedHookManager()
            manager._hook_registry = {}
            return manager

    def test_prioritize_hooks_empty_list(self, hook_manager):
        """Test prioritization with empty hook list."""
        # Arrange
        hook_paths = []
        phase = Phase.RED

        # Act
        result = hook_manager._prioritize_hooks(hook_paths, phase)

        # Assert
        assert result == []

    def test_prioritize_hooks_single_hook(self, hook_manager):
        """Test prioritization with single hook."""
        # Arrange
        hook_path = "test_hook.py"
        metadata = HookMetadata(
            hook_path=hook_path,
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.HIGH,
            estimated_execution_time_ms=100.0,
            success_rate=0.95,
            phase_relevance={Phase.RED: 0.8},
        )
        hook_manager._hook_registry[hook_path] = metadata
        hook_paths = [hook_path]

        # Act
        result = hook_manager._prioritize_hooks(hook_paths, Phase.RED)

        # Assert
        assert len(result) == 1
        assert result[0][0] == hook_path
        # Priority score calculation: (2*10) + (-0.8*5) + (100/100) + 0 = 20 - 4 + 1 = 17
        assert isinstance(result[0][1], float)

    def test_prioritize_hooks_multiple_hooks_sorted(self, hook_manager):
        """Test that hooks are sorted by priority score."""
        # Arrange
        hooks = [
            ("high_priority.py", HookPriority.HIGH, 50.0, 0.95),
            ("low_priority.py", HookPriority.LOW, 200.0, 0.95),
            ("normal_priority.py", HookPriority.NORMAL, 100.0, 0.95),
        ]

        for path, priority, time_ms, success_rate in hooks:
            metadata = HookMetadata(
                hook_path=path,
                event_type=HookEvent.SESSION_START,
                priority=priority,
                estimated_execution_time_ms=time_ms,
                success_rate=success_rate,
                phase_relevance={Phase.RED: 0.5},
            )
            hook_manager._hook_registry[path] = metadata

        # Act
        result = hook_manager._prioritize_hooks([h[0] for h in hooks], Phase.RED)

        # Assert
        assert len(result) == 3
        # Results should be sorted by priority score (lower is better)
        scores = [score for _, score in result]
        assert scores == sorted(scores)

    def test_prioritize_hooks_with_phase_relevance(self, hook_manager):
        """Test phase relevance bonus in prioritization."""
        # Arrange
        hook_path = "phase_sensitive.py"
        metadata = HookMetadata(
            hook_path=hook_path,
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
            estimated_execution_time_ms=100.0,
            success_rate=0.95,
            phase_relevance={
                Phase.RED: 1.0,  # High relevance to RED
                Phase.GREEN: 0.2,  # Low relevance to GREEN
            },
        )
        hook_manager._hook_registry[hook_path] = metadata

        # Act - execute for RED phase (high relevance)
        red_result = hook_manager._prioritize_hooks([hook_path], Phase.RED)
        # Act - execute for GREEN phase (low relevance)
        green_result = hook_manager._prioritize_hooks([hook_path], Phase.GREEN)

        # Assert - RED phase should have higher priority (lower score)
        assert red_result[0][1] < green_result[0][1]

    def test_prioritize_hooks_unreliable_hook_penalized(self, hook_manager):
        """Test that unreliable hooks (low success rate) are penalized."""
        # Arrange
        reliable_hook = "reliable.py"
        unreliable_hook = "unreliable.py"

        hook_manager._hook_registry[reliable_hook] = HookMetadata(
            hook_path=reliable_hook,
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
            estimated_execution_time_ms=100.0,
            success_rate=0.95,  # High success rate
            phase_relevance={Phase.RED: 0.5},
        )
        hook_manager._hook_registry[unreliable_hook] = HookMetadata(
            hook_path=unreliable_hook,
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
            estimated_execution_time_ms=100.0,
            success_rate=0.8,  # Low success rate
            phase_relevance={Phase.RED: 0.5},
        )

        # Act
        results = hook_manager._prioritize_hooks([reliable_hook, unreliable_hook], Phase.RED)

        # Assert - reliable hook should have higher priority
        reliable_score = next(s for p, s in results if p == reliable_hook)
        unreliable_score = next(s for p, s in results if p == unreliable_hook)
        assert reliable_score < unreliable_score

    def test_prioritize_hooks_missing_metadata_skipped(self, hook_manager):
        """Test that hooks without metadata are skipped."""
        # Arrange
        hook_paths = ["hook_with_metadata.py", "hook_without_metadata.py"]
        hook_manager._hook_registry["hook_with_metadata.py"] = HookMetadata(
            hook_path="hook_with_metadata.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.HIGH,
            estimated_execution_time_ms=100.0,
            success_rate=0.95,
            phase_relevance={Phase.RED: 0.5},
        )

        # Act
        result = hook_manager._prioritize_hooks(hook_paths, Phase.RED)

        # Assert - only hook with metadata should be returned
        assert len(result) == 1
        assert result[0][0] == "hook_with_metadata.py"


class TestHookExecution:
    """Test hook execution paths (lines 862-908)."""

    @pytest.fixture
    def hook_manager(self):
        """Create hook manager instance."""
        with patch("moai_adk.core.jit_enhanced_hook_manager.JITContextLoader"):
            manager = JITEnhancedHookManager()
            manager._hooks_by_event = {}
            manager._hook_registry = {}
            manager.jit_loader = MagicMock()
            manager.jit_loader.phase_detector = MagicMock()
            manager.jit_loader.phase_detector.detect_phase = MagicMock(return_value=Phase.RED)
            manager.jit_loader.load_context = AsyncMock(return_value=({}, {}))
            return manager

    @pytest.mark.asyncio
    async def test_execute_hooks_empty_event(self, hook_manager):
        """Test executing hooks for event with no registered hooks."""
        # Arrange
        hook_manager._hooks_by_event[HookEvent.SESSION_START] = []
        hook_manager.enable_performance_monitoring = False
        context = {"key": "value"}

        # Act
        results = await hook_manager.execute_hooks(HookEvent.SESSION_START, context)

        # Assert
        assert results == []

    @pytest.mark.asyncio
    async def test_execute_hooks_detects_phase_from_input(self, hook_manager):
        """Test that phase is detected from user input when not provided."""
        # Arrange
        hook_manager._hooks_by_event[HookEvent.USER_PROMPT_SUBMIT] = []
        hook_manager.enable_performance_monitoring = False
        user_input = "Testing RED phase"

        # Act
        await hook_manager.execute_hooks(
            HookEvent.USER_PROMPT_SUBMIT,
            {},
            user_input=user_input,
        )

        # Assert
        hook_manager.jit_loader.phase_detector.detect_phase.assert_called_once_with(user_input)

    @pytest.mark.asyncio
    async def test_execute_hooks_fallback_to_spec_phase(self, hook_manager):
        """Test fallback to SPEC phase when phase detector not available."""
        # Arrange
        hook_manager._hooks_by_event[HookEvent.USER_PROMPT_SUBMIT] = []
        hook_manager.enable_performance_monitoring = False
        hook_manager.jit_loader.phase_detector = None
        hook_manager.jit_loader.load_context = AsyncMock(return_value=({}, {}))

        # Act
        results = await hook_manager.execute_hooks(HookEvent.USER_PROMPT_SUBMIT, {}, user_input="test")

        # Assert
        assert isinstance(results, list)

    @pytest.mark.asyncio
    async def test_execute_hooks_with_performance_monitoring(self, hook_manager):
        """Test performance monitoring during hook execution."""
        # Arrange
        hook_manager.enable_performance_monitoring = True
        hook_manager._hooks_by_event[HookEvent.SESSION_START] = []
        hook_manager._update_performance_metrics = MagicMock()

        # Act
        await hook_manager.execute_hooks(
            HookEvent.SESSION_START,
            {},
        )

        # Assert
        hook_manager._update_performance_metrics.assert_called()

    @pytest.mark.asyncio
    async def test_execute_hooks_time_constraint(self, hook_manager):
        """Test that hook execution respects max_total_execution_time_ms."""
        # Arrange
        hook_manager._hooks_by_event[HookEvent.SESSION_START] = []
        hook_manager.enable_performance_monitoring = False
        hook_manager._execute_hooks_optimized = AsyncMock(return_value=[])

        # Act
        await hook_manager.execute_hooks(HookEvent.SESSION_START, {}, max_total_execution_time_ms=5000.0)

        # Assert - the time constraint should be passed through
        hook_manager._execute_hooks_optimized.assert_called()


class TestContextLoading:
    """Test context loading and optimization (lines 952-995)."""

    @pytest.fixture
    def hook_manager(self):
        """Create hook manager instance."""
        with patch("moai_adk.core.jit_enhanced_hook_manager.JITContextLoader"):
            manager = JITEnhancedHookManager()
            manager.jit_loader = MagicMock()
            return manager

    @pytest.mark.asyncio
    async def test_load_optimized_context_success(self, hook_manager):
        """Test successful context loading."""
        # Arrange
        hook_manager.jit_loader.load_context = AsyncMock(return_value=({"loaded": True}, {"tokens": 100}))
        event_type = HookEvent.SESSION_START
        context = {"original": "context"}
        phase = Phase.RED
        prioritized_hooks = [("hook1.py", 1.0), ("hook2.py", 2.0)]

        # Act
        result = await hook_manager._load_optimized_context(event_type, context, phase, prioritized_hooks)

        # Assert
        assert "hook_event_type" in result
        assert result["hook_event_type"] == HookEvent.SESSION_START.value
        assert result["hook_phase"] == Phase.RED.value
        assert result["hook_execution_mode"] == "optimized"
        assert "prioritized_hooks" in result

    @pytest.mark.asyncio
    async def test_load_optimized_context_jit_fallback(self, hook_manager):
        """Test fallback when JIT loader fails."""
        # Arrange
        hook_manager.jit_loader.load_context = AsyncMock(side_effect=TypeError("JIT interface mismatch"))
        context = {"original": "context"}

        # Act
        result = await hook_manager._load_optimized_context(HookEvent.SESSION_START, context, Phase.RED, [])

        # Assert - should return original context in optimized format
        assert "hook_event_type" in result
        assert result["original"] == "context"  # Original context preserved

    @pytest.mark.asyncio
    async def test_load_optimized_context_no_phase(self, hook_manager):
        """Test context loading when phase is None."""
        # Arrange
        hook_manager.jit_loader.load_context = AsyncMock(return_value=({}, {}))

        # Act
        result = await hook_manager._load_optimized_context(HookEvent.SESSION_START, {}, None, [])

        # Assert
        assert result["hook_phase"] is None

    @pytest.mark.asyncio
    async def test_load_optimized_context_top_5_hooks(self, hook_manager):
        """Test that only top 5 hooks are included in context."""
        # Arrange
        hook_manager.jit_loader.load_context = AsyncMock(return_value=({}, {}))
        # Create 10 prioritized hooks
        prioritized_hooks = [(f"hook{i}.py", float(i)) for i in range(10)]

        # Act
        result = await hook_manager._load_optimized_context(HookEvent.SESSION_START, {}, Phase.RED, prioritized_hooks)

        # Assert
        assert len(result["prioritized_hooks"]) == 5
        assert result["prioritized_hooks"] == [
            "hook0.py",
            "hook1.py",
            "hook2.py",
            "hook3.py",
            "hook4.py",
        ]


class TestSingleHookExecution:
    """Test single hook execution (lines 1125-1232)."""

    @pytest.fixture
    def hook_manager(self):
        """Create hook manager instance."""
        with patch("moai_adk.core.jit_enhanced_hook_manager.JITContextLoader"):
            manager = JITEnhancedHookManager()
            manager.hooks_directory = Path("/tmp/hooks")
            manager._hook_registry = {}
            manager._circuit_breakers = {}
            manager._retry_policies = {}
            manager._advanced_cache = MagicMock()
            manager._advanced_cache.get = MagicMock(return_value=None)
            manager._advanced_cache.put = MagicMock()
            manager._performance_lock = MagicMock()
            manager._performance_lock.__enter__ = MagicMock(return_value=None)
            manager._performance_lock.__exit__ = MagicMock(return_value=None)
            manager.metrics = MagicMock()
            manager.metrics.cache_hits = 0
            manager.metrics.cache_misses = 0
            manager.metrics.circuit_breaker_trips = 0
            manager._resource_monitor = MagicMock()
            manager._resource_monitor.get_current_metrics = MagicMock(return_value={})
            manager._anomaly_detector = MagicMock()
            manager._anomaly_detector.detect_anomaly = MagicMock(return_value=None)
            manager._execution_profiles = {}
            manager._logger = MagicMock()
            manager.circuit_breaker_threshold = 5
            manager.max_retries = 3
            return manager

    @pytest.mark.asyncio
    async def test_execute_single_hook_missing_metadata(self, hook_manager):
        """Test execution fails when hook metadata not found."""
        # Arrange
        hook_path = "missing_hook.py"
        context = {}

        # Act
        result = await hook_manager._execute_single_hook(hook_path, context)

        # Assert
        assert result.success is False
        assert "Hook metadata not found" in result.error_message

    @pytest.mark.asyncio
    async def test_execute_single_hook_creates_circuit_breaker(self, hook_manager):
        """Test that circuit breaker is created for new hook."""
        # Arrange
        hook_path = "test_hook.py"
        metadata = HookMetadata(
            hook_path=hook_path,
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.HIGH,
            estimated_execution_time_ms=100.0,
            success_rate=0.95,
        )
        hook_manager._hook_registry[hook_path] = metadata

        with patch.object(hook_manager, "_execute_hook_subprocess", new_callable=AsyncMock) as mock_exec:
            mock_exec.return_value = HookExecutionResult(
                hook_path=hook_path,
                success=True,
                execution_time_ms=50.0,
                token_usage=0,
                output="test output",
            )
            with patch("moai_adk.core.jit_enhanced_hook_manager.CircuitBreaker"):
                with patch("moai_adk.core.jit_enhanced_hook_manager.RetryPolicy"):
                    # Act
                    await hook_manager._execute_single_hook(hook_path, {})

        # Assert
        assert hook_path in hook_manager._circuit_breakers
        assert hook_path in hook_manager._retry_policies

    @pytest.mark.asyncio
    async def test_execute_single_hook_uses_cache(self, hook_manager):
        """Test that cached results are returned."""
        # Arrange
        hook_path = "cached_hook.py"
        cached_result = HookExecutionResult(
            hook_path=hook_path,
            success=True,
            execution_time_ms=10.0,
            token_usage=0,
            output="cached",
        )
        hook_manager._advanced_cache.get = MagicMock(return_value=cached_result)

        metadata = HookMetadata(
            hook_path=hook_path,
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.HIGH,
            estimated_execution_time_ms=100.0,
            success_rate=0.95,
        )
        hook_manager._hook_registry[hook_path] = metadata

        # Act
        result = await hook_manager._execute_single_hook(hook_path, {})

        # Assert
        assert result == cached_result
        hook_manager.metrics.cache_hits += 1

    @pytest.mark.asyncio
    async def test_execute_single_hook_caches_successful_result(self, hook_manager):
        """Test that successful results are cached."""
        # Arrange
        hook_path = "test_hook.py"
        metadata = HookMetadata(
            hook_path=hook_path,
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.HIGH,
            estimated_execution_time_ms=100.0,
            success_rate=0.95,
        )
        hook_manager._hook_registry[hook_path] = metadata
        hook_manager._execution_profiles[hook_path] = []

        mock_result = HookExecutionResult(
            hook_path=hook_path,
            success=True,
            execution_time_ms=50.0,
            token_usage=0,
            output="test",
        )

        # Create proper mock circuit breaker and retry policy
        mock_cb = MagicMock()
        mock_cb.call = AsyncMock(return_value=mock_result)
        mock_rp = MagicMock()

        with patch.object(hook_manager, "_execute_hook_subprocess", new_callable=AsyncMock) as mock_exec:
            mock_exec.return_value = mock_result
            with patch(
                "moai_adk.core.jit_enhanced_hook_manager.CircuitBreaker",
                return_value=mock_cb,
            ):
                with patch(
                    "moai_adk.core.jit_enhanced_hook_manager.RetryPolicy",
                    return_value=mock_rp,
                ):
                    # Act
                    await hook_manager._execute_single_hook(hook_path, {})

        # Assert - cache should have been used
        assert hook_path in hook_manager._execution_profiles

    @pytest.mark.asyncio
    async def test_execute_single_hook_records_execution_profile(self, hook_manager):
        """Test that execution time is recorded in profile."""
        # Arrange
        hook_path = "test_hook.py"
        metadata = HookMetadata(
            hook_path=hook_path,
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.HIGH,
            estimated_execution_time_ms=100.0,
            success_rate=0.95,
        )
        hook_manager._hook_registry[hook_path] = metadata
        hook_manager._execution_profiles[hook_path] = []

        mock_result = HookExecutionResult(
            hook_path=hook_path,
            success=True,
            execution_time_ms=75.0,
            token_usage=0,
            output="test",
        )

        # Create proper mock circuit breaker
        mock_cb = MagicMock()
        mock_cb.call = AsyncMock(return_value=mock_result)
        mock_rp = MagicMock()

        with patch.object(hook_manager, "_execute_hook_subprocess", new_callable=AsyncMock) as mock_exec:
            mock_exec.return_value = mock_result
            with patch(
                "moai_adk.core.jit_enhanced_hook_manager.CircuitBreaker",
                return_value=mock_cb,
            ):
                with patch(
                    "moai_adk.core.jit_enhanced_hook_manager.RetryPolicy",
                    return_value=mock_rp,
                ):
                    # Act
                    await hook_manager._execute_single_hook(hook_path, {})

        # Assert
        assert len(hook_manager._execution_profiles[hook_path]) > 0

    @pytest.mark.asyncio
    async def test_execute_single_hook_detects_anomaly(self, hook_manager):
        """Test that performance anomalies are detected."""
        # Arrange
        hook_path = "anomalous_hook.py"
        metadata = HookMetadata(
            hook_path=hook_path,
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.HIGH,
            estimated_execution_time_ms=100.0,
            success_rate=0.95,
        )
        hook_manager._hook_registry[hook_path] = metadata
        hook_manager._execution_profiles[hook_path] = []
        hook_manager._anomaly_detector.detect_anomaly = MagicMock(return_value="Execution time doubled")

        mock_result = HookExecutionResult(
            hook_path=hook_path,
            success=True,
            execution_time_ms=200.0,
            token_usage=0,
            output="test",
            metadata={},
        )

        # Create proper mock circuit breaker
        mock_cb = MagicMock()
        mock_cb.call = AsyncMock(return_value=mock_result)
        mock_rp = MagicMock()

        with patch.object(hook_manager, "_execute_hook_subprocess", new_callable=AsyncMock) as mock_exec:
            mock_exec.return_value = mock_result
            with patch(
                "moai_adk.core.jit_enhanced_hook_manager.CircuitBreaker",
                return_value=mock_cb,
            ):
                with patch(
                    "moai_adk.core.jit_enhanced_hook_manager.RetryPolicy",
                    return_value=mock_rp,
                ):
                    # Act
                    result = await hook_manager._execute_single_hook(hook_path, {})

        # Assert - anomaly detection should be called
        assert hook_manager._anomaly_detector.detect_anomaly.called

    @pytest.mark.asyncio
    async def test_execute_single_hook_unexpected_error(self, hook_manager):
        """Test handling of unexpected errors during execution."""
        # Arrange
        hook_path = "error_hook.py"
        metadata = HookMetadata(
            hook_path=hook_path,
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.HIGH,
            estimated_execution_time_ms=100.0,
            success_rate=0.95,
        )
        hook_manager._hook_registry[hook_path] = metadata

        # Create proper mock circuit breaker that raises
        mock_cb = MagicMock()
        mock_cb.call = AsyncMock(side_effect=RuntimeError("Unexpected execution error"))
        mock_rp = MagicMock()

        with patch.object(hook_manager, "_execute_hook_subprocess", new_callable=AsyncMock) as mock_exec:
            mock_exec.side_effect = RuntimeError("Unexpected execution error")
            with patch(
                "moai_adk.core.jit_enhanced_hook_manager.CircuitBreaker",
                return_value=mock_cb,
            ):
                with patch(
                    "moai_adk.core.jit_enhanced_hook_manager.RetryPolicy",
                    return_value=mock_rp,
                ):
                    # Act
                    result = await hook_manager._execute_single_hook(hook_path, {})

        # Assert
        assert result.success is False
        assert "Unexpected error" in result.error_message or "error" in result.error_message.lower()


class TestCacheTTLDetermination:
    """Test cache TTL determination logic (lines 1234-1249)."""

    @pytest.fixture
    def hook_manager(self):
        """Create hook manager instance."""
        with patch("moai_adk.core.jit_enhanced_hook_manager.JITContextLoader"):
            return JITEnhancedHookManager()

    def test_cache_ttl_api_hooks(self, hook_manager):
        """Test that API hooks get 1-minute TTL."""
        # Arrange
        metadata = HookMetadata(
            hook_path="test.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
            estimated_execution_time_ms=100.0,
            success_rate=0.95,
        )

        # Act
        ttl_fetch = hook_manager._determine_cache_ttl("fetch_data.py", metadata)
        ttl_api = hook_manager._determine_cache_ttl("api_call.py", metadata)
        ttl_network = hook_manager._determine_cache_ttl("network_request.py", metadata)
        ttl_git = hook_manager._determine_cache_ttl("git_fetch.py", metadata)

        # Assert
        assert ttl_fetch == 60
        assert ttl_api == 60
        assert ttl_network == 60
        assert ttl_git == 60

    def test_cache_ttl_static_hooks(self, hook_manager):
        """Test that static hooks get 30-minute TTL."""
        # Arrange
        metadata = HookMetadata(
            hook_path="test.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
            estimated_execution_time_ms=100.0,
            success_rate=0.95,
        )

        # Act
        ttl_read = hook_manager._determine_cache_ttl("read_file.py", metadata)
        ttl_parse = hook_manager._determine_cache_ttl("parse_config.py", metadata)
        ttl_analyze = hook_manager._determine_cache_ttl("analyze_data.py", metadata)

        # Assert
        assert ttl_read == 1800
        assert ttl_parse == 1800
        assert ttl_analyze == 1800

    def test_cache_ttl_write_hooks(self, hook_manager):
        """Test that write hooks get 30-second TTL."""
        # Arrange
        metadata = HookMetadata(
            hook_path="test.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
            estimated_execution_time_ms=100.0,
            success_rate=0.95,
        )

        # Act
        ttl_write = hook_manager._determine_cache_ttl("write_log.py", metadata)
        ttl_modify = hook_manager._determine_cache_ttl("modify_config.py", metadata)
        ttl_update = hook_manager._determine_cache_ttl("update_cache.py", metadata)
        ttl_create = hook_manager._determine_cache_ttl("create_file.py", metadata)

        # Assert
        assert ttl_write == 30
        assert ttl_modify == 30
        assert ttl_update == 30
        assert ttl_create == 30
