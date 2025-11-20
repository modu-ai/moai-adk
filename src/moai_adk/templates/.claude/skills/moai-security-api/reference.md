# API Security Reference

Complete reference guide for OWASP API Security Top 10 and advanced patterns.

---

## OWASP API Security Top 10 (2023) - Detailed

### 1. Broken Object Level Authorization (BOLA)

**Description**: API exposes endpoints that handle object identifiers without proper authorization checks.

**Attack Example**:

```http
GET /api/users/123/orders
Authorization: Bearer <user_456_token>

# User 456 shouldn't see User 123's orders!
```

**Mitigation**:

```python
@app.get("/users/{user_id}/orders")
async def get_user_orders(user_id: int, current_user: User = Depends(get_current_user)):
    # Verify user owns this resource
    if current_user.id != user_id and not current_user.is_admin:
        raise HTTPException(403, "Access denied")

    return await Order.get_by_user(user_id)
```

**Prevention Checklist**:

- [ ] Validate user owns requested resource
- [ ] Implement object-level permission checks
- [ ] Use UUIDs instead of sequential IDs
- [ ] Log unauthorized access attempts

---

### 2. Broken Authentication

**Description**: Weak authentication mechanisms allowing attackers to compromise tokens or credentials.

**Common Issues**:

- Weak password policies
- No MFA
- Long-lived tokens
- Credentials in URLs

**Mitigation**:

```python
# Strong password policy
PASSWORD_REQUIREMENTS = {
    "min_length": 12,
    "require_uppercase": True,
    "require_lowercase": True,
    "require_digits": True,
    "require_special": True,
}

# Short-lived JWT
ACCESS_TOKEN_EXPIRE_MINUTES = 15  # Not 24 hours!
REFRESH_TOKEN_EXPIRE_DAYS = 7

# MFA enforcement
@app.post("/login")
async def login(username: str, password: str):
    user = await authenticate(username, password)
    if user.mfa_enabled:
        otp_token = generate_otp_challenge(user)
        return {"requires_mfa": True, "otp_token": otp_token}
    # ... issue tokens
```

---

### 3. Broken Object Property Level Authorization

**Description**: API returns more data than necessary, exposing sensitive fields.

**Attack Example**:

```json
// Response exposes internal fields
{
  "id": 123,
  "name": "John Doe",
  "email": "john@example.com",
  "password_hash": "$2b$12$...", // ❌ Should not be exposed
  "is_admin": true, // ❌ Internal field
  "created_at": "2025-01-01"
}
```

**Mitigation**:

```python
from pydantic import BaseModel

class UserPublic(BaseModel):
    """Public user representation - excludes sensitive fields."""
    id: int
    name: str
    email: str
    created_at: datetime

    class Config:
        from_attributes = True

@app.get("/users/{user_id}", response_model=UserPublic)
async def get_user(user_id: int):
    user = await User.get(user_id)
    return user  # Pydantic automatically filters fields
```

---

### 4. Unrestricted Resource Consumption

**Description**: No limits on API usage, leading to DoS or excessive billing.

**Attack Scenarios**:

- No pagination (fetch 1M records)
- No rate limiting
- Expensive operations without throttling
- Large file uploads

**Mitigation**:

```python
from fastapi import Query

@app.get("/users")
async def list_users(
    page: int = Query(1, ge=1),
    limit: int = Query(20, ge=1, le=100)  # Max 100 per request
):
    offset = (page - 1) * limit
    users = await User.get_paginated(offset=offset, limit=limit)
    return {
        "users": users,
        "page": page,
        "limit": limit,
        "total": await User.count()
    }

# File upload limits
@app.post("/upload")
async def upload_file(file: UploadFile):
    MAX_FILE_SIZE = 10 * 1024 * 1024  # 10MB

    size = 0
    chunks = []
    async for chunk in file.stream():
        size += len(chunk)
        if size > MAX_FILE_SIZE:
            raise HTTPException(413, "File too large")
        chunks.append(chunk)

    # Process file...
```

---

### 5. Broken Function Level Authorization

**Description**: Regular users can access admin/privileged functions.

**Attack Example**:

```http
DELETE /api/admin/users/123
Authorization: Bearer <regular_user_token>

# Regular user shouldn't access admin endpoints!
```

**Mitigation**:

```python
# Route-level protection
admin_router = APIRouter(prefix="/admin", dependencies=[Depends(require_admin)])

@admin_router.delete("/users/{user_id}")
async def delete_user(user_id: int):
    # Only reachable if user is admin
    await User.delete(user_id)
    return {"status": "deleted"}

# Function-level check
def require_admin(current_user: User = Depends(get_current_user)):
    if not current_user.is_admin:
        raise HTTPException(403, "Admin access required")
    return current_user
```

---

### 6. Unrestricted Access to Sensitive Business Flows

**Description**: Lack of flow control allows automated attacks (credential stuffing, scalping).

**Scenarios**:

- Mass account creation
- Automated ticket purchasing
- Credential stuffing attacks

**Mitigation**:

```python
from slowapi import Limiter
from slowapi.util import get_remote_address

limiter = Limiter(key_func=get_remote_address)

# Stricter limits on sensitive endpoints
@app.post("/register")
@limiter.limit("5/hour")  # Only 5 registrations per hour per IP
async def register(user_data: UserCreate):
    # Add CAPTCHA verification
    if not verify_captcha(user_data.captcha_token):
        raise HTTPException(400, "Invalid CAPTCHA")

    # Detect suspicious patterns
    if await is_suspicious_activity(user_data.email):
        await send_alert_to_security_team()
        raise HTTPException(429, "Too many requests")

    # Proceed with registration...
```

---

### 7. Server Side Request Forgery (SSRF)

**Description**: Attacker tricks server into making requests to internal services.

**Attack Example**:

```http
POST /api/fetch-url
{
  "url": "http://169.254.169.254/latest/meta-data/iam/security-credentials/"
}

# Accesses AWS metadata service!
```

**Mitigation**:

```python
import ipaddress
from urllib.parse import urlparse

BLOCKED_IP_RANGES = [
    ipaddress.ip_network("10.0.0.0/8"),      # Private
    ipaddress.ip_network("172.16.0.0/12"),   # Private
    ipaddress.ip_network("192.168.0.0/16"),  # Private
    ipaddress.ip_network("169.254.0.0/16"),  # Link-local (AWS metadata)
    ipaddress.ip_network("127.0.0.0/8"),     # Loopback
]

ALLOWED_DOMAINS = ["api.trusted-partner.com", "cdn.example.com"]

async def safe_fetch_url(url: str):
    """Safely fetch URL with SSRF protection."""
    # Parse and validate URL
    parsed = urlparse(url)

    # Check protocol
    if parsed.scheme not in ["http", "https"]:
        raise ValueError("Invalid protocol")

    # Check domain allowlist
    if parsed.hostname not in ALLOWED_DOMAINS:
        raise ValueError("Domain not allowed")

    # Resolve IP and check against blocklist
    import socket
    ip = socket.gethostbyname(parsed.hostname)
    ip_obj = ipaddress.ip_address(ip)

    for blocked_range in BLOCKED_IP_RANGES:
        if ip_obj in blocked_range:
            raise ValueError("IP address blocked")

    # Fetch with timeout
    async with httpx.AsyncClient(timeout=5.0) as client:
        response = await client.get(url)
        return response.text
```

---

### 8. Security Misconfiguration

**Common Issues**:

- Debug mode enabled in production
- Default credentials
- Verbose error messages
- Missing security headers

**Mitigation**:

```python
# Production configuration
app = FastAPI(
    debug=False,  # Never True in production
    docs_url=None,  # Disable Swagger in production
    redoc_url=None,
)

# Security headers middleware
from fastapi.middleware.trustedhost import TrustedHostMiddleware
from starlette.middleware.httpsredirect import HTTPSRedirectMiddleware

app.add_middleware(TrustedHostMiddleware, allowed_hosts=["example.com", "*.example.com"])
app.add_middleware(HTTPSRedirectMiddleware)

@app.middleware("http")
async def add_security_headers(request, call_next):
    response = await call_next(request)
    response.headers["X-Content-Type-Options"] = "nosniff"
    response.headers["X-Frame-Options"] = "DENY"
    response.headers["X-XSS-Protection"] = "1; mode=block"
    response.headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains"
    return response

# Generic error responses
@app.exception_handler(Exception)
async def generic_exception_handler(request, exc):
    # Log detailed error internally
    logger.error(f"Unhandled exception: {exc}", exc_info=True)

    # Return generic message to client
    return JSONResponse(
        status_code=500,
        content={"detail": "Internal server error"}
    )
```

---

### 9. Improper Inventory Management

**Issues**:

- Undocumented endpoints
- Deprecated versions still active
- No API versioning

**Mitigation**:

```python
# API versioning
from fastapi import APIRouter

v1_router = APIRouter(prefix="/v1")
v2_router = APIRouter(prefix="/v2")

# Deprecation warnings
@v1_router.get("/users", deprecated=True)
async def get_users_v1():
    return {
        "warning": "This endpoint is deprecated. Use /v2/users instead.",
        "users": []
    }

@v2_router.get("/users")
async def get_users_v2():
    return {"users": []}

app.include_router(v1_router)
app.include_router(v2_router)

# OpenAPI documentation with versions
app = FastAPI(
    title="My API",
    version="2.0.0",
    description="API v2 - v1 deprecated, will be removed 2026-01-01"
)
```

---

### 10. Unsafe Consumption of APIs

**Description**: Blindly trusting third-party API responses.

**Mitigation**:

```python
from pydantic import BaseModel, ValidationError

class ThirdPartyResponse(BaseModel):
    """Validate external API responses."""
    user_id: int
    name: str
    email: str

async def consume_third_party_api(url: str):
    async with httpx.AsyncClient(timeout=10.0) as client:
        try:
            response = await client.get(url)
            response.raise_for_status()

            # Validate response structure
            data = ThirdPartyResponse(**response.json())

            # Additional business logic validation
            if data.user_id <= 0:
                raise ValueError("Invalid user_id")

            return data

        except httpx.HTTPError as e:
            logger.error(f"Third-party API error: {e}")
            raise HTTPException(502, "External service unavailable")
        except ValidationError as e:
            logger.error(f"Invalid third-party response: {e}")
            raise HTTPException(502, "Invalid external response")
```

---

## Advanced Security Patterns

### API Key Rotation

```python
class APIKey(BaseModel):
    key: str
    created_at: datetime
    expires_at: datetime
    last_used: datetime | None

async def rotate_api_key(old_key: str) -> str:
    """Generate new API key while deprecating old one."""
    # Mark old key as deprecated (grace period: 30 days)
    await APIKey.update(
        key=old_key,
        status="deprecated",
        expires_at=datetime.utcnow() + timedelta(days=30)
    )

    # Generate new key
    new_key = secrets.token_urlsafe(32)
    await APIKey.create(
        key=new_key,
        created_at=datetime.utcnow(),
        expires_at=datetime.utcnow() + timedelta(days=365)
    )

    return new_key
```

### Request Signing

```python
import hmac
import hashlib

def sign_request(payload: dict, secret: str) -> str:
    """Sign request payload with HMAC."""
    message = json.dumps(payload, sort_keys=True).encode()
    signature = hmac.new(
        secret.encode(),
        message,
        hashlib.sha256
    ).hexdigest()
    return signature

def verify_request_signature(payload: dict, signature: str, secret: str) -> bool:
    """Verify request signature."""
    expected_signature = sign_request(payload, secret)
    return hmac.compare_digest(expected_signature, signature)
```

---

## Security Testing

### Automated Security Scanning

```bash
# Install security tools
pip install bandit safety

# Scan for security issues
bandit -r app/
safety check --json

# API security testing
pip install owasp-zap-api-python

# Run ZAP scan
zap-cli quick-scan --self-contained http://localhost:8000
```

---

**See also**: [SKILL.md](./SKILL.md) for overview and quick start guide.
