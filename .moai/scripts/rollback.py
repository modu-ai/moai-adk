#!/usr/bin/env python3
"""
MoAI-ADK Git Rollback Script v0.3.0 (통합 시스템 기반)
체크포인트 기반 안전한 롤백 시스템 – 통합 체크포인트 시스템 사용

@REQ:GIT-ROLLBACK-001
@FEATURE:ROLLBACK-SYSTEM-001
@API:GET-ROLLBACK
@DESIGN:CHECKPOINT-ROLLBACK-003
@TECH:PERSONAL-MODE-ONLY-001
"""

import argparse
import sys
from pathlib import Path
from typing import Any

import click

# 새로운 통합 시스템 import
sys.path.append(str(Path(__file__).parent / "utils"))
from checkpoint_system import CheckpointError, CheckpointInfo, CheckpointSystem


class MoAIRollback:
    """롤백 관리 시스템 (통합 시스템 래퍼)"""

    def __init__(self):
        self.project_root = Path(__file__).resolve().parents[2]
        self.checkpoint_system = CheckpointSystem(self.project_root)

    def list_available_checkpoints(self, limit: int = 10) -> list[CheckpointInfo]:
        """사용 가능한 체크포인트 목록 조회"""
        return self.checkpoint_system.list_checkpoints(limit)

    def rollback_to_checkpoint(
        self, tag_or_index: str, confirm: bool = False
    ) -> dict[str, Any]:
        """체크포인트로 롤백"""
        try:
            # 확인 절차
            checkpoint = self.checkpoint_system.get_checkpoint_info(tag_or_index)
            if not checkpoint:
                return {
                    "success": False,
                    "error": f"체크포인트를 찾을 수 없습니다: {tag_or_index}",
                }

            if not confirm:
                click.echo("롤백할 체크포인트:")
                click.echo(f"  태그: {checkpoint.tag}")
                click.echo(f"  메시지: {checkpoint.message}")
                click.echo(f"  생성일: {checkpoint.created_at}")
                click.echo(f"  커밋: {checkpoint.commit_hash}")

                response = input("\n정말 롤백하시겠습니까? (y/N): ")
                if response.lower() != "y":
                    return {"success": False, "error": "사용자가 취소했습니다"}

            # 롤백 실행
            result_checkpoint = self.checkpoint_system.rollback_to_checkpoint(
                tag_or_index
            )

            return {
                "success": True,
                "tag": result_checkpoint.tag,
                "message": result_checkpoint.message,
                "commit_hash": result_checkpoint.commit_hash,
            }

        except CheckpointError as e:
            return {"success": False, "error": str(e)}

    def rollback_interactive(self) -> dict[str, Any]:
        """대화형 롤백"""
        checkpoints = self.list_available_checkpoints()

        if not checkpoints:
            return {"success": False, "error": "사용 가능한 체크포인트가 없습니다"}

        click.echo("\n사용 가능한 체크포인트:")
        click.echo("-" * 80)
        for i, cp in enumerate(checkpoints):
            auto_marker = "(자동)" if cp.is_auto else "(수동)"
            click.echo(f"{i:2d}: {cp.tag} {auto_marker}")
            click.echo(f"     메시지: {cp.message}")
            click.echo(f"     생성일: {cp.created_at}")
            click.echo(f"     커밋: {cp.commit_hash[:8]}")
            click.echo()

        try:
            choice = input("롤백할 체크포인트 번호를 입력하세요 (취소: q): ")
            if choice.lower() == "q":
                return {"success": False, "error": "사용자가 취소했습니다"}

            index = int(choice)
            if 0 <= index < len(checkpoints):
                return self.rollback_to_checkpoint(str(index), confirm=False)
            else:
                return {"success": False, "error": "유효하지 않은 번호입니다"}

        except ValueError:
            return {"success": False, "error": "올바른 번호를 입력하세요"}

    def show_rollback_preview(self, tag_or_index: str) -> dict[str, Any]:
        """롤백 미리보기"""
        try:
            checkpoint = self.checkpoint_system.get_checkpoint_info(tag_or_index)
            if not checkpoint:
                return {
                    "success": False,
                    "error": f"체크포인트를 찾을 수 없습니다: {tag_or_index}",
                }

            # 현재 상태와 롤백 대상 비교
            current_checkpoints = self.checkpoint_system.list_checkpoints(1)
            current = current_checkpoints[0] if current_checkpoints else None

            preview = {
                "success": True,
                "target_checkpoint": checkpoint.to_dict(),
                "current_checkpoint": current.to_dict() if current else None,
                "rollback_safe": True,  # 체크포인트 시스템은 항상 안전
            }

            return preview

        except Exception as e:
            return {"success": False, "error": str(e)}


def main():
    """CLI 엔트리포인트"""
    parser = argparse.ArgumentParser(description="MoAI 체크포인트 롤백 시스템")

    subparsers = parser.add_subparsers(dest="command", help="사용 가능한 명령어")

    # list 명령어
    list_parser = subparsers.add_parser("list", help="체크포인트 목록 조회")
    list_parser.add_argument(
        "--limit", type=int, default=10, help="표시할 체크포인트 수"
    )

    # rollback 명령어
    rollback_parser = subparsers.add_parser("rollback", help="체크포인트로 롤백")
    rollback_parser.add_argument("target", help="롤백할 체크포인트 (태그 또는 인덱스)")
    rollback_parser.add_argument(
        "--yes", "-y", action="store_true", help="확인 없이 실행"
    )

    # interactive 명령어
    subparsers.add_parser("interactive", help="대화형 롤백")

    # preview 명령어
    preview_parser = subparsers.add_parser("preview", help="롤백 미리보기")
    preview_parser.add_argument("target", help="미리볼 체크포인트 (태그 또는 인덱스)")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return

    rollback_system = MoAIRollback()

    try:
        if args.command == "list":
            checkpoints = rollback_system.list_available_checkpoints(args.limit)
            click.echo(f"\n체크포인트 목록 ({len(checkpoints)}개):")
            click.echo("-" * 80)
            for i, cp in enumerate(checkpoints):
                auto_marker = "(자동)" if cp.is_auto else "(수동)"
                click.echo(f"{i:2d}: {cp.tag} {auto_marker}")
                click.echo(f"     메시지: {cp.message}")
                click.echo(f"     생성일: {cp.created_at}")
                click.echo()

        elif args.command == "rollback":
            result = rollback_system.rollback_to_checkpoint(args.target, args.yes)
            if result["success"]:
                click.echo(f"✅ 롤백 완료: {result['tag']}")
                click.echo(f"   메시지: {result['message']}")
                click.echo(f"   커밋: {result['commit_hash'][:8]}")
            else:
                click.echo(f"❌ 롤백 실패: {result['error']}")

        elif args.command == "interactive":
            result = rollback_system.rollback_interactive()
            if result["success"]:
                click.echo(f"✅ 롤백 완료: {result['tag']}")
            else:
                click.echo(f"❌ 롤백 실패: {result['error']}")

        elif args.command == "preview":
            result = rollback_system.show_rollback_preview(args.target)
            if result["success"]:
                click.echo("📋 롤백 미리보기:")
                click.echo(f"   대상: {result['target_checkpoint']['tag']}")
                click.echo(f"   메시지: {result['target_checkpoint']['message']}")
                click.echo(f"   생성일: {result['target_checkpoint']['created_at']}")
                click.echo(
                    f"   안전성: {'안전' if result['rollback_safe'] else '주의 필요'}"
                )
            else:
                click.echo(f"❌ 미리보기 실패: {result['error']}")

    except Exception as e:
        click.echo(f"❌ 오류 발생: {e}")


if __name__ == "__main__":
    main()
