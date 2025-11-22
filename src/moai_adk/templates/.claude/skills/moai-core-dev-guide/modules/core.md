    assert exc_info.value.code == "AUTH_FAILED"
```

**Run RED Phase**:
```bash
$ pytest tests/test_auth.py -v
============================= test session starts =============================
collected 2 items

tests/test_auth.py::test_login_with_valid_credentials FAILED          [ 50%]
tests/test_auth.py::test_login_with_invalid_password FAILED            [100%]

============================== 2 failed in 0.15s ==============================
```

### Pattern 3: GREEN Phase - Minimal Implementation

**Objective**: Write just enough code to make tests pass.

```python
# app/auth.py
import sqlite3
import hashlib
import jwt
import datetime
from typing import Optional

class AuthenticationError(Exception):
    def __init__(self, message: str, code: str = "AUTH_ERROR"):
        super().__init__(message)
        self.code = code

class AuthService:
    """JWT-based authentication service."""
    
    def __init__(self, db_path: str, secret_key: str = "test-secret"):
        self.db_path = db_path
        self.secret_key = secret_key
        self.conn = None
    
    def initialize(self):
        self.conn = sqlite3.connect(self.db_path)
        cursor = self.conn.cursor()
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                username TEXT UNIQUE NOT NULL,
                password_hash TEXT NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        """)
        self.conn.commit()
    
    def create_user(self, username: str, password: str):
        password_hash = hashlib.sha256(password.encode()).hexdigest()
        cursor = self.conn.cursor()
        cursor.execute(
            "INSERT INTO users (username, password_hash) VALUES (?, ?)",
            (username, password_hash)
        )
        self.conn.commit()
    
    def login(self, username: str, password: str) -> str:
        """Authenticate user and generate JWT token."""
        # Input validation
        if not username:
            raise ValueError("Username cannot be empty")
        if not password:
            raise ValueError("Password cannot be empty")
        
        # Hash password and check against database
        password_hash = hashlib.sha256(password.encode()).hexdigest()
        cursor = self.conn.cursor()
        cursor.execute(
            "SELECT id FROM users WHERE username = ? AND password_hash = ?",
            (username, password_hash)
        )
        user = cursor.fetchone()
        
        if user is None:
            raise AuthenticationError("Invalid credentials", code="AUTH_FAILED")
        
        # Generate JWT token
        payload = {
            "user_id": user[0],
            "username": username,
            "exp": datetime.datetime.utcnow() + datetime.timedelta(hours=1)
        }
        token = jwt.encode(payload, self.secret_key, algorithm="HS256")
        return token
    
    def is_token_valid(self, token: str) -> bool:
        try:
            jwt.decode(token, self.secret_key, algorithms=["HS256"])
            return True
        except jwt.ExpiredSignatureError:
            return False
        except jwt.InvalidTokenError:
            return False
    
    def close(self):
        if self.conn:
            self.conn.close()
```

### Pattern 4: REFACTOR Phase - Code Improvement

**Objective**: Improve code quality without changing behavior.

**Before (GREEN phase - minimal)**:
```python
def login(self, username: str, password: str) -> str:
    if not username:
        raise ValueError("Username cannot be empty")
    if not password:
        raise ValueError("Password cannot be empty")
    
    password_hash = hashlib.sha256(password.encode()).hexdigest()
    cursor = self.conn.cursor()
    cursor.execute(
        "SELECT id FROM users WHERE username = ? AND password_hash = ?",
        (username, password_hash)
    )
    user = cursor.fetchone()
    
    if user is None:
        raise AuthenticationError("Invalid credentials", code="AUTH_FAILED")
    
    payload = {
        "user_id": user[0],
        "username": username,
        "exp": datetime.datetime.utcnow() + datetime.timedelta(hours=1)
    }
    token = jwt.encode(payload, self.secret_key, algorithm="HS256")
    return token
```

**After (REFACTOR phase - improved)**:
```python
def login(self, username: str, password: str) -> str:
    """Authenticate user and generate JWT token.
    
    Args:
        username: User's login name (1-255 characters)
        password: Plain-text password (will be hashed)
    
    Returns:
        JWT token string valid for 1 hour
    
    Raises:
        ValidationError: If input validation fails
        AuthenticationError: If credentials are invalid
    """
    self._validate_login_inputs(username, password)
    user_id = self._authenticate_user(username, password)
    return self._generate_token(user_id, username)

def _validate_login_inputs(self, username: str, password: str) -> None:
    """Validate login input parameters."""
    if not username:
        raise ValidationError("Username cannot be empty")
    if not password:
        raise ValidationError("Password cannot be empty")
    if len(username) > 255:
        raise ValidationError("Username too long (max 255 characters)")

def _authenticate_user(self, username: str, password: str) -> int:
    """Authenticate user against database."""
    password_hash = hashlib.sha256(password.encode()).hexdigest()
    cursor = self.conn.cursor()
    cursor.execute(
        "SELECT id FROM users WHERE username = ? AND password_hash = ?",
        (username, password_hash)
    )
    user = cursor.fetchone()
    
    if user is None:
        raise AuthenticationError("Invalid credentials", code="AUTH_FAILED")
    
    return user[0]

def _generate_token(self, user_id: int, username: str) -> str:
    """Generate JWT token with 1-hour expiry."""
    payload = {
        "user_id": user_id,
        "username": username,
        "exp": datetime.datetime.utcnow() + datetime.timedelta(hours=1),
        "iat": datetime.datetime.utcnow()
    }
    return jwt.encode(payload, self.secret_key, algorithm="HS256")
```

### Pattern 5: TRUST 5 Principles Implementation

**TRUST 5 Framework**:

1. **T**est-driven: RED → GREEN → REFACTOR mandatory
2. **R**eadable: Clear naming, documentation, type hints
3. **U**nified: Consistent patterns, style guides
4. **S**ecured: OWASP compliance, security reviews
5. **E**valuated: Metrics, coverage ≥85%, performance benchmarks

**T: Test-Driven Example**:
```python
# ✅ CORRECT: Test-first approach
# Step 1: Write failing test (RED)
def test_delete_user_removes_from_database(auth_service, user_factory):
    user_id = user_factory("alice", "password")
    
    auth_service.delete_user(user_id)
