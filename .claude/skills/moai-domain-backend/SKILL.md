---
name: moai-domain-backend
description: >
  Backend development specialist covering API design, database integration,
  microservices architecture, and modern backend patterns.
  Use when user asks about API design, REST or GraphQL endpoints, server implementation,
  authentication, authorization, middleware, or backend service architecture.
  Do NOT use for database-specific schema design or query optimization
  (use moai-domain-database instead) or frontend implementation
  (use moai-domain-frontend instead).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Bash(npm:*) Bash(npx:*) Bash(node:*) Bash(uv:*) Bash(pip:*) Bash(pytest:*) Bash(ruff:*) Bash(docker:*) Bash(curl:*) Bash(go:*) Bash(cargo:*) Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-01-11"
  modularized: "false"
  tags: "backend, api, database, microservices, architecture"
  author: "MoAI-ADK Team"

# MoAI Extension: Triggers
triggers:
  keywords:
    - backend
    - API
    - server
    - authentication
    - authorization
    - REST
    - GraphQL
    - gRPC
    - microservices
    - database
    - endpoint
    - middleware
    - FastAPI
    - Express
    - Django
    - Flask
    - serverless
    - caching
    - Redis
    - PostgreSQL
    - MongoDB
---

# Backend Development Specialist

## Quick Reference

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

Create a FastAPI application with authentication and response models. Define a Pydantic UserResponse model with id, email, and name fields. Implement list_users and create_user endpoints with HTTPBearer security dependency. The list endpoint returns a list of UserResponse objects, while the create endpoint accepts a UserCreate model and returns a single UserResponse.

GraphQL Implementation:

Use Strawberry to define GraphQL types. Create a User type with id, email, and name fields. Define a Query type with a users resolver that returns a list of User objects asynchronously. Generate the schema by passing the Query type to strawberry.Schema.

### Database Integration Patterns

PostgreSQL with SQLAlchemy:

Define SQLAlchemy models using declarative_base. Create a User model with id as primary key, email as unique string, and name as string column. Configure the engine with connection pooling parameters including pool_size of 20, max_overflow of 30, and pool_pre_ping enabled for connection health checks.

MongoDB with Motor:

Create a UserService class that initializes with an AsyncIOMotorClient. Set up the database and users collection in the constructor. Create indexes for email (unique) and created_at fields. Implement create_user method that inserts a document and returns the inserted_id as string.

### Microservices Architecture

Service Discovery with Consul:

Create a ServiceRegistry class that connects to Consul. Implement register_service method that registers a service with name, id, port, and health check endpoint. Implement discover_service method that queries healthy services and returns list of adddess:port strings.

Event-Driven Architecture:

Create an EventBus class using aio_pika for AMQP messaging. Implement connect method to establish connection and channel. Implement publish_event method that serializes event type and data as JSON and publishes to the default exchange with routing_key matching the event type.

---

## Advanced Patterns

### Caching Strategies

Redis Integration:

Create a CacheManager class with Redis connection. Implement a cache_result decorator that accepts ttl parameter. The decorator generates cache keys from function name and arguments, checks Redis for cached results, executes the function on cache miss, and stores results with expiration. Use json.loads and json.dumps for serialization.

### Security Implementation

JWT Authentication:

Create a SecurityManager class with CryptContext for bcrypt password hashing. Implement hash_password and verify_password methods using the context. Implement create_access_token that encodes a JWT with expiration time using HS256 algorithm. Default expiration is 15 minutes if not specified.

### Performance Optimization

Database Connection Pooling:

Create an optimized SQLAlchemy engine with QueuePool, pool_size 20, max_overflow 30, pool_pre_ping enabled, and pool_recycle of 3600 seconds. Add event listeners for before_cursor_execute and after_cursor_execute to track query timing. Log warnings for queries exceeding 100ms threshold.

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

## Resources

For working code examples, see [examples.md](examples.md).

Status: Production Ready
Last Updated: 2026-01-11
Maintained by: MoAI-ADK Backend Team
