#!/usr/bin/env python3
"""
MoAI 체크포인트 관리자 v0.4.0 (통합 시스템 기반)
개인 모드 전용 안전한 실험 환경 제공 – 통합 체크포인트 시스템 사용

@REQ:GIT-CHECKPOINT-001
@FEATURE:CHECKPOINT-SYSTEM-001
@API:GET-CHECKPOINT
@DESIGN:CHECKPOINT-WORKFLOW-004
@TECH:PERSONAL-MODE-001
"""

from __future__ import annotations

import sys
from pathlib import Path
from typing import Any

import click

# 새로운 통합 시스템 import
sys.path.append(str(Path(__file__).parent / "utils"))
from checkpoint_system import CheckpointError, CheckpointSystem


class CheckpointManager:
    """체크포인트 관리자 (통합 시스템 래퍼)"""

    def __init__(self) -> None:
        self.project_root = Path(__file__).resolve().parents[2]
        self.checkpoint_system = CheckpointSystem(self.project_root)

    def create_checkpoint(self, message: str = "Manual checkpoint", is_auto: bool = False) -> dict[str, Any]:
        """체크포인트 생성"""
        try:
            checkpoint = self.checkpoint_system.create_checkpoint(message, is_auto)
            return {
                "success": True,
                "tag": checkpoint.tag,
                "commit_hash": checkpoint.commit_hash,
                "message": checkpoint.message,
                "created_at": checkpoint.created_at,
                "file_count": checkpoint.file_count
            }
        except CheckpointError as e:
            return {"success": False, "error": str(e)}

    def list_checkpoints(self, limit: int | None = None) -> list[dict[str, Any]]:
        """체크포인트 목록 조회"""
        try:
            checkpoints = self.checkpoint_system.list_checkpoints(limit)
            return [cp.to_dict() for cp in checkpoints]
        except Exception as e:
            return [{"error": str(e)}]

    def rollback_to_checkpoint(self, tag_or_index: str) -> dict[str, Any]:
        """체크포인트로 롤백"""
        try:
            checkpoint = self.checkpoint_system.rollback_to_checkpoint(tag_or_index)
            return {
                "success": True,
                "tag": checkpoint.tag,
                "commit_hash": checkpoint.commit_hash,
                "message": checkpoint.message
            }
        except CheckpointError as e:
            return {"success": False, "error": str(e)}

    def delete_checkpoint(self, tag_or_index: str) -> dict[str, Any]:
        """체크포인트 삭제"""
        try:
            success = self.checkpoint_system.delete_checkpoint(tag_or_index)
            return {"success": success}
        except Exception as e:
            return {"success": False, "error": str(e)}

    def auto_checkpoint_if_needed(self, message: str = "Auto checkpoint") -> dict[str, Any] | None:
        """필요 시 자동 체크포인트 생성"""
        try:
            if self.checkpoint_system.should_create_auto_checkpoint():
                return self.create_checkpoint(message, is_auto=True)
            return None
        except Exception as e:
            return {"success": False, "error": str(e)}

    def get_checkpoint_info(self, tag_or_index: str) -> dict[str, Any] | None:
        """체크포인트 정보 조회"""
        try:
            checkpoint = self.checkpoint_system.get_checkpoint_info(tag_or_index)
            return checkpoint.to_dict() if checkpoint else None
        except Exception as e:
            return {"error": str(e)}

    # 하위 호환성을 위한 레거시 메서드들
    def create_auto_checkpoint(self, message: str = "Auto checkpoint") -> dict[str, Any]:
        """자동 체크포인트 생성 (레거시 호환)"""
        return self.create_checkpoint(message, is_auto=True)

    def get_latest_checkpoint(self) -> dict[str, Any] | None:
        """최신 체크포인트 조회 (레거시 호환)"""
        checkpoints = self.list_checkpoints(limit=1)
        return checkpoints[0] if checkpoints and "error" not in checkpoints[0] else None


def main():
    """CLI 엔트리포인트"""
    manager = CheckpointManager()

    if len(sys.argv) < 2:
        click.echo("사용법: python checkpoint_manager.py <command> [args]")
        click.echo("명령어:")
        click.echo("  create <message>     - 체크포인트 생성")
        click.echo("  list [limit]         - 체크포인트 목록")
        click.echo("  rollback <tag>       - 체크포인트로 롤백")
        click.echo("  delete <tag>         - 체크포인트 삭제")
        click.echo("  info <tag>           - 체크포인트 정보")
        return

    command = sys.argv[1]

    try:
        if command == "create":
            message = sys.argv[2] if len(sys.argv) > 2 else "Manual checkpoint"
            result = manager.create_checkpoint(message)
            click.echo(f"체크포인트 생성: {result}")

        elif command == "list":
            limit = int(sys.argv[2]) if len(sys.argv) > 2 else None
            checkpoints = manager.list_checkpoints(limit)
            click.echo(f"체크포인트 목록 ({len(checkpoints)}개):")
            for i, cp in enumerate(checkpoints):
                if "error" not in cp:
                    click.echo(f"  {i}: {cp['tag']} - {cp['message']}")

        elif command == "rollback":
            if len(sys.argv) < 3:
                click.echo("오류: 롤백할 체크포인트를 지정하세요")
                return
            tag = sys.argv[2]
            result = manager.rollback_to_checkpoint(tag)
            click.echo(f"롤백 결과: {result}")

        elif command == "delete":
            if len(sys.argv) < 3:
                click.echo("오류: 삭제할 체크포인트를 지정하세요")
                return
            tag = sys.argv[2]
            result = manager.delete_checkpoint(tag)
            click.echo(f"삭제 결과: {result}")

        elif command == "info":
            if len(sys.argv) < 3:
                click.echo("오류: 조회할 체크포인트를 지정하세요")
                return
            tag = sys.argv[2]
            info = manager.get_checkpoint_info(tag)
            click.echo(f"체크포인트 정보: {info}")

        else:
            click.echo(f"알 수 없는 명령어: {command}")

    except Exception as e:
        click.echo(f"오류 발생: {e}")


if __name__ == "__main__":
    main()
