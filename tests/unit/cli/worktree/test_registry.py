"""
Minimal import and instantiation tests for Worktree Registry.

These tests verify that the module can be imported and basic classes
can be instantiated without errors.
"""

import pytest
import tempfile
import json
from pathlib import Path
from unittest.mock import MagicMock, patch

from moai_adk.cli.worktree.registry import WorktreeRegistry
from moai_adk.cli.worktree.models import WorktreeInfo


class TestImports:
    """Test that all classes can be imported."""

    def test_worktree_registry_importable(self):
        """Test WorktreeRegistry can be imported."""
        assert WorktreeRegistry is not None


class TestWorktreeRegistryInstantiation:
    """Test WorktreeRegistry class instantiation."""

    def test_worktree_registry_init_with_temp_dir(self):
        """Test WorktreeRegistry can be instantiated with a temp directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir)
            registry = WorktreeRegistry(worktree_root)
            assert registry is not None
            assert registry.worktree_root == worktree_root

    def test_worktree_registry_creates_registry_path(self):
        """Test WorktreeRegistry creates registry path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir)
            registry = WorktreeRegistry(worktree_root)
            assert hasattr(registry, "registry_path")
            assert str(registry.registry_path).endswith(".moai-worktree-registry.json")

    def test_worktree_registry_has_data_dict(self):
        """Test WorktreeRegistry has _data dictionary."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir)
            registry = WorktreeRegistry(worktree_root)
            assert hasattr(registry, "_data")
            assert isinstance(registry._data, dict)

    def test_worktree_registry_methods_exist(self):
        """Test WorktreeRegistry has expected methods."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir)
            registry = WorktreeRegistry(worktree_root)
            # Check for expected method names
            assert hasattr(registry, "_load")
            assert hasattr(registry, "_save")
            assert hasattr(registry, "register")
            assert hasattr(registry, "unregister")
            assert hasattr(registry, "get")
            assert hasattr(registry, "list_all")


class TestWorktreeRegistryLoad:
    """Test WorktreeRegistry load functionality."""

    def test_worktree_registry_load_nonexistent(self):
        """Test WorktreeRegistry loads empty dict for nonexistent file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir) / "nonexistent"
            registry = WorktreeRegistry(worktree_root)
            assert isinstance(registry._data, dict)

    def test_worktree_registry_load_empty_file(self):
        """Test WorktreeRegistry loads empty file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir)
            worktree_root.mkdir(exist_ok=True)
            registry_path = worktree_root / ".moai-worktree-registry.json"
            registry_path.write_text("")

            registry = WorktreeRegistry(worktree_root)
            assert isinstance(registry._data, dict)


class TestWorktreeRegistryBasicOperations:
    """Test WorktreeRegistry basic operations."""

    def test_worktree_registry_list_all_empty(self):
        """Test list_all returns empty list for new registry."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir)
            registry = WorktreeRegistry(worktree_root)
            result = registry.list_all()
            assert isinstance(result, list)
            assert len(result) == 0

    def test_worktree_registry_get_nonexistent(self):
        """Test get returns None for nonexistent spec_id."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir)
            registry = WorktreeRegistry(worktree_root)
            result = registry.get("SPEC-001")
            assert result is None

    @patch("moai_adk.cli.worktree.registry.WorktreeInfo")
    def test_worktree_registry_register(self, mock_worktree_info):
        """Test register method exists and is callable."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir)
            registry = WorktreeRegistry(worktree_root)

            # Create mock WorktreeInfo
            mock_info = MagicMock()
            mock_info.spec_id = "SPEC-001"
            mock_info.to_dict.return_value = {"spec_id": "SPEC-001"}

            # Should not raise
            registry.register(mock_info)

    def test_worktree_registry_unregister(self):
        """Test unregister method is callable."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir)
            registry = WorktreeRegistry(worktree_root)
            # Should not raise
            registry.unregister("SPEC-001")

    @patch("moai_adk.cli.worktree.registry.WorktreeInfo")
    def test_worktree_registry_sync_with_git(self, mock_worktree_info):
        """Test sync_with_git method exists and is callable."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir)
            registry = WorktreeRegistry(worktree_root)
            mock_repo = MagicMock()
            # Should not raise
            registry.sync_with_git(mock_repo)


class TestWorktreeRegistryFileHandling:
    """Test WorktreeRegistry file handling."""

    def test_worktree_registry_creates_parent_directory(self):
        """Test WorktreeRegistry creates parent directory if needed."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir) / "subdir" / "nested"
            registry = WorktreeRegistry(worktree_root)
            # Should create parent directories
            assert worktree_root.exists()

    def test_worktree_registry_saves_to_file(self):
        """Test WorktreeRegistry saves registry to file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir)
            registry = WorktreeRegistry(worktree_root)

            # Registry file should exist after instantiation
            assert registry.registry_path.exists()

    def test_worktree_registry_preserves_json_format(self):
        """Test WorktreeRegistry preserves JSON format."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir)
            registry = WorktreeRegistry(worktree_root)

            # Manually add data
            registry._data = {"test": "value"}
            registry._save()

            # Read back and verify JSON
            with open(registry.registry_path, "r") as f:
                content = json.load(f)
                assert content == {"test": "value"}


class TestWorktreeRegistryDataPersistence:
    """Test WorktreeRegistry data persistence."""

    def test_worktree_registry_persistence_between_instances(self):
        """Test WorktreeRegistry persistence between instances."""
        with tempfile.TemporaryDirectory() as tmpdir:
            worktree_root = Path(tmpdir)

            # Create first instance and add data
            registry1 = WorktreeRegistry(worktree_root)
            registry1._data = {"SPEC-001": {"spec_id": "SPEC-001", "path": "/test"}}
            registry1._save()

            # Create second instance and verify data persists
            registry2 = WorktreeRegistry(worktree_root)
            assert "SPEC-001" in registry2._data
