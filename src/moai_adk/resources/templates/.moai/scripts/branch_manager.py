#!/usr/bin/env python3
"""
MoAI 브랜치 관리자 v0.2.0 (통합 시스템 기반)
모드별 최적화된 스마트 브랜치 관리 시스템 – 통합 Git 워크플로우 사용

@REQ:GIT-BRANCH-001
@FEATURE:BRANCH-MANAGEMENT-001
@API:GET-BRANCH
@DESIGN:MODE-BASED-WORKFLOW-002
@TECH:GITFLOW-INTEGRATION-001
"""

import sys
import argparse
import json
from pathlib import Path
from typing import Any, Dict, List, Optional

import click

# 새로운 통합 시스템 import
sys.path.append(str(Path(__file__).parent / "utils"))
from git_workflow import GitWorkflow, GitWorkflowError
from project_helper import ProjectHelper


class BranchManager:
    """브랜치 관리자 (통합 시스템 래퍼)"""

    def __init__(self):
        self.project_root = Path(__file__).resolve().parents[2]
        self.git_workflow = GitWorkflow(self.project_root)
        self.config = ProjectHelper.load_config(self.project_root)
        self.mode = self.config.get("mode", "personal")

    def create_feature_branch(self, feature_name: str, from_branch: Optional[str] = None) -> Dict[str, Any]:
        """기능 브랜치 생성"""
        try:
            branch_name = self.git_workflow.create_feature_branch(feature_name, from_branch)
            return {
                "success": True,
                "branch_name": branch_name,
                "mode": self.mode,
                "base_branch": from_branch or self.git_workflow._get_default_branch()
            }
        except GitWorkflowError as e:
            return {"success": False, "error": str(e)}

    def create_hotfix_branch(self, fix_name: str) -> Dict[str, Any]:
        """핫픽스 브랜치 생성"""
        try:
            branch_name = self.git_workflow.create_hotfix_branch(fix_name)
            return {
                "success": True,
                "branch_name": branch_name,
                "mode": self.mode,
                "type": "hotfix"
            }
        except GitWorkflowError as e:
            return {"success": False, "error": str(e)}

    def get_branch_status(self) -> Dict[str, Any]:
        """브랜치 상태 조회"""
        try:
            status = self.git_workflow.get_branch_status()
            status["manager_mode"] = self.mode
            return status
        except Exception as e:
            return {"error": str(e)}

    def switch_branch(self, branch_name: str) -> Dict[str, Any]:
        """브랜치 전환"""
        try:
            # 변경사항이 있으면 체크포인트 생성 (개인 모드)
            if self.mode == "personal" and self.git_workflow.git.has_uncommitted_changes():
                self.git_workflow.checkpoint_system.create_checkpoint(
                    f"Pre-switch to {branch_name}", is_auto=True
                )

            self.git_workflow.git.switch_branch(branch_name)
            return {
                "success": True,
                "current_branch": branch_name,
                "mode": self.mode
            }
        except Exception as e:
            return {"success": False, "error": str(e)}

    def list_branches(self) -> Dict[str, Any]:
        """브랜치 목록 조회"""
        try:
            local_branches = self.git_workflow.git.get_local_branches()
            current_branch = self.git_workflow.git.get_current_branch()

            branches_info = []
            for branch in local_branches:
                is_current = branch == current_branch
                branch_info = {
                    "name": branch,
                    "is_current": is_current,
                    "type": self._classify_branch(branch)
                }
                branches_info.append(branch_info)

            return {
                "success": True,
                "branches": branches_info,
                "current_branch": current_branch,
                "total_count": len(local_branches)
            }
        except Exception as e:
            return {"success": False, "error": str(e)}

    def delete_branch(self, branch_name: str, force: bool = False) -> Dict[str, Any]:
        """브랜치 삭제"""
        try:
            current_branch = self.git_workflow.git.get_current_branch()
            if branch_name == current_branch:
                return {"success": False, "error": "현재 브랜치는 삭제할 수 없습니다"}

            self.git_workflow.git.delete_branch(branch_name, force)
            return {
                "success": True,
                "deleted_branch": branch_name,
                "force": force
            }
        except Exception as e:
            return {"success": False, "error": str(e)}

    def cleanup_merged_branches(self, dry_run: bool = True) -> Dict[str, Any]:
        """병합된 브랜치 정리"""
        try:
            merged_branches = self.git_workflow.cleanup_merged_branches(dry_run)
            return {
                "success": True,
                "merged_branches": merged_branches,
                "dry_run": dry_run,
                "count": len(merged_branches)
            }
        except Exception as e:
            return {"success": False, "error": str(e)}

    def sync_branch(self, push: bool = True) -> Dict[str, Any]:
        """브랜치 동기화"""
        try:
            success = self.git_workflow.sync_with_remote(push)
            current_branch = self.git_workflow.git.get_current_branch()

            return {
                "success": success,
                "branch": current_branch,
                "push": push,
                "mode": self.mode
            }
        except Exception as e:
            return {"success": False, "error": str(e)}

    def _classify_branch(self, branch_name: str) -> str:
        """브랜치 유형 분류"""
        if branch_name.startswith("feature/"):
            return "feature"
        elif branch_name.startswith("hotfix/"):
            return "hotfix"
        elif branch_name.startswith("release/"):
            return "release"
        elif branch_name in ["main", "master", "develop", "dev"]:
            return "main"
        else:
            return "other"


def main():
    """CLI 엔트리포인트"""
    parser = argparse.ArgumentParser(description="MoAI 브랜치 관리 시스템")

    subparsers = parser.add_subparsers(dest="command", help="사용 가능한 명령어")

    # create 명령어
    create_parser = subparsers.add_parser("create", help="브랜치 생성")
    create_parser.add_argument("type", choices=["feature", "hotfix"], help="브랜치 유형")
    create_parser.add_argument("name", help="브랜치/기능 이름")
    create_parser.add_argument("--from", dest="from_branch", help="기준 브랜치")

    # list 명령어
    subparsers.add_parser("list", help="브랜치 목록 조회")

    # switch 명령어
    switch_parser = subparsers.add_parser("switch", help="브랜치 전환")
    switch_parser.add_argument("branch", help="전환할 브랜치명")

    # delete 명령어
    delete_parser = subparsers.add_parser("delete", help="브랜치 삭제")
    delete_parser.add_argument("branch", help="삭제할 브랜치명")
    delete_parser.add_argument("--force", "-f", action="store_true", help="강제 삭제")

    # status 명령어
    subparsers.add_parser("status", help="브랜치 상태 조회")

    # cleanup 명령어
    cleanup_parser = subparsers.add_parser("cleanup", help="병합된 브랜치 정리")
    cleanup_parser.add_argument("--execute", action="store_true", help="실제 삭제 실행")

    # sync 명령어
    sync_parser = subparsers.add_parser("sync", help="브랜치 동기화")
    sync_parser.add_argument("--no-push", action="store_true", help="푸시 건너뛰기")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return

    manager = BranchManager()

    try:
        if args.command == "create":
            if args.type == "feature":
                result = manager.create_feature_branch(args.name, args.from_branch)
            elif args.type == "hotfix":
                result = manager.create_hotfix_branch(args.name)

            if result["success"]:
                click.echo(f"✅ {args.type} 브랜치 생성 완료: {result['branch_name']}")
                click.echo(f"   모드: {result['mode']}")
            else:
                click.echo(f"❌ 브랜치 생성 실패: {result['error']}")

        elif args.command == "list":
            result = manager.list_branches()
            if result["success"]:
                click.echo(f"\n브랜치 목록 ({result['total_count']}개):")
                click.echo("-" * 60)
                for branch in result["branches"]:
                    marker = "* " if branch["is_current"] else "  "
                    type_marker = f"[{branch['type']}]" if branch['type'] != 'other' else ""
                    click.echo(f"{marker}{branch['name']} {type_marker}")
                click.echo(f"\n현재 브랜치: {result['current_branch']}")
            else:
                click.echo(f"❌ 브랜치 목록 조회 실패: {result['error']}")

        elif args.command == "switch":
            result = manager.switch_branch(args.branch)
            if result["success"]:
                click.echo(f"✅ 브랜치 전환 완료: {result['current_branch']}")
                click.echo(f"   모드: {result['mode']}")
            else:
                click.echo(f"❌ 브랜치 전환 실패: {result['error']}")

        elif args.command == "delete":
            result = manager.delete_branch(args.branch, args.force)
            if result["success"]:
                force_marker = " (강제)" if result["force"] else ""
                click.echo(f"✅ 브랜치 삭제 완료: {result['deleted_branch']}{force_marker}")
            else:
                click.echo(f"❌ 브랜치 삭제 실패: {result['error']}")

        elif args.command == "status":
            result = manager.get_branch_status()
            if "error" not in result:
                click.echo(f"📋 브랜치 상태:")
                click.echo(f"   현재 브랜치: {result['current_branch']}")
                click.echo(f"   관리 모드: {result['manager_mode']}")
                click.echo(f"   변경사항: {'있음' if result['has_uncommitted_changes'] else '없음'}")
                click.echo(f"   원격 저장소: {'연결됨' if result['has_remote'] else '없음'}")
                click.echo(f"   작업 트리: {'깨끗함' if result['clean_working_tree'] else '수정됨'}")
            else:
                click.echo(f"❌ 상태 조회 실패: {result['error']}")

        elif args.command == "cleanup":
            dry_run = not args.execute
            result = manager.cleanup_merged_branches(dry_run)
            if result["success"]:
                if result["count"] > 0:
                    click.echo(f"{'🔍 발견된' if dry_run else '✅ 정리된'} 병합 브랜치 ({result['count']}개):")
                    for branch in result["merged_branches"]:
                        click.echo(f"  - {branch}")
                    if dry_run:
                        click.echo("\n실제 삭제하려면 --execute 옵션을 사용하세요")
                else:
                    click.echo("🎉 정리할 병합 브랜치가 없습니다")
            else:
                click.echo(f"❌ 브랜치 정리 실패: {result['error']}")

        elif args.command == "sync":
            push = not args.no_push
            result = manager.sync_branch(push)
            if result["success"]:
                sync_type = "푸시 포함 동기화" if push else "풀만 실행"
                click.echo(f"✅ {sync_type} 완료: {result['branch']}")
                click.echo(f"   모드: {result['mode']}")
            else:
                click.echo(f"❌ 동기화 실패: {result['error']}")

    except Exception as e:
        click.echo(f"❌ 오류 발생: {e}")


if __name__ == "__main__":
    main()