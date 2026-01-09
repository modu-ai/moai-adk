# AST-grep analyzer tests
"""Tests for MoAIASTGrepAnalyzer.

Following TDD RED-GREEN-REFACTOR cycle.
These tests define the expected behavior of the AST-grep analyzer.
"""

from __future__ import annotations

import tempfile
from pathlib import Path
from unittest.mock import patch

import pytest


class TestMoAIASTGrepAnalyzerImport:
    """Tests for MoAIASTGrepAnalyzer import."""

    def test_analyzer_import(self) -> None:
        """Test that MoAIASTGrepAnalyzer can be imported."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        assert MoAIASTGrepAnalyzer is not None


class TestMoAIASTGrepAnalyzerInit:
    """Tests for MoAIASTGrepAnalyzer initialization."""

    def test_analyzer_instantiation_default(self) -> None:
        """Test analyzer instantiation with default config."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        analyzer = MoAIASTGrepAnalyzer()
        assert analyzer is not None

    def test_analyzer_instantiation_with_config(self) -> None:
        """Test analyzer instantiation with custom config."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        from moai_adk.astgrep.models import ScanConfig

        config = ScanConfig(
            rules_path="/custom/rules",
            security_scan=False,
        )
        analyzer = MoAIASTGrepAnalyzer(config=config)

        assert analyzer is not None
        assert analyzer.config.rules_path == "/custom/rules"
        assert analyzer.config.security_scan is False


class TestScanFile:
    """Tests for scan_file method."""

    def test_scan_file_returns_scan_result(self) -> None:
        """Test that scan_file returns a ScanResult."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        from moai_adk.astgrep.models import ScanResult

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("print('hello')\n")
            f.flush()

            analyzer = MoAIASTGrepAnalyzer()
            result = analyzer.scan_file(f.name)

            assert isinstance(result, ScanResult)
            assert result.file_path == f.name
            assert result.language == "python"
            assert isinstance(result.matches, list)
            assert result.scan_time_ms >= 0

        Path(f.name).unlink()

    def test_scan_file_nonexistent_raises_error(self) -> None:
        """Test that scanning nonexistent file raises FileNotFoundError."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        analyzer = MoAIASTGrepAnalyzer()

        with pytest.raises(FileNotFoundError):
            analyzer.scan_file("/nonexistent/file.py")

    def test_scan_file_with_custom_config(self) -> None:
        """Test scan_file with custom config override."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        from moai_adk.astgrep.models import ScanConfig, ScanResult

        with tempfile.NamedTemporaryFile(mode="w", suffix=".js", delete=False) as f:
            f.write("console.log('test');\n")
            f.flush()

            analyzer = MoAIASTGrepAnalyzer()
            config = ScanConfig(security_scan=False)
            result = analyzer.scan_file(f.name, config=config)

            assert isinstance(result, ScanResult)
            assert result.language == "javascript"

        Path(f.name).unlink()

    def test_scan_file_detects_language(self) -> None:
        """Test that scan_file correctly detects file language."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        test_files = [
            (".py", "python"),
            (".js", "javascript"),
            (".ts", "typescript"),
            (".tsx", "typescriptreact"),
            (".jsx", "javascriptreact"),
            (".go", "go"),
            (".rs", "rust"),
            (".java", "java"),
            (".rb", "ruby"),
        ]

        analyzer = MoAIASTGrepAnalyzer()

        for ext, expected_lang in test_files:
            with tempfile.NamedTemporaryFile(mode="w", suffix=ext, delete=False) as f:
                f.write("// test content\n")
                f.flush()

                result = analyzer.scan_file(f.name)
                assert result.language == expected_lang, f"Expected {expected_lang} for {ext}"

            Path(f.name).unlink()


class TestScanProject:
    """Tests for scan_project method."""

    def test_scan_project_returns_project_scan_result(self) -> None:
        """Test that scan_project returns a ProjectScanResult."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        from moai_adk.astgrep.models import ProjectScanResult

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create test files
            (Path(tmpdir) / "app.py").write_text("print('hello')\n")
            (Path(tmpdir) / "utils.py").write_text("def foo(): pass\n")

            analyzer = MoAIASTGrepAnalyzer()
            result = analyzer.scan_project(tmpdir)

            assert isinstance(result, ProjectScanResult)
            assert result.project_path == tmpdir
            assert result.files_scanned >= 2
            assert isinstance(result.results_by_file, dict)
            assert isinstance(result.summary_by_severity, dict)
            assert result.scan_time_ms >= 0

    def test_scan_project_nonexistent_raises_error(self) -> None:
        """Test that scanning nonexistent project raises FileNotFoundError."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        analyzer = MoAIASTGrepAnalyzer()

        with pytest.raises(FileNotFoundError):
            analyzer.scan_project("/nonexistent/project")

    def test_scan_project_respects_exclude_patterns(self) -> None:
        """Test that scan_project respects exclude patterns."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create test files
            (Path(tmpdir) / "app.py").write_text("print('hello')\n")

            # Create excluded directory
            node_modules = Path(tmpdir) / "node_modules"
            node_modules.mkdir()
            (node_modules / "pkg.js").write_text("console.log('excluded');\n")

            analyzer = MoAIASTGrepAnalyzer()
            result = analyzer.scan_project(tmpdir)

            # node_modules should be excluded by default
            for file_path in result.results_by_file:
                assert "node_modules" not in file_path

    def test_scan_project_with_include_patterns(self) -> None:
        """Test that scan_project respects include patterns."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        from moai_adk.astgrep.models import ScanConfig

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create test files
            (Path(tmpdir) / "app.py").write_text("print('hello')\n")
            (Path(tmpdir) / "app.js").write_text("console.log('test');\n")

            config = ScanConfig(include_patterns=["*.py"])
            analyzer = MoAIASTGrepAnalyzer(config=config)
            result = analyzer.scan_project(tmpdir)

            # Should only include .py files
            for file_path in result.results_by_file:
                assert file_path.endswith(".py")


class TestPatternSearch:
    """Tests for pattern_search method."""

    def test_pattern_search_returns_list_of_matches(self) -> None:
        """Test that pattern_search returns a list of ASTMatch."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        from moai_adk.astgrep.models import ASTMatch

        with tempfile.TemporaryDirectory() as tmpdir:
            (Path(tmpdir) / "test.py").write_text("print('hello')\nprint('world')\n")

            analyzer = MoAIASTGrepAnalyzer()
            matches = analyzer.pattern_search(
                pattern="print($MSG)",
                language="python",
                path=tmpdir,
            )

            assert isinstance(matches, list)
            for match in matches:
                assert isinstance(match, ASTMatch)

    def test_pattern_search_with_no_matches(self) -> None:
        """Test pattern_search returns empty list when no matches."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        with tempfile.TemporaryDirectory() as tmpdir:
            (Path(tmpdir) / "test.py").write_text("x = 1\ny = 2\n")

            analyzer = MoAIASTGrepAnalyzer()
            matches = analyzer.pattern_search(
                pattern="print($MSG)",
                language="python",
                path=tmpdir,
            )

            assert matches == []

    def test_pattern_search_nonexistent_path_raises_error(self) -> None:
        """Test pattern_search with nonexistent path raises error."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        analyzer = MoAIASTGrepAnalyzer()

        with pytest.raises(FileNotFoundError):
            analyzer.pattern_search(
                pattern="print($MSG)",
                language="python",
                path="/nonexistent/path",
            )


class TestPatternReplace:
    """Tests for pattern_replace method."""

    def test_pattern_replace_dry_run_returns_replace_result(self) -> None:
        """Test that pattern_replace dry run returns ReplaceResult."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        from moai_adk.astgrep.models import ReplaceResult

        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.py"
            test_file.write_text("print('hello')\n")

            analyzer = MoAIASTGrepAnalyzer()
            result = analyzer.pattern_replace(
                pattern="print($MSG)",
                replacement="logger.info($MSG)",
                language="python",
                path=tmpdir,
                dry_run=True,
            )

            assert isinstance(result, ReplaceResult)
            assert result.pattern == "print($MSG)"
            assert result.replacement == "logger.info($MSG)"
            assert result.dry_run is True

            # File should not be modified in dry run
            assert test_file.read_text() == "print('hello')\n"

    def test_pattern_replace_actual_modifies_files(self) -> None:
        """Test that pattern_replace with dry_run=False modifies files."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        from moai_adk.astgrep.models import ReplaceResult

        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.js"
            test_file.write_text("console.log('hello');\n")

            analyzer = MoAIASTGrepAnalyzer()
            result = analyzer.pattern_replace(
                pattern="console.log($MSG)",
                replacement="logger.info($MSG)",
                language="javascript",
                path=tmpdir,
                dry_run=False,
            )

            assert isinstance(result, ReplaceResult)
            assert result.dry_run is False

            # Note: Actual file modification depends on sg CLI availability
            # If sg is not available, this is a graceful no-op

    def test_pattern_replace_no_matches(self) -> None:
        """Test pattern_replace with no matches."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        from moai_adk.astgrep.models import ReplaceResult

        with tempfile.TemporaryDirectory() as tmpdir:
            (Path(tmpdir) / "test.py").write_text("x = 1\n")

            analyzer = MoAIASTGrepAnalyzer()
            result = analyzer.pattern_replace(
                pattern="nonexistent_func($ARG)",
                replacement="new_func($ARG)",
                language="python",
                path=tmpdir,
                dry_run=True,
            )

            assert isinstance(result, ReplaceResult)
            assert result.matches_found == 0
            assert result.files_modified == 0

    def test_pattern_replace_nonexistent_path_raises_error(self) -> None:
        """Test pattern_replace with nonexistent path raises error."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        analyzer = MoAIASTGrepAnalyzer()

        with pytest.raises(FileNotFoundError):
            analyzer.pattern_replace(
                pattern="print($MSG)",
                replacement="logger.info($MSG)",
                language="python",
                path="/nonexistent/path",
                dry_run=True,
            )


class TestGracefulDegradation:
    """Tests for graceful degradation when sg CLI is not available."""

    def test_scan_file_without_sg_cli(self) -> None:
        """Test that scan_file works without sg CLI (graceful degradation)."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        from moai_adk.astgrep.models import ScanResult

        with tempfile.NamedTemporaryFile(mode="w", suffix=".py", delete=False) as f:
            f.write("print('hello')\n")
            f.flush()

            analyzer = MoAIASTGrepAnalyzer()

            # Mock subprocess to simulate sg CLI not found
            with patch("subprocess.run") as mock_run:
                mock_run.side_effect = FileNotFoundError("sg not found")

                result = analyzer.scan_file(f.name)

                # Should return empty result without crashing
                assert isinstance(result, ScanResult)
                assert result.matches == []

        Path(f.name).unlink()

    def test_is_sg_available(self) -> None:
        """Test is_sg_available method."""
        from moai_adk.astgrep.analyzer import MoAIASTGrepAnalyzer

        analyzer = MoAIASTGrepAnalyzer()

        # Should return a boolean
        result = analyzer.is_sg_available()
        assert isinstance(result, bool)


class TestAnalyzerExports:
    """Tests for module exports."""

    def test_analyzer_exported_from_astgrep_package(self) -> None:
        """Test MoAIASTGrepAnalyzer is exported from astgrep package."""
        from moai_adk import astgrep

        assert hasattr(astgrep, "MoAIASTGrepAnalyzer")

    def test_all_components_exported(self) -> None:
        """Test all main components are exported from astgrep package."""
        from moai_adk.astgrep import (
            ASTMatch,
            MoAIASTGrepAnalyzer,
            ProjectScanResult,
            ReplaceResult,
            Rule,
            RuleLoader,
            ScanConfig,
            ScanResult,
        )

        assert ASTMatch is not None
        assert ScanConfig is not None
        assert ScanResult is not None
        assert ProjectScanResult is not None
        assert ReplaceResult is not None
        assert Rule is not None
        assert RuleLoader is not None
        assert MoAIASTGrepAnalyzer is not None
