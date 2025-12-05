"""Comprehensive test suite for WorktreeRegistry with 95%+ coverage.

This test module provides extensive coverage of the WorktreeRegistry class,
including all public and private methods, edge cases, exception handling,
file I/O operations, and data validation scenarios.
"""

import json
import pytest
import tempfile
from datetime import datetime
from pathlib import Path
from unittest.mock import MagicMock, patch, mock_open, call

from moai_adk.cli.worktree.registry import WorktreeRegistry
from moai_adk.cli.worktree.models import WorktreeInfo


class TestWorktreeRegistryInitialization:
    """Test WorktreeRegistry initialization scenarios."""

    def test_init_creates_registry_file_in_existing_directory(self, tmp_path):
        """Test registry file creation in existing directory."""
        worktree_root = tmp_path
        registry = WorktreeRegistry(worktree_root)

        assert registry.worktree_root == worktree_root
        assert registry.registry_path == worktree_root / ".moai-worktree-registry.json"
        assert registry.registry_path.exists()
        assert isinstance(registry._data, dict)

    def test_init_creates_nested_directories(self, tmp_path):
        """Test registry creates nested parent directories."""
        nested_path = tmp_path / "a" / "b" / "c" / "d"
        registry = WorktreeRegistry(nested_path)

        assert registry.worktree_root.exists()
        assert registry.registry_path.exists()
        assert registry.worktree_root == nested_path

    def test_init_with_existing_registry_file_loads_data(self, tmp_path):
        """Test that existing registry file is loaded on initialization."""
        worktree_root = tmp_path
        registry_path = worktree_root / ".moai-worktree-registry.json"

        # Create a registry file with data
        test_data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/path/to/worktree",
                "branch": "feature/spec-001",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-02T00:00:00Z",
                "status": "active",
            }
        }
        registry_path.write_text(json.dumps(test_data))

        registry = WorktreeRegistry(worktree_root)

        assert "SPEC-001" in registry._data
        assert registry._data["SPEC-001"]["spec_id"] == "SPEC-001"

    def test_init_with_empty_registry_file(self, tmp_path):
        """Test initialization with empty registry file."""
        worktree_root = tmp_path
        registry_path = worktree_root / ".moai-worktree-registry.json"
        registry_path.write_text("")

        registry = WorktreeRegistry(worktree_root)

        assert isinstance(registry._data, dict)
        assert len(registry._data) == 0

    def test_init_creates_empty_registry_file_if_not_exists(self, tmp_path):
        """Test that empty registry file is created if it doesn't exist."""
        worktree_root = tmp_path
        registry = WorktreeRegistry(worktree_root)

        assert registry.registry_path.exists()
        content = registry.registry_path.read_text()
        assert content == "{}"


class TestLoadMethod:
    """Test the _load method in detail."""

    def test_load_with_invalid_json_initializes_empty(self, tmp_path):
        """Test that invalid JSON is handled gracefully."""
        worktree_root = tmp_path
        registry_path = worktree_root / ".moai-worktree-registry.json"
        registry_path.write_text("{invalid json}")

        registry = WorktreeRegistry(worktree_root)

        assert isinstance(registry._data, dict)
        assert len(registry._data) == 0

    def test_load_with_corrupted_file_initializes_empty(self, tmp_path):
        """Test that corrupted files don't crash initialization."""
        worktree_root = tmp_path
        registry_path = worktree_root / ".moai-worktree-registry.json"
        registry_path.write_text("corrupted content\n\nmore junk")

        registry = WorktreeRegistry(worktree_root)

        assert isinstance(registry._data, dict)

    def test_load_with_whitespace_only_file(self, tmp_path):
        """Test loading file with only whitespace."""
        worktree_root = tmp_path
        registry_path = worktree_root / ".moai-worktree-registry.json"
        registry_path.write_text("   \n\n   \t\t  \n")

        registry = WorktreeRegistry(worktree_root)

        assert isinstance(registry._data, dict)
        assert len(registry._data) == 0

    def test_load_validates_data_structure(self, tmp_path):
        """Test that _load calls _validate_data."""
        worktree_root = tmp_path
        registry_path = worktree_root / ".moai-worktree-registry.json"

        # Write valid data
        valid_data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/path",
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            }
        }
        registry_path.write_text(json.dumps(valid_data))

        registry = WorktreeRegistry(worktree_root)

        assert "SPEC-001" in registry._data


class TestValidateDataMethod:
    """Test the _validate_data method."""

    def test_validate_data_with_non_dict_returns_empty(self, tmp_path):
        """Test validation with non-dict input returns empty dict."""
        registry = WorktreeRegistry(tmp_path)

        result = registry._validate_data([])
        assert result == {}

        result = registry._validate_data("string")
        assert result == {}

        result = registry._validate_data(123)
        assert result == {}

        result = registry._validate_data(None)
        assert result == {}

    def test_validate_data_filters_non_dict_entries(self, tmp_path):
        """Test that non-dict entries are filtered out."""
        registry = WorktreeRegistry(tmp_path)

        raw_data = {
            "SPEC-001": "not a dict",
            "SPEC-002": [],
            "SPEC-003": 123,
        }

        result = registry._validate_data(raw_data)
        assert result == {}

    def test_validate_data_filters_entries_missing_required_fields(self, tmp_path):
        """Test that entries with missing required fields are filtered."""
        registry = WorktreeRegistry(tmp_path)

        raw_data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/path",
                "branch": "main",
                # Missing created_at, last_accessed, status
            },
            "SPEC-002": {
                "spec_id": "SPEC-002",
                # Missing all other fields
            },
        }

        result = registry._validate_data(raw_data)
        assert result == {}

    def test_validate_data_filters_entries_with_non_string_fields(self, tmp_path):
        """Test that entries with non-string required fields are filtered."""
        registry = WorktreeRegistry(tmp_path)

        raw_data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": 123,  # Should be string
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            },
            "SPEC-002": {
                "spec_id": "SPEC-002",
                "path": "/path",
                "branch": ["main"],  # Should be string
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            },
        }

        result = registry._validate_data(raw_data)
        assert result == {}

    def test_validate_data_keeps_valid_entries(self, tmp_path):
        """Test that valid entries are preserved."""
        registry = WorktreeRegistry(tmp_path)

        valid_entry = {
            "spec_id": "SPEC-001",
            "path": "/path",
            "branch": "main",
            "created_at": "2025-01-01T00:00:00Z",
            "last_accessed": "2025-01-01T00:00:00Z",
            "status": "active",
        }

        raw_data = {"SPEC-001": valid_entry}

        result = registry._validate_data(raw_data)
        assert "SPEC-001" in result
        assert result["SPEC-001"] == valid_entry

    def test_validate_data_mixed_valid_invalid_entries(self, tmp_path):
        """Test that only valid entries are kept when mixed."""
        registry = WorktreeRegistry(tmp_path)

        valid_entry = {
            "spec_id": "SPEC-VALID",
            "path": "/path",
            "branch": "main",
            "created_at": "2025-01-01T00:00:00Z",
            "last_accessed": "2025-01-01T00:00:00Z",
            "status": "active",
        }

        raw_data = {
            "SPEC-VALID": valid_entry,
            "SPEC-INVALID": "not a dict",
            "SPEC-MISSING": {"spec_id": "incomplete"},
        }

        result = registry._validate_data(raw_data)
        assert len(result) == 1
        assert "SPEC-VALID" in result
        assert "SPEC-INVALID" not in result
        assert "SPEC-MISSING" not in result


class TestSaveMethod:
    """Test the _save method."""

    def test_save_creates_json_file(self, tmp_path):
        """Test that _save creates a valid JSON file."""
        registry = WorktreeRegistry(tmp_path)
        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/path",
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            }
        }

        registry._save()

        assert registry.registry_path.exists()
        saved_data = json.loads(registry.registry_path.read_text())
        assert saved_data == registry._data

    def test_save_creates_parent_directories(self, tmp_path):
        """Test that _save creates parent directories if needed."""
        nested_path = tmp_path / "a" / "b" / "c"
        registry_path = nested_path / ".moai-worktree-registry.json"

        # Manually set paths without triggering __init__
        registry = WorktreeRegistry.__new__(WorktreeRegistry)
        registry.worktree_root = nested_path
        registry.registry_path = registry_path
        registry._data = {"test": "data"}

        registry._save()

        assert nested_path.exists()
        assert registry_path.exists()

    def test_save_overwrites_existing_file(self, tmp_path):
        """Test that _save overwrites existing registry file."""
        registry = WorktreeRegistry(tmp_path)

        # First save
        registry._data = {"SPEC-001": {"spec_id": "SPEC-001", "path": "/path1"}}
        registry._save()

        # Second save with different data
        registry._data = {"SPEC-002": {"spec_id": "SPEC-002", "path": "/path2"}}
        registry._save()

        saved_data = json.loads(registry.registry_path.read_text())
        assert "SPEC-001" not in saved_data
        assert "SPEC-002" in saved_data

    def test_save_uses_proper_json_formatting(self, tmp_path):
        """Test that saved JSON uses proper indentation."""
        registry = WorktreeRegistry(tmp_path)
        registry._data = {"test": "data"}
        registry._save()

        content = registry.registry_path.read_text()
        assert "\n" in content  # Should have indentation
        assert content.strip().endswith("}")


class TestRegisterMethod:
    """Test the register method."""

    def test_register_adds_worktree_to_data(self, tmp_path):
        """Test that register adds WorktreeInfo to registry."""
        registry = WorktreeRegistry(tmp_path)

        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/path/to/worktree"),
            branch="feature/spec-001",
            created_at="2025-01-01T00:00:00Z",
            last_accessed="2025-01-01T00:00:00Z",
            status="active",
        )

        registry.register(info)

        assert "SPEC-001" in registry._data
        assert registry._data["SPEC-001"]["spec_id"] == "SPEC-001"

    def test_register_persists_to_disk(self, tmp_path):
        """Test that register saves to disk."""
        registry = WorktreeRegistry(tmp_path)

        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/path/to/worktree"),
            branch="feature/spec-001",
            created_at="2025-01-01T00:00:00Z",
            last_accessed="2025-01-01T00:00:00Z",
            status="active",
        )

        registry.register(info)

        # Create new instance to verify persistence
        registry2 = WorktreeRegistry(tmp_path)
        assert "SPEC-001" in registry2._data

    def test_register_overwrites_existing_spec_id(self, tmp_path):
        """Test that register overwrites existing SPEC ID."""
        registry = WorktreeRegistry(tmp_path)

        info1 = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/path1"),
            branch="main",
            created_at="2025-01-01T00:00:00Z",
            last_accessed="2025-01-01T00:00:00Z",
            status="active",
        )

        info2 = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/path2"),
            branch="develop",
            created_at="2025-01-02T00:00:00Z",
            last_accessed="2025-01-02T00:00:00Z",
            status="merged",
        )

        registry.register(info1)
        registry.register(info2)

        assert registry._data["SPEC-001"]["path"] == "/path2"
        assert registry._data["SPEC-001"]["branch"] == "develop"


class TestUnregisterMethod:
    """Test the unregister method."""

    def test_unregister_removes_existing_spec_id(self, tmp_path):
        """Test that unregister removes a spec_id."""
        registry = WorktreeRegistry(tmp_path)

        # Add data first
        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/path",
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            }
        }
        registry._save()

        registry.unregister("SPEC-001")

        assert "SPEC-001" not in registry._data

    def test_unregister_persists_to_disk(self, tmp_path):
        """Test that unregister saves to disk."""
        registry = WorktreeRegistry(tmp_path)

        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/path",
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            }
        }
        registry._save()

        registry.unregister("SPEC-001")

        # Verify persistence
        registry2 = WorktreeRegistry(tmp_path)
        assert "SPEC-001" not in registry2._data

    def test_unregister_nonexistent_spec_id_is_safe(self, tmp_path):
        """Test that unregistering nonexistent spec_id doesn't crash."""
        registry = WorktreeRegistry(tmp_path)

        # Should not raise
        registry.unregister("SPEC-NONEXISTENT")

    def test_unregister_does_not_affect_other_entries(self, tmp_path):
        """Test that unregister only removes target spec_id."""
        registry = WorktreeRegistry(tmp_path)

        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/path1",
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            },
            "SPEC-002": {
                "spec_id": "SPEC-002",
                "path": "/path2",
                "branch": "develop",
                "created_at": "2025-01-02T00:00:00Z",
                "last_accessed": "2025-01-02T00:00:00Z",
                "status": "active",
            },
        }

        registry.unregister("SPEC-001")

        assert "SPEC-001" not in registry._data
        assert "SPEC-002" in registry._data


class TestGetMethod:
    """Test the get method."""

    def test_get_returns_worktree_info_for_existing_spec_id(self, tmp_path):
        """Test that get returns WorktreeInfo for existing spec_id."""
        registry = WorktreeRegistry(tmp_path)

        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/path",
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            }
        }

        result = registry.get("SPEC-001")

        assert result is not None
        assert isinstance(result, WorktreeInfo)
        assert result.spec_id == "SPEC-001"
        assert result.branch == "main"

    def test_get_returns_none_for_nonexistent_spec_id(self, tmp_path):
        """Test that get returns None for nonexistent spec_id."""
        registry = WorktreeRegistry(tmp_path)

        result = registry.get("SPEC-NONEXISTENT")

        assert result is None

    def test_get_converts_path_to_path_object(self, tmp_path):
        """Test that get converts path string to Path object."""
        registry = WorktreeRegistry(tmp_path)

        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/path/to/worktree",
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            }
        }

        result = registry.get("SPEC-001")

        assert isinstance(result.path, Path)
        assert result.path == Path("/path/to/worktree")


class TestListAllMethod:
    """Test the list_all method."""

    def test_list_all_returns_empty_list_for_empty_registry(self, tmp_path):
        """Test that list_all returns empty list for empty registry."""
        registry = WorktreeRegistry(tmp_path)

        result = registry.list_all()

        assert isinstance(result, list)
        assert len(result) == 0

    def test_list_all_returns_all_worktrees(self, tmp_path):
        """Test that list_all returns all registered worktrees."""
        registry = WorktreeRegistry(tmp_path)

        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/path1",
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            },
            "SPEC-002": {
                "spec_id": "SPEC-002",
                "path": "/path2",
                "branch": "develop",
                "created_at": "2025-01-02T00:00:00Z",
                "last_accessed": "2025-01-02T00:00:00Z",
                "status": "merged",
            },
        }

        result = registry.list_all()

        assert len(result) == 2
        spec_ids = {w.spec_id for w in result}
        assert spec_ids == {"SPEC-001", "SPEC-002"}

    def test_list_all_returns_worktree_info_objects(self, tmp_path):
        """Test that list_all returns WorktreeInfo objects."""
        registry = WorktreeRegistry(tmp_path)

        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/path",
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            }
        }

        result = registry.list_all()

        assert all(isinstance(w, WorktreeInfo) for w in result)


class TestSyncWithGitMethod:
    """Test the sync_with_git method."""

    def test_sync_with_git_removes_nonexistent_worktrees(self, tmp_path):
        """Test that sync_with_git removes entries for missing worktrees."""
        registry = WorktreeRegistry(tmp_path)

        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/nonexistent/path",
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            }
        }

        mock_repo = MagicMock()
        mock_repo.git.worktree.return_value = "worktree /real/path\ndetached"

        registry.sync_with_git(mock_repo)

        assert "SPEC-001" not in registry._data

    def test_sync_with_git_keeps_existing_worktrees(self, tmp_path):
        """Test that sync_with_git keeps entries for existing worktrees."""
        registry = WorktreeRegistry(tmp_path)

        path = "/real/path"
        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": path,
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            }
        }

        mock_repo = MagicMock()
        mock_repo.git.worktree.return_value = f"worktree {path}\ndetached"

        registry.sync_with_git(mock_repo)

        assert "SPEC-001" in registry._data

    def test_sync_with_git_handles_multiple_worktrees(self, tmp_path):
        """Test sync_with_git with multiple worktrees."""
        registry = WorktreeRegistry(tmp_path)

        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/path1",
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            },
            "SPEC-002": {
                "spec_id": "SPEC-002",
                "path": "/nonexistent",
                "branch": "develop",
                "created_at": "2025-01-02T00:00:00Z",
                "last_accessed": "2025-01-02T00:00:00Z",
                "status": "active",
            },
        }

        mock_repo = MagicMock()
        mock_repo.git.worktree.return_value = "worktree /path1\ndetached"

        registry.sync_with_git(mock_repo)

        assert "SPEC-001" in registry._data
        assert "SPEC-002" not in registry._data

    def test_sync_with_git_handles_empty_worktree_list(self, tmp_path):
        """Test sync_with_git with empty worktree list."""
        registry = WorktreeRegistry(tmp_path)

        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/path",
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            }
        }

        mock_repo = MagicMock()
        mock_repo.git.worktree.return_value = ""

        registry.sync_with_git(mock_repo)

        assert "SPEC-001" not in registry._data

    def test_sync_with_git_handles_exception(self, tmp_path):
        """Test that sync_with_git handles exceptions gracefully."""
        registry = WorktreeRegistry(tmp_path)

        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": "/path",
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            }
        }

        mock_repo = MagicMock()
        mock_repo.git.worktree.side_effect = Exception("Git error")

        # Should not raise
        registry.sync_with_git(mock_repo)

        # Data should remain unchanged
        assert "SPEC-001" in registry._data

    def test_sync_with_git_skips_invalid_data_entries(self, tmp_path):
        """Test sync_with_git handles invalid data gracefully."""
        registry = WorktreeRegistry(tmp_path)

        registry._data = {
            "SPEC-001": "not a dict",  # Invalid
            "SPEC-002": {"path": "/path"},  # Missing path
        }

        mock_repo = MagicMock()
        mock_repo.git.worktree.return_value = "worktree /real/path\ndetached"

        # Should not raise
        registry.sync_with_git(mock_repo)

    def test_sync_with_git_ignores_non_worktree_lines(self, tmp_path):
        """Test that sync_with_git ignores non-worktree lines."""
        registry = WorktreeRegistry(tmp_path)

        path = "/real/path"
        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                "path": path,
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            }
        }

        mock_repo = MagicMock()
        mock_repo.git.worktree.return_value = f"worktree {path}\ndetached\nbranch /foo/bar"

        registry.sync_with_git(mock_repo)

        assert "SPEC-001" in registry._data


class TestRecoverFromDiskMethod:
    """Test the recover_from_disk method."""

    def test_recover_from_disk_returns_zero_if_root_not_exists(self, tmp_path):
        """Test recover_from_disk returns 0 if root doesn't exist."""
        nonexistent_path = tmp_path / "nonexistent"
        registry = WorktreeRegistry.__new__(WorktreeRegistry)
        registry.worktree_root = nonexistent_path
        registry._data = {}

        result = registry.recover_from_disk()

        assert result == 0

    def test_recover_from_disk_discovers_worktree_directories(self, tmp_path):
        """Test recover_from_disk discovers worktree directories."""
        worktree_root = tmp_path
        worktree_dir = worktree_root / "SPEC-001"
        worktree_dir.mkdir()
        (worktree_dir / ".git").mkdir()

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}  # Clear any initial data

        result = registry.recover_from_disk()

        assert result >= 1
        assert "SPEC-001" in registry._data

    def test_recover_from_disk_skips_hidden_files(self, tmp_path):
        """Test recover_from_disk skips hidden files and directories."""
        worktree_root = tmp_path
        hidden_dir = worktree_root / ".hidden"
        hidden_dir.mkdir()
        (hidden_dir / ".git").mkdir()

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}

        result = registry.recover_from_disk()

        assert result == 0
        assert ".hidden" not in registry._data

    def test_recover_from_disk_skips_non_directories(self, tmp_path):
        """Test recover_from_disk skips non-directory items."""
        worktree_root = tmp_path
        file_path = worktree_root / "file.txt"
        file_path.write_text("content")

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}

        result = registry.recover_from_disk()

        assert result == 0

    def test_recover_from_disk_skips_items_without_git(self, tmp_path):
        """Test recover_from_disk skips directories without .git."""
        worktree_root = tmp_path
        regular_dir = worktree_root / "not_a_worktree"
        regular_dir.mkdir()

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}

        result = registry.recover_from_disk()

        assert result == 0

    def test_recover_from_disk_skips_already_registered(self, tmp_path):
        """Test recover_from_disk skips already registered worktrees."""
        worktree_root = tmp_path
        worktree_dir = worktree_root / "SPEC-001"
        worktree_dir.mkdir()
        (worktree_dir / ".git").mkdir()

        registry = WorktreeRegistry(worktree_root)
        registry._data = {"SPEC-001": {"spec_id": "SPEC-001", "path": str(worktree_dir)}}

        result = registry.recover_from_disk()

        assert result == 0

    def test_recover_from_disk_detects_branch_from_git_file(self, tmp_path):
        """Test recover_from_disk detects branch from .git file."""
        worktree_root = tmp_path
        worktree_dir = worktree_root / "SPEC-001"
        worktree_dir.mkdir()

        git_dir = tmp_path / ".git" / "worktrees" / "SPEC-001"
        git_dir.mkdir(parents=True)
        (git_dir / "HEAD").write_text("ref: refs/heads/feature/spec-001\n")

        git_file = worktree_dir / ".git"
        git_file.write_text(f"gitdir: {git_dir}\n")

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}

        result = registry.recover_from_disk()

        assert result >= 1
        worktree = registry._data.get("SPEC-001")
        assert worktree is not None
        assert worktree["branch"] == "feature/spec-001"

    def test_recover_from_disk_uses_default_branch_on_error(self, tmp_path):
        """Test recover_from_disk uses default branch name on error."""
        worktree_root = tmp_path
        worktree_dir = worktree_root / "SPEC-001"
        worktree_dir.mkdir()

        git_file = worktree_dir / ".git"
        git_file.write_text("gitdir: /nonexistent/path\n")

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}

        result = registry.recover_from_disk()

        assert result >= 1
        worktree = registry._data.get("SPEC-001")
        assert worktree["branch"] == "feature/SPEC-001"

    def test_recover_from_disk_handles_git_directory(self, tmp_path):
        """Test recover_from_disk handles .git as directory (not worktree)."""
        worktree_root = tmp_path
        worktree_dir = worktree_root / "SPEC-001"
        worktree_dir.mkdir()
        (worktree_dir / ".git").mkdir()  # .git as directory, not worktree

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}

        result = registry.recover_from_disk()

        assert result >= 1
        worktree = registry._data.get("SPEC-001")
        assert worktree is not None

    def test_recover_from_disk_persists_recovered_entries(self, tmp_path):
        """Test recover_from_disk persists recovered entries."""
        worktree_root = tmp_path
        worktree_dir = worktree_root / "SPEC-001"
        worktree_dir.mkdir()
        (worktree_dir / ".git").mkdir()

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}

        registry.recover_from_disk()

        # Create new instance and verify persistence
        registry2 = WorktreeRegistry(worktree_root)
        assert "SPEC-001" in registry2._data

    def test_recover_from_disk_recovers_multiple_worktrees(self, tmp_path):
        """Test recover_from_disk recovers multiple worktrees."""
        worktree_root = tmp_path

        for i in range(1, 4):
            worktree_dir = worktree_root / f"SPEC-{i:03d}"
            worktree_dir.mkdir()
            (worktree_dir / ".git").mkdir()

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}

        result = registry.recover_from_disk()

        assert result >= 3
        for i in range(1, 4):
            assert f"SPEC-{i:03d}" in registry._data

    def test_recover_from_disk_sets_status_to_recovered(self, tmp_path):
        """Test recover_from_disk sets status to 'recovered'."""
        worktree_root = tmp_path
        worktree_dir = worktree_root / "SPEC-001"
        worktree_dir.mkdir()
        (worktree_dir / ".git").mkdir()

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}

        registry.recover_from_disk()

        worktree = registry._data.get("SPEC-001")
        assert worktree["status"] == "recovered"

    def test_recover_from_disk_does_not_re_save_if_no_recovery(self, tmp_path):
        """Test recover_from_disk doesn't save if nothing recovered."""
        worktree_root = tmp_path

        registry = WorktreeRegistry(worktree_root)
        initial_mtime = registry.registry_path.stat().st_mtime

        # Simulate nothing to recover
        registry._data = {}
        result = registry.recover_from_disk()

        assert result == 0
        # File should not have been modified (within a reasonable tolerance)
        assert abs(registry.registry_path.stat().st_mtime - initial_mtime) < 1.0


class TestIntegrationScenarios:
    """Integration tests for complex scenarios."""

    def test_full_workflow_register_and_retrieve(self, tmp_path):
        """Test full workflow of registering and retrieving."""
        registry = WorktreeRegistry(tmp_path)

        # Register multiple worktrees
        for i in range(1, 4):
            info = WorktreeInfo(
                spec_id=f"SPEC-{i:03d}",
                path=Path(f"/path/to/worktree-{i}"),
                branch=f"feature/spec-{i}",
                created_at=f"2025-01-{i:02d}T00:00:00Z",
                last_accessed=f"2025-01-{i:02d}T00:00:00Z",
                status="active",
            )
            registry.register(info)

        # Verify all registered
        all_worktrees = registry.list_all()
        assert len(all_worktrees) == 3

        # Retrieve specific worktree
        retrieved = registry.get("SPEC-001")
        assert retrieved.branch == "feature/spec-1"

    def test_persistent_data_across_instances(self, tmp_path):
        """Test data persists across registry instances."""
        # Create first registry and add data
        registry1 = WorktreeRegistry(tmp_path)
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/path/to/worktree"),
            branch="feature/spec-001",
            created_at="2025-01-01T00:00:00Z",
            last_accessed="2025-01-01T00:00:00Z",
            status="active",
        )
        registry1.register(info)

        # Create second registry and verify persistence
        registry2 = WorktreeRegistry(tmp_path)
        assert "SPEC-001" in registry2._data
        retrieved = registry2.get("SPEC-001")
        assert retrieved.spec_id == "SPEC-001"

    def test_corrupted_registry_recovery(self, tmp_path):
        """Test recovery from corrupted registry file."""
        registry_path = tmp_path / ".moai-worktree-registry.json"

        # Write corrupted JSON
        registry_path.write_text("{invalid json: }")

        # Should not crash during initialization
        registry = WorktreeRegistry(tmp_path)
        assert isinstance(registry._data, dict)

        # Should be able to use registry normally
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/path"),
            branch="main",
            created_at="2025-01-01T00:00:00Z",
            last_accessed="2025-01-01T00:00:00Z",
            status="active",
        )
        registry.register(info)

        # New registry should load valid data
        registry2 = WorktreeRegistry(tmp_path)
        assert "SPEC-001" in registry2._data


class TestEdgeCases:
    """Test edge cases and boundary conditions."""

    def test_registry_with_special_characters_in_path(self, tmp_path):
        """Test registry handles special characters in paths."""
        registry = WorktreeRegistry(tmp_path)

        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/path/with spaces/and-dashes/and_underscores"),
            branch="feature/branch",
            created_at="2025-01-01T00:00:00Z",
            last_accessed="2025-01-01T00:00:00Z",
            status="active",
        )
        registry.register(info)

        retrieved = registry.get("SPEC-001")
        assert retrieved.path == Path("/path/with spaces/and-dashes/and_underscores")

    def test_registry_with_unicode_in_branch(self, tmp_path):
        """Test registry handles unicode in branch names."""
        registry = WorktreeRegistry(tmp_path)

        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/path"),
            branch="feature/spec-001-中文",
            created_at="2025-01-01T00:00:00Z",
            last_accessed="2025-01-01T00:00:00Z",
            status="active",
        )
        registry.register(info)

        retrieved = registry.get("SPEC-001")
        assert retrieved.branch == "feature/spec-001-中文"

    def test_registry_with_very_long_spec_id(self, tmp_path):
        """Test registry handles very long spec IDs."""
        registry = WorktreeRegistry(tmp_path)

        long_spec_id = "SPEC-" + "X" * 200

        info = WorktreeInfo(
            spec_id=long_spec_id,
            path=Path("/path"),
            branch="main",
            created_at="2025-01-01T00:00:00Z",
            last_accessed="2025-01-01T00:00:00Z",
            status="active",
        )
        registry.register(info)

        retrieved = registry.get(long_spec_id)
        assert retrieved.spec_id == long_spec_id

    def test_registry_with_large_number_of_entries(self, tmp_path):
        """Test registry handles large number of entries."""
        registry = WorktreeRegistry(tmp_path)

        for i in range(100):
            info = WorktreeInfo(
                spec_id=f"SPEC-{i:05d}",
                path=Path(f"/path/{i}"),
                branch=f"feature/spec-{i}",
                created_at="2025-01-01T00:00:00Z",
                last_accessed="2025-01-01T00:00:00Z",
                status="active",
            )
            registry.register(info)

        all_worktrees = registry.list_all()
        assert len(all_worktrees) == 100

    def test_sync_with_git_skips_entries_without_path_field(self, tmp_path):
        """Test sync_with_git handles data entries missing path field."""
        registry = WorktreeRegistry(tmp_path)

        # Create data with missing 'path' field
        registry._data = {
            "SPEC-001": {
                "spec_id": "SPEC-001",
                # 'path' field is missing
                "branch": "main",
                "created_at": "2025-01-01T00:00:00Z",
                "last_accessed": "2025-01-01T00:00:00Z",
                "status": "active",
            }
        }

        mock_repo = MagicMock()
        mock_repo.git.worktree.return_value = "worktree /real/path\ndetached"

        registry.sync_with_git(mock_repo)

        # Entry without path field should be removed
        assert "SPEC-001" not in registry._data

    def test_recover_from_disk_handles_broken_gitdir_reference(self, tmp_path):
        """Test recover_from_disk handles broken gitdir references."""
        worktree_root = tmp_path
        worktree_dir = worktree_root / "SPEC-001"
        worktree_dir.mkdir()

        # Create .git file with reference to nonexistent gitdir
        git_file = worktree_dir / ".git"
        git_file.write_text("gitdir: /nonexistent/gitdir\n")

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}

        result = registry.recover_from_disk()

        # Should still recover worktree with default branch
        assert result >= 1
        assert "SPEC-001" in registry._data
        assert registry._data["SPEC-001"]["branch"] == "feature/SPEC-001"

    def test_recover_from_disk_handles_malformed_head_file(self, tmp_path):
        """Test recover_from_disk handles malformed HEAD file."""
        worktree_root = tmp_path
        worktree_dir = worktree_root / "SPEC-001"
        worktree_dir.mkdir()

        git_dir = tmp_path / ".git" / "worktrees" / "SPEC-001"
        git_dir.mkdir(parents=True)

        # Create malformed HEAD file (not readable format)
        (git_dir / "HEAD").write_text("corrupt data\ninvalid format\n")

        git_file = worktree_dir / ".git"
        git_file.write_text(f"gitdir: {git_dir}\n")

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}

        result = registry.recover_from_disk()

        # Should still recover but with default branch
        assert result >= 1
        assert "SPEC-001" in registry._data
        assert registry._data["SPEC-001"]["branch"] == "feature/SPEC-001"

    def test_sync_with_git_with_data_having_non_dict_and_missing_path(self, tmp_path):
        """Test sync_with_git handles mix of invalid data types."""
        registry = WorktreeRegistry(tmp_path)

        registry._data = {
            "SPEC-001": "invalid string",
            "SPEC-002": {"no_path_field": "value"},
            "SPEC-003": None,
        }

        mock_repo = MagicMock()
        mock_repo.git.worktree.return_value = "worktree /some/path\ndetached"

        # Should not crash and should clean up invalid data
        registry.sync_with_git(mock_repo)

        assert len(registry._data) == 0

    def test_recover_from_disk_with_gitdir_containing_spaces(self, tmp_path):
        """Test recover_from_disk handles gitdir paths with spaces."""
        worktree_root = tmp_path
        worktree_dir = worktree_root / "SPEC-001"
        worktree_dir.mkdir()

        git_dir = tmp_path / "path with spaces" / ".git" / "worktrees" / "SPEC-001"
        git_dir.mkdir(parents=True)
        (git_dir / "HEAD").write_text("ref: refs/heads/feature/spec-001\n")

        git_file = worktree_dir / ".git"
        git_file.write_text(f"gitdir: {git_dir}\n")

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}

        result = registry.recover_from_disk()

        assert result >= 1
        assert "SPEC-001" in registry._data
        assert registry._data["SPEC-001"]["branch"] == "feature/spec-001"

    def test_recover_from_disk_gitdir_line_without_gitdir_prefix(self, tmp_path):
        """Test recover_from_disk handles .git file with no gitdir line."""
        worktree_root = tmp_path
        worktree_dir = worktree_root / "SPEC-001"
        worktree_dir.mkdir()

        # Create .git file without gitdir line
        git_file = worktree_dir / ".git"
        git_file.write_text("some other content\nno gitdir line\n")

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}

        result = registry.recover_from_disk()

        # Should still recover with default branch
        assert result >= 1
        assert "SPEC-001" in registry._data
        assert registry._data["SPEC-001"]["branch"] == "feature/SPEC-001"

    def test_recover_from_disk_exception_in_branch_detection(self, tmp_path):
        """Test recover_from_disk handles exceptions during branch detection."""
        worktree_root = tmp_path
        worktree_dir = worktree_root / "SPEC-001"
        worktree_dir.mkdir()

        # Create a .git file that will cause an exception when trying to read gitdir
        git_file = worktree_dir / ".git"
        # Write a line that starts with 'gitdir:' but with special characters
        git_file.write_text("gitdir: /path/that/will/\x00cause/error\n")

        registry = WorktreeRegistry(worktree_root)
        registry._data = {}

        # This should not raise an exception
        result = registry.recover_from_disk()

        # Should still recover with default branch despite error
        assert result >= 1
        assert "SPEC-001" in registry._data
        assert registry._data["SPEC-001"]["branch"] == "feature/SPEC-001"

    def test_recover_from_disk_with_unreadable_git_file(self, tmp_path):
        """Test recover_from_disk handles unreadable .git files."""
        worktree_root = tmp_path
        worktree_dir = worktree_root / "SPEC-001"
        worktree_dir.mkdir()

        git_file = worktree_dir / ".git"
        git_file.write_text("gitdir: /valid/path\n")

        # Make the git_file unreadable by changing permissions
        git_file.chmod(0o000)

        try:
            registry = WorktreeRegistry(worktree_root)
            registry._data = {}

            # Should handle permission error gracefully
            result = registry.recover_from_disk()

            # May or may not recover depending on OS permission handling
            # but should not crash
            assert isinstance(result, int)
        finally:
            # Restore permissions for cleanup
            git_file.chmod(0o644)
