"""Tests for worktree manager with project_name namespace support.

These tests validate that the WorktreeManager can organize worktrees
by project name, creating directory structures like:
    /Users/goos/moai/worktrees/{project-name}/SPEC-001
instead of:
    /Users/goos/moai/worktrees/SPEC-001
"""

import json
from pathlib import Path
from typing import Generator

import pytest
from git import Repo

from moai_adk.cli.worktree.manager import WorktreeManager


@pytest.fixture
def temp_repo_dir(tmp_path: Path) -> Generator[Path, None, None]:
    """Create a temporary Git repository for testing.

    Yields:
        Path to the temporary repository.
    """
    repo_dir = tmp_path / "test_repo"
    repo_dir.mkdir(parents=True, exist_ok=True)

    # Initialize Git repository with explicit initial branch name
    repo = Repo.init(repo_dir, initial_branch="main")
    repo.config_writer().set_value("user", "name", "Test User").release()
    repo.config_writer().set_value("user", "email", "test@example.com").release()

    # Create initial commit
    test_file = repo_dir / "README.md"
    test_file.write_text("# Test Repo")
    repo.index.add([str(test_file)])
    repo.index.commit("Initial commit")

    yield repo_dir


@pytest.fixture
def temp_worktree_root(tmp_path: Path) -> Generator[Path, None, None]:
    """Create a temporary worktree root directory.

    Yields:
        Path to the temporary worktree root.
    """
    worktree_root = tmp_path / "worktrees"
    worktree_root.mkdir(parents=True, exist_ok=True)
    yield worktree_root


@pytest.fixture
def manager(temp_repo_dir: Path, temp_worktree_root: Path) -> WorktreeManager:
    """Create a WorktreeManager instance for testing.

    Args:
        temp_repo_dir: Temporary repository directory.
        temp_worktree_root: Temporary worktree root directory.

    Returns:
        WorktreeManager instance.
    """
    return WorktreeManager(repo_path=temp_repo_dir, worktree_root=temp_worktree_root)


class TestWorktreePathWithProjectName:
    """Test worktree path generation with project_name support."""

    def test_worktree_path_includes_project_name(self, manager: WorktreeManager, temp_repo_dir: Path) -> None:
        """Test that worktree path includes project_name as intermediate directory.

        Current behavior (failing):
            /Users/goos/moai/worktrees/SPEC-001

        Expected behavior (passing):
            /Users/goos/moai/worktrees/{project-name}/SPEC-001
        """
        spec_id = "SPEC-001"

        # Create worktree with explicit project_name
        manager_with_project = WorktreeManager(
            repo_path=temp_repo_dir,
            worktree_root=manager.worktree_root,
            project_name="my-project",
        )

        # Create worktree
        info = manager_with_project.create(spec_id=spec_id, base_branch="main")

        # Verify path includes project_name
        expected_path = manager.worktree_root / "my-project" / spec_id
        assert info.path == expected_path, f"Expected path {expected_path}, got {info.path}"

    def test_project_name_defaults_to_repo_name(self, manager: WorktreeManager, temp_repo_dir: Path) -> None:
        """Test that project_name defaults to repository name when not provided.

        When project_name is not explicitly provided, it should:
        1. Auto-detect from the repository directory name
        2. Use detected name for worktree path organization
        """
        spec_id = "SPEC-AUTH-001"

        # Create manager without explicit project_name
        manager_auto = WorktreeManager(
            repo_path=temp_repo_dir,
            worktree_root=manager.worktree_root,
        )

        # Create worktree
        info = manager_auto.create(spec_id=spec_id, base_branch="main")

        # Verify path uses repo name as project name
        repo_name = temp_repo_dir.name  # "test_repo"
        expected_path = manager.worktree_root / repo_name / spec_id
        assert info.path == expected_path, (
            f"Expected path {expected_path}, got {info.path}. Repository name: {repo_name}"
        )

    def test_multiple_projects_namespace_isolation(self, temp_repo_dir: Path, temp_worktree_root: Path) -> None:
        """Test that multiple projects maintain isolated namespace directories.

        Scenario:
        - Project A creates worktree SPEC-001
        - Project B creates worktree SPEC-001
        - Both should exist in separate directories

        Expected structure:
            /worktrees/project-a/SPEC-001
            /worktrees/project-b/SPEC-001
        """
        # Create manager for project A
        manager_a = WorktreeManager(
            repo_path=temp_repo_dir,
            worktree_root=temp_worktree_root,
            project_name="project-a",
        )

        # Create manager for project B
        manager_b = WorktreeManager(
            repo_path=temp_repo_dir,
            worktree_root=temp_worktree_root,
            project_name="project-b",
        )

        # Create worktrees with same spec_id in both projects
        # Note: Each worktree needs a unique branch name to avoid Git conflicts
        spec_id_a = "SPEC-PROJECT-A"
        spec_id_b = "SPEC-PROJECT-B"
        info_a = manager_a.create(spec_id=spec_id_a, base_branch="main")
        info_b = manager_b.create(spec_id=spec_id_b, base_branch="main")

        # Verify paths are isolated
        assert info_a.path == temp_worktree_root / "project-a" / spec_id_a
        assert info_b.path == temp_worktree_root / "project-b" / spec_id_b
        assert info_a.path != info_b.path

        # Verify both paths exist
        assert info_a.path.exists()
        assert info_b.path.exists()

        # Verify both are registered separately
        assert manager_a.registry.get(spec_id_a) is not None
        assert manager_b.registry.get(spec_id_b) is not None

    def test_same_spec_id_different_projects_can_coexist(self, temp_repo_dir: Path, temp_worktree_root: Path) -> None:
        """Test that identical spec_id values can exist across projects."""
        manager_a = WorktreeManager(
            repo_path=temp_repo_dir,
            worktree_root=temp_worktree_root,
            project_name="project-a",
        )
        manager_b = WorktreeManager(
            repo_path=temp_repo_dir,
            worktree_root=temp_worktree_root,
            project_name="project-b",
        )

        spec_id = "SPEC-SHARED-001"
        info_a = manager_a.create(
            spec_id=spec_id,
            branch_name="feature/project-a-SPEC-SHARED-001",
            base_branch="main",
        )
        info_b = manager_b.create(
            spec_id=spec_id,
            branch_name="feature/project-b-SPEC-SHARED-001",
            base_branch="main",
        )

        assert info_a.path == temp_worktree_root / "project-a" / spec_id
        assert info_b.path == temp_worktree_root / "project-b" / spec_id
        assert info_a.path != info_b.path

        assert manager_a.registry.get(spec_id, project_name="project-a") is not None
        assert manager_b.registry.get(spec_id, project_name="project-b") is not None

    def test_worktree_creation_creates_project_namespace_directory(
        self, manager: WorktreeManager, temp_repo_dir: Path
    ) -> None:
        """Test that creating a worktree creates project_name directory if needed.

        When creating a worktree with project_name:
        1. Project namespace directory should be created
        2. Worktree should be created under project namespace
        3. Directory structure should be clean
        """
        spec_id = "SPEC-BUILD-001"

        manager_with_project = WorktreeManager(
            repo_path=temp_repo_dir,
            worktree_root=manager.worktree_root,
            project_name="build-system",
        )

        # Verify project namespace doesn't exist yet
        project_dir = manager.worktree_root / "build-system"
        assert not project_dir.exists()

        # Create worktree
        info = manager_with_project.create(spec_id=spec_id, base_branch="main")

        # Verify project namespace was created
        assert project_dir.exists()
        assert project_dir.is_dir()

        # Verify worktree is under project namespace
        assert info.path.parent == project_dir
        assert info.path.exists()

    def test_backward_compatibility_worktree_without_project_name(
        self, manager: WorktreeManager, temp_repo_dir: Path
    ) -> None:
        """Test that worktrees created without project_name are still accessible.

        This ensures backward compatibility with existing code that
        doesn't use project_name parameter.
        """
        spec_id = "SPEC-LEGACY-001"

        # Create worktree using default behavior (no project_name)
        info = manager.create(spec_id=spec_id, base_branch="main")

        # When no project_name is provided, should use repo name
        repo_name = temp_repo_dir.name
        expected_path = manager.worktree_root / repo_name / spec_id
        assert info.path == expected_path

        # Should be accessible via registry
        retrieved = manager.registry.get(spec_id)
        assert retrieved is not None
        assert retrieved.path == expected_path

    def test_registry_stores_worktree_with_project_namespace(
        self, manager: WorktreeManager, temp_repo_dir: Path
    ) -> None:
        """Test that registry correctly stores worktree info with project_name path.

        Registry should persist worktree information including the full path
        with project_name in the directory structure.
        """
        spec_id = "SPEC-REG-001"

        manager_with_project = WorktreeManager(
            repo_path=temp_repo_dir,
            worktree_root=manager.worktree_root,
            project_name="registry-test",
        )

        # Create worktree
        manager_with_project.create(spec_id=spec_id, base_branch="main")

        # Verify registry contains correct path
        registry_info = manager_with_project.registry.get(spec_id)
        assert registry_info is not None
        assert registry_info.path == manager.worktree_root / "registry-test" / spec_id

        # Verify registry file contains correct path
        registry_file = manager.worktree_root / ".moai-worktree-registry.json"
        assert registry_file.exists()

        with open(registry_file, "r") as f:
            registry_data = json.load(f)
            assert spec_id in registry_data
            assert Path(registry_data[spec_id]["path"]) == manager.worktree_root / "registry-test" / spec_id
