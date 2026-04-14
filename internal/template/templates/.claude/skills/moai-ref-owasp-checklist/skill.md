---
name: moai-ref-owasp-checklist
description: >
  OWASP Top 10 security checklist, authentication patterns, input validation,
  and HTTP security headers reference. Agent-extending skill that amplifies
  expert-security and expert-backend expertise with production-grade security patterns.
  NOT for: frontend UI, DevOps deployment, performance optimization, testing strategy.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-03-30"
  tags: "owasp, security, checklist, authentication, validation, reference"
  agent: "expert-security"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000

# MoAI Extension: Triggers
triggers:
  keywords: ["owasp", "security", "vulnerability", "injection", "xss", "csrf", "auth"]
  agents: ["expert-security", "expert-backend"]
  phases: ["run"]
---

# OWASP Security Checklist Reference

## Target Agents

- `expert-security` - Primary: applies checklist during security audits
- `expert-backend` - Secondary: applies during API implementation

## OWASP API Security Top 10

| Rank | Vulnerability | Check | Defense |
|------|-------------|-------|---------|
| A1 | **BOLA** (Broken Object Level Authorization) | Can user A access user B's resources? | Verify object ownership at every endpoint |
| A2 | **Broken Authentication** | Weak passwords, unlimited login attempts? | bcrypt (cost 12+), rate limit, MFA |
| A3 | **Broken Object Property Level Authorization** | Are hidden fields exposed in responses? | Response DTOs, field-level filtering |
| A4 | **Unrestricted Resource Consumption** | Can mass requests crash the server? | Rate limiting, enforce pagination limits |
| A5 | **Broken Function Level Authorization** | Can regular users call admin APIs? | RBAC middleware, permission checks |
| A6 | **SSRF** (Server-Side Request Forgery) | Can URL input access internal resources? | URL whitelist, block internal IPs |
| A7 | **Security Misconfiguration** | Debug mode, default accounts exposed? | Separate prod config, inspect headers |
| A8 | **Lack of Automated Threat Protection** | Can APIs be called in abnormal sequences? | State machine validation, business rules |
| A9 | **Improper Asset Management** | Unused APIs, old versions exposed? | API inventory, version deprecation |
| A10 | **Unsafe API Consumption** | Are external API responses trusted blindly? | Validate external responses, set timeouts |

## Authentication Checklist

### Password Policy
- Minimum 8 characters, show strength meter (not strict rules)
- bcrypt (cost factor 12+) or Argon2id
- Temporary lock after 5 failed attempts (15 min) or CAPTCHA
- Prevent reuse of last 5 passwords

### JWT Configuration
| Setting | Recommended Value |
|---------|------------------|
| Access Token Expiry | 15-30 minutes |
| Refresh Token Expiry | 7-14 days |
| Algorithm | RS256 (asymmetric) or HS256 |
| Storage | httpOnly + secure + sameSite cookie |
| Payload | Minimal: userId, role only (no PII) |
| Renewal | Silent refresh or token rotation |

### Session Security
- Regenerate session ID after login
- Invalidate session on logout (server-side)
- Set session timeout (30 min idle)
- Bind session to IP/User-Agent (optional, strict)

## HTTP Security Headers

| Header | Value | Purpose |
|--------|-------|---------|
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Force HTTPS |
| `X-Content-Type-Options` | `nosniff` | Prevent MIME sniffing |
| `X-Frame-Options` | `DENY` or `SAMEORIGIN` | Prevent clickjacking |
| `Content-Security-Policy` | `default-src 'self'` | Prevent XSS |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Limit referrer |
| `Permissions-Policy` | `camera=(), microphone=()` | Restrict browser features |

## Input Validation Checklist

| Type | Method | Tool |
|------|--------|------|
| Schema validation | Type + structure check | Zod, Joi, pydantic, Go validator |
| Length limits | Min/max constraints | Schema definitions |
| SQL Injection | Parameterized queries | ORM (Prisma, GORM, SQLAlchemy) |
| XSS Prevention | HTML escaping | DOMPurify (client), server escape |
| Path Traversal | Path normalization | filepath.Clean + whitelist |
| File Upload | Type + size validation | MIME type + magic number check |
| CORS | Origin whitelist | Never `origin: '*'` with credentials |

## Sensitive Data Handling

| Data Type | Storage | Transmission | Logging |
|----------|---------|-------------|---------|
| Passwords | bcrypt hash only | HTTPS only | NEVER |
| API Keys | Environment variables | Header (Authorization) | Masked (first 4 chars) |
| PII | Encrypted (AES-256) | HTTPS only | Masked |
| Credit Cards | Tokenized (payment provider) | Provider SDK | NEVER |
| Sessions | httpOnly cookie | HTTPS only | NEVER |

## Security Review Severity Levels

| Level | Label | Action | Example |
|-------|-------|--------|---------|
| P0 | CRITICAL | Block release | SQL injection, auth bypass |
| P1 | HIGH | Fix before merge | Missing authorization check |
| P2 | MEDIUM | Fix within sprint | Weak password policy |
| P3 | LOW | Track in backlog | Missing security header |

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "This is an internal application, OWASP does not apply" | Internal applications are reachable from compromised internal services. OWASP applies to all web applications. |
| "The framework handles XSS protection" | Frameworks protect default rendering paths. Dynamic HTML insertion, innerHTML, and template literals bypass the protection. |
| "We do not store sensitive data, so encryption is unnecessary" | Session tokens, API keys, and PII are sensitive data. If the application has users, it has sensitive data. |
| "Security headers are just defense-in-depth, not critical" | Each security header blocks a specific attack class. Missing CSP enables XSS even when output is escaped. |
| "I will do a security review before release" | Late security reviews find issues that are expensive to fix. Secure coding practices prevent them from the start. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- User input rendered in HTML without escaping or sanitization
- SQL query built with string concatenation instead of parameterized queries
- Authentication token stored in localStorage instead of httpOnly cookie
- Missing Content-Security-Policy header on response
- Secrets (API keys, passwords) found in source code or configuration files committed to git

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] OWASP Top 10 checklist reviewed for the change (show which items were evaluated)
- [ ] User input sanitized before rendering in HTML output
- [ ] All database queries use parameterized statements
- [ ] Security headers present (CSP, X-Frame-Options, X-Content-Type-Options)
- [ ] No secrets found in source code (show grep results for common secret patterns)
- [ ] Authentication tokens use httpOnly, Secure, SameSite cookie attributes

<!-- moai:evolvable-end -->
