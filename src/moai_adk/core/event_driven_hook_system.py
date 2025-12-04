"""
Event-Driven Hook System Architecture

Phase 3: Enterprise-grade event-driven architecture for hook management with
message queuing, resource isolation, and asynchronous event processing.

Key Features:
- Event-driven architecture with message queuing
- Resource isolation between hook types
- Asynchronous event processing capabilities
- Scalable message broker integration
- Event persistence and recovery
- Hook workflow orchestration
- Enterprise-grade reliability and performance
"""

import asyncio
import json
import logging
import time
import uuid
from abc import ABC, abstractmethod
from collections import defaultdict, deque
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional, Set

# Import existing systems
from .jit_enhanced_hook_manager import (
    HookEvent,
)

logger = logging.getLogger(__name__)


class EventType(Enum):
    """Event types in the event-driven system"""

    HOOK_EXECUTION_REQUEST = "hook_execution_request"
    HOOK_EXECUTION_COMPLETED = "hook_execution_completed"
    HOOK_EXECUTION_FAILED = "hook_execution_failed"
    SYSTEM_ALERT = "system_alert"
    RESOURCE_STATUS_CHANGE = "resource_status_change"
    HEALTH_CHECK = "health_check"
    PERFORMANCE_UPDATE = "performance_update"
    BATCH_EXECUTION_REQUEST = "batch_execution_request"
    WORKFLOW_ORCHESTRATION = "workflow_orchestration"


class EventPriority(Enum):
    """Event priority levels for message queuing"""

    CRITICAL = 1  # System-critical events (security, failures)
    HIGH = 2  # High-impact events (performance, alerts)
    NORMAL = 3  # Standard events (hook execution)
    LOW = 4  # Background events (analytics, metrics)
    BULK = 5  # Bulk processing (batch operations)


class ResourceIsolationLevel(Enum):
    """Resource isolation levels for hook execution"""

    SHARED = "shared"  # Share resources across all hooks
    TYPE_ISOLATED = "type"  # Isolate by hook type (event type)
    PRIORITY_ISOLATED = "priority"  # Isolate by priority level
    FULL_ISOLATION = "full"  # Complete isolation for each hook


class MessageBrokerType(Enum):
    """Supported message broker types"""

    MEMORY = "memory"  # In-memory message broker
    REDIS = "redis"  # Redis message broker
    RABBITMQ = "rabbitmq"  # RabbitMQ message broker
    KAFKA = "kafka"  # Apache Kafka
    AWS_SQS = "aws_sqs"  # AWS SQS


@dataclass
class Event:
    """Base event class for the event-driven system"""

    event_id: str
    event_type: EventType
    priority: EventPriority
    timestamp: datetime
    payload: Dict[str, Any]
    source: str = ""
    correlation_id: Optional[str] = None
    causation_id: Optional[str] = None  # Event that caused this event
    metadata: Dict[str, Any] = field(default_factory=dict)
    retry_count: int = 0
    max_retries: int = 3
    timeout_seconds: float = 60.0
    tags: Dict[str, str] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        """Convert event to dictionary for serialization"""
        return {
            "event_id": self.event_id,
            "event_type": self.event_type.value,
            "priority": self.priority.value,
            "timestamp": self.timestamp.isoformat(),
            "payload": self.payload,
            "source": self.source,
            "correlation_id": self.correlation_id,
            "causation_id": self.causation_id,
            "metadata": self.metadata,
            "retry_count": self.retry_count,
            "max_retries": self.max_retries,
            "timeout_seconds": self.timeout_seconds,
            "tags": self.tags,
        }

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Event":
        """Create event from dictionary"""
        return cls(
            event_id=data["event_id"],
            event_type=EventType(data["event_type"]),
            priority=EventPriority(data["priority"]),
            timestamp=datetime.fromisoformat(data["timestamp"]),
            payload=data["payload"],
            source=data.get("source", ""),
            correlation_id=data.get("correlation_id"),
            causation_id=data.get("causation_id"),
            metadata=data.get("metadata", {}),
            retry_count=data.get("retry_count", 0),
            max_retries=data.get("max_retries", 3),
            timeout_seconds=data.get("timeout_seconds", 60.0),
            tags=data.get("tags", {}),
        )


@dataclass
class HookExecutionEvent(Event):
    """Specialized event for hook execution"""

    hook_path: str = ""
    hook_event_type: Optional[HookEvent] = None
    execution_context: Dict[str, Any] = field(default_factory=dict)
    isolation_level: ResourceIsolationLevel = ResourceIsolationLevel.SHARED

    def __post_init__(self) -> None:
        """Post-initialization setup"""
        if self.event_type == EventType.HOOK_EXECUTION_REQUEST and not self.hook_path:
            raise ValueError("hook_path is required for hook execution requests")


@dataclass
class WorkflowEvent(Event):
    """Specialized event for workflow orchestration"""

    workflow_id: str = ""
    step_id: str = ""
    workflow_definition: Dict[str, Any] = field(default_factory=dict)
    execution_state: Dict[str, Any] = field(default_factory=dict)

    def __post_init__(self) -> None:
        """Post-initialization setup"""
        if self.event_type == EventType.WORKFLOW_ORCHESTRATION and not self.workflow_id:
            raise ValueError("workflow_id is required for workflow events")


class MessageBroker(ABC):
    """Abstract base class for message brokers"""

    @abstractmethod
    async def publish(self, topic: str, event: Event) -> bool:
        """Publish event to topic"""
        pass

    @abstractmethod
    async def subscribe(self, topic: str, callback: Callable[[Event], None]) -> str:
        """Subscribe to topic with callback"""
        pass

    @abstractmethod
    async def unsubscribe(self, subscription_id: str) -> bool:
        """Unsubscribe from topic"""
        pass

    @abstractmethod
    async def create_queue(self, queue_name: str, config: Dict[str, Any]) -> bool:
        """Create message queue with configuration"""
        pass

    @abstractmethod
    async def delete_queue(self, queue_name: str) -> bool:
        """Delete message queue"""
        pass

    @abstractmethod
    def get_stats(self) -> Dict[str, Any]:
        """Get message broker statistics"""
        pass


class InMemoryMessageBroker(MessageBroker):
    """In-memory message broker for development and testing"""

    def __init__(self, max_queue_size: int = 10000):
        self.max_queue_size = max_queue_size
        self.queues: Dict[str, deque] = defaultdict(lambda: deque(maxlen=max_queue_size))
        self.subscribers: Dict[str, List[tuple[str, Callable[[Event], None]]]] = defaultdict(list)
        self._lock = asyncio.Lock()
        self._stats = {
            "messages_published": 0,
            "messages_delivered": 0,
            "queues_created": 0,
            "active_subscriptions": 0,
            "failed_publishes": 0,
        }

    async def publish(self, topic: str, event: Event) -> bool:
        """Publish event to topic"""
        try:
            async with self._lock:
                if len(self.queues[topic]) >= self.max_queue_size:
                    # Remove oldest message if queue is full
                    self.queues[topic].popleft()

                self.queues[topic].append(event)
                self._stats["messages_published"] += 1

                # Notify subscribers
                for sub_id, callback in self.subscribers[topic]:
                    try:
                        # Create task for async callback
                        asyncio.create_task(self._safe_callback(callback, event))
                        self._stats["messages_delivered"] += 1
                    except Exception as e:
                        logger.error(f"Error in subscriber callback: {e}")

            return True
        except Exception as e:
            logger.error(f"Error publishing event to {topic}: {e}")
            self._stats["failed_publishes"] += 1
            return False

    async def _safe_callback(self, callback: Callable[[Event], None], event: Event) -> None:
        """Safely execute callback with error handling"""
        try:
            if asyncio.iscoroutinefunction(callback):
                result = callback(event)  # type: ignore[func-returns-value]
                if result is not None:
                    await result
            else:
                callback(event)
        except Exception as e:
            logger.error(f"Error in event callback: {e}")

    async def subscribe(self, topic: str, callback: Callable[[Event], None]) -> str:
        """Subscribe to topic with callback"""
        subscription_id = str(uuid.uuid4())
        async with self._lock:
            self.subscribers[topic].append((subscription_id, callback))
            self._stats["active_subscriptions"] += 1
        return subscription_id

    async def unsubscribe(self, subscription_id: str) -> bool:
        """Unsubscribe from topic"""
        async with self._lock:
            for topic in list(self.subscribers.keys()):
                subscribers = self.subscribers[topic]
                original_len = len(subscribers)
                self.subscribers[topic] = [
                    (sub_id, callback) for sub_id, callback in subscribers if sub_id != subscription_id
                ]
                if len(self.subscribers[topic]) < original_len:
                    self._stats["active_subscriptions"] -= 1
                    return True
        return False

    async def create_queue(self, queue_name: str, config: Dict[str, Any]) -> bool:
        """Create message queue with configuration"""
        async with self._lock:
            # Queue is created automatically on first use
            self.queues[queue_name] = deque(maxlen=config.get("max_size", self.max_queue_size))
            self._stats["queues_created"] += 1
        return True

    async def delete_queue(self, queue_name: str) -> bool:
        """Delete message queue"""
        async with self._lock:
            if queue_name in self.queues:
                del self.queues[queue_name]
            if queue_name in self.subscribers:
                del self.subscribers[queue_name]
        return True

    def get_stats(self) -> Dict[str, Any]:
        """Get message broker statistics"""
        return {
            **self._stats,
            "queue_count": len(self.queues),
            "total_queued_messages": sum(len(queue) for queue in self.queues.values()),
        }


class RedisMessageBroker(MessageBroker):
    """Redis-based message broker for production use"""

    def __init__(self, redis_url: str = "redis://localhost:6379/0"):
        self.redis_url = redis_url
        self._redis = None
        self._pubsub = None
        self._subscribers: Dict[str, List[tuple[str, Callable[[Event], None]]]] = defaultdict(list)
        self._stats = {
            "messages_published": 0,
            "messages_delivered": 0,
            "queues_created": 0,
            "active_subscriptions": 0,
            "failed_publishes": 0,
        }

    async def _connect(self) -> None:
        """Connect to Redis"""
        if self._redis is None:
            try:
                import redis.asyncio as redis_module

                redis_client = redis_module.from_url(self.redis_url)
                self._redis = redis_client
                self._pubsub = redis_client.pubsub()  # type: ignore[union-attr]
            except ImportError:
                raise ImportError("redis package is required for RedisMessageBroker")
            except Exception as e:
                raise ConnectionError(f"Failed to connect to Redis: {e}")

    async def publish(self, topic: str, event: Event) -> bool:
        """Publish event to Redis topic"""
        try:
            await self._connect()
            if self._redis is None:
                raise ConnectionError("Redis connection not established")
            message = json.dumps(event.to_dict())
            await self._redis.publish(topic, message)
            self._stats["messages_published"] += 1
            return True
        except Exception as e:
            logger.error(f"Error publishing event to Redis topic {topic}: {e}")
            self._stats["failed_publishes"] += 1
            return False

    async def subscribe(self, topic: str, callback: Callable[[Event], None]) -> str:
        """Subscribe to Redis topic"""
        try:
            await self._connect()
            if self._pubsub is None:
                raise ConnectionError("Redis pubsub not established")
            subscription_id = str(uuid.uuid4())

            # Add to local subscribers
            self._subscribers[topic].append((subscription_id, callback))

            # Subscribe to Redis pubsub
            await self._pubsub.subscribe(topic)

            # Start listener task
            asyncio.create_task(self._listen_to_topic(topic))

            self._stats["active_subscriptions"] += 1
            return subscription_id
        except Exception as e:
            logger.error(f"Error subscribing to Redis topic {topic}: {e}")
            return ""

    async def _listen_to_topic(self, topic: str) -> None:
        """Listen to Redis topic and call callbacks"""
        if self._pubsub is None:
            return
        try:
            async for message in self._pubsub.listen():
                if message["type"] == "message":
                    try:
                        event_data = json.loads(message["data"])
                        event = Event.from_dict(event_data)

                        # Call all subscribers for this topic
                        for subscription_id, callback in self._subscribers[topic]:
                            try:
                                if asyncio.iscoroutinefunction(callback):
                                    await callback(event)
                                else:
                                    callback(event)
                                self._stats["messages_delivered"] += 1
                            except Exception as e:
                                logger.error(f"Error in subscriber callback: {e}")
                    except Exception as e:
                        logger.error(f"Error processing Redis message: {e}")
        except Exception as e:
            logger.error(f"Error in Redis listener for topic {topic}: {e}")

    async def unsubscribe(self, subscription_id: str) -> bool:
        """Unsubscribe from Redis topic"""
        try:
            # Remove from local subscribers
            for topic, subscribers in self._subscribers.items():
                self._subscribers[topic] = [
                    (sub_id, callback) for sub_id, callback in subscribers if sub_id != subscription_id
                ]

            self._stats["active_subscriptions"] -= 1
            return True
        except Exception as e:
            logger.error(f"Error unsubscribing from Redis: {e}")
            return False

    async def create_queue(self, queue_name: str, config: Dict[str, Any]) -> bool:
        """Create Redis queue"""
        try:
            await self._connect()
            # Use Redis list as queue
            self._stats["queues_created"] += 1
            return True
        except Exception as e:
            logger.error(f"Error creating Redis queue {queue_name}: {e}")
            return False

    async def delete_queue(self, queue_name: str) -> bool:
        """Delete Redis queue"""
        try:
            await self._connect()
            if self._redis is None:
                raise ConnectionError("Redis connection not established")
            await self._redis.delete(queue_name)
            return True
        except Exception as e:
            logger.error(f"Error deleting Redis queue {queue_name}: {e}")
            return False

    def get_stats(self) -> Dict[str, Any]:
        """Get Redis message broker statistics"""
        return self._stats


class ResourcePool:
    """Resource pool for hook execution with isolation"""

    def __init__(self, isolation_level: ResourceIsolationLevel, max_concurrent: int = 10):
        self.isolation_level = isolation_level
        self.max_concurrent = max_concurrent
        self._pools: Dict[str, asyncio.Semaphore] = {}
        self._active_executions: Dict[str, Set[str]] = defaultdict(set)
        self._lock = asyncio.Lock()
        self._stats: Dict[str, Any] = {
            "total_executions": 0,
            "active_executions": 0,
            "pool_utilization": {},
            "isolation_violations": 0,
        }

    async def get_semaphore(self, hook_path: str, event_type: HookEvent) -> asyncio.Semaphore:
        """Get semaphore for hook based on isolation level"""
        pool_key = self._get_pool_key(hook_path, event_type)

        async with self._lock:
            if pool_key not in self._pools:
                self._pools[pool_key] = asyncio.Semaphore(self.max_concurrent)
            return self._pools[pool_key]

    def _get_pool_key(self, hook_path: str, event_type: HookEvent) -> str:
        """Generate pool key based on isolation level"""
        if self.isolation_level == ResourceIsolationLevel.SHARED:
            return "shared"
        elif self.isolation_level == ResourceIsolationLevel.TYPE_ISOLATED:
            return event_type.value
        elif self.isolation_level == ResourceIsolationLevel.PRIORITY_ISOLATED:
            # Extract priority from hook path or use default
            if "security" in hook_path or "validation" in hook_path:
                return "critical"
            elif "performance" in hook_path:
                return "high"
            else:
                return "normal"
        elif self.isolation_level == ResourceIsolationLevel.FULL_ISOLATION:
            return hook_path
        else:
            return "shared"

    async def acquire_execution_slot(self, hook_path: str, event_type: HookEvent) -> bool:
        """Acquire execution slot in appropriate pool"""
        pool_key = self._get_pool_key(hook_path, event_type)
        semaphore = await self.get_semaphore(hook_path, event_type)

        try:
            await semaphore.acquire()
            async with self._lock:
                self._active_executions[pool_key].add(hook_path)
                self._stats["active_executions"] += 1
                self._stats["total_executions"] += 1

                # Update pool utilization
                available = semaphore._value
                total = self.max_concurrent
                utilization = ((total - available) / total) * 100
                self._stats["pool_utilization"][pool_key] = utilization

            return True
        except Exception as e:
            logger.error(f"Error acquiring execution slot for {hook_path}: {e}")
            return False

    async def release_execution_slot(self, hook_path: str, event_type: HookEvent) -> None:
        """Release execution slot"""
        pool_key = self._get_pool_key(hook_path, event_type)
        semaphore = await self.get_semaphore(hook_path, event_type)

        async with self._lock:
            self._active_executions[pool_key].discard(hook_path)
            self._stats["active_executions"] = max(0, self._stats["active_executions"] - 1)

            # Update pool utilization
            available = semaphore._value
            total = self.max_concurrent
            utilization = ((total - available) / total) * 100
            self._stats["pool_utilization"][pool_key] = utilization

        semaphore.release()

    def get_stats(self) -> Dict[str, Any]:
        """Get resource pool statistics"""

        async def get_async_stats():
            async with self._lock:
                return {
                    **self._stats,
                    "pool_count": len(self._pools),
                    "active_pools": [
                        {
                            "pool_key": key,
                            "active_executions": len(executions),
                            "executions": list(executions),
                        }
                        for key, executions in self._active_executions.items()
                    ],
                }

        # Return synchronous version for compatibility
        try:
            loop = asyncio.get_event_loop()
            if loop.is_running():
                return self._stats
            else:
                return loop.run_until_complete(get_async_stats())
        except Exception:
            return self._stats


class EventProcessor:
    """Event processor for handling hook execution events"""

    def __init__(self, resource_pool: ResourcePool):
        self.resource_pool = resource_pool
        self._handlers: Dict[EventType, List[Callable]] = defaultdict(list)
        self._processing_stats: Dict[str, Any] = {
            "events_processed": 0,
            "events_failed": 0,
            "average_processing_time_ms": 0.0,
            "by_event_type": defaultdict(int),
        }

    def register_handler(self, event_type: EventType, handler: Callable) -> None:
        """Register event handler for event type"""
        self._handlers[event_type].append(handler)

    async def process_event(self, event: Event) -> bool:
        """Process event with appropriate handlers"""
        start_time = time.time()
        success = False

        try:
            handlers = self._handlers.get(event.event_type, [])
            if not handlers:
                logger.warning(f"No handlers registered for event type: {event.event_type}")
                return True

            # Process event with all registered handlers
            for handler in handlers:
                try:
                    if asyncio.iscoroutinefunction(handler):
                        await handler(event)
                    else:
                        handler(event)
                    success = True
                except Exception as e:
                    logger.error(f"Error in event handler for {event.event_type}: {e}")
                    success = False

            # Update statistics
            processing_time_ms = (time.time() - start_time) * 1000
            self._update_stats(event.event_type, processing_time_ms, success)

            return success

        except Exception as e:
            logger.error(f"Error processing event {event.event_id}: {e}")
            processing_time_ms = (time.time() - start_time) * 1000
            self._update_stats(event.event_type, processing_time_ms, False)
            return False

    def _update_stats(self, event_type: EventType, processing_time_ms: float, success: bool) -> None:
        """Update processing statistics"""
        events_processed: int = self._processing_stats["events_processed"]
        events_processed += 1
        self._processing_stats["events_processed"] = events_processed
        by_event_type: Dict[str, int] = self._processing_stats["by_event_type"]
        by_event_type[event_type.value] = by_event_type.get(event_type.value, 0) + 1

        if not success:
            events_failed: int = self._processing_stats["events_failed"]
            self._processing_stats["events_failed"] = events_failed + 1

        # Update average processing time
        total_events = events_processed
        current_avg: float = self._processing_stats["average_processing_time_ms"]
        self._processing_stats["average_processing_time_ms"] = (
            current_avg * (total_events - 1) + processing_time_ms
        ) / total_events

    def get_stats(self) -> Dict[str, Any]:
        """Get event processing statistics"""
        events_processed: int = self._processing_stats["events_processed"]
        events_failed: int = self._processing_stats["events_failed"]
        return {
            **self._processing_stats,
            "success_rate": ((events_processed - events_failed) / max(events_processed, 1)) * 100,
            "handlers_registered": {event_type.value: len(handlers) for event_type, handlers in self._handlers.items()},
        }


class EventDrivenHookSystem:
    """
    Event-Driven Hook System Architecture

    Enterprise-grade event-driven system for hook management with message queuing,
    resource isolation, and asynchronous event processing capabilities.
    """

    def __init__(
        self,
        message_broker_type: MessageBrokerType = MessageBrokerType.MEMORY,
        isolation_level: ResourceIsolationLevel = ResourceIsolationLevel.TYPE_ISOLATED,
        max_concurrent_hooks: int = 10,
        enable_persistence: bool = True,
        persistence_path: Optional[Path] = None,
        redis_url: str = "redis://localhost:6379/0",
    ):
        """Initialize Event-Driven Hook System

        Args:
            message_broker_type: Type of message broker to use
            isolation_level: Resource isolation level for hook execution
            max_concurrent_hooks: Maximum concurrent hook executions per pool
            enable_persistence: Enable event persistence for recovery
            persistence_path: Path for event persistence storage
            redis_url: Redis URL for Redis message broker
        """
        self.message_broker_type = message_broker_type
        self.isolation_level = isolation_level
        self.max_concurrent_hooks = max_concurrent_hooks
        self.enable_persistence = enable_persistence
        self.persistence_path = persistence_path or Path.cwd() / ".moai" / "cache" / "event_system"
        self.redis_url = redis_url

        # Initialize message broker
        self.message_broker = self._create_message_broker()

        # Initialize resource pool
        self.resource_pool = ResourcePool(isolation_level, max_concurrent_hooks)

        # Initialize event processor
        self.event_processor = EventProcessor(self.resource_pool)

        # System state
        self._running = False
        self._startup_time = datetime.now()
        self._event_loops: List[asyncio.Task] = []

        # Event persistence
        self._pending_events: Dict[str, Event] = {}
        self._processed_events: Set[str] = set()

        # System metrics
        self._system_metrics = {
            "events_published": 0,
            "events_processed": 0,
            "events_failed": 0,
            "hook_executions": 0,
            "system_uptime_seconds": 0,
            "average_event_latency_ms": 0.0,
        }

        # Setup persistence directory
        if self.enable_persistence:
            self.persistence_path.mkdir(parents=True, exist_ok=True)

    def _create_message_broker(self) -> MessageBroker:
        """Create message broker based on type"""
        if self.message_broker_type == MessageBrokerType.MEMORY:
            return InMemoryMessageBroker()
        elif self.message_broker_type == MessageBrokerType.REDIS:
            return RedisMessageBroker(self.redis_url)
        else:
            logger.warning(f"Message broker {self.message_broker_type} not implemented, using in-memory")
            return InMemoryMessageBroker()

    async def start(self) -> None:
        """Start the event-driven hook system"""
        if self._running:
            return

        logger.info("Starting Event-Driven Hook System...")

        try:
            # Load persisted events if enabled
            if self.enable_persistence:
                await self._load_persisted_events()

            # Register event handlers
            self._register_event_handlers()

            # Create message queues for different event types
            await self._setup_message_queues()

            # Start event processing loops
            await self._start_event_loops()

            self._running = True
            logger.info("Event-Driven Hook System started successfully")

        except Exception as e:
            logger.error(f"Error starting Event-Driven Hook System: {e}")
            raise

    async def stop(self) -> None:
        """Stop the event-driven hook system"""
        if not self._running:
            return

        logger.info("Stopping Event-Driven Hook System...")

        try:
            # Cancel event loops
            for loop_task in self._event_loops:
                loop_task.cancel()

            # Wait for loops to finish
            if self._event_loops:
                await asyncio.gather(*self._event_loops, return_exceptions=True)

            # Persist pending events if enabled
            if self.enable_persistence:
                await self._persist_events()

            self._running = False
            logger.info("Event-Driven Hook System stopped")

        except Exception as e:
            logger.error(f"Error stopping Event-Driven Hook System: {e}")

    def _register_event_handlers(self) -> None:
        """Register event handlers for different event types"""
        # Hook execution request handler
        self.event_processor.register_handler(EventType.HOOK_EXECUTION_REQUEST, self._handle_hook_execution_request)

        # Hook execution completion handler
        self.event_processor.register_handler(EventType.HOOK_EXECUTION_COMPLETED, self._handle_hook_execution_completed)

        # Hook execution failure handler
        self.event_processor.register_handler(EventType.HOOK_EXECUTION_FAILED, self._handle_hook_execution_failed)

        # System alert handler
        self.event_processor.register_handler(EventType.SYSTEM_ALERT, self._handle_system_alert)

        # Health check handler
        self.event_processor.register_handler(EventType.HEALTH_CHECK, self._handle_health_check)

        # Batch execution handler
        self.event_processor.register_handler(EventType.BATCH_EXECUTION_REQUEST, self._handle_batch_execution_request)

        # Workflow orchestration handler
        self.event_processor.register_handler(EventType.WORKFLOW_ORCHESTRATION, self._handle_workflow_orchestration)

    async def _setup_message_queues(self) -> None:
        """Setup message queues for different event types"""
        queue_configs = {
            "hook_execution_high": {
                "max_size": 1000,
                "priority": EventPriority.HIGH.value,
            },
            "hook_execution_normal": {
                "max_size": 5000,
                "priority": EventPriority.NORMAL.value,
            },
            "hook_execution_low": {
                "max_size": 10000,
                "priority": EventPriority.LOW.value,
            },
            "system_events": {
                "max_size": 1000,
                "priority": EventPriority.CRITICAL.value,
            },
            "analytics": {"max_size": 20000, "priority": EventPriority.BULK.value},
        }

        for queue_name, config in queue_configs.items():
            await self.message_broker.create_queue(queue_name, config)

    async def _start_event_loops(self) -> None:
        """Start event processing loops"""
        # Start main event processing loop
        event_loop = asyncio.create_task(self._event_processing_loop())
        self._event_loops.append(event_loop)

        # Start metrics collection loop
        metrics_loop = asyncio.create_task(self._metrics_collection_loop())
        self._event_loops.append(metrics_loop)

        # Start cleanup loop
        cleanup_loop = asyncio.create_task(self._cleanup_loop())
        self._event_loops.append(cleanup_loop)

        # Start persistence loop if enabled
        if self.enable_persistence:
            persistence_loop = asyncio.create_task(self._persistence_loop())
            self._event_loops.append(persistence_loop)

    async def _event_processing_loop(self) -> None:
        """Main event processing loop"""
        logger.info("Starting event processing loop")

        while self._running:
            try:
                # Process events from different queues based on priority
                await self._process_events_by_priority()
                await asyncio.sleep(0.1)  # Prevent busy waiting
            except Exception as e:
                logger.error(f"Error in event processing loop: {e}")
                await asyncio.sleep(1)

    async def _process_events_by_priority(self) -> None:
        """Process events by priority order"""
        # Process critical system events first
        await self._process_queue_events("system_events", EventPriority.CRITICAL)

        # Then process high priority hook executions
        await self._process_queue_events("hook_execution_high", EventPriority.HIGH)

        # Then normal priority
        await self._process_queue_events("hook_execution_normal", EventPriority.NORMAL)

        # Then low priority
        await self._process_queue_events("hook_execution_low", EventPriority.LOW)

        # Finally bulk analytics events
        await self._process_queue_events("analytics", EventPriority.BULK)

    async def _process_queue_events(self, queue_name: str, priority: EventPriority) -> None:
        """Process events from a specific queue"""
        # This is a simplified implementation
        # In a real implementation, you would poll the message broker
        pass

    async def _metrics_collection_loop(self) -> None:
        """Collect system metrics"""
        logger.info("Starting metrics collection loop")

        while self._running:
            try:
                # Update system metrics
                self._update_system_metrics()

                # Publish performance update event
                await self._publish_performance_update()

                await asyncio.sleep(30)  # Collect metrics every 30 seconds
            except Exception as e:
                logger.error(f"Error in metrics collection loop: {e}")
                await asyncio.sleep(30)

    async def _cleanup_loop(self) -> None:
        """Cleanup old events and resources"""
        logger.info("Starting cleanup loop")

        while self._running:
            try:
                # Clean up old processed events
                await self._cleanup_old_events()

                # Clean up completed workflow executions
                await self._cleanup_completed_workflows()

                await asyncio.sleep(300)  # Cleanup every 5 minutes
            except Exception as e:
                logger.error(f"Error in cleanup loop: {e}")
                await asyncio.sleep(300)

    async def _persistence_loop(self) -> None:
        """Persist events for recovery"""
        logger.info("Starting persistence loop")

        while self._running:
            try:
                await self._persist_events()
                await asyncio.sleep(60)  # Persist every minute
            except Exception as e:
                logger.error(f"Error in persistence loop: {e}")
                await asyncio.sleep(60)

    async def _handle_hook_execution_request(self, event: HookExecutionEvent) -> None:
        """Handle hook execution request event"""
        try:
            # Acquire execution slot from resource pool
            acquired = await self.resource_pool.acquire_execution_slot(
                event.hook_path, event.hook_event_type or HookEvent.SESSION_START
            )

            if not acquired:
                # Publish failed event
                failure_event = Event(
                    event_id=str(uuid.uuid4()),
                    event_type=EventType.HOOK_EXECUTION_FAILED,
                    priority=EventPriority.HIGH,
                    timestamp=datetime.now(),
                    payload={
                        "hook_path": event.hook_path,
                        "reason": "Resource pool full",
                        "original_event_id": event.event_id,
                    },
                    correlation_id=event.correlation_id,
                    causation_id=event.event_id,
                )
                await self.message_broker.publish("system_events", failure_event)
                return

            try:
                # Execute hook (this would integrate with existing hook manager)
                await self._execute_hook_event(event)

                # Publish completion event
                completion_event = Event(
                    event_id=str(uuid.uuid4()),
                    event_type=EventType.HOOK_EXECUTION_COMPLETED,
                    priority=EventPriority.NORMAL,
                    timestamp=datetime.now(),
                    payload={
                        "hook_path": event.hook_path,
                        "execution_time": 0.0,  # Would be filled by actual execution
                        "success": True,
                        "original_event_id": event.event_id,
                    },
                    correlation_id=event.correlation_id,
                    causation_id=event.event_id,
                )
                await self.message_broker.publish("hook_execution_normal", completion_event)

                self._system_metrics["hook_executions"] += 1

            finally:
                # Always release the execution slot
                await self.resource_pool.release_execution_slot(
                    event.hook_path, event.hook_event_type or HookEvent.SESSION_START
                )

        except Exception as e:
            logger.error(f"Error handling hook execution request: {e}")

            # Publish failure event
            failure_event = Event(
                event_id=str(uuid.uuid4()),
                event_type=EventType.HOOK_EXECUTION_FAILED,
                priority=EventPriority.HIGH,
                timestamp=datetime.now(),
                payload={
                    "hook_path": event.hook_path,
                    "reason": str(e),
                    "original_event_id": event.event_id,
                },
                correlation_id=event.correlation_id,
                causation_id=event.event_id,
            )
            await self.message_broker.publish("system_events", failure_event)

    async def _execute_hook_event(self, event: HookExecutionEvent) -> None:
        """Execute hook event (integration point with existing hook system)"""
        # This is where you would integrate with the existing JITEnhancedHookManager
        # For now, we'll simulate execution
        logger.info(f"Executing hook: {event.hook_path} with isolation: {event.isolation_level}")

        # Simulate execution time
        await asyncio.sleep(0.1)

        # In real implementation, you would call:
        # result = await self.hook_manager._execute_single_hook(event.hook_path, event.execution_context)

    async def _handle_hook_execution_completed(self, event: Event) -> None:
        """Handle hook execution completion event"""
        logger.debug(f"Hook execution completed: {event.payload}")
        self._system_metrics["events_processed"] += 1

    async def _handle_hook_execution_failed(self, event: Event) -> None:
        """Handle hook execution failure event"""
        logger.warning(f"Hook execution failed: {event.payload}")
        self._system_metrics["events_failed"] += 1

    async def _handle_system_alert(self, event: Event) -> None:
        """Handle system alert event"""
        logger.warning(f"System alert: {event.payload}")
        # Here you could trigger additional alerting mechanisms

    async def _handle_health_check(self, event: Event) -> None:
        """Handle health check event"""
        logger.debug(f"Health check: {event.payload}")
        # Respond with health status

    async def _handle_batch_execution_request(self, event: Event) -> None:
        """Handle batch execution request event"""
        logger.info(f"Batch execution request: {event.payload}")
        # Process multiple hooks in batch

    async def _handle_workflow_orchestration(self, event: WorkflowEvent) -> None:
        """Handle workflow orchestration event"""
        logger.info(f"Workflow orchestration: {event.workflow_id}")
        # Execute workflow steps

    async def _publish_performance_update(self) -> None:
        """Publish performance update event"""
        performance_event = Event(
            event_id=str(uuid.uuid4()),
            event_type=EventType.PERFORMANCE_UPDATE,
            priority=EventPriority.LOW,
            timestamp=datetime.now(),
            payload=self._system_metrics,
            source="event_system",
        )
        await self.message_broker.publish("analytics", performance_event)

    def _update_system_metrics(self) -> None:
        """Update system metrics"""
        self._system_metrics["system_uptime_seconds"] = (datetime.now() - self._startup_time).total_seconds()

    async def _cleanup_old_events(self) -> None:
        """Clean up old processed events"""
        cutoff_time = datetime.now() - timedelta(hours=24)

        # Remove old processed events
        old_events = [
            event_id
            for event_id in self._processed_events
            if event_id in self._pending_events and self._pending_events[event_id].timestamp < cutoff_time
        ]

        for event_id in old_events:
            self._processed_events.discard(event_id)
            if event_id in self._pending_events:
                del self._pending_events[event_id]

    async def _cleanup_completed_workflows(self) -> None:
        """Clean up completed workflow executions"""
        # Implementation would clean up completed workflow state
        pass

    async def _persist_events(self) -> None:
        """Persist events to disk for recovery"""
        if not self.enable_persistence:
            return

        try:
            # Persist pending events
            pending_file = self.persistence_path / "pending_events.json"
            pending_data = {event_id: event.to_dict() for event_id, event in self._pending_events.items()}

            with open(pending_file, "w") as f:
                json.dump(pending_data, f, indent=2)

            # Persist processed events
            processed_file = self.persistence_path / "processed_events.json"
            with open(processed_file, "w") as f:
                json.dump(list(self._processed_events), f, indent=2)

            # Persist system metrics
            metrics_file = self.persistence_path / "system_metrics.json"
            with open(metrics_file, "w") as f:
                json.dump(self._system_metrics, f, indent=2)

        except Exception as e:
            logger.error(f"Error persisting events: {e}")

    async def _load_persisted_events(self) -> None:
        """Load persisted events for recovery"""
        if not self.enable_persistence:
            return

        try:
            # Load pending events
            pending_file = self.persistence_path / "pending_events.json"
            if pending_file.exists():
                with open(pending_file, "r") as f:
                    pending_data = json.load(f)

                self._pending_events = {
                    event_id: Event.from_dict(event_data) for event_id, event_data in pending_data.items()
                }

            # Load processed events
            processed_file = self.persistence_path / "processed_events.json"
            if processed_file.exists():
                with open(processed_file, "r") as f:
                    processed_data = json.load(f)
                self._processed_events = set(processed_data)

            # Load system metrics
            metrics_file = self.persistence_path / "system_metrics.json"
            if metrics_file.exists():
                with open(metrics_file, "r") as f:
                    self._system_metrics = json.load(f)

        except Exception as e:
            logger.error(f"Error loading persisted events: {e}")

    async def publish_hook_execution_event(
        self,
        hook_path: str,
        event_type: HookEvent,
        execution_context: Dict[str, Any],
        isolation_level: Optional[ResourceIsolationLevel] = None,
        priority: EventPriority = EventPriority.NORMAL,
        correlation_id: Optional[str] = None,
    ) -> str:
        """Publish hook execution event"""

        event = HookExecutionEvent(
            event_id=str(uuid.uuid4()),
            event_type=EventType.HOOK_EXECUTION_REQUEST,
            priority=priority,
            timestamp=datetime.now(),
            payload={
                "hook_path": hook_path,
                "event_type": event_type.value,
                "execution_context": execution_context,
            },
            source="event_system",
            correlation_id=correlation_id or str(uuid.uuid4()),
            hook_path=hook_path,
            hook_event_type=event_type,
            execution_context=execution_context,
            isolation_level=isolation_level or self.isolation_level,
        )

        # Determine queue based on priority
        queue_name = self._get_queue_name_by_priority(priority)

        # Publish event
        success = await self.message_broker.publish(queue_name, event)

        if success:
            self._system_metrics["events_published"] += 1
            if self.enable_persistence:
                self._pending_events[event.event_id] = event

        return event.event_id

    def _get_queue_name_by_priority(self, priority: EventPriority) -> str:
        """Get queue name based on event priority"""
        if priority == EventPriority.CRITICAL:
            return "system_events"
        elif priority == EventPriority.HIGH:
            return "hook_execution_high"
        elif priority == EventPriority.NORMAL:
            return "hook_execution_normal"
        elif priority == EventPriority.LOW:
            return "hook_execution_low"
        else:
            return "analytics"

    async def publish_system_alert(
        self,
        alert_type: str,
        message: str,
        severity: EventPriority = EventPriority.HIGH,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> str:
        """Publish system alert event"""

        alert_event = Event(
            event_id=str(uuid.uuid4()),
            event_type=EventType.SYSTEM_ALERT,
            priority=severity,
            timestamp=datetime.now(),
            payload={
                "alert_type": alert_type,
                "message": message,
                "metadata": metadata or {},
            },
            source="event_system",
        )

        await self.message_broker.publish("system_events", alert_event)
        self._system_metrics["events_published"] += 1

        return alert_event.event_id

    async def get_system_status(self) -> Dict[str, Any]:
        """Get comprehensive system status"""
        return {
            "status": "running" if self._running else "stopped",
            "uptime_seconds": self._system_metrics["system_uptime_seconds"],
            "message_broker_type": self.message_broker_type.value,
            "isolation_level": self.isolation_level.value,
            "max_concurrent_hooks": self.max_concurrent_hooks,
            "persistence_enabled": self.enable_persistence,
            "system_metrics": self._system_metrics,
            "message_broker_stats": self.message_broker.get_stats(),
            "resource_pool_stats": self.resource_pool.get_stats(),
            "event_processor_stats": self.event_processor.get_stats(),
            "pending_events_count": len(self._pending_events),
            "processed_events_count": len(self._processed_events),
        }

    def get_event_flow_diagram(self) -> Dict[str, Any]:
        """Get event flow diagram for visualization"""
        return {
            "event_types": [event_type.value for event_type in EventType],
            "priorities": [priority.value for priority in EventPriority],
            "isolation_levels": [level.value for level in ResourceIsolationLevel],
            "message_broker_type": self.message_broker_type.value,
            "flow_pattern": {
                "hook_execution": [
                    "HOOK_EXECUTION_REQUEST -> Resource Pool -> Hook Execution -> HOOK_EXECUTION_COMPLETED"
                ],
                "system_alerts": ["SYSTEM_ALERT -> Alert Handlers -> Notification"],
                "batch_processing": ["BATCH_EXECUTION_REQUEST -> Batch Queue -> Parallel Execution -> Results"],
                "workflow_orchestration": [
                    "WORKFLOW_ORCHESTRATION -> Workflow Engine -> Step Execution -> State Update"
                ],
            },
        }


# Global instance for easy access
_event_system: Optional[EventDrivenHookSystem] = None


def get_event_system() -> EventDrivenHookSystem:
    """Get or create global event system instance"""
    global _event_system
    if _event_system is None:
        _event_system = EventDrivenHookSystem()
    return _event_system


# Convenience functions
async def start_event_system() -> None:
    """Start the event system"""
    system = get_event_system()
    await system.start()


async def stop_event_system() -> None:
    """Stop the event system"""
    system = get_event_system()
    await system.stop()


async def execute_hook_with_event_system(
    hook_path: str,
    event_type: HookEvent,
    execution_context: Dict[str, Any],
    isolation_level: Optional[ResourceIsolationLevel] = None,
    priority: EventPriority = EventPriority.NORMAL,
) -> str:
    """Execute hook using the event system"""
    system = get_event_system()
    return await system.publish_hook_execution_event(
        hook_path, event_type, execution_context, isolation_level, priority
    )


if __name__ == "__main__":
    # Example usage
    async def main():
        print("🚀 Starting Event-Driven Hook System Architecture...")

        # Initialize system with Redis broker and type isolation
        event_system = EventDrivenHookSystem(
            message_broker_type=MessageBrokerType.MEMORY,  # Use MEMORY for demo
            isolation_level=ResourceIsolationLevel.TYPE_ISOLATED,
            max_concurrent_hooks=5,
            enable_persistence=True,
        )

        try:
            # Start the system
            await event_system.start()

            # Publish some test hook execution events
            print("\n📤 Publishing hook execution events...")

            event_ids = []
            for i in range(3):
                event_id = await event_system.publish_hook_execution_event(
                    hook_path=f"test_hook_{i}.py",
                    event_type=HookEvent.SESSION_START,
                    execution_context={"test": True, "iteration": i},
                    isolation_level=ResourceIsolationLevel.FULL_ISOLATION,
                    priority=EventPriority.NORMAL,
                )
                event_ids.append(event_id)
                print(f"  Published event {i + 1}: {event_id}")

            # Publish a system alert
            alert_id = await event_system.publish_system_alert(
                alert_type="TEST_ALERT",
                message="This is a test alert from the event system",
                severity=EventPriority.HIGH,
            )
            print(f"\n🚨 Published system alert: {alert_id}")

            # Let events process
            print("\n⏳ Processing events...")
            await asyncio.sleep(2)

            # Get system status
            status = await event_system.get_system_status()
            print("\n📊 System Status:")
            print(f"  Status: {status['status']}")
            print(f"  Uptime: {status['uptime_seconds']:.1f}s")
            print(f"  Events Published: {status['system_metrics']['events_published']}")
            print(f"  Events Processed: {status['system_metrics']['events_processed']}")
            print(f"  Hook Executions: {status['system_metrics']['hook_executions']}")
            print(f"  Pending Events: {status['pending_events_count']}")

            # Get message broker stats
            broker_stats = status["message_broker_stats"]
            print("\n📨 Message Broker Stats:")
            print(f"  Messages Published: {broker_stats.get('messages_published', 0)}")
            print(f"  Messages Delivered: {broker_stats.get('messages_delivered', 0)}")

            # Get resource pool stats
            pool_stats = status["resource_pool_stats"]
            print("\n🏊 Resource Pool Stats:")
            print(f"  Total Executions: {pool_stats.get('total_executions', 0)}")
            print(f"  Active Executions: {pool_stats.get('active_executions', 0)}")

            # Get event flow diagram
            flow_diagram = event_system.get_event_flow_diagram()
            print("\n🌊 Event Flow Diagram:")
            print(f"  Event Types: {len(flow_diagram['event_types'])}")
            print(f"  Isolation Levels: {flow_diagram['isolation_levels']}")
            print(f"  Message Broker: {flow_diagram['message_broker_type']}")

            print("\n✅ Event-Driven Hook System demo completed successfully!")

        except Exception as e:
            print(f"\n❌ Demo failed: {str(e)}")
            import traceback

            traceback.print_exc()

        finally:
            # Stop the system
            print("\n🛑 Stopping event system...")
            await event_system.stop()
            print("✅ System stopped")

    # Run the demo
    asyncio.run(main())
