# Backend Architecture Examples

## Example 1: Microservices with Kubernetes & Istio

**Production microservices deployment**:

```yaml
# Kubernetes Deployment with Istio injection
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-service
  labels:
    app: api
    version: v1
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
        version: v1
      annotations:
        sidecar.istio.io/inject: "true"
    spec:
      containers:
      - name: api
        image: myregistry.azurecr.io/api:1.0.0
        ports:
        - containerPort: 8080
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 1000m
            memory: 1Gi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10

---
# Service exposing the deployment
apiVersion: v1
kind: Service
metadata:
  name: api-service
spec:
  ports:
  - port: 80
    targetPort: 8080
    name: http
  selector:
    app: api

---
# Istio VirtualService for traffic management
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: api-virtual-service
spec:
  hosts:
  - api-service
  http:
  - match:
    - uri:
        prefix: "/v2"
    route:
    - destination:
        host: api-service
        subset: v2
      weight: 10
    - destination:
        host: api-service
        subset: v1
      weight: 90
    timeout: 30s
    retries:
      attempts: 3
      perTryTimeout: 10s

---
# DestinationRule for load balancing and circuit breaking
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: api-destination-rule
spec:
  host: api-service
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

## Example 2: FastAPI with Async Database Operations

**Modern async Python backend**:

```python
from fastapi import FastAPI, HTTPException, Depends
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy import Column, Integer, String, DateTime
from sqlalchemy.ext.declarative import declarative_base
from pydantic import BaseModel
import asyncio
from datetime import datetime

app = FastAPI()

# Database setup
DATABASE_URL = "postgresql+asyncpg://user:password@localhost/db"
engine = create_async_engine(DATABASE_URL, echo=False)
async_session = sessionmaker(engine, class_=AsyncSession, expire_on_commit=False)

Base = declarative_base()

# Models
class UserModel(Base):
    __tablename__ = "users"
    id = Column(Integer, primary_key=True)
    name = Column(String)
    email = Column(String, unique=True)
    created_at = Column(DateTime, default=datetime.utcnow)

class UserSchema(BaseModel):
    name: str
    email: str

# Database dependency
async def get_session() -> AsyncSession:
    async with async_session() as session:
        yield session

# Endpoints
@app.post("/users")
async def create_user(user: UserSchema, session: AsyncSession = Depends(get_session)):
    """Create user with async database operation"""
    db_user = UserModel(name=user.name, email=user.email)
    session.add(db_user)
    await session.commit()
    await session.refresh(db_user)
    return db_user

@app.get("/users/{user_id}")
async def get_user(user_id: int, session: AsyncSession = Depends(get_session)):
    """Fetch user by ID"""
    from sqlalchemy import select
    result = await session.execute(select(UserModel).where(UserModel.id == user_id))
    user = result.scalar_one_or_none()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return user
```

---

## Example 3: Event-Driven Architecture

**Asynchronous event processing**:

```python
import asyncio
from dataclasses import dataclass
from typing import List, Callable
from datetime import datetime

@dataclass
class Event:
    type: str
    timestamp: datetime
    data: dict

class EventBus:
    """Central event bus for async event handling"""

    def __init__(self):
        self.subscribers: dict[str, List[Callable]] = {}

    def subscribe(self, event_type: str, handler: Callable):
        """Register event handler"""
        if event_type not in self.subscribers:
            self.subscribers[event_type] = []
        self.subscribers[event_type].append(handler)

    async def publish(self, event: Event):
        """Publish event to all subscribers"""
        handlers = self.subscribers.get(event.type, [])
        tasks = [handler(event) for handler in handlers]
        await asyncio.gather(*tasks)

# Usage
event_bus = EventBus()

async def on_user_created(event: Event):
    """Handler for user creation event"""
    user_data = event.data
    print(f"User created: {user_data['email']}")
    # Send welcome email, create subscription, etc.

async def on_user_created_notify(event: Event):
    """Another handler for same event"""
    print(f"Notifying managers about new user")

event_bus.subscribe("user.created", on_user_created)
event_bus.subscribe("user.created", on_user_created_notify)

# Publish event
async def create_user(email: str):
    # ... create user in database
    event = Event(
        type="user.created",
        timestamp=datetime.utcnow(),
        data={"email": email}
    )
    await event_bus.publish(event)
```

---

## Example 4: Database Connection Pooling

**Optimized database connections**:

```python
from sqlalchemy import create_engine
from sqlalchemy.pool import QueuePool

# Optimal connection pool configuration
engine = create_engine(
    'postgresql://user:password@localhost:5432/mydb',
    poolclass=QueuePool,
    pool_size=20,              # Base connections
    max_overflow=10,           # Additional connections
    pool_timeout=30,           # Wait timeout
    pool_recycle=3600,         # Recycle after 1 hour
    pool_pre_ping=True,        # Check connection health
    echo_pool=False            # Disable logging
)

# Redis connection pool
import redis

redis_pool = redis.ConnectionPool(
    host='localhost',
    port=6379,
    max_connections=50,
    socket_connect_timeout=5,
    socket_keepalive=True
)

redis_client = redis.Redis(connection_pool=redis_pool)
```

---

## Example 5: API Rate Limiting

**Tiered rate limiting by user**:

```python
from fastapi import FastAPI, HTTPException
from slowapi import Limiter
from slowapi.util import get_remote_address

app = FastAPI()
limiter = Limiter(key_func=get_remote_address)

# Rate limit tiers
RATE_LIMITS = {
    "free": "10/minute",
    "pro": "100/minute",
    "enterprise": "unlimited"
}

def get_user_tier(user_id: int) -> str:
    # Fetch from database or cache
    return "pro"

@app.get("/api/data")
@limiter.limit("10/minute")  # Default limit
async def get_data(user_id: int):
    tier = get_user_tier(user_id)
    if tier == "free":
        # Apply free tier limit
        pass
    return {"data": "response"}
```

---

## Example 6: Health Checks & Graceful Shutdown

**Production-ready health monitoring**:

```python
from fastapi import FastAPI
from contextlib import asynccontextmanager

app = FastAPI()

# Graceful shutdown
@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    print("Starting application")
    await initialize_database()
    yield
    # Shutdown
    print("Shutting down application")
    await cleanup_resources()

app = FastAPI(lifespan=lifespan)

@app.get("/health/live")
async def health_live():
    """Liveness probe - is app running?"""
    return {"status": "alive"}

@app.get("/health/ready")
async def health_ready():
    """Readiness probe - is app ready for traffic?"""
    db_ok = await check_database()
    cache_ok = await check_cache()

    if not (db_ok and cache_ok):
        return {"status": "not_ready", "db": db_ok, "cache": cache_ok}, 503

    return {"status": "ready"}
```

---

## Example 7: Structured Logging

**Production-grade logging**:

```python
import logging
import json
from pythonjsonlogger import jsonlogger

# Configure JSON logging
logHandler = logging.StreamHandler()
formatter = jsonlogger.JsonFormatter()
logHandler.setFormatter(formatter)

logger = logging.getLogger()
logger.addHandler(logHandler)
logger.setLevel(logging.INFO)

# Usage with context
import contextvars

request_id = contextvars.ContextVar('request_id', default='')

def log_with_context(level, message, **kwargs):
    """Log with request context"""
    logger.log(level, message, extra={
        "request_id": request_id.get(),
        **kwargs
    })

# In API handlers
@app.post("/users")
async def create_user(user_data: dict):
    request_id.set(uuid.uuid4())
    log_with_context(logging.INFO, "Creating user", email=user_data['email'])
    # ... create user
    log_with_context(logging.INFO, "User created successfully", user_id=user.id)
```

---

## Example 8: Circuit Breaker Pattern

**Prevent cascading failures**:

```python
from enum import Enum
from datetime import datetime, timedelta

class CircuitState(Enum):
    CLOSED = "closed"
    OPEN = "open"
    HALF_OPEN = "half_open"

class CircuitBreaker:
    """Protect against failing services"""

    def __init__(self, failure_threshold=5, timeout=60):
        self.failure_threshold = failure_threshold
        self.timeout = timeout
        self.failure_count = 0
        self.state = CircuitState.CLOSED
        self.last_failure_time = None

    async def call(self, func, *args, **kwargs):
        """Execute function with circuit breaker"""
        if self.state == CircuitState.OPEN:
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

# Usage
breaker = CircuitBreaker()

@app.get("/external-service")
async def call_external_service():
    return await breaker.call(external_api_call)
```

---

## Example 9: Message Queue Processing

**Async message processing with RabbitMQ**:

```python
import aio_pika
import json
from typing import Callable

class MessageQueueHandler:
    """Handle async message processing"""

    def __init__(self, amqp_url: str):
        self.amqp_url = amqp_url
        self.connection = None
        self.channel = None

    async def connect(self):
        self.connection = await aio_pika.connect_robust(self.amqp_url)
        self.channel = await self.connection.channel()

    async def declare_queue(self, queue_name: str):
        """Declare queue and exchange"""
        exchange = await self.channel.declare_exchange(
            'tasks',
            aio_pika.ExchangeType.TOPIC,
            durable=True
        )

        queue = await self.channel.declare_queue(
            queue_name,
            durable=True
        )

        await queue.bind(exchange, 'task.*')
        return queue

    async def consume_messages(self, queue_name: str, handler: Callable):
        """Consume and process messages"""
        queue = await self.declare_queue(queue_name)

        async with queue.iterator() as queue_iter:
            async for message in queue_iter:
                async with message.process():
                    try:
                        body = json.loads(message.body)
                        await handler(body)
                    except Exception as e:
                        print(f"Error processing message: {e}")

    async def publish_message(self, queue_name: str, data: dict):
        """Publish message to queue"""
        exchange = await self.channel.get_exchange('tasks')
        message = aio_pika.Message(body=json.dumps(data).encode())
        await exchange.publish(message, routing_key=f'task.{queue_name}')

# Usage
queue_handler = MessageQueueHandler('amqp://guest:guest@localhost/')

async def process_order(order_data: dict):
    print(f"Processing order: {order_data['order_id']}")
    # Process order logic

async def consume_orders():
    await queue_handler.connect()
    await queue_handler.consume_messages('orders', process_order)
```

---

## Example 10: Dependency Injection Pattern

**Clean architecture with dependency injection**:

```python
from abc import ABC, abstractmethod
from typing import List
from fastapi import FastAPI, Depends

# Abstract interfaces
class UserRepository(ABC):
    @abstractmethod
    async def get_user(self, user_id: int):
        pass

    @abstractmethod
    async def create_user(self, user_data: dict):
        pass

class EmailService(ABC):
    @abstractmethod
    async def send_welcome_email(self, email: str):
        pass

# Concrete implementations
class PostgresUserRepository(UserRepository):
    def __init__(self, session):
        self.session = session

    async def get_user(self, user_id: int):
        # Fetch from PostgreSQL
        pass

    async def create_user(self, user_data: dict):
        # Insert into PostgreSQL
        pass

class SMTPEmailService(EmailService):
    def __init__(self, smtp_config):
        self.smtp_config = smtp_config

    async def send_welcome_email(self, email: str):
        # Send via SMTP
        pass

# Dependency container
class Container:
    def __init__(self):
        self.user_repo = None
        self.email_service = None

    def setup(self):
        self.user_repo = PostgresUserRepository(session)
        self.email_service = SMTPEmailService(smtp_config)

# FastAPI with dependency injection
app = FastAPI()
container = Container()
container.setup()

def get_user_repository() -> UserRepository:
    return container.user_repo

def get_email_service() -> EmailService:
    return container.email_service

@app.post("/users")
async def create_user(
    user_data: dict,
    user_repo: UserRepository = Depends(get_user_repository),
    email_service: EmailService = Depends(get_email_service)
):
    new_user = await user_repo.create_user(user_data)
    await email_service.send_welcome_email(user_data['email'])
    return new_user
```

---

**Last Updated**: 2025-11-22
**Version**: 4.0.0
