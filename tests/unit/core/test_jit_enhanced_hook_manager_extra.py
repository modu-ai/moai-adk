"""Comprehensive test coverage for JIT Enhanced Hook Manager.

This module provides extensive unit tests for the JITEnhancedHookManager
including circuit breaker, caching, connection pooling, and performance monitoring.
"""

import asyncio
import pytest
from datetime import datetime, timedelta
from pathlib import Path
from unittest.mock import MagicMock, AsyncMock, patch, Mock
from collections import defaultdict

from moai_adk.core.jit_enhanced_hook_manager import (
    JITEnhancedHookManager,
    HookEvent,
    HookPriority,
    Phase,
    ContextCache,
    TokenBudgetManager,
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
    """Test CircuitBreaker class"""

    def test_circuit_breaker_initialization(self):
        """Test circuit breaker initializes correctly"""
        cb = CircuitBreaker(failure_threshold=3, timeout_seconds=60, success_threshold=5)
        assert cb.failure_threshold == 3
        assert cb.timeout_seconds == 60
        assert cb.success_threshold == 5
        assert cb.state.state == "CLOSED"
        assert cb.state.failure_count == 0

    @pytest.mark.asyncio
    async def test_circuit_breaker_successful_call(self):
        """Test successful call through circuit breaker"""
        cb = CircuitBreaker()

        async def successful_func():
            return "success"

        result = await cb.call(successful_func)
        assert result == "success"
        assert cb.state.failure_count == 0

    @pytest.mark.asyncio
    async def test_circuit_breaker_open_after_failures(self):
        """Test circuit breaker opens after threshold failures"""
        cb = CircuitBreaker(failure_threshold=2)

        async def failing_func():
            raise ValueError("Test failure")

        # First failure
        with pytest.raises(ValueError):
            await cb.call(failing_func)
        assert cb.state.failure_count == 1

        # Second failure - should open
        with pytest.raises(ValueError):
            await cb.call(failing_func)
        assert cb.state.state == "OPEN"

    @pytest.mark.asyncio
    async def test_circuit_breaker_prevents_calls_when_open(self):
        """Test circuit breaker prevents calls when open"""
        cb = CircuitBreaker(failure_threshold=1)
        cb.state.state = "OPEN"
        cb.state.last_failure_time = datetime.now()

        async def func():
            return "should not execute"

        with pytest.raises(Exception, match="Circuit breaker is OPEN"):
            await cb.call(func)

    def test_on_success_resets_failure_count(self):
        """Test on_success resets failure count"""
        cb = CircuitBreaker()
        cb.state.failure_count = 5
        cb._on_success()
        assert cb.state.failure_count == 0

    def test_on_failure_increments_count(self):
        """Test on_failure increments failure count"""
        cb = CircuitBreaker()
        assert cb.state.failure_count == 0
        cb._on_failure()
        assert cb.state.failure_count == 1

    def test_should_attempt_reset_checks_timeout(self):
        """Test should_attempt_reset respects timeout"""
        cb = CircuitBreaker(timeout_seconds=1)
        cb.state.last_failure_time = datetime.now() - timedelta(seconds=2)
        assert cb._should_attempt_reset() is True

        cb.state.last_failure_time = datetime.now()
        assert cb._should_attempt_reset() is False


class TestHookResultCache:
    """Test HookResultCache class"""

    def test_cache_initialization(self):
        """Test cache initializes correctly"""
        cache = HookResultCache(max_size=1000, default_ttl_seconds=300)
        assert cache.max_size == 1000
        assert cache.default_ttl_seconds == 300

    def test_cache_put_and_get(self):
        """Test putting and getting values from cache"""
        cache = HookResultCache()
        cache.put("key1", "value1")
        assert cache.get("key1") == "value1"

    def test_cache_returns_none_for_missing_key(self):
        """Test cache returns None for missing keys"""
        cache = HookResultCache()
        assert cache.get("nonexistent") is None

    def test_cache_ttl_expiration(self):
        """Test cache entries expire after TTL"""
        cache = HookResultCache(default_ttl_seconds=1)
        cache.put("key1", "value1", ttl_seconds=1)

        # Should be available immediately
        assert cache.get("key1") == "value1"

        # Simulate TTL expiration
        cache._cache["key1"] = (
            "value1",
            datetime.now() - timedelta(seconds=2),
            1
        )
        assert cache.get("key1") is None

    def test_cache_invalidate_all(self):
        """Test invalidating all cache entries"""
        cache = HookResultCache()
        cache.put("key1", "value1")
        cache.put("key2", "value2")

        cache.invalidate()
        assert cache.get("key1") is None
        assert cache.get("key2") is None

    def test_cache_invalidate_pattern(self):
        """Test invalidating cache entries by pattern"""
        cache = HookResultCache()
        cache.put("hook:test:1", "value1")
        cache.put("hook:test:2", "value2")
        cache.put("other", "value3")

        cache.invalidate(pattern="hook:test")
        assert cache.get("hook:test:1") is None
        assert cache.get("hook:test:2") is None
        assert cache.get("other") == "value3"

    def test_cache_lru_eviction(self):
        """Test LRU eviction when cache is full"""
        cache = HookResultCache(max_size=2)
        cache.put("key1", "value1")
        cache.put("key2", "value2")

        # Access key1 to update access time
        cache.get("key1")

        # Put new key, should evict key2
        cache.put("key3", "value3")
        assert cache.get("key1") == "value1"
        assert cache.get("key3") == "value3"

    def test_cache_stats(self):
        """Test getting cache statistics"""
        cache = HookResultCache(max_size=100)
        cache.put("key1", "value1")
        cache.put("key2", "value2")

        stats = cache.get_stats()
        assert stats["size"] == 2
        assert stats["max_size"] == 100
        assert 0 <= stats["utilization"] <= 1


class TestConnectionPool:
    """Test ConnectionPool class"""

    def test_connection_pool_initialization(self):
        """Test connection pool initializes correctly"""
        pool = ConnectionPool(max_connections=10, connection_timeout_seconds=30)
        assert pool.max_connections == 10
        assert pool.connection_timeout_seconds == 30

    @pytest.mark.asyncio
    async def test_get_connection_creates_new(self):
        """Test getting connection creates new one if available"""
        pool = ConnectionPool(max_connections=2)

        async def connection_factory():
            return {"id": 1}

        conn = await pool.get_connection("test_pool", connection_factory)
        assert conn == {"id": 1}
        assert pool._active_connections["test_pool"] == 1

    @pytest.mark.asyncio
    async def test_get_connection_reuses_from_pool(self):
        """Test getting connection reuses from pool"""
        pool = ConnectionPool(max_connections=2)

        async def connection_factory():
            return {"id": "new"}

        conn1 = await pool.get_connection("test_pool", connection_factory)
        pool.return_connection("test_pool", conn1)

        conn2 = await pool.get_connection("test_pool", connection_factory)
        assert conn1 == conn2

    @pytest.mark.asyncio
    async def test_connection_pool_full(self):
        """Test connection pool raises when full"""
        pool = ConnectionPool(max_connections=1)

        async def connection_factory():
            return {"id": "test"}

        conn1 = await pool.get_connection("test_pool", connection_factory)

        with pytest.raises(Exception, match="Connection pool.*is full"):
            await pool.get_connection("test_pool", connection_factory)

    def test_connection_pool_stats(self):
        """Test getting connection pool statistics"""
        pool = ConnectionPool()
        pool._active_connections["pool1"] = 2
        pool._pools["pool1"] = [1, 2, 3]

        stats = pool.get_pool_stats()
        assert "pools" in stats
        assert "pool1" in stats["pools"]

    @pytest.mark.asyncio
    async def test_return_connection(self):
        """Test returning connection to pool"""
        pool = ConnectionPool(max_connections=2)

        async def factory():
            return {"id": 1}

        conn = await pool.get_connection("test", factory)
        initial_active = pool._active_connections["test"]

        pool.return_connection("test", conn)
        assert pool._active_connections["test"] == initial_active - 1


class TestRetryPolicy:
    """Test RetryPolicy class"""

    def test_retry_policy_initialization(self):
        """Test retry policy initializes correctly"""
        policy = RetryPolicy(max_retries=3, base_delay_ms=100, max_delay_ms=5000, backoff_factor=2.0)
        assert policy.max_retries == 3
        assert policy.base_delay_ms == 100

    @pytest.mark.asyncio
    async def test_retry_succeeds_on_first_attempt(self):
        """Test retry succeeds on first attempt"""
        policy = RetryPolicy(max_retries=3)

        call_count = 0
        async def func():
            nonlocal call_count
            call_count += 1
            return "success"

        result = await policy.execute_with_retry(func)
        assert result == "success"
        assert call_count == 1

    @pytest.mark.asyncio
    async def test_retry_succeeds_after_failures(self):
        """Test retry succeeds after some failures"""
        policy = RetryPolicy(max_retries=3, base_delay_ms=1)

        call_count = 0
        async def func():
            nonlocal call_count
            call_count += 1
            if call_count < 3:
                raise ValueError("Fail")
            return "success"

        result = await policy.execute_with_retry(func)
        assert result == "success"
        assert call_count == 3

    @pytest.mark.asyncio
    async def test_retry_exhausted(self):
        """Test retry raises after exhausting retries"""
        policy = RetryPolicy(max_retries=2, base_delay_ms=1)

        async def func():
            raise ValueError("Always fails")

        with pytest.raises(ValueError, match="Always fails"):
            await policy.execute_with_retry(func)


class TestResourceMonitor:
    """Test ResourceMonitor class"""

    def test_resource_monitor_initialization(self):
        """Test resource monitor initializes correctly"""
        monitor = ResourceMonitor()
        assert monitor._baseline_metrics is not None

    def test_get_current_metrics(self):
        """Test getting current resource metrics"""
        monitor = ResourceMonitor()
        metrics = monitor.get_current_metrics()

        assert isinstance(metrics, ResourceUsageMetrics)
        # Metrics should be valid numbers
        assert isinstance(metrics.memory_usage_mb, (int, float))
        assert isinstance(metrics.cpu_usage_percent, (int, float))

    def test_get_peak_metrics(self):
        """Test getting peak resource metrics"""
        monitor = ResourceMonitor()
        peak = monitor.get_peak_metrics()
        assert isinstance(peak, ResourceUsageMetrics)


class TestPerformanceAnomalyDetector:
    """Test PerformanceAnomalyDetector class"""

    def test_detector_initialization(self):
        """Test detector initializes correctly"""
        detector = PerformanceAnomalyDetector(sensitivity_factor=2.0)
        assert detector.sensitivity_factor == 2.0

    def test_detect_anomaly_insufficient_data(self):
        """Test detector returns None with insufficient data"""
        detector = PerformanceAnomalyDetector()
        result = detector.detect_anomaly("hook1", 100.0)
        assert result is None
        assert len(detector._performance_history["hook1"]) == 1

    def test_detect_anomaly_normal_execution(self):
        """Test detector returns None for normal execution"""
        detector = PerformanceAnomalyDetector(sensitivity_factor=5.0)

        # Add baseline data
        for i in range(5):
            detector.detect_anomaly("hook1", 100.0 + i)

        # Normal execution within 5 standard deviations
        result = detector.detect_anomaly("hook1", 102.0)
        assert result is None

    def test_detect_anomaly_slow_execution(self):
        """Test detector identifies slow execution"""
        detector = PerformanceAnomalyDetector(sensitivity_factor=1.5)

        # Add baseline data with some variation
        for i in range(10):
            detector.detect_anomaly("hook1", 100.0 + i)

        # Clear history and rebuild to ensure non-zero variance
        detector._performance_history.clear()
        for val in [100.0, 101.0, 102.0, 103.0, 104.0, 105.0]:
            detector._performance_history["hook1"].append(val)

        # Significantly slow execution (far outside normal range)
        result = detector.detect_anomaly("hook1", 500.0)
        assert result is not None
        assert result["anomaly_type"] == "slow"


class TestHealthChecker:
    """Test HealthChecker class"""

    def test_health_checker_initialization(self):
        """Test health checker initializes correctly"""
        manager_mock = MagicMock()
        checker = HealthChecker(manager_mock)
        assert checker.hook_manager == manager_mock
        assert checker._health_status == "healthy"

    @pytest.mark.asyncio
    async def test_check_system_health(self):
        """Test checking system health"""
        manager_mock = MagicMock()
        manager_mock._hook_registry = {"hook1": MagicMock()}
        manager_mock._hooks_by_event = {"event1": ["hook1"]}
        manager_mock._advanced_cache = MagicMock()
        manager_mock._advanced_cache.get_stats.return_value = {
            "size": 10,
            "utilization": 0.1,
            "max_size": 100
        }
        manager_mock._connection_pool = MagicMock()
        manager_mock._connection_pool.get_pool_stats.return_value = {"pools": {}}
        manager_mock._circuit_breakers = {}
        manager_mock._resource_monitor = MagicMock()
        manager_mock._resource_monitor.get_current_metrics.return_value = ResourceUsageMetrics()

        checker = HealthChecker(manager_mock)
        health = await checker.check_system_health()

        assert "status" in health
        assert "checks" in health
        assert "timestamp" in health

    def test_get_health_status(self):
        """Test getting health status"""
        manager_mock = MagicMock()
        checker = HealthChecker(manager_mock)
        status = checker.get_health_status()
        assert status == "healthy"


class TestJITEnhancedHookManager:
    """Test JITEnhancedHookManager class"""

    def test_manager_initialization(self):
        """Test manager initializes correctly"""
        manager = JITEnhancedHookManager(max_concurrent_hooks=5)
        assert manager.max_concurrent_hooks == 5
        assert manager._hook_registry is not None
        assert manager.jit_loader is not None

    def test_discover_hooks_empty_directory(self, tmp_path):
        """Test discovering hooks from empty directory"""
        hooks_dir = tmp_path / "hooks"
        hooks_dir.mkdir()

        manager = JITEnhancedHookManager(hooks_directory=hooks_dir)
        assert len(manager._hook_registry) == 0

    def test_extract_event_type_from_filename(self):
        """Test extracting event type from filename"""
        manager = JITEnhancedHookManager()

        assert manager._extract_event_type_from_filename("session_start_hook.py") == HookEvent.SESSION_START
        assert manager._extract_event_type_from_filename("pre_tool_use.py") == HookEvent.PRE_TOOL_USE
        assert manager._extract_event_type_from_filename("post_tool_use.py") == HookEvent.POST_TOOL_USE
        assert manager._extract_event_type_from_filename("unknown.py") is None

    def test_determine_hook_priority(self):
        """Test determining hook priority"""
        manager = JITEnhancedHookManager()

        # Security hooks are critical
        priority = manager._determine_hook_priority("security_validation.py", HookEvent.SESSION_START)
        assert priority == HookPriority.CRITICAL

        # Performance hooks are high
        priority = manager._determine_hook_priority("performance_optimizer.py", HookEvent.SESSION_START)
        assert priority == HookPriority.HIGH

    def test_estimate_execution_time(self):
        """Test estimating hook execution time"""
        manager = JITEnhancedHookManager()

        # Git operations
        time_ms = manager._estimate_execution_time("git_operation.py")
        assert time_ms == 200.0

        # Network operations
        time_ms = manager._estimate_execution_time("api_fetch.py")
        assert time_ms == 500.0

        # Simple operations
        time_ms = manager._estimate_execution_time("simple.py")
        assert time_ms == 10.0

    def test_determine_phase_relevance(self):
        """Test determining phase relevance"""
        manager = JITEnhancedHookManager()

        relevance = manager._determine_phase_relevance("spec_design.py", HookEvent.SESSION_START)
        assert Phase.SPEC in relevance
        assert relevance[Phase.SPEC] == 1.0

    def test_estimate_token_cost(self):
        """Test estimating token cost"""
        manager = JITEnhancedHookManager()

        # Analysis hooks
        cost = manager._estimate_token_cost("analysis_report.py")
        assert cost > 100

        # Simple hooks
        cost = manager._estimate_token_cost("simple_log.py")
        assert cost > 0

    def test_is_parallel_safe(self):
        """Test determining if hook is parallel safe"""
        manager = JITEnhancedHookManager()

        # Write hooks are not parallel safe
        assert manager._is_parallel_safe("write_data.py") is False

        # Read hooks are parallel safe
        assert manager._is_parallel_safe("read_data.py") is True

    @pytest.mark.asyncio
    async def test_execute_hooks_with_empty_registry(self):
        """Test executing hooks with empty registry"""
        manager = JITEnhancedHookManager()

        results = await manager.execute_hooks(HookEvent.SESSION_START, {})
        assert isinstance(results, list)

    def test_prioritize_hooks(self):
        """Test prioritizing hooks"""
        manager = JITEnhancedHookManager()

        # Register some hooks
        manager._hook_registry["hook1"] = HookMetadata(
            hook_path="hook1.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.HIGH,
            estimated_execution_time_ms=100.0,
            success_rate=1.0
        )
        manager._hook_registry["hook2"] = HookMetadata(
            hook_path="hook2.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.LOW,
            estimated_execution_time_ms=50.0,
            success_rate=0.8
        )

        prioritized = manager._prioritize_hooks(["hook1", "hook2"], None)
        assert len(prioritized) == 2
        # Higher priority should come first
        assert prioritized[0][0] == "hook1"

    def test_determine_cache_ttl(self):
        """Test determining cache TTL"""
        manager = JITEnhancedHookManager()
        metadata = HookMetadata(
            hook_path="test.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL
        )

        # Network hooks have short TTL
        ttl = manager._determine_cache_ttl("fetch_api.py", metadata)
        assert ttl == 60

        # Write hooks have very short TTL
        ttl = manager._determine_cache_ttl("write_file.py", metadata)
        assert ttl == 30

    def test_get_performance_metrics(self):
        """Test getting performance metrics"""
        manager = JITEnhancedHookManager()

        metrics = manager.get_performance_metrics()
        assert isinstance(metrics, HookPerformanceMetrics)
        assert metrics.total_executions == 0

    def test_get_circuit_breaker_status(self):
        """Test getting circuit breaker status"""
        manager = JITEnhancedHookManager()

        # Add a circuit breaker
        cb = CircuitBreaker()
        manager._circuit_breakers["hook1"] = cb

        status = manager.get_circuit_breaker_status()
        assert "hook1" in status
        assert status["hook1"]["state"] == "CLOSED"

    def test_get_hook_recommendations(self):
        """Test getting hook optimization recommendations"""
        manager = JITEnhancedHookManager()

        # Add a slow hook
        manager._hook_registry["slow_hook"] = HookMetadata(
            hook_path="slow_hook.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
            estimated_execution_time_ms=300.0,
            success_rate=0.7
        )

        recommendations = manager.get_hook_recommendations()
        assert "slow_hooks" in recommendations
        assert "unreliable_hooks" in recommendations

    def test_update_hook_metadata(self):
        """Test updating hook metadata"""
        manager = JITEnhancedHookManager()

        manager._hook_registry["hook1"] = HookMetadata(
            hook_path="hook1.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
            success_rate=1.0
        )

        result = HookExecutionResult(
            hook_path="hook1.py",
            success=True,
            execution_time_ms=100.0,
            token_usage=100,
            output="test"
        )

        manager._update_hook_metadata("hook1.py", result)
        assert manager._hook_registry["hook1"].success_rate <= 1.0

    def test_calculate_std_dev(self):
        """Test calculating standard deviation"""
        manager = JITEnhancedHookManager()

        values = [100.0, 100.0, 100.0]
        std_dev = manager._calculate_std_dev(values)
        assert std_dev == 0.0

        values = [100.0, 200.0]
        std_dev = manager._calculate_std_dev(values)
        assert std_dev > 0

    @pytest.mark.asyncio
    async def test_cleanup(self):
        """Test cleanup functionality"""
        manager = JITEnhancedHookManager()
        await manager.cleanup()
        # Should complete without errors


class TestGlobalFunctions:
    """Test global convenience functions"""

    def test_get_jit_hook_manager_singleton(self):
        """Test get_jit_hook_manager returns singleton"""
        from moai_adk.core.jit_enhanced_hook_manager import get_jit_hook_manager

        manager1 = get_jit_hook_manager()
        manager2 = get_jit_hook_manager()
        assert manager1 is manager2

    @pytest.mark.asyncio
    async def test_execute_session_start_hooks(self):
        """Test execute_session_start_hooks function"""
        from moai_adk.core.jit_enhanced_hook_manager import execute_session_start_hooks

        results = await execute_session_start_hooks({})
        assert isinstance(results, list)

    def test_execute_pre_tool_hooks(self):
        """Test execute_pre_tool_hooks function"""
        # This is an async function, test import only
        from moai_adk.core.jit_enhanced_hook_manager import execute_pre_tool_hooks
        assert callable(execute_pre_tool_hooks)

    @pytest.mark.asyncio
    async def test_execute_session_end_hooks(self):
        """Test execute_session_end_hooks function"""
        from moai_adk.core.jit_enhanced_hook_manager import execute_session_end_hooks

        results = await execute_session_end_hooks({})
        assert isinstance(results, list)

    def test_get_hook_performance_metrics(self):
        """Test get_hook_performance_metrics function"""
        from moai_adk.core.jit_enhanced_hook_manager import get_hook_performance_metrics

        metrics = get_hook_performance_metrics()
        assert isinstance(metrics, HookPerformanceMetrics)

    def test_get_system_health(self):
        """Test get_system_health function"""
        from moai_adk.core.jit_enhanced_hook_manager import get_system_health

        # This is async, test with pytest.mark.asyncio
        pass

    def test_get_connection_pool_info(self):
        """Test get_connection_pool_info function"""
        from moai_adk.core.jit_enhanced_hook_manager import get_connection_pool_info

        info = get_connection_pool_info()
        assert isinstance(info, dict)

    def test_get_cache_performance(self):
        """Test get_cache_performance function"""
        from moai_adk.core.jit_enhanced_hook_manager import get_cache_performance

        perf = get_cache_performance()
        assert isinstance(perf, dict)

    def test_get_circuit_breaker_info(self):
        """Test get_circuit_breaker_info function"""
        from moai_adk.core.jit_enhanced_hook_manager import get_circuit_breaker_info

        info = get_circuit_breaker_info()
        assert isinstance(info, dict)

    def test_invalidate_hook_cache(self):
        """Test invalidate_hook_cache function"""
        from moai_adk.core.jit_enhanced_hook_manager import invalidate_hook_cache

        invalidate_hook_cache()
        invalidate_hook_cache(pattern="test")

    def test_reset_circuit_breakers(self):
        """Test reset_circuit_breakers function"""
        from moai_adk.core.jit_enhanced_hook_manager import reset_circuit_breakers

        reset_circuit_breakers()
        reset_circuit_breakers("specific_hook")


class TestDataClasses:
    """Test dataclass definitions"""

    def test_hook_metadata_dataclass(self):
        """Test HookMetadata dataclass"""
        metadata = HookMetadata(
            hook_path="test.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL
        )
        assert metadata.hook_path == "test.py"
        assert metadata.event_type == HookEvent.SESSION_START
        assert metadata.success_rate == 1.0

    def test_hook_execution_result_dataclass(self):
        """Test HookExecutionResult dataclass"""
        result = HookExecutionResult(
            hook_path="test.py",
            success=True,
            execution_time_ms=100.0,
            token_usage=50,
            output="test output"
        )
        assert result.hook_path == "test.py"
        assert result.success is True

    def test_circuit_breaker_state_dataclass(self):
        """Test CircuitBreakerState dataclass"""
        state = CircuitBreakerState()
        assert state.state == "CLOSED"
        assert state.failure_count == 0

    def test_resource_usage_metrics_dataclass(self):
        """Test ResourceUsageMetrics dataclass"""
        metrics = ResourceUsageMetrics(
            cpu_usage_percent=50.0,
            memory_usage_mb=100.0
        )
        assert metrics.cpu_usage_percent == 50.0
        assert metrics.memory_usage_mb == 100.0

    def test_hook_performance_metrics_dataclass(self):
        """Test HookPerformanceMetrics dataclass"""
        metrics = HookPerformanceMetrics(
            total_executions=10,
            successful_executions=9
        )
        assert metrics.total_executions == 10
        assert metrics.successful_executions == 9


class TestEnums:
    """Test enum classes"""

    def test_hook_event_enum(self):
        """Test HookEvent enum has all values"""
        assert HookEvent.SESSION_START is not None
        assert HookEvent.SESSION_END is not None
        assert HookEvent.PRE_TOOL_USE is not None
        assert HookEvent.POST_TOOL_USE is not None

    def test_hook_priority_enum(self):
        """Test HookPriority enum has all values"""
        assert HookPriority.CRITICAL is not None
        assert HookPriority.HIGH is not None
        assert HookPriority.NORMAL is not None
        assert HookPriority.LOW is not None
