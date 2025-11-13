"""Unit tests for slash command diagnostics

Tests for diagnosing Claude Code slash command loading issues:
- validate_command_file(): YAML front matter and required fields validation
- scan_command_files(): Recursive .md file scanning
- diagnose_slash_commands(): Comprehensive diagnostic
"""

from pathlib import Path
from textwrap import dedent


class TestValidateCommandFile:
    """Test validate_command_file() function"""

    def test_valid_command_file(self, tmp_path):
        """Should validate a properly formatted command file"""
        # Arrange: Create a valid command file
        cmd_file = tmp_path / "test.md"
        cmd_file.write_text(
            dedent(
                """
                ---
                name: alfred:0-project
                description: Initialize project documentation
                ---

                # Command content
                """
            ).strip()
        )

        # Import will fail until we implement the module
        from moai_adk.core.diagnostics.slash_commands import validate_command_file

        # Act
        result = validate_command_file(cmd_file)

        # Assert
        assert result["valid"] is True
        assert "errors" not in result or len(result["errors"]) == 0

    def test_missing_frontmatter(self, tmp_path):
        """Should detect missing YAML front matter"""
        # Arrange: Create a file without front matter
        cmd_file = tmp_path / "invalid.md"
        cmd_file.write_text("# Just markdown content\nNo YAML here")

        from moai_adk.core.diagnostics.slash_commands import validate_command_file

        # Act
        result = validate_command_file(cmd_file)

        # Assert
        assert result["valid"] is False
        assert any("front matter" in err.lower() for err in result["errors"])

    def test_missing_name_field(self, tmp_path):
        """Should detect missing 'name' field"""
        # Arrange: Create a file with incomplete front matter
        cmd_file = tmp_path / "no-name.md"
        cmd_file.write_text(
            dedent(
                """
                ---
                description: Missing name field
                ---

                # Content
                """
            ).strip()
        )

        from moai_adk.core.diagnostics.slash_commands import validate_command_file

        # Act
        result = validate_command_file(cmd_file)

        # Assert
        assert result["valid"] is False
        assert any("name" in err.lower() for err in result["errors"])

    def test_missing_description_field(self, tmp_path):
        """Should detect missing 'description' field"""
        # Arrange: Create a file with no description
        cmd_file = tmp_path / "no-desc.md"
        cmd_file.write_text(
            dedent(
                """
                ---
                name: alfred:test
                ---

                # Content
                """
            ).strip()
        )

        from moai_adk.core.diagnostics.slash_commands import validate_command_file

        # Act
        result = validate_command_file(cmd_file)

        # Assert
        assert result["valid"] is False
        assert any("description" in err.lower() for err in result["errors"])

    def test_invalid_yaml_syntax(self, tmp_path):
        """Should detect YAML parsing errors"""
        # Arrange: Create a file with broken YAML
        cmd_file = tmp_path / "bad-yaml.md"
        cmd_file.write_text(
            dedent(
                """
                ---
                name: alfred:test
                description: [unclosed bracket
                ---

                # Content
                """
            ).strip()
        )

        from moai_adk.core.diagnostics.slash_commands import validate_command_file

        # Act
        result = validate_command_file(cmd_file)

        # Assert
        assert result["valid"] is False
        assert len(result["errors"]) > 0

    def test_incomplete_frontmatter_delimiter(self, tmp_path):
        """Should detect missing closing delimiter"""
        # Arrange: Create a file with only opening ---
        cmd_file = tmp_path / "incomplete.md"
        cmd_file.write_text(
            dedent(
                """
                ---
                name: alfred:test
                description: Test

                # Content (no closing delimiter)
                """
            ).strip()
        )

        from moai_adk.core.diagnostics.slash_commands import validate_command_file

        # Act
        result = validate_command_file(cmd_file)

        # Assert
        assert result["valid"] is False
        assert any("format" in err.lower() for err in result["errors"])

    def test_file_not_found(self):
        """Should handle file not found gracefully"""
        from moai_adk.core.diagnostics.slash_commands import validate_command_file

        # Act
        result = validate_command_file(Path("/nonexistent/file.md"))

        # Assert
        assert result["valid"] is False
        assert len(result["errors"]) > 0


class TestScanCommandFiles:
    """Test scan_command_files() function"""

    def test_scan_finds_md_files(self, tmp_path):
        """Should find all .md files in directory tree"""
        # Arrange: Create nested .md files
        (tmp_path / "cmd1.md").write_text("# Command 1")
        (tmp_path / "cmd2.md").write_text("# Command 2")
        subdir = tmp_path / "alfred"
        subdir.mkdir()
        (subdir / "cmd3.md").write_text("# Command 3")

        from moai_adk.core.diagnostics.slash_commands import scan_command_files

        # Act
        files = scan_command_files(tmp_path)

        # Assert
        assert len(files) == 3
        assert all(f.suffix == ".md" for f in files)

    def test_scan_ignores_non_md_files(self, tmp_path):
        """Should ignore non-.md files"""
        # Arrange: Create mixed file types
        (tmp_path / "cmd.md").write_text("# Command")
        (tmp_path / "readme.txt").write_text("Text file")
        (tmp_path / "notes.json").write_text("{}")

        from moai_adk.core.diagnostics.slash_commands import scan_command_files

        # Act
        files = scan_command_files(tmp_path)

        # Assert
        assert len(files) == 1
        assert files[0].name == "cmd.md"

    def test_scan_empty_directory(self, tmp_path):
        """Should return empty list for directory with no .md files"""
        from moai_adk.core.diagnostics.slash_commands import scan_command_files

        # Act
        files = scan_command_files(tmp_path)

        # Assert
        assert files == []

    def test_scan_nonexistent_directory(self):
        """Should handle nonexistent directory gracefully"""
        from moai_adk.core.diagnostics.slash_commands import scan_command_files

        # Act
        files = scan_command_files(Path("/nonexistent/directory"))

        # Assert
        assert files == []


class TestDiagnoseSlashCommands:
    """Test diagnose_slash_commands() function"""

    def test_diagnose_all_valid_commands(self, tmp_path, monkeypatch):
        """Should report all commands as valid"""
        # Arrange: Create valid command files
        commands_dir = tmp_path / ".claude" / "commands"
        commands_dir.mkdir(parents=True)

        for i in range(3):
            (commands_dir / f"cmd{i}.md").write_text(
                dedent(
                    f"""
                    ---
                    name: test:cmd{i}
                    description: Test command {i}
                    ---

                    # Content
                    """
                ).strip()
            )

        # Mock the commands directory path
        monkeypatch.chdir(tmp_path)

        from moai_adk.core.diagnostics.slash_commands import diagnose_slash_commands

        # Act
        result = diagnose_slash_commands()

        # Assert
        assert result["total_files"] == 3
        assert result["valid_commands"] == 3
        assert len(result["details"]) == 3

    def test_diagnose_mixed_valid_invalid(self, tmp_path, monkeypatch):
        """Should distinguish between valid and invalid commands"""
        # Arrange: Create mix of valid and invalid files
        commands_dir = tmp_path / ".claude" / "commands"
        commands_dir.mkdir(parents=True)

        # Valid file
        (commands_dir / "valid.md").write_text(
            dedent(
                """
                ---
                name: test:valid
                description: Valid command
                ---

                # Content
                """
            ).strip()
        )

        # Invalid file (missing name)
        (commands_dir / "invalid.md").write_text(
            dedent(
                """
                ---
                description: Missing name
                ---

                # Content
                """
            ).strip()
        )

        monkeypatch.chdir(tmp_path)

        from moai_adk.core.diagnostics.slash_commands import diagnose_slash_commands

        # Act
        result = diagnose_slash_commands()

        # Assert
        assert result["total_files"] == 2
        assert result["valid_commands"] == 1
        assert len(result["details"]) == 2

    def test_diagnose_no_commands_directory(self, tmp_path, monkeypatch):
        """Should handle missing .claude/commands directory"""
        monkeypatch.chdir(tmp_path)

        from moai_adk.core.diagnostics.slash_commands import diagnose_slash_commands

        # Act
        result = diagnose_slash_commands()

        # Assert
        assert "error" in result
        assert "not found" in result["error"].lower()

    def test_diagnose_empty_commands_directory(self, tmp_path, monkeypatch):
        """Should handle empty commands directory"""
        # Arrange: Create empty directory
        commands_dir = tmp_path / ".claude" / "commands"
        commands_dir.mkdir(parents=True)

        monkeypatch.chdir(tmp_path)

        from moai_adk.core.diagnostics.slash_commands import diagnose_slash_commands

        # Act
        result = diagnose_slash_commands()

        # Assert
        assert result["total_files"] == 0
        assert result["valid_commands"] == 0
        assert result["details"] == []


class TestDoctorCheckCommandsIntegration:
    """Integration tests for doctor --check-commands option"""

    def test_doctor_check_commands_option_exists(self):
        """Should accept --check-commands option"""
        from click.testing import CliRunner

        from moai_adk.cli.commands.doctor import doctor

        runner = CliRunner()
        result = runner.invoke(doctor, ["--check-commands"])

        # Should not fail with unknown option
        assert "no such option" not in result.output.lower()

    def test_doctor_check_commands_shows_diagnostics(self, tmp_path, monkeypatch):
        """Should display slash command diagnostics"""
        # Arrange: Create test commands
        commands_dir = tmp_path / ".claude" / "commands"
        commands_dir.mkdir(parents=True)

        (commands_dir / "test.md").write_text(
            dedent(
                """
                ---
                name: test:cmd
                description: Test
                ---

                # Content
                """
            ).strip()
        )

        monkeypatch.chdir(tmp_path)

        from click.testing import CliRunner

        from moai_adk.cli.commands.doctor import doctor

        runner = CliRunner()

        # Act
        result = runner.invoke(doctor, ["--check-commands"])

        # Assert
        assert result.exit_code == 0
        assert "command" in result.output.lower()
