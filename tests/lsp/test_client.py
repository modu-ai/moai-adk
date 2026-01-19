# LSP Client Tests - RED Phase
"""Tests for MoAI LSP client interface."""

from pathlib import Path
from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from moai_adk.lsp.client import LanguageSession, MoAILSPClient
from moai_adk.lsp.models import (
    Diagnostic,
    DiagnosticSeverity,
    DocumentSymbol,
    HoverInfo,
    Location,
    Position,
    SymbolKind,
    WorkspaceEdit,
)


class TestMoAILSPClientCreation:
    """Tests for LSP client initialization."""

    def test_create_client(self):
        """Client should initialize with project root."""
        client = MoAILSPClient(project_root=Path("/project"))
        assert client.project_root == Path("/project")

    def test_create_client_with_string_path(self):
        """Client should accept string path."""
        client = MoAILSPClient(project_root="/project")
        assert client.project_root == Path("/project")

    def test_client_loads_lsp_config(self):
        """Client should load .lsp.json on init."""
        with patch.object(MoAILSPClient, "_load_config") as mock_load:
            MoAILSPClient(project_root="/project")
            mock_load.assert_called_once()


@pytest.mark.asyncio
class TestMoAILSPClientDiagnostics:
    """Tests for diagnostic functionality."""

    async def test_get_diagnostics_for_file(self):
        """Client should return diagnostics for a file."""
        client = MoAILSPClient(project_root="/project")

        # Mock server response
        mock_diagnostics = [
            {
                "range": {"start": {"line": 5, "character": 0}, "end": {"line": 5, "character": 10}},
                "severity": 1,
                "code": "E001",
                "source": "pyright",
                "message": "Undefined variable 'x'",
            }
        ]

        with patch.object(client, "_request_diagnostics", return_value=mock_diagnostics):
            diagnostics = await client.get_diagnostics("/project/src/main.py")

        assert len(diagnostics) == 1
        assert isinstance(diagnostics[0], Diagnostic)
        assert diagnostics[0].severity == DiagnosticSeverity.ERROR
        assert diagnostics[0].message == "Undefined variable 'x'"

    async def test_get_diagnostics_empty(self):
        """Client should return empty list when no diagnostics."""
        client = MoAILSPClient(project_root="/project")

        with patch.object(client, "_request_diagnostics", return_value=[]):
            diagnostics = await client.get_diagnostics("/project/src/clean.py")

        assert diagnostics == []

    async def test_get_diagnostics_multiple(self):
        """Client should return multiple diagnostics."""
        client = MoAILSPClient(project_root="/project")

        mock_diagnostics = [
            {
                "range": {"start": {"line": 1, "character": 0}, "end": {"line": 1, "character": 5}},
                "severity": 1,
                "code": "E001",
                "source": "pyright",
                "message": "Error 1",
            },
            {
                "range": {"start": {"line": 2, "character": 0}, "end": {"line": 2, "character": 5}},
                "severity": 2,
                "code": "W001",
                "source": "pyright",
                "message": "Warning 1",
            },
        ]

        with patch.object(client, "_request_diagnostics", return_value=mock_diagnostics):
            diagnostics = await client.get_diagnostics("/project/src/file.py")

        assert len(diagnostics) == 2
        assert diagnostics[0].is_error()
        assert not diagnostics[1].is_error()


@pytest.mark.asyncio
class TestMoAILSPClientReferences:
    """Tests for find references functionality."""

    async def test_find_references(self):
        """Client should find symbol references."""
        client = MoAILSPClient(project_root="/project")

        mock_refs = [
            {
                "uri": "file:///project/src/main.py",
                "range": {"start": {"line": 10, "character": 5}, "end": {"line": 10, "character": 15}},
            },
            {
                "uri": "file:///project/src/util.py",
                "range": {"start": {"line": 20, "character": 0}, "end": {"line": 20, "character": 10}},
            },
        ]

        with patch.object(client, "_request_references", return_value=mock_refs):
            refs = await client.find_references("/project/src/main.py", Position(line=5, character=10))

        assert len(refs) == 2
        assert all(isinstance(r, Location) for r in refs)
        assert refs[0].uri == "file:///project/src/main.py"

    async def test_find_references_none(self):
        """Client should return empty list when no references found."""
        client = MoAILSPClient(project_root="/project")

        with patch.object(client, "_request_references", return_value=[]):
            refs = await client.find_references("/project/src/main.py", Position(line=1, character=0))

        assert refs == []


@pytest.mark.asyncio
class TestMoAILSPClientRename:
    """Tests for rename symbol functionality."""

    async def test_rename_symbol(self):
        """Client should return workspace edit for rename."""
        client = MoAILSPClient(project_root="/project")

        mock_edit = {
            "changes": {
                "file:///project/src/main.py": [
                    {
                        "range": {"start": {"line": 5, "character": 0}, "end": {"line": 5, "character": 7}},
                        "newText": "new_name",
                    },
                    {
                        "range": {"start": {"line": 10, "character": 4}, "end": {"line": 10, "character": 11}},
                        "newText": "new_name",
                    },
                ]
            }
        }

        with patch.object(client, "_request_rename", return_value=mock_edit):
            edit = await client.rename_symbol("/project/src/main.py", Position(line=5, character=3), "new_name")

        assert isinstance(edit, WorkspaceEdit)
        assert edit.file_count() == 1
        assert len(edit.changes["file:///project/src/main.py"]) == 2

    async def test_rename_symbol_multiple_files(self):
        """Client should handle renames across multiple files."""
        client = MoAILSPClient(project_root="/project")

        mock_edit = {
            "changes": {
                "file:///project/src/main.py": [
                    {
                        "range": {"start": {"line": 5, "character": 0}, "end": {"line": 5, "character": 7}},
                        "newText": "renamed",
                    }
                ],
                "file:///project/src/util.py": [
                    {
                        "range": {"start": {"line": 15, "character": 0}, "end": {"line": 15, "character": 7}},
                        "newText": "renamed",
                    }
                ],
            }
        }

        with patch.object(client, "_request_rename", return_value=mock_edit):
            edit = await client.rename_symbol("/project/src/main.py", Position(line=5, character=0), "renamed")

        assert edit.file_count() == 2


@pytest.mark.asyncio
class TestMoAILSPClientHover:
    """Tests for hover info functionality."""

    async def test_get_hover_info(self):
        """Client should return hover information."""
        client = MoAILSPClient(project_root="/project")

        mock_hover = {
            "contents": "def foo(x: int) -> str:\n    '''Docstring'''",
            "range": {"start": {"line": 5, "character": 0}, "end": {"line": 5, "character": 3}},
        }

        with patch.object(client, "_request_hover", return_value=mock_hover):
            hover = await client.get_hover_info("/project/src/main.py", Position(line=5, character=1))

        assert isinstance(hover, HoverInfo)
        assert "def foo" in hover.contents
        assert hover.range is not None

    async def test_get_hover_info_no_range(self):
        """Client should handle hover without range."""
        client = MoAILSPClient(project_root="/project")

        mock_hover = {"contents": "Some documentation"}

        with patch.object(client, "_request_hover", return_value=mock_hover):
            hover = await client.get_hover_info("/project/src/main.py", Position(line=10, character=5))

        if hover is None:
            raise AssertionError("hover should not be None")
        assert hover.contents == "Some documentation"
        assert hover.range is None

    async def test_get_hover_info_none(self):
        """Client should return None when no hover info available."""
        client = MoAILSPClient(project_root="/project")

        with patch.object(client, "_request_hover", return_value=None):
            hover = await client.get_hover_info("/project/src/main.py", Position(line=1, character=0))

        assert hover is None


@pytest.mark.asyncio
class TestMoAILSPClientServerManagement:
    """Tests for server lifecycle management."""

    async def test_start_server_for_language(self):
        """Client should start server for a language."""
        client = MoAILSPClient(project_root="/project")

        # Mock _get_or_create_session to avoid complex initialization
        mock_session = MagicMock(spec=LanguageSession)
        mock_session.language = "python"
        mock_session.initialized = True

        with patch.object(client, "_get_or_create_session", new_callable=AsyncMock) as mock_get_session:
            mock_get_session.return_value = mock_session
            await client.ensure_server_running("python")
            mock_get_session.assert_called_once_with("python")

    async def test_stop_all_servers(self):
        """Client should stop all servers on cleanup."""
        client = MoAILSPClient(project_root="/project")

        with patch.object(client.server_manager, "stop_all_servers", new_callable=AsyncMock) as mock_stop:
            await client.cleanup()
            mock_stop.assert_called_once()

    async def test_get_language_for_file(self):
        """Client should determine language from file extension."""
        client = MoAILSPClient(project_root="/project")

        with patch.object(client.server_manager, "get_language_for_file") as mock_get:
            mock_get.return_value = "python"
            lang = client.get_language_for_file("/project/src/main.py")
            assert lang == "python"


class TestMoAILSPClientHelpers:
    """Tests for helper methods."""

    def test_file_to_uri(self):
        """Client should convert file path to URI."""
        client = MoAILSPClient(project_root="/project")
        uri = client._file_to_uri("/project/src/main.py")
        assert uri == "file:///project/src/main.py"

    def test_uri_to_file(self):
        """Client should convert URI to file path."""
        client = MoAILSPClient(project_root="/project")
        path = client._uri_to_file("file:///project/src/main.py")
        assert path == "/project/src/main.py"

    def test_parse_diagnostic(self):
        """Client should parse diagnostic from LSP response."""
        client = MoAILSPClient(project_root="/project")
        raw = {
            "range": {"start": {"line": 5, "character": 0}, "end": {"line": 5, "character": 10}},
            "severity": 2,
            "code": "W001",
            "source": "mypy",
            "message": "Type warning",
        }
        diag = client._parse_diagnostic(raw)
        assert diag.severity == DiagnosticSeverity.WARNING
        assert diag.source == "mypy"

    def test_parse_location(self):
        """Client should parse location from LSP response."""
        client = MoAILSPClient(project_root="/project")
        raw = {
            "uri": "file:///project/test.py",
            "range": {"start": {"line": 10, "character": 5}, "end": {"line": 10, "character": 15}},
        }
        loc = client._parse_location(raw)
        assert loc.uri == "file:///project/test.py"
        assert loc.range.start.line == 10


@pytest.mark.asyncio
class TestMoAILSPClientGoToDefinition:
    """Tests for go to definition functionality."""

    async def test_go_to_definition(self):
        """Client should return definition locations."""
        client = MoAILSPClient(project_root="/project")

        mock_definition = {
            "uri": "file:///project/src/module.py",
            "range": {"start": {"line": 15, "character": 0}, "end": {"line": 15, "character": 20}},
        }

        with patch.object(client, "_request_definition", new_callable=AsyncMock) as mock_req:
            mock_req.return_value = mock_definition
            locations = await client.go_to_definition("/project/src/main.py", Position(line=10, character=5))

        assert len(locations) == 1
        assert isinstance(locations[0], Location)
        assert locations[0].uri == "file:///project/src/module.py"
        assert locations[0].range.start.line == 15

    async def test_go_to_definition_list_response(self):
        """Client should handle list response from definition."""
        client = MoAILSPClient(project_root="/project")

        mock_definitions = [
            {
                "uri": "file:///project/src/module.py",
                "range": {"start": {"line": 15, "character": 0}, "end": {"line": 15, "character": 20}},
            },
            {
                "uri": "file:///project/src/other.py",
                "range": {"start": {"line": 5, "character": 0}, "end": {"line": 5, "character": 10}},
            },
        ]

        with patch.object(client, "_request_definition", new_callable=AsyncMock) as mock_req:
            mock_req.return_value = mock_definitions
            locations = await client.go_to_definition("/project/src/main.py", Position(line=10, character=5))

        assert len(locations) == 2
        assert all(isinstance(loc, Location) for loc in locations)
        assert locations[0].uri == "file:///project/src/module.py"
        assert locations[1].uri == "file:///project/src/other.py"

    async def test_go_to_definition_empty(self):
        """Client should return empty list when no definition found."""
        client = MoAILSPClient(project_root="/project")

        with patch.object(client, "_request_definition", new_callable=AsyncMock) as mock_req:
            mock_req.return_value = []
            locations = await client.go_to_definition("/project/src/main.py", Position(line=1, character=0))

        assert locations == []


@pytest.mark.asyncio
class TestMoAILSPClientDocumentSymbols:
    """Tests for document symbols functionality."""

    async def test_get_document_symbols(self):
        """Client should return document symbols."""
        client = MoAILSPClient(project_root="/project")

        mock_symbols = [
            {
                "name": "MyClass",
                "kind": 5,  # Class
                "range": {"start": {"line": 5, "character": 0}, "end": {"line": 20, "character": 0}},
                "selectionRange": {"start": {"line": 5, "character": 6}, "end": {"line": 5, "character": 13}},
                "detail": "class MyClass",
                "children": [
                    {
                        "name": "my_method",
                        "kind": 6,  # Method
                        "range": {"start": {"line": 10, "character": 4}, "end": {"line": 15, "character": 4}},
                        "selectionRange": {"start": {"line": 10, "character": 8}, "end": {"line": 10, "character": 17}},
                    }
                ],
            },
            {
                "name": "helper_function",
                "kind": 12,  # Function
                "range": {"start": {"line": 25, "character": 0}, "end": {"line": 30, "character": 0}},
                "selectionRange": {"start": {"line": 25, "character": 4}, "end": {"line": 25, "character": 19}},
            },
        ]

        with patch.object(client, "_request_document_symbols", new_callable=AsyncMock) as mock_req:
            mock_req.return_value = mock_symbols
            symbols = await client.get_document_symbols("/project/src/main.py")

        assert len(symbols) == 2
        assert all(isinstance(s, DocumentSymbol) for s in symbols)
        assert symbols[0].name == "MyClass"
        assert symbols[0].kind == SymbolKind.CLASS
        assert len(symbols[0].children) == 1
        assert symbols[0].children[0].name == "my_method"
        assert symbols[0].children[0].kind == SymbolKind.METHOD
        assert symbols[1].name == "helper_function"
        assert symbols[1].kind == SymbolKind.FUNCTION

    async def test_get_document_symbols_empty(self):
        """Client should return empty list when no symbols."""
        client = MoAILSPClient(project_root="/project")

        with patch.object(client, "_request_document_symbols", new_callable=AsyncMock) as mock_req:
            mock_req.return_value = []
            symbols = await client.get_document_symbols("/project/src/empty.py")

        assert symbols == []

    async def test_get_document_symbols_nested(self):
        """Client should handle deeply nested symbols."""
        client = MoAILSPClient(project_root="/project")

        mock_symbols = [
            {
                "name": "OuterClass",
                "kind": 5,
                "range": {"start": {"line": 0, "character": 0}, "end": {"line": 50, "character": 0}},
                "selectionRange": {"start": {"line": 0, "character": 6}, "end": {"line": 0, "character": 16}},
                "children": [
                    {
                        "name": "InnerClass",
                        "kind": 5,
                        "range": {
                            "start": {"line": 10, "character": 4},
                            "end": {"line": 40, "character": 4},
                        },
                        "selectionRange": {
                            "start": {"line": 10, "character": 10},
                            "end": {"line": 10, "character": 20},
                        },
                        "children": [
                            {
                                "name": "inner_method",
                                "kind": 6,
                                "range": {
                                    "start": {"line": 20, "character": 8},
                                    "end": {"line": 30, "character": 8},
                                },
                                "selectionRange": {
                                    "start": {"line": 20, "character": 12},
                                    "end": {"line": 20, "character": 24},
                                },
                            }
                        ],
                    }
                ],
            }
        ]

        with patch.object(client, "_request_document_symbols", new_callable=AsyncMock) as mock_req:
            mock_req.return_value = mock_symbols
            symbols = await client.get_document_symbols("/project/src/nested.py")

        assert len(symbols) == 1
        outer = symbols[0]
        assert outer.name == "OuterClass"
        assert len(outer.children) == 1
        inner = outer.children[0]
        assert inner.name == "InnerClass"
        assert len(inner.children) == 1
        assert inner.children[0].name == "inner_method"
