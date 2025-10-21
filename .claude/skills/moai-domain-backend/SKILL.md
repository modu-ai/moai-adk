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
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand for backend architecture requests |
| Trigger cues | Service layering, API orchestration, caching, background job design discussions. |
| Tier | 4 |

## What it does

Provides expertise in backend server architecture, RESTful API design, caching strategies, database optimization, and horizontal/vertical scalability patterns.

## When to use

- Engages when backend or service-architecture questions come up.
- “Backend architecture”, “API design”, “Caching strategy”, “Scalability”
- Automatically invoked when working with backend projects
- Backend SPEC implementation (`/alfred:2-run`)

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
```bash
$ make test-backend
$ k6 run perf/api-smoke.js
```

## Inputs
- Domain-specific design documents and user requirements.
- Project technology stack and operational constraints.

## Outputs
- Domain-specific architecture or implementation guidelines.
- Recommended list of associated sub-agents/skills.

## Failure Modes
- When the domain document does not exist or is ambiguous.
- When the project strategy is unconfirmed and cannot be specified.

## Dependencies
- `.moai/project/` document and latest technical briefing are required.

## References
- AWS. "AWS Well-Architected Framework." https://docs.aws.amazon.com/wellarchitected/latest/framework/ (accessed 2025-03-29).
- Heroku. "The Twelve-Factor App." https://12factor.net/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Codified input/output and failure responses for domain skills.

## Works well with

- alfred-trust-validation (backend testing)
- web-api-expert (API design)
- database-expert (database optimization)

## Best Practices
- Record supporting documentation (version/link) for each domain decision.
- Review performance, security, and operational requirements simultaneously at an early stage.
