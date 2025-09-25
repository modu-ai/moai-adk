#!/usr/bin/env python3
"""
MoAI 동기화 관리자 v0.1.0
모드별 최적화된 안전한 원격 저장소 동기화

@REQ:GIT-SYNC-001
@FEATURE:REMOTE-SYNC-001
@API:SYNC-INTERFACE-001
@DESIGN:MODE-BASED-SYNC-001
@TECH:PUSH-PULL-STRATEGY-001
"""

import sys
import json
import subprocess
from pathlib import Path

import click

class SyncManager:
    """동기화 관리자

    @FEATURE:REMOTE-SYNC-001
    @API:SYNC-INTERFACE-001
    """

    def __init__(self):
        self.project_root = Path(__file__).resolve().parents[2]
        self.config_path = self.project_root / ".moai" / "config.json"

    def load_config(self) -> dict:
        """프로젝트 설정 로드

        @DATA:CONFIG-LOAD-001
        @API:CONFIG-ACCESS-001
        """
        try:
            with open(self.config_path, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception:
            return {"project": {"mode": "personal"}}

    def get_sync_status(self):
        """동기화 상태 확인

        @API:SYNC-STATUS-001
        @DATA:REMOTE-STATUS-001
        """
        click.echo("=== 동기화 상태 확인 ===")

        try:
            # 현재 브랜치
            current_branch = subprocess.run(
                ["git", "branch", "--show-current"],
                capture_output=True, text=True, check=True, cwd=self.project_root
            ).stdout.strip()

            click.echo(f"📍 현재 브랜치: {current_branch}")

            # 원격 상태 업데이트
            subprocess.run(["git", "fetch"], capture_output=True, check=True, cwd=self.project_root)

            # Push 필요한 커밋
            try:
                ahead_result = subprocess.run(
                    ["git", "log", f"origin/{current_branch}..HEAD", "--oneline"],
                    capture_output=True, text=True, cwd=self.project_root
                )
                ahead_count = len(ahead_result.stdout.strip().split('\n')) if ahead_result.stdout.strip() else 0
                click.echo(f"📤 Push 필요: {ahead_count}개 커밋")
            except:
                click.echo("📤 Push 필요: 확인 불가 (원격 브랜치 없음)")

            # Pull 필요한 커밋
            try:
                behind_result = subprocess.run(
                    ["git", "log", f"HEAD..origin/{current_branch}", "--oneline"],
                    capture_output=True, text=True, cwd=self.project_root
                )
                behind_count = len(behind_result.stdout.strip().split('\n')) if behind_result.stdout.strip() else 0
                click.echo(f"📥 Pull 필요: {behind_count}개 커밋")
            except:
                click.echo("📥 Pull 필요: 확인 불가 (원격 브랜치 없음)")

        except subprocess.CalledProcessError as e:
            click.echo(f"❌ 상태 확인 실패: {e}")

    def push_changes(self):
        """원격으로 Push

        @API:GIT-PUSH-001
        @DESIGN:PUSH-STRATEGY-001
        """
        click.echo("📤 원격으로 Push 중...")

        config = self.load_config()
        mode = config.get("project", {}).get("mode", "personal")

        try:
            current_branch = subprocess.run(
                ["git", "branch", "--show-current"],
                capture_output=True, text=True, check=True, cwd=self.project_root
            ).stdout.strip()

            if mode == "team":
                # 팀 모드: 안전한 push
                subprocess.run(["git", "push", "origin", current_branch], check=True, cwd=self.project_root)
                click.echo("✅ 팀 모드 Push 완료")
            else:
                # 개인 모드: 선택적 push
                click.echo("🎯 개인 모드: 백업이 필요한 경우만 Push")
                subprocess.run(["git", "push", "origin", current_branch], check=True, cwd=self.project_root)
                click.echo("✅ 개인 모드 Push 완료")

        except subprocess.CalledProcessError as e:
            click.echo(f"❌ Push 실패: {e}")

    def pull_changes(self):
        """원격에서 Pull

        @API:GIT-PULL-001
        @DESIGN:PULL-STRATEGY-001
        """
        click.echo("📥 원격에서 Pull 중...")

        try:
            # 변경사항이 있으면 stash
            status_result = subprocess.run(["git", "status", "--porcelain"], capture_output=True, text=True, cwd=self.project_root)
            if status_result.stdout.strip():
                click.echo("💾 현재 변경사항을 stash에 저장")
                subprocess.run(["git", "stash", "push", "-m", "auto-stash-before-pull"], check=True, cwd=self.project_root)
                need_stash_pop = True
            else:
                need_stash_pop = False

            # Pull 실행
            subprocess.run(["git", "pull"], check=True, cwd=self.project_root)
            click.echo("✅ Pull 완료")

            # stash 복원
            if need_stash_pop:
                subprocess.run(["git", "stash", "pop"], check=True, cwd=self.project_root)
                click.echo("✅ 변경사항 복원 완료")

        except subprocess.CalledProcessError as e:
            click.echo(f"❌ Pull 실패: {e}")

    def auto_sync(self):
        """자동 동기화

        @FEATURE:AUTO-SYNC-001
        @DESIGN:MODE-BASED-SYNC-001
        """
        click.echo("🔄 자동 동기화 시작")

        config = self.load_config()
        mode = config.get("project", {}).get("mode", "personal")

        click.echo(f"🎯 모드: {mode}")

        if mode == "team":
            # 팀 모드: Pull → Push 순서
            click.echo("👥 팀 모드: 원격 우선 동기화")
            self.pull_changes()
            self.push_changes()
        else:
            # 개인 모드: Push → Pull 순서 (선택적)
            click.echo("👤 개인 모드: 로컬 우선 동기화")
            self.push_changes()
            # 개인 모드에서는 pull 선택적

        click.echo("✅ 자동 동기화 완료")

    def run(self, args: list):
        """메인 실행 함수

        @API:SYNC-CLI-001
        @DESIGN:COMMAND-DISPATCH-001
        """
        if not args:
            self.get_sync_status()
            return

        action = args[0]

        if action == "--status":
            self.get_sync_status()
        elif action == "push":
            self.push_changes()
        elif action == "pull":
            self.pull_changes()
        elif action == "--auto":
            self.auto_sync()
        elif action == "--safe":
            # 안전 모드: 상태 확인 후 진행
            self.get_sync_status()
            click.echo("\n🔒 안전 모드: 위 상태를 확인 후 동기화를 진행하시겠습니까?")
            user_input = input("계속하려면 'y' 입력: ")
            if user_input.lower() == 'y':
                self.auto_sync()
            else:
                click.echo("동기화 취소됨")
        else:
            click.echo("❌ 알 수 없는 동기화 명령어")
            click.echo("사용법: python3 sync_manager.py [--status|push|pull|--auto|--safe]")

def main():
    """진입점

    @API:MAIN-ENTRY-001
    @TECH:CLI-INTERFACE-001
    """
    manager = SyncManager()
    manager.run(sys.argv[1:])

if __name__ == "__main__":
    main()
