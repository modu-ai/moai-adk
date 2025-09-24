"""@TEST:DOCS-SITE-001 Basic documentation site tests

Tests for MkDocs Material based documentation site functionality.
Following TRUST 5 principles for TDD implementation.

@REQ:DOCS-SITE-001 → @TASK:DOCS-TEST-001
"""
import pytest
import os
from pathlib import Path
from unittest.mock import patch, MagicMock

from moai_adk.core.docs.documentation_builder import DocumentationBuilder


class TestDocumentationSite:
    """@TEST:DOCS-SITE-001 Test basic documentation site functionality"""

    def test_should_create_mkdocs_config_file(self, tmp_path):
        """@TEST:DOCS-SITE-002 Should create mkdocs.yml configuration"""
        builder = DocumentationBuilder(str(tmp_path))

        builder.initialize_site()

        mkdocs_config = tmp_path / "mkdocs.yml"
        assert mkdocs_config.exists(), "mkdocs.yml should be created"

        # Check Material theme configuration
        config_content = mkdocs_config.read_text()
        assert "theme:" in config_content
        assert "material" in config_content

    def test_should_create_docs_directory_structure(self, tmp_path):
        """@TEST:DOCS-SITE-003 Should create docs directory with proper structure"""
        builder = DocumentationBuilder(str(tmp_path))

        builder.initialize_site()

        docs_dir = tmp_path / "docs"
        assert docs_dir.exists(), "docs directory should be created"

        # Check required directories
        assert (docs_dir / "getting-started").exists()
        assert (docs_dir / "guide").exists()
        assert (docs_dir / "development").exists()
        assert (docs_dir / "examples").exists()

    def test_should_create_essential_documentation_files(self, tmp_path):
        """@TEST:DOCS-SITE-004 Should create essential documentation files"""
        builder = DocumentationBuilder(str(tmp_path))

        builder.initialize_site()

        docs_dir = tmp_path / "docs"

        # Check essential files
        assert (docs_dir / "index.md").exists()
        assert (docs_dir / "getting-started" / "installation.md").exists()
        assert (docs_dir / "guide" / "workflow.md").exists()

    def test_should_validate_mkdocs_configuration(self, tmp_path):
        """@TEST:DOCS-SITE-005 Should validate MkDocs configuration is valid"""
        builder = DocumentationBuilder(str(tmp_path))

        builder.initialize_site()

        # Should be able to validate config without errors
        is_valid = builder.validate_config()
        assert is_valid, "MkDocs configuration should be valid"

    def test_should_build_documentation_successfully(self, tmp_path):
        """@TEST:DOCS-SITE-006 Should build documentation without errors"""
        builder = DocumentationBuilder(str(tmp_path))
        builder.initialize_site()

        # Mock MkDocs build process by patching at the module level
        with patch.object(builder, '_build_status', {"success": True, "error": None}):
            result = builder.build_docs()

            assert result is True, "Documentation build should succeed"

    def test_should_handle_missing_docs_directory(self, tmp_path):
        """@TEST:DOCS-SITE-007 Should handle missing docs directory gracefully"""
        builder = DocumentationBuilder(str(tmp_path))

        # Should return False but not raise exception
        result = builder.build_docs()  # Without initialization
        assert result is False, "Should return False for missing docs directory"

        status = builder.get_build_status()
        assert "docs directory not found" in status["error"]

    def test_should_provide_build_status_information(self, tmp_path):
        """@TEST:DOCS-SITE-008 Should provide build status and error information"""
        builder = DocumentationBuilder(str(tmp_path))
        builder.initialize_site()

        # Test successful build status
        result = builder.build_docs()
        status = builder.get_build_status()

        assert result is True, "Build should succeed"
        assert status["success"] is True
        assert status["error"] is None

        # Test failed build status - simulate by calling on missing docs
        builder2 = DocumentationBuilder(str(tmp_path / "nonexistent"))
        result2 = builder2.build_docs()
        status2 = builder2.get_build_status()

        assert result2 is False, "Failed build should return False"
        assert "docs directory not found" in status2["error"]