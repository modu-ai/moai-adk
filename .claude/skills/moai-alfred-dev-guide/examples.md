# Alfred Development Guide - Examples

## Example 1: Complete SPEC-First TDD Workflow

### Phase 1: Create SPEC with EARS Format

**Command**: `/alfred:1-plan "User authentication system"`

**Generated SPEC** (`.moai/specs/SPEC-AUTH-001/spec.md`):

```yaml
---
id: AUTH-001
title: "User Authentication System"
version: 0.1.0
status: active
created: 2025-11-13
updated: 2025-11-13
priority: high
author: alfred
---

# User Authentication System SPEC

## Requirements

### Ubiquitous Requirements
- The system shall provide user authentication via email and password.
- The system shall validate email format (RFC 5322) before storage.
- The system shall hash passwords using bcrypt (minimum 12 rounds).

### Event-driven Requirements
- WHEN a user submits signup form, the system shall create user account.
- WHEN email verification link is clicked, the system shall activate account.
- WHEN login fails 3 times, the system shall lock account for 1 hour.

### State-driven Requirements
- WHILE user is authenticated, the system shall allow access to protected resources.
- WHILE session is active, the system shall maintain JWT token validity.

### Optional Features
- WHERE 2FA is enabled, the system may require additional verification.
- WHERE social login is configured, the system may allow OAuth authentication.

### Constraints
- IF password is invalid, the system shall return 401 Unauthorized.
- IF email already exists, the system shall return 409 Conflict.
- The system shall NOT store plaintext passwords.
- The system shall enforce 8-character minimum password length.
```

### Phase 2: TDD Implementation

**Command**: `/alfred:2-run SPEC-AUTH-001`

**RED Phase - Write Failing Tests**:

```python
# tests/test_auth.py
import pytest
from datetime import datetime, timedelta
from src.auth import signup, login, verify_email
from src.models import User
from src.exceptions import AuthenticationError, ValidationError

class TestSignup:
    """@TEST:AUTH-001 - User signup functionality"""
    
    def test_signup_valid_user(self):
        """@SPEC:AUTH-001 - The system shall create user account."""
        response = signup(email="user@example.com", password="securePass123")
        
        assert response["status"] == "created"
        assert "user_id" in response
        
        user = User.find_by_email("user@example.com")
        assert user is not None
        assert user.verified is False
        assert user.password_hash is not None
        assert user.password_hash != "securePass123"  # Hashed
    
    def test_signup_invalid_email(self):
        """@SPEC:AUTH-001 - The system shall validate email format."""
        with pytest.raises(ValidationError, match="Invalid email"):
            signup(email="invalid-email", password="securePass123")
        
        with pytest.raises(ValidationError, match="Invalid email"):
            signup(email="user@", password="securePass123")
    
    def test_signup_weak_password(self):
        """@SPEC:AUTH-001 - The system shall enforce password requirements."""
        with pytest.raises(ValidationError, match="Password must be at least 8 characters"):
            signup(email="user@example.com", password="short")
        
        with pytest.raises(ValidationError, match="Password must contain at least one number"):
            signup(email="user@example.com", password="longbutnonumber")
    
    def test_signup_duplicate_email(self):
        """@SPEC:AUTH-001 - IF email already exists, return 409 Conflict."""
        # First signup
        signup(email="user@example.com", password="securePass123")
        
        # Duplicate signup
        with pytest.raises(ValidationError, match="Email already exists"):
            signup(email="user@example.com", password="differentPass123")

class TestLogin:
    """@TEST:AUTH-001 - User login functionality"""
    
    def test_login_valid_credentials(self):
        """@SPEC:AUTH-001 - WHILE user is authenticated, allow access."""
        # Create verified user
        user_data = signup(email="user@example.com", password="securePass123")
        verify_email(user_data["verification_token"])
        
        # Login
        token = login(email="user@example.com", password="securePass123")
        
        assert token is not None
        assert isinstance(token, str)
        assert len(token) > 50  # JWT tokens are long
    
    def test_login_invalid_credentials(self):
        """@SPEC:AUTH-001 - IF password is invalid, return 401 Unauthorized."""
        signup(email="user@example.com", password="securePass123")
        verify_email(get_verification_token("user@example.com"))
        
        with pytest.raises(AuthenticationError):
            login(email="user@example.com", password="wrongPassword")
    
    def test_login_account_lockout(self):
        """@SPEC:AUTH-001 - WHEN login fails 3 times, lock account."""
        user_data = signup(email="user@example.com", password="securePass123")
        verify_email(user_data["verification_token"])
        
        # 3 failed attempts
        for i in range(3):
            with pytest.raises(AuthenticationError):
                login(email="user@example.com", password=f"wrongPassword{i}")
        
        # 4th attempt should be locked
        with pytest.raises(AuthenticationError, match="Account locked"):
            login(email="user@example.com", password="securePass123")
```

**GREEN Phase - Minimal Implementation**:

```python
# src/auth.py
import bcrypt
import jwt
from datetime import datetime, timedelta
from typing import Dict, Optional
from src.models import User
from src.email import send_verification_email
from src.config import Settings

# @CODE:AUTH-001
def signup(email: str, password: str) -> Dict[str, str]:
    """Create new user account with email verification."""
    # Basic validation
    if "@" not in email or "." not in email.split("@")[1]:
        raise ValidationError("Invalid email")
    
    if len(password) < 8 or not any(c.isdigit() for c in password):
        raise ValidationError("Password must be at least 8 characters and contain at least one number")
    
    # Check for duplicate email
    if User.find_by_email(email):
        raise ValidationError("Email already exists")
    
    # Hash password
    hashed = bcrypt.hashpw(password.encode(), bcrypt.gensalt(12))
    
    # Create user
    user = User.create(email=email, password_hash=hashed, verified=False)
    
    # Generate verification token and send email
    token = generate_verification_token(user.id)
    send_verification_email(email, token)
    
    return {
        "status": "created", 
        "user_id": user.id,
        "verification_token": token
    }

# @CODE:AUTH-001
def login(email: str, password: str) -> str:
    """Authenticate user and return JWT token."""
    user = User.find_by_email(email)
    if not user:
        raise AuthenticationError("Invalid credentials")
    
    if user.locked and user.locked_until > datetime.now():
        raise AuthenticationError("Account locked")
    
    if not bcrypt.checkpw(password.encode(), user.password_hash):
        user.failed_attempts += 1
        if user.failed_attempts >= 3:
            user.locked = True
            user.locked_until = datetime.now() + timedelta(hours=1)
        user.save()
        raise AuthenticationError("Invalid credentials")
    
    # Reset failed attempts and create token
    user.failed_attempts = 0
    user.save()
    
    token = jwt.encode({
        "user_id": user.id,
        "exp": datetime.utcnow() + timedelta(hours=24)
    }, Settings.SECRET_KEY)
    
    return token
```

**REFACTOR Phase - Improved Implementation**:

```python
# src/auth.py (Refactored)
import bcrypt
import jwt
import re
from datetime import datetime, timedelta
from typing import Dict, Optional
from src.models import User
from src.email import send_verification_email
from src.config import Settings
from src.exceptions import AuthenticationError, ValidationError

# Constants
MIN_PASSWORD_LENGTH = 8
MAX_FAILED_ATTEMPTS = 3
LOCKOUT_DURATION_HOURS = 1

# @CODE:AUTH-001 (Refactored)
class AuthService:
    """Service class for user authentication operations."""
    
    @staticmethod
    def _validate_email(email: str) -> None:
        """Validate email format using RFC 5322 basic pattern."""
        pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
        if not re.match(pattern, email):
            raise ValidationError("Invalid email format")
    
    @staticmethod
    def _validate_password_strength(password: str) -> None:
        """Validate password meets security requirements."""
        if len(password) < MIN_PASSWORD_LENGTH:
            raise ValidationError(f"Password must be at least {MIN_PASSWORD_LENGTH} characters")
        
        if not re.search(r'\d', password):
            raise ValidationError("Password must contain at least one number")
        
        if not re.search(r'[A-Z]', password):
            raise ValidationError("Password must contain at least one uppercase letter")
    
    @staticmethod
    def _handle_failed_login(user: User) -> None:
        """Track failed login attempts and lock account if needed."""
        user.failed_attempts += 1
        if user.failed_attempts >= MAX_FAILED_ATTEMPTS:
            user.locked_until = datetime.now() + timedelta(hours=LOCKOUT_DURATION_HOURS)
        user.save()
    
    @staticmethod
    def _reset_login_attempts(user: User) -> None:
        """Reset failed attempt counter on successful login."""
        user.failed_attempts = 0
        user.locked_until = None
        user.save()
    
    @classmethod
    def signup(cls, email: str, password: str) -> Dict[str, str]:
        """Create new user account with validation and security."""
        cls._validate_email(email)
        cls._validate_password_strength(password)
        
        if User.find_by_email(email):
            raise ValidationError("Email already exists")
        
        # Hash password with bcrypt
        password_hash = bcrypt.hashpw(password.encode(), bcrypt.gensalt(12))
        
        # Create user
        user = User.create(
            email=email, 
            password_hash=password_hash,
            verified=False
        )
        
        # Send verification email
        token = cls._generate_verification_token(user.id)
        send_verification_email(email, token)
        
        return {
            "status": "created",
            "user_id": str(user.id),
            "message": "Verification email sent"
        }
    
    @classmethod
    def login(cls, email: str, password: str) -> str:
        """Authenticate user and return JWT token."""
        user = User.find_by_email(email)
        if not user:
            raise AuthenticationError("Invalid credentials")
        
        if not user.verified:
            raise AuthenticationError("Please verify your email first")
        
        if user.is_locked():
            raise AuthenticationError("Account temporarily locked")
        
        if not bcrypt.checkpw(password.encode(), user.password_hash):
            cls._handle_failed_login(user)
            raise AuthenticationError("Invalid credentials")
        
        cls._reset_login_attempts(user)
        return cls._generate_jwt_token(user.id)
    
    @staticmethod
    def _generate_verification_token(user_id: str) -> str:
        """Generate email verification token."""
        return jwt.encode({
            "user_id": str(user_id),
            "type": "email_verification",
            "exp": datetime.utcnow() + timedelta(hours=24)
        }, Settings.SECRET_KEY)
    
    @staticmethod
    def _generate_jwt_token(user_id: str) -> str:
        """Generate JWT authentication token."""
        return jwt.encode({
            "user_id": str(user_id),
            "type": "authentication",
            "exp": datetime.utcnow() + timedelta(hours=24)
        }, Settings.SECRET_KEY)

# Convenience functions for backward compatibility
def signup(email: str, password: str) -> Dict[str, str]:
    return AuthService.signup(email, password)

def login(email: str, password: str) -> str:
    return AuthService.login(email, password)
```

### Phase 3: Documentation Sync

**Command**: `/alfred:3-sync auto SPEC-AUTH-001`

**Updated SPEC Status**:
```yaml
status: completed
version: 1.0.0
updated: 2025-11-13
completed: 2025-11-13
```

**README.md Updates**:
```markdown
# @DOC:AUTH-001 - Authentication System

## Features
- Email/password authentication with bcrypt hashing
- Email verification system
- Account lockout after failed attempts
- JWT-based session management

## Usage

### Signup
```python
from src.auth import signup

response = signup("user@example.com", "SecurePass123")
# Check email for verification link
```

### Login
```python
from src.auth import login

token = login("user@example.com", "SecurePass123")
# Use token for API requests
```

## Security
- Passwords hashed with bcrypt (12 rounds)
- Email verification required
- Account lockout after 3 failed attempts
- JWT tokens with 24-hour expiration

*See SPEC-AUTH-001 for complete requirements*
```

## Example 2: Context Engineering in Practice

### Scenario: Large Codebase Development

**Context Budget Analysis**:
```bash
# Phase 1: Planning (/alfred:1-plan)
Context budget: 5MB maximum
Always load:
✅ .moai/project/product.md (150KB) - Project overview
✅ .moai/project/structure.md (50KB) - Codebase layout
Context used: 200KB / 5MB (4%)

# Phase 2: Implementation (/alfred:2-run SPEC-AUTH-001)
Context budget: 200MB maximum  
Always load:
✅ .moai/specs/SPEC-AUTH-001/spec.md (25KB) - Target requirements
✅ .claude/skills/moai-alfred-dev-guide/SKILL.md (45KB) - This guide
Optional load:
✅ tests/test_auth.py (15KB) - Existing tests (if any)
✅ src/auth.py (30KB) - Current implementation (if exists)
Context used: ~100KB / 200MB (0.05%)

# Phase 3: Sync (/alfred:3-sync)
Context budget: 50MB maximum
Always load:
✅ .moai/reports/last-sync.md (20KB) - Previous sync state
✅ Modified SPEC files only
Context used: ~40KB / 50MB (0.08%)
```

### JIT Loading Strategy

**Implementation with selective loading**:
```python
# Example: Agent loads only what's needed for each phase
class ContextManager:
    def load_context_for_phase(self, phase: str, spec_id: Optional[str] = None):
        if phase == "planning":
            return self._load_project_context()
        elif phase == "implementation" and spec_id:
            return self._load_spec_context(spec_id)
        elif phase == "sync":
            return self._load_sync_context()
    
    def _load_project_context(self):
        """Load minimal project overview for planning."""
        return {
            "product_spec": self._load_file(".moai/project/product.md"),
            "structure": self._load_file(".moai/project/structure.md"),
            "tech_stack": self._load_file(".moai/project/tech.md")
        }
    
    def _load_spec_context(self, spec_id: str):
        """Load specific SPEC and related implementation."""
        return {
            "spec": self._load_file(f".moai/specs/SPEC-{spec_id}/spec.md"),
            "guide": self._load_file(".claude/skills/moai-alfred-dev-guide/SKILL.md"),
            "current_code": self._load_existing_code(spec_id),
            "existing_tests": self._load_existing_tests(spec_id)
        }
```

## Example 3: TRUST 5 Validation in Action

### Quality Gate Implementation

```python
# src/quality/trust_validator.py
class TRUSTValidator:
    """Validates TRUST 5 principles compliance."""
    
    def validate_spec_auth_001(self) -> Dict[str, bool]:
        """@SPEC:AUTH-001 - Complete TRUST validation."""
        results = {
            "test_driven": self._validate_test_driven(),
            "readable": self._validate_readable(),
            "unified": self._validate_unified(),
            "secured": self._validate_secured(),
            "evaluated": self._validate_evaluated()
        }
        
        self._generate_trust_report("AUTH-001", results)
        return results
    
    def _validate_test_driven(self) -> bool:
        """Check RED-GREEN-REFACTOR cycle completeness."""
        checks = [
            self._has_spec_file("AUTH-001"),
            self._has_test_file("AUTH-001"),
            self._has_implementation("AUTH-001"),
            self._tests_fail_without_implementation("AUTH-001"),
            self._implementation_passes_tests("AUTH-001")
        ]
        return all(checks)
    
    def _validate_readable(self) -> bool:
        """Check code readability standards."""
        metrics = self._analyze_code_metrics("src/auth.py")
        return (
            metrics["cyclomatic_complexity"] < 10 and
            metrics["function_length_avg"] < 20 and
            metrics["docstring_coverage"] > 80
        )
    
    def _validate_unified(self) -> bool:
        """Check consistency across codebase."""
        return (
            self._consistent_naming_conventions("src/") and
            self._consistent_error_handling("src/") and
            self._consistent_test_structure("tests/")
        )
    
    def _validate_secured(self) -> bool:
        """Check security compliance."""
        security_checks = [
            not self._has_hardcoded_secrets("src/"),
            self._input_validation_present("src/auth.py"),
            self._error_handling_safe("src/auth.py"),
            self._passwords_hashed("src/auth.py")
        ]
        return all(security_checks)
    
    def _validate_evaluated(self) -> bool:
        """Check evaluation metrics."""
        coverage = self._get_test_coverage("src/auth.py")
        return coverage >= 85

# Usage
validator = TRUSTValidator()
results = validator.validate_spec_auth_001()
print(f"TRUST 5 Compliance: {sum(results.values())}/5")

# Expected output:
# TRUST 5 Compliance: 5/5
# ✅ Test-Driven: Complete RED-GREEN-REFACTOR cycle
# ✅ Readable: High code quality metrics
# ✅ Unified: Consistent patterns and conventions
# ✅ Secured: No security vulnerabilities found
# ✅ Evaluated: 92% test coverage achieved
```

## Example 4: Multi-Agent Orchestration

### Alfred Delegation Patterns

```python
# Example of Alfred coordinating specialist agents
def alfred_orchestrate_feature_development(feature_description: str):
    """Complete feature development orchestration."""
    
    # Phase 1: Planning
    plan_result = Task(
        subagent_type="spec-builder",
        prompt=f"Create detailed SPEC for: {feature_description}",
        context=_load_planning_context()
    )
    
    spec_id = plan_result["spec_id"]
    
    # Phase 2: Implementation
    implementation_result = Task(
        subagent_type="tdd-implementer",
        prompt=f"Implement SPEC-{spec_id} following TDD principles",
        context=_load_implementation_context(spec_id)
    )
    
    # Phase 3: Quality Validation
    quality_result = Task(
        subagent_type="quality-gate",
        prompt=f"Validate SPEC-{spec_id} meets TRUST 5 principles",
        context={"spec_id": spec_id, "implementation": implementation_result}
    )
    
    # Phase 4: Documentation Sync
    sync_result = Task(
        subagent_type="doc-syncer",
        prompt=f"Synchronize documentation for SPEC-{spec_id}",
        context={
            "spec_id": spec_id,
            "quality_report": quality_result,
            "implementation": implementation_result
        }
    )
    
    # Phase 5: Git Operations
    git_result = Task(
        subagent_type="git-manager",
        prompt=f"Commit and create PR for SPEC-{spec_id}",
        context={
            "spec_id": spec_id,
            "files_changed": sync_result["modified_files"],
            "commit_type": "feature"
        }
    )
    
    return {
        "spec_id": spec_id,
        "status": "completed",
        "pr_url": git_result["pr_url"],
        "quality_score": quality_result["trust_score"]
    }

# Alfred maintains traceability throughout the process
def generate_traceability_report(spec_id: str):
    """@TAG:SYSTEM - Generate complete traceability report."""
    
    tags = {
        "spec": rg(f"@SPEC:{spec_id}", ".moai/specs/"),
        "test": rg(f"@TEST:{spec_id}", "tests/"),
        "code": rg(f"@CODE:{spec_id}", "src/"),
        "doc": rg(f"@DOC:{spec_id}", "docs/")
    }
    
    report = f"""
# Traceability Report: SPEC-{spec_id}

## TAG Chain Verification
- @SPEC:{spec_id}: {len(tags['spec'])} locations
- @TEST:{spec_id}: {len(tags['test'])} locations  
- @CODE:{spec_id}: {len(tags['code'])} locations
- @DOC:{spec_id}: {len(tags['doc'])} locations

## Chain Integrity
✅ Complete chain: SPEC → TEST → CODE → DOC
✅ All TAGs properly formatted
✅ No orphaned CODE or TEST tags
✅ Documentation synchronized with implementation

## Quality Metrics
- Test Coverage: 92%
- TRUST 5 Score: 5/5
- Security Review: Passed
- Performance: Meets requirements
"""
    
    return report
```

---

*Examples optimized for real-world development scenarios*  
*Complete end-to-end workflows with best practices*  
*Enterprise-ready patterns and implementations*
