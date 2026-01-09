# MoAI LSP Client
"""LSP client interface for MoAI-ADK.

This module provides the main LSP client interface for getting diagnostics,
finding references, renaming symbols, and other LSP operations.
"""

from __future__ import annotations

from pathlib import Path
from typing import Any

from moai_adk.lsp.models import (
    Diagnostic,
    DiagnosticSeverity,
    HoverInfo,
    Location,
    Position,
    Range,
    TextEdit,
    WorkspaceEdit,
)
from moai_adk.lsp.server_manager import LSPServerManager


class LSPClientError(Exception):
    """Error raised by the LSP client."""

    pass


class MoAILSPClient:
    """LSP client for MoAI-ADK.

    Provides a high-level interface for LSP operations including:
    - Getting diagnostics for files
    - Finding references to symbols
    - Renaming symbols across the project
    - Getting hover information

    Attributes:
        project_root: The root directory of the project.
        server_manager: Manager for LSP server processes.
    """

    def __init__(self, project_root: str | Path) -> None:
        """Initialize the LSP client.

        Args:
            project_root: Path to the project root directory.
        """
        self.project_root = Path(project_root)
        self.server_manager = LSPServerManager()
        self._load_config()

    def _load_config(self) -> None:
        """Load LSP configuration from .lsp.json file."""
        config_path = self.project_root / ".lsp.json"
        if config_path.exists():
            self.server_manager.load_config(config_path)

    # ==========================================================================
    # Public API
    # ==========================================================================

    async def get_diagnostics(self, file_path: str) -> list[Diagnostic]:
        """Get diagnostics for a file.

        Args:
            file_path: Path to the file.

        Returns:
            List of Diagnostic objects.
        """
        raw_diagnostics = await self._request_diagnostics(file_path)
        return [self._parse_diagnostic(d) for d in raw_diagnostics]

    async def find_references(self, file_path: str, position: Position) -> list[Location]:
        """Find all references to the symbol at position.

        Args:
            file_path: Path to the file.
            position: Position of the symbol.

        Returns:
            List of Location objects.
        """
        raw_refs = await self._request_references(file_path, position)
        return [self._parse_location(r) for r in raw_refs]

    async def rename_symbol(self, file_path: str, position: Position, new_name: str) -> WorkspaceEdit:
        """Rename the symbol at position.

        Args:
            file_path: Path to the file.
            position: Position of the symbol.
            new_name: New name for the symbol.

        Returns:
            WorkspaceEdit with all changes.
        """
        raw_edit = await self._request_rename(file_path, position, new_name)
        return self._parse_workspace_edit(raw_edit)

    async def get_hover_info(self, file_path: str, position: Position) -> HoverInfo | None:
        """Get hover information for position.

        Args:
            file_path: Path to the file.
            position: Position to get hover info for.

        Returns:
            HoverInfo or None if not available.
        """
        raw_hover = await self._request_hover(file_path, position)
        if raw_hover is None:
            return None
        return self._parse_hover_info(raw_hover)

    def get_language_for_file(self, file_path: str) -> str | None:
        """Get the language for a file.

        Args:
            file_path: Path to the file.

        Returns:
            Language identifier or None.
        """
        return self.server_manager.get_language_for_file(file_path)

    async def ensure_server_running(self, language: str) -> None:
        """Ensure an LSP server is running for a language.

        Args:
            language: Language identifier.
        """
        await self.server_manager.start_server(language)

    async def cleanup(self) -> None:
        """Clean up by stopping all servers."""
        await self.server_manager.stop_all_servers()

    # ==========================================================================
    # Internal request methods (to be overridden in tests or by subclasses)
    # ==========================================================================

    async def _request_diagnostics(self, file_path: str) -> list[dict[str, Any]]:
        """Request diagnostics from the LSP server.

        Args:
            file_path: Path to the file.

        Returns:
            Raw diagnostic data from server.
        """
        # This would normally communicate with the LSP server
        # For now, return empty list (actual implementation in Phase 2+)
        return []

    async def _request_references(self, file_path: str, position: Position) -> list[dict[str, Any]]:
        """Request references from the LSP server.

        Args:
            file_path: Path to the file.
            position: Position of the symbol.

        Returns:
            Raw reference data from server.
        """
        return []

    async def _request_rename(self, file_path: str, position: Position, new_name: str) -> dict[str, Any]:
        """Request rename from the LSP server.

        Args:
            file_path: Path to the file.
            position: Position of the symbol.
            new_name: New name for the symbol.

        Returns:
            Raw workspace edit data from server.
        """
        return {"changes": {}}

    async def _request_hover(self, file_path: str, position: Position) -> dict[str, Any] | None:
        """Request hover info from the LSP server.

        Args:
            file_path: Path to the file.
            position: Position to get hover info for.

        Returns:
            Raw hover data from server or None.
        """
        return None

    # ==========================================================================
    # Helper methods
    # ==========================================================================

    def _file_to_uri(self, file_path: str) -> str:
        """Convert a file path to a file:// URI.

        Args:
            file_path: Path to the file.

        Returns:
            File URI string.
        """
        if not file_path.startswith("/"):
            file_path = "/" + file_path
        return f"file://{file_path}"

    def _uri_to_file(self, uri: str) -> str:
        """Convert a file:// URI to a file path.

        Args:
            uri: File URI string.

        Returns:
            File path string.
        """
        if uri.startswith("file://"):
            return uri[7:]
        return uri

    def _parse_diagnostic(self, raw: dict[str, Any]) -> Diagnostic:
        """Parse a Diagnostic from raw LSP data.

        Args:
            raw: Raw diagnostic dictionary from LSP.

        Returns:
            Diagnostic instance.
        """
        range_data = raw["range"]
        return Diagnostic(
            range=self._parse_range(range_data),
            severity=DiagnosticSeverity(raw["severity"]),
            code=raw.get("code"),
            source=raw.get("source", ""),
            message=raw["message"],
        )

    def _parse_location(self, raw: dict[str, Any]) -> Location:
        """Parse a Location from raw LSP data.

        Args:
            raw: Raw location dictionary from LSP.

        Returns:
            Location instance.
        """
        return Location(
            uri=raw["uri"],
            range=self._parse_range(raw["range"]),
        )

    def _parse_range(self, raw: dict[str, Any]) -> Range:
        """Parse a Range from raw LSP data.

        Args:
            raw: Raw range dictionary from LSP.

        Returns:
            Range instance.
        """
        return Range(
            start=Position(
                line=raw["start"]["line"],
                character=raw["start"]["character"],
            ),
            end=Position(
                line=raw["end"]["line"],
                character=raw["end"]["character"],
            ),
        )

    def _parse_workspace_edit(self, raw: dict[str, Any]) -> WorkspaceEdit:
        """Parse a WorkspaceEdit from raw LSP data.

        Args:
            raw: Raw workspace edit dictionary from LSP.

        Returns:
            WorkspaceEdit instance.
        """
        changes: dict[str, list[TextEdit]] = {}

        raw_changes = raw.get("changes", {})
        for uri, edits in raw_changes.items():
            changes[uri] = [
                TextEdit(
                    range=self._parse_range(edit["range"]),
                    new_text=edit["newText"],
                )
                for edit in edits
            ]

        return WorkspaceEdit(changes=changes)

    def _parse_hover_info(self, raw: dict[str, Any]) -> HoverInfo:
        """Parse HoverInfo from raw LSP data.

        Args:
            raw: Raw hover dictionary from LSP.

        Returns:
            HoverInfo instance.
        """
        contents = raw.get("contents", "")

        # Handle MarkupContent format
        if isinstance(contents, dict):
            contents = contents.get("value", "")

        range_data = raw.get("range")
        hover_range = self._parse_range(range_data) if range_data else None

        return HoverInfo(contents=contents, range=hover_range)
