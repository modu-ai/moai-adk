# Rate Limiting Patterns

Rate limiting strategies for protecting MCP servers.

---

## Token Bucket Algorithm

```python
from collections import defaultdict
from datetime import datetime, timedelta
import time

class TokenBucket:
    """Token bucket rate limiter."""

    def __init__(self, capacity: float, refill_rate: float):
        """
        Args:
            capacity: Max tokens (requests per window)
            refill_rate: Tokens per second
        """
        self.capacity = capacity
        self.refill_rate = refill_rate
        self.tokens = capacity
        self.last_refill = datetime.now()

    def consume(self, tokens: int = 1) -> bool:
        """Try to consume tokens."""
        self._refill()

        if self.tokens >= tokens:
            self.tokens -= tokens
            return True

        return False

    def _refill(self):
        """Refill tokens based on elapsed time."""
        now = datetime.now()
        elapsed = (now - self.last_refill).total_seconds()
        tokens_to_add = elapsed * self.refill_rate

        self.tokens = min(self.capacity, self.tokens + tokens_to_add)
        self.last_refill = now

# Global rate limiter: 100 requests per minute
global_limiter = TokenBucket(capacity=100, refill_rate=100/60)

@server.tool()
def rate_limited_operation(param: str) -> dict:
    """Operation with rate limiting."""
    if not global_limiter.consume():
        raise ValueError(
            "Rate limit exceeded. "
            "Maximum 100 requests per minute."
        )

    return execute_operation(param)
```

---

## Per-User Rate Limiting

```python
class PerUserRateLimiter:
    """Rate limiting per user."""

    def __init__(self, requests_per_minute: int = 100):
        self.limiters = {}
        self.requests_per_minute = requests_per_minute
        self.refill_rate = requests_per_minute / 60

    def is_allowed(self, user_id: str) -> bool:
        """Check if user is within rate limit."""
        if user_id not in self.limiters:
            self.limiters[user_id] = TokenBucket(
                capacity=self.requests_per_minute,
                refill_rate=self.refill_rate
            )

        return self.limiters[user_id].consume()

user_limiter = PerUserRateLimiter(requests_per_minute=100)

@server.tool()
def user_rate_limited_operation(user_id: str, param: str) -> dict:
    """Per-user rate limited operation."""
    if not user_limiter.is_allowed(user_id):
        raise ValueError(
            f"Rate limit exceeded for user {user_id}. "
            f"Maximum 100 requests per minute."
        )

    return execute_operation(user_id, param)
```

---

## Sliding Window Counter

```python
class SlidingWindowRateLimiter:
    """Sliding window rate limiting."""

    def __init__(self, max_requests: int, window_seconds: int):
        self.max_requests = max_requests
        self.window = window_seconds
        self.requests = defaultdict(list)

    def is_allowed(self, identifier: str) -> bool:
        """Check if request is allowed."""
        now = datetime.now()
        cutoff = now - timedelta(seconds=self.window)

        # Remove old requests outside window
        self.requests[identifier] = [
            ts for ts in self.requests[identifier]
            if ts > cutoff
        ]

        # Check limit
        if len(self.requests[identifier]) < self.max_requests:
            self.requests[identifier].append(now)
            return True

        return False

# Example: 10 requests per minute per IP
ip_limiter = SlidingWindowRateLimiter(max_requests=10, window_seconds=60)

@server.tool()
def ip_rate_limited_operation(client_ip: str, param: str) -> dict:
    """Rate limited by IP address."""
    if not ip_limiter.is_allowed(client_ip):
        raise ValueError(
            f"Rate limit exceeded for IP {client_ip}. "
            f"Maximum 10 requests per minute."
        )

    return execute_operation(param)
```

---

## Tiered Rate Limiting

```python
class TieredRateLimiter:
    """Different limits for different tiers."""

    def __init__(self):
        self.tiers = {
            "free": {"requests_per_minute": 10, "requests_per_day": 1000},
            "pro": {"requests_per_minute": 100, "requests_per_day": 100000},
            "enterprise": {"requests_per_minute": 10000, "requests_per_day": 1000000}
        }
        self.user_tiers = {}
        self.limiters = defaultdict(dict)

    def set_user_tier(self, user_id: str, tier: str):
        """Set user's tier."""
        if tier not in self.tiers:
            raise ValueError(f"Unknown tier: {tier}")
        self.user_tiers[user_id] = tier

    def is_allowed(self, user_id: str) -> bool:
        """Check if user is within their tier limits."""
        tier = self.user_tiers.get(user_id, "free")
        limits = self.tiers[tier]

        minute_key = f"{user_id}:minute"
        day_key = f"{user_id}:day"

        # Create limiters if needed
        if minute_key not in self.limiters[user_id]:
            self.limiters[user_id][minute_key] = TokenBucket(
                capacity=limits["requests_per_minute"],
                refill_rate=limits["requests_per_minute"] / 60
            )
        if day_key not in self.limiters[user_id]:
            self.limiters[user_id][day_key] = TokenBucket(
                capacity=limits["requests_per_day"],
                refill_rate=limits["requests_per_day"] / 86400
            )

        # Check both limits
        minute_ok = self.limiters[user_id][minute_key].consume()
        day_ok = self.limiters[user_id][day_key].consume()

        return minute_ok and day_ok

tiered_limiter = TieredRateLimiter()

@server.tool()
def tiered_rate_limited_operation(user_id: str, param: str) -> dict:
    """Operation with tier-based rate limiting."""
    if not tiered_limiter.is_allowed(user_id):
        tier = tiered_limiter.user_tiers.get(user_id, "free")
        limits = tiered_limiter.tiers[tier]

        raise ValueError(
            f"Rate limit exceeded for {tier} tier. "
            f"Maximum {limits['requests_per_minute']} requests per minute, "
            f"{limits['requests_per_day']} per day."
        )

    return execute_operation(user_id, param)
```

---

## Redis-Based Rate Limiting

```python
import redis

class RedisRateLimiter:
    """Distributed rate limiting using Redis."""

    def __init__(self, redis_client: redis.Redis):
        self.redis = redis_client

    def is_allowed(
        self,
        identifier: str,
        max_requests: int,
        window_seconds: int
    ) -> bool:
        """Check rate limit using Redis."""
        key = f"rate_limit:{identifier}"

        # Increment counter
        current = self.redis.incr(key)

        # Set expiry on first request
        if current == 1:
            self.redis.expire(key, window_seconds)

        return current <= max_requests

    def get_remaining(
        self,
        identifier: str,
        max_requests: int
    ) -> int:
        """Get remaining requests."""
        key = f"rate_limit:{identifier}"
        current = int(self.redis.get(key) or 0)
        return max(0, max_requests - current)

redis_client = redis.Redis(host="localhost", port=6379)
redis_limiter = RedisRateLimiter(redis_client)

@server.tool()
def distributed_rate_limited_operation(user_id: str, param: str) -> dict:
    """Operation with distributed rate limiting."""
    if not redis_limiter.is_allowed(user_id, max_requests=100, window_seconds=60):
        remaining = redis_limiter.get_remaining(user_id, max_requests=100)

        raise ValueError(
            f"Rate limit exceeded for user {user_id}. "
            f"Remaining requests: {remaining}/100"
        )

    return execute_operation(user_id, param)
```

---

## Rate Limit Headers

### Response Headers

```python
@server.tool()
def operation_with_headers(user_id: str, param: str) -> dict:
    """Operation that includes rate limit headers."""
    result = execute_operation(user_id, param)

    # Get rate limit info
    remaining = redis_limiter.get_remaining(user_id, max_requests=100)
    reset_time = redis_limiter.get_reset_time(user_id)

    return {
        "status": "success",
        "data": result,
        "rate_limit": {
            "limit": 100,
            "remaining": remaining,
            "reset": reset_time
        }
    }
```

---

## Quota Management

```python
class QuotaManager:
    """Manage request quotas."""

    def __init__(self, redis_client: redis.Redis):
        self.redis = redis_client

    def add_quota(self, user_id: str, quota: int, days: int = 30):
        """Add quota to user."""
        key = f"quota:{user_id}"
        self.redis.set(key, quota)
        self.redis.expire(key, days * 86400)

    def use_quota(self, user_id: str, amount: int = 1) -> int:
        """Use quota. Returns remaining."""
        key = f"quota:{user_id}"
        remaining = self.redis.decrby(key, amount)
        return max(0, remaining)

    def get_quota(self, user_id: str) -> int:
        """Get remaining quota."""
        key = f"quota:{user_id}"
        quota = self.redis.get(key)
        return int(quota) if quota else 0

quota_manager = QuotaManager(redis_client)

@server.tool()
def quota_limited_operation(user_id: str, param: str) -> dict:
    """Operation with quota limiting."""
    remaining = quota_manager.get_quota(user_id)

    if remaining <= 0:
        raise ValueError(
            f"Quota exceeded for user {user_id}. "
            f"Quota resets in 30 days."
        )

    # Use 1 quota
    quota_manager.use_quota(user_id, amount=1)

    return execute_operation(user_id, param)
```

---

## Best Practices

✅ **Rate Limiting**:
- Use token bucket for most cases
- Implement per-user limits
- Use tiered limits for SaaS
- Include rate limit info in responses
- Monitor and adjust limits
- Log rate limit violations

✅ **Configuration**:
- Free tier: 100-1000 req/day
- Pro tier: 10K-100K req/day
- Enterprise: unlimited or custom
- Minute limits prevent bursts
- Daily limits prevent overuse

❌ **Avoid**:
- Too restrictive limits
- No per-user limits
- Missing error messages
- Not returning remaining quota
- Ignoring distributed systems
- Inflexible rate limit policies

---

**Last Updated**: 2025-11-27
