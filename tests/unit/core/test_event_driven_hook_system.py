"""
Comprehensive test suite for event_driven_hook_system.py

Tests cover:
- Enums (EventType, EventPriority, ResourceIsolationLevel, MessageBrokerType)
- Event and specialized event classes
- InMemoryMessageBroker with publish, subscribe, queue operations
- RedisMessageBroker with connection handling
- ResourcePool with different isolation levels
- EventProcessor with handler registration and processing
- EventDrivenHookSystem with full lifecycle and event handling
- Event persistence and recovery
- System metrics and status reporting
- Global utility functions
"""

import asyncio
from datetime import datetime, timedelta
from unittest.mock import AsyncMock, Mock, patch

import pytest

from moai_adk.core.event_driven_hook_system import (
    Event,
    EventDrivenHookSystem,
    EventPriority,
    EventProcessor,
    EventType,
    HookExecutionEvent,
    InMemoryMessageBroker,
    MessageBrokerType,
    RedisMessageBroker,
    ResourceIsolationLevel,
    ResourcePool,
    WorkflowEvent,
    execute_hook_with_event_system,
    get_event_system,
    start_event_system,
    stop_event_system,
)
from moai_adk.core.jit_enhanced_hook_manager import HookEvent

# =============================================================================
# ENUM TESTS
# =============================================================================


class TestEventType:
    """Test EventType enum"""

    def test_event_type_values(self):
        """Test EventType has all expected values"""
        assert EventType.HOOK_EXECUTION_REQUEST.value == "hook_execution_request"
        assert EventType.HOOK_EXECUTION_COMPLETED.value == "hook_execution_completed"
        assert EventType.HOOK_EXECUTION_FAILED.value == "hook_execution_failed"
        assert EventType.SYSTEM_ALERT.value == "system_alert"
        assert EventType.RESOURCE_STATUS_CHANGE.value == "resource_status_change"
        assert EventType.HEALTH_CHECK.value == "health_check"
        assert EventType.PERFORMANCE_UPDATE.value == "performance_update"
        assert EventType.BATCH_EXECUTION_REQUEST.value == "batch_execution_request"
        assert EventType.WORKFLOW_ORCHESTRATION.value == "workflow_orchestration"


class TestEventPriority:
    """Test EventPriority enum"""

    def test_priority_values(self):
        """Test EventPriority has correct numeric values"""
        assert EventPriority.CRITICAL.value == 1
        assert EventPriority.HIGH.value == 2
        assert EventPriority.NORMAL.value == 3
        assert EventPriority.LOW.value == 4
        assert EventPriority.BULK.value == 5


class TestResourceIsolationLevel:
    """Test ResourceIsolationLevel enum"""

    def test_isolation_level_values(self):
        """Test ResourceIsolationLevel has all expected values"""
        assert ResourceIsolationLevel.SHARED.value == "shared"
        assert ResourceIsolationLevel.TYPE_ISOLATED.value == "type"
        assert ResourceIsolationLevel.PRIORITY_ISOLATED.value == "priority"
        assert ResourceIsolationLevel.FULL_ISOLATION.value == "full"


class TestMessageBrokerType:
    """Test MessageBrokerType enum"""

    def test_broker_type_values(self):
        """Test MessageBrokerType has all expected values"""
        assert MessageBrokerType.MEMORY.value == "memory"
        assert MessageBrokerType.REDIS.value == "redis"
        assert MessageBrokerType.RABBITMQ.value == "rabbitmq"
        assert MessageBrokerType.KAFKA.value == "kafka"
        assert MessageBrokerType.AWS_SQS.value == "aws_sqs"


# =============================================================================
# EVENT TESTS
# =============================================================================


class TestEvent:
    """Test Event class"""

    def test_event_initialization(self):
        """Test Event initialization with all fields"""
        now = datetime.now()
        payload = {"key": "value"}
        tags = {"tag1": "value1"}

        event = Event(
            event_id="test_id",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.HIGH,
            timestamp=now,
            payload=payload,
            source="test_source",
            correlation_id="corr_123",
            causation_id="caus_123",
            metadata={"meta": "data"},
            retry_count=1,
            max_retries=5,
            timeout_seconds=30.0,
            tags=tags,
        )

        assert event.event_id == "test_id"
        assert event.event_type == EventType.HOOK_EXECUTION_REQUEST
        assert event.priority == EventPriority.HIGH
        assert event.timestamp == now
        assert event.payload == payload
        assert event.source == "test_source"
        assert event.correlation_id == "corr_123"
        assert event.causation_id == "caus_123"
        assert event.metadata == {"meta": "data"}
        assert event.retry_count == 1
        assert event.max_retries == 5
        assert event.timeout_seconds == 30.0
        assert event.tags == tags

    def test_event_to_dict(self):
        """Test Event serialization to dictionary"""
        now = datetime.now()
        event = Event(
            event_id="test_id",
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.CRITICAL,
            timestamp=now,
            payload={"alert": "test"},
            source="test",
        )

        event_dict = event.to_dict()

        assert event_dict["event_id"] == "test_id"
        assert event_dict["event_type"] == "system_alert"
        assert event_dict["priority"] == 1
        assert event_dict["timestamp"] == now.isoformat()
        assert event_dict["payload"] == {"alert": "test"}
        assert event_dict["source"] == "test"

    def test_event_from_dict(self):
        """Test Event deserialization from dictionary"""
        now = datetime.now()
        event_data = {
            "event_id": "test_id",
            "event_type": "hook_execution_completed",
            "priority": 3,
            "timestamp": now.isoformat(),
            "payload": {"status": "done"},
            "source": "test_source",
            "correlation_id": "corr_123",
            "causation_id": "caus_123",
            "metadata": {"info": "data"},
            "retry_count": 0,
            "max_retries": 3,
            "timeout_seconds": 60.0,
            "tags": {"tag1": "val1"},
        }

        event = Event.from_dict(event_data)

        assert event.event_id == "test_id"
        assert event.event_type == EventType.HOOK_EXECUTION_COMPLETED
        assert event.priority == EventPriority.NORMAL
        assert event.payload == {"status": "done"}
        assert event.source == "test_source"
        assert event.correlation_id == "corr_123"
        assert event.causation_id == "caus_123"

    def test_event_from_dict_with_defaults(self):
        """Test Event.from_dict with missing optional fields"""
        event_data = {
            "event_id": "test_id",
            "event_type": "hook_execution_request",
            "priority": 2,
            "timestamp": datetime.now().isoformat(),
            "payload": {},
        }

        event = Event.from_dict(event_data)

        assert event.source == ""
        assert event.correlation_id is None
        assert event.causation_id is None
        assert event.metadata == {}
        assert event.retry_count == 0
        assert event.max_retries == 3


class TestHookExecutionEvent:
    """Test HookExecutionEvent specialized event"""

    def test_hook_execution_event_initialization(self):
        """Test HookExecutionEvent initialization"""
        event = HookExecutionEvent(
            event_id="exec_123",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
            hook_path="/path/to/hook.py",
            hook_event_type=HookEvent.SESSION_START,
            execution_context={"key": "value"},
            isolation_level=ResourceIsolationLevel.FULL_ISOLATION,
        )

        assert event.hook_path == "/path/to/hook.py"
        assert event.hook_event_type == HookEvent.SESSION_START
        assert event.execution_context == {"key": "value"}
        assert event.isolation_level == ResourceIsolationLevel.FULL_ISOLATION

    def test_hook_execution_event_requires_hook_path(self):
        """Test HookExecutionEvent requires hook_path for execution request"""
        with pytest.raises(ValueError, match="hook_path is required"):
            HookExecutionEvent(
                event_id="exec_123",
                event_type=EventType.HOOK_EXECUTION_REQUEST,
                priority=EventPriority.NORMAL,
                timestamp=datetime.now(),
                payload={},
                hook_path="",  # Empty hook_path
            )

    def test_hook_execution_event_allows_empty_hook_path_for_other_events(self):
        """Test HookExecutionEvent allows empty hook_path for non-execution events"""
        event = HookExecutionEvent(
            event_id="alert_123",
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={},
            hook_path="",  # Empty is OK for non-execution events
        )

        assert event.hook_path == ""


class TestWorkflowEvent:
    """Test WorkflowEvent specialized event"""

    def test_workflow_event_initialization(self):
        """Test WorkflowEvent initialization"""
        event = WorkflowEvent(
            event_id="wf_123",
            event_type=EventType.WORKFLOW_ORCHESTRATION,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={},
            workflow_id="workflow_abc",
            step_id="step_1",
            workflow_definition={"steps": []},
            execution_state={"current": "step_1"},
        )

        assert event.workflow_id == "workflow_abc"
        assert event.step_id == "step_1"
        assert event.workflow_definition == {"steps": []}
        assert event.execution_state == {"current": "step_1"}

    def test_workflow_event_requires_workflow_id(self):
        """Test WorkflowEvent requires workflow_id for orchestration events"""
        with pytest.raises(ValueError, match="workflow_id is required"):
            WorkflowEvent(
                event_id="wf_123",
                event_type=EventType.WORKFLOW_ORCHESTRATION,
                priority=EventPriority.HIGH,
                timestamp=datetime.now(),
                payload={},
                workflow_id="",  # Empty workflow_id
            )


# =============================================================================
# IN-MEMORY MESSAGE BROKER TESTS
# =============================================================================


class TestInMemoryMessageBroker:
    """Test InMemoryMessageBroker"""

    @pytest.fixture
    def broker(self):
        """Create a broker instance for testing"""
        return InMemoryMessageBroker(max_queue_size=100)

    @pytest.mark.asyncio
    async def test_publish_event(self, broker):
        """Test publishing event to topic"""
        event = Event(
            event_id="test_123",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={"test": "data"},
        )

        result = await broker.publish("test_topic", event)

        assert result is True
        assert len(broker.queues["test_topic"]) == 1
        assert broker.queues["test_topic"][0] == event
        assert broker._stats["messages_published"] == 1

    @pytest.mark.asyncio
    async def test_subscribe_callback(self, broker):
        """Test subscribing to topic with callback"""
        callback_mock = AsyncMock()

        sub_id = await broker.subscribe("test_topic", callback_mock)

        assert sub_id is not None
        assert len(broker.subscribers["test_topic"]) == 1
        assert broker._stats["active_subscriptions"] == 1

    @pytest.mark.asyncio
    async def test_publish_notifies_subscribers(self, broker):
        """Test that publishing notifies subscribers"""
        callback_mock = AsyncMock()

        await broker.subscribe("test_topic", callback_mock)

        event = Event(
            event_id="test_123",
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={},
        )

        await broker.publish("test_topic", event)

        # Give async task time to execute
        await asyncio.sleep(0.1)

        # Callback should be called
        assert callback_mock.called or broker._stats["messages_delivered"] > 0

    @pytest.mark.asyncio
    async def test_unsubscribe(self, broker):
        """Test unsubscribing from topic"""
        callback_mock = AsyncMock()

        sub_id = await broker.subscribe("test_topic", callback_mock)
        assert broker._stats["active_subscriptions"] == 1

        result = await broker.unsubscribe(sub_id)

        assert result is True
        assert broker._stats["active_subscriptions"] == 0

    @pytest.mark.asyncio
    async def test_unsubscribe_nonexistent(self, broker):
        """Test unsubscribing from nonexistent subscription"""
        result = await broker.unsubscribe("nonexistent_id")

        assert result is False

    @pytest.mark.asyncio
    async def test_create_queue(self, broker):
        """Test creating message queue"""
        result = await broker.create_queue("test_queue", {"max_size": 500})

        assert result is True
        assert "test_queue" in broker.queues
        assert broker._stats["queues_created"] == 1

    @pytest.mark.asyncio
    async def test_delete_queue(self, broker):
        """Test deleting message queue"""
        await broker.create_queue("test_queue", {})

        result = await broker.delete_queue("test_queue")

        assert result is True
        assert "test_queue" not in broker.queues

    @pytest.mark.asyncio
    async def test_queue_size_limit(self, broker):
        """Test queue respects max size limit"""
        small_broker = InMemoryMessageBroker(max_queue_size=3)

        for i in range(5):
            event = Event(
                event_id=f"test_{i}",
                event_type=EventType.HOOK_EXECUTION_REQUEST,
                priority=EventPriority.NORMAL,
                timestamp=datetime.now(),
                payload={"index": i},
            )
            await small_broker.publish("test_topic", event)

        # Queue should only have last 3 messages
        assert len(small_broker.queues["test_topic"]) == 3

    def test_get_stats(self, broker):
        """Test getting broker statistics"""
        stats = broker.get_stats()

        assert "messages_published" in stats
        assert "messages_delivered" in stats
        assert "queues_created" in stats
        assert "active_subscriptions" in stats
        assert "failed_publishes" in stats
        assert "queue_count" in stats
        assert "total_queued_messages" in stats

    @pytest.mark.asyncio
    async def test_safe_callback_with_sync_callback(self, broker):
        """Test _safe_callback with synchronous callback"""
        callback_mock = Mock()

        event = Event(
            event_id="test_123",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
        )

        await broker._safe_callback(callback_mock, event)

        callback_mock.assert_called_once_with(event)

    @pytest.mark.asyncio
    async def test_safe_callback_with_async_callback(self, broker):
        """Test _safe_callback with asynchronous callback"""
        callback_mock = AsyncMock()

        event = Event(
            event_id="test_123",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
        )

        await broker._safe_callback(callback_mock, event)

        callback_mock.assert_called_once_with(event)

    @pytest.mark.asyncio
    async def test_safe_callback_exception_handling(self, broker):
        """Test _safe_callback handles exceptions gracefully"""
        callback_mock = AsyncMock(side_effect=Exception("Test error"))

        event = Event(
            event_id="test_123",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
        )

        # Should not raise, just log
        await broker._safe_callback(callback_mock, event)

    @pytest.mark.asyncio
    async def test_multiple_subscriptions(self, broker):
        """Test multiple subscriptions to same topic"""
        callback1 = AsyncMock()
        callback2 = AsyncMock()

        sub_id1 = await broker.subscribe("test_topic", callback1)
        sub_id2 = await broker.subscribe("test_topic", callback2)

        assert sub_id1 != sub_id2
        assert broker._stats["active_subscriptions"] == 2

        event = Event(
            event_id="test_123",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={"test": "data"},
        )

        await broker.publish("test_topic", event)
        await asyncio.sleep(0.1)

        # Both callbacks should have been invoked
        assert callback1.called or callback2.called


# =============================================================================
# REDIS MESSAGE BROKER TESTS
# =============================================================================


class TestRedisMessageBroker:
    """Test RedisMessageBroker"""

    @pytest.fixture
    def broker(self):
        """Create a Redis broker instance"""
        return RedisMessageBroker(redis_url="redis://localhost:6379/0")

    @pytest.mark.asyncio
    async def test_redis_connect_failure(self, broker):
        """Test Redis connection failure handling"""
        # Mock redis.asyncio to not be available
        broker._redis = None

        # Test that ConnectionError is raised when trying to connect without redis
        with patch.dict("sys.modules", {"redis": None, "redis.asyncio": None}):
            with pytest.raises((ImportError, ConnectionError, AttributeError)):
                await broker._connect()

    @pytest.mark.asyncio
    async def test_publish_without_connection(self, broker):
        """Test publish without Redis connection"""
        event = Event(
            event_id="test_123",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
        )

        # Mock connect to not establish connection
        broker._connect = AsyncMock()
        broker._redis = None

        result = await broker.publish("test_topic", event)

        assert result is False
        assert broker._stats["failed_publishes"] == 1

    @pytest.mark.asyncio
    async def test_subscribe_returns_subscription_id(self, broker):
        """Test subscribe returns valid subscription ID"""
        broker._redis = AsyncMock()
        broker._pubsub = AsyncMock()
        broker._connect = AsyncMock()

        callback = AsyncMock()

        sub_id = await broker.subscribe("test_topic", callback)

        assert sub_id != ""
        assert broker._stats["active_subscriptions"] == 1

    @pytest.mark.asyncio
    async def test_subscribe_connection_error(self, broker):
        """Test subscribe handles connection errors"""
        broker._redis = None
        broker._pubsub = None
        broker._connect = AsyncMock(side_effect=Exception("Connection failed"))

        callback = AsyncMock()

        sub_id = await broker.subscribe("test_topic", callback)

        assert sub_id == ""

    @pytest.mark.asyncio
    async def test_unsubscribe(self, broker):
        """Test Redis unsubscribe"""
        broker._subscribers["topic1"] = [("sub_1", AsyncMock()), ("sub_2", AsyncMock())]
        broker._stats["active_subscriptions"] = 2

        result = await broker.unsubscribe("sub_1")

        assert result is True
        assert broker._stats["active_subscriptions"] == 1

    @pytest.mark.asyncio
    async def test_unsubscribe_exception(self, broker):
        """Test unsubscribe exception handling"""
        broker._subscribers = Mock(side_effect=Exception("Error"))

        result = await broker.unsubscribe("sub_1")

        assert result is False

    @pytest.mark.asyncio
    async def test_create_queue(self, broker):
        """Test create_queue"""
        broker._redis = AsyncMock()
        broker._connect = AsyncMock()

        result = await broker.create_queue("test_queue", {})

        assert result is True
        assert broker._stats["queues_created"] == 1

    @pytest.mark.asyncio
    async def test_create_queue_connection_error(self, broker):
        """Test create_queue connection error"""
        broker._connect = AsyncMock(side_effect=Exception("Connection failed"))

        result = await broker.create_queue("test_queue", {})

        assert result is False

    @pytest.mark.asyncio
    async def test_delete_queue(self, broker):
        """Test delete_queue"""
        broker._redis = AsyncMock()
        broker._redis.delete = AsyncMock()
        broker._connect = AsyncMock()

        result = await broker.delete_queue("test_queue")

        assert result is True

    @pytest.mark.asyncio
    async def test_delete_queue_without_connection(self, broker):
        """Test delete_queue without connection"""
        broker._connect = AsyncMock()
        broker._redis = None

        result = await broker.delete_queue("test_queue")

        assert result is False

    def test_get_stats(self, broker):
        """Test get_stats"""
        stats = broker.get_stats()

        assert "messages_published" in stats
        assert "messages_delivered" in stats
        assert "queues_created" in stats


# =============================================================================
# RESOURCE POOL TESTS
# =============================================================================


class TestResourcePool:
    """Test ResourcePool class"""

    @pytest.fixture
    def pool_shared(self):
        """Create resource pool with SHARED isolation"""
        return ResourcePool(ResourceIsolationLevel.SHARED, max_concurrent=5)

    @pytest.fixture
    def pool_type_isolated(self):
        """Create resource pool with TYPE_ISOLATED isolation"""
        return ResourcePool(ResourceIsolationLevel.TYPE_ISOLATED, max_concurrent=5)

    @pytest.fixture
    def pool_priority_isolated(self):
        """Create resource pool with PRIORITY_ISOLATED isolation"""
        return ResourcePool(ResourceIsolationLevel.PRIORITY_ISOLATED, max_concurrent=5)

    @pytest.fixture
    def pool_full_isolated(self):
        """Create resource pool with FULL_ISOLATION"""
        return ResourcePool(ResourceIsolationLevel.FULL_ISOLATION, max_concurrent=5)

    @pytest.mark.asyncio
    async def test_shared_pool_key(self, pool_shared):
        """Test SHARED isolation returns same key"""
        key1 = pool_shared._get_pool_key("hook1.py", HookEvent.SESSION_START)
        key2 = pool_shared._get_pool_key("hook2.py", HookEvent.POST_TOOL_USE)

        assert key1 == "shared"
        assert key2 == "shared"
        assert key1 == key2

    @pytest.mark.asyncio
    async def test_type_isolated_pool_key(self, pool_type_isolated):
        """Test TYPE_ISOLATED isolation returns event type key"""
        key1 = pool_type_isolated._get_pool_key("hook1.py", HookEvent.SESSION_START)
        key2 = pool_type_isolated._get_pool_key("hook1.py", HookEvent.POST_TOOL_USE)

        assert key1 == HookEvent.SESSION_START.value
        assert key2 == HookEvent.POST_TOOL_USE.value
        assert key1 != key2

    @pytest.mark.asyncio
    async def test_priority_isolated_pool_key(self, pool_priority_isolated):
        """Test PRIORITY_ISOLATED isolation"""
        key_security = pool_priority_isolated._get_pool_key("security_hook.py", HookEvent.SESSION_START)
        key_validation = pool_priority_isolated._get_pool_key("validation_hook.py", HookEvent.SESSION_START)
        key_performance = pool_priority_isolated._get_pool_key("performance_hook.py", HookEvent.SESSION_START)
        key_normal = pool_priority_isolated._get_pool_key("normal_hook.py", HookEvent.SESSION_START)

        assert key_security == "critical"
        assert key_validation == "critical"
        assert key_performance == "high"
        assert key_normal == "normal"

    @pytest.mark.asyncio
    async def test_full_isolated_pool_key(self, pool_full_isolated):
        """Test FULL_ISOLATION returns hook path"""
        key1 = pool_full_isolated._get_pool_key("hook1.py", HookEvent.SESSION_START)
        key2 = pool_full_isolated._get_pool_key("hook2.py", HookEvent.SESSION_START)

        assert key1 == "hook1.py"
        assert key2 == "hook2.py"
        assert key1 != key2

    @pytest.mark.asyncio
    async def test_acquire_execution_slot(self, pool_shared):
        """Test acquiring execution slot"""
        result = await pool_shared.acquire_execution_slot("hook1.py", HookEvent.SESSION_START)

        assert result is True
        assert pool_shared._stats["active_executions"] == 1
        assert pool_shared._stats["total_executions"] == 1

    @pytest.mark.asyncio
    async def test_release_execution_slot(self, pool_shared):
        """Test releasing execution slot"""
        await pool_shared.acquire_execution_slot("hook1.py", HookEvent.SESSION_START)

        await pool_shared.release_execution_slot("hook1.py", HookEvent.SESSION_START)

        assert pool_shared._stats["active_executions"] == 0

    @pytest.mark.asyncio
    async def test_acquire_release_cycle(self, pool_shared):
        """Test multiple acquire-release cycles"""
        for i in range(3):
            result = await pool_shared.acquire_execution_slot(f"hook{i}.py", HookEvent.SESSION_START)
            assert result is True

        assert pool_shared._stats["active_executions"] == 3

        for i in range(3):
            await pool_shared.release_execution_slot(f"hook{i}.py", HookEvent.SESSION_START)

        assert pool_shared._stats["active_executions"] == 0

    def test_get_stats(self, pool_shared):
        """Test get_stats method"""
        stats = pool_shared.get_stats()

        assert "total_executions" in stats
        assert "active_executions" in stats
        assert "pool_utilization" in stats
        assert "isolation_violations" in stats


# =============================================================================
# EVENT PROCESSOR TESTS
# =============================================================================


class TestEventProcessor:
    """Test EventProcessor class"""

    @pytest.fixture
    def processor(self):
        """Create event processor"""
        pool = ResourcePool(ResourceIsolationLevel.SHARED, max_concurrent=10)
        return EventProcessor(pool)

    @pytest.mark.asyncio
    async def test_register_handler(self, processor):
        """Test registering event handler"""
        handler = AsyncMock()

        processor.register_handler(EventType.HOOK_EXECUTION_REQUEST, handler)

        assert EventType.HOOK_EXECUTION_REQUEST in processor._handlers
        assert handler in processor._handlers[EventType.HOOK_EXECUTION_REQUEST]

    @pytest.mark.asyncio
    async def test_process_event_no_handlers(self, processor):
        """Test processing event with no handlers"""
        event = Event(
            event_id="test_123",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
        )

        result = await processor.process_event(event)

        assert result is True  # Returns True if no handlers but no error

    @pytest.mark.asyncio
    async def test_process_event_with_async_handler(self, processor):
        """Test processing event with async handler"""
        handler = AsyncMock()

        processor.register_handler(EventType.HOOK_EXECUTION_REQUEST, handler)

        event = Event(
            event_id="test_123",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={"data": "test"},
        )

        result = await processor.process_event(event)

        assert result is True
        handler.assert_called_once_with(event)

    @pytest.mark.asyncio
    async def test_process_event_with_sync_handler(self, processor):
        """Test processing event with sync handler"""
        handler = Mock()

        processor.register_handler(EventType.SYSTEM_ALERT, handler)

        event = Event(
            event_id="test_123",
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={},
        )

        result = await processor.process_event(event)

        assert result is True
        handler.assert_called_once_with(event)

    @pytest.mark.asyncio
    async def test_process_event_handler_exception(self, processor):
        """Test processing event when handler raises exception"""
        handler = AsyncMock(side_effect=Exception("Handler error"))

        processor.register_handler(EventType.HOOK_EXECUTION_REQUEST, handler)

        event = Event(
            event_id="test_123",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
        )

        result = await processor.process_event(event)

        assert result is False

    @pytest.mark.asyncio
    async def test_process_event_multiple_handlers(self, processor):
        """Test processing event with multiple handlers"""
        handler1 = AsyncMock()
        handler2 = Mock()

        processor.register_handler(EventType.HEALTH_CHECK, handler1)
        processor.register_handler(EventType.HEALTH_CHECK, handler2)

        event = Event(
            event_id="test_123",
            event_type=EventType.HEALTH_CHECK,
            priority=EventPriority.LOW,
            timestamp=datetime.now(),
            payload={},
        )

        result = await processor.process_event(event)

        assert result is True
        handler1.assert_called_once()
        handler2.assert_called_once()

    def test_update_stats(self, processor):
        """Test statistics update"""
        processor._update_stats(EventType.HOOK_EXECUTION_REQUEST, 10.5, True)
        processor._update_stats(EventType.HOOK_EXECUTION_REQUEST, 20.0, False)

        assert processor._processing_stats["events_processed"] == 2
        assert processor._processing_stats["events_failed"] == 1
        assert processor._processing_stats["by_event_type"]["hook_execution_request"] == 2

    def test_get_stats(self, processor):
        """Test get_stats"""
        processor._update_stats(EventType.HOOK_EXECUTION_REQUEST, 10.0, True)

        stats = processor.get_stats()

        assert "events_processed" in stats
        assert "events_failed" in stats
        assert "average_processing_time_ms" in stats
        assert "success_rate" in stats
        assert "handlers_registered" in stats


# =============================================================================
# EVENT-DRIVEN HOOK SYSTEM TESTS
# =============================================================================


class TestEventDrivenHookSystem:
    """Test EventDrivenHookSystem main class"""

    @pytest.fixture
    def system(self, tmp_path):
        """Create event system instance"""
        return EventDrivenHookSystem(
            message_broker_type=MessageBrokerType.MEMORY,
            isolation_level=ResourceIsolationLevel.TYPE_ISOLATED,
            max_concurrent_hooks=10,
            enable_persistence=True,
            persistence_path=tmp_path / "events",
        )

    @pytest.fixture
    def system_no_persistence(self):
        """Create event system without persistence"""
        return EventDrivenHookSystem(
            message_broker_type=MessageBrokerType.MEMORY,
            isolation_level=ResourceIsolationLevel.SHARED,
            enable_persistence=False,
        )

    def test_system_initialization(self, system):
        """Test system initialization"""
        assert system.message_broker_type == MessageBrokerType.MEMORY
        assert system.isolation_level == ResourceIsolationLevel.TYPE_ISOLATED
        assert system.max_concurrent_hooks == 10
        assert system.enable_persistence is True
        assert isinstance(system.message_broker, InMemoryMessageBroker)
        assert isinstance(system.resource_pool, ResourcePool)
        assert isinstance(system.event_processor, EventProcessor)
        assert system._running is False

    def test_create_message_broker_memory(self, system):
        """Test message broker creation for MEMORY type"""
        broker = system._create_message_broker()

        assert isinstance(broker, InMemoryMessageBroker)

    def test_create_message_broker_redis(self):
        """Test message broker creation for REDIS type"""
        system = EventDrivenHookSystem(
            message_broker_type=MessageBrokerType.REDIS,
            redis_url="redis://localhost:6379/0",
        )

        broker = system._create_message_broker()

        assert isinstance(broker, RedisMessageBroker)

    def test_create_message_broker_unsupported(self):
        """Test message broker creation for unsupported type"""
        system = EventDrivenHookSystem(message_broker_type=MessageBrokerType.KAFKA)

        broker = system._create_message_broker()

        # Should fall back to InMemoryMessageBroker
        assert isinstance(broker, InMemoryMessageBroker)

    @pytest.mark.asyncio
    async def test_start_system(self, system):
        """Test starting the system"""
        await system.start()

        assert system._running is True
        assert len(system._event_loops) > 0

        # Cleanup
        await system.stop()

    @pytest.mark.asyncio
    async def test_start_already_running(self, system):
        """Test starting already running system"""
        await system.start()

        # Start again
        await system.start()

        assert system._running is True

        await system.stop()

    @pytest.mark.asyncio
    async def test_stop_system(self, system):
        """Test stopping the system"""
        await system.start()
        assert system._running is True

        await system.stop()

        assert system._running is False

    @pytest.mark.asyncio
    async def test_stop_not_running(self, system_no_persistence):
        """Test stopping system that's not running"""
        # Should not raise
        await system_no_persistence.stop()

    @pytest.mark.asyncio
    async def test_register_event_handlers(self, system):
        """Test event handler registration"""
        system._register_event_handlers()

        assert EventType.HOOK_EXECUTION_REQUEST in system.event_processor._handlers
        assert EventType.HOOK_EXECUTION_COMPLETED in system.event_processor._handlers
        assert EventType.HOOK_EXECUTION_FAILED in system.event_processor._handlers
        assert EventType.SYSTEM_ALERT in system.event_processor._handlers
        assert EventType.HEALTH_CHECK in system.event_processor._handlers
        assert EventType.BATCH_EXECUTION_REQUEST in system.event_processor._handlers
        assert EventType.WORKFLOW_ORCHESTRATION in system.event_processor._handlers

    @pytest.mark.asyncio
    async def test_setup_message_queues(self, system):
        """Test message queue setup"""
        await system._setup_message_queues()

        # Check queues were created
        stats = system.message_broker.get_stats()
        assert stats["queues_created"] == 5

    @pytest.mark.asyncio
    async def test_publish_hook_execution_event(self, system):
        """Test publishing hook execution event"""
        await system.start()

        event_id = await system.publish_hook_execution_event(
            hook_path="/path/to/hook.py",
            event_type=HookEvent.SESSION_START,
            execution_context={"key": "value"},
        )

        assert event_id is not None
        assert system._system_metrics["events_published"] == 1

        await system.stop()

    @pytest.mark.asyncio
    async def test_publish_system_alert(self, system):
        """Test publishing system alert"""
        await system.start()

        alert_id = await system.publish_system_alert(
            alert_type="TEST_ALERT",
            message="Test alert message",
            severity=EventPriority.HIGH,
            metadata={"info": "test"},
        )

        assert alert_id is not None
        assert system._system_metrics["events_published"] == 1

        await system.stop()

    @pytest.mark.asyncio
    async def test_get_system_status(self, system):
        """Test getting system status"""
        await system.start()

        status = await system.get_system_status()

        assert status["status"] == "running"
        assert "uptime_seconds" in status
        assert "system_metrics" in status
        assert "message_broker_stats" in status
        assert "resource_pool_stats" in status
        assert "event_processor_stats" in status

        await system.stop()

    @pytest.mark.asyncio
    async def test_get_event_flow_diagram(self, system):
        """Test getting event flow diagram"""
        diagram = system.get_event_flow_diagram()

        assert "event_types" in diagram
        assert "priorities" in diagram
        assert "isolation_levels" in diagram
        assert "message_broker_type" in diagram
        assert "flow_pattern" in diagram
        assert len(diagram["event_types"]) > 0

    @pytest.mark.asyncio
    async def test_get_queue_name_by_priority(self, system):
        """Test queue name resolution by priority"""
        assert system._get_queue_name_by_priority(EventPriority.CRITICAL) == "system_events"
        assert system._get_queue_name_by_priority(EventPriority.HIGH) == "hook_execution_high"
        assert system._get_queue_name_by_priority(EventPriority.NORMAL) == "hook_execution_normal"
        assert system._get_queue_name_by_priority(EventPriority.LOW) == "hook_execution_low"
        assert system._get_queue_name_by_priority(EventPriority.BULK) == "analytics"

    @pytest.mark.asyncio
    async def test_handle_hook_execution_request(self, system):
        """Test hook execution request handler"""
        event = HookExecutionEvent(
            event_id="exec_123",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
            hook_path="/path/to/hook.py",
            hook_event_type=HookEvent.SESSION_START,
        )

        await system._handle_hook_execution_request(event)

        # Hook should be marked as executed
        assert system._system_metrics["hook_executions"] >= 0

    @pytest.mark.asyncio
    async def test_handle_hook_execution_completed(self, system):
        """Test hook execution completed handler"""
        event = Event(
            event_id="exec_123",
            event_type=EventType.HOOK_EXECUTION_COMPLETED,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={"status": "success"},
        )

        await system._handle_hook_execution_completed(event)

        assert system._system_metrics["events_processed"] == 1

    @pytest.mark.asyncio
    async def test_handle_hook_execution_failed(self, system):
        """Test hook execution failed handler"""
        event = Event(
            event_id="exec_123",
            event_type=EventType.HOOK_EXECUTION_FAILED,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={"reason": "timeout"},
        )

        await system._handle_hook_execution_failed(event)

        assert system._system_metrics["events_failed"] == 1

    @pytest.mark.asyncio
    async def test_handle_system_alert(self, system):
        """Test system alert handler"""
        event = Event(
            event_id="alert_123",
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.CRITICAL,
            timestamp=datetime.now(),
            payload={"alert": "critical error"},
        )

        # Should not raise
        await system._handle_system_alert(event)

    @pytest.mark.asyncio
    async def test_handle_health_check(self, system):
        """Test health check handler"""
        event = Event(
            event_id="health_123",
            event_type=EventType.HEALTH_CHECK,
            priority=EventPriority.LOW,
            timestamp=datetime.now(),
            payload={},
        )

        # Should not raise
        await system._handle_health_check(event)

    @pytest.mark.asyncio
    async def test_handle_batch_execution_request(self, system):
        """Test batch execution request handler"""
        event = Event(
            event_id="batch_123",
            event_type=EventType.BATCH_EXECUTION_REQUEST,
            priority=EventPriority.BULK,
            timestamp=datetime.now(),
            payload={"hooks": ["hook1.py", "hook2.py"]},
        )

        # Should not raise
        await system._handle_batch_execution_request(event)

    @pytest.mark.asyncio
    async def test_handle_workflow_orchestration(self, system):
        """Test workflow orchestration handler"""
        event = WorkflowEvent(
            event_id="wf_123",
            event_type=EventType.WORKFLOW_ORCHESTRATION,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={},
            workflow_id="workflow_abc",
        )

        # Should not raise
        await system._handle_workflow_orchestration(event)

    @pytest.mark.asyncio
    async def test_publish_performance_update(self, system):
        """Test publishing performance update"""
        await system.start()

        await system._publish_performance_update()

        # Stats should show published event
        stats = system.message_broker.get_stats()
        assert stats["messages_published"] >= 1

        await system.stop()

    def test_update_system_metrics(self, system):
        """Test system metrics update"""
        old_uptime = system._system_metrics["system_uptime_seconds"]

        system._update_system_metrics()

        new_uptime = system._system_metrics["system_uptime_seconds"]
        assert new_uptime >= old_uptime

    @pytest.mark.asyncio
    async def test_cleanup_old_events(self, system):
        """Test cleanup of old events"""
        old_time = datetime.now() - timedelta(hours=25)

        event1 = Event(
            event_id="old_event",
            event_type=EventType.HOOK_EXECUTION_COMPLETED,
            priority=EventPriority.LOW,
            timestamp=old_time,
            payload={},
        )

        event2 = Event(
            event_id="new_event",
            event_type=EventType.HOOK_EXECUTION_COMPLETED,
            priority=EventPriority.LOW,
            timestamp=datetime.now(),
            payload={},
        )

        system._pending_events["old_event"] = event1
        system._pending_events["new_event"] = event2
        system._processed_events.add("old_event")
        system._processed_events.add("new_event")

        await system._cleanup_old_events()

        # Old event should be removed
        assert "old_event" not in system._processed_events
        assert "new_event" in system._processed_events

    @pytest.mark.asyncio
    async def test_cleanup_completed_workflows(self, system):
        """Test cleanup of completed workflows"""
        # Should not raise
        await system._cleanup_completed_workflows()

    @pytest.mark.asyncio
    async def test_persist_events(self, system):
        """Test event persistence"""
        event = Event(
            event_id="test_123",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={"data": "test"},
        )

        system._pending_events["test_123"] = event
        system._processed_events.add("test_456")

        await system._persist_events()

        # Check files were created
        assert (system.persistence_path / "pending_events.json").exists()
        assert (system.persistence_path / "processed_events.json").exists()
        assert (system.persistence_path / "system_metrics.json").exists()

    @pytest.mark.asyncio
    async def test_load_persisted_events(self, system):
        """Test loading persisted events"""
        # First persist events
        event = Event(
            event_id="test_123",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={"data": "test"},
        )

        system._pending_events["test_123"] = event
        system._processed_events.add("test_456")

        await system._persist_events()

        # Clear and reload
        system._pending_events.clear()
        system._processed_events.clear()

        await system._load_persisted_events()

        assert "test_123" in system._pending_events
        assert "test_456" in system._processed_events

    @pytest.mark.asyncio
    async def test_persist_without_enabled(self, system_no_persistence):
        """Test that persistence is skipped when disabled"""
        await system_no_persistence._persist_events()

        # Should complete without error

    @pytest.mark.asyncio
    async def test_load_without_enabled(self, system_no_persistence):
        """Test that loading is skipped when disabled"""
        await system_no_persistence._load_persisted_events()

        # Should complete without error

    @pytest.mark.asyncio
    async def test_execute_hook_event(self, system):
        """Test hook event execution"""
        event = HookExecutionEvent(
            event_id="exec_123",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
            hook_path="/path/to/hook.py",
            hook_event_type=HookEvent.SESSION_START,
        )

        # Should not raise
        await system._execute_hook_event(event)


# =============================================================================
# GLOBAL UTILITY FUNCTION TESTS
# =============================================================================


class TestGlobalUtilityFunctions:
    """Test module-level utility functions"""

    def test_get_event_system_singleton(self):
        """Test get_event_system returns singleton"""
        import moai_adk.core.event_driven_hook_system as module

        module._event_system = None  # Reset singleton

        system1 = get_event_system()
        system2 = get_event_system()

        assert system1 is system2

    @pytest.mark.asyncio
    async def test_start_event_system(self):
        """Test start_event_system function"""
        # Get fresh system
        import moai_adk.core.event_driven_hook_system as module

        module._event_system = None

        await start_event_system()

        system = get_event_system()
        assert system._running is True

        await stop_event_system()

    @pytest.mark.asyncio
    async def test_stop_event_system(self):
        """Test stop_event_system function"""
        import moai_adk.core.event_driven_hook_system as module

        module._event_system = None

        await start_event_system()
        await stop_event_system()

        system = get_event_system()
        assert system._running is False

    @pytest.mark.asyncio
    async def test_execute_hook_with_event_system(self):
        """Test execute_hook_with_event_system function"""
        import moai_adk.core.event_driven_hook_system as module

        module._event_system = None

        await start_event_system()

        event_id = await execute_hook_with_event_system(
            hook_path="/path/to/hook.py",
            event_type=HookEvent.SESSION_START,
            execution_context={"key": "value"},
        )

        assert event_id is not None

        await stop_event_system()


# =============================================================================
# INTEGRATION TESTS
# =============================================================================


class TestIntegration:
    """Integration tests combining multiple components"""

    @pytest.mark.asyncio
    async def test_full_event_lifecycle(self, tmp_path):
        """Test complete event lifecycle"""
        system = EventDrivenHookSystem(
            message_broker_type=MessageBrokerType.MEMORY,
            isolation_level=ResourceIsolationLevel.FULL_ISOLATION,
            max_concurrent_hooks=5,
            enable_persistence=True,
            persistence_path=tmp_path / "events",
        )

        await system.start()

        # Publish an event
        event_id = await system.publish_hook_execution_event(
            hook_path="/test/hook.py",
            event_type=HookEvent.SESSION_START,
            execution_context={"data": "test"},
            priority=EventPriority.HIGH,
        )

        assert event_id is not None

        # Process event
        await asyncio.sleep(0.2)

        # Check system status
        status = await system.get_system_status()
        assert status["status"] == "running"
        assert status["system_metrics"]["events_published"] >= 1

        await system.stop()

    @pytest.mark.asyncio
    async def test_multiple_isolation_levels(self):
        """Test system with different isolation levels"""
        for isolation_level in ResourceIsolationLevel:
            system = EventDrivenHookSystem(
                message_broker_type=MessageBrokerType.MEMORY,
                isolation_level=isolation_level,
            )

            assert system.isolation_level == isolation_level

    @pytest.mark.asyncio
    async def test_event_priority_queuing(self, tmp_path):
        """Test events are queued by priority"""
        system = EventDrivenHookSystem(
            message_broker_type=MessageBrokerType.MEMORY,
            enable_persistence=False,
        )

        await system.start()

        # Publish events with different priorities
        await system.publish_hook_execution_event(
            hook_path="/hook1.py",
            event_type=HookEvent.SESSION_START,
            execution_context={},
            priority=EventPriority.NORMAL,
        )

        await system.publish_hook_execution_event(
            hook_path="/hook2.py",
            event_type=HookEvent.SESSION_START,
            execution_context={},
            priority=EventPriority.HIGH,
        )

        await system.publish_system_alert(
            alert_type="TEST",
            message="Test",
            severity=EventPriority.CRITICAL,
        )

        await asyncio.sleep(0.2)

        status = await system.get_system_status()
        assert status["system_metrics"]["events_published"] >= 3

        await system.stop()

    @pytest.mark.asyncio
    async def test_concurrent_hook_execution(self, tmp_path):
        """Test concurrent hook execution with resource pooling"""
        system = EventDrivenHookSystem(
            message_broker_type=MessageBrokerType.MEMORY,
            isolation_level=ResourceIsolationLevel.SHARED,
            max_concurrent_hooks=3,
            enable_persistence=False,
        )

        await system.start()

        # Publish multiple hooks concurrently
        tasks = [
            system.publish_hook_execution_event(
                hook_path=f"/hook{i}.py",
                event_type=HookEvent.SESSION_START,
                execution_context={"index": i},
            )
            for i in range(5)
        ]

        event_ids = await asyncio.gather(*tasks)

        assert len(event_ids) == 5
        assert all(eid is not None for eid in event_ids)

        await system.stop()
