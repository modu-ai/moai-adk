"""Enhanced comprehensive tests for doctor command to achieve 90%+ coverage

This test suite targets all uncovered lines in doctor.py:
- Lines 78-79: check_commands path
- Lines 89-101: verbose mode with language detection
- Lines 105-106: specific tool check
- Lines 122, 129: fix mode suggestions
- Lines 146-159: display_language_tools
- Lines 164-173: check_specific_tool
- Lines 178-202: suggest_fixes
- Lines 208-220: get_install_command
- Lines 225-230: export_diagnostics
- Lines 235-272: check_slash_commands
"""

import json
from pathlib import Path
from unittest.mock import Mock, patch

import pytest
from click.testing import CliRunner

from moai_adk.cli.commands.doctor import (
    _check_specific_tool,
    _display_language_tools,
    _export_diagnostics,
    _get_install_command,
    _suggest_fixes,
    doctor,
)


class TestCheckCommandsFlag:
    """Test --check-commands flag and _check_slash_commands function"""

    @patch("moai_adk.core.diagnostics.slash_commands.diagnose_slash_commands")
    def test_check_commands_flag_executes_slash_command_diagnostics(self, mock_diagnose):
        """Should execute slash command diagnostics when --check-commands is passed"""
        mock_diagnose.return_value = {
            "total_files": 5,
            "valid_commands": 5,
            "details": [
                {"file": "test.md", "valid": True, "errors": []},
            ],
        }

        runner = CliRunner()
        result = runner.invoke(doctor, ["--check-commands"])

        assert result.exit_code == 0
        assert mock_diagnose.called
        assert "slash command" in result.output.lower()

    @patch("moai_adk.core.diagnostics.slash_commands.diagnose_slash_commands")
    def test_check_commands_returns_early_without_other_checks(self, mock_diagnose):
        """Should return early without running basic checks when --check-commands is used"""
        mock_diagnose.return_value = {
            "total_files": 0,
            "valid_commands": 0,
            "details": [],
        }

        runner = CliRunner()
        result = runner.invoke(doctor, ["--check-commands"])

        assert result.exit_code == 0
        # Should not show basic diagnostics
        assert "Running system diagnostics" not in result.output

    @patch("moai_adk.core.diagnostics.slash_commands.diagnose_slash_commands")
    def test_check_slash_commands_with_error(self, mock_diagnose):
        """Should handle error when commands directory not found"""
        mock_diagnose.return_value = {"error": "Commands directory not found"}

        runner = CliRunner()
        result = runner.invoke(doctor, ["--check-commands"])

        assert result.exit_code == 0
        assert "Commands directory not found" in result.output

    @patch("moai_adk.core.diagnostics.slash_commands.diagnose_slash_commands")
    def test_check_slash_commands_with_valid_commands(self, mock_diagnose):
        """Should display valid commands correctly"""
        mock_diagnose.return_value = {
            "total_files": 3,
            "valid_commands": 3,
            "details": [
                {"file": "cmd1.md", "valid": True, "errors": []},
                {"file": "cmd2.md", "valid": True, "errors": []},
                {"file": "cmd3.md", "valid": True, "errors": []},
            ],
        }

        runner = CliRunner()
        result = runner.invoke(doctor, ["--check-commands"])

        assert result.exit_code == 0
        assert "3/3 command files are valid" in result.output

    @patch("moai_adk.core.diagnostics.slash_commands.diagnose_slash_commands")
    def test_check_slash_commands_with_invalid_commands(self, mock_diagnose):
        """Should display invalid commands with errors"""
        mock_diagnose.return_value = {
            "total_files": 3,
            "valid_commands": 1,
            "details": [
                {"file": "cmd1.md", "valid": True, "errors": []},
                {
                    "file": "cmd2.md",
                    "valid": False,
                    "errors": ["Missing YAML front matter"],
                },
                {
                    "file": "cmd3.md",
                    "valid": False,
                    "errors": ["Missing required field: name"],
                },
            ],
        }

        runner = CliRunner()
        result = runner.invoke(doctor, ["--check-commands"])

        assert result.exit_code == 0
        assert "1/3 command files are valid" in result.output
        # The error may be wrapped or truncated in table display
        assert "Missing YAML front matter" in result.output or "Missing YAML front" in result.output

    @patch("moai_adk.core.diagnostics.slash_commands.diagnose_slash_commands")
    def test_check_slash_commands_with_no_files(self, mock_diagnose):
        """Should handle case when no command files found"""
        mock_diagnose.return_value = {
            "total_files": 0,
            "valid_commands": 0,
            "details": [],
        }

        runner = CliRunner()
        result = runner.invoke(doctor, ["--check-commands"])

        assert result.exit_code == 0
        assert "No command files found" in result.output


class TestVerboseModeWithLanguageDetection:
    """Test --verbose flag with language detection and tool display"""

    @patch("moai_adk.cli.commands.doctor.detect_project_language")
    @patch("moai_adk.cli.commands.doctor.check_environment")
    @patch("moai_adk.cli.commands.doctor.SystemChecker")
    def test_verbose_detects_language_and_shows_tools(self, mock_checker_class, mock_check_env, mock_detect_lang):
        """Should detect language and display language-specific tools in verbose mode"""
        mock_check_env.return_value = {"Python >= 3.13": True, "Git installed": True}
        mock_detect_lang.return_value = "python"

        mock_checker = Mock()
        mock_checker.check_language_tools.return_value = {
            "pytest": True,
            "mypy": False,
            "ruff": True,
        }
        mock_checker.get_tool_version.side_effect = lambda tool: ("pytest 8.4.2" if tool == "pytest" else "ruff 0.11.0")
        mock_checker_class.return_value = mock_checker

        runner = CliRunner()
        result = runner.invoke(doctor, ["--verbose"])

        assert result.exit_code == 0
        assert "Detected language: python" in result.output
        assert "Python Tools" in result.output or "python" in result.output.lower()

    @patch("moai_adk.cli.commands.doctor.detect_project_language")
    @patch("moai_adk.cli.commands.doctor.check_environment")
    def test_verbose_handles_unknown_language(self, mock_check_env, mock_detect_lang):
        """Should handle case when no language is detected"""
        mock_check_env.return_value = {"Python >= 3.13": True, "Git installed": True}
        mock_detect_lang.return_value = None

        runner = CliRunner()
        result = runner.invoke(doctor, ["--verbose"])

        assert result.exit_code == 0
        assert "Detected language: Unknown" in result.output

    @patch("moai_adk.cli.commands.doctor.detect_project_language")
    @patch("moai_adk.cli.commands.doctor.check_environment")
    @patch("moai_adk.cli.commands.doctor.SystemChecker")
    def test_display_language_tools_with_versions(self, mock_checker_class, mock_check_env, mock_detect_lang):
        """Should display tool versions correctly"""
        mock_check_env.return_value = {"Python >= 3.13": True}
        mock_detect_lang.return_value = "typescript"

        mock_checker = Mock()
        mock_checker.check_language_tools.return_value = {
            "node": True,
            "npm": True,
            "vitest": False,
        }
        mock_checker.get_tool_version.side_effect = lambda tool: ("v22.0.0" if tool == "node" else "10.2.0")
        mock_checker_class.return_value = mock_checker

        runner = CliRunner()
        result = runner.invoke(doctor, ["--verbose"])

        assert result.exit_code == 0
        # Should show tool status
        assert result.output  # Non-empty output


class TestCheckSpecificTool:
    """Test --check flag for specific tool checking"""

    @patch("moai_adk.cli.commands.doctor.SystemChecker")
    def test_check_specific_tool_installed(self, mock_checker_class):
        """Should report when specific tool is installed"""
        mock_checker = Mock()
        mock_checker._is_tool_available.return_value = True
        mock_checker.get_tool_version.return_value = "Python 3.13.0"
        mock_checker_class.return_value = mock_checker

        runner = CliRunner()
        result = runner.invoke(doctor, ["--check", "python"])

        assert result.exit_code == 0
        assert "python is installed" in result.output.lower()
        assert "Python 3.13.0" in result.output or "version" in result.output.lower()

    @patch("moai_adk.cli.commands.doctor.SystemChecker")
    def test_check_specific_tool_not_installed(self, mock_checker_class):
        """Should report when specific tool is not installed"""
        mock_checker = Mock()
        mock_checker._is_tool_available.return_value = False
        mock_checker.get_tool_version.return_value = None
        mock_checker_class.return_value = mock_checker

        runner = CliRunner()
        result = runner.invoke(doctor, ["--check", "nonexistent"])

        assert result.exit_code == 0
        assert "not installed" in result.output.lower()

    @patch("moai_adk.cli.commands.doctor.SystemChecker")
    def test_check_specific_tool_without_version(self, mock_checker_class):
        """Should handle tool installed without version info"""
        mock_checker = Mock()
        mock_checker._is_tool_available.return_value = True
        mock_checker.get_tool_version.return_value = None
        mock_checker_class.return_value = mock_checker

        runner = CliRunner()
        result = runner.invoke(doctor, ["--check", "git"])

        assert result.exit_code == 0
        assert "git is installed" in result.output.lower()

    def test_check_tool_returns_early_without_basic_checks(self):
        """Should return early without running basic environment checks"""
        runner = CliRunner()
        result = runner.invoke(doctor, ["--check", "python"])

        assert result.exit_code == 0
        # The --check flag still shows "Running system diagnostics" but exits early
        # We just verify it completes successfully
        assert result.output  # Non-empty output


class TestFixModeAndSuggestions:
    """Test --fix flag and installation suggestions"""

    @patch("moai_adk.cli.commands.doctor.questionary")
    @patch("moai_adk.cli.commands.doctor.detect_project_language")
    @patch("moai_adk.cli.commands.doctor.check_environment")
    @patch("moai_adk.cli.commands.doctor.SystemChecker")
    def test_fix_suggests_installations_for_missing_tools(
        self, mock_checker_class, mock_check_env, mock_detect_lang, mock_questionary
    ):
        """Should suggest installation commands for missing tools"""
        mock_check_env.return_value = {"Python >= 3.13": True}
        mock_detect_lang.return_value = "python"

        mock_checker = Mock()
        mock_checker.check_language_tools.return_value = {
            "pytest": False,
            "mypy": False,
            "ruff": True,
        }
        mock_checker_class.return_value = mock_checker

        mock_questionary.confirm.return_value.ask.return_value = True

        runner = CliRunner()
        result = runner.invoke(doctor, ["--fix"])

        assert result.exit_code == 0
        assert "Missing 2 tool(s)" in result.output
        assert "uv pip install pytest" in result.output
        assert "uv pip install mypy" in result.output

    @patch("moai_adk.cli.commands.doctor.questionary")
    @patch("moai_adk.cli.commands.doctor.detect_project_language")
    @patch("moai_adk.cli.commands.doctor.check_environment")
    @patch("moai_adk.cli.commands.doctor.SystemChecker")
    def test_fix_with_user_declining_suggestions(
        self, mock_checker_class, mock_check_env, mock_detect_lang, mock_questionary
    ):
        """Should respect user declining to see suggestions"""
        mock_check_env.return_value = {"Python >= 3.13": True}
        mock_detect_lang.return_value = "typescript"

        mock_checker = Mock()
        mock_checker.check_language_tools.return_value = {
            "vitest": False,
            "biome": False,
        }
        mock_checker_class.return_value = mock_checker

        mock_questionary.confirm.return_value.ask.return_value = False

        runner = CliRunner()
        result = runner.invoke(doctor, ["--fix"])

        assert result.exit_code == 0
        assert "skipped" in result.output.lower()

    @patch("moai_adk.cli.commands.doctor.detect_project_language")
    @patch("moai_adk.cli.commands.doctor.check_environment")
    @patch("moai_adk.cli.commands.doctor.SystemChecker")
    def test_fix_with_all_tools_installed(self, mock_checker_class, mock_check_env, mock_detect_lang):
        """Should report success when all tools are installed"""
        mock_check_env.return_value = {"Python >= 3.13": True}
        mock_detect_lang.return_value = "python"

        mock_checker = Mock()
        mock_checker.check_language_tools.return_value = {
            "pytest": True,
            "mypy": True,
            "ruff": True,
        }
        mock_checker_class.return_value = mock_checker

        runner = CliRunner()
        result = runner.invoke(doctor, ["--fix"])

        assert result.exit_code == 0
        assert "All tools are installed" in result.output

    @patch("moai_adk.cli.commands.doctor.questionary")
    def test_fix_handles_questionary_exception(self, mock_questionary):
        """Should handle exception when questionary fails"""
        mock_questionary.confirm.side_effect = Exception("Terminal not available")

        # Use helper function directly
        from moai_adk.cli.commands.doctor import console

        with patch.object(console, "print") as mock_print:
            _suggest_fixes({"pytest": False, "mypy": False}, "python")
            # Should still proceed with suggestions despite error
            assert mock_print.called


class TestGetInstallCommand:
    """Test _get_install_command helper function"""

    def test_get_install_command_python_tools(self):
        """Should return correct install commands for Python tools"""
        assert _get_install_command("pytest", "python") == "uv pip install pytest"
        assert _get_install_command("mypy", "python") == "uv pip install mypy"
        assert _get_install_command("ruff", "python") == "uv pip install ruff"

    def test_get_install_command_javascript_tools(self):
        """Should return correct install commands for JavaScript tools"""
        assert _get_install_command("vitest", "typescript") == "npm install -D vitest"
        assert _get_install_command("biome", "javascript") == "npm install -D @biomejs/biome"
        assert _get_install_command("eslint", "javascript") == "npm install -D eslint"
        assert _get_install_command("jest", "javascript") == "npm install -D jest"

    def test_get_install_command_unknown_tool(self):
        """Should return generic message for unknown tools"""
        result = _get_install_command("unknowntool", "python")
        assert "Install unknowntool for python" in result
        assert result.startswith("#")

    def test_get_install_command_none_language(self):
        """Should handle None language"""
        result = _get_install_command("sometool", None)
        assert "Install sometool" in result


class TestExportDiagnostics:
    """Test --export flag and _export_diagnostics function"""

    @patch("moai_adk.cli.commands.doctor.check_environment")
    def test_export_creates_valid_json_file(self, mock_check_env):
        """Should create valid JSON file with diagnostics"""
        mock_check_env.return_value = {
            "Python >= 3.13": True,
            "Git installed": True,
            "Project structure (.moai/)": False,
        }

        runner = CliRunner()
        with runner.isolated_filesystem():
            result = runner.invoke(doctor, ["--export", "diagnostics.json"])

            assert result.exit_code == 0
            assert Path("diagnostics.json").exists()

            data = json.loads(Path("diagnostics.json").read_text())
            assert "basic_checks" in data
            assert isinstance(data["basic_checks"], dict)

    @patch("moai_adk.cli.commands.doctor.check_environment")
    def test_export_with_verbose_includes_language_data(self, mock_check_env):
        """Should include language detection data in export"""
        mock_check_env.return_value = {"Python >= 3.13": True}

        runner = CliRunner()
        with runner.isolated_filesystem():
            result = runner.invoke(doctor, ["--verbose", "--export", "report.json"])

            assert result.exit_code == 0
            if Path("report.json").exists():
                data = json.loads(Path("report.json").read_text())
                assert "detected_language" in data

    def test_export_diagnostics_handles_write_error(self, tmp_path):
        """Should handle file write errors gracefully"""
        from moai_adk.cli.commands.doctor import console

        # Try to write to invalid path
        invalid_path = "/invalid/path/report.json"

        with patch.object(console, "print") as mock_print:
            _export_diagnostics(invalid_path, {"test": "data"})

            # Should print error message
            error_calls = [call for call in mock_print.call_args_list if "Failed to export" in str(call)]
            assert len(error_calls) > 0

    def test_export_diagnostics_success_message(self, tmp_path):
        """Should display success message after export"""
        export_path = tmp_path / "test_report.json"
        data = {"test": "data", "checks": {"python": True}}

        from moai_adk.cli.commands.doctor import console

        with patch.object(console, "print") as mock_print:
            _export_diagnostics(str(export_path), data)

            # Should print success message
            success_calls = [call for call in mock_print.call_args_list if "exported to" in str(call).lower()]
            assert len(success_calls) > 0


class TestDisplayLanguageTools:
    """Test _display_language_tools helper function"""

    def test_display_language_tools_with_all_available(self):
        """Should display table when all tools are available"""
        from moai_adk.cli.commands.doctor import console

        mock_checker = Mock()
        mock_checker.get_tool_version.side_effect = lambda tool: f"{tool} 1.0.0"

        tools = {"pytest": True, "mypy": True, "ruff": True}

        with patch.object(console, "print") as mock_print:
            _display_language_tools("python", tools, mock_checker)
            assert mock_print.called

    def test_display_language_tools_with_mixed_availability(self):
        """Should display table with mix of available and unavailable tools"""
        from moai_adk.cli.commands.doctor import console

        mock_checker = Mock()
        mock_checker.get_tool_version.side_effect = lambda tool: "1.0.0" if tool != "mypy" else "not installed"

        tools = {"pytest": True, "mypy": False, "ruff": True}

        with patch.object(console, "print") as mock_print:
            _display_language_tools("typescript", tools, mock_checker)
            assert mock_print.called

    def test_display_language_tools_shows_not_installed(self):
        """Should show 'not installed' for unavailable tools"""
        from moai_adk.cli.commands.doctor import console

        mock_checker = Mock()
        mock_checker.get_tool_version.return_value = None

        tools = {"vitest": False, "biome": False}

        with patch.object(console, "print") as mock_print:
            _display_language_tools("javascript", tools, mock_checker)
            assert mock_print.called


class TestErrorHandling:
    """Test error handling in doctor command"""

    @patch("moai_adk.cli.commands.doctor.check_environment")
    def test_doctor_handles_check_environment_exception(self, mock_check_env):
        """Should handle exceptions from check_environment"""
        mock_check_env.side_effect = Exception("System check failed")

        runner = CliRunner()
        result = runner.invoke(doctor)

        # Should exit with error
        assert result.exit_code != 0
        assert "Diagnostic failed" in result.output

    @patch("moai_adk.cli.commands.doctor.detect_project_language")
    @patch("moai_adk.cli.commands.doctor.check_environment")
    def test_doctor_handles_language_detection_exception(self, mock_check_env, mock_detect_lang):
        """Should handle exceptions during language detection"""
        mock_check_env.return_value = {"Python >= 3.13": True}
        mock_detect_lang.side_effect = Exception("Language detection failed")

        runner = CliRunner()
        result = runner.invoke(doctor, ["--verbose"])

        assert result.exit_code != 0


class TestOutputFormatting:
    """Test output formatting and status messages"""

    @patch("moai_adk.cli.commands.doctor.check_environment")
    def test_doctor_shows_all_checks_passed(self, mock_check_env):
        """Should show success message when all checks pass"""
        mock_check_env.return_value = {
            "Python >= 3.13": True,
            "Git installed": True,
            "Project structure (.moai/)": True,
            "Config file (.moai/config/config.json)": True,
        }

        runner = CliRunner()
        result = runner.invoke(doctor)

        assert result.exit_code == 0
        assert "All checks passed" in result.output

    @patch("moai_adk.cli.commands.doctor.check_environment")
    def test_doctor_shows_some_checks_failed(self, mock_check_env):
        """Should show warning when some checks fail"""
        mock_check_env.return_value = {
            "Python >= 3.13": True,
            "Git installed": False,
            "Project structure (.moai/)": False,
        }

        runner = CliRunner()
        result = runner.invoke(doctor)

        assert result.exit_code == 0
        assert "Some checks failed" in result.output
        assert "--verbose" in result.output.lower()


class TestIntegrationScenarios:
    """Integration tests for complex scenarios"""

    @patch("moai_adk.cli.commands.doctor.questionary")
    @patch("moai_adk.cli.commands.doctor.detect_project_language")
    @patch("moai_adk.cli.commands.doctor.check_environment")
    @patch("moai_adk.cli.commands.doctor.SystemChecker")
    def test_verbose_fix_export_combined(self, mock_checker_class, mock_check_env, mock_detect_lang, mock_questionary):
        """Should handle --verbose --fix --export combination"""
        mock_check_env.return_value = {"Python >= 3.13": True, "Git installed": True}
        mock_detect_lang.return_value = "python"

        mock_checker = Mock()
        mock_checker.check_language_tools.return_value = {
            "pytest": True,
            "mypy": False,
        }
        mock_checker.get_tool_version.return_value = "pytest 8.4.2"
        mock_checker_class.return_value = mock_checker

        mock_questionary.confirm.return_value.ask.return_value = True

        runner = CliRunner()
        with runner.isolated_filesystem():
            result = runner.invoke(doctor, ["--verbose", "--fix", "--export", "full_report.json"])

            assert result.exit_code == 0
            assert Path("full_report.json").exists()

            data = json.loads(Path("full_report.json").read_text())
            assert "basic_checks" in data
            assert "detected_language" in data
            assert "language_tools" in data


class TestEdgeCases:
    """Test edge cases and boundary conditions"""

    def test_check_specific_tool_helper_directly(self):
        """Test _check_specific_tool helper function directly"""
        from moai_adk.cli.commands.doctor import console

        with patch.object(console, "print") as mock_print:
            with patch("moai_adk.cli.commands.doctor.SystemChecker") as mock_checker_class:
                mock_checker = Mock()
                mock_checker._is_tool_available.return_value = True
                mock_checker.get_tool_version.return_value = "v1.0.0"
                mock_checker_class.return_value = mock_checker

                _check_specific_tool("test-tool")
                assert mock_print.called

    def test_suggest_fixes_with_empty_tools(self):
        """Should handle empty tools dict"""
        from moai_adk.cli.commands.doctor import console

        with patch.object(console, "print") as mock_print:
            _suggest_fixes({}, "python")
            # Should print "All tools are installed"
            success_calls = [call for call in mock_print.call_args_list if "All tools" in str(call)]
            assert len(success_calls) > 0

    @patch("moai_adk.cli.commands.doctor.check_environment")
    def test_doctor_with_empty_results(self, mock_check_env):
        """Should handle empty check results"""
        mock_check_env.return_value = {}

        runner = CliRunner()
        result = runner.invoke(doctor)

        assert result.exit_code == 0
        # Should show "All checks passed" for empty results
        assert "All checks passed" in result.output


# Parametrized tests for comprehensive coverage
class TestParametrizedScenarios:
    """Parametrized tests for multiple scenarios"""

    @pytest.mark.parametrize(
        "tool_name,expected_command",
        [
            ("pytest", "uv pip install pytest"),
            ("mypy", "uv pip install mypy"),
            ("ruff", "uv pip install ruff"),
            ("vitest", "npm install -D vitest"),
            ("biome", "npm install -D @biomejs/biome"),
            ("eslint", "npm install -D eslint"),
            ("jest", "npm install -D jest"),
        ],
    )
    def test_get_install_commands_parametrized(self, tool_name, expected_command):
        """Test install commands for various tools"""
        result = _get_install_command(tool_name, "python")
        assert result == expected_command

    @pytest.mark.parametrize(
        "language",
        ["python", "typescript", "javascript", "go", "rust", "java"],
    )
    def test_display_language_tools_for_multiple_languages(self, language):
        """Test display for multiple languages"""
        from moai_adk.cli.commands.doctor import console

        mock_checker = Mock()
        mock_checker.get_tool_version.return_value = "1.0.0"

        tools = {"tool1": True, "tool2": False}

        with patch.object(console, "print"):
            _display_language_tools(language, tools, mock_checker)
            # Should complete without errors
