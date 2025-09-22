#!/usr/bin/env python3
"""
MoAI-ADK 자동 버전 업데이트 스크립트

사용법:
    python scripts/update_version.py <new_version>
    python scripts/update_version.py 0.2.1 --verify
    python scripts/update_version.py 0.2.1 --dry-run

기능:
    - _version.py 파일의 버전 업데이트
    - 전체 프로젝트의 버전 동기화
    - 변경사항 검증
    - Git 커밋까지 자동화
"""

import sys
import re
import argparse
from pathlib import Path
from datetime import datetime

# MoAI-ADK 모듈 임포트를 위한 경로 설정
current_dir = Path(__file__).parent
project_root = current_dir.parent
sys.path.insert(0, str(project_root / "src"))

try:
    from moai_adk.core.version_sync import VersionSyncManager
    from moai_adk.utils.logger import get_logger
except ImportError as e:
    print(f"❌ MoAI-ADK 모듈을 가져올 수 없습니다: {e}")
    print("프로젝트 루트에서 실행하거나 패키지를 설치하세요.")
    sys.exit(1)

logger = get_logger(__name__)


def validate_version_format(version: str) -> bool:
    """버전 형식 검증 (MAJOR.MINOR.PATCH)"""
    pattern = r"^[0-9]+\.[0-9]+\.[0-9]+$"
    return bool(re.match(pattern, version))


def update_version_file(new_version: str, dry_run: bool = False) -> bool:
    """_version.py 파일의 버전 정보 업데이트"""
    version_file = project_root / "src" / "moai_adk" / "_version.py"

    if not version_file.exists():
        logger.error("버전 파일을 찾을 수 없습니다: %s", version_file)
        return False

    try:
        # 현재 파일 내용 읽기
        with open(version_file, 'r', encoding='utf-8') as f:
            content = f.read()

        original_content = content

        # 버전 패턴들 업데이트
        patterns = [
            (r'__version__\s*=\s*"[^"]*"', f'__version__ = "{new_version}"'),
            (r'"moai_adk":\s*"[^"]*"', f'"moai_adk": "{new_version}"'),
            (r'"core":\s*"[^"]*"', f'"core": "{new_version}"'),
            (r'"templates":\s*"[^"]*"', f'"templates": "{new_version}"'),
        ]

        for pattern, replacement in patterns:
            if re.search(pattern, content):
                content = re.sub(pattern, replacement, content)
                logger.debug("패턴 업데이트: %s", pattern)

        # 변경사항이 있는지 확인
        if content == original_content:
            logger.warning("버전 파일에서 업데이트할 패턴을 찾지 못했습니다")
            return False

        # 드라이 런이 아닌 경우에만 파일 쓰기
        if not dry_run:
            with open(version_file, 'w', encoding='utf-8') as f:
                f.write(content)
            logger.info("버전 파일 업데이트 완료: %s", version_file)
        else:
            logger.info("드라이 런: 버전 파일 업데이트 시뮬레이션 완료")

        return True

    except Exception as e:
        logger.error("버전 파일 업데이트 실패: %s", e)
        return False


def run_git_commands(new_version: str, dry_run: bool = False) -> bool:
    """Git 커밋 및 태그 생성"""
    import subprocess

    if dry_run:
        logger.info("드라이 런: Git 명령어 시뮬레이션")
        return True

    try:
        # Git 상태 확인
        result = subprocess.run(['git', 'status', '--porcelain'],
                              capture_output=True, text=True, cwd=project_root)

        if result.stdout.strip():
            logger.info("변경사항이 감지되었습니다. Git 커밋을 진행합니다.")

            # 변경사항 스테이징
            subprocess.run(['git', 'add', '.'], cwd=project_root, check=True)

            # 커밋 메시지 생성
            commit_message = f"""chore(release): bump version to v{new_version}

- 전체 프로젝트 버전 정보 동기화
- 템플릿 및 문서 버전 업데이트
- 자동화된 버전 관리 시스템 적용

🤖 Generated with MoAI-ADK Version Sync System

Co-Authored-By: MoAI-ADK <noreply@moai.dev>"""

            # 커밋 실행
            subprocess.run(['git', 'commit', '-m', commit_message],
                         cwd=project_root, check=True)

            # 태그 생성
            subprocess.run(['git', 'tag', f'v{new_version}', '-m', f'Release v{new_version}'],
                         cwd=project_root, check=True)

            logger.info("Git 커밋 및 태그 생성 완료: v%s", new_version)
            return True
        else:
            logger.info("커밋할 변경사항이 없습니다.")
            return True

    except subprocess.CalledProcessError as e:
        logger.error("Git 명령어 실행 실패: %s", e)
        return False
    except Exception as e:
        logger.error("Git 작업 중 오류 발생: %s", e)
        return False


def show_next_steps(new_version: str):
    """다음 단계 안내"""
    print(f"\n{'='*70}")
    print(f"🗿  MoAI-ADK 버전 업데이트 완료: v{new_version}")
    print(f"{'='*70}")
    print("\n✅ 완료된 작업:")
    print(f"   • _version.py 파일 업데이트")
    print(f"   • 프로젝트 전체 버전 동기화")
    print(f"   • Git 커밋 및 태그 생성")
    print(f"   • 템플릿 변수 자동 적용")

    print(f"\n🚀 다음 단계 (선택사항):")
    print(f"   1. 변경사항 검토: git log --oneline -5")
    print(f"   2. 원격 저장소 푸시: git push origin main --tags")
    print(f"   3. 패키지 빌드: python -m build")
    print(f"   4. PyPI 업로드: python -m twine upload dist/*")

    print(f"\n💡 버전 확인:")
    print(f"   • python -c \"from moai_adk import __version__; print(__version__)\"")
    print(f"   • git tag -l | tail -5")
    print(f"{'='*70}\n")


def main():
    """메인 함수"""
    parser = argparse.ArgumentParser(
        description="MoAI-ADK 자동 버전 업데이트 도구",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
예제:
  python scripts/update_version.py 0.2.1
  python scripts/update_version.py 0.2.1 --verify
  python scripts/update_version.py 0.2.1 --dry-run
  python scripts/update_version.py 0.2.1 --no-git
        """
    )

    parser.add_argument("version", help="새 버전 (예: 0.2.1)")
    parser.add_argument("--dry-run", action="store_true",
                       help="실제 변경하지 않고 시뮬레이션만 실행")
    parser.add_argument("--verify", action="store_true",
                       help="업데이트 후 버전 동기화 검증 실행")
    parser.add_argument("--no-git", action="store_true",
                       help="Git 커밋 및 태그 생성 건너뛰기")

    args = parser.parse_args()

    # 버전 형식 검증
    if not validate_version_format(args.version):
        print("❌ 오류: 버전은 MAJOR.MINOR.PATCH 형식이어야 합니다 (예: 0.2.1)")
        sys.exit(1)

    print(f"🗿 MoAI-ADK 버전 업데이트 시작: v{args.version}")
    print(f"프로젝트 루트: {project_root}")

    if args.dry_run:
        print("🔍 드라이 런 모드: 실제 파일은 변경되지 않습니다")

    # 1단계: _version.py 파일 업데이트
    print(f"\n1️⃣ 버전 파일 업데이트 중...")
    if not update_version_file(args.version, args.dry_run):
        print("❌ 버전 파일 업데이트 실패")
        sys.exit(1)

    # 2단계: 전체 프로젝트 동기화
    print(f"\n2️⃣ 프로젝트 전체 버전 동기화 중...")
    try:
        # 동기화 실행 전에 새 버전으로 _version.py가 업데이트되어야 함
        sync_manager = VersionSyncManager(str(project_root))
        results = sync_manager.sync_all_versions(dry_run=args.dry_run)

        if results:
            total_files = sum(len(files) for files in results.values())
            print(f"✅ {total_files}개 파일에서 버전 정보 동기화 완료")
        else:
            print("ℹ️  동기화할 파일이 없습니다")

    except Exception as e:
        logger.error("버전 동기화 실패: %s", e)
        print("❌ 버전 동기화 실패")
        sys.exit(1)

    # 3단계: 검증 (옵션)
    if args.verify:
        print(f"\n3️⃣ 버전 동기화 검증 중...")
        try:
            inconsistencies = sync_manager.verify_sync()
            if inconsistencies:
                print("⚠️  일부 버전 불일치가 발견되었습니다")
                for pattern, files in inconsistencies.items():
                    print(f"   패턴 '{pattern}': {len(files)}개 파일")
            else:
                print("✅ 모든 버전 정보가 일치합니다")
        except Exception as e:
            logger.error("검증 실패: %s", e)
            print("⚠️  검증 중 오류 발생")

    # 4단계: Git 커밋 (옵션)
    if not args.no_git and not args.dry_run:
        print(f"\n4️⃣ Git 커밋 및 태그 생성 중...")
        if not run_git_commands(args.version, args.dry_run):
            print("❌ Git 작업 실패")
            sys.exit(1)
    elif args.no_git:
        print(f"\n4️⃣ Git 작업 건너뛰기 (--no-git 옵션)")

    # 완료 메시지 및 다음 단계 안내
    if not args.dry_run:
        show_next_steps(args.version)
    else:
        print(f"\n✅ 드라이 런 완료 - 실제 파일은 변경되지 않았습니다")
        print(f"실제 업데이트를 원하면 --dry-run 옵션을 제거하고 다시 실행하세요.")


if __name__ == "__main__":
    main()
