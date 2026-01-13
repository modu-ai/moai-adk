"""Additional coverage tests for LSP models.

Tests for lines not covered by existing tests.
"""


from moai_adk.lsp.models import (
    Diagnostic,
    DiagnosticSeverity,
    Position,
    Range,
    TextDocumentIdentifier,
    TextEdit,
)


class TestRangeContainsMultiLine:
    """Test Range.contains() method for multi-line ranges."""

    def test_contains_position_before_range(self):
        """Should return False for position before range."""
        start = Position(line=5, character=0)
        end = Position(line=10, character=0)
        range_obj = Range(start=start, end=end)

        position = Position(line=3, character=5)

        assert not range_obj.contains(position)

    def test_contains_position_after_range(self):
        """Should return False for position after range."""
        start = Position(line=5, character=0)
        end = Position(line=10, character=0)
        range_obj = Range(start=start, end=end)

        position = Position(line=12, character=5)

        assert not range_obj.contains(position)

    def test_contains_position_on_start_line(self):
        """Should check character position on start line."""
        start = Position(line=5, character=10)
        end = Position(line=10, character=0)
        range_obj = Range(start=start, end=end)

        # Position before start character
        position_before = Position(line=5, character=5)
        assert not range_obj.contains(position_before)

        # Position after start character
        position_after = Position(line=5, character=15)
        assert range_obj.contains(position_after)

    def test_contains_position_on_end_line(self):
        """Should check character position on end line."""
        start = Position(line=5, character=0)
        end = Position(line=10, character=20)
        range_obj = Range(start=start, end=end)

        # Position before end character
        position_before = Position(line=10, character=15)
        assert range_obj.contains(position_before)

        # Position after end character
        position_after = Position(line=10, character=25)
        assert not range_obj.contains(position_after)

    def test_contains_position_in_middle_lines(self):
        """Should return True for any character in middle lines."""
        start = Position(line=5, character=0)
        end = Position(line=10, character=0)
        range_obj = Range(start=start, end=end)

        # Middle line positions
        position1 = Position(line=7, character=0)
        position2 = Position(line=7, character=100)
        position3 = Position(line=9, character=50)

        assert range_obj.contains(position1)
        assert range_obj.contains(position2)
        assert range_obj.contains(position3)


class TestTextDocumentIdentifierFromPath:
    """Test TextDocumentIdentifier.from_path method."""

    def test_from_path_with_absolute_path(self):
        """Should create identifier with absolute path."""
        identifier = TextDocumentIdentifier.from_path("/absolute/path.py")

        assert identifier.uri == "file:///absolute/path.py"

    def test_from_path_with_relative_path(self):
        """Should add leading slash to relative path."""
        identifier = TextDocumentIdentifier.from_path("relative/path.py")

        assert identifier.uri == "file:///relative/path.py"


class TestTextEditMethods:
    """Test TextEdit method variations."""

    def test_is_delete_with_empty_text(self):
        """Should return True when new_text is empty."""
        range_obj = Range(
            start=Position(line=0, character=0),
            end=Position(line=0, character=5),
        )
        edit = TextEdit(range=range_obj, new_text="")

        assert edit.is_delete()
        assert not edit.is_insert()

    def test_is_insert_with_zero_width_range(self):
        """Should return True for zero-width range with non-empty text."""
        range_obj = Range(
            start=Position(line=5, character=10),
            end=Position(line=5, character=10),
        )
        edit = TextEdit(range=range_obj, new_text="new text")

        assert edit.is_insert()
        assert not edit.is_delete()

    def test_is_replace(self):
        """Should return False for both delete and insert on replace."""
        range_obj = Range(
            start=Position(line=0, character=0),
            end=Position(line=0, character=5),
        )
        edit = TextEdit(range=range_obj, new_text="replacement")

        assert not edit.is_delete()
        assert not edit.is_insert()


class TestDiagnosticIsError:
    """Test Diagnostic.is_error method with all severities."""

    def test_is_error_with_error_severity(self):
        """Should return True for error severity."""
        diagnostic = Diagnostic(
            range=Range(
                start=Position(line=0, character=0),
                end=Position(line=0, character=1),
            ),
            severity=DiagnosticSeverity.ERROR,
            code=None,
            source="test",
            message="Error message",
        )

        assert diagnostic.is_error()

    def test_is_error_with_warning_severity(self):
        """Should return False for warning severity."""
        diagnostic = Diagnostic(
            range=Range(
                start=Position(line=0, character=0),
                end=Position(line=0, character=1),
            ),
            severity=DiagnosticSeverity.WARNING,
            code=None,
            source="test",
            message="Warning message",
        )

        assert not diagnostic.is_error()

    def test_is_error_with_hint_severity(self):
        """Should return False for hint severity."""
        diagnostic = Diagnostic(
            range=Range(
                start=Position(line=0, character=0),
                end=Position(line=0, character=1),
            ),
            severity=DiagnosticSeverity.HINT,
            code=None,
            source="test",
            message="Hint message",
        )

        assert not diagnostic.is_error()
