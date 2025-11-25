# Claude Code Memory System - Practical Examples

## Example 1: Session Memory Storage

```python
# memory/session_store.py
from moai_cc_memory import MemoryStore, MemoryKey

class SessionMemory(MemoryStore):
    """User session memory"""

    async def save_user_context(self, user_id: str, context: dict):
        """Save user context"""

        key = MemoryKey(
            scope="session",
            user_id=user_id,
            component="context"
        )

        await self.set(key, context, ttl=3600)  # 1 hour

    async def load_user_context(self, user_id: str) -> dict:
        """Load user context"""

        key = MemoryKey(
            scope="session",
            user_id=user_id,
            component="context"
        )

        return await self.get(key)
```

## Example 2: Token Budget Tracking

```python
# memory/token_budget.py
from moai_cc_memory import TokenBudgetTracker

class ContextualTokenBudget(TokenBudgetTracker):
    """Contextual token budget tracking"""

    async def track_request(self, request_id: str, tokens_used: int):
        """Track tokens per request"""

        # Check budget
        remaining = await self.get_remaining_budget()

        if tokens_used > remaining:
            raise BudgetExceededError(
                f"Token exceeded: {tokens_used} > {remaining}"
            )

        # Record token usage
        await self.record_usage(
            request_id=request_id,
            tokens=tokens_used,
            timestamp=datetime.now()
        )
```

## Example 3: Conversation History Management

```python
# memory/conversation_history.py
from moai_cc_memory import ConversationMemory

class ContextAwareConversation(ConversationMemory):
    """Context-aware conversation memory"""

    async def add_message(self, role: str, content: str, context: dict = None):
        """Add message"""

        message = {
            "role": role,
            "content": content,
            "context": context or {},
            "timestamp": datetime.now()
        }

        await self.append(message)

        # Length limit (keep only recent 50 messages)
        if len(self.history) > 50:
            await self.trim_old_messages(keep=50)

    async def get_context_summary(self) -> str:
        """Get current context summary"""

        # Extract context from recent messages
        recent = self.history[-10:]

        summary = "\n".join([
            f"{msg['role']}: {msg['content'][:100]}"
            for msg in recent
        ])

        return summary
```

## Example 4: Cache Layer

```python
# memory/cache_layer.py
from moai_cc_memory import MemoryCache

class LRUMemoryCache(MemoryCache):
    """LRU cache implementation"""

    def __init__(self, max_size: int = 1000):
        super().__init__()
        self.max_size = max_size
        self.access_times = {}

    async def get(self, key: str):
        """Retrieve from cache"""

        value = await super().get(key)

        if value is not None:
            # Update access time
            self.access_times[key] = datetime.now()

        return value

    async def set(self, key: str, value):
        """Store in cache"""

        # Remove oldest item if size exceeded
        if len(self._data) >= self.max_size:
            oldest_key = min(
                self.access_times,
                key=self.access_times.get
            )
            await self.delete(oldest_key)

        await super().set(key, value)
        self.access_times[key] = datetime.now()
```

## Example 5: State Machine Memory

```python
# memory/state_machine.py
from moai_cc_memory import StatefulMemory

class WorkflowState(StatefulMemory):
    """Workflow state memory"""

    async def transition(self, state: str, data: dict = None):
        """State transition"""

        # Verify valid transition
        if not self.is_valid_transition(self.current_state, state):
            raise InvalidTransitionError(
                f"{self.current_state} -> {state} not allowed"
            )

        # Save state
        await self.set_state(state, data or {})

        # Record transition history
        await self.record_transition(
            from_state=self.current_state,
            to_state=state,
            timestamp=datetime.now()
        )
```

## Example 6: Distributed Memory Synchronization

```python
# memory/distributed_sync.py
from moai_cc_memory import DistributedMemory

class RedisBackedMemory(DistributedMemory):
    """Redis-backed distributed memory"""

    async def publish_state(self, channel: str, state: dict):
        """Publish state"""

        await self.redis.publish(
            channel,
            json.dumps(state)
        )

    async def subscribe_to_state(self, channel: str, callback):
        """Subscribe to state"""

        async def on_message(message):
            state = json.loads(message['data'])
            await callback(state)

        await self.redis.subscribe(channel, callback=on_message)
```

---

**Last Updated**: 2025-11-22
**Status**: Production Ready
