#!/usr/bin/env python3
"""
MoAI-ADK Test Migration Automation Script
Automated test file migration for package restructuring
"""

import re
import ast
import sys
from pathlib import Path
from typing import Dict, List, Set, Tuple
from dataclasses import dataclass


@dataclass
class ImportMigration:
    """Represents a single import migration rule."""

    old_import: str
    new_import: str
    affected_files: List[str]
    complexity_score: int  # 1-5, where 5 is most complex


class TestMigrationAnalyzer:
    """Analyzes and automates test file migration for package restructuring."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.tests_dir = project_root / "tests"
        self.src_dir = project_root / "src" / "moai_adk"

        # Migration mapping based on proposed restructuring
        self.migration_rules = {
            "moai_adk.config_manager": "moai_adk.config.config_manager",
            "moai_adk.security": "moai_adk.core.security.security",
            "moai_adk.config": "moai_adk.config.config",
            "moai_adk.installer": "moai_adk.install.installer",
            "moai_adk.git_manager": "moai_adk.core.managers.git_manager",
            "moai_adk.system_manager": "moai_adk.core.managers.system_manager",
            "moai_adk.directory_manager": "moai_adk.core.managers.directory_manager",
            "moai_adk.file_manager": "moai_adk.core.managers.file_manager",
            "moai_adk.progress_tracker": "moai_adk.utils.progress_tracker",
            "moai_adk.installation_result": "moai_adk.utils.installation_result",
            "moai_adk.validator": "moai_adk.core.validation.validator",
            "moai_adk.template_engine": "moai_adk.core.template_engine",
            "moai_adk.version_sync": "moai_adk.core.version_sync",
        }

    def analyze_test_files(self) -> Dict[str, List[ImportMigration]]:
        """Analyze all test files and identify required migrations."""
        migrations = {}

        for test_file in self.tests_dir.rglob("*.py"):
            if test_file.name.startswith("test_"):
                file_migrations = self._analyze_single_test_file(test_file)
                if file_migrations:
                    migrations[str(test_file.relative_to(self.project_root))] = (
                        file_migrations
                    )

        return migrations

    def _analyze_single_test_file(self, test_file: Path) -> List[ImportMigration]:
        """Analyze a single test file for required import migrations."""
        migrations = []

        try:
            with open(test_file, "r") as f:
                content = f.read()

            tree = ast.parse(content)

            for node in ast.walk(tree):
                if isinstance(node, (ast.Import, ast.ImportFrom)):
                    migration = self._check_import_node(node, test_file)
                    if migration:
                        migrations.append(migration)

        except (SyntaxError, UnicodeDecodeError) as e:
            print(f"Warning: Could not parse {test_file}: {e}")

        return migrations

    def _check_import_node(self, node: ast.AST, test_file: Path) -> ImportMigration:
        """Check if an import node requires migration."""
        if isinstance(node, ast.ImportFrom) and node.module:
            module = node.module
            if module in self.migration_rules:
                complexity = self._calculate_complexity(node, test_file)
                return ImportMigration(
                    old_import=f"from {module} import {', '.join(alias.name for alias in node.names)}",
                    new_import=f"from {self.migration_rules[module]} import {', '.join(alias.name for alias in node.names)}",
                    affected_files=[str(test_file.relative_to(self.project_root))],
                    complexity_score=complexity,
                )

        elif isinstance(node, ast.Import):
            for alias in node.names:
                if alias.name in self.migration_rules:
                    complexity = self._calculate_complexity(node, test_file)
                    return ImportMigration(
                        old_import=f"import {alias.name}",
                        new_import=f"import {self.migration_rules[alias.name]}",
                        affected_files=[str(test_file.relative_to(self.project_root))],
                        complexity_score=complexity,
                    )

        return None

    def _calculate_complexity(self, node: ast.AST, test_file: Path) -> int:
        """Calculate migration complexity score (1-5)."""
        base_score = 1

        # Factor in file size
        file_size = test_file.stat().st_size
        if file_size > 50000:  # >50KB
            base_score += 2
        elif file_size > 20000:  # >20KB
            base_score += 1

        # Factor in number of imports from same module
        if isinstance(node, ast.ImportFrom) and node.names:
            if len(node.names) > 3:
                base_score += 1
            if any(
                alias.name in ["SecurityManager", "ConfigManager"]
                for alias in node.names
            ):
                base_score += 1  # These are complex interfaces

        return min(base_score, 5)

    def generate_migration_script(
        self, migrations: Dict[str, List[ImportMigration]]
    ) -> str:
        """Generate automated migration script."""
        script = '''#!/usr/bin/env python3
"""
Automated Test Migration Script for MoAI-ADK Package Restructuring
Generated by TestMigrationAnalyzer
"""

import re
from pathlib import Path
from typing import Dict, List

class TestFileMigrator:
    """Automates the migration of test file imports."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.migration_map = {
'''

        # Add migration mappings
        for old, new in self.migration_rules.items():
            script += f'            "{old}": "{new}",\n'

        script += '''        }

    def migrate_file(self, file_path: Path) -> bool:
        """Migrate imports in a single file."""
        try:
            with open(file_path, 'r') as f:
                content = f.read()

            original_content = content

            # Apply migration rules
            for old_module, new_module in self.migration_map.items():
                # Handle 'from X import Y' pattern
                pattern1 = rf"from {re.escape(old_module)} import"
                replacement1 = f"from {new_module} import"
                content = re.sub(pattern1, replacement1, content)

                # Handle 'import X' pattern
                pattern2 = rf"import {re.escape(old_module)}(?=\\s|$)"
                replacement2 = f"import {new_module}"
                content = re.sub(pattern2, replacement2, content)

                # Handle 'import X as Y' pattern
                pattern3 = rf"import {re.escape(old_module)} as"
                replacement3 = f"import {new_module} as"
                content = re.sub(pattern3, replacement3, content)

            # Write back if changed
            if content != original_content:
                with open(file_path, 'w') as f:
                    f.write(content)
                print(f"✅ Migrated: {file_path}")
                return True
            else:
                print(f"⏭️  No changes: {file_path}")
                return False

        except Exception as e:
            print(f"❌ Error migrating {file_path}: {e}")
            return False

    def migrate_all_tests(self) -> Dict[str, bool]:
        """Migrate all test files."""
        results = {}
        tests_dir = self.project_root / "tests"

        for test_file in tests_dir.rglob("*.py"):
            if test_file.name.startswith("test_"):
                results[str(test_file)] = self.migrate_file(test_file)

        return results

    def validate_imports(self) -> List[str]:
        """Validate that all imports can be resolved after migration."""
        errors = []
        tests_dir = self.project_root / "tests"

        for test_file in tests_dir.rglob("*.py"):
            if test_file.name.startswith("test_"):
                try:
                    # Try to compile the file to check for import errors
                    with open(test_file, 'r') as f:
                        content = f.read()

                    compile(content, test_file, 'exec')

                except SyntaxError as e:
                    errors.append(f"{test_file}: {e}")
                except Exception as e:
                    errors.append(f"{test_file}: {e}")

        return errors

if __name__ == "__main__":
    import sys

    if len(sys.argv) != 2:
        print("Usage: python test_migration.py <project_root>")
        sys.exit(1)

    project_root = Path(sys.argv[1])
    migrator = TestFileMigrator(project_root)

    print("🚀 Starting test file migration...")
    results = migrator.migrate_all_tests()

    # Summary
    successful = sum(1 for r in results.values() if r)
    total = len(results)
    print(f"\n📊 Migration Summary:")
    print(f"   Files processed: {total}")
    print(f"   Files migrated: {successful}")
    print(f"   Files unchanged: {total - successful}")

    # Validation
    print("\n🔍 Validating imports...")
    errors = migrator.validate_imports()

    if errors:
        print("❌ Import validation errors:")
        for error in errors:
            print(f"   {error}")
        sys.exit(1)
    else:
        print("✅ All imports validated successfully!")
'''

        return script

    def generate_test_structure_setup(self) -> str:
        """Generate script to create new test directory structure."""
        return '''#!/usr/bin/env python3
"""
Test Structure Setup Script
Creates new directory structure for reorganized packages
"""

from pathlib import Path

def setup_test_structure(project_root: Path):
    """Set up new test directory structure."""
    tests_dir = project_root / "tests"

    # New directory structure
    new_dirs = [
        "unit/config",
        "unit/utils",
        "unit/core/managers",
        "unit/core/security",
        "unit/core/validation",
        "unit/install",
        "integration",
        "performance"
    ]

    for dir_path in new_dirs:
        full_path = tests_dir / dir_path
        full_path.mkdir(parents=True, exist_ok=True)

        # Create __init__.py files
        init_file = full_path / "__init__.py"
        if not init_file.exists():
            init_file.write_text('"""Test package."""\n')

        print(f"✅ Created: {full_path}")

    # Create centralized conftest.py
    conftest_content = """\"\"\"
Centralized test configuration and fixtures for MoAI-ADK.
\"\"\"

import pytest
import tempfile
from pathlib import Path
from unittest.mock import MagicMock

# Import the reorganized modules
from moai_adk.core.security.security import SecurityManager
from moai_adk.config.config_manager import ConfigManager
from moai_adk.utils.progress_tracker import ProgressTracker

@pytest.fixture
def temp_dir():
    \"\"\"Create a temporary directory for testing.\"\"\"
    with tempfile.TemporaryDirectory() as temp_dir:
        yield Path(temp_dir)

@pytest.fixture
def mock_security_manager():
    \"\"\"Centralized security manager mock for all tests.\"\"\"
    return MagicMock(spec=SecurityManager)

@pytest.fixture
def mock_config_manager():
    \"\"\"Centralized config manager mock for all tests.\"\"\"
    return MagicMock(spec=ConfigManager)

@pytest.fixture
def mock_progress_tracker():
    \"\"\"Centralized progress tracker mock for all tests.\"\"\"
    return MagicMock(spec=ProgressTracker)

@pytest.fixture
def isolated_test_env(temp_dir, monkeypatch):
    \"\"\"Isolated test environment with temporary directory.\"\"\"
    # Set up isolated environment
    monkeypatch.setenv("MOAI_ADK_TEST_MODE", "1")
    monkeypatch.setenv("MOAI_ADK_CONFIG_DIR", str(temp_dir))
    yield temp_dir
"""

    conftest_file = tests_dir / "conftest.py"
    conftest_file.write_text(conftest_content)
    print(f"✅ Created: {conftest_file}")

    # Create pytest.ini
    pytest_ini_content = """[tool:pytest]
minversion = 6.0
addopts =
    --maxfail=5
    --tb=short
    -ra
    --strict-markers
    --disable-warnings
    --cov=moai_adk
    --cov-report=term-missing
    --cov-report=html:coverage_html
    --junit-xml=test-reports/junit.xml

markers =
    slow: marks tests as slow (deselect with '-m "not slow"')
    config: marks tests as config-related
    core: marks tests as core-related
    utils: marks tests as utils-related
    install: marks tests as install-related
    integration: marks tests as integration tests
    performance: marks tests as performance tests

testpaths = tests
python_files = test_*.py
python_classes = Test*
python_functions = test_*

filterwarnings =
    ignore::DeprecationWarning
    ignore::PendingDeprecationWarning
"""

    pytest_ini_file = project_root / "pytest.ini"
    pytest_ini_file.write_text(pytest_ini_content)
    print(f"✅ Created: {pytest_ini_file}")

    print("\n🎉 Test structure setup complete!")

if __name__ == "__main__":
    import sys

    if len(sys.argv) != 2:
        print("Usage: python setup_test_structure.py <project_root>")
        sys.exit(1)

    project_root = Path(sys.argv[1])
    setup_test_structure(project_root)
'''


def main():
    """Main execution function."""
    if len(sys.argv) != 2:
        print("Usage: python test_migration_automation.py <project_root>")
        sys.exit(1)

    project_root = Path(sys.argv[1])
    analyzer = TestMigrationAnalyzer(project_root)

    print("🔍 Analyzing test files for migration requirements...")
    migrations = analyzer.analyze_test_files()

    # Generate detailed report
    print("\n📋 Migration Analysis Report:")
    print("=" * 60)

    total_files = len(migrations)
    total_migrations = sum(len(migs) for migs in migrations.values())
    high_complexity = sum(
        1 for migs in migrations.values() for mig in migs if mig.complexity_score >= 4
    )

    print(f"Files requiring migration: {total_files}")
    print(f"Total import migrations: {total_migrations}")
    print(f"High-complexity migrations: {high_complexity}")

    print("\nDetailed Migration Plan:")
    print("-" * 40)

    for file_path, file_migrations in migrations.items():
        print(f"\n📄 {file_path}")
        for mig in file_migrations:
            complexity_emoji = (
                "🔴"
                if mig.complexity_score >= 4
                else "🟡"
                if mig.complexity_score >= 2
                else "🟢"
            )
            print(f"   {complexity_emoji} {mig.old_import}")
            print(f"      → {mig.new_import}")

    # Generate migration script
    print("\n🛠️  Generating migration script...")
    migration_script = analyzer.generate_migration_script(migrations)

    script_path = project_root / "migrate_test_imports.py"
    with open(script_path, "w") as f:
        f.write(migration_script)

    print(f"✅ Migration script created: {script_path}")

    # Generate test structure setup script
    setup_script = analyzer.generate_test_structure_setup()
    setup_script_path = project_root / "setup_test_structure.py"
    with open(setup_script_path, "w") as f:
        f.write(setup_script)

    print(f"✅ Test structure setup script created: {setup_script_path}")

    print("\n🚀 Next Steps:")
    print("1. Run: python setup_test_structure.py <project_root>")
    print("2. Run: python migrate_test_imports.py <project_root>")
    print("3. Execute: make test to validate migration")
    print("4. Review and adjust any failing tests")


if __name__ == "__main__":
    main()
