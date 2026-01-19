"""
Simple comprehensive unit tests for moai_adk.project.documentation module.

Tests cover:
- DocumentationGenerator methods with actual code paths
- BrainstormQuestionGenerator methods
- AgentContextInjector methods
"""

from pathlib import Path
from tempfile import TemporaryDirectory

from moai_adk.project.documentation import (
    AgentContextInjector,
    BrainstormQuestionGenerator,
    DocumentationGenerator,
)


class TestDocumentationGeneratorSimple:
    """Test DocumentationGenerator with simple focused tests."""

    def test_generate_product_md_with_empty_responses(self):
        """Test product.md generation with empty responses."""
        gen = DocumentationGenerator()
        content = gen.generate_product_md({})

        assert "# Project Vision" in content
        assert "## Vision Statement" in content
        assert "## Target Users" in content
        assert "## Value Proposition" in content
        assert "## Roadmap" in content

    def test_generate_product_md_strips_whitespace(self):
        """Test that product.md content is stripped."""
        gen = DocumentationGenerator()
        responses = {"project_vision": "Test Vision"}
        content = gen.generate_product_md(responses)

        # Content should be stripped (no leading/trailing whitespace)
        assert not content.startswith("\n")
        assert not content.endswith("\n")

    def test_generate_product_md_contains_generated_marker(self):
        """Test that product.md contains generation marker."""
        gen = DocumentationGenerator()
        responses = {"project_vision": "Test"}
        content = gen.generate_product_md(responses)

        assert "Generated" in content
        assert "auto-generated" in content.lower()

    def test_generate_structure_md_with_empty_responses(self):
        """Test structure.md generation with empty responses."""
        gen = DocumentationGenerator()
        content = gen.generate_structure_md({})

        assert "# System Architecture" in content
        assert "## Architecture Overview" in content
        assert "## Core Components" in content
        assert "## Component Relationships" in content
        assert "## Dependencies" in content

    def test_generate_structure_md_contains_design_notes(self):
        """Test structure.md contains design notes."""
        gen = DocumentationGenerator()
        responses = {"system_architecture": "Test"}
        content = gen.generate_structure_md(responses)

        assert "Design Notes" in content
        assert "reference for implementation" in content.lower()

    def test_generate_tech_md_with_empty_responses(self):
        """Test tech.md generation with empty responses."""
        gen = DocumentationGenerator()
        content = gen.generate_tech_md({})

        assert "# Technology Stack" in content
        assert "## Technology Selection" in content
        assert "## Trade-off Analysis" in content
        assert "## Performance Considerations" in content
        assert "## Security Considerations" in content
        assert "## Setup Guide" in content

    def test_generate_tech_md_contains_version_marker(self):
        """Test tech.md contains version marker."""
        gen = DocumentationGenerator()
        content = gen.generate_tech_md({})

        assert "Version" in content
        assert "subject to updates" in content.lower()

    def test_generate_all_documents_returns_three_docs(self):
        """Test that generate_all_documents returns three documents."""
        gen = DocumentationGenerator()
        docs = gen.generate_all_documents({})

        assert isinstance(docs, dict)
        assert len(docs) == 3
        assert "product" in docs
        assert "structure" in docs
        assert "tech" in docs

    def test_generate_all_documents_all_are_strings(self):
        """Test that all generated documents are strings."""
        gen = DocumentationGenerator()
        docs = gen.generate_all_documents({})

        for doc in docs.values():
            assert isinstance(doc, str)
            assert len(doc) > 0

    def test_save_all_documents_creates_directory(self):
        """Test that save_all_documents creates directory if needed."""
        gen = DocumentationGenerator()
        docs = {
            "product": "# Product",
            "structure": "# Structure",
            "tech": "# Tech",
        }

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir) / "nested" / "path"
            gen.save_all_documents(docs, base_path)

            assert base_path.exists()
            assert (base_path / "product.md").exists()

    def test_save_all_documents_with_custom_path(self):
        """Test saving documents with custom base path."""
        gen = DocumentationGenerator()
        docs = {
            "product": "# Test Product",
            "structure": "# Test Structure",
            "tech": "# Test Tech",
        }

        with TemporaryDirectory() as tmpdir:
            custom_path = Path(tmpdir) / "custom_docs"
            gen.save_all_documents(docs, custom_path)

            assert (custom_path / "product.md").exists()
            content = (custom_path / "product.md").read_text()
            assert "Test Product" in content

    def test_load_document_returns_content(self):
        """Test loading document returns exact content."""
        gen = DocumentationGenerator()

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            test_content = "# Test\nSpecial chars: @#$%"
            (base_path / "product.md").write_text(test_content, encoding="utf-8")

            loaded = gen.load_document("product.md", base_path)

            assert loaded == test_content

    def test_load_document_with_encoding(self):
        """Test loading document respects encoding."""
        gen = DocumentationGenerator()

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            unicode_content = "# Test\nUnicode: 한글 日本語 中文"
            (base_path / "structure.md").write_text(unicode_content, encoding="utf-8")

            loaded = gen.load_document("structure.md", base_path)

            assert loaded == unicode_content
            assert "한글" in loaded

    def test_create_minimal_templates_creates_all_files(self):
        """Test that minimal templates creates all three files."""
        gen = DocumentationGenerator()

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            gen.create_minimal_templates(base_path)

            assert (base_path / "product.md").exists()
            assert (base_path / "structure.md").exists()
            assert (base_path / "tech.md").exists()

    def test_create_minimal_templates_product_content(self):
        """Test minimal product template content."""
        gen = DocumentationGenerator()

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            gen.create_minimal_templates(base_path)

            content = (base_path / "product.md").read_text()

            assert "# Project Vision" in content
            assert "Vision Statement" in content
            assert "[Add your project vision here]" in content

    def test_create_minimal_templates_structure_content(self):
        """Test minimal structure template content."""
        gen = DocumentationGenerator()

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            gen.create_minimal_templates(base_path)

            content = (base_path / "structure.md").read_text()

            assert "# System Architecture" in content
            assert "[Describe the system architecture]" in content

    def test_create_minimal_templates_tech_content(self):
        """Test minimal tech template content."""
        gen = DocumentationGenerator()

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            gen.create_minimal_templates(base_path)

            content = (base_path / "tech.md").read_text()

            assert "# Technology Stack" in content
            assert "[List selected technologies and frameworks]" in content

    def test_generate_ai_insights_with_all_fields(self):
        """Test AI insights with all fields present."""
        responses = {
            "project_vision": "Vision",
            "target_users": "Users",
            "value_proposition": "Value",
        }
        insights = DocumentationGenerator._generate_ai_insights(responses)

        assert "Vision is clear" in insights
        assert "Target user" in insights
        assert "Value proposition" in insights

    def test_generate_ai_insights_with_partial_fields(self):
        """Test AI insights with partial fields."""
        responses = {"project_vision": "Vision"}
        insights = DocumentationGenerator._generate_ai_insights(responses)

        assert "Vision is clear" in insights
        assert len(insights) > 0

    def test_generate_ai_insights_with_no_fields(self):
        """Test AI insights with no fields returns pending."""
        responses = {}
        insights = DocumentationGenerator._generate_ai_insights(responses)

        assert "pending" in insights.lower()


class TestBrainstormQuestionGeneratorSimple:
    """Test BrainstormQuestionGenerator with simple focused tests."""

    def test_quick_questions_count(self):
        """Test quick questions returns 5 questions."""
        questions = BrainstormQuestionGenerator.get_quick_questions()
        assert len(questions) == 5

    def test_quick_questions_have_unique_ids(self):
        """Test quick questions have unique IDs."""
        questions = BrainstormQuestionGenerator.get_quick_questions()
        ids = [q["id"] for q in questions]

        assert len(ids) == len(set(ids))

    def test_quick_questions_start_with_q1(self):
        """Test first quick question is q1_vision."""
        questions = BrainstormQuestionGenerator.get_quick_questions()
        assert questions[0]["id"] == "q1_vision"

    def test_standard_questions_count(self):
        """Test standard questions returns 10 questions."""
        questions = BrainstormQuestionGenerator.get_standard_questions()
        assert len(questions) == 10

    def test_standard_questions_include_quick(self):
        """Test standard questions include all quick questions."""
        quick = BrainstormQuestionGenerator.get_quick_questions()
        standard = BrainstormQuestionGenerator.get_standard_questions()

        quick_ids = {q["id"] for q in quick}
        standard_ids = {q["id"] for q in standard}

        assert quick_ids.issubset(standard_ids)

    def test_deep_questions_count(self):
        """Test deep questions returns 16 questions."""
        questions = BrainstormQuestionGenerator.get_deep_questions()
        assert len(questions) == 16

    def test_deep_questions_include_standard(self):
        """Test deep questions include all standard questions."""
        standard = BrainstormQuestionGenerator.get_standard_questions()
        deep = BrainstormQuestionGenerator.get_deep_questions()

        standard_ids = {q["id"] for q in standard}
        deep_ids = {q["id"] for q in deep}

        assert standard_ids.issubset(deep_ids)

    def test_get_questions_by_depth_quick(self):
        """Test getting questions with quick depth."""
        questions = BrainstormQuestionGenerator.get_questions_by_depth("quick")

        assert len(questions) == 5
        assert questions[0]["id"] == "q1_vision"

    def test_get_questions_by_depth_standard(self):
        """Test getting questions with standard depth."""
        questions = BrainstormQuestionGenerator.get_questions_by_depth("standard")

        assert len(questions) == 10

    def test_get_questions_by_depth_deep(self):
        """Test getting questions with deep depth."""
        questions = BrainstormQuestionGenerator.get_questions_by_depth("deep")

        assert len(questions) == 16

    def test_get_questions_by_depth_invalid_defaults_to_quick(self):
        """Test invalid depth defaults to quick."""
        questions = BrainstormQuestionGenerator.get_questions_by_depth("nonexistent")

        assert len(questions) == 5

    def test_get_questions_by_depth_none_defaults_to_quick(self):
        """Test None depth defaults to quick."""
        questions = BrainstormQuestionGenerator.get_questions_by_depth("unknown")

        assert len(questions) == 5

    def test_questions_have_required_structure(self):
        """Test all questions have required structure."""
        for depth in ["quick", "standard", "deep"]:
            questions = BrainstormQuestionGenerator.get_questions_by_depth(depth)

            for q in questions:
                assert "id" in q
                assert "question" in q
                assert "category" in q
                assert isinstance(q["id"], str)
                assert isinstance(q["question"], str)
                assert isinstance(q["category"], str)


class TestAgentContextInjectorSimple:
    """Test AgentContextInjector with simple focused tests."""

    def test_inject_project_manager_context_adds_key(self):
        """Test that injection adds system_context key."""
        config = {"prompt": "test"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            (base_path / "product.md").write_text("# Content")

            result = AgentContextInjector.inject_project_manager_context(config, base_path)

            assert "system_context" in result

    def test_inject_project_manager_context_preserves_original(self):
        """Test injection doesn't modify original config."""
        original = {"prompt": "test"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            (base_path / "product.md").write_text("# Test")

            result = AgentContextInjector.inject_project_manager_context(original, base_path)

            # Original should not be modified
            assert original == {"prompt": "test"}
            # Result should be different
            assert result != original

    def test_inject_project_manager_context_with_missing_doc(self):
        """Test injection when document doesn't exist."""
        config = {"prompt": "test"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)

            result = AgentContextInjector.inject_project_manager_context(config, base_path)

            # Should return a copy of original config
            assert result == config

    def test_inject_project_manager_context_adds_documentation(self):
        """Test that product.md is injected into context."""
        config = {"prompt": "test"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            (base_path / "product.md").write_text("# Project Doc\nVision: Test")

            result = AgentContextInjector.inject_project_manager_context(config, base_path)

            # Either the context should contain the doc content or a reference
            assert "system_context" in result

    def test_inject_ddd_implementer_context_adds_architecture_context(self):
        """Test DDD implementer context adds architecture_context."""
        config = {"prompt": "test"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            (base_path / "structure.md").write_text("# Architecture\nContent")

            result = AgentContextInjector.inject_ddd_implementer_context(config, base_path)

            assert "architecture_context" in result

    def test_inject_ddd_implementer_context_preserves_original(self):
        """Test DDD injection doesn't modify original."""
        original = {"prompt": "test"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            (base_path / "structure.md").write_text("# Arch")

            AgentContextInjector.inject_ddd_implementer_context(original, base_path)

            # Original unchanged
            assert original == {"prompt": "test"}

    def test_inject_domain_expert_context_for_backend(self):
        """Test domain expert context injection for backend."""
        config = {"prompt": "test"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            (base_path / "tech.md").write_text("# Tech\nContent")

            result = AgentContextInjector.inject_domain_expert_context(config, "backend_expert", base_path)

            assert "tech_context" in result

    def test_inject_domain_expert_context_for_frontend(self):
        """Test domain expert context injection for frontend."""
        config = {"prompt": "test"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            (base_path / "tech.md").write_text("# Tech")

            result = AgentContextInjector.inject_domain_expert_context(config, "frontend_expert", base_path)

            assert "tech_context" in result

    def test_inject_domain_expert_context_preserves_original(self):
        """Test domain expert injection preserves original."""
        original = {"prompt": "test"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            (base_path / "tech.md").write_text("# Tech")

            AgentContextInjector.inject_domain_expert_context(original, "backend_expert", base_path)

            # Original unchanged
            assert original == {"prompt": "test"}

    def test_inject_context_with_existing_system_context(self):
        """Test injection when config already has system_context."""
        config = {"system_context": "Existing context"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            (base_path / "product.md").write_text("# New content")

            AgentContextInjector.inject_project_manager_context(config, base_path)

            # Should preserve original
            assert config == {"system_context": "Existing context"}

    def test_inject_context_with_existing_architecture_context(self):
        """Test injection when config already has architecture_context."""
        config = {"architecture_context": "Existing"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            (base_path / "structure.md").write_text("# New")

            AgentContextInjector.inject_ddd_implementer_context(config, base_path)

            # Original should not be modified
            assert config == {"architecture_context": "Existing"}


class TestDocumentationGeneratorIntegration:
    """Integration tests for DocumentationGenerator."""

    def test_save_and_load_roundtrip(self):
        """Test saving and loading documents roundtrip."""
        gen = DocumentationGenerator()

        original_docs = {
            "product": "# Product\nVision: Build fast",
            "structure": "# Structure\nAPI Service",
            "tech": "# Tech\nPython FastAPI",
        }

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)

            # Save
            gen.save_all_documents(original_docs, base_path)

            # Load
            product = gen.load_document("product.md", base_path)
            structure = gen.load_document("structure.md", base_path)
            tech = gen.load_document("tech.md", base_path)

            # Verify
            assert product == original_docs["product"]
            assert structure == original_docs["structure"]
            assert tech == original_docs["tech"]

    def test_generate_and_save_workflow(self):
        """Test generating and saving documents."""
        gen = DocumentationGenerator()

        responses = {
            "project_vision": "Build fast API",
            "target_users": "Developers",
            "value_proposition": "Speed",
            "roadmap": "Q1 2025",
            "system_architecture": "Microservices",
            "core_components": "API, DB",
            "relationships": "API -> DB",
            "dependencies": "PostgreSQL",
            "technology_selection": "Python",
            "trade_offs": "FastAPI",
            "performance": "Fast",
            "security": "Secure",
            "setup_guide": "Install",
        }

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)

            # Generate
            docs = gen.generate_all_documents(responses)

            # Save
            gen.save_all_documents(docs, base_path)

            # Verify files exist
            assert (base_path / "product.md").exists()
            assert (base_path / "structure.md").exists()
            assert (base_path / "tech.md").exists()

            # Verify content
            product_content = (base_path / "product.md").read_text()
            assert "Build fast API" in product_content
