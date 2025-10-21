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
- 도메인 관련 설계 문서 및 사용자 요구사항.
- 프로젝트 기술 스택 및 운영 제약.

## Outputs
- 도메인 특화 아키텍처 또는 구현 가이드라인.
- 연관 서브 에이전트/스킬 권장 목록.

## Failure Modes
- 도메인 근거 문서가 없거나 모호할 때.
- 프로젝트 전략이 미확정이라 구체화할 수 없을 때.

## Dependencies
- `.moai/project/` 문서와 최신 기술 브리핑이 필요합니다.

## References
- AWS. "AWS Well-Architected Framework." https://docs.aws.amazon.com/wellarchitected/latest/framework/ (accessed 2025-03-29).
- Heroku. "The Twelve-Factor App." https://12factor.net/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 도메인 스킬에 대한 입력/출력 및 실패 대응을 명문화했습니다.

## Works well with

- alfred-trust-validation (backend testing)
- web-api-expert (API design)
- database-expert (database optimization)

## Best Practices
- 도메인 결정 사항마다 근거 문서(버전/링크)를 기록합니다.
- 성능·보안·운영 요구사항을 초기 단계에서 동시에 검토하세요.
