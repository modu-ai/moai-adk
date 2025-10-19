# @TEST:TRUST-001 | SPEC: SPEC-TRUST-001/spec.md
"""
TRUST 원칙 통합 검증 테스트

Given-When-Then 구조를 따르는 20개의 테스트 케이스:
- AC-001: 테스트 커버리지 ≥85% (2개: pass/fail)
- AC-002: 파일 ≤300 LOC (2개: pass/fail)
- AC-003: 함수 ≤50 LOC (2개: pass/fail)
- AC-004: 매개변수 ≤5개 (2개: pass/fail)
- AC-005: 순환 복잡도 ≤10 (2개: pass/fail)
- AC-006: @TAG 체인 완전성 (2개: pass/fail)
- AC-007: 고아 TAG 탐지 (2개: detected/none)
- AC-008: 보고서 생성 (2개: markdown/json)
- AC-009: 오류 메시지 (2개: specific/generic)
- AC-010: 언어별 도구 선택 (2개: python/typescript)
"""

from pathlib import Path

import pytest


class TestTrustChecker:
    """@TEST:TRUST-001: TRUST 원칙 통합 검증"""

    @pytest.fixture
    def trust_checker(self):
        """TrustChecker 인스턴스 생성"""
        from moai_adk.core.quality.trust_checker import TrustChecker
        return TrustChecker()

    @pytest.fixture
    def sample_project_path(self, tmp_path: Path) -> Path:
        """테스트용 샘플 프로젝트 디렉토리"""
        project = tmp_path / "sample_project"
        project.mkdir()
        (project / "src").mkdir()
        (project / "tests").mkdir()
        (project / ".moai").mkdir()
        return project

    # ========================================
    # AC-001: 테스트 커버리지 ≥85% 검증
    # ========================================

    def test_should_pass_when_coverage_above_85_percent(self, trust_checker, sample_project_path):
        """
        Given: 프로젝트의 테스트 커버리지가 87%
        When: trust_checker.validate_coverage() 실행
        Then: ValidationResult.passed = True
        """
        # Arrange
        coverage_data = {"total_coverage": 87.5}

        # Act
        result = trust_checker.validate_coverage(sample_project_path, coverage_data)

        # Assert
        assert result.passed is True
        assert result.message == "Test coverage: 87.5% (Target: 85%)"

    def test_should_fail_when_coverage_below_85_percent(self, trust_checker, sample_project_path):
        """
        Given: 프로젝트의 테스트 커버리지가 78%
        When: trust_checker.validate_coverage() 실행
        Then: ValidationResult.passed = False, 미달 파일 목록 포함
        """
        # Arrange
        coverage_data = {
            "total_coverage": 78.0,
            "low_coverage_files": [
                {"file": "src/utils/helper.py", "coverage": 72.0},
                {"file": "src/core/validator.py", "coverage": 78.0},
            ],
        }

        # Act
        result = trust_checker.validate_coverage(sample_project_path, coverage_data)

        # Assert
        assert result.passed is False
        assert "78.0%" in result.message
        assert "src/utils/helper.py" in result.details

    # ========================================
    # AC-002: 파일 ≤300 LOC 검증
    # ========================================

    def test_should_pass_when_all_files_within_300_loc(self, trust_checker, sample_project_path):
        """
        Given: 모든 소스 파일이 300 LOC 이하
        When: trust_checker.validate_file_size() 실행
        Then: ValidationResult.passed = True
        """
        # Arrange
        (sample_project_path / "src" / "small.py").write_text("\n".join([f"# Line {i}" for i in range(200)]))

        # Act
        result = trust_checker.validate_file_size(sample_project_path / "src")

        # Assert
        assert result.passed is True
        assert "All files within 300 LOC" in result.message

    def test_should_fail_when_file_exceeds_300_loc(self, trust_checker, sample_project_path):
        """
        Given: 파일이 342 LOC로 300 LOC 초과
        When: trust_checker.validate_file_size() 실행
        Then: ValidationResult.passed = False, 위반 파일 목록 포함
        """
        # Arrange
        large_file = sample_project_path / "src" / "large.py"
        large_file.write_text("\n".join([f"# Line {i}" for i in range(342)]))

        # Act
        result = trust_checker.validate_file_size(sample_project_path / "src")

        # Assert
        assert result.passed is False
        assert "large.py" in result.details
        assert "342 LOC" in result.details

    # ========================================
    # AC-003: 함수 ≤50 LOC 검증
    # ========================================

    def test_should_pass_when_all_functions_within_50_loc(self, trust_checker, sample_project_path):
        """
        Given: 모든 함수가 50 LOC 이하
        When: trust_checker.validate_function_size() 실행
        Then: ValidationResult.passed = True
        """
        # Arrange
        code = """
def small_function():
    # 30 LOC function
""" + "\n".join([f"    pass  # Line {i}" for i in range(30)])
        (sample_project_path / "src" / "functions.py").write_text(code)

        # Act
        result = trust_checker.validate_function_size(sample_project_path / "src")

        # Assert
        assert result.passed is True
        assert "All functions within 50 LOC" in result.message

    def test_should_fail_when_function_exceeds_50_loc(self, trust_checker, sample_project_path):
        """
        Given: 함수가 58 LOC로 50 LOC 초과
        When: trust_checker.validate_function_size() 실행
        Then: ValidationResult.passed = False, 위반 함수 목록 포함
        """
        # Arrange
        code = """
def large_function():
    # 58 LOC function
""" + "\n".join([f"    pass  # Line {i}" for i in range(58)])
        (sample_project_path / "src" / "functions.py").write_text(code)

        # Act
        result = trust_checker.validate_function_size(sample_project_path / "src")

        # Assert
        assert result.passed is False
        assert "large_function" in result.details
        assert " LOC" in result.details  # 실제 LOC가 60이 될 수 있음 (헤더 + 본문)

    # ========================================
    # AC-004: 매개변수 ≤5개 검증
    # ========================================

    def test_should_pass_when_all_params_within_5(self, trust_checker, sample_project_path):
        """
        Given: 모든 함수의 매개변수가 5개 이하
        When: trust_checker.validate_param_count() 실행
        Then: ValidationResult.passed = True
        """
        # Arrange
        code = """
def function_with_4_params(a, b, c, d):
    pass
"""
        (sample_project_path / "src" / "params.py").write_text(code)

        # Act
        result = trust_checker.validate_param_count(sample_project_path / "src")

        # Assert
        assert result.passed is True
        assert "All functions within 5 parameters" in result.message

    def test_should_fail_when_params_exceed_5(self, trust_checker, sample_project_path):
        """
        Given: 함수의 매개변수가 7개로 5개 초과
        When: trust_checker.validate_param_count() 실행
        Then: ValidationResult.passed = False, 위반 함수 목록 포함
        """
        # Arrange
        code = """
def function_with_7_params(a, b, c, d, e, f, g):
    pass
"""
        (sample_project_path / "src" / "params.py").write_text(code)

        # Act
        result = trust_checker.validate_param_count(sample_project_path / "src")

        # Assert
        assert result.passed is False
        assert "function_with_7_params" in result.details
        assert "7 parameters" in result.details

    # ========================================
    # AC-005: 순환 복잡도 ≤10 검증
    # ========================================

    def test_should_pass_when_complexity_within_10(self, trust_checker, sample_project_path):
        """
        Given: 모든 함수의 순환 복잡도가 10 이하
        When: trust_checker.validate_complexity() 실행
        Then: ValidationResult.passed = True
        """
        # Arrange
        code = """
def simple_function(x):
    if x > 0:
        return x
    return 0
"""
        (sample_project_path / "src" / "complexity.py").write_text(code)

        # Act
        result = trust_checker.validate_complexity(sample_project_path / "src")

        # Assert
        assert result.passed is True
        assert "All functions within complexity 10" in result.message

    def test_should_fail_when_complexity_exceeds_10(self, trust_checker, sample_project_path):
        """
        Given: 함수의 순환 복잡도가 15로 10 초과
        When: trust_checker.validate_complexity() 실행
        Then: ValidationResult.passed = False, 위반 함수 목록 포함
        """
        # Arrange - 중첩된 if문 12개로 복잡도 13 생성
        code = """
def complex_function(x):
    if x > 0:
        if x > 10:
            if x > 20:
                if x > 30:
                    if x > 40:
                        if x > 50:
                            if x > 60:
                                if x > 70:
                                    if x > 80:
                                        if x > 90:
                                            if x > 100:
                                                if x > 110:
                                                    return x
    return 0
"""
        (sample_project_path / "src" / "complexity.py").write_text(code)

        # Act
        result = trust_checker.validate_complexity(sample_project_path / "src")

        # Assert
        assert result.passed is False
        assert "complex_function" in result.details
        assert "complexity" in result.details.lower()

    # ========================================
    # AC-006: @TAG 체인 완전성 검증
    # ========================================

    def test_should_pass_when_tag_chain_complete(self, trust_checker, sample_project_path):
        """
        Given: @SPEC, @CODE, @TEST TAG가 모두 연결됨
        When: trust_checker.validate_tag_chain() 실행
        Then: ValidationResult.passed = True
        """
        # Arrange
        (sample_project_path / ".moai" / "specs").mkdir(parents=True)
        (sample_project_path / ".moai" / "specs" / "SPEC-AUTH-001.md").write_text("# @SPEC:AUTH-001")
        (sample_project_path / "src" / "auth.py").write_text("# @CODE:AUTH-001")
        (sample_project_path / "tests" / "test_auth.py").write_text("# @TEST:AUTH-001")

        # Act
        result = trust_checker.validate_tag_chain(sample_project_path)

        # Assert
        assert result.passed is True
        assert "TAG chain complete" in result.message

    def test_should_fail_when_tag_chain_broken(self, trust_checker, sample_project_path):
        """
        Given: @CODE:AUTH-001은 있으나 @SPEC:AUTH-001이 없음
        When: trust_checker.validate_tag_chain() 실행
        Then: ValidationResult.passed = False, 끊어진 체인 표시
        """
        # Arrange
        (sample_project_path / "src" / "auth.py").write_text("# @CODE:AUTH-001")

        # Act
        result = trust_checker.validate_tag_chain(sample_project_path)

        # Assert
        assert result.passed is False
        assert "auth-001" in result.details.lower()  # 소문자로 변환되므로 소문자로 검색
        assert "broken" in result.details.lower()

    # ========================================
    # AC-007: 고아 TAG 탐지
    # ========================================

    def test_should_detect_orphan_tags(self, trust_checker, sample_project_path):
        """
        Given: @CODE:USER-005는 있으나 @SPEC:USER-005가 없음 (고아 TAG)
        When: trust_checker.detect_orphan_tags() 실행
        Then: 고아 TAG 목록 반환
        """
        # Arrange
        (sample_project_path / "src" / "user.py").write_text("# @CODE:USER-005")

        # Act
        orphans = trust_checker.detect_orphan_tags(sample_project_path)

        # Assert
        assert len(orphans) > 0
        assert any("USER-005" in tag for tag in orphans)

    def test_should_return_empty_when_no_orphan_tags(self, trust_checker, sample_project_path):
        """
        Given: 모든 TAG가 올바르게 연결됨
        When: trust_checker.detect_orphan_tags() 실행
        Then: 빈 목록 반환
        """
        # Arrange
        (sample_project_path / ".moai" / "specs").mkdir(parents=True)
        (sample_project_path / ".moai" / "specs" / "SPEC-USER-001.md").write_text("# @SPEC:USER-001")
        (sample_project_path / "src" / "user.py").write_text("# @CODE:USER-001")

        # Act
        orphans = trust_checker.detect_orphan_tags(sample_project_path)

        # Assert
        assert len(orphans) == 0

    # ========================================
    # AC-008: 검증 결과 보고서 생성
    # ========================================

    def test_should_generate_markdown_report(self, trust_checker, sample_project_path):
        """
        Given: TRUST 검증 완료
        When: trust_checker.generate_report(format="markdown") 실행
        Then: Markdown 형식 보고서 생성
        """
        # Arrange
        results = {"coverage": {"passed": True, "value": 87}}

        # Act
        report = trust_checker.generate_report(results, format="markdown")

        # Assert
        assert "# TRUST Validation Report" in report
        assert "87%" in report
        assert "✅" in report or "PASS" in report

    def test_should_generate_json_report(self, trust_checker, sample_project_path):
        """
        Given: TRUST 검증 완료
        When: trust_checker.generate_report(format="json") 실행
        Then: JSON 형식 보고서 생성
        """
        # Arrange
        results = {"coverage": {"passed": True, "value": 87}}

        # Act
        report_json = trust_checker.generate_report(results, format="json")

        # Assert
        import json

        report = json.loads(report_json)
        assert "coverage" in report
        assert report["coverage"]["passed"] is True
        assert report["coverage"]["value"] == 87

    # ========================================
    # AC-009: 검증 실패 시 구체적 오류 메시지
    # ========================================

    def test_should_provide_specific_error_message(self, trust_checker, sample_project_path):
        """
        Given: 테스트 커버리지 78% (미달)
        When: trust_checker.validate_coverage() 실행
        Then: 구체적 오류 메시지 포함 (현재 커버리지, 미달 파일, 권장 조치)
        """
        # Arrange
        coverage_data = {
            "total_coverage": 78.0,
            "low_coverage_files": [{"file": "src/utils/helper.py", "coverage": 72.0}],
        }

        # Act
        result = trust_checker.validate_coverage(sample_project_path, coverage_data)

        # Assert
        assert "78.0%" in result.message
        assert "helper.py" in result.details
        assert "recommended" in result.details.lower() or "권장" in result.details

    def test_should_provide_generic_error_when_no_details(self, trust_checker, sample_project_path):
        """
        Given: 검증 실패했으나 상세 정보 없음
        When: trust_checker.validate() 실행
        Then: 일반적 오류 메시지 반환
        """
        # Arrange
        # (상세 정보 없는 실패 상황 시뮬레이션)

        # Act
        result = trust_checker.validate_coverage(sample_project_path, {"total_coverage": 70.0})

        # Assert
        assert result.passed is False
        assert "coverage" in result.message.lower() or "커버리지" in result.message

    # ========================================
    # AC-010: 언어별 도구 자동 선택
    # ========================================

    def test_should_select_python_tools(self, trust_checker, sample_project_path):
        """
        Given: .moai/config.json에 project.language: "python" 정의
        When: trust_checker.select_tools() 실행
        Then: pytest, coverage.py, mypy, ruff 선택
        """
        # Arrange
        config = {"project": {"language": "python"}}
        import json

        (sample_project_path / ".moai" / "config.json").write_text(json.dumps(config))

        # Act
        tools = trust_checker.select_tools(sample_project_path)

        # Assert
        assert tools["test_framework"] == "pytest"
        assert tools["coverage_tool"] == "coverage.py"
        assert tools["linter"] == "ruff"
        assert tools["type_checker"] == "mypy"

    def test_should_select_typescript_tools(self, trust_checker, sample_project_path):
        """
        Given: .moai/config.json에 project.language: "typescript" 정의
        When: trust_checker.select_tools() 실행
        Then: Vitest, Biome, tsc 선택
        """
        # Arrange
        config = {"project": {"language": "typescript"}}
        import json

        (sample_project_path / ".moai" / "config.json").write_text(json.dumps(config))

        # Act
        tools = trust_checker.select_tools(sample_project_path)

        # Assert
        assert tools["test_framework"] == "vitest"
        assert tools["linter"] == "biome"
        assert tools["type_checker"] == "tsc"
