"""Registry for managing Git worktree metadata."""

import json
from pathlib import Path

from moai_adk.cli.worktree.models import WorktreeInfo


class WorktreeRegistry:
    """Manages Git worktree metadata persistence.

    This class handles storing and retrieving worktree information from
    a JSON registry file. It ensures registry consistency and provides
    CRUD operations for worktree metadata.
    """

    def __init__(self, worktree_root: Path) -> None:
        """Initialize the registry.

        Creates the registry file if it doesn't exist.

        Args:
            worktree_root: Root directory for worktrees.
        """
        self.worktree_root = worktree_root
        self.registry_path = worktree_root / ".moai-worktree-registry.json"
        self._data: dict[str, dict] = {}
        self._load()

    def _load(self) -> None:
        """Load registry from disk.

        Initializes empty registry if file doesn't exist.
        """
        if self.registry_path.exists():
            try:
                with open(self.registry_path, "r") as f:
                    content = f.read().strip()
                    if content:
                        self._data = json.loads(content)
                    else:
                        self._data = {}
            except (json.JSONDecodeError, IOError):
                self._data = {}
        else:
            # Create parent directory if needed
            self.registry_path.parent.mkdir(parents=True, exist_ok=True)
            self._data = {}
            self._save()

    def _save(self) -> None:
        """Save registry to disk."""
        self.registry_path.parent.mkdir(parents=True, exist_ok=True)
        with open(self.registry_path, "w") as f:
            json.dump(self._data, f, indent=2)

    def register(self, info: WorktreeInfo) -> None:
        """Register a new worktree.

        Args:
            info: WorktreeInfo instance to register.
        """
        self._data[info.spec_id] = info.to_dict()
        self._save()

    def unregister(self, spec_id: str) -> None:
        """Unregister a worktree.

        Args:
            spec_id: SPEC ID to unregister.
        """
        if spec_id in self._data:
            del self._data[spec_id]
            self._save()

    def get(self, spec_id: str) -> WorktreeInfo | None:
        """Get worktree information by SPEC ID.

        Args:
            spec_id: SPEC ID to retrieve.

        Returns:
            WorktreeInfo if found, None otherwise.
        """
        if spec_id in self._data:
            return WorktreeInfo.from_dict(self._data[spec_id])
        return None

    def list_all(self) -> list[WorktreeInfo]:
        """List all registered worktrees.

        Returns:
            List of WorktreeInfo instances.
        """
        return [WorktreeInfo.from_dict(data) for data in self._data.values()]

    def sync_with_git(self, repo) -> None:
        """Synchronize registry with actual Git worktree state.

        Removes entries for worktrees that no longer exist on disk.

        Args:
            repo: GitPython Repo instance.
        """
        # Get list of actual Git worktrees
        try:
            worktrees = repo.git.worktree("list", "--porcelain").split("\n")
            actual_paths = set()

            for line in worktrees:
                if line.strip() and line.startswith("worktree "):
                    # Parse worktree list output - lines start with "worktree "
                    path = line[9:].strip()  # Remove "worktree " prefix
                    if path:
                        actual_paths.add(path)

            # Remove registry entries for non-existent worktrees
            spec_ids_to_remove = []
            for spec_id, data in self._data.items():
                if data["path"] not in actual_paths:
                    spec_ids_to_remove.append(spec_id)

            for spec_id in spec_ids_to_remove:
                self.unregister(spec_id)

        except Exception:
            # If sync fails, just continue
            pass
