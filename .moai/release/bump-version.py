#!/usr/bin/env python3
"""
bump-version.py - 의미 있는 버전 관리 유틸리티

역할:
  - Semantic Versioning (semver) 준수
  - pyproject.toml 버전 업데이트
  - 소스 코드 __init__.py 버전 동기화
  - CHANGELOG.md 버전 섹션 추가
  - git 태그 생성

사용:
  python bump-version.py patch   # 0.22.5 → 0.22.6
  python bump-version.py minor   # 0.22.5 → 0.23.0
  python bump-version.py major   # 0.22.5 → 1.0.0
  python bump-version.py --current  # 현재 버전 표시
"""

import re
import sys
import subprocess
from datetime import datetime
from pathlib import Path
from typing import Tuple, Optional


class VersionBumper:
    """의미 있는 버전 범프 관리"""

    def __init__(self, project_root: Optional[Path] = None):
        """
        Args:
            project_root: 프로젝트 루트 디렉토리 (기본값: 현재 디렉토리)
        """
        self.project_root = project_root or Path.cwd()
        self.pyproject_path = self.project_root / "pyproject.toml"
        self.init_path = self.project_root / "src" / "moai_adk" / "__init__.py"
        self.changelog_path = self.project_root / "CHANGELOG.md"

        # 파일 존재 확인
        if not self.pyproject_path.exists():
            raise FileNotFoundError(f"pyproject.toml 찾을 수 없음: {self.pyproject_path}")

    def get_current_version(self) -> str:
        """현재 버전 읽기"""
        content = self.pyproject_path.read_text()

        # 정규식으로 버전 추출
        match = re.search(r'version\s*=\s*"([^"]+)"', content)
        if not match:
            raise ValueError("pyproject.toml에서 버전 정보를 찾을 수 없음")

        return match.group(1)

    def parse_version(self, version: str) -> Tuple[int, int, int]:
        """버전 문자열을 (major, minor, patch)로 파싱"""
        parts = version.split(".")
        if len(parts) != 3:
            raise ValueError(f"잘못된 버전 형식: {version} (기대: X.Y.Z)")

        try:
            return tuple(int(p) for p in parts)  # type: ignore
        except ValueError:
            raise ValueError(f"버전 번호는 숫자여야 함: {version}")

    def format_version(self, major: int, minor: int, patch: int) -> str:
        """(major, minor, patch)를 버전 문자열로 포맷"""
        return f"{major}.{minor}.{patch}"

    def bump_patch(self, version: str) -> str:
        """패치 버전 증가"""
        major, minor, patch = self.parse_version(version)
        return self.format_version(major, minor, patch + 1)

    def bump_minor(self, version: str) -> str:
        """마이너 버전 증가 (패치 리셋)"""
        major, minor, _ = self.parse_version(version)
        return self.format_version(major, minor + 1, 0)

    def bump_major(self, version: str) -> str:
        """메이저 버전 증가 (마이너, 패치 리셋)"""
        major, _, _ = self.parse_version(version)
        return self.format_version(major + 1, 0, 0)

    def update_pyproject(self, new_version: str) -> None:
        """pyproject.toml 버전 업데이트"""
        content = self.pyproject_path.read_text()
        updated = re.sub(
            r'version\s*=\s*"[^"]+"',
            f'version = "{new_version}"',
            content,
        )
        self.pyproject_path.write_text(updated)
        print(f"✓ pyproject.toml 업데이트: {new_version}")

    def update_init_file(self, new_version: str) -> None:
        """소스 코드 __init__.py 버전 업데이트"""
        if not self.init_path.exists():
            print(f"⚠ {self.init_path} 찾을 수 없음 (스킵)")
            return

        content = self.init_path.read_text()

        # __version__ 찾기 및 업데이트
        if "__version__" in content:
            updated = re.sub(
                r'__version__\s*=\s*"[^"]+"',
                f'__version__ = "{new_version}"',
                content,
            )
            self.init_path.write_text(updated)
            print(f"✓ __init__.py 업데이트: {new_version}")
        else:
            print(f"⚠ __version__ 찾을 수 없음 (스킵)")

    def update_changelog(self, new_version: str) -> None:
        """CHANGELOG.md에 새 버전 섹션 추가"""
        if not self.changelog_path.exists():
            print(f"⚠ {self.changelog_path} 찾을 수 없음 (스킵)")
            return

        today = datetime.now().strftime("%Y-%m-%d")
        new_section = f"""# v{new_version} ({today})

## 주요 변경사항

- 기능 추가/개선 사항 여기에 입력
- 버그 수정 사항
- 주요 변경 사항

## 설치

\\`\\`\\`bash
pip install moai-adk=={new_version}
\\`\\`\\`

---

"""

        content = self.changelog_path.read_text()

        # 첫 번째 제목(# ) 뒤에 삽입
        lines = content.split("\n")
        insert_pos = 0

        for i, line in enumerate(lines):
            if line.startswith("# "):
                insert_pos = i + 1
                break

        new_lines = lines[:insert_pos] + new_section.split("\n") + lines[insert_pos:]
        updated_content = "\n".join(new_lines)

        self.changelog_path.write_text(updated_content)
        print(f"✓ CHANGELOG.md 업데이트: v{new_version} 섹션 추가")

    def create_git_tag(self, new_version: str) -> None:
        """git 태그 생성 (선택사항)"""
        try:
            tag_name = f"v{new_version}"

            # 태그 존재 여부 확인
            result = subprocess.run(
                ["git", "tag", "-l", tag_name],
                cwd=self.project_root,
                capture_output=True,
                text=True,
            )

            if result.stdout.strip():
                print(f"⚠ git 태그 이미 존재: {tag_name} (스킵)")
                return

            # 새 태그 생성
            subprocess.run(
                ["git", "tag", "-a", tag_name, "-m", f"Release v{new_version}"],
                cwd=self.project_root,
                check=True,
                capture_output=True,
            )
            print(f"✓ git 태그 생성: {tag_name}")

        except subprocess.CalledProcessError as e:
            print(f"⚠ git 태그 생성 실패: {e}")
        except FileNotFoundError:
            print("⚠ git 명령 찾을 수 없음 (git 태그 스킵)")

    def bump(self, bump_type: str) -> str:
        """버전 범프 실행"""
        if bump_type not in ("patch", "minor", "major"):
            raise ValueError(f"잘못된 범프 타입: {bump_type} (patch/minor/major)")

        current = self.get_current_version()
        print(f"\n현재 버전: {current}")

        # 새 버전 계산
        if bump_type == "patch":
            new_version = self.bump_patch(current)
        elif bump_type == "minor":
            new_version = self.bump_minor(current)
        else:  # major
            new_version = self.bump_major(current)

        print(f"새 버전:  {new_version} ({bump_type})")

        # 파일 업데이트
        print("\n파일 업데이트 중...")
        self.update_pyproject(new_version)
        self.update_init_file(new_version)
        self.update_changelog(new_version)

        # git 태그 생성
        print("\ngit 태그 생성 중...")
        self.create_git_tag(new_version)

        print(f"\n✅ 버전 범프 완료: {current} → {new_version}")
        return new_version


def main():
    """메인 진입점"""
    if len(sys.argv) < 2:
        print("사용: bump-version.py [patch|minor|major|--current]")
        print("")
        print("예제:")
        print("  python bump-version.py patch   # 0.22.5 → 0.22.6")
        print("  python bump-version.py minor   # 0.22.5 → 0.23.0")
        print("  python bump-version.py major   # 0.22.5 → 1.0.0")
        print("  python bump-version.py --current")
        sys.exit(1)

    try:
        bumper = VersionBumper()

        if sys.argv[1] == "--current":
            current = bumper.get_current_version()
            print(f"현재 버전: {current}")
        else:
            bumper.bump(sys.argv[1])

    except (FileNotFoundError, ValueError, subprocess.CalledProcessError) as e:
        print(f"❌ 오류: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
