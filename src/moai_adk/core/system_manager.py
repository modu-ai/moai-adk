"""
@FEATURE:SYSTEM-001 System utilities and environment checks for MoAI-ADK

Handles Node.js/npm detection, ccusage availability checks,
and other system-level validations.
"""

import subprocess
from typing import Any

import click

from ..utils.logger import get_logger

logger = get_logger(__name__)


class SystemManager:
    """@TASK:SYSTEM-MANAGER-001 Manages system-level checks and validations."""

    def __init__(self):
        """Initialize system manager."""

    def check_nodejs_and_npm(self) -> bool:
        """
        Check if Node.js and npm are installed, and verify ccusage can be used.

        Returns:
            bool: True if Node.js environment is properly set up
        """
        logger.info("Node.js 환경 확인 시작")
        click.echo("\n📋 Node.js 환경 확인 중...")

        # Check Node.js
        if not self._check_command_exists("node"):
            logger.warning("Node.js가 설치되어 있지 않음")
            click.echo("⚠️  Node.js가 설치되어 있지 않습니다.")
            click.echo("   ccusage statusLine 기능을 사용하려면 Node.js가 필요합니다.")
            click.echo("   Node.js 설치: https://nodejs.org")
            return False

        # Check npm
        if not self._check_command_exists("npm"):
            logger.warning("npm이 설치되어 있지 않음")
            click.echo("⚠️  npm이 설치되어 있지 않습니다.")
            click.echo("   ccusage statusLine 기능을 사용하려면 npm이 필요합니다.")
            return False

        # Get versions and test ccusage
        return self._validate_nodejs_environment()

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

    def _validate_nodejs_environment(self) -> bool:
        """Validate Node.js environment and ccusage availability."""
        try:
            # Get Node.js and npm versions
            node_result = subprocess.run(
                ["node", "--version"],
                capture_output=True,
                text=True,
                check=True
            )
            npm_result = subprocess.run(
                ["npm", "--version"],
                capture_output=True,
                text=True,
                check=True
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
                check=False  # Don't fail on non-zero exit
            )

            if ccusage_result.returncode == 0 or "ccusage" in ccusage_result.stdout.lower():
                logger.info("ccusage 패키지 접근 가능 확인됨")
                click.echo("✅ ccusage 패키지 접근 가능 확인됨")
                click.echo("💡 statusLine에서 실시간 Claude Code 사용량 추적이 활성화됩니다.")
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

    def get_system_info(self) -> dict[str, Any]:
        """
        Get comprehensive system information.

        Returns:
            dict: System information including OS, Python, Node.js, etc.
        """
        import platform
        import sys

        system_info = {
            'platform': {
                'system': platform.system(),
                'release': platform.release(),
                'machine': platform.machine(),
                'processor': platform.processor(),
            },
            'python': {
                'version': sys.version,
                'version_info': {
                    'major': sys.version_info.major,
                    'minor': sys.version_info.minor,
                    'micro': sys.version_info.micro,
                },
                'executable': sys.executable,
            },
            'nodejs': self._get_nodejs_info(),
            'git': {'available': self._check_command_exists('git')},
        }

        # Add package managers info
        system_info['package_managers'] = self._get_package_managers_info()

        return system_info

    def _get_nodejs_info(self) -> dict[str, Any]:
        """Get Node.js environment information."""
        nodejs_info = {
            'node_available': self._check_command_exists('node'),
            'npm_available': self._check_command_exists('npm'),
            'yarn_available': self._check_command_exists('yarn'),
            'pnpm_available': self._check_command_exists('pnpm'),
        }

        if nodejs_info['node_available']:
            try:
                result = subprocess.run(
                    ["node", "--version"],
                    capture_output=True,
                    text=True,
                    check=True
                )
                nodejs_info['node_version'] = result.stdout.strip()
            except Exception as e:
                logger.error("Failed to get Node.js version: %s", e)

        if nodejs_info['npm_available']:
            try:
                result = subprocess.run(
                    ["npm", "--version"],
                    capture_output=True,
                    text=True,
                    check=True
                )
                nodejs_info['npm_version'] = result.stdout.strip()
            except Exception as e:
                logger.error("Failed to get npm version: %s", e)

        # Test ccusage if npm available
        if nodejs_info['npm_available']:
            nodejs_info['ccusage_available'] = self._quick_ccusage_test()

        return nodejs_info

    def _quick_ccusage_test(self) -> bool:
        """Quick test for ccusage availability without output."""
        try:
            result = subprocess.run(
                ["npx", "-y", "ccusage", "--help"],
                capture_output=True,
                text=True,
                timeout=10,
                check=False
            )
            return result.returncode == 0 or "ccusage" in result.stdout.lower()
        except Exception:
            return False

    def _get_package_managers_info(self) -> dict[str, bool]:
        """Get information about available package managers."""
        return {
            'pip': self._check_command_exists('pip'),
            'pip3': self._check_command_exists('pip3'),
            'conda': self._check_command_exists('conda'),
            'brew': self._check_command_exists('brew'),  # macOS
            'apt': self._check_command_exists('apt'),    # Ubuntu/Debian
            'yum': self._check_command_exists('yum'),    # CentOS/RHEL
            'dnf': self._check_command_exists('dnf'),    # Fedora
            'choco': self._check_command_exists('choco'), # Windows
            'winget': self._check_command_exists('winget'), # Windows 10+
        }

    def check_python_version(self, min_version: tuple = (3, 8)) -> bool:
        """
        Check if Python version meets minimum requirements.

        Args:
            min_version: Minimum required version as tuple (major, minor)

        Returns:
            bool: True if version is sufficient
        """
        import sys

        current_version = (sys.version_info.major, sys.version_info.minor)
        return current_version >= min_version

    def detect_project_type(self, project_path) -> dict[str, Any]:
        """
        Detect project type based on existing files.

        Args:
            project_path: Path to project directory

        Returns:
            dict: Detected project information
        """
        from pathlib import Path

        project_path = Path(project_path)
        detected = {
            'type': 'unknown',
            'language': 'unknown',
            'frameworks': [],
            'build_tools': [],
            'files_found': []
        }

        # Check for various project files
        project_files = {
            'package.json': {'type': 'nodejs', 'language': 'javascript'},
            'requirements.txt': {'type': 'python', 'language': 'python'},
            'pyproject.toml': {'type': 'python', 'language': 'python'},
            'Cargo.toml': {'type': 'rust', 'language': 'rust'},
            'go.mod': {'type': 'go', 'language': 'go'},
            'pom.xml': {'type': 'java', 'language': 'java'},
            'build.gradle': {'type': 'java', 'language': 'java'},
            'Gemfile': {'type': 'ruby', 'language': 'ruby'},
            'composer.json': {'type': 'php', 'language': 'php'},
        }

        for file_name, info in project_files.items():
            file_path = project_path / file_name
            if file_path.exists():
                detected['files_found'].append(file_name)
                detected['type'] = info['type']
                detected['language'] = info['language']

        # Detect frameworks and build tools
        if (project_path / 'package.json').exists():
            detected.update(self._analyze_package_json(project_path / 'package.json'))

        return detected

    def _analyze_package_json(self, package_json_path) -> dict[str, Any]:
        """Analyze package.json for frameworks and dependencies."""
        import json

        try:
            with open(package_json_path, encoding='utf-8') as f:
                package_data = json.load(f)

            frameworks = []
            build_tools = []

            # Check dependencies and devDependencies
            all_deps = {}
            all_deps.update(package_data.get('dependencies', {}))
            all_deps.update(package_data.get('devDependencies', {}))

            # Detect frameworks
            framework_indicators = {
                'react': ['react', '@types/react'],
                'vue': ['vue', '@vue/cli'],
                'angular': ['@angular/core', '@angular/cli'],
                'svelte': ['svelte'],
                'nextjs': ['next'],
                'nuxtjs': ['nuxt'],
                'express': ['express'],
                'fastify': ['fastify'],
            }

            for framework, indicators in framework_indicators.items():
                if any(indicator in all_deps for indicator in indicators):
                    frameworks.append(framework)

            # Detect build tools
            build_tool_indicators = {
                'webpack': ['webpack'],
                'vite': ['vite'],
                'rollup': ['rollup'],
                'parcel': ['parcel'],
                'typescript': ['typescript', '@types/node'],
            }

            for tool, indicators in build_tool_indicators.items():
                if any(indicator in all_deps for indicator in indicators):
                    build_tools.append(tool)

            return {
                'frameworks': frameworks,
                'build_tools': build_tools,
                'has_scripts': bool(package_data.get('scripts')),
                'scripts': list(package_data.get('scripts', {}).keys()),
            }

        except Exception as e:
            logger.error("Error analyzing package.json: %s", e)
            return {'frameworks': [], 'build_tools': []}

    def should_create_package_json(self, config) -> bool:
        """
        Check if package.json should be created based on project configuration.

        Args:
            config: Project configuration

        Returns:
            bool: True if package.json should be created
        """
        # Only create package.json for explicit Node.js/web projects
        return config.runtime.name in ["node", "tsx"] or any(
            tech in config.tech_stack
            for tech in ["nextjs", "react", "vue", "angular", "svelte"]
        )
