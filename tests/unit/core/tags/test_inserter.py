# @TEST:INS-001, @TEST:INS-002, @TEST:INS-003
"""Test suite for markdown TAG insertion.

Tests TAG insertion into markdown headers, chain reference formatting,
and error handling for file operations.

@SPEC:DOC-TAG-001: @DOC tag automatic generation infrastructure
"""

import pytest
from pathlib import Path
from unittest.mock import patch, mock_open, MagicMock
from src.moai_adk.core.tags.inserter import insert_tag_to_markdown, format_tag_header


def test_insert_tag_into_header():
    """Insert TAG comment into markdown header"""
    original = """# User Guide

This is a guide.
"""
    expected = """# @DOC:AUTH-001 | Chain: @SPEC:AUTH-001 -> @DOC:AUTH-001
# User Guide

This is a guide.
"""

    with patch("pathlib.Path.read_text", return_value=original), \
         patch("pathlib.Path.write_text") as mock_write:

        result = insert_tag_to_markdown(
            Path("guide.md"),
            "@DOC:AUTH-001",
            "@SPEC:AUTH-001"
        )

        assert result is True
        mock_write.assert_called_once()
        written_content = mock_write.call_args[0][0]
        assert "@DOC:AUTH-001" in written_content
        assert "Chain: @SPEC:AUTH-001" in written_content


def test_format_tag_header():
    """Format TAG header with chain reference"""
    header = format_tag_header("@DOC:AUTH-001", "@SPEC:AUTH-001")
    assert header == "# @DOC:AUTH-001 | Chain: @SPEC:AUTH-001 -> @DOC:AUTH-001"


def test_insert_tag_file_not_found():
    """Handle file not found error"""
    with patch("pathlib.Path.read_text", side_effect=FileNotFoundError):
        result = insert_tag_to_markdown(
            Path("missing.md"),
            "@DOC:AUTH-001",
            "@SPEC:AUTH-001"
        )
        assert result is False


def test_insert_tag_permission_error():
    """Handle permission error"""
    with patch("pathlib.Path.read_text", return_value="# Content"), \
         patch("pathlib.Path.write_text", side_effect=PermissionError):

        result = insert_tag_to_markdown(
            Path("readonly.md"),
            "@DOC:AUTH-001",
            "@SPEC:AUTH-001"
        )
        assert result is False
