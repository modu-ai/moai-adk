"""
Pytest configuration and shared fixtures for foundation module tests.

This conftest.py provides:
- Shared fixtures for all foundation module tests
- Test configuration settings
- Common test data and utilities
"""

from typing import Any, Dict

import pytest

# ============================================================================
# Sample Data Fixtures
# ============================================================================


@pytest.fixture
def sample_endpoint() -> Dict[str, Any]:
    """Sample API endpoint for testing."""
    return {"method": "GET", "path": "/api/v1/users", "status_code": 200}


@pytest.fixture
def sample_service() -> Dict[str, Any]:
    """Sample microservice configuration for testing."""
    return {
        "name": "user-service",
        "domain": "auth",
        "endpoints": [{"path": "/users", "method": "GET"}, {"path": "/users", "method": "POST"}],
    }


@pytest.fixture
def sample_database_schema() -> Dict[str, Any]:
    """Sample database schema for testing."""
    return {
        "users": {"id": "INTEGER PRIMARY KEY", "name": "VARCHAR(100)", "email": "VARCHAR(255)"},
        "orders": {
            "PRIMARY KEY": "(order_id, product_id)",
            "order_id": "INTEGER",
            "product_id": "INTEGER",
            "quantity": "INTEGER",
        },
    }


@pytest.fixture
def sample_git_version() -> str:
    """Sample Git version string."""
    return "git version 2.45.0"


@pytest.fixture
def sample_commit_message() -> str:
    """Sample valid Conventional Commit message."""
    return "feat(auth): add user authentication"


@pytest.fixture
def sample_performance_metrics() -> Dict[str, Any]:
    """Sample frontend performance metrics."""
    return {"lcp_seconds": 1.8, "fid_milliseconds": 45, "cls_value": 0.08, "bundle_size_kb": 150}


@pytest.fixture
def sample_user_permissions() -> Dict[str, Any]:
    """Sample user with permissions for testing."""
    return {"id": "user-123", "name": "Test User", "permissions": ["read:users", "write:posts", "admin:system"]}


@pytest.fixture
def mlflow_config() -> Dict[str, Any]:
    """Sample MLflow configuration."""
    return {
        "experiment_name": "test_experiment",
        "run_name": "test_run_001",
        "tracking_uri": "http://localhost:5000",
        "tags": {"framework": "pytorch", "environment": "test"},
    }


# ============================================================================
# Async Testing Fixtures
# ============================================================================


@pytest.fixture
def event_loop_policy():
    """Event loop policy for async tests."""
    import asyncio

    return asyncio.DefaultEventLoopPolicy()


# ============================================================================
# Test Utilities
# ============================================================================


def assert_valid_response(response: Dict[str, Any], required_keys: list):
    """Assert response has all required keys."""
    for key in required_keys:
        assert key in response, f"Missing required key: {key}"


def assert_valid_validation(result: Dict[str, Any]):
    """Assert validation result structure."""
    assert "valid" in result
    assert isinstance(result["valid"], bool)
    assert "errors" in result or "violations" in result


# ============================================================================
# Pytest Configuration
# ============================================================================


def pytest_configure(config):
    """Configure pytest with custom markers."""
    config.addinivalue_line("markers", "asyncio: mark test as async test")
    config.addinivalue_line("markers", "unit: mark test as unit test")
    config.addinivalue_line("markers", "integration: mark test as integration test")
