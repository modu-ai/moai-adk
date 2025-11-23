---
name: moai-security-api__advanced-security-patterns
description: gRPC mTLS, webhook security, API versioning, CORS, and enterprise integration patterns
parent_skill: moai-security-api
module_type: advanced
---

## Advanced Security Patterns

### gRPC Security with mTLS

```javascript
const grpc = require('@grpc/grpc-js');
const fs = require('fs');

// mTLS Server Configuration
function createSecureServer() {
  const rootCert = fs.readFileSync('/secure/ca-cert.pem');
  const serverCert = fs.readFileSync('/secure/server-cert.pem');
  const serverKey = fs.readFileSync('/secure/server-key.pem');
  
  const serverCredentials = grpc.ServerCredentials.createSsl(
    rootCert,
    [{ cert_chain: serverCert, private_key: serverKey }]
  );
  
  const server = new grpc.Server();
  
  // JWT Interceptor for authentication
  const jwtInterceptor = (options, nextCall) => {
    const metadata = options.metadata || new grpc.Metadata();
    const token = metadata.get('authorization')[0];
    
    try {
      const decoded = jwt.verify(token, getPublicKey());
      options.metadata = metadata;
      options.metadata.set('user', JSON.stringify(decoded));
    } catch (error) {
      throw new grpc.status.PERMISSION_DENIED('Invalid token');
    }
    
    return nextCall(options);
  };
  
  server.bind('0.0.0.0:50051', serverCredentials);
  return server;
}

// Client mTLS Configuration
function createSecureClient() {
  const rootCert = fs.readFileSync('/secure/ca-cert.pem');
  const clientCert = fs.readFileSync('/secure/client-cert.pem');
  const clientKey = fs.readFileSync('/secure/client-key.pem');
  
  const clientCredentials = grpc.credentials.createSsl(
    rootCert, clientKey, clientCert
  );
  
  return new ServiceClient('api-server:50051', clientCredentials);
}
```

### Webhook Security (HMAC-SHA256)

```javascript
const crypto = require('crypto');

// Secure webhook delivery
async function sendSecureWebhook(event, url, secret) {
  const timestamp = Math.floor(Date.now() / 1000);
  const payload = JSON.stringify({ ...event, timestamp });
  
  // HMAC-SHA256 signature
  const signature = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'X-Webhook-Signature': `sha256=${signature}`,
      'X-Webhook-Timestamp': timestamp.toString(),
      'Content-Type': 'application/json'
    },
    body: payload,
    timeout: 30000
  });
  
  if (!response.ok) {
    throw new Error(`Webhook delivery failed: ${response.status}`);
  }
  
  return response.json();
}

// Webhook verification endpoint
app.post('/webhooks/stripe', 
  express.raw({ type: 'application/json' }),
  (req, res) => {
    const signature = req.headers['stripe-signature'];
    const timestamp = parseInt(req.headers['x-webhook-timestamp']);
    
    // Prevent replay attacks
    const age = Math.floor(Date.now() / 1000) - timestamp;
    if (age > 300) { // 5 minutes
      return res.status(401).json({ error: 'Webhook expired' });
    }
    
    // Verify signature
    const [version, hash] = signature.split(',')[0].split('=');
    const expected = crypto
      .createHmac('sha256', process.env.WEBHOOK_SECRET)
      .update(`${timestamp}.${req.body}`)
      .digest('hex');
    
    if (!crypto.timingSafeEqual(Buffer.from(hash), Buffer.from(expected))) {
      return res.status(401).json({ error: 'Invalid signature' });
    }
    
    // Process webhook safely
    const event = JSON.parse(req.body);
    processWebhookEvent(event);
    
    res.json({ received: true });
  }
);
```

### CORS & Security Headers

```javascript
const cors = require('cors');
const helmet = require('helmet');

// Production CORS configuration
const corsOptions = {
  origin: process.env.ALLOWED_ORIGINS?.split(',') || ['https://app.example.com'],
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-API-Key'],
  maxAge: 86400 // 24 hours
};

// Security headers configuration
app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      scriptSrc: ["'self'", "'unsafe-inline'"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      imgSrc: ["'self'", "data:", "https:"],
      connectSrc: ["'self'", process.env.API_BASE_URL]
    }
  },
  hsts: { maxAge: 31536000, includeSubDomains: true },
  noSniff: true,
  xssFilter: true,
  referrerPolicy: { policy: 'strict-origin-when-cross-origin' }
}));

app.use(cors(corsOptions));
```

### API Versioning with Deprecation

```javascript
// Semantic versioning with backward compatibility
function apiVersionMiddleware(req, res, next) {
  const version = req.path.match(/\/v(\d+)\//)?.[1] || '1';
  const currentVersion = 2;
  
  req.apiVersion = parseInt(version);
  
  // Deprecation warnings for old versions
  if (req.apiVersion < currentVersion) {
    res.set({
      'Deprecation': 'true',
      'Sunset': new Date('2026-01-01').toUTCString(),
      'Link': `</api/v${currentVersion}${req.path.replace(/\/v\d+/, '')}>; rel="successor-version"`
    });
  }
  
  next();
}

// Route handlers by version
app.get('/api/v1/users', legacyUserHandler); // Deprecated
app.get('/api/v2/users', currentUserHandler);  // Current
app.get('/api/v3/users', nextGenUserHandler);  // Beta
```

**Related**: [Parent Skill](../SKILL.md) | [Authentication Module](authentication-authorization.md)
