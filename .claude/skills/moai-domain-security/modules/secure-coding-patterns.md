# Secure Coding Patterns

## Input Validation & Sanitization

### Pattern 1: Type-Safe Input Validation

**Principle**: Never trust user input. Validate type, format, range, and business logic.

**Implementation (FastAPI + Pydantic)**:
```python
from pydantic import BaseModel, Field, validator, EmailStr
from typing import Optional
from datetime import datetime

class UserRegistration(BaseModel):
    """Type-safe user registration with comprehensive validation."""

    username: str = Field(
        min_length=3,
        max_length=20,
        regex=r'^[a-zA-Z0-9_]+$',
        description="Alphanumeric username only"
    )
    email: EmailStr
    password: str = Field(min_length=12, max_length=128)
    age: int = Field(ge=13, le=120, description="Must be 13+ years old")
    referral_code: Optional[str] = Field(None, max_length=10)

    @validator('username')
    def username_must_not_be_reserved(cls, v):
        """Block reserved usernames."""
        reserved = {'admin', 'root', 'system', 'api', 'support'}
        if v.lower() in reserved:
            raise ValueError('Username is reserved')
        return v

    @validator('password')
    def password_strength(cls, v):
        """Enforce strong password requirements."""
        if not any(c.isupper() for c in v):
            raise ValueError('Password must contain uppercase letter')
        if not any(c.islower() for c in v):
            raise ValueError('Password must contain lowercase letter')
        if not any(c.isdigit() for c in v):
            raise ValueError('Password must contain digit')
        if not any(c in '!@#$%^&*()_+-=[]{}|;:,.<>?' for c in v):
            raise ValueError('Password must contain special character')
        return v

# Usage
from fastapi import FastAPI, HTTPException

app = FastAPI()

@app.post("/register")
async def register_user(user: UserRegistration):
    """Registration with automatic validation."""
    # Pydantic validates automatically before this code runs
    # If validation fails, FastAPI returns 422 Unprocessable Entity

    # Business logic here
    return {"message": "User registered successfully"}
```

### Pattern 2: Allowlist-Based Validation

**Principle**: Reject by default, allow explicitly defined inputs only.

**Implementation**:
```python
from enum import Enum
from typing import Literal

class FileType(str, Enum):
    """Allowlist of permitted file types."""
    PDF = "pdf"
    JPEG = "jpeg"
    PNG = "png"
    DOCX = "docx"

class SortOrder(str, Enum):
    ASC = "asc"
    DESC = "desc"

class SearchQuery(BaseModel):
    """Allowlist-based query parameters."""

    query: str = Field(max_length=200, regex=r'^[a-zA-Z0-9 \-_]+$')
    file_type: FileType  # Only accepts enum values
    sort_by: Literal["created_at", "updated_at", "name"]  # Only these 3 fields
    sort_order: SortOrder = SortOrder.DESC
    page: int = Field(ge=1, le=1000, default=1)
    per_page: int = Field(ge=10, le=100, default=20)

@app.get("/search")
async def search_files(query: SearchQuery):
    """Search with strict allowlist validation."""
    # All inputs validated against allowlist
    # No SQL injection risk from sort_by field
    return {"results": []}
```

---

## Authentication & Authorization Patterns

### Pattern 3: Secure JWT Implementation

**Best Practices**: Short-lived access tokens, secure refresh mechanism, proper validation.

**Implementation**:
```python
from datetime import datetime, timedelta
import jwt
from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials

security = HTTPBearer()

class JWTService:
    """Production-ready JWT service."""

    def __init__(self):
        self.access_secret = "your-access-secret-256-bits"  # Use env var
        self.refresh_secret = "your-refresh-secret-256-bits"  # Different key
        self.algorithm = "HS256"
        self.access_token_expire = timedelta(minutes=15)
        self.refresh_token_expire = timedelta(days=7)

    def create_access_token(self, user_id: str, roles: list[str]) -> str:
        """Generate short-lived access token."""
        payload = {
            "sub": user_id,
            "roles": roles,
            "type": "access",
            "iat": datetime.utcnow(),
            "exp": datetime.utcnow() + self.access_token_expire
        }
        return jwt.encode(payload, self.access_secret, algorithm=self.algorithm)

    def create_refresh_token(self, user_id: str) -> str:
        """Generate long-lived refresh token."""
        payload = {
            "sub": user_id,
            "type": "refresh",
            "iat": datetime.utcnow(),
            "exp": datetime.utcnow() + self.refresh_token_expire
        }
        return jwt.encode(payload, self.refresh_secret, algorithm=self.algorithm)

    def verify_access_token(self, token: str) -> dict:
        """Validate access token with comprehensive checks."""
        try:
            payload = jwt.decode(
                token,
                self.access_secret,
                algorithms=[self.algorithm],
                options={
                    "verify_signature": True,
                    "verify_exp": True,
                    "verify_iat": True,
                    "require": ["sub", "roles", "type", "exp"]
                }
            )

            # Verify token type
            if payload.get("type") != "access":
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Invalid token type"
                )

            return payload

        except jwt.ExpiredSignatureError:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Token expired"
            )
        except jwt.InvalidTokenError:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid token"
            )

jwt_service = JWTService()

async def get_current_user(credentials: HTTPAuthorizationCredentials = Depends(security)):
    """Dependency: Extract user from JWT."""
    token = credentials.credentials
    payload = jwt_service.verify_access_token(token)
    return {"user_id": payload["sub"], "roles": payload["roles"]}

def require_role(required_role: str):
    """Decorator: Enforce role-based access control."""
    async def role_checker(user: dict = Depends(get_current_user)):
        if required_role not in user["roles"]:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=f"Role '{required_role}' required"
            )
        return user
    return role_checker

@app.get("/admin/dashboard")
async def admin_dashboard(user: dict = Depends(require_role("admin"))):
    """Admin-only endpoint with RBAC."""
    return {"message": f"Welcome, admin {user['user_id']}"}
```

---

## Cryptography & Data Protection

### Pattern 4: Secure Password Hashing

**Best Practice**: Use bcrypt/argon2 with high work factor, unique salt per password.

**Implementation**:
```python
import bcrypt
from typing import Optional

class PasswordService:
    """Secure password hashing with bcrypt."""

    def __init__(self, rounds: int = 12):
        """
        Initialize with work factor.

        Args:
            rounds: bcrypt work factor (10-14 recommended, 12 is standard)
                   Each +1 doubles computation time
        """
        self.rounds = rounds

    def hash_password(self, password: str) -> str:
        """Hash password with bcrypt."""
        # bcrypt generates unique salt automatically
        salt = bcrypt.gensalt(rounds=self.rounds)
        hashed = bcrypt.hashpw(password.encode('utf-8'), salt)
        return hashed.decode('utf-8')

    def verify_password(self, password: str, hashed: str) -> bool:
        """Verify password against hash."""
        try:
            return bcrypt.checkpw(password.encode('utf-8'), hashed.encode('utf-8'))
        except Exception:
            return False

    def needs_rehash(self, hashed: str) -> bool:
        """Check if password needs rehashing (work factor changed)."""
        try:
            # Extract current rounds from hash
            current_rounds = bcrypt.gensalt(hashed.encode('utf-8'))
            return current_rounds != self.rounds
        except Exception:
            return True

password_service = PasswordService(rounds=12)

# Usage
hashed = password_service.hash_password("SecurePassword123!")
is_valid = password_service.verify_password("SecurePassword123!", hashed)
```

### Pattern 5: Encryption at Rest

**Best Practice**: AES-256-GCM for authenticated encryption, proper key management.

**Implementation**:
```python
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
import os
import base64

class EncryptionService:
    """AES-256-GCM encryption for sensitive data."""

    def __init__(self, key: bytes = None):
        """
        Initialize with encryption key.

        Args:
            key: 256-bit (32 bytes) encryption key
        """
        self.key = key or os.urandom(32)  # Use secure key from env/vault
        self.cipher = AESGCM(self.key)

    def encrypt(self, plaintext: str, associated_data: str = None) -> str:
        """
        Encrypt data with authenticated encryption.

        Args:
            plaintext: Data to encrypt
            associated_data: Additional authenticated data (not encrypted)

        Returns:
            Base64-encoded nonce + ciphertext
        """
        nonce = os.urandom(12)  # 96-bit nonce for GCM

        ciphertext = self.cipher.encrypt(
            nonce,
            plaintext.encode('utf-8'),
            associated_data.encode('utf-8') if associated_data else None
        )

        # Return: nonce + ciphertext (base64 encoded)
        encrypted = base64.b64encode(nonce + ciphertext).decode('utf-8')
        return encrypted

    def decrypt(self, encrypted: str, associated_data: str = None) -> str:
        """
        Decrypt authenticated encrypted data.

        Args:
            encrypted: Base64-encoded nonce + ciphertext
            associated_data: Additional authenticated data (must match encryption)

        Returns:
            Decrypted plaintext

        Raises:
            InvalidTag: If ciphertext tampered or associated_data mismatch
        """
        encrypted_bytes = base64.b64decode(encrypted)

        nonce = encrypted_bytes[:12]
        ciphertext = encrypted_bytes[12:]

        plaintext = self.cipher.decrypt(
            nonce,
            ciphertext,
            associated_data.encode('utf-8') if associated_data else None
        )

        return plaintext.decode('utf-8')

# Usage
encryption = EncryptionService()

# Encrypt sensitive data
encrypted_ssn = encryption.encrypt("123-45-6789", associated_data="user:12345")

# Decrypt with matching associated_data
decrypted_ssn = encryption.decrypt(encrypted_ssn, associated_data="user:12345")
```

---

## Error Handling Security

### Pattern 6: Safe Error Messages

**Principle**: Never expose internal details in error responses.

**Implementation**:
```python
import logging
from fastapi import Request, status
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError

logger = logging.getLogger(__name__)

class SecureErrorHandler:
    """Production-safe error handling."""

    @staticmethod
    async def generic_exception_handler(request: Request, exc: Exception):
        """Handle unexpected exceptions safely."""

        # Log full error details internally
        logger.error(
            "Unexpected error occurred",
            extra={
                "path": request.url.path,
                "method": request.method,
                "error": str(exc),
                "type": type(exc).__name__
            },
            exc_info=True  # Include stack trace in logs
        )

        # Return generic message to client
        return JSONResponse(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            content={"detail": "An internal error occurred"}
        )

    @staticmethod
    async def validation_exception_handler(request: Request, exc: RequestValidationError):
        """Handle validation errors with safe messages."""

        # Log validation failures
        logger.warning(
            "Validation error",
            extra={
                "path": request.url.path,
                "errors": exc.errors()
            }
        )

        # Return sanitized validation errors (no internal details)
        safe_errors = [
            {
                "field": ".".join(str(loc) for loc in err["loc"]),
                "message": err["msg"]
            }
            for err in exc.errors()
        ]

        return JSONResponse(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            content={"detail": "Validation error", "errors": safe_errors}
        )

# Register error handlers
app.add_exception_handler(Exception, SecureErrorHandler.generic_exception_handler)
app.add_exception_handler(RequestValidationError, SecureErrorHandler.validation_exception_handler)
```

---

**Version**: 1.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready
