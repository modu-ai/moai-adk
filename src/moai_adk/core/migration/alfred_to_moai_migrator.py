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
                logger.info("ℹ️  Alfred → Moai 마이그레이션이 이미 완료되었습니다")
                if migration_state.get("timestamp"):
                    logger.info(f"타임스탬프: {migration_state['timestamp']}")
                return False
        except Exception as e:
            logger.debug(f"Config check error: {e}")

        # Check if any alfred folder exists
        has_alfred = any(folder.exists() for folder in self.alfred_folders.values())

        if has_alfred:
            detected = [
                name
                for name, folder in self.alfred_folders.items()
                if folder.exists()
            ]
            logger.debug(f"Alfred 폴더 감지됨: {', '.join(detected)}")

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
            logger.info("\n[1/5] 프로젝트 백업 중...")

            # Step 1: Create or use existing backup
            if backup_path is None:
                try:
                    backup_path = self.backup_manager.create_backup(
                        "alfred_to_moai_migration"
                    )
                    logger.info(f"✅ 백업 완료: {backup_path}")
                except Exception as e:
                    logger.error("❌ 에러: 백업 생성 실패")
                    logger.error(f"원인: {str(e)}")
                    return False
            else:
                logger.info(f"✅ 기존 백업 사용: {backup_path}")

            # Step 2: Detect alfred folders
            logger.info("\n[2/5] Alfred 폴더 감지됨: ", end="")
            alfred_detected = {
                name: folder
                for name, folder in self.alfred_folders.items()
                if folder.exists()
            }

            if not alfred_detected:
                logger.warning("Alfred 폴더 없음 - 마이그레이션 건너뜀")
                return True

            logger.info(", ".join(alfred_detected.keys()))

            # Step 3: Verify moai folders exist (should be created in Phase 1)
            logger.info("\n[3/5] Moai 템플릿 설치 확인 중...")
            missing_moai = [
                name
                for name, folder in self.moai_folders.items()
                if not folder.exists()
            ]

            if missing_moai:
                logger.error(
                    f"❌ Moai 폴더 없음: {', '.join(missing_moai)}"
                )
                logger.error("Phase 1 구현이 먼저 필요합니다 (package template moai 구조)")
                self._rollback_migration(backup_path)
                return False

            logger.info("✅ Moai 템플릿 설치됨")

            # Step 4: Update settings.json hooks
            logger.info("\n[4/5] 경로 업데이트 중...")
            try:
                self._update_settings_json_hooks()
                logger.info("✅ settings.json Hook 경로 업데이트됨")
            except Exception as e:
                logger.error("❌ 에러: settings.json 업데이트 실패")
                logger.error(f"원인: {str(e)}")
                self._rollback_migration(backup_path)
                return False

            # Step 5: Delete alfred folders
            logger.info("\n[5/5] 정리 작업 중...")
            try:
                self._delete_alfred_folders(alfred_detected)
                logger.info("✅ Alfred 폴더 삭제됨")
            except Exception as e:
                logger.error("❌ 에러: Alfred 폴더 삭제 실패")
                logger.error(f"원인: {str(e)}")
                self._rollback_migration(backup_path)
                return False

            # Step 6: Verify migration
            logger.info("\n[6/6] 마이그레이션 검증 중...")
            if not self._verify_migration():
                logger.error("❌ 마이그레이션 검증 실패")
                self._rollback_migration(backup_path)
                return False

            logger.info("✅ 마이그레이션 검증 통과")

            # Step 7: Record migration status
            logger.info("\n마이그레이션 상태 기록 중...")
            try:
                self._record_migration_state(backup_path, len(alfred_detected))
                logger.info("✅ 마이그레이션 상태 기록됨")
            except Exception as e:
                logger.warning(f"⚠️  상태 기록 실패: {str(e)}")
                # Don't rollback for this, migration was successful

            logger.info("\n✅ Alfred → Moai 마이그레이션 완료!")
            return True

        except Exception as e:
            logger.error(f"\n❌ 예상치 못한 에러: {str(e)}")
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
                    logger.debug(f"삭제됨: {folder}")
                except Exception as e:
                    raise Exception(f"{name} 폴더 삭제 실패: {str(e)}")

    def _update_settings_json_hooks(self) -> None:
        """
        Update settings.json to replace alfred paths with moai paths

        Raises:
            Exception: If update fails
        """
        if not self.settings_path.exists():
            logger.warning(f"settings.json 파일 없음: {self.settings_path}")
            return

        try:
            # Read settings.json
            with open(self.settings_path, "r", encoding="utf-8") as f:
                settings_content = f.read()

            # Replace all alfred references with moai
            # Pattern: .claude/hooks/alfred/ → .claude/hooks/moai/
            updated_content = settings_content.replace(
                ".claude/hooks/alfred/", ".claude/hooks/moai/"
            )
            updated_content = updated_content.replace(
                ".claude/commands/alfred/", ".claude/commands/moai/"
            )
            updated_content = updated_content.replace(
                ".claude/agents/alfred/", ".claude/agents/moai/"
            )

            # Write back to file
            with open(self.settings_path, "w", encoding="utf-8") as f:
                f.write(updated_content)

            # Verify JSON validity
            with open(self.settings_path, "r", encoding="utf-8") as f:
                json.load(f)  # This will raise if JSON is invalid

            logger.debug("settings.json 업데이트 및 검증 완료")

        except json.JSONDecodeError as e:
            raise Exception(f"settings.json JSON 형식 오류: {str(e)}")
        except Exception as e:
            raise Exception(f"settings.json 업데이트 실패: {str(e)}")

    def _verify_migration(self) -> bool:
        """
        Verify migration was successful

        Returns:
            True if migration is valid, False otherwise
        """
        # Check moai folders exist
        for name, folder in self.moai_folders.items():
            if not folder.exists():
                logger.error(f"❌ Moai {name} 폴더 없음: {folder}")
                return False

        # Check alfred folders are deleted
        for name, folder in self.alfred_folders.items():
            if folder.exists():
                logger.warning(f"⚠️  Alfred {name} 폴더 아직 존재: {folder}")
                return False

        # Check settings.json has no alfred references
        if self.settings_path.exists():
            try:
                with open(self.settings_path, "r", encoding="utf-8") as f:
                    settings_content = f.read()

                if "alfred" in settings_content.lower():
                    logger.error("❌ settings.json에 아직 alfred 참조 존재")
                    return False

                if "moai" not in settings_content.lower():
                    logger.warning("⚠️  settings.json에 moai 참조 없음")

            except Exception as e:
                logger.error(f"❌ settings.json 검증 실패: {str(e)}")
                return False

        logger.debug("마이그레이션 검증 완료")
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
            logger.info("\n🔄 자동 롤백 시작...")
            logger.info("[1/3] 프로젝트 복원 중...")

            # Restore from backup
            self.backup_manager.restore_backup(backup_path)

            logger.info("✅ 프로젝트 복원됨")
            logger.info("[2/3] 마이그레이션 상태 초기화...")

            # Clear migration state in config
            try:
                config = self.config_manager.load()
                if "migration" in config and "alfred_to_moai" in config["migration"]:
                    del config["migration"]["alfred_to_moai"]
                    self.config_manager.save(config)
            except Exception as e:
                logger.warning(f"⚠️  상태 초기화 실패: {str(e)}")

            logger.info("✅ 롤백 완료")
            logger.info(
                "💡 팁: 에러를 해결한 후 다시 `moai-adk update` 명령을 실행하세요"
            )

        except Exception as e:
            logger.error(f"\n❌ 롤백 실패: {str(e)}")
            logger.error(
                "⚠️  수동 복구 필요: 백업에서 수동으로 복원하세요: "
                f"{backup_path}"
            )

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
