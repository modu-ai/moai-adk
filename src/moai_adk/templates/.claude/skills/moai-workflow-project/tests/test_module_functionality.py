"""
Module Functionality Tests

Tests for individual module functionality including:
- DocumentationManager
- LanguageInitializer
- TemplateOptimizer
- Module-specific features and capabilities
"""

import tempfile
import unittest
from pathlib import Path

from moai_menu_project import MoaiMenuProject
from moai_menu_project.modules.documentation_manager import DocumentationManager
from moai_menu_project.modules.language_initializer import LanguageInitializer
from moai_menu_project.modules.template_optimizer import TemplateOptimizer


class TestDocumentationManager(unittest.TestCase):
    """Test DocumentationManager functionality."""

    def setUp(self):
        """Set up test environment."""
        self.test_dir = Path(tempfile.mkdtemp())
        self.config = {
            "project": {"name": "Test Project", "type": "web_application"},
            "language": {"conversation_language": "en"},
        }
        self.doc_manager = DocumentationManager(str(self.test_dir), self.config)

    def tearDown(self):
        """Clean up test environment."""
        import shutil

        shutil.rmtree(self.test_dir, ignore_errors=True)

    def test_documentation_structure_initialization(self):
        """Test documentation structure initialization."""

        result = self.doc_manager.initialize_documentation_structure()

        self.assertTrue(result["success"])
        self.assertIn("created_files", result)

        # Verify docs directory was created
        docs_dir = self.test_dir / "docs"
        self.assertTrue(docs_dir.exists())

    def test_spec_based_documentation_generation(self):
        """Test documentation generation from SPEC data."""

        spec_data = {
            "id": "SPEC-001",
            "title": "User Authentication",
            "description": "Implement secure authentication system",
            "requirements": [
                "User registration with email verification",
                "JWT token generation and validation",
            ],
            "api_endpoints": [
                {
                    "path": "/api/auth/login",
                    "method": "POST",
                    "description": "User login endpoint",
                }
            ],
        }

        result = self.doc_manager.generate_documentation_from_spec(spec_data)

        self.assertTrue(result["success"])
        self.assertIn("feature_docs", result)
        self.assertIn("api_docs", result)

    def test_multilingual_documentation_support(self):
        """Test multilingual documentation capabilities."""

        # Test with Korean language
        config_ko = self.config.copy()
        config_ko["language"] = {"conversation_language": "ko"}

        doc_manager_ko = DocumentationManager(str(self.test_dir), config_ko)
        result = doc_manager_ko.initialize_documentation_structure()

        self.assertTrue(result["success"])

    def test_template_based_generation(self):
        """Test template-based documentation generation."""

        # Test different project types
        project_types = ["web_application", "mobile_application", "cli_tool", "library"]

        for project_type in project_types:
            config = self.config.copy()
            config["project"]["type"] = project_type

            doc_manager = DocumentationManager(str(self.test_dir), config)
            result = doc_manager._generate_product_doc(project_type, "en")

            # Should generate appropriate content
            self.assertIsInstance(result, str)
            self.assertTrue(len(result) > 0)

    def test_export_functionality(self):
        """Test documentation export capabilities."""

        # Test markdown export
        result = self.doc_manager.export_documentation("markdown")
        self.assertTrue(result["success"])

        # Test HTML export
        result = self.doc_manager.export_documentation("html")
        self.assertTrue(result["success"])


class TestLanguageInitializer(unittest.TestCase):
    """Test LanguageInitializer functionality."""

    def setUp(self):
        """Set up test environment."""
        self.test_dir = Path(tempfile.mkdtemp())
        self.config = {
            "project": {"name": "Test Project"},
            "language": {"conversation_language": "en"},
        }
        self.lang_init = LanguageInitializer(str(self.test_dir), self.config)

    def tearDown(self):
        """Clean up test environment."""
        import shutil

        shutil.rmtree(self.test_dir, ignore_errors=True)

    def test_language_detection(self):
        """Test automatic language detection."""

        # Create test files with Korean content
        test_file = self.test_dir / "test.py"
        test_file.write_text(
            """
# 한국어 주석 예제
def calculate_score():
    # 점수 계산 함수
    score = 100  # 최종 점수
    return score
        """,
            encoding="utf-8",
        )

        detected_lang = self.lang_init.detect_project_language()
        self.assertIsInstance(detected_lang, str)

    def test_language_configuration_initialization(self):
        """Test language configuration initialization."""

        result = self.lang_init.initialize_language_configuration(
            language="ko", user_name="테스트 사용자", domains=["backend", "frontend"]
        )

        self.assertTrue(result["success"])
        self.assertEqual(result["language"], "ko")

    def test_multilingual_structure_creation(self):
        """Test multilingual documentation structure creation."""

        result = self.lang_init.create_multilingual_documentation_structure("ko")

        self.assertTrue(result["success"])
        self.assertIn("created_directories", result)

        # Verify directory structure
        ko_docs_dir = self.test_dir / "docs/ko"
        self.assertTrue(ko_docs_dir.exists())

    def test_agent_prompt_localization(self):
        """Test agent prompt localization."""

        base_prompt = "Generate user authentication system"
        result = self.lang_init.localize_agent_prompts(base_prompt, "ko")

        self.assertIsInstance(result, str)
        self.assertTrue(len(result) > 0)

    def test_language_settings_update(self):
        """Test language settings update functionality."""

        updates = {
            "language.conversation_language": "ja",
            "language.agent_prompt_language": "english",
        }

        result = self.lang_init.update_language_settings(updates)

        self.assertTrue(result["success"])

    def test_language_status_reporting(self):
        """Test language status reporting."""

        status = self.lang_init.get_language_status()

        self.assertIn("current_language", status)
        self.assertIn("supported_languages", status)
        self.assertIn("configuration", status)

    def test_token_cost_analysis(self):
        """Test token cost analysis for different languages."""

        # Test cost analysis for different languages
        languages = ["en", "ko", "ja", "zh"]

        for lang in languages:
            cost_info = self.lang_init.get_token_cost_analysis(lang)

            self.assertIn("language", cost_info)
            self.assertIn("cost_impact", cost_info)
            self.assertIsInstance(cost_info["cost_impact"], (int, float))


class TestTemplateOptimizer(unittest.TestCase):
    """Test TemplateOptimizer functionality."""

    def setUp(self):
        """Set up test environment."""
        self.test_dir = Path(tempfile.mkdtemp())
        self.config = {"project": {"name": "Test Project", "type": "web_application"}}
        self.template_opt = TemplateOptimizer(str(self.test_dir), self.config)

        # Create test template files
        self._create_test_templates()

    def tearDown(self):
        """Clean up test environment."""
        import shutil

        shutil.rmtree(self.test_dir, ignore_errors=True)

    def _create_test_templates(self):
        """Create test template files for optimization."""

        templates_dir = self.test_dir / "templates"
        templates_dir.mkdir(exist_ok=True)

        # Create a test template with optimization opportunities
        test_template = """
# Test Template

This template contains redundant content and can be optimized.

## Section 1

Content here...

## Section 1  # Duplicate section

Duplicate content here...

## Section 2

More content...

Extra whitespace:



Complex logic that could be simplified:
{% if condition %}
    {% if nested_condition %}
        Content
    {% endif %}
{% endif %}
        """

        (templates_dir / "test_template.md").write_text(test_template, encoding="utf-8")

    def test_template_analysis(self):
        """Test template analysis functionality."""

        analysis = self.template_opt.analyze_project_templates()

        self.assertTrue(analysis["success"])
        self.assertIn("analyzed_files", analysis)
        self.assertIn("optimization_opportunities", analysis)

    def test_optimization_creation(self):
        """Test optimized template creation."""

        options = {
            "backup_first": True,
            "apply_size_optimizations": True,
            "apply_performance_optimizations": True,
        }

        result = self.template_opt.create_optimized_templates(options)

        self.assertTrue(result["success"])
        self.assertIn("optimization_results", result)

    def test_backup_creation(self):
        """Test template backup creation."""

        backup_result = self.template_opt._create_template_backup("test-backup")

        self.assertTrue(backup_result["success"])
        self.assertIn("backup_path", backup_result)

    def test_benchmark_functionality(self):
        """Test performance benchmarking."""

        benchmark_result = self.template_opt.benchmark_template_performance()

        self.assertTrue(benchmark_result["success"])
        self.assertIn("performance_metrics", benchmark_result)

    def test_optimization_recommendations(self):
        """Test optimization recommendations."""

        # Get recommendations based on analysis
        analysis = self.template_opt.analyze_project_templates()
        recommendations = self.template_opt.get_optimization_recommendations(analysis)

        self.assertIsInstance(recommendations, list)

    def test_complexity_analysis(self):
        """Test template complexity analysis."""

        analysis = self.template_opt.analyze_project_templates()

        # Check complexity metrics
        for file_analysis in analysis.get("analyzed_files", []):
            if "complexity_score" in file_analysis:
                self.assertIsInstance(file_analysis["complexity_score"], (int, float))
                self.assertGreaterEqual(file_analysis["complexity_score"], 0)


class TestModuleInteractions(unittest.TestCase):
    """Test interactions between modules."""

    def setUp(self):
        """Set up test environment."""
        self.test_dir = Path(tempfile.mkdtemp())
        self.project = MoaiMenuProject(str(self.test_dir))

    def tearDown(self):
        """Clean up test environment."""
        import shutil

        shutil.rmtree(self.test_dir, ignore_errors=True)

    def test_language_documentation_integration(self):
        """Test integration between language and documentation modules."""

        # Initialize with Korean language
        self.project.initialize_complete_project(language="ko", project_type="web_application")

        # Generate documentation
        spec_data = {
            "id": "SPEC-001",
            "title": "테스트 기능",
            "description": "한국어 테스트 기능 구현",
        }

        docs_result = self.project.generate_documentation_from_spec(spec_data)

        # Should have localization information
        self.assertTrue(docs_result["success"])

    def test_optimization_documentation_integration(self):
        """Test integration between optimization and documentation."""

        # Initialize and optimize
        self.project.initialize_complete_project(optimization_enabled=True)

        # Generate documentation after optimization
        docs_result = self.project.documentation_manager.export_documentation("markdown")

        self.assertTrue(docs_result["success"])


if __name__ == "__main__":
    unittest.main()
