"""
Unit tests for the git manager module.

Tests the GitManager class and its Git operations methods
to ensure proper Git repository management and security validation.
"""

import pytest
import tempfile
import subprocess
from pathlib import Path
from unittest.mock import patch, MagicMock, call

from moai_adk.core.git_manager import GitManager
from moai_adk.core.security import SecurityManager
from moai_adk.core.file_manager import FileManager


class TestGitManager:
    """Test cases for GitManager class."""

    @pytest.fixture
    def temp_dir(self):
        """Create a temporary directory for testing."""
        with tempfile.TemporaryDirectory() as temp_dir:
            yield Path(temp_dir)

    @pytest.fixture
    def security_manager(self):
        """Create a mock security manager for testing."""
        return MagicMock(spec=SecurityManager)

    @pytest.fixture
    def file_manager(self):
        """Create a mock file manager for testing."""
        return MagicMock(spec=FileManager)

    @pytest.fixture
    def git_manager(self, security_manager, file_manager):
        """Create a GitManager instance for testing."""
        return GitManager(security_manager, file_manager)

    def test_init_with_managers(self, security_manager, file_manager):
        """Test GitManager initialization with managers."""
        manager = GitManager(security_manager, file_manager)
        assert manager.security_manager == security_manager
        assert manager.file_manager == file_manager

    def test_init_without_managers(self):
        """Test GitManager initialization without managers."""
        manager = GitManager()
        assert isinstance(manager.security_manager, SecurityManager)
        assert manager.file_manager is None

    @patch("subprocess.run")
    def test_check_git_available_success(self, mock_run, git_manager):
        """Test checking Git availability when Git is installed."""
        mock_run.return_value = MagicMock(returncode=0)

        result = git_manager._check_git_available()

        assert result is True
        mock_run.assert_called_once_with(
            ["git", "--version"], capture_output=True, text=True, check=True
        )

    @patch("subprocess.run")
    def test_check_git_available_not_found(self, mock_run, git_manager):
        """Test checking Git availability when Git is not installed."""
        mock_run.side_effect = FileNotFoundError()

        result = git_manager._check_git_available()

        assert result is False

    @patch("subprocess.run")
    def test_check_git_available_command_failed(self, mock_run, git_manager):
        """Test checking Git availability when Git command fails."""
        mock_run.side_effect = subprocess.CalledProcessError(1, ["git", "--version"])

        result = git_manager._check_git_available()

        assert result is False

    def test_initialize_git_repository_already_exists(self, git_manager, temp_dir):
        """Test Git initialization when repository already exists."""
        # Create .git directory
        git_dir = temp_dir / ".git"
        git_dir.mkdir()

        success, was_initialized = git_manager.initialize_git_repository(temp_dir)

        assert success is True
        assert was_initialized is False

    @patch.object(GitManager, "_check_git_available")
    @patch.object(GitManager, "_offer_git_installation")
    def test_initialize_git_repository_git_not_available(
        self, mock_offer_install, mock_check_git, git_manager, temp_dir
    ):
        """Test Git initialization when Git is not available."""
        mock_check_git.return_value = False
        mock_offer_install.return_value = False

        success, was_initialized = git_manager.initialize_git_repository(temp_dir)

        assert success is False
        assert was_initialized is False
        mock_offer_install.assert_called_once()

    @patch.object(GitManager, "_check_git_available")
    @patch.object(GitManager, "_initialize_repository")
    def test_initialize_git_repository_success(
        self, mock_init_repo, mock_check_git, git_manager, temp_dir
    ):
        """Test successful Git repository initialization."""
        mock_check_git.return_value = True
        mock_init_repo.return_value = True

        # Mock file manager
        git_manager.file_manager.create_gitignore.return_value = True

        success, was_initialized = git_manager.initialize_git_repository(temp_dir)

        assert success is True
        assert was_initialized is True
        mock_init_repo.assert_called_once_with(temp_dir)

    @patch.object(GitManager, "_check_git_available")
    @patch.object(GitManager, "_initialize_repository")
    def test_initialize_git_repository_initialization_failed(
        self, mock_init_repo, mock_check_git, git_manager, temp_dir
    ):
        """Test Git initialization when repository initialization fails."""
        mock_check_git.return_value = True
        mock_init_repo.return_value = False

        success, was_initialized = git_manager.initialize_git_repository(temp_dir)

        assert success is False
        assert was_initialized is False

    @patch.object(GitManager, "_check_git_available")
    @patch.object(GitManager, "_initialize_repository")
    def test_initialize_git_repository_creates_gitignore(
        self, mock_init_repo, mock_check_git, git_manager, temp_dir
    ):
        """Test that Git initialization creates .gitignore if needed."""
        mock_check_git.return_value = True
        mock_init_repo.return_value = True

        success, was_initialized = git_manager.initialize_git_repository(temp_dir)

        assert success is True
        assert was_initialized is True

        # Should create .gitignore
        expected_gitignore_path = temp_dir / ".gitignore"
        git_manager.file_manager.create_gitignore.assert_called_once_with(
            expected_gitignore_path
        )

    @patch.object(GitManager, "_check_git_available")
    @patch.object(GitManager, "_initialize_repository")
    def test_initialize_git_repository_no_file_manager(
        self, mock_init_repo, mock_check_git, temp_dir
    ):
        """Test Git initialization without file manager."""
        git_manager = GitManager()  # No file manager
        mock_check_git.return_value = True
        mock_init_repo.return_value = True

        success, was_initialized = git_manager.initialize_git_repository(temp_dir)

        assert success is True
        assert was_initialized is True

    @patch("subprocess.run")
    def test_initialize_repository_success(self, mock_run, git_manager, temp_dir):
        """Test successful repository initialization."""
        mock_run.return_value = MagicMock(returncode=0)

        # Mock security validation to return True
        git_manager.security_manager.validate_subprocess_path.return_value = True

        result = git_manager._initialize_repository(temp_dir)

        assert result is True
        mock_run.assert_called_once_with(
            ["git", "init"], cwd=temp_dir, capture_output=True, text=True
        )

    @patch("subprocess.run")
    def test_initialize_repository_security_failure(
        self, mock_run, git_manager, temp_dir
    ):
        """Test repository initialization when security validation fails."""
        # Mock security validation to return False
        git_manager.security_manager.validate_subprocess_path.return_value = False

        result = git_manager._initialize_repository(temp_dir)

        assert result is False
        mock_run.assert_not_called()

    @patch("subprocess.run")
    def test_initialize_repository_git_failed(self, mock_run, git_manager, temp_dir):
        """Test repository initialization when git init fails."""
        mock_run.return_value = MagicMock(returncode=1, stderr="Permission denied")

        # Mock security validation to return True
        git_manager.security_manager.validate_subprocess_path.return_value = True

        result = git_manager._initialize_repository(temp_dir)

        assert result is False

    @patch("subprocess.run")
    def test_initialize_repository_exception(self, mock_run, git_manager, temp_dir):
        """Test repository initialization with exception."""
        mock_run.side_effect = Exception("Unexpected error")

        # Mock security validation to return True
        git_manager.security_manager.validate_subprocess_path.return_value = True

        result = git_manager._initialize_repository(temp_dir)

        assert result is False

    @patch("builtins.input")
    @patch("builtins.print")
    @patch.object(GitManager, "_get_git_install_command")
    def test_offer_git_installation_user_accepts(
        self, mock_get_cmd, mock_print, mock_input, git_manager
    ):
        """Test offering Git installation when user accepts."""
        mock_input.return_value = "y"
        mock_get_cmd.return_value = ["brew", "install", "git"]

        with patch.object(git_manager, "_install_git_with_command", return_value=True):
            result = git_manager._offer_git_installation()

        assert result is True

    @patch("builtins.input")
    @patch("builtins.print")
    def test_offer_git_installation_user_declines(
        self, mock_print, mock_input, git_manager
    ):
        """Test offering Git installation when user declines."""
        mock_input.return_value = "n"

        result = git_manager._offer_git_installation()

        assert result is False

    @patch("builtins.input")
    @patch("builtins.print")
    def test_offer_git_installation_keyboard_interrupt(
        self, mock_print, mock_input, git_manager
    ):
        """Test offering Git installation with keyboard interrupt."""
        mock_input.side_effect = KeyboardInterrupt()

        result = git_manager._offer_git_installation()

        assert result is False

    @patch.object(GitManager, "_check_command_exists")
    def test_get_git_install_command_macos_with_brew(self, mock_check_cmd, git_manager):
        """Test getting Git install command for macOS with Homebrew."""
        mock_check_cmd.return_value = True

        result = git_manager._get_git_install_command("darwin")

        assert result == ["brew", "install", "git"]

    @patch.object(GitManager, "_check_command_exists")
    def test_get_git_install_command_macos_without_brew(
        self, mock_check_cmd, git_manager
    ):
        """Test getting Git install command for macOS without Homebrew."""
        mock_check_cmd.return_value = False

        result = git_manager._get_git_install_command("darwin")

        assert result is None

    @patch.object(GitManager, "_check_command_exists")
    def test_get_git_install_command_linux_apt(self, mock_check_cmd, git_manager):
        """Test getting Git install command for Linux with APT."""

        def mock_check_side_effect(cmd):
            return cmd == "apt"

        mock_check_cmd.side_effect = mock_check_side_effect

        result = git_manager._get_git_install_command("linux")

        expected = [
            "sudo",
            "apt",
            "update",
            "&&",
            "sudo",
            "apt",
            "install",
            "-y",
            "git",
        ]
        assert result == expected

    @patch.object(GitManager, "_check_command_exists")
    def test_get_git_install_command_windows(self, mock_check_cmd, git_manager):
        """Test getting Git install command for Windows."""
        result = git_manager._get_git_install_command("windows")

        assert result is None

    @patch("subprocess.run")
    def test_check_command_exists_success(self, mock_run, git_manager):
        """Test checking command existence when command exists."""
        mock_run.return_value = MagicMock(returncode=0)

        result = git_manager._check_command_exists("brew")

        assert result is True

    @patch("subprocess.run")
    def test_check_command_exists_not_found(self, mock_run, git_manager):
        """Test checking command existence when command doesn't exist."""
        mock_run.side_effect = FileNotFoundError()

        result = git_manager._check_command_exists("nonexistent")

        assert result is False

    @patch("subprocess.run")
    def test_install_git_with_command_success(self, mock_run, git_manager):
        """Test successful Git installation."""
        mock_run.return_value = MagicMock(returncode=0)

        result = git_manager._install_git_with_command(
            ["brew", "install", "git"], "darwin"
        )

        assert result is True

    @patch("subprocess.run")
    def test_install_git_with_command_failed(self, mock_run, git_manager):
        """Test failed Git installation."""
        mock_run.return_value = MagicMock(returncode=1, stderr="Installation failed")

        result = git_manager._install_git_with_command(
            ["brew", "install", "git"], "darwin"
        )

        assert result is False

    @patch("subprocess.run")
    def test_install_git_with_command_linux_complex(self, mock_run, git_manager):
        """Test Git installation on Linux with complex command."""
        mock_run.return_value = MagicMock(returncode=0)

        complex_cmd = [
            "sudo",
            "apt",
            "update",
            "&&",
            "sudo",
            "apt",
            "install",
            "-y",
            "git",
        ]
        result = git_manager._install_git_with_command(complex_cmd, "linux")

        assert result is True
        # Should use shell=True for commands with &&
        mock_run.assert_called_once_with(
            " ".join(complex_cmd),
            shell=True,
            capture_output=True,
            text=True,
            timeout=300,
        )

    @patch("subprocess.run")
    def test_install_git_with_command_timeout(self, mock_run, git_manager):
        """Test Git installation timeout."""
        mock_run.side_effect = subprocess.TimeoutExpired("brew", 300)

        result = git_manager._install_git_with_command(
            ["brew", "install", "git"], "darwin"
        )

        assert result is False

    def test_check_git_status_not_git_repo(self, git_manager, temp_dir):
        """Test Git status check when not a Git repository."""
        result = git_manager.check_git_status(temp_dir)

        assert result["is_git_repo"] is False
        assert "error" in result

    def test_check_git_status_security_failure(self, git_manager, temp_dir):
        """Test Git status check when security validation fails."""
        # Create .git directory
        (temp_dir / ".git").mkdir()

        # Mock security validation to return False
        git_manager.security_manager.validate_subprocess_path.return_value = False

        result = git_manager.check_git_status(temp_dir)

        assert result["is_git_repo"] is True
        assert "error" in result
        assert "Security validation failed" in result["error"]

    @patch("subprocess.run")
    def test_check_git_status_clean_repo(self, mock_run, git_manager, temp_dir):
        """Test Git status check for clean repository."""
        # Create .git directory
        (temp_dir / ".git").mkdir()

        # Mock security validation to return True
        git_manager.security_manager.validate_subprocess_path.return_value = True

        # Mock git status output (empty = clean)
        mock_run.return_value = MagicMock(stdout="", returncode=0)

        result = git_manager.check_git_status(temp_dir)

        assert result["is_git_repo"] is True
        assert result["is_clean"] is True
        assert result["total_changes"] == 0
        assert result["modified_files"] == []
        assert result["untracked_files"] == []
        assert result["staged_files"] == []

    @patch("subprocess.run")
    def test_check_git_status_with_changes(self, mock_run, git_manager, temp_dir):
        """Test Git status check with various file changes."""
        # Create .git directory
        (temp_dir / ".git").mkdir()

        # Mock security validation to return True
        git_manager.security_manager.validate_subprocess_path.return_value = True

        # Mock git status output with various changes
        status_output = """M  modified_file.py
A  added_file.py
 M unstaged_file.py
?? untracked_file.py"""

        mock_run.return_value = MagicMock(stdout=status_output, returncode=0)

        result = git_manager.check_git_status(temp_dir)

        assert result["is_git_repo"] is True
        assert result["is_clean"] is False
        assert result["total_changes"] == 4
        assert "modified_file.py" in result["staged_files"]
        assert "added_file.py" in result["staged_files"]
        assert "unstaged_file.py" in result["modified_files"]
        assert "untracked_file.py" in result["untracked_files"]

    @patch("subprocess.run")
    def test_check_git_status_command_failed(self, mock_run, git_manager, temp_dir):
        """Test Git status check when git command fails."""
        # Create .git directory
        (temp_dir / ".git").mkdir()

        # Mock security validation to return True
        git_manager.security_manager.validate_subprocess_path.return_value = True

        # Mock git command failure
        mock_run.side_effect = subprocess.CalledProcessError(
            1, ["git", "status"], stderr="Error"
        )

        result = git_manager.check_git_status(temp_dir)

        assert result["is_git_repo"] is True
        assert "error" in result

    @patch.object(GitManager, "_check_git_available")
    @patch.object(GitManager, "check_git_status")
    @patch.object(GitManager, "_get_remote_info")
    def test_get_git_info_complete(
        self, mock_remote_info, mock_status, mock_check_git, git_manager, temp_dir
    ):
        """Test getting complete Git information."""
        # Create .git directory
        (temp_dir / ".git").mkdir()

        # Mock return values
        mock_check_git.return_value = True
        mock_status.return_value = {"is_clean": True}
        mock_remote_info.return_value = {
            "remotes": {"origin": {"fetch": "git@github.com:test/repo.git"}}
        }

        result = git_manager.get_git_info(temp_dir)

        assert result["git_available"] is True
        assert result["is_git_repo"] is True
        assert "status" in result
        assert "remote_info" in result

    @patch.object(GitManager, "_check_git_available")
    def test_get_git_info_git_not_available(
        self, mock_check_git, git_manager, temp_dir
    ):
        """Test getting Git info when Git is not available."""
        mock_check_git.return_value = False

        result = git_manager.get_git_info(temp_dir)

        assert result["git_available"] is False
        assert result["is_git_repo"] is False

    @patch("subprocess.run")
    def test_get_remote_info_success(self, mock_run, git_manager, temp_dir):
        """Test getting Git remote information."""
        # Mock security validation to return True
        git_manager.security_manager.validate_subprocess_path.return_value = True

        # Mock git remote output
        remote_output = """origin	git@github.com:test/repo.git (fetch)
origin	git@github.com:test/repo.git (push)
upstream	https://github.com/upstream/repo.git (fetch)
upstream	https://github.com/upstream/repo.git (push)"""

        mock_run.return_value = MagicMock(stdout=remote_output, returncode=0)

        result = git_manager._get_remote_info(temp_dir)

        assert "remotes" in result
        remotes = result["remotes"]
        assert "origin" in remotes
        assert "upstream" in remotes
        assert remotes["origin"]["fetch"] == "git@github.com:test/repo.git"
        assert remotes["upstream"]["fetch"] == "https://github.com/upstream/repo.git"

    @patch("subprocess.run")
    def test_get_remote_info_security_failure(self, mock_run, git_manager, temp_dir):
        """Test getting remote info when security validation fails."""
        # Mock security validation to return False
        git_manager.security_manager.validate_subprocess_path.return_value = False

        result = git_manager._get_remote_info(temp_dir)

        assert "error" in result
        assert "Security validation failed" in result["error"]
        mock_run.assert_not_called()

    @patch("subprocess.run")
    def test_get_remote_info_no_remotes(self, mock_run, git_manager, temp_dir):
        """Test getting remote info when no remotes exist."""
        # Mock security validation to return True
        git_manager.security_manager.validate_subprocess_path.return_value = True

        # Mock empty git remote output
        mock_run.return_value = MagicMock(stdout="", returncode=0)

        result = git_manager._get_remote_info(temp_dir)

        assert result == {"remotes": {}}

    @patch("subprocess.run")
    def test_get_remote_info_command_failed(self, mock_run, git_manager, temp_dir):
        """Test getting remote info when git command fails."""
        # Mock security validation to return True
        git_manager.security_manager.validate_subprocess_path.return_value = True

        # Mock git command failure
        mock_run.side_effect = subprocess.CalledProcessError(
            1, ["git", "remote"], stderr="Error"
        )

        result = git_manager._get_remote_info(temp_dir)

        assert "error" in result

    def test_integration_workflow(self, git_manager, temp_dir):
        """Test complete Git management workflow."""
        # Mock all external dependencies
        with patch.object(git_manager, "_check_git_available", return_value=True):
            with patch.object(git_manager, "_initialize_repository", return_value=True):
                git_manager.security_manager.validate_subprocess_path.return_value = (
                    True
                )

                # Initialize repository
                success, was_initialized = git_manager.initialize_git_repository(
                    temp_dir
                )
                assert success is True
                assert was_initialized is True

                # Check status (simulate clean repo)
                with patch("subprocess.run") as mock_run:
                    mock_run.return_value = MagicMock(stdout="", returncode=0)

                    # Create .git for status check
                    (temp_dir / ".git").mkdir()

                    status = git_manager.check_git_status(temp_dir)
                    assert status["is_git_repo"] is True
                    assert status["is_clean"] is True

                # Get full Git info
                git_info = git_manager.get_git_info(temp_dir)
                assert git_info["git_available"] is True
                assert git_info["is_git_repo"] is True
