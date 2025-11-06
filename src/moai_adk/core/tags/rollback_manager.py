#!/usr/bin/env python3
# @CODE:TAG-ROLLBACK-MANAGER-001 | @SPEC:TAG-ROLLBACK-001
"""TAG 시스템 롤백 관리자

TAG 정책 시스템에서 문제 발생 시 안전한 롤백을 제공하는 관리자.
체크포인트 기반 복구와 이력 추적 기능을 지원.

주요 기능:
- 체크포인트 생성 및 관리
- 안전한 롤백 실행
- 이력 추적 및 로깅
- 비상 복구 시스템

@SPEC:TAG-ROLLBACK-001
"""

import json
import shutil
import time
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional


@dataclass
class Checkpoint:
    """체크포인트 정보

    Attributes:
        id: 고유 체크포인트 ID
        timestamp: 생성 시간
        description: 체크포인트 설명
        file_states: 파일 상태 정보
        metadata: 추가 메타데이터
    """
    id: str
    timestamp: datetime
    description: str
    file_states: Dict[str, str]  # {file_path: content_hash}
    metadata: Dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        """딕셔너리로 변환"""
        return {
            "id": self.id,
            "timestamp": self.timestamp.isoformat(),
            "description": self.description,
            "file_states": self.file_states,
            "metadata": self.metadata
        }


@dataclass
class RollbackConfig:
    """롤백 시스템 설정

    Attributes:
        checkpoints_dir: 체크포인트 저장 디렉토리
        max_checkpoints: 최대 체크포인트 수
        auto_cleanup: 자동 정리 활성화
        backup_before_rollback: 롤백 전 백업 생성
        rollback_timeout: 롤백 타임아웃 (초)
    """
    checkpoints_dir: str = ".moai/checkpoints"
    max_checkpoints: int = 10
    auto_cleanup: bool = True
    backup_before_rollback: bool = True
    rollback_timeout: int = 30


class RollbackManager:
    """TAG 시스템 롤백 관리자

    체크포인트 기반 롤백 시스템을 제공하여 TAG 정책 시스템의 안정성 보장.
    문제 발생 시 신속하고 안전한 복구를 지원.

    Usage:
        config = RollbackConfig()
        manager = RollbackManager(config=config)

        # 체크포인트 생성
        checkpoint_id = manager.create_checkpoint("작업 전 상태")

        # 롤백 실행
        success = manager.rollback_to_checkpoint(checkpoint_id)

        # 최신 체크포인트로 롤백
        success = manager.rollback_to_latest()
    """

    def __init__(self, config: Optional[RollbackConfig] = None):
        """초기화

        Args:
            config: 롤백 설정 (기본: RollbackConfig())
        """
        self.config = config or RollbackConfig()
        self.checkpoints_dir = Path(self.config.checkpoints_dir)
        self.checkpoints_dir.mkdir(parents=True, exist_ok=True)

    def create_checkpoint(self, description: str,
                         files: Optional[List[str]] = None,
                         metadata: Optional[Dict[str, Any]] = None) -> str:
        """체크포인트 생성

        Args:
            description: 체크포인트 설명
            files: 포함할 파일 목록 (None이면 자동 탐지)
            metadata: 추가 메타데이터

        Returns:
            생성된 체크포인트 ID
        """
        checkpoint_id = self._generate_checkpoint_id()
        timestamp = datetime.now()

        # 파일 상태 수집
        if files is None:
            files = self._discover_project_files()

        file_states = self._collect_file_states(files)

        # 체크포인트 생성
        checkpoint = Checkpoint(
            id=checkpoint_id,
            timestamp=timestamp,
            description=description,
            file_states=file_states,
            metadata=metadata or {}
        )

        # 체크포인트 저장
        self._save_checkpoint(checkpoint)

        # 파일 백업 생성
        self._backup_files(checkpoint_id, files)

        # 오래된 체크포인트 정리
        if self.config.auto_cleanup:
            self._cleanup_old_checkpoints()

        return checkpoint_id

    def rollback_to_checkpoint(self, checkpoint_id: str) -> bool:
        """특정 체크포인트로 롤백

        Args:
            checkpoint_id: 롤백할 체크포인트 ID

        Returns:
            성공 여부
        """
        try:
            checkpoint = self._load_checkpoint(checkpoint_id)
            if not checkpoint:
                return False

            # 롤백 전 백업 생성
            if self.config.backup_before_rollback:
                self._create_rollback_backup(checkpoint_id)

            # 파일 복원
            success = self._restore_files(checkpoint)

            if success:
                # 롤백 로그 기록
                self._log_rollback(checkpoint_id, checkpoint)

            return success

        except Exception:
            return False

    def rollback_to_latest(self) -> bool:
        """최신 체크포인트로 롤백

        Returns:
            성공 여부
        """
        latest_id = self.get_latest_checkpoint_id()
        if not latest_id:
            return False

        return self.rollback_to_checkpoint(latest_id)

    def list_checkpoints(self) -> List[Dict[str, Any]]:
        """체크포인트 목록 조회

        Returns:
            체크포인트 정보 목록
        """
        checkpoints = []

        for checkpoint_file in self.checkpoints_dir.glob("checkpoint_*.json"):
            try:
                with open(checkpoint_file, 'r', encoding='utf-8') as f:
                    checkpoint_data = json.load(f)
                    checkpoints.append(checkpoint_data)
            except Exception:
                continue

        # 시간순 정렬 (최신 먼저)
        checkpoints.sort(key=lambda x: x.get('timestamp', ''), reverse=True)

        return checkpoints

    def get_latest_checkpoint_id(self) -> Optional[str]:
        """최신 체크포인트 ID 조회

        Returns:
            최신 체크포인트 ID 또는 None
        """
        checkpoints = self.list_checkpoints()
        return checkpoints[0]['id'] if checkpoints else None

    def delete_checkpoint(self, checkpoint_id: str) -> bool:
        """체크포인트 삭제

        Args:
            checkpoint_id: 삭제할 체크포인트 ID

        Returns:
            성공 여부
        """
        try:
            # 체크포인트 파일 삭제
            checkpoint_file = self.checkpoints_dir / f"checkpoint_{checkpoint_id}.json"
            if checkpoint_file.exists():
                checkpoint_file.unlink()

            # 백업 파일 삭제
            backup_dir = self.checkpoints_dir / f"backup_{checkpoint_id}"
            if backup_dir.exists():
                shutil.rmtree(backup_dir)

            return True

        except Exception:
            return False

    def emergency_rollback(self) -> bool:
        """비상 롤백

        모든 변경 사항을 취소하고 가장 안전한 상태로 복귀.

        Returns:
            성공 여부
        """
        try:
            # 가장 오래된 안정 체크포인트 찾기
            stable_checkpoints = self._find_stable_checkpoints()
            if not stable_checkpoints:
                return False

            # 가장 오래된 안정 체크포인트로 롤백
            oldest_stable = stable_checkpoints[-1]  # 오래된 순서 정렬
            return self.rollback_to_checkpoint(oldest_stable['id'])

        except Exception:
            return False

    def validate_checkpoint_integrity(self, checkpoint_id: str) -> bool:
        """체크포인트 무결성 검증

        Args:
            checkpoint_id: 검증할 체크포인트 ID

        Returns:
            무결성 여부
        """
        try:
            checkpoint = self._load_checkpoint(checkpoint_id)
            if not checkpoint:
                return False

            # 파일 무결성 검증
            backup_dir = self.checkpoints_dir / f"backup_{checkpoint_id}"
            if not backup_dir.exists():
                return False

            # 백업 파일 존재 확인
            for file_path in checkpoint.file_states.keys():
                backup_file = backup_dir / file_path.replace('/', '_')
                if not backup_file.exists():
                    return False

            return True

        except Exception:
            return False

    def _generate_checkpoint_id(self) -> str:
        """체크포인트 ID 생성

        Returns:
            고유 체크포인트 ID
        """
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        random_suffix = str(int(time.time() * 1000))[-6:]
        return f"ckpt_{timestamp}_{random_suffix}"

    def _discover_project_files(self) -> List[str]:
        """프로젝트 파일 자동 탐지

        Returns:
            프로젝트 파일 경로 목록
        """
        files = []
        important_patterns = [
            "src/**/*.py",
            "tests/**/*.py",
            "**/*.md",
            "**/*.json",
            "**/*.yml",
            "**/*.yaml",
            ".claude/**/*",
            ".moai/**/*"
        ]

        for pattern in important_patterns:
            for path in Path(".").glob(pattern):
                if path.is_file():
                    files.append(str(path))

        return list(set(files))  # 중복 제거

    def _collect_file_states(self, files: List[str]) -> Dict[str, str]:
        """파일 상태 수집

        Args:
            files: 파일 경로 목록

        Returns:
            {file_path: content_hash} 딕셔너리
        """
        file_states = {}

        for file_path in files:
            try:
                path = Path(file_path)
                if path.exists():
                    content = path.read_text(encoding="utf-8", errors="ignore")
                    # 간단한 해시 (실제로는 hashlib 사용 권장)
                    content_hash = str(hash(content))
                    file_states[file_path] = content_hash
            except Exception:
                continue

        return file_states

    def _save_checkpoint(self, checkpoint: Checkpoint) -> None:
        """체크포인트 저장

        Args:
            checkpoint: 저장할 체크포인트
        """
        checkpoint_file = self.checkpoints_dir / f"checkpoint_{checkpoint.id}.json"
        with open(checkpoint_file, 'w', encoding='utf-8') as f:
            json.dump(checkpoint.to_dict(), f, indent=2, ensure_ascii=False)

    def _backup_files(self, checkpoint_id: str, files: List[str]) -> None:
        """파일 백업 생성

        Args:
            checkpoint_id: 체크포인트 ID
            files: 백업할 파일 목록
        """
        backup_dir = self.checkpoints_dir / f"backup_{checkpoint_id}"
        backup_dir.mkdir(exist_ok=True)

        for file_path in files:
            try:
                path = Path(file_path)
                if path.exists():
                    # 파일명에 /를 _로 변경하여 유효한 파일명으로 변환
                    backup_name = file_path.replace('/', '_')
                    backup_file = backup_dir / backup_name
                    shutil.copy2(path, backup_file)
            except Exception:
                continue

    def _load_checkpoint(self, checkpoint_id: str) -> Optional[Checkpoint]:
        """체크포인트 로드

        Args:
            checkpoint_id: 로드할 체크포인트 ID

        Returns:
            Checkpoint 객체 또는 None
        """
        try:
            checkpoint_file = self.checkpoints_dir / f"checkpoint_{checkpoint_id}.json"
            if not checkpoint_file.exists():
                return None

            with open(checkpoint_file, 'r', encoding='utf-8') as f:
                data = json.load(f)

            return Checkpoint(
                id=data['id'],
                timestamp=datetime.fromisoformat(data['timestamp']),
                description=data['description'],
                file_states=data['file_states'],
                metadata=data.get('metadata', {})
            )

        except Exception:
            return None

    def _restore_files(self, checkpoint: Checkpoint) -> bool:
        """파일 복원

        Args:
            checkpoint: 복원할 체크포인트

        Returns:
            성공 여부
        """
        backup_dir = self.checkpoints_dir / f"backup_{checkpoint.id}"
        if not backup_dir.exists():
            return False

        success_count = 0
        total_files = len(checkpoint.file_states)

        for file_path in checkpoint.file_states.keys():
            try:
                backup_name = file_path.replace('/', '_')
                backup_file = backup_dir / backup_name
                target_file = Path(file_path)

                if backup_file.exists():
                    # 대상 디렉토리 생성
                    target_file.parent.mkdir(parents=True, exist_ok=True)
                    # 파일 복원
                    shutil.copy2(backup_file, target_file)
                    success_count += 1

            except Exception:
                continue

        return success_count == total_files

    def _create_rollback_backup(self, checkpoint_id: str) -> None:
        """롤백 전 백업 생성

        Args:
            checkpoint_id: 롤백할 체크포인트 ID
        """
        backup_id = self._generate_checkpoint_id()
        self.create_checkpoint(
            description=f"롤백 전 백업 (from {checkpoint_id})",
            metadata={"rollback_from": checkpoint_id}
        )

    def _cleanup_old_checkpoints(self) -> None:
        """오래된 체크포인트 정리"""
        checkpoints = self.list_checkpoints()

        if len(checkpoints) > self.config.max_checkpoints:
            # 가장 오래된 체크포인트 삭제
            old_checkpoints = checkpoints[self.config.max_checkpoints:]
            for checkpoint in old_checkpoints:
                self.delete_checkpoint(checkpoint['id'])

    def _find_stable_checkpoints(self) -> List[Dict[str, Any]]:
        """안정적인 체크포인트 찾기

        Returns:
            안정적인 체크포인트 목록
        """
        checkpoints = self.list_checkpoints()
        stable_checkpoints = []

        for checkpoint in checkpoints:
            # 무결성 검증 통과한 체크포인트만 선택
            if self.validate_checkpoint_integrity(checkpoint['id']):
                stable_checkpoints.append(checkpoint)

        return stable_checkpoints

    def _log_rollback(self, checkpoint_id: str, checkpoint: Checkpoint) -> None:
        """롤백 로그 기록

        Args:
            checkpoint_id: 롤백한 체크포인트 ID
            checkpoint: 체크포인트 정보
        """
        log_file = self.checkpoints_dir / "rollback.log"

        log_entry = {
            "timestamp": datetime.now().isoformat(),
            "checkpoint_id": checkpoint_id,
            "description": checkpoint.description,
            "file_count": len(checkpoint.file_states)
        }

        try:
            with open(log_file, 'a', encoding='utf-8') as f:
                f.write(json.dumps(log_entry, ensure_ascii=False) + '\n')
        except Exception:
            pass

    def get_rollback_history(self) -> List[Dict[str, Any]]:
        """롤백 이력 조회

        Returns:
            롤백 이력 목록
        """
        log_file = self.checkpoints_dir / "rollback.log"
        history = []

        if log_file.exists():
            try:
                with open(log_file, 'r', encoding='utf-8') as f:
                    for line in f:
                        if line.strip():
                            history.append(json.loads(line.strip()))
            except Exception:
                pass

        # 최신 이력부터 정렬
        history.sort(key=lambda x: x.get('timestamp', ''), reverse=True)
        return history
