"""
Unit tests for moai_adk.project.documentation module.

Tests cover:
- DocumentationGenerator class
- BrainstormQuestionGenerator class
- AgentContextInjector class
"""

from pathlib import Path
from tempfile import TemporaryDirectory

from moai_adk.project.documentation import (
    AgentContextInjector,
    BrainstormQuestionGenerator,
    DocumentationGenerator,
)


class TestDocumentationGenerator:
    """Test DocumentationGenerator class."""

    def test_init(self):
        """Test initialization."""
        gen = DocumentationGenerator()
        assert gen.templates is not None
        assert "product" in gen.templates
        assert "structure" in gen.templates
        assert "tech" in gen.templates

    def test_load_templates(self):
        """Test template loading."""
        templates = DocumentationGenerator._load_templates()
        assert "product" in templates
        assert "structure" in templates
        assert "tech" in templates
        assert "{vision}" in templates["product"]

    def test_generate_product_md(self):
        """Test product.md generation."""
        gen = DocumentationGenerator()
        responses = {
            "project_vision": "Build fast API",
            "target_users": "Python developers",
            "value_proposition": "Speed and simplicity",
            "roadmap": "v1.0 in Q1 2025",
        }
        content = gen.generate_product_md(responses)

        assert "# Project Vision" in content
        assert "Build fast API" in content
        assert "Python developers" in content

    def test_generate_structure_md(self):
        """Test structure.md generation."""
        gen = DocumentationGenerator()
        responses = {
            "system_architecture": "Microservices",
            "core_components": "API, Database, Cache",
            "relationships": "API calls Database",
            "dependencies": "PostgreSQL, Redis",
        }
        content = gen.generate_structure_md(responses)

        assert "# System Architecture" in content
        assert "Microservices" in content

    def test_generate_tech_md(self):
        """Test tech.md generation."""
        gen = DocumentationGenerator()
        responses = {
            "technology_selection": "Python, FastAPI, PostgreSQL",
            "trade_offs": "FastAPI over Django",
            "performance": "Sub-100ms latency",
            "security": "OAuth2, TLS",
            "setup_guide": "pip install",
        }
        content = gen.generate_tech_md(responses)

        assert "# Technology Stack" in content
        assert "Python, FastAPI, PostgreSQL" in content

    def test_generate_all_documents(self):
        """Test generating all documents."""
        gen = DocumentationGenerator()
        responses = {
            "project_vision": "Fast API",
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
        docs = gen.generate_all_documents(responses)

        assert len(docs) == 3
        assert "product" in docs
        assert "structure" in docs
        assert "tech" in docs

    def test_generate_ai_insights_with_data(self):
        """Test AI insights generation with data."""
        responses = {
            "project_vision": "Build API",
            "target_users": "Developers",
            "value_proposition": "Speed",
        }
        insights = DocumentationGenerator._generate_ai_insights(responses)

        assert "Vision is clear" in insights
        assert "Target user" in insights
        assert "Value proposition" in insights

    def test_generate_ai_insights_empty(self):
        """Test AI insights generation with no data."""
        responses = {}
        insights = DocumentationGenerator._generate_ai_insights(responses)

        assert "pending" in insights.lower()

    def test_save_all_documents(self):
        """Test saving documents to disk."""
        gen = DocumentationGenerator()
        docs = {
            "product": "# Product\nContent",
            "structure": "# Structure\nContent",
            "tech": "# Tech\nContent",
        }

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            gen.save_all_documents(docs, base_path)

            assert (base_path / "product.md").exists()
            assert (base_path / "structure.md").exists()
            assert (base_path / "tech.md").exists()

    def test_load_document(self):
        """Test loading document from disk."""
        gen = DocumentationGenerator()

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            base_path.mkdir(exist_ok=True)

            # Create a test file
            content = "# Test Content"
            (base_path / "product.md").write_text(content)

            # Load it
            loaded = gen.load_document("product.md", base_path)
            assert loaded == content

    def test_load_document_not_found(self):
        """Test loading non-existent document."""
        gen = DocumentationGenerator()

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            loaded = gen.load_document("nonexistent.md", base_path)
            assert loaded is None

    def test_create_minimal_templates(self):
        """Test creating minimal template files."""
        gen = DocumentationGenerator()

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            gen.create_minimal_templates(base_path)

            assert (base_path / "product.md").exists()
            assert (base_path / "structure.md").exists()
            assert (base_path / "tech.md").exists()
            assert "[Add your project vision here]" in (base_path / "product.md").read_text()


class TestBrainstormQuestionGenerator:
    """Test BrainstormQuestionGenerator class."""

    def test_get_quick_questions(self):
        """Test getting quick questions."""
        questions = BrainstormQuestionGenerator.get_quick_questions()

        assert len(questions) == 5
        assert all("id" in q for q in questions)
        assert all("question" in q for q in questions)
        assert all("category" in q for q in questions)

    def test_get_standard_questions(self):
        """Test getting standard questions."""
        questions = BrainstormQuestionGenerator.get_standard_questions()

        assert len(questions) == 10
        assert all("id" in q for q in questions)
        quick = BrainstormQuestionGenerator.get_quick_questions()
        assert len(questions) > len(quick)

    def test_get_deep_questions(self):
        """Test getting deep questions."""
        questions = BrainstormQuestionGenerator.get_deep_questions()

        assert len(questions) == 16
        standard = BrainstormQuestionGenerator.get_standard_questions()
        assert len(questions) > len(standard)

    def test_get_questions_by_depth_quick(self):
        """Test getting questions by depth (quick)."""
        questions = BrainstormQuestionGenerator.get_questions_by_depth("quick")
        assert len(questions) == 5

    def test_get_questions_by_depth_standard(self):
        """Test getting questions by depth (standard)."""
        questions = BrainstormQuestionGenerator.get_questions_by_depth("standard")
        assert len(questions) == 10

    def test_get_questions_by_depth_deep(self):
        """Test getting questions by depth (deep)."""
        questions = BrainstormQuestionGenerator.get_questions_by_depth("deep")
        assert len(questions) == 16

    def test_get_questions_by_depth_invalid(self):
        """Test getting questions with invalid depth."""
        questions = BrainstormQuestionGenerator.get_questions_by_depth("invalid")
        # Should default to quick
        assert len(questions) == 5

    def test_quick_questions_structure(self):
        """Test structure of quick questions."""
        questions = BrainstormQuestionGenerator.get_quick_questions()

        for q in questions:
            assert "id" in q
            assert "question" in q
            assert "category" in q
            assert isinstance(q["id"], str)
            assert isinstance(q["question"], str)
            assert isinstance(q["category"], str)


class TestAgentContextInjector:
    """Test AgentContextInjector class."""

    def test_inject_project_manager_context(self):
        """Test injecting project manager context."""
        agent_config = {"system_prompt": "You are a project manager"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            base_path.mkdir(exist_ok=True)

            # Create a product.md file
            (base_path / "product.md").write_text("# Product\nVision: Build fast API")

            result = AgentContextInjector.inject_project_manager_context(agent_config, base_path)

            assert "system_context" in result
            # Check that project documentation was injected
            assert "Project Documentation" in result["system_context"] or "Build fast API" in result["system_context"]

    def test_inject_ddd_implementer_context(self):
        """Test injecting DDD implementer context."""
        agent_config = {"system_prompt": "You are a DDD implementer"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            base_path.mkdir(exist_ok=True)

            # Create a structure.md file
            (base_path / "structure.md").write_text("# Architecture\nMicroservices")

            result = AgentContextInjector.inject_ddd_implementer_context(agent_config, base_path)

            assert "architecture_context" in result

    def test_inject_domain_expert_context(self):
        """Test injecting domain expert context."""
        agent_config = {"system_prompt": "You are a domain expert"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            base_path.mkdir(exist_ok=True)

            # Create a tech.md file
            (base_path / "tech.md").write_text("# Technology\nPython, FastAPI")

            result = AgentContextInjector.inject_domain_expert_context(agent_config, "backend_expert", base_path)

            assert "tech_context" in result

    def test_inject_context_no_document(self):
        """Test injection when document doesn't exist."""
        agent_config = {"system_prompt": "Test"}

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            base_path.mkdir(exist_ok=True)

            result = AgentContextInjector.inject_project_manager_context(agent_config, base_path)

            # Should return original config when document doesn't exist
            assert result == agent_config

    def test_context_injection_preserves_original(self):
        """Test that context injection doesn't modify original config."""
        original_config = {"system_prompt": "Test", "model": "gpt-4"}
        agent_config = original_config.copy()

        with TemporaryDirectory() as tmpdir:
            base_path = Path(tmpdir)
            base_path.mkdir(exist_ok=True)
            (base_path / "product.md").write_text("# Content")

            result = AgentContextInjector.inject_project_manager_context(agent_config, base_path)

            # Original should not be modified
            assert agent_config == original_config
            # Result should have new content
            assert result != original_config
