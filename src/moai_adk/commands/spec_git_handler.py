"""
@TASK:GIT-HANDLER-001 SpecGitHandler Module
@REQ:TRUST-COMPLIANCE-001 → @DESIGN:MODULE-SPLIT-001 → @TASK:GIT-HANDLER-001

Git workflow handling module following TRUST principles:
- T: Test-driven Git operations
- R: Readable Git workflow logic
- U: Unified Git workflow responsibility
- S: Secure Git operations with error handling
- T: Trackable Git workflow execution
"""

import logging
from pathlib import Path

from ..core.exceptions import GitLockedException
from ..core.git_strategy import GitStrategyBase, PersonalGitStrategy, TeamGitStrategy

# Logging setup
logger = logging.getLogger(__name__)


class SpecGitHandler:
    """
    @TASK:SPEC-GIT-WORKFLOW-001 Git workflow management for SPEC command

    TRUST principles applied:
    - T: Testable Git strategy pattern
    - R: Clear Git workflow logic
    - U: Single responsibility (Git operations only)
    - S: Secure error handling for Git locks
    - T: Trackable Git workflow execution
    """

    def __init__(self, project_dir: Path, config=None):
        """Initialize SpecGitHandler

        Args:
            project_dir: Project directory path
            config: Configuration manager or dict

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

        # Git strategy instance (lazy initialization)
        self._git_strategy: GitStrategyBase | None = None

        logger.debug(f"SpecGitHandler initialized: {self.project_dir}")

    def _get_git_strategy(self) -> GitStrategyBase:
        """Get Git strategy instance based on current mode

        Returns:
            Appropriate Git strategy instance
        """
        if self._git_strategy is None:
            mode = self._get_current_mode()

            if mode == "team":
                self._git_strategy = TeamGitStrategy(self.project_dir, self.config)
            else:
                self._git_strategy = PersonalGitStrategy(self.project_dir, self.config)

            logger.debug(
                f"Git strategy created: {mode} -> {self._git_strategy.__class__.__name__}"
            )

        return self._git_strategy

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

    def should_create_branch(self) -> bool:
        """Determine if branch should be created

        Returns:
            True if branch should be created, False otherwise
        """
        mode = self._get_current_mode()
        should_create = mode == "team"

        logger.debug(f"Branch creation decision: {should_create} (mode: {mode})")
        return should_create

    def execute_git_workflow(self, spec_name: str):
        """Execute Git workflow for SPEC creation

        Args:
            spec_name: Name of the specification for branch naming

        Raises:
            GitLockedException: Re-raised for graceful handling
            Exception: Other Git-related exceptions
        """
        try:
            git_strategy = self._get_git_strategy()

            # Create branch name from spec name
            branch_name = f"spec-{spec_name.lower()}"

            # Execute Git workflow within context
            with git_strategy.work_context(branch_name):
                logger.info(f"Git workflow executed: {spec_name}")

        except GitLockedException:
            logger.warning("Git locked - skipping branch operations")
            # Re-raise for caller to handle gracefully
            raise
        except Exception as e:
            logger.error(f"Git workflow execution failed: {e}")
            raise

    def get_handler_status(self) -> dict:
        """Get handler status information

        Returns:
            Dictionary containing handler status
        """
        git_strategy = self._get_git_strategy()

        return {
            "project_dir": str(self.project_dir),
            "mode": self._get_current_mode(),
            "git_strategy": git_strategy.__class__.__name__,
            "repository_status": git_strategy.get_repository_status(),
        }