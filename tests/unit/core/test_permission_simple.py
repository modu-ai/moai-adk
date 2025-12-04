"""
Simple, working tests for moai_adk.core.unified_permission_manager module.

Focus: PermissionManager class, permission validation, and grants.
Target: 60%+ code coverage with AAA pattern.
"""

import pytest
import json
import tempfile
from pathlib import Path
from unittest.mock import patch, MagicMock, mock_open
from moai_adk.core.unified_permission_manager import (
    UnifiedPermissionManager,
    PermissionMode,
    PermissionSeverity,
    ResourceType,
    PermissionRule,
    ValidationResult,
    PermissionAudit,
    validate_agent_permission,
    check_tool_permission,
    auto_fix_all_agent_permissions,
    get_permission_stats,
)


class TestPermissionModeEnum:
    """Test PermissionMode enum."""

    def test_permission_mode_values(self):
        """Test all permission mode values are defined."""
        # Assert
        assert PermissionMode.ACCEPT_EDITS.value == "acceptEdits"
        assert PermissionMode.BYPASS_PERMISSIONS.value == "bypassPermissions"
        assert PermissionMode.DEFAULT.value == "default"
        assert PermissionMode.DONT_ASK.value == "dontAsk"
        assert PermissionMode.PLAN.value == "plan"

    def test_permission_mode_enum_members(self):
        """Test permission mode enum has expected members."""
        # Assert
        assert hasattr(PermissionMode, "ACCEPT_EDITS")
        assert hasattr(PermissionMode, "BYPASS_PERMISSIONS")
        assert hasattr(PermissionMode, "DEFAULT")


class TestPermissionSeverityEnum:
    """Test PermissionSeverity enum."""

    def test_severity_values(self):
        """Test all severity values are defined."""
        # Assert
        assert PermissionSeverity.LOW.value == "low"
        assert PermissionSeverity.MEDIUM.value == "medium"
        assert PermissionSeverity.HIGH.value == "high"
        assert PermissionSeverity.CRITICAL.value == "critical"


class TestUnifiedPermissionManagerInit:
    """Test UnifiedPermissionManager initialization."""

    def test_init_default_config_path(self):
        """Test initialization with default config path."""
        # Arrange & Act
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Assert
        assert manager.config_path == ".claude/settings.json"
        assert manager.enable_logging is True

    def test_init_custom_config_path(self):
        """Test initialization with custom config path."""
        # Arrange & Act
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager(config_path="/custom/path.json")

        # Assert
        assert manager.config_path == "/custom/path.json"

    def test_init_initializes_caches_and_logs(self):
        """Test that initialization sets up caches."""
        # Arrange & Act
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Assert
        assert isinstance(manager.permission_cache, dict)
        assert isinstance(manager.audit_log, list)
        assert len(manager.audit_log) == 0

    def test_init_initializes_stats(self):
        """Test that stats are initialized."""
        # Arrange & Act
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Assert
        assert "validations_performed" in manager.stats
        assert manager.stats["validations_performed"] == 0
        assert manager.stats["auto_corrections"] == 0

    def test_init_sets_role_hierarchy(self):
        """Test role hierarchy is initialized."""
        # Arrange & Act
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Assert
        assert "admin" in manager.role_hierarchy
        assert "developer" in manager.role_hierarchy["admin"]


class TestLoadConfiguration:
    """Test configuration loading."""

    def test_load_configuration_file_exists(self):
        """Test loading existing configuration file."""
        # Arrange
        config_data = {"agents": {}}
        with tempfile.NamedTemporaryFile(mode="w", delete=False, suffix=".json") as f:
            json.dump(config_data, f)
            config_path = f.name

        try:
            # Act
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager(config_path=config_path)

            # Assert
            assert manager.config == config_data
        finally:
            Path(config_path).unlink()

    def test_load_configuration_file_not_found(self):
        """Test handling missing configuration file."""
        # Arrange & Act
        with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
            manager = UnifiedPermissionManager(config_path="/nonexistent/file.json")

        # Assert
        assert manager.config == {}

    def test_load_configuration_invalid_json(self):
        """Test handling invalid JSON in configuration."""
        # Arrange
        with tempfile.NamedTemporaryFile(mode="w", delete=False, suffix=".json") as f:
            f.write("invalid json {{{")
            config_path = f.name

        try:
            # Act
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager(config_path=config_path)

            # Assert
            assert manager.config == {}
        finally:
            Path(config_path).unlink()


class TestValidateAgentPermission:
    """Test agent permission validation."""

    def test_validate_valid_permission_mode(self):
        """Test validation of valid permission mode."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        agent_config = {
            "permissionMode": "acceptEdits",
            "description": "Test agent",
            "systemPrompt": "Test prompt",
        }

        # Act
        result = manager.validate_agent_permission("test-agent", agent_config)

        # Assert
        assert result.valid is True
        assert result.auto_corrected is False

    def test_validate_invalid_permission_mode_auto_correction(self):
        """Test auto-correction of invalid permission mode."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        agent_config = {
            "permissionMode": "invalid",
            "description": "Test agent",
            "systemPrompt": "Test prompt",
        }

        # Act
        result = manager.validate_agent_permission("backend-expert", agent_config)

        # Assert
        assert result.auto_corrected is True
        assert result.corrected_mode is not None
        assert result.corrected_mode in {"acceptEdits", "plan", "default"}

    def test_validate_missing_required_fields(self):
        """Test validation detects missing required fields."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        agent_config = {
            "permissionMode": "plan",
        }

        # Act
        result = manager.validate_agent_permission("test-agent", agent_config)

        # Assert
        assert len(result.warnings) > 0


class TestSuggestPermissionMode:
    """Test permission mode suggestion."""

    def test_suggest_security_agent_mode(self):
        """Test security agent gets restrictive mode."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Act
        mode = manager._suggest_permission_mode("security-expert")

        # Assert
        assert mode == PermissionMode.PLAN.value

    def test_suggest_expert_agent_mode(self):
        """Test expert agent gets ACCEPT_EDITS mode."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Act
        mode = manager._suggest_permission_mode("backend-expert")

        # Assert
        assert mode == PermissionMode.ACCEPT_EDITS.value

    def test_suggest_analyzer_mode(self):
        """Test analyzer agent gets PLAN mode."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Act
        mode = manager._suggest_permission_mode("code-analyzer")

        # Assert
        assert mode == PermissionMode.PLAN.value

    def test_suggest_unknown_agent_defaults(self):
        """Test unknown agent gets default mode."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Act
        mode = manager._suggest_permission_mode("unknown-agent-xyz")

        # Assert
        assert mode == PermissionMode.DEFAULT.value


class TestValidateToolPermissions:
    """Test tool permission validation."""

    def test_validate_tool_permissions_with_dangerous_tools(self):
        """Test detection of dangerous tools."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        dangerous_tools = ["Bash(rm -rf:*)", "Bash(sudo:*)"]

        # Act
        result = manager.validate_tool_permissions(dangerous_tools)

        # Assert
        assert len(result.warnings) > 0
        assert result.severity == PermissionSeverity.HIGH

    def test_validate_tool_permissions_safe_tools(self):
        """Test validation of safe tools."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        safe_tools = ["Read", "Task", "AskUserQuestion"]

        # Act
        result = manager.validate_tool_permissions(safe_tools)

        # Assert
        assert result.valid is True


class TestCheckToolPermission:
    """Test tool permission checking."""

    def test_check_admin_has_all_permissions(self):
        """Test that admin role has all tool permissions."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Act
        result = manager.check_tool_permission("admin", "Task", "execute")

        # Assert
        assert result is True

    def test_check_developer_has_limited_permissions(self):
        """Test that developer role has limited permissions."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Act
        result = manager.check_tool_permission("developer", "Task", "execute")

        # Assert
        assert result is True

    def test_check_user_restricted_permissions(self):
        """Test that user role has restricted permissions."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Act
        result = manager.check_tool_permission("user", "Bash", "execute")

        # Assert
        assert result is False

    def test_check_permission_caching(self):
        """Test that permission checks are cached."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Act
        result1 = manager.check_tool_permission("admin", "Task", "execute")
        result2 = manager.check_tool_permission("admin", "Task", "execute")

        # Assert
        assert result1 == result2
        assert len(manager.permission_cache) > 0


class TestValidateConfiguration:
    """Test configuration validation."""

    def test_validate_configuration_file_not_found(self):
        """Test validation of missing configuration file."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Act
        result = manager.validate_configuration("/nonexistent/file.json")

        # Assert
        assert result.valid is False
        assert result.severity == PermissionSeverity.CRITICAL

    def test_validate_configuration_with_valid_file(self):
        """Test validation of valid configuration file."""
        # Arrange
        config_data = {"agents": {}, "permissions": {}}
        with tempfile.NamedTemporaryFile(mode="w", delete=False, suffix=".json") as f:
            json.dump(config_data, f)
            config_path = f.name

        try:
            with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
                mock_load.return_value = {}
                with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                    manager = UnifiedPermissionManager()

            # Act
            result = manager.validate_configuration(config_path)

            # Assert
            assert result.valid is True
        finally:
            Path(config_path).unlink()


class TestAutoFixAgentPermissions:
    """Test automatic permission fixing."""

    def test_auto_fix_agent_permissions(self):
        """Test auto-fixing agent permissions."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {"agents": {}}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                with patch.object(UnifiedPermissionManager, "_save_configuration"):
                    manager = UnifiedPermissionManager()

        agent_config = {
            "permissionMode": "invalid",
            "description": "Test",
            "systemPrompt": "Test",
        }
        manager.config["agents"]["test-agent"] = agent_config

        # Act
        result = manager.auto_fix_agent_permissions("test-agent")

        # Assert
        assert result.auto_corrected is True

    def test_auto_fix_all_agents(self):
        """Test auto-fixing all agent permissions."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {"agents": {}}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                with patch.object(UnifiedPermissionManager, "_save_configuration"):
                    manager = UnifiedPermissionManager()

        # Act
        results = manager.auto_fix_all_agents()

        # Assert
        assert isinstance(results, dict)
        assert len(results) > 0


class TestPermissionAuditing:
    """Test permission auditing and statistics."""

    def test_get_permission_stats(self):
        """Test getting permission statistics."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Act
        stats = manager.get_permission_stats()

        # Assert
        assert "validations_performed" in stats
        assert "auto_corrections" in stats
        assert "cached_permissions" in stats
        assert "audit_log_entries" in stats

    def test_get_recent_audits(self):
        """Test retrieving recent audit entries."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Act
        audits = manager.get_recent_audits(limit=10)

        # Assert
        assert isinstance(audits, list)
        assert len(audits) <= 10

    def test_audit_log_max_size(self):
        """Test that audit log is limited to max size."""
        # Arrange
        with patch.object(UnifiedPermissionManager, "_load_configuration") as mock_load:
            mock_load.return_value = {}
            with patch.object(UnifiedPermissionManager, "_validate_all_permissions"):
                manager = UnifiedPermissionManager()

        # Add many audit entries
        for i in range(1500):
            manager._audit_permission_change(
                ResourceType.AGENT,
                f"agent-{i}",
                "test_action",
                None,
                None,
                f"Test audit {i}",
                False,
            )

        # Act & Assert
        assert len(manager.audit_log) <= 1000


class TestGlobalConvenienceFunctions:
    """Test global convenience functions."""

    def test_validate_agent_permission_function(self):
        """Test global validate_agent_permission function."""
        # Arrange
        agent_config = {
            "permissionMode": "acceptEdits",
            "description": "Test",
            "systemPrompt": "Test",
        }

        # Act
        result = validate_agent_permission("test-agent", agent_config)

        # Assert
        assert isinstance(result, ValidationResult)
        assert result.valid is True

    def test_check_tool_permission_function(self):
        """Test global check_tool_permission function."""
        # Act
        result = check_tool_permission("admin", "Task", "execute")

        # Assert
        assert isinstance(result, bool)
        assert result is True

    def test_auto_fix_all_agent_permissions_function(self):
        """Test global auto_fix_all_agent_permissions function."""
        # Act
        results = auto_fix_all_agent_permissions()

        # Assert
        assert isinstance(results, dict)

    def test_get_permission_stats_function(self):
        """Test global get_permission_stats function."""
        # Act
        stats = get_permission_stats()

        # Assert
        assert isinstance(stats, dict)
        assert "validations_performed" in stats


class TestPermissionRule:
    """Test PermissionRule dataclass."""

    def test_permission_rule_creation(self):
        """Test creating a permission rule."""
        # Arrange & Act
        rule = PermissionRule(
            resource_type=ResourceType.TOOL,
            resource_name="Task",
            action="execute",
            allowed=True,
        )

        # Assert
        assert rule.resource_type == ResourceType.TOOL
        assert rule.resource_name == "Task"
        assert rule.action == "execute"
        assert rule.allowed is True

    def test_permission_rule_with_conditions(self):
        """Test creating a permission rule with conditions."""
        # Arrange & Act
        rule = PermissionRule(
            resource_type=ResourceType.TOOL,
            resource_name="Bash",
            action="execute",
            allowed=True,
            conditions={"max_size": "1GB"},
        )

        # Assert
        assert rule.conditions["max_size"] == "1GB"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
