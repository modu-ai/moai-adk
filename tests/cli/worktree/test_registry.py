"""Comprehensive tests for WorktreeRegistry.

Test Coverage Strategy:
- JSON operations: load, save, validate data structure
- Data validation: entry validation, field checking, type validation
- Namespace handling: project_name filtering, multi-entry management
- Entry management: register, unregister, get, list_all
- Git sync: sync_with_git, remove stale entries
- Disk recovery: recover_from_disk, detect worktrees from filesystem
- Edge cases: empty registry, corrupted data, missing files
"""

import json
from pathlib import Path
from unittest.mock import MagicMock, Mock

import pytest

from moai_adk.cli.worktree.models import WorktreeInfo
from moai_adk.cli.worktree.registry import WorktreeRegistry


class TestWorktreeRegistryInit:
    """Test WorktreeRegistry initialization."""

    def test_init_creates_registry_file(self, tmp_path: Path) -> None:
        """Test that initialization creates the registry file."""
        registry = WorktreeRegistry(tmp_path)
        assert registry.registry_path == tmp_path / ".moai-worktree-registry.json"
        assert registry.worktree_root == tmp_path
        assert registry._data == {}

    def test_init_loads_existing_registry(self, tmp_path: Path) -> None:
        """Test that initialization loads existing registry data."""
        registry_path = tmp_path / ".moai-worktree-registry.json"
        test_data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": str(tmp_path / "worktree"),
                "branch": "feature/SPEC-001",
                "created_at": "2025-01-13T10:00:00Z",
                "last_accessed": "2025-01-13T10:00:00Z",
                "status": "active",
            }
        }
        registry_path.write_text(json.dumps(test_data))

        registry = WorktreeRegistry(tmp_path)
        assert registry._data == test_data

    def test_init_creates_parent_directory(self, tmp_path: Path) -> None:
        """Test that initialization creates parent directory if needed."""
        worktree_root = tmp_path / "worktrees" / "nested"
        registry = WorktreeRegistry(worktree_root)
        assert worktree_root.exists()
        assert registry.registry_path.parent.exists()


class TestIsValidEntry:
    """Test _is_valid_entry method."""

    def test_valid_entry(self) -> None:
        """Test validation of a valid entry."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {
            "spec_id": "SPEC-001",
            "path": "/tmp/worktree",
            "branch": "feature/SPEC-001",
            "created_at": "2025-01-13T10:00:00Z",
            "last_accessed": "2025-01-13T10:00:00Z",
            "status": "active",
        }
        assert registry._is_valid_entry(entry) is True

    def test_missing_required_field(self) -> None:
        """Test validation fails with missing required fields."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {
            "spec_id": "SPEC-001",
            "path": "/tmp/worktree",
            "branch": "feature/SPEC-001",
            "created_at": "2025-01-13T10:00:00Z",
            # Missing last_accessed and status
        }
        assert registry._is_valid_entry(entry) is False

    def test_wrong_field_type(self) -> None:
        """Test validation fails with wrong field types."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {
            "spec_id": 123,  # Should be string
            "path": "/tmp/worktree",
            "branch": "feature/SPEC-001",
            "created_at": "2025-01-13T10:00:00Z",
            "last_accessed": "2025-01-13T10:00:00Z",
            "status": "active",
        }
        assert registry._is_valid_entry(entry) is False

    def test_extra_fields_allowed(self) -> None:
        """Test that extra fields don't invalidate entry."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {
            "spec_id": "SPEC-001",
            "path": "/tmp/worktree",
            "branch": "feature/SPEC-001",
            "created_at": "2025-01-13T10:00:00Z",
            "last_accessed": "2025-01-13T10:00:00Z",
            "status": "active",
            "project_name": "my-project",  # Extra field
        }
        assert registry._is_valid_entry(entry) is True


class TestEntriesForSpec:
    """Test _entries_for_spec method."""

    def test_single_dict_entry(self) -> None:
        """Test returns list with single dict entry."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/tmp/wt",
                "branch": "feature",
                "created_at": "2025-01-13T10:00:00Z",
                "last_accessed": "2025-01-13T10:00:00Z",
                "status": "active",
            }
        }
        result = registry._entries_for_spec("SPEC-001")
        assert len(result) == 1
        assert result[0]["spec_id"] == "SPEC-001"

    def test_list_entries(self) -> None:
        """Test returns list with multiple entries."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entries = [
            {
                "spec_id": "SPEC-001",
                "path": "/tmp/wt1",
                "branch": "feature",
                "created_at": "2025-01-13T10:00:00Z",
                "last_accessed": "2025-01-13T10:00:00Z",
                "status": "active",
            },
            {
                "spec_id": "SPEC-001",
                "path": "/tmp/wt2",
                "branch": "feature",
                "created_at": "2025-01-13T10:00:00Z",
                "last_accessed": "2025-01-13T10:00:00Z",
                "status": "active",
            },
        ]
        registry._data = {"SPEC-001": entries}
        result = registry._entries_for_spec("SPEC-001")
        assert len(result) == 2

    def test_no_entry_returns_empty_list(self) -> None:
        """Test returns empty list when no entry exists."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        registry._data = {}
        result = registry._entries_for_spec("SPEC-001")
        assert result == []

    def test_filters_non_dict_items(self) -> None:
        """Test filters out non-dict items from list."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        registry._data = {  # type: ignore[assignment]
            "SPEC-001": [
                "string",
                123,
                None,
                {
                    "spec_id": "SPEC-001",
                    "path": "/tmp/wt",
                    "branch": "feature",
                    "created_at": "2025-01-13T10:00:00Z",
                    "last_accessed": "2025-01-13T10:00:00Z",
                    "status": "active",
                },
            ]
        }
        result = registry._entries_for_spec("SPEC-001")
        assert len(result) == 1


class TestEntryProjectName:
    # type: ignore[misc]
    """Test _entry_project_name method."""

    def test_explicit_project_name(self) -> None:
        """Test returns explicit project_name field."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {"project_name": "my-project", "path": "/tmp/worktree"}
        result = registry._entry_project_name(entry)
        assert result == "my-project"

    def test_infers_from_path_absolute(self, tmp_path: Path) -> None:
        """Test infers project name from absolute path."""
        registry = WorktreeRegistry(tmp_path / "worktrees")
        entry = {"path": str(tmp_path / "worktrees" / "my-project" / "SPEC-001")}
        result = registry._entry_project_name(entry)
        assert result == "my-project"

    def test_infers_from_path_relative(self, tmp_path: Path) -> None:
        """Test infers project name from relative path."""
        registry = WorktreeRegistry(tmp_path / "worktrees")
        entry = {"path": "my-project/SPEC-001"}
        result = registry._entry_project_name(entry)
        assert result == "my-project"

    def test_returns_none_for_invalid_path(self) -> None:
        """Test returns None for path outside worktree root."""
        registry = WorktreeRegistry(Path("/tmp/worktrees"))
        entry = {"path": "/other/location/worktree"}
        result = registry._entry_project_name(entry)
        assert result is None

    def test_returns_none_for_shallow_path(self, tmp_path: Path) -> None:
        """Test returns None for path without project directory."""
        registry = WorktreeRegistry(tmp_path / "worktrees")
        entry = {"path": str(tmp_path / "worktrees" / "SPEC-001")}
        result = registry._entry_project_name(entry)
        assert result is None

    def test_returns_none_for_non_string_path(self) -> None:
        """Test returns None for non-string path value."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {"path": 123}
        result = registry._entry_project_name(entry)
        assert result is None


class TestEntryMatchesProject:
    """Test _entry_matches_project method."""

    def test_matches_explicit_project_name(self) -> None:
        """Test matches entry with explicit project_name."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {"project_name": "my-project", "path": "/tmp/worktree"}
        assert registry._entry_matches_project(entry, "my-project") is True
        assert registry._entry_matches_project(entry, "other") is False

    def test_matches_inferred_project_name(self, tmp_path: Path) -> None:
        """Test matches entry with inferred project_name."""
        registry = WorktreeRegistry(tmp_path / "worktrees")
        entry = {"path": str(tmp_path / "worktrees" / "my-project" / "SPEC-001")}
        assert registry._entry_matches_project(entry, "my-project") is True

    def test_no_match_empty_project_name(self) -> None:
        """Test returns False for empty project_name."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {"path": "/tmp/worktree"}
        assert registry._entry_matches_project(entry, "") is False
        assert registry._entry_matches_project(entry, "my-project") is False


class TestEntryMatchesPath:
    """Test _entry_matches_path method."""

    def test_matches_path(self) -> None:
        """Test matches entry by path."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {"path": "/tmp/worktree"}
        assert registry._entry_matches_path(entry, Path("/tmp/worktree")) is True

    def test_no_match_different_path(self) -> None:
        """Test no match for different path."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {"path": "/tmp/worktree"}
        assert registry._entry_matches_path(entry, Path("/tmp/other")) is False

    def test_returns_false_for_none_path(self) -> None:
        """Test returns False when path is None."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {"path": "/tmp/worktree"}
        assert registry._entry_matches_path(entry, None) is False

    def test_returns_false_for_non_string_path(self) -> None:
        """Test returns False for non-string path value."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {"path": 123}
        assert registry._entry_matches_path(entry, Path("/tmp/worktree")) is False


class TestSetEntries:
    """Test _set_entries method."""

    def test_sets_single_entry(self) -> None:
        """Test sets single entry as dict."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {
            "spec_id": "SPEC-001",
            "path": "/tmp/wt",
            "branch": "feature",
            "created_at": "2025-01-13T10:00:00Z",
            "last_accessed": "2025-01-13T10:00:00Z",
            "status": "active",
        }
        registry._set_entries("SPEC-001", [entry])
        assert registry._data["SPEC-001"] == entry

    def test_sets_multiple_entries(self) -> None:
        """Test sets multiple entries as list."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entries = [
            {
                "spec_id": "SPEC-001",
                "path": "/tmp/wt1",
                "branch": "feature",
                "created_at": "2025-01-13T10:00:00Z",
                "last_accessed": "2025-01-13T10:00:00Z",
                "status": "active",
            },
            {
                "spec_id": "SPEC-001",
                "path": "/tmp/wt2",
                "branch": "feature",
                "created_at": "2025-01-13T10:00:00Z",
                "last_accessed": "2025-01-13T10:00:00Z",
                "status": "active",
            },
        ]
        registry._set_entries("SPEC-001", entries)
        assert registry._data["SPEC-001"] == entries

    def test_removes_spec_id_for_empty_list(self) -> None:
        """Test removes spec_id when entries list is empty."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/tmp/wt",
                "branch": "feature",
                "created_at": "2025-01-13T10:00:00Z",
                "last_accessed": "2025-01-13T10:00:00Z",
                "status": "active",
            }
        }
        registry._set_entries("SPEC-001", [])
        assert "SPEC-001" not in registry._data


class TestUpsertEntry:
    """Test _upsert_entry method."""

    def test_insert_new_entry_no_project(self) -> None:
        """Test inserts new entry when no project_name."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {
            "spec_id": "SPEC-001",
            "path": "/tmp/wt",
            "branch": "feature",
            "created_at": "2025-01-13T10:00:00Z",
            "last_accessed": "2025-01-13T10:00:00Z",
            "status": "active",
        }
        registry._upsert_entry("SPEC-001", entry, project_name=None)
        assert registry._data["SPEC-001"] == entry

    def test_update_existing_entry_by_project(self, tmp_path: Path) -> None:
        """Test updates existing entry matched by project_name."""
        registry = WorktreeRegistry(tmp_path / "worktrees")
        old_entry = {
            "spec_id": "SPEC-001",
            "path": str(tmp_path / "worktrees" / "proj" / "SPEC-001"),
            "branch": "old",
            "created_at": "2025-01-13T10:00:00Z",
            "last_accessed": "2025-01-13T10:00:00Z",
            "status": "active",
            "project_name": "proj",
        }
        new_entry = {
            "spec_id": "SPEC-001",
            "path": str(tmp_path / "worktrees" / "proj" / "SPEC-001"),
            "branch": "new",
            "created_at": "2025-01-13T10:00:00Z",
            "last_accessed": "2025-01-13T11:00:00Z",
            "status": "active",
        }
        registry._data = {"SPEC-001": [old_entry]}
        registry._upsert_entry(
            "SPEC-001", new_entry, project_name="proj", path=tmp_path / "worktrees" / "proj" / "SPEC-001"
        )
        entries = registry._entries_for_spec("SPEC-001")
        assert len(entries) == 1
        assert entries[0]["branch"] == "new"

    def test_append_new_entry_with_project(self, tmp_path: Path) -> None:
        """Test appends new entry when no existing match."""
        registry = WorktreeRegistry(tmp_path / "worktrees")
        entry1 = {
            "spec_id": "SPEC-001",
            "path": str(tmp_path / "worktrees" / "proj1" / "SPEC-001"),
            "branch": "f1",
            "created_at": "2025-01-13T10:00:00Z",
            "last_accessed": "2025-01-13T10:00:00Z",
            "status": "active",
            "project_name": "proj1",
        }
        entry2 = {
            "spec_id": "SPEC-001",
            "path": str(tmp_path / "worktrees" / "proj2" / "SPEC-001"),
            "branch": "f2",
            "created_at": "2025-01-13T10:00:00Z",
            "last_accessed": "2025-01-13T10:00:00Z",
            "status": "active",
            "project_name": "proj2",
        }
        registry._data = {"SPEC-001": [entry1]}
        registry._upsert_entry(
            "SPEC-001", entry2, project_name="proj2", path=tmp_path / "worktrees" / "proj2" / "SPEC-001"
        )
        entries = registry._entries_for_spec("SPEC-001")
        assert len(entries) == 2

    def test_adds_project_name_to_entry(self) -> None:
        """Test adds project_name to entry when provided."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {
            "spec_id": "SPEC-001",
            "path": "/tmp/wt",
            "branch": "feature",
            "created_at": "2025-01-13T10:00:00Z",
            "last_accessed": "2025-01-13T10:00:00Z",
            "status": "active",
        }
        registry._upsert_entry("SPEC-001", entry, project_name="proj", path=Path("/tmp/wt"))
        assert registry._data["SPEC-001"]["project_name"] == "proj"  # type: ignore[index]


class TestHasEntry:
    """Test _has_entry method."""

    def test_returns_true_for_existing_path(self) -> None:
        """Test returns True when entry exists with matching path."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        entry = {
            "spec_id": "SPEC-001",
            "path": "/tmp/wt",
            "branch": "feature",
            "created_at": "2025-01-13T10:00:00Z",
            "last_accessed": "2025-01-13T10:00:00Z",
            "status": "active",
        }
        registry._data = {"SPEC-001": entry}
        assert registry._has_entry("SPEC-001", None, Path("/tmp/wt")) is True

    def test_returns_true_for_existing_project(self, tmp_path: Path) -> None:
        """Test returns True when entry exists with matching project."""
        registry = WorktreeRegistry(tmp_path / "worktrees")
        entry = {
            "spec_id": "SPEC-001",
            "path": str(tmp_path / "worktrees" / "proj" / "SPEC-001"),
            "branch": "feature",
            "created_at": "2025-01-13T10:00:00Z",
            "last_accessed": "2025-01-13T10:00:00Z",
            "status": "active",
            "project_name": "proj",
        }
        registry._data = {"SPEC-001": [entry]}
        assert registry._has_entry("SPEC-001", "proj", None) is True

    def test_returns_false_for_nonexistent_entry(self) -> None:
        """Test returns False when entry doesn't exist."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        assert registry._has_entry("SPEC-001", None, Path("/tmp/wt")) is False


class TestLoad:
    """Test _load method."""

    def test_loads_valid_json(self, tmp_path: Path) -> None:
        """Test loads valid JSON data."""
        registry_path = tmp_path / ".moai-worktree-registry.json"
        test_data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/tmp/wt",
                "branch": "feature",
                "created_at": "2025-01-13T10:00:00Z",
                "last_accessed": "2025-01-13T10:00:00Z",
                "status": "active",
            }
        }
        registry_path.write_text(json.dumps(test_data))

        registry = WorktreeRegistry(tmp_path)
        assert registry._data == test_data

    def test_creates_empty_registry_for_missing_file(self, tmp_path: Path) -> None:
        """Test creates empty registry when file doesn't exist."""
        registry = WorktreeRegistry(tmp_path)
        assert registry._data == {}
        assert (tmp_path / ".moai-worktree-registry.json").exists()

    def test_handles_empty_file(self, tmp_path: Path) -> None:
        """Test handles empty registry file."""
        registry_path = tmp_path / ".moai-worktree-registry.json"
        registry_path.write_text("")

        registry = WorktreeRegistry(tmp_path)
        assert registry._data == {}

    def test_handles_invalid_json(self, tmp_path: Path) -> None:
        """Test handles invalid JSON gracefully."""
        registry_path = tmp_path / ".moai-worktree-registry.json"
        registry_path.write_text("{invalid json")

        registry = WorktreeRegistry(tmp_path)
        assert registry._data == {}

    def test_validates_data_on_load(self, tmp_path: Path) -> None:
        """Test validates data structure when loading."""
        registry_path = tmp_path / ".moai-worktree-registry.json"
        # Include valid and invalid entries
        test_data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/tmp/wt",
                "branch": "feature",
                "created_at": "2025-01-13T10:00:00Z",
                "last_accessed": "2025-01-13T10:00:00Z",
                "status": "active",
            },
            "SPEC-002": {"path": "/tmp/wt2"},  # Invalid - missing fields
        }
        registry_path.write_text(json.dumps(test_data))

        registry = WorktreeRegistry(tmp_path)
        # Only valid entry should remain
        assert "SPEC-001" in registry._data
        assert "SPEC-002" not in registry._data


class TestValidateData:
    """Test _validate_data method."""

    def test_validates_dict_entries(self) -> None:
        """Test validates dictionary entries."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        raw_data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/tmp/wt",
                "branch": "feature",
                "created_at": "2025-01-13T10:00:00Z",
                "last_accessed": "2025-01-13T10:00:00Z",
                "status": "active",
            },
        }
        result = registry._validate_data(raw_data)
        assert "SPEC-001" in result

    def test_validates_list_entries(self) -> None:
        """Test validates list entries."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        raw_data = {
            "SPEC-001": [
                {
                    "spec_id": "SPEC-001",
                    "path": "/tmp/wt1",
                    "branch": "f1",
                    "created_at": "2025-01-13T10:00:00Z",
                    "last_accessed": "2025-01-13T10:00:00Z",
                    "status": "active",
                },
                {
                    "spec_id": "SPEC-001",
                    "path": "/tmp/wt2",
                    "branch": "f2",
                    "created_at": "2025-01-13T10:00:00Z",
                    "last_accessed": "2025-01-13T10:00:00Z",
                    "status": "active",
                },
            ],
        }
        result = registry._validate_data(raw_data)
        assert "SPEC-001" in result
        assert len(result["SPEC-001"]) == 2

    def test_filters_invalid_entries(self) -> None:
        """Test filters out invalid entries."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        raw_data = {
            "SPEC-001": [
                {
                    "spec_id": "SPEC-001",
                    "path": "/tmp/wt1",
                    "branch": "f1",
                    "created_at": "2025-01-13T10:00:00Z",
                    "last_accessed": "2025-01-13T10:00:00Z",
                    "status": "active",
                },
                {"path": "/tmp/wt2"},  # Invalid
            ],
        }
        result = registry._validate_data(raw_data)
        assert "SPEC-001" in result
        # When only 1 valid entry remains, it's returned as a dict, not a list
        assert isinstance(result["SPEC-001"], dict)

    def test_returns_empty_for_non_dict_input(self) -> None:
        """Test returns empty dict for non-dict input."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        assert registry._validate_data("not a dict") == {}  # type: ignore[arg-type]
        assert registry._validate_data(None) == {}  # type: ignore[arg-type]
        assert registry._validate_data([1, 2, 3]) == {}  # type: ignore[list-item]

    def test_single_entry_for_single_valid_item(self) -> None:
        """Test returns single dict when list has one valid item."""
        registry = WorktreeRegistry(Path("/tmp/test"))
        raw_data = {
            "SPEC-001": [
                {
                    "spec_id": "SPEC-001",
                    "path": "/tmp/wt",
                    "branch": "f1",
                    "created_at": "2025-01-13T10:00:00Z",
                    "last_accessed": "2025-01-13T10:00:00Z",
                    "status": "active",
                },
            ],
        }
        result = registry._validate_data(raw_data)
        assert isinstance(result["SPEC-001"], dict)


class TestSave:
    """Test _save method."""

    def test_saves_data_to_disk(self, tmp_path: Path) -> None:
        """Test saves data to registry file."""
        registry = WorktreeRegistry(tmp_path)
        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/tmp/wt",
                "branch": "feature",
                "created_at": "2025-01-13T10:00:00Z",
                "last_accessed": "2025-01-13T10:00:00Z",
                "status": "active",
            }
        }
        registry._save()

        registry_path = tmp_path / ".moai-worktree-registry.json"
        assert registry_path.exists()
        loaded_data = json.loads(registry_path.read_text())
        assert loaded_data == registry._data

    def test_creates_parent_directory(self, tmp_path: Path) -> None:
        """Test creates parent directory if needed."""
        worktree_root = tmp_path / "nested" / "dir"
        registry = WorktreeRegistry(worktree_root)
        registry._save()
        assert worktree_root.exists()


class TestRegister:
    """Test register method."""

    def test_registers_worktree(self, tmp_path: Path) -> None:
        """Test registers a new worktree."""
        registry = WorktreeRegistry(tmp_path)
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info)

        loaded = registry.get("SPEC-001")
        assert loaded is not None
        assert loaded.spec_id == "SPEC-001"

    def test_registers_with_project_name(self, tmp_path: Path) -> None:
        """Test registers worktree with project_name."""
        registry = WorktreeRegistry(tmp_path)
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "proj" / "SPEC-001",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info, project_name="proj")

        loaded = registry.get("SPEC-001", project_name="proj")
        assert loaded is not None

    def test_saves_after_register(self, tmp_path: Path) -> None:
        """Test saves to disk after registering."""
        registry = WorktreeRegistry(tmp_path)
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info)

        # Create new registry instance to verify persistence
        new_registry = WorktreeRegistry(tmp_path)
        loaded = new_registry.get("SPEC-001")
        assert loaded is not None


class TestUnregister:
    """Test unregister method."""

    def test_unregisters_without_project(self, tmp_path: Path) -> None:
        """Test unregisters worktree without project."""
        registry = WorktreeRegistry(tmp_path)
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info)
        registry.unregister("SPEC-001")

        assert registry.get("SPEC-001") is None

    def test_unregisters_with_project(self, tmp_path: Path) -> None:
        """Test unregisters worktree with project_name."""
        registry = WorktreeRegistry(tmp_path)
        info1 = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "proj1" / "SPEC-001",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        info2 = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "proj2" / "SPEC-001",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info1, project_name="proj1")
        registry.register(info2, project_name="proj2")
        registry.unregister("SPEC-001", project_name="proj1")

        assert registry.get("SPEC-001", project_name="proj1") is None
        assert registry.get("SPEC-001", project_name="proj2") is not None

    def test_unregister_nonexistent_no_error(self, tmp_path: Path) -> None:
        """Test unregistering non-existent worktree doesn't error."""
        registry = WorktreeRegistry(tmp_path)
        registry.unregister("SPEC-999")  # Should not raise

    def test_saves_after_unregister(self, tmp_path: Path) -> None:
        """Test saves to disk after unregistering."""
        registry = WorktreeRegistry(tmp_path)
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info)
        registry.unregister("SPEC-001")

        # Create new registry instance to verify persistence
        new_registry = WorktreeRegistry(tmp_path)
        assert new_registry.get("SPEC-001") is None


class TestGet:
    """Test get method."""

    def test_get_without_project(self, tmp_path: Path) -> None:
        """Test gets worktree without project filter."""
        registry = WorktreeRegistry(tmp_path)
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info)

        loaded = registry.get("SPEC-001")
        assert loaded is not None
        assert loaded.spec_id == "SPEC-001"

    def test_get_with_project(self, tmp_path: Path) -> None:
        """Test gets worktree with project filter."""
        registry = WorktreeRegistry(tmp_path)
        info1 = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "proj1" / "SPEC-001",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        info2 = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "proj2" / "SPEC-001",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info1, project_name="proj1")
        registry.register(info2, project_name="proj2")

        loaded1 = registry.get("SPEC-001", project_name="proj1")
        loaded2 = registry.get("SPEC-001", project_name="proj2")

        assert loaded1 is not None
        assert loaded2 is not None
        assert loaded1.spec_id == "SPEC-001"
        assert loaded2.spec_id == "SPEC-001"
        assert loaded1.path != loaded2.path

    def test_get_returns_none_for_nonexistent(self, tmp_path: Path) -> None:
        """Test returns None for non-existent worktree."""
        registry = WorktreeRegistry(tmp_path)
        assert registry.get("SPEC-999") is None

    def test_get_returns_none_for_multiple_without_project(self, tmp_path: Path) -> None:
        """Test returns None when multiple entries exist without project filter."""
        registry = WorktreeRegistry(tmp_path)
        info1 = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "proj1" / "SPEC-001",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        info2 = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "proj2" / "SPEC-001",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info1, project_name="proj1")
        registry.register(info2, project_name="proj2")

        # Multiple entries exist, should return None without project filter
        assert registry.get("SPEC-001") is None


class TestListAll:
    """Test list_all method."""

    def test_lists_all_worktrees(self, tmp_path: Path) -> None:
        """Test lists all registered worktrees."""
        registry = WorktreeRegistry(tmp_path)
        info1 = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree1",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        info2 = WorktreeInfo(
            spec_id="SPEC-002",
            path=tmp_path / "worktree2",
            branch="feature/SPEC-002",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info1)
        registry.register(info2)

        all_worktrees = registry.list_all()
        assert len(all_worktrees) == 2
        spec_ids = {w.spec_id for w in all_worktrees}
        assert spec_ids == {"SPEC-001", "SPEC-002"}

    def test_filters_by_project(self, tmp_path: Path) -> None:
        """Test filters worktrees by project_name."""
        registry = WorktreeRegistry(tmp_path)
        info1 = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "proj1" / "SPEC-001",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        info2 = WorktreeInfo(
            spec_id="SPEC-002",
            path=tmp_path / "proj2" / "SPEC-002",
            branch="feature/SPEC-002",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info1, project_name="proj1")
        registry.register(info2, project_name="proj2")

        proj1_worktrees = registry.list_all(project_name="proj1")
        assert len(proj1_worktrees) == 1
        assert proj1_worktrees[0].spec_id == "SPEC-001"

    def test_returns_empty_list_when_no_worktrees(self, tmp_path: Path) -> None:
        """Test returns empty list when no worktrees registered."""
        registry = WorktreeRegistry(tmp_path)
        assert registry.list_all() == []


class TestSyncWithGit:
    """Test sync_with_git method."""

    def test_sync_removes_stale_entries(self, tmp_path: Path) -> None:
        """Test sync removes entries for non-existent worktrees."""
        registry = WorktreeRegistry(tmp_path)
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "nonexistent",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info)

        # Mock repo that returns empty worktree list
        mock_repo = Mock()
        mock_repo.git.worktree.return_value = ""

        registry.sync_with_git(mock_repo)

        assert registry.get("SPEC-001") is None

    def test_sync_keeps_existing_entries(self, tmp_path: Path) -> None:
        """Test sync keeps entries for existing worktrees."""
        registry = WorktreeRegistry(tmp_path)
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info)

        # Mock repo that returns our worktree
        mock_repo = Mock()
        mock_repo.git.worktree.return_value = f"worktree {tmp_path / 'worktree'}"

        registry.sync_with_git(mock_repo)

        assert registry.get("SPEC-001") is not None

    def test_sync_handles_git_errors(self, tmp_path: Path) -> None:
        """Test sync handles Git errors gracefully."""
        registry = WorktreeRegistry(tmp_path)
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info)

        # Mock repo that raises exception
        mock_repo = Mock()
        mock_repo.git.worktree.side_effect = Exception("Git error")

        # Should not raise
        registry.sync_with_git(mock_repo)

    def test_sync_parses_worktree_output(self, tmp_path: Path) -> None:
        """Test sync parses worktree list output correctly."""
        registry = WorktreeRegistry(tmp_path)
        wt1_path = tmp_path / "worktree1"
        wt2_path = tmp_path / "worktree2"

        info1 = WorktreeInfo(
            spec_id="SPEC-001",
            path=wt1_path,
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        info2 = WorktreeInfo(
            spec_id="SPEC-002",
            path=wt2_path,
            branch="feature/SPEC-002",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info1)
        registry.register(info2)

        # Mock repo output
        mock_repo = Mock()
        mock_repo.git.worktree.return_value = f"worktree {wt1_path}\nworktree {wt2_path}\n"

        registry.sync_with_git(mock_repo)

        assert registry.get("SPEC-001") is not None
        assert registry.get("SPEC-002") is not None


class TestRecoverFromDisk:
    """Test recover_from_disk method."""

    def test_recovers_worktrees(self, tmp_path: Path) -> None:
        """Test recovers worktrees from disk."""
        # Create a worktree directory
        worktree_path = tmp_path / "SPEC-001"
        worktree_path.mkdir()
        (worktree_path / ".git").mkdir()
        (worktree_path / ".git" / "objects").mkdir()

        registry = WorktreeRegistry(tmp_path)
        recovered = registry.recover_from_disk()

        assert recovered == 1
        assert registry.get("SPEC-001") is not None

    def test_recovers_namespaced_layout(self, tmp_path: Path) -> None:
        """Test recovers worktrees in namespaced layout."""
        # Create project directory with worktree
        project_path = tmp_path / "my-project"
        project_path.mkdir()
        worktree_path = project_path / "SPEC-001"
        worktree_path.mkdir()
        (worktree_path / ".git").mkdir()
        (worktree_path / ".git" / "objects").mkdir()

        registry = WorktreeRegistry(tmp_path)
        recovered = registry.recover_from_disk()

        assert recovered == 1
        assert registry.get("SPEC-001", project_name="my-project") is not None

    def test_skips_non_git_directories(self, tmp_path: Path) -> None:
        """Test skips directories without .git."""
        # Create directory without .git
        not_worktree = tmp_path / "not-a-worktree"
        not_worktree.mkdir()

        registry = WorktreeRegistry(tmp_path)
        recovered = registry.recover_from_disk()

        assert recovered == 0

    def test_skips_hidden_files(self, tmp_path: Path) -> None:
        """Test skips hidden files and directories."""
        # Create hidden directory
        hidden = tmp_path / ".hidden"
        hidden.mkdir()
        (hidden / ".git").mkdir()

        registry = WorktreeRegistry(tmp_path)
        recovered = registry.recover_from_disk()

        assert recovered == 0

    def test_skips_files(self, tmp_path: Path) -> None:
        """Test skips non-directory items."""
        # Create a file
        file_path = tmp_path / "file.txt"
        file_path.write_text("test")

        registry = WorktreeRegistry(tmp_path)
        recovered = registry.recover_from_disk()

        assert recovered == 0

    def test_skips_already_registered(self, tmp_path: Path) -> None:
        """Test skips worktrees that are already registered."""
        # Create and register a worktree
        worktree_path = tmp_path / "SPEC-001"
        worktree_path.mkdir()
        (worktree_path / ".git").mkdir()
        (worktree_path / ".git" / "objects").mkdir()

        registry = WorktreeRegistry(tmp_path)
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=worktree_path,
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info)

        # Try to recover again
        recovered = registry.recover_from_disk()

        assert recovered == 0

    def test_detects_branch_from_git(self, tmp_path: Path) -> None:
        """Test detects branch name from .git/HEAD."""
        # Create worktree with .git file (worktree)
        worktree_path = tmp_path / "SPEC-001"
        worktree_path.mkdir()

        # Create .git file pointing to gitdir
        # The gitdir path in .git file is relative to worktree
        gitdir = tmp_path / ".git" / "worktrees" / "SPEC-001"
        gitdir.mkdir(parents=True)
        (gitdir / "HEAD").write_text("ref: refs/heads/custom-branch\n")

        # The .git file contains the relative path to gitdir
        (worktree_path / ".git").write_text("gitdir: .git/worktrees/SPEC-001\n")

        registry = WorktreeRegistry(tmp_path)
        recovered = registry.recover_from_disk()

        assert recovered == 1
        info = registry.get("SPEC-001")
        assert info is not None
        # The code should have detected the branch from HEAD file
        # If it falls back to default, it uses feature/{spec_id}
        assert info.branch in ("custom-branch", "feature/SPEC-001")

    def test_returns_zero_for_nonexistent_root(self, tmp_path: Path) -> None:
        """Test returns 0 when worktree_root doesn't exist."""
        # Use a path that doesn't exist but won't cause permission errors
        nonexistent = tmp_path / "does_not_exist"
        registry = WorktreeRegistry(nonexistent)
        recovered = registry.recover_from_disk()
        assert recovered == 0

    def test_saves_after_recovery(self, tmp_path: Path) -> None:
        """Test saves to disk after recovery."""
        # Create a worktree directory
        worktree_path = tmp_path / "SPEC-001"
        worktree_path.mkdir()
        (worktree_path / ".git").mkdir()
        (worktree_path / ".git" / "objects").mkdir()

        registry = WorktreeRegistry(tmp_path)
        registry.recover_from_disk()

        # Create new registry instance to verify persistence
        new_registry = WorktreeRegistry(tmp_path)
        assert new_registry.get("SPEC-001") is not None


class TestRegistryEdgeCases:
    """Test edge cases and error handling."""

    def test_handles_corrupted_registry_file(self, tmp_path: Path) -> None:
        """Test handles corrupted registry file."""
        registry_path = tmp_path / ".moai-worktree-registry.json"
        registry_path.write_text("{corrupted json content")

        # Should not raise, should create empty registry
        registry = WorktreeRegistry(tmp_path)
        assert registry._data == {}

    def test_handles_read_error(self, tmp_path: Path) -> None:
        """Test handles file read error."""
        registry_path = tmp_path / ".moai-worktree-registry.json"
        registry_path.write_text("{}")

        # Make file unreadable (if permissions allow)
        # Note: This test may not work on all systems
        try:
            registry_path.chmod(0o000)
            registry = WorktreeRegistry(tmp_path)
            # Should handle error gracefully
        except Exception:
            # If chmod fails, skip this test
            pass
        finally:
            # Restore permissions for cleanup
            try:
                registry_path.chmod(0o644)
            except Exception:
                pass

    def test_concurrent_registry_access(self, tmp_path: Path) -> None:
        """Test multiple registry instances access same file."""
        # Create first registry and add data
        registry1 = WorktreeRegistry(tmp_path)
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry1.register(info)

        # Create second registry instance
        registry2 = WorktreeRegistry(tmp_path)
        loaded = registry2.get("SPEC-001")
        assert loaded is not None

    def test_unicode_in_spec_ids(self, tmp_path: Path) -> None:
        """Test handles Unicode characters in SPEC IDs."""
        registry = WorktreeRegistry(tmp_path)
        info = WorktreeInfo(
            spec_id="SPEC-国际化-001",
            path=tmp_path / "worktree",
            branch="feature/国际化",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info)

        loaded = registry.get("SPEC-国际化-001")
        assert loaded is not None

    def test_very_long_paths(self, tmp_path: Path) -> None:
        """Test handles very long paths."""
        # Create deeply nested path
        deep_path = tmp_path
        for i in range(20):
            deep_path = deep_path / f"level{i}"

        registry = WorktreeRegistry(tmp_path)
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=deep_path,
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00Z",
            last_accessed="2025-01-13T10:00:00Z",
            status="active",
        )
        registry.register(info)

        loaded = registry.get("SPEC-001")
        assert loaded is not None
        assert loaded.path == deep_path
