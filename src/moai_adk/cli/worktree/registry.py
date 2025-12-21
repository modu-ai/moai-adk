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
        self._data: dict[str, dict | list[dict]] = {}
        self._load()

    def _is_valid_entry(self, entry: dict) -> bool:
        """Check if a registry entry has the required fields and types."""
        required_fields = {
            "spec_id",
            "path",
            "branch",
            "created_at",
            "last_accessed",
            "status",
        }

        if not required_fields.issubset(entry.keys()):
            return False

        return all(isinstance(entry.get(f), str) for f in required_fields)

    def _entries_for_spec(self, spec_id: str) -> list[dict]:
        """Return all entries for a spec_id, normalized as a list."""
        entry = self._data.get(spec_id)
        if isinstance(entry, list):
            return [item for item in entry if isinstance(item, dict)]
        if isinstance(entry, dict):
            return [entry]
        return []

    def _entry_project_name(self, entry: dict) -> str | None:
        """Infer project_name for an entry from explicit field or path."""
        project_name = entry.get("project_name")
        if isinstance(project_name, str) and project_name:
            return project_name

        path_value = entry.get("path")
        if not isinstance(path_value, str):
            return None

        path = Path(path_value)
        if not path.is_absolute():
            path = self.worktree_root / path

        try:
            relative = path.relative_to(self.worktree_root)
        except Exception:
            return None

        if len(relative.parts) >= 2:
            return relative.parts[0]
        return None

    def _entry_matches_project(self, entry: dict, project_name: str) -> bool:
        """Check whether a registry entry matches a project_name."""
        if not project_name:
            return False
        return self._entry_project_name(entry) == project_name

    def _entry_matches_path(self, entry: dict, path: Path | None) -> bool:
        """Check whether a registry entry matches a path."""
        if path is None:
            return False
        path_value = entry.get("path")
        return isinstance(path_value, str) and path_value == str(path)

    def _set_entries(self, spec_id: str, entries: list[dict]) -> None:
        """Set normalized entries for a spec_id in the registry."""
        if not entries:
            self._data.pop(spec_id, None)
            return

        if len(entries) == 1:
            self._data[spec_id] = entries[0]
        else:
            self._data[spec_id] = entries

    def _upsert_entry(
        self,
        spec_id: str,
        entry: dict,
        project_name: str | None = None,
        path: Path | None = None,
    ) -> None:
        """Insert or update a registry entry without writing to disk."""
        if project_name:
            entry = dict(entry)
            entry["project_name"] = project_name

        if project_name is None:
            self._data[spec_id] = entry
            return

        entries = self._entries_for_spec(spec_id)
        if not entries:
            self._data[spec_id] = entry
            return

        for index, existing in enumerate(entries):
            if self._entry_matches_project(existing, project_name) or self._entry_matches_path(existing, path):
                entries[index] = entry
                self._set_entries(spec_id, entries)
                return

        entries.append(entry)
        self._set_entries(spec_id, entries)

    def _has_entry(self, spec_id: str, project_name: str | None, path: Path | None) -> bool:
        """Check whether an entry already exists for this spec/project/path."""
        entries = self._entries_for_spec(spec_id)
        if not entries:
            return False

        for entry in entries:
            if self._entry_matches_path(entry, path):
                return True
            if project_name and self._entry_matches_project(entry, project_name):
                return True

        return False

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

    def _validate_data(self, raw_data: dict) -> dict[str, dict | list[dict]]:
        """Validate registry data structure.

        Filters out invalid entries and ensures all entries have required fields.

        Args:
            raw_data: Raw data loaded from JSON file.

        Returns:
            Validated data dictionary with only valid entries.
        """
        if not isinstance(raw_data, dict):
            return {}

        validated: dict[str, dict | list[dict]] = {}

        for spec_id, entry in raw_data.items():
            if isinstance(entry, dict):
                if self._is_valid_entry(entry):
                    validated[spec_id] = entry
                continue

            if isinstance(entry, list):
                valid_entries = [item for item in entry if isinstance(item, dict) and self._is_valid_entry(item)]
                if not valid_entries:
                    continue
                validated[spec_id] = valid_entries[0] if len(valid_entries) == 1 else valid_entries
                continue

        return validated

    def _save(self) -> None:
        """Save registry to disk."""
        self.registry_path.parent.mkdir(parents=True, exist_ok=True)
        with open(self.registry_path, "w") as f:
            json.dump(self._data, f, indent=2)

    def register(self, info: WorktreeInfo, project_name: str | None = None) -> None:
        """Register a new worktree.

        Args:
            info: WorktreeInfo instance to register.
            project_name: Project name for namespace organization.
        """
        self._upsert_entry(info.spec_id, info.to_dict(), project_name=project_name, path=info.path)
        self._save()

    def unregister(self, spec_id: str, project_name: str | None = None) -> None:
        """Unregister a worktree.

        Args:
            spec_id: SPEC ID to unregister.
            project_name: Project name for namespace organization.
        """
        if project_name is None:
            if spec_id in self._data:
                del self._data[spec_id]
                self._save()
            return

        entries = self._entries_for_spec(spec_id)
        if not entries:
            return

        remaining = [entry for entry in entries if not self._entry_matches_project(entry, project_name)]
        if len(remaining) == len(entries):
            return

        self._set_entries(spec_id, remaining)
        self._save()

    def get(self, spec_id: str, project_name: str | None = None) -> WorktreeInfo | None:
        """Get worktree information by SPEC ID.

        Args:
            spec_id: SPEC ID to retrieve.
            project_name: Project name for namespace organization.

        Returns:
            WorktreeInfo if found, None otherwise.
        """
        entries = self._entries_for_spec(spec_id)
        if not entries:
            return None

        if project_name:
            for entry in entries:
                if self._entry_matches_project(entry, project_name):
                    return WorktreeInfo.from_dict(entry)
            return None

        if len(entries) == 1:
            return WorktreeInfo.from_dict(entries[0])
        return None

    def list_all(self, project_name: str | None = None) -> list[WorktreeInfo]:
        """List all registered worktrees.

        Args:
            project_name: Project name for namespace organization.

        Returns:
            List of WorktreeInfo instances.
        """
        items: list[WorktreeInfo] = []
        for spec_id in self._data:
            for entry in self._entries_for_spec(spec_id):
                if project_name and not self._entry_matches_project(entry, project_name):
                    continue
                items.append(WorktreeInfo.from_dict(entry))
        return items

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
            changed = False
            for spec_id in list(self._data.keys()):
                entries = self._entries_for_spec(spec_id)
                if not entries:
                    self._data.pop(spec_id, None)
                    changed = True
                    continue

                kept_entries = []
                for entry in entries:
                    path_value = entry.get("path")
                    if not isinstance(path_value, str):
                        changed = True
                        continue
                    if path_value in actual_paths:
                        kept_entries.append(entry)
                    else:
                        changed = True

                if kept_entries:
                    normalized = kept_entries[0] if len(kept_entries) == 1 else kept_entries
                    if self._data.get(spec_id) != normalized:
                        self._data[spec_id] = normalized
                        changed = True
                else:
                    self._data.pop(spec_id, None)
                    changed = True

            if changed:
                self._save()

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

        def register_candidate(worktree_path: Path, spec_id: str, project_name: str | None) -> None:
            nonlocal recovered

            if self._has_entry(spec_id, project_name, worktree_path):
                return

            git_path = worktree_path / ".git"
            if not git_path.exists():
                return

            # Try to detect branch name
            branch = f"feature/{spec_id}"
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

            now = datetime.now().isoformat() + "Z"
            info_dict = {
                "spec_id": spec_id,
                "path": str(worktree_path),
                "branch": branch,
                "created_at": now,
                "last_accessed": now,
                "status": "recovered",
            }
            self._upsert_entry(spec_id, info_dict, project_name=project_name, path=worktree_path)
            recovered += 1

        for item in self.worktree_root.iterdir():
            # Skip registry file and hidden files
            if item.name.startswith("."):
                continue

            # Skip if not a directory
            if not item.is_dir():
                continue

            # Legacy layout: worktree directly under root
            if (item / ".git").exists():
                register_candidate(item, item.name, None)
                continue

            # Namespaced layout: /worktrees/{project}/{spec_id}
            for child in item.iterdir():
                if child.name.startswith("."):
                    continue
                if not child.is_dir():
                    continue
                if not (child / ".git").exists():
                    continue
                register_candidate(child, child.name, item.name)

        if recovered > 0:
            self._save()

        return recovered
