---
name: moai-domain-backend
description: Server architecture, API design, caching strategies, and scalability
  patterns
allowed-tools:
- Read
- Bash
---

# Backend Expert

## What it does

Provides expertise in backend server architecture, RESTful API design, caching strategies, database optimization, and horizontal/vertical scalability patterns.

## When to use

- "백엔드 아키텍처", "API 설계", "캐싱 전략", "확장성"
- Automatically invoked when working with backend projects
- Backend SPEC implementation (`/alfred:2-build`)

## How it works

**Server Architecture**:
- **Layered architecture**: Controller → Service → Repository
- **Microservices**: Service decomposition, inter-service communication
- **Monoliths**: When appropriate (team size, complexity)
- **Serverless**: Functions as a Service (AWS Lambda, Cloud Functions)

**API Design**:
- **RESTful principles**: Resource-based, stateless
- **GraphQL**: Schema-first design
- **gRPC**: Protocol buffers for high performance
- **WebSockets**: Real-time bidirectional communication

**Caching Strategies**:
- **Redis**: In-memory data store
- **Memcached**: Distributed caching
- **Cache invalidation**: TTL, cache-aside pattern
- **CDN caching**: Static asset delivery

**Database Optimization**:
- **Connection pooling**: Reuse connections
- **Query optimization**: EXPLAIN analysis
- **Read replicas**: Horizontal scaling
- **Sharding**: Data partitioning

**Scalability Patterns**:
- **Horizontal scaling**: Load balancing across instances
- **Vertical scaling**: Increasing instance resources
- **Async processing**: Message queues (RabbitMQ, Kafka)
- **Rate limiting**: API throttling

## Examples

### Example 1: Layered architecture implementation
User: "/alfred:2-build BACKEND-001"
Claude: (creates RED API test, GREEN layered implementation, REFACTOR with caching)

### Example 2: Redis caching integration
User: "Redis 캐싱 추가"
Claude: (implements cache-aside pattern with Redis)

## Works well with

- alfred-trust-validation (backend testing)
- web-api-expert (API design)
- database-expert (database optimization)
