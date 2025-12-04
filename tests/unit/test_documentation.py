"""Tests for moai_adk.project.documentation module."""

import pytest
import tempfile
from pathlib import Path
from unittest.mock import patch, MagicMock

from moai_adk.project.documentation import (
    DocumentationGenerator,
    BrainstormQuestionGenerator,
    AgentContextInjector,
)


class TestDocumentationGeneratorInit:
    """Test DocumentationGenerator initialization."""

    def test_init_creates_instance(self):
        """Test that init creates a valid instance."""
        gen = DocumentationGenerator()
        assert isinstance(gen, DocumentationGenerator)

    def test_init_loads_templates(self):
        """Test that init loads templates."""
        gen = DocumentationGenerator()
        assert hasattr(gen, "templates")
        assert isinstance(gen.templates, dict)

    def test_templates_have_required_keys(self):
        """Test that templates have all required keys."""
        gen = DocumentationGenerator()
        assert "product" in gen.templates
        assert "structure" in gen.templates
        assert "tech" in gen.templates

    def test_templates_are_strings(self):
        """Test that all templates are strings."""
        gen = DocumentationGenerator()
        for key, template in gen.templates.items():
            assert isinstance(template, str)
            assert len(template) > 0


class TestGenerateProductMd:
    """Test product.md generation."""

    def test_generate_product_md_basic(self):
        """Test basic product.md generation."""
        gen = DocumentationGenerator()
        responses = {
            "project_vision": "Build fast API",
            "target_users": "Python developers",
            "value_proposition": "Speed and simplicity",
            "roadmap": "v1.0 in Q1 2025",
        }

        content = gen.generate_product_md(responses)

        assert isinstance(content, str)
        assert "Build fast API" in content
        assert "Python developers" in content
        assert "Speed and simplicity" in content

    def test_generate_product_md_includes_headers(self):
        """Test that product.md includes proper headers."""
        gen = DocumentationGenerator()
        responses = {
            "project_vision": "Test vision",
            "target_users": "Test users",
        }

        content = gen.generate_product_md(responses)

        assert "# Project Vision" in content
        assert "## Vision Statement" in content
        assert "## Target Users" in content

    def test_generate_product_md_with_empty_responses(self):
        """Test product.md generation with empty responses."""
        gen = DocumentationGenerator()
        responses = {}

        content = gen.generate_product_md(responses)

        assert isinstance(content, str)
        assert "# Project Vision" in content

    def test_generate_product_md_includes_ai_insights(self):
        """Test that product.md includes AI insights."""
        gen = DocumentationGenerator()
        responses = {
            "project_vision": "Test",
            "target_users": "Users",
            "value_proposition": "Value",
        }

        content = gen.generate_product_md(responses)

        assert "AI Analysis" in content or "AI analysis" in content or "Insights" in content

    def test_generate_product_md_partial_responses(self):
        """Test product.md with partial responses."""
        gen = DocumentationGenerator()
        responses = {"project_vision": "Only vision provided"}

        content = gen.generate_product_md(responses)

        assert "Only vision provided" in content
        assert isinstance(content, str)


class TestGenerateStructureMd:
    """Test structure.md generation."""

    def test_generate_structure_md_basic(self):
        """Test basic structure.md generation."""
        gen = DocumentationGenerator()
        responses = {
            "system_architecture": "Microservices",
            "core_components": "API, Database, Cache",
            "relationships": "API calls Database",
            "dependencies": "PostgreSQL, Redis",
        }

        content = gen.generate_structure_md(responses)

        assert isinstance(content, str)
        assert "Microservices" in content
        assert "API, Database, Cache" in content

    def test_generate_structure_md_includes_headers(self):
        """Test that structure.md includes proper headers."""
        gen = DocumentationGenerator()
        responses = {"system_architecture": "Test"}

        content = gen.generate_structure_md(responses)

        assert "# System Architecture" in content
        assert "## Architecture Overview" in content
        assert "## Core Components" in content

    def test_generate_structure_md_with_empty_responses(self):
        """Test structure.md with empty responses."""
        gen = DocumentationGenerator()
        responses = {}

        content = gen.generate_structure_md(responses)

        assert isinstance(content, str)
        assert "# System Architecture" in content


class TestGenerateTechMd:
    """Test tech.md generation."""

    def test_generate_tech_md_basic(self):
        """Test basic tech.md generation."""
        gen = DocumentationGenerator()
        responses = {
            "technology_selection": "Python, FastAPI, PostgreSQL",
            "trade_offs": "FastAPI over Django",
            "performance": "Sub-100ms latency",
            "security": "OAuth2, TLS",
        }

        content = gen.generate_tech_md(responses)

        assert isinstance(content, str)
        assert "Python, FastAPI, PostgreSQL" in content
        assert "FastAPI over Django" in content

    def test_generate_tech_md_includes_headers(self):
        """Test that tech.md includes proper headers."""
        gen = DocumentationGenerator()
        responses = {"technology_selection": "Test"}

        content = gen.generate_tech_md(responses)

        assert "# Technology Stack" in content
        assert "## Technology Selection" in content
        assert "## Trade-off Analysis" in content

    def test_generate_tech_md_with_empty_responses(self):
        """Test tech.md with empty responses."""
        gen = DocumentationGenerator()
        responses = {}

        content = gen.generate_tech_md(responses)

        assert isinstance(content, str)
        assert "# Technology Stack" in content


class TestGenerateAllDocuments:
    """Test all documents generation."""

    def test_generate_all_documents_returns_dict(self):
        """Test that generate_all_documents returns a dict."""
        gen = DocumentationGenerator()
        responses = {
            "project_vision": "Vision",
            "system_architecture": "Architecture",
            "technology_selection": "Tech",
        }

        docs = gen.generate_all_documents(responses)

        assert isinstance(docs, dict)

    def test_generate_all_documents_has_three_keys(self):
        """Test that result has product, structure, tech keys."""
        gen = DocumentationGenerator()
        responses = {}

        docs = gen.generate_all_documents(responses)

        assert "product" in docs
        assert "structure" in docs
        assert "tech" in docs
        assert len(docs) == 3

    def test_generate_all_documents_values_are_strings(self):
        """Test that all document values are strings."""
        gen = DocumentationGenerator()
        responses = {}

        docs = gen.generate_all_documents(responses)

        for doc_content in docs.values():
            assert isinstance(doc_content, str)

    def test_generate_all_documents_with_full_responses(self):
        """Test generate_all_documents with complete responses."""
        gen = DocumentationGenerator()
        responses = {
            "project_vision": "Fast API framework",
            "target_users": "Python developers",
            "value_proposition": "Speed and simplicity",
            "roadmap": "v1.0 in 2025",
            "system_architecture": "Microservices",
            "core_components": "API, Database, Cache",
            "relationships": "API calls DB",
            "dependencies": "PostgreSQL, Redis",
            "technology_selection": "Python, FastAPI",
            "trade_offs": "FastAPI over Django",
            "performance": "Sub-100ms",
            "security": "OAuth2",
            "setup_guide": "pip install",
        }

        docs = gen.generate_all_documents(responses)

        assert "Fast API framework" in docs["product"]
        assert "Microservices" in docs["structure"]
        assert "Python, FastAPI" in docs["tech"]


class TestSaveAllDocuments:
    """Test saving documents to disk."""

    def test_save_all_documents_creates_files(self):
        """Test that save_all_documents creates files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            gen = DocumentationGenerator()

            documents = {
                "product": "# Product\nContent",
                "structure": "# Structure\nContent",
                "tech": "# Tech\nContent",
            }

            gen.save_all_documents(documents, base_path)

            assert (base_path / "product.md").exists()
            assert (base_path / "structure.md").exists()
            assert (base_path / "tech.md").exists()

    def test_save_all_documents_creates_directory(self):
        """Test that save_all_documents creates missing directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir) / "nested/path"
            gen = DocumentationGenerator()

            documents = {
                "product": "Content",
                "structure": "Content",
                "tech": "Content",
            }

            gen.save_all_documents(documents, base_path)

            assert base_path.exists()

    def test_save_all_documents_writes_content(self):
        """Test that documents are written with correct content."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            gen = DocumentationGenerator()

            documents = {
                "product": "# Product\nTest content",
                "structure": "# Structure\nArchitecture",
                "tech": "# Tech\nStack info",
            }

            gen.save_all_documents(documents, base_path)

            # Verify content
            product_content = (base_path / "product.md").read_text(encoding="utf-8")
            assert "Test content" in product_content

    def test_save_all_documents_with_utf8(self):
        """Test that save preserves UTF-8 characters."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            gen = DocumentationGenerator()

            documents = {
                "product": "# 제품\n한글 콘텐츠",
                "structure": "# 構造",
                "tech": "# 技術",
            }

            gen.save_all_documents(documents, base_path)

            product_content = (base_path / "product.md").read_text(encoding="utf-8")
            assert "제품" in product_content
            assert "한글" in product_content

    def test_save_all_documents_partial_documents(self):
        """Test saving with only some documents."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            gen = DocumentationGenerator()

            documents = {"product": "# Product"}

            gen.save_all_documents(documents, base_path)

            assert (base_path / "product.md").exists()


class TestLoadDocument:
    """Test loading documents from disk."""

    def test_load_document_returns_content(self):
        """Test that load_document returns file content."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            base_path.mkdir(exist_ok=True)

            # Create a document
            doc_path = base_path / "product.md"
            doc_path.write_text("# Product\nContent", encoding="utf-8")

            gen = DocumentationGenerator()
            content = gen.load_document("product.md", base_path)

            assert content == "# Product\nContent"

    def test_load_document_nonexistent_returns_none(self):
        """Test that loading nonexistent document returns None."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            gen = DocumentationGenerator()

            result = gen.load_document("nonexistent.md", base_path)

            assert result is None

    def test_load_document_preserves_utf8(self):
        """Test that load_document preserves UTF-8."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            base_path.mkdir(exist_ok=True)

            # Create a document with UTF-8
            doc_path = base_path / "product.md"
            doc_path.write_text("# 제품\n한글 내용", encoding="utf-8")

            gen = DocumentationGenerator()
            content = gen.load_document("product.md", base_path)

            assert "제품" in content
            assert "한글" in content


class TestBrainstormQuestionGenerator:
    """Test BrainstormQuestionGenerator."""

    def test_get_quick_questions_returns_list(self):
        """Test that get_quick_questions returns list."""
        questions = BrainstormQuestionGenerator.get_quick_questions()
        assert isinstance(questions, list)
        assert len(questions) > 0

    def test_quick_questions_have_required_fields(self):
        """Test that quick questions have required fields."""
        questions = BrainstormQuestionGenerator.get_quick_questions()

        for q in questions:
            assert "id" in q
            assert "question" in q
            assert "category" in q

    def test_get_standard_questions_includes_quick(self):
        """Test that standard questions include quick questions."""
        quick = BrainstormQuestionGenerator.get_quick_questions()
        standard = BrainstormQuestionGenerator.get_standard_questions()

        assert len(standard) > len(quick)
        # First items should be quick questions
        for i, q in enumerate(quick):
            assert standard[i]["id"] == q["id"]

    def test_get_deep_questions_includes_standard(self):
        """Test that deep questions include standard questions."""
        standard = BrainstormQuestionGenerator.get_standard_questions()
        deep = BrainstormQuestionGenerator.get_deep_questions()

        assert len(deep) > len(standard)

    def test_get_questions_by_depth(self):
        """Test get_questions_by_depth for all depths."""
        quick = BrainstormQuestionGenerator.get_questions_by_depth("quick")
        standard = BrainstormQuestionGenerator.get_questions_by_depth("standard")
        deep = BrainstormQuestionGenerator.get_questions_by_depth("deep")

        assert len(quick) < len(standard) < len(deep)

    def test_get_questions_by_depth_unknown_returns_quick(self):
        """Test that unknown depth returns quick questions."""
        result = BrainstormQuestionGenerator.get_questions_by_depth("unknown")
        quick = BrainstormQuestionGenerator.get_quick_questions()

        assert result == quick


class TestAgentContextInjector:
    """Test AgentContextInjector."""

    def test_inject_project_manager_context_creates_copy(self):
        """Test that inject_project_manager_context creates a copy."""
        agent_config = {"name": "project_manager", "model": "claude-3"}
        result = AgentContextInjector.inject_project_manager_context(agent_config)

        # Should be a copy, not same object
        assert result is not agent_config

    def test_inject_project_manager_context_preserves_original(self):
        """Test that original config is not modified."""
        agent_config = {"name": "project_manager"}
        original_keys = set(agent_config.keys())

        AgentContextInjector.inject_project_manager_context(agent_config)

        # Original should be unchanged
        assert set(agent_config.keys()) == original_keys

    def test_inject_project_manager_context_with_existing_context(self):
        """Test injection with existing context."""
        agent_config = {"system_context": "Existing context"}

        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            (base_path / "product.md").write_text("# Product", encoding="utf-8")

            result = AgentContextInjector.inject_project_manager_context(agent_config, base_path)

            # Should contain both existing and new context
            assert "Existing context" in result["system_context"]
            assert "Project Documentation" in result["system_context"]

    def test_inject_tdd_implementer_context(self):
        """Test inject_tdd_implementer_context."""
        agent_config = {"name": "tdd_implementer"}

        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            result = AgentContextInjector.inject_tdd_implementer_context(agent_config, base_path)

            assert result is not agent_config
            assert isinstance(result, dict)

    def test_inject_domain_expert_context(self):
        """Test inject_domain_expert_context."""
        agent_config = {"name": "backend_expert"}

        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            result = AgentContextInjector.inject_domain_expert_context(
                agent_config, "backend_expert", base_path
            )

            assert result is not agent_config
            assert isinstance(result, dict)


class TestCreateMinimalTemplates:
    """Test create_minimal_templates."""

    def test_create_minimal_templates_creates_files(self):
        """Test that create_minimal_templates creates files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            gen = DocumentationGenerator()

            gen.create_minimal_templates(base_path)

            assert (base_path / "product.md").exists()
            assert (base_path / "structure.md").exists()
            assert (base_path / "tech.md").exists()

    def test_create_minimal_templates_content(self):
        """Test that minimal templates have placeholder content."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            gen = DocumentationGenerator()

            gen.create_minimal_templates(base_path)

            product = (base_path / "product.md").read_text(encoding="utf-8")
            assert "# Project Vision" in product
            assert "[Add your project vision here]" in product

    def test_create_minimal_templates_creates_directory(self):
        """Test that create_minimal_templates creates directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir) / "nested/path"
            gen = DocumentationGenerator()

            gen.create_minimal_templates(base_path)

            assert base_path.exists()


class TestGenerateAiInsights:
    """Test _generate_ai_insights static method."""

    def test_generate_ai_insights_with_all_keys(self):
        """Test insights generation with all required keys."""
        responses = {
            "project_vision": "Vision",
            "target_users": "Users",
            "value_proposition": "Value",
        }

        insights = DocumentationGenerator._generate_ai_insights(responses)

        assert isinstance(insights, str)
        assert "Vision" in insights or "vision" in insights

    def test_generate_ai_insights_with_empty_responses(self):
        """Test insights with empty responses."""
        responses = {}

        insights = DocumentationGenerator._generate_ai_insights(responses)

        assert isinstance(insights, str)

    def test_generate_ai_insights_partial_responses(self):
        """Test insights with partial responses."""
        responses = {"project_vision": "Vision only"}

        insights = DocumentationGenerator._generate_ai_insights(responses)

        assert isinstance(insights, str)
        assert len(insights) > 0
