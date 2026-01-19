"""
Comprehensive TDD test suite for unified_permission_manager.py

This test suite covers:
- All enums (PermissionMode, PermissionSeverity, ResourceType)
- Data classes (PermissionRule, ValidationResult, PermissionAudit)
- UnifiedPermissionManager class with all methods
- Permission validation and auto-correction
- Role-based access control
- Permission caching
- Audit logging
- Configuration file handling
- Security validation
- Edge cases and error conditions

Coverage Target: 100%
"""

import json
import os
import tempfile

import pytest

from moai_adk.core.unified_permission_manager import (
    PermissionAudit,
    PermissionMode,
    PermissionRule,
    PermissionSeverity,
    ResourceType,
    UnifiedPermissionManager,
    ValidationResult,
)

# =============================================================================
# ENUM TESTS
# =============================================================================


class TestPermissionMode:
    """Test PermissionMode enum"""

    def test_all_permission_modes(self):
        """Test all PermissionMode values"""
        assert PermissionMode.ACCEPT_EDITS.value == "acceptEdits"
        assert PermissionMode.BYPASS_PERMISSIONS.value == "bypassPermissions"
        assert PermissionMode.DEFAULT.value == "default"
        assert PermissionMode.DONT_ASK.value == "dontAsk"
        assert PermissionMode.PLAN.value == "plan"


class TestPermissionSeverity:
    """Test PermissionSeverity enum"""

    def test_all_severity_levels(self):
        """Test all PermissionSeverity values"""
        assert PermissionSeverity.LOW.value == "low"
        assert PermissionSeverity.MEDIUM.value == "medium"
        assert PermissionSeverity.HIGH.value == "high"
        assert PermissionSeverity.CRITICAL.value == "critical"


class TestResourceType:
    """Test ResourceType enum"""

    def test_all_resource_types(self):
        """Test all ResourceType values"""
        assert ResourceType.AGENT.value == "agent"
        assert ResourceType.TOOL.value == "tool"
        assert ResourceType.FILE.value == "file"
        assert ResourceType.COMMAND.value == "command"
        assert ResourceType.SETTING.value == "setting"


# =============================================================================
# DATA CLASS TESTS
# =============================================================================


class TestPermissionRule:
    """Test PermissionRule dataclass"""

    def test_initialization(self):
        """Test PermissionRule initialization"""
        rule = PermissionRule(
            resource_type=ResourceType.AGENT,
            resource_name="backend-expert",
            action="execute",
            allowed=True,
        )
        assert rule.resource_type == ResourceType.AGENT
        assert rule.resource_name == "backend-expert"
        assert rule.action == "execute"
        assert rule.allowed is True
        assert rule.conditions is None
        assert rule.expires_at is None

    def test_with_conditions(self):
        """Test PermissionRule with conditions"""
        conditions = {"max_tokens": 50000, "phase": "RED"}
        rule = PermissionRule(
            resource_type=ResourceType.TOOL,
            resource_name="Bash",
            action="execute",
            allowed=False,
            conditions=conditions,
            expires_at=1234567890.0,
        )
        assert rule.conditions == conditions
        assert rule.expires_at == 1234567890.0


class TestValidationResult:
    """Test ValidationResult dataclass"""

    def test_initialization(self):
        """Test ValidationResult initialization"""
        result = ValidationResult(valid=True)
        assert result.valid is True
        assert result.corrected_mode is None
        assert result.warnings == []
        assert result.errors == []
        assert result.severity == PermissionSeverity.LOW
        assert result.auto_corrected is False

    def test_with_correction(self):
        """Test ValidationResult with correction"""
        result = ValidationResult(
            valid=False,
            corrected_mode="acceptEdits",
            warnings=["Low disk space"],
            errors=["Invalid configuration"],
            severity=PermissionSeverity.HIGH,
            auto_corrected=True,
        )
        assert result.valid is False
        assert result.corrected_mode == "acceptEdits"
        assert "Low disk space" in result.warnings
        assert "Invalid configuration" in result.errors
        assert result.severity == PermissionSeverity.HIGH
        assert result.auto_corrected is True


class TestPermissionAudit:
    """Test PermissionAudit dataclass"""

    def test_initialization(self):
        """Test PermissionAudit initialization"""
        audit = PermissionAudit(
            timestamp=1234567890.0,
            user_id="user123",
            resource_type=ResourceType.AGENT,
            resource_name="backend-expert",
            action="permission_change",
            old_permissions={"permissionMode": "ask"},
            new_permissions={"permissionMode": "acceptEdits"},
            reason="Invalid mode corrected",
            auto_corrected=True,
        )
        assert audit.timestamp == 1234567890.0
        assert audit.user_id == "user123"
        assert audit.resource_type == ResourceType.AGENT
        assert audit.resource_name == "backend-expert"
        assert audit.action == "permission_change"
        assert audit.old_permissions == {"permissionMode": "ask"}
        assert audit.new_permissions == {"permissionMode": "acceptEdits"}
        assert audit.reason == "Invalid mode corrected"
        assert audit.auto_corrected is True


# =============================================================================
# UNIFIED PERMISSION MANAGER TESTS
# =============================================================================


class TestUnifiedPermissionManager:
    """Test UnifiedPermissionManager class"""

    @pytest.fixture
    def temp_config(self):
        """Create temporary config file"""
        config_data = {
            "agents": {
                "backend-expert": {
                    "permissionMode": "acceptEdits",
                    "description": "Backend development expert",
                    "systemPrompt": "You are a backend expert",
                    "model": "claude-sonnet-4-20250514",
                },
                "invalid-agent": {
                    "permissionMode": "invalid_mode",  # Invalid
                    "description": "Invalid agent",
                    "systemPrompt": "Test",
                },
            },
            "projectSettings": {
                "allowedTools": ["Task", "Read", "Write"],
            },
        }

        fd, path = tempfile.mkstemp(suffix=".json")
        with os.fdopen(fd, "w") as f:
            json.dump(config_data, f)

        yield path

        # Cleanup
        if os.path.exists(path):
            os.remove(path)

    @pytest.fixture
    def manager(self, temp_config):
        """Create manager instance with temp config"""
        return UnifiedPermissionManager(
            config_path=temp_config,
            enable_logging=False,  # Disable for tests
        )

    def test_initialization(self, manager):
        """Test manager initialization"""
        assert manager.config_path is not None
        assert manager.enable_logging is False
        assert isinstance(manager.permission_cache, dict)
        assert isinstance(manager.audit_log, list)
        assert isinstance(manager.stats, dict)
        assert "validations_performed" in manager.stats
        assert "auto_corrections" in manager.stats
        assert "security_violations" in manager.stats
        assert "permission_denied" in manager.stats

    def test_role_hierarchy(self, manager):
        """Test role hierarchy setup"""
        assert "admin" in manager.role_hierarchy
        assert "developer" in manager.role_hierarchy
        assert "user" in manager.role_hierarchy

        # Check inheritance
        assert "developer" in manager.role_hierarchy["admin"]
        assert "user" in manager.role_hierarchy["developer"]

    def test_default_permissions(self, manager):
        """Test default permission mappings"""
        assert "backend-expert" in UnifiedPermissionManager.DEFAULT_PERMISSIONS
        assert "frontend-expert" in UnifiedPermissionManager.DEFAULT_PERMISSIONS
        assert UnifiedPermissionManager.DEFAULT_PERMISSIONS["backend-expert"] == PermissionMode.ACCEPT_EDITS

    def test_valid_permission_modes(self, manager):
        """Test valid permission modes constant"""
        assert "acceptEdits" in UnifiedPermissionManager.VALID_PERMISSION_MODES
        assert "bypassPermissions" in UnifiedPermissionManager.VALID_PERMISSION_MODES
        assert "default" in UnifiedPermissionManager.VALID_PERMISSION_MODES
        assert "dontAsk" in UnifiedPermissionManager.VALID_PERMISSION_MODES
        assert "plan" in UnifiedPermissionManager.VALID_PERMISSION_MODES


# =============================================================================
# CONFIGURATION LOADING TESTS
# =============================================================================


class TestConfigurationLoading:
    """Test configuration loading"""

    def test_load_configuration_success(self):
        """Test successful configuration loading"""
        config_data = {
            "agents": {"test-agent": {"permissionMode": "default"}},
            "projectSettings": {"allowedTools": ["Task"]},
        }

        fd, path = tempfile.mkstemp(suffix=".json")
        with os.fdopen(fd, "w") as f:
            json.dump(config_data, f)

        try:
            manager = UnifiedPermissionManager(config_path=path, enable_logging=False)
            assert manager.config == config_data
        finally:
            if os.path.exists(path):
                os.remove(path)

    def test_load_configuration_file_not_found(self):
        """Test loading when file doesn't exist"""
        manager = UnifiedPermissionManager(
            config_path="/nonexistent/path/settings.json",
            enable_logging=False,
        )
        assert manager.config == {}

    def test_load_configuration_invalid_json(self):
        """Test loading with invalid JSON"""
        fd, path = tempfile.mkstemp(suffix=".json")
        with os.fdopen(fd, "w") as f:
            f.write("{ invalid json }")

        try:
            manager = UnifiedPermissionManager(config_path=path, enable_logging=False)
            assert manager.config == {}
        finally:
            if os.path.exists(path):
                os.remove(path)

    def test_load_configuration_empty_file(self):
        """Test loading empty configuration file"""
        fd, path = tempfile.mkstemp(suffix=".json")
        with os.fdopen(fd, "w") as f:
            json.dump({}, f)

        try:
            manager = UnifiedPermissionManager(config_path=path, enable_logging=False)
            assert manager.config == {}
        finally:
            if os.path.exists(path):
                os.remove(path)


# =============================================================================
# PERMISSION VALIDATION TESTS
# =============================================================================


class TestPermissionValidation:
    """Test permission validation"""

    @pytest.fixture
    def manager(self):
        """Create manager instance"""
        return UnifiedPermissionManager(enable_logging=False)

    def test_validate_agent_permission_valid(self, manager):
        """Test validation of valid agent permission"""
        agent_config = {
            "permissionMode": "acceptEdits",
            "description": "Test agent",
            "systemPrompt": "You are a test agent",
            "model": "claude-sonnet-4-20250514",
        }

        result = manager.validate_agent_permission("test-agent", agent_config)
        assert result.valid is True
        assert result.auto_corrected is False

    def test_validate_agent_permission_invalid_mode(self, manager):
        """Test validation of invalid permission mode"""
        agent_config = {
            "permissionMode": "invalid_mode",
            "description": "Test agent",
            "systemPrompt": "Test",
        }

        result = manager.validate_agent_permission("test-agent", agent_config)
        assert result.valid is True  # Still valid after correction
        assert result.auto_corrected is True
        assert result.corrected_mode is not None
        assert "invalid_mode" in str(result.errors)
        assert result.severity == PermissionSeverity.HIGH

    def test_validate_agent_permission_missing_model(self, manager):
        """Test validation with missing model field"""
        agent_config = {
            "permissionMode": "default",
            "description": "Test agent",
            "model": "",  # Empty model
        }

        result = manager.validate_agent_permission("test-agent", agent_config)
        assert "model" in str(result.errors[0]).lower()

    def test_validate_agent_permission_missing_required_fields(self, manager):
        """Test validation with missing required fields"""
        agent_config = {
            "permissionMode": "default",
            # Missing description and systemPrompt
        }

        result = manager.validate_agent_permission("test-agent", agent_config)
        assert len(result.warnings) >= 2
        assert any("description" in w for w in result.warnings)
        assert any("systemPrompt" in w for w in result.warnings)

    def test_validate_all_permissions(self, manager):
        """Test validating all permissions in configuration"""
        config = {
            "agents": {
                "agent1": {"permissionMode": "acceptEdits", "description": "Test", "systemPrompt": "Test"},
                "agent2": {"permissionMode": "invalid", "description": "Test", "systemPrompt": "Test"},
            },
            "projectSettings": {
                "allowedTools": ["Task", "Read"],
            },
        }

        manager.config = config
        manager._validate_all_permissions()

        # agent2 should have been corrected
        assert manager.config["agents"]["agent2"]["permissionMode"] != "invalid"


# =============================================================================
# PERMISSION MODE SUGGESTION TESTS
# =============================================================================


class TestPermissionModeSuggestion:
    """Test permission mode suggestion logic"""

    @pytest.fixture
    def manager(self):
        """Create manager instance"""
        return UnifiedPermissionManager(enable_logging=False)

    def test_suggest_mode_for_security_agent(self, manager):
        """Test mode suggestion for security-focused agent"""
        mode = manager._suggest_permission_mode("security-expert")
        assert mode == PermissionMode.PLAN.value

    def test_suggest_mode_for_audit_agent(self, manager):
        """Test mode suggestion for audit agent"""
        mode = manager._suggest_permission_mode("audit-agent")
        assert mode == PermissionMode.PLAN.value

    def test_suggest_mode_for_expert_agent(self, manager):
        """Test mode suggestion for expert agent"""
        mode = manager._suggest_permission_mode("backend-expert")
        assert mode == PermissionMode.ACCEPT_EDITS.value

    def test_suggest_mode_for_implementer_agent(self, manager):
        """Test mode suggestion for implementer agent"""
        mode = manager._suggest_permission_mode("tdd-implementer")
        assert mode == PermissionMode.ACCEPT_EDITS.value

    def test_suggest_mode_for_planner_agent(self, manager):
        """Test mode suggestion for planner agent"""
        mode = manager._suggest_permission_mode("planner-agent")
        assert mode == PermissionMode.PLAN.value

    def test_suggest_mode_for_manager_agent(self, manager):
        """Test mode suggestion for manager agent"""
        mode = manager._suggest_permission_mode("workflow-manager")
        assert mode == PermissionMode.ACCEPT_EDITS.value

    def test_suggest_mode_from_defaults(self, manager):
        """Test mode suggestion from default mappings"""
        mode = manager._suggest_permission_mode("api-designer")
        assert mode == PermissionMode.PLAN.value

    def test_suggest_mode_default_fallback(self, manager):
        """Test mode suggestion fallback to default"""
        mode = manager._suggest_permission_mode("unknown-random-agent")
        assert mode == PermissionMode.DEFAULT.value


# =============================================================================
# TOOL PERMISSION TESTS
# =============================================================================


class TestToolPermissions:
    """Test tool permission validation"""

    @pytest.fixture
    def manager(self):
        """Create manager instance"""
        return UnifiedPermissionManager(enable_logging=False)

    def test_validate_tool_permissions_safe(self, manager):
        """Test validation of safe tool permissions"""
        tools = ["Task", "Read", "Write", "Edit"]
        result = manager.validate_tool_permissions(tools)
        assert result.valid is True
        assert result.auto_corrected is False

    def test_validate_tool_permissions_dangerous(self, manager):
        """Test validation of dangerous tools"""
        tools = ["Bash(rm -rf:*)", "Task", "Read"]
        result = manager.validate_tool_permissions(tools)
        assert "Dangerous tool" in result.warnings[0]
        assert result.severity == PermissionSeverity.HIGH

    def test_validate_tool_permissions_multiple_dangerous(self, manager):
        """Test validation with multiple dangerous tools"""
        tools = [
            "Bash(rm -rf:*)",
            "Bash(sudo:*)",
            "Bash(chmod -R 777:*)",
            "Bash(dd:*)",
        ]
        result = manager.validate_tool_permissions(tools)
        assert len(result.warnings) == 4
        assert manager.stats["security_violations"] == 4


# =============================================================================
# PERMISSION CHECKING TESTS
# =============================================================================


class TestPermissionChecking:
    """Test permission checking logic"""

    @pytest.fixture
    def manager(self):
        """Create manager instance"""
        return UnifiedPermissionManager(enable_logging=False)

    def test_check_tool_permission_admin(self, manager):
        """Test admin permission for all tools"""
        permitted = manager.check_tool_permission("admin", "Bash", "execute")
        assert permitted is True

    def test_check_tool_permission_developer(self, manager):
        """Test developer permission for allowed tools"""
        permitted = manager.check_tool_permission("developer", "Task", "execute")
        assert permitted is True

    def test_check_tool_permission_developer_restricted(self, manager):
        """Test developer permission for restricted tools"""
        permitted = manager.check_tool_permission("developer", "Bash", "execute")
        assert permitted is True  # Bash is in developer allowed list

    def test_check_tool_permission_user_limited(self, manager):
        """Test user limited permissions"""
        permitted = manager.check_tool_permission("user", "Task", "execute")
        assert permitted is True

        permitted = manager.check_tool_permission("user", "Write", "execute")
        assert permitted is False

    def test_check_tool_permission_cache(self, manager):
        """Test permission caching"""
        # First call
        manager.check_tool_permission("user", "Task", "execute")

        # Should be cached
        cache_key = "user:Task:execute"
        assert cache_key in manager.permission_cache

        # Second call should use cache
        manager.check_tool_permission("user", "Task", "execute")
        assert manager.stats["validations_performed"] >= 2

    def test_check_tool_permission_denied(self, manager):
        """Test permission denied increments stats"""
        permitted = manager.check_tool_permission("user", "Write", "execute")
        assert permitted is False
        assert manager.stats["permission_denied"] >= 1

    def test_check_direct_permission_wildcard(self, manager):
        """Test wildcard permission"""
        permitted = manager._check_direct_permission("admin", "AnyTool", "execute")
        assert permitted is True

    def test_check_direct_permission_exact_match(self, manager):
        """Test exact tool match"""
        permitted = manager._check_direct_permission("developer", "Task", "execute")
        assert permitted is True

    def test_check_direct_permission_bash_pattern(self, manager):
        """Test Bash pattern matching"""
        permitted = manager._check_direct_permission("developer", "Bash(anything)", "execute")
        assert permitted is True

    def test_role_hierarchy_inheritance(self, manager):
        """Test role hierarchy permission inheritance"""
        # User should not have Write permission
        user_permitted = manager._check_direct_permission("user", "Write", "execute")
        assert user_permitted is False

        # But admin should (via developer inheritance)
        # Admin has all permissions via wildcard
        admin_permitted = manager.check_tool_permission("admin", "Write", "execute")
        assert admin_permitted is True


# =============================================================================
# CONFIGURATION VALIDATION TESTS
# =============================================================================


class TestConfigurationValidation:
    """Test configuration validation"""

    @pytest.fixture
    def manager(self):
        """Create manager instance"""
        return UnifiedPermissionManager(enable_logging=False)

    def test_validate_configuration_valid_file(self):
        """Test validation of valid configuration file"""
        config_data = {
            "agents": {
                "test-agent": {
                    "permissionMode": "acceptEdits",
                    "description": "Test",
                    "systemPrompt": "Test",
                },
            },
        }

        fd, path = tempfile.mkstemp(suffix=".json")
        with os.fdopen(fd, "w") as f:
            json.dump(config_data, f)

        try:
            manager = UnifiedPermissionManager(config_path=path, enable_logging=False)
            result = manager.validate_configuration(path)
            assert result.valid is True
        finally:
            if os.path.exists(path):
                os.remove(path)

    def test_validate_configuration_file_not_found(self, manager):
        """Test validation when file not found"""
        result = manager.validate_configuration("/nonexistent/file.json")
        assert result.valid is False
        assert "not found" in result.errors[0]
        assert result.severity == PermissionSeverity.CRITICAL

    def test_validate_configuration_invalid_json(self, manager):
        """Test validation with invalid JSON"""
        fd, path = tempfile.mkstemp(suffix=".json")
        with os.fdopen(fd, "w") as f:
            f.write("{ invalid }")

        try:
            result = manager.validate_configuration(path)
            assert result.valid is False
            assert "Invalid JSON" in result.errors[0]
            assert result.severity == PermissionSeverity.CRITICAL
        finally:
            if os.path.exists(path):
                os.remove(path)


# =============================================================================
# AUDIT LOG TESTS
# =============================================================================


class TestAuditLogging:
    """Test audit logging functionality"""

    @pytest.fixture
    def manager(self):
        """Create manager instance"""
        return UnifiedPermissionManager(enable_logging=False)

    def test_audit_permission_change(self, manager):
        """Test audit log entry creation"""
        manager._audit_permission_change(
            resource_type=ResourceType.AGENT,
            resource_name="test-agent",
            action="permission_mode_correction",
            old_permissions={"permissionMode": "invalid"},
            new_permissions={"permissionMode": "acceptEdits"},
            reason="Auto-correction applied",
            auto_corrected=True,
        )

        assert len(manager.audit_log) == 1

        audit_entry = manager.audit_log[0]
        assert audit_entry.resource_name == "test-agent"
        assert audit_entry.action == "permission_mode_correction"
        assert audit_entry.auto_corrected is True

    def test_audit_log_with_user(self, manager):
        """Test audit log with user ID"""
        manager._audit_permission_change(
            resource_type=ResourceType.TOOL,
            resource_name="Bash",
            action="permission_grant",
            old_permissions={"allowed": False},
            new_permissions={"allowed": True},
            reason="Admin granted access",
            auto_corrected=False,
            user_id="admin123",
        )

        audit_entry = manager.audit_log[0]
        assert audit_entry.user_id == "admin123"

    def test_get_audit_log(self, manager):
        """Test retrieving audit log"""
        manager._audit_permission_change(
            resource_type=ResourceType.AGENT,
            resource_name="test-agent",
            action="test",
            old_permissions={},
            new_permissions={},
            reason="Test entry",
            auto_corrected=False,
        )

        log = manager.get_audit_log()
        assert len(log) == 1

    def test_filter_audit_log_by_resource(self, manager):
        """Test filtering audit log by resource"""
        # Add multiple entries
        for i in range(3):
            manager._audit_permission_change(
                resource_type=ResourceType.AGENT,
                resource_name=f"agent-{i}",
                action="test",
                old_permissions={},
                new_permissions={},
                reason="Test",
                auto_corrected=False,
            )

        # Filter for specific agent
        filtered = manager.get_audit_log(resource_name="agent-1")
        assert len(filtered) == 1
        assert filtered[0].resource_name == "agent-1"


# =============================================================================
# STATISTICS TESTS
# =============================================================================


class TestStatistics:
    """Test statistics tracking"""

    @pytest.fixture
    def manager(self):
        """Create manager instance"""
        return UnifiedPermissionManager(enable_logging=False)

    def test_get_statistics(self, manager):
        """Test getting statistics"""
        stats = manager.get_statistics()
        assert "validations_performed" in stats
        assert "auto_corrections" in stats
        assert "security_violations" in stats
        assert "permission_denied" in stats

    def test_statistics_increment_on_validation(self, manager):
        """Test statistics increment on validation"""
        initial_count = manager.stats["validations_performed"]

        agent_config = {
            "permissionMode": "acceptEdits",
            "description": "Test",
            "systemPrompt": "Test",
        }
        manager.validate_agent_permission("test", agent_config)

        assert manager.stats["validations_performed"] == initial_count + 1

    def test_statistics_increment_on_correction(self, manager):
        """Test statistics increment on auto-correction"""
        agent_config = {
            "permissionMode": "invalid",
            "description": "Test",
            "systemPrompt": "Test",
        }
        manager.validate_agent_permission("test", agent_config)

        assert manager.stats["auto_corrections"] >= 1


# =============================================================================
# CONFIGURATION SAVING TESTS
# =============================================================================


class TestConfigurationSaving:
    """Test configuration saving"""

    def test_save_configuration(self):
        """Test saving corrected configuration"""
        config_data = {
            "agents": {
                "test-agent": {
                    "permissionMode": "invalid",  # Will be corrected
                    "description": "Test",
                    "systemPrompt": "Test",
                },
            },
        }

        fd, path = tempfile.mkstemp(suffix=".json")
        with os.fdopen(fd, "w") as f:
            json.dump(config_data, f)

        try:
            manager = UnifiedPermissionManager(config_path=path, enable_logging=False)

            # Configuration should have been validated and corrected
            with open(path, "r") as f:
                saved_config = json.load(f)

            # Permission mode should have been corrected
            assert saved_config["agents"]["test-agent"]["permissionMode"] != "invalid"

        finally:
            if os.path.exists(path):
                os.remove(path)


# =============================================================================
# CACHE INVALIDATION TESTS
# =============================================================================


class TestCacheInvalidation:
    """Test permission cache invalidation"""

    @pytest.fixture
    def manager(self):
        """Create manager instance"""
        return UnifiedPermissionManager(enable_logging=False)

    def test_clear_permission_cache(self, manager):
        """Test clearing permission cache"""
        # Add some entries to cache
        manager.permission_cache["user1:Task:execute"] = True
        manager.permission_cache["user2:Read:execute"] = True

        assert len(manager.permission_cache) == 2

        manager.clear_permission_cache()

        assert len(manager.permission_cache) == 0

    def test_invalidate_resource_cache(self, manager):
        """Test invalidating specific resource from cache"""
        # Add cache entries
        manager.permission_cache["user1:Bash:execute"] = True
        manager.permission_cache["user1:Read:execute"] = True
        manager.permission_cache["user2:Bash:execute"] = True

        # Invalidate Bash permissions
        manager.invalidate_resource_cache("Bash")

        # Bash entries should be removed
        assert "user1:Bash:execute" not in manager.permission_cache
        assert "user2:Bash:execute" not in manager.permission_cache

        # Other entries should remain
        assert "user1:Read:execute" in manager.permission_cache


# =============================================================================
# INTEGRATION TESTS
# =============================================================================


class TestIntegration:
    """Integration tests combining multiple components"""

    def test_full_validation_workflow(self):
        """Test complete validation and correction workflow"""
        config_data = {
            "agents": {
                "security-agent": {
                    "permissionMode": "ask",  # Invalid
                    "description": "Security agent",
                    "systemPrompt": "Security expert",
                },
                "backend-expert": {
                    "permissionMode": "invalid",  # Invalid
                    "description": "Backend expert",
                    "systemPrompt": "Backend expert",
                },
            },
            "projectSettings": {
                "allowedTools": ["Task", "Bash(rm -rf:*)", "Read"],
            },
        }

        fd, path = tempfile.mkstemp(suffix=".json")
        with os.fdopen(fd, "w") as f:
            json.dump(config_data, f)

        try:
            manager = UnifiedPermissionManager(config_path=path, enable_logging=False)

            # Check that corrections were made
            assert manager.config["agents"]["security-agent"]["permissionMode"] == "plan"
            assert manager.config["agents"]["backend-expert"]["permissionMode"] == "acceptEdits"

            # Check stats
            assert manager.stats["auto_corrections"] == 2

            # Check security violations for dangerous tool
            assert manager.stats["security_violations"] >= 1

            # Check audit log
            assert len(manager.audit_log) >= 2

        finally:
            if os.path.exists(path):
                os.remove(path)

    def test_role_based_access_control_workflow(self):
        """Test complete RBAC workflow"""
        manager = UnifiedPermissionManager(enable_logging=False)

        # Admin can do anything
        assert manager.check_tool_permission("admin", "AnyTool", "execute") is True

        # Developer has broader access
        assert manager.check_tool_permission("developer", "Write", "execute") is True
        assert manager.check_tool_permission("developer", "Bash", "execute") is True

        # User has limited access
        assert manager.check_tool_permission("user", "Task", "execute") is True
        assert manager.check_tool_permission("user", "Write", "execute") is False


# =============================================================================
# EDGE CASES AND ERROR CONDITIONS
# =============================================================================


class TestEdgeCases:
    """Test edge cases and error handling"""

    @pytest.fixture
    def manager(self):
        """Create manager instance"""
        return UnifiedPermissionManager(enable_logging=False)

    def test_empty_configuration(self):
        """Test handling empty configuration"""
        fd, path = tempfile.mkstemp(suffix=".json")
        with os.fdopen(fd, "w") as f:
            json.dump({}, f)

        try:
            manager = UnifiedPermissionManager(config_path=path, enable_logging=False)
            assert manager.config == {}
        finally:
            if os.path.exists(path):
                os.remove(path)

    def test_malformed_configuration(self):
        """Test handling malformed configuration"""
        fd, path = tempfile.mkstemp(suffix=".json")
        with os.fdopen(fd, "w") as f:
            f.write("completely malformed {{{")

        try:
            manager = UnifiedPermissionManager(config_path=path, enable_logging=False)
            assert manager.config == {}
        finally:
            if os.path.exists(path):
                os.remove(path)

    def test_permission_mode_case_sensitivity(self, manager):
        """Test permission mode is case-sensitive"""
        agent_config = {
            "permissionMode": "acceptedits",  # Wrong case
            "description": "Test",
            "systemPrompt": "Test",
        }

        result = manager.validate_agent_permission("test", agent_config)
        assert result.auto_corrected is True

    def test_concurrent_validation_safety(self, manager):
        """Test that concurrent validations are safe"""
        import threading

        def validate_agent(agent_id):
            agent_config = {
                "permissionMode": f"invalid_{agent_id}",
                "description": "Test",
                "systemPrompt": "Test",
            }
            manager.validate_agent_permission(f"agent_{agent_id}", agent_config)

        threads = [threading.Thread(target=validate_agent, args=(i,)) for i in range(10)]

        for t in threads:
            t.start()

        for t in threads:
            t.join()

        # All validations should have completed
        assert manager.stats["validations_performed"] >= 10


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
