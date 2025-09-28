"""
@TASK:SPEC-REFACTOR-001 Refactored SpecCommand Orchestrator
@REQ:TRUST-COMPLIANCE-001 → @DESIGN:MODULE-SPLIT-001 → @TASK:SPEC-REFACTOR-001

Refactored SPEC command orchestrator following TRUST principles:
- T: Test-driven orchestration design
- R: Readable command flow
- U: Unified orchestration responsibility
- S: Secure module integration
- T: Trackable command execution
"""

import logging
from pathlib import Path

from .spec_validator import SpecValidator
from .spec_file_generator import SpecFileGenerator
from .spec_git_handler import SpecGitHandler
from ..core.exceptions import GitLockedException

# Logging setup
logger = logging.getLogger(__name__)


class SpecCommand:
    """
    @TASK:SPEC-MAIN-001 Enhanced SPEC command orchestrator using modular components

    TRUST principles applied:
    - T: Testable orchestration structure
    - R: Clear command execution flow
    - U: Unified command orchestration (single responsibility)
    - S: Secure module integration with error handling
    - T: Trackable execution through modules
    """

    def __init__(self, project_dir: Path, config=None, skip_branch: bool = False,
                 validator=None, file_generator=None, git_handler=None):
        """Initialize SpecCommand orchestrator

        Args:
            project_dir: Project directory path
            config: Configuration manager instance or dict
            skip_branch: Skip branch creation option
            validator: SpecValidator instance (for dependency injection)
            file_generator: SpecFileGenerator instance (for dependency injection)
            git_handler: SpecGitHandler instance (for dependency injection)

        Raises:
            ValueError: If project_dir is invalid
        """
        # Input validation
        if not isinstance(project_dir, Path):
            raise ValueError(f"project_dir must be a Path object: {type(project_dir)}")

        if not project_dir.exists():
            raise ValueError(f"Project directory does not exist: {project_dir}")

        self.project_dir = project_dir.resolve()
        self.config = config
        self.skip_branch = skip_branch

        # Initialize modules (dependency injection pattern)
        self._validator = validator or SpecValidator()
        self._file_generator = file_generator or SpecFileGenerator(self.project_dir, self._get_current_mode())
        self._git_handler = git_handler or SpecGitHandler(self.project_dir, self.config)

        logger.debug(
            f"SpecCommand initialized: {self.project_dir}, skip_branch={skip_branch}"
        )

    def _get_current_mode(self) -> str:
        """Get current operation mode

        Returns:
            Current mode ('personal' or 'team')
        """
        if self.config:
            try:
                # ConfigManager object case
                if hasattr(self.config, "get_mode"):
                    return self.config.get_mode()
                # Dictionary case
                elif isinstance(self.config, dict):
                    return self.config.get("mode", "personal")
            except (AttributeError, KeyError):
                pass

        return "personal"

    def execute(
        self, spec_name: str, description: str, skip_branch: bool | None = None
    ):
        """Execute SPEC command using modular components

        Args:
            spec_name: Specification name
            description: Specification description
            skip_branch: Branch creation skip option (None uses default)

        Raises:
            ValueError: Invalid input from modules
            GitLockedException: Git operation locked (handled gracefully)
        """
        # Update skip option if provided
        if skip_branch is not None:
            self.skip_branch = skip_branch

        # Execution logging
        self._log_execution_start(spec_name, description)

        try:
            # Step 1: Validate inputs
            validated_spec_name = self._validator.validate_spec_name(spec_name)
            validated_description = self._validator.validate_description(description)

            # Step 2: Create SPEC file
            spec_file_path = self._file_generator.create_spec_file(
                validated_spec_name, validated_description
            )

            # Step 3: Execute Git workflow (if needed)
            if not self.skip_branch and self._git_handler.should_create_branch():
                try:
                    self._git_handler.execute_git_workflow(validated_spec_name)
                except GitLockedException:
                    logger.warning("Git locked - SPEC file created without branch")
                    # Continue execution, file was already created

            # Success logging
            self._log_execution_success(validated_spec_name, spec_file_path)

        except Exception as e:
            self._log_execution_error(spec_name, e)
            raise

    def execute_with_mode(
        self, mode: str, spec_name: str = "test-spec", description: str = "테스트 명세"
    ):
        """Execute with specific mode strategy

        Args:
            mode: Execution mode (personal/team)
            spec_name: Specification name
            description: Specification description

        Raises:
            ValueError: Unsupported mode
        """
        # Mode validation
        if mode not in ["personal", "team"]:
            raise ValueError(f"지원하지 않는 모드입니다: {mode}")

        logger.info(f"모드별 실행: {mode}")

        if mode == "personal" and self.skip_branch:
            # Personal mode with branch skip
            self.execute(spec_name, description, skip_branch=True)
        elif mode == "team":
            # Team mode always creates branches
            self.execute(spec_name, description, skip_branch=False)
        else:
            # Default execution
            self.execute(spec_name, description)

    def get_command_status(self) -> dict:
        """Get command status information

        Returns:
            Dictionary containing command status
        """
        git_handler_status = self._git_handler.get_handler_status()
        specs_dir = self.project_dir / ".moai" / "specs"

        return {
            "project_dir": str(self.project_dir),
            "skip_branch": self.skip_branch,
            "specs_dir_exists": specs_dir.exists(),
            "git_handler_status": git_handler_status,
        }

    def _log_execution_start(self, spec_name: str, description: str):
        """Log execution start"""
        logger.info(
            "SPEC command execution started",
            extra={
                "command": "spec",
                "spec_name": spec_name,
                "description": description[:50] + "..."
                if len(description) > 50
                else description,
                "mode": self._get_current_mode(),
                "skip_branch": self.skip_branch,
            },
        )

    def _log_execution_success(self, spec_name: str, spec_file_path: Path):
        """Log execution success"""
        logger.info(f"SPEC command execution completed: {spec_name} -> {spec_file_path}")

    def _log_execution_error(self, spec_name: str, error: Exception):
        """Log execution error"""
        logger.error(f"SPEC command execution failed: {spec_name}, error: {error}")