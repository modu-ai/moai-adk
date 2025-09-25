#!/usr/bin/env python3
"""
🗿 MoAI-ADK 버전 자동 관리 스크립트

Usage:
    python scripts/bump_version.py patch    # 0.1.24 -> 0.1.25
    python scripts/bump_version.py minor    # 0.1.24 -> 0.2.0
    python scripts/bump_version.py major    # 0.1.24 -> 1.0.0
    python scripts/bump_version.py 0.2.5    # 직접 버전 지정
"""

import re
import sys
import json
from pathlib import Path
from typing import Tuple, Optional

class VersionBumper:
    def __init__(self):
        self.project_root = Path(__file__).parent.parent
        self.files_to_update = [
            "pyproject.toml",
            "src/moai_adk/_version.py",
            ".moai/version.json"
        ]

    def get_current_version(self) -> str:
        """pyproject.toml에서 현재 버전을 읽어옵니다."""
        pyproject_path = self.project_root / "pyproject.toml"

        if not pyproject_path.exists():
            raise FileNotFoundError("pyproject.toml을 찾을 수 없습니다.")

        content = pyproject_path.read_text(encoding='utf-8')
        match = re.search(r'version\s*=\s*"([^"]+)"', content)

        if not match:
            raise ValueError("pyproject.toml에서 버전을 찾을 수 없습니다.")

        return match.group(1)

    def parse_version(self, version: str) -> Tuple[int, int, int]:
        """버전 문자열을 (major, minor, patch) 튜플로 파싱합니다."""
        match = re.match(r'(\d+)\.(\d+)\.(\d+)', version)
        if not match:
            raise ValueError(f"잘못된 버전 형식: {version}")

        return (int(match.group(1)), int(match.group(2)), int(match.group(3)))

    def bump_version(self, current: str, bump_type: str) -> str:
        """버전을 업데이트합니다."""
        if re.match(r'\d+\.\d+\.\d+', bump_type):
            # 직접 버전 지정
            return bump_type

        major, minor, patch = self.parse_version(current)

        if bump_type == "patch":
            patch += 1
        elif bump_type == "minor":
            minor += 1
            patch = 0
        elif bump_type == "major":
            major += 1
            minor = 0
            patch = 0
        else:
            raise ValueError(f"지원하지 않는 bump_type: {bump_type}")

        return f"{major}.{minor}.{patch}"

    def update_pyproject_toml(self, new_version: str) -> None:
        """pyproject.toml 버전을 업데이트합니다."""
        file_path = self.project_root / "pyproject.toml"
        content = file_path.read_text(encoding='utf-8')

        # version = "0.1.24" 형태 업데이트
        new_content = re.sub(
            r'version\s*=\s*"[^"]+"',
            f'version = "{new_version}"',
            content
        )

        file_path.write_text(new_content, encoding='utf-8')
        print(f"✅ pyproject.toml 업데이트 완료: {new_version}")

    def update_version_py(self, new_version: str) -> None:
        """_version.py 파일을 업데이트합니다."""
        file_path = self.project_root / "src/moai_adk/_version.py"
        content = file_path.read_text(encoding='utf-8')

        # __version__ = "0.1.17" 업데이트
        content = re.sub(
            r'__version__\s*=\s*"[^"]+"',
            f'__version__ = "0.1.17"',
            content
        )

        # VERSIONS 딕셔너리의 moai_adk, core, templates 업데이트
        content = re.sub(
            r'"moai_adk":\s*"[^"]+"',
            f'"moai_adk": "{new_version}"',
            content
        )
        content = re.sub(
            r'"core":\s*"[^"]+"',
            f'"core": "{new_version}"',
            content
        )
        content = re.sub(
            r'"templates":\s*"[^"]+"',
            f'"templates": "{new_version}"',
            content
        )

        file_path.write_text(content, encoding='utf-8')
        print(f"✅ _version.py 업데이트 완료: {new_version}")

    def update_moai_version_json(self, new_version: str) -> None:
        """/.moai/version.json 파일을 업데이트합니다."""
        file_path = self.project_root / ".moai/version.json"

        if file_path.exists():
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    data = json.load(f)

                data["template_version"] = new_version
                data["package_version"] = new_version

                with open(file_path, 'w', encoding='utf-8') as f:
                    json.dump(data, f, indent=2, ensure_ascii=False)

                print(f"✅ .moai/version.json 업데이트 완료: {new_version}")
            except (json.JSONDecodeError, KeyError) as e:
                print(f"⚠️ .moai/version.json 업데이트 실패: {e}")
        else:
            print("⚠️ .moai/version.json 파일이 없습니다.")

    def verify_versions(self, expected_version: str) -> bool:
        """모든 파일의 버전이 일치하는지 확인합니다."""
        print(f"\n🔍 버전 일관성 검증 중... (예상 버전: {expected_version})")

        errors = []

        # pyproject.toml 확인
        try:
            current = self.get_current_version()
            if current != expected_version:
                errors.append(f"pyproject.toml: {current} != {expected_version}")
            else:
                print(f"✅ pyproject.toml: {current}")
        except Exception as e:
            errors.append(f"pyproject.toml 읽기 실패: {e}")

        # _version.py 확인
        version_py_path = self.project_root / "src/moai_adk/_version.py"
        try:
            content = version_py_path.read_text(encoding='utf-8')
            match = re.search(r'__version__\s*=\s*"([^"]+)"', content)
            if match:
                version = match.group(1)
                if version != expected_version:
                    errors.append(f"_version.py __version__: {version} != {expected_version}")
                else:
                    print(f"✅ _version.py __version__: {version}")
            else:
                errors.append("_version.py에서 __version__을 찾을 수 없음")
        except Exception as e:
            errors.append(f"_version.py 읽기 실패: {e}")

        # .moai/version.json 확인
        version_json_path = self.project_root / ".moai/version.json"
        if version_json_path.exists():
            try:
                with open(version_json_path, 'r', encoding='utf-8') as f:
                    data = json.load(f)

                template_ver = data.get("template_version")
                package_ver = data.get("package_version")

                if template_ver != expected_version:
                    errors.append(f".moai/version.json template_version: {template_ver} != {expected_version}")
                else:
                    print(f"✅ .moai/version.json template_version: {template_ver}")

                if package_ver != expected_version:
                    errors.append(f".moai/version.json package_version: {package_ver} != {expected_version}")
                else:
                    print(f"✅ .moai/version.json package_version: {package_ver}")

            except Exception as e:
                errors.append(f".moai/version.json 읽기 실패: {e}")

        if errors:
            print(f"\n❌ 버전 불일치 발견:")
            for error in errors:
                print(f"  - {error}")
            return False
        else:
            print(f"\n✅ 모든 파일의 버전이 {expected_version}으로 일치합니다!")
            return True

    def run(self, bump_type: str) -> None:
        """버전 업데이트를 실행합니다."""
        try:
            # 현재 버전 확인
            current_version = self.get_current_version()
            print(f"📋 현재 버전: {current_version}")

            # 새 버전 계산
            new_version = self.bump_version(current_version, bump_type)
            print(f"🚀 새 버전: {new_version}")

            # 확인 요청
            response = input(f"\n{current_version} -> {new_version} 으로 업데이트하시겠습니까? (y/N): ")
            if response.lower() not in ['y', 'yes', '네', 'ㅇ']:
                print("❌ 취소되었습니다.")
                return

            # 파일들 업데이트
            print(f"\n📦 버전 업데이트 중...")
            self.update_pyproject_toml(new_version)
            self.update_version_py(new_version)
            self.update_moai_version_json(new_version)

            # 검증
            if self.verify_versions(new_version):
                print(f"\n🎉 버전 업데이트 완료!")
                print(f"\n💡 다음 단계:")
                print(f"   1. pip install -e . 로 개발 모드 재설치")
                print(f"   2. python -m build 로 패키지 빌드")
                print(f"   3. git add . && git commit -m '🔖 chore: bump version to {new_version}'")
            else:
                print(f"\n❌ 버전 업데이트에 문제가 있습니다. 수동으로 확인해주세요.")
                sys.exit(1)

        except Exception as e:
            print(f"❌ 오류 발생: {e}")
            sys.exit(1)

def main():
    if len(sys.argv) != 2:
        print("사용법: python scripts/bump_version.py <patch|minor|major|버전번호>")
        print("예시:")
        print("  python scripts/bump_version.py patch    # 0.1.24 -> 0.1.25")
        print("  python scripts/bump_version.py minor    # 0.1.24 -> 0.2.0")
        print("  python scripts/bump_version.py major    # 0.1.24 -> 1.0.0")
        print("  python scripts/bump_version.py 0.2.5    # 직접 버전 지정")
        sys.exit(1)

    bump_type = sys.argv[1]
    bumper = VersionBumper()
    bumper.run(bump_type)

if __name__ == "__main__":
    main()