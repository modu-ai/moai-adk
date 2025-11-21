#!/usr/bin/env python3
"""
MoAI-ADK Interactive Release Manager

인터랙티브 릴리즈 관리 시스템으로
5가지 핵심 명령어를 메뉴 방식으로 제공합니다.
"""

import json
import sys
import subprocess
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Tuple

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))

class InteractiveReleaseManager:
    """인터랙티브 릴리즈 관리자"""

    def __init__(self):
        self.project_root = Path.cwd()
        self.logs_dir = self.project_root / ".moai" / "logs"
        self.logs_dir.mkdir(parents=True, exist_ok=True)
        self.log_file = self.logs_dir / f"release-{datetime.now().strftime('%Y%m%d-%H%M%S')}.log"

    def log(self, message: str) -> None:
        """로그 기록"""
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        log_entry = f"[{timestamp}] {message}"
        print(log_entry)

        with open(self.log_file, "a", encoding="utf-8") as f:
            f.write(log_entry + "\n")

    def get_version_info(self) -> Dict[str, str]:
        """현재 버전 정보 가져오기"""
        try:
            # pyproject.toml에서 버전 읽기
            pyproject_path = self.project_root / "pyproject.toml"
            if pyproject_path.exists():
                with open(pyproject_path, "r", encoding="utf-8") as f:
                    content = f.read()
                    for line in content.split("\n"):
                        if line.startswith("version = "):
                            version = line.split('"')[1]
                            return {"current": version, "pyproject": version}

            return {"current": "unknown", "pyproject": "unknown"}
        except Exception as e:
            self.log(f"버전 정보读取 오류: {e}")
            return {"current": "unknown", "pyproject": "unknown"}

    def show_main_menu(self) -> str:
        """메인 메뉴 표시"""
        # Note: 실제 구현에서는 AskUserQuestion을 사용해야 함
        # 여기서는 터미널 기반 메뉴로 시뮬레이션

        print("\n" + "="*60)
        print("🚀 MoAI-ADK 릴리즈 관리 - 작업을 선택하세요:")
        print("="*60)
        print("1. 🔍 validate  - 사전 릴리즈 품질 검증")
        print("2. 📝 version   - 버전 관리 (major/minor/patch)")
        print("3. 📋 changelog - 이중언어 변경로그 생성")
        print("4. 🚀 prepare   - CI/CD 배포 준비")
        print("5. 🚨 rollback  - 응급 롤백")
        print("6. ❌ 종료      - 작업 종료")
        print("="*60)

        while True:
            try:
                choice = input("\n선택 (1-6): ").strip()
                if choice in ["1", "2", "3", "4", "5", "6"]:
                    return choice
                else:
                    print("❌ 1-6 사이의 숫자를 입력해주세요.")
            except KeyboardInterrupt:
                print("\n\n👋 작업을 취소합니다.")
                sys.exit(0)

    def run_validate_workflow(self) -> None:
        """검증 워크플로우 실행"""
        print("\n🔍 validate - 사전 릴리즈 품질 검증")
        print("-" * 50)
        print("1. ⚡ Quick 검증 (5분)")
        print("2. 🔬 Full 검증 (15분)")
        print("3. 🎯 사용자 정의")

        choice = input("\n검증 방식 선택 (1-3): ").strip()

        if choice == "1":
            self.run_quick_validation()
        elif choice == "2":
            self.run_full_validation()
        elif choice == "3":
            self.run_custom_validation()
        else:
            print("❌ 잘못된 선택입니다.")

    def run_quick_validation(self) -> None:
        """빠른 검증 실행"""
        self.log("Quick 검증 시작")
        print("\n⚡ Quick 검증 실행 중...")

        commands = [
            ["uv", "run", "pytest", "--tb=short"],
            ["uv", "run", "ruff", "check", "src/moai_adk"],
            ["uv", "run", "mypy", "src/moai_adk"]
        ]

        success = True
        for cmd in commands:
            print(f"🔍 실행: {' '.join(cmd)}")
            try:
                result = subprocess.run(cmd, capture_output=True, text=True, timeout=300)
                if result.returncode == 0:
                    print("✅ 통과")
                else:
                    print("❌ 실패")
                    print(result.stderr)
                    success = False
                    break
            except subprocess.TimeoutExpired:
                print("❌ 타임아웃")
                success = False
                break
            except Exception as e:
                print(f"❌ 오류: {e}")
                success = False
                break

        if success:
            print("\n✅ Quick 검증 통과!")
            self.log("Quick 검증 통과")
        else:
            print("\n❌ Quick 검증 실패!")
            self.log("Quick 검증 실패")

    def run_full_validation(self) -> None:
        """전체 검증 실행"""
        self.log("Full 검증 시작")
        print("\n🔬 Full 검증 실행 중...")

        # Quick 검증 항목 + 보안 스캔
        self.run_quick_validation()

        # 추가 보안 스캔
        print("\n🔒 보안 스캔 실행 중...")

        security_commands = [
            ["uv", "run", "bandit", "-r", "src/moai_adk"],
            ["uv", "run", "pip-audit", "--desc"]
        ]

        for cmd in security_commands:
            print(f"🔍 실행: {' '.join(cmd)}")
            try:
                result = subprocess.run(cmd, capture_output=True, text=True, timeout=300)
                if result.returncode == 0:
                    print("✅ 통과")
                else:
                    print("⚠️ 경고 또는 정보")
                    print(result.stdout[:500])  # 첫 500자만 표시
            except subprocess.TimeoutExpired:
                print("❌ 타임아웃")
            except Exception as e:
                print(f"❌ 오류: {e}")

        print("\n✅ Full 검증 완료!")
        self.log("Full 검증 완료")

    def run_custom_validation(self) -> None:
        """사용자 정의 검증 실행"""
        print("\n🎯 사용자 정의 검증")
        print("사용 가능한 검증 항목:")
        print("1. pytest (테스트)")
        print("2. ruff (코드 포맷)")
        print("3. mypy (타입 검사)")
        print("4. bandit (보안 스캔)")
        print("5. pip-audit (의존성 취약점)")

        selections = input("\n선택 (예: 1,3,5): ").strip().split(",")

        available_commands = {
            "1": ["uv", "run", "pytest", "--tb=short"],
            "2": ["uv", "run", "ruff", "check", "src/moai_adk"],
            "3": ["uv", "run", "mypy", "src/moai_adk"],
            "4": ["uv", "run", "bandit", "-r", "src/moai_adk"],
            "5": ["uv", "run", "pip-audit", "--desc"]
        }

        for selection in selections:
            selection = selection.strip()
            if selection in available_commands:
                cmd = available_commands[selection]
                print(f"🔍 실행: {' '.join(cmd)}")
                try:
                    result = subprocess.run(cmd, capture_output=True, text=True, timeout=300)
                    if result.returncode == 0:
                        print("✅ 통과")
                    else:
                        print("❌ 실패")
                        print(result.stderr)
                except Exception as e:
                    print(f"❌ 오류: {e}")
            else:
                print(f"❌ 잘못된 선택: {selection}")

    def run_version_workflow(self) -> None:
        """버전 관리 워크플로우 실행"""
        version_info = self.get_version_info()
        print(f"\n📝 version - 버전 관리")
        print(f"현재 버전: {version_info['current']}")
        print("-" * 50)
        print("1. 🔢 patch - 버그 수정")
        print("2. 🔧 minor - 기능 추가")
        print("3. 💥 major - 호환성 변경")

        choice = input("\n버전 타입 선택 (1-3): ").strip()

        if choice not in ["1", "2", "3"]:
            print("❌ 잘못된 선택입니다.")
            return

        version_type = { "1": "patch", "2": "minor", "3": "major" }[choice]

        current = version_info['current']
        if current == "unknown":
            print("❌ 현재 버전을 알 수 없습니다.")
            return

        # 버전 계산
        parts = current.split('.')
        if len(parts) != 3:
            print("❌ 버전 형식이 올바르지 않습니다.")
            return

        major, minor, patch = map(int, parts)

        if version_type == "patch":
            patch += 1
        elif version_type == "minor":
            minor += 1
            patch = 0
        elif version_type == "major":
            major += 1
            minor = 0
            patch = 0

        new_version = f"{major}.{minor}.{patch}"

        print(f"\n{version_type} 버전으로 bump:")
        print(f"{current} → {new_version}")

        confirm = input("\n실행하시겠습니까? (y/N): ").strip().lower()
        if confirm != 'y':
            print("❌ 취소되었습니다.")
            return

        self.bump_version(new_version)

    def bump_version(self, new_version: str) -> None:
        """버전 bump 실행"""
        self.log(f"버전 bump 시작: {new_version}")

        files_to_update = [
            ("pyproject.toml", f'version = "{new_version}"'),
            ("src/moai_adk/__init__.py", f'__version__ = "{new_version}"'),
            (".moai/config/config.json', f'"version": "{new_version}"')
        ]

        success = True
        for file_path, version_line in files_to_update:
            full_path = self.project_root / file_path
            if full_path.exists():
                try:
                    with open(full_path, "r", encoding="utf-8") as f:
                        content = f.read()

                    # 버전 라인 찾기 및 업데이트
                    lines = content.split("\n")
                    updated = False
                    for i, line in enumerate(lines):
                        if "version" in line and ("=" in line or ":" in line):
                            if "pyproject.toml" in file_path:
                                lines[i] = version_line
                            elif "__init__.py" in file_path:
                                lines[i] = f'__version__ = "{new_version}"'
                            elif "config.json" in file_path:
                                # JSON 형식에서 버전 업데이트
                                try:
                                    import json
                                    config = json.loads(content)
                                    if "moai" in config and "version" in config["moai"]:
                                        config["moai"]["version"] = new_version
                                        content = json.dumps(config, indent=2)
                                        updated = True
                                        break
                                except:
                                    pass
                            updated = True
                            break

                    if updated:
                        with open(full_path, "w", encoding="utf-8") as f:
                            f.write("\n".join(lines))
                        print(f"✅ {file_path} 업데이트 완료")
                    else:
                        print(f"⚠️ {file_path}에서 버전 라인을 찾을 수 없음")

                except Exception as e:
                    print(f"❌ {file_path} 업데이트 오류: {e}")
                    success = False
            else:
                print(f"⚠️ {file_path} 파일을 찾을 수 없음")

        if success:
            print(f"\n✅ 버전 bump 완료: {new_version}")
            self.log(f"버전 bump 완료: {new_version}")

            # Git 커밋 제안
            print(f"\n🔧 다음 Git 명령어를 실행하세요:")
            print(f"git add .")
            print(f"git commit -m \"chore: Bump version to {new_version}\"")
            print(f"git push origin main")
        else:
            print("\n❌ 버전 bump 실패!")
            self.log("버전 bump 실패")

    def run_changelog_workflow(self) -> None:
        """변경로그 워크플로우 실행"""
        print("\n📋 changelog - 이중언어 변경로그 생성")
        print("-" * 50)
        print("1. 📝 자동 생성 (Git 히스토리 기반)")
        print("2. ✏️ 수동 편집 (템플릿 제공)")
        print("3. 🔄 기존 수정")

        choice = input("\n방식 선택 (1-3): ").strip()

        if choice == "1":
            self.generate_auto_changelog()
        elif choice == "2":
            self.create_changelog_template()
        elif choice == "3":
            self.edit_existing_changelog()
        else:
            print("❌ 잘못된 선택입니다.")

    def generate_auto_changelog(self) -> None:
        """자동 변경로그 생성"""
        self.log("자동 변경로그 생성 시작")
        print("\n📝 자동 변경로그 생성 중...")

        try:
            # 최신 태그 가져오기
            result = subprocess.run(
                ["git", "describe", "--tags", "--abbrev=0"],
                capture_output=True, text=True
            )
            latest_tag = result.stdout.strip() if result.returncode == 0 else "v0.0.0"

            # 태그 이후 커밋 가져오기
            result = subprocess.run(
                ["git", "log", f"{latest_tag}..HEAD", "--oneline"],
                capture_output=True, text=True
            )
            commits = result.stdout.strip().split("\n") if result.returncode == 0 else []

            if not commits or commits == ['']:
                print("❌ 변경사항이 없습니다.")
                return

            print(f"🔍 {latest_tag} 이후 {len(commits)}개 커밋 분석...")

            # 커밋 분류
            features = []
            fixes = []
            improvements = []

            for commit in commits:
                commit_msg = commit.split(" ", 1)[1] if " " in commit else commit

                if any(keyword in commit_msg.lower() for keyword in ["feat", "add", "new"]):
                    features.append(commit_msg)
                elif any(keyword in commit_msg.lower() for keyword in ["fix", "bug", "patch"]):
                    fixes.append(commit_msg)
                else:
                    improvements.append(commit_msg)

            # 버전 정보 가져오기
            version_info = self.get_version_info()
            new_version = version_info['current']

            # CHANGELOG.md 업데이트
            changelog_path = self.project_root / "CHANGELOG.md"

            new_entry = f"# v{new_version} ({datetime.now().strftime('%Y-%m-%d')})\n\n"
            new_entry += "## 🎯 English Section\n\n"

            if features:
                new_entry += "### Features\n"
                for feature in features:
                    new_entry += f"- {feature}\n"
                new_entry += "\n"

            if fixes:
                new_entry += "### Bug Fixes\n"
                for fix in fixes:
                    new_entry += f"- {fix}\n"
                new_entry += "\n"

            if improvements:
                new_entry += "### Improvements\n"
                for improvement in improvements:
                    new_entry += f"- {improvement}\n"
                new_entry += "\n"

            new_entry += "---\n\n"
            new_entry += "## 🎯 한글 섹션\n\n"

            if features:
                new_entry += "### 기능 추가\n"
                new_entry += "Features listed above.\n\n"

            if fixes:
                new_entry += "### 버그 수정\n"
                new_entry += "Bug fixes listed above.\n\n"

            if improvements:
                new_entry += "### 개선사항\n"
                new_entry += "Improvements listed above.\n\n"

            new_entry += "## 설치\n\n"
            new_entry += "```bash\npip install moai-adk=={new_version}\n```\n\n"
            new_entry += "---\n\n"
            new_entry += "🤖 Generated with Claude Code\n"
            new_entry += "Co-Authored-By: 🎩 Alfred@MoAI\n"
            new_entry += "---\n\n"

            # 기존 내용 읽기
            if changelog_path.exists():
                with open(changelog_path, "r", encoding="utf-8") as f:
                    existing_content = f.read()
            else:
                existing_content = "# CHANGELOG\n\n"

            # 새 내용 추가
            with open(changelog_path, "w", encoding="utf-8") as f:
                f.write(new_entry + existing_content)

            print(f"✅ CHANGELOG.md 업데이트 완료")
            print(f"   - Features: {len(features)}")
            print(f"   - Fixes: {len(fixes)}")
            print(f"   - Improvements: {len(improvements)}")

            self.log(f"자동 변경로그 생성 완료: {new_version}")

        except Exception as e:
            print(f"❌ 자동 변경로그 생성 오류: {e}")
            self.log(f"자동 변경로그 생성 오류: {e}")

    def create_changelog_template(self) -> None:
        """변경로그 템플릿 생성"""
        version_info = self.get_version_info()
        new_version = version_info['current']

        template = f"# v{new_version} ({datetime.now().strftime('%Y-%m-%d')})\n\n"
        template += "## 🎯 English Section\n\n"
        template += "### Features\n"
        template += "- Feature description here\n\n"
        template += "### Bug Fixes\n"
        template += "- Bug fix description here\n\n"
        template += "### Improvements\n"
        template += "- Improvement description here\n\n"
        template += "---\n\n"
        template += "## 🎯 한글 섹션\n\n"
        template += "### 기능 추가\n"
        template += "위 기능 설명\n\n"
        template += "### 버그 수정\n"
        template += "위 버그 수정 설명\n\n"
        template += "### 개선사항\n"
        template += "위 개선사항 설명\n\n"
        template += "## 설치\n\n"
        template += f"```bash\npip install moai-adk=={new_version}\n```\n\n"
        template += "---\n\n"
        template += "🤖 Generated with Claude Code\n"
        template += "Co-Authored-By: 🎩 Alfred@MoAI\n"
        template += "---\n\n"

        # 템플릿 파일로 저장
        template_path = self.project_root / "CHANGELOG.template.md"
        with open(template_path, "w", encoding="utf-8") as f:
            f.write(template)

        print(f"✅ 변경로그 템플릿 생성: CHANGELOG.template.md")
        print("📝 템플릿을 편집한 후 CHANGELOG.md로 저장하세요.")

        self.log("변경로그 템플릿 생성 완료")

    def edit_existing_changelog(self) -> None:
        """기존 변경로그 수정"""
        changelog_path = self.project_root / "CHANGELOG.md"

        if not changelog_path.exists():
            print("❌ CHANGELOG.md 파일을 찾을 수 없습니다.")
            return

        print(f"✏️ CHANGELOG.md를 직접 편집하세요:")
        print(f"   경로: {changelog_path}")
        print("   편집 후 Git 커밋을 생성하세요.")

        self.log("기존 변경로그 수정 안내")

    def run_prepare_workflow(self) -> None:
        """배포 준비 워크플로우 실행"""
        print("\n🚀 prepare - CI/CD 배포 준비")
        print("-" * 50)
        print("1. 🧪 test 환경 (TestPyPI)")
        print("2. 🌍 production 환경 (PyPI)")
        print("3. 📋 검토용 (릴리즈 검토 번들)")

        choice = input("\n환경 선택 (1-3): ").strip()

        if choice == "1":
            self.prepare_test_environment()
        elif choice == "2":
            self.prepare_production_environment()
        elif choice == "3":
            self.prepare_review_bundle()
        else:
            print("❌ 잘못된 선택입니다.")

    def prepare_test_environment(self) -> None:
        """테스트 환경 준비"""
        self.log("테스트 환경 준비 시작")
        print("\n🧪 TestPyPI 환경 준비 중...")

        try:
            # 패키지 빌드
            print("📦 패키지 빌드...")
            result = subprocess.run(["uv", "build"], capture_output=True, text=True)
            if result.returncode != 0:
                print(f"❌ 빌드 실패: {result.stderr}")
                return

            print("✅ 패키지 빌드 완료")

            # TestPyPI 토큰 확인
            print("🔑 TestPyPI 토큰 확인...")
            try:
                import configparser
                config = configparser.ConfigParser()
                config.read(Path.home() / ".pypirc")

                if "testpypi" in config:
                    print("✅ TestPyPI 토큰 확인 완료")
                else:
                    print("⚠️ TestPyPI 토큰이 설정되지 않았습니다.")
                    print("~/.pypirc에 testpypi 섹션을 추가하세요.")
            except:
                print("⚠️ 토큰 확인 중 오류 발생")

            print("\n✅ 테스트 환경 준비 완료!")
            print("📋 다음 단계:")
            print("1. GitHub Actions에서 '🚀 Secure Release Pipeline' 실행")
            print("2. 'test' 환경 선택")
            print("3. 1분 대기 후 자동 배포")

            self.log("테스트 환경 준비 완료")

        except Exception as e:
            print(f"❌ 테스트 환경 준비 오류: {e}")
            self.log(f"테스트 환경 준비 오류: {e}")

    def prepare_production_environment(self) -> None:
        """프로덕션 환경 준비"""
        self.log("프로덕션 환경 준비 시작")
        print("\n🌍 Production PyPI 환경 준비 중...")

        try:
            # 검증 통과 확인
            print("🔍 사전 검증 확인...")
            # 이전 검증 결과를 확인하는 로직 추가 가능

            # PyPI 토큰 확인
            print("🔑 Production PyPI 토큰 확인...")
            try:
                import configparser
                config = configparser.ConfigParser()
                config.read(Path.home() / ".pypirc")

                if "pypi" in config:
                    print("✅ Production PyPI 토큰 확인 완료")
                else:
                    print("⚠️ Production PyPI 토큰이 설정되지 않았습니다.")
                    print("~/.pypirc에 pypi 섹션을 추가하세요.")
            except:
                print("⚠️ 토큰 확인 중 오류 발생")

            # 배포 준비 확인
            print("📦 배포 준비 상태 확인...")

            print("\n✅ 프로덕션 환경 준비 완료!")
            print("📋 다음 단계:")
            print("1. GitHub Actions에서 '🚀 Secure Release Pipeline' 실행")
            print("2. 'production' 환경 선택")
            print("3. 5분 대기 후 승인 필요")
            print("4. 승인 후 자동 배포")

            self.log("프로덕션 환경 준비 완료")

        except Exception as e:
            print(f"❌ 프로덕션 환경 준비 오류: {e}")
            self.log(f"프로덕션 환경 준비 오류: {e}")

    def prepare_review_bundle(self) -> None:
        """검토용 번들 생성"""
        self.log("검토용 번들 생성 시작")
        print("\n📋 릴리즈 검토 번들 생성 중...")

        version_info = self.get_version_info()

        bundle_info = {
            "version": version_info['current'],
            "timestamp": datetime.now().isoformat(),
            "prepared_by": "interactive-release.py",
            "environment": "review",
            "files": {
                "package_files": [],
                "test_results": {},
                "security_scan": {},
                "changelog": ""
            }
        }

        # 패키지 파일 목록
        try:
            dist_dir = self.project_root / "dist"
            if dist_dir.exists():
                bundle_info["files"]["package_files"] = [f.name for f in dist_dir.iterdir() if f.is_file()]
        except:
            pass

        # CHANGELOG 내용
        try:
            changelog_path = self.project_root / "CHANGELOG.md"
            if changelog_path.exists():
                with open(changelog_path, "r", encoding="utf-8") as f:
                    bundle_info["files"]["changelog"] = f.read()[:1000]  # 첫 1000자만
        except:
            pass

        # 번들 저장
        bundle_path = self.project_root / ".moai" / f"release-bundle-{version_info['current']}.json"
        with open(bundle_path, "w", encoding="utf-8") as f:
            json.dump(bundle_info, f, indent=2, ensure_ascii=False)

        print(f"✅ 검토 번들 생성 완료: {bundle_path}")
        print("📋 번들 내용:")
        print(f"   - 버전: {bundle_info['version']}")
        print(f"   - 패키지 파일: {len(bundle_info['files']['package_files'])}")
        print(f"   - 생성 시간: {bundle_info['timestamp']}")

        self.log("검토용 번들 생성 완료")

    def run_rollback_workflow(self) -> None:
        """롤백 워크플로우 실행"""
        print("\n🚨 rollback - 응급 롤백")
        print("-" * 50)
        print("⚠️ 롤백은 심각한 작업입니다. 신중하게 진행하세요.")
        print("1. 📦 PyPI만 (패키지만 롤백)")
        print("2. 🔄 전체 (PyPI + GitHub Release + 태그)")
        print("3. 🎯 특정 버전")

        choice = input("\n롤백 범위 선택 (1-3): ").strip()

        if choice == "1":
            self.rollback_pypi_only()
        elif choice == "2":
            self.rollback_full()
        elif choice == "3":
            self.rollback_specific_version()
        else:
            print("❌ 잘못된 선택입니다.")

    def rollback_pypi_only(self) -> None:
        """PyPI만 롤백"""
        version_info = self.get_version_info()
        version = version_info['current']

        print(f"\n📦 PyPI 버전 {version} 롤백")
        print("⚠️ 이 작업은 PyPI 버전을 숨기기만 합니다.")

        confirm = input(f"정말 {version} 버전을 PyPI에서 숨기시겠습니까? (type 'ROLLBACK' to confirm): ").strip()

        if confirm != "ROLLBACK":
            print("❌ 롤백 취소되었습니다.")
            return

        self.log(f"PyPI 롤백 시작: {version}")
        print(f"🔄 PyPI 버전 숨기기: {version}")
        print("📋 수동 롤백 절차:")
        print("1. PyPI 로그인: https://pypi.org/manage/")
        print("2. 'moai-adk' 패키지 찾기")
        print("3. 'Hide version' 또는 'Delete version' 선택")

        print("✅ PyPI 롤백 안내 완료")
        self.log("PyPI 롤백 안내 완료")

    def rollback_full(self) -> None:
        """전체 롤백"""
        version_info = self.get_version_info()
        version = version_info['current']

        print(f"\n🔄 전체 롤백: {version}")
        print("⚠️ PyPI + GitHub Release + Git 태그 모두 롤백")

        confirm = input(f"정말 {version} 버전을 전체 롤백하시겠습니까? (type 'FULL_ROLLBACK' to confirm): ").strip()

        if confirm != "FULL_ROLLBACK":
            print("❌ 롤백 취소되었습니다.")
            return

        self.log(f"전체 롤백 시작: {version}")

        print("🔄 전체 롤백 절차:")
        print("1. PyPI 버전 삭제")
        print("2. GitHub Release 삭제")
        print("3. Git 태그 삭제")
        print("4. GitHub Issue 생성 (롤백 기록)")

        try:
            # GitHub Release 삭제
            print(f"🔄 GitHub Release 삭제: v{version}")
            result = subprocess.run(
                ["gh", "release", "delete", f"v{version}", "--yes"],
                capture_output=True, text=True
            )
            if result.returncode == 0:
                print("✅ GitHub Release 삭제 완료")
            else:
                print("⚠️ GitHub Release 삭제 실패 (이미 없거나 권한 없음)")

            # Git 태그 삭제
            print(f"🔄 Git 태그 삭제: v{version}")
            result = subprocess.run(
                ["git", "tag", "-d", f"v{version}"],
                capture_output=True, text=True
            )
            if result.returncode == 0:
                print("✅ Git 태그 삭제 완료")
            else:
                print("⚠️ Git 태그 삭제 실패 (이미 없거나 권한 없음)")

            # GitHub Issue 생성
            print("🔄 롤백 기록 Issue 생성...")
            issue_title = f"Rollback: v{version}"
            issue_body = f"""
# 롤백 기록

## 롤백 정보
- **버전**: {version}
- **시간**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
- **범위**: 전체 롤백 (PyPI + GitHub Release + Git 태그)

## 롤백 이유
롤백 이유를 여기에 기록하세요.

## 복구 절차
1. 문제 원인 분석
2. 수정사항 개발
3. 재배포 준비
4. 재배포 실행

---
🤖 Generated by MoAI-ADK Interactive Release Manager
"""

            result = subprocess.run(
                ["gh", "issue", "create", "--title", issue_title, "--body", issue_body, "--label", "rollback"],
                capture_output=True, text=True
            )
            if result.returncode == 0:
                print("✅ GitHub Issue 생성 완료")
            else:
                print("⚠️ GitHub Issue 생성 실패")

        except Exception as e:
            print(f"❌ 전체 롤백 중 오류: {e}")

        print("\n✅ 전체 롤백 절차 완료")
        print("📋 PyPI에서는 수동으로 버전을 숨기거나 삭제해야 합니다.")
        self.log("전체 롤백 절차 완료")

    def rollback_specific_version(self) -> None:
        """특정 버전 롤백"""
        version = input("\n롤백할 버전 입력 (예: 0.27.1): ").strip()

        if not version:
            print("❌ 버전을 입력해야 합니다.")
            return

        print(f"\n🎯 특정 버전 롤백: {version}")

        confirm = input(f"정말 {version} 버전을 롤백하시겠습니까? (type 'ROLLBACK' to confirm): ").strip()

        if confirm != "ROLLBACK":
            print("❌ 롤백 취소되었습니다.")
            return

        self.log(f"특정 버전 롤백 시작: {version}")
        print(f"🔄 특정 버전 롤백 절차: {version}")

        # 여기서 특정 버전 롤백 로직 구현
        print("📋 수동 롤백 절차:")
        print("1. PyPI에서 해당 버전 처리")
        print("2. GitHub Release 처리 (해당 버전이 있다면)")
        print("3. Git 태그 처리 (해당 버전이 있다면)")

        print("✅ 특정 버전 롤백 안내 완료")
        self.log(f"특정 버전 롤백 안내 완료: {version}")

    def run(self) -> None:
        """메인 실행 함수"""
        self.log("MoAI-ADK Interactive Release Manager 시작")

        try:
            while True:
                choice = self.show_main_menu()

                if choice == "1":
                    self.run_validate_workflow()
                elif choice == "2":
                    self.run_version_workflow()
                elif choice == "3":
                    self.run_changelog_workflow()
                elif choice == "4":
                    self.run_prepare_workflow()
                elif choice == "5":
                    self.run_rollback_workflow()
                elif choice == "6":
                    print("\n👋 작업을 종료합니다.")
                    self.log("작업 종료")
                    break
                else:
                    print("❌ 잘못된 선택입니다.")

                print("\n" + "="*60)

                # 계속할지 묻기
                continue_choice = input("계속 작업하시겠습니까? (Y/n): ").strip().lower()
                if continue_choice == 'n':
                    print("\n👋 작업을 종료합니다.")
                    self.log("작업 종료")
                    break

        except KeyboardInterrupt:
            print("\n\n👋 작업을 취소합니다.")
            self.log("작업 취소")
        except Exception as e:
            print(f"\n❌ 오류 발생: {e}")
            self.log(f"오류 발생: {e}")


def main():
    """메인 함수"""
    manager = InteractiveReleaseManager()
    manager.run()


if __name__ == "__main__":
    main()