# Skill: moai-baas-auth0-ext

## Metadata

```yaml
skill_id: moai-baas-auth0-ext
skill_name: Auth0 Enterprise Authentication & Identity Management
version: 1.0.0
created_date: 2025-11-09
language: english
triggers:
  - keywords: ["Auth0", "Enterprise Auth", "SAML", "OIDC", "Identity", "SSO"]
  - contexts: ["auth0-detected", "pattern-h", "enterprise-authentication"]
agents:
  - security-expert
  - backend-expert
  - devops-expert
freedom_level: high
word_count: 1000
context7_references:
  - url: "https://auth0.com/docs/get-started"
    topic: "Auth0 Integration & Setup"
  - url: "https://auth0.com/docs/protocols/openid-connect"
    topic: "OpenID Connect (OIDC) Protocol"
  - url: "https://auth0.com/docs/saml/saml-configuration"
    topic: "SAML 2.0 Configuration"
  - url: "https://auth0.com/docs/rules"
    topic: "Rules & Hooks for Custom Logic"
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

---

## 📚 Content

### 1. Auth0 Enterprise Architecture (150 words)

**Auth0** is an enterprise-grade identity and authentication platform supporting complex authentication scenarios.

**Core Philosophy**:
```
Consumer Authentication:
  Client → Simple provider (Google, Facebook)
  └─ Good for: MVPs, consumer apps

Enterprise Authentication:
  Client → Auth0 → Multiple identity providers (SAML, Active Directory, OIDC)
  └─ Good for: B2B, large organizations, compliance-heavy scenarios
```

**Auth0 Platform Components**:
```
┌──────────────────────────────────────────┐
│ Auth0 (Enterprise Identity Platform)     │
├──────────────────────────────────────────┤
│                                           │
│ 1. Universal Login                       │
│    └─ Branded login experience            │
│                                           │
│ 2. Social & OAuth Connections            │
│    └─ Google, GitHub, Facebook, etc.     │
│                                           │
│ 3. Enterprise Connections (SAML/OIDC)   │
│    └─ Active Directory, Okta, etc.       │
│                                           │
│ 4. Multi-Factor Authentication (MFA)     │
│    └─ Authenticator apps, SMS, push      │
│                                           │
│ 5. Custom Database Connections           │
│    └─ Legacy systems integration          │
│                                           │
│ 6. Rules & Hooks                         │
│    └─ Custom authentication logic         │
│                                           │
│ 7. Management API                        │
│    └─ Programmatic user management       │
│                                           │
│ 8. Logs & Analytics                      │
│    └─ Security & compliance reporting     │
└──────────────────────────────────────────┘
```

**When to Use Auth0**:
- ✅ Enterprise customers require SAML/OIDC
- ✅ Complex authentication flows
- ✅ Compliance requirements (SOC 2, HIPAA)
- ✅ Multi-tenant applications
- ✅ Password-less authentication
- ⚠️ Higher cost ($0.08 per MAU - monthly active user)

---

### 2. Auth0 Integration & SDK Setup (200 words)

**Auth0 Configuration**:

```bash
# 1. Create Auth0 application
# Dashboard: Applications → Create Application → Regular Web App/SPA

# 2. Configure allowed URLs
# Allowed Callback URLs: http://localhost:3000/callback
# Allowed Logout URLs: http://localhost:3000
# Allowed Web Origins: http://localhost:3000
```

**Frontend Integration (React)**:

```typescript
// src/auth0-provider.tsx
import React from "react";
import { Auth0Provider } from "@auth0/auth0-react";

export function AppWithAuth0({ children }) {
  return (
    <Auth0Provider
      domain={process.env.REACT_APP_AUTH0_DOMAIN}
      clientId={process.env.REACT_APP_AUTH0_CLIENT_ID}
      redirectUri={window.location.origin}
      audience={process.env.REACT_APP_AUTH0_AUDIENCE}
      scope="openid profile email"
    >
      {children}
    </Auth0Provider>
  );
}

// Usage in component
import { useAuth0 } from "@auth0/auth0-react";

export function LoginButton() {
  const { loginWithRedirect } = useAuth0();

  return (
    <button onClick={() => loginWithRedirect()}>Log In with Auth0</button>
  );
}

export function UserProfile() {
  const { user, isAuthenticated } = useAuth0();

  if (!isAuthenticated) return null;

  return (
    <div>
      <img src={user.picture} alt={user.name} />
      <h2>{user.name}</h2>
      <p>{user.email}</p>
    </div>
  );
}
```

**Backend Integration (Node.js)**:

```typescript
// server.ts
import { ManagementClient } from "auth0";

const management = new ManagementClient({
  domain: process.env.AUTH0_DOMAIN,
  clientId: process.env.AUTH0_MGMT_CLIENT_ID,
  clientSecret: process.env.AUTH0_MGMT_CLIENT_SECRET,
});

// Verify JWT token
import { auth } from "express-oauth2-jwt-bearer";

app.use(
  auth({
    audience: process.env.AUTH0_AUDIENCE,
    issuerBaseURL: `https://${process.env.AUTH0_DOMAIN}/`,
  })
);

// Protected route
app.get("/api/profile", (req, res) => {
  const userId = req.auth.sub; // From JWT
  res.json({ userId, email: req.auth.payload.email });
});
```

---

### 3. SAML & Enterprise Connections (200 words)

**SAML 2.0** enables enterprise customers to use their corporate identity providers (Okta, Active Directory, etc.).

**Configuring SAML Connection**:

```bash
# 1. Create SAML Enterprise Connection
# Dashboard: Connections → Enterprise → SAML
# Name: "Company SAML"

# 2. Upload IdP Metadata (from customer's Okta/Salesforce)
# - Entity ID
# - Single Sign-On URL
# - Certificate

# 3. Configure your metadata
# Auth0 provides:
# - Entity ID: urn:auth0:yourtenant:samlp
# - ACS URL: https://yourtenant.auth0.com/login/callback?connection=Company%20SAML
```

**SAML Rules for Custom Logic**:

```javascript
// Auth0 Rules (server-side, execute on every login)
function addCompanyMetadata(user, context, callback) {
  // Extract company from SAML attributes
  const company = user['http://schemas.xmlsoap.org/ws/2005/05/identity/claims/companyname'];

  // Store in metadata
  context.idToken['custom:company'] = company;
  context.idToken['custom:department'] = user['department'];

  callback(null, user, context);
}
```

**OIDC Configuration**:

```typescript
// OpenID Connect flow (for applications)
// 1. Redirect to Auth0 authorize endpoint
// GET https://yourtenant.auth0.com/authorize
//   ?client_id=xxx
//   &response_type=code
//   &redirect_uri=http://localhost:3000/callback
//   &scope=openid%20profile%20email
//   &state=xyz123

// 2. Auth0 redirects back with code
// GET http://localhost:3000/callback?code=abc123&state=xyz123

// 3. Exchange code for tokens
const tokens = await fetch(`https://yourtenant.auth0.com/oauth/token`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    client_id: "xxx",
    client_secret: "yyy",
    code: "abc123",
    redirect_uri: "http://localhost:3000/callback",
    grant_type: "authorization_code",
  }),
});

// Response includes:
// {
//   "access_token": "...",
//   "id_token": "...",  (JWT containing user info)
//   "token_type": "Bearer"
// }
```

---

### 4. Multi-Factor Authentication & Security (200 words)

**MFA in Auth0** provides multiple authentication methods.

**MFA Options**:
- SMS (one-time codes)
- Authenticator apps (Google Authenticator, Duo)
- Biometric (fingerprint, face)
- Security keys (FIDO2)

**Enforcing MFA**:

```javascript
// Auth0 Rule - Require MFA for admins
function enforceMFAForAdmins(user, context, callback) {
  // Check if user is admin
  const isAdmin = user.roles && user.roles.includes("admin");

  if (isAdmin && !context.multifactor) {
    return callback(
      new UnauthorizedError("MFA is required for admin accounts")
    );
  }

  callback(null, user, context);
}

// Auth0 Hook - Custom MFA logic
exports.onExecutePostLogin = async (event, api) => {
  if (event.user.user_metadata?.mfa_required) {
    api.authentication.challengeWithAny([
      {
        authenticator_types: ["otp"],
      },
    ]);
  }
};
```

**OAuth 2.0 Device Flow** (for IoT):

```typescript
// 1. Client initiates
const deviceAuth = await fetch(`https://yourtenant.auth0.com/oauth/device_authorization`, {
  method: "POST",
  body: `client_id=${CLIENT_ID}&scope=openid%20profile`,
});

// 2. Response includes device_code & user_code
// User enters code in browser at: https://yourtenant.auth0.com/device

// 3. Poll token endpoint
const tokenResponse = await fetch(`https://yourtenant.auth0.com/oauth/token`, {
  method: "POST",
  body: JSON.stringify({
    client_id: CLIENT_ID,
    device_code: DEVICE_CODE,
    grant_type: "urn:ietf:params:oauth:grant-type:device_code",
  }),
});
```

---

### 5. Rules, Hooks & Custom Logic (200 words)

**Rules** (deprecated → use Hooks/Actions) execute custom code during authentication.

**Actions** (new architecture - recommended):

```javascript
// actions/enrich-profile
exports.onExecutePostLogin = async (event, api) => {
  // Fetch additional user data
  const metadata = await fetchUserMetadata(event.user.user_id);

  // Add to ID token
  api.idToken.setCustomClaim("custom:company", metadata.company);
  api.idToken.setCustomClaim("custom:role", metadata.role);

  // Log authentication event
  console.log(`User ${event.user.email} authenticated via ${event.connection}`);
};

// For flagging risky logins
exports.onExecutePostLogin = async (event, api) => {
  if (isLocationAnomalous(event)) {
    api.authentication.challengeWithOOB(
      {
        channel: "sms",
        authenticator_type: "oob",
      },
      api
    );
  }
};
```

**Webhooks** for external integrations:

```bash
# Configure Webhook in Dashboard
# POST https://yourapp.com/auth0-webhook
# Include: Authorization header with webhook secret

# Webhook events:
# - user.created
# - user.deleted
# - user.updated
# - failed_login
# - success_login
```

**Management API** for programmatic access:

```typescript
import { ManagementClient } from "auth0";

const mgmt = new ManagementClient({
  domain: process.env.AUTH0_DOMAIN,
  clientId: process.env.AUTH0_MGMT_CLIENT_ID,
  clientSecret: process.env.AUTH0_MGMT_CLIENT_SECRET,
});

// Create user
await mgmt.users.create({
  email: "user@example.com",
  password: "SecurePassword123!",
  connection: "Username-Password-Authentication",
  user_metadata: { plan: "premium" },
});

// Assign role
await mgmt.users.assignRoles({ id: USER_ID }, { roles: [ROLE_ID] });

// Revoke token
await mgmt.tokens.revoke({ token: REFRESH_TOKEN });
```

---

### 6. Cost Optimization & Common Issues (50 words)

| Issue | Solution |
|-------|----------|
| **High MAU costs** | Review inactive users, purge old accounts |
| **SAML not working** | Check IdP metadata, certificate expiry |
| **MFA enrollment low** | Use Rules to encourage opt-in gradually |
| **Token expiry issues** | Implement refresh token rotation |

---

## 🎯 Usage

### Invocation from Agents
```python
Skill("moai-baas-auth0-ext")
# Load when Pattern H (Auth0 Enterprise) detected
```

### Context7 Integration
When Auth0 platform detected:
- SAML & OIDC enterprise flows
- Multi-factor authentication setup
- Rules, Hooks, & Actions for custom logic
- Management API for user provisioning

---

## 📚 Reference Materials

- [Auth0 Getting Started](https://auth0.com/docs/get-started)
- [OpenID Connect (OIDC) Protocol](https://auth0.com/docs/protocols/openid-connect)
- [SAML 2.0 Configuration](https://auth0.com/docs/saml/saml-configuration)
- [Rules & Hooks](https://auth0.com/docs/rules)
- [Management API](https://auth0.com/docs/api/management/v2)

---

## ✅ Validation Checklist

- [x] Enterprise architecture & platform components
- [x] Frontend & backend SDK integration
- [x] SAML & OIDC protocol configuration
- [x] Multi-factor authentication (MFA)
- [x] Rules, Hooks, Actions & Management API
- [x] Cost optimization & troubleshooting
- [x] 1000-word target
- [x] English language (policy compliant)
