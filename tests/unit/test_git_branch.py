"""Unit tests for git/branch.py module

Tests for branch naming utilities.
"""

from moai_adk.core.git.branch import generate_branch_name


class TestGenerateBranchName:
    """Test generate_branch_name function"""

    def test_generate_branch_name_simple_spec_id(self):
        """Should generate correct branch name from simple SPEC ID"""
        result = generate_branch_name("AUTH-001")
        assert result == "feature/SPEC-AUTH-001"

    def test_generate_branch_name_complex_spec_id(self):
        """Should generate correct branch name from complex SPEC ID"""
        result = generate_branch_name("CORE-GIT-001")
        assert result == "feature/SPEC-CORE-GIT-001"

    def test_generate_branch_name_multiple_hyphens(self):
        """Should handle SPEC ID with multiple hyphens"""
        result = generate_branch_name("UPDATE-REFACTOR-001")
        assert result == "feature/SPEC-UPDATE-REFACTOR-001"

    def test_generate_branch_name_prefix_format(self):
        """Branch name should always start with feature/SPEC-"""
        result = generate_branch_name("TEST-999")
        assert result.startswith("feature/SPEC-")

    def test_generate_branch_name_preserves_spec_id(self):
        """Should preserve original SPEC ID in branch name"""
        spec_id = "INSTALLER-SEC-042"
        result = generate_branch_name(spec_id)
        assert spec_id in result
        assert result == f"feature/SPEC-{spec_id}"
