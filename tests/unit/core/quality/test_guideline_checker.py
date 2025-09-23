"""
Tests for GuidelineChecker module - TDD GREEN phase.

@TEST:UNIT-GUIDELINES Test guideline compliance validation system
"""

import pytest
from pathlib import Path
from unittest.mock import Mock, patch, mock_open

from moai_adk.core.quality.guideline_checker import GuidelineChecker, GuidelineError


class TestGuidelineChecker:
    """Test suite for GuidelineChecker class - GREEN phase tests that should pass."""

    def test_function_length_should_detect_violations_over_50_loc(self):
        """
        @TEST:UNIT-FUNCTION-LENGTH
        Test that functions exceeding 50 LOC are detected as violations.

        Given: Python 파일에 51줄짜리 함수가 있음
        When: 함수 길이 검사를 실행함
        Then: 위반 사항이 감지되어야 함
        """
        # Given
        project_path = Path("/fake/project")
        checker = GuidelineChecker(project_path)
        test_file = Path("/fake/project/test_module.py")

        # Create a long function (51 lines) for testing
        long_function_code = """
def long_function():
    # Line 1
    x = 1
    # Line 3
    x += 1
    # Line 5
    x += 1
    # Continue for 51 lines...
""" + "\n".join([f"    # Line {i}" for i in range(6, 52)]) + "\n    return x"

        # Mock file parsing to return AST with long function
        with patch.object(checker, '_parse_python_file') as mock_parse:
            import ast
            mock_parse.return_value = ast.parse(long_function_code)

            # When
            violations = checker.check_function_length(test_file)

            # Then
            assert len(violations) == 1
            assert violations[0]["function_name"] == "long_function"
            assert violations[0]["line_count"] > 50

    def test_file_size_should_detect_violations_over_300_loc(self):
        """
        @TEST:UNIT-FILE-SIZE
        Test that files exceeding 300 LOC are detected as violations.

        Given: Python 파일이 301줄로 구성되어 있음
        When: 파일 크기 검사를 실행함
        Then: 위반 사항이 감지되어야 함
        """
        # Given
        project_path = Path("/fake/project")
        checker = GuidelineChecker(project_path)
        test_file = Path("/fake/project/large_module.py")

        # Mock file line counting to return 301 lines
        with patch.object(checker, '_count_file_lines') as mock_count:
            mock_count.return_value = 301

            # When
            result = checker.check_file_size(test_file)

            # Then
            assert result["violation"] is True
            assert result["line_count"] == 301
            assert result["line_count"] > 300

    def test_parameter_count_should_detect_violations_over_5_params(self):
        """
        @TEST:UNIT-PARAMETER-COUNT
        Test that functions with more than 5 parameters are detected as violations.

        Given: Python 함수가 6개의 매개변수를 가지고 있음
        When: 매개변수 개수 검사를 실행함
        Then: 위반 사항이 감지되어야 함
        """
        # Given
        project_path = Path("/fake/project")
        checker = GuidelineChecker(project_path)
        test_file = Path("/fake/project/many_params.py")

        # Create function with 6 parameters
        many_params_code = "def func_with_many_params(a, b, c, d, e, f):\n    return a + b + c + d + e + f"

        # Mock file parsing
        with patch.object(checker, '_parse_python_file') as mock_parse:
            import ast
            mock_parse.return_value = ast.parse(many_params_code)

            # When
            violations = checker.check_parameter_count(test_file)

            # Then
            assert len(violations) == 1
            assert violations[0]["function_name"] == "func_with_many_params"
            assert violations[0]["parameter_count"] == 6

    def test_complexity_should_detect_violations_over_10(self):
        """
        @TEST:UNIT-COMPLEXITY
        Test that functions with complexity > 10 are detected as violations.

        Given: Python 함수의 순환 복잡도가 11임
        When: 복잡도 검사를 실행함
        Then: 위반 사항이 감지되어야 함
        """
        # Given
        project_path = Path("/fake/project")
        checker = GuidelineChecker(project_path)
        test_file = Path("/fake/project/complex_function.py")

        # Create complex function with multiple if statements (complexity > 10)
        complex_code = """def complex_function(x):
    if x > 1:
        if x > 2:
            if x > 3:
                if x > 4:
                    if x > 5:
                        if x > 6:
                            if x > 7:
                                if x > 8:
                                    if x > 9:
                                        if x > 10:
                                            return x
    return 0"""

        with patch.object(checker, '_parse_python_file') as mock_parse:
            import ast
            mock_parse.return_value = ast.parse(complex_code)

            # When
            violations = checker.check_complexity(test_file)

            # Then
            assert len(violations) == 1
            assert violations[0]["function_name"] == "complex_function"
            assert violations[0]["complexity"] > 10

    def test_project_scan_should_find_all_violations(self):
        """
        @TEST:UNIT-PROJECT-SCAN
        Test that project scanning identifies all guideline violations.

        Given: 프로젝트에 다양한 가이드라인 위반 파일들이 있음
        When: 전체 프로젝트 스캔을 실행함
        Then: 모든 위반 사항이 발견되어야 함
        """
        # Given
        project_path = Path("/fake/project")
        checker = GuidelineChecker(project_path)

        # Mock Path methods using patch
        with patch('pathlib.Path.exists') as mock_exists, \
             patch('pathlib.Path.rglob') as mock_rglob:
            mock_exists.return_value = True
            mock_rglob.return_value = [Path("/fake/project/test.py")]

            # Mock all check methods to return empty results
            with patch.object(checker, 'check_function_length') as mock_func, \
                 patch.object(checker, 'check_file_size') as mock_file, \
                 patch.object(checker, 'check_parameter_count') as mock_params, \
                 patch.object(checker, 'check_complexity') as mock_complex:

                mock_func.return_value = []
                mock_file.return_value = {"violation": False}
                mock_params.return_value = []
                mock_complex.return_value = []

                # When
                violations = checker.scan_project()

                # Then
                assert "function_length" in violations
                assert "file_size" in violations
                assert "parameter_count" in violations
                assert "complexity" in violations

    def test_violation_report_should_provide_comprehensive_summary(self):
        """
        @TEST:UNIT-VIOLATION-REPORT
        Test that violation report provides comprehensive violation summary.

        Given: 여러 종류의 가이드라인 위반이 존재함
        When: 위반 리포트 생성을 요청함
        Then: 포괄적인 위반 요약 정보가 반환되어야 함
        """
        # Given
        project_path = Path("/fake/project")
        checker = GuidelineChecker(project_path)

        # Mock scan_project to return violations
        mock_violations = {
            "function_length": [{"function_name": "test"}],
            "file_size": [{"file_path": "/test.py"}],
            "parameter_count": [],
            "complexity": []
        }

        with patch.object(checker, 'scan_project') as mock_scan:
            mock_scan.return_value = mock_violations

            # When
            report = checker.generate_violation_report()

            # Then
            assert "violations" in report
            assert "summary" in report
            assert report["summary"]["total_violations"] == 2
            assert "violation_breakdown" in report["summary"]

    def test_single_file_validation_should_check_all_guidelines(self):
        """
        @TEST:UNIT-SINGLE-FILE
        Test that single file validation checks all guideline rules.

        Given: 하나의 Python 파일이 있음
        When: 단일 파일 검증을 실행함
        Then: 모든 가이드라인 규칙이 검사되어야 함
        """
        # Given
        project_path = Path("/fake/project")
        checker = GuidelineChecker(project_path)
        test_file = Path("/fake/project/single_file.py")

        # Mock all checks to return no violations
        with patch.object(checker, 'check_function_length') as mock_func, \
             patch.object(checker, 'check_file_size') as mock_file, \
             patch.object(checker, 'check_parameter_count') as mock_params, \
             patch.object(checker, 'check_complexity') as mock_complex:

            mock_func.return_value = []
            mock_file.return_value = {"violation": False}
            mock_params.return_value = []
            mock_complex.return_value = []

            # When
            is_valid = checker.validate_single_file(test_file)

            # Then
            assert is_valid is True
            mock_func.assert_called_once_with(test_file)
            mock_file.assert_called_once_with(test_file)
            mock_params.assert_called_once_with(test_file)
            mock_complex.assert_called_once_with(test_file)

    def test_guideline_checker_should_initialize_with_trust_limits(self):
        """
        @TEST:UNIT-INIT-LIMITS
        Test that GuidelineChecker initializes with TRUST 5 principle limits.

        Given: 프로젝트 경로가 주어짐
        When: GuidelineChecker를 초기화함
        Then: TRUST 5원칙 제한값들이 올바르게 설정되어야 함
        """
        # Given
        project_path = Path("/fake/project")

        # When
        checker = GuidelineChecker(project_path)

        # Then - 이 부분은 초기화에서 성공해야 함
        assert checker.project_path == project_path
        assert checker.max_function_lines == 50
        assert checker.max_file_lines == 300
        assert checker.max_parameters == 5
        assert checker.max_complexity == 10

    def test_function_length_violation_should_include_specific_details(self):
        """
        @TEST:UNIT-FUNCTION-DETAILS
        Test that function length violations include specific violation details.

        Given: 51줄의 함수가 있는 파일이 있음
        When: 함수 길이 검사를 실행함
        Then: 함수명, 줄 수, 시작 라인 정보가 포함되어야 함
        """
        # Given
        project_path = Path("/fake/project")
        checker = GuidelineChecker(project_path)
        test_file = Path("/fake/project/long_function.py")

        # Create a long function for testing
        long_function_code = """def long_function():
""" + "\n".join([f"    # Line {i}" for i in range(1, 52)]) + "\n    return True"

        with patch.object(checker, '_parse_python_file') as mock_parse:
            import ast
            mock_parse.return_value = ast.parse(long_function_code)

            # When
            violations = checker.check_function_length(test_file)

            # Then
            assert len(violations) > 0
            assert "function_name" in violations[0]
            assert "line_count" in violations[0]
            assert "start_line" in violations[0]
            assert violations[0]["line_count"] > 50
            assert violations[0]["function_name"] == "long_function"

    def test_file_size_violation_should_include_file_metrics(self):
        """
        @TEST:UNIT-FILE-METRICS
        Test that file size violations include comprehensive file metrics.

        Given: 301줄의 파일이 있음
        When: 파일 크기 검사를 실행함
        Then: 파일명, 총 줄 수, 위반 여부 정보가 포함되어야 함
        """
        # Given
        project_path = Path("/fake/project")
        checker = GuidelineChecker(project_path)
        test_file = Path("/fake/project/large_file.py")

        # Mock file line counting to return 301 lines
        with patch.object(checker, '_count_file_lines') as mock_count:
            mock_count.return_value = 301

            # When
            result = checker.check_file_size(test_file)

            # Then
            assert "file_path" in result
            assert "line_count" in result
            assert "violation" in result
            assert result["violation"] is True
            assert result["line_count"] > 300
            assert result["line_count"] == 301