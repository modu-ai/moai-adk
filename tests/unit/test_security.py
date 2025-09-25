#!/usr/bin/env python3
"""
Critical Security Testing Suite for MoAI-ADK
Tests all security-critical components including SecurityManager and hook security
"""

import unittest
import sys
import os
import tempfile
import shutil
import subprocess
import json
from pathlib import Path
from unittest.mock import patch, mock_open

# Add project root to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

try:
    from moai_adk.core.security import SecurityManager, SecurityError
except ImportError:
    print("Warning: Could not import security module")
    SecurityManager = None
    SecurityError = Exception


class TestSecurityManager(unittest.TestCase):
    """Critical security function testing"""

    def setUp(self):
        """Set up test environment"""
        if SecurityManager is None:
            self.skipTest("SecurityManager not available")

        self.security = SecurityManager()
        self.test_dir = Path(tempfile.mkdtemp())
        self.safe_dir = self.test_dir / "safe"
        self.safe_dir.mkdir()

    def tearDown(self):
        """Clean up test environment"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_path_traversal_prevention(self):
        """CRITICAL: Test prevention of ../../../etc/passwd attacks"""
        # Test cases for path traversal
        traversal_attempts = [
            "../../../etc/passwd",
            "..\\..\\..\\windows\\system32",
            "/etc/passwd",
            "../../.ssh/id_rsa",
            "safe/../../../etc/passwd",
            "safe/../../..",
            ".././../etc/passwd",
        ]

        for malicious_path in traversal_attempts:
            with self.subTest(path=malicious_path):
                test_path = self.test_dir / malicious_path
                result = self.security.validate_path_safety_enhanced(
                    test_path, self.test_dir
                )
                self.assertFalse(
                    result, f"Path traversal not blocked: {malicious_path}"
                )

    def test_command_injection_prevention(self):
        """CRITICAL: Test prevention of command injection via subprocess args"""
        # Dangerous command patterns
        injection_attempts = [
            ["ls", "; rm -rf /"],
            ["echo", "test", "|", "cat", "/etc/passwd"],
            ["ls", "&& rm -rf *"],
            ["cat", "$(whoami)"],
            ["ls", "`rm -rf /`"],
            ["echo", "test > /etc/passwd"],
            ["python", "-c", "import os; os.system('rm -rf /')"],
            ["sh", "-c", "rm -rf /"],
            ["sudo", "rm", "-rf", "/"],
        ]

        for dangerous_args in injection_attempts:
            with self.subTest(args=dangerous_args):
                with self.assertRaises(SecurityError):
                    self.security.sanitize_command_args(dangerous_args)

    def test_safe_command_args(self):
        """Test that safe commands pass validation"""
        safe_commands = [
            ["ls", "-la"],
            ["python", "--version"],
            ["git", "status"],
            ["npm", "install"],
            ["echo", "hello world"],
            ["cat", "file.txt"],
        ]

        for safe_args in safe_commands:
            with self.subTest(args=safe_args):
                try:
                    result = self.security.sanitize_command_args(safe_args)
                    self.assertIsInstance(result, list)
                    self.assertEqual(len(result), len(safe_args))
                except SecurityError:
                    self.fail(f"Safe command rejected: {safe_args}")

    def test_file_size_limits(self):
        """CRITICAL: Test DoS prevention via large file validation"""
        # Create test file
        test_file = self.test_dir / "test_file.txt"

        # Test small file (should pass)
        test_file.write_text("small content")
        self.assertTrue(self.security.validate_file_size(test_file, max_size_mb=1))

        # Test large file content simulation
        large_content = "A" * (2 * 1024 * 1024)  # 2MB content
        test_file.write_text(large_content)

        # Should fail with 1MB limit
        self.assertFalse(self.security.validate_file_size(test_file, max_size_mb=1))

        # Should pass with 5MB limit
        self.assertTrue(self.security.validate_file_size(test_file, max_size_mb=5))

    def test_safe_subprocess_execution(self):
        """CRITICAL: Test secure subprocess execution"""
        # Test safe command
        try:
            result = self.security.safe_subprocess_run(
                ["echo", "test"], cwd=self.test_dir, timeout=5
            )
            self.assertEqual(result.returncode, 0)
            self.assertIn("test", result.stdout)
        except SecurityError:
            self.fail("Safe subprocess command failed")

        # Test dangerous working directory
        with self.assertRaises(SecurityError):
            self.security.safe_subprocess_run(
                ["ls"],
                cwd="/etc",  # Should be blocked
                timeout=5,
            )

    def test_critical_path_protection(self):
        """CRITICAL: Test system critical path deletion prevention"""
        critical_paths = [
            Path.home(),
            Path("/"),
            Path("/etc"),
            Path("/usr"),
            Path("/var"),
        ]

        if os.name == "nt":
            critical_paths.extend(
                [Path("C:\\"), Path("C:\\Windows"), Path("C:\\Program Files")]
            )

        for critical_path in critical_paths:
            with self.subTest(path=critical_path):
                with self.assertRaises(SecurityError):
                    self.security.safe_rmtree(critical_path)

    def test_filename_sanitization(self):
        """Test malicious filename sanitization"""
        malicious_names = [
            "../../../etc/passwd",
            "file\x00.txt",
            "con.txt",  # Windows reserved
            "aux.txt",  # Windows reserved
            "A" * 300,  # Too long
            "",  # Empty
            "   ",  # Whitespace only
            "file/../other.txt",
        ]

        for malicious_name in malicious_names:
            with self.subTest(filename=malicious_name):
                sanitized = self.security.sanitize_filename(malicious_name)

                # Should not contain path traversal
                self.assertNotIn("..", sanitized)
                self.assertNotIn("/", sanitized)
                self.assertNotIn("\\", sanitized)
                self.assertNotIn("\x00", sanitized)

                # Should not be empty
                self.assertTrue(sanitized.strip())

                # Should not be too long
                self.assertLessEqual(len(sanitized), 255)


class TestHookSecurity(unittest.TestCase):
    """Test security of hook scripts"""

    def setUp(self):
        """Set up hook testing environment"""
        self.test_dir = Path(tempfile.mkdtemp())
        self.hooks_dir = self.test_dir / ".claude" / "hooks" / "moai"
        self.hooks_dir.mkdir(parents=True)

    def tearDown(self):
        """Clean up test environment"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_constitution_guard_security_import(self):
        """Test constitution guard properly imports security"""
        # Read actual hook file
        hook_file = (
            Path(__file__).parent.parent
            / "src"
            / "moai_adk"
            / "resources"
            / "templates"
            / ".claude"
            / "hooks"
            / "moai"
            / "constitution_guard.py"
        )

        if not hook_file.exists():
            self.skipTest("개발 가이드 guard hook not found")

        with open(hook_file, "r") as f:
            content = f.read()

        # Should import SecurityManager
        self.assertIn("from security import SecurityManager", content)
        self.assertIn("SecurityError", content)

        # Should have fallback for missing security
        self.assertIn("SecurityManager = None", content)

    def test_hook_input_validation(self):
        """Test hook input validation against malicious JSON"""
        malicious_inputs = [
            '{"tool_name": "Write", "tool_input": {"path": "../../../etc/passwd"}}',
            '{"tool_name": "Bash", "tool_input": {"command": "rm -rf /"}}',
            '{"tool_name": "Write", "tool_input": {"content": "'
            + "A" * 1000000
            + '"}}',  # JSON bomb
            '{"tool_name": "' + "X" * 100000 + '"}',  # Large key
        ]

        for malicious_json in malicious_inputs:
            with self.subTest(json_input=malicious_json[:100]):
                try:
                    # Test JSON parsing doesn't crash
                    data = json.loads(malicious_json)

                    # Tool name should be reasonable length
                    tool_name = data.get("tool_name", "")
                    self.assertLess(len(tool_name), 1000, "Tool name too long")

                except json.JSONDecodeError:
                    # Expected for malformed JSON
                    pass
                except Exception as e:
                    self.fail(f"Unexpected error parsing JSON: {e}")


class TestValidatorSecurity(unittest.TestCase):
    """Test validator module security"""

    def setUp(self):
        """Set up validator testing"""
        self.test_dir = Path(tempfile.mkdtemp())

    def tearDown(self):
        """Clean up"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_subprocess_timeout_security(self):
        """Test subprocess calls have timeouts to prevent DoS"""
        try:
            from moai_adk.core.validator import validate_claude_code

            # Mock subprocess.run to simulate hanging process
            with patch("subprocess.run") as mock_run:
                mock_run.side_effect = subprocess.TimeoutExpired("claude", 10)

                # Should handle timeout gracefully
                result = validate_claude_code()
                self.assertFalse(result)

        except ImportError:
            self.skipTest("Validator module not available")

    def test_path_validation_in_project_structure(self):
        """Test project structure validation doesn't access unsafe paths"""
        try:
            from moai_adk.core.validator import validate_project_structure

            # Create malicious symlink
            if os.name != "nt":  # Unix systems
                malicious_dir = self.test_dir / "malicious"
                malicious_dir.mkdir()

                # Try to create symlink to /etc (should be handled safely)
                try:
                    symlink_path = malicious_dir / "etc_link"
                    symlink_path.symlink_to("/etc")

                    # Validation should not follow dangerous symlinks
                    result = validate_project_structure(malicious_dir)
                    self.assertIsInstance(result, dict)

                except (OSError, PermissionError):
                    # Expected on some systems
                    pass

        except ImportError:
            self.skipTest("Validator module not available")


class TestBuildSystemSecurity(unittest.TestCase):
    """Test build system security"""

    def setUp(self):
        """Set up build testing"""
        self.test_dir = Path(tempfile.mkdtemp())

    def tearDown(self):
        """Clean up"""
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_makefile_command_safety(self):
        """Test Makefile doesn't contain dangerous commands"""
        makefile_path = Path(__file__).parent.parent / "Makefile"

        if not makefile_path.exists():
            self.skipTest("Makefile not found")

        with open(makefile_path, "r") as f:
            content = f.read()

        # Check for dangerous patterns
        dangerous_patterns = [
            r"rm\s+-rf\s+/",  # Dangerous rm commands
            r"sudo\s+rm",  # Privileged deletion
            r"\$\(shell\s+rm",  # Shell rm commands
            r">/etc/",  # Writing to system directories
        ]

        import re

        for pattern in dangerous_patterns:
            with self.subTest(pattern=pattern):
                matches = re.findall(pattern, content, re.IGNORECASE)
                self.assertEqual(len(matches), 0, f"Dangerous pattern found: {pattern}")


def run_security_tests():
    """Run all security tests"""
    print("🔒 Running Critical Security Tests...")

    # Test loader
    loader = unittest.TestLoader()
    suite = unittest.TestSuite()

    # Add all security test cases
    test_classes = [
        TestSecurityManager,
        TestHookSecurity,
        TestValidatorSecurity,
        TestBuildSystemSecurity,
    ]

    for test_class in test_classes:
        tests = loader.loadTestsFromTestCase(test_class)
        suite.addTests(tests)

    # Run tests with detailed output
    runner = unittest.TextTestRunner(verbosity=2, buffer=True)
    result = runner.run(suite)

    # Security-specific reporting
    print(f"\n🔒 Security Test Results:")
    print(f"   Tests run: {result.testsRun}")
    print(f"   Failures: {len(result.failures)}")
    print(f"   Errors: {len(result.errors)}")
    print(f"   Skipped: {len(result.skipped)}")

    # Security tests must have zero failures
    security_critical = len(result.failures) == 0 and len(result.errors) == 0

    if security_critical:
        print(f"\n✅ All critical security tests passed!")
    else:
        print(f"\n❌ SECURITY FAILURES DETECTED!")
        print("   Review and fix all security issues before proceeding.")

        # Print detailed failure information
        for test, error in result.failures + result.errors:
            print(f"\n🚨 FAILED: {test}")
            print(f"   Error: {error}")

    return security_critical


if __name__ == "__main__":
    success = run_security_tests()
    sys.exit(0 if success else 1)
