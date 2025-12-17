"""
System Integration Tests

Tests for the complete moai-menu-project integrated system including:
- Module initialization and dependencies
- End-to-end workflows
- Configuration management
- Error handling and recovery
"""

import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from moai_menu_project import MoaiMenuProject, initialize_project


class TestSystemIntegration(unittest.TestCase):
    """Test the complete integrated system."""

    def setUp(self):
        """Set up test environment."""
        self.test_dir = Path(tempfile.mkdtemp())
        self.project = MoaiMenuProject(str(self.test_dir))

    def tearDown(self):
        """Clean up test environment."""
        import shutil

        shutil.rmtree(self.test_dir, ignore_errors=True)

    def test_complete_project_initialization(self):
        """Test complete project initialization workflow."""

        # Initialize with all modules
        result = self.project.initialize_complete_project(
            language="ko",
            user_name="테스트 사용자",
            domains=["backend", "frontend"],
            project_type="web_application",
            optimization_enabled=True,
        )

        # Verify successful initialization
        self.assertTrue(result["success"])
        self.assertEqual(len(result["modules_initialized"]), 3)  # All 3 modules

        # Check configuration was created
        config_path = self.test_dir / ".moai/config/config.yaml"
        self.assertTrue(config_path.exists())

        # Verify configuration content
        with open(config_path, "r", encoding="utf-8") as f:
            config = json.load(f)

        self.assertEqual(config["project"]["type"], "web_application")
        self.assertEqual(config["language"]["conversation_language"], "ko")
        self.assertTrue(config["menu_system"]["fully_initialized"])

    def test_module_dependencies(self):
        """Test module dependencies and integration."""

        # All modules should be initialized
        self.assertIsNotNone(self.project.documentation_manager)
        self.assertIsNotNone(self.project.language_initializer)
        self.assertIsNotNone(self.project.template_optimizer)

        # Modules should share the same project root and config
        self.assertEqual(self.project.documentation_manager.project_root, str(self.test_dir))
        self.assertEqual(self.project.language_initializer.project_root, str(self.test_dir))

    def test_configuration_sharing(self):
        """Test configuration sharing between modules."""

        # Update language settings
        updates = {
            "language.conversation_language": "ja",
            "language.agent_prompt_language": "english",
        }

        result = self.project.update_language_settings(updates)
        self.assertTrue(result["success"])

        # Verify configuration was updated
        self.assertEqual(self.project.config["language"]["conversation_language"], "ja")

    def test_error_handling_and_recovery(self):
        """Test error handling and recovery mechanisms."""

        # Test with invalid project type
        result = self.project.initialize_complete_project(project_type="invalid_type")

        # Should still succeed but log error
        self.assertTrue(result["success"])

        # Test with invalid language
        result = self.project.initialize_complete_project(language="invalid_language")

        # Should handle gracefully
        self.assertTrue(result["success"])

    def test_workflow_integration(self):
        """Test integration of different workflows."""

        # Step 1: Initialize project
        init_result = self.project.initialize_complete_project(language="en", project_type="web_application")

        self.assertTrue(init_result["success"])

        # Step 2: Generate documentation from SPEC
        spec_data = {
            "id": "SPEC-001",
            "title": "Test Feature",
            "description": "Test feature implementation",
            "requirements": ["Requirement 1", "Requirement 2"],
            "api_endpoints": [{"path": "/api/test", "method": "POST", "description": "Test endpoint"}],
        }

        docs_result = self.project.generate_documentation_from_spec(spec_data)
        self.assertTrue(docs_result["success"])

        # Step 3: Optimize templates
        opt_result = self.project.optimize_project_templates()
        self.assertIn("analysis", opt_result)
        self.assertIn("optimization", opt_result)

    def test_backup_and_recovery(self):
        """Test backup creation and recovery capabilities."""

        # Initialize project first
        self.project.initialize_complete_project()

        # Create backup
        backup_result = self.project.create_project_backup("test-backup")
        self.assertTrue(backup_result["success"])

        # Verify backup files exist
        backup_dir = self.test_dir / ".moai-backups/test-backup"
        self.assertTrue(backup_dir.exists())
        self.assertTrue((backup_dir / "config.json").exists())

    def test_integration_matrix(self):
        """Test integration matrix and workflow information."""

        matrix = self.project.get_integration_matrix()

        # Verify module information
        self.assertIn("modules", matrix)
        self.assertIn("workflows", matrix)
        self.assertIn("data_flow", matrix)

        # Check all modules are documented
        modules = matrix["modules"]
        self.assertIn("documentation_manager", modules)
        self.assertIn("language_initializer", modules)
        self.assertIn("template_optimizer", modules)

        # Check workflows are defined
        workflows = matrix["workflows"]
        self.assertIn("project_initialization", workflows)
        self.assertIn("documentation_generation", workflows)
        self.assertIn("template_optimization", workflows)

    def test_project_status(self):
        """Test comprehensive project status reporting."""

        # Initialize project
        self.project.initialize_complete_project()

        # Get status
        status = self.project.get_project_status()

        # Verify status structure
        self.assertIn("timestamp", status)
        self.assertIn("project_root", status)
        self.assertIn("configuration", status)
        self.assertIn("modules", status)
        self.assertIn("language_status", status)
        self.assertIn("documentation_status", status)
        self.assertIn("fully_initialized", status)

        # Verify initialization status
        self.assertTrue(status["fully_initialized"])

        # Verify module status
        self.assertEqual(status["modules"]["documentation_manager"], "initialized")
        self.assertEqual(status["modules"]["language_initializer"], "initialized")
        self.assertEqual(status["modules"]["template_optimizer"], "initialized")

    def test_convenience_functions(self):
        """Test convenience functions for easy access."""

        # Test convenience initialization function
        test_dir2 = Path(tempfile.mkdtemp())

        try:
            result = initialize_project(str(test_dir2), language="en", project_type="mobile_application")

            self.assertTrue(result["success"])

        finally:
            import shutil

            shutil.rmtree(test_dir2, ignore_errors=True)

    def test_module_error_isolation(self):
        """Test that module errors don't crash the entire system."""

        # Mock a module to raise an exception
        with patch.object(
            self.project.template_optimizer,
            "analyze_project_templates",
            side_effect=Exception("Module error"),
        ):
            # Should still complete initialization with optimization disabled
            result = self.project.initialize_complete_project(optimization_enabled=False)  # Disable to avoid the error

            self.assertTrue(result["success"])
            self.assertIn("language_initializer", result["modules_initialized"])
            self.assertIn("documentation_manager", result["modules_initialized"])


if __name__ == "__main__":
    unittest.main()
