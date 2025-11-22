---
name: moai-security-api__rate-limiting-protection
description: Token bucket rate limiting, distributed Redis-based rate limiting, and DoS protection
parent_skill: moai-security-api
module_type: core
---

## Rate Limiting & Protection Patterns

### Token Bucket Rate Limiting (Redis Lua Script)

```javascript
const rateLimitLuaScript = `
  local key = KEYS[1]
  local capacity = tonumber(ARGV[1])
  local refill_rate = tonumber(ARGV[2])
  local now = tonumber(ARGV[3])
  
  local tokens = tonumber(redis.call('GET', key) or capacity)
  local last_refill = tonumber(redis.call('GET', key .. ':refill') or now)
  
  -- Calculate tokens to add
  local time_passed = now - last_refill
  local tokens_to_add = math.floor(time_passed * refill_rate / 1000)
  tokens = math.min(capacity, tokens + tokens_to_add)
  
  if tokens >= 1 then
    tokens = tokens - 1
    redis.call('SET', key, tokens)
    redis.call('SET', key .. ':refill', now)
    redis.call('EXPIRE', key, 3600)
    redis.call('EXPIRE', key .. ':refill', 3600)
    return 1
  else
    redis.call('SET', key, tokens)
    redis.call('SET', key .. ':refill', now)
    redis.call('EXPIRE', key, 3600)
    redis.call('EXPIRE', key .. ':refill', 3600)
    return 0
  end
`;

async function rateLimit(capacity = 100, refillRate = 10) {
  return async (req, res, next) => {
    const userId = req.user?.id || req.client?.id || 'anonymous';
    const key = `ratelimit:${userId}`;
    
    try {
      const allowed = await redisClient.eval(rateLimitLuaScript, {
        keys: [key],
        arguments: [capacity, refillRate, Date.now()]
      });
      
      if (!allowed) {
        return res.status(429).json({
          error: 'Rate limit exceeded',
          retry_after: Math.ceil(capacity / refillRate)
        });
      }
      
      // Add rate limit headers
      res.set({
        'X-RateLimit-Limit': capacity,
        'X-RateLimit-Remaining': Math.max(0, capacity - (await redisClient.get(key) || 0)),
        'X-RateLimit-Reset': new Date(Date.now() + 1000).toUTCString()
      });
      
      next();
    } catch (error) {
      console.error('Rate limiting error:', error);
      next(); // Fail open on rate limiting errors
    }
  };
}
```

### GraphQL Query Complexity Analysis

```javascript
const { ApolloServer } = require('@apollo/server');

// Query Complexity Analysis
function calculateQueryComplexity(document, operation) {
  // Simple complexity calculation
  let complexity = 0;
  let depth = 0;
  
  const visitNode = (node, currentDepth) => {
    depth = Math.max(depth, currentDepth);
    
    if (node.kind === 'Field') {
      complexity += 1; // Base cost per field
      
      // Higher cost for expensive fields
      const expensiveFields = ['users', 'posts', 'analytics'];
      if (expensiveFields.includes(node.name.value)) {
        complexity += 10;
      }
    }
    
    if (node.selectionSet) {
      node.selectionSet.selections.forEach(selection => 
        visitNode(selection, currentDepth + 1)
      );
    }
  };
  
  visitNode(operation, 0);
  return { complexity, depth };
}

// Apollo Server Security Configuration
const server = new ApolloServer({
  typeDefs,
  resolvers,
  
  plugins: [{
    requestDidStart() {
      return {
        didResolveOperation(requestContext) {
          const { complexity, depth } = calculateQueryComplexity(
            requestContext.document,
            requestContext.request.operation
          );
          
          // Prevent DoS queries
          if (complexity > 1000) {
            throw new Error(`Query too complex: ${complexity} (max: 1000)`);
          }
          
          if (depth > 7) {
            throw new Error(`Query too deep: ${depth} (max: 7)`);
          }
        }
      };
    }
  }],
  
  // Security settings
  introspection: process.env.NODE_ENV !== 'production',
  executionTimeoutMs: 5000,
  
  context: ({ req }) => ({
    user: req.user,
    tenantId: req.tenantId,
    scopes: req.user?.scopes || []
  })
});

// Field-level Authorization
const resolvers = {
  Query: {
    users: (parent, args, context) => {
      if (!context.scopes.includes('read:users')) {
        throw new ForbiddenError('Missing read:users scope');
      }
      
      return db.users.findByTenant(context.tenantId);
    },
    
    sensitiveData: (parent, args, context) => {
      if (!context.scopes.includes('read:sensitive') || 
          !context.roles.includes('admin')) {
        throw new ForbiddenError('Admin access required');
      }
      
      return db.sensitive.findByTenant(context.tenantId);
    }
  }
};
```

**Related**: [Parent Skill](../SKILL.md) | [Authentication Module](authentication-authorization.md)
