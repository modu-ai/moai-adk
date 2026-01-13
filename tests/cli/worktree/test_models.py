"""Comprehensive tests for WorktreeInfo data model.

Test Coverage Strategy:
- to_dict() serialization: Valid data, Path object conversion, all fields
- from_dict() deserialization: Valid data, Path reconstruction, all fields
- Round-trip consistency: Serialize then deserialize, verify equality
- Edge cases: Empty strings, special characters in paths, Unicode handling
- Type safety: Invalid types, missing fields, malformed data
"""

from pathlib import Path

import pytest

from moai_adk.cli.worktree.models import WorktreeInfo


class TestWorktreeInfoInit:
    """Test WorktreeInfo initialization and basic properties."""

    def test_init_with_all_fields(self, tmp_path: Path) -> None:
        """Test initialization with all required fields."""
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T12:00:00",
            status="active",
        )

        assert info.spec_id == "SPEC-001"
        assert info.path == tmp_path / "worktree"
        assert info.branch == "feature/SPEC-001"
        assert info.created_at == "2025-01-13T10:00:00"
        assert info.last_accessed == "2025-01-13T12:00:00"
        assert info.status == "active"

    def test_init_with_different_statuses(self, tmp_path: Path) -> None:
        """Test initialization with different status values."""
        statuses = ["active", "merged", "stale", "archived", "deleted"]

        for status in statuses:
            info = WorktreeInfo(
                spec_id="SPEC-001",
                path=tmp_path / "worktree",
                branch="feature/test",
                created_at="2025-01-13T10:00:00",
                last_accessed="2025-01-13T10:00:00",
                status=status,
            )
            assert info.status == status

    def test_init_with_complex_spec_ids(self, tmp_path: Path) -> None:
        """Test initialization with various SPEC ID formats."""
        spec_ids = [
            "SPEC-001",
            "SPEC-AUTH-001",
            "SPEC-FE-2025-001",
            "spec-001",  # lowercase
            "SPEC_001",  # underscore
        ]

        for spec_id in spec_ids:
            info = WorktreeInfo(
                spec_id=spec_id,
                path=tmp_path / "worktree",
                branch=f"feature/{spec_id}",
                created_at="2025-01-13T10:00:00",
                last_accessed="2025-01-13T10:00:00",
                status="active",
            )
            assert info.spec_id == spec_id


class TestWorktreeInfoToDict:
    """Test WorktreeInfo.to_dict() serialization."""

    def test_to_dict_returns_all_fields(self, tmp_path: Path) -> None:
        """Test that to_dict() returns all fields with correct types."""
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T12:00:00",
            status="active",
        )

        result = info.to_dict()

        assert isinstance(result, dict)
        assert set(result.keys()) == {"spec_id", "path", "branch", "created_at", "last_accessed", "status"}

    def test_to_dict_converts_path_to_string(self, tmp_path: Path) -> None:
        """Test that Path objects are converted to strings."""
        worktree_path = tmp_path / "worktrees" / "SPEC-001"
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=worktree_path,
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T12:00:00",
            status="active",
        )

        result = info.to_dict()

        assert isinstance(result["path"], str)
        assert result["path"] == str(worktree_path)

    def test_to_dict_preserves_all_values(self, tmp_path: Path) -> None:
        """Test that all field values are preserved without modification."""
        original_data = {
            "spec_id": "SPEC-AUTH-001",
            "path": tmp_path / "worktrees" / "SPEC-AUTH-001",
            "branch": "feature/SPEC-AUTH-001-user-auth",
            "created_at": "2025-01-13T09:30:15.123456",
            "last_accessed": "2025-01-13T14:22:33.987654",
            "status": "active",
        }

        info = WorktreeInfo(**original_data)
        result = info.to_dict()

        assert result["spec_id"] == original_data["spec_id"]
        assert result["path"] == str(original_data["path"])
        assert result["branch"] == original_data["branch"]
        assert result["created_at"] == original_data["created_at"]
        assert result["last_accessed"] == original_data["last_accessed"]
        assert result["status"] == original_data["status"]

    def test_to_dict_with_special_characters_in_path(self, tmp_path: Path) -> None:
        """Test to_dict() with paths containing special characters."""
        # Create path with spaces and special characters
        special_path = tmp_path / "worktrees" / "SPEC-001 (dev)" / "project files"

        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=special_path,
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T10:00:00",
            status="active",
        )

        result = info.to_dict()

        assert isinstance(result["path"], str)
        assert str(special_path) in result["path"]

    def test_to_dict_with_unicode_in_fields(self, tmp_path: Path) -> None:
        """Test to_dict() with Unicode characters in fields."""
        info = WorktreeInfo(
            spec_id="SPEC-国际化-001",
            path=tmp_path / "worktree",
            branch="feature/国际化",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T12:00:00",
            status="active",
        )

        result = info.to_dict()

        assert result["spec_id"] == "SPEC-国际化-001"
        assert result["branch"] == "feature/国际化"


class TestWorktreeInfoFromDict:
    """Test WorktreeInfo.from_dict() deserialization."""

    def test_from_dict_creates_valid_instance(self, tmp_path: Path) -> None:
        """Test that from_dict() creates a valid WorktreeInfo instance."""
        data = {
            "spec_id": "SPEC-001",
            "path": str(tmp_path / "worktree"),
            "branch": "feature/SPEC-001",
            "created_at": "2025-01-13T10:00:00",
            "last_accessed": "2025-01-13T12:00:00",
            "status": "active",
        }

        info = WorktreeInfo.from_dict(data)

        assert isinstance(info, WorktreeInfo)
        assert info.spec_id == "SPEC-001"
        assert info.path == tmp_path / "worktree"
        assert info.branch == "feature/SPEC-001"
        assert info.created_at == "2025-01-13T10:00:00"
        assert info.last_accessed == "2025-01-13T12:00:00"
        assert info.status == "active"

    def test_from_dict_converts_string_to_path(self, tmp_path: Path) -> None:
        """Test that string paths are converted to Path objects."""
        path_str = str(tmp_path / "worktrees" / "SPEC-001")
        data = {
            "spec_id": "SPEC-001",
            "path": path_str,
            "branch": "feature/SPEC-001",
            "created_at": "2025-01-13T10:00:00",
            "last_accessed": "2025-01-13T12:00:00",
            "status": "active",
        }

        info = WorktreeInfo.from_dict(data)

        assert isinstance(info.path, Path)
        assert info.path == Path(path_str)

    def test_from_dict_with_all_statuses(self, tmp_path: Path) -> None:
        """Test from_dict() with different status values."""
        statuses = ["active", "merged", "stale", "archived", "deleted"]

        for status in statuses:
            data = {
                "spec_id": "SPEC-001",
                "path": str(tmp_path / "worktree"),
                "branch": "feature/test",
                "created_at": "2025-01-13T10:00:00",
                "last_accessed": "2025-01-13T10:00:00",
                "status": status,
            }

            info = WorktreeInfo.from_dict(data)
            assert info.status == status

    def test_from_dict_preserves_timestamps(self, tmp_path: Path) -> None:
        """Test that ISO 8601 timestamps are preserved exactly."""
        timestamps = [
            "2025-01-13T10:00:00",
            "2025-01-13T10:00:00.123456",
            "2025-01-13T10:00:00Z",
            "2025-01-13T10:00:00+00:00",
        ]

        for ts in timestamps:
            data = {
                "spec_id": "SPEC-001",
                "path": str(tmp_path / "worktree"),
                "branch": "feature/test",
                "created_at": ts,
                "last_accessed": ts,
                "status": "active",
            }

            info = WorktreeInfo.from_dict(data)
            assert info.created_at == ts
            assert info.last_accessed == ts


class TestWorktreeInfoRoundTrip:
    """Test round-trip serialization/deserialization consistency."""

    def test_round_trip_preserves_all_data(self, tmp_path: Path) -> None:
        """Test that serializing then deserializing preserves all data."""
        original = WorktreeInfo(
            spec_id="SPEC-FE-2025-001",
            path=tmp_path / "worktrees" / "SPEC-FE-2025-001",
            branch="feature/SPEC-FE-2025-001-user-dashboard",
            created_at="2025-01-13T08:15:30.987654",
            last_accessed="2025-01-13T16:45:22.123456",
            status="active",
        )

        # Serialize to dict
        serialized = original.to_dict()

        # Deserialize from dict
        restored = WorktreeInfo.from_dict(serialized)

        # Verify all fields match
        assert restored.spec_id == original.spec_id
        assert restored.path == original.path
        assert restored.branch == original.branch
        assert restored.created_at == original.created_at
        assert restored.last_accessed == original.last_accessed
        assert restored.status == original.status

    def test_multiple_round_trips(self, tmp_path: Path) -> None:
        """Test that multiple serialize/deserialize cycles remain consistent."""
        original = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T12:00:00",
            status="active",
        )

        current = original

        # Perform multiple round trips
        for _ in range(5):
            serialized = current.to_dict()
            current = WorktreeInfo.from_dict(serialized)

        # Final result should match original
        assert current.spec_id == original.spec_id
        assert current.path == original.path
        assert current.branch == original.branch
        assert current.created_at == original.created_at
        assert current.last_accessed == original.last_accessed
        assert current.status == original.status


class TestWorktreeInfoEdgeCases:
    """Test edge cases and boundary conditions."""

    def test_empty_string_fields(self, tmp_path: Path) -> None:
        """Test handling of empty strings in string fields."""
        info = WorktreeInfo(
            spec_id="",
            path=tmp_path / "worktree",
            branch="",
            created_at="",
            last_accessed="",
            status="",
        )

        assert info.spec_id == ""
        assert info.branch == ""
        assert info.created_at == ""
        assert info.last_accessed == ""
        assert info.status == ""

    def test_very_long_spec_id(self, tmp_path: Path) -> None:
        """Test handling of very long SPEC IDs."""
        long_spec_id = "SPEC-" + "A" * 1000

        info = WorktreeInfo(
            spec_id=long_spec_id,
            path=tmp_path / "worktree",
            branch="feature/test",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T10:00:00",
            status="active",
        )

        assert info.spec_id == long_spec_id

    def test_deeply_nested_path(self, tmp_path: Path) -> None:
        """Test handling of deeply nested directory paths."""
        deep_path = tmp_path
        for i in range(10):
            deep_path = deep_path / f"level{i}"

        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=deep_path,
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T10:00:00",
            status="active",
        )

        assert info.path == deep_path

    def test_path_with_parent_references(self, tmp_path: Path) -> None:
        """Test paths containing parent directory references."""
        # Create a path with ../ components (resolved)
        base = tmp_path / "worktrees"
        test_path = (base / ".." / "worktrees" / "SPEC-001").resolve()

        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=test_path,
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T10:00:00",
            status="active",
        )

        assert info.path == test_path

    def test_unicode_path_on_windows_style(self, tmp_path: Path) -> None:
        """Test paths that might occur on Windows systems."""
        # Simulate Windows-style path components
        windows_like_path = tmp_path / "C:" / "Users" / "test" / "worktree"

        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=windows_like_path,
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T10:00:00",
            status="active",
        )

        result = info.to_dict()
        restored = WorktreeInfo.from_dict(result)

        assert restored.path == windows_like_path


class TestWorktreeInfoErrorCases:
    """Test error handling and invalid input cases."""

    def test_from_dict_missing_field(self, tmp_path: Path) -> None:
        """Test from_dict() with missing required fields."""
        # Missing 'status' field
        incomplete_data = {
            "spec_id": "SPEC-001",
            "path": str(tmp_path / "worktree"),
            "branch": "feature/SPEC-001",
            "created_at": "2025-01-13T10:00:00",
            "last_accessed": "2025-01-13T12:00:00",
            # 'status' is missing
        }

        with pytest.raises(KeyError):
            WorktreeInfo.from_dict(incomplete_data)

    def test_from_dict_with_path_object(self, tmp_path: Path) -> None:
        """Test from_dict() accepts Path objects (Python's Path() handles this)."""
        # Note: Path() accepts both strings and Path objects, so this works
        path_obj = tmp_path / "worktree"
        data = {
            "spec_id": "SPEC-001",
            "path": path_obj,  # Path object instead of string
            "branch": "feature/SPEC-001",
            "created_at": "2025-01-13T10:00:00",
            "last_accessed": "2025-01-13T12:00:00",
            "status": "active",
        }

        # Path() constructor accepts Path objects and converts them
        info = WorktreeInfo.from_dict(data)
        assert info.path == path_obj

    def test_from_dict_empty_dict(self) -> None:
        """Test from_dict() with completely empty dictionary."""
        with pytest.raises(KeyError):
            WorktreeInfo.from_dict({})


class TestWorktreeInfoDataclassBehavior:
    """Test standard dataclass behaviors."""

    def test_equality(self, tmp_path: Path) -> None:
        """Test that two WorktreeInfo instances with same data are equal."""
        info1 = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T12:00:00",
            status="active",
        )

        info2 = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T12:00:00",
            status="active",
        )

        assert info1 == info2

    def test_inequality(self, tmp_path: Path) -> None:
        """Test that WorktreeInfo instances with different data are not equal."""
        info1 = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree1",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T12:00:00",
            status="active",
        )

        info2 = WorktreeInfo(
            spec_id="SPEC-002",
            path=tmp_path / "worktree2",
            branch="feature/SPEC-002",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T12:00:00",
            status="active",
        )

        assert info1 != info2

    def test_repr(self, tmp_path: Path) -> None:
        """Test that repr() provides useful information."""
        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=tmp_path / "worktree",
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T12:00:00",
            status="active",
        )

        repr_str = repr(info)

        assert "WorktreeInfo" in repr_str
        assert "SPEC-001" in repr_str


class TestWorktreeInfoImmutability:
    """Test immutability aspects of WorktreeInfo."""

    def test_path_object_not_affected_by_external_changes(self, tmp_path: Path) -> None:
        """Test that the stored path is independent of external Path object changes."""
        original_path = tmp_path / "original"

        info = WorktreeInfo(
            spec_id="SPEC-001",
            path=original_path,
            branch="feature/SPEC-001",
            created_at="2025-01-13T10:00:00",
            last_accessed="2025-01-13T12:00:00",
            status="active",
        )

        # Store the path reference
        stored_path = info.path

        # Modify the original variable (creates new Path object)
        original_path = tmp_path / "modified"

        # The stored path in info should remain unchanged
        assert stored_path == tmp_path / "original"
        assert info.path == tmp_path / "original"
