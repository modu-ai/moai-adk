---
name: moai-domain-backend
description: Backend development specialist covering API design, database integration, microservices architecture, and modern backend patterns
version: 1.0.0
category: domain
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
tags:
 - backend
 - api
 - database
 - microservices
 - architecture
updated: 2025-12-06
status: active
author: MoAI-ADK Team
---

# Backend Development Specialist

## Quick Reference (30 seconds)

Backend Development Mastery - Comprehensive backend development patterns covering API design, database integration, microservices, and modern architecture patterns.

Core Capabilities:
- API Design: REST, GraphQL, gRPC with OpenAPI 3.1
- Database Integration: PostgreSQL, MongoDB, Redis, caching strategies
- Microservices: Service mesh, distributed patterns, event-driven architecture
- Security: Authentication, authorization, OWASP compliance
- Performance: Caching, optimization, monitoring, scaling

When to Use:
- Backend API development and architecture
- Database design and optimization
- Microservices implementation
- Performance optimization and scaling
- Security integration for backend systems

---

## Implementation Guide

### API Design Patterns

RESTful API Architecture:
```python
from fastapi import FastAPI, Depends, HTTPException
from fastapi.security import HTTPBearer
from pydantic import BaseModel
from typing import List, Optional

app = FastAPI(title="Modern API", version="2.0.0")
security = HTTPBearer()

class UserResponse(BaseModel):
 id: int
 email: str
 name: str

@app.get("/users", response_model=List[UserResponse])
async def list_users(token: str = Depends(security)):
 """List users with authentication."""
 return await user_service.get_all_users()

@app.post("/users", response_model=UserResponse)
async def create_user(user: UserCreate):
 """Create new user with validation."""
 return await user_service.create(user)
```

GraphQL Implementation:
```python
import strawberry
from typing import List

@strawberry.type
class User:
 id: int
 email: str
 name: str

@strawberry.type
class Query:
 @strawberry.field
 async def users(self) -> List[User]:
 return await user_service.get_all_users()

schema = strawberry.Schema(query=Query)
```

### Database Integration Patterns

PostgreSQL with SQLAlchemy:
```python
from sqlalchemy import create_engine, Column, Integer, String
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

Base = declarative_base()

class User(Base):
 __tablename__ = "users"

 id = Column(Integer, primary_key=True)
 email = Column(String, unique=True)
 name = Column(String)

# Connection pooling and optimization
engine = create_engine(
 DATABASE_URL,
 pool_size=20,
 max_overflow=30,
 pool_pre_ping=True
)
```

MongoDB with Motor:
```python
from motor.motor_asyncio import AsyncIOMotorClient
from pymongo import IndexModel

class UserService:
 def __init__(self, client: AsyncIOMotorClient):
 self.db = client.myapp
 self.users = self.db.users

 # Index optimization
 self.users.create_indexes([
 IndexModel("email", unique=True),
 IndexModel("created_at")
 ])

 async def create_user(self, user_data: dict) -> str:
 result = await self.users.insert_one(user_data)
 return str(result.inserted_id)
```

### Microservices Architecture

Service Discovery with Consul:
```python
import consul

class ServiceRegistry:
 def __init__(self, consul_host="localhost", consul_port=8500):
 self.consul = consul.Consul(host=consul_host, port=consul_port)

 def register_service(self, service_name: str, service_id: str, port: int):
 self.consul.agent.service.register(
 name=service_name,
 service_id=service_id,
 port=port,
 check=consul.Check.http(f"http://localhost:{port}/health", interval="10s")
 )

 def discover_service(self, service_name: str) -> List[str]:
 _, services = self.consul.health.service(service_name, passing=True)
 return [f"{s['Service']['Address']}:{s['Service']['Port']}" for s in services]
```

Event-Driven Architecture:
```python
import asyncio
from aio_pika import connect_robust

class EventBus:
 def __init__(self, amqp_url: str):
 self.connection = None
 self.channel = None
 self.amqp_url = amqp_url

 async def connect(self):
 self.connection = await connect_robust(self.amqp_url)
 self.channel = await self.connection.channel()

 async def publish_event(self, event_type: str, data: dict):
 await self.channel.default_exchange.publish(
 aio_pika.Message(
 json.dumps({"type": event_type, "data": data}).encode(),
 content_type="application/json"
 ),
 routing_key=event_type
 )
```

---

## Advanced Patterns

### Caching Strategies

Redis Integration:
```python
import redis.asyncio as redis
from functools import wraps
import json
import hashlib

class CacheManager:
 def __init__(self, redis_url: str):
 self.redis = redis.from_url(redis_url)

 def cache_result(self, ttl: int = 3600):
 def decorator(func):
 @wraps(func)
 async def wrapper(*args, kwargs):
 cache_key = self._generate_cache_key(func.__name__, args, kwargs)

 # Try to get from cache
 cached = await self.redis.get(cache_key)
 if cached:
 return json.loads(cached)

 # Execute function and cache result
 result = await func(*args, kwargs)
 await self.redis.setex(
 cache_key,
 ttl,
 json.dumps(result, default=str)
 )
 return result
 return wrapper
 return decorator
```

### Security Implementation

JWT Authentication:
```python
import jwt
from datetime import datetime, timedelta
from passlib.context import CryptContext

class SecurityManager:
 def __init__(self, secret_key: str):
 self.secret_key = secret_key
 self.pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")

 def hash_password(self, password: str) -> str:
 return self.pwd_context.hash(password)

 def verify_password(self, plain_password: str, hashed_password: str) -> bool:
 return self.pwd_context.verify(plain_password, hashed_password)

 def create_access_token(self, data: dict, expires_delta: timedelta = None) -> str:
 to_encode = data.copy()
 if expires_delta:
 expire = datetime.utcnow() + expires_delta
 else:
 expire = datetime.utcnow() + timedelta(minutes=15)

 to_encode.update({"exp": expire})
 return jwt.encode(to_encode, self.secret_key, algorithm="HS256")
```

### Performance Optimization

Database Connection Pooling:
```python
from sqlalchemy.pool import QueuePool
from sqlalchemy import event

def create_optimized_engine(database_url: str):
 engine = create_engine(
 database_url,
 poolclass=QueuePool,
 pool_size=20,
 max_overflow=30,
 pool_pre_ping=True,
 pool_recycle=3600,
 echo=False
 )

 @event.listens_for(engine, "before_cursor_execute")
 def receive_before_cursor_execute(conn, cursor, statement, parameters, context, executemany):
 context._query_start_time = time.time()

 @event.listens_for(engine, "after_cursor_execute")
 def receive_after_cursor_execute(conn, cursor, statement, parameters, context, executemany):
 total = time.time() - context._query_start_time
 if total > 0.1: # Log slow queries
 logger.warning(f"Slow query: {total:.2f}s - {statement[:100]}")

 return engine
```

---

## Works Well With

- moai-domain-frontend - Full-stack development integration
- moai-domain-database - Advanced database patterns
- moai-foundation-core - MCP server development patterns for backend services
- moai-quality-security - Security validation and compliance
- moai-foundation-core - Core architectural principles

---

## Technology Stack

Primary Technologies:
- Languages: Python 3.13+, Node.js 20+, Go 1.23
- Frameworks: FastAPI, Django, Express.js, Gin
- Databases: PostgreSQL 16+, MongoDB 7+, Redis 7+
- Message Queues: RabbitMQ, Apache Kafka, Redis Pub/Sub
- Containerization: Docker, Kubernetes
- Monitoring: Prometheus, Grafana, OpenTelemetry

Integration Patterns:
- RESTful APIs with OpenAPI 3.1
- GraphQL with Apollo Federation
- gRPC for high-performance services
- Event-driven architecture with CQRS
- API Gateway patterns
- Circuit breakers and resilience patterns

---

Status: Production Ready
Last Updated: 2025-11-30
Maintained by: MoAI-ADK Backend Team
