#!/usr/bin/env python3
"""
개발 가이드 5원칙 검증 스크립트 (strict/relaxed 지원)
MoAI-ADK의 개발 가이드 5원칙 준수 여부를 자동 검증합니다.

동작 모드:
- 기본(완화): 현실적 기준으로 오탐을 줄임
- --strict: 기존 엄격 기준 유지(파일 수/커버리지 비율 기반 등)
"""

import json
import os
import subprocess
import sys
from pathlib import Path
from typing import Dict, List, Tuple

class 개발 가이드Checker:
    def __init__(self, project_root: str = ".", strict: bool = False):
        self.project_root = Path(project_root)
        self.config_path = self.project_root / ".moai" / "config.json"
        self.violations: List[Tuple[str, str, str]] = []  # (원칙, 위반내용, 권장사항)
        self.strict = strict

    def load_config(self) -> Dict:
        """프로젝트 설정 로드"""
        if not self.config_path.exists():
            print(f"❌ 설정 파일을 찾을 수 없습니다: {self.config_path}")
            return {}

        with open(self.config_path, 'r', encoding='utf-8') as f:
            return json.load(f)

    def check_simplicity(self, config: Dict) -> bool:
        """Simplicity 원칙 검증 (프로젝트 복잡도 ≤ 3개)

        - strict: src 내 *.py 파일(\_\_init__ 제외) 총량으로 판단(기존 방식)
        - relaxed: src 바로 하위의 유의미한 상위 모듈(디렉터리) 개수로 판단
        """
        max_projects = config.get('constitution', {}).get('principles', {}).get('simplicity', {}).get('max_projects', 3)
        src_dir = self.project_root / "src"
        if not src_dir.exists():
            return True

        if self.strict:
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

        # relaxed: 상위 모듈(디렉터리) 수로 판단
        top_modules = [d for d in src_dir.iterdir() if d.is_dir() and d.name not in {"__pycache__", "tests"}]
        # 모듈로 볼 수 있는 디렉터리만 카운트 (python 파일 포함)
        module_count = 0
        for d in top_modules:
            if any(p.suffix == ".py" for p in d.rglob("*.py")):
                module_count += 1
        if module_count > max_projects:
            self.violations.append((
                "Simplicity",
                f"상위 모듈 {module_count}개가 허용 한도 {max_projects}개를 초과",
                "상위 구조를 단순화하거나 모듈을 통합하세요"
            ))
            return False
        return True

    def check_architecture(self) -> bool:
        """Architecture 원칙 검증 (라이브러리 분리)

        - strict: 기대 계층 디렉터리 2개 이상 존재 요구(기존)
        - relaxed: 1개 이상 있으면 통과(없으면 위반)
        """
        src_dir = self.project_root / "src"
        if not src_dir.exists():
            return True
        expected_dirs = ["models", "services", "controllers", "utils"]
        found_dirs = [d.name for d in src_dir.iterdir() if d.is_dir()]
        overlap = len(set(expected_dirs) & set(found_dirs))
        if self.strict:
            if overlap < 2:
                self.violations.append((
                    "Architecture",
                    "라이브러리 분리 구조가 불명확함",
                    "models, services, controllers 등으로 계층을 분리하세요"
                ))
                return False
            return True
        # relaxed
        if overlap < 1:
            self.violations.append((
                "Architecture",
                "계층 분리 디렉터리가 감지되지 않음",
                "최소 한 개 계층 디렉터리(models/services/controllers/utils)부터 구성하세요"
            ))
            return False
        return True

    def check_testing(self) -> bool:
        """Testing 원칙 검증 (TDD, 커버리지 목표) - 언어 중립"""
        root = self.project_root
        tests_found = 0
        # Python
        tests_found += len(list(root.rglob("tests/test_*.py")))
        # JS/TS
        tests_found += len(list(root.rglob("**/*.test.js")))
        tests_found += len(list(root.rglob("**/*.spec.js")))
        tests_found += len(list(root.rglob("**/*.test.ts")))
        tests_found += len(list(root.rglob("**/*.spec.ts")))
        tests_found += 1 if any(p.is_dir() for p in root.rglob("**/__tests__")) else 0
        # Go
        tests_found += len(list(root.rglob("**/*_test.go")))
        # Rust
        tests_found += 1 if (root / "tests").exists() and list((root / "tests").rglob("*.rs")) else 0
        # Java
        tests_found += 1 if (root / "src" / "test").exists() else 0
        # C#
        tests_found += 1 if any("Tests" in str(p) for p in root.rglob("**/*.csproj")) else 0
        # C/C++ (CTest or tests dir with c/cpp)
        tests_found += 1 if (root / "CTestTestfile.cmake").exists() else 0
        tests_found += 1 if ((root / "tests").exists() and list((root / "tests").rglob("*.c")) or list((root / "tests").rglob("*.cpp"))) else 0

        if self.strict:
            return tests_found > 0  # 엄격 모드에서는 언어별 비율 검증은 도구 단계에 위임

        # relaxed: 하나라도 있으면 통과
        if tests_found == 0:
            self.violations.append((
                "Testing",
                "테스트 파일이 발견되지 않음",
                "언어에 맞는 테스트 파일을 추가하세요 (예: *_test.go, *.test.ts, tests/test_*.py 등)"
            ))
            return False
        return True

    def check_observability(self) -> bool:
        """Observability 원칙 검증 (구조화 로깅) - 언어 중립"""
        root = self.project_root
        patterns = [
            ("*.py", ["import logging", "logger"]) ,
            ("*.js", ["winston", "pino", "logger"]),
            ("*.ts", ["winston", "pino", "logger"]),
            ("*.go", ["\nlog.", "zap.", "zerolog."]),
            ("*.rs", ["log::", "tracing", "env_logger", "tracing_subscriber"]),
            ("*.java", ["org.slf4j", "java.util.logging", "log4j", "Logger "]),
            ("*.cs", ["Microsoft.Extensions.Logging", "ILogger<", "ILogger "]),
            ("*.cpp", ["spdlog", "glog", "BOOST_LOG"]),
            ("*.c", ["syslog", "glog"]) ,
        ]
        for glob, needles in patterns:
            for fp in root.rglob(f"**/{glob}"):
                try:
                    txt = fp.read_text(encoding="utf-8", errors="ignore")
                except Exception:
                    continue
                if any(n in txt for n in needles):
                    return True
        self.violations.append((
            "Observability",
            "구조화 로깅/로깅 프레임워크 사용 흔적이 없음",
            "언어에 맞는 로깅 프레임워크를 도입하세요 (예: Python logging, JS winston/pino, Go log/zap, Rust tracing/log, Java SLF4J, .NET ILogger, C++ spdlog 등)"
        ))
        return False

    def check_versioning(self) -> bool:
        """Versioning 원칙 검증 (시맨틱 버전 관리) - 언어 중립"""
        vf = [
            self.project_root / "pyproject.toml",
            self.project_root / "package.json",
            self.project_root / "go.mod",
            self.project_root / "Cargo.toml",
            self.project_root / "pom.xml",
        ]
        # *.csproj, *.sln 중 하나도 허용
        csproj = list(self.project_root.rglob("**/*.csproj"))
        sln = list(self.project_root.rglob("**/*.sln"))
        if csproj or sln:
            return True
        if any(p.exists() for p in vf):
            return True
        self.violations.append((
            "Versioning",
            "버전 관리 파일이 없음",
            "언어에 맞는 버전/의존성 파일을 구성하세요 (예: package.json, go.mod, Cargo.toml, pom.xml, *.csproj 등)"
        ))
        return False

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

        print("🏛️ 개발 가이드 5원칙 검증")
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
            print("🎉 모든 개발 가이드 원칙을 준수합니다!")
            return 0

        print("\n🔴 위반 사항 및 권장 조치:")
        for principle, violation, recommendation in self.violations:
            print(f"\n[{principle}]")
            print(f"  ❌ 문제: {violation}")
            print(f"  💡 권장: {recommendation}")

        print(f"\n⚖️ 개발 가이드 준수율: {(passed/total)*100:.1f}%")

        return 1 if len(self.violations) > 0 else 0

def main():
    import argparse
    parser = argparse.ArgumentParser(description="개발 가이드 5원칙 검증")
    parser.add_argument("--project-root", "-p", default=".", help="프로젝트 루트 경로")
    parser.add_argument("--strict", action="store_true", help="엄격 모드(기존 기준)")

    args = parser.parse_args()

    checker = 개발 가이드Checker(args.project_root, strict=args.strict)
    passed, total = checker.run_verification()

    return checker.generate_report(passed, total)

if __name__ == "__main__":
    sys.exit(main())
