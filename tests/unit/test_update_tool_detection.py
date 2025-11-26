"""Unit tests for update command tool detection

Tests the _detect_tool_installer() function which detects
which package manager installed moai-adk (uv tool, pipx, or pip).

RED Phase: These tests should initially fail.
"""

from unittest.mock import MagicMock, patch

from moai_adk.cli.commands.update import _detect_tool_installer


class TestToolDetection:
    """Test tool detection for update command"""

    def test_detect_uv_tool_installed(self):
        """Test detection of uv tool installation."""
        # Mock subprocess.run to simulate uv tool list output
        with patch("subprocess.run") as mock_run:
            # Setup mock for uv tool list (success)
            mock_run.return_value = MagicMock(returncode=0, stdout="moai-adk 0.6.1")

            result = _detect_tool_installer()

            # Assertions
            assert result == ["uv", "tool", "upgrade", "moai-adk"]
            mock_run.assert_called_once()

    def test_detect_pipx_installed(self):
        """Test detection of pipx installation."""
        with patch("subprocess.run") as mock_run:
            # First call: uv tool list fails
            # Second call: pipx list succeeds
            mock_run.side_effect = [
                MagicMock(returncode=1, stdout=""),  # uv fails
                MagicMock(returncode=0, stdout="moai-adk 0.6.1"),  # pipx succeeds
            ]

            result = _detect_tool_installer()

            # Should return pipx upgrade command
            assert result == ["pipx", "upgrade", "moai-adk"]
            assert mock_run.call_count == 2

    def test_detect_pip_fallback(self):
        """Test detection falls back to pip."""
        with patch("subprocess.run") as mock_run:
            # First call: uv tool list fails
            # Second call: pipx list fails
            # Third call: pip show succeeds
            mock_run.side_effect = [
                MagicMock(returncode=1, stdout=""),  # uv fails
                MagicMock(returncode=1, stdout=""),  # pipx fails
                MagicMock(returncode=0, stdout="Name: moai-adk"),  # pip succeeds
            ]

            result = _detect_tool_installer()

            # Should return pip install command
            assert result == ["pip", "install", "--upgrade", "moai-adk"]
            assert mock_run.call_count == 3

    def test_detect_no_tools_available(self):
        """Test when no package manager detects moai-adk."""
        with patch("subprocess.run") as mock_run:
            # All calls fail
            mock_run.side_effect = [
                MagicMock(returncode=1, stdout=""),  # uv fails
                MagicMock(returncode=1, stdout=""),  # pipx fails
                MagicMock(returncode=1, stdout=""),  # pip fails
            ]

            result = _detect_tool_installer()

            # Should return None when all fail
            assert result is None
            assert mock_run.call_count == 3

    def test_detect_prioritizes_uv_over_pipx(self):
        """Test that uv tool is prioritized over pipx when both available."""
        with patch("subprocess.run") as mock_run:
            # Both uv and pipx return success
            mock_run.side_effect = [
                MagicMock(returncode=0, stdout="moai-adk 0.6.1"),  # uv succeeds
            ]

            result = _detect_tool_installer()

            # Should return uv command (not pipx)
            assert result == ["uv", "tool", "upgrade", "moai-adk"]
            # Should only call uv, not pipx (priority check)
            assert mock_run.call_count == 1

    def test_detect_handles_timeout(self):
        """Test graceful handling of subprocess timeout."""
        import subprocess

        with patch("subprocess.run") as mock_run:
            # First call: timeout
            # Second call: success with pipx
            mock_run.side_effect = [
                subprocess.TimeoutExpired("uv", 5),  # timeout
                MagicMock(returncode=0, stdout="moai-adk 0.6.1"),  # pipx succeeds
            ]

            result = _detect_tool_installer()

            # Should continue to next tool after timeout
            assert result == ["pipx", "upgrade", "moai-adk"]
            assert mock_run.call_count == 2

    def test_detect_handles_file_not_found(self):
        """Test graceful handling when tool binary not found."""
        with patch("subprocess.run") as mock_run:
            # First call: FileNotFoundError (tool not installed)
            # Second call: success with pipx
            mock_run.side_effect = [
                FileNotFoundError("uv not found"),  # uv not installed
                MagicMock(returncode=0, stdout="moai-adk 0.6.1"),  # pipx succeeds
            ]

            result = _detect_tool_installer()

            # Should continue to next tool after FileNotFoundError
            assert result == ["pipx", "upgrade", "moai-adk"]
            assert mock_run.call_count == 2

    def test_detect_handles_subprocess_error(self):
        """Test graceful handling of generic subprocess errors."""
        with patch("subprocess.run") as mock_run:
            # First call: OSError
            # Second call: success with pipx
            mock_run.side_effect = [
                OSError("Permission denied"),  # generic OS error
                MagicMock(returncode=0, stdout="moai-adk 0.6.1"),  # pipx succeeds
            ]

            result = _detect_tool_installer()

            # Should continue to next tool after OSError
            assert result == ["pipx", "upgrade", "moai-adk"]
            assert mock_run.call_count == 2
