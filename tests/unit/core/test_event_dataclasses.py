"""Tests for Event-driven system dataclasses."""

from datetime import datetime

from moai_adk.core.event_driven_hook_system import (
    Event,
    EventPriority,
    EventType,
    MessageBrokerType,
    ResourceIsolationLevel,
)


class TestEventEnums:
    """Test event system enums."""

    def test_event_type_values(self):
        """Test event types are defined."""
        assert hasattr(EventType, "HOOK_EXECUTION_REQUEST")
        assert hasattr(EventType, "HOOK_EXECUTION_COMPLETED")
        assert hasattr(EventType, "HOOK_EXECUTION_FAILED")

    def test_event_priority_values(self):
        """Test event priorities are defined."""
        assert EventPriority.CRITICAL.value == 1
        assert EventPriority.HIGH.value == 2
        assert EventPriority.NORMAL.value == 3
        assert EventPriority.LOW.value == 4
        assert EventPriority.BULK.value == 5

    def test_isolation_level_values(self):
        """Test resource isolation levels."""
        assert hasattr(ResourceIsolationLevel, "SHARED")
        assert hasattr(ResourceIsolationLevel, "TYPE_ISOLATED")

    def test_message_broker_types(self):
        """Test message broker types."""
        assert hasattr(MessageBrokerType, "MEMORY")
        assert hasattr(MessageBrokerType, "REDIS")


class TestEventDataclass:
    """Test Event dataclass."""

    def test_event_creation(self):
        """Test creating an Event instance."""
        event = Event(
            event_id="evt_001",
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=EventPriority.NORMAL,
            timestamp=datetime.now(),
            payload={"hook": "test"},
        )
        assert event.event_id == "evt_001"
        assert event.priority == EventPriority.NORMAL

    def test_event_default_values(self):
        """Test Event default values."""
        event = Event(
            event_id="evt_002",
            event_type=EventType.HEALTH_CHECK,
            priority=EventPriority.HIGH,
            timestamp=datetime.now(),
            payload={},
        )
        assert event.source == ""
        assert event.correlation_id is None
        assert event.retry_count == 0
        assert event.max_retries == 3

    def test_event_with_correlation_id(self):
        """Test Event with correlation ID for tracing."""
        event = Event(
            event_id="evt_003",
            event_type=EventType.WORKFLOW_ORCHESTRATION,
            priority=EventPriority.CRITICAL,
            timestamp=datetime.now(),
            payload={"workflow": "spec"},
            correlation_id="corr_abc123",
            causation_id="evt_002",
        )
        assert event.correlation_id == "corr_abc123"
        assert event.causation_id == "evt_002"

    def test_event_to_dict(self):
        """Test converting Event to dictionary."""
        timestamp = datetime.now()
        event = Event(
            event_id="evt_004",
            event_type=EventType.HOOK_EXECUTION_COMPLETED,
            priority=EventPriority.NORMAL,
            timestamp=timestamp,
            payload={"result": "success"},
            source="hook_manager",
        )
        event_dict = event.to_dict()
        assert event_dict["event_id"] == "evt_004"
        assert event_dict["source"] == "hook_manager"

    def test_event_with_metadata(self):
        """Test Event with custom metadata."""
        metadata = {
            "execution_time_ms": 45.5,
            "tokens_used": 150,
        }
        event = Event(
            event_id="evt_005",
            event_type=EventType.PERFORMANCE_UPDATE,
            priority=EventPriority.LOW,
            timestamp=datetime.now(),
            payload={},
            metadata=metadata,
        )
        assert event.metadata == metadata

    def test_event_batch_processing(self):
        """Test processing multiple events."""
        events = []
        for i in range(10):
            event = Event(
                event_id=f"evt_{i:03d}",
                event_type=EventType.HOOK_EXECUTION_REQUEST,
                priority=EventPriority.NORMAL,
                timestamp=datetime.now(),
                payload={"index": i},
            )
            events.append(event)

        assert len(events) == 10
        assert all(e.priority == EventPriority.NORMAL for e in events)

    def test_event_priority_comparison(self):
        """Test comparing event priorities."""
        critical = EventPriority.CRITICAL
        normal = EventPriority.NORMAL
        assert critical.value < normal.value
