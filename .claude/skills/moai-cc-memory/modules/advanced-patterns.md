# Advanced Patterns - Claude Code Memory

## Pattern 1: Multi-Layer Memory Architecture

Implement tiered memory system (L1 hot cache, L2 warm storage, L3 cold archive).

```python
from typing import Dict, Any, Optional, Callable
from enum import Enum
import time

class MemoryTier(Enum):
    """Memory storage tiers."""
    L1_HOT = "l1_hot"       # In-memory cache, very fast
    L2_WARM = "l2_warm"     # Disk cache, moderate speed
    L3_COLD = "l3_cold"     # Archive, slow but persistent

@dataclass
class MemoryEntry:
    """Entry in memory system."""
    key: str
    value: Any
    tier: MemoryTier
    timestamp: float
    access_count: int = 0
    ttl: Optional[int] = None

class TieredMemorySystem:
    """Multi-tier memory with automatic promotion/demotion."""

    def __init__(
        self,
        l1_size_mb: int = 512,
        l2_path: str = './.moai/cache',
        l3_path: str = './.moai/archive'
    ):
        self.l1: Dict[str, MemoryEntry] = {}  # Hot cache
        self.l1_size_bytes = l1_size_mb * 1024 * 1024
        self.l1_current_bytes = 0

        self.l2_path = l2_path  # Warm storage
        self.l3_path = l3_path  # Cold archive

        Path(self.l2_path).mkdir(parents=True, exist_ok=True)
        Path(self.l3_path).mkdir(parents=True, exist_ok=True)

    def put(
        self,
        key: str,
        value: Any,
        ttl_seconds: Optional[int] = None
    ):
        """Put value in memory system."""

        entry_size = len(pickle.dumps(value))

        # Check if value fits in L1
        if self.l1_current_bytes + entry_size <= self.l1_size_bytes:
            self.l1[key] = MemoryEntry(
                key=key,
                value=value,
                tier=MemoryTier.L1_HOT,
                timestamp=time.time(),
                ttl=ttl_seconds
            )
            self.l1_current_bytes += entry_size
        else:
            # Move to L2
            self._save_to_l2(key, value, ttl_seconds)

    def get(self, key: str) -> Optional[Any]:
        """Get value from memory system."""

        # Try L1
        if key in self.l1:
            entry = self.l1[key]

            # Check expiration
            if entry.ttl and time.time() - entry.timestamp > entry.ttl:
                del self.l1[key]
                return None

            entry.access_count += 1
            return entry.value

        # Try L2
        value = self._load_from_l2(key)
        if value is not None:
            # Promote to L1 on access
            self.put(key, value)
            return value

        # Try L3
        return self._load_from_l3(key)

    def _save_to_l2(self, key: str, value: Any, ttl: Optional[int]):
        """Save to warm storage (disk)."""
        filepath = Path(self.l2_path) / f"{key}.pkl"
        with open(filepath, 'wb') as f:
            pickle.dump((value, ttl, time.time()), f)

    def _load_from_l2(self, key: str) -> Optional[Any]:
        """Load from warm storage."""
        filepath = Path(self.l2_path) / f"{key}.pkl"
        if filepath.exists():
            with open(filepath, 'rb') as f:
                value, ttl, timestamp = pickle.load(f)
                if ttl and time.time() - timestamp > ttl:
                    filepath.unlink()
                    return None
                return value
        return None

    def _save_to_l3(self, key: str, value: Any):
        """Archive to cold storage."""
        # Compress and store in archive
        pass

    def get_stats(self) -> Dict[str, Any]:
        """Get memory system statistics."""
        return {
            'l1_entries': len(self.l1),
            'l1_usage_mb': self.l1_current_bytes / (1024 * 1024),
            'l1_capacity_mb': self.l1_size_bytes / (1024 * 1024),
            'l2_files': len(list(Path(self.l2_path).glob('*.pkl'))),
            'l3_files': len(list(Path(self.l3_path).glob('*.pkl'))),
        }
```

## Pattern 2: Session State Management with Snapshots

Track and restore session state with point-in-time snapshots.

```python
from dataclasses import dataclass, field
import json

@dataclass
class SessionSnapshot:
    """Point-in-time snapshot of session state."""
    timestamp: float
    session_id: str
    state: Dict[str, Any]
    metadata: Dict[str, Any] = field(default_factory=dict)
    version: int = 1

class SessionStateManager:
    """Manage session state with snapshots."""

    def __init__(self, session_id: str):
        self.session_id = session_id
        self.state: Dict[str, Any] = {}
        self.snapshots: List[SessionSnapshot] = []
        self.max_snapshots = 100

    def set_state(self, key: str, value: Any):
        """Set session state."""
        self.state[key] = value

    def get_state(self, key: str, default: Any = None) -> Any:
        """Get session state."""
        return self.state.get(key, default)

    def create_snapshot(self, metadata: Dict[str, Any] = None) -> SessionSnapshot:
        """Create session state snapshot."""

        snapshot = SessionSnapshot(
            timestamp=time.time(),
            session_id=self.session_id,
            state=self.state.copy(),
            metadata=metadata or {}
        )

        self.snapshots.append(snapshot)

        # Keep only recent snapshots
        if len(self.snapshots) > self.max_snapshots:
            self.snapshots = self.snapshots[-self.max_snapshots:]

        return snapshot

    def restore_snapshot(self, snapshot_index: int = -1) -> bool:
        """Restore to previous snapshot."""

        if snapshot_index >= len(self.snapshots):
            return False

        snapshot = self.snapshots[snapshot_index]
        self.state = snapshot.state.copy()

        return True

    def get_snapshot_history(self) -> List[Dict]:
        """Get history of snapshots."""

        return [
            {
                'timestamp': s.timestamp,
                'snapshot_index': i,
                'metadata': s.metadata,
                'state_keys': list(s.state.keys())
            }
            for i, s in enumerate(self.snapshots)
        ]

    def save_snapshots(self, filepath: str):
        """Save snapshots to disk."""

        data = [
            {
                'timestamp': s.timestamp,
                'session_id': s.session_id,
                'state': s.state,
                'metadata': s.metadata,
                'version': s.version
            }
            for s in self.snapshots
        ]

        with open(filepath, 'w') as f:
            json.dump(data, f, indent=2, default=str)

    def load_snapshots(self, filepath: str):
        """Load snapshots from disk."""

        with open(filepath, 'r') as f:
            data = json.load(f)

        for item in data:
            self.snapshots.append(SessionSnapshot(
                timestamp=item['timestamp'],
                session_id=item['session_id'],
                state=item['state'],
                metadata=item['metadata'],
                version=item.get('version', 1)
            ))
```

## Pattern 3: Event-Driven State Updates

Track state changes through event streams with audit trail.

```python
from abc import ABC, abstractmethod
from dataclasses import dataclass

@dataclass
class StateChangeEvent:
    """Event representing state change."""
    event_id: str
    timestamp: float
    key: str
    old_value: Any
    new_value: Any
    source: str
    metadata: Dict[str, Any]

class EventListener(ABC):
    """Base class for state change listeners."""

    @abstractmethod
    async def on_state_change(self, event: StateChangeEvent):
        """Handle state change event."""
        pass

class EventDrivenStateManager:
    """Manage state with event-driven updates."""

    def __init__(self):
        self.state: Dict[str, Any] = {}
        self.listeners: List[EventListener] = []
        self.event_log: List[StateChangeEvent] = []

    def subscribe(self, listener: EventListener):
        """Subscribe to state changes."""
        self.listeners.append(listener)

    def unsubscribe(self, listener: EventListener):
        """Unsubscribe from state changes."""
        self.listeners.remove(listener)

    async def set_state(
        self,
        key: str,
        value: Any,
        source: str = 'direct',
        metadata: Dict[str, Any] = None
    ):
        """Set state and emit event."""

        old_value = self.state.get(key)
        self.state[key] = value

        # Create event
        event = StateChangeEvent(
            event_id=str(uuid.uuid4()),
            timestamp=time.time(),
            key=key,
            old_value=old_value,
            new_value=value,
            source=source,
            metadata=metadata or {}
        )

        # Log event
        self.event_log.append(event)

        # Notify listeners
        for listener in self.listeners:
            await listener.on_state_change(event)

    def get_audit_trail(
        self,
        key: Optional[str] = None,
        limit: int = 100
    ) -> List[StateChangeEvent]:
        """Get audit trail of state changes."""

        events = self.event_log

        if key:
            events = [e for e in events if e.key == key]

        return events[-limit:]
```

## Pattern 4: Distributed Session Synchronization

Synchronize session state across multiple processes/servers.

```python
from typing import Coroutine
import asyncio

class DistributedSessionSync:
    """Synchronize session state across distributed nodes."""

    def __init__(self, node_id: str):
        self.node_id = node_id
        self.local_state: Dict[str, Any] = {}
        self.peers: List[str] = []
        self.sync_queue: asyncio.Queue = asyncio.Queue()

    async def add_peer(self, peer_id: str):
        """Register peer node."""
        self.peers.append(peer_id)

    async def broadcast_state_change(
        self,
        key: str,
        value: Any
    ):
        """Broadcast state change to all peers."""

        message = {
            'type': 'state_change',
            'source_node': self.node_id,
            'key': key,
            'value': value,
            'timestamp': time.time()
        }

        # Add to sync queue
        await self.sync_queue.put(message)

        # Send to all peers
        for peer_id in self.peers:
            await self._send_to_peer(peer_id, message)

    async def _send_to_peer(self, peer_id: str, message: Dict):
        """Send message to peer."""
        # Implementation depends on transport (gRPC, HTTP, etc.)
        pass

    async def sync_worker(self):
        """Background worker for state synchronization."""

        while True:
            message = await self.sync_queue.get()

            try:
                # Apply state change
                if message['type'] == 'state_change':
                    self.local_state[message['key']] = message['value']

                # Acknowledge
                await self._send_ack(message['source_node'], message)

            except Exception as e:
                # Handle sync error
                print(f"Sync error: {e}")

    async def _send_ack(self, node_id: str, message: Dict):
        """Send acknowledgment."""
        pass
```

## Pattern 5: Memory Compression and Decompression

Compress memory state for efficient storage and transmission.

```python
import pickle
import zlib

class CompressedMemory:
    """Compress memory state automatically."""

    def __init__(self, compression_level: int = 9):
        self.compression_level = compression_level
        self.compressed_data: Dict[str, bytes] = {}

    def store_compressed(self, key: str, value: Any) -> int:
        """Store value compressed."""

        # Serialize
        serialized = pickle.dumps(value)

        # Compress
        compressed = zlib.compress(serialized, self.compression_level)

        # Store
        self.compressed_data[key] = compressed

        # Return compression ratio
        return len(compressed) / len(serialized)

    def retrieve_decompressed(self, key: str) -> Optional[Any]:
        """Retrieve and decompress value."""

        if key not in self.compressed_data:
            return None

        compressed = self.compressed_data[key]

        # Decompress
        decompressed = zlib.decompress(compressed)

        # Deserialize
        return pickle.loads(decompressed)

    def get_compression_stats(self) -> Dict[str, Any]:
        """Get compression statistics."""

        total_compressed = sum(
            len(data) for data in self.compressed_data.values()
        )

        return {
            'total_compressed_bytes': total_compressed,
            'total_compressed_mb': total_compressed / (1024 * 1024),
            'entries': len(self.compressed_data),
            'avg_compressed_bytes': total_compressed / len(self.compressed_data) if self.compressed_data else 0
        }
```

## Pattern 6: Memory Leak Detection

Monitor and detect memory leaks in session state.

```python
from weakref import WeakSet
import sys

class MemoryLeakDetector:
    """Detect memory leaks in session state."""

    def __init__(self):
        self.tracked_objects: WeakSet = WeakSet()
        self.baseline_memory: int = 0

    def track_object(self, obj: Any):
        """Track object for leak detection."""
        self.tracked_objects.add(obj)

    def set_baseline(self):
        """Set memory baseline."""
        self.baseline_memory = self.get_memory_usage()

    def get_memory_usage(self) -> int:
        """Get current memory usage."""
        return sum(
            sys.getsizeof(obj) for obj in self.tracked_objects
        )

    def detect_leaks(self, threshold_mb: float = 100.0) -> List[Dict]:
        """Detect potential memory leaks."""

        current_memory = self.get_memory_usage()
        memory_growth = current_memory - self.baseline_memory

        leaks = []

        if memory_growth > threshold_mb * 1024 * 1024:
            leaks.append({
                'type': 'memory_growth',
                'baseline_mb': self.baseline_memory / (1024 * 1024),
                'current_mb': current_memory / (1024 * 1024),
                'growth_mb': memory_growth / (1024 * 1024),
                'threshold_mb': threshold_mb
            })

        return leaks

    def get_object_breakdown(self) -> Dict[str, int]:
        """Get memory usage by object type."""

        breakdown = {}

        for obj in self.tracked_objects:
            obj_type = type(obj).__name__
            size = sys.getsizeof(obj)

            breakdown[obj_type] = breakdown.get(obj_type, 0) + size

        return breakdown
```

## Pattern 7: Context Variable Management

Manage request-local context variables across async boundaries.

```python
from contextvars import ContextVar
from typing import Optional, Dict, Any

class AsyncContextManager:
    """Manage context variables across async calls."""

    def __init__(self):
        self.request_id: ContextVar[str] = ContextVar('request_id')
        self.user_id: ContextVar[Optional[str]] = ContextVar('user_id', default=None)
        self.session_data: ContextVar[Dict[str, Any]] = ContextVar('session_data', default={})

    def set_request_context(
        self,
        request_id: str,
        user_id: Optional[str] = None,
        session_data: Dict[str, Any] = None
    ):
        """Set request context."""
        self.request_id.set(request_id)
        self.user_id.set(user_id)
        self.session_data.set(session_data or {})

    def get_request_id(self) -> str:
        """Get current request ID."""
        return self.request_id.get()

    def get_user_id(self) -> Optional[str]:
        """Get current user ID."""
        return self.user_id.get()

    def get_session_data(self) -> Dict[str, Any]:
        """Get current session data."""
        return self.session_data.get()

    def set_session_value(self, key: str, value: Any):
        """Set value in session context."""
        data = self.session_data.get()
        data[key] = value
        self.session_data.set(data)

    async def run_with_context(
        self,
        coro,
        request_id: str,
        user_id: Optional[str] = None,
        session_data: Dict[str, Any] = None
    ):
        """Run coroutine with request context."""
        self.set_request_context(request_id, user_id, session_data)
        return await coro

# Usage
context_manager = AsyncContextManager()

async def handle_request():
    # Context is available throughout async call chain
    request_id = context_manager.get_request_id()
    user_id = context_manager.get_user_id()
    # Use context variables...

await context_manager.run_with_context(
    handle_request(),
    request_id='req-123',
    user_id='user-456'
)
```

---

**Advanced Patterns Summary**: 7 enterprise patterns for tiered memory architecture, session snapshots, event-driven updates, distributed sync, compression, leak detection, and context management.

