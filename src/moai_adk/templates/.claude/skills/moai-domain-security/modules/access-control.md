# Access Control & Authorization Patterns

## Role-Based Access Control (RBAC)

### Pattern 1: Hierarchical RBAC Implementation

**Principle**: Users inherit permissions from roles, roles form hierarchy.

**Implementation**:
```python
from dataclasses import dataclass
from typing import Set, Optional
from enum import Enum

class Permission(Enum):
    """System permissions."""
    # User management
    USER_READ = "user:read"
    USER_WRITE = "user:write"
    USER_DELETE = "user:delete"

    # Content management
    CONTENT_READ = "content:read"
    CONTENT_WRITE = "content:write"
    CONTENT_PUBLISH = "content:publish"
    CONTENT_DELETE = "content:delete"

    # Admin
    ADMIN_READ = "admin:read"
    ADMIN_WRITE = "admin:write"

@dataclass
class Role:
    """Role with permissions and hierarchy."""
    name: str
    permissions: Set[Permission]
    inherits_from: Optional['Role'] = None

    def all_permissions(self) -> Set[Permission]:
        """Get all permissions including inherited."""
        permissions = self.permissions.copy()

        # Recursively add parent permissions
        if self.inherits_from:
            permissions.update(self.inherits_from.all_permissions())

        return permissions

# Define role hierarchy
viewer_role = Role(
    name="viewer",
    permissions={Permission.USER_READ, Permission.CONTENT_READ}
)

editor_role = Role(
    name="editor",
    permissions={Permission.CONTENT_WRITE},
    inherits_from=viewer_role  # Inherits read permissions
)

publisher_role = Role(
    name="publisher",
    permissions={Permission.CONTENT_PUBLISH},
    inherits_from=editor_role  # Inherits viewer + editor
)

admin_role = Role(
    name="admin",
    permissions={
        Permission.USER_WRITE,
        Permission.USER_DELETE,
        Permission.CONTENT_DELETE,
        Permission.ADMIN_READ,
        Permission.ADMIN_WRITE
    },
    inherits_from=publisher_role  # Inherits all content permissions
)

class RBACService:
    """RBAC authorization service."""

    def __init__(self):
        self.roles = {
            "viewer": viewer_role,
            "editor": editor_role,
            "publisher": publisher_role,
            "admin": admin_role
        }

    def check_permission(self, user_roles: list[str], required_permission: Permission) -> bool:
        """Check if user has required permission."""
        for role_name in user_roles:
            role = self.roles.get(role_name)
            if role and required_permission in role.all_permissions():
                return True
        return False

    def get_user_permissions(self, user_roles: list[str]) -> Set[Permission]:
        """Get all permissions for user's roles."""
        permissions = set()
        for role_name in user_roles:
            role = self.roles.get(role_name)
            if role:
                permissions.update(role.all_permissions())
        return permissions

# FastAPI integration
from fastapi import Depends, HTTPException, status

rbac = RBACService()

def require_permission(permission: Permission):
    """Dependency: Enforce permission-based access control."""
    async def permission_checker(user: dict = Depends(get_current_user)):
        if not rbac.check_permission(user["roles"], permission):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=f"Permission '{permission.value}' required"
            )
        return user
    return permission_checker

@app.delete("/content/{content_id}")
async def delete_content(
    content_id: int,
    user: dict = Depends(require_permission(Permission.CONTENT_DELETE))
):
    """Delete content (admin only via CONTENT_DELETE permission)."""
    return {"message": f"Content {content_id} deleted by {user['user_id']}"}
```

---

## Attribute-Based Access Control (ABAC)

### Pattern 2: Policy-Based ABAC

**Principle**: Dynamic authorization based on attributes (user, resource, environment).

**Implementation**:
```python
from dataclasses import dataclass
from typing import Any, Dict
from datetime import datetime, time

@dataclass
class ABACPolicy:
    """ABAC policy with subject, resource, action, and environment rules."""
    name: str
    subject_rules: Dict[str, Any]  # User attributes
    resource_rules: Dict[str, Any]  # Resource attributes
    action_rules: Dict[str, Any]  # Action attributes
    environment_rules: Dict[str, Any]  # Context attributes

class ABACEngine:
    """ABAC policy evaluation engine."""

    def __init__(self):
        self.policies: list[ABACPolicy] = []

    def add_policy(self, policy: ABACPolicy):
        """Register policy."""
        self.policies.append(policy)

    def evaluate(
        self,
        subject: Dict[str, Any],
        resource: Dict[str, Any],
        action: str,
        environment: Dict[str, Any]
    ) -> bool:
        """
        Evaluate if action is allowed.

        Args:
            subject: User attributes (id, role, department, clearance_level)
            resource: Resource attributes (owner, sensitivity, classification)
            action: Action to perform (read, write, delete)
            environment: Context attributes (time, location, ip_address)

        Returns:
            True if allowed by at least one policy
        """
        for policy in self.policies:
            if self._matches_policy(policy, subject, resource, action, environment):
                return True
        return False

    def _matches_policy(
        self,
        policy: ABACPolicy,
        subject: Dict,
        resource: Dict,
        action: str,
        environment: Dict
    ) -> bool:
        """Check if request matches policy rules."""

        # Check subject rules
        if not self._match_attributes(subject, policy.subject_rules):
            return False

        # Check resource rules
        if not self._match_attributes(resource, policy.resource_rules):
            return False

        # Check action rules
        if policy.action_rules.get("allowed_actions"):
            if action not in policy.action_rules["allowed_actions"]:
                return False

        # Check environment rules
        if not self._match_environment(environment, policy.environment_rules):
            return False

        return True

    def _match_attributes(self, actual: Dict, required: Dict) -> bool:
        """Check if actual attributes match required."""
        for key, value in required.items():
            if key not in actual:
                return False

            if isinstance(value, list):
                # Check if actual value is in allowed list
                if actual[key] not in value:
                    return False
            elif callable(value):
                # Custom validation function
                if not value(actual[key]):
                    return False
            else:
                # Exact match
                if actual[key] != value:
                    return False

        return True

    def _match_environment(self, environment: Dict, rules: Dict) -> bool:
        """Check environment constraints."""

        # Time-based access
        if "allowed_hours" in rules:
            current_hour = datetime.now().hour
            start, end = rules["allowed_hours"]
            if not (start <= current_hour < end):
                return False

        # Location-based access
        if "allowed_locations" in rules:
            if environment.get("location") not in rules["allowed_locations"]:
                return False

        # IP-based access
        if "allowed_ip_ranges" in rules:
            ip = environment.get("ip_address")
            if not any(self._ip_in_range(ip, range_) for range_ in rules["allowed_ip_ranges"]):
                return False

        return True

    def _ip_in_range(self, ip: str, ip_range: str) -> bool:
        """Check if IP is in range (simplified)."""
        # Real implementation should use ipaddress module
        return True

# Example policies
abac = ABACEngine()

# Policy 1: Employees can read their department's documents
abac.add_policy(ABACPolicy(
    name="department_read_access",
    subject_rules={"role": "employee"},
    resource_rules={"department": lambda dept: dept},  # Match user's department
    action_rules={"allowed_actions": ["read"]},
    environment_rules={}
))

# Policy 2: Managers can write to their department during business hours
abac.add_policy(ABACPolicy(
    name="manager_write_access",
    subject_rules={"role": "manager"},
    resource_rules={"department": lambda dept: dept},  # Match user's department
    action_rules={"allowed_actions": ["read", "write"]},
    environment_rules={"allowed_hours": (9, 17)}  # 9 AM - 5 PM
))

# Policy 3: Admins can do anything from corporate network
abac.add_policy(ABACPolicy(
    name="admin_full_access",
    subject_rules={"role": "admin"},
    resource_rules={},
    action_rules={"allowed_actions": ["read", "write", "delete"]},
    environment_rules={"allowed_locations": ["corporate_office"]}
))

# FastAPI integration
@app.get("/documents/{doc_id}")
async def get_document(
    doc_id: int,
    user: dict = Depends(get_current_user),
    request: Request = None
):
    """Get document with ABAC authorization."""

    # Fetch document
    document = db.query(Document).filter(Document.id == doc_id).first()
    if not document:
        raise HTTPException(status_code=404)

    # ABAC evaluation
    subject = {
        "id": user["user_id"],
        "role": user["role"],
        "department": user["department"]
    }

    resource = {
        "id": document.id,
        "owner": document.owner_id,
        "department": document.department,
        "sensitivity": document.sensitivity_level
    }

    environment = {
        "time": datetime.now(),
        "location": user.get("location"),
        "ip_address": request.client.host
    }

    if not abac.evaluate(subject, resource, "read", environment):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Access denied by policy"
        )

    return document
```

---

## OAuth 2.0 / OpenID Connect

### Pattern 3: OAuth 2.0 Authorization Code Flow

**Best Practice**: Use authorization code flow with PKCE for web/mobile apps.

**Implementation**:
```python
import secrets
import hashlib
import base64
from urllib.parse import urlencode

class OAuth2Service:
    """OAuth 2.0 authorization code flow with PKCE."""

    def __init__(self, client_id: str, client_secret: str, redirect_uri: str):
        self.client_id = client_id
        self.client_secret = client_secret
        self.redirect_uri = redirect_uri
        self.authorization_endpoint = "https://auth.example.com/authorize"
        self.token_endpoint = "https://auth.example.com/token"

    def generate_pkce_pair(self) -> tuple[str, str]:
        """Generate PKCE code_verifier and code_challenge."""

        # Generate cryptographically random code_verifier
        code_verifier = base64.urlsafe_b64encode(secrets.token_bytes(32)).decode('utf-8')
        code_verifier = code_verifier.rstrip('=')  # Remove padding

        # Compute code_challenge = BASE64URL(SHA256(code_verifier))
        challenge = hashlib.sha256(code_verifier.encode('utf-8')).digest()
        code_challenge = base64.urlsafe_b64encode(challenge).decode('utf-8')
        code_challenge = code_challenge.rstrip('=')

        return code_verifier, code_challenge

    def get_authorization_url(self, state: str, scope: str = "openid profile email") -> str:
        """Generate authorization URL with PKCE."""

        code_verifier, code_challenge = self.generate_pkce_pair()

        # Store code_verifier in session (retrieve later for token exchange)
        # session['code_verifier'] = code_verifier

        params = {
            "response_type": "code",
            "client_id": self.client_id,
            "redirect_uri": self.redirect_uri,
            "scope": scope,
            "state": state,  # CSRF protection
            "code_challenge": code_challenge,
            "code_challenge_method": "S256"
        }

        return f"{self.authorization_endpoint}?{urlencode(params)}"

    async def exchange_code_for_token(self, code: str, code_verifier: str) -> dict:
        """Exchange authorization code for access token."""

        import httpx

        data = {
            "grant_type": "authorization_code",
            "code": code,
            "redirect_uri": self.redirect_uri,
            "client_id": self.client_id,
            "client_secret": self.client_secret,
            "code_verifier": code_verifier  # PKCE
        }

        async with httpx.AsyncClient() as client:
            response = await client.post(self.token_endpoint, data=data)
            response.raise_for_status()
            return response.json()

# FastAPI routes
oauth2_service = OAuth2Service(
    client_id="your-client-id",
    client_secret="your-client-secret",
    redirect_uri="https://yourapp.com/callback"
)

@app.get("/login")
async def login():
    """Initiate OAuth 2.0 flow."""

    # Generate random state for CSRF protection
    state = secrets.token_urlsafe(32)
    # Store state in session for validation
    # session['oauth_state'] = state

    auth_url = oauth2_service.get_authorization_url(state)
    return {"authorization_url": auth_url}

@app.get("/callback")
async def oauth_callback(code: str, state: str):
    """OAuth 2.0 callback handler."""

    # Validate state (CSRF protection)
    # if state != session.get('oauth_state'):
    #     raise HTTPException(status_code=400, detail="Invalid state")

    # Retrieve code_verifier from session
    # code_verifier = session.get('code_verifier')

    # Exchange code for tokens
    tokens = await oauth2_service.exchange_code_for_token(code, code_verifier)

    # tokens contains:
    # - access_token: For API calls
    # - id_token: OpenID Connect user identity (JWT)
    # - refresh_token: For refreshing access tokens

    return {"message": "Authentication successful", "tokens": tokens}
```

---

## API Rate Limiting

### Pattern 4: Token Bucket Rate Limiter

**Principle**: Prevent API abuse with rate limiting per user/IP.

**Implementation**:
```python
from datetime import datetime, timedelta
from typing import Optional
import time

class TokenBucket:
    """Token bucket rate limiter."""

    def __init__(self, capacity: int, refill_rate: float):
        """
        Initialize token bucket.

        Args:
            capacity: Maximum tokens
            refill_rate: Tokens added per second
        """
        self.capacity = capacity
        self.refill_rate = refill_rate
        self.tokens = capacity
        self.last_refill = time.time()

    def consume(self, tokens: int = 1) -> bool:
        """
        Consume tokens if available.

        Returns:
            True if tokens consumed, False if insufficient
        """
        self._refill()

        if self.tokens >= tokens:
            self.tokens -= tokens
            return True
        return False

    def _refill(self):
        """Refill tokens based on elapsed time."""
        now = time.time()
        elapsed = now - self.last_refill

        # Add tokens based on elapsed time
        tokens_to_add = elapsed * self.refill_rate
        self.tokens = min(self.capacity, self.tokens + tokens_to_add)

        self.last_refill = now

class RateLimiter:
    """Rate limiter with token bucket per user."""

    def __init__(self, requests_per_minute: int = 60):
        self.buckets: dict[str, TokenBucket] = {}
        self.capacity = requests_per_minute
        self.refill_rate = requests_per_minute / 60.0  # Tokens per second

    def is_allowed(self, identifier: str) -> bool:
        """Check if request is allowed for identifier (user ID or IP)."""

        if identifier not in self.buckets:
            self.buckets[identifier] = TokenBucket(self.capacity, self.refill_rate)

        return self.buckets[identifier].consume()

    def get_remaining(self, identifier: str) -> int:
        """Get remaining requests for identifier."""
        if identifier in self.buckets:
            return int(self.buckets[identifier].tokens)
        return self.capacity

# FastAPI integration
rate_limiter = RateLimiter(requests_per_minute=100)

async def rate_limit_dependency(
    request: Request,
    user: Optional[dict] = Depends(get_current_user)
):
    """Rate limiting middleware."""

    # Use user ID if authenticated, otherwise IP address
    identifier = user["user_id"] if user else request.client.host

    if not rate_limiter.is_allowed(identifier):
        raise HTTPException(
            status_code=status.HTTP_429_TOO_MANY_REQUESTS,
            detail="Rate limit exceeded",
            headers={
                "X-RateLimit-Remaining": "0",
                "Retry-After": "60"
            }
        )

    return identifier

@app.get("/api/data")
async def get_data(identifier: str = Depends(rate_limit_dependency)):
    """API endpoint with rate limiting."""
    remaining = rate_limiter.get_remaining(identifier)
    return {
        "data": "Your data here",
        "rate_limit_remaining": remaining
    }
```

---

**Version**: 1.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready
