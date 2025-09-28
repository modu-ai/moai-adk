"""
@FEATURE:TEMPLATE-PROCESSOR-001 Template Processing for MoAI-ADK

@TASK:TEMPLATE-PROCESSOR-001 Handles template variable substitution and file type detection
@DESIGN:MODULE-SPLIT-001 Extracted from file_operations.py for TRUST principle compliance
"""

import logging
import shutil
from pathlib import Path

from .template_manager import TemplateManager

logger = logging.getLogger(__name__)


class TemplateProcessor:
    """Template processing manager for variable substitution and file type detection."""

    def __init__(self):
        """Initialize template processor."""
        self.template_manager = TemplateManager()

    def copy_directory_with_substitution(
        self,
        source_dir: Path,
        target_dir: Path,
        context: dict[str, str],
        overwrite: bool,
        exclude_subdirs: list[str] | None = None,
    ) -> None:
        """
        Recursively copy directory with template substitution for text files.

        Args:
            source_dir: Source directory path
            target_dir: Target directory path
            context: Context variables for template substitution
            overwrite: Whether to overwrite existing files
            exclude_subdirs: Subdirectories to exclude
        """
        # Create target directory
        target_dir.mkdir(parents=True, exist_ok=True)

        for source_item in source_dir.iterdir():
            target_item = target_dir / source_item.name

            # Skip excluded subdirectories
            if exclude_subdirs and source_item.is_dir() and source_item.name in exclude_subdirs:
                logger.debug(f"Excluding subdirectory: {source_item}")
                continue

            if source_item.is_dir():
                # Recursively copy subdirectory
                self.copy_directory_with_substitution(
                    source_item, target_item, context, overwrite, exclude_subdirs
                )
            else:
                # Copy file with potential substitution
                if not target_item.exists() or overwrite:
                    self.copy_file_with_substitution(source_item, target_item, context)
                else:
                    logger.debug(f"File exists, skipping: {target_item}")

    def copy_file_with_substitution(
        self, source_file: Path, target_file: Path, context: dict[str, str]
    ) -> None:
        """
        Copy file with template substitution if it's a text file.

        Args:
            source_file: Source file path
            target_file: Target file path
            context: Context variables for template substitution
        """
        try:
            # Check if file should be processed as template
            if self.should_process_as_template(source_file):
                # Read, substitute, and write
                content = source_file.read_text(encoding="utf-8")
                processed_content = self.template_manager.unified_substitute_template_variables(
                    content, context
                )
                target_file.write_text(processed_content, encoding="utf-8")
                logger.debug(f"Template substitution applied to: {target_file}")
            else:
                # Binary copy for non-text files
                shutil.copy2(source_file, target_file)
                logger.debug(f"Binary copy: {source_file} -> {target_file}")

        except UnicodeDecodeError:
            # Fall back to binary copy for files that can't be decoded as text
            shutil.copy2(source_file, target_file)
            logger.debug(f"Binary copy (decode failed): {source_file} -> {target_file}")
        except Exception as e:
            logger.warning(f"Failed to copy {source_file}: {e}")
            # Fall back to binary copy
            shutil.copy2(source_file, target_file)

    def should_process_as_template(self, file_path: Path) -> bool:
        """
        Determine if a file should be processed as a template.

        Args:
            file_path: Path to check

        Returns:
            bool: True if file should be templated
        """
        # Skip binary file extensions
        binary_extensions = {
            '.exe', '.dll', '.so', '.dylib', '.bin', '.dat', '.db', '.sqlite',
            '.png', '.jpg', '.jpeg', '.gif', '.bmp', '.ico', '.svg',
            '.mp3', '.mp4', '.avi', '.mov', '.wav', '.pdf', '.zip', '.tar', '.gz',
            '.class', '.jar', '.pyc', '.pyo', '.whl', '.egg',
        }

        if file_path.suffix.lower() in binary_extensions:
            return False

        # Process text file extensions
        text_extensions = {
            '.md', '.json', '.yml', '.yaml', '.txt', '.py', '.js', '.ts', '.html',
            '.css', '.xml', '.ini', '.cfg', '.conf', '.sh', '.bat', '.ps1',
            '.sql', '.env', '.gitignore', '.dockerignore', '.editorconfig',
        }

        # Process if extension is known text or if no extension
        return file_path.suffix.lower() in text_extensions or not file_path.suffix