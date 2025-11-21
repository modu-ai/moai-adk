      const verified = jwt.verify(idToken, publicKey, {
        algorithms: ['RS256'],  // Ensure RS256
        issuer: this.config.issuerUrl,
        audience: this.config.clientId,
      });
      
      return verified;
    } catch (error) {
      throw new Error(`Token verification failed: ${error.message}`);
    }
  }
  
  // Validate Access Token
  validateAccessToken(accessToken) {
    const decoded = jwt.decode(accessToken, { complete: true });
    
    if (!decoded) {
      throw new Error('Invalid token format');
    }
    
    // 1. Check expiration
    const { payload } = decoded;
    const now = Math.floor(Date.now() / 1000);
    
    if (payload.exp <= now) {
      throw new Error('Token expired');
    }
    
    // 2. Check issuer
    if (payload.iss !== this.config.issuerUrl) {
      throw new Error('Invalid issuer');
    }
    
    return payload;
  }
  
  // Convert JWK to PEM format
  jwkToPem(jwk) {
    // Implementation using node-jose or similar
    // Returns PEM-formatted public key
    // ... (crypto conversion logic)
  }
}

// Usage
const oidcValidator = new OIDCValidator({
  issuerUrl: 'https://auth.example.com',
  clientId: 'my-app-id',
  clientSecret: 'my-secret',
  redirectUri: 'https://myapp.com/auth/callback',
});

await oidcValidator.initialize();

// In middleware
app.use((req, res, next) => {
  const authHeader = req.headers.authorization;
  if (!authHeader) {
    return res.status(401).json({ error: 'Missing authorization' });
  }
  
  const token = authHeader.replace('Bearer ', '');
  
  try {
    req.user = oidcValidator.validateIdToken(token);
    next();
  } catch (error) {
    res.status(401).json({ error: error.message });
  }
});
```

### Pattern 3: SCIM 2.0 User Provisioning

```javascript
class SCIMUserProvisioner {
  constructor(config) {
    this.config = config;
  }
  
  // Handle SCIM provisioning webhook from IdP
  handleScimWebhook(scimEvent) {
    switch (scimEvent.resourceType) {
      case 'User':
        return this.handleUserEvent(scimEvent);
      case 'Group':
        return this.handleGroupEvent(scimEvent);
      default:
        throw new Error(`Unknown resource type: ${scimEvent.resourceType}`);
    }
  }
  
  async handleUserEvent(event) {
    const { externalId, attributes } = event;
    
    switch (event.eventType) {
      case 'user.created':
        return this.createUser(attributes);
      
      case 'user.updated':
        return this.updateUser(externalId, attributes);
      
      case 'user.deleted':
        return this.deleteUser(externalId);
      
      default:
        throw new Error(`Unknown event: ${event.eventType}`);
    }
  }
  
  async createUser(attributes) {
    // Validate required fields
    if (!attributes.email || !attributes.userName) {
      throw new Error('Missing required fields');
    }
    
    // Create database record
    const user = await db.users.create({
      externalId: attributes.externalId,
      email: attributes.email,
      displayName: attributes.displayName,
      givenName: attributes.givenName,
      familyName: attributes.familyName,
      active: attributes.active ?? true,
      groups: attributes.groups || [],
    });
    
    return user;
  }
  
  async updateUser(externalId, attributes) {
    const user = await db.users.findByExternalId(externalId);
    
    if (!user) {
      throw new Error(`User not found: ${externalId}`);
    }
    
    // Update fields
    const updated = await db.users.update(user.id, {
      displayName: attributes.displayName,
      active: attributes.active,
      groups: attributes.groups || [],
    });
    
    return updated;
  }
  
  async deleteUser(externalId) {
    const user = await db.users.findByExternalId(externalId);
    
    if (!user) {
      throw new Error(`User not found: ${externalId}`);
    }
    
    // Soft delete (mark as inactive)
    await db.users.update(user.id, { active: false });
    
    return { success: true };
  }
  
  async handleGroupEvent(event) {
    // Similar pattern for group provisioning
    // Handle group.created, group.updated, group.deleted
  }
}

// Express endpoint for SCIM webhooks
app.post('/scim/webhook', async (req, res) => {
  try {
    // Verify webhook signature
    if (!verifyWebhookSignature(req)) {
      return res.status(401).json({ error: 'Invalid signature' });
    }
    
    const provisioner = new SCIMUserProvisioner(config);
    const result = await provisioner.handleScimWebhook(req.body);
    
    res.json(result);
  } catch (error) {
    console.error('SCIM webhook error:', error);
    res.status(400).json({ error: error.message });
  }
});
```

### Pattern 4: JWT Bearer Token in APIs

```javascript
class JWTMiddleware {
  constructor(publicKey) {
    this.publicKey = publicKey;
  }
  
  middleware() {
    return (req, res, next) => {
      const authHeader = req.headers.authorization;
      
      if (!authHeader || !authHeader.startsWith('Bearer ')) {
        return res.status(401).json({ error: 'Missing token' });
      }
      
      const token = authHeader.slice(7);  // Remove "Bearer "
      
      try {
        const payload = jwt.verify(token, this.publicKey, {
          algorithms: ['RS256'],
        });
        
        req.user = {
          id: payload.sub,
          email: payload.email,
          scope: payload.scope ? payload.scope.split(' ') : [],
        };
        
        next();
      } catch (error) {
        res.status(401).json({ error: 'Invalid token' });
      }
    };
  }
}

// Usage
const jwtMiddleware = new JWTMiddleware(publicKey);
app.use('/api', jwtMiddleware.middleware());

app.get('/api/protected', (req, res) => {
  res.json({
    message: `Hello, ${req.user.email}`,
    scopes: req.user.scope,
  });
});
```


## Checklist

- [ ] SAML 2.0 strategy configured with certificate verification
- [ ] OIDC provider discovery working
- [ ] JWT token signature validation implemented
- [ ] Token expiration checks in place
- [ ] SCIM webhook handling for user provisioning
- [ ] JIT (Just-In-Time) provisioning working
- [ ] Multi-protocol SSO (SAML + OIDC) tested
- [ ] Identity threat intelligence integrated
- [ ] SSO logout (SLO) working
- [ ] Performance tested at scale





## Context7 Integration

### Related Libraries & Tools
- [OAuth 2.0](/oauth-xx): Authorization framework
- [OpenID Connect](/openid/connect): Identity layer

### Official Documentation
- [Documentation](https://oauth.net/2/)
- [API Reference](https://openid.net/developers/specs/)

### Version-Specific Guides
Latest stable version: 2.1
- [Release Notes](https://oauth.net/2.1/)
- [Migration Guide](https://oauth.net/2.1/)
