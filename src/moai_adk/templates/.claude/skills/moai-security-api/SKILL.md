---
name: moai-security-api
version: 4.0.0
updated: 2025-11-20
status: stable
description: API security patterns - authentication, authorization, rate limiting, OWASP
allowed-tools: [Read, Bash, WebSearch, WebFetch]
---

# API Security Expert

**Secure API Design & Implementation**

> **Focus**: Authentication, Authorization, Rate Limiting, OWASP API Top 10  
> **Stack**: OAuth 2.0, JWT, API Keys, CORS

---

## Overview

Comprehensive patterns for securing RESTful and GraphQL APIs.

### Core Security Layers

1.  **Authentication**: Who are you? (OAuth, JWT, API keys)
2.  **Authorization**: What can you do? (RBAC, ABAC)
3.  **Rate Limiting**: Prevent abuse (token bucket, sliding window)
4.  **Input Validation**: Prevent injection attacks

---

## Quick Start

### 1. JWT Authentication

Issue and verify JWT tokens for API access.

**Key Concepts**:

- Token structure: `Header.Payload.Signature`
- Short-lived tokens (<1 hour)
- Refresh token rotation

See: [examples.md](./examples.md#jwt-authentication) for implementation

### 2. Role-Based Access Control (RBAC)

Enforce permissions based on user roles.

**Pattern**: Decorator/middleware checks user role before allowing access.

See: [examples.md](./examples.md#rbac-implementation) for code

### 3. Rate Limiting

Prevent API abuse with token bucket algorithm.

**Common Limits**:

- Public endpoints: 100 req/min
- Authenticated: 1000 req/min
- Admin: Unlimited

See: [examples.md](./examples.md#rate-limiting) for implementation

### 4. CORS Configuration

Restrict cross-origin requests to trusted domains.

**Critical**: Never use `allow_origins=["*"]` in production.

See: [examples.md](./examples.md#cors-setup) for configuration

---

## OWASP API Security Top 10 (2023)

| #   | Vulnerability                                   | Mitigation                       |
| --- | ----------------------------------------------- | -------------------------------- |
| 1   | Broken Object Level Authorization               | Validate user owns resource      |
| 2   | Broken Authentication                           | OAuth 2.0, MFA                   |
| 3   | Broken Object Property Level Authorization      | Don't expose internal fields     |
| 4   | Unrestricted Resource Consumption               | Rate limiting, pagination        |
| 5   | Broken Function Level Authorization             | Verify permissions per endpoint  |
| 6   | Unrestricted Access to Sensitive Business Flows | CAPTCHA, anomaly detection       |
| 7   | Server Side Request Forgery (SSRF)              | Validate URLs, block private IPs |
| 8   | Security Misconfiguration                       | Disable debug, remove defaults   |
| 9   | Improper Inventory Management                   | Document endpoints, versioning   |
| 10  | Unsafe Consumption of APIs                      | Validate third-party responses   |

See: [reference.md](./reference.md#owasp-details) for detailed mitigations

---

## Best Practices

1.  **HTTPS Only**: Enforce TLS 1.3+
2.  **Short-Lived Tokens**: JWT expiry <1 hour
3.  **API Versioning**: `/v1/users`, `/v2/users`
4.  **Logging**: Log auth failures, suspicious patterns
5.  **Error Messages**: Don't leak system details

---

## Validation Checklist

- [ ] **Auth**: JWT/OAuth implemented?
- [ ] **AuthZ**: RBAC/ABAC enforced?
- [ ] **Rate Limiting**: Configured per endpoint?
- [ ] **CORS**: Restricted to trusted origins?
- [ ] **HTTPS**: TLS 1.3+ enforced?
- [ ] **Input**: Pydantic/Zod validation used?

---

## Related Skills

- `moai-security-auth`: Authentication patterns
- `moai-security-devsecops`: Security testing
- `moai-domain-backend`: API design

---

## Additional Resources

- [examples.md](./examples.md): Code implementations
- [reference.md](./reference.md): Complete API reference

---

**Last Updated**: 2025-11-20
