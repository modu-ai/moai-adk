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
        Validates data structure and removes invalid entries.
        """
        if self.registry_path.exists():
            try:
                with open(self.registry_path, "r") as f:
                    content = f.read().strip()
                    if content:
                        raw_data = json.loads(content)
                        # Validate and filter data
                        self._data = self._validate_data(raw_data)
                    else:
                        self._data = {}
            except (json.JSONDecodeError, IOError):
                self._data = {}
        else:
            # Create parent directory if needed
            self.registry_path.parent.mkdir(parents=True, exist_ok=True)
            self._data = {}
            self._save()

    def _validate_data(self, raw_data: dict) -> dict[str, dict]:
        """Validate registry data structure.

        Filters out invalid entries and ensures all entries have required fields.

        Args:
            raw_data: Raw data loaded from JSON file.

        Returns:
            Validated data dictionary with only valid entries.
        """
        if not isinstance(raw_data, dict):
            return {}

        required_fields = {"spec_id", "path", "branch", "created_at", "last_accessed", "status"}
        validated = {}

        for spec_id, entry in raw_data.items():
            # Skip if entry is not a dictionary
            if not isinstance(entry, dict):
                continue

            # Skip if missing required fields
            if not required_fields.issubset(entry.keys()):
                continue

            # Validate field types
            if not all(isinstance(entry.get(f), str) for f in required_fields):
                continue

            validated[spec_id] = entry

        return validated

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
                # Defensive check: ensure data is a dict with 'path' key
                if not isinstance(data, dict):
                    spec_ids_to_remove.append(spec_id)
                    continue
                if "path" not in data:
                    spec_ids_to_remove.append(spec_id)
                    continue
                if data["path"] not in actual_paths:
                    spec_ids_to_remove.append(spec_id)

            for spec_id in spec_ids_to_remove:
                self.unregister(spec_id)

        except Exception:
            # If sync fails, just continue
            pass

    def recover_from_disk(self) -> int:
        """Recover worktree registry from existing worktree directories.

        Scans the worktree_root directory for existing worktrees and
        registers them if they have valid Git structure.

        Returns:
            Number of worktrees recovered.
        """
        from datetime import datetime

        recovered = 0

        if not self.worktree_root.exists():
            return 0

        for item in self.worktree_root.iterdir():
            # Skip registry file and hidden files
            if item.name.startswith("."):
                continue

            # Skip if not a directory
            if not item.is_dir():
                continue

            # Skip if already registered
            if item.name in self._data:
                continue

            # Check if it's a valid worktree (has .git file or directory)
            git_path = item / ".git"
            if not git_path.exists():
                continue

            # Try to detect branch name
            branch = f"feature/{item.name}"
            try:
                if git_path.is_file():
                    # It's a worktree - read the gitdir to find HEAD
                    with open(git_path, "r") as f:
                        for line in f:
                            if line.startswith("gitdir:"):
                                gitdir = Path(line[8:].strip())
                                head_file = gitdir / "HEAD"
                                if head_file.exists():
                                    with open(head_file, "r") as hf:
                                        head_content = hf.read().strip()
                                        if head_content.startswith("ref: refs/heads/"):
                                            branch = head_content[16:]
                                break
            except Exception:
                pass

            # Create WorktreeInfo and register
            now = datetime.now().isoformat() + "Z"
            info_dict = {
                "spec_id": item.name,
                "path": str(item),
                "branch": branch,
                "created_at": now,
                "last_accessed": now,
                "status": "recovered",
            }
            self._data[item.name] = info_dict
            recovered += 1

        if recovered > 0:
            self._save()

        return recovered
