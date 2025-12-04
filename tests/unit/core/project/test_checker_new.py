"""
Comprehensive tests for SystemChecker module.

Tests cover:
- SystemChecker class initialization
- check_all method
- check_language_tools method
- get_tool_version method
- get_platform_specific_message helper
- get_permission_fix_message helper
- check_environment helper
"""

import platform
import shutil
import subprocess
import tempfile
from pathlib import Path
from unittest import mock
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.core.project.checker import (
    SystemChecker,
    check_environment,
    get_permission_fix_message,
    get_platform_specific_message,
)


class TestSystemChecker:
    """Test suite for SystemChecker class."""

    def test_initialization(self):
        """Test SystemChecker initialization."""
        # Arrange & Act
        checker = SystemChecker()

        # Assert
        assert checker is not None
        assert hasattr(checker, "REQUIRED_TOOLS")
        assert hasattr(checker, "OPTIONAL_TOOLS")
        assert hasattr(checker, "LANGUAGE_TOOLS")

    def test_required_tools_contain_git_and_python(self):
        """Test REQUIRED_TOOLS contains git and python."""
        # Arrange
        checker = SystemChecker()

        # Act
        required = checker.REQUIRED_TOOLS

        # Assert
        assert "git" in required
        assert "python" in required

    def test_optional_tools_contain_gh_and_docker(self):
        """Test OPTIONAL_TOOLS contains gh and docker."""
        # Arrange
        checker = SystemChecker()

        # Act
        optional = checker.OPTIONAL_TOOLS

        # Assert
        assert "gh" in optional
        assert "docker" in optional

    def test_language_tools_contains_major_languages(self):
        """Test LANGUAGE_TOOLS contains major programming languages."""
        # Arrange
        checker = SystemChecker()

        # Act
        languages = checker.LANGUAGE_TOOLS

        # Assert
        assert "python" in languages
        assert "typescript" in languages
        assert "javascript" in languages
        assert "go" in languages
        assert "rust" in languages
        assert "java" in languages

    def test_check_all_returns_dict(self):
        """Test check_all returns dictionary of tool availability."""
        # Arrange
        checker = SystemChecker()

        # Act
        result = checker.check_all()

        # Assert
        assert isinstance(result, dict)
        assert "git" in result
        assert "python" in result
        assert isinstance(result["git"], bool)

    def test_check_all_includes_required_and_optional(self):
        """Test check_all includes both required and optional tools."""
        # Arrange
        checker = SystemChecker()

        # Act
        result = checker.check_all()

        # Assert
        assert len(result) >= 4  # At least git, python, gh, docker

    def test_check_tool_returns_bool(self):
        """Test _check_tool returns boolean."""
        # Arrange
        checker = SystemChecker()

        # Act
        result = checker._check_tool("git --version")

        # Assert
        assert isinstance(result, bool)

    def test_check_tool_returns_true_for_existing_tool(self):
        """Test _check_tool returns True for git (usually available)."""
        # Arrange
        checker = SystemChecker()

        # Act
        with patch("shutil.which", return_value="/usr/bin/git"):
            result = checker._check_tool("git --version")

        # Assert
        assert result is True

    def test_check_tool_returns_false_for_nonexistent_tool(self):
        """Test _check_tool returns False for nonexistent tool."""
        # Arrange
        checker = SystemChecker()

        # Act
        with patch("shutil.which", return_value=None):
            result = checker._check_tool("nonexistent_tool --version")

        # Assert
        assert result is False

    def test_check_tool_handles_empty_command(self):
        """Test _check_tool handles empty command gracefully."""
        # Arrange
        checker = SystemChecker()

        # Act
        result = checker._check_tool("")

        # Assert
        assert result is False

    def test_check_language_tools_python_has_required_tools(self):
        """Test check_language_tools returns required tools for Python."""
        # Arrange
        checker = SystemChecker()

        # Act
        with patch.object(checker, "_is_tool_available", return_value=True):
            result = checker.check_language_tools("python")

        # Assert
        assert "python3" in result
        assert "pip" in result

    def test_check_language_tools_typescript_has_required_tools(self):
        """Test check_language_tools returns required tools for TypeScript."""
        # Arrange
        checker = SystemChecker()

        # Act
        with patch.object(checker, "_is_tool_available", return_value=True):
            result = checker.check_language_tools("typescript")

        # Assert
        assert "node" in result
        assert "npm" in result

    def test_check_language_tools_go_has_required_tools(self):
        """Test check_language_tools returns required tools for Go."""
        # Arrange
        checker = SystemChecker()

        # Act
        with patch.object(checker, "_is_tool_available", return_value=True):
            result = checker.check_language_tools("go")

        # Assert
        assert "go" in result

    def test_check_language_tools_rust_has_required_tools(self):
        """Test check_language_tools returns required tools for Rust."""
        # Arrange
        checker = SystemChecker()

        # Act
        with patch.object(checker, "_is_tool_available", return_value=True):
            result = checker.check_language_tools("rust")

        # Assert
        assert "rustc" in result
        assert "cargo" in result

    def test_check_language_tools_java_has_required_tools(self):
        """Test check_language_tools returns required tools for Java."""
        # Arrange
        checker = SystemChecker()

        # Act
        with patch.object(checker, "_is_tool_available", return_value=True):
            result = checker.check_language_tools("java")

        # Assert
        assert "java" in result
        assert "javac" in result

    def test_check_language_tools_returns_dict(self):
        """Test check_language_tools returns dictionary."""
        # Arrange
        checker = SystemChecker()

        # Act
        result = checker.check_language_tools("python")

        # Assert
        assert isinstance(result, dict)

    def test_check_language_tools_returns_empty_for_none(self):
        """Test check_language_tools returns empty dict for None language."""
        # Arrange
        checker = SystemChecker()

        # Act
        result = checker.check_language_tools(None)

        # Assert
        assert result == {}

    def test_check_language_tools_returns_empty_for_unknown_language(self):
        """Test check_language_tools returns empty dict for unknown language."""
        # Arrange
        checker = SystemChecker()

        # Act
        result = checker.check_language_tools("unknown_language")

        # Assert
        assert result == {}

    def test_check_language_tools_is_case_insensitive(self):
        """Test check_language_tools is case insensitive."""
        # Arrange
        checker = SystemChecker()

        # Act
        with patch.object(checker, "_is_tool_available", return_value=True):
            result = checker.check_language_tools("PYTHON")

        # Assert
        assert "python3" in result

    def test_is_tool_available_returns_bool(self):
        """Test _is_tool_available returns boolean."""
        # Arrange
        checker = SystemChecker()

        # Act
        with patch("shutil.which", return_value="/usr/bin/python3"):
            result = checker._is_tool_available("python3")

        # Assert
        assert isinstance(result, bool)

    def test_is_tool_available_checks_shutil_which(self):
        """Test _is_tool_available uses shutil.which."""
        # Arrange
        checker = SystemChecker()

        # Act
        with patch("shutil.which") as mock_which:
            mock_which.return_value = "/usr/bin/python3"
            checker._is_tool_available("python3")

        # Assert
        mock_which.assert_called_once_with("python3")

    def test_get_tool_version_returns_string(self):
        """Test get_tool_version returns string version."""
        # Arrange
        checker = SystemChecker()

        # Act
        with patch("shutil.which", return_value="/usr/bin/python3"):
            with patch("subprocess.run") as mock_run:
                mock_run.return_value = MagicMock(returncode=0, stdout="Python 3.11.0")
                result = checker.get_tool_version("python3")

        # Assert
        assert isinstance(result, str)
        assert "Python" in result

    def test_get_tool_version_returns_none_for_nonexistent(self):
        """Test get_tool_version returns None for nonexistent tool."""
        # Arrange
        checker = SystemChecker()

        # Act
        with patch("shutil.which", return_value=None):
            result = checker.get_tool_version("nonexistent")

        # Assert
        assert result is None

    def test_get_tool_version_returns_none_for_none_input(self):
        """Test get_tool_version returns None for None input."""
        # Arrange
        checker = SystemChecker()

        # Act
        result = checker.get_tool_version(None)

        # Assert
        assert result is None

    def test_get_tool_version_handles_subprocess_timeout(self):
        """Test get_tool_version handles subprocess timeout gracefully."""
        # Arrange
        checker = SystemChecker()

        # Act
        with patch("shutil.which", return_value="/usr/bin/tool"):
            with patch("subprocess.run") as mock_run:
                mock_run.side_effect = subprocess.TimeoutExpired("tool", 2)
                result = checker.get_tool_version("tool")

        # Assert
        assert result is None

    def test_get_tool_version_handles_oserror(self):
        """Test get_tool_version handles OSError gracefully."""
        # Arrange
        checker = SystemChecker()

        # Act
        with patch("shutil.which", return_value="/usr/bin/tool"):
            with patch("subprocess.run") as mock_run:
                mock_run.side_effect = OSError("Permission denied")
                result = checker.get_tool_version("tool")

        # Assert
        assert result is None

    def test_extract_version_line_extracts_first_line(self):
        """Test _extract_version_line extracts first line."""
        # Arrange
        checker = SystemChecker()
        output = "Python 3.11.0\ndetail line 2\ndetail line 3"

        # Act
        result = checker._extract_version_line(output)

        # Assert
        assert result == "Python 3.11.0"

    def test_extract_version_line_strips_whitespace(self):
        """Test _extract_version_line strips whitespace."""
        # Arrange
        checker = SystemChecker()
        output = "  Python 3.11.0  \n"

        # Act
        result = checker._extract_version_line(output)

        # Assert
        assert result == "Python 3.11.0"

    def test_extract_version_line_handles_empty_output(self):
        """Test _extract_version_line handles empty output."""
        # Arrange
        checker = SystemChecker()
        output = ""

        # Act
        result = checker._extract_version_line(output)

        # Assert
        assert result == ""


class TestHelperFunctions:
    """Test suite for module-level helper functions."""

    def test_check_environment_returns_dict(self):
        """Test check_environment returns dictionary."""
        # Arrange & Act
        result = check_environment()

        # Assert
        assert isinstance(result, dict)

    def test_check_environment_checks_python_version(self):
        """Test check_environment includes Python version check."""
        # Arrange & Act
        result = check_environment()

        # Assert
        assert "Python >= 3.11" in result
        assert isinstance(result["Python >= 3.11"], bool)

    def test_check_environment_checks_git_installed(self):
        """Test check_environment includes git check."""
        # Arrange & Act
        result = check_environment()

        # Assert
        assert "Git installed" in result
        assert isinstance(result["Git installed"], bool)

    def test_check_environment_checks_moai_directory(self):
        """Test check_environment includes .moai/ directory check."""
        # Arrange & Act
        result = check_environment()

        # Assert
        assert "Project structure (.moai/)" in result
        assert isinstance(result["Project structure (.moai/)"], bool)

    def test_check_environment_checks_config_file(self):
        """Test check_environment includes config.json check."""
        # Arrange & Act
        result = check_environment()

        # Assert
        assert "Config file (.moai/config/config.json)" in result
        assert isinstance(result["Config file (.moai/config/config.json)"], bool)

    def test_get_platform_specific_message_returns_unix_on_unix(self):
        """Test get_platform_specific_message returns Unix message on Unix."""
        # Arrange
        unix_msg = "chmod 755 file"
        windows_msg = "Check permissions"

        # Act
        with patch("platform.system", return_value="Linux"):
            result = get_platform_specific_message(unix_msg, windows_msg)

        # Assert
        assert result == unix_msg

    def test_get_platform_specific_message_returns_windows_on_windows(self):
        """Test get_platform_specific_message returns Windows message on Windows."""
        # Arrange
        unix_msg = "chmod 755 file"
        windows_msg = "Check permissions"

        # Act
        with patch("platform.system", return_value="Windows"):
            result = get_platform_specific_message(unix_msg, windows_msg)

        # Assert
        assert result == windows_msg

    def test_get_platform_specific_message_returns_unix_on_macos(self):
        """Test get_platform_specific_message returns Unix message on macOS."""
        # Arrange
        unix_msg = "chmod 755 file"
        windows_msg = "Check permissions"

        # Act
        with patch("platform.system", return_value="Darwin"):
            result = get_platform_specific_message(unix_msg, windows_msg)

        # Assert
        assert result == unix_msg

    def test_get_permission_fix_message_returns_unix_on_unix(self):
        """Test get_permission_fix_message returns Unix message on Unix."""
        # Arrange
        path = ".moai"

        # Act
        with patch("platform.system", return_value="Linux"):
            result = get_permission_fix_message(path)

        # Assert
        assert "chmod" in result
        assert path in result

    def test_get_permission_fix_message_returns_windows_on_windows(self):
        """Test get_permission_fix_message returns Windows message on Windows."""
        # Arrange
        path = ".moai"

        # Act
        with patch("platform.system", return_value="Windows"):
            result = get_permission_fix_message(path)

        # Assert
        assert "administrator" in result.lower() or "properties" in result.lower()

    def test_get_permission_fix_message_includes_path(self):
        """Test get_permission_fix_message includes the provided path."""
        # Arrange
        path = "test_path"

        # Act
        result = get_permission_fix_message(path)

        # Assert
        assert path in result


class TestLanguageToolsIntegration:
    """Test suite for language-specific tool checking integration."""

    def test_python_tools_categories(self):
        """Test Python tools include required, recommended, optional."""
        # Arrange
        checker = SystemChecker()

        # Act
        tools = checker.LANGUAGE_TOOLS["python"]

        # Assert
        assert "required" in tools
        assert "recommended" in tools
        assert "optional" in tools

    def test_typescript_tools_categories(self):
        """Test TypeScript tools include required, recommended, optional."""
        # Arrange
        checker = SystemChecker()

        # Act
        tools = checker.LANGUAGE_TOOLS["typescript"]

        # Assert
        assert "required" in tools
        assert "recommended" in tools
        assert "optional" in tools

    def test_go_tools_categories(self):
        """Test Go tools include required, recommended, optional."""
        # Arrange
        checker = SystemChecker()

        # Act
        tools = checker.LANGUAGE_TOOLS["go"]

        # Assert
        assert "required" in tools
        assert "recommended" in tools
        assert "optional" in tools

    def test_rust_tools_categories(self):
        """Test Rust tools include required, recommended, optional."""
        # Arrange
        checker = SystemChecker()

        # Act
        tools = checker.LANGUAGE_TOOLS["rust"]

        # Assert
        assert "required" in tools
        assert "recommended" in tools
        assert "optional" in tools

    def test_java_tools_categories(self):
        """Test Java tools include required, recommended, optional."""
        # Arrange
        checker = SystemChecker()

        # Act
        tools = checker.LANGUAGE_TOOLS["java"]

        # Assert
        assert "required" in tools
        assert "recommended" in tools
        assert "optional" in tools

    def test_all_language_tools_have_required(self):
        """Test all languages have at least required tools."""
        # Arrange
        checker = SystemChecker()

        # Act & Assert
        for language, tools_config in checker.LANGUAGE_TOOLS.items():
            assert "required" in tools_config, f"{language} missing required tools"
            assert (
                len(tools_config["required"]) > 0
            ), f"{language} has no required tools"
