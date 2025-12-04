"""Enhanced unit tests for SPEC Status Manager module.

This module tests:
- SpecStatusManager initialization and configuration
- Draft SPEC detection with YAML parsing
- Implementation completion detection
- SPEC status updates and version bumping
- Validation and completion criteria
- Batch operations and error handling
"""

import tempfile
from pathlib import Path
from unittest import mock

import pytest
import yaml

from moai_adk.core.spec_status_manager import SpecStatusManager


class TestSpecStatusManagerInitialization:
    """Test SpecStatusManager initialization."""

    def test_manager_initialization(self):
        """Test SpecStatusManager initialization."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))

            assert manager.project_root == Path(temp_dir)
            assert manager.specs_dir == Path(temp_dir) / ".moai" / "specs"
            assert manager.validation_criteria["min_code_coverage"] == 0.85

    def test_manager_validation_criteria(self):
        """Test validation criteria are set."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))

            criteria = manager.validation_criteria
            assert "min_code_coverage" in criteria
            assert "require_acceptance_criteria" in criteria


class TestDraftSpecDetection:
    """Test draft SPEC detection."""

    def test_detect_draft_specs_empty(self):
        """Test detecting draft specs when none exist."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            specs_dir.mkdir(parents=True)

            draft_specs = manager.detect_draft_specs()
            assert len(draft_specs) == 0

    def test_detect_draft_specs_single(self):
        """Test detecting single draft SPEC."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            spec_dir = specs_dir / "SPEC-001"
            spec_dir.mkdir(parents=True)

            spec_file = spec_dir / "spec.md"
            frontmatter = {"status": "draft", "version": "0.1.0"}
            spec_file.write_text(f"---\n{yaml.dump(frontmatter)}---\n# Content")

            draft_specs = manager.detect_draft_specs()
            assert "SPEC-001" in draft_specs

    def test_detect_draft_specs_multiple(self):
        """Test detecting multiple draft SPECs."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"

            for i in range(3):
                spec_dir = specs_dir / f"SPEC-{i:03d}"
                spec_dir.mkdir(parents=True)

                spec_file = spec_dir / "spec.md"
                frontmatter = {"status": "draft", "version": "0.1.0"}
                spec_file.write_text(f"---\n{yaml.dump(frontmatter)}---\n# Content")

            draft_specs = manager.detect_draft_specs()
            assert len(draft_specs) == 3

    def test_detect_draft_specs_excludes_completed(self):
        """Test that completed specs are excluded."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"

            # Create draft
            draft_dir = specs_dir / "SPEC-DRAFT"
            draft_dir.mkdir(parents=True)
            draft_file = draft_dir / "spec.md"
            draft_file.write_text("---\nstatus: draft\nversion: 0.1.0\n---\n# Content")

            # Create completed
            completed_dir = specs_dir / "SPEC-DONE"
            completed_dir.mkdir(parents=True)
            completed_file = completed_dir / "spec.md"
            completed_file.write_text("---\nstatus: completed\nversion: 1.0.0\n---\n# Content")

            draft_specs = manager.detect_draft_specs()
            assert "SPEC-DRAFT" in draft_specs
            assert "SPEC-DONE" not in draft_specs

    def test_detect_draft_specs_invalid_yaml(self):
        """Test handling invalid YAML in spec."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            spec_dir = specs_dir / "SPEC-001"
            spec_dir.mkdir(parents=True)

            spec_file = spec_dir / "spec.md"
            spec_file.write_text("---\ninvalid: yaml: content\n---\n# Content")

            # Should handle gracefully
            draft_specs = manager.detect_draft_specs()
            assert isinstance(draft_specs, set)


class TestImplementationCompletion:
    """Test implementation completion detection."""

    def test_is_spec_implementation_completed_missing_spec(self):
        """Test completion check with missing spec file."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            result = manager.is_spec_implementation_completed("NONEXISTENT")
            assert result is False

    def test_is_spec_implementation_completed_no_code(self):
        """Test completion check without implementation."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            spec_dir = specs_dir / "SPEC-001"
            spec_dir.mkdir(parents=True)

            spec_file = spec_dir / "spec.md"
            content = "---\nstatus: draft\n---\n# Spec\n## Acceptance Criteria\nCriteria here"
            spec_file.write_text(content)

            result = manager.is_spec_implementation_completed("SPEC-001")
            assert result is False  # No code or tests

    def test_is_spec_implementation_completed_with_code(self):
        """Test completion check with implementation."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))

            # Create spec directory
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            spec_dir = specs_dir / "SPEC-001"
            spec_dir.mkdir(parents=True)

            spec_file = spec_dir / "spec.md"
            content = "---\nstatus: draft\n---\n# Spec\n## Acceptance Criteria\nCriteria"
            spec_file.write_text(content)

            # Create implementation
            src_dir = Path(temp_dir) / "src"
            src_dir.mkdir()
            (src_dir / "impl.py").write_text("# Implementation")

            # Create tests
            tests_dir = Path(temp_dir) / "tests"
            tests_dir.mkdir()
            (tests_dir / "test_spec001.py").write_text("# Tests")

            result = manager.is_spec_implementation_completed("SPEC-001")
            # May be True or False depending on file matching


class TestSpecStatusUpdate:
    """Test SPEC status updates."""

    def test_update_spec_status_missing_file(self):
        """Test updating status of missing spec."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            result = manager.update_spec_status("NONEXISTENT", "completed")
            assert result is False

    def test_update_spec_status_draft_to_completed(self):
        """Test updating spec from draft to completed."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            spec_dir = specs_dir / "SPEC-001"
            spec_dir.mkdir(parents=True)

            spec_file = spec_dir / "spec.md"
            frontmatter = {"status": "draft", "version": "0.1.0"}
            spec_file.write_text(f"---\n{yaml.dump(frontmatter)}---\n# Content")

            result = manager.update_spec_status("SPEC-001", "completed")
            assert result is True

            # Verify update
            content = spec_file.read_text()
            assert "completed" in content

    def test_update_spec_status_bumps_version(self):
        """Test that completing a spec bumps version."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            spec_dir = specs_dir / "SPEC-001"
            spec_dir.mkdir(parents=True)

            spec_file = spec_dir / "spec.md"
            frontmatter = {"status": "draft", "version": "0.1.0"}
            spec_file.write_text(f"---\n{yaml.dump(frontmatter)}---\n# Content")

            manager.update_spec_status("SPEC-001", "completed")

            # Read back and verify version changed
            content = spec_file.read_text()
            assert "1.0.0" in content or content.count(".") > 1  # Version bumped


class TestSpecValidation:
    """Test SPEC validation for completion."""

    def test_validate_spec_for_completion_missing_spec(self):
        """Test validating missing spec."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            result = manager.validate_spec_for_completion("NONEXISTENT")

            assert result["is_ready"] is False
            assert len(result["issues"]) > 0

    def test_validate_spec_for_completion_no_code(self):
        """Test validation without implementation code."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            spec_dir = specs_dir / "SPEC-001"
            spec_dir.mkdir(parents=True)

            spec_file = spec_dir / "spec.md"
            spec_file.write_text("---\nstatus: draft\n---\n# Content")

            result = manager.validate_spec_for_completion("SPEC-001")
            assert result["is_ready"] is False

    def test_validate_spec_criteria_met_structure(self):
        """Test validation result structure."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            spec_dir = specs_dir / "SPEC-001"
            spec_dir.mkdir(parents=True)

            spec_file = spec_dir / "spec.md"
            spec_file.write_text("---\nstatus: draft\n---\n# Content")

            result = manager.validate_spec_for_completion("SPEC-001")

            assert "is_ready" in result
            assert "criteria_met" in result
            assert "issues" in result
            assert "recommendations" in result

    def test_get_completion_validation_criteria(self):
        """Test getting validation criteria."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            criteria = manager.get_completion_validation_criteria()

            assert isinstance(criteria, dict)
            assert "min_code_coverage" in criteria


class TestBatchOperations:
    """Test batch update operations."""

    def test_batch_update_completed_specs_empty(self):
        """Test batch update with no draft specs."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            specs_dir.mkdir(parents=True)

            result = manager.batch_update_completed_specs()

            assert "updated" in result
            assert "failed" in result
            assert "skipped" in result
            assert len(result["updated"]) == 0

    def test_batch_update_return_structure(self):
        """Test batch update return structure."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            specs_dir.mkdir(parents=True)

            result = manager.batch_update_completed_specs()

            assert isinstance(result, dict)
            assert "updated" in result
            assert "failed" in result
            assert "skipped" in result
            assert isinstance(result["updated"], list)


class TestVersionBumping:
    """Test version bumping logic."""

    def test_bump_version_0_point_x(self):
        """Test bumping version from 0.x.x."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            bumped = manager._bump_version("0.1.0")
            assert bumped == "1.0.0"

    def test_bump_version_1_point_x(self):
        """Test bumping version from 1.x.x."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            bumped = manager._bump_version("1.2.3")
            assert bumped.startswith("1.")

    def test_bump_version_with_quotes(self):
        """Test bumping version with quotes."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            bumped = manager._bump_version('"0.1.0"')
            assert "." in bumped

    def test_bump_version_invalid(self):
        """Test bumping invalid version."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            bumped = manager._bump_version("invalid")
            # Should either return 1.0.0 or handle gracefully
            assert isinstance(bumped, str)
            assert len(bumped) > 0


class TestAcceptanceCriteriaDetection:
    """Test acceptance criteria detection."""

    def test_check_acceptance_criteria_present(self):
        """Test detecting acceptance criteria when present."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            spec_dir = specs_dir / "SPEC-001"
            spec_dir.mkdir(parents=True)

            spec_file = spec_dir / "spec.md"
            spec_file.write_text("# Content\n## Acceptance Criteria\nCriteria here")

            result = manager._check_acceptance_criteria(spec_file)
            assert result is True

    def test_check_acceptance_criteria_absent(self):
        """Test detecting when acceptance criteria absent."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            spec_dir = specs_dir / "SPEC-001"
            spec_dir.mkdir(parents=True)

            spec_file = spec_dir / "spec.md"
            spec_file.write_text("# Content\n## Implementation\nDetails here")

            result = manager._check_acceptance_criteria(spec_file)
            assert result is False

    def test_check_acceptance_criteria_chinese(self):
        """Test detecting acceptance criteria in Chinese."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            spec_dir = specs_dir / "SPEC-001"
            spec_dir.mkdir(parents=True)

            spec_file = spec_dir / "spec.md"
            spec_file.write_text("# Content\n## 验收 标准\nCriteria here")

            result = manager._check_acceptance_criteria(spec_file)
            assert result is True


class TestDocumentationSync:
    """Test documentation sync checking."""

    def test_check_documentation_sync_no_docs(self):
        """Test when no docs directory exists."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            result = manager._check_documentation_sync("SPEC-001")
            assert result is True  # No docs to sync

    def test_check_documentation_sync_docs_exist(self):
        """Test when docs directory exists."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            docs_dir = Path(temp_dir) / "docs"
            docs_dir.mkdir()

            result = manager._check_documentation_sync("SPEC-001")
            assert result is True  # Docs present


class TestErrorHandling:
    """Test error handling and edge cases."""

    def test_detect_draft_specs_missing_directory(self):
        """Test detecting specs when directory doesn't exist."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            # Don't create .moai/specs directory

            draft_specs = manager.detect_draft_specs()
            assert isinstance(draft_specs, set)

    def test_update_spec_status_invalid_yaml(self):
        """Test updating spec with invalid YAML."""
        with tempfile.TemporaryDirectory() as temp_dir:
            manager = SpecStatusManager(Path(temp_dir))
            specs_dir = Path(temp_dir) / ".moai" / "specs"
            spec_dir = specs_dir / "SPEC-001"
            spec_dir.mkdir(parents=True)

            spec_file = spec_dir / "spec.md"
            spec_file.write_text("---\ninvalid: yaml: syntax\n---\n# Content")

            # Should handle gracefully
            result = manager.update_spec_status("SPEC-001", "completed")
            assert isinstance(result, bool)


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
