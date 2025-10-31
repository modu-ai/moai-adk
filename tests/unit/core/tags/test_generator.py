# @TEST:GEN-001, @TEST:GEN-002, @TEST:GEN-003
"""Test suite for TAG ID generation and duplicate detection.

Tests sequential TAG ID generation, duplicate detection using ripgrep,
and domain validation for @DOC tags.

@SPEC:DOC-TAG-001: @DOC tag automatic generation infrastructure
"""

from unittest.mock import MagicMock, patch

import pytest

from src.moai_adk.core.tags.generator import detect_duplicates, generate_doc_tag


def test_generate_first_doc_tag():
    """Generate first @DOC tag for domain"""
    # No existing tags
    tag = generate_doc_tag("AUTH", existing_ids=[])
    assert tag == "@DOC:AUTH-001"


def test_generate_next_doc_tag():
    """Generate next sequential TAG ID"""
    existing = ["@DOC:AUTH-001", "@DOC:AUTH-002"]
    tag = generate_doc_tag("AUTH", existing_ids=existing)
    assert tag == "@DOC:AUTH-003"


def test_generate_skips_gaps():
    """Generate next TAG based on highest number, ignoring gaps"""
    existing = ["@DOC:AUTH-001", "@DOC:AUTH-003", "@DOC:AUTH-005"]
    tag = generate_doc_tag("AUTH", existing_ids=existing)
    assert tag == "@DOC:AUTH-006"


def test_generate_different_domains():
    """Different domains have independent numbering"""
    auth_tag = generate_doc_tag("AUTH", [])
    api_tag = generate_doc_tag("API", [])
    assert auth_tag == "@DOC:AUTH-001"
    assert api_tag == "@DOC:API-001"


def test_generate_multi_word_domain():
    """Handle multi-word domains with hyphens"""
    tag = generate_doc_tag("CLI-TOOL", [])
    assert tag == "@DOC:CLI-TOOL-001"


def test_generate_filters_other_domains():
    """Only count tags from same domain"""
    existing = [
        "@DOC:AUTH-001",
        "@DOC:API-001",
        "@DOC:AUTH-002",
        "@DOC:DB-001"
    ]
    tag = generate_doc_tag("AUTH", existing_ids=existing)
    assert tag == "@DOC:AUTH-003"


def test_invalid_domain_format():
    """Validate domain format (uppercase alphanumeric + hyphens)"""
    with pytest.raises(ValueError, match="Invalid domain"):
        generate_doc_tag("invalid-lowercase", [])

    with pytest.raises(ValueError, match="Invalid domain"):
        generate_doc_tag("123", [])  # Must start with letter

    with pytest.raises(ValueError, match="Invalid domain"):
        generate_doc_tag("A_B", [])  # No underscores


def test_detect_duplicates_with_ripgrep():
    """Detect existing TAGs using ripgrep"""
    # Mock ripgrep output
    with patch("subprocess.run") as mock_run:
        mock_result = MagicMock()
        mock_result.returncode = 0
        mock_result.stdout = "@DOC:AUTH-001\n@DOC:AUTH-002\n@DOC:AUTH-003"
        mock_run.return_value = mock_result

        duplicates = detect_duplicates("AUTH", "docs/")
        assert len(duplicates) == 3
        assert "@DOC:AUTH-001" in duplicates
        assert "@DOC:AUTH-002" in duplicates
        assert "@DOC:AUTH-003" in duplicates


def test_detect_duplicates_no_matches():
    """Handle no matches from ripgrep (new domain)"""
    with patch("subprocess.run") as mock_run:
        mock_result = MagicMock()
        mock_result.returncode = 1  # ripgrep exit code 1 = no matches
        mock_result.stdout = ""
        mock_run.return_value = mock_result

        duplicates = detect_duplicates("NEWDOMAIN", "docs/")
        assert duplicates == []


def test_detect_duplicates_ripgrep_error():
    """Handle ripgrep execution errors"""
    with patch("subprocess.run") as mock_run:
        mock_result = MagicMock()
        mock_result.returncode = 2
        mock_result.stderr = "ripgrep error"
        mock_run.return_value = mock_result

        with pytest.raises(RuntimeError, match="ripgrep error"):
            detect_duplicates("AUTH", "docs/")


def test_detect_duplicates_ripgrep_not_found():
    """Handle ripgrep not installed"""
    with patch("subprocess.run", side_effect=FileNotFoundError):
        with pytest.raises(RuntimeError, match="ripgrep .* not found"):
            detect_duplicates("AUTH", "docs/")


def test_detect_duplicates_timeout():
    """Handle ripgrep timeout"""
    import subprocess
    with patch("subprocess.run", side_effect=subprocess.TimeoutExpired("rg", 5)):
        with pytest.raises(RuntimeError, match="ripgrep timeout"):
            detect_duplicates("AUTH", "docs/")
