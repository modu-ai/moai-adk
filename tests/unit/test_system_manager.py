"""
Unit tests for the system manager module.

Tests the SystemManager class and its system validation methods
to ensure proper environment detection and validation.
"""

import pytest
import json
import tempfile
import subprocess
from pathlib import Path
from unittest.mock import patch, MagicMock, mock_open

from moai_adk.core.system_manager import SystemManager


class TestSystemManager:
    """Test cases for SystemManager class."""

    @pytest.fixture
    def system_manager(self):
        """Create a SystemManager instance for testing."""
        return SystemManager()

    @pytest.fixture
    def temp_dir(self):
        """Create a temporary directory for testing."""
        with tempfile.TemporaryDirectory() as temp_dir:
            yield Path(temp_dir)

    def test_init(self, system_manager):
        """Test SystemManager initialization."""
        assert isinstance(system_manager, SystemManager)

    @patch("subprocess.run")
    def test_check_command_exists_success(self, mock_run, system_manager):
        """Test checking command existence when command exists."""
        mock_run.return_value = MagicMock(returncode=0)

        result = system_manager._check_command_exists("node")

        assert result is True
        mock_run.assert_called_once_with(
            ["node", "--version"], capture_output=True, text=True, check=True
        )

    @patch("subprocess.run")
    def test_check_command_exists_not_found(self, mock_run, system_manager):
        """Test checking command existence when command doesn't exist."""
        mock_run.side_effect = FileNotFoundError()

        result = system_manager._check_command_exists("nonexistent")

        assert result is False

    @patch("subprocess.run")
    def test_check_command_exists_command_failed(self, mock_run, system_manager):
        """Test checking command existence when command fails."""
        mock_run.side_effect = subprocess.CalledProcessError(1, ["node", "--version"])

        result = system_manager._check_command_exists("node")

        assert result is False

    @patch.object(SystemManager, "_check_command_exists")
    @patch("builtins.print")
    def test_check_nodejs_and_npm_no_node(
        self, mock_print, mock_check_cmd, system_manager
    ):
        """Test Node.js check when Node.js is not installed."""

        def mock_check_side_effect(cmd):
            return cmd != "node"

        mock_check_cmd.side_effect = mock_check_side_effect

        result = system_manager.check_nodejs_and_npm()

        assert result is False

    @patch.object(SystemManager, "_check_command_exists")
    @patch("builtins.print")
    def test_check_nodejs_and_npm_no_npm(
        self, mock_print, mock_check_cmd, system_manager
    ):
        """Test Node.js check when npm is not installed."""

        def mock_check_side_effect(cmd):
            return cmd == "node"

        mock_check_cmd.side_effect = mock_check_side_effect

        result = system_manager.check_nodejs_and_npm()

        assert result is False

    @patch.object(SystemManager, "_check_command_exists")
    @patch.object(SystemManager, "_validate_nodejs_environment")
    @patch("builtins.print")
    def test_check_nodejs_and_npm_success(
        self, mock_print, mock_validate, mock_check_cmd, system_manager
    ):
        """Test successful Node.js and npm check."""
        mock_check_cmd.return_value = True
        mock_validate.return_value = True

        result = system_manager.check_nodejs_and_npm()

        assert result is True
        mock_validate.assert_called_once()

    @patch("subprocess.run")
    @patch.object(SystemManager, "_test_ccusage_availability")
    @patch("builtins.print")
    def test_validate_nodejs_environment_success(
        self, mock_print, mock_ccusage, mock_run, system_manager
    ):
        """Test successful Node.js environment validation."""
        # Mock subprocess results
        mock_run.side_effect = [
            MagicMock(stdout="v18.17.0\n", returncode=0),  # node version
            MagicMock(stdout="9.6.7\n", returncode=0),  # npm version
        ]
        mock_ccusage.return_value = True

        result = system_manager._validate_nodejs_environment()

        assert result is True
        assert mock_run.call_count == 2

    @patch("subprocess.run")
    @patch("builtins.print")
    def test_validate_nodejs_environment_command_failed(
        self, mock_print, mock_run, system_manager
    ):
        """Test Node.js environment validation when command fails."""
        mock_run.side_effect = subprocess.CalledProcessError(1, ["node", "--version"])

        result = system_manager._validate_nodejs_environment()

        assert result is False

    @patch("subprocess.run")
    @patch("builtins.print")
    def test_validate_nodejs_environment_exception(
        self, mock_print, mock_run, system_manager
    ):
        """Test Node.js environment validation with unexpected exception."""
        mock_run.side_effect = Exception("Unexpected error")

        result = system_manager._validate_nodejs_environment()

        assert result is False

    @patch("subprocess.run")
    @patch("builtins.print")
    def test_test_ccusage_availability_success(
        self, mock_print, mock_run, system_manager
    ):
        """Test successful ccusage availability test."""
        mock_run.return_value = MagicMock(
            returncode=0, stdout="ccusage help information"
        )

        result = system_manager._test_ccusage_availability()

        assert result is True
        mock_run.assert_called_once_with(
            ["npx", "-y", "ccusage", "--help"],
            capture_output=True,
            text=True,
            timeout=30,
            check=False,
        )

    @patch("subprocess.run")
    @patch("builtins.print")
    def test_test_ccusage_availability_ccusage_in_output(
        self, mock_print, mock_run, system_manager
    ):
        """Test ccusage availability when ccusage appears in output despite non-zero exit."""
        mock_run.return_value = MagicMock(
            returncode=1, stdout="Error: ccusage requires parameters"
        )

        result = system_manager._test_ccusage_availability()

        assert result is True

    @patch("subprocess.run")
    @patch("builtins.print")
    def test_test_ccusage_availability_failed(
        self, mock_print, mock_run, system_manager
    ):
        """Test ccusage availability test failure."""
        mock_run.return_value = MagicMock(returncode=1, stdout="Command not found")

        result = system_manager._test_ccusage_availability()

        assert result is False

    @patch("subprocess.run")
    @patch("builtins.print")
    def test_test_ccusage_availability_timeout(
        self, mock_print, mock_run, system_manager
    ):
        """Test ccusage availability test timeout."""
        mock_run.side_effect = subprocess.TimeoutExpired("npx", 30)

        result = system_manager._test_ccusage_availability()

        assert result is False

    @patch("subprocess.run")
    @patch("builtins.print")
    def test_test_ccusage_availability_exception(
        self, mock_print, mock_run, system_manager
    ):
        """Test ccusage availability test with exception."""
        mock_run.side_effect = Exception("Network error")

        result = system_manager._test_ccusage_availability()

        assert result is False

    @patch("platform.system")
    @patch("platform.release")
    @patch("platform.machine")
    @patch("platform.processor")
    @patch.object(SystemManager, "_get_nodejs_info")
    @patch.object(SystemManager, "_check_command_exists")
    @patch.object(SystemManager, "_get_package_managers_info")
    def test_get_system_info_complete(
        self,
        mock_pkg_mgrs,
        mock_check_git,
        mock_nodejs_info,
        mock_processor,
        mock_machine,
        mock_release,
        mock_system,
        system_manager,
    ):
        """Test getting complete system information."""
        # Mock platform information
        mock_system.return_value = "Darwin"
        mock_release.return_value = "21.6.0"
        mock_machine.return_value = "arm64"
        mock_processor.return_value = "arm"

        # Mock other components
        mock_check_git.return_value = True
        mock_nodejs_info.return_value = {
            "node_available": True,
            "node_version": "v18.17.0",
        }
        mock_pkg_mgrs.return_value = {"pip": True, "brew": True}

        result = system_manager.get_system_info()

        # Check structure
        assert "platform" in result
        assert "python" in result
        assert "nodejs" in result
        assert "git" in result
        assert "package_managers" in result

        # Check platform info
        assert result["platform"]["system"] == "Darwin"
        assert result["platform"]["machine"] == "arm64"

        # Check Python info
        assert "version" in result["python"]
        assert "version_info" in result["python"]

        # Check Git info
        assert result["git"]["available"] is True

    @patch.object(SystemManager, "_check_command_exists")
    @patch("subprocess.run")
    def test_get_nodejs_info_all_available(
        self, mock_run, mock_check_cmd, system_manager
    ):
        """Test getting Node.js info when all tools are available."""

        # Mock command existence checks
        def mock_check_side_effect(cmd):
            return cmd in ["node", "npm", "yarn", "pnpm"]

        mock_check_cmd.side_effect = mock_check_side_effect

        # Mock version queries
        mock_run.side_effect = [
            MagicMock(stdout="v18.17.0\n", returncode=0),  # node version
            MagicMock(stdout="9.6.7\n", returncode=0),  # npm version
        ]

        with patch.object(system_manager, "_quick_ccusage_test", return_value=True):
            result = system_manager._get_nodejs_info()

        assert result["node_available"] is True
        assert result["npm_available"] is True
        assert result["yarn_available"] is True
        assert result["pnpm_available"] is True
        assert result["node_version"] == "v18.17.0"
        assert result["npm_version"] == "9.6.7"
        assert result["ccusage_available"] is True

    @patch.object(SystemManager, "_check_command_exists")
    def test_get_nodejs_info_nothing_available(self, mock_check_cmd, system_manager):
        """Test getting Node.js info when nothing is available."""
        mock_check_cmd.return_value = False

        result = system_manager._get_nodejs_info()

        assert result["node_available"] is False
        assert result["npm_available"] is False
        assert result["yarn_available"] is False
        assert result["pnpm_available"] is False

    @patch.object(SystemManager, "_check_command_exists")
    @patch("subprocess.run")
    def test_get_nodejs_info_version_query_failed(
        self, mock_run, mock_check_cmd, system_manager
    ):
        """Test getting Node.js info when version queries fail."""
        mock_check_cmd.return_value = True
        mock_run.side_effect = Exception("Command failed")

        with patch.object(system_manager, "_quick_ccusage_test", return_value=False):
            result = system_manager._get_nodejs_info()

        assert result["node_available"] is True
        assert result["npm_available"] is True
        assert "node_version" not in result
        assert "npm_version" not in result

    @patch("subprocess.run")
    def test_quick_ccusage_test_success(self, mock_run, system_manager):
        """Test quick ccusage test success."""
        mock_run.return_value = MagicMock(returncode=0, stdout="ccusage help")

        result = system_manager._quick_ccusage_test()

        assert result is True

    @patch("subprocess.run")
    def test_quick_ccusage_test_ccusage_in_output(self, mock_run, system_manager):
        """Test quick ccusage test with ccusage in output."""
        mock_run.return_value = MagicMock(returncode=1, stdout="ccusage error")

        result = system_manager._quick_ccusage_test()

        assert result is True

    @patch("subprocess.run")
    def test_quick_ccusage_test_failed(self, mock_run, system_manager):
        """Test quick ccusage test failure."""
        mock_run.return_value = MagicMock(returncode=1, stdout="command not found")

        result = system_manager._quick_ccusage_test()

        assert result is False

    @patch("subprocess.run")
    def test_quick_ccusage_test_exception(self, mock_run, system_manager):
        """Test quick ccusage test with exception."""
        mock_run.side_effect = Exception("Network error")

        result = system_manager._quick_ccusage_test()

        assert result is False

    @patch.object(SystemManager, "_check_command_exists")
    def test_get_package_managers_info(self, mock_check_cmd, system_manager):
        """Test getting package managers information."""

        def mock_check_side_effect(cmd):
            return cmd in ["pip", "brew", "apt"]

        mock_check_cmd.side_effect = mock_check_side_effect

        result = system_manager._get_package_managers_info()

        expected_managers = [
            "pip",
            "pip3",
            "conda",
            "brew",
            "apt",
            "yum",
            "dnf",
            "choco",
            "winget",
        ]
        for manager in expected_managers:
            assert manager in result

        assert result["pip"] is True
        assert result["brew"] is True
        assert result["apt"] is True
        assert result["yum"] is False

    def test_check_python_version_sufficient(self, system_manager):
        """Test Python version check when version is sufficient."""
        import sys

        current_version = (sys.version_info.major, sys.version_info.minor)

        # Test with version lower than current
        test_version = (current_version[0], max(0, current_version[1] - 1))
        result = system_manager.check_python_version(test_version)

        assert result is True

    def test_check_python_version_insufficient(self, system_manager):
        """Test Python version check when version is insufficient."""
        # Test with impossibly high version
        result = system_manager.check_python_version((99, 99))

        assert result is False

    def test_check_python_version_default(self, system_manager):
        """Test Python version check with default minimum version."""
        result = system_manager.check_python_version()

        # Should be True for any reasonable Python environment running tests
        assert isinstance(result, bool)

    def test_detect_project_type_empty_directory(self, system_manager, temp_dir):
        """Test project type detection in empty directory."""
        result = system_manager.detect_project_type(temp_dir)

        assert result["type"] == "unknown"
        assert result["language"] == "unknown"
        assert result["frameworks"] == []
        assert result["build_tools"] == []
        assert result["files_found"] == []

    def test_detect_project_type_python_project(self, system_manager, temp_dir):
        """Test project type detection for Python project."""
        # Create Python project files
        (temp_dir / "requirements.txt").write_text("django==4.0\n")
        (temp_dir / "pyproject.toml").write_text("[tool.poetry]\n")

        result = system_manager.detect_project_type(temp_dir)

        assert result["type"] == "python"
        assert result["language"] == "python"
        assert "requirements.txt" in result["files_found"]
        assert "pyproject.toml" in result["files_found"]

    def test_detect_project_type_nodejs_project(self, system_manager, temp_dir):
        """Test project type detection for Node.js project."""
        # Create package.json
        package_json_content = {
            "name": "test-project",
            "version": "1.0.0",
            "dependencies": {"react": "^18.0.0", "express": "^4.18.0"},
            "devDependencies": {"typescript": "^4.8.0", "webpack": "^5.74.0"},
            "scripts": {"dev": "webpack serve", "build": "webpack build"},
        }

        (temp_dir / "package.json").write_text(json.dumps(package_json_content))

        result = system_manager.detect_project_type(temp_dir)

        assert result["type"] == "nodejs"
        assert result["language"] == "javascript"
        assert "package.json" in result["files_found"]
        assert "react" in result["frameworks"]
        assert "express" in result["frameworks"]
        assert "typescript" in result["build_tools"]
        assert "webpack" in result["build_tools"]

    def test_detect_project_type_rust_project(self, system_manager, temp_dir):
        """Test project type detection for Rust project."""
        (temp_dir / "Cargo.toml").write_text("[package]\nname = 'test'\n")

        result = system_manager.detect_project_type(temp_dir)

        assert result["type"] == "rust"
        assert result["language"] == "rust"
        assert "Cargo.toml" in result["files_found"]

    def test_detect_project_type_multiple_files(self, system_manager, temp_dir):
        """Test project type detection with multiple project files."""
        # Create multiple project files - last one should win
        (temp_dir / "requirements.txt").write_text("django==4.0\n")
        (temp_dir / "Cargo.toml").write_text("[package]\nname = 'test'\n")

        result = system_manager.detect_project_type(temp_dir)

        # Should detect Rust (last one processed)
        assert result["type"] == "rust"
        assert result["language"] == "rust"
        assert len(result["files_found"]) == 2

    def test_analyze_package_json_success(self, system_manager, temp_dir):
        """Test successful package.json analysis."""
        package_json_content = {
            "name": "test-project",
            "dependencies": {"next": "^13.0.0", "react": "^18.0.0"},
            "devDependencies": {"vite": "^4.0.0", "typescript": "^4.8.0"},
            "scripts": {"dev": "next dev", "build": "next build"},
        }

        package_json_path = temp_dir / "package.json"
        with open(package_json_path, "w", encoding="utf-8") as f:
            json.dump(package_json_content, f)

        result = system_manager._analyze_package_json(package_json_path)

        assert "nextjs" in result["frameworks"]
        assert "react" in result["frameworks"]
        assert "vite" in result["build_tools"]
        assert "typescript" in result["build_tools"]
        assert result["has_scripts"] is True
        assert "dev" in result["scripts"]
        assert "build" in result["scripts"]

    def test_analyze_package_json_file_not_found(self, system_manager, temp_dir):
        """Test package.json analysis when file doesn't exist."""
        nonexistent_path = temp_dir / "nonexistent.json"

        result = system_manager._analyze_package_json(nonexistent_path)

        assert result["frameworks"] == []
        assert result["build_tools"] == []

    def test_analyze_package_json_invalid_json(self, system_manager, temp_dir):
        """Test package.json analysis with invalid JSON."""
        package_json_path = temp_dir / "package.json"
        package_json_path.write_text('{"invalid": json}')

        result = system_manager._analyze_package_json(package_json_path)

        assert result["frameworks"] == []
        assert result["build_tools"] == []

    def test_analyze_package_json_minimal(self, system_manager, temp_dir):
        """Test package.json analysis with minimal content."""
        package_json_content = {"name": "minimal-project"}

        package_json_path = temp_dir / "package.json"
        with open(package_json_path, "w", encoding="utf-8") as f:
            json.dump(package_json_content, f)

        result = system_manager._analyze_package_json(package_json_path)

        assert result["frameworks"] == []
        assert result["build_tools"] == []
        assert result["has_scripts"] is False
        assert result["scripts"] == []

    def test_should_create_package_json_node_runtime(self, system_manager):
        """Test package.json creation decision for Node.js runtime."""
        from moai_adk.config import Config, RuntimeConfig

        config = Config(name="test", template="standard", runtime=RuntimeConfig("node"))

        result = system_manager.should_create_package_json(config)
        assert result is True

    def test_should_create_package_json_tsx_runtime(self, system_manager):
        """Test package.json creation decision for TSX runtime."""
        from moai_adk.config import Config, RuntimeConfig

        config = Config(name="test", template="standard", runtime=RuntimeConfig("tsx"))

        result = system_manager.should_create_package_json(config)
        assert result is True

    def test_should_create_package_json_web_framework(self, system_manager):
        """Test package.json creation decision for web frameworks."""
        from moai_adk.config import Config, RuntimeConfig

        config = Config(
            name="test",
            template="standard",
            runtime=RuntimeConfig("python"),
            tech_stack=["nextjs", "react"],
        )

        result = system_manager.should_create_package_json(config)
        assert result is True

    def test_should_create_package_json_python_only(self, system_manager):
        """Test package.json creation decision for Python-only project."""
        from moai_adk.config import Config, RuntimeConfig

        config = Config(
            name="test",
            template="standard",
            runtime=RuntimeConfig("python"),
            tech_stack=["django", "python"],
        )

        result = system_manager.should_create_package_json(config)
        assert result is False

    def test_integration_full_system_check(self, system_manager):
        """Test complete system check workflow."""
        with patch.object(system_manager, "_check_command_exists") as mock_check:
            with patch("subprocess.run") as mock_run:
                # Mock system commands
                mock_check.return_value = True

                # Mock Node.js and npm version outputs
                mock_run.side_effect = [
                    MagicMock(stdout="v18.17.0\n", returncode=0),  # node version
                    MagicMock(stdout="9.6.7\n", returncode=0),  # npm version
                    MagicMock(stdout="ccusage help", returncode=0),  # ccusage test
                ]

                # Test Node.js check
                nodejs_result = system_manager.check_nodejs_and_npm()
                assert nodejs_result is True

                # Test system info
                system_info = system_manager.get_system_info()
                assert "platform" in system_info
                assert "python" in system_info
                assert "nodejs" in system_info

    def test_framework_detection_comprehensive(self, system_manager, temp_dir):
        """Test comprehensive framework detection."""
        # Create package.json with multiple frameworks
        package_json_content = {
            "name": "multi-framework-project",
            "dependencies": {
                "react": "^18.0.0",
                "vue": "^3.0.0",
                "express": "^4.18.0",
                "next": "^13.0.0",
            },
            "devDependencies": {
                "@angular/core": "^15.0.0",
                "svelte": "^3.0.0",
                "webpack": "^5.0.0",
                "vite": "^4.0.0",
                "rollup": "^3.0.0",
            },
        }

        package_json_path = temp_dir / "package.json"
        with open(package_json_path, "w", encoding="utf-8") as f:
            json.dump(package_json_content, f)

        result = system_manager._analyze_package_json(package_json_path)

        # Should detect all frameworks
        expected_frameworks = ["react", "vue", "angular", "svelte", "nextjs", "express"]
        for framework in expected_frameworks:
            assert framework in result["frameworks"]

        # Should detect all build tools
        expected_tools = ["webpack", "vite", "rollup"]
        for tool in expected_tools:
            assert tool in result["build_tools"]
