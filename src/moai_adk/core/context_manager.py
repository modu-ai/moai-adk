"""
CLAUDE.md Import Processor

Processes @path/to/file imports in CLAUDE.md files.

Based on Claude Code official documentation:
- Import syntax: @path/to/file or @~/absolute/path
- Recursive imports (max 5-depth)
- Ignored in code blocks/spans
- Both relative and absolute paths supported
"""

from __future__ import annotations

import json
import logging
import os
import re
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

# Import pattern matches @path but not in code
IMPORT_PATTERN = r"@([^\s\])`>@]+)"
# Matches code blocks
CODE_BLOCK_PATTERN = r"```[\s\S]*?```"
# Matches inline code
INLINE_CODE_PATTERN = r"`[^`]+`"

logger = logging.getLogger(__name__)


class ClaudeMDImporter:
    """Process @path imports in CLAUDE.md files."""

    MAX_RECURSION_DEPTH = 5

    def __init__(self, base_path: Path):
        """
        Initialize the importer with a base path for resolving relative imports.

        Args:
            base_path: Base directory for resolving relative import paths
        """
        self.base_path = base_path.resolve()
        self.import_cache: dict[str, str] = {}
        self.recursion_stack: List[str] = []

    def process_imports(self, content: str, depth: int = 0) -> Tuple[str, List[str]]:
        """
        Process @path imports in CLAUDE.md content.

        Args:
            content: CLAUDE.md content
            depth: Current recursion depth

        Returns:
            Tuple of (processed_content, list_of_imported_files)
        """
        if depth >= self.MAX_RECURSION_DEPTH:
            logger.warning(f"Max recursion depth {self.MAX_RECURSION_DEPTH} reached")
            return content, []

        # Track recursion to prevent cycles
        self.recursion_stack.append(str(self.base_path))

        imported_files = []

        # Remove code blocks (imports ignored in code)
        code_blocks = {}
        content = self._extract_and_replace_code_blocks(content, code_blocks)

        # Process imports
        lines = content.split("\n")
        processed_lines = []

        for line in lines:
            processed_line, imported = self._process_line(line, depth)
            processed_lines.append(processed_line)
            imported_files.extend(imported)

        # Restore code blocks
        content = "\n".join(processed_lines)
        content = self._restore_code_blocks(content, code_blocks)

        self.recursion_stack.pop()
        return content, imported_files

    def _extract_and_replace_code_blocks(self, content: str, store: dict) -> str:
        """
        Extract code blocks and replace with placeholders.

        Args:
            content: Content to process
            store: Dictionary to store extracted blocks

        Returns:
            Content with code blocks replaced by placeholders
        """
        idx = 0
        for match in re.finditer(CODE_BLOCK_PATTERN, content):
            placeholder = f"__CODE_BLOCK_{idx}__"
            store[placeholder] = match.group(0)
            content = content[: match.start()] + placeholder + content[match.end() :]
            idx += 1
        return content

    def _restore_code_blocks(self, content: str, store: dict) -> str:
        """
        Restore code blocks from placeholders.

        Args:
            content: Content with placeholders
            store: Dictionary with stored code blocks

        Returns:
            Content with code blocks restored
        """
        for placeholder, code_block in store.items():
            content = content.replace(placeholder, code_block)
        return content

    def _process_line(self, line: str, depth: int) -> Tuple[str, List[str]]:
        """
        Process a single line for imports.

        Args:
            line: Line to process
            depth: Current recursion depth

        Returns:
            Tuple of (processed_line, list_of_imported_files)
        """
        imported_files = []
        processed_line = line

        for match in re.finditer(IMPORT_PATTERN, line):
            import_path = match.group(1)
            # Skip if in inline code (already replaced code blocks)
            if "`" in import_path:
                continue

            imported_content, import_files = self._load_import(import_path, depth)
            if imported_content:
                processed_line = processed_line.replace(match.group(0), imported_content)
                imported_files.extend(import_files)

        return processed_line, imported_files

    def _load_import(self, import_path: str, depth: int) -> Tuple[str, List[str]]:
        """
        Load and process an imported file.

        Args:
            import_path: Path to import (relative or absolute with ~/)
            depth: Current recursion depth

        Returns:
            Tuple of (imported_content, list_of_imported_files)
        """
        # Resolve path
        if import_path.startswith("~/"):
            # User home directory import
            resolved_path = Path(import_path).expanduser()
        else:
            # Relative path from base_path
            resolved_path = self.base_path / import_path

        # Check cache
        cache_key = str(resolved_path)
        if cache_key in self.import_cache:
            return self.import_cache[cache_key], []

        # Check for circular imports
        if cache_key in self.recursion_stack:
            logger.warning(f"Circular import detected: {import_path}")
            return f"[Circular import detected: {import_path}]", []

        # Read file
        if not resolved_path.exists():
            logger.warning(f"Import not found: {import_path}")
            return f"[Import not found: {import_path}]", []

        try:
            content = resolved_path.read_text(encoding="utf-8", errors="replace")

            # Recursive processing
            importer = ClaudeMDImporter(resolved_path.parent)
            importer.recursion_stack = self.recursion_stack.copy()
            processed_content, nested_imports = importer.process_imports(content, depth + 1)

            # Cache result
            self.import_cache[cache_key] = processed_content

            return processed_content, [cache_key] + nested_imports

        except Exception as e:
            logger.error(f"Import error: {import_path} - {e}")
            return f"[Import error: {import_path} - {e}]", []


def process_claude_md_imports(content: str, base_path: Path) -> Tuple[str, List[str]]:
    """
    Process @path imports in CLAUDE.md content.

    Convenience function that creates an importer and processes the content.

    Args:
        content: CLAUDE.md content with @path imports
        base_path: Base directory for resolving relative imports

    Returns:
        Tuple of (processed_content, list_of_imported_files)
    """
    importer = ClaudeMDImporter(base_path)
    return importer.process_imports(content)


# ============================================================================
# Context Manager for Phase Results (Alfred Workflow)
# ============================================================================


class ContextManager:
    """
    Manages context for Alfred workflow phases.

    Provides utilities for storing and retrieving phase results,
    validating paths, and managing template variables.
    """

    def __init__(self, project_root: str):
        """
        Initialize the context manager.

        Args:
            project_root: Root directory of the project
        """
        self.project_root = os.path.abspath(project_root)
        self.memory_dir = os.path.join(self.project_root, ".moai", "memory")
        os.makedirs(self.memory_dir, exist_ok=True)

    @property
    def state_dir(self) -> str:
        """
        Get the state directory path (alias for memory_dir).

        Returns:
            Path to the state directory
        """
        return self.memory_dir

    def get_state_dir(self) -> str:
        """
        Get the state directory path.

        Returns:
            Path to the state directory
        """
        return self.state_dir

    def get_phase_result_path(self, phase_name: str) -> str:
        """
        Get the file path for a phase result.

        Args:
            phase_name: Name of the phase

        Returns:
            Absolute path to the phase result file
        """
        return os.path.join(self.memory_dir, "command-state", f"{phase_name}.json")

    def load_phase_result(self, phase_name: str) -> Optional[Dict[str, Any]]:
        """
        Load a phase result from disk.

        Args:
            phase_name: Name of the phase to load

        Returns:
            Phase result dictionary or None if not found
        """
        result_path = self.get_phase_result_path(phase_name)
        if not os.path.exists(result_path):
            return None

        try:
            with open(result_path, "r", encoding="utf-8") as f:
                return json.load(f)
        except (json.JSONDecodeError, IOError):
            return None

    def save_phase_result(self, phase_name: str, result: Dict[str, Any]) -> None:
        """
        Save a phase result to disk.

        Args:
            phase_name: Name of the phase
            result: Phase result dictionary
        """
        result_path = self.get_phase_result_path(phase_name)
        os.makedirs(os.path.dirname(result_path), exist_ok=True)

        # Add timestamp if not present
        if "timestamp" not in result:
            result["timestamp"] = datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")

        with open(result_path, "w", encoding="utf-8") as f:
            json.dump(result, f, indent=2, ensure_ascii=False)


def validate_and_convert_path(path: str, base_path: Optional[str] = None) -> str:
    """
    Validate and convert a path to absolute.

    Args:
        path: Path to validate (relative or absolute)
        base_path: Base directory for relative paths (defaults to cwd)

    Returns:
        Absolute path

    Raises:
        ValueError: If path is invalid
        FileNotFoundError: If path doesn't exist and check_exists=True
    """
    if base_path is None:
        base_path = os.getcwd()

    # Convert to absolute path
    if os.path.isabs(path):
        abs_path = path
    else:
        abs_path = os.path.abspath(os.path.join(base_path, path))

    # Normalize path
    abs_path = os.path.normpath(abs_path)

    return abs_path


def _is_path_within_root(path: str, root: str) -> bool:
    """
    Check if a path is within a root directory.

    Args:
        path: Path to check
        root: Root directory

    Returns:
        True if path is within root, False otherwise
    """
    # Normalize both paths
    path = os.path.normpath(os.path.abspath(path))
    root = os.path.normpath(os.path.abspath(root))

    # Check if path starts with root
    return path.startswith(root + os.sep) or path == root


def _cleanup_temp_file(fd: Optional[int], path: Optional[str]) -> None:
    """
    Clean up a temporary file safely.

    Args:
        fd: File descriptor (may be None)
        path: Path to the file (may be None)

    Closes the file descriptor if provided and removes the file if it exists.
    """
    if fd is not None:
        try:
            os.close(fd)
        except OSError:
            pass

    if path is not None and os.path.exists(path):
        try:
            os.remove(path)
        except OSError:
            pass


def validate_no_template_vars(content: str) -> bool:
    """
    Validate that content contains no template variables.

    Args:
        content: Content to validate

    Returns:
        True if no template variables found
    """
    import re

    # Check for {{VAR}} pattern
    template_pattern = r"\{\{[A-Z_]+\}\}"
    return not re.search(template_pattern, content)


def substitute_template_variables(content: str, variables: Dict[str, str]) -> Tuple[str, List[str]]:
    """
    Substitute template variables in content.

    Args:
        content: Content with template variables
        variables: Dictionary of variable substitutions

    Returns:
        Tuple of (substituted_content, list_of_found_variables)
    """

    found_vars = []
    substituted = content

    for var_name, var_value in variables.items():
        pattern = f"{{{{{var_name}}}}}"
        if pattern in substituted:
            found_vars.append(var_name)
            substituted = substituted.replace(pattern, var_value)

    return substituted, found_vars


def load_phase_result(phase_name: str, project_root: Optional[str] = None) -> Optional[Dict[str, Any]]:
    """
    Load a phase result from disk.

    Convenience function that creates a ContextManager.

    Args:
        phase_name: Name of the phase to load
        project_root: Project root directory (defaults to cwd)

    Returns:
        Phase result dictionary or None if not found
    """
    if project_root is None:
        project_root = os.getcwd()

    manager = ContextManager(project_root)
    return manager.load_phase_result(phase_name)


def save_phase_result(
    phase_name: str,
    result: Dict[str, Any],
    project_root: Optional[str] = None,
) -> None:
    """
    Save a phase result to disk.

    Convenience function that creates a ContextManager.

    Args:
        phase_name: Name of the phase
        result: Phase result dictionary
        project_root: Project root directory (defaults to cwd)
    """
    if project_root is None:
        project_root = os.getcwd()

    manager = ContextManager(project_root)
    manager.save_phase_result(phase_name, result)
