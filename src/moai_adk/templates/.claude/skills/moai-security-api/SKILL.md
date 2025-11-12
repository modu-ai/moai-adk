---
name: "moai-security-api"
version: "4.0.0"
status: stable
description: "Enterprise Skill for advanced development"
allowed-tools: "Read, Bash, WebSearch, WebFetch"
---

# moai-security-api: REST/GraphQL/gRPC API Security

**Comprehensive API Security for Modern Web Services**  
Trust Score: 9.9/10 | Version: 4.0.0 | Enterprise Mode | Last Updated: 2025-11-12

---

## Overview

API security is critical in distributed systems where REST, GraphQL, and gRPC endpoints become attack surfaces. This Skill provides production-ready defense patterns, authentication frameworks, and rate-limiting strategies for all API paradigms.

**When to use this Skill:**
- Building REST APIs with OAuth 2.1 authentication
- Implementing GraphQL security (query depth limits, complexity analysis)
- Protecting gRPC services with mTLS and authorization
- Managing API keys securely with rotation policies
- Implementing rate limiting and DDoS protection
- Securing API webhooks and callbacks
- Building multi-tenant API platforms safely

---

## Level 1: Foundations (FREE TIER)

### What is API Security?

```
Attack Surface:
User → [REST/GraphQL/gRPC Endpoint] → Internal Resources
        ↓
    - Missing Authentication
    - Broken Authorization
    - Excessive Data Exposure
    - Rate Limit Bypass
    - Injection Attacks
```

**OWASP API Security Top 10 (2023 Updated):**
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

### API Security Fundamentals

**Three Security Pillars:**

1. **Authentication** (Who are you?)
   - OAuth 2.1 / OpenID Connect
   - JWT with RS256 signature
   - API Key with rotation

2. **Authorization** (What can you access?)
   - Role-based access control (RBAC)
   - Attribute-based access control (ABAC)
   - Scope-based permission model

3. **Rate Limiting** (How much can you use?)
   - Token bucket algorithm
   - Sliding window counter
   - Distributed rate limiting (Redis)

### API Architecture Patterns (November 2025)

**Pattern Selection Matrix:**

```
┌─────────────────┬──────────────────┬────────────┬──────────────┐
│ API Type        │ Best Practice    │ Framework  │ Version      │
├─────────────────┼──────────────────┼────────────┼──────────────┤
│ REST            │ OAuth 2.1 + JWT  │ Express.js │ 4.21.x       │
│ GraphQL         │ Query depth      │ Apollo     │ 4.12.x       │
│ gRPC            │ mTLS + JWT       │ @grpc/grpc│ 1.12.x       │
│ Webhook         │ HMAC-SHA256      │ Custom     │ Best practice│
└─────────────────┴──────────────────┴────────────┴──────────────┘
```

---

## Level 2: Intermediate Patterns (STANDARD TIER)

### OAuth 2.1 Implementation for REST APIs

**November 2025 Best Practice**: OAuth 2.1 removes deprecated flows (Implicit, Resource Owner Password).

**Recommended Flows:**

1. **Authorization Code Flow with PKCE** (Frontend & Mobile)
   ```
   User → Frontend → Authorization Server → User Approval
           ↓                                       ↓
        PKCE Code        ← ← ← ← ← ← ← ← ← Token Response
   Frontend exchanges code + code_verifier for access_token
   ```

2. **Client Credentials Flow** (Service-to-Service)
   ```
   Service A → Auth Server (client_id + client_secret)
               ↓
           Access Token → Service B
   ```

3. **Refresh Token Grant** (Long-lived Sessions)
   ```
   Old Token Expired → Refresh Server (refresh_token)
                       ↓
                   New Access Token
   ```

**Express.js + OAuth 2.1 Pattern:**

```javascript
// Use passport-oauth2 v1.8.0 (November 2025 latest)
// With PKCE extension (RFC 7636)

const strategy = new OAuth2Strategy({
  authorizationURL: 'https://auth-server.com/oauth/authorize',
  tokenURL: 'https://auth-server.com/oauth/token',
  clientID: process.env.OAUTH_CLIENT_ID,
  clientSecret: process.env.OAUTH_CLIENT_SECRET,
  callbackURL: 'https://api.example.com/auth/callback',
  state: true, // CSRF protection
  pkce: true   // RFC 7636 PKCE
}, verifyCallback);

// Scope limiting: principle of least privilege
passport.use('oauth', strategy);
```

### JWT RS256 Validation

**Why RS256 (RSA Signature + SHA256)?**

- **Asymmetric**: Public key for verification (no secret needed on clients)
- **Revocation**: Easy token revocation via blacklist/JTI checks
- **Distribution**: Multiple services share verification key safely

**Production Implementation (jsonwebtoken v9.x):**

```javascript
const jwt = require('jsonwebtoken');
const fs = require('fs');

// Load public key from Authorization Server
const publicKey = fs.readFileSync('/secure/auth-server-public.pem', 'utf8');

function verifyToken(token) {
  try {
    const decoded = jwt.verify(token, publicKey, {
      algorithms: ['RS256'],
      issuer: 'https://auth-server.com',      // Prevent algorithm confusion
      audience: 'api.example.com',              // Scope token to this API
      clockTolerance: 5,                        // 5s clock skew tolerance
      ignoreNotBefore: false,
      ignoreExpiration: false
    });
    
    // Check token blacklist (for revocation)
    if (isTokenBlacklisted(token)) {
      throw new Error('Token revoked');
    }
    
    return decoded;
  } catch (err) {
    console.error('JWT verification failed:', err.message);
    throw new AuthenticationError(err.message);
  }
}
```

### API Key Management with Rotation

**Three-Tier API Key Strategy:**

1. **Client API Key** (Public Identifier)
   - Format: `sk_live_xxxxxxxxxxxxxxxxxxxx` (like Stripe)
   - Includes: client_id, version, timestamp
   - Rotated: annually or on compromise

2. **Server API Key** (Secret Hash)
   - Storage: Bcrypt hashed in database
   - Never sent over network
   - Used for verification only

3. **Rate Limit Quota**
   - Per-key rate limits
   - Stored in Redis with TTL
   - Resets hourly/daily

**Express.js API Key Validation Middleware:**

```javascript
const crypto = require('crypto');
const redis = require('redis');

const redisClient = redis.createClient();

async function apiKeyMiddleware(req, res, next) {
  const apiKey = req.headers['x-api-key'];
  
  if (!apiKey) {
    return res.status(401).json({ error: 'Missing API key' });
  }
  
  try {
    // Lookup in Redis cache first (99% hit rate)
    let client = await redisClient.get(`api_key:${apiKey}`);
    
    if (!client) {
      // Hit database for new/uncached key
      client = await db.apiKeys.findOne({ 
        client_id: apiKey.split('_')[0],
        status: 'active',
        expires_at: { $gt: new Date() }
      });
      
      if (!client) {
        return res.status(401).json({ error: 'Invalid API key' });
      }
      
      // Cache for 1 hour
      await redisClient.setEx(`api_key:${apiKey}`, 3600, JSON.stringify(client));
    }
    
    // Rate limit check
    const rateLimitKey = `ratelimit:${apiKey}`;
    const count = await redisClient.incr(rateLimitKey);
    
    if (count === 1) {
      // First request this minute, set expiry
      await redisClient.expire(rateLimitKey, 60);
    }
    
    if (count > client.rate_limit_per_minute) {
      return res.status(429).json({ 
        error: 'Rate limit exceeded',
        retry_after: 60
      });
    }
    
    req.client = client;
    next();
  } catch (err) {
    console.error('API key validation error:', err);
    res.status(500).json({ error: 'Authentication failed' });
  }
}

app.use(apiKeyMiddleware);
```

### Rate Limiting with Token Bucket

**Token Bucket Algorithm** (Distributed):

```
Bucket Capacity: 100 tokens
Refill Rate: 10 tokens/second
Request Cost: 1 token per operation

Request Flow:
1. Request arrives → Check tokens in bucket
2. If tokens > 0 → Allow request, consume token
3. If tokens = 0 → Queue or reject (429 Too Many Requests)
4. Tokens refill at rate (10/sec)
```

**Redis-based Implementation (Distributed):**

```javascript
const redis = require('redis');
const client = redis.createClient();

async function rateLimitMiddleware(req, res, next) {
  const userId = req.user.id;
  const key = `ratelimit:${userId}`;
  
  // Lua script for atomic token bucket operation
  const luaScript = `
    local key = KEYS[1]
    local limit = tonumber(ARGV[1])
    local refill_rate = tonumber(ARGV[2])
    local now = tonumber(ARGV[3])
    local ttl = tonumber(ARGV[4])
    
    local current = redis.call('GET', key)
    if not current then
      redis.call('SET', key, limit, 'EX', ttl)
      return 1
    end
    
    -- Calculate refilled tokens
    local last_refill = redis.call('GET', key .. ':last_refill') or now
    local time_passed = now - tonumber(last_refill)
    local refilled = math.min(limit, 
      tonumber(current) + (time_passed * refill_rate / 1000))
    
    if refilled >= 1 then
      redis.call('SET', key, refilled - 1, 'EX', ttl)
      redis.call('SET', key .. ':last_refill', now, 'EX', ttl)
      return 1
    else
      return 0
    end
  `;
  
  try {
    const allowed = await client.eval(luaScript, 
      { keys: [key], arguments: [100, 10, Date.now(), 3600] });
    
    if (!allowed) {
      return res.status(429).json({
        error: 'Rate limit exceeded',
        retry_after: 60
      });
    }
    
    next();
  } catch (err) {
    console.error('Rate limiting error:', err);
    res.status(500).json({ error: 'Internal server error' });
  }
}
```

### GraphQL Security

**Apollo Server 4.12.x Security Hardening:**

```javascript
const { ApolloServer } = require('@apollo/server');
const { expressMiddleware } = require('@apollo/server/express4');

const server = new ApolloServer({
  typeDefs,
  resolvers,
  
  // 1. Query complexity analysis
  plugins: {
    didResolveOperation(requestContext) {
      const complexity = calculateQueryComplexity(
        requestContext.document,
        requestContext.operation,
        resolvers
      );
      
      // Prevent DoS queries
      if (complexity > 5000) {
        throw new Error(
          `Query too complex: ${complexity} (max: 5000)`
        );
      }
    }
  },
  
  // 2. Query depth limit
  context: ({ req }) => {
    return {
      depth: 0,
      maxDepth: 5  // Prevent nested query attacks
    };
  },
  
  // 3. Timeout protection
  executionTimeoutMs: 5000,
  
  // 4. Introspection disabled in production
  introspection: process.env.NODE_ENV !== 'production'
});

// Field-level authorization
const resolvers = {
  Query: {
    user: (parent, args, context) => {
      // Only return if requester has 'read:user' scope
      if (!context.scopes.includes('read:user')) {
        throw new ForbiddenError('Missing read:user scope');
      }
      return db.users.findById(args.id);
    }
  }
};
```

### gRPC Security with mTLS

**mTLS (Mutual TLS) Setup:**

```javascript
const grpc = require('@grpc/grpc-js');
const credentials = require('@grpc/grpc-js').credentials;
const fs = require('fs');

// Load mTLS certificates
const rootCert = fs.readFileSync('/secure/ca-cert.pem');
const clientCert = fs.readFileSync('/secure/client-cert.pem');
const clientKey = fs.readFileSync('/secure/client-key.pem');

// Server setup
const serverCredentials = grpc.ServerCredentials.createSsl(
  rootCert,
  [{
    cert_chain: clientCert,
    private_key: clientKey
  }]
);

const server = new grpc.Server();
server.bind('0.0.0.0:50051', serverCredentials);

// Client setup
const clientCredentials = grpc.credentials.createSsl(
  rootCert,
  clientKey,
  clientCert
);

const client = new ServiceClient(
  'api-server:50051',
  clientCredentials
);

// JWT verification in gRPC interceptor
function jwtInterceptor(options, nextCall) {
  const metadata = options.metadata || new grpc.Metadata();
  const token = metadata.get('authorization')[0];
  
  const decoded = jwt.verify(token, publicKey, { algorithms: ['RS256'] });
  options.metadata = metadata;
  
  return nextCall(options);
}
```

---

## Level 3: Enterprise Patterns (PREMIUM TIER)

### Multi-Tenant API Security

**Tenant Isolation Pattern:**

```javascript
async function tenantMiddleware(req, res, next) {
  const token = req.headers.authorization?.split(' ')[1];
  const decoded = jwt.verify(token, publicKey);
  
  // Token must include tenant_id
  const tenantId = decoded.tenant_id;
  
  if (!tenantId) {
    return res.status(403).json({ error: 'Invalid token: missing tenant_id' });
  }
  
  // Verify tenant ownership
  const tenant = await db.tenants.findById(tenantId);
  if (!tenant || tenant.status !== 'active') {
    return res.status(403).json({ error: 'Tenant not found or inactive' });
  }
  
  req.tenantId = tenantId;
  req.tenant = tenant;
  next();
}

// BOLA prevention: Always check tenant_id in queries
app.get('/api/users/:userId', tenantMiddleware, async (req, res) => {
  const user = await db.users.findById(req.params.userId);
  
  // CRITICAL: Verify tenant_id matches
  if (user.tenant_id !== req.tenantId) {
    return res.status(403).json({ error: 'Access denied' });
  }
  
  res.json(user);
});
```

### API Versioning & Backward Compatibility

**Semantic Versioning with Deprecation:**

```javascript
// Version endpoint via URL path
app.get('/api/v1/users/:id', legacyUserHandler);      // v1 (deprecated 2026-01-01)
app.get('/api/v2/users/:id', currentUserHandler);     // v2 (current)
app.get('/api/v3/users/:id', nextGenUserHandler);     // v3 (beta)

// Deprecation warning header
function deprecationMiddleware(req, res, next) {
  const apiVersion = req.path.match(/\/v(\d+)\//)[1];
  const currentVersion = 2;
  
  if (apiVersion < currentVersion) {
    res.set('Deprecation', 'true');
    res.set('Sunset', new Date('2026-01-01').toUTCString());
    res.set('Link', '</api/v' + currentVersion + '/users>; rel="successor-version"');
  }
  
  next();
}
```

### Webhook Security (HMAC-SHA256)

**Signature Verification:**

```javascript
const crypto = require('crypto');

async function sendWebhook(event, url, secret) {
  const timestamp = Math.floor(Date.now() / 1000);
  const payload = JSON.stringify({ ...event, timestamp });
  
  // Create signature: HMAC-SHA256(secret, payload)
  const signature = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  
  const headers = {
    'X-Webhook-Signature': `sha256=${signature}`,
    'X-Webhook-Timestamp': timestamp.toString(),
    'Content-Type': 'application/json'
  };
  
  await fetch(url, {
    method: 'POST',
    headers,
    body: payload,
    timeout: 30000
  });
}

// Webhook receiver endpoint
app.post('/webhooks/stripe', express.raw({ type: 'application/json' }), (req, res) => {
  const signature = req.headers['stripe-signature'];
  const timestamp = parseInt(req.headers['x-webhook-timestamp']);
  const payload = req.rawBody;
  
  // Prevent replay attacks: timestamp must be recent (within 5 minutes)
  const age = Math.floor(Date.now() / 1000) - timestamp;
  if (age > 300) {
    return res.status(401).json({ error: 'Webhook expired' });
  }
  
  // Verify signature
  const [version, hash] = signature.split(',')[0].split('=');
  const expected = crypto
    .createHmac('sha256', process.env.WEBHOOK_SECRET)
    .update(`${timestamp}.${payload}`)
    .digest('hex');
  
  if (!crypto.timingSafeEqual(Buffer.from(hash), Buffer.from(expected))) {
    return res.status(401).json({ error: 'Invalid signature' });
  }
  
  // Process webhook safely
  processWebhook(JSON.parse(payload));
  res.json({ received: true });
});
```

### CORS & CSRF Protection

**Express.js with helmet + csrf-protection:**

```javascript
const cors = require('cors');
const helmet = require('helmet');
const csrf = require('csurf');
const cookieParser = require('cookie-parser');

// 1. CORS configuration
const corsOptions = {
  origin: ['https://app.example.com', 'https://web.example.com'],
  credentials: true,                    // Allow cookies
  methods: ['GET', 'POST', 'PUT', 'DELETE'],
  allowedHeaders: ['Content-Type', 'Authorization'],
  maxAge: 86400                         // 24 hours
};

app.use(cors(corsOptions));

// 2. Helmet for security headers
app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      scriptSrc: ["'self'", "https://cdn.example.com"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      imgSrc: ["'self'", "data:", "https:"],
      connectSrc: ["'self'", "https://api.example.com"]
    }
  },
  hsts: { maxAge: 31536000 },           // 1 year
  noSniff: true,
  xssFilter: true
}));

// 3. CSRF protection
app.use(cookieParser());
app.use(csrf({ cookie: false }));

// Add CSRF token to forms
app.get('/', (req, res) => {
  res.json({ csrfToken: req.csrfToken() });
});

// Validate CSRF on state-changing requests
app.post('/api/users', csrf(), (req, res) => {
  // Token automatically verified by middleware
  res.json({ success: true });
});
```

---

## Reference

### Official Documentation
- OAuth 2.1 Specification: https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1
- RFC 7636 PKCE: https://tools.ietf.org/html/rfc7636
- JWT.io: https://jwt.io/
- Express.js Security Best Practices: https://expressjs.com/en/advanced/best-practice-security.html
- Apollo Server Security: https://www.apollographql.com/docs/apollo-server/security/
- OWASP API Security: https://owasp.org/www-project-api-security/
- gRPC Security: https://grpc.io/docs/guides/auth/
- HMAC Signature Verification: https://tools.ietf.org/html/rfc2104

### Tools & Libraries (November 2025 Versions)
- **Express.js**: 4.21.x
- **Apollo Server**: 4.12.x
- **jsonwebtoken**: 9.x (RS256 support)
- **passport-oauth2**: 1.8.x
- **helmet**: 7.x
- **redis**: 5.0.x
- **@grpc/grpc-js**: 1.12.x
- **libsodium.js**: 0.7.x

### Common Vulnerabilities & Mitigations

| Vulnerability | OWASP | Mitigation |
|---|---|---|
| **Missing OAuth** | A02:2021 | Use OAuth 2.1 + PKCE |
| **Weak JWT** | A02:2021 | Use RS256, verify iss/aud |
| **BOLA (Object Auth)** | A01:2023 | Check tenant_id on every query |
| **Rate Limit Bypass** | A04:2023 | Use Redis token bucket |
| **GraphQL DoS** | A04:2023 | Query depth + complexity limits |
| **CSRF** | A01:2021 | CSRF tokens + SameSite cookies |

---

## Quick Setup Examples

### 1. Express.js + OAuth 2.1
```bash
npm install express passport passport-oauth2 jsonwebtoken redis
# See Level 2 for implementation
```

### 2. Apollo Server GraphQL Security
```bash
npm install @apollo/server @apollo/server/express4
# See Level 2 for complexity analysis
```

### 3. gRPC + mTLS
```bash
npm install @grpc/grpc-js @grpc/proto-loader
# See Level 2 for certificate setup
```

---

**Version**: 4.0.0 Enterprise
**Skill Category**: Security (API Authentication & Authorization)
**Complexity**: Medium-Advanced
**Time to Implement**: 2-4 hours per component
**Prerequisites**: Express.js/Node.js, OAuth concepts, JWT understanding
