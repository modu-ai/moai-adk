"""
@FEATURE:GIT-002 Git installation management for MoAI-ADK

Handles Git installation detection and automated installation
with cross-platform support and user interaction.
"""

import platform
import subprocess
from pathlib import Path

import click

from ..utils.logger import get_logger

logger = get_logger(__name__)


class GitInstallationManager:
    """@TASK:GIT-INSTALL-001 Manages Git installation and availability checking."""

    def __init__(self):
        """Initialize Git installation manager."""
        self.os_name = platform.system().lower()

    def check_git_available(self) -> bool:
        """Check if git is available in the system."""
        try:
            subprocess.run(
                ["git", "--version"], capture_output=True, text=True, check=True
            )
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            logger.warning("Git not found in system")
            return False

    def offer_git_installation(self) -> bool:
        """Offer to install git and attempt installation if user agrees."""
        logger.warning("Git is not installed on your system")
        click.echo("\n" + "=" * 60)
        click.echo("🔧 Git is not installed on your system.")
        click.echo(
            "   Git is required for MoAI-ADK version control and CI/CD features."
        )
        click.echo("=" * 60)

        # Show installation command based on OS
        install_cmd = self._get_git_install_command()

        # Ask for user confirmation
        logger.info("사용자에게 Git 자동 설치 여부 확인 중")
        click.echo("\n🤔 Git을 자동으로 설치하시겠습니까? (y/N): ", nl=False)

        try:
            response = input().strip().lower()
            if response in ["y", "yes", "예"]:
                if install_cmd and self.os_name != "windows":
                    logger.info(f"Git 자동 설치 시작: {' '.join(install_cmd)}")
                    click.echo(f"🚀 Git 설치 중... (명령어: {' '.join(install_cmd)})")
                    return self._install_git_with_command(install_cmd)
                else:
                    logger.warning("자동 설치가 지원되지 않는 환경")
                    click.echo("⚠️ 자동 설치가 지원되지 않는 환경입니다.")
                    click.echo(
                        "   위 안내에 따라 수동으로 Git을 설치한 후 다시 실행해주세요."
                    )
                    return False
            else:
                logger.info("사용자가 Git 설치를 거부")
                click.echo("⏭️ Git 설치를 건너뛰었습니다.")
                return False

        except (KeyboardInterrupt, EOFError):
            logger.info("사용자가 Git 설치를 취소")
            click.echo("\n⏭️ Git 설치를 건너뛰었습니다.")
            return False

    def _get_git_install_command(self) -> list | None:
        """Get Git installation command based on OS."""
        install_cmd = None

        if self.os_name == "darwin":  # macOS
            if self._check_command_exists("brew"):
                install_cmd = ["brew", "install", "git"]
                logger.info("macOS Homebrew 환경에서 Git 설치 가능")
                click.echo("💡 Homebrew를 사용하여 Git을 설치할 수 있습니다:")
                click.echo("   brew install git")
            else:
                logger.info("macOS 환경에서 Git 수동 설치 안내")
                click.echo("💡 Git 설치 방법:")
                click.echo("   1. Homebrew 설치 후: brew install git")
                click.echo(
                    "   2. 또는 https://git-scm.com/download/mac 에서 직접 다운로드"
                )

        elif self.os_name == "linux":
            # Check for different package managers
            if self._check_command_exists("apt"):
                install_cmd = [
                    "sudo", "apt", "update", "&&",
                    "sudo", "apt", "install", "-y", "git"
                ]
                logger.info("Linux APT 환경에서 Git 설치 가능")
                click.echo("💡 APT를 사용하여 Git을 설치할 수 있습니다:")
                click.echo("   sudo apt update && sudo apt install -y git")
            elif self._check_command_exists("yum"):
                install_cmd = ["sudo", "yum", "install", "-y", "git"]
                logger.info("Linux YUM 환경에서 Git 설치 가능")
                click.echo("💡 YUM을 사용하여 Git을 설치할 수 있습니다:")
                click.echo("   sudo yum install -y git")
            elif self._check_command_exists("dnf"):
                install_cmd = ["sudo", "dnf", "install", "-y", "git"]
                logger.info("Linux DNF 환경에서 Git 설치 가능")
                click.echo("💡 DNF를 사용하여 Git을 설치할 수 있습니다:")
                click.echo("   sudo dnf install -y git")
            else:
                logger.info("Linux 환경에서 Git 수동 설치 안내")
                click.echo("💡 패키지 매니저를 통해 Git을 설치하세요:")
                click.echo("   - Ubuntu/Debian: sudo apt install git")
                click.echo("   - CentOS/RHEL: sudo yum install git")
                click.echo("   - Fedora: sudo dnf install git")

        elif self.os_name == "windows":
            logger.info("Windows 환경에서 Git 수동 설치 안내")
            click.echo("💡 Git 설치 방법:")
            click.echo(
                "   1. https://git-scm.com/download/win 에서 Git for Windows 다운로드"
            )
            click.echo("   2. 또는 Chocolatey 사용: choco install git")
            click.echo("   3. 또는 Winget 사용: winget install Git.Git")

        return install_cmd

    def _check_command_exists(self, command: str) -> bool:
        """Check if a command exists in the system."""
        try:
            subprocess.run(
                [command, "--version"], capture_output=True, text=True, check=True
            )
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            return False

    def _install_git_with_command(self, install_cmd: list) -> bool:
        """Install git using the provided command."""
        try:
            if self.os_name == "linux" and "&&" in install_cmd:
                # Handle complex commands by executing them separately for security
                commands = []
                current_cmd = []

                for part in install_cmd:
                    if part == "&&":
                        if current_cmd:
                            commands.append(current_cmd)
                            current_cmd = []
                    else:
                        current_cmd.append(part)

                if current_cmd:
                    commands.append(current_cmd)

                # Execute each command separately
                for cmd in commands:
                    result = subprocess.run(
                        cmd,
                        capture_output=True,
                        text=True,
                        timeout=300,  # 5 minute timeout
                    )
                    # If any command fails, stop execution
                    if result.returncode != 0:
                        break
            else:
                result = subprocess.run(
                    install_cmd,
                    capture_output=True,
                    text=True,
                    timeout=300,  # 5 minute timeout
                )

            if result.returncode == 0:
                logger.info("Git 설치가 완료되었습니다")
                click.echo("✅ Git 설치가 완료되었습니다!")
                return True
            else:
                logger.error(f"Git 설치 중 오류 발생: {result.stderr}")
                click.echo("❌ Git 설치 중 오류 발생:")
                click.echo(f"   {result.stderr}")
                return False

        except subprocess.TimeoutExpired:
            logger.error("Git 설치가 시간 초과되었습니다")
            click.echo("❌ Git 설치가 시간 초과되었습니다.")
            return False
        except Exception as e:
            logger.error(f"Git 설치 중 예상치 못한 오류 발생: {e}")
            click.echo(f"❌ Git 설치 중 예상치 못한 오류 발생: {e}")
            return False