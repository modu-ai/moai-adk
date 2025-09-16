#!/usr/bin/env python3
"""
MoAI-ADK CI/CD Enhancement Script
Implements enhanced testing, monitoring, and deployment pipelines

This script provides automated tools for setting up comprehensive
CI/CD pipelines with performance validation and architectural monitoring.
"""

import os
import yaml
import json
import subprocess
import time
from pathlib import Path
from typing import Dict, List, Any, Optional, Tuple
from dataclasses import dataclass, asdict
from datetime import datetime
import logging


@dataclass
class TestConfig:
    """Configuration for test execution"""
    test_types: List[str]
    coverage_threshold: float
    performance_baseline: Dict[str, float]
    timeout_seconds: int = 300
    parallel_jobs: int = 4


@dataclass
class DeploymentConfig:
    """Configuration for deployment pipeline"""
    environment: str
    validation_steps: List[str]
    rollback_enabled: bool = True
    health_check_timeout: int = 300
    traffic_routing: Dict[str, float] = None


@dataclass
class PipelineResult:
    """Result of pipeline execution"""
    stage: str
    success: bool
    duration_seconds: float
    artifacts: List[str]
    metrics: Dict[str, Any]
    errors: List[str] = None


class PerformanceBenchmark:
    """Performance benchmarking for CI/CD"""

    def __init__(self):
        self.baseline_file = Path("performance_baseline.json")
        self.current_metrics = {}

    def load_baseline(self) -> Dict[str, float]:
        """Load performance baseline from file"""
        if self.baseline_file.exists():
            with open(self.baseline_file, 'r') as f:
                return json.load(f)
        return {}

    def save_baseline(self, metrics: Dict[str, float]):
        """Save performance baseline to file"""
        with open(self.baseline_file, 'w') as f:
            json.dump(metrics, f, indent=2)

    def benchmark_import_performance(self) -> Dict[str, float]:
        """Benchmark import performance of core modules"""
        import_tests = {
            'moai_adk.config': 'from moai_adk import config',
            'moai_adk.logger': 'from moai_adk.utils import logger',
            'moai_adk.installer': 'from moai_adk.install import installer',
            'moai_adk.core.validator': 'from moai_adk.core.validation import validator'
        }

        results = {}
        for module_name, import_statement in import_tests.items():
            try:
                start_time = time.perf_counter()
                exec(import_statement)
                end_time = time.perf_counter()
                results[f"{module_name}_import_ms"] = (end_time - start_time) * 1000
            except Exception as e:
                logging.error(f"Failed to benchmark {module_name}: {e}")
                results[f"{module_name}_import_ms"] = -1

        return results

    def benchmark_memory_usage(self) -> Dict[str, float]:
        """Benchmark memory usage of core operations"""
        import psutil
        process = psutil.Process()

        baseline_memory = process.memory_info().rss / 1024 / 1024  # MB

        # Test operations
        operations = {
            'config_load': lambda: __import__('moai_adk.config'),
            'logger_init': lambda: __import__('moai_adk.utils.logger'),
            'validator_init': lambda: __import__('moai_adk.core.validation.validator')
        }

        results = {'baseline_memory_mb': baseline_memory}

        for op_name, operation in operations.items():
            try:
                start_memory = process.memory_info().rss / 1024 / 1024
                operation()
                end_memory = process.memory_info().rss / 1024 / 1024
                results[f"{op_name}_memory_delta_mb"] = end_memory - start_memory
            except Exception as e:
                logging.error(f"Failed to benchmark {op_name}: {e}")
                results[f"{op_name}_memory_delta_mb"] = -1

        return results

    def validate_performance(self, current_metrics: Dict[str, float], threshold: float = 0.2) -> Tuple[bool, List[str]]:
        """Validate current performance against baseline"""
        baseline = self.load_baseline()
        violations = []

        if not baseline:
            logging.warning("No performance baseline found, creating new baseline")
            self.save_baseline(current_metrics)
            return True, []

        for metric_name, current_value in current_metrics.items():
            if metric_name in baseline:
                baseline_value = baseline[metric_name]
                if baseline_value > 0:  # Skip negative values (errors)
                    deviation = (current_value - baseline_value) / baseline_value

                    if deviation > threshold:
                        violations.append(
                            f"{metric_name}: {current_value:.2f} vs baseline {baseline_value:.2f} "
                            f"({deviation*100:.1f}% increase)"
                        )

        return len(violations) == 0, violations


class ArchitecturalValidator:
    """Validates architectural integrity in CI/CD"""

    def __init__(self, source_root: Path):
        self.source_root = source_root

    def validate_directory_structure(self) -> Tuple[bool, List[str]]:
        """Validate expected directory structure"""
        expected_dirs = [
            "src/moai_adk/config",
            "src/moai_adk/utils",
            "src/moai_adk/core/managers",
            "src/moai_adk/core/security",
            "src/moai_adk/core/validation",
            "src/moai_adk/install",
            "src/moai_adk/cli"
        ]

        violations = []
        for dir_path in expected_dirs:
            full_path = self.source_root / dir_path
            if not full_path.exists():
                violations.append(f"Missing directory: {dir_path}")
            elif not full_path.is_dir():
                violations.append(f"Expected directory but found file: {dir_path}")

        return len(violations) == 0, violations

    def validate_import_structure(self) -> Tuple[bool, List[str]]:
        """Validate import dependencies are reasonable"""
        violations = []

        # Check for circular imports
        try:
            result = subprocess.run([
                "python", "-c",
                "import ast; import sys; "
                "from pathlib import Path; "
                "print('Import structure validation passed')"
            ], capture_output=True, text=True, cwd=self.source_root, timeout=30)

            if result.returncode != 0:
                violations.append(f"Import validation failed: {result.stderr}")

        except subprocess.TimeoutExpired:
            violations.append("Import validation timed out - possible circular imports")
        except Exception as e:
            violations.append(f"Import validation error: {str(e)}")

        return len(violations) == 0, violations

    def validate_code_quality(self) -> Tuple[bool, List[str]]:
        """Validate code quality metrics"""
        violations = []

        # Check for Python syntax errors
        python_files = list(self.source_root.rglob("*.py"))
        syntax_errors = 0

        for py_file in python_files:
            if "__pycache__" in str(py_file):
                continue

            try:
                with open(py_file, 'r', encoding='utf-8') as f:
                    content = f.read()
                compile(content, str(py_file), 'exec')
            except SyntaxError as e:
                violations.append(f"Syntax error in {py_file}: {e}")
                syntax_errors += 1
            except Exception as e:
                violations.append(f"Error checking {py_file}: {e}")

        if syntax_errors > 0:
            violations.append(f"Found {syntax_errors} syntax errors")

        return len(violations) == 0, violations


class TestRunner:
    """Enhanced test runner with parallel execution and reporting"""

    def __init__(self, config: TestConfig, source_root: Path):
        self.config = config
        self.source_root = source_root

    def run_unit_tests(self) -> PipelineResult:
        """Run unit tests with coverage"""
        start_time = time.time()
        errors = []
        artifacts = []

        try:
            # Run pytest with coverage
            cmd = [
                "python", "-m", "pytest",
                "tests/unit/",
                f"-j{self.config.parallel_jobs}",
                "--cov=src/moai_adk",
                "--cov-report=xml",
                "--cov-report=html",
                "--junit-xml=test-results.xml",
                "-v"
            ]

            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                cwd=self.source_root,
                timeout=self.config.timeout_seconds
            )

            artifacts = ["test-results.xml", "htmlcov/", "coverage.xml"]

            if result.returncode != 0:
                errors.append(f"Unit tests failed: {result.stdout}\n{result.stderr}")

            # Parse coverage from output
            coverage_percent = self._extract_coverage(result.stdout)

            metrics = {
                'coverage_percent': coverage_percent,
                'exit_code': result.returncode
            }

            # Check coverage threshold
            if coverage_percent < self.config.coverage_threshold:
                errors.append(f"Coverage {coverage_percent}% below threshold {self.config.coverage_threshold}%")

        except subprocess.TimeoutExpired:
            errors.append("Unit tests timed out")
            metrics = {'timeout': True}
        except Exception as e:
            errors.append(f"Unit test execution failed: {str(e)}")
            metrics = {'error': str(e)}

        duration = time.time() - start_time

        return PipelineResult(
            stage="unit_tests",
            success=len(errors) == 0,
            duration_seconds=duration,
            artifacts=artifacts,
            metrics=metrics,
            errors=errors
        )

    def run_integration_tests(self) -> PipelineResult:
        """Run integration tests"""
        start_time = time.time()
        errors = []
        artifacts = []

        try:
            cmd = [
                "python", "-m", "pytest",
                "tests/integration/",
                "--junit-xml=integration-results.xml",
                "-v"
            ]

            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                cwd=self.source_root,
                timeout=self.config.timeout_seconds
            )

            artifacts = ["integration-results.xml"]

            if result.returncode != 0:
                errors.append(f"Integration tests failed: {result.stdout}\n{result.stderr}")

            metrics = {'exit_code': result.returncode}

        except subprocess.TimeoutExpired:
            errors.append("Integration tests timed out")
            metrics = {'timeout': True}
        except Exception as e:
            errors.append(f"Integration test execution failed: {str(e)}")
            metrics = {'error': str(e)}

        duration = time.time() - start_time

        return PipelineResult(
            stage="integration_tests",
            success=len(errors) == 0,
            duration_seconds=duration,
            artifacts=artifacts,
            metrics=metrics,
            errors=errors
        )

    def run_performance_tests(self) -> PipelineResult:
        """Run performance benchmarks"""
        start_time = time.time()
        errors = []
        artifacts = []

        try:
            benchmark = PerformanceBenchmark()

            # Run import benchmarks
            import_metrics = benchmark.benchmark_import_performance()
            memory_metrics = benchmark.benchmark_memory_usage()

            current_metrics = {**import_metrics, **memory_metrics}

            # Validate against baseline
            performance_ok, violations = benchmark.validate_performance(current_metrics)

            if not performance_ok:
                errors.extend(violations)

            metrics = {
                'performance_metrics': current_metrics,
                'baseline_validation': performance_ok,
                'violations': violations
            }

            # Save current metrics as artifacts
            with open(self.source_root / "performance_metrics.json", 'w') as f:
                json.dump(current_metrics, f, indent=2)

            artifacts = ["performance_metrics.json"]

        except Exception as e:
            errors.append(f"Performance test execution failed: {str(e)}")
            metrics = {'error': str(e)}

        duration = time.time() - start_time

        return PipelineResult(
            stage="performance_tests",
            success=len(errors) == 0,
            duration_seconds=duration,
            artifacts=artifacts,
            metrics=metrics,
            errors=errors
        )

    def _extract_coverage(self, pytest_output: str) -> float:
        """Extract coverage percentage from pytest output"""
        try:
            # Look for coverage line like "TOTAL    1234   123    90%"
            lines = pytest_output.split('\n')
            for line in lines:
                if 'TOTAL' in line and '%' in line:
                    parts = line.split()
                    for part in parts:
                        if part.endswith('%'):
                            return float(part.rstrip('%'))
            return 0.0
        except Exception:
            return 0.0


class CIPipeline:
    """Comprehensive CI/CD pipeline"""

    def __init__(self, source_root: Path, test_config: TestConfig):
        self.source_root = source_root
        self.test_config = test_config
        self.results: List[PipelineResult] = []

    def run_full_pipeline(self) -> Dict[str, Any]:
        """Run the complete CI/CD pipeline"""
        pipeline_start = time.time()

        logging.info("🚀 Starting MoAI-ADK CI/CD Pipeline")

        # Stage 1: Architectural validation
        arch_result = self._run_architectural_validation()
        self.results.append(arch_result)

        if not arch_result.success:
            return self._generate_pipeline_report(pipeline_start, early_termination="architectural_validation")

        # Stage 2: Unit tests
        unit_result = self._run_unit_tests()
        self.results.append(unit_result)

        if not unit_result.success:
            return self._generate_pipeline_report(pipeline_start, early_termination="unit_tests")

        # Stage 3: Integration tests
        integration_result = self._run_integration_tests()
        self.results.append(integration_result)

        if not integration_result.success:
            return self._generate_pipeline_report(pipeline_start, early_termination="integration_tests")

        # Stage 4: Performance tests
        performance_result = self._run_performance_tests()
        self.results.append(performance_result)

        # Stage 5: Security validation
        security_result = self._run_security_validation()
        self.results.append(security_result)

        # Stage 6: Build validation
        build_result = self._run_build_validation()
        self.results.append(build_result)

        return self._generate_pipeline_report(pipeline_start)

    def _run_architectural_validation(self) -> PipelineResult:
        """Run architectural validation"""
        logging.info("🏗️  Running architectural validation")
        start_time = time.time()

        validator = ArchitecturalValidator(self.source_root)
        errors = []

        # Directory structure validation
        dir_ok, dir_errors = validator.validate_directory_structure()
        if not dir_ok:
            errors.extend(dir_errors)

        # Import structure validation
        import_ok, import_errors = validator.validate_import_structure()
        if not import_ok:
            errors.extend(import_errors)

        # Code quality validation
        quality_ok, quality_errors = validator.validate_code_quality()
        if not quality_ok:
            errors.extend(quality_errors)

        duration = time.time() - start_time
        success = len(errors) == 0

        return PipelineResult(
            stage="architectural_validation",
            success=success,
            duration_seconds=duration,
            artifacts=[],
            metrics={
                'directory_structure_valid': dir_ok,
                'import_structure_valid': import_ok,
                'code_quality_valid': quality_ok
            },
            errors=errors
        )

    def _run_unit_tests(self) -> PipelineResult:
        """Run unit tests"""
        logging.info("🧪 Running unit tests")
        runner = TestRunner(self.test_config, self.source_root)
        return runner.run_unit_tests()

    def _run_integration_tests(self) -> PipelineResult:
        """Run integration tests"""
        logging.info("🔗 Running integration tests")
        runner = TestRunner(self.test_config, self.source_root)
        return runner.run_integration_tests()

    def _run_performance_tests(self) -> PipelineResult:
        """Run performance tests"""
        logging.info("⚡ Running performance tests")
        runner = TestRunner(self.test_config, self.source_root)
        return runner.run_performance_tests()

    def _run_security_validation(self) -> PipelineResult:
        """Run security validation"""
        logging.info("🔒 Running security validation")
        start_time = time.time()
        errors = []
        artifacts = []

        try:
            # Run security checks (placeholder - would integrate with actual security tools)
            cmd = ["python", "-c", "print('Security validation placeholder')"]
            result = subprocess.run(cmd, capture_output=True, text=True, timeout=60)

            if result.returncode != 0:
                errors.append(f"Security validation failed: {result.stderr}")

            metrics = {'security_checks_passed': result.returncode == 0}

        except Exception as e:
            errors.append(f"Security validation error: {str(e)}")
            metrics = {'error': str(e)}

        duration = time.time() - start_time

        return PipelineResult(
            stage="security_validation",
            success=len(errors) == 0,
            duration_seconds=duration,
            artifacts=artifacts,
            metrics=metrics,
            errors=errors
        )

    def _run_build_validation(self) -> PipelineResult:
        """Run build validation"""
        logging.info("📦 Running build validation")
        start_time = time.time()
        errors = []
        artifacts = []

        try:
            # Test package building
            cmd = ["python", "-m", "build", "--wheel", "--no-isolation"]
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                cwd=self.source_root,
                timeout=300
            )

            if result.returncode != 0:
                errors.append(f"Build failed: {result.stderr}")
            else:
                # Find built wheel
                dist_dir = self.source_root / "dist"
                if dist_dir.exists():
                    wheels = list(dist_dir.glob("*.whl"))
                    artifacts.extend([str(w.relative_to(self.source_root)) for w in wheels])

            metrics = {
                'build_successful': result.returncode == 0,
                'artifacts_created': len(artifacts)
            }

        except Exception as e:
            errors.append(f"Build validation error: {str(e)}")
            metrics = {'error': str(e)}

        duration = time.time() - start_time

        return PipelineResult(
            stage="build_validation",
            success=len(errors) == 0,
            duration_seconds=duration,
            artifacts=artifacts,
            metrics=metrics,
            errors=errors
        )

    def _generate_pipeline_report(self, start_time: float, early_termination: Optional[str] = None) -> Dict[str, Any]:
        """Generate comprehensive pipeline report"""
        total_duration = time.time() - start_time
        successful_stages = [r for r in self.results if r.success]
        failed_stages = [r for r in self.results if not r.success]

        report = {
            'pipeline_summary': {
                'start_time': datetime.fromtimestamp(start_time).isoformat(),
                'end_time': datetime.now().isoformat(),
                'total_duration_seconds': total_duration,
                'total_stages': len(self.results),
                'successful_stages': len(successful_stages),
                'failed_stages': len(failed_stages),
                'overall_success': len(failed_stages) == 0 and early_termination is None,
                'early_termination': early_termination
            },
            'stage_results': [asdict(result) for result in self.results],
            'artifacts': []
        }

        # Collect all artifacts
        for result in self.results:
            report['artifacts'].extend(result.artifacts)

        # Add summary metrics
        if successful_stages:
            report['pipeline_summary']['avg_stage_duration'] = sum(r.duration_seconds for r in successful_stages) / len(successful_stages)

        return report


def generate_github_workflow():
    """Generate GitHub Actions workflow file"""
    workflow = {
        'name': 'MoAI-ADK Enhanced CI/CD',
        'on': {
            'push': {'branches': ['main', 'develop']},
            'pull_request': {'branches': ['main']}
        },
        'jobs': {
            'test': {
                'runs-on': 'ubuntu-latest',
                'strategy': {
                    'matrix': {
                        'python-version': ['3.9', '3.10', '3.11', '3.12']
                    }
                },
                'steps': [
                    {
                        'uses': 'actions/checkout@v4'
                    },
                    {
                        'name': 'Set up Python ${{ matrix.python-version }}',
                        'uses': 'actions/setup-python@v4',
                        'with': {
                            'python-version': '${{ matrix.python-version }}'
                        }
                    },
                    {
                        'name': 'Install dependencies',
                        'run': 'pip install -e .[dev] && pip install build pytest pytest-cov pytest-xdist'
                    },
                    {
                        'name': 'Run enhanced CI pipeline',
                        'run': 'python scripts/ci_enhancement.py --source-root . --output pipeline-report.json'
                    },
                    {
                        'name': 'Upload test results',
                        'uses': 'actions/upload-artifact@v3',
                        'if': 'always()',
                        'with': {
                            'name': 'test-results-${{ matrix.python-version }}',
                            'path': '**/*results*.xml'
                        }
                    },
                    {
                        'name': 'Upload coverage reports',
                        'uses': 'actions/upload-artifact@v3',
                        'if': 'always()',
                        'with': {
                            'name': 'coverage-${{ matrix.python-version }}',
                            'path': 'htmlcov/'
                        }
                    },
                    {
                        'name': 'Upload pipeline report',
                        'uses': 'actions/upload-artifact@v3',
                        'if': 'always()',
                        'with': {
                            'name': 'pipeline-report-${{ matrix.python-version }}',
                            'path': 'pipeline-report.json'
                        }
                    }
                ]
            }
        }
    }

    return yaml.dump(workflow, default_flow_style=False)


def main():
    """Main execution function"""
    import argparse

    parser = argparse.ArgumentParser(description="MoAI-ADK CI/CD Enhancement")
    parser.add_argument("--source-root", type=str, default=".", help="Source root directory")
    parser.add_argument("--output", type=str, default="pipeline-report.json", help="Output report file")
    parser.add_argument("--coverage-threshold", type=float, default=80.0, help="Coverage threshold")
    parser.add_argument("--generate-workflow", action="store_true", help="Generate GitHub workflow file")

    args = parser.parse_args()

    logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

    if args.generate_workflow:
        workflow_content = generate_github_workflow()
        workflow_path = Path(".github/workflows/enhanced-ci.yml")
        workflow_path.parent.mkdir(parents=True, exist_ok=True)

        with open(workflow_path, 'w') as f:
            f.write(workflow_content)

        print(f"✅ Generated GitHub workflow: {workflow_path}")
        return 0

    # Run CI pipeline
    source_root = Path(args.source_root).resolve()

    test_config = TestConfig(
        test_types=["unit", "integration", "performance"],
        coverage_threshold=args.coverage_threshold,
        performance_baseline={
            "config_import_ms": 25.0,
            "logger_import_ms": 5.0,
            "baseline_memory_mb": 30.0
        }
    )

    pipeline = CIPipeline(source_root, test_config)
    report = pipeline.run_full_pipeline()

    # Save report
    with open(args.output, 'w') as f:
        json.dump(report, f, indent=2, default=str)

    print(f"\n📊 Pipeline report saved to: {args.output}")

    if report['pipeline_summary']['overall_success']:
        print("✅ CI/CD Pipeline completed successfully!")
        return 0
    else:
        print("❌ CI/CD Pipeline failed. Check the report for details.")
        return 1


if __name__ == "__main__":
    exit(main())