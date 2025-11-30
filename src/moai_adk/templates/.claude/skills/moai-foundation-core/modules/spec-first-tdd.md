# SPEC-First TDD - Specification-Driven Development

Purpose: Specification-driven test-driven development workflow ensuring clear requirements before implementation through EARS format and RED-GREEN-REFACTOR cycles.

Version: 1.0.0
Last Updated: 2025-11-25

---

## Quick Reference (30 seconds)

SPEC-First TDD is MoAI-ADK's development methodology combining:

1. SPEC Generation - EARS format requirements (/moai:1-plan)
2. Test-Driven Development - RED-GREEN-REFACTOR (/moai:2-run)
3. Documentation Sync - Auto-generated docs (/moai:3-sync)

Three-Phase Workflow:
```
Phase 1: SPEC → spec-builder → .moai/specs/SPEC-XXX/spec.md
Phase 2: TDD → tdd-implementer → Code + Tests (≥85% coverage)
Phase 3: Docs → docs-manager → API docs + diagrams
```

Token Budget: SPEC 30K | TDD 180K | Docs 40K | Total 250K

Key Practice: Execute `/clear` after Phase 1 to save 45-50K tokens.

EARS Patterns:
- Ubiquitous: System SHALL always...
- Event-driven: WHEN <event>, system SHALL...
- State-driven: WHILE <state>, system SHALL...
- Unwanted: System SHALL NOT...
- Optional: WHERE possible, system SHOULD...

---

## Implementation Guide (5 minutes)

### Phase 1: SPEC Generation

Purpose: Define clear, testable requirements in EARS format before coding.

Workflow:
```bash
# 1. Generate SPEC
/moai:1-plan "Implement user authentication with JWT tokens"

# 2. spec-builder creates:
.moai/specs/SPEC-001/
 spec.md # EARS format requirements
 acceptance.md # Acceptance criteria
 complexity.yaml # Complexity analysis

# 3. Execute /clear (mandatory)
/clear # Saves 45-50K tokens, prepares clean context
```

EARS Format Structure:

```markdown
---
spec_id: SPEC-001
title: User Authentication System
version: 1.0.0
complexity: Medium
estimated_effort: 8-12 hours
---

## Requirements

### SPEC-001-REQ-01: User Registration (Ubiquitous)
Pattern: Ubiquitous
Statement: The system SHALL register users with email and password validation.

Acceptance Criteria:
- Email format validated (RFC 5322)
- Password strength: ≥8 chars, mixed case, numbers, symbols
- Duplicate email rejected with clear error
- Success returns user ID and confirmation email sent

Test Coverage Target: ≥90%

---

### SPEC-001-REQ-02: JWT Token Generation (Event-driven)
Pattern: Event-driven
Statement: WHEN a user successfully authenticates, the system SHALL generate a JWT token with 1-hour expiry.

Acceptance Criteria:
- Token includes user ID, email, role claims
- Token signed with RS256 algorithm
- Expiry set to 1 hour from generation
- Refresh token generated with 7-day expiry

Test Coverage Target: ≥95%

---

### SPEC-001-REQ-03: Token Validation (State-driven)
Pattern: State-driven
Statement: WHILE a request includes Authorization header, the system SHALL validate JWT token before processing.

Acceptance Criteria:
- Expired tokens rejected with 401 Unauthorized
- Invalid signature rejected with 401 Unauthorized
- Valid token extracts user claims successfully
- Token blacklist checked (revoked tokens)

Test Coverage Target: ≥95%

---

### SPEC-001-REQ-04: Weak Password Prevention (Unwanted)
Pattern: Unwanted
Statement: The system SHALL NOT allow passwords from common password lists (top 10K).

Acceptance Criteria:
- Common passwords rejected (e.g., "password123")
- Sequential patterns rejected (e.g., "abc123")
- User-specific patterns rejected (e.g., email prefix)
- Clear error message with improvement suggestions

Test Coverage Target: ≥85%

---

### SPEC-001-REQ-05: OAuth2 Integration (Optional)
Pattern: Optional
Statement: WHERE user chooses, the system SHOULD support OAuth2 authentication via Google and GitHub.

Acceptance Criteria:
- OAuth2 providers configurable
- User can link multiple providers to one account
- Provider-specific profile data merged
- Graceful fallback if provider unavailable

Test Coverage Target: ≥80%
```

Complexity Analysis:

```yaml
# .moai/specs/SPEC-001/complexity.yaml
complexity_metrics:
 total_requirements: 5
 critical_requirements: 3
 
 complexity_breakdown:
 SPEC-001-REQ-01: Medium # Standard CRUD + validation
 SPEC-001-REQ-02: Medium # JWT library integration
 SPEC-001-REQ-03: High # Security validation logic
 SPEC-001-REQ-04: Low # Lookup validation
 SPEC-001-REQ-05: High # External API integration
 
 estimated_effort:
 development: 8 hours
 testing: 4 hours
 total: 12 hours
 
 risk_factors:
 - Security-critical functionality
 - External OAuth2 provider dependencies
 - Token expiry edge cases
 
 dependencies:
 - PyJWT library
 - bcrypt library
 - OAuth2 client libraries
```

---

### Phase 2: Test-Driven Development

Purpose: Implement requirements through RED-GREEN-REFACTOR cycles with ≥85% coverage.

RED-GREEN-REFACTOR Cycle:

```python
# ====================================
# RED PHASE: Write failing test first
# ====================================

import pytest
from src.auth.registration import register_user
from src.models.user import User
from src.exceptions import ValidationError

def test_register_user_with_valid_data():
 """SPEC-001-REQ-01: User registration with valid data."""
 # Arrange
 email = "user@example.com"
 password = "SecureP@ssw0rd"
 
 # Act
 result = register_user(email=email, password=password)
 
 # Assert
 assert result.success is True
 assert result.user.email == email
 assert result.user.id is not None
 assert result.confirmation_sent is True

# Run test → FAILS (functions don't exist yet)

# ====================================
# GREEN PHASE: Minimal implementation
# ====================================

# src/auth/registration.py
from dataclasses import dataclass
from src.models.user import User
import bcrypt
import re

@dataclass
class RegistrationResult:
 success: bool
 user: User
 confirmation_sent: bool

def register_user(email: str, password: str) -> RegistrationResult:
 """Register new user with email and password."""
 # Minimal code to pass test
 user = User(id=1, email=email, password_hash=password)
 return RegistrationResult(
 success=True,
 user=user,
 confirmation_sent=True
 )

# Run test → PASSES (test now passes)

# ====================================
# REFACTOR PHASE: Improve quality
# ====================================

# src/auth/registration.py (refactored)
from dataclasses import dataclass
from typing import Optional
from src.models.user import User
from src.database import db_session
from src.email import send_confirmation_email
from src.exceptions import ValidationError
import bcrypt
import re

EMAIL_REGEX = re.compile(r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$')
PASSWORD_MIN_LENGTH = 8

@dataclass
class RegistrationResult:
 """Result of user registration attempt."""
 success: bool
 user: Optional[User]
 confirmation_sent: bool
 error: Optional[str] = None

def register_user(email: str, password: str) -> RegistrationResult:
 """Register new user with email and password.
 
 Implements SPEC-001-REQ-01: User Registration (Ubiquitous)
 
 Args:
 email: User email address (must be valid format)
 password: User password (≥8 chars, mixed case, numbers, symbols)
 
 Returns:
 RegistrationResult with user data or error
 
 Raises:
 ValidationError: If email or password invalid
 
 Example:
 >>> result = register_user("user@example.com", "SecureP@ssw0rd")
 >>> result.success
 True
 >>> result.user.email
 'user@example.com'
 """
 # Validate email format (RFC 5322 compliant)
 if not EMAIL_REGEX.match(email):
 raise ValidationError("Invalid email format")
 
 # Check for duplicate email
 existing_user = User.query.filter_by(email=email).first()
 if existing_user:
 raise ValidationError("Email already registered")
 
 # Validate password strength
 if not _is_password_strong(password):
 raise ValidationError(
 "Password must be ≥8 characters with mixed case, numbers, and symbols"
 )
 
 # Hash password securely
 password_hash = bcrypt.hashpw(
 password.encode('utf-8'),
 bcrypt.gensalt(rounds=12)
 ).decode('utf-8')
 
 # Create user
 user = User(email=email, password_hash=password_hash)
 db_session.add(user)
 db_session.commit()
 
 # Send confirmation email
 confirmation_sent = send_confirmation_email(user.email, user.id)
 
 return RegistrationResult(
 success=True,
 user=user,
 confirmation_sent=confirmation_sent
 )

def _is_password_strong(password: str) -> bool:
 """Validate password strength.
 
 Requirements:
 - At least 8 characters
 - Contains uppercase letter
 - Contains lowercase letter
 - Contains digit
 - Contains special symbol
 """
 if len(password) < PASSWORD_MIN_LENGTH:
 return False
 
 has_upper = any(c.isupper() for c in password)
 has_lower = any(c.islower() for c in password)
 has_digit = any(c.isdigit() for c in password)
 has_symbol = any(not c.isalnum() for c in password)
 
 return all([has_upper, has_lower, has_digit, has_symbol])

# Run test → PASSES (with improved implementation)
# Coverage: 95% (meets ≥90% target)
```

Comprehensive Test Suite:

```python
# tests/auth/test_registration.py
import pytest
from src.auth.registration import register_user
from src.exceptions import ValidationError

class TestUserRegistration:
 """Test suite for SPEC-001-REQ-01: User Registration."""
 
 def test_register_user_success(self):
 """Test successful registration with valid data."""
 result = register_user("user@example.com", "SecureP@ssw0rd")
 assert result.success is True
 assert result.user.email == "user@example.com"
 assert result.confirmation_sent is True
 
 def test_register_user_invalid_email(self):
 """Test registration fails with invalid email format."""
 with pytest.raises(ValidationError, match="Invalid email format"):
 register_user("invalid-email", "SecureP@ssw0rd")
 
 def test_register_user_duplicate_email(self):
 """Test registration fails with duplicate email."""
 register_user("user@example.com", "SecureP@ssw0rd")
 
 with pytest.raises(ValidationError, match="Email already registered"):
 register_user("user@example.com", "AnotherP@ssw0rd")
 
 def test_register_user_weak_password(self):
 """Test registration fails with weak password."""
 weak_passwords = [
 "short", # Too short
 "alllowercase1", # No uppercase
 "ALLUPPERCASE1", # No lowercase
 "NoNumbersHere!", # No digits
 "NoSymbols123" # No symbols
 ]
 
 for weak_pwd in weak_passwords:
 with pytest.raises(ValidationError, match="Password must be"):
 register_user("user@example.com", weak_pwd)
 
 def test_register_user_password_hashing(self):
 """Test password is properly hashed (not stored plain text)."""
 result = register_user("user@example.com", "SecureP@ssw0rd")
 
 # Password hash should not equal plain text
 assert result.user.password_hash != "SecureP@ssw0rd"
 
 # Hash should be bcrypt format (starts with $2b$)
 assert result.user.password_hash.startswith("$2b$")
```

Coverage Report:

```bash
# Run tests with coverage
pytest tests/auth/test_registration.py --cov=src/auth/registration --cov-report=html

# Output:
---------- coverage: platform darwin, python 3.13.0 -----------
Name Stmts Miss Cover Missing
-------------------------------------------------------
src/auth/registration.py 42 2 95% 87, 92
-------------------------------------------------------
TOTAL 42 2 95%

# Coverage meets ≥90% target 
```

---

### Phase 3: Documentation Synchronization

Purpose: Auto-generate comprehensive documentation from implementation.

Workflow:
```bash
# 1. Generate documentation
/moai:3-sync SPEC-001

# 2. docs-manager creates:
.moai/specs/SPEC-001/
 docs/
 api.md # API reference
 architecture.md # Architecture diagram
 testing.md # Test report
 report.md # Implementation summary
```

Auto-Generated API Documentation:

```markdown
# API Documentation - User Authentication

## Endpoints

### POST /api/auth/register
Register new user account.

Request:
```json
{
 "email": "user@example.com",
 "password": "SecureP@ssw0rd"
}
```

Response (201 Created):
```json
{
 "success": true,
 "user": {
 "id": 123,
 "email": "user@example.com",
 "created_at": "2025-11-25T10:30:00Z"
 },
 "message": "Confirmation email sent"
}
```

Errors:
- `400 Bad Request`: Invalid email or weak password
- `409 Conflict`: Email already registered

Coverage: 95% (meets ≥90% target)

Implementation: `SPEC-001-REQ-01`
```

---

## Advanced Implementation (10+ minutes)

### EARS Pattern Advanced Usage

Complex Event-Driven Requirements:

```markdown
### SPEC-002-REQ-03: Multi-Factor Authentication (Event-driven + State-driven)
Pattern: Event-driven + State-driven
Statement: 
- WHEN a user attempts login with MFA enabled (Event)
- WHILE the MFA verification is pending (State)
- The system SHALL send TOTP code and require verification within 5 minutes

Acceptance Criteria:
1. Event trigger: Login attempt detected
2. State check: User has MFA enabled
3. Action: Generate TOTP code (6 digits, 30s validity)
4. Notification: Send code via SMS or email
5. Verification: User submits code within 5 minutes
6. Expiry: Code expires after 5 minutes
7. Rate limiting: Max 3 failed attempts, then 15-minute lockout

Test Scenarios:
- Happy path: User submits valid code within time
- Expired code: User submits code after 5 minutes
- Invalid code: User submits incorrect code
- Rate limit: User exceeds 3 failed attempts
- Disabled MFA: User without MFA enabled
```

Implementation:

```python
# RED PHASE: Complex test scenario
def test_mfa_verification_with_valid_code():
 """SPEC-002-REQ-03: MFA verification happy path."""
 # Arrange
 user = create_user_with_mfa_enabled()
 login_attempt = initiate_login(user.email, user.password)
 
 # System generates TOTP code
 totp_code = get_pending_totp_code(user.id)
 
 # Act
 result = verify_mfa(
 user.id,
 code=totp_code,
 timestamp=datetime.now()
 )
 
 # Assert
 assert result.success is True
 assert result.token is not None # JWT issued
 assert result.mfa_verified is True

def test_mfa_verification_with_expired_code():
 """SPEC-002-REQ-03: MFA code expiry."""
 user = create_user_with_mfa_enabled()
 login_attempt = initiate_login(user.email, user.password)
 totp_code = get_pending_totp_code(user.id)
 
 # Act - submit code after 6 minutes (expired)
 result = verify_mfa(
 user.id,
 code=totp_code,
 timestamp=datetime.now() + timedelta(minutes=6)
 )
 
 # Assert
 assert result.success is False
 assert result.error == "Code expired"

def test_mfa_rate_limiting():
 """SPEC-002-REQ-03: Rate limiting after failed attempts."""
 user = create_user_with_mfa_enabled()
 login_attempt = initiate_login(user.email, user.password)
 
 # 3 failed attempts
 for _ in range(3):
 verify_mfa(user.id, code="000000") # Invalid code
 
 # Act - 4th attempt should be blocked
 result = verify_mfa(user.id, code="123456")
 
 # Assert
 assert result.success is False
 assert result.error == "Too many failed attempts. Try again in 15 minutes."

# GREEN + REFACTOR PHASE
from datetime import datetime, timedelta
from typing import Optional
import pyotp

@dataclass
class MFAVerificationResult:
 success: bool
 token: Optional[str]
 mfa_verified: bool
 error: Optional[str] = None

class MFAManager:
 """Manage multi-factor authentication."""
 
 TOTP_VALIDITY_SECONDS = 300 # 5 minutes
 MAX_FAILED_ATTEMPTS = 3
 LOCKOUT_MINUTES = 15
 
 def verify_mfa(
 self,
 user_id: int,
 code: str,
 timestamp: datetime = None
 ) -> MFAVerificationResult:
 """Verify MFA code.
 
 Implements SPEC-002-REQ-03: Multi-Factor Authentication
 
 Args:
 user_id: User ID attempting verification
 code: 6-digit TOTP code
 timestamp: Verification timestamp (default: now)
 
 Returns:
 MFAVerificationResult with verification status
 """
 timestamp = timestamp or datetime.now()
 
 # Get user and pending verification
 user = User.query.get(user_id)
 pending_mfa = PendingMFA.query.filter_by(
 user_id=user_id,
 verified=False
 ).first()
 
 if not pending_mfa:
 return MFAVerificationResult(
 success=False,
 token=None,
 mfa_verified=False,
 error="No pending MFA verification"
 )
 
 # Check rate limiting
 if self._is_rate_limited(user_id):
 return MFAVerificationResult(
 success=False,
 token=None,
 mfa_verified=False,
 error=f"Too many failed attempts. Try again in {self.LOCKOUT_MINUTES} minutes."
 )
 
 # Check expiry (5 minutes from generation)
 if timestamp - pending_mfa.created_at > timedelta(seconds=self.TOTP_VALIDITY_SECONDS):
 return MFAVerificationResult(
 success=False,
 token=None,
 mfa_verified=False,
 error="Code expired"
 )
 
 # Verify TOTP code
 totp = pyotp.TOTP(user.mfa_secret)
 if not totp.verify(code, valid_window=1):
 self._record_failed_attempt(user_id)
 return MFAVerificationResult(
 success=False,
 token=None,
 mfa_verified=False,
 error="Invalid code"
 )
 
 # Success - mark verified and generate JWT
 pending_mfa.verified = True
 pending_mfa.verified_at = timestamp
 db_session.commit()
 
 jwt_token = self._generate_jwt_token(user)
 
 return MFAVerificationResult(
 success=True,
 token=jwt_token,
 mfa_verified=True
 )
 
 def _is_rate_limited(self, user_id: int) -> bool:
 """Check if user is rate limited."""
 recent_failures = FailedMFAAttempt.query.filter(
 FailedMFAAttempt.user_id == user_id,
 FailedMFAAttempt.timestamp > datetime.now() - timedelta(minutes=self.LOCKOUT_MINUTES)
 ).count()
 
 return recent_failures >= self.MAX_FAILED_ATTEMPTS
 
 def _record_failed_attempt(self, user_id: int):
 """Record failed MFA attempt for rate limiting."""
 attempt = FailedMFAAttempt(
 user_id=user_id,
 timestamp=datetime.now()
 )
 db_session.add(attempt)
 db_session.commit()
```

### SPEC-TDD Integration Patterns

Pattern 1: Iterative SPEC Refinement:

```python
# Initial SPEC (v1.0.0)
SPEC-003-REQ-01: File upload with size limit (10MB)

# Implementation reveals edge case
# → User uploads 9.9MB file successfully
# → But total storage exceeds user quota

# Refined SPEC (v1.1.0)
SPEC-003-REQ-01: File upload with size and quota validation
- Single file limit: 10MB
- User quota limit: 100MB total
- Validation: Check both limits before accepting upload

# Updated tests
def test_file_upload_exceeds_quota():
 """SPEC-003-REQ-01 v1.1.0: Quota validation."""
 user = create_user(quota_limit_mb=100)
 upload_files(user, total_size_mb=95) # Existing files
 
 # Attempt upload within file limit but exceeds quota
 result = upload_file(user, file_size_mb=8)
 
 assert result.success is False
 assert result.error == "Upload would exceed storage quota"
```

Pattern 2: SPEC-Driven Test Generation:

```python
# Automated test generation from SPEC

class SPECTestGenerator:
 """Generate tests automatically from SPEC requirements."""
 
 def generate_tests_for_requirement(self, requirement: dict) -> str:
 """Generate test code from SPEC requirement."""
 
 spec_id = requirement['id']
 pattern = requirement['pattern']
 acceptance_criteria = requirement['acceptance_criteria']
 
 if pattern == 'Event-driven':
 return self._generate_event_driven_tests(spec_id, acceptance_criteria)
 elif pattern == 'State-driven':
 return self._generate_state_driven_tests(spec_id, acceptance_criteria)
 elif pattern == 'Unwanted':
 return self._generate_unwanted_tests(spec_id, acceptance_criteria)
 else:
 return self._generate_standard_tests(spec_id, acceptance_criteria)
 
 def _generate_event_driven_tests(self, spec_id: str, criteria: list) -> str:
 """Generate tests for event-driven requirements."""
 
 test_template = f'''
def test_{spec_id.lower().replace('-', '_')}_event_triggered():
 """{spec_id}: Event-driven requirement test."""
 # Arrange
 {{setup_code}}
 
 # Act - Trigger event
 {{trigger_event}}
 
 # Assert - Verify system response
 {{assertions}}

def test_{spec_id.lower().replace('-', '_')}_no_event():
 """{spec_id}: No action when event not triggered."""
 # Arrange
 {{setup_code}}
 
 # Act - No event triggered
 {{no_event_code}}
 
 # Assert - System remains unchanged
 {{no_change_assertions}}
'''
 return test_template

# Usage
generator = SPECTestGenerator()
requirement = {
 'id': 'SPEC-001-REQ-02',
 'pattern': 'Event-driven',
 'acceptance_criteria': [...]
}
test_code = generator.generate_tests_for_requirement(requirement)
```

### Continuous SPEC-TDD Workflow

Automated Pipeline:

```yaml
# .github/workflows/spec-tdd-pipeline.yml
name: SPEC-First TDD Pipeline

on:
 push:
 paths:
 - '.moai/specs/'
 - 'src/'
 - 'tests/'

jobs:
 spec-validation:
 name: "Phase 1: SPEC Validation"
 runs-on: ubuntu-latest
 steps:
 - uses: actions/checkout@v4
 
 - name: Validate SPEC format
 run: python .moai/scripts/validate_spec.py
 
 - name: Check requirement traceability
 run: python .moai/scripts/check_traceability.py
 
 - name: Generate test scaffolding
 run: python .moai/scripts/generate_test_scaffolding.py
 
 tdd-implementation:
 name: "Phase 2: TDD Implementation"
 needs: spec-validation
 runs-on: ubuntu-latest
 steps:
 - uses: actions/checkout@v4
 
 - name: Run RED phase tests
 run: |
 pytest tests/ -v --tb=short || true # Allow failures
 echo "RED phase: Expected failures "
 
 - name: Verify tests exist for all requirements
 run: python .moai/scripts/verify_test_coverage_mapping.py
 
 quality-gates:
 name: "Phase 3: Quality Gates"
 needs: tdd-implementation
 runs-on: ubuntu-latest
 steps:
 - uses: actions/checkout@v4
 
 - name: Run tests with coverage
 run: pytest --cov=src --cov-fail-under=85
 
 - name: Validate TRUST 5
 run: python .moai/scripts/validate_trust5.py
 
 - name: Generate documentation
 run: python .moai/scripts/generate_docs.py
```

---

## Works Well With

Agents:
- spec-builder - EARS format SPEC generation
- tdd-implementer - RED-GREEN-REFACTOR execution
- quality-gate - TRUST 5 validation
- docs-manager - Documentation generation

Skills:
- moai-foundation-ears - EARS format patterns
- moai-foundation-trust - Quality framework
- moai-essentials-testing-integration - Test frameworks

Commands:
- /moai:1-plan - SPEC generation (Phase 1)
- /moai:2-run - TDD implementation (Phase 2)
- /moai:3-sync - Documentation sync (Phase 3)
- /clear - Token optimization between phases

Memory:
- Skill("moai-foundation-core") modules/execution-rules.md - SPEC decision criteria
- @.moai/specs/ - SPEC storage location

---

Version: 1.0.0
Last Updated: 2025-11-25
Status: Production Ready
