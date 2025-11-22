# Auth0 - Practical Examples

## Example 1: Basic Login Implementation

```javascript
// Auth0 JavaScript SDK integration
import { Auth0Client } from '@auth0/auth0-spa-js';

let auth0 = null;

const configureClient = async () => {
    auth0 = await Auth0Client.create({
        domain: process.env.AUTH0_DOMAIN,
        clientId: process.env.AUTH0_CLIENT_ID,
        cacheLocation: 'memory'
    });
};

const login = async () => {
    try {
        await auth0.loginWithPopup({
            redirect_uri: window.location.origin
        });
        updateUI();
    } catch (error) {
        console.error('Login failed:', error);
    }
};

const logout = () => {
    auth0.logout({
        returnTo: window.location.origin
    });
};

const updateUI = async () => {
    const isAuthenticated = await auth0.isAuthenticated();
    if (isAuthenticated) {
        const user = await auth0.getUser();
        document.getElementById('profile').innerText = JSON.stringify(user);
    }
};

window.addEventListener('load', configureClient);
```

## Example 2: OAuth2 Authorization Code Flow

```python
# Auth0 OAuth2 implementation in Python
from fastapi import FastAPI, HTTPException
from authlib.integrations.starlette_client import OAuth
import os

app = FastAPI()
oauth = OAuth()

oauth.register(
    name='auth0',
    client_id=os.getenv('AUTH0_CLIENT_ID'),
    client_secret=os.getenv('AUTH0_CLIENT_SECRET'),
    server_metadata_url=f"https://{os.getenv('AUTH0_DOMAIN')}/.well-known/openid-configuration",
    client_kwargs={'scope': 'openid email profile'}
)

@app.get('/login')
async def login(request):
    redirect_uri = request.url_for('callback')
    return await oauth.auth0.authorize_redirect(request, redirect_uri)

@app.get('/callback')
async def callback(request):
    try:
        token = await oauth.auth0.authorize_access_token(request)
        user = token.get('userinfo')
        # Store user in database
        return {'user': user}
    except Exception as error:
        raise HTTPException(status_code=400, detail=str(error))

@app.post('/logout')
async def logout(request):
    request.session.clear()
    return {'status': 'logged out'}
```

## Example 3: Multi-Tenant Configuration

```typescript
// Auth0 multi-tenant setup with organization support
interface Auth0OrgConfig {
    domain: string;
    clientId: string;
    organizationId: string;
    redirectUri: string;
}

class Auth0MultiTenant {
    private clients: Map<string, Auth0Client> = new Map();

    async registerTenant(orgId: string, config: Auth0OrgConfig) {
        const client = new Auth0Client({
            domain: config.domain,
            clientId: config.clientId,
            cacheLocation: 'localstorage',
            authorizationParams: {
                organization: orgId
            }
        });

        this.clients.set(orgId, client);
    }

    async loginToTenant(orgId: string) {
        const client = this.clients.get(orgId);
        if (!client) {
            throw new Error(`Organization ${orgId} not configured`);
        }

        await client.loginWithPopup({
            organization: orgId
        });
    }

    async getTenantUser(orgId: string) {
        const client = this.clients.get(orgId);
        return client?.getUser();
    }
}
```

## Example 4: Role-Based Access Control (RBAC)

```python
# Auth0 RBAC implementation
from functools import wraps
from flask import request, abort

def require_role(required_roles):
    """Decorator to enforce role-based access"""
    def decorator(f):
        @wraps(f)
        def decorated_function(*args, **kwargs):
            token = request.headers.get('Authorization', '').split(' ')[1]
            decoded = jwt.decode(
                token,
                algorithms=['RS256'],
                audience=os.getenv('AUTH0_API_AUDIENCE'),
                issuer=f"https://{os.getenv('AUTH0_DOMAIN')}/"
            )

            user_roles = decoded.get('http://localhost/roles', [])
            if not any(role in required_roles for role in user_roles):
                abort(403)

            return f(*args, **kwargs)
        return decorated_function
    return decorator

@app.route('/admin')
@require_role(['admin', 'moderator'])
def admin_endpoint():
    return {'message': 'Admin access granted'}

@app.route('/user-data')
@require_role(['user', 'admin'])
def user_endpoint():
    return {'message': 'User access granted'}
```

## Example 5: API Protection with JWT Validation

```javascript
// Express.js middleware for Auth0 JWT validation
const express = require('express');
const jwt = require('express-jwt');
const jwksRsa = require('jwks-rsa');

const app = express();

const checkJwt = jwt({
    secret: jwksRsa.expressJwtSecret({
        cache: true,
        rateLimit: true,
        jwksRequestsPerMinute: 5,
        jwksUri: `https://${process.env.AUTH0_DOMAIN}/.well-known/jwks.json`
    }),
    audience: process.env.AUTH0_API_AUDIENCE,
    issuer: `https://${process.env.AUTH0_DOMAIN}/`,
    algorithms: ['RS256']
});

app.get('/api/protected', checkJwt, (req, res) => {
    res.json({
        message: 'This is a protected route',
        user: req.user
    });
});

app.get('/api/users/:id', checkJwt, (req, res) => {
    // Only allow users to access their own data
    if (req.user.sub === req.params.id || req.user.permissions.includes('read:users')) {
        res.json({ userId: req.params.id });
    } else {
        res.status(403).json({ error: 'Forbidden' });
    }
});
```

## Example 6: Custom Database Connection

```python
# Auth0 custom database authentication
class Auth0CustomDB:
    def __init__(self, db_connection):
        self.db = db_connection
        self.auth0_mgmt = Auth0ManagementAPI()

    async def login_script(self, email, password):
        """Custom database login script"""
        user = await self.db.find_user(email=email)

        if not user:
            return False

        # Verify password against custom database
        if not self.verify_password(password, user['password_hash']):
            return False

        # Sync user to Auth0 on successful login
        await self.auth0_mgmt.create_or_update_user({
            'email': email,
            'user_id': str(user['id']),
            'email_verified': True
        })

        return True

    async def signup_script(self, email, password, email_verified):
        """Custom database signup script"""
        if await self.db.user_exists(email=email):
            return False

        hashed_password = self.hash_password(password)
        user = await self.db.create_user(
            email=email,
            password_hash=hashed_password,
            email_verified=email_verified
        )

        return True

    def verify_password(self, password, hash):
        # Implement bcrypt verification
        pass

    def hash_password(self, password):
        # Implement bcrypt hashing
        pass
```

## Example 7: Social Identity Provider Integration

```javascript
// Auth0 social provider configuration
const socialProviders = {
    google: {
        strategy: 'google-oauth2',
        scope: ['profile', 'email'],
        options: {
            access_type: 'offline',
            prompt: 'consent'
        }
    },
    github: {
        strategy: 'github',
        scope: ['user:email'],
        options: {
            allow_signup: true
        }
    },
    linkedin: {
        strategy: 'linkedin-oauth2',
        scope: ['r_basicprofile', 'r_emailaddress'],
        options: {}
    }
};

class SocialAuth {
    async loginWithProvider(providerName) {
        const provider = socialProviders[providerName];
        if (!provider) {
            throw new Error(`Unknown provider: ${providerName}`);
        }

        const response = await fetch('/api/login-social', {
            method: 'POST',
            body: JSON.stringify({
                strategy: provider.strategy,
                scope: provider.scope
            })
        });

        return response.json();
    }
}
```

## Example 8: Email Verification and Passwordless

```python
# Auth0 passwordless authentication
from auth0.authentication import GetToken
from auth0.management import Auth0

class Auth0Passwordless:
    def __init__(self, domain, client_id, client_secret):
        self.domain = domain
        self.client_id = client_id
        self.client_secret = client_secret
        self.mgmt = Auth0(domain, client_secret, client_id=client_id)

    async def send_verification_email(self, email):
        """Send email verification code"""
        result = await self.mgmt.users.create({
            'email': email,
            'email_verified': False,
            'connection': 'email'
        })
        return result

    async def verify_email_code(self, email, code):
        """Verify email code and issue token"""
        response = await self.verify_code(email, code)
        return response

    async def send_passwordless_link(self, email):
        """Send passwordless login link"""
        return await self.send_verification_email(email)

    async def magic_link_callback(self, token):
        """Handle magic link callback"""
        # Validate token and authenticate user
        return self.authenticate_with_token(token)
```

## Example 9: User Metadata Management

```typescript
// Auth0 user metadata and app_metadata
interface UserMetadata {
    first_name?: string;
    last_name?: string;
    preferences?: {
        language: string;
        theme: 'light' | 'dark';
    };
}

interface AppMetadata {
    roles: string[];
    department?: string;
    subscription_tier?: string;
}

class Auth0UserMetadata {
    async updateUserMetadata(userId: string, metadata: UserMetadata) {
        const response = await fetch(
            `https://${this.domain}/api/v2/users/${userId}`,
            {
                method: 'PATCH',
                headers: {
                    'Authorization': `Bearer ${this.managementToken}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    user_metadata: metadata
                })
            }
        );
        return response.json();
    }

    async updateAppMetadata(userId: string, metadata: AppMetadata) {
        const response = await fetch(
            `https://${this.domain}/api/v2/users/${userId}`,
            {
                method: 'PATCH',
                headers: {
                    'Authorization': `Bearer ${this.managementToken}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    app_metadata: metadata
                })
            }
        );
        return response.json();
    }
}
```

## Example 10: Token Refresh and Rotation

```python
# Auth0 token refresh and rotation
import asyncio
from datetime import datetime, timedelta

class Auth0TokenManager:
    def __init__(self, client):
        self.client = client
        self.access_token = None
        self.refresh_token = None
        self.token_expires_at = None
        self.refresh_task = None

    async def refresh_tokens(self):
        """Refresh access token using refresh token"""
        if not self.refresh_token:
            raise ValueError('No refresh token available')

        response = await self.client.post('/oauth/token', {
            'client_id': self.client_id,
            'client_secret': self.client_secret,
            'grant_type': 'refresh_token',
            'refresh_token': self.refresh_token
        })

        self.access_token = response['access_token']
        self.token_expires_at = datetime.now() + timedelta(
            seconds=response['expires_in']
        )

        # Schedule next refresh
        self.schedule_refresh()

    def schedule_refresh(self):
        """Schedule token refresh before expiration"""
        if self.refresh_task:
            self.refresh_task.cancel()

        # Refresh 30 seconds before expiration
        refresh_delay = (
            (self.token_expires_at - datetime.now()).total_seconds() - 30
        )

        self.refresh_task = asyncio.create_task(
            asyncio.sleep(refresh_delay)
        ).then(self.refresh_tokens)
```

## Example 11: Rules and Hooks for Authentication

```javascript
// Auth0 Rules for custom authentication logic
// This runs during login and issues tokens

module.exports = function(user, context, callback) {
    // Add custom claims to token
    context.idToken['http://example.com/custom_claim'] = 'custom_value';

    // Add roles from database
    context.accessToken['roles'] = user.app_metadata?.roles || [];

    // Prevent login for inactive users
    if (user.app_metadata?.is_active === false) {
        return callback(new UnauthorizedError('Account is inactive'));
    }

    // Add location information
    context.idToken['http://example.com/location'] = context.request.ip;

    // Redirect to MFA if needed
    if (user.app_metadata?.requires_mfa) {
        context.multifactorAuthentication = {
            provider: 'google-authenticator',
            allow_remember_browser: false
        };
    }

    callback(null, user, context);
};
```

## Example 12: Logging and Monitoring

```python
# Auth0 monitoring and analytics
class Auth0Monitor:
    def __init__(self):
        self.logs_buffer = []
        self.metrics = {
            'login_success': 0,
            'login_failed': 0,
            'signup': 0,
            'mfa_success': 0,
            'mfa_failed': 0
        }

    async def fetch_auth0_logs(self, page=0, per_page=100):
        """Fetch Auth0 logs for monitoring"""
        response = await self.mgmt.log_entries.all(
            page=page,
            per_page=per_page,
            sort='date:-1'
        )
        return response

    def process_auth_event(self, event):
        """Process authentication event"""
        event_type = event.get('type')

        if event_type == 's':  # Successful login
            self.metrics['login_success'] += 1
        elif event_type == 'f':  # Failed login
            self.metrics['login_failed'] += 1
        elif event_type == 'ss':  # Successful signup
            self.metrics['signup'] += 1

    def get_metrics(self):
        """Get authentication metrics"""
        total_logins = (
            self.metrics['login_success'] + self.metrics['login_failed']
        )
        success_rate = (
            self.metrics['login_success'] / total_logins
            if total_logins > 0 else 0
        )

        return {
            **self.metrics,
            'success_rate': success_rate * 100
        }
```

## Example 13: Tenant Configuration

```yaml
# Auth0 Tenant Configuration
domain: example.auth0.com
clientId: ${AUTH0_CLIENT_ID}
clientSecret: ${AUTH0_CLIENT_SECRET}

# API Configuration
api:
  audience: https://api.example.com
  identifier: https://api.example.com

# Connection Configuration
connections:
  - type: Username-Password-Authentication
    options:
      password_policy: good
      requires_username: true
      password_history_size: 5
  - type: google-oauth2
  - type: github
  - type: facebook

# Application Settings
applications:
  - name: Web App
    type: spa
    allowedOrigins: ['http://localhost:3000']
    allowedLogoutUrls: ['http://localhost:3000']
  - name: Mobile App
    type: native
    tokenEndpointAuthMethod: none
```

---

**Context7 Integration**: Use `/auth0/auth0-js`, `/auth0/auth0-python`, `/auth0/auth0-spring-boot` for latest SDK documentation.

**Version**: 1.0.0 | **Last Updated**: 2025-11-22
