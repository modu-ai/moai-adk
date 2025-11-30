"""Data models for Git Worktree CLI."""

from dataclasses import dataclass
from pathlib import Path


@dataclass
class WorktreeInfo:
    """Metadata information for a Git worktree.

    This dataclass stores all essential information about a managed worktree,
    including its location, associated branch, creation time, and status.
    """

    spec_id: str
    """The SPEC ID associated with this worktree (e.g., 'SPEC-AUTH-001')."""

    path: Path
    """Absolute path to the worktree directory."""

    branch: str
    """The Git branch associated with this worktree."""

    created_at: str
    """ISO 8601 timestamp of worktree creation."""

    last_accessed: str
    """ISO 8601 timestamp of last access."""

    status: str
    """Current status of the worktree ('active', 'merged', 'stale')."""

    def to_dict(self) -> dict:
        """Convert WorktreeInfo to dictionary for JSON serialization.

        Returns:
            Dictionary with all worktree metadata.
        """
        return {
            "spec_id": self.spec_id,
            "path": str(self.path),
            "branch": self.branch,
            "created_at": self.created_at,
            "last_accessed": self.last_accessed,
            "status": self.status,
        }

    @classmethod
    def from_dict(cls, data: dict) -> "WorktreeInfo":
        """Create WorktreeInfo from dictionary (JSON deserialization).

        Args:
            data: Dictionary containing worktree metadata.

        Returns:
            New WorktreeInfo instance.
        """
        return cls(
            spec_id=data["spec_id"],
            path=Path(data["path"]),
            branch=data["branch"],
            created_at=data["created_at"],
            last_accessed=data["last_accessed"],
            status=data["status"],
        )
