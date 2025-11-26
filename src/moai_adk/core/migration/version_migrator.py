"""
Main version migration orchestrator for MoAI-ADK

Coordinates version detection, backup creation, file migration,
and cleanup processes for automatic project upgrades.
"""

import logging
from pathlib import Path
from typing import Any, Dict

from .backup_manager import BackupManager
from .file_migrator import FileMigrator
from .version_detector import VersionDetector

logger = logging.getLogger(__name__)


class VersionMigrator:
    """Main migration orchestrator for MoAI-ADK version upgrades"""

    def __init__(self, project_root: Path):
        """
        Initialize version migrator

        Args:
            project_root: Root directory of the project
        """
        self.project_root = Path(project_root)
        self.detector = VersionDetector(project_root)
        self.backup_manager = BackupManager(project_root)
        self.file_migrator = FileMigrator(project_root)

    def detect_version(self) -> str:
        """
        Detect current project version

        Returns:
            Version string (e.g., "0.23.0", "0.24.0+", "unknown")
        """
        return self.detector.detect_version()

    def needs_migration(self) -> bool:
        """
        Check if project needs migration

        Returns:
            True if migration is needed, False otherwise
        """
        return self.detector.needs_migration()

    def get_migration_info(self) -> Dict[str, Any]:
        """
        Get detailed migration information

        Returns:
            Dictionary with migration details
        """
        version = self.detector.detect_version()
        needs_migration = self.detector.needs_migration()
        plan = self.detector.get_migration_plan()

        return {
            "current_version": version,
            "needs_migration": needs_migration,
            "target_version": "0.24.0" if needs_migration else version,
            "migration_plan": plan,
            "file_count": len(plan.get("move", [])),
        }

    def migrate_to_v024(self, dry_run: bool = False, cleanup: bool = True) -> bool:
        """
        Migrate project from v0.23.0 to v0.24.0

        Args:
            dry_run: If True, only show what would be done
            cleanup: If True, remove old files after successful migration

        Returns:
            True if migration was successful, False otherwise
        """
        if not self.needs_migration():
            logger.info("✅ Project is already up to date")
            return True

        if dry_run:
            logger.info("🔍 DRY RUN MODE - No changes will be made")
            self._show_migration_plan()
            return True

        logger.info("🚀 Starting migration to v0.24.0...")

        try:
            # Step 1: Create backup
            logger.info("\n💾 Step 1: Creating backup...")
            backup_path = self.backup_manager.create_backup("pre_v024_migration")
            logger.info(f"✅ Backup created: {backup_path}")

            # Step 2: Get migration plan
            logger.info("\n📋 Step 2: Analyzing migration requirements...")
            plan = self.detector.get_migration_plan()
            logger.info(f"   - {len(plan['create'])} directories to create")
            logger.info(f"   - {len(plan['move'])} files to move")

            # Step 3: Execute migration
            logger.info("\n🔄 Step 3: Executing migration...")
            results = self.file_migrator.execute_migration_plan(plan)

            if not results["success"]:
                logger.error("❌ Migration failed with errors:")
                for error in results["errors"]:
                    logger.error(f"   - {error}")
                logger.info(f"\n🔙 Rolling back using backup: {backup_path}")
                self.backup_manager.restore_backup(backup_path)
                return False

            logger.info("✅ Migration completed successfully!")
            logger.info(f"   - {results['created_dirs']} directories created")
            logger.info(f"   - {results['moved_files']} files moved")

            # Step 4: Verify migration
            logger.info("\n🔍 Step 4: Verifying migration...")
            if self._verify_migration():
                logger.info("✅ Verification passed")

                # Step 5: Cleanup (optional)
                if cleanup:
                    logger.info("\n🗑️  Step 5: Cleaning up old files...")
                    cleaned = self.file_migrator.cleanup_old_files(plan["cleanup"])
                    logger.info(f"✅ Cleaned up {cleaned} old files")
                else:
                    logger.info("\n⏭️  Step 5: Skipped cleanup (old files preserved)")

                logger.info("\n✨ Migration to v0.24.0 completed successfully!")
                return True
            else:
                logger.error("❌ Verification failed")
                logger.info(f"🔙 Rolling back using backup: {backup_path}")
                self.backup_manager.restore_backup(backup_path)
                return False

        except Exception as e:
            logger.error(f"❌ Migration failed with exception: {e}")
            logger.info("🔙 Attempting rollback...")
            try:
                latest_backup = self.backup_manager.get_latest_backup()
                if latest_backup:
                    self.backup_manager.restore_backup(latest_backup)
                    logger.info("✅ Rollback completed")
            except Exception as rollback_error:
                logger.error(f"❌ Rollback failed: {rollback_error}")
            return False

    def _show_migration_plan(self) -> None:
        """Display migration plan without executing"""
        plan = self.detector.get_migration_plan()

        print("\n📋 Migration Plan:")
        print("\n📁 Directories to create:")
        for directory in plan.get("create", []):
            print(f"   + {directory}")

        print("\n📄 Files to move:")
        for move_op in plan.get("move", []):
            print(f"   • {move_op['description']}")
            print(f"     {move_op['from']} → {move_op['to']}")

        print("\n🗑️  Files to cleanup after migration:")
        for cleanup_file in plan.get("cleanup", []):
            print(f"   - {cleanup_file}")

    def _verify_migration(self) -> bool:
        """
        Verify migration was successful

        Returns:
            True if verification passed
        """
        # Check that new config exists
        new_config = self.project_root / ".moai" / "config" / "config.json"
        if not new_config.exists():
            logger.error("Verification failed: new config.json not found")
            return False

        # Check that new config directory exists
        config_dir = self.project_root / ".moai" / "config"
        if not config_dir.is_dir():
            logger.error("Verification failed: config directory not found")
            return False

        logger.debug("All verification checks passed")
        return True

    def check_status(self) -> Dict[str, Any]:
        """
        Check migration status and return detailed information

        Returns:
            Dictionary with status information
        """
        version_info = self.detector.get_version_info()
        migration_info = self.get_migration_info()
        backups = self.backup_manager.list_backups()

        return {
            "version": version_info,
            "migration": migration_info,
            "backups": {
                "count": len(backups),
                "latest": backups[0] if backups else None,
            },
        }

    def rollback_to_latest_backup(self) -> bool:
        """
        Rollback to the most recent backup

        Returns:
            True if rollback was successful
        """
        latest_backup = self.backup_manager.get_latest_backup()

        if not latest_backup:
            logger.error("No backup found to rollback to")
            return False

        logger.info(f"🔙 Rolling back to backup: {latest_backup}")
        return self.backup_manager.restore_backup(latest_backup)
