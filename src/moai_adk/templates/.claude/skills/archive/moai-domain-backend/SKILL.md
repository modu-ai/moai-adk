---
name: moai-domain-backend
description: Enterprise Backend Architecture with modern frameworks and best practices
version: 1.0.0
tier: Domain
status: active
test_coverage: 90
last_updated: 2025-11-24
---

# Enterprise Backend Architecture

## Quick Reference (30 seconds)

**moai-domain-backend** provides production-ready patterns for REST API design, microservices architecture, async/await optimization, authentication, error handling, and performance optimization. Use this when designing scalable backend systems with FastAPI, Django, Express, or similar frameworks.

**Core Capabilities**: REST/GraphQL API validation, microservice boundaries, async concurrency, JWT/OAuth2 auth, structured error handling, caching strategies, metrics collection.

---

## What It Does

### 1. API Design Validation
Validates REST endpoint configuration against RFC 7231 standards, enforces HTTP method and status code combinations, standardizes error response formats (RFC 7807), and supports multiple API versioning strategies.

### 2. Microservice Architecture
Defines service boundaries based on domain-driven design (DDD), configures inter-service communication patterns (REST, async, gRPC), sets up service discovery and registry (Consul, Eureka, etcd), and manages API gateway patterns.

### 3. Async/Await Optimization
Concurrent operation execution with timeout handling, async context managers for resource management, error handling in async code, and performance optimization for I/O-bound operations.

### 4. Authentication & Authorization
JWT token generation and validation (HS256), OAuth2 authorization code flow implementation, role-based access control (RBAC) with permission checking, and API key management.

### 5. Error Handling & Logging
Standardized error response formatting, correlation ID tracking for distributed tracing, structured async logging with context preservation, and error classification.

### 6. Performance Optimization
Multi-tier caching strategy (Redis, in-memory, CDN), rate limiting configuration (token bucket, sliding window), query optimization recommendations, and connection pooling management.

### 7. Metrics Collection
Real-time request/response performance tracking, error rate calculation and monitoring, service health status assessment, and performance trend analysis.

---

## When to Use

### ✅ Use moai-domain-backend when:
- Designing new backend systems with proven patterns
- Building REST or GraphQL APIs with validation
- Implementing microservices with clear boundaries
- Adding JWT/OAuth2 authentication
- Optimizing performance with caching
- Setting up monitoring and observability
- Handling errors systematically

### ❌ Avoid moai-domain-backend when:
- Building simple CRUD endpoints
- Working with legacy monolith codebases
- Implementing only frontend authentication
- Building CLI tools or scripts

---

## Key Features

1. **RFC-Compliant API Design** - RFC 7231 HTTP Semantics, RFC 7807 Problem Details, RESTful conventions
2. **Domain-Driven Service Decomposition** - Bounded contexts, service communication, event-driven support
3. **Production-Grade Async Patterns** - Concurrent task execution, timeout handling, memory efficiency
4. **Enterprise Security** - Cryptographic token generation, HMAC-SHA256 signing, automatic expiration
5. **Structured Observability** - Trace ID generation, contextual logging, metric aggregation
6. **Scalability Patterns** - Caching with invalidation, rate limiting with bursts, connection pooling
7. **Distributed Systems Patterns** - Circuit breaker, retry logic, bulkhead isolation, fallbacks

---

## Core Concepts (9 Architectural Patterns)

### 1-9. API Design, Microservices, Async, Auth, Error Handling, Performance, Versioning, Distributed Systems, Observability

Core patterns covering REST/GraphQL/gRPC design, microservice boundaries, concurrent execution, JWT/OAuth2, standardized errors, multi-tier caching, API versioning, resilience patterns, and comprehensive monitoring.

---

## Best Practices (DO and DON'T)

### ✅ DO
- Design APIs from consumer perspective
- Use async/await for I/O operations
- Implement proper error handling
- Validate all inputs
- Use connection pooling
- Cache strategically
- Monitor everything
- Version your APIs
- Implement rate limiting
- Use correlation IDs

### ❌ DON'T
- Block event loop with sync operations
- Hardcode configuration
- Log sensitive data
- Trust client input
- Ignore error handling
- Mix business logic with HTTP concerns
- Share databases between services
- Implement custom crypto
- Deploy without monitoring
- Skip database migrations

---

## Implementation Patterns (5 Scenarios)

### Pattern 1: FastAPI REST API with JWT
POST /api/v1/users with authentication, GET /api/v1/users/{id} public endpoint, validate input, structured error handling.

### Pattern 2: Async Concurrent Operations
Fetch multiple users concurrently with timeout, handle failures gracefully, return results.

### Pattern 3: Microservice Circuit Breaker
Detect cascading failures, open circuit on threshold, retry in half-open state, recover gracefully.

### Pattern 4: Error Handling with Logging
Request context with trace ID, structured logging with metadata, proper status codes, never expose secrets.

### Pattern 5: Redis Caching with Invalidation
Cache decorator with TTL, cache key generation, cache warming, automatic invalidation on updates.

---

## Real-World Examples (3 Systems)

### Example 1: E-Commerce Product API
List products with pagination (cache 1h), details with related items (cache 30m), admin operations invalidate caches, high-volume read optimization.

### Example 2: SaaS Multi-Tenant API
JWT authentication, tenant isolation, RBAC validation, per-tenant rate limiting, audit logging, version management.

### Example 3: Real-Time Notification Service
WebSocket connections with JWT auth, channel subscriptions, async message processing, offline message storage with retry.

---

## Performance Tips

1. Database: Indexes, column selection, pagination, pooling, EXPLAIN ANALYZE
2. Caching: Read-heavy data, short TTL for volatile, cache warming, hit rate monitoring
3. API: Compression, HTTP headers, field selection, request logging, percentile tracking
4. Async: I/O operations, batching, timeouts, queue depth monitoring, pooling
5. Monitoring: Percentiles, error rates, alerting, slow query logs, distributed tracing

---

**Version**: 1.0.0 | **Updated**: 2025-11-24 | **Coverage**: 90%+ | **Tier**: Domain
