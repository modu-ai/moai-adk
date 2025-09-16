#!/usr/bin/env python3
"""
Migration Rollback Script for MoAI-ADK

Safe rollback mechanism to restore the package to pre-migration state
if issues are detected during or after the restructuring process.
"""

import subprocess
import shutil
import sys
from pathlib import Path
from typing import List, Optional, Dict
from datetime import datetime
import json


class MigrationRollback:
    """Handles safe rollback of MoAI-ADK migration"""

    def __init__(self, base_path: Path):
        self.base_path = base_path
        self.backup_path = base_path / "migration_backups"
        self.log_file = base_path / "rollback.log"

    def log(self, message: str, level: str = "INFO"):
        """Log rollback progress"""
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        log_entry = f"[{timestamp}] {level}: {message}\n"

        print(f"{level}: {message}")
        with open(self.log_file, "a") as f:
            f.write(log_entry)

    def get_available_backups(self) -> Dict[str, Dict]:
        """Get list of available backup points"""
        backups = {}

        # Git tag backups
        try:
            result = subprocess.run([
                "git", "tag", "-l", "migration-backup-*"
            ], capture_output=True, text=True, cwd=self.base_path)

            if result.returncode == 0:
                for tag in result.stdout.strip().split('\n'):
                    if tag.startswith('migration-backup-'):
                        backup_name = tag.replace('migration-backup-', '')
                        backups[backup_name] = {
                            "type": "git_tag",
                            "tag": tag,
                            "available": True
                        }

        except subprocess.CalledProcessError:
            self.log("Could not fetch git tags", "WARNING")

        # Filesystem backups
        if self.backup_path.exists():
            for backup_dir in self.backup_path.iterdir():
                if backup_dir.is_dir():
                    backups[backup_dir.name] = {
                        "type": "filesystem",
                        "path": backup_dir,
                        "available": True
                    }

        return backups

    def rollback_to_git_tag(self, tag: str) -> bool:
        """Rollback using git tag"""
        try:
            self.log(f"Rolling back to git tag: {tag}")

            # Stash any uncommitted changes
            subprocess.run([
                "git", "stash", "push", "-m", f"Pre-rollback stash {datetime.now()}"
            ], cwd=self.base_path)

            # Hard reset to the tag
            result = subprocess.run([
                "git", "reset", "--hard", tag
            ], check=True, capture_output=True, text=True, cwd=self.base_path)

            # Clean untracked files
            subprocess.run([
                "git", "clean", "-fd"
            ], check=True, cwd=self.base_path)

            # Reinstall package
            subprocess.run([
                "pip", "install", "-e", ".", "--force-reinstall"
            ], check=True, cwd=self.base_path)

            self.log(f"Successfully rolled back to: {tag}")
            return True

        except subprocess.CalledProcessError as e:
            self.log(f"Git rollback failed: {e}", "ERROR")
            return False

    def rollback_to_filesystem_backup(self, backup_path: Path) -> bool:
        """Rollback using filesystem backup"""
        try:
            self.log(f"Rolling back from filesystem backup: {backup_path}")

            src_path = self.base_path / "src" / "moai_adk"

            # Remove current source
            if src_path.exists():
                shutil.rmtree(src_path)

            # Restore from backup
            shutil.copytree(backup_path, src_path)

            # Reinstall package
            subprocess.run([
                "pip", "install", "-e", ".", "--force-reinstall"
            ], check=True, cwd=self.base_path)

            self.log(f"Successfully rolled back from filesystem backup")
            return True

        except Exception as e:
            self.log(f"Filesystem rollback failed: {e}", "ERROR")
            return False

    def verify_rollback(self) -> bool:
        """Verify that rollback was successful"""
        self.log("Verifying rollback success...")

        verification_tests = [
            ("Package Import", "python -c 'import moai_adk; print(\"Import OK\")'"),
            ("CLI Version", "moai --version"),
            ("Basic Functionality", "python -c 'from moai_adk import Config; c = Config(); print(\"Config OK\")'"),
        ]

        all_passed = True

        for test_name, command in verification_tests:
            try:
                result = subprocess.run(
                    command,
                    shell=True,
                    capture_output=True,
                    text=True,
                    timeout=30,
                    cwd=self.base_path
                )

                if result.returncode == 0:
                    self.log(f"✅ {test_name}: PASS")
                else:
                    self.log(f"❌ {test_name}: FAIL - {result.stderr.strip()}", "ERROR")
                    all_passed = False

            except subprocess.TimeoutExpired:
                self.log(f"❌ {test_name}: TIMEOUT", "ERROR")
                all_passed = False
            except Exception as e:
                self.log(f"❌ {test_name}: ERROR - {e}", "ERROR")
                all_passed = False

        return all_passed

    def clean_migration_artifacts(self) -> bool:
        """Clean up migration-related files and directories"""
        try:
            self.log("Cleaning migration artifacts...")

            artifacts_to_clean = [
                "migration.log",
                "rollback.log",
                "validation_report.json",
                "migration_backups",
                "scripts/migration_toolkit.py",
                "scripts/validate_migration.py",
                "scripts/rollback_migration.py",
                "ROUND_2_MIGRATION_STRATEGY.md",
            ]

            for artifact in artifacts_to_clean:
                artifact_path = self.base_path / artifact
                if artifact_path.exists():
                    if artifact_path.is_dir():
                        shutil.rmtree(artifact_path)
                        self.log(f"Removed directory: {artifact}")
                    else:
                        artifact_path.unlink()
                        self.log(f"Removed file: {artifact}")

            # Remove migration git tags
            try:
                result = subprocess.run([
                    "git", "tag", "-l", "migration-backup-*"
                ], capture_output=True, text=True, cwd=self.base_path)

                if result.returncode == 0:
                    for tag in result.stdout.strip().split('\n'):
                        if tag.startswith('migration-backup-'):
                            subprocess.run([
                                "git", "tag", "-d", tag
                            ], cwd=self.base_path)
                            self.log(f"Removed git tag: {tag}")

            except subprocess.CalledProcessError:
                self.log("Could not clean git tags", "WARNING")

            return True

        except Exception as e:
            self.log(f"Failed to clean artifacts: {e}", "ERROR")
            return False

    def interactive_rollback(self) -> bool:
        """Interactive rollback with user choices"""
        print("\n🗿 MoAI-ADK Migration Rollback Tool")
        print("=" * 40)

        backups = self.get_available_backups()

        if not backups:
            print("❌ No backups available for rollback!")
            print("Cannot proceed with rollback.")
            return False

        print(f"\n📁 Available backup points ({len(backups)}):")
        backup_list = list(backups.items())

        for i, (name, info) in enumerate(backup_list, 1):
            backup_type = info['type'].replace('_', ' ').title()
            print(f"  {i}. {name} ({backup_type})")

        try:
            choice = input(f"\nSelect backup to rollback to (1-{len(backup_list)}, or 'q' to quit): ").strip()

            if choice.lower() == 'q':
                print("Rollback cancelled by user.")
                return False

            selection = int(choice) - 1
            if selection < 0 or selection >= len(backup_list):
                print("❌ Invalid selection!")
                return False

            backup_name, backup_info = backup_list[selection]

            # Confirm rollback
            print(f"\n⚠️  ROLLBACK CONFIRMATION")
            print(f"About to rollback to: {backup_name}")
            print(f"Backup type: {backup_info['type']}")
            print(f"This will PERMANENTLY LOSE any changes made after this backup point!")

            confirm = input("\nType 'ROLLBACK' to confirm: ").strip()

            if confirm != 'ROLLBACK':
                print("Rollback cancelled.")
                return False

            # Perform rollback
            print(f"\n🔄 Rolling back to: {backup_name}")

            if backup_info['type'] == 'git_tag':
                success = self.rollback_to_git_tag(backup_info['tag'])
            elif backup_info['type'] == 'filesystem':
                success = self.rollback_to_filesystem_backup(backup_info['path'])
            else:
                print(f"❌ Unknown backup type: {backup_info['type']}")
                return False

            if not success:
                print("❌ Rollback failed!")
                return False

            # Verify rollback
            if not self.verify_rollback():
                print("❌ Rollback verification failed!")
                return False

            # Ask about cleanup
            cleanup = input("\nClean up migration artifacts? (y/N): ").strip().lower()
            if cleanup == 'y':
                self.clean_migration_artifacts()

            print("\n✅ Rollback completed successfully!")
            print("Package has been restored to pre-migration state.")

            return True

        except (ValueError, KeyboardInterrupt):
            print("\n❌ Invalid input or operation cancelled.")
            return False

    def emergency_rollback(self, backup_name: str = None) -> bool:
        """Emergency rollback without user interaction"""
        self.log("Starting emergency rollback...")

        backups = self.get_available_backups()

        if not backups:
            self.log("No backups available for emergency rollback!", "ERROR")
            return False

        # Use specified backup or the most recent one
        if backup_name and backup_name in backups:
            selected_backup = backup_name
            backup_info = backups[backup_name]
        else:
            # Use the first available backup (assuming most recent)
            selected_backup, backup_info = next(iter(backups.items()))

        self.log(f"Emergency rollback to: {selected_backup}")

        # Perform rollback
        if backup_info['type'] == 'git_tag':
            success = self.rollback_to_git_tag(backup_info['tag'])
        elif backup_info['type'] == 'filesystem':
            success = self.rollback_to_filesystem_backup(backup_info['path'])
        else:
            self.log(f"Unknown backup type: {backup_info['type']}", "ERROR")
            return False

        if not success:
            self.log("Emergency rollback failed!", "ERROR")
            return False

        # Verify rollback
        if not self.verify_rollback():
            self.log("Emergency rollback verification failed!", "ERROR")
            return False

        self.log("Emergency rollback completed successfully!")
        return True

    def status_check(self) -> Dict:
        """Check current migration status"""
        status = {
            "migration_in_progress": False,
            "migration_completed": False,
            "rollback_needed": False,
            "backups_available": 0,
            "current_structure": "unknown"
        }

        # Check for migration artifacts
        migration_files = [
            "migration.log",
            "ROUND_2_MIGRATION_STRATEGY.md",
            "scripts/migration_toolkit.py"
        ]

        for file in migration_files:
            if (self.base_path / file).exists():
                status["migration_in_progress"] = True
                break

        # Check current package structure
        src_path = self.base_path / "src" / "moai_adk"
        if src_path.exists():
            if (src_path / "utils").exists():
                status["current_structure"] = "new"
                status["migration_completed"] = True
            else:
                status["current_structure"] = "old"

        # Count available backups
        backups = self.get_available_backups()
        status["backups_available"] = len(backups)

        # Basic health check
        try:
            result = subprocess.run([
                "python", "-c", "import moai_adk; print('OK')"
            ], capture_output=True, text=True, cwd=self.base_path, timeout=10)

            if result.returncode != 0:
                status["rollback_needed"] = True

        except:
            status["rollback_needed"] = True

        return status


def main():
    """Main entry point for rollback script"""
    import argparse

    parser = argparse.ArgumentParser(description="MoAI-ADK Migration Rollback")
    parser.add_argument("--base-path", type=Path, default=Path.cwd(),
                        help="Base path of the project")
    parser.add_argument("--backup", type=str,
                        help="Specific backup to rollback to")
    parser.add_argument("--emergency", action="store_true",
                        help="Emergency rollback without confirmation")
    parser.add_argument("--status", action="store_true",
                        help="Check migration status")
    parser.add_argument("--list-backups", action="store_true",
                        help="List available backups")

    args = parser.parse_args()

    rollback = MigrationRollback(args.base_path)

    if args.status:
        status = rollback.status_check()
        print("\n🗿 MoAI-ADK Migration Status")
        print("=" * 30)
        for key, value in status.items():
            print(f"{key.replace('_', ' ').title()}: {value}")
        return

    if args.list_backups:
        backups = rollback.get_available_backups()
        print(f"\n📁 Available Backups ({len(backups)})")
        print("=" * 25)
        if backups:
            for name, info in backups.items():
                backup_type = info['type'].replace('_', ' ').title()
                print(f"- {name} ({backup_type})")
        else:
            print("No backups available.")
        return

    if args.emergency:
        success = rollback.emergency_rollback(args.backup)
        sys.exit(0 if success else 1)

    else:
        success = rollback.interactive_rollback()
        sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()