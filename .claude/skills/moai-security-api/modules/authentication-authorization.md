---
name: moai-security-api__authentication-authorization
description: OAuth 2.1, JWT, API key authentication and RBAC/ABAC authorization patterns
parent_skill: moai-security-api
module_type: core
---

## Authentication & Authorization Patterns

### OAuth 2.1 + JWT Security Framework

```javascript
const jwt = require('jsonwebtoken');
const passport = require('passport');
const { Strategy: OAuth2Strategy } = require('passport-oauth2');
const redis = require('redis');

// Redis client for distributed rate limiting
const redisClient = redis.createClient();

// OAuth 2.1 with PKCE (November 2025 best practice)
const oauthStrategy = new OAuth2Strategy({
  authorizationURL: 'https://auth-server.com/oauth/authorize',
  tokenURL: 'https://auth-server.com/oauth/token',
  clientID: process.env.OAUTH_CLIENT_ID,
  clientSecret: process.env.OAUTH_CLIENT_SECRET,
  callbackURL: 'https://api.example.com/auth/callback',
  state: true, // CSRF protection
  pkce: true   // RFC 7636 PKCE
}, verifyCallback);

passport.use('oauth', oauthStrategy);

// JWT RS256 Verification
function verifyJWT(token) {
  try {
    const decoded = jwt.verify(token, getPublicKey(), {
      algorithms: ['RS256'],
      issuer: 'https://auth-server.com',
      audience: 'api.example.com',
      clockTolerance: 5
    });
    
    // Check blacklist for revoked tokens
    if (isTokenBlacklisted(token)) {
      throw new Error('Token revoked');
    }
    
    return decoded;
  } catch (error) {
    throw new AuthenticationError(`Invalid token: ${error.message}`);
  }
}

// Authentication Middleware
async function authenticate(req, res, next) {
  try {
    const authHeader = req.headers.authorization;
    
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return res.status(401).json({ error: 'Missing or invalid authorization header' });
    }
    
    const token = authHeader.substring(7);
    const decoded = verifyJWT(token);
    
    // Attach user info to request
    req.user = {
      id: decoded.sub,
      email: decoded.email,
      scopes: decoded.scope?.split(' ') || [],
      tenantId: decoded.tenant_id,
      roles: decoded.roles || []
    };
    
    next();
  } catch (error) {
    res.status(401).json({ error: error.message });
  }
}

// Authorization Middleware (Scope-based)
function authorize(requiredScopes) {
  return (req, res, next) => {
    const userScopes = req.user.scopes || [];
    const hasRequiredScope = requiredScopes.every(scope => 
      userScopes.includes(scope)
    );
    
    if (!hasRequiredScope) {
      return res.status(403).json({ 
        error: 'Insufficient permissions',
        required: requiredScopes,
        provided: userScopes
      });
    }
    
    next();
  };
}

// API Key Management with Rate Limiting
async function authenticateAPIKey(req, res, next) {
  const apiKey = req.headers['x-api-key'];
  
  if (!apiKey) {
    return res.status(401).json({ error: 'Missing API key' });
  }
  
  try {
    // Lookup API key in cache or database
    let client = await redisClient.get(`api_key:${apiKey}`);
    
    if (!client) {
      client = await db.apiKeys.findOne({ 
        client_id: apiKey.split('_')[0],
        status: 'active',
        expires_at: { $gt: new Date() }
      });
      
      if (!client) {
        return res.status(401).json({ error: 'Invalid or expired API key' });
      }
      
      // Cache for 1 hour
      await redisClient.setEx(`api_key:${apiKey}`, 3600, JSON.stringify(client));
    }
    
    // Check rate limits
    const rateLimitKey = `ratelimit:${apiKey}`;
    const count = await redisClient.incr(rateLimitKey);
    
    if (count === 1) {
      await redisClient.expire(rateLimitKey, 60); // 1 minute window
    }
    
    if (count > client.rate_limit_per_minute) {
      return res.status(429).json({ 
        error: 'Rate limit exceeded',
        retry_after: 60
      });
    }
    
    req.client = client;
    next();
  } catch (error) {
    console.error('API key validation error:', error);
    res.status(500).json({ error: 'Authentication failed' });
  }
}
```

### Multi-Tenant Security Patterns

```javascript
// Tenant Isolation Middleware
async function tenantMiddleware(req, res, next) {
  const tenantId = req.user?.tenant_id || req.client?.tenant_id;
  
  if (!tenantId) {
    return res.status(403).json({ error: 'Tenant ID required' });
  }
  
  // Verify tenant exists and is active
  const tenant = await db.tenants.findById(tenantId);
  if (!tenant || tenant.status !== 'active') {
    return res.status(403).json({ error: 'Invalid or inactive tenant' });
  }
  
  req.tenantId = tenantId;
  req.tenant = tenant;
  next();
}

// BOLA Prevention: Always check tenant_id in queries
function tenantIsolated(queryField = 'tenant_id') {
  return (req, res, next) => {
    // Add tenant filter to all database queries
    req.tenantFilter = { [queryField]: req.tenantId };
    next();
  };
}

// Secure database query with tenant isolation
app.get('/api/users/:id', 
  authenticate(),
  tenantIsolated('tenant_id'),
  async (req, res) => {
    const user = await db.users.findOne({
      _id: req.params.id,
      ...req.tenantFilter // CRITICAL: Tenant isolation
    });
    
    if (!user) {
      return res.status(404).json({ error: 'User not found' });
    }
    
    // Additional ownership check for sensitive data
    if (user.tenant_id !== req.tenantId) {
      return res.status(403).json({ error: 'Access denied' });
    }
    
    res.json(user);
  }
);
```

**Related**: [Parent Skill](../SKILL.md) | [Rate Limiting Module](rate-limiting-protection.md)
