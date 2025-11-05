# Phase 2: Run - TDD Implementation

The `/alfred:2-run` command executes the complete Test-Driven Development cycle, transforming your specifications into production-ready code through the proven RED→GREEN→REFACTOR methodology.

## Overview

**Purpose**: Implement specifications using rigorous TDD methodology with 85%+ test coverage guarantee.

**Command Format**:
```bash
/alfred:2-run SPEC-ID
```

**Typical Duration**: 5-15 minutes
**Output**: Complete implementation with tests, documentation, and quality validation

## Alfred's TDD Process

### Phase 1: Implementation Planning

Before writing any code, Alfred's **implementation-planner** analyzes the specification and creates a detailed implementation strategy.

#### Architecture Analysis

```bash
# Example: /alfred:2-run AUTH-001
```

Alfred analyzes the SPEC and determines:

1. **Technology Stack**
   - Primary framework (FastAPI, Express, etc.)
   - Database requirements (PostgreSQL, MongoDB, etc.)
   - Authentication libraries (JWT, OAuth providers)
   - Testing frameworks (pytest, Jest, etc.)

2. **Project Structure**
   ```
   src/
   ├── auth/
   │   ├── __init__.py
   │   ├── models.py      # Data models
   │   ├── service.py     # Business logic
   │   ├── api.py         # HTTP endpoints
   │   └── utils.py       # Helper functions
   └── tests/
       ├── test_models.py
       ├── test_service.py
       └── test_api.py
   ```

3. **TAG Assignment Strategy**
   - `@CODE:AUTH-001:MODEL` - Data models and schemas
   - `@CODE:AUTH-001:SERVICE` - Business logic layer
   - `@CODE:AUTH-001:API` - HTTP endpoint layer
   - `@CODE:AUTH-001:UTILS` - Utility functions

4. **Expert Activation**
   Based on SPEC keywords, Alfred activates relevant experts:
   - **backend-expert**: For API design and architecture
   - **security-expert**: For authentication and security
   - **database-expert**: For data persistence

#### Risk Assessment and Mitigation

Alfred identifies potential implementation challenges:

```markdown
## Implementation Risks

### High Risk
- Token security implementation complexity
- Password hashing performance under load
- Session management edge cases

### Medium Risk
- Email validation regex complexity
- Rate limiting implementation
- Error handling consistency

### Mitigation Strategies
- Use established libraries (PyJWT, bcrypt)
- Implement comprehensive logging
- Add performance monitoring
```

### Phase 2: TDD Execution

Alfred's **tdd-implementer** executes the complete TDD cycle with rigorous adherence to best practices.

#### 🔴 RED Phase - Write Failing Tests

**Objective**: Create comprehensive test coverage before any implementation.

**Test Categories**:

1. **Happy Path Tests**
   ```python
   def test_login_with_valid_credentials_should_return_tokens():
       """WHEN valid credentials provided, SHALL return access and refresh tokens"""
       response = client.post("/auth/login", json={
           "email": "user@example.com",
           "password": "SecurePass123!"
       })
       assert response.status_code == 200
       data = response.json()
       assert "access_token" in data
       assert "refresh_token" in data
       assert data["token_type"] == "bearer"
   ```

2. **Edge Case Tests**
   ```python
   def test_login_with_invalid_email_should_return_401():
       """WHEN invalid email format provided, SHALL return 401 error"""
       response = client.post("/auth/login", json={
           "email": "invalid-email",
           "password": "password123"
       })
       assert response.status_code == 401
       assert "error" in response.json()

   def test_login_with_wrong_password_should_return_401():
       """WHEN wrong password provided, SHALL return 401 error"""
       response = client.post("/auth/login", json={
           "email": "user@example.com",
           "password": "wrongpassword"
       })
       assert response.status_code == 401
   ```

3. **Security Tests**
   ```python
   def test_login_should_be_case_sensitive():
       """Email authentication SHALL be case sensitive"""
       response = client.post("/auth/login", json={
           "email": "User@Example.com",  # Different case
           "password": "SecurePass123!"
       })
       assert response.status_code == 401

   def test_login_should_reject_sql_injection_attempts():
       """Login SHALL prevent SQL injection attacks"""
       malicious_input = "'; DROP TABLE users; --"
       response = client.post("/auth/login", json={
           "email": malicious_input,
           "password": "password"
       })
       assert response.status_code == 400  # Bad request, not 500
   ```

4. **Performance Tests**
   ```python
   def test_login_response_time_should_be_under_500ms():
       """Login response time SHALL be under 500ms"""
       import time
       start_time = time.time()
       response = client.post("/auth/login", json={
           "email": "user@example.com",
           "password": "SecurePass123!"
       })
       end_time = time.time()
       assert response.status_code == 200
       assert (end_time - start_time) < 0.5
   ```

**Running RED Tests**:
```bash
pytest tests/test_auth.py -v
# Expected: All tests fail (No implementation yet)
```

**Commit RED Phase**:
```bash
git add tests/test_auth.py
git commit -m "🔴 test(AUTH-001): add failing authentication tests"
```

#### 🟢 GREEN Phase - Minimal Implementation

**Objective**: Write the simplest possible code to make all tests pass.

**Implementation Strategy**:

1. **Start with Data Models**
   ```python
   # src/auth/models.py
   # @CODE:AUTH-001:MODEL | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

   from pydantic import BaseModel, EmailStr, Field, validator
   from typing import Optional

   class LoginRequest(BaseModel):
       """@CODE:AUTH-001:MODEL - Login request validation"""
       email: EmailStr = Field(..., description="User email address")
       password: str = Field(..., min_length=8, max_length=128, description="User password")

       @validator('password')
       def validate_password_strength(cls, v):
           if not any(c.isupper() for c in v):
               raise ValueError('Password must contain at least one uppercase letter')
           if not any(c.islower() for c in v):
               raise ValueError('Password must contain at least one lowercase letter')
           if not any(c.isdigit() for c in v):
               raise ValueError('Password must contain at least one digit')
           return v

   class TokenResponse(BaseModel):
       """@CODE:AUTH-001:MODEL - Token response model"""
       access_token: str = Field(..., description="JWT access token")
       refresh_token: str = Field(..., description="JWT refresh token")
       token_type: str = Field(default="bearer", description="Token type")
       expires_in: int = Field(..., description="Token expiration time in seconds")
   ```

2. **Implement Business Logic**
   ```python
   # src/auth/service.py
   # @CODE:AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

   import bcrypt
   import jwt
   from datetime import datetime, timedelta
   from typing import Optional, Tuple
   from .models import LoginRequest, TokenResponse
   from ..database import get_user_by_email, verify_user_password

   class AuthService:
       """@CODE:AUTH-001:SERVICE - Authentication business logic"""

       def __init__(self, secret_key: str, algorithm: str = "HS256"):
           self.secret_key = secret_key
           self.algorithm = algorithm

       async def authenticate(self, login_data: LoginRequest) -> TokenResponse:
           """Authenticate user and return tokens"""
           # Find user by email
           user = await get_user_by_email(login_data.email)
           if not user:
               raise AuthenticationError("Invalid credentials")

           # Verify password
           if not await verify_user_password(user.id, login_data.password):
               raise AuthenticationError("Invalid credentials")

           # Generate tokens
           access_token = self._generate_access_token(user.id)
           refresh_token = self._generate_refresh_token(user.id)

           return TokenResponse(
               access_token=access_token,
               refresh_token=refresh_token,
               expires_in=900  # 15 minutes
           )

       def _generate_access_token(self, user_id: str) -> str:
           """Generate JWT access token"""
           expires = datetime.utcnow() + timedelta(minutes=15)
           payload = {
               "sub": user_id,
               "exp": expires,
               "type": "access"
           }
           return jwt.encode(payload, self.secret_key, algorithm=self.algorithm)

       def _generate_refresh_token(self, user_id: str) -> str:
           """Generate JWT refresh token"""
           expires = datetime.utcnow() + timedelta(days=7)
           payload = {
               "sub": user_id,
               "exp": expires,
               "type": "refresh"
           }
           return jwt.encode(payload, self.secret_key, algorithm=self.algorithm)

   class AuthenticationError(Exception):
       """Custom authentication error"""
       pass
   ```

3. **Create API Endpoints**
   ```python
   # src/auth/api.py
   # @CODE:AUTH-001:API | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py

   from fastapi import APIRouter, HTTPException, Depends, status
   from fastapi.security import HTTPBearer
   from .models import LoginRequest, TokenResponse
   from .service import AuthService, AuthenticationError

   router = APIRouter(prefix="/auth", tags=["authentication"])
   security = HTTPBearer()

   # Dependency injection
   async def get_auth_service() -> AuthService:
       return AuthService(secret_key="your-secret-key-here")

   @router.post("/login", response_model=TokenResponse, status_code=status.HTTP_200_OK)
   async def login(
       login_data: LoginRequest,
       auth_service: AuthService = Depends(get_auth_service)
   ):
       """Authenticate user and return JWT tokens"""
       try:
           tokens = await auth_service.authenticate(login_data)
           return tokens
       except AuthenticationError:
           raise HTTPException(
               status_code=status.HTTP_401_UNAUTHORIZED,
               detail="Invalid credentials",
               headers={"WWW-Authenticate": "Bearer"},
           )
       except Exception as e:
           # Log the error for debugging
           logger.error(f"Login error: {str(e)}")
           raise HTTPException(
               status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
               detail="Internal server error"
           )
   ```

**Running GREEN Tests**:
```bash
pytest tests/test_auth.py -v
# Expected: All tests pass
```

**Commit GREEN Phase**:
```bash
git add src/auth/
git commit -m "🟢 feat(AUTH-001): implement authentication service"
```

#### <span class="material-icons">recycling</span> REFACTOR Phase - Code Quality Improvement

**Objective**: Improve code quality while maintaining 100% test coverage.

**Refactoring Activities**:

1. **Apply SOLID Principles**
   ```python
   # Refactored service with dependency injection
   # src/auth/service.py (Improved)

   from abc import ABC, abstractmethod
   from typing import Protocol

   class UserRepository(Protocol):
       """Protocol for user repository"""
       async def get_by_email(self, email: str) -> Optional[User]:
           ...

       async def verify_password(self, user_id: str, password: str) -> bool:
           ...

   class TokenGenerator(ABC):
       """Abstract base for token generation"""
       @abstractmethod
       def generate_access_token(self, user_id: str) -> str:
           ...

       @abstractmethod
       def generate_refresh_token(self, user_id: str) -> str:
           ...

   class JWTTokenGenerator(TokenGenerator):
       """JWT token implementation"""
       def __init__(self, secret_key: str, algorithm: str = "HS256"):
           self.secret_key = secret_key
           self.algorithm = algorithm

       def generate_access_token(self, user_id: str) -> str:
           expires = datetime.utcnow() + timedelta(minutes=15)
           payload = {"sub": user_id, "exp": expires, "type": "access"}
           return jwt.encode(payload, self.secret_key, algorithm=self.algorithm)

       def generate_refresh_token(self, user_id: str) -> str:
           expires = datetime.utcnow() + timedelta(days=7)
           payload = {"sub": user_id, "exp": expires, "type": "refresh"}
           return jwt.encode(payload, self.secret_key, algorithm=self.algorithm)

   class AuthService:
       """Improved authentication service with dependency injection"""
       def __init__(
           self,
           user_repository: UserRepository,
           token_generator: TokenGenerator,
           rate_limiter: Optional[RateLimiter] = None
       ):
           self.user_repository = user_repository
           self.token_generator = token_generator
           self.rate_limiter = rate_limiter

       async def authenticate(self, login_data: LoginRequest) -> TokenResponse:
           """Authenticate user with rate limiting and security checks"""
           # Rate limiting check
           if self.rate_limiter:
               await self.rate_limiter.check_rate_limit(login_data.email)

           # Find and verify user
           user = await self.user_repository.get_by_email(login_data.email)
           if not user or not await self.user_repository.verify_password(
               user.id, login_data.password
           ):
               raise AuthenticationError("Invalid credentials")

           # Generate tokens
           access_token = self.token_generator.generate_access_token(user.id)
           refresh_token = self.token_generator.generate_refresh_token(user.id)

           return TokenResponse(
               access_token=access_token,
               refresh_token=refresh_token,
               expires_in=900
           )
   ```

2. **Add Comprehensive Error Handling**
   ```python
   # src/auth/exceptions.py
   # @CODE:AUTH-001:EXCEPTIONS | SPEC: SPEC-AUTH-001.md

   class AuthenticationError(Exception):
       """Base authentication error"""
       def __init__(self, message: str, error_code: str = None):
           self.message = message
           self.error_code = error_code
           super().__init__(message)

   class InvalidCredentialsError(AuthenticationError):
       """Invalid credentials provided"""
       def __init__(self):
           super().__init__("Invalid credentials", "AUTH_001")

   class AccountLockedException(AuthenticationError):
       """Account is locked"""
       def __init__(self, unlock_time: datetime):
           super().__init__(f"Account locked until {unlock_time}", "AUTH_002")
           self.unlock_time = unlock_time

   class RateLimitExceededError(AuthenticationError):
       """Rate limit exceeded"""
       def __init__(self, retry_after: int):
           super().__init__(f"Rate limit exceeded. Try again in {retry_after} seconds", "AUTH_003")
           self.retry_after = retry_after
   ```

3. **Implement Security Best Practices**
   ```python
   # src/auth/security.py
   # @CODE:AUTH-001:SECURITY | SPEC: SPEC-AUTH-001.md

   import secrets
   import hashlib
   from typing import Optional
   from datetime import datetime, timedelta

   class SecurityManager:
       """Security utilities for authentication"""

       @staticmethod
       def hash_password(password: str) -> str:
           """Hash password using bcrypt"""
           salt = bcrypt.gensalt(rounds=12)
           return bcrypt.hashpw(password.encode('utf-8'), salt).decode('utf-8')

       @staticmethod
       def verify_password(password: str, hashed: str) -> bool:
           """Verify password against hash"""
           return bcrypt.checkpw(password.encode('utf-8'), hashed.encode('utf-8'))

       @staticmethod
       def generate_secure_token(length: int = 32) -> str:
           """Generate cryptographically secure token"""
           return secrets.token_urlsafe(length)

       @staticmethod
       def is_strong_password(password: str) -> tuple[bool, list[str]]:
           """Check password strength"""
           errors = []
           if len(password) < 8:
               errors.append("Password must be at least 8 characters")
           if len(password) > 128:
               errors.append("Password must be less than 128 characters")
           if not any(c.isupper() for c in password):
               errors.append("Password must contain uppercase letter")
           if not any(c.islower() for c in password):
               errors.append("Password must contain lowercase letter")
           if not any(c.isdigit() for c in password):
               errors.append("Password must contain digit")
           if not any(c in "!@#$%^&*()_+-=[]{}|;:,.<>?" for c in password):
               errors.append("Password must contain special character")

           return len(errors) == 0, errors
   ```

**Verify Refactored Code**:
```bash
# Run all tests
pytest tests/ -v --cov=src

# Expected: All tests pass, 85%+ coverage
```

**Commit REFACTOR Phase**:
```bash
git add src/auth/
git commit -m "<span class="material-icons">recycling</span> refactor(AUTH-001): improve code quality and security"
```

### Phase 3: Quality Validation

Alfred's **trust-checker** and **quality-gate** validate the implementation against production standards.

#### TRUST 5 Validation

1. **Test First <span class="material-icons">check_circle</span>**
   - Test coverage: 100% for new code
   - All tests passing
   - Edge cases covered

2. **Readable <span class="material-icons">check_circle</span>**
   - Function length < 50 lines
   - Clear variable names
   - Proper documentation

3. **Unified <span class="material-icons">check_circle</span>**
   - Consistent API patterns
   - Proper error handling
   - Type safety

4. **Secured <span class="material-icons">check_circle</span>**
   - Input validation
   - Secure password handling
   - Rate limiting

5. **Trackable <span class="material-icons">check_circle</span>**
   - All code tagged with @CODE:AUTH-001
   - Git history clean
   - Documentation linked

#### Security and Performance Checks

```python
# Security validation
<span class="material-icons">check_circle</span> Passwords hashed with bcrypt (12 rounds)
<span class="material-icons">check_circle</span> JWT tokens use proper signing
<span class="material-icons">check_circle</span> Rate limiting implemented
<span class="material-icons">check_circle</span> Input validation on all endpoints
<span class="material-icons">check_circle</span> SQL injection protection
<span class="material-icons">check_circle</span> XSS prevention in error messages

# Performance validation
<span class="material-icons">check_circle</span> Login response time < 500ms
<span class="material-icons">check_circle</span> Token validation < 100ms
<span class="material-icons">check_circle</span> Database queries optimized
<span class="material-icons">check_circle</span> Memory usage within limits
```

## Advanced TDD Patterns

### Test Doubles and Mocking

```python
# tests/conftest.py
import pytest
from unittest.mock import AsyncMock, Mock
from src.auth.service import AuthService, JWTTokenGenerator

@pytest.fixture
def mock_user_repository():
    repo = AsyncMock()
    return repo

@pytest.fixture
def token_generator():
    return JWTTokenGenerator(secret_key="test-secret")

@pytest.fixture
def auth_service(mock_user_repository, token_generator):
    return AuthService(
        user_repository=mock_user_repository,
        token_generator=token_generator
    )

@pytest.fixture
def sample_user():
    return User(
        id="user-123",
        email="test@example.com",
        hashed_password="$2b$12$..."
    )
```

### Integration Testing

```python
# tests/test_auth_integration.py
import pytest
from fastapi.testclient import TestClient
from src.main import app

client = TestClient(app)

class TestAuthenticationIntegration:
    """Integration tests for authentication flow"""

    async def test_complete_auth_flow(self):
        """Test complete authentication workflow"""
        # Register user
        register_response = client.post("/auth/register", json={
            "email": "integration@example.com",
            "password": "SecurePass123!"
        })
        assert register_response.status_code == 201

        # Login
        login_response = client.post("/auth/login", json={
            "email": "integration@example.com",
            "password": "SecurePass123!"
        })
        assert login_response.status_code == 200
        tokens = login_response.json()

        # Use access token
        protected_response = client.get(
            "/auth/me",
            headers={"Authorization": f"Bearer {tokens['access_token']}"}
        )
        assert protected_response.status_code == 200

        # Refresh token
        refresh_response = client.post("/auth/refresh", json={
            "refresh_token": tokens["refresh_token"]
        })
        assert refresh_response.status_code == 200
```

### Property-Based Testing

```python
# tests/test_auth_property.py
import pytest
from hypothesis import given, strategies as st
from src.auth.security import SecurityManager

class TestSecurityProperties:
    """Property-based tests for security functions"""

    @given(st.text(min_size=8, max_size=128))
    def test_password_hash_properties(self, password):
        """Test password hashing properties"""
        # Hash password
        hashed = SecurityManager.hash_password(password)

        # Different passwords should have different hashes
        hashed2 = SecurityManager.hash_password(password + "!")
        assert hashed != hashed2

        # Same password should have different hashes (due to salt)
        hashed3 = SecurityManager.hash_password(password)
        assert hashed != hashed3

        # Verification should work
        assert SecurityManager.verify_password(password, hashed)
        assert not SecurityManager.verify_password(password + "!", hashed)

    @given(st.text())
    def test_password_strength_validation(self, password):
        """Test password strength validation"""
        is_strong, errors = SecurityManager.is_strong_password(password)

        if is_strong:
            # Should have no errors
            assert len(errors) == 0
        else:
            # Should have specific errors
            assert len(errors) > 0
            for error in errors:
                assert isinstance(error, str)
```

## Common Patterns and Best Practices

### Error Handling Patterns

```python
# Custom error hierarchy
class AuthError(Exception):
    """Base authentication error"""
    pass

class ValidationError(AuthError):
    """Validation error"""
    def __init__(self, field: str, message: str):
        self.field = field
        self.message = message
        super().__init__(f"{field}: {message}")

# Consistent error responses
def handle_auth_error(func):
    """Decorator for consistent error handling"""
    def wrapper(*args, **kwargs):
        try:
            return func(*args, **kwargs)
        except ValidationError as e:
            raise HTTPException(
                status_code=400,
                detail={"error": "validation_error", "field": e.field, "message": e.message}
            )
        except AuthError as e:
            raise HTTPException(
                status_code=401,
                detail={"error": "authentication_error", "message": str(e)}
            )
    return wrapper
```

### Configuration Management

```python
# src/auth/config.py
# @CODE:AUTH-001:CONFIG | SPEC: SPEC-AUTH-001.md

from pydantic import BaseSettings, Field
from typing import Optional

class AuthConfig(BaseSettings):
    """Authentication configuration"""

    # JWT settings
    jwt_secret_key: str = Field(..., env="JWT_SECRET_KEY")
    jwt_algorithm: str = Field(default="HS256", env="JWT_ALGORITHM")
    jwt_access_token_expire_minutes: int = Field(default=15, env="JWT_ACCESS_EXPIRE_MINUTES")
    jwt_refresh_token_expire_days: int = Field(default=7, env="JWT_REFRESH_EXPIRE_DAYS")

    # Password settings
    password_min_length: int = Field(default=8, env="PASSWORD_MIN_LENGTH")
    password_max_length: int = Field(default=128, env="PASSWORD_MAX_LENGTH")
    bcrypt_rounds: int = Field(default=12, env="BCRYPT_ROUNDS")

    # Rate limiting
    rate_limit_requests_per_minute: int = Field(default=5, env="RATE_LIMIT_RPM")
    rate_limit_burst_size: int = Field(default=10, env="RATE_LIMIT_BURST")

    # Security
    require_email_verification: bool = Field(default=True, env="REQUIRE_EMAIL_VERIFICATION")
    session_timeout_minutes: int = Field(default=30, env="SESSION_TIMEOUT_MINUTES")

    class Config:
        env_file = ".env"
        case_sensitive = False

# Global configuration instance
auth_config = AuthConfig()
```

## Troubleshooting

### Common TDD Issues

**Tests pass but implementation is incomplete**:
- Add more comprehensive test cases
- Test edge cases and error conditions
- Include performance and security tests

**Implementation becomes complex**:
- Break down into smaller components
- Use dependency injection
- Apply SOLID principles

**Test suite runs slowly**:
- Use mocks for external dependencies
- Optimize database operations
- Run tests in parallel

### Getting Help

```bash
# Check test coverage
pytest --cov=src --cov-report=html

# Run specific tests
pytest tests/test_auth.py::test_login_with_valid_credentials -v

# Debug failing tests
pytest tests/test_auth.py -vv --pdb

# Get help with implementation
/alfred:9-feedback
```

## Next Steps

After completing `/alfred:2-run`:

1. **Review Implementation**: Ensure all requirements are met
2. **Manual Testing**: Test the implementation manually
3. **Documentation Sync**: Run `/alfred:3-sync` to update documentation
4. **Code Review**: Share with team for review (if applicable)

The TDD implementation phase ensures your code is robust, well-tested, and meets production standards. By following the RED→GREEN→REFACTOR cycle, you create software that is maintainable, secure, and reliable! <span class="material-icons">rocket_launch</span>