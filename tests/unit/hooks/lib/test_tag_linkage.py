"""Tests for TAG linkage manager module (T4: Linkage Manager)."""

from pathlib import Path

import pytest

from moai_adk.tag_system import linkage as tag_linkage
from moai_adk.tag_system import validator


class TestLinkageDatabase:
    """Test linkage database operations (T4: TAG↔CODE Mapping)."""

    def test_create_new_database(self, tmp_path):
        """Test creating new linkage database."""
        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        # Database should be created with empty mappings
        assert db_path.exists()

        data = manager._load_database()
        assert data == {"tags": [], "files": {}}

    def test_load_existing_database(self, tmp_path):
        """Test loading existing linkage database."""
        db_path = tmp_path / "linkage.json"

        # Create existing database
        existing_data = {
            "tags": [
                {
                    "spec_id": "SPEC-AUTH-001",
                    "verb": "impl",
                    "file_path": "auth.py",
                    "line": 10,
                }
            ],
            "files": {"auth.py": ["SPEC-AUTH-001"]},
        }
        tag_linkage.atomic_write_json(db_path, existing_data)

        # Load database
        manager = tag_linkage.LinkageManager(db_path)

        # Verify data loaded
        assert len(manager.get_all_tags()) == 1

    def test_add_tag_to_database(self, tmp_path):
        """Test adding TAG to linkage database (T4.1)."""

        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        tag = validator.TAG(
            spec_id="SPEC-AUTH-001",
            verb="impl",
            file_path=Path("auth.py"),
            line=10,
        )

        manager.add_tag(tag)

        # Verify TAG added
        tags = manager.get_all_tags()
        assert len(tags) == 1
        assert tags[0]["spec_id"] == "SPEC-AUTH-001"

    def test_add_multiple_tags(self, tmp_path):
        """Test adding multiple TAGs to database."""

        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        tags = [
            validator.TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl"),
            validator.TAG("SPEC-AUTH-002", Path("test.py"), 5, "verify"),
            validator.TAG("SPEC-AUTH-003", Path("auth.py"), 20, "impl"),
        ]

        for tag in tags:
            manager.add_tag(tag)

        # Verify all TAGs added
        all_tags = manager.get_all_tags()
        assert len(all_tags) == 3


class TestQueryOperations:
    """Test database query operations (T4.2)."""

    def test_get_code_locations_by_spec_id(self, tmp_path):
        """Test querying code locations by SPEC-ID."""

        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        # Add TAGs for same SPEC-ID
        manager.add_tag(validator.TAG("SPEC-AUTH-001", Path("a.py"), 10, "impl"))
        manager.add_tag(validator.TAG("SPEC-AUTH-001", Path("b.py"), 20, "verify"))
        manager.add_tag(validator.TAG("SPEC-AUTH-001", Path("c.py"), 30, "impl"))

        # Query locations
        locations = manager.get_code_locations("SPEC-AUTH-001")

        assert len(locations) == 3
        assert any(loc["file_path"] == "a.py" and loc["line"] == 10 for loc in locations)
        assert any(loc["file_path"] == "b.py" and loc["line"] == 20 for loc in locations)
        assert any(loc["file_path"] == "c.py" and loc["line"] == 30 for loc in locations)

    def test_get_code_locations_nonexistent_spec(self, tmp_path):
        """Test querying nonexistent SPEC-ID."""
        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        locations = manager.get_code_locations("SPEC-NONEXISTENT")

        assert locations == []

    def test_get_tags_by_file(self, tmp_path):
        """Test getting all TAGs for a file."""

        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        # Add multiple TAGs for same file
        manager.add_tag(validator.TAG("SPEC-TAG-001", Path("auth.py"), 10, "impl"))
        manager.add_tag(validator.TAG("SPEC-TAG-002", Path("auth.py"), 20, "verify"))
        manager.add_tag(validator.TAG("SPEC-TAG-003", Path("other.py"), 10, "impl"))

        # Query TAGs for auth.py
        tags = manager.get_tags_by_file(Path("auth.py"))

        assert len(tags) == 2
        assert all(tag["file_path"] == "auth.py" for tag in tags)

    def test_get_all_spec_ids(self, tmp_path):
        """Test getting all unique SPEC-IDs."""

        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        # Add TAGs with duplicate SPEC-IDs
        manager.add_tag(validator.TAG("SPEC-TAG-001", Path("a.py"), 10, "impl"))
        manager.add_tag(validator.TAG("SPEC-TAG-001", Path("b.py"), 10, "verify"))
        manager.add_tag(validator.TAG("SPEC-TAG-002", Path("c.py"), 10, "impl"))

        spec_ids = manager.get_all_spec_ids()

        assert len(spec_ids) == 2
        assert "SPEC-001" in spec_ids
        assert "SPEC-002" in spec_ids


class TestRemoveOperations:
    """Test TAG removal operations (T4.1)."""

    def test_remove_file_tags(self, tmp_path):
        """Test removing all TAGs for a file."""

        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        # Add TAGs for multiple files
        manager.add_tag(validator.TAG("SPEC-TAG-001", Path("deleted.py"), 10, "impl"))
        manager.add_tag(validator.TAG("SPEC-TAG-002", Path("deleted.py"), 10, "impl"))
        manager.add_tag(validator.TAG("SPEC-TAG-003", Path("kept.py"), 10, "impl"))

        # Remove TAGs for deleted.py
        manager.remove_file_tags(Path("deleted.py"))

        # Verify removed
        assert manager.get_tags_by_file(Path("deleted.py")) == []
        assert len(manager.get_tags_by_file(Path("kept.py"))) == 1

    def test_remove_tag(self, tmp_path):
        """Test removing specific TAG."""

        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        tag = validator.TAG("SPEC-TAG-001", Path("test.py"), 10, "impl")
        manager.add_tag(tag)

        # Remove TAG
        manager.remove_tag(tag)

        # Verify removed
        assert manager.get_tags_by_file(Path("test.py")) == []

    def test_clear_database(self, tmp_path):
        """Test clearing entire database."""

        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        # Add TAGs
        manager.add_tag(validator.TAG("SPEC-TAG-001", Path("a.py"), 10, "impl"))
        manager.add_tag(validator.TAG("SPEC-TAG-002", Path("b.py"), 10, "impl"))

        # Clear database
        manager.clear()

        # Verify cleared
        assert manager.get_all_tags() == []
        assert manager.get_all_spec_ids() == []


class TestAtomicWrites:
    """Test atomic write operations (T4.1)."""

    def test_atomic_write_prevents_corruption(self, tmp_path):
        """Test that atomic writes prevent corruption."""

        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        # Add initial TAG
        manager.add_tag(validator.TAG("SPEC-TAG-001", Path("test.py"), 10, "impl"))

        # Verify database valid
        data = manager._load_database()
        assert "tags" in data
        assert "files" in data

    def test_database_persists_across_instances(self, tmp_path):
        """Test that database persists across manager instances."""

        db_path = tmp_path / "linkage.json"

        # Create first instance and add TAG
        manager1 = tag_linkage.LinkageManager(db_path)
        manager1.add_tag(validator.TAG("SPEC-TAG-001", Path("test.py"), 10, "impl"))

        # Create second instance and verify TAG present
        manager2 = tag_linkage.LinkageManager(db_path)
        tags = manager2.get_all_tags()

        assert len(tags) == 1
        assert tags[0]["spec_id"] == "SPEC-001"


class TestOrphanedTags:
    """Test orphaned TAG detection (T4.1)."""

    def test_find_orphaned_tags(self, tmp_path, monkeypatch):
        """Test finding TAGs with nonexistent SPEC documents."""

        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        # Add TAG for nonexistent SPEC
        manager.add_tag(validator.TAG("SPEC-DELETED-001", Path("orphan.py"), 10, "impl"))

        # Mock spec directory to not contain SPEC-DELETED
        def mock_spec_exists(spec_id):
            return False

        monkeypatch.setattr(tag_linkage, "spec_document_exists", mock_spec_exists)

        # Find orphaned TAGs
        orphans = manager.find_orphaned_tags()

        assert len(orphans) == 1
        assert orphans[0]["spec_id"] == "SPEC-DELETED"


class TestPerformance:
    """Test performance for large databases."""

    def test_large_database_performance(self, tmp_path):
        """Test query performance with large database."""
        import time

        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        # Add 10,000 TAGs
        for i in range(10000):
            spec_id = f"SPEC-PERF-{i % 100:03d}"
            manager.add_tag(validator.TAG(spec_id, Path(f"file{i}.py"), 1, "impl"))

        # Query performance
        start = time.time()
        locations = manager.get_code_locations("SPEC-PERF-001")
        elapsed = time.time() - start

        # Should have 100 matches (10000 / 100 unique SPECs)
        assert len(locations) == 100
        assert elapsed < 0.1  # <100ms (T4.1)


class TestBidirectionalMapping:
    """Test bidirectional TAG↔CODE mapping (T4)."""

    def test_spec_to_files_mapping(self, tmp_path):
        """Test SPEC-ID → Files mapping."""

        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        manager.add_tag(validator.TAG("SPEC-TAG-001", Path("a.py"), 10, "impl"))
        manager.add_tag(validator.TAG("SPEC-TAG-001", Path("b.py"), 10, "verify"))

        locations = manager.get_code_locations("SPEC-001")

        assert len(locations) == 2
        assert locations[0]["file_path"] in ["a.py", "b.py"]
        assert locations[1]["file_path"] in ["a.py", "b.py"]

    def test_file_to_tags_mapping(self, tmp_path):
        """Test File → TAGs mapping (reverse lookup)."""

        db_path = tmp_path / "linkage.json"
        manager = tag_linkage.LinkageManager(db_path)

        manager.add_tag(validator.TAG("SPEC-TAG-001", Path("auth.py"), 10, "impl"))
        manager.add_tag(validator.TAG("SPEC-TAG-002", Path("auth.py"), 20, "verify"))
        manager.add_tag(validator.TAG("SPEC-TAG-003", Path("other.py"), 10, "impl"))

        tags = manager.get_tags_by_file(Path("auth.py"))

        assert len(tags) == 2
        assert tags[0]["spec_id"] in ["SPEC-001", "SPEC-002"]
        assert tags[1]["spec_id"] in ["SPEC-001", "SPEC-002"]
