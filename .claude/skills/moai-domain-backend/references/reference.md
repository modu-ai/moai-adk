# Backend Development Reference

## API Reference

### FastAPI Application Setup

Complete application structure:
```python
from fastapi import FastAPI, Depends, HTTPException, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.security import OAuth2PasswordBearer
from contextlib import asynccontextmanager

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    await database.connect()
    await cache.connect()
    yield
    # Shutdown
    await database.disconnect()
    await cache.disconnect()

app = FastAPI(
    title="API Service",
    version="2.0.0",
    lifespan=lifespan,
    docs_url="/docs",
    redoc_url="/redoc",
    openapi_url="/openapi.json"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
```

### Database Connection Patterns

SQLAlchemy Async Engine:
```python
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker

DATABASE_URL = "postgresql+asyncpg://user:pass@localhost/db"

engine = create_async_engine(
    DATABASE_URL,
    pool_size=20,
    max_overflow=30,
    pool_pre_ping=True,
    pool_recycle=3600,
    echo=False
)

async_session = sessionmaker(
    engine, class_=AsyncSession, expire_on_commit=False
)

async def get_db():
    async with async_session() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
```

Motor (MongoDB) Connection:
```python
from motor.motor_asyncio import AsyncIOMotorClient

MONGODB_URL = "mongodb://localhost:27017"

client = AsyncIOMotorClient(
    MONGODB_URL,
    maxPoolSize=50,
    minPoolSize=10,
    maxIdleTimeMS=50000,
    waitQueueTimeoutMS=5000
)

db = client.myapp

async def get_mongodb():
    return db
```

Redis Connection:
```python
import redis.asyncio as redis

REDIS_URL = "redis://localhost:6379"

redis_pool = redis.ConnectionPool.from_url(
    REDIS_URL,
    max_connections=50,
    decode_responses=True
)

async def get_redis():
    return redis.Redis(connection_pool=redis_pool)
```

### Authentication Middleware

JWT Authentication:
```python
from jose import jwt, JWTError
from datetime import datetime, timedelta
from passlib.context import CryptContext

pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="auth/login")

SECRET_KEY = "your-secret-key"
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 30

def create_access_token(data: dict, expires_delta: timedelta = None):
    to_encode = data.copy()
    expire = datetime.utcnow() + (expires_delta or timedelta(minutes=15))
    to_encode.update({"exp": expire})
    return jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)

async def get_current_user(token: str = Depends(oauth2_scheme)):
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Could not validate credentials",
        headers={"WWW-Authenticate": "Bearer"},
    )
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        user_id: str = payload.get("sub")
        if user_id is None:
            raise credentials_exception
    except JWTError:
        raise credentials_exception
    return await get_user_by_id(user_id)
```

---

## Configuration Options

### Application Configuration

```yaml
# config/settings.yaml
app:
  name: "Backend Service"
  version: "2.0.0"
  debug: false
  environment: "production"

server:
  host: "0.0.0.0"
  port: 8000
  workers: 4
  reload: false

database:
  postgresql:
    url: "postgresql+asyncpg://user:pass@localhost/db"
    pool_size: 20
    max_overflow: 30
    pool_pre_ping: true
  mongodb:
    url: "mongodb://localhost:27017"
    database: "myapp"
    max_pool_size: 50
  redis:
    url: "redis://localhost:6379"
    max_connections: 50

security:
  secret_key: "${SECRET_KEY}"
  algorithm: "HS256"
  access_token_expire_minutes: 30
  refresh_token_expire_days: 7

cors:
  allow_origins:
    - "http://localhost:3000"
    - "https://myapp.com"
  allow_credentials: true
  allow_methods: ["*"]
  allow_headers: ["*"]

logging:
  level: "INFO"
  format: "json"
  handlers:
    - console
    - file
```

### Environment Variables

APP_ENV - Application environment (development, staging, production)
DATABASE_URL - Primary database connection string
MONGODB_URL - MongoDB connection string
REDIS_URL - Redis connection string
SECRET_KEY - JWT signing key
CORS_ORIGINS - Comma-separated allowed origins
LOG_LEVEL - Logging level (DEBUG, INFO, WARNING, ERROR)

---

## Integration Patterns

### Service Layer Pattern

```python
from abc import ABC, abstractmethod
from typing import Generic, TypeVar, List, Optional

T = TypeVar('T')

class BaseService(ABC, Generic[T]):
    @abstractmethod
    async def get(self, id: int) -> Optional[T]:
        pass

    @abstractmethod
    async def get_all(self, skip: int = 0, limit: int = 100) -> List[T]:
        pass

    @abstractmethod
    async def create(self, obj: T) -> T:
        pass

    @abstractmethod
    async def update(self, id: int, obj: T) -> Optional[T]:
        pass

    @abstractmethod
    async def delete(self, id: int) -> bool:
        pass

class UserService(BaseService[User]):
    def __init__(self, db: AsyncSession, cache: Redis):
        self.db = db
        self.cache = cache

    async def get(self, id: int) -> Optional[User]:
        # Check cache first
        cached = await self.cache.get(f"user:{id}")
        if cached:
            return User.parse_raw(cached)

        # Query database
        user = await self.db.get(User, id)
        if user:
            await self.cache.setex(f"user:{id}", 3600, user.json())
        return user
```

### Repository Pattern

```python
from sqlalchemy import select, update, delete
from sqlalchemy.ext.asyncio import AsyncSession

class UserRepository:
    def __init__(self, session: AsyncSession):
        self.session = session

    async def find_by_id(self, user_id: int) -> Optional[User]:
        result = await self.session.execute(
            select(User).where(User.id == user_id)
        )
        return result.scalar_one_or_none()

    async def find_by_email(self, email: str) -> Optional[User]:
        result = await self.session.execute(
            select(User).where(User.email == email)
        )
        return result.scalar_one_or_none()

    async def create(self, user_data: UserCreate) -> User:
        user = User(**user_data.dict())
        self.session.add(user)
        await self.session.flush()
        return user

    async def update(self, user_id: int, data: UserUpdate) -> Optional[User]:
        await self.session.execute(
            update(User)
            .where(User.id == user_id)
            .values(**data.dict(exclude_unset=True))
        )
        return await self.find_by_id(user_id)
```

### Event-Driven Architecture

```python
import asyncio
from aio_pika import connect_robust, Message, ExchangeType

class EventPublisher:
    def __init__(self, amqp_url: str):
        self.amqp_url = amqp_url
        self.connection = None
        self.channel = None
        self.exchange = None

    async def connect(self):
        self.connection = await connect_robust(self.amqp_url)
        self.channel = await self.connection.channel()
        self.exchange = await self.channel.declare_exchange(
            "events", ExchangeType.TOPIC, durable=True
        )

    async def publish(self, event_type: str, data: dict):
        message = Message(
            json.dumps(data).encode(),
            content_type="application/json",
            headers={"event_type": event_type}
        )
        await self.exchange.publish(message, routing_key=event_type)

class EventSubscriber:
    def __init__(self, amqp_url: str):
        self.amqp_url = amqp_url
        self.handlers = {}

    def on(self, event_type: str):
        def decorator(func):
            self.handlers[event_type] = func
            return func
        return decorator

    async def start(self):
        connection = await connect_robust(self.amqp_url)
        channel = await connection.channel()
        exchange = await channel.declare_exchange(
            "events", ExchangeType.TOPIC, durable=True
        )
        queue = await channel.declare_queue("", exclusive=True)

        for event_type in self.handlers:
            await queue.bind(exchange, routing_key=event_type)

        async with queue.iterator() as queue_iter:
            async for message in queue_iter:
                async with message.process():
                    event_type = message.headers.get("event_type")
                    if event_type in self.handlers:
                        data = json.loads(message.body)
                        await self.handlers[event_type](data)
```

---

## Troubleshooting

### Common Issues

Issue: Connection pool exhausted
Symptoms: Requests timeout, "too many connections" errors
Solution:
- Increase pool_size and max_overflow in database configuration
- Check for connection leaks (ensure proper context manager usage)
- Implement connection health checks with pool_pre_ping

Issue: Slow database queries
Symptoms: High response times, database CPU spikes
Solution:
- Use EXPLAIN ANALYZE to identify slow queries
- Add appropriate indexes based on query patterns
- Implement query result caching with Redis
- Consider read replicas for read-heavy workloads

Issue: Memory leaks in async operations
Symptoms: Gradual memory increase, eventual OOM
Solution:
- Use async context managers properly
- Implement proper cleanup in lifespan handlers
- Monitor task cancellation and cleanup
- Use weak references for caches where appropriate

Issue: CORS errors in browser
Symptoms: Cross-origin requests blocked
Solution:
- Verify allow_origins includes client domain
- Check allow_credentials setting for cookie-based auth
- Ensure preflight OPTIONS requests are handled
- Add explicit headers for custom request headers

Issue: JWT token expiration issues
Symptoms: Users logged out unexpectedly
Solution:
- Implement refresh token rotation
- Use sliding window expiration for active users
- Add token refresh middleware
- Handle token refresh in frontend interceptors

### Performance Optimization

Query Optimization:
- Use select_related/joinedload for N+1 query prevention
- Implement pagination with cursor-based approach for large datasets
- Use database-level aggregations instead of application-level
- Cache frequently accessed, rarely changed data

Connection Management:
- Tune pool sizes based on actual workload
- Use connection poolers (PgBouncer) for high-concurrency
- Implement circuit breakers for external service calls
- Monitor connection metrics and adjust accordingly

Async Best Practices:
- Use asyncio.gather for concurrent operations
- Implement proper timeout handling
- Use semaphores to limit concurrent external calls
- Profile async code with py-spy or similar tools

---

## External Resources

### Frameworks
- FastAPI: https://fastapi.tiangolo.com/
- Django: https://www.djangoproject.com/
- Flask: https://flask.palletsprojects.com/
- Starlette: https://www.starlette.io/

### Databases
- SQLAlchemy: https://docs.sqlalchemy.org/
- Motor (MongoDB): https://motor.readthedocs.io/
- Redis-py: https://redis-py.readthedocs.io/
- asyncpg: https://magicstack.github.io/asyncpg/

### Message Queues
- RabbitMQ: https://www.rabbitmq.com/documentation.html
- Apache Kafka: https://kafka.apache.org/documentation/
- aio-pika: https://aio-pika.readthedocs.io/

### Security
- OWASP API Security: https://owasp.org/www-project-api-security/
- python-jose: https://python-jose.readthedocs.io/
- Passlib: https://passlib.readthedocs.io/

### Monitoring
- Prometheus: https://prometheus.io/docs/
- OpenTelemetry: https://opentelemetry.io/docs/
- Grafana: https://grafana.com/docs/

### Best Practices
- 12 Factor App: https://12factor.net/
- Microservices Patterns: https://microservices.io/patterns/
- API Design Guidelines: https://github.com/microsoft/api-guidelines

---

Version: 1.0.0
Last Updated: 2025-12-06
