# LSP Data Models
"""LSP (Language Server Protocol) data models.

These dataclasses represent the core LSP types used for communication
between the client and language servers. All types follow the LSP 3.17
specification.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from enum import IntEnum
from typing import Union


class DiagnosticSeverity(IntEnum):
    """Severity level for diagnostics.

    Values match LSP specification:
    - ERROR = 1: Reports an error
    - WARNING = 2: Reports a warning
    - INFORMATION = 3: Reports an information
    - HINT = 4: Reports a hint
    """

    ERROR = 1
    WARNING = 2
    INFORMATION = 3
    HINT = 4


@dataclass
class Position:
    """A position in a text document.

    Position in a text document expressed as zero-based line and
    zero-based character offset. A position is between two characters
    like an 'insert' cursor in an editor.

    Attributes:
        line: Line position in a document (zero-based).
        character: Character offset on a line in a document (zero-based).
    """

    line: int
    character: int


@dataclass
class Range:
    """A range in a text document.

    A range in a text document expressed as start and end positions.
    A range is comparable to a selection in an editor.

    Attributes:
        start: The range's start position (inclusive).
        end: The range's end position (exclusive).
    """

    start: Position
    end: Position

    def contains(self, position: Position) -> bool:
        """Check if a position is within this range.

        Args:
            position: The position to check.

        Returns:
            True if the position is within the range, False otherwise.
        """
        # Check if position is on the same line for single-line ranges
        if self.start.line == self.end.line:
            return position.line == self.start.line and self.start.character <= position.character <= self.end.character

        # Multi-line range check
        if position.line < self.start.line or position.line > self.end.line:
            return False

        if position.line == self.start.line:
            return position.character >= self.start.character

        if position.line == self.end.line:
            return position.character <= self.end.character

        return True

    def is_single_line(self) -> bool:
        """Check if the range spans only a single line.

        Returns:
            True if start and end are on the same line.
        """
        return self.start.line == self.end.line


@dataclass
class Location:
    """A location inside a resource.

    Represents a location inside a resource, such as a line inside a
    text file.

    Attributes:
        uri: The resource URI (e.g., file:///path/to/file.py).
        range: The range within the resource.
    """

    uri: str
    range: Range


@dataclass
class Diagnostic:
    """A diagnostic representing an issue in source code.

    Represents a diagnostic, such as a compiler error or warning.
    Diagnostic objects are only valid in the scope of a resource.

    Attributes:
        range: The range at which the message applies.
        severity: The diagnostic's severity.
        code: The diagnostic's code (can be string, int, or None).
        source: The diagnostic's source (e.g., 'pyright', 'mypy').
        message: The diagnostic's message.
    """

    range: Range
    severity: DiagnosticSeverity
    code: Union[str, int, None]
    source: str
    message: str

    def is_error(self) -> bool:
        """Check if this diagnostic is an error.

        Returns:
            True if severity is ERROR, False otherwise.
        """
        return self.severity == DiagnosticSeverity.ERROR


@dataclass
class TextDocumentIdentifier:
    """Identifies a text document using a URI.

    Text documents are identified using a URI. On the protocol level,
    URIs are passed as strings.

    Attributes:
        uri: The text document's URI (e.g., file:///path/to/file.py).
    """

    uri: str

    @classmethod
    def from_path(cls, file_path: str) -> TextDocumentIdentifier:
        """Create a TextDocumentIdentifier from a file path.

        Args:
            file_path: The file path (absolute or relative).

        Returns:
            A TextDocumentIdentifier with a file:// URI.
        """
        # Ensure the path starts with /
        if not file_path.startswith("/"):
            file_path = "/" + file_path
        return cls(uri=f"file://{file_path}")


@dataclass
class TextDocumentPositionParams:
    """Parameters for text document position requests.

    A parameter literal used in requests to pass a text document
    and a position inside that document.

    Attributes:
        text_document: The text document.
        position: The position inside the text document.
    """

    text_document: TextDocumentIdentifier
    position: Position


@dataclass
class TextEdit:
    """A text edit applicable to a text document.

    A textual edit applicable to a text document.

    Attributes:
        range: The range of the text document to be manipulated.
        new_text: The string to be inserted. Empty for delete operations.
    """

    range: Range
    new_text: str

    def is_delete(self) -> bool:
        """Check if this edit represents a deletion.

        Returns:
            True if new_text is empty (deletion), False otherwise.
        """
        return self.new_text == ""

    def is_insert(self) -> bool:
        """Check if this edit represents an insertion.

        An insertion is when the range is zero-width (start == end)
        and new_text is non-empty.

        Returns:
            True if this is an insertion, False otherwise.
        """
        return (
            self.range.start.line == self.range.end.line
            and self.range.start.character == self.range.end.character
            and self.new_text != ""
        )


@dataclass
class WorkspaceEdit:
    """A workspace edit represents changes to many resources.

    A workspace edit represents changes to many resources managed
    in the workspace. The edit should be applied atomically.

    Attributes:
        changes: Holds changes to existing resources, keyed by URI.
    """

    changes: dict[str, list[TextEdit]] = field(default_factory=dict)

    def file_count(self) -> int:
        """Get the number of files affected by this edit.

        Returns:
            The number of unique file URIs in the changes.
        """
        return len(self.changes)


@dataclass
class HoverInfo:
    """Hover information for a symbol.

    The result of a hover request.

    Attributes:
        contents: The hover's content (can be markdown).
        range: Optional range for the symbol being hovered.
    """

    contents: str
    range: Range | None = None
