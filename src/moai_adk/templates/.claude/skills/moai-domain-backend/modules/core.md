    db.add_all(users)
    await db.commit()
    return users
```

## Deployment Patterns

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
      - name: api
        image: user-service:latest
        ports:
        - containerPort: 8000
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

## Installation Commands

```bash
# Core backend stack
pip install fastapi==0.118.0
pip install uvicorn[standard]
pip install sqlalchemy[asyncio]==2.0.0
pip install asyncpg
pip install pydantic==2.8.0

# Production addons
pip install redis
pip install slowapi  # Rate limiting
pip install fastapi-cache[redis]
pip install python-multipart  # File uploads

# Observability
pip install opentelemetry-api
pip install opentelemetry-sdk
pip install opentelemetry-exporter-jaeger
pip install prometheus-client

# Development
pip install pytest-asyncio
pip install httpx  # Async HTTP client for testing
```

## Best Practices

1. **Async by Default**: Use async/await for I/O operations
2. **Connection Pooling**: Configure appropriate pool sizes
3. **Error Handling**: Implement graceful degradation
4. **Security**: Input validation, rate limiting, authentication
5. **Testing**: Unit and integration tests with pytest-asyncio
6. **Documentation**: OpenAPI/Swagger auto-generation
7. **Monitoring**: Structured logging and metrics
8. **Versioning**: API versioning strategy


**Version**: 4.0.0 Enterprise
**Last Updated**: 2025-11-13
**Status**: Production Ready
**Enterprise Grade**: ✅ Full Enterprise Support


## Advanced Patterns

## Level 3: Advanced Integration

### Rate Limiting & Caching

```python
from slowapi import Limiter, _rate_limit_exceeded_handler
from slowapi.util import get_remote_address
from fastapi import Request

# Rate limiting
limiter = Limiter(key_func=get_remote_address)
app.state.limiter = limiter

@app.get("/api/users")
@limiter.limit("100/minute")
async def list_users(request: Request, db: AsyncSession = Depends(get_db)):
    # API with rate limiting
    users = await db.execute(select(User))
    return users.scalars().all()

# Redis caching
import redis.asyncio as redis
from fastapi_cache import FastAPICache, Coder
from fastapi_cache.backends.redis import RedisBackend

@app.post("/compute-heavy")
@cache(expire=60)  # Cache for 60 seconds
async def compute_heavy_operation(data: InputData):
    # Expensive computation cached in Redis
    result = await expensive_calculation(data)
    return result
```

### Microservices with Service Discovery

```python
# Service registration with Consul
import aiohttp
import asyncio

class ServiceRegistry:
    def __init__(self, consul_url: str):
        self.consul_url = consul_url
        self.service_id = f"user-service-{uuid4()}"

    async def register(self):
        async with aiohttp.ClientSession() as session:
            await session.put(
                f"{self.consul_url}/v1/agent/service/register",
                json={
                    "ID": self.service_id,
                    "Name": "user-service",
                    "Address": "user-service",
                    "Port": 8000,
                    "Check": {
                        "HTTP": "http://user-service:8000/health",
                        "Interval": "10s"
                    }
                }
            )

# Service discovery and load balancing
async def call_service(service_name: str, endpoint: str):
    services = await discover_services(service_name)
    selected = random.choice(services)
    url = f"http://{selected['Address']}:{selected['Port']}{endpoint}"

    async with aiohttp.ClientSession() as session:
        async with session.get(url) as response:
            return await response.json()
```

### API Gateway & Service Mesh

```yaml
# Istio Virtual Service
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: user-service
spec:
  http:
  - match:
    - uri:
        prefix: "/api/v1/users"
    route:
    - destination:
        host: user-service
        port:
          number: 8000
      weight: 100
    fault:
      delay:
        percentage:
          value: 0.1
        fixedDelay: 5s
```

### OpenTelemetry Observability

```python
from opentelemetry import trace, baggage
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

# Initialize tracing
tracer_provider = TracerProvider()
trace.set_tracer_provider(tracer_provider)

jaeger_exporter = JaegerExporter(
    agent_host_name="jaeger",
    agent_port=6831,
)

span_processor = BatchSpanProcessor(jaeger_exporter)
tracer_provider.add_span_processor(span_processor)

# Custom tracing in endpoints
@app.get("/users/{user_id}")
async def get_user(user_id: int, db: AsyncSession = Depends(get_db)):
    tracer = trace.get_tracer(__name__)

    with tracer.start_as_current_span("get_user") as span:
        span.set_attribute("user.id", user_id)

        with tracer.start_as_current_span("database_query"):
            result = await db.execute(select(User).where(User.id == user_id))
            user = result.scalar_one_or_none()

        if user:
            span.set_attribute("user.found", True)
            return user
        else:
            span.set_attribute("user.found", False)
            raise HTTPException(status_code=404)
```


## Context7 Integration

### Related Libraries & Tools
- [FastAPI](/tiangolo/fastapi): Modern, fast (high-performance), web framework for building APIs
- [Django](/django/django): High-level Python web framework for rapid development
- [SQLAlchemy](/sqlalchemy/sqlalchemy): Python SQL toolkit and Object Relational Mapper
- [Uvicorn](/encode/uvicorn): Lightning-fast ASGI server implementation
- [Pydantic](/pydantic/pydantic): Data validation using Python type annotations

### Official Documentation
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [Django 5.2](https://docs.djangoproject.com/en/5.2/)
- [SQLAlchemy 2.0](https://docs.sqlalchemy.org/en/20/)
- [Uvicorn](https://www.uvicorn.org/)
- [asyncpg](https://magicstack.github.io/asyncpg/)

### Version-Specific Guides
Latest stable version: FastAPI 0.118+, Django 5.2 LTS, SQLAlchemy 2.0
- [FastAPI 0.118 Release](https://github.com/tiangolo/fastapi/releases)
- [Django 5.2 Release Notes](https://docs.djangoproject.com/en/5.2/releases/5.2/)
- [SQLAlchemy 2.0 Migration](https://docs.sqlalchemy.org/en/20/changelog/migration_20.html)
- [Async Python Best Practices](https://fastapi.tiangolo.com/async/)

