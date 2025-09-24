#!/usr/bin/env python3
"""
MoAI Git Lock Manager - 에이전트 간 Git 작업 동기화

동시에 실행되는 여러 MoAI 에이전트가 Git 작업을 할 때 충돌을 방지하는
잠금 메커니즘을 제공합니다.

@TASK:GIT-LOCK-001
"""

import json
import os
import time
import fcntl
from pathlib import Path
from datetime import datetime, timedelta
import logging

class GitLockManager:
    """Git 작업 잠금 관리자"""

    def __init__(self, project_root: str = None):
        self.project_root = Path(project_root or os.getcwd())
        self.lock_dir = self.project_root / ".moai" / "locks"
        self.lock_dir.mkdir(parents=True, exist_ok=True)

        self.lock_file = self.lock_dir / "git_operations.lock"
        self.metadata_file = self.lock_dir / "git_lock_metadata.json"

        # 설정
        self.lock_timeout = 300  # 5분 타임아웃
        self.check_interval = 0.5  # 0.5초마다 잠금 확인

        # 로깅 설정
        logging.basicConfig(level=logging.INFO)
        self.logger = logging.getLogger(__name__)

    def acquire_git_lock(self, agent_name: str, operation: str,
                        description: str = "") -> bool:
        """
        Git 작업 잠금 획득

        Args:
            agent_name: 에이전트 이름 (spec-builder, code-builder, git-manager)
            operation: 작업 유형 (branch, commit, sync, checkpoint)
            description: 작업 설명

        Returns:
            bool: 잠금 획득 성공 여부
        """
        lock_info = {
            "agent": agent_name,
            "operation": operation,
            "description": description,
            "timestamp": datetime.now().isoformat(),
            "pid": os.getpid()
        }

        start_time = time.time()

        while time.time() - start_time < self.lock_timeout:
            try:
                # 잠금 파일 생성 시도
                with open(self.lock_file, 'w') as f:
                    # 파일 잠금 (non-blocking)
                    fcntl.flock(f.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)

                    # 잠금 메타데이터 저장
                    json.dump(lock_info, f, indent=2)
                    f.flush()

                    # 메타데이터 파일 업데이트
                    self._update_lock_metadata(lock_info)

                    self.logger.info(f"🔒 Git 잠금 획득: {agent_name} - {operation}")
                    return True

            except (OSError, IOError) as e:
                # 잠금이 이미 존재하는 경우
                current_lock = self._get_current_lock_info()
                if current_lock:
                    remaining = self._get_remaining_time(current_lock)
                    self.logger.info(
                        f"⏳ Git 잠금 대기 중... "
                        f"현재: {current_lock['agent']} ({remaining}초 남음)"
                    )

                # 잠시 대기 후 재시도
                time.sleep(self.check_interval)

                # 데드락 방지: 오래된 잠금 정리
                self._cleanup_stale_locks()

        self.logger.error(f"❌ Git 잠금 획득 실패: {agent_name} - {operation} (타임아웃)")
        return False

    def release_git_lock(self, agent_name: str) -> bool:
        """
        Git 작업 잠금 해제

        Args:
            agent_name: 에이전트 이름

        Returns:
            bool: 잠금 해제 성공 여부
        """
        try:
            if self.lock_file.exists():
                current_lock = self._get_current_lock_info()

                if current_lock and current_lock.get("agent") == agent_name:
                    # 잠금 해제
                    self.lock_file.unlink()

                    # 잠금 해제 기록
                    release_info = {
                        "agent": agent_name,
                        "released_at": datetime.now().isoformat(),
                        "duration_seconds": self._calculate_lock_duration(current_lock)
                    }
                    self._record_lock_release(release_info)

                    self.logger.info(f"🔓 Git 잠금 해제: {agent_name}")
                    return True
                else:
                    self.logger.warning(f"⚠️ 잠금 해제 실패: {agent_name} (소유자가 아님)")
                    return False
            else:
                self.logger.info(f"ℹ️ 해제할 잠금이 없음: {agent_name}")
                return True

        except Exception as e:
            self.logger.error(f"❌ 잠금 해제 오류: {e}")
            return False

    def is_git_locked(self) -> bool:
        """Git 작업이 잠금 중인지 확인"""
        return self.lock_file.exists() and self._is_lock_valid()

    def get_lock_status(self) -> dict:
        """현재 잠금 상태 조회"""
        if not self.is_git_locked():
            return {"locked": False}

        lock_info = self._get_current_lock_info()
        if not lock_info:
            return {"locked": False}

        return {
            "locked": True,
            "agent": lock_info.get("agent"),
            "operation": lock_info.get("operation"),
            "description": lock_info.get("description"),
            "started_at": lock_info.get("timestamp"),
            "remaining_seconds": self._get_remaining_time(lock_info)
        }

    def force_release_lock(self, reason: str = "강제 해제") -> bool:
        """강제로 잠금 해제 (응급 상황용)"""
        try:
            if self.lock_file.exists():
                current_lock = self._get_current_lock_info()

                # 강제 해제 기록
                force_release_info = {
                    "forced_release": True,
                    "reason": reason,
                    "original_lock": current_lock,
                    "released_at": datetime.now().isoformat()
                }
                self._record_lock_release(force_release_info)

                self.lock_file.unlink()
                self.logger.warning(f"🚨 Git 잠금 강제 해제: {reason}")
                return True
        except Exception as e:
            self.logger.error(f"❌ 강제 해제 실패: {e}")
            return False

        return False

    def _get_current_lock_info(self) -> dict:
        """현재 잠금 정보 조회"""
        try:
            if self.lock_file.exists():
                with open(self.lock_file, 'r') as f:
                    return json.load(f)
        except Exception as e:
            self.logger.error(f"잠금 정보 읽기 실패: {e}")
        return {}

    def _is_lock_valid(self) -> bool:
        """잠금이 유효한지 확인 (타임아웃 체크)"""
        lock_info = self._get_current_lock_info()
        if not lock_info:
            return False

        try:
            lock_time = datetime.fromisoformat(lock_info["timestamp"])
            elapsed = (datetime.now() - lock_time).total_seconds()
            return elapsed < self.lock_timeout
        except Exception:
            return False

    def _get_remaining_time(self, lock_info: dict) -> int:
        """잠금 남은 시간 계산"""
        try:
            lock_time = datetime.fromisoformat(lock_info["timestamp"])
            elapsed = (datetime.now() - lock_time).total_seconds()
            remaining = max(0, self.lock_timeout - elapsed)
            return int(remaining)
        except Exception:
            return 0

    def _cleanup_stale_locks(self):
        """오래된 잠금 정리"""
        if self.lock_file.exists() and not self._is_lock_valid():
            self.logger.warning("🧹 오래된 잠금 파일 정리")
            try:
                self.lock_file.unlink()
            except Exception as e:
                self.logger.error(f"잠금 파일 정리 실패: {e}")

    def _update_lock_metadata(self, lock_info: dict):
        """잠금 메타데이터 업데이트"""
        metadata = {
            "current_lock": lock_info,
            "lock_history": self._get_lock_history()
        }

        try:
            with open(self.metadata_file, 'w') as f:
                json.dump(metadata, f, indent=2)
        except Exception as e:
            self.logger.error(f"메타데이터 업데이트 실패: {e}")

    def _get_lock_history(self) -> list:
        """잠금 이력 조회"""
        try:
            if self.metadata_file.exists():
                with open(self.metadata_file, 'r') as f:
                    metadata = json.load(f)
                    return metadata.get("lock_history", [])
        except Exception:
            pass
        return []

    def _record_lock_release(self, release_info: dict):
        """잠금 해제 기록"""
        try:
            history = self._get_lock_history()
            history.append(release_info)

            # 최근 100개만 유지
            if len(history) > 100:
                history = history[-100:]

            metadata = {"lock_history": history}
            with open(self.metadata_file, 'w') as f:
                json.dump(metadata, f, indent=2)

        except Exception as e:
            self.logger.error(f"해제 기록 실패: {e}")

    def _calculate_lock_duration(self, lock_info: dict) -> float:
        """잠금 지속 시간 계산"""
        try:
            lock_time = datetime.fromisoformat(lock_info["timestamp"])
            duration = (datetime.now() - lock_time).total_seconds()
            return round(duration, 2)
        except Exception:
            return 0.0

# CLI 인터페이스
if __name__ == "__main__":
    import sys

    if len(sys.argv) < 2:
        print("사용법: git_lock_manager.py <command> [args...]")
        print("명령어: acquire, release, status, force-release")
        sys.exit(1)

    manager = GitLockManager()
    command = sys.argv[1]

    if command == "acquire":
        if len(sys.argv) < 4:
            print("사용법: acquire <agent_name> <operation> [description]")
            sys.exit(1)

        agent_name = sys.argv[2]
        operation = sys.argv[3]
        description = sys.argv[4] if len(sys.argv) > 4 else ""

        success = manager.acquire_git_lock(agent_name, operation, description)
        sys.exit(0 if success else 1)

    elif command == "release":
        if len(sys.argv) < 3:
            print("사용법: release <agent_name>")
            sys.exit(1)

        agent_name = sys.argv[2]
        success = manager.release_git_lock(agent_name)
        sys.exit(0 if success else 1)

    elif command == "status":
        status = manager.get_lock_status()
        print(json.dumps(status, indent=2))

    elif command == "force-release":
        reason = sys.argv[2] if len(sys.argv) > 2 else "강제 해제"
        success = manager.force_release_lock(reason)
        sys.exit(0 if success else 1)

    else:
        print(f"알 수 없는 명령어: {command}")
        sys.exit(1)