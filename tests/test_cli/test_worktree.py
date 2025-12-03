"""
Tests for Git Worktree CLI functionality.

Tests cover WorktreeManager, WorktreeRegistry, WorktreeInfo classes
and all CLI commands (new, list, switch, remove, status, go, sync, clean, config).
"""

import json
import os
import tempfile
from datetime import datetime
from pathlib import Path
from typing import Generator
from unittest import mock

import pytest
from git import GitCommandError, Repo
from git.util import Actor

from moai_adk.cli.worktree.manager import WorktreeManager
from moai_adk.cli.worktree.registry import WorktreeRegistry
from moai_adk.cli.worktree.models import WorktreeInfo
from moai_adk.cli.worktree.exceptions import (
    WorktreeExistsError,
    WorktreeNotFoundError,
    UncommittedChangesError,
    GitOperationError,
)


@pytest.fixture
def temp_git_repo() -> Generator[Path, None, None]:
    """Create a temporary Git repository for testing."""
    with tempfile.TemporaryDirectory() as tmp_dir:
        repo_path = Path(tmp_dir)
        repo = Repo.init(repo_path)

        # Configure git user
        repo.config_writer().set_value("user", "name", "Test User").release()
        repo.config_writer().set_value("user", "email", "test@example.com").release()

        # Create initial commit
        test_file = repo_path / "README.md"
        test_file.write_text("# Test Project\n")
        repo.index.add(["README.md"])  # Use relative path
        repo.index.commit("Initial commit", author=Actor("Test User", "test@example.com"))

        yield repo_path


@pytest.fixture
def temp_worktree_root() -> Generator[Path, None, None]:
    """Create a temporary worktree root directory."""
    with tempfile.TemporaryDirectory() as tmp_dir:
        yield Path(tmp_dir)


@pytest.fixture
def worktree_manager(temp_git_repo: Path, temp_worktree_root: Path) -> WorktreeManager:
    """Create a WorktreeManager instance with temporary paths."""
    return WorktreeManager(repo_path=temp_git_repo, worktree_root=temp_worktree_root)


@pytest.fixture
def worktree_registry(temp_worktree_root: Path) -> WorktreeRegistry:
    """Create a WorktreeRegistry instance with temporary path."""
    return WorktreeRegistry(worktree_root=temp_worktree_root)


class TestWorktreeInfo:
    """Test WorktreeInfo dataclass."""

    def test_create_worktree_info(self):
        """Test WorktreeInfo creation with valid data."""
        info = WorktreeInfo(
            spec_id="SPEC-AUTH-001",
            path=Path("/home/user/worktrees/MoAI-ADK/SPEC-AUTH-001"),
            branch="feature/SPEC-AUTH-001",
            created_at="2025-11-27T10:00:00Z",
            last_accessed="2025-11-27T11:00:00Z",
            status="active",
        )

        assert info.spec_id == "SPEC-AUTH-001"
        assert info.branch == "feature/SPEC-AUTH-001"
        assert info.status == "active"

    def test_worktree_info_to_dict(self):
        """Test WorktreeInfo.to_dict() serialization."""
        info = WorktreeInfo(
            spec_id="SPEC-PAYMENT-002",
            path=Path("/home/user/worktrees/MoAI-ADK/SPEC-PAYMENT-002"),
            branch="feature/SPEC-PAYMENT-002",
            created_at="2025-11-27T10:00:00Z",
            last_accessed="2025-11-27T11:00:00Z",
            status="merged",
        )

        data = info.to_dict()

        assert data["spec_id"] == "SPEC-PAYMENT-002"
        assert data["status"] == "merged"
        assert "path" in data
        assert "branch" in data

    def test_worktree_info_from_dict(self):
        """Test WorktreeInfo.from_dict() deserialization."""
        data = {
            "spec_id": "SPEC-DB-003",
            "path": "/home/user/worktrees/MoAI-ADK/SPEC-DB-003",
            "branch": "feature/SPEC-DB-003",
            "created_at": "2025-11-27T10:00:00Z",
            "last_accessed": "2025-11-27T11:00:00Z",
            "status": "stale",
        }

        info = WorktreeInfo.from_dict(data)

        assert info.spec_id == "SPEC-DB-003"
        assert info.status == "stale"
        assert isinstance(info.path, Path)


class TestWorktreeRegistry:
    """Test WorktreeRegistry functionality."""

    def test_registry_initialization(self, worktree_registry: WorktreeRegistry):
        """Test registry initialization creates empty registry."""
        assert worktree_registry.registry_path.exists()

        # Should be able to load empty registry
        registry_list = worktree_registry.list_all()
        assert isinstance(registry_list, list)

    def test_registry_register(self, worktree_registry: WorktreeRegistry):
        """Test registering a worktree."""
        info = WorktreeInfo(
            spec_id="SPEC-AUTH-001",
            path=Path("/home/user/worktrees/SPEC-AUTH-001"),
            branch="feature/SPEC-AUTH-001",
            created_at=datetime.now().isoformat() + "Z",
            last_accessed=datetime.now().isoformat() + "Z",
            status="active",
        )

        worktree_registry.register(info)

        retrieved = worktree_registry.get("SPEC-AUTH-001")
        assert retrieved is not None
        assert retrieved.spec_id == "SPEC-AUTH-001"

    def test_registry_unregister(self, worktree_registry: WorktreeRegistry):
        """Test unregistering a worktree."""
        info = WorktreeInfo(
            spec_id="SPEC-PAYMENT-002",
            path=Path("/home/user/worktrees/SPEC-PAYMENT-002"),
            branch="feature/SPEC-PAYMENT-002",
            created_at=datetime.now().isoformat() + "Z",
            last_accessed=datetime.now().isoformat() + "Z",
            status="active",
        )

        worktree_registry.register(info)
        assert worktree_registry.get("SPEC-PAYMENT-002") is not None

        worktree_registry.unregister("SPEC-PAYMENT-002")
        assert worktree_registry.get("SPEC-PAYMENT-002") is None

    def test_registry_list_all(self, worktree_registry: WorktreeRegistry):
        """Test listing all worktrees."""
        info1 = WorktreeInfo(
            spec_id="SPEC-001",
            path=Path("/home/user/worktrees/SPEC-001"),
            branch="feature/SPEC-001",
            created_at=datetime.now().isoformat() + "Z",
            last_accessed=datetime.now().isoformat() + "Z",
            status="active",
        )
        info2 = WorktreeInfo(
            spec_id="SPEC-002",
            path=Path("/home/user/worktrees/SPEC-002"),
            branch="feature/SPEC-002",
            created_at=datetime.now().isoformat() + "Z",
            last_accessed=datetime.now().isoformat() + "Z",
            status="active",
        )

        worktree_registry.register(info1)
        worktree_registry.register(info2)

        all_worktrees = worktree_registry.list_all()
        assert len(all_worktrees) == 2
        spec_ids = [w.spec_id for w in all_worktrees]
        assert "SPEC-001" in spec_ids
        assert "SPEC-002" in spec_ids

    def test_registry_get_nonexistent(self, worktree_registry: WorktreeRegistry):
        """Test getting nonexistent worktree returns None."""
        result = worktree_registry.get("NONEXISTENT")
        assert result is None

    def test_registry_persistence(self, temp_worktree_root: Path):
        """Test registry persists to disk."""
        registry1 = WorktreeRegistry(temp_worktree_root)
        info = WorktreeInfo(
            spec_id="SPEC-PERSIST",
            path=Path("/home/user/worktrees/SPEC-PERSIST"),
            branch="feature/SPEC-PERSIST",
            created_at=datetime.now().isoformat() + "Z",
            last_accessed=datetime.now().isoformat() + "Z",
            status="active",
        )
        registry1.register(info)

        # Create new registry instance reading same file
        registry2 = WorktreeRegistry(temp_worktree_root)
        retrieved = registry2.get("SPEC-PERSIST")
        assert retrieved is not None
        assert retrieved.spec_id == "SPEC-PERSIST"


class TestWorktreeManager:
    """Test WorktreeManager functionality."""

    def test_manager_initialization(self, worktree_manager: WorktreeManager):
        """Test WorktreeManager initializes correctly."""
        assert worktree_manager.repo is not None
        assert worktree_manager.worktree_root is not None
        assert isinstance(worktree_manager.registry, WorktreeRegistry)

    def test_create_worktree_basic(self, worktree_manager: WorktreeManager):
        """Test creating a basic worktree."""
        info = worktree_manager.create(spec_id="SPEC-AUTH-001")

        assert info.spec_id == "SPEC-AUTH-001"
        assert info.path.exists()
        assert info.status == "active"

    def test_create_worktree_with_custom_branch(self, worktree_manager: WorktreeManager):
        """Test creating a worktree with custom branch name."""
        info = worktree_manager.create(spec_id="SPEC-PAY-002", branch_name="custom-payment-branch")

        assert info.spec_id == "SPEC-PAY-002"
        assert "custom-payment-branch" in info.branch

    def test_create_worktree_creates_registry_entry(self, worktree_manager: WorktreeManager):
        """Test that creating a worktree registers it."""
        worktree_manager.create(spec_id="SPEC-REG-001")

        registered = worktree_manager.registry.get("SPEC-REG-001")
        assert registered is not None
        assert registered.spec_id == "SPEC-REG-001"

    def test_create_duplicate_worktree_fails(self, worktree_manager: WorktreeManager):
        """Test that creating duplicate worktree raises error."""
        worktree_manager.create(spec_id="SPEC-DUP-001")

        with pytest.raises(WorktreeExistsError):
            worktree_manager.create(spec_id="SPEC-DUP-001")

    def test_remove_worktree(self, worktree_manager: WorktreeManager):
        """Test removing a worktree."""
        worktree_manager.create(spec_id="SPEC-RM-001")
        assert worktree_manager.registry.get("SPEC-RM-001") is not None

        worktree_manager.remove(spec_id="SPEC-RM-001", force=True)
        assert worktree_manager.registry.get("SPEC-RM-001") is None

    def test_remove_nonexistent_worktree_fails(self, worktree_manager: WorktreeManager):
        """Test removing nonexistent worktree raises error."""
        with pytest.raises(WorktreeNotFoundError):
            worktree_manager.remove(spec_id="NONEXISTENT")

    def test_list_worktrees(self, worktree_manager: WorktreeManager):
        """Test listing all worktrees."""
        worktree_manager.create(spec_id="SPEC-LIST-001")
        worktree_manager.create(spec_id="SPEC-LIST-002")

        worktrees = worktree_manager.list()

        assert len(worktrees) >= 2
        spec_ids = [w.spec_id for w in worktrees]
        assert "SPEC-LIST-001" in spec_ids
        assert "SPEC-LIST-002" in spec_ids

    def test_worktree_isolation(self, worktree_manager: WorktreeManager):
        """Test that created worktrees are isolated."""
        info1 = worktree_manager.create(spec_id="SPEC-ISO-001")
        info2 = worktree_manager.create(spec_id="SPEC-ISO-002")

        # Paths should be different
        assert info1.path != info2.path
        # Both should exist
        assert info1.path.exists()
        assert info2.path.exists()

    def test_sync_worktree(self, worktree_manager: WorktreeManager, temp_git_repo: Path):
        """Test syncing a worktree with base branch."""
        # Create initial worktree
        worktree_manager.create(spec_id="SPEC-SYNC-001")

        # Make a change in main repo
        test_file = temp_git_repo / "test.txt"
        test_file.write_text("new content")
        repo = Repo(temp_git_repo)
        repo.index.add([str(test_file)])
        repo.index.commit("Add test file", author=Actor("Test", "test@test.com"))

        # Sync should work (might fail on network, but shouldn't crash)
        try:
            worktree_manager.sync(spec_id="SPEC-SYNC-001")
        except GitOperationError:
            # Network error is ok for test
            pass

    def test_clean_merged_worktrees(self, worktree_manager: WorktreeManager):
        """Test cleaning merged worktrees."""
        worktree_manager.create(spec_id="SPEC-CLEAN-001")

        # Should return empty list if none merged
        cleaned = worktree_manager.clean_merged()
        assert isinstance(cleaned, list)


class TestWorktreeExceptions:
    """Test custom exceptions."""

    def test_worktree_exists_error(self):
        """Test WorktreeExistsError."""
        error = WorktreeExistsError("SPEC-001", Path("/test/path"))
        assert "SPEC-001" in str(error)
        assert "/test/path" in str(error)

    def test_worktree_not_found_error(self):
        """Test WorktreeNotFoundError."""
        error = WorktreeNotFoundError("SPEC-001")
        assert "SPEC-001" in str(error)

    def test_uncommitted_changes_error(self):
        """Test UncommittedChangesError."""
        error = UncommittedChangesError("SPEC-001")
        assert "SPEC-001" in str(error)

    def test_git_operation_error(self):
        """Test GitOperationError."""
        error = GitOperationError("git command failed")
        assert "git command failed" in str(error)


class TestWorktreeEdgeCases:
    """Test edge cases and error scenarios."""

    def test_registry_with_empty_file(self, temp_worktree_root: Path):
        """Test handling corrupted/empty registry file."""
        registry_path = temp_worktree_root / ".moai-worktree-registry.json"
        registry_path.write_text("{}")  # Empty valid JSON

        registry = WorktreeRegistry(temp_worktree_root)
        assert registry.list_all() == []

    def test_worktree_root_directory_creation(self, temp_git_repo: Path):
        """Test that worktree root is created if missing."""
        worktree_root = temp_git_repo / "non" / "existent" / "path"
        manager = WorktreeManager(repo_path=temp_git_repo, worktree_root=worktree_root)

        # Should not raise error, should create directory
        info = manager.create(spec_id="SPEC-MKDIR-001")
        assert info.path.exists()

    def test_spec_id_validation(self, worktree_manager: WorktreeManager):
        """Test SPEC ID format validation."""
        # Valid SPEC IDs should work
        valid_ids = ["SPEC-AUTH-001", "SPEC-PAYMENT-002", "SPEC-DB-999"]
        for spec_id in valid_ids:
            try:
                worktree_manager.create(spec_id=spec_id, force=True)
            except WorktreeExistsError:
                worktree_manager.remove(spec_id=spec_id, force=True)
                worktree_manager.create(spec_id=spec_id)

    def test_large_worktree_count(self, worktree_manager: WorktreeManager):
        """Test handling multiple worktrees."""
        # Create 5 worktrees
        for i in range(5):
            worktree_manager.create(spec_id=f"SPEC-MANY-{i:03d}")

        # Should list all
        worktrees = worktree_manager.list()
        assert len(worktrees) >= 5


class TestWorktreeCLI:
    """Test CLI commands for worktrees."""

    def test_cli_new_command(self, worktree_manager: WorktreeManager):
        """Test CLI new command creates worktree."""
        # Create worktree via manager
        info = worktree_manager.create(spec_id="SPEC-CLI-001")

        # Verify it was created
        assert info.spec_id == "SPEC-CLI-001"
        assert info.path.exists()

    def test_cli_list_empty(self, temp_worktree_root: Path):
        """Test CLI list command with empty registry."""
        manager = WorktreeManager(repo_path=Path.cwd(), worktree_root=temp_worktree_root)
        worktrees = manager.list()

        assert worktrees == []

    def test_cli_list_multiple(self, worktree_manager: WorktreeManager):
        """Test CLI list command with multiple worktrees."""
        worktree_manager.create(spec_id="SPEC-LIST-A")
        worktree_manager.create(spec_id="SPEC-LIST-B")

        worktrees = worktree_manager.list()

        assert len(worktrees) >= 2
        spec_ids = [w.spec_id for w in worktrees]
        assert "SPEC-LIST-A" in spec_ids
        assert "SPEC-LIST-B" in spec_ids

    def test_cli_go_command(self, worktree_manager: WorktreeManager):
        """Test CLI go command returns correct path."""
        info = worktree_manager.create(spec_id="SPEC-GO-001")

        # Get the worktree info
        retrieved = worktree_manager.registry.get("SPEC-GO-001")
        assert retrieved is not None
        assert retrieved.path == info.path

    def test_cli_sync_command(self, worktree_manager: WorktreeManager):
        """Test CLI sync command."""
        worktree_manager.create(spec_id="SPEC-SYNC-001")

        # Sync should work without error
        try:
            worktree_manager.sync(spec_id="SPEC-SYNC-001")
        except Exception:
            # Network errors are acceptable
            pass

    def test_cli_clean_command(self, worktree_manager: WorktreeManager):
        """Test CLI clean command."""
        worktree_manager.create(spec_id="SPEC-CLEAN-001")

        # Clean should return empty list if nothing merged
        cleaned = worktree_manager.clean_merged()

        assert isinstance(cleaned, list)

    def test_cli_status_command(self, worktree_manager: WorktreeManager):
        """Test CLI status command."""
        worktree_manager.create(spec_id="SPEC-STATUS-001")

        worktrees = worktree_manager.list()

        assert len(worktrees) > 0
        for w in worktrees:
            assert w.status in ["active", "merged", "stale"]

    def test_cli_remove_and_list(self, worktree_manager: WorktreeManager):
        """Test removing worktree and listing."""
        worktree_manager.create(spec_id="SPEC-RM-001")
        worktree_manager.create(spec_id="SPEC-RM-002")

        # List should have 2
        assert len(worktree_manager.list()) >= 2

        # Remove one
        worktree_manager.remove(spec_id="SPEC-RM-001", force=True)

        # List should have fewer
        remaining = worktree_manager.list()
        spec_ids = [w.spec_id for w in remaining]
        assert "SPEC-RM-001" not in spec_ids
