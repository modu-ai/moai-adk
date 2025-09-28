"""
@FEATURE:SCRIPT-INSTALL-001 Script installation utilities for MoAI-ADK

Handles installation of hook scripts, verification scripts, and output styles
with specialized logic for different script types.
"""

from pathlib import Path
from typing import Any

from ..utils.logger import get_logger
from .file_copier import FileCopier
from .security import SecurityManager

logger = get_logger(__name__)


class ScriptInstaller:
    """@TASK:SCRIPT-INSTALLER-001 Handles installation of MoAI scripts and styles."""

    def __init__(
        self, template_dir: Path, security_manager: SecurityManager = None
    ):
        """
        Initialize script installer.

        Args:
            template_dir: Directory containing template files
            security_manager: Security manager instance for validation
        """
        self.template_dir = template_dir
        self.security_manager = security_manager or SecurityManager()
        self.file_copier = FileCopier(security_manager)

    def copy_hook_scripts(self, target_dir: Path) -> list[Path]:
        """
        Copy MoAI Hook scripts to target directory.

        Args:
            target_dir: Directory to copy hooks to

        Returns:
            List[Path]: List of copied hook files
        """
        hook_files = []
        source_dir = self.template_dir / ".claude" / "hooks" / "moai"

        # Ensure target directory exists
        target_dir.mkdir(parents=True, exist_ok=True)

        # Core hook files
        core_hooks = [
            "pre_write_guard.py",
            "policy_block.py",
            "tag_validator.py",
            "steering_guard.py",
            "run_tests_and_report.py",
            "session_start_notice.py",
            "language_detector.py",
        ]

        for hook_name in core_hooks:
            source_file = source_dir / hook_name
            target_file = target_dir / hook_name

            if source_file.exists():
                # Use FileCopier with executable permissions for Python files
                mode = 0o755 if hook_name.endswith(".py") else None
                if self.file_copier.copy_file_with_permissions(
                    source_file, target_file, mode
                ):
                    hook_files.append(target_file)
                    logger.debug("Installed hook: %s", hook_name)

        return hook_files

    def copy_verification_scripts(self, target_dir: Path) -> list[Path]:
        """
        Copy MoAI verification scripts to target directory.

        Args:
            target_dir: Directory to copy scripts to

        Returns:
            List[Path]: List of copied script files
        """
        script_files = []
        source_dir = self.template_dir / "scripts"

        # Ensure target directory exists
        target_dir.mkdir(parents=True, exist_ok=True)

        # Verification scripts
        verification_scripts = [
            "validate_stage.py",
            "check-secrets.py",
            "check-licenses.py",
            "check-traceability.py",
            "check_coverage.py",
            "validate_tags.py",
            "check_constitution.py",
            "repair_tags.py",
            # Note: run-tests.sh removed - use Python test_runner.py instead
        ]

        for script_name in verification_scripts:
            source_file = source_dir / script_name
            target_file = target_dir / script_name

            if source_file.exists():
                # All verification scripts need executable permissions
                if self.file_copier.copy_file_with_permissions(
                    source_file, target_file, 0o755
                ):
                    script_files.append(target_file)
                    logger.debug("Installed verification script: %s", script_name)

        return script_files

    def install_output_styles(
        self, target_dir: Path, context: dict[str, Any]
    ) -> list[Path]:
        """
        Install MoAI-ADK output styles with template rendering.

        Args:
            target_dir: Directory to install styles to
            context: Template context for rendering

        Returns:
            List[Path]: List of installed style files
        """
        target_dir.mkdir(parents=True, exist_ok=True)
        installed_files = []
        template_styles_dir = self.template_dir / ".claude" / "output-styles"

        if not template_styles_dir.exists():
            logger.warning(
                "Output styles template directory not found: %s", template_styles_dir
            )
            return installed_files

        # Style files to install
        style_files = [
            "expert.md",  # Expert mode - concise and efficient
            "beginner.md",  # Beginner mode - detailed explanations
            "study.md",  # Study mode - in-depth principle explanations
            "mentor.md",  # Mentoring mode - 1:1 pair programming simulation
            "audit.md",  # Quality audit mode - continuous auditing
        ]

        for style_file in style_files:
            source_file = template_styles_dir / style_file
            if not source_file.exists():
                logger.warning("Style template not found: %s", source_file)
                continue

            target_file = target_dir / style_file

            # Copy and render template using FileCopier
            if self.file_copier.copy_and_render_template(
                source_file, target_file, context
            ):
                installed_files.append(target_file)

        return installed_files

    def create_gitignore(self, gitignore_path: Path) -> bool:
        """
        Create a comprehensive .gitignore file.

        Args:
            gitignore_path: Path where .gitignore should be created

        Returns:
            bool: True if successful
        """
        gitignore_content = """# MoAI-ADK specific
.moai/reports/
.moai/indexes/*.json

# Python
__pycache__/
*.py[cod]
*$py.class
*.so
.Python
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/
pip-wheel-metadata/
share/python-wheels/
*.egg-info/
.installed.cfg
*.egg
MANIFEST

# Virtual environments
.env
.venv
env/
venv/
ENV/
env.bak/
venv.bak/

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Logs
*.log
logs/

# Node.js (if applicable)
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# Rust (if applicable)
target/
Cargo.lock

# Go (if applicable)
vendor/
"""

        try:
            # Security validation
            if not self.security_manager.validate_file_creation(
                gitignore_path, gitignore_path.parent
            ):
                logger.error("Security validation failed for .gitignore")
                return False

            with open(gitignore_path, "w", encoding="utf-8") as f:
                f.write(gitignore_content)

            logger.info("Created .gitignore file: %s", gitignore_path)
            return True

        except Exception as e:
            logger.error("Failed to create .gitignore: %s", e)
            return False