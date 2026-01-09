# LSP Models Tests - RED Phase
"""Tests for LSP data models (Position, Range, Location, Diagnostic, etc.)."""

from dataclasses import asdict

# Import from the module we're about to create
from moai_adk.lsp.models import (
    Diagnostic,
    DiagnosticSeverity,
    HoverInfo,
    Location,
    Position,
    Range,
    TextDocumentIdentifier,
    TextDocumentPositionParams,
    TextEdit,
    WorkspaceEdit,
)


class TestPosition:
    """Tests for Position dataclass."""

    def test_create_position(self):
        """Position should store line and character."""
        pos = Position(line=10, character=5)
        assert pos.line == 10
        assert pos.character == 5

    def test_position_zero_based(self):
        """Position should accept zero-based indices."""
        pos = Position(line=0, character=0)
        assert pos.line == 0
        assert pos.character == 0

    def test_position_equality(self):
        """Two positions with same values should be equal."""
        pos1 = Position(line=5, character=3)
        pos2 = Position(line=5, character=3)
        assert pos1 == pos2

    def test_position_inequality(self):
        """Positions with different values should not be equal."""
        pos1 = Position(line=5, character=3)
        pos2 = Position(line=5, character=4)
        assert pos1 != pos2

    def test_position_to_dict(self):
        """Position should be convertible to dict."""
        pos = Position(line=10, character=5)
        d = asdict(pos)
        assert d == {"line": 10, "character": 5}

    def test_position_from_dict(self):
        """Position should be creatable from dict."""
        data = {"line": 10, "character": 5}
        pos = Position(**data)
        assert pos.line == 10
        assert pos.character == 5


class TestRange:
    """Tests for Range dataclass."""

    def test_create_range(self):
        """Range should store start and end positions."""
        start = Position(line=0, character=0)
        end = Position(line=0, character=10)
        range_ = Range(start=start, end=end)
        assert range_.start == start
        assert range_.end == end

    def test_range_equality(self):
        """Two ranges with same positions should be equal."""
        range1 = Range(start=Position(line=0, character=0), end=Position(line=0, character=10))
        range2 = Range(start=Position(line=0, character=0), end=Position(line=0, character=10))
        assert range1 == range2

    def test_range_to_dict(self):
        """Range should be convertible to nested dict."""
        range_ = Range(start=Position(line=1, character=2), end=Position(line=3, character=4))
        d = asdict(range_)
        assert d == {"start": {"line": 1, "character": 2}, "end": {"line": 3, "character": 4}}

    def test_range_contains_position(self):
        """Range should check if a position is within it."""
        range_ = Range(start=Position(line=5, character=0), end=Position(line=5, character=20))
        pos_inside = Position(line=5, character=10)
        pos_outside = Position(line=6, character=0)
        assert range_.contains(pos_inside) is True
        assert range_.contains(pos_outside) is False

    def test_range_single_line(self):
        """Range should detect single-line ranges."""
        single = Range(start=Position(line=5, character=0), end=Position(line=5, character=20))
        multi = Range(start=Position(line=5, character=0), end=Position(line=10, character=0))
        assert single.is_single_line() is True
        assert multi.is_single_line() is False


class TestLocation:
    """Tests for Location dataclass."""

    def test_create_location(self):
        """Location should store URI and range."""
        loc = Location(
            uri="file:///path/to/file.py",
            range=Range(start=Position(line=10, character=0), end=Position(line=10, character=20)),
        )
        assert loc.uri == "file:///path/to/file.py"
        assert loc.range.start.line == 10

    def test_location_to_dict(self):
        """Location should be convertible to dict."""
        loc = Location(
            uri="file:///test.py", range=Range(start=Position(line=1, character=2), end=Position(line=3, character=4))
        )
        d = asdict(loc)
        assert d["uri"] == "file:///test.py"
        assert d["range"]["start"]["line"] == 1


class TestDiagnosticSeverity:
    """Tests for DiagnosticSeverity enum."""

    def test_severity_values(self):
        """DiagnosticSeverity should have correct LSP values."""
        assert DiagnosticSeverity.ERROR == 1
        assert DiagnosticSeverity.WARNING == 2
        assert DiagnosticSeverity.INFORMATION == 3
        assert DiagnosticSeverity.HINT == 4

    def test_severity_is_int_enum(self):
        """DiagnosticSeverity should be an IntEnum for JSON serialization."""
        assert isinstance(DiagnosticSeverity.ERROR, int)


class TestDiagnostic:
    """Tests for Diagnostic dataclass."""

    def test_create_diagnostic(self):
        """Diagnostic should store all required fields."""
        diag = Diagnostic(
            range=Range(start=Position(line=5, character=0), end=Position(line=5, character=10)),
            severity=DiagnosticSeverity.ERROR,
            code="E001",
            source="pyright",
            message="Undefined variable 'x'",
        )
        assert diag.severity == DiagnosticSeverity.ERROR
        assert diag.code == "E001"
        assert diag.source == "pyright"
        assert diag.message == "Undefined variable 'x'"

    def test_diagnostic_optional_code(self):
        """Diagnostic code can be None."""
        diag = Diagnostic(
            range=Range(start=Position(line=0, character=0), end=Position(line=0, character=5)),
            severity=DiagnosticSeverity.WARNING,
            code=None,
            source="mypy",
            message="Type mismatch",
        )
        assert diag.code is None

    def test_diagnostic_code_as_int(self):
        """Diagnostic code can be an integer."""
        diag = Diagnostic(
            range=Range(start=Position(line=0, character=0), end=Position(line=0, character=5)),
            severity=DiagnosticSeverity.INFORMATION,
            code=42,
            source="eslint",
            message="Some rule",
        )
        assert diag.code == 42

    def test_diagnostic_to_dict(self):
        """Diagnostic should be convertible to dict for JSON-RPC."""
        diag = Diagnostic(
            range=Range(start=Position(line=1, character=0), end=Position(line=1, character=10)),
            severity=DiagnosticSeverity.ERROR,
            code="E001",
            source="test",
            message="Test error",
        )
        d = asdict(diag)
        assert d["severity"] == 1  # DiagnosticSeverity.ERROR value
        assert d["message"] == "Test error"

    def test_diagnostic_is_error(self):
        """Diagnostic should have helper to check if it's an error."""
        error = Diagnostic(
            range=Range(start=Position(line=0, character=0), end=Position(line=0, character=1)),
            severity=DiagnosticSeverity.ERROR,
            code=None,
            source="test",
            message="Error",
        )
        warning = Diagnostic(
            range=Range(start=Position(line=0, character=0), end=Position(line=0, character=1)),
            severity=DiagnosticSeverity.WARNING,
            code=None,
            source="test",
            message="Warning",
        )
        assert error.is_error() is True
        assert warning.is_error() is False


class TestTextDocumentIdentifier:
    """Tests for TextDocumentIdentifier."""

    def test_create_text_document_identifier(self):
        """TextDocumentIdentifier should store URI."""
        doc = TextDocumentIdentifier(uri="file:///path/to/file.py")
        assert doc.uri == "file:///path/to/file.py"

    def test_text_document_identifier_from_path(self):
        """TextDocumentIdentifier should be creatable from file path."""
        doc = TextDocumentIdentifier.from_path("/path/to/file.py")
        assert doc.uri == "file:///path/to/file.py"


class TestTextDocumentPositionParams:
    """Tests for TextDocumentPositionParams."""

    def test_create_params(self):
        """TextDocumentPositionParams should store doc and position."""
        params = TextDocumentPositionParams(
            text_document=TextDocumentIdentifier(uri="file:///test.py"), position=Position(line=10, character=5)
        )
        assert params.text_document.uri == "file:///test.py"
        assert params.position.line == 10


class TestTextEdit:
    """Tests for TextEdit dataclass."""

    def test_create_text_edit(self):
        """TextEdit should store range and new text."""
        edit = TextEdit(
            range=Range(start=Position(line=5, character=0), end=Position(line=5, character=10)), new_text="replaced"
        )
        assert edit.new_text == "replaced"
        assert edit.range.start.line == 5

    def test_text_edit_delete(self):
        """TextEdit with empty new_text represents deletion."""
        edit = TextEdit(
            range=Range(start=Position(line=0, character=0), end=Position(line=0, character=10)), new_text=""
        )
        assert edit.is_delete() is True

    def test_text_edit_insert(self):
        """TextEdit with zero-width range represents insertion."""
        edit = TextEdit(
            range=Range(start=Position(line=5, character=10), end=Position(line=5, character=10)), new_text="inserted"
        )
        assert edit.is_insert() is True


class TestWorkspaceEdit:
    """Tests for WorkspaceEdit dataclass."""

    def test_create_workspace_edit(self):
        """WorkspaceEdit should store changes by document URI."""
        edit = WorkspaceEdit(
            changes={
                "file:///test.py": [
                    TextEdit(
                        range=Range(start=Position(line=0, character=0), end=Position(line=0, character=5)),
                        new_text="new",
                    )
                ]
            }
        )
        assert "file:///test.py" in edit.changes
        assert len(edit.changes["file:///test.py"]) == 1

    def test_workspace_edit_empty(self):
        """WorkspaceEdit can be empty."""
        edit = WorkspaceEdit(changes={})
        assert len(edit.changes) == 0

    def test_workspace_edit_file_count(self):
        """WorkspaceEdit should report number of affected files."""
        edit = WorkspaceEdit(changes={"file:///a.py": [], "file:///b.py": []})
        assert edit.file_count() == 2


class TestHoverInfo:
    """Tests for HoverInfo dataclass."""

    def test_create_hover_info(self):
        """HoverInfo should store contents and optional range."""
        hover = HoverInfo(
            contents="def foo(): ...",
            range=Range(start=Position(line=0, character=0), end=Position(line=0, character=3)),
        )
        assert hover.contents == "def foo(): ..."
        assert hover.range is not None

    def test_hover_info_no_range(self):
        """HoverInfo range can be None."""
        hover = HoverInfo(contents="Some documentation", range=None)
        assert hover.contents == "Some documentation"
        assert hover.range is None

    def test_hover_info_markdown_contents(self):
        """HoverInfo can have markdown contents."""
        hover = HoverInfo(contents="```python\ndef foo():\n    pass\n```", range=None)
        assert "```python" in hover.contents
