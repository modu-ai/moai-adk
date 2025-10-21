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
- Domain-related design documents and user requirements.
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
- Microsoft. "REST API Design Guidelines." https://learn.microsoft.com/azure/architecture/best-practices/api-design (accessed 2025-03-29).
- OpenAPI Initiative. "OpenAPI Specification." https://spec.openapis.org/oas/latest.html (accessed 2025-03-29).

## Changelog
- 2025-03-29: Codified input/output and failure responses for domain skills.

## Works well with

- alfred-trust-validation (API security validation)
- backend-expert (server implementation)
- security-expert (authentication patterns)

## Best Practices
- Record supporting documentation (version/link) for each domain decision.
- Review performance, security, and operational requirements simultaneously at an early stage.
