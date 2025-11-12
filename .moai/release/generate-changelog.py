#!/usr/bin/env python3
"""
generate-changelog.py - 자동 CHANGELOG 생성 유틸리티

역할:
  - git 커밋 히스토리에서 변경사항 자동 추출
  - 커밋 타입별 분류 (feat, fix, docs, etc)
  - Conventional Commits 형식 지원
  - CHANGELOG.md 자동 업데이트

사용:
  python generate-changelog.py              # 마지막 릴리스 이후 변경사항
  python generate-changelog.py --all        # 전체 히스토리
  python generate-changelog.py --preview    # 미리보기 (파일 수정 안함)
"""

import re
import subprocess
import sys
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Tuple, Optional


class ChangelogGenerator:
    """자동 CHANGELOG 생성"""

    # 커밋 타입별 섹션 매핑
    COMMIT_TYPE_MAP = {
        "feat": "✨ 새로운 기능",
        "fix": "🐛 버그 수정",
        "docs": "📚 문서",
        "style": "🎨 스타일",
        "refactor": "♻️  리팩토링",
        "perf": "⚡ 성능 개선",
        "test": "✅ 테스트",
        "ci": "🔧 CI/CD",
        "chore": "🔨 유지보수",
    }

    def __init__(self, project_root: Optional[Path] = None):
        """
        Args:
            project_root: 프로젝트 루트 디렉토리
        """
        self.project_root = project_root or Path.cwd()
        self.changelog_path = self.project_root / "CHANGELOG.md"

    def get_last_tag(self) -> Optional[str]:
        """마지막 git 태그 가져오기"""
        try:
            result = subprocess.run(
                ["git", "describe", "--tags", "--abbrev=0"],
                cwd=self.project_root,
                capture_output=True,
                text=True,
            )

            if result.returncode == 0:
                return result.stdout.strip()
        except FileNotFoundError:
            pass

        return None

    def get_commits(self, from_tag: Optional[str] = None) -> List[Tuple[str, str, str]]:
        """
        커밋 목록 가져오기

        Returns:
            (타입, 제목, 본문) 튜플 리스트
        """
        try:
            # git log 명령어 구성
            if from_tag:
                range_spec = f"{from_tag}..HEAD"
            else:
                range_spec = "HEAD"

            result = subprocess.run(
                [
                    "git",
                    "log",
                    range_spec,
                    "--pretty=format:%s%n%b---COMMIT_SEPARATOR---",
                ],
                cwd=self.project_root,
                capture_output=True,
                text=True,
            )

            if result.returncode != 0:
                return []

            commits = []
            commit_blocks = result.stdout.split("---COMMIT_SEPARATOR---")

            for block in commit_blocks:
                block = block.strip()
                if not block:
                    continue

                lines = block.split("\n")
                subject = lines[0] if lines else ""

                # Conventional Commits 형식 파싱: type(scope): message
                match = re.match(r"^(\w+)(?:\(.+\))?:\s*(.+)$", subject)

                if match:
                    commit_type = match.group(1)
                    message = match.group(2)
                    body = "\n".join(lines[1:]).strip()

                    commits.append((commit_type, message, body))
                else:
                    # 형식이 맞지 않는 커밋도 추가 (chore로 분류)
                    commits.append(("chore", subject, ""))

            return commits

        except FileNotFoundError:
            print("❌ git 명령을 찾을 수 없음", file=sys.stderr)
            return []

    def categorize_commits(
        self, commits: List[Tuple[str, str, str]]
    ) -> Dict[str, List[str]]:
        """
        커밋을 타입별로 분류

        Returns:
            {섹션: [메시지]} 딕셔너리
        """
        categorized: Dict[str, List[str]] = {}

        for commit_type, message, _ in commits:
            # 타입에 맞는 섹션 찾기
            section = self.COMMIT_TYPE_MAP.get(commit_type, "기타")

            if section not in categorized:
                categorized[section] = []

            # 중복 제거
            if message not in categorized[section]:
                categorized[section].append(message)

        return categorized

    def format_changelog_section(
        self, version: str, changes: Dict[str, List[str]], include_date: bool = True
    ) -> str:
        """CHANGELOG 섹션 포맷"""
        date_str = f" ({datetime.now().strftime('%Y-%m-%d')})" if include_date else ""
        section = f"## v{version}{date_str}\n\n"

        if not changes:
            section += "변경사항 없음\n\n"
            return section

        for category in [
            "✨ 새로운 기능",
            "🐛 버그 수정",
            "📚 문서",
            "🎨 스타일",
            "♻️  리팩토링",
            "⚡ 성능 개선",
            "✅ 테스트",
            "🔧 CI/CD",
            "🔨 유지보수",
            "기타",
        ]:
            if category in changes and changes[category]:
                section += f"### {category}\n\n"
                for message in sorted(set(changes[category])):
                    section += f"- {message}\n"
                section += "\n"

        return section

    def generate_from_commits(
        self, from_version: Optional[str] = None, preview: bool = False
    ) -> str:
        """커밋 히스토리에서 CHANGELOG 생성"""
        # 마지막 태그 찾기
        last_tag = self.get_last_tag()

        if last_tag:
            print(f"마지막 릴리스 태그: {last_tag}")
        else:
            print("이전 릴리스 태그 없음 (전체 히스토리 사용)")

        # 커밋 가져오기
        commits = self.get_commits(last_tag)

        if not commits:
            print("변경사항 없음")
            return ""

        print(f"수집된 커밋: {len(commits)}개")

        # 분류하기
        changes = self.categorize_commits(commits)

        # 버전 결정
        version = from_version or self._get_next_version()

        # CHANGELOG 섹션 생성
        section = self.format_changelog_section(version, changes)

        if preview:
            print("\n미리보기:\n")
            print(section)
        else:
            # 기존 CHANGELOG 읽기
            if self.changelog_path.exists():
                existing_content = self.changelog_path.read_text()
            else:
                existing_content = "# CHANGELOG\n\n모든 주목할 만한 변경사항은 이 파일에 문서화됩니다.\n\n"

            # 새 섹션을 기존 내용 앞에 삽입
            lines = existing_content.split("\n")
            insert_pos = 0

            for i, line in enumerate(lines):
                if line.startswith("## v") or (
                    i > 0 and lines[i - 1].startswith("# CHANGELOG")
                ):
                    insert_pos = i
                    break

            new_content = (
                "\n".join(lines[:insert_pos])
                + "\n"
                + section
                + "\n".join(lines[insert_pos:])
            )

            self.changelog_path.write_text(new_content)
            print(f"✓ CHANGELOG.md 업데이트: v{version}")

        return section

    def _get_next_version(self) -> str:
        """다음 버전 추정"""
        try:
            from bump_version import VersionBumper

            bumper = VersionBumper(self.project_root)
            current = bumper.get_current_version()
            return bumper.bump_patch(current)
        except (ImportError, Exception):
            # fallback: 현재 시간 사용
            return datetime.now().strftime("%Y.%m.%d")

    def generate_all(self, preview: bool = False) -> str:
        """전체 히스토리에서 CHANGELOG 생성"""
        print("전체 히스토리 기반 CHANGELOG 생성...")
        return self.generate_from_commits(from_version=None, preview=preview)


def main():
    """메인 진입점"""
    preview = "--preview" in sys.argv
    all_history = "--all" in sys.argv

    try:
        generator = ChangelogGenerator()

        if all_history:
            generator.generate_all(preview=preview)
        else:
            generator.generate_from_commits(preview=preview)

        if not preview:
            print("\n✅ CHANGELOG 생성 완료")

    except Exception as e:
        print(f"❌ 오류: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
