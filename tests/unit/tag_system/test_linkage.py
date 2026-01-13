"""Comprehensive TDD tests for TAG linkage manager module.

Targets 100% coverage for linkage.py including:
- spec_document_exists function (lines 24-41)
- LinkageManager initialization (lines 54-61)
- LinkageManager database operations (lines 63-331)
- Error handling and edge cases
"""

import json
from pathlib import Path
from unittest.mock import patch

from moai_adk.tag_system import linkage
from moai_adk.tag_system.linkage import LinkageManager


class TestSpecDocumentExists:
    """Test spec_document_exists function."""

    def test_spec_document_exists_with_valid_spec(self):
        """Test spec_document_exists with existing SPEC directory."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create .moai/specs directory structure
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            specs_dir.mkdir(parents=True)
            spec_dir = specs_dir / "SPEC-AUTH-001"
            spec_dir.mkdir()

            # Change to temp directory
            original_cwd = Path.cwd()
            import os
            os.chdir(temp_dir)

            try:
                result = linkage.spec_document_exists("SPEC-AUTH-001")
                assert result is True
            finally:
                os.chdir(original_cwd)

    def test_spec_document_exists_with_nonexistent_spec(self):
        """Test spec_document_exists with nonexistent SPEC."""
        import tempfile
        import os

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create .moai/specs directory structure
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            specs_dir.mkdir(parents=True)

            original_cwd = Path.cwd()
            os.chdir(temp_dir)

            try:
                result = linkage.spec_document_exists("SPEC-NONEXISTENT-999")
                assert result is False
            finally:
                os.chdir(original_cwd)

    def test_spec_document_exists_without_moai_dir(self):
        """Test spec_document_exists when .moai directory doesn't exist."""
        import tempfile
        import os

        with tempfile.TemporaryDirectory() as temp_dir:
            original_cwd = Path.cwd()
            os.chdir(temp_dir)

            try:
                result = linkage.spec_document_exists("SPEC-AUTH-001")
                assert result is False
            finally:
                os.chdir(original_cwd)

    def test_spec_document_exists_with_specs_dir_not_directory(self):
        """Test spec_document_exists when specs is not a directory."""
        import tempfile
        import os

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create .moai/specs as a file instead of directory
            moai_dir = Path(temp_dir) / ".moai"
            moai_dir.mkdir()
            specs_file = moai_dir / "specs"
            specs_file.write_text("not a directory")

            original_cwd = Path.cwd()
            os.chdir(temp_dir)

            try:
                result = linkage.spec_document_exists("SPEC-AUTH-001")
                assert result is False
            finally:
                os.chdir(original_cwd)

    def test_spec_document_exists_with_spec_file_not_directory(self):
        """Test spec_document_exists when SPEC-ID is a file not directory."""
        import tempfile
        import os

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create .moai/specs/SPEC-AUTH-001 as a file
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            specs_dir.mkdir(parents=True)
            spec_file = specs_dir / "SPEC-AUTH-001"
            spec_file.write_text("not a directory")

            original_cwd = Path.cwd()
            os.chdir(temp_dir)

            try:
                result = linkage.spec_document_exists("SPEC-AUTH-001")
                assert result is False
            finally:
                os.chdir(original_cwd)


class TestLinkageManagerInitialization:
    """Test LinkageManager initialization and database setup."""

    def test_linkage_manager_init_creates_database(self):
        """Test that LinkageManager creates database file on init."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"

            # Database shouldn't exist yet
            assert not db_path.exists()

            # Initialize manager (triggers database creation)
            _manager = linkage.LinkageManager(db_path)

            # Database should now exist
            assert db_path.exists()

            # Should have proper structure
            data = json.loads(db_path.read_text())
            assert "tags" in data
            assert "files" in data
            assert data["tags"] == []
            assert data["files"] == {}

    def test_linkage_manager_init_with_existing_database(self):
        """Test LinkageManager with existing database."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"

            # Create initial database
            initial_data = {
                "tags": [
                    {
                        "spec_id": "SPEC-AUTH-001",
                        "verb": "impl",
                        "file_path": "auth.py",
                        "line": 10
                    }
                ],
                "files": {
                    "auth.py": ["SPEC-AUTH-001"]
                }
            }
            db_path.write_text(json.dumps(initial_data))

            # Initialize manager
            _manager = linkage.LinkageManager(db_path)

            # Should load existing data
            tags = _manager.get_all_tags()
            assert len(tags) == 1
            assert tags[0]["spec_id"] == "SPEC-AUTH-001"

    def test_linkage_manager_init_with_path_object(self):
        """Test LinkageManager initialization with Path object."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            assert manager.db_path == db_path
            assert db_path.exists()

    def test_linkage_manager_init_with_string_path(self):
        """Test LinkageManager initialization with string path."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path_str = str(Path(temp_dir) / "linkage.json")
            # Convert to Path for type safety
            manager = LinkageManager(Path(db_path_str))

            assert manager.db_path == Path(db_path_str)
            assert Path(db_path_str).exists()


class TestLinkageManagerAddTag:
    """Test LinkageManager.add_tag method."""

    def test_add_tag_success(self):
        """Test adding a TAG successfully."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            result = manager.add_tag(tag)

            assert result is True

            # Verify tag was added
            tags = manager.get_all_tags()
            assert len(tags) == 1
            assert tags[0]["spec_id"] == "SPEC-AUTH-001"

    def test_add_tag_duplicate_prevented(self):
        """Test that duplicate TAGs are prevented."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            # Add same tag twice
            result1 = manager.add_tag(tag)
            result2 = manager.add_tag(tag)

            assert result1 is True
            assert result2 is True  # Still returns True

            # Should only have one entry
            tags = manager.get_all_tags()
            assert len(tags) == 1

    def test_add_tag_updates_file_index(self):
        """Test that add_tag updates file index."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-USER-002", Path("auth.py"), 20, "verify")

            manager.add_tag(tag1)
            manager.add_tag(tag2)

            # Load database directly
            data = json.loads(db_path.read_text())

            # File index should exist
            assert "auth.py" in data["files"]
            assert "SPEC-AUTH-001" in data["files"]["auth.py"]
            assert "SPEC-USER-002" in data["files"]["auth.py"]

    def test_add_tag_multiple_files(self):
        """Test adding TAGs from multiple files."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-USER-002", Path("user.py"), 20, "verify")

            manager.add_tag(tag1)
            manager.add_tag(tag2)

            tags = manager.get_all_tags()
            assert len(tags) == 2

            # Both files should be in index
            data = json.loads(db_path.read_text())
            assert "auth.py" in data["files"]
            assert "user.py" in data["files"]

    def test_add_tag_same_spec_multiple_times(self):
        """Test adding same SPEC-ID from multiple locations."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-AUTH-001", Path("user.py"), 20, "verify")

            manager.add_tag(tag1)
            manager.add_tag(tag2)

            # Both TAGs should be present
            tags = manager.get_all_tags()
            assert len(tags) == 2

            # File index should have both files
            data = json.loads(db_path.read_text())
            assert "SPEC-AUTH-001" in data["files"]["auth.py"]
            assert "SPEC-AUTH-001" in data["files"]["user.py"]


class TestLinkageManagerRemoveTag:
    """Test LinkageManager.remove_tag method."""

    def test_remove_tag_success(self):
        """Test removing a TAG successfully."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            manager.add_tag(tag)
            assert len(manager.get_all_tags()) == 1

            result = manager.remove_tag(tag)

            assert result is True
            assert len(manager.get_all_tags()) == 0

    def test_remove_tag_nonexistent(self):
        """Test removing a TAG that doesn't exist."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-NONEXISTENT-999", Path("fake.py"), 99, "impl")

            result = manager.remove_tag(tag)

            # Should still return True (idempotent)
            assert result is True

    def test_remove_tag_updates_file_index(self):
        """Test that remove_tag updates file index."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-USER-002", Path("auth.py"), 20, "verify")

            manager.add_tag(tag1)
            manager.add_tag(tag2)
            manager.remove_tag(tag1)

            # File index should only have SPEC-USER-002
            data = json.loads(db_path.read_text())
            assert data["files"]["auth.py"] == ["SPEC-USER-002"]

    def test_remove_tag_cleans_up_empty_file_entry(self):
        """Test that remove_tag cleans up empty file entries."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            manager.add_tag(tag)
            manager.remove_tag(tag)

            # File entry should be removed
            data = json.loads(db_path.read_text())
            assert "auth.py" not in data["files"]

    def test_remove_tag_only_from_specific_location(self):
        """Test removing TAG from one location keeps it in others."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-AUTH-001", Path("user.py"), 20, "verify")

            manager.add_tag(tag1)
            manager.add_tag(tag2)
            manager.remove_tag(tag1)

            # Should still have one TAG
            tags = manager.get_all_tags()
            assert len(tags) == 1
            assert tags[0]["file_path"] == "user.py"


class TestLinkageManagerRemoveFileTags:
    """Test LinkageManager.remove_file_tags method."""

    def test_remove_file_tags_success(self):
        """Test removing all TAGs for a file."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-USER-002", Path("auth.py"), 20, "verify")
            tag3 = TAG("SPEC-OTHER-003", Path("other.py"), 30, "depends")

            manager.add_tag(tag1)
            manager.add_tag(tag2)
            manager.add_tag(tag3)

            result = manager.remove_file_tags(Path("auth.py"))

            assert result is True

            # auth.py TAGs should be gone
            auth_tags = manager.get_tags_by_file(Path("auth.py"))
            assert len(auth_tags) == 0

            # other.py TAGs should remain
            other_tags = manager.get_tags_by_file(Path("other.py"))
            assert len(other_tags) == 1

    def test_remove_file_tags_nonexistent_file(self):
        """Test removing TAGs for file with no TAGs."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            result = manager.remove_file_tags(Path("nonexistent.py"))

            # Should still return True (idempotent)
            assert result is True

    def test_remove_file_tags_removes_file_index(self):
        """Test that remove_file_tags removes file from index."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            manager.add_tag(tag)
            manager.remove_file_tags(Path("auth.py"))

            # File should be removed from index
            data = json.loads(db_path.read_text())
            assert "auth.py" not in data["files"]


class TestLinkageManagerGetAllTags:
    """Test LinkageManager.get_all_tags method."""

    def test_get_all_tags_empty_database(self):
        """Test get_all_tags with empty database."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            tags = manager.get_all_tags()
            assert tags == []

    def test_get_all_tags_with_tags(self):
        """Test get_all_tags returns all TAGs."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-USER-002", Path("user.py"), 20, "verify")

            manager.add_tag(tag1)
            manager.add_tag(tag2)

            tags = manager.get_all_tags()
            assert len(tags) == 2

    def test_get_all_tags_returns_dictionaries(self):
        """Test that get_all_tags returns dictionaries."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            manager.add_tag(tag)

            tags = manager.get_all_tags()
            assert isinstance(tags[0], dict)
            assert "spec_id" in tags[0]
            assert "verb" in tags[0]
            assert "file_path" in tags[0]
            assert "line" in tags[0]


class TestLinkageManagerGetCodeLocations:
    """Test LinkageManager.get_code_locations method."""

    def test_get_code_locations_existing_spec(self):
        """Test get_code_locations for existing SPEC-ID."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-AUTH-001", Path("user.py"), 20, "verify")

            manager.add_tag(tag1)
            manager.add_tag(tag2)

            locations = manager.get_code_locations("SPEC-AUTH-001")

            assert len(locations) == 2

            # Check location structure
            assert locations[0]["file_path"] == "auth.py"
            assert locations[0]["line"] == 10
            assert locations[0]["verb"] == "impl"

    def test_get_code_locations_nonexistent_spec(self):
        """Test get_code_locations for nonexistent SPEC-ID."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            locations = manager.get_code_locations("SPEC-NONEXISTENT-999")

            assert locations == []

    def test_get_code_locations_excludes_spec_id(self):
        """Test that get_code_locations doesn't include spec_id in results."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            manager.add_tag(tag)

            locations = manager.get_code_locations("SPEC-AUTH-001")

            # Should not have spec_id field
            assert "spec_id" not in locations[0]


class TestLinkageManagerGetTagsByFile:
    """Test LinkageManager.get_tags_by_file method."""

    def test_get_tags_by_file_existing_file(self):
        """Test get_tags_by_file for existing file."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-USER-002", Path("auth.py"), 20, "verify")

            manager.add_tag(tag1)
            manager.add_tag(tag2)

            tags = manager.get_tags_by_file(Path("auth.py"))

            assert len(tags) == 2
            assert tags[0]["spec_id"] == "SPEC-AUTH-001"
            assert tags[1]["spec_id"] == "SPEC-USER-002"

    def test_get_tags_by_file_nonexistent_file(self):
        """Test get_tags_by_file for file with no TAGs."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            tags = manager.get_tags_by_file(Path("nonexistent.py"))

            assert tags == []

    def test_get_tags_by_file_with_path_object(self):
        """Test get_tags_by_file with Path object."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            manager.add_tag(tag)

            tags = manager.get_tags_by_file(Path("auth.py"))
            assert len(tags) == 1

    def test_get_tags_by_file_with_string_path(self):
        """Test get_tags_by_file with string path."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            manager.add_tag(tag)

            # Convert string to Path for type safety
            tags = manager.get_tags_by_file(Path("auth.py"))
            assert len(tags) == 1


class TestLinkageManagerGetAllSpecIds:
    """Test LinkageManager.get_all_spec_ids method."""

    def test_get_all_spec_ids_empty_database(self):
        """Test get_all_spec_ids with empty database."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            spec_ids = manager.get_all_spec_ids()
            assert spec_ids == []

    def test_get_all_spec_ids_returns_unique(self):
        """Test get_all_spec_ids returns unique SPEC-IDs."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-AUTH-001", Path("user.py"), 20, "verify")
            tag3 = TAG("SPEC-USER-002", Path("admin.py"), 30, "depends")

            manager.add_tag(tag1)
            manager.add_tag(tag2)
            manager.add_tag(tag3)

            spec_ids = manager.get_all_spec_ids()

            # Should be unique
            assert len(spec_ids) == 2
            assert "SPEC-AUTH-001" in spec_ids
            assert "SPEC-USER-002" in spec_ids

    def test_get_all_spec_ids_returns_sorted(self):
        """Test get_all_spec_ids returns sorted list."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag1 = TAG("SPEC-USER-002", Path("user.py"), 20, "verify")
            tag2 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag3 = TAG("SPEC-OTHER-003", Path("other.py"), 30, "depends")

            manager.add_tag(tag1)
            manager.add_tag(tag2)
            manager.add_tag(tag3)

            spec_ids = manager.get_all_spec_ids()

            # Should be sorted
            assert spec_ids == ["SPEC-AUTH-001", "SPEC-OTHER-003", "SPEC-USER-002"]


class TestLinkageManagerFindOrphanedTags:
    """Test LinkageManager.find_orphaned_tags method."""

    def test_find_orphaned_tags_with_orphans(self):
        """Test find_orphaned_tags finds orphaned TAGs."""
        import tempfile
        import os

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create .moai/specs with only one SPEC
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            specs_dir.mkdir(parents=True)
            spec_dir = specs_dir / "SPEC-AUTH-001"
            spec_dir.mkdir()

            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            original_cwd = Path.cwd()
            os.chdir(temp_dir)

            try:
                from moai_adk.tag_system.validator import TAG
                tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
                tag2 = TAG("SPEC-ORPHAN-999", Path("orphan.py"), 20, "verify")

                manager.add_tag(tag1)
                manager.add_tag(tag2)

                orphans = manager.find_orphaned_tags()

                # Should find orphaned TAG
                assert len(orphans) == 1
                assert orphans[0]["spec_id"] == "SPEC-ORPHAN-999"

            finally:
                os.chdir(original_cwd)

    def test_find_orphaned_tags_no_orphans(self):
        """Test find_orphaned_tags when all TAGs have SPECs."""
        import tempfile
        import os

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create .moai/specs with all SPECs
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            specs_dir.mkdir(parents=True)
            spec_dir1 = specs_dir / "SPEC-AUTH-001"
            spec_dir2 = specs_dir / "SPEC-USER-002"
            spec_dir1.mkdir()
            spec_dir2.mkdir()

            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            original_cwd = Path.cwd()
            os.chdir(temp_dir)

            try:
                from moai_adk.tag_system.validator import TAG
                tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
                tag2 = TAG("SPEC-USER-002", Path("user.py"), 20, "verify")

                manager.add_tag(tag1)
                manager.add_tag(tag2)

                orphans = manager.find_orphaned_tags()

                # Should not find orphans
                assert len(orphans) == 0

            finally:
                os.chdir(original_cwd)

    def test_find_orphaned_tags_empty_database(self):
        """Test find_orphaned_tags with empty database."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            orphans = manager.find_orphaned_tags()

            assert orphans == []


class TestLinkageManagerClear:
    """Test LinkageManager.clear method."""

    def test_clear_removes_all_tags(self):
        """Test clear removes all TAGs."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-USER-002", Path("user.py"), 20, "verify")

            manager.add_tag(tag1)
            manager.add_tag(tag2)

            assert len(manager.get_all_tags()) == 2

            result = manager.clear()

            assert result is True
            assert len(manager.get_all_tags()) == 0

    def test_clear_resets_database_structure(self):
        """Test clear resets database to initial structure."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            manager.add_tag(tag)
            manager.clear()

            # Database should have proper structure
            data = json.loads(db_path.read_text())
            assert "tags" in data
            assert "files" in data
            assert data["tags"] == []
            assert data["files"] == {}


class TestLinkageManagerDatabaseOperations:
    """Test LinkageManager database loading and writing."""

    def test_load_database_creates_structure_if_missing(self):
        """Test _load_database creates structure if missing."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"

            # Create database with incomplete structure
            db_path.write_text('{"tags": []}')

            manager = linkage.LinkageManager(db_path)

            # Should add missing 'files' key when loading
            # Note: _load_database doesn't write back, so we need to trigger a write
            # by calling a method that uses _load_database
            tags = manager.get_all_tags()
            assert tags == []

            # The structure is corrected in memory when loaded
            data = manager._load_database()
            assert "files" in data

    def test_load_database_handles_invalid_json(self):
        """Test _load_database handles invalid JSON."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"

            # Write invalid JSON
            db_path.write_text("not valid json {")

            manager = linkage.LinkageManager(db_path)

            # Should create new database
            tags = manager.get_all_tags()
            assert tags == []

    def test_load_database_handles_non_dict_data(self):
        """Test _load_database handles non-dict data."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"

            # Write non-dict JSON
            db_path.write_text('["not", "a", "dict"]')

            manager = linkage.LinkageManager(db_path)

            # Should create new database
            tags = manager.get_all_tags()
            assert tags == []

    def test_write_database_atomic(self):
        """Test _write_database uses atomic write."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            # Write test data
            data = {"tags": [{"spec_id": "SPEC-AUTH-001"}], "files": {}}
            result = manager._write_database(data)

            assert result is True

            # Verify file was written
            assert db_path.exists()

            # Verify content
            loaded_data = json.loads(db_path.read_text())
            assert loaded_data["tags"][0]["spec_id"] == "SPEC-AUTH-001"


class TestLinkageManagerErrorHandling:
    """Test LinkageManager error handling."""

    def test_load_database_returns_empty_when_missing_tags_key(self):
        """Test _load_database adds tags key when missing."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"

            # Create database with only files key
            db_path.write_text('{"files": {}}')

            _manager = linkage.LinkageManager(db_path)

            # Should add missing tags key when loaded
            data = _manager._load_database()
            assert "tags" in data

    def test_load_database_returns_empty_when_missing_files_key(self):
        """Test _load_database adds files key when missing."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"

            # Create database with only tags key
            db_path.write_text('{"tags": []}')

            _manager = linkage.LinkageManager(db_path)

            # Should add missing files key when loaded
            data = _manager._load_database()
            assert "files" in data

    def test_load_database_returns_empty_when_nonexistent_file(self):
        """Test _load_database returns empty dict when file doesn't exist."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"

            # Create manager (which creates the file)
            manager = linkage.LinkageManager(db_path)

            # Delete the file to test the non-existent case
            db_path.unlink()

            # Should return empty structure when file doesn't exist
            data = manager._load_database()
            assert data == {"tags": [], "files": {}}

    def test_add_tag_with_exception_returns_false(self):
        """Test add_tag returns False when exception occurs."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            # Mock _write_database to raise exception
            with patch.object(manager, "_write_database", side_effect=Exception("Database error")):
                result = manager.add_tag(tag)
                assert result is False

    def test_remove_tag_with_exception_returns_false(self):
        """Test remove_tag returns False when exception occurs."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            # Add tag first
            manager.add_tag(tag)

            # Mock _write_database to raise exception
            with patch.object(manager, "_write_database", side_effect=Exception("Database error")):
                result = manager.remove_tag(tag)
                assert result is False

    def test_remove_file_tags_with_exception_returns_false(self):
        """Test remove_file_tags returns False when exception occurs."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            # Mock _write_database to raise exception
            with patch.object(manager, "_write_database", side_effect=Exception("Database error")):
                result = manager.remove_file_tags(Path("auth.py"))
                assert result is False

    def test_add_tag_with_error_returns_false(self):
        """Test add_tag returns False on error."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create directory instead of file
            db_path = Path(temp_dir) / "linkage.json"
            db_path.mkdir()

            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            # Should return False on error
            result = manager.add_tag(tag)
            assert result is False

    def test_remove_tag_with_error_returns_false(self):
        """Test remove_tag returns False on error."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create directory instead of file
            db_path = Path(temp_dir) / "linkage.json"
            db_path.mkdir()

            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            # Should return False on error
            result = manager.remove_tag(tag)
            assert result is False

    def test_clear_with_error_returns_false(self):
        """Test clear returns False on error."""
        import tempfile

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create directory instead of file
            db_path = Path(temp_dir) / "linkage.json"
            db_path.mkdir()

            manager = linkage.LinkageManager(db_path)

            # Should return False on error
            result = manager.clear()
            assert result is False


class TestAtomicWriteReexports:
    """Test atomic_write function re-exports."""

    def test_atomic_write_text_reexported(self):
        """Test atomic_write_text is re-exported from linkage."""
        assert hasattr(linkage, "atomic_write_text")
        assert linkage.atomic_write_text == linkage.atomic_ops.atomic_write_text

    def test_atomic_write_json_reexported(self):
        """Test atomic_write_json is re-exported from linkage."""
        assert hasattr(linkage, "atomic_write_json")
        assert linkage.atomic_write_json == linkage.atomic_ops.atomic_write_json
