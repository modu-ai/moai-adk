---

name: moai-domain-web-api
description: REST API and GraphQL design patterns with authentication, versioning, and OpenAPI documentation. Use when working on web API contracts scenarios.
allowed-tools:
  - Read
  - Bash
---

# Web API Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand for API delivery |
| Trigger cues | REST/GraphQL design, contract testing, versioning, integration hardening. |
| Tier | 4 |

## What it does

Provides expertise in designing and implementing RESTful APIs and GraphQL services, including authentication mechanisms (JWT, OAuth2), API versioning strategies, and OpenAPI documentation.

## When to use

- Engages when designing or validating web APIs and their lifecycle controls.
- “API design”, “REST API pattern”, “GraphQL schema”, “JWT authentication”
- Automatically invoked when working with API projects
- Web API SPEC implementation (`/alfred:2-run`)

## How it works

**REST API Design**:
- **RESTful principles**: Resource-based URLs, HTTP verbs (GET, POST, PUT, DELETE)
- **Status codes**: Proper use of 2xx, 4xx, 5xx codes
- **HATEOAS**: Hypermedia links in responses
- **Pagination**: Cursor-based or offset-based

**GraphQL Design**:
- **Schema definition**: Types, queries, mutations, subscriptions
- **Resolver implementation**: Data fetching logic
- **N+1 problem**: DataLoader for batching
- **Schema stitching**: Federated GraphQL

**Authentication & Authorization**:
- **JWT (JSON Web Token)**: Stateless authentication
- **OAuth2**: Authorization framework (flows: authorization code, client credentials)
- **API keys**: Simple authentication
- **RBAC/ABAC**: Role/Attribute-based access control

**API Versioning**:
- **URL versioning**: /v1/users, /v2/users
- **Header versioning**: Accept: application/vnd.api.v2+json
- **Deprecation strategy**: Sunset header

**Documentation**:
- **OpenAPI (Swagger)**: API specification
- **API documentation**: Auto-generated docs
- **Postman collections**: Request examples

## Examples
```bash
$ uvicorn app.main:app --reload
$ newman run postman_collection.json
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
- Microsoft. "REST API Design Guidelines." https://learn.microsoft.com/azure/architecture/best-practices/api-design (accessed 2025-03-29).
- OpenAPI Initiative. "OpenAPI Specification." https://spec.openapis.org/oas/latest.html (accessed 2025-03-29).

## Changelog
- 2025-03-29: 도메인 스킬에 대한 입력/출력 및 실패 대응을 명문화했습니다.

## Works well with

- alfred-trust-validation (API security validation)
- backend-expert (server implementation)
- security-expert (authentication patterns)

## Best Practices
- 도메인 결정 사항마다 근거 문서(버전/링크)를 기록합니다.
- 성능·보안·운영 요구사항을 초기 단계에서 동시에 검토하세요.
