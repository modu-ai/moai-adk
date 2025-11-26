"""
Unit Tests for Unified Permission Manager

Production-ready test suite covering all permission validation, auto-correction,
and security features of the Unified Permission Manager.

Author: MoAI-ADK Core Team
Version: 1.0.0
"""

import json
import os
import sys
import tempfile
import unittest

# Add the src directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "src"))

from moai_adk.core.unified_permission_manager import (
    PermissionAudit,
    PermissionMode,
    PermissionSeverity,
    ResourceType,
    UnifiedPermissionManager,
    ValidationResult,
    auto_fix_all_agent_permissions,
    check_tool_permission,
    get_permission_stats,
    validate_agent_permission,
)


class TestUnifiedPermissionManager(unittest.TestCase):
    """Comprehensive test suite for Unified Permission Manager"""

    def setUp(self):
        """Set up test fixtures"""
        # Create a temporary configuration file
        self.temp_dir = tempfile.mkdtemp()
        self.config_path = os.path.join(self.temp_dir, "test_settings.json")

        # Create test configuration with permission issues
        self.test_config = {
            "agents": {
                "backend-expert": {
                    "permissionMode": "ask",  # Invalid mode
                    "description": "Backend development expert",
                    "systemPrompt": "Backend development assistance",
                },
                "security-expert": {
                    "permissionMode": "auto",  # Invalid mode
                    "description": "Security and compliance expert",
                },
                "api-designer": {
                    "permissionMode": "plan",  # Valid mode
                    "description": "API design and documentation expert",
                },
            },
            "permissions": {
                "allowedTools": ["Task", "Read", "Write", "Bash"],
                "deniedTools": ["Bash(rm -rf:*)", "Bash(sudo:*)"],
            },
            "sandbox": {"allowUnsandboxedCommands": False},
        }

        # Write test configuration
        with open(self.config_path, "w") as f:
            json.dump(self.test_config, f)

        # Create permission manager instance
        self.manager = UnifiedPermissionManager(config_path=self.config_path, enable_logging=False)

    def tearDown(self):
        """Clean up test fixtures"""
        import shutil

        shutil.rmtree(self.temp_dir, ignore_errors=True)

    def test_invalid_permission_mode_correction(self):
        """Test correction of invalid permission modes"""
        # Test agent with 'ask' permission mode (invalid)
        agent_config = {"permissionMode": "ask", "description": "Test agent"}

        result = self.manager.validate_agent_permission("test-agent", agent_config)

        self.assertTrue(result.auto_corrected)
        self.assertIsNotNone(result.corrected_mode)
        self.assertIn(result.corrected_mode, [mode.value for mode in PermissionMode])
        self.assertGreater(len(result.errors), 0)

    def test_valid_permission_mode_preservation(self):
        """Test that valid permission modes are preserved"""
        # Test agent with valid permission mode
        agent_config = {"permissionMode": "acceptEdits", "description": "Test agent"}

        result = self.manager.validate_agent_permission("test-agent", agent_config)

        self.assertFalse(result.auto_corrected)
        self.assertIsNone(result.corrected_mode)
        self.assertEqual(agent_config["permissionMode"], "acceptEdits")

    def test_permission_mode_suggestion_logic(self):
        """Test permission mode suggestion logic based on agent names"""
        test_cases = [
            ("security-expert", PermissionMode.PLAN),
            ("backend-expert", PermissionMode.ACCEPT_EDITS),
            ("api-designer", PermissionMode.PLAN),
            ("tdd-implementer", PermissionMode.ACCEPT_EDITS),
            ("unknown-agent", PermissionMode.DEFAULT),
        ]

        for agent_name, expected_mode in test_cases:
            with self.subTest(agent_name=agent_name):
                suggested_mode = self.manager._suggest_permission_mode(agent_name)
                self.assertEqual(suggested_mode, expected_mode.value)

    def test_tool_permission_checking(self):
        """Test tool permission checking with role hierarchy"""
        # Test admin role (should have all permissions)
        self.assertTrue(self.manager.check_tool_permission("admin", "Task", "execute"))
        self.assertTrue(self.manager.check_tool_permission("admin", "Bash", "execute"))

        # Test developer role
        self.assertTrue(self.manager.check_tool_permission("developer", "Task", "execute"))
        self.assertTrue(self.manager.check_tool_permission("developer", "Read", "execute"))

        # Test user role (limited permissions)
        self.assertTrue(self.manager.check_tool_permission("user", "Task", "execute"))
        self.assertTrue(self.manager.check_tool_permission("user", "Read", "execute"))
        self.assertFalse(self.manager.check_tool_permission("user", "Write", "execute"))

    def test_permission_caching(self):
        """Test that permission results are cached for performance"""
        # First check should populate cache
        result1 = self.manager.check_tool_permission("developer", "Task", "execute")

        # Second check should use cache
        result2 = self.manager.check_tool_permission("developer", "Task", "execute")

        self.assertEqual(result1, result2)

        stats = self.manager.get_permission_stats()
        self.assertGreater(stats["cached_permissions"], 0)

    def test_dangerous_tool_detection(self):
        """Test detection of dangerous tools in allowed tools"""
        dangerous_tools = ["Bash(rm -rf:*)", "Bash(sudo:*)", "Bash(chmod -R 777:*)", "Bash(git push --force:*)"]

        for tool in dangerous_tools:
            with self.subTest(tool=tool):
                result = self.manager.validate_tool_permissions([tool])
                self.assertGreater(len(result.warnings), 0)

    def test_configuration_validation(self):
        """Test configuration file validation"""
        # Test with valid configuration
        result = self.manager.validate_configuration(self.config_path)
        self.assertTrue(result.valid)

        # Test with invalid JSON
        invalid_config_path = os.path.join(self.temp_dir, "invalid.json")
        with open(invalid_config_path, "w") as f:
            f.write("{ invalid json }")

        result = self.manager.validate_configuration(invalid_config_path)
        self.assertFalse(result.valid)
        self.assertEqual(result.severity, PermissionSeverity.CRITICAL)

    def test_sandbox_security_validation(self):
        """Test sandbox security validation"""
        # Test secure sandbox configuration
        secure_config = {"sandbox": {"allowUnsandboxedCommands": False}}
        self.assertTrue(self.manager._validate_sandbox_settings(secure_config))

        # Test insecure sandbox configuration
        insecure_config = {
            "sandbox": {"allowUnsandboxedCommands": True, "validatedCommands": ["rm -rf /", "sudo rm -rf"]}
        }
        self.assertFalse(self.manager._validate_sandbox_settings(insecure_config))

    def test_mcp_server_validation(self):
        """Test MCP server security validation"""
        # Test secure MCP configuration
        secure_config = {"mcpServers": {"context7": {"command": "npx", "args": ["-y", "@upstash/context7-mcp@latest"]}}}
        self.assertTrue(self.manager._validate_mcp_servers(secure_config))

        # Test insecure MCP configuration
        insecure_config = {
            "mcpServers": {"dangerous-server": {"command": "some-command", "args": ["--insecure", "--allow-all"]}}
        }
        self.assertFalse(self.manager._validate_mcp_servers(insecure_config))

    def test_auto_fix_single_agent(self):
        """Test auto-fixing a single agent's permissions"""
        agent_name = "backend-expert"

        # Get initial state
        original_config = self.manager.config["agents"][agent_name].copy()

        # Auto-fix the agent
        result = self.manager.auto_fix_agent_permissions(agent_name)

        # Verify correction
        self.assertTrue(result.auto_corrected)
        self.assertIsNotNone(result.corrected_mode)

        # Verify configuration was updated
        updated_config = self.manager.config["agents"][agent_name]
        self.assertNotEqual(original_config["permissionMode"], updated_config["permissionMode"])

    def test_auto_fix_all_agents(self):
        """Test auto-fixing all agent permissions"""
        # Fix all agents
        results = self.manager.auto_fix_all_agents()

        # Verify all problematic agents were fixed
        self.assertGreater(len(results), 0)

        fixed_count = sum(1 for result in results.values() if result.auto_corrected)
        self.assertGreater(fixed_count, 0)

        # Verify all agents now have valid permission modes
        for agent_name, agent_config in self.manager.config["agents"].items():
            permission_mode = agent_config.get("permissionMode", "default")
            self.assertIn(permission_mode, [mode.value for mode in PermissionMode])

    def test_audit_logging(self):
        """Test audit logging functionality"""
        # Trigger a permission change
        self.manager.auto_fix_agent_permissions("backend-expert")

        # Check audit log
        audits = self.manager.get_recent_audits()
        self.assertGreater(len(audits), 0)

        # Verify audit entry structure
        audit = audits[-1]
        self.assertIsInstance(audit, PermissionAudit)
        self.assertEqual(audit.resource_type, ResourceType.AGENT)
        self.assertEqual(audit.resource_name, "backend-expert")
        self.assertTrue(audit.auto_corrected)

    def test_permission_statistics(self):
        """Test permission management statistics"""
        # Perform some operations
        self.manager.validate_agent_permission("test-agent", {"permissionMode": "invalid"})
        self.manager.check_tool_permission("user", "Write", "execute")

        # Check statistics
        stats = self.manager.get_permission_stats()

        self.assertIn("validations_performed", stats)
        self.assertIn("auto_corrections", stats)
        self.assertIn("permission_denied", stats)
        self.assertIn("cached_permissions", stats)

        self.assertGreater(stats["validations_performed"], 0)

    def test_role_hierarchy_inheritance(self):
        """Test role hierarchy and permission inheritance"""
        # Admin should inherit all subordinate permissions
        self.assertTrue(self.manager.check_tool_permission("admin", "Write", "execute"))
        self.assertTrue(self.manager.check_tool_permission("admin", "Bash", "execute"))

        # Developer should inherit user permissions but not admin-only ones
        self.assertTrue(self.manager.check_tool_permission("developer", "Task", "execute"))
        self.assertTrue(self.manager.check_tool_permission("developer", "Read", "execute"))

    def test_configuration_backup_and_save(self):
        """Test configuration backup and save functionality"""
        # Make a change that triggers a save
        self.manager.auto_fix_agent_permissions("security-expert")

        # Verify backup was created
        backup_files = [f for f in os.listdir(self.temp_dir) if f.startswith("test_settings.json.backup.")]
        self.assertGreater(len(backup_files), 0)

    def test_missing_agent_configuration_creation(self):
        """Test creation of default configurations for missing agents"""
        # Get initial agent count
        initial_count = len(self.manager.config.get("agents", {}))

        # Auto-fix all agents (should create missing ones)
        results = self.manager.auto_fix_all_agents()

        # Should have created configurations for debug log agents
        final_count = len(self.manager.config.get("agents", {}))
        self.assertGreater(final_count, initial_count)

        # Check that some agents were auto-created
        auto_created_count = sum(
            1
            for result in results.values()
            if result.auto_corrected and "Created default configuration" in str(result.warnings)
        )
        self.assertGreater(auto_created_count, 0)


class TestConvenienceFunctions(unittest.TestCase):
    """Test convenience functions for easy integration"""

    def setUp(self):
        """Set up test fixtures"""
        self.agent_config = {"permissionMode": "invalid", "description": "Test agent"}

    def test_validate_agent_permission_function(self):
        """Test convenience function for agent permission validation"""
        result = validate_agent_permission("test-agent", self.agent_config)
        self.assertIsInstance(result, ValidationResult)
        self.assertTrue(result.auto_corrected)

    def test_check_tool_permission_function(self):
        """Test convenience function for tool permission checking"""
        result = check_tool_permission("admin", "Task", "execute")
        self.assertIsInstance(result, bool)
        self.assertTrue(result)

    def test_auto_fix_all_agent_permissions_function(self):
        """Test convenience function for fixing all agent permissions"""
        results = auto_fix_all_agent_permissions()
        self.assertIsInstance(results, dict)

    def test_get_permission_stats_function(self):
        """Test convenience function for getting statistics"""
        stats = get_permission_stats()
        self.assertIsInstance(stats, dict)
        self.assertIn("validations_performed", stats)


class TestSecurityCompliance(unittest.TestCase):
    """Test security compliance features"""

    def setUp(self):
        """Set up test fixtures"""
        self.temp_dir = tempfile.mkdtemp()
        self.config_path = os.path.join(self.temp_dir, "security_test.json")

        # Create test configuration with security issues
        self.insecure_config = {
            "permissions": {
                "allowedTools": ["Task", "Bash(rm -rf:*)", "Bash(sudo rm -rf *)"],
                "deniedTools": [],  # Missing dangerous tool denials
            },
            "sandbox": {"allowUnsandboxedCommands": True, "validatedCommands": ["rm -rf /", "sudo rm -rf /"]},
            "mcpServers": {"insecure-server": {"command": "server", "args": ["--insecure", "--disable-ssl"]}},
        }

    def tearDown(self):
        """Clean up test fixtures"""
        import shutil

        shutil.rmtree(self.temp_dir, ignore_errors=True)

    def test_file_permission_validation(self):
        """Test file permission validation"""
        # Test configuration missing dangerous tool denials
        with open(self.config_path, "w") as f:
            json.dump(self.insecure_config, f)

        manager = UnifiedPermissionManager(config_path=self.config_path, enable_logging=False)
        result = manager.validate_configuration(self.config_path)

        # Should fail security validation
        self.assertFalse(result.valid)

    def test_required_tools_validation(self):
        """Test that essential tools are required"""
        config_missing_essentials = {"permissions": {"allowedTools": []}}  # Missing essential tools

        result = UnifiedPermissionManager._validate_allowed_tools(None, config_missing_essentials)
        self.assertFalse(result)

    def test_comprehensive_security_validation(self):
        """Test comprehensive security validation"""
        with open(self.config_path, "w") as f:
            json.dump(self.insecure_config, f)

        manager = UnifiedPermissionManager(config_path=self.config_path, enable_logging=False)

        # All security checks should fail
        self.assertFalse(manager._validate_file_permissions(self.insecure_config))
        self.assertFalse(manager._validate_sandbox_settings(self.insecure_config))
        self.assertFalse(manager._validate_mcp_servers(self.insecure_config))


def run_integration_tests():
    """Run integration tests with real-world scenarios"""
    print("\n🔧 Running Integration Tests...")
    print("-" * 40)

    # Create temporary configuration for integration testing
    temp_dir = tempfile.mkdtemp()
    config_path = os.path.join(temp_dir, "integration_test.json")

    try:
        # Test scenario: Configuration with multiple permission issues
        problematic_config = {
            "agents": {
                "backend-expert": {"permissionMode": "ask"},
                "security-expert": {"permissionMode": "auto"},
                "api-designer": {"permissionMode": "invalid"},
                "frontend-expert": {"permissionMode": ""},
                "database-expert": {},  # Missing permissionMode
            },
            "permissions": {"allowedTools": ["Task", "Bash(rm -rf:*)"], "deniedTools": []},
            "sandbox": {"allowUnsandboxedCommands": True},
        }

        with open(config_path, "w") as f:
            json.dump(problematic_config, f)

        print("1. Creating UnifiedPermissionManager with problematic configuration...")
        manager = UnifiedPermissionManager(config_path=config_path, enable_logging=False)

        print("2. Auto-fixing all agent permissions...")
        results = manager.auto_fix_all_agents()

        fixed_count = sum(1 for result in results.values() if result.auto_corrected)
        print(f"   Fixed {fixed_count} agent configurations")

        print("3. Validating corrected configuration...")
        validation_result = manager.validate_configuration(config_path)
        print(f"   Configuration valid: {validation_result.valid}")

        print("4. Checking permission statistics...")
        stats = manager.get_permission_stats()
        print(f"   Total validations: {stats['validations_performed']}")
        print(f"   Auto corrections: {stats['auto_corrections']}")

        print("5. Testing role-based permissions...")
        test_results = [
            ("admin", "Bash", manager.check_tool_permission("admin", "Bash", "execute")),
            ("user", "Write", manager.check_tool_permission("user", "Write", "execute")),
            ("developer", "Read", manager.check_tool_permission("developer", "Read", "execute")),
        ]

        for role, tool, allowed in test_results:
            print(f"   {role} can use {tool}: {allowed}")

        print("6. Exporting audit report...")
        audit_path = os.path.join(temp_dir, "audit_report.json")
        manager.export_audit_report(audit_path)
        print(f"   Audit report exported to: {audit_path}")

        print("\n✅ Integration tests completed successfully!")

    finally:
        import shutil

        shutil.rmtree(temp_dir, ignore_errors=True)


if __name__ == "__main__":
    # Run unit tests
    print("🚀 Running Unified Permission Manager Unit Tests...")
    print("=" * 60)

    unittest.main(verbosity=2, exit=False)

    # Run integration tests
    run_integration_tests()
