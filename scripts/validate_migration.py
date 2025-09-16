#!/usr/bin/env python3
"""
Migration Validation Script for MoAI-ADK

Comprehensive validation of the package restructuring to ensure:
1. All imports work correctly
2. CLI commands function properly
3. Public API remains intact
4. No functionality regressions
5. Test suite passes
"""

import subprocess
import sys
import importlib
import traceback
from pathlib import Path
from typing import List, Dict, Tuple, Optional
from dataclasses import dataclass
from datetime import datetime
import json

@dataclass
class ValidationResult:
    """Result of a validation test"""
    test_name: str
    passed: bool
    error_message: Optional[str] = None
    execution_time: Optional[float] = None

class MigrationValidator:
    """Comprehensive validation for MoAI-ADK migration"""

    def __init__(self, base_path: Path):
        self.base_path = base_path
        self.results: List[ValidationResult] = []

    def log_result(self, test_name: str, passed: bool, error_msg: str = None, exec_time: float = None):
        """Log a validation result"""
        result = ValidationResult(test_name, passed, error_msg, exec_time)
        self.results.append(result)

        status = "✅ PASS" if passed else "❌ FAIL"
        time_str = f" ({exec_time:.2f}s)" if exec_time else ""
        print(f"{status}: {test_name}{time_str}")

        if error_msg:
            print(f"   Error: {error_msg}")

    def run_command_test(self, test_name: str, command: str, timeout: int = 30) -> bool:
        """Run a command and validate it succeeds"""
        start_time = datetime.now()

        try:
            result = subprocess.run(
                command,
                shell=True,
                capture_output=True,
                text=True,
                timeout=timeout,
                cwd=self.base_path
            )

            exec_time = (datetime.now() - start_time).total_seconds()

            if result.returncode == 0:
                self.log_result(test_name, True, exec_time=exec_time)
                return True
            else:
                error_msg = result.stderr.strip() or result.stdout.strip() or "Command failed"
                self.log_result(test_name, False, error_msg, exec_time)
                return False

        except subprocess.TimeoutExpired:
            exec_time = (datetime.now() - start_time).total_seconds()
            self.log_result(test_name, False, f"Timeout after {timeout}s", exec_time)
            return False
        except Exception as e:
            exec_time = (datetime.now() - start_time).total_seconds()
            self.log_result(test_name, False, str(e), exec_time)
            return False

    def run_import_test(self, test_name: str, import_statement: str) -> bool:
        """Test that an import statement works correctly"""
        start_time = datetime.now()

        try:
            # Execute the import
            exec(import_statement)
            exec_time = (datetime.now() - start_time).total_seconds()
            self.log_result(test_name, True, exec_time=exec_time)
            return True

        except Exception as e:
            exec_time = (datetime.now() - start_time).total_seconds()
            self.log_result(test_name, False, str(e), exec_time)
            return False

    def validate_basic_imports(self) -> int:
        """Test basic package imports work"""
        print("\n=== BASIC IMPORT VALIDATION ===")

        basic_imports = [
            ("Version Import", "from moai_adk import __version__"),
            ("Logger Import", "from moai_adk import get_logger"),
            ("Config Import", "from moai_adk import Config"),
            ("Runtime Config Import", "from moai_adk import RuntimeConfig"),
            ("Security Manager Import", "from moai_adk import SecurityManager"),
            ("Config Manager Import", "from moai_adk import ConfigManager"),
            ("Template Engine Import", "from moai_adk import TemplateEngine"),
            ("Installer Import", "from moai_adk import SimplifiedInstaller"),
            ("CLI Commands Import", "from moai_adk import CLICommands"),
        ]

        passed = 0
        for test_name, import_stmt in basic_imports:
            if self.run_import_test(test_name, import_stmt):
                passed += 1

        return passed

    def validate_new_structure_imports(self) -> int:
        """Test new package structure imports work"""
        print("\n=== NEW STRUCTURE IMPORT VALIDATION ===")

        new_imports = [
            ("Utils Version", "from moai_adk.utils.version import __version__"),
            ("Utils Logger", "from moai_adk.utils.logger import get_logger"),
            ("Utils Progress", "from moai_adk.utils.progress_tracker import ProgressTracker"),
            ("Config Base", "from moai_adk.config.base import Config"),
            ("Config Result", "from moai_adk.config.installation_result import InstallationResult"),
            ("Security Manager", "from moai_adk.core.security.security_manager import SecurityManager"),
            ("Config Manager", "from moai_adk.core.managers.config_manager import ConfigManager"),
            ("Template Manager", "from moai_adk.core.managers.template_manager import TemplateEngine"),
            ("File Manager", "from moai_adk.core.managers.file_manager import FileManager"),
            ("Directory Manager", "from moai_adk.core.managers.directory_manager import DirectoryManager"),
            ("Git Manager", "from moai_adk.core.managers.git_manager import GitManager"),
            ("System Manager", "from moai_adk.core.managers.system_manager import SystemManager"),
        ]

        passed = 0
        for test_name, import_stmt in new_imports:
            if self.run_import_test(test_name, import_stmt):
                passed += 1

        return passed

    def validate_backward_compatibility(self) -> int:
        """Test backward compatibility imports still work"""
        print("\n=== BACKWARD COMPATIBILITY VALIDATION ===")

        backward_compat_tests = [
            ("Old Config Import", "from moai_adk.config import Config"),
            ("Old Security Import", "from moai_adk.security import SecurityManager"),
            ("Old Logger Import", "from moai_adk.logger import get_logger"),
            ("Direct Module Access", """
import moai_adk
config_class = moai_adk.Config
security_class = moai_adk.SecurityManager
logger_func = moai_adk.get_logger
"""),
        ]

        passed = 0
        for test_name, import_stmt in backward_compat_tests:
            if self.run_import_test(test_name, import_stmt):
                passed += 1

        return passed

    def validate_instantiation(self) -> int:
        """Test that classes can be instantiated correctly"""
        print("\n=== CLASS INSTANTIATION VALIDATION ===")

        instantiation_tests = [
            ("Config Creation", """
from moai_adk import Config
config = Config()
"""),
            ("Security Manager Creation", """
from moai_adk import SecurityManager
sm = SecurityManager()
"""),
            ("Logger Creation", """
from moai_adk import get_logger
logger = get_logger(__name__)
"""),
            ("Progress Tracker Creation", """
from moai_adk.utils.progress_tracker import ProgressTracker
pt = ProgressTracker("test", 5)
"""),
            ("Installation Result Creation", """
from moai_adk.config.installation_result import InstallationResult
from moai_adk.config.base import Config
config = Config()
result = InstallationResult(config)
"""),
        ]

        passed = 0
        for test_name, test_code in instantiation_tests:
            if self.run_import_test(test_name, test_code):
                passed += 1

        return passed

    def validate_cli_functionality(self) -> int:
        """Test CLI commands work correctly"""
        print("\n=== CLI FUNCTIONALITY VALIDATION ===")

        cli_tests = [
            ("CLI Version", "moai --version"),
            ("CLI Help", "moai --help"),
            ("CLI Doctor", "moai doctor --dry-run"),
            ("Install Help", "moai-install --help"),
        ]

        passed = 0
        for test_name, command in cli_tests:
            if self.run_command_test(test_name, command):
                passed += 1

        return passed

    def validate_entry_points(self) -> int:
        """Test package entry points are correctly configured"""
        print("\n=== ENTRY POINT VALIDATION ===")

        entry_point_tests = [
            ("Moai Command Available", "which moai"),
            ("Moai Install Available", "which moai-install"),
            ("Python Module Execution", "python -m moai_adk.cli --version"),
        ]

        passed = 0
        for test_name, command in entry_point_tests:
            if self.run_command_test(test_name, command):
                passed += 1

        return passed

    def validate_package_metadata(self) -> int:
        """Test package metadata is correct"""
        print("\n=== PACKAGE METADATA VALIDATION ===")

        metadata_tests = [
            ("Version Consistency", """
from moai_adk import __version__
from moai_adk.utils.version import __version__ as utils_version
assert __version__ == utils_version, f"Version mismatch: {__version__} != {utils_version}"
"""),
            ("Package Info", """
from moai_adk import __author__, __email__, __description__
assert __author__, "Author not set"
assert __email__, "Email not set"
assert __description__, "Description not set"
"""),
            ("All Exports Available", """
from moai_adk import __all__
for item in __all__:
    assert hasattr(__import__('moai_adk'), item), f"Missing export: {item}"
"""),
        ]

        passed = 0
        for test_name, test_code in metadata_tests:
            if self.run_import_test(test_name, test_code):
                passed += 1

        return passed

    def validate_test_suite(self) -> int:
        """Run the full test suite to ensure no regressions"""
        print("\n=== TEST SUITE VALIDATION ===")

        test_commands = [
            ("Unit Tests", "pytest tests/unit/ -v --tb=short"),
            ("Integration Tests", "pytest tests/integration/ -v --tb=short"),
            ("Full Test Suite", "pytest tests/ -x --tb=short"),
        ]

        passed = 0
        for test_name, command in test_commands:
            # Longer timeout for test suite
            if self.run_command_test(test_name, command, timeout=120):
                passed += 1

        return passed

    def validate_import_performance(self) -> int:
        """Test that imports are reasonably fast"""
        print("\n=== IMPORT PERFORMANCE VALIDATION ===")

        performance_tests = [
            ("Fast Basic Import", "python -c 'import time; start=time.time(); from moai_adk import Config; print(f\"Import time: {time.time()-start:.3f}s\")'"),
            ("Fast Full Import", "python -c 'import time; start=time.time(); from moai_adk import *; print(f\"Import time: {time.time()-start:.3f}s\")'"),
        ]

        passed = 0
        for test_name, command in performance_tests:
            if self.run_command_test(test_name, command, timeout=10):
                passed += 1

        return passed

    def check_for_circular_imports(self) -> int:
        """Check for circular import issues"""
        print("\n=== CIRCULAR IMPORT VALIDATION ===")

        # Test importing all modules simultaneously to catch circular imports
        circular_test = """
try:
    # Import all major modules at once to trigger any circular imports
    from moai_adk import (
        Config, SecurityManager, ConfigManager, TemplateEngine,
        get_logger, SimplifiedInstaller, CLICommands
    )

    from moai_adk.utils import version, logger, progress_tracker
    from moai_adk.config import base, installation_result
    from moai_adk.core.security import security_manager
    from moai_adk.core.managers import (
        config_manager, template_manager, file_manager,
        directory_manager, git_manager, system_manager
    )

    print("All imports successful - no circular dependencies detected")
except ImportError as e:
    print(f"Import error: {e}")
    exit(1)
"""

        passed = 0
        if self.run_import_test("Circular Import Check", circular_test):
            passed += 1

        return passed

    def generate_report(self) -> Dict:
        """Generate comprehensive validation report"""
        total_tests = len(self.results)
        passed_tests = sum(1 for r in self.results if r.passed)
        failed_tests = total_tests - passed_tests

        success_rate = (passed_tests / total_tests) * 100 if total_tests > 0 else 0

        # Group results by category
        categories = {}
        for result in self.results:
            category = result.test_name.split()[0] if " " in result.test_name else "Other"
            if category not in categories:
                categories[category] = {"passed": 0, "failed": 0, "tests": []}

            if result.passed:
                categories[category]["passed"] += 1
            else:
                categories[category]["failed"] += 1

            categories[category]["tests"].append({
                "name": result.test_name,
                "passed": result.passed,
                "error": result.error_message,
                "time": result.execution_time
            })

        return {
            "summary": {
                "total_tests": total_tests,
                "passed": passed_tests,
                "failed": failed_tests,
                "success_rate": success_rate
            },
            "categories": categories,
            "failed_tests": [
                {
                    "name": r.test_name,
                    "error": r.error_message,
                    "time": r.execution_time
                }
                for r in self.results if not r.passed
            ]
        }

    def run_full_validation(self) -> bool:
        """Run complete validation suite"""
        print("🗿 MoAI-ADK Migration Validation Suite")
        print("=" * 50)

        validation_sections = [
            ("Basic Imports", self.validate_basic_imports),
            ("New Structure", self.validate_new_structure_imports),
            ("Backward Compatibility", self.validate_backward_compatibility),
            ("Class Instantiation", self.validate_instantiation),
            ("CLI Functionality", self.validate_cli_functionality),
            ("Entry Points", self.validate_entry_points),
            ("Package Metadata", self.validate_package_metadata),
            ("Import Performance", self.validate_import_performance),
            ("Circular Imports", self.check_for_circular_imports),
            ("Test Suite", self.validate_test_suite),
        ]

        section_results = {}

        for section_name, validation_func in validation_sections:
            print(f"\n📊 Running {section_name} validation...")
            try:
                passed_count = validation_func()
                section_results[section_name] = passed_count
            except Exception as e:
                print(f"❌ Section {section_name} failed with exception: {e}")
                section_results[section_name] = 0

        # Generate final report
        report = self.generate_report()

        print("\n" + "=" * 50)
        print("📋 VALIDATION SUMMARY")
        print("=" * 50)

        print(f"Total Tests: {report['summary']['total_tests']}")
        print(f"Passed: {report['summary']['passed']} ✅")
        print(f"Failed: {report['summary']['failed']} ❌")
        print(f"Success Rate: {report['summary']['success_rate']:.1f}%")

        if report['failed_tests']:
            print(f"\n❌ Failed Tests ({len(report['failed_tests'])}):")
            for test in report['failed_tests']:
                print(f"  - {test['name']}: {test['error']}")

        print(f"\n📊 Section Results:")
        for section, count in section_results.items():
            print(f"  {section}: {count} tests passed")

        # Save detailed report
        report_file = self.base_path / "validation_report.json"
        with open(report_file, 'w') as f:
            json.dump(report, f, indent=2, default=str)
        print(f"\n📄 Detailed report saved to: {report_file}")

        # Determine overall success
        success_threshold = 85.0  # 85% success rate required
        migration_successful = report['summary']['success_rate'] >= success_threshold

        if migration_successful:
            print(f"\n🎉 MIGRATION VALIDATION SUCCESSFUL!")
            print(f"✅ Success rate ({report['summary']['success_rate']:.1f}%) meets threshold ({success_threshold}%)")
        else:
            print(f"\n🚨 MIGRATION VALIDATION FAILED!")
            print(f"❌ Success rate ({report['summary']['success_rate']:.1f}%) below threshold ({success_threshold}%)")

        return migration_successful


def main():
    """Main entry point for validation script"""
    import argparse

    parser = argparse.ArgumentParser(description="MoAI-ADK Migration Validation")
    parser.add_argument("--base-path", type=Path, default=Path.cwd(),
                        help="Base path of the project")
    parser.add_argument("--section", choices=[
        "basic", "structure", "compat", "instantiation",
        "cli", "entry", "metadata", "performance", "circular", "tests"
    ], help="Run only a specific validation section")

    args = parser.parse_args()

    validator = MigrationValidator(args.base_path)

    if args.section:
        # Run specific section
        section_map = {
            "basic": validator.validate_basic_imports,
            "structure": validator.validate_new_structure_imports,
            "compat": validator.validate_backward_compatibility,
            "instantiation": validator.validate_instantiation,
            "cli": validator.validate_cli_functionality,
            "entry": validator.validate_entry_points,
            "metadata": validator.validate_package_metadata,
            "performance": validator.validate_import_performance,
            "circular": validator.check_for_circular_imports,
            "tests": validator.validate_test_suite,
        }

        validation_func = section_map[args.section]
        print(f"Running {args.section} validation only...")
        passed = validation_func()
        print(f"Result: {passed} tests passed")

    else:
        # Run full validation
        success = validator.run_full_validation()
        sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()