#!/usr/bin/env python3
"""
MoAI-ADK 통합 체크포인트 시스템

@REQ:CHECKPOINT-SYSTEM-001
@FEATURE:CHECKPOINT-MANAGEMENT-001
@API:GET-CHECKPOINT
@DESIGN:UNIFIED-CHECKPOINT-001
"""

import json
import logging
from datetime import UTC, datetime, timedelta, timezone
from pathlib import Path
from typing import Any

from constants import (
    AUTO_CHECKPOINT_INTERVAL_MINUTES,
    BACKUP_RETENTION_DAYS,
    CHECKPOINT_MESSAGE_MAX_LENGTH,
    CHECKPOINT_TAG_PREFIX,
    ERROR_MESSAGES,
    MAX_CHECKPOINTS,
)
from git_helper import GitCommandError, GitHelper
from project_helper import ProjectHelper

logger = logging.getLogger(__name__)

# 한국 시간대 (KST) 상수 정의 - UTC+9
KST = timezone(timedelta(hours=9))

# 시간 관련 상수
CHECKPOINT_TIME_TOLERANCE_SECONDS = 60  # 시간 비교 허용 오차


def get_kst_now() -> datetime:
    """
    한국 시간(KST) 기준 현재 시간 반환

    Returns:
        datetime: KST 시간대가 적용된 현재 시간

    Note:
        기존 UTC 기반 체크포인트와의 호환성을 위해
        항상 시간대 정보를 포함하여 반환합니다.
    """
    return datetime.now(KST)


def convert_utc_to_kst(utc_datetime: datetime) -> datetime:
    """
    UTC 시간을 KST 시간으로 변환

    Args:
        utc_datetime: UTC 시간대의 datetime 객체

    Returns:
        datetime: KST로 변환된 datetime 객체

    Raises:
        ValueError: 입력값이 datetime 객체가 아닌 경우
    """
    if not isinstance(utc_datetime, datetime):
        raise ValueError("입력값은 datetime 객체여야 합니다")

    if utc_datetime.tzinfo is None:
        # timezone 정보가 없으면 UTC로 간주
        utc_datetime = utc_datetime.replace(tzinfo=UTC)

    return utc_datetime.astimezone(KST)


class CheckpointError(Exception):
    """체크포인트 관련 오류"""


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

    def to_dict(self) -> dict[str, Any]:
        return {
            "tag": self.tag,
            "commit_hash": self.commit_hash,
            "message": self.message,
            "created_at": self.created_at,
            "file_count": self.file_count,
            "is_auto": self.is_auto
        }

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "CheckpointInfo":
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

    def __init__(self, project_root: Path | None = None):
        self.project_root = project_root or ProjectHelper.find_project_root()
        self.git = GitHelper(self.project_root)
        self.checkpoints_dir = self.project_root / ".moai" / "checkpoints"
        self.metadata_file = self.checkpoints_dir / "metadata.json"
        self.checkpoints_dir.mkdir(parents=True, exist_ok=True)

    def create_checkpoint(self, message: str, is_auto: bool = False) -> CheckpointInfo:
        """
        체크포인트 생성

        Args:
            message: 체크포인트 메시지
            is_auto: 자동 체크포인트 여부

        Returns:
            CheckpointInfo: 생성된 체크포인트 정보

        Raises:
            CheckpointError: 체크포인트 생성 실패 시
        """
        # 입력 검증 강화
        if not isinstance(message, str):
            raise CheckpointError("메시지는 문자열이어야 합니다")

        if not message.strip():
            raise CheckpointError("메시지가 비어있습니다")

        # 가드절 적용
        if not self.git.is_git_repo():
            raise CheckpointError(ERROR_MESSAGES["not_git_repo"])

        try:
            # 메시지 길이 제한 적용
            sanitized_message = self._sanitize_checkpoint_message(message)

            # 변경사항 확인 및 처리
            has_changes = self.git.has_uncommitted_changes()
            if not has_changes:
                logger.info("변경사항이 없어 현재 HEAD에 태그만 생성합니다.")

            # 변경사항이 있는 경우에만 staging
            if has_changes:
                self._stage_changes_safely()

            # KST 기준 타임스탬프 생성
            kst_now = get_kst_now()
            timestamp = kst_now.strftime("%Y%m%d_%H%M%S")
            tag_name = f"{CHECKPOINT_TAG_PREFIX}{timestamp}"

            # 중복 태그 방지
            existing_tags = self._get_existing_tags()
            if tag_name in existing_tags:
                # 초 단위 추가로 중복 방지
                timestamp = kst_now.strftime("%Y%m%d_%H%M%S_%f")[:17]  # 마이크로초 3자리
                tag_name = f"{CHECKPOINT_TAG_PREFIX}{timestamp}"

            # Git 작업 수행
            commit_message = f"📍 {'Auto-' if is_auto else ''}Checkpoint: {sanitized_message}"
            commit_hash = self._create_commit_and_tag(
                has_changes, commit_message, tag_name
            )

            # 체크포인트 정보 생성
            checkpoint = CheckpointInfo(
                tag=tag_name,
                commit_hash=commit_hash,
                message=sanitized_message,
                created_at=kst_now.isoformat(),
                file_count=self._count_tracked_files(),
                is_auto=is_auto
            )

            # 메타데이터 저장 및 정리
            self._save_checkpoint_metadata(checkpoint)
            self._cleanup_old_checkpoints()

            logger.info(f"체크포인트 생성 완료: {tag_name} (KST: {kst_now.strftime('%Y-%m-%d %H:%M:%S')})")
            return checkpoint

        except GitCommandError as e:
            logger.error(f"Git 명령 실행 실패: {e}")
            raise CheckpointError(f"Git 명령 실행 실패: {e}")
        except CheckpointError:
            # CheckpointError는 그대로 전파
            raise
        except Exception as e:
            logger.error(f"체크포인트 생성 중 예상치 못한 오류: {e}")
            raise CheckpointError(f"체크포인트 생성 실패: {e}")

    def _sanitize_checkpoint_message(self, message: str) -> str:
        """체크포인트 메시지 정제 및 검증"""
        # 앞뒤 공백 제거
        message = message.strip()

        # 길이 제한 적용
        if len(message) > CHECKPOINT_MESSAGE_MAX_LENGTH:
            message = message[:CHECKPOINT_MESSAGE_MAX_LENGTH - 3] + "..."

        # 개행 문자 제거 (Git 태그 메시지에서 문제가 될 수 있음)
        message = message.replace('\n', ' ').replace('\r', ' ')

        # 연속된 공백 정리
        import re
        message = re.sub(r'\s+', ' ', message)

        return message

    def _stage_changes_safely(self) -> None:
        """안전한 변경사항 staging"""
        try:
            self.git.stage_all_changes()
        except GitCommandError as e:
            logger.error(f"변경사항 staging 실패: {e}")
            raise CheckpointError(f"변경사항을 staging할 수 없습니다: {e}")

    def _get_existing_tags(self) -> set:
        """기존 태그 목록 조회"""
        try:
            result = self.git.run_command(["git", "tag", "-l", f"{CHECKPOINT_TAG_PREFIX}*"])
            return set(result.stdout.strip().split('\n')) if result.stdout.strip() else set()
        except GitCommandError:
            logger.warning("기존 태그 목록 조회 실패, 빈 세트 반환")
            return set()

    def _create_commit_and_tag(self, has_changes: bool, commit_message: str, tag_name: str) -> str:
        """커밋 및 태그 생성"""
        try:
            if has_changes:
                commit_hash = self.git.commit(commit_message, allow_empty=False)
            else:
                commit_hash = self.git.run_command(["git", "rev-parse", "HEAD"]).stdout.strip()

            self.git.create_tag(tag_name, commit_message)
            return commit_hash

        except GitCommandError as e:
            logger.error(f"커밋 또는 태그 생성 실패: {e}")
            raise CheckpointError(f"Git 작업 실패: {e}")

    def list_checkpoints(self, limit: int | None = None) -> list[CheckpointInfo]:
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
        """
        자동 체크포인트 생성 조건 확인

        Returns:
            bool: 자동 체크포인트 생성이 필요한 경우 True

        Note:
            KST 기준으로 시간 간격을 계산합니다.
        """
        try:
            checkpoints = self.list_checkpoints(limit=1)
            if not checkpoints:
                logger.debug("기존 체크포인트가 없어 자동 체크포인트 생성 조건 충족")
                return True

            last_checkpoint = checkpoints[0]
            last_time = datetime.fromisoformat(last_checkpoint.created_at)
            now = get_kst_now()

            # 시간대가 다를 수 있으므로 KST로 통일하여 비교
            if last_time.tzinfo is None:
                # 레거시 데이터: UTC로 간주하고 KST로 변환
                last_time = last_time.replace(tzinfo=UTC).astimezone(KST)
            elif last_time.tzinfo != KST:
                # 다른 시간대인 경우 KST로 변환
                last_time = last_time.astimezone(KST)

            time_diff_minutes = (now - last_time).total_seconds() / 60
            needed = time_diff_minutes >= AUTO_CHECKPOINT_INTERVAL_MINUTES

            logger.debug(
                f"자동 체크포인트 조건 확인: "
                f"마지막={last_time.strftime('%Y-%m-%d %H:%M:%S')}, "
                f"현재={now.strftime('%Y-%m-%d %H:%M:%S')}, "
                f"차이={time_diff_minutes:.1f}분, "
                f"필요={needed}"
            )

            return needed

        except Exception as e:
            logger.error(f"자동 체크포인트 조건 확인 실패: {e}")
            # 안전을 위해 False 반환 (무한 체크포인트 생성 방지)
            return False

    def _find_checkpoint(self, tag_or_index: str) -> CheckpointInfo | None:
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

    def _load_checkpoint_metadata(self) -> dict[str, Any]:
        """체크포인트 메타데이터 로드"""
        if not self.metadata_file.exists():
            return {"checkpoints": [], "version": "1.0"}

        try:
            with open(self.metadata_file, encoding='utf-8') as f:
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

        cutoff_date = get_kst_now() - timedelta(days=BACKUP_RETENTION_DAYS)
        for checkpoint in checkpoints:
            created_date = datetime.fromisoformat(checkpoint.created_at)
            if created_date < cutoff_date and checkpoint.is_auto:
                self.delete_checkpoint(checkpoint.tag)

    def get_checkpoint_info(self, tag_or_index: str) -> CheckpointInfo | None:
        """태그 또는 인덱스로 체크포인트 정보 조회"""
        try:
            return self._find_checkpoint(tag_or_index)
        except Exception as exc:
            logger.error(f"체크포인트 조회 실패: {exc}")
            return None


def create_checkpoint(
    message: str, project_root: Path | None = None, is_auto: bool = False
) -> CheckpointInfo:
    """체크포인트 생성 편의 함수"""
    system = CheckpointSystem(project_root)
    return system.create_checkpoint(message, is_auto)


def rollback_to_checkpoint(
    tag_or_index: str, project_root: Path | None = None
) -> CheckpointInfo:
    """체크포인트 롤백 편의 함수"""
    system = CheckpointSystem(project_root)
    return system.rollback_to_checkpoint(tag_or_index)


def list_checkpoints(
    limit: int | None = None, project_root: Path | None = None
) -> list[CheckpointInfo]:
    """체크포인트 목록 조회 편의 함수"""
    system = CheckpointSystem(project_root)
    return system.list_checkpoints(limit)
