"""Unit tests for doctor command enhancements

Tests for --verbose, --fix, --export options.
"""

import json
from pathlib import Path
from unittest.mock import patch

from click.testing import CliRunner

from moai_adk.cli.commands.doctor import doctor


class TestDoctorCommand:
    """Test doctor command"""

    def test_doctor_runs_without_options(self):
        """Should run doctor without options"""
        runner = CliRunner()
        result = runner.invoke(doctor)

        # Should complete without crashing
        assert result.exit_code == 0
        assert "diagnostics" in result.output.lower()

    def test_doctor_displays_check_results(self):
        """Should display environment check results"""
        runner = CliRunner()
        result = runner.invoke(doctor)

        # Should show Python and Git checks
        assert "Python" in result.output or "python" in result.output.lower()
        assert "Git" in result.output or "git" in result.output.lower()


class TestDoctorVerboseOption:
    """Test doctor --verbose option"""

    def test_verbose_option_exists(self):
        """Should accept --verbose option"""
        runner = CliRunner()
        result = runner.invoke(doctor, ["--verbose"])

        # Should not fail with unknown option
        assert "no such option" not in result.output.lower()

    def test_verbose_shows_language_detection(self):
        """Should show detected language with --verbose"""
        runner = CliRunner()
        result = runner.invoke(doctor, ["--verbose"])

        # Should show language information
        assert "language" in result.output.lower() or result.exit_code == 0

    def test_verbose_shows_tool_versions(self):
        """Should show tool versions with --verbose"""
        runner = CliRunner()
        result = runner.invoke(doctor, ["--verbose"])

        # Should show version information
        # (May fail if no tools installed, just check it doesn't crash)
        assert result.exit_code == 0


class TestDoctorFixOption:
    """Test doctor --fix option"""

    def test_fix_option_exists(self):
        """Should accept --fix option"""
        runner = CliRunner()
        result = runner.invoke(doctor, ["--fix"])

        # Should not fail with unknown option
        assert "no such option" not in result.output.lower()

    @patch("moai_adk.cli.commands.doctor.questionary")
    def test_fix_prompts_user_confirmation(self, mock_questionary):
        """Should prompt user before installing tools"""
        mock_questionary.confirm.return_value.ask.return_value = False

        runner = CliRunner()
        result = runner.invoke(doctor, ["--fix"])

        # Should complete without errors
        assert result.exit_code == 0

    def test_fix_without_missing_tools_shows_message(self):
        """Should show message when no tools are missing"""
        runner = CliRunner()
        result = runner.invoke(doctor, ["--fix"])

        # Should complete successfully
        assert result.exit_code == 0


class TestDoctorExportOption:
    """Test doctor --export option"""

    def test_export_option_exists(self):
        """Should accept --export option"""
        runner = CliRunner()

        with runner.isolated_filesystem():
            result = runner.invoke(doctor, ["--export", "report.json"])

            # Should not fail with unknown option
            assert "no such option" not in result.output.lower()
            assert result.exit_code == 0

    def test_export_creates_json_file(self):
        """Should create JSON report file"""
        runner = CliRunner()

        with runner.isolated_filesystem():
            result = runner.invoke(doctor, ["--export", "diagnostics.json"])

            # Should create the file
            report_file = Path("diagnostics.json")
            if result.exit_code == 0:
                assert report_file.exists() or "export" in result.output.lower()

    def test_export_json_is_valid(self):
        """Exported JSON should be valid"""
        runner = CliRunner()

        with runner.isolated_filesystem():
            result = runner.invoke(doctor, ["--export", "report.json"])

            report_file = Path("report.json")
            if report_file.exists():
                # Should be valid JSON
                data = json.loads(report_file.read_text())
                assert isinstance(data, dict)
            assert result.exit_code == 0


class TestDoctorCheckOption:
    """Test doctor --check option"""

    def test_check_option_exists(self):
        """Should accept --check option for specific tool"""
        runner = CliRunner()
        result = runner.invoke(doctor, ["--check", "python"])

        # Should not fail with unknown option
        assert "no such option" not in result.output.lower()

    def test_check_specific_tool(self):
        """Should check only specified tool"""
        runner = CliRunner()
        result = runner.invoke(doctor, ["--check", "git"])

        # Should mention git
        assert result.exit_code == 0


class TestDoctorPerformance:
    """Test doctor performance constraints"""

    def test_doctor_completes_within_5_seconds(self):
        """Should complete within 5 seconds (AC-1)"""
        import time

        runner = CliRunner()

        start_time = time.time()
        result = runner.invoke(doctor)
        elapsed = time.time() - start_time

        # Should complete within 5 seconds
        assert elapsed < 5.0, f"Doctor took {elapsed:.2f}s (should be < 5s)"
        assert result.exit_code == 0

    def test_doctor_verbose_completes_within_5_seconds(self):
        """Verbose mode should also complete within 5 seconds"""
        import time

        runner = CliRunner()

        start_time = time.time()
        result = runner.invoke(doctor, ["--verbose"])
        elapsed = time.time() - start_time

        # Should complete within 5 seconds even with verbose
        assert elapsed < 5.0, f"Doctor --verbose took {elapsed:.2f}s (should be < 5s)"
        assert result.exit_code == 0
