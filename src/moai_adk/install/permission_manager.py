"""
@FEATURE:PERMISSION-MANAGER-001 Permission Management for MoAI-ADK

@TASK:PERMISSION-MANAGER-001 Handles cross-platform file permission management
@DESIGN:MODULE-SPLIT-001 Extracted from file_operations.py for TRUST principle compliance
"""

import json
import logging
from pathlib import Path

logger = logging.getLogger(__name__)


class PermissionManager:
    """Cross-platform permission manager for MoAI-ADK files."""

    def ensure_hook_permissions(self, claude_root: Path) -> None:
        """
        Ensure .claude/hooks/moai/*.py files have execution permissions.

        Args:
            claude_root: .claude directory path
        """
        try:
            hooks_dir = claude_root / "hooks" / "moai"
            if not hooks_dir.exists():
                return

            python_files = list(hooks_dir.glob("*.py"))
            for py_file in python_files:
                # Add execute permission for owner
                current_permissions = py_file.stat().st_mode
                py_file.chmod(current_permissions | 0o100)

            if python_files:
                logger.info(f"Set execution permissions for {len(python_files)} hook files")

        except Exception as e:
            logger.warning(f"Failed to set hook permissions: {e}")

    def customize_settings_json(self, claude_root: Path, python_command: str) -> None:
        """
        Customize .claude/settings.json with Python command.

        Args:
            claude_root: .claude directory path
            python_command: Python command to use
        """
        try:
            settings_path = claude_root / "settings.json"
            if not settings_path.exists():
                logger.warning(f"settings.json not found: {settings_path}")
                return

            # Read current settings
            with open(settings_path) as f:
                settings = json.load(f)

            # Update Python command in hooks if present
            if "hooks" in settings:
                for hook_name, hook_config in settings["hooks"].items():
                    if isinstance(hook_config, dict) and "command" in hook_config:
                        # Replace python/python3 with user's preference
                        command = hook_config["command"]
                        if command.startswith(("python ", "python3 ")):
                            hook_config["command"] = command.replace(
                                command.split()[0], python_command, 1
                            )

            # Write back
            with open(settings_path, "w") as f:
                json.dump(settings, f, indent=2)

            logger.info(f"Customized settings.json with Python command: {python_command}")

        except Exception as e:
            logger.warning(f"Failed to customize settings.json: {e}")