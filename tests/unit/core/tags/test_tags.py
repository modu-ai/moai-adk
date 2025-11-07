# @TEST:VAL-002
"""Test suite for TAG suggestion orchestrator.

Tests integrated TAG suggestion, domain inference, and chain validation.

@SPEC:DOC-TAG-001: @DOC tag automatic generation infrastructure
"""

from pathlib import Path
from unittest.mock import patch

from src.moai_adk.core.tags.tags import (
    TagSuggestion,
    _infer_domain_from_path,
    suggest_tag_for_file,
    validate_tag_chain,
)


def test_suggest_tag_with_spec():
    """Suggest TAG when SPEC is found"""
    with patch("src.moai_adk.core.tags.tags.find_related_spec", return_value="AUTH-001"), \
         patch("src.moai_adk.core.tags.tags.detect_duplicates", return_value=[]), \
         patch("src.moai_adk.core.tags.tags.calculate_confidence", return_value=0.85):

        suggestion = suggest_tag_for_file(Path("docs/auth/guide.md"))

        assert suggestion.tag_id == "@DOC:AUTH-001"
        assert suggestion.chain_ref == "@SPEC:AUTH-001"
        assert suggestion.confidence == 0.85
        assert suggestion.domain == "AUTH"


def test_suggest_tag_without_spec():
    """Suggest TAG when no SPEC found (infer domain)"""
    with patch("src.moai_adk.core.tags.tags.find_related_spec", return_value=None), \
         patch("src.moai_adk.core.tags.tags.detect_duplicates", return_value=[]):

        suggestion = suggest_tag_for_file(Path("docs/tutorial/setup.md"))

        assert suggestion.tag_id == "@DOC:TUTORIAL-001"
        assert suggestion.chain_ref is None
        assert suggestion.confidence == 0.3
        assert suggestion.domain == "TUTORIAL"


def test_suggest_tag_increments_number():
    """Generate next sequential TAG number"""
    existing = ["@DOC:AUTH-001", "@DOC:AUTH-002"]

    with patch("src.moai_adk.core.tags.tags.find_related_spec", return_value="AUTH-001"), \
         patch("src.moai_adk.core.tags.tags.detect_duplicates", return_value=existing), \
         patch("src.moai_adk.core.tags.tags.calculate_confidence", return_value=0.9):

        suggestion = suggest_tag_for_file(Path("docs/auth/advanced.md"))

        assert suggestion.tag_id == "@DOC:AUTH-003"


def test_infer_domain_from_path():
    """Infer domain from file path structure"""
    assert _infer_domain_from_path(Path("docs/auth/guide.md")) == "AUTH"
    assert _infer_domain_from_path(Path("docs/api/reference.md")) == "API"
    assert _infer_domain_from_path(Path("docs/cli-tool/commands.md")) == "CLI-TOOL"


def test_infer_domain_short_path():
    """Handle short paths without subdirectory"""
    assert _infer_domain_from_path(Path("README.md")) == "DOC"


def test_validate_tag_chain_valid():
    """Validate correct TAG chain"""
    assert validate_tag_chain(
        "@DOC:AUTH-001",
        "@SPEC:AUTH-001 -> @DOC:AUTH-001"
    ) is True


def test_validate_tag_chain_domain_mismatch():
    """Detect domain mismatch in chain"""
    assert validate_tag_chain(
        "@DOC:AUTH-001",
        "@SPEC:API-001 -> @DOC:AUTH-001"
    ) is False


def test_validate_tag_chain_missing_tag():
    """Detect missing TAG in chain"""
    assert validate_tag_chain(
        "@DOC:AUTH-002",
        "@SPEC:AUTH-001 -> @DOC:AUTH-001"
    ) is False


def test_tag_suggestion_dataclass():
    """Verify TagSuggestion dataclass structure"""
    suggestion = TagSuggestion(
        tag_id="@DOC:AUTH-001",
        chain_ref="@SPEC:AUTH-001",
        confidence=0.85,
        domain="AUTH",
        file_path=Path("docs/auth/guide.md"),
    )

    assert suggestion.tag_id == "@DOC:AUTH-001"
    assert suggestion.chain_ref == "@SPEC:AUTH-001"
    assert suggestion.confidence == 0.85
    assert suggestion.domain == "AUTH"
    assert suggestion.file_path == Path("docs/auth/guide.md")
