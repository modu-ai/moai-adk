#!/usr/bin/env python3
"""
MoAI 커밋 도우미 v0.1.0
자동 메시지 생성 및 Constitution 5원칙 준수 커밋 시스템

@REQ:GIT-COMMIT-001
@FEATURE:AUTO-COMMIT-001
@API:COMMIT-INTERFACE-001
@DESIGN:COMMIT-WORKFLOW-001
@TECH:CLAUDE-CODE-STD-001
"""

import sys
import subprocess
from pathlib import Path
from datetime import datetime

class CommitHelper:
    """커밋 도우미

    @FEATURE:AUTO-COMMIT-001
    @API:COMMIT-INTERFACE-001
    """

    def __init__(self):
        self.project_root = Path(__file__).resolve().parents[2]

    def get_changed_files(self) -> list:
        """스테이징된 변경 파일 목록 가져오기

        @API:GIT-STATUS-001
        @DATA:FILE-LIST-001
        """
        try:
            result = subprocess.run(
                ["git", "diff", "--cached", "--name-only"],
                capture_output=True, text=True, check=True
            )
            return result.stdout.strip().split('\n') if result.stdout.strip() else []
        except Exception:
            return []

    def generate_auto_message(self) -> str:
        """자동 커밋 메시지 생성

        @FEATURE:AUTO-MESSAGE-001
        @DESIGN:MESSAGE-PATTERN-001
        """
        changed_files = self.get_changed_files()
        file_count = len(changed_files)

        if not changed_files:
            return "📝 빈 커밋"

        # 파일 유형별 분석
        if any('.md' in f or 'spec' in f.lower() or 'SPEC' in f for f in changed_files):
            return f"📝 명세 및 문서 업데이트 ({file_count}개 파일)"
        elif any('test' in f.lower() for f in changed_files):
            return f"🧪 테스트 추가 및 개선 ({file_count}개 파일)"
        elif any(f.endswith(('.py', '.js', '.ts', '.java', '.go', '.rs')) for f in changed_files):
            return f"✨ 기능 구현 및 개선 ({file_count}개 파일)"
        elif any('config' in f.lower() or f.endswith(('.json', '.yml', '.yaml', '.toml')) for f in changed_files):
            return f"⚙️ 설정 및 구성 업데이트 ({file_count}개 파일)"
        elif any('.claude' in f for f in changed_files):
            return f"🔧 Claude Code 설정 업데이트 ({file_count}개 파일)"
        else:
            return f"🔧 프로젝트 업데이트 ({file_count}개 파일)"

    def create_commit_message(self, user_input: str) -> str:
        """완전한 커밋 메시지 생성

        @DESIGN:COMMIT-MESSAGE-001
        @TECH:CLAUDE-CODE-STD-001
        """
        if user_input == "--auto":
            commit_msg = self.generate_auto_message()
            detail = "자동 생성된 커밋 메시지"
        elif user_input.startswith("--checkpoint"):
            checkpoint_msg = user_input.replace("--checkpoint", "").strip()
            commit_msg = f"🔄 체크포인트: {checkpoint_msg}" if checkpoint_msg else "🔄 자동 체크포인트"
            detail = f"체크포인트 생성: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}"
        else:
            commit_msg = user_input if user_input else "📝 업데이트"
            detail = "사용자 지정 커밋 메시지"

        # Constitution 준수 footer 추가
        full_message = f"""{commit_msg}

{detail}

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"""

        return full_message

    def execute_commit(self, message: str):
        """커밋 실행

        @API:GIT-COMMIT-001
        @TECH:SUBPROCESS-001
        """
        try:
            result = subprocess.run(
                ["git", "commit", "-m", message],
                capture_output=True, text=True, check=True
            )

            # 커밋 성공 메시지
            commit_hash = subprocess.run(
                ["git", "rev-parse", "--short", "HEAD"],
                capture_output=True, text=True
            ).stdout.strip()

            print(f"✅ 커밋 완료: {commit_hash}")
            print(f"📝 메시지: {message.split(chr(10))[0]}")  # 첫 번째 줄만 표시

        except subprocess.CalledProcessError as e:
            print(f"❌ 커밋 실패: {e.stderr}")
            sys.exit(1)

    def run(self, args: list):
        """메인 실행 함수

        @API:COMMIT-INTERFACE-001
        @DESIGN:COMMIT-WORKFLOW-001
        """
        user_input = " ".join(args) if args else "--auto"

        print(f"📝 커밋 처리: {user_input}")

        # 변경사항 확인
        changed_files = self.get_changed_files()
        if not changed_files:
            print("⚠️ 스테이징된 변경사항이 없습니다.")
            return

        # 커밋 메시지 생성
        commit_message = self.create_commit_message(user_input)

        # 커밋 실행
        self.execute_commit(commit_message)

def main():
    """진입점

    @API:MAIN-ENTRY-001
    @TECH:CLI-INTERFACE-001
    """
    helper = CommitHelper()
    helper.run(sys.argv[1:])

if __name__ == "__main__":
    main()