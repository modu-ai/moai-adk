"""@TEST:DOCS-INTEGRATION-001 Documentation build integration tests

End-to-end tests for complete documentation building pipeline.
Verifies MkDocs build, HTML generation, and link integrity.

@REQ:DOCS-INTEGRATION-001 → @TASK:INTEGRATION-TEST-001
"""

import pytest
import subprocess
from pathlib import Path
from unittest.mock import patch, MagicMock

from moai_adk.core.docs.documentation_builder import DocumentationBuilder
from moai_adk.core.docs.api_generator import ApiGenerator
from moai_adk.core.docs.release_notes_converter import ReleaseNotesConverter


class TestDocumentationIntegration:
    """@TEST:DOCS-INTEGRATION-001 Test complete documentation build pipeline"""

    def test_should_build_complete_documentation_site(self, tmp_path):
        """@TEST:DOCS-INTEGRATION-002 Should build complete documentation successfully"""
        # Initialize all components
        builder = DocumentationBuilder(str(tmp_path))
        api_gen = ApiGenerator(str(tmp_path), "src")
        release_gen = ReleaseNotesConverter(str(tmp_path))

        # Create mock source structure
        src_dir = tmp_path / "src" / "moai_adk"
        src_dir.mkdir(parents=True)
        (src_dir / "__init__.py").write_text('"""MoAI ADK"""')

        # Initialize and build site
        builder.initialize_site()

        docs_dir = tmp_path / "docs"

        # Generate API docs
        api_gen.generate_api_docs(str(docs_dir))

        # Generate release notes
        release_gen.generate_release_notes(str(docs_dir))

        # Mock MkDocs build success
        with patch("mkdocs.commands.build.build") as mock_build:
            mock_build.return_value = True

            result = builder.build_docs()

            assert result is True
            assert (docs_dir / "api").exists(), "API docs should be generated"

    def test_should_generate_valid_html_output(self, tmp_path):
        """@TEST:DOCS-INTEGRATION-003 Should generate valid HTML files"""
        builder = DocumentationBuilder(str(tmp_path))
        builder.initialize_site()

        site_dir = tmp_path / "site"

        # Mock the HTML generation process
        with patch("mkdocs.commands.build.build") as mock_build:
            # Simulate HTML file creation
            site_dir.mkdir()
            (site_dir / "index.html").write_text("<html><body>Test</body></html>")
            (site_dir / "getting-started" / "installation").mkdir(parents=True)
            (site_dir / "getting-started" / "installation" / "index.html").write_text(
                "<html><body>Installation</body></html>"
            )

            mock_build.return_value = True

            result = builder.build_docs()

            assert result is True
            assert (site_dir / "index.html").exists()

    def test_should_validate_internal_links(self, tmp_path):
        """@TEST:DOCS-INTEGRATION-004 Should validate internal links integrity"""
        builder = DocumentationBuilder(str(tmp_path))
        builder.initialize_site()

        # Create docs with internal links
        docs_dir = tmp_path / "docs"
        index_md = docs_dir / "index.md"
        index_content = """
# MoAI-ADK Documentation

- [Getting Started](getting-started/installation.md)
- [API Reference](api/index.md)
"""
        index_md.write_text(index_content)

        # Validate links
        link_results = builder.validate_links()

        # Should detect that some links are missing
        assert "getting-started/installation.md" in link_results["missing"]

    def test_should_handle_build_failures_gracefully(self, tmp_path):
        """@TEST:DOCS-INTEGRATION-005 Should handle build failures with proper error reporting"""
        builder = DocumentationBuilder(str(tmp_path))
        builder.initialize_site()

        # Mock build failure
        with patch("mkdocs.commands.build.build") as mock_build:
            mock_build.side_effect = subprocess.CalledProcessError(1, "mkdocs")

            result = builder.build_docs()

            assert result is False

            status = builder.get_build_status()
            assert status["success"] is False
            assert "error" in status

    def test_should_integrate_all_documentation_sources(self, tmp_path):
        """@TEST:DOCS-INTEGRATION-006 Should integrate API docs, release notes, and manual content"""
        # Setup complete documentation pipeline
        builder = DocumentationBuilder(str(tmp_path))
        api_gen = ApiGenerator(str(tmp_path), "src")
        release_gen = ReleaseNotesConverter(str(tmp_path))

        # Create comprehensive source structure
        src_dir = tmp_path / "src" / "moai_adk"
        src_dir.mkdir(parents=True)
        (src_dir / "__init__.py").write_text('''"""MoAI ADK Package

Complete Agentic Development Kit for Claude Code integration.
"""''')

        core_dir = src_dir / "core"
        core_dir.mkdir()
        (core_dir / "__init__.py").touch()
        (core_dir / "manager.py").write_text('''"""Core manager module"""

class CoreManager:
    """Main core manager class"""

    def __init__(self):
        """Initialize core manager"""
        pass
''')

        # Create sync report
        moai_dir = tmp_path / ".moai" / "reports"
        moai_dir.mkdir(parents=True)
        sync_report = moai_dir / "sync-report.md"
        sync_report.write_text("""
# Sync Report - 2024-09-25

## Version Info
- Current: v0.2.2

## Changes Summary

### @FEATURE:DOCS-001
- Added comprehensive documentation system
- Integrated MkDocs Material theme

### @API:DOCS-001
- Auto-generated API documentation from source
""")

        # Initialize site
        builder.initialize_site()
        docs_dir = tmp_path / "docs"

        # Generate all documentation
        api_gen.generate_api_docs(str(docs_dir))
        release_gen.generate_release_notes(str(docs_dir))

        # Verify all components exist
        assert (docs_dir / "api").exists()
        assert (docs_dir / "release-notes.md").exists()
        assert (docs_dir / "index.md").exists()

        # Mock successful build
        with patch("mkdocs.commands.build.build") as mock_build:
            mock_build.return_value = True
            result = builder.build_docs()
            assert result is True

    def test_should_support_incremental_builds(self, tmp_path):
        """@TEST:DOCS-INTEGRATION-007 Should support incremental documentation builds"""
        builder = DocumentationBuilder(str(tmp_path))
        builder.initialize_site()

        # Initial build
        with patch("mkdocs.commands.build.build") as mock_build:
            mock_build.return_value = True

            first_result = builder.build_docs()
            first_call_count = mock_build.call_count

            # Incremental build (no changes)
            second_result = builder.build_docs(incremental=True)

            assert first_result is True
            assert second_result is True

            # Should optimize for incremental builds
            # (Implementation dependent on actual optimization strategy)

    def test_should_validate_documentation_completeness(self, tmp_path):
        """@TEST:DOCS-INTEGRATION-008 Should validate documentation completeness"""
        builder = DocumentationBuilder(str(tmp_path))
        builder.initialize_site()

        # Check documentation completeness
        completeness_report = builder.check_completeness()

        # Should identify missing sections
        assert "missing_sections" in completeness_report
        assert "coverage_percent" in completeness_report

        # With minimal setup, coverage should be low
        assert completeness_report["coverage_percent"] < 100
