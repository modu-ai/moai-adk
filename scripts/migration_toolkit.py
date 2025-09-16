#!/usr/bin/env python3
"""
MoAI-ADK Architecture Migration Toolkit
Automated implementation of comprehensive improvement plan

This script provides automated tools for implementing the architectural
improvements identified in the comprehensive analysis.
"""

import os
import shutil
import json
import ast
import time
import subprocess
import psutil
from pathlib import Path
from typing import Dict, List, Tuple, Optional, Set
from dataclasses import dataclass, asdict
from datetime import datetime
import re


@dataclass
class MigrationConfig:
    """Configuration for migration process"""
    source_root: Path
    backup_root: Path
    validation_enabled: bool = True
    rollback_enabled: bool = True
    performance_baseline: bool = True
    dry_run: bool = False


@dataclass
class MigrationStep:
    """Individual migration step"""
    name: str
    description: str
    action_type: str  # 'move', 'create', 'update', 'validate'
    source_path: Optional[str] = None
    target_path: Optional[str] = None
    dependencies: List[str] = None
    rollback_action: Optional[str] = None
    validation_check: Optional[str] = None


class PerformanceMonitor:
    """Monitor performance during migration"""

    def __init__(self):
        self.metrics = {}
        self.baselines = {}

    def capture_baseline(self, name: str):
        """Capture performance baseline"""
        process = psutil.Process()
        self.baselines[name] = {
            'timestamp': datetime.now(),
            'memory_mb': round(process.memory_info().rss / 1024 / 1024, 2),
            'cpu_percent': process.cpu_percent(),
        }

    def capture_metric(self, name: str, operation: str):
        """Capture performance metric during operation"""
        start_time = time.perf_counter()
        start_memory = psutil.Process().memory_info().rss / 1024 / 1024

        def end_capture():
            end_time = time.perf_counter()
            end_memory = psutil.Process().memory_info().rss / 1024 / 1024
            self.metrics[f"{name}_{operation}"] = {
                'duration_ms': round((end_time - start_time) * 1000, 2),
                'memory_delta_mb': round(end_memory - start_memory, 2),
                'timestamp': datetime.now()
            }

        return end_capture

    def generate_report(self) -> Dict:
        """Generate performance report"""
        return {
            'baselines': self.baselines,
            'metrics': self.metrics,
            'summary': {
                'total_operations': len(self.metrics),
                'avg_duration_ms': sum(m['duration_ms'] for m in self.metrics.values()) / len(self.metrics) if self.metrics else 0,
                'total_memory_delta_mb': sum(m['memory_delta_mb'] for m in self.metrics.values()),
            }
        }


class ImportAnalyzer:
    """Analyze and update import statements"""

    def __init__(self):
        self.import_map = {
            # Old -> New import mappings
            'moai_adk.config': 'moai_adk.config.settings',
            'moai_adk.logger': 'moai_adk.utils.logger',
            'moai_adk.progress_tracker': 'moai_adk.utils.progress',
            'moai_adk._version': 'moai_adk.utils.version',
            'moai_adk.core.config_manager': 'moai_adk.config.manager',
            'moai_adk.core.directory_manager': 'moai_adk.core.managers.directory',
            'moai_adk.core.file_manager': 'moai_adk.core.managers.file',
            'moai_adk.core.git_manager': 'moai_adk.core.managers.git',
            'moai_adk.core.system_manager': 'moai_adk.core.managers.system',
            'moai_adk.core.security': 'moai_adk.core.security.validator',
            'moai_adk.core.validator': 'moai_adk.core.validation.validator',
        }

    def analyze_file_imports(self, file_path: Path) -> Dict:
        """Analyze imports in a Python file"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()

            tree = ast.parse(content)
            imports = []
            relative_imports = []

            for node in ast.walk(tree):
                if isinstance(node, ast.Import):
                    for alias in node.names:
                        imports.append({
                            'type': 'import',
                            'module': alias.name,
                            'alias': alias.asname,
                            'line': node.lineno
                        })

                elif isinstance(node, ast.ImportFrom):
                    module_name = node.module or ''
                    if node.level > 0:  # Relative import
                        relative_imports.append({
                            'type': 'from_relative',
                            'module': module_name,
                            'level': node.level,
                            'names': [alias.name for alias in node.names],
                            'line': node.lineno
                        })
                    else:
                        imports.append({
                            'type': 'from_absolute',
                            'module': module_name,
                            'names': [alias.name for alias in node.names],
                            'line': node.lineno
                        })

            return {
                'file_path': str(file_path),
                'absolute_imports': imports,
                'relative_imports': relative_imports,
                'needs_update': any(imp['module'] in self.import_map for imp in imports)
            }

        except Exception as e:
            return {
                'file_path': str(file_path),
                'error': str(e),
                'absolute_imports': [],
                'relative_imports': [],
                'needs_update': False
            }

    def update_imports(self, file_path: Path, dry_run: bool = False) -> Dict:
        """Update imports in a Python file"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()

            lines = content.splitlines()
            updated_lines = []
            changes_made = []

            for line_num, line in enumerate(lines, 1):
                updated_line = line
                for old_import, new_import in self.import_map.items():
                    if old_import in line:
                        updated_line = line.replace(old_import, new_import)
                        if updated_line != line:
                            changes_made.append({
                                'line_number': line_num,
                                'old': line,
                                'new': updated_line
                            })

                updated_lines.append(updated_line)

            if changes_made and not dry_run:
                with open(file_path, 'w', encoding='utf-8') as f:
                    f.write('\n'.join(updated_lines))

            return {
                'file_path': str(file_path),
                'changes_made': changes_made,
                'success': True
            }

        except Exception as e:
            return {
                'file_path': str(file_path),
                'error': str(e),
                'changes_made': [],
                'success': False
            }


class ArchitectureMigrator:
    """Main migration orchestrator"""

    def __init__(self, config: MigrationConfig):
        self.config = config
        self.monitor = PerformanceMonitor()
        self.import_analyzer = ImportAnalyzer()
        self.migration_log = []
        self.rollback_actions = []

    def create_migration_plan(self) -> List[MigrationStep]:
        """Create detailed migration plan"""
        steps = [
            # Phase 1: Create new directory structure
            MigrationStep(
                name="create_config_dir",
                description="Create config/ directory",
                action_type="create",
                target_path="src/moai_adk/config",
                rollback_action="remove_directory"
            ),
            MigrationStep(
                name="create_utils_dir",
                description="Create utils/ directory",
                action_type="create",
                target_path="src/moai_adk/utils",
                rollback_action="remove_directory"
            ),
            MigrationStep(
                name="create_managers_dir",
                description="Create core/managers/ directory",
                action_type="create",
                target_path="src/moai_adk/core/managers",
                rollback_action="remove_directory"
            ),
            MigrationStep(
                name="create_security_dir",
                description="Create core/security/ directory",
                action_type="create",
                target_path="src/moai_adk/core/security",
                rollback_action="remove_directory"
            ),
            MigrationStep(
                name="create_validation_dir",
                description="Create core/validation/ directory",
                action_type="create",
                target_path="src/moai_adk/core/validation",
                rollback_action="remove_directory"
            ),

            # Phase 2: Move files to new locations
            MigrationStep(
                name="move_config",
                description="Move config.py to config/settings.py",
                action_type="move",
                source_path="src/moai_adk/config.py",
                target_path="src/moai_adk/config/settings.py",
                rollback_action="move_back"
            ),
            MigrationStep(
                name="move_logger",
                description="Move logger.py to utils/logger.py",
                action_type="move",
                source_path="src/moai_adk/logger.py",
                target_path="src/moai_adk/utils/logger.py",
                rollback_action="move_back"
            ),
            MigrationStep(
                name="move_progress_tracker",
                description="Move progress_tracker.py to utils/progress.py",
                action_type="move",
                source_path="src/moai_adk/progress_tracker.py",
                target_path="src/moai_adk/utils/progress.py",
                rollback_action="move_back"
            ),
            MigrationStep(
                name="move_version",
                description="Move _version.py to utils/version.py",
                action_type="move",
                source_path="src/moai_adk/_version.py",
                target_path="src/moai_adk/utils/version.py",
                rollback_action="move_back"
            ),
            MigrationStep(
                name="move_config_manager",
                description="Move core/config_manager.py to config/manager.py",
                action_type="move",
                source_path="src/moai_adk/core/config_manager.py",
                target_path="src/moai_adk/config/manager.py",
                rollback_action="move_back"
            ),
            MigrationStep(
                name="move_directory_manager",
                description="Move core/directory_manager.py to core/managers/directory.py",
                action_type="move",
                source_path="src/moai_adk/core/directory_manager.py",
                target_path="src/moai_adk/core/managers/directory.py",
                rollback_action="move_back"
            ),
            MigrationStep(
                name="move_file_manager",
                description="Move core/file_manager.py to core/managers/file.py",
                action_type="move",
                source_path="src/moai_adk/core/file_manager.py",
                target_path="src/moai_adk/core/managers/file.py",
                rollback_action="move_back"
            ),
            MigrationStep(
                name="move_git_manager",
                description="Move core/git_manager.py to core/managers/git.py",
                action_type="move",
                source_path="src/moai_adk/core/git_manager.py",
                target_path="src/moai_adk/core/managers/git.py",
                rollback_action="move_back"
            ),
            MigrationStep(
                name="move_system_manager",
                description="Move core/system_manager.py to core/managers/system.py",
                action_type="move",
                source_path="src/moai_adk/core/system_manager.py",
                target_path="src/moai_adk/core/managers/system.py",
                rollback_action="move_back"
            ),
            MigrationStep(
                name="move_security",
                description="Move core/security.py to core/security/validator.py",
                action_type="move",
                source_path="src/moai_adk/core/security.py",
                target_path="src/moai_adk/core/security/validator.py",
                rollback_action="move_back"
            ),
            MigrationStep(
                name="move_validator",
                description="Move core/validator.py to core/validation/validator.py",
                action_type="move",
                source_path="src/moai_adk/core/validator.py",
                target_path="src/moai_adk/core/validation/validator.py",
                rollback_action="move_back"
            ),

            # Phase 3: Create __init__.py files
            MigrationStep(
                name="create_config_init",
                description="Create config/__init__.py",
                action_type="create",
                target_path="src/moai_adk/config/__init__.py",
                rollback_action="remove_file"
            ),
            MigrationStep(
                name="create_utils_init",
                description="Create utils/__init__.py",
                action_type="create",
                target_path="src/moai_adk/utils/__init__.py",
                rollback_action="remove_file"
            ),
            MigrationStep(
                name="create_managers_init",
                description="Create core/managers/__init__.py",
                action_type="create",
                target_path="src/moai_adk/core/managers/__init__.py",
                rollback_action="remove_file"
            ),
            MigrationStep(
                name="create_security_init",
                description="Create core/security/__init__.py",
                action_type="create",
                target_path="src/moai_adk/core/security/__init__.py",
                rollback_action="remove_file"
            ),
            MigrationStep(
                name="create_validation_init",
                description="Create core/validation/__init__.py",
                action_type="create",
                target_path="src/moai_adk/core/validation/__init__.py",
                rollback_action="remove_file"
            ),

            # Phase 4: Update imports
            MigrationStep(
                name="update_imports",
                description="Update all import statements",
                action_type="update",
                validation_check="validate_imports"
            ),

            # Phase 5: Validation
            MigrationStep(
                name="validate_structure",
                description="Validate new directory structure",
                action_type="validate",
                validation_check="validate_architecture"
            ),
            MigrationStep(
                name="run_tests",
                description="Run test suite",
                action_type="validate",
                validation_check="run_test_suite"
            ),
        ]

        return steps

    def execute_step(self, step: MigrationStep) -> bool:
        """Execute a single migration step"""
        try:
            self.log_action(f"Executing step: {step.name} - {step.description}")

            end_capture = self.monitor.capture_metric("migration", step.name)

            if step.action_type == "create":
                if step.target_path:
                    target = self.config.source_root / step.target_path
                    if not self.config.dry_run:
                        target.mkdir(parents=True, exist_ok=True)
                    self.rollback_actions.append(("remove_directory", str(target)))

            elif step.action_type == "move":
                if step.source_path and step.target_path:
                    source = self.config.source_root / step.source_path
                    target = self.config.source_root / step.target_path

                    if source.exists():
                        if not self.config.dry_run:
                            target.parent.mkdir(parents=True, exist_ok=True)
                            shutil.move(str(source), str(target))
                        self.rollback_actions.append(("move_back", str(target), str(source)))

            elif step.action_type == "update":
                if step.name == "update_imports":
                    self.update_all_imports()

            elif step.action_type == "validate":
                if step.validation_check:
                    if not self.run_validation(step.validation_check):
                        end_capture()
                        return False

            end_capture()
            self.log_action(f"Successfully completed step: {step.name}")
            return True

        except Exception as e:
            self.log_action(f"Error in step {step.name}: {str(e)}")
            return False

    def update_all_imports(self):
        """Update imports in all Python files"""
        python_files = list(self.config.source_root.rglob("*.py"))

        for py_file in python_files:
            if "test" in str(py_file) or "__pycache__" in str(py_file):
                continue

            result = self.import_analyzer.update_imports(py_file, self.config.dry_run)
            if result['changes_made']:
                self.log_action(f"Updated imports in {py_file}: {len(result['changes_made'])} changes")

    def run_validation(self, validation_type: str) -> bool:
        """Run validation checks"""
        try:
            if validation_type == "validate_imports":
                return self.validate_imports()
            elif validation_type == "validate_architecture":
                return self.validate_architecture()
            elif validation_type == "run_test_suite":
                return self.run_test_suite()
            return True

        except Exception as e:
            self.log_action(f"Validation error ({validation_type}): {str(e)}")
            return False

    def validate_imports(self) -> bool:
        """Validate that all imports work"""
        test_imports = [
            "moai_adk.config.settings",
            "moai_adk.utils.logger",
            "moai_adk.core.managers.directory",
            "moai_adk.core.security.validator"
        ]

        for module in test_imports:
            try:
                subprocess.run(
                    ["python", "-c", f"import {module}"],
                    check=True,
                    capture_output=True,
                    cwd=self.config.source_root.parent
                )
            except subprocess.CalledProcessError:
                self.log_action(f"Import validation failed for {module}")
                return False

        return True

    def validate_architecture(self) -> bool:
        """Validate new architecture structure"""
        required_dirs = [
            "src/moai_adk/config",
            "src/moai_adk/utils",
            "src/moai_adk/core/managers",
            "src/moai_adk/core/security",
            "src/moai_adk/core/validation"
        ]

        for dir_path in required_dirs:
            if not (self.config.source_root / dir_path).exists():
                self.log_action(f"Architecture validation failed: missing {dir_path}")
                return False

        return True

    def run_test_suite(self) -> bool:
        """Run the test suite to ensure everything works"""
        try:
            result = subprocess.run(
                ["python", "-m", "pytest", "tests/", "-v"],
                capture_output=True,
                text=True,
                cwd=self.config.source_root.parent
            )

            if result.returncode == 0:
                self.log_action("All tests passed")
                return True
            else:
                self.log_action(f"Tests failed: {result.stdout}\n{result.stderr}")
                return False

        except Exception as e:
            self.log_action(f"Test execution error: {str(e)}")
            return False

    def create_backup(self):
        """Create full backup before migration"""
        if self.config.rollback_enabled and not self.config.dry_run:
            self.log_action("Creating backup...")
            shutil.copytree(
                self.config.source_root,
                self.config.backup_root,
                ignore=shutil.ignore_patterns('__pycache__', '*.pyc', '.git')
            )
            self.log_action(f"Backup created at {self.config.backup_root}")

    def rollback(self):
        """Rollback all changes"""
        self.log_action("Starting rollback...")

        for action_data in reversed(self.rollback_actions):
            try:
                action_type = action_data[0]
                if action_type == "remove_directory":
                    path = Path(action_data[1])
                    if path.exists():
                        shutil.rmtree(path)

                elif action_type == "move_back":
                    source = Path(action_data[1])
                    target = Path(action_data[2])
                    if source.exists():
                        shutil.move(str(source), str(target))

                elif action_type == "remove_file":
                    path = Path(action_data[1])
                    if path.exists():
                        path.unlink()

            except Exception as e:
                self.log_action(f"Rollback error: {str(e)}")

        self.log_action("Rollback completed")

    def execute_migration(self) -> Dict:
        """Execute full migration process"""
        start_time = datetime.now()
        self.monitor.capture_baseline("migration_start")

        try:
            # Create backup
            if self.config.rollback_enabled:
                self.create_backup()

            # Create and execute migration plan
            steps = self.create_migration_plan()
            successful_steps = 0
            failed_steps = []

            for step in steps:
                if self.execute_step(step):
                    successful_steps += 1
                else:
                    failed_steps.append(step.name)
                    if not self.config.dry_run:
                        break  # Stop on first failure

            # Generate final report
            end_time = datetime.now()
            duration = end_time - start_time

            migration_report = {
                'migration_summary': {
                    'start_time': start_time.isoformat(),
                    'end_time': end_time.isoformat(),
                    'duration_seconds': duration.total_seconds(),
                    'total_steps': len(steps),
                    'successful_steps': successful_steps,
                    'failed_steps': failed_steps,
                    'success': len(failed_steps) == 0
                },
                'performance_report': self.monitor.generate_report(),
                'migration_log': self.migration_log,
                'rollback_available': self.config.rollback_enabled,
                'dry_run': self.config.dry_run
            }

            return migration_report

        except Exception as e:
            self.log_action(f"Critical migration error: {str(e)}")
            if self.config.rollback_enabled and not self.config.dry_run:
                self.rollback()

            return {
                'migration_summary': {
                    'success': False,
                    'error': str(e)
                },
                'migration_log': self.migration_log
            }

    def log_action(self, message: str):
        """Log migration action"""
        timestamp = datetime.now().isoformat()
        log_entry = f"[{timestamp}] {message}"
        self.migration_log.append(log_entry)
        print(log_entry)


def main():
    """Main execution function"""
    import argparse

    parser = argparse.ArgumentParser(description="MoAI-ADK Architecture Migration Toolkit")
    parser.add_argument("--source-root", type=str, default=".",
                      help="Source root directory")
    parser.add_argument("--backup-root", type=str, default="./backup_migration",
                      help="Backup directory")
    parser.add_argument("--dry-run", action="store_true",
                      help="Run in dry-run mode (no actual changes)")
    parser.add_argument("--no-rollback", action="store_true",
                      help="Disable rollback capability")
    parser.add_argument("--output", type=str, default="migration_report.json",
                      help="Output file for migration report")

    args = parser.parse_args()

    # Create migration configuration
    config = MigrationConfig(
        source_root=Path(args.source_root).resolve(),
        backup_root=Path(args.backup_root).resolve(),
        validation_enabled=True,
        rollback_enabled=not args.no_rollback,
        performance_baseline=True,
        dry_run=args.dry_run
    )

    # Execute migration
    migrator = ArchitectureMigrator(config)
    report = migrator.execute_migration()

    # Save report
    with open(args.output, 'w') as f:
        json.dump(report, f, indent=2, default=str)

    print(f"\nMigration report saved to: {args.output}")

    if report['migration_summary']['success']:
        print("✅ Migration completed successfully!")
    else:
        print("❌ Migration failed. Check the report for details.")

    return 0 if report['migration_summary']['success'] else 1


if __name__ == "__main__":
    exit(main())