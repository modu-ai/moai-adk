"""Comprehensive test coverage for status command.

Focus on uncovered code paths with mocked dependencies.
Tests actual code paths without side effects.
"""

import tempfile
from pathlib import Path
from unittest.mock import patch

import yaml
from click.testing import CliRunner

from moai_adk.cli.commands.status import status


class TestStatusCommand:
    """Test status command function."""

    def test_status_no_config_file(self):
        """Test status command when config file doesn't exist."""
        # Act
        runner = CliRunner()
        with runner.isolated_filesystem():
            result = runner.invoke(status, [])

            # Assert
            assert result.exit_code != 0
            assert "No .moai/config/config.yaml found" in result.output

    def test_status_with_config_file(self):
        """Test status command with valid config file."""
        runner = CliRunner()
        with runner.isolated_filesystem() as tmpdir:
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
            with open(config_path, "w") as f:
                yaml.dump(config_data, f)

            # Patch Path.cwd() to return tmpdir instead of the mocked value from conftest
            with patch("pathlib.Path.cwd", return_value=project_path):
                result = runner.invoke(status, [])

            # Assert
            assert result.exit_code == 0
            assert "personal" in result.output

    def test_status_with_legacy_config_format(self):
        """Test status command with legacy config format."""
        runner = CliRunner()
        with runner.isolated_filesystem() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            # Legacy format with mode/locale at top level
            config_data = {
                "mode": "team",
                "locale": "ko",
            }
            with open(config_path, "w") as f:
                yaml.dump(config_data, f)

            with patch("pathlib.Path.cwd", return_value=project_path):
                result = runner.invoke(status, [])

            # Assert
            assert result.exit_code == 0
            assert "team" in result.output

    def test_status_counts_spec_documents(self):
        """Test status command counts SPEC documents."""
        runner = CliRunner()
        with runner.isolated_filesystem() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": "personal", "locale": "en"}}
            with open(config_path, "w") as f:
                yaml.dump(config_data, f)

            # Create SPEC files
            specs_dir = project_path / ".moai" / "specs"
            for i in range(3):
                spec_path = specs_dir / f"SPEC-AUTH-00{i + 1}"
                spec_path.mkdir(parents=True, exist_ok=True)
                (spec_path / "spec.md").write_text(f"# SPEC-AUTH-00{i + 1}")

            with patch("pathlib.Path.cwd", return_value=project_path):
                result = runner.invoke(status, [])

            # Assert
            assert result.exit_code == 0
            assert "3" in result.output

    def test_status_no_specs_directory(self):
        """Test status command when specs directory doesn't exist."""
        runner = CliRunner()
        with runner.isolated_filesystem() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": "personal", "locale": "en"}}
            with open(config_path, "w") as f:
                yaml.dump(config_data, f)

            with patch("pathlib.Path.cwd", return_value=project_path):
                result = runner.invoke(status, [])

            # Assert
            assert result.exit_code == 0
            assert "0" in result.output

    def test_status_with_multiple_specs(self):
        """Test status command with multiple SPEC documents."""
        runner = CliRunner()
        with runner.isolated_filesystem() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": "team", "locale": "ko"}}
            with open(config_path, "w") as f:
                yaml.dump(config_data, f)

            # Create multiple SPEC files
            specs_dir = project_path / ".moai" / "specs"
            spec_ids = [
                "SPEC-AUTH-001",
                "SPEC-DB-001",
                "SPEC-API-001",
                "SPEC-UI-001",
                "SPEC-TEST-001",
            ]
            for spec_id in spec_ids:
                spec_path = specs_dir / spec_id
                spec_path.mkdir(parents=True, exist_ok=True)
                (spec_path / "spec.md").write_text(f"# {spec_id}")

            with patch("pathlib.Path.cwd", return_value=project_path):
                result = runner.invoke(status, [])

            # Assert
            assert result.exit_code == 0
            assert "5" in result.output
            assert "team" in result.output

    def test_status_config_with_unknown_values(self):
        """Test status with config values set to unknown."""
        runner = CliRunner()
        with runner.isolated_filesystem() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            # Config with missing values
            config_data = {"project": {}}
            with open(config_path, "w") as f:
                yaml.dump(config_data, f)

            with patch("pathlib.Path.cwd", return_value=project_path):
                result = runner.invoke(status, [])

            # Assert
            assert result.exit_code == 0
            assert "unknown" in result.output

    def test_status_output_format(self):
        """Test status command output format."""
        runner = CliRunner()
        with runner.isolated_filesystem() as tmpdir:
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
            with open(config_path, "w") as f:
                yaml.dump(config_data, f)

            with patch("pathlib.Path.cwd", return_value=project_path):
                result = runner.invoke(status, [])

            # Assert
            assert result.exit_code == 0
            assert "Mode" in result.output
            assert "Locale" in result.output
            assert "SPECs" in result.output

    def test_status_with_empty_locale(self):
        """Test status with empty locale."""
        runner = CliRunner()
        with runner.isolated_filesystem() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {
                "project": {
                    "mode": "team",
                    "locale": None,
                }
            }
            with open(config_path, "w") as f:
                yaml.dump(config_data, f)

            with patch("pathlib.Path.cwd", return_value=project_path):
                result = runner.invoke(status, [])

            # Assert
            assert result.exit_code == 0

    def test_status_invalid_yaml_config(self):
        """Test status with invalid YAML config file."""
        runner = CliRunner()
        with runner.isolated_filesystem() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            # Write invalid YAML
            with open(config_path, "w") as f:
                f.write("invalid: yaml: content:")

            with patch("pathlib.Path.cwd", return_value=project_path):
                result = runner.invoke(status, [])

            # Assert
            assert result.exit_code != 0


class TestStatusEdgeCases:
    """Test edge cases and error conditions."""

    def test_status_with_special_characters_in_locale(self):
        """Test status with special characters in locale."""
        runner = CliRunner()
        with runner.isolated_filesystem() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {
                "project": {
                    "mode": "personal",
                    "locale": "en_US-utf8",
                }
            }
            with open(config_path, "w") as f:
                yaml.dump(config_data, f)

            with patch("pathlib.Path.cwd", return_value=project_path):
                result = runner.invoke(status, [])

            # Assert
            assert result.exit_code == 0
            assert "en_US-utf8" in result.output

    def test_status_with_nested_specs(self):
        """Test status correctly counts only top-level SPEC directories."""
        runner = CliRunner()
        with runner.isolated_filesystem() as tmpdir:
            project_path = Path(tmpdir)
            config_dir = project_path / ".moai" / "config"
            config_dir.mkdir(parents=True, exist_ok=True)
            config_path = config_dir / "config.yaml"
            config_data = {"project": {"mode": "personal", "locale": "en"}}
            with open(config_path, "w") as f:
                yaml.dump(config_data, f)

            # Create SPEC files with nested structure
            specs_dir = project_path / ".moai" / "specs"
            spec_path = specs_dir / "SPEC-MAIN-001"
            spec_path.mkdir(parents=True, exist_ok=True)
            (spec_path / "spec.md").write_text("# Main")

            # Create nested (should not be counted)
            nested_path = spec_path / "SPEC-NESTED-001"
            nested_path.mkdir(parents=True, exist_ok=True)
            (nested_path / "spec.md").write_text("# Nested")

            with patch("pathlib.Path.cwd", return_value=project_path):
                result = runner.invoke(status, [])

            # Assert
            assert result.exit_code == 0
            assert "1" in result.output
