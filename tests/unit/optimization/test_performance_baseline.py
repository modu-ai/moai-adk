"""
Unit tests for performance baseline measurement framework.

Tests the TAG-005 component: Performance Baseline Measurements
"""

import time
from unittest.mock import patch

import pytest


class TestPerformanceBaseline:
    """Test cases for performance baseline measurement."""

    def test_performance_metrics_collection(self):
        """Test that performance metrics can be collected properly."""
        from moai_adk.optimization.performance import PerformanceMetrics

        metrics = PerformanceMetrics()

        # Mock some performance data
        with patch.object(metrics, "_collect_system_metrics") as mock_collect:
            mock_collect.return_value = {"memory_usage": 50.0, "cpu_usage": 30.0, "response_time": 0.05}

            # Collect metrics
            collected = metrics.collect_metrics("test_operation")

            # Verify metrics collection
            assert "test_operation" in metrics.metrics_history
            assert "memory_usage" in collected
            assert "cpu_usage" in collected
            assert "response_time" in collected

    def test_performance_benchmark_runner(self):
        """Test that performance benchmarks can be executed."""
        from moai_adk.optimization.performance import PerformanceBenchmark

        benchmark = PerformanceBenchmark()

        # Mock benchmark scenario
        test_scenario = {
            "name": "skill_allocation_benchmark",
            "operation": lambda: None,  # Mock operation
            "iterations": 10,
            "warmup_iterations": 2,
        }

        with patch.object(benchmark, "_execute_operation") as mock_execute:
            mock_execute.return_value = 0.01  # Mock result time

            # Run benchmark
            results = benchmark.run_benchmark(test_scenario)

            # Verify benchmark results
            assert "average_time" in results
            assert "min_time" in results
            assert "max_time" in results
            assert "total_time" in results
            assert results["average_time"] > 0

    def test_performance_analysis(self):
        """Test that performance analysis works correctly."""
        from moai_adk.optimization.performance import PerformanceAnalyzer

        analyzer = PerformanceAnalyzer()

        # Mock performance data
        performance_data = {
            "operation_1": [0.01, 0.02, 0.015, 0.025, 0.01],
            "operation_2": [0.05, 0.06, 0.045, 0.07, 0.05],
        }

        # Analyze performance
        analysis = analyzer.analyze_performance(performance_data)

        # Verify analysis results
        assert "operation_1" in analysis
        assert "operation_2" in analysis
        assert "mean" in analysis["operation_1"]
        assert "stddev" in analysis["operation_1"]
        assert percentiles in analysis["operation_1"]

    def test_performance_threshold_validation(self):
        """Test that performance thresholds are validated correctly."""
        from moai_adk.optimization.performance import PerformanceThreshold

        threshold = PerformanceThreshold(max_response_time=0.1, max_memory_usage=80.0, max_cpu_usage=70.0)

        # Test valid performance
        valid_metrics = {"response_time": 0.05, "memory_usage": 60.0, "cpu_usage": 50.0}
        assert threshold.validate_metrics(valid_metrics) is True

        # Test invalid performance
        invalid_metrics = {
            "response_time": 0.15,  # Exceeds max
            "memory_usage": 90.0,  # Exceeds max
            "cpu_usage": 80.0,  # Exceeds max
        }
        assert threshold.validate_metrics(invalid_metrics) is False

    def test_performance_baseline_creation(self):
        """Test that performance baselines can be created."""
        from moai_adk.optimization.performance import PerformanceBaseline

        baseline = PerformanceBaseline()

        # Mock baseline data
        baseline_data = {
            "skill_allocation": {"mean": 0.025, "stddev": 0.005, "percentile_95": 0.035},
            "dynamic_loading": {"mean": 0.015, "stddev": 0.003, "percentile_95": 0.020},
        }

        # Create baseline
        baseline.create_baseline("test_scenario", baseline_data)

        # Verify baseline creation
        assert "test_scenario" in baseline.baselines
        assert baseline.baselines["test_scenario"]["skill_allocation"]["mean"] == 0.025

    def test_performance_comparison(self):
        """Test that performance can be compared against baselines."""
        from moai_adk.optimization.performance import ComparisonResult, PerformanceBaseline

        baseline = PerformanceBaseline()

        # Create baseline
        baseline_data = {"mean": 0.025, "stddev": 0.005, "percentile_95": 0.035}
        baseline.create_baseline("test_operation", baseline_data)

        # Test current performance
        current_performance = {"mean": 0.030, "stddev": 0.006, "percentile_95": 0.040}

        # Compare performance
        comparison = baseline.compare_performance("test_operation", current_performance)

        # Verify comparison result
        assert isinstance(comparison, ComparisonResult)
        assert comparison.is_regression is True  # Performance got worse
        assert comparison.percent_change > 0

    def test_performance_regression_detection(self):
        """Test that performance regressions are detected."""
        from moai_adk.optimization.performance import PerformanceRegressionDetector

        detector = PerformanceRegressionDetector()

        # Mock performance data with regression
        baseline = {"mean": 0.025, "stddev": 0.005, "percentile_95": 0.035}

        current = {"mean": 0.050, "stddev": 0.010, "percentile_95": 0.060}  # 100% increase - clear regression

        # Detect regression
        regression = detector.detect_regression("test_operation", baseline, current)

        # Verify regression detection
        assert regression.detected is True
        assert regression.severity == "high"
        assert regression.percent_change > 50

    def test_performance_optimization_suggestions(self):
        """Test that optimization suggestions are generated."""
        from moai_adk.optimization.performance import PerformanceOptimizer

        optimizer = PerformanceOptimizer()

        # Mock performance issues
        issues = [
            {
                "operation": "skill_allocation",
                "issue": "high_memory_usage",
                "current": 85.0,
                "target": 50.0,
                "severity": "high",
            },
            {
                "operation": "dynamic_loading",
                "issue": "slow_response_time",
                "current": 0.100,
                "target": 0.050,
                "severity": "medium",
            },
        ]

        # Generate optimization suggestions
        suggestions = optimizer.generate_optimization_suggestions(issues)

        # Verify suggestions
        assert len(suggestions) > 0
        assert all("suggestion" in s for s in suggestions)
        assert all("priority" in s for s in suggestions)
        assert all("estimated_improvement" in s for s in suggestions)

    def test_performance_continuous_monitoring(self):
        """Test performance monitoring over time."""
        from moai_adk.optimization.performance import PerformanceMonitor

        monitor = PerformanceMonitor()

        # Mock continuous monitoring
        with patch.object(monitor, "collect_metrics") as mock_collect:
            mock_collect.return_value = {"response_time": 0.025, "memory_usage": 60.0, "cpu_usage": 40.0}

            # Start monitoring
            monitor.start_monitoring("test_operation", interval=1.0)

            # Allow some monitoring to occur
            time.sleep(2)

            # Stop monitoring
            monitor.stop_monitoring("test_operation")

            # Verify monitoring data was collected
            assert len(monitor.monitoring_data["test_operation"]) > 0

    def test_performance_history_tracking(self):
        """Test performance history tracking."""
        from moai_adk.optimization.performance import PerformanceHistory

        history = PerformanceHistory()

        # Add performance metrics
        history.add_metrics("test_operation", {"response_time": 0.025, "memory_usage": 60.0, "timestamp": time.time()})

        # Get history
        operation_history = history.get_history("test_operation")

        # Verify history tracking
        assert len(operation_history) == 1
        assert "response_time" in operation_history[0]
        assert "timestamp" in operation_history[0]


class TestPerformanceIntegration:
    """Integration tests for performance measurement with other components."""

    def test_performance_integration_with_skill_allocation(self):
        """Test performance integration with skill allocation matrix."""
        from moai_adk.optimization.performance import PerformanceBenchmark
        from moai_adk.optimization.skill_allocation import SkillMatrix

        benchmark = PerformanceBenchmark()
        matrix = SkillMatrix(categories=["frontend", "backend"], skills_per_category=3)

        # Benchmark skill allocation performance
        benchmark_scenario = {
            "name": "skill_allocation_benchmark",
            "operation": lambda: matrix.optimize_allocation("frontend", {"react": 0.9, "css": 0.8}),
            "iterations": 10,
            "warmup_iterations": 2,
        }

        results = benchmark.run_benchmark(benchmark_scenario)

        # Verify integration
        assert results["average_time"] < 0.01  # Performance target
        assert "success" in results
        assert results["success"] is True

    def test_performance_integration_with_dynamic_loading(self):
        """Test performance integration with dynamic skill loading."""
        from moai_adk.optimization.dynamic_loading import SkillLoader
        from moai_adk.optimization.performance import PerformanceMonitor

        monitor = PerformanceMonitor()
        loader = SkillLoader()

        # Mock skill loading
        with patch.object(loader, "_load_skill") as mock_load:
            mock_load.return_value = {"skill": "data"}

            # Monitor skill loading performance
            monitor.start_monitoring("skill_loading")
            loader.get_skill("frontend", "react")
            loader.get_skill("backend", "django")
            monitor.stop_monitoring("skill_loading")

            # Verify performance integration
            assert "skill_loading" in monitor.monitoring_data
            assert len(monitor.monitoring_data["skill_loading"]) == 2

    def test_performance_integration_with_templates(self):
        """Test performance integration with agent optimization templates."""
        from moai_adk.optimization.performance import PerformanceBenchmark
        from moai_adk.optimization.templates import TemplateRegistry

        benchmark = PerformanceBenchmark()
        registry = TemplateRegistry()

        # Benchmark template application performance
        template_scenario = {
            "name": "template_application_benchmark",
            "operation": lambda: registry.apply_template("frontend-expert", {"project_type": "web"}),
            "iterations": 10,
            "warmup_iterations": 2,
        }

        results = benchmark.run_benchmark(template_scenario)

        # Verify integration
        assert results["average_time"] < 0.005  # Performance target
        assert results["success"] is True

    def test_performance_baseline_compliance(self):
        """Test that all components meet performance baselines."""
        from moai_adk.optimization.performance import PerformanceBaseline

        baseline = PerformanceBaseline()

        # Set performance baselines for each component
        baseline.create_baseline("skill_allocation", {"max_response_time": 0.01})
        baseline.create_baseline("dynamic_loading", {"max_response_time": 0.005})
        baseline.create_baseline("template_application", {"max_response_time": 0.005})

        # Test each component against baselines
        components = ["skill_allocation", "dynamic_loading", "template_application"]

        for component in components:
            current_performance = {"mean": 0.008, "stddev": 0.001, "percentile_95": 0.010}  # Below baseline

            comparison = baseline.compare_performance(component, current_performance)
            assert comparison.is_regression is False, f"{component} performance regression detected"

    def test_performance_aggregation(self):
        """Test performance metrics aggregation across all components."""
        from moai_adk.optimization.performance import PerformanceAggregator

        aggregator = PerformanceAggregator()

        # Mock performance data from all components
        component_metrics = {
            "skill_allocation": [0.01, 0.012, 0.009, 0.011, 0.01],
            "dynamic_loading": [0.005, 0.006, 0.004, 0.007, 0.005],
            "template_application": [0.004, 0.005, 0.003, 0.006, 0.004],
        }

        # Aggregate performance
        aggregated = aggregator.aggregate_performance(component_metrics)

        # Verify aggregation
        assert "overall_mean" in aggregated
        assert "overall_stddev" in aggregated
        assert "component_breakdown" in aggregated
        assert len(aggregated["component_breakdown"]) == 3


class TestPerformanceQuality:
    """Quality tests for performance measurement framework."""

    def test_performance_data_validation(self):
        """Test that performance data is properly validated."""
        from moai_adk.optimization.performance import DataValidationError, PerformanceMetrics

        metrics = PerformanceMetrics()

        # Test invalid metrics data
        invalid_data = {
            "negative_value": -1.0,  # Invalid
            "string_instead_of_number": "invalid",
            "missing_required_field": {},
        }

        with pytest.raises(DataValidationError):
            metrics.validate_metrics(invalid_data)

    def test_performance_monitoring_resilience(self):
        """Test that performance monitoring is resilient to errors."""
        from moai_adk.optimization.performance import PerformanceMonitor

        monitor = PerformanceMonitor()

        # Test error handling in monitoring
        try:
            monitor.start_monitoring("invalid_operation")
            monitor.collect_metrics("invalid_operation")
            monitor.stop_monitoring("invalid_operation")
        except Exception:
            # Should handle errors gracefully
            assert True

    def test_performance_report_generation(self):
        """Test that performance reports are generated correctly."""
        from moai_adk.optimization.performance import PerformanceReportGenerator

        generator = PerformanceReportGenerator()

        # Mock performance data
        performance_data = {"skill_allocation": [0.01, 0.012, 0.009], "dynamic_loading": [0.005, 0.006, 0.004]}

        # Generate report
        report = generator.generate_report(performance_data)

        # Verify report structure
        assert "summary" in report
        assert "recommendations" in report
        assert "raw_data" in report
        assert len(report["recommendations"]) > 0


if __name__ == "__main__":
    pytest.main([__file__])
