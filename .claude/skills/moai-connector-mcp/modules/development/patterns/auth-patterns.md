# Authentication Patterns

Security and authentication patterns for MCP servers.

---

## OAuth2 Authentication

### OAuth2 Provider Setup

```python
from fastmcp import FastMCP
from fastmcp.auth import OAuth2Provider
from datetime import datetime

server = FastMCP("oauth-server")

# Configure OAuth2
oauth = OAuth2Provider(
    authorize_url="https://auth.company.com/oauth/authorize",
    token_url="https://auth.company.com/oauth/token",
    scopes=["read:data", "write:data", "admin"]
)

# Protect tools with OAuth
@server.auth(oauth)
@server.tool()
def get_user_data(user_id: str) -> dict:
    """Get user data (requires OAuth token)."""
    # Token validation is automatic
    return fetch_user(user_id)

@server.auth(oauth)
@server.tool()
def update_user_data(user_id: str, data: dict) -> dict:
    """Update user (requires write:data scope)."""
    # In production, check scope: write:data
    updated = update_user(user_id, data)
    return {"status": "success", "user_id": user_id}

# Protect resources with OAuth
@server.auth(oauth)
@server.resource("user://{user_id}/private")
def get_private_data(user_id: str) -> dict:
    """Get private user data (OAuth required)."""
    return fetch_private(user_id)
```

---

## API Key Authentication

### API Key Validation

```python
from fastmcp.auth import APIKeyAuth

# Setup API key authentication
api_auth = APIKeyAuth(header="X-API-Key")

# Protect admin operations
@server.auth(api_auth)
@server.tool()
def create_admin_user(email: str, name: str) -> dict:
    """Create admin user (API key required)."""
    user_id = create_user(email, name, role="admin")
    return {"status": "success", "user_id": user_id}

@server.auth(api_auth)
@server.tool()
def delete_user(user_id: str, reason: str) -> dict:
    """Delete user (API key required)."""
    delete(user_id, reason)
    return {"status": "success", "deleted": user_id}

@server.auth(api_auth)
@server.resource("admin://system/stats")
def get_system_stats() -> dict:
    """System statistics (API key required)."""
    return {
        "total_users": count_users(),
        "active_sessions": count_sessions(),
        "api_calls_today": count_api_calls_today()
    }
```

---

## Multi-Auth Strategy

### Different Auth for Different Operations

```python
from fastmcp.auth import OAuth2Provider, APIKeyAuth

server = FastMCP("multi-auth-server")

# OAuth for user-facing operations
oauth = OAuth2Provider(
    authorize_url="https://auth.company.com/authorize",
    token_url="https://auth.company.com/token",
    scopes=["read:profile", "write:profile"]
)

# API Key for service-to-service
api_auth = APIKeyAuth(header="X-API-Key")

# User operations (OAuth)
@server.auth(oauth)
@server.tool()
def get_my_profile() -> dict:
    """Get current user profile (OAuth)."""
    return fetch_current_user()

@server.auth(oauth)
@server.tool()
def update_my_profile(data: dict) -> dict:
    """Update current user profile (OAuth)."""
    return update_user(get_current_user_id(), data)

# Admin operations (API Key)
@server.auth(api_auth)
@server.tool()
def list_all_users() -> list[dict]:
    """List all users (API key required)."""
    return fetch_all_users()

@server.auth(api_auth)
@server.tool()
def get_user_by_id(user_id: str) -> dict:
    """Get user by ID (API key required)."""
    return fetch_user(user_id)
```

---

## Custom Authentication

### Token Validation

```python
import jwt
from datetime import datetime, timedelta

class CustomTokenAuth:
    def __init__(self, secret_key: str):
        self.secret_key = secret_key

    def validate_token(self, token: str) -> dict:
        """Validate JWT token."""
        try:
            payload = jwt.decode(
                token,
                self.secret_key,
                algorithms=["HS256"]
            )

            # Check expiration
            if "exp" in payload:
                exp = datetime.fromtimestamp(payload["exp"])
                if datetime.now() > exp:
                    raise ValueError("Token expired")

            return payload

        except jwt.InvalidTokenError as e:
            raise ValueError(f"Invalid token: {str(e)}")

# Usage
auth = CustomTokenAuth(secret_key="your-secret-key")

@server.tool()
def protected_operation(token: str, param: str) -> dict:
    """Operation with custom token validation."""
    # Validate token
    payload = auth.validate_token(token)
    user_id = payload["user_id"]

    return execute_operation(user_id, param)
```

---

## Rate Limiting by User

### Per-User Rate Limiting

```python
from collections import defaultdict
from datetime import datetime, timedelta

class RateLimiter:
    def __init__(self, max_requests: int, window_seconds: int):
        self.max_requests = max_requests
        self.window = window_seconds
        self.requests = defaultdict(list)

    def is_allowed(self, user_id: str) -> bool:
        """Check if user is within rate limit."""
        now = datetime.now()
        cutoff = now - timedelta(seconds=self.window)

        # Remove old requests
        self.requests[user_id] = [
            ts for ts in self.requests[user_id]
            if ts > cutoff
        ]

        # Check limit
        if len(self.requests[user_id]) < self.max_requests:
            self.requests[user_id].append(now)
            return True

        return False

limiter = RateLimiter(max_requests=100, window_seconds=60)

@server.tool()
def rate_limited_operation(user_id: str, param: str) -> dict:
    """Operation with rate limiting."""
    if not limiter.is_allowed(user_id):
        raise ValueError(
            f"Rate limit exceeded for user {user_id}. "
            f"Max 100 requests per minute."
        )

    return execute_operation(user_id, param)
```

---

## Environment-Based Credentials

### Secure Credential Management

```python
import os
from functools import lru_cache

class CredentialManager:
    @staticmethod
    @lru_cache(maxsize=1)
    def get_secret(key: str) -> str:
        """Get secret from environment or vault."""
        secret = os.getenv(key)

        if not secret:
            raise ValueError(
                f"Required secret not configured: {key}. "
                f"Set {key} environment variable."
            )

        return secret

    @staticmethod
    def get_api_key() -> str:
        """Get API key."""
        return CredentialManager.get_secret("API_KEY")

    @staticmethod
    def get_oauth_client_id() -> str:
        """Get OAuth client ID."""
        return CredentialManager.get_secret("OAUTH_CLIENT_ID")

    @staticmethod
    def get_oauth_client_secret() -> str:
        """Get OAuth client secret."""
        return CredentialManager.get_secret("OAUTH_CLIENT_SECRET")

# Usage
try:
    api_key = CredentialManager.get_api_key()
except ValueError as e:
    print(f"Configuration error: {e}")
    exit(1)
```

---

## Scope-Based Authorization

### Check Scopes in Protected Operations

```python
from typing import Optional

def has_scope(token_scopes: list[str], required_scope: str) -> bool:
    """Check if token has required scope."""
    return required_scope in token_scopes

@server.auth(oauth)
@server.tool()
def delete_user(user_id: str, reason: str) -> dict:
    """Delete user (requires admin:write scope)."""
    # In production, extract and check scopes from OAuth token
    token_scopes = extract_scopes_from_context()

    if not has_scope(token_scopes, "admin:write"):
        raise PermissionError(
            "Insufficient permissions. "
            "Required scope: admin:write"
        )

    perform_deletion(user_id, reason)
    return {"status": "success", "deleted": user_id}

@server.auth(oauth)
@server.tool()
def read_sensitive_data(data_id: str) -> dict:
    """Read sensitive data (requires data:read scope)."""
    token_scopes = extract_scopes_from_context()

    if not has_scope(token_scopes, "data:read"):
        raise PermissionError(
            "Insufficient permissions. "
            "Required scope: data:read"
        )

    return fetch_data(data_id)
```

---

## Best Practices

✅ **Security**:
- Never hardcode credentials
- Use environment variables for secrets
- Validate all tokens
- Check scopes for operations
- Implement rate limiting
- Log authentication events
- Use HTTPS for token transmission

✅ **Implementation**:
- Separate auth mechanisms by use case
- Provide clear error messages
- Document required scopes
- Implement token refresh
- Monitor for suspicious activity

❌ **Avoid**:
- Storing passwords in plain text
- Ignoring token expiration
- Bypassing scope checks
- Mixing auth mechanisms carelessly
- Logging sensitive data

---

**Last Updated**: 2025-11-27
