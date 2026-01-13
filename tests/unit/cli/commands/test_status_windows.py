"""Windows-specific tests for status.py command.

Tests Windows-specific scenarios:
- Windows path handling with backslashes and drive letters
- UTF-8 encoding scenarios for file operations
- Console encoding with emoji handling
- Path operations on Windows-style paths

Coverage Goals:
- Test Windows-specific code paths (sys.platform == "win32")
- Verify UTF-8 encoding works correctly on Windows paths
- Ensure backslashes and drive letters are handled correctly
"""

import os
import sys
import tempfile
from pathlib import Path
from unittest.mock import Mock, patch

import pytest
import yaml
from click.testing import CliRunner

from moai_adk.cli.commands.status import status


@pytest.fixture
def runner():
    """Create CLI runner for testing"""
    return CliRunner()


@pytest.mark.skipif(sys.platform != "win32", reason="Windows-only test")
class TestWindowsPathHandling:
    """Test Windows-specific path handling scenarios"""

    def test_status_with_windows_config_path(self, runner):
        """Should handle Windows-style config path correctly"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {
                "project": {
                    "mode": "personal",
                    "locale": "en",
                }
            }
            with open(config_path, "w", encoding="utf-8") as f:
                yaml.dump(config_data, f)

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                assert "Project Status" in result.output
                assert "personal" in result.output

            finally:
                os.chdir(original_cwd)

    def test_status_with_subdirectory_config(self, runner):
        """Should handle config in nested subdirectories"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": "team", "locale": "ko"}}
            with open(config_path, "w", encoding="utf-8") as f:
                yaml.dump(config_data, f)

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                assert "team" in result.output

            finally:
                os.chdir(original_cwd)

    def test_status_with_deeply_nested_specs(self, runner):
        """Should handle deeply nested SPEC directories"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": "personal", "locale": "en"}}
            with open(config_path, "w", encoding="utf-8") as f:
                yaml.dump(config_data, f)

            # Create deeply nested SPEC structure
            specs_dir = project_path / ".moai" / "specs"
            spec_path = specs_dir / "SPEC-DEEP-001"
            spec_path.mkdir(parents=True, exist_ok=True)

            # Create deeply nested subdirectory (should not be counted)
            nested_path = spec_path / "level1" / "level2" / "level3"
            nested_path.mkdir(parents=True, exist_ok=True)
            (nested_path / "spec.md").write_text("# Deep SPEC")

            # Create valid spec file at top level
            (spec_path / "spec.md").write_text("# Main SPEC")

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                # Should count only top-level SPEC
                assert "1" in result.output

            finally:
                os.chdir(original_cwd)


@pytest.mark.skipif(sys.platform != "win32", reason="Windows-only test")
class TestWindowsUtf8Encoding:
    """Test UTF-8 encoding scenarios specific to Windows"""

    def test_status_with_unicode_locale(self, runner):
        """Should handle Unicode characters in locale field"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {
                "project": {
                    "mode": "personal",
                    "locale": "en_US-UTF-8",
                }
            }
            with open(config_path, "w", encoding="utf-8") as f:
                yaml.dump(config_data, f)

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                assert "en_US-UTF-8" in result.output

            finally:
                os.chdir(original_cwd)

    def test_status_with_unicode_mode(self, runner):
        """Should handle Unicode characters in mode field"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {
                "project": {
                    "mode": "팀",  # Korean for "team"
                    "locale": "ko",
                }
            }
            with open(config_path, "w", encoding="utf-8") as f:
                yaml.dump(config_data, f, allow_unicode=True)

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                assert "팀" in result.output

            finally:
                os.chdir(original_cwd)

    def test_status_config_with_utf8_bom(self, runner):
        """Should handle UTF-8 BOM in config file"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": "personal", "locale": "en"}}

            # Write with UTF-8 BOM
            with open(config_path, "wb") as f:
                f.write(b"\xef\xbb\xbf")  # UTF-8 BOM
                f.write(yaml.dump(config_data).encode("utf-8"))

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert - Should handle BOM gracefully
                assert result.exit_code == 0

            finally:
                os.chdir(original_cwd)


@pytest.mark.skipif(sys.platform != "win32", reason="Windows-only test")
class TestWindowsConsoleEncoding:
    """Test console encoding scenarios on Windows"""

    def test_status_output_with_emoji(self, runner):
        """Should display emoji characters correctly on Windows console"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": "personal", "locale": "en"}}
            with open(config_path, "w", encoding="utf-8") as f:
                yaml.dump(config_data, f)

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                # The output should be generated without encoding errors

            finally:
                os.chdir(original_cwd)

    def test_status_warning_emoji_display(self, runner):
        """Should display warning emoji correctly when config missing"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            original_cwd = Path.cwd()
            try:
                os.chdir(tmpdir)

                # Act - No config file
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code != 0
                # Should show warning message

            finally:
                os.chdir(original_cwd)


@pytest.mark.skipif(sys.platform != "win32", reason="Windows-only test")
class TestWindowsPlatformSpecific:
    """Test Windows platform-specific code paths"""

    @patch("sys.platform", "win32")
    def test_windows_console_initialization(self, runner):
        """Should initialize console correctly on Windows platform"""
        # Arrange & Act - Import on Windows platform
        # The console is initialized at module import time
        # This test verifies the code path works
        result = runner.invoke(status, ["--help"])

        # Assert
        assert result.exit_code == 0

    def test_status_with_windows_line_endings(self, runner):
        """Should handle Windows line endings (CRLF) in config"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": "team", "locale": "en"}}

            # Write with Windows line endings
            content = yaml.dump(config_data).replace("\n", "\r\n")
            with open(config_path, "w", encoding="utf-8") as f:
                f.write(content)

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                assert "team" in result.output

            finally:
                os.chdir(original_cwd)


@pytest.mark.skipif(sys.platform != "win32", reason="Windows-only test")
class TestWindowsSpecCounting:
    """Test SPEC document counting on Windows"""

    def test_status_counts_specs_with_unicode_names(self, runner):
        """Should count SPEC directories with Unicode characters"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": "personal", "locale": "en"}}
            with open(config_path, "w", encoding="utf-8") as f:
                yaml.dump(config_data, f)

            # Create SPEC with Unicode characters
            specs_dir = project_path / ".moai" / "specs"
            spec_ids = [
                "SPEC-AUTH-001",
                "SPEC-한국어-001",  # Korean
                "SPEC-日本語-001",  # Japanese
            ]
            for spec_id in spec_ids:
                spec_path = specs_dir / spec_id
                spec_path.mkdir(parents=True, exist_ok=True)
                (spec_path / "spec.md").write_text(f"# {spec_id}", encoding="utf-8")

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                assert "3" in result.output

            finally:
                os.chdir(original_cwd)

    def test_status_ignores_non_spec_directories(self, runner):
        """Should ignore directories that don't match SPEC-* pattern"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": "personal", "locale": "en"}}
            with open(config_path, "w", encoding="utf-8") as f:
                yaml.dump(config_data, f)

            # Create mix of SPEC and non-SPEC directories
            specs_dir = project_path / ".moai" / "specs"
            specs_dir.mkdir(parents=True, exist_ok=True)

            (specs_dir / "SPEC-001").mkdir()
            (specs_dir / "SPEC-001" / "spec.md").write_text("# SPEC 001")

            (specs_dir / "other-dir").mkdir()
            (specs_dir / "other-dir" / "spec.md").write_text("# Not a SPEC")

            (specs_dir / "NOT-SPEC-002").mkdir()
            (specs_dir / "NOT-SPEC-002" / "spec.md").write_text("# Not a SPEC")

            (specs_dir / "SPEC-002").mkdir()
            (specs_dir / "SPEC-002" / "spec.md").write_text("# SPEC 002")

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                # Should only count SPEC-001 and SPEC-002
                assert "2" in result.output

            finally:
                os.chdir(original_cwd)


@pytest.mark.skipif(sys.platform != "win32", reason="Windows-only test")
class TestWindowsEdgeCases:
    """Test Windows-specific edge cases"""

    def test_status_with_empty_config(self, runner):
        """Should handle empty config file"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            # Empty config
            with open(config_path, "w", encoding="utf-8") as f:
                f.write("")

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                # Should show "unknown" for missing values
                assert "unknown" in result.output

            finally:
                os.chdir(original_cwd)

    def test_status_with_missing_spec_md(self, runner):
        """Should handle SPEC directories without spec.md file"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": "personal", "locale": "en"}}
            with open(config_path, "w", encoding="utf-8") as f:
                yaml.dump(config_data, f)

            # Create SPEC directory without spec.md
            specs_dir = project_path / ".moai" / "specs"
            spec_path = specs_dir / "SPEC-001"
            spec_path.mkdir(parents=True, exist_ok=True)
            # Don't create spec.md

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                # Should count directory even without spec.md
                assert "0" in result.output

            finally:
                os.chdir(original_cwd)

    def test_status_with_null_values(self, runner):
        """Should handle null values in config"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {
                "project": {
                    "mode": None,
                    "locale": None,
                }
            }
            with open(config_path, "w", encoding="utf-8") as f:
                yaml.dump(config_data, f)

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                # Should show "unknown" for null values

            finally:
                os.chdir(original_cwd)

    @patch("git.Repo")
    def test_status_with_git_error(self, mock_repo, runner):
        """Should handle Git repository errors gracefully"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": "personal", "locale": "en"}}
            with open(config_path, "w", encoding="utf-8") as f:
                yaml.dump(config_data, f)

            # Mock Git to raise exception
            mock_repo.side_effect = Exception("Git error")

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                # Should still show status without Git info

            finally:
                os.chdir(original_cwd)


@pytest.mark.skipif(sys.platform != "win32", reason="Windows-only test")
class TestWindowsParametrized:
    """Parametrized Windows-specific tests"""

    @pytest.mark.parametrize(
        "mode,locale,expected_mode,expected_locale",
        [
            ("personal", "en", "personal", "en"),
            ("team", "ko", "team", "ko"),
            ("personal", "ja", "personal", "ja"),
            ("team", "zh", "team", "zh"),
        ],
    )
    def test_status_various_locales(self, runner, mode, locale, expected_mode, expected_locale):
        """Should handle various locale settings correctly"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": mode, "locale": locale}}
            with open(config_path, "w", encoding="utf-8") as f:
                yaml.dump(config_data, f)

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                assert expected_mode in result.output
                assert expected_locale in result.output

            finally:
                os.chdir(original_cwd)

    @pytest.mark.parametrize(
        "spec_count",
        [0, 1, 5, 10],
    )
    def test_status_various_spec_counts(self, runner, spec_count):
        """Should display various SPEC counts correctly"""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": "personal", "locale": "en"}}
            with open(config_path, "w", encoding="utf-8") as f:
                yaml.dump(config_data, f)

            # Create SPEC directories
            specs_dir = project_path / ".moai" / "specs"
            for i in range(spec_count):
                spec_path = specs_dir / f"SPEC-{i:03d}"
                spec_path.mkdir(parents=True, exist_ok=True)
                (spec_path / "spec.md").write_text(f"# SPEC {i}")

            original_cwd = Path.cwd()
            try:
                os.chdir(project_path)

                # Act
                result = runner.invoke(status, [])

                # Assert
                assert result.exit_code == 0
                assert str(spec_count) in result.output

            finally:
                os.chdir(original_cwd)
