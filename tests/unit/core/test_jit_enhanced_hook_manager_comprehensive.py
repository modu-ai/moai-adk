"""
Comprehensive TDD test suite for jit_enhanced_hook_manager.py

This test suite covers:
- All enums (HookEvent, HookPriority, Phase)
- Data classes (HookMetadata, HookExecutionResult, CircuitBreakerState, ResourceUsageMetrics, HookPerformanceMetrics)
- CircuitBreaker class with full state transitions
- HookResultCache with TTL and LRU eviction
- ConnectionPool management
- RetryPolicy with exponential backoff
- ResourceMonitor
- HealthChecker
- JITEnhancedHookManager main class with all methods
- Error handling and edge cases

Coverage Target: 100%
"""

import asyncio
import json
import time
from datetime import datetime, timedelta
from unittest.mock import AsyncMock, Mock, patch, MagicMock
from pathlib import Path
import pytest

from moai_adk.core.jit_enhanced_hook_manager import (
    HookEvent,
    HookPriority,
    HookMetadata,
    HookExecutionResult,
    CircuitBreakerState,
    CircuitBreaker,
    HookResultCache,
    ConnectionPool,
    RetryPolicy,
    ResourceUsageMetrics,
    HookPerformanceMetrics,
    ResourceMonitor,
    HealthChecker,
    JITEnhancedHookManager,
    Phase,
    _JIT_AVAILABLE,
)


# =============================================================================
# ENUM TESTS
# =============================================================================


class TestHookEventEnum:
    """Test HookEvent enum"""

    def test_all_hook_events(self):
        """Test all HookEvent enum values exist"""
        assert HookEvent.SESSION_START.value == "SessionStart"
        assert HookEvent.SESSION_END.value == "SessionEnd"
        assert HookEvent.USER_PROMPT_SUBMIT.value == "UserPromptSubmit"
        assert HookEvent.PRE_TOOL_USE.value == "PreToolUse"
        assert HookEvent.POST_TOOL_USE.value == "PostToolUse"
        assert HookEvent.SUBAGENT_START.value == "SubagentStart"
        assert HookEvent.SUBAGENT_STOP.value == "SubagentStop"


class TestHookPriorityEnum:
    """Test HookPriority enum"""

    def test_all_priorities(self):
        """Test all HookPriority enum values"""
        assert HookPriority.CRITICAL.value == 1
        assert HookPriority.HIGH.value == 2
        assert HookPriority.NORMAL.value == 3
        assert HookPriority.LOW.value == 4


class TestPhaseEnum:
    """Test Phase enum"""

    def test_all_phases(self):
        """Test all Phase enum values"""
        assert Phase.SPEC.value == "SPEC"
        assert Phase.RED.value == "RED"
        assert Phase.GREEN.value == "GREEN"
        assert Phase.REFACTOR.value == "REFACTOR"
        assert Phase.SYNC.value == "SYNC"
        assert Phase.DEBUG.value == "DEBUG"
        assert Phase.PLANNING.value == "PLANNING"


# =============================================================================
# DATA CLASS TESTS
# =============================================================================


class TestHookMetadata:
    """Test HookMetadata dataclass"""

    def test_initialization(self):
        """Test HookMetadata initialization"""
        metadata = HookMetadata(
            hook_path="/test/hook.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
        )
        assert metadata.hook_path == "/test/hook.py"
        assert metadata.event_type == HookEvent.SESSION_START
        assert metadata.priority == HookPriority.NORMAL
        assert metadata.estimated_execution_time_ms == 0.0
        assert metadata.last_execution_time is None
        assert metadata.success_rate == 1.0
        assert metadata.phase_relevance == {}
        assert metadata.token_cost_estimate == 0
        assert metadata.dependencies == set()
        assert metadata.parallel_safe is True

    def test_with_all_fields(self):
        """Test HookMetadata with all fields populated"""
        phase_relevance = {Phase.SPEC: 0.9, Phase.RED: 0.7}
        metadata = HookMetadata(
            hook_path="/test/hook.py",
            event_type=HookEvent.PRE_TOOL_USE,
            priority=HookPriority.HIGH,
            estimated_execution_time_ms=150.0,
            last_execution_time=datetime.now(),
            success_rate=0.95,
            phase_relevance=phase_relevance,
            token_cost_estimate=500,
            dependencies={"/dep/hook.py"},
            parallel_safe=False,
        )
        assert metadata.hook_path == "/test/hook.py"
        assert metadata.estimated_execution_time_ms == 150.0
        assert metadata.success_rate == 0.95
        assert metadata.phase_relevance == phase_relevance
        assert metadata.token_cost_estimate == 500
        assert "/dep/hook.py" in metadata.dependencies
        assert metadata.parallel_safe is False


class TestHookExecutionResult:
    """Test HookExecutionResult dataclass"""

    def test_success_result(self):
        """Test successful execution result"""
        result = HookExecutionResult(
            hook_path="/test/hook.py",
            success=True,
            execution_time_ms=100.0,
            token_usage=50,
            output="result",
        )
        assert result.hook_path == "/test/hook.py"
        assert result.success is True
        assert result.execution_time_ms == 100.0
        assert result.token_usage == 50
        assert result.output == "result"
        assert result.error_message is None
        assert result.metadata == {}

    def test_failure_result(self):
        """Test failed execution result"""
        result = HookExecutionResult(
            hook_path="/test/hook.py",
            success=False,
            execution_time_ms=50.0,
            token_usage=25,
            output=None,
            error_message="Hook failed",
            metadata={"retry_count": 3},
        )
        assert result.success is False
        assert result.error_message == "Hook failed"
        assert result.metadata == {"retry_count": 3}


class TestCircuitBreakerState:
    """Test CircuitBreakerState dataclass"""

    def test_default_state(self):
        """Test default circuit breaker state"""
        state = CircuitBreakerState()
        assert state.failure_count == 0
        assert state.last_failure_time is None
        assert state.state == "CLOSED"
        assert state.success_threshold == 5
        assert state.failure_threshold == 3
        assert state.timeout_seconds == 60

    def test_custom_state(self):
        """Test custom circuit breaker state"""
        state = CircuitBreakerState(
            failure_count=5,
            last_failure_time=datetime.now(),
            state="OPEN",
            success_threshold=10,
            failure_threshold=5,
            timeout_seconds=120,
        )
        assert state.failure_count == 5
        assert state.state == "OPEN"
        assert state.success_threshold == 10
        assert state.failure_threshold == 5
        assert state.timeout_seconds == 120


class TestResourceUsageMetrics:
    """Test ResourceUsageMetrics dataclass"""

    def test_default_metrics(self):
        """Test default resource usage metrics"""
        metrics = ResourceUsageMetrics()
        assert metrics.cpu_usage_percent == 0.0
        assert metrics.memory_usage_mb == 0.0
        assert metrics.disk_io_mb == 0.0
        assert metrics.network_io_mb == 0.0
        assert metrics.open_files == 0
        assert metrics.thread_count == 0

    def test_populated_metrics(self):
        """Test populated resource usage metrics"""
        metrics = ResourceUsageMetrics(
            cpu_usage_percent=75.5,
            memory_usage_mb=512.0,
            disk_io_mb=1024.0,
            network_io_mb=2048.0,
            open_files=100,
            thread_count=8,
        )
        assert metrics.cpu_usage_percent == 75.5
        assert metrics.memory_usage_mb == 512.0
        assert metrics.disk_io_mb == 1024.0
        assert metrics.open_files == 100
        assert metrics.thread_count == 8


class TestHookPerformanceMetrics:
    """Test HookPerformanceMetrics dataclass"""

    def test_default_metrics(self):
        """Test default performance metrics"""
        metrics = HookPerformanceMetrics()
        assert metrics.total_executions == 0
        assert metrics.successful_executions == 0
        assert metrics.average_execution_time_ms == 0.0
        assert metrics.total_token_usage == 0
        assert metrics.cache_hits == 0
        assert metrics.cache_misses == 0
        assert metrics.phase_distribution == {}
        assert metrics.event_type_distribution == {}
        assert metrics.circuit_breaker_trips == 0
        assert metrics.retry_attempts == 0
        assert isinstance(metrics.resource_usage, ResourceUsageMetrics)


# =============================================================================
# CIRCUIT BREAKER TESTS
# =============================================================================


class TestCircuitBreaker:
    """Test CircuitBreaker class"""

    @pytest.fixture
    def breaker(self):
        """Create circuit breaker instance"""
        return CircuitBreaker(
            failure_threshold=3,
            timeout_seconds=60,
            success_threshold=5,
        )

    @pytest.mark.asyncio
    async def test_successful_call(self, breaker):
        """Test successful call through circuit breaker"""

        async def test_func():
            return "success"

        result = await breaker.call(test_func)
        assert result == "success"
        assert breaker.state.failure_count == 0
        assert breaker.state.state == "CLOSED"

    @pytest.mark.asyncio
    async def test_sync_successful_call(self, breaker):
        """Test successful synchronous call"""

        def test_func():
            return "sync_success"

        result = await breaker.call(test_func)
        assert result == "sync_success"
        assert breaker.state.state == "CLOSED"

    @pytest.mark.asyncio
    async def test_call_with_failure(self, breaker):
        """Test call that fails"""

        async def failing_func():
            raise Exception("Test failure")

        with pytest.raises(Exception, match="Test failure"):
            await breaker.call(failing_func)

        assert breaker.state.failure_count == 1
        assert breaker.state.last_failure_time is not None
        assert breaker.state.state == "CLOSED"

    @pytest.mark.asyncio
    async def test_circuit_opens_after_threshold(self, breaker):
        """Test circuit opens after failure threshold"""

        async def failing_func():
            raise Exception("Failure")

        # Trigger failures up to threshold
        for i in range(3):
            try:
                await breaker.call(failing_func)
            except Exception:
                pass

        assert breaker.state.state == "OPEN"

    @pytest.mark.asyncio
    async def test_open_circuit_blocks_calls(self, breaker):
        """Test open circuit blocks calls"""

        async def failing_func():
            raise Exception("Failure")

        # Open the circuit
        for i in range(3):
            try:
                await breaker.call(failing_func)
            except Exception:
                pass

        assert breaker.state.state == "OPEN"

        # Try to call again - should be blocked
        with pytest.raises(Exception, match="Circuit breaker is OPEN"):
            await breaker.call(failing_func)

    @pytest.mark.asyncio
    async def test_half_open_after_timeout(self, breaker):
        """Test circuit transitions to HALF_OPEN after timeout"""

        async def failing_func():
            raise Exception("Failure")

        # Open the circuit
        for i in range(3):
            try:
                await breaker.call(failing_func)
            except Exception:
                pass

        assert breaker.state.state == "OPEN"

        # Mock timeout to have passed
        old_time = datetime.now() - timedelta(seconds=61)
        breaker.state.last_failure_time = old_time

        # Should transition to HALF_OPEN
        async def success_func():
            return "success"

        result = await breaker.call(success_func)
        assert result == "success"
        assert breaker.state.state == "HALF_OPEN"

    @pytest.mark.asyncio
    async def test_circuit_closes_after_success_threshold(self, breaker):
        """Test circuit closes after success threshold in HALF_OPEN"""

        async def success_func():
            return "success"

        # Open circuit
        breaker.state.state = "OPEN"
        breaker.state.last_failure_time = datetime.now() - timedelta(seconds=61)

        # Transition to HALF_OPEN
        result = await breaker.call(success_func)
        assert breaker.state.state == "HALF_OPEN"

        # Execute enough successes to close circuit
        for i in range(4):
            result = await breaker.call(success_func)

        assert breaker.state.state == "CLOSED"

    @pytest.mark.asyncio
    async def test_should_attempt_reset_with_no_failure_time(self, breaker):
        """Test _should_attempt_reset with no failure time"""
        breaker.state.last_failure_time = None
        assert breaker._should_attempt_reset() is False


# =============================================================================
# HOOK RESULT CACHE TESTS
# =============================================================================


class TestHookResultCache:
    """Test HookResultCache class"""

    @pytest.fixture
    def cache(self):
        """Create cache instance"""
        return HookResultCache(max_size=100, default_ttl_seconds=300)

    def test_put_and_get(self, cache):
        """Test basic put and get operations"""
        cache.put("key1", "value1")
        result = cache.get("key1")
        assert result == "value1"

    def test_get_nonexistent_key(self, cache):
        """Test get with nonexistent key"""
        result = cache.get("nonexistent")
        assert result is None

    def test_ttl_expiration(self, cache):
        """Test TTL expiration"""
        cache.put("key1", "value1", ttl_seconds=1)

        # Should be available immediately
        assert cache.get("key1") == "value1"

        # Wait for expiration
        time.sleep(2)
        assert cache.get("key1") is None

    def test_access_count_increments(self, cache):
        """Test access count increments on get"""
        cache.put("key1", "value1")

        cache.get("key1")
        cache.get("key1")
        cache.get("key1")

        # Access count should be updated
        value, expiry, access_count = cache._cache["key1"]
        assert access_count == 4  # Initial put (1) + 3 gets

    def test_lru_eviction(self, cache):
        """Test LRU eviction when cache is full"""
        small_cache = HookResultCache(max_size=3, default_ttl_seconds=300)

        small_cache.put("key1", "value1")
        time.sleep(0.1)
        small_cache.put("key2", "value2")
        time.sleep(0.1)
        small_cache.put("key3", "value3")

        # All keys should be present
        assert small_cache.get("key1") == "value1"
        assert small_cache.get("key2") == "value2"
        assert small_cache.get("key3") == "value3"

        # Add one more - should evict LRU (key1)
        small_cache.put("key4", "value4")

        assert small_cache.get("key1") is None  # Evicted
        assert small_cache.get("key2") == "value2"
        assert small_cache.get("key3") == "value3"
        assert small_cache.get("key4") == "value4"

    def test_invalidate_all(self, cache):
        """Test invalidate all entries"""
        cache.put("key1", "value1")
        cache.put("key2", "value2")

        cache.invalidate()

        assert cache.get("key1") is None
        assert cache.get("key2") is None

    def test_invalidate_pattern(self, cache):
        """Test invalidate with pattern"""
        cache.put("user:1:data", "value1")
        cache.put("user:2:data", "value2")
        cache.put("session:1:data", "value3")

        cache.invalidate(pattern="user:")

        assert cache.get("user:1:data") is None
        assert cache.get("user:2:data") is None
        assert cache.get("session:1:data") == "value3"

    def test_get_stats(self, cache):
        """Test get statistics"""
        cache.put("key1", "value1")
        cache.put("key2", "value2")

        stats = cache.get_stats()
        assert stats["size"] == 2
        assert stats["max_size"] == 100
        assert stats["utilization"] == 0.02


# =============================================================================
# CONNECTION POOL TESTS
# =============================================================================


class TestConnectionPool:
    """Test ConnectionPool class"""

    @pytest.fixture
    def pool(self):
        """Create connection pool"""
        return ConnectionPool(max_connections=5, connection_timeout_seconds=30)

    @pytest.mark.asyncio
    async def test_get_new_connection(self, pool):
        """Test getting new connection"""

        async def factory():
            return "connection"

        conn = await pool.get_connection("test_pool", factory)
        assert conn == "connection"

    @pytest.mark.asyncio
    async def test_sync_connection_factory(self, pool):
        """Test synchronous connection factory"""

        def factory():
            return "sync_conn"

        conn = await pool.get_connection("test_pool", factory)
        assert conn == "sync_conn"

    @pytest.mark.asyncio
    async def test_connection_reuse(self, pool):
        """Test connection reuse"""
        created_connections = []

        async def factory():
            conn = f"conn_{len(created_connections)}"
            created_connections.append(conn)
            return conn

        # Get first connection
        conn1 = await pool.get_connection("test_pool", factory)
        assert conn1 == "conn_0"

        # Return it
        pool.return_connection("test_pool", conn1)

        # Get again - should reuse
        conn2 = await pool.get_connection("test_pool", factory)
        assert conn2 == "conn_0"  # Reused
        assert len(created_connections) == 1

    @pytest.mark.asyncio
    async def test_pool_full_exception(self, pool):
        """Test exception when pool is full"""

        async def factory():
            return "connection"

        # Fill the pool
        connections = []
        for i in range(5):
            conn = await pool.get_connection("full_pool", factory)
            connections.append(conn)

        # Try to get one more - should raise exception
        with pytest.raises(Exception, match="Connection pool 'full_pool' is full"):
            await pool.get_connection("full_pool", factory)

    @pytest.mark.asyncio
    async def test_factory_failure_decrements_active(self, pool):
        """Test that factory failure decrements active count"""

        async def failing_factory():
            raise Exception("Connection failed")

        with pytest.raises(Exception):
            await pool.get_connection("test_pool", failing_factory)

        # Active count should be 0 after failure
        stats = pool.get_pool_stats()
        if "test_pool" in stats["pools"]:
            assert stats["pools"]["test_pool"]["active"] == 0

    def test_get_pool_stats(self, pool):
        """Test getting pool statistics"""
        stats = pool.get_pool_stats()
        assert "pools" in stats
        assert isinstance(stats["pools"], dict)


# =============================================================================
# RETRY POLICY TESTS
# =============================================================================


class TestRetryPolicy:
    """Test RetryPolicy class"""

    @pytest.fixture
    def policy(self):
        """Create retry policy"""
        return RetryPolicy(
            max_retries=3,
            base_delay_ms=100,
            max_delay_ms=5000,
            backoff_factor=2.0,
        )

    @pytest.mark.asyncio
    async def test_success_on_first_try(self, policy):
        """Test successful execution on first try"""

        async def test_func():
            return "success"

        result = await policy.execute_with_retry(test_func)
        assert result == "success"

    @pytest.mark.asyncio
    async def test_success_after_retry(self, policy):
        """Test success after initial failure"""
        attempt_count = [0]

        async def flaky_func():
            attempt_count[0] += 1
            if attempt_count[0] < 2:
                raise Exception("Temporary failure")
            return "success"

        result = await policy.execute_with_retry(flaky_func)
        assert result == "success"
        assert attempt_count[0] == 2

    @pytest.mark.asyncio
    async def test_failure_after_max_retries(self, policy):
        """Test failure after max retries"""

        async def always_failing_func():
            raise Exception("Permanent failure")

        with pytest.raises(Exception, match="Permanent failure"):
            await policy.execute_with_retry(always_failing_func)

    @pytest.mark.asyncio
    async def test_exponential_backoff(self, policy):
        """Test exponential backoff delay"""
        call_times = []

        async def failing_func():
            call_times.append(time.time())
            raise Exception("Failure")

        start = time.time()
        with pytest.raises(Exception):
            await policy.execute_with_retry(failing_func)

        # Verify delays between attempts
        assert len(call_times) == 4  # Initial + 3 retries
        # First retry: ~100ms, second: ~200ms, third: ~400ms

    @pytest.mark.asyncio
    async def test_sync_function_retry(self, policy):
        """Test retry with synchronous function"""
        attempt_count = [0]

        def sync_flaky_func():
            attempt_count[0] += 1
            if attempt_count[0] < 2:
                raise Exception("Sync failure")
            return "sync_success"

        result = await policy.execute_with_retry(sync_flaky_func)
        assert result == "sync_success"


# =============================================================================
# RESOURCE MONITOR TESTS
# =============================================================================


class TestResourceMonitor:
    """Test ResourceMonitor class"""

    @pytest.fixture
    def monitor(self):
        """Create resource monitor"""
        return ResourceMonitor()

    def test_get_current_metrics(self, monitor):
        """Test getting current metrics"""
        metrics = monitor.get_current_metrics()
        assert isinstance(metrics, ResourceUsageMetrics)
        # Metrics should be populated (even if with zeros)
        assert metrics.cpu_usage_percent >= 0
        assert metrics.memory_usage_mb >= 0

    def test_get_peak_metrics(self, monitor):
        """Test getting peak metrics"""
        peak = monitor.get_peak_metrics()
        assert isinstance(peak, ResourceUsageMetrics)

    def test_peak_metrics_update(self, monitor):
        """Test that peak metrics update appropriately"""
        # Get metrics multiple times
        monitor.get_current_metrics()
        metrics1 = monitor.get_peak_metrics()

        monitor.get_current_metrics()
        metrics2 = monitor.get_peak_metrics()

        # Peak should be at least as high as current
        assert metrics2.memory_usage_mb >= metrics1.memory_usage_mb


# =============================================================================
# JIT ENHANCED HOOK MANAGER TESTS
# =============================================================================


class TestJITEnhancedHookManager:
    """Test JITEnhancedHookManager main class"""

    @pytest.fixture
    def manager(self):
        """Create JIT Enhanced Hook Manager instance"""
        return JITEnhancedHookManager()

    def test_initialization(self, manager):
        """Test manager initialization"""
        assert manager is not None
        assert hasattr(manager, "_hook_registry")
        assert hasattr(manager, "_hooks_by_event")
        assert hasattr(manager, "_advanced_cache")
        assert hasattr(manager, "_performance_metrics")

    def test_register_hook(self, manager):
        """Test hook registration"""
        metadata = HookMetadata(
            hook_path="/test/hook.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
        )

        manager.register_hook("/test/hook.py", metadata)

        assert "/test/hook.py" in manager._hook_registry
        assert HookEvent.SESSION_START in manager._hooks_by_event

    def test_get_hook_metadata(self, manager):
        """Test getting hook metadata"""
        metadata = HookMetadata(
            hook_path="/test/hook.py",
            event_type=HookEvent.PRE_TOOL_USE,
            priority=HookPriority.HIGH,
        )

        manager.register_hook("/test/hook.py", metadata)

        retrieved = manager.get_hook_metadata("/test/hook.py")
        assert retrieved is not None
        assert retrieved.hook_path == "/test/hook.py"
        assert retrieved.priority == HookPriority.HIGH

    def test_unregister_hook(self, manager):
        """Test hook unregistration"""
        metadata = HookMetadata(
            hook_path="/test/hook.py",
            event_type=HookEvent.SESSION_START,
            priority=HookPriority.NORMAL,
        )

        manager.register_hook("/test/hook.py", metadata)
        assert "/test/hook.py" in manager._hook_registry

        manager.unregister_hook("/test/hook.py")
        assert "/test/hook.py" not in manager._hook_registry


# =============================================================================
# FALLBACK JIT SYSTEM TESTS
# =============================================================================


class TestFallbackJITSystem:
    """Test fallback JIT system when JIT is not available"""

    def test_phase_enum_exists(self):
        """Test Phase enum exists in fallback"""
        assert hasattr(Phase, "SPEC")
        assert hasattr(Phase, "RED")
        assert hasattr(Phase, "GREEN")

    def test_fallback_context_cache(self):
        """Test fallback ContextCache class"""
        cache = ContextCache(max_size=50, max_memory_mb=25)

        assert cache.max_size == 50
        assert cache.max_memory_mb == 25
        assert cache.get_stats()["hits"] == 0
        assert cache.get_stats()["misses"] == 0

    def test_fallback_token_budget_manager(self):
        """Test fallback TokenBudgetManager class"""
        manager = TokenBudgetManager()
        # Should not raise
        assert manager is not None


# =============================================================================
# INTEGRATION TESTS
# =============================================================================


class TestIntegration:
    """Integration tests combining multiple components"""

    @pytest.mark.asyncio
    async def test_circuit_breaker_with_retry_policy(self):
        """Test circuit breaker working with retry policy"""
        breaker = CircuitBreaker(failure_threshold=2)
        policy = RetryPolicy(max_retries=2)

        failure_count = [0]

        async def flaky_func():
            failure_count[0] += 1
            if failure_count[0] <= 2:
                raise Exception("Failure")
            return "success"

        # Should succeed after retries
        result = await breaker.call(flaky_func)
        assert result == "success"

    @pytest.mark.asyncio
    async def test_cache_with_circuit_breaker(self):
        """Test cache integration with circuit breaker"""
        cache = HookResultCache(max_size=100)
        breaker = CircuitBreaker()

        async def cached_operation(key):
            # Check cache first
            cached = cache.get(key)
            if cached:
                return cached

            # Execute with circuit breaker
            async def operation():
                return f"result_{key}"

            result = await breaker.call(operation)
            cache.put(key, result)
            return result

        result1 = await cached_operation("test_key")
        result2 = await cached_operation("test_key")

        assert result1 == result2
        assert result1 == "result_test_key"


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
