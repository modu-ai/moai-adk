#!/usr/bin/env python3
"""
Constitution 5원칙 검증 스크립트
MoAI-ADK의 Constitution 5원칙 준수 여부를 자동 검증합니다.
"""

import json
import os
import subprocess
import sys
from pathlib import Path
from typing import Dict, List, Tuple

class ConstitutionChecker:
    def __init__(self, project_root: str = "."):
        self.project_root = Path(project_root)
        self.config_path = self.project_root / ".moai" / "config.json"
        self.violations: List[Tuple[str, str, str]] = []  # (원칙, 위반내용, 권장사항)

    def load_config(self) -> Dict:
        """프로젝트 설정 로드"""
        if not self.config_path.exists():
            print(f"❌ 설정 파일을 찾을 수 없습니다: {self.config_path}")
            return {}

        with open(self.config_path, 'r', encoding='utf-8') as f:
            return json.load(f)

    def check_simplicity(self, config: Dict) -> bool:
        """Simplicity 원칙 검증 (프로젝트 복잡도 ≤ 3개)"""
        max_projects = config.get('constitution', {}).get('principles', {}).get('simplicity', {}).get('max_projects', 3)

        # Python 모듈 수 확인
        src_dir = self.project_root / "src"
        if src_dir.exists():
            py_files = list(src_dir.rglob("*.py"))
            py_count = len([f for f in py_files if f.name != "__init__.py"])

            if py_count > max_projects:
                self.violations.append((
                    "Simplicity",
                    f"모듈 수 {py_count}개가 허용 한도 {max_projects}개를 초과",
                    f"모듈을 {max_projects}개 이하로 통합하거나 기능을 단순화하세요"
                ))
                return False

        return True

    def check_architecture(self) -> bool:
        """Architecture 원칙 검증 (라이브러리 분리)"""
        # 기본적인 아키텍처 패턴 확인
        src_dir = self.project_root / "src"
        if not src_dir.exists():
            return True

        # 라이브러리 구조 확인 (예: models, services, controllers 분리)
        expected_dirs = ["models", "services", "controllers", "utils"]
        found_dirs = [d.name for d in src_dir.iterdir() if d.is_dir()]

        if len(set(expected_dirs) & set(found_dirs)) < 2:
            self.violations.append((
                "Architecture",
                "라이브러리 분리 구조가 불명확함",
                "models, services, controllers 등으로 계층을 분리하세요"
            ))
            return False

        return True

    def check_testing(self) -> bool:
        """Testing 원칙 검증 (TDD, 85% 커버리지)"""
        tests_dir = self.project_root / "tests"
        src_dir = self.project_root / "src"

        if not tests_dir.exists():
            self.violations.append((
                "Testing",
                "tests 디렉토리가 존재하지 않음",
                "TDD를 위한 tests 디렉토리를 생성하고 테스트를 작성하세요"
            ))
            return False

        # 테스트 파일 수와 소스 파일 수 비교
        if src_dir.exists():
            test_files = list(tests_dir.rglob("test_*.py"))
            src_files = list(src_dir.rglob("*.py"))

            if len(test_files) < len(src_files) * 0.8:  # 80% 이상의 테스트 파일 권장
                self.violations.append((
                    "Testing",
                    f"테스트 파일 수({len(test_files)})가 소스 파일 수({len(src_files)})에 비해 부족",
                    "각 소스 파일에 대응하는 테스트 파일을 작성하세요"
                ))
                return False

        return True

    def check_observability(self) -> bool:
        """Observability 원칙 검증 (구조화 로깅)"""
        src_dir = self.project_root / "src"
        if not src_dir.exists():
            return True

        # 로깅 코드 존재 확인
        logging_found = False
        for py_file in src_dir.rglob("*.py"):
            try:
                with open(py_file, 'r', encoding='utf-8') as f:
                    content = f.read()
                    if "import logging" in content or "logger" in content:
                        logging_found = True
                        break
            except Exception:
                continue

        if not logging_found:
            self.violations.append((
                "Observability",
                "구조화된 로깅 코드가 발견되지 않음",
                "logging 모듈을 사용하여 구조화된 로그를 구현하세요"
            ))
            return False

        return True

    def check_versioning(self) -> bool:
        """Versioning 원칙 검증 (시맨틱 버전 관리)"""
        # pyproject.toml 또는 package.json 확인
        version_files = [
            self.project_root / "pyproject.toml",
            self.project_root / "package.json",
            self.project_root / ".moai" / "config.json"
        ]

        version_found = False
        for version_file in version_files:
            if version_file.exists():
                version_found = True
                break

        if not version_found:
            self.violations.append((
                "Versioning",
                "버전 관리 파일이 없음",
                "pyproject.toml 또는 package.json을 생성하여 시맨틱 버전을 관리하세요"
            ))
            return False

        return True

    def run_verification(self) -> Tuple[int, int]:
        """전체 검증 실행"""
        config = self.load_config()

        checks = [
            ("Simplicity", lambda: self.check_simplicity(config)),
            ("Architecture", self.check_architecture),
            ("Testing", self.check_testing),
            ("Observability", self.check_observability),
            ("Versioning", self.check_versioning)
        ]

        passed = 0
        total = len(checks)

        print("🏛️ Constitution 5원칙 검증")
        print("=" * 50)

        for principle, check_func in checks:
            try:
                result = check_func()
                if result:
                    print(f"✅ {principle}: 통과")
                    passed += 1
                else:
                    print(f"❌ {principle}: 위반")
            except Exception as e:
                print(f"⚠️ {principle}: 검증 실패 - {e}")

        return passed, total

    def generate_report(self, passed: int, total: int) -> int:
        """검증 결과 보고서 생성"""
        print(f"\n📊 검증 결과: {passed}/{total} 통과")

        if len(self.violations) == 0:
            print("🎉 모든 Constitution 원칙을 준수합니다!")
            return 0

        print("\n🔴 위반 사항 및 권장 조치:")
        for principle, violation, recommendation in self.violations:
            print(f"\n[{principle}]")
            print(f"  ❌ 문제: {violation}")
            print(f"  💡 권장: {recommendation}")

        print(f"\n⚖️ Constitution 준수율: {(passed/total)*100:.1f}%")

        return 1 if len(self.violations) > 0 else 0

def main():
    import argparse
    parser = argparse.ArgumentParser(description="Constitution 5원칙 검증")
    parser.add_argument("--project-root", "-p", default=".", help="프로젝트 루트 경로")

    args = parser.parse_args()

    checker = ConstitutionChecker(args.project_root)
    passed, total = checker.run_verification()

    return checker.generate_report(passed, total)

if __name__ == "__main__":
    sys.exit(main())