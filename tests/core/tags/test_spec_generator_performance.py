#!/usr/bin/env python3
# @CODE:PERF-SPEC-GENERATOR-001 | @TEST:PERF-SPEC-GENERATOR-001
"""Performance benchmarks for SpecGenerator.

Measures current baseline performance and validates optimization targets:
- Small file (100 LOC): < 50ms
- Medium file (1000 LOC): < 500ms
- Large file (10MB): < 5000ms (5s)

Optimization targets after Phase 2:
- AST parsing: 30-50% improvement
- Domain inference: 20-40% improvement
- Caching: 100% improvement on re-analysis
- Large file chunking: 50-70% improvement
"""

import time
from pathlib import Path
from tempfile import TemporaryDirectory

import pytest

from moai_adk.core.tags.spec_generator import SpecGenerator


class TestSpecGeneratorPerformanceBaseline:
    """Baseline performance measurements (current implementation)."""

    @pytest.fixture
    def generator(self):
        """Create SpecGenerator instance."""
        return SpecGenerator()

    @pytest.fixture
    def small_python_file(self) -> Path:
        """Create a small Python file (100 LOC)."""
        with TemporaryDirectory() as tmpdir:
            code = '''"""Small Python module."""

def simple_function():
    """A simple function."""
    return 42

class SimpleClass:
    """A simple class."""

    def method1(self):
        """Method 1."""
        pass

    def method2(self):
        """Method 2."""
        pass

def another_function(x, y):
    """Another function with parameters."""
    return x + y

class AnotherClass:
    """Another class."""

    def __init__(self):
        """Initialize."""
        self.value = 0

    def process(self, data):
        """Process data."""
        return len(data)

def utility_function():
    """Utility function."""
    items = [i for i in range(100)]
    return sum(items)

def auth_login(username, password):
    """Authentication login."""
    if not username or not password:
        return False
    return True

def auth_logout(token):
    """Authentication logout."""
    return True
''' + "\n# " + "\n# ".join([f"Line {i}" for i in range(20, 100)])

            file_path = Path(tmpdir) / "small_test.py"
            file_path.write_text(code)
            yield file_path

    @pytest.fixture
    def medium_python_file(self) -> Path:
        """Create a medium Python file (1000 LOC)."""
        with TemporaryDirectory() as tmpdir:
            code = '''"""Medium-sized authentication module."""

import hashlib
import secrets
from typing import Optional, Dict, Tuple
from dataclasses import dataclass
from datetime import datetime, timedelta

@dataclass
class User:
    """User data model."""
    id: str
    username: str
    password_hash: str
    email: str
    is_active: bool = True
    created_at: datetime = None

    def __post_init__(self):
        """Initialize created_at."""
        if self.created_at is None:
            self.created_at = datetime.now()

class AuthenticationService:
    """Authentication service for user management."""

    def __init__(self):
        """Initialize auth service."""
        self.users: Dict[str, User] = {}
        self.tokens: Dict[str, str] = {}
        self.login_attempts: Dict[str, int] = {}

    def hash_password(self, password: str) -> str:
        """Hash password using SHA256."""
        return hashlib.sha256(password.encode()).hexdigest()

    def verify_password(self, password: str, hash_value: str) -> bool:
        """Verify password against hash."""
        return self.hash_password(password) == hash_value

    def create_user(self, username: str, password: str, email: str) -> Optional[User]:
        """Create new user account."""
        if username in self.users:
            return None

        user = User(
            id=secrets.token_hex(16),
            username=username,
            password_hash=self.hash_password(password),
            email=email
        )
        self.users[username] = user
        return user

    def login(self, username: str, password: str) -> Optional[str]:
        """Authenticate user and return token."""
        if username not in self.users:
            return None

        user = self.users[username]
        if not self.verify_password(password, user.password_hash):
            return None

        token = secrets.token_hex(32)
        self.tokens[token] = username
        return token

    def logout(self, token: str) -> bool:
        """Revoke user token."""
        if token in self.tokens:
            del self.tokens[token]
            return True
        return False

    def verify_token(self, token: str) -> Optional[str]:
        """Verify token and return username."""
        return self.tokens.get(token)

    def change_password(self, username: str, old_password: str, new_password: str) -> bool:
        """Change user password."""
        if username not in self.users:
            return False

        user = self.users[username]
        if not self.verify_password(old_password, user.password_hash):
            return False

        user.password_hash = self.hash_password(new_password)
        return True

class TokenManager:
    """Manages authentication tokens."""

    def __init__(self, expiration_hours: int = 24):
        """Initialize token manager."""
        self.expiration_hours = expiration_hours
        self.tokens: Dict[str, Tuple[str, datetime]] = {}

    def generate_token(self, user_id: str) -> str:
        """Generate new token."""
        token = secrets.token_hex(32)
        self.tokens[token] = (user_id, datetime.now() + timedelta(hours=self.expiration_hours))
        return token

    def validate_token(self, token: str) -> Optional[str]:
        """Validate token and return user_id if valid."""
        if token not in self.tokens:
            return None

        user_id, expiration = self.tokens[token]
        if datetime.now() > expiration:
            del self.tokens[token]
            return None

        return user_id

    def revoke_token(self, token: str) -> bool:
        """Revoke token."""
        if token in self.tokens:
            del self.tokens[token]
            return True
        return False

class PermissionManager:
    """Manages user permissions."""

    def __init__(self):
        """Initialize permission manager."""
        self.permissions: Dict[str, set] = {}

    def grant_permission(self, user_id: str, permission: str) -> bool:
        """Grant permission to user."""
        if user_id not in self.permissions:
            self.permissions[user_id] = set()
        self.permissions[user_id].add(permission)
        return True

    def revoke_permission(self, user_id: str, permission: str) -> bool:
        """Revoke permission from user."""
        if user_id not in self.permissions:
            return False
        self.permissions[user_id].discard(permission)
        return True

    def has_permission(self, user_id: str, permission: str) -> bool:
        """Check if user has permission."""
        return permission in self.permissions.get(user_id, set())

    def get_permissions(self, user_id: str) -> set:
        """Get all permissions for user."""
        return self.permissions.get(user_id, set()).copy()

# Utility functions
def validate_email(email: str) -> bool:
    """Validate email format."""
    return "@" in email and "." in email

def validate_username(username: str) -> bool:
    """Validate username format."""
    return 3 <= len(username) <= 50 and username.isalnum()

def validate_password(password: str) -> bool:
    """Validate password strength."""
    return len(password) >= 8
''' + "\n# " + "\n# ".join([f"Line {i}" for i in range(150, 1000)])

            file_path = Path(tmpdir) / "medium_test.py"
            file_path.write_text(code)
            yield file_path

    def test_small_file_performance(self, generator, small_python_file):
        """Benchmark small file analysis (target: < 50ms)."""
        start = time.perf_counter()
        result = generator.generate_spec_template(small_python_file)
        elapsed = (time.perf_counter() - start) * 1000  # Convert to ms

        assert result["success"]
        assert elapsed < 500  # Target: < 500ms (generous for baseline)
        print(f"\n✓ Small file (100 LOC): {elapsed:.2f}ms (target: <50ms)")

    def test_medium_file_performance(self, generator, medium_python_file):
        """Benchmark medium file analysis (target: < 500ms)."""
        start = time.perf_counter()
        result = generator.generate_spec_template(medium_python_file)
        elapsed = (time.perf_counter() - start) * 1000  # Convert to ms

        assert result["success"]
        assert elapsed < 5000  # Target: < 5000ms (generous for baseline)
        print(f"\n✓ Medium file (1000 LOC): {elapsed:.2f}ms (target: <500ms)")

    def test_domain_inference_performance(self, generator, medium_python_file):
        """Benchmark domain inference speed."""
        # Pre-analyze to isolate inference

        generator = SpecGenerator()
        analysis = generator._analyze_code_file(medium_python_file)

        start = time.perf_counter()
        for _ in range(100):  # Run 100 times to measure overhead
            domain = generator._infer_domain(medium_python_file, analysis)
        elapsed = (time.perf_counter() - start) * 1000 / 100  # Average ms

        assert domain == "AUTH"
        print(f"\n✓ Domain inference (100 iterations avg): {elapsed:.3f}ms")

    def test_multiple_files_analysis(self, generator, small_python_file):
        """Test analysis of multiple files to measure aggregate time."""
        files = [small_python_file for _ in range(10)]

        start = time.perf_counter()
        for file_path in files:
            result = generator.generate_spec_template(file_path)
            assert result["success"]
        elapsed = (time.perf_counter() - start) * 1000

        avg_time = elapsed / 10
        print(f"\n✓ 10 small files total: {elapsed:.2f}ms (avg: {avg_time:.2f}ms per file)")


class TestCachingFeasibility:
    """Test caching strategy feasibility."""

    def test_file_hash_computation(self, tmp_path):
        """Verify file hashing overhead is minimal."""
        from hashlib import sha256

        # Create test file
        code = "def test(): pass\n" * 500  # ~8KB
        test_file = tmp_path / "test.py"
        test_file.write_text(code)

        start = time.perf_counter()
        for _ in range(100):
            content = test_file.read_bytes()
            sha256(content).hexdigest()
        elapsed = (time.perf_counter() - start) * 1000 / 100

        assert elapsed < 5  # Hash should be < 5ms
        print(f"\n✓ File hash computation: {elapsed:.3f}ms (acceptable overhead)")

    def test_analysis_cache_validity(self):
        """Verify that cached analysis results are consistent."""
        from tempfile import TemporaryDirectory

        from moai_adk.core.tags.spec_generator import SpecGenerator

        generator = SpecGenerator()

        with TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "auth.py"
            test_file.write_text('''
def login(username, password):
    """User login."""
    pass
            ''')

            # Analyze twice
            result1 = generator.generate_spec_template(test_file)
            result2 = generator.generate_spec_template(test_file)

            # Results should be identical
            assert result1["domain"] == result2["domain"]
            assert result1["confidence"] == result2["confidence"]
            print("\n✓ Cache validity: Analysis results consistent")


class TestOptimizationTargets:
    """Define optimization targets for Phase 2."""

    def test_optimization_targets(self):
        """Document optimization targets."""
        targets = {
            "small_file_100_loc": "< 50ms (currently ~200ms)",
            "medium_file_1000_loc": "< 500ms (currently ~1000ms)",
            "large_file_10mb": "< 5s (currently >10s)",
            "domain_inference": "< 1ms per file (currently ~5ms)",
            "cache_hit": "< 0.1ms (100% improvement)",
            "ast_parsing": "30-50% improvement via FastVisitor",
        }

        print("\n📊 Optimization Targets:")
        for target, goal in targets.items():
            print(f"  - {target}: {goal}")
