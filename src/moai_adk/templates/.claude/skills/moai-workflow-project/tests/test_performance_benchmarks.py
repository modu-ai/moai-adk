"""
Performance Benchmark Tests

Performance tests and benchmarks for the moai-menu-project system including:
- Load time and response speed measurements
- Memory usage analysis
- Large dataset handling
- Concurrent access testing
- Resource efficiency metrics
"""

import os
import tempfile
import time
import unittest
from pathlib import Path

from moai_menu_project import MoaiMenuProject


class TestPerformanceBenchmarks(unittest.TestCase):
    """Performance benchmarking tests."""

    def setUp(self):
        """Set up test environment."""
        self.test_dir = Path(tempfile.mkdtemp())
        self.project = MoaiMenuProject(str(self.test_dir))
        self.performance_results = {}

    def tearDown(self):
        """Clean up test environment."""
        import shutil

        shutil.rmtree(self.test_dir, ignore_errors=True)

    def measure_time(self, func, *args, **kwargs):
        """Measure execution time of a function."""
        start_time = time.time()
        result = func(*args, **kwargs)
        end_time = time.time()
        execution_time = end_time - start_time

        return result, execution_time

    def test_initialization_performance(self):
        """Test project initialization performance."""

        # Measure initialization time
        result, init_time = self.measure_time(
            self.project.initialize_complete_project,
            language="en",
            project_type="web_application",
            optimization_enabled=True,
        )

        # Record performance metric
        self.performance_results["initialization_time"] = init_time

        # Performance assertions
        self.assertLess(init_time, 10.0, "Initialization should complete within 10 seconds")
        self.assertTrue(result["success"])

        print(f"✅ Initialization completed in {init_time:.2f} seconds")

    def test_module_load_performance(self):
        """Test individual module load performance."""

        # Test DocumentationManager load time
        from moai_menu_project.modules.documentation_manager import DocumentationManager

        start_time = time.time()
        DocumentationManager(str(self.test_dir), self.project.config)
        doc_load_time = time.time() - start_time

        # Test LanguageInitializer load time
        from moai_menu_project.modules.language_initializer import LanguageInitializer

        start_time = time.time()
        LanguageInitializer(str(self.test_dir), self.project.config)
        lang_load_time = time.time() - start_time

        # Test TemplateOptimizer load time
        from moai_menu_project.modules.template_optimizer import TemplateOptimizer

        start_time = time.time()
        TemplateOptimizer(str(self.test_dir), self.project.config)
        template_load_time = time.time() - start_time

        # Performance assertions
        self.assertLess(doc_load_time, 1.0, "DocumentationManager should load within 1 second")
        self.assertLess(lang_load_time, 1.0, "LanguageInitializer should load within 1 second")
        self.assertLess(template_load_time, 1.0, "TemplateOptimizer should load within 1 second")

        print(f"✅ DocumentationManager loaded in {doc_load_time:.3f} seconds")
        print(f"✅ LanguageInitializer loaded in {lang_load_time:.3f} seconds")
        print(f"✅ TemplateOptimizer loaded in {template_load_time:.3f} seconds")

    def test_documentation_generation_performance(self):
        """Test documentation generation performance."""

        # Initialize project first
        self.project.initialize_complete_project()

        # Create test SPEC data
        spec_data = {
            "id": "SPEC-001",
            "title": "Performance Test Feature",
            "description": "Testing documentation generation performance",
            "requirements": [f"Requirement {i}" for i in range(50)],  # Large requirement list
            "api_endpoints": [
                {
                    "path": f"/api/endpoint/{i}",
                    "method": "POST",
                    "description": f"Test endpoint {i}",
                }
                for i in range(20)  # 20 endpoints
            ],
        }

        # Measure documentation generation time
        result, gen_time = self.measure_time(self.project.generate_documentation_from_spec, spec_data)

        # Record performance metric
        self.performance_results["documentation_generation_time"] = gen_time

        # Performance assertions
        self.assertLess(gen_time, 5.0, "Documentation generation should complete within 5 seconds")
        self.assertTrue(result["success"])

        print(
            f"✅ Documentation generated in {gen_time:.2f} seconds with {len(spec_data['requirements'])} requirements"
        )

    def test_template_optimization_performance(self):
        """Test template optimization performance."""

        # Create multiple test template files
        templates_dir = self.test_dir / "templates"
        templates_dir.mkdir(exist_ok=True)

        # Create 10 test templates
        for i in range(10):
            template_content = f"""
# Template {i}

This is test template {i} with multiple sections.

## Section 1
Content for section 1 of template {i}.

## Section 2
Content for section 2 of template {i}.

## Section 3
Content for section 3 of template {i}.

Repeated content for testing optimization: {"test " * 100}

Complex logic:
{{% if condition_{i} %}}
    {{% if nested_condition_{i} %}}
        Content {i}
    {{% endif %}}
{{% endif %}}
            """

            (templates_dir / f"template_{i}.md").write_text(template_content, encoding="utf-8")

        # Measure template analysis time
        result, analysis_time = self.measure_time(self.project.template_optimizer.analyze_project_templates)

        # Measure optimization time
        result, opt_time = self.measure_time(self.project.template_optimizer.create_optimized_templates)

        # Record performance metrics
        self.performance_results["template_analysis_time"] = analysis_time
        self.performance_results["template_optimization_time"] = opt_time

        # Performance assertions
        self.assertLess(analysis_time, 3.0, "Template analysis should complete within 3 seconds")
        self.assertLess(opt_time, 5.0, "Template optimization should complete within 5 seconds")

        print(f"✅ Template analysis completed in {analysis_time:.2f} seconds for 10 templates")
        print(f"✅ Template optimization completed in {opt_time:.2f} seconds")

    def test_language_detection_performance(self):
        """Test language detection performance with various project sizes."""

        # Create test project files with different content
        files_to_create = [
            ("main.py", "# Main Python file\nimport os\nprint('Hello World')"),
            ("config.json", '{"name": "test", "language": "en"}'),
            ("README.md", "# Test Project\\nThis is a test project"),
            ("docs/guide.md", "# User Guide\\nThis guide explains..."),
        ]

        # Create multiple language variations
        korean_files = [
            (
                "korean_module.py",
                "# 한국어 모듈\\ndef calculate():\\n    # 계산 함수\\n    return 100",
            ),
            ("korean_readme.md", "# 한국어 프로젝트\\n이것은 테스트 프로젝트입니다"),
        ]

        # Create test files
        for file_path, content in files_to_create + korean_files:
            full_path = self.test_dir / file_path
            full_path.parent.mkdir(parents=True, exist_ok=True)
            full_path.write_text(content, encoding="utf-8")

        # Measure language detection time
        result, detection_time = self.measure_time(self.project.language_initializer.detect_project_language)

        # Record performance metric
        self.performance_results["language_detection_time"] = detection_time

        # Performance assertions
        self.assertLess(detection_time, 1.0, "Language detection should complete within 1 second")

        print(f"✅ Language detection completed in {detection_time:.3f} seconds")

    def test_memory_usage_analysis(self):
        """Test memory usage during operations."""

        try:
            import gc  # noqa: F401 - memory profiling availability check

            import psutil

            # Get initial memory usage
            process = psutil.Process(os.getpid())
            initial_memory = process.memory_info().rss / 1024 / 1024  # MB

            # Perform memory-intensive operations
            self.project.initialize_complete_project(optimization_enabled=True)

            # Create large SPEC data
            large_spec = {
                "id": "SPEC-LARGE",
                "title": "Large Feature Test",
                "description": "Testing memory usage with large data",
                "requirements": [f"Requirement {i}" for i in range(1000)],
                "api_endpoints": [
                    {
                        "path": f"/api/endpoint/{i}",
                        "method": "POST",
                        "description": f"Test endpoint {i} with large content {'x' * 100}",
                    }
                    for i in range(100)
                ],
            }

            self.project.generate_documentation_from_spec(large_spec)

            # Get final memory usage
            final_memory = process.memory_info().rss / 1024 / 1024  # MB
            memory_increase = final_memory - initial_memory

            # Record performance metric
            self.performance_results["memory_increase_mb"] = memory_increase

            # Memory assertions (should be reasonable)
            self.assertLess(memory_increase, 100, "Memory increase should be less than 100MB")

            print(f"✅ Memory usage increased by {memory_increase:.2f} MB during operations")

        except ImportError:
            self.skipTest("psutil not available for memory testing")

    def test_concurrent_operations(self):
        """Test system performance under concurrent operations."""

        import concurrent.futures

        def initialize_subproject(index):
            """Initialize a subproject for concurrent testing."""
            subproject_dir = self.test_dir / f"subproject_{index}"
            subproject_dir.mkdir(exist_ok=True)

            subproject = MoaiMenuProject(str(subproject_dir))
            return subproject.initialize_complete_project(language="en", project_type="web_application")

        # Test with 5 concurrent operations
        start_time = time.time()

        with concurrent.futures.ThreadPoolExecutor(max_workers=5) as executor:
            futures = [executor.submit(initialize_subproject, i) for i in range(5)]
            results = [future.result() for future in concurrent.futures.as_completed(futures)]

        concurrent_time = time.time() - start_time

        # Record performance metric
        self.performance_results["concurrent_operations_time"] = concurrent_time

        # Verify all operations succeeded
        self.assertEqual(len(results), 5)
        self.assertTrue(all(result["success"] for result in results))

        # Performance assertions
        self.assertLess(
            concurrent_time,
            20.0,
            "5 concurrent operations should complete within 20 seconds",
        )

        print(f"✅ 5 concurrent operations completed in {concurrent_time:.2f} seconds")

    def test_large_project_handling(self):
        """Test performance with large project structures."""

        # Create a large project structure
        dirs_to_create = [
            "src/backend/module1",
            "src/backend/module2",
            "src/frontend/components",
            "docs/api",
            "docs/guides",
            "tests/unit",
            "tests/integration",
            "config/dev",
            "config/prod",
            "scripts/deploy",
        ]

        # Create directory structure
        for dir_path in dirs_to_create:
            full_path = self.test_dir / dir_path
            full_path.mkdir(parents=True, exist_ok=True)

        # Create template files in each directory
        template_count = 0
        for dir_path in dirs_to_create:
            for i in range(5):  # 5 templates per directory
                requirements = "\n".join([f"- Requirement {j}" for j in range(10)])
                endpoints = "\n".join([f"- Endpoint {j}" for j in range(5)])
                template_content = f"""
# Template for {dir_path}/{i}

## Content Section
This is template {i} in {dir_path}.

## Requirements
{requirements}

## API Endpoints
{endpoints}

Complex structure:
{{% for item in items %}}
    {{% if item.active %}}
        {{ item.name }}
    {{% endif %}}
{{% endfor %}}
                """

                template_file = self.test_dir / dir_path / f"template_{i}.md"
                template_file.write_text(template_content, encoding="utf-8")
                template_count += 1

        # Measure performance with large project
        start_time = time.time()

        # Initialize
        init_result = self.project.initialize_complete_project(optimization_enabled=True)

        # Analyze templates
        analysis_result = self.project.template_optimizer.analyze_project_templates()

        # Optimize templates
        opt_result = self.project.template_optimizer.create_optimized_templates()

        total_time = time.time() - start_time

        # Record performance metrics
        self.performance_results["large_project_time"] = total_time
        self.performance_results["templates_analyzed"] = len(analysis_result.get("analyzed_files", []))

        # Performance assertions
        self.assertLess(
            total_time,
            30.0,
            "Large project operations should complete within 30 seconds",
        )
        self.assertTrue(init_result["success"])
        self.assertTrue(analysis_result["success"])
        self.assertTrue(opt_result["success"])

        print(f"✅ Large project ({template_count} templates) processed in {total_time:.2f} seconds")

    def test_performance_regression_detection(self):
        """Test detection of performance regressions."""

        # Define performance baselines (these should be updated based on actual measurements)
        performance_baselines = {
            "initialization_time": 10.0,  # seconds
            "documentation_generation_time": 5.0,  # seconds
            "template_analysis_time": 3.0,  # seconds
            "language_detection_time": 1.0,  # seconds
            "memory_increase_mb": 100,  # MB
        }

        # Run all performance tests to get metrics
        self.test_initialization_performance()
        self.test_documentation_generation_performance()
        self.test_template_optimization_performance()
        self.test_language_detection_performance()

        # Check for regressions
        regressions_detected = []

        for metric, baseline in performance_baselines.items():
            if metric in self.performance_results:
                actual_value = self.performance_results[metric]
                if actual_value > baseline:
                    regressions_detected.append(
                        {
                            "metric": metric,
                            "baseline": baseline,
                            "actual": actual_value,
                            "regression_percentage": ((actual_value - baseline) / baseline) * 100,
                        }
                    )

        # Assert no significant regressions
        if regressions_detected:
            regression_details = "\\n".join(
                [
                    f"  - {reg['metric']}: {reg['actual']:.2f} (baseline: {reg['baseline']:.2f}, "
                    f"+{reg['regression_percentage']:.1f}%)"
                    for reg in regressions_detected
                ]
            )
            print(f"⚠️  Performance regressions detected:\\n{regression_details}")
        else:
            print("✅ No performance regressions detected")

        # Performance summary
        print("\\n📊 Performance Summary:")
        for metric, value in self.performance_results.items():
            print(f"  - {metric}: {value:.3f}")


if __name__ == "__main__":
    unittest.main()
