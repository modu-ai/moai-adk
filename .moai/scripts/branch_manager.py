#!/usr/bin/env python3
"""
MoAI 브랜치 관리자 v0.1.0
모드별 최적화된 스마트 브랜치 관리 시스템

@REQ:GIT-BRANCH-001
@FEATURE:BRANCH-MANAGEMENT-001
@API:BRANCH-INTERFACE-001
@DESIGN:MODE-BASED-WORKFLOW-001
@TECH:GITFLOW-INTEGRATION-001
"""

import sys
import json
import subprocess
import re
from pathlib import Path
from datetime import datetime

class BranchManager:
    """브랜치 관리자

    @FEATURE:BRANCH-MANAGEMENT-001
    @API:BRANCH-INTERFACE-001
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

    def get_current_branch(self) -> str:
        """현재 브랜치 이름 가져오기

        @API:GIT-BRANCH-STATUS-001
        @DATA:BRANCH-INFO-001
        """
        try:
            result = subprocess.run(
                ["git", "branch", "--show-current"],
                capture_output=True, text=True, check=True
            )
            return result.stdout.strip()
        except Exception:
            return "unknown"

    def list_branches(self):
        """브랜치 목록 표시

        @API:BRANCH-LIST-001
        @DATA:BRANCH-DISPLAY-001
        """
        print("=== 브랜치 목록 ===")

        try:
            # 로컬 브랜치
            result = subprocess.run(["git", "branch"], capture_output=True, text=True, check=True)
            print("📋 로컬 브랜치:")
            for line in result.stdout.split('\n'):
                if line.strip():
                    current_marker = "👉 " if line.startswith('*') else "   "
                    branch_name = line.strip().replace('*', '').strip()
                    print(f"{current_marker}{branch_name}")

            # 원격 브랜치
            result = subprocess.run(["git", "branch", "-r"], capture_output=True, text=True)
            if result.stdout.strip():
                print("\n🌐 원격 브랜치:")
                for line in result.stdout.split('\n')[:5]:  # 상위 5개만 표시
                    if line.strip() and not 'HEAD' in line:
                        print(f"   {line.strip()}")

        except subprocess.CalledProcessError:
            print("❌ 브랜치 목록을 가져올 수 없습니다.")

    def generate_branch_name(self, description: str, action_type: str = "feature") -> str:
        """모드별 브랜치명 생성

        @DESIGN:BRANCH-NAMING-001
        @FEATURE:NAME-GENERATION-001
        """
        config = self.load_config()
        mode = config.get("project", {}).get("mode", "personal")

        # 설명을 브랜치명에 적합하게 변환
        safe_description = re.sub(r'[^a-zA-Z0-9가-힣\s-]', '', description)
        safe_description = re.sub(r'\s+', '-', safe_description.strip())
        safe_description = safe_description.lower()

        if mode == "team":
            # 팀 모드: SPEC ID 연동
            # TODO: 실제 SPEC ID 추출 로직 구현
            spec_id = "SPEC-001"  # 임시값
            return f"{action_type}/{spec_id}-{safe_description}"
        else:
            # 개인 모드: 간단한 형식
            if action_type == "experiment":
                date_str = datetime.now().strftime("%m%d")
                return f"experiment/{date_str}-{safe_description}"
            else:
                return f"{action_type}/{safe_description}"

    def create_branch(self, description: str):
        """새 브랜치 생성

        @FEATURE:BRANCH-CREATE-001
        @API:GIT-CHECKOUT-001
        """
        if not description:
            print("❌ 브랜치 설명이 필요합니다.")
            print("사용법: /moai:git:branch create \"새로운 기능\"")
            return

        config = self.load_config()
        mode = config.get("project", {}).get("mode", "personal")

        # 브랜치명 생성
        branch_name = self.generate_branch_name(description)

        print(f"🌿 새 브랜치 생성 (모드: {mode})")
        print(f"📝 설명: {description}")
        print(f"🏷️ 브랜치명: {branch_name}")

        try:
            # 변경사항 확인
            result = subprocess.run(["git", "status", "--porcelain"], capture_output=True, text=True)
            if result.stdout.strip():
                print("⚠️ 변경사항이 있습니다. 스테이징 후 진행합니다.")
                subprocess.run(["git", "add", "-A"], check=True)

            # 브랜치 생성 및 전환
            subprocess.run(["git", "checkout", "-b", branch_name], check=True)
            print(f"✅ 브랜치 '{branch_name}' 생성 및 전환 완료")

            # 팀 모드에서 원격 연결 설정
            if mode == "team":
                try:
                    subprocess.run(["git", "push", "-u", "origin", branch_name],
                                 capture_output=True, check=True)
                    print("🌐 원격 브랜치 연결 완료")
                except subprocess.CalledProcessError:
                    print("⚠️ 원격 브랜치 연결 실패 (나중에 push 필요)")

        except subprocess.CalledProcessError as e:
            print(f"❌ 브랜치 생성 실패: {e}")

    def switch_branch(self, branch_name: str):
        """브랜치 전환

        @API:BRANCH-SWITCH-001
        @FEATURE:STASH-MANAGEMENT-001
        """
        if not branch_name:
            print("❌ 전환할 브랜치명이 필요합니다.")
            return

        print(f"🔄 브랜치 전환: {branch_name}")

        try:
            # 변경사항 확인
            result = subprocess.run(["git", "status", "--porcelain"], capture_output=True, text=True)
            if result.stdout.strip():
                print("💾 변경사항을 stash에 저장합니다.")
                subprocess.run(["git", "stash", "push", "-m", f"auto-stash-before-switch-{datetime.now().strftime('%H%M%S')}"], check=True)

            # 브랜치 전환
            subprocess.run(["git", "checkout", branch_name], check=True)
            print(f"✅ '{branch_name}' 브랜치로 전환 완료")

            # stash 복원 여부 확인
            result = subprocess.run(["git", "stash", "list"], capture_output=True, text=True)
            if result.stdout.strip():
                recent_stash = result.stdout.split('\n')[0]
                if "auto-stash-before-switch" in recent_stash:
                    user_input = input("💾 이전 변경사항을 복원하시겠습니까? (y/N): ")
                    if user_input.lower() == 'y':
                        subprocess.run(["git", "stash", "pop"], check=True)
                        print("✅ 변경사항 복원 완료")

        except subprocess.CalledProcessError as e:
            print(f"❌ 브랜치 전환 실패: {e}")

    def delete_branch(self, branch_name: str):
        """브랜치 삭제

        @API:BRANCH-DELETE-001
        @FEATURE:BRANCH-CLEANUP-001
        """
        if not branch_name:
            print("❌ 삭제할 브랜치명이 필요합니다.")
            return

        current_branch = self.get_current_branch()
        if branch_name == current_branch:
            print("❌ 현재 브랜치는 삭제할 수 없습니다. 다른 브랜치로 전환 후 시도하세요.")
            return

        print(f"🗑️ 브랜치 삭제: {branch_name}")

        try:
            # 로컬 브랜치 삭제
            subprocess.run(["git", "branch", "-d", branch_name], check=True)
            print(f"✅ 로컬 브랜치 '{branch_name}' 삭제 완료")

            # 원격 브랜치 삭제 (선택적)
            try:
                subprocess.run(["git", "push", "origin", "--delete", branch_name],
                             capture_output=True, check=True)
                print("🌐 원격 브랜치도 삭제 완료")
            except subprocess.CalledProcessError:
                print("⚠️ 원격 브랜치 삭제 실패 (존재하지 않거나 권한 없음)")

        except subprocess.CalledProcessError as e:
            print(f"❌ 브랜치 삭제 실패: {e}")
            print("💡 강제 삭제가 필요하면 'git branch -D {branch_name}' 사용")

    def clean_branches(self):
        """정리 작업

        @FEATURE:AUTO-CLEANUP-001
        @DESIGN:BRANCH-LIFECYCLE-001
        """
        print("🧹 브랜치 정리 작업")

        config = self.load_config()
        mode = config.get("project", {}).get("mode", "personal")

        try:
            # 병합된 브랜치 찾기
            result = subprocess.run(
                ["git", "branch", "--merged", "HEAD"],
                capture_output=True, text=True, check=True
            )

            merged_branches = []
            current_branch = self.get_current_branch()

            for line in result.stdout.split('\n'):
                branch_name = line.strip().replace('*', '').strip()
                if (branch_name and
                    branch_name != current_branch and
                    branch_name not in ['main', 'master', 'develop'] and
                    not branch_name.startswith('checkpoint_')):
                    merged_branches.append(branch_name)

            if merged_branches:
                print("병합된 브랜치들:")
                for branch in merged_branches:
                    print(f"  🔗 {branch}")

                if mode == "personal":
                    # 개인 모드: 자동 정리
                    for branch in merged_branches:
                        subprocess.run(["git", "branch", "-d", branch], check=True)
                        print(f"🗑️ 삭제: {branch}")
                    print(f"✅ {len(merged_branches)}개 브랜치 정리 완료")
                else:
                    # 팀 모드: 확인 후 정리
                    user_input = input(f"{len(merged_branches)}개 브랜치를 삭제하시겠습니까? (y/N): ")
                    if user_input.lower() == 'y':
                        for branch in merged_branches:
                            subprocess.run(["git", "branch", "-d", branch], check=True)
                            print(f"🗑️ 삭제: {branch}")
                        print(f"✅ {len(merged_branches)}개 브랜치 정리 완료")
            else:
                print("정리할 브랜치가 없습니다.")

        except subprocess.CalledProcessError as e:
            print(f"❌ 브랜치 정리 실패: {e}")

    def show_status(self):
        """브랜치 시스템 상태

        @API:STATUS-DISPLAY-001
        @DATA:BRANCH-STATS-001
        """
        print("=== 브랜치 시스템 상태 ===")

        config = self.load_config()
        mode = config.get("project", {}).get("mode", "personal")
        current_branch = self.get_current_branch()

        print(f"🎯 모드: {mode}")
        print(f"📍 현재 브랜치: {current_branch}")

        # 브랜치 통계
        try:
            local_result = subprocess.run(["git", "branch"], capture_output=True, text=True)
            local_count = len([line for line in local_result.stdout.split('\n') if line.strip()])

            remote_result = subprocess.run(["git", "branch", "-r"], capture_output=True, text=True)
            remote_count = len([line for line in remote_result.stdout.split('\n') if line.strip() and 'HEAD' not in line])

            print(f"📋 로컬 브랜치: {local_count}개")
            print(f"🌐 원격 브랜치: {remote_count}개")

            # 체크포인트 브랜치
            checkpoint_result = subprocess.run(["git", "branch", "--list", "checkpoint_*"], capture_output=True, text=True)
            checkpoint_count = len([line for line in checkpoint_result.stdout.split('\n') if line.strip()])
            print(f"💾 체크포인트: {checkpoint_count}개")

        except Exception:
            print("브랜치 통계를 가져올 수 없습니다.")

    def run(self, args: list):
        """메인 실행 함수

        @API:BRANCH-CLI-001
        @DESIGN:COMMAND-DISPATCH-001
        """
        if not args:
            self.list_branches()
            return

        action = args[0]

        if action == "list":
            self.list_branches()
        elif action == "create" and len(args) > 1:
            description = " ".join(args[1:])
            self.create_branch(description)
        elif action == "switch" and len(args) > 1:
            branch_name = args[1]
            self.switch_branch(branch_name)
        elif action == "delete" and len(args) > 1:
            branch_name = args[1]
            self.delete_branch(branch_name)
        elif action == "clean":
            self.clean_branches()
        elif action == "--status":
            self.show_status()
        else:
            print("❌ 알 수 없는 명령어입니다.")
            print("사용법: python3 branch_manager.py [list|create|switch|delete|clean|--status] [옵션]")

def main():
    """진입점

    @API:MAIN-ENTRY-001
    @TECH:CLI-INTERFACE-001
    """
    manager = BranchManager()
    manager.run(sys.argv[1:])

if __name__ == "__main__":
    main()