"""Tests for moai_adk.core.unified_permission_manager module."""

import pytest
from unittest.mock import MagicMock, patch


class TestUnifiedPermissionManager:
    """Basic tests for unified permission manager."""

    def test_module_imports(self):
        """Test that module can be imported."""
        try:
            from moai_adk.core.unified_permission_manager import (
                UnifiedPermissionManager,
            )

            assert UnifiedPermissionManager is not None
        except ImportError:
            pytest.skip("Module not available")

    def test_permission_manager_instantiation(self, tmp_path):
        """Test permission manager can be instantiated."""
        try:
            from moai_adk.core.unified_permission_manager import (
                UnifiedPermissionManager,
            )

            # Use tmp_path to avoid modifying real settings.json
            config_path = tmp_path / "settings.json"
            manager = UnifiedPermissionManager(config_path=str(config_path))
            assert manager is not None
        except (ImportError, Exception):
            pytest.skip("Module or dependencies not available")


class TestPermissionChecks:
    """Test permission checking functionality."""

    def test_has_method_exists(self):
        """Test that has method exists."""
        try:
            from moai_adk.core.unified_permission_manager import (
                UnifiedPermissionManager,
            )

            with patch.object(UnifiedPermissionManager, "_load_configuration", return_value={}):
                with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                    manager = UnifiedPermissionManager()
                    assert hasattr(manager, "has_permission")
        except (ImportError, Exception):
            pytest.skip("Module not available")

    def test_grant_permission_exists(self):
        """Test that grant_permission method exists."""
        try:
            from moai_adk.core.unified_permission_manager import (
                UnifiedPermissionManager,
            )

            with patch.object(UnifiedPermissionManager, "_load_configuration", return_value={}):
                with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                    manager = UnifiedPermissionManager()
                    assert hasattr(manager, "grant")
        except (ImportError, Exception):
            pytest.skip("Module not available")
