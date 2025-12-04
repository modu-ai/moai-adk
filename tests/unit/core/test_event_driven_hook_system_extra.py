"""Comprehensive test coverage for Event-Driven Hook System.

This module provides extensive unit tests for the EventDrivenHookSystem
including message brokers, event processing, resource isolation, and workflows.
"""

import asyncio
import pytest
import uuid
import json
from datetime import datetime, timedelta
from pathlib import Path
from unittest.mock import MagicMock, AsyncMock, patch, Mock
from collections import defaultdict, deque

from moai_adk.core.event_driven_hook_system import (
    Event,
    EventType,
    EventPriority,
    ResourceIsolationLevel,
    MessageBrokerType,
    HookExecutionEvent,
    WorkflowEvent,
    MessageBroker,
    InMemoryMessageBroker,
    RedisMessageBroker,
    ResourcePool,
    EventProcessor,
    EventDrivenHookSystem,
    get_event_system,
)
from moai_adk.core.jit_enhanced_hook_manager import HookEvent


class TestEventDataclass:
    """Test Event dataclass"""

    def test_event_initialization(self):
        """Test event initializes correctly"""
        event = Event(
            event_id="test_id",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={"test": "data"}
        )
        assert event.event_id == "test_id"
        assert event.event_type == EventType.HOOK_EXECUTION_REQUEST
        assert event.priority == EventPriority.NORMAL

    def test_event_to_dict(self):
        """Test event to_dict conversion"""
        timestamp = datetime.now()
        event = Event(
            event_id="test_id",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=timestamp,
            payload={"test": "data"}
        )
        event_dict = event.to_dict()

        assert event_dict["event_id"] == "test_id"
        assert event_dict["event_type"] == EventType.HOOK_EXECUTION_REQUEST.value
        assert event_dict["priority"] == EventPriority.NORMAL.value

    def test_event_from_dict(self):
        """Test event from_dict creation"""
        timestamp = datetime.now()
        event_data = {
            "event_id": "test_id",
            "event_type": "hook_execution_request",
            "priority": 3,
            "timestamp": timestamp.isoformat(),
            "payload": {"test": "data"},
            "source": "test_source",
            "retry_count": 1,
            "max_retries": 3,
            "timeout_seconds": 60.0,
            "tags": {}
        }

        event = Event.from_dict(event_data)
        assert event.event_id == "test_id"
        assert event.event_type == EventType.HOOK_EXECUTION_REQUEST


class TestHookExecutionEvent:
    """Test HookExecutionEvent dataclass"""

    def test_hook_execution_event_initialization(self):
        """Test hook execution event initializes correctly"""
        event = HookExecutionEvent(
            event_id="test_id",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
            hook_path="test_hook.py",
            hook_event_type=HookEvent.SESSION_START
        )
        assert event.hook_path == "test_hook.py"
        assert event.hook_event_type == HookEvent.SESSION_START

    def test_hook_execution_event_validation(self):
        """Test hook execution event validates hook_path"""
        with pytest.raises(ValueError, match="hook_path is required"):
            HookExecutionEvent(
                event_id="test_id",
                event_type=EventType.HOOK_EXECUTION_REQUEST,
                priority=EventPriority.NORMAL,
                timestamp=datetime.now(),
                payload={}
            )


class TestWorkflowEvent:
    """Test WorkflowEvent dataclass"""

    def test_workflow_event_initialization(self):
        """Test workflow event initializes correctly"""
        event = WorkflowEvent(
            event_id="test_id",
            event_type=EventType.WORKFLOW_ORCHESTRATION,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
            workflow_id="workflow_123"
        )
        assert event.workflow_id == "workflow_123"

    def test_workflow_event_validation(self):
        """Test workflow event validates workflow_id"""
        with pytest.raises(ValueError, match="workflow_id is required"):
            WorkflowEvent(
                event_id="test_id",
                event_type=EventType.WORKFLOW_ORCHESTRATION,
                priority=EventPriority.NORMAL,
                timestamp=datetime.now(),
                payload={}
            )


class TestInMemoryMessageBroker:
    """Test InMemoryMessageBroker class"""

    def test_broker_initialization(self):
        """Test in-memory broker initializes correctly"""
        broker = InMemoryMessageBroker(max_queue_size=1000)
        assert broker.max_queue_size == 1000

    @pytest.mark.asyncio
    async def test_publish_event(self):
        """Test publishing event to broker"""
        broker = InMemoryMessageBroker()
        event = Event(
            event_id="test",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={}
        )

        result = await broker.publish("test_topic", event)
        assert result is True
        assert broker._stats["messages_published"] == 1

    @pytest.mark.asyncio
    async def test_subscribe_to_topic(self):
        """Test subscribing to topic"""
        broker = InMemoryMessageBroker()

        async def callback(event):
            pass

        sub_id = await broker.subscribe("test_topic", callback)
        assert sub_id is not None
        assert len(broker.subscribers["test_topic"]) == 1

    @pytest.mark.asyncio
    async def test_unsubscribe_from_topic(self):
        """Test unsubscribing from topic"""
        broker = InMemoryMessageBroker()

        async def callback(event):
            pass

        sub_id = await broker.subscribe("test_topic", callback)
        result = await broker.unsubscribe(sub_id)
        assert result is True
        assert len(broker.subscribers["test_topic"]) == 0

    @pytest.mark.asyncio
    async def test_create_queue(self):
        """Test creating message queue"""
        broker = InMemoryMessageBroker()

        result = await broker.create_queue("test_queue", {"max_size": 100})
        assert result is True
        assert "test_queue" in broker.queues

    @pytest.mark.asyncio
    async def test_delete_queue(self):
        """Test deleting message queue"""
        broker = InMemoryMessageBroker()

        await broker.create_queue("test_queue", {})
        result = await broker.delete_queue("test_queue")
        assert result is True
        assert "test_queue" not in broker.queues

    def test_get_stats(self):
        """Test getting broker statistics"""
        broker = InMemoryMessageBroker()
        stats = broker.get_stats()

        assert "messages_published" in stats
        assert "messages_delivered" in stats
        assert "queue_count" in stats

    @pytest.mark.asyncio
    async def test_safe_callback_async(self):
        """Test safe callback execution for async function"""
        broker = InMemoryMessageBroker()
        called = False

        async def async_callback(event):
            nonlocal called
            called = True

        event = Event(
            event_id="test",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={}
        )

        await broker._safe_callback(async_callback, event)
        await asyncio.sleep(0.1)
        assert called is True


class TestRedisMessageBroker:
    """Test RedisMessageBroker class"""

    def test_redis_broker_initialization(self):
        """Test Redis broker initializes correctly"""
        broker = RedisMessageBroker(redis_url="redis://localhost:6379/0")
        assert broker.redis_url == "redis://localhost:6379/0"

    @pytest.mark.asyncio
    async def test_redis_broker_publish_without_connection(self):
        """Test Redis broker handles missing connection gracefully"""
        broker = RedisMessageBroker()

        event = Event(
            event_id="test",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={}
        )

        with patch.object(broker, '_connect', side_effect=ConnectionError("No Redis")):
            result = await broker.publish("test_topic", event)
            assert result is False


class TestResourcePool:
    """Test ResourcePool class"""

    def test_resource_pool_initialization(self):
        """Test resource pool initializes correctly"""
        pool = ResourcePool(ResourceIsolationLevel.SHARED, max_concurrent=10)
        assert pool.isolation_level == ResourceIsolationLevel.SHARED
        assert pool.max_concurrent == 10

    @pytest.mark.asyncio
    async def test_get_semaphore_shared_isolation(self):
        """Test getting semaphore with shared isolation"""
        pool = ResourcePool(ResourceIsolationLevel.SHARED)

        sem1 = await pool.get_semaphore("hook1", HookEvent.SESSION_START)
        sem2 = await pool.get_semaphore("hook2", HookEvent.SESSION_START)

        assert sem1 is sem2  # Same semaphore for shared isolation

    @pytest.mark.asyncio
    async def test_get_semaphore_type_isolated(self):
        """Test getting semaphore with type isolation"""
        pool = ResourcePool(ResourceIsolationLevel.TYPE_ISOLATED)

        sem1 = await pool.get_semaphore("hook1", HookEvent.SESSION_START)
        sem2 = await pool.get_semaphore("hook1", HookEvent.SESSION_END)

        assert sem1 is not sem2  # Different semaphores for different event types

    @pytest.mark.asyncio
    async def test_acquire_release_execution_slot(self):
        """Test acquiring and releasing execution slots"""
        pool = ResourcePool(ResourceIsolationLevel.SHARED)

        acquired = await pool.acquire_execution_slot("hook1", HookEvent.SESSION_START)
        assert acquired is True
        assert pool._stats["active_executions"] == 1

        await pool.release_execution_slot("hook1", HookEvent.SESSION_START)
        assert pool._stats["active_executions"] == 0

    def test_get_pool_stats(self):
        """Test getting pool statistics"""
        pool = ResourcePool(ResourceIsolationLevel.SHARED)
        stats = pool.get_stats()

        assert "total_executions" in stats
        assert "active_executions" in stats


class TestEventProcessor:
    """Test EventProcessor class"""

    def test_event_processor_initialization(self):
        """Test event processor initializes correctly"""
        pool = ResourcePool(ResourceIsolationLevel.SHARED)
        processor = EventProcessor(pool)

        assert processor.resource_pool == pool
        assert len(processor._handlers) == 0

    def test_register_handler(self):
        """Test registering event handler"""
        pool = ResourcePool(ResourceIsolationLevel.SHARED)
        processor = EventProcessor(pool)

        def handler(event):
            pass

        processor.register_handler(EventType.HOOK_EXECUTION_REQUEST, handler)
        assert len(processor._handlers[EventType.HOOK_EXECUTION_REQUEST]) == 1

    @pytest.mark.asyncio
    async def test_process_event_no_handlers(self):
        """Test processing event with no handlers"""
        pool = ResourcePool(ResourceIsolationLevel.SHARED)
        processor = EventProcessor(pool)

        event = Event(
            event_id="test",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={}
        )

        result = await processor.process_event(event)
        assert result is True

    @pytest.mark.asyncio
    async def test_process_event_with_handler(self):
        """Test processing event with handler"""
        pool = ResourcePool(ResourceIsolationLevel.SHARED)
        processor = EventProcessor(pool)

        handled = False

        async def handler(event):
            nonlocal handled
            handled = True

        processor.register_handler(EventType.HOOK_EXECUTION_REQUEST, handler)

        event = Event(
            event_id="test",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={}
        )

        result = await processor.process_event(event)
        await asyncio.sleep(0.1)
        assert handled is True

    def test_get_stats(self):
        """Test getting processor statistics"""
        pool = ResourcePool(ResourceIsolationLevel.SHARED)
        processor = EventProcessor(pool)

        stats = processor.get_stats()
        assert "events_processed" in stats
        assert "events_failed" in stats


class TestEventDrivenHookSystem:
    """Test EventDrivenHookSystem class"""

    def test_system_initialization(self):
        """Test event system initializes correctly"""
        system = EventDrivenHookSystem(
            message_broker_type=MessageBrokerType.MEMORY,
            isolation_level=ResourceIsolationLevel.SHARED,
            max_concurrent_hooks=5
        )
        assert system.message_broker_type == MessageBrokerType.MEMORY
        assert system.max_concurrent_hooks == 5
        assert not system._running

    def test_create_memory_message_broker(self):
        """Test creating in-memory message broker"""
        system = EventDrivenHookSystem(message_broker_type=MessageBrokerType.MEMORY)
        assert isinstance(system.message_broker, InMemoryMessageBroker)

    def test_create_redis_message_broker(self):
        """Test creating Redis message broker"""
        system = EventDrivenHookSystem(message_broker_type=MessageBrokerType.REDIS)
        assert isinstance(system.message_broker, RedisMessageBroker)

    @pytest.mark.asyncio
    async def test_start_system(self):
        """Test starting event system"""
        system = EventDrivenHookSystem()

        await system.start()
        assert system._running is True

        await system.stop()

    @pytest.mark.asyncio
    async def test_stop_system(self):
        """Test stopping event system"""
        system = EventDrivenHookSystem()

        await system.start()
        await system.stop()

        assert system._running is False

    def test_register_event_handlers(self):
        """Test registering event handlers"""
        system = EventDrivenHookSystem()
        system._register_event_handlers()

        # Check that handlers are registered
        assert len(system.event_processor._handlers) > 0

    @pytest.mark.asyncio
    async def test_setup_message_queues(self):
        """Test setting up message queues"""
        system = EventDrivenHookSystem()

        await system._setup_message_queues()
        # Queues should be created in message broker

    @pytest.mark.asyncio
    async def test_publish_hook_execution_event(self):
        """Test publishing hook execution event"""
        system = EventDrivenHookSystem()

        event_id = await system.publish_hook_execution_event(
            hook_path="test_hook.py",
            event_type=HookEvent.SESSION_START,
            execution_context={"test": "data"}
        )

        assert event_id is not None
        assert system._system_metrics["events_published"] == 1

    @pytest.mark.asyncio
    async def test_publish_system_alert(self):
        """Test publishing system alert"""
        system = EventDrivenHookSystem()

        alert_id = await system.publish_system_alert(
            alert_type="TEST",
            message="Test alert",
            severity=EventPriority.HIGH
        )

        assert alert_id is not None
        assert system._system_metrics["events_published"] == 1

    def test_get_queue_name_by_priority(self):
        """Test getting queue name by priority"""
        system = EventDrivenHookSystem()

        queue = system._get_queue_name_by_priority(EventPriority.CRITICAL)
        assert queue == "system_events"

        queue = system._get_queue_name_by_priority(EventPriority.HIGH)
        assert queue == "hook_execution_high"

        queue = system._get_queue_name_by_priority(EventPriority.BULK)
        assert queue == "analytics"

    @pytest.mark.asyncio
    async def test_get_system_status(self):
        """Test getting system status"""
        system = EventDrivenHookSystem()

        status = await system.get_system_status()
        assert "status" in status
        assert "message_broker_type" in status
        assert "system_metrics" in status

    def test_get_event_flow_diagram(self):
        """Test getting event flow diagram"""
        system = EventDrivenHookSystem()

        diagram = system.get_event_flow_diagram()
        assert "event_types" in diagram
        assert "flow_pattern" in diagram
        assert "message_broker_type" in diagram

    @pytest.mark.asyncio
    async def test_persist_and_load_events(self, tmp_path):
        """Test persisting and loading events"""
        system = EventDrivenHookSystem(
            enable_persistence=True,
            persistence_path=tmp_path
        )

        # Add pending event
        event = Event(
            event_id="test_id",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={}
        )
        system._pending_events["test_id"] = event

        # Persist
        await system._persist_events()

        # Load
        system2 = EventDrivenHookSystem(
            enable_persistence=True,
            persistence_path=tmp_path
        )
        await system2._load_persisted_events()

        assert "test_id" in system2._pending_events


class TestEnums:
    """Test enum classes"""

    def test_event_type_enum(self):
        """Test EventType enum has all values"""
        assert EventType.HOOK_EXECUTION_REQUEST is not None
        assert EventType.HOOK_EXECUTION_COMPLETED is not None
        assert EventType.HOOK_EXECUTION_FAILED is not None
        assert EventType.SYSTEM_ALERT is not None

    def test_event_priority_enum(self):
        """Test EventPriority enum has all values"""
        assert EventPriority.CRITICAL is not None
        assert EventPriority.HIGH is not None
        assert EventPriority.NORMAL is not None
        assert EventPriority.LOW is not None
        assert EventPriority.BULK is not None

    def test_resource_isolation_level_enum(self):
        """Test ResourceIsolationLevel enum has all values"""
        assert ResourceIsolationLevel.SHARED is not None
        assert ResourceIsolationLevel.TYPE_ISOLATED is not None
        assert ResourceIsolationLevel.PRIORITY_ISOLATED is not None
        assert ResourceIsolationLevel.FULL_ISOLATION is not None

    def test_message_broker_type_enum(self):
        """Test MessageBrokerType enum has all values"""
        assert MessageBrokerType.MEMORY is not None
        assert MessageBrokerType.REDIS is not None
        assert MessageBrokerType.RABBITMQ is not None
        assert MessageBrokerType.KAFKA is not None


class TestGlobalFunctions:
    """Test global convenience functions"""

    def test_get_event_system_singleton(self):
        """Test get_event_system returns singleton"""
        system1 = get_event_system()
        system2 = get_event_system()
        assert system1 is system2

    @pytest.mark.asyncio
    async def test_start_event_system(self):
        """Test start_event_system function"""
        from moai_adk.core.event_driven_hook_system import start_event_system, stop_event_system

        await start_event_system()
        await stop_event_system()

    @pytest.mark.asyncio
    async def test_execute_hook_with_event_system(self):
        """Test execute_hook_with_event_system function"""
        from moai_adk.core.event_driven_hook_system import execute_hook_with_event_system

        event_id = await execute_hook_with_event_system(
            hook_path="test.py",
            event_type=HookEvent.SESSION_START,
            execution_context={}
        )
        assert event_id is not None
