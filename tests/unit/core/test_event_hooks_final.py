"""
Final comprehensive tests for EventDrivenHookSystem - targeting 60%+ coverage.

Focus areas:
- Event dataclass creation and serialization (to_dict/from_dict)
- EventType, EventPriority, ResourceIsolationLevel, MessageBrokerType enums
- InMemoryMessageBroker basic operations
- Event and specialized event classes (HookExecutionEvent, WorkflowEvent)
- ResourcePool and EventProcessor basic initialization

All tests use @patch to mock async operations and external dependencies.
"""

import uuid
from datetime import datetime

import pytest

from moai_adk.core.event_driven_hook_system import (
    Event,
    EventPriority,
    EventProcessor,
    EventType,
    HookExecutionEvent,
    InMemoryMessageBroker,
    MessageBrokerType,
    ResourceIsolationLevel,
    ResourcePool,
    WorkflowEvent,
)
from moai_adk.core.jit_enhanced_hook_manager import HookEvent

# ============================================================================
# Test EventType Enum
# ============================================================================


class TestEventType:
    """Test EventType enum values and behavior."""

    def test_all_event_types_defined(self):
        """Test all event types are properly defined."""
        # Act & Assert
        assert EventType.HOOK_EXECUTION_REQUEST.value == "hook_execution_request"
        assert EventType.HOOK_EXECUTION_COMPLETED.value == "hook_execution_completed"
        assert EventType.HOOK_EXECUTION_FAILED.value == "hook_execution_failed"
        assert EventType.SYSTEM_ALERT.value == "system_alert"
        assert EventType.RESOURCE_STATUS_CHANGE.value == "resource_status_change"
        assert EventType.HEALTH_CHECK.value == "health_check"
        assert EventType.PERFORMANCE_UPDATE.value == "performance_update"
        assert EventType.BATCH_EXECUTION_REQUEST.value == "batch_execution_request"
        assert EventType.WORKFLOW_ORCHESTRATION.value == "workflow_orchestration"

    def test_event_type_iteration(self):
        """Test iterating through event types."""
        # Act
        types = list(EventType)

        # Assert
        assert len(types) >= 9


# ============================================================================
# Test EventPriority Enum
# ============================================================================


class TestEventPriority:
    """Test EventPriority enum values and ordering."""

    def test_priority_levels(self):
        """Test all priority levels are defined."""
        # Act & Assert
        assert EventPriority.CRITICAL.value == 1
        assert EventPriority.HIGH.value == 2
        assert EventPriority.NORMAL.value == 3
        assert EventPriority.LOW.value == 4
        assert EventPriority.BULK.value == 5

    def test_priority_ordering(self):
        """Test priorities can be compared."""
        # Act & Assert
        assert EventPriority.CRITICAL.value < EventPriority.HIGH.value
        assert EventPriority.HIGH.value < EventPriority.NORMAL.value
        assert EventPriority.NORMAL.value < EventPriority.LOW.value


# ============================================================================
# Test ResourceIsolationLevel Enum
# ============================================================================


class TestResourceIsolationLevel:
    """Test ResourceIsolationLevel enum values."""

    def test_isolation_levels(self):
        """Test all isolation levels are defined."""
        # Act & Assert
        assert ResourceIsolationLevel.SHARED.value == "shared"
        assert ResourceIsolationLevel.TYPE_ISOLATED.value == "type"
        assert ResourceIsolationLevel.PRIORITY_ISOLATED.value == "priority"
        assert ResourceIsolationLevel.FULL_ISOLATION.value == "full"


# ============================================================================
# Test MessageBrokerType Enum
# ============================================================================


class TestMessageBrokerType:
    """Test MessageBrokerType enum values."""

    def test_broker_types(self):
        """Test all message broker types are defined."""
        # Act & Assert
        assert MessageBrokerType.MEMORY.value == "memory"
        assert MessageBrokerType.REDIS.value == "redis"
        assert MessageBrokerType.RABBITMQ.value == "rabbitmq"
        assert MessageBrokerType.KAFKA.value == "kafka"
        assert MessageBrokerType.AWS_SQS.value == "aws_sqs"


# ============================================================================
# Test Event Dataclass
# ============================================================================


class TestEventDataclass:
    """Test Event dataclass creation and serialization."""

    def test_event_creation(self):
        """Test creating an Event instance."""
        # Arrange
        event_id = str(uuid.uuid4())
        timestamp = datetime.now()

        # Act
        event = Event(
            event_id=event_id,
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.HIGH,
            timestamp=timestamp,
            payload={"message": "Test alert"},
        )

        # Assert
        assert event.event_id == event_id
        assert event.event_type == EventType.SYSTEM_ALERT
        assert event.priority == EventPriority.HIGH
        assert event.payload == {"message": "Test alert"}
        assert event.source == ""
        assert event.correlation_id is None
        assert event.retry_count == 0

    def test_event_with_all_fields(self):
        """Test creating Event with all optional fields."""
        # Arrange
        event_id = str(uuid.uuid4())
        correlation_id = str(uuid.uuid4())
        causation_id = str(uuid.uuid4())
        timestamp = datetime.now()

        # Act
        event = Event(
            event_id=event_id,
            event_type=EventType.PERFORMANCE_UPDATE,
            priority=EventPriority.LOW,
            timestamp=timestamp,
            payload={"cpu": 75, "memory": 82},
            source="monitoring-service",
            correlation_id=correlation_id,
            causation_id=causation_id,
            metadata={"region": "us-east-1"},
            retry_count=1,
            max_retries=5,
            timeout_seconds=120.0,
            tags={"env": "production"},
        )

        # Assert
        assert event.source == "monitoring-service"
        assert event.correlation_id == correlation_id
        assert event.causation_id == causation_id
        assert event.metadata == {"region": "us-east-1"}
        assert event.retry_count == 1
        assert event.max_retries == 5
        assert event.timeout_seconds == 120.0
        assert event.tags == {"env": "production"}

    def test_event_to_dict(self):
        """Test converting Event to dictionary."""
        # Arrange
        event_id = str(uuid.uuid4())
        timestamp = datetime.now()
        event = Event(
            event_id=event_id,
            event_type=EventType.HEALTH_CHECK,
            priority=EventPriority.NORMAL,
            timestamp=timestamp,
            payload={"status": "healthy"},
            source="health-monitor",
        )

        # Act
        result = event.to_dict()

        # Assert
        assert isinstance(result, dict)
        assert result["event_id"] == event_id
        assert result["event_type"] == "health_check"
        assert result["priority"] == 3
        assert result["source"] == "health-monitor"
        assert result["payload"] == {"status": "healthy"}

    def test_event_from_dict(self):
        """Test creating Event from dictionary."""
        # Arrange
        event_id = str(uuid.uuid4())
        timestamp = datetime.now().isoformat()
        data = {
            "event_id": event_id,
            "event_type": "system_alert",
            "priority": 2,
            "timestamp": timestamp,
            "payload": {"alert": "test"},
            "source": "test-source",
            "correlation_id": str(uuid.uuid4()),
            "metadata": {"key": "value"},
        }

        # Act
        event = Event.from_dict(data)

        # Assert
        assert event.event_id == event_id
        assert event.event_type == EventType.SYSTEM_ALERT
        assert event.priority == EventPriority.HIGH
        assert event.payload == {"alert": "test"}
        assert event.source == "test-source"

    def test_event_serialization_roundtrip(self):
        """Test Event serialization and deserialization roundtrip."""
        # Arrange
        original_event = Event(
            event_id=str(uuid.uuid4()),
            event_type=EventType.BATCH_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={"items": 100},
            source="batch-processor",
        )

        # Act
        dict_form = original_event.to_dict()
        reconstructed = Event.from_dict(dict_form)

        # Assert
        assert reconstructed.event_id == original_event.event_id
        assert reconstructed.event_type == original_event.event_type
        assert reconstructed.priority == original_event.priority
        assert reconstructed.payload == original_event.payload


# ============================================================================
# Test HookExecutionEvent
# ============================================================================


class TestHookExecutionEvent:
    """Test HookExecutionEvent specialized event class."""

    def test_hook_execution_event_creation(self):
        """Test creating a HookExecutionEvent."""
        # Arrange
        event_id = str(uuid.uuid4())
        hook_path = "/hooks/test.py"

        # Act
        event = HookExecutionEvent(
            event_id=event_id,
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={"hook": "test"},
            hook_path=hook_path,
            hook_event_type=HookEvent.PRE_TOOL_USE,
        )

        # Assert
        assert event.hook_path == hook_path
        assert event.hook_event_type == HookEvent.PRE_TOOL_USE
        assert event.isolation_level == ResourceIsolationLevel.SHARED

    def test_hook_execution_event_with_context(self):
        """Test HookExecutionEvent with execution context."""
        # Arrange
        event = HookExecutionEvent(
            event_id=str(uuid.uuid4()),
            event_type=EventType.HOOK_EXECUTION_COMPLETED,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={"result": "success"},
            hook_path="/hooks/post.py",
            execution_context={"duration": 1.5, "retries": 0},
            isolation_level=ResourceIsolationLevel.TYPE_ISOLATED,
        )

        # Act & Assert
        assert event.execution_context == {"duration": 1.5, "retries": 0}
        assert event.isolation_level == ResourceIsolationLevel.TYPE_ISOLATED

    def test_hook_execution_event_validation(self):
        """Test HookExecutionEvent validation of required fields."""
        # Arrange & Act & Assert
        with pytest.raises(ValueError, match="hook_path is required"):
            HookExecutionEvent(
                event_id=str(uuid.uuid4()),
                event_type=EventType.HOOK_EXECUTION_REQUEST,
                priority=EventPriority.HIGH,
                timestamp=datetime.now(),
                payload={},
                hook_path="",  # Empty hook path should trigger validation
            )


# ============================================================================
# Test WorkflowEvent
# ============================================================================


class TestWorkflowEvent:
    """Test WorkflowEvent specialized event class."""

    def test_workflow_event_creation(self):
        """Test creating a WorkflowEvent."""
        # Arrange
        event_id = str(uuid.uuid4())
        workflow_id = str(uuid.uuid4())

        # Act
        event = WorkflowEvent(
            event_id=event_id,
            event_type=EventType.WORKFLOW_ORCHESTRATION,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={"action": "start"},
            workflow_id=workflow_id,
            step_id="step-1",
        )

        # Assert
        assert event.workflow_id == workflow_id
        assert event.step_id == "step-1"
        assert event.workflow_definition == {}
        assert event.execution_state == {}

    def test_workflow_event_with_definitions(self):
        """Test WorkflowEvent with workflow and execution definitions."""
        # Arrange
        workflow_id = str(uuid.uuid4())
        event = WorkflowEvent(
            event_id=str(uuid.uuid4()),
            event_type=EventType.WORKFLOW_ORCHESTRATION,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={"status": "executing"},
            workflow_id=workflow_id,
            step_id="step-2",
            workflow_definition={"steps": ["start", "process", "end"]},
            execution_state={"current_step": 2, "progress": 66},
        )

        # Act & Assert
        assert event.workflow_definition["steps"] == ["start", "process", "end"]
        assert event.execution_state["progress"] == 66

    def test_workflow_event_validation(self):
        """Test WorkflowEvent validation of required fields."""
        # Arrange & Act & Assert
        with pytest.raises(ValueError, match="workflow_id is required"):
            WorkflowEvent(
                event_id=str(uuid.uuid4()),
                event_type=EventType.WORKFLOW_ORCHESTRATION,
                priority=EventPriority.NORMAL,
                timestamp=datetime.now(),
                payload={},
                workflow_id="",  # Empty workflow_id should trigger validation
            )


# ============================================================================
# Test InMemoryMessageBroker
# ============================================================================


class TestInMemoryMessageBroker:
    """Test InMemoryMessageBroker basic operations."""

    @pytest.mark.asyncio
    async def test_broker_initialization(self):
        """Test initializing InMemoryMessageBroker."""
        # Arrange & Act
        broker = InMemoryMessageBroker(max_queue_size=5000)

        # Assert
        assert broker.max_queue_size == 5000
        assert len(broker.queues) == 0
        assert len(broker.subscribers) == 0

    @pytest.mark.asyncio
    async def test_publish_event(self):
        """Test publishing an event to a topic."""
        # Arrange
        broker = InMemoryMessageBroker()
        event = Event(
            event_id=str(uuid.uuid4()),
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={"test": "data"},
        )

        # Act
        result = await broker.publish("alerts", event)

        # Assert
        assert result is True
        assert len(broker.queues["alerts"]) == 1

    @pytest.mark.asyncio
    async def test_create_queue(self):
        """Test creating a message queue."""
        # Arrange
        broker = InMemoryMessageBroker()

        # Act
        result = await broker.create_queue("test_queue", {"max_size": 1000})

        # Assert
        assert result is True
        assert "test_queue" in broker.queues

    @pytest.mark.asyncio
    async def test_delete_queue(self):
        """Test deleting a message queue."""
        # Arrange
        broker = InMemoryMessageBroker()
        await broker.create_queue("temp_queue", {})

        # Act
        result = await broker.delete_queue("temp_queue")

        # Assert
        assert result is True
        assert "temp_queue" not in broker.queues

    @pytest.mark.asyncio
    async def test_get_stats(self):
        """Test getting broker statistics."""
        # Arrange
        broker = InMemoryMessageBroker()
        event = Event(
            event_id=str(uuid.uuid4()),
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={},
        )

        # Act
        await broker.publish("test", event)
        stats = broker.get_stats()

        # Assert
        assert "messages_published" in stats
        assert stats["messages_published"] == 1
        assert "queue_count" in stats
        assert "total_queued_messages" in stats

    @pytest.mark.asyncio
    async def test_multiple_publishes(self):
        """Test publishing multiple events."""
        # Arrange
        broker = InMemoryMessageBroker()
        events = [
            Event(
                event_id=str(uuid.uuid4()),
                event_type=EventType.PERFORMANCE_UPDATE,
                priority=EventPriority.NORMAL,
                timestamp=datetime.now(),
                payload={"metric": i},
            )
            for i in range(5)
        ]

        # Act
        for event in events:
            await broker.publish("metrics", event)

        # Assert
        assert len(broker.queues["metrics"]) == 5


# ============================================================================
# Test ResourcePool
# ============================================================================


class TestResourcePool:
    """Test ResourcePool initialization and basic operations."""

    def test_resource_pool_initialization_shared(self):
        """Test initializing ResourcePool with SHARED isolation."""
        # Arrange & Act
        pool = ResourcePool(
            isolation_level=ResourceIsolationLevel.SHARED,
            max_concurrent=10,
        )

        # Assert
        assert pool.isolation_level == ResourceIsolationLevel.SHARED
        assert pool.max_concurrent == 10

    def test_resource_pool_initialization_type_isolated(self):
        """Test initializing ResourcePool with TYPE_ISOLATED isolation."""
        # Arrange & Act
        pool = ResourcePool(
            isolation_level=ResourceIsolationLevel.TYPE_ISOLATED,
            max_concurrent=20,
        )

        # Assert
        assert pool.isolation_level == ResourceIsolationLevel.TYPE_ISOLATED
        assert pool.max_concurrent == 20

    def test_resource_pool_initialization_full_isolation(self):
        """Test initializing ResourcePool with FULL_ISOLATION."""
        # Arrange & Act
        pool = ResourcePool(
            isolation_level=ResourceIsolationLevel.FULL_ISOLATION,
            max_concurrent=5,
        )

        # Assert
        assert pool.isolation_level == ResourceIsolationLevel.FULL_ISOLATION
        assert pool.max_concurrent == 5

    def test_resource_pool_stats(self):
        """Test getting ResourcePool statistics."""
        # Arrange
        pool = ResourcePool(
            isolation_level=ResourceIsolationLevel.SHARED,
            max_concurrent=10,
        )

        # Act
        stats = pool.get_stats()

        # Assert
        assert isinstance(stats, dict)


# ============================================================================
# Test EventProcessor
# ============================================================================


class TestEventProcessor:
    """Test EventProcessor initialization and basic operations."""

    def test_event_processor_initialization(self):
        """Test initializing EventProcessor."""
        # Arrange
        pool = ResourcePool(ResourceIsolationLevel.SHARED, max_concurrent=10)

        # Act
        processor = EventProcessor(pool)

        # Assert
        assert processor.resource_pool == pool

    def test_register_event_handler(self):
        """Test registering an event handler."""
        # Arrange
        pool = ResourcePool(ResourceIsolationLevel.SHARED, max_concurrent=10)
        processor = EventProcessor(pool)

        def test_handler(event):
            return True

        # Act
        processor.register_handler(EventType.SYSTEM_ALERT, test_handler)

        # Assert
        assert EventType.SYSTEM_ALERT in processor._handlers
        assert len(processor._handlers[EventType.SYSTEM_ALERT]) == 1

    def test_register_multiple_handlers(self):
        """Test registering multiple event handlers."""
        # Arrange
        pool = ResourcePool(ResourceIsolationLevel.SHARED, max_concurrent=10)
        processor = EventProcessor(pool)

        handlers = {}
        for event_type in [
            EventType.SYSTEM_ALERT,
            EventType.HEALTH_CHECK,
            EventType.PERFORMANCE_UPDATE,
        ]:
            handlers[event_type] = lambda e: True

        # Act
        for event_type, handler in handlers.items():
            processor.register_handler(event_type, handler)

        # Assert
        assert len(processor._handlers) == 3

    def test_event_processor_stats(self):
        """Test getting EventProcessor statistics."""
        # Arrange
        pool = ResourcePool(ResourceIsolationLevel.SHARED, max_concurrent=10)
        processor = EventProcessor(pool)

        # Act
        stats = processor.get_stats()

        # Assert
        assert isinstance(stats, dict)


# ============================================================================
# Test Event Default Values
# ============================================================================


class TestEventDefaults:
    """Test Event and specialized event default values."""

    def test_event_default_source(self):
        """Test Event default source value."""
        # Arrange & Act
        event = Event(
            event_id=str(uuid.uuid4()),
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={},
        )

        # Assert
        assert event.source == ""

    def test_event_default_retry_count(self):
        """Test Event default retry_count."""
        # Arrange & Act
        event = Event(
            event_id=str(uuid.uuid4()),
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={},
        )

        # Assert
        assert event.retry_count == 0
        assert event.max_retries == 3
        assert event.timeout_seconds == 60.0

    def test_event_default_metadata(self):
        """Test Event default metadata."""
        # Arrange & Act
        event = Event(
            event_id=str(uuid.uuid4()),
            event_type=EventType.SYSTEM_ALERT,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={},
        )

        # Assert
        assert event.metadata == {}
        assert event.tags == {}
        assert event.correlation_id is None

    def test_hook_execution_event_default_isolation(self):
        """Test HookExecutionEvent default isolation level."""
        # Arrange & Act
        event = HookExecutionEvent(
            event_id=str(uuid.uuid4()),
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
            hook_path="/test.py",
        )

        # Assert
        assert event.isolation_level == ResourceIsolationLevel.SHARED
        assert event.execution_context == {}

    def test_workflow_event_default_definitions(self):
        """Test WorkflowEvent default definitions."""
        # Arrange & Act
        event = WorkflowEvent(
            event_id=str(uuid.uuid4()),
            event_type=EventType.WORKFLOW_ORCHESTRATION,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={},
            workflow_id=str(uuid.uuid4()),
        )

        # Assert
        assert event.workflow_definition == {}
        assert event.execution_state == {}
        assert event.step_id == ""


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
