#!/usr/bin/env python3
"""
🔍 MoAI-ADK 버전 일관성 검사 스크립트

이 스크립트는 프로젝트의 모든 버전이 일치하는지 확인합니다.
Pre-commit hook이나 CI에서 사용할 수 있습니다.

Usage:
    python scripts/check_version_consistency.py

Exit codes:
    0: 모든 버전이 일치
    1: 버전 불일치 발견
"""

import re
import sys
import json
from pathlib import Path
from typing import Dict, List, Tuple

class VersionChecker:
    def __init__(self):
        self.project_root = Path(__file__).parent.parent
        self.version_sources = {
            "pyproject.toml": self._get_pyproject_version,
            "_version.py": self._get_version_py_version,
            ".moai/version.json": self._get_moai_version_json
        }

    def _get_pyproject_version(self) -> str:
        """pyproject.toml에서 버전을 읽어옵니다."""
        file_path = self.project_root / "pyproject.toml"
        content = file_path.read_text(encoding='utf-8')
        match = re.search(r'version\s*=\s*"([^"]+)"', content)

        if not match:
            raise ValueError("pyproject.toml에서 버전을 찾을 수 없습니다.")

        return match.group(1)

    def _get_version_py_version(self) -> Dict[str, str]:
        """_version.py에서 여러 버전을 읽어옵니다."""
        file_path = self.project_root / "src/moai_adk/_version.py"
        content = file_path.read_text(encoding='utf-8')

        versions = {}

        # __version__ 추출
        main_match = re.search(r'__version__\s*=\s*"([^"]+)"', content)
        if main_match:
            versions["__version__"] = main_match.group(1)

        # VERSIONS 딕셔너리에서 주요 버전들 추출
        moai_match = re.search(r'"moai_adk":\s*"([^"]+)"', content)
        if moai_match:
            versions["moai_adk"] = moai_match.group(1)

        core_match = re.search(r'"core":\s*"([^"]+)"', content)
        if core_match:
            versions["core"] = core_match.group(1)

        templates_match = re.search(r'"templates":\s*"([^"]+)"', content)
        if templates_match:
            versions["templates"] = templates_match.group(1)

        return versions

    def _get_moai_version_json(self) -> Dict[str, str]:
        """/.moai/version.json에서 버전을 읽어옵니다."""
        file_path = self.project_root / ".moai/version.json"

        if not file_path.exists():
            return {}

        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                data = json.load(f)

            return {
                "template_version": data.get("template_version", ""),
                "package_version": data.get("package_version", "")
            }
        except (json.JSONDecodeError, KeyError):
            return {}

    def check_consistency(self) -> Tuple[bool, List[str]]:
        """모든 버전의 일관성을 확인합니다."""
        issues = []

        try:
            # 모든 버전 정보 수집
            pyproject_version = self._get_pyproject_version()
            version_py_versions = self._get_version_py_version()
            moai_versions = self._get_moai_version_json()

            print(f"📋 발견된 버전들:")
            print(f"  pyproject.toml: {pyproject_version}")

            for key, value in version_py_versions.items():
                print(f"  _version.py {key}: {value}")

            for key, value in moai_versions.items():
                if value:  # 값이 있을 때만 출력
                    print(f"  .moai/version.json {key}: {value}")

            # 기준 버전 (pyproject.toml)
            base_version = pyproject_version

            # _version.py 검사
            for key, version in version_py_versions.items():
                if version != base_version:
                    issues.append(f"_version.py {key}: {version} != {base_version}")

            # .moai/version.json 검사 (파일이 있을 때만)
            if moai_versions:
                for key, version in moai_versions.items():
                    if version and version != base_version:
                        issues.append(f".moai/version.json {key}: {version} != {base_version}")

            return len(issues) == 0, issues

        except Exception as e:
            issues.append(f"버전 검사 중 오류 발생: {e}")
            return False, issues

    def run(self) -> int:
        """버전 일관성 검사를 실행합니다."""
        print("🔍 MoAI-ADK 버전 일관성 검사 시작...")

        is_consistent, issues = self.check_consistency()

        if is_consistent:
            print("✅ 모든 버전이 일치합니다!")
            return 0
        else:
            print("❌ 버전 불일치 발견:")
            for issue in issues:
                print(f"  - {issue}")

            print("\n💡 해결 방법:")
            print("  python scripts/bump_version.py <현재버전>  # 모든 버전을 통일")
            print("  예: python scripts/bump_version.py 0.1.25")

            return 1

def main():
    checker = VersionChecker()
    exit_code = checker.run()
    sys.exit(exit_code)

if __name__ == "__main__":
    main()