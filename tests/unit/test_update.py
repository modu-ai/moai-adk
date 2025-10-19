# @TEST:TEST-COVERAGE-001 | SPEC: SPEC-TEST-COVERAGE-001.md
"""Unit tests for update.py command

Tests for package upgrade functionality (not template updates).
"""

from unittest.mock import Mock, patch

from click.testing import CliRunner

from moai_adk.cli.commands.update import detect_install_method, get_latest_version, update, upgrade_package


class TestGetLatestVersion:
    """Test get_latest_version function"""

    def test_get_latest_version_success(self):
        """Test successful PyPI version fetch"""
        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = Mock()
            mock_response.read.return_value = b'{"info": {"version": "0.4.0"}}'
            mock_response.__enter__ = Mock(return_value=mock_response)
            mock_response.__exit__ = Mock(return_value=False)
            mock_urlopen.return_value = mock_response

            version = get_latest_version()
            assert version == "0.4.0"

    def test_get_latest_version_timeout(self):
        """Test PyPI fetch timeout"""
        with patch("urllib.request.urlopen", side_effect=TimeoutError):
            version = get_latest_version()
            assert version is None

    def test_get_latest_version_network_error(self):
        """Test PyPI fetch network error"""
        import urllib.error

        with patch("urllib.request.urlopen", side_effect=urllib.error.URLError("Network error")):
            version = get_latest_version()
            assert version is None

    def test_get_latest_version_invalid_json(self):
        """Test PyPI fetch with invalid JSON response"""
        with patch("urllib.request.urlopen") as mock_urlopen:
            mock_response = Mock()
            mock_response.read.return_value = b"invalid json"
            mock_response.__enter__ = Mock(return_value=mock_response)
            mock_response.__exit__ = Mock(return_value=False)
            mock_urlopen.return_value = mock_response

            version = get_latest_version()
            assert version is None


class TestDetectInstallMethod:
    """Test detect_install_method function"""

    def test_detect_uv_tool(self):
        """Test detection of uv tool installation"""
        with patch("subprocess.run") as mock_run:
            # First call: uv tool list (success with moai-adk in output)
            mock_run.return_value = Mock(returncode=0, stdout="moai-adk v0.3.13")
            method = detect_install_method()
            assert method == "uv-tool"

    def test_detect_uv_pip(self):
        """Test detection of uv pip installation"""
        with patch("subprocess.run") as mock_run:
            # First call: uv tool list (fails or no moai-adk)
            # Second call: uv --version (success)
            def side_effect(*args, **kwargs):
                cmd = args[0]
                if "tool" in cmd:
                    return Mock(returncode=1, stdout="")
                elif "--version" in cmd:
                    return Mock(returncode=0, stdout="uv 0.1.0")
                return Mock(returncode=1)

            mock_run.side_effect = side_effect
            method = detect_install_method()
            assert method == "uv-pip"

    def test_detect_pip(self):
        """Test detection of pip installation (fallback)"""
        with patch("subprocess.run", side_effect=FileNotFoundError):
            method = detect_install_method()
            assert method == "pip"

    def test_detect_uv_tool_timeout(self):
        """Test uv tool detection with timeout"""
        import subprocess

        with patch("subprocess.run", side_effect=subprocess.TimeoutExpired("uv", 5)):
            method = detect_install_method()
            # Should fallback to pip
            assert method == "pip"


class TestUpgradePackage:
    """Test upgrade_package function"""

    def test_upgrade_uv_tool_success(self):
        """Test successful upgrade via uv tool"""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = Mock(returncode=0, stdout="", stderr="")
            success = upgrade_package("uv-tool", "0.4.0")
            assert success is True
            mock_run.assert_called_once()
            assert "uv" in mock_run.call_args[0][0]
            assert "tool" in mock_run.call_args[0][0]
            assert "upgrade" in mock_run.call_args[0][0]

    def test_upgrade_uv_pip_success(self):
        """Test successful upgrade via uv pip"""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = Mock(returncode=0, stdout="", stderr="")
            success = upgrade_package("uv-pip", "0.4.0")
            assert success is True
            mock_run.assert_called_once()
            assert "uv" in mock_run.call_args[0][0]
            assert "pip" in mock_run.call_args[0][0]
            assert "install" in mock_run.call_args[0][0]

    def test_upgrade_pip_success(self):
        """Test successful upgrade via pip"""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = Mock(returncode=0, stdout="", stderr="")
            success = upgrade_package("pip", "0.4.0")
            assert success is True
            mock_run.assert_called_once()
            assert "pip" in mock_run.call_args[0][0]
            assert "install" in mock_run.call_args[0][0]

    def test_upgrade_failure(self):
        """Test failed upgrade"""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = Mock(returncode=1, stdout="", stderr="Error message")
            success = upgrade_package("pip", "0.4.0")
            assert success is False

    def test_upgrade_timeout(self):
        """Test upgrade timeout"""
        import subprocess

        with patch("subprocess.run", side_effect=subprocess.TimeoutExpired("pip", 120)):
            success = upgrade_package("pip", "0.4.0")
            assert success is False

    def test_upgrade_invalid_method(self):
        """Test upgrade with invalid install method"""
        success = upgrade_package("invalid-method", "0.4.0")
        assert success is False


class TestUpdateCommand:
    """Test update command"""

    def test_update_help(self):
        """Test update --help"""
        runner = CliRunner()
        result = runner.invoke(update, ["--help"])
        assert result.exit_code == 0
        assert "Upgrade moai-adk package" in result.output
        assert "--check" in result.output

    def test_update_check_update_available(self):
        """Test update --check when new version is available"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.update.get_latest_version") as mock_get_version:
            with patch("moai_adk.cli.commands.update.__version__", "0.3.0"):
                mock_get_version.return_value = "0.4.0"
                result = runner.invoke(update, ["--check"])
                assert result.exit_code == 0
                assert "Checking versions" in result.output
                assert "0.3.0" in result.output
                assert "0.4.0" in result.output
                assert "Update available" in result.output

    def test_update_check_already_up_to_date(self):
        """Test update --check when already up to date"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.update.get_latest_version") as mock_get_version:
            with patch("moai_adk.cli.commands.update.__version__", "0.4.0"):
                mock_get_version.return_value = "0.4.0"
                result = runner.invoke(update, ["--check"])
                assert result.exit_code == 0
                assert "Already up to date" in result.output

    def test_update_check_development_version(self):
        """Test update --check when local version is newer (development)"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.update.get_latest_version") as mock_get_version:
            with patch("moai_adk.cli.commands.update.__version__", "0.5.0"):
                mock_get_version.return_value = "0.4.0"
                result = runner.invoke(update, ["--check"])
                assert result.exit_code == 0
                assert "Development version" in result.output

    def test_update_check_pypi_failure(self):
        """Test update --check when PyPI fetch fails"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.update.get_latest_version") as mock_get_version:
            mock_get_version.return_value = None
            result = runner.invoke(update, ["--check"])
            assert result.exit_code == 0
            assert "Unable to fetch from PyPI" in result.output
            assert "Cannot check for updates" in result.output

    def test_update_already_up_to_date(self):
        """Test update when already up to date (no upgrade needed)"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.update.get_latest_version") as mock_get_version:
            with patch("moai_adk.cli.commands.update.__version__", "0.4.0"):
                mock_get_version.return_value = "0.4.0"
                result = runner.invoke(update)
                assert result.exit_code == 0
                assert "Already up to date" in result.output

    def test_update_successful_upgrade_uv_tool(self):
        """Test successful upgrade via uv tool"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.update.get_latest_version") as mock_get_version:
            with patch("moai_adk.cli.commands.update.__version__", "0.3.0"):
                with patch("moai_adk.cli.commands.update.detect_install_method") as mock_detect:
                    with patch("moai_adk.cli.commands.update.upgrade_package") as mock_upgrade:
                        mock_get_version.return_value = "0.4.0"
                        mock_detect.return_value = "uv-tool"
                        mock_upgrade.return_value = True

                        result = runner.invoke(update)
                        assert result.exit_code == 0
                        assert "Detected installation method: uv-tool" in result.output
                        assert "Update complete" in result.output
                        assert "For template updates, run: moai-adk init ." in result.output

    def test_update_successful_upgrade_uv_pip(self):
        """Test successful upgrade via uv pip"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.update.get_latest_version") as mock_get_version:
            with patch("moai_adk.cli.commands.update.__version__", "0.3.0"):
                with patch("moai_adk.cli.commands.update.detect_install_method") as mock_detect:
                    with patch("moai_adk.cli.commands.update.upgrade_package") as mock_upgrade:
                        mock_get_version.return_value = "0.4.0"
                        mock_detect.return_value = "uv-pip"
                        mock_upgrade.return_value = True

                        result = runner.invoke(update)
                        assert result.exit_code == 0
                        assert "Detected installation method: uv-pip" in result.output
                        assert "Update complete" in result.output

    def test_update_successful_upgrade_pip(self):
        """Test successful upgrade via pip"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.update.get_latest_version") as mock_get_version:
            with patch("moai_adk.cli.commands.update.__version__", "0.3.0"):
                with patch("moai_adk.cli.commands.update.detect_install_method") as mock_detect:
                    with patch("moai_adk.cli.commands.update.upgrade_package") as mock_upgrade:
                        mock_get_version.return_value = "0.4.0"
                        mock_detect.return_value = "pip"
                        mock_upgrade.return_value = True

                        result = runner.invoke(update)
                        assert result.exit_code == 0
                        assert "Detected installation method: pip" in result.output
                        assert "Update complete" in result.output

    def test_update_failed_upgrade_uv_tool(self):
        """Test failed upgrade via uv tool (shows manual instructions)"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.update.get_latest_version") as mock_get_version:
            with patch("moai_adk.cli.commands.update.__version__", "0.3.0"):
                with patch("moai_adk.cli.commands.update.detect_install_method") as mock_detect:
                    with patch("moai_adk.cli.commands.update.upgrade_package") as mock_upgrade:
                        mock_get_version.return_value = "0.4.0"
                        mock_detect.return_value = "uv-tool"
                        mock_upgrade.return_value = False

                        result = runner.invoke(update)
                        assert result.exit_code == 1
                        assert "Upgrade failed" in result.output
                        assert "uv tool upgrade moai-adk" in result.output

    def test_update_failed_upgrade_uv_pip(self):
        """Test failed upgrade via uv pip (shows manual instructions)"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.update.get_latest_version") as mock_get_version:
            with patch("moai_adk.cli.commands.update.__version__", "0.3.0"):
                with patch("moai_adk.cli.commands.update.detect_install_method") as mock_detect:
                    with patch("moai_adk.cli.commands.update.upgrade_package") as mock_upgrade:
                        mock_get_version.return_value = "0.4.0"
                        mock_detect.return_value = "uv-pip"
                        mock_upgrade.return_value = False

                        result = runner.invoke(update)
                        assert result.exit_code == 1
                        assert "Upgrade failed" in result.output
                        assert "uv pip install --upgrade moai-adk" in result.output

    def test_update_failed_upgrade_pip(self):
        """Test failed upgrade via pip (shows manual instructions)"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.update.get_latest_version") as mock_get_version:
            with patch("moai_adk.cli.commands.update.__version__", "0.3.0"):
                with patch("moai_adk.cli.commands.update.detect_install_method") as mock_detect:
                    with patch("moai_adk.cli.commands.update.upgrade_package") as mock_upgrade:
                        mock_get_version.return_value = "0.4.0"
                        mock_detect.return_value = "pip"
                        mock_upgrade.return_value = False

                        result = runner.invoke(update)
                        assert result.exit_code == 1
                        assert "Upgrade failed" in result.output
                        assert "pip install --upgrade moai-adk" in result.output

    def test_update_pypi_failure_returns_early(self):
        """Test update returns early when PyPI fetch fails"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.update.get_latest_version") as mock_get_version:
            with patch("moai_adk.cli.commands.update.detect_install_method") as mock_detect:
                mock_get_version.return_value = None

                result = runner.invoke(update)
                assert result.exit_code == 0
                assert "Unable to fetch from PyPI" in result.output
                # Should NOT call detect_install_method
                mock_detect.assert_not_called()

    def test_update_exception_handling(self):
        """Test update handles unexpected exceptions"""
        runner = CliRunner()

        with patch("moai_adk.cli.commands.update.get_latest_version", side_effect=RuntimeError("Test error")):
            result = runner.invoke(update)
            assert result.exit_code != 0
            assert "Update failed" in result.output
