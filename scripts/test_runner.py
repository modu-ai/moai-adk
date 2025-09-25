#!/usr/bin/env python3
"""
MoAI-ADK Test Runner
Cross-platform Python replacement for run-tests.sh
"""

import argparse
import json
import os
import re
import subprocess
import sys
from pathlib import Path
from typing import Dict, List, Optional, Tuple


class Colors:
    """ANSI color codes for terminal output"""
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'  # No Color

    @classmethod
    def disable_on_windows(cls):
        """Disable colors on Windows if colorama not available"""
        if os.name == 'nt':
            try:
                import colorama
                colorama.init()
            except ImportError:
                cls.RED = cls.GREEN = cls.YELLOW = cls.BLUE = cls.NC = ''


class TestRunner:
    def __init__(self, python_cmd: str = 'python3', verbose: bool = False,
                 coverage: bool = False, junit: bool = False):
        self.python_cmd = python_cmd
        self.verbose = verbose
        self.coverage = coverage
        self.junit = junit

        # Test statistics
        self.total_tests = 0
        self.total_failures = 0
        self.total_errors = 0
        self.total_skipped = 0
        self.all_success = True

        # Coverage settings
        self.coverage_available = False
        self.coverage_cmd = [python_cmd]

        # Project paths
        script_dir = Path(__file__).parent
        self.project_root = script_dir.parent

        # Initialize colors
        Colors.disable_on_windows()

    def print_header(self) -> None:
        """Print test runner header"""
        print(f"{Colors.BLUE}============================================{Colors.NC}")
        print(f"{Colors.BLUE}🧪 MoAI-ADK Test Suite Runner{Colors.NC}")
        print(f"{Colors.BLUE}============================================{Colors.NC}")

    def print_section(self, title: str) -> None:
        """Print section header"""
        print(f"\n{Colors.YELLOW}📋 {title}{Colors.NC}")
        print("----------------------------------------")

    def print_success(self, message: str) -> None:
        """Print success message"""
        print(f"{Colors.GREEN}✅ {message}{Colors.NC}")

    def print_error(self, message: str) -> None:
        """Print error message"""
        print(f"{Colors.RED}❌ {message}{Colors.NC}")

    def print_warning(self, message: str) -> None:
        """Print warning message"""
        print(f"{Colors.YELLOW}⚠️ {message}{Colors.NC}")

    def run_command(self, cmd: List[str], cwd: Optional[Path] = None) -> subprocess.CompletedProcess:
        """Run command and return result"""
        if cwd is None:
            cwd = self.project_root
        return subprocess.run(cmd, cwd=cwd, capture_output=True, text=True)

    def check_environment(self) -> bool:
        """Check Python environment and required modules"""
        self.print_section("Environment Check")

        # Check Python version
        result = self.run_command([self.python_cmd, '--version'])
        if result.returncode == 0:
            print(f"Python: {result.stdout.strip()}")
        else:
            self.print_error("Python not available")
            return False

        print(f"Project Root: {self.project_root}")

        # Check required modules
        print("Checking Python modules...")
        required_modules = ["unittest", "json", "pathlib", "hashlib", "tempfile"]

        for module in required_modules:
            result = self.run_command([self.python_cmd, '-c', f'import {module}'])
            if result.returncode == 0:
                self.print_success(f"{module} module available")
            else:
                self.print_error(f"{module} module not available")
                return False

        # Check optional modules
        optional_modules = ["coverage"]
        for module in optional_modules:
            result = self.run_command([self.python_cmd, '-c', f'import {module}'])
            if result.returncode == 0:
                self.print_success(f"{module} module available")
            else:
                self.print_warning(f"{module} module not available (optional)")

        return True

    def setup_coverage(self) -> None:
        """Setup coverage if requested and available"""
        if self.coverage:
            result = self.run_command([self.python_cmd, '-c', 'import coverage'])
            if result.returncode == 0:
                self.coverage_cmd = [
                    self.python_cmd, '-m', 'coverage', 'run',
                    '--source=.', '--omit=tests/*,*/__pycache__/*'
                ]
                self.coverage_available = True
                self.print_success("Coverage enabled")
            else:
                self.print_warning("Coverage requested but not available")
                self.coverage_available = False

    def parse_test_output(self, output: str, test_type: str) -> Tuple[int, int, int, int]:
        """Parse test output and extract statistics"""
        tests = failures = errors = skipped = 0

        # Look for test statistics patterns
        tests_match = re.search(r'Tests run: (\d+)', output)
        failures_match = re.search(r'Failures: (\d+)', output)
        errors_match = re.search(r'Errors: (\d+)', output)
        skipped_match = re.search(r'Skipped: (\d+)', output)

        if tests_match:
            tests = int(tests_match.group(1))
        if failures_match:
            failures = int(failures_match.group(1))
        if errors_match:
            errors = int(errors_match.group(1))
        if skipped_match:
            skipped = int(skipped_match.group(1))

        return tests, failures, errors, skipped

    def run_test_file(self, test_file: Path, test_name: str, success_pattern: str) -> bool:
        """Run a specific test file"""
        if not test_file.exists():
            self.print_warning(f"{test_name} not found ({test_file})")
            return True

        print(f"Running {test_name}...")

        cmd = self.coverage_cmd + [str(test_file)]
        result = self.run_command(cmd)

        if self.verbose:
            output = result.stdout + result.stderr
        else:
            # Filter output to show only important lines
            full_output = result.stdout + result.stderr
            important_lines = []
            for line in full_output.split('\n'):
                if any(pattern in line for pattern in
                      ['Tests run', 'Failures', 'Errors', 'Skipped', '✅', '❌']):
                    important_lines.append(line)
            output = '\n'.join(important_lines)

        if output.strip():
            print(output)

        # Check if tests passed
        success = success_pattern in (result.stdout + result.stderr)
        if success:
            self.print_success(f"{test_name} passed")
        else:
            self.print_error(f"{test_name} failed")
            self.all_success = False

        # Parse and update statistics
        tests, failures, errors, skipped = self.parse_test_output(
            result.stdout + result.stderr, test_name
        )
        self.total_tests += tests
        self.total_failures += failures
        self.total_errors += errors
        self.total_skipped += skipped

        return success

    def run_hook_tests(self) -> None:
        """Run hook system tests"""
        self.print_section("Hook System Tests")
        test_file = self.project_root / "tests" / "test_hooks.py"
        self.run_test_file(test_file, "Hook system tests", "All tests passed!")

    def run_build_tests(self) -> None:
        """Run build system tests"""
        self.print_section("Build System Tests")
        test_file = self.project_root / "tests" / "test_build.py"
        self.run_test_file(test_file, "Build system tests", "All build tests passed!")

    def validate_configurations(self) -> None:
        """Validate JSON configuration files"""
        self.print_section("Configuration Validation")
        print("Validating JSON configurations...")

        # Check Claude settings.json
        claude_settings = self.project_root / "src" / "templates" / ".claude" / "settings.json"
        if claude_settings.exists():
            try:
                with open(claude_settings, 'r', encoding='utf-8') as f:
                    json.load(f)
                self.print_success("Claude settings.json is valid")
            except json.JSONDecodeError:
                self.print_error("Claude settings.json is invalid")
                self.all_success = False
        else:
            self.print_warning("Claude settings.json not found")

        # Check MoAI config.json
        moai_config = self.project_root / "src" / "templates" / ".moai" / "config.json"
        if moai_config.exists():
            try:
                with open(moai_config, 'r', encoding='utf-8') as f:
                    json.load(f)
                self.print_success("MoAI config.json is valid")
            except json.JSONDecodeError:
                self.print_error("MoAI config.json is invalid")
                self.all_success = False
        else:
            self.print_warning("MoAI config.json not found")

    def test_build_system_integration(self) -> None:
        """Test build system integration"""
        self.print_section("Build System Integration")
        build_py = self.project_root / "build.py"
        if build_py.exists():
            print("Testing build system status...")
            result = self.run_command([self.python_cmd, str(build_py), 'status'])
            if result.returncode == 0:
                self.print_success("Build system status check passed")
            else:
                self.print_error("Build system status check failed")
                self.all_success = False
        else:
            self.print_error("build.py not found")
            self.all_success = False

    def check_hook_permissions(self) -> None:
        """Check hook script permissions"""
        self.print_section("Hook Scripts Permissions")
        hook_dir = self.project_root / "src" / "templates" / ".claude" / "hooks" / "moai"

        if hook_dir.exists():
            print("Checking hook script permissions...")
            for script in hook_dir.glob("*.py"):
                if os.access(script, os.X_OK):
                    self.print_success(f"{script.name} has execute permission")
                else:
                    self.print_warning(f"{script.name} missing execute permission")
        else:
            self.print_warning("Hook scripts directory not found")

    def generate_coverage_report(self) -> None:
        """Generate coverage report"""
        if self.coverage_available and self.coverage:
            self.print_section("Coverage Report")

            print("Generating coverage report...")

            # Combine coverage data
            self.run_command([self.python_cmd, '-m', 'coverage', 'combine'])

            # Generate text report
            result = self.run_command([self.python_cmd, '-m', 'coverage', 'report', '--show-missing'])
            if result.stdout:
                print(result.stdout)

            # Generate HTML report
            html_result = self.run_command([
                self.python_cmd, '-m', 'coverage', 'html', '-d', 'coverage_html'
            ])
            if html_result.returncode == 0:
                self.print_success("HTML coverage report generated: coverage_html/index.html")

    def generate_junit_report(self) -> None:
        """Generate JUnit XML report"""
        if self.junit:
            self.print_section("JUnit XML Report")

            # Check if xmlrunner is available
            result = self.run_command([self.python_cmd, '-c', 'import xmlrunner'])
            if result.returncode == 0:
                print("Generating JUnit XML report...")

                # Create test reports directory
                reports_dir = self.project_root / "test-reports"
                reports_dir.mkdir(exist_ok=True)

                # Run tests with XML output
                for test_name, pattern in [("hooks", "test_hooks.py"), ("build", "test_build.py")]:
                    test_file = self.project_root / "tests" / pattern
                    if test_file.exists():
                        test_output_dir = reports_dir / test_name
                        test_output_dir.mkdir(exist_ok=True)
                        self.run_command([
                            self.python_cmd, '-m', 'xmlrunner', 'discover',
                            '-s', 'tests', '-p', pattern, '-o', str(test_output_dir)
                        ])

                self.print_success("JUnit XML reports generated in test-reports/")
            else:
                self.print_warning("unittest-xml-reporting not available for JUnit reports")
                print("Install with: pip install unittest-xml-reporting")

    def print_summary(self) -> None:
        """Print final test summary"""
        self.print_section("Test Summary")
        print(f"Total Tests: {self.total_tests}")
        print(f"Failures: {self.total_failures}")
        print(f"Errors: {self.total_errors}")
        print(f"Skipped: {self.total_skipped}")

        if self.all_success:
            self.print_success("All tests passed! 🎉")
        else:
            self.print_error("Some tests failed! 💥")

    def run_all_tests(self) -> bool:
        """Run all test suites"""
        # Change to project root
        os.chdir(self.project_root)

        self.print_header()

        # Environment check
        if not self.check_environment():
            return False

        # Setup coverage
        self.setup_coverage()

        # Run test suites
        self.run_hook_tests()
        self.run_build_tests()
        self.validate_configurations()
        self.test_build_system_integration()
        self.check_hook_permissions()

        # Generate reports
        self.generate_coverage_report()
        self.generate_junit_report()

        # Print summary
        self.print_summary()

        return self.all_success


def main():
    """Main function"""
    parser = argparse.ArgumentParser(description="MoAI-ADK Test Suite Runner")
    parser.add_argument("--verbose", "-v", action="store_true", help="Verbose output")
    parser.add_argument("--coverage", "-c", action="store_true", help="Run with coverage report")
    parser.add_argument("--junit", "-j", action="store_true", help="Generate JUnit XML report")
    parser.add_argument("--python", default="python3", help="Python command to use (default: python3)")

    args = parser.parse_args()

    runner = TestRunner(
        python_cmd=args.python,
        verbose=args.verbose,
        coverage=args.coverage,
        junit=args.junit
    )

    success = runner.run_all_tests()
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()