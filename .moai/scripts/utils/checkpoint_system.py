#!/usr/bin/env python3
"""
MoAI-ADK 통합 체크포인트 시스템

@REQ:CHECKPOINT-SYSTEM-001
@FEATURE:CHECKPOINT-MANAGEMENT-001
@API:GET-CHECKPOINT
@DESIGN:UNIFIED-CHECKPOINT-001
"""

import json
from datetime import datetime, timezone, timedelta
from pathlib import Path
from typing import Dict, List, Optional, Any
import logging

from constants import (
    CHECKPOINT_TAG_PREFIX, MAX_CHECKPOINTS, CHECKPOINT_MESSAGE_MAX_LENGTH,
    AUTO_CHECKPOINT_INTERVAL_MINUTES, BACKUP_RETENTION_DAYS,
    ERROR_MESSAGES
)
from git_helper import GitHelper, GitCommandError
from project_helper import ProjectHelper

logger = logging.getLogger(__name__)


class CheckpointError(Exception):
    """체크포인트 관련 오류"""
    pass


class CheckpointInfo:
    """체크포인트 정보"""

    def __init__(self, tag: str, commit_hash: str, message: str, created_at: str,
                 file_count: int = 0, is_auto: bool = False):
        self.tag = tag
        self.commit_hash = commit_hash
        self.message = message
        self.created_at = created_at
        self.file_count = file_count
        self.is_auto = is_auto

    def to_dict(self) -> Dict[str, Any]:
        return {
            "tag": self.tag,
            "commit_hash": self.commit_hash,
            "message": self.message,
            "created_at": self.created_at,
            "file_count": self.file_count,
            "is_auto": self.is_auto
        }

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "CheckpointInfo":
        # 새 형식
        if "tag" in data:
            return cls(
                tag=data["tag"],
                commit_hash=data["commit_hash"],
                message=data["message"],
                created_at=data["created_at"],
                file_count=data.get("file_count", 0),
                is_auto=data.get("is_auto", False)
            )
        # 구 형식 호환성
        elif "id" in data:
            return cls(
                tag=f"moai_cp/{data['id'].replace('checkpoint_', '')}",
                commit_hash=data["commit"],
                message=data.get("message", "Legacy checkpoint"),
                created_at=data["timestamp"],
                file_count=data.get("files_changed", 0),
                is_auto=data.get("kind") == "auto"
            )
        else:
            raise ValueError(f"Invalid checkpoint data format: {data}")


class CheckpointSystem:
    """통합 체크포인트 관리 시스템"""

    def __init__(self, project_root: Optional[Path] = None):
        self.project_root = project_root or ProjectHelper.find_project_root()
        self.git = GitHelper(self.project_root)
        self.checkpoints_dir = self.project_root / ".moai" / "checkpoints"
        self.metadata_file = self.checkpoints_dir / "metadata.json"
        self.checkpoints_dir.mkdir(parents=True, exist_ok=True)

    def create_checkpoint(self, message: str, is_auto: bool = False) -> CheckpointInfo:
        """체크포인트 생성"""
        try:
            if not self.git.is_git_repo():
                raise CheckpointError(ERROR_MESSAGES["not_git_repo"])

            if len(message) > CHECKPOINT_MESSAGE_MAX_LENGTH:
                message = message[:CHECKPOINT_MESSAGE_MAX_LENGTH - 3] + "..."

            has_changes = self.git.has_uncommitted_changes()
            if not has_changes:
                logger.info("변경사항이 없어 현재 HEAD에 태그만 생성합니다.")

            if has_changes:
                self.git.stage_all_changes()

            timestamp = datetime.now(timezone.utc).strftime("%Y%m%d_%H%M%S")
            tag_name = f"{CHECKPOINT_TAG_PREFIX}{timestamp}"

            commit_message = f"📍 {'Auto-' if is_auto else ''}Checkpoint: {message}"
            if has_changes:
                commit_hash = self.git.commit(commit_message, allow_empty=False)
            else:
                commit_hash = self.git.run_command(["git", "rev-parse", "HEAD"]).stdout.strip()
            self.git.create_tag(tag_name, commit_message)

            checkpoint = CheckpointInfo(
                tag=tag_name,
                commit_hash=commit_hash,
                message=message,
                created_at=datetime.now(timezone.utc).isoformat(),
                file_count=self._count_tracked_files(),
                is_auto=is_auto
            )

            self._save_checkpoint_metadata(checkpoint)
            self._cleanup_old_checkpoints()

            logger.info(f"체크포인트 생성 완료: {tag_name}")
            return checkpoint

        except GitCommandError as e:
            raise CheckpointError(f"Git 명령 실행 실패: {e}")
        except Exception as e:
            raise CheckpointError(f"체크포인트 생성 실패: {e}")

    def list_checkpoints(self, limit: Optional[int] = None) -> List[CheckpointInfo]:
        """체크포인트 목록 조회"""
        try:
            metadata = self._load_checkpoint_metadata()
            checkpoints = []

            for cp in metadata.get("checkpoints", []):
                try:
                    checkpoint = CheckpointInfo.from_dict(cp)
                    checkpoints.append(checkpoint)
                except Exception as e:
                    logger.warning(f"체크포인트 정보 파싱 실패: {e}, data: {cp}")
                    continue

            checkpoints.sort(key=lambda x: x.created_at, reverse=True)

            if limit:
                checkpoints = checkpoints[:limit]

            return checkpoints

        except Exception as e:
            logger.error(f"체크포인트 목록 조회 실패: {e}")
            return []

    def rollback_to_checkpoint(self, tag_or_index: str) -> CheckpointInfo:
        """체크포인트로 롤백"""
        try:
            if self.git.has_uncommitted_changes():
                stash_id = self.git.stash_push("Pre-rollback stash")
                logger.info(f"변경사항을 임시 저장했습니다: {stash_id}")

            checkpoint = self._find_checkpoint(tag_or_index)
            if not checkpoint:
                raise CheckpointError(f"체크포인트를 찾을 수 없습니다: {tag_or_index}")

            self.git.run_command(["git", "reset", "--hard", checkpoint.commit_hash])
            logger.info(f"체크포인트로 롤백 완료: {checkpoint.tag}")
            return checkpoint

        except GitCommandError as e:
            raise CheckpointError(f"롤백 실패: {e}")
        except Exception as e:
            raise CheckpointError(f"롤백 중 오류 발생: {e}")

    def delete_checkpoint(self, tag_or_index: str) -> bool:
        """체크포인트 삭제"""
        try:
            checkpoint = self._find_checkpoint(tag_or_index)
            if not checkpoint:
                return False

            try:
                self.git.run_command(["git", "tag", "-d", checkpoint.tag])
            except GitCommandError:
                logger.warning(f"Git 태그 삭제 실패: {checkpoint.tag}")

            backup_path = self.checkpoints_dir / f"{checkpoint.tag}.tar.gz"
            if backup_path.exists():
                backup_path.unlink()

            self._remove_checkpoint_metadata(checkpoint.tag)
            logger.info(f"체크포인트 삭제 완료: {checkpoint.tag}")
            return True

        except Exception as e:
            logger.error(f"체크포인트 삭제 실패: {e}")
            return False

    def should_create_auto_checkpoint(self) -> bool:
        """자동 체크포인트 생성 조건 확인"""
        try:
            checkpoints = self.list_checkpoints(limit=1)
            if not checkpoints:
                return True

            last_checkpoint = checkpoints[0]
            last_time = datetime.fromisoformat(last_checkpoint.created_at)
            now = datetime.now(timezone.utc)

            time_diff = (now - last_time).total_seconds() / 60
            return time_diff >= AUTO_CHECKPOINT_INTERVAL_MINUTES

        except Exception as e:
            logger.error(f"자동 체크포인트 조건 확인 실패: {e}")
            return False

    def _find_checkpoint(self, tag_or_index: str) -> Optional[CheckpointInfo]:
        """태그 또는 인덱스로 체크포인트 찾기"""
        checkpoints = self.list_checkpoints()

        if tag_or_index.isdigit():
            index = int(tag_or_index)
            if 0 <= index < len(checkpoints):
                return checkpoints[index]

        for checkpoint in checkpoints:
            if checkpoint.tag == tag_or_index or checkpoint.tag.endswith(tag_or_index):
                return checkpoint

        return None

    def _count_tracked_files(self) -> int:
        """추적 중인 파일 수 계산"""
        try:
            result = self.git.run_command(["git", "ls-files"])
            return len(result.stdout.splitlines())
        except GitCommandError:
            return 0

    def _load_checkpoint_metadata(self) -> Dict[str, Any]:
        """체크포인트 메타데이터 로드"""
        if not self.metadata_file.exists():
            return {"checkpoints": [], "version": "1.0"}

        try:
            with open(self.metadata_file, 'r', encoding='utf-8') as f:
                return json.load(f)
        except (json.JSONDecodeError, Exception) as e:
            logger.error(f"메타데이터 로드 실패: {e}")
            return {"checkpoints": [], "version": "1.0"}

    def _save_checkpoint_metadata(self, checkpoint: CheckpointInfo) -> None:
        """체크포인트 메타데이터 저장"""
        metadata = self._load_checkpoint_metadata()
        metadata["checkpoints"] = [
            cp for cp in metadata["checkpoints"]
            if cp.get("tag") != checkpoint.tag
        ]
        metadata["checkpoints"].append(checkpoint.to_dict())

        with open(self.metadata_file, 'w', encoding='utf-8') as f:
            json.dump(metadata, f, indent=2, ensure_ascii=False)

    def _remove_checkpoint_metadata(self, tag: str) -> None:
        """체크포인트 메타데이터에서 제거"""
        metadata = self._load_checkpoint_metadata()
        metadata["checkpoints"] = [
            cp for cp in metadata["checkpoints"]
            if cp.get("tag") != tag
        ]

        with open(self.metadata_file, 'w', encoding='utf-8') as f:
            json.dump(metadata, f, indent=2, ensure_ascii=False)

    def _cleanup_old_checkpoints(self) -> None:
        """오래된 체크포인트 정리"""
        checkpoints = self.list_checkpoints()

        if len(checkpoints) > MAX_CHECKPOINTS:
            old_checkpoints = checkpoints[MAX_CHECKPOINTS:]
            for checkpoint in old_checkpoints:
                self.delete_checkpoint(checkpoint.tag)

        cutoff_date = datetime.now(timezone.utc) - timedelta(days=BACKUP_RETENTION_DAYS)
        for checkpoint in checkpoints:
            created_date = datetime.fromisoformat(checkpoint.created_at)
            if created_date < cutoff_date and checkpoint.is_auto:
                self.delete_checkpoint(checkpoint.tag)

    def get_checkpoint_info(self, tag_or_index: str) -> Optional[CheckpointInfo]:
        """태그 또는 인덱스로 체크포인트 정보 조회"""
        try:
            return self._find_checkpoint(tag_or_index)
        except Exception as exc:
            logger.error(f"체크포인트 조회 실패: {exc}")
            return None


def create_checkpoint(
    message: str, project_root: Optional[Path] = None, is_auto: bool = False
) -> CheckpointInfo:
    """체크포인트 생성 편의 함수"""
    system = CheckpointSystem(project_root)
    return system.create_checkpoint(message, is_auto)


def rollback_to_checkpoint(
    tag_or_index: str, project_root: Optional[Path] = None
) -> CheckpointInfo:
    """체크포인트 롤백 편의 함수"""
    system = CheckpointSystem(project_root)
    return system.rollback_to_checkpoint(tag_or_index)


def list_checkpoints(
    limit: Optional[int] = None, project_root: Optional[Path] = None
) -> List[CheckpointInfo]:
    """체크포인트 목록 조회 편의 함수"""
    system = CheckpointSystem(project_root)
    return system.list_checkpoints(limit)
