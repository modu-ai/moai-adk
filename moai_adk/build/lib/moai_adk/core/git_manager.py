"""
Git repository management utilities for MoAI-ADK.

Handles Git repository initialization, validation, and operations
with security validation and error handling.
"""

import subprocess
import platform
from pathlib import Path
from typing import Tuple

from ..logger import get_logger
from .security import SecurityManager
from .file_manager import FileManager

logger = get_logger(__name__)


class GitManager:
    """Manages Git operations for MoAI-ADK installation."""

    def __init__(self, security_manager: SecurityManager = None, file_manager: FileManager = None):
        """
        Initialize Git manager.

        Args:
            security_manager: Security manager instance for validation
            file_manager: File manager for .gitignore creation
        """
        self.security_manager = security_manager or SecurityManager()
        self.file_manager = file_manager

    def initialize_repository(self, project_path: Path) -> bool:
        """
        Initialize git repository if not already initialized.

        Args:
            project_path: Project root path

        Returns:
            bool: True if git repo exists or was successfully initialized
        """
        try:
            # Check if already a git repository
            git_dir = project_path / ".git"
            if git_dir.exists():
                logger.info("Git repository already exists - skipping initialization")
                return True

            # Check if git is available
            if not self._check_git_available():
                if self._offer_git_installation():
                    # Try again after installation
                    if not self._check_git_available():
                        logger.error("Git still not available after installation attempt")
                        return False
                else:
                    logger.info("Git installation declined - skipping git initialization")
                    return False

            # Initialize git repository
            if not self._initialize_repository(project_path):
                return False

            # Create initial .gitignore if it doesn't exist
            gitignore_path = project_path / ".gitignore"
            if not gitignore_path.exists() and self.file_manager:
                self.file_manager.create_gitignore(gitignore_path)

            logger.info("Git repository initialized successfully")
            return True

        except Exception as e:
            logger.error("Error initializing git repository: %s", e)
            return False

    def _check_git_available(self) -> bool:
        """Check if git is available in the system."""
        try:
            subprocess.run(
                ["git", "--version"],
                capture_output=True,
                text=True,
                check=True
            )
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            logger.warning("Git not found in system")
            return False

    def _initialize_repository(self, project_path: Path) -> bool:
        """Initialize git repository with security validation."""
        try:
            # Security validation
            if not self.security_manager.validate_subprocess_path(project_path, project_path):
                logger.error("Security: Invalid path for git initialization: %s", project_path)
                return False

            result = subprocess.run(
                ["git", "init"],
                cwd=project_path,
                capture_output=True,
                text=True
            )

            if result.returncode == 0:
                logger.debug("Git init completed successfully")
                return True
            else:
                logger.error("Failed to initialize git: %s", result.stderr)
                return False

        except Exception as e:
            logger.error("Error during git initialization: %s", e)
            return False

    def _offer_git_installation(self) -> bool:
        """Offer to install git and attempt installation if user agrees."""
        print("\n" + "=" * 60)
        print("🔧 Git is not installed on your system.")
        print("   Git is required for MoAI-ADK version control and CI/CD features.")
        print("=" * 60)

        # Show installation command based on OS
        os_name = platform.system().lower()
        install_cmd = self._get_git_install_command(os_name)

        # Ask for user confirmation
        print("\n🤔 Git을 자동으로 설치하시겠습니까? (y/N): ", end="", flush=True)

        try:
            response = input().strip().lower()
            if response in ['y', 'yes', '예']:
                if install_cmd and os_name != "windows":
                    print(f"🚀 Git 설치 중... (명령어: {' '.join(install_cmd)})")
                    return self._install_git_with_command(install_cmd, os_name)
                else:
                    print("⚠️ 자동 설치가 지원되지 않는 환경입니다.")
                    print("   위 안내에 따라 수동으로 Git을 설치한 후 다시 실행해주세요.")
                    return False
            else:
                print("⏭️ Git 설치를 건너뛰었습니다.")
                return False

        except (KeyboardInterrupt, EOFError):
            print("\n⏭️ Git 설치를 건너뛰었습니다.")
            return False

    def _get_git_install_command(self, os_name: str) -> list | None:
        """Get Git installation command based on OS."""
        install_cmd = None

        if os_name == "darwin":  # macOS
            if self._check_command_exists("brew"):
                install_cmd = ["brew", "install", "git"]
                print("💡 Homebrew를 사용하여 Git을 설치할 수 있습니다:")
                print("   brew install git")
            else:
                print("💡 Git 설치 방법:")
                print("   1. Homebrew 설치 후: brew install git")
                print("   2. 또는 https://git-scm.com/download/mac 에서 직접 다운로드")

        elif os_name == "linux":
            # Check for different package managers
            if self._check_command_exists("apt"):
                install_cmd = ["sudo", "apt", "update", "&&", "sudo", "apt", "install", "-y", "git"]
                print("💡 APT를 사용하여 Git을 설치할 수 있습니다:")
                print("   sudo apt update && sudo apt install -y git")
            elif self._check_command_exists("yum"):
                install_cmd = ["sudo", "yum", "install", "-y", "git"]
                print("💡 YUM을 사용하여 Git을 설치할 수 있습니다:")
                print("   sudo yum install -y git")
            elif self._check_command_exists("dnf"):
                install_cmd = ["sudo", "dnf", "install", "-y", "git"]
                print("💡 DNF를 사용하여 Git을 설치할 수 있습니다:")
                print("   sudo dnf install -y git")
            else:
                print("💡 패키지 매니저를 통해 Git을 설치하세요:")
                print("   - Ubuntu/Debian: sudo apt install git")
                print("   - CentOS/RHEL: sudo yum install git")
                print("   - Fedora: sudo dnf install git")

        elif os_name == "windows":
            print("💡 Git 설치 방법:")
            print("   1. https://git-scm.com/download/win 에서 Git for Windows 다운로드")
            print("   2. 또는 Chocolatey 사용: choco install git")
            print("   3. 또는 Winget 사용: winget install Git.Git")

        return install_cmd

    def _check_command_exists(self, command: str) -> bool:
        """Check if a command exists in the system."""
        try:
            subprocess.run(
                [command, "--version"],
                capture_output=True,
                text=True,
                check=True
            )
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            return False

    def _install_git_with_command(self, install_cmd: list, os_name: str) -> bool:
        """Install git using the provided command."""
        try:
            if os_name == "linux" and "&&" in install_cmd:
                # Handle complex commands with &&
                cmd_str = " ".join(install_cmd)
                result = subprocess.run(
                    cmd_str,
                    shell=True,
                    capture_output=True,
                    text=True,
                    timeout=300  # 5 minute timeout
                )
            else:
                result = subprocess.run(
                    install_cmd,
                    capture_output=True,
                    text=True,
                    timeout=300  # 5 minute timeout
                )

            if result.returncode == 0:
                print("✅ Git 설치가 완료되었습니다!")
                return True
            else:
                print(f"❌ Git 설치 중 오류 발생:")
                print(f"   {result.stderr}")
                return False

        except subprocess.TimeoutExpired:
            print("❌ Git 설치가 시간 초과되었습니다.")
            return False
        except Exception as e:
            print(f"❌ Git 설치 중 예상치 못한 오류 발생: {e}")
            return False

    def check_git_status(self, project_path: Path) -> dict:
        """
        Check the status of Git repository.

        Args:
            project_path: Project root path

        Returns:
            dict: Git status information
        """
        try:
            if not (project_path / ".git").exists():
                return {
                    'is_git_repo': False,
                    'error': 'Not a git repository'
                }

            # Security validation
            if not self.security_manager.validate_subprocess_path(project_path, project_path):
                return {
                    'is_git_repo': True,
                    'error': 'Security validation failed for git status'
                }

            # Get git status
            result = subprocess.run(
                ["git", "status", "--porcelain"],
                cwd=project_path,
                capture_output=True,
                text=True,
                check=True
            )

            # Parse status output
            status_lines = result.stdout.strip().split('\n') if result.stdout.strip() else []

            modified_files = []
            untracked_files = []
            staged_files = []

            for line in status_lines:
                if len(line) >= 3:
                    status_code = line[:2]
                    filename = line[3:]

                    if status_code[0] in ['M', 'A', 'D', 'R', 'C']:
                        staged_files.append(filename)
                    if status_code[1] in ['M', 'D']:
                        modified_files.append(filename)
                    if status_code == '??':
                        untracked_files.append(filename)

            return {
                'is_git_repo': True,
                'is_clean': len(status_lines) == 0,
                'modified_files': modified_files,
                'untracked_files': untracked_files,
                'staged_files': staged_files,
                'total_changes': len(status_lines)
            }

        except subprocess.CalledProcessError as e:
            logger.error("Git status command failed: %s", e.stderr)
            return {
                'is_git_repo': True,
                'error': f'Git command failed: {e.stderr}'
            }
        except Exception as e:
            logger.error("Error checking git status: %s", e)
            return {
                'is_git_repo': False,
                'error': str(e)
            }

    def get_git_info(self, project_path: Path) -> dict:
        """
        Get comprehensive Git repository information.

        Args:
            project_path: Project root path

        Returns:
            dict: Git repository information
        """
        git_info = {
            'git_available': self._check_git_available(),
            'is_git_repo': (project_path / ".git").exists(),
            'status': {},
            'remote_info': {},
        }

        if git_info['git_available'] and git_info['is_git_repo']:
            git_info['status'] = self.check_git_status(project_path)
            git_info['remote_info'] = self._get_remote_info(project_path)

        return git_info

    def _get_remote_info(self, project_path: Path) -> dict:
        """Get Git remote information."""
        try:
            # Security validation
            if not self.security_manager.validate_subprocess_path(project_path, project_path):
                return {'error': 'Security validation failed'}

            result = subprocess.run(
                ["git", "remote", "-v"],
                cwd=project_path,
                capture_output=True,
                text=True,
                check=True
            )

            remotes = {}
            for line in result.stdout.strip().split('\n'):
                if line:
                    parts = line.split()
                    if len(parts) >= 2:
                        remote_name = parts[0]
                        remote_url = parts[1]
                        remote_type = parts[2] if len(parts) > 2 else ''

                        if remote_name not in remotes:
                            remotes[remote_name] = {}

                        if '(fetch)' in remote_type:
                            remotes[remote_name]['fetch'] = remote_url
                        elif '(push)' in remote_type:
                            remotes[remote_name]['push'] = remote_url

            return {'remotes': remotes}

        except subprocess.CalledProcessError as e:
            return {'error': f'Git remote command failed: {e.stderr}'}
        except Exception as e:
            return {'error': str(e)}

    def create_gitignore(self, gitignore_path: Path) -> bool:
        """
        Create a comprehensive .gitignore file.

        Args:
            gitignore_path: Path where .gitignore should be created

        Returns:
            bool: True if successful
        """
        if self.file_manager:
            return self.file_manager.create_gitignore(gitignore_path)
        else:
            logger.warning("FileManager not available for gitignore creation")
            return False