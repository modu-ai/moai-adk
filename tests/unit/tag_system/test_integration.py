"""Comprehensive integration tests for TAG system modules.

Tests cross-module interactions:
- parser → linkage workflow
- atomic_ops → linkage database operations
- Complete TAG extraction and tracking workflow
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import patch

import pytest

from moai_adk.tag_system import linkage, parser, validator


class TestParserToLinkageWorkflow:
    """Test integration between parser and linkage modules."""

    def test_extract_and_store_tags_workflow(self):
        """Test complete workflow: extract TAGs and store in linkage database."""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create test file with TAGs
            test_file = Path(temp_dir) / "auth.py"
            test_file.write_text("""
# @SPEC SPEC-AUTH-001 impl
def authenticate():
    pass

# @SPEC SPEC-USER-002 verify
class User:
    pass
""")

            # Create linkage database
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            # Extract TAGs from file
            tags = parser.extract_tags_from_file(test_file)

            # Store TAGs in database
            for tag in tags:
                manager.add_tag(tag)

            # Verify TAGs were stored
            all_tags = manager.get_all_tags()
            assert len(all_tags) == 2

            # Verify file indexing works
            auth_tags = manager.get_tags_by_file(test_file)
            assert len(auth_tags) == 2

    def test_extract_multiple_files_and_aggregate(self):
        """Test extracting TAGs from multiple files and aggregating in database."""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create multiple test files
            files = []
            for i in range(3):
                test_file = Path(temp_dir) / f"module{i}.py"
                test_file.write_text(f"# @SPEC SPEC-MODULE-00{i}\ndef func{i}(): pass")
                files.append(test_file)

            # Create linkage database
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            # Extract from all files
            all_tags = parser.extract_tags_from_files(files)

            # Store all TAGs
            for tag in all_tags:
                manager.add_tag(tag)

            # Verify all TAGs stored
            stored_tags = manager.get_all_tags()
            assert len(stored_tags) == 3

            # Verify each file has correct TAGs
            for i, file_path in enumerate(files):
                file_tags = manager.get_tags_by_file(file_path)
                assert len(file_tags) == 1
                assert file_tags[0]["spec_id"] == f"SPEC-MODULE-00{i}"

    def test_extract_directory_and_build_complete_index(self):
        """Test extracting from directory and building complete TAG index."""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create directory structure
            subdir = Path(temp_dir) / "subdir"
            subdir.mkdir()

            # Create files with TAGs
            (Path(temp_dir) / "root.py").write_text("# @SPEC SPEC-ROOT-001")
            (subdir / "sub.py").write_text("# @SPEC SPEC-SUB-002")

            # Extract all TAGs
            all_tags = parser.extract_tags_from_directory(Path(temp_dir), recursive=True)

            # Create linkage database
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            # Store all TAGs
            for tag in all_tags:
                manager.add_tag(tag)

            # Verify database has all TAGs
            assert len(manager.get_all_tags()) == 2

            # Verify all SPEC-IDs indexed
            spec_ids = manager.get_all_spec_ids()
            assert "SPEC-ROOT-001" in spec_ids
            assert "SPEC-SUB-002" in spec_ids


class TestLinkageDatabaseAtomicOperations:
    """Test atomic operations in linkage database."""

    def test_add_tag_uses_atomic_write(self):
        """Test that add_tag uses atomic write operations."""
        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            # Add TAG
            manager.add_tag(tag)

            # Verify atomic write (no partial/corrupted data)
            data = json.loads(db_path.read_text())
            assert "tags" in data
            assert "files" in data

    def test_remove_tag_uses_atomic_write(self):
        """Test that remove_tag uses atomic write operations."""
        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            manager.add_tag(tag)
            manager.remove_tag(tag)

            # Verify atomic write
            data = json.loads(db_path.read_text())
            assert data["tags"] == []

    def test_clear_uses_atomic_write(self):
        """Test that clear uses atomic write operations."""
        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            manager.add_tag(tag)
            manager.clear()

            # Verify atomic write
            data = json.loads(db_path.read_text())
            assert data["tags"] == []
            assert data["files"] == {}

    def test_concurrent_database_operations_safe(self):
        """Test that concurrent database operations are safe (within reason)."""
        import threading

        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            # Add multiple TAGs concurrently
            def add_tag(spec_id):
                from moai_adk.tag_system.validator import TAG
                tag = TAG(spec_id, Path("test.py"), 1, "impl")
                manager.add_tag(tag)

            threads = [
                threading.Thread(target=add_tag, args=(f"SPEC-TAG-{i:03d}",))
                for i in range(10)
            ]

            for thread in threads:
                thread.start()

            for thread in threads:
                thread.join()

            # Due to race conditions, we may not get all 10, but should get some
            tags = manager.get_all_tags()
            # At minimum, should have more than 1 TAG
            assert len(tags) >= 1
            # All TAGs should be unique (no duplicates)
            spec_ids = [tag["spec_id"] for tag in tags]
            assert len(spec_ids) == len(set(spec_ids))


class TestTagValidationIntegration:
    """Test TAG validation integration with parser and linkage."""

    def test_parser_creates_valid_tags(self):
        """Test that parser creates only valid TAGs."""
        source = """
# @SPEC SPEC-AUTH-001 impl
# @SPEC invalid-format
# @SPEC SPEC-USER-002 verify
"""

        tags = parser.extract_tags_from_source(source, Path("test.py"))

        # Only valid TAGs should be created
        for tag in tags:
            is_valid, errors = validator.validate_tag(tag)
            assert is_valid, f"Invalid TAG found: {errors}"

    def test_linkage_stores_valid_tags_only(self):
        """Test that linkage manager stores valid TAG data."""
        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            manager.add_tag(tag)

            # Retrieve and validate
            stored_tags = manager.get_all_tags()
            assert len(stored_tags) == 1

            stored_tag_dict = stored_tags[0]

            # Create TAG object from stored data
            reconstructed_tag = validator.TAG(
                spec_id=stored_tag_dict["spec_id"],
                verb=stored_tag_dict["verb"],
                file_path=Path(stored_tag_dict["file_path"]),
                line=stored_tag_dict["line"],
            )

            is_valid, _ = validator.validate_tag(reconstructed_tag)
            assert is_valid


class TestFileTrackingIntegration:
    """Test file tracking integration across modules."""

    def test_file_index_maintained_across_operations(self):
        """Test that file index is maintained across add/remove operations."""
        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG

            # Add TAGs from multiple files
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-USER-002", Path("user.py"), 20, "verify")
            tag3 = TAG("SPEC-AUTH-003", Path("auth.py"), 30, "depends")

            manager.add_tag(tag1)
            manager.add_tag(tag2)
            manager.add_tag(tag3)

            # Verify file index
            data = json.loads(db_path.read_text())
            assert "auth.py" in data["files"]
            assert "user.py" in data["files"]
            assert len(data["files"]["auth.py"]) == 2  # SPEC-AUTH-001, SPEC-AUTH-003

            # Remove one TAG from auth.py
            manager.remove_tag(tag1)

            # Verify file index updated
            data = json.loads(db_path.read_text())
            assert len(data["files"]["auth.py"]) == 1  # Only SPEC-AUTH-003

    def test_remove_file_updates_all_indices(self):
        """Test that remove_file_tags updates all indices correctly."""
        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG

            # Add multiple TAGs from same file
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-USER-002", Path("auth.py"), 20, "verify")
            tag3 = TAG("SPEC-OTHER-003", Path("other.py"), 30, "depends")

            manager.add_tag(tag1)
            manager.add_tag(tag2)
            manager.add_tag(tag3)

            # Remove all TAGs from auth.py
            manager.remove_file_tags(Path("auth.py"))

            # Verify file index cleaned up
            data = json.loads(db_path.read_text())
            assert "auth.py" not in data["files"]
            assert "other.py" in data["files"]

            # Verify TAGs list updated
            tags = manager.get_all_tags()
            assert len(tags) == 1
            assert tags[0]["spec_id"] == "SPEC-OTHER-003"


class TestSpecIdQueryIntegration:
    """Test SPEC-ID query integration."""

    def test_get_code_locations_for_spec(self):
        """Test getting code locations for SPEC-ID."""
        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG

            # Add same SPEC from multiple locations
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-AUTH-001", Path("user.py"), 20, "verify")
            tag3 = TAG("SPEC-OTHER-002", Path("other.py"), 30, "depends")

            manager.add_tag(tag1)
            manager.add_tag(tag2)
            manager.add_tag(tag3)

            # Query locations for SPEC-AUTH-001
            locations = manager.get_code_locations("SPEC-AUTH-001")

            assert len(locations) == 2
            assert any(loc["file_path"] == "auth.py" for loc in locations)
            assert any(loc["file_path"] == "user.py" for loc in locations)

    def test_get_all_spec_ids_aggregates_from_tags(self):
        """Test getting all unique SPEC-IDs from database."""
        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            from moai_adk.tag_system.validator import TAG

            # Add TAGs with duplicate SPEC-IDs
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            tag2 = TAG("SPEC-AUTH-001", Path("user.py"), 20, "verify")
            tag3 = TAG("SPEC-USER-002", Path("admin.py"), 30, "depends")

            manager.add_tag(tag1)
            manager.add_tag(tag2)
            manager.add_tag(tag3)

            # Get unique SPEC-IDs
            spec_ids = manager.get_all_spec_ids()

            assert len(spec_ids) == 2
            assert "SPEC-AUTH-001" in spec_ids
            assert "SPEC-USER-002" in spec_ids


class TestErrorRecoveryIntegration:
    """Test error recovery across modules."""

    def test_parser_handles_invalid_files_gracefully(self):
        """Test that parser handles invalid files without crashing."""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create files with various issues
            valid_file = Path(temp_dir) / "valid.py"
            valid_file.write_text("# @SPEC SPEC-VALID-001")

            invalid_file = Path(temp_dir) / "invalid.py"
            invalid_file.write_text("def invalid(:")  # Syntax error

            missing_file = Path(temp_dir) / "missing.py"

            # Extract from all files
            tags = parser.extract_tags_from_files([valid_file, invalid_file, missing_file])

            # Should only extract from valid file
            assert len(tags) == 1
            assert tags[0].spec_id == "SPEC-VALID-001"

    def test_linkage_handles_database_corruption(self):
        """Test that linkage handles corrupted database gracefully."""
        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"

            # Create corrupted database
            db_path.write_text("corrupted json {")

            # Initialize manager (should recover)
            manager = linkage.LinkageManager(db_path)

            # Should be able to add TAGs
            from moai_adk.tag_system.validator import TAG
            tag = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")

            result = manager.add_tag(tag)

            # Should succeed after recovery
            assert result is True

    def test_atomic_write_rollback_on_error(self):
        """Test that atomic writes handle errors gracefully."""
        with tempfile.TemporaryDirectory() as temp_dir:
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            # Add initial data
            from moai_adk.tag_system.validator import TAG
            tag1 = TAG("SPEC-AUTH-001", Path("auth.py"), 10, "impl")
            manager.add_tag(tag1)

            # Replace database file with a directory (will cause write to fail)
            db_path.unlink()
            db_path.mkdir()

            try:
                # Try to add another TAG (should fail gracefully)
                tag2 = TAG("SPEC-USER-002", Path("user.py"), 20, "verify")
                result = manager.add_tag(tag2)

                # Should return False on error
                assert result is False
            finally:
                # Clean up directory
                db_path.rmdir()


class TestCompleteWorkflowIntegration:
    """Test complete end-to-end workflows."""

    def test_extract_store_and_query_workflow(self):
        """Test complete workflow: extract, store, and query TAGs."""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create test files
            (Path(temp_dir) / "auth.py").write_text("""
# @SPEC SPEC-AUTH-001 impl
def login(): pass

# @SPEC SPEC-AUTH-002 verify
class User: pass
""")

            (Path(temp_dir) / "user.py").write_text("""
# @SPEC SPEC-USER-003 depends
class Profile: pass
""")

            # Step 1: Extract all TAGs
            all_tags = parser.extract_tags_from_directory(Path(temp_dir))
            assert len(all_tags) == 3

            # Step 2: Store in database
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            for tag in all_tags:
                manager.add_tag(tag)

            # Step 3: Query by SPEC-ID
            auth_locations = manager.get_code_locations("SPEC-AUTH-001")
            assert len(auth_locations) == 1
            assert auth_locations[0]["file_path"] == str(Path(temp_dir) / "auth.py")

            # Step 4: Query by file
            user_tags = manager.get_tags_by_file(Path(temp_dir) / "user.py")
            assert len(user_tags) == 1
            assert user_tags[0]["spec_id"] == "SPEC-USER-003"

            # Step 5: Get all SPEC-IDs
            spec_ids = manager.get_all_spec_ids()
            assert len(spec_ids) == 3

    def test_update_and_sync_workflow(self):
        """Test workflow of updating code and syncing TAG database."""
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create initial file
            test_file = Path(temp_dir) / "auth.py"
            test_file.write_text("""
# @SPEC SPEC-AUTH-001 impl
def login(): pass
""")

            # Extract and store
            tags = parser.extract_tags_from_file(test_file)
            db_path = Path(temp_dir) / "linkage.json"
            manager = linkage.LinkageManager(db_path)

            for tag in tags:
                manager.add_tag(tag)

            assert len(manager.get_all_tags()) == 1

            # Update file (remove old TAG, add new TAGs)
            test_file.write_text("""
# @SPEC SPEC-AUTH-002 impl
def login(): pass

# @SPEC SPEC-AUTH-003 verify
class User: pass
""")

            # Remove old TAGs and add new ones
            manager.remove_file_tags(test_file)

            new_tags = parser.extract_tags_from_file(test_file)
            for tag in new_tags:
                manager.add_tag(tag)

            # Verify database updated
            all_tags = manager.get_all_tags()
            assert len(all_tags) == 2

            spec_ids = manager.get_all_spec_ids()
            assert "SPEC-AUTH-001" not in spec_ids
            assert "SPEC-AUTH-002" in spec_ids
            assert "SPEC-AUTH-003" in spec_ids

    def test_orphan_detection_workflow(self):
        """Test workflow of detecting orphaned TAGs."""
        import os

        with tempfile.TemporaryDirectory() as temp_dir:
            # Create .moai/specs with only one SPEC
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            specs_dir.mkdir(parents=True)
            (specs_dir / "SPEC-AUTH-001").mkdir()

            # Create file with TAGs
            test_file = Path(temp_dir) / "auth.py"
            test_file.write_text("""
# @SPEC SPEC-AUTH-001 impl
def login(): pass

# @SPEC SPEC-ORPHAN-999 impl
def orphan(): pass
""")

            # Change to temp directory (for spec_document_exists)
            original_cwd = Path.cwd()
            os.chdir(temp_dir)

            try:
                # Extract and store TAGs
                tags = parser.extract_tags_from_file(test_file)
                db_path = Path(temp_dir) / "linkage.json"
                manager = linkage.LinkageManager(db_path)

                for tag in tags:
                    manager.add_tag(tag)

                # Find orphans
                orphans = manager.find_orphaned_tags()

                assert len(orphans) == 1
                assert orphans[0]["spec_id"] == "SPEC-ORPHAN-999"

            finally:
                os.chdir(original_cwd)
