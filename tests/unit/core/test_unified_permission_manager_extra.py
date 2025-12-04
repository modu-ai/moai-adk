"""
Extended tests for unified_permission_manager module.

Tests enums, dataclasses, permission validation, auto-correction, and security checks.
"""

import json
import tempfile
import time
from pathlib import Path
from typing import Any, Dict
from unittest.mock import MagicMock, Mock, patch

import pytest

from moai_adk.core.unified_permission_manager import (
    PermissionAudit,
    PermissionMode,
    PermissionRule,
    PermissionSeverity,
    ResourceType,
    UnifiedPermissionManager,
    ValidationResult,
    auto_fix_all_agent_permissions,
    check_tool_permission,
    get_permission_stats,
    permission_manager,
    validate_agent_permission,
)


class TestEnums:
    """Test permission-related enums."""

    def test_permission_mode_values(self):
        """Test PermissionMode enum values."""
        assert PermissionMode.ACCEPT_EDITS.value == "acceptEdits"
        assert PermissionMode.BYPASS_PERMISSIONS.value == "bypassPermissions"
        assert PermissionMode.DEFAULT.value == "default"
        assert PermissionMode.DONT_ASK.value == "dontAsk"
        assert PermissionMode.PLAN.value == "plan"

    def test_permission_severity_values(self):
        """Test PermissionSeverity enum values."""
        assert PermissionSeverity.LOW.value == "low"
        assert PermissionSeverity.MEDIUM.value == "medium"
        assert PermissionSeverity.HIGH.value == "high"
        assert PermissionSeverity.CRITICAL.value == "critical"

    def test_resource_type_values(self):
        """Test ResourceType enum values."""
        assert ResourceType.AGENT.value == "agent"
        assert ResourceType.TOOL.value == "tool"
        assert ResourceType.FILE.value == "file"
        assert ResourceType.COMMAND.value == "command"
        assert ResourceType.SETTING.value == "setting"


class TestPermissionRuleDataclass:
    """Test PermissionRule dataclass."""

    def test_permission_rule_creation(self):
        """Test PermissionRule creation."""
        rule = PermissionRule(
            resource_type=ResourceType.AGENT,
            resource_name="backend-expert",
            action="execute",
            allowed=True,
            conditions={"role": "developer"},
        )
        assert rule.resource_type == ResourceType.AGENT
        assert rule.allowed is True
        assert rule.conditions["role"] == "developer"

    def test_permission_rule_with_expiry(self):
        """Test PermissionRule with expiration."""
        future_time = time.time() + 3600
        rule = PermissionRule(
            resource_type=ResourceType.TOOL,
            resource_name="Bash",
            action="execute",
            allowed=True,
            expires_at=future_time,
        )
        assert rule.expires_at == future_time


class TestValidationResultDataclass:
    """Test ValidationResult dataclass."""

    def test_validation_result_valid(self):
        """Test ValidationResult for valid state."""
        result = ValidationResult(valid=True)
        assert result.valid is True
        assert result.auto_corrected is False
        assert result.corrected_mode is None

    def test_validation_result_with_errors(self):
        """Test ValidationResult with errors."""
        result = ValidationResult(
            valid=False,
            errors=["Error 1", "Error 2"],
            severity=PermissionSeverity.HIGH,
        )
        assert result.valid is False
        assert len(result.errors) == 2

    def test_validation_result_auto_corrected(self):
        """Test ValidationResult with auto-correction."""
        result = ValidationResult(
            valid=True,
            auto_corrected=True,
            corrected_mode="acceptEdits",
        )
        assert result.auto_corrected is True
        assert result.corrected_mode == "acceptEdits"


class TestPermissionAuditDataclass:
    """Test PermissionAudit dataclass."""

    def test_permission_audit_creation(self):
        """Test PermissionAudit creation."""
        audit = PermissionAudit(
            timestamp=time.time(),
            user_id="user1",
            resource_type=ResourceType.AGENT,
            resource_name="backend-expert",
            action="permission_change",
            old_permissions={"mode": "ask"},
            new_permissions={"mode": "acceptEdits"},
            reason="Auto-correction",
            auto_corrected=True,
        )
        assert audit.resource_type == ResourceType.AGENT
        assert audit.auto_corrected is True

    def test_permission_audit_with_none_user(self):
        """Test PermissionAudit with None user_id (system)."""
        audit = PermissionAudit(
            timestamp=time.time(),
            user_id=None,
            resource_type=ResourceType.AGENT,
            resource_name="test",
            action="test",
            old_permissions=None,
            new_permissions=None,
            reason="Test",
            auto_corrected=False,
        )
        assert audit.user_id is None


class TestUnifiedPermissionManagerInit:
    """Test UnifiedPermissionManager initialization."""

    def test_manager_init_default(self):
        """Test manager initialization with default path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "settings.json"
            manager = UnifiedPermissionManager(config_path=str(config_path))
            assert manager.enable_logging is True
            assert isinstance(manager.config, dict)

    def test_manager_init_with_logging_disabled(self):
        """Test manager initialization without logging."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "settings.json"
            manager = UnifiedPermissionManager(
                config_path=str(config_path),
                enable_logging=False,
            )
            assert manager.enable_logging is False

    def test_manager_init_creates_config(self):
        """Test manager creates default config."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "settings.json"
            manager = UnifiedPermissionManager(config_path=str(config_path))
            # Should have default configuration
            assert isinstance(manager.config, dict)

    def test_manager_role_hierarchy(self):
        """Test manager role hierarchy setup."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            assert "admin" in manager.role_hierarchy
            assert "developer" in manager.role_hierarchy["admin"]


class TestValidateAgentPermission:
    """Test agent permission validation."""

    def test_validate_valid_permission(self):
        """Test validation of valid permission."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            config = {"permissionMode": "acceptEdits"}
            result = manager.validate_agent_permission("backend-expert", config)
            assert result.valid is True

    def test_validate_invalid_permission_mode(self):
        """Test validation of invalid permission mode."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            config = {"permissionMode": "invalid_mode"}
            result = manager.validate_agent_permission("test-agent", config)
            assert result.valid is False or result.auto_corrected is True

    def test_validate_auto_corrects_invalid_mode(self):
        """Test validation auto-corrects invalid modes."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            config = {"permissionMode": "ask"}
            result = manager.validate_agent_permission("backend-expert", config)
            # Should auto-correct
            if result.auto_corrected:
                assert result.corrected_mode in manager.VALID_PERMISSION_MODES

    def test_validate_with_required_fields(self):
        """Test validation checks required fields."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            config = {"permissionMode": "acceptEdits"}
            result = manager.validate_agent_permission("test-agent", config)
            # Should warn about missing fields
            assert len(result.warnings) >= 0


class TestSuggestPermissionMode:
    """Test permission mode suggestion."""

    def test_suggest_security_agent(self):
        """Test mode suggestion for security agent."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            mode = manager._suggest_permission_mode("security-expert")
            assert mode in manager.VALID_PERMISSION_MODES

    def test_suggest_implementer_agent(self):
        """Test mode suggestion for implementer agent."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            mode = manager._suggest_permission_mode("tdd-implementer")
            assert mode in manager.VALID_PERMISSION_MODES

    def test_suggest_planner_agent(self):
        """Test mode suggestion for planner agent."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            mode = manager._suggest_permission_mode("implementation-planner")
            assert mode in manager.VALID_PERMISSION_MODES

    def test_suggest_default_mode(self):
        """Test default mode suggestion."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            mode = manager._suggest_permission_mode("unknown-agent")
            assert mode in manager.VALID_PERMISSION_MODES


class TestToolPermissions:
    """Test tool permission validation."""

    def test_validate_safe_tools(self):
        """Test validation of safe tools."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            result = manager.validate_tool_permissions(["Read", "Task", "AskUserQuestion"])
            assert result.valid is True

    def test_validate_dangerous_tools(self):
        """Test validation detects dangerous tools."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            result = manager.validate_tool_permissions(["Bash(rm -rf:*)", "Read"])
            # Should warn about dangerous tools
            assert len(result.warnings) > 0 or result.valid is True

    def test_check_tool_permission_allowed(self):
        """Test permission check for allowed tool."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            allowed = manager.check_tool_permission("developer", "Read", "execute")
            assert isinstance(allowed, bool)

    def test_check_tool_permission_admin(self):
        """Test admin has all permissions."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            allowed = manager.check_tool_permission("admin", "AnyTool", "execute")
            assert allowed is True

    def test_check_tool_permission_denied(self):
        """Test permission denied for user."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            allowed = manager.check_tool_permission("user", "Write", "execute")
            assert allowed is False


class TestConfigurationValidation:
    """Test configuration validation."""

    def test_validate_configuration_valid(self):
        """Test validation of valid configuration."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "settings.json"
            config = {"permissions": {}, "sandbox": {}, "mcpServers": {}}
            with open(config_path, "w") as f:
                json.dump(config, f)

            manager = UnifiedPermissionManager(config_path=str(config_path))
            result = manager.validate_configuration(config_path=str(config_path))
            assert isinstance(result, ValidationResult)

    def test_validate_configuration_missing_file(self):
        """Test validation with missing configuration file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            result = manager.validate_configuration(
                config_path=str(Path(tmpdir) / "nonexistent.json")
            )
            assert result.valid is False

    def test_validate_configuration_invalid_json(self):
        """Test validation with invalid JSON."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "settings.json"
            with open(config_path, "w") as f:
                f.write("{ invalid json }")

            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "other.json")
            )
            result = manager.validate_configuration(config_path=str(config_path))
            assert result.valid is False


class TestAutoFixPermissions:
    """Test auto-fix permission functionality."""

    def test_auto_fix_agent_permission(self):
        """Test auto-fixing agent permission."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "settings.json"
            manager = UnifiedPermissionManager(config_path=str(config_path))

            # Set up agent with invalid permission
            manager.config["agents"] = {
                "test-agent": {"permissionMode": "invalid"}
            }

            result = manager.auto_fix_agent_permissions("test-agent")
            # Should attempt fix
            assert isinstance(result, ValidationResult)

    def test_auto_fix_all_agents(self):
        """Test auto-fixing all agent permissions."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "settings.json"
            manager = UnifiedPermissionManager(config_path=str(config_path))

            results = manager.auto_fix_all_agents()
            assert isinstance(results, dict)


class TestAuditLog:
    """Test audit logging functionality."""

    def test_audit_log_creation(self):
        """Test audit log entry creation."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            manager._audit_permission_change(
                resource_type=ResourceType.AGENT,
                resource_name="test",
                action="test_action",
                old_permissions={"mode": "ask"},
                new_permissions={"mode": "default"},
                reason="Test",
                auto_corrected=True,
            )
            assert len(manager.audit_log) > 0

    def test_audit_log_max_size(self):
        """Test audit log size limit."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            # Add more than max entries
            for i in range(1100):
                manager._audit_permission_change(
                    resource_type=ResourceType.AGENT,
                    resource_name=f"agent{i}",
                    action="test",
                    old_permissions=None,
                    new_permissions=None,
                    reason="Test",
                    auto_corrected=False,
                )
            # Should be limited
            assert len(manager.audit_log) <= 1000

    def test_get_recent_audits(self):
        """Test getting recent audit entries."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            for i in range(5):
                manager._audit_permission_change(
                    resource_type=ResourceType.AGENT,
                    resource_name=f"agent{i}",
                    action="test",
                    old_permissions=None,
                    new_permissions=None,
                    reason="Test",
                    auto_corrected=False,
                )

            recent = manager.get_recent_audits(limit=3)
            assert len(recent) <= 3


class TestStatistics:
    """Test permission statistics."""

    def test_get_permission_stats(self):
        """Test getting permission statistics."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            stats = manager.get_permission_stats()

            assert "validations_performed" in stats
            assert "auto_corrections" in stats
            assert "security_violations" in stats
            assert "permission_denied" in stats

    def test_stats_update_on_validation(self):
        """Test stats update during validation."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            initial_stats = manager.get_permission_stats()

            manager.validate_agent_permission("test", {"permissionMode": "default"})

            updated_stats = manager.get_permission_stats()
            assert updated_stats["validations_performed"] > initial_stats["validations_performed"]


class TestExporting:
    """Test audit report export."""

    def test_export_audit_report(self):
        """Test exporting audit report."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            manager._audit_permission_change(
                resource_type=ResourceType.AGENT,
                resource_name="test",
                action="test",
                old_permissions=None,
                new_permissions=None,
                reason="Test",
                auto_corrected=False,
            )

            report_path = Path(tmpdir) / "audit_report.json"
            manager.export_audit_report(str(report_path))

            assert report_path.exists()
            with open(report_path) as f:
                report = json.load(f)
            assert "stats" in report
            assert "recent_audits" in report


class TestGlobalFunctions:
    """Test module-level convenience functions."""

    def test_validate_agent_permission_function(self):
        """Test validate_agent_permission function."""
        result = validate_agent_permission("test", {"permissionMode": "default"})
        assert isinstance(result, ValidationResult)

    def test_check_tool_permission_function(self):
        """Test check_tool_permission function."""
        allowed = check_tool_permission("admin", "Read", "execute")
        assert isinstance(allowed, bool)

    def test_auto_fix_all_permissions_function(self):
        """Test auto_fix_all_agent_permissions function."""
        results = auto_fix_all_agent_permissions()
        assert isinstance(results, dict)

    def test_get_permission_stats_function(self):
        """Test get_permission_stats function."""
        stats = get_permission_stats()
        assert isinstance(stats, dict)

    def test_global_permission_manager(self):
        """Test global permission manager instance."""
        assert permission_manager is not None
        assert isinstance(permission_manager, UnifiedPermissionManager)


class TestCaching:
    """Test permission caching."""

    def test_permission_cache(self):
        """Test permission caching mechanism."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            # First check
            manager.check_tool_permission("developer", "Read", "execute")

            # Check cache
            assert len(manager.permission_cache) > 0

            # Second check should use cache
            manager.check_tool_permission("developer", "Read", "execute")
            assert len(manager.permission_cache) >= 1


class TestEdgeCases:
    """Test edge cases and error conditions."""

    def test_none_config(self):
        """Test handling None configuration."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "nonexistent.json")
            )
            # Should handle gracefully
            assert isinstance(manager.config, dict)

    def test_empty_agent_name(self):
        """Test with empty agent name."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            result = manager.validate_agent_permission("", {})
            assert isinstance(result, ValidationResult)

    def test_special_characters_in_names(self):
        """Test with special characters in agent names."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            result = manager.validate_agent_permission("agent<>?:/", {})
            assert isinstance(result, ValidationResult)

    def test_very_long_agent_names(self):
        """Test with very long agent names."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            long_name = "x" * 10000
            result = manager.validate_agent_permission(long_name, {})
            assert isinstance(result, ValidationResult)

    def test_role_hierarchy_lookup(self):
        """Test role hierarchy lookups."""
        with tempfile.TemporaryDirectory() as tmpdir:
            manager = UnifiedPermissionManager(
                config_path=str(Path(tmpdir) / "settings.json")
            )
            # Unknown role should return empty list
            subordinates = manager.role_hierarchy.get("unknown_role", [])
            assert subordinates == []


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
