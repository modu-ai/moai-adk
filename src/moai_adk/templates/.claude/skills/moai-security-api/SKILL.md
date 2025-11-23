---
name: moai-security-api
description: Comprehensive API security for REST, GraphQL, and gRPC services
version: 1.0.1
modularized: true
tags:
  - security
  - enterprise
  - threat-modeling
  - api
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**Name**: moai-security-api
**Domain**: API Security & Authentication
**Freedom Level**: high
**Target Users**: Backend developers, API architects, security engineers
**Invocation**: Skill("moai-security-api")
**Progressive Disclosure**: SKILL.md (core) → modules/ (detailed patterns)
**Last Updated**: 2025-11-23
**Modularized**: true

---

## 🎯 Quick Reference (30 seconds)

**Purpose**: Comprehensive security for REST, GraphQL, and gRPC APIs.

**OWASP API Top 10 (2023) Coverage**:
1. Broken Object Level Authorization (BOLA)
2. Broken Authentication
3. Excessive Data Exposure
4. Lack of Resources & Rate Limiting
5. Broken Function Level Authorization
6. Mass Assignment
7. Cross-Site Scripting (XSS)
8. Broken API Versioning
9. Improper Assets Management
10. Insufficient Logging & Monitoring

---

## 📚 Core Patterns (5-10 minutes)

### Pattern 1: Authentication & Authorization

**Key Concept**: Proper user identity verification and access control

**Three Security Pillars**:
1. **Authentication** - Who are you? (OAuth 2.1, JWT, API Keys)
2. **Authorization** - What can you access? (RBAC, ABAC, Scopes)
3. **Rate Limiting** - How much can you use? (Token bucket, Sliding window)

**Best Practice**:
```javascript
// ✅ CORRECT - Auth + Authz + Rate limiting
app.get('/api/users/:id',
  authenticate(),
  authorize(['users:read']),
  rateLimit(100, '1h'),
  (req, res) => {
    // Access user if authorized
  }
);

// ❌ WRONG - No security at all
app.get('/api/users/:id', (req, res) => {
  res.json(db.users.all()); // Data exposure!
});
```

### Pattern 2: BOLA Prevention (Broken Object Level Authorization)

**Key Concept**: Verify user owns resource before returning it

**Vulnerable Code**:
```python
# ❌ WRONG - No authorization check
@app.get('/api/users/{user_id}')
def get_user(user_id):
    return db.users.find_by_id(user_id)  # Returns ANY user!
```

**Secure Code**:
```python
# ✅ CORRECT - Check authorization
@app.get('/api/users/{user_id}')
def get_user(user_id, current_user):
    user = db.users.find_by_id(user_id)
    if user.id != current_user.id and not current_user.is_admin:
        raise AuthorizationError("Forbidden")
    return user
```

### Pattern 3: Rate Limiting Implementation

**Key Concept**: Prevent abuse through resource limits

**Algorithm**: Token bucket
```
1. Bucket contains N tokens
2. Each request consumes 1 token
3. Bucket refills at rate R tokens/second
4. Request denied if bucket empty
```

**Implementation**:
```python
from redis import Redis
redis = Redis()

def rate_limit(user_id, limit=100, window=3600):
    key = f"rate_limit:{user_id}"
    current = redis.incr(key)
    if current == 1:
        redis.expire(key, window)
    if current > limit:
        raise RateLimitError("Exceeded limit")
```

### Pattern 4: Input Validation & Injection Prevention

**Key Concept**: Sanitize all user input

**Common Injections**:
- SQL Injection
- Command Injection
- XML External Entity (XXE)
- GraphQL Injection

**Defense**:
```python
# ✅ Use parameterized queries
db.query("SELECT * FROM users WHERE id = ?", [user_id])

# ❌ NEVER concatenate
db.query(f"SELECT * FROM users WHERE id = {user_id}")
```

### Pattern 5: Logging & Monitoring

**Key Concept**: Track and audit all API access

**Critical Events to Log**:
- Authentication attempts (success/failure)
- Authorization decisions (allow/deny)
- Data access (who accessed what)
- Errors and exceptions
- Rate limit violations

```python
# ✅ Comprehensive logging
def handle_request(req):
    logger.info(f"User {req.user.id} accessed {req.path}")
    try:
        # Process request
    except Exception as e:
        logger.error(f"Error for {req.user.id}: {e}")
        raise
```

---

## 📖 Advanced Documentation

This Skill uses Progressive Disclosure. For detailed patterns:

- **[modules/oauth-jwt.md](modules/oauth-jwt.md)** - OAuth 2.1 & JWT implementation
- **[modules/graphql-security.md](modules/graphql-security.md)** - GraphQL-specific security
- **[modules/rate-limiting.md](modules/rate-limiting.md)** - Rate limiting strategies
- **[modules/reference.md](modules/reference.md)** - OWASP checklist & API security

---

## 🔒 Security Checklist

- [ ] Authentication enforced on all endpoints
- [ ] Authorization checked per user + resource
- [ ] Rate limiting implemented
- [ ] Input validation on all user input
- [ ] Output encoding (prevent XSS)
- [ ] CORS properly configured
- [ ] API versioning strategy defined
- [ ] Logging & monitoring in place
- [ ] Error messages don't expose sensitive data
- [ ] Secrets not hardcoded

---

## 🔗 Integration with Other Skills

**Complementary Skills**:
- Skill("moai-security-identity") - Identity management
- Skill("moai-domain-backend") - Backend architecture
- Skill("moai-domain-monitoring") - Security monitoring

---

## 📈 Version History

**1.0.1** (2025-11-23)
- 🔄 Refactored with Progressive Disclosure
- ✨ Added OWASP Top 10 mapping
- ✨ Core patterns highlighted

**1.0.0** (2025-11-12)
- ✨ REST, GraphQL, gRPC security patterns
- ✨ OAuth 2.1 & JWT implementation

---

**Maintained by**: alfred
**Domain**: API Security
**Generated with**: MoAI-ADK Skill Factory
