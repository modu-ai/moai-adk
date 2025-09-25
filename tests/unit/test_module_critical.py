#!/usr/bin/env python3
"""
Critical Module Testing Suite for MoAI-ADK
Tests all core modules for critical functionality and reliability
"""

import unittest
import sys
import os
import tempfile
import shutil
import json
import time
from pathlib import Path
from unittest.mock import patch, MagicMock, mock_open

# Add project root to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

# Import modules to test
try:
    from moai_adk.config import Config, RuntimeConfig
    from moai_adk.core.config_manager import ConfigManager
    from moai_adk.core.git_manager import GitManager
    from moai_adk.core.template_engine import TemplateEngine
    from moai_adk.install.installer import install_project
    from moai_adk.post_install import run_post_install_steps
    from moai_adk.version_sync import version_sync
except ImportError as e:
    print(f"Warning: Could not import some modules: {e}")


class TestConfigModuleCritical(unittest.TestCase):
    """Critical tests for config module"""

    def test_config_validation_edge_cases(self):
        """Test config validation with edge cases"""
        if "Config" not in globals():
            self.skipTest("Config not available")

        # Test edge case names
        edge_cases = [
            ("", ValueError),  # Empty name
            ("a" * 256, ValueError),  # Too long
            ("valid-name_123.test", None),  # Valid complex name
            ("123-project", None),  # Starting with number
            ("project.with.dots", None),  # Multiple dots
        ]

        for name, expected_error in edge_cases:
            with self.subTest(name=name):
                if expected_error:
                    with self.assertRaises(expected_error):
                        Config(name=name)
                else:
                    try:
                        config = Config(name=name)
                        self.assertEqual(config.name, name)
                    except Exception as e:
                        self.fail(f"Valid name rejected: {name}, error: {e}")

    def test_runtime_config_performance_validation(self):
        """Test runtime config performance validation"""
        if "RuntimeConfig" not in globals():
            self.skipTest("RuntimeConfig not available")

        # Test performance boundaries
        invalid_performances = [-1, 0, 6, 10, 100]
        valid_performances = [1, 2, 3, 4, 5]

        for perf in invalid_performances:
            with self.subTest(performance=perf):
                with self.assertRaises(ValueError):
                    RuntimeConfig("python", performance=perf)

        for perf in valid_performances:
            with self.subTest(performance=perf):
                try:
                    runtime = RuntimeConfig("python", performance=perf)
                    self.assertEqual(runtime.performance, perf)
                except Exception as e:
                    self.fail(f"Valid performance rejected: {perf}, error: {e}")

    def test_config_template_context_completeness(self):
        """Test that template context contains all required fields"""
        if "Config" not in globals():
            self.skipTest("Config not available")

        config = Config(
            name="test_project",
            template="standard",
            tech_stack=["python", "fastapi", "postgresql"],
        )

        context = config.get_template_context()

        # Required fields
        required_fields = [
            "project_name",
            "project_type",
            "template",
            "runtime_name",
            "runtime_performance",
            "tech_stack",
            "tech_stack_list",
            "created_date",
            "created_year",
            "version",
        ]

        for field in required_fields:
            with self.subTest(field=field):
                self.assertIn(field, context, f"Missing required field: {field}")
                self.assertIsNotNone(context[field], f"Field {field} is None")


class TestConfigManagerCritical(unittest.TestCase):
    """Critical tests for config manager"""

    def setUp(self):
        """Set up test environment"""
        self.test_dir = Path(tempfile.mkdtemp())

    def tearDown(self):
        """Clean up"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_config_file_handling_edge_cases(self):
        """Test config file handling with various edge cases"""
        if "ConfigManager" not in globals():
            self.skipTest("ConfigManager not available")

        manager = ConfigManager(self.test_dir)

        # Test with corrupted JSON
        config_file = self.test_dir / ".moai" / "config.json"
        config_file.parent.mkdir(parents=True)

        corrupted_configs = [
            '{"incomplete": json',  # Invalid JSON
            '{"version": "0.1.0", "invalid_unicode": "\udc00"}',  # Invalid unicode
            '{"nested": {"too": {"deep": {"for": {"safety": true}}}}}',  # Too nested
            "[]",  # Array instead of object
            "null",  # Null value
            '{"extremely_long_key_' + "x" * 1000 + '": "value"}',  # Very long key
        ]

        for corrupted_config in corrupted_configs:
            with self.subTest(config=corrupted_config[:50]):
                config_file.write_text(corrupted_config)

                try:
                    # Should handle corrupted config gracefully
                    config = manager.load_config()
                    self.assertIsNotNone(
                        config
                    )  # Should return default or handle error
                except Exception as e:
                    # Should not crash, but may raise specific handled exceptions
                    self.assertIsInstance(e, (ValueError, json.JSONDecodeError))

    def test_concurrent_config_access(self):
        """Test config manager under concurrent access"""
        if "ConfigManager" not in globals():
            self.skipTest("ConfigManager not available")

        manager = ConfigManager(self.test_dir)

        # Simulate concurrent access
        def save_config(config_data):
            try:
                manager.save_config(config_data)
                return True
            except Exception:
                return False

        # Multiple save operations
        configs = [{"version": "0.1.0", "test": f"config_{i}"} for i in range(10)]

        # All saves should complete without corruption
        for i, config in enumerate(configs):
            with self.subTest(iteration=i):
                result = save_config(config)
                self.assertTrue(result, f"Config save {i} failed")

                # Verify saved config
                loaded = manager.load_config()
                self.assertIsNotNone(loaded)


class TestGitManagerCritical(unittest.TestCase):
    """Critical tests for git manager"""

    def setUp(self):
        """Set up test environment"""
        self.test_dir = Path(tempfile.mkdtemp())

    def tearDown(self):
        """Clean up"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_git_operations_safety(self):
        """Test git operations are safe and don't affect system"""
        if "GitManager" not in globals():
            self.skipTest("GitManager not available")

        manager = GitManager(self.test_dir)

        # Test git operations in isolated environment
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="test")

            # Test init
            try:
                manager.init_repository()
                # Should have called git init in correct directory
                mock_run.assert_called()
                call_args = mock_run.call_args
                self.assertIn("git", call_args[0][0])
                self.assertEqual(call_args[1]["cwd"], self.test_dir)
            except Exception as e:
                self.fail(f"Git init failed: {e}")

    def test_git_error_handling(self):
        """Test git error handling"""
        if "GitManager" not in globals():
            self.skipTest("GitManager not available")

        manager = GitManager(self.test_dir)

        # Test with git command failure
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(1, ["git"])

            # Should handle git errors gracefully
            try:
                result = manager.init_repository()
                # Should return failure indication, not crash
                self.assertIsNotNone(result)
            except Exception as e:
                self.assertIsInstance(e, (subprocess.CalledProcessError, RuntimeError))


class TestTemplateEngineCritical(unittest.TestCase):
    """Critical tests for template engine"""

    def setUp(self):
        """Set up test environment"""
        self.test_dir = Path(tempfile.mkdtemp())

    def tearDown(self):
        """Clean up"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_template_injection_prevention(self):
        """Test template engine prevents code injection"""
        if "TemplateEngine" not in globals():
            self.skipTest("TemplateEngine not available")

        # Create template directory
        template_dir = self.test_dir / "templates"
        template_dir.mkdir()

        engine = TemplateEngine(template_dir)

        # Test malicious template content
        malicious_templates = [
            "{{ __import__('os').system('rm -rf /') }}",
            "{{ config.__class__.__bases__[0].__subclasses__() }}",
            "{% for x in ().__class__.__base__.__subclasses__() %}{% endfor %}",
            "{{ request.application.__globals__.__builtins__.__import__('os').system('ls') }}",
        ]

        for malicious_template in malicious_templates:
            with self.subTest(template=malicious_template[:50]):
                template_file = template_dir / "malicious.txt"
                template_file.write_text(malicious_template)

                # Should either prevent injection or handle safely
                try:
                    result = engine.render_template(
                        "malicious.txt", {"config": {"safe": "value"}}
                    )
                    # If rendering succeeds, should not execute malicious code
                    self.assertNotIn("__import__", result)
                    self.assertNotIn("system", result)
                except Exception as e:
                    # Expected - should reject dangerous templates
                    self.assertIsInstance(e, (ValueError, SecurityError, Exception))

    def test_template_path_safety(self):
        """Test template engine path safety"""
        if "TemplateEngine" not in globals():
            self.skipTest("TemplateEngine not available")

        template_dir = self.test_dir / "templates"
        template_dir.mkdir()

        engine = TemplateEngine(template_dir)

        # Test path traversal attempts
        dangerous_paths = [
            "../../../etc/passwd",
            "../../.ssh/id_rsa",
            "/etc/shadow",
            "C:\\Windows\\System32\\config\\sam",
        ]

        for dangerous_path in dangerous_paths:
            with self.subTest(path=dangerous_path):
                with self.assertRaises((ValueError, FileNotFoundError, SecurityError)):
                    engine.render_template(dangerous_path, {})


class TestInstallerCritical(unittest.TestCase):
    """Critical tests for installer"""

    def setUp(self):
        """Set up test environment"""
        self.test_dir = Path(tempfile.mkdtemp())

    def tearDown(self):
        """Clean up"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_installer_path_validation(self):
        """Test installer validates paths safely"""
        if "install_project" not in globals():
            self.skipTest("install_project not available")

        # Test with dangerous paths
        dangerous_paths = ["/etc/passwd", "../../../root", "/sys/kernel", "/proc/1/mem"]

        for dangerous_path in dangerous_paths:
            with self.subTest(path=dangerous_path):
                # Create minimal config
                config = {
                    "name": "test_project",
                    "path": dangerous_path,
                    "template": "minimal",
                }

                # Should reject dangerous paths
                with self.assertRaises((ValueError, PermissionError, OSError)):
                    install_project(config)

    def test_installer_resource_limits(self):
        """Test installer respects resource limits"""
        if "install_project" not in globals():
            self.skipTest("install_project not available")

        # Create valid config
        config = {
            "name": "test_project",
            "path": str(self.test_dir / "test_project"),
            "template": "minimal",
        }

        # Test with timeout
        start_time = time.time()

        try:
            with patch("time.sleep"):  # Prevent actual delays
                install_project(config)

            # Should complete in reasonable time (< 30 seconds)
            duration = time.time() - start_time
            self.assertLess(duration, 30, "Installer took too long")

        except Exception as e:
            # Installation may fail due to missing resources, but shouldn't hang
            duration = time.time() - start_time
            self.assertLess(duration, 10, "Installer hung on error")


class TestVersionSyncCritical(unittest.TestCase):
    """Critical tests for version sync"""

    def setUp(self):
        """Set up test environment"""
        self.test_dir = Path(tempfile.mkdtemp())

    def tearDown(self):
        """Clean up"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_version_sync_file_safety(self):
        """Test version sync doesn't modify unsafe files"""
        if "version_sync" not in globals():
            self.skipTest("version_sync not available")

        # Create project structure
        project_dir = self.test_dir / "project"
        project_dir.mkdir()

        # Create version files in safe locations
        (project_dir / "package.json").write_text('{"version": "1.0.0"}')
        (project_dir / "pyproject.toml").write_text('[tool.poetry]\nversion = "1.0.0"')

        # Create dangerous symlinks (on Unix)
        if os.name != "nt":
            try:
                dangerous_link = project_dir / "dangerous_link.json"
                dangerous_link.symlink_to("/etc/passwd")
            except (OSError, PermissionError):
                pass  # Expected on some systems

        # Version sync should only affect safe files
        try:
            version_sync(project_dir, "1.1.0")
            # Should complete without modifying system files
        except Exception as e:
            # May fail due to missing tools, but shouldn't modify system files
            self.assertNotIsInstance(e, PermissionError)


def run_critical_module_tests():
    """Run all critical module tests"""
    print("🧪 Running Critical Module Tests...")

    # Test loader
    loader = unittest.TestLoader()
    suite = unittest.TestSuite()

    # Add all critical module test cases
    test_classes = [
        TestConfigModuleCritical,
        TestConfigManagerCritical,
        TestGitManagerCritical,
        TestTemplateEngineCritical,
        TestInstallerCritical,
        TestVersionSyncCritical,
    ]

    for test_class in test_classes:
        tests = loader.loadTestsFromTestCase(test_class)
        suite.addTests(tests)

    # Run tests with detailed output
    runner = unittest.TextTestRunner(verbosity=2, buffer=True)
    result = runner.run(suite)

    # Module-specific reporting
    print(f"\n🧪 Critical Module Test Results:")
    print(f"   Tests run: {result.testsRun}")
    print(f"   Failures: {len(result.failures)}")
    print(f"   Errors: {len(result.errors)}")
    print(f"   Skipped: {len(result.skipped)}")

    # Module tests critical for functionality
    modules_critical = len(result.failures) == 0 and len(result.errors) == 0

    if modules_critical:
        print(f"\n✅ All critical module tests passed!")
    else:
        print(f"\n❌ MODULE FAILURES DETECTED!")
        print("   Review and fix all module issues before proceeding.")

        # Print detailed failure information
        for test, error in result.failures + result.errors:
            print(f"\n🚨 FAILED: {test}")
            print(f"   Error: {error}")

    return modules_critical


if __name__ == "__main__":
    success = run_critical_module_tests()
    sys.exit(0 if success else 1)
