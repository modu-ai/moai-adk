# AST-grep models tests
"""Tests for AST-grep data models.

Following TDD RED-GREEN-REFACTOR cycle.
These tests define the expected behavior of AST-grep data models.
"""

from __future__ import annotations

from moai_adk.lsp.models import Position, Range


class TestASTMatch:
    """Tests for ASTMatch dataclass."""

    def test_astmatch_import(self) -> None:
        """Test that ASTMatch can be imported."""
        from moai_adk.astgrep.models import ASTMatch

        assert ASTMatch is not None

    def test_astmatch_creation_with_all_fields(self) -> None:
        """Test ASTMatch creation with all required fields."""
        from moai_adk.astgrep.models import ASTMatch

        match_range = Range(
            start=Position(line=10, character=5),
            end=Position(line=10, character=20),
        )
        match = ASTMatch(
            rule_id="no-console",
            severity="warning",
            message="Avoid using console.log",
            file_path="/path/to/file.js",
            range=match_range,
            suggested_fix="logger.info($MSG)",
        )

        assert match.rule_id == "no-console"
        assert match.severity == "warning"
        assert match.message == "Avoid using console.log"
        assert match.file_path == "/path/to/file.js"
        assert match.range == match_range
        assert match.suggested_fix == "logger.info($MSG)"

    def test_astmatch_creation_without_suggested_fix(self) -> None:
        """Test ASTMatch creation with optional suggested_fix as None."""
        from moai_adk.astgrep.models import ASTMatch

        match_range = Range(
            start=Position(line=5, character=0),
            end=Position(line=5, character=15),
        )
        match = ASTMatch(
            rule_id="sql-injection",
            severity="error",
            message="Potential SQL injection",
            file_path="/path/to/query.py",
            range=match_range,
            suggested_fix=None,
        )

        assert match.rule_id == "sql-injection"
        assert match.severity == "error"
        assert match.suggested_fix is None

    def test_astmatch_severity_values(self) -> None:
        """Test that ASTMatch accepts all valid severity values."""
        from moai_adk.astgrep.models import ASTMatch

        valid_severities = ["error", "warning", "info", "hint"]
        match_range = Range(
            start=Position(line=0, character=0),
            end=Position(line=0, character=10),
        )

        for severity in valid_severities:
            match = ASTMatch(
                rule_id="test-rule",
                severity=severity,
                message="Test message",
                file_path="/test.py",
                range=match_range,
                suggested_fix=None,
            )
            assert match.severity == severity


class TestScanConfig:
    """Tests for ScanConfig dataclass."""

    def test_scanconfig_import(self) -> None:
        """Test that ScanConfig can be imported."""
        from moai_adk.astgrep.models import ScanConfig

        assert ScanConfig is not None

    def test_scanconfig_default_values(self) -> None:
        """Test ScanConfig default values."""
        from moai_adk.astgrep.models import ScanConfig

        config = ScanConfig()

        assert config.rules_path is None
        assert config.security_scan is True
        assert config.include_patterns == []
        assert "node_modules" in config.exclude_patterns
        assert ".git" in config.exclude_patterns
        assert "__pycache__" in config.exclude_patterns

    def test_scanconfig_custom_values(self) -> None:
        """Test ScanConfig with custom values."""
        from moai_adk.astgrep.models import ScanConfig

        config = ScanConfig(
            rules_path="/custom/rules.yml",
            security_scan=False,
            include_patterns=["*.py", "*.js"],
            exclude_patterns=["vendor", "dist"],
        )

        assert config.rules_path == "/custom/rules.yml"
        assert config.security_scan is False
        assert config.include_patterns == ["*.py", "*.js"]
        assert config.exclude_patterns == ["vendor", "dist"]

    def test_scanconfig_mutable_defaults_are_isolated(self) -> None:
        """Test that mutable default values are properly isolated."""
        from moai_adk.astgrep.models import ScanConfig

        config1 = ScanConfig()
        config2 = ScanConfig()

        # Modify config1's lists
        config1.include_patterns.append("*.txt")
        config1.exclude_patterns.append("custom_exclude")

        # config2 should be unaffected
        assert "*.txt" not in config2.include_patterns
        assert "custom_exclude" not in config2.exclude_patterns


class TestScanResult:
    """Tests for ScanResult dataclass."""

    def test_scanresult_import(self) -> None:
        """Test that ScanResult can be imported."""
        from moai_adk.astgrep.models import ScanResult

        assert ScanResult is not None

    def test_scanresult_creation(self) -> None:
        """Test ScanResult creation with matches."""
        from moai_adk.astgrep.models import ASTMatch, ScanResult

        match_range = Range(
            start=Position(line=10, character=0),
            end=Position(line=10, character=25),
        )
        matches = [
            ASTMatch(
                rule_id="test-rule",
                severity="warning",
                message="Test message",
                file_path="/test.py",
                range=match_range,
                suggested_fix=None,
            )
        ]

        result = ScanResult(
            file_path="/test.py",
            matches=matches,
            scan_time_ms=15.5,
            language="python",
        )

        assert result.file_path == "/test.py"
        assert len(result.matches) == 1
        assert result.scan_time_ms == 15.5
        assert result.language == "python"

    def test_scanresult_empty_matches(self) -> None:
        """Test ScanResult with no matches."""
        from moai_adk.astgrep.models import ScanResult

        result = ScanResult(
            file_path="/clean.py",
            matches=[],
            scan_time_ms=5.0,
            language="python",
        )

        assert result.file_path == "/clean.py"
        assert result.matches == []
        assert result.scan_time_ms == 5.0


class TestProjectScanResult:
    """Tests for ProjectScanResult dataclass."""

    def test_projectscanresult_import(self) -> None:
        """Test that ProjectScanResult can be imported."""
        from moai_adk.astgrep.models import ProjectScanResult

        assert ProjectScanResult is not None

    def test_projectscanresult_creation(self) -> None:
        """Test ProjectScanResult creation."""
        from moai_adk.astgrep.models import ASTMatch, ProjectScanResult, ScanResult

        match_range = Range(
            start=Position(line=5, character=0),
            end=Position(line=5, character=20),
        )
        file_result = ScanResult(
            file_path="/project/app.py",
            matches=[
                ASTMatch(
                    rule_id="test-rule",
                    severity="error",
                    message="Error found",
                    file_path="/project/app.py",
                    range=match_range,
                    suggested_fix=None,
                )
            ],
            scan_time_ms=10.0,
            language="python",
        )

        result = ProjectScanResult(
            project_path="/project",
            files_scanned=5,
            total_matches=1,
            results_by_file={"/project/app.py": file_result},
            summary_by_severity={"error": 1, "warning": 0, "info": 0, "hint": 0},
            scan_time_ms=50.0,
        )

        assert result.project_path == "/project"
        assert result.files_scanned == 5
        assert result.total_matches == 1
        assert len(result.results_by_file) == 1
        assert result.summary_by_severity["error"] == 1
        assert result.scan_time_ms == 50.0

    def test_projectscanresult_empty(self) -> None:
        """Test ProjectScanResult with no matches."""
        from moai_adk.astgrep.models import ProjectScanResult

        result = ProjectScanResult(
            project_path="/clean_project",
            files_scanned=10,
            total_matches=0,
            results_by_file={},
            summary_by_severity={"error": 0, "warning": 0, "info": 0, "hint": 0},
            scan_time_ms=100.0,
        )

        assert result.files_scanned == 10
        assert result.total_matches == 0
        assert len(result.results_by_file) == 0


class TestReplaceResult:
    """Tests for ReplaceResult dataclass."""

    def test_replaceresult_import(self) -> None:
        """Test that ReplaceResult can be imported."""
        from moai_adk.astgrep.models import ReplaceResult

        assert ReplaceResult is not None

    def test_replaceresult_dry_run(self) -> None:
        """Test ReplaceResult for dry run operation."""
        from moai_adk.astgrep.models import ReplaceResult

        changes = [
            {
                "file_path": "/app.js",
                "old_code": "console.log(msg)",
                "new_code": "logger.info(msg)",
                "range": {"start": {"line": 10, "character": 0}, "end": {"line": 10, "character": 16}},
            }
        ]

        result = ReplaceResult(
            pattern="console.log($MSG)",
            replacement="logger.info($MSG)",
            matches_found=3,
            files_modified=2,
            changes=changes,
            dry_run=True,
        )

        assert result.pattern == "console.log($MSG)"
        assert result.replacement == "logger.info($MSG)"
        assert result.matches_found == 3
        assert result.files_modified == 2
        assert len(result.changes) == 1
        assert result.dry_run is True

    def test_replaceresult_actual_replacement(self) -> None:
        """Test ReplaceResult for actual replacement."""
        from moai_adk.astgrep.models import ReplaceResult

        result = ReplaceResult(
            pattern="var $NAME = $VALUE",
            replacement="const $NAME = $VALUE",
            matches_found=10,
            files_modified=5,
            changes=[],  # Simplified for test
            dry_run=False,
        )

        assert result.matches_found == 10
        assert result.files_modified == 5
        assert result.dry_run is False

    def test_replaceresult_no_matches(self) -> None:
        """Test ReplaceResult when no matches found."""
        from moai_adk.astgrep.models import ReplaceResult

        result = ReplaceResult(
            pattern="nonexistent_pattern",
            replacement="new_pattern",
            matches_found=0,
            files_modified=0,
            changes=[],
            dry_run=True,
        )

        assert result.matches_found == 0
        assert result.files_modified == 0
        assert result.changes == []


class TestModelExports:
    """Tests for module exports."""

    def test_all_models_exported_from_package(self) -> None:
        """Test that all models are exported from the package."""
        from moai_adk.astgrep.models import (
            ASTMatch,
            ProjectScanResult,
            ReplaceResult,
            ScanConfig,
            ScanResult,
        )

        assert ASTMatch is not None
        assert ScanConfig is not None
        assert ScanResult is not None
        assert ProjectScanResult is not None
        assert ReplaceResult is not None

    def test_models_can_be_imported_from_astgrep(self) -> None:
        """Test that models can be imported from astgrep package."""
        from moai_adk import astgrep

        assert hasattr(astgrep, "ASTMatch")
        assert hasattr(astgrep, "ScanConfig")
        assert hasattr(astgrep, "ScanResult")
        assert hasattr(astgrep, "ProjectScanResult")
        assert hasattr(astgrep, "ReplaceResult")
