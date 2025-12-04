"""
Alfred to Moai folder structure migration for MoAI-ADK

Handles automatic migration from legacy alfred/ folders to new moai/ structure.
- Creates backup before migration
- Installs fresh moai/ templates from package
- Deletes alfred/ folders
- Updates settings.json Hook paths
- Records migration status in config.json
- Provides automatic rollback on failure
"""

import json
import logging
import shutil
from datetime import datetime
from pathlib import Path
from typing import Optional

from .backup_manager import BackupManager

logger = logging.getLogger(__name__)


class AlfredToMoaiMigrator:
    """Handles automatic migration from Alfred to Moai folder structure"""

    def __init__(self, project_root: Path):
        """
        Initialize Alfred to Moai migrator

        Args:
            project_root: Root directory of the project
        """
        self.project_root = Path(project_root)
        self.claude_root = self.project_root / ".claude"
        self.config_path = self.project_root / ".moai" / "config" / "config.json"
        self.settings_path = self.claude_root / "settings.json"
        self.backup_manager = BackupManager(project_root)

        # Define folder paths
        self.alfred_folders = {
            "commands": self.claude_root / "commands" / "alfred",
            "agents": self.claude_root / "agents" / "alfred",
            "hooks": self.claude_root / "hooks" / "alfred",
        }

        self.moai_folders = {
            "commands": self.claude_root / "commands" / "moai",
            "agents": self.claude_root / "agents" / "moai",
            "hooks": self.claude_root / "hooks" / "moai",
        }

    def _load_config(self) -> dict:
        """
        Load config.json

        Returns:
            Dictionary from config.json, or empty dict if not found
        """
        if not self.config_path.exists():
            return {}

        try:
            return json.loads(self.config_path.read_text(encoding="utf-8"))
        except Exception as e:
            logger.warning(f"Failed to load config.json: {e}")
            return {}

    def _save_config(self, config: dict) -> None:
        """
        Save config.json

        Args:
            config: Configuration dictionary to save

        Raises:
            Exception: If save fails
        """
        self.config_path.parent.mkdir(parents=True, exist_ok=True)
        self.config_path.write_text(json.dumps(config, indent=2, ensure_ascii=False))

    def needs_migration(self) -> bool:
        """
        Check if Alfred to Moai migration is needed

        Returns:
            True if migration is needed, False otherwise
        """
        # Check if already migrated
        try:
            config = self._load_config()
            migration_state = config.get("migration", {}).get("alfred_to_moai", {})
            if migration_state.get("migrated"):
                logger.info("ℹ️  Alfred → Moai migration already completed")
                if migration_state.get("timestamp"):
                    logger.info(f"Timestamp: {migration_state['timestamp']}")
                return False
        except Exception as e:
            logger.debug(f"Config check error: {e}")

        # Check if any alfred folder exists
        has_alfred = any(folder.exists() for folder in self.alfred_folders.values())

        if has_alfred:
            detected = [name for name, folder in self.alfred_folders.items() if folder.exists()]
            logger.debug(f"Alfred folders detected: {', '.join(detected)}")

        return has_alfred

    def execute_migration(self, backup_path: Optional[Path] = None) -> bool:
        """
        Execute Alfred to Moai migration

        Args:
            backup_path: Path to use for backup (if None, creates new backup)

        Returns:
            True if migration successful, False otherwise
        """
        try:
            logger.info("\n[1/5] Backing up project...")

            # Step 1: Create or use existing backup
            if backup_path is None:
                try:
                    backup_path = self.backup_manager.create_backup("alfred_to_moai_migration")
                    logger.info(f"✅ Backup completed: {backup_path}")
                except Exception as e:
                    logger.error("❌ Error: Backup failed")
                    logger.error(f"Cause: {str(e)}")
                    return False
            else:
                logger.info(f"✅ Using existing backup: {backup_path}")

            # Step 2: Detect alfred folders
            logger.info("\n[2/5] Alfred folders detected:")
            alfred_detected = {name: folder for name, folder in self.alfred_folders.items() if folder.exists()}

            if not alfred_detected:
                logger.warning("No Alfred folders found - skipping migration")
                return True

            logger.info(", ".join(alfred_detected.keys()))

            # Step 3: Verify moai folders exist (should be created in Phase 1)
            logger.info("\n[3/5] Verifying Moai template installation...")
            missing_moai = [name for name, folder in self.moai_folders.items() if not folder.exists()]

            if missing_moai:
                logger.error(f"❌ Missing Moai folders: {', '.join(missing_moai)}")
                logger.error("Phase 1 implementation required first (package template moai structure)")
                self._rollback_migration(backup_path)
                return False

            logger.info("✅ Moai templates installed")

            # Step 4: Update settings.json hooks
            logger.info("\n[4/5] Updating paths...")
            try:
                self._update_settings_json_hooks()
                logger.info("✅ settings.json Hook paths updated")
            except Exception as e:
                logger.error("❌ Error: Failed to update settings.json")
                logger.error(f"Cause: {str(e)}")
                self._rollback_migration(backup_path)
                return False

            # Step 5: Delete alfred folders
            logger.info("\n[5/5] Cleaning up...")
            try:
                self._delete_alfred_folders(alfred_detected)
                logger.info("✅ Alfred folders deleted")
            except Exception as e:
                logger.error("❌ Error: Failed to delete Alfred folders")
                logger.error(f"Cause: {str(e)}")
                self._rollback_migration(backup_path)
                return False

            # Step 6: Verify migration
            logger.info("\n[6/6] Verifying migration...")
            if not self._verify_migration():
                logger.error("❌ Migration verification failed")
                self._rollback_migration(backup_path)
                return False

            logger.info("✅ Migration verification passed")

            # Step 7: Record migration status
            logger.info("\nRecording migration status...")
            try:
                self._record_migration_state(backup_path, len(alfred_detected))
                logger.info("✅ Migration status recorded")
            except Exception as e:
                logger.warning(f"⚠️  Failed to record status: {str(e)}")
                # Don't rollback for this, migration was successful

            logger.info("\n✅ Alfred → Moai migration completed!")
            return True

        except Exception as e:
            logger.error(f"\n❌ Unexpected error: {str(e)}")
            if backup_path:
                self._rollback_migration(backup_path)
            return False

    def _delete_alfred_folders(self, alfred_detected: dict) -> None:
        """
        Delete Alfred folders

        Args:
            alfred_detected: Dictionary of detected alfred folders

        Raises:
            Exception: If deletion fails
        """
        for name, folder in alfred_detected.items():
            if folder.exists():
                try:
                    shutil.rmtree(folder)
                    logger.debug(f"Deleted: {folder}")
                except Exception as e:
                    raise Exception(f"Failed to delete {name} folder: {str(e)}")

    def _update_settings_json_hooks(self) -> None:
        """
        Update settings.json to replace alfred paths with moai paths

        Raises:
            Exception: If update fails
        """
        if not self.settings_path.exists():
            logger.warning(f"settings.json file missing: {self.settings_path}")
            return

        try:
            # Read settings.json
            with open(self.settings_path, "r", encoding="utf-8") as f:
                settings_content = f.read()

            # Replace all alfred references with moai
            # Pattern: .claude/hooks/alfred/ → .claude/hooks/moai/
            updated_content = settings_content.replace(".claude/hooks/alfred/", ".claude/hooks/moai/")
            updated_content = updated_content.replace(".claude/commands/alfred/", ".claude/commands/moai/")
            updated_content = updated_content.replace(".claude/agents/alfred/", ".claude/agents/moai/")

            # Write back to file
            with open(self.settings_path, "w", encoding="utf-8") as f:
                f.write(updated_content)

            # Verify JSON validity
            with open(self.settings_path, "r", encoding="utf-8") as f:
                json.load(f)  # This will raise if JSON is invalid

            logger.debug("settings.json update and verification completed")

        except json.JSONDecodeError as e:
            raise Exception(f"settings.json JSON format error: {str(e)}")
        except Exception as e:
            raise Exception(f"Failed to update settings.json: {str(e)}")

    def _verify_migration(self) -> bool:
        """
        Verify migration was successful

        Returns:
            True if migration is valid, False otherwise
        """
        # Check moai folders exist
        for name, folder in self.moai_folders.items():
            if not folder.exists():
                logger.error(f"❌ Missing Moai {name} folder: {folder}")
                return False

        # Check alfred folders are deleted
        for name, folder in self.alfred_folders.items():
            if folder.exists():
                logger.warning(f"⚠️  Alfred {name} folder still exists: {folder}")
                return False

        # Check settings.json hooks paths (ignore pattern matching strings like "Bash(alfred:*)")
        if self.settings_path.exists():
            try:
                with open(self.settings_path, "r", encoding="utf-8") as f:
                    settings_content = f.read()

                # Only check for hooks/alfred paths, not pattern strings
                if (
                    ".claude/hooks/alfred/" in settings_content
                    or ".claude/commands/alfred/" in settings_content
                    or ".claude/agents/alfred/" in settings_content
                ):
                    logger.error("❌ settings.json still contains alfred hook paths")
                    return False

                if "moai" not in settings_content.lower():
                    logger.warning("⚠️  No moai references in settings.json")

            except Exception as e:
                logger.error(f"❌ settings.json verification failed: {str(e)}")
                return False

        logger.debug("Migration verification completed")
        return True

    def _record_migration_state(self, backup_path: Path, folders_count: int) -> None:
        """
        Record migration state in config.json

        Args:
            backup_path: Path to the backup
            folders_count: Number of folders migrated

        Raises:
            Exception: If recording fails
        """
        try:
            config = self._load_config()

            # Initialize migration section if not exists
            if "migration" not in config:
                config["migration"] = {}

            config["migration"]["alfred_to_moai"] = {
                "migrated": True,
                "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                "folders_installed": 3,  # commands, agents, hooks
                "folders_removed": folders_count,
                "backup_path": str(backup_path),
                "package_version": self._get_package_version(),
            }

            self._save_config(config)
            logger.debug("Migration state recorded in config.json")

        except Exception as e:
            raise Exception(f"Migration state recording failed: {str(e)}")

    def _rollback_migration(self, backup_path: Path) -> None:
        """
        Rollback migration from backup

        Args:
            backup_path: Path to the backup to restore from
        """
        try:
            logger.info("\n🔄 Starting automatic rollback...")
            logger.info("[1/3] Restoring project...")

            # Restore from backup
            self.backup_manager.restore_backup(backup_path)

            logger.info("✅ Project restored")
            logger.info("[2/3] Resetting migration state...")

            # Clear migration state in config
            try:
                config = self._load_config()
                if "migration" in config and "alfred_to_moai" in config["migration"]:
                    del config["migration"]["alfred_to_moai"]
                    self._save_config(config)
            except Exception as e:
                logger.warning(f"⚠️  Failed to reset state: {str(e)}")

            logger.info("✅ Rollback completed")
            logger.info("💡 Tip: Run `moai-adk update` again after resolving the error")

        except Exception as e:
            logger.error(f"\n❌ Rollback failed: {str(e)}")
            logger.error(f"⚠️  Manual recovery required: Please restore manually from backup: {backup_path}")

    def _get_package_version(self) -> str:
        """
        Get current package version

        Returns:
            Version string
        """
        try:
            config = self._load_config()
            return config.get("moai", {}).get("version", "unknown")
        except Exception:
            return "unknown"
