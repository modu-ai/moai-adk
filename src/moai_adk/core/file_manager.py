"""
File management utilities for MoAI-ADK.

Handles file copying, template rendering, and file system operations
with security validation and error handling.
"""

import shutil
from pathlib import Path
from string import Template
from typing import List, Dict, Any

from ..utils.logger import get_logger
from .security import SecurityManager

logger = get_logger(__name__)


class FileManager:
    """Manages file operations for MoAI-ADK installation."""

    def __init__(self, template_dir: Path, security_manager: SecurityManager = None):
        """
        Initialize file manager.

        Args:
            template_dir: Directory containing template files
            security_manager: Security manager instance for validation
        """
        self.template_dir = template_dir
        self.security_manager = security_manager or SecurityManager()

    def copy_template_files(
        self,
        source_dir: Path,
        target_dir: Path,
        pattern: str,
        preserve_permissions: bool = False
    ) -> List[Path]:
        """
        Copy template files matching pattern with security validation.

        Args:
            source_dir: Source directory containing templates
            target_dir: Target directory for copied files
            pattern: Glob pattern to match files
            preserve_permissions: Whether to preserve file permissions

        Returns:
            List[Path]: List of successfully copied files
        """
        if not source_dir.exists():
            logger.warning("Source directory not found: %s", source_dir)
            return []

        created_files = []
        target_dir.mkdir(parents=True, exist_ok=True)

        try:
            for source_file in source_dir.glob(pattern):
                if not source_file.is_file():
                    continue

                # Security validation
                target_file = target_dir / source_file.name
                if not self.security_manager.validate_file_creation(target_file, target_dir):
                    logger.error("Security validation failed for file: %s", target_file)
                    continue

                # Copy file
                shutil.copy2(source_file, target_file)

                # Set permissions if specified
                if preserve_permissions and source_file.stat().st_mode:
                    target_file.chmod(source_file.stat().st_mode)

                created_files.append(target_file)
                logger.debug("Copied template file: %s -> %s", source_file, target_file)

        except Exception as e:
            logger.error("Error copying template files: %s", e)

        return created_files

    def render_template_file(self, template_path: Path, context: Dict[str, Any]) -> str:
        """
        Render a template file with context variables.

        Args:
            template_path: Path to template file
            context: Variables to substitute in template

        Returns:
            str: Rendered template content
        """
        try:
            with open(template_path, "r", encoding="utf-8") as f:
                template_content = f.read()

            template = Template(template_content)
            return template.safe_substitute(context)

        except FileNotFoundError:
            logger.error("Template file not found: %s", template_path)
            return ""
        except Exception as e:
            logger.error("Error rendering template %s: %s", template_path, e)
            return template_content if 'template_content' in locals() else ""

    def copy_and_render_template(
        self,
        source_path: Path,
        target_path: Path,
        context: Dict[str, Any],
        create_dirs: bool = True
    ) -> bool:
        """
        Copy and render a single template file.

        Args:
            source_path: Source template file
            target_path: Target file path
            context: Template variables
            create_dirs: Whether to create parent directories

        Returns:
            bool: True if successful
        """
        try:
            # Validate source file exists
            if not source_path.exists():
                logger.error("Template source not found: %s", source_path)
                return False

            # Security validation
            if not self.security_manager.validate_file_creation(
                target_path, target_path.parent.parent
            ):
                logger.error("Security validation failed for: %s", target_path)
                return False

            # Create target directory if needed
            if create_dirs:
                target_path.parent.mkdir(parents=True, exist_ok=True)

            # Render template
            rendered_content = self.render_template_file(source_path, context)

            # Write rendered content
            with open(target_path, "w", encoding="utf-8") as f:
                f.write(rendered_content)

            logger.debug("Rendered template: %s -> %s", source_path, target_path)
            return True

        except Exception as e:
            logger.error("Error copying and rendering template: %s", e)
            return False

    def copy_hook_scripts(self, target_dir: Path) -> List[Path]:
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
                # Security validation
                if not self.security_manager.validate_file_creation(target_file, target_dir):
                    logger.error("Security validation failed for hook: %s", hook_name)
                    continue

                shutil.copy2(source_file, target_file)

                # Set executable permissions for Python files
                if hook_name.endswith('.py'):
                    target_file.chmod(0o755)

                hook_files.append(target_file)
                logger.debug("Installed hook: %s", hook_name)

        return hook_files

    def copy_verification_scripts(self, target_dir: Path) -> List[Path]:
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
            "run-tests.sh",
        ]

        for script_name in verification_scripts:
            source_file = source_dir / script_name
            target_file = target_dir / script_name

            if source_file.exists():
                # Security validation
                if not self.security_manager.validate_file_creation(target_file, target_dir):
                    logger.error("Security validation failed for script: %s", script_name)
                    continue

                shutil.copy2(source_file, target_file)

                # Set executable permissions
                target_file.chmod(0o755)
                script_files.append(target_file)
                logger.debug("Installed verification script: %s", script_name)

        return script_files

    def install_output_styles(
        self,
        target_dir: Path,
        context: Dict[str, Any]
    ) -> List[Path]:
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
            logger.warning("Output styles template directory not found: %s", template_styles_dir)
            return installed_files

        # Style files to install
        style_files = [
            "expert.md",      # 전문가용 - 간결하고 효율적
            "beginner.md",    # 초보자용 - 상세한 설명
            "study.md",       # 심화학습용 - 깊이 있는 원리 설명
            "mentor.md",      # 멘토링용 - 1:1 페어 프로그래밍 시뮬레이션
            "audit.md",       # 품질 검증용 - 지속적 감사
        ]

        for style_file in style_files:
            source_file = template_styles_dir / style_file
            if not source_file.exists():
                logger.warning("Style template not found: %s", source_file)
                continue

            target_file = target_dir / style_file

            # Security validation
            if not self.security_manager.validate_file_creation(target_file, target_dir):
                logger.error("Security validation failed for style: %s", style_file)
                continue

            # Copy and render template
            if self.copy_and_render_template(source_file, target_file, context):
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

            with open(gitignore_path, 'w', encoding='utf-8') as f:
                f.write(gitignore_content)

            logger.info("Created .gitignore file: %s", gitignore_path)
            return True

        except Exception as e:
            logger.error("Failed to create .gitignore: %s", e)
            return False
