---
name: moai-domain-backend
description: Enterprise Backend Architecture with modern frameworks, Context7 integration, and production-ready patterns for 2025
version: 1.0.0
modularized: false
tags:
  - architecture
  - enterprise
  - patterns
  - backend
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: domain, moai, backend  


## Quick Reference (30 seconds)

# Enterprise Backend Architecture Expert

**Latest Frameworks (2025)**:
- **FastAPI 0.115+** - Modern async Python with enhanced dependencies
- **Django 5.0+** - Async views, signals improvements, database optimization
- **Express.js 5.x** - Enhanced middleware, async error handling
- **HTTP/3 (QUIC)** - Next-generation protocol support

**Key Capabilities**:
- Modern async/await patterns across all frameworks
- Advanced dependency injection and middleware
- Database connection pooling and optimization
- Real-time capabilities and WebSocket support
- Production-grade error handling

**When to Use**:
- Backend architecture design and framework selection
- API design with RESTful or GraphQL patterns
- Database integration and performance optimization
- Authentication, authorization, and security implementation

---

## Implementation Guide

### FastAPI 0.115+ (2025 LTS)

**Enhanced Dependencies with Background Tasks**:
```python
from fastapi import FastAPI, Depends, BackgroundTasks
from typing import Annotated
from contextvars import ContextVar

app = FastAPI()

# Context variable for request tracking
request_context: ContextVar[dict] = ContextVar("request_context", default={})

async def get_database():
    """Database dependency with proper cleanup."""
    db = await create_connection_pool()
    try:
        yield db
    finally:
        await db.close()

async def verify_token(token: str = Header(...)):
    """Enhanced token verification with caching."""
    user = await cached_token_verification(token)
    if not user:
        raise HTTPException(status_code=401, detail="Invalid token")
    return user

@app.post("/tasks/")
async def create_task(
    task_data: TaskCreate,
    background_tasks: BackgroundTasks,
    db: Annotated[Database, Depends(get_database)],
    user: Annotated[User, Depends(verify_token)]
):
    """Create task with background processing."""
    task = await db.tasks.create(task_data, user_id=user.id)
    
    # Add background task for notifications
    background_tasks.add_task(send_notification, task.id, user.email)
    background_tasks.add_task(update_analytics, task.type)
    
    return {"task_id": task.id, "status": "processing"}
```

**Advanced Middleware with Context Variables**:
```python
from starlette.middleware.base import BaseHTTPMiddleware

class RequestContextMiddleware(BaseHTTPMiddleware):
    """Track request context across async calls."""
    
    async def dispatch(self, request, call_next):
        # Set request context
        token = request_context.set({
            "request_id": str(uuid.uuid4()),
            "path": request.url.path,
            "method": request.method
        })
        
        try:
            response = await call_next(request)
            return response
        finally:
            request_context.reset(token)

app.add_middleware(RequestContextMiddleware)
```

### Django 5.0+ (2025 LTS)

**Async Views with Enhanced Signals**:
```python
from django.http import JsonResponse
from django.views import View
from asgiref.sync import sync_to_async
import asyncio

class AsyncAPIView(View):
    """Modern async class-based view."""
    
    async def get(self, request, *args, **kwargs):
        # Parallel async operations
        user_data, stats_data = await asyncio.gather(
            self.get_user_data(request.user.id),
            self.get_stats_data(request.user.id)
        )
        
        return JsonResponse({
            "user": user_data,
            "stats": stats_data
        })
    
    @sync_to_async
    def get_user_data(self, user_id):
        return User.objects.get(id=user_id).to_dict()
    
    @sync_to_async
    def get_stats_data(self, user_id):
        return UserStats.objects.filter(user_id=user_id).aggregate(
            total_views=Count('id'),
            avg_duration=Avg('duration')
        )
```

**Async Signal Handlers**:
```python
from django.db.models.signals import post_save
from django.dispatch import receiver
import asyncio

@receiver(post_save, sender=Order)
async def process_order_async(sender, instance, created, **kwargs):
    """Async signal handler for order processing."""
    if created:
        await asyncio.gather(
            send_confirmation_email(instance.email),
            update_inventory(instance.items),
            trigger_fulfillment(instance.id)
        )
```

**Database Optimization with Connection Pooling**:
```python
# settings.py
DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.postgresql',
        'NAME': 'mydb',
        'USER': 'user',
        'PASSWORD': 'password',
        'HOST': 'localhost',
        'PORT': '5432',
        'CONN_MAX_AGE': 600,  # Connection pooling
        'CONN_HEALTH_CHECKS': True,  # Django 5.0+
        'OPTIONS': {
            'connect_timeout': 10,
            'options': '-c statement_timeout=30000'
        }
    }
}
```

### Express.js 5.x (2024+)

**Enhanced Middleware Chaining**:
```javascript
const express = require('express');
const app = express();

// Async error handling middleware
const asyncHandler = (fn) => (req, res, next) => {
    Promise.resolve(fn(req, res, next)).catch(next);
};

// Enhanced logging middleware
app.use((req, res, next) => {
    req.startTime = Date.now();
    res.on('finish', () => {
        const duration = Date.now() - req.startTime;
        console.log(`${req.method} ${req.path} - ${res.statusCode} (${duration}ms)`);
    });
    next();
});

// Database connection middleware
app.use(asyncHandler(async (req, res, next) => {
    req.db = await getDbConnection();
    next();
}));

// Routes with async/await
app.get('/users/:id', asyncHandler(async (req, res) => {
    const user = await req.db.users.findById(req.params.id);
    if (!user) {
        throw new NotFoundError('User not found');
    }
    res.json(user);
}));

// Error handling middleware (must be last)
app.use((err, req, res, next) => {
    console.error(err.stack);
    res.status(err.statusCode || 500).json({
        error: err.message,
        ...(process.env.NODE_ENV === 'development' && { stack: err.stack })
    });
});
```

### HTTP/2 & HTTP/3 Support

**Node.js HTTP/2 Server**:
```javascript
const http2 = require('http2');
const fs = require('fs');

const server = http2.createSecureServer({
    key: fs.readFileSync('server-key.pem'),
    cert: fs.readFileSync('server-cert.pem')
});

server.on('stream', (stream, headers) => {
    // HTTP/2 server push
    stream.pushStream({ ':path': '/style.css' }, (err, pushStream) => {
        if (err) throw err;
        pushStream.respondWithFile('style.css');
    });
    
    // Main response
    stream.respond({
        'content-type': 'text/html',
        ':status': 200
    });
    stream.end('<html><body>Hello HTTP/2</body></html>');
});

server.listen(3000);
```

### Database Connection Pooling

**PostgreSQL with PgBouncer**:
```python
from sqlalchemy import create_engine
from sqlalchemy.pool import QueuePool

# Optimized connection pool
engine = create_engine(
    'postgresql://user:pass@localhost/db',
    poolclass=QueuePool,
    pool_size=20,          # Core connections
    max_overflow=10,       # Additional connections
    pool_timeout=30,       # Connection timeout
    pool_recycle=3600,     # Recycle after 1 hour
    pool_pre_ping=True,    # Check connection health
    echo_pool=True         # Debug pool activity
)

# Connection with automatic retry
async def get_connection():
    retries = 3
    for attempt in range(retries):
        try:
            async with engine.begin() as conn:
                yield conn
                break
        except OperationalError:
            if attempt == retries - 1:
                raise
            await asyncio.sleep(0.1 * (2 ** attempt))
```

### Query Optimization Patterns

**Index Strategy**:
```sql
-- Composite indexes for common queries
CREATE INDEX idx_users_email_active ON users(email, is_active);
CREATE INDEX idx_orders_user_status ON orders(user_id, status);

-- Partial indexes for filtered queries
CREATE INDEX idx_active_users ON users(email) WHERE is_active = true;

-- BRIN indexes for time-series data
CREATE INDEX idx_logs_created ON logs USING BRIN(created_at);

-- GiST indexes for full-text search
CREATE INDEX idx_posts_search ON posts USING GiST(to_tsvector('english', content));
```

**Query Analysis**:
```python
async def analyze_slow_queries(db_connection):
    """Identify and optimize slow queries."""
    slow_queries = await db_connection.execute("""
        SELECT query, calls, total_time, mean_time
        FROM pg_stat_statements
        WHERE mean_time > 100  -- queries > 100ms
        ORDER BY total_time DESC
        LIMIT 20
    """)
    
    for query in slow_queries:
        # Analyze execution plan
        explain = await db_connection.execute(f"EXPLAIN ANALYZE {query.query}")
        print(f"Query: {query.query[:100]}...")
        print(f"Mean time: {query.mean_time}ms")
        print(f"Execution plan: {explain}")
```

---

## Advanced Patterns

### gRPC-Web Implementation

**Modern gRPC Service**:
```python
from grpc import aio
import user_service_pb2
import user_service_pb2_grpc

class UserServicer(user_service_pb2_grpc.UserServiceServicer):
    """Async gRPC service implementation."""
    
    async def GetUser(self, request, context):
        user = await db.users.find_one({"id": request.user_id})
        if not user:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details('User not found')
            return user_service_pb2.UserResponse()
        
        return user_service_pb2.UserResponse(
            user_id=user['id'],
            email=user['email'],
            name=user['name']
        )

async def serve():
    server = aio.server()
    user_service_pb2_grpc.add_UserServiceServicer_to_server(
        UserServicer(), server
    )
    server.add_insecure_port('[::]:50051')
    await server.start()
    await server.wait_for_termination()
```

### Real-time WebSocket Support

**FastAPI WebSocket with Broadcast**:
```python
from fastapi import WebSocket, WebSocketDisconnect
from typing import List

class ConnectionManager:
    def __init__(self):
        self.active_connections: List[WebSocket] = []
    
    async def connect(self, websocket: WebSocket):
        await websocket.accept()
        self.active_connections.append(websocket)
    
    def disconnect(self, websocket: WebSocket):
        self.active_connections.remove(websocket)
    
    async def broadcast(self, message: str):
        for connection in self.active_connections:
            await connection.send_text(message)

manager = ConnectionManager()

@app.websocket("/ws/{client_id}")
async def websocket_endpoint(websocket: WebSocket, client_id: str):
    await manager.connect(websocket)
    try:
        while True:
            data = await websocket.receive_text()
            await manager.broadcast(f"Client {client_id}: {data}")
    except WebSocketDisconnect:
        manager.disconnect(websocket)
        await manager.broadcast(f"Client {client_id} disconnected")
```

---

## Reference & Resources

### Context7 Documentation Access

**Latest Framework Patterns**:
- FastAPI: `/fastapi/fastapi` - Latest async patterns, dependencies, middleware
- Django: `/django/django` - Async views, signals, ORM optimization
- Express: `/expressjs/express` - Middleware, routing, error handling

**Performance Optimization**:
- PostgreSQL: `/postgresql/postgresql` - Indexing, query optimization
- Redis: `/redis/redis` - Caching strategies, pub/sub patterns

---

## Best Practices

### DO
- ✅ Use async/await for I/O-bound operations
- ✅ Implement proper connection pooling
- ✅ Add comprehensive error handling
- ✅ Optimize database queries with indexes
- ✅ Use middleware for cross-cutting concerns
- ✅ Implement request validation and sanitization
- ✅ Monitor performance with profiling tools

### DON'T
- ❌ Block event loop with synchronous operations
- ❌ Create new connections per request
- ❌ Skip error handling in async code
- ❌ Use N+1 queries without optimization
- ❌ Ignore middleware execution order
- ❌ Skip input validation and sanitization
- ❌ Deploy without performance testing

---

**Last Updated**: 2025-11-22
**Version**: 5.0.0
**Status**: Production Ready (2025 Standards)

---

## Works Well With

- `moai-domain-database` - Database design and optimization patterns
- `moai-domain-web-api` - RESTful/GraphQL API design patterns
- `moai-security-api` - Backend API security best practices
- `moai-essentials-perf` - Backend performance optimization
- `moai-domain-devops` - Deployment and CI/CD integration
- `moai-domain-testing` - Backend testing strategies

---

## Best Practices

### ✅ DO
- **Use async/await patterns** - Modern frameworks (FastAPI, Django 5.0+) support async natively
- **Implement proper dependency injection** - Clean separation of concerns and testability
- **Apply Context7 latest patterns** - Leverage up-to-date framework best practices
- **Use connection pooling** - Optimize database connections for scalability
- **Implement comprehensive error handling** - Structured error responses with proper HTTP codes
- **Enable request validation** - Use Pydantic/Zod for automatic input validation
- **Apply rate limiting** - Protect APIs from abuse and DoS attacks

### ❌ DON'T
- **Block event loops** - Avoid synchronous I/O in async contexts
- **Skip input validation** - Always validate and sanitize user input
- **Hardcode configurations** - Use environment variables and config files
- **Ignore error handling** - Implement proper exception handling and logging
- **Skip database migrations** - Use Alembic/Prisma for schema versioning
- **Expose sensitive data** - Filter responses and implement proper authorization

---

