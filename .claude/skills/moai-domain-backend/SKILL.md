---
name: moai-domain-backend
description: Provides backend architecture and scaling guidance; use when the project targets server-side APIs or infrastructure design decisions.
allowed-tools:
  - Read
  - Bash
---

# Backend Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Version | 2.0.0 |
| Created | 2025-10-22 |
| Updated | 2025-10-22 |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand for backend architecture requests |
| Trigger cues | Service layering, API orchestration, caching, background job design, cloud-native patterns, Kubernetes, Istio, observability, microservices, serverless, event-driven architecture. |
| Tier | 4 |

## What it does

Provides expertise in backend server architecture, RESTful API design, caching strategies, database optimization, cloud-native patterns, observability, and horizontal/vertical scalability patterns with modern tooling (Kubernetes, Istio, OpenTelemetry, Prometheus).

## When to use

- Engages when backend or service-architecture questions come up.
- "Backend architecture", "API design", "Caching strategy", "Scalability", "Kubernetes", "Observability"
- Automatically invoked when working with backend projects
- Backend SPEC implementation (`/alfred:2-run`)
- Cloud-native architecture planning
- Microservices design and orchestration

## How it works

### Server Architecture Patterns

**Layered Architecture**:
- **Controller → Service → Repository** pattern
- Clear separation of concerns
- Testable business logic
- Database abstraction layer

**Microservices**:
- **Service decomposition**: Domain-driven design boundaries
- **Container orchestration**: Kubernetes 1.31.x
- **Service mesh**: Istio 1.21.x for traffic management, security, observability
- **Inter-service communication**: REST, gRPC, message queues
- **API Gateway**: Kong, Ambassador, Traefik
- **Service discovery**: Kubernetes DNS, Consul

**Monoliths**:
- When appropriate (small team, low complexity, rapid prototyping)
- Modular monolith patterns
- Clear internal boundaries for future extraction

**Serverless**:
- **AWS Lambda**: Event-driven functions
- **Google Cloud Functions**: HTTP and event triggers
- **Azure Functions**: Durable Functions for workflows
- Cold start mitigation strategies
- Stateless design principles

**Event-Driven Architecture**:
- **CQRS**: Command Query Responsibility Segregation
- **Event Sourcing**: Immutable event logs
- **Apache Kafka 3.7.x**: Distributed streaming platform
- **RabbitMQ 3.13.x**: Message broker
- **Event versioning**: Schema evolution strategies

### API Design

**RESTful APIs**:
- **Resource-based**: Nouns over verbs
- **HTTP methods**: GET, POST, PUT, PATCH, DELETE
- **Stateless**: No server-side session
- **HATEOAS**: Hypermedia as the Engine of Application State
- **Pagination**: Cursor-based or offset-based
- **Filtering**: Query parameters for resource selection

**GraphQL**:
- **Schema-first**: Type definitions drive development
- **Resolver patterns**: Efficient data fetching
- **DataLoader**: Batching and caching
- **Subscriptions**: Real-time updates via WebSockets

**gRPC**:
- **Protocol Buffers**: Strongly-typed contracts
- **HTTP/2**: Multiplexing, server push
- **Streaming**: Unary, server-streaming, client-streaming, bidirectional

**WebSockets**:
- **Bidirectional communication**: Real-time updates
- **Connection management**: Heartbeats, reconnection
- **Authentication**: Token-based or session-based

### Security (OWASP API Security Top 10 2023)

- **API1:2023 Broken Object Level Authorization**: Verify user permissions
- **API2:2023 Broken Authentication**: Strong auth mechanisms
- **API3:2023 Broken Object Property Level Authorization**: Filter sensitive fields
- **API4:2023 Unrestricted Resource Consumption**: Rate limiting, pagination
- **API5:2023 Broken Function Level Authorization**: Role-based access control
- **API6:2023 Unrestricted Access to Sensitive Business Flows**: Anti-automation
- **API7:2023 Server Side Request Forgery**: Validate URLs, whitelist
- **API8:2023 Security Misconfiguration**: Secure defaults, hardening
- **API9:2023 Improper Inventory Management**: API versioning, documentation
- **API10:2023 Unsafe Consumption of APIs**: Validate external API responses

### Caching Strategies

**Redis 7.2.x**:
- **In-memory data store**: Key-value, hashes, lists, sets
- **Pub/Sub**: Real-time messaging
- **Lua scripting**: Atomic operations
- **Redis Cluster**: Horizontal scaling
- **Persistence**: RDB snapshots, AOF logs

**Memcached 1.6.x**:
- **Distributed caching**: Simple key-value store
- **LRU eviction**: Automatic memory management
- **Multi-threaded**: High concurrency

**Cache Patterns**:
- **Cache-aside**: Application manages cache
- **Write-through**: Write to cache and database
- **Write-behind**: Asynchronous database writes
- **Refresh-ahead**: Proactive cache warming

**CDN Caching**:
- **CloudFlare**, **CloudFront**, **Fastly**
- **Edge caching**: Geographically distributed
- **Cache invalidation**: Purge, TTL

### Database Optimization

**PostgreSQL 16.x**:
- **EXPLAIN ANALYZE**: Query plan analysis
- **Connection pooling**: PgBouncer, pgpool-II
- **Read replicas**: Streaming replication
- **Partitioning**: Table partitioning for large datasets
- **JSONB indexing**: GIN, GiST indexes

**MongoDB 8.0.x**:
- **Aggregation pipeline**: Complex queries
- **Sharding**: Horizontal scaling
- **Read concern/Write concern**: Consistency tuning
- **Change streams**: Real-time data feeds

**Cassandra 4.1.x**:
- **Wide-column store**: Time-series data
- **Tunable consistency**: CAP theorem trade-offs
- **Data modeling**: Partition keys, clustering columns

**Redis 7.2.x** (as database):
- **RediSearch**: Full-text search
- **RedisJSON**: Native JSON support
- **RedisTimeSeries**: Time-series data

### Cloud-Native Patterns

**Kubernetes 1.31.x**:
- **Deployments**: Rolling updates, rollbacks
- **StatefulSets**: Persistent storage for databases
- **Services**: ClusterIP, NodePort, LoadBalancer
- **Ingress**: HTTP/HTTPS routing
- **ConfigMaps/Secrets**: Configuration management
- **Horizontal Pod Autoscaler**: CPU/memory-based scaling
- **Vertical Pod Autoscaler**: Resource request optimization

**Istio 1.21.x Service Mesh**:
- **Traffic management**: Canary deployments, A/B testing
- **Security**: mTLS, authorization policies
- **Observability**: Distributed tracing, metrics
- **Resilience**: Circuit breaking, retries, timeouts

**Docker 27.0+**:
- **Multi-stage builds**: Smaller images
- **BuildKit**: Efficient caching
- **Docker Compose**: Local development
- **Image scanning**: Vulnerability detection

### Observability Stack (2025-10-22)

**OpenTelemetry 1.24.0**:
- **Unified telemetry**: Traces, metrics, logs
- **Language SDKs**: Auto-instrumentation
- **Collector**: Vendor-agnostic pipeline
- **Context propagation**: Distributed tracing

**Prometheus 2.48.x**:
- **Metrics collection**: Pull-based
- **PromQL**: Query language
- **Alertmanager**: Alert routing
- **Service discovery**: Kubernetes integration

**Jaeger 1.51.x**:
- **Distributed tracing**: Request flow visualization
- **Span analysis**: Latency breakdown
- **Service dependency graph**: Architecture mapping

**ELK Stack**:
- **Elasticsearch 8.x**: Log search and analytics
- **Logstash 8.x**: Log aggregation and transformation
- **Kibana 8.x**: Visualization and dashboards
- **Filebeat 8.x**: Log shipping

**Grafana 10.x**:
- **Unified dashboards**: Metrics, logs, traces
- **Data source plugins**: Prometheus, Loki, Tempo
- **Alerting**: Multi-channel notifications

### Scalability Patterns

**Horizontal Scaling**:
- **Load balancers**: NGINX, HAProxy, AWS ALB
- **Stateless services**: No local state
- **Session management**: Redis, database

**Vertical Scaling**:
- **Resource limits**: CPU, memory allocation
- **Database connection pooling**: Reduce overhead

**Async Processing**:
- **Apache Kafka 3.7.x**: Event streaming
- **RabbitMQ 3.13.x**: Task queues
- **Celery**: Python task queue
- **Bull/BullMQ**: Node.js job queues

**Rate Limiting**:
- **Token bucket**: Fixed rate with burst
- **Sliding window**: Time-based limits
- **Distributed rate limiting**: Redis-based

## Examples

See `examples.md` for production-ready patterns:
- Microservices with Kubernetes 1.31 + Istio 1.21
- Event-driven with Kafka 3.7
- Serverless with AWS Lambda
- Observability with OpenTelemetry 1.24 + Prometheus 2.48

## Inputs
- Domain-specific design documents and user requirements.
- Project technology stack and operational constraints.
- Performance, scalability, and availability requirements.

## Outputs
- Domain-specific architecture or implementation guidelines.
- Recommended list of associated sub-agents/skills.
- Tool version recommendations and deployment strategies.

## Failure Modes
- When the domain document does not exist or is ambiguous.
- When the project strategy is unconfirmed and cannot be specified.
- When performance/scalability requirements are not quantified.

## Dependencies
- `.moai/project/` document and latest technical briefing are required.
- `reference.md` for architecture decision matrices.
- `examples.md` for production patterns.

## References
- AWS. "AWS Well-Architected Framework." https://docs.aws.amazon.com/wellarchitected/latest/framework/ (accessed 2025-10-22).
- Heroku. "The Twelve-Factor App." https://12factor.net/ (accessed 2025-10-22).
- OWASP. "API Security Top 10 2023." https://owasp.org/API-Security/editions/2023/en/0x11-t10/ (accessed 2025-10-22).
- Kubernetes. "Kubernetes Documentation." https://kubernetes.io/docs/ (accessed 2025-10-22).
- OpenTelemetry. "OpenTelemetry Documentation." https://opentelemetry.io/docs/ (accessed 2025-10-22).
- Istio. "Istio Documentation." https://istio.io/latest/docs/ (accessed 2025-10-22).

## Changelog
- 2025-10-22: v2.0.0 - Added cloud-native patterns (Kubernetes 1.31, Istio 1.21), observability stack (OpenTelemetry 1.24, Prometheus 2.48, Jaeger 1.51), OWASP API Security Top 10 2023, latest database versions (PostgreSQL 16, MongoDB 8, Redis 7.2, Cassandra 4.1).
- 2025-03-29: v1.0.0 - Codified input/output and failure responses for domain skills.

## Works well with

- moai-alfred-trust-validation (backend testing)
- moai-domain-web-api (API design)
- moai-domain-database (database optimization)
- moai-domain-devops (CI/CD, infrastructure)
- moai-domain-security (OWASP, SAST)

## Best Practices
- Record supporting documentation (version/link) for each domain decision.
- Review performance, security, and operational requirements simultaneously at an early stage.
- Use observability stack from day one (OpenTelemetry + Prometheus + Jaeger).
- Apply OWASP API Security Top 10 2023 guidelines.
- Choose architecture patterns based on team size, complexity, and operational maturity.
- Test scalability patterns with realistic load (k6, Gatling).
