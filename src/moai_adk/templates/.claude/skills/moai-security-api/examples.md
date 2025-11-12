# moai-security-api: Production Examples

## Example 1: Complete REST API with OAuth 2.1 + JWT

```javascript
// server.js - Express.js API with OAuth 2.1
const express = require('express');
const passport = require('passport');
const OAuth2Strategy = require('passport-oauth2');
const jwt = require('jsonwebtoken');
const redis = require('redis');
const helmet = require('helmet');
const rateLimit = require('express-rate-limit');

const app = express();
const redisClient = redis.createClient();

// Security middleware
app.use(helmet());
app.use(express.json({ limit: '10kb' }));

// Rate limiting
const limiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100, // Limit each IP to 100 requests per windowMs
  standardHeaders: true,
  legacyHeaders: false
});

app.use(limiter);

// OAuth 2.1 Strategy
passport.use('oauth', new OAuth2Strategy({
  authorizationURL: 'https://auth-provider.com/oauth/authorize',
  tokenURL: 'https://auth-provider.com/oauth/token',
  clientID: process.env.OAUTH_CLIENT_ID,
  clientSecret: process.env.OAUTH_CLIENT_SECRET,
  callbackURL: 'https://api.example.com/auth/callback',
  state: true,
  pkce: true
}, async (accessToken, refreshToken, profile, done) => {
  try {
    // Verify user and create/update in DB
    const user = await db.users.findOrCreate({
      oauthId: profile.id,
      email: profile.email
    });
    done(null, user);
  } catch (err) {
    done(err);
  }
}));

// Endpoint: Start OAuth flow
app.get('/auth/login', passport.authenticate('oauth'));

// Endpoint: OAuth callback
app.get('/auth/callback',
  passport.authenticate('oauth', { session: false }),
  (req, res) => {
    const token = jwt.sign(
      { id: req.user.id, scope: 'read:api' },
      process.env.JWT_SECRET,
      { algorithm: 'RS256', expiresIn: '7d' }
    );
    
    res.json({ token, user: req.user });
  }
);

// Middleware: JWT verification
function jwtAuth(req, res, next) {
  const authHeader = req.headers.authorization;
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'Missing token' });
  }
  
  const token = authHeader.slice(7);
  
  try {
    const decoded = jwt.verify(token, process.env.JWT_PUBLIC_KEY, {
      algorithms: ['RS256'],
      issuer: 'https://api.example.com'
    });
    
    req.user = decoded;
    next();
  } catch (err) {
    return res.status(401).json({ error: 'Invalid token' });
  }
}

// Endpoint: Get user profile
app.get('/api/users/me', jwtAuth, (req, res) => {
  res.json({ id: req.user.id, role: req.user.role });
});

// Endpoint: List users (admin only)
app.get('/api/users', jwtAuth, (req, res) => {
  if (req.user.role !== 'admin') {
    return res.status(403).json({ error: 'Forbidden' });
  }
  
  res.json([/* users */]);
});

app.listen(3000);
```

## Example 2: GraphQL API with Query Complexity Limits

```javascript
// graphql-server.js - Apollo Server with security
const { ApolloServer } = require('@apollo/server');
const { expressMiddleware } = require('@apollo/server/express4');
const express = require('express');
const { getComplexity, simpleEstimator } = require('@graphql-query-estimator/server');

const typeDefs = `#graphql
  type User {
    id: ID!
    name: String!
    email: String!
    posts(limit: Int = 10): [Post!]!
  }
  
  type Post {
    id: ID!
    title: String!
    content: String!
    comments: [Comment!]!
  }
  
  type Comment {
    id: ID!
    text: String!
    author: User!
  }
  
  type Query {
    user(id: ID!): User
    posts(limit: Int = 20): [Post!]!
  }
`;

const resolvers = {
  Query: {
    user: async (_, { id }, context) => {
      if (!context.user) throw new Error('Unauthorized');
      return db.users.findById(id);
    },
    posts: async (_, { limit }, context) => {
      return db.posts.find({}).limit(limit);
    }
  },
  Post: {
    comments: async (post) => {
      return db.comments.find({ postId: post.id });
    }
  }
};

const server = new ApolloServer({
  typeDefs,
  resolvers,
  
  // Security plugin: Check complexity before execution
  plugins: {
    async didResolveOperation(requestContext) {
      const complexity = getComplexity({
        schema: server.schema,
        operationAST: requestContext.document,
        estimators: [simpleEstimator({ defaultComplexity: 1 })]
      });
      
      // Reject complex queries
      if (complexity > 1000) {
        throw new Error(
          `Query is too complex: ${complexity} (max 1000)`
        );
      }
      
      // Log complexity for monitoring
      console.log(`Query complexity: ${complexity}`);
    }
  }
});

const app = express();

// Authentication middleware
app.use((req, res, next) => {
  const token = req.headers.authorization?.split(' ')[1];
  if (token) {
    try {
      const decoded = jwt.verify(token, process.env.JWT_SECRET);
      req.user = decoded;
    } catch (err) {
      // Invalid token
    }
  }
  next();
});

app.use('/graphql', expressMiddleware(server, {
  context: async ({ req }) => ({
    user: req.user,
    db
  })
}));

app.listen(4000);
```

## Example 3: API Key Management with Rate Limiting

```javascript
// api-key-manager.js
const crypto = require('crypto');
const redis = require('redis');
const bcrypt = require('bcryptjs');

class APIKeyManager {
  constructor() {
    this.redis = redis.createClient();
  }
  
  // Generate new API key
  async generateKey(clientId, name, rateLimitPerMinute = 100) {
    // Format: sk_live_xxxxxxxxxxxxxxxxxxxx
    const key = `sk_${process.env.NODE_ENV}_${crypto.randomBytes(16).toString('hex')}`;
    
    // Hash key for storage
    const hashedKey = bcrypt.hashSync(key, 10);
    
    // Store in database
    const stored = await db.apiKeys.create({
      client_id: clientId,
      key_hash: hashedKey,
      name,
      rate_limit_per_minute: rateLimitPerMinute,
      created_at: new Date(),
      status: 'active'
    });
    
    // Cache for quick lookup (hash → metadata)
    await this.redis.setex(
      `api_key:${crypto.createHash('sha256').update(key).digest('hex')}`,
      3600, // 1 hour
      JSON.stringify({
        client_id: clientId,
        rate_limit: rateLimitPerMinute
      })
    );
    
    return { key, id: stored.id };
  }
  
  // Validate API key and check rate limit
  async validateAndRateLimit(apiKey) {
    const keyHash = crypto.createHash('sha256').update(apiKey).digest('hex');
    
    // Check cache first
    let metadata = await this.redis.get(`api_key:${keyHash}`);
    
    if (!metadata) {
      // Hit database
      const key = await db.apiKeys.findOne({ 
        key_hash: bcrypt.hashSync(apiKey, 10),
        status: 'active'
      });
      
      if (!key) throw new Error('Invalid API key');
      
      metadata = {
        client_id: key.client_id,
        rate_limit: key.rate_limit_per_minute
      };
      
      // Cache result
      await this.redis.setex(
        `api_key:${keyHash}`,
        3600,
        JSON.stringify(metadata)
      );
    } else {
      metadata = JSON.parse(metadata);
    }
    
    // Rate limit check
    const rateLimitKey = `ratelimit:${apiKey}`;
    const count = await this.redis.incr(rateLimitKey);
    
    if (count === 1) {
      await this.redis.expire(rateLimitKey, 60);
    }
    
    if (count > metadata.rate_limit) {
      throw new Error('Rate limit exceeded');
    }
    
    return metadata;
  }
  
  // Rotate keys (generate new, mark old as deprecated)
  async rotateKey(keyId) {
    const oldKey = await db.apiKeys.findById(keyId);
    
    // Mark old as deprecated (grace period for rotation)
    await db.apiKeys.update(keyId, {
      status: 'deprecated',
      deprecated_at: new Date()
    });
    
    // Generate new key
    return this.generateKey(
      oldKey.client_id,
      `${oldKey.name} (rotated)`,
      oldKey.rate_limit_per_minute
    );
  }
  
  // Revoke key immediately
  async revokeKey(keyId) {
    await db.apiKeys.update(keyId, {
      status: 'revoked',
      revoked_at: new Date()
    });
    
    // Clear cache
    const key = await db.apiKeys.findById(keyId);
    await this.redis.del(`api_key:${crypto.createHash('sha256').update(key.key).digest('hex')}`);
  }
}

// Middleware usage
const keyManager = new APIKeyManager();

app.use(async (req, res, next) => {
  const apiKey = req.headers['x-api-key'];
  
  if (!apiKey) {
    return res.status(401).json({ error: 'Missing API key' });
  }
  
  try {
    req.client = await keyManager.validateAndRateLimit(apiKey);
    next();
  } catch (err) {
    if (err.message.includes('Rate limit')) {
      return res.status(429).json({ error: err.message });
    }
    return res.status(401).json({ error: 'Unauthorized' });
  }
});
```

## Example 4: Webhook Signature Verification (Stripe Pattern)

```javascript
// webhook-handler.js
const express = require('express');
const crypto = require('crypto');

const app = express();

// Use raw body for webhook signature verification
app.post('/webhooks/stripe',
  express.raw({ type: 'application/json' }),
  (req, res) => {
    const sig = req.headers['stripe-signature'];
    const rawBody = req.rawBody;
    
    try {
      // Verify signature
      const event = stripe.webhooks.constructEvent(
        rawBody,
        sig,
        process.env.WEBHOOK_SECRET
      );
      
      // Handle event
      switch (event.type) {
        case 'payment_intent.succeeded':
          handlePaymentSuccess(event.data.object);
          break;
        case 'customer.subscription.deleted':
          handleSubscriptionCanceled(event.data.object);
          break;
      }
      
      res.json({ received: true });
    } catch (err) {
      console.error('Webhook signature verification failed:', err);
      res.status(400).send(`Webhook Error: ${err.message}`);
    }
  }
);

// Manual verification example
function verifyWebhookSignature(payload, signature, secret) {
  const [t, v1] = signature.split(',');
  const timestamp = t.split('=')[1];
  const hash = v1.split('=')[1];
  
  // Prevent replay attacks
  const now = Math.floor(Date.now() / 1000);
  if (Math.abs(now - parseInt(timestamp)) > 300) {
    throw new Error('Webhook timestamp too old');
  }
  
  // Verify signature
  const signedContent = `${timestamp}.${payload}`;
  const expected = crypto
    .createHmac('sha256', secret)
    .update(signedContent)
    .digest('hex');
  
  if (!crypto.timingSafeEqual(Buffer.from(hash), Buffer.from(expected))) {
    throw new Error('Invalid signature');
  }
  
  return true;
}
```

## Example 5: Multi-Tenant API with Tenant Isolation

```javascript
// multi-tenant-api.js
const express = require('express');
const jwt = require('jsonwebtoken');

const app = express();

// Middleware: Extract and verify tenant
app.use((req, res, next) => {
  const token = req.headers.authorization?.split(' ')[1];
  
  if (!token) {
    return res.status(401).json({ error: 'Missing token' });
  }
  
  try {
    const decoded = jwt.verify(token, process.env.JWT_SECRET);
    
    // Token MUST include tenant_id
    if (!decoded.tenant_id) {
      return res.status(403).json({ error: 'Invalid token: missing tenant_id' });
    }
    
    req.user = decoded;
    req.tenantId = decoded.tenant_id;
    next();
  } catch (err) {
    return res.status(401).json({ error: 'Invalid token' });
  }
});

// CRITICAL: Always filter by tenant_id
app.get('/api/users', (req, res) => {
  // Filter by tenant_id - prevent BOLA/IDOR
  const users = db.users.find({ tenant_id: req.tenantId });
  res.json(users);
});

// CRITICAL: Verify tenant ownership on record access
app.get('/api/users/:userId', (req, res) => {
  const user = db.users.findById(req.params.userId);
  
  // Check tenant ownership
  if (user.tenant_id !== req.tenantId) {
    return res.status(403).json({ error: 'Access denied' });
  }
  
  res.json(user);
});

// CRITICAL: Verify tenant ownership on write operations
app.put('/api/users/:userId', (req, res) => {
  const user = db.users.findById(req.params.userId);
  
  if (user.tenant_id !== req.tenantId) {
    return res.status(403).json({ error: 'Access denied' });
  }
  
  // User can only update their own data (not admin)
  if (req.user.id !== req.params.userId && req.user.role !== 'admin') {
    return res.status(403).json({ error: 'Access denied' });
  }
  
  const updated = db.users.update(req.params.userId, req.body);
  res.json(updated);
});
```

