---
name: moai-domain-web-api
description: REST API and GraphQL design patterns with authentication, versioning,
  and OpenAPI documentation
allowed-tools:
- Read
- Bash
---

# Web API Expert

## What it does

Provides expertise in designing and implementing RESTful APIs and GraphQL services, including authentication mechanisms (JWT, OAuth2), API versioning strategies, and OpenAPI documentation.

## When to use

- "API 설계", "REST API 패턴", "GraphQL 스키마", "JWT 인증"
- Automatically invoked when working with API projects
- Web API SPEC implementation (`/alfred:2-build`)

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

### Example 1: REST API with JWT
User: "/alfred:2-build API-AUTH-001"
Claude: (creates RED API test, GREEN implementation with JWT middleware, REFACTOR)

### Example 2: GraphQL schema design
User: "GraphQL 스키마 설계"
Claude: (designs schema with proper types and resolvers)

## Works well with

- alfred-trust-validation (API security validation)
- backend-expert (server implementation)
- security-expert (authentication patterns)
