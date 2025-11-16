# SAML 2.0 & OIDC - Production Code Examples

## Example 1: SAML 2.0 Strategy

```javascript
// auth/saml-strategy.js
const passport = require('passport');
const { Strategy } = require('@node-saml/passport-saml');
const fs = require('fs');

module.exports = (app) => {
  const samlStrategy = new Strategy(
    {
      entryPoint: process.env.SAML_ENTRY_POINT,
      issuer: process.env.SAML_ISSUER,
      callbackURL: process.env.SAML_CALLBACK_URL,
      cert: fs.readFileSync(process.env.SAML_CERT_PATH, 'utf-8'),
      privateKey: fs.readFileSync(process.env.SAML_PRIVATE_KEY_PATH, 'utf-8'),
      validateInResponseTo: true,
      wantAssertionsSigned: true,
      identifierFormat: 'urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress',
    },
    (profile, done) => {
      const user = {
        id: profile.nameID,
        email: profile.attributes.email,
        displayName: profile.attributes.displayName,
        groups: profile.attributes.groups || [],
      };
      
      done(null, user);
    }
  );
  
  passport.use('saml', samlStrategy);
  
  app.get('/auth/saml', passport.authenticate('saml'));
  
  app.post('/auth/saml/callback', (req, res, next) => {
    passport.authenticate('saml', (err, user) => {
      if (err || !user) {
        return res.redirect('/login?error=authentication_failed');
      }
      
      req.logIn(user, (err) => {
        if (err) return next(err);
        res.redirect('/dashboard');
      });
    })(req, res, next);
  });
  
  app.get('/auth/saml/metadata', (req, res) => {
    const metadata = samlStrategy.generateServiceProviderMetadata();
    res.type('application/xml').send(metadata);
  });
};
```

## Example 2: OIDC Authorization Code Flow

```javascript
// auth/oidc-auth.js
const { Issuer } = require('openid-client');
const passport = require('passport');
const { Strategy } = require('openid-client').Passport;

module.exports = async (app) => {
  // Discover OIDC provider
  const issuer = await Issuer.discover(process.env.OIDC_ISSUER_URL);
  
  const client = new issuer.Client({
    client_id: process.env.OIDC_CLIENT_ID,
    client_secret: process.env.OIDC_CLIENT_SECRET,
    redirect_uris: [process.env.OIDC_REDIRECT_URI],
    response_types: ['code'],
  });
  
  const oidcStrategy = new Strategy(
    { client, params: {} },
    (tokenSet, userInfo, done) => {
      const user = {
        id: userInfo.sub,
        email: userInfo.email,
        displayName: userInfo.name,
        emailVerified: userInfo.email_verified,
      };
      
      done(null, user);
    }
  );
  
  passport.use('oidc', oidcStrategy);
  
  app.get('/auth/oidc', passport.authenticate('oidc'));
  
  app.get('/auth/oidc/callback', (req, res, next) => {
    passport.authenticate('oidc', {
      successRedirect: '/dashboard',
      failureRedirect: '/login',
    })(req, res, next);
  });
};
```

## Example 3: JWT Bearer Token Validation

```javascript
// middleware/jwt-auth.js
const jwt = require('jsonwebtoken');
const fs = require('fs');

class JWTAuthMiddleware {
  constructor() {
    this.publicKey = fs.readFileSync(process.env.JWT_PUBLIC_KEY_PATH, 'utf-8');
  }
  
  middleware() {
    return (req, res, next) => {
      const authHeader = req.headers.authorization;
      
      if (!authHeader || !authHeader.startsWith('Bearer ')) {
        return res.status(401).json({ error: 'Missing token' });
      }
      
      const token = authHeader.slice(7);
      
      try {
        const payload = jwt.verify(token, this.publicKey, {
          algorithms: ['RS256'],
          issuer: process.env.JWT_ISSUER,
          audience: process.env.JWT_AUDIENCE,
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

module.exports = JWTAuthMiddleware;
```

## Example 4: SCIM User Provisioning Webhook

```javascript
// webhooks/scim-provisioning.js
class SCIMProvisioningHandler {
  constructor(db) {
    this.db = db;
  }
  
  async handleWebhook(req, res) {
    // Verify webhook signature
    if (!this.verifySignature(req)) {
      return res.status(401).json({ error: 'Invalid signature' });
    }
    
    const { resourceType, eventType, attributes } = req.body;
    
    try {
      switch (resourceType) {
        case 'User':
          if (eventType === 'user.created') {
            await this.createUser(attributes);
          } else if (eventType === 'user.updated') {
            await this.updateUser(attributes);
          } else if (eventType === 'user.deleted') {
            await this.deleteUser(attributes.id);
          }
          break;
        
        case 'Group':
          if (eventType === 'group.created') {
            await this.createGroup(attributes);
          }
          break;
      }
      
      res.json({ success: true });
    } catch (error) {
      res.status(400).json({ error: error.message });
    }
  }
  
  async createUser(attributes) {
    const user = {
      externalId: attributes.id,
      email: attributes.email,
      displayName: attributes.displayName,
      active: attributes.active ?? true,
      createdAt: new Date(),
    };
    
    await this.db.users.insertOne(user);
    console.log(`User created: ${attributes.email}`);
  }
  
  async updateUser(attributes) {
    await this.db.users.updateOne(
      { externalId: attributes.id },
      {
        $set: {
          displayName: attributes.displayName,
          active: attributes.active,
          updatedAt: new Date(),
        },
      }
    );
  }
  
  async deleteUser(userId) {
    await this.db.users.updateOne(
      { externalId: userId },
      { $set: { active: false, deletedAt: new Date() } }
    );
  }
  
  verifySignature(req) {
    // Implement HMAC signature verification
    const signature = req.headers['x-webhook-signature'];
    const body = JSON.stringify(req.body);
    const hash = crypto
      .createHmac('sha256', process.env.SCIM_WEBHOOK_SECRET)
      .update(body)
      .digest('hex');
    
    return signature === hash;
  }
}

module.exports = SCIMProvisioningHandler;
```

## Example 5: Context7 MCP Identity Threat Detection

```javascript
// security/context7-identity-threats.js
const { Context7Client } = require('context7-mcp');

class IdentityThreatDetection {
  constructor(apiKey) {
    this.context7 = new Context7Client({ apiKey });
  }
  
  async checkUserIdentity(user) {
    const threat = await this.context7.query({
      type: 'identity_threat',
      email: user.email,
      userId: user.id,
      tags: ['compromise', 'fraud'],
    });
    
    return {
      safe: threat.severity === 0,
      severity: threat.severity,
      actions: this.getSecurityActions(threat),
    };
  }
  
  getSecurityActions(threat) {
    const actions = [];
    
    if (threat.severity >= 'HIGH') {
      actions.push('BLOCK_LOGIN');
      actions.push('REQUIRE_MFA');
      actions.push('ALERT_SECURITY_TEAM');
    } else if (threat.severity === 'MEDIUM') {
      actions.push('REQUIRE_MFA');
      actions.push('LOG_EVENT');
    }
    
    return actions;
  }
}

module.exports = IdentityThreatDetection;
```

