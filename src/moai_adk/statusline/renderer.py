"""
Statusline renderer for Claude Code status display

"""

# type: ignore

from dataclasses import dataclass
from typing import List

from .config import StatuslineConfig  # type: ignore[attr-defined]


@dataclass
class StatuslineData:
    """Status line data structure containing all necessary information"""

    model: str
    version: str
    memory_usage: str
    branch: str
    git_status: str
    duration: str
    directory: str
    active_task: str
    update_available: bool = False
    latest_version: str = ""


class StatuslineRenderer:
    """Renders status information in various modes (compact, extended, minimal)"""

    # Constraints for each mode
    _MODE_CONSTRAINTS = {
        "compact": 80,
        "extended": 120,
        "minimal": 40,
    }

    def __init__(self):
        """Initialize renderer with configuration"""
        self._config = StatuslineConfig()
        self._format_config = self._config.get_format_config()
        self._display_config = self._config.get_display_config()

    def render(self, data: StatuslineData, mode: str = "compact") -> str:
        """
        Render statusline with given data in specified mode

        Args:
            data: StatuslineData instance with all required fields
            mode: Display mode - "compact" (80 chars), "extended" (120 chars), "minimal" (40 chars)

        Returns:
            Formatted statusline string
        """
        render_method = {
            "compact": self._render_compact,
            "extended": self._render_extended,
            "minimal": self._render_minimal,
        }.get(mode, self._render_compact)

        return render_method(data)

    def _render_compact(self, data: StatuslineData) -> str:
        """
        Render compact mode: [MODEL] [DURATION] | [DIR] | [VERSION] | [BRANCH] | [GIT] | [TASK]
        Constraint: <= 80 characters

        Args:
            data: StatuslineData instance

        Returns:
            Formatted statusline string (max 80 chars)
        """
        max_length = self._MODE_CONSTRAINTS["compact"]
        parts = self._build_compact_parts(data)
        result = self._format_config.separator.join(parts)

        # Adjust if too long
        if len(result) > max_length:
            result = self._fit_to_constraint(data, max_length)

        return result

    def _build_compact_parts(self, data: StatuslineData) -> List[str]:
        """
        Build parts list for compact mode with labeled sections
        Format: 🤖 Model | 🗿 Version | Output Style | 📋 TODO | 🔀 Git: Branch | Changes

        Args:
            data: StatuslineData instance

        Returns:
            List of parts to be joined
        """
        parts = []

        # Add model if display enabled (most important - cloud service context)
        if self._display_config.model:
            parts.append(f"🤖 {data.model}")

        # Add version if display enabled (system status)
        if self._display_config.version:
            parts.append(f"🗿 Ver {data.version}")

        # Add git status if display enabled and status not empty
        if self._display_config.git_status and data.git_status:
            parts.append(f"📊 {data.git_status}")

        # Add Git info (development context)
        if self._display_config.branch:
            parts.append(f"🔀 {data.branch}")

        # Add active_task if display enabled and not empty
        if self._display_config.active_task and data.active_task.strip():
            parts.append(data.active_task)

        return parts

    def _fit_to_constraint(self, data: StatuslineData, max_length: int) -> str:
        """
        Fit statusline to character constraint by truncating
        Format: 🤖 Model | 🗿 Version | Output Style | 📋 TODO | 🔀 Git: Branch | Changes

        Args:
            data: StatuslineData instance
            max_length: Maximum allowed length

        Returns:
            Truncated statusline string
        """
        # Try with truncated branch first
        truncated_branch = self._truncate_branch(data.branch, max_length=20)
        parts = [
            f"🤖 {data.model}",
            f"🗿 Ver {data.version}",
        ]

        # Add git status if display enabled and status not empty
        if self._display_config.git_status and data.git_status:
            parts.append(f"📊 {data.git_status}")

        # Add Git info
        parts.append(f"🔀 {truncated_branch}")

        # Only add active_task if it's not empty
        if data.active_task.strip():
            parts.append(data.active_task)

        result = self._format_config.separator.join(parts)

        # If still too long, try more aggressive branch truncation
        if len(result) > max_length:
            truncated_branch = self._truncate_branch(data.branch, max_length=12)
            parts = [
                f"🤖 {data.model}",
                f"🗿 Ver {data.version}",
            ]
            if data.git_status:
                parts.append(f"📊 {data.git_status}")
            parts.append(f"🔀 {truncated_branch}")
            if data.active_task.strip():
                parts.append(data.active_task)
            result = self._format_config.separator.join(parts)

        # If still too long, remove active_task
        if len(result) > max_length and data.active_task.strip():
            parts = [
                f"🤖 {data.model}",
                f"🗿 Ver {data.version}",
            ]
            if data.git_status:
                parts.append(f"📊 {data.git_status}")
            parts.append(f"🔀 {truncated_branch}")
            result = self._format_config.separator.join(parts)

        # Final fallback to minimal if still too long
        if len(result) > max_length:
            result = self._render_minimal(data)

        return result

    def _render_extended(self, data: StatuslineData) -> str:
        """
        Render extended mode: Full path and detailed info with labels
        Constraint: <= 120 characters
        Format: 🤖 Model | 🗿 Version | Output Style | 📋 TODO | 🔀 Git: Branch | Changes

        Args:
            data: StatuslineData instance

        Returns:
            Formatted statusline string (max 120 chars)
        """
        branch = self._truncate_branch(data.branch, max_length=30)

        parts = []

        # Add model if display enabled
        if self._display_config.model:
            parts.append(f"🤖 {data.model}")

        # Add version if display enabled
        if self._display_config.version:
            parts.append(f"🗿 Ver {data.version}")

        # Add git status if display enabled and status not empty
        if self._display_config.git_status and data.git_status:
            parts.append(f"📊 {data.git_status}")

        # Add Git info (development context)
        if self._display_config.branch:
            parts.append(f"🔀 {branch}")

        # Add active_task if display enabled and not empty
        if self._display_config.active_task and data.active_task.strip():
            parts.append(data.active_task)

        result = self._format_config.separator.join(parts)

        # If exceeds limit, try truncating branch
        if len(result) > 120:
            branch = self._truncate_branch(data.branch, max_length=20)
            parts = [
                f"🤖 {data.model}",
                f"🗿 Ver {data.version}",
            ]

            # Add optional parts only if they have content
            if self._display_config.output_style and data.output_style.strip():
                # Use Yoda-specific emoji for Yoda styles only
                if "yoda" in data.output_style.lower() or "Yoda Master" in data.output_style:
                    parts.append(f"🧙 {data.output_style}")
                else:
                    parts.append(f"{data.output_style}")
            if self._display_config.todo_count and data.todo_count.strip():
                parts.append(f"📋 {data.todo_count}")

            parts.append(f"🔀 {branch}")

            if data.git_status:
                parts.append(f"Changes: {data.git_status}")
            if data.active_task.strip():
                parts.append(data.active_task)
            result = self._format_config.separator.join(parts)

        return result

    def _render_minimal(self, data: StatuslineData) -> str:
        """
        Render minimal mode: Extreme space constraint with minimal labels
        Constraint: <= 40 characters
        Format: 🤖 Model | 🗿 Ver Version | Changes: +staged M modified ? untracked

        Args:
            data: StatuslineData instance

        Returns:
            Formatted statusline string (max 40 chars)
        """
        parts = []

        # Add model if display enabled
        if self._display_config.model:
            parts.append(f"🤖 {data.model}")

        # Add version if display enabled
        if self._display_config.version:
            parts.append(f"🗿 Ver {self._truncate_version(data.version)}")

        result = self._format_config.separator.join(parts)

        # Add git_status if it fits (use abbreviated format for minimal)
        # and if display is enabled and status not empty
        if self._display_config.git_status and data.git_status:
            status_label = f"Chg: {data.git_status}"
            if (
                len(result) + len(status_label) + len(self._format_config.separator)
                <= 40
            ):
                result += f"{self._format_config.separator}{status_label}"

        return result

    @staticmethod
    def _truncate_branch(branch: str, max_length: int = 20) -> str:
        """
        Truncate branch name intelligently, preserving SPEC ID if present

        Args:
            branch: Branch name to truncate
            max_length: Maximum allowed length

        Returns:
            Truncated branch name
        """
        if len(branch) <= max_length:
            return branch

        # Try to preserve SPEC ID in feature branches
        if "SPEC" in branch:
            parts = branch.split("-")
            for i, part in enumerate(parts):
                if "SPEC" in part and i + 1 < len(parts):
                    # Found SPEC ID, include it
                    spec_truncated = "-".join(parts[: i + 2])
                    if len(spec_truncated) <= max_length:
                        return spec_truncated

        # Simple truncation with ellipsis for very long names
        return f"{branch[:max_length-1]}…" if len(branch) > max_length else branch

    @staticmethod
    def _truncate_version(version: str) -> str:
        """
        Truncate version string for minimal display by removing 'v' prefix

        Args:
            version: Version string (e.g., "v0.20.1" or "0.20.1")

        Returns:
            Truncated version string
        """
        if version.startswith("v"):
            return version[1:]
        return version
