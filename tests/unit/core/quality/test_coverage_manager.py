"""
Tests for CoverageManager module.

@TEST:UNIT-COVERAGE Test coverage management and validation
"""

import pytest
from pathlib import Path
from unittest.mock import Mock, patch

from moai_adk.core.quality import CoverageManager, CoverageError


class TestCoverageManager:
    """Test suite for CoverageManager class."""

    def test_coverage_should_meet_minimum_threshold(self):
        """
        @TEST:UNIT-THRESHOLD
        Test that coverage meets the minimum 85% threshold.

        Given: 프로젝트의 최소 커버리지 임계값이 85%로 설정됨
        When: 커버리지 측정을 실행함
        Then: 실제 커버리지가 85% 이상이어야 함
        """
        # Given
        project_path = Path("/fake/project")
        coverage_manager = CoverageManager(project_path)
        coverage_manager.set_minimum_threshold(85.0)

        # When
        actual_coverage = coverage_manager.measure_coverage()

        # Then
        assert actual_coverage >= 85.0

    def test_coverage_should_fail_below_threshold(self):
        """
        @TEST:UNIT-VALIDATION
        Test that coverage validation fails when below threshold.

        Given: 최소 커버리지가 85%로 설정되고 실제 커버리지가 70%임
        When: 커버리지 검증을 수행함
        Then: CoverageError가 발생해야 함
        """
        # Given
        project_path = Path("/fake/project")
        coverage_manager = CoverageManager(project_path)
        coverage_manager.set_minimum_threshold(85.0)

        # When & Then
        with pytest.raises(CoverageError) as exc_info:
            coverage_manager.validate_coverage(70.0)

        assert "70.0%" in str(exc_info.value)
        assert "85.0%" in str(exc_info.value)

    def test_coverage_should_generate_report(self):
        """
        @TEST:UNIT-REPORT
        Test that coverage report is generated correctly.

        Given: 프로젝트에 테스트 파일들이 존재함
        When: 커버리지 리포트 생성을 요청함
        Then: 구조화된 커버리지 리포트가 반환되어야 함
        """
        # Given
        project_path = Path("/fake/project")
        coverage_manager = CoverageManager(project_path)

        # When
        report = coverage_manager.generate_report()

        # Then
        assert "coverage_percentage" in report
        assert "total_lines" in report
        assert "covered_lines" in report
        assert "uncovered_files" in report
        assert isinstance(report["uncovered_files"], list)

    def test_coverage_should_identify_uncovered_lines(self):
        """
        @TEST:UNIT-ANALYSIS
        Test that uncovered lines are properly identified.

        Given: 프로젝트에 일부 라인이 커버되지 않은 파일이 있음
        When: 커버되지 않은 라인 분석을 요청함
        Then: 커버되지 않은 라인 정보가 반환되어야 함
        """
        # Given
        project_path = Path("/fake/project")
        coverage_manager = CoverageManager(project_path)

        # When
        uncovered_lines = coverage_manager.get_uncovered_lines()

        # Then
        assert isinstance(uncovered_lines, dict)

    def test_coverage_manager_should_initialize_with_config(self):
        """
        @TEST:UNIT-INIT
        Test that CoverageManager initializes with proper configuration.

        Given: 프로젝트 경로와 설정이 주어짐
        When: CoverageManager를 초기화함
        Then: 올바른 설정으로 초기화되어야 함
        """
        # Given
        project_path = Path("/fake/project")

        # When
        coverage_manager = CoverageManager(project_path)

        # Then
        assert coverage_manager.project_path == project_path
        assert coverage_manager.minimum_threshold == 85.0
        assert coverage_manager.exclude_patterns == []

    def test_coverage_should_handle_pytest_integration(self):
        """
        @TEST:UNIT-PYTEST
        Test integration with pytest coverage tool.

        Given: pytest와 pytest-cov가 설치되어 있음
        When: pytest 기반 커버리지 측정을 실행함
        Then: pytest-cov를 통한 정확한 커버리지 측정이 이루어져야 함
        """
        # Given
        project_path = Path("/fake/project")
        coverage_manager = CoverageManager(project_path)

        # When
        pytest_results = coverage_manager.run_pytest_coverage()

        # Then
        assert "total_coverage" in pytest_results
        assert "line_coverage" in pytest_results
        assert "branch_coverage" in pytest_results
        assert isinstance(pytest_results["total_coverage"], (int, float))

    def test_coverage_should_exclude_specified_files(self):
        """
        @TEST:UNIT-EXCLUDE
        Test that specified files are excluded from coverage.

        Given: 특정 파일들이 커버리지에서 제외되도록 설정됨
        When: 커버리지 측정을 실행함
        Then: 지정된 파일들이 커버리지 계산에서 제외되어야 함
        """
        # Given
        project_path = Path("/fake/project")
        coverage_manager = CoverageManager(project_path)
        exclude_patterns = ["*/tests/*", "*/migrations/*", "*/__pycache__/*"]

        # When
        coverage_manager.set_exclude_patterns(exclude_patterns)

        # Then
        assert coverage_manager.exclude_patterns == exclude_patterns
