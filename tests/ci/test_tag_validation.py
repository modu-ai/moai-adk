#!/usr/bin/env python3
# @TEST:DOC-TAG-004 | Component 2: CI validator tests
"""Test suite for CI/CD TAG validation

This module tests the CIValidator that integrates with GitHub Actions:
- PR changed file detection via GitHub API
- Validation report generation
- Markdown comment formatting for PRs
- Strict vs. info mode behavior
- JSON serialization for GitHub Actions

Following TDD RED-GREEN-REFACTOR cycle.
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import Mock, patch

import pytest

from moai_adk.core.tags.pre_commit_validator import (
    PreCommitValidator,
    ValidationError,
    ValidationResult,
    ValidationWarning,
)

# Import will fail initially (RED phase) - that's expected
try:
    from moai_adk.core.tags.ci_validator import CIValidator
except ImportError:
    # Allow tests to be written before implementation
    CIValidator = None


@pytest.mark.skipif(CIValidator is None, reason="CIValidator not implemented yet")
class TestCIValidatorInitialization:
    """Test CIValidator initialization and inheritance"""

    def test_inherits_from_pre_commit_validator(self):
        """CIValidator should inherit from PreCommitValidator"""
        validator = CIValidator()
        assert isinstance(validator, PreCommitValidator)

    def test_default_initialization(self):
        """CIValidator should initialize with sensible defaults"""
        validator = CIValidator()
        assert validator.strict_mode is False
        assert validator.check_orphans is True

    def test_custom_github_token(self):
        """CIValidator should accept GitHub token for API calls"""
        validator = CIValidator(github_token="test_token_123")
        assert validator.github_token == "test_token_123"

    def test_custom_repo_info(self):
        """CIValidator should accept repository owner/name"""
        validator = CIValidator(repo_owner="testuser", repo_name="testrepo")
        assert validator.repo_owner == "testuser"
        assert validator.repo_name == "testrepo"


@pytest.mark.skipif(CIValidator is None, reason="CIValidator not implemented yet")
class TestGitHubAPIMocking:
    """Test GitHub API interaction with mocks"""

    @patch('requests.get')
    def test_get_pr_changed_files_success(self, mock_get):
        """Should fetch changed files from GitHub API"""
        # Mock successful API response
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = [
            {"filename": "src/auth.py", "status": "modified"},
            {"filename": "tests/test_auth.py", "status": "added"},
            {"filename": "README.md", "status": "modified"}
        ]
        mock_get.return_value = mock_response

        validator = CIValidator(
            github_token="test_token",
            repo_owner="testuser",
            repo_name="testrepo"
        )

        files = validator.get_pr_changed_files(pr_number=42)

        assert len(files) == 3
        assert "src/auth.py" in files
        assert "tests/test_auth.py" in files
        assert "README.md" in files

    @patch('requests.get')
    def test_get_pr_changed_files_api_error(self, mock_get):
        """Should handle API errors gracefully"""
        # Mock API error
        mock_response = Mock()
        mock_response.status_code = 404
        mock_response.raise_for_status.side_effect = Exception("Not Found")
        mock_get.return_value = mock_response

        validator = CIValidator(
            github_token="test_token",
            repo_owner="testuser",
            repo_name="testrepo"
        )

        files = validator.get_pr_changed_files(pr_number=9999)

        # Should return empty list on error
        assert files == []

    @patch('requests.get')
    def test_get_pr_changed_files_with_auth_header(self, mock_get):
        """Should include authorization header in API request"""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = []
        mock_get.return_value = mock_response

        validator = CIValidator(
            github_token="secret_token",
            repo_owner="testuser",
            repo_name="testrepo"
        )

        validator.get_pr_changed_files(pr_number=1)

        # Verify authorization header was included
        mock_get.assert_called_once()
        call_args = mock_get.call_args
        headers = call_args[1].get('headers', {})
        assert 'Authorization' in headers
        assert headers['Authorization'] == 'Bearer secret_token'


@pytest.mark.skipif(CIValidator is None, reason="CIValidator not implemented yet")
class TestPRValidation:
    """Test PR validation workflow"""

    @patch('requests.get')
    def test_validate_pr_changes_no_errors(self, mock_get):
        """PR with valid TAGs should pass validation"""
        # Mock API response
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = [
            {"filename": "test_file.py", "status": "modified"}
        ]
        mock_get.return_value = mock_response

        validator = CIValidator(
            github_token="token",
            repo_owner="user",
            repo_name="repo"
        )

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create test file with valid TAGs
            test_file = Path(tmpdir) / "test_file.py"
            test_file.write_text("""
# @CODE:AUTH-001
def authenticate():
    pass

# @TEST:AUTH-001
def test_authenticate():
    pass
""")

            # Mock file path resolution
            with patch.object(Path, 'exists', return_value=True), \
                 patch.object(Path, 'is_file', return_value=True), \
                 patch.object(Path, 'read_text', return_value=test_file.read_text()):

                result = validator.validate_pr_changes(pr_number=1)

                assert result.is_valid is True
                assert len(result.errors) == 0

    @patch('requests.get')
    def test_validate_pr_changes_with_errors(self, mock_get):
        """PR with TAG errors should fail validation"""
        # Mock API response
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = [
            {"filename": "test_file.py", "status": "modified"}
        ]
        mock_get.return_value = mock_response

        validator = CIValidator(
            github_token="token",
            repo_owner="user",
            repo_name="repo"
        )

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create test file with duplicate TAGs
            test_file = Path(tmpdir) / "test_file.py"
            test_file.write_text("""
# @CODE:AUTH-001
def func1():
    pass

# @CODE:AUTH-001
def func2():
    pass
""")

            # Mock file path resolution
            with patch.object(Path, 'exists', return_value=True), \
                 patch.object(Path, 'is_file', return_value=True), \
                 patch.object(Path, 'read_text', return_value=test_file.read_text()):

                result = validator.validate_pr_changes(pr_number=1)

                assert result.is_valid is False
                assert len(result.errors) > 0


@pytest.mark.skipif(CIValidator is None, reason="CIValidator not implemented yet")
class TestReportGeneration:
    """Test validation report generation"""

    def test_generate_report_success(self):
        """Should generate structured report for successful validation"""
        validator = CIValidator()

        result = ValidationResult(
            is_valid=True,
            errors=[],
            warnings=[]
        )

        report = validator.generate_report(result)

        assert isinstance(report, dict)
        assert report['status'] == 'success'
        assert report['is_valid'] is True
        assert 'summary' in report
        assert 'errors' in report
        assert 'warnings' in report
        assert 'statistics' in report

    def test_generate_report_with_errors(self):
        """Should generate report with error details"""
        validator = CIValidator()

        error = ValidationError(
            message="Duplicate TAG found",
            tag="@CODE:TEST-001",
            locations=[("file1.py", 10), ("file2.py", 20)]
        )

        result = ValidationResult(
            is_valid=False,
            errors=[error],
            warnings=[]
        )

        report = validator.generate_report(result)

        assert report['status'] == 'failure'
        assert report['is_valid'] is False
        assert len(report['errors']) == 1
        assert report['errors'][0]['message'] == "Duplicate TAG found"
        assert report['errors'][0]['tag'] == "@CODE:TEST-001"
        assert len(report['errors'][0]['locations']) == 2

    def test_generate_report_with_warnings(self):
        """Should generate report with warning details"""
        validator = CIValidator()

        warning = ValidationWarning(
            message="CODE TAG without corresponding TEST",
            tag="@CODE:AUTH-001",
            location=("auth.py", 15)
        )

        result = ValidationResult(
            is_valid=True,
            errors=[],
            warnings=[warning]
        )

        report = validator.generate_report(result)

        assert report['status'] == 'success_with_warnings'
        assert report['is_valid'] is True
        assert len(report['warnings']) == 1
        assert report['warnings'][0]['message'] == "CODE TAG without corresponding TEST"

    def test_generate_report_statistics(self):
        """Should include validation statistics"""
        validator = CIValidator()

        result = ValidationResult(
            is_valid=False,
            errors=[
                ValidationError("Error 1", "@CODE:TEST-001", [("f1.py", 1)]),
                ValidationError("Error 2", "@CODE:TEST-002", [("f2.py", 2)])
            ],
            warnings=[
                ValidationWarning("Warning 1", "@CODE:TEST-003", ("f3.py", 3))
            ]
        )

        report = validator.generate_report(result)

        stats = report['statistics']
        assert stats['total_errors'] == 2
        assert stats['total_warnings'] == 1
        assert stats['total_issues'] == 3

    def test_report_json_serialization(self):
        """Report should be JSON serializable"""
        validator = CIValidator()

        result = ValidationResult(
            is_valid=True,
            errors=[],
            warnings=[]
        )

        report = validator.generate_report(result)

        # Should not raise exception
        json_str = json.dumps(report)
        assert isinstance(json_str, str)

        # Should be deserializable
        parsed = json.loads(json_str)
        assert parsed['status'] == 'success'


@pytest.mark.skipif(CIValidator is None, reason="CIValidator not implemented yet")
class TestPRCommentFormatting:
    """Test markdown comment generation for GitHub PRs"""

    def test_format_pr_comment_success(self):
        """Should format success message for PR"""
        validator = CIValidator()

        result = ValidationResult(
            is_valid=True,
            errors=[],
            warnings=[]
        )

        comment = validator.format_pr_comment(
            result=result,
            pr_url="https://github.com/user/repo/pull/42"
        )

        assert isinstance(comment, str)
        assert "✅" in comment or "success" in comment.lower()
        assert "TAG validation passed" in comment or "no issues" in comment.lower()
        assert len(comment) > 0

    def test_format_pr_comment_with_errors(self):
        """Should format error details in PR comment"""
        validator = CIValidator()

        error = ValidationError(
            message="Duplicate TAG found",
            tag="@CODE:AUTH-001",
            locations=[("auth.py", 10), ("auth_v2.py", 20)]
        )

        result = ValidationResult(
            is_valid=False,
            errors=[error],
            warnings=[]
        )

        comment = validator.format_pr_comment(
            result=result,
            pr_url="https://github.com/user/repo/pull/42"
        )

        assert "❌" in comment or "error" in comment.lower()
        assert "Duplicate TAG" in comment
        assert "@CODE:AUTH-001" in comment
        assert "auth.py" in comment
        assert "auth_v2.py" in comment

    def test_format_pr_comment_with_warnings(self):
        """Should format warnings in PR comment"""
        validator = CIValidator()

        warning = ValidationWarning(
            message="CODE TAG without corresponding TEST",
            tag="@CODE:AUTH-001",
            location=("auth.py", 15)
        )

        result = ValidationResult(
            is_valid=True,
            errors=[],
            warnings=[warning]
        )

        comment = validator.format_pr_comment(
            result=result,
            pr_url="https://github.com/user/repo/pull/42"
        )

        assert "⚠️" in comment or "warning" in comment.lower()
        assert "CODE TAG without corresponding TEST" in comment
        assert "@CODE:AUTH-001" in comment

    def test_format_pr_comment_includes_table(self):
        """Should include results table in comment"""
        validator = CIValidator()

        result = ValidationResult(
            is_valid=False,
            errors=[
                ValidationError("Error 1", "@CODE:TEST-001", [("f1.py", 1)])
            ],
            warnings=[
                ValidationWarning("Warning 1", "@CODE:TEST-002", ("f2.py", 2))
            ]
        )

        comment = validator.format_pr_comment(
            result=result,
            pr_url="https://github.com/user/repo/pull/42"
        )

        # Should contain markdown table
        assert "|" in comment
        assert "---" in comment or "Errors" in comment

    def test_format_pr_comment_includes_docs_link(self):
        """Should include link to documentation for error resolution"""
        validator = CIValidator()

        result = ValidationResult(
            is_valid=False,
            errors=[
                ValidationError("Duplicate TAG", "@CODE:TEST-001", [("f1.py", 1)])
            ],
            warnings=[]
        )

        comment = validator.format_pr_comment(
            result=result,
            pr_url="https://github.com/user/repo/pull/42"
        )

        # Should contain documentation link or action items
        assert "docs" in comment.lower() or "how to fix" in comment.lower() or "resolution" in comment.lower()


@pytest.mark.skipif(CIValidator is None, reason="CIValidator not implemented yet")
class TestStrictVsInfoMode:
    """Test strict mode vs. info mode behavior"""

    def test_info_mode_allows_warnings(self):
        """Info mode should pass validation with warnings"""
        validator = CIValidator(strict_mode=False)

        with tempfile.TemporaryDirectory() as tmpdir:
            # File with orphan CODE (warning only)
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            result = validator.validate_files([str(file1)])

            # Should pass in info mode
            assert result.is_valid is True
            assert len(result.warnings) > 0

    def test_strict_mode_blocks_warnings(self):
        """Strict mode should fail validation on warnings"""
        validator = CIValidator(strict_mode=True)

        with tempfile.TemporaryDirectory() as tmpdir:
            # File with orphan CODE (warning in info mode)
            file1 = Path(tmpdir) / "file1.py"
            file1.write_text("# @CODE:TEST-001\n")

            result = validator.validate_files([str(file1)])

            # Should fail in strict mode
            assert result.is_valid is False
            assert len(result.warnings) > 0

    def test_strict_mode_in_report(self):
        """Report should indicate strict mode status"""
        validator_strict = CIValidator(strict_mode=True)
        validator_info = CIValidator(strict_mode=False)

        result = ValidationResult(is_valid=True, errors=[], warnings=[])

        report_strict = validator_strict.generate_report(result)
        report_info = validator_info.generate_report(result)

        assert 'strict_mode' in report_strict
        assert report_strict['strict_mode'] is True
        assert report_info['strict_mode'] is False


@pytest.mark.skipif(CIValidator is None, reason="CIValidator not implemented yet")
class TestEnvironmentVariableHandling:
    """Test handling of environment variables for GitHub Actions"""

    def test_github_token_from_env(self):
        """Should read GitHub token from environment variable"""
        with patch.dict('os.environ', {'GITHUB_TOKEN': 'env_token_123'}):
            validator = CIValidator()
            assert validator.github_token == 'env_token_123'

    def test_repo_info_from_env(self):
        """Should read repo info from GitHub Actions environment"""
        with patch.dict('os.environ', {
            'GITHUB_REPOSITORY': 'testuser/testrepo'
        }):
            validator = CIValidator()
            assert validator.repo_owner == 'testuser'
            assert validator.repo_name == 'testrepo'

    def test_pr_number_from_env(self):
        """Should extract PR number from GitHub event"""
        with patch.dict('os.environ', {
            'GITHUB_EVENT_PATH': '/tmp/event.json'
        }):
            # Mock event file
            event_data = {'pull_request': {'number': 42}}
            with patch('builtins.open', create=True) as mock_open:
                mock_open.return_value.__enter__.return_value.read.return_value = json.dumps(event_data)

                validator = CIValidator()
                pr_number = validator.get_pr_number_from_event()

                assert pr_number == 42


@pytest.mark.skipif(CIValidator is None, reason="CIValidator not implemented yet")
class TestIntegrationWorkflow:
    """Test complete CI validation workflow"""

    @patch('requests.get')
    def test_complete_validation_workflow(self, mock_get):
        """Test complete workflow: fetch files -> validate -> generate report -> format comment"""
        # Mock GitHub API
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = [
            {"filename": "src/auth.py", "status": "modified"}
        ]
        mock_get.return_value = mock_response

        validator = CIValidator(
            github_token="token",
            repo_owner="user",
            repo_name="repo"
        )

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create test file
            test_file = Path(tmpdir) / "src" / "auth.py"
            test_file.parent.mkdir(parents=True)
            test_file.write_text("# @CODE:AUTH-001\n# @TEST:AUTH-001\n")

            # Mock file resolution
            with patch.object(Path, 'exists', return_value=True), \
                 patch.object(Path, 'is_file', return_value=True), \
                 patch.object(Path, 'read_text', return_value=test_file.read_text()):

                # 1. Validate PR changes
                result = validator.validate_pr_changes(pr_number=1)

                # 2. Generate report
                report = validator.generate_report(result)

                # 3. Format PR comment
                comment = validator.format_pr_comment(result, "https://github.com/user/repo/pull/1")

                # Verify workflow
                assert result.is_valid is True
                assert report['status'] == 'success'
                assert isinstance(comment, str)
                assert len(comment) > 0


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
