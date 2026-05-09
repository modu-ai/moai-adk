---
name: moai-ref-api-patterns
description: >
  REST/GraphQL API design patterns, error handling conventions, and input validation
  reference for backend development. Agent-extending skill that amplifies expert-backend
  expertise with production-grade API patterns. Use when designing APIs, implementing
  endpoints, or reviewing backend code.
  NOT for: frontend development, DevOps, database schema design, security audits.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-03-30"
  tags: "api, rest, graphql, patterns, backend, reference"
  agent: "expert-backend"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000

# MoAI Extension: Triggers
triggers:
  keywords: ["api", "endpoint", "rest", "graphql", "route", "handler"]
  agents: ["expert-backend"]
  phases: ["run"]
---

# API Patterns Reference

## Target Agent

`expert-backend` - Applies these patterns directly to API implementation and review.

## RESTful API Design Conventions

| Principle | Convention | Example |
|-----------|-----------|---------|
| Resource Naming | Plural nouns, lowercase, kebab-case | `/api/v1/user-profiles` |
| Collection | GET returns array with pagination | `GET /users?page=1&limit=20` |
| Single Resource | GET returns object | `GET /users/{id}` |
| Create | POST to collection | `POST /users` |
| Update (full) | PUT to resource | `PUT /users/{id}` |
| Update (partial) | PATCH to resource | `PATCH /users/{id}` |
| Delete | DELETE to resource | `DELETE /users/{id}` |
| Nested Resources | Max 2 levels deep | `/users/{id}/posts` |
| Filtering | Query params | `?status=active&role=admin` |
| Sorting | Sort param | `?sort=-created_at,name` |
| Versioning | URL prefix | `/api/v1/`, `/api/v2/` |

## HTTP Status Code Guide

| Category | Code | When to Use |
|----------|------|-------------|
| Success | 200 OK | Successful GET, PUT, PATCH, DELETE |
| Success | 201 Created | Successful POST (resource created) |
| Success | 204 No Content | Successful DELETE (no body) |
| Client Error | 400 Bad Request | Malformed request, validation failure |
| Client Error | 401 Unauthorized | Missing or invalid authentication |
| Client Error | 403 Forbidden | Authenticated but not authorized |
| Client Error | 404 Not Found | Resource does not exist |
| Client Error | 409 Conflict | Resource state conflict (duplicate) |
| Client Error | 422 Unprocessable | Valid syntax but semantic error |
| Client Error | 429 Too Many | Rate limit exceeded |
| Server Error | 500 Internal | Unexpected server error |
| Server Error | 503 Service Unavailable | Maintenance or overload |

## Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Input validation failed",
    "details": [
      {"field": "email", "message": "Must be a valid email address"},
      {"field": "age", "message": "Must be between 0 and 150"}
    ],
    "request_id": "req_abc123"
  }
}
```

Rules:
- Never expose stack traces or internal details in production
- Always include request_id for traceability
- Use consistent error codes (ENUM, not free text)
- Login failures: "Invalid email or password" (never reveal which)

## Pagination Pattern

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8,
    "has_next": true,
    "has_prev": false
  }
}
```

For cursor-based (large datasets):
```json
{
  "data": [...],
  "cursor": {
    "next": "eyJpZCI6MTAwfQ==",
    "has_more": true
  }
}
```

## Input Validation Checklist

| Validation | Method | Tool |
|-----------|--------|------|
| Type validation | Schema validation | Zod, Joi, pydantic, Go validator |
| Length limits | Min/max constraints | Schema min/max |
| Pattern matching | Regex | Email, URL, phone patterns |
| Range validation | Number/date bounds | min/max values |
| Enumeration | Allowed values | enum types |
| SQL Injection | Parameterized queries | ORM (Prisma, GORM, SQLAlchemy) |
| XSS | HTML escaping | Template engines, DOMPurify |
| Path Traversal | Path normalization | filepath.Clean + whitelist |

## Rate Limiting Strategy

| Target | Limit | Key |
|--------|-------|-----|
| Auth endpoints | 5 req/min | IP |
| General API | 100 req/min | User token |
| File upload | 10 req/hour | User token |
| Public API | 30 req/min | IP |

Response headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`, `Retry-After` (on 429).

## API Versioning Strategy

| Strategy | Use Case | Example |
|----------|----------|---------|
| URL prefix | Most APIs | `/api/v1/users` |
| Header | Internal APIs | `Accept: application/vnd.api+json; version=2` |
| Query param | Simple APIs | `/users?version=2` |

Breaking changes that require version bump:
- Removing or renaming fields
- Changing field types
- Removing endpoints
- Changing authentication methods

Non-breaking changes (no version bump needed):
- Adding new optional fields
- Adding new endpoints
- Adding new query parameters

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "REST naming conventions are just aesthetics" | Consistent resource naming is how clients discover and predict endpoints. Inconsistency multiplies documentation burden. |
| "GraphQL solves over-fetching, so I do not need to design response shapes" | GraphQL shifts complexity to the resolver layer. Poorly designed schemas create N+1 queries and authorization gaps. |
| "Error codes are internal details, clients just need the message" | Clients need machine-readable error codes for programmatic handling. Messages are for humans, codes are for code. |
| "PATCH and PUT are interchangeable" | PATCH applies partial updates; PUT replaces the entire resource. Using them incorrectly breaks idempotency expectations. |
| "I will version the API when it becomes necessary" | Versioning after breaking changes forces emergency migrations. Plan versioning from the first release. |

**Hyrum's Law**: Every observable API behavior will eventually be depended on by clients. Undocumented response fields, error formats, and timing characteristics become implicit contracts.

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- API returns different error formats across endpoints
- Resource names use verbs instead of nouns (e.g., /getUser instead of /users/:id)
- No pagination on list endpoints that can return unbounded results
- Breaking change deployed without API version bump
- GraphQL schema allows unbounded depth or circular queries without limits

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] All endpoints follow consistent naming convention (nouns, plurals, nested resources)
- [ ] Error responses use a standard format with machine-readable error code
- [ ] List endpoints implement pagination with documented limits
- [ ] API versioning strategy present and enforced (URL path, header, or query param)
- [ ] Breaking vs non-breaking change classification documented for recent changes
- [ ] Input validation returns 400 with specific field-level error details

<!-- moai:evolvable-end -->
