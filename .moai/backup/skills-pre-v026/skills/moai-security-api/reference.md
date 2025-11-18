# moai-security-api: Reference & Official Documentation

## Official Standards & Specifications

### OAuth 2.1 & OpenID Connect
- OAuth 2.1 Specification: https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1
- OAuth 2.0 PKCE (RFC 7636): https://tools.ietf.org/html/rfc7636
- OpenID Connect Core: https://openid.net/specs/openid-connect-core-1_0.html
- OAuth 2.0 Scope Recommendations: https://tools.ietf.org/html/draft-ietf-oauth-v2-scopes

### JWT & Token Management
- JWT.io Registry: https://jwt.io/
- RFC 7519 (JWT Claims): https://tools.ietf.org/html/rfc7519
- RFC 7797 (Unencrypted JWT): https://tools.ietf.org/html/rfc7797
- JWE (JSON Web Encryption): https://tools.ietf.org/html/rfc7516

### API Security Standards
- OWASP API Security Top 10: https://owasp.org/www-project-api-security/
- REST API Best Practices: https://restfulapi.net/
- GraphQL Security Guidelines: https://cheatsheetseries.owasp.org/cheatsheets/GraphQL_Cheat_Sheet.html
- gRPC Security: https://grpc.io/docs/guides/auth/

### HTTP Security Headers
- Content-Security-Policy (CSP): https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
- CORS Specification: https://fetch.spec.whatwg.org/#http-cors-protocol
- Strict-Transport-Security (HSTS): https://tools.ietf.org/html/rfc6797
- X-Content-Type-Options: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options

## Framework Documentation

### Express.js
- Official Security Best Practices: https://expressjs.com/en/advanced/best-practice-security.html
- Passport.js Integration: http://www.passportjs.org/
- Express Middleware Reference: https://expressjs.com/en/resources/middleware.html
- CORS Package: https://github.com/expressjs/cors
- Helmet.js Security Headers: https://helmetjs.github.io/

### Apollo Server (GraphQL)
- Apollo Server Documentation: https://www.apollographql.com/docs/apollo-server/
- GraphQL Security Best Practices: https://www.apollographql.com/docs/apollo-server/security/authentication/
- Query Complexity Analysis: https://www.apollographql.com/docs/apollo-server/security/complexity-analysis/
- Rate Limiting: https://www.apollographql.com/docs/apollo-server/security/rate-limiting/

### Node.js gRPC
- gRPC Official Documentation: https://grpc.io/docs/languages/node/quickstart/
- gRPC Security: https://grpc.io/docs/guides/auth/
- mTLS Configuration: https://grpc.io/docs/guides/auth/#mutual-tls
- Protocol Buffers: https://developers.google.com/protocol-buffers

### Rate Limiting & Redis
- Redis Official: https://redis.io/
- Redis Lua Scripting: https://redis.io/commands/eval/
- Token Bucket Algorithm: https://en.wikipedia.org/wiki/Token_bucket
- Distributed Rate Limiting Patterns: https://stripe.com/blog/rate-limiters

## Libraries & Tools (November 2025)

### Authentication & Authorization
- **jsonwebtoken** (9.0.x): https://github.com/auth0/node-jsonwebtoken
- **passport** (0.7.x): https://github.com/jaredhanson/passport
- **passport-oauth2** (1.8.x): https://github.com/jaredhanson/passport-oauth2
- **@passport-js/passport-google-oauth20**: https://github.com/jaredhanson/passport-google-oauth2
- **@passport-js/passport-github2**: https://github.com/cfsghost/passport-github

### API Frameworks
- **Express.js** (4.21.x): https://github.com/expressjs/express
- **Fastify** (5.0.x): https://github.com/fastify/fastify
- **@apollo/server** (4.12.x): https://github.com/apollographql/apollo-server
- **@grpc/grpc-js** (1.12.x): https://github.com/grpc/grpc-node

### Security Headers & Validation
- **helmet** (7.0.x): https://github.com/helmetjs/helmet
- **express-validator** (7.0.x): https://github.com/express-validator/express-validator
- **cors** (2.8.x): https://github.com/expressjs/cors
- **csurf** (1.11.x): https://github.com/expressjs/csurf

### Rate Limiting & Caching
- **redis** (5.0.x): https://github.com/redis/node-redis
- **express-rate-limit** (7.2.x): https://github.com/nfriedly/express-rate-limit
- **ioredis** (5.3.x): https://github.com/luin/ioredis

## Common Vulnerabilities & CWE References

### Broken Object Level Authorization (BOLA/IDOR)
- CWE-639: https://cwe.mitre.org/data/definitions/639.html
- OWASP A01:2021: https://owasp.org/Top10/A01_2021-Broken_Access_Control/
- Mitigation: Always verify tenant_id/user_id on every query

### Broken Function Level Authorization (BFLA)
- CWE-269: https://cwe.mitre.org/data/definitions/269.html
- OWASP A05:2021: https://owasp.org/Top10/A05_2021-Broken_Access_Control/
- Mitigation: Check user.role before executing admin operations

### Excessive Data Exposure
- CWE-200: https://cwe.mitre.org/data/definitions/200.html
- OWASP A03:2021: https://owasp.org/Top10/A03_2021-Injection/
- Mitigation: Return only necessary fields in API responses

### Rate Limit Bypass
- CWE-770: https://cwe.mitre.org/data/definitions/770.html
- OWASP A04:2023: https://owasp.org/www-project-api-security/
- Mitigation: Use distributed rate limiting with Redis

### Broken API Versioning
- No standard CWE, but related to obsolescence
- OWASP API Security Top 10: https://owasp.org/www-project-api-security/
- Mitigation: Use deprecation headers, version endpoints explicitly

## Testing & Validation Tools

### API Testing
- Postman: https://www.postman.com/
- Insomnia: https://insomnia.rest/
- Thunder Client: https://www.thunderclient.com/
- REST Client (VS Code): https://marketplace.visualstudio.com/items?itemName=humao.rest-client

### Security Testing
- OWASP ZAP: https://www.zaproxy.org/
- Burp Suite: https://portswigger.net/burp
- SQLmap: https://sqlmap.org/
- API Fortress: https://www.apitesting.com/

### Load Testing
- Apache JMeter: https://jmeter.apache.org/
- Locust: https://locust.io/
- k6: https://k6.io/
- Artillery: https://artillery.io/

## Industry Standards & Best Practices

### NIST Cybersecurity Framework
- NIST CSF Overview: https://www.nist.gov/cyberframework
- NIST SP 800-53 (Security Controls): https://csrc.nist.gov/publications/detail/sp/800-53/rev-5/final
- NIST SP 800-63B (Authentication): https://pages.nist.gov/800-63-3/sp800-63b.html

### CIS Controls
- CIS Controls v8: https://www.cisecurity.org/cis-controls/v8
- API Security: https://www.cisecurity.org/blog/securing-apis/

### PCI-DSS Compliance
- PCI-DSS v4.0: https://www.pcisecuritystandards.org/
- Secure Coding Practices: https://www.pcisecuritystandards.org/documents/PCI_Secure_Coding_Practices.pdf

## Conference Talks & Articles

### OWASP Events
- AppSec Global: https://www.globalappsec.org/
- OWASP Europe Conference: https://owasp.org/www-community/events

### API Security Articles
- "The 2023 State of API Security": https://www.neuralegion.com/blog/api-security-report/
- "API Security Best Practices": https://owasp.org/www-project-api-security/
- "Protecting APIs from OWASP Top 10": https://blog.postman.com/

## Related Skills

- **moai-security-auth**: Authentication patterns (OAuth 2.1, JWT, MFA)
- **moai-security-encryption**: Encryption & TLS 1.3
- **moai-security-owasp**: OWASP Top 10 defense
- **moai-domain-web-api**: REST API best practices
- **moai-domain-database**: SQL injection prevention

