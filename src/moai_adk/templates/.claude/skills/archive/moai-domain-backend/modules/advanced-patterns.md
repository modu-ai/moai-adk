# Advanced Backend Patterns

## Distributed System Architecture

### CQRS Pattern (Command Query Responsibility Segregation)

**Pattern Overview**:
Separate read and write operations into distinct services with different data models.

**Implementation (FastAPI + PostgreSQL)**:
```python
from fastapi import FastAPI
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

app = FastAPI()

# Write side (Command)
write_engine = create_engine("postgresql://user:pass@write-db:5432/db")
write_session = sessionmaker(bind=write_engine)

# Read side (Query)
read_engine = create_engine("postgresql://user:pass@read-db:5432/db")
read_session = sessionmaker(bind=read_engine)

@app.post("/orders/")
async def create_order(order_data: OrderCreate):
    """Command: Write to write database"""
    session = write_session()
    db_order = Order(**order_data.dict())
    session.add(db_order)
    session.commit()

    # Publish event for eventual consistency
    await publish_event(OrderCreated(order_id=db_order.id))
    return {"order_id": db_order.id}

@app.get("/orders/{order_id}")
async def get_order(order_id: int):
    """Query: Read from read-optimized database"""
    session = read_session()
    order = session.query(OrderReadModel).filter_by(id=order_id).first()
    return order
```

**Benefits**:
- Independent scaling of read and write workloads
- Optimized data models for each operation
- Eventual consistency allows for eventual synchronization
- Better performance for read-heavy workloads

---

### Event Sourcing Pattern

**Core Concept**: Store all changes to application state as immutable events in event log.

**Implementation (Python)**:
```python
from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import List

class OrderEvent(Enum):
    CREATED = "order_created"
    PAYMENT_RECEIVED = "payment_received"
    SHIPPED = "shipped"
    DELIVERED = "delivered"

@dataclass
class EventLog:
    event_type: OrderEvent
    aggregate_id: str
    timestamp: datetime
    payload: dict
    version: int

class OrderAggregate:
    def __init__(self, order_id: str):
        self.order_id = order_id
        self.status = "pending"
        self.events: List[EventLog] = []

    def create_order(self, user_id: str, items: List[dict]):
        """Create new order"""
        event = EventLog(
            event_type=OrderEvent.CREATED,
            aggregate_id=self.order_id,
            timestamp=datetime.utcnow(),
            payload={"user_id": user_id, "items": items},
            version=1
        )
        self.events.append(event)
        self.apply_event(event)

    def apply_event(self, event: EventLog):
        """Apply event to current state"""
        if event.event_type == OrderEvent.CREATED:
            self.status = "created"
        elif event.event_type == OrderEvent.PAYMENT_RECEIVED:
            self.status = "paid"
        elif event.event_type == OrderEvent.SHIPPED:
            self.status = "shipped"
        elif event.event_type == OrderEvent.DELIVERED:
            self.status = "delivered"

    def replay_events(self, events: List[EventLog]):
        """Rebuild state from event history"""
        for event in sorted(events, key=lambda e: e.version):
            self.apply_event(event)
```

**Advantages**:
- Complete audit trail of all state changes
- Time travel debugging capabilities
- Natural fit for eventual consistency
- Event replay for data recovery

---

### Saga Pattern for Distributed Transactions

**Pattern**: Coordinated set of local transactions across multiple services.

**Implementation (Choreography-based)**:
```python
from enum import Enum
import asyncio
from aio_pika import connect, Exchange, Queue

class PaymentService:
    async def process_payment(self, order_id: str, amount: float):
        """Process payment and emit event"""
        try:
            # Process payment
            transaction_id = await self.charge_card(order_id, amount)

            # Emit PaymentCompleted event
            await self.event_bus.publish_event(
                "payment.completed",
                {"order_id": order_id, "transaction_id": transaction_id}
            )
        except Exception as e:
            # Emit PaymentFailed event for compensation
            await self.event_bus.publish_event(
                "payment.failed",
                {"order_id": order_id, "reason": str(e)}
            )

class InventoryService:
    async def handle_payment_completed(self, event: dict):
        """Reserve inventory after payment"""
        order_id = event["order_id"]
        try:
            await self.reserve_inventory(order_id)
            await self.event_bus.publish_event(
                "inventory.reserved",
                {"order_id": order_id}
            )
        except OutOfStockError:
            # Compensating transaction
            await self.event_bus.publish_event(
                "inventory.failed",
                {"order_id": order_id}
            )

class OrderService:
    async def handle_inventory_reserved(self, event: dict):
        """Complete order after inventory reserved"""
        order_id = event["order_id"]
        await self.mark_order_confirmed(order_id)

    async def handle_inventory_failed(self, event: dict):
        """Compensate: Refund payment"""
        order_id = event["order_id"]
        await self.cancel_order(order_id)
        await self.event_bus.publish_event(
            "order.refunded",
            {"order_id": order_id}
        )
```

---

### Service Mesh with Istio

**Key Capabilities**:
- Traffic management with fine-grained routing rules
- Security policies (mutual TLS, authorization policies)
- Observability (distributed tracing, metrics)
- Resilience patterns (circuit breaker, retry, timeout)

**VirtualService Example**:
```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: product-service
spec:
  hosts:
  - product-service
  http:
  - match:
    - uri:
        prefix: "/v2"
    route:
    - destination:
        host: product-service
        subset: v2
      weight: 10
    - destination:
        host: product-service
        subset: v1
      weight: 90
  - route:
    - destination:
        host: product-service
        subset: v1
---
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: product-service
spec:
  host: product-service
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        http1MaxPendingRequests: 50
        maxRequestsPerConnection: 2
    loadBalancer:
      simple: LEAST_REQUEST
    outlierDetection:
      consecutive5xxErrors: 5
      interval: 30s
      baseEjectionTime: 30s
```

---

### Advanced Caching Strategies

**Multi-Level Caching**:
```python
from functools import wraps
import redis
import json
from typing import Any, Callable

class CacheManager:
    def __init__(self):
        self.redis_client = redis.Redis(host='localhost', port=6379)
        self.local_cache = {}

    async def get(self, key: str, ttl: int = 3600) -> Any:
        """Get from cache with fallthrough strategy"""
        # L1: Local memory cache
        if key in self.local_cache:
            return self.local_cache[key]

        # L2: Redis cache
        cached = self.redis_client.get(key)
        if cached:
            value = json.loads(cached)
            self.local_cache[key] = value  # Populate L1
            return value

        return None

    async def set(self, key: str, value: Any, ttl: int = 3600):
        """Set in both cache layers"""
        self.local_cache[key] = value
        self.redis_client.setex(
            key,
            ttl,
            json.dumps(value)
        )

    def cache_aside(self, ttl: int = 3600):
        """Decorator for cache-aside pattern"""
        def decorator(func: Callable):
            async def wrapper(*args, **kwargs):
                cache_key = f"{func.__name__}:{str(args)}:{str(kwargs)}"

                # Try cache first
                cached_value = await self.get(cache_key)
                if cached_value is not None:
                    return cached_value

                # Cache miss: call function and cache result
                result = await func(*args, **kwargs)
                await self.set(cache_key, result, ttl)
                return result

            return wrapper
        return decorator
```

---

### Rate Limiting & Throttling

**Token Bucket Algorithm**:
```python
from time import time
from threading import Lock

class TokenBucket:
    def __init__(self, capacity: int, refill_rate: float):
        self.capacity = capacity
        self.tokens = capacity
        self.refill_rate = refill_rate  # tokens per second
        self.last_refill = time()
        self.lock = Lock()

    def allow_request(self) -> bool:
        """Check if request is allowed"""
        with self.lock:
            current_time = time()
            elapsed = current_time - self.last_refill

            # Refill tokens based on elapsed time
            self.tokens = min(
                self.capacity,
                self.tokens + elapsed * self.refill_rate
            )
            self.last_refill = current_time

            if self.tokens >= 1:
                self.tokens -= 1
                return True
            return False

# FastAPI middleware integration
from fastapi import Request, HTTPException

rate_limiters = {}

async def rate_limit_middleware(request: Request, call_next):
    """Apply rate limiting by user"""
    user_id = request.query_params.get("user_id", "anonymous")

    if user_id not in rate_limiters:
        rate_limiters[user_id] = TokenBucket(capacity=100, refill_rate=10)

    if not rate_limiters[user_id].allow_request():
        raise HTTPException(status_code=429, detail="Rate limit exceeded")

    return await call_next(request)
```

---

## Database Optimization Techniques

### Query Optimization with Explain Analysis

```python
async def analyze_slow_query(db_connection, query: str):
    """Analyze query execution plan"""
    result = await db_connection.execute(f"EXPLAIN ANALYZE {query}")

    # Parse execution plan
    execution_plan = result.all()
    for row in execution_plan:
        print(row[0])  # Print execution plan

    # Look for sequential scans (potential issues)
    plan_text = "\n".join([row[0] for row in execution_plan])
    if "Seq Scan" in plan_text:
        print("WARNING: Sequential scan detected - consider adding index")
```

### Composite Indexes for Complex Queries

```sql
-- Complex query on multiple columns
CREATE INDEX idx_orders_user_status_date ON orders (user_id, status, created_at DESC);

-- Partial index for active records only
CREATE INDEX idx_active_orders ON orders (user_id, created_at DESC)
WHERE status != 'cancelled';

-- BRIN index for time-series data (much smaller than B-tree)
CREATE INDEX idx_logs_date ON logs USING BRIN (created_at);

-- GiST for full-text search
CREATE INDEX idx_content_search ON documents USING GiST (to_tsvector('english', content));
```

---

## Error Handling & Resilience

### Circuit Breaker Pattern

```python
from enum import Enum
from datetime import datetime, timedelta

class CircuitState(Enum):
    CLOSED = "closed"      # Normal operation
    OPEN = "open"          # Failing, reject requests
    HALF_OPEN = "half_open"  # Testing recovery

class CircuitBreaker:
    def __init__(self, failure_threshold: int = 5, timeout: int = 60):
        self.failure_threshold = failure_threshold
        self.timeout = timeout  # seconds
        self.failure_count = 0
        self.state = CircuitState.CLOSED
        self.last_failure_time = None

    async def call(self, func, *args, **kwargs):
        """Execute function with circuit breaker protection"""
        if self.state == CircuitState.OPEN:
            # Check if timeout has passed
            if datetime.now() - self.last_failure_time > timedelta(seconds=self.timeout):
                self.state = CircuitState.HALF_OPEN
            else:
                raise Exception("Circuit breaker is OPEN")

        try:
            result = await func(*args, **kwargs)
            self._on_success()
            return result
        except Exception as e:
            self._on_failure()
            raise

    def _on_success(self):
        self.failure_count = 0
        self.state = CircuitState.CLOSED

    def _on_failure(self):
        self.failure_count += 1
        self.last_failure_time = datetime.now()
        if self.failure_count >= self.failure_threshold:
            self.state = CircuitState.OPEN
```

---

**Last Updated**: 2025-11-22
**Version**: 4.0.0
