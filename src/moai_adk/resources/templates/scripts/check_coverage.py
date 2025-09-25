#!/usr/bin/env python3
# @TASK:COVERAGE-CHECK-011
"""
MoAI-ADK Test Coverage Checker v0.1.12
테스트 커버리지 측정 및 임계값 검증

이 스크립트는 프로젝트의 테스트 커버리지를:
- pytest-cov를 사용하여 정확히 측정
- 최소 80% 임계값 검증
- 미커버 코드 위치 상세 리포트
- HTML 커버리지 리포트 생성
- 개발 가이드 5원칙 중 Testing 원칙 준수 확인
"""

import json
import re
import subprocess
import sys
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Any


@dataclass
class CoverageResult:
    """커버리지 결과 구조"""

    total_statements: int
    covered_statements: int
    coverage_percentage: float
    missing_lines: dict[str, list[int]]  # 파일별 미커버 라인
    branch_coverage: float | None = None


@dataclass
class FileCoverage:
    """파일별 커버리지 정보"""

    file_path: str
    statements: int
    missing: int
    coverage: float
    missing_lines: list[int]


class CoverageChecker:
    """테스트 커버리지 검사기"""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.moai_dir = project_root / ".moai"

        # 설정 로드
        self.config = self.load_coverage_config()

        # 결과 저장
        self.coverage_data = {}
        self.violations = []

    def load_coverage_config(self) -> dict[str, Any]:
        """커버리지 설정 로드"""

        default_config = {
            "min_coverage": 80.0,
            "min_branch_coverage": 75.0,
            "include_patterns": ["src/**/*.py", "app/**/*.py", "lib/**/*.py"],
            "exclude_patterns": [
                "tests/**/*.py",
                "test_*.py",
                "*_test.py",
                "setup.py",
                "conftest.py",
                "*/migrations/*",
                "*/venv/*",
                "*/node_modules/*",
            ],
            "fail_under": True,
            "show_missing": True,
            "skip_covered": False,
            "precision": 2,
        }

        # .moai/config.json에서 설정 읽기
        config_file = self.moai_dir / "config.json"
        if config_file.exists():
            try:
                moai_config = json.loads(config_file.read_text())
                coverage_config = moai_config.get("coverage", {})

                # 기본 설정 업데이트
                default_config.update(coverage_config)

            except Exception as error:
                print(f"Warning: Failed to load coverage config: {error}")

        return default_config

    def detect_test_framework(self) -> str:
        """사용 중인 테스트 프레임워크 감지"""

        # pytest 확인
        if self.has_pytest():
            return "pytest"

        # unittest 확인 (Python 기본)
        test_files = list(self.project_root.rglob("test_*.py"))
        if test_files:
            return "unittest"

        # Node.js 테스트 프레임워크 확인
        package_json = self.project_root / "package.json"
        if package_json.exists():
            try:
                pkg_data = json.loads(package_json.read_text())
                deps = {
                    **pkg_data.get("dependencies", {}),
                    **pkg_data.get("devDependencies", {}),
                }

                if "jest" in deps:
                    return "jest"
                elif "mocha" in deps:
                    return "mocha"

            except:
                pass

        return "unknown"

    def has_pytest(self) -> bool:
        """pytest 사용 가능 여부 확인"""
        try:
            result = subprocess.run(
                ["python", "-m", "pytest", "--version"], capture_output=True, timeout=10
            )
            return result.returncode == 0
        except:
            return False

    def has_pytest_cov(self) -> bool:
        """pytest-cov 사용 가능 여부 확인"""
        try:
            result = subprocess.run(
                ["python", "-c", 'import pytest_cov; print("available")'],
                capture_output=True,
                timeout=10,
            )
            return result.returncode == 0
        except:
            return False

    def run_pytest_coverage(self) -> CoverageResult:
        """pytest-cov로 커버리지 측정"""

        if not self.has_pytest_cov():
            raise RuntimeError(
                "pytest-cov not available. Install with: pip install pytest-cov"
            )

        # 커버리지 대상 디렉토리 결정
        src_dirs = []
        for pattern in self.config["include_patterns"]:
            base_dir = pattern.split("/")[0]
            if (self.project_root / base_dir).exists():
                src_dirs.append(base_dir)

        if not src_dirs:
            src_dirs = ["src"]  # 기본값

        # pytest 명령어 구성
        cmd = [
            "python",
            "-m",
            "pytest",
            "--cov=" + ",".join(src_dirs),
            "--cov-report=term-missing",
            "--cov-report=json:coverage.json",
            "--cov-report=html:htmlcov",
            "--cov-branch",
            f"--cov-fail-under={self.config['min_coverage']}",
            "-v",
        ]

        # 테스트 디렉토리 추가
        test_dirs = []
        for test_dir in ["tests", "test"]:
            if (self.project_root / test_dir).exists():
                test_dirs.append(test_dir)

        if test_dirs:
            cmd.extend(test_dirs)
        else:
            # 테스트 파일 패턴 추가
            cmd.extend(["-k", "test_"])

        print(f"Running: {' '.join(cmd)}")

        try:
            # pytest 실행
            result = subprocess.run(
                cmd,
                cwd=self.project_root,
                capture_output=True,
                text=True,
                timeout=300,  # 5분 타임아웃
            )

            print("STDOUT:", result.stdout[-1000:])  # 마지막 1000자만 출력
            if result.stderr:
                print("STDERR:", result.stderr[-500:])  # 에러 출력

            # coverage.json 파일에서 결과 파싱
            coverage_file = self.project_root / "coverage.json"
            if coverage_file.exists():
                coverage_data = json.loads(coverage_file.read_text())
                return self.parse_coverage_json(coverage_data)
            else:
                # 표준 출력에서 커버리지 정보 추출
                return self.parse_coverage_output(result.stdout)

        except subprocess.TimeoutExpired:
            raise RuntimeError("Coverage test timeout after 5 minutes")
        except Exception as error:
            raise RuntimeError(f"Coverage test failed: {error}")

    def parse_coverage_json(self, coverage_data: dict) -> CoverageResult:
        """coverage.json 파일 파싱"""

        totals = coverage_data.get("totals", {})
        files = coverage_data.get("files", {})

        total_statements = totals.get("num_statements", 0)
        covered_statements = totals.get("covered_lines", 0)
        coverage_percentage = totals.get("percent_covered", 0.0)

        # 파일별 미커버 라인 정보
        missing_lines = {}
        for file_path, file_data in files.items():
            if file_data.get("missing_lines"):
                missing_lines[file_path] = file_data["missing_lines"]

        return CoverageResult(
            total_statements=total_statements,
            covered_statements=covered_statements,
            coverage_percentage=coverage_percentage,
            missing_lines=missing_lines,
            branch_coverage=totals.get("percent_covered_display", None),
        )

    def parse_coverage_output(self, output: str) -> CoverageResult:
        """pytest 출력에서 커버리지 정보 추출"""

        # 커버리지 퍼센트 찾기
        coverage_pattern = r"TOTAL\s+\d+\s+\d+\s+(\d+)%"
        coverage_match = re.search(coverage_pattern, output)

        if coverage_match:
            coverage_percentage = float(coverage_match.group(1))
        else:
            coverage_percentage = 0.0

        # 파일별 미커버 라인 추출 (간단 버전)
        missing_lines = {}

        # 미커버 라인 패턴: filename.py    10     2    80%   5-6
        missing_pattern = r"(\S+\.py)\s+\d+\s+\d+\s+\d+%\s+([0-9,-]+)"
        missing_matches = re.findall(missing_pattern, output)

        for file_path, line_ranges in missing_matches:
            lines = self.parse_line_ranges(line_ranges)
            if lines:
                missing_lines[file_path] = lines

        return CoverageResult(
            total_statements=0,  # 정확한 수치는 JSON에서만 가능
            covered_statements=0,
            coverage_percentage=coverage_percentage,
            missing_lines=missing_lines,
        )

    def parse_line_ranges(self, line_ranges: str) -> list[int]:
        """라인 범위 문자열을 라인 번호 리스트로 변환"""
        lines = []

        for part in line_ranges.split(","):
            part = part.strip()
            if "-" in part:
                # 범위: 5-10
                start, end = part.split("-")
                lines.extend(range(int(start), int(end) + 1))
            else:
                # 단일 라인: 15
                lines.append(int(part))

        return lines

    def run_unittest_coverage(self) -> CoverageResult:
        """unittest + coverage.py로 커버리지 측정"""

        try:
            # coverage 모듈 확인
            subprocess.run(
                ["python", "-c", "import coverage"],
                check=True,
                capture_output=True,
                timeout=10,
            )

        except:
            raise RuntimeError(
                "coverage.py not available. Install with: pip install coverage"
            )

        # 커버리지 실행
        cmd = [
            "python",
            "-m",
            "coverage",
            "run",
            "-m",
            "unittest",
            "discover",
            "-s",
            "tests",
            "-p",
            "test_*.py",
        ]

        try:
            # 테스트 실행
            subprocess.run(cmd, check=True, cwd=self.project_root, timeout=180)

            # 리포트 생성
            report_result = subprocess.run(
                ["python", "-m", "coverage", "report", "--format=json"],
                capture_output=True,
                text=True,
                cwd=self.project_root,
                timeout=30,
            )

            if report_result.returncode == 0 and report_result.stdout:
                coverage_data = json.loads(report_result.stdout)
                return self.parse_coverage_json(coverage_data)

            # JSON 실패시 텍스트 리포트
            report_result = subprocess.run(
                ["python", "-m", "coverage", "report"],
                capture_output=True,
                text=True,
                cwd=self.project_root,
                timeout=30,
            )

            return self.parse_coverage_output(report_result.stdout)

        except subprocess.CalledProcessError as error:
            raise RuntimeError(f"unittest coverage failed: {error}")

    def analyze_coverage_quality(self, result: CoverageResult) -> dict[str, Any]:
        """커버리지 품질 분석"""

        analysis = {
            "overall_grade": "FAIL",
            "meets_minimum": False,
            "issues": [],
            "recommendations": [],
        }

        # 최소 커버리지 확인
        min_coverage = self.config["min_coverage"]
        if result.coverage_percentage >= min_coverage:
            analysis["meets_minimum"] = True
            analysis["overall_grade"] = "PASS"
        else:
            deficit = min_coverage - result.coverage_percentage
            analysis["issues"].append(
                f"Coverage {result.coverage_percentage:.1f}% below minimum {min_coverage}% (deficit: {deficit:.1f}%)"
            )

        # 브랜치 커버리지 확인
        if result.branch_coverage:
            min_branch = self.config["min_branch_coverage"]
            if result.branch_coverage < min_branch:
                analysis["issues"].append(
                    f"Branch coverage {result.branch_coverage:.1f}% below minimum {min_branch}%"
                )

        # 미커버 파일 분석
        if result.missing_lines:
            critical_files = []
            for file_path, lines in result.missing_lines.items():
                if len(lines) > 10:  # 10줄 이상 미커버
                    critical_files.append((file_path, len(lines)))

            if critical_files:
                analysis["issues"].append(
                    f"{len(critical_files)} files have >10 uncovered lines"
                )

        # 권장사항 생성
        if not analysis["meets_minimum"]:
            analysis["recommendations"].extend(
                [
                    f"Add tests to increase coverage by {deficit:.1f}%",
                    "Focus on critical business logic first",
                    "Use coverage report to identify specific uncovered lines",
                ]
            )

        if result.missing_lines:
            most_missing = max(result.missing_lines.items(), key=lambda x: len(x[1]))
            analysis["recommendations"].append(
                f"Start with {most_missing[0]} ({len(most_missing[1])} uncovered lines)"
            )

        return analysis

    def generate_report(
        self, result: CoverageResult, analysis: dict[str, Any]
    ) -> dict[str, Any]:
        """커버리지 리포트 생성"""

        return {
            "test_coverage": {
                "total_coverage": round(result.coverage_percentage, 2),
                "branch_coverage": round(result.branch_coverage or 0, 2),
                "target_coverage": self.config["min_coverage"],
                "meets_target": analysis["meets_minimum"],
                "grade": analysis["overall_grade"],
            },
            "statistics": {
                "total_statements": result.total_statements,
                "covered_statements": result.covered_statements,
                "missing_statements": result.total_statements
                - result.covered_statements,
                "files_with_missing_coverage": len(result.missing_lines),
            },
            "quality_analysis": analysis,
            "uncovered_files": [
                {
                    "file": file_path,
                    "missing_lines_count": len(lines),
                    "missing_lines": lines[:10],  # 처음 10개만
                }
                for file_path, lines in result.missing_lines.items()
            ],
            "recommendations": analysis["recommendations"],
            "scan_info": {
                "timestamp": datetime.now().isoformat(),
                "framework": self.detect_test_framework(),
                "config": self.config,
            },
        }

    def run_coverage_check(self) -> dict[str, Any]:
        """전체 커버리지 검사 실행"""

        print("🧪 Starting test coverage analysis...")

        framework = self.detect_test_framework()
        print(f"  Detected test framework: {framework}")

        try:
            if framework == "pytest":
                result = self.run_pytest_coverage()
            elif framework == "unittest":
                result = self.run_unittest_coverage()
            else:
                raise RuntimeError(f"Unsupported test framework: {framework}")

            print(f"  Coverage: {result.coverage_percentage:.1f}%")

            # 품질 분석
            analysis = self.analyze_coverage_quality(result)

            # 리포트 생성
            report = self.generate_report(result, analysis)

            return report

        except Exception as error:
            return {
                "test_coverage": {
                    "total_coverage": 0,
                    "meets_target": False,
                    "grade": "ERROR",
                },
                "error": str(error),
                "recommendations": ["Fix test configuration and try again"],
                "scan_info": {
                    "timestamp": datetime.now().isoformat(),
                    "framework": framework,
                },
            }


def main():
    """메인 실행 함수"""

    project_root = Path.cwd()

    # MoAI 프로젝트 확인
    if not (project_root / ".moai").exists():
        print("❌ This is not a MoAI-ADK project")
        sys.exit(1)

    try:
        # 커버리지 검사 실행
        checker = CoverageChecker(project_root)
        report = checker.run_coverage_check()

        # 결과 출력
        print("\n" + "=" * 60)
        print("🧪 TEST COVERAGE REPORT")
        print("=" * 60)

        coverage_info = report["test_coverage"]
        print(f"Overall Coverage: {coverage_info['total_coverage']:.2f}%")
        print(f"Target Coverage: {coverage_info['target_coverage']:.1f}%")
        print(f"Status: {'✅ PASS' if coverage_info['meets_target'] else '❌ FAIL'}")

        if "statistics" in report:
            stats = report["statistics"]
            print("\nStatistics:")
            print(f"  Total Statements: {stats['total_statements']}")
            print(f"  Covered: {stats['covered_statements']}")
            print(f"  Missing: {stats['missing_statements']}")
            print(
                f"  Files with Missing Coverage: {stats['files_with_missing_coverage']}"
            )

        # 품질 이슈
        if "quality_analysis" in report and report["quality_analysis"]["issues"]:
            print("\n⚠️  Issues:")
            for issue in report["quality_analysis"]["issues"]:
                print(f"  • {issue}")

        # 권장사항
        if report.get("recommendations"):
            print("\n💡 Recommendations:")
            for rec in report["recommendations"]:
                print(f"  • {rec}")

        # 미커버 파일 상위 5개
        if report.get("uncovered_files"):
            print("\n📋 Top Uncovered Files:")
            sorted_files = sorted(
                report["uncovered_files"],
                key=lambda x: x["missing_lines_count"],
                reverse=True,
            )
            for file_info in sorted_files[:5]:
                print(
                    f"  • {file_info['file']}: {file_info['missing_lines_count']} lines"
                )

        # 리포트 파일 저장
        report_file = project_root / ".moai" / "reports" / "coverage_report.json"
        report_file.parent.mkdir(parents=True, exist_ok=True)
        report_file.write_text(json.dumps(report, indent=2))

        print(f"\n📄 Detailed report saved to: {report_file}")

        # HTML 리포트 경로 안내
        html_report = project_root / "htmlcov" / "index.html"
        if html_report.exists():
            print(f"🌐 HTML report: {html_report}")

        # Exit code
        sys.exit(0 if coverage_info.get("meets_target", False) else 1)

    except Exception as error:
        print(f"❌ Coverage check failed: {error}")
        sys.exit(1)


if __name__ == "__main__":
    main()
