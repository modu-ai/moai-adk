# @CODE:TRUST-001 | SPEC: SPEC-TRUST-001/spec.md | TEST: tests/unit/core/quality/test_trust_checker.py
"""
TRUST 원칙 통합 검증 시스템

TRUST 5원칙:
- T: Test First (테스트 커버리지 ≥85%)
- R: Readable (파일 ≤300 LOC, 함수 ≤50 LOC, 매개변수 ≤5개)
- U: Unified (타입 안전성)
- S: Secured (보안 취약점 스캔)
- T: Trackable (TAG 체인 무결성)
"""

import ast
import json
from pathlib import Path
from typing import Any

from moai_adk.core.quality.validators.base_validator import ValidationResult

# ========================================
# 상수 정의 (의도를 드러내는 이름)
# ========================================
MIN_TEST_COVERAGE_PERCENT = 85
MAX_FILE_LINES_OF_CODE = 300
MAX_FUNCTION_LINES_OF_CODE = 50
MAX_FUNCTION_PARAMETERS = 5
MAX_CYCLOMATIC_COMPLEXITY = 10

# 파일 인코딩
DEFAULT_FILE_ENCODING = "utf-8"

# TAG 접두사
TAG_PREFIX_SPEC = "@SPEC:"
TAG_PREFIX_CODE = "@CODE:"
TAG_PREFIX_TEST = "@TEST:"


class TrustChecker:
    """TRUST 원칙 통합 검증기"""

    def __init__(self):
        """TrustChecker 초기화"""
        self.results: dict[str, ValidationResult] = {}

    # ========================================
    # T: Test First - Coverage Validation
    # ========================================

    def validate_coverage(self, project_path: Path, coverage_data: dict[str, Any]) -> ValidationResult:
        """
        테스트 커버리지 검증 (≥85%)

        Args:
            project_path: 프로젝트 경로
            coverage_data: 커버리지 데이터 (total_coverage, low_coverage_files)

        Returns:
            ValidationResult: 검증 결과
        """
        total_coverage = coverage_data.get("total_coverage", 0)

        if total_coverage >= MIN_TEST_COVERAGE_PERCENT:
            return ValidationResult(
                passed=True, message=f"Test coverage: {total_coverage}% (Target: {MIN_TEST_COVERAGE_PERCENT}%)"
            )

        # 실패 시 상세 정보 생성
        low_files = coverage_data.get("low_coverage_files", [])
        details = f"Current coverage: {total_coverage}% (Target: {MIN_TEST_COVERAGE_PERCENT}%)\n"
        details += "Low coverage files:\n"
        for file_info in low_files:
            details += f"  - {file_info['file']}: {file_info['coverage']}%\n"
        details += "\nRecommended: Add more test cases to increase coverage."

        return ValidationResult(
            passed=False,
            message=f"Test coverage: {total_coverage}% (Target: {MIN_TEST_COVERAGE_PERCENT}%)",
            details=details,
        )

    # ========================================
    # R: Readable - Code Constraints
    # ========================================

    def validate_file_size(self, src_path: Path) -> ValidationResult:
        """
        파일 크기 검증 (≤300 LOC)

        Args:
            src_path: 소스 코드 디렉토리 경로

        Returns:
            ValidationResult: 검증 결과
        """
        # 입력 검증 (보안 강화)
        if not src_path.exists():
            return ValidationResult(passed=False, message=f"Source path does not exist: {src_path}", details="")

        if not src_path.is_dir():
            return ValidationResult(passed=False, message=f"Source path is not a directory: {src_path}", details="")

        violations = []

        for py_file in src_path.rglob("*.py"):
            # 가드절 적용 (가독성 향상)
            if py_file.name.startswith("test_"):
                continue

            try:
                lines = py_file.read_text(encoding="utf-8").splitlines()
                loc = len(lines)

                if loc > MAX_FILE_LINES_OF_CODE:
                    violations.append(f"{py_file.name}: {loc} LOC (Limit: {MAX_FILE_LINES_OF_CODE})")
            except (UnicodeDecodeError, PermissionError):
                # 보안: 파일 접근 오류 처리
                continue

        if not violations:
            return ValidationResult(passed=True, message="All files within 300 LOC")

        details = "Files exceeding 300 LOC:\n" + "\n".join(f"  - {v}" for v in violations)
        details += "\n\nRecommended: Refactor large files into smaller modules."

        return ValidationResult(passed=False, message=f"{len(violations)} files exceed 300 LOC", details=details)

    def validate_function_size(self, src_path: Path) -> ValidationResult:
        """
        함수 크기 검증 (≤50 LOC)

        Args:
            src_path: 소스 코드 디렉토리 경로

        Returns:
            ValidationResult: 검증 결과
        """
        violations = []

        for py_file in src_path.rglob("*.py"):
            if py_file.name.startswith("test_"):
                continue

            try:
                content = py_file.read_text()
                tree = ast.parse(content)
                lines = content.splitlines()

                for node in ast.walk(tree):
                    if isinstance(node, ast.FunctionDef):
                        # AST 라인 번호는 1-based
                        start_line = node.lineno
                        end_line = node.end_lineno if node.end_lineno else start_line  # type: ignore

                        # 실제 함수 라인 수 계산 (데코레이터 제외)
                        func_lines = lines[start_line - 1:end_line]
                        func_loc = len(func_lines)

                        if func_loc > MAX_FUNCTION_LINES_OF_CODE:
                            violations.append(
                                f"{py_file.name}::{node.name}(): {func_loc} LOC (Limit: {MAX_FUNCTION_LINES_OF_CODE})"
                            )
            except SyntaxError:
                continue

        if not violations:
            return ValidationResult(passed=True, message="All functions within 50 LOC")

        details = "Functions exceeding 50 LOC:\n" + "\n".join(f"  - {v}" for v in violations)
        details += "\n\nRecommended: Extract complex functions into smaller ones."

        return ValidationResult(passed=False, message=f"{len(violations)} functions exceed 50 LOC", details=details)

    def validate_param_count(self, src_path: Path) -> ValidationResult:
        """
        매개변수 개수 검증 (≤5개)

        Args:
            src_path: 소스 코드 디렉토리 경로

        Returns:
            ValidationResult: 검증 결과
        """
        violations = []

        for py_file in src_path.rglob("*.py"):
            if py_file.name.startswith("test_"):
                continue

            try:
                tree = ast.parse(py_file.read_text())
                for node in ast.walk(tree):
                    if isinstance(node, ast.FunctionDef):
                        param_count = len(node.args.args)
                        if param_count > MAX_FUNCTION_PARAMETERS:
                            violations.append(
                                f"{py_file.name}::{node.name}(): {param_count} parameters "
                                f"(Limit: {MAX_FUNCTION_PARAMETERS})"
                            )
            except SyntaxError:
                continue

        if not violations:
            return ValidationResult(passed=True, message="All functions within 5 parameters")

        details = "Functions exceeding 5 parameters:\n" + "\n".join(f"  - {v}" for v in violations)
        details += "\n\nRecommended: Use data classes or parameter objects."

        return ValidationResult(
            passed=False, message=f"{len(violations)} functions exceed 5 parameters", details=details
        )

    def validate_complexity(self, src_path: Path) -> ValidationResult:
        """
        순환 복잡도 검증 (≤10)

        Args:
            src_path: 소스 코드 디렉토리 경로

        Returns:
            ValidationResult: 검증 결과
        """
        violations = []

        for py_file in src_path.rglob("*.py"):
            if py_file.name.startswith("test_"):
                continue

            try:
                tree = ast.parse(py_file.read_text())
                for node in ast.walk(tree):
                    if isinstance(node, ast.FunctionDef):
                        complexity = self._calculate_complexity(node)
                        if complexity > MAX_CYCLOMATIC_COMPLEXITY:
                            violations.append(
                                f"{py_file.name}::{node.name}(): complexity {complexity} "
                                f"(Limit: {MAX_CYCLOMATIC_COMPLEXITY})"
                            )
            except SyntaxError:
                continue

        if not violations:
            return ValidationResult(passed=True, message="All functions within complexity 10")

        details = "Functions exceeding complexity 10:\n" + "\n".join(f"  - {v}" for v in violations)
        details += "\n\nRecommended: Simplify complex logic using guard clauses."

        return ValidationResult(
            passed=False, message=f"{len(violations)} functions exceed complexity 10", details=details
        )

    def _calculate_complexity(self, node: ast.FunctionDef) -> int:
        """
        순환 복잡도 계산 (McCabe complexity)

        Args:
            node: 함수 AST 노드

        Returns:
            int: 순환 복잡도
        """
        complexity = 1
        for child in ast.walk(node):
            # 분기문마다 +1
            if isinstance(child, (ast.If, ast.While, ast.For, ast.ExceptHandler, ast.With)):
                complexity += 1
            # and/or 연산자마다 +1
            elif isinstance(child, ast.BoolOp):
                complexity += len(child.values) - 1
            # elif는 이미 ast.If로 카운트되므로 별도 처리 불필요
        return complexity

    # ========================================
    # T: Trackable - TAG Chain Validation
    # ========================================

    def validate_tag_chain(self, project_path: Path) -> ValidationResult:
        """
        TAG 체인 완전성 검증

        Args:
            project_path: 프로젝트 경로

        Returns:
            ValidationResult: 검증 결과
        """
        specs_dir = project_path / ".moai" / "specs"
        src_dir = project_path / "src"

        # TAG 스캔
        spec_tags = self._scan_tags(specs_dir, "@SPEC:")
        code_tags = self._scan_tags(src_dir, "@CODE:")

        # 체인 검증
        broken_chains = []
        for code_tag in code_tags:
            tag_id = code_tag.split(":")[-1]
            if not any(tag_id in spec_tag for spec_tag in spec_tags):
                broken_chains.append(f"@CODE:{tag_id} (no @SPEC:{tag_id})")

        if not broken_chains:
            return ValidationResult(passed=True, message="TAG chain complete")

        details = "broken tag chains:\n" + "\n".join(f"  - {chain.lower()}" for chain in broken_chains)
        details += "\n\nrecommended: add missing spec documents or fix tag references."

        return ValidationResult(passed=False, message=f"{len(broken_chains)} broken TAG chains", details=details)

    def detect_orphan_tags(self, project_path: Path) -> list[str]:
        """
        고아 TAG 탐지

        Args:
            project_path: 프로젝트 경로

        Returns:
            list[str]: 고아 TAG 목록
        """
        specs_dir = project_path / ".moai" / "specs"
        src_dir = project_path / "src"

        spec_tags = self._scan_tags(specs_dir, "@SPEC:")
        code_tags = self._scan_tags(src_dir, "@CODE:")

        orphans = []
        for code_tag in code_tags:
            tag_id = code_tag.split(":")[-1]
            if not any(tag_id in spec_tag for spec_tag in spec_tags):
                orphans.append(code_tag)

        return orphans

    def _scan_tags(self, directory: Path, tag_prefix: str) -> list[str]:
        """
        디렉토리에서 TAG 스캔

        Args:
            directory: 스캔할 디렉토리
            tag_prefix: TAG 접두사 (예: "@SPEC:", "@CODE:")

        Returns:
            list[str]: 발견된 TAG 목록
        """
        if not directory.exists():
            return []

        tags = []
        for file in directory.rglob("*"):
            if file.is_file():
                try:
                    content = file.read_text()
                    for line in content.splitlines():
                        if tag_prefix in line:
                            tags.append(line.strip())
                except (UnicodeDecodeError, PermissionError):
                    continue

        return tags

    # ========================================
    # Report Generation
    # ========================================

    def generate_report(self, results: dict[str, Any], format: str = "markdown") -> str:
        """
        검증 결과 보고서 생성

        Args:
            results: 검증 결과 딕셔너리
            format: 보고서 형식 ("markdown" 또는 "json")

        Returns:
            str: 보고서 문자열
        """
        if format == "json":
            return json.dumps(results, indent=2)

        # Markdown 형식
        report = "# TRUST Validation Report\n\n"

        for category, result in results.items():
            status = "✅ PASS" if result.get("passed", False) else "❌ FAIL"
            value = result.get('value', 'N/A')
            # 숫자인 경우 % 기호 추가
            if isinstance(value, (int, float)):
                value_str = f"{value}%"
            else:
                value_str = str(value)

            report += f"## {category.upper()}\n"
            report += f"**Status**: {status}\n"
            report += f"**Value**: {value_str}\n\n"

        return report

    # ========================================
    # Tool Selection
    # ========================================

    def select_tools(self, project_path: Path) -> dict[str, str]:
        """
        언어별 도구 자동 선택

        Args:
            project_path: 프로젝트 경로

        Returns:
            dict[str, str]: 선택된 도구 딕셔너리
        """
        config_path = project_path / ".moai" / "config.json"
        if not config_path.exists():
            return {
                "test_framework": "pytest",
                "coverage_tool": "coverage.py",
                "linter": "ruff",
                "type_checker": "mypy",
            }

        config = json.loads(config_path.read_text())
        language = config.get("project", {}).get("language", "python")

        if language == "python":
            return {
                "test_framework": "pytest",
                "coverage_tool": "coverage.py",
                "linter": "ruff",
                "type_checker": "mypy",
            }
        elif language == "typescript":
            return {
                "test_framework": "vitest",
                "linter": "biome",
                "type_checker": "tsc",
            }

        # 기본값 (Python)
        return {
            "test_framework": "pytest",
            "coverage_tool": "coverage.py",
            "linter": "ruff",
            "type_checker": "mypy",
        }
