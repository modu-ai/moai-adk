# Rate Limiting Patterns

## Overview

Enterprise rate limiting strategies for API protection.

## Implementation Patterns

### Token Bucket Algorithm

```python
from redis import Redis
from datetime import datetime

class TokenBucket:
    def __init__(self, rate: int, capacity: int):
        self.rate = rate  # Tokens per second
        self.capacity = capacity
        self.redis = Redis()

    def consume(self, user_id: str, tokens: int = 1) -> bool:
        key = f"rate_limit:{user_id}"

        # Get current tokens
        current = self.redis.get(key)
        if current is None:
            current = self.capacity
        else:
            current = int(current)

        # Refill tokens
        last_refill = self.redis.get(f"{key}:last_refill")
        if last_refill:
            elapsed = datetime.now().timestamp() - float(last_refill)
            refill = min(self.capacity, current + int(elapsed * self.rate))
            current = refill

        # Consume tokens
        if current >= tokens:
            self.redis.set(key, current - tokens)
            self.redis.set(f"{key}:last_refill", datetime.now().timestamp())
            return True

        return False
```

### Sliding Window

```python
from collections import deque
from time import time

class SlidingWindow:
    def __init__(self, limit: int, window: int):
        self.limit = limit
        self.window = window  # seconds
        self.requests = deque()

    def allow_request(self) -> bool:
        now = time()

        # Remove old requests
        while self.requests and self.requests[0] < now - self.window:
            self.requests.popleft()

        # Check limit
        if len(self.requests) < self.limit:
            self.requests.append(now)
            return True

        return False
```

---
**Last Updated**: 2025-11-23
**Status**: Production Ready
