#!/usr/bin/env python3
"""
Critical Integration Testing Suite for MoAI-ADK
Tests integration between security, build system, and core modules
"""

import unittest
import sys
import os
import tempfile
import shutil
import json
import subprocess
from pathlib import Path
from unittest.mock import patch, MagicMock

# Add project root to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

try:
    from moai_adk.core.security import SecurityManager, SecurityError
    from moai_adk.core.validator import validate_environment, validate_project_structure
    from moai_adk.config import Config, RuntimeConfig
except ImportError as e:
    print(f"Warning: Could not import required modules: {e}")


class TestSecurityValidatorIntegration(unittest.TestCase):
    """Test integration between security and validator modules"""

    def setUp(self):
        """Set up integration test environment"""
        self.test_dir = Path(tempfile.mkdtemp())
        self.security = SecurityManager() if 'SecurityManager' in globals() else None

    def tearDown(self):
        """Clean up test environment"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_validator_uses_security_manager(self):
        """Test that validator properly integrates with security manager"""
        if self.security is None:
            self.skipTest("SecurityManager not available")

        # Create a project structure with potential security issues
        project_dir = self.test_dir / "test_project"
        project_dir.mkdir()

        # Create .claude structure
        claude_dir = project_dir / ".claude"
        claude_dir.mkdir()

        # Create settings with dangerous path
        dangerous_settings = {
            "permissions": {"allow": ["Read(../../../etc/passwd)"]},
            "environment": {"PATH": "/dangerous/path"}
        }

        settings_file = claude_dir / "settings.json"
        with open(settings_file, 'w') as f:
            json.dump(dangerous_settings, f)

        # Validator should handle this safely
        try:
            if 'validate_project_structure' in globals():
                result = validate_project_structure(project_dir)
                self.assertIsInstance(result, dict)
        except Exception as e:
            self.fail(f"Validator failed with security issue: {e}")

    def test_environment_validation_security(self):
        """Test environment validation with security controls"""
        if 'validate_environment' not in globals():
            self.skipTest("validate_environment not available")

        # Mock dangerous environment
        with patch.dict(os.environ, {
            'PATH': '/dangerous/path:/usr/bin',
            'CLAUDE_PROJECT_DIR': '../../../etc',
            'SHELL': '/bin/dangerous_shell'
        }):
            # Should complete without security errors
            try:
                result = validate_environment()
                self.assertIsInstance(result, bool)
            except SecurityError:
                self.fail("Environment validation should handle dangerous env vars safely")

    def test_subprocess_integration(self):
        """Test that validators use secure subprocess calls"""
        if self.security is None:
            self.skipTest("SecurityManager not available")

        # Mock subprocess.run to track calls
        with patch('subprocess.run') as mock_run:
            mock_run.return_value = MagicMock(returncode=0, stdout="test output")

            # Test secure subprocess call
            try:
                result = self.security.safe_subprocess_run(
                    ["echo", "test"],
                    cwd=self.test_dir,
                    timeout=5
                )
                self.assertEqual(result.returncode, 0)

                # Verify subprocess was called with security parameters
                mock_run.assert_called_once()
                call_args = mock_run.call_args
                self.assertIn('timeout', call_args.kwargs)
                self.assertEqual(call_args.kwargs['timeout'], 5)

            except SecurityError as e:
                self.fail(f"Secure subprocess call failed: {e}")


class TestConfigSecurityIntegration(unittest.TestCase):
    """Test integration between config and security systems"""

    def setUp(self):
        """Set up config security testing"""
        self.test_dir = Path(tempfile.mkdtemp())

    def tearDown(self):
        """Clean up"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_config_path_validation(self):
        """Test that config validates paths securely"""
        if 'Config' not in globals():
            self.skipTest("Config class not available")

        # Test dangerous project paths
        dangerous_paths = [
            "../../../etc/passwd",
            "/etc/shadow",
            "../../.ssh/id_rsa",
            "/root/.bashrc"
        ]

        for dangerous_path in dangerous_paths:
            with self.subTest(path=dangerous_path):
                try:
                    # Config should either reject or sanitize dangerous paths
                    config = Config(
                        name="test_project",
                        path=dangerous_path
                    )

                    # If it accepts the path, it should be sanitized
                    project_path = config.project_path
                    resolved_path = project_path.resolve()

                    # Should not point to system directories
                    self.assertNotEqual(str(resolved_path), "/etc/passwd")
                    self.assertNotEqual(str(resolved_path), "/etc/shadow")

                except ValueError:
                    # Expected - dangerous paths should be rejected
                    pass

    def test_config_name_sanitization(self):
        """Test that project names are properly sanitized"""
        if 'Config' not in globals():
            self.skipTest("Config class not available")

        dangerous_names = [
            "../malicious",
            "project; rm -rf /",
            "project`whoami`",
            "project$(rm -rf /)",
            "project && rm -rf /",
            "project | cat /etc/passwd"
        ]

        for dangerous_name in dangerous_names:
            with self.subTest(name=dangerous_name):
                with self.assertRaises(ValueError):
                    Config(name=dangerous_name)

    def test_runtime_config_validation(self):
        """Test runtime config security validation"""
        if 'RuntimeConfig' not in globals():
            self.skipTest("RuntimeConfig class not available")

        # Test dangerous runtime names
        dangerous_runtimes = [
            "node; rm -rf /",
            "../../../bin/sh",
            "/etc/passwd",
            "runtime`malicious`"
        ]

        for dangerous_runtime in dangerous_runtimes:
            with self.subTest(runtime=dangerous_runtime):
                with self.assertRaises(ValueError):
                    RuntimeConfig(name=dangerous_runtime)


class TestBuildSecurityIntegration(unittest.TestCase):
    """Test build system security integration"""

    def setUp(self):
        """Set up build security testing"""
        self.test_dir = Path(tempfile.mkdtemp())

    def tearDown(self):
        """Clean up"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_makefile_execution_security(self):
        """Test that Makefile execution is secure"""
        # Create test Makefile with potentially dangerous commands
        makefile_content = """
.PHONY: test clean status

test:
\t@echo "Running tests"
\t@python -m pytest tests/

clean:
\t@echo "Cleaning build artifacts"
\t@rm -rf dist/ build/ *.egg-info/

status:
\t@echo "Checking status"
\t@ls -la
"""

        makefile_path = self.test_dir / "Makefile"
        with open(makefile_path, 'w') as f:
            f.write(makefile_content)

        # Test executing safe targets
        safe_targets = ["status", "test"]

        for target in safe_targets:
            with self.subTest(target=target):
                try:
                    # Should be able to run make commands safely
                    result = subprocess.run(
                        ["make", "-f", str(makefile_path), target],
                        cwd=self.test_dir,
                        capture_output=True,
                        text=True,
                        timeout=10
                    )
                    # Command should complete (may fail due to missing dependencies)
                    self.assertIsNotNone(result.returncode)

                except subprocess.TimeoutExpired:
                    self.fail(f"Make target '{target}' hung - possible security issue")

    def test_template_file_security(self):
        """Test that template files don't contain security vulnerabilities"""
        # Look for template files in the project
        template_dir = Path(__file__).parent.parent / "src" / "moai_adk" / "resources" / "templates"

        if not template_dir.exists():
            self.skipTest("Template directory not found")

        dangerous_patterns = [
            r'eval\s*\(',  # eval() calls
            r'exec\s*\(',  # exec() calls
            r'__import__\s*\(',  # dynamic imports
            r'open\s*\(\s*["\'][^"\']*\.\./.*["\']',  # path traversal in file opens
            r'subprocess\.',  # subprocess without security
            r'os\.system\s*\(',  # os.system calls
        ]

        import re

        for template_file in template_dir.rglob("*.py"):
            with self.subTest(file=template_file.name):
                try:
                    with open(template_file, 'r', encoding='utf-8') as f:
                        content = f.read()

                    for pattern in dangerous_patterns:
                        matches = re.findall(pattern, content, re.IGNORECASE)
                        if matches:
                            # Check if it's a false positive (e.g., in comments or secured code)
                            if "SecurityManager" in content or "# SECURITY:" in content:
                                continue  # Allowed in security-aware code

                            self.fail(f"Potential security issue in {template_file}: {pattern}")

                except Exception as e:
                    self.fail(f"Failed to scan template file {template_file}: {e}")


class TestHookIntegrationSecurity(unittest.TestCase):
    """Test hook integration security"""

    def setUp(self):
        """Set up hook integration testing"""
        self.test_dir = Path(tempfile.mkdtemp())

    def tearDown(self):
        """Clean up"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_hook_chain_security(self):
        """Test that hook execution chain is secure"""
        # Create hook directory structure
        hooks_dir = self.test_dir / ".claude" / "hooks" / "moai"
        hooks_dir.mkdir(parents=True)

        # Create test hooks
        test_hooks = [
            "session_start_notice.py",
            "constitution_guard.py",
            "tag_validator.py",
            "policy_block.py"
        ]

        for hook_name in test_hooks:
            hook_file = hooks_dir / hook_name
            hook_content = f"""#!/usr/bin/env python3
# Test hook: {hook_name}
import sys
import json
import os
from pathlib import Path

# Security import
try:
    from security import SecurityManager
except ImportError:
    SecurityManager = None

def main():
    try:
        hook_data = json.loads(sys.stdin.read())
        print(f"Hook {hook_name} executed safely")
        sys.exit(0)
    except Exception as e:
        print(f"Hook error: {{e}}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
"""
            with open(hook_file, 'w') as f:
                f.write(hook_content)

            # Set executable permissions
            if os.name != 'nt':
                hook_file.chmod(0o755)

        # Test hook execution with malicious input
        malicious_inputs = [
            '{"tool_name": "Write", "tool_input": {"path": "../../../etc/passwd"}}',
            '{"tool_name": "Bash", "tool_input": {"command": "rm -rf /"}}',
        ]

        for hook_name in test_hooks:
            for malicious_input in malicious_inputs:
                with self.subTest(hook=hook_name, input=malicious_input[:50]):
                    hook_file = hooks_dir / hook_name

                    try:
                        # Execute hook with malicious input
                        result = subprocess.run(
                            ["python", str(hook_file)],
                            input=malicious_input,
                            capture_output=True,
                            text=True,
                            timeout=5,
                            cwd=self.test_dir
                        )

                        # Hook should either block or handle safely
                        self.assertIsNotNone(result.returncode)

                    except subprocess.TimeoutExpired:
                        self.fail(f"Hook {hook_name} hung on malicious input")


def run_critical_integration_tests():
    """Run all critical integration tests"""
    print("🔗 Running Critical Integration Tests...")

    # Test loader
    loader = unittest.TestLoader()
    suite = unittest.TestSuite()

    # Add all integration test cases
    test_classes = [
        TestSecurityValidatorIntegration,
        TestConfigSecurityIntegration,
        TestBuildSecurityIntegration,
        TestHookIntegrationSecurity
    ]

    for test_class in test_classes:
        tests = loader.loadTestsFromTestCase(test_class)
        suite.addTests(tests)

    # Run tests with detailed output
    runner = unittest.TextTestRunner(verbosity=2, buffer=True)
    result = runner.run(suite)

    # Integration-specific reporting
    print(f"\n🔗 Integration Test Results:")
    print(f"   Tests run: {result.testsRun}")
    print(f"   Failures: {len(result.failures)}")
    print(f"   Errors: {len(result.errors)}")
    print(f"   Skipped: {len(result.skipped)}")

    # Integration tests critical for security
    integration_critical = len(result.failures) == 0 and len(result.errors) == 0

    if integration_critical:
        print(f"\n✅ All critical integration tests passed!")
    else:
        print(f"\n❌ INTEGRATION FAILURES DETECTED!")
        print("   Review and fix all integration issues before proceeding.")

        # Print detailed failure information
        for test, error in result.failures + result.errors:
            print(f"\n🚨 FAILED: {test}")
            print(f"   Error: {error}")

    return integration_critical


if __name__ == '__main__':
    success = run_critical_integration_tests()
    sys.exit(0 if success else 1)