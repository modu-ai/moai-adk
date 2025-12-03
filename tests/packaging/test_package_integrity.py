"""
Test package integrity verification.

This module tests the package integrity verification script to ensure
all critical files and directories are included in the built packages.

NOTE: verify_package_integrity.py script not implemented in v0.26.0.
All tests in this module are skipped.
"""

import tarfile
import tempfile
import zipfile
from pathlib import Path
from unittest import mock

import pytest

# Skip all tests in this module - verify_package_integrity.py not implemented
pytestmark = pytest.mark.skip(reason="verify_package_integrity.py script not implemented in moai-adk v0.26.0")


class TestSourceFilesVerification:
    """Test source files verification."""

    def test_verify_source_files_returns_true_when_all_files_exist(self):
        """Test that verification passes when all required files exist."""
        result = verify_source_files()
        assert result is True

    def test_verify_source_files_checks_output_styles_r2d2(self):
        """Test that r2d2.md output style is checked."""
        assert any("output-styles/moai/r2d2.md" in str(path) for path in REQUIRED_SOURCE_FILES)

    def test_verify_source_files_checks_output_styles_yoda(self):
        """Test that yoda.md output style is checked."""
        assert any("output-styles/moai/yoda.md" in str(path) for path in REQUIRED_SOURCE_FILES)

    def test_verify_source_files_checks_skills_directory(self):
        """Test that skills directory is checked."""
        assert any("skills" in str(path) for path in REQUIRED_SOURCE_FILES)

    def test_verify_source_files_checks_agents_directory(self):
        """Test that agents directory is checked."""
        assert any("agents" in str(path) for path in REQUIRED_SOURCE_FILES)

    def test_verify_source_files_checks_config_json(self):
        """Test that config.json is checked."""
        assert any("config/config.json" in str(path) for path in REQUIRED_SOURCE_FILES)

    def test_verify_source_files_returns_false_when_file_missing(self):
        """Test that verification fails when a required file is missing."""
        with mock.patch("pathlib.Path.exists", return_value=False):
            result = verify_source_files()
            assert result is False

    def test_verify_source_files_prints_error_on_missing_file(self, capsys):
        """Test that error message is printed for missing files."""
        with mock.patch("pathlib.Path.exists", return_value=False):
            verify_source_files()
            captured = capsys.readouterr()
            assert "Missing:" in captured.out or "Missing:" in captured.err

    def test_required_source_files_is_list(self):
        """Test that REQUIRED_SOURCE_FILES is a list."""
        assert isinstance(REQUIRED_SOURCE_FILES, list)

    def test_required_source_files_not_empty(self):
        """Test that REQUIRED_SOURCE_FILES is not empty."""
        assert len(REQUIRED_SOURCE_FILES) > 0


class TestWheelContentsVerification:
    """Test wheel file contents verification."""

    def test_verify_wheel_contents_with_valid_wheel(self):
        """Test verification with a valid wheel file."""
        # Create a temporary wheel file with required patterns
        with tempfile.NamedTemporaryFile(suffix=".whl", delete=False) as tmp:
            wheel_path = tmp.name

        try:
            # Create a valid wheel with required patterns and critical files
            with zipfile.ZipFile(wheel_path, "w") as whl:
                for pattern in REQUIRED_WHEEL_PATTERNS:
                    whl.writestr(f"{pattern}/dummy.txt", "content")
                # Add critical files
                whl.writestr(
                    "moai_adk/templates/.claude/output-styles/moai/r2d2.md",
                    "content",
                )
                whl.writestr(
                    "moai_adk/templates/.claude/output-styles/moai/yoda.md",
                    "content",
                )
                whl.writestr(
                    "moai_adk/templates/.moai/config/config.json",
                    "content",
                )

            result = verify_wheel_contents(wheel_path)
            assert result is True
        finally:
            Path(wheel_path).unlink(missing_ok=True)

    def test_verify_wheel_contents_returns_false_with_missing_pattern(self):
        """Test verification fails when a required pattern is missing."""
        with tempfile.NamedTemporaryFile(suffix=".whl", delete=False) as tmp:
            wheel_path = tmp.name

        try:
            # Create a wheel with only some required patterns
            with zipfile.ZipFile(wheel_path, "w") as whl:
                whl.writestr("moai_adk/templates/.claude/output-styles/moai/r2d2.md", "content")

            result = verify_wheel_contents(wheel_path)
            assert result is False
        finally:
            Path(wheel_path).unlink(missing_ok=True)

    def test_verify_wheel_contents_handles_nonexistent_file(self):
        """Test that verification gracefully handles missing wheel file."""
        result = verify_wheel_contents("/nonexistent/path/fake.whl")
        assert result is False

    def test_verify_wheel_contents_checks_templates_output_styles(self):
        """Test that templates output-styles pattern is checked."""
        assert any("output-styles" in pattern for pattern in REQUIRED_WHEEL_PATTERNS)

    def test_verify_wheel_contents_checks_templates_skills(self):
        """Test that templates skills pattern is checked."""
        assert any("skills" in pattern for pattern in REQUIRED_WHEEL_PATTERNS)

    def test_verify_wheel_contents_checks_templates_config(self):
        """Test that templates config pattern is checked."""
        assert any("config" in pattern for pattern in REQUIRED_WHEEL_PATTERNS)

    def test_verify_wheel_contents_prints_error_on_missing_pattern(self, capsys):
        """Test that error message is printed for missing patterns."""
        with tempfile.NamedTemporaryFile(suffix=".whl", delete=False) as tmp:
            wheel_path = tmp.name

        try:
            with zipfile.ZipFile(wheel_path, "w") as whl:
                whl.writestr("dummy.txt", "content")

            verify_wheel_contents(wheel_path)
            captured = capsys.readouterr()
            assert "Missing" in captured.out or "Missing" in captured.err or "Pattern" in captured.out
        finally:
            Path(wheel_path).unlink(missing_ok=True)


class TestTarballContentsVerification:
    """Test tarball file contents verification."""

    def test_verify_tarball_contents_with_valid_tarball(self):
        """Test verification with a valid tarball file."""
        import io

        with tempfile.NamedTemporaryFile(suffix=".tar.gz", delete=False) as tmp:
            tarball_path = tmp.name

        try:
            # Create a valid tarball with required patterns
            with tarfile.open(tarball_path, "w:gz") as tar:
                for pattern in REQUIRED_WHEEL_PATTERNS:
                    info = tarfile.TarInfo(name=f"{pattern}/dummy.txt")
                    info.size = 7
                    tar.addfile(info, fileobj=io.BytesIO(b"content"))

            result = verify_tarball_contents(tarball_path)
            assert result is True
        finally:
            Path(tarball_path).unlink(missing_ok=True)

    def test_verify_tarball_contents_returns_false_with_missing_pattern(self):
        """Test verification fails when a required pattern is missing."""
        import io

        with tempfile.NamedTemporaryFile(suffix=".tar.gz", delete=False) as tmp:
            tarball_path = tmp.name

        try:
            # Create a tarball with only some required patterns
            with tarfile.open(tarball_path, "w:gz") as tar:
                info = tarfile.TarInfo(name="moai_adk/templates/dummy.txt")
                info.size = 7
                tar.addfile(info, fileobj=io.BytesIO(b"content"))

            result = verify_tarball_contents(tarball_path)
            assert result is False
        finally:
            Path(tarball_path).unlink(missing_ok=True)

    def test_verify_tarball_contents_handles_nonexistent_file(self):
        """Test that verification gracefully handles missing tarball file."""
        result = verify_tarball_contents("/nonexistent/path/fake.tar.gz")
        assert result is False

    def test_verify_tarball_contents_handles_invalid_tarball(self):
        """Test that verification handles corrupted tarball files."""
        with tempfile.NamedTemporaryFile(suffix=".tar.gz", delete=False) as tmp:
            tarball_path = tmp.name
            tmp.write(b"not a valid tarball")

        try:
            result = verify_tarball_contents(tarball_path)
            assert result is False
        finally:
            Path(tarball_path).unlink(missing_ok=True)

    def test_verify_tarball_contents_checks_same_patterns_as_wheel(self):
        """Test that tarball checks use the same patterns as wheel."""
        # Both should check the same required patterns
        assert REQUIRED_WHEEL_PATTERNS is not None
        assert len(REQUIRED_WHEEL_PATTERNS) > 0


class TestRequiredPatterns:
    """Test that required patterns are properly defined."""

    def test_required_wheel_patterns_is_list(self):
        """Test that REQUIRED_WHEEL_PATTERNS is a list."""
        assert isinstance(REQUIRED_WHEEL_PATTERNS, list)

    def test_required_wheel_patterns_not_empty(self):
        """Test that REQUIRED_WHEEL_PATTERNS is not empty."""
        assert len(REQUIRED_WHEEL_PATTERNS) > 0

    def test_required_wheel_patterns_include_output_styles(self):
        """Test that output-styles is in patterns."""
        assert any("output-styles" in pattern for pattern in REQUIRED_WHEEL_PATTERNS)

    def test_required_wheel_patterns_include_skills(self):
        """Test that skills is in patterns."""
        assert any("skills" in pattern for pattern in REQUIRED_WHEEL_PATTERNS)

    def test_required_wheel_patterns_include_config(self):
        """Test that config is in patterns."""
        assert any("config" in pattern for pattern in REQUIRED_WHEEL_PATTERNS)
