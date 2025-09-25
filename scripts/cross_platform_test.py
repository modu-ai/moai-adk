#!/usr/bin/env python3
"""
MoAI-ADK Cross-Platform Compatibility Test Suite
Tests Windows/macOS/Linux compatibility for all MoAI-ADK components.

@TASK:CROSS-PLATFORM-001 Comprehensive platform compatibility verification
"""

import os
import sys
import platform
import subprocess
import json
from pathlib import Path
from typing import Dict, List, Tuple, Optional

class Colors:
    """ANSI color codes for terminal output"""
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    PURPLE = '\033[0;35m'
    CYAN = '\033[0;36m'
    NC = '\033[0m'  # No Color

    @classmethod
    def disable_on_windows(cls):
        """Disable colors on Windows if colorama not available"""
        if sys.platform == 'win32':
            try:
                import colorama
                colorama.init()
            except ImportError:
                cls.RED = cls.GREEN = cls.YELLOW = cls.BLUE = cls.NC = ''
                cls.PURPLE = cls.CYAN = ''

class CrossPlatformTester:
    def __init__(self):
        self.project_root = Path(__file__).parent.parent
        self.results = []
        Colors.disable_on_windows()

    def print_header(self, title: str) -> None:
        """Print test section header"""
        print(f"\n{Colors.CYAN}{'='*60}")
        print(f"🧪 {title}")
        print(f"{'='*60}{Colors.NC}")

    def print_success(self, message: str) -> None:
        """Print success message"""
        print(f"{Colors.GREEN}✅ {message}{Colors.NC}")

    def print_error(self, message: str) -> None:
        """Print error message"""
        print(f"{Colors.RED}❌ {message}{Colors.NC}")

    def print_warning(self, message: str) -> None:
        """Print warning message"""
        print(f"{Colors.YELLOW}⚠️ {message}{Colors.NC}")

    def print_info(self, message: str) -> None:
        """Print info message"""
        print(f"{Colors.BLUE}ℹ️ {message}{Colors.NC}")

    def run_command(self, cmd: List[str], description: str,
                   timeout: int = 60) -> Tuple[bool, str, str]:
        """Run command and return success, stdout, stderr"""
        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=timeout,
                cwd=self.project_root
            )
            return result.returncode == 0, result.stdout, result.stderr
        except subprocess.TimeoutExpired:
            return False, "", f"Command timed out after {timeout}s"
        except Exception as e:
            return False, "", str(e)

    def test_system_info(self) -> Dict:
        """Gather system information"""
        self.print_header("System Information")

        info = {
            'platform': platform.platform(),
            'system': platform.system(),
            'machine': platform.machine(),
            'python_version': platform.python_version(),
            'python_implementation': platform.python_implementation(),
            'architecture': platform.architecture()[0]
        }

        for key, value in info.items():
            self.print_info(f"{key.replace('_', ' ').title()}: {value}")

        return info

    def test_python_tools(self) -> bool:
        """Test all Python tools"""
        self.print_header("Python Tools Compatibility")

        tools = [
            (['python3', 'scripts/version_manager.py', 'status'], 'Version Manager'),
            (['python3', 'scripts/test_runner.py', '--help'], 'Test Runner'),
            (['python3', 'scripts/build.py', '--help'], 'Build System'),
        ]

        all_passed = True

        for cmd, name in tools:
            success, stdout, stderr = self.run_command(cmd, name)
            if success:
                self.print_success(f"{name}: Working correctly")
            else:
                self.print_error(f"{name}: Failed - {stderr}")
                all_passed = False

        return all_passed

    def test_cli_commands(self) -> bool:
        """Test CLI commands"""
        self.print_header("CLI Commands Test")

        commands = [
            (['moai', '--version'], 'Version Command'),
            (['moai', '--help'], 'Help Command'),
            (['moai', 'doctor'], 'Doctor Command'),
        ]

        all_passed = True

        for cmd, name in commands:
            success, stdout, stderr = self.run_command(cmd, name)
            if success:
                self.print_success(f"{name}: Working correctly")
            else:
                self.print_error(f"{name}: Failed - {stderr}")
                all_passed = False

        return all_passed

    def test_file_operations(self) -> bool:
        """Test cross-platform file operations"""
        self.print_header("File Operations Test")

        test_dir = self.project_root / "test_cross_platform"
        test_file = test_dir / "test_file.txt"

        try:
            # Create test directory
            test_dir.mkdir(exist_ok=True)
            self.print_success("Directory creation: Working")

            # Create test file
            test_file.write_text("Test content\n", encoding='utf-8')
            self.print_success("File creation: Working")

            # Read test file
            content = test_file.read_text(encoding='utf-8')
            if content == "Test content\n":
                self.print_success("File reading: Working")
            else:
                self.print_error("File reading: Content mismatch")
                return False

            # Test file permissions (Unix-like systems)
            if sys.platform != 'win32':
                test_file.chmod(0o755)
                self.print_success("File permissions: Working")
            else:
                self.print_info("File permissions: Skipped on Windows")

            # Cleanup
            test_file.unlink()
            test_dir.rmdir()
            self.print_success("File cleanup: Working")

            return True

        except Exception as e:
            self.print_error(f"File operations: Failed - {e}")
            return False

    def test_path_handling(self) -> bool:
        """Test cross-platform path handling"""
        self.print_header("Path Handling Test")

        try:
            # Test Path operations
            test_path = Path("test") / "subdir" / "file.txt"
            self.print_success(f"Path construction: {test_path}")

            # Test absolute vs relative
            abs_path = test_path.absolute()
            self.print_success(f"Absolute path: {abs_path}")

            # Test path components
            parts = test_path.parts
            self.print_success(f"Path parts: {parts}")

            # Test path separator handling
            path_str = str(test_path)
            if sys.platform == 'win32':
                expected_sep = '\\'
            else:
                expected_sep = '/'

            if expected_sep in path_str:
                self.print_success(f"Path separator ({expected_sep}): Correct")
            else:
                self.print_warning(f"Path separator: Unexpected format - {path_str}")

            return True

        except Exception as e:
            self.print_error(f"Path handling: Failed - {e}")
            return False

    def test_dependency_imports(self) -> bool:
        """Test all required dependencies can be imported"""
        self.print_header("Dependency Imports Test")

        dependencies = [
            'click',
            'colorama',
            'toml',
            'watchdog',
            'jsonschema',
            'git',  # gitpython
            'jinja2',
            'yaml',  # pyyaml
        ]

        all_passed = True

        for dep in dependencies:
            try:
                __import__(dep)
                self.print_success(f"{dep}: Imported successfully")
            except ImportError as e:
                self.print_error(f"{dep}: Import failed - {e}")
                all_passed = False

        return all_passed

    def test_encoding_support(self) -> bool:
        """Test Unicode/UTF-8 support"""
        self.print_header("Encoding Support Test")

        try:
            # Test Unicode strings
            unicode_text = "MoAI-ADK: 한글 테스트 🚀 © ® ™"
            self.print_success(f"Unicode display: {unicode_text}")

            # Test file encoding
            test_file = self.project_root / "test_encoding.txt"
            test_file.write_text(unicode_text, encoding='utf-8')
            read_text = test_file.read_text(encoding='utf-8')

            if read_text == unicode_text:
                self.print_success("UTF-8 file encoding: Working")
            else:
                self.print_error("UTF-8 file encoding: Failed")
                return False

            # Cleanup
            test_file.unlink()

            return True

        except Exception as e:
            self.print_error(f"Encoding support: Failed - {e}")
            return False

    def generate_report(self, test_results: Dict) -> str:
        """Generate cross-platform test report"""
        self.print_header("Test Report Generation")

        report = {
            'timestamp': str(Path(__file__).stat().st_mtime),
            'system_info': test_results['system_info'],
            'test_results': {
                'python_tools': test_results.get('python_tools', False),
                'cli_commands': test_results.get('cli_commands', False),
                'file_operations': test_results.get('file_operations', False),
                'path_handling': test_results.get('path_handling', False),
                'dependency_imports': test_results.get('dependency_imports', False),
                'encoding_support': test_results.get('encoding_support', False),
            },
            'overall_status': all(test_results.get(key, False) for key in [
                'python_tools', 'cli_commands', 'file_operations',
                'path_handling', 'dependency_imports', 'encoding_support'
            ])
        }

        report_file = self.project_root / f"cross_platform_report_{platform.system().lower()}.json"

        try:
            with open(report_file, 'w', encoding='utf-8') as f:
                json.dump(report, f, indent=2, ensure_ascii=False)

            self.print_success(f"Report saved: {report_file}")
            return str(report_file)

        except Exception as e:
            self.print_error(f"Report generation failed: {e}")
            return ""

    def run_all_tests(self) -> bool:
        """Run all cross-platform tests"""
        print(f"{Colors.PURPLE}")
        print("🗿 MoAI-ADK Cross-Platform Compatibility Test Suite")
        print("=" * 60)
        print(f"Testing on: {platform.system()} {platform.release()}")
        print(f"{Colors.NC}")

        test_results = {}

        # System information
        test_results['system_info'] = self.test_system_info()

        # Run all tests
        test_results['python_tools'] = self.test_python_tools()
        test_results['cli_commands'] = self.test_cli_commands()
        test_results['file_operations'] = self.test_file_operations()
        test_results['path_handling'] = self.test_path_handling()
        test_results['dependency_imports'] = self.test_dependency_imports()
        test_results['encoding_support'] = self.test_encoding_support()

        # Generate report
        report_file = self.generate_report(test_results)

        # Summary
        self.print_header("Test Summary")

        passed_tests = sum(1 for key, value in test_results.items()
                          if key != 'system_info' and value)
        total_tests = len(test_results) - 1  # Exclude system_info

        if passed_tests == total_tests:
            self.print_success(f"All tests passed! ({passed_tests}/{total_tests})")
            self.print_success(f"✅ {platform.system()} is fully supported!")
            return True
        else:
            self.print_error(f"Some tests failed ({passed_tests}/{total_tests})")
            self.print_warning(f"⚠️ {platform.system()} has compatibility issues")
            return False

def main():
    """Main function"""
    tester = CrossPlatformTester()
    success = tester.run_all_tests()

    return 0 if success else 1

if __name__ == "__main__":
    sys.exit(main())