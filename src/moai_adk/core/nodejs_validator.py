"""
@FEATURE:NODEJS-001 Node.js environment validation for MoAI-ADK

Handles Node.js, npm, and ccusage validation with comprehensive error handling.
Extracted from system_manager.py for TRUST compliance (≤300 LOC).
"""

import subprocess
from typing import Any

import click

from ..utils.logger import get_logger
from .command_checker import CommandChecker

logger = get_logger(__name__)


class NodejsValidator:
    """@TASK:NODEJS-VALIDATOR-001 Validates Node.js environment and related tools."""

    def __init__(self):
        """Initialize Node.js validator with command checker dependency."""
        self.command_checker = CommandChecker()

    def check_nodejs_and_npm(self) -> bool:
        """
        Check if Node.js and npm are installed, and verify ccusage can be used.

        Returns:
            bool: True if Node.js environment is properly set up
        """
        logger.info("Node.js 환경 확인 시작")
        click.echo("\n📋 Node.js 환경 확인 중...")

        # Check Node.js
        if not self.command_checker.check_command_exists("node"):
            logger.warning("Node.js가 설치되어 있지 않음")
            click.echo("⚠️  Node.js가 설치되어 있지 않습니다.")
            click.echo("   ccusage statusLine 기능을 사용하려면 Node.js가 필요합니다.")
            click.echo("   Node.js 설치: https://nodejs.org")
            return False

        # Check npm
        if not self.command_checker.check_command_exists("npm"):
            logger.warning("npm이 설치되어 있지 않음")
            click.echo("⚠️  npm이 설치되어 있지 않습니다.")
            click.echo("   ccusage statusLine 기능을 사용하려면 npm이 필요합니다.")
            return False

        # Get versions and test ccusage
        return self._validate_nodejs_environment()

    def _validate_nodejs_environment(self) -> bool:
        """Validate Node.js environment and ccusage availability."""
        try:
            # Get Node.js and npm versions
            node_result = subprocess.run(
                ["node", "--version"], capture_output=True, text=True, check=True
            )
            npm_result = subprocess.run(
                ["npm", "--version"], capture_output=True, text=True, check=True
            )

            node_version = node_result.stdout.strip()
            npm_version = npm_result.stdout.strip()

            logger.info(f"Node.js {node_version}, npm {npm_version} 감지됨")
            click.echo(f"✅ Node.js {node_version} 감지됨")
            click.echo(f"✅ npm {npm_version} 감지됨")

            # Test ccusage availability
            return self._test_ccusage_availability()

        except subprocess.CalledProcessError as e:
            logger.error("Node.js/npm version check failed: %s", e)
            click.echo(f"❌ Node.js/npm 버전 확인 실패: {e}")
            return False
        except Exception as e:
            logger.error("Node.js environment validation error: %s", e)
            click.echo(f"❌ Node.js 환경 검사 중 오류: {e}")
            return False

    def _test_ccusage_availability(self) -> bool:
        """Test ccusage package accessibility."""
        logger.info("ccusage 패키지 접근 테스트 시작")
        click.echo("📦 ccusage 패키지 접근 테스트 중...")
        try:
            # Test if npx can access ccusage (without actually running it)
            ccusage_result = subprocess.run(
                ["npx", "-y", "ccusage", "--help"],
                capture_output=True,
                text=True,
                timeout=30,
                check=False,  # Don't fail on non-zero exit
            )

            if (
                ccusage_result.returncode == 0
                or "ccusage" in ccusage_result.stdout.lower()
            ):
                logger.info("ccusage 패키지 접근 가능 확인됨")
                click.echo("✅ ccusage 패키지 접근 가능 확인됨")
                click.echo(
                    "💡 statusLine에서 실시간 Claude Code 사용량 추적이 활성화됩니다."
                )
                return True
            else:
                logger.warning("ccusage 패키지 접근 실패")
                click.echo("⚠️  ccusage 패키지 접근 실패")
                click.echo("   statusLine 기능이 제한될 수 있습니다.")
                return False

        except subprocess.TimeoutExpired:
            logger.warning("ccusage 접근 테스트 시간 초과")
            click.echo("⚠️  ccusage 접근 테스트 시간 초과")
            click.echo("   네트워크 상태를 확인해주세요.")
            return False
        except Exception as e:
            logger.error("ccusage test error: %s", e)
            click.echo(f"⚠️  ccusage 테스트 중 오류: {e}")
            return False

    def get_nodejs_info(self) -> dict[str, Any]:
        """Get Node.js environment information."""
        nodejs_info = {
            "node_available": self.command_checker.check_command_exists("node"),
            "npm_available": self.command_checker.check_command_exists("npm"),
            "yarn_available": self.command_checker.check_command_exists("yarn"),
            "pnpm_available": self.command_checker.check_command_exists("pnpm"),
        }

        if nodejs_info["node_available"]:
            try:
                result = subprocess.run(
                    ["node", "--version"], capture_output=True, text=True, check=True
                )
                nodejs_info["node_version"] = result.stdout.strip()
            except Exception as e:
                logger.error("Failed to get Node.js version: %s", e)

        if nodejs_info["npm_available"]:
            try:
                result = subprocess.run(
                    ["npm", "--version"], capture_output=True, text=True, check=True
                )
                nodejs_info["npm_version"] = result.stdout.strip()
            except Exception as e:
                logger.error("Failed to get npm version: %s", e)

        # Test ccusage if npm available
        if nodejs_info["npm_available"]:
            nodejs_info["ccusage_available"] = self._quick_ccusage_test()

        logger.info(f"Node.js environment info collected: {nodejs_info}")
        return nodejs_info

    def _quick_ccusage_test(self) -> bool:
        """Quick test for ccusage availability without output."""
        try:
            result = subprocess.run(
                ["npx", "-y", "ccusage", "--help"],
                capture_output=True,
                text=True,
                timeout=10,
                check=False,
            )
            is_available = result.returncode == 0 or "ccusage" in result.stdout.lower()
            logger.debug(f"ccusage quick test result: {is_available}")
            return is_available
        except Exception as e:
            logger.debug(f"ccusage quick test failed: {e}")
            return False