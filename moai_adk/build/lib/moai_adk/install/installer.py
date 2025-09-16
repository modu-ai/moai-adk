"""
Simplified MoAI-ADK Project Installer

Simplified installation system using embedded package resources instead of symbolic links.
Ensures perfect compatibility with Claude Code by directly copying all resources through ResourceManager.
"""

from pathlib import Path
from typing import List, Optional
from collections.abc import Callable

from ..logger import get_logger
from ..config import Config
from ..installation_result import InstallationResult
from ..progress_tracker import ProgressTracker
from ..core.security import SecurityManager
from ..core.directory_manager import DirectoryManager
from ..core.config_manager import ConfigManager
from ..core.git_manager import GitManager
from .resource_manager import ResourceManager

logger = get_logger(__name__)


class SimplifiedInstaller:
    """
    Simplified MoAI-ADK project installation manager

    Installation system that directly copies embedded package resources
    instead of symbolic links for stable operation across all platforms.
    """

    def __init__(self, config: Config):
        """
        Initialize installation manager

        Args:
            config: Project configuration
        """
        self.config = config
        self.progress = ProgressTracker()

        # Initialize core managers
        self.security_manager = SecurityManager()
        self.directory_manager = DirectoryManager(self.security_manager)
        self.config_manager = ConfigManager()
        self.git_manager = GitManager()
        self.resource_manager = ResourceManager()

        logger.info("SimplifiedInstaller initialized for: %s", config.project_path)

    def install(self, progress_callback: Optional[Callable[[str, int, int], None]] = None) -> InstallationResult:
        """
        Execute MoAI-ADK project installation

        Args:
            progress_callback: Progress callback function

        Returns:
            InstallationResult: Installation result
        """
        files_created: List[str] = []
        errors: List[str] = []

        try:
            # Step 1: 프로젝트 디렉토리 생성
            self.progress.update_progress("📁 프로젝트 디렉토리 준비 중...", progress_callback)
            self.directory_manager.create_project_directory(self.config)
            files_created.append(str(self.config.project_path))

            # Step 2: 기본 디렉토리 구조 생성
            self.progress.update_progress("🏗️ 디렉토리 구조 생성 중...", progress_callback)
            directories = self._create_basic_structure()
            files_created.extend([str(d) for d in directories])

            # Step 3: Claude Code 리소스 복사
            self.progress.update_progress("🤖 Claude Code 리소스 복사 중...", progress_callback)
            claude_files = self._install_claude_resources()
            files_created.extend([str(f) for f in claude_files])

            # Step 4: MoAI 리소스 복사
            self.progress.update_progress("🗿 MoAI 리소스 복사 중...", progress_callback)
            moai_files = self._install_moai_resources()
            files_created.extend([str(f) for f in moai_files])

            # Step 5: GitHub 워크플로우 복사 (선택적)
            if self.config.include_github:
                self.progress.update_progress("🐙 GitHub 워크플로우 설정 중...", progress_callback)
                github_files = self._install_github_workflows()
                files_created.extend([str(f) for f in github_files])

            # Step 6: 프로젝트 메모리 파일 생성
            self.progress.update_progress("📝 프로젝트 메모리 생성 중...", progress_callback)
            if self._install_project_memory():
                files_created.append(str(self.config.project_path / 'CLAUDE.md'))

            # Step 7: 설정 파일 생성
            self.progress.update_progress("⚙️ 설정 파일 생성 중...", progress_callback)
            config_files = self._create_configuration_files()
            files_created.extend([str(f) for f in config_files])

            # Step 8: Git 저장소 초기화 (선택적)
            if self.config.initialize_git:
                self.progress.update_progress("📦 Git 저장소 초기화 중...", progress_callback)
                git_files = self._initialize_git_repository()
                files_created.extend([str(f) for f in git_files])

            # Step 9: 설치 검증
            self.progress.update_progress("✅ 설치 검증 중...", progress_callback)
            if not self._verify_installation():
                errors.append("Installation verification failed")

            self.progress.update_progress("🎉 설치 완료!", progress_callback)

            return InstallationResult(
                success=len(errors) == 0,
                project_path=str(self.config.project_path),
                files_created=files_created,
                errors=errors,
                next_steps=self._generate_next_steps(),
                config=self.config
            )

        except Exception as e:
            error_msg = f"Installation failed: {e}"
            logger.error(error_msg)
            errors.append(error_msg)

            return InstallationResult(
                success=False,
                project_path=str(self.config.project_path),
                files_created=files_created,
                errors=errors,
                next_steps=["Fix the errors above and retry installation"],
                config=self.config
            )

    def _create_basic_structure(self) -> List[Path]:
        """기본 디렉토리 구조 생성"""
        directories = [
            # Claude Code 표준 디렉토리
            self.config.project_path / ".claude",
            self.config.project_path / ".claude" / "logs",

            # MoAI 문서 시스템
            self.config.project_path / ".moai",
            self.config.project_path / ".moai" / "steering",
            self.config.project_path / ".moai" / "specs",
            self.config.project_path / ".moai" / "memory" / "decisions",
            self.config.project_path / ".moai" / "indexes",
            self.config.project_path / ".moai" / "reports",
        ]

        created_dirs = []
        for directory in directories:
            try:
                if not self.security_manager.validate_path_safety(directory, self.config.project_path):
                    logger.error("Security validation failed for directory: %s", directory)
                    continue

                directory.mkdir(parents=True, exist_ok=True)
                created_dirs.append(directory)
                logger.debug("Created directory: %s", directory)

            except Exception as e:
                logger.error("Failed to create directory %s: %s", directory, e)

        return created_dirs

    def _install_claude_resources(self) -> List[Path]:
        """Claude Code 리소스 설치"""
        try:
            return self.resource_manager.copy_claude_resources(
                self.config.project_path,
                overwrite=self.config.force_overwrite
            )
        except Exception as e:
            logger.error("Failed to install Claude resources: %s", e)
            return []

    def _install_moai_resources(self) -> List[Path]:
        """MoAI 리소스 설치"""
        try:
            return self.resource_manager.copy_moai_resources(
                self.config.project_path,
                overwrite=self.config.force_overwrite
            )
        except Exception as e:
            logger.error("Failed to install MoAI resources: %s", e)
            return []

    def _install_github_workflows(self) -> List[Path]:
        """GitHub 워크플로우 설치"""
        try:
            return self.resource_manager.copy_github_resources(
                self.config.project_path,
                overwrite=self.config.force_overwrite
            )
        except Exception as e:
            logger.error("Failed to install GitHub workflows: %s", e)
            return []

    def _install_project_memory(self) -> bool:
        """프로젝트 메모리 파일 생성"""
        try:
            return self.resource_manager.copy_project_memory(
                self.config.project_path,
                overwrite=self.config.force_overwrite
            )
        except Exception as e:
            logger.error("Failed to create project memory: %s", e)
            return False

    def _create_configuration_files(self) -> List[Path]:
        """설정 파일 생성"""
        config_files = []

        try:
            # Claude Code 설정
            claude_settings = self.config.project_path / ".claude" / "settings.json"
            if self.config_manager.create_claude_settings(claude_settings, self.config):
                config_files.append(claude_settings)

            # MoAI 설정
            moai_config = self.config.project_path / ".moai" / "config.json"
            if self.config_manager.create_moai_config(moai_config, self.config):
                config_files.append(moai_config)

            logger.info("Created %d configuration files", len(config_files))
            return config_files

        except Exception as e:
            logger.error("Failed to create configuration files: %s", e)
            return []

    def _initialize_git_repository(self) -> List[Path]:
        """Git 저장소 초기화"""
        git_files = []

        try:
            if self.git_manager.initialize_repository(self.config.project_path):
                git_files.append(self.config.project_path / ".git")

                # .gitignore 생성
                gitignore_path = self.config.project_path / ".gitignore"
                if self.git_manager.create_gitignore(gitignore_path):
                    git_files.append(gitignore_path)

            logger.info("Initialized Git repository with %d files", len(git_files))
            return git_files

        except Exception as e:
            logger.error("Failed to initialize Git repository: %s", e)
            return []

    def _verify_installation(self) -> bool:
        """설치 검증"""
        try:
            return self.resource_manager.validate_project_resources(self.config.project_path)
        except Exception as e:
            logger.error("Installation verification failed: %s", e)
            return False

    def _generate_next_steps(self) -> List[str]:
        """다음 단계 안내 생성"""
        next_steps = [
            "🎯 프로젝트가 성공적으로 설치되었습니다!",
            "",
            "다음 단계:",
            "1. 프로젝트 디렉토리로 이동:",
            f"   cd {self.config.project_path}",
            "",
            "2. 프로젝트 초기화:",
            "   /moai:1-project init",
            "",
            "3. 첫 번째 기능 개발:",
            "   /moai:2-spec <feature-name> \"기능 설명\"",
            "",
            "4. 도움말:",
            "   /moai:help",
        ]

        # Claude Code 관련 안내
        if hasattr(self.config, 'claude_version'):
            next_steps.extend([
                "",
                "🤖 Claude Code 통합:",
                "- 모든 MoAI 명령어가 슬래시 명령어로 설치되었습니다",
                "- /moai:1-project부터 시작하세요",
            ])

        return next_steps


# 하위 호환성을 위한 별칭
ProjectInstaller = SimplifiedInstaller