---
name: moai-security-api
description: Comprehensive API security for REST, GraphQL, and gRPC services with OAuth 2.1, JWT, rate limiting, and multi-tenant patterns
modularized: true
modules:
  - authentication-authorization
  - rate-limiting-protection
  - advanced-security-patterns
---

## Quick Reference (30 seconds)

# moai-security-api

**API Security Expert**

> **Trust Score**: 9.9/10 | **Version**: 4.0.0 | **Enterprise Security**

---

## Core Purpose

Comprehensive API security for REST, GraphQL, and gRPC services with production-ready authentication, authorization, and protection patterns.

**API Attack Surface**:
```
User → [REST/GraphQL/gRPC Endpoint] → Internal Resources
        ↓
    - Missing Authentication
    - Broken Authorization  
    - Excessive Data Exposure
    - Rate Limit Bypass
    - Injection Attacks
```

**OWASP API Security Top 10 (2023)**:
1. **Broken Object Level Authorization** (BOLA)
2. **Broken Authentication**
3. **Excessive Data Exposure**
4. **Lack of Resources & Rate Limiting**
5. **Broken Function Level Authorization** (BFLA)
6. **Mass Assignment**
7. **Cross-Site Scripting (XSS)**
8. **Broken API Versioning**
9. **Improper Assets Management**
10. **Insufficient Logging & Monitoring**

**Three Security Pillars**:

**1. Authentication** (Who are you?)
- OAuth 2.1 Authorization Code with PKCE
- JWT with RS256 signatures
- API Key with rotation policies

**2. Authorization** (What can you access?)
- Role-based access control (RBAC)
- Attribute-based access control (ABAC)
- Scope-based permission model

**3. Rate Limiting** (How much can you use?)
- Token bucket algorithm
- Sliding window counter
- Distributed rate limiting (Redis)

**Quick Defense Implementation**:
```javascript
// NEVER do this:
app.get('/api/users', (req, res) => {
  // No authentication, no authorization
  res.json(db.users.all()); // Data exposure!
});

// ALWAYS do this:
app.get('/api/users', 
  authenticate(), // Verify JWT/API key
  authorize('read:users'), // Check scope/role
  rateLimit(), // Prevent abuse
  (req, res) => {
    const users = db.users.findByTenant(req.tenantId);
    res.json(users); // Tenant-isolated data
  }
);
```

---

## Modules

### 1. [Authentication & Authorization](modules/authentication-authorization.md)
OAuth 2.1, JWT, API key authentication, and RBAC/ABAC authorization patterns.

**Topics**:
- OAuth 2.1 with PKCE implementation
- JWT RS256 verification with issuer/audience checks
- API key management with rate limiting
- Multi-tenant data isolation
- BOLA prevention

### 2. [Rate Limiting & Protection](modules/rate-limiting-protection.md)
Token bucket rate limiting, distributed Redis-based rate limiting, and DoS protection.

**Topics**:
- Token bucket algorithm (Redis Lua script)
- GraphQL query complexity analysis
- API request throttling
- Distributed rate limiting
- DoS attack prevention

### 3. [Advanced Security Patterns](modules/advanced-security-patterns.md)
gRPC mTLS, webhook security, API versioning, CORS, and enterprise integration.

**Topics**:
- gRPC mTLS server and client configuration
- Webhook HMAC-SHA256 signature verification
- API versioning with deprecation strategy
- CORS configuration with origin whitelist
- Security headers (Helmet)

---

## Deployment Checklist

✅ **Essential Security Controls**:
- [ ] OAuth 2.1 with PKCE implementation
- [ ] JWT RS256 verification with issuer/audience checks
- [ ] API key management with rate limiting
- [ ] Token bucket rate limiting (Redis-based)
- [ ] Multi-tenant data isolation
- [ ] BOLA prevention on all endpoints

✅ **Enterprise Security Integration**:
- [ ] CORS configuration with origin whitelist
- [ ] Security headers (Helmet)
- [ ] API versioning with deprecation strategy
- [ ] Webhook signature verification
- [ ] GraphQL query complexity limits
- [ ] gRPC mTLS configuration

✅ **Monitoring & Compliance**:
- [ ] Security event logging
- [ ] Rate limit monitoring
- [ ] API usage analytics
- [ ] Token revocation tracking
- [ ] Audit trail for sensitive operations

---

## Version History

**v4.0.0** (2025-11-13)
- ✨ Modularized structure (3 focused modules)
- ✨ Enhanced OAuth 2.1 with PKCE patterns
- ✨ Comprehensive multi-tenant security
- ✨ Production-ready implementation examples

**v3.0.0** (2025-11-12)
- ✨ Advanced GraphQL security patterns
- ✨ gRPC mTLS implementation
- ✨ Webhook security with HMAC

---

**Generated with**: MoAI-ADK Skill Factory    
**Last Updated**: 2025-11-13  
**Security Classification**: Enterprise API Security  
**Optimization**: Modularized for progressive loading
