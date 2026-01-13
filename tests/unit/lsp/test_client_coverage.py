"""Additional coverage tests for LSP client module.

Tests for lines not covered by existing tests.
"""


import pytest

from moai_adk.lsp.client import MoAILSPClient
from moai_adk.lsp.models import DiagnosticSeverity, HoverInfo, Position


class TestMoAILSPClientConfig:
    """Test LSP configuration loading."""

    def test_load_config_when_file_exists(self, tmp_path):
        """Should load config from .lsp.json file."""
        config_file = tmp_path / ".lsp.json"
        config_file.write_text('{"servers": {}}')

        client = MoAILSPClient(tmp_path)

        # Config should be loaded
        assert client.server_manager is not None


class TestMoAILSPClientInternalRequests:
    """Test internal request methods with default implementations."""

    @pytest.mark.asyncio
    async def test_request_diagnostics_returns_empty_list(self, tmp_path):
        """Should return empty list by default."""
        client = MoAILSPClient(tmp_path)

        result = await client._request_diagnostics("test.py")

        assert result == []

    @pytest.mark.asyncio
    async def test_request_references_returns_empty_list(self, tmp_path):
        """Should return empty list by default."""
        client = MoAILSPClient(tmp_path)
        position = Position(line=0, character=0)

        result = await client._request_references("test.py", position)

        assert result == []

    @pytest.mark.asyncio
    async def test_request_rename_returns_empty_changes(self, tmp_path):
        """Should return empty changes dict by default."""
        client = MoAILSPClient(tmp_path)
        position = Position(line=0, character=0)

        result = await client._request_rename("test.py", position, "new_name")

        assert result == {"changes": {}}

    @pytest.mark.asyncio
    async def test_request_hover_returns_none(self, tmp_path):
        """Should return None by default."""
        client = MoAILSPClient(tmp_path)
        position = Position(line=0, character=0)

        result = await client._request_hover("test.py", position)

        assert result is None


class TestMoAILSPClientURIConversion:
    """Test URI conversion methods."""

    def test_file_to_uri_adds_slash_for_relative_path(self, tmp_path):
        """Should add leading slash to relative paths."""
        client = MoAILSPClient(tmp_path)

        result = client._file_to_uri("relative/path.py")

        assert result == "file:///relative/path.py"

    def test_file_to_uri_preserves_absolute_path(self, tmp_path):
        """Should preserve leading slash for absolute paths."""
        client = MoAILSPClient(tmp_path)

        result = client._file_to_uri("/absolute/path.py")

        assert result == "file:///absolute/path.py"

    def test_uri_to_file_without_file_prefix(self, tmp_path):
        """Should return URI as-is when it doesn't start with file://"""
        client = MoAILSPClient(tmp_path)

        result = client._uri_to_file("http://example.com/file.py")

        assert result == "http://example.com/file.py"


class TestMoAILSPClientParseHoverInfo:
    """Test hover info parsing."""

    def test_parse_hover_info_with_markup_content(self, tmp_path):
        """Should parse MarkupContent format correctly."""
        client = MoAILSPClient(tmp_path)

        raw = {
            "contents": {"kind": "markdown", "value": "Hover text"},
            "range": {
                "start": {"line": 0, "character": 0},
                "end": {"line": 0, "character": 5},
            },
        }

        result = client._parse_hover_info(raw)

        assert isinstance(result, HoverInfo)
        assert result.contents == "Hover text"
        assert result.range is not None


class TestMoAILSPClientParseDiagnostic:
    """Test diagnostic parsing with various severities."""

    def test_parse_diagnostic_with_error_severity(self, tmp_path):
        """Should parse diagnostic with error severity."""
        client = MoAILSPClient(tmp_path)

        raw = {
            "range": {
                "start": {"line": 0, "character": 0},
                "end": {"line": 0, "character": 5},
            },
            "severity": 1,  # Error
            "message": "Test error",
        }

        result = client._parse_diagnostic(raw)

        assert result.severity == DiagnosticSeverity.ERROR
        assert result.message == "Test error"
