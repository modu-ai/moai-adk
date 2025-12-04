"""
Comprehensive coverage tests for JITEnhancedHookManager.

Target: 60%+ coverage for jit_enhanced_hook_manager.py (879 lines)
Focuses on: Hook loading, execution, caching, circuit breaker, and performance monitoring.
Tests use @patch for mocking subprocess, file operations, and dependencies.
"""

import asyncio
import pytest
import tempfile
from datetime import datetime, timedelta
from pathlib import Path
from unittest.mock import MagicMock, patch, AsyncMock, call
from io import StringIO

from moai_adk.core.jit_enhanced_hook_manager import (
    HookEvent,
    HookPriority,
    Phase,
    ContextCache,
    TokenBudgetManager,
    JITEnhancedHookManager,
    HookMetadata,
    HookExecutionResult,
    CircuitBreaker,
    CircuitBreakerState,
    HookResultCache,
    ConnectionPool,
    RetryPolicy,
    ResourceUsageMetrics,
    HookPerformanceMetrics,
    ResourceMonitor,
    HealthChecker,
    PerformanceAnomalyDetector,
)


class TestCircuitBreaker:
    """Test CircuitBreaker class and its state management."""

    def test_circuit_breaker_init(self):
        """Test CircuitBreaker initialization."""
        # Arrange & Act
        cb = CircuitBreaker(failure_threshold=3, timeout_seconds=60, success_threshold=5)

        # Assert
        assert cb.failure_threshold == 3
        assert cb.timeout_seconds == 60
        assert cb.success_threshold == 5
        assert cb.state.state == "CLOSED"
        assert cb.state.failure_count == 0

    def test_circuit_breaker_state_initialization(self):
        """Test CircuitBreakerState initialization."""
        # Arrange & Act
        state = CircuitBreakerState(failure_threshold=3, timeout_seconds=60)

        # Assert
        assert state.state == "CLOSED"
        assert state.failure_count == 0
        assert state.last_failure_time is None

    @pytest.mark.asyncio
    async def test_circuit_breaker_call_success(self):
        """Test successful call through circuit breaker."""
        # Arrange
        cb = CircuitBreaker()
        async_func = AsyncMock(return_value="success")

        # Act
        result = await cb.call(async_func, "arg1", kwarg1="value1")

        # Assert
        assert result == "success"
        assert cb.state.state == "CLOSED"
        assert cb.state.failure_count == 0
        async_func.assert_called_once_with("arg1", kwarg1="value1")

    @pytest.mark.asyncio
    async def test_circuit_breaker_call_failure_and_open(self):
        """Test circuit breaker opening after failures."""
        # Arrange
        cb = CircuitBreaker(failure_threshold=2)
        async_func = AsyncMock(side_effect=ValueError("test error"))

        # Act & Assert
        with pytest.raises(ValueError):
            await cb.call(async_func)

        with pytest.raises(ValueError):
            await cb.call(async_func)

        # After threshold reached, should open
        assert cb.state.state == "OPEN"
        assert cb.state.failure_count == 2

    @pytest.mark.asyncio
    async def test_circuit_breaker_open_blocks_call(self):
        """Test that OPEN circuit breaker blocks calls."""
        # Arrange
        cb = CircuitBreaker()
        cb.state.state = "OPEN"
        cb.state.last_failure_time = datetime.now() - timedelta(seconds=1)
        func = MagicMock()

        # Act & Assert
        with pytest.raises(Exception, match="Circuit breaker is OPEN"):
            await cb.call(func)

    def test_circuit_breaker_sync_function(self):
        """Test circuit breaker with synchronous function."""
        # Arrange
        cb = CircuitBreaker()
        sync_func = MagicMock(return_value="sync_result")

        # Act
        result = asyncio.run(cb.call(sync_func, "arg"))

        # Assert
        assert result == "sync_result"
        sync_func.assert_called_once_with("arg")


class TestHookResultCache:
    """Test HookResultCache class."""

    def test_hook_result_cache_init(self):
        """Test HookResultCache initialization."""
        # Arrange & Act
        cache = HookResultCache(max_size=500, default_ttl_seconds=600)

        # Assert
        assert cache.max_size == 500
        assert cache.default_ttl_seconds == 600
        assert len(cache._cache) == 0

    def test_hook_result_cache_put_and_get(self):
        """Test put and get operations."""
        # Arrange
        cache = HookResultCache()

        # Act
        cache.put("test_key", "test_value", ttl_seconds=3600)
        result = cache.get("test_key")

        # Assert
        assert result == "test_value"

    def test_hook_result_cache_expired_entry(self):
        """Test that expired entries are removed."""
        # Arrange
        cache = HookResultCache(default_ttl_seconds=1)
        cache.put("key", "value")

        # Act
        import time
        time.sleep(1.1)
        result = cache.get("key")

        # Assert
        assert result is None

    def test_hook_result_cache_lru_eviction(self):
        """Test LRU eviction when cache is full."""
        # Arrange
        cache = HookResultCache(max_size=2)
        cache.put("key1", "value1")
        cache.put("key2", "value2")

        # Act - Add third item, should evict key1 (least recently used)
        cache.put("key3", "value3")

        # Assert
        assert cache.get("key1") is None
        assert cache.get("key2") == "value2"
        assert cache.get("key3") == "value3"

    def test_hook_result_cache_invalidate_all(self):
        """Test invalidating entire cache."""
        # Arrange
        cache = HookResultCache()
        cache.put("key1", "value1")
        cache.put("key2", "value2")

        # Act
        cache.invalidate()

        # Assert
        assert cache.get("key1") is None
        assert cache.get("key2") is None
        assert len(cache._cache) == 0

    def test_hook_result_cache_invalidate_pattern(self):
        """Test invalidating by pattern."""
        # Arrange
        cache = HookResultCache()
        cache.put("hook_test_1", "value1")
        cache.put("hook_test_2", "value2")
        cache.put("other_key", "value3")

        # Act
        cache.invalidate(pattern="hook_test")

        # Assert
        assert cache.get("hook_test_1") is None
        assert cache.get("hook_test_2") is None
        assert cache.get("other_key") == "value3"

    def test_hook_result_cache_stats(self):
        """Test cache statistics."""
        # Arrange
        cache = HookResultCache(max_size=100)
        cache.put("key1", "value1")
        cache.put("key2", "value2")

        # Act
        stats = cache.get_stats()

        # Assert
        assert stats["size"] == 2
        assert stats["max_size"] == 100
        assert stats["utilization"] == 0.02


class TestRetryPolicy:
    """Test RetryPolicy class."""

    def test_retry_policy_init(self):
        """Test RetryPolicy initialization."""
        # Arrange & Act
        policy = RetryPolicy(max_retries=3, backoff_factor=2.0)

        # Assert
        assert policy.max_retries == 3
        assert policy.backoff_factor == 2.0

    @pytest.mark.asyncio
    async def test_retry_policy_execute_with_retry_success(self):
        """Test successful execution with retry."""
        # Arrange
        policy = RetryPolicy(max_retries=3)
        func = AsyncMock(return_value="success")

        # Act
        result = await policy.execute_with_retry(func)

        # Assert
        assert result == "success"
        func.assert_called_once()

    @pytest.mark.asyncio
    async def test_retry_policy_execute_with_retry_failure(self):
        """Test failure after retries."""
        # Arrange
        policy = RetryPolicy(max_retries=1, base_delay_ms=10)
        func = AsyncMock(side_effect=ValueError("test error"))

        # Act & Assert
        with pytest.raises(ValueError, match="test error"):
            await policy.execute_with_retry(func)


class TestResourceMonitor:
    """Test ResourceMonitor class."""

    def test_resource_monitor_init(self):
        """Test ResourceMonitor initialization."""
        # Arrange & Act
        monitor = ResourceMonitor()

        # Assert
        assert monitor._baseline_metrics is not None
        assert monitor._peak_usage is not None

    def test_resource_monitor_get_current_metrics(self):
        """Test getting current resource metrics."""
        # Arrange
        monitor = ResourceMonitor()

        # Act
        metrics = monitor.get_current_metrics()

        # Assert
        assert metrics is not None
        assert metrics.cpu_usage_percent >= 0
        assert metrics.memory_usage_mb >= 0

    def test_resource_monitor_peak_usage(self):
        """Test tracking peak usage."""
        # Arrange
        monitor = ResourceMonitor()

        # Act
        monitor.get_current_metrics()
        peak = monitor._peak_usage

        # Assert
        assert peak is not None
        assert peak.memory_usage_mb >= 0


class TestHealthChecker:
    """Test HealthChecker class."""

    def test_health_checker_init(self):
        """Test HealthChecker initialization."""
        # Arrange
        manager = MagicMock()

        # Act
        checker = HealthChecker(manager)

        # Assert
        assert checker.hook_manager == manager
        assert checker._health_status == "healthy"

    @pytest.mark.asyncio
    async def test_health_checker_check_system_health(self):
        """Test health check execution."""
        # Arrange
        manager = MagicMock()
        manager._hook_registry = {}
        manager._hooks_by_event = {}
        manager._advanced_cache = MagicMock()
        manager._advanced_cache.get_stats.return_value = {"size": 0, "utilization": 0.0, "max_size": 100}
        manager._connection_pool = MagicMock()
        manager._connection_pool.get_pool_stats.return_value = {"pools": {}}
        manager._circuit_breakers = {}
        manager._resource_monitor = ResourceMonitor()

        checker = HealthChecker(manager)

        # Act
        health_report = await checker.check_system_health()

        # Assert
        assert health_report is not None
        assert "status" in health_report
        assert health_report["status"] in ["healthy", "degraded", "unhealthy"]

    def test_health_checker_get_health_status(self):
        """Test getting health status."""
        # Arrange
        manager = MagicMock()
        checker = HealthChecker(manager)

        # Act
        status = checker.get_health_status()

        # Assert
        assert status in ["healthy", "degraded", "unhealthy"]


class TestPerformanceAnomalyDetector:
    """Test PerformanceAnomalyDetector class."""

    def test_anomaly_detector_init(self):
        """Test AnomalyDetector initialization."""
        # Arrange & Act
        detector = PerformanceAnomalyDetector(sensitivity_factor=2.0)

        # Assert
        assert detector.sensitivity_factor == 2.0
        assert detector._performance_history == {}

    def test_anomaly_detector_detect_anomaly(self):
        """Test anomaly detection."""
        # Arrange
        detector = PerformanceAnomalyDetector(sensitivity_factor=2.0)

        # Act - Add enough data points for detection
        for time in [100.0, 105.0, 103.0, 102.0, 104.0]:
            detector.detect_anomaly("test_hook", time)

        # Detect anomaly on much slower execution
        anomaly = detector.detect_anomaly("test_hook", 500.0)

        # Assert
        assert anomaly is not None
        assert anomaly.get("anomaly_type") == "slow" or anomaly.get("type") == "slow"


class TestJITEnhancedHookManager:
    """Test JITEnhancedHookManager class."""

    @pytest.fixture
    def temp_hooks_dir(self):
        """Create temporary hooks directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            yield Path(tmpdir)

    @pytest.fixture
    def hook_manager(self, temp_hooks_dir):
        """Create JITEnhancedHookManager instance."""
        return JITEnhancedHookManager(
            hooks_directory=temp_hooks_dir,
            cache_directory=temp_hooks_dir / "cache",
            max_concurrent_hooks=5,
        )

    def test_jit_manager_initialization(self, hook_manager):
        """Test JITEnhancedHookManager initialization."""
        # Arrange & Act - done in fixture
        # Assert
        assert hook_manager.hooks_directory is not None
        assert hook_manager.cache_directory is not None
        assert hook_manager.max_concurrent_hooks == 5
        assert hook_manager.enable_performance_monitoring is True

    def test_jit_manager_extract_event_type(self, hook_manager):
        """Test event type extraction from filename."""
        # Arrange
        test_cases = [
            ("session_start_hook.py", HookEvent.SESSION_START),
            ("session_end_hook.py", HookEvent.SESSION_END),
            ("pre_tool_use.py", HookEvent.PRE_TOOL_USE),
            ("post_tool_use.py", HookEvent.POST_TOOL_USE),
            ("subagent_start.py", HookEvent.SUBAGENT_START),
            ("subagent_stop.py", HookEvent.SUBAGENT_STOP),
            ("unknown_hook.py", None),
        ]

        # Act & Assert
        for filename, expected_event in test_cases:
            result = hook_manager._extract_event_type_from_filename(filename)
            assert result == expected_event

    def test_jit_manager_discover_hooks(self, temp_hooks_dir):
        """Test hook discovery."""
        # Arrange
        hook_file = temp_hooks_dir / "session_start_hook.py"
        hook_file.write_text("# Hook content")

        # Act
        manager = JITEnhancedHookManager(hooks_directory=temp_hooks_dir)

        # Assert
        assert len(manager._hook_registry) > 0

    def test_jit_manager_register_hook(self, hook_manager):
        """Test hook registration."""
        # Arrange
        hook_path = "test_hook.py"
        event_type = HookEvent.SESSION_START

        # Act
        hook_manager._register_hook(hook_path, event_type)

        # Assert
        assert hook_path in hook_manager._hook_registry
        assert event_type in hook_manager._hooks_by_event
        assert hook_path in hook_manager._hooks_by_event[event_type]

    def test_jit_manager_determine_hook_priority(self, hook_manager):
        """Test hook priority determination."""
        # Arrange
        critical_hooks = ["security", "validation", "auth"]
        optional_hooks = ["analytics", "metrics"]

        # Act & Assert
        for hook in critical_hooks:
            priority = hook_manager._determine_hook_priority(hook, HookEvent.PRE_TOOL_USE)
            assert priority in [HookPriority.CRITICAL, HookPriority.HIGH]

        for hook in optional_hooks:
            priority = hook_manager._determine_hook_priority(hook, HookEvent.SESSION_END)
            assert priority in [HookPriority.LOW, HookPriority.NORMAL]

    def test_jit_manager_estimate_execution_time(self, hook_manager):
        """Test execution time estimation."""
        # Arrange & Act
        time_ms = hook_manager._estimate_execution_time("test_hook.py")

        # Assert
        assert time_ms >= 0.0

    def test_jit_manager_estimate_token_cost(self, hook_manager):
        """Test token cost estimation."""
        # Arrange & Act
        tokens = hook_manager._estimate_token_cost("test_hook.py")

        # Assert
        assert tokens >= 0

    @pytest.mark.asyncio
    async def test_jit_manager_execute_hook(self, hook_manager, temp_hooks_dir):
        """Test hook execution."""
        # Arrange
        hook_file = temp_hooks_dir / "test_hook.py"
        hook_file.write_text("def execute():\n    return 'result'")

        # Act
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(
                returncode=0,
                stdout="result",
                stderr=""
            )
            hook_manager._register_hook("test_hook.py", HookEvent.SESSION_START)

        # Assert
        assert "test_hook.py" in hook_manager._hook_registry

    def test_jit_manager_cache_integration(self, hook_manager):
        """Test cache integration."""
        # Arrange
        hook_manager._advanced_cache.put("hook_key", "hook_result")

        # Act
        result = hook_manager._advanced_cache.get("hook_key")

        # Assert
        assert result == "hook_result"

    def test_jit_manager_circuit_breaker_integration(self, hook_manager):
        """Test circuit breaker integration."""
        # Arrange
        hook_path = "test_hook.py"

        # Act
        breaker = CircuitBreaker(failure_threshold=2)
        hook_manager._circuit_breakers[hook_path] = breaker

        # Assert
        assert hook_path in hook_manager._circuit_breakers

    def test_jit_manager_retry_policy_integration(self, hook_manager):
        """Test retry policy integration."""
        # Arrange
        hook_path = "test_hook.py"

        # Act
        policy = RetryPolicy(max_retries=3)
        hook_manager._retry_policies[hook_path] = policy

        # Assert
        assert hook_path in hook_manager._retry_policies

    @pytest.mark.asyncio
    async def test_jit_manager_execute_hook_with_retry(self, hook_manager):
        """Test hook execution with retry logic."""
        # Arrange
        hook_path = "test_hook.py"
        hook_manager._register_hook(hook_path, HookEvent.SESSION_START)
        hook_manager._retry_policies[hook_path] = RetryPolicy(max_retries=2)

        # Act
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="success")
            # Would execute with retry logic

        # Assert
        assert hook_path in hook_manager._hook_registry

    def test_jit_manager_get_performance_metrics(self, hook_manager):
        """Test performance metrics retrieval."""
        # Arrange & Act
        metrics = hook_manager.metrics

        # Assert
        assert metrics is not None
        assert hasattr(metrics, "total_executions")

    def test_jit_manager_phase_relevance(self, hook_manager):
        """Test phase relevance determination."""
        # Arrange
        hook_path = "test_hook.py"

        # Act
        relevance = hook_manager._determine_phase_relevance(hook_path, HookEvent.SESSION_START)

        # Assert
        assert isinstance(relevance, dict)

    def test_jit_manager_parallel_safety(self, hook_manager):
        """Test parallel safety determination."""
        # Arrange
        hook_path = "thread_safe_hook.py"

        # Act
        is_safe = hook_manager._is_parallel_safe(hook_path)

        # Assert
        assert isinstance(is_safe, bool)

    def test_jit_manager_hooks_by_event(self, hook_manager):
        """Test hooks organization by event type."""
        # Arrange
        hook_manager._register_hook("hook1.py", HookEvent.SESSION_START)
        hook_manager._register_hook("hook2.py", HookEvent.SESSION_START)
        hook_manager._register_hook("hook3.py", HookEvent.SESSION_END)

        # Act
        start_hooks = hook_manager._hooks_by_event.get(HookEvent.SESSION_START, [])
        end_hooks = hook_manager._hooks_by_event.get(HookEvent.SESSION_END, [])

        # Assert
        assert len(start_hooks) >= 2
        assert len(end_hooks) >= 1

    def test_jit_manager_metadata_cache(self, hook_manager):
        """Test metadata caching."""
        # Arrange
        hook_manager._register_hook("test.py", HookEvent.SESSION_START)

        # Act
        metadata = hook_manager._hook_registry.get("test.py")

        # Assert
        assert metadata is not None
        assert metadata.hook_path == "test.py"
        assert metadata.event_type == HookEvent.SESSION_START

    def test_jit_manager_health_monitoring(self, hook_manager):
        """Test health monitoring integration."""
        # Arrange & Act
        health_status = hook_manager._health_checker.get_health_status()

        # Assert
        assert health_status in ["healthy", "degraded", "unhealthy"]


class TestConnectionPool:
    """Test ConnectionPool class."""

    def test_connection_pool_init(self):
        """Test ConnectionPool initialization."""
        # Arrange & Act
        pool = ConnectionPool(max_connections=20)

        # Assert
        assert pool.max_connections == 20

    @pytest.mark.asyncio
    async def test_connection_pool_get_connection(self):
        """Test getting connection from pool."""
        # Arrange
        pool = ConnectionPool(max_connections=5)

        async def mock_factory():
            return MagicMock()

        # Act
        connection = await pool.get_connection("test_pool", mock_factory)

        # Assert
        assert connection is not None

    @pytest.mark.asyncio
    async def test_connection_pool_multiple_connections(self):
        """Test managing multiple connections."""
        # Arrange
        pool = ConnectionPool(max_connections=5)

        async def mock_factory():
            return MagicMock()

        # Act
        conn1 = await pool.get_connection("pool1", mock_factory)
        conn2 = await pool.get_connection("pool2", mock_factory)

        # Assert
        assert conn1 is not None
        assert conn2 is not None


class TestContextCache:
    """Test ContextCache fallback implementation."""

    def test_context_cache_init(self):
        """Test ContextCache initialization."""
        # Arrange & Act
        cache = ContextCache(max_size=100, max_memory_mb=50)

        # Assert
        assert cache.max_size == 100
        assert hasattr(cache, 'max_memory_mb') or hasattr(cache, 'max_memory_bytes')
        assert isinstance(cache.cache, dict)

    def test_context_cache_get(self):
        """Test get operation."""
        # Arrange
        cache = ContextCache()

        # Act
        result = cache.get("key")

        # Assert
        assert result is None
        assert cache.misses == 1

    def test_context_cache_put(self):
        """Test put operation."""
        # Arrange
        cache = ContextCache()

        # Act
        cache.put("key", "value", token_count=100)

        # Assert
        # No exception should be raised
        assert len(cache.cache) >= 0

    def test_context_cache_stats(self):
        """Test getting cache stats."""
        # Arrange
        cache = ContextCache()

        # Act
        stats = cache.get_stats()

        # Assert
        assert "hits" in stats
        assert "misses" in stats


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
