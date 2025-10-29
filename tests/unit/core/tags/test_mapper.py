# @TEST:MAP-001, @TEST:MAP-002
"""Test suite for SPEC-DOC mapping and confidence scoring.

Tests domain-based SPEC mapping, explicit metadata extraction,
and confidence scoring for TAG chain references.

@SPEC:DOC-TAG-001: @DOC tag automatic generation infrastructure
"""

import pytest
from pathlib import Path
from unittest.mock import patch, mock_open
from src.moai_adk.core.tags.mapper import find_related_spec, calculate_confidence


def test_find_spec_by_domain_match():
    """Find SPEC by matching domain from file path"""
    # Mock file system: docs/auth/setup.md should map to SPEC AUTH-001
    with patch("pathlib.Path.exists", return_value=True), \
         patch("pathlib.Path.glob") as mock_glob:

        # Mock SPEC directory containing AUTH-001
        mock_spec_dir = Path(".moai/specs/SPEC-AUTH-001")
        mock_glob.return_value = [mock_spec_dir]

        # Mock spec.md with AUTH-001 metadata
        spec_content = "---\nid: AUTH-001\n---\n# Authentication"
        with patch("builtins.open", mock_open(read_data=spec_content)):
            spec_id = find_related_spec(Path("docs/auth/setup.md"))
            assert spec_id == "AUTH-001"


def test_find_spec_no_match():
    """Return None when no SPEC matches domain"""
    with patch("pathlib.Path.glob", return_value=[]):
        spec_id = find_related_spec(Path("docs/random/file.md"))
        assert spec_id is None


def test_find_spec_multi_word_domain():
    """Handle multi-word domains like 'cli-tool'"""
    with patch("pathlib.Path.exists", return_value=True), \
         patch("pathlib.Path.glob") as mock_glob:

        mock_spec_dir = Path(".moai/specs/SPEC-CLI-TOOL-001")
        mock_glob.return_value = [mock_spec_dir]

        spec_content = "---\nid: CLI-TOOL-001\n---\n# CLI Tool"
        with patch("builtins.open", mock_open(read_data=spec_content)):
            spec_id = find_related_spec(Path("docs/cli-tool/commands.md"))
            assert spec_id == "CLI-TOOL-001"


def test_find_spec_case_insensitive():
    """Match domains case-insensitively"""
    with patch("pathlib.Path.exists", return_value=True), \
         patch("pathlib.Path.glob") as mock_glob:

        mock_spec_dir = Path(".moai/specs/SPEC-AUTH-001")
        mock_glob.return_value = [mock_spec_dir]

        spec_content = "---\nid: AUTH-001\n---\n# Auth"
        with patch("builtins.open", mock_open(read_data=spec_content)):
            # File path uses lowercase 'auth'
            spec_id = find_related_spec(Path("docs/auth/guide.md"))
            assert spec_id == "AUTH-001"


def test_calculate_confidence_high():
    """High confidence: domain + keyword match"""
    # File in 'auth' directory discussing authentication
    confidence = calculate_confidence(
        "AUTH-001",
        Path("docs/auth/authentication-guide.md")
    )
    assert confidence >= 0.8


def test_calculate_confidence_medium():
    """Medium confidence: domain match only"""
    confidence = calculate_confidence(
        "AUTH-001",
        Path("docs/auth/overview.md")
    )
    assert 0.4 <= confidence < 0.8


def test_calculate_confidence_low():
    """Low confidence: no domain match"""
    confidence = calculate_confidence(
        "AUTH-001",
        Path("docs/api/endpoints.md")
    )
    assert confidence < 0.4


def test_calculate_confidence_exact_spec_reference():
    """Very high confidence: file path contains SPEC ID"""
    confidence = calculate_confidence(
        "AUTH-001",
        Path("docs/auth/AUTH-001-implementation.md")
    )
    assert confidence >= 0.95


def test_find_spec_multiple_candidates():
    """When multiple SPECs match, return most recent"""
    with patch("pathlib.Path.exists", return_value=True), \
         patch("pathlib.Path.glob") as mock_glob:

        # Two AUTH specs: AUTH-001 and AUTH-002
        spec1 = Path(".moai/specs/SPEC-AUTH-001")
        spec2 = Path(".moai/specs/SPEC-AUTH-002")
        mock_glob.return_value = [spec1, spec2]

        # Mock reading both spec files
        def mock_open_func(file, *args, **kwargs):
            if "AUTH-001" in str(file):
                return mock_open(read_data="---\nid: AUTH-001\n---")()
            else:
                return mock_open(read_data="---\nid: AUTH-002\n---")()

        with patch("builtins.open", mock_open_func):
            # Should return higher number (most recent)
            spec_id = find_related_spec(Path("docs/auth/new-feature.md"))
            assert spec_id == "AUTH-002"


def test_find_spec_invalid_yaml():
    """Handle SPEC with invalid YAML gracefully"""
    with patch("pathlib.Path.exists", return_value=True), \
         patch("pathlib.Path.glob") as mock_glob:

        mock_spec_dir = Path(".moai/specs/SPEC-AUTH-001")
        mock_glob.return_value = [mock_spec_dir]

        # Invalid YAML
        spec_content = "---\ninvalid: yaml: syntax:\n---"
        with patch("builtins.open", mock_open(read_data=spec_content)):
            spec_id = find_related_spec(Path("docs/auth/guide.md"))
            # Should handle error and return None
            assert spec_id is None
