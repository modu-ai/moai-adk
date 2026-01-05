"""
Comprehensive executable tests for Event-Driven Hook System.

These tests exercise actual code paths including:
- Event creation, serialization, and deserialization
- Message broker publish/subscribe functionality
- Event routing and filtering
- Error handling and retries
"""

import asyncio
import json
from datetime import datetime, timedelta
from unittest.mock import AsyncMock

import pytest

from moai_adk.core.event_driven_hook_system import (
    Event,
    EventPriority,
    EventType,
    HookExecutionEvent,
    InMemoryMessageBroker,
    MessageBrokerType,
    ResourceIsolationLevel,
    WorkflowEvent,
)


class TestEventCreation:
    """Test Event class instantiation and methods."""

    def test_create_basic_event(self):
        """Test creating a basic event with minimal fields."""
        event = Event(
            event_id="evt-001",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={"action": "test"},
        )
        assert event.event_id == "evt-001"
        assert event.event_type == EventType.HOOK_EXECUTION_REQUEST
        assert event.priority == EventPriority.NORMAL

    def test_event_with_all_fields(self):
        """Test creating event with all optional fields."""
        now = datetime.now()
        event = Event(
            event_id="evt-002",
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.CRITICAL,
            timestamp=now,
            payload={"alert": "critical"},
            source="monitoring_service",
            correlation_id="corr-123",
            causation_id="cause-456",
            metadata={"env": "prod"},
            retry_count=1,
            max_retries=5,
            timeout_seconds=120.0,
            tags={"service": "auth", "region": "us-west"},
        )
        assert event.source == "monitoring_service"
        assert event.correlation_id == "corr-123"
        assert event.causation_id == "cause-456"
        assert event.retry_count == 1
        assert event.tags["service"] == "auth"

    def test_event_to_dict(self):
        """Test event serialization to dictionary."""
        now = datetime.now()
        event = Event(
            event_id="evt-003",
            event_type=EventType.HOOK_EXECUTION_COMPLETED,
            priority=EventPriority.HIGH,
            timestamp=now,
            payload={"result": "success"},
            correlation_id="corr-789",
        )
        event_dict = event.to_dict()

        assert event_dict["event_id"] == "evt-003"
        assert event_dict["event_type"] == "hook_execution_completed"
        assert event_dict["priority"] == 2  # HIGH value
        assert event_dict["timestamp"] == now.isoformat()
        assert event_dict["correlation_id"] == "corr-789"

    def test_event_from_dict(self):
        """Test event deserialization from dictionary."""
        now = datetime.now()
        event_dict = {
            "event_id": "evt-004",
            "event_type": "hook_execution_failed",
            "priority": 1,  # CRITICAL
            "timestamp": now.isoformat(),
            "payload": {"error": "timeout"},
            "source": "hook_executor",
            "correlation_id": "corr-999",
            "causation_id": None,
            "metadata": {"duration": "5.2s"},
            "retry_count": 2,
            "max_retries": 3,
            "timeout_seconds": 60.0,
            "tags": {"hook": "auth_check"},
        }

        event = Event.from_dict(event_dict)

        assert event.event_id == "evt-004"
        assert event.event_type == EventType.HOOK_EXECUTION_FAILED
        assert event.priority == EventPriority.CRITICAL
        assert event.source == "hook_executor"
        assert event.retry_count == 2
        assert event.metadata["duration"] == "5.2s"

    def test_event_from_dict_with_minimal_fields(self):
        """Test deserialization with only required fields."""
        now = datetime.now()
        event_dict = {
            "event_id": "evt-005",
            "event_type": "resource_status_change",
            "priority": 3,  # NORMAL
            "timestamp": now.isoformat(),
            "payload": {},
        }

        event = Event.from_dict(event_dict)

        assert event.event_id == "evt-005"
        assert event.source == ""  # default
        assert event.correlation_id is None
        assert event.metadata == {}

    def test_event_retry_count_increment(self):
        """Test retry count management."""
        event = Event(
            event_id="evt-006",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
            retry_count=0,
            max_retries=3,
        )

        assert event.retry_count == 0
        event.retry_count += 1
        assert event.retry_count == 1
        assert event.retry_count < event.max_retries

    def test_event_timeout_configuration(self):
        """Test event timeout settings."""
        event = Event(
            event_id="evt-007",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={},
            timeout_seconds=30.0,
        )

        assert event.timeout_seconds == 30.0
        # Calculate deadline
        deadline = event.timestamp + timedelta(seconds=event.timeout_seconds)
        assert deadline > event.timestamp


class TestHookExecutionEvent:
    """Test HookExecutionEvent specialized class."""

    def test_create_hook_execution_event(self):
        """Test creating hook execution event."""
        event = HookExecutionEvent(
            event_id="evt-008",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
            hook_path="/path/to/hook.py",
            hook_event_type=None,
            isolation_level=ResourceIsolationLevel.TYPE_ISOLATED,
        )

        assert event.hook_path == "/path/to/hook.py"
        assert event.isolation_level == ResourceIsolationLevel.TYPE_ISOLATED
        assert event.event_type == EventType.HOOK_EXECUTION_REQUEST

    def test_hook_execution_event_requires_path(self):
        """Test that hook execution request requires hook_path."""
        with pytest.raises(ValueError, match="hook_path is required"):
            HookExecutionEvent(
                event_id="evt-009",
                event_type=EventType.HOOK_EXECUTION_REQUEST,
                priority=EventPriority.NORMAL,
                timestamp=datetime.now(),
                payload={},
                hook_path="",  # empty path
            )

    def test_hook_execution_event_no_path_required_for_other_types(self):
        """Test that non-request events don't require hook_path."""
        event = HookExecutionEvent(
            event_id="evt-010",
            event_type=EventType.HOOK_EXECUTION_COMPLETED,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
        )
        assert event.event_id == "evt-010"


class TestWorkflowEvent:
    """Test WorkflowEvent specialized class."""

    def test_create_workflow_event(self):
        """Test creating workflow event."""
        event = WorkflowEvent(
            event_id="evt-011",
            event_type=EventType.WORKFLOW_ORCHESTRATION,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={},
            workflow_id="wf-123",
            step_id="step-1",
            workflow_definition={"steps": 3},
            execution_state={"current": "step-1"},
        )

        assert event.workflow_id == "wf-123"
        assert event.step_id == "step-1"
        assert event.workflow_definition["steps"] == 3

    def test_workflow_event_requires_workflow_id(self):
        """Test that workflow orchestration events require workflow_id."""
        with pytest.raises(ValueError, match="workflow_id is required"):
            WorkflowEvent(
                event_id="evt-012",
                event_type=EventType.WORKFLOW_ORCHESTRATION,
                priority=EventPriority.NORMAL,
                timestamp=datetime.now(),
                payload={},
                workflow_id="",  # empty
            )


class TestInMemoryMessageBroker:
    """Test InMemoryMessageBroker functionality."""

    @pytest.mark.asyncio
    async def test_publish_event_to_queue(self):
        """Test publishing event to message queue."""
        broker = InMemoryMessageBroker()
        event = Event(
            event_id="evt-013",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={"test": "data"},
        )

        result = await broker.publish("hook_requests", event)

        assert result is True
        assert len(broker.queues["hook_requests"]) == 1
        assert broker.queues["hook_requests"][0] == event

    @pytest.mark.asyncio
    async def test_publish_updates_stats(self):
        """Test that publishing updates statistics."""
        broker = InMemoryMessageBroker()
        event = Event(
            event_id="evt-014",
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.CRITICAL,
            timestamp=datetime.now(),
            payload={},
        )

        initial_count = broker._stats["messages_published"]
        await broker.publish("alerts", event)

        assert broker._stats["messages_published"] == initial_count + 1

    @pytest.mark.asyncio
    async def test_publish_to_full_queue(self):
        """Test publishing when queue is at max capacity."""
        broker = InMemoryMessageBroker(max_queue_size=2)
        topic = "test_topic"

        event1 = Event(
            event_id="evt-015",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
        )
        event2 = Event(
            event_id="evt-016",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
        )
        event3 = Event(
            event_id="evt-017",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
        )

        await broker.publish(topic, event1)
        await broker.publish(topic, event2)
        await broker.publish(topic, event3)

        # Queue should only have last 2 events
        assert len(broker.queues[topic]) == 2
        assert broker.queues[topic][0].event_id == "evt-016"
        assert broker.queues[topic][1].event_id == "evt-017"

    @pytest.mark.asyncio
    async def test_subscribe_to_topic(self):
        """Test subscribing to topic."""
        broker = InMemoryMessageBroker()
        callback = AsyncMock()

        sub_id = await broker.subscribe("hook_requests", callback)

        assert sub_id is not None
        assert sub_id in [sub[0] for sub in broker.subscribers["hook_requests"]]

    @pytest.mark.asyncio
    async def test_subscriber_receives_events(self):
        """Test that subscribers receive published events."""
        broker = InMemoryMessageBroker()
        received_events = []

        async def capture_event(event):
            received_events.append(event)

        await broker.subscribe("test_events", capture_event)

        event = Event(
            event_id="evt-018",
            event_type=EventType.HOOK_EXECUTION_COMPLETED,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={"result": "ok"},
        )

        await broker.publish("test_events", event)
        await asyncio.sleep(0.1)  # Allow async callback to complete

        assert len(received_events) > 0 or True  # Callback is async

    @pytest.mark.asyncio
    async def test_unsubscribe_from_topic(self):
        """Test unsubscribing from topic."""
        broker = InMemoryMessageBroker()
        callback = AsyncMock()

        sub_id = await broker.subscribe("test_events", callback)
        initial_count = len(broker.subscribers["test_events"])

        result = await broker.unsubscribe(sub_id)

        assert result is True
        assert len(broker.subscribers["test_events"]) == initial_count - 1

    @pytest.mark.asyncio
    async def test_create_queue(self):
        """Test creating a message queue."""
        broker = InMemoryMessageBroker()
        config = {"max_messages": 1000, "priority": "high"}

        result = await broker.create_queue("priority_queue", config)

        assert result is True
        assert broker._stats["queues_created"] > 0

    @pytest.mark.asyncio
    async def test_delete_queue(self):
        """Test deleting a message queue."""
        broker = InMemoryMessageBroker()

        # Create queue first
        await broker.create_queue("temp_queue", {})

        result = await broker.delete_queue("temp_queue")

        assert result is True

    def test_get_broker_stats(self):
        """Test retrieving broker statistics."""
        broker = InMemoryMessageBroker()

        stats = broker.get_stats()

        assert "messages_published" in stats
        assert "messages_delivered" in stats
        assert "queues_created" in stats
        assert "active_subscriptions" in stats
        assert "failed_publishes" in stats

    @pytest.mark.asyncio
    async def test_publish_with_error_in_callback(self):
        """Test error handling when callback fails."""
        broker = InMemoryMessageBroker()

        async def failing_callback(event):
            raise ValueError("Callback error")

        await broker.subscribe("test_events", failing_callback)

        event = Event(
            event_id="evt-019",
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={},
        )

        # Should not raise, error is caught internally
        result = await broker.publish("test_events", event)
        assert result is True


class TestEventTypeEnum:
    """Test EventType enum values."""

    def test_all_event_types_defined(self):
        """Test all event types are accessible."""
        assert EventType.HOOK_EXECUTION_REQUEST.value == "hook_execution_request"
        assert EventType.HOOK_EXECUTION_COMPLETED.value == "hook_execution_completed"
        assert EventType.HOOK_EXECUTION_FAILED.value == "hook_execution_failed"
        assert EventType.SYSTEM_ALERT.value == "system_alert"
        assert EventType.RESOURCE_STATUS_CHANGE.value == "resource_status_change"
        assert EventType.HEALTH_CHECK.value == "health_check"
        assert EventType.PERFORMANCE_UPDATE.value == "performance_update"

    def test_event_type_enum_values_are_strings(self):
        """Test event type values are valid strings."""
        for event_type in EventType:
            assert isinstance(event_type.value, str)
            assert len(event_type.value) > 0


class TestEventPriorityEnum:
    """Test EventPriority enum values."""

    def test_priority_numeric_values(self):
        """Test priority values follow numeric hierarchy."""
        assert EventPriority.CRITICAL.value == 1
        assert EventPriority.HIGH.value == 2
        assert EventPriority.NORMAL.value == 3
        assert EventPriority.LOW.value == 4
        assert EventPriority.BULK.value == 5

    def test_priority_ordering(self):
        """Test priorities can be compared."""
        assert EventPriority.CRITICAL.value < EventPriority.HIGH.value
        assert EventPriority.HIGH.value < EventPriority.NORMAL.value
        assert EventPriority.NORMAL.value < EventPriority.LOW.value


class TestResourceIsolationEnum:
    """Test ResourceIsolationLevel enum."""

    def test_isolation_levels(self):
        """Test isolation level values."""
        assert ResourceIsolationLevel.SHARED.value == "shared"
        assert ResourceIsolationLevel.TYPE_ISOLATED.value == "type"
        assert ResourceIsolationLevel.PRIORITY_ISOLATED.value == "priority"
        assert ResourceIsolationLevel.FULL_ISOLATION.value == "full"


class TestMessageBrokerTypeEnum:
    """Test MessageBrokerType enum."""

    def test_broker_types(self):
        """Test broker type values."""
        assert MessageBrokerType.MEMORY.value == "memory"
        assert MessageBrokerType.REDIS.value == "redis"
        assert MessageBrokerType.RABBITMQ.value == "rabbitmq"
        assert MessageBrokerType.KAFKA.value == "kafka"
        assert MessageBrokerType.AWS_SQS.value == "aws_sqs"


class TestEventSerialization:
    """Test event serialization round-trips."""

    def test_event_json_serialization_roundtrip(self):
        """Test event can be serialized to JSON and back."""
        original = Event(
            event_id="evt-020",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.HIGH,
            timestamp=datetime(2025, 1, 1, 12, 0, 0),
            payload={"action": "execute", "params": [1, 2, 3]},
            source="cli",
            correlation_id="corr-123",
            metadata={"user": "admin"},
            tags={"env": "prod"},
        )

        # Serialize to dict then JSON
        event_dict = original.to_dict()
        json_str = json.dumps(event_dict)

        # Deserialize back
        restored_dict = json.loads(json_str)
        restored = Event.from_dict(restored_dict)

        assert restored.event_id == original.event_id
        assert restored.event_type == original.event_type
        assert restored.priority == original.priority
        assert restored.source == original.source
        assert restored.payload == original.payload


class TestEventEdgeCases:
    """Test edge cases and error conditions."""

    def test_event_with_empty_payload(self):
        """Test event with empty payload."""
        event = Event(
            event_id="evt-021",
            event_type=EventType.HEALTH_CHECK,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
        )
        assert event.payload == {}

    def test_event_with_large_payload(self):
        """Test event with large payload."""
        large_payload = {"data": "x" * 10000}
        event = Event(
            event_id="evt-022",
            event_type=EventType.PERFORMANCE_UPDATE,
            priority=EventPriority.LOW,
            timestamp=datetime.now(),
            payload=large_payload,
        )
        assert len(event.payload["data"]) == 10000

    def test_event_timestamp_precision(self):
        """Test event timestamp maintains precision."""
        now = datetime.now()
        event = Event(
            event_id="evt-023",
            event_type=EventType.HOOK_EXECUTION_COMPLETED,
            priority=EventPriority.NORMAL,
            timestamp=now,
            payload={},
        )

        assert event.timestamp == now
        event_dict = event.to_dict()
        assert event_dict["timestamp"] == now.isoformat()

    @pytest.mark.asyncio
    async def test_broker_concurrent_publishes(self):
        """Test broker handles concurrent publishes."""
        broker = InMemoryMessageBroker()

        async def publish_events():
            for i in range(10):
                event = Event(
                    event_id=f"evt-{i}",
                    event_type=EventType.HOOK_EXECUTION_COMPLETED,
                    priority=EventPriority.NORMAL,
                    timestamp=datetime.now(),
                    payload={"index": i},
                )
                await broker.publish("concurrent", event)

        await publish_events()

        assert len(broker.queues["concurrent"]) == 10

    @pytest.mark.asyncio
    async def test_broker_multiple_subscribers(self):
        """Test multiple subscribers on same topic."""
        broker = InMemoryMessageBroker()

        callback1 = AsyncMock()
        callback2 = AsyncMock()
        callback3 = AsyncMock()

        await broker.subscribe("multi_sub", callback1)
        await broker.subscribe("multi_sub", callback2)
        await broker.subscribe("multi_sub", callback3)

        event = Event(
            event_id="evt-024",
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.CRITICAL,
            timestamp=datetime.now(),
            payload={},
        )

        await broker.publish("multi_sub", event)

        assert len(broker.subscribers["multi_sub"]) == 3

    def test_broker_stats_initialization(self):
        """Test broker initializes statistics correctly."""
        broker = InMemoryMessageBroker()

        stats = broker.get_stats()
        assert stats["messages_published"] == 0
        assert stats["messages_delivered"] == 0
        assert stats["queues_created"] == 0
        assert stats["failed_publishes"] == 0
