"""Test suite for SPEC Status Manager

This module implements TDD tests for automated SPEC status updates
from 'draft' to 'completed' after sync operations.
"""

import tempfile
from pathlib import Path

import pytest
import yaml

from moai_adk.core.spec_status_manager import SpecStatusManager


class TestSpecStatusManager:
    """Test cases for SPEC status detection and updates"""

    @pytest.fixture
    def temp_project_dir(self):
        """Create a temporary project directory structure"""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_dir = Path(temp_dir)

            # Create .moai/specs structure
            specs_dir = project_dir / ".moai" / "specs"
            specs_dir.mkdir(parents=True)

            # Create src and tests directories
            (project_dir / "src").mkdir()
            (project_dir / "tests").mkdir()

            yield project_dir

    @pytest.fixture
    def sample_draft_spec(self, temp_project_dir):
        """Create a sample draft SPEC file"""
        spec_dir = temp_project_dir / ".moai" / "specs" / "SPEC-TEST-001"
        spec_dir.mkdir()

        spec_content = {
            "title": "SPEC-TEST-001: Test Feature",
            "version": "0.1.0",
            "status": "draft",
            "date": "2025-11-11",
            "author": "Alfred",
            "category": "feature",
            "priority": "high",
        }

        spec_file = spec_dir / "spec.md"
        with open(spec_file, "w") as f:
            f.write(
                f"---\n{yaml.dump(spec_content)}---\n\n# Test SPEC\n\n## Implementation\n\n# REMOVED_ORPHAN_CODE:TEST-001\n# REMOVED_ORPHAN_TEST:TEST-001\n"
            )

        return spec_file

    @pytest.fixture
    def sample_completed_spec(self, temp_project_dir):
        """Create a sample completed SPEC file"""
        spec_dir = temp_project_dir / ".moai" / "specs" / "SPEC-COMPLETE-001"
        spec_dir.mkdir()

        spec_content = {
            "title": "SPEC-COMPLETE-001: Completed Feature",
            "version": "1.0.0",
            "status": "completed",
            "date": "2025-11-10",
            "author": "Alfred",
            "category": "feature",
            "priority": "high",
        }

        spec_file = spec_dir / "spec.md"
        with open(spec_file, "w") as f:
            f.write(
                f"---\n{yaml.dump(spec_content)}---\n\n# Completed SPEC\n\n## Implementation\n\n# REMOVED_ORPHAN_CODE:COMPLETE-001\n# REMOVED_ORPHAN_TEST:COMPLETE-001\n"
            )

        return spec_file

    @pytest.fixture
    def spec_status_manager(self, temp_project_dir):
        """Create SpecStatusManager instance"""
        return SpecStatusManager(temp_project_dir)

    def test_detect_draft_specs(self, spec_status_manager, sample_draft_spec, sample_completed_spec):
        """Test detection of draft SPEC files"""
        draft_specs = spec_status_manager.detect_draft_specs()

        assert len(draft_specs) == 1
        assert "SPEC-TEST-001" in draft_specs
        assert "SPEC-COMPLETE-001" not in draft_specs

    def test_is_spec_implementation_completed_all_codes_present(self, spec_status_manager, temp_project_dir):
        """Test completion detection when all required codes are present"""
        # Create corresponding code and test files
        spec_dir = temp_project_dir / ".moai" / "specs" / "SPEC-FULL-001"
        spec_dir.mkdir()

        spec_content = {"title": "SPEC-FULL-001: Full Implementation", "version": "0.1.0", "status": "draft"}

        spec_file = spec_dir / "spec.md"
        with open(spec_file, "w") as f:
            f.write(
                f"""---
{yaml.dump(spec_content)}---
# Full Implementation SPEC

## Implementation Plan

# REMOVED_ORPHAN_CODE:FULL-001-001: Main implementation
# REMOVED_ORPHAN_CODE:FULL-001-002: Helper function
# REMOVED_ORPHAN_TEST:FULL-001-001: Unit tests
# REMOVED_ORPHAN_TEST:FULL-001-002: Integration tests
"""
            )

        # Create code files with matching codes
        src_file = temp_project_dir / "src" / "main.py"
        src_file.write_text(
            """
# # REMOVED_ORPHAN_CODE:FULL-001-001
def main_function():
    pass

# # REMOVED_ORPHAN_CODE:FULL-001-002
def helper_function():
    pass
"""
        )

        # Create test files with matching codes
        test_file = temp_project_dir / "tests" / "test_main.py"
        test_file.write_text(
            """
# # REMOVED_ORPHAN_TEST:FULL-001-001
def test_main_function():
    assert main_function()

# # REMOVED_ORPHAN_TEST:FULL-001-002
def test_helper_function():
    assert helper_function()
"""
        )

        # Test completion detection
        is_completed = spec_status_manager.is_spec_implementation_completed("SPEC-FULL-001")
        assert is_completed is True

    def test_is_spec_implementation_completed_missing_codes(self, spec_status_manager, temp_project_dir):
        """Test completion detection when some codes are missing"""
        # Create SPEC with multiple codes
        spec_dir = temp_project_dir / ".moai" / "specs" / "SPEC-PARTIAL-001"
        spec_dir.mkdir()

        spec_content = {"title": "SPEC-PARTIAL-001: Partial Implementation", "version": "0.1.0", "status": "draft"}

        spec_file = spec_dir / "spec.md"
        with open(spec_file, "w") as f:
            f.write(
                f"""---
{yaml.dump(spec_content)}---
# Partial Implementation SPEC

## Implementation Plan

# REMOVED_ORPHAN_CODE:PARTIAL-001-001: Implemented function
# REMOVED_ORPHAN_CODE:PARTIAL-001-002: Missing function
# REMOVED_ORPHAN_TEST:PARTIAL-001-001: Implemented test
"""
            )

        # Create only some of the required files
        src_file = temp_project_dir / "src" / "partial.py"
        src_file.write_text(
            """
# # REMOVED_ORPHAN_CODE:PARTIAL-001-001
def implemented_function():
    pass
# Missing # REMOVED_ORPHAN_CODE:PARTIAL-001-002
"""
        )

        test_file = temp_project_dir / "tests" / "test_partial.py"
        test_file.write_text(
            """
# # REMOVED_ORPHAN_TEST:PARTIAL-001-001
def test_implemented_function():
    assert implemented_function()
"""
        )

        # Test completion detection should be False
        is_completed = spec_status_manager.is_spec_implementation_completed("SPEC-PARTIAL-001")
        assert is_completed is False

    def test_update_spec_status_to_completed(self, spec_status_manager, sample_draft_spec):
        """Test updating SPEC status from draft to completed"""
        # Read initial content
        with open(sample_draft_spec, "r") as f:
            f.read()

        # Update status
        success = spec_status_manager.update_spec_status("SPEC-TEST-001", "completed")
        assert success is True

        # Verify the update
        with open(sample_draft_spec, "r") as f:
            updated_content = f.read()

        assert "status: completed" in updated_content
        assert "status: draft" not in updated_content

        # Version should be bumped to at least 1.0.0
        assert "version:" in updated_content

    def test_update_spec_status_version_bump(self, spec_status_manager, temp_project_dir):
        """Test version bumping when updating status"""
        # Create SPEC with low version number
        spec_dir = temp_project_dir / ".moai" / "specs" / "SPEC-VERSION-001"
        spec_dir.mkdir()

        spec_content = {"title": "SPEC-VERSION-001: Version Test", "version": "0.0.1", "status": "draft"}

        spec_file = spec_dir / "spec.md"
        with open(spec_file, "w") as f:
            f.write(f"---\n{yaml.dump(spec_content)}---\n\n# Version Test\n")

        # Update status
        success = spec_status_manager.update_spec_status("SPEC-VERSION-001", "completed")
        assert success is True

        # Verify version bump
        with open(spec_file, "r") as f:
            updated_content = f.read()

        # Should be bumped to 1.0.0 or higher (check with or without quotes)
        assert (
            'version: "1.0.0"' in updated_content
            or "version: '1.0.0'" in updated_content
            or "version: 1.0.0" in updated_content
        )

    def test_update_spec_status_error_handling(self, spec_status_manager):
        """Test error handling for invalid SPEC IDs"""
        # Test with non-existent SPEC
        success = spec_status_manager.update_spec_status("SPEC-NONEXISTENT", "completed")
        assert success is False

    def test_get_completion_validation_criteria(self, spec_status_manager):
        """Test getting completion validation criteria"""
        criteria = spec_status_manager.get_completion_validation_criteria()

        assert isinstance(criteria, dict)
        assert "min_code_coverage" in criteria
        assert "require_acceptance_criteria" in criteria
        assert "max_open_tasks" in criteria

    def test_validate_spec_for_completion(self, spec_status_manager, temp_project_dir):
        """Test validation criteria for SPEC completion"""
        # Create a well-structured SPEC
        spec_dir = temp_project_dir / ".moai" / "specs" / "SPEC-VALID-001"
        spec_dir.mkdir()

        spec_content = {
            "title": "SPEC-VALID-001: Valid SPEC",
            "version": "0.1.0",
            "status": "draft",
            "date": "2025-11-11",
        }

        spec_file = spec_dir / "spec.md"
        with open(spec_file, "w") as f:
            f.write(
                f"""---
{yaml.dump(spec_content)}---
# Valid SPEC

## Acceptance Criteria
- All functionality implemented
- Tests passing
- Documentation updated

# REMOVED_ORPHAN_CODE:VALID-001-001
# REMOVED_ORPHAN_TEST:VALID-001-001
"""
            )

        # Create implementation files
        (temp_project_dir / "src" / "valid.py").write_text(
            """
# # REMOVED_ORPHAN_CODE:VALID-001-001
def valid_function():
    return True
"""
        )

        (temp_project_dir / "tests" / "test_valid.py").write_text(
            """
# # REMOVED_ORPHAN_TEST:VALID-001-001
def test_valid_function():
    assert valid_function()
"""
        )

        # Test validation
        validation_result = spec_status_manager.validate_spec_for_completion("SPEC-VALID-001")

        assert isinstance(validation_result, dict)
        assert "is_ready" in validation_result
        assert "issues" in validation_result
        assert "criteria_met" in validation_result

    def test_batch_update_completed_specs(self, spec_status_manager, temp_project_dir):
        """Test batch updating of completed SPECs"""
        # Create multiple draft specs
        for i, spec_id in enumerate(["SPEC-BATCH-001", "SPEC-BATCH-002", "SPEC-BATCH-003"]):
            spec_dir = temp_project_dir / ".moai" / "specs" / spec_id
            spec_dir.mkdir()

            spec_content = {"title": f"{spec_id}: Batch Test", "version": "0.1.0", "status": "draft"}

            spec_file = spec_dir / "spec.md"
            with open(spec_file, "w") as f:
                f.write(
                    f"""---
{yaml.dump(spec_content)}---
# Batch Test {i}

## Acceptance Criteria
- Function implemented correctly
- All tests passing

"""
                )

            # Create docs directory and sync report to pass docs_synced check
            (temp_project_dir / "docs").mkdir(exist_ok=True)
            reports_dir = temp_project_dir / ".moai" / "reports"
            reports_dir.mkdir(parents=True, exist_ok=True)

            # Create a recent sync report
            sync_report = reports_dir / "sync-report-2025-11-11.md"
            sync_report.write_text("# Sync Report\n\nSync completed successfully.")

            # Create implementation for first two specs only
            if i < 2:
                (temp_project_dir / "src" / f"batch_{i}.py").write_text(
                    f"""
def batch_function_{i}():
    return True
"""
                )
                (temp_project_dir / "tests" / f"test_batch_{i}.py").write_text(
                    f"""
def test_batch_function_{i}():
    assert batch_function_{i}()
"""
                )

        # Run batch update
        results = spec_status_manager.batch_update_completed_specs()

        assert isinstance(results, dict)
        assert "updated" in results
        assert "failed" in results
        assert "skipped" in results

        # Should update 2 specs, skip 1 (incomplete implementation)
        assert len(results["updated"]) == 2
        assert len(results["skipped"]) == 1

    def test_integration_with_existing_codes(self, spec_status_manager, temp_project_dir):
        """Test integration with existing code system"""
        # Create SPEC with complex code structure
        spec_dir = temp_project_dir / ".moai" / "specs" / "SPEC-INTEGRATION-001"
        spec_dir.mkdir()

        spec_file = spec_dir / "spec.md"
        spec_file.write_text(
            """---
title: SPEC-INTEGRATION-001: Integration Test
version: 0.1.0
status: draft
---
# Integration Test

## Complex Code Structure

# REMOVED_ORPHAN_CODE:INTEGRATION-001-001: Code implementation
# REMOVED_ORPHAN_TEST:INTEGRATION-001-001: Test implementation
"""
        )

        # Test code scanning integration
        completion_result = spec_status_manager.is_spec_implementation_completed("SPEC-INTEGRATION-001")
        assert isinstance(completion_result, bool)
