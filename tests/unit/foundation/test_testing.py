"""
Unit tests for moai_adk.foundation.testing module.

Tests cover:
- TestStatus enum
- TestResult dataclass
- CoverageReport dataclass
- All manager and analyzer classes
- Utility functions
"""

import json
from datetime import datetime
from pathlib import Path
from unittest import mock

import pytest

from moai_adk.foundation.testing import (
    CoverageAnalyzer,
    CoverageReport,
    QualityGateEngine,
    TestAutomationOrchestrator,
    TestDataManager,
    TestReportingSpecialist,
    TestResult,
    TestStatus,
    TestingFrameworkManager,
    TestingMetricsCollector,
    export_test_results,
    generate_test_report,
    validate_test_configuration,
)


class TestTestStatus:
    """Test TestStatus enumeration."""

    def test_status_values(self):
        """Test all status values are defined."""
        assert TestStatus.PASSED.value == "passed"
        assert TestStatus.FAILED.value == "failed"
        assert TestStatus.SKIPPED.value == "skipped"
        assert TestStatus.RUNNING.value == "running"

    def test_status_members(self):
        """Test status members exist."""
        assert hasattr(TestStatus, "PASSED")
        assert hasattr(TestStatus, "FAILED")
        assert hasattr(TestStatus, "SKIPPED")
        assert hasattr(TestStatus, "RUNNING")


class TestTestResult:
    """Test TestResult dataclass."""

    def test_create_minimal_result(self):
        """Test creating minimal test result."""
        result = TestResult(name="test_basic", status=TestStatus.PASSED, duration=0.5)
        assert result.name == "test_basic"
        assert result.status == TestStatus.PASSED
        assert result.duration == 0.5
        assert result.error_message is None
        assert result.metadata is None

    def test_create_full_result(self):
        """Test creating full test result."""
        metadata = {"module": "test_core", "line": 42}
        result = TestResult(
            name="test_failure",
            status=TestStatus.FAILED,
            duration=1.2,
            error_message="AssertionError: expected True",
            metadata=metadata,
        )
        assert result.error_message == "AssertionError: expected True"
        assert result.metadata == metadata


class TestCoverageReport:
    """Test CoverageReport dataclass."""

    def test_create_coverage_report(self):
        """Test creating coverage report."""
        report = CoverageReport(
            total_lines=1000,
            covered_lines=850,
            percentage=85.0,
            branches=200,
            covered_branches=160,
            branch_percentage=80.0,
            by_file={"module.py": {"lines": 100, "covered": 90}},
            by_module={"core": {"percentage": 85.0}},
        )
        assert report.total_lines == 1000
        assert report.covered_lines == 850
        assert report.percentage == 85.0


class TestTestingFrameworkManager:
    """Test TestingFrameworkManager class."""

    def test_init_with_default_config(self):
        """Test initialization with default config."""
        manager = TestingFrameworkManager()
        assert manager.config == {}

    def test_init_with_custom_config(self):
        """Test initialization with custom config."""
        config = {"pytest": {"markers": ["unit", "integration"]}}
        manager = TestingFrameworkManager(config=config)
        assert manager.config == config

    def test_configure_pytest_environment(self):
        """Test pytest environment configuration."""
        manager = TestingFrameworkManager()
        config = manager.configure_pytest_environment()

        assert "fixtures" in config
        assert "markers" in config
        assert "options" in config
        assert "testpaths" in config
        assert "addopts" in config

    def test_setup_jest_environment(self):
        """Test Jest environment setup."""
        manager = TestingFrameworkManager()
        config = manager.setup_jest_environment()

        assert "jest_config" in config
        assert "npm_scripts" in config
        assert "package_config" in config
        assert config["jest_config"]["testEnvironment"] == "node"

    def test_configure_playwright_e2e(self):
        """Test Playwright E2E configuration."""
        manager = TestingFrameworkManager()
        config = manager.configure_playwright_e2e()

        assert "playwright_config" in config
        assert "test_config" in config
        assert "browsers" in config
        assert config["playwright_config"]["testDir"] == "tests/e2e"

    def test_setup_api_testing(self):
        """Test API testing setup."""
        manager = TestingFrameworkManager()
        config = manager.setup_api_testing()

        assert "rest_assured_config" in config
        assert "test_data" in config
        assert "assertion_helpers" in config
        assert config["rest_assured_config"]["base_url"] == "http://localhost:8080"


class TestQualityGateEngine:
    """Test QualityGateEngine class."""

    def test_init_with_default_config(self):
        """Test initialization with default config."""
        engine = QualityGateEngine()
        assert engine.config == {}
        assert engine.quality_thresholds["max_complexity"] == 10
        assert engine.quality_thresholds["min_coverage"] == 85

    def test_init_with_custom_config(self):
        """Test initialization with custom config."""
        config = {"enabled": True}
        engine = QualityGateEngine(config=config)
        assert engine.config == config

    def test_setup_code_quality_checks(self):
        """Test code quality checks setup."""
        engine = QualityGateEngine()
        config = engine.setup_code_quality_checks()

        assert "linters" in config
        assert "formatters" in config
        assert "rules" in config
        assert "thresholds" in config
        assert "pylint" in config["linters"]

    def test_configure_security_scanning(self):
        """Test security scanning configuration."""
        engine = QualityGateEngine()
        config = engine.configure_security_scanning()

        assert "scan_tools" in config
        assert "vulnerability_levels" in config
        assert "exclusions" in config
        assert "reporting" in config
        assert "bandit" in config["scan_tools"]

    def test_setup_performance_tests(self):
        """Test performance tests setup."""
        engine = QualityGateEngine()
        config = engine.setup_performance_tests()

        assert "benchmarks" in config
        assert "thresholds" in config
        assert "tools" in config
        assert "scenarios" in config


class TestCoverageAnalyzer:
    """Test CoverageAnalyzer class."""

    def test_init_with_default_config(self):
        """Test initialization with default config."""
        analyzer = CoverageAnalyzer()
        assert analyzer.config == {}
        assert analyzer.coverage_thresholds["min_line_coverage"] == 85.0

    def test_analyze_code_coverage(self):
        """Test code coverage analysis."""
        analyzer = CoverageAnalyzer()
        report = analyzer.analyze_code_coverage()

        assert "summary" in report
        assert "details" in report
        assert "recommendations" in report
        assert "trends" in report
        assert report["summary"]["percentage"] == 85.0

    def test_generate_coverage_badges(self):
        """Test coverage badge generation."""
        analyzer = CoverageAnalyzer()
        badges = analyzer.generate_coverage_badges()

        assert "badges" in badges
        assert "badge_config" in badges
        assert "line_coverage" in badges["badges"]

    def test_track_coverage_trends(self):
        """Test coverage trend tracking."""
        analyzer = CoverageAnalyzer()
        trends = analyzer.track_coverage_trends()

        assert "historical_data" in trends
        assert "trend_analysis" in trends
        assert "2024-01" in trends["historical_data"]

    def test_set_coverage_thresholds(self):
        """Test setting coverage thresholds."""
        analyzer = CoverageAnalyzer()
        new_thresholds = {"min_line_coverage": 90.0, "min_branch_coverage": 85.0}
        result = analyzer.set_coverage_thresholds(new_thresholds)

        assert result["thresholds_set"] is True
        assert analyzer.coverage_thresholds["min_line_coverage"] == 90.0

    def test_enforce_quality_gates(self):
        """Test quality gate enforcement."""
        analyzer = CoverageAnalyzer()
        result = analyzer.enforce_quality_gates()

        assert "status" in result
        assert "passed_gates" in result
        assert "failed_gates" in result
        assert "details" in result

    def test_collect_test_metrics(self):
        """Test metrics collection."""
        analyzer = CoverageAnalyzer()
        metrics = analyzer.collect_test_metrics()

        assert "execution_metrics" in metrics
        assert "quality_metrics" in metrics
        assert "performance_metrics" in metrics


class TestTestAutomationOrchestrator:
    """Test TestAutomationOrchestrator class."""

    def test_init_with_default_config(self):
        """Test initialization."""
        orchestrator = TestAutomationOrchestrator()
        assert orchestrator.config == {}

    def test_setup_ci_pipeline(self):
        """Test CI pipeline setup."""
        orchestrator = TestAutomationOrchestrator()
        config = orchestrator.setup_ci_pipeline()

        assert "pipeline_config" in config
        assert "triggers" in config
        assert "jobs" in config
        assert "artifacts" in config

    def test_configure_parallel_execution(self):
        """Test parallel execution configuration."""
        orchestrator = TestAutomationOrchestrator()
        config = orchestrator.configure_parallel_execution()

        assert "execution_strategy" in config
        assert "workers" in config
        assert "distribution" in config
        assert "isolation" in config

    def test_manage_test_data(self):
        """Test test data management."""
        orchestrator = TestAutomationOrchestrator()
        config = orchestrator.manage_test_data()

        assert "data_sources" in config
        assert "fixtures" in config
        assert "cleanup" in config
        assert "validation" in config

    def test_orchestrate_test_runs(self):
        """Test test run orchestration."""
        orchestrator = TestAutomationOrchestrator()
        config = orchestrator.orchestrate_test_runs()

        assert "test_runs" in config
        assert "orchestration_config" in config
        assert "unit_tests" in config["test_runs"]


class TestTestReportingSpecialist:
    """Test TestReportingSpecialist class."""

    def test_init_with_default_config(self):
        """Test initialization."""
        specialist = TestReportingSpecialist()
        assert specialist.config == {}

    def test_generate_test_reports(self):
        """Test report generation."""
        specialist = TestReportingSpecialist()
        report = specialist.generate_test_reports()

        assert "summary" in report
        assert "details" in report
        assert "trends" in report
        assert "recommendations" in report
        assert report["summary"]["total_tests"] == 1500

    def test_create_quality_dashboard(self):
        """Test quality dashboard creation."""
        specialist = TestReportingSpecialist()
        dashboard = specialist.create_quality_dashboard()

        assert "widgets" in dashboard
        assert "data_sources" in dashboard
        assert "refresh_interval" in dashboard
        assert "filters" in dashboard

    def test_analyze_test_failures(self):
        """Test failure analysis."""
        specialist = TestReportingSpecialist()
        analysis = specialist.analyze_test_failures()

        assert "failure_summary" in analysis
        assert "root_causes" in analysis
        assert "patterns" in analysis
        assert "recommendations" in analysis

    def test_track_test_trends(self):
        """Test trend tracking."""
        specialist = TestReportingSpecialist()
        trends = specialist.track_test_trends()

        assert "historical_data" in trends
        assert "trend_analysis" in trends
        assert "predictions" in trends
        assert "insights" in trends


class TestTestDataManager:
    """Test TestDataManager class."""

    def test_init_with_default_config(self):
        """Test initialization."""
        manager = TestDataManager()
        assert manager.config == {}

    def test_create_test_datasets(self):
        """Test test dataset creation."""
        manager = TestDataManager()
        datasets = manager.create_test_datasets()

        assert "test_datasets" in datasets
        assert "data_validation" in datasets
        assert "data_management" in datasets
        assert "user_data" in datasets["test_datasets"]

    def test_manage_test_fixtures(self):
        """Test fixture management."""
        manager = TestDataManager()
        fixtures = manager.manage_test_fixtures()

        assert "fixture_config" in fixtures
        assert "fixture_lifecycle" in fixtures
        assert "database_fixtures" in fixtures["fixture_config"]

    def test_setup_test_environments(self):
        """Test environment setup."""
        manager = TestDataManager()
        envs = manager.setup_test_environments()

        assert "environments" in envs
        assert "environment_setup" in envs
        assert "environment_isolation" in envs
        assert "development" in envs["environments"]

    def test_cleanup_test_artifacts(self):
        """Test artifact cleanup."""
        manager = TestDataManager()
        cleanup = manager.cleanup_test_artifacts()

        assert "cleanup_strategies" in cleanup
        assert "cleanup_schedule" in cleanup
        assert "cleanup_metrics" in cleanup


class TestTestingMetricsCollector:
    """Test TestingMetricsCollector class."""

    def test_init_with_default_config(self):
        """Test initialization."""
        collector = TestingMetricsCollector()
        assert collector.config == {}
        assert collector.metrics_history == []

    def test_collect_test_metrics(self):
        """Test metrics collection."""
        collector = TestingMetricsCollector()
        metrics = collector.collect_test_metrics()

        assert "execution_metrics" in metrics
        assert "quality_metrics" in metrics
        assert "performance_metrics" in metrics
        assert "team_metrics" in metrics

    def test_calculate_quality_scores(self):
        """Test quality score calculation."""
        collector = TestingMetricsCollector()
        scores = collector.calculate_quality_scores()

        assert "weights" in scores
        assert "raw_scores" in scores
        assert "weighted_scores" in scores
        assert "overall_score" in scores
        assert "grade" in scores
        assert "recommendations" in scores

    def test_track_test_efficiency(self):
        """Test efficiency tracking."""
        collector = TestingMetricsCollector()
        efficiency = collector.track_test_efficiency()

        assert "efficiency_metrics" in efficiency
        assert "productivity_metrics" in efficiency
        assert "efficiency_trends" in efficiency
        assert "efficiency_benchmarks" in efficiency

    def test_generate_test_analytics(self):
        """Test analytics generation."""
        collector = TestingMetricsCollector()
        analytics = collector.generate_test_analytics()

        assert "executive_summary" in analytics
        assert "detailed_analytics" in analytics
        assert "actionable_insights" in analytics
        assert "future_predictions" in analytics


class TestUtilityFunctions:
    """Test utility functions."""

    def test_generate_test_report_empty(self):
        """Test report generation with empty results."""
        report = generate_test_report([])
        assert report["summary"]["total_tests"] == 0
        assert report["summary"]["passed_tests"] == 0
        assert report["summary"]["pass_rate"] == 0

    def test_generate_test_report_with_results(self):
        """Test report generation with results."""
        results = [
            TestResult("test1", TestStatus.PASSED, 0.5),
            TestResult("test2", TestStatus.FAILED, 1.0),
            TestResult("test3", TestStatus.SKIPPED, 0.0),
        ]
        report = generate_test_report(results)

        assert report["summary"]["total_tests"] == 3
        assert report["summary"]["passed_tests"] == 1
        assert report["summary"]["failed_tests"] == 1
        assert report["summary"]["skipped_tests"] == 1
        assert report["summary"]["pass_rate"] == pytest.approx(33.33, abs=0.01)

    def test_export_test_results_json(self):
        """Test exporting results to JSON."""
        results = {"summary": {"total": 10, "passed": 8}}
        json_str = export_test_results(results, format="json")

        parsed = json.loads(json_str)
        assert parsed["summary"]["total"] == 10

    def test_export_test_results_xml(self):
        """Test exporting results to XML."""
        results = {
            "summary": {"total": 2, "passed": 1},
            "details": [
                {"name": "test1", "status": "PASSED", "duration": 0.5},
                {"name": "test2", "status": "FAILED", "duration": 1.0},
            ],
        }
        xml_str = export_test_results(results, format="xml")

        assert "<test_results>" in xml_str
        assert "<test name=" in xml_str
        assert "status=" in xml_str

    def test_export_test_results_unsupported_format(self):
        """Test exporting with unsupported format."""
        results = {"summary": {}}
        with pytest.raises(ValueError, match="Unsupported format"):
            export_test_results(results, format="csv")

    def test_validate_test_configuration_valid(self):
        """Test validating valid configuration."""
        config = {
            "frameworks": ["pytest"],
            "test_paths": ["."],
            "thresholds": {"min_coverage": 85},
        }
        result = validate_test_configuration(config)

        assert result["is_valid"] is True
        assert len(result["errors"]) == 0

    def test_validate_test_configuration_missing_fields(self):
        """Test validating configuration with missing fields."""
        config = {"frameworks": ["pytest"]}
        result = validate_test_configuration(config)

        assert result["is_valid"] is False
        assert len(result["errors"]) > 0

    def test_validate_test_configuration_invalid_coverage(self):
        """Test validating configuration with invalid coverage."""
        config = {
            "frameworks": ["pytest"],
            "test_paths": ["."],
            "thresholds": {"min_coverage": 150},
        }
        result = validate_test_configuration(config)

        assert result["is_valid"] is False
        assert any("coverage" in err.lower() for err in result["errors"])

    def test_validate_test_configuration_max_duration(self):
        """Test validating configuration with invalid max_duration."""
        config = {
            "frameworks": ["pytest"],
            "test_paths": ["."],
            "thresholds": {"max_duration": -5},
        }
        result = validate_test_configuration(config)

        assert result["is_valid"] is False
        assert any("duration" in err.lower() for err in result["errors"])

    def test_validate_test_configuration_nonexistent_path(self):
        """Test validating configuration with nonexistent test paths."""
        config = {
            "frameworks": ["pytest"],
            "test_paths": ["/nonexistent/path/to/tests"],
            "thresholds": {"min_coverage": 85},
        }
        result = validate_test_configuration(config)

        assert "warnings" in result
        assert any("does not exist" in w.lower() for w in result.get("warnings", []))

    def test_validate_test_configuration_valid_with_recommendations(self):
        """Test validating valid configuration generates recommendations."""
        config = {
            "frameworks": ["pytest"],
            "test_paths": ["."],
            "thresholds": {"min_coverage": 85, "max_duration": 300},
        }
        result = validate_test_configuration(config)

        assert result["is_valid"] is True
        assert "recommendations" in result
        assert len(result["recommendations"]) > 0


class TestMainFunction:
    """Test the main demonstration function."""

    @mock.patch("moai_adk.foundation.testing.TestingFrameworkManager")
    @mock.patch("moai_adk.foundation.testing.QualityGateEngine")
    @mock.patch("moai_adk.foundation.testing.CoverageAnalyzer")
    @mock.patch("moai_adk.foundation.testing.TestAutomationOrchestrator")
    @mock.patch("moai_adk.foundation.testing.TestReportingSpecialist")
    @mock.patch("moai_adk.foundation.testing.TestDataManager")
    @mock.patch("moai_adk.foundation.testing.TestingMetricsCollector")
    @mock.patch("builtins.print")
    def test_main_function(
        self,
        mock_print,
        mock_metrics,
        mock_data,
        mock_reporting,
        mock_automation,
        mock_coverage,
        mock_quality,
        mock_framework,
    ):
        """Test the main function runs without errors."""
        from moai_adk.foundation.testing import main

        # Setup mocks
        mock_framework_instance = mock.MagicMock()
        mock_framework.return_value = mock_framework_instance
        mock_framework_instance.configure_pytest_environment.return_value = {
            "pytest": {}
        }

        mock_quality_instance = mock.MagicMock()
        mock_quality.return_value = mock_quality_instance
        mock_quality_instance.setup_code_quality_checks.return_value = {"linters": []}

        mock_coverage_instance = mock.MagicMock()
        mock_coverage.return_value = mock_coverage_instance
        mock_coverage_instance.analyze_code_coverage.return_value = {
            "summary": {"percentage": 85}
        }

        mock_automation_instance = mock.MagicMock()
        mock_automation.return_value = mock_automation_instance
        mock_automation_instance.setup_ci_pipeline.return_value = {"stages": []}

        mock_reporting_instance = mock.MagicMock()
        mock_reporting.return_value = mock_reporting_instance
        mock_reporting_instance.generate_test_reports.return_value = {
            "summary": {"total_tests": 10}
        }

        mock_data_instance = mock.MagicMock()
        mock_data.return_value = mock_data_instance
        mock_data_instance.create_test_datasets.return_value = {"test_datasets": []}

        mock_metrics_instance = mock.MagicMock()
        mock_metrics.return_value = mock_metrics_instance
        mock_metrics_instance.collect_test_metrics.return_value = {}
        mock_metrics_instance.calculate_quality_scores.return_value = {"grade": "A"}

        # Run main
        result = main()

        # Assert
        assert result is True
        mock_print.assert_called()
        assert mock_print.call_count >= 10  # Should print demo output
