#!/usr/bin/env python3
"""
MoAI Git 통합 관리자 v1.0.0
단일 진입점을 통한 모든 Git 서브커맨드 처리

서브커맨드:
- branch: 브랜치 생성/전환/삭제
- commit: 스마트 커밋 (--auto 지원)
- checkpoint: 체크포인트 생성/관리
- rollback: 안전한 롤백
- sync: 원격 동기화
"""

import sys
import json
import subprocess
from pathlib import Path
from typing import List, Optional

class MoAIGitManager:
    """MoAI Git 통합 관리자"""

    def __init__(self):
        self.project_root = Path(__file__).resolve().parents[2]
        self.scripts_dir = self.project_root / ".moai" / "scripts"
        self.config_path = self.project_root / ".moai" / "config.json"

    def load_config(self) -> dict:
        """프로젝트 설정 로드"""
        try:
            with open(self.config_path, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception:
            return {"project": {"mode": "personal"}}

    def show_help(self):
        """도움말 표시"""
        print("🎯 MoAI Git 통합 관리자")
        print()
        print("사용법: python3 git_manager.py <서브커맨드> [옵션]")
        print()
        print("서브커맨드:")
        print("  branch    브랜치 생성/전환/삭제")
        print("  commit    스마트 커밋 (--auto 지원)")
        print("  checkpoint 체크포인트 생성/관리")
        print("  rollback  안전한 롤백")
        print("  sync      원격 동기화")
        print("  help      이 도움말 표시")
        print()
        print("예시:")
        print("  python3 git_manager.py branch create feature/new-feature")
        print("  python3 git_manager.py commit --auto")
        print("  python3 git_manager.py checkpoint '실험 시작'")
        print("  python3 git_manager.py rollback --last")
        print("  python3 git_manager.py sync --auto")

    def handle_branch(self, options: List[str]):
        """브랜치 관리"""
        print("🌿 브랜치 관리")
        print(f"옵션: {' '.join(options)}")

        if not options:
            # 현재 브랜치 표시
            result = subprocess.run(["git", "branch", "--show-current"],
                                   capture_output=True, text=True)
            if result.returncode == 0:
                print(f"현재 브랜치: {result.stdout.strip()}")

            # 모든 브랜치 목록
            result = subprocess.run(["git", "branch"], capture_output=True, text=True)
            if result.returncode == 0:
                print("브랜치 목록:")
                print(result.stdout)
            return

        action = options[0]
        config = self.load_config()
        mode = config.get("project", {}).get("mode", "personal")

        if action == "create" and len(options) > 1:
            branch_name = options[1]
            print(f"브랜치 생성: {branch_name} (모드: {mode})")

            # 브랜치 생성 및 전환
            subprocess.run(["git", "checkout", "-b", branch_name])
            print(f"✅ 브랜치 '{branch_name}' 생성 및 전환 완료")

        elif action == "switch" and len(options) > 1:
            branch_name = options[1]
            print(f"브랜치 전환: {branch_name}")
            subprocess.run(["git", "checkout", branch_name])

        elif action == "delete" and len(options) > 1:
            branch_name = options[1]
            print(f"브랜치 삭제: {branch_name}")
            subprocess.run(["git", "branch", "-d", branch_name])

        else:
            print("❌ 올바르지 않은 브랜치 명령어")
            print("사용법: branch [create|switch|delete] <브랜치명>")

    def handle_commit(self, options: List[str]):
        """스마트 커밋"""
        print("📝 스마트 커밋")

        # 변경사항 스테이징
        subprocess.run(["git", "add", "-A"])
        print("✅ 모든 변경사항 스테이징 완료")

        if options and options[0] == "--auto":
            # 자동 메시지 생성
            commit_msg = self._generate_auto_commit_message()
        else:
            # 사용자 제공 메시지
            commit_msg = " ".join(options) if options else "업데이트"

        # Constitution 준수 footer 추가
        full_message = f"""{commit_msg}

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"""

        # 커밋 실행
        result = subprocess.run(["git", "commit", "-m", full_message],
                               capture_output=True, text=True)

        if result.returncode == 0:
            print(f"✅ 커밋 완료: {commit_msg}")
        else:
            print(f"❌ 커밋 실패: {result.stderr}")

    def _generate_auto_commit_message(self) -> str:
        """자동 커밋 메시지 생성"""
        try:
            # 변경된 파일 확인
            result = subprocess.run(["git", "diff", "--cached", "--name-only"],
                                   capture_output=True, text=True)
            changed_files = result.stdout.strip().split('\n') if result.stdout.strip() else []
            file_count = len(changed_files)

            # 파일 유형별 분석
            if any('.md' in f or 'spec' in f.lower() or 'SPEC' in f for f in changed_files):
                return f"📝 명세 및 문서 업데이트 ({file_count}개 파일)"
            elif any('test' in f.lower() for f in changed_files):
                return f"🧪 테스트 추가 및 개선 ({file_count}개 파일)"
            elif any(f.endswith(('.py', '.js', '.ts', '.java')) for f in changed_files):
                return f"✨ 기능 구현 및 개선 ({file_count}개 파일)"
            elif any('config' in f.lower() or f.endswith(('.json', '.yml', '.yaml')) for f in changed_files):
                return f"⚙️ 설정 및 구성 업데이트 ({file_count}개 파일)"
            else:
                return f"🔧 프로젝트 업데이트 ({file_count}개 파일)"

        except Exception:
            return "📝 파일 업데이트"

    def handle_checkpoint(self, options: List[str]):
        """체크포인트 관리"""
        print("💾 체크포인트 관리")

        config = self.load_config()
        mode = config.get("project", {}).get("mode", "personal")

        if mode != "personal":
            print("⚠️ 체크포인트는 개인 모드에서만 지원됩니다.")
            print(f"현재 모드: {mode}")
            return

        if options and options[0] == "--list":
            # 체크포인트 목록
            result = subprocess.run(["git", "branch"], capture_output=True, text=True)
            if result.returncode == 0:
                checkpoints = [line.strip() for line in result.stdout.split('\n')
                              if 'checkpoint_' in line]
                if checkpoints:
                    print("체크포인트 목록:")
                    for cp in sorted(checkpoints, reverse=True)[:10]:
                        print(f"  {cp}")
                else:
                    print("체크포인트가 없습니다.")
            return

        # 체크포인트 생성
        from datetime import datetime
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        checkpoint_id = f"checkpoint_{timestamp}"

        message = " ".join(options) if options else f"Auto checkpoint {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}"

        # 변경사항 스테이징
        subprocess.run(["git", "add", "-A"])

        # 체크포인트 커밋
        commit_msg = f"""🔄 Checkpoint: {message}

타임스탬프: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
체크포인트 ID: {checkpoint_id}

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"""

        subprocess.run(["git", "commit", "-m", commit_msg])

        # 체크포인트 브랜치 생성
        subprocess.run(["git", "branch", checkpoint_id, "HEAD"])

        print(f"✅ 체크포인트 생성 완료: {checkpoint_id}")
        print(f"📋 메시지: {message}")

    def handle_rollback(self, options: List[str]):
        """안전한 롤백"""
        print("🔄 안전한 롤백")

        # 기존 rollback.py 스크립트 활용
        rollback_script = self.scripts_dir / "rollback.py"
        if rollback_script.exists():
            cmd = ["python3", str(rollback_script)] + options
            subprocess.run(cmd)
        else:
            print("❌ rollback.py 스크립트를 찾을 수 없습니다.")

    def handle_sync(self, options: List[str]):
        """원격 동기화"""
        print("🔄 원격 동기화")

        config = self.load_config()
        mode = config.get("project", {}).get("mode", "personal")

        if options and options[0] == "--auto":
            # 자동 동기화
            if mode == "team":
                print("팀 모드: 원격 저장소와 동기화")
                subprocess.run(["git", "push", "origin", "HEAD"])
            else:
                print("개인 모드: 로컬 작업만 정리")
        else:
            # 수동 동기화
            print(f"모드: {mode}")
            print("원격 상태 확인...")
            subprocess.run(["git", "fetch"])
            subprocess.run(["git", "status"])

    def run(self, args: List[str]):
        """메인 실행 함수"""
        if not args:
            self.show_help()
            return

        subcommand = args[0]
        options = args[1:] if len(args) > 1 else []

        # 서브커맨드별 처리
        handlers = {
            'branch': self.handle_branch,
            'commit': self.handle_commit,
            'checkpoint': self.handle_checkpoint,
            'rollback': self.handle_rollback,
            'sync': self.handle_sync,
            'help': lambda x: self.show_help()
        }

        handler = handlers.get(subcommand)
        if handler:
            try:
                handler(options)
            except Exception as e:
                print(f"❌ 오류 발생: {e}")
                print("자세한 도움말: python3 git_manager.py help")
        else:
            print(f"❌ 알 수 없는 서브커맨드: {subcommand}")
            self.show_help()

def main():
    """진입점"""
    manager = MoAIGitManager()
    manager.run(sys.argv[1:])

if __name__ == "__main__":
    main()