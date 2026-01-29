# Backend Development Examples

Multishot prompting examples demonstrating practical backend development patterns covering API design, database integration, microservices, and security.

---

## Example 1: RESTful API with Authentication

**Scenario**: Create a secure user management API with JWT authentication.

**Input**:
```python
from fastapi import FastAPI, Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel, EmailStr
from typing import List
import jwt
from datetime import datetime, timedelta

app = FastAPI(title="User Management API", version="1.0.0")
security = HTTPBearer()

# Models
class UserCreate(BaseModel):
    email: EmailStr
    password: str
    name: str

class UserResponse(BaseModel):
    id: int
    email: str
    name: str
    created_at: datetime

class TokenResponse(BaseModel):
    access_token: str
    token_type: str
    expires_in: int
```

**Output**:
```python
# Complete implementation with authentication
SECRET_KEY = "your-secret-key"  # Use environment variable in production
ALGORITHM = "HS256"

def create_access_token(user_id: int) -> str:
    expire = datetime.utcnow() + timedelta(hours=24)
    payload = {"sub": str(user_id), "exp": expire}
    return jwt.encode(payload, SECRET_KEY, algorithm=ALGORITHM)

async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security)
) -> int:
    try:
        payload = jwt.decode(
            credentials.credentials,
            SECRET_KEY,
            algorithms=[ALGORITHM]
        )
        user_id = int(payload.get("sub"))
        return user_id
    except jwt.ExpiredSignatureError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Token has expired"
        )
    except jwt.InvalidTokenError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token"
        )

@app.post("/auth/login", response_model=TokenResponse)
async def login(email: str, password: str):
    user = await user_service.authenticate(email, password)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid credentials"
        )
    token = create_access_token(user.id)
    return TokenResponse(
        access_token=token,
        token_type="bearer",
        expires_in=86400
    )

@app.get("/users/me", response_model=UserResponse)
async def get_current_user_profile(
    user_id: int = Depends(get_current_user)
):
    user = await user_service.get_by_id(user_id)
    return user
```

**Explanation**: This pattern demonstrates secure JWT authentication with proper error handling, token expiration, and dependency injection for protected routes.

---

## Example 2: Database Repository Pattern with Caching

**Scenario**: Implement a repository pattern with Redis caching for optimal performance.

**Input**:
```python
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
import redis.asyncio as redis
import json
from typing import Optional, List

class UserRepository:
    def __init__(
        self,
        session: AsyncSession,
        cache: redis.Redis
    ):
        self.session = session
        self.cache = cache
        self.cache_ttl = 3600  # 1 hour
```

**Output**:
```python
class UserRepository:
    def __init__(
        self,
        session: AsyncSession,
        cache: redis.Redis
    ):
        self.session = session
        self.cache = cache
        self.cache_ttl = 3600

    async def get_by_id(self, user_id: int) -> Optional[User]:
        # Try cache first
        cache_key = f"user:{user_id}"
        cached = await self.cache.get(cache_key)

        if cached:
            return User(**json.loads(cached))

        # Query database
        result = await self.session.execute(
            select(User).where(User.id == user_id)
        )
        user = result.scalar_one_or_none()

        if user:
            # Cache the result
            await self.cache.setex(
                cache_key,
                self.cache_ttl,
                json.dumps(user.to_dict())
            )

        return user

    async def get_all(
        self,
        skip: int = 0,
        limit: int = 100
    ) -> List[User]:
        result = await self.session.execute(
            select(User)
            .offset(skip)
            .limit(limit)
            .order_by(User.created_at.desc())
        )
        return result.scalars().all()

    async def create(self, user_data: UserCreate) -> User:
        user = User(**user_data.model_dump())
        self.session.add(user)
        await self.session.commit()
        await self.session.refresh(user)

        # Invalidate list cache
        await self.cache.delete("users:list:*")

        return user

    async def update(
        self,
        user_id: int,
        user_data: UserUpdate
    ) -> Optional[User]:
        user = await self.get_by_id(user_id)
        if not user:
            return None

        for key, value in user_data.model_dump(exclude_unset=True).items():
            setattr(user, key, value)

        await self.session.commit()

        # Invalidate cache
        await self.cache.delete(f"user:{user_id}")

        return user

    async def delete(self, user_id: int) -> bool:
        user = await self.get_by_id(user_id)
        if not user:
            return False

        await self.session.delete(user)
        await self.session.commit()

        # Invalidate cache
        await self.cache.delete(f"user:{user_id}")

        return True
```

**Explanation**: The repository pattern encapsulates data access logic with Redis caching for read operations and automatic cache invalidation on writes.

---

## Example 3: Event-Driven Microservice Architecture

**Scenario**: Implement an order processing service with event-driven communication.

**Input**:
```python
from aio_pika import connect_robust, Message, ExchangeType
import json
from dataclasses import dataclass
from datetime import datetime
from typing import Optional

@dataclass
class OrderCreatedEvent:
    order_id: str
    user_id: int
    total_amount: float
    items: list
    created_at: datetime
```

**Output**:
```python
import asyncio
from aio_pika import connect_robust, Message, ExchangeType
import json
from dataclasses import dataclass, asdict
from datetime import datetime
from typing import Optional, Callable, Dict
import logging

logger = logging.getLogger(__name__)

@dataclass
class OrderCreatedEvent:
    order_id: str
    user_id: int
    total_amount: float
    items: list
    created_at: datetime

class EventBus:
    def __init__(self, amqp_url: str):
        self.amqp_url = amqp_url
        self.connection = None
        self.channel = None
        self.exchange = None
        self.handlers: Dict[str, Callable] = {}

    async def connect(self):
        self.connection = await connect_robust(self.amqp_url)
        self.channel = await self.connection.channel()
        self.exchange = await self.channel.declare_exchange(
            "events",
            ExchangeType.TOPIC,
            durable=True
        )
        logger.info("Connected to message broker")

    async def publish(self, event_type: str, event_data: dict):
        message = Message(
            json.dumps({
                "type": event_type,
                "data": event_data,
                "timestamp": datetime.utcnow().isoformat()
            }).encode(),
            content_type="application/json",
            delivery_mode=2  # Persistent
        )
        await self.exchange.publish(message, routing_key=event_type)
        logger.info(f"Published event: {event_type}")

    async def subscribe(
        self,
        event_type: str,
        handler: Callable,
        queue_name: str
    ):
        queue = await self.channel.declare_queue(
            queue_name,
            durable=True
        )
        await queue.bind(self.exchange, routing_key=event_type)

        async def process_message(message):
            async with message.process():
                try:
                    data = json.loads(message.body.decode())
                    await handler(data)
                except Exception as e:
                    logger.error(f"Error processing message: {e}")
                    raise

        await queue.consume(process_message)
        logger.info(f"Subscribed to: {event_type}")

class OrderService:
    def __init__(self, event_bus: EventBus, db_session):
        self.event_bus = event_bus
        self.db = db_session

    async def create_order(self, order_data: dict) -> Order:
        # Create order in database
        order = Order(**order_data)
        self.db.add(order)
        await self.db.commit()

        # Publish event
        event = OrderCreatedEvent(
            order_id=str(order.id),
            user_id=order.user_id,
            total_amount=order.total_amount,
            items=order.items,
            created_at=order.created_at
        )
        await self.event_bus.publish(
            "order.created",
            asdict(event)
        )

        return order

class NotificationService:
    def __init__(self, event_bus: EventBus):
        self.event_bus = event_bus

    async def start(self):
        await self.event_bus.subscribe(
            "order.created",
            self.handle_order_created,
            "notification-service-orders"
        )

    async def handle_order_created(self, event_data: dict):
        order_data = event_data["data"]
        user_id = order_data["user_id"]
        order_id = order_data["order_id"]

        # Send notification
        await self.send_email(
            user_id=user_id,
            subject=f"Order {order_id} Confirmed",
            body=f"Your order for ${order_data['total_amount']} has been confirmed."
        )
        logger.info(f"Sent order confirmation for {order_id}")
```

**Explanation**: This pattern demonstrates event-driven architecture with RabbitMQ, enabling loose coupling between services and reliable message delivery.

---

## Common Patterns

### Pattern 1: Circuit Breaker for External Services

Protect your service from cascading failures:

```python
from tenacity import (
    retry,
    stop_after_attempt,
    wait_exponential,
    retry_if_exception_type
)
import httpx
from dataclasses import dataclass
from datetime import datetime, timedelta

@dataclass
class CircuitBreakerState:
    failures: int = 0
    last_failure: datetime = None
    is_open: bool = False

class CircuitBreaker:
    def __init__(
        self,
        failure_threshold: int = 5,
        recovery_timeout: int = 30
    ):
        self.failure_threshold = failure_threshold
        self.recovery_timeout = timedelta(seconds=recovery_timeout)
        self.state = CircuitBreakerState()

    def can_execute(self) -> bool:
        if not self.state.is_open:
            return True

        if datetime.utcnow() - self.state.last_failure > self.recovery_timeout:
            self.state.is_open = False
            self.state.failures = 0
            return True

        return False

    def record_failure(self):
        self.state.failures += 1
        self.state.last_failure = datetime.utcnow()

        if self.state.failures >= self.failure_threshold:
            self.state.is_open = True

    def record_success(self):
        self.state.failures = 0
        self.state.is_open = False

class ExternalPaymentService:
    def __init__(self, base_url: str):
        self.base_url = base_url
        self.circuit_breaker = CircuitBreaker()
        self.client = httpx.AsyncClient()

    @retry(
        stop=stop_after_attempt(3),
        wait=wait_exponential(multiplier=1, min=1, max=10),
        retry=retry_if_exception_type(httpx.TransportError)
    )
    async def process_payment(self, payment_data: dict) -> dict:
        if not self.circuit_breaker.can_execute():
            raise ServiceUnavailableError("Payment service circuit open")

        try:
            response = await self.client.post(
                f"{self.base_url}/payments",
                json=payment_data,
                timeout=10.0
            )
            response.raise_for_status()
            self.circuit_breaker.record_success()
            return response.json()
        except Exception as e:
            self.circuit_breaker.record_failure()
            raise
```

### Pattern 2: Request Validation Middleware

Comprehensive request validation:

```python
from fastapi import Request, HTTPException
from starlette.middleware.base import BaseHTTPMiddleware
from pydantic import ValidationError
import time
import logging

logger = logging.getLogger(__name__)

class RequestValidationMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        start_time = time.time()
        request_id = request.headers.get("X-Request-ID", str(uuid.uuid4()))

        # Add request ID to context
        request.state.request_id = request_id

        # Log incoming request
        logger.info(f"[{request_id}] {request.method} {request.url.path}")

        try:
            response = await call_next(request)

            # Add response headers
            response.headers["X-Request-ID"] = request_id
            response.headers["X-Response-Time"] = str(time.time() - start_time)

            logger.info(
                f"[{request_id}] Completed {response.status_code} "
                f"in {time.time() - start_time:.3f}s"
            )

            return response

        except ValidationError as e:
            logger.warning(f"[{request_id}] Validation error: {e}")
            raise HTTPException(status_code=422, detail=e.errors())
        except Exception as e:
            logger.error(f"[{request_id}] Unexpected error: {e}")
            raise
```

### Pattern 3: Database Connection Pool Management

Optimized database connections:

```python
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import QueuePool
from contextlib import asynccontextmanager

class DatabaseManager:
    def __init__(self, database_url: str):
        self.engine = create_async_engine(
            database_url,
            poolclass=QueuePool,
            pool_size=20,
            max_overflow=30,
            pool_pre_ping=True,
            pool_recycle=3600,
            echo=False
        )
        self.async_session = sessionmaker(
            self.engine,
            class_=AsyncSession,
            expire_on_commit=False
        )

    @asynccontextmanager
    async def get_session(self):
        session = self.async_session()
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()

    async def health_check(self) -> bool:
        try:
            async with self.get_session() as session:
                await session.execute("SELECT 1")
            return True
        except Exception:
            return False
```

---

## Anti-Patterns (Patterns to Avoid)

### Anti-Pattern 1: N+1 Query Problem

**Problem**: Making individual database queries for related entities.

```python
# Incorrect approach
async def get_orders_with_items():
    orders = await session.execute(select(Order))
    for order in orders.scalars():
        # N+1 problem: one query per order
        items = await session.execute(
            select(OrderItem).where(OrderItem.order_id == order.id)
        )
        order.items = items.scalars().all()
    return orders
```

**Solution**: Use eager loading with joins.

```python
# Correct approach
async def get_orders_with_items():
    result = await session.execute(
        select(Order)
        .options(selectinload(Order.items))
        .order_by(Order.created_at.desc())
    )
    return result.scalars().all()
```

### Anti-Pattern 2: Synchronous Operations in Async Context

**Problem**: Blocking the event loop with synchronous operations.

```python
# Incorrect approach
@app.get("/data")
async def get_data():
    # This blocks the event loop!
    data = requests.get("https://api.example.com/data")
    return data.json()
```

**Solution**: Use async-compatible libraries.

```python
# Correct approach
@app.get("/data")
async def get_data():
    async with httpx.AsyncClient() as client:
        response = await client.get("https://api.example.com/data")
        return response.json()
```

### Anti-Pattern 3: Hardcoded Configuration

**Problem**: Hardcoding configuration values in code.

```python
# Incorrect approach
DATABASE_URL = "postgresql://user:password@localhost:5432/db"
SECRET_KEY = "my-super-secret-key"
```

**Solution**: Use environment variables with validation.

```python
# Correct approach
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    database_url: str
    secret_key: str
    redis_url: str = "redis://localhost:6379"
    debug: bool = False

    class Config:
        env_file = ".env"
        case_sensitive = False

settings = Settings()
```

---

## Integration Examples

### Health Check Endpoint

```python
from fastapi import APIRouter
from datetime import datetime

router = APIRouter(prefix="/health", tags=["Health"])

@router.get("")
async def health_check(
    db: DatabaseManager = Depends(get_db),
    cache: redis.Redis = Depends(get_cache)
):
    checks = {
        "status": "healthy",
        "timestamp": datetime.utcnow().isoformat(),
        "checks": {}
    }

    # Database check
    checks["checks"]["database"] = await db.health_check()

    # Cache check
    try:
        await cache.ping()
        checks["checks"]["cache"] = True
    except Exception:
        checks["checks"]["cache"] = False

    # Overall status
    if not all(checks["checks"].values()):
        checks["status"] = "degraded"

    return checks
```

### Structured Logging

```python
import structlog
from fastapi import FastAPI

def configure_logging():
    structlog.configure(
        processors=[
            structlog.stdlib.filter_by_level,
            structlog.stdlib.add_logger_name,
            structlog.stdlib.add_log_level,
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.processors.JSONRenderer()
        ],
        wrapper_class=structlog.stdlib.BoundLogger,
        context_class=dict,
        logger_factory=structlog.stdlib.LoggerFactory(),
        cache_logger_on_first_use=True
    )

logger = structlog.get_logger()

@app.middleware("http")
async def log_requests(request: Request, call_next):
    logger.info(
        "request_started",
        method=request.method,
        path=request.url.path,
        client_ip=request.client.host
    )
    response = await call_next(request)
    logger.info(
        "request_completed",
        status_code=response.status_code
    )
    return response
```

---

*For additional patterns and advanced configurations, see the related skills and documentation.*
