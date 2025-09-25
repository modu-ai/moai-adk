"""@TEST:API-GEN-001 API documentation generation tests

Tests for automatic API documentation generation from source code.
Includes docstring parsing and navigation structure creation.

@REQ:API-GEN-001 → @TASK:API-TEST-001
"""

import pytest
from pathlib import Path
from unittest.mock import patch, MagicMock

from moai_adk.core.docs.api_generator import ApiGenerator


class TestApiGeneration:
    """@TEST:API-GEN-001 Test automatic API documentation generation"""

    def test_should_scan_python_modules(self, tmp_path):
        """@TEST:API-GEN-002 Should scan and find Python modules in source tree"""
        # Create mock source structure
        src_dir = tmp_path / "src" / "moai_adk"
        src_dir.mkdir(parents=True)

        (src_dir / "__init__.py").write_text('"""MoAI ADK package"""')
        (src_dir / "core.py").write_text('"""Core module"""')

        generator = ApiGenerator(str(tmp_path), "src")
        modules = generator.scan_modules()

        assert len(modules) >= 2, "Should find at least 2 modules"
        assert any("__init__" in mod.name for mod in modules)
        assert any("core" in mod.name for mod in modules)

    def test_should_parse_docstrings_from_modules(self, tmp_path):
        """@TEST:API-GEN-003 Should parse docstrings and extract documentation"""
        # Create module with docstrings
        src_dir = tmp_path / "src" / "moai_adk"
        src_dir.mkdir(parents=True)

        module_content = '''"""Module docstring"""

def test_function():
    """Function docstring

    Returns:
        bool: Test result
    """
    return True

class TestClass:
    """Class docstring"""

    def method(self):
        """Method docstring"""
        pass
'''
        (src_dir / "example.py").write_text(module_content)

        generator = ApiGenerator(str(tmp_path), "src")

        doc_info = generator.parse_module_docs("src.moai_adk.example")

        assert doc_info["module_doc"] == "Module docstring"
        assert "test_function" in doc_info["functions"]
        assert "TestClass" in doc_info["classes"]

    def test_should_generate_navigation_structure(self, tmp_path):
        """@TEST:API-GEN-004 Should generate proper navigation structure"""
        # Create mock module structure
        src_dir = tmp_path / "src" / "moai_adk"
        src_dir.mkdir(parents=True)

        core_dir = src_dir / "core"
        core_dir.mkdir()

        (src_dir / "__init__.py").touch()
        (core_dir / "__init__.py").touch()
        (core_dir / "manager.py").touch()

        generator = ApiGenerator(str(tmp_path), "src")
        nav_structure = generator.generate_nav_structure()

        assert "moai_adk" in nav_structure
        assert "core" in nav_structure["moai_adk"]

    def test_should_create_api_markdown_files(self, tmp_path):
        """@TEST:API-GEN-005 Should create markdown files for API documentation"""
        # Setup mock source
        src_dir = tmp_path / "src" / "moai_adk"
        src_dir.mkdir(parents=True)
        (src_dir / "__init__.py").write_text('"""Package init"""')

        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()

        generator = ApiGenerator(str(tmp_path), "src")

        generator.generate_api_docs(str(docs_dir))

        # Check if API docs are created
        api_dir = docs_dir / "api"
        assert api_dir.exists(), "API directory should be created"

        # Should have at least one markdown file
        md_files = list(api_dir.glob("**/*.md"))
        assert len(md_files) > 0, "Should create markdown files"

    def test_should_handle_modules_without_docstrings(self, tmp_path):
        """@TEST:API-GEN-006 Should handle modules without docstrings gracefully"""
        src_dir = tmp_path / "src" / "moai_adk"
        src_dir.mkdir(parents=True)

        # Module without docstring
        (src_dir / "no_docs.py").write_text("pass")

        generator = ApiGenerator(str(tmp_path), "src")

        doc_info = generator.parse_module_docs("src.moai_adk.no_docs")

        assert doc_info["module_doc"] == "", "Should handle missing docstrings"
        assert isinstance(doc_info["functions"], dict)
        assert isinstance(doc_info["classes"], dict)

    def test_should_update_mkdocs_nav_with_api(self, tmp_path):
        """@TEST:API-GEN-007 Should update mkdocs.yml navigation with API section"""
        # Create mock mkdocs.yml
        mkdocs_config = tmp_path / "mkdocs.yml"
        mkdocs_config.write_text("""
site_name: Test
nav:
  - Home: index.md
""")

        generator = ApiGenerator(str(tmp_path), "src")

        generator.update_mkdocs_nav(str(mkdocs_config))

        config_content = mkdocs_config.read_text()
        assert "API Reference" in config_content

    def test_should_generate_module_index(self, tmp_path):
        """@TEST:API-GEN-008 Should generate module index with proper links"""
        src_dir = tmp_path / "src" / "moai_adk"
        src_dir.mkdir(parents=True)

        (src_dir / "__init__.py").touch()
        (src_dir / "core.py").touch()

        generator = ApiGenerator(str(tmp_path), "src")

        index_content = generator.generate_module_index()

        assert "# API Reference" in index_content
        assert "moai_adk" in index_content
